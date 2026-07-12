# Contributing

Thanks for your interest in contributing! This document describes the conventions
and workflow for this repository.

This project follows a [Code of Conduct](./CODE_OF_CONDUCT.md). By participating,
you are expected to uphold it.

## Language

**All communication in this repository must be in English.** This includes, but is
not limited to:

- Commit messages
- Issues
- Pull requests (titles and descriptions)
- Code comments and documentation
- Review discussions

English does not have to be your native language, and you are very welcome here
regardless of your fluency. Feel free to use translation tools — we only ask that
your communication is clear and concise. Don't let language hold you back from
contributing.

## Commit messages and pull request titles

**Commit messages and pull request titles must follow the
[Conventional Commits](https://www.conventionalcommits.org/) specification.**

The format is:

```
<type>(<optional scope>): <description>
```

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`,
`ci`, `chore`, `revert`.

Examples:

```
feat: add comment add and comment list commands
feat(ado): emit multilineFieldsFormat ops for markdown fields
fix(config): stop the .abwi.toml search at the filesystem root
docs: document the --format html fallback
ci: pin actions to commit SHAs
```

This is enforced by [commitizen](https://commitizen-tools.github.io/commitizen/) via
the pre-commit hooks (`commit-msg` and `pre-push` stages). Since pull requests are
expected to use a Conventional Commit title as well, keep the PR title consistent
with the change.

## Development setup

This project uses [mise](https://mise.jdx.dev) to manage the toolchain and
[pre-commit](https://pre-commit.com) (run via [prek](https://github.com/j178/prek))
for the hooks.

```sh
# Install the pinned toolchain (prek, dprint, commitizen, etc.)
# and wire up the git hooks (via mise's postinstall hook)
mise install
```

`mise install` runs a `postinstall` hook that installs the prek git hooks for
you, so a fresh clone needs only this one command. If you ever need to wire the
hooks up by hand, run `prek install`.

If mise is [activated](https://mise.jdx.dev/getting-started.html) in your shell,
the toolchain is on your `PATH` and the commands below can be run as-is.
Otherwise, prefix them with `mise exec --` (e.g. `mise exec -- go test ./...`).

## Before you open a pull request

Most checks run automatically through the git hooks (assuming `prek install`):

- **`pre-commit`** — formatting, `golangci-lint`, and `go mod tidy`
- **`pre-push`** — `go test ./...`

So a normal `git commit` and `git push` already run the formatters, linters, and
the test suite. To run the pre-commit hooks on demand:

```sh
prek run
```

`mise.lock` is generated automatically by the `mise-lock` hook whenever a mise
configuration file changes, so commit the updated lockfile together with your
configuration change.

## Project layout

The implementation is entirely in Go, in `cmd/` and `internal/`; `npm/` and
`pypi/` only package and ship the prebuilt binary.

```
cmd/abwi/          CLI entry point
internal/cli/      Command tree, flag parsing, rendering
internal/config/   TOML config loading/merging with origin tracking
internal/ado/      Azure DevOps SDK wrapper, JSON Patch building, REST fallback
npm/, pypi/        Binary-distribution wrappers — no real logic
```

See [AGENTS.md](./AGENTS.md) for a more detailed map and architecture notes.

## Testing

Tests live next to the code they cover in each `internal` package
(`internal/*/*_test.go`). No test hits the real Azure DevOps API — everything
runs offline against in-memory inputs.

```sh
go test ./...                    # run all tests
go test ./internal/ado -run Patch -v   # focus on a subset
```

The suite also runs automatically on `git push` via the `pre-push` hook, but
running it directly is faster while iterating.

Add or update tests for any behavior change.

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](./LICENSE).
