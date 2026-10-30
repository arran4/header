---
title: "Trace redirects"
weight: 20
---

# Trace redirects

To follow HTTP redirects and HTML meta-refresh redirects:

```sh
header trace-redirects https://example.com
```

Use `header trace-redirects --help` for the current redirect limit, delay and cookie options. Avoid using the fetcher as an unrestricted proxy in a privileged network environment; see the repository's [security policy](https://github.com/arran4/header/blob/main/SECURITY.md).
