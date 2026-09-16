package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"golang.org/x/text/width"
	_ "modernc.org/sqlite"
)

// Inputs (edit for each run, or pass the card ID as the first command line argument):
const defaultCardID = 19375

func main() {
	cardID := defaultCardID
	if len(os.Args) > 1 {
		parsed, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalf("error strconv.Atoi %q: %v", os.Args[1], err)
		}
		cardID = parsed
	}

	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sql.Open: %v", err)
	}
	defer db.Close()

	fmt.Printf("card_id %d\n", cardID)
	if err := printCardRow(db, cardID); err != nil {
		log.Fatalf("error printCardRow: %v", err)
	}
	if err := printCardTexts(db, cardID); err != nil {
		log.Fatalf("error printCardTexts: %v", err)
	}
	if err := printCardPassword(db, cardID); err != nil {
		log.Fatalf("error printCardPassword: %v", err)
	}
	if err := printSetCards(db, cardID); err != nil {
		log.Fatalf("error printSetCards: %v", err)
	}
}

// printCardRow prints the stats row from `cards`,
// falling back to `cards_rush` because Standard and Rush Duel cards
// share one card_id space and a given ID lives in exactly one of the two.
func printCardRow(db *sql.DB, cardID int) error {
	for _, table := range []string{"cards", "cards_rush"} {
		rows, err := queryRows(db, "SELECT * FROM "+table+" WHERE card_id = ?", cardID)
		if err != nil {
			return fmt.Errorf("error queryRows %s: %w", table, err)
		}
		if len(rows) == 0 {
			continue
		}
		printSection(table)
		printVertical(rows[0])
		return nil
	}
	return fmt.Errorf("card_id %d is in neither cards nor cards_rush", cardID)
}

// printCardTexts prints every stored column, one block per locale,
// so the name, its katakana reading, the effect, the Pendulum effect,
// and the display text of attribute and monster type all show up.
func printCardTexts(db *sql.DB, cardID int) error {
	rows, err := queryRows(db, "SELECT * FROM card_texts WHERE card_id = ? ORDER BY lang", cardID)
	if err != nil {
		return fmt.Errorf("error queryRows card_texts: %w", err)
	}
	if len(rows) == 0 {
		printSection("card_texts (no row)")
		return nil
	}
	for _, row := range rows {
		printSection("card_texts " + row.value("lang"))
		printVertical(row)
	}
	return nil
}

// printCardPassword prints the ygocdb.com password row, absent for Rush Duel
// cards and for the handful of Standard cards ygocdb.com does not list yet.
func printCardPassword(db *sql.DB, cardID int) error {
	rows, err := queryRows(db, "SELECT * FROM card_passwords WHERE card_id = ?", cardID)
	if err != nil {
		return fmt.Errorf("error queryRows card_passwords: %w", err)
	}
	if len(rows) == 0 {
		printSection("card_passwords (no row)")
		return nil
	}
	printSection("card_passwords")
	printVertical(rows[0])
	return nil
}

// printSetCards prints every print of the card across all locales,
// joining `sets` for the set name and `rarities` for the canonical rarity name.
func printSetCards(db *sql.DB, cardID int) error {
	rows, err := queryRows(db, `
        SELECT sc.card_set_code, sc.set_code,
               COALESCE(NULLIF(s.name_en, ''), NULLIF(s.name_ja, ''), NULLIF(s.name_ko, ''), '') AS set_name,
               s.game_version, sc.release_date,
               sc.rarity_code, sc.rarity_name, r.rarity_name_en
        FROM set_cards sc
        LEFT JOIN sets s ON s.set_code = sc.set_code
        LEFT JOIN rarities r ON r.rarity_code = sc.rarity_code
        WHERE sc.card_id = ?
        ORDER BY sc.release_date, sc.card_set_code, sc.rarity_code`, cardID)
	if err != nil {
		return fmt.Errorf("error queryRows set_cards: %w", err)
	}
	if len(rows) == 0 {
		printSection("set_cards (no row)")
		return nil
	}
	printSection(fmt.Sprintf("set_cards (%d prints)", len(rows)))
	printTable(rows)
	return nil
}

func printSection(title string) {
	fmt.Printf("\n--- %s ---\n", title)
}

// printVertical prints one column per line, for rows too wide to read as a table.
func printVertical(row dbRow) {
	nameWidth := 0
	for _, column := range row.columns {
		if n := displayWidth(column); n > nameWidth {
			nameWidth = n
		}
	}
	// card effects span several lines; indent the continuation lines
	// so they stay under the first line instead of under the column name.
	indent := strings.Repeat(" ", nameWidth+4)
	for i, column := range row.columns {
		value := strings.ReplaceAll(row.values[i], "\n", "\n"+indent)
		fmt.Printf("  %s  %s\n", pad(column, nameWidth), value)
	}
}

func printTable(rows []dbRow) {
	if len(rows) == 0 {
		return
	}

	widths := make([]int, len(rows[0].columns))
	for i, column := range rows[0].columns {
		widths[i] = displayWidth(column)
	}
	for _, row := range rows {
		for i, value := range row.values {
			if n := displayWidth(value); n > widths[i] {
				widths[i] = n
			}
		}
	}

	cells := make([]string, len(rows[0].columns))
	for i, column := range rows[0].columns {
		cells[i] = pad(column, widths[i])
	}
	fmt.Printf("  %s\n", strings.Join(cells, "  "))
	for _, row := range rows {
		for i, value := range row.values {
			cells[i] = pad(value, widths[i])
		}
		fmt.Printf("  %s\n", strings.Join(cells, "  "))
	}
}

// pad appends spaces so s occupies targetWidth terminal columns,
// leaving the text itself left-aligned within the column.
func pad(s string, targetWidth int) string {
	if n := displayWidth(s); n < targetWidth {
		return s + strings.Repeat(" ", targetWidth-n)
	}
	return s
}

// displayWidth counts the terminal columns a string occupies,
// which is neither its byte count nor its rune count:
// the kanji, kana, and Hangul filling this database render two columns wide,
// so counting runes would leave every localized cell short of its neighbours.
func displayWidth(s string) int {
	total := 0
	for _, r := range s {
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			total += 2
		default:
			total++
		}
	}
	return total
}

// queryRows reads any query into column-name plus stringified-value pairs,
// so a `SELECT *` prints every column without this script
// restating the schema of each table.
func queryRows(db *sql.DB, query string, args ...any) ([]dbRow, error) {
	sqlRows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error db.Query: %w", err)
	}
	defer sqlRows.Close()

	columns, err := sqlRows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error sqlRows.Columns: %w", err)
	}

	var rows []dbRow
	for sqlRows.Next() {
		scanned := make([]any, len(columns))
		for i := range scanned {
			scanned[i] = new(sql.NullString)
		}
		if err := sqlRows.Scan(scanned...); err != nil {
			return nil, fmt.Errorf("error sqlRows.Scan: %w", err)
		}
		values := make([]string, len(columns))
		for i, cell := range scanned {
			// NULL only reaches here through the LEFT JOINs in printSetCards;
			// every table column is declared NOT NULL.
			if nullable := cell.(*sql.NullString); nullable.Valid {
				values[i] = nullable.String
			} else {
				values[i] = "NULL"
			}
		}
		rows = append(rows, dbRow{columns: columns, values: values})
	}
	if err := sqlRows.Err(); err != nil {
		return nil, fmt.Errorf("error sqlRows.Err: %w", err)
	}
	return rows, nil
}

// dbRow is one result row, keeping the column order the query returned.
type dbRow struct {
	columns []string
	values  []string
}

func (r dbRow) value(column string) string {
	for i, c := range r.columns {
		if c == column {
			return r.values[i]
		}
	}
	return ""
}
