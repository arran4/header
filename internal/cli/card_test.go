package cli

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/header/internal/cli/cards"
)

func TestCardStdout_OnlyBinaryPNGOnStdout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Test Page</title></head><body></body></html>`))
	}))
	defer ts.Close()

	// Capture stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	outDone := make(chan struct{})
	errDone := make(chan struct{})

	go func() {
		_, _ = io.Copy(&stdoutBuf, rOut)
		close(outDone)
	}()
	go func() {
		_, _ = io.Copy(&stderrBuf, rErr)
		close(errDone)
	}()

	err := CardStdout(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		false,
		"",
		ts.URL,
	)

	_ = wOut.Close()
	_ = wErr.Close()
	<-outDone
	<-errDone

	os.Stdout = origStdout
	os.Stderr = origStderr

	if err != nil {
		t.Fatalf("CardStdout failed: %v", err)
	}

	stdoutBytes := stdoutBuf.Bytes()
	stderrBytes := stderrBuf.Bytes()

	// Verify stdout contains ONLY valid PNG binary data
	if len(stdoutBytes) < 8 {
		t.Fatalf("stdout bytes too short: %d", len(stdoutBytes))
	}
	// PNG signature: \x89PNG\r\n\x1a\n
	pngSig := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if !bytes.HasPrefix(stdoutBytes, pngSig) {
		t.Errorf("stdout does not begin with PNG signature")
	}

	// Verify PNG decodes successfully
	_, err = png.Decode(bytes.NewReader(stdoutBytes))
	if err != nil {
		t.Errorf("stdout is not valid PNG: %v", err)
	}

	// Verify diagnostics appear on stderr, not stdout
	stderrStr := string(stderrBytes)
	if !strings.Contains(stderrStr, "card-cover:") {
		t.Errorf("expected card-cover diagnostic on stderr, got: %q", stderrStr)
	}
}

func TestCardFile_StrictFailureDoesNotOverwriteExistingGoodFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Page has no cover image metadata
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>No Cover</title></head><body></body></html>`))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "card.png")
	originalGoodBytes := []byte("ORIGINAL_GOOD_CARD_CONTENT")
	if err := os.WriteFile(outPath, originalGoodBytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Run with requireCover = true -> MUST fail and MUST NOT touch outPath
	err := CardFile(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		true, // requireCover
		"",
		outPath,
		ts.URL,
	)
	if err == nil {
		t.Fatal("expected error with requireCover=true, got nil")
	}

	// Verify original file is untouched
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Equal(data, originalGoodBytes) {
		t.Errorf("file was modified or overwritten on strict failure! got %q, want %q", data, originalGoodBytes)
	}
}

func TestCardFile_NonStrictPlaceholderSupported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>No Cover</title></head><body></body></html>`))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "card.png")

	err := CardFile(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		false, // requireCover = false (default backward compatible)
		"",
		outPath,
		ts.URL,
	)
	if err != nil {
		t.Fatalf("CardFile failed in non-strict mode: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("output file is not valid PNG: %v", err)
	}
}

func TestCard_QualityReportJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Report Page</title><meta property="og:image" content="http://example.com/bad.png"></head><body></body></html>`))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "quality-report.json")
	outPath := filepath.Join(tempDir, "card.png")

	err := CardFile(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		false,
		reportPath,
		outPath,
		ts.URL,
	)
	if err != nil {
		t.Fatalf("CardFile failed: %v", err)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	var report cards.CardQualityReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("failed to parse quality report JSON: %v", err)
	}

	if report.InputURL != ts.URL {
		t.Errorf("InputURL = %q, want %q", report.InputURL, ts.URL)
	}
	if report.Generator != "basic" {
		t.Errorf("Generator = %q, want 'basic'", report.Generator)
	}
	if !report.UsedPlaceholder {
		t.Errorf("UsedPlaceholder = false, want true")
	}
	if report.CoverStatus == "" {
		t.Errorf("CoverStatus is empty")
	}
}

func TestCard_QualityReportWriteFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Page</title></head><body></body></html>`))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "card.png")
	invalidReportPath := filepath.Join(tempDir, "nonexistent-dir-xyz", "report.json")

	err := CardFile(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		false,
		invalidReportPath,
		outPath,
		ts.URL,
	)
	if err == nil {
		t.Fatalf("expected error when quality report destination is invalid, got nil")
	}
	if !strings.Contains(err.Error(), "quality report") {
		t.Errorf("error %q should mention quality report failure", err.Error())
	}
}

func TestCard_SensitiveURLRedactionInQualityReport(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Secret Page</title>
<meta property="og:image" content="http://imguser:imgpassword@` + r.Host + `/img?sig=sec123">
</head><body></body></html>`))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "report.json")
	outPath := filepath.Join(tempDir, "card.png")

	credentialURL := strings.Replace(ts.URL, "http://", "http://loginuser:loginpass@", 1) + "/page?token=auth123"

	err := CardFile(
		"Mozilla/5.0",
		"basic",
		1,
		5*time.Millisecond,
		20*time.Millisecond,
		1*time.Second,
		1*time.Second,
		false,
		reportPath,
		outPath,
		credentialURL,
	)
	if err != nil {
		t.Fatalf("CardFile failed: %v", err)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed reading report: %v", err)
	}
	reportStr := string(reportBytes)

	for _, secret := range []string{"loginpass", "auth123", "imgpassword", "sec123"} {
		if strings.Contains(reportStr, secret) {
			t.Errorf("quality report leaked secret %q:\n%s", secret, reportStr)
		}
	}
}
