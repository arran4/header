package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/header"

	"github.com/arran4/header/diagnostic"
	"github.com/arran4/header/internal/ui"
	"github.com/arran4/header/lint"
	_ "github.com/arran4/header/lint/lists/all" // Ensure rules are loaded
	_ "github.com/arran4/header/lint/lists/recommended"
	"github.com/arran4/header/templates"
	"github.com/fatih/color"
)

// To avoid unused import issues if lint isn't used directly here, we wrap it
func importLint(info *header.PageInfo, lists, rules []string) {
	lint.Run(info, lists, rules)
}

// Extract is a subcommand `header extract` -- Extract metadata from URL.
//
// Flags:
//   testMode:          --test                                   Run in test mode to show diagnostics/missing data
//   lintListFlags:     --lint-list                              Run linting with specific lists (comma separated). Default is 'recommended' if flag is used
//   lintRuleFlags:     --lint-rule                              Run linting with specific rules (comma separated).
//   lintRun:           --lint-run                               Run default linting
//   showImage:         --show-image                             Display the cover image (if available)
//   style:             --style (default: "default")             The output style template to use (e.g. default, plain)
//   userAgent:         --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//   showUnknown:       --show-unknown                           Show unknown metadata tags
//   excludeUnknown:    --exclude-unknown                        Comma separated list of unknown tags to exclude
//   traceRedirects:    --trace-redirects                        Trace redirects like trace-redirects tool
//   followMetaRefresh: --follow-meta-redirects                  Follow meta refresh redirects
//   updateReferer:     --update-referer (default: true)         Update Referer header on redirect
//   storeCookies:      --store-cookies (default: true)          Store cookies across redirects
//   redirectLimit:     --redirect-limit (default: 20)           Maximum number of redirects to follow
//   redirectDelay:     --redirect-delay (default: 0)            Delay between redirects (e.g. 1s, 500ms)
//   httprettyFlag:     --httpretty                              Print HTTP requests and responses
//   urlStr:            @1 (required)                            URL to extract
//
// Aliases: show
func Extract(
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
	opts := header.FetchOptions{
		UserAgent:         userAgent,
		TraceRedirects:    traceRedirects,
		FollowMetaRefresh: followMetaRefresh,
		UpdateReferer:     updateReferer,
		StoreCookies:      storeCookies,
		RedirectLimit:     redirectLimit,
		RedirectDelay:     redirectDelay,
		Httpretty:         httprettyFlag,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("Error fetching URL: %w", err)
	}

	if traceRedirects && len(info.Trace) > 0 {
		fmt.Printf("Tracing %s\n", urlStr)
		for i, url := range info.Trace {
			fmt.Printf("%d: %s\n", i+1, url)
		}
		fmt.Println("Done")
	}

	if showUnknown {
		var excluded []string
		if excludeUnknown != "" {
			parts := strings.Split(excludeUnknown, ",")
			for _, p := range parts {
				excluded = append(excluded, strings.TrimSpace(p))
			}
		}
		info.Mutate(func(i *header.PageInfo) {
			i.UnknownMeta = i.KnownMetaTracker.ProcessUnknown(excluded)
		})
	}

	tmplContent, err := templates.FS.ReadFile(style + ".tmpl")
	if err != nil {
		return fmt.Errorf("Error loading template '%s': %w", style, err)
	}

	funcMap := template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
		"json": func(v interface{}) (string, error) {
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil { return "", fmt.Errorf("failed to marshal json: %w", err) }
			return string(b), nil
		},
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("Error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		return fmt.Errorf("Error executing template: %w", err)
	}
	fmt.Print(buf.String())
	fmt.Println()

	if testMode {
		diagnostic.Run(info)
	}

	isLintRun := lintRun || lintListFlags != "" || lintRuleFlags != ""
	if isLintRun {
		lists := []string{}
		if lintListFlags != "" {
			lists = strings.Split(lintListFlags, ",")
			for i, t := range lists {
				lists[i] = strings.TrimSpace(t)
			}
		}

		rules := []string{}
		if lintRuleFlags != "" {
			rules = strings.Split(lintRuleFlags, ",")
			for i, t := range rules {
				rules[i] = strings.TrimSpace(t)
			}
		}

		importLint(info, lists, rules)
	}

	if showImage {
		images := info.CoverImages()
		if len(images) > 0 {
			ui.DisplayImage(images[0], "default")
		} else {
			fmt.Println("No cover image found to display.")
		}
	}

	return nil
}
