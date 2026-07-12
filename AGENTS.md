# AGENTS.md

Guidance for AI coding agents working in this repository.

## What this project is

`abwi` is a Go CLI for reading and writing Azure Boards work items with
first-class Markdown support. Azure Boards can store large text fields as
native Markdown, but `az devops` cannot write that format; `abwi` makes it
the default by adding the required `multilineFieldsFormat` JSON Patch entry
automatically. The scope is deliberately narrow: work item read/write only
(`create`, `update`, `show`, `list`, `comment`, `link`, `unlink`, `config`) —
no other Azure DevOps services.

## Repository layout

```
cmd/abwi/          CLI entry point
internal/cli/      Command tree, flag parsing, rendering
internal/config/   TOML config loading/merging with origin tracking
internal/ado/      Azure DevOps SDK wrapper, JSON Patch building, REST fallback
```

The implementation lives entirely in `cmd/` and `internal/`. `internal/cli`
owns everything user-facing (flags, `@file`/`@-` expansion, output rendering),
`internal/config` resolves `.abwi.toml` / the global config / environment
variables with per-key origin tracking, and `internal/ado` is the only package
that talks to Azure DevOps.

## Development

This project uses [mise](https://mise.jdx.dev) for the toolchain and
[pre-commit](https://pre-commit.com) (run via [prek](https://github.com/j178/prek)).

```sh
mise install      # install pinned toolchain AND wire up git hooks
```

`mise install` runs a `postinstall` hook that also installs the prek git hooks,
so this one command is enough on a fresh clone. Run `prek install` by hand only
if you need to re-wire the hooks.

mise provides the pinned toolchain. If it is activated in your shell the tools
are on `PATH`; otherwise prefix commands with `mise exec --`.

Day-to-day commands:

```sh
go build ./...
go test ./...
go vet ./...
```

The git hooks (installed by `mise install`) cover the rest automatically:
formatting, `golangci-lint`, and `go mod tidy` on `pre-commit`, and
`go test ./...` on `pre-push`. To run the pre-commit hooks on demand:

```sh
prek run          # against staged files
```

`golangci-lint run` / `golangci-lint fmt` are still handy while iterating.

## Conventions

- **Commit messages and PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/)**
  (`feat`, `fix`, `docs`, `refactor`, `test`, `chore`, …). Enforced by commitizen
  via the `commit-msg` / `pre-push` hooks. See [CONTRIBUTING.md](./CONTRIBUTING.md).
- All repository communication (commits, PRs, comments, code) is in **English**.
- Go code is formatted with **gofumpt + goimports** (local prefix
  `github.com/lohn/abwi`); non-Go files with **dprint**. Let the hooks
  format — don't hand-format.
- Keep the three READMEs (`README.md`, `README.ja.md`, `README.ko.md`) consistent
  when changing user-facing behavior.
- Japanese/Korean documentation style: translations must read as natural,
  idiomatic technical writing in each language — never a literal rendering of
  the English text. In Japanese, do not hard-wrap in the middle of a sentence
  (GitHub renders the soft break as a stray half-width space between CJK
  characters); break lines only after 。/、 or adjacent to ASCII tokens such as
  English words and inline code, or leave the paragraph unwrapped. Korean may
  wrap at any inter-word space.
- Unit tests live next to the code they cover (`internal/*/*_test.go`); no
  mocking frameworks. Add or update tests for behavior changes.
- Azure DevOps API notes:
  - Writing a large text field as Markdown requires a JSON Patch op on
    `/multilineFieldsFormat/<ref>` with value `"Markdown"`, alongside the
    usual `/fields/<ref>` op (see `internal/ado/patch.go`).
  - Work item comments go through plain REST with
    `?format=markdown&api-version=7.1-preview.4` — the Go SDK's comments
    client has no format parameter (see `internal/ado/comment.go`).
  - The SDK v7 fields listing is `GetWorkItemFields`, returning
    `WorkItemField2` (see `internal/ado/workitem.go`).
