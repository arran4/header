package header

import (
	"strings"

	"golang.org/x/net/html"
)

type appLinksExtractor struct{}

func (e appLinksExtractor) Name() string { return "AppLinks" }

func (e appLinksExtractor) Init(info *PageInfo) Processor {
	return &appLinksProcessor{info: info}
}

type appLinksProcessor struct {
	info *PageInfo
}

func (p *appLinksProcessor) ProcessToken(t html.Token) bool {
	if t.Type == html.StartTagToken || t.Type == html.SelfClosingTagToken {
		if t.Data == "meta" {
			var property, content string
			for _, attr := range t.Attr {
				switch attr.Key {
				case "property":
					property = attr.Val
				case "content":
					content = attr.Val
				}
			}

			p.info.Mutate(func(i *PageInfo) {
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
		} else if t.Data == "body" {
			return true
		}
	}
	return false
}

func init() {
	RegisterExtractor(appLinksExtractor{})
}
