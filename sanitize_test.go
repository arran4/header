package header

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "plain URL without secrets",
			input:    "https://example.com/path/to/page.html?view=grid",
			expected: "https://example.com/path/to/page.html?view=grid",
		},
		{
			name:     "user and password in userinfo",
			input:    "https://alice:supersecret@example.com/api/image.png",
			expected: "https://REDACTED:REDACTED@example.com/api/image.png",
		},
		{
			name:     "token only as username in userinfo",
			input:    "https://ghp_secrettoken123@github.com/repo/image.png",
			expected: "https://REDACTED@github.com/repo/image.png",
		},
		{
			name:     "sensitive query parameters",
			input:    "https://example.com/img.jpg?access_token=secret123&category=nature&api_key=key456",
			expected: "https://example.com/img.jpg?access_token=REDACTED&api_key=REDACTED&category=nature",
		},
		{
			name:     "AWS S3 signed query parameters",
			input:    "https://bucket.s3.amazonaws.com/photo.jpg?X-Amz-Signature=sig123&X-Amz-Credential=cred456&X-Amz-Security-Token=tok789&id=1",
			expected: "https://bucket.s3.amazonaws.com/photo.jpg?X-Amz-Credential=REDACTED&X-Amz-Security-Token=REDACTED&X-Amz-Signature=REDACTED&id=1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeURL(tc.input)
			if got != tc.expected {
				t.Errorf("SanitizeURL(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if got := SanitizeError(nil); got != "" {
			t.Errorf("expected empty string for nil error, got %q", got)
		}
	})

	t.Run("url.Error with credentials", func(t *testing.T) {
		err := &url.Error{
			Op:  "Get",
			URL: "https://user:password123@example.com/private/cover.png?token=secrettoken",
			Err: errors.New("dial tcp: lookup example.com: no such host"),
		}
		sanitized := SanitizeError(err)

		if strings.Contains(sanitized, "password123") {
			t.Errorf("sanitized error leaked password: %q", sanitized)
		}
		if strings.Contains(sanitized, "secrettoken") {
			t.Errorf("sanitized error leaked token: %q", sanitized)
		}
		if !strings.Contains(sanitized, "example.com/private/cover.png") {
			t.Errorf("sanitized error lost path context: %q", sanitized)
		}
		if !strings.Contains(sanitized, "no such host") {
			t.Errorf("sanitized error lost root cause: %q", sanitized)
		}
	})

	t.Run("generic error with embedded URL", func(t *testing.T) {
		err := fmt.Errorf("failed fetching https://api_key_user:pass@service.internal/hook?sig=abcd1234: connection reset by peer")
		sanitized := SanitizeError(err)

		if strings.Contains(sanitized, "pass") && strings.Contains(sanitized, "api_key_user") {
			t.Errorf("leaked credentials: %q", sanitized)
		}
		if strings.Contains(sanitized, "abcd1234") {
			t.Errorf("leaked signature: %q", sanitized)
		}
		if !strings.Contains(sanitized, "connection reset by peer") {
			t.Errorf("lost diagnostic message: %q", sanitized)
		}
	})
}
