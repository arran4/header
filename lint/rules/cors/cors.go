package cors

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

// RuleID is the unique identifier for the cors lint rule.
const RuleID = "cors"

// corsRule is a lint rule that checks if referenced resources support Cross-Origin Resource Sharing (CORS).
// What it checks:
// - Extracts URLs for images (like `og:image`, `twitter:image`) and icons.
// - Sends an HTTP HEAD request to these URLs, supplying an `Origin` header that mimics the current page's origin.
// - Inspects the `Access-Control-Allow-Origin` response header.
// - Warns if the header is missing or does not allow the current page's origin when the resource is hosted on a different origin.
//
// Why it checks this:
// - Resources served without proper CORS headers cannot be fetched via JavaScript (e.g., using `fetch()`) from different origins.
// - This is particularly problematic for web applications that need to draw cross-origin images onto a `<canvas>` element (which taints the canvas and prevents `toDataURL` export if CORS is missing).
//
// References:
// - MDN Cross-Origin Resource Sharing (CORS): https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
type corsRule struct{}

func (r corsRule) ID() string   { return RuleID }
func (r corsRule) Name() string { return "CORS Checks" }

// HTTPClient represents the interface we need from the HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r corsRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
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

	targets := make(map[string]string) // url -> url

	add := func(u string) {
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
		targets[u] = u
	}

	// Images
	add(info.OGImage)
	add(info.OGImageSecureURL)
	add(info.TwitterImage)
	for _, icon := range info.Icons {
		add(icon.URL)
	}

	if len(targets) == 0 {
		return results
	}

	origin := "https://example.com"
	if info.URL != "" {
		base, err := url.Parse(info.URL)
		if err == nil {
			origin = fmt.Sprintf("%s://%s", base.Scheme, base.Host)
		}
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, t := range targets {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()

			req, err := http.NewRequest("HEAD", targetUrl, nil)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Invalid URL %s: %v", targetUrl, err)})
				mu.Unlock()
				return
			}
			req.Header.Set("User-Agent", "Header-Lint-Bot/1.0")
			req.Header.Set("Origin", origin)

			resp, err := client.Do(req)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Unreachable URL %s: %v", targetUrl, err)})
				mu.Unlock()
				return
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode >= 400 {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("URL %s returned status %d", targetUrl, resp.StatusCode)})
				mu.Unlock()
				return
			}

			// Some servers might not return CORS headers for HEAD requests, but we'll check it anyway
			acao := resp.Header.Get("Access-Control-Allow-Origin")
			if acao == "" {
				// We can't strictly say it's an error because the image might be served from the same origin,
				// or the server doesn't need to serve CORS if it's not requested with CORS.
				// Let's only warn if it's on a different origin.
				targetParsed, err := url.Parse(targetUrl)
				if err == nil {
					targetOrigin := fmt.Sprintf("%s://%s", targetParsed.Scheme, targetParsed.Host)
					if targetOrigin != origin {
						mu.Lock()
						results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("URL %s on different origin might missing CORS header Access-Control-Allow-Origin", targetUrl)})
						mu.Unlock()
					}
				}
			} else if acao != "*" && !strings.Contains(acao, origin) {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("URL %s has Access-Control-Allow-Origin %s which does not allow %s", targetUrl, acao, origin)})
				mu.Unlock()
			}
		}(t)
	}

	wg.Wait()

	return results
}

func init() {
	lint.RegisterRule(corsRule{})
}
