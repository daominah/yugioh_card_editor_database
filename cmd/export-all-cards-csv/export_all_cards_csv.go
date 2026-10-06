// export-all-cards-csv writes human-readable files from the existing `data/yugioh.db`,
// in a few seconds:
//   - `data/yugioh_cards.csv`: every card with an English name, without effect text
//   - `data/yugioh_cards.xlsx`: the same card list, rows colored by card type
//   - `data/yugioh_sets.csv`: every set (OCG, TCG, Rush Duel), sorted by release date
//
// It does not crawl: to include new cards,
// run `cmd/crawl-konami-db-full` first (it rebuilds the database, a few minutes).
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/xuri/excelize/v2"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/core"
	"github.com/daominah/yugioh_card_editor_database/pkg/driver/sqlite"
	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}
	dataDir := filepath.Join(projectRoot, "data")

	db, err := sqlite.Open(filepath.Join(dataDir, "yugioh.db"))
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer db.Close()

	sets, err := db.ListSets()
	if err != nil {
		log.Fatalf("error ListSets: %v", err)
	}
	// Stable keeps the set_code order of ListSets for ties
	// (same release date and English name, often both OCG sets without NameEN),
	// so re-running on the same database writes the same file.
	sort.Stable(konami.SortKonamiSetsByReleaseDate(sets))
	writeCSV(konami.MarshalKonamiSetsToCSV(sets), filepath.Join(dataDir, "yugioh_sets.csv"))

	cardsTable := readCardsTable(db, sets)
	writeCSV(cardsTable, filepath.Join(dataDir, "yugioh_cards.csv"))
	writeCardsXLSX(cardsTable, filepath.Join(dataDir, "yugioh_cards.xlsx"))
}

// readCardsTable returns every card with an English name
// (without effect text, sorted by name) as rows for human view and search,
// the first row is the header.
// Both yugioh_cards.csv and yugioh_cards.xlsx are written from these rows.
//
// Set full names are the English names Konami prints on English card pages,
// mostly upper case (for example "DIMENSION OF CHAOS").
// Before this command, they came from a manually saved Yugipedia page in title case.
func readCardsTable(db core.DatabaseReader, sets []konami.KonamiSet) [][]string {
	cards, err := db.ListCardsEN()
	if err != nil {
		log.Fatalf("error ListCardsEN: %v", err)
	}
	mapSetsNameEN := make(map[string]string)
	for _, s := range sets {
		mapSetsNameEN[s.Abbreviation] = s.NameEN
	}
	sort.Sort(konami.SortCardNames(cards))
	return konami.ToCSV(cards, mapSetsNameEN)
}

// writeCSV writes records to path, the first record is the header.
func writeCSV(records [][]string, path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("error os.Create %v: %v", path, err)
	}
	defer f.Close()
	if err := csv.NewWriter(f).WriteAll(records); err != nil {
		log.Fatalf("error csv.Writer.WriteAll %v: %v", path, err)
	}
	log.Printf("wrote %v rows (without header) to %v", len(records)-1, path)
}

// writeCardsXLSX writes the cards table as a spreadsheet
// that colors each row by card type (Normal Monster, other Monster, Spell, Trap),
// alternating a lighter and darker shade between even and odd rows.
// The colors are fixed cell fills computed here,
// not conditional formatting (formulas evaluated by the spreadsheet app),
// to keep the big sheet fast to open and export.
// Fills move with their rows when the user sorts,
// so card type colors stay right, but the even and odd shades no longer alternate.
//
// Every cell is plain text aligned left, including numbers like CardID and ATK,
// so values show exactly as in the CSV
// (the card named "7", a password with leading zeros like "08233522").
// Column widths, font, and colors copy the previously hand-made file.
func writeCardsXLSX(records [][]string, path string) {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Sheet1"

	for i, width := range cardsTableColumnWidths {
		column, _ := excelize.ColumnNumberToName(i + 1)
		if err := f.SetColWidth(sheet, column, column, width); err != nil {
			log.Fatalf("error SetColWidth %v: %v", column, err)
		}
	}
	// Hide the unused columns, so the sheet ends at the last data column.
	lastColumn, _ := excelize.ColumnNumberToName(len(records[0]))
	firstUnusedColumn, _ := excelize.ColumnNumberToName(len(records[0]) + 1)
	if err := f.SetColVisible(sheet, firstUnusedColumn+":XFD", false); err != nil {
		log.Fatalf("error SetColVisible: %v", err)
	}
	// The PDF page (Height 19.9", Width 19", see README) fits 100 rows at this height,
	// but only 85 rows at the excelize default 15 points.
	// Set both the sheet default and each row's height,
	// as the previously hand-made file did,
	// in case a spreadsheet app reads only one of them.
	isCustomHeight := true
	defaultRowHeight := cardsTableRowHeight
	if err := f.SetSheetProps(sheet, &excelize.SheetPropsOptions{
		DefaultRowHeight: &defaultRowHeight,
		CustomHeight:     &isCustomHeight,
	}); err != nil {
		log.Fatalf("error SetSheetProps: %v", err)
	}
	// Google Sheets keeps these margins when exporting to PDF.
	// With top and bottom 1.05", the 19.9" page height leaves 17.79" (1,281 points),
	// exactly 100 rows of 12.8 points (header plus 99 cards on the first page).
	// Without margins in the file, Google uses narrower ones and fits 103 rows.
	if err := f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Top: &cardsTablePageMarginTopBottom, Bottom: &cardsTablePageMarginTopBottom,
		Left: &cardsTablePageMarginOther, Right: &cardsTablePageMarginOther,
		Header: &cardsTablePageMarginOther, Footer: &cardsTablePageMarginOther,
	}); err != nil {
		log.Fatalf("error SetPageMargins: %v", err)
	}

	headerStyle := newCardsTableStyle(f, true, "")
	bodyStyle := newCardsTableStyle(f, false, "")
	// fillStyles caches one style per fill color.
	fillStyles := make(map[string]int)
	cardTypeIndex := slices.Index(records[0], "CardType")
	cardSubtypeIndex := slices.Index(records[0], "CardSubtype")
	for i, record := range records {
		sheetRow := i + 1 // header is spreadsheet row 1
		firstCell := fmt.Sprintf("A%v", sheetRow)
		row := make([]any, len(record))
		for j, value := range record {
			row[j] = value
		}
		if err := f.SetSheetRow(sheet, firstCell, &row); err != nil {
			log.Fatalf("error SetSheetRow %v: %v", firstCell, err)
		}
		if err := f.SetRowHeight(sheet, sheetRow, cardsTableRowHeight); err != nil {
			log.Fatalf("error SetRowHeight %v: %v", sheetRow, err)
		}
		if i == 0 {
			setRowStyle(f, sheet, firstCell, fmt.Sprintf("%v%v", lastColumn, sheetRow), headerStyle)
			continue
		}

		// Column A (Row) stays white, the fill covers column B to the last column.
		setRowStyle(f, sheet, firstCell, firstCell, bodyStyle)
		color := cardsTableRowColor(record[cardTypeIndex], record[cardSubtypeIndex], sheetRow)
		style := bodyStyle
		if color != "" {
			var ok bool
			style, ok = fillStyles[color]
			if !ok {
				style = newCardsTableStyle(f, false, color)
				fillStyles[color] = style
			}
		}
		setRowStyle(f, sheet, fmt.Sprintf("B%v", sheetRow), fmt.Sprintf("%v%v", lastColumn, sheetRow), style)
	}

	if err := f.SaveAs(path); err != nil {
		log.Fatalf("error SaveAs %v: %v", path, err)
	}
	log.Printf("wrote %v cards to %v", len(records)-1, path)
}

// newCardsTableStyle returns a yugioh_cards.xlsx cell style:
// Arial 10, text number format "@", aligned left,
// with a solid fill when fillColor is not empty.
func newCardsTableStyle(f *excelize.File, isBold bool, fillColor string) int {
	style := &excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 10, Bold: isBold},
		NumFmt:    49, // built-in format "@": plain text
		Alignment: &excelize.Alignment{Horizontal: "left"},
	}
	if fillColor != "" {
		style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{fillColor}}
	}
	id, err := f.NewStyle(style)
	if err != nil {
		log.Fatalf("error NewStyle isBold=%v fillColor=%v: %v", isBold, fillColor, err)
	}
	return id
}

func setRowStyle(f *excelize.File, sheet, fromCell, toCell string, style int) {
	if err := f.SetCellStyle(sheet, fromCell, toCell, style); err != nil {
		log.Fatalf("error SetCellStyle %v:%v: %v", fromCell, toCell, err)
	}
}

// cardsTableRowColor returns the yugioh_cards.xlsx fill color of a data row,
// or empty for an unknown card type.
// sheetRow is the 1-based spreadsheet row number,
// its parity picks the shade the same way as the formula ISEVEN(ROW())
// in the previously hand-made file's conditional formatting.
func cardsTableRowColor(cardType, cardSubtype string, sheetRow int) string {
	group := cardType
	if cardType == string(konami.Monster) && cardSubtype == "Normal" {
		group = "MonsterNormal"
	}
	shades, ok := cardsTableRowColors[group]
	if !ok {
		return ""
	}
	if sheetRow%2 == 0 {
		return shades.even
	}
	return shades.odd
}

// cardsTableRowHeight is the yugioh_cards.xlsx row height in points,
// copied from the previously hand-made file.
const cardsTableRowHeight = 12.8

// cardsTablePageMarginTopBottom and cardsTablePageMarginOther are the yugioh_cards.xlsx
// page margins in inches, copied from the previously hand-made file
// (LibreOffice defaults: 2.67 cm top and bottom, 2 cm for the others).
var (
	cardsTablePageMarginTopBottom = 1.05277777777778
	cardsTablePageMarginOther     = 0.7875
)

// cardsTableColumnWidths are the yugioh_cards.xlsx column widths
// in the order of konami.ToCSV columns, copied from the previously hand-made file.
var cardsTableColumnWidths = []float64{
	6.12,         // Row, up to 5 digits
	51.07,        // ENName
	12.76, 12.76, // CardType, CardSubtype
	8.17, 12.76, // CardID, CardPasswd
	8.17, 12.76, // Attribute, Type
	6.12, 6.12, 6.12, 6.12, 6.12, // Level, ATK, DEF, Tuner, ENYear
	13.24, 51.07, // ENSet, ENSetFullName
}

// cardsTableRowColors are the yugioh_cards.xlsx fill colors per card type group,
// copied from the previously hand-made file.
// "MonsterNormal" is a Monster with CardSubtype "Normal",
// "Monster" covers every other Monster.
var cardsTableRowColors = map[string]struct{ even, odd string }{
	"MonsterNormal": {even: "FFFFA6", odd: "FFFF6D"}, // light yellow, yellow
	"Monster":       {even: "FF972F", odd: "FF860D"}, // light orange, orange
	"Spell":         {even: "77BC65", odd: "3FAF46"}, // light green, green
	"Trap":          {even: "A1467E", odd: "BF819E"}, // purple, light purple
}
