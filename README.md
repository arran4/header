# header

`header` is a Go command-line application that extracts human-readable metadata from a given URL. It scrapes OpenGraph information, Twitter cards, JSON-LD, and YouTube specifics (watch time, video IDs, titles), displaying them nicely on the terminal.

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

## Commands

### `show` / `extract` (default)
Extracts human-readable metadata from a given URL.

```bash
header https://youtube.com/watch?v=dQw4w9WgXcQ
header extract https://youtube.com/watch?v=dQw4w9WgXcQ
```

**Display Cover Images**
By default, the `-show-image` flag will download the image and use your system's default image viewer (`xdg-open`, `open`, or `start`) to display it.
```bash
header extract -show-image https://example.com/some/article
```
*Experimental Shiny UI:* You can build the app with the `shiny` tag to display the image using an experimental UI window:
```bash
go build -tags shiny ./cmd/header
./header extract -show-image https://example.com/some/article
```

**Test Diagnostics**
If you want to debug what metadata a site is missing, you can use `-test`:
```bash
header extract -test https://example.com
```

### `edit`
Edits HTML file head/meta fields.

```bash
header edit head -S '<link rel="canonical" href="https://example.com/"/>' index.html
header edit meta -s "new value" "description" index.html
```
Options:
- `-s`: set (remove other versions if any)
- `-S`: add to set (add a new one, only works where valid)
- `-r`: replace (only works if replacing 1 value)
- `-n`: replace number x, only works with -r

### `trace-redirects`
Traces URL HTTP standard redirects as well as HTML meta-refresh redirects.

```bash
header trace-redirects https://example.com/redirect
```
Customization flags include `-store-cookies`, `-redirect-limit`, `-update-referer`, and `-redirect-delay`.

### `verify`
Checks all URLs and other details; a subset of lint rules mostly concerned with valid info (URL reachability and image validity).

```bash
header verify https://example.com
```

### `lint`
Lints the metadata of a URL.

```bash
header lint https://example.com
header lint -list all https://example.com
header lint -rule twitter,opengraph https://example.com
```

### `lint-rule-lists`
Lists available lint rule lists.

```bash
header lint-rule-lists
```

### `lint-rules`
Lists available lint rules.

```bash
header lint-rules
```

### `view`
Views images or icons from a URL.

```bash
header view https://example.com/image.png
```

### `httpretty`
Fetches a URL and prints HTTP requests and responses using httpretty.

```bash
header httpretty https://example.com
```

## Development & Testing

Run unit and integration tests using:

```bash
go test ./...
```
