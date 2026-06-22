package konami

import (
	"reflect"
	"testing"
)

func TestCardPrintSetCode(t *testing.T) {
	cases := []struct {
		position string
		want     string
	}{
		// regional codes: prefix before the first '-' is the set abbreviation
		{"LOCH-JP077", "LOCH"},
		{"DBLE-KRS03", "DBLE"},
		{"OP28-EN001", "OP28"},
		// no '-' means the whole position is the abbreviation (rare/legacy)
		{"LOB001", "LOB001"},
		// leading '-' has no prefix to extract; return as-is
		{"-EN001", "-EN001"},
		// empty
		{"", ""},
	}
	for _, c := range cases {
		t.Run(c.position, func(t *testing.T) {
			// GIVEN a CardPrint with the given Position
			p := CardPrint{Position: c.position}

			// WHEN SetCode is called
			got := p.SetCode()

			// THEN the result matches the expected set abbreviation
			if got != c.want {
				t.Errorf("Position %q SetCode got %q, want %q", c.position, got, c.want)
			}
		})
	}
}

func TestDedupCardPrints(t *testing.T) {
	t.Run("empty input returns empty slice", func(t *testing.T) {
		// WHEN deduplicating an empty slice
		got := DedupCardPrints(nil)

		// THEN the result is empty (not nil-vs-empty asserted, just length 0)
		if len(got) != 0 {
			t.Errorf("got %d entries, want 0", len(got))
		}
	})

	t.Run("no duplicates: every entry kept in original order", func(t *testing.T) {
		// GIVEN three distinct (Position, RarityCode) pairs
		in := []CardPrint{
			{Position: "POTE-JP001", RarityCode: "SR", RarityName: "スーパーレア仕様"},
			{Position: "POTE-JP001", RarityCode: "SE", RarityName: "シークレットレア仕様"},
			{Position: "POTE-JP002", RarityCode: "N", RarityName: "ノーマル仕様"},
		}

		// WHEN deduplicating
		got := DedupCardPrints(in)

		// THEN all three are returned in the original order
		if !reflect.DeepEqual(got, in) {
			t.Errorf("got %+v, want %+v", got, in)
		}
	})

	t.Run("same key with SPECIAL Ver. subvariant: shortest RarityName wins", func(t *testing.T) {
		// GIVEN one (Position, RarityCode) listed twice on the same Konami page,
		// the second occurrence carrying a "(SPECIAL Ver.)" suffix
		base := CardPrint{
			Position: "QCAC-JP019", RarityCode: "QCSE",
			RarityName: "クォーターセンチュリーシークレットレア仕様",
		}
		special := CardPrint{
			Position: "QCAC-JP019", RarityCode: "QCSE",
			RarityName: "クォーターセンチュリーシークレットレア仕様(SPECIAL Ver.)",
		}

		// WHEN deduplicating with the longer variant first, then the shorter one
		got := DedupCardPrints([]CardPrint{special, base})

		// THEN the shorter (canonical) RarityName is the one kept
		if len(got) != 1 {
			t.Fatalf("got %d entries, want 1", len(got))
		}
		if got[0].RarityName != base.RarityName {
			t.Errorf("RarityName got %q, want %q", got[0].RarityName, base.RarityName)
		}
	})

	t.Run("non-empty RarityName preferred over empty", func(t *testing.T) {
		// GIVEN one entry with empty RarityName and one with a real value,
		// sharing the same (Position, RarityCode)
		empty := CardPrint{Position: "X-001", RarityCode: "UR", RarityName: ""}
		real := CardPrint{Position: "X-001", RarityCode: "UR", RarityName: "Ultra Rare"}

		// WHEN deduplicating with the empty entry encountered first
		got := DedupCardPrints([]CardPrint{empty, real})

		// THEN the non-empty entry wins (shorter empty string does NOT outrank it)
		if len(got) != 1 {
			t.Fatalf("got %d entries, want 1", len(got))
		}
		if got[0].RarityName != real.RarityName {
			t.Errorf("RarityName got %q, want %q", got[0].RarityName, real.RarityName)
		}
	})

	t.Run("different RarityCode on same Position: both kept", func(t *testing.T) {
		// GIVEN one Position printed at two different rarities (Ultra and Secret)
		ur := CardPrint{Position: "POTE-JP001", RarityCode: "SR", RarityName: "スーパーレア仕様"}
		se := CardPrint{Position: "POTE-JP001", RarityCode: "SE", RarityName: "シークレットレア仕様"}

		// WHEN deduplicating
		got := DedupCardPrints([]CardPrint{ur, se})

		// THEN both rows are kept (the PK is composite, not Position alone)
		if len(got) != 2 {
			t.Fatalf("got %d entries, want 2", len(got))
		}
	})

	t.Run("order of remaining entries preserved", func(t *testing.T) {
		// GIVEN a mixed slice where some entries have duplicates and some do not,
		// arranged so the kept entries appear at indices 0, 2, 3 of the input
		in := []CardPrint{
			{Position: "A-001", RarityCode: "N", RarityName: "ノーマル"},          // kept (index 0)
			{Position: "A-001", RarityCode: "N", RarityName: "ノーマル(SPECIAL)"}, // dropped: longer
			{Position: "B-001", RarityCode: "SR", RarityName: "スーパーレア"},       // kept (index 2)
			{Position: "C-001", RarityCode: "UR", RarityName: "ウルトラレア"},       // kept (index 3)
		}

		// WHEN deduplicating
		got := DedupCardPrints(in)

		// THEN the three kept entries appear in their original input order
		if len(got) != 3 {
			t.Fatalf("got %d entries, want 3", len(got))
		}
		wantPositions := []string{"A-001", "B-001", "C-001"}
		for i, p := range got {
			if p.Position != wantPositions[i] {
				t.Errorf("entry %d Position got %q, want %q", i, p.Position, wantPositions[i])
			}
		}
	})
}
