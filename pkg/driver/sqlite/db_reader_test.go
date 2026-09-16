package sqlite

import (
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

func TestGetCard(t *testing.T) {
	db := openRealDB(t)

	for _, tc := range []struct {
		desc string
		id   konami.CardID
		want konami.Card
	}{
		{
			desc: "CardID 4007 Blue-Eyes White Dragon",
			id:   "4007",
			want: konami.Card{
				CardName:             "Blue-Eyes White Dragon",
				CardNameEN:           "Blue-Eyes White Dragon",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterNormal,
				MonsterAttribute:     konami.LIGHT,
				MonsterType:          konami.Dragon,
				MonsterLevelRankLink: 8,
				MonsterATK:           3000,
				MonsterATKStr:        "3000",
				MonsterDEF:           2500,
				MonsterDEFStr:        "2500",
				IsNonEffectMonster:   true,
				MiscKonamiCardID:     "4007",
				MiscCardPassword:     "89631139",
			},
		},
		{
			// password has a zero prefix to verify it is stored as text, not int
			desc: "CardID 8409 Ally of Justice Cycle Reader",
			id:   "8409",
			want: konami.Card{
				CardName:             "Ally of Justice Cycle Reader",
				CardNameEN:           "Ally of Justice Cycle Reader",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterEffect,
				MonsterAttribute:     konami.DARK,
				MonsterType:          konami.Machine,
				MonsterLevelRankLink: 3,
				MonsterATK:           1000,
				MonsterATKStr:        "1000",
				MonsterDEF:           1000,
				MonsterDEFStr:        "1000",
				MonsterAbilities:     []konami.MonsterAbility{konami.Tuner},
				MiscKonamiCardID:     "8409",
				MiscCardPassword:     "08233522",
			},
		},
		{
			desc: "CardID 11232 Shaddoll Falco",
			id:   "11232",
			want: konami.Card{
				CardName:             "Shaddoll Falco",
				CardNameEN:           "Shaddoll Falco",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterEffect,
				MonsterAttribute:     konami.DARK,
				MonsterType:          konami.Spellcaster,
				MonsterLevelRankLink: 2,
				MonsterATK:           600,
				MonsterATKStr:        "600",
				MonsterDEF:           1400,
				MonsterDEFStr:        "1400",
				MonsterAbilities:     []konami.MonsterAbility{konami.Flip, konami.Tuner},
				MiscKonamiCardID:     "11232",
				MiscCardPassword:     "37445295",
			},
		},
		{
			desc: "CardID 11353 Qliphort Scout",
			id:   "11353",
			want: konami.Card{
				CardName:             "Qliphort Scout",
				CardNameEN:           "Qliphort Scout",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterNormal,
				MonsterAttribute:     konami.EARTH,
				MonsterType:          konami.Machine,
				MonsterLevelRankLink: 5,
				MonsterATK:           1000,
				MonsterATKStr:        "1000",
				MonsterDEF:           2800,
				MonsterDEFStr:        "2800",
				IsNonEffectMonster:   true,
				IsPendulum:           true,
				PendulumScale:        9,
				MiscKonamiCardID:     "11353",
				MiscCardPassword:     "65518099",
			},
		},
		{
			desc: "CardID 12788 Zoodiac Drident",
			id:   "12788",
			want: konami.Card{
				CardName:             "Zoodiac Drident",
				CardNameEN:           "Zoodiac Drident",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterXyz,
				MonsterAttribute:     konami.EARTH,
				MonsterType:          konami.BeastWarrior,
				MonsterLevelRankLink: 4,
				MonsterATK:           0,
				MonsterATKStr:        "?",
				MonsterDEF:           0,
				MonsterDEFStr:        "?",
				MiscKonamiCardID:     "12788",
				MiscCardPassword:     "48905153",
			},
		},
		{
			desc: "CardID 12953 Supreme King Z-ARC",
			id:   "12953",
			want: konami.Card{
				CardName:             "Supreme King Z-ARC",
				CardNameEN:           "Supreme King Z-ARC",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterFusion,
				MonsterAttribute:     konami.DARK,
				MonsterType:          konami.Dragon,
				MonsterLevelRankLink: 12,
				MonsterATK:           4000,
				MonsterATKStr:        "4000",
				MonsterDEF:           4000,
				MonsterDEFStr:        "4000",
				IsPendulum:           true,
				PendulumScale:        1,
				MiscKonamiCardID:     "12953",
				MiscCardPassword:     "13331639",
			},
		},
		{
			desc: "CardID 14491 Monk of the Tenyi",
			id:   "14491",
			want: konami.Card{
				CardName:             "Monk of the Tenyi",
				CardNameEN:           "Monk of the Tenyi",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterLink,
				MonsterAttribute:     konami.EARTH,
				MonsterType:          konami.Wyrm,
				MonsterLevelRankLink: 1,
				MonsterATK:           1000,
				MonsterATKStr:        "1000",
				MonsterDEFStr:        "-",
				MonsterLinkArrows:    []konami.MonsterLinkArrow{konami.Down},
				IsNonEffectMonster:   true,
				MiscKonamiCardID:     "14491",
				MiscCardPassword:     "32519092",
			},
		},
		{
			desc: "CardID 19375 Centur-Ion Legatia",
			id:   "19375",
			want: konami.Card{
				CardName:             "Centur-Ion Legatia",
				CardNameEN:           "Centur-Ion Legatia",
				CardType:             konami.Monster,
				CardSubtype:          konami.MonsterSynchro,
				MonsterAttribute:     konami.LIGHT,
				MonsterType:          konami.Machine,
				MonsterLevelRankLink: 12,
				MonsterATK:           3500,
				MonsterATKStr:        "3500",
				MonsterDEF:           2000,
				MonsterDEFStr:        "2000",
				MiscKonamiCardID:     "19375",
				MiscCardPassword:     "15982593",
			},
		},
		{
			desc: "CardID 22089 The Fallen & The Virtuous",
			id:   "22089",
			want: konami.Card{
				CardName:         "The Fallen & The Virtuous",
				CardNameEN:       "The Fallen & The Virtuous",
				CardType:         konami.Spell,
				CardSubtype:      konami.SpellQuickPlay,
				MiscKonamiCardID: "22089",
				MiscCardPassword: "30271097",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// WHEN reading the card for the EN locale from the crawled DB
			got, err := db.GetCard(tc.id, "en")
			if err != nil {
				t.Fatalf("error GetCard: %v", err)
			}

			// THEN all key fields match the published Konami data
			checkGetCard(t, got, tc.want)
		})
	}
}

func TestGetCardMissingTextAndPassword(t *testing.T) {
	db := openTestDB(t)

	// GIVEN card 4007 is inserted with no card_texts or card_passwords rows
	if err := db.UpsertCard(konami.Card{
		CardNameEN:       "Blue-Eyes White Dragon",
		CardType:         konami.Monster,
		CardSubtype:      konami.MonsterNormal,
		MonsterAttribute: konami.LIGHT,
		MonsterType:      konami.Dragon,
		MiscKonamiCardID: "4007",
	}); err != nil {
		t.Fatalf("error UpsertCard: %v", err)
	}

	// WHEN reading the card for the EN locale
	got, err := db.GetCard("4007", "en")
	if err != nil {
		t.Fatalf("error GetCard: %v", err)
	}

	// THEN card stats from the cards row are present, text and password are empty
	if got.CardNameEN != "Blue-Eyes White Dragon" {
		t.Errorf("CardNameEN got %q, want %q", got.CardNameEN, "Blue-Eyes White Dragon")
	}
	if got.CardName != "" {
		t.Errorf("CardName got %q, want empty (no card_texts row)", got.CardName)
	}
	if got.MiscCardPassword != "" {
		t.Errorf("MiscCardPassword got %q, want empty (no card_passwords row)", got.MiscCardPassword)
	}
}

func TestGetCardNotFound(t *testing.T) {
	db := openTestDB(t)

	// WHEN reading a card ID that does not exist in cards
	_, err := db.GetCard("9999", "en")

	// THEN an error wrapping sql.ErrNoRows is returned
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("error got %v, want wrapping sql.ErrNoRows", err)
	}
}

func TestGetMapSetNumberToCardName(t *testing.T) {
	db := openRealDB(t)

	// WHEN fetching the full set number to card name map from the crawled DB
	got, err := db.GetMapSetNumberToCardName()
	if err != nil {
		t.Fatalf("error GetMapSetNumberToCardName: %v", err)
	}

	// THEN known set codes resolve to the correct card names
	if got["RATE-EN053"] != "Zoodiac Drident" {
		t.Errorf("RATE-EN053 got %q, want %q", got["RATE-EN053"], "Zoodiac Drident")
	}
	if got["NECH-EN021"] != "Qliphort Scout" {
		t.Errorf("NECH-EN021 got %q, want %q", got["NECH-EN021"], "Qliphort Scout")
	}
}

func TestGetCardCounts(t *testing.T) {
	db := openRealDB(t)

	// WHEN counting cards from the fully crawled DB
	got, err := db.GetCardCounts()
	if err != nil {
		t.Fatalf("error GetCardCounts: %v", err)
	}

	// THEN each top-level card type is present
	if got.ByType[konami.Monster] == 0 {
		t.Errorf("ByType[Monster] got 0, want > 0")
	}
	if got.ByType[konami.Spell] == 0 {
		t.Errorf("ByType[Spell] got 0, want > 0")
	}
	if got.ByType[konami.Trap] == 0 {
		t.Errorf("ByType[Trap] got 0, want > 0")
	}

	// THEN known monster subtypes are present
	if got.BySubtype[konami.MonsterNormal] == 0 {
		t.Errorf("BySubtype[MonsterNormal] got 0, want > 0")
	}
	if got.BySubtype[konami.MonsterEffect] == 0 {
		t.Errorf("BySubtype[MonsterEffect] got 0, want > 0")
	}

	// THEN known monster attributes are present
	if got.ByMonsterAttribute[konami.DARK] == 0 {
		t.Errorf("ByMonsterAttribute[DARK] got 0, want > 0")
	}
	if got.ByMonsterAttribute[konami.LIGHT] == 0 {
		t.Errorf("ByMonsterAttribute[LIGHT] got 0, want > 0")
	}

	// THEN known monster types are present
	if got.ByMonsterType[konami.Dragon] == 0 {
		t.Errorf("ByMonsterType[Dragon] got 0, want > 0")
	}
	if got.ByMonsterType[konami.Warrior] == 0 {
		t.Errorf("ByMonsterType[Warrior] got 0, want > 0")
	}

	// THEN the most common monster level (4) is present
	if got.ByMonsterLevel[4] == 0 {
		t.Errorf("ByMonsterLevel[4] got 0, want > 0")
	}

	// THEN ATK 1500 is present (many well-known monsters have 1500 ATK)
	if got.ByMonsterATK[1500] == 0 {
		t.Errorf("ByMonsterATK[1500] got 0, want > 0")
	}

	// THEN DEF 2000 is present (many well-known monsters have 2000 DEF)
	if got.ByMonsterDEF[2000] == 0 {
		t.Errorf("ByMonsterDEF[2000] got 0, want > 0")
	}
}

// checkGetCard compares the key fields of two konami.Card values.
// Slice fields (MonsterAbilities, MonsterLinkArrows) treat nil and empty as equal.
func checkGetCard(t *testing.T, got, want konami.Card) {
	t.Helper()
	if got.CardName != want.CardName {
		t.Errorf("CardName got %q, want %q", got.CardName, want.CardName)
	}
	if got.CardNameEN != want.CardNameEN {
		t.Errorf("CardNameEN got %q, want %q", got.CardNameEN, want.CardNameEN)
	}
	if got.CardType != want.CardType {
		t.Errorf("CardType got %q, want %q", got.CardType, want.CardType)
	}
	if got.CardSubtype != want.CardSubtype {
		t.Errorf("CardSubtype got %q, want %q", got.CardSubtype, want.CardSubtype)
	}
	if got.MonsterAttribute != want.MonsterAttribute {
		t.Errorf("MonsterAttribute got %q, want %q", got.MonsterAttribute, want.MonsterAttribute)
	}
	if got.MonsterType != want.MonsterType {
		t.Errorf("MonsterType got %q, want %q", got.MonsterType, want.MonsterType)
	}
	if got.MonsterLevelRankLink != want.MonsterLevelRankLink {
		t.Errorf("MonsterLevelRankLink got %d, want %d", got.MonsterLevelRankLink, want.MonsterLevelRankLink)
	}
	if got.MonsterATK != want.MonsterATK {
		t.Errorf("MonsterATK got %v, want %v", got.MonsterATK, want.MonsterATK)
	}
	if got.MonsterATKStr != want.MonsterATKStr {
		t.Errorf("MonsterATKStr got %q, want %q", got.MonsterATKStr, want.MonsterATKStr)
	}
	if got.MonsterDEF != want.MonsterDEF {
		t.Errorf("MonsterDEF got %v, want %v", got.MonsterDEF, want.MonsterDEF)
	}
	if got.MonsterDEFStr != want.MonsterDEFStr {
		t.Errorf("MonsterDEFStr got %q, want %q", got.MonsterDEFStr, want.MonsterDEFStr)
	}
	if !slices.Equal(got.MonsterAbilities, want.MonsterAbilities) {
		t.Errorf("MonsterAbilities got %v, want %v", got.MonsterAbilities, want.MonsterAbilities)
	}
	if !slices.Equal(got.MonsterLinkArrows, want.MonsterLinkArrows) {
		t.Errorf("MonsterLinkArrows got %v, want %v", got.MonsterLinkArrows, want.MonsterLinkArrows)
	}
	if got.IsNonEffectMonster != want.IsNonEffectMonster {
		t.Errorf("IsNonEffectMonster got %v, want %v", got.IsNonEffectMonster, want.IsNonEffectMonster)
	}
	if got.IsPendulum != want.IsPendulum {
		t.Errorf("IsPendulum got %v, want %v", got.IsPendulum, want.IsPendulum)
	}
	if got.PendulumScale != want.PendulumScale {
		t.Errorf("PendulumScale got %d, want %d", got.PendulumScale, want.PendulumScale)
	}
	if got.MiscKonamiCardID != want.MiscKonamiCardID {
		t.Errorf("MiscKonamiCardID got %q, want %q", got.MiscKonamiCardID, want.MiscKonamiCardID)
	}
	if got.MiscCardPassword != want.MiscCardPassword {
		t.Errorf("MiscCardPassword got %q, want %q", got.MiscCardPassword, want.MiscCardPassword)
	}
}
