package header

import (
	"encoding/json"
	"strings"
	"golang.org/x/net/html"
)

// extractJSONLD searches for application/ld+json scripts and extracts data into PageInfo.
func extractJSONLD(doc *html.Node, info *PageInfo) {
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

			if isJSONLD {
				var b strings.Builder
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						b.WriteString(c.Data)
					}
				}

				// Simple best-effort unmarshaling
				// JSON-LD can be an object or an array of objects
				content := strings.TrimSpace(b.String())
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

	walk(doc)
}

func extractDataFromMap(data map[string]interface{}, info *PageInfo) {
	if t, ok := data["@type"].(string); ok && info.JSONLDType == "" {
		info.JSONLDType = t
	}
	if n, ok := data["name"].(string); ok && info.JSONLDName == "" {
		info.JSONLDName = n
	}
	if t, ok := data["headline"].(string); ok && info.Title == "" {
		info.Title = t
	}
}
