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
	"slices"
	"sync"
	"time"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/core"
	"github.com/daominah/yugioh_card_editor_database/pkg/driver/sqlite"
	"github.com/daominah/yugioh_card_editor_database/pkg/driver/ygocdb"
	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
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
	latestKnownCardID     = 23902
	latestKnownRushCardID = 24004

	isIncludingCardPassword = true // 3rd-party data ygocdb.com

	// isFillingNameENFromAsiaEnglish runs fillCardNameENFromAsiaEnglish after the standard passes.
	// Off: on the 2026-11-04 full crawl, all 404 cards still missing an English name
	// had no "ae" page either, so it filled 0 names for about 400 extra requests.
	isFillingNameENFromAsiaEnglish = false
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
	if isFillingNameENFromAsiaEnglish {
		fillCardNameENFromAsiaEnglish(db, htmlCacheDir)
	}

	// Rush Duel / Duel Links: only JA and KO are available on rushdb.
	for _, locale := range []string{"ja", "ko"} {
		log.Printf("begin crawlLocale %v %v", konami.RushDB, locale)
		crawlLocale(db, htmlCacheDir, konami.RushDB, locale, nLimitRush)
	}

	writeChosenSets(db)

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
	// data/ may not exist yet, e.g. in a checkout without a committed yugioh.db.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatalf("error os.MkdirAll %v: %v", filepath.Dir(path), err)
	}
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
				if card.CardNameEN == "" {
					cardsMissingNameEN.add(cardID)
				}
			}
			// Konami's current JA pages no longer print the English name,
			// so the EN page name fills cards.card_name_en for those cards.
			if locale == "en" && card.CardName != "" && cardsMissingNameEN.remove(cardID) {
				if err := database.FillCardNameEN(string(cardID), card.CardName); err != nil {
					log.Printf("error FillCardNameEN cardID=%v: %v", cardID, err)
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
			recordSetName(p.SetCode(), konami.YuGiOhVersion(gameVersion), locale, p)
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

// fillCardNameENFromAsiaEnglish fills cards.card_name_en
// for cards still missing it after the "ja" and "en" passes,
// mostly OCG-only cards that have no EN (TCG) page.
// It reads the "ae" (English Asia, the OCG English edition) page name.
// Only the name is used: "ae" card texts and set prints are not stored.
// It visits only the missing card IDs instead of the whole ID range,
// so it cannot use crawlLocale (whose early stop needs consecutive IDs).
func fillCardNameENFromAsiaEnglish(database core.Database, htmlCacheDir string) {
	const locale = "ae"
	beginT := time.Now()
	cardIDs := cardsMissingNameEN.sorted()
	log.Printf("begin fillCardNameENFromAsiaEnglish missing=%v", len(cardIDs))

	baseURL := fmt.Sprintf("https://www.db.yugioh-card.com/%v/card_search.action", konami.StandardDB)
	cacheDir := filepath.Join(htmlCacheDir, string(konami.StandardDB), locale)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("error os.MkdirAll %v: %v", cacheDir, err)
	}
	httpClient := &http.Client{Timeout: 128 * time.Second}

	var (
		mu          sync.Mutex
		filledCount int
		wg          sync.WaitGroup
	)
	cardIDCh := make(chan konami.CardID, maxGoroutines)
	for range maxGoroutines {
		wg.Go(func() {
			for cardID := range cardIDCh {
				cardNum := cardID.Int()
				cachePath := filepath.Join(cacheDir, fmt.Sprintf("%v.html", cardNum))
				bodyBs, err := readOrFetchHTML(httpClient, cachePath, baseURL, locale, cardNum, latestKnownCardID)
				if err != nil {
					log.Printf("error cardID=%v locale=%v: %v", cardID, locale, err)
					continue
				}
				if bytes.Contains(bodyBs, []byte(notFoundMarkers[locale])) {
					continue
				}
				nameEN := konami.NewParser(bodyBs, cardID).LocaleText().Name
				if nameEN == "" {
					continue
				}
				if err := database.FillCardNameEN(string(cardID), nameEN); err != nil {
					log.Printf("error FillCardNameEN cardID=%v: %v", cardID, err)
					continue
				}
				mu.Lock()
				filledCount++
				mu.Unlock()
			}
		})
	}
	for _, cardID := range cardIDs {
		cardIDCh <- cardID
	}
	close(cardIDCh)
	wg.Wait()
	log.Printf("end fillCardNameENFromAsiaEnglish filled=%v stillMissing=%v duration=%v",
		filledCount, len(cardIDs)-filledCount, formatDuration(time.Since(beginT)))
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
	"ae": "Card information not found",
	"ja": "カード情報がありません",
	"ko": "카드 정보가 없습니다",
}

// cpuProfile, if non-empty, enables CPU profiling
// and writes the sampled profile to that path.
// Inspect with `go tool pprof <path>` (top, list, web).
var cpuProfile = flag.String("cpuprofile", "", "write CPU profile to this file")

// setNameKey identifies the pack names of one set code in one locale.
// Konami's product ids differ per locale (the same booster has a different
// pid on the JA, KO and EN pages), so names cannot be matched across locales
// and each locale picks its own most representative name.
type setNameKey struct {
	setCode string
	locale  string
}

// setNameStat accumulates the prints that list one pack name.
type setNameStat struct {
	date       string // earliest non-empty print date
	printCount int
}

// Pack names seen on card pages, collected while crawling and written
// by writeChosenSets after every pass, so the result does not depend
// on which card page a worker goroutine happens to parse first.
var (
	setNamesMu      sync.Mutex
	setNames        = make(map[setNameKey]map[string]*setNameStat)
	setGameVersions = make(map[string]konami.YuGiOhVersion) // set code -> version of the first pass that saw it
)

// recordSetName notes that print p lists a pack name under setCode in locale.
func recordSetName(setCode string, version konami.YuGiOhVersion, locale string, p konami.CardPrint) {
	if p.SetName == "" {
		return
	}
	setNamesMu.Lock()
	defer setNamesMu.Unlock()
	if _, ok := setGameVersions[setCode]; !ok {
		setGameVersions[setCode] = version
	}
	key := setNameKey{setCode, locale}
	byName := setNames[key]
	if byName == nil {
		byName = make(map[string]*setNameStat)
		setNames[key] = byName
	}
	stat := byName[p.SetName]
	if stat == nil {
		stat = &setNameStat{}
		byName[p.SetName] = stat
	}
	stat.printCount++
	if p.Date != "" && (stat.date == "" || p.Date < stat.date) {
		stat.date = p.Date
	}
}

// writeChosenSets writes one sets row per set code:
// each locale's name is the one konami.ChooseSetName prefers,
// and release_date comes from the first locale that has a date:
// EN first for TCG sets, JA first otherwise
// (JA is the earliest release on Konami's DB, TCG sets have no JA page).
func writeChosenSets(database core.Database) {
	setNamesMu.Lock()
	defer setNamesMu.Unlock()
	rows := make(map[string]konami.KonamiSet)
	// Only some JA and KO names carry Konami's bracketed English title,
	// drop it for consistency and keep it as name_en when no EN page names the set.
	englishTitles := make(map[string]map[string]string) // set code -> bare name -> English title
	for key, byName := range setNames {
		if key.locale == "en" {
			continue
		}
		merged := make(map[string]*setNameStat)
		for name, stat := range byName {
			base, english := konami.SplitEnglishTitle(name)
			if english != "" {
				if englishTitles[key.setCode] == nil {
					englishTitles[key.setCode] = make(map[string]string)
				}
				englishTitles[key.setCode][base] = english
			}
			m := merged[base]
			if m == nil {
				m = &setNameStat{}
				merged[base] = m
			}
			m.printCount += stat.printCount
			if stat.date != "" && (m.date == "" || stat.date < m.date) {
				m.date = stat.date
			}
		}
		setNames[key] = merged
	}
	for _, locale := range []string{"ja", "ko", "en"} {
		for key, byName := range setNames {
			if key.locale != locale {
				continue
			}
			candidates := make([]konami.SetNameCandidate, 0, len(byName))
			for name, stat := range byName {
				candidates = append(candidates, konami.SetNameCandidate{
					Name: name, Date: stat.date, PrintCount: stat.printCount})
			}
			chosen := konami.ChooseSetName(candidates)
			row, ok := rows[key.setCode]
			if !ok {
				row = konami.KonamiSet{
					Abbreviation:  key.setCode,
					YuGiOhVersion: setGameVersions[key.setCode],
				}
			}
			switch locale {
			case "ja":
				row.NameJA = chosen.Name
			case "ko":
				row.NameKO = chosen.Name
			case "en":
				row.NameEN = chosen.Name
			}
			rows[key.setCode] = row
		}
	}
	for setCode, row := range rows {
		if row.NameEN == "" {
			// Only one JA or KO name can be chosen, so prefer the JA one.
			if title := englishTitles[setCode][row.NameJA]; title != "" {
				row.NameEN = title
			} else if title := englishTitles[setCode][row.NameKO]; title != "" {
				row.NameEN = title
			}
		}
		localesByPriority := []string{"ja", "ko", "en"}
		if row.YuGiOhVersion == konami.TCG {
			localesByPriority = []string{"en", "ko", "ja"}
		}
		for _, locale := range localesByPriority {
			chosenName := map[string]string{"ja": row.NameJA, "ko": row.NameKO, "en": row.NameEN}[locale]
			if stat := setNames[setNameKey{setCode, locale}][chosenName]; stat != nil && stat.date != "" {
				row.ReleaseDate = stat.date
				break
			}
		}
		rows[setCode] = row
	}
	for _, row := range rows {
		if err := database.UpsertSet(row); err != nil {
			log.Printf("error UpsertSet %v: %v", row.Abbreviation, err)
		}
	}
	log.Printf("end write sets, count=%v", len(rows))
}

// cardsMissingNameEN holds the card IDs whose "ja" pass found no English name,
// because Konami's current JA pages no longer print it.
// The "en" pass removes the cards it fills,
// and fillCardNameENFromAsiaEnglish tries the rest when isFillingNameENFromAsiaEnglish is on.
var cardsMissingNameEN = &cardIDSet{ids: make(map[konami.CardID]bool)}

// cardIDSet is a set of card IDs, safe for the crawler's worker goroutines.
type cardIDSet struct {
	mu  sync.Mutex
	ids map[konami.CardID]bool
}

func (s *cardIDSet) add(id konami.CardID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ids[id] = true
}

// remove deletes id and reports whether it was in the set.
func (s *cardIDSet) remove(id konami.CardID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ids[id] {
		return false
	}
	delete(s.ids, id)
	return true
}

// sorted returns the card IDs in numeric order.
func (s *cardIDSet) sorted() []konami.CardID {
	s.mu.Lock()
	defer s.mu.Unlock()
	ret := make([]konami.CardID, 0, len(s.ids))
	for id := range s.ids {
		ret = append(ret, id)
	}
	slices.SortFunc(ret, func(a, b konami.CardID) int { return a.Int() - b.Int() })
	return ret
}
