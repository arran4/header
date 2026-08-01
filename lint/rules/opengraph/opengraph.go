package opengraph

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "opengraph"

// opengraphRule is a lint rule that checks for the presence of basic OpenGraph metadata.
// What it checks:
// - Verifies the existence of fundamental OpenGraph tags: `og:title`, `og:type`, and `og:image` (or `og:image:secure_url`).
//
// Why it checks this:
// - These properties are the minimum required fields for a web page to be transformed into a rich object in a social graph.
// - Missing these fields can result in poor or broken previews when sharing links on platforms like Facebook, LinkedIn, Discord, and others.
//
// References:
// - The Open Graph protocol documentation: https://ogp.me/
type opengraphRule struct{}

func (r opengraphRule) ID() string   { return RuleID }
func (r opengraphRule) Name() string { return "OpenGraph Basic Checks" }

func (r opengraphRule) Check(info *header.PageInfo, _ ...any) []lint.Result {
	var results []lint.Result

	// These are typically required OpenGraph fields
	if info.OGTitle == "" {
		results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Missing og:title"})
	}
	if info.OGType == "" {
		results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Missing og:type"})
	}
	if info.OGImage == "" && info.OGImageSecureURL == "" {
		results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Missing og:image (and og:image:secure_url)"})
	}

	return results
}

func init() {
	lint.RegisterRule(opengraphRule{})
}
