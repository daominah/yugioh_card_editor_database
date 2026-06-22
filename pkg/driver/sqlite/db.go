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

//go:embed init_schema_indexes.sql
var initSchemaIndexes string

// DB wraps sql.DB with Yu-Gi-Oh card data write operations.
type DB struct {
	sql *sql.DB
}

// Open creates or opens the SQLite database at path, applies the schema
// (tables only — non-PK indexes live in init_schema_indexes.sql and are
// created on demand via CreateIndexes), and returns a ready-to-use DB.
// WAL mode is enabled; MaxOpenConns(1) serializes all writes through a
// single connection, avoiding "database is locked" errors when goroutines
// flush results concurrently.
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
	if _, err := sqlDB.Exec(initSchemaIndexes); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error init schema indexes: %w", err)
	}
	return &DB{sql: sqlDB}, nil
}

// OpenForBulkLoad opens a brand-new DB with the most aggressive performance
// settings: no rollback journal, no fsync, exclusive locking, large in-memory
// page cache. The DB is created without non-PK indexes (CreateIndexes must
// be called after the bulk load). On crash mid-load the file is in an
// undefined state and must be deleted; callers using this code path are
// expected to re-run the loader rather than recover.
//
// page_size and locking_mode must be set before any table exists, so this
// constructor is for fresh files only — callers should delete pre-existing
// db/db-wal/db-shm files first.
func OpenForBulkLoad(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("error sql.Open: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	pragmas := []string{
		// page_size and locking_mode must be set before schema.
		`PRAGMA page_size = 65536`,
		`PRAGMA locking_mode = EXCLUSIVE`,
		// No rollback journal; crash mid-write = corrupt DB.
		`PRAGMA journal_mode = OFF`,
		// Never fsync; power loss = corrupt DB.
		`PRAGMA synchronous = OFF`,
		// Temp B-trees in RAM, not a temp file.
		`PRAGMA temp_store = MEMORY`,
		// 1 GB page cache cap (-N = N KB). The cache fills lazily and
		// never exceeds the actual DB size, so this is just a ceiling.
		`PRAGMA cache_size = -1048576`,
		// 1 GB mmap cap. mmap is virtual memory; the kernel pages in only
		// what's touched, so this is also a ceiling, not an allocation.
		`PRAGMA mmap_size = 1073741824`,
		`PRAGMA foreign_keys = OFF`,
	}
	for _, p := range pragmas {
		if _, err := sqlDB.Exec(p); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("error %s: %w", p, err)
		}
	}
	if _, err := sqlDB.Exec(initSchema); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error init schema: %w", err)
	}
	return &DB{sql: sqlDB}, nil
}

// CreateIndexes runs the embedded init_schema_indexes.sql, which creates the
// non-PK indexes used by aggregate and lookup queries. Intended to be called
// after a bulk-load crawl that opened the DB with OpenForBulkLoad: building
// the indexes once at the end as a sort-then-build is dramatically cheaper
// than rebalancing per row during insert.
func (db *DB) CreateIndexes() error {
	if _, err := db.sql.Exec(initSchemaIndexes); err != nil {
		return fmt.Errorf("error CreateIndexes: %w", err)
	}
	return nil
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

// UpsertSet inserts or updates a set row. On conflict, each name_<lang>
// column is overwritten only if the new value is non-empty;
// values collected from earlier locale passes are preserved.
// The caller is expected to populate exactly one of NameJA/NameKO/NameEN per call,
// matching the source page's locale.
func (db *DB) UpsertSet(s konami.KonamiSet) error {
	_, err := db.sql.Exec(`
        INSERT INTO sets ( set_code,  game_version,  release_date,  name_ja,  name_ko,  name_en)
        VALUES           (?,          ?,              ?,             ?,         ?,         ?)
        ON CONFLICT (set_code)
            DO UPDATE
            SET release_date = CASE WHEN excluded.release_date != '' THEN excluded.release_date ELSE sets.release_date END,
                name_ja      = CASE WHEN excluded.name_ja      != '' THEN excluded.name_ja      ELSE sets.name_ja      END,
                name_ko      = CASE WHEN excluded.name_ko      != '' THEN excluded.name_ko      ELSE sets.name_ko      END,
                name_en      = CASE WHEN excluded.name_en      != '' THEN excluded.name_en      ELSE sets.name_en      END`,
		s.Abbreviation, string(s.YuGiOhVersion), s.ReleaseDate,
		s.NameJA, s.NameKO, s.NameEN,
	)
	if err != nil {
		return fmt.Errorf("error UpsertSet %v: %w", s.Abbreviation, err)
	}
	return nil
}

// UpsertSetCards writes a batch of prints in a single transaction. Each
// CardPrint carries its own CardID, so the slice can mix prints from
// different cards (the crawler passes one card's prints per call).
// rarity_name is the localized name from the source page; the language is
// recoverable from card_set_code's region suffix.
//
// On a (card_set_code, rarity_code) PK collision (same code re-inserted by a
// later locale crawl), first-write-wins: existing non-empty rarity_name and
// release_date are preserved. JA crawls first (see crawler locale order),
// so JA names stick even on EN- or KR-region multi-locale promo codes
// (WCPS, WCS, 2010..2012 promo series, ADC1, etc.). Caller is expected to
// dedup intra-page duplicates with konami.DedupCardPrints before calling
// this — the SQL path does not log conflicts.
func (db *DB) UpsertSetCards(prints []konami.CardPrint) error {
	if len(prints) == 0 {
		return nil
	}
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("error db.Begin: %w", err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT INTO set_cards
            ( card_set_code,  set_code,  card_id,  rarity_code,  rarity_name,  release_date)
        VALUES
            (?, ?, ?, ?, ?, ?)
        ON CONFLICT (card_set_code, rarity_code)
            DO UPDATE SET
                rarity_name  = CASE WHEN set_cards.rarity_name  != '' THEN set_cards.rarity_name  ELSE excluded.rarity_name  END,
                release_date = CASE WHEN set_cards.release_date != '' THEN set_cards.release_date ELSE excluded.release_date END`)
	if err != nil {
		return fmt.Errorf("error tx.Prepare: %w", err)
	}
	defer stmt.Close()
	for _, p := range prints {
		if p.Position == "" {
			continue
		}
		if _, err := stmt.Exec(
			p.Position, p.SetCode(), string(p.CardID),
			p.RarityCode, p.RarityName, p.Date,
		); err != nil {
			return fmt.Errorf("error stmt.Exec %v: %w", p.Position, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error tx.Commit: %w", err)
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
