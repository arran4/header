package header

import (
	"strings"

	"golang.org/x/net/html"
)

type openGraphExtractor struct{}

func (e openGraphExtractor) Name() string { return "OpenGraph" }

func (e openGraphExtractor) Init(info *PageInfo) Processor {
	return &openGraphProcessor{info: info}
}

type openGraphProcessor struct {
	info *PageInfo
}

func (p *openGraphProcessor) ProcessToken(t html.Token) bool {
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
				if strings.HasPrefix(property, "og:") {
					switch property {
					case "og:type":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGType = content
					case "og:image":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImage = content
					case "og:image:secure_url":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImageSecureURL = content
					case "og:image:type":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImageType = content
					case "og:image:width":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImageWidth = content
					case "og:image:height":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImageHeight = content
					case "og:image:alt":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGImageAlt = content
					case "og:description":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGDescription = content
					case "og:title":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGTitle = content
						if i.Title == "" {
							i.Title = content
						}
					case "og:video", "og:video:url":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideo = content
					case "og:video:secure_url":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideoSecureURL = content
					case "og:video:type":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideoType = content
					case "og:video:width":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideoWidth = content
					case "og:video:height":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideoHeight = content
					case "og:video:alt":
						i.KnownMetaTracker.MarkKnown(property)
						i.OGVideoAlt = content
					}
				}
			})
		case "body":
			return true
		}
	}
	return false
}

func init() {
	RegisterExtractor(openGraphExtractor{})
}
