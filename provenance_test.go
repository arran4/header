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
