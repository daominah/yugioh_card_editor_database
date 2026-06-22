package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	_ "modernc.org/sqlite"
)

// Adjustable knobs (edit before running):
const outputFileName = "set_number_card_name_lookup.json"

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sql.Open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT set_cards.card_set_code, cards.card_name_en
		FROM set_cards
		JOIN cards ON set_cards.card_id = cards.card_id
		ORDER BY set_cards.card_set_code
	`)
	if err != nil {
		log.Fatalf("error db.Query: %v", err)
	}
	defer rows.Close()

	lookup := make(map[string]string)

	for rows.Next() {
		var setCode, cardNameEN string
		if err := rows.Scan(&setCode, &cardNameEN); err != nil {
			log.Fatalf("error rows.Scan: %v", err)
		}
		if _, exists := lookup[setCode]; !exists {
			lookup[setCode] = cardNameEN
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("error rows.Err: %v", err)
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
}
