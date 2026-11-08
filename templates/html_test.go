package templates_test

import (
	"strings"
	"testing"

	"github.com/arran4/header"
	"github.com/arran4/header/templates"
)

func renderHTML(t *testing.T, info *header.PageInfo) string {
	t.Helper()
	tmpl, err := templates.LoadTemplate("html", "")
	if err != nil {
		t.Fatalf("LoadTemplate('html', '') failed: %v", err)
	}

	out, err := templates.Render(tmpl, info)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out
}

func TestHTMLTemplate_DeterministicStructure(t *testing.T) {
	info := &header.PageInfo{
		Title:              "Example title",
		Description:        "Example description",
		OGTitle:            "Example title",
		OGDescription:      "Example description",
		TwitterTitle:       "Example title",
		CanonicalLink:      "https://example.com/",
	}

	out := renderHTML(t, info)

	expected := `<head>
  <title>Example title</title>
  <meta name="description" content="Example description">
  <link rel="canonical" href="https://example.com/">
  <meta property="og:title" content="Example title">
  <meta property="og:description" content="Example description">
  <meta name="twitter:title" content="Example title">
</head>
`
	if out != expected {
		t.Errorf("mismatch.\nExpected:\n%s\nGot:\n%s", expected, out)
	}
}

func TestHTMLTemplate_EscapingAndInjection(t *testing.T) {
	info := &header.PageInfo{
		Title:         `<script>alert("XSS & title")</script>`,
		Description:   `" onclick="alert('inject')" <tag>`,
		OGTitle:       `" onmouseover="bad()" & <test>`,
		CanonicalLink: `https://example.com/?a=1&b=2" onclick="bad()`,
	}

	out := renderHTML(t, info)

	// Verify unescaped markup / script tags are NOT present
	if strings.Contains(out, "<script>") || strings.Contains(out, "</script>") {
		t.Errorf("unescaped script tag found in output: %s", out)
	}
	if strings.Contains(out, `" onclick=`) || strings.Contains(out, `" onmouseover=`) {
		t.Errorf("unescaped attribute injection found in output: %s", out)
	}

	// Verify proper HTML escaping
	if !strings.Contains(out, "&lt;script&gt;alert(&#34;XSS &amp; title&#34;)&lt;/script&gt;") {
		t.Errorf("expected escaped title, got: %s", out)
	}
	if !strings.Contains(out, "&#34; onclick=&#34;alert(&#39;inject&#39;)&#34; &lt;tag&gt;") {
		t.Errorf("expected escaped description attribute, got: %s", out)
	}
	if !strings.Contains(out, "https://example.com/?a=1&amp;b=2&#34; onclick=&#34;bad()") {
		t.Errorf("expected escaped canonical link, got: %s", out)
	}
}

func TestHTMLTemplate_PreserveDistinctSources(t *testing.T) {
	// Grouped text view coalesces these, but synthetic HTML MUST preserve distinct tags
	info := &header.PageInfo{
		Title:        "Same Title",
		OGTitle:      "Same Title",
		TwitterTitle: "Same Title",
	}

	out := renderHTML(t, info)

	if !strings.Contains(out, "<title>Same Title</title>") {
		t.Errorf("missing HTML title: %s", out)
	}
	if !strings.Contains(out, `<meta property="og:title" content="Same Title">`) {
		t.Errorf("missing og:title: %s", out)
	}
	if !strings.Contains(out, `<meta name="twitter:title" content="Same Title">`) {
		t.Errorf("missing twitter:title: %s", out)
	}
}

func TestHTMLTemplate_MissingFields(t *testing.T) {
	info := &header.PageInfo{
		URL: "https://example.com/only-url",
	}

	out := renderHTML(t, info)

	// No fabricated tags
	expected := "<head>\n</head>\n"
	if out != expected {
		t.Errorf("expected empty <head>, got: %q", out)
	}
}

func TestHTMLTemplate_LinksAndFeeds(t *testing.T) {
	info := &header.PageInfo{
		CanonicalLink: "https://example.com/",
		ShortLink:     "https://example.com/s",
		Icons: []header.Icon{
			{URL: "https://example.com/favicon.png", Sizes: "32x32"},
		},
		FeedLinks: []header.FeedLink{
			{URL: "https://example.com/rss.xml", Type: "application/rss+xml", Title: "Main Feed"},
		},
		AlternateLinks: []header.AlternateLink{
			{URL: "https://m.example.com/", Media: "handheld"},
		},
	}

	out := renderHTML(t, info)

	if !strings.Contains(out, `<link rel="canonical" href="https://example.com/">`) {
		t.Errorf("missing canonical link: %s", out)
	}
	if !strings.Contains(out, `<link rel="shortlink" href="https://example.com/s">`) {
		t.Errorf("missing shortlink: %s", out)
	}
	if !strings.Contains(out, `<link rel="icon" href="https://example.com/favicon.png" sizes="32x32">`) {
		t.Errorf("missing icon: %s", out)
	}
	if !strings.Contains(out, `<link rel="alternate" type="application/rss+xml" href="https://example.com/rss.xml" title="Main Feed">`) {
		t.Errorf("missing feed link: %s", out)
	}
	if !strings.Contains(out, `<link rel="alternate" href="https://m.example.com/" media="handheld">`) {
		t.Errorf("missing alternate link: %s", out)
	}
}

func TestHTMLTemplate_AppLinksAndUnknown(t *testing.T) {
	info := &header.PageInfo{
		AppLinksIOSName: "iOS App",
		AppLinksIOSURL:  "app://ios",
		UnknownMeta: []header.UnknownMeta{
			{Key: "custom-token", Value: "token-123"},
		},
	}

	out := renderHTML(t, info)

	if !strings.Contains(out, `<meta property="al:ios:app_name" content="iOS App">`) {
		t.Errorf("missing AppLinks tag: %s", out)
	}
	if !strings.Contains(out, `<meta name="custom-token" content="token-123">`) {
		t.Errorf("missing unknown meta tag: %s", out)
	}
}

func TestHTMLTemplate_DuplicatesAndMalformed(t *testing.T) {
	info := &header.PageInfo{
		Title: "First Title",
		UnknownMeta: []header.UnknownMeta{
			{Key: "duplicate-tag", Value: "val1"},
			{Key: "duplicate-tag", Value: "val2"},
			{Key: `bad"name<tag>`, Value: `bad"value'quote`},
		},
	}

	out := renderHTML(t, info)

	// Both duplicate entries should appear deterministically
	if strings.Count(out, `<meta name="duplicate-tag"`) != 2 {
		t.Errorf("expected 2 duplicate-tag entries, got: %s", out)
	}
	if !strings.Contains(out, `<meta name="duplicate-tag" content="val1">`) ||
		!strings.Contains(out, `<meta name="duplicate-tag" content="val2">`) {
		t.Errorf("missing duplicate tag values: %s", out)
	}

	// Malformed keys and values must be escaped
	if strings.Contains(out, "<tag>") || strings.Contains(out, `bad"name`) {
		t.Errorf("unescaped malformed key: %s", out)
	}
	if !strings.Contains(out, `bad&#34;name&lt;tag&gt;`) || !strings.Contains(out, `bad&#34;value&#39;quote`) {
		t.Errorf("expected escaped malformed key and value, got: %s", out)
	}
}

