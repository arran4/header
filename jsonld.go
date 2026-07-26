package header

import (
	"encoding/json"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type jsonldExtractor struct{}

func (e jsonldExtractor) Name() string { return "JSON-LD" }

func (e jsonldExtractor) Extract(r io.Reader, info *PageInfo) error {
	z := html.NewTokenizer(r)

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
			if t.Data == "script" {
				var isJSONLD bool
				for _, attr := range t.Attr {
					if attr.Key == "type" && attr.Val == "application/ld+json" {
						isJSONLD = true
						break
					}
				}

				if isJSONLD {
					tt = z.Next()
					if tt == html.TextToken {
						content := strings.TrimSpace(z.Token().Data)
						if strings.HasPrefix(content, "[") {
							var data []map[string]interface{}
							if err := json.Unmarshal([]byte(content), &data); err == nil && len(data) > 0 {
								extractDataFromMap(data[0], info)
							}
						} else {
							var data map[string]interface{}
							if err := json.Unmarshal([]byte(content), &data); err == nil {
								extractDataFromMap(data, info)
							}
						}
					}
				}
			} else if t.Data == "body" {
				// Once we hit body, we can typically stop parsing for metadata
				return nil
			}
		}
	}
}

func extractDataFromMap(data map[string]interface{}, info *PageInfo) {
	info.Mutate(func(i *PageInfo) {
		if t, ok := data["@type"].(string); ok {
			if i.JSONLDType == "" {
				i.JSONLDType = t
			}
		}
		if n, ok := data["name"].(string); ok {
			if i.JSONLDName == "" {
				i.JSONLDName = n
			}
		}
		if t, ok := data["headline"].(string); ok {
			if i.Title == "" {
				i.Title = t
			}
		}
	})
}

func init() {
	RegisterExtractor(jsonldExtractor{})
}
