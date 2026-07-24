package consistency

import (
	"testing"
	"github.com/arran4/header"
)

func TestConsistencyRule(t *testing.T) {
	rule := consistencyRule{}

	info := &header.PageInfo{
		OGTitle: "Title A",
		TwitterTitle: "Title B",
	}
	results := rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 info for differing titles, got %d", len(results))
	}
}
