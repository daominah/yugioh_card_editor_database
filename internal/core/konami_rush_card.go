package core

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
)

// CardRushDuel represents a Yu-Gi-Oh Rush Duel card.
// It embeds Card for all standard fields and adds Rush Duel specific fields.
type CardRushDuel struct {
	Card

	// RushIsLegend marks the card as a Legend card, a Rush Duel exclusive category
	// for powerful OCG/TCG cards imported into the Rush Duel format.
	// Deck-building rule: at most 1 Legend card per card type (Monster, Spell, Trap)
	// may be included in a deck.
	RushIsLegend bool

	// RushMaximumATK is the MAXIMUM ATK value of the center Maximum Monster card
	// (e.g. "3400"). Empty string means the card is not a Maximum Monster.
	// A Maximum Summon requires 3 specific Maximum Monster cards in hand played
	// simultaneously; only the center card has a MAXIMUM ATK value printed above
	// its normal ATK/DEF, which applies only while in Maximum Mode.
	RushMaximumATK string
}

// mapRushDuelSpellTrapTypes maps the Rush Duel attribute-slot text for Spell
// and Trap cards to their CardType and CardSubtype. Spell/Trap cards display
// their category (e.g. "通常魔法") in the slot where Monster cards show an
// element attribute, so this lookup distinguishes the two cases.
var mapRushDuelSpellTrapTypes = map[string]struct {
	cardType    CardType
	cardSubtype CardSubtype
}{
	"通常魔法":   {Spell, SpellNormal},
	"速攻魔法":   {Spell, SpellQuickPlay},
	"フィールド魔法": {Spell, SpellField},
	"装備魔法":   {Spell, SpellEquip},
	"永続魔法":   {Spell, SpellContinuous},
	"通常罠":    {Trap, TrapNormal},
	"永続罠":    {Trap, TrapContinuous},
	"カウンター罠": {Trap, TrapCounter},
}

// MapRushDuelMonsterAttributes maps Rush Duel HTML attribute text to MonsterAttribute.
// Rush Duel uses localized text like "光属性" (Japanese), unlike the standard
// DB which uses "LIGHT".
var MapRushDuelMonsterAttributes = map[string]MonsterAttribute{
	"光属性": LIGHT,
	"闇属性": DARK,
	"地属性": EARTH,
	"炎属性": FIRE,
	"水属性": WATER,
	"風属性": WIND,
	"神属性": DIVINE,
}

// MapRushDuelMonsterTypesJP maps Japanese monster type names to MonsterType.
// English Rush Duel uses "Dragon-Type" format handled by stripping the "-Type"
// suffix and looking up in MapMonsterTypes.
var MapRushDuelMonsterTypesJP = map[string]MonsterType{
	"悪魔族":    Fiend,
	"アンデット族": Zombie,
	"機械族":    Machine,
	"魔法使い族":  Spellcaster,
	"戦士族":    Warrior,
	"竜族":     Dragon,
	"ドラゴン族":  Dragon,
	"天使族":    Fairy,
	"海竜族":    SeaSerpent,
	"昆虫族":    Insect,
	"岩石族":    Rock,
	"獣族":     Beast,
	"獣戦士族":   BeastWarrior,
	"植物族":    Plant,
	"炎族":     Pyro,
	"雷族":     Thunder,
	"爬虫類族":   Reptile,
	"魚族":     Fish,
	"サイキック族": Psychic,
	"水族":     Aqua,
	"恐竜族":    Dinosaur,
	"鳥獣族":    WingedBeast,
	"神獣族":    DivineBeast,
	"創造神族":   CreatorGod,
	"サイバース族": Cyberse,
	"幻想魔族":   Illusion,
	"幻竜族":    Wyrm,
}

// ParseRushDuelCardHTML parses a Rush Duel card detail page from the Konami
// Rush Duel card database (db.yugioh-card.com/rushdb) and returns a CardRushDuel.
func ParseRushDuelCardHTML(cardPageHTML []byte, cardID CardID) CardRushDuel {
	c := CardRushDuel{Card: Card{MiscKonamiCardID: cardID}}
	root := textproc.HTMLParseToNode(cardPageHTML)

	// getNode returns the first matched node or an empty node on no match
	getNode := func(parent *html.Node, xpath string) *html.Node {
		nodes, _ := textproc.HTMLXPath(parent, xpath)
		if len(nodes) == 0 {
			return &html.Node{}
		}
		return nodes[0]
	}

	// Remove related-card section to avoid picking up their item_box values
	relationNode := getNode(root, `//*[@id="relationCard"]`)
	if relationNode.Parent != nil {
		relationNode.Parent.RemoveChild(relationNode)
	}

	// Card name: h1 contains a <span class="ruby"> with an alternate reading
	// followed by a direct text node with the actual card name. We walk the
	// h1's children to collect only text nodes, skipping element children.
	cardNameNode := getNode(root, `//*[@id="cardname"]`)
	c.CardName = rushDuelH1TextOnly(cardNameNode)

	// Card effect: second CardText div contains the effect in item_box_text
	cardTexts, err := textproc.HTMLXPath(root, `//*[@class="CardText"]`)
	if len(cardTexts) < 2 {
		log.Printf("error rushDuel cardID %v cardTexts len: %v, %v\n", cardID, len(cardTexts), err)
		return c
	}
	cardEffect := textproc.HTMLGetText(cardTexts[1])
	cardEffect = strings.TrimSpace(cardEffect)
	cardEffect = strings.TrimPrefix(cardEffect, "Card Text")
	cardEffect = strings.TrimPrefix(cardEffect, "カードテキスト")
	c.CardEffect = strings.TrimSpace(cardEffect)

	// Legend: presence of <div id="legend"> marks the card as a Legend card
	legendNode := getNode(root, `//*[@id="legend"]`)
	if legendNode.Parent != nil {
		c.RushIsLegend = true
	}

	// All item_box_value spans in order: [attribute/type, level, (maxATK?), ATK, DEF]
	values, _ := textproc.HTMLXPath(root, `//span[@class="item_box_value"]`)

	// Spell/Trap cards show their category text (e.g. "通常魔法") in the first
	// slot; Monster cards show an element attribute (e.g. "光属性") there.
	if len(values) >= 1 {
		attributeS := strings.TrimSpace(textproc.HTMLGetText(values[0]))
		if st, ok := mapRushDuelSpellTrapTypes[attributeS]; ok {
			c.CardType = st.cardType
			c.CardSubtype = st.cardSubtype
		} else {
			c.CardType = Monster
			found := false
			c.MonsterAttribute, found = MapRushDuelMonsterAttributes[attributeS]
			if !found {
				log.Printf("error rushDuel cardID %v MonsterAttribute: %q\n", cardID, attributeS)
			}
		}
	}

	if c.CardType == Monster {
		// MAXIMUM ATK: center Maximum Monster cards have an extra item_box with
		// class "maxatk" before the regular ATK/DEF item_boxes.
		maxAtkValueNodes, _ := textproc.HTMLXPath(root,
			`//*[contains(@class,"maxatk")]//*[@class="item_box_value"]`)
		if len(maxAtkValueNodes) > 0 {
			c.RushMaximumATK = strings.TrimSpace(
				textproc.HTMLGetText(maxAtkValueNodes[0]))
		}

		if len(values) >= 2 {
			levelS := textproc.HTMLGetText(values[1])
			var digits []rune
			for _, r := range levelS {
				if NumberChars[r] {
					digits = append(digits, r)
				}
			}
			c.MonsterLevelRankLink, err = strconv.Atoi(string(digits))
			if err != nil {
				log.Printf("error rushDuel cardID %v MonsterLevelRankLink: %q\n", cardID, levelS)
			}
		}

		// ATK index is 2 normally; shifts to 3 when MAXIMUM ATK occupies index 2
		atkIdx := 2
		if c.RushMaximumATK != "" {
			atkIdx = 3
		}
		if len(values) > atkIdx {
			atkS := strings.TrimSpace(textproc.HTMLGetText(values[atkIdx]))
			c.MonsterATKStr = atkS
			c.MonsterATK, err = strconv.ParseFloat(atkS, 64)
			if err != nil && !SpecialATKDEF[atkS] {
				log.Printf("error rushDuel cardID %v MonsterATK: %q\n", cardID, atkS)
			}
		}
		if len(values) > atkIdx+1 {
			defS := strings.TrimSpace(textproc.HTMLGetText(values[atkIdx+1]))
			c.MonsterDEFStr = defS
			c.MonsterDEF, err = strconv.ParseFloat(defS, 64)
			if err != nil && !SpecialATKDEF[defS] {
				log.Printf("error rushDuel cardID %v MonsterDEF: %q\n", cardID, defS)
			}
		}

		// Species: "Dragon-Type / Normal", "Winged Beast / Maximum/Effect", "鳥獣族/マキシマム/効果"
		speciesS := textproc.HTMLGetText(getNode(root, `//*[@class="species"]`))
		rushDuelParseSpecies(&c, speciesS)
	}

	// Set and year: same update_list structure as the standard Konami DB
	konamiSets, err := textproc.HTMLXPath(root,
		`//*[@id="update_list"]//*[starts-with(@class,"t_row")]`)
	if err != nil {
		log.Printf("error rushDuel cardID %v konamiSets: %v\n", cardID, err)
	}
	for i := len(konamiSets) - 1; i >= 0; i-- {
		firstSet := konamiSets[i]
		ymd := textproc.HTMLGetText(getNode(firstSet, `//*[@class="time"]`))
		if len(ymd) >= 4 {
			c.MiscYear = ymd[:4]
		}
		c.MiscKonamiSet = textproc.HTMLGetText(
			getNode(firstSet, `//*[@class="card_number"]`))
		if c.MiscYear != "" && c.MiscKonamiSet != "" {
			break
		}
	}

	return c
}

// rushDuelH1TextOnly returns the direct text-node content of the first h1
// inside node, skipping child elements such as <span class="ruby">.
// Rush Duel card name h1s contain a ruby span with an alternate reading
// followed by the actual card name as a plain text node.
func rushDuelH1TextOnly(node *html.Node) string {
	h1nodes, _ := textproc.HTMLXPath(node, `//h1`)
	if len(h1nodes) == 0 {
		return strings.TrimSpace(textproc.HTMLGetText(node))
	}
	var parts []string
	for child := h1nodes[0].FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			t := strings.TrimSpace(child.Data)
			if t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, " ")
}

// rushDuelParseSpecies parses the species/type line of a Rush Duel monster card
// and fills MonsterType, CardSubtype, MonsterAbilities, and IsNonEffectMonster
// on the given card. The species text uses "/" or "／" as separators and
// contains monster type, subtype keywords, and ability keywords, e.g.:
// "Dragon-Type / Normal", "Winged Beast / Maximum/Effect", "鳥獣族/マキシマム/効果"
func rushDuelParseSpecies(card *CardRushDuel, speciesText string) {
	speciesText = strings.ReplaceAll(speciesText, "／", "/")
	tokens := strings.Split(speciesText, "/")

	card.IsNonEffectMonster = true
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		switch token {
		case "Normal", "通常":
			card.CardSubtype = MonsterNormal
			continue
		case "Effect", "効果":
			card.CardSubtype = MonsterEffect
			card.IsNonEffectMonster = false
			continue
		case "Maximum", "マキシマム":
			card.MonsterAbilities = append(card.MonsterAbilities, RushMaximum)
			continue
		case "Fusion", "融合":
			card.CardSubtype = MonsterFusion
			card.IsNonEffectMonster = false
			continue
		case "Ritual", "儀式":
			card.CardSubtype = MonsterRitual
			card.IsNonEffectMonster = false
			continue
		}

		// English: "Dragon-Type" → strip "-Type" → lookup in MapMonsterTypes
		cleanType := strings.TrimSuffix(token, "-Type")
		if mt, ok := MapMonsterTypes[cleanType]; ok {
			card.MonsterType = mt
			continue
		}
		// English without suffix: "Winged Beast", "Dragon", etc.
		if mt, ok := MapMonsterTypes[token]; ok {
			card.MonsterType = mt
			continue
		}
		// Japanese
		if mt, ok := MapRushDuelMonsterTypesJP[token]; ok {
			card.MonsterType = mt
			continue
		}
	}

	sort.Slice(card.MonsterAbilities, func(i, j int) bool {
		return card.MonsterAbilities[i] < card.MonsterAbilities[j]
	})
}
