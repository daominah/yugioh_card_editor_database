package sqlite

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

//go:embed init_schema.sql
var initSchema string

// DB wraps sql.DB with Yu-Gi-Oh card data write operations.
type DB struct {
	sql *sql.DB
}

// Open creates or opens the SQLite database at path, applies the schema, and
// returns a ready-to-use DB. WAL mode is enabled; MaxOpenConns(1) serializes
// all writes through a single connection, avoiding "database is locked" errors
// when goroutines flush results concurrently.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("error sql.Open: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error PRAGMA journal_mode=WAL: %w", err)
	}
	if _, err := sqlDB.Exec(initSchema); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error init schema: %w", err)
	}
	return &DB{sql: sqlDB}, nil
}

func (db *DB) Close() error {
	return db.sql.Close()
}

// UpsertCard writes TCG/OCG card stats. The crawler gates this call to only
// the "ja" locale pass (see crawl-konami-db-full), so card stats are written
// once per card from the JA page and never overwritten by EN/KO crawls.
// INSERT OR REPLACE is therefore used for re-runnability of the JA pass
// itself (e.g. after a parser fix), not for cross-locale corrections.
func (db *DB) UpsertCard(c konami.Card) error {
	abilities := jsonSlice(c.MonsterAbilities)
	linkArrows := jsonSlice(c.MonsterLinkArrows)
	_, err := db.sql.Exec(`
        INSERT OR REPLACE INTO cards
            ( card_id,  card_name_en,  card_type,  card_subtype,  attribute,  monster_type,
              level_rank_link,  atk,  atk_str,  def,  def_str,
              abilities,  link_arrows,  is_pendulum,  pendulum_scale,  is_non_effect,
              is_special_summon_only,
              year,  creator)
        VALUES
            (?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?,
             ?, ?)`,
		string(c.MiscKonamiCardID), c.CardNameEN,
		string(c.CardType), string(c.CardSubtype),
		string(c.MonsterAttribute), string(c.MonsterType),
		c.MonsterLevelRankLink,
		int(c.MonsterATK), c.MonsterATKStr,
		int(c.MonsterDEF), c.MonsterDEFStr,
		abilities, linkArrows,
		boolToInt(c.IsPendulum), c.PendulumScale, boolToInt(c.IsNonEffectMonster),
		boolToInt(c.IsSpecialSummonOnly),
		c.MiscYear, c.MiscCreator,
	)
	if err != nil {
		return fmt.Errorf("error UpsertCard %v: %w", c.MiscKonamiCardID, err)
	}
	return nil
}

// UpsertCardRush writes Rush Duel / Duel Links card stats. Like UpsertCard
// the crawler gates this to the "ja" pass only; KO crawls produce no writes
// to cards_rush. The Card.CardNameEN field on the embedded struct is ignored
// here — Konami's Rush JA pages have no EN <span>, so the value is always
// empty for Rush, and cards_rush has no card_name_en column.
func (db *DB) UpsertCardRush(c konami.CardRushDuel) error {
	abilities := jsonSlice(c.MonsterAbilities)
	_, err := db.sql.Exec(`
        INSERT OR REPLACE INTO cards_rush
            ( card_id,  card_type,  card_subtype,  attribute,  monster_type,
              level_rank_link,  atk,  atk_str,  def,  def_str,
              abilities,  is_non_effect,  rush_is_legend,  rush_max_atk)
        VALUES
            (?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?)`,
		string(c.MiscKonamiCardID),
		string(c.CardType), string(c.CardSubtype),
		string(c.MonsterAttribute), string(c.MonsterType),
		c.MonsterLevelRankLink,
		int(c.MonsterATK), c.MonsterATKStr,
		int(c.MonsterDEF), c.MonsterDEFStr,
		abilities,
		boolToInt(c.IsNonEffectMonster), boolToInt(c.RushIsLegend), c.RushMaximumATK,
	)
	if err != nil {
		return fmt.Errorf("error UpsertCardRush %v: %w", c.MiscKonamiCardID, err)
	}
	return nil
}

// UpsertCardPassword writes the YGOCDB password row for one cardID.
// card_passwords is its own table because passwords come from a third-party
// source (ygocdb.com) on a different lifecycle than the Konami crawl.
func (db *DB) UpsertCardPassword(cardID, password, cardName string) error {
	_, err := db.sql.Exec(`
        INSERT OR REPLACE INTO card_passwords
            (card_id, password, card_name)
        VALUES (?, ?, ?)`,
		cardID, password, cardName,
	)
	if err != nil {
		return fmt.Errorf("error UpsertCardPassword %v: %w", cardID, err)
	}
	return nil
}

// UpsertCardText writes locale-specific name and effect text for one card.
// Called once per locale per card; rows in card_texts are keyed by
// (card_id, lang) so the three locale crawls coexist without conflict.
func (db *DB) UpsertCardText(cardID, lang string, t konami.CardLocaleText) error {
	_, err := db.sql.Exec(`
        INSERT OR REPLACE INTO card_texts
            ( card_id,  lang,  name,  name_katakana,  effect,  pendulum_effect,  attribute_text,  monster_type_text)
        VALUES
            (?, ?, ?, ?, ?, ?, ?, ?)`,
		cardID, lang,
		t.Name, t.NamePronunciation,
		t.Effect, t.PendulumEffect,
		t.AttributeText, t.MonsterTypeText,
	)
	if err != nil {
		return fmt.Errorf("error UpsertCardText %v/%v: %w", cardID, lang, err)
	}
	return nil
}

// UpsertSet inserts or updates a set row. On conflict, only the name column for
// locale is updated; names collected from other locales are preserved.
func (db *DB) UpsertSet(setCode, gameVersion, date, setName, locale string) error {
	var nameEn, nameJa, nameKo string
	switch locale {
	case "en":
		nameEn = setName
	case "ja":
		nameJa = setName
	case "ko":
		nameKo = setName
	}
	_, err := db.sql.Exec(`
        INSERT INTO sets ( set_code,  game_version,  release_date,  name_en,  name_ja,  name_ko)
        VALUES           (?,          ?,              ?,             ?,         ?,         ?)
        ON CONFLICT (set_code)
            DO UPDATE
            SET release_date = CASE WHEN excluded.release_date != '' THEN excluded.release_date ELSE sets.release_date END,
                name_en      = CASE WHEN excluded.name_en      != '' THEN excluded.name_en      ELSE sets.name_en      END,
                name_ja      = CASE WHEN excluded.name_ja      != '' THEN excluded.name_ja      ELSE sets.name_ja      END,
                name_ko      = CASE WHEN excluded.name_ko      != '' THEN excluded.name_ko      ELSE sets.name_ko      END`,
		setCode, gameVersion, date, nameEn, nameJa, nameKo,
	)
	if err != nil {
		return fmt.Errorf("error UpsertSet %v: %w", setCode, err)
	}
	return nil
}

// UpsertSetCard inserts or replaces one printed card entry in a regional set.
func (db *DB) UpsertSetCard(position, setCode, cardID string, p konami.CardPrint) error {
	_, err := db.sql.Exec(`
        INSERT OR REPLACE INTO set_cards
            ( card_set_code,  set_code,  card_id,  rarity_code,  rarity_name,  release_date)
        VALUES
            (?, ?, ?, ?, ?, ?)`,
		position, setCode, cardID,
		p.RarityCode, p.RarityName, p.Date,
	)
	if err != nil {
		return fmt.Errorf("error UpsertSetCard %v: %w", position, err)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// jsonSlice marshals a slice of string-typed values to a JSON array string.
// Returns "[]" for nil and empty slices, avoiding the JSON "null" value that
// json.Marshal produces for nil slices.
func jsonSlice[T ~string](items []T) string {
	if len(items) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(items)
	return string(b)
}
