package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cli "github.com/arran4/header/internal/cli"
)

func TestPrintStyles(t *testing.T) {
	var buf bytes.Buffer
	err := cli.PrintStyles(&buf)
	if err != nil {
		t.Fatalf("PrintStyles failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Available Styles:") {
		t.Errorf("expected header 'Available Styles:', got: %s", out)
	}

	// Verify all known styles are present
	for _, style := range []string{"all_fields", "all_fields_plain", "default", "discord", "json", "jsonld", "opengraph", "twitter"} {
		if !strings.Contains(out, style) {
			t.Errorf("expected style %q in output, got: %s", style, out)
		}
	}
}

func TestStyleExport(t *testing.T) {
	t.Run("export nonexistent style returns error", func(t *testing.T) {
		err := cli.StyleExport("", false, "nonexistent_style")
		if err == nil {
			t.Fatal("expected error exporting nonexistent style, got nil")
		}
		if !strings.Contains(err.Error(), "available styles:") {
			t.Errorf("expected available styles listed in error, got: %v", err)
		}
	})

	t.Run("export to file and refuse overwrite without force", func(t *testing.T) {
		tmpDir := t.TempDir()
		outPath := filepath.Join(tmpDir, "exported.tmpl")

		// First export succeeds
		if err := cli.StyleExport(outPath, false, "all_fields"); err != nil {
			t.Fatalf("first export failed: %v", err)
		}

		content1, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		if len(content1) == 0 {
			t.Fatal("exported file is empty")
		}

		// Second export without force fails
		err = cli.StyleExport(outPath, false, "all_fields")
		if err == nil {
			t.Fatal("expected error on overwrite without force, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("unexpected error message: %v", err)
		}

		// Third export with force succeeds
		if err := cli.StyleExport(outPath, true, "default"); err != nil {
			t.Fatalf("export with force failed: %v", err)
		}

		content2, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read overwritten file: %v", err)
		}
		if bytes.Equal(content1, content2) {
			t.Error("expected overwritten content to differ")
		}
	})
}
