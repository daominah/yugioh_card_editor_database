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

	//t.Skip("only need to run once at the start of coding to prepare data")

	// isSkipExisting avoids re-downloading fixtures that are already on disk,
	// so re-runs only fetch the newly-added cardIDs from Konami's server.
	const isSkipExisting = true

	for _, locale := range []string{"en", "ja", "ko"} {
		for _, cardID := range []string{
			"4007",  // Blue-Eyes White Dragon, test Normal Monster
			"4343",  // Raigeki, test Normal Spell
			"4386",  // Blue-Eyes Ultimate Dragon, test Fusion Monster
			"4861",  // Solemn Judgment, test Counter Trap
			"4960",  // Imperial Order, test Continuous Trap
			"6341",  // King of the Skull Servants, test ATK ? DEF 0
			"6996",  // Advanced Ritual Art, test Ritual Spell
			"8409",  // Ally of Justice Cycle Reader, test Effect Monster, ability Tuner, card password with zero prefix "08233522"
			"8933",  // Effect Veiler, test Effect Monster with ability Tuner and ATK and DEF as 0 (not "?")
			"11232", // Shaddoll Falco, test multiple abilities Flip and Tuner
			"11353", // Qliphort Scout, test Pendulum Normal Monster
			"11973", // Majespecter Unicorn - Kirin, test Pendulum Effect Monster
			"12788", // Zoodiac Drident, test ATK and DEF as "?"
			"12828", // Clear Wing Fast Dragon, test Pendulum and Synchro
			"12953", // Supreme King Z-ARC, test Pendulum Fusion Monster
			"13587", // Ghost Belle & Haunted Mansion, test Effect Monster
			"13631", // Infinite Impermanence, test Normal Trap
			"14356", // Time Thief Redoer, test Xyz Monster with effect has bullet points
			"14439", // Endymion, the Mighty Master of Magic, test Pendulum Monster with long effect
			"14491", // Monk of the Tenyi, test Link Monster that has no effect
			"14496", // Apollousa, Bow of the Goddess, test ATK as "?"
			"15296", // Triple Tactics Talent, test Normal Spell
			"15299", // Forbidden Droplet, test Quick-Play Spell
			"15524", // Divine Arsenal AA-ZEUS - Sky Thunder, test high rank Xyz Monster
			"15741", // Underworld Goddess of the Closed World, test many link arrows
			"16386", // Baronne de Fleur, test Synchro Monster
			"16849", // D/D/D Deviser King Deus Machinex, test Pendulum Xyz Monster
			"17474", // Tearlaments Sulliek, test Continuous Trap
			"17746", // Garura, Wings of Resonant Life, test Fusion Monster
			"17808", // Branded Regained, test Continuous Spell
			"18022", // Mikanko Water Arabesque, test Equip Spell
			"18177", // Evigishki Neremanas, test Ritual Monster
			"18792", // Cornfield Coatl, test new Illusion monster type
			"19188", // S:P Little Knight, test Link Monster LINK-2
			"19375", // Centur-Ion Legatia, test Synchro Monster generic Lv12
			"19376", // Stand Up Centur-Ion!, test Field Spell
			"19521", // Prayers of the Voiceless Voice, test Ritual Spell
			"20536", // Primite Drillbeam, test Quick-Play Spell
			"20578", // Ryzeal Detonator, test Xyz Monster Rank-4
			"21627", // Miracle Raven, test TCG-exclusive Pendulum Ritual Monster
			"22617", // latest Korean card in 2026-04
			"23141", // latest Japanese card in 2026-04
		} {
			outputDirLang := outputDir + "_" + locale
			outputFile := filepath.Join(outputDirLang,
				fmt.Sprintf("test_konami_%v.html", cardID))
			if isSkipExisting {
				if _, err := os.Stat(outputFile); err == nil {
					//t.Logf("skip existing %v", outputFile)
					continue
				}
			}
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

	//go:embed test_html_en/test_konami_4861.html
	test_konami_4861 []byte
	// 4861: Solemn Judgment (Counter Trap)

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

	//go:embed test_html_en/test_konami_11353.html
	test_konami_11353 []byte
	// 11353: Qliphort Scout (Pendulum Normal Monster)

	//go:embed test_html_en/test_konami_11973.html
	test_konami_11973 []byte
	// 11973: Majespecter Unicorn - Kirin (Pendulum Effect Monster)

	//go:embed test_html_en/test_konami_12788.html
	test_konami_12788 []byte
	// 12788: Zoodiac Drident (Xyz Monster)

	//go:embed test_html_en/test_konami_12828.html
	test_konami_12828 []byte
	// 12828: Clear Wing Fast Dragon (Synchro Monster)

	//go:embed test_html_en/test_konami_12953.html
	test_konami_12953 []byte
	// 12953: Supreme King Z-ARC (Pendulum Fusion Monster)

	//go:embed test_html_en/test_konami_13587.html
	test_konami_13587 []byte
	// 13587: Ghost Belle & Haunted Mansion (Effect Monster)

	//go:embed test_html_en/test_konami_13631.html
	test_konami_13631 []byte
	// 13631: Infinite Impermanence (Normal Trap)

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

	//go:embed test_html_en/test_konami_15296.html
	test_konami_15296 []byte
	// 15296: Triple Tactics Talent (Normal Spell)

	//go:embed test_html_en/test_konami_15299.html
	test_konami_15299 []byte
	// 15299: Forbidden Droplet (Quick-Play Spell)

	//go:embed test_html_en/test_konami_15524.html
	test_konami_15524 []byte
	// 15524: Divine Arsenal AA-ZEUS - Sky Thunder (Xyz Monster)

	//go:embed test_html_en/test_konami_15741.html
	test_konami_15741 []byte
	// 15741: Underworld Goddess of the Closed World (Link Monster)

	//go:embed test_html_en/test_konami_16386.html
	test_konami_16386 []byte
	// 16386: Baronne de Fleur (Synchro Monster)

	//go:embed test_html_en/test_konami_16849.html
	test_konami_16849 []byte
	// 16849: D/D/D Deviser King Deus Machinex (Pendulum Xyz Monster)

	//go:embed test_html_en/test_konami_17474.html
	test_konami_17474 []byte
	// 17474: Tearlaments Sulliek (Continuous Trap)

	//go:embed test_html_en/test_konami_17746.html
	test_konami_17746 []byte
	// 17746: Garura, Wings of Resonant Life (Fusion Monster)

	//go:embed test_html_en/test_konami_17808.html
	test_konami_17808 []byte
	// 17808: Branded Regained (Continuous Spell)

	//go:embed test_html_en/test_konami_18022.html
	test_konami_18022 []byte
	// 18022: Mikanko Water Arabesque (Equip Spell)

	//go:embed test_html_en/test_konami_18177.html
	test_konami_18177 []byte
	// 18177: Evigishki Neremanas (Ritual Monster)

	//go:embed test_html_en/test_konami_18792.html
	test_konami_18792 []byte
	// 18792: Cornfield Coatl (Effect Monster)

	//go:embed test_html_en/test_konami_19188.html
	test_konami_19188 []byte
	// 19188: S:P Little Knight (Link Monster, LINK-2)

	//go:embed test_html_en/test_konami_19375.html
	test_konami_19375 []byte
	// 19375: Centur-Ion Legatia (Synchro Monster, Lv-12)

	//go:embed test_html_en/test_konami_19376.html
	test_konami_19376 []byte
	// 19376: Stand Up Centur-Ion! (Field Spell)

	//go:embed test_html_en/test_konami_19521.html
	test_konami_19521 []byte
	// 19521: Prayers of the Voiceless Voice (Ritual Spell)

	//go:embed test_html_en/test_konami_20536.html
	test_konami_20536 []byte
	// 20536: Primite Drillbeam (Quick-Play Spell)

	//go:embed test_html_en/test_konami_20578.html
	test_konami_20578 []byte
	// 20578: Ryzeal Detonator (Xyz Monster, Rank-4)

	//go:embed test_html_en/test_konami_21627.html
	test_konami_21627 []byte
	// 21627: Miracle Raven (Pendulum Ritual Monster, TCG-exclusive)
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

		{pageHTML: test_konami_4861, cardID: "4861",
			want: Card{
				CardName:    "Solemn Judgment",
				CardType:    Trap,
				CardSubtype: TrapCounter,
				CardEffect:  `When a monster(s) would be Summoned, OR a Spell/Trap Card is activated: Pay half your LP; negate the Summon or activation, and if you do, destroy that card.`,

				MiscKonamiSet:    "MRD-127",
				MiscKonamiCardID: "4861",
				MiscYear:         "2002",
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

		{pageHTML: test_konami_11353, cardID: "11353",
			want: Card{
				CardName:    "Qliphort Scout",
				CardType:    Monster,
				CardSubtype: MonsterNormal,
				CardEffect: `Booting in Replica Mode...
An error has occurred while executing C:\sophia\zefra.exe
Unknown publisher.
Allow C:\tierra\qliphort.exe ?
...[Y]
Booting in Autonomy Mode...`,

				MonsterAttribute:     EARTH,
				MonsterType:          Machine,
				MonsterLevelRankLink: 5,
				MonsterATK:           1000,
				MonsterDEF:           2800,
				IsNonEffectMonster:   true,

				MiscKonamiSet:    "NECH-EN021",
				MiscKonamiCardID: "11353",
				MiscYear:         "2014",

				IsPendulum:     true,
				PendulumScale:  9,
				PendulumEffect: `You cannot Special Summon monsters, except "Qli" monsters. This effect cannot be negated. Once per turn: You can pay 800 LP; add 1 "Qli" card from your Deck to your hand, except "Qliphort Scout".`,
			}},

		{pageHTML: test_konami_11973, cardID: "11973",
			want: Card{
				CardName:    "Majespecter Unicorn - Kirin",
				CardType:    Monster,
				CardSubtype: MonsterEffect,
				CardEffect:  `(Quick Effect): You can target 1 Pendulum Monster in your Monster Zone and 1 monster your opponent controls; return them to the hand(s). You can only use this effect of "Majespecter Unicorn - Kirin" once per turn. While this card is in your Monster Zone, your opponent cannot target it with card effects, also it cannot be destroyed by your opponent's card effects.`,

				MonsterAttribute:     WIND,
				MonsterType:          Spellcaster,
				MonsterLevelRankLink: 6,
				MonsterATK:           2000,
				MonsterDEF:           2000,

				MiscKonamiSet:    "DOCS-EN029",
				MiscKonamiCardID: "11973",
				MiscYear:         "2015",

				IsPendulum:    true,
				PendulumScale: 2,
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

		{pageHTML: test_konami_12953, cardID: "12953",
			want: Card{
				CardName:    "Supreme King Z-ARC",
				CardType:    Monster,
				CardSubtype: MonsterFusion,
				CardEffect: `4 Dragon monsters (1 Fusion, 1 Synchro, 1 Xyz, and 1 Pendulum)
Must be Fusion Summoned. If this card is Special Summoned: Destroy all cards your opponent controls. Cannot be destroyed by your opponent's card effects. Your opponent cannot target this card with card effects. When this card destroys an opponent's monster by battle: You can Special Summon 1 "Supreme King Dragon" monster from your Deck or Extra Deck. If this card in the Monster Zone is destroyed by battle or card effect: You can place this card in your Pendulum Zone.`,

				MonsterAttribute:     DARK,
				MonsterType:          Dragon,
				MonsterLevelRankLink: 12,
				MonsterATK:           4000,
				MonsterDEF:           4000,

				MiscKonamiSet:    "MACR-EN039",
				MiscKonamiCardID: "12953",
				MiscYear:         "2017",

				IsPendulum:     true,
				PendulumScale:  1,
				PendulumEffect: `Fusion, Synchro, and Xyz Monsters your opponent controls cannot activate their effects. Once per turn, when a card(s) is added from the Main Deck to your opponent's hand (except during the Draw Phase or the Damage Step): You can destroy that card(s).`,
			}},

		{pageHTML: test_konami_13587, cardID: "13587",
			want: Card{
				CardName:    "Ghost Belle & Haunted Mansion",
				CardType:    Monster,
				CardSubtype: MonsterEffect,
				CardEffect: `When a card or effect is activated that includes any of these effects (Quick Effect): You can discard this card; negate that activation.
● Add a card(s) from the GY to the hand, Deck, and/or Extra Deck.
● Special Summon a Monster Card(s) from the GY.
● Banish a card(s) from the GY.
You can only use this effect of "Ghost Belle & Haunted Mansion" once per turn.`,

				MonsterAttribute:     EARTH,
				MonsterType:          Zombie,
				MonsterLevelRankLink: 3,
				MonsterATKStr:        "0",
				MonsterDEF:           1800,
				MonsterAbilities:     []MonsterAbility{Tuner},

				MiscKonamiSet:    "FLOD-EN033",
				MiscKonamiCardID: "13587",
				MiscYear:         "2018",
			}},

		{pageHTML: test_konami_13631, cardID: "13631",
			want: Card{
				CardName:    "Infinite Impermanence",
				CardType:    Trap,
				CardSubtype: TrapNormal,
				CardEffect:  `Target 1 face-up monster your opponent controls; negate its effects (until the end of this turn), then if this card was Set before activation and is on the field at resolution, for the rest of this turn all other Spell/Trap effects in this column are negated. If you control no cards, you can activate this card from your hand.`,

				MiscKonamiSet:    "FLOD-EN077",
				MiscKonamiCardID: "13631",
				MiscYear:         "2018",
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

		{pageHTML: test_konami_15296, cardID: "15296",
			want: Card{
				CardName:    "Triple Tactics Talent",
				CardType:    Spell,
				CardSubtype: SpellNormal,
				CardEffect: `If your opponent has activated a monster effect during your Main Phase this turn: Activate 1 of these effects;
● Draw 2 cards.
● Take control of 1 monster your opponent controls until the End Phase.
● Look at your opponent's hand, and choose 1 card from it to shuffle into the Deck.
You can only activate 1 "Triple Tactics Talent" per turn.`,

				MiscKonamiSet:    "ROTD-EN062",
				MiscKonamiCardID: "15296",
				MiscYear:         "2020",
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

		{pageHTML: test_konami_16386, cardID: "16386",
			want: Card{
				CardName:    "Baronne de Fleur",
				CardType:    Monster,
				CardSubtype: MonsterSynchro,
				CardEffect: `1 Tuner + 1+ non-Tuner monsters
Once per turn: You can target 1 card on the field; destroy it. Once while face-up on the field, when a card or effect is activated (Quick Effect): You can negate the activation, and if you do, destroy that card. You can only use the previous effect of "Baronne de Fleur" once per turn. Once per turn, during the Standby Phase: You can target 1 Level 9 or lower monster in your GY; return this card to the Extra Deck, and if you do, Special Summon that monster.`,

				MonsterAttribute:     WIND,
				MonsterType:          Warrior,
				MonsterLevelRankLink: 10,
				MonsterATK:           3000,
				MonsterDEF:           2400,

				MiscKonamiSet:    "LED8-EN024",
				MiscKonamiCardID: "16386",
				MiscYear:         "2021",
			}},

		{pageHTML: test_konami_16849, cardID: "16849",
			want: Card{
				CardName:    "D/D/D Deviser King Deus Machinex",
				CardType:    Monster,
				CardSubtype: MonsterXyz,
				CardEffect: `2 Level 10 Fiend monsters
You can also Xyz Summon this card by using a "D/D/D" monster you control as material. (Transfer its materials to this card.) You can only control 1 "D/D/D Deviser King Deus Machinex" in your Monster Zone. Once per Chain, when a Monster Card your opponent controls activates its effect (Quick Effect): You can either detach 2 materials from this card, or destroy 1 "Dark Contract" card you control, and if you do, attach that opponent's card to this card as material. Once per turn, during your Standby Phase: You can place this card in your Pendulum Zone.`,

				MonsterAttribute:     DARK,
				MonsterType:          Fiend,
				MonsterLevelRankLink: 10,
				MonsterATK:           3000,
				MonsterDEF:           3000,

				MiscKonamiSet:    "BACH-EN044",
				MiscKonamiCardID: "16849",
				MiscYear:         "2022",

				IsPendulum:     true,
				PendulumScale:  10,
				PendulumEffect: `While you have a card in your other Pendulum Zone: You can target 1 Pendulum Monster you control or in your GY; Special Summon the card in your other Pendulum Zone, and if you do, place that targeted Pendulum Monster in your Pendulum Zone. You can only use this effect of "D/D/D Deviser King Deus Machinex" once per turn.`,
			}},

		{pageHTML: test_konami_17474, cardID: "17474",
			want: Card{
				CardName:    "Tearlaments Sulliek",
				CardType:    Trap,
				CardSubtype: TrapContinuous,
				CardEffect:  `If you control a "Tearlaments" monster or "Visas Starfrost": You can target 1 Effect Monster your opponent controls; negate its effects, then send 1 monster you control to the GY. If this card is sent to the GY by card effect: You can add 1 "Tearlaments" monster from your Deck to your hand. You can only use each effect of "Tearlaments Sulliek" once per turn.`,

				MiscKonamiSet:    "POTE-EN072",
				MiscKonamiCardID: "17474",
				MiscYear:         "2022",
			}},

		{pageHTML: test_konami_17746, cardID: "17746",
			want: Card{
				CardName:    "Garura, Wings of Resonant Life",
				CardType:    Monster,
				CardSubtype: MonsterFusion,
				CardEffect: `2 monsters with the same Type and Attribute, but different names
Any battle damage your opponent takes from battles involving this card is doubled. If this card is sent to the GY: You can draw 1 card. You can only use this effect of "Garura, Wings of Resonant Life" once per turn.`,

				MonsterAttribute:     DARK,
				MonsterType:          WingedBeast,
				MonsterLevelRankLink: 6,
				MonsterATK:           1500,
				MonsterDEF:           2400,

				MiscKonamiSet:    "POTE-EN082",
				MiscKonamiCardID: "17746",
				MiscYear:         "2022",
			}},

		{pageHTML: test_konami_17808, cardID: "17808",
			want: Card{
				CardName:    "Branded Regained",
				CardType:    Spell,
				CardSubtype: SpellContinuous,
				CardEffect:  `If a LIGHT or DARK monster(s) is banished (except during the Damage Step): You can target 1 of those monsters; place that monster on the bottom of the Deck, and if you do, draw 1 card. You can only use this effect of "Branded Regained" once per turn. Once per turn, if your opponent Normal or Special Summons a monster (except during the Damage Step): You can target 1 "Bystial" monster in your GY; Special Summon it. You can only activate this effect of "Branded Regained" once per Chain.`,

				MiscKonamiSet:    "DABL-EN053",
				MiscKonamiCardID: "17808",
				MiscYear:         "2022",
			}},

		{pageHTML: test_konami_18022, cardID: "18022",
			want: Card{
				CardName:    "Mikanko Water Arabesque",
				CardType:    Spell,
				CardSubtype: SpellEquip,
				CardEffect:  `The equipped monster cannot be destroyed by card effects. During your Main Phase: You can Special Summon 1 "Mikanko" monster from your hand or Deck, with a different original name than the equipped monster, and if you do, equip it with this card, then return the monster that was previously equipped with this card to the hand. You can only use this effect of "Mikanko Water Arabesque" once per turn.`,

				MiscKonamiSet:    "AMDE-EN032",
				MiscKonamiCardID: "18022",
				MiscYear:         "2023",
			}},

		{pageHTML: test_konami_18177, cardID: "18177",
			want: Card{
				CardName:    "Evigishki Neremanas",
				CardType:    Monster,
				CardSubtype: MonsterRitual,
				CardEffect:  `You can Ritual Summon this card with any "Gishki" Ritual Spell. If this card is Ritual Summoned: You can target 1 WATER monster in your GY; Special Summon it. Cannot be destroyed by battle with a monster Special Summoned from the Extra Deck. Once per turn, when your opponent activates a monster effect (Quick Effect): You can return 1 "Gishki" Ritual Monster you control to the hand, and if you do, negate the activation, and if you do that, shuffle it into the Deck.`,

				MonsterAttribute:     WATER,
				MonsterType:          Spellcaster,
				MonsterLevelRankLink: 10,
				MonsterATK:           3000,
				MonsterDEF:           1800,

				MiscKonamiSet:    "PHHY-EN032",
				MiscKonamiCardID: "18177",
				MiscYear:         "2023",
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

		{pageHTML: test_konami_19188, cardID: "19188",
			want: Card{
				CardName:    "S:P Little Knight",
				CardType:    Monster,
				CardSubtype: MonsterLink,
				CardEffect: `2 Effect Monsters
If this card is Link Summoned using a Fusion, Synchro, Xyz, or Link Monster as material: You can target 1 card on the field or in either GY; banish it, also your monsters cannot attack directly this turn. When your opponent activates a card or effect (Quick Effect): You can target 2 face-up monsters on the field, including a monster you control; banish both until the End Phase. You can only use each effect of "S:P Little Knight" once per turn.`,

				MonsterAttribute:     DARK,
				MonsterType:          Warrior,
				MonsterLevelRankLink: 2,
				MonsterATK:           1600,
				MonsterDEF:           0,
				MonsterDEFStr:        "-",
				MonsterLinkArrows:    []MonsterLinkArrow{Left, Right},

				MiscKonamiSet:    "AGOV-EN046",
				MiscKonamiCardID: "19188",
				MiscYear:         "2023",
			}},

		{pageHTML: test_konami_19375, cardID: "19375",
			want: Card{
				CardName:    "Centur-Ion Legatia",
				CardNameEN:  "Centur-Ion Legatia",
				CardType:    Monster,
				CardSubtype: MonsterSynchro,
				CardEffect: `1 Tuner + 1+ non-Tuner monsters
Your monsters with 2000 or less ATK cannot be destroyed by battle. You can only use each of the following effects of "Centur-Ion Legatia" once per turn. If this card is Special Summoned: You can draw 1 card, then you can destroy the monster your opponent controls with the highest ATK (your choice, if tied). During the End Phase: You can place 1 non-Synchro "Centur-Ion" monster from your hand or GY in your Spell & Trap Zone as a face-up Continuous Trap.`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Machine,
				MonsterLevelRankLink: 12,
				MonsterATK:           3500,
				MonsterDEF:           2000,

				MiscKonamiSet:    "VASM-EN019",
				MiscKonamiCardID: "19375",
				MiscYear:         "2023",
			}},

		{pageHTML: test_konami_19376, cardID: "19376",
			want: Card{
				CardName:    "Stand Up Centur-Ion!",
				CardType:    Spell,
				CardSubtype: SpellField,
				CardEffect:  `Cannot be destroyed by your opponent's card effects while you control a "Centur-Ion" Monster Card. You can only use each of the following effects of "Stand Up Centur-Ion!" once per turn. During your Main Phase, if this card was activated this turn: You can send 1 card from your hand to the GY; place 1 "Centur-Ion" monster from your Deck in your Spell & Trap Zone as a face-up Continuous Trap. If a monster(s) is Special Summoned, you can: Immediately after this effect resolves, Synchro Summon 1 Synchro Monster, using monsters you control as material, including a "Centur-Ion" monster.`,

				MiscKonamiSet:    "VASM-EN020",
				MiscKonamiCardID: "19376",
				MiscYear:         "2023",
			}},

		{pageHTML: test_konami_19521, cardID: "19521",
			want: Card{
				CardName:    "Prayers of the Voiceless Voice",
				CardType:    Spell,
				CardSubtype: SpellRitual,
				CardEffect:  `This card can be used to Ritual Summon any LIGHT Ritual Monster. You must also Tribute LIGHT monsters from your hand or field whose total Levels equal or exceed the Level of the Ritual Monster. If a face-up LIGHT Ritual Monster(s) you control leaves the field by an opponent's card effect (except during the Damage Step): You can banish this card from your GY; Special Summon 1 "Sauravis, the Ancient and Ascended", "Saffira, Queen of Dragons", or "Skull Guardian, Protector of the Voiceless Voice" from your hand or Deck, ignoring its Summoning conditions. You can only use this effect of "Prayers of the Voiceless Voice" once per turn.`,

				MiscKonamiSet:    "PHNI-EN066",
				MiscKonamiCardID: "19521",
				MiscYear:         "2024",
			}},

		{pageHTML: test_konami_20536, cardID: "20536",
			want: Card{
				CardName:    "Primite Drillbeam",
				CardType:    Spell,
				CardSubtype: SpellQuickPlay,
				CardEffect:  `Reveal 1 "Primite" card, or 1 Normal Monster, in your hand, except "Primite Drillbeam" (or if you control a Normal Monster or a Level 5 or higher "Primite" monster, except a Token, you can activate this effect without revealing a card), then target 1 face-up card on the field; negate its effects, and if you do, banish it. During your Main Phase, if you control a "Primite" monster: You can Set this card from your GY. You can only use each effect of "Primite Drillbeam" once per turn.`,

				MiscKonamiSet:    "ROTA-EN060",
				MiscKonamiCardID: "20536",
				MiscYear:         "2024",
			}},

		{pageHTML: test_konami_20578, cardID: "20578",
			want: Card{
				CardName:    "Ryzeal Detonator",
				CardType:    Monster,
				CardSubtype: MonsterXyz,
				CardEffect: `2+ Level 4 "Ryzeal" monsters
When your opponent activates a card or effect (Quick Effect): You can detach 1 material from this card, then target 1 card on the field; destroy it. You can only use each of the following effects of "Ryzeal Detonator" once per turn. If this card is Special Summoned: You can attach 1 monster from your GY to this card as material. If an Xyz Monster(s) you control would be destroyed by battle or card effect, you can detach 1 material from this card instead.`,

				MonsterAttribute:     LIGHT,
				MonsterType:          Pyro,
				MonsterLevelRankLink: 4,
				MonsterATK:           3000,
				MonsterDEF:           2500,

				MiscKonamiSet:    "CRBR-EN007",
				MiscKonamiCardID: "20578",
				MiscYear:         "2024",
			}},

		{pageHTML: test_konami_21627, cardID: "21627",
			want: Card{
				CardName:    "Miracle Raven",
				CardType:    Monster,
				CardSubtype: MonsterRitual,
				CardEffect:  `You can Ritual Summon this card with "Miracle Raven". Must be Ritual Summoned. This Ritual Summoned card is unaffected by your opponent's activated effects. If you Ritual Summon exactly 1 Ritual Monster with a card effect that requires use of monsters, this card you control can be used as the entire Tribute. If this card is Tributed for a Ritual Summon: You can add 1 Ritual Monster from your Deck to your hand. You can only use this effect of "Miracle Raven" once per turn.`,

				MonsterAttribute:     DARK,
				MonsterType:          Fiend,
				MonsterLevelRankLink: 1,
				MonsterATK:           300,
				MonsterDEF:           300,

				MiscKonamiSet:    "DUAD-EN084",
				MiscKonamiCardID: "21627",
				MiscYear:         "2025",

				IsPendulum:     true,
				PendulumScale:  0,
				PendulumEffect: `Once per turn, during your Main Phase: You can Ritual Summon this card, by Tributing monsters from your hand or field whose total Levels equal or exceed 1.`,
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
