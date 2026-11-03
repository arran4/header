package header

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

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

func TestRetry_ConnectionFailureThenRetry(t *testing.T) {
	var count int32
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			reqNum := atomic.AddInt32(&count, 1)
			if reqNum == 1 {
				// Close connection immediately to simulate reset/EOF
				_ = conn.Close()
			} else {
				// Send minimal HTTP 200 response
				_, _ = fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nok")
				_ = conn.Close()
			}
		}
	}()

	urlStr := fmt.Sprintf("http://%s", l.Addr().String())
	req, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	policy := RetryPolicy{
		Attempts: 3,
		Delay:    5 * time.Millisecond,
		MaxDelay: 20 * time.Millisecond,
	}

	client := &http.Client{Timeout: 1 * time.Second}
	resp, attempts, err := ExecuteRequestWithRetry(context.Background(), client, req, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts < 2 {
		t.Errorf("attempts = %d, want >= 2", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
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
	if len(info.Trace) < 2 {
		t.Errorf("Trace length = %d, want >= 2", len(info.Trace))
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
