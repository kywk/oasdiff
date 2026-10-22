# 從 Git 修訂版本載入規格 (Loading Specs from Git Revisions)

[English](../GIT-REVISION.md) | [繁體中文](GIT-REVISION.md)

oasdiff 可以直接使用標準的 `<ref>:<path>` 語法從任何 Git 修訂版本（Revision）載入 OpenAPI 規格——無需額外旗標、無需暫存檔案，也無需手動執行 `git show`。

## 語法

```
<git-ref>:<path-to-spec>
```

範例：

| 表達式 | 意涵 |
|---|---|
| `HEAD:openapi.yaml` | 當前 Commit 中的 `openapi.yaml` 規格 |
| `HEAD~1:openapi.yaml` | 當前 HEAD 前一個 Commit 中的規格 |
| `origin/main:openapi.yaml` | 遠端 `main` 分支上的 `openapi.yaml` 規格 |
| `v2.3.0:api/openapi.yaml` | 標籤（Tag）`v2.3.0` 上的 `api/openapi.yaml` 規格 |
| `abc1234:openapi.yaml` | 特定 Commit SHA 中的規格 |
| `e69de29:openapi.yaml` | 特定 Blob SHA 中的規格（Git 外部 diff 協定傳遞的形式 — 請參閱 [Git Diff 驅動程式](GIT-DIFF-DRIVER.md)） |

## 使用方式

比較當前工作目錄的規格與 `origin/main` 上的版本：

```bash
oasdiff breaking origin/main:openapi.yaml openapi.yaml
```

跨兩個標籤進行比較：

```bash
oasdiff changelog v1.0.0:openapi.yaml v2.0.0:openapi.yaml
```

比較昨天的 `HEAD` 與今天的版本：

```bash
oasdiff diff HEAD~1:openapi.yaml HEAD:openapi.yaml
```

## 運作原理

oasdiff 會偵測 `<ref>:<path>` 格式，並在內部執行 `git show <ref>:<path>` 以擷取規格內容。規格中的相對路徑 `$ref` 會依據規格的路徑進行解析，行為與從本機檔案載入完全一致。

該指令必須在包含該規格的 Git 儲存庫目錄中執行。

## 抓取遺失的 Commit (`--fetch`)

oasdiff 預設僅從本機物件資料庫中讀取 Git 修訂版本；它預設不會修改您的儲存庫。若您引用的 Commit 不在本地 Clone 中（常見於審查他人的 PR Commit，或在淺層 Clone (Shallow Clone) 環境中），`git show` 會失敗，oasdiff 會印出抓取該 Commit 的精確指令：

```
git fetch origin <ref>
```

傳入 `--fetch` 旗標可讓 oasdiff 在讀取規格前自動針對 `origin` 遠端執行 fetch：

```bash
oasdiff changelog --fetch <base-sha>:openapi.yaml <revision-sha>:openapi.yaml
```

`--fetch` 會將遺失的 Git 物件下載至您的本機儲存庫（不會移動任何分支或 Ref）。由於這會變更儲存庫狀態，因此屬於 Opt-in（需明確傳入啟用）。只有其 Commit 確實不存在時才會觸發 fetch；已存在的 Commit 會直接讀取。

## GitHub Actions 整合

Git 修訂版本語法在 CI/CD 中特別有用。以下工作流程可檢測 Base 分支與 PR 分支之間的破壞性變更，無需額外的 checkout 步驟或暫存檔：

```yaml
name: API breaking changes

on:
  pull_request:

jobs:
  oasdiff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0          # Git ref 需要完整歷史紀錄

      - uses: oasdiff/oasdiff-action/breaking@v0
        with:
          base: 'origin/${{ github.base_ref }}:openapi.yaml'
          revision: 'HEAD:openapi.yaml'
```

> **附註：** 必須設定 `fetch-depth: 0`。`actions/checkout` 預設使用的淺層 Clone 不包含 Git 修訂版本語法所依賴的歷史紀錄或遠端 Ref。

## 延伸閱讀

- [Git Diff 驅動程式](GIT-DIFF-DRIVER.md) — 將 oasdiff 設定為 Git 外部 diff 驅動程式，讓 `git log --patch` 與 `git diff` 直接在行內輸出人類易讀的 OpenAPI 變更日誌。
