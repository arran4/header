package cloudflare

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "cloudflare"

// cloudflareRule is a lint rule that checks if the URL is blocked by Cloudflare's security mechanisms.
// What it checks:
// - Sends a HEAD request to the target URL.
// - Inspects the `Server` header for "cloudflare" and checks the HTTP status code.
// - Detects 403 Forbidden responses, which often indicate a Cloudflare challenge page (e.g., Turnstile or CAPTCHA).
// - Detects 503 Service Unavailable responses, which can indicate a Cloudflare gateway error or block.
//
// Why it checks this:
// - Automated fetchers, social media crawlers, and link preview generators are frequently blocked by Cloudflare's Bot Management or Web Application Firewall (WAF) rules.
// - If the link returns a 403 or 503 from Cloudflare, the rich preview will likely fail to generate on platforms like Discord or Twitter because they cannot fetch the metadata.
//
// References:
// - Cloudflare Bot Management: https://developers.cloudflare.com/bots/
// - Cloudflare HTTP Status Codes: https://developers.cloudflare.com/support/troubleshooting/cloudflare-errors/troubleshooting-cloudflare-5xx-errors/
type cloudflareRule struct{}

func (r cloudflareRule) ID() string   { return RuleID }
func (r cloudflareRule) Name() string { return "Cloudflare Gateway Checks" }

// HTTPClient represents the interface we need from the HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r cloudflareRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	if info.URL == "" {
		return results
	}

	var client HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, opt := range ops {
		switch o := opt.(type) {
		case HTTPClient:
			client = o
		}
	}

	req, err := http.NewRequest("HEAD", info.URL, nil)
	if err != nil {
		results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Invalid URL %s: %v", info.URL, err)})
		return results
	}
	req.Header.Set("User-Agent", "Header-Lint-Bot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Unreachable URL %s: %v", info.URL, err)})
		return results
	}
	defer func() { _ = resp.Body.Close() }()

	server := resp.Header.Get("Server")
	if strings.Contains(strings.ToLower(server), "cloudflare") && resp.StatusCode == http.StatusForbidden {
		results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("URL %s is returning a Cloudflare challenge (403 Forbidden)", info.URL)})
	} else if strings.Contains(strings.ToLower(server), "cloudflare") && resp.StatusCode == http.StatusServiceUnavailable {
		results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("URL %s is returning a Cloudflare gateway error (503 Service Unavailable)", info.URL)})
	}

	return results
}

func init() {
	lint.RegisterRule(cloudflareRule{})
}
