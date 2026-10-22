# 規格驗證 (Validate)

[English](../VALIDATE.md) | [繁體中文](VALIDATE.md)

`oasdiff validate <spec>` 用於檢查單一 OpenAPI 規格是否違反 RFC 規範：無效的 `type` 型別值、缺少必填欄位、格式錯誤的路徑、不合法的正規表示式、無法解析的 `$ref`，以及類似的結構性問題。它會根據 OpenAPI 與 JSON Schema 規範進行規格檢驗。

此指令不是可自訂風格的 Linter（例如 Spectral）；所有發現的違規項目皆為規範定義的違規，外加少數 oasdiff 原生提供的檢驗規則——針對在 OpenAPI 語法上合法但存在矛盾或未明確定義的情境（例如重複的 enum 值、模糊不清的參數序列化，以及在必填參數或屬性上宣告了預設值，而該預設值在語意上永遠不會生效）。每一項發現皆被分類為特定的嚴重度層級（error、warning 或 info）。

這與比較兩份規格的 `breaking` / `changelog` 不同；`validate` 著眼於單一規格，負責回答「這是否為合法的 OpenAPI 文件？」。

## 使用方式

```bash
oasdiff validate openapi.yaml
```

規格來源可以是本機檔案路徑、URL、Git ref（例如 `main:openapi.yaml`，請參閱 [Git 修訂版本](GIT-REVISION.md)），或使用 `-` 從標準輸入（stdin）讀取：

```bash
oasdiff validate https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml
cat openapi.yaml | oasdiff validate -
```

## 輸出結果

預設輸出為人類易讀的文字：包含一行統計摘要，其後列出每個違規項目區塊，每個項目皆附有穩定的規則 ID 與 `file:line:column` 檔案行號位置（當載入器能追蹤來源位置時）：

```
2 findings: 2 error, 0 warning, 0 info
error	[duplicate-operation-id] at openapi.yaml:42:7
	duplicate operation id "listPets"
error	[security-scheme-type-invalid] at openapi.yaml:88:9
	...
```

使用 `-f yaml` 或 `-f json` 可產生結構化輸出。每項發現包含 `id`、`text`、`level`、`section`、`source`（`file` / `line` / `column`）以及穩定的 `fingerprint` 指紋。

```bash
oasdiff validate -f json openapi.yaml
```

在 CI 流程中，使用 `-f githubactions` 會為每項違規產生 GitHub Actions 註解（錨定至其 file/line/column），讓違規直接以行內方式顯示於 Pull Request 的 Files Changed 標籤頁中，並將 `error_count` / `warning_count` / `info_count` 輸出為步驟輸出（Step outputs）。

所有違規項目會在單次執行中全數回報（Multi-error），而非僅在遇到第一個錯誤時中斷。

## 嚴重度層級 (Severities)

每項發現皆源自 OpenAPI / JSON Schema 規範，並依據影響程度進行分類：

- **error（錯誤）** — 規格無法被可靠解析與使用：缺少必填欄位、無法解析的 `$ref`、無效型別、格式錯誤的路徑，以及類似的結構性損壞。此外也包括 `duplicate-operation-id`，因為非唯一的 operationId 會使程式碼生成工具失效。
- **warning（警告）** — 結構上合法但存在實質風險：在舊版文件中使用了 3.1 專用欄位、被靜默忽略的 `$ref` 同級欄位、路徑衝突、重複參數、與 Schema 不相符的 `default` 預設值、重複的 enum 列舉值（JSON Schema 規範 enum 項目應為唯一；重複通常代表複製貼上錯誤），以及將結構化型別與純量型別混合的參數型別聯集（例如 `type: [array, integer]` 缺乏明確的序列化標準：伺服器無法辨識 `?token=5` 究竟代表 `["5"]` 還是 `5`）。
- **info（資訊）** — 純資訊性質：例如 `example` 範例值與 Schema 定義不相符（API 契約本身依然有效）。

`--fail-on` 決定何種嚴重度會導致指令失敗（請參閱結束代碼）。目前分類為固定設定；未來可能會支援依規則自訂。

## 參數旗標 (Flags)

| 參數旗標 | 預設值 | 說明 |
|---|---|---|
| `-f, --format` | `text` | 輸出格式：`text`、`yaml`、`json` 或 `githubactions` |
| `-o, --fail-on` | `ERR` | 當違規項目的嚴重度大於或等於此等級時以代碼 1 結束：`ERR`、`WARN` 或 `INFO` |
| `--color` | `auto` | 何時為文字輸出上色：`auto`、`always`、`never` |
| `--allow-external-refs` | `true` | 是否解析外部 `$ref`；驗證不可信任的規格時可設為 `false` 以防範 SSRF |

## 結束代碼 (Exit codes)

| 代碼 | 意涵 |
|---|---|
| `0` | 沒有達到或超過 `--fail-on` 嚴重度的違規項目 |
| `1` | 至少存在一項達到或超過 `--fail-on` 嚴重度的違規項目 |
| `102` | 無法載入規格檔案 |

這使其非常適合作為 CI 的品質關卡：預設情況下 `oasdiff validate openapi.yaml` 在有任何 error 時即會中斷步驟（warning 與 info 仍會印出但不會使指令失敗）。使用 `--fail-on WARN` 亦會在有警告時失敗，或使用 `--fail-on INFO` 在有任何發現時皆失敗。

## 規則識別碼 (Rule IDs)

每項違規皆有穩定且以烤肉串命名法（kebab-case）命名的規則 ID，例如 `path-parameters-mismatch`、`duplicate-operation-id`、`unresolved-ref`、`schema-pattern-regex-invalid` 與 `security-scheme-type-invalid`。尚未指派專屬 ID 的違規項目則會歸類在統括性的 `spec-validation-error` 之下。

## 意見反饋

發現問題？請[在此建立 Issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[validate]%20)，標題請加上 `[validate]`。
