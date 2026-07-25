package twitter

import (
	"github.com/arran4/header"
	"testing"
)

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
