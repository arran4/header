package opengraph

import (
	"github.com/arran4/header"
	"testing"
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
