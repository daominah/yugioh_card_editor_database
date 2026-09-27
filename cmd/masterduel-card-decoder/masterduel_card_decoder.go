package main

// Ranks the next guesses of a Master Duel card decoder round
// by how well each splits the remaining candidates,
// following masterduel-card-decoder.md in this directory.

import (
	"cmp"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
	_ "modernc.org/sqlite"
)

// Inputs (edit for each guess).
// confirmed holds at most 1 value per stat; excluded piles up the unmatched values.
// Use UndefinedBattleStat for an ATK or DEF of "?".
var (
	confirmed = Filter{
		Types: []konami.MonsterType{konami.Dinosaur},
		ATKs:  []int{2000},
	}
	excluded = Filter{
		Attributes: []konami.MonsterAttribute{konami.EARTH},
		Frames:     []CardFrame{Effect},
		Levels:     []int{4},
		DEFs:       []int{0},
	}
	// excludedCandidates drops cards the game lacks (usually new cards).
	// The filter uses only the card ID key; the English name is for a human double check.
	// Each comment is the card's first release date.
	excludedCandidates = map[int]string{
		10112: "Holactie the Creator of Light",  // 2011-12-10
		14367: "Exodia, the Legendary Defender", // 2019-02-09
		22952: "Celtic Mystic",                  // 2026-04-25
		23359: "D-HERO ドレッドノートガイ",               // 2026-07-18
	}
)

// Adjustable knobs (edit before running):
const (
	suggestionLimit = 8
	topValuesLimit  = 12
	// The partition check compares every candidate against every candidate,
	// so its duration grows with the square of the candidate count:
	// about 2s for 2600 candidates, 24s for the whole pool of ~10000 monsters.
	// Over this count the check still runs, after a warning.
	partitionCandidatesLimit = 3000
	// A hint is scarce, so recommend it only when its average left
	// is at most this fraction of the best guess's.
	hintClearWinRatio = 0.6
)

func main() {
	monsters, err := loadMonsters()
	if err != nil {
		log.Fatalf("error loadMonsters: %v", err)
	}
	candidates := filterCandidates(monsters)
	fmt.Printf("candidates: %d\n", len(candidates))

	unconfirmed := listUnconfirmedStats()
	printTopValues(candidates, unconfirmed)

	if len(candidates) == 0 {
		log.Fatalf("error no candidates: check the confirmed and excluded filters")
	}
	hintAverageLeft := printHints(candidates, unconfirmed)
	warnIfSlow(len(candidates))
	guesses := groupSameStats(candidates, unconfirmed)
	partitions := partitionGuesses(guesses, candidates, unconfirmed)

	// Neither order is proven better, so when their top guesses differ,
	// print both and let the player pick.
	byAverageLeft := sortPartitions(partitions, compareAverageLeftFirst)
	byWorstCase := sortPartitions(partitions, compareWorstCaseFirst)
	printPartitions("average left first", byAverageLeft, len(candidates), unconfirmed)
	if byWorstCase[0].Guess != byAverageLeft[0].Guess {
		printPartitions("worst case first", byWorstCase, len(candidates), unconfirmed)
	}
	printNextAction(hintAverageLeft, byAverageLeft[0], len(candidates))
}

func loadMonsters() ([]Monster, error) {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		return nil, fmt.Errorf("error base.GetProjectRootDir: %w", err)
	}
	db, err := sql.Open("sqlite", filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		return nil, fmt.Errorf("error sql.Open: %w", err)
	}
	defer db.Close()
	monsters, err := readMonsters(db)
	if err != nil {
		return nil, fmt.Errorf("error readMonsters: %w", err)
	}
	return monsters, nil
}

func readMonsters(db *sql.DB) ([]Monster, error) {
	rows, err := db.Query(`
		SELECT c.card_id, COALESCE(NULLIF(t.name, ''), c.card_name_en), c.attribute,
		       c.card_subtype, c.is_pendulum, c.monster_type, c.level_rank_link,
		       c.atk, c.atk_str, c.def, c.def_str
		FROM cards c
		LEFT JOIN card_texts t ON t.card_id = c.card_id AND t.lang = 'en'
		WHERE c.card_type = 'Monster'`)
	if err != nil {
		return nil, fmt.Errorf("error db.Query: %w", err)
	}
	defer rows.Close()
	var monsters []Monster
	for rows.Next() {
		var (
			m                               Monster
			attribute, subtype, monsterType string
			isPendulum, level, atk, def     int
			atkStr, defStr                  string
		)
		err := rows.Scan(&m.CardID, &m.Name, &attribute, &subtype, &isPendulum,
			&monsterType, &level, &atk, &atkStr, &def, &defStr)
		if err != nil {
			return nil, fmt.Errorf("error rows.Scan: %w", err)
		}
		// The game treats Pendulum as its own frame.
		frame := CardFrame(strings.TrimPrefix(subtype, "Monster"))
		if isPendulum == 1 {
			frame = Pendulum
		}
		if atkStr == "?" {
			atk = UndefinedBattleStat
		}
		if defStr == "?" {
			def = UndefinedBattleStat
		}
		m.Stats = map[Stat]string{
			Attribute: attribute,
			Frame:     string(frame),
			Type:      monsterType,
			Level:     strconv.Itoa(level),
			ATK:       formatBattleStat(atk),
			DEF:       formatBattleStat(def),
		}
		monsters = append(monsters, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error rows.Err: %w", err)
	}
	return monsters, nil
}

func formatBattleStat(value int) string {
	if value == UndefinedBattleStat {
		return "?"
	}
	return strconv.Itoa(value)
}

func filterCandidates(monsters []Monster) []Monster {
	confirmedValues, excludedValues := confirmed.values(), excluded.values()
	var candidates []Monster
	for _, m := range monsters {
		if isCandidate(m, confirmedValues, excludedValues) {
			candidates = append(candidates, m)
		}
	}
	return candidates
}

func isCandidate(m Monster, confirmedValues, excludedValues map[Stat][]string) bool {
	if _, isExcluded := excludedCandidates[m.CardID]; isExcluded {
		return false
	}
	for s, values := range confirmedValues {
		if !slices.Contains(values, m.Stats[s]) {
			return false
		}
	}
	for s, values := range excludedValues {
		if slices.Contains(values, m.Stats[s]) {
			return false
		}
	}
	return true
}

func listUnconfirmedStats() []Stat {
	confirmedValues := confirmed.values()
	var unconfirmed []Stat
	for _, s := range allStats {
		if len(confirmedValues[s]) == 0 {
			unconfirmed = append(unconfirmed, s)
		}
	}
	return unconfirmed
}

// printTopValues shows the most common values of each unconfirmed stat,
// for the "Remaining candidates stats" table of the round notes.
func printTopValues(candidates []Monster, unconfirmed []Stat) {
	fmt.Println("\n--- top values ---")
	for _, s := range unconfirmed {
		counts := map[string]int{}
		for _, c := range candidates {
			counts[c.Stats[s]]++
		}
		var values []string
		for v := range counts {
			values = append(values, v)
		}
		slices.SortFunc(values, func(a, b string) int {
			return cmp.Or(cmp.Compare(counts[b], counts[a]), cmp.Compare(a, b))
		})
		var parts []string
		for _, v := range values[:min(len(values), topValuesLimit)] {
			parts = append(parts, fmt.Sprintf("%s (%d)", v, counts[v]))
		}
		fmt.Printf("%-9s %s\n", s, strings.Join(parts, ", "))
	}
}

// printHints shows the average left and worst case after a hint on each unconfirmed stat.
// Returns the average over those stats, as the game picks the hinted stat.
func printHints(candidates []Monster, unconfirmed []Stat) float64 {
	fmt.Println("\n--- hints (average left, worst case) ---")
	var averageSum float64
	for _, s := range unconfirmed {
		counts := map[string]int{}
		for _, c := range candidates {
			counts[c.Stats[s]]++
		}
		squareSum, worstCase := 0, 0
		for _, count := range counts {
			squareSum += count * count
			worstCase = max(worstCase, count)
		}
		average := float64(squareSum) / float64(len(candidates))
		averageSum += average
		fmt.Printf("%-9s average left=%.2f, worst=%d\n", s, average, worstCase)
	}
	if len(unconfirmed) == 0 {
		return 0
	}
	randomAverage := averageSum / float64(len(unconfirmed))
	fmt.Printf("%-9s average left=%.2f\n", "random", randomAverage)
	return randomAverage
}

// printNextAction prints the hint or guess decision last, where the player looks.
func printNextAction(hintAverageLeft float64, bestGuess Partition, candidatesCount int) {
	guessAverageLeft := float64(bestGuess.SquareSum) / float64(candidatesCount)
	fmt.Println("\n--- next action ---")
	if hintAverageLeft > 0 && hintAverageLeft <= hintClearWinRatio*guessAverageLeft {
		fmt.Printf("REQUEST A HINT: leaves %.2f on average, the best guess %s leaves %.2f\n",
			hintAverageLeft, bestGuess.Guess.Cards[0].Name, guessAverageLeft)
		return
	}
	fmt.Printf("GUESS %s: leaves %.2f on average, a hint leaves %.2f\n",
		bestGuess.Guess.Cards[0].Name, guessAverageLeft, hintAverageLeft)
}

func warnIfSlow(candidatesCount int) {
	if candidatesCount > partitionCandidatesLimit {
		log.Printf("warning: %d candidates is over %d, the partition check can take a while",
			candidatesCount, partitionCandidatesLimit)
	}
}

// groupSameStats collapses cards with identical unconfirmed stats into one guess,
// led by the lowest card ID (older cards are usually better known).
func groupSameStats(candidates []Monster, unconfirmed []Stat) []*Group {
	groups := map[string]*Group{}
	var sorted []*Group
	for _, c := range candidates {
		var keyParts []string
		for _, s := range unconfirmed {
			keyParts = append(keyParts, c.Stats[s])
		}
		key := strings.Join(keyParts, "|")
		g, ok := groups[key]
		if !ok {
			g = &Group{Stats: c.Stats}
			groups[key] = g
			sorted = append(sorted, g)
		}
		g.Cards = append(g.Cards, c)
	}
	for _, g := range sorted {
		slices.SortFunc(g.Cards, func(a, b Monster) int { return cmp.Compare(a.CardID, b.CardID) })
	}
	return sorted
}

func partitionGuesses(guesses []*Group, candidates []Monster, unconfirmed []Stat) []Partition {
	var partitions []Partition
	for _, guess := range guesses {
		partitions = append(partitions, partitionCandidates(guess, candidates, unconfirmed))
	}
	return partitions
}

// partitionCandidates treats each candidate as the hidden monster
// and groups the candidates by the match pattern the guess would get.
func partitionCandidates(guess *Group, candidates []Monster, unconfirmed []Stat) Partition {
	p := Partition{Guess: guess, Groups: map[string][]Monster{}}
	for _, hidden := range candidates {
		pattern := ""
		for _, s := range unconfirmed {
			if guess.Stats[s] == hidden.Stats[s] {
				pattern += "1"
			} else {
				pattern += "0"
			}
		}
		p.Groups[pattern] = append(p.Groups[pattern], hidden)
	}
	// Matching every stat wins the round, so that group leaves nothing to guess.
	solved := strings.Repeat("1", len(unconfirmed))
	for pattern, group := range p.Groups {
		if pattern == solved {
			continue
		}
		p.WorstCase = max(p.WorstCase, len(group))
		p.SquareSum += len(group) * len(group)
	}
	return p
}

func sortPartitions(partitions []Partition, compare func(a, b Partition) int) []Partition {
	sorted := slices.Clone(partitions)
	slices.SortFunc(sorted, compare)
	return sorted
}

// compareAverageLeftFirst leaves the fewest cards on average,
// accepting a worse unlucky result.
// SquareSum / len(candidates) is the average left, over every candidate as the hidden monster,
// so comparing SquareSum ranks by that average.
func compareAverageLeftFirst(a, b Partition) int {
	return cmp.Or(
		cmp.Compare(a.SquareSum, b.SquareSum),
		cmp.Compare(a.WorstCase, b.WorstCase),
		cmp.Compare(a.Guess.Cards[0].CardID, b.Guess.Cards[0].CardID))
}

// compareWorstCaseFirst limits how many cards an unlucky result leaves.
func compareWorstCaseFirst(a, b Partition) int {
	return cmp.Or(
		cmp.Compare(a.WorstCase, b.WorstCase),
		cmp.Compare(a.SquareSum, b.SquareSum),
		cmp.Compare(a.Guess.Cards[0].CardID, b.Guess.Cards[0].CardID))
}

func printPartitions(order string, partitions []Partition, candidatesCount int, unconfirmed []Stat) {
	fmt.Printf("\n--- partitions, %s (average left, worst case) ---\n", order)
	for _, p := range partitions[:min(len(partitions), suggestionLimit)] {
		var statParts []string
		for _, s := range unconfirmed {
			statParts = append(statParts, fmt.Sprintf("%s=%s", s, p.Guess.Stats[s]))
		}
		var same []string
		for _, c := range p.Guess.Cards[1:] {
			same = append(same, fmt.Sprintf("%s [%d]", c.Name, c.CardID))
		}
		fmt.Printf("%s [%d] | %s | average left=%.2f, worst=%d | same: %s\n",
			p.Guess.Cards[0].Name, p.Guess.Cards[0].CardID, strings.Join(statParts, ", "),
			float64(p.SquareSum)/float64(candidatesCount), p.WorstCase, strings.Join(same, "; "))
		if candidatesCount <= 20 {
			printMatchGroups(p, unconfirmed)
		}
	}
}

// printMatchGroups lists which candidates each match pattern would leave,
// only useful when candidates are few enough to read.
func printMatchGroups(p Partition, unconfirmed []Stat) {
	var patterns []string
	for pattern := range p.Groups {
		patterns = append(patterns, pattern)
	}
	slices.Sort(patterns)
	for _, pattern := range patterns {
		var matched []string
		for j, s := range unconfirmed {
			if pattern[j] == '1' {
				matched = append(matched, string(s))
			}
		}
		var names []string
		for _, c := range p.Groups[pattern] {
			names = append(names, c.Name)
		}
		fmt.Printf("    matched [%s] -> %s\n", strings.Join(matched, ", "), strings.Join(names, "; "))
	}
}

// allStats lists the 6 stats the game compares, in the game's display order.
var allStats = []Stat{Attribute, Frame, Type, Level, ATK, DEF}

// Stat is one of the 6 stats the game compares between the guess and the hidden monster.
type Stat string

// Stat enum values.
const (
	Attribute Stat = "Attribute"
	Frame     Stat = "Frame"
	Type      Stat = "Type"
	Level     Stat = "Level"
	ATK       Stat = "ATK"
	DEF       Stat = "DEF"
)

// CardFrame is the monster frame as the game shows it,
// which differs from konami.CardSubtype: no "Monster" prefix, and Pendulum is its own frame.
type CardFrame string

// CardFrame enum values.
const (
	Normal   CardFrame = "Normal"
	Effect   CardFrame = "Effect"
	Ritual   CardFrame = "Ritual"
	Fusion   CardFrame = "Fusion"
	Synchro  CardFrame = "Synchro"
	Xyz      CardFrame = "Xyz"
	Link     CardFrame = "Link"
	Pendulum CardFrame = "Pendulum"
)

// UndefinedBattleStat stands for an ATK or DEF printed as "?".
// The value is not unknown but defined by the card's effect,
// and counts as 0 wherever that effect does not apply.
const UndefinedBattleStat = -1

// Filter lists stat values; an empty field does not filter that stat.
type Filter struct {
	Attributes []konami.MonsterAttribute
	Frames     []CardFrame
	Types      []konami.MonsterType
	Levels     []int
	ATKs       []int
	DEFs       []int
}

// values converts the filter to the same wording as Monster.Stats.
func (f Filter) values() map[Stat][]string {
	v := map[Stat][]string{}
	for _, a := range f.Attributes {
		v[Attribute] = append(v[Attribute], string(a))
	}
	for _, frame := range f.Frames {
		v[Frame] = append(v[Frame], string(frame))
	}
	for _, t := range f.Types {
		v[Type] = append(v[Type], string(t))
	}
	for _, level := range f.Levels {
		v[Level] = append(v[Level], strconv.Itoa(level))
	}
	for _, atk := range f.ATKs {
		v[ATK] = append(v[ATK], formatBattleStat(atk))
	}
	for _, def := range f.DEFs {
		v[DEF] = append(v[DEF], formatBattleStat(def))
	}
	return v
}

// Monster holds the stats of one monster card as the game shows them.
type Monster struct {
	CardID int
	Name   string
	Stats  map[Stat]string
}

// Group is a set of candidates sharing identical unconfirmed stats,
// so guessing any of them gets the same result.
type Group struct {
	Stats map[Stat]string
	Cards []Monster
}

// Partition groups the candidates by the match pattern one guess would get.
type Partition struct {
	Guess     *Group
	Groups    map[string][]Monster // key: "1" or "0" per unconfirmed stat
	WorstCase int                  // largest group left to guess, excluding the win
	SquareSum int                  // sum of group sizes squared, excluding the win
}
