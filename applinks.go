package header

import (
	"strings"

	"golang.org/x/net/html"
)

type appLinksExtractor struct{}

func (e appLinksExtractor) Name() string { return "AppLinks" }

func (e appLinksExtractor) Extract(n *html.Node, info *PageInfo) error {
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
	RegisterExtractor(appLinksExtractor{})
}
