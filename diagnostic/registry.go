package diagnostic

import (
	"fmt"
	"github.com/arran4/header"
)

// Severity indicates the level of a diagnostic result.
type Severity int

const (
	// SeverityInfo is an informational diagnostic.
	SeverityInfo Severity = iota
	// SeverityWarning is a warning diagnostic.
	SeverityWarning
	// SeverityError is an error diagnostic.
	SeverityError
)

// Result is a single diagnostic result.
type Result struct {
	Severity Severity
	Message  string
}

// Rule is an interface for checking a PageInfo for issues.
type Rule interface {
	Name() string
	Check(info *header.PageInfo) []Result
}

var registry []Rule

// Register adds a rule to the diagnostic registry.
func Register(rule Rule) {
	registry = append(registry, rule)
}

// Run executes all registered diagnostic rules against the given PageInfo.
func Run(info *header.PageInfo) {
	fmt.Println("--- Diagnostics ---")

	totalIssues := 0
	for _, rule := range registry {
		results := rule.Check(info)
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
		fmt.Println("All common metadata fields are present and no issues detected.")
	}
	fmt.Println("-------------------")
}
