package header

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// EditOptions configures the edit operation.
type EditOptions struct {
	CommandType  string // "head" or "meta"
	MetaName     string
	SetVal       *string
	AddVal       *string
	ReplaceVal   *string
	ReplaceIndex int
}

// FindNodes finds all HTML nodes with the given tag that match the predicate.
func FindNodes(n *html.Node, tag string, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	var travel func(*html.Node)
	travel = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tag {
			if match == nil || match(node) {
				found = append(found, node)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			travel(c)
		}
	}
	travel(n)
	return found
}

// GetOrCreateHead finds the <head> element or creates it if it doesn't exist.
func GetOrCreateHead(doc *html.Node) *html.Node {
	heads := FindNodes(doc, "head", nil)
	if len(heads) > 0 {
		return heads[0]
	}
	htmls := FindNodes(doc, "html", nil)
	var htmlNode *html.Node
	if len(htmls) > 0 {
		htmlNode = htmls[0]
	} else {
		htmlNode = doc
	}
	head := &html.Node{
		Type: html.ElementNode,
		Data: "head",
	}
	if htmlNode == doc {
		htmlNode.AppendChild(head)
	} else {
		htmlNode.InsertBefore(head, htmlNode.FirstChild)
	}
	return head
}

// RemoveNodes removes the specified nodes from their parents.
func RemoveNodes(nodes []*html.Node) {
	for _, n := range nodes {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}
}

// SetContent removes all children of a node and appends a single text node.
func SetContent(n *html.Node, text string) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		n.RemoveChild(c)
		c = next
	}
	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}

// ApplyEdit parses an HTML document from the reader, applies the edit options, and returns the modified document.
func ApplyEdit(r io.Reader, opts EditOptions) (*html.Node, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	if opts.CommandType == "head" {
		if opts.SetVal != nil {
			titles := FindNodes(doc, "title", nil)
			if len(titles) > 0 {
				for i := 1; i < len(titles); i++ {
					titles[i].Parent.RemoveChild(titles[i])
				}
				SetContent(titles[0], *opts.SetVal)
			} else {
				head := GetOrCreateHead(doc)
				t := &html.Node{Type: html.ElementNode, Data: "title"}
				SetContent(t, *opts.SetVal)
				head.AppendChild(t)
			}
		} else if opts.AddVal != nil {
			return nil, errors.New("cannot add (-S) multiple title tags")
		} else if opts.ReplaceVal != nil {
			titles := FindNodes(doc, "title", nil)
			if len(titles) == 0 {
				return nil, errors.New("no title to replace")
			}
			idx := 0
			if opts.ReplaceIndex != -1 {
				idx = opts.ReplaceIndex
			}
			if idx < 0 || idx >= len(titles) {
				return nil, fmt.Errorf("invalid index %d for replace", idx)
			}
			SetContent(titles[idx], *opts.ReplaceVal)
		} else {
			return nil, errors.New("must specify -s, -S, or -r")
		}
	} else if opts.CommandType == "meta" {
		matchFunc := func(n *html.Node) bool {
			for _, a := range n.Attr {
				if (a.Key == "name" || a.Key == "property") && a.Val == opts.MetaName {
					return true
				}
			}
			return false
		}
		keyAttr := "name"
		if strings.HasPrefix(opts.MetaName, "og:") || strings.HasPrefix(opts.MetaName, "fb:") || strings.HasPrefix(opts.MetaName, "article:") {
			keyAttr = "property"
		}

		if opts.SetVal != nil {
			metas := FindNodes(doc, "meta", matchFunc)
			RemoveNodes(metas)
			head := GetOrCreateHead(doc)
			m := &html.Node{
				Type: html.ElementNode,
				Data: "meta",
				Attr: []html.Attribute{
					{Key: keyAttr, Val: opts.MetaName},
					{Key: "content", Val: *opts.SetVal},
				},
			}
			head.AppendChild(m)
		} else if opts.AddVal != nil {
			head := GetOrCreateHead(doc)
			m := &html.Node{
				Type: html.ElementNode,
				Data: "meta",
				Attr: []html.Attribute{
					{Key: keyAttr, Val: opts.MetaName},
					{Key: "content", Val: *opts.AddVal},
				},
			}
			head.AppendChild(m)
		} else if opts.ReplaceVal != nil {
			metas := FindNodes(doc, "meta", matchFunc)
			if len(metas) == 0 {
				return nil, fmt.Errorf("no meta %s to replace", opts.MetaName)
			}
			if len(metas) > 1 && opts.ReplaceIndex == -1 {
				return nil, fmt.Errorf("multiple meta %s found, use -n to specify which to replace", opts.MetaName)
			}
			idx := 0
			if opts.ReplaceIndex != -1 {
				idx = opts.ReplaceIndex
			}
			if idx < 0 || idx >= len(metas) {
				return nil, fmt.Errorf("invalid index %d for replace", idx)
			}
			target := metas[idx]
			foundContent := false
			for i, a := range target.Attr {
				if a.Key == "content" {
					target.Attr[i].Val = *opts.ReplaceVal
					foundContent = true
					break
				}
			}
			if !foundContent {
				target.Attr = append(target.Attr, html.Attribute{Key: "content", Val: *opts.ReplaceVal})
			}
		} else {
			return nil, errors.New("must specify -s, -S, or -r")
		}
	} else {
		return nil, fmt.Errorf("unknown edit type: %s", opts.CommandType)
	}

	return doc, nil
}
