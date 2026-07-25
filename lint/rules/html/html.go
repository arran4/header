package html

import (
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "html"

type htmlRule struct{}

func (r htmlRule) ID() string   { return RuleID }
func (r htmlRule) Name() string { return "HTML Good Practices" }

func (r htmlRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	if info.Title == "" {
		results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Missing HTML <title>"})
	}

	// Icon checks
	hasIconSizes := false
	hasStandardPWASizes := false // 192x192 or 512x512

	for _, icon := range info.Icons {
		if icon.Sizes != "" {
			hasIconSizes = true
			if strings.Contains(icon.Sizes, "192x192") || strings.Contains(icon.Sizes, "512x512") {
				hasStandardPWASizes = true
			}
		}
	}

	if len(info.Icons) > 0 && !hasIconSizes {
		results = append(results, lint.Result{Severity: lint.SeverityInfo, Message: "Icons present but none specify 'sizes' attribute"})
	}

	if len(info.Icons) > 0 && !hasStandardPWASizes {
		results = append(results, lint.Result{Severity: lint.SeverityInfo, Message: "Standard PWA icon sizes (192x192, 512x512) are missing"})
	}

	// Basic mime type format checks
	for _, altLink := range info.AlternateLinks {
		if altLink.Media != "" {
			if !strings.Contains(altLink.Media, "/") && !strings.Contains(altLink.Media, "(") {
				results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: "Alternate link has an unusual media/type attribute format: " + altLink.Media})
			}
		}
	}

	return results
}

func init() {
	lint.RegisterRule(htmlRule{})
}
