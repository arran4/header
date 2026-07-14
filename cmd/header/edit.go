package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func findNodes(n *html.Node, tag string, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		if match == nil || match(n) {
			found = append(found, n)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found = append(found, findNodes(c, tag, match)...)
	}
	return found
}

func getOrCreateHead(doc *html.Node) *html.Node {
	heads := findNodes(doc, "head", nil)
	if len(heads) > 0 {
		return heads[0]
	}
	htmls := findNodes(doc, "html", nil)
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

func removeNodes(nodes []*html.Node) {
	for _, n := range nodes {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}
}

func setContent(n *html.Node, text string) {
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

func editCmd(args []string) int {
	fs := flag.NewFlagSet("header edit", flag.ContinueOnError)

	setVal := fs.String("s", "", "set (remove other versions if any)")
	addVal := fs.String("S", "", "add to set (add a new one, only works where valid)")
	replaceVal := fs.String("r", "", "replace (only works if replacing 1 value)")
	replaceIndex := fs.Int("n", -1, "replace number x, only works with -r")

	// We need to parse positional arguments before flags, because standard flag package
	// stops parsing flags when it hits a positional argument.
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  header edit head <file> -s <title>\n")
		fmt.Fprintf(os.Stderr, "  header edit meta <name> <file> -s <value>\n")
		return 2
	}

	commandType := args[0]
	var targetFile string
	var metaName string
	var flagIndex int

	if commandType == "head" {
		targetFile = args[1]
		flagIndex = 2
	} else if commandType == "meta" {
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: header edit meta <name> <file> ...\n")
			return 2
		}
		metaName = args[1]
		targetFile = args[2]
		flagIndex = 3
	} else {
		fmt.Fprintf(os.Stderr, "Unknown edit type: %s\n", commandType)
		return 2
	}

	if err := fs.Parse(args[flagIndex:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing edit flags: %v\n", err)
		return 2
	}

	var inputData []byte
	var err error
	if targetFile == "-" {
		inputData, err = io.ReadAll(os.Stdin)
	} else {
		inputData, err = os.ReadFile(targetFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return 1
	}

	doc, err := html.Parse(bytes.NewReader(inputData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
		return 1
	}

	if commandType == "head" {
		if *setVal != "" {
			titles := findNodes(doc, "title", nil)
			if len(titles) > 0 {
				for i := 1; i < len(titles); i++ {
					titles[i].Parent.RemoveChild(titles[i])
				}
				setContent(titles[0], *setVal)
			} else {
				head := getOrCreateHead(doc)
				t := &html.Node{Type: html.ElementNode, Data: "title"}
				setContent(t, *setVal)
				head.AppendChild(t)
			}
		} else if *addVal != "" {
			fmt.Fprintf(os.Stderr, "Error: cannot add (-S) multiple title tags.\n")
			return 1
		} else if *replaceVal != "" {
			titles := findNodes(doc, "title", nil)
			if len(titles) == 0 {
				fmt.Fprintf(os.Stderr, "Error: no title to replace.\n")
				return 1
			}
			idx := 0
			if *replaceIndex != -1 {
				idx = *replaceIndex
			}
			if idx < 0 || idx >= len(titles) {
				fmt.Fprintf(os.Stderr, "Error: invalid index %d for replace.\n", idx)
				return 1
			}
			setContent(titles[idx], *replaceVal)
		} else {
			fmt.Fprintf(os.Stderr, "Error: must specify -s, -S, or -r.\n")
			return 2
		}
	} else if commandType == "meta" {
		matchFunc := func(n *html.Node) bool {
			for _, a := range n.Attr {
				if (a.Key == "name" || a.Key == "property") && a.Val == metaName {
					return true
				}
			}
			return false
		}
		keyAttr := "name"
		if strings.HasPrefix(metaName, "og:") || strings.HasPrefix(metaName, "fb:") || strings.HasPrefix(metaName, "article:") {
			keyAttr = "property"
		}

		if *setVal != "" {
			metas := findNodes(doc, "meta", matchFunc)
			removeNodes(metas)
			head := getOrCreateHead(doc)
			m := &html.Node{
				Type: html.ElementNode,
				Data: "meta",
				Attr: []html.Attribute{
					{Key: keyAttr, Val: metaName},
					{Key: "content", Val: *setVal},
				},
			}
			head.AppendChild(m)
		} else if *addVal != "" {
			head := getOrCreateHead(doc)
			m := &html.Node{
				Type: html.ElementNode,
				Data: "meta",
				Attr: []html.Attribute{
					{Key: keyAttr, Val: metaName},
					{Key: "content", Val: *addVal},
				},
			}
			head.AppendChild(m)
		} else if *replaceVal != "" {
			metas := findNodes(doc, "meta", matchFunc)
			if len(metas) == 0 {
				fmt.Fprintf(os.Stderr, "Error: no meta %s to replace.\n", metaName)
				return 1
			}
			if len(metas) > 1 && *replaceIndex == -1 {
				fmt.Fprintf(os.Stderr, "Error: multiple meta %s found, use -n to specify which to replace.\n", metaName)
				return 1
			}
			idx := 0
			if *replaceIndex != -1 {
				idx = *replaceIndex
			}
			if idx < 0 || idx >= len(metas) {
				fmt.Fprintf(os.Stderr, "Error: invalid index %d for replace.\n", idx)
				return 1
			}
			target := metas[idx]
			foundContent := false
			for i, a := range target.Attr {
				if a.Key == "content" {
					target.Attr[i].Val = *replaceVal
					foundContent = true
					break
				}
			}
			if !foundContent {
				target.Attr = append(target.Attr, html.Attribute{Key: "content", Val: *replaceVal})
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: must specify -s, -S, or -r.\n")
			return 2
		}
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
		return 1
	}

	if targetFile == "-" {
		fmt.Print(buf.String())
	} else {
		if err := os.WriteFile(targetFile, buf.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			return 1
		}
	}

	return 0
}
