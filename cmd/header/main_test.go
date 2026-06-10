package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	_ "embed"

	"github.com/rogpeppe/go-internal/testscript"
)

var ts *httptest.Server

//go:embed testdata/example.html
var exampleHTML []byte

//go:embed testdata/missing.html
var missingHTML []byte

func TestMain(m *testing.M) {
	// Start server that runs for the entire duration of the tests, including child processes!
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/example" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(exampleHTML)
		} else if r.URL.Path == "/missing" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(missingHTML)
		} else if r.URL.Path == "/image.jpg" {
			w.Header().Set("Content-Type", "image/jpeg")
			fmt.Fprint(w, "fake image data")
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	os.Exit(testscript.RunMain(m, map[string]func() int{
		"header": func() int {
			// Restore original os.Args after the call
			origArgs := os.Args
			defer func() { os.Args = origArgs }()

			// We need a fresh set of flags per run since `flag.Parse()` is called
			newArgs := make([]string, len(origArgs))
			copy(newArgs, origArgs)

			os.Args = newArgs
			main()
			return 0
		},
	}))
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "../../testdata",
		Setup: func(env *testscript.Env) error {
			// Replace localhost placeholders with actual server URL
			env.Setenv("TEST_SERVER_URL", ts.URL)
			return nil
		},
	})
}
