package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error GetProjectRootDir: %v", err)
	}
	outputPath := filepath.Join(projectRoot, "web/konami_data/konami_db.json")
	outputDir := filepath.Dir(outputPath)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Fatalf("error output directory does not exist: %v", outputDir)
	}
	log.Printf("output result file path: %v", outputPath)

	cardLanguage := "en"
	// cardLanguage := "ja"  // TODO: handle Japanese card text

	var result []konami.Card
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	maxGoroutines := make(chan bool, 32)
	maxFoundCardID := 0
	countFound := make(map[int]bool)
	countNotFound := make(map[int]bool)

	beginFetchT := time.Now()
	log.Printf("begin downloading Konami database at %v", beginFetchT.Format(time.RFC3339))
	log.Printf("================================================")
	httpClient := &http.Client{Timeout: 64 * time.Second}

	// the first CardID is 4007 "Blue-Eyes White Dragon",
	// the last CardID in 2026-01 is 22355 "Charmers of the Grand Circle",
	// source https://www.db.yugioh-card.com/yugiohdb/card_list.action?clm=3&wname=CardSearch,
	// the loop will stop when not found for 400 consecutive CardIDs,
	// the EstimatedLastCardID only for estimating remaining time.
	const FirstCardID = 4000
	const EstimatedLastCardID = 24000
	for i := FirstCardID; true; i++ {
		mu.Lock()
		currentMaxFound := maxFoundCardID
		mu.Unlock()
		if currentMaxFound > 0 && i-currentMaxFound > 400 {
			break
		}
		// log estimate speed every 1000 cardIDs
		if (i-FirstCardID)%200 == 0 && i > FirstCardID {
			elapsed := time.Since(beginFetchT)
			processed := i - FirstCardID
			avgDurFor1000Cards := elapsed * 1000 / time.Duration(processed)
			remainingCardIDs := EstimatedLastCardID - i
			estimatedRemainingDuration := avgDurFor1000Cards * time.Duration(remainingCardIDs) / 1000
			mu.Lock()
			foundCount := len(countFound)
			notFoundCount := len(countNotFound)
			mu.Unlock()
			var percentage float64
			if processed > 0 {
				percentage = float64(foundCount) / float64(processed) * 100
			}

			log.Printf("proc %5d cards in %-10s, avg speed: %-10s/Kcards, est remaining: %-10s, max cardID: %5d",
				processed, formatDuration(elapsed), formatDuration(avgDurFor1000Cards), formatDuration(estimatedRemainingDuration), currentMaxFound)
			log.Printf("found: %5d, not found: %5d, found percentage: %.2f%%", foundCount, notFoundCount, percentage)
			log.Printf("================================================")
		}
		maxGoroutines <- true
		wg.Add(1)
		go func(i int) {
			defer func() {
				<-maxGoroutines
				wg.Add(-1)
			}()
			cardID := konami.CardID(fmt.Sprintf("%v", i))
			cardURL := `https://www.db.yugioh-card.com/yugiohdb/card_search.action` +
				fmt.Sprintf(`?ope=2&request_locale=%v&cid=%v`, cardLanguage, cardID)
			w, err := httpClient.Get(cardURL)
			if err != nil {
				log.Printf("error cardID %v http.Get: %v", cardID, err)
				return
			}
			bodyBs, err := io.ReadAll(w.Body)
			if err != nil {
				log.Printf("error cardID %v io.ReadAll: %v", cardID, err)
				return
			}
			if !(200 <= w.StatusCode && w.StatusCode < 400) {
				log.Printf("error cardID %v StatusCode: %v, body: %s", cardID, w.StatusCode, bodyBs)
				return
			}
			if bytes.Contains(bodyBs, []byte("Card information not found")) {
				//log.Printf("cardID %v not found", cardID)
				mu.Lock()
				countNotFound[i] = true
				mu.Unlock()
				return
			}

			// TODO: probably should store all response HTML so we can re-parse without re-fetching

			card := konami.ParseKonamiCardHTML(bodyBs, cardID)
			if card.CardName == "" {
				log.Fatalf("cardID %v ParseKonamiCardHTML returned empty CardName", cardID)
				return
			}

			// log.Printf("ok cardID %v\n", cardID)
			mu.Lock()
			result = append(result, card)
			countFound[i] = true
			if maxFoundCardID < i {
				maxFoundCardID = i
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	log.Printf("end downloading Konami database, duration: %v", time.Since(beginFetchT))
	// Output: 7m29s
	log.Printf("maxFoundCardID: %v", maxFoundCardID)
	// Output: 22355

	sort.Slice(result, func(i, j int) bool {
		return result[i].MiscKonamiCardID.Int() < result[j].MiscKonamiCardID.Int()
	})
	beauty, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		log.Fatalf("error json.MarshalIndent: %v", err)
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("error os.Create: %v", err)
	}
	_, err = outputFile.Write(beauty)
	if err != nil {
		log.Fatalf("error outputFile.Write: %v", err)
	}
	err = outputFile.Sync()
	if err != nil {
		log.Fatalf("error outputFile.Sync: %v", err)
	}
	err = outputFile.Close()
	if err != nil {
		log.Fatalf("error outputFile.Close: %v", err)
	}
	log.Printf("main returned")
	log.Printf("you probably want to run 'cmd/add-card-password' to get final output 'web/konami_data/konami_db_en.js'")
}

// formatDuration formats duration with millisecond precision.
// If duration >= 60s, shows as "1m23.456s", otherwise "12.345s"
func formatDuration(d time.Duration) string {
	if d < 60*time.Second {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm%.3fs", minutes, seconds)
}
