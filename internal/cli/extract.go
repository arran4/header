package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"

	"strings"

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

func ExtractCmd(args []string) int {
	fs := flag.NewFlagSet("header extract", flag.ContinueOnError)
	testMode := fs.Bool("test", false, "Run in test mode to show diagnostics/missing data")
	lintListFlags := fs.String("lint-list", "", "Run linting with specific lists (comma-separated). Default is 'recommended' if flag is used")
	lintRuleFlags := fs.String("lint-rule", "", "Run linting with specific rules (comma-separated).")
	lintRun := false
	fs.BoolVar(&lintRun, "lint-run", false, "Run default linting")
	showImage := fs.Bool("show-image", false, "Display the cover image (if available)")
	style := fs.String("style", "default", "The output style template to use (e.g. default, plain)")
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")
	traceRedirects := fs.Bool("trace-redirects", false, "Trace redirects like trace-redirects tool")
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
		fmt.Fprintf(os.Stderr, "Usage: header extract [flags] <url>\n")
		fs.PrintDefaults()
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent:         *userAgent,
		TraceRedirects:    *traceRedirects,
		FollowMetaRefresh: *followMetaRefresh,
		UpdateReferer:     *updateReferer,
		StoreCookies:      *storeCookies,
		RedirectLimit:     *redirectLimit,
		RedirectDelay:     *redirectDelay,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	if *traceRedirects && len(info.Trace) > 0 {
		fmt.Printf("Tracing %s\n", urlStr)
		for i, url := range info.Trace {
			fmt.Printf("%d: %s\n", i+1, url)
		}
		fmt.Println("Done")
	}

	tmplContent, err := templates.FS.ReadFile(*style + ".tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading template '%s': %v\n", *style, err)
		return 1
	}

	funcMap := template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		return 1
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
	}
	fmt.Print(buf.String())
	fmt.Println()

	if *testMode {
		diagnostic.Run(info)
	}

	isLintRun := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "lint-list" || f.Name == "lint-rule" {
			isLintRun = true
		}
	})

	if isLintRun || lintRun {
		lists := []string{}
		if *lintListFlags != "" {
			lists = strings.Split(*lintListFlags, ",")
			for i, t := range lists {
				lists[i] = strings.TrimSpace(t)
			}
		}

		rules := []string{}
		if *lintRuleFlags != "" {
			rules = strings.Split(*lintRuleFlags, ",")
			for i, t := range rules {
				rules[i] = strings.TrimSpace(t)
			}
		}

		importLint(info, lists, rules)
	}

	if *showImage {
		images := info.CoverImages()
		if len(images) > 0 {
			ui.DisplayImage(images[0], "default")
		} else {
			fmt.Println("No cover image found to display.")
		}
	}

	return 0
}
