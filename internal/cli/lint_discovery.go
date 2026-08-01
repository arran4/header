package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/arran4/header/lint"
	"github.com/arran4/header/lint/lists"
)

// PrintLists displays the available lint lists.
func PrintLists() {
	fmt.Println("Available Lint Lists:")
	allLists := lists.GetAll()
	for name, rules := range allLists {
		fmt.Printf("  - %s: [%s]\n", name, strings.Join(rules, ", "))
	}
}

// PrintRules displays the available individual lint rules.
func PrintRules() {
	fmt.Println("Available Lint Rules:")
	allRules := lint.GetRegisteredRules()
	for id, rule := range allRules {
		fmt.Printf("  - %s: %s\n", id, rule.Name())
	}
}

// LintRuleListsCmd handles the lint-rule-lists subcommand.
func LintRuleListsCmd(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: header lint-rule-lists <command>\n\nCommands:\n  list        List available lint lists\n  list-rules  List available lint rules\n")
		return 2
	}

	switch args[0] {
	case "list":
		PrintLists()
		return 0
	case "list-rules":
		PrintRules()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s' for lint-rule-lists.\n", args[0])
		return 2
	}
}

// LintRulesCmd handles the lint-rules subcommand.
func LintRulesCmd(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: header lint-rules <command>\n\nCommands:\n  list        List available lint rules\n")
		return 2
	}

	switch args[0] {
	case "list":
		PrintRules()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s' for lint-rules.\n", args[0])
		return 2
	}
}
