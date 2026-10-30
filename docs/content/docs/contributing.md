---
title: "Contributing"
weight: 50
---

# Contributing

The project is written in Go. Build from a checkout with Go 1.26 or later.

```sh
go test ./...
go test -race -count=1 ./...
go vet ./...
go generate ./...
```

Generated command sources under `cmd/` are produced from definitions under `internal/cli/` using the pinned `gosubc` dependency. Review [AGENTS.md](https://github.com/arran4/header/blob/main/AGENTS.md) and the [generation notes](https://github.com/arran4/header/blob/main/internal/generate/README.md) before changing commands.

## Documentation

The documentation source lives in `docs/` and uses Hugo Book as a pinned submodule. Preview with:

```sh
git submodule update --init --recursive
hugo server --source docs
```

Check the production build with `hugo --source docs --minify`. A PR builds the site but does not deploy it; merges to `main` deploy via GitHub Actions once the Pages source is configured.

Update an existing reader-oriented page when implementing behaviour rather than creating disposable implementation-plan files. For Go API signatures, update the source documentation and link to pkg.go.dev.
