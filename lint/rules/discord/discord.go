package discord

import (
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "discord"

// discordRule is a lint rule that checks for the presence of metadata required for Discord inline video playback.
// What it checks:
// - Verifies that if the page seems to be a video page (like an Instagram reel or short), or already has video intent like og:video,
//   it also provides `twitter:player` since Discord relies heavily on `twitter:player` (or specific oEmbed/OG tags) for inline playback.
//
// Why it checks this:
// - Without `twitter:player`, Discord will not embed the video inline, and instead only show an image preview.
//
// References:
// - Discord Developer Documentation on Embeds.
type discordRule struct{}

func (r discordRule) ID() string   { return RuleID }
func (r discordRule) Name() string { return "Discord Inline Video Checks" }

func (r discordRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	isVideoIntent := false

	if info.OGVideo != "" {
		isVideoIntent = true
	}

	// Instagram Reels/Shorts are videos
	if strings.Contains(info.URL, "/reels/") || strings.Contains(info.URL, "/reel/") || strings.Contains(info.URL, "/shorts/") {
		isVideoIntent = true
	}

	for _, al := range info.AlternateLinks {
		if strings.Contains(al.URL, "/reels/") || strings.Contains(al.URL, "/reel/") || strings.Contains(al.URL, "/shorts/") {
			isVideoIntent = true
		}
	}

	if isVideoIntent {
		if info.TwitterPlayer == "" {
			results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Missing twitter:player which is typically required for Discord inline video playback"})
		}
	}

	return results
}

func init() {
	lint.RegisterRule(discordRule{})
}
