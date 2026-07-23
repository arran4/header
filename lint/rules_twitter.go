package lint

import (
	"github.com/arran4/header"
)

type twitterRule struct{}

func (r twitterRule) Name() string { return "Twitter Card Checks" }

func (r twitterRule) Tags() []string { return []string{"", "default", "twitter"} }

func (r twitterRule) Check(info *header.PageInfo) []Result {
	var results []Result

	if info.TwitterCard != "" {
		validCards := map[string]bool{
			"summary":             true,
			"summary_large_image": true,
			"app":                 true,
			"player":              true,
		}

		if !validCards[info.TwitterCard] {
			results = append(results, Result{SeverityError, "Invalid twitter:card value: " + info.TwitterCard})
		}

		if info.TwitterCard == "summary_large_image" {
			if info.TwitterImage == "" {
				results = append(results, Result{SeverityError, "twitter:card is summary_large_image but twitter:image is missing"})
			}
		}

		if info.TwitterCard == "player" {
			if info.TwitterPlayer == "" {
				results = append(results, Result{SeverityError, "twitter:card is player but twitter:player is missing"})
			}
		}
	} else {
		// If they have twitter fields but no card, it's usually bad
		if info.TwitterTitle != "" || info.TwitterDescription != "" || info.TwitterImage != "" {
			results = append(results, Result{SeverityWarning, "Twitter metadata found but twitter:card is missing (defaults to summary, but should be explicit)"})
		}
	}

	return results
}

func init() {
	Register(twitterRule{})
}
