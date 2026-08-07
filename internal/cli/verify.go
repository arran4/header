package cli

import (
	"fmt"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
	_ "github.com/arran4/header/lint/lists/verify" // Load verify list
)

// Verify is a subcommand `header verify` -- Checks all the urls, and other details, it's a subset of the lint rules that is mostly concerned with valid info.
//
// Flags:
//   userAgent:     --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//   httprettyFlag: --httpretty                              Print HTTP requests and responses
//   urlStr:        @1 (required)                            URL to verify
func Verify(userAgent string, httprettyFlag bool, urlStr string) error {
	opts := header.FetchOptions{
		UserAgent:         userAgent,
		TraceRedirects:    false,
		FollowMetaRefresh: false,
		UpdateReferer:     true,
		StoreCookies:      true,
		RedirectLimit:     20,
		RedirectDelay:     0,
		Httpretty:         httprettyFlag,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("Error fetching URL: %w", err)
	}

	lint.Run(info, []string{"verify"}, nil)

	return nil
}
