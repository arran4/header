# CLI generation

`go generate ./...` uses the Go tool dependency pinned in `go.mod` (v0.0.31).
It disables timestamps and project provenance so commits do not change output.

`templates/main.go.gotmpl` is the upstream v0.0.31 main template with only its
nested `go:generate` directive replaced by a pointer to the repository command.
The upstream directive would run a second, unpinned generation during
`go generate ./...`, overriding the deterministic output. No dispatch templates
are overridden. Review this small overlay when upgrading the generator.

`templates/agents.md.gotmpl` keeps generated contributor guidance consistent
with the module-pinned generation command and preserves the annotation examples.
