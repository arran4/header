package templates_test

import (
	"strings"
	"testing"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/templates"
	"github.com/fatih/color"
)

func renderStyle(t *testing.T, style string, info *header.PageInfo) string {
	t.Helper()
	color.NoColor = true
	tmpl, err := templates.LoadTemplate(style, "")
	if err != nil {
		t.Fatalf("LoadTemplate(%q, '') failed: %v", style, err)
	}

	out, err := templates.Render(tmpl, info)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out
}

func TestDefaultTemplate_RichOverview(t *testing.T) {
	uploadDate, _ := time.Parse(time.RFC3339, "2026-05-01T12:00:00Z")
	info := &header.PageInfo{
		URL:           "https://example.com/blog/post?utm_source=rss",
		CanonicalLink: "https://example.com/blog/post",
		ShortLink:     "https://example.com/p/123",

		// Differing titles and descriptions
		Title:              "HTML Page Title",
		OGTitle:            "Social Title",
		TwitterTitle:       "Social Title",
		Description:        "HTML Page Description",
		OGDescription:      "Social Description",
		TwitterDescription: "Social Description",

		// Social sharing
		OGType:      "article",
		TwitterCard: "summary_large_image",

		// Cover images and metadata
		OGImage:          "https://example.com/og.jpg",
		TwitterImage:     "https://example.com/og.jpg",
		OGImageAlt:       "Article banner",
		TwitterImageAlt:  "Article banner",
		OGImageWidth:     "1200",
		OGImageHeight:    "630",
		OGImageType:      "image/jpeg",

		// Video
		OGVideo:          "https://example.com/clip.mp4",
		OGVideoType:      "video/mp4",
		OGVideoWidth:     "1920",
		OGVideoHeight:    "1080",

		// JSON-LD structured data summary
		JSONLDType:             "NewsArticle",
		JSONLDID:               "https://example.com/article/123",
		JSONLDAuthor:           "Alice Journalist",
		JSONLDPublisher:        "The Daily Tech",
		JSONLDUploadDate:       "2026-05-01T12:00:00Z",
		JSONLDUploadDateParsed: uploadDate,

		// Discovery links
		FeedLinks: []header.FeedLink{
			{URL: "https://example.com/feed.xml", Title: "Tech RSS", Type: "application/rss+xml"},
		},
		AlternateLinks: []header.AlternateLink{
			{URL: "https://m.example.com/blog/post", Media: "only screen and (max-width: 640px)"},
		},
		Icons: []header.Icon{
			{URL: "https://example.com/favicon.png", Sizes: "32x32"},
		},

		// App links
		AppLinksIOSName:       "DailyTech iOS",
		AppLinksIOSURL:        "dailytech://post/123",
		AppLinksIOSAppStoreID: "987654321",

		// Page metadata
		Keywords:   "tech, go, metadata",
		ThemeColor: "#007acc",
	}

	out := renderStyle(t, "default", info)

	// Verify sections exist
	expectedParts := []string{
		"URL: https://example.com/blog/post?utm_source=rss",
		"Canonical URL: https://example.com/blog/post",
		"Shortlink: https://example.com/p/123",
		"Title (HTML):",
		"HTML Page Title",
		"Title (OpenGraph, Twitter):",
		"Social Title",
		"Description (HTML):",
		"HTML Page Description",
		"Description (OpenGraph, Twitter):",
		"Social Description",
		"Social Sharing:",
		"OpenGraph Type: article",
		"Twitter Card: summary_large_image",
		"Cover Images:",
		"https://example.com/og.jpg (OpenGraph, Twitter)",
		"Image Alt (OpenGraph, Twitter): Article banner",
		"Image Dimensions: 1200x630 (image/jpeg)",
		"Video:",
		"OpenGraph Video: https://example.com/clip.mp4 (1920x1080, video/mp4)",
		"Structured Data (JSON-LD):",
		"Type: NewsArticle",
		"ID: https://example.com/article/123",
		"Author: Alice Journalist",
		"Publisher: The Daily Tech",
		"Discovery Links:",
		`Feed Link: https://example.com/feed.xml ("Tech RSS", application/rss+xml)`,
		"Alternate Link: https://m.example.com/blog/post (only screen and (max-width: 640px))",
		"Icon: https://example.com/favicon.png (32x32)",
		"App Links:",
		"iOS: DailyTech iOS (dailytech://post/123) [App Store ID: 987654321]",
		"Page Metadata:",
		"Theme Color: #007acc",
		"Keywords: tech, go, metadata",
	}

	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("missing expected part %q in default output:\n%s", part, out)
		}
	}

	// Compare with quick output to ensure default is significantly richer
	quickOut := renderStyle(t, "quick", info)
	if strings.Contains(quickOut, "Social Sharing:") {
		t.Errorf("quick output should not contain Social Sharing")
	}
	if strings.Contains(quickOut, "Structured Data (JSON-LD):") {
		t.Errorf("quick output should not contain Structured Data")
	}
	if strings.Contains(quickOut, "Discovery Links:") {
		t.Errorf("quick output should not contain Discovery Links")
	}
	if strings.Contains(quickOut, "App Links:") {
		t.Errorf("quick output should not contain App Links")
	}
	if strings.Contains(quickOut, "Page Metadata:") {
		t.Errorf("quick output should not contain Page Metadata")
	}
}

func TestDefaultTemplate_Sparse(t *testing.T) {
	info := &header.PageInfo{
		URL:   "https://example.com/sparse",
		Title: "Sparse Page Title",
	}

	out := renderStyle(t, "default", info)
	quickOut := renderStyle(t, "quick", info)

	// Sparse page should produce identical concise output in both default and quick
	expected := "URL: https://example.com/sparse\n\nTitle (HTML):\n  Sparse Page Title\n"
	if out != expected {
		t.Errorf("sparse default output mismatch.\nExpected:\n%q\nGot:\n%q", expected, out)
	}
	if quickOut != expected {
		t.Errorf("sparse quick output mismatch.\nExpected:\n%q\nGot:\n%q", expected, quickOut)
	}

	// Ensure no empty headings
	forbidden := []string{
		"Canonical URL:",
		"Shortlink:",
		"Description",
		"Social Sharing:",
		"Cover Images:",
		"Video:",
		"Structured Data",
		"Discovery Links:",
		"App Links:",
		"Page Metadata:",
	}
	for _, term := range forbidden {
		if strings.Contains(out, term) {
			t.Errorf("sparse output should not contain %q: %s", term, out)
		}
	}
}

func TestQuickTemplate_ReproducesPreChangeFormat(t *testing.T) {
	info := &header.PageInfo{
		URL:                "https://example.com/",
		Title:              "Title 1",
		OGTitle:            "Title 1",
		TwitterTitle:       "Title 1",
		Description:        "Desc 1",
		OGDescription:      "Desc 1",
		OGImage:            "https://example.com/img.jpg",
		TwitterImage:       "https://example.com/img.jpg",
		OGType:             "website",
		TwitterCard:        "summary",
		FeedLinks: []header.FeedLink{
			{URL: "https://example.com/rss.xml"},
		},
	}

	quickOut := renderStyle(t, "quick", info)

	// Quick should only contain URL, Title, Description, and Cover Images
	expected := `URL: https://example.com/

Title (HTML, OpenGraph, Twitter):
  Title 1

Description (HTML, OpenGraph):
  Desc 1

Cover Images:
  https://example.com/img.jpg (OpenGraph, Twitter)
`
	if quickOut != expected {
		t.Errorf("quick output mismatch.\nExpected:\n%s\nGot:\n%s", expected, quickOut)
	}
}
