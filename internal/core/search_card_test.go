package core

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mywrap/gofast"
)

// testCardsData will be loaded from the data file in TestMain
var testCardsData []byte

func init() {
	projectRoot, err := gofast.GetProjectRootGit()
	if err != nil {
		log.Fatalf("error gofast.GetProjectRootGit: %v", err)
	}
	dataFilePath := filepath.Join(projectRoot, "web/konami_data/konami_db_en.js")
	testCardsData, err = os.ReadFile(dataFilePath)
	if err != nil {
		log.Fatalf("error os.ReadFile: %v", err)
	}
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
		if card.CardType != Monster {
			if card.CardSubtype != SpellContinuous && card.CardSubtype != TrapContinuous {
				continue
			}
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
		if card.CardType != Monster {
			if card.CardSubtype != SpellContinuous && card.CardSubtype != TrapContinuous {
				continue
			}
		}
		cardEffect := strings.ToLower(card.CardEffect)
		if strings.Contains(cardEffect, "you cannot") &&
			(strings.Contains(cardEffect, "the turn you activate this effect")) {
			continue
		}
		t.Logf("%v\n%v", card.CardName, card.CardEffect)
		t.Logf("________________________________________________")
	}
}

func TestCardDatabase_SearchCardEffect_Lingering(t *testing.T) {
	projectRoot, err := gofast.GetProjectRootGit()
	if err != nil {
		t.Fatalf("error gofast.GetProjectRootGit: %v", err)
	}
	dataFilePath := filepath.Join(projectRoot, "web/konami_data/konami_db_en.js")
	cardsData, err := os.ReadFile(dataFilePath)
	if err != nil {
		t.Fatalf("error os.ReadFile: %v", err)
	}
	cardsData = bytes.TrimPrefix(cardsData, []byte(`const CardDatabase = `))

	db, err := NewCardDatabase(cardsData)
	if err != nil {
		t.Fatalf("error NewCardDatabase: %v", err)
	}
	ids, err := db.SearchCardEffect(`cannot be negated`)
	if err != nil {
		t.Fatalf("error SearchCardEffect: %v", err)
	}
	t.Logf("len(ids): %v", len(ids))
	for _, id := range ids {
		card := db.GetCard(id)

		// custom filter that is not in the search query here:
		if strings.Contains(card.CardEffect, "except") {
			continue
		}

		t.Logf("%v\n%v", card.CardName, card.CardEffect)
		t.Logf("________________________________________________")
	}
}
