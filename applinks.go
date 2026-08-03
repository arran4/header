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
		switch t.Data {
		case "meta":
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
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksIOSName = content
					case "al:ios:app_store_id":
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksIOSAppStoreID = content
					case "al:ios:url":
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksIOSURL = content
					case "al:android:app_name":
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksAndroidName = content
					case "al:android:package":
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksAndroidPackage = content
					case "al:android:url":
						i.KnownMetaTracker.MarkKnown(property)
						i.AppLinksAndroidURL = content
					}
				}

				// Instagram specific
				if property == "instapp:owner_user_id" {
					i.InstagramOwnerUserID = content
				}
			})
		case "body":
			return true
		}
	}
	return false
}

func init() {
	RegisterExtractor(appLinksExtractor{})
}
