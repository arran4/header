package recommended

import (
	"github.com/arran4/header/lint"
	"github.com/arran4/header/lint/rules/consistency"
	"github.com/arran4/header/lint/rules/html"
	"github.com/arran4/header/lint/rules/opengraph"
	"github.com/arran4/header/lint/rules/twitter"
)

func init() {
	lint.RegisterList("recommended", []string{
		consistency.RuleID,
		html.RuleID,
		opengraph.RuleID,
		twitter.RuleID,
	})
}
