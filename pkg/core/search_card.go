package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

type CardDatabase struct {
	cards          map[konami.CardID]konami.Card // map key is cardID, e.g. "4007" for "Blue-Eyes White Dragon"
	cardEffects    map[konami.CardID]string      // map cardID to lowercase card effect
	cardPenEffects map[konami.CardID]string      // map cardID to lowercase card pendulum effect
}

func NewCardDatabase(jsonData []byte) (*CardDatabase, error) {
	var cards []konami.Card
	err := json.Unmarshal(jsonData, &cards)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	d := &CardDatabase{
		cards:          make(map[konami.CardID]konami.Card),
		cardEffects:    make(map[konami.CardID]string),
		cardPenEffects: make(map[konami.CardID]string),
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

func (d *CardDatabase) GetCard(cardID konami.CardID) konami.Card {
	return d.cards[cardID]
}

// SearchCardEffect searches text in card effects using regular expression,
// query is case-insensitive.
func (d *CardDatabase) SearchCardEffect(query string) ([]konami.CardID, error) {
	query = strings.ToLower(query)
	matcher, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile: %w", err)
	}
	uniqueIDs := make(map[konami.CardID]bool)
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
	var result konami.SortCardIDs
	for id := range uniqueIDs {
		result = append(result, id)
	}
	sort.Sort(result)
	return result, nil
}

func (d *CardDatabase) SearchCardName(query string) []konami.CardID {
	// not implemented
	return nil
}

func (d *CardDatabase) SearchCard(query string) []konami.CardID {
	// not implemented
	return nil
}
