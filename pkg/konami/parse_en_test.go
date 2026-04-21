package konami

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareData(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error Getwd: %v", err)
	}
	outputDir := filepath.Join(pwd, "test_html")
	t.Logf("outputDir TestPrepareData: %v", outputDir)

	t.Skip("only need to run once at the start of coding to prepare data")

	for _, locale := range []string{"en", "ja", "ko"} {
		for _, cardID := range []string{
			"4007",  // Blue-Eyes White Dragon, test Normal Monster
			"4343",  // Raigeki, test Normal Spell
			"4386",  // Blue-Eyes Ultimate Dragon, test Fusion Monster
			"4960",  // Imperial Order, test Continuous Trap
			"6341",  // King of the Skull Servants, test ATK ? DEF 0
			"6996",  // Advanced Ritual Art, test Ritual Spell
			"8409",  // Ally of Justice Cycle Reader, test Effect Monster, ability Tuner, card password with zero prefix "08233522"
			"8933",  // Effect Veiler, test Effect Monster with ability Tuner and ATK and DEF as 0 (not "?")
			"11232", // Shaddoll Falco, test multiple abilities Flip and Tuner
			"12788", // Zoodiac Drident, test ATK and DEF as "?"
			"12828", // Clear Wing Fast Dragon, test Pendulum and Synchro
			"14356", // Time Thief Redoer, test Xyz Monster with effect has bullet points
			"14439", // Endymion, the Mighty Master of Magic, test Pendulum Monster with long effect
			"14491", // Monk of the Tenyi, test Link Monster that has no effect
			"14496", // Apollousa, Bow of the Goddess, test ATK as "?"
			"15741", // Underworld Goddess of the Closed World, test many link arrows
			"15299", // Forbidden Droplet, test Quick-Play Spell
			"15524", // Divine Arsenal AA-ZEUS - Sky Thunder, test high rank Xyz Monster
			"18792", // Cornfield Coatl, test new Illusion monster type
		} {
			cardURL := `https://www.db.yugioh-card.com/yugiohdb/card_search.action?ope=2` +
				fmt.Sprintf(`&request_locale=%v&cid=%v`, locale, cardID)
			t.Logf("http.Get: %v", cardURL)
			w, err := http.Get(cardURL)
			if err != nil {
				t.Fatalf("error cardID %v http.Get: %v", cardID, err)
			}
			bodyBs, err := io.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("error cardID %v io.ReadAll: %v", cardID, err)
			}
			outputDirLang := outputDir + "_" + locale
			outputFile := filepath.Join(outputDirLang,
				fmt.Sprintf("test_konami_%v.html", cardID))
			f, err := os.Create(outputFile)
			if err != nil {
				t.Logf("error cardID %v os.OpenFile: %v", cardID, err)
				continue
			}
			_, err = f.Write(bodyBs)
			if err != nil {
				t.Logf("error cardID %v f.Write: %v", cardID, err)
				continue
			}
			err = f.Sync()
			if err != nil {
				t.Logf("error cardID %v f.Sync: %v", cardID, err)
			}
			t.Logf("written %v", outputFile)
			err = f.Close()
			if err != nil {
				t.Logf("error cardID %v f.Close: %v", cardID, err)
			}
		}
	}
}

var (
	//go:embed test_html_en/test_konami_4007.html
	test_konami_4007 []byte
	// 4007: Blue-Eyes White Dragon (Normal Monster)

	//go:embed test_html_en/test_konami_4343.html
	test_konami_4343 []byte
	// 4343: Raigeki (Normal Spell)

	//go:embed test_html_en/test_konami_4386.html
	test_konami_4386 []byte
	// 4386: Blue-Eyes Ultimate Dragon (Fusion)

	//go:embed test_html_en/test_konami_4960.html
	test_konami_4960 []byte
	// 4960: Imperial Order (Continuous Trap)

	//go:embed test_html_en/test_konami_6341.html
	test_konami_6341 []byte
	// 6341: King of the Skull Servants (Effect Monster)

	//go:embed test_html_en/test_konami_6996.html
	test_konami_6996 []byte
	// 6996: Advanced Ritual Art (Ritual Spell)

	//go:embed test_html_en/test_konami_8409.html
	test_konami_8409 []byte
	// 8409: Ally of Justice Cycle Reader (Effect Monster/Tuner)

	//go:embed test_html_en/test_konami_8933.html
	test_konami_8933 []byte
	// 8933: Effect Veiler (Effect Monster/Tuner)

	//go:embed test_html_en/test_konami_11232.html
	test_konami_11232 []byte
	// 11232: Shaddoll Falco (Effect Monster/Flip)

	//go:embed test_html_en/test_konami_12788.html
	test_konami_12788 []byte
	// 12788: Zoodiac Drident (Xyz Monster)

	//go:embed test_html_en/test_konami_12828.html
	test_konami_12828 []byte
	// 12828: Clear Wing Fast Dragon (Synchro Monster)

	//go:embed test_html_en/test_konami_14356.html
	test_konami_14356 []byte
	// 14356: Time Thief Redoer (Xyz Monster)

	//go:embed test_html_en/test_konami_14439.html
	test_konami_14439 []byte
	// 14439: Endymion, the Mighty Master of Magic (Pendulum Monster)

	//go:embed test_html_en/test_konami_14491.html
	test_konami_14491 []byte
	// 14491: Monk of the Tenyi (Link Monster/non-effect)

	//go:embed test_html_en/test_konami_14496.html
	test_konami_14496 []byte
	// 14496: Apollousa, Bow of the Goddess (Link Monster)

	//go:embed test_html_en/test_konami_15299.html
	test_konami_15299 []byte
	// 15299: Forbidden Droplet (Quick-Play Spell)

	//go:embed test_html_en/test_konami_15524.html
	test_konami_15524 []byte
	// 15524: Divine Arsenal AA-ZEUS - Sky Thunder (Xyz Monster)

	//go:embed test_html_en/test_konami_15741.html
	test_konami_15741 []byte
	// 15741: Underworld Goddess of the Closed World (Link Monster)

	//go:embed test_html_en/test_konami_18792.html
	test_konami_18792 []byte
	// 18792: Cornfield Coatl (Effect Monster)
)

func TestParseKonamiCardHTML(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     Card
	}{
		{pageHTML: test_konami_4007, cardID: "4007",
			want: Card{
				CardName:    "Blue-Eyes White Dragon",
				CardType:    Monster,
				CardSubtype: MonsterNormal,
				CardEffect:  `This legendary dragon is a powerful engine of destruction. Virtually invincible, very few have faced this awesome creature and lived to tell the tale.`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Dragon,
				MonsterLevelRankLink: 8,
				MonsterATK:           3000,
				MonsterDEF:           2500,
				IsNonEffectMonster:   true,

				MiscKonamiSet:    "LOB-001",
				MiscKonamiCardID: "4007",
				MiscYear:         "2002",
			}},

		{pageHTML: test_konami_4343, cardID: "4343",
			want: Card{
				CardName:         "Raigeki",
				CardType:         Spell,
				CardSubtype:      SpellNormal,
				CardEffect:       `Destroy all monsters your opponent controls.`,
				MiscKonamiSet:    "LOB-053",
				MiscKonamiCardID: "4343",
				MiscYear:         "2002",
			}},

		{pageHTML: test_konami_4386, cardID: "4386",
			want: Card{
				CardName:    "Blue-Eyes Ultimate Dragon",
				CardType:    Monster,
				CardSubtype: MonsterFusion,
				CardEffect:  `"Blue-Eyes White Dragon" + "Blue-Eyes White Dragon" + "Blue-Eyes White Dragon"`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Dragon,
				MonsterLevelRankLink: 12,
				MonsterATK:           4500,
				MonsterDEF:           3800,
				IsNonEffectMonster:   true,

				MiscKonamiSet:    "GLD1-EN028",
				MiscKonamiCardID: "4386",
				MiscYear:         "2008",
			}},

		{pageHTML: test_konami_4960, cardID: "4960",
			want: Card{
				CardName:    "Imperial Order",
				CardType:    Trap,
				CardSubtype: TrapContinuous,
				CardEffect:  `Negate all Spell effects on the field. Once per turn, during the Standby Phase, you must pay 700 LP (this is not optional), or this card is destroyed.`,

				MiscKonamiSet:    "PSV-104",
				MiscKonamiCardID: "4960",
				MiscYear:         "2002",
			}},

		{pageHTML: test_konami_6341, cardID: "6341",
			want: Card{
				CardName:             "King of the Skull Servants",
				CardType:             Monster,
				CardSubtype:          MonsterEffect,
				CardEffect:           `The original ATK of this card is the combined number of "King of the Skull Servants" and "Skull Servant" in your GY x 1000. When this card is destroyed by battle and sent to the GY: You can banish 1 other "King of the Skull Servants" or 1 "Skull Servant" from your GY; Special Summon this card.`,
				MonsterAttribute:     DARK,
				MonsterType:          Zombie,
				MonsterLevelRankLink: 1,
				MonsterATKStr:        "?",
				MonsterDEFStr:        "0",
				MiscKonamiSet:        "TLM-EN032",
				MiscKonamiCardID:     "6341",
				MiscCardPassword:     "36021814",
				MiscYear:             "2005",
			}},

		{pageHTML: test_konami_6996, cardID: "6996",
			want: Card{
				CardName:    "Advanced Ritual Art",
				CardType:    Spell,
				CardSubtype: SpellRitual,
				CardEffect:  `This card is used to Ritual Summon any 1 Ritual Monster. You must also send Normal Monsters from your Deck to the GY whose total Levels equal the Level of that Ritual Monster.`,

				MiscKonamiSet:    "STON-EN045",
				MiscKonamiCardID: "6996",
				MiscYear:         "2007",
			}},

		{pageHTML: test_konami_8409, cardID: "8409",
			want: Card{
				CardName:             "Ally of Justice Cycle Reader",
				CardType:             Monster,
				CardSubtype:          MonsterEffect,
				CardEffect:           "(Quick Effect): You can discard this card to the GY, then target up to 2 LIGHT monsters in your opponent's GY; banish those targets.",
				CardArt:              "",
				MonsterAttribute:     DARK,
				MonsterType:          Machine,
				MonsterLevelRankLink: 3,
				MonsterATK:           1000,
				MonsterATKStr:        "1000",
				MonsterDEF:           1000,
				MonsterDEFStr:        "1000",
				MonsterAbilities:     []MonsterAbility{Tuner},
				MonsterLinkArrows:    []MonsterLinkArrow{},
				IsPendulum:           false,
				PendulumScale:        0,
				PendulumEffect:       "",
				MiscKonamiSet:        "DT03-EN080",
				MiscKonamiCardID:     "8409",
				MiscCardPassword:     "08233522",
				MiscYear:             "2010",
				MiscCreator:          "daominah.github.io",
			}},

		{pageHTML: test_konami_8933, cardID: "8933",
			want: Card{
				CardName:             "Effect Veiler",
				CardType:             Monster,
				CardSubtype:          MonsterEffect,
				CardEffect:           `During your opponent's Main Phase (Quick Effect): You can send this card from your hand to the GY, then target 1 Effect Monster your opponent controls; negate the effects of that face-up monster your opponent controls, until the end of this turn.`,
				MonsterAttribute:     LIGHT,
				MonsterType:          Spellcaster,
				MonsterLevelRankLink: 1,
				MonsterATK:           0,
				MonsterDEF:           0,
				MonsterAbilities:     []MonsterAbility{Tuner},
				MiscKonamiSet:        "DREV-EN002",
				MiscKonamiCardID:     "8933",
				MiscYear:             "2010",
			}},

		{pageHTML: test_konami_11232, cardID: "11232",
			want: Card{
				CardName:    "Shaddoll Falco",
				CardType:    Monster,
				CardSubtype: MonsterEffect,
				CardEffect: `FLIP: You can target 1 "Shaddoll" monster in your GY, except "Shaddoll Falco"; Special Summon it in face-down Defense Position.
If this card is sent to the GY by a card effect: You can Special Summon it in face-down Defense Position. You can only use 1 "Shaddoll Falco" effect per turn, and only once that turn.`,

				MonsterAttribute:     DARK,
				MonsterType:          Spellcaster,
				MonsterLevelRankLink: 2,
				MonsterATK:           600,
				MonsterDEF:           1400,
				MonsterAbilities:     []MonsterAbility{Flip, Tuner},

				MiscKonamiSet:    "DUEA-EN023",
				MiscKonamiCardID: "11232",
				MiscYear:         "2014",
			}},

		{pageHTML: test_konami_12788, cardID: "12788",
			want: Card{
				CardName:    "Zoodiac Drident",
				CardType:    Monster,
				CardSubtype: MonsterXyz,
				CardEffect: `4 Level 4 monsters
Once per turn, you can also Xyz Summon "Zoodiac Drident" by using 1 "Zoodiac" monster you control with a different name as material. (Transfer its materials to this card.) Gains ATK/DEF equal to the ATK/DEF of all "Zoodiac" monsters attached to it as material. Once per turn (Quick Effect): You can detach 1 material from this card, then target 1 face-up card on the field; destroy it.`,
				MonsterAttribute:     EARTH,
				MonsterType:          BeastWarrior,
				MonsterLevelRankLink: 4,
				MonsterATK:           0,
				MonsterATKStr:        "?",
				MonsterDEF:           0,
				MonsterDEFStr:        "?",
				MiscKonamiSet:        "RATE-EN053",
				MiscKonamiCardID:     "12788",
				MiscYear:             "2017",
			}},

		{pageHTML: test_konami_12828, cardID: "12828",
			want: Card{
				CardName:    "Clear Wing Fast Dragon",
				CardType:    Monster,
				CardSubtype: MonsterSynchro,
				CardEffect: `1 Tuner + 1+ non-Tuner WIND monsters
(Quick Effect): You can target 1 face-up monster your opponent controls that was Special Summoned from the Extra Deck; until the end of this turn, change its ATK to 0, also negate that face-up monster's effects. You can only use this effect of "Clear Wing Fast Dragon" once per turn. If this card in the Monster Zone is destroyed by battle or card effect: You can place this card in your Pendulum Zone.`,

				MonsterAttribute:     WIND,
				MonsterType:          Dragon,
				MonsterLevelRankLink: 7,
				MonsterATK:           2500,
				MonsterDEF:           2000,

				MiscKonamiSet:    "YA02-EN001",
				MiscKonamiCardID: "12828",
				MiscYear:         "2017",

				IsPendulum:     true,
				PendulumScale:  4,
				PendulumEffect: `You can send 1 face-up "Speedroid" Tuner and 1 face-up non-Tuner monster you control to the GY, whose total Levels equal 7; Special Summon this card from your Pendulum Zone. You can only use this effect of "Clear Wing Fast Dragon" once per turn.`,
			}},

		{pageHTML: test_konami_14356, cardID: "14356",
			want: Card{
				CardName:    "Time Thief Redoer",
				CardType:    Monster,
				CardSubtype: MonsterXyz,
				CardEffect: `2 Level 4 monsters
Once per turn, during the Standby Phase: You can attach the top card of your opponent's Deck to this card as material. (Quick Effect): You can detach up to 3 different types of materials from this card, then apply the following effect(s) depending on what was detached.
● Monster: Banish this card until the End Phase.
● Spell: Draw 1 card.
● Trap: Place 1 face-up card your opponent controls on top of the Deck.
You can only use this effect of "Time Thief Redoer" once per turn.`,

				MonsterAttribute:     DARK,
				MonsterType:          Psychic,
				MonsterLevelRankLink: 4,
				MonsterATK:           2400,
				MonsterDEF:           2000,

				MiscKonamiSet:    "SAST-EN085",
				MiscKonamiCardID: "14356",
				MiscYear:         "2019",
			}},

		{pageHTML: test_konami_14439, cardID: "14439",
			want: Card{
				CardName:    "Endymion, the Mighty Master of Magic",
				CardType:    Monster,
				CardSubtype: MonsterEffect,
				CardEffect:  "Once per turn, when a Spell/Trap Card or effect is activated (Quick Effect): You can return 1 card you control with a Spell Counter to the hand, and if you do, negate the activation, and if you do that, destroy it. Then, you can place the same number of Spell Counters on this card that the returned card had. While this card has a Spell Counter, your opponent cannot target it with card effects, also it cannot be destroyed by your opponent's card effects. When this card with a Spell Counter is destroyed by battle: You can add 1 Normal Spell from your Deck to your hand.",

				MonsterAttribute:     DARK,
				MonsterType:          Spellcaster,
				MonsterLevelRankLink: 7,
				MonsterATK:           2800,
				MonsterDEF:           1700,

				MiscKonamiSet:    "SR08-EN001",
				MiscKonamiCardID: "14439",
				MiscYear:         "2019",

				IsPendulum:     true,
				PendulumScale:  8,
				PendulumEffect: `You can remove 6 Spell Counters from your field; Special Summon this card from the Pendulum Zone, then count the number of cards you control that can have a Spell Counter, destroy up to that many cards on the field, and if you do, place Spell Counters on this card equal to the number of cards destroyed. You can only use this effect of "Endymion, the Mighty Master of Magic" once per turn.`,
			}},

		{pageHTML: test_konami_14491, cardID: "14491",
			want: Card{
				CardName:    "Monk of the Tenyi",
				CardType:    Monster,
				CardSubtype: MonsterLink,
				CardEffect:  `1 non-Link "Tenyi" monster`,

				MonsterAttribute:     EARTH,
				MonsterType:          Wyrm,
				MonsterLevelRankLink: 1,
				MonsterATK:           1000,
				MonsterDEF:           0,
				MonsterLinkArrows:    []MonsterLinkArrow{Down},
				IsNonEffectMonster:   true,

				MiscKonamiSet:    "RIRA-EN043",
				MiscKonamiCardID: "14491",
				MiscYear:         "2019",
			}},

		{pageHTML: test_konami_14496, cardID: "14496",
			want: Card{
				CardName:    "Apollousa, Bow of the Goddess",
				CardType:    Monster,
				CardSubtype: MonsterLink,
				CardEffect: `2+ monsters with different names, except Tokens
You can only control 1 "Apollousa, Bow of the Goddess". The original ATK of this card becomes 800 x the number of Link Materials used for its Link Summon. Once per Chain, when your opponent activates a monster effect (Quick Effect): You can make this card lose exactly 800 ATK, and if you do, negate the activation.`,

				MonsterAttribute:     WIND,
				MonsterType:          Fairy,
				MonsterLevelRankLink: 4,
				MonsterATK:           0,
				MonsterATKStr:        "?",
				MonsterDEF:           0,
				MonsterLinkArrows:    []MonsterLinkArrow{DownLeft, Down, DownRight, Up},

				MiscKonamiSet:    "RIRA-EN048",
				MiscKonamiCardID: "14496",
				MiscYear:         "2019",
			}},

		{pageHTML: test_konami_15299, cardID: "15299",
			want: Card{
				CardName:    "Forbidden Droplet",
				CardType:    Spell,
				CardSubtype: SpellQuickPlay,
				CardEffect:  `Send any number of other cards from your hand and/or field to the GY; choose that many Effect Monsters your opponent controls, and until the end of this turn, their ATK is halved, also their effects are negated. In response to this card's activation, your opponent cannot activate cards, or the effects of cards, with the same original type (Monster/Spell/Trap) as the cards sent to the GY to activate this card. You can only activate 1 "Forbidden Droplet" per turn.`,

				MiscKonamiSet:    "ROTD-EN065",
				MiscKonamiCardID: "15299",
				MiscYear:         "2020",
			}},

		{pageHTML: test_konami_15524, cardID: "15524",
			want: Card{
				CardName:    "Divine Arsenal AA-ZEUS - Sky Thunder",
				CardType:    Monster,
				CardSubtype: MonsterXyz,
				CardEffect: `2 Level 12 monsters
Once per turn, if an Xyz Monster battled this turn, you can also Xyz Summon "Divine Arsenal AA-ZEUS - Sky Thunder" by using 1 Xyz Monster you control as material (transfer its materials to this card). (Quick Effect): You can detach 2 materials from this card; send all other cards from the field to the GY. Once per turn, if another card(s) you control is destroyed by battle or an opponent's card effect: You can attach 1 card from your hand, Deck, or Extra Deck to this card.`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Machine,
				MonsterLevelRankLink: 12,
				MonsterATK:           3000,
				MonsterDEF:           3000,

				MiscKonamiSet:    "PHRA-EN045",
				MiscKonamiCardID: "15524",
				MiscYear:         "2020",
			}},

		{pageHTML: test_konami_15741, cardID: "15741",
			want: Card{
				CardName:    "Underworld Goddess of the Closed World",
				CardType:    Monster,
				CardSubtype: MonsterLink,
				CardEffect: `4+ Effect Monsters
You can also use 1 monster your opponent controls as material to Link Summon this card. If this card is Link Summoned: You can negate the effects of all face-up monsters your opponent currently controls. This Link Summoned card is unaffected by your opponent's activated effects, unless they target this card. Once per turn, when your opponent activates a card or effect that Special Summons a monster(s) from the GY (Quick Effect): You can negate the activation.`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Fiend,
				MonsterLevelRankLink: 5,
				MonsterATK:           3000,
				MonsterDEF:           0,
				MonsterLinkArrows:    []MonsterLinkArrow{Down, DownRight, Right, Up, UpRight},

				MiscKonamiSet:    "BLVO-EN050",
				MiscKonamiCardID: "15741",
				MiscYear:         "2021",
			}},

		{pageHTML: test_konami_18792, cardID: "18792",
			want: Card{
				CardName:    "Cornfield Coatl",
				CardType:    Monster,
				CardSubtype: MonsterEffect,
				CardEffect:  `If this card battles a monster, neither can be destroyed by that battle. You can only use each of the following effects of "Cornfield Coatl" once per turn. You can discard this card; add 1 monster that mentions "Chimera Fusion" from your Deck to your hand, except "Cornfield Coatl". When your opponent activates a card or effect that targets a card(s) you control, while you control "Chimera the Flying Mythical Beast" (Quick Effect): You can banish this card from your field or GY; negate that effect, and if you do, destroy that card.`,

				MonsterAttribute:     WIND,
				MonsterType:          Illusion,
				MonsterLevelRankLink: 4,
				MonsterATK:           500,
				MonsterDEF:           1700,

				MiscKonamiSet:    "DUNE-EN005",
				MiscKonamiCardID: "18792",
				MiscYear:         "2023",
			}},
	} {
		got := ParseKonamiCardHTML(c.pageHTML, CardID(c.cardID))
		if got.CardName != c.want.CardName {
			t.Errorf(`error cardID %v CardName got "%v", want "%v"`, c.cardID, got.CardName, c.want.CardName)
		}
		if got.CardType != c.want.CardType {
			t.Errorf(`error cardID %v CardType got "%v", want "%v"`, c.cardID, got.CardType, c.want.CardType)
		}
		if got.CardSubtype != c.want.CardSubtype {
			t.Errorf(`error cardID %v CardSubtype got "%v", want "%v"`, c.cardID, got.CardSubtype, c.want.CardSubtype)
		}
		if got.CardEffect != c.want.CardEffect {
			t.Errorf("error cardID %v CardEffect got:\n%v\n, want:\n%v\n", c.cardID, got.CardEffect, c.want.CardEffect)
		}
		if got.MonsterAttribute != c.want.MonsterAttribute {
			t.Errorf(`error cardID %v MonsterAttribute got "%v", want "%v"`, c.cardID, got.MonsterAttribute, c.want.MonsterAttribute)
		}
		if got.MonsterType != c.want.MonsterType {
			t.Errorf(`error cardID %v MonsterType got "%v", want "%v"`, c.cardID, got.MonsterType, c.want.MonsterType)
		}
		if got.MonsterLevelRankLink != c.want.MonsterLevelRankLink {
			t.Errorf(`error cardID %v MonsterLevelRankLink got "%v", want "%v"`, c.cardID, got.MonsterLevelRankLink, c.want.MonsterLevelRankLink)
		}
		if got.MonsterATK != c.want.MonsterATK {
			t.Errorf(`error cardID %v MonsterATK got "%v", want "%v"`, c.cardID, got.MonsterATK, c.want.MonsterATK)
		}
		if got.MonsterDEF != c.want.MonsterDEF {
			t.Errorf(`error cardID %v MonsterDEF got "%v", want "%v"`, c.cardID, got.MonsterDEF, c.want.MonsterDEF)
		}
		if !CheckEqualArray(got.MonsterAbilities, c.want.MonsterAbilities) {
			t.Errorf(`error cardID %v MonsterAbilities got "%+v", want "%+v"`, c.cardID, got.MonsterAbilities, c.want.MonsterAbilities)
		}
		if !CheckEqualArray(got.MonsterLinkArrows, c.want.MonsterLinkArrows) {
			t.Errorf(`error cardID %v MonsterLinkArrows got "%+v", want "%+v"`, c.cardID, got.MonsterLinkArrows, c.want.MonsterLinkArrows)
		}
		if got.IsNonEffectMonster != c.want.IsNonEffectMonster {
			t.Errorf(`error cardID %v IsNonEffectMonster got "%v", want "%v"`, c.cardID, got.IsNonEffectMonster, c.want.IsNonEffectMonster)
		}

		if got.IsPendulum != c.want.IsPendulum {
			t.Errorf(`error cardID %v IsPendulum got "%v", want "%v"`, c.cardID, got.IsPendulum, c.want.IsPendulum)
		}
		if got.PendulumScale != c.want.PendulumScale {
			t.Errorf(`error cardID %v PendulumScale got "%v", want "%v"`, c.cardID, got.PendulumScale, c.want.PendulumScale)
		}
		if got.PendulumEffect != c.want.PendulumEffect {
			t.Errorf(`error cardID %v PendulumEffect got "%v", want "%v"`, c.cardID, got.PendulumEffect, c.want.PendulumEffect)
		}

		if got.MiscKonamiSet != c.want.MiscKonamiSet {
			t.Errorf(`error cardID %v MiscKonamiSet got "%v", want "%v"`, c.cardID, got.MiscKonamiSet, c.want.MiscKonamiSet)
		}
		if got.MiscKonamiCardID != c.want.MiscKonamiCardID {
			t.Errorf(`error cardID %v MiscKonamiCardID got "%v", want "%v"`, c.cardID, got.MiscKonamiCardID, c.want.MiscKonamiCardID)
		}
		if got.MiscYear != c.want.MiscYear {
			t.Errorf(`error cardID %v MiscYear got "%v", want "%v"`, c.cardID, got.MiscYear, c.want.MiscYear)
		}

		if got.CardName == "Apollousa, Bow of the Goddess" &&
			got.MonsterATKStr != c.want.MonsterATKStr {
			t.Errorf(`error cardID %v MonsterATKStr got "%v", want "%v"`, c.cardID, got.MonsterATKStr, c.want.MonsterATKStr)
		}

		if got.CardName == "Effect Veiler" &&
			got.MonsterATKStr == "?" {
			t.Errorf(`error cardID %v MonsterATKStr got "%v", want "0" or empty`, c.cardID, got.MonsterATKStr)
		}
	}
}
