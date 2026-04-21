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
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

const (
	firstTCGCardID         = 4000
	firstRushCardID        = 15000
	maxConsecutiveNotFound = 400
	maxGoroutines          = 32
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	dbLocalFile := filepath.Join(projectRoot, "yugioh.db")
	dbSQLite, err := sqlite.Open(dbLocalFile)
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer dbSQLite.Close()
	log.Printf("database: %v", dbLocalFile)
	var db core.Database = dbSQLite

	htmlCacheDir := filepath.Join(projectRoot, "html_cache")

	// TCG/OCG: JA first for widest card ID coverage, then EN to correct stats
	// that JA could not parse (attribute names use EN strings in lookup maps),
	// then KO for Korean text and KR set codes.
	for _, locale := range []string{"ja", "en", "ko"} {
		log.Printf("begin crawlLocale %v %v", konami.StandardDB, locale)
		crawlLocale(db, htmlCacheDir, konami.StandardDB, locale)
	}

	// Rush Duel / Duel Links: only JA and KO are available on rushdb.
	for _, locale := range []string{"ja", "ko"} {
		log.Printf("begin crawlLocale %v %v", konami.RushDB, locale)
		crawlLocale(db, htmlCacheDir, konami.RushDB, locale)
	}

	log.Printf("crawl-konami-db-full done")
}

func crawlLocale(database core.Database, htmlCacheDir string, konamiDB konami.KonamiDB, locale string) {
	baseURL := fmt.Sprintf("https://www.db.yugioh-card.com/%v/card_search.action", konamiDB)
	cacheDir := filepath.Join(htmlCacheDir, string(konamiDB), locale)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("error os.MkdirAll %v: %v", cacheDir, err)
	}

	firstCardID := firstTCGCardID
	if konamiDB == konami.RushDB {
		firstCardID = firstRushCardID
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

	httpClient := &http.Client{Timeout: 64 * time.Second}

	var (
		mu             sync.Mutex
		maxFoundCardID int
		wg             sync.WaitGroup
		semaphore      = make(chan struct{}, maxGoroutines)
	)

	beginT := time.Now()
	log.Printf("begin locale=%v gameVersion=%v at %v", locale, gameVersion, beginT.Format(time.RFC3339))

	for i := firstCardID; ; i++ {
		mu.Lock()
		currentMax := maxFoundCardID
		mu.Unlock()
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

			bodyBs, err := readOrFetchHTML(httpClient, cachePath, baseURL, locale, cardID)
			if err != nil {
				log.Printf("error cardID=%v locale=%v: %v", cardID, locale, err)
				return
			}
			if bytes.Contains(bodyBs, []byte("Card information not found")) {
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
			if err := database.UpsertCardText(string(cardID), locale, text); err != nil {
				log.Printf("error UpsertCardText cardID=%v locale=%v: %v", cardID, locale, err)
			}
			if konamiDB == konami.RushDB {
				if err := database.UpsertCardRush(rushCard); err != nil {
					log.Printf("error UpsertCardRush cardID=%v: %v", cardID, err)
				}
			} else {
				if err := database.UpsertCard(card); err != nil {
					log.Printf("error UpsertCard cardID=%v: %v", cardID, err)
				}
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
	log.Printf("done locale=%v maxFoundCardID=%v duration=%v", locale, maxFoundCardID, formatDuration(time.Since(beginT)))
}

// readOrFetchHTML returns HTML for cardID: from the on-disk cache if present,
// otherwise fetches from the Konami DB and writes the response to the cache.
// The cache allows re-parsing during development without re-downloading.
func readOrFetchHTML(client *http.Client, cachePath, baseURL, locale string, cardID konami.CardID) ([]byte, error) {
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	url := fmt.Sprintf("%v?ope=2&request_locale=%v&cid=%v", baseURL, locale, cardID)
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
		log.Printf("warn: failed to cache cardID=%v locale=%v: %v", cardID, locale, err)
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
