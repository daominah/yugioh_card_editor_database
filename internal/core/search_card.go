package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type CardDatabase struct {
	cards          map[CardID]Card   // map key is cardID, e.g. "4007" for "Blue-Eyes White Dragon"
	cardEffects    map[CardID]string // map cardID to lowercase card effect
	cardPenEffects map[CardID]string // map cardID to lowercase card pendulum effect
}

func NewCardDatabase(jsonData []byte) (*CardDatabase, error) {
	var cards []Card
	err := json.Unmarshal(jsonData, &cards)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %v", err)
	}
	d := &CardDatabase{
		cards:          make(map[CardID]Card),
		cardEffects:    make(map[CardID]string),
		cardPenEffects: make(map[CardID]string),
	}
	for _, card := range cards {
		d.cards[card.MiscKonamiCardID] = card
		d.cardEffects[card.MiscKonamiCardID] = strings.ToLower(card.CardEffect)
		if card.IsPendulum {
			d.cardPenEffects[card.MiscKonamiCardID] = strings.ToLower(card.PendulumEffect)
		}
	}
	return d, nil
}

func (d *CardDatabase) GetCard(cardID CardID) Card {
	return d.cards[cardID]
}

// SearchCardEffect searches text in card effects using regular expression,
// query is case-insensitive.
func (d *CardDatabase) SearchCardEffect(query string) ([]CardID, error) {
	query = strings.ToLower(query)
	matcher, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile: %v", err)
	}
	uniqueIDs := make(map[CardID]bool)
	for id, effect := range d.cardEffects {
		if matcher.MatchString(effect) {
			uniqueIDs[id] = true
		}
	}
	for id, effect := range d.cardPenEffects {
		if matcher.MatchString(effect) {
			uniqueIDs[id] = true
		}
	}
	var result SortCardIDs
	for id := range uniqueIDs {
		result = append(result, id)
	}
	sort.Sort(result)
	return result, nil
}

func (d *CardDatabase) SearchCardName(query string) []CardID {
	// not implemented
	return nil
}

func (d *CardDatabase) SearchCard(query string) []CardID {
	// not implemented
	return nil
}
