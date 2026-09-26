package core

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

func TestInitMapCardsPassword(t *testing.T) {
	var passwordsSource MapCardsPasswordInitiator = &YgocdbStaticData{}
	t.Logf("testing %#v", passwordsSource)
	cardPasswords, err := passwordsSource.InitMapCardsPassword()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("len(cardPasswords): %v", len(cardPasswords))
	for cid, password := range map[konami.CardID]string{
		"4007":  "89631139", // Blue-Eyes White Dragon
		"9455":  "23434538", // Maxx "C"
		"8409":  "08233522", // Ally of Justice Cycle Reader, should pad with 0 to 8 digits
		"14356": "55285840", // Time Thief Redoer
		"14876": "14532163", // Lightning Storm
		"19372": "15005145", // Centur-Ion Primera
		"20500": "42141493", // Mulcharmy Fuwalos
	} {
		if cardPasswords[cid].Password != password {
			t.Errorf("cardPasswords[%v].Password = %v, want %v", cid, cardPasswords[cid].Password, password)
		}
	}
}

func TestBuildCardNameToPassword(t *testing.T) {
	t.Skip("one-off: outputs already generated at ygocdb_card_name_to_password.json and ygocdb_card_name_to_main_or_extra.json")

	// GIVEN the static YGOCDB card password data
	cardPasswords, err := (&YgocdbStaticData{}).InitMapCardsPassword()
	if err != nil {
		t.Fatal(err)
	}

	// WHEN building the reverse map from card name to password
	nameToPassword := make(map[string]string, len(cardPasswords))
	for _, cp := range cardPasswords {
		nameToPassword[cp.CardName] = cp.Password
	}

	// THEN write the map to a JSON file
	data, err := json.MarshalIndent(nameToPassword, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := "ygocdb_card_name_to_password.json"
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d entries to %v", len(nameToPassword), outputPath)

	// GIVEN the Konami card database with card subtypes
	var cards []konami.Card
	if err := json.Unmarshal(testCardsData, &cards); err != nil {
		t.Fatalf("error json.Unmarshal cards: %v", err)
	}

	// WHEN building the map from card name to main/extra deck placement
	extraSubtypes := map[konami.CardSubtype]bool{
		konami.MonsterFusion:  true,
		konami.MonsterSynchro: true,
		konami.MonsterXyz:     true,
		konami.MonsterLink:    true,
	}
	nameToMainOrExtra := make(map[string]string, len(cards))
	for _, card := range cards {
		placement := "MAIN"
		if extraSubtypes[card.CardSubtype] {
			placement = "EXTRA"
		}
		nameToMainOrExtra[card.CardName] = placement
	}

	// THEN write the map to a JSON file
	data2, err := json.MarshalIndent(nameToMainOrExtra, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	outputPath2 := "ygocdb_card_name_to_main_or_extra.json"
	if err := os.WriteFile(outputPath2, data2, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d entries to %v", len(nameToMainOrExtra), outputPath2)
}
