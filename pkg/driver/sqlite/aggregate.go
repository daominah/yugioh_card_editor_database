package sqlite

import (
	"fmt"

	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

// Compile-time check that *DB implements core.DatabaseAggregate.
var _ core.DatabaseAggregate = (*DB)(nil)

// FetchEnumVariants reads grouped (canonical, lang, text, count) rows from a
// (cards|cards_rush) JOIN card_texts pairing, restricted to monster rows
// with a non-empty canonical value. The output feeds konami.AggregateMonster*.
func (db *DB) FetchEnumVariants(cardsTable, canonicalCol, textCol string) ([]konami.EnumVariantInput, error) {
	rows, err := db.sql.Query(fmt.Sprintf(`
		SELECT %[1]s.%[2]s, card_texts.lang, card_texts.%[3]s, COUNT(*)
		FROM %[1]s JOIN card_texts ON %[1]s.card_id = card_texts.card_id
		WHERE %[1]s.card_type = 'Monster' AND %[1]s.%[2]s != ''
		GROUP BY %[1]s.%[2]s, card_texts.lang, card_texts.%[3]s
		ORDER BY %[1]s.%[2]s, card_texts.lang, 4 DESC`,
		cardsTable, canonicalCol, textCol))
	if err != nil {
		return nil, fmt.Errorf("error db.Query %v.%v × %v: %w", cardsTable, canonicalCol, textCol, err)
	}
	defer rows.Close()
	var out []konami.EnumVariantInput
	for rows.Next() {
		var v konami.EnumVariantInput
		if err := rows.Scan(&v.Canonical, &v.Lang, &v.Text, &v.Count); err != nil {
			return nil, fmt.Errorf("error rows.Scan: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
}

// FetchRarityVariants reads grouped (card_set_code, rarity_code, rarity_name,
// count) rows from set_cards. The output feeds konami.AggregateRarities.
func (db *DB) FetchRarityVariants() ([]konami.RarityVariantInput, error) {
	rows, err := db.sql.Query(`
		SELECT card_set_code, rarity_code, rarity_name, COUNT(*)
		FROM set_cards
		WHERE rarity_code != '' AND rarity_name != ''
		GROUP BY card_set_code, rarity_code, rarity_name`)
	if err != nil {
		return nil, fmt.Errorf("error db.Query set_cards: %w", err)
	}
	defer rows.Close()
	var out []konami.RarityVariantInput
	for rows.Next() {
		var v konami.RarityVariantInput
		if err := rows.Scan(&v.CardSetCode, &v.RarityCode, &v.RarityName, &v.Count); err != nil {
			return nil, fmt.Errorf("error rows.Scan: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
}

// UpsertMonsterAttributes writes rows into the monster_attributes lookup table.
func (db *DB) UpsertMonsterAttributes(rows []konami.MonsterAttributeRow) error {
	for _, r := range rows {
		if _, err := db.sql.Exec(
			`INSERT OR REPLACE INTO monster_attributes (attribute, text_ja, text_ko, text_en) VALUES (?, ?, ?, ?)`,
			string(r.Attribute), r.TextJA, r.TextKO, r.TextEN,
		); err != nil {
			return fmt.Errorf("error db.Exec INSERT monster_attributes %v: %w", r.Attribute, err)
		}
	}
	return nil
}

// UpsertMonsterAttributesRush writes rows into monster_attributes_rush.
func (db *DB) UpsertMonsterAttributesRush(rows []konami.MonsterAttributeRushRow) error {
	for _, r := range rows {
		if _, err := db.sql.Exec(
			`INSERT OR REPLACE INTO monster_attributes_rush (attribute, text_ja, text_ko) VALUES (?, ?, ?)`,
			string(r.Attribute), r.TextJA, r.TextKO,
		); err != nil {
			return fmt.Errorf("error db.Exec INSERT monster_attributes_rush %v: %w", r.Attribute, err)
		}
	}
	return nil
}

// UpsertMonsterTypes writes rows into monster_types.
func (db *DB) UpsertMonsterTypes(rows []konami.MonsterTypeRow) error {
	for _, r := range rows {
		if _, err := db.sql.Exec(
			`INSERT OR REPLACE INTO monster_types (monster_type, text_ja, text_ko, text_en) VALUES (?, ?, ?, ?)`,
			string(r.MonsterType), r.TextJA, r.TextKO, r.TextEN,
		); err != nil {
			return fmt.Errorf("error db.Exec INSERT monster_types %v: %w", r.MonsterType, err)
		}
	}
	return nil
}

// UpsertMonsterTypesRush writes rows into monster_types_rush.
func (db *DB) UpsertMonsterTypesRush(rows []konami.MonsterTypeRushRow) error {
	for _, r := range rows {
		if _, err := db.sql.Exec(
			`INSERT OR REPLACE INTO monster_types_rush (monster_type, text_ja, text_ko) VALUES (?, ?, ?)`,
			string(r.MonsterType), r.TextJA, r.TextKO,
		); err != nil {
			return fmt.Errorf("error db.Exec INSERT monster_types_rush %v: %w", r.MonsterType, err)
		}
	}
	return nil
}

// UpsertRarities writes rows into the rarities lookup table. CountCards is
// not written here; call UpdateRarityCardCounts after this for that field.
func (db *DB) UpsertRarities(rows []konami.RarityRow) error {
	for _, r := range rows {
		if _, err := db.sql.Exec(
			`INSERT OR REPLACE INTO rarities (rarity_code, alias_code, rarity_name_ja, rarity_name_ko, rarity_name_en) VALUES (?, ?, ?, ?, ?)`,
			r.RarityCode, r.AliasCode, r.RarityNameJA, r.RarityNameKO, r.RarityNameEN,
		); err != nil {
			return fmt.Errorf("error db.Exec INSERT rarities %v: %w", r.RarityCode, err)
		}
	}
	return nil
}

// UpdateRarityCardCounts recomputes rarities.count_cards from set_cards (one
// SQL UPDATE; counts distinct card_id per rarity_code, ignoring card_id=0).
func (db *DB) UpdateRarityCardCounts() error {
	if _, err := db.sql.Exec(`
		UPDATE rarities SET count_cards = (
			SELECT COUNT(DISTINCT card_id)
			FROM set_cards
			WHERE set_cards.rarity_code = rarities.rarity_code
			  AND card_id != 0
		)`); err != nil {
		return fmt.Errorf("error db.Exec UPDATE count_cards: %w", err)
	}
	return nil
}
