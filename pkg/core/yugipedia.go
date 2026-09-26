package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mywrap/textproc"
	"golang.org/x/net/html"

	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

// ParseYugipediaSetChronology parses Yugipedia's Set Chronology HTML page
// (https://yugipedia.com/wiki/Set_chronology) and returns the OCG and TCG
// sets it lists. The page must be fetched and saved separately (Yugipedia
// has bot protection); pkg/core/set_yugipedia.html is the embedded snapshot
// used by tests.
//
// Yugipedia is an English wiki, so for both OCG and TCG rows the parser
// fills only KonamiSet.NameEN; NameJA and NameKO stay empty (filled later
// from the Konami crawl).
func ParseYugipediaSetChronology(htmlData []byte) ([]konami.KonamiSet, error) {
	root := textproc.HTMLParseToNode(htmlData)

	// because of XPath preceding and following do not work the same way in different libraries and XPath versions,
	// we need the following splitting approach to get the OCG and TCG tables separately:

	// split the HTML to get part1 including OCG and TCG tables (part2 is Rush Duel and beyond)
	rushDuelBegin, err := textproc.HTMLXPath(root, `//h2[.//*[@id="Rush_Duel"]]`)
	if err != nil {
		return nil, err
	}
	if len(rushDuelBegin) != 1 {
		return nil, fmt.Errorf("unexpected HTML structure, cannot find where Rush Duel section begins, len(rushDuelBegin): %v", len(rushDuelBegin))
	}
	partOCGAndTCG, _, err := SplitHTML(root, rushDuelBegin[0])
	if err != nil {
		return nil, fmt.Errorf("error SplitHTML at rushDuelBegin: %w", err)
	}

	// split OCG and TCG parts
	tcgBegin, err := textproc.HTMLXPath(partOCGAndTCG, `//h2[.//*[@id="TCG"]]`)
	if err != nil {
		return nil, err
	}
	if len(tcgBegin) != 1 {
		return nil, fmt.Errorf("unexpected HTML structure, cannot find where TCG section begins, len(tcgBegin): %v", len(tcgBegin))
	}
	partOCG, partTCG, err := SplitHTML(partOCGAndTCG, tcgBegin[0])
	if err != nil {
		return nil, fmt.Errorf("error SplitHTML at tcgBegin: %w", err)
	}

	// now extract all tables from OCG and TCG parts
	ocgTables, err := textproc.HTMLXPath(partOCG, `//table`)
	if err != nil {
		return nil, fmt.Errorf("error HTMLXPath ocgTables: %w", err)
	}
	tcgTables, err := textproc.HTMLXPath(partTCG, `//table`)
	if err != nil {
		return nil, fmt.Errorf("error HTMLXPath tcgTables: %w", err)
	}
	if len(ocgTables) == 0 && len(tcgTables) == 0 {
		return nil, errors.New("no tables found for OCG or TCG, check the XPath")
	}
	var results []konami.KonamiSet
	for _, v := range []struct {
		ygoVer konami.YuGiOhVersion
		tables []*html.Node
	}{
		{ygoVer: konami.OCG, tables: ocgTables},
		{ygoVer: konami.TCG, tables: tcgTables},
	} {
		for _, table := range v.tables {
			rows, err := textproc.HTMLXPath(table, `//tr`)
			if err != nil {
				return nil, fmt.Errorf("error HTMLXPath tr: %w", err)
			}
			for _, row := range rows {
				cells, err := textproc.HTMLXPath(row, `//td`)
				if err != nil {
					return nil, err
				}
				if len(cells) < 4 {
					continue
				}
				s := konami.KonamiSet{
					YuGiOhVersion: v.ygoVer,
					Abbreviation:  strings.ReplaceAll(textproc.HTMLGetText(cells[0]), "\n", " "),
					// Yugipedia is an English wiki: NameEN only; the Konami
					// crawler fills NameJA / NameKO from a different source.
					NameEN:      strings.ReplaceAll(textproc.HTMLGetText(cells[1]), "\n", " "),
					ReleaseDate: "1970-01-01",
				}

				isParseTimeSuccess := false
				releaseDateStr := textproc.HTMLGetText(cells[3])
				if releaseDateStr == "" {
					continue
				}
				releaseDateStr = strings.TrimPrefix(releaseDateStr, "Unknown ")
				// parse release date: "4 February 1999", "March 2001", "Unknown 2003", ...
				for _, timeFormat := range []string{
					"2 January 2006",
					"January 2006",
					"2006",
				} {
					releaseDate, err := time.Parse(timeFormat, releaseDateStr)
					if err == nil {
						s.ReleaseDate = releaseDate.Format("2006-01-02")
						isParseTimeSuccess = true
						break
					}
				}
				if !isParseTimeSuccess {
					continue
				}
				results = append(results, s)
			}
		}
	}
	return results, nil
}

// SplitHTML splits the whole HTML node into two parts at the cutFrom node.
// The cutFrom node itself will be included in part2.
// Both part1 and part2 will have a single root node,
// parent will be cloned for both parts, but only the relevant children will be included.
//
// Used by ParseYugipediaSetChronology to slice the chronology page into the
// OCG, TCG, and Rush Duel sections (XPath preceding/following axes vary
// across libraries, so manual splitting is more portable).
func SplitHTML(whole *html.Node, cutFrom *html.Node) (part1, part2 *html.Node, err error) {
	if whole == nil || cutFrom == nil {
		return nil, nil, errors.New("nil node")
	}

	// Verify cutFrom is inside whole
	if !isDescendant(whole, cutFrom) {
		return nil, nil, errors.New("cut node not inside whole")
	}

	// Path from whole down to cutFrom.Parent (inclusive): [whole, ..., cutFrom.Parent]
	path := make([]*html.Node, 0)
	for p := cutFrom.Parent; p != nil; p = p.Parent {
		path = append(path, p)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	part1 = shallowClone(whole)
	part2 = shallowClone(whole)

	// Walk the tree: on the path from whole to cutFrom, clone for both parts; before path -> part1, cutFrom and after -> part2.
	var walk func(n *html.Node, p1, p2 *html.Node, path []*html.Node, pi int, afterCut bool)
	walk = func(n *html.Node, p1, p2 *html.Node, path []*html.Node, pi int, afterCut bool) {
		pathChild := cutFrom
		if pi+1 < len(path) {
			pathChild = path[pi+1]
		}
		passedPath := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c == pathChild {
				if c != cutFrom {
					cloned1 := shallowClone(c)
					cloned2 := shallowClone(c)
					p1.AppendChild(cloned1)
					p2.AppendChild(cloned2)
					walk(c, cloned1, cloned2, path, pi+1, afterCut)
				} else {
					cloned := shallowClone(c)
					p2.AppendChild(cloned)
					walk(c, p1, cloned, path, pi+1, true)
				}
				passedPath = true
			} else if passedPath || afterCut || c == cutFrom {
				cloned := shallowClone(c)
				p2.AppendChild(cloned)
				walk(c, p1, cloned, path, pi+1, true)
			} else {
				cloned := shallowClone(c)
				p1.AppendChild(cloned)
				walk(c, cloned, p2, path, pi+1, false)
			}
		}
	}
	walk(whole, part1, part2, path, 0, false)

	return part1, part2, nil
}

func shallowClone(n *html.Node) *html.Node {
	return &html.Node{
		Type: n.Type,
		Data: n.Data,
		Attr: cloneAttrs(n.Attr),
	}
}

func cloneAttrs(attrs []html.Attribute) []html.Attribute {
	out := make([]html.Attribute, len(attrs))
	copy(out, attrs)
	return out
}

func isDescendant(root, target *html.Node) bool {
	var found bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == target {
			found = true
			return
		}
		for c := n.FirstChild; c != nil && !found; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return found
}
