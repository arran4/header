package header

import (
	"strings"

	"golang.org/x/net/html"
)

type openGraphExtractor struct{}

func (e openGraphExtractor) Name() string { return "OpenGraph" }

func (e openGraphExtractor) Extract(n *html.Node, info *PageInfo) error {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "property":
					property = attr.Val
				case "content":
					content = attr.Val
				}
			}

			info.Mutate(func(i *PageInfo) {
				if strings.HasPrefix(property, "og:") {
					switch property {
					case "og:type":
						i.OGType = content
					case "og:image":
						i.OGImage = content
					case "og:image:secure_url":
						i.OGImageSecureURL = content
					case "og:image:type":
						i.OGImageType = content
					case "og:image:width":
						i.OGImageWidth = content
					case "og:image:height":
						i.OGImageHeight = content
					case "og:image:alt":
						i.OGImageAlt = content
					case "og:description":
						i.OGDescription = content
					case "og:title":
						i.OGTitle = content
						if i.Title == "" {
							i.Title = content
						}
					case "og:video", "og:video:url":
						i.OGVideo = content
					case "og:video:secure_url":
						i.OGVideoSecureURL = content
					case "og:video:type":
						i.OGVideoType = content
					case "og:video:width":
						i.OGVideoWidth = content
					case "og:video:height":
						i.OGVideoHeight = content
					case "og:video:alt":
						i.OGVideoAlt = content
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
	RegisterExtractor(openGraphExtractor{})
}
