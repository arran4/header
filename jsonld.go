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

func extractPersonOrOrgName(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			return name
		}
	case []interface{}:
		if len(v) > 0 {
			return extractPersonOrOrgName(v[0])
		}
	}
	return ""
}

func extractDataFromMap(data map[string]interface{}, info *PageInfo) {
	info.Mutate(func(i *PageInfo) {
		if t, ok := data["@type"].(string); ok {
			if i.JSONLDType == "" {
				i.JSONLDType = t
			}
		}
		if id, ok := data["@id"].(string); ok {
			if i.JSONLDID == "" {
				i.JSONLDID = id
			}
		}
		if n, ok := data["name"].(string); ok {
			if i.JSONLDName == "" {
				i.JSONLDName = n
			}
		}
		if d, ok := data["description"].(string); ok {
			if i.JSONLDDescription == "" {
				i.JSONLDDescription = d
			}
		}

		var uploadDate string
		if d, ok := data["uploadDate"].(string); ok {
			uploadDate = d
		} else if d, ok := data["datePublished"].(string); ok {
			uploadDate = d
		} else if d, ok := data["dateModified"].(string); ok {
			uploadDate = d
		}
		if uploadDate != "" && i.JSONLDUploadDate == "" {
			i.JSONLDUploadDate = uploadDate
		}

		if d, ok := data["duration"].(string); ok {
			if i.JSONLDDuration == "" {
				i.JSONLDDuration = d
			}
		}
		if u, ok := data["thumbnailUrl"].(string); ok {
			if i.JSONLDThumbnailURL == "" {
				i.JSONLDThumbnailURL = u
			}
		}

		if a, ok := data["author"]; ok {
			if name := extractPersonOrOrgName(a); name != "" && i.JSONLDAuthor == "" {
				i.JSONLDAuthor = name
			}
		}
		if p, ok := data["publisher"]; ok {
			if name := extractPersonOrOrgName(p); name != "" && i.JSONLDPublisher == "" {
				i.JSONLDPublisher = name
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
