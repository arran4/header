package templates_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/header"
	"github.com/arran4/header/templates"
)

func TestListStyles(t *testing.T) {
	styles, err := templates.ListStyles()
	if err != nil {
		t.Fatalf("ListStyles failed: %v", err)
	}

	if len(styles) == 0 {
		t.Fatal("expected at least one style, got none")
	}

	// Verify styles are sorted alphabetically
	for i := 1; i < len(styles); i++ {
		if styles[i-1].Name >= styles[i].Name {
			t.Errorf("styles not sorted: %s before %s", styles[i-1].Name, styles[i].Name)
		}
	}

	// Verify standard styles are present
	expected := []string{"all_fields", "all_fields_plain", "default", "discord", "html", "json", "jsonld", "opengraph", "quick", "twitter"}
	styleMap := make(map[string]templates.StyleInfo)
	for _, s := range styles {
		styleMap[s.Name] = s
	}

	for _, name := range expected {
		s, ok := styleMap[name]
		if !ok {
			t.Errorf("expected style %q not found", name)
		}
		if s.Description == "" {
			t.Errorf("expected description for style %q, got empty string", name)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	t.Run("existing style without extension", func(t *testing.T) {
		content, err := templates.GetTemplate("default")
		if err != nil {
			t.Fatalf("GetTemplate('default') failed: %v", err)
		}
		if len(content) == 0 {
			t.Fatal("expected non-empty template content")
		}
	})

	t.Run("existing style with extension", func(t *testing.T) {
		content, err := templates.GetTemplate("default.tmpl")
		if err != nil {
			t.Fatalf("GetTemplate('default.tmpl') failed: %v", err)
		}
		if len(content) == 0 {
			t.Fatal("expected non-empty template content")
		}
	})

	t.Run("nonexistent style", func(t *testing.T) {
		_, err := templates.GetTemplate("nonexistent_style")
		if err == nil {
			t.Fatal("expected error for nonexistent style, got nil")
		}
		if !strings.Contains(err.Error(), "error loading template") || !strings.Contains(err.Error(), "available styles:") {
			t.Errorf("error message should list available styles, got: %v", err)
		}
	})
}

func TestLoadTemplate(t *testing.T) {
	info := &header.PageInfo{
		URL:   "https://example.com/",
		Title: "Example",
	}

	t.Run("default embedded style", func(t *testing.T) {
		tmpl, err := templates.LoadTemplate("", "")
		if err != nil {
			t.Fatalf("LoadTemplate('', '') failed: %v", err)
		}
		out, err := templates.Render(tmpl, info)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if !strings.Contains(out, "https://example.com/") {
			t.Errorf("expected output to contain URL, got: %s", out)
		}
	})

	t.Run("custom style file", func(t *testing.T) {
		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom.tmpl")
		customContent := "Custom: {{.URL}} - {{.Title}}"
		if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}

		tmpl, err := templates.LoadTemplate("", customPath)
		if err != nil {
			t.Fatalf("LoadTemplate with custom file failed: %v", err)
		}
		out, err := templates.Render(tmpl, info)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if strings.TrimSpace(out) != "Custom: https://example.com/ - Example" {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("both custom style file and explicit non-default style returns error", func(t *testing.T) {
		_, err := templates.LoadTemplate("json", "/path/to/custom.tmpl")
		if err == nil {
			t.Fatal("expected error when specifying both --style and --style-file, got nil")
		}
	})

	t.Run("custom style file with default style allowed", func(t *testing.T) {
		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom.tmpl")
		if err := os.WriteFile(customPath, []byte("URL: {{.URL}}"), 0644); err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}
		tmpl, err := templates.LoadTemplate("default", customPath)
		if err != nil {
			t.Fatalf("LoadTemplate('default', customPath) failed: %v", err)
		}
		out, err := templates.Render(tmpl, info)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if strings.TrimSpace(out) != "URL: https://example.com/" {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("missing custom style file returns actionable error", func(t *testing.T) {
		_, err := templates.LoadTemplate("", "/nonexistent/template.tmpl")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
		if !strings.Contains(err.Error(), "error reading template file") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("invalid template syntax returns actionable error", func(t *testing.T) {
		tmpDir := t.TempDir()
		badPath := filepath.Join(tmpDir, "bad.tmpl")
		if err := os.WriteFile(badPath, []byte("Bad {{.URL"), 0644); err != nil {
			t.Fatalf("failed to write bad template: %v", err)
		}
		_, err := templates.LoadTemplate("", badPath)
		if err == nil {
			t.Fatal("expected syntax error, got nil")
		}
		if !strings.Contains(err.Error(), "error parsing template") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("template execution error returns error without output", func(t *testing.T) {
		tmpDir := t.TempDir()
		execErrPath := filepath.Join(tmpDir, "exec_err.tmpl")
		// Call a method that doesn't exist or fail execution
		if err := os.WriteFile(execErrPath, []byte("Hello {{.NonExistentField.Something}}"), 0644); err != nil {
			t.Fatalf("failed to write exec error template: %v", err)
		}
		tmpl, err := templates.LoadTemplate("", execErrPath)
		if err == nil {
			// If parsing succeeded, execution should fail
			out, execErr := templates.Render(tmpl, info)
			if execErr == nil {
				t.Fatal("expected execution error, got nil")
			}
			if out != "" {
				t.Errorf("expected empty output on execution error, got: %q", out)
			}
		}
	})
}
