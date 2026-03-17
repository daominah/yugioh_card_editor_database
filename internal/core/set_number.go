package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
)

// YuGiOhVersion enum, can be OCG, TCG, RushDuel, ...
type YuGiOhVersion string

// YuGiOhVersion enum values
const (
	OCG YuGiOhVersion = "OCG"
	TCG YuGiOhVersion = "TCG"
)

type KonamiSet struct {
	YuGiOhVersion YuGiOhVersion
	Abbreviation  string
	Name          string
	Type          string
	ReleaseDate   string // YYYY-MM-DD
}

func ParseYugipediaSetChronology(htmlData []byte) ([]KonamiSet, error) {
	root := textproc.HTMLParseToNode(htmlData)

	// because of XPath preceding and following do not work the same way in different libraries and XPath versions,
	// we need the following splitting approach to get the OCG and TCG tables separately:

	// split the HTML to get part1 including OCG and TCG tables (part2 is Rush Duel and beyond)
	rushDuelBegin, err := textproc.HTMLXPath(root, `//h2[.//*[@id="Rush_Duel"]]`)
	if err != nil {
		return nil, err
	}
	if len(rushDuelBegin) != 1 {
		return nil, fmt.Errorf("unexpected HTML structure, cannot find where Rush Duel section begins, len(rushDuelBegin): %v", len(rushDuelBegin))
	}
	partOCGAndTCG, _, err := SplitHTML(root, rushDuelBegin[0])
	if err != nil {
		return nil, fmt.Errorf("error SplitHTML at rushDuelBegin: %v", err)
	}

	// split OCG and TCG parts
	tcgBegin, err := textproc.HTMLXPath(partOCGAndTCG, `//h2[.//*[@id="TCG"]]`)
	if err != nil {
		return nil, err
	}
	if len(tcgBegin) != 1 {
		return nil, fmt.Errorf("unexpected HTML structure, cannot find where TCG section begins, len(tcgBegin): %v", len(tcgBegin))
	}
	partOCG, partTCG, err := SplitHTML(partOCGAndTCG, tcgBegin[0])
	if err != nil {
		return nil, fmt.Errorf("error SplitHTML at tcgBegin: %v", err)
	}

	// now extract all tables from OCG and TCG parts
	ocgTables, err := textproc.HTMLXPath(partOCG, `//table`)
	if err != nil {
		return nil, fmt.Errorf("error HTMLXPath ocgTables: %v", err)
	}
	tcgTables, err := textproc.HTMLXPath(partTCG, `//table`)
	if err != nil {
		return nil, fmt.Errorf("error HTMLXPath tcgTables: %v", err)
	}
	//println("debug: len(ocgTables):", len(ocgTables), "len(tcgTables):", len(tcgTables))
	if len(ocgTables) == 0 && len(tcgTables) == 0 {
		return nil, errors.New("no tables found for OCG or TCG, check the XPath")
	}
	var results []KonamiSet
	for _, v := range []struct {
		ygoVer YuGiOhVersion
		tables []*html.Node
	}{
		{ygoVer: OCG, tables: ocgTables},
		{ygoVer: TCG, tables: tcgTables},
	} {
		for _, table := range v.tables {
			rows, err := textproc.HTMLXPath(table, `//tr`)
			if err != nil {
				return nil, fmt.Errorf("error HTMLXPath tr: %v", err)
			}
			for _, row := range rows {
				//rowText := textproc.HTMLGetText(row)
				//if strings.Contains(rowText, "Magic Ruler") {
				//	println("debug row matched text: ", rowText)
				//	println("debug =========================================")
				//}

				cells, err := textproc.HTMLXPath(row, `//td`)
				if err != nil {
					return nil, err
				}
				if len(cells) < 4 {
					continue
				}
				s := KonamiSet{
					YuGiOhVersion: v.ygoVer,
					Abbreviation:  strings.ReplaceAll(textproc.HTMLGetText(cells[0]), "\n", " "),
					Name:          strings.ReplaceAll(textproc.HTMLGetText(cells[1]), "\n", " "),
					Type:          strings.ReplaceAll(textproc.HTMLGetText(cells[2]), "\n", " "),
					ReleaseDate:   "1970-01-01",
				}

				isParseTimeSuccess := false
				releaseDateStr := textproc.HTMLGetText(cells[3])
				if releaseDateStr == "" {
					continue
				}
				releaseDateStr = strings.TrimPrefix(releaseDateStr, "Unknown ")
				//println("debug parsing KonamiSet: ", s.Abbreviation, s.Name, s.Type, releaseDateStr)
				// parse release date: "4 February 1999", "March 2001", "Unknown 2003", ...
				for _, timeFormat := range []string{
					"2 January 2006",
					"January 2006",
					"2006",
				} {
					releaseDate, err := time.Parse(timeFormat, releaseDateStr)
					if err == nil {
						s.ReleaseDate = releaseDate.Format("2006-01-02")
						isParseTimeSuccess = true
						break
					}
				}
				if !isParseTimeSuccess {
					//println("debug unexpected release date format:", releaseDateStr)
					continue
				}
				results = append(results, s)
			}
		}
	}
	return results, nil
}

type SortKonamiSetsByReleaseDate []KonamiSet

func (s SortKonamiSetsByReleaseDate) Len() int { return len(s) }
func (s SortKonamiSetsByReleaseDate) Less(i, j int) bool {
	if s[i].ReleaseDate == s[j].ReleaseDate {
		if s[i].Name == s[j].Name {
			return s[i].YuGiOhVersion < s[j].YuGiOhVersion
		}
		return s[i].Name < s[j].Name
	}
	return s[i].ReleaseDate < s[j].ReleaseDate
}
func (s SortKonamiSetsByReleaseDate) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// MarshalKonamiSetsToCSV output can be used for csv.Writer.WriteAll
func MarshalKonamiSetsToCSV(sets []KonamiSet) [][]string {
	outputFields := []string{
		"YuGiOhVersion", "Abbreviation", "Name", "Type", "ReleaseDate",
	}
	records := [][]string{outputFields}
	for _, s := range sets {
		record := []string{
			string(s.YuGiOhVersion),
			s.Abbreviation,
			s.Name,
			s.Type,
			s.ReleaseDate,
		}
		records = append(records, record)
	}
	return records
}

// UnmarshalCSVToMapSetAbbreviationToName reads a CSV data (created by MarshalKonamiSetsToCSV)
// and returns a map of set abbreviation to set name
// (prefer TCG name if same abbreviation, as crawled cards data are in English)
func UnmarshalCSVToMapSetAbbreviationToName(records [][]string) map[string]string {
	setMap := make(map[string]string)
	if len(records) <= 1 {
		return setMap
	}
	for _, record := range records[1:] { // Skip header row
		if len(record) < 2 {
			continue
		}
		version := record[0]
		abbreviation := record[1]
		name := record[2]
		// prefer TCG name if the same abbreviation exists
		if existingName, exists := setMap[abbreviation]; exists {
			if version == "TCG" {
				setMap[abbreviation] = name
			} else if version == "OCG" && existingName == "" {
				setMap[abbreviation] = name
			}
		} else {
			setMap[abbreviation] = name
		}
	}
	return setMap
}
