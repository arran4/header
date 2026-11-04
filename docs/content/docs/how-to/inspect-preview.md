---
title: "Inspect social preview metadata"
weight: 10
---

# Inspect social preview metadata

Fetch a page and focus on the metadata that social platforms use:

```sh
header extract --style opengraph https://example.com
header extract --style twitter https://example.com
```

The output only includes properties found on the page; an absent Open Graph image, for example, is not invented.

To run built-in metadata checks:

```sh
header lint https://example.com
header lint --list all https://example.com
```

To generate a PNG preview card from a webpage:

```sh
header card list
header card file --out card.png https://example.com
```

Card commands support bounded HTTP retries and automatic cover candidate fallback across Open Graph (`og:image`), Open Graph Secure URL (`og:image:secure_url`), and Twitter (`twitter:image`).

To enforce that a card must have a valid cover image and fail otherwise:

```sh
header card file --require-cover --out card.png https://example.com
```

To output machine-readable card quality diagnostics to JSON:

```sh
header card file --quality-report report.json --out card.png https://example.com
```

When piping `header card stdout`, only valid binary PNG data is emitted on standard output. Human diagnostics (`card-quality:` and `card-cover:`) are written to standard error.

### Cover image statuses

The card quality report records one of the following image statuses:
- `loaded`: The candidate was successfully downloaded and decoded.
- `no-image-metadata`: No image candidate metadata was found in the page.
- `http-status`: Candidate download failed with an HTTP error status (e.g. 404).
- `timeout`: Candidate download timed out.
- `network-error`: Network connection failure.
- `decode-error`: Image format was corrupted or unreadable.
- `image-too-large`: Image file size or dimensions exceeded safety bounds.
- `all-candidates-failed`: All available candidates were attempted and failed.

In default (non-strict) mode, cards fall back to a visual placeholder when cover images cannot be loaded. Strict mode (`--require-cover`) exits with a nonzero status and does not overwrite existing good files.

The card file is written locally. Be aware that downloaded website images remain the property of their respective owners.

## GitHub repository card quality

For a `https://github.com/OWNER/REPO` page, the card generators use the repository identity from the URL as the title and strip GitHub's known account-creation suffix from Open Graph/Twitter descriptions. They do not change the original fetched `PageInfo`; `header extract --style json` continues to report the page's actual metadata.

The description is still derived from the original page and may not match its author's preferred public description. When no meaningful description remains, `header card file`, `card clipboard`, and `card stdout` emit `card-quality: needs-owner-description` to standard error. A valid PNG is a *candidate*, not publication approval. Review source and selected copy before sharing.

This deliberately avoids making hundreds of separate GitHub REST API calls during batch generation. If a separate catalogue already holds GitHub repository descriptions, preserve that data and its provenance for the owner's review.
