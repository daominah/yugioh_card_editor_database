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

func TestPrepareRushDuelData(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error Getwd: %v", err)
	}
	outputDir := filepath.Join(pwd, "test_rushduel_html")
	t.Logf("outputDir TestPrepareRushDuelData: %v", outputDir)

	t.Skip("only need to run once at the start of coding to prepare data")

	for _, locale := range []string{"ja", "ko"} {
		for _, cardID := range []string{
			"15184", // Blue-Eyes White Dragon
			"16372", // Pot of Greed
			"19096", // Harpy Lady Sisters [L]
			"19097", // Harpy Lady Sisters (center Maximum Monster)
			"19105", // Harpie's Full Dress (Equip Spell)
			"19342", // Negate Attack
			"20453", // Elemental HERO Flame Wingman
			"21588", // Magician of Black Chaos (Ritual Monster)
			"21597", // Ritual of Black Chaos (Ritual Spell — pairs with 21588)
		} {
			cardURL := `https://www.db.yugioh-card.com/rushdb/card_search.action?ope=2` +
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
	//go:embed test_rushduel_html_ja/test_konami_15184.html
	test_rushduel_15184 []byte
	// 15184: Blue-Eyes White Dragon

	//go:embed test_rushduel_html_ja/test_konami_16372.html
	test_rushduel_16372 []byte
	// 16372: Pot of Greed

	//go:embed test_rushduel_html_ja/test_konami_19096.html
	test_rushduel_19096 []byte
	// 19096: Harpy Lady Sisters [L]

	//go:embed test_rushduel_html_ja/test_konami_19097.html
	test_rushduel_19097 []byte
	// 19097: Harpy Lady Sisters (center Maximum Monster)

	//go:embed test_rushduel_html_ja/test_konami_19105.html
	test_konami_19105 []byte
	// 19105: Harpie's Full Dress (Equip Spell)

	//go:embed test_rushduel_html_ja/test_konami_19342.html
	test_rushduel_19342 []byte
	// 19342: Negate Attack

	//go:embed test_rushduel_html_ja/test_konami_20453.html
	test_rushduel_20453 []byte
	// 20453: Elemental HERO Flame Wingman

	//go:embed test_rushduel_html_ja/test_konami_21588.html
	test_rushduel_21588 []byte
	// 21588: Magician of Black Chaos (Ritual Monster)

	//go:embed test_rushduel_html_ja/test_konami_21597.html
	test_rushduel_21597 []byte
	// 21597: Ritual of Black Chaos (Ritual Spell — pairs with 21588)
)

func TestParseRushDuelCardHTML(t *testing.T) {
	for _, c := range []struct {
		cardID   string
		pageHTML []byte
		want     CardRushDuel
	}{
		{pageHTML: test_rushduel_15184, cardID: "15184",
			want: CardRushDuel{
				Card: Card{
					CardName:    "青眼の白龍", // Blue-Eyes White Dragon
					CardType:    Monster,
					CardSubtype: MonsterNormal,
					CardEffect:  `高い攻撃力を誇る伝説のドラゴン。どんな相手でも粉砕する、その破壊力は計り知れない。`,
					// A legendary dragon with immense attack power. Its destructive force is immeasurable, capable of crushing any opponent.

					MonsterAttribute:     LIGHT,
					MonsterType:          Dragon,
					MonsterLevelRankLink: 8,
					MonsterATK:           3000,
					MonsterDEF:           2500,
					IsNonEffectMonster:   true,

					MiscKonamiSet:    "RD/KP01-JP000",
					MiscKonamiCardID: "15184",
					MiscYear:         "2020",
				},
				RushIsLegend:   true,
				RushMaximumATK: "",
			}},

		{pageHTML: test_rushduel_16372, cardID: "16372",
			want: CardRushDuel{
				Card: Card{
					CardName:    "強欲な壺", // Pot of Greed
					CardType:    Spell,
					CardSubtype: SpellNormal,
					CardEffect: `【条件】なし
【効果】自分は2枚ドローする。`,
					// [Condition] None.
					// [Effect] Draw 2 cards.

					MiscKonamiSet:    "RD/G001-JP003",
					MiscKonamiCardID: "16372",
					MiscYear:         "2021",
				},
				RushIsLegend: true,
			}},

		{pageHTML: test_rushduel_19096, cardID: "19096",
			want: CardRushDuel{
				Card: Card{
					CardName:    "ハーピィ三姉妹［L］", // Harpy Lady Sisters [L]
					CardType:    Monster,
					CardSubtype: MonsterEffect,
					CardEffect: `墓地にいるこのカードのカード名は「ハーピィ三姉妹」になる。
【条件】なし
【永続効果】このカードのカード名は「ハーピィ三姉妹」になり、相手の罠カードの効果では破壊されない。このカードがマキシマムモードの場合、さらにこのカードのレベルは5上がる。`,
					// While in the Graveyard, this card's name becomes "Harpy Lady Sisters".
					// [Condition] None.
					// [Continuous Effect] This card's name becomes "Harpy Lady Sisters" and cannot be destroyed by the opponent's Trap card effects. If this card is in Maximum Mode, also increase this card's Level by 5.

					MonsterAttribute:     WIND,
					MonsterType:          WingedBeast,
					MonsterLevelRankLink: 5,
					MonsterATK:           2100,
					MonsterDEF:           0,
					MonsterAbilities:     []MonsterAbility{RushMaximum},
					IsNonEffectMonster:   false,

					MiscKonamiSet:    "RD/TB01-JP001",
					MiscKonamiCardID: "19096",
					MiscYear:         "2023",
				},
				RushIsLegend:   false,
				RushMaximumATK: "",
			}},

		{pageHTML: test_rushduel_19097, cardID: "19097",
			want: CardRushDuel{
				Card: Card{
					CardName:    "ハーピィ三姉妹", // Harpy Lady Sisters (center Maximum Monster)
					CardType:    Monster,
					CardSubtype: MonsterEffect,
					CardEffect: `「ハーピィ三姉妹[L]」「ハーピィ三姉妹[R]」と揃えてマキシマム召喚できる。
【条件】自分の墓地のモンスター(風属性/鳥獣族)2体をデッキに戻して発動できる。
【効果】自分フィールドの表側表示モンスター1体を選び、その攻撃力をターン終了時まで500アップする。このカードがマキシマムモードの場合、さらに相手に500ダメージを与える。`,
					// Can be Maximum Summoned together with "Harpy Lady Sisters [L]" and "Harpy Lady Sisters [R]".
					// [Condition] Return 2 monsters (WIND/Winged Beast) from your Graveyard to your Deck.
					// [Effect] Choose 1 face-up monster on your field and increase its ATK by 500 until end of turn. If this card is in Maximum Mode, also deal 500 damage to the opponent.

					MonsterAttribute:     WIND,
					MonsterType:          WingedBeast,
					MonsterLevelRankLink: 5,
					MonsterATK:           2100,
					MonsterDEF:           0,
					MonsterAbilities:     []MonsterAbility{RushMaximum},
					IsNonEffectMonster:   false,

					MiscKonamiSet:    "RD/TB01-JP002",
					MiscKonamiCardID: "19097",
					MiscYear:         "2023",
				},
				RushIsLegend:   false,
				RushMaximumATK: "3400",
			}},

		{pageHTML: test_konami_19105, cardID: "19105",
			want: CardRushDuel{
				Card: Card{
					CardName:    "ハーピィズフルドレス", // Harpie's Full Dress
					CardType:    Spell,
					CardSubtype: SpellEquip,
					CardEffect: `【条件】自分フィールドの表側表示の「ハーピィ・レディ」または「ハーピィ三姉妹」1体に装備できる。
【効果】装備モンスターの攻撃力は800アップし、守備力は400アップする。自分フィールドの表側表示モンスターが1体または3体の場合、装備モンスターの攻撃は貫通する。`,
					// [Condition] Equip to 1 face-up "Harpie Lady" or "Harpy Lady Sisters" on your field.
					// [Effect] The equipped monster gains 800 ATK and 400 DEF. If you control 1 or 3 face-up monsters, the equipped monster's attacks pierce.

					MiscKonamiSet:    "RD/TB01-JP012",
					MiscKonamiCardID: "19105",
					MiscYear:         "2023",
				},
				RushIsLegend: false,
			}},

		{pageHTML: test_rushduel_19342, cardID: "19342",
			want: CardRushDuel{
				Card: Card{
					CardName:    "攻撃の無力化", // Negate Attack
					CardType:    Trap,
					CardSubtype: TrapNormal,
					CardEffect: `【条件】相手モンスターの攻撃宣言時に発動できる。
【効果】その攻撃を無効にする。このターン、相手は攻撃宣言できない。`,
					// [Condition] Can be activated when the opponent declares an attack with a monster.
					// [Effect] Negate that attack. The opponent cannot declare attacks this turn.

					MiscKonamiSet:    "RD/KP14-JP065",
					MiscKonamiCardID: "19342",
					MiscYear:         "2023",
				},
				RushIsLegend: true,
			}},

		{pageHTML: test_rushduel_20453, cardID: "20453",
			want: CardRushDuel{
				Card: Card{
					CardName:    "E・HERO フレイム・ウィングマン", // Elemental HERO Flame Wingman
					CardType:    Monster,
					CardSubtype: MonsterEffect,
					CardEffect: `「E・HERO フェザーマン」+「E・HERO バーストレディ」
このカードはフュージョン召喚でしか特殊召喚できない。
【条件】なし
【永続効果】このカードが戦闘でモンスターを破壊し墓地へ送った時、[そのモンスターの元々の攻撃力]だけ相手にダメージを与える。`,
					// "Elemental HERO Featherman" + "Elemental HERO Burst Lady"
					// This card can only be Special Summoned by Fusion Summon.
					// [Condition] None.
					// [Continuous Effect] When this card destroys a monster in battle and sends it to the Graveyard, deal damage to the opponent equal to [that monster's original ATK].

					MonsterAttribute:     WIND,
					MonsterType:          Warrior,
					MonsterLevelRankLink: 6,
					MonsterATK:           2100,
					MonsterDEF:           1200,
					IsNonEffectMonster:   false,

					MiscKonamiSet:    "RD/SD0B-JPS01",
					MiscKonamiCardID: "20453",
					MiscYear:         "2024",
				},
				RushIsLegend:   false,
				RushMaximumATK: "",
			}},

		{pageHTML: test_rushduel_21588, cardID: "21588",
			want: CardRushDuel{
				Card: Card{
					CardName:    "マジシャン・オブ・ブラックカオス", // Magician of Black Chaos
					CardType:    Monster,
					CardSubtype: MonsterEffect,
					CardEffect: `墓地にいるこのカードのカード名は「ブラック・マジシャン」になる。
【条件】なし
【効果】なし`,
					// While in the Graveyard, this card's name becomes "Dark Magician".
					// [Condition] None.
					// [Effect] None.

					MonsterAttribute:     DARK,
					MonsterType:          Spellcaster,
					MonsterLevelRankLink: 8,
					MonsterATK:           2800,
					MonsterDEF:           2600,
					IsNonEffectMonster:   false,

					MiscKonamiSet:    "RD/SD0E-JPS01",
					MiscKonamiCardID: "21588",
					MiscYear:         "2025",
				},
				RushIsLegend:   false,
				RushMaximumATK: "",
			}},

		{pageHTML: test_rushduel_21597, cardID: "21597",
			want: CardRushDuel{
				// Ritual Spell card. Currently the Rush parser misclassifies it as
				// a Monster with attribute "リチュアル魔法"; this case is a
				// regression test for that bug (cards_rush had 32 such rows).
				Card: Card{
					CardName:    "カオス－黒魔術の儀式", // Ritual of Black Chaos
					CardType:    Spell,
					CardSubtype: SpellRitual,

					MiscKonamiCardID: "21597",
				},
			}},
	} {
		got := ParseRushDuelCardHTML(c.pageHTML, CardID(c.cardID))
		//t.Logf("parsed card %v: %+v", c.cardID, got)
		want := c.want
		if got.CardName != want.CardName {
			t.Errorf("cardID %v CardName got %q, want %q", c.cardID, got.CardName, want.CardName)
		}
		if got.CardType != want.CardType {
			t.Errorf("cardID %v CardType got %q, want %q", c.cardID, got.CardType, want.CardType)
		}
		if got.CardSubtype != want.CardSubtype {
			t.Errorf("cardID %v CardSubtype got %q, want %q", c.cardID, got.CardSubtype, want.CardSubtype)
		}
		if want.CardEffect != "" && got.CardEffect != want.CardEffect {
			t.Errorf("cardID %v CardEffect got %q, want %q", c.cardID, got.CardEffect, want.CardEffect)
		}
		if got.MonsterAttribute != want.MonsterAttribute {
			t.Errorf("cardID %v MonsterAttribute got %q, want %q", c.cardID, got.MonsterAttribute, want.MonsterAttribute)
		}
		if got.MonsterType != want.MonsterType {
			t.Errorf("cardID %v MonsterType got %q, want %q", c.cardID, got.MonsterType, want.MonsterType)
		}
		if got.MonsterLevelRankLink != want.MonsterLevelRankLink {
			t.Errorf("cardID %v MonsterLevelRankLink got %v, want %v", c.cardID, got.MonsterLevelRankLink, want.MonsterLevelRankLink)
		}
		if got.MonsterATK != want.MonsterATK {
			t.Errorf("cardID %v MonsterATK got %v, want %v", c.cardID, got.MonsterATK, want.MonsterATK)
		}
		if got.MonsterDEF != want.MonsterDEF {
			t.Errorf("cardID %v MonsterDEF got %v, want %v", c.cardID, got.MonsterDEF, want.MonsterDEF)
		}
		if !CheckEqualArray(got.MonsterAbilities, want.MonsterAbilities) {
			t.Errorf("cardID %v MonsterAbilities got %v, want %v", c.cardID, got.MonsterAbilities, want.MonsterAbilities)
		}
		if got.IsNonEffectMonster != want.IsNonEffectMonster {
			t.Errorf("cardID %v IsNonEffectMonster got %v, want %v", c.cardID, got.IsNonEffectMonster, want.IsNonEffectMonster)
		}
		if got.RushIsLegend != want.RushIsLegend {
			t.Errorf("cardID %v RushIsLegend got %v, want %v", c.cardID, got.RushIsLegend, want.RushIsLegend)
		}
		if got.RushMaximumATK != want.RushMaximumATK {
			t.Errorf("cardID %v RushMaximumATK got %q, want %q", c.cardID, got.RushMaximumATK, want.RushMaximumATK)
		}
		if want.MiscKonamiSet != "" && got.MiscKonamiSet != want.MiscKonamiSet {
			t.Errorf("cardID %v MiscKonamiSet got %q, want %q", c.cardID, got.MiscKonamiSet, want.MiscKonamiSet)
		}
		if want.MiscYear != "" && got.MiscYear != want.MiscYear {
			t.Errorf("cardID %v MiscYear got %q, want %q", c.cardID, got.MiscYear, want.MiscYear)
		}
	}
}
