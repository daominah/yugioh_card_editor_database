package konami

import (
	"github.com/mywrap/textproc"
	"golang.org/x/net/html"
)

// Parser holds a parsed Konami card-detail page and exposes the per-card
// extractors as methods. The crawl uses one Parser per cardPageHTML so the
// HTML is parsed exactly once and the DOM tree is shared across the four
// extractors below; calling each Parse* convenience function on the same
// bytes would otherwise reparse the ~150 KB document four times per card.
//
// The relationCard section (sibling-card preview embedded by Konami) is
// removed once at construction time, so all extractors see the same cleaned
// tree without each one having to repeat the removal.
type Parser struct {
	root   *html.Node
	cardID CardID
}

// NewParser parses cardPageHTML into a *html.Node, removes the relationCard
// section, and returns a ready-to-use Parser. Cardpage parsing is the most
// expensive thing the crawler does per card; reuse this Parser across all
// extractors instead of calling the byte-level Parse* helpers.
func NewParser(cardPageHTML []byte, cardID CardID) *Parser {
	root := textproc.HTMLParseToNode(cardPageHTML)
	if rel := firstNode(root, `//*[@id="relationCard"]`); rel.Parent != nil {
		rel.Parent.RemoveChild(rel)
	}
	return &Parser{root: root, cardID: cardID}
}

// Card extracts the standard (TCG/OCG) card stats. Empty Card if the page
// is not a yugiohdb card page.
func (p *Parser) Card() Card { return parseKonamiCardFromNode(p.root, p.cardID) }

// RushCard extracts the Rush Duel card stats. Use Card for yugiohdb pages.
func (p *Parser) RushCard() CardRushDuel { return parseRushDuelCardFromNode(p.root, p.cardID) }

// LocaleText extracts the page-locale text fields (kanji/hangul/EN name,
// kana pronunciation, EN-name span, effect text, locale-specific attribute
// and monster-type strings).
func (p *Parser) LocaleText() CardLocaleText { return parseCardLocaleTextFromNode(p.root, p.cardID) }

// Prints extracts every printed appearance of the card from the
// #update_list section.
func (p *Parser) Prints() []CardPrint { return parseCardPrintsFromNode(p.root, p.cardID) }

// firstNode is the local equivalent of the per-function getNode closures
// in the extractors: returns the first XPath match or a stub *html.Node so
// callers can do ".Parent != nil" guards without nil checks.
func firstNode(parent *html.Node, xpath string) *html.Node {
	nodes, _ := textproc.HTMLXPath(parent, xpath)
	if len(nodes) == 0 {
		return &html.Node{}
	}
	return nodes[0]
}
