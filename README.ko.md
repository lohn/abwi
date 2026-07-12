# abwi

[![CI](https://github.com/lohn/abwi/actions/workflows/ci.yaml/badge.svg)](https://github.com/lohn/abwi/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lohn/abwi)](https://goreportcard.com/report/github.com/lohn/abwi)
[![npm](https://img.shields.io/npm/v/@abwi/cli.svg)](https://www.npmjs.com/package/@abwi/cli)
[![PyPI](https://img.shields.io/pypi/v/abwi.svg)](https://pypi.org/project/abwi/)

> [English](README.md) | [日本語](README.ja.md) | **한국어**

Azure Boards work item을 읽고 쓰기 위한 CLI입니다. Markdown을 일급으로 지원하며,
`az boards`로는 어려웠던 일이 `abwi`에서는 기본 동작이 됩니다.

## 왜 필요한가

- Azure Boards의 큰 텍스트 필드는 **네이티브 Markdown**을 지원하지만, `az devops`로는
  쓸 수 없습니다. 포맷을 지정하려면 JSON Patch 문서에 `multilineFieldsFormat` 항목을
  추가해야 하는데, Azure CLI는 이를 보내지 않기 때문입니다. `abwi`는 이 항목을
  자동으로 붙여 주므로 Description, Repro Steps, Acceptance Criteria 같은 필드가
  HTML로 이스케이프된 문자열이 아니라 진짜 Markdown으로 저장됩니다.
- 여러 줄 필드는 모두 `-f <refname>=<value>` 형태로 한꺼번에 지정할 수 있습니다.
  값은 curl 방식으로, `@file`은 파일을, `@-`는 표준 입력을 읽고, 맨 앞의 `\@`는
  리터럴 `@`의 이스케이프입니다. 필드마다 전용 플래그를 외울 필요가 없습니다.

> [!WARNING]
> 필드의 Markdown 전환은 **work item 단위로 되돌릴 수 없습니다**. 한 번 Markdown으로
> 저장한 필드를 HTML로 되돌리는 것은 불가능합니다. 이는 Azure DevOps의 사양이지
> `abwi`의 제한이 아닙니다. 조직이나 프로세스가 아직 Markdown으로 옮겨 갈 수 없다면
> [`--format html` 폴백](#--format-html-폴백)을 사용하세요.

## 설치

### npm을 통해

```bash
npm install -g @abwi/cli
abwi --help
```

### PyPI를 통해

```bash
pip install abwi
abwi --help
```

### GitHub Releases를 통해

[Releases](https://github.com/lohn/abwi/releases)에서 플랫폼용 바이너리를 다운로드하세요.

### Go를 통해

```bash
go install github.com/lohn/abwi/cmd/abwi@latest
```

## 인증

**Entra ID(기본값).** Azure CLI로 한 번 로그인해 두면 `abwi`가 그 자격 증명을
그대로 사용합니다:

```bash
az login
```

**PAT(명시적 옵트인, 권장하지 않음).** **글로벌** 설정 파일에 `auth = "pat"`을 쓰거나
`--auth pat`을 넘긴 뒤, 토큰을 환경 변수 `ABWI_PAT`로 내보냅니다(없으면 Azure CLI와
같은 `AZURE_DEVOPS_EXT_PAT`를 참조합니다):

```bash
export ABWI_PAT=...          # 또는 AZURE_DEVOPS_EXT_PAT
abwi --auth pat show 123
```

PAT는 오직 환경 변수에서만 읽으며 설정 파일에는 쓸 수 없습니다. 그래서 토큰을
실수로 커밋하는 사고가 일어나지 않습니다.

또한 `auth` 키는 **글로벌 설정 파일에서만** 유효합니다. 체크아웃한 저장소 쪽에서
인증 방식을 바꿀 수 있어서는 안 되므로, 저장소 로컬 `.abwi.toml`에 적힌 `auth`는
경고를 출력한 뒤 무시됩니다.

## 설정

`abwi`는 두 개의 TOML 파일에서 설정을 병합합니다:

- **로컬**: `.abwi.toml`. 현재 디렉터리에서 위쪽으로 올라가며 찾습니다(저장소
  루트에 두는 것을 권장합니다)
- **글로벌**: 사용자별 설정 디렉터리 — Linux에서는 `~/.config/abwi/config.toml`
  (Go의 `os.UserConfigDir`를 따르므로 macOS에서는
  `~/Library/Application Support/abwi/config.toml`, Windows에서는
  `%AppData%\abwi\config.toml`)

우선순위는 높은 순서대로 **플래그 > 환경 변수(`ABWI_ORG`, `ABWI_PROJECT`) >
로컬 파일 > 글로벌 파일**입니다.

```toml
# .abwi.toml — 저장소 루트에 커밋
org = "https://dev.azure.com/myorg"
project = "MyProject"
default-type = "Product Backlog Item"

# -f용 축약 이름. 전체 필드 참조 이름으로 확장됨
[aliases]
ac = "Microsoft.VSTS.Common.AcceptanceCriteria"
repro = "Microsoft.VSTS.TCM.ReproSteps"
```

설정 가능한 키:

| 키             | 설명                                                      | 기본값     |
| -------------- | --------------------------------------------------------- | ---------- |
| `org`          | 조직 URL(`https://dev.azure.com/<org>`)                   | —          |
| `project`      | 프로젝트 이름                                             | —          |
| `format`       | 큰 텍스트 포맷: `markdown` / `html`                       | `markdown` |
| `auth`         | 인증 방식: `entra` / `pat`. **글로벌 설정 전용**          | `entra`    |
| `default-type` | `create`에서 `--type`을 생략했을 때 사용할 work item 타입 | —          |
| `[aliases]`    | `-f`용 축약 이름 테이블. 전체 참조 이름으로 확장됨        | —          |

`abwi config`를 실행하면 최종 값과 각 값의 출처를 확인할 수 있습니다 — 아래
[사용법](#사용법)을 참고하세요.

## 사용법

work item을 만듭니다(`--type`을 생략하면 설정의 `default-type`이 사용됩니다.
`-d`/`-f` 값은 `@file`과 `@-`를 지원합니다):

```bash
abwi create -T Bug -t "Crash when saving a draft" \
  -d @description.md \
  -f Microsoft.VSTS.TCM.ReproSteps=@repro.md

# 같은 작업을 설정의 [aliases] 축약 이름으로
abwi create -T Bug -t "Crash when saving a draft" -d @description.md -f repro=@repro.md
```

기존 work item의 필드를 수정합니다(`@-`로 표준 입력에서 Markdown을 읽습니다.
값이 리터럴 `@`로 시작해야 한다면 `\@`로 이스케이프하세요):

```bash
generate-criteria | abwi update 123 -s Active -f ac=@- \
  -f System.Description='\@mentions start with an escaped at-sign'
```

work item을 표시합니다(`--json`은 원본 응답을 출력):

```bash
abwi show 123
```

work item을 나열합니다. 기본으로는 나에게 할당된 항목을 최근 변경 순으로 보여
줍니다. 플래그로 좁히거나 WIQL 쿼리로 통째로 바꿀 수도 있습니다:

```bash
abwi list -T Bug -s Active --limit 20
abwi list --all                            # 내 것만이 아니라 전부
abwi list --assignee "someone@example.com" # 특정 사용자의 것
abwi list --wiql @query.wiql               # WIQL로 자유롭게
```

댓글을 읽고 씁니다(기본으로 Markdown으로 게시됩니다):

```bash
abwi comment add 123 "Reproduced on \`main\`; see #456."
abwi comment add 123 @-        # 댓글 본문을 표준 입력에서
abwi comment list 123
```

work item끼리 연결하거나 연결을 해제합니다(`--type`에는 `parent`, `child`,
`related`(기본값) 또는 `System.LinkTypes.*` 참조 이름을 그대로 지정할 수
있습니다):

```bash
abwi link 123 456 --type parent   # #456을 #123의 부모로 만든다
abwi unlink 123 456               # 링크가 여러 개면 --type으로 좁힌다
```

최종 설정을 값마다 출처(`flag`, `env`, `local`, `global`, `default`)와 함께
표시합니다:

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

`abwi config <key>`는 값만, `--json`은 전체를 JSON으로 출력합니다.

### `--format html` 폴백

조직이 아직 Markdown work item을 지원하지 않거나 되돌릴 수 없는 전환을 피하고
싶다면, `--format html`을 넘기거나 설정에 `format = "html"`을 쓰세요. 이
모드에서도 여러분이 쓰는 것은 여전히 Markdown입니다 — 전송 전에 `abwi`가
[goldmark](https://github.com/yuin/goldmark)로 HTML로 변환하며, 필드는 HTML
포맷으로 유지됩니다:

```bash
abwi create -T Bug -t "Crash when saving a draft" --format html -d @description.md
```

## 라이선스

MIT © [lohn](https://github.com/lohn)
