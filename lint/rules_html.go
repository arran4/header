package lint

import (
	"strings"

	"github.com/arran4/header"
)

type htmlRule struct{}

func (r htmlRule) Name() string { return "HTML Good Practices" }

func (r htmlRule) Tags() []string { return []string{"", "default", "html"} }

func (r htmlRule) Check(info *header.PageInfo) []Result {
	var results []Result

	if info.Title == "" {
		results = append(results, Result{SeverityWarning, "Missing HTML <title>"})
	}

	// Icon checks
	hasIconSizes := false
	hasStandardPWASizes := false // 192x192 or 512x512

	for _, icon := range info.Icons {
		if icon.Sizes != "" {
			hasIconSizes = true
			if strings.Contains(icon.Sizes, "192x192") || strings.Contains(icon.Sizes, "512x512") {
				hasStandardPWASizes = true
			}
		}
	}

	if len(info.Icons) > 0 && !hasIconSizes {
		results = append(results, Result{SeverityInfo, "Icons present but none specify 'sizes' attribute"})
	}

	if len(info.Icons) > 0 && !hasStandardPWASizes {
		results = append(results, Result{SeverityInfo, "Standard PWA icon sizes (192x192, 512x512) are missing"})
	}

	// Basic mime type format checks (mostly on alternate links for RSS/Atom/etc)
	for _, altLink := range info.AlternateLinks {
		if altLink.Media != "" {
			// Actually 'media' attribute in <link rel="alternate"> is for media queries.
			// Wait, the struct maps type attribute to Media? Let's assume so based on context or maybe it's just 'type'
			if !strings.Contains(altLink.Media, "/") && !strings.Contains(altLink.Media, "(") {
				// If it's a type like "application/rss+xml" it has a slash. If it's media query like "(max-width: 600px)" it has parens.
				// This is a rough check.
				results = append(results, Result{SeverityWarning, "Alternate link has an unusual media/type attribute format: " + altLink.Media})
			}
		}
	}

	return results
}

func init() {
	Register(htmlRule{})
}
