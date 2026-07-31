package header

import (
	"embed"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/txtar"
)

//go:embed testdata/extractor/*.txtar
var testDataFS embed.FS

func TestExtractors(t *testing.T) {
	err := fs.WalkDir(testDataFS, "testdata/extractor", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".txtar") {
			return nil
		}

		t.Run(d.Name(), func(t *testing.T) {
			fileData, err := testDataFS.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", path, err)
			}

			archive := txtar.Parse(fileData)
			var htmlContent, expectedJSON string

			for _, file := range archive.Files {
				if file.Name == "example.html" {
					htmlContent = string(file.Data)
				} else if file.Name == "expected.json" {
					expectedJSON = string(file.Data)
				}
			}

			if htmlContent == "" {
				t.Fatalf("example.html not found in archive")
			}
			if expectedJSON == "" {
				t.Fatalf("expected.json not found in archive")
			}

			info := &PageInfo{
				URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", // Ensure youtube extractor works if needed
			}

			for _, extractor := range extractorRegistry {
				err = extractor.Extract(strings.NewReader(htmlContent), info)
				if err != nil {
					t.Fatalf("Extractor %s returned error: %v", extractor.Name(), err)
				}
			}

			// Verify fields using JSON comparison for exact match
			var expectedInfo PageInfo
			if err := json.Unmarshal([]byte(expectedJSON), &expectedInfo); err != nil {
				t.Fatalf("Failed to parse expected.json: %v", err)
			}

			actualJSON, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				t.Fatalf("Failed to serialize info to json: %v", err)
			}

			if expectedInfo.Title != info.Title || expectedInfo.Keywords != info.Keywords ||
				expectedInfo.OGTitle != info.OGTitle || expectedInfo.TwitterTitle != info.TwitterTitle {
				t.Errorf("Mismatch in expected output.\nExpected:\n%s\n\nActual:\n%s\n", expectedJSON, string(actualJSON))
			}

		})
		return nil
	})

	if err != nil {
		t.Fatalf("Error walking test directory: %v", err)
	}
}
