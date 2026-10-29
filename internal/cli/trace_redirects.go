package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/arran4/header"
)

// TraceRedirects is a subcommand `header trace-redirects` -- Trace URL redirects.
//
// Flags:
//
//	followMetaRefresh: --follow-meta-redirects (default: true)  Follow meta refresh redirects
//	updateReferer:     --update-referer (default: true)         Update Referer header on redirect
//	storeCookies:      --store-cookies (default: true)          Store cookies across redirects
//	redirectLimit:     --redirect-limit (default: 20)           Maximum number of redirects to follow
//	redirectDelay:     --redirect-delay (default: 0)            Delay between redirects (e.g. 1s, 500ms)
//	userAgent:         --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//	httprettyFlag:     --httpretty                              Print HTTP requests and responses
//	urlStr:            @1 (required)                            URL to trace
func TraceRedirects(
	followMetaRefresh bool,
	updateReferer bool,
	storeCookies bool,
	redirectLimit int,
	redirectDelay time.Duration,
	userAgent string,
	httprettyFlag bool,
	urlStr string,
) error {
	opts := header.FetchOptions{
		UserAgent:         userAgent,
		TraceRedirects:    true,
		FollowMetaRefresh: followMetaRefresh,
		UpdateReferer:     updateReferer,
		StoreCookies:      storeCookies,
		RedirectLimit:     redirectLimit,
		RedirectDelay:     redirectDelay,
		Httpretty:         httprettyFlag,
	}

	fmt.Printf("Tracing %s\n", urlStr)

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
	}

	if info != nil {
		for i, url := range info.Trace {
			fmt.Printf("%d: %s\n", i+1, url)
		}
		if len(info.Trace) > 0 {
			fmt.Println("Done")
		}
	}

	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}
	return nil
}
