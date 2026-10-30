package cards

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/arran4/header"
)

// CardContent is presentation text selected from raw page metadata.
// Source fields record the provenance of each selection. This does not
// change PageInfo or imply that the card is approved for publication.
type CardContent struct {
	Title             string
	Description       string
	TitleSource       string
	DescriptionSource string
	Quality           string
}

const (
	// CardQualityCandidate means no known metadata-quality issue was detected.
	// It does not mean that a human approved the card.
	CardQualityCandidate = "candidate"
	// CardQualityNeedsOwnerDescription means no meaningful description exists.
	CardQualityNeedsOwnerDescription = "needs-owner-description"
	// CardQualityNeedsReview means the source has suspicious or missing text.
	CardQualityNeedsReview = "needs-review"
)

var githubRepositoryPart = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// SelectCardContent chooses presentation text without overwriting the raw
// Open Graph, Twitter or HTML values. For github.com/owner/repo URLs, GitHub's
// HTML title and known signup suffix are not useful on repository cards.
// Do not apply those domain-specific rules to GitHub subpages or other hosts.
func SelectCardContent(info *header.PageInfo) CardContent {
	result := CardContent{
		TitleSource:       "none",
		DescriptionSource: "none",
		Quality:           CardQualityCandidate,
	}
	if info == nil {
		result.Title = "No Title"
		result.Quality = CardQualityNeedsOwnerDescription
		return result
	}

	repo := githubRepositoryName(info.URL)
	if repo != "" {
		// This identity comes directly from the requested URL and cannot be
		// confused with the GitHub HTML shell's branding text.
		result.Title = repo
		result.TitleSource = "github-repository-url"
	} else {
		switch {
		case strings.TrimSpace(info.Title) != "":
			result.Title = info.Title
			result.TitleSource = "page-html-title"
		case strings.TrimSpace(info.OGTitle) != "":
			result.Title = info.OGTitle
			result.TitleSource = "page-og-title"
		case strings.TrimSpace(info.TwitterTitle) != "":
			result.Title = info.TwitterTitle
			result.TitleSource = "page-twitter-title"
		default:
			result.Title = "No Title"
			result.Quality = CardQualityNeedsReview
		}
	}

	sources := []struct {
		text   string
		source string
	}{
		{info.OGDescription, "page-og-description"},
		{info.TwitterDescription, "page-twitter-description"},
	}
	sawGithubBoilerplate := false
	for _, source := range sources {
		desc := strings.TrimSpace(source.text)
		if desc == "" {
			continue
		}
		if repo != "" {
			clean, trimmed := stripGitHubDescriptionBoilerplate(desc, repo)
			if trimmed {
				sawGithubBoilerplate = true
				desc = clean
				if desc != "" {
					result.DescriptionSource = source.source + "+github-boilerplate-removed"
				}
			} else if containsGitHubSignupBoilerplate(desc) {
				// Unknown/mismatched GitHub boilerplate is not trustworthy
				// repository description text; require human review.
				sawGithubBoilerplate = true
				continue
			}
		}
		if desc == "" {
			continue
		}
		result.Description = desc
		if result.DescriptionSource == "none" {
			result.DescriptionSource = source.source
		}
		break
	}

	if result.Description == "" {
		result.Quality = CardQualityNeedsOwnerDescription
	} else if result.Quality == CardQualityCandidate && sawGithubBoilerplate && containsGitHubSignupBoilerplate(result.Description) {
		result.Quality = CardQualityNeedsReview
	}
	return result
}

// githubRepositoryName only accepts a plain GitHub repository page URL. This
// intentionally excludes GitHub user pages, issues, pulls and file subpaths.
func githubRepositoryName(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") ||
		!strings.EqualFold(u.Hostname(), "github.com") || u.Port() != "" || u.User != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return ""
	}
	for _, part := range parts {
		if part == "" || !githubRepositoryPart.MatchString(part) || part == "." || part == ".." {
			return ""
		}
	}
	return parts[0] + "/" + parts[1]
}

// stripGitHubDescriptionBoilerplate removes ONLY GitHub's known suffix for
// the exact repository named in the URL. The leading text is GitHub's own
// page metadata; no description is inferred from the README or made up.
func stripGitHubDescriptionBoilerplate(text, repo string) (string, bool) {
	text = strings.TrimSpace(text)
	suffix := "Contribute to " + repo + " development by creating an account on GitHub."
	if len(text) < len(suffix) || !strings.EqualFold(text[len(text)-len(suffix):], suffix) {
		return text, false
	}
	prefix := strings.TrimSpace(text[:len(text)-len(suffix)])
	prefix = strings.TrimSpace(strings.TrimSuffix(prefix, "."))
	return prefix, true
}

func containsGitHubSignupBoilerplate(text string) bool {
	return strings.Contains(strings.ToLower(text), "development by creating an account on github")
}
