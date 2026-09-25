package konami

import (
	"log"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
)

// toCardType translates a localized spell/trap label to the canonical CardType.
// Unrecognized labels are passed through as-is, preserving new card types from Konami.
func toCardType(s string) CardType {
	switch s {
	case "Normal Spell", "Quick-Play Spell", "Ritual Spell",
		"Continuous Spell", "Field Spell", "Equip Spell",
		"通常魔法", "速攻魔法", "儀式魔法", "永続魔法", "フィールド魔法", "装備魔法",
		"일반 마법", "속공 마법", "의식 마법", "지속 마법", "필드 마법", "장착 마법":
		return Spell
	case "Normal Trap", "Counter Trap", "Continuous Trap",
		"通常罠", "カウンター罠", "永続罠",
		"일반 함정", "카운터 함정", "지속 함정":
		return Trap
	default:
		return CardType(s)
	}
}

// toCardSubtype translates a localized spell/trap label to the canonical CardSubtype.
// Unrecognized labels are passed through as-is.
// Each case groups the EN, JA, and KO spellings for the same subtype.
func toCardSubtype(s string) CardSubtype {
	switch s {
	case "Normal Spell", "通常魔法", "일반 마법":
		return SpellNormal
	case "Quick-Play Spell", "速攻魔法", "속공 마법":
		return SpellQuickPlay
	case "Ritual Spell", "儀式魔法", "의식 마법":
		return SpellRitual
	case "Continuous Spell", "永続魔法", "지속 마법":
		return SpellContinuous
	case "Field Spell", "フィールド魔法", "필드 마법":
		return SpellField
	case "Equip Spell", "装備魔法", "장착 마법":
		return SpellEquip
	case "Normal Trap", "通常罠", "일반 함정":
		return TrapNormal
	case "Counter Trap", "カウンター罠", "카운터 함정":
		return TrapCounter
	case "Continuous Trap", "永続罠", "지속 함정":
		return TrapContinuous
	default:
		return CardSubtype(s)
	}
}

// toMonsterExtraSubtype returns the extra-deck CardSubtype for recognized
// extra-deck keywords, or empty string for all other species tokens.
func toMonsterExtraSubtype(s string) CardSubtype {
	switch s {
	case "Ritual", "儀式", "의식":
		return MonsterRitual
	case "Fusion", "融合", "융합":
		return MonsterFusion
	case "Synchro", "シンクロ", "싱크로":
		return MonsterSynchro
	case "Xyz", "エクシーズ", "엑시즈":
		return MonsterXyz
	case "Link", "リンク", "링크":
		return MonsterLink
	default:
		return ""
	}
}

// toMonsterEffectSubtype returns MonsterNormal or MonsterEffect for the
// recognized Normal/Effect keywords, or empty string for all other tokens.
func toMonsterEffectSubtype(s string) CardSubtype {
	switch s {
	case "Normal", "通常", "일반":
		return MonsterNormal
	case "Effect", "効果", "효과":
		return MonsterEffect
	default:
		return ""
	}
}

// toMonsterType translates a localized monster type string to the canonical
// MonsterType. Returns empty string for non-type tokens (e.g. "Effect", "Tuner").
func toMonsterType(s string) MonsterType {
	switch s {
	case "Aqua", "水族", "물족":
		return Aqua
	case "Beast", "獣族", "야수족":
		return Beast
	case "Beast-Warrior", "獣戦士族", "야수전사족":
		return BeastWarrior
	case "Creator God", "創造神族", "창조신족":
		return CreatorGod
	case "Cyberse", "サイバース族", "사이버스족":
		return Cyberse
	case "Dinosaur", "恐竜族", "공룡족":
		return Dinosaur
	case "Divine-Beast", "幻神獣族", "환신야수족":
		return DivineBeast
	case "Dragon", "ドラゴン族", "드래곤족":
		return Dragon
	case "Fairy", "天使族", "천사족":
		return Fairy
	case "Fiend", "悪魔族", "악마족":
		return Fiend
	case "Fish", "魚族", "어류족":
		return Fish
	case "Illusion", "Illusion Type", "幻想魔族", "환상마족":
		return Illusion
	case "Insect", "昆虫族", "곤충족":
		return Insect
	case "Machine", "機械族", "기계족":
		return Machine
	case "Plant", "植物族", "식물족":
		return Plant
	case "Psychic", "サイキック族", "사이킥족":
		return Psychic
	case "Pyro", "炎族", "화염족":
		return Pyro
	case "Reptile", "爬虫類族", "파충류족":
		return Reptile
	case "Rock", "岩石族", "암석족":
		return Rock
	case "Sea Serpent", "海竜族", "해룡족":
		return SeaSerpent
	case "Spellcaster", "魔法使い族", "마법사족":
		return Spellcaster
	case "Thunder", "雷族", "번개족":
		return Thunder
	case "Warrior", "戦士族", "전사족":
		return Warrior
	case "Winged Beast", "鳥獣族", "비행야수족":
		return WingedBeast
	case "Wyrm", "幻竜族", "환룡족":
		return Wyrm
	case "Zombie", "アンデット族", "언데드족":
		return Zombie
	default:
		return ""
	}
}

// toMonsterAttribute translates a localized attribute string to the canonical
// MonsterAttribute. Unrecognized attributes are passed through as-is.
func toMonsterAttribute(s string) MonsterAttribute {
	switch s {
	case "DARK", "闇属性", "어둠":
		return DARK
	case "EARTH", "地属性", "땅":
		return EARTH
	case "FIRE", "炎属性", "화염":
		return FIRE
	case "LIGHT", "光属性", "빛":
		return LIGHT
	case "WATER", "水属性", "물":
		return WATER
	case "WIND", "風属性", "바람":
		return WIND
	case "DIVINE", "神属性", "신":
		return DIVINE
	default:
		return MonsterAttribute(s)
	}
}

// toMonsterAbility returns the canonical MonsterAbility for recognized ability
// keywords, or empty string for non-ability species tokens.
func toMonsterAbility(s string) MonsterAbility {
	switch s {
	case "Flip", "リバース", "리버스":
		return Flip
	case "Gemini", "デュアル", "듀얼":
		return Gemini
	case "Spirit", "スピリット", "스피릿":
		return Spirit
	case "Toon", "トゥーン", "툰":
		return Toon
	case "Tuner", "チューナー", "튜너":
		return Tuner
	case "Union", "ユニオン", "유니온":
		return Union
	default:
		return ""
	}
}

// isSkipAbilityKeyword reports whether the species token is handled as
// CardSubtype or IsPendulum elsewhere and should not be treated as a MonsterAbility.
func isSkipAbilityKeyword(s string) bool {
	switch s {
	case "Normal", "Effect", "Pendulum",
		"通常", "効果", "ペンデュラム",
		"일반", "효과", "펜듈럼":
		return true
	default:
		return false
	}
}

// isSpecialSummonOnlyKeyword reports whether the species token marks the
// monster as Special Summon-only ("Nomi" / "Semi-Nomi"): cannot be Normal
// Summoned or Set, can only be Special Summoned via card-text conditions.
// Konami's EN page omits this token from the species line, so only JA and KO
// values are listed.
func isSpecialSummonOnlyKeyword(s string) bool {
	switch s {
	case "特殊召喚", "특수 소환":
		return true
	default:
		return false
	}
}

// isEffectKeyword reports whether the species token marks the monster as an
// effect monster (clears the IsNonEffectMonster flag).
func isEffectKeyword(s string) bool {
	switch s {
	case "Effect", "効果", "효과":
		return true
	default:
		return false
	}
}

// isNumberChar reports whether r is an ASCII digit.
func isNumberChar(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

// isSpecialATKDEF reports whether the ATK/DEF string is a non-numeric sentinel
// value printed on some cards: "?" for variable, "-" for N/A.
func isSpecialATKDEF(s string) bool {
	switch s {
	case "?", "-":
		return true
	default:
		return false
	}
}

// linkArrowFromChar returns the MonsterLinkArrow for a numpad-style character
// in a link monster's CSS class string, or empty string for non-arrow characters.
// The numpad layout: 7=UpLeft 8=Up 9=UpRight / 4=Left 6=Right / 1=DownLeft 2=Down 3=DownRight.
func linkArrowFromChar(r rune) MonsterLinkArrow {
	switch r {
	case '7':
		return UpLeft
	case '8':
		return Up
	case '9':
		return UpRight
	case '4':
		return Left
	case '6':
		return Right
	case '1':
		return DownLeft
	case '2':
		return Down
	case '3':
		return DownRight
	default:
		return ""
	}
}

// linkArrowSortOrder returns the display-order position for a link arrow,
// matching the numpad layout used in the CSS class encoding.
func linkArrowSortOrder(arrow MonsterLinkArrow) int {
	switch arrow {
	case UpLeft:
		return 7
	case Up:
		return 8
	case UpRight:
		return 9
	case Left:
		return 4
	case Right:
		return 6
	case DownLeft:
		return 1
	case Down:
		return 2
	case DownRight:
		return 3
	default:
		return 0
	}
}

// ParseKonamiCardHTML parses a Konami yugiohdb card detail page and returns
// a Card. Convenience wrapper; for the bulk crawl, use Parser to share one
// parsed DOM tree across LocaleText / Prints / Card / RushCard.
func ParseKonamiCardHTML(cardPageHTML []byte, cardID CardID) Card {
	return NewParser(cardPageHTML, cardID).Card()
}

// htmlGetCardText is textproc.HTMLGetText for card text (effect, Pendulum effect),
// with two differences:
//
//   - The effect prefix (circled number with its colon, JA "①：", KO "①:")
//     stays as written on the Konami page, like on the printed card,
//     because Unicode normalization (NFKC) would turn it into "1:".
//     The rest of the text is still normalized, so searching "700" matches JA "７００".
//   - Since 2026-09, Konami pages write line breaks in card text as escaped "&lt;br&gt;",
//     which decodes to a literal "<br>" text instead of a <br> element,
//     so it is converted to a newline here.
//     Older pages with a <br> element become a newline like in HTMLGetText.
func htmlGetCardText(node *html.Node) string {
	var buf strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			isExcluded := n.Parent != nil && (n.Parent.Data == "script" || n.Parent.Data == "style")
			if !isExcluded {
				buf.WriteString(n.Data)
				buf.WriteString("\n")
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	text := regexpLiteralBr.ReplaceAllString(buf.String(), "\n")
	text = textproc.RemoveRedundantSpace(strings.TrimSpace(text))
	return normalizeExceptEffectPrefix(text)
}

// normalizeExceptEffectPrefix applies textproc.NormalizeText to the text
// between effect prefixes, keeping each prefix unchanged.
func normalizeExceptEffectPrefix(text string) string {
	var buf strings.Builder
	last := 0
	for _, loc := range regexpEffectPrefix.FindAllStringIndex(text, -1) {
		buf.WriteString(textproc.NormalizeText(text[last:loc[0]]))
		buf.WriteString(text[loc[0]:loc[1]])
		last = loc[1]
	}
	buf.WriteString(textproc.NormalizeText(text[last:]))
	return buf.String()
}

// regexpEffectPrefix matches a circled number ① to ⑳ with its optional colon,
// full-width "：" on JA pages or ":" on KO pages.
var regexpEffectPrefix = regexp.MustCompile(`[①-⑳][：:]?`)

var regexpLiteralBr = regexp.MustCompile(`(?i)<br\s*/?>`)

func parseKonamiCardFromNode(root *html.Node, cardID CardID) Card {
	c := Card{MiscKonamiCardID: cardID}

	// getNode always return one non-empty node, ignores error
	getNode := func(parent *html.Node, xpath string) *html.Node {
		node, _ := textproc.HTMLXPath(parent, xpath)
		if len(node) == 0 {
			return &html.Node{}
		}
		return node[0]
	}

	// Extract names from the h1 inside #cardname. parseCardNameAndPronunciation
	// returns three parts (see its docstring): localized printed name from
	// text nodes, kana pronunciation from <span class="ruby">, and the EN
	// name from the bare <span>. Only the JA crawl pass actually persists
	// CardNameEN to cards.card_name_en (gated in the crawler); for EN/KO
	// pages this field is parsed but the crawler does not call UpsertCard.
	h1NameNode := getNode(root, `//*[@id="cardname"]//h1`)
	if h1NameNode.Parent != nil {
		c.CardName, _, c.CardNameEN = parseCardNameAndPronunciation(h1NameNode)
	}

	cardTexts, err := textproc.HTMLXPath(root, `//*[@class="CardText"]`)
	if len(cardTexts) < 2 {
		log.Printf("error cardID %v cardTexts len: %v, %v\n", cardID, len(cardTexts), err)
		return c
	}
	cardText1 := htmlGetCardText(cardTexts[1])
	cardText1 = strings.TrimSpace(cardText1)
	cardText1 = strings.TrimPrefix(cardText1, "Card Text")
	cardText1 = strings.TrimPrefix(cardText1, "カードテキスト")
	cardText1 = strings.TrimPrefix(cardText1, "카드 텍스트")
	cardText1 = strings.TrimSpace(cardText1)
	cardText1 = strings.ReplaceAll(cardText1, "��", "● ") // workaround Konami text bug
	// replace "●" immediately followed by normal character with an added space "● "
	cardText1 = regexp.MustCompile(`(?m)^●([^\s])`).ReplaceAllString(cardText1, "● $1")
	c.CardEffect = cardText1

	es, _ := textproc.HTMLXPath(root, `//div[@class="frame"]//span[contains(@class,"item_box")]`)
	if len(es) == 2 { // Spell or Trap
		konamiCardSubtype := textproc.HTMLGetText(es[1])
		c.CardSubtype = toCardSubtype(konamiCardSubtype)
		c.CardType = toCardType(konamiCardSubtype)
	} else {
		c.CardType = Monster
		monsterAbilitiesS := textproc.HTMLGetText(
			getNode(root, `//*[@class="species"]`))
		var monsterAbilities []string
		for v := range strings.SplitSeq(monsterAbilitiesS, "/") {
			monsterAbilities = append(monsterAbilities, strings.TrimSpace(v))
		}
		c.IsNonEffectMonster = true
		for _, v := range monsterAbilities {
			matched := false
			if sub := toMonsterExtraSubtype(v); sub != "" {
				c.CardSubtype = sub
				matched = true
			}
			if c.CardSubtype == "" {
				if sub := toMonsterEffectSubtype(v); sub != "" {
					c.CardSubtype = sub
					matched = true
				}
			}
			if mt := toMonsterType(v); mt != "" {
				c.MonsterType = mt
				matched = true
			}
			if ab := toMonsterAbility(v); ab != "" {
				c.MonsterAbilities = append(c.MonsterAbilities, ab)
				matched = true
			}
			if isSkipAbilityKeyword(v) {
				matched = true
			}
			if isSpecialSummonOnlyKeyword(v) {
				c.IsSpecialSummonOnly = true
				matched = true
			}
			if !matched && v != "" {
				log.Printf("warn cardID %v unrecognized species part: %q\n", cardID, v)
			}
			if isEffectKeyword(v) {
				c.IsNonEffectMonster = false
			}
		}
		if c.CardSubtype == "" {
			log.Printf("error cardID %v CardSubtype: %v\n", cardID, monsterAbilitiesS)
		}
		if c.MonsterType == "" {
			log.Printf("error cardID %v MonsterType: %v\n", cardID, monsterAbilitiesS)
		}
		slices.Sort(c.MonsterAbilities)

		if c.CardSubtype == MonsterLink {
			// Match by class prefix instead of alt: alt is localized
			// ("Link" / "リンク" / "링크"), the "icon_img_set link<digits>"
			// class string is identical across locale pages.
			linkNode := getNode(root, `//*[starts-with(@class, "icon_img_set link")]`)
			// log.Printf("linkNode.Attr: %+v\n", linkNode.Attr)
			for _, attr := range linkNode.Attr {
				if attr.Key != "class" {
					continue
				}
				// log.Printf("linkNode.Attr.class: %#v\n", attr.Val)
				for _, char := range attr.Val {
					if ar := linkArrowFromChar(char); ar != "" {
						c.MonsterLinkArrows = append(c.MonsterLinkArrows, ar)
					}
				}
			}
			sort.Slice(c.MonsterLinkArrows, func(i, j int) bool {
				return linkArrowSortOrder(c.MonsterLinkArrows[i]) < linkArrowSortOrder(c.MonsterLinkArrows[j])
			})
		}

		values, _ := textproc.HTMLXPath(root, `//span[@class="item_box_value"]`)
		if len(values) >= 4 {
			attributeS := textproc.HTMLGetText(values[0])
			c.MonsterAttribute = toMonsterAttribute(attributeS)
			levelRankLinkS := textproc.HTMLGetText(values[1])
			var levelRankLinkA []rune
			for _, char := range levelRankLinkS {
				if isNumberChar(char) {
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
				if !isSpecialATKDEF(c.MonsterATKStr) {
					log.Printf("error cardID %v MonsterATK: %v\n", cardID, c.MonsterATKStr)
				}
			}
			c.MonsterDEFStr = textproc.HTMLGetText(values[3])
			c.MonsterDEF, err = strconv.ParseFloat(c.MonsterDEFStr, 64)
			if err != nil {
				if !isSpecialATKDEF(c.MonsterDEFStr) {
					log.Printf("error cardID %v MonsterDEF: %v\n", cardID, c.MonsterDEFStr)
				}
			}
			if len(values) >= 5 {
				c.IsPendulum = true
				penScaleS := textproc.HTMLGetText(values[4])
				var penScaleDigits []rune
				for _, ch := range penScaleS {
					if isNumberChar(ch) {
						penScaleDigits = append(penScaleDigits, ch)
					}
				}
				c.PendulumScale, err = strconv.Atoi(string(penScaleDigits))
				if err != nil {
					log.Printf("error cardID %v PendulumScale: %v\n", cardID, penScaleS)
				}
				c.PendulumEffect = htmlGetCardText(
					getNode(root, `//div[contains(@class,"pen_effect")]`))
			}
		}
	}

	konamiSets, err := textproc.HTMLXPath(root, `//*[@id="update_list"]//*[starts-with(@class,"t_row")]`)
	if err != nil {
		log.Printf("error cardID %v konamiSets: %v\n", cardID, err)
	}
	for i := len(konamiSets) - 1; i >= 0; i-- {
		firstSet := konamiSets[i]
		// log.Printf("debug cardID %v firstSet: %v\n", cardID, textproc.HTMLGetText(firstSet))
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

// CheckEqualArray can check if MonsterAbilities or MonsterLinkArrows are equal
func CheckEqualArray[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
