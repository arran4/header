package header

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/arran4/header/templates"
	"github.com/fatih/color"
	"golang.org/x/net/html"
)

func renderDefault(t *testing.T, info *PageInfo) string {
	t.Helper()
	color.NoColor = true
	funcMap := template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
	}

	tmplContent, err := templates.FS.ReadFile("default.tmpl")
	if err != nil {
		t.Fatalf("failed to read default.tmpl: %v", err)
	}

	tmpl, err := template.New("default.tmpl").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		t.Fatalf("failed to parse default.tmpl: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		t.Fatalf("failed to execute default.tmpl: %v", err)
	}
	return buf.String()
}

func extractHTML(htmlContent string) *PageInfo {
	info := &PageInfo{
		URL: "https://example.com/",
	}
	processors := make([]Processor, len(extractorRegistry))
	active := make([]bool, len(extractorRegistry))
	activeCount := len(extractorRegistry)
	for i, ext := range extractorRegistry {
		processors[i] = ext.Init(info)
		active[i] = true
	}
	z := html.NewTokenizer(strings.NewReader(htmlContent))
	for activeCount > 0 {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		tk := z.Token()
		for i, p := range processors {
			if active[i] {
				if p.ProcessToken(tk) {
					active[i] = false
					activeCount--
				}
			}
		}
	}
	return info
}

func TestProvenance_IdenticalMetadata(t *testing.T) {
	info := &PageInfo{
		URL:                "https://example.com/",
		Title:              "Example title",
		HTMLTitle:          "Example title",
		OGTitle:            "Example title",
		TwitterTitle:       "Example title",
		OGDescription:      "Example description",
		TwitterDescription: "Example description",
		OGImage:            "https://example.com/cover.jpg",
		TwitterImage:       "https://example.com/cover.jpg",
	}

	out := renderDefault(t, info)

	expected := `URL: https://example.com/

Title (HTML, OpenGraph, Twitter):
  Example title

Description (OpenGraph, Twitter):
  Example description

Cover Images:
  https://example.com/cover.jpg (OpenGraph, Twitter)
`
	if out != expected {
		t.Errorf("mismatch.\nExpected:\n%s\nGot:\n%s", expected, out)
	}
}

func TestProvenance_ConflictingMetadata(t *testing.T) {
	info := &PageInfo{
		URL:                "https://example.com/",
		Title:              "HTML Title",
		OGTitle:            "OG/Twitter Title",
		TwitterTitle:       "OG/Twitter Title",
		Description:        "HTML Description",
		OGDescription:      "OG Description",
		TwitterDescription: "Twitter Description",
		OGImage:            "https://example.com/og.jpg",
		TwitterImage:       "https://example.com/tw.jpg",
	}

	out := renderDefault(t, info)

	expected := `URL: https://example.com/

Title (HTML):
  HTML Title

Title (OpenGraph, Twitter):
  OG/Twitter Title

Description (HTML):
  HTML Description

Description (OpenGraph):
  OG Description

Description (Twitter):
  Twitter Description

Cover Images:
  https://example.com/og.jpg (OpenGraph)
  https://example.com/tw.jpg (Twitter)
`
	if out != expected {
		t.Errorf("mismatch.\nExpected:\n%s\nGot:\n%s", expected, out)
	}
}

func TestProvenance_MissingValues(t *testing.T) {
	t.Run("only Twitter metadata", func(t *testing.T) {
		info := &PageInfo{
			URL:                "https://example.com/",
			TwitterTitle:       "Twitter Title Only",
			TwitterDescription: "Twitter Desc Only",
		}

		out := renderDefault(t, info)

		// Must not claim HTML exists
		if strings.Contains(out, "HTML") {
			t.Errorf("output should not claim HTML exists: %s", out)
		}

		expected := `URL: https://example.com/

Title (Twitter):
  Twitter Title Only

Description (Twitter):
  Twitter Desc Only
`
		if out != expected {
			t.Errorf("mismatch.\nExpected:\n%s\nGot:\n%s", expected, out)
		}
	})

	t.Run("no title, no description, no images", func(t *testing.T) {
		info := &PageInfo{
			URL: "https://example.com/empty",
		}

		out := renderDefault(t, info)
		expected := "URL: https://example.com/empty\n"
		if out != expected {
			t.Errorf("mismatch.\nExpected:\n%q\nGot:\n%q", expected, out)
		}
	})
}

func TestProvenance_SourceOrderingAndDeduplication(t *testing.T) {
	info := &PageInfo{
		URL:              "https://example.com/",
		Title:            "Shared Title",
		HTMLTitle:        "Shared Title",
		OGTitle:          "Shared Title",
		TwitterTitle:     "Shared Title",
		JSONLDName:       "Shared Title",
		OGImage:          "https://example.com/cover.jpg",
		OGImageSecureURL: "https://example.com/cover.jpg", // duplicate OpenGraph source for same URL
		TwitterImage:     "https://example.com/cover.jpg",
	}

	out := renderDefault(t, info)

	// Verify title source order: HTML, OpenGraph, Twitter, JSON-LD
	if !strings.Contains(out, "Title (HTML, OpenGraph, Twitter, JSON-LD):") {
		t.Errorf("expected Title sources sorted in canonical order, got: %s", out)
	}

	// Verify cover image deduplication: OpenGraph appears once, Twitter appears once
	if !strings.Contains(out, "https://example.com/cover.jpg (OpenGraph, Twitter)") {
		t.Errorf("expected deduplicated cover image sources, got: %s", out)
	}
}

func TestProvenance_SemanticFieldSeparation(t *testing.T) {
	info := &PageInfo{
		URL:           "https://example.com/",
		Title:         "Identical Text",
		OGDescription: "Identical Text",
	}

	out := renderDefault(t, info)

	// Title and Description must remain separate sections despite identical string values
	if !strings.Contains(out, "Title (HTML):") || !strings.Contains(out, "Description (OpenGraph):") {
		t.Errorf("different semantic fields should not be collapsed: %s", out)
	}
}

func TestProvenance_HTMLExtraction(t *testing.T) {
	htmlData := `<!DOCTYPE html>
<html>
<head>
  <title>Extracted Title</title>
  <meta name="description" content="Extracted HTML Description">
  <meta property="og:title" content="Extracted Title">
  <meta property="og:description" content="Extracted OG Description">
</head>
<body></body>
</html>`

	info := extractHTML(htmlData)

	if info.Title != "Extracted Title" {
		t.Errorf("expected HTML Title 'Extracted Title', got %q", info.Title)
	}
	if info.Description != "Extracted HTML Description" {
		t.Errorf("expected HTML Description 'Extracted HTML Description', got %q", info.Description)
	}

	out := renderDefault(t, info)
	if !strings.Contains(out, "Title (HTML, OpenGraph):") {
		t.Errorf("expected Title grouped across HTML and OpenGraph, got: %s", out)
	}
	if !strings.Contains(out, "Description (HTML):") || !strings.Contains(out, "Description (OpenGraph):") {
		t.Errorf("expected Descriptions separated since they differ, got: %s", out)
	}
}

func TestProvenance_TitleProvenanceAndPrecedence(t *testing.T) {
	t.Run("OG-only title accurately tracked without HTML misattribution", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><meta property="og:title" content="OpenGraph Only Title"></head><body></body></html>`
		info := extractHTML(html)

		if info.HTMLTitle != "" {
			t.Errorf("expected empty HTMLTitle, got %q", info.HTMLTitle)
		}
		if info.OGTitle != "OpenGraph Only Title" {
			t.Errorf("expected OGTitle 'OpenGraph Only Title', got %q", info.OGTitle)
		}
		if info.Title != "OpenGraph Only Title" {
			t.Errorf("expected fallback Title 'OpenGraph Only Title', got %q", info.Title)
		}

		grouped := info.GroupedTitles()
		if len(grouped) != 1 {
			t.Fatalf("expected 1 grouped title, got %d: %+v", len(grouped), grouped)
		}
		if grouped[0].SourcesString() != "OpenGraph" || grouped[0].Value != "OpenGraph Only Title" {
			t.Errorf("unexpected grouped title: %+v", grouped[0])
		}

		out := renderDefault(t, info)
		if !strings.Contains(out, "Title (OpenGraph):") {
			t.Errorf("expected 'Title (OpenGraph):' in default output: %s", out)
		}
		if strings.Contains(out, "Title (HTML") {
			t.Errorf("output must not attribute title to HTML: %s", out)
		}
	})

	t.Run("Twitter-only title accurately tracked without HTML misattribution", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><meta name="twitter:title" content="Twitter Only Title"></head><body></body></html>`
		info := extractHTML(html)

		if info.HTMLTitle != "" {
			t.Errorf("expected empty HTMLTitle, got %q", info.HTMLTitle)
		}
		if info.TwitterTitle != "Twitter Only Title" {
			t.Errorf("expected TwitterTitle 'Twitter Only Title', got %q", info.TwitterTitle)
		}
		if info.Title != "Twitter Only Title" {
			t.Errorf("expected fallback Title 'Twitter Only Title', got %q", info.Title)
		}

		grouped := info.GroupedTitles()
		if len(grouped) != 1 {
			t.Fatalf("expected 1 grouped title, got %d: %+v", len(grouped), grouped)
		}
		if grouped[0].SourcesString() != "Twitter" || grouped[0].Value != "Twitter Only Title" {
			t.Errorf("unexpected grouped title: %+v", grouped[0])
		}

		out := renderDefault(t, info)
		if !strings.Contains(out, "Title (Twitter):") {
			t.Errorf("expected 'Title (Twitter):' in default output: %s", out)
		}
		if strings.Contains(out, "Title (HTML") {
			t.Errorf("output must not attribute title to HTML: %s", out)
		}
	})

	t.Run("HTML-only title accurately tracked", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><title>HTML Only Title</title></head><body></body></html>`
		info := extractHTML(html)

		if info.HTMLTitle != "HTML Only Title" {
			t.Errorf("expected HTMLTitle 'HTML Only Title', got %q", info.HTMLTitle)
		}
		if info.Title != "HTML Only Title" {
			t.Errorf("expected Title 'HTML Only Title', got %q", info.Title)
		}

		grouped := info.GroupedTitles()
		if len(grouped) != 1 {
			t.Fatalf("expected 1 grouped title, got %d: %+v", len(grouped), grouped)
		}
		if grouped[0].SourcesString() != "HTML" || grouped[0].Value != "HTML Only Title" {
			t.Errorf("unexpected grouped title: %+v", grouped[0])
		}

		out := renderDefault(t, info)
		if !strings.Contains(out, "Title (HTML):") {
			t.Errorf("expected 'Title (HTML):' in default output: %s", out)
		}
	})

	t.Run("Conflicting titles preserve HTML precedence and distinct sources", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head>
			<title>Original HTML Title</title>
			<meta property="og:title" content="OpenGraph Title">
			<meta name="twitter:title" content="Twitter Title">
		</head><body></body></html>`
		info := extractHTML(html)

		if info.HTMLTitle != "Original HTML Title" {
			t.Errorf("expected HTMLTitle 'Original HTML Title', got %q", info.HTMLTitle)
		}
		if info.OGTitle != "OpenGraph Title" {
			t.Errorf("expected OGTitle 'OpenGraph Title', got %q", info.OGTitle)
		}
		if info.TwitterTitle != "Twitter Title" {
			t.Errorf("expected TwitterTitle 'Twitter Title', got %q", info.TwitterTitle)
		}
		if info.Title != "Original HTML Title" {
			t.Errorf("expected preferred Title to be HTML title 'Original HTML Title', got %q", info.Title)
		}

		grouped := info.GroupedTitles()
		if len(grouped) != 3 {
			t.Fatalf("expected 3 distinct grouped titles, got %d: %+v", len(grouped), grouped)
		}

		out := renderDefault(t, info)
		if !strings.Contains(out, "Title (HTML):") || !strings.Contains(out, "Title (OpenGraph):") || !strings.Contains(out, "Title (Twitter):") {
			t.Errorf("expected distinct title sections for each source: %s", out)
		}
	})

	t.Run("Matching titles across sources coalesced correctly", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head>
			<title>Uniform Title</title>
			<meta property="og:title" content="Uniform Title">
			<meta name="twitter:title" content="Uniform Title">
		</head><body></body></html>`
		info := extractHTML(html)

		if info.HTMLTitle != "Uniform Title" || info.OGTitle != "Uniform Title" || info.TwitterTitle != "Uniform Title" {
			t.Errorf("expected matching titles across sources, got HTML=%q, OG=%q, TW=%q", info.HTMLTitle, info.OGTitle, info.TwitterTitle)
		}
		if info.Title != "Uniform Title" {
			t.Errorf("expected Title 'Uniform Title', got %q", info.Title)
		}

		grouped := info.GroupedTitles()
		if len(grouped) != 1 {
			t.Fatalf("expected 1 grouped title, got %d: %+v", len(grouped), grouped)
		}
		if grouped[0].SourcesString() != "HTML, OpenGraph, Twitter" || grouped[0].Value != "Uniform Title" {
			t.Errorf("unexpected grouped title: %+v", grouped[0])
		}

		out := renderDefault(t, info)
		if !strings.Contains(out, "Title (HTML, OpenGraph, Twitter):") {
			t.Errorf("expected grouped Title (HTML, OpenGraph, Twitter): %s", out)
		}
	})

	t.Run("Tag order does not affect title precedence or provenance", func(t *testing.T) {
		// Order 1: OG before HTML
		html1 := `<!DOCTYPE html><html><head>
			<meta property="og:title" content="OpenGraph Title">
			<title>True HTML Title</title>
			<meta name="twitter:title" content="Twitter Title">
		</head><body></body></html>`

		// Order 2: Twitter before HTML before OG
		html2 := `<!DOCTYPE html><html><head>
			<meta name="twitter:title" content="Twitter Title">
			<title>True HTML Title</title>
			<meta property="og:title" content="OpenGraph Title">
		</head><body></body></html>`

		// Order 3: HTML first
		html3 := `<!DOCTYPE html><html><head>
			<title>True HTML Title</title>
			<meta property="og:title" content="OpenGraph Title">
			<meta name="twitter:title" content="Twitter Title">
		</head><body></body></html>`

		info1 := extractHTML(html1)
		info2 := extractHTML(html2)
		info3 := extractHTML(html3)

		for idx, info := range []*PageInfo{info1, info2, info3} {
			if info.Title != "True HTML Title" {
				t.Errorf("order %d: expected Title 'True HTML Title', got %q", idx+1, info.Title)
			}
			if info.HTMLTitle != "True HTML Title" {
				t.Errorf("order %d: expected HTMLTitle 'True HTML Title', got %q", idx+1, info.HTMLTitle)
			}
			if info.OGTitle != "OpenGraph Title" {
				t.Errorf("order %d: expected OGTitle 'OpenGraph Title', got %q", idx+1, info.OGTitle)
			}
			if info.TwitterTitle != "Twitter Title" {
				t.Errorf("order %d: expected TwitterTitle 'Twitter Title', got %q", idx+1, info.TwitterTitle)
			}
		}

		// When HTML title is absent, OG takes precedence over Twitter regardless of order
		ogBeforeTwitter := `<!DOCTYPE html><html><head>
			<meta property="og:title" content="Preferred OG Title">
			<meta name="twitter:title" content="Fallback Twitter Title">
		</head><body></body></html>`
		twitterBeforeOG := `<!DOCTYPE html><html><head>
			<meta name="twitter:title" content="Fallback Twitter Title">
			<meta property="og:title" content="Preferred OG Title">
		</head><body></body></html>`

		infoOGFirst := extractHTML(ogBeforeTwitter)
		infoTWFirst := extractHTML(twitterBeforeOG)

		if infoOGFirst.Title != "Preferred OG Title" {
			t.Errorf("expected OG title precedence, got %q", infoOGFirst.Title)
		}
		if infoTWFirst.Title != "Preferred OG Title" {
			t.Errorf("expected OG title precedence even when Twitter tag comes first, got %q", infoTWFirst.Title)
		}
		if infoOGFirst.HTMLTitle != "" || infoTWFirst.HTMLTitle != "" {
			t.Errorf("HTMLTitle must remain empty when no <title> tag exists")
		}
	})
}
