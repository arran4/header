package templates_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/arran4/header"
	"github.com/arran4/header/templates"
	"github.com/fatih/color"
	"github.com/rogpeppe/go-internal/txtar"
)

func TestTemplates(t *testing.T) {
	color.NoColor = true // Disable color for deterministic output matching

	funcMap := template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
		"json": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "  ")
			return string(b)
		},
	}

	testFiles, err := filepath.Glob("testdata/*.txtar")
	if err != nil {
		t.Fatalf("Failed to glob testdata: %v", err)
	}
	if len(testFiles) == 0 {
		t.Skip("No txtar files found in testdata")
	}

	for _, file := range testFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			archive, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatalf("Failed to parse txtar file %s: %v", file, err)
			}

			var inputJSON []byte
			var description string

			for _, f := range archive.Files {
				switch f.Name {
				case "description":
					description = strings.TrimSpace(string(f.Data))
				case "input.json":
					inputJSON = f.Data
				}
			}

			if inputJSON == nil {
				t.Fatal("Missing input.json in txtar archive")
			}

			var info header.PageInfo
			if err := json.Unmarshal(inputJSON, &info); err != nil {
				t.Fatalf("Failed to unmarshal input.json: %v", err)
			}

			tmplFiles := []string{"all_fields.tmpl", "all_fields_plain.tmpl", "default.tmpl", "plain.tmpl", "json.tmpl", "markdown.tmpl"}
			for _, tmplName := range tmplFiles {
				t.Run(tmplName, func(t *testing.T) {
					expectedOutputName := "expected_" + strings.TrimSuffix(tmplName, ".tmpl") + ".txt"
					if tmplName == "all_fields.tmpl" || tmplName == "all_fields_plain.tmpl" {
						expectedOutputName = "expected.txt"
					}
					var expectedOutput string
					for _, f := range archive.Files {
						if f.Name == expectedOutputName {
							expectedOutput = string(f.Data)
							break
						}
					}

					tmplContent, err := templates.FS.ReadFile(tmplName)
					if err != nil {
						t.Fatalf("Failed to load template %s: %v", tmplName, err)
					}

					tmpl, err := template.New("output").Funcs(funcMap).Parse(string(tmplContent))
					if err != nil {
						t.Fatalf("Failed to parse template: %v", err)
					}

					var buf bytes.Buffer
					if err := tmpl.Execute(&buf, &info); err != nil {
						t.Fatalf("Failed to execute template: %v", err)
					}

					actual := buf.String()
					if actual != expectedOutput {
						t.Errorf("Template output mismatch.\nDescription: %s\nExpected:\n%q\nGot:\n%q\n\nExpected exact:\n%s\nGot exact:\n%s",
							description, expectedOutput, actual, expectedOutput, actual)
					}
				})
			}
		})
	}
}
