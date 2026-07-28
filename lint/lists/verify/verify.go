package verify

import (
	"github.com/arran4/header/lint/lists"
	"github.com/arran4/header/lint/rules/images"
	"github.com/arran4/header/lint/rules/urls"
)

func init() {
	lists.Register("verify", []string{
		images.RuleID,
		urls.RuleID,
	})
}
