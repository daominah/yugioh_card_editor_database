package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/daominah/yugioh_card_editor/internal/core"
	"github.com/daominah/yugioh_card_editor/internal/driver/ygocdb"
	"github.com/mywrap/gofast"
)

// main will read and write files in this repository:
//
// * input files:
//   - cards data from Konami database: `web/konami_data/konami_db.json`
//   - cards password: `internal\core\ygocdb_card_password.json`
//   - card set name: `internal/core/yugioh_sets.csv`
//
// * output files:
//   - `web/konami_data/konami_db_en.js`: all cards data as a JavaScript variable, for web asset
//   - `internal/core/yugioh_cards.csv`: cards data (without effect), for human view and search
func main() {
	projectRoot, err := gofast.GetProjectRootGit()
	if err != nil {
		log.Fatalf("error GetProjectRootGit: %v", err)
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

	// read the Konami sets name from CSV file
	setsNameCSVFile := filepath.Join(projectRoot, "internal/core/yugioh_sets.csv")
	csvFile, err := os.Open(setsNameCSVFile)
	if err != nil {
		log.Fatalf("error setsNameCSVFile: %v", err)
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("error csvReader.ReadAll: %v", err)
	}
	mapKonamiSetsFullName := core.UnmarshalCSVToMapSetAbbreviationToName(records)
	log.Printf("len(mapKonamiSetsFullName): %v", len(mapKonamiSetsFullName))
	if len(mapKonamiSetsFullName) == 0 {
		log.Fatalf("error empty mapKonamiSetsFullName")
	}

	// read crawled cards data from Konami database
	crawledOutputPath := filepath.Join(projectRoot, "web/konami_data/konami_db.json")
	cardsDatabaseB, err := os.ReadFile(crawledOutputPath)
	if err != nil {
		log.Fatalf("error os.ReadFile: %v", err)
	}
	var cards []core.Card
	err = json.Unmarshal(cardsDatabaseB, &cards)
	if err != nil {
		log.Fatalf("error json.Unmarshal file %v data: %v", crawledOutputPath, err)
	}
	// this log line will be captured by the GitHub action to its summary and commit message,
	// do not change it without updating the GitHub action too.
	log.Printf("len(cards): %v", len(cards))

	// add the card password based on the card ID
	for i, card := range cards {
		cards[i].MiscCardPassword = cardPasswords[card.MiscKonamiCardID]
	}
	updatedData, err := json.MarshalIndent(cards, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal updated data: %v", err)
	}

	// the final output will be a JavaScript file that has a variable named CardDatabase,
	// which contains the updated data.
	lastUpdated := time.Now().In(core.VietnamTimezone).Format(time.RFC3339)
	thisFileAction := `github.com/daominah/yugioh_card_editor/cmd/add_card_password`
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

	// write all cards data to csv too, but sort by name (instead of id in the JS file)
	outputFileShortCardsData := filepath.Join(projectRoot, "internal/core/yugioh_cards.csv")
	outputCSVFile, err := os.Create(outputFileShortCardsData)
	if err != nil {
		log.Fatalf("error creating CSV file: %v", err)
	}
	sort.Sort(core.SortCardNames(cards))
	csvWriter := csv.NewWriter(outputCSVFile)
	err = csvWriter.WriteAll(core.ToCSV(cards, mapKonamiSetsFullName))
	if err != nil {
		log.Fatalf("error csv.NewWriter.WriteAll: %v", err)
	}
	absPath, err := filepath.Abs(outputCSVFile.Name())
	if err != nil {
		log.Fatalf("error outputCSVFile absolute path: %v", err)
	}
	outputCSVFile.Close()
	log.Printf("wrote short cards data too: %v", absPath)

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
