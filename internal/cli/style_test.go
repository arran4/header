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
	for _, style := range []string{"all_fields", "all_fields_plain", "default", "discord", "html", "json", "jsonld", "opengraph", "quick", "twitter"} {
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

	t.Run("atomic creation on nonexistent path succeeds with 0644 mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		outPath := filepath.Join(tmpDir, "exported.tmpl")

		if err := cli.StyleExport(outPath, false, "all_fields"); err != nil {
			t.Fatalf("StyleExport failed: %v", err)
		}

		fi, err := os.Lstat(outPath)
		if err != nil {
			t.Fatalf("failed to stat exported file: %v", err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("exported file should not be a symlink")
		}
		if perm := fi.Mode().Perm(); perm != 0644 {
			t.Errorf("expected 0644 permissions, got %#o", perm)
		}

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		if len(content) == 0 {
			t.Fatal("exported file is empty")
		}
	})

	t.Run("collision on existing regular file without force fails and preserves content", func(t *testing.T) {
		tmpDir := t.TempDir()
		outPath := filepath.Join(tmpDir, "existing.tmpl")
		originalData := []byte("original content that must not be overwritten")
		if err := os.WriteFile(outPath, originalData, 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}

		err := cli.StyleExport(outPath, false, "all_fields")
		if err == nil {
			t.Fatal("expected collision error without force, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}

		currentData, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !bytes.Equal(currentData, originalData) {
			t.Errorf("existing file content was corrupted: got %q, want %q", string(currentData), string(originalData))
		}
	})

	t.Run("collision on existing symlink without force fails and preserves target", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, "target.tmpl")
		symlinkPath := filepath.Join(tmpDir, "link.tmpl")

		targetData := []byte("target file content that must not be overwritten")
		if err := os.WriteFile(targetPath, targetData, 0644); err != nil {
			t.Fatalf("failed to create target file: %v", err)
		}
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		err := cli.StyleExport(symlinkPath, false, "all_fields")
		if err == nil {
			t.Fatal("expected collision error on symlink without force, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}

		// Verify target was not modified
		currentTargetData, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("failed to read target: %v", err)
		}
		if !bytes.Equal(currentTargetData, targetData) {
			t.Errorf("target file content was modified: got %q, want %q", string(currentTargetData), string(targetData))
		}

		// Verify link remains a symlink
		fi, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to stat symlink: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("symlink was replaced: mode is %v", fi.Mode())
		}
	})

	t.Run("collision on dangling symlink without force fails and does not create target", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonexistentTarget := filepath.Join(tmpDir, "nonexistent.tmpl")
		symlinkPath := filepath.Join(tmpDir, "dangling.tmpl")

		if err := os.Symlink(nonexistentTarget, symlinkPath); err != nil {
			t.Fatalf("failed to create dangling symlink: %v", err)
		}

		err := cli.StyleExport(symlinkPath, false, "all_fields")
		if err == nil {
			t.Fatal("expected collision error on dangling symlink without force, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}

		if _, err := os.Lstat(nonexistentTarget); !os.IsNotExist(err) {
			t.Errorf("nonexistent target should not have been created: %v", err)
		}

		fi, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to lstat dangling symlink: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("dangling symlink was unexpectedly modified: mode %v", fi.Mode())
		}
	})

	t.Run("export to existing directory fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		dirPath := filepath.Join(tmpDir, "destination_dir")
		if err := os.Mkdir(dirPath, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		errWithoutForce := cli.StyleExport(dirPath, false, "all_fields")
		if errWithoutForce == nil {
			t.Fatal("expected error exporting to directory without force, got nil")
		}
		if !strings.Contains(errWithoutForce.Error(), "directory") {
			t.Errorf("expected 'directory' error message, got: %v", errWithoutForce)
		}

		errWithForce := cli.StyleExport(dirPath, true, "all_fields")
		if errWithForce == nil {
			t.Fatal("expected error exporting to directory with force, got nil")
		}
		if !strings.Contains(errWithForce.Error(), "directory") {
			t.Errorf("expected 'directory' error message, got: %v", errWithForce)
		}
	})

	t.Run("permission failure in read-only directory fails", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("skipping permission failure test as root user")
		}

		tmpDir := t.TempDir()
		roDir := filepath.Join(tmpDir, "readonly")
		if err := os.Mkdir(roDir, 0555); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		defer func() {
			_ = os.Chmod(roDir, 0755)
		}()

		outPath := filepath.Join(roDir, "test.tmpl")
		err := cli.StyleExport(outPath, false, "all_fields")
		if err == nil {
			t.Fatal("expected permission error, got nil")
		}

		errForce := cli.StyleExport(outPath, true, "all_fields")
		if errForce == nil {
			t.Fatal("expected permission error with force, got nil")
		}
	})

	t.Run("force overwrite on regular file succeeds atomically", func(t *testing.T) {
		tmpDir := t.TempDir()
		outPath := filepath.Join(tmpDir, "exported.tmpl")

		if err := cli.StyleExport(outPath, false, "all_fields"); err != nil {
			t.Fatalf("initial export failed: %v", err)
		}
		content1, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read initial export: %v", err)
		}

		if err := cli.StyleExport(outPath, true, "default"); err != nil {
			t.Fatalf("force overwrite failed: %v", err)
		}
		content2, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read overwritten export: %v", err)
		}
		if bytes.Equal(content1, content2) {
			t.Errorf("expected file content to change after force overwrite")
		}
	})

	t.Run("force overwrite on symlink replaces symlink without mutating target", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, "target.tmpl")
		symlinkPath := filepath.Join(tmpDir, "symlink.tmpl")

		targetOriginal := []byte("original target content that must remain untouched")
		if err := os.WriteFile(targetPath, targetOriginal, 0644); err != nil {
			t.Fatalf("failed to create target file: %v", err)
		}
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		if err := cli.StyleExport(symlinkPath, true, "quick"); err != nil {
			t.Fatalf("force export on symlink failed: %v", err)
		}

		// Target file MUST NOT have been modified
		targetAfter, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("failed to read target: %v", err)
		}
		if !bytes.Equal(targetAfter, targetOriginal) {
			t.Errorf("target was modified: got %q, want %q", string(targetAfter), string(targetOriginal))
		}

		// symlinkPath MUST now be a regular file, NOT a symlink
		fi, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to lstat export path: %v", err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("export path is still a symlink; expected regular file replacement")
		}

		content, err := os.ReadFile(symlinkPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		if len(content) == 0 {
			t.Errorf("exported file is empty")
		}
	})

	t.Run("force overwrite on dangling symlink replaces symlink with regular file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonexistentTarget := filepath.Join(tmpDir, "nonexistent.tmpl")
		symlinkPath := filepath.Join(tmpDir, "dangling.tmpl")

		if err := os.Symlink(nonexistentTarget, symlinkPath); err != nil {
			t.Fatalf("failed to create dangling symlink: %v", err)
		}

		if err := cli.StyleExport(symlinkPath, true, "quick"); err != nil {
			t.Fatalf("force export on dangling symlink failed: %v", err)
		}

		// Target file must not have been created
		if _, err := os.Lstat(nonexistentTarget); !os.IsNotExist(err) {
			t.Errorf("target file should not exist: %v", err)
		}

		// symlinkPath must now be a regular file
		fi, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to lstat export path: %v", err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("expected regular file, got symlink: mode %v", fi.Mode())
		}
	})
}
