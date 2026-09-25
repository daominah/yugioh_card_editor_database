package konami

import (
	_ "embed"
	"strings"
	"testing"
)

// Fixtures in test_html_escaped_br/ were fetched on 2026-09-16,
// after Konami changed card pages to wrap card text in <div class="text_linebreak">
// and write line breaks as escaped text "&lt;br&gt;" instead of a <br> element.
// Older fixtures in test_html_en/ etc. still use the <br> element.
var (
	//go:embed test_html_escaped_br/test_konami_15524_en.html
	test_escaped_br_15524_en []byte
	// 15524: Divine Arsenal AA-ZEUS - Sky Thunder (Xyz Monster, materials line)

	//go:embed test_html_escaped_br/test_konami_14356_en.html
	test_escaped_br_14356_en []byte
	// 14356: Time Thief Redoer (Xyz Monster, bullet points)

	//go:embed test_html_escaped_br/test_konami_12144_en.html
	test_escaped_br_12144_en []byte
	// 12144: Guiding Ariadne (Pendulum Monster, bullet points in Pendulum Effect)

	//go:embed test_html_escaped_br/test_konami_15524_ja.html
	test_escaped_br_15524_ja []byte
	// 15524: 天霆號アーゼウス (Xyz Monster, numbered effects)

	//go:embed test_html_escaped_br/test_rushduel_19105_ja.html
	test_escaped_br_rush_19105_ja []byte
	// 19105: Rush Duel Equip Spell with 【条件】 and 【効果】 lines
)

func TestParseKonamiCardHTMLEscapedBr(t *testing.T) {
	for _, c := range []struct {
		cardID             string
		pageHTML           []byte
		wantEffect         string
		wantPendulumEffect string
	}{
		{cardID: "15524", pageHTML: test_escaped_br_15524_en,
			wantEffect: `2 Level 12 monsters
Once per turn, if an Xyz Monster battled this turn, you can also Xyz Summon "Divine Arsenal AA-ZEUS - Sky Thunder" by using 1 Xyz Monster you control (transfer its materials). (Quick Effect): You can detach 2 materials from this card; send all other cards from the field to the GY. Once per turn, if another card(s) you control is destroyed by battle or an opponent's card effect: You can attach 1 card from your hand, Deck, or Extra Deck to this card.`,
		},
		{cardID: "14356", pageHTML: test_escaped_br_14356_en,
			wantEffect: `2 Level 4 monsters
Once per turn, during the Standby Phase: You can attach the top card of your opponent's Deck to this card. (Quick Effect): You can detach up to 3 different types of materials from this card, then apply the following effect(s) depending on what was detached.
● Monster: Banish this card until the End Phase.
● Spell: Draw 1 card.
● Trap: Place 1 face-up card your opponent controls on top of the Deck.
You can only use this effect of "Time Thief Redoer" once per turn.`,
		},
		{cardID: "12144", pageHTML: test_escaped_br_12144_en,
			wantEffect: `If this card is destroyed by battle or card effect: You can reveal 3 Counter Traps from your Deck, your opponent chooses 1 for you to add to your hand, and you shuffle the rest back into your Deck.`,
			wantPendulumEffect: `Apply these effects.
●You do not pay LP to activate Counter Trap Cards.
●You do not discard to activate Counter Trap Cards.`,
		},
	} {
		// GIVEN a card page in Konami's new markup, with escaped "<br>" in the card text
		// WHEN the page is parsed as a Standard card
		got := ParseKonamiCardHTML(c.pageHTML, CardID(c.cardID))

		// THEN each "<br>" becomes a line break in the effects
		if got.CardEffect != c.wantEffect {
			t.Errorf("error cardID %v CardEffect got:\n%v\n, want:\n%v\n", c.cardID, got.CardEffect, c.wantEffect)
		}
		if got.PendulumEffect != c.wantPendulumEffect {
			t.Errorf("error cardID %v PendulumEffect got:\n%v\n, want:\n%v\n", c.cardID, got.PendulumEffect, c.wantPendulumEffect)
		}
	}
}

func TestParseCardLocaleTextEscapedBr(t *testing.T) {
	// GIVEN a Japanese card page in Konami's new markup, whose effect has 4 lines
	// WHEN the page is parsed as locale text for the database
	got := ParseCardLocaleText(test_escaped_br_15524_ja, "15524")

	// THEN the stored effect has 4 lines and no literal "<br>"
	if strings.Contains(got.Effect, "<br") {
		t.Errorf("error Effect contains literal <br>:\n%v", got.Effect)
	}
	if lines := strings.Split(got.Effect, "\n"); len(lines) != 4 {
		t.Errorf("error Effect got %v lines, want 4:\n%v", len(lines), got.Effect)
	}
}

func TestParseRushDuelCardHTMLEscapedBr(t *testing.T) {
	// GIVEN a Rush Duel card page in Konami's new markup, with escaped "<br>" in the card text
	// WHEN the page is parsed as a Rush Duel card
	got := ParseRushDuelCardHTML(test_escaped_br_rush_19105_ja, "19105")

	// THEN the condition and the effect are on separate lines
	want := `【条件】自分フィールドの表側表示の「ハーピィ・レディ」または「ハーピィ三姉妹」1体に装備できる。
【効果】装備モンスターの攻撃力は800アップし、守備力は400アップする。自分フィールドの表側表示モンスターが1体または3体の場合、装備モンスターの攻撃は貫通する。`
	if got.CardEffect != want {
		t.Errorf("error CardEffect got:\n%v\n, want:\n%v\n", got.CardEffect, want)
	}
}
