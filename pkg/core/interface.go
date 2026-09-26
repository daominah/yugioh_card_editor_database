// Package core contains business logic. Can be tested without external
// resources (database, HTTP, websocket, message queue, file, ..)
package core

import "github.com/daominah/yugioh_card_editor_database/pkg/konami"

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

// DatabaseReader is the read contract for querying a fully crawled and aggregated yugioh.db.
// It is intended for use after cmd/crawl-konami-db-full
// and cmd/aggregate-type-attr-rarity have both completed,
// so all cards, card_texts, card_passwords, and set_cards rows are present.
// Implemented by pkg/driver/sqlite.
type DatabaseReader interface {
	// GetCard returns the card assembled from cards, card_texts (for lang), and card_passwords.
	// Returns a wrapped sql.ErrNoRows when cardID is not found in cards.
	GetCard(cardID konami.CardID, lang string) (konami.Card, error)
	// GetMapSetNumberToCardName returns a map from card_set_code (e.g. "LOB-001")
	// to the card's English name.
	// When the same code appears with multiple rarities, the first occurrence wins.
	GetMapSetNumberToCardName() (map[string]string, error)
	// GetCardCounts returns card counts grouped by type, subtype, monster attribute,
	// monster type, monster level, monster ATK, and monster DEF.
	// Monster-specific dimensions only count rows where card_type = 'Monster'.
	GetCardCounts() (konami.CardCounts, error)
}
