package lint

import (
	"github.com/arran4/header"
)

type opengraphRule struct{}

func (r opengraphRule) Name() string { return "OpenGraph Basic Checks" }

func (r opengraphRule) Tags() []string { return []string{"", "default", "opengraph"} }

func (r opengraphRule) Check(info *header.PageInfo) []Result {
	var results []Result

	// These are typically required OpenGraph fields
	if info.OGTitle == "" {
		results = append(results, Result{SeverityWarning, "Missing og:title"})
	}
	if info.OGType == "" {
		results = append(results, Result{SeverityWarning, "Missing og:type"})
	}
	if info.OGImage == "" && info.OGImageSecureURL == "" {
		results = append(results, Result{SeverityWarning, "Missing og:image (and og:image:secure_url)"})
	}

	// URL is technically part of the basic requirements for opengraph but we might just use the canonical
	if info.URL == "" {
		// Not strictly an OG property in our struct maybe, but usually og:url is checked.
		// Wait, our struct doesn't have an explicitly mapped OGUrl, but CanonicalLink is usually it.
		// We'll skip og:url specifically here since we only have CanonicalLink or just URL.
	}

	return results
}

func init() {
	Register(opengraphRule{})
}
