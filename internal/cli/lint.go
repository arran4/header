package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
	_ "github.com/arran4/header/lint/lists/all" // Load all lists and rules by default
	_ "github.com/arran4/header/lint/lists/recommended"
)

func printAvailableLintRules() {
	PrintLists()
	fmt.Println()
	PrintRules()
}

func LintCmd(args []string) int {
	// First check if the user is calling `header lint list` or `header lint list-rules` as a pseudo-subcommand
	if len(args) > 0 && (args[0] == "list" || args[0] == "list-rules") {
		printAvailableLintRules()
		return 0
	}

	fs := flag.NewFlagSet("header lint", flag.ContinueOnError)
	listFlag := fs.String("list", "", "Comma-separated list of lint lists to run (e.g. 'all', 'recommended'). Default is 'recommended'.")
	ruleFlag := fs.String("rule", "", "Comma-separated list of specific lint rules to run (e.g. 'twitter,opengraph').")
	listRulesFlag := fs.Bool("list-rules", false, "List available lint rules and lists, then exit.")
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")
	httprettyFlag := fs.Bool("httpretty", false, "Print HTTP requests and responses")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	if *listRulesFlag {
		printAvailableLintRules()
		return 0
	}

	pargs := fs.Args()
	if len(pargs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: header lint [flags] <url>\n")
		fmt.Fprintf(os.Stderr, "       header lint list\n")
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
		Httpretty:         *httprettyFlag,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	listsToRun := []string{}
	if *listFlag != "" {
		listsToRun = strings.Split(*listFlag, ",")
		for i, t := range listsToRun {
			listsToRun[i] = strings.TrimSpace(t)
		}
	}

	rulesToRun := []string{}
	if *ruleFlag != "" {
		rulesToRun = strings.Split(*ruleFlag, ",")
		for i, t := range rulesToRun {
			rulesToRun[i] = strings.TrimSpace(t)
		}
	}

	lint.Run(info, listsToRun, rulesToRun)

	return 0
}
