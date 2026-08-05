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

Example Output:
```
URL: https://youtube.com/watch?v=dQw4w9WgXcQ
Title: Rick Astley - Never Gonna Give You Up (Official Video) (4K Remaster) - YouTube
Description: The official video for “Never Gonna Give You Up” by Rick Astley. Never: The Autobiography 📚 OUT NOW! Follow this link to get your copy and listen to Rick’s ...
Cover Images:
  - https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg
```

**Output Formats**
The tool supports multiple output formatting templates using the `-style` flag:
- `default`: The default format printing concise, highlighted essential fields (URL, title, description, cover images).
- `plain`: The same as `default` but without color highlights.
- `all_fields`: A detailed format printing all captured metadata grouped logically.
- `all_fields_plain`: A detailed plain text format.
- `json`: Prints the raw page metadata as JSON.
- `discord`: A format tailored for Discord metadata embedding.
- `jsonld`: Prints human readable JSON-LD metadata.
- `opengraph`: Prints OpenGraph metadata only.
- `twitter`: Prints Twitter card metadata only.

Example of retrieving all fields in JSON:
```bash
header extract -style json https://example.com
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

The `edit` command operates in two modes: `head` and `meta`.
- `header edit head [flags] <file>` targets the root `<head>` block, allowing you to manipulate standard HTML links and elements.
- `header edit meta [flags] <name> <file>` specifically targets a `<meta name="...">` or `<meta property="...">` tag for modification.

Options:
- `-s`: set (removes all existing occurrences of this tag and inserts the new one).
- `-S`: add to set (appends a new one, does not remove existing tags).
- `-r`: replace (replaces exactly 1 existing value).
- `-n`: replace number x (only works with `-r`, specifies which index to replace).

Use `-` as the filename to read from `stdin` and write to `stdout`, which is useful for chaining in pipes.

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
