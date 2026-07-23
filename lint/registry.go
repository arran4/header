package lint

import (
	"fmt"
	"github.com/arran4/header"
)

// Severity indicates the level of a lint result.
type Severity int

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
	Name() string
	Tags() []string
	Check(info *header.PageInfo) []Result
}

var registry []Rule

// Register adds a rule to the lint registry.
func Register(rule Rule) {
	registry = append(registry, rule)
}

// Run executes all registered lint rules matching the given tags against the given PageInfo.
func Run(info *header.PageInfo, tags []string) {
	fmt.Println("--- Lint ---")

	totalIssues := 0
	for _, rule := range registry {
		if !matchTags(rule.Tags(), tags) {
			continue
		}
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
		fmt.Println("No lint issues detected.")
	}
	fmt.Println("------------")
}

func matchTags(ruleTags []string, userTags []string) bool {
	// If the user specified no tags or just "", it means they want default rules
	wantsDefault := false
	if len(userTags) == 0 {
		wantsDefault = true
	} else {
		for _, ut := range userTags {
			if ut == "" || ut == "default" {
				wantsDefault = true
				break
			}
		}
	}

	// Check if this rule is a default rule (has "" tag)
	isDefaultRule := false
	if len(ruleTags) == 0 {
		isDefaultRule = true
	} else {
		for _, rt := range ruleTags {
			if rt == "" || rt == "default" {
				isDefaultRule = true
				break
			}
		}
	}

	// If the user requested default rules and this is a default rule, match.
	if wantsDefault && isDefaultRule {
		return true
	}

	// Otherwise, check for specific tag overlap
	for _, ut := range userTags {
		if ut == "" || ut == "default" {
			continue // handled above
		}
		for _, rt := range ruleTags {
			if ut == rt {
				return true
			}
		}
	}

	return false
}
