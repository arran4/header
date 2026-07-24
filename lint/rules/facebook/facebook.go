package facebook

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

type facebookRule struct{}

func (r facebookRule) Name() string { return "Facebook Specific Checks" }

func (r facebookRule) Tags() []string { return []string{"facebook"} }

func (r facebookRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	// Basic facebook rules - often they want fb:app_id or similar.
	// We'll just put a placeholder here that passes unless something specific is found.
	return results
}

func init() {
	lint.Register(facebookRule{})
}
