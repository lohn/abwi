# abwi

> [English](README.md) | [日本語](README.ja.md) | **한국어**

Azure Boards work item을 읽고 쓰는 CLI입니다. Markdown을 일급으로 지원하여,
`az boards`에서는 어려운 일을 `abwi`는 기본값으로 만듭니다.

## 왜 필요한가

- Azure Boards는 큰 텍스트 필드에서 **네이티브 Markdown**을 지원하지만,
  `az devops`는 이를 쓸 수 없습니다. 필드의 포맷을 지정하려면 JSON Patch 문서에
  `multilineFieldsFormat` 항목을 추가로 넣어야 하는데, Azure CLI는 이를 전혀
  보내지 않기 때문입니다. `abwi`는 이 항목을 자동으로 추가하므로 Description,
  Repro Steps, Acceptance Criteria 같은 필드가 HTML로 이스케이프된 텍스트가
  아닌 진짜 Markdown으로 저장됩니다.
- 모든 여러 줄 필드는 `-f <refname>=<value>`로 **범용적으로** 지정합니다.
  입력은 curl 스타일입니다. `@file`은 파일을 읽고, `@-`는 표준 입력을 읽으며,
  맨 앞의 `\@`는 리터럴 `@`를 이스케이프합니다. 필드별 플래그를 외울 필요가
  없습니다.

> [!WARNING]
> 필드를 Markdown으로 전환하는 것은 **work item 단위로 되돌릴 수 없습니다**.
> 특정 work item의 필드에 한 번 Markdown이 저장되면, Azure DevOps는 이를 다시
> HTML로 변환하는 것을 허용하지 않습니다. 이는 Azure DevOps의 제한이지
> `abwi`의 제한이 아닙니다. 조직이나 프로세스가 아직 Markdown을 받아들일
> 준비가 되지 않았다면 [`--format html` 폴백](#--format-html-폴백)을 사용하세요.

## 설치

```bash
go install github.com/lohn/abwi/cmd/abwi@latest
```

## 인증

**Entra ID(기본값).** Azure CLI로 한 번 로그인하면 `abwi`가 Azure CLI 자격
증명을 자동으로 사용합니다:

```bash
az login
```

**PAT(명시적 옵트인, 권장하지 않음).** 설정 파일에서 `auth = "pat"`을
설정하고(또는 `--auth pat`을 전달) 토큰을 `ABWI_PAT`으로 내보내세요(설정되어
있지 않으면 Azure CLI가 사용하는 변수인 `AZURE_DEVOPS_EXT_PAT`로 대체됩니다):

```bash
export ABWI_PAT=...          # 또는 AZURE_DEVOPS_EXT_PAT
abwi --auth pat show 123
```

PAT 값은 **오직** 환경 변수에서만 읽고 설정 파일에서는 절대 읽지 않으므로,
토큰이 실수로 커밋되는 일은 없습니다.

## 설정

`abwi`는 두 개의 TOML 파일에서 설정을 병합합니다:

- **로컬**: `.abwi.toml`. 현재 디렉터리에서 위로 올라가며 찾습니다(저장소
  루트에 두세요)
- **글로벌**: 사용자별 설정 디렉터리 — Linux에서는
  `~/.config/abwi/config.toml`(Go의 `os.UserConfigDir`를 따르므로 macOS에서는
  `~/Library/Application Support/abwi/config.toml`, Windows에서는
  `%AppData%\abwi\config.toml`)

우선순위(높은 것부터): **플래그 > 환경 변수(`ABWI_ORG`, `ABWI_PROJECT`) >
로컬 파일 > 글로벌 파일**.

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

사용 가능한 키: `org`, `project`, `format`(`markdown`/`html`, 기본값
`markdown`), `auth`(`entra`/`pat`, 기본값 `entra`), `default-type`, 그리고
`[aliases]` 테이블.

`abwi config`를 실행하면 해석된 값과 각 값의 출처를 확인할 수 있습니다.
아래 [사용법](#사용법)을 참조하세요.

## 사용법

work item을 생성합니다(`--type`은 설정의 `default-type`으로 대체됩니다.
`-d`/`-f` 값은 `@file`과 `@-`를 지원합니다):

```bash
abwi create -T Bug -t "Crash when saving a draft" \
  -d @description.md \
  -f Microsoft.VSTS.TCM.ReproSteps=@repro.md

# 같은 작업을 설정의 [aliases] 축약 이름으로
abwi create -T Bug -t "Crash when saving a draft" -d @description.md -f repro=@repro.md
```

기존 work item의 필드를 업데이트합니다(`@-`로 표준 입력에서 Markdown을
읽습니다. 값이 리터럴 `@`로 시작해야 하면 `\@`를 사용하세요):

```bash
generate-criteria | abwi update 123 -s Active -f ac=@- \
  -f System.Description='\@mentions start with an escaped at-sign'
```

work item을 표시합니다(`--json`으로 원시 응답 표시):

```bash
abwi show 123
```

work item을 나열합니다. 기본적으로 내 것이 최근 변경 순으로 표시됩니다.
플래그로 필터링하거나 전체 WIQL 쿼리로 직접 제어할 수 있습니다:

```bash
abwi list -T Bug -s Active --limit 20
abwi list --all                            # 내 것만이 아니라 모두의 것
abwi list --assignee "someone@example.com" # 다른 사람의 것
abwi list --wiql @query.wiql               # 완전한 제어
```

댓글(기본적으로 Markdown으로 게시됩니다):

```bash
abwi comment add 123 "Reproduced on \`main\`; see #456."
abwi comment add 123 @-        # 댓글 본문을 표준 입력에서
abwi comment list 123
```

링크 및 링크 해제(`--type`: `parent`, `child`, 기본값인 `related`, 또는 전체
`System.LinkTypes.*` 참조 이름):

```bash
abwi link 123 456 --type parent   # #456을 #123의 부모로 만들기
abwi unlink 123 456               # 링크가 여러 개면 --type으로 구분
```

해석된 설정을 표시합니다. 각 값에는 출처(`flag`, `env`, `local`, `global`,
`default`)가 주석으로 표시됩니다:

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

`abwi config <key>`는 단일 값을 출력하고, `--json`은 전체를 JSON으로
출력합니다.

### `--format html` 폴백

조직이 아직 Markdown work item을 지원하지 않는다면(또는 되돌릴 수 없는 전환을
피하고 싶다면) `--format html`을 전달하거나 설정에서 `format = "html"`을
지정하세요. 여전히 Markdown을 *쓴다*는 점은 그대로입니다. `abwi`가 전송 전에
[goldmark](https://github.com/yuin/goldmark)로 HTML로 변환하며, 필드는 HTML
포맷으로 유지됩니다:

```bash
abwi create -T Bug -t "Crash when saving a draft" --format html -d @description.md
```

## 라이선스

MIT © [lohn](https://github.com/lohn)
