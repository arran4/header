---
title: "Getting started"
weight: 10
---

# Getting started

**Goal:** Install `header`, extract metadata from a URL, and inspect the result.

## Prerequisites

Install a supported Go toolchain (Go 1.26 or later to build from source) and have network access to a webpage. Alternatively, download a compatible binary from [GitHub Releases](https://github.com/arran4/header/releases) when available.

## Install

```sh
go install github.com/arran4/header/cmd/header@latest
```

Check the installed command:

```sh
header --help
```

## Extract metadata

```sh
header extract --style json https://example.com
```

This requests the page and prints the extracted metadata in JSON. The actual fields depend on the remote page and can change as its content changes.

To see the available formatting styles and flags:

```sh
header extract --help
```

## Troubleshooting

If a request fails, first verify network access and the URL, then compare the [available commands](reference/) against your installed version. For deeper diagnosis see [Troubleshooting](how-to/troubleshooting/).

For structured extraction in Go applications, see the authoritative [package reference on pkg.go.dev](https://pkg.go.dev/github.com/arran4/header).
