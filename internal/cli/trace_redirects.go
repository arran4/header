package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/header"
)

func TraceRedirectsCmd(args []string) int {
	fs := flag.NewFlagSet("header trace-redirects", flag.ContinueOnError)
	followMetaRefresh := fs.Bool("follow-meta-redirects", true, "Follow meta refresh redirects")
	updateReferer := fs.Bool("update-referer", true, "Update Referer header on redirect")
	storeCookies := fs.Bool("store-cookies", true, "Store cookies across redirects")
	redirectLimit := fs.Int("redirect-limit", 20, "Maximum number of redirects to follow")
	redirectDelay := fs.Duration("redirect-delay", 0, "Delay between redirects (e.g. 1s, 500ms)")
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")
	httprettyFlag := fs.Bool("httpretty", false, "Print HTTP requests and responses")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: header trace-redirects [flags] <url>\n")
		fs.PrintDefaults()
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent:         *userAgent,
		TraceRedirects:    true,
		FollowMetaRefresh: *followMetaRefresh,
		UpdateReferer:     *updateReferer,
		StoreCookies:      *storeCookies,
		RedirectLimit:     *redirectLimit,
		RedirectDelay:     *redirectDelay,
		Httpretty:         *httprettyFlag,
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
		return 1
	}
	return 0
}
