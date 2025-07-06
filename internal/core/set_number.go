package core

import (
	"strings"
	"time"

	"github.com/mywrap/textproc"
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
	elems, err := textproc.HTMLXPath(root, `//h2//i | //tr`)
	if err != nil {
		return nil, err
	}
	var results []KonamiSet
	var currentYuGiOhVersion YuGiOhVersion
	for _, elem := range elems {
		// check if is `//h2//i`, if text is OCG or TCG then handle
		tag := strings.TrimSpace(elem.Data)
		if tag == "i" {
			text := textproc.HTMLGetText(elem)
			//println("debug //h2//i:", text)
			switch text {
			case "OCG":
				currentYuGiOhVersion = OCG
			case "TCG":
				currentYuGiOhVersion = TCG
			default:
				currentYuGiOhVersion = ""
			}
			continue
		}
		if currentYuGiOhVersion == "" {
			continue
		}
		// parse table rows
		cells, err := textproc.HTMLXPath(elem, `//td`)
		if err != nil {
			return nil, err
		}
		if len(cells) < 5 {
			continue
		}
		// assign cells to KonamiSet fields, cells[4] is Notes, ignore it
		s := KonamiSet{
			YuGiOhVersion: currentYuGiOhVersion,
			Abbreviation:  strings.ReplaceAll(textproc.HTMLGetText(cells[0]), "\n", " "),
			Name:          strings.ReplaceAll(textproc.HTMLGetText(cells[1]), "\n", " "),
			Type:          strings.ReplaceAll(textproc.HTMLGetText(cells[2]), "\n", " "),
			ReleaseDate:   "1970-01-01",
		}

		// parse release date: "4 February 1999", "March 2001", "Unknown 2003", ...

		isParseTimeSuccess := false
		releaseDateStr := textproc.HTMLGetText(cells[3])
		if releaseDateStr == "" {
			continue
		}
		releaseDateStr = strings.TrimPrefix(releaseDateStr, "Unknown ")
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
			rowText := textproc.HTMLGetText(elem)
			println("debug unexpected release date format in row:", rowText)
		}
		results = append(results, s)
	}
	return results, nil
}

type SortKonamiSetsByReleaseDate []KonamiSet

func (s SortKonamiSetsByReleaseDate) Len() int { return len(s) }
func (s SortKonamiSetsByReleaseDate) Less(i, j int) bool {
	if s[i].ReleaseDate == s[j].ReleaseDate {
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
