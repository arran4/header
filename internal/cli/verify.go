package cli

import (
	"fmt"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
	_ "github.com/arran4/header/lint/lists/verify" // Load verify list
)

// Verify is a subcommand `header verify` -- Checks all the urls, and other details, it's a subset of the lint rules that is mostly concerned with valid info.
//
// Flags:
//
//	userAgent:     --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//	httprettyFlag: --httpretty                              Print HTTP requests and responses
//	retryAttempts: --retry-attempts (default: 3)            Number of attempts for idempotent HTTP requests (1 means no retries)
//	retryDelay:    --retry-delay (default: 250ms)           Initial delay between retries
//	retryMaxDelay: --retry-max-delay (default: 5s)          Maximum delay between retries
//	timeout:       --timeout (default: 20s)                 Maximum total request timeout
//	urlStr:        @1 (required)                            URL to verify
func Verify(
	userAgent string,
	httprettyFlag bool,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	urlStr string,
) error {
	opts := header.FetchOptions{
		UserAgent:         userAgent,
		TraceRedirects:    false,
		FollowMetaRefresh: false,
		UpdateReferer:     true,
		StoreCookies:      true,
		RedirectLimit:     20,
		RedirectDelay:     0,
		Httpretty:         httprettyFlag,
		RetryAttempts:     retryAttempts,
		RetryDelay:        retryDelay,
		RetryMaxDelay:     retryMaxDelay,
		Timeout:           timeout,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	lint.Run(info, []string{"verify"}, nil)

	return nil
}
