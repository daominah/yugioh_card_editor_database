package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/driver/sqlite"
	"github.com/daominah/yugioh_card_editor/pkg/driver/ygocdb"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

const (
	firstTCGCardID         = 4007
	firstRushCardID        = 15150
	maxConsecutiveNotFound = 400
	maxGoroutines          = 8
	// Caps on successful cards per locale; 0 = full crawl. Split between
	// Standard (TCG/OCG) and Rush so each can be sampled or fully crawled
	// independently.
	nLimit     = 0
	nLimitRush = 0

	// isCacheOnly forbids every HTTP request: HTML pages must already exist in
	// data/html_cache/, and the YGOCDB password feed comes from the embedded
	// snapshot (pkg/core/ygocdb_card_password.json) instead of the live API.
	// Useful for reparsing after a parser fix, or for fully offline runs.
	isCacheOnly = false
	// Hand-maintained snapshot of Konami's latest released cardID.
	// Used by readOrFetchHTML to apply a 24h cache TTL above this threshold:
	// caches at or above are time-sensitive (a "not found" tail or a recent card
	// page may change), caches below are trusted indefinitely.
	latestKnownCardID     = 23141 // but the latest "en" is 22355
	latestKnownRushCardID = 23097

	isIncludingCardPassword = true // 3rd-party data ygocdb.com
)

// notFoundMarkers is the localized "card not found" string Konami renders inside
// the <nav id="pan_nav"> breadcrumb when a cardID has no card. The English string
// only appears on en pages, so the marker must be looked up per locale — otherwise
// ja/ko not-found pages bypass the early return, falsely advance maxFoundCardID,
// and disable the maxConsecutiveNotFound early-stop.
var notFoundMarkers = map[string]string{
	"en": "Card information not found",
	"ja": "カード情報がありません",
	"ko": "카드 정보가 없습니다",
}

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	dataDir := filepath.Join(projectRoot, "data")
	dbLocalFile := filepath.Join(dataDir, "yugioh.db")
	dbSQLite, err := sqlite.Open(dbLocalFile)
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer dbSQLite.Close()
	log.Printf("database: %v", dbLocalFile)
	var db core.Database = dbSQLite

	htmlCacheDir := filepath.Join(dataDir, "html_cache")

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

	log.Printf("crawl-konami-db-full done")
}

// populateCardPasswords fetches YGOCDB's password table and writes it to
// card_passwords. Konami's official DB does not expose the 8-digit password,
// so this is a separate step against a third-party source. When isCacheOnly
// is true the embedded snapshot at pkg/core/ygocdb_card_password.json is used
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
	for cardID, cp := range cardPasswords {
		if err := db.UpsertCardPassword(string(cardID), cp.Password, cp.CardName); err != nil {
			log.Printf("error UpsertCardPassword cardID=%v: %v", cardID, err)
		}
	}
	log.Printf("done card_passwords count=%v duration=%v",
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
		semaphore      = make(chan struct{}, maxGoroutines)
	)

	beginT := time.Now()
	log.Printf("begin locale=%v gameVersion=%v at %v", locale, gameVersion, beginT.Format(time.RFC3339))

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

		semaphore <- struct{}{}
		wg.Add(1)
		go func(cardNum int) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

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

			// Parse outside the DB lock — HTML parsing is CPU-bound with no shared state.
			text := konami.ParseCardLocaleText(bodyBs, cardID)
			prints := konami.ParseCardPrints(bodyBs, cardID)
			var card konami.Card
			var rushCard konami.CardRushDuel
			if konamiDB == konami.RushDB {
				rushCard = konami.ParseRushDuelCardHTML(bodyBs, cardID)
			} else {
				card = konami.ParseKonamiCardHTML(bodyBs, cardID)
			}

			// DB writes: the sqlite driver serializes these through MaxOpenConns(1).
			// Skip cards that parsed as empty — the page existed but the parser
			// could not extract a card_type, meaning the HTML format was unrecognised.
			// successCount is incremented only after a successful parse so nLimit
			// counts saved cards, not just pages that exist.
			//
			// Card stats (cards / cards_rush) are language-independent, so they
			// are written only during the "ja" crawl pass. The "en" and "ko"
			// passes still parse cards (to drive the empty-parse skip and
			// successCount) but do not call UpsertCard/UpsertCardRush — they
			// only contribute card_texts and set_cards rows below. The locale
			// order in main() puts "ja" first, so stats land before any later
			// locale's text/set rows are written for the same card.
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
			for _, p := range prints {
				if p.Position == "" {
					continue
				}
				setCode := positionToSetCode(p.Position)
				if err := database.UpsertSet(setCode, gameVersion, p.Date, p.SetName, locale); err != nil {
					log.Printf("error UpsertSet %v: %v", setCode, err)
				}
				if err := database.UpsertSetCard(p.Position, setCode, string(cardID), p); err != nil {
					log.Printf("error UpsertSetCard %v: %v", p.Position, err)
				}
			}
		}(i)
	}
	wg.Wait()
	log.Printf("done locale=%v maxFoundCardID=%v latestKnownID=%v duration=%v",
		locale, maxFoundCardID, latestKnownID, formatDuration(time.Since(beginT)))
	if maxFoundCardID > latestKnownID {
		log.Printf("warn: maxFoundCardID=%v exceeds latestKnownID=%v — bump the const to skip wasted re-downloads next run",
			maxFoundCardID, latestKnownID)
	}
}

// readOrFetchHTML returns HTML for cardID: from the on-disk cache if present,
// otherwise fetches from the Konami DB and writes the response to the cache.
// The cache allows re-parsing during development without re-downloading.
//
// Cache freshness rule: caches for cardNum >= latestKnownID expire after 24h
// because Konami may publish a new card at that ID (or fill a tail "not found"
// gap with a real release). Caches below the threshold are reused indefinitely.
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

// positionToSetCode extracts the set abbreviation from a full card number.
// "LOCH-JP077" -> "LOCH", "LOB-001" -> "LOB", "OP28-EN001" -> "OP28"
func positionToSetCode(position string) string {
	if i := strings.Index(position, "-"); i > 0 {
		return position[:i]
	}
	return position
}

func formatDuration(d time.Duration) string {
	if d < 60*time.Second {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm%.3fs", minutes, seconds)
}
