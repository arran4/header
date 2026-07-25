package all

import (
	"github.com/arran4/header/lint/lists"
	"github.com/arran4/header/lint/rules/consistency"
	"github.com/arran4/header/lint/rules/facebook"
	"github.com/arran4/header/lint/rules/html"
	"github.com/arran4/header/lint/rules/opengraph"
	"github.com/arran4/header/lint/rules/twitter"
	"github.com/arran4/header/lint/rules/urls"
)

func init() {
	lists.Register("all", []string{
		consistency.RuleID,
		facebook.RuleID,
		html.RuleID,
		opengraph.RuleID,
		twitter.RuleID,
		urls.RuleID,
	})

	// Set default list mapping
	lists.Register("default", []string{
		consistency.RuleID,
		html.RuleID,
		opengraph.RuleID,
		twitter.RuleID,
		urls.RuleID,
	})
}
