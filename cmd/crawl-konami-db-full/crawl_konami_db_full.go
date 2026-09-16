package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/driver/sqlite"
	"github.com/daominah/yugioh_card_editor/pkg/driver/ygocdb"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

// Adjustable knobs for the crawl. Edit and re-run.
const (
	firstTCGCardID         = 4007
	firstRushCardID        = 15150
	maxConsecutiveNotFound = 400
	maxGoroutines          = 8
	// Caps on successful cards per locale; 0 = full crawl.
	// Split between Standard (TCG/OCG) and Rush
	// so each can be sampled or fully crawled independently.
	nLimit     = 0
	nLimitRush = 0

	// isCacheOnly forbids every HTTP request:
	// HTML pages must already exist in data/html_cache/,
	// and the YGOCDB password feed comes from the embedded snapshot
	// (pkg/core/ygocdb_card_password.json) instead of the live API.
	// Useful for reparsing after a parser fix, or for fully offline runs.
	isCacheOnly = false
	// Hand-maintained snapshot of Konami's latest released cardID.
	// Used by readOrFetchHTML to apply a 7-day cache TTL above this threshold:
	// caches at or above are time-sensitive
	// (a "not found" tail or a recent card page may change),
	// caches below are trusted indefinitely.
	latestKnownCardID     = 23428 // the latest "en" is 22355, "ko" is 22617
	latestKnownRushCardID = 23471

	isIncludingCardPassword = true // 3rd-party data ygocdb.com
)

func main() {
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatalf("error os.Create cpuprofile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("error pprof.StartCPUProfile: %v", err)
		}
		defer pprof.StopCPUProfile()
		log.Printf("CPU profile will be written to %v", *cpuProfile)
	}

	beginT := time.Now()
	defer func() {
		log.Printf("end crawl-konami-db-full, duration: %v", time.Since(beginT))
	}()

	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	dataDir := filepath.Join(projectRoot, "data")
	dbLocalFile := filepath.Join(dataDir, "yugioh.db")
	htmlCacheDir := filepath.Join(dataDir, "html_cache")

	dbSQLite, writer := openFreshBulkDB(dbLocalFile)
	defer dbSQLite.Close()
	var db core.Database = writer

	// TCG/OCG: JA first for widest card ID coverage, then EN to correct stats
	// that JA could not parse (attribute names use EN strings in lookup maps),
	// then KO for Korean text and KR set codes.
	for _, locale := range []string{"ja", "en", "ko"} {
		log.Printf("begin crawlLocale %v %v", konami.StandardDB, locale)
		crawlLocale(db, htmlCacheDir, konami.StandardDB, locale, nLimit)
	}

	// Rush Duel / Duel Links: only JA and KO are available on rushdb.
	for _, locale := range []string{"ja", "ko"} {
		log.Printf("begin crawlLocale %v %v", konami.RushDB, locale)
		crawlLocale(db, htmlCacheDir, konami.RushDB, locale, nLimitRush)
	}

	if isIncludingCardPassword {
		populateCardPasswords(db)
	}

	finalizeBulkLoad(dbSQLite, writer)
}

// openFreshBulkDB removes any pre-existing yugioh.db / -wal / -shm files
// at path (the bulk-load pragmas in OpenForBulkLoad require a brand-new file)
// and opens a fresh DB with a single-tx writer ready for the crawl.
//
// The returned TxWriter funnels every parser goroutine's write
// through one background goroutine that owns a single *sql.Tx;
// the whole crawl is one transaction that commits at the end,
// so the page cache absorbs most of the work and we hit disk once.
func openFreshBulkDB(path string) (*sqlite.DB, *sqlite.TxWriter) {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		p := path + suffix
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			log.Fatalf("error os.Remove %v: %v", p, err)
		}
	}
	dbSQLite, err := sqlite.OpenForBulkLoad(path)
	if err != nil {
		log.Fatalf("error sqlite.OpenForBulkLoad: %v", err)
	}
	log.Printf("database: %v (bulk-load mode)", path)
	writer, err := dbSQLite.BeginTxWriter()
	if err != nil {
		log.Fatalf("error BeginTxWriter: %v", err)
	}
	return dbSQLite, writer
}

// finalizeBulkLoad commits the single bulk-load transaction
// and creates the non-PK indexes.
// Indexes are deferred to the end of the load
// so per-row inserts touch only the PK B-trees;
// they're built once here as a sort-then-build
// instead of being rebalanced incrementally per row.
func finalizeBulkLoad(db *sqlite.DB, writer *sqlite.TxWriter) {
	commitBegin := time.Now()
	if err := writer.Commit(); err != nil {
		log.Fatalf("error TxWriter.Commit: %v", err)
	}
	log.Printf("end commit, duration: %v", formatDuration(time.Since(commitBegin)))

	indexBegin := time.Now()
	if err := db.CreateIndexes(); err != nil {
		log.Fatalf("error CreateIndexes: %v", err)
	}
	log.Printf("end create indexes, duration: %v", formatDuration(time.Since(indexBegin)))
}

// populateCardPasswords fetches YGOCDB's password table
// and writes it to card_passwords.
// Konami's official DB does not expose the 8-digit password,
// so this is a separate step against a third-party source.
// When isCacheOnly is true,
// the embedded snapshot at pkg/core/ygocdb_card_password.json is used
// so the run stays fully offline; otherwise the live YGOCDB API is queried.
func populateCardPasswords(db core.Database) {
	beginT := time.Now()
	var src core.MapCardsPasswordInitiator
	if isCacheOnly {
		src = &core.YgocdbStaticData{}
	} else {
		src = &ygocdb.YgocdbDownloadFreshData{}
	}
	cardPasswords, err := src.InitMapCardsPassword()
	if err != nil {
		log.Printf("error InitMapCardsPassword: %v", err)
		return
	}
	log.Printf("begin insert card_passwords")
	for cardID, cp := range cardPasswords {
		if err := db.UpsertCardPassword(string(cardID), cp.Password, cp.CardName); err != nil {
			log.Printf("error UpsertCardPassword cardID=%v: %v", cardID, err)
		}
	}
	log.Printf("end insert card_passwords count=%v duration=%v",
		len(cardPasswords), formatDuration(time.Since(beginT)))
}

func crawlLocale(database core.Database, htmlCacheDir string,
	konamiDB konami.KonamiDB, locale string, nLimit int) {
	baseURL := fmt.Sprintf("https://www.db.yugioh-card.com/%v/card_search.action", konamiDB)
	cacheDir := filepath.Join(htmlCacheDir, string(konamiDB), locale)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("error os.MkdirAll %v: %v", cacheDir, err)
	}

	notFoundMarker, ok := notFoundMarkers[locale]
	if !ok {
		log.Fatalf("no notFoundMarker for locale=%v", locale)
	}

	firstCardID := firstTCGCardID
	latestKnownID := latestKnownCardID
	if konamiDB == konami.RushDB {
		firstCardID = firstRushCardID
		latestKnownID = latestKnownRushCardID
	}

	var gameVersion string
	switch {
	case konamiDB == konami.RushDB:
		gameVersion = "RushDuel"
	case locale == "en":
		gameVersion = "TCG"
	default:
		gameVersion = "OCG"
	}

	httpClient := &http.Client{Timeout: 128 * time.Second}

	var (
		mu             sync.Mutex
		maxFoundCardID int
		successCount   int
		wg             sync.WaitGroup
	)

	beginT := time.Now()
	log.Printf("begin locale=%v gameVersion=%v at %v", locale, gameVersion, beginT.Format(time.RFC3339))

	// processCard is the per-card pipeline.
	// It runs on every cardID a worker dequeues from cardIDCh:
	// read HTML (cache or fetch), short-circuit the "card not found" page,
	// parse once with konami.Parser,
	// and write all derived rows through the TxWriter.
	// Hoisted out of the dispatch loop so a fixed worker pool can call it
	// instead of spawning a goroutine per cardID,
	// eliminating ~70k goroutine spawns per crawl
	// and the scheduler contention they caused.
	processCard := func(cardNum int) {
		cardID := konami.CardID(fmt.Sprintf("%v", cardNum))
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("%v.html", cardNum))

		bodyBs, err := readOrFetchHTML(httpClient, cachePath, baseURL, locale, cardNum, latestKnownID)
		if err != nil {
			log.Printf("error cardID=%v locale=%v: %v", cardID, locale, err)
			return
		}
		if bytes.Contains(bodyBs, []byte(notFoundMarker)) {
			return
		}

		mu.Lock()
		if maxFoundCardID < cardNum {
			maxFoundCardID = cardNum
		}
		mu.Unlock()

		// One Parser instance per page so html.Parse runs once
		// and the DOM tree (~30% CPU per card across XPath calls)
		// is reused by every extractor below.
		parser := konami.NewParser(bodyBs, cardID)
		text := parser.LocaleText()
		prints := parser.Prints()
		var card konami.Card
		var rushCard konami.CardRushDuel
		if konamiDB == konami.RushDB {
			rushCard = parser.RushCard()
		} else {
			card = parser.Card()
		}

		// DB writes: the sqlite driver serializes these through MaxOpenConns(1).
		// Skip cards that parsed as empty: the page existed
		// but the parser could not extract a card_type,
		// meaning the HTML format was unrecognised.
		// successCount is incremented only after a successful parse,
		// so nLimit counts saved cards, not just pages that exist.
		//
		// Card stats (cards / cards_rush) are language-independent,
		// so they are written only during the "ja" crawl pass.
		// The "en" and "ko" passes still parse cards
		// (to drive the empty-parse skip and successCount)
		// but do not call UpsertCard/UpsertCardRush;
		// they only contribute card_texts and set_cards rows below.
		// The locale order in main() puts "ja" first,
		// so stats land before any later locale's text/set rows
		// are written for the same card.
		if konamiDB == konami.RushDB {
			if rushCard.CardType == "" {
				log.Printf("warn: empty parse cardID=%v locale=%v rushdb, skipping", cardID, locale)
				return
			}
			if locale == "ja" {
				if err := database.UpsertCardRush(rushCard); err != nil {
					log.Printf("error UpsertCardRush cardID=%v: %v", cardID, err)
				}
			}
		} else {
			if card.CardType == "" {
				log.Printf("warn: empty parse cardID=%v locale=%v, skipping", cardID, locale)
				return
			}
			if locale == "ja" {
				if err := database.UpsertCard(card); err != nil {
					log.Printf("error UpsertCard cardID=%v: %v", cardID, err)
				}
			}
		}

		mu.Lock()
		successCount++
		mu.Unlock()
		if err := database.UpsertCardText(string(cardID), locale, text); err != nil {
			log.Printf("error UpsertCardText cardID=%v locale=%v: %v", cardID, locale, err)
		}
		prints = konami.DedupCardPrints(prints)
		for _, p := range prints {
			if p.Position == "" {
				continue
			}
			set := konami.KonamiSet{
				Abbreviation:  p.SetCode(),
				YuGiOhVersion: konami.YuGiOhVersion(gameVersion),
				ReleaseDate:   p.Date,
			}
			switch locale {
			case "ja":
				set.NameJA = p.SetName
			case "ko":
				set.NameKO = p.SetName
			case "en":
				set.NameEN = p.SetName
			}
			upsertSetIfChanged(database, set)
		}
		if err := database.UpsertSetCards(prints); err != nil {
			log.Printf("error UpsertSetCards cardID=%v: %v", cardID, err)
		}
	}

	// Worker pool: maxGoroutines long-lived workers drain cardIDCh.
	// Channel buffer = maxGoroutines provides the same back-pressure
	// as the previous semaphore (the producer blocks when workers can't keep up),
	// so the producer's view of maxFoundCardID stays close to actually-processed.
	cardIDCh := make(chan int, maxGoroutines)
	for range maxGoroutines {
		wg.Go(func() {
			for cardNum := range cardIDCh {
				processCard(cardNum)
			}
		})
	}

	for i := firstCardID; ; i++ {
		mu.Lock()
		currentMax := maxFoundCardID
		done := nLimit > 0 && successCount >= nLimit
		mu.Unlock()
		if done {
			break
		}
		if currentMax > 0 && i-currentMax > maxConsecutiveNotFound {
			break
		}

		if (i-firstCardID)%1000 == 0 && i > firstCardID {
			elapsed := time.Since(beginT)
			processed := i - firstCardID
			mu.Lock()
			foundMax := maxFoundCardID
			mu.Unlock()
			log.Printf("locale=%v processed=%d maxFoundID=%d elapsed=%v",
				locale, processed, foundMax, formatDuration(elapsed))
		}

		cardIDCh <- i
	}
	close(cardIDCh)
	wg.Wait()
	log.Printf("________________________________________")
	log.Printf("done %v locale=%v maxFoundCardID=%v latestKnownID=%v duration=%v",
		konamiDB, locale, maxFoundCardID, latestKnownID, formatDuration(time.Since(beginT)))
	if maxFoundCardID > latestKnownID {
		log.Printf("warn: maxFoundCardID=%v exceeds latestKnownID=%v, bump the const to skip wasted re-downloads next run",
			maxFoundCardID, latestKnownID)
	}
	log.Printf("________________________________________")
}

// readOrFetchHTML returns HTML for cardID: from the on-disk cache if present,
// otherwise fetches from the Konami DB and writes the response to the cache.
// The cache allows re-parsing during development without re-downloading.
//
// Cache freshness rule: caches for cardNum >= latestKnownID expire after 7 days
// because Konami may publish a new card at that ID
// (or fill a tail "not found" gap with a real release).
// Caches below the threshold are reused indefinitely.
// isCacheOnly=true overrides the freshness rule: no network is allowed at all,
// so any cached file is returned even when stale.
func readOrFetchHTML(client *http.Client, cachePath, baseURL, locale string, cardNum, latestKnownID int) ([]byte, error) {
	if stat, err := os.Stat(cachePath); err == nil {
		fresh := isCacheOnly ||
			cardNum < latestKnownID ||
			time.Since(stat.ModTime()) <= 7*24*time.Hour
		if fresh {
			if data, err := os.ReadFile(cachePath); err == nil {
				return data, nil
			}
		}
	}

	if isCacheOnly {
		return nil, fmt.Errorf("cache miss and isCacheOnly=true")
	}

	url := fmt.Sprintf("%v?ope=2&request_locale=%v&cid=%v", baseURL, locale, cardNum)
	log.Printf("begin http Get %v", url)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error http.Get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error io.ReadAll: %w", err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 400) {
		return nil, fmt.Errorf("HTTP %v", resp.StatusCode)
	}

	if err := os.WriteFile(cachePath, body, 0644); err != nil {
		log.Printf("warn: failed to cache cardID=%v locale=%v: %v", cardNum, locale, err)
	}
	return body, nil
}

func formatDuration(d time.Duration) string {
	if d < 60*time.Second {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm%.3fs", minutes, seconds)
}

// notFoundMarkers is the localized "card not found" string
// that Konami renders inside the <nav id="pan_nav"> breadcrumb
// when a cardID has no card.
// The English string only appears on en pages,
// so the marker must be looked up per locale; otherwise
// ja/ko not-found pages bypass the early return,
// falsely advance maxFoundCardID,
// and disable the maxConsecutiveNotFound early-stop.
var notFoundMarkers = map[string]string{
	"en": "Card information not found",
	"ja": "カード情報がありません",
	"ko": "카드 정보가 없습니다",
}

// cpuProfile, if non-empty, enables CPU profiling
// and writes the sampled profile to that path.
// Inspect with `go tool pprof <path>` (top, list, web).
var cpuProfile = flag.String("cpuprofile", "", "write CPU profile to this file")

// upsertSetIfChanged is a write-through cache over Database.UpsertSet:
// it skips the SQL call when the merge of the in-memory cached row
// and the new value would produce no field changes.
// Konami's per-card pages echo the same set name
// across every card that belongs to a given set,
// so most UpsertSet calls in a hot loop carry no new data
// and only burn the writer mutex; this cache filters them out.
// Crosses locale passes intentionally: the JA pass populates NameJA,
// the EN pass populates NameEN on the same cached row,
// and each genuinely-new locale name triggers exactly one SQL call.
var (
	seenSetsMu sync.Mutex
	seenSets   = make(map[string]konami.KonamiSet) // set_code -> last merged value
)

func upsertSetIfChanged(database core.Database, fresh konami.KonamiSet) {
	seenSetsMu.Lock()
	cached, exists := seenSets[fresh.Abbreviation]
	merged := cached
	if !exists {
		merged = fresh
	} else {
		// First-write-wins per field:
		// only fill empty cached fields from non-empty fresh fields.
		// Mirrors UpsertSet's ON CONFLICT semantics.
		if merged.NameJA == "" && fresh.NameJA != "" {
			merged.NameJA = fresh.NameJA
		}
		if merged.NameKO == "" && fresh.NameKO != "" {
			merged.NameKO = fresh.NameKO
		}
		if merged.NameEN == "" && fresh.NameEN != "" {
			merged.NameEN = fresh.NameEN
		}
		if merged.ReleaseDate == "" && fresh.ReleaseDate != "" {
			merged.ReleaseDate = fresh.ReleaseDate
		}
	}
	if exists && merged == cached {
		seenSetsMu.Unlock()
		return // nothing new to write
	}
	seenSets[fresh.Abbreviation] = merged
	seenSetsMu.Unlock()
	if err := database.UpsertSet(merged); err != nil {
		log.Printf("error UpsertSet %v: %v", merged.Abbreviation, err)
	}
}
