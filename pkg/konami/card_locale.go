package konami

import (
	"slices"
	"strings"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
)

// CardPrint is one printed appearance of a card in a specific regional set.
type CardPrint struct {
	Date       string // "YYYY-MM-DD" as scraped from the Konami card page
	Position   string // full card number, e.g. "LOB-001", "LOCH-JP077"
	SetName    string // e.g. "Legend of Blue-Eyes White Dragon"
	RarityCode string // short code, e.g. "UL", "SE"
	RarityName string // full name, e.g. "Ultimate Rare"
}

// CardLocaleText holds locale-specific text for one card in one language.
type CardLocaleText struct {
	// Name is the official printed name: kanji for JA (e.g. 青眼の白龍),
	// Hangul for KO, English for EN.
	Name string
	// NamePronunciation is the phonetic kana reading taken from the
	// <span class="ruby"> child of <h1> on JA pages. The previous parser
	// also lumped the bare <span> (English name) into this field; the
	// current parser narrows it to the ruby span only, so this field is
	// kana-only. Empty for EN and KO pages.
	NamePronunciation string
	// NameEnglishOnPage is the EN name as printed on the page's <h1> bare
	// <span> (no class). JA and KO pages carry it; EN pages do not (the
	// page is already EN). Always empty for Rush JA pages (no EN span).
	// Used by the crawler to populate cards.card_name_en during the JA pass.
	NameEnglishOnPage string
	Effect            string
	PendulumEffect    string
	// AttributeText is the localized attribute display: "LIGHT" / "光属性" / "빛".
	// Empty for Spell and Trap cards.
	AttributeText string
	// MonsterTypeText is the localized type display: "Dragon" / "ドラゴン族" /
	// "드래곤족" / Rush-only kanji like "竜族". Sourced verbatim from the first
	// slash-separated token of <p class="species">, but only when the page
	// is a monster page (detected via len(item_box_value) >= 4, the
	// ATK/DEF/level/attribute slots). Empty for spells, traps, and any
	// non-monster row.
	MonsterTypeText string
}

// ParseCardPrints extracts all prints from the #update_list section of a
// Konami card page (yugiohdb or rushdb). One t_row per regional set print.
func ParseCardPrints(cardPageHTML []byte, cardID CardID) []CardPrint {
	root := textproc.HTMLParseToNode(cardPageHTML)
	getNode := func(parent *html.Node, xpath string) *html.Node {
		nodes, _ := textproc.HTMLXPath(parent, xpath)
		if len(nodes) == 0 {
			return &html.Node{}
		}
		return nodes[0]
	}

	rows, _ := textproc.HTMLXPath(root,
		`//*[@id="update_list"]//*[starts-with(@class,"t_row")]`)
	prints := make([]CardPrint, 0, len(rows))
	for _, row := range rows {
		position := strings.TrimSpace(
			textproc.HTMLGetText(getNode(row, `//*[@class="card_number"]`)))
		if position == "" {
			continue
		}
		date := strings.TrimSpace(
			textproc.HTMLGetText(getNode(row, `//*[@class="time"]`)))
		setName := strings.TrimSpace(
			textproc.HTMLGetText(getNode(row, `.//*[contains(@class,"pack_name")]`)))
		rarityCode := strings.TrimSpace(
			textproc.HTMLGetText(getNode(row, `.//*[contains(@class,"lr_icon")]//p`)))
		rarityName := strings.TrimSpace(
			textproc.HTMLGetText(getNode(row, `.//*[contains(@class,"lr_icon")]//span`)))
		prints = append(prints, CardPrint{
			Date:       date,
			Position:   position,
			SetName:    setName,
			RarityCode: rarityCode,
			RarityName: rarityName,
		})
	}
	return prints
}

// ParseCardLocaleText extracts locale-specific name and effect text from a
// Konami card page. For JA pages, NamePronunciation holds the katakana from
// the ruby span and Name holds the printed kanji. For EN/KO pages
// NamePronunciation is empty.
func ParseCardLocaleText(cardPageHTML []byte, cardID CardID) CardLocaleText {
	root := textproc.HTMLParseToNode(cardPageHTML)
	getNode := func(parent *html.Node, xpath string) *html.Node {
		nodes, _ := textproc.HTMLXPath(parent, xpath)
		if len(nodes) == 0 {
			return &html.Node{}
		}
		return nodes[0]
	}

	var t CardLocaleText

	// Name: text-nodes only from h1#cardname, skipping child element nodes
	// (the ruby span on JA pages holds the katakana pronunciation, not the name).
	h1 := getNode(root, `//*[@id="cardname"]//h1`)
	if h1.Parent != nil {
		t.Name, t.NamePronunciation, t.NameEnglishOnPage = parseCardNameAndPronunciation(h1)
	}

	// Effect text
	cardTexts, _ := textproc.HTMLXPath(root, `//*[@class="CardText"]`)
	if len(cardTexts) >= 2 {
		effect := strings.TrimSpace(textproc.HTMLGetText(cardTexts[1]))
		effect = strings.TrimPrefix(effect, "Card Text")
		effect = strings.TrimPrefix(effect, "カードテキスト")
		effect = strings.TrimPrefix(effect, "카드 텍스트")
		t.Effect = strings.TrimSpace(effect)
	}

	// Pendulum effect (empty string for non-pendulum cards)
	t.PendulumEffect = strings.TrimSpace(
		textproc.HTMLGetText(getNode(root, `//div[contains(@class,"pen_effect")]`)))

	// First item_box_value span is the attribute text for monsters. Spell/Trap
	// cards have only 1-2 such spans (their subtype text), so checking >= 4
	// distinguishes monster pages from spell/trap pages: the four monster slots
	// are attribute, level/rank/link, ATK, and DEF.
	//
	// On the same monster-page condition, the first slash-separated token of
	// <p class="species"> is the localized monster type (Konami's species
	// format is type-first; subsequent tokens are subtype/abilities like
	// "Effect"/"通常"/"Tuner"). Stored verbatim — no enum lookup needed.
	values, _ := textproc.HTMLXPath(root, `//span[@class="item_box_value"]`)
	if len(values) >= 4 {
		t.AttributeText = strings.TrimSpace(textproc.HTMLGetText(values[0]))
		speciesText := textproc.HTMLGetText(getNode(root, `//*[@class="species"]`))
		if tokens := strings.Split(speciesText, "/"); len(tokens) > 0 {
			t.MonsterTypeText = strings.TrimSpace(tokens[0])
		}
	}

	return t
}

// parseCardNameAndPronunciation splits a h1#cardname node into three parts.
// Konami's <h1> structure is locale-dependent:
//   - JA: <span class="ruby">[kana]</span>  [kanji text]  <span>[EN name]</span>
//   - KO:                                   [hangul text] <span>[EN name]</span>
//   - EN:                                   [EN text]
//   - Rush JA: <span class="ruby">[kana]</span>  [kana again]
//
// The discriminator across all of these is the "ruby" class:
//   - text-node children always belong to name (kanji/hangul/EN as printed)
//   - element with class="ruby" is the kana pronunciation (JA + Rush JA)
//   - element WITHOUT class="ruby" is the EN-name span (JA and KO carry it)
//
// Returning EN-on-page separately lets the crawler write cards.card_name_en
// from the JA pass while keeping name_katakana kana-only on JA and empty on KO.
func parseCardNameAndPronunciation(h1 *html.Node) (name, pronunciation, nameEN string) {
	var nameParts, pronunciationParts, nameENParts []string
	for child := h1.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			if t := strings.TrimSpace(child.Data); t != "" {
				nameParts = append(nameParts, t)
			}
		case html.ElementNode:
			t := strings.TrimSpace(textproc.HTMLGetText(child))
			if t == "" {
				continue
			}
			if hasClass(child, "ruby") {
				pronunciationParts = append(pronunciationParts, t)
			} else {
				nameENParts = append(nameENParts, t)
			}
		}
	}
	return strings.Join(nameParts, " "),
		strings.Join(pronunciationParts, " "),
		strings.Join(nameENParts, " ")
}

// hasClass reports whether n's class attribute contains the given token.
func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" && slices.Contains(strings.Fields(attr.Val), class) {
			return true
		}
	}
	return false
}
