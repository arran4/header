package twitter

import (
	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "twitter"

// twitterRule is a lint rule that validates the structure and requirements of Twitter Card metadata.
// What it checks:
// - Validates the `twitter:card` attribute against the allowed values (`summary`, `summary_large_image`, `app`, `player`).
// - If `twitter:card` is set to `summary_large_image`, it ensures the `twitter:image` attribute is also provided.
// - If `twitter:card` is set to `player`, it ensures the `twitter:player` attribute is provided.
// - Warns if Twitter-specific metadata (like title, description, or image) is present, but the explicit `twitter:card` type declaration is missing.
//
// Why it checks this:
// - Twitter requires specific field combinations for its various card types to render correctly in the timeline.
// - Providing an invalid card type or missing required dependent fields will result in Twitter falling back to a default link layout or failing to show a rich preview.
//
// References:
// - Twitter Cards Getting Started Guide: https://developer.x.com/en/docs/twitter-for-websites/cards/guides/getting-started
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
