package header

import (
	"bytes"
	"testing"
	"golang.org/x/net/html"
)

func TestExtractJSONLD(t *testing.T) {
	htmlContent := `
	<html>
	<head>
		<script type="application/ld+json">
		{
			"@context": "https://schema.org",
			"@type": "NewsArticle",
			"headline": "JSON-LD Test Headline",
			"name": "JSON-LD Test Name"
		}
		</script>
	</head>
	<body></body>
	</html>
	`

	doc, err := html.Parse(bytes.NewReader([]byte(htmlContent)))
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
