package konami

import (
	_ "embed"
	"testing"
)

var (
	//go:embed test_html_ko/test_konami_4007.html
	test_konami_ko_4007 []byte
	// 4007: Blue-Eyes White Dragon (Normal Monster)

	//go:embed test_html_ko/test_konami_4343.html
	test_konami_ko_4343 []byte
	// 4343: Raigeki (Normal Spell)

	//go:embed test_html_ko/test_konami_4960.html
	test_konami_ko_4960 []byte
	// 4960: Imperial Order (Continuous Trap)
)

func TestParseKonamiCardHTML_KO(t *testing.T) {
	// ParseKonamiCardHTML was designed for EN pages: its type/subtype/attribute
	// lookup maps contain only English strings. KO pages still produce
	// CardType "Monster" (inferred from ATK/DEF presence), level, ATK/DEF,
	// and set code, but CardSubtype, MonsterAttribute, MonsterType are empty.
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     Card
	}{
		// GIVEN a KO page for a Normal Monster
		{cardID: "4007", pageHTML: test_konami_ko_4007, want: Card{
			CardName:             "푸른 눈의 백룡",
			CardType:             Monster,
			CardSubtype:          MonsterNormal,
			MonsterAttribute:     LIGHT,
			MonsterType:          Dragon,
			MonsterLevelRankLink: 8,
			MonsterATK:           3000,
			MonsterDEF:           2500,
			IsNonEffectMonster:   true,
			MiscKonamiSet:        "SDK-K001",
			MiscKonamiCardID:     "4007",
			MiscYear:             "2003",
		}},

		// GIVEN a KO page for a Normal Spell
		{cardID: "4343", pageHTML: test_konami_ko_4343, want: Card{
			CardName:         "번개",
			CardType:         Spell,
			CardSubtype:      SpellNormal,
			MiscKonamiSet:    "LOB-K053",
			MiscKonamiCardID: "4343",
			MiscYear:         "2003",
		}},

		// GIVEN a KO page for a Continuous Trap
		{cardID: "4960", pageHTML: test_konami_ko_4960, want: Card{
			CardName:         "왕궁의 칙명",
			CardType:         Trap,
			CardSubtype:      TrapContinuous,
			MiscKonamiSet:    "PSV-K104",
			MiscKonamiCardID: "4960",
			MiscYear:         "2004",
		}},
	} {
		// WHEN parsing the KO card page
		got := ParseKonamiCardHTML(c.pageHTML, CardID(c.cardID))

		// THEN the parseable fields match
		if got.CardName != c.want.CardName {
			t.Errorf("cardID %v CardName got %q, want %q", c.cardID, got.CardName, c.want.CardName)
		}
		if got.CardType != c.want.CardType {
			t.Errorf("cardID %v CardType got %q, want %q", c.cardID, got.CardType, c.want.CardType)
		}
		if got.CardSubtype != c.want.CardSubtype {
			t.Errorf("cardID %v CardSubtype got %q, want %q", c.cardID, got.CardSubtype, c.want.CardSubtype)
		}
		if got.MonsterAttribute != c.want.MonsterAttribute {
			t.Errorf("cardID %v MonsterAttribute got %q, want %q", c.cardID, got.MonsterAttribute, c.want.MonsterAttribute)
		}
		if got.MonsterType != c.want.MonsterType {
			t.Errorf("cardID %v MonsterType got %q, want %q", c.cardID, got.MonsterType, c.want.MonsterType)
		}
		if got.MonsterLevelRankLink != c.want.MonsterLevelRankLink {
			t.Errorf("cardID %v MonsterLevelRankLink got %d, want %d", c.cardID, got.MonsterLevelRankLink, c.want.MonsterLevelRankLink)
		}
		if got.MonsterATK != c.want.MonsterATK {
			t.Errorf("cardID %v MonsterATK got %.0f, want %.0f", c.cardID, got.MonsterATK, c.want.MonsterATK)
		}
		if got.MonsterDEF != c.want.MonsterDEF {
			t.Errorf("cardID %v MonsterDEF got %.0f, want %.0f", c.cardID, got.MonsterDEF, c.want.MonsterDEF)
		}
		if got.IsNonEffectMonster != c.want.IsNonEffectMonster {
			t.Errorf("cardID %v IsNonEffectMonster got %v, want %v", c.cardID, got.IsNonEffectMonster, c.want.IsNonEffectMonster)
		}
		if got.MiscKonamiSet != c.want.MiscKonamiSet {
			t.Errorf("cardID %v MiscKonamiSet got %q, want %q", c.cardID, got.MiscKonamiSet, c.want.MiscKonamiSet)
		}
		if got.MiscKonamiCardID != c.want.MiscKonamiCardID {
			t.Errorf("cardID %v MiscKonamiCardID got %q, want %q", c.cardID, got.MiscKonamiCardID, c.want.MiscKonamiCardID)
		}
		if got.MiscYear != c.want.MiscYear {
			t.Errorf("cardID %v MiscYear got %q, want %q", c.cardID, got.MiscYear, c.want.MiscYear)
		}
	}
}

func TestParseCardLocaleText_KO(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     CardLocaleText
	}{
		// GIVEN a KO monster page with Hangul name and EN pronunciation
		{cardID: "4007", pageHTML: test_konami_ko_4007, want: CardLocaleText{
			Name:              "푸른 눈의 백룡",
			NamePronunciation: "Blue-Eyes White Dragon",
			Effect:            "높은 공격력을 자랑하는 전설의 드래곤. 어떠한 상대라도 분쇄해 버리는 파괴력은 상상을 초월한다.",
			AttributeText:     "빛",
		}},

		// GIVEN a KO spell page
		{cardID: "4343", pageHTML: test_konami_ko_4343, want: CardLocaleText{
			Name:              "번개",
			NamePronunciation: "Raigeki",
			Effect:            "1: 상대 필드의 몬스터를 전부 파괴한다.",
		}},

		// GIVEN a KO trap page
		{cardID: "4960", pageHTML: test_konami_ko_4960, want: CardLocaleText{
			Name:              "왕궁의 칙명",
			NamePronunciation: "Imperial Order",
			Effect:            "이 카드의 컨트롤러는 서로의 스탠바이 페이즈마다 700 LP를 지불한다. 700 LP 지불할 수 없을 경우 이 카드를 파괴한다. 1: 이 카드가 마법 & 함정 존에 존재하는 한, 필드의 모든 마법 카드의 효과는 무효화된다.",
		}},
	} {
		// WHEN parsing locale text from the KO page
		got := ParseCardLocaleText(c.pageHTML, CardID(c.cardID))

		// THEN the name, pronunciation, effect, and attribute text match
		if got.Name != c.want.Name {
			t.Errorf("cardID %v Name got %q, want %q", c.cardID, got.Name, c.want.Name)
		}
		if got.NamePronunciation != c.want.NamePronunciation {
			t.Errorf("cardID %v NamePronunciation got %q, want %q", c.cardID, got.NamePronunciation, c.want.NamePronunciation)
		}
		if got.Effect != c.want.Effect {
			t.Errorf("cardID %v Effect got %q, want %q", c.cardID, got.Effect, c.want.Effect)
		}
		if got.AttributeText != c.want.AttributeText {
			t.Errorf("cardID %v AttributeText got %q, want %q", c.cardID, got.AttributeText, c.want.AttributeText)
		}
	}
}

func TestParseCardPrints_KO(t *testing.T) {
	for _, c := range []struct {
		cardID    string
		pageHTML  []byte
		wantFirst CardPrint
		wantCount int // minimum expected print count
	}{
		// GIVEN a KO monster page with many prints
		{cardID: "4007", pageHTML: test_konami_ko_4007,
			wantFirst: CardPrint{
				Date:       "2026-01-14",
				Position:   "LPST-KR003",
				SetName:    "리미티드 팩 -스탬프 에디션-",
				RarityCode: "UR",
				RarityName: "울트라 레어",
			},
			wantCount: 38,
		},

		// GIVEN a KO spell page
		{cardID: "4343", pageHTML: test_konami_ko_4343,
			wantFirst: CardPrint{
				Date:       "2026-01-14",
				Position:   "LPST-KR028",
				SetName:    "리미티드 팩 -스탬프 에디션-",
				RarityCode: "UR",
				RarityName: "울트라 레어",
			},
			wantCount: 13,
		},

		// GIVEN a KO trap page with fewer prints
		{cardID: "4960", pageHTML: test_konami_ko_4960,
			wantFirst: CardPrint{
				Date:       "2019-06-19",
				Position:   "SD36-KR040",
				SetName:    "스트럭처 덱 - 리볼버 -",
				RarityCode: "N",
				RarityName: "노멀",
			},
			wantCount: 4,
		},
	} {
		// WHEN parsing prints from the KO page
		prints := ParseCardPrints(c.pageHTML, CardID(c.cardID))

		// THEN the print count is at least the expected minimum
		if len(prints) < c.wantCount {
			t.Errorf("cardID %v prints count got %d, want >= %d", c.cardID, len(prints), c.wantCount)
		}

		// THEN the first print matches the expected values
		if len(prints) == 0 {
			t.Errorf("cardID %v got 0 prints", c.cardID)
			continue
		}
		got := prints[0]
		if got.Date != c.wantFirst.Date {
			t.Errorf("cardID %v first print Date got %q, want %q", c.cardID, got.Date, c.wantFirst.Date)
		}
		if got.Position != c.wantFirst.Position {
			t.Errorf("cardID %v first print Position got %q, want %q", c.cardID, got.Position, c.wantFirst.Position)
		}
		if got.SetName != c.wantFirst.SetName {
			t.Errorf("cardID %v first print SetName got %q, want %q", c.cardID, got.SetName, c.wantFirst.SetName)
		}
		if got.RarityCode != c.wantFirst.RarityCode {
			t.Errorf("cardID %v first print RarityCode got %q, want %q", c.cardID, got.RarityCode, c.wantFirst.RarityCode)
		}
		if got.RarityName != c.wantFirst.RarityName {
			t.Errorf("cardID %v first print RarityName got %q, want %q", c.cardID, got.RarityName, c.wantFirst.RarityName)
		}
	}
}
