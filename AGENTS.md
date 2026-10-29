- Ensure output formats and templates are updated for new properties.

## API and template output contracts

- When a property changes, trace it through all *supported* template and output formats, including documentation and user-visible defaults. Add one understandable test per affected format and at least one negative/empty case when applicable; don't mechanically touch unrelated templates.
- Preserve existing caller-visible names, ordering and formatting unless the change explicitly defines a new output contract. Separate semantic input from rendering where possible; do not normalise source values destructively if exact representation is promised.
- If the project has a check/preview mode, ensure it does not mutate user files. Do not create a dry-run command just for compliance with generic policy.
- Prefer narrow input/output seams and small fixtures, but retain actual filesystem integration checks where the output path, permissions or write semantics are being tested.

- Regenerate files under `cmd/` from `internal/cli` definitions with `go generate ./...`; do not hand-edit generated files.

## Engineering guidance and agent audit

Use https://arran4.github.io/blog/post/2026/055-suggested-coding-standards-for-repositories-and-agents/ selectively, without superseding the more specific rules above. A coding agent must inspect implementation/tests, distinguish verified gaps from hypotheses, and record results and existing issue overlap at https://github.com/arran4/header/issues/53. Do not equate adding instructions with completing the implementation audit; do not merge automatically.
