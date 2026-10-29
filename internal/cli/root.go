package cli

import (
	"fmt"
	"time"
)

// Root is a subcommand `header` -- Extract metadata from URL.
//
// Flags:
//
//	testMode:          --test                                   Run in test mode to show diagnostics/missing data
//	lintListFlags:     --lint-list                              Run linting with specific lists (comma separated). Default is 'recommended' if flag is used
//	lintRuleFlags:     --lint-rule                              Run linting with specific rules (comma separated).
//	lintRun:           --lint-run                               Run default linting
//	showImage:         --show-image                             Display the cover image (if available)
//	style:             --style (default: "default")             The output style template to use (e.g. default, all_fields_plain)
//	userAgent:         --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//	showUnknown:       --show-unknown                           Show unknown metadata tags
//	excludeUnknown:    --exclude-unknown                        Comma separated list of unknown tags to exclude
//	traceRedirects:    --trace-redirects                        Trace redirects like trace-redirects tool
//	followMetaRefresh: --follow-meta-redirects                  Follow meta refresh redirects
//	updateReferer:     --update-referer (default: true)         Update Referer header on redirect
//	storeCookies:      --store-cookies (default: true)          Store cookies across redirects
//	redirectLimit:     --redirect-limit (default: 20)           Maximum number of redirects to follow
//	redirectDelay:     --redirect-delay (default: 0)            Delay between redirects (e.g. 1s, 500ms)
//	httprettyFlag:     --httpretty                              Print HTTP requests and responses
//	urlStr:            @1 (optional)                                       URL to extract
func Root(
	testMode bool,
	lintListFlags string,
	lintRuleFlags string,
	lintRun bool,
	showImage bool,
	style string,
	userAgent string,
	showUnknown bool,
	excludeUnknown string,
	traceRedirects bool,
	followMetaRefresh bool,
	updateReferer bool,
	storeCookies bool,
	redirectLimit int,
	redirectDelay time.Duration,
	httprettyFlag bool,
	urlStr string,
) error {
	if urlStr == "" {
		fmt.Println("Usage: header <command> [args...]")
		fmt.Println("\nCommands:")
		fmt.Println("  show            Extract metadata from URL (default if no matching command is found)")
		fmt.Println("  edit            Edit HTML file head/meta fields")
		fmt.Println("  trace-redirects Trace URL redirects")
		fmt.Println("  verify           Checks all the urls, and other details, it's a subset of the lint rules that is mostly concerned with valid info")
		fmt.Println("  lint             Lint the metadata of a URL")
		fmt.Println("  httpretty        Fetch a URL and print HTTP requests and responses using httpretty")
		fmt.Println("  lint-rule-lists  List available lint rule lists")
		fmt.Println("  lint-rules       List available lint rules")
		fmt.Println("  view             View images or icons from a URL (e.g. view image <url>)")
		fmt.Println("  card             Export a card (as an image) including the image, description and a couple other things")
		return fmt.Errorf("URL is required")
	}

	return Extract(
		testMode,
		lintListFlags,
		lintRuleFlags,
		lintRun,
		showImage,
		style,
		userAgent,
		showUnknown,
		excludeUnknown,
		traceRedirects,
		followMetaRefresh,
		updateReferer,
		storeCookies,
		redirectLimit,
		redirectDelay,
		httprettyFlag,
		urlStr,
	)
}
