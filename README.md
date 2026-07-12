# abwi

[![CI](https://github.com/lohn/abwi/actions/workflows/ci.yaml/badge.svg)](https://github.com/lohn/abwi/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lohn/abwi)](https://goreportcard.com/report/github.com/lohn/abwi)
[![npm](https://img.shields.io/npm/v/@abwi/cli.svg)](https://www.npmjs.com/package/@abwi/cli)
[![PyPI](https://img.shields.io/pypi/v/abwi.svg)](https://pypi.org/project/abwi/)

> **English** | [日本語](README.ja.md) | [한국어](README.ko.md)

A CLI for reading and writing Azure Boards work items with first-class
Markdown support — what `az boards` makes hard, `abwi` makes the default.

## Why

- Azure Boards supports **native Markdown** in large text fields, but
  `az devops` cannot write it: setting a field's format requires an extra
  `multilineFieldsFormat` entry in the JSON Patch document, which the Azure
  CLI never sends. `abwi` adds it automatically, so Description, Repro Steps,
  Acceptance Criteria, and friends land as real Markdown — not HTML-escaped
  text.
- Every multiline field is addressed **generically** with
  `-f <refname>=<value>`, with curl-style input: `@file` reads a file, `@-`
  reads stdin, and a leading `\@` escapes a literal `@`. No per-field flags to
  memorize.

> [!WARNING]
> Switching a field to Markdown is **irreversible per work item** — once a
> field on a given work item stores Markdown, Azure DevOps does not allow
> converting it back to HTML. This is an Azure DevOps limitation, not an
> `abwi` one. If your organization or process is not ready for Markdown, use
> the [`--format html` fallback](#the---format-html-fallback).

## Install

### via npm

```bash
npm install -g @abwi/cli
abwi --help
```

### via PyPI

```bash
pip install abwi
abwi --help
```

### via GitHub Releases

Download the binary for your platform from [Releases](https://github.com/lohn/abwi/releases).

### via Go

```bash
go install github.com/lohn/abwi/cmd/abwi@latest
```

## Authentication

**Entra ID (default).** Sign in once with the Azure CLI; `abwi` picks up the
Azure CLI credential automatically:

```bash
az login
```

**PAT (explicit opt-in, not recommended).** Set `auth = "pat"` in the
**global** config file (or pass `--auth pat`) and export the token as
`ABWI_PAT` (falling back to `AZURE_DEVOPS_EXT_PAT`, the variable the Azure
CLI uses):

```bash
export ABWI_PAT=...          # or AZURE_DEVOPS_EXT_PAT
abwi --auth pat show 123
```

PAT values are read **only** from environment variables, never from config
files, so a token can't end up committed by accident.

The `auth` key is honored **only in the global config file** — a checked-out
repository must not be able to switch your authentication mode, so `auth` in
a repo-local `.abwi.toml` is ignored with a warning.

## Configuration

`abwi` merges configuration from two TOML files:

- **Local**: `.abwi.toml`, found by walking up from the current directory
  (put it at your repository root)
- **Global**: the per-user config directory —
  `~/.config/abwi/config.toml` on Linux (Go's `os.UserConfigDir`, so
  `~/Library/Application Support/abwi/config.toml` on macOS and
  `%AppData%\abwi\config.toml` on Windows)

Precedence, highest first: **flags > environment (`ABWI_ORG`,
`ABWI_PROJECT`) > local file > global file**.

```toml
# .abwi.toml — committed at the repository root
org = "https://dev.azure.com/myorg"
project = "MyProject"
default-type = "Product Backlog Item"

# Shorthand names for -f, expanded to full field reference names
[aliases]
ac = "Microsoft.VSTS.Common.AcceptanceCriteria"
repro = "Microsoft.VSTS.TCM.ReproSteps"
```

Available keys:

| Key            | Description                                                         | Default    |
| -------------- | ------------------------------------------------------------------- | ---------- |
| `org`          | Organization URL (`https://dev.azure.com/<org>`)                    | —          |
| `project`      | Project name                                                        | —          |
| `format`       | Large text format: `markdown` / `html`                              | `markdown` |
| `auth`         | Authentication: `entra` / `pat` — **global config only**            | `entra`    |
| `default-type` | Work item type used when `create --type` is omitted                 | —          |
| `[aliases]`    | Table of shorthand names for `-f`, expanded to full reference names | —          |

Run `abwi config` to see the resolved values and where each one came from —
see [Usage](#usage) below.

## Usage

Create a work item (`--type` falls back to `default-type` from the config;
`-d`/`-f` values support `@file` and `@-`):

```bash
abwi create -T Bug -t "Crash when saving a draft" \
  -d @description.md \
  -f Microsoft.VSTS.TCM.ReproSteps=@repro.md

# The same, using the [aliases] shorthand from the config
abwi create -T Bug -t "Crash when saving a draft" -d @description.md -f repro=@repro.md
```

Update fields of an existing work item (Markdown from stdin via `@-`; use
`\@` when a value must start with a literal `@`):

```bash
generate-criteria | abwi update 123 -s Active -f ac=@- \
  -f System.Description='\@mentions start with an escaped at-sign'
```

Show a work item (`--json` for the raw response):

```bash
abwi show 123
```

List work items — yours by default, most recently changed first; filter with
flags or take over with a full WIQL query:

```bash
abwi list -T Bug -s Active --limit 20
abwi list --all                            # everyone's, not just yours
abwi list --assignee "someone@example.com" # someone else's
abwi list --wiql @query.wiql               # full control
```

Comments (posted as Markdown by default):

```bash
abwi comment add 123 "Reproduced on \`main\`; see #456."
abwi comment add 123 @-        # comment body from stdin
abwi comment list 123
```

Link and unlink (`--type`: `parent`, `child`, `related` — the default — or a
full `System.LinkTypes.*` reference name):

```bash
abwi link 123 456 --type parent   # make #456 the parent of #123
abwi unlink 123 456               # --type disambiguates multiple links
```

Show the resolved configuration, each value annotated with its origin
(`flag`, `env`, `local`, `global`, or `default`):

```console
$ abwi config
# global: /home/you/.config/abwi/config.toml
# local:  /home/you/src/myrepo/.abwi.toml
org = "https://dev.azure.com/myorg"  # local
project = "MyProject"  # env
format = "markdown"  # default
auth = "entra"  # default
default-type = "Product Backlog Item"  # local

[aliases]
ac = "Microsoft.VSTS.Common.AcceptanceCriteria"
repro = "Microsoft.VSTS.TCM.ReproSteps"
```

`abwi config <key>` prints a single value; `--json` prints the whole thing as
JSON.

### The `--format html` fallback

If your organization does not support Markdown work items yet (or you want to
avoid the irreversible switch), pass `--format html` or set
`format = "html"` in the config. You still _write_ Markdown — `abwi` converts
it to HTML via [goldmark](https://github.com/yuin/goldmark) before sending,
and the fields stay in HTML format:

```bash
abwi create -T Bug -t "Crash when saving a draft" --format html -d @description.md
```

## License

MIT © [lohn](https://github.com/lohn)
