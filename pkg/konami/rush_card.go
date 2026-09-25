package konami

import (
	"log"
	"slices"
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

// rushDuelSpellTrapType returns the CardType and CardSubtype for a localized
// Rush Duel spell/trap category label (e.g. "通常魔法", "일반 마법"), or empty
// strings when the label is a monster attribute instead.
func rushDuelSpellTrapType(s string) (CardType, CardSubtype) {
	switch s {
	case "通常魔法", "일반 마법":
		return Spell, SpellNormal
	case "速攻魔法", "속공 마법":
		return Spell, SpellQuickPlay
	case "フィールド魔法", "필드 마법":
		return Spell, SpellField
	case "装備魔法", "장착 마법":
		return Spell, SpellEquip
	case "永続魔法", "지속 마법":
		return Spell, SpellContinuous
	case "リチュアル魔法", "리추얼 마법":
		return Spell, SpellRitual
	case "通常罠", "일반 함정":
		return Trap, TrapNormal
	case "永続罠", "지속 함정":
		return Trap, TrapContinuous
	case "カウンター罠", "카운터 함정":
		return Trap, TrapCounter
	default:
		return "", ""
	}
}

// rushDuelMonsterAttribute translates a Rush Duel JA or KO attribute label to
// the canonical MonsterAttribute. Unrecognized labels are passed through as-is.
func rushDuelMonsterAttribute(s string) MonsterAttribute {
	switch s {
	case "光属性", "빛":
		return LIGHT
	case "闇属性", "어둠":
		return DARK
	case "地属性", "땅":
		return EARTH
	case "炎属性", "화염":
		return FIRE
	case "水属性", "물":
		return WATER
	case "風属性", "바람":
		return WIND
	case "神属性", "신":
		return DIVINE
	default:
		return MonsterAttribute(s)
	}
}

// rushDuelMonsterTypeJP translates a Rush Duel JA monster type name to the
// canonical MonsterType, or empty string for unrecognized tokens.
// Rush Duel uses different kanji for some types (e.g. "竜族" and "神獣族")
// that do not appear in the standard Konami DB.
func rushDuelMonsterTypeJP(s string) MonsterType {
	switch s {
	case "悪魔族":
		return Fiend
	case "アンデット族":
		return Zombie
	case "機械族":
		return Machine
	case "魔法使い族":
		return Spellcaster
	case "戦士族":
		return Warrior
	case "竜族", "ドラゴン族":
		return Dragon
	case "天使族":
		return Fairy
	case "海竜族":
		return SeaSerpent
	case "昆虫族":
		return Insect
	case "岩石族":
		return Rock
	case "獣族":
		return Beast
	case "獣戦士族":
		return BeastWarrior
	case "植物族":
		return Plant
	case "炎族":
		return Pyro
	case "雷族":
		return Thunder
	case "爬虫類族":
		return Reptile
	case "魚族":
		return Fish
	case "サイキック族":
		return Psychic
	case "水族":
		return Aqua
	case "恐竜族":
		return Dinosaur
	case "鳥獣族":
		return WingedBeast
	case "神獣族":
		return DivineBeast
	case "創造神族":
		return CreatorGod
	case "サイバース族":
		return Cyberse
	case "幻想魔族":
		return Illusion
	case "幻竜族":
		return Wyrm
	case "魔導騎士族":
		return MagicalKnight
	case "サイボーグ族":
		return Cyborg
	case "ハイドラゴン族":
		return HighDragon
	case "天界戦士族":
		return CelestialWarrior
	case "オメガサイキック族":
		return OmegaPsychic
	case "ギャラクシー族":
		return Galaxy
	default:
		return ""
	}
}

// ParseRushDuelCardHTML parses a Rush Duel card detail page from the Konami
// Rush Duel card database (db.yugioh-card.com/rushdb) and returns a
// CardRushDuel. Convenience wrapper; for the bulk crawl, use Parser to share
// one parsed DOM tree across LocaleText / Prints / Card / RushCard.
func ParseRushDuelCardHTML(cardPageHTML []byte, cardID CardID) CardRushDuel {
	return NewParser(cardPageHTML, cardID).RushCard()
}

func parseRushDuelCardFromNode(root *html.Node, cardID CardID) CardRushDuel {
	c := CardRushDuel{Card: Card{MiscKonamiCardID: cardID}}

	// getNode returns the first matched node or an empty node on no match
	getNode := func(parent *html.Node, xpath string) *html.Node {
		nodes, _ := textproc.HTMLXPath(parent, xpath)
		if len(nodes) == 0 {
			return &html.Node{}
		}
		return nodes[0]
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
	cardEffect := htmlGetCardText(cardTexts[1])
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
		if ct, cs := rushDuelSpellTrapType(attributeS); ct != "" {
			c.CardType = ct
			c.CardSubtype = cs
		} else {
			c.CardType = Monster
			c.MonsterAttribute = rushDuelMonsterAttribute(attributeS)
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
				if isNumberChar(r) {
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
			if err != nil && !isSpecialATKDEF(atkS) {
				log.Printf("error rushDuel cardID %v MonsterATK: %q\n", cardID, atkS)
			}
		}
		if len(values) > atkIdx+1 {
			defS := strings.TrimSpace(textproc.HTMLGetText(values[atkIdx+1]))
			c.MonsterDEFStr = defS
			c.MonsterDEF, err = strconv.ParseFloat(defS, 64)
			if err != nil && !isSpecialATKDEF(defS) {
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

		// English: "Dragon-Type" → strip "-Type" → lookup in toMonsterType
		cleanType := strings.TrimSuffix(token, "-Type")
		if mt := toMonsterType(cleanType); mt != "" {
			card.MonsterType = mt
			continue
		}
		// English without suffix: "Winged Beast", "Dragon", etc.
		if mt := toMonsterType(token); mt != "" {
			card.MonsterType = mt
			continue
		}
		// Japanese
		if mt := rushDuelMonsterTypeJP(token); mt != "" {
			card.MonsterType = mt
			continue
		}
	}

	slices.Sort(card.MonsterAbilities)
}
