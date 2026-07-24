package all

import (
	"testing"
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

func TestAllRulesLoaded(t *testing.T) {
	// Register will happen in init() of the imported packages
	// We can check if rules count is > 0

	// Create a dummy page info to trigger rules
	info := &header.PageInfo{}

	// Just verify it doesn't panic
	lint.Run(info, []string{"all"})
}
