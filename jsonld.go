package header

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
)

type jsonldExtractor struct{}

func (e jsonldExtractor) Name() string { return "JSON-LD" }

func (e jsonldExtractor) Extract(n *html.Node, info *PageInfo) error {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			var isJSONLD bool
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					isJSONLD = true
					break
				}
			}

			if isJSONLD && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				content := strings.TrimSpace(n.FirstChild.Data)
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

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return nil
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
