package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/header"
)

func HttprettyCmd(args []string) int {
	fs := flag.NewFlagSet("header httpretty", flag.ContinueOnError)
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")
	followMetaRefresh := fs.Bool("follow-meta-redirects", false, "Follow meta refresh redirects")
	updateReferer := fs.Bool("update-referer", true, "Update Referer header on redirect")
	storeCookies := fs.Bool("store-cookies", true, "Store cookies across redirects")
	redirectLimit := fs.Int("redirect-limit", 20, "Maximum number of redirects to follow")
	redirectDelay := fs.Duration("redirect-delay", 0, "Delay between redirects (e.g. 1s, 500ms)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: header httpretty [flags] <url>\n")
		fs.PrintDefaults()
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent:         *userAgent,
		TraceRedirects:    false,
		FollowMetaRefresh: *followMetaRefresh,
		UpdateReferer:     *updateReferer,
		StoreCookies:      *storeCookies,
		RedirectLimit:     *redirectLimit,
		RedirectDelay:     *redirectDelay,
		Httpretty:         true,
	}

	_, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	return 0
}
