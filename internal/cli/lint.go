package cli

import (
	"fmt"
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

// Lint is a subcommand `header lint` -- Lint the metadata of a URL.
//
// Flags:
//   listFlag:      --list                                   Comma separated list of lint lists to run (e.g. 'all', 'recommended'). Default is 'recommended'.
//   ruleFlag:      --rule                                   Comma separated list of specific lint rules to run (e.g. 'twitter,opengraph').
//   listRulesFlag: --list-rules                             List available lint rules and lists, then exit.
//   userAgent:     --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//   httprettyFlag: --httpretty                              Print HTTP requests and responses
//   urlStr:        @1 (required)                            URL to lint
func Lint(listFlag string, ruleFlag string, listRulesFlag bool, userAgent string, httprettyFlag bool, urlStr string) error {
	if listRulesFlag {
		printAvailableLintRules()
		return nil
	}

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

	listsToRun := []string{}
	if listFlag != "" {
		listsToRun = strings.Split(listFlag, ",")
		for i, t := range listsToRun {
			listsToRun[i] = strings.TrimSpace(t)
		}
	}

	rulesToRun := []string{}
	if ruleFlag != "" {
		rulesToRun = strings.Split(ruleFlag, ",")
		for i, t := range rulesToRun {
			rulesToRun[i] = strings.TrimSpace(t)
		}
	}

	lint.Run(info, listsToRun, rulesToRun)

	return nil
}

// LintList is a subcommand `header lint list` -- List available lint lists.
func LintList() error {
	printAvailableLintRules()
	return nil
}

// LintListRules is a subcommand `header lint list-rules` -- List available lint rules.
func LintListRules() error {
	printAvailableLintRules()
	return nil
}
