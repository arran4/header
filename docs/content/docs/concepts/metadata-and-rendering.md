---
title: "Extraction and rendering"
weight: 10
---

# Extraction and rendering

The program first fetches a page and extracts its available metadata into a `PageInfo` value. Supported groups include Open Graph, Twitter Cards, JSON-LD, App Links, HTML link relationships and selected site-specific information.

Different output styles render selected properties rather than changing the underlying extraction. The `json` style exposes structured fields; the Open Graph and Twitter styles focus on their respective families. This is why a field may be present in one style but omitted by design in another.

Lint rules and rule lists inspect the extracted result and report issues. The `card` commands use the same metadata to generate shareable images, while `edit` operates on local HTML documents.

For source-level details, see [`PageInfo` on pkg.go.dev](https://pkg.go.dev/github.com/arran4/header#PageInfo), [the template files](https://github.com/arran4/header/tree/main/templates) and [lint rules](https://github.com/arran4/header/tree/main/lint/rules). These are authoritative for exact fields and supported behaviour.
