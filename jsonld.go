package header

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
)

type jsonldExtractor struct{}

func (e jsonldExtractor) Name() string { return "JSON-LD" }

func (e jsonldExtractor) Init(info *PageInfo) Processor {
	return &jsonldProcessor{info: info}
}

type jsonldProcessor struct {
	info     *PageInfo
	inScript bool
}

func (p *jsonldProcessor) ProcessToken(t html.Token) bool {
	if t.Type == html.StartTagToken {
		if t.Data == "script" {
			for _, attr := range t.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					p.inScript = true
					break
				}
			}
		} else if t.Data == "body" {
			return true // stop
		}
	} else if t.Type == html.TextToken && p.inScript {
		content := strings.TrimSpace(t.Data)
		if strings.HasPrefix(content, "[") {
			var data []map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err == nil && len(data) > 0 {
				extractDataFromMap(data[0], p.info)
			}
		} else {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err == nil {
				extractDataFromMap(data, p.info)
			}
		}
	} else if t.Type == html.EndTagToken {
		if t.Data == "script" {
			p.inScript = false
		}
	}
	return false
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
