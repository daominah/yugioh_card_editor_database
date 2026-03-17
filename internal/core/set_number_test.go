package core

import (
	_ "embed"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"testing"
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
	sort.Sort(SortKonamiSetsByReleaseDate(konamiSets))
	outputCSVFile, err := os.Create("yugioh_sets.csv")
	if err != nil {
		t.Fatalf("error creating CSV file: %v", err)
	}
	csvWriter := csv.NewWriter(outputCSVFile)
	err = csvWriter.WriteAll(MarshalKonamiSetsToCSV(konamiSets))
	if err != nil {
		t.Fatalf("error csv.NewWriter.WriteAll: %v", err)
	}
	absPath, err := filepath.Abs(outputCSVFile.Name())
	if err != nil {
		t.Fatalf("error outputCSVFile absolute path: %v", err)
	}
	outputCSVFile.Close()
	t.Logf("wrote Konami sets data too: %v", absPath)

	mapSetsAbbreviationName := UnmarshalCSVToMapSetAbbreviationToName(MarshalKonamiSetsToCSV(konamiSets))
	mapYugiohVersion := make(map[string]YuGiOhVersion)
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
		YuGiOhVersion YuGiOhVersion
	}{
		{"LB", "Legend of Blue Eyes White Dragon", OCG},
		{"LOB", "Legend of Blue Eyes White Dragon", TCG},
		{"SD47", "Structure Deck: Advent of the Eyes of Blue", OCG},
		{"SDWD", "Structure Deck: Blue-Eyes White Destiny", TCG},
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
