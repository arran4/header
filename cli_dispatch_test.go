package header_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// A local proxy proves that tokens after -- reach the URL action instead of
// invoking a command, without resolving or contacting an external host.
func TestCLIDashDashDisablesCommands(t *testing.T) {
	requests := make(chan string, 1)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Host
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()
	bin := buildCLI(t)
	for _, token := range []string{"extract", "help", "version"} {
		t.Run(token, func(t *testing.T) {
			cmd := exec.Command(bin, "--", token)
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				switch strings.ToUpper(key) {
				case "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY":
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "HTTPS_PROXY="+proxy.URL, "NO_PROXY=")
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "error fetching URL") {
				t.Fatalf("expected URL action failure, got %v: %s", err, out)
			}
			select {
			case host := <-requests:
				if host != token+":443" {
					t.Fatalf("requested %q, want %q", host, token+":443")
				}
			default:
				t.Fatal("URL action did not reach the local proxy")
			}
		})
	}
}
