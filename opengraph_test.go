package header

import (
	"bytes"
	"testing"
	"golang.org/x/net/html"
)

func TestExtractOpenGraph(t *testing.T) {
	htmlContent := `
	<html>
	<head>
		<title>Test Title</title>
		<meta property="og:type" content="article" />
		<meta property="og:image" content="http://example.com/image.jpg" />
		<meta name="twitter:card" content="summary_large_image" />
		<meta name="twitter:image" content="http://example.com/twitter.jpg" />
	</head>
	<body></body>
	</html>
	`

	doc, err := html.Parse(bytes.NewReader([]byte(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	info := &PageInfo{}
	extractOpenGraph(doc, info)

	if info.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", info.Title)
	}
	if info.OGType != "article" {
		t.Errorf("Expected OGType 'article', got '%s'", info.OGType)
	}
	if info.OGImage != "http://example.com/image.jpg" {
		t.Errorf("Expected OGImage 'http://example.com/image.jpg', got '%s'", info.OGImage)
	}
	if info.TwitterCard != "summary_large_image" {
		t.Errorf("Expected TwitterCard 'summary_large_image', got '%s'", info.TwitterCard)
	}
	if info.TwitterImage != "http://example.com/twitter.jpg" {
		t.Errorf("Expected TwitterImage 'http://example.com/twitter.jpg', got '%s'", info.TwitterImage)
	}
}
