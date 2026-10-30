---
title: "Troubleshooting"
weight: 30
---

# Troubleshooting

| Symptom | First check |
| --- | --- |
| The request fails | Confirm the URL is reachable and that the host accepts your client and User-Agent |
| A metadata field is missing | Inspect the source document; `header` does not fabricate absent properties |
| A rendered output is unexpected | Run `header extract --style json URL` and compare the raw parsed properties |
| A generated card fails | Verify destination permissions and image availability |

To diagnose a page with the built-in rules:

```sh
header extract --test https://example.com
header verify https://example.com
```

Before publishing diagnostics, remove private URLs, credentials, cookies and personal data. Reproduce the problem with a small public test page if possible. To report a reproducible bug, use [GitHub Issues](https://github.com/arran4/header/issues).
