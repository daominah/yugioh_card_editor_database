package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor/internal/core"
	"github.com/mywrap/gofast"
)

func main() {
	log.SetFlags(log.Lshortfile)

	cardPasswords, err := core.InitMapCardsPassword()
	if err != nil {
		log.Fatalf("error InitMapCardsPassword: %v", err)
		return
	}
	log.Printf("len(cardPasswords): %v", len(cardPasswords))
	crawledOutputPath := "web/konami_data/konami_db.json"
	cardsDatabaseB, err := os.ReadFile(crawledOutputPath)
	if err != nil {
		log.Fatalf("error os.ReadFile: %v", err)
	}
	var cards []core.Card
	err = json.Unmarshal(cardsDatabaseB, &cards)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON data: %v", err)
	}
	log.Printf("len(cards): %v", len(cards))

	for i, card := range cards {
		cards[i].MiscCardPassword = cardPasswords[card.MiscKonamiCardID]
	}
	updatedData, err := json.MarshalIndent(cards, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal updated data: %v", err)
	}

	// the final output will be a JavaScript file that has a variable named CardDatabase,
	// which contains the updated data.
	finalOutputData := fmt.Sprintf("const CardDatabase = %s\n", updatedData)
	projectRoot, err := gofast.GetProjectRootGit()
	finalOutputPath := filepath.Join(projectRoot, "web/konami_data/konami_db_en.js")
	err = os.WriteFile(finalOutputPath, []byte(finalOutputData), 0644)
	if err != nil {
		log.Fatalf("Failed to write final output: %v", err)
	}
	log.Printf("successfully updated the card database to %s", finalOutputPath)
	log.Printf("main returned")
}
