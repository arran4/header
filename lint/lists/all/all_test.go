package all

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
	"testing"
)

func TestAllRulesLoaded(_ *testing.T) {
	// Register will happen in init() of the imported packages
	// Create a dummy page info to trigger rules
	info := &header.PageInfo{}

	// Just verify it doesn't panic
	lint.Run(info, []string{"all"}, nil)
}
