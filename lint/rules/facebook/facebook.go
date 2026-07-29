package facebook

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "facebook"

// facebookRule is a lint rule placeholder for checking Facebook-specific metadata.
// What it checks:
// - Currently acts as a placeholder for future Facebook-specific tag validations (e.g., `fb:app_id`, `fb:pages`, or `article:author`).
//
// Why it checks this:
// - While Facebook heavily utilizes OpenGraph tags (`og:*`), they also maintain proprietary tags that enable insights and deeper platform integration.
//
// References:
// - Meta for Developers - Sharing for Web: https://developers.facebook.com/docs/sharing/webmasters/
type facebookRule struct{}

func (r facebookRule) ID() string   { return RuleID }
func (r facebookRule) Name() string { return "Facebook Specific Checks" }

func (r facebookRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	// Basic facebook rules - often they want fb:app_id or similar.
	// We'll just put a placeholder here that passes unless something specific is found.
	return results
}

func init() {
	lint.RegisterRule(facebookRule{})
}
