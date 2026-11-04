package header

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var urlRegex = regexp.MustCompile(`https?://[^\s"'<>]+`)

// SanitizeURL scrubs userinfo (passwords and usernames) and sensitive query parameters from a URL.
func SanitizeURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if u.User != nil {
		if _, hasPass := u.User.Password(); hasPass {
			u.User = url.UserPassword("REDACTED", "REDACTED")
		} else {
			u.User = url.User("REDACTED")
		}
	}

	if u.RawQuery != "" {
		q := u.Query()
		modified := false
		for k := range q {
			lk := strings.ToLower(k)
			if isSensitiveParam(lk) {
				q.Set(k, "REDACTED")
				modified = true
			}
		}
		if modified {
			u.RawQuery = q.Encode()
		}
	}

	return u.String()
}

func isSensitiveParam(lowerKey string) bool {
	sensitiveKeys := []string{
		"token", "secret", "key", "auth", "password", "pass",
		"signature", "sig", "credential", "cred", "access",
		"bearer", "apikey", "api_key", "session",
	}
	for _, s := range sensitiveKeys {
		if strings.Contains(lowerKey, s) {
			return true
		}
	}
	return false
}

// SanitizeErrorMessage scrubs URLs with sensitive parameters or userinfo from error messages.
func SanitizeErrorMessage(msg string) string {
	if msg == "" {
		return ""
	}
	return urlRegex.ReplaceAllStringFunc(msg, func(matched string) string {
		return SanitizeURL(matched)
	})
}

// SanitizeError returns a sanitized error message string from err,
// ensuring user credentials and sensitive tokens are not leaked.
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		sanitizedURL := SanitizeURL(urlErr.URL)
		var innerErrStr string
		if urlErr.Err != nil {
			innerErrStr = SanitizeErrorMessage(urlErr.Err.Error())
		}
		if innerErrStr != "" {
			return fmt.Sprintf("%s %q: %s", urlErr.Op, sanitizedURL, innerErrStr)
		}
		return fmt.Sprintf("%s %q", urlErr.Op, sanitizedURL)
	}
	return SanitizeErrorMessage(err.Error())
}
