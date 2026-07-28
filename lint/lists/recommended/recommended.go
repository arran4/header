package recommended

import (
	"github.com/arran4/header/lint/lists"
	"github.com/arran4/header/lint/rules/consistency"
	"github.com/arran4/header/lint/rules/html"
	"github.com/arran4/header/lint/rules/images"
	"github.com/arran4/header/lint/rules/opengraph"
	"github.com/arran4/header/lint/rules/twitter"
)

func init() {
	lists.Register("recommended", []string{
		consistency.RuleID,
		html.RuleID,
		images.RuleID,
		opengraph.RuleID,
		twitter.RuleID,
	})
}
