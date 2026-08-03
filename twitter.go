package header

import (
	"strings"

	"golang.org/x/net/html"
)

type twitterExtractor struct{}

func (e twitterExtractor) Name() string { return "Twitter" }

func (e twitterExtractor) Init(info *PageInfo) Processor {
	return &twitterProcessor{info: info}
}

type twitterProcessor struct {
	info *PageInfo
}

func (p *twitterProcessor) ProcessToken(t html.Token) bool {
	if t.Type == html.StartTagToken || t.Type == html.SelfClosingTagToken {
		switch t.Data {
		case "meta":
			var name, content string
			for _, attr := range t.Attr {
				switch attr.Key {
				case "name":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}

			p.info.Mutate(func(i *PageInfo) {
				if strings.HasPrefix(name, "twitter:") {
					switch name {
					case "twitter:card":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterCard = content
					case "twitter:image":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterImage = content
					case "twitter:image:alt":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterImageAlt = content
					case "twitter:title":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterTitle = content
						if i.Title == "" {
							i.Title = content
						}
					case "twitter:description":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterDescription = content
					case "twitter:player":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterPlayer = content
					case "twitter:player:width":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterPlayerWidth = content
					case "twitter:player:height":
						i.KnownMetaTracker.MarkKnown(name)
						i.TwitterPlayerHeight = content
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
	RegisterExtractor(twitterExtractor{})
}
