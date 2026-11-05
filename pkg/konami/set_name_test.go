package konami

import "testing"

func TestSplitEnglishTitle(t *testing.T) {
	for _, c := range []struct {
		name, wantBase, wantEnglish string
	}{
		{"エイジ・オブ・オーバーロード [ AGE OF OVERLORD ]", "エイジ・オブ・オーバーロード", "AGE OF OVERLORD"},
		{"イモータル・フェニックス [IMMORTAL PHOENIX]", "イモータル・フェニックス", "IMMORTAL PHOENIX"},
		{"BEGINNER'S EDITION 1 [ 2011 ]", "BEGINNER'S EDITION 1 [ 2011 ]", ""},
		{"デュエルロワイヤル デッキセットEX[ ブラック・マジシャン ]", "デュエルロワイヤル デッキセットEX[ ブラック・マジシャン ]", ""},
		{"EX-R [ 二人対戦用構築済みデッキセット ]", "EX-R [ 二人対戦用構築済みデッキセット ]", ""},
		{"スターストライク・ブラスト", "スターストライク・ブラスト", ""},
		{"[ ONLY BRACKET ]", "[ ONLY BRACKET ]", ""},
		{"", "", ""},
	} {
		// GIVEN a pack name, WHEN the English title is split off
		base, english := SplitEnglishTitle(c.name)
		// THEN only a Latin-script bracket title is removed from the name
		if base != c.wantBase || english != c.wantEnglish {
			t.Errorf("%q: got (%q, %q), want (%q, %q)",
				c.name, base, english, c.wantBase, c.wantEnglish)
		}
	}
}

func TestChooseSetName(t *testing.T) {
	for _, c := range []struct {
		desc       string
		candidates []SetNameCandidate
		want       string
	}{
		{"a bundle bonus card loses to the main booster",
			[]SetNameCandidate{
				{"STARSTRIKE BLAST 2パックセット 特典カード", "2010-08-26", 30},
				{"スターストライク・ブラスト [ STARSTRIKE BLAST ]", "2010-07-17", 5},
			}, "スターストライク・ブラスト [ STARSTRIKE BLAST ]"},
		{"the bonus loses even when it has more prints",
			[]SetNameCandidate{
				{"Yu-Gi-Oh! WORLD CHAMPIONSHIP 2017 来場記念カード", "2017-01-01", 3},
				{"Yu-Gi-Oh! WORLD CHAMPIONSHIP 2017", "2017-01-01", 2},
			}, "Yu-Gi-Oh! WORLD CHAMPIONSHIP 2017"},
		{"a katakana bonus pack loses even with more prints",
			[]SetNameCandidate{
				{"エイジ・オブ・オーバーロード　+1ボーナスパック", "2018-01-01", 131},
				{"エイジ・オブ・オーバーロード", "2018-01-01", 41},
			}, "エイジ・オブ・オーバーロード"},
		{"a generic name beats its qualified variant",
			[]SetNameCandidate{
				{"Loppi限定 20th ANNIVERSARY デュエルセット（オシリスの天空竜）", "2019-09-26", 1},
				{"Loppi限定 20th ANNIVERSARY デュエルセット", "2018-10-31", 1},
			}, "Loppi限定 20th ANNIVERSARY デュエルセット"},
		{"more prints win among equal kinds of names",
			[]SetNameCandidate{
				{"20th ANNIVERSARY DUELIST BOX", "2018-01-01", 42},
				{"20th ANNIVERSARY LEGEND COLLECTION", "2018-01-02", 200},
			}, "20th ANNIVERSARY LEGEND COLLECTION"},
		{"an earlier date wins on equal prints",
			[]SetNameCandidate{
				{"トーナメントパック 2017 Vol.3", "2017-07-01", 14},
				{"トーナメントパック 2017 Vol.2", "2017-04-01", 14},
			}, "トーナメントパック 2017 Vol.2"},
		{"an unknown date loses to a known date",
			[]SetNameCandidate{
				{"B", "", 1},
				{"A", "2020-01-01", 1},
			}, "A"},
		{"equal evidence falls back to byte order",
			[]SetNameCandidate{
				{"B", "2020-01-01", 1},
				{"A", "2020-01-01", 1},
			}, "A"},
		{"a single candidate is returned as is",
			[]SetNameCandidate{{"ONLY", "", 0}}, "ONLY"},
		{"no candidate gives a zero value", nil, ""},
	} {
		// GIVEN candidates in one order, WHEN the name is chosen
		got := ChooseSetName(c.candidates).Name
		// THEN the most representative name wins
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.desc, got, c.want)
		}
		// AND the result does not depend on the input order
		reversed := make([]SetNameCandidate, len(c.candidates))
		for i, v := range c.candidates {
			reversed[len(c.candidates)-1-i] = v
		}
		if got2 := ChooseSetName(reversed).Name; got2 != got {
			t.Errorf("%s: reversed input chose %q, want %q", c.desc, got2, got)
		}
	}
}
