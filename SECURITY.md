# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `header`, please report it privately through GitHub's Private Vulnerability Reporting feature, when enabled:

1. Navigate to the repository's **Security** tab on GitHub.
2. Under "Reporting", click **Report a vulnerability**.
3. Provide details and reproduction steps for the issue.

Please do not disclose security issues through public issues, pull requests, or discussions until they have been addressed.

If the private reporting button is unavailable, use the contact links on the
[maintainer's profile](https://github.com/arran4) to request a private channel.
Do not include credentials or sensitive reproduction data in a public issue.

## Fetching untrusted pages

This is a local CLI and library for fetching user-selected HTTP(S) URLs. It can
reach loopback and private networks, and redirects and page metadata may point
to other hosts. Do not expose it as an unrestricted server-side URL-fetching
service. Privileged deployments need their own network egress controls.

HTML extraction defaults to a 2 MiB body cap, a 20-second timeout per request,
and at most 20 request hops. Image viewing and card generation have request
timeouts but do not yet bound downloaded bytes or decoded pixel counts.
Use care with hostile pages, images and metadata. There is currently no overall
cancellation context spanning redirects, delays and downstream image requests.

HTTP diagnostics (`--httpretty` / `httpretty`) intentionally print request and
response details; treat the output as sensitive. Avoid credential-bearing URLs
and review cookies, query strings and headers before sharing logs. Image viewer
fallbacks leave temporary `header-cover-*.jpg` files in the system temporary
directory for asynchronous viewers; remove these files when no longer needed.
