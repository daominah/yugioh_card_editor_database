package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/driver/sqlite"
)

// outputFileName file will store a map of card set number to English card name,
// e.g. "DPKB-EN001": "Blue-Eyes White Dragon".
// The map is used by yugioh_shop_order\pkg\item\set_number_card_name_lookup.json
const outputFileName = "set_number_card_name_lookup.json"

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	db, err := sqlite.Open(filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer db.Close()

	lookup, err := db.GetMapSetNumberToCardName()
	if err != nil {
		log.Fatalf("error db.GetMapSetNumberToCardName: %v", err)
	}

	outputPath := filepath.Join(projectRoot, "cmd", "export-set-number-card-name", outputFileName)
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("error os.Create %s: %v", outputPath, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(lookup); err != nil {
		log.Fatalf("error encoder.Encode: %v", err)
	}

	log.Printf("outputfile path: %v", outputPath)
}
