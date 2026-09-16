package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"github.com/daominah/yugioh_card_editor/pkg/driver/sqlite"
	"github.com/daominah/yugioh_card_editor/pkg/konami"
)

// Adjustable knobs (edit before running):
const (
	outputFileName   = "count_card_group_type.json"
	top10MDFileName  = "top10_monster_stats.md"
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}

	db, err := sqlite.Open(filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer db.Close()

	counts, err := db.GetCardCounts()
	if err != nil {
		log.Fatalf("error db.GetCardCounts: %v", err)
	}

	cmdDir := filepath.Join(projectRoot, "cmd", "count-card-group-stats")

	jsonPath := filepath.Join(cmdDir, outputFileName)
	f, err := os.Create(jsonPath)
	if err != nil {
		log.Fatalf("error os.Create %s: %v", jsonPath, err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(counts); err != nil {
		log.Fatalf("error encoder.Encode: %v", err)
	}

	mdPath := filepath.Join(cmdDir, top10MDFileName)
	if err := writeTop10MD(counts, mdPath); err != nil {
		log.Fatalf("error writeTop10MD: %v", err)
	}

	printCounts(counts)
}

func writeTop10MD(counts konami.CardCounts, outputPath string) error {
	subtypeMonster, _, _ := splitSubtypes(counts.BySubtype)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error os.Create %s: %w", outputPath, err)
	}
	defer f.Close()

	cardFrameEntries := sortStringEntries(subtypeMonster)
	for i := range cardFrameEntries {
		cardFrameEntries[i].key = strings.TrimPrefix(cardFrameEntries[i].key, "Monster")
	}

	total := counts.ByType[konami.Monster]
	updatedOn := time.Now().Format("2006-01")
	fmt.Fprintln(f, "# Top 10 Monster Stats")
	fmt.Fprintf(f, "\nTotal monsters: %d (updated on %s)\n", total, updatedOn)
	cardFrameNote := `Note: "Effect" counts main-deck Effect monsters only.`
	writeMDSection(f, "By Monster Card Frame", cardFrameNote, "", cardFrameEntries, total)
	writeMDSection(f, "By Monster Attribute", "", "", sortStringEntries(counts.ByMonsterAttribute), total)
	writeMDSection(f, "By Monster Type", "", "", sortStringEntries(counts.ByMonsterType), total)
	writeMDSection(f, "By Monster Level/Rank/Link", "", "Level/Rank/Link", sortIntEntries(counts.ByMonsterLevel), total)
	writeMDSection(f, "By Monster ATK", "", "ATK", sortIntEntries(counts.ByMonsterATK), total)
	// Link monsters have no DEF (stored as def=0 with def_str="-").
	// Subtract their count from the DEF=0 bucket.
	defEntries := sortIntEntries(counts.ByMonsterDEF)
	linkCount := counts.BySubtype[konami.MonsterLink]
	for i := range defEntries {
		if defEntries[i].key == "0" {
			defEntries[i].count -= linkCount
			break
		}
	}
	slices.SortFunc(defEntries, func(a, b entry) int {
		if n := cmp.Compare(b.count, a.count); n != 0 {
			return n
		}
		return cmp.Compare(a.key, b.key)
	})
	defSubtitle := fmt.Sprintf("Total monsters (Link excluded): %d", total-linkCount)
	writeMDSection(f, "By Monster DEF", defSubtitle, "DEF", defEntries, total-linkCount)
	return nil
}

func writeMDSection(f *os.File, title string, subtitle string, keyHeader string, entries []entry, total int) {
	if len(entries) > 10 {
		entries = entries[:10]
	}

	// maxKeyLen accounts for the ** bold markers added to each data key.
	maxKeyLen := len(keyHeader)
	for _, e := range entries {
		if n := len(e.key) + 4; n > maxKeyLen {
			maxKeyLen = n
		}
	}
	maxCountLen := len("Count")
	for _, e := range entries {
		if n := len(fmt.Sprintf("%d", e.count)); n > maxCountLen {
			maxCountLen = n
		}
	}
	maxPctLen := len("Share")
	for _, e := range entries {
		if n := len(fmt.Sprintf("%.1f%%", float64(e.count)*100/float64(total))); n > maxPctLen {
			maxPctLen = n
		}
	}

	// `:` + `-`*(len+1) left-aligns; `-`*(len+1) + `:` right-aligns.
	sepKey := ":" + strings.Repeat("-", maxKeyLen+1)
	sepCount := strings.Repeat("-", maxCountLen+1) + ":"
	sepPct := strings.Repeat("-", maxPctLen+1) + ":"

	fmt.Fprintf(f, "\n## %s\n", title)
	fmt.Fprintf(f, "\n| %-*s | %*s | %*s |\n", maxKeyLen, keyHeader, maxCountLen, "Count", maxPctLen, "Share")
	fmt.Fprintf(f, "|%s|%s|%s|\n", sepKey, sepCount, sepPct)
	for _, e := range entries {
		pct := fmt.Sprintf("%.1f%%", float64(e.count)*100/float64(total))
		fmt.Fprintf(f, "| %-*s | %*d | %*s |\n", maxKeyLen, "**"+e.key+"**", maxCountLen, e.count, maxPctLen, pct)
	}
	if subtitle != "" {
		fmt.Fprintf(f, "\n%s\n", subtitle)
	}
}

const separator = "\n" + "________________________________________" + "\n"

func printCounts(counts konami.CardCounts) {
	subtypeMonster, subtypeSpell, subtypeTrap := splitSubtypes(counts.BySubtype)

	printStringMap("ByType", counts.ByType)
	fmt.Print(separator)
	printStringMap("BySubtypeMonster", subtypeMonster)
	fmt.Print(separator)
	printStringMap("BySubtypeSpell", subtypeSpell)
	fmt.Print(separator)
	printStringMap("BySubtypeTrap", subtypeTrap)
	fmt.Print(separator)
	printStringMap("ByMonsterAttribute", counts.ByMonsterAttribute)
	fmt.Print(separator)
	printStringMap("ByMonsterType", counts.ByMonsterType)
	fmt.Print(separator)
	printIntMap("ByMonsterLevel", counts.ByMonsterLevel)
	fmt.Print(separator)
	printIntMap("ByMonsterATK", counts.ByMonsterATK)
	fmt.Print(separator)
	printIntMap("ByMonsterDEF", counts.ByMonsterDEF)
}

func printStringMap[K ~string](label string, m map[K]int) {
	entries := sortStringEntries(m)
	maxKeyLen := 0
	for _, e := range entries {
		if len(e.key) > maxKeyLen {
			maxKeyLen = len(e.key)
		}
	}
	fmt.Printf("%s:\n", label)
	for _, e := range entries {
		fmt.Printf("\t%-*s: %d\n", maxKeyLen, e.key, e.count)
	}
}

func printIntMap(label string, m map[int]int) {
	entries := sortIntEntries(m)
	maxKeyLen := 0
	for _, e := range entries {
		if len(e.key) > maxKeyLen {
			maxKeyLen = len(e.key)
		}
	}
	fmt.Printf("%s:\n", label)
	for _, e := range entries {
		fmt.Printf("\t%-*s: %d\n", maxKeyLen, e.key, e.count)
	}
}

// splitSubtypes partitions BySubtype into three maps by card type prefix.
func splitSubtypes(m map[konami.CardSubtype]int) (monster, spell, trap map[konami.CardSubtype]int) {
	monster = make(map[konami.CardSubtype]int)
	spell = make(map[konami.CardSubtype]int)
	trap = make(map[konami.CardSubtype]int)
	for k, v := range m {
		switch {
		case len(k) >= 7 && k[:7] == "Monster":
			monster[k] = v
		case len(k) >= 5 && k[:5] == "Spell":
			spell[k] = v
		case len(k) >= 4 && k[:4] == "Trap":
			trap[k] = v
		}
	}
	return
}

// sortStringEntries returns m's entries sorted by count desc, key asc.
func sortStringEntries[K ~string](m map[K]int) []entry {
	entries := make([]entry, 0, len(m))
	for k, v := range m {
		entries = append(entries, entry{string(k), v})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if n := cmp.Compare(b.count, a.count); n != 0 {
			return n
		}
		return cmp.Compare(a.key, b.key)
	})
	return entries
}

// sortIntEntries returns m's entries sorted by count desc, key asc.
func sortIntEntries(m map[int]int) []entry {
	type pair struct{ key, count int }
	pairs := make([]pair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, pair{k, v})
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		if n := cmp.Compare(b.count, a.count); n != 0 {
			return n
		}
		return cmp.Compare(a.key, b.key)
	})
	result := make([]entry, len(pairs))
	for i, p := range pairs {
		result[i] = entry{fmt.Sprintf("%d", p.key), p.count}
	}
	return result
}

type entry struct {
	key   string
	count int
}
