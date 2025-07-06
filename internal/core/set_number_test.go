package core

import (
	_ "embed"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

//go:embed set_number_ocg.html
var yugipediaSetChronologyOCG []byte // from https://yugipedia.com/wiki/Set_chronology#OCG
//go:embed set_number_tcg.html
var yugipediaSetChronologyTCG []byte // from https://yugipedia.com/wiki/Set_chronology#TCG

func TestParseYugipediaSetChronology(t *testing.T) {
	var allVersionsKonamiSets []KonamiSet
	for _, yugipediaSetChronology := range [][]byte{
		yugipediaSetChronologyOCG,
		yugipediaSetChronologyTCG,
	} {
		t.Logf("yugipediaSetChronology size: %v", len(yugipediaSetChronology))
		if len(yugipediaSetChronology) == 0 {
			t.Fatalf("yugipediaSetChronology is empty")
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
			if i < 5 {
				t.Logf("konamiSet: %+v", konamiSet)
			}
			if konamiSet.ReleaseDate == "1970-01-01" {
				t.Fatalf("konamiSet.ReleaseDate is the default value, probably not parsed correctly")
			}
		}
		allVersionsKonamiSets = append(allVersionsKonamiSets, konamiSets...)
	}

	// write to a CSV file
	sort.Sort(SortKonamiSetsByReleaseDate(allVersionsKonamiSets))
	outputCSVFile, err := os.Create("yugioh_sets.csv")
	if err != nil {
		t.Fatalf("error creating CSV file: %v", err)
	}
	csvWriter := csv.NewWriter(outputCSVFile)
	err = csvWriter.WriteAll(MarshalKonamiSetsToCSV(allVersionsKonamiSets))
	if err != nil {
		t.Fatalf("error csv.NewWriter.WriteAll: %v", err)
	}
	absPath, err := filepath.Abs(outputCSVFile.Name())
	if err != nil {
		t.Fatalf("error outputCSVFile absolute path: %v", err)
	}
	outputCSVFile.Close()
	t.Logf("wrote Konami sets data too: %v", absPath)

	mapSetsAbbreviationName := UnmarshalCSVToMapSetAbbreviationToName(MarshalKonamiSetsToCSV(allVersionsKonamiSets))
	for _, c := range []struct {
		Abbreviation string
		Name         string
	}{
		{"LB", "Legend of Blue Eyes White Dragon"},
		{"LOB", "Legend of Blue Eyes White Dragon"},
		{"SD47", "Structure Deck: Advent of the Eyes of Blue"},
		{"SDWD", "Structure Deck: Blue-Eyes White Destiny"},
		{"POTE", "Power of the Elements"},
	} {
		if name, ok := mapSetsAbbreviationName[c.Abbreviation]; !ok || name != c.Name {
			t.Fatalf("error UnmarshalCSVToMapSetAbbreviationToName got: %v, want: %v", name, c.Name)
		}
	}
}
