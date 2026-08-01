package urls

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

// RuleID is the unique identifier for the URL validation lint rule.
const RuleID = "urls"

// urlCheckRule is a lint rule that checks the reachability and content types of referenced URLs.
// What it checks:
// - Extracts URLs referenced in the page metadata (e.g., images, canonical links).
// - Resolves relative URLs to absolute URLs using the page's base URL.
// - Sends an HTTP HEAD request to each extracted URL.
// - Verifies the URL is reachable (status code < 400).
// - Validates that the returned `Content-Type` matches the expected type (e.g., verifying image URLs actually return an `image/*` content type).
//
// Why it checks this:
// - Broken links (404s, 500s) degrade the user experience and can negatively impact SEO and link previews.
// - Returning an incorrect `Content-Type` (like serving an HTML error page instead of an image) can cause social platforms to silently fail to display the media.
//
// References:
// - HTTP HEAD Method: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD
// - HTTP Content-Type Header: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
type urlCheckRule struct{}

func (r urlCheckRule) ID() string   { return RuleID }
func (r urlCheckRule) Name() string { return "URL Reachability & Content-Type Checks" }

// HTTPClient represents the interface we need from the HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r urlCheckRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	var client HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, opt := range ops {
		switch o := opt.(type) {
		case HTTPClient:
			client = o
		}
	}

	type checkTarget struct {
		url          string
		expectedType string // prefix like "image/"
	}

	targets := make(map[string]checkTarget)

	add := func(u, expectedType string) {
		if u == "" {
			return
		}
		// Resolve relative URLs if info.URL is available
		if info.URL != "" {
			base, err := url.Parse(info.URL)
			if err == nil {
				parsedURL, err := url.Parse(u)
				if err == nil {
					u = base.ResolveReference(parsedURL).String()
				}
			}
		}
		targets[u] = checkTarget{url: u, expectedType: expectedType}
	}

	// Images
	add(info.OGImage, "image/")
	add(info.OGImageSecureURL, "image/")
	add(info.TwitterImage, "image/")
	for _, icon := range info.Icons {
		add(icon.URL, "image/")
	}

	// Links
	add(info.CanonicalLink, "") // Canonical usually HTML, but we don't strictly require a specific content type
	add(info.ShortLink, "")

	if len(targets) == 0 {
		return results
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, t := range targets {
		wg.Add(1)
		go func(t checkTarget) {
			defer wg.Done()

			req, err := http.NewRequest("HEAD", t.url, nil)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Invalid URL %s: %v", t.url, err)})
				mu.Unlock()
				return
			}
			req.Header.Set("User-Agent", "Header-Lint-Bot/1.0")

			resp, err := client.Do(req)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Unreachable URL %s: %v", t.url, err)})
				mu.Unlock()
				return
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode >= 400 {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("URL %s returned status %d", t.url, resp.StatusCode)})
				mu.Unlock()
				return
			}

			if t.expectedType != "" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), t.expectedType) {
					mu.Lock()
					results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("URL %s expected %s but got Content-Type %s", t.url, t.expectedType, contentType)})
					mu.Unlock()
				}
			}
		}(t)
	}

	wg.Wait()

	return results
}

func init() {
	lint.RegisterRule(urlCheckRule{})
}
