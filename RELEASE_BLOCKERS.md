# Final release blockers — 29 October 2026

This follows up PR #56 at `8528cee25b24a534b6522864fbb6fd2b99c5f10d`.
It narrows the remaining decisions; it does not repeat the publication audit.

## 1. Hosted Actions scheduling

[Run 40832281247](https://github.com/arran4/header/actions/runs/40832281247)
accepted the workflow and created its job graph. `Route event` failed in about
two seconds with `runner_id=0`, no runner name and zero steps. Its three
dependent jobs were skipped. The check run advertises two annotations, but
the available PAT cannot read them (HTTP 403); the connector can list the checks
but does not expose the annotation endpoint. Empty job logs do not identify
the billing cause.

The read-only account Actions-billing and October usage requests also returned
HTTP 403. Account budget state could not be established. A request to a
personal-budget route returned 404 and is not evidence that no budget exists.
No payment details, balances or credential values were retrieved or recorded.

Public repositories did schedule real runners on the same date:

- [go-subcommand run 40830167358](https://github.com/arran4/go-subcommand/actions/runs/40830167358)
  at 06:35 UTC passed, including Route event and code checks on assigned runners.
- [g2 run 40796380220](https://github.com/arran4/g2/actions/runs/40796380220)
  at 00:47 UTC passed.

This is a pre-execution scheduling/admission failure, not an observed failure
of the route shell script, checkout, generator, tests or linter. A private Actions
quota/payment restriction or account/repository hard budget is the leading
explanation, not a verified specific account setting. Public success does not
rule it out: standard hosted runners in public repositories are free, while
private usage draws on the owner's allowance and payment/budget controls.
See [Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
and [budget scopes and stopping usage](https://docs.github.com/en/billing/concepts/budgets-and-alerts).

Owner action: read the two run annotations, then inspect the personal account's
Billing & licensing usage, account/repository Actions budgets with stopping
enabled, allowance exhaustion, and any payment restriction. Resolve the stated
restriction or contact GitHub Support if none is shown, then rerun the failed
jobs on the current PR head while the repository is still private. Increasing
paid limits or changing a payment method is an owner decision. No speculative
CI YAML edits or repeated reruns are warranted from the evidence available.
The two annotation messages and a redacted account restriction status were
requested from the owner because the available tools cannot read them.

## 2. Exact historical reachability

The following came from exact blob membership in `git rev-list --objects` for
every ref advertised by `git ls-remote origin`, including PR head/merge refs.
It is more precise than checking an introducing commit's ancestry alone.

| Material | Blob / introduction | Advertised reachability at review |
| --- | --- | --- |
| Old GitHub 404/Octocat card, 36,531 bytes | `29cbc6c95e487a124623bc2f70762933b4278676`, introduced by `57f80121d43de3cb8d8ce86e92dfd45a21330994` | In main and PR #56 ancestry; present at main's tip and PR head tips #49, #50, #51, #52, #54 and #55. Also reachable through #56 head/merge and the PR branch, whose tips use the replacement card. |
| GoReleaser executable, 65,593,496 bytes | `f203332741a45ba048b226aa93090d22cb6f5b92`, introduced by `7ba9683aad489403538ca5aa1e7ba61f025ac72b` | Only `refs/heads/fix-goreleaser-v2-695992398468588233`. No advertised PR head/merge ref contains it; commit-to-PR lookup returns no associated PR. Not reachable from main or #56. |
| Old header executable, 12,698,143 bytes | `3f970d60bdc421675864fe5291e5ddec7256ad2e`, introduced by `a46541dec96cb8ca7cfb91265375854812770829`, deleted by `4f9f31006cb2b28f477ed162652928bc0c89e2e9` | In all six advertised branch histories, all 55 advertised PR-head histories, and #56's merge-ref history. Not present at any advertised tip. |

`HEAD` aliases main. Local stale branches/remotes are not counted as live remote
branches. No tags are advertised. This identifies reachable refs, not GitHub's
unadvertised caches, detached commit URLs, forks, or other people's clones.

### Licensing and recommended disposition

**Octocat card: prefer targeted removal unless explicit retention permission is
obtained.** This is GitHub's illustration in a product's project card, not
project-owned artwork covered by MIT. A copyright credit alone is not a new
license. GitHub's policy permits some personal-avatar/link uses, but that does
not establish permission for this product-card use. See the
[logo policy](https://docs.github.com/en/site-policy/other-site-policies/github-logo-policy)
and [brand legal terms](https://brand.github.com/foundations/logo#legal).
PR #56 fixes the current tree; it does not erase the historical image.
Permission to retain it would avoid a disruptive history operation.

**GoReleaser binary: recommend removing the abandoned branch, subject to owner
approval.** Build metadata identifies GoReleaser v2.1.0 / Go 1.22.5 and 283
dependency modules. [GoReleaser v2.1.0 is MIT licensed](https://github.com/goreleaser/goreleaser/blob/v2.1.0/LICENSE.md),
so redistribution of that upstream software is permitted under its notice
conditions. The executable is byte-for-byte identical to `goreleaser` inside the official
v2.1.0 Linux x86_64 archive, whose SHA-256 matches upstream checksums.txt.
The upstream archive includes LICENSE.md. Its accompanying SBOM declares
NOASSERTION for many packages, so it is not a complete dependency-license
inventory. MIT for the main
program does not waive dependency license/NOTICE obligations.
The project's current notices cover header's graph, not those 283 old modules.
Retention would need the version-specific dependency notices and any applicable
source/other obligations checked. Removing an obsolete tool artifact avoids
that unrelated distribution burden. Deleting this branch removes the only
advertised ref to this blob, without rewriting main; direct cached access and
server garbage collection cannot be guaranteed merely by branch deletion.

**Old header binary: recommend retention with the historical notices now added
to THIRD_PARTY_NOTICES.md.** Its embedded build information records Go 1.24.3,
`-tags=shiny`, nine dependency modules, source revision `2e96d3393760b55bb6edcb580202e47b3f1bee05`
and `vcs.modified=true`. All nine source-module checksums match the embedded
checksums. Four license texts were already included; the five older Go modules
share a BSD text with older copyright/Google Inc. wording, now included exactly.
The Go 1.24.3 runtime license already matches the runtime notice. These are
MIT/BSD terms requiring attribution, not an established prohibition on retaining
the executable. It is dynamically linked to libc rather than bundling libc;
no historical executable was run. It remains unsupported, not reproducibly
identified with a clean source tree, and unsuitable as the first official
release. Standalone redistribution must carry the notices. Retention is far less
disruptive than rewriting every maintained branch and old PR ancestry for a
non-secret artifact. Owners may still approve removal for repository size or
provenance reasons.

### What an approved removal would and would not accomplish

1. Keep a private backup and an exact ref/object manifest before any destructive
   action. Any rewrite is a separate, explicitly approved operation; no rewrite
   or branch deletion has been performed here.
2. For the GoReleaser binary, deleting only the abandoned branch is sufficient
   to remove all *advertised* references currently found. Do not claim it purges
   GitHub's cached objects or outside clones.
3. To remove the Octocat image from maintained branch history, a future rewrite
   should target that exact blob, preserving the replacement card and unrelated
   source changes. Removing every historical `header-card.png` indiscriminately
   would also discard the new original asset. Removing the old header executable
   would similarly target its exact blob and every affected maintained branch.
4. Force-pushing maintained branch histories does not update GitHub's read-only
   `refs/pull/*` namespace. Old PR heads listed above would still expose the image;
   all old PR heads would still expose the header executable. Closing a PR or
   deleting its branch does not delete those refs. Rewrites also change commit
   IDs/signatures and can invalidate old diffs/reviews; collaborators need clean
   clones or careful rebases to prevent reintroduction.
5. GitHub documents PR/cache cleanup via Support for qualifying sensitive-data
   removal, and explicitly says it does not remove non-sensitive data through
   that process. Do not assume this non-secret artwork/binary cleanup qualifies.
   See [history-removal limitations](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
   and [read-only PR refs](https://docs.github.com/en/pull-requests/how-tos/review-pull-requests/checking-out-pull-requests-locally).
   If confirmed removal of the artwork from the public surface is required and
   permission/appropriate GitHub assistance is unavailable, keep the old repository
   private and publish an approved clean-history copy in a separate repository.
   A new public copy, migration, rename, or recreation also requires approval.

## 3. PR review and decision gates

At the reviewed head GitHub reports `mergeable=true`, `draft=false`, with current
main `a548acb84bf08c8d4bd8dbd84e14487a682d6da2` as base. The PR has no reviewer
submissions or inline review comments. The source diff, generator overlay,
fixtures, generated output, release configuration and attribution packaging were
reviewed without unrelated refactors. The suspected need for an extra Homebrew
PR token was checked against GoReleaser v2.19.0 source: cask publishing first
selects the repository token, and an empty separate PR token reuses that client.
The existing TAP_GITHUB_TOKEN wiring is therefore sufficient in configuration;
actual token presence/grants remain an owner release prerequisite.

`Fixes #53` is justified for the issue's guidance plus implementation/test audit
acceptance: repository-specific instructions, all applicable output formats,
negative/error paths, recorded commands/results, existing upstream issue overlap
and proposed follow-ups are present in code and the issue report. #53 does not
require a public-visibility change or a production release. It remains open until
the PR is reviewed/merged; this report does not close it manually.

- **Before merging #56:** restore private Actions scheduling, rerun the latest
  head, inspect actual code checks and obtain normal review approval. The historical
  retention/removal decisions need not be executed to merge this non-destructive PR.
- **Before changing the existing repository to public:** resolve the historical
  Octocat rights/removal path, approve the retained old header binary with notices,
  approve removal or fully attributed retention of the abandoned GoReleaser branch,
  and accept the already identified existing personal metadata. Recheck any approved
  cleanup against remote PR refs; a new file or ordinary branch rewrite is not proof
  of complete removal. Do not change visibility as a billing workaround.
- **Before the first binary release:** merge the reviewed changes with successful
  checks; choose an explicit version/tag; validate the release token and tap token's
  contents/PR grants without disclosing values; decide macOS signing/notarization
  and communicate any unsigned-install limitations; use fresh binaries, bundled
  license/third-party notices and verified checksums, not historical executables.
  Snapshot/config success does not demonstrate a live release or tap publication.
