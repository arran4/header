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

The card file is written locally. Be aware that downloaded website images remain the property of their respective owners.

## GitHub repository card quality

For a `https://github.com/OWNER/REPO` page, the card generators use the repository identity from the URL as the title and strip GitHub's known account-creation suffix from Open Graph/Twitter descriptions. They do not change the original fetched `PageInfo`; `header extract --style json` continues to report the page's actual metadata.

The description is still derived from the original page and may not match its author's preferred public description. When no meaningful description remains, `header card file`, `card clipboard`, and `card stdout` emit `card-quality: needs-owner-description` to standard error. A valid PNG is a *candidate*, not publication approval. Review source and selected copy before sharing.

This deliberately avoids making hundreds of separate GitHub REST API calls during batch generation. If a separate catalogue already holds GitHub repository descriptions, preserve that data and its provenance for the owner's review.
