package diagnostic

import (
	"github.com/arran4/header"
)

type missingFieldRule struct{}

func (r missingFieldRule) Name() string { return "MissingFieldCheck" }

func (r missingFieldRule) Check(info *header.PageInfo) []Result {
	var results []Result

	if info.Title == "" {
		results = append(results, Result{SeverityError, "Missing Title"})
	}
	if info.OGType == "" {
		results = append(results, Result{SeverityWarning, "Missing OpenGraph Data"})
	}
	if info.TwitterCard == "" {
		results = append(results, Result{SeverityWarning, "Missing Twitter Card Data"})
	}
	if info.JSONLDType == "" {
		results = append(results, Result{SeverityWarning, "Missing JSON-LD Metadata"})
	}
	if len(info.CoverImages()) == 0 {
		results = append(results, Result{SeverityWarning, "No Cover Images found"})
	}

	return results
}

func init() {
	Register(missingFieldRule{})
}
