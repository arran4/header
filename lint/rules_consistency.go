package lint

import (
	"github.com/arran4/header"
)

type consistencyRule struct{}

func (r consistencyRule) Name() string { return "Metadata Consistency Checks" }

func (r consistencyRule) Tags() []string { return []string{"", "default", "consistency"} }

func (r consistencyRule) Check(info *header.PageInfo) []Result {
	var results []Result

	if info.OGTitle != "" && info.TwitterTitle != "" && info.OGTitle != info.TwitterTitle {
		results = append(results, Result{SeverityInfo, "og:title and twitter:title differ"})
	}

	if info.OGDescription != "" && info.TwitterDescription != "" && info.OGDescription != info.TwitterDescription {
		results = append(results, Result{SeverityInfo, "og:description and twitter:description differ"})
	}

	return results
}

func init() {
	Register(consistencyRule{})
}
