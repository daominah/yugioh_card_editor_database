package konami

import (
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
	// NamePronunciation is the phonetic katakana pronunciation from the ruby
	// span on JA pages. Empty for EN and KO.
	NamePronunciation string
	Effect            string
	PendulumEffect    string
	// AttributeText is the localized attribute display: "LIGHT" / "光属性" / "빛".
	// Empty for Spell and Trap cards.
	AttributeText string
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
		t.Name, t.NamePronunciation = parseCardNameAndPronunciation(h1)
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
	// distinguishes monster pages from spell/trap pages.
	values, _ := textproc.HTMLXPath(root, `//span[@class="item_box_value"]`)
	if len(values) >= 4 {
		t.AttributeText = strings.TrimSpace(textproc.HTMLGetText(values[0]))
	}

	return t
}

// parseCardNameAndPronunciation splits a h1#cardname node:
//   - name: text from direct text-node children only (kanji/Hangul/EN name)
//   - pronunciation: text from element children (katakana ruby span on JA pages)
//
// On EN/KO pages the h1 contains only a text node, so pronunciation is always empty.
func parseCardNameAndPronunciation(h1 *html.Node) (name, pronunciation string) {
	var nameParts, pronunciationParts []string
	for child := h1.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			if t := strings.TrimSpace(child.Data); t != "" {
				nameParts = append(nameParts, t)
			}
		case html.ElementNode:
			if t := strings.TrimSpace(textproc.HTMLGetText(child)); t != "" {
				pronunciationParts = append(pronunciationParts, t)
			}
		}
	}
	return strings.Join(nameParts, " "), strings.Join(pronunciationParts, " ")
}
