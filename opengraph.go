package header

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type openGraphExtractor struct{}

func (e openGraphExtractor) Name() string { return "OpenGraph" }

func (e openGraphExtractor) Extract(r io.Reader, info *PageInfo) error {
	z := html.NewTokenizer(r)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		}

		if tt == html.StartTagToken || tt == html.SelfClosingTagToken {
			t := z.Token()
			if t.Data == "meta" {
				var property, name, content string
				for _, attr := range t.Attr {
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
						info.SetOGType(content)
					case "og:image":
						info.SetOGImage(content)
					case "og:description":
						info.SetOGDescription(content)
					case "og:title":
						info.SetTitleIfEmpty(content)
					}
				}

				// Twitter Card
				if strings.HasPrefix(name, "twitter:") {
					switch name {
					case "twitter:card":
						info.SetTwitterCard(content)
					case "twitter:image":
						info.SetTwitterImage(content)
					case "twitter:title":
						info.SetTwitterTitle(content)
					case "twitter:description":
						info.SetTwitterDescription(content)
					}
				}
			} else if t.Data == "title" {
				tt = z.Next()
				if tt == html.TextToken {
					info.SetTitleIfEmpty(z.Token().Data)
				}
			} else if t.Data == "body" {
				// Once we hit body, we can typically stop parsing for metadata
				return nil
			}
		}
	}
}

func init() {
	RegisterExtractor(openGraphExtractor{})
}
