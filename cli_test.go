package header_test

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var ts *httptest.Server

//go:embed testdata/server/example.html
var exampleHTML []byte

//go:embed testdata/server/missing.html
var missingHTML []byte

//go:embed testdata/server/issue.html
var issueHTML []byte

//go:embed testdata/server/media.html
var mediaHTML []byte

func TestMain(m *testing.M) {
	// Start server that runs for the entire duration of the tests, including child processes!
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/example":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(exampleHTML)
		case "/missing":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(missingHTML)
		case "/issue":
			w.Header().Set("Content-Type", "text/html")
			// We need to inject the actual URL into the HTML so they match
			html := string(issueHTML)
			// Find "http://127.0.0.1:0/issue" and replace with actual server URL
			html = strings.Replace(html, "http://127.0.0.1:0/issue", ts.URL+"/issue", 1)
			_, _ = w.Write([]byte(html))
		case "/media":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(mediaHTML)
		case "/image.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = fmt.Fprint(w, "fake image data")
		case "/redirect1":
			w.Header().Set("Location", "/redirect2")
			w.WriteHeader(http.StatusFound)
		case "/redirect2":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><meta http-equiv="refresh" content="0; url=/redirect3"></head><body>redirecting...</body></html>`))
		case "/redirect3":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body>Final page</body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))

	code := m.Run()
	ts.Close()
	os.Exit(code)
}
func TestCLI(t *testing.T) {
	runTxtarScript(t, "testdata/cli/txtar", map[string]string{
		"TEST_SERVER_URL": ts.URL,
	})
}
