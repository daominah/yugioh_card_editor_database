// Package core contains business logic. Can be tested without external
// resources (database, HTTP, websocket, message queue, file, ..)
package core

import "github.com/daominah/yugioh_card_editor/pkg/konami"

// App now is just a placeholder
// (NewHandlerAPI receives this App when init, but it is not used yet)
type App struct {
	Database Database
}

// Database is the write contract used by the crawler to persist parsed
// Konami data. Implemented by pkg/driver/sqlite.
type Database interface {
	UpsertCard(c konami.Card) error
	UpsertCardRush(c konami.CardRushDuel) error
	UpsertCardPassword(cardID, password, cardName string) error
	UpsertCardText(cardID, lang string, t konami.CardLocaleText) error
	UpsertSet(set konami.KonamiSet) error
	UpsertSetCards(prints []konami.CardPrint) error
}

// DatabaseAggregate is the data source/sink contract for Aggregate. It
// reads grouped raw rows from the source tables (cards, cards_rush, set_cards
// joined with card_texts) and writes the chosen most-frequent variants into
// the canonical lookup tables (monster_attributes, monster_types, *_rush,
// rarities). Implemented by pkg/driver/sqlite.
type DatabaseAggregate interface {
	FetchEnumVariants(cardsTable, canonicalCol, textCol string) ([]konami.EnumVariantInput, error)
	FetchRarityVariants() ([]konami.RarityVariantInput, error)
	UpsertMonsterAttributes(rows []konami.MonsterAttributeRow) error
	UpsertMonsterAttributesRush(rows []konami.MonsterAttributeRushRow) error
	UpsertMonsterTypes(rows []konami.MonsterTypeRow) error
	UpsertMonsterTypesRush(rows []konami.MonsterTypeRushRow) error
	UpsertRarities(rows []konami.RarityRow) error
	UpdateRarityCardCounts() error
}
