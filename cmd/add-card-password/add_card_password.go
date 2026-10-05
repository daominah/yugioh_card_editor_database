package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "modernc.org/sqlite"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/core"
	"github.com/daominah/yugioh_card_editor_database/pkg/driver/ygocdb"
	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

// main will read and write files in this repository:
//
// * input files:
//   - cards data from Konami database: `web/konami_data/konami_db.json`
//   - cards password: `pkg/core/ygocdb_card_password.json`
//   - JA-only fields snapshot: `data/yugioh.db`
//     (CardNameEN, IsSpecialSummonOnly).
//     The daily EN crawl that produces `konami_db.json`
//     cannot source these two fields from the EN page;
//     they are read from the SQLite snapshot instead.
//     The snapshot is refreshed manually by `cmd/crawl-konami-db-full`,
//     so values can be stale,
//     and brand-new cards may be missing from the snapshot entirely
//     (left empty).
//
// * output files:
//   - `web/konami_data/konami_db_en.js`: all cards data as a JavaScript variable, for web asset
//
// The human-readable `data/yugioh_cards.csv` is written by `cmd/export-all-cards-csv`.
func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error GetProjectRootDir: %v", err)
		return
	}
	log.Printf("projectRoot: %v", projectRoot)

	// read the card cards passwords, static or fresh download
	var passwordsSource core.MapCardsPasswordInitiator
	if false {
		passwordsSource = &core.YgocdbStaticData{}
	} else {
		passwordsSource = &ygocdb.YgocdbDownloadFreshData{}
	}
	log.Printf("using passwordsSource: %#v", passwordsSource)
	cardPasswords, err := passwordsSource.InitMapCardsPassword()
	if err != nil {
		log.Fatalf("error %#v InitMapCardsPassword: %v", passwordsSource, err)
		return
	}
	log.Printf("len(cardPasswords): %v", len(cardPasswords))

	// read crawled cards data from Konami database
	crawledOutputPath := filepath.Join(projectRoot, "web/konami_data/konami_db.json")
	cardsDatabaseB, err := os.ReadFile(crawledOutputPath)
	if err != nil {
		log.Fatalf("error os.ReadFile: %v", err)
	}
	var cards []konami.Card
	err = json.Unmarshal(cardsDatabaseB, &cards)
	if err != nil {
		log.Fatalf("error json.Unmarshal file %v data: %v", crawledOutputPath, err)
	}
	// this log line will be captured by the GitHub action to its summary and commit message,
	// do not change it without updating the GitHub action too.
	log.Printf("len(cards): %v", len(cards))

	// read JA-only Card fields from the SQLite snapshot.
	// The daily EN crawl cannot source these from Konami's EN page
	// (CardNameEN: no <span> in the h1;
	// IsSpecialSummonOnly: token absent from the species line),
	// so values are taken from `data/yugioh.db`,
	// which is refreshed manually by `cmd/crawl-konami-db-full`.
	// Cards added since the last full re-crawl will be missing here
	// and keep the empty defaults from the EN crawl.
	sqliteDBPath := filepath.Join(projectRoot, "data/yugioh.db")
	sqliteDB, err := sql.Open("sqlite", sqliteDBPath)
	if err != nil {
		log.Fatalf("error sql.Open %v: %v", sqliteDBPath, err)
	}
	defer sqliteDB.Close()
	type cardJaFields struct {
		CardNameEN          string
		IsSpecialSummonOnly bool
	}
	mapCardJaFields := make(map[konami.CardID]cardJaFields)
	rows, err := sqliteDB.Query(`SELECT card_id, card_name_en, is_special_summon_only FROM cards`)
	if err != nil {
		log.Fatalf("error sqliteDB.Query: %v", err)
	}
	for rows.Next() {
		var cardID, isSpecialSummonOnly int
		var cardNameEN string
		if err := rows.Scan(&cardID, &cardNameEN, &isSpecialSummonOnly); err != nil {
			log.Fatalf("error rows.Scan: %v", err)
		}
		mapCardJaFields[konami.CardID(strconv.Itoa(cardID))] = cardJaFields{
			CardNameEN:          cardNameEN,
			IsSpecialSummonOnly: isSpecialSummonOnly != 0,
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("error rows.Err: %v", err)
	}
	rows.Close()
	log.Printf("len(mapCardJaFields): %v (from %v)", len(mapCardJaFields), sqliteDBPath)

	// add the card password and JA-only fields based on the card ID
	for i, card := range cards {
		cards[i].MiscCardPassword = cardPasswords[card.MiscKonamiCardID].Password
		if jaFields, ok := mapCardJaFields[card.MiscKonamiCardID]; ok {
			cards[i].CardNameEN = jaFields.CardNameEN
			cards[i].IsSpecialSummonOnly = jaFields.IsSpecialSummonOnly
		}
	}
	updatedData, err := json.MarshalIndent(cards, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal updated data: %v", err)
	}

	// the final output will be a JavaScript file that has a variable named CardDatabase,
	// which contains the updated data.
	lastUpdated := time.Now().In(base.VietnamTimezone).Format(time.RFC3339)
	thisFileAction := `github.com/daominah/yugioh_card_editor_database/cmd/add-card-password`
	commentLine := fmt.Sprintf("// CardDatabase was updated at %s\n// by %v\n", lastUpdated, thisFileAction)
	log.Printf("commentLine:\n%s", commentLine)

	constLine := fmt.Sprintf("const CardDatabase = %s\n", updatedData)
	finalOutputData := commentLine + constLine

	finalOutputPath := filepath.Join(projectRoot, "web/konami_data/konami_db_en.js")
	err = os.WriteFile(finalOutputPath, []byte(finalOutputData), 0644)
	if err != nil {
		log.Fatalf("Failed to write final output: %v", err)
	}
	log.Printf("SUCCESSFULLY UPDATED THE CARD DATABASE to %s", finalOutputPath)

	// write back the card passwords to the crawled JSON file
	// code here write updatedData to crawledOutputPath
	err = os.WriteFile(crawledOutputPath, updatedData, 0644)
	if err != nil {
		log.Fatalf("Failed to write updated card database: %v", err)
	} else {
		log.Printf("added passwords to crawled JSON file (gitignored): %v", crawledOutputPath)
	}

	log.Printf("main returned")
}
