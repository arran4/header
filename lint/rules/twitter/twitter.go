package twitter

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "twitter"

type twitterRule struct{}

func (r twitterRule) ID() string   { return RuleID }
func (r twitterRule) Name() string { return "Twitter Card Checks" }

func (r twitterRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	if info.TwitterCard != "" {
		validCards := map[string]bool{
			"summary":             true,
			"summary_large_image": true,
			"app":                 true,
			"player":              true,
		}

		if !validCards[info.TwitterCard] {
			results = append(results, lint.Result{Severity: lint.SeverityError, Message: "Invalid twitter:card value: " + info.TwitterCard})
		}

		if info.TwitterCard == "summary_large_image" {
			if info.TwitterImage == "" {
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: "twitter:card is summary_large_image but twitter:image is missing"})
			}
		}

		if info.TwitterCard == "player" {
			if info.TwitterPlayer == "" {
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: "twitter:card is player but twitter:player is missing"})
			}
		}
	} else {
		// If they have twitter fields but no card, it's usually bad
		if info.TwitterTitle != "" || info.TwitterDescription != "" || info.TwitterImage != "" {
			results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Twitter metadata found but twitter:card is missing (defaults to summary, but should be explicit)"})
		}
	}

	return results
}

func init() {
	lint.RegisterRule(twitterRule{})
}
