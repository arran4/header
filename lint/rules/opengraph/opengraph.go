package opengraph

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

type opengraphRule struct{}

func (r opengraphRule) Name() string { return "OpenGraph Basic Checks" }

func (r opengraphRule) Tags() []string { return []string{"", "default", "opengraph"} }

func (r opengraphRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
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
	lint.Register(opengraphRule{})
}
