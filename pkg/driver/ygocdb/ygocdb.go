// Package ygocdb implements MapCardsPasswordInitiator interface for YGOCDB download data source
package ygocdb

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

const ygocdbURL = "https://ygocdb.com/api/v0/cards.zip"

// YgocdbDownloadFreshData implements MapCardsPasswordInitiator by downloading fresh data from YGOCDB API.
type YgocdbDownloadFreshData struct{}

// InitMapCardsPassword downloads and returns the map of cardID to CardPassword from YGOCDB API.
func (d *YgocdbDownloadFreshData) InitMapCardsPassword() (map[konami.CardID]core.CardPassword, error) {
	beginT := time.Now()
	log.Printf("begin downloading card passwords data %s", ygocdbURL)
	resp, err := http.Get(ygocdbURL)
	log.Printf("end downloading card passwords data %s, duration: %v", ygocdbURL, time.Since(beginT))
	if err != nil {
		return nil, fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("zip.NewReader: %w", err)
	}

	for _, file := range zipReader.File {
		if file.Name == "cards.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("file.Open: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("io.ReadAll cards.json: %w", err)
			}
			return core.ParseYGOCDBData(data)
		}
	}

	return nil, fmt.Errorf("cards.json not found in zip file")
}
