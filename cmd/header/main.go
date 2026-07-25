package main

import (
	"fmt"
	"os"

	"github.com/arran4/header/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: header <command> [args...]\n")
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  show            Extract metadata from URL (default if no matching command is found)\n")
		fmt.Fprintf(os.Stderr, "  edit            Edit HTML file head/meta fields\n")
		fmt.Fprintf(os.Stderr, "  trace-redirects Trace URL redirects\n")
		fmt.Fprintf(os.Stderr, "  diagnose         Not implemented\n")
		fmt.Fprintf(os.Stderr, "  debug            Not implemented\n")
		fmt.Fprintf(os.Stderr, "  verify           Not implemented\n")
		fmt.Fprintf(os.Stderr, "  lint             Lint the metadata of a URL\n")
		fmt.Fprintf(os.Stderr, "  lint-rule-lists  List available lint rule lists\n")
		fmt.Fprintf(os.Stderr, "  lint-rules       List available lint rules\n")
		fmt.Fprintf(os.Stderr, "  view             Not implemented (e.g. view image <url>)\n")
		os.Exit(2)
	}

	command := os.Args[1]

	switch command {
	case "edit":
		os.Exit(cli.EditCmd(os.Args[2:]))
	case "show", "extract":
		os.Exit(cli.ExtractCmd(os.Args[2:]))
	case "trace-redirects":
		os.Exit(cli.TraceRedirectsCmd(os.Args[2:]))
	case "lint":
		os.Exit(cli.LintCmd(os.Args[2:]))
	case "lint-rule-lists":
		os.Exit(cli.LintRuleListsCmd(os.Args[2:]))
	case "lint-rules":
		os.Exit(cli.LintRulesCmd(os.Args[2:]))
	case "diagnose", "debug", "verify", "view":
		fmt.Fprintf(os.Stderr, "Command '%s' is not implemented yet.\n", command)
		os.Exit(1)
	default:
		// To remain somewhat backwards compatible with old usage where people did `header https://...`,
		// if the command isn't recognized, fallback to extract treating the command as an arg.
		// e.g. `header -test https://example.com`
		os.Exit(cli.ExtractCmd(os.Args[1:]))
	}
}
