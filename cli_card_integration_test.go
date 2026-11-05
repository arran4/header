package header_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/arran4/header/internal/cli/cards"
)

func createValidPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{100, 150, 200, 255}}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestCLI_CardStdout_BinaryOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>CLI Card Test</title></head><body></body></html>`))
	}))
	defer ts.Close()

	bin := buildCLI(t)
	cmd := exec.Command(bin, "card", "stdout", ts.URL)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("card stdout failed: %v\nstderr: %s", err, stderr.String())
	}

	stdoutBytes := stdout.Bytes()
	if len(stdoutBytes) < 8 {
		t.Fatalf("stdout too short: %d", len(stdoutBytes))
	}

	pngSig := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if !bytes.HasPrefix(stdoutBytes, pngSig) {
		t.Fatalf("stdout does not begin with PNG magic signature")
	}

	if _, err := png.Decode(bytes.NewReader(stdoutBytes)); err != nil {
		t.Fatalf("stdout failed to decode as PNG: %v", err)
	}

	// Stderr must receive human diagnostics
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "card-cover:") {
		t.Errorf("expected card-cover on stderr, got: %q", stderrStr)
	}
}

func TestCLI_CardFile_StrictFailurePreservesExistingFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>No Cover Page</title></head><body></body></html>`))
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "preserved_card.png")
	origBytes := []byte("DO_NOT_OVERWRITE_THIS_FILE")
	if err := os.WriteFile(outPath, origBytes, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "card", "file", "--require-cover", "--out", outPath, ts.URL)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected nonzero exit code when require-cover fails, got success: %s", out)
	}

	// Verify original file is untouched
	readBack, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading file: %v", err)
	}
	if !bytes.Equal(readBack, origBytes) {
		t.Fatalf("file was overwritten or modified on strict failure!")
	}
}

func TestCLI_CardFile_NonStrictPlaceholder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Placeholder Page</title></head><body></body></html>`))
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "placeholder_card.png")

	cmd := exec.Command(bin, "card", "file", "--out", outPath, ts.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("card file failed in default non-strict mode: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading card: %v", err)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("card output is not valid PNG: %v", err)
	}
}

func TestCLI_CardFile_QualityReportJSON(t *testing.T) {
	pngBytes := createValidPNG(16, 16)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			html := `<!DOCTYPE html><html><head><title>Report Test</title><meta property="og:image" content="/img.png"></head><body></body></html>`
			_, _ = w.Write([]byte(html))
		case "/img.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "output_card.png")
	reportPath := filepath.Join(tempDir, "report.json")

	cmd := exec.Command(bin, "card", "file", "--quality-report", reportPath, "--out", outPath, ts.URL+"/html")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("card file failed: %v\n%s", err, out)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed reading report JSON: %v", err)
	}

	var report cards.CardQualityReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("failed parsing report JSON: %v", err)
	}

	if report.InputURL != ts.URL+"/html" {
		t.Errorf("InputURL = %q, want %q", report.InputURL, ts.URL+"/html")
	}
	if report.CoverStatus != cards.ImageStatusLoaded {
		t.Errorf("CoverStatus = %q, want %q", report.CoverStatus, cards.ImageStatusLoaded)
	}
	if report.UsedPlaceholder {
		t.Errorf("UsedPlaceholder = true, want false")
	}
	if len(report.Candidates) == 0 {
		t.Errorf("Candidates list is empty")
	}
}

func TestCLI_CardHelp_FlagsPresent(t *testing.T) {
	bin := buildCLI(t)
	for _, sub := range []string{"file", "stdout", "clipboard"} {
		t.Run(sub, func(t *testing.T) {
			cmd := exec.Command(bin, "card", sub, "--help")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed running help: %v\n%s", err, out)
			}
			outStr := string(out)
			for _, expectedFlag := range []string{
				"--retry-attempts",
				"--retry-delay",
				"--retry-max-delay",
				"--timeout",
				"--image-timeout",
				"--require-cover",
				"--quality-report",
			} {
				if !strings.Contains(outStr, expectedFlag) {
					t.Errorf("subcommand %s help missing flag %s", sub, expectedFlag)
				}
			}
		})
	}
}

func TestCLI_CardFile_UnwritableQualityReportPathFails(t *testing.T) {
	pngBytes := createValidPNG(16, 16)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Unwritable Report</title><meta property="og:image" content="/img.png"></head><body></body></html>`))
		case "/img.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "preserved_card.png")
	origBytes := []byte("DO_NOT_TOUCH_CARD")
	if err := os.WriteFile(outPath, origBytes, 0644); err != nil {
		t.Fatal(err)
	}

	unwritableReport := filepath.Join(tempDir, "nonexistent-dir-12345", "sub", "report.json")

	cmd := exec.Command(bin, "card", "file", "--quality-report", unwritableReport, "--out", outPath, ts.URL+"/html")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error when quality-report path is unwritable, got success: %s", out)
	}
	if !strings.Contains(string(out), "quality report") {
		t.Errorf("output %q should mention quality report failure", string(out))
	}

	// Verify original file is untouched
	readBack, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading file: %v", err)
	}
	if !bytes.Equal(readBack, origBytes) {
		t.Fatalf("output file was modified despite quality report write failure!")
	}
}

func TestCLI_CardFile_SensitiveURLRedactionInReport(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/doc":
			w.Header().Set("Content-Type", "text/html")
			// Candidate URL contains credentials and secret tokens
			serverHost := r.Host
			html := `<!DOCTYPE html><html><head><title>Secure Page</title>
<meta property="og:image" content="http://imguser:imgpass@` + serverHost + `/img?secret_token=topsecret123">
</head><body></body></html>`
			_, _ = w.Write([]byte(html))
		case "/img":
			// Return 500 error to test that error message and candidate outcome don't leak secrets
			http.Error(w, "internal image failure", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "card.png")
	reportPath := filepath.Join(tempDir, "redacted_report.json")

	// Input URL contains credentials and secret parameters
	inputURL := strings.Replace(ts.URL, "http://", "http://loginuser:loginpass@", 1) + "/doc?access_token=supersecret456"

	cmd := exec.Command(bin, "card", "file",
		"--retry-attempts", "1",
		"--quality-report", reportPath,
		"--out", outPath,
		inputURL,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("card file failed: %v\n%s", err, out)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed reading report: %v", err)
	}
	reportStr := string(reportBytes)

	// Verify no credentials or secret tokens leaked anywhere in the JSON report
	leaks := []string{"loginpass", "supersecret456", "imgpass", "topsecret123"}
	for _, leak := range leaks {
		if strings.Contains(reportStr, leak) {
			t.Errorf("quality report leaked secret token or password %q:\n%s", leak, reportStr)
		}
	}

	var report cards.CardQualityReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("failed unmarshaling report: %v", err)
	}

	if !strings.Contains(report.InputURL, "REDACTED") {
		t.Errorf("InputURL was not redacted: %q", report.InputURL)
	}
	if len(report.Candidates) > 0 {
		cand := report.Candidates[0]
		if !strings.Contains(cand.URL, "REDACTED") {
			t.Errorf("Candidate URL was not redacted: %q", cand.URL)
		}
	}
}

func TestCLI_CardFile_RedirectUpdatesFinalURLAndResolvesRelativeCandidates(t *testing.T) {
	pngBytes := createValidPNG(16, 16)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/entry":
			http.Redirect(w, r, "/nested/section/page.html", http.StatusMovedPermanently)
		case "/nested/section/page.html":
			w.Header().Set("Content-Type", "text/html")
			html := `<!DOCTYPE html><html><head><title>Redirect Doc</title>
<meta property="og:image" content="../images/cover.png">
</head><body></body></html>`
			_, _ = w.Write([]byte(html))
		case "/nested/images/cover.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "card.png")
	reportPath := filepath.Join(tempDir, "redirect_report.json")

	cmd := exec.Command(bin, "card", "file",
		"--quality-report", reportPath,
		"--out", outPath,
		ts.URL+"/entry",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("card file failed: %v\n%s", err, out)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed reading report: %v", err)
	}
	var report cards.CardQualityReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("failed unmarshaling report: %v", err)
	}

	expectedFinalURL := ts.URL + "/nested/section/page.html"
	expectedSelectedCover := ts.URL + "/nested/images/cover.png"

	if report.InputURL != ts.URL+"/entry" {
		t.Errorf("report.InputURL = %q, want %q", report.InputURL, ts.URL+"/entry")
	}
	if report.FinalURL != expectedFinalURL {
		t.Errorf("report.FinalURL = %q, want %q", report.FinalURL, expectedFinalURL)
	}
	if report.SelectedCoverURL != expectedSelectedCover {
		t.Errorf("report.SelectedCoverURL = %q, want %q", report.SelectedCoverURL, expectedSelectedCover)
	}
	if report.CoverStatus != cards.ImageStatusLoaded {
		t.Errorf("report.CoverStatus = %q, want %q", report.CoverStatus, cards.ImageStatusLoaded)
	}
}

func TestCLI_View_ImageFlagsApplied(t *testing.T) {
	var imageRequests int32
	pngBytes := createValidPNG(16, 16)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta property="og:image" content="/img.png"></head></html>`))
		case "/img.png":
			reqNum := atomic.AddInt32(&imageRequests, 1)
			if reqNum == 1 {
				// Transient 500 on first attempt
				http.Error(w, "transient error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	// Running with retry-attempts 3 should retry the image download and succeed!
	cmd := exec.Command(bin, "view",
		"--viewer", "mock",
		"--retry-attempts", "3",
		"--retry-delay", "10ms",
		"--retry-max-delay", "50ms",
		"--image-timeout", "2s",
		"image",
		ts.URL+"/page.html",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("view failed: %v\n%s", err, out)
	}

	if atomic.LoadInt32(&imageRequests) != 2 {
		t.Errorf("expected 2 image requests due to retry, got %d", atomic.LoadInt32(&imageRequests))
	}
}

func TestCLI_Extract_ShowImageFlagsApplied(t *testing.T) {
	var imageRequests int32
	pngBytes := createValidPNG(16, 16)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/extract_page.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta property="og:image" content="/extract_img.png"></head></html>`))
		case "/extract_img.png":
			reqNum := atomic.AddInt32(&imageRequests, 1)
			if reqNum == 1 {
				// Transient 500 on first attempt
				http.Error(w, "transient error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	cmd := exec.Command(bin, "extract",
		"--show-image",
		"--retry-attempts", "3",
		"--retry-delay", "10ms",
		"--retry-max-delay", "50ms",
		"--image-timeout", "2s",
		ts.URL+"/extract_page.html",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}

	if atomic.LoadInt32(&imageRequests) != 2 {
		t.Errorf("expected 2 image requests due to retry, got %d", atomic.LoadInt32(&imageRequests))
	}
}

func TestCLI_HTTPStatusErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/not-found":
			http.NotFound(w, r)
		case "/forbidden":
			http.Error(w, "Access Denied", http.StatusForbidden)
		case "/server-error":
			http.Error(w, "Unavailable", http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	for _, tc := range []struct {
		name       string
		subcmd     string
		path       string
		expectCode int
	}{
		{"extract-404", "extract", "/not-found", 404},
		{"extract-403", "extract", "/forbidden", 403},
		{"extract-503", "extract", "/server-error", 503},
		{"lint-404", "lint", "/not-found", 404},
		{"verify-404", "verify", "/not-found", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(bin, tc.subcmd, "--retry-attempts", "1", ts.URL+tc.path)
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected non-zero exit code for %s %d, got success: %s", tc.subcmd, tc.expectCode, out)
			}
			outStr := string(out)
			if !strings.Contains(outStr, fmt.Sprintf("%d", tc.expectCode)) {
				t.Errorf("output %q should contain HTTP status code %d", outStr, tc.expectCode)
			}
		})
	}
}

func TestCLI_View_ImageFailureNonZeroExit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page_with_bad_image.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta property="og:image" content="/missing.jpg"></head></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	cmd := exec.Command(bin, "view",
		"--viewer", "mock",
		"--retry-attempts", "1",
		"image",
		ts.URL+"/page_with_bad_image.html",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected view to fail with nonzero exit code when image 404s, got success")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Failed to download image") && !strings.Contains(stderrStr, "404") {
		t.Errorf("stderr %q should contain failure diagnostic", stderrStr)
	}
}

func TestCLI_Extract_ShowImageFailureNonZeroExit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page_with_bad_image.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Bad Img</title><meta property="og:image" content="/missing.jpg"></head></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	cmd := exec.Command(bin, "extract",
		"--show-image",
		"--retry-attempts", "1",
		ts.URL+"/page_with_bad_image.html",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected extract --show-image to fail with nonzero exit code when image 404s, got success")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Failed to download image") && !strings.Contains(stderrStr, "404") {
		t.Errorf("stderr %q should contain failure diagnostic", stderrStr)
	}
}

func TestCLI_View_RawURLSanitizationInStatus(t *testing.T) {
	pngData := createValidPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/doc":
			w.Header().Set("Content-Type", "text/html")
			// Image URL contains sensitive userinfo and signed token
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
<meta property="og:image" content="http://imguser:secretpass@` + r.Host + `/secret.png?token=topsecrettoken&sig=signedval">
</head></html>`))
		case "/secret.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bin := buildCLI(t)

	// Case 1: Automatic fallback status message must redact credentials and tokens
	cmd := exec.Command(bin, "view",
		"--viewer", "default",
		"--retry-attempts", "1",
		"image",
		ts.URL+"/doc",
	)
	out, _ := cmd.CombinedOutput()
	outStr := string(out)

	if strings.Contains(outStr, "secretpass") {
		t.Errorf("automatic status message leaked secretpass: %s", outStr)
	}
	if strings.Contains(outStr, "topsecrettoken") {
		t.Errorf("automatic status message leaked topsecrettoken: %s", outStr)
	}
	if strings.Contains(outStr, "signedval") {
		t.Errorf("automatic status message leaked signedval: %s", outStr)
	}
	if !strings.Contains(outStr, "REDACTED") {
		t.Errorf("automatic status message did not redact URL: %s", outStr)
	}

	// Case 2: Explicit print-url preserves exact raw URL
	cmdPrint := exec.Command(bin, "view",
		"--viewer", "print-url",
		"image",
		ts.URL+"/doc",
	)
	outPrint, errPrint := cmdPrint.CombinedOutput()
	if errPrint != nil {
		t.Fatalf("print-url failed: %v\n%s", errPrint, outPrint)
	}
	outPrintStr := string(outPrint)
	if !strings.Contains(outPrintStr, "secretpass") || !strings.Contains(outPrintStr, "topsecrettoken") {
		t.Errorf("print-url should preserve exact URL for caller: %s", outPrintStr)
	}
}

func TestCLI_View_UnsupportedViewerReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta property="og:image" content="/img.png"></head></html>`))
	}))
	defer ts.Close()

	bin := buildCLI(t)
	cmd := exec.Command(bin, "view", "--viewer", "unsupported-viewer-xyz", "image", ts.URL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected view to fail with nonzero exit code for unsupported viewer, got success")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, `unsupported viewer "unsupported-viewer-xyz" requested`) {
		t.Errorf("stderr %q should mention unsupported viewer", stderrStr)
	}
}
