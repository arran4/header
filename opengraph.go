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
				var httpEquiv string
				var property, name, content string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "http-equiv":
						httpEquiv = attr.Val
					case "property":
						property = attr.Val
					case "name":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}

				// Meta Refresh
				if strings.EqualFold(httpEquiv, "refresh") {
					info.SetMetaRefresh(content)
				}

				// Standard Metadata
				if strings.EqualFold(name, "keywords") {
					info.SetKeywords(content)
				}

				// OpenGraph
				if strings.HasPrefix(property, "og:") {
					switch property {
					case "og:type":
						info.SetOGType(content)
					case "og:image":
						info.SetOGImage(content)
					case "og:image:secure_url":
						info.SetOGImageSecureURL(content)
					case "og:image:type":
						info.SetOGImageType(content)
					case "og:image:width":
						info.SetOGImageWidth(content)
					case "og:image:height":
						info.SetOGImageHeight(content)
					case "og:image:alt":
						info.SetOGImageAlt(content)
					case "og:description":
						info.SetOGDescription(content)
					case "og:title":
						info.SetOGTitle(content)
					case "og:video", "og:video:url":
						info.SetOGVideo(content)
					case "og:video:secure_url":
						info.SetOGVideoSecureURL(content)
					case "og:video:type":
						info.SetOGVideoType(content)
					case "og:video:width":
						info.SetOGVideoWidth(content)
					case "og:video:height":
						info.SetOGVideoHeight(content)
					case "og:video:alt":
						info.SetOGVideoAlt(content)
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
					case "twitter:player":
						info.SetTwitterPlayer(content)
					case "twitter:player:width":
						info.SetTwitterPlayerWidth(content)
					case "twitter:player:height":
						info.SetTwitterPlayerHeight(content)
					}
				}
			case "title":
				tt = z.Next()
				if tt == html.TextToken {
					info.SetTitleIfEmpty(z.Token().Data)
			}
			case "link":
				var rel, href, sizes, media string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "rel":
						rel = attr.Val
					case "href":
						href = attr.Val
					case "sizes":
						sizes = attr.Val
					case "media":
						media = attr.Val
					}
				}
				if strings.EqualFold(rel, "canonical") && href != "" {
					info.SetCanonicalLink(href)
				}
				if strings.EqualFold(rel, "shortlink") && href != "" {
					info.SetShortLink(href)
				}
				// Support non-standard shortlinkUrl variant often found in the wild (e.g. YouTube)
				if strings.EqualFold(rel, "shortlinkurl") && href != "" {
					info.SetShortLink(href)
				}
				if strings.EqualFold(rel, "alternate") && href != "" {
					if strings.EqualFold(media, "handheld") || strings.Contains(strings.ToLower(media), "max-width") {
						info.SetAlternateHandheldLink(href)
					}
				}
				if href != "" && strings.Contains(strings.ToLower(rel), "icon") {
					info.AddIcon(href, sizes)
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
