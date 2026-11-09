package templates

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

// StyleInfo contains metadata about an extraction style.
type StyleInfo struct {
	Name        string
	Description string
}

var knownDescriptions = map[string]string{
	"all_fields":       "Complete extracted metadata with terminal colors",
	"all_fields_plain": "Complete extracted metadata without terminal colors",
	"default":          "Concise human-readable metadata summary",
	"discord":          "Discord embed-oriented metadata preview",
	"html":             "Deterministic synthetic HTML <head> metadata",
	"json":             "Raw extracted metadata as formatted JSON",
	"jsonld":           "Extracted JSON-LD structured data",
	"opengraph":        "Open Graph protocol metadata properties",
	"twitter":          "Twitter card metadata properties",
}

// ListStyles enumerates all styles available in the embedded filesystem in sorted order.
func ListStyles() ([]StyleInfo, error) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	var styles []StyleInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		styleName := strings.TrimSuffix(name, ".tmpl")
		desc := knownDescriptions[styleName]
		styles = append(styles, StyleInfo{
			Name:        styleName,
			Description: desc,
		})
	}

	sort.Slice(styles, func(i, j int) bool {
		return styles[i].Name < styles[j].Name
	})

	return styles, nil
}

// GetTemplate retrieves the raw embedded template source for the given style name.
func GetTemplate(name string) ([]byte, error) {
	cleanName := strings.TrimSuffix(strings.TrimSpace(name), ".tmpl")
	if cleanName == "" {
		styles, _ := ListStyles()
		return nil, fmt.Errorf("error loading template: empty style name; available styles: %s", styleNamesList(styles))
	}

	content, err := FS.ReadFile(cleanName + ".tmpl")
	if err != nil {
		styles, _ := ListStyles()
		return nil, fmt.Errorf("error loading template %q: unknown style; available styles: %s", name, styleNamesList(styles))
	}
	return content, nil
}

func styleNamesList(styles []StyleInfo) string {
	names := make([]string, len(styles))
	for i, s := range styles {
		names[i] = s.Name
	}
	return strings.Join(names, ", ")
}

// FuncMap returns the standard template functions available to extraction templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
		"json": func(v interface{}) (string, error) {
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal json: %w", err)
			}
			return string(b), nil
		},
	}
}

// LoadTemplate loads, validates, and parses a template from either the embedded catalogue or a custom file.
func LoadTemplate(style, styleFile string) (*template.Template, error) {
	if styleFile != "" && style != "" && style != "default" {
		return nil, errors.New("cannot specify both --style and --style-file")
	}

	var content []byte
	var tmplName string

	if styleFile != "" {
		data, err := os.ReadFile(styleFile)
		if err != nil {
			return nil, fmt.Errorf("error reading template file %q: %w", styleFile, err)
		}
		content = data
		tmplName = filepath.Base(styleFile)
	} else {
		targetStyle := style
		if targetStyle == "" {
			targetStyle = "default"
		}
		data, err := GetTemplate(targetStyle)
		if err != nil {
			return nil, err
		}
		content = data
		tmplName = targetStyle
	}

	tmpl, err := template.New(tmplName).Funcs(FuncMap()).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}
	return tmpl, nil
}

// Render executes the template against data into a buffer.
// If execution fails, an error is returned and no partial output is emitted.
func Render(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return buf.String(), nil
}
