# Git Diff 驅動程式 (Git Diff Driver)

[English](../GIT-DIFF-DRIVER.md) | [繁體中文](GIT-DIFF-DRIVER.md)

oasdiff 可以作為 [Git 外部 diff 驅動程式（git external diff driver）](https://git-scm.com/docs/git#Documentation/git.txt-codeGITEXTERNALDIFFcode) 執行，讓 `git log --patch`、`git diff` 與 `git show` 直接在行內輸出人類易讀的 OpenAPI 變更日誌，而非原始未加工的 YAML 純文字 diff。

靈感來自 [Jamie Tanna 關於將 oasdiff 作為 git diff 驅動程式的文章](https://www.jvt.me/posts/2026/04/11/oasdiff-driver/)。該文中介紹的舊版 Bash 包裝腳本已不再需要。

## 設定方式

在您的 Git 儲存庫中執行以下兩行一次性設定指令：

```bash
git config diff.oasdiff.command "oasdiff git-diff-driver"
echo "openapi.yaml diff=oasdiff" >> .gitattributes
```

將 `openapi.yaml` 替換為實際的規格路徑（多規格專案可使用萬用字元如 `**/openapi.yaml`）。將 `.gitattributes` 的變更 Commit 簽入，以便團隊成員共用。

## 使用方式

只要在 Git 指令中傳入 `--ext-diff` 旗標，驅動程式就會在任何輸出 diff 的 Git 指令中生效。以下是三種常見情境：

```bash
# 瀏覽規格的完整歷史變更
git log --patch --ext-diff -- openapi.yaml

# 檢查單一 Commit
git show --ext-diff <commit>

# 檢視即將 Commit 的變更（工作目錄 vs HEAD）
git diff --ext-diff -- openapi.yaml
```

在上述每種情況下，原本的 YAML 差異區塊（Hunks）都會被替換為 `oasdiff changelog` 所輸出的人類易讀變更日誌。以 `git log` 為例：

```
commit aed51e7b25195d82c3edd76caab75c4e1a2a2922
Author: Jane <jane@example.com>
Date:   Thu May 28 15:14:56 2026 +0300

    v2

1 changes: 0 error, 0 warning, 1 info
info  [response-non-success-status-added] at openapi.yaml
        in API GET /pets
                added the non-success response with the status `404`


commit 7dd4bdb866153dcd3e7cbc47d41953f05ff671c6
Author: Jane <jane@example.com>
Date:   Thu May 28 15:14:56 2026 +0300

    v1

Added openapi.yaml
```

若未傳入 `--ext-diff`，Git 會回退至預設的純文字 diff。若希望在日常工作流程中將 `--ext-diff` 設為預設值，可在 Git 設定中配置 `diff.external`（請參閱 Git 官方文件）。

## 處理的情境

| 情境 | 輸出內容 |
|---|---|
| 兩版本皆存在（一般修改） | 兩個 Blob 版本之間的完整變更日誌 |
| 檔案新增（Root Commit 或新檔案） | `Added <path>` |
| 檔案刪除 | `Removed <path>` |
| 僅權限模式變更（chmod） | 空白 — Git 本身機制即會印出 mode 差異 |
| 載入錯誤（格式損壞的規格等） | 在行內直接呈現錯誤訊息；diff 管線會繼續處理其他檔案 |

## 為什麼使用 `cat-file` 而非 Git 的暫存檔

`git-diff-driver` 會忽略 Git 傳遞的暫存檔路徑（`/tmp/.git-blob-xxxxx`），而是直接透過 `git cat-file blob <hex>` 重新讀取每個 Blob。輸出中的來源標籤會是 Blob 的短 SHA 加上儲存庫內的真實路徑（例如 `abc1234:openapi.yaml`），而非無意義的臨時檔案路徑。

## 限制

- `--ext-diff` 旗標是 Git 本身的設計限制——驅動程式無法自動強制啟用。

## 延伸閱讀

- [從 Git 修訂版本載入規格](GIT-REVISION.md) — oasdiff 通用支援的 `<ref>:<path>` 語法（包含 Blob Hash）。
- [破壞性變更與變更日誌](BREAKING-CHANGES.md) — Git Diff 驅動程式底層所執行的指令與詳細輸出說明。
