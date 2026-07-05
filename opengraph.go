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
			switch t.Data {
			case "meta":
				var property, name, content string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "property":
						property = attr.Val
					case "name":
						name = attr.Val
					case "content":
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
					case "og:image:alt":
						info.SetOGImageAlt(content)
					case "og:description":
						info.SetOGDescription(content)
					case "og:title":
						info.SetOGTitle(content)
					}
				}

				// Twitter Card
				if strings.HasPrefix(name, "twitter:") {
					switch name {
					case "twitter:card":
						info.SetTwitterCard(content)
					case "twitter:image":
						info.SetTwitterImage(content)
					case "twitter:image:alt":
						info.SetTwitterImageAlt(content)
					case "twitter:title":
						info.SetTwitterTitle(content)
					case "twitter:description":
						info.SetTwitterDescription(content)
					}
				}
			case "title":
				tt = z.Next()
				if tt == html.TextToken {
					info.SetTitleIfEmpty(z.Token().Data)
			}
			case "link":
				var rel, href string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "rel":
						rel = attr.Val
					case "href":
						href = attr.Val
					}
				}
				if rel == "canonical" {
					info.SetCanonicalLink(href)
				}
			case "body":
				// Once we hit body, we can typically stop parsing for metadata
				return nil
			}
		}
	}
}

func init() {
	RegisterExtractor(openGraphExtractor{})
}
