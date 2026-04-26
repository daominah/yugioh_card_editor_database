package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

func TestUpsertCard(t *testing.T) {
	db := openTestDB(t)

	// WHEN upserting a TCG/OCG monster card with all fields populated
	card := konami.Card{
		CardName:             "青眼の白龍",
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
		MonsterAbilities:     nil,
		MonsterLinkArrows:    nil,
		IsPendulum:           false,
		PendulumScale:        0,
		IsNonEffectMonster:   true,
		MiscKonamiCardID:     "4007",
		MiscCardPassword:     "89631139",
		MiscYear:             "2002",
		MiscCreator:          "",
	}
	err := db.UpsertCard(card)
	if err != nil {
		t.Fatalf("error UpsertCard: %v", err)
	}

	// THEN the card row can be read back with matching values
	var gotNameEN, gotType, gotAttr, gotYear string
	var gotLevel, gotATK int
	err = db.sql.QueryRow(
		`SELECT card_name_en, card_type, attribute, level_rank_link, atk, year FROM cards WHERE card_id = ?`,
		"4007",
	).Scan(&gotNameEN, &gotType, &gotAttr, &gotLevel, &gotATK, &gotYear)
	if err != nil {
		t.Fatalf("error SELECT cards: %v", err)
	}
	if gotNameEN != "Blue-Eyes White Dragon" {
		t.Errorf("card_name_en got %q, want %q", gotNameEN, "Blue-Eyes White Dragon")
	}
	if gotType != "Monster" {
		t.Errorf("card_type got %q, want %q", gotType, "Monster")
	}
	if gotAttr != "LIGHT" {
		t.Errorf("attribute got %q, want %q", gotAttr, "LIGHT")
	}
	if gotLevel != 8 {
		t.Errorf("level_rank_link got %d, want %d", gotLevel, 8)
	}
	if gotATK != 3000 {
		t.Errorf("atk got %d, want %d", gotATK, 3000)
	}
	if gotYear != "2002" {
		t.Errorf("year got %q, want %q", gotYear, "2002")
	}
}

func TestUpsertCardRush(t *testing.T) {
	db := openTestDB(t)

	// WHEN upserting a Rush Duel monster card
	card := konami.CardRushDuel{
		Card: konami.Card{
			CardName:             "Blue-Eyes White Dragon",
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
			MiscKonamiCardID:     "15150",
		},
		RushIsLegend: true,
	}
	err := db.UpsertCardRush(card)
	if err != nil {
		t.Fatalf("error UpsertCardRush: %v", err)
	}

	// THEN the card row can be read back
	var gotLevel int
	var gotIsLegend int
	err = db.sql.QueryRow(
		`SELECT level_rank_link, rush_is_legend FROM cards_rush WHERE card_id = ?`,
		"15150",
	).Scan(&gotLevel, &gotIsLegend)
	if err != nil {
		t.Fatalf("error SELECT cards_rush: %v", err)
	}
	if gotLevel != 8 {
		t.Errorf("level_rank_link got %d, want %d", gotLevel, 8)
	}
	if gotIsLegend != 1 {
		t.Errorf("rush_is_legend got %d, want %d", gotIsLegend, 1)
	}
}

func TestUpsertCardPassword(t *testing.T) {
	db := openTestDB(t)

	// WHEN upserting a YGOCDB password row
	err := db.UpsertCardPassword("4007", "89631139", "Blue-Eyes White Dragon")
	if err != nil {
		t.Fatalf("error UpsertCardPassword: %v", err)
	}

	// THEN the row round-trips and is keyed by card_id
	var gotPassword, gotName string
	err = db.sql.QueryRow(
		`SELECT password, card_name FROM card_passwords WHERE card_id = ?`, "4007",
	).Scan(&gotPassword, &gotName)
	if err != nil {
		t.Fatalf("error SELECT card_passwords: %v", err)
	}
	if gotPassword != "89631139" {
		t.Errorf("password got %q, want %q", gotPassword, "89631139")
	}
	if gotName != "Blue-Eyes White Dragon" {
		t.Errorf("card_name got %q, want %q", gotName, "Blue-Eyes White Dragon")
	}

	// AND a re-upsert of the same card_id replaces the existing row
	err = db.UpsertCardPassword("4007", "89631139", "Blue-Eyes White Dragon (updated)")
	if err != nil {
		t.Fatalf("error second UpsertCardPassword: %v", err)
	}
	var n int
	_ = db.sql.QueryRow(`SELECT COUNT(*) FROM card_passwords WHERE card_id = ?`, "4007").Scan(&n)
	if n != 1 {
		t.Errorf("after re-upsert got %d rows, want 1", n)
	}
}

func TestUpsertCardText(t *testing.T) {
	db := openTestDB(t)

	// WHEN upserting JA locale text with katakana pronunciation and a localized type
	text := konami.CardLocaleText{
		Name:              "青眼の白龍",
		NamePronunciation: "ブルーアイズ・ホワイト・ドラゴン",
		Effect:            "高い攻撃力を誇る伝説のドラゴン。",
		AttributeText:     "光属性",
		MonsterTypeText:   "ドラゴン族",
	}
	err := db.UpsertCardText("4007", "ja", text)
	if err != nil {
		t.Fatalf("error UpsertCardText: %v", err)
	}

	// THEN the per-locale text columns round-trip
	var gotName, gotKatakana, gotAttrText, gotTypeText string
	err = db.sql.QueryRow(
		`SELECT name, name_katakana, attribute_text, monster_type_text FROM card_texts WHERE card_id = ? AND lang = ?`,
		"4007", "ja",
	).Scan(&gotName, &gotKatakana, &gotAttrText, &gotTypeText)
	if err != nil {
		t.Fatalf("error SELECT card_texts: %v", err)
	}
	if gotName != "青眼の白龍" {
		t.Errorf("name got %q, want %q", gotName, "青眼の白龍")
	}
	if gotKatakana != "ブルーアイズ・ホワイト・ドラゴン" {
		t.Errorf("name_katakana got %q, want %q", gotKatakana, "ブルーアイズ・ホワイト・ドラゴン")
	}
	if gotAttrText != "光属性" {
		t.Errorf("attribute_text got %q, want %q", gotAttrText, "光属性")
	}
	if gotTypeText != "ドラゴン族" {
		t.Errorf("monster_type_text got %q, want %q", gotTypeText, "ドラゴン族")
	}
}

func TestUpsertSetAndSetCard(t *testing.T) {
	db := openTestDB(t)

	// WHEN upserting a set from JA locale, then updating it from EN locale
	err := db.UpsertSet("LOB", "OCG", "2002-02-04", "青眼の白龍伝説", "ja")
	if err != nil {
		t.Fatalf("error UpsertSet JA: %v", err)
	}
	err = db.UpsertSet("LOB", "TCG", "2002-03-08", "Legend of Blue Eyes White Dragon", "en")
	if err != nil {
		t.Fatalf("error UpsertSet EN: %v", err)
	}

	// THEN both locale names are preserved
	var gotNameEN, gotNameJA string
	err = db.sql.QueryRow(
		`SELECT name_en, name_ja FROM sets WHERE set_code = ?`, "LOB",
	).Scan(&gotNameEN, &gotNameJA)
	if err != nil {
		t.Fatalf("error SELECT sets: %v", err)
	}
	if gotNameEN != "Legend of Blue Eyes White Dragon" {
		t.Errorf("name_en got %q, want %q", gotNameEN, "Legend of Blue Eyes White Dragon")
	}
	if gotNameJA != "青眼の白龍伝説" {
		t.Errorf("name_ja got %q, want %q", gotNameJA, "青眼の白龍伝説")
	}

	// WHEN upserting a set_card entry
	p := konami.CardPrint{
		Date:       "2002-02-04",
		Position:   "LOB-001",
		SetName:    "Legend of Blue Eyes White Dragon",
		RarityCode: "UL",
		RarityName: "Ultimate Rare",
	}
	err = db.UpsertSetCard("LOB-001", "LOB", "4007", p)
	if err != nil {
		t.Fatalf("error UpsertSetCard: %v", err)
	}

	// THEN the card_set_code column stores the full card number
	var gotCardID, gotRarityCode string
	err = db.sql.QueryRow(
		`SELECT card_id, rarity_code FROM set_cards WHERE card_set_code = ?`, "LOB-001",
	).Scan(&gotCardID, &gotRarityCode)
	if err != nil {
		t.Fatalf("error SELECT set_cards: %v", err)
	}
	if gotCardID != "4007" {
		t.Errorf("card_id got %q, want %q", gotCardID, "4007")
	}
	if gotRarityCode != "UL" {
		t.Errorf("rarity_code got %q, want %q", gotRarityCode, "UL")
	}
}

func openTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("error Open: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})
	return db
}
