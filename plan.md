1.  **Add `Httpretty` to `FetchOptions` in `fetch.go`**:
    *   Include a boolean `Httpretty` in the `header.FetchOptions` struct.
    *   In `FetchURL`, if `opts.Httpretty` is true, configure the `http.Client`'s `Transport` with `httpretty.Logger`.

2.  **Add `-httpretty` flag to CLI commands**:
    *   Update `internal/cli/extract.go`, `lint.go`, `verify.go`, `view.go`, `trace_redirects.go` to include an `-httpretty` flag.
    *   Pass the flag value to `FetchOptions.Httpretty`.

3.  **Create a new subcommand `httpretty`**:
    *   In `cmd/header/main.go`, add a new case `httpretty` which calls `cli.HttprettyCmd`.
    *   Create `internal/cli/httpretty.go` implementing `HttprettyCmd`. This command will essentially function identically to `ExtractCmd` (or directly wrap it) but will have `Httpretty` enabled by default. Alternatively, it can just be a dedicated command that fetches a URL with `httpretty` enabled.
    *   Since the prompt mentions "a subcommand that prints out httpretty", a dedicated command `httpretty` that fetches a URL and prints its HTTP exchanges using `httpretty` (and perhaps discards the body or prints it too) makes sense.

4.  **Run pre-commit instructions**:
    *   Check for tests passing.

5.  **Submit changes**.
