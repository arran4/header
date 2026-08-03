package consistency

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

// RuleID is the unique identifier for the consistency lint rule.
const RuleID = "consistency"

// consistencyRule is a lint rule that checks for parity between different metadata ecosystems.
// What it checks:
// - Cross-references the OpenGraph title (`og:title`) against the Twitter title (`twitter:title`).
// - Cross-references the OpenGraph description (`og:description`) against the Twitter description (`twitter:description`).
// - Emits an info-level notice if they are both populated but contain different values.
//
// Why it checks this:
// - Maintaining separate, duplicate tags for OpenGraph and Twitter is common, but differences between them are usually unintentional sync issues caused by CMS bugs or human error.
// - It helps ensure consistent messaging across different social sharing platforms.
type consistencyRule struct{}

func (r consistencyRule) ID() string   { return RuleID }
func (r consistencyRule) Name() string { return "Metadata Consistency Checks" }

func (r consistencyRule) Check(info *header.PageInfo, _ ...any) []lint.Result {
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
