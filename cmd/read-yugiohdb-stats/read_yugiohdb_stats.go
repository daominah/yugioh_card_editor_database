package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
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
		{"cards special summon only", "SELECT COUNT(*) FROM cards WHERE is_special_summon_only = 1"},
		{"cards_rush total", "SELECT COUNT(*) FROM cards_rush"},
		{"cards_rush empty (card_type='')", "SELECT COUNT(*) FROM cards_rush WHERE card_type = ''"},
		{"card_texts total", "SELECT COUNT(*) FROM card_texts"},
		{"card_texts empty (name='')", "SELECT COUNT(*) FROM card_texts WHERE name = ''"},
		{"card_texts lang='ja'", "SELECT COUNT(*) FROM card_texts WHERE lang = 'ja'"},
		{"card_texts lang='ko'", "SELECT COUNT(*) FROM card_texts WHERE lang = 'ko'"},
		{"card_texts lang='en'", "SELECT COUNT(*) FROM card_texts WHERE lang = 'en'"},
		{"card_passwords total", "SELECT COUNT(*) FROM card_passwords"},
		// card_passwords comes from ygocdb.com, which also lists card ids absent from
		// the Konami crawl, so the matched count is the meaningful coverage number.
		{"card_passwords matched to cards", "SELECT COUNT(DISTINCT card_id) FROM card_passwords WHERE card_id IN (SELECT card_id FROM cards)"},
		{"sets total", "SELECT COUNT(*) FROM sets"},
		{"set_cards total", "SELECT COUNT(*) FROM set_cards"},
		{"rarities total", "SELECT COUNT(*) FROM rarities"},
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

	// Konami card IDs are assigned sequentially, so MAX(card_id) per lang
	// gives the most recently registered card on each locale's pages.
	// Restricted to standard (OCG/TCG) cards by joining `cards`, since Rush
	// Duel reprints get fresh card_ids and would otherwise win the MAX.
	fmt.Println("\n--- latest standard card per language (by max card_id, excluding Rush Duel) ---")
	rowsLatest, err := db.Query(`
        SELECT t.lang, t.card_id, t.name
        FROM card_texts t JOIN cards c ON c.card_id = t.card_id
        WHERE (t.lang, t.card_id) IN (
            SELECT t2.lang, MAX(t2.card_id)
            FROM card_texts t2 JOIN cards c2 ON c2.card_id = t2.card_id
            GROUP BY t2.lang
        )
        ORDER BY t.lang`)
	if err != nil {
		log.Fatalf("error latest-per-lang: %v", err)
	}
	defer rowsLatest.Close()
	for rowsLatest.Next() {
		var lang, name string
		var cardID int
		if err := rowsLatest.Scan(&lang, &cardID, &name); err != nil {
			log.Printf("error rows.Scan: %v", err)
			continue
		}
		fmt.Printf("  lang=%-2s card_id=%-6d %s\n", lang, cardID, name)
	}

	fmt.Println("\n--- card_passwords match rate against Standard cards ---")
	var stdTotal, stdMatched int
	_ = db.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&stdTotal)
	_ = db.QueryRow(`
        SELECT COUNT(*) FROM cards c
        WHERE EXISTS (SELECT 1 FROM card_passwords p WHERE p.card_id = c.card_id)`).Scan(&stdMatched)
	fmt.Printf("  matched: %d / %d (%.1f%%)\n",
		stdMatched, stdTotal, 100*float64(stdMatched)/float64(stdTotal))

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

	// suspicious empty fields: every check below should ideally count 0;
	// non-zero counts surface rows that the crawler/parser failed to populate.
	fmt.Println("\n--- suspicious empty fields ---")
	audits := []struct {
		label string
		query string
	}{
		// cards: every row should have card_type / card_subtype / card_name_en (filled by ja pass)
		{"cards.card_type=''", "SELECT COUNT(*) FROM cards WHERE card_type = ''"},
		{"cards.card_subtype=''", "SELECT COUNT(*) FROM cards WHERE card_subtype = ''"},
		{"cards.card_name_en=''", "SELECT COUNT(*) FROM cards WHERE card_name_en = ''"},
		{"cards.year=''", "SELECT COUNT(*) FROM cards WHERE year = ''"},
		// monster rows must have an attribute and a monster_type
		{"cards Monster attribute=''", "SELECT COUNT(*) FROM cards WHERE card_type = 'Monster' AND attribute = ''"},
		{"cards Monster monster_type=''", "SELECT COUNT(*) FROM cards WHERE card_type = 'Monster' AND monster_type = ''"},
		{"cards Monster level_rank_link=0 (non-Link)",
			"SELECT COUNT(*) FROM cards WHERE card_type = 'Monster' AND card_subtype != 'MonsterLink' AND level_rank_link = 0"},
		{"cards MonsterLink link_arrows='[]'",
			"SELECT COUNT(*) FROM cards WHERE card_subtype = 'MonsterLink' AND link_arrows = '[]'"},
		{"cards Pendulum pendulum_scale=0",
			"SELECT COUNT(*) FROM cards WHERE is_pendulum = 1 AND pendulum_scale = 0"},

		// cards_rush: every row should have card_type / card_subtype
		{"cards_rush.card_type=''", "SELECT COUNT(*) FROM cards_rush WHERE card_type = ''"},
		{"cards_rush.card_subtype=''", "SELECT COUNT(*) FROM cards_rush WHERE card_subtype = ''"},
		{"cards_rush Monster attribute=''",
			"SELECT COUNT(*) FROM cards_rush WHERE card_type = 'Monster' AND attribute = ''"},
		{"cards_rush Monster monster_type=''",
			"SELECT COUNT(*) FROM cards_rush WHERE card_type = 'Monster' AND monster_type = ''"},

		// card_texts: every row needs a name; monster rows need attribute_text / monster_type_text
		{"card_texts.name=''", "SELECT COUNT(*) FROM card_texts WHERE name = ''"},
		{"card_texts ja monster attribute_text=''",
			`SELECT COUNT(*) FROM card_texts t JOIN cards c ON c.card_id = t.card_id
             WHERE t.lang = 'ja' AND c.card_type = 'Monster' AND t.attribute_text = ''`},
		{"card_texts ja monster monster_type_text=''",
			`SELECT COUNT(*) FROM card_texts t JOIN cards c ON c.card_id = t.card_id
             WHERE t.lang = 'ja' AND c.card_type = 'Monster' AND t.monster_type_text = ''`},
		{"card_texts ja effect=''",
			"SELECT COUNT(*) FROM card_texts WHERE lang = 'ja' AND effect = ''"},

		// sets: every set row needs at least one localized name and a release_date
		{"sets all-three-names empty",
			"SELECT COUNT(*) FROM sets WHERE name_ja = '' AND name_ko = '' AND name_en = ''"},
		{"sets.release_date=''", "SELECT COUNT(*) FROM sets WHERE release_date = ''"},
		{"sets.game_version=''", "SELECT COUNT(*) FROM sets WHERE game_version = ''"},

		// set_cards: card_id=0 means the print could not be linked to a card row
		{"set_cards.card_id=0", "SELECT COUNT(*) FROM set_cards WHERE card_id = 0"},
		{"set_cards.set_code=''", "SELECT COUNT(*) FROM set_cards WHERE set_code = ''"},
		{"set_cards.rarity_code=''", "SELECT COUNT(*) FROM set_cards WHERE rarity_code = ''"},
		{"set_cards.rarity_name=''", "SELECT COUNT(*) FROM set_cards WHERE rarity_name = ''"},
		{"set_cards.release_date=''", "SELECT COUNT(*) FROM set_cards WHERE release_date = ''"},
		{"set_cards orphan set_code (no sets row)",
			"SELECT COUNT(*) FROM set_cards sc LEFT JOIN sets s ON s.set_code = sc.set_code WHERE s.set_code IS NULL"},

		// card_passwords: every row should have a non-empty 8-digit password
		{"card_passwords.password=''", "SELECT COUNT(*) FROM card_passwords WHERE password = ''"},
		{"card_passwords.card_name=''", "SELECT COUNT(*) FROM card_passwords WHERE card_name = ''"},
	}
	for _, a := range audits {
		var n int
		if err := db.QueryRow(a.query).Scan(&n); err != nil {
			log.Printf("error audit %q: %v", a.label, err)
			continue
		}
		marker := "  "
		if n > 0 {
			marker = "!!"
		}
		fmt.Printf("  %s %-44s %d\n", marker, a.label, n)
	}

	fmt.Println("\n--- sample MonsterLink rows with empty link_arrows (JA name) ---")
	rowsLink, err := db.Query(`
        SELECT c.card_id, c.card_subtype, c.attribute, c.monster_type,
               c.level_rank_link, c.atk, t.name
        FROM cards c LEFT JOIN card_texts t ON c.card_id = t.card_id AND t.lang = 'ja'
        WHERE c.card_subtype = 'MonsterLink' AND c.link_arrows = '[]'
        ORDER BY c.card_id LIMIT 20`)
	if err != nil {
		log.Fatalf("error link audit: %v", err)
	}
	defer rowsLink.Close()
	for rowsLink.Next() {
		var cardID, sub, attr, mt, name string
		var lv, atk int
		_ = rowsLink.Scan(&cardID, &sub, &attr, &mt, &lv, &atk, &name)
		fmt.Printf("  card_id=%-6s sub=%-12s attr=%-6s mt=%-14s LINK-%d atk=%-5d %s\n",
			cardID, sub, attr, mt, lv, atk, name)
	}
}
