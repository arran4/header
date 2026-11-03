package header

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// RetryPolicy specifies the bounds for retrying idempotent HTTP GET requests.
type RetryPolicy struct {
	Attempts int           // Total attempts (e.g. 3; 1 means initial request only, no retries)
	Delay    time.Duration // Initial delay between retries (e.g. 250ms)
	MaxDelay time.Duration // Maximum delay cap (e.g. 5s)
}

// DefaultRetryPolicy returns the standard default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		Attempts: 3,
		Delay:    250 * time.Millisecond,
		MaxDelay: 5 * time.Second,
	}
}

// ParseRetryAfter parses an HTTP Retry-After header value, which can be either
// a non-negative decimal integer representing seconds, or an HTTP-date.
func ParseRetryAfter(headerVal string, now time.Time) (time.Duration, bool) {
	headerVal = strings.TrimSpace(headerVal)
	if headerVal == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(headerVal); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	if t, err := http.ParseTime(headerVal); err == nil {
		d := t.Sub(now)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}

// IsRetryableStatus returns true for transient HTTP response status codes.
// Retry 408 (Request Timeout), 429 (Too Many Requests), 500, 502, 503, 504.
func IsRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:       // 504
		return true
	default:
		return false
	}
}

// IsPermanentStatus returns true for permanent HTTP error responses that should not be retried.
func IsPermanentStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusBadRequest, // 400
		http.StatusUnauthorized, // 401
		http.StatusForbidden,    // 403
		http.StatusNotFound,     // 404
		http.StatusMethodNotAllowed: // 405
		return true
	default:
		return false
	}
}

// IsRetryableError returns true if the network error is transient and eligible for retry.
func IsRetryableError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx != nil && (ctx.Err() != nil || errors.Is(err, context.Canceled)) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		if ctx != nil && ctx.Err() != nil {
			return false
		}
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.EOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "closed network connection") ||
		strings.Contains(msg, "server closed idle connection") {
		return true
	}
	return false
}

// CalculateBackoff computes the exponential backoff duration with jitter,
// respecting Retry-After where available, bounded by maxDelay.
func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration, retryAfter string, now time.Time) time.Duration {
	if baseDelay <= 0 {
		baseDelay = 250 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 5 * time.Second
	}
	if d, ok := ParseRetryAfter(retryAfter, now); ok {
		if d > maxDelay {
			return maxDelay
		}
		return d
	}

	shift := attempt - 1
	if shift < 0 {
		shift = 0
	}
	if shift > 10 {
		shift = 10
	}
	delay := baseDelay * (1 << shift)
	jitterLimit := int64(delay / 4)
	if jitterLimit > 0 {
		jitter := time.Duration(rand.Int63n(jitterLimit))
		delay += jitter
	}
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// HTTPDoer is an interface for making HTTP requests (e.g. *http.Client).
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ExecuteRequestWithRetry executes an HTTP request using the given retry policy.
// It returns the response (if any), the number of attempts made, and any terminal error.
func ExecuteRequestWithRetry(ctx context.Context, client HTTPDoer, req *http.Request, policy RetryPolicy) (*http.Response, int, error) {
	if policy.Attempts <= 0 {
		policy.Attempts = 3
	}
	if policy.Delay <= 0 {
		policy.Delay = 250 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = 5 * time.Second
	}

	attempts := 0
	for {
		attempts++
		if ctx != nil && ctx.Err() != nil {
			return nil, attempts, ctx.Err()
		}

		attemptReq := req
		if ctx != nil {
			attemptReq = req.Clone(ctx)
		}

		resp, err := client.Do(attemptReq)

		shouldRetry := false
		var retryDelay time.Duration

		if err != nil {
			if ctx != nil && ctx.Err() != nil {
				return nil, attempts, ctx.Err()
			}
			if attempts < policy.Attempts && IsRetryableError(ctx, err) {
				shouldRetry = true
				retryDelay = CalculateBackoff(attempts, policy.Delay, policy.MaxDelay, "", time.Now())
			} else {
				return nil, attempts, err
			}
		} else {
			if IsRetryableStatus(resp.StatusCode) && attempts < policy.Attempts {
				shouldRetry = true
				retryAfter := resp.Header.Get("Retry-After")
				_ = resp.Body.Close()
				retryDelay = CalculateBackoff(attempts, policy.Delay, policy.MaxDelay, retryAfter, time.Now())
			} else {
				return resp, attempts, nil
			}
		}

		if shouldRetry {
			if ctx != nil {
				select {
				case <-ctx.Done():
					return nil, attempts, ctx.Err()
				case <-time.After(retryDelay):
				}
			} else {
				time.Sleep(retryDelay)
			}
		}
	}
}
