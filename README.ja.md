# abwi

> [English](README.md) | **日本語** | [한국어](README.ko.md)

Azure Boards の work item を読み書きする CLI です。Markdown をファーストクラスで
サポートし、`az boards` では難しいことを `abwi` はデフォルトにします。

## なぜ必要か

- Azure Boards は大きなテキストフィールドで**ネイティブ Markdown** をサポート
  していますが、`az devops` はそれを書き込めません。フィールドのフォーマットを
  指定するには JSON Patch ドキュメントに追加の `multilineFieldsFormat` エントリが
  必要で、Azure CLI はこれを一切送信しないためです。`abwi` はこのエントリを
  自動的に追加するため、Description、Repro Steps、Acceptance Criteria などの
  フィールドが HTML エスケープされたテキストではなく、本物の Markdown として
  保存されます。
- すべての複数行フィールドは `-f <refname>=<value>` で**汎用的に**指定できます。
  入力は curl スタイルです。`@file` はファイルを、`@-` は標準入力を読み込み、
  先頭の `\@` はリテラルの `@` をエスケープします。フィールドごとのフラグを
  覚える必要はありません。

> [!WARNING]
> フィールドの Markdown への切り替えは **work item ごとに不可逆**です。ある
> work item のフィールドに一度 Markdown が保存されると、Azure DevOps はそれを
> HTML に戻すことを許可しません。これは Azure DevOps の制限であり、`abwi` の
> 制限ではありません。組織やプロセスがまだ Markdown を受け入れる準備ができて
> いない場合は、[`--format html` フォールバック](#--format-html-フォールバック)を
> 使用してください。

## インストール

```bash
go install github.com/lohn/abwi/cmd/abwi@latest
```

## 認証

**Entra ID（デフォルト）。** Azure CLI で一度サインインすれば、`abwi` が
Azure CLI の資格情報を自動的に利用します：

```bash
az login
```

**PAT（明示的なオプトイン、非推奨）。** 設定ファイルで `auth = "pat"` を設定し
（または `--auth pat` を渡し）、トークンを `ABWI_PAT` としてエクスポートします
（未設定の場合は、Azure CLI が使う変数 `AZURE_DEVOPS_EXT_PAT` にフォールバック
します）：

```bash
export ABWI_PAT=...          # または AZURE_DEVOPS_EXT_PAT
abwi --auth pat show 123
```

PAT の値は環境変数**のみ**から読み取られ、設定ファイルから読み取られることは
決してありません。そのため、トークンが誤ってコミットされてしまうことはありません。

## 設定

`abwi` は 2 つの TOML ファイルから設定をマージします：

- **ローカル**: `.abwi.toml`。カレントディレクトリから上方向に探索して
  見つけます（リポジトリルートに置いてください）
- **グローバル**: ユーザーごとの設定ディレクトリ — Linux では
  `~/.config/abwi/config.toml`（Go の `os.UserConfigDir` に従うため、macOS では
  `~/Library/Application Support/abwi/config.toml`、Windows では
  `%AppData%\abwi\config.toml`）

優先順位（高い順）：**フラグ > 環境変数（`ABWI_ORG`、`ABWI_PROJECT`）>
ローカルファイル > グローバルファイル**。

```toml
# .abwi.toml — リポジトリルートにコミットする
org = "https://dev.azure.com/myorg"
project = "MyProject"
default-type = "Product Backlog Item"

# -f 用の短縮名。完全なフィールド参照名に展開される
[aliases]
ac = "Microsoft.VSTS.Common.AcceptanceCriteria"
repro = "Microsoft.VSTS.TCM.ReproSteps"
```

利用可能なキー：`org`、`project`、`format`（`markdown`/`html`、デフォルト
`markdown`）、`auth`（`entra`/`pat`、デフォルト `entra`）、`default-type`、
および `[aliases]` テーブル。

`abwi config` を実行すると、解決された値と各値の出所を確認できます —
下記の[使い方](#使い方)を参照してください。

## 使い方

work item を作成します（`--type` は設定の `default-type` にフォールバック
します。`-d`/`-f` の値は `@file` と `@-` をサポートします）：

```bash
abwi create -T Bug -t "Crash when saving a draft" \
  -d @description.md \
  -f Microsoft.VSTS.TCM.ReproSteps=@repro.md

# 同じ操作を、設定の [aliases] 短縮名を使って
abwi create -T Bug -t "Crash when saving a draft" -d @description.md -f repro=@repro.md
```

既存の work item のフィールドを更新します（`@-` で標準入力から Markdown を
読み込みます。値をリテラルの `@` で始める必要がある場合は `\@` を使います）：

```bash
generate-criteria | abwi update 123 -s Active -f ac=@- \
  -f System.Description='\@mentions start with an escaped at-sign'
```

work item を表示します（`--json` で生のレスポンスを表示）：

```bash
abwi show 123
```

work item を一覧表示します。デフォルトでは自分のものを、変更が新しい順に
表示します。フラグで絞り込むか、完全な WIQL クエリですべてを制御できます：

```bash
abwi list -T Bug -s Active --limit 20
abwi list --all                            # 自分のものだけでなく全員分
abwi list --assignee "someone@example.com" # 他の人のもの
abwi list --wiql @query.wiql               # 完全な制御
```

コメント（デフォルトで Markdown として投稿されます）：

```bash
abwi comment add 123 "Reproduced on \`main\`; see #456."
abwi comment add 123 @-        # コメント本文を標準入力から
abwi comment list 123
```

リンクとリンク解除（`--type`：`parent`、`child`、デフォルトの `related`、
または完全な `System.LinkTypes.*` 参照名）：

```bash
abwi link 123 456 --type parent   # #456 を #123 の親にする
abwi unlink 123 456               # 複数のリンクがある場合は --type で特定
```

解決された設定を表示します。各値には出所（`flag`、`env`、`local`、`global`、
`default`）が注釈として付きます：

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

`abwi config <key>` は単一の値を表示し、`--json` は全体を JSON として表示します。

### `--format html` フォールバック

組織がまだ Markdown の work item に対応していない場合（あるいは不可逆な
切り替えを避けたい場合）は、`--format html` を渡すか、設定で `format = "html"`
を設定してください。それでも Markdown を*書く*ことに変わりはありません —
`abwi` が送信前に [goldmark](https://github.com/yuin/goldmark) で HTML に変換し、
フィールドは HTML フォーマットのまま維持されます：

```bash
abwi create -T Bug -t "Crash when saving a draft" --format html -d @description.md
```

## ライセンス

MIT © [lohn](https://github.com/lohn)
