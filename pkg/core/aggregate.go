package core

import (
	"fmt"
	"strings"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

// Aggregate orchestrates all five lookup-table aggregation passes against db.
// Idempotent and safe to re-run after every crawl.
//
// For each pass it prints a header, a one-line summary per chosen row, a
// WARN block for any (key, lang) cell that has multiple text variants
// (the most-frequent is picked), and a "→ seeded N rows" line on success.
func Aggregate(db DatabaseAggregate) error {
	if err := runMonsterAttributesPass(db); err != nil {
		return fmt.Errorf("error runMonsterAttributesPass: %w", err)
	}
	if err := runMonsterTypesPass(db); err != nil {
		return fmt.Errorf("error runMonsterTypesPass: %w", err)
	}
	if err := runMonsterAttributesRushPass(db); err != nil {
		return fmt.Errorf("error runMonsterAttributesRushPass: %w", err)
	}
	if err := runMonsterTypesRushPass(db); err != nil {
		return fmt.Errorf("error runMonsterTypesRushPass: %w", err)
	}
	if err := runRaritiesPass(db); err != nil {
		return fmt.Errorf("error runRaritiesPass: %w", err)
	}
	return nil
}

func runMonsterAttributesPass(db DatabaseAggregate) error {
	printPassHeader("monster_attributes", "cards", "attribute", "attribute_text")
	inputs, err := db.FetchEnumVariants("cards", "attribute", "attribute_text")
	if err != nil {
		return err
	}
	rows, anomalies := konami.AggregateMonsterAttributes(inputs)
	printAnomalies(anomalies)
	for _, r := range rows {
		fmt.Printf("  %-15s  ja=%q  ko=%q  en=%q\n", r.Attribute, r.TextJA, r.TextKO, r.TextEN)
	}
	if len(rows) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	if err := db.UpsertMonsterAttributes(rows); err != nil {
		return err
	}
	fmt.Printf("  → seeded %d rows into monster_attributes\n", len(rows))
	return nil
}

func runMonsterTypesPass(db DatabaseAggregate) error {
	printPassHeader("monster_types", "cards", "monster_type", "monster_type_text")
	inputs, err := db.FetchEnumVariants("cards", "monster_type", "monster_type_text")
	if err != nil {
		return err
	}
	rows, anomalies := konami.AggregateMonsterTypes(inputs)
	printAnomalies(anomalies)
	for _, r := range rows {
		fmt.Printf("  %-15s  ja=%q  ko=%q  en=%q\n", r.MonsterType, r.TextJA, r.TextKO, r.TextEN)
	}
	if len(rows) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	if err := db.UpsertMonsterTypes(rows); err != nil {
		return err
	}
	fmt.Printf("  → seeded %d rows into monster_types\n", len(rows))
	return nil
}

func runMonsterAttributesRushPass(db DatabaseAggregate) error {
	printPassHeader("monster_attributes_rush", "cards_rush", "attribute", "attribute_text")
	inputs, err := db.FetchEnumVariants("cards_rush", "attribute", "attribute_text")
	if err != nil {
		return err
	}
	rows, anomalies := konami.AggregateMonsterAttributesRush(inputs)
	printAnomalies(anomalies)
	for _, r := range rows {
		fmt.Printf("  %-15s  ja=%q  ko=%q\n", r.Attribute, r.TextJA, r.TextKO)
	}
	if len(rows) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	if err := db.UpsertMonsterAttributesRush(rows); err != nil {
		return err
	}
	fmt.Printf("  → seeded %d rows into monster_attributes_rush\n", len(rows))
	return nil
}

func runMonsterTypesRushPass(db DatabaseAggregate) error {
	printPassHeader("monster_types_rush", "cards_rush", "monster_type", "monster_type_text")
	inputs, err := db.FetchEnumVariants("cards_rush", "monster_type", "monster_type_text")
	if err != nil {
		return err
	}
	rows, anomalies := konami.AggregateMonsterTypesRush(inputs)
	printAnomalies(anomalies)
	for _, r := range rows {
		fmt.Printf("  %-15s  ja=%q  ko=%q\n", r.MonsterType, r.TextJA, r.TextKO)
	}
	if len(rows) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	if err := db.UpsertMonsterTypesRush(rows); err != nil {
		return err
	}
	fmt.Printf("  → seeded %d rows into monster_types_rush\n", len(rows))
	return nil
}

func runRaritiesPass(db DatabaseAggregate) error {
	fmt.Printf("\n=== rarities (from set_cards.rarity_code × lang inferred from card_set_code) ===\n")
	inputs, err := db.FetchRarityVariants()
	if err != nil {
		return err
	}
	rows, anomalies, skipped := konami.AggregateRarities(inputs)
	if skipped > 0 {
		fmt.Printf("  (skipped %d rows whose card_set_code region maps to a non-en/ja/ko locale)\n", skipped)
	}
	for _, a := range anomalies {
		fmt.Printf("  WARN: rarity_code %q lang=%s has %d name variants:\n", a.Key, a.Lang, len(a.Variants))
		printVariants(a.Variants)
	}
	for _, r := range rows {
		fmt.Printf("  %-6s  ja=%q  ko=%q  en=%q\n", r.RarityCode, r.RarityNameJA, r.RarityNameKO, r.RarityNameEN)
	}
	if len(rows) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	if err := db.UpsertRarities(rows); err != nil {
		return err
	}
	fmt.Printf("  → seeded %d rows into rarities\n", len(rows))
	if err := db.UpdateRarityCardCounts(); err != nil {
		return err
	}
	fmt.Println("  → updated count_cards")
	return nil
}

func printPassHeader(lookupTable, cardsTable, canonicalCol, textCol string) {
	fmt.Printf("\n=== %s (from %s.%s × card_texts.%s) ===\n",
		lookupTable, cardsTable, canonicalCol, textCol)
}

func printAnomalies(anomalies []konami.Anomaly) {
	for _, a := range anomalies {
		fmt.Printf("  WARN: %s %s has %d variants:\n", a.Key, a.Lang, len(a.Variants))
		printVariants(a.Variants)
	}
}

func printVariants(vs []konami.TextVariant) {
	for i, v := range vs {
		mark := strings.Repeat(" ", 7)
		if i == 0 {
			mark = "  pick "
		}
		fmt.Printf("    %s %q (count=%d)\n", mark, v.Text, v.Count)
	}
}
