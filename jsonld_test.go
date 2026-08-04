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
	if info.JSONLDID != "https://example.com/article/1" {
		t.Errorf("Expected JSONLDID 'https://example.com/article/1', got '%s'", info.JSONLDID)
	}
	if info.JSONLDDescription != "JSON-LD Test Description" {
		t.Errorf("Expected JSONLDDescription 'JSON-LD Test Description', got '%s'", info.JSONLDDescription)
	}
	if info.JSONLDUploadDate != "2023-10-24T23:57:33-07:00" {
		t.Errorf("Expected JSONLDUploadDate '2023-10-24T23:57:33-07:00', got '%s'", info.JSONLDUploadDate)
	}
	if info.JSONLDDuration != "PT3M33S" {
		t.Errorf("Expected JSONLDDuration 'PT3M33S', got '%s'", info.JSONLDDuration)
	}
	if info.JSONLDThumbnailURL != "https://example.com/thumbnail.jpg" {
		t.Errorf("Expected JSONLDThumbnailURL 'https://example.com/thumbnail.jpg', got '%s'", info.JSONLDThumbnailURL)
	}
	if info.JSONLDAuthor != "John Doe" {
		t.Errorf("Expected JSONLDAuthor 'John Doe', got '%s'", info.JSONLDAuthor)
	}
	if info.JSONLDPublisher != "Example News" {
		t.Errorf("Expected JSONLDPublisher 'Example News', got '%s'", info.JSONLDPublisher)
	}
}
