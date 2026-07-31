package header

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type feedExtractor struct{}

func (e feedExtractor) Name() string { return "Feed" }

func (e feedExtractor) Extract(r io.Reader, info *PageInfo) error {
	z := html.NewTokenizer(r)

	isFeed := false

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		}

		if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "rss" || t.Data == "feed" || t.Data == "channel" {
				isFeed = true
			}

			if isFeed && t.Data == "title" {
				tt = z.Next()
				if tt == html.TextToken {
					data := z.Token().Data
					info.Mutate(func(i *PageInfo) {
						if i.Title == "" {
							i.Title = strings.TrimSpace(data)
						}
					})
					return nil
				}
			}
		}
	}
}

func init() {
	RegisterExtractor(feedExtractor{})
}
