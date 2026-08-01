package header

import (
	"strings"

	"golang.org/x/net/html"
)

type twitterExtractor struct{}

func (e twitterExtractor) Name() string { return "Twitter" }

func (e twitterExtractor) Extract(n *html.Node, info *PageInfo) error {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}

			info.Mutate(func(i *PageInfo) {
				if strings.HasPrefix(name, "twitter:") {
					switch name {
					case "twitter:card":
						i.TwitterCard = content
					case "twitter:image":
						i.TwitterImage = content
					case "twitter:image:alt":
						i.TwitterImageAlt = content
					case "twitter:title":
						i.TwitterTitle = content
						if i.Title == "" {
							i.Title = content
						}
					case "twitter:description":
						i.TwitterDescription = content
					case "twitter:player":
						i.TwitterPlayer = content
					case "twitter:player:width":
						i.TwitterPlayerWidth = content
					case "twitter:player:height":
						i.TwitterPlayerHeight = content
					}
				}
			})
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return nil
}

func init() {
	RegisterExtractor(twitterExtractor{})
}
