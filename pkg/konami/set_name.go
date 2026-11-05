package konami

import (
	"slices"
	"strings"
)

// SetNameCandidate is one distinct pack name that Konami card pages of one locale
// list under one set code, with the evidence used to rank it.
//
// Konami reuses a set code for several products:
// bonus cards of a bundle ("STBL-JP001" is in both the STARSTRIKE BLAST booster
// and its "2 pack set" bonus card), numbered series (Vol.1 to Vol.4),
// and yearly event families. The sets table has one row per set code,
// so ChooseSetName picks the most representative name.
type SetNameCandidate struct {
	Name       string
	Date       string // earliest print date "YYYY-MM-DD" under this name, "" if unknown
	PrintCount int    // number of card prints that list this name
}

// bonusNameMarkers are substrings of pack names that are bonus cards,
// event giveaways or bundle extras rather than the main product, in JA, KO and EN.
var bonusNameMarkers = []string{
	"特典", "来場", "同梱", "付録", "参加賞", "上位賞",
	"특전", "동봉", "부록", "참가상", "상위상", "내방",
	"ボーナス", "보너스", "+1",
	"bonus", "bundle", "participation", "visitor",
}

// SplitEnglishTitle splits a trailing bracketed English title off a JA or KO pack name:
// "エイジ・オブ・オーバーロード [ AGE OF OVERLORD ]" gives
// ("エイジ・オブ・オーバーロード", "AGE OF OVERLORD").
// Konami writes it on only some names, so the bare name is the consistent form.
// Brackets that are not an English title stay in the name:
// a year ("BEGINNER'S EDITION 1 [ 2011 ]", no letters)
// or a Japanese subtitle ("デュエルロワイヤル デッキセットEX[ ブラック・マジシャン ]").
func SplitEnglishTitle(name string) (base string, english string) {
	if !strings.HasSuffix(name, "]") {
		return name, ""
	}
	open := strings.LastIndex(name, "[")
	if open < 0 {
		return name, ""
	}
	inside := strings.TrimSpace(name[open+1 : len(name)-1])
	base = strings.TrimSpace(name[:open])
	if inside == "" || base == "" {
		return name, ""
	}
	hasLetter := false
	for _, r := range inside {
		if r > 0x7e {
			return name, ""
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
	}
	if !hasLetter {
		return name, ""
	}
	return base, inside
}

func isBonusSetName(name string) bool {
	lower := strings.ToLower(name)
	for _, marker := range bonusNameMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// ChooseSetName returns the candidate that best names the set code,
// or a zero value when candidates is empty. The result does not depend
// on the order of candidates (the crawl visits card pages in random order).
// Preference order:
//  1. A name without a bonus marker: "STARSTRIKE BLAST" over its bonus card.
//  2. A generic name that is a prefix of another candidate's name:
//     "WORLD CHAMPIONSHIP 2017" over "WORLD CHAMPIONSHIP 2017 来場記念カード".
//  3. The name with more prints (the main product).
//  4. The earlier release date.
//  5. The smaller name in byte order, as the final tie-break.
func ChooseSetName(candidates []SetNameCandidate) SetNameCandidate {
	if len(candidates) == 0 {
		return SetNameCandidate{}
	}
	isGeneric := func(c SetNameCandidate) bool {
		for _, other := range candidates {
			if other.Name != c.Name && strings.HasPrefix(other.Name, c.Name) {
				return true
			}
		}
		return false
	}
	// Dates "" (unknown) sort after known dates.
	isEarlier := func(a, b string) bool {
		if a == "" || b == "" {
			return a != "" && b == ""
		}
		return a < b
	}
	better := func(a, b SetNameCandidate) bool {
		if isBonusSetName(a.Name) != isBonusSetName(b.Name) {
			return !isBonusSetName(a.Name)
		}
		if isGeneric(a) != isGeneric(b) {
			return isGeneric(a)
		}
		if a.PrintCount != b.PrintCount {
			return a.PrintCount > b.PrintCount
		}
		if a.Date != b.Date {
			return isEarlier(a.Date, b.Date)
		}
		return a.Name < b.Name
	}
	sorted := slices.Clone(candidates)
	slices.SortFunc(sorted, func(a, b SetNameCandidate) int {
		switch {
		case better(a, b):
			return -1
		case better(b, a):
			return 1
		}
		return 0
	})
	return sorted[0]
}
