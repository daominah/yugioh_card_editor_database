package core

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"

	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

//go:embed set_yugipedia.html
var yugipediaSetChronology []byte

func TestParseYugipediaSetChronology(t *testing.T) {

	// probably Yugipedia has some bot protection, so the following code does not work,
	// we can manually go to the page and save the HTML content, and save as set_yugipedia.html
	//
	//yugipediaResp, err := http.Get("https://yugipedia.com/wiki/Set_chronology#OCG")
	//if err != nil {
	//	t.Fatalf("error http.Get Yugipedia Set Chronology page: %v", err)
	//}
	//yugipediaSetChronology, err = io.ReadAll(yugipediaResp.Body)
	//if err != nil {
	//	t.Fatalf("error io.ReadAll: %v", err)
	//}
	//if yugipediaResp.StatusCode != 200 {
	//	t.Fatalf("error fetching Yugipedia Set Chronology page, status code: %v, body: %s", yugipediaResp.StatusCode, yugipediaSetChronology)
	//}

	if len(yugipediaSetChronology) == 0 {
		t.Fatalf("empty yugipediaSetChronology data")
	}

	konamiSets, err := ParseYugipediaSetChronology(yugipediaSetChronology)
	if err != nil {
		t.Fatalf("error ParseYugipediaSetChronology: %v", err)
	}
	t.Logf("len(konamiSets): %v", len(konamiSets))
	if len(konamiSets) == 0 {
		t.Fatalf("empty ParseYugipediaSetChronology result")
	}
	for i, konamiSet := range konamiSets {
		if i < 2 {
			t.Logf("example konamiSet: %+v", konamiSet)
		}
		if konamiSet.ReleaseDate == "1970-01-01" {
			t.Fatalf("konamiSet.ReleaseDate is the default value, probably not parsed correctly")
		}
	}

	// write to a CSV file
	sort.Sort(konami.SortKonamiSetsByReleaseDate(konamiSets))
	outputCSVFile, err := os.Create("yugioh_sets.csv")
	if err != nil {
		t.Fatalf("error creating CSV file: %v", err)
	}
	csvWriter := csv.NewWriter(outputCSVFile)
	err = csvWriter.WriteAll(konami.MarshalKonamiSetsToCSV(konamiSets))
	if err != nil {
		t.Fatalf("error csv.NewWriter.WriteAll: %v", err)
	}
	absPath, err := filepath.Abs(outputCSVFile.Name())
	if err != nil {
		t.Fatalf("error outputCSVFile absolute path: %v", err)
	}
	outputCSVFile.Close()
	t.Logf("wrote Konami sets data too: %v", absPath)

	mapSetsAbbreviationName := konami.UnmarshalCSVToMapSetAbbreviationToName(konami.MarshalKonamiSetsToCSV(konamiSets))
	mapYugiohVersion := make(map[string]konami.YuGiOhVersion)
	for _, set := range konamiSets {
		if _, found := mapYugiohVersion[set.Abbreviation]; !found {
			mapYugiohVersion[set.Abbreviation] = set.YuGiOhVersion
		} else {
			// set to empty if both OCG and TCG exist
			if set.YuGiOhVersion != mapYugiohVersion[set.Abbreviation] {
				mapYugiohVersion[set.Abbreviation] = ""
			}
		}
	}
	for _, c := range []struct {
		Abbreviation  string
		Name          string
		YuGiOhVersion konami.YuGiOhVersion
	}{
		{"LB", "Legend of Blue Eyes White Dragon", konami.OCG},
		{"LOB", "Legend of Blue Eyes White Dragon", konami.TCG},
		{"SD47", "Structure Deck: Advent of the Eyes of Blue", konami.OCG},
		{"SDWD", "Structure Deck: Blue-Eyes White Destiny", konami.TCG},
		{"POTE", "Power of the Elements", ""},
	} {
		if mapSetsAbbreviationName[c.Abbreviation] != c.Name {
			t.Errorf("error UnmarshalCSVToMapSetAbbreviationToName set %v, got: %v, want: %v", c.Abbreviation, mapSetsAbbreviationName[c.Abbreviation], c.Name)
		}
		if c.YuGiOhVersion != "" && mapYugiohVersion[c.Abbreviation] != c.YuGiOhVersion {
			t.Errorf("error YuGiOhVersion set %v got: %v, want: %v", c.Abbreviation, mapYugiohVersion[c.Abbreviation], c.YuGiOhVersion)
		}
	}
}

func TestSplitHTML(t *testing.T) {
	testHTMLData := `
		<!DOCTYPE html>
		<html lang="en">
		<head><title>Test HTML</title></head>
		<body>
		<h2><span id="OCG">OCG</span></h2>
		<table>
		  <tr>
			<td>OCGa1</td>
		  </tr>
		</table>
		<table>
		  <tr>
			<td>OCGb1</td>
			<td>OCGb2</td>
		  </tr>
		</table>
		<table>
		  <tr>
			<td>OCGc11</td>
			<td>OCGc12</td>
			<td>OCGc13</td>
		  </tr>
		  <tr>
			<td>OCGc21</td>
			<td>OCGc22</td>
			<td>OCGc23</td>
		  </tr>
		</table>
		<h2><span id="TCG">TCG</span></h2>
		<table>
		  <tr>
			<td>TCGd</td>
			<td>TCGd</td>
		  </tr>
		</table>
		<table>
		  <tr>
			<td>TCGe</td>
		  </tr>
		</table>
		<h2><span id="Rush_Duel">Rush Duel</span></h2>
		<table></table>
		</body>
		</html>

	`

	root, err := html.Parse(bytes.NewReader([]byte(testHTMLData)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	rushDuelNode := findNodeByID(root, "Rush_Duel")
	if rushDuelNode == nil {
		t.Fatalf("Rush_Duel node not found")
	}

	part1, part2, err := SplitHTML(root, rushDuelNode)
	if err != nil {
		t.Fatalf("splitHTML failed: %v", err)
	}

	tcgNode := findNodeByID(part1, "TCG")
	if tcgNode == nil {
		t.Fatalf("TCG node not found in part1")
	}

	part1OCG, part1TCG, err := SplitHTML(part1, tcgNode)
	if err != nil {
		t.Fatalf("splitHTML failed at TCG: %v", err)
	}

	// Validate the results
	if !containsNodeWithID(part1OCG, "OCG") {
		t.Errorf("part1OCG does not contain OCG section")
	}
	if !containsNodeWithID(part1TCG, "TCG") {
		t.Errorf("part1TCG does not contain TCG section")
	}
	if !containsNodeWithID(part2, "Rush_Duel") {
		t.Errorf("part2 does not contain Rush_Duel section")
	}

	// Validate the number of tables in each part
	ocgTables, err := textproc.HTMLXPath(part1OCG, `//table`)
	if err != nil {
		t.Fatalf("error HTMLXPath ocgTables: %v", err)
	}
	if len(ocgTables) != 3 {
		t.Errorf("Expected 3 tables in OCG part, got %d", len(ocgTables))
	}
	var ocgTexts []string
	ocgCells, err := textproc.HTMLXPath(part1OCG, `//table//td`)
	if err != nil {
		t.Fatalf("error HTMLXPath ocgCells: %v", err)
	}
	for _, row := range ocgCells {
		ocgTexts = append(ocgTexts, textproc.HTMLGetText(row))
	}
	if len(ocgCells) != 9 {
		t.Errorf("Expected 9 cells in OCG tables, got %d", len(ocgCells))
	}
	if strings.Join(ocgTexts, " ") != "OCGa1 OCGb1 OCGb2 OCGc11 OCGc12 OCGc13 OCGc21 OCGc22 OCGc23" {
		t.Errorf("Unexpected OCG cells text: %s", strings.Join(ocgTexts, " "))
	}

	tcgTables, err := textproc.HTMLXPath(part1TCG, `//table`)
	if err != nil {
		t.Fatalf("error HTMLXPath tcgTables: %v", err)
	}
	if len(tcgTables) != 2 {
		t.Errorf("Expected 2 tables in TCG part, got %d", len(tcgTables))
	}
	tcgRows, err := textproc.HTMLXPath(part1TCG, `//table//tr`)
	if err != nil {
		t.Fatalf("error HTMLXPath tcgRows: %v", err)
	}
	if len(tcgRows) != 2 {
		t.Errorf("error unexpected number of TCG rows, got: %d, want: 2", len(tcgRows))
	}
	tcgText := textproc.HTMLGetText(part1TCG)
	tcgTextOneLine := strings.Join(strings.Fields(tcgText), " ")
	expectedTCGText := "TCG TCGd TCGd TCGe"
	if tcgTextOneLine != expectedTCGText {
		t.Errorf("Unexpected TCG text: %s", tcgTextOneLine)
	}
}

func findNodeByID(root *html.Node, id string) *html.Node {
	var result *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != nil {
			return
		}
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				result = n
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return result
}

func containsNodeWithID(root *html.Node, id string) bool {
	return findNodeByID(root, id) != nil
}
