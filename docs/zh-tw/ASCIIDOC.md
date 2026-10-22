# Asciidoc 輸出格式 (Asciidoc Output Format)

[English](../ASCIIDOC.md) | [繁體中文](ASCIIDOC.md)

oasdiff 的 `diff` 與 `changelog` 指令支援以 [Asciidoc](https://asciidoc.org/) 作為輸出格式。

## 使用方式

使用 `--format asciidoc`（或 `-f asciidoc`）旗標：

```bash
# 以 Asciidoc 輸出完整 Diff
oasdiff diff base.yaml revision.yaml -f asciidoc

# 以 Asciidoc 輸出變更日誌
oasdiff changelog base.yaml revision.yaml -f asciidoc

# 以 Asciidoc 輸出破壞性變更
oasdiff breaking base.yaml revision.yaml -f asciidoc

# 將輸出寫入檔案
oasdiff changelog base.yaml revision.yaml -f asciidoc > changelog.adoc
```

## 自訂範本 (Custom Templates)

您可以為變更日誌輸出提供自訂的 Asciidoc 範本：

```bash
oasdiff changelog base.yaml revision.yaml --template my-template.adoc -f asciidoc
```

範本採用 Go 的 `text/template` 語法，並能存取與 [Markdown 與 HTML 範本](CHANGELOG-TEMPLATE.md) 相同的資料與函式：

- `.GroupedChanges` — 依端點或區段分組的變更清單
- `.BaseVersion` / `.RevisionVersion` — 規格版本字串
- `.GetVersionTitle()` — 格式化後的版本比較字串（如 `1.0.0 vs. 2.0.0`）
- `pathGroups .GroupedChanges` — 依路徑與操作排序的 API 路徑變更清單
- `sectionGroups .GroupedChanges` — 依區段排序的安全性與元件變更清單
- `capitalize "string"` — 將字串首字轉為大寫

### 最小自訂範本範例

```asciidoc
= API Changelog {{ .GetVersionTitle }}
{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
== API Changes
{{ range . }}
=== {{ .Group.Operation }} {{ .Group.Path }}
{{ range .Changes }}* {{ if .IsBreaking }}*[BREAKING]* {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ end }}
{{ range sectionGroups .GroupedChanges }}
== {{ capitalize .Group.Section }}
{{ range .Changes }}* {{ if .IsBreaking }}*[BREAKING]* {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ else }}
No changes detected
{{ end }}
```

## 輸出結構

### Diff (`oasdiff diff -f asciidoc`)

Diff 輸出使用 Asciidoc 標題層級與巢狀無序清單：

```asciidoc
=== New Endpoints: 1
* GET /pets/{petId}

=== Deleted Endpoints: 1
* POST /pets

=== Modified Endpoints: 2
* PUT /pets
** Request body changed
*** Content changed
```

### 變更日誌 (`oasdiff changelog -f asciidoc`)

變更日誌輸出依端點分組，並標註破壞性變更：

```asciidoc
= API Changelog 1.0.0 vs. 2.0.0

== API Changes

=== GET /pets
* *[BREAKING]* removed the query parameter 'limit'
* added the query parameter 'page_size'

=== POST /pets
* request body became optional
```
