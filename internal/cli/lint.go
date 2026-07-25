package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
	_ "github.com/arran4/header/lint/rules/all" // Load all rules by default
)

func LintCmd(args []string) int {
	fs := flag.NewFlagSet("header lint", flag.ContinueOnError)
	listFlag := fs.String("list", "", "Comma-separated list of lint lists to run (e.g. 'all', 'recommended'). Default is 'default'.")
	ruleFlag := fs.String("rule", "", "Comma-separated list of specific lint rules to run (e.g. 'twitter,opengraph').")
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: header lint [flags] <url>\n")
		fs.PrintDefaults()
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent:         *userAgent,
		TraceRedirects:    false,
		FollowMetaRefresh: false,
		UpdateReferer:     true,
		StoreCookies:      true,
		RedirectLimit:     20,
		RedirectDelay:     0,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	lists := []string{}
	if *listFlag != "" {
		lists = strings.Split(*listFlag, ",")
		for i, t := range lists {
			lists[i] = strings.TrimSpace(t)
		}
	}

	rules := []string{}
	if *ruleFlag != "" {
		rules = strings.Split(*ruleFlag, ",")
		for i, t := range rules {
			rules[i] = strings.TrimSpace(t)
		}
	}

	lint.Run(info, lists, rules)

	return 0
}
