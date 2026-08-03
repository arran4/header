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
			var property string
			var itemprop string
			for _, attr := range t.Attr {
				switch attr.Key {
				case "http-equiv":
					httpEquiv = attr.Val
				case "name":
					name = attr.Val
				case "property":
					property = attr.Val
				case "itemprop":
					itemprop = attr.Val
				case "content":
					content = attr.Val
				}
			}

			p.info.Mutate(func(i *PageInfo) {
				var tagKey string
				if name != "" {
					tagKey = name
				} else if property != "" {
					tagKey = property
				} else if itemprop != "" {
					tagKey = itemprop
				} else if httpEquiv != "" {
					tagKey = httpEquiv
				}

				if tagKey != "" {
					i.KnownMetaTracker.AddRaw(tagKey, content)
				}

				// Meta Refresh
				if strings.EqualFold(httpEquiv, "refresh") {
					i.KnownMetaTracker.MarkKnown(httpEquiv)
					if i.MetaRefresh == "" {
						i.MetaRefresh = content
					}
				}

				// Standard Metadata
				if strings.EqualFold(name, "keywords") {
					i.KnownMetaTracker.MarkKnown(name)
					i.Keywords = content
				}
				if strings.EqualFold(name, "theme-color") {
					i.KnownMetaTracker.MarkKnown(name)
					i.ThemeColor = content
				}
				if strings.EqualFold(name, "application-title") {
					i.KnownMetaTracker.MarkKnown(name)
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
				if rel != "" && href != "" {
					i.KnownMetaTracker.AddRaw("link:"+rel, href)
				}

				if strings.EqualFold(rel, "canonical") && href != "" {
					i.KnownMetaTracker.MarkKnown("link:"+rel)
					i.CanonicalLink = href
				}
				if strings.EqualFold(rel, "shortlink") && href != "" {
					i.KnownMetaTracker.MarkKnown("link:"+rel)
					i.ShortLink = href
				}
				// Support non-standard shortlinkUrl variant often found in the wild (e.g. YouTube)
				if strings.EqualFold(rel, "shortlinkurl") && href != "" {
					i.KnownMetaTracker.MarkKnown("link:"+rel)
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
					i.KnownMetaTracker.MarkKnown("link:"+rel)
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
					i.KnownMetaTracker.MarkKnown("link:"+rel)
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
