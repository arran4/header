// Package lint provides a framework for analyzing extracted HTML metadata.
package lint

import (
	"fmt"
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/lint/lists"
)

// Severity indicates the level of a lint result.
type Severity int

// The available severities for a lint result.
const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// Result is a single lint result.
type Result struct {
	Severity Severity
	Message  string
}

// Rule is an interface for checking a PageInfo for issues.
type Rule interface {
	ID() string
	Name() string
	Check(info *header.PageInfo, ops ...any) []Result
}

var (
	ruleRegistry = make(map[string]Rule)
)

// RegisterRule adds a rule to the lint registry.
func RegisterRule(rule Rule) {
	ruleRegistry[rule.ID()] = rule
}

// GetRegisteredRules returns all registered rules in the registry.
func GetRegisteredRules() map[string]Rule {
	ret := make(map[string]Rule)
	for k, v := range ruleRegistry {
		ret[k] = v
	}
	return ret
}

// Run executes specific rules and/or lists of rules against the given PageInfo.
func Run(info *header.PageInfo, listNames []string, ruleIDs []string, ops ...any) {
	fmt.Println("--- Lint ---")

	rulesToRun := make(map[string]Rule)

	// Resolve lists
	for _, ln := range listNames {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if ids, ok := lists.Get(ln); ok {
			for _, id := range ids {
				if r, exists := ruleRegistry[id]; exists {
					rulesToRun[id] = r
				} else {
					fmt.Printf("[WARN] List '%s' references unknown rule '%s'\n", ln, id)
				}
			}
		} else {
			fmt.Printf("[WARN] Unknown rule list '%s'\n", ln)
		}
	}

	// Resolve specific rules
	for _, id := range ruleIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if r, ok := ruleRegistry[id]; ok {
			rulesToRun[id] = r
		} else {
			fmt.Printf("[WARN] Unknown rule '%s'\n", id)
		}
	}

	// Default behavior if nothing specified
	if len(listNames) == 0 && len(ruleIDs) == 0 {
		// Run 'recommended' list if it exists, otherwise run everything for backward compatibility / fallback
		if ids, ok := lists.Get("recommended"); ok {
			for _, id := range ids {
				if r, exists := ruleRegistry[id]; exists {
					rulesToRun[id] = r
				}
			}
		} else {
			// No recommended list defined, run all rules
			for id, r := range ruleRegistry {
				rulesToRun[id] = r
			}
		}
	}

	totalIssues := 0
	for _, rule := range rulesToRun {
		results := rule.Check(info, ops...)
		if len(results) > 0 {
			for _, res := range results {
				totalIssues++
				severityStr := "INFO"
				switch res.Severity {
				case SeverityWarning:
					severityStr = "WARN"
				case SeverityError:
					severityStr = "ERROR"
				}
				fmt.Printf("[%s] %s: %s\n", severityStr, rule.Name(), res.Message)
			}
		}
	}

	if totalIssues == 0 {
		fmt.Println("No lint issues detected.")
	}
	fmt.Println("------------")
}
