package konami

import (
	_ "embed"
	"testing"
)

var (
	//go:embed test_html_ja/test_konami_4007.html
	test_konami_ja_4007 []byte
	// 4007: Blue-Eyes White Dragon (Normal Monster)

	//go:embed test_html_ja/test_konami_4343.html
	test_konami_ja_4343 []byte
	// 4343: Raigeki (Normal Spell)

	//go:embed test_html_ja/test_konami_4960.html
	test_konami_ja_4960 []byte
	// 4960: Imperial Order (Continuous Trap)
)

func TestParseKonamiCardHTML_JA(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     Card
	}{
		// GIVEN a JA page for a Normal Monster
		{cardID: "4007", pageHTML: test_konami_ja_4007, want: Card{
			CardName:             "青眼の白龍",
			CardType:             Monster,
			CardSubtype:          MonsterNormal,
			MonsterAttribute:     LIGHT,
			MonsterType:          Dragon,
			MonsterLevelRankLink: 8,
			MonsterATK:           3000,
			MonsterDEF:           2500,
			IsNonEffectMonster:   true,
			MiscKonamiSet:        "LB-01",
			MiscKonamiCardID:     "4007",
			MiscYear:             "2000",
		}},

		// GIVEN a JA page for a Normal Spell
		{cardID: "4343", pageHTML: test_konami_ja_4343, want: Card{
			CardName:         "サンダー・ボルト",
			CardType:         Spell,
			CardSubtype:      SpellNormal,
			MiscKonamiSet:    "LB-52",
			MiscKonamiCardID: "4343",
			MiscYear:         "2000",
		}},

		// GIVEN a JA page for a Continuous Trap
		{cardID: "4960", pageHTML: test_konami_ja_4960, want: Card{
			CardName:         "王宮の勅命",
			CardType:         Trap,
			CardSubtype:      TrapContinuous,
			MiscKonamiSet:    "CA-33",
			MiscKonamiCardID: "4960",
			MiscYear:         "2000",
		}},
	} {
		// WHEN parsing the JA card page
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

func TestParseCardLocaleText_JA(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     CardLocaleText
	}{
		// GIVEN a JA monster page with kanji name, katakana ruby, and EN-name span
		{cardID: "4007", pageHTML: test_konami_ja_4007, want: CardLocaleText{
			Name:              "青眼の白龍",
			NamePronunciation: "ブルーアイズ・ホワイト・ドラゴン",
			NameEnglishOnPage: "Blue-Eyes White Dragon",
			Effect:            "高い攻撃力を誇る伝説のドラゴン。どんな相手でも粉砕する、その破壊力は計り知れない。",
			AttributeText:     "光属性",
			MonsterTypeText:   "ドラゴン族",
		}},

		// GIVEN a JA spell page (no monster slots, so attribute/type text stay empty)
		{cardID: "4343", pageHTML: test_konami_ja_4343, want: CardLocaleText{
			Name:              "サンダー・ボルト",
			NamePronunciation: "サンダー・ボルト",
			NameEnglishOnPage: "Raigeki",
			Effect:            "1:相手フィールドのモンスターを全て破壊する。",
		}},

		// GIVEN a JA trap page
		{cardID: "4960", pageHTML: test_konami_ja_4960, want: CardLocaleText{
			Name:              "王宮の勅命",
			NamePronunciation: "おうきゅうのちょくめい",
			NameEnglishOnPage: "Imperial Order",
			Effect:            "このカードのコントローラーはお互いのスタンバイフェイズ毎に700LPを払う。700LP払えない場合このカードを破壊する。1:このカードが魔法&罠ゾーンに存在する限り、フィールドの全ての魔法カードの効果は無効化される。",
		}},
	} {
		// WHEN parsing locale text from the JA page
		got := ParseCardLocaleText(c.pageHTML, CardID(c.cardID))

		// THEN every locale field matches: name, kana pronunciation, EN-name span, effect, attribute, type
		if got.Name != c.want.Name {
			t.Errorf("cardID %v Name got %q, want %q", c.cardID, got.Name, c.want.Name)
		}
		if got.NamePronunciation != c.want.NamePronunciation {
			t.Errorf("cardID %v NamePronunciation got %q, want %q", c.cardID, got.NamePronunciation, c.want.NamePronunciation)
		}
		if got.NameEnglishOnPage != c.want.NameEnglishOnPage {
			t.Errorf("cardID %v NameEnglishOnPage got %q, want %q", c.cardID, got.NameEnglishOnPage, c.want.NameEnglishOnPage)
		}
		if got.Effect != c.want.Effect {
			t.Errorf("cardID %v Effect got %q, want %q", c.cardID, got.Effect, c.want.Effect)
		}
		if got.AttributeText != c.want.AttributeText {
			t.Errorf("cardID %v AttributeText got %q, want %q", c.cardID, got.AttributeText, c.want.AttributeText)
		}
		if got.MonsterTypeText != c.want.MonsterTypeText {
			t.Errorf("cardID %v MonsterTypeText got %q, want %q", c.cardID, got.MonsterTypeText, c.want.MonsterTypeText)
		}
	}
}

func TestParseCardPrints_JA(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		// first expected print (earliest in page order, which is newest release)
		wantFirst CardPrint
		wantCount int // minimum expected print count
	}{
		// GIVEN a JA monster page with many prints
		{cardID: "4007", pageHTML: test_konami_ja_4007,
			wantFirst: CardPrint{
				Date:       "2025-12-13",
				Position:   "LPST-JP003",
				SetName:    "LIMITED PACK ー STAMP EDITION ー",
				RarityCode: "UR",
				RarityName: "ウルトラレア仕様",
			},
			wantCount: 60,
		},

		// GIVEN a JA spell page
		{cardID: "4343", pageHTML: test_konami_ja_4343,
			wantFirst: CardPrint{
				Date:       "2025-12-13",
				Position:   "LPST-JP028",
				SetName:    "LIMITED PACK ー STAMP EDITION ー",
				RarityCode: "UR",
				RarityName: "ウルトラレア仕様",
			},
			wantCount: 20,
		},

		// GIVEN a JA trap page with fewer prints
		{cardID: "4960", pageHTML: test_konami_ja_4960,
			wantFirst: CardPrint{
				Date:       "2019-06-22",
				Position:   "SD36-JP040",
				SetName:    "ストラクチャーデッキ - リボルバー -",
				RarityCode: "N",
				RarityName: "ノーマル仕様",
			},
			wantCount: 6,
		},
	} {
		// WHEN parsing prints from the JA page
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
