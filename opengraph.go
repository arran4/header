package header

import (
	"strings"
	"golang.org/x/net/html"
)

func extractOpenGraph(doc *html.Node, info *PageInfo) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, name, content string
			for _, attr := range n.Attr {
				if attr.Key == "property" {
					property = attr.Val
				} else if attr.Key == "name" {
					name = attr.Val
				} else if attr.Key == "content" {
					content = attr.Val
				}
			}

			// OpenGraph
			if strings.HasPrefix(property, "og:") {
				switch property {
				case "og:type":
					info.OGType = content
				case "og:image":
					info.OGImage = content
				case "og:description":
					info.OGDescription = content
				case "og:title":
					if info.Title == "" {
						info.Title = content
					}
				}
			}

			// Twitter Card
			if strings.HasPrefix(name, "twitter:") {
				switch name {
				case "twitter:card":
					info.TwitterCard = content
				case "twitter:image":
					info.TwitterImage = content
				case "twitter:title":
					info.TwitterTitle = content
					if info.Title == "" {
						info.Title = content
					}
				case "twitter:description":
					info.TwitterDescription = content
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "title" && info.Title == "" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				info.Title = n.FirstChild.Data
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
}
