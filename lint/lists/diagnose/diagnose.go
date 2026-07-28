package diagnose

import (
	"github.com/arran4/header/lint/lists"
	"github.com/arran4/header/lint/rules/cloudflare"
	"github.com/arran4/header/lint/rules/cors"
)

func init() {
	lists.Register("diagnose", []string{
		cors.RuleID,
		cloudflare.RuleID,
	})
}
