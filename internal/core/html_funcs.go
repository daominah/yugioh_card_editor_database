package core

import (
	"errors"

	"golang.org/x/net/html"
)

// SplitHTML splits the whole HTML node into two parts at the cutFrom node.
// The cutFrom node itself will be included in part2.
// Both part1 and part2 will have a single root node,
// parent will be cloned for both parts, but only the relevant children will be included.
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
