package diagnostic

import (
	"fmt"
	"github.com/arran4/header"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

type Result struct {
	Severity Severity
	Message  string
}

type Rule interface {
	Name() string
	Check(info *header.PageInfo) []Result
}

var registry []Rule

func Register(rule Rule) {
	registry = append(registry, rule)
}

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
