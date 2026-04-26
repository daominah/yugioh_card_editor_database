package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	_ "modernc.org/sqlite"
)

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

	counts := []struct {
		label string
		query string
	}{
		{"cards total", "SELECT COUNT(*) FROM cards"},
		{"cards empty (card_type='')", "SELECT COUNT(*) FROM cards WHERE card_type = ''"},
		{"cards_rush total", "SELECT COUNT(*) FROM cards_rush"},
		{"cards_rush empty (card_type='')", "SELECT COUNT(*) FROM cards_rush WHERE card_type = ''"},
		{"card_texts total", "SELECT COUNT(*) FROM card_texts"},
		{"card_texts empty (name='')", "SELECT COUNT(*) FROM card_texts WHERE name = ''"},
		{"sets total", "SELECT COUNT(*) FROM sets"},
		{"set_cards total", "SELECT COUNT(*) FROM set_cards"},
		{"monster_types total", "SELECT COUNT(*) FROM monster_types"},
		{"monster_types_rush total", "SELECT COUNT(*) FROM monster_types_rush"},
		{"monster_attributes total", "SELECT COUNT(*) FROM monster_attributes"},
		{"monster_attributes_rush total", "SELECT COUNT(*) FROM monster_attributes_rush"},
	}
	for _, c := range counts {
		var n int
		if err := db.QueryRow(c.query).Scan(&n); err != nil {
			log.Printf("error query %q: %v", c.query, err)
			continue
		}
		fmt.Printf("%-40s %d\n", c.label, n)
	}

	fmt.Println("\n--- sample empty cards (card_type='') ---")
	rows, err := db.Query(`SELECT card_id, card_type, attribute, monster_type FROM cards WHERE card_type = '' LIMIT 10`)
	if err != nil {
		log.Fatalf("error querying empty cards: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cardID, cardType, attribute, monsterType string
		if err := rows.Scan(&cardID, &cardType, &attribute, &monsterType); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  card_id=%-8s card_type=%q attribute=%q monster_type=%q\n",
			cardID, cardType, attribute, monsterType)
	}

	fmt.Println("\n--- sample non-empty cards ---")
	rows2, err := db.Query(`SELECT card_id, card_type, card_subtype, attribute, monster_type, level_rank_link, atk, def FROM cards WHERE card_type != '' LIMIT 10`)
	if err != nil {
		log.Fatalf("error querying non-empty cards: %v", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var cardID, cardType, cardSubtype, attribute, monsterType string
		var level, atk, def int
		if err := rows2.Scan(&cardID, &cardType, &cardSubtype, &attribute, &monsterType, &level, &atk, &def); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  card_id=%-8s type=%-15s subtype=%-20s attr=%-6s mt=%-12s lv=%d atk=%d def=%d\n",
			cardID, cardType, cardSubtype, attribute, monsterType, level, atk, def)
	}

	fmt.Println("\n--- distinct attribute values in cards_rush ---")
	rows3, err := db.Query(`SELECT attribute, COUNT(*) as n FROM cards_rush GROUP BY attribute ORDER BY n DESC`)
	if err != nil {
		log.Fatalf("error querying cards_rush attributes: %v", err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var attr string
		var n int
		if err := rows3.Scan(&attr, &n); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  attribute=%-24q  count=%d\n", attr, n)
	}

	fmt.Println("\n--- sample card_texts JA (name_katakana inspection) ---")
	rows5, err := db.Query(`SELECT card_id, name, name_katakana FROM card_texts WHERE lang = 'ja' LIMIT 10`)
	if err != nil {
		log.Fatalf("error querying card_texts ja: %v", err)
	}
	defer rows5.Close()
	for rows5.Next() {
		var cardID, name, nameKat string
		if err := rows5.Scan(&cardID, &name, &nameKat); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  card_id=%-6s name=%q\n      name_katakana=%q\n", cardID, name, nameKat)
	}

	fmt.Println("\n--- sample cards_rush rows ---")
	rows4, err := db.Query(`SELECT card_id, card_type, attribute, monster_type, atk, def FROM cards_rush ORDER BY card_id LIMIT 20`)
	if err != nil {
		log.Fatalf("error querying cards_rush rows: %v", err)
	}
	defer rows4.Close()
	for rows4.Next() {
		var cardID, cardType, attr, monsterType string
		var atk, def int
		if err := rows4.Scan(&cardID, &cardType, &attr, &monsterType, &atk, &def); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  card_id=%-8s type=%-8s attr=%-24s mt=%-20s atk=%d def=%d\n",
			cardID, cardType, attr, monsterType, atk, def)
	}

	fmt.Println("\n--- card_texts per locale ---")
	rowsLoc, err := db.Query(`SELECT lang, COUNT(*) FROM card_texts GROUP BY lang ORDER BY lang`)
	if err != nil {
		log.Fatalf("error per-locale: %v", err)
	}
	defer rowsLoc.Close()
	for rowsLoc.Next() {
		var lang string
		var n int
		_ = rowsLoc.Scan(&lang, &n)
		fmt.Printf("  lang=%-4s count=%d\n", lang, n)
	}

	fmt.Println("\n--- is_special_summon_only counts ---")
	var ssoTrue, ssoFalse int
	_ = db.QueryRow(`SELECT COUNT(*) FROM cards WHERE is_special_summon_only = 1`).Scan(&ssoTrue)
	_ = db.QueryRow(`SELECT COUNT(*) FROM cards WHERE is_special_summon_only = 0`).Scan(&ssoFalse)
	fmt.Printf("  true:  %d\n  false: %d\n", ssoTrue, ssoFalse)

	fmt.Println("\n--- sample is_special_summon_only=1 cards (JA names) ---")
	rowsSSO, err := db.Query(`
        SELECT c.card_id, c.monster_type, c.card_subtype, t.name
        FROM cards c LEFT JOIN card_texts t ON c.card_id = t.card_id AND t.lang = 'ja'
        WHERE c.is_special_summon_only = 1
        ORDER BY c.card_id LIMIT 12`)
	if err != nil {
		log.Fatalf("error sso sample: %v", err)
	}
	defer rowsSSO.Close()
	for rowsSSO.Next() {
		var cardID, mt, sub, name string
		_ = rowsSSO.Scan(&cardID, &mt, &sub, &name)
		fmt.Printf("  %-6s %-12s %-18s %s\n", cardID, mt, sub, name)
	}

	fmt.Println("\n--- cards_rush rows where attribute looks wrong (non-canonical) ---")
	rowsAttr, err := db.Query(`
        SELECT card_id, card_type, card_subtype, attribute, monster_type
        FROM cards_rush
        WHERE attribute NOT IN ('','LIGHT','DARK','EARTH','FIRE','WATER','WIND','DIVINE')
        LIMIT 10`)
	if err != nil {
		log.Fatalf("error attr audit: %v", err)
	}
	defer rowsAttr.Close()
	for rowsAttr.Next() {
		var cardID, cardType, sub, attr, mt string
		_ = rowsAttr.Scan(&cardID, &cardType, &sub, &attr, &mt)
		fmt.Printf("  %-7s type=%-8s sub=%-15s attr=%-12s mt=%s\n", cardID, cardType, sub, attr, mt)
	}
}
