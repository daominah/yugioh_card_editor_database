package core

import (
	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
	"log"
	"sort"
	"strconv"
	"strings"
)

// MapCardType keys are strings on Konami website
var MapCardType = map[string]CardType{
	"Normal Spell":     Spell,
	"Quick-Play Spell": Spell,
	"Ritual Spell":     Spell,
	"Continuous Spell": Spell,
	"Field Spell":      Spell,
	"Equip Spell":      Spell,
	"Normal Trap":      Trap,
	"Counter Trap":     Trap,
	"Continuous Trap":  Trap,
}

// MapCardSubtype keys are strings on Konami website
var MapCardSubtype = map[string]CardSubtype{
	"Normal Spell":     SpellNormal,
	"Quick-Play Spell": SpellQuickPlay,
	"Ritual Spell":     SpellRitual,
	"Continuous Spell": SpellContinuous,
	"Field Spell":      SpellField,
	"Equip Spell":      SpellEquip,
	"Normal Trap":      TrapNormal,
	"Counter Trap":     TrapCounter,
	"Continuous Trap":  TrapContinuous,
}

var MapCardSubtypeMonsterExtra = map[string]CardSubtype{
	"Ritual":  MonsterRitual,
	"Fusion":  MonsterFusion,
	"Synchro": MonsterSynchro,
	"Xyz":     MonsterXyz,
	"Link":    MonsterLink,
}

var MapCardSubtypeMonsterEffect = map[string]CardSubtype{
	"Normal": MonsterNormal,
	"Effect": MonsterEffect,
}

var MapMonsterTypes = map[string]MonsterType{
	"Aqua":          Aqua,
	"Beast":         Beast,
	"Beast-Warrior": BeastWarrior,
	"Creator God":   CreatorGod,
	"Cyberse":       Cyberse,
	"Dinosaur":      Dinosaur,
	"Divine-Beast":  DivineBeast,
	"Dragon":        Dragon,
	"Fairy":         Fairy,
	"Fiend":         Fiend,
	"Fish":          Fish,
	"Illusion Type": Illusion,
	"Insect":        Insect,
	"Machine":       Machine,
	"Plant":         Plant,
	"Psychic":       Psychic,
	"Pyro":          Pyro,
	"Reptile":       Reptile,
	"Rock":          Rock,
	"Sea Serpent":   SeaSerpent,
	"Spellcaster":   Spellcaster,
	"Thunder":       Thunder,
	"Warrior":       Warrior,
	"Winged Beast":  WingedBeast,
	"Wyrm":          Wyrm,
	"Zombie":        Zombie,
}

var MapMonsterAttributes = map[string]MonsterAttribute{
	"DARK": DARK, "EARTH": EARTH, "FIRE": FIRE,
	"LIGHT": LIGHT, "WATER": WATER, "WIND": WIND, "DIVINE": DIVINE,
}

var NumberChars = map[rune]bool{
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
}

var MapLinkArrow = map[rune]MonsterLinkArrow{
	'7': UpLeft, '8': Up, '9': UpRight,
	'4': Left, '6': Right,
	'1': DownLeft, '2': Down, '3': DownRight,
}

var linkArrowOrder = map[MonsterLinkArrow]int{
	UpLeft: 7, Up: 8, UpRight: 9,
	Left: 4, Right: 6,
	DownLeft: 1, Down: 2, DownRight: 3,
}

var SpecialATKDEF = map[string]bool{"?": true, "-": true}

func ParseKonamiCardHTML(cardPageHTML []byte, cardID string) Card {
	c := Card{MiscKonamiCardID: cardID}
	root := textproc.HTMLParseToNode(cardPageHTML)

	// getNode always return one non-empty node, ignores error
	getNode := func(parent *html.Node, xpath string) *html.Node {
		node, _ := textproc.HTMLXPath(parent, xpath)
		if len(node) == 0 {
			return &html.Node{}
		}
		return node[0]
	}

	// data of cards related to the parsing card, unused so remove
	relationNode := getNode(root, `//*[@id="relationCard"]`)
	if relationNode.Parent != nil {
		relationNode.Parent.RemoveChild(relationNode)
	}

	c.CardName = textproc.HTMLGetText(getNode(root, `//*[@id="cardname"]`))

	cardTexts, err := textproc.HTMLXPath(root, `//*[@class="CardText"]`)
	if len(cardTexts) < 2 {
		log.Printf("error cardID %v cardTexts len: %v, %v\n", cardID, len(cardTexts), err)
		return c
	}
	cardText1 := textproc.HTMLGetText(cardTexts[1])
	cardText1 = strings.TrimSpace(cardText1)
	cardText1 = strings.TrimPrefix(cardText1, "Card Text")
	cardText1 = strings.TrimSpace(cardText1)
	cardText1 = strings.ReplaceAll(cardText1, "��", "● ") // workaround Konami text bug
	c.CardEffect = cardText1

	es, _ := textproc.HTMLXPath(root, `//span[contains(@class,"item_box")]`)
	if len(es) == 2 { // Spell or Trap
		konamiCardSubtype := textproc.HTMLGetText(es[1])
		found := false
		c.CardSubtype, found = MapCardSubtype[konamiCardSubtype]
		if !found {
			log.Printf("error cardID %v CardSubtype: %v\n", cardID, konamiCardSubtype)
		}
		c.CardType, found = MapCardType[konamiCardSubtype]
		if !found {
			log.Printf("error cardID %v CardType: %v\n", cardID, konamiCardSubtype)
		}
	} else {
		c.CardType = Monster
		monsterAbilitiesS := textproc.HTMLGetText(
			getNode(root, `//*[@class="species"]`))
		var monsterAbilities []string
		for _, v := range strings.Split(monsterAbilitiesS, "/") {
			monsterAbilities = append(monsterAbilities, strings.TrimSpace(v))
		}
		c.IsNonEffectMonster = true
		for _, v := range monsterAbilities {
			isAbility := true // not count CardSubtype or MonsterType or Normal or Effect
			if cardSubtype, ok := MapCardSubtypeMonsterExtra[v]; ok {
				c.CardSubtype = cardSubtype
				isAbility = false
			}
			if c.CardSubtype == "" {
				if cardSubtype, ok := MapCardSubtypeMonsterEffect[v]; ok {
					c.CardSubtype = cardSubtype
					isAbility = false
				}
			}
			if monsterType, ok := MapMonsterTypes[v]; ok {
				c.MonsterType = monsterType
				isAbility = false
			}
			if isAbility && v != "Normal" && v != "Effect" && v != "Pendulum" {
				c.MonsterAbilities = append(c.MonsterAbilities, MonsterAbility(v))
			}
			if v == "Effect" {
				c.IsNonEffectMonster = false
			}
		}
		if c.CardSubtype == "" {
			log.Printf("error cardID %v CardSubtype: %v\n", cardID, monsterAbilitiesS)
		}
		if c.MonsterType == "" {
			log.Printf("error cardID %v MonsterType: %v\n", cardID, monsterAbilitiesS)
		}
		sort.Slice(c.MonsterAbilities, func(i, j int) bool {
			return c.MonsterAbilities[i] < c.MonsterAbilities[j]
		})

		if c.CardSubtype == MonsterLink {
			linkNode := getNode(root, `//*[@alt="Link"]`)
			//log.Printf("linkNode.Attr: %+v\n", linkNode.Attr)
			for _, attr := range linkNode.Attr {
				if attr.Key != "class" {
					continue
				}
				//log.Printf("linkNode.Attr.class: %#v\n", attr.Val)
				for _, char := range []rune(attr.Val) {
					ar, found := MapLinkArrow[char]
					if found {
						c.MonsterLinkArrows = append(c.MonsterLinkArrows, ar)
					}
				}
			}
			sort.Slice(c.MonsterLinkArrows, func(i, j int) bool {
				return linkArrowOrder[c.MonsterLinkArrows[i]] < linkArrowOrder[c.MonsterLinkArrows[j]]
			})
		}

		values, _ := textproc.HTMLXPath(root, `//span[@class="item_box_value"]`)
		if len(values) >= 4 {
			attributeS := textproc.HTMLGetText(values[0])
			found := false
			c.MonsterAttribute, found = MapMonsterAttributes[attributeS]
			if !found {
				log.Printf("error cardID %v MonsterAttribute: %v\n", cardID, attributeS)
			}
			levelRankLinkS := textproc.HTMLGetText(values[1])
			var levelRankLinkA []rune
			for _, char := range []rune(levelRankLinkS) {
				if NumberChars[char] {
					levelRankLinkA = append(levelRankLinkA, char)
				}
			}
			c.MonsterLevelRankLink, err = strconv.Atoi(string(levelRankLinkA))
			if err != nil {
				log.Printf("error cardID %v MonsterLevelRankLink: %v\n", cardID, levelRankLinkS)
			}
			c.MonsterATKStr = textproc.HTMLGetText(values[2])
			c.MonsterATK, err = strconv.ParseFloat(c.MonsterATKStr, 64)
			if err != nil {
				if !SpecialATKDEF[c.MonsterATKStr] {
					log.Printf("error cardID %v MonsterATK: %v\n", cardID, c.MonsterATKStr)
				}
			}
			c.MonsterDEFStr = textproc.HTMLGetText(values[3])
			c.MonsterDEF, err = strconv.ParseFloat(c.MonsterDEFStr, 64)
			if err != nil {
				if !SpecialATKDEF[c.MonsterDEFStr] {
					log.Printf("error cardID %v MonsterDEF: %v\n", cardID, c.MonsterDEFStr)
				}
			}
			if len(values) >= 5 {
				c.IsPendulum = true
				penScaleS := textproc.HTMLGetText(values[4])
				c.PendulumScale, err = strconv.Atoi(penScaleS)
				if err != nil {
					log.Printf("error cardID %v PendulumScale: %v\n", cardID, penScaleS)
				}
				c.PendulumEffect = textproc.HTMLGetText(
					getNode(root, `//div[contains(@class,"pen_effect")]`))
			}
		}
	}

	konamiSets, err := textproc.HTMLXPath(root, `//*[@id="update_list"]//*[@class="t_row"]`)
	for i := len(konamiSets) - 1; i >= 0; i-- {
		firstSet := konamiSets[i]
		//log.Printf("debug cardID %v firstSet: %v\n", cardID, textproc.HTMLGetText(firstSet))
		ymd := textproc.HTMLGetText(getNode(firstSet, `//*[@class="time"]`))
		if len(ymd) >= 4 {
			c.MiscYear = ymd[:4]
		}
		c.MiscKonamiSet = textproc.HTMLGetText(getNode(firstSet, `//*[@class="card_number"]`))
		if c.MiscYear != "" && c.MiscKonamiSet != "" {
			break
		}
	}

	return c
}
