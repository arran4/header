package consistency

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "consistency"

type consistencyRule struct{}

func (r consistencyRule) ID() string   { return RuleID }
func (r consistencyRule) Name() string { return "Metadata Consistency Checks" }

func (r consistencyRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	if info.OGTitle != "" && info.TwitterTitle != "" && info.OGTitle != info.TwitterTitle {
		results = append(results, lint.Result{Severity: lint.SeverityInfo, Message: "og:title and twitter:title differ"})
	}

	if info.OGDescription != "" && info.TwitterDescription != "" && info.OGDescription != info.TwitterDescription {
		results = append(results, lint.Result{Severity: lint.SeverityInfo, Message: "og:description and twitter:description differ"})
	}

	return results
}

func init() {
	lint.RegisterRule(consistencyRule{})
}
