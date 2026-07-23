package lint

import (
	"testing"

	"github.com/arran4/header"
)

func TestOpenGraphRule(t *testing.T) {
	rule := opengraphRule{}

	info := &header.PageInfo{} // empty
	results := rule.Check(info)

	if len(results) != 3 {
		t.Errorf("Expected 3 warnings for empty OpenGraph data, got %d", len(results))
	}

	info = &header.PageInfo{
		OGTitle: "Test",
		OGType:  "website",
		OGImage: "https://example.com/image.jpg",
	}
	results = rule.Check(info)

	if len(results) != 0 {
		t.Errorf("Expected 0 warnings for valid OpenGraph data, got %d", len(results))
	}
}

func TestTwitterRule(t *testing.T) {
	rule := twitterRule{}

	info := &header.PageInfo{
		TwitterCard: "invalid",
	}
	results := rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 error for invalid card, got %d", len(results))
	}

	info = &header.PageInfo{
		TwitterCard: "summary_large_image",
	}
	results = rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 error for missing image, got %d", len(results))
	}

	info = &header.PageInfo{
		TwitterCard:  "summary_large_image",
		TwitterImage: "https://example.com/image.jpg",
	}
	results = rule.Check(info)
	if len(results) != 0 {
		t.Errorf("Expected 0 errors for valid large image card, got %d", len(results))
	}
}

func TestConsistencyRule(t *testing.T) {
	rule := consistencyRule{}

	info := &header.PageInfo{
		OGTitle:      "Title A",
		TwitterTitle: "Title B",
	}
	results := rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 info for differing titles, got %d", len(results))
	}
}

func TestHTMLRule(t *testing.T) {
	rule := htmlRule{}

	info := &header.PageInfo{}
	results := rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 warning for missing title, got %d", len(results))
	}
}
