package facebook

import (
	"testing"
	"github.com/arran4/header"
)

func TestFacebookRule(t *testing.T) {
	rule := facebookRule{}

	info := &header.PageInfo{} // empty
	results := rule.Check(info)

	if len(results) != 0 {
		t.Errorf("Expected 0 warnings for empty facebook data, got %d", len(results))
	}
}
