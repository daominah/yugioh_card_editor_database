package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

// GetCard returns the card assembled from three tables:
//   - cards: stats written by the JA crawl pass
//   - card_texts: locale text for the requested lang
//     (LEFT JOIN: missing rows leave CardName, CardEffect, and PendulumEffect empty)
//   - card_passwords: 8-digit password from ygocdb.com
//     (LEFT JOIN: missing rows leave MiscCardPassword empty)
//
// MiscKonamiSet and CardArt are not stored in the DB and are left empty.
// Returns a wrapped sql.ErrNoRows when cardID is not found in cards.
func (db *DB) GetCard(cardID konami.CardID, lang string) (konami.Card, error) {
	var (
		cardNameEN, cardType, cardSubtype string
		attribute, monsterType            string
		levelRankLink, atk, def           int
		atkStr, defStr                    string
		abilitiesJSON, linkArrowsJSON     string
		isPendulum, pendulumScale         int
		isNonEffect, isSpecialSummonOnly  int
		year, creator                     string
		name, effect, pendulumEffect      string
		password                          string
	)
	err := db.sql.QueryRow(`
        SELECT
            c.card_name_en, c.card_type, c.card_subtype,
            c.attribute, c.monster_type, c.level_rank_link,
            c.atk, c.atk_str, c.def, c.def_str,
            c.abilities, c.link_arrows,
            c.is_pendulum, c.pendulum_scale, c.is_non_effect, c.is_special_summon_only,
            c.year, c.creator,
            COALESCE(t.name, ''), COALESCE(t.effect, ''), COALESCE(t.pendulum_effect, ''),
            COALESCE(p.password, '')
        FROM cards c
        LEFT JOIN card_texts t ON t.card_id = c.card_id AND t.lang = ?
        LEFT JOIN card_passwords p ON p.card_id = c.card_id
        WHERE c.card_id = ?`,
		lang, string(cardID),
	).Scan(
		&cardNameEN, &cardType, &cardSubtype,
		&attribute, &monsterType, &levelRankLink,
		&atk, &atkStr, &def, &defStr,
		&abilitiesJSON, &linkArrowsJSON,
		&isPendulum, &pendulumScale, &isNonEffect, &isSpecialSummonOnly,
		&year, &creator,
		&name, &effect, &pendulumEffect,
		&password,
	)
	if err != nil {
		return konami.Card{}, fmt.Errorf("error GetCard %v: %w", cardID, err)
	}

	var abilities []konami.MonsterAbility
	if err := json.Unmarshal([]byte(abilitiesJSON), &abilities); err != nil {
		abilities = nil
	}
	var linkArrows []konami.MonsterLinkArrow
	if err := json.Unmarshal([]byte(linkArrowsJSON), &linkArrows); err != nil {
		linkArrows = nil
	}

	return konami.Card{
		CardName:             name,
		CardNameEN:           cardNameEN,
		CardType:             konami.CardType(cardType),
		CardSubtype:          konami.CardSubtype(cardSubtype),
		CardEffect:           effect,
		MonsterAttribute:     konami.MonsterAttribute(attribute),
		MonsterType:          konami.MonsterType(monsterType),
		MonsterLevelRankLink: levelRankLink,
		MonsterATK:           float64(atk),
		MonsterATKStr:        atkStr,
		MonsterDEF:           float64(def),
		MonsterDEFStr:        defStr,
		MonsterAbilities:     abilities,
		MonsterLinkArrows:    linkArrows,
		IsNonEffectMonster:   isNonEffect != 0,
		IsSpecialSummonOnly:  isSpecialSummonOnly != 0,
		IsPendulum:           isPendulum != 0,
		PendulumScale:        pendulumScale,
		PendulumEffect:       pendulumEffect,
		MiscKonamiCardID:     cardID,
		MiscCardPassword:     password,
		MiscYear:             year,
		MiscCreator:          creator,
	}, nil
}

// GetCardCounts returns card counts grouped across seven dimensions.
// Seven GROUP BY queries are issued; each returns (key, count) pairs.
func (db *DB) GetCardCounts() (konami.CardCounts, error) {
	result := konami.CardCounts{
		ByType:             make(map[konami.CardType]int),
		BySubtype:          make(map[konami.CardSubtype]int),
		ByMonsterAttribute: make(map[konami.MonsterAttribute]int),
		ByMonsterType:      make(map[konami.MonsterType]int),
		ByMonsterLevel:     make(map[int]int),
		ByMonsterATK:       make(map[int]int),
		ByMonsterDEF:       make(map[int]int),
	}

	type query struct {
		sql  string
		scan func(rows *sql.Rows) error
	}
	queries := []query{
		{
			sql: `SELECT card_type, COUNT(*) FROM cards GROUP BY card_type`,
			scan: func(rows *sql.Rows) error {
				var key string
				var count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByType[konami.CardType(key)] = count
				return nil
			},
		},
		{
			sql: `SELECT card_subtype, COUNT(*) FROM cards GROUP BY card_subtype`,
			scan: func(rows *sql.Rows) error {
				var key string
				var count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.BySubtype[konami.CardSubtype(key)] = count
				return nil
			},
		},
		{
			sql: `SELECT attribute, COUNT(*) FROM cards WHERE card_type = 'Monster' GROUP BY attribute`,
			scan: func(rows *sql.Rows) error {
				var key string
				var count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByMonsterAttribute[konami.MonsterAttribute(key)] = count
				return nil
			},
		},
		{
			sql: `SELECT monster_type, COUNT(*) FROM cards WHERE card_type = 'Monster' GROUP BY monster_type`,
			scan: func(rows *sql.Rows) error {
				var key string
				var count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByMonsterType[konami.MonsterType(key)] = count
				return nil
			},
		},
		{
			sql: `SELECT level_rank_link, COUNT(*) FROM cards WHERE card_type = 'Monster' GROUP BY level_rank_link`,
			scan: func(rows *sql.Rows) error {
				var key, count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByMonsterLevel[key] = count
				return nil
			},
		},
		{
			sql: `SELECT atk, COUNT(*) FROM cards WHERE card_type = 'Monster' GROUP BY atk`,
			scan: func(rows *sql.Rows) error {
				var key, count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByMonsterATK[key] = count
				return nil
			},
		},
		{
			sql: `SELECT def, COUNT(*) FROM cards WHERE card_type = 'Monster' GROUP BY def`,
			scan: func(rows *sql.Rows) error {
				var key, count int
				if err := rows.Scan(&key, &count); err != nil {
					return err
				}
				result.ByMonsterDEF[key] = count
				return nil
			},
		},
	}

	for _, q := range queries {
		rows, err := db.sql.Query(q.sql)
		if err != nil {
			return konami.CardCounts{}, fmt.Errorf("error GetCardCounts query %q: %w", q.sql, err)
		}
		for rows.Next() {
			if err := q.scan(rows); err != nil {
				rows.Close()
				return konami.CardCounts{}, fmt.Errorf("error GetCardCounts scan %q: %w", q.sql, err)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return konami.CardCounts{}, fmt.Errorf("error GetCardCounts rows.Err %q: %w", q.sql, err)
		}
		rows.Close()
	}

	return result, nil
}

// GetMapSetNumberToCardName returns a map from card_set_code (e.g. "LOB-001")
// to the card's English name.
// When the same code appears with multiple rarities,
// the first occurrence (ordered by card_set_code) wins.
func (db *DB) GetMapSetNumberToCardName() (map[string]string, error) {
	rows, err := db.sql.Query(`
        SELECT set_cards.card_set_code, cards.card_name_en
        FROM set_cards
        JOIN cards ON set_cards.card_id = cards.card_id
        ORDER BY set_cards.card_set_code`)
	if err != nil {
		return nil, fmt.Errorf("error GetMapSetNumberToCardName: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var setCode, cardNameEN string
		if err := rows.Scan(&setCode, &cardNameEN); err != nil {
			return nil, fmt.Errorf("error GetMapSetNumberToCardName rows.Scan: %w", err)
		}
		if _, exists := result[setCode]; !exists {
			result[setCode] = cardNameEN
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error GetMapSetNumberToCardName rows.Err: %w", err)
	}
	return result, nil
}
