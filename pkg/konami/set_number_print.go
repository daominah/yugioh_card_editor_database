package konami

// YuGiOhVersion enum, can be OCG, TCG, RushDuel, ...
type YuGiOhVersion string

// YuGiOhVersion enum values
const (
	OCG YuGiOhVersion = "OCG"
	TCG YuGiOhVersion = "TCG"
)

// KonamiSet represents one Yu-Gi-Oh set (booster pack, structure deck, etc.).
// Shape mirrors the SQL `sets` table: Abbreviation = set_code,
// YuGiOhVersion = game_version, NameJA/NameKO/NameEN = name_{ja,ko,en},
// ReleaseDate = release_date. Per-locale name fields default to empty when
// no crawled card page of that locale lists the set
// (for example an OCG-only set has no NameEN).
//
// Locale field order is JA, KO, EN throughout the codebase (struct fields,
// SQL columns, langs slices, output): JA is the source of truth (earliest
// release on Konami's DB), KO is the next localized variant, and EN is the
// final translation.
type KonamiSet struct {
	YuGiOhVersion YuGiOhVersion
	Abbreviation  string
	NameJA        string
	NameKO        string
	NameEN        string
	ReleaseDate   string // YYYY-MM-DD
}

type SortKonamiSetsByReleaseDate []KonamiSet

func (s SortKonamiSetsByReleaseDate) Len() int { return len(s) }
func (s SortKonamiSetsByReleaseDate) Less(i, j int) bool {
	if s[i].ReleaseDate == s[j].ReleaseDate {
		if s[i].NameEN == s[j].NameEN {
			return s[i].YuGiOhVersion < s[j].YuGiOhVersion
		}
		return s[i].NameEN < s[j].NameEN
	}
	return s[i].ReleaseDate < s[j].ReleaseDate
}
func (s SortKonamiSetsByReleaseDate) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// MarshalKonamiSetsToCSV output can be used for csv.Writer.WriteAll
func MarshalKonamiSetsToCSV(sets []KonamiSet) [][]string {
	outputFields := []string{
		"YuGiOhVersion", "Abbreviation", "NameJA", "NameKO", "NameEN", "ReleaseDate",
	}
	records := [][]string{outputFields}
	for _, s := range sets {
		record := []string{
			string(s.YuGiOhVersion),
			s.Abbreviation,
			s.NameJA,
			s.NameKO,
			s.NameEN,
			s.ReleaseDate,
		}
		records = append(records, record)
	}
	return records
}
