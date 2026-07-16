# header

`header` is a Go command-line application that extracts human-readable metadata from a given URL. It scrapes OpenGraph information, Twitter cards, JSON-LD, and YouTube specifics (watch time, video IDs, titles), displaying them nicely on the terminal.

It also supports checking for missing information (`--test` mode) and displaying cover images directly (`--show-image`).

## Installation

### GitHub Releases
Download binaries from: https://github.com/arran4/header/releases

### Homebrew
```bash
brew tap arran4/homebrew-tap
brew install header
```

### Go install
```bash
go install github.com/arran4/header/cmd/header@latest
```

## Usage

Provide a URL to extract information:
```bash
header https://youtube.com/watch?v=dQw4w9WgXcQ
```

### Display Cover Images

By default, the `--show-image` flag will download the image and use your system's default image viewer (`xdg-open`, `open`, or `start`) to display it.

```bash
header --show-image https://example.com/some/article
```

**Experimental Shiny UI:**
You can build the app with the `shiny` tag to display the image using an experimental UI window, instead of spawning an external executable:

```bash
go build -tags shiny ./cmd/header
./header --show-image https://example.com/some/article
```

### Test Diagnostics

If you want to debug what metadata a site is missing, you can use `--test`:

```bash
header --test https://example.com
```

This will print out a list of any missing common fields (like no OpenGraph Data, no Title, etc.).

## Development & Testing

Run unit and integration tests using:

```bash
go test ./...
```

### Redirect Tracing

`header` replaces the `trace-redirects` application and supports tracing HTTP standard redirects as well as HTML meta-refresh redirects.

```bash
header trace-redirects https://example.com/redirect
```

You can customize the behaviour using various flags like `-store-cookies`, `-redirect-limit`, `-update-referer`, and `-redirect-delay`.

Alternatively, use the `-trace-redirects` and `-follow-meta-redirects` flags on the default `extract` command to trace redirects before extracting the metadata.
