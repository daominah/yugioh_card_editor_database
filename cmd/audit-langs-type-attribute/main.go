// audit-monster-lookups joins canonical enums in cards / cards_rush against
// the localized text columns in card_texts, picks the most-frequent localized
// string per (enum, language) cell, and writes those into the four enum
// dictionaries (monster_attributes, monster_types, *_rush). Anomalies — any
// cell where multiple distinct localized strings appear — are printed to
// stdout so the user can decide whether parse.go's hardcoded switch needs
// updating. Idempotent: safe to re-run after every crawl.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/daominah/yugioh_card_editor/pkg/base"
)

// task describes one audit-and-seed pass over a (canonical column, locale
// text column) pairing for one source/lookup table couple.
type task struct {
	cardsTable   string   // "cards" or "cards_rush"
	canonicalCol string   // "attribute" or "monster_type"
	textCol      string   // "attribute_text" or "monster_type_text"
	lookupTable  string   // "monster_attributes", "monster_types", or *_rush
	pkCol        string   // PK column of lookup table (matches canonicalCol value)
	langs        []string // locales whose text columns the lookup table has
}

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}
	db, err := sql.Open("sqlite", filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sql.Open: %v", err)
	}
	defer db.Close()

	tasks := []task{
		{cardsTable: "cards", canonicalCol: "attribute", textCol: "attribute_text",
			lookupTable: "monster_attributes", pkCol: "attribute", langs: []string{"en", "ja", "ko"}},
		{cardsTable: "cards", canonicalCol: "monster_type", textCol: "monster_type_text",
			lookupTable: "monster_types", pkCol: "monster_type", langs: []string{"en", "ja", "ko"}},
		{cardsTable: "cards_rush", canonicalCol: "attribute", textCol: "attribute_text",
			lookupTable: "monster_attributes_rush", pkCol: "attribute", langs: []string{"ja", "ko"}},
		{cardsTable: "cards_rush", canonicalCol: "monster_type", textCol: "monster_type_text",
			lookupTable: "monster_types_rush", pkCol: "monster_type", langs: []string{"ja", "ko"}},
	}
	for _, t := range tasks {
		if err := runTask(db, t); err != nil {
			log.Printf("error task %v: %v", t.lookupTable, err)
		}
	}
}

// runTask audits one source-table × column pairing and writes the chosen
// localized strings into the lookup table. Prints warnings for any (enum,
// lang) cell with multiple variants, choosing the most frequent.
func runTask(db *sql.DB, t task) error {
	fmt.Printf("\n=== %s (from %s.%s × card_texts.%s) ===\n", t.lookupTable, t.cardsTable, t.canonicalCol, t.textCol)

	// variants[canonical][lang] = ordered list of (text, count), most-frequent first
	variants := map[string]map[string][]variant{}

	rows, err := db.Query(fmt.Sprintf(`
		SELECT %[1]s.%[2]s, card_texts.lang, card_texts.%[3]s, COUNT(*)
		FROM %[1]s JOIN card_texts ON %[1]s.card_id = card_texts.card_id
		WHERE %[1]s.card_type = 'Monster' AND %[1]s.%[2]s != ''
		GROUP BY %[1]s.%[2]s, card_texts.lang, card_texts.%[3]s
		ORDER BY %[1]s.%[2]s, card_texts.lang, 4 DESC`,
		t.cardsTable, t.canonicalCol, t.textCol))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var canonical, lang, text string
		var n int
		if err := rows.Scan(&canonical, &lang, &text, &n); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		if _, ok := variants[canonical]; !ok {
			variants[canonical] = map[string][]variant{}
		}
		variants[canonical][lang] = append(variants[canonical][lang], variant{text: text, count: n})
	}

	// Sorted canonical list for stable output
	canonicals := make([]string, 0, len(variants))
	for c := range variants {
		canonicals = append(canonicals, c)
	}
	sort.Strings(canonicals)

	// Print chosen rows + flag anomalies, collect cells for INSERT
	type chosenRow struct {
		canonical string
		texts     map[string]string // lang → chosen text
	}
	var chosen []chosenRow
	for _, c := range canonicals {
		row := chosenRow{canonical: c, texts: map[string]string{}}
		for _, lang := range t.langs {
			vs := variants[c][lang]
			if len(vs) == 0 {
				continue // no data for this (enum, lang) — silent
			}
			row.texts[lang] = vs[0].text
			if len(vs) > 1 {
				fmt.Printf("  WARN: %s %s has %d variants:\n", c, lang, len(vs))
				for i, v := range vs {
					mark := "       "
					if i == 0 {
						mark = "  pick "
					}
					fmt.Printf("    %s %q (count=%d)\n", mark, v.text, v.count)
				}
			}
		}
		chosen = append(chosen, row)
	}

	// Print a one-line summary per canonical
	for _, row := range chosen {
		parts := make([]string, 0, len(t.langs))
		for _, lang := range t.langs {
			parts = append(parts, fmt.Sprintf("%s=%q", lang, row.texts[lang]))
		}
		fmt.Printf("  %-15s  %s\n", row.canonical, strings.Join(parts, "  "))
	}

	// Write INSERT OR REPLACE rows. Column list: PK + text_<lang> for each lang.
	if len(chosen) == 0 {
		fmt.Println("  (no rows to seed — source data is empty)")
		return nil
	}
	cols := []string{t.pkCol}
	placeholders := []string{"?"}
	for _, lang := range t.langs {
		cols = append(cols, "text_"+lang)
		placeholders = append(placeholders, "?")
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO %s (%s) VALUES (%s)`,
		t.lookupTable, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	for _, row := range chosen {
		args := []any{row.canonical}
		for _, lang := range t.langs {
			args = append(args, row.texts[lang]) // empty string OK if no data
		}
		if _, err := db.Exec(stmt, args...); err != nil {
			return fmt.Errorf("insert %s=%v: %w", t.canonicalCol, row.canonical, err)
		}
	}
	fmt.Printf("  → seeded %d rows into %s\n", len(chosen), t.lookupTable)
	return nil
}

type variant struct {
	text  string
	count int
}
