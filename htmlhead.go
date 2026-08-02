package header

import (
	"strings"

	"golang.org/x/net/html"
)

type htmlHeadExtractor struct{}

func (e htmlHeadExtractor) Name() string { return "HTMLHead" }

func (e htmlHeadExtractor) Init(info *PageInfo) Processor {
	return &htmlHeadProcessor{info: info}
}

type htmlHeadProcessor struct {
	info    *PageInfo
	inTitle bool
}

func (p *htmlHeadProcessor) ProcessToken(t html.Token) bool {
	if t.Type == html.StartTagToken || t.Type == html.SelfClosingTagToken {
		switch t.Data {
		case "meta":
			var httpEquiv string
			var name, content string
			for _, attr := range t.Attr {
				switch attr.Key {
				case "http-equiv":
					httpEquiv = attr.Val
				case "name":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}

			p.info.Mutate(func(i *PageInfo) {
				// Meta Refresh
				if strings.EqualFold(httpEquiv, "refresh") {
					if i.MetaRefresh == "" {
						i.MetaRefresh = content
					}
				}

				// Standard Metadata
				if strings.EqualFold(name, "keywords") {
					i.Keywords = content
				}
				if strings.EqualFold(name, "theme-color") {
					i.ThemeColor = content
				}
				if strings.EqualFold(name, "application-title") {
					i.ApplicationTitle = content
				}
			})
		case "title":
			p.inTitle = true
		case "link":
			var rel, href, sizes, media, typ, title string
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
				case "type":
					typ = attr.Val
				case "title":
					title = attr.Val
				}
			}
			p.info.Mutate(func(i *PageInfo) {
				if strings.EqualFold(rel, "canonical") && href != "" {
					i.CanonicalLink = href
				}
				if strings.EqualFold(rel, "shortlink") && href != "" {
					i.ShortLink = href
				}
				// Support non-standard shortlinkUrl variant often found in the wild (e.g. YouTube)
				if strings.EqualFold(rel, "shortlinkurl") && href != "" {
					i.ShortLink = href
				}

				isAlternate := false
				for _, token := range strings.Fields(rel) {
					if strings.EqualFold(token, "alternate") {
						isAlternate = true
						break
					}
				}

				if isAlternate && href != "" {
					typeLower := strings.ToLower(typ)
					hrefLower := strings.ToLower(href)
					isFeed := strings.Contains(typeLower, "rss") || strings.Contains(typeLower, "atom") || strings.Contains(typeLower, "json")
					if !isFeed && typeLower == "" {
						isFeed = strings.HasSuffix(hrefLower, ".rss") || strings.HasSuffix(hrefLower, ".atom") || strings.HasSuffix(hrefLower, ".xml")
					}

					if isFeed {
						found := false
						for _, existing := range i.FeedLinks {
							if existing.URL == href && existing.Type == typ {
								found = true
								break
							}
						}
						if !found {
							i.FeedLinks = append(i.FeedLinks, FeedLink{URL: href, Type: typ, Title: title})
						}
					} else {
						found := false
						for _, existing := range i.AlternateLinks {
							if existing.URL == href && existing.Media == media {
								found = true
								break
							}
						}
						if !found {
							i.AlternateLinks = append(i.AlternateLinks, AlternateLink{URL: href, Media: media})
						}
					}
				}
				if href != "" && strings.Contains(strings.ToLower(rel), "icon") {
					found := false
					for _, existing := range i.Icons {
						if existing.URL == href && existing.Sizes == sizes {
							found = true
							break
						}
					}
					if !found {
						i.Icons = append(i.Icons, Icon{URL: href, Sizes: sizes})
					}
				}
			})
		case "body":
			return true
		}
	} else if t.Type == html.TextToken {
		if p.inTitle {
			data := t.Data
			p.info.Mutate(func(i *PageInfo) {
				if i.Title == "" {
					i.Title = data
				}
			})
		}
	} else if t.Type == html.EndTagToken {
		if t.Data == "title" {
			p.inTitle = false
		}
	}

	return false
}

func init() {
	RegisterExtractor(htmlHeadExtractor{})
}
