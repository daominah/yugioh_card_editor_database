package ygocdb

import (
	"testing"

	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

func TestYgocdbDownloadData_InitMapCardsPassword(t *testing.T) {
	t.Skip("skipping this external API call to ygocdb.com, comment this line to enable the test")
	var passwordsSource core.MapCardsPasswordInitiator = &YgocdbDownloadFreshData{}
	t.Logf("testing %#v", passwordsSource)
	cardPasswords, err := passwordsSource.InitMapCardsPassword()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("len(cardPasswords): %v", len(cardPasswords))

	// Test a few known cards
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
