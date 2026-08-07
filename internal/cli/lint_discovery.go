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

// LintRuleListsList is a subcommand `header lint-rule-lists list` -- List available lint lists.
func LintRuleListsList() error {
	PrintLists()
	return nil
}

// LintRuleListsListRules is a subcommand `header lint-rule-lists list-rules` -- List available lint rules.
func LintRuleListsListRules() error {
	PrintRules()
	return nil
}

// LintRulesList is a subcommand `header lint-rules list` -- List available lint rules.
func LintRulesList() error {
	PrintRules()
	return nil
}
