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
	defer resp.Body.Close()

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
