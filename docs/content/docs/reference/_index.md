---
title: "CLI reference"
weight: 40
---

# CLI reference

Use the installed executable's help for exact flags, positional arguments and defaults rather than a hand-maintained copy.

```sh
header --help
header extract --help
header card --help
```

## Commands

| Command | Purpose |
| --- | --- |
| `extract` (or default URL argument) | Fetch and render website metadata |
| `lint` / `verify` | Inspect metadata and URL validity |
| `lint-rules` / `lint-rule-lists` | Discover available lint rules and lists |
| `trace-redirects` | Inspect HTTP and meta-refresh redirects |
| `edit head` / `edit meta` | Modify local HTML head elements |
| `card` | Export a shareable preview image |
| `view` | View linked images or icons |
| `httpretty` | Print HTTP request and response details |

Output styles include `default`, `all_fields`, `all_fields_plain`, `json`, `discord`, `jsonld`, `opengraph` and `twitter`. Use `header extract --help` to check the current version.

## Common HTTP Options

HTTP-backed commands support bounded retries for idempotent requests:

- `--retry-attempts`: Total attempts for idempotent GET requests (default: `3`; `1` disables retries).
- `--retry-delay`: Initial delay between retries (default: `250ms`).
- `--retry-max-delay`: Maximum backoff delay cap (default: `5s`).
- `--timeout`: Maximum total request timeout (default: `20s`).
- `--image-timeout`: Image download timeout for image and card commands (default: `10s`).

Retry policy automatically retries transient network errors and HTTP 408, 429, 500, 502, 503, and 504 responses, respecting `Retry-After` headers and context deadlines. Permanent 400, 401, 403, and 404 responses are not retried.

Go package APIs are documented at [pkg.go.dev/github.com/arran4/header](https://pkg.go.dev/github.com/arran4/header).
