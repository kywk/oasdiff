# 使用自訂範本輸出 OpenAPI 變更日誌 (OpenAPI Changelog with Custom Template)

[English](../CHANGELOG-TEMPLATE.md) | [繁體中文](CHANGELOG-TEMPLATE.md)

當使用 `markdown` 或 `html` 格式時，您可以透過提供自訂範本檔案來自訂變更日誌的輸出排版：

```bash
oasdiff changelog base.yaml revision.yaml --template my-template.md -f markdown
oasdiff changelog base.yaml revision.yaml --template my-template.html -f html
```

## 範本資料 (Template Data)

範本中可以存取以下變數欄位與函式：

### 資料欄位 (Data Fields)

| 欄位 | 說明 |
|---|---|
| `.GroupedChanges` | 依端點分組（針對 API 變更）或依區段分組（針對安全性 / components 變更）的變更集合 |
| `.BaseVersion` | 基準規格版本字串 |
| `.RevisionVersion` | 修訂規格版本字串 |
| `.GetVersionTitle()` | 格式化後的版本比較字串，例如 `1.0.0 vs. 2.0.0` |

### 範本函式 (Template Functions)

| 函式 | 說明 |
|---|---|
| `pathGroups .GroupedChanges` | 回傳依路徑（path）與操作（operation）排序後的 API 路徑變更清單 |
| `sectionGroups .GroupedChanges` | 回傳依區段名稱排序後的安全性與元件變更清單 |
| `capitalize "string"` | 將字串的第一個字母轉換為大寫 |

由 `pathGroups` 與 `sectionGroups` 回傳的每個分組項目包含：
- `.Group.Path` — API 路徑（例如 `/pets`）
- `.Group.Operation` — HTTP 方法（例如 `GET`）
- `.Group.Section` — 非路徑分組的區段名稱（例如 `security`、`components`）
- `.Changes` — 變更項目切片（Slice），每個項目包含 `.Text`、`.IsBreaking` 與 `.Comment`

## 範本範例

完整範例範本可參閱 [`examples/custom-changelog-template.md`](../../examples/custom-changelog-template.md)。

以下為最小可行範例：

```markdown
# API Changelog {{ .GetVersionTitle }}

{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
## API Changes
{{ range . }}
### {{ .Group.Operation }} {{ .Group.Path }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning: **BREAKING**: {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ end }}

{{ range sectionGroups .GroupedChanges }}
## {{ capitalize .Group.Section }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning: **BREAKING**: {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ else }}
No changes
{{ end }}
```

## 附註

- `pathGroups` 與 `sectionGroups` 會回傳已排序的項目，確保多次執行時輸出順序完全一致。
- 若沒有任何路徑變更，`API Changes` 區段將會被完全省略。
- 每個區段分組（例如 Security、Components）僅在包含至少一項變更時才會顯示。
- 使用 `{{ if .GroupedChanges }}` 可在變更日誌為空時顯示「No changes」等提示訊息。
