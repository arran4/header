package header

import (
	_ "embed"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

//go:embed testdata/jsonld.html
var jsonldHTML []byte

func TestExtractJSONLD(t *testing.T) {
	info := &PageInfo{}

	extractor := jsonldExtractor{}
	processor := extractor.Init(info)

	z := html.NewTokenizer(strings.NewReader(string(jsonldHTML)))
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		tk := z.Token()
		if processor.ProcessToken(tk) {
			break
		}
	}

	if info.JSONLDType != "NewsArticle" {
		t.Errorf("Expected JSONLDType 'NewsArticle', got '%s'", info.JSONLDType)
	}
	if info.Title != "JSON-LD Test Headline" {
		t.Errorf("Expected Title 'JSON-LD Test Headline', got '%s'", info.Title)
	}
	if info.JSONLDName != "JSON-LD Test Name" {
		t.Errorf("Expected JSONLDName 'JSON-LD Test Name', got '%s'", info.JSONLDName)
	}
}
