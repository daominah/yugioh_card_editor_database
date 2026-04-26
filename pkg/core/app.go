// Package core contains business logic. Can be tested without external
// resources (database, HTTP, websocket, message queue, file, ..)
package core

import "github.com/daominah/yugioh_card_editor/pkg/konami"

// App now is just a placeholder
// (NewHandlerAPI receives this App when init, but it is not used yet)
type App struct {
	Database Database
}

type Database interface {
	UpsertCard(c konami.Card) error
	UpsertCardRush(c konami.CardRushDuel) error
	UpsertCardPassword(cardID, password, cardName string) error
	UpsertCardText(cardID, lang string, t konami.CardLocaleText) error
	UpsertSet(setCode, gameVersion, date, setName, locale string) error
	UpsertSetCard(position, setCode, cardID string, p konami.CardPrint) error
}
