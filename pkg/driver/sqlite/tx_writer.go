package sqlite

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

// TxWriter funnels write operations from many goroutines through a single
// background goroutine that owns one *sql.Tx. Each Upsert* method enqueues
// the SQL it wants to run on a buffered channel; the writer goroutine drains
// the channel and executes statements serially against the tx.
//
// This pattern lets the crawler keep its existing parallel-parser layout
// (parsers call database.UpsertX as before, satisfying core.Database) while
// reducing 200k+ implicit-transaction commits to a single explicit Commit
// at the end of the run. Combined with OpenForBulkLoad's pragmas (no journal,
// no fsync, exclusive lock), the bulk crawl avoids almost all per-row I/O.
//
// Errors from individual Upsert calls are logged inside the writer goroutine
// rather than returned to the caller — the parser goroutines have nowhere
// useful to surface a write error mid-stream, and the crawler already treats
// existing UpsertX errors as log-and-continue. Commit() returns the only
// error that can fail the whole run.
type TxWriter struct {
	tx   *sql.Tx
	stmt *sql.Stmt // prepared INSERT for set_cards (the hot path)
	ops  chan func() error
	done chan error
}

// BeginTxWriter starts a new transaction on db, prepares the set_cards INSERT
// once for reuse across goroutines, and spawns the writer goroutine. The
// returned TxWriter satisfies core.Database; callers Commit() once at the end.
func (db *DB) BeginTxWriter() (*TxWriter, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return nil, fmt.Errorf("error db.Begin: %w", err)
	}
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
		tx.Rollback()
		return nil, fmt.Errorf("error tx.Prepare set_cards: %w", err)
	}
	w := &TxWriter{
		tx:   tx,
		stmt: stmt,
		ops:  make(chan func() error, 1024),
		done: make(chan error, 1),
	}
	go w.run()
	return w, nil
}

func (w *TxWriter) run() {
	for op := range w.ops {
		if err := op(); err != nil {
			log.Printf("error TxWriter op: %v", err)
		}
	}
	w.stmt.Close()
	w.done <- w.tx.Commit()
}

// Commit closes the op channel, waits for the writer goroutine to drain
// every pending op, and commits the transaction. Returns the commit error,
// if any. Must be called exactly once.
func (w *TxWriter) Commit() error {
	close(w.ops)
	return <-w.done
}

func (w *TxWriter) UpsertCard(c konami.Card) error {
	abilities := jsonSlice(c.MonsterAbilities)
	linkArrows := jsonSlice(c.MonsterLinkArrows)
	w.ops <- func() error {
		_, err := w.tx.Exec(`
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
			return fmt.Errorf("UpsertCard %v: %w", c.MiscKonamiCardID, err)
		}
		return nil
	}
	return nil
}

func (w *TxWriter) UpsertCardRush(c konami.CardRushDuel) error {
	abilities := jsonSlice(c.MonsterAbilities)
	w.ops <- func() error {
		_, err := w.tx.Exec(`
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
			return fmt.Errorf("UpsertCardRush %v: %w", c.MiscKonamiCardID, err)
		}
		return nil
	}
	return nil
}

func (w *TxWriter) UpsertCardPassword(cardID, password, cardName string) error {
	w.ops <- func() error {
		_, err := w.tx.Exec(`
            INSERT OR REPLACE INTO card_passwords
                (card_id, password, card_name)
            VALUES (?, ?, ?)`,
			cardID, password, cardName)
		if err != nil {
			return fmt.Errorf("UpsertCardPassword %v: %w", cardID, err)
		}
		return nil
	}
	return nil
}

func (w *TxWriter) UpsertCardText(cardID, lang string, t konami.CardLocaleText) error {
	w.ops <- func() error {
		_, err := w.tx.Exec(`
            INSERT OR REPLACE INTO card_texts
                ( card_id,  lang,  name,  name_katakana,  effect,  pendulum_effect,  attribute_text,  monster_type_text)
            VALUES
                (?, ?, ?, ?, ?, ?, ?, ?)`,
			cardID, lang,
			t.Name, t.NamePronunciation,
			t.Effect, t.PendulumEffect,
			t.AttributeText, t.MonsterTypeText)
		if err != nil {
			return fmt.Errorf("UpsertCardText %v/%v: %w", cardID, lang, err)
		}
		return nil
	}
	return nil
}

func (w *TxWriter) UpsertSet(s konami.KonamiSet) error {
	w.ops <- func() error {
		_, err := w.tx.Exec(`
            INSERT INTO sets ( set_code,  game_version,  release_date,  name_ja,  name_ko,  name_en)
            VALUES           (?,          ?,              ?,             ?,         ?,         ?)
            ON CONFLICT (set_code)
                DO UPDATE
                SET release_date = CASE WHEN excluded.release_date != '' THEN excluded.release_date ELSE sets.release_date END,
                    name_ja      = CASE WHEN excluded.name_ja      != '' THEN excluded.name_ja      ELSE sets.name_ja      END,
                    name_ko      = CASE WHEN excluded.name_ko      != '' THEN excluded.name_ko      ELSE sets.name_ko      END,
                    name_en      = CASE WHEN excluded.name_en      != '' THEN excluded.name_en      ELSE sets.name_en      END`,
			s.Abbreviation, string(s.YuGiOhVersion), s.ReleaseDate,
			s.NameJA, s.NameKO, s.NameEN)
		if err != nil {
			return fmt.Errorf("UpsertSet %v: %w", s.Abbreviation, err)
		}
		return nil
	}
	return nil
}

func (w *TxWriter) UpsertSetCards(prints []konami.CardPrint) error {
	if len(prints) == 0 {
		return nil
	}
	// Snapshot the slice — caller may mutate after returning.
	snapshot := append([]konami.CardPrint(nil), prints...)
	w.ops <- func() error {
		for _, p := range snapshot {
			if p.Position == "" {
				continue
			}
			if _, err := w.stmt.Exec(
				p.Position, p.SetCode(), string(p.CardID),
				p.RarityCode, p.RarityName, p.Date,
			); err != nil {
				return fmt.Errorf("UpsertSetCards %v: %w", p.Position, err)
			}
		}
		return nil
	}
	return nil
}
