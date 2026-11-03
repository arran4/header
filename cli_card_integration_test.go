package header_test

import (
	"bytes"
	"encoding/json"
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
