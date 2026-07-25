package html

import (
	"github.com/arran4/header"
	"testing"
)

func TestHTMLRule(t *testing.T) {
	rule := htmlRule{}

	info := &header.PageInfo{}
	results := rule.Check(info)
	if len(results) != 1 {
		t.Errorf("Expected 1 warning for missing title, got %d", len(results))
	}
}
