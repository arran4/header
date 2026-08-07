package cli

import (
	"fmt"
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

// LintRuleLists is a subcommand `header lint-rule-lists` -- List available lint lists and rules.
func LintRuleLists() error {
	PrintLists()
	fmt.Println()
	PrintRules()
	return nil
}

// LintRules is a subcommand `header lint-rules` -- List available lint rules.
func LintRules() error {
	PrintRules()
	return nil
}
