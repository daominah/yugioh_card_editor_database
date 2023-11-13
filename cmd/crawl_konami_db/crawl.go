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
	"strconv"
	"sync"

	"github.com/daominah/yugioh_card_editor/internal/core"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error os.Getwd: %v", err)
	}
	outputPath := filepath.Join(pwd, "web/konami_data/konami_db.json")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("error os.OpenFile: %v", err)
	}

	cardLanguage := "en"
	//cardLanguage := "ja"  // TODO: handle Japanese card text

	isUseProxy := true

	var result []core.Card

	var httpClients []*http.Client
	proxyURLs := []string{
		"http://127.0.0.1:24001",
		"http://127.0.0.1:24002",
		"http://127.0.0.1:24003",
		"http://127.0.0.1:24004",
		"http://127.0.0.1:24005",
		"http://127.0.0.1:24006",
		"http://127.0.0.1:24007",
		"http://127.0.0.1:24008",
		"http://127.0.0.1:24009",
		"http://127.0.0.1:24010",
		"http://127.0.0.1:24011",
		"http://127.0.0.1:24012",
		"http://127.0.0.1:24013",
		"http://127.0.0.1:24014",
		"http://127.0.0.1:24015",
		"http://127.0.0.1:24016",
		"http://127.0.0.1:24017",
		"http://127.0.0.1:24018",
		"http://127.0.0.1:24019",
		"http://127.0.0.1:24020",
		"http://127.0.0.1:24021",
		"http://127.0.0.1:24022",
		"http://127.0.0.1:24023",
		"http://127.0.0.1:24024",
		"http://127.0.0.1:24025",
		"http://127.0.0.1:24026",
		"http://127.0.0.1:24027",
		"http://127.0.0.1:24028",
		"http://127.0.0.1:24029",
		"http://127.0.0.1:24030",
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

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	maxGoroutines := make(chan bool, 15)

	// first card I know: CardID 4007: "Blue-Eyes White Dragon"
	// latest card I know: CardID 19507: "Promethean Princess, Bestower of Flames"
	const maxCardID = 20000

	for i := 4000; i < maxCardID; i++ {
		//for i := 5000; i < maxCardID; i++ {
		//for i := 18792; i < maxCardID; i++ {
		if i > 19507 {
			break
		}
		maxGoroutines <- true
		wg.Add(1)
		go func(i int) {
			defer func() {
				<-maxGoroutines
				wg.Add(-1)
			}()
			idxHC := rand.Intn(len(httpClients))
			httpClient := httpClients[idxHC]
			cardID := fmt.Sprintf("%v", i)
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
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		cardI, _ := strconv.Atoi(result[i].MiscKonamiCardID)
		cardJ, _ := strconv.Atoi(result[j].MiscKonamiCardID)
		return cardI < cardJ
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
	log.Println("main returned")
}
