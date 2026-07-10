package header

import (
	"strings"
	"testing"
)

func TestExtractOpenGraph(t *testing.T) {
	htmlContent := `
	<html>
	<head>
		<title>Test Title</title>
		<meta property="og:title" content="Test OG Title" />
		<meta property="og:type" content="article" />
		<meta property="og:image" content="http://example.com/image.jpg" />
		<meta property="og:image:alt" content="A test image alt text" />
		<meta name="twitter:title" content="Test Twitter Title" />
		<meta name="twitter:card" content="summary_large_image" />
		<meta name="twitter:image" content="http://example.com/twitter.jpg" />
		<meta name="twitter:image:alt" content="A test twitter image alt text" />
		<link rel="canonical" href="https://example.com/canonical" />
		<link rel="shortcut icon" href="https://www.youtube.com/s/desktop/6740ecb6/img/favicon.ico" type="image/x-icon">
		<link rel="icon" href="https://www.youtube.com/s/desktop/6740ecb6/img/favicon_32x32.png" sizes="32x32">
		<link rel="icon" href="https://www.youtube.com/s/desktop/6740ecb6/img/favicon_48x48.png" sizes="48x48">
		<link rel="icon" href="https://www.youtube.com/s/desktop/6740ecb6/img/favicon_144x144.png" sizes="144x144">
		<link rel="alternate" media="handheld" href="https://m.youtube.com/">
	</head>
	<body></body>
	</html>
	`

	info := &PageInfo{}
	extractor := openGraphExtractor{}
	err := extractor.Extract(strings.NewReader(htmlContent), info)
	if err != nil {
		t.Fatalf("Extractor returned error: %v", err)
	}

	if info.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", info.Title)
	}
	if info.OGTitle != "Test OG Title" {
		t.Errorf("Expected OGTitle 'Test OG Title', got '%s'", info.OGTitle)
	}
	if info.TwitterTitle != "Test Twitter Title" {
		t.Errorf("Expected TwitterTitle 'Test Twitter Title', got '%s'", info.TwitterTitle)
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
	if info.OGImageAlt != "A test image alt text" {
		t.Errorf("Expected OGImageAlt 'A test image alt text', got '%s'", info.OGImageAlt)
	}
	if info.TwitterImageAlt != "A test twitter image alt text" {
		t.Errorf("Expected TwitterImageAlt 'A test twitter image alt text', got '%s'", info.TwitterImageAlt)
	}
	if info.CanonicalLink != "https://example.com/canonical" {
		t.Errorf("Expected CanonicalLink 'https://example.com/canonical', got '%s'", info.CanonicalLink)
	}
	if info.AlternateLink != "https://m.youtube.com/" {
		t.Errorf("Expected AlternateLink 'https://m.youtube.com/', got '%s'", info.AlternateLink)
	}
	if len(info.Icons) != 4 {
		t.Fatalf("Expected 4 Icons, got %d", len(info.Icons))
	}
	expectedIcons := []Icon{
		{URL: "https://www.youtube.com/s/desktop/6740ecb6/img/favicon.ico", Sizes: ""},
		{URL: "https://www.youtube.com/s/desktop/6740ecb6/img/favicon_32x32.png", Sizes: "32x32"},
		{URL: "https://www.youtube.com/s/desktop/6740ecb6/img/favicon_48x48.png", Sizes: "48x48"},
		{URL: "https://www.youtube.com/s/desktop/6740ecb6/img/favicon_144x144.png", Sizes: "144x144"},
	}
	for i, expected := range expectedIcons {
		if info.Icons[i] != expected {
			t.Errorf("Expected Icon[%d] '%v', got '%v'", i, expected, info.Icons[i])
		}
	}
}
