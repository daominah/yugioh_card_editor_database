package konami

import (
	"regexp"
	"sort"
	"strings"
)

// Locale field order is JA, KO, EN throughout the codebase (struct fields,
// SQL columns, langs slices, output): JA is the source of truth (earliest
// release on Konami's DB), KO is the next localized variant, and EN is the
// final translation.

// MonsterAttributeRow corresponds to one row of the monster_attributes lookup
// table: a canonical attribute (e.g. "LIGHT") with localized display text.
type MonsterAttributeRow struct {
	Attribute MonsterAttribute
	TextJA    string
	TextKO    string
	TextEN    string
}

// MonsterAttributeRushRow corresponds to one row of monster_attributes_rush.
// Rush has no EN crawl pass, so there is no text_en column.
type MonsterAttributeRushRow struct {
	Attribute MonsterAttribute
	TextJA    string
	TextKO    string
}

// MonsterTypeRow corresponds to one row of monster_types.
type MonsterTypeRow struct {
	MonsterType MonsterType
	TextJA      string
	TextKO      string
	TextEN      string
}

// MonsterTypeRushRow corresponds to one row of monster_types_rush.
type MonsterTypeRushRow struct {
	MonsterType MonsterType
	TextJA      string
	TextKO      string
}

// RarityRow corresponds to one row of the rarities lookup table. CountCards
// is left at zero by AggregateRarities; it is filled in by a separate pass
// (DatabaseAggregate.UpdateRarityCardCounts) that counts distinct card_id
// per rarity_code in set_cards.
type RarityRow struct {
	RarityCode   string
	AliasCode    string
	RarityNameJA string
	RarityNameKO string
	RarityNameEN string
	CountCards   int
}

// EnumVariantInput is one (canonical enum value, lang, localized text, count)
// tuple, the result of grouping cards × card_texts in a SQL GROUP BY. The
// pure aggregate functions consume these to pick the most-frequent text per
// (enum, lang).
type EnumVariantInput struct {
	Canonical string // e.g. "LIGHT", "Dragon"
	Lang      string // "ja", "ko", "en"
	Text      string // localized text from card_texts
	Count     int
}

// RarityVariantInput is one (card_set_code, rarity_code, rarity_name, count)
// tuple from set_cards. The aggregator infers the language from
// card_set_code's region suffix (see inferLangFromSetCode).
type RarityVariantInput struct {
	CardSetCode string
	RarityCode  string
	RarityName  string
	Count       int
}

// TextVariant is a (text, count) pair used to rank localized strings within
// a (key, language) cell.
type TextVariant struct {
	Text  string
	Count int
}

// Anomaly indicates a (key, language) cell that has more than one distinct
// localized text. Variants is sorted by Count descending, so Variants[0] is
// the value the aggregate function picked.
type Anomaly struct {
	Key      string // canonical enum value or rarity_code
	Lang     string
	Variants []TextVariant
}

// AggregateMonsterAttributes picks the most-frequent localized text per
// (attribute, lang) cell across ja/ko/en. Returns one row per distinct
// attribute and an anomaly entry for every cell with more than one variant.
func AggregateMonsterAttributes(inputs []EnumVariantInput) ([]MonsterAttributeRow, []Anomaly) {
	chosen, anomalies := aggregateEnum(inputs, []string{"ja", "ko", "en"})
	rows := make([]MonsterAttributeRow, 0, len(chosen))
	for _, c := range sortedKeys(chosen) {
		rows = append(rows, MonsterAttributeRow{
			Attribute: MonsterAttribute(c),
			TextJA:    chosen[c]["ja"],
			TextKO:    chosen[c]["ko"],
			TextEN:    chosen[c]["en"],
		})
	}
	return rows, anomalies
}

// AggregateMonsterAttributesRush is the Rush variant: ja/ko only.
func AggregateMonsterAttributesRush(inputs []EnumVariantInput) ([]MonsterAttributeRushRow, []Anomaly) {
	chosen, anomalies := aggregateEnum(inputs, []string{"ja", "ko"})
	rows := make([]MonsterAttributeRushRow, 0, len(chosen))
	for _, c := range sortedKeys(chosen) {
		rows = append(rows, MonsterAttributeRushRow{
			Attribute: MonsterAttribute(c),
			TextJA:    chosen[c]["ja"],
			TextKO:    chosen[c]["ko"],
		})
	}
	return rows, anomalies
}

// AggregateMonsterTypes picks the most-frequent localized text per
// (monster_type, lang) cell across ja/ko/en.
func AggregateMonsterTypes(inputs []EnumVariantInput) ([]MonsterTypeRow, []Anomaly) {
	chosen, anomalies := aggregateEnum(inputs, []string{"ja", "ko", "en"})
	rows := make([]MonsterTypeRow, 0, len(chosen))
	for _, c := range sortedKeys(chosen) {
		rows = append(rows, MonsterTypeRow{
			MonsterType: MonsterType(c),
			TextJA:      chosen[c]["ja"],
			TextKO:      chosen[c]["ko"],
			TextEN:      chosen[c]["en"],
		})
	}
	return rows, anomalies
}

// AggregateMonsterTypesRush is the Rush variant: ja/ko only.
func AggregateMonsterTypesRush(inputs []EnumVariantInput) ([]MonsterTypeRushRow, []Anomaly) {
	chosen, anomalies := aggregateEnum(inputs, []string{"ja", "ko"})
	rows := make([]MonsterTypeRushRow, 0, len(chosen))
	for _, c := range sortedKeys(chosen) {
		rows = append(rows, MonsterTypeRushRow{
			MonsterType: MonsterType(c),
			TextJA:      chosen[c]["ja"],
			TextKO:      chosen[c]["ko"],
		})
	}
	return rows, anomalies
}

// AggregateRarities groups the set_cards inputs by (rarity_code, lang) — with
// lang inferred from each row's card_set_code region suffix — picks the
// most-frequent rarity_name per cell, and emits one RarityRow per distinct
// rarity_code. AliasCode is filled from the hard-coded rarityAliases map;
// CountCards is left at zero (compute it separately from set_cards).
//
// skipped is the total Count of input rows whose card_set_code maps to a
// non-ja/ko/en locale (e.g. French "FRP" prints).
func AggregateRarities(inputs []RarityVariantInput) (rows []RarityRow, anomalies []Anomaly, skipped int) {
	langs := []string{"ja", "ko", "en"}
	// counts[lang][rarity_code][name] aggregates input counts across many
	// card_set_codes that share the same (rarity_code, lang, name) triple.
	counts := map[string]map[string]map[string]int{}
	for _, lang := range langs {
		counts[lang] = map[string]map[string]int{}
	}
	for _, in := range inputs {
		lang := inferLangFromSetCode(in.CardSetCode)
		if lang == "" {
			skipped += in.Count
			continue
		}
		if counts[lang][in.RarityCode] == nil {
			counts[lang][in.RarityCode] = map[string]int{}
		}
		counts[lang][in.RarityCode][in.RarityName] += in.Count
	}

	// variants[lang][code] = sorted []TextVariant, most-frequent first.
	variants := map[string]map[string][]TextVariant{}
	for _, lang := range langs {
		variants[lang] = map[string][]TextVariant{}
		for code, byName := range counts[lang] {
			vs := make([]TextVariant, 0, len(byName))
			for name, n := range byName {
				vs = append(vs, TextVariant{Text: name, Count: n})
			}
			sort.Slice(vs, func(i, j int) bool {
				if vs[i].Count != vs[j].Count {
					return vs[i].Count > vs[j].Count
				}
				return vs[i].Text < vs[j].Text
			})
			variants[lang][code] = vs
		}
	}

	// Collect the union of rarity codes across all languages, sorted.
	codeSet := map[string]struct{}{}
	for _, lang := range langs {
		for code := range variants[lang] {
			codeSet[code] = struct{}{}
		}
	}
	codes := make([]string, 0, len(codeSet))
	for c := range codeSet {
		codes = append(codes, c)
	}
	sort.Strings(codes)

	for _, code := range codes {
		row := RarityRow{RarityCode: code, AliasCode: rarityAliases[code]}
		for _, lang := range langs {
			vs := variants[lang][code]
			if len(vs) == 0 {
				continue
			}
			switch lang {
			case "ja":
				row.RarityNameJA = vs[0].Text
			case "ko":
				row.RarityNameKO = vs[0].Text
			case "en":
				row.RarityNameEN = vs[0].Text
			}
			if len(vs) > 1 {
				anomalies = append(anomalies, Anomaly{Key: code, Lang: lang, Variants: vs})
			}
		}
		rows = append(rows, row)
	}
	return rows, anomalies, skipped
}

// aggregateEnum is the shared internal pass for the four enum-text lookups.
// Returns chosen[canonical][lang] = most-frequent text and a flat list of
// anomalies for cells with multiple variants.
func aggregateEnum(inputs []EnumVariantInput, langs []string) (map[string]map[string]string, []Anomaly) {
	// variants[canonical][lang] = list of (text, count) pairs.
	variants := map[string]map[string][]TextVariant{}
	for _, in := range inputs {
		if variants[in.Canonical] == nil {
			variants[in.Canonical] = map[string][]TextVariant{}
		}
		variants[in.Canonical][in.Lang] = append(variants[in.Canonical][in.Lang],
			TextVariant{Text: in.Text, Count: in.Count})
	}
	// Sort each cell's variants by count desc so Variants[0] is the picked one.
	for _, byLang := range variants {
		for _, vs := range byLang {
			sort.Slice(vs, func(i, j int) bool {
				if vs[i].Count != vs[j].Count {
					return vs[i].Count > vs[j].Count
				}
				return vs[i].Text < vs[j].Text
			})
		}
	}

	chosen := map[string]map[string]string{}
	var anomalies []Anomaly
	for _, c := range sortedKeys(variants) {
		chosen[c] = map[string]string{}
		for _, lang := range langs {
			vs := variants[c][lang]
			if len(vs) == 0 {
				continue
			}
			chosen[c][lang] = vs[0].Text
			if len(vs) > 1 {
				anomalies = append(anomalies, Anomaly{Key: c, Lang: lang, Variants: vs})
			}
		}
	}
	return chosen, anomalies
}

// rarityAliases maps each rarity_code to its symmetric peer alias. Pairs are
// bidirectional: if A→B then B→A. Codes absent from this map have no known alias.
var rarityAliases = map[string]string{
	"C":    "N", // EN Common  ↔  JA Normal (ノーマル仕様)
	"N":    "C",
	"PGR":  "PG", // Prismatic God Rare ↔ Platinum Gold Rare
	"PG":   "PGR",
	"STAR": "PSE", // Starlight Rare ↔ Prismatic Secret Rare (older term)
	"PSE":  "STAR",
	"PS":   "EXSE", // Prismatic Secret ↔ Extra Secret Rare
	"EXSE": "PS",
}

// regionRe extracts the alphabetic region prefix between '-' and the first
// digit in a card_set_code, e.g. "LOCH-JP077"→"JP", "DBLE-KRS03"→"KRS".
// Returns empty string for codes with no '-' or no letter prefix (old codes
// like "LOB-001").
var regionRe = regexp.MustCompile(`-([A-Za-z]+)\d`)

// inferLangFromSetCode maps a card_set_code's region prefix to a crawl
// language ("en"/"ja"/"ko"), returning "" when the code does not belong to
// any of the three crawled locales (e.g. French "FRP" prints).
//
// Region prefixes observed in the wild:
//   - "EN", "ENA", "ENB", ..., "AE" → en (Asia-English uses "AE")
//   - "JP", "JPA", "JPB", ..., "B", "J" → ja
//   - "KR", "KRA", "KRS", ..., "K", "SEKR" → ko
//   - empty (codes like "LOB-001"): treated as ja, matching UpsertSetCards'
//     first-write-wins (JA pass crawls first, so the stored name is JA).
func inferLangFromSetCode(code string) string {
	region := ""
	if m := regionRe.FindStringSubmatch(code); m != nil {
		region = strings.ToUpper(m[1])
	}
	switch {
	case region == "AE":
		return "en"
	case region == "SEKR":
		return "ko"
	case strings.HasPrefix(region, "EN"):
		return "en"
	case strings.HasPrefix(region, "JP"):
		return "ja"
	case strings.HasPrefix(region, "KR"):
		return "ko"
	case region == "K":
		return "ko"
	case region == "J" || region == "B" || region == "":
		return "ja"
	default:
		return ""
	}
}

// sortedKeys returns the keys of m in lexicographic order, regardless of the
// value type. Used to produce stable output across runs.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
