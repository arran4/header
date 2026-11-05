package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/header"
)

func TestDisplayImageWithOptions_UnsupportedViewer(t *testing.T) {
	err := DisplayImageWithOptions("http://example.com/test.jpg", "invalid-viewer", header.FetchOptions{})
	if err == nil {
		t.Fatal("expected error for unsupported viewer, got nil")
	}
	if !strings.Contains(err.Error(), `unsupported viewer "invalid-viewer" requested`) {
		t.Errorf("error = %q, want message containing unsupported viewer", err.Error())
	}
}

func TestDisplayImageWithOptions_PrintURL(t *testing.T) {
	err := DisplayImageWithOptions("http://example.com/test.jpg", "print-url", header.FetchOptions{})
	if err != nil {
		t.Fatalf("expected nil error for print-url, got: %v", err)
	}
}

func TestDisplayImageWithOptions_MockViewerSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = fmt.Fprint(w, "fake image bytes")
	}))
	defer ts.Close()

	opts := header.FetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}
	err := DisplayImageWithOptions(ts.URL+"/img.jpg", "mock", opts)
	if err != nil {
		t.Fatalf("expected nil error for mock with 200 OK, got: %v", err)
	}
}

func TestDisplayImageWithOptions_MockViewerFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer ts.Close()

	opts := header.FetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}
	err := DisplayImageWithOptions(ts.URL+"/notfound.jpg", "mock", opts)
	if err == nil {
		t.Fatal("expected error for mock viewer on 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want 404 mentioned", err.Error())
	}
}
