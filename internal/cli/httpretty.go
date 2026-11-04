package cli

import (
	"fmt"
	"time"

	"github.com/arran4/header"
)

// Httpretty is a subcommand `header httpretty` -- Fetch a URL and print HTTP requests and responses using httpretty.
//
// Flags:
//
//	userAgent:         --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//	followMetaRefresh: --follow-meta-redirects                  Follow meta refresh redirects
//	updateReferer:     --update-referer (default: true)         Update Referer header on redirect
//	storeCookies:      --store-cookies (default: true)          Store cookies across redirects
//	redirectLimit:     --redirect-limit (default: 20)           Maximum number of redirects to follow
//	redirectDelay:     --redirect-delay (default: 0)            Delay between redirects (e.g. 1s, 500ms)
//	retryAttempts:     --retry-attempts (default: 3)            Number of attempts for idempotent HTTP requests (1 means no retries)
//	retryDelay:        --retry-delay (default: 250ms)           Initial delay between retries
//	retryMaxDelay:     --retry-max-delay (default: 5s)          Maximum delay between retries
//	timeout:           --timeout (default: 20s)                 Maximum total request timeout
//	urlStr:            @1 (required)                            URL to fetch
func Httpretty(
	userAgent string,
	followMetaRefresh bool,
	updateReferer bool,
	storeCookies bool,
	redirectLimit int,
	redirectDelay time.Duration,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	urlStr string,
) error {
	opts := header.FetchOptions{
		UserAgent:         userAgent,
		TraceRedirects:    false,
		FollowMetaRefresh: followMetaRefresh,
		UpdateReferer:     updateReferer,
		StoreCookies:      storeCookies,
		RedirectLimit:     redirectLimit,
		RedirectDelay:     redirectDelay,
		Httpretty:         true,
		RetryAttempts:     retryAttempts,
		RetryDelay:        retryDelay,
		RetryMaxDelay:     retryMaxDelay,
		Timeout:           timeout,
	}

	_, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	return nil
}
