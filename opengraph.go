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

				info.Mutate(func(i *PageInfo) {
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

					// App Links
					if strings.HasPrefix(property, "al:") {
						switch property {
						case "al:ios:app_name":
							i.AppLinksIOSName = content
						case "al:ios:app_store_id":
							i.AppLinksIOSAppStoreID = content
						case "al:ios:url":
							i.AppLinksIOSURL = content
						case "al:android:app_name":
							i.AppLinksAndroidName = content
						case "al:android:package":
							i.AppLinksAndroidPackage = content
						case "al:android:url":
							i.AppLinksAndroidURL = content
						}
					}

					// Instagram specific
					if property == "instapp:owner_user_id" {
						i.InstagramOwnerUserID = content
					}

					// OpenGraph
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

					// Twitter Card
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
			case "title":
				tt = z.Next()
				if tt == html.TextToken {
					data := z.Token().Data
					info.Mutate(func(i *PageInfo) {
						if i.Title == "" {
							i.Title = data
						}
					})
				}
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
				info.Mutate(func(i *PageInfo) {
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
					if strings.EqualFold(rel, "alternate") && href != "" {
						typeLower := strings.ToLower(typ)
						if strings.Contains(typeLower, "rss") || strings.Contains(typeLower, "atom") || strings.Contains(typeLower, "json") {
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
				// Once we hit body, we can typically stop parsing for metadata
				return nil
			}
		}
	}
}

func init() {
	RegisterExtractor(openGraphExtractor{})
}
