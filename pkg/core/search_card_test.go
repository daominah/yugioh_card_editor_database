package core

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

// testCardsData will be loaded from the data file in TestMain
var testCardsData []byte

func init() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}
	dataFilePath := filepath.Join(projectRoot, "web/konami_data/konami_db_en.js")
	testCardsData, err = os.ReadFile(dataFilePath)
	if err != nil {
		log.Fatalf("error os.ReadFile: %v", err)
	}
	// testCardsData looks like:
	_ = `
// CardDatabase was updated at 2026-01-28T08:24:06+07:00
// by github.com/daominah/yugioh_card_editor_database/cmd/add-card-password
const CardDatabase = [
	...
]
`
	// remove comment lines and the prefix
	lines := bytes.Split(testCardsData, []byte{'\n'})
	var dataLines [][]byte
	for _, line := range lines {
		trimmedLine := bytes.TrimSpace(line)
		if len(trimmedLine) == 0 {
			continue
		}
		if bytes.HasPrefix(trimmedLine, []byte("//")) {
			continue
		}
		dataLines = append(dataLines, line)
	}
	testCardsData = bytes.Join(dataLines, []byte{'\n'})
	testCardsData = bytes.TrimPrefix(testCardsData, []byte(`const CardDatabase = `))
}

func TestCardDatabase_SearchCardEffectContain(t *testing.T) {
	db, err := NewCardDatabase(testCardsData)
	if err != nil {
		t.Fatalf("error NewCardDatabase: %v", err)
	}
	for i, c := range []struct {
		query    string
		wantName string
	}{
		{
			query:    "Virtually invincible",
			wantName: "Blue-Eyes White Dragon",
		},
		{
			query:    `invincible.*tell the tale`,
			wantName: "Blue-Eyes White Dragon",
		},
		{
			query:    "cannot be destroyed",
			wantName: "Blue-Eyes Jet Dragon",
		},
		{
			query:    `cannot.*destroy`,
			wantName: "Blue-Eyes Jet Dragon",
		},
		{
			query:    "neither player can special summon",
			wantName: "Fossil Dyna Pachycephalo",
		},
		{
			query:    `(neither.*summon)|(your opponent cannot.*summon)`,
			wantName: "Fossil Dyna Pachycephalo",
		},
	} {
		ids, err := db.SearchCardEffect(c.query)
		if err != nil {
			t.Errorf("error SearchCardEffect i %v: %v", i, err)
			continue
		}
		var cardNames []string
		found := false
		for _, id := range ids {
			cardName := db.GetCard(id).CardName
			cardNames = append(cardNames, cardName)
			if cardName == c.wantName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("test case %v: want card name %v, got %#v", i, c.wantName, cardNames)
		}
	}
}

func TestCardDatabase_SearchCardEffect_Floodgate_Summon(t *testing.T) {
	db, err := NewCardDatabase(testCardsData)
	if err != nil {
		t.Fatalf("error NewCardDatabase: %v", err)
	}
	ids, err := db.SearchCardEffect(`(neither[^.]*summon)|(your opponent cannot[^.]*summon)`)
	if err != nil {
		t.Fatalf("error SearchCardEffect: %v", err)
	}
	t.Logf("len(ids): %v", len(ids))
	for _, id := range ids {
		card := db.GetCard(id)
		if card.CardType != konami.Monster {
			continue
		}
		if card.MonsterLevelRankLink > 4 {
			continue
		}
		t.Logf("%v\n%v", card.CardName, card.CardEffect)
		t.Logf("________________________________________________")
	}
}

func TestCardDatabase_SearchCardEffect_Floodgate_Effect(t *testing.T) {
	db, err := NewCardDatabase(testCardsData)
	if err != nil {
		t.Fatalf("error NewCardDatabase: %v", err)
	}
	ids, err := db.SearchCardEffect(`(neither[^.]*activate[^.]*effect)|(cannot[^.]*activate[^.]*effect)|(negate[^.]*effect[^.]*all)|(negate[^.]*all[^.]*effect)`)
	if err != nil {
		t.Fatalf("error SearchCardEffect: %v", err)
	}
	t.Logf("len(ids): %v", len(ids))
	for _, id := range ids {
		card := db.GetCard(id)
		if card.CardType != konami.Monster {
			if card.CardSubtype != konami.SpellContinuous && card.CardSubtype != konami.TrapContinuous {
				continue
			}
		}
		cardEffect := strings.ToLower(card.CardEffect)
		if strings.Contains(cardEffect, "you cannot") &&
			(strings.Contains(cardEffect, "the turn you activate this effect")) {
			continue
		}
		//t.Logf("%v\n%v", card.CardName, card.CardEffect)
		//t.Logf("________________________________________________")
	}
}

func TestCardDatabase_SearchHandTrap(t *testing.T) {
	db, err := NewCardDatabase(testCardsData)
	if err != nil {
		t.Fatalf("error NewCardDatabase: %v", err)
	}

	queryActivationConditionOR := []string{
		`Quick Effect`,
		`during either player`,
		`When.*opponent.*declare.*attack`,
		`When.*opponent.*activate`,
	}
	queryCostOR := []string{
		`can send this card from your hand`,
		`can discard this card`,
		`Special Summon this card from your hand`,
	}
	queriesAND := []string{
		strings.Join(queryActivationConditionOR, "|"),
		strings.Join(queryCostOR, "|"),
	}

	mergeResults := make(map[konami.CardID]int)
	for _, query := range queriesAND {
		ids, err := db.SearchCardEffect(query)
		if err != nil {
			t.Fatalf("error SearchCardEffect: %v", err)
		}
		for _, id := range ids {
			if _, exists := mergeResults[id]; !exists {
				mergeResults[id] = 0
			}
			mergeResults[id]++
		}
	}
	var ids []konami.CardID
	for id := range mergeResults {
		// the following line is kind of set intersection,
		// only store cardID that match both activation condition and cost
		if mergeResults[id] == len(queriesAND) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)

	t.Logf("len(ids): %v", len(ids))
	for _, id := range ids {
		card := db.GetCard(id)
		t.Logf("%v\n%v", card.CardName, card.CardEffect)
		t.Logf("________________________________________________")
	}
}
