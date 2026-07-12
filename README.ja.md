# abwi

> [English](README.md) | **日本語** | [한국어](README.ko.md)

Azure Boards の work item を読み書きするための CLI です。
Markdown を第一級でサポートしており、`az boards` では難しかったことが `abwi` では標準の動作になります。

## なぜ必要か

- Azure Boards の大きめのテキストフィールドは**ネイティブの Markdown** に対応していますが、`az devops` からは書き込めません。
  フォーマットの指定には JSON Patch に `multilineFieldsFormat` という追加エントリが必要で、Azure CLI はこれを送ってくれないためです。
  `abwi` はこのエントリを自動で付与するので、Description・Repro Steps・Acceptance Criteria といったフィールドが、
  HTML エスケープされた文字列ではなく本物の Markdown として保存されます。
- 複数行フィールドはどれも `-f <refname>=<value>` の形でまとめて指定できます。
  値は curl 流に、`@file` でファイルを、`@-` で標準入力を読み込み、
  先頭の `\@` はリテラルの `@` のエスケープになります。
  フィールドごとの専用フラグを覚える必要はありません。

> [!WARNING]
> フィールドの Markdown 化は **work item 単位で不可逆**です。
> 一度 Markdown で保存したフィールドを HTML に戻すことはできません。
> これは Azure DevOps 側の仕様であり、`abwi` の制限ではありません。
> 組織やプロセスがまだ Markdown に移行できない場合は、
> [`--format html` フォールバック](#--format-html-フォールバック)を使ってください。

## インストール

```bash
go install github.com/lohn/abwi/cmd/abwi@latest
```

## 認証

**Entra ID（既定）。**
Azure CLI で一度サインインしておけば、`abwi` はその資格情報をそのまま使います：

```bash
az login
```

**PAT（明示的なオプトイン・非推奨）。**
**グローバル**設定ファイルに `auth = "pat"` を書くか `--auth pat` を渡したうえで、
トークンを環境変数 `ABWI_PAT` に設定します
（未設定なら、Azure CLI と同じ `AZURE_DEVOPS_EXT_PAT` を参照します）：

```bash
export ABWI_PAT=...          # または AZURE_DEVOPS_EXT_PAT
abwi --auth pat show 123
```

PAT を読み取るのは環境変数からだけで、設定ファイルには書けない作りです。
そのため、トークンをうっかりコミットしてしまう事故は起きません。

また、`auth` キーは**グローバル設定ファイルでのみ**有効です。
チェックアウトしたリポジトリの側から認証方式を切り替えられるべきではないため、
リポジトリローカルの `.abwi.toml` に書かれた `auth` は警告を出したうえで無視されます。

## 設定

`abwi` は 2 つの TOML ファイルの設定をマージして使います：

- **ローカル**：`.abwi.toml`。カレントディレクトリから上へ向かって探索します
  （リポジトリルートに置く想定です）
- **グローバル**：ユーザーごとの設定ディレクトリにある設定ファイル。
  Linux では `~/.config/abwi/config.toml` です
  （Go の `os.UserConfigDir` に従うため、macOS では
  `~/Library/Application Support/abwi/config.toml`、
  Windows では `%AppData%\abwi\config.toml` になります）

優先順位は高い順に、
**フラグ > 環境変数（`ABWI_ORG`・`ABWI_PROJECT`）> ローカル > グローバル** です。

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

設定可能なキー:

| キー           | 内容                                                       | 既定値     |
| -------------- | ---------------------------------------------------------- | ---------- |
| `org`          | 組織 URL（`https://dev.azure.com/<org>`）                  | —          |
| `project`      | プロジェクト名                                             | —          |
| `format`       | 大きめテキストのフォーマット：`markdown` / `html`          | `markdown` |
| `auth`         | 認証方式：`entra` / `pat`。**グローバル設定専用**          | `entra`    |
| `default-type` | `create` で `--type` を省略したときに使う work item タイプ | —          |
| `[aliases]`    | `-f` 用の短縮名テーブル。完全な参照名に展開される          | —          |

解決後の値とその出どころは `abwi config` で確認できます。
詳しくは下の[使い方](#使い方)を見てください。

## 使い方

work item を作成します
（`--type` を省略すると設定の `default-type` が使われます。
`-d` と `-f` の値は `@file`・`@-` に対応しています）：

```bash
abwi create -T Bug -t "Crash when saving a draft" \
  -d @description.md \
  -f Microsoft.VSTS.TCM.ReproSteps=@repro.md

# 同じ操作を、設定の [aliases] 短縮名で
abwi create -T Bug -t "Crash when saving a draft" -d @description.md -f repro=@repro.md
```

既存の work item を更新します
（`@-` で標準入力から Markdown を読み込みます。
値を `@` そのもので始めたいときは `\@` でエスケープします）：

```bash
generate-criteria | abwi update 123 -s Active -f ac=@- \
  -f System.Description='\@mentions start with an escaped at-sign'
```

work item を表示します（`--json` で生のレスポンスを出力）：

```bash
abwi show 123
```

work item を一覧します。
既定では自分にアサインされたものを、更新が新しい順に表示します。
フラグで絞り込むほか、WIQL クエリで丸ごと差し替えることもできます：

```bash
abwi list -T Bug -s Active --limit 20
abwi list --all                            # 自分のぶんだけでなく全部
abwi list --assignee "someone@example.com" # 特定の人のもの
abwi list --wiql @query.wiql               # WIQL で自由に
```

コメントを読み書きします（既定で Markdown として投稿されます）：

```bash
abwi comment add 123 "Reproduced on \`main\`; see #456."
abwi comment add 123 @-        # コメント本文を標準入力から
abwi comment list 123
```

work item 同士をリンクします。リンクの削除もできます
（`--type` には `parent`・`child`・`related`（既定）のほか、
`System.LinkTypes.*` の参照名をそのまま指定できます）：

```bash
abwi link 123 456 --type parent   # #456 を #123 の親にする
abwi unlink 123 456               # リンクが複数あるときは --type で絞り込む
```

解決後の設定を、値ごとの出どころ
（`flag`・`env`・`local`・`global`・`default`）付きで表示します：

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

`abwi config <key>` は値だけを、`--json` は全体を JSON で出力します。

### `--format html` フォールバック

組織がまだ Markdown の work item に対応していない場合や、
不可逆な切り替えを避けたい場合は、
`--format html` を渡すか設定に `format = "html"` を書いてください。
このモードでも、書くのはこれまでどおり Markdown のままです。
送信前に `abwi` が [goldmark](https://github.com/yuin/goldmark) で HTML へ変換するので、
フィールドは HTML フォーマットのまま保たれます：

```bash
abwi create -T Bug -t "Crash when saving a draft" --format html -d @description.md
```

## ライセンス

MIT © [lohn](https://github.com/lohn)
