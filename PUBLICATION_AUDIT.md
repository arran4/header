# Publication-readiness audit — 29 October 2026

Assessment: **CONDITIONAL GO after the conditions below are satisfied.**
No confirmed credential exposure was found. The proposed tree is ready for code
review; this is not approval to change visibility, merge, or publish a release.

## Publication conditions

1. Review the historical third-party artwork and committed executables described
   below. The new MIT licence does not grant rights to that material. Either
   establish the required permissions/notices or approve a separately reviewed
   history-remediation plan before making the complete history public. Nothing
   in this PR rewrites history or deletes remote branches.
2. Inspect the PR's hosted checks before merging. Local verification is recorded
   below; historical failed hosted runs are not passing test evidence. If Actions
   cannot start, restore its availability and rerun checks rather than treating
   skipped jobs as successful.
3. Confirm that exposing the existing maintainer commit identities, public sample
   social-media identifiers and past issue/review conversations is acceptable.
   No confidential customer data or non-test internal service endpoint was found.

A future binary release additionally needs a working tap-scoped token, a decision
on macOS signing/notarization, and an actual tagged release. These are release
prerequisites, not reasons to change visibility during this task.

## Recovery and scope

The interrupted `/workspace` tree was present on `main` at
`a548acb84bf08c8d4bd8dbd84e14487a682d6da2`. Before fetching or generating,
`/tmp/header-handoff/` received binary working/index patches, the untracked file
list, refs/reflog/status, and a tar archive of the tracked and untracked files.
The new branch is `codex/public-release-readiness`, based on that same current
`origin/main` commit. No reset, clean, destructive checkout or automatic conflict
resolution was used.

Recovered files included `LICENSE`, `SECURITY.md`, the AppLinks/Instagram/theme
fixture, `tools.go`, CI/release/README changes, source annotations, generated
commands, test fixes and staged removal of three `.orig` files. The recovered
AGENTS.md matched closed, unmerged PR #54; its useful guidance is retained.
Closed PR #55 had no implementation diff and is not reopened.

`tools.go` was replaced by Go's native `tool` directive, supported since Go 1.24,
with the dependency version in go.mod. The generated commands were regenerated,
not discarded or manually patched. New work adds absent metadata coverage,
process-level dispatch/error regressions, deterministic generation, attribution
packaging and an original project card. No metadata parser or template-output
architecture was redesigned.

## Issue #53: implementation and negative-path evidence

Checked sources include `types.go:13` (JSON fields), `types.go:152`
(`GroupedProperties`), `types.go:216` (theme/AppLinks/Instagram rows),
`applinks.go`, `htmlhead.go`, `templates/*.tmpl`,
`templates/templates_test.go:69`, `extractor_test.go:17`, `cli_test.go:97`,
`internal/cli/{root,extract,edit}.go`, `cmd/header/root.go`, and the CLI txtars.

| Representation | Verified contract |
| --- | --- |
| `all_fields`, `all_fields_plain` | Both render GroupedProperties; all six AppLinks fields, Instagram owner ID and theme colour appear with stable names and order. |
| `json` and extractor JSON fixtures | PageInfo uses its existing exported names and omitempty; the new fixture checks exact JSON formatting and field order. |
| `default`, `discord` | Concise summaries deliberately omit these extra fields. |
| `jsonld`, `opengraph`, `twitter` | Domain-specific outputs deliberately select their own fields and omit unrelated metadata. |

`templates/testdata/applinks_instagram_themecolor.txtar` runs all eight templates,
not merely the full-output pair. `metadata_absent.txtar` runs the same eight and
checks absence of placeholder rows/JSON keys. Existing YouTube/all-fields,
canonical-link, grouping, missing-field and extractor fixtures remain in use.
No source value normalization or field/template output change was needed.
README and generated usage now list actual supported styles; `plain` was an
obsolete advertised style, not an existing supported template.

`testdata/cli/txtar/dispatch.txtar` exercises root URL extraction, explicit
`extract`, `show`, help/version, `--` before URLs, unknown flags, missing URL and
unknown template errors using the local fixture server. `cli_dispatch_test.go:14`
uses a local rejecting HTTPS proxy to prove `-- extract`, `-- help` and
`-- version` reach the URL action without external DNS or network requests.
`edit_errors.txtar` verifies rejected edits preserve the source file and missing
input does not create a file; existing edit txtars retain real filesystem/stdin
coverage. The test runner now supports stderr assertions. The fixture server no
longer reads a global server pointer from a handler during server startup.

The original dispatch defect is already fixed upstream:
[arran4/go-subcommand#466](https://github.com/arran4/go-subcommand/issues/466) is
closed; [PR #491](https://github.com/arran4/go-subcommand/pull/491) merged as
`e19afc83200134db0887eee2f9eb13750e2476fe`, and the
[v0.0.31 release](https://github.com/arran4/go-subcommand/releases/tag/v0.0.31)
includes that merge. No duplicate upstream issue or dispatch workaround is added.

## Generation and compatibility

Canonical command: `go generate ./...`; pinned tool: **gosubc v0.0.31**.
`go tool gosubc` resolves the module tool rather than a binary on PATH.
Source `Flags:` blocks now have a blank line and tab indentation. Generation
emits no formatting warnings. Timestamps and project-commit provenance are
disabled; GeneratorVersion remains v0.0.31.

The sole template overlay is the upstream main template with its nested,
unpinned generation directive replaced by a pointer to the canonical command.
Without that change, `go generate ./...` would run a second generation from
`cmd/header`, silently reintroducing PATH selection and timestamps. See
`internal/generate/README.md`. No root/dispatch template is overridden.

Expected generated differences: registered commands run before the root URL
action; `--` prevents command recognition; ordinary action errors receive one
context prefix; wrapped exit-code errors use errors.As. The newer generator
also changes import formatting, canonical flag names in generated tests and
flag helper code. Source diagnostic strings are lowercased where required by
the existing staticcheck rules; consumers should not parse error prose. Metadata
field names, ordering and rendered template bytes are unchanged.

## Confidentiality and history coverage

Fetched current branches, tags and all available PR heads. The remote inventory
contained five branches, no tags and no releases. Together with local historical
refs, 214 commits and 768 reachable blobs (80,232,495 bytes) were inventoried
before this PR. Deleted paths and old blobs were included, not just main's tree.
Gitleaks scanned full history and a separate export of every reachable blob.
TruffleHog independently scanned Git history with live credential verification
disabled. Candidate values were never included in reports or GitHub comments.

GitHub inventory covered all 55 issues/PRs, all 138 conversation comments,
260 submitted reviews and 371 inline review comments, including closed PRs #54
and #55. Bodies and API responses were archived privately in `/tmp/header-audit/`
and scanned with both scanners. The single discussion candidate is a public
GitHub App OAuth client identifier in API metadata, not a client secret.

| Scan | Exit / redacted result |
| --- | --- |
| Gitleaks full Git history (`--all --full-history --redact`) | 0; no findings (151 commits with scanable diffs, including reachable histories). |
| Gitleaks every reachable blob export | 0; no findings. |
| Gitleaks working tree | 0; no findings. |
| Gitleaks archived GitHub discussions | 0; no findings. |
| TruffleHog Git (`--no-verification --no-update --json`) | 0; 30 unverified GitLab-pattern candidates, all false positives in the historical GoReleaser executable. |
| TruffleHog working tree (excluding .git) | 0; no candidates. |
| TruffleHog archived GitHub responses | 0; one public App client-ID false positive. |
| Gitleaks / TruffleHog 112 available workflow log files | Both exit 0; zero findings/candidates. |

All 30 binary candidate strings also occur in the independently installed
GoReleaser executable. The historical executable's build metadata identifies
GoReleaser v2.1.0 / Go 1.22.5. No live validation or credential rotation was
needed for those compiled strings. A scanner exit of 0 alone was not used to
classify TruffleHog findings: every candidate was reconciled separately.
Gitleaks was installed from the existing workspace toolchain (its version command
has no embedded release version); TruffleHog reports 3.99.7.

Personal data found consists of maintainer commit identities/contact addresses,
bot identities, publicly recognizable sample social-account IDs and example
person names. Historical issue diagnostics contain a private-range DNS resolver
address; source blobs use loopback test servers. These are local/test context,
not evidence of credentials or a confidential deployed service. Workflow secret
references are GITHUB_TOKEN and TAP_GITHUB_TOKEN, never their stored values.

All 299 listed Actions runs were inspected for downloadable logs. 27 runs
provided 112 log files; 55 returned empty archives and 217 downloads failed
(all 217 returned HTTP 410: expired/unavailable). Both scanners
checked the available log contents. Repository artifact listing returned zero
artifacts. Sample latest/main/card job log endpoints returned BlobNotFound;
run-log downloads for those runs were empty ZIPs. Empty or unavailable logs are
not proof that the underlying commands ran successfully. The audit does not
cover deleted GitHub discussions, expired/deleted artifacts/logs, inaccessible
server-side dangling objects, repository secret values, or private vulnerability
reports. Secrets are deliberately not retrieved. The task did not authorize a
history rewrite; if a real credential is discovered later, rotate/revoke it
first, then review history and cached GitHub surfaces before publication.

## Licensing and historical material

MIT matches the pre-existing GoReleaser licence declaration and the author's
project metadata. `LICENSE` is the standard MIT text, with the maintainer's
name and 2026 date. The author still controls permission to license original
contributions. Dependency terms remain separate: `THIRD_PARTY_NOTICES.md`
contains module-graph licence texts, the Go runtime licence and the embedded
Bigelow & Holmes font notice. Archives and Linux packages include those notices.
Small HTML fixtures contain constructed metadata, titles and sample identifiers,
not complete scraped articles, lyrics or media files.

The old `header-card.png` (blob
`29cbc6c95e487a124623bc2f70762933b4278676`, introduced by `57f8012`)
contains GitHub's error-page Octocat illustration and “Page not found” title.
It is replaced with a card generated only from original project text,
placeholders, a local QR code and the attributed Go font. The new workflow uses
`.github/scripts/header-card.go`, so a private/404 GitHub page cannot silently
become the README card again. Both basic and darkmode generation were checked.
Historical artwork is still reachable; MIT does not relicense it. Consult
[GitHub's brand/legal guidance](https://brand.github.com/foundations/logo#legal)
and obtain permission or approve removal of the historical asset before public
exposure.

The abandoned remote branch `fix-goreleaser-v2-695992398468588233` at `7ba9683`
contains a 65,593,496-byte `goreleaser` binary (blob
`f203332741a45ba048b226aa93090d22cb6f5b92`). Another old commit introduced a
12,698,143-byte `header` binary (blob
`3f970d60bdc421675864fe5291e5ddec7256ad2e`, introduced by `a46541d`, later
removed by `4f9f310`). They are not in the proposed tree, but remain reachable.
Their source/dependency attribution and redistribution terms are not proven by
this PR's current notices. Review retention/removal explicitly. Removing a
branch alone may leave PR refs reachable. No malware analysis or execution of
historical executables was performed.

## CI, release and deployment concerns

Available default-branch schedule run
[36938140662](https://github.com/arran4/header/actions/runs/36938140662)
actually failed staticcheck ST1005 in internal/cli/edit.go; the retained diagnostic
casing fixes and final v2 lint run address that verified failure. An older release run
[30791693599](https://github.com/arran4/header/actions/runs/30791693599)
failed because the repository had no tags; this audit used snapshot mode and did
not manufacture a tag. The latest
main/card/PR runs had no steps/logs available, so an Actions-budget explanation
from the previous audit remains unverified here.

CI defaults to contents:read. Release contents:write is job-scoped; card automation
has contents/pull-requests write only in its own job. Lint is pinned to v2.15.0,
uses go.mod's Go version, and keeps the v2 configuration. The recovered v1 config
was unsuitable for the v2 CI action and was corrected. CI now checks generation
without modifying user files, vet, race tests and GoReleaser config. Generation
runs in Actions' disposable checkout; no new dry-run contract is invented.

GoReleaser v2.19.0 config uses current archive/cask keys and an explicit eight-
target matrix. A safe snapshot created all eight binaries/archives, 12 Linux
packages, checksums and a local Homebrew cask; publish/announce were skipped.
That is stronger evidence than config validation, but it does not prove live
release/tap credentials work. TAP_GITHUB_TOKEN is now passed to GoReleaser; its
existence and grants were not inspected. The cask publisher needs a separate
[tap-scoped write token](https://goreleaser.com/customization/publish/homebrew_casks/#github-actions).
No tags or releases were created and no Homebrew repository was updated.

macOS signing/notarization remains unconfigured and was not tested on macOS.
Decide that distribution policy before advertising a production cask; see
[GoReleaser's signing notes](https://goreleaser.com/customization/publish/homebrew_casks/#signing-and-notarizing).
The existing manual release-mode/version-override scaffolding is not a verified
version-bump workflow; actual tagged releases remain the supported publishing
trigger. Do not infer a successful release from a snapshot or empty Actions log.

URL hardening was evaluated without removing arbitrary user-selected URL access:

- `fetch.go:31` provides per-request timeout (20 seconds), a 2 MiB HTML reader cap,
  and a 20-hop request/redirect bound. Truncation at the cap is currently silent.
- Redirects are handled manually; metadata refresh and downstream image/URL lint
  fetches can cross hosts, including private/loopback addresses. No SSRF-safe
  service deployment boundary or overall cancellation context is provided.
- `normaliseURL` defaults bare inputs to HTTPS; HTTP transport ultimately rejects
  unsupported schemes. Its case-sensitive scheme handling is not a comprehensive
  URL validator. No non-HTTP file transport is registered.
- `internal/cli/cards/basic.go:64` / `:107`, `internal/ui/image.go:40` and
  `lint/rules/images/images.go:52` have timeouts, but image byte/pixel limits are
  absent. Timeouts do not bound image memory allocation. Shiny decodes directly.
- Viewer temp files use os.CreateTemp and are not removed automatically because
  viewers start asynchronously; failed download/viewer paths can leave files too.
- Explicit HTTP diagnostics can include cookies, query strings and URL credentials.
  Full URLs may also appear in errors, redirect traces and Referer headers.

These are documented local-CLI limitations, not a claim of safety for an
unrestricted privileged URL-fetching service. Before such a deployment, require
network egress controls, bounded images and credential-safe diagnostics.
Proposed separate hardening work: (a) byte/pixel limits with oversized-image and
cleanup regressions; (b) context-aware fetch/redirect/image cancellation and
sensitive diagnostic redaction. Those changes are not hidden in this release PR.

## Verification record

Go project checks used Go 1.26.0 on linux/amd64. The v2 linter binary was built
with Go 1.26.10; GoReleaser v2.19.0 reports Go 1.27.3. The old installed v1.64.8
linter is not the final CI-equivalent check.

| Command | Final result |
| --- | --- |
| `go test ./...` | Exit 0, including CLI and template txtars. |
| `go test -race -count=1 ./...` | Exit 0. |
| `go vet ./...` | Exit 0. |
| `/tmp/header-tools/golangci-lint run ./...` (v2.15.0) | Exit 0; zero issues. |
| `go generate ./...` | Exit 0; no warnings. Repeated generation has identical generated bytes. |
| `go test ./templates -run 'TestTemplates/(applinks_instagram_themecolor|metadata_absent)' -count=1 -v` | Exit 0; 16 template scenarios. |
| `/home/user/bin/goreleaser check` | Exit 0; config validated. |
| `/home/user/bin/goreleaser release --snapshot --clean --skip=publish,announce --parallelism 2` | Exit 0; eight build targets, archives, packages, cask and checksums. |
| Snapshot executable `version`, `--help` | Exit 0; generated help and snapshot build metadata. |
| `go run ./.github/scripts/header-card.go --type basic` and `--type darkmode` | Exit 0; PNGs produced locally. |
| `git diff --check` and ignored-artifact checks | Exit 0; root header/header.exe and dist are ignored; three tracked .orig files removed. |

Clean-worktree generation, the final snapshot smoke test, hosted CI and final head
SHA are recorded with the PR/issue delivery evidence so they refer to the exact
committed revision rather than a self-referential SHA in this file.

Recommended repository metadata (not changed by this PR): description “Extract
and inspect web metadata from your terminal”; homepage
`https://github.com/arran4/header#readme`; topics `go`, `cli`, `metadata`,
`opengraph`, `twitter-cards`, `json-ld`, `web-scraping`. Current repository metadata
has no description/homepage/topics. Visibility remains private.
