package header

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/jsonld.html
var jsonldHTML []byte

func TestExtractJSONLD(t *testing.T) {
	info := &PageInfo{}
	extractor := jsonldExtractor{}
	err := extractor.Extract(strings.NewReader(string(jsonldHTML)), info)
	if err != nil {
		t.Fatalf("Extractor returned error: %v", err)
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
