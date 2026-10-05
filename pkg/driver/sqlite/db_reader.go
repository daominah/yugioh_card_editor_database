package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

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

// ListCardsEN returns every TCG/OCG card that has an English name
// (card_texts lang='en'), with fields for the human-readable CSV export.
// Rush Duel cards are excluded: Konami has no English Rush pages.
//
// MiscKonamiSet and MiscYear come from the card's first English print,
// ordered by release date (empty dates last), then by card_set_code.
// A print is English when its card number has the "-EN" region
// (for example "AGOV-EN022"),
// or it is an early TCG number without region (for example "LOB-001").
// sets.game_version alone is not enough:
// a set code shared by Japanese, English, and Korean prints keeps the first pass's version,
// for example "PP01" is "TCG" but also covers Korean "PP01-KR006".
// Cards without an English print leave both fields empty.
func (db *DB) ListCardsEN() ([]konami.Card, error) {
	rows, err := db.sql.Query(`
        WITH english_prints AS (
            SELECT sc.card_id, sc.card_set_code, sc.release_date,
                ROW_NUMBER() OVER (
                    PARTITION BY sc.card_id
                    ORDER BY sc.release_date = '', sc.release_date, sc.card_set_code
                ) AS print_order
            FROM set_cards sc
            LEFT JOIN sets s ON s.set_code = sc.set_code
            WHERE sc.card_set_code GLOB '*-EN*'
                OR (sc.card_set_code GLOB '*-[0-9]*' AND s.game_version = 'TCG')
        )
        SELECT
            c.card_id, t.name, c.card_type, c.card_subtype,
            c.attribute, c.monster_type, c.level_rank_link,
            c.atk, c.atk_str, c.def, c.def_str, c.abilities,
            COALESCE(p.password, ''),
            COALESCE(e.card_set_code, ''), COALESCE(e.release_date, '')
        FROM cards c
        JOIN card_texts t ON t.card_id = c.card_id AND t.lang = 'en' AND t.name != ''
        LEFT JOIN card_passwords p ON p.card_id = c.card_id
        LEFT JOIN english_prints e ON e.card_id = c.card_id AND e.print_order = 1
        ORDER BY c.card_id`)
	if err != nil {
		return nil, fmt.Errorf("error ListCardsEN: %w", err)
	}
	defer rows.Close()

	var cards []konami.Card
	for rows.Next() {
		var (
			cardID                                   int
			name, cardType, cardSubtype              string
			attribute, monsterType                   string
			levelRankLink, atk, def                  int
			atkStr, defStr, abilitiesJSON            string
			password, firstSetCode, firstReleaseDate string
		)
		err := rows.Scan(
			&cardID, &name, &cardType, &cardSubtype,
			&attribute, &monsterType, &levelRankLink,
			&atk, &atkStr, &def, &defStr, &abilitiesJSON,
			&password, &firstSetCode, &firstReleaseDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error ListCardsEN rows.Scan: %w", err)
		}
		var abilities []konami.MonsterAbility
		if err := json.Unmarshal([]byte(abilitiesJSON), &abilities); err != nil {
			abilities = nil
		}
		var year string
		if len(firstReleaseDate) >= 4 {
			year = firstReleaseDate[:4]
		}
		cards = append(cards, konami.Card{
			CardName:             name,
			CardNameEN:           name,
			CardType:             konami.CardType(cardType),
			CardSubtype:          konami.CardSubtype(cardSubtype),
			MonsterAttribute:     konami.MonsterAttribute(attribute),
			MonsterType:          konami.MonsterType(monsterType),
			MonsterLevelRankLink: levelRankLink,
			MonsterATK:           float64(atk),
			MonsterATKStr:        atkStr,
			MonsterDEF:           float64(def),
			MonsterDEFStr:        defStr,
			MonsterAbilities:     abilities,
			MiscKonamiCardID:     konami.CardID(strconv.Itoa(cardID)),
			MiscCardPassword:     password,
			MiscKonamiSet:        firstSetCode,
			MiscYear:             year,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error ListCardsEN rows.Err: %w", err)
	}
	return cards, nil
}

// ListSets returns every set (OCG, TCG, and Rush Duel)
// with the names collected from Japanese, Korean, and English card pages,
// ordered by set_code.
// A name is empty when no card page of that language lists the set.
func (db *DB) ListSets() ([]konami.KonamiSet, error) {
	rows, err := db.sql.Query(`
        SELECT set_code, game_version, release_date, name_ja, name_ko, name_en
        FROM sets
        ORDER BY set_code`)
	if err != nil {
		return nil, fmt.Errorf("error ListSets: %w", err)
	}
	defer rows.Close()

	var sets []konami.KonamiSet
	for rows.Next() {
		var s konami.KonamiSet
		var gameVersion string
		if err := rows.Scan(&s.Abbreviation, &gameVersion, &s.ReleaseDate, &s.NameJA, &s.NameKO, &s.NameEN); err != nil {
			return nil, fmt.Errorf("error ListSets rows.Scan: %w", err)
		}
		s.YuGiOhVersion = konami.YuGiOhVersion(gameVersion)
		sets = append(sets, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error ListSets rows.Err: %w", err)
	}
	return sets, nil
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
