package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/daominah/yugioh_card_editor/internal/core"
	"github.com/mywrap/gofast"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	projectRoot, err := gofast.GetProjectRootGit()
	if err != nil {
		log.Fatalf("error os.Getwd: %v", err)
	}
	outputPath := filepath.Join(projectRoot, "web/konami_data/konami_db.json")
	if _, err := os.Stat(outputPath); err == nil { // file existed
		backup := outputPath + ".backup"
		copyFile(outputPath, backup)
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("error os.OpenFile: %v", err)
	}
	log.Printf("output result file path: %v", outputPath)

	cardLanguage := "en"
	// cardLanguage := "ja"  // TODO: handle Japanese card text

	var httpClients []*http.Client
	isUseProxy := false
	if !isUseProxy {
		httpClients = []*http.Client{{}}
	} else {
		proxyURLs := []string{
			"http://127.0.0.1:24001",
			"http://127.0.0.1:24002",
			"http://127.0.0.1:24003",
			"http://127.0.0.1:24004",
			"http://127.0.0.1:24005",
			"http://127.0.0.1:24006",
			"http://127.0.0.1:24007",
			"http://127.0.0.1:24008",
		}
		for i := 0; i < len(proxyURLs); i++ {
			proxyUrl, err := url.Parse(proxyURLs[i])
			if err != nil {
				log.Fatalf("error url.Parse proxyURLs: %v", err)
			}
			var hc *http.Client
			if isUseProxy {
				hc = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
			} else {
				hc = &http.Client{}
			}
			httpClients = append(httpClients, hc)
		}
	}

	var result []core.Card
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	maxGoroutines := make(chan bool, 32)
	beginFetchT := time.Now()
	maxFoundCardID := 0

	// first card I know: CardID 4007: "Blue-Eyes White Dragon";
	// latest card I know: CardID 21639: from "DUelist's ADvance" English released in 2025-07-04;
	// check latest set here: https://www.db.yugioh-card.com/yugiohdb/card_list.action?clm=3&wname=CardSearch
	const cardIDMin = 4000
	const cardIDMax = 24000
	for i := cardIDMin; i < cardIDMax; i++ {
		maxGoroutines <- true
		wg.Add(1)
		go func(i int) {
			defer func() {
				<-maxGoroutines
				wg.Add(-1)
			}()
			idxHC := rand.Intn(len(httpClients))
			httpClient := httpClients[idxHC]
			cardID := core.CardID(fmt.Sprintf("%v", i))
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
				log.Printf("cardID %v not found", cardID)
				return
			}
			card := core.ParseKonamiCardHTML(bodyBs, cardID)
			if card.CardName == "" {
				log.Printf("cardID %v empty CardName: %+v", cardID, card)
				return
			}
			log.Printf("ok cardID %v\n", cardID)
			mu.Lock()
			result = append(result, card)
			if maxFoundCardID < i {
				maxFoundCardID = i
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	log.Printf("fetching Konami database duration: %v", time.Since(beginFetchT))
	// Output: 10m19s
	log.Printf("maxFoundCardID: %v", maxFoundCardID)
	// Output: 20719

	sort.Slice(result, func(i, j int) bool {
		return result[i].MiscKonamiCardID.Int() < result[j].MiscKonamiCardID.Int()
	})
	beauty, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		log.Println("error json.MarshalIndent:", err)
	}
	_, err = outputFile.Write(beauty)
	if err != nil {
		log.Println("error outputFile.Write:", err)
	}
	err = outputFile.Sync()
	if err != nil {
		log.Println("error outputFile.Sync:", err)
	}
	err = outputFile.Close()
	if err != nil {
		log.Println("error outputFile.Close:", err)
	}
	log.Printf("main returned")
	log.Printf("you probably want to run 'cmd/add_card_password' to get final output 'web/konami_data/konami_db_en.js'")
}

// copyFile overwrite if target file existed
func copyFile(sourceFullPath string, targetFullPath string) bool {
	sourceFile, err := os.Open(sourceFullPath)
	if err != nil {
		log.Printf("error os.ReadFile: %v", err)
		return false
	}
	targetFile, err := os.Create(targetFullPath)
	if err != nil {
		log.Printf("error os.Create: %v", err)
		return false
	}
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		log.Printf("error io.Copy: %v", err)
		return false
	}
	return true
}
