package main

import "testing"

func TestCardsTableRowColor(t *testing.T) {
	for _, c := range []struct {
		desc        string
		cardType    string
		cardSubtype string
		sheetRow    int
		want        string
	}{
		{"Normal Monster on an even row", "Monster", "Normal", 2, "FFFFA6"},
		{"Normal Monster on an odd row", "Monster", "Normal", 3, "FFFF6D"},
		{"Effect Monster on an even row", "Monster", "Effect", 2, "FF972F"},
		{"Link Monster on an odd row", "Monster", "Link", 3, "FF860D"},
		{"Normal Spell is green, not yellow", "Spell", "Normal", 2, "77BC65"},
		{"Trap on an odd row", "Trap", "Counter", 3, "BF819E"},
		{"unknown card type", "", "", 2, ""},
	} {
		// WHEN choosing the fill color of a card row in the spreadsheet
		got := cardsTableRowColor(c.cardType, c.cardSubtype, c.sheetRow)

		// THEN the color matches the card type and the row's even or odd shade
		if got != c.want {
			t.Errorf("%v: got %q, want %q", c.desc, got, c.want)
		}
	}
}
