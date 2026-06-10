package header

import (
	"bytes"
	_ "embed"
	"testing"

	"golang.org/x/net/html"
)

//go:embed testdata/jsonld.html
var jsonldHTML []byte

func TestExtractJSONLD(t *testing.T) {
	doc, err := html.Parse(bytes.NewReader(jsonldHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	info := &PageInfo{}
	extractJSONLD(doc, info)

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
