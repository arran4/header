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
