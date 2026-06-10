package main

import (
	"fmt"
	"net/url"

	"github.com/arran4/header"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

type DiagnosticResult struct {
	Severity Severity
	Message  string
}

type DiagnosticRule interface {
	Name() string
	Check(info *header.PageInfo) []DiagnosticResult
}

var diagnosticRegistry []DiagnosticRule

func RegisterDiagnosticRule(rule DiagnosticRule) {
	diagnosticRegistry = append(diagnosticRegistry, rule)
}

type MissingFieldRule struct{}

func (r MissingFieldRule) Name() string { return "MissingFieldCheck" }

func (r MissingFieldRule) Check(info *header.PageInfo) []DiagnosticResult {
	var results []DiagnosticResult

	if info.Title == "" {
		results = append(results, DiagnosticResult{SeverityError, "Missing Title"})
	}
	if info.OGType == "" {
		results = append(results, DiagnosticResult{SeverityWarning, "Missing OpenGraph Data"})
	}
	if info.TwitterCard == "" {
		results = append(results, DiagnosticResult{SeverityWarning, "Missing Twitter Card Data"})
	}
	if info.JSONLDType == "" {
		results = append(results, DiagnosticResult{SeverityWarning, "Missing JSON-LD Metadata"})
	}
	if len(info.CoverImages()) == 0 {
		results = append(results, DiagnosticResult{SeverityWarning, "No Cover Images found"})
	}

	return results
}

type YouTubeRule struct{}

func (r YouTubeRule) Name() string { return "YouTubeCheck" }

func (r YouTubeRule) Check(info *header.PageInfo) []DiagnosticResult {
	var results []DiagnosticResult

	u, err := url.Parse(info.URL)
	if err != nil {
		results = append(results, DiagnosticResult{SeverityError, fmt.Sprintf("Failed to parse URL: %v", err)})
		return results
	}

	hostname := u.Hostname()
	if hostname == "youtube.com" || hostname == "www.youtube.com" || hostname == "youtu.be" {
		if info.YouTubeVideoID == "" {
			results = append(results, DiagnosticResult{SeverityWarning, "YouTube URL detected but failed to extract Video ID/Watch time."})
		}
	}

	return results
}

func init() {
	RegisterDiagnosticRule(MissingFieldRule{})
	RegisterDiagnosticRule(YouTubeRule{})
}

func runDiagnostics(info *header.PageInfo) {
	fmt.Println("--- Diagnostics ---")

	totalIssues := 0
	for _, rule := range diagnosticRegistry {
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
