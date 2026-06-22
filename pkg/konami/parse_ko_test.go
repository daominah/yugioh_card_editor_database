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

	//go:embed test_html_ko/test_konami_4386.html
	test_konami_ko_4386 []byte
	// 4386: Blue-Eyes Ultimate Dragon (Fusion Monster)

	//go:embed test_html_ko/test_konami_4861.html
	test_konami_ko_4861 []byte
	// 4861: Solemn Judgment (Counter Trap)

	//go:embed test_html_ko/test_konami_4960.html
	test_konami_ko_4960 []byte
	// 4960: Imperial Order (Continuous Trap)

	//go:embed test_html_ko/test_konami_6341.html
	test_konami_ko_6341 []byte
	// 6341: King of the Skull Servants (Effect Monster)

	//go:embed test_html_ko/test_konami_6996.html
	test_konami_ko_6996 []byte
	// 6996: Advanced Ritual Art (Ritual Spell)

	//go:embed test_html_ko/test_konami_8409.html
	test_konami_ko_8409 []byte
	// 8409: Ally of Justice Cycle Reader (Effect Monster/Tuner)

	//go:embed test_html_ko/test_konami_8933.html
	test_konami_ko_8933 []byte
	// 8933: Effect Veiler (Effect Monster/Tuner)

	//go:embed test_html_ko/test_konami_11232.html
	test_konami_ko_11232 []byte
	// 11232: Shaddoll Falco (Effect Monster/Flip)

	//go:embed test_html_ko/test_konami_11353.html
	test_konami_ko_11353 []byte
	// 11353: Qliphort Scout (Pendulum Normal Monster)

	//go:embed test_html_ko/test_konami_11973.html
	test_konami_ko_11973 []byte
	// 11973: Majespecter Unicorn - Kirin (Pendulum Effect Monster)

	//go:embed test_html_ko/test_konami_12788.html
	test_konami_ko_12788 []byte
	// 12788: Zoodiac Drident (Xyz Monster)

	//go:embed test_html_ko/test_konami_12828.html
	test_konami_ko_12828 []byte
	// 12828: Clear Wing Fast Dragon (Synchro Monster)

	//go:embed test_html_ko/test_konami_12953.html
	test_konami_ko_12953 []byte
	// 12953: Supreme King Z-ARC (Pendulum Fusion Monster)

	//go:embed test_html_ko/test_konami_13587.html
	test_konami_ko_13587 []byte
	// 13587: Ghost Belle & Haunted Mansion (Effect Monster)

	//go:embed test_html_ko/test_konami_13631.html
	test_konami_ko_13631 []byte
	// 13631: Infinite Impermanence (Normal Trap)

	//go:embed test_html_ko/test_konami_14356.html
	test_konami_ko_14356 []byte
	// 14356: Time Thief Redoer (Xyz Monster)

	//go:embed test_html_ko/test_konami_14439.html
	test_konami_ko_14439 []byte
	// 14439: Endymion, the Mighty Master of Magic (Pendulum Monster)

	//go:embed test_html_ko/test_konami_14491.html
	test_konami_ko_14491 []byte
	// 14491: Monk of the Tenyi (Link Monster/non-effect)

	//go:embed test_html_ko/test_konami_14496.html
	test_konami_ko_14496 []byte
	// 14496: Apollousa, Bow of the Goddess (Link Monster, LINK-4)

	//go:embed test_html_ko/test_konami_15296.html
	test_konami_ko_15296 []byte
	// 15296: Triple Tactics Talent (Normal Spell)

	//go:embed test_html_ko/test_konami_15299.html
	test_konami_ko_15299 []byte
	// 15299: Forbidden Droplet (Quick-Play Spell)

	//go:embed test_html_ko/test_konami_15524.html
	test_konami_ko_15524 []byte
	// 15524: Divine Arsenal AA-ZEUS - Sky Thunder (Xyz Monster)

	//go:embed test_html_ko/test_konami_15741.html
	test_konami_ko_15741 []byte
	// 15741: Underworld Goddess of the Closed World (Link Monster, LINK-5)

	//go:embed test_html_ko/test_konami_16386.html
	test_konami_ko_16386 []byte
	// 16386: Baronne de Fleur (Synchro Monster)

	//go:embed test_html_ko/test_konami_16849.html
	test_konami_ko_16849 []byte
	// 16849: D/D/D Deviser King Deus Machinex (Pendulum Xyz Monster)

	//go:embed test_html_ko/test_konami_17474.html
	test_konami_ko_17474 []byte
	// 17474: Tearlaments Sulliek (Continuous Trap)

	//go:embed test_html_ko/test_konami_17746.html
	test_konami_ko_17746 []byte
	// 17746: Garura, Wings of Resonant Life (Fusion Monster)

	//go:embed test_html_ko/test_konami_17808.html
	test_konami_ko_17808 []byte
	// 17808: Branded Regained (Continuous Spell)

	//go:embed test_html_ko/test_konami_18022.html
	test_konami_ko_18022 []byte
	// 18022: Mikanko Water Arabesque (Equip Spell)

	//go:embed test_html_ko/test_konami_18177.html
	test_konami_ko_18177 []byte
	// 18177: Evigishki Neremanas (Ritual Monster)

	//go:embed test_html_ko/test_konami_18792.html
	test_konami_ko_18792 []byte
	// 18792: Cornfield Coatl (Effect Monster)

	//go:embed test_html_ko/test_konami_19188.html
	test_konami_ko_19188 []byte
	// 19188: S:P Little Knight (Link Monster, LINK-2)

	//go:embed test_html_ko/test_konami_19376.html
	test_konami_ko_19376 []byte
	// 19376: Stand Up Centur-Ion! (Field Spell)

	//go:embed test_html_ko/test_konami_19521.html
	test_konami_ko_19521 []byte
	// 19521: Prayers of the Voiceless Voice (Ritual Spell)

	//go:embed test_html_ko/test_konami_20536.html
	test_konami_ko_20536 []byte
	// 20536: Primite Drillbeam (Quick-Play Spell)

	//go:embed test_html_ko/test_konami_20578.html
	test_konami_ko_20578 []byte
	// 20578: Ryzeal Detonator (Xyz Monster, Rank-4)

	//go:embed test_html_ko/test_konami_21627.html
	test_konami_ko_21627 []byte
	// 21627: Miracle Raven (Pendulum Ritual Monster).
	// TCG-exclusive (DUAD-EN084, OP29-EN007); no KO print exists, Konami's
	// KO database returns an empty stub.

	//go:embed test_html_ko/test_konami_22617.html
	test_konami_ko_22617 []byte
	// 22617: Stealth Ange - Cradle (latest Korean release as of 2026-04;
	// KO page lists the card but the EN-name span is still empty)

	//go:embed test_html_ko/test_konami_23141.html
	test_konami_ko_23141 []byte
	// 23141: latest Japanese release as of 2026-04; KO page is still
	// an empty stub (Korean release not yet published)
)

func TestParseKonamiCardHTML_KO_DevPanelSmoke(t *testing.T) {
	// Smoke test: every dev-panel KO fixture must parse to a Card with
	// a non-empty CardName and the expected MiscKonamiCardID. Replace
	// individual entries with full assertion blocks as needed.
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
	}{
		{cardID: "4861", pageHTML: test_konami_ko_4861},
		{cardID: "11353", pageHTML: test_konami_ko_11353},
		{cardID: "11973", pageHTML: test_konami_ko_11973},
		{cardID: "12953", pageHTML: test_konami_ko_12953},
		{cardID: "13587", pageHTML: test_konami_ko_13587},
		{cardID: "13631", pageHTML: test_konami_ko_13631},
		{cardID: "15296", pageHTML: test_konami_ko_15296},
		{cardID: "16386", pageHTML: test_konami_ko_16386},
		{cardID: "16849", pageHTML: test_konami_ko_16849},
		{cardID: "17474", pageHTML: test_konami_ko_17474},
		{cardID: "17746", pageHTML: test_konami_ko_17746},
		{cardID: "17808", pageHTML: test_konami_ko_17808},
		{cardID: "18022", pageHTML: test_konami_ko_18022},
		{cardID: "18177", pageHTML: test_konami_ko_18177},
		{cardID: "19376", pageHTML: test_konami_ko_19376},
		{cardID: "19521", pageHTML: test_konami_ko_19521},
		{cardID: "20536", pageHTML: test_konami_ko_20536},
		{cardID: "20578", pageHTML: test_konami_ko_20578},
		// 21627 (Miracle Raven) is EN-only at crawl time; the KO page is an
		// empty stub. Re-add once Konami publishes the KO card page.
	} {
		// WHEN parsing the KO card page
		got := ParseKonamiCardHTML(c.pageHTML, CardID(c.cardID))

		// THEN the parsed Card carries the expected card ID and a non-empty name
		if string(got.MiscKonamiCardID) != c.cardID {
			t.Errorf("cardID %v MiscKonamiCardID got %q, want %q",
				c.cardID, got.MiscKonamiCardID, c.cardID)
		}
		if got.CardName == "" {
			t.Errorf("cardID %v CardName is empty", c.cardID)
		}
	}
}

func TestParseKonamiCardHTML_KO_LinkArrows(t *testing.T) {
	for _, c := range []struct {
		cardID     string
		pageHTML   []byte
		wantArrows []MonsterLinkArrow
	}{
		// GIVEN a KO page for Apollousa, Bow of the Goddess (LINK-4)
		{cardID: "14496", pageHTML: test_konami_ko_14496,
			wantArrows: []MonsterLinkArrow{DownLeft, Down, DownRight, Up}},

		// GIVEN a KO page for Underworld Goddess of the Closed World (LINK-5)
		{cardID: "15741", pageHTML: test_konami_ko_15741,
			wantArrows: []MonsterLinkArrow{Down, DownRight, Right, Up, UpRight}},

		// GIVEN a KO page for S:P Little Knight (LINK-2, side arrows only)
		{cardID: "19188", pageHTML: test_konami_ko_19188,
			wantArrows: []MonsterLinkArrow{Left, Right}},
	} {
		// WHEN parsing the KO card page
		got := ParseKonamiCardHTML(c.pageHTML, CardID(c.cardID))

		// THEN the link arrows match the same set the EN parser produces
		if !CheckEqualArray(got.MonsterLinkArrows, c.wantArrows) {
			t.Errorf("cardID %v MonsterLinkArrows got %+v, want %+v",
				c.cardID, got.MonsterLinkArrows, c.wantArrows)
		}
	}
}

func TestParseKonamiCardHTML_KO(t *testing.T) {
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
		// GIVEN a KO monster page (Hangul name + EN-name span; no kana ruby on KO)
		{cardID: "4007", pageHTML: test_konami_ko_4007, want: CardLocaleText{
			Name:              "푸른 눈의 백룡",
			NamePronunciation: "",
			NameEnglishOnPage: "Blue-Eyes White Dragon",
			Effect:            "높은 공격력을 자랑하는 전설의 드래곤. 어떠한 상대라도 분쇄해 버리는 파괴력은 상상을 초월한다.",
			AttributeText:     "빛",
			MonsterTypeText:   "드래곤족",
		}},

		// GIVEN a KO spell page
		{cardID: "4343", pageHTML: test_konami_ko_4343, want: CardLocaleText{
			Name:              "번개",
			NamePronunciation: "",
			NameEnglishOnPage: "Raigeki",
			Effect:            "1: 상대 필드의 몬스터를 전부 파괴한다.",
		}},

		// GIVEN a KO trap page
		{cardID: "4960", pageHTML: test_konami_ko_4960, want: CardLocaleText{
			Name:              "왕궁의 칙명",
			NamePronunciation: "",
			NameEnglishOnPage: "Imperial Order",
			Effect:            "이 카드의 컨트롤러는 서로의 스탠바이 페이즈마다 700 LP를 지불한다. 700 LP 지불할 수 없을 경우 이 카드를 파괴한다. 1: 이 카드가 마법 & 함정 존에 존재하는 한, 필드의 모든 마법 카드의 효과는 무효화된다.",
		}},
	} {
		// WHEN parsing locale text from the KO page
		got := ParseCardLocaleText(c.pageHTML, CardID(c.cardID))

		// THEN every locale field matches; KO has no kana so NamePronunciation is empty,
		// and the EN-name span is captured separately in NameEnglishOnPage
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

func TestParseCardLocaleText_KO_NameEnglishOnPage(t *testing.T) {
	// Konami's KO card pages embed the official English name in a bare
	// <span> next to the hangul <h1>. This single source covers all
	// KO-cataloged cards, including OCG-only releases that have no EN page.
	// 21627 (Miracle Raven, TCG-exclusive) has no KO page, so the parser
	// returns an empty NameEnglishOnPage.
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     string
	}{
		{cardID: "4007", pageHTML: test_konami_ko_4007, want: "Blue-Eyes White Dragon"},
		{cardID: "4343", pageHTML: test_konami_ko_4343, want: "Raigeki"},
		{cardID: "4386", pageHTML: test_konami_ko_4386, want: "Blue-Eyes Ultimate Dragon"},
		{cardID: "4861", pageHTML: test_konami_ko_4861, want: "Solemn Judgment"},
		{cardID: "4960", pageHTML: test_konami_ko_4960, want: "Imperial Order"},
		{cardID: "6341", pageHTML: test_konami_ko_6341, want: "King of the Skull Servants"},
		{cardID: "6996", pageHTML: test_konami_ko_6996, want: "Advanced Ritual Art"},
		{cardID: "8409", pageHTML: test_konami_ko_8409, want: "Ally of Justice Cycle Reader"},
		{cardID: "8933", pageHTML: test_konami_ko_8933, want: "Effect Veiler"},
		{cardID: "11232", pageHTML: test_konami_ko_11232, want: "Shaddoll Falco"},
		{cardID: "11353", pageHTML: test_konami_ko_11353, want: "Qliphort Scout"},
		{cardID: "11973", pageHTML: test_konami_ko_11973, want: "Majespecter Unicorn - Kirin"},
		{cardID: "12788", pageHTML: test_konami_ko_12788, want: "Zoodiac Drident"},
		{cardID: "12828", pageHTML: test_konami_ko_12828, want: "Clear Wing Fast Dragon"},
		{cardID: "12953", pageHTML: test_konami_ko_12953, want: "Supreme King Z-ARC"},
		{cardID: "13587", pageHTML: test_konami_ko_13587, want: "Ghost Belle & Haunted Mansion"},
		{cardID: "13631", pageHTML: test_konami_ko_13631, want: "Infinite Impermanence"},
		{cardID: "14356", pageHTML: test_konami_ko_14356, want: "Time Thief Redoer"},
		{cardID: "14439", pageHTML: test_konami_ko_14439, want: "Endymion, the Mighty Master of Magic"},
		{cardID: "14491", pageHTML: test_konami_ko_14491, want: "Monk of the Tenyi"},
		{cardID: "14496", pageHTML: test_konami_ko_14496, want: "Apollousa, Bow of the Goddess"},
		{cardID: "15296", pageHTML: test_konami_ko_15296, want: "Triple Tactics Talent"},
		{cardID: "15299", pageHTML: test_konami_ko_15299, want: "Forbidden Droplet"},
		{cardID: "15524", pageHTML: test_konami_ko_15524, want: "Divine Arsenal AA-ZEUS - Sky Thunder"},
		{cardID: "15741", pageHTML: test_konami_ko_15741, want: "Underworld Goddess of the Closed World"},
		{cardID: "16386", pageHTML: test_konami_ko_16386, want: "Baronne de Fleur"},
		{cardID: "16849", pageHTML: test_konami_ko_16849, want: "D/D/D Deviser King Deus Machinex"},
		{cardID: "17474", pageHTML: test_konami_ko_17474, want: "Tearlaments Sulliek"},
		{cardID: "17746", pageHTML: test_konami_ko_17746, want: "Garura, Wings of Resonant Life"},
		{cardID: "17808", pageHTML: test_konami_ko_17808, want: "Branded Regained"},
		{cardID: "18022", pageHTML: test_konami_ko_18022, want: "Mikanko Water Arabesque"},
		{cardID: "18177", pageHTML: test_konami_ko_18177, want: "Evigishki Neremanas"},
		{cardID: "18792", pageHTML: test_konami_ko_18792, want: "Cornfield Coatl"},
		{cardID: "19188", pageHTML: test_konami_ko_19188, want: "S:P Little Knight"},
		{cardID: "19376", pageHTML: test_konami_ko_19376, want: "Stand Up Centur-Ion!"},
		{cardID: "19521", pageHTML: test_konami_ko_19521, want: "Prayers of the Voiceless Voice"},
		{cardID: "20536", pageHTML: test_konami_ko_20536, want: "Primite Drillbeam"},
		{cardID: "20578", pageHTML: test_konami_ko_20578, want: "Ryzeal Detonator"},
		// 21627 Miracle Raven: TCG-exclusive, KO page is an empty stub
		{cardID: "21627", pageHTML: test_konami_ko_21627, want: ""},
		// 22617: latest Korean release 2026-04, KO page lists the card but
		// Konami has not published the official English name yet
		{cardID: "22617", pageHTML: test_konami_ko_22617, want: ""},
		// 23141: latest Japanese release 2026-04, KO page is an empty stub
		// (Korean release not yet published)
		{cardID: "23141", pageHTML: test_konami_ko_23141, want: ""},
	} {
		// WHEN parsing locale text from the KO page
		got := ParseCardLocaleText(c.pageHTML, CardID(c.cardID))

		// THEN the English-name span on the KO page matches the expected name
		if got.NameEnglishOnPage != c.want {
			t.Errorf("cardID %v NameEnglishOnPage got %q, want %q",
				c.cardID, got.NameEnglishOnPage, c.want)
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
