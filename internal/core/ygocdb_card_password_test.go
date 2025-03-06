package core

import (
	"testing"
)

func TestInitMapCardsPassword(t *testing.T) {
	cardPasswords, err := InitMapCardsPassword()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("len(cardPasswords): %v", len(cardPasswords))
	for cid, password := range map[CardID]string{
		"4007":  "89631139", // Blue-Eyes White Dragon
		"9455":  "23434538", // Maxx "C"
		"14356": "55285840", // Time Thief Redoer
		"14876": "14532163", // Lightning Storm
		"19372": "15005145", // Centur-Ion Primera
		"20500": "42141493", // Mulcharmy Fuwalos
	} {
		if cardPasswords[cid] != password {
			t.Errorf("cardPasswords[%v] = %v, want %v", cid, cardPasswords[cid], password)
		}
	}
}
