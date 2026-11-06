package header

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestRetry_PostNeverRetried(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader("payload"))
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    5 * time.Millisecond,
		MaxDelay: 20 * time.Millisecond,
	}

	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (POST must never be retried)", attempts)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("server requests = %d, want 1", atomic.LoadInt32(&count))
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}

func TestRetry_ImmediateSuccess(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    10 * time.Millisecond,
		MaxDelay: 50 * time.Millisecond,
	}

	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("server requests = %d, want 1", atomic.LoadInt32(&count))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRetry_Transient500ThenSuccess(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reqNum := atomic.AddInt32(&count, 1)
		if reqNum == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("recovered"))
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    5 * time.Millisecond,
		MaxDelay: 20 * time.Millisecond,
	}

	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("server requests = %d, want 2", atomic.LoadInt32(&count))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRetry_429WithRetryAfter(t *testing.T) {
	t.Run("seconds", func(t *testing.T) {
		var count int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			reqNum := atomic.AddInt32(&count, 1)
			if reqNum == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
		policy := RetryPolicy{
			Attempts: 3,
			Delay:    5 * time.Millisecond,
			MaxDelay: 20 * time.Millisecond,
		}

		resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if attempts != 2 {
			t.Errorf("attempts = %d, want 2", attempts)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("http-date", func(t *testing.T) {
		var count int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			reqNum := atomic.AddInt32(&count, 1)
			if reqNum == 1 {
				// Past date means 0 duration
				w.Header().Set("Retry-After", time.Now().Add(-1*time.Minute).UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
		policy := RetryPolicy{
			Attempts: 3,
			Delay:    5 * time.Millisecond,
			MaxDelay: 20 * time.Millisecond,
		}

		resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if attempts != 2 {
			t.Errorf("attempts = %d, want 2", attempts)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})
}

func TestRetry_TimeoutThenRetry(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reqNum := atomic.AddInt32(&count, 1)
		if reqNum == 1 {
			time.Sleep(60 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := ts.Client()
	client.Timeout = 30 * time.Millisecond

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    5 * time.Millisecond,
		MaxDelay: 20 * time.Millisecond,
	}

	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), client, req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// retryDoerFunc lets retry tests control request outcomes without relying on
// HTTP transport connection scheduling or a raw TCP server.
type retryDoerFunc func(*http.Request) (*http.Response, error)

func (f retryDoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRetry_ConnectionFailureThenRetry(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "EOF", err: io.EOF},
		{name: "connection reset", err: syscall.ECONNRESET},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := retryDoerFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return nil, &url.Error{Op: "Get", URL: req.URL.String(), Err: tc.err}
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})

			req, err := http.NewRequest(http.MethodGet, "http://example.test", nil)
			if err != nil {
				t.Fatal(err)
			}
			policy := RetryPolicy{
				Attempts: 3,
				Delay:    5 * time.Millisecond,
				MaxDelay: 20 * time.Millisecond,
			}

			resp, attempts, err := ExecuteRequestWithRetry(context.Background(), client, req, policy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if attempts != 2 {
				t.Errorf("attempts = %d, want 2", attempts)
			}
			if calls != 2 {
				t.Errorf("client calls = %d, want 2", calls)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want 200", resp.StatusCode)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if string(body) != "ok" {
				t.Errorf("body = %q, want ok", body)
			}
		})
	}
}

func TestRetry_Permanent403And404WithoutRetries(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		t.Run(fmt.Sprintf("status-%d", status), func(t *testing.T) {
			var count int32
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&count, 1)
				w.WriteHeader(status)
			}))
			defer ts.Close()

			req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
			policy := RetryPolicy{
				Attempts: 3,
				Delay:    5 * time.Millisecond,
				MaxDelay: 20 * time.Millisecond,
			}

			resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if attempts != 1 {
				t.Errorf("attempts = %d, want 1 (no retries for permanent errors)", attempts)
			}
			if atomic.LoadInt32(&count) != 1 {
				t.Errorf("server requests = %d, want 1", atomic.LoadInt32(&count))
			}
			if resp.StatusCode != status {
				t.Errorf("status = %d, want %d", resp.StatusCode, status)
			}
		})
	}
}

func TestRetry_RetryExhaustion(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    5 * time.Millisecond,
		MaxDelay: 15 * time.Millisecond,
	}

	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("server requests = %d, want 3", atomic.LoadInt32(&count))
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}

func TestRetry_CancellationDuringBackoff(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after short delay so it cancels during backoff
	go func() {
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 5,
		Delay:    500 * time.Millisecond, // Long backoff to ensure cancel happens during sleep
		MaxDelay: 1 * time.Second,
	}

	start := time.Now()
	_, attempts, err := ExecuteRequestWithRetry(ctx, ts.Client(), req, policy)
	duration := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if duration > 200*time.Millisecond {
		t.Errorf("cancellation took too long: %v", duration)
	}
}

func TestFetchURL_RedirectAndRelativeURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			w.Header().Set("Location", "/final-page")
			w.WriteHeader(http.StatusFound)
		case "/final-page":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Final Page Title</title></head><body>Hello</body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	opts := FetchOptions{
		TraceRedirects: true,
		RetryAttempts:  3,
		RetryDelay:     5 * time.Millisecond,
		RetryMaxDelay:  15 * time.Millisecond,
		Timeout:        1 * time.Second,
	}

	info, err := FetchURL(ts.URL+"/start", opts)
	if err != nil {
		t.Fatalf("FetchURL error: %v", err)
	}

	if info.Title != "Final Page Title" {
		t.Errorf("Title = %q, want 'Final Page Title'", info.Title)
	}
	if info.URL != ts.URL+"/final-page" {
		t.Errorf("URL = %q, want %q", info.URL, ts.URL+"/final-page")
	}
	if len(info.Trace) < 2 {
		t.Errorf("Trace length = %d, want >= 2", len(info.Trace))
	}
}

func TestCalculateBackoff_RetryAfterNotCapped(t *testing.T) {
	// Server requested 60s, maxDelay is 5s.
	// CalculateBackoff must return 60s rather than capping to 5s.
	now := time.Now()
	delay := CalculateBackoff(1, 250*time.Millisecond, 5*time.Second, "60", now)
	if delay != 60*time.Second {
		t.Errorf("delay = %v, want 60s", delay)
	}
}

func TestRetry_429WithRetryAfterExceedingMaxDelay(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.Header().Set("Retry-After", "120") // 120 seconds
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    10 * time.Millisecond,
		MaxDelay: 50 * time.Millisecond, // 50ms budget < 120s
	}

	start := time.Now()
	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), ts.Client(), req, policy)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (must not retry prematurely)", attempts)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("server requests = %d, want 1", atomic.LoadInt32(&count))
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", resp.StatusCode)
	}
	if duration > 1*time.Second {
		t.Errorf("execution took %v, should have returned immediately", duration)
	}
}

func TestFetchURL_HTTPStatuses(t *testing.T) {
	var requests403 int32
	var requests404 int32
	var requests503 int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status-403":
			atomic.AddInt32(&requests403, 1)
			http.Error(w, "Forbidden", http.StatusForbidden)
		case "/status-404":
			atomic.AddInt32(&requests404, 1)
			http.NotFound(w, r)
		case "/status-503":
			atomic.AddInt32(&requests503, 1)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
	}))
	defer ts.Close()

	opts := FetchOptions{
		RetryAttempts: 3,
		RetryDelay:    5 * time.Millisecond,
		RetryMaxDelay: 20 * time.Millisecond,
		Timeout:       1 * time.Second,
	}

	t.Run("403 Forbidden terminal error without retrying", func(t *testing.T) {
		_, err := FetchURL(ts.URL+"/status-403", opts)
		if err == nil {
			t.Fatalf("expected error for 403, got nil")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("error %q should mention 403", err.Error())
		}
		if atomic.LoadInt32(&requests403) != 1 {
			t.Errorf("403 requests = %d, want 1 (no retries)", atomic.LoadInt32(&requests403))
		}
	})

	t.Run("404 Not Found terminal error without retrying", func(t *testing.T) {
		_, err := FetchURL(ts.URL+"/status-404", opts)
		if err == nil {
			t.Fatalf("expected error for 404, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error %q should mention 404", err.Error())
		}
		if atomic.LoadInt32(&requests404) != 1 {
			t.Errorf("404 requests = %d, want 1 (no retries)", atomic.LoadInt32(&requests404))
		}
	})

	t.Run("503 retries exhausted returns error", func(t *testing.T) {
		_, err := FetchURL(ts.URL+"/status-503", opts)
		if err == nil {
			t.Fatalf("expected error for exhausted 503, got nil")
		}
		if !strings.Contains(err.Error(), "503") {
			t.Errorf("error %q should mention 503", err.Error())
		}
		if atomic.LoadInt32(&requests503) != 3 {
			t.Errorf("503 requests = %d, want 3 attempts", atomic.LoadInt32(&requests503))
		}
	})
}

func TestFetchURL_FinalURLAndRelativeAssetResolution(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/entry":
			http.Redirect(w, r, "/nested/section/document.html", http.StatusMovedPermanently)
		case "/nested/section/document.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>Nested Document</title>
  <meta property="og:image" content="../images/cover.jpg">
  <link rel="icon" href="favicon.ico">
</head>
<body>Nested</body>
</html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	opts := FetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	info, err := FetchURL(ts.URL+"/entry", opts)
	if err != nil {
		t.Fatalf("FetchURL error: %v", err)
	}

	expectedFinalURL := ts.URL + "/nested/section/document.html"
	if info.URL != expectedFinalURL {
		t.Errorf("info.URL = %q, want %q", info.URL, expectedFinalURL)
	}

	// Verify relative OG image resolves against the final document URL:
	// ../images/cover.jpg resolved against /nested/section/document.html -> /nested/images/cover.jpg
	expectedCover := ts.URL + "/nested/images/cover.jpg"
	coverImages := info.CoverImages()
	if len(coverImages) == 0 {
		t.Fatalf("expected at least one cover image")
	}
	// Note: info.OGImage contains the raw metadata attribute ("../images/cover.jpg"),
	// ResolveAssetURL or resolving against info.URL gives the correct resolved URL.
	docURL, _ := url.Parse(info.URL)
	refURL, _ := url.Parse(info.OGImage)
	resolved := docURL.ResolveReference(refURL).String()
	if resolved != expectedCover {
		t.Errorf("resolved relative cover = %q, want %q", resolved, expectedCover)
	}
}

func TestFetchURL_CrossHostRedirect(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/destination" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Target Host</title></head></html>`))
			return
		}
		http.NotFound(w, r)
	}))
	defer targetServer.Close()

	originServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL+"/destination", http.StatusFound)
	}))
	defer originServer.Close()

	opts := FetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	info, err := FetchURL(originServer.URL, opts)
	if err != nil {
		t.Fatalf("FetchURL cross-host error: %v", err)
	}

	if info.URL != targetServer.URL+"/destination" {
		t.Errorf("info.URL = %q, want %q", info.URL, targetServer.URL+"/destination")
	}
	if info.Title != "Target Host" {
		t.Errorf("info.Title = %q, want 'Target Host'", info.Title)
	}
}

func TestFetchURL_TotalDeadlineEnforcement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	opts := FetchOptions{
		Timeout:       50 * time.Millisecond, // Total deadline is 50ms
		RetryAttempts: 3,
		RetryDelay:    10 * time.Millisecond,
	}

	start := time.Now()
	_, err := FetchURL(ts.URL, opts)
	duration := time.Since(start)

	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if duration > 120*time.Millisecond {
		t.Errorf("FetchURL exceeded total deadline: took %v", duration)
	}
}
