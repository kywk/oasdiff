# 與 Git Ref 比對多個規格檔案 (Checking Many Specs Against a Git Ref)

[English](../BREAKING-FILES.md) | [繁體中文](BREAKING-FILES.md)

`oasdiff breaking` 本身即可將單一規格與其在特定 Git ref 中的版本進行比較，因為兩者引數皆可傳入 Git 修訂版本：

```bash
oasdiff breaking origin/main:openapi.yaml openapi.yaml
```

而 `oasdiff breaking-files` 則能一次針對多個規格檔案進行上述操作，每個規格會依據自身的相對路徑自動推導出在 Base Ref 中的基準版本：

```bash
oasdiff breaking-files --base origin/main --fail-on ERR openapi.yaml users.yaml orders.yaml
```

兩者的核心差異在於「比較的次數」：`breaking` 指令始終只進行**單次**比較（無論引數是檔案、萬用字元 Glob 或 Git 修訂版本）。而 `breaking-files` 則是**每個規格獨立進行一次比較**並分別輸出結果，因此它能明確指名到底是哪一份規格發生了破壞性變更；只要其中任一份規格包含達到或超過 `--fail-on` 的變更，指令便會以非零代碼結束。

這非常適用於已經知道哪些檔案發生變更、且需要對所有檔案進行整體裁決的情境，例如 Pre-commit Hook 或由 `git diff --name-only` 驅動的 CI 任務。

## 運作行為

若某個規格在 Base Ref 中不存在，代表它是新建立的規格；由於全新 API 沒有前一個版本可被破壞，因此會直接跳過：

```
$ oasdiff breaking-files --base origin/main --fail-on ERR new-api.yaml openapi.yaml
=== new-api.yaml ===
new file, not in base ref, skipped
=== openapi.yaml ===
1 changes: 1 error, 0 warning, 0 info
error	[api-path-removed-without-deprecation] at openapi.yaml
	in API GET /pets
		api path removed without deprecation
```

傳入的每個引數必須是工作目錄中的規格檔案路徑。標準輸入（stdin）、URL 或 Git 修訂版本字串無法在特定 Ref 中查詢路徑，因此不被接受。若在 Base Ref 中不存在則不屬於錯誤：即上述的新增檔案情境。

比對結果會逐一印出，因此若指定了結構化 `--format` 或 `--template`，每個規格會各自獨立渲染一次，而非合併成單一文件。

## 與 Composed 模式的區別

[跨多檔案模式 (Composed Mode)](COMPOSED.md) 是將比對到 Base Glob 與 Revision Glob 的所有規格合併為單一比較，適用於「多個檔案共同描述同一套 API」的情境。而 `breaking-files` 保持各規格彼此獨立，適用於「每個檔案各自是一套獨立 API」的架構。因此本指令不支援 Composed 模式，如果您的單一 API 被切分成多個檔案，`breaking-files` 並非合適的選擇。

## Pre-commit hook 整合

oasdiff 提供基於 `breaking-files` 打造的 [pre-commit](https://pre-commit.com/) hook，讓破壞性變更在 Commit 前就被攔截，而不必等到 CI。將其加入您的 `.pre-commit-config.yaml`：

```yaml
repos:
  - repo: https://github.com/oasdiff/oasdiff
    rev: v1.30.0
    hooks:
      - id: oasdiff-breaking
```

`rev` 釘選了 hook 採用的 oasdiff 發行版本。執行 `pre-commit autoupdate` 可自動更新至最新標籤，無需手動追蹤發行。

Hook 的 `files` 正則表達式決定哪些暫存的規格會被傳遞給 oasdiff，預設會比對 `openapi.yaml`、`openapi.yml` 與 `openapi.json`。您可以依專案配置進行覆寫。請注意 `args` 會替換 hook 的預設引數而非附加，因此覆寫時請保留 `--base`：

```yaml
      - id: oasdiff-breaking
        files: (^|/)api/.*\.yaml$
        args: [--base, origin/main, --fail-on, WARN]
```

pre-commit 僅在 hook 失敗時才會顯示輸出，而這正是您需要報告的時刻：阻擋 Commit 的那次執行會明確告訴您是哪一份規格破壞了相容性以及破壞的原因。

完整的設定範例與註解請參閱 [`examples/.pre-commit-config.yaml`](../../examples/.pre-commit-config.yaml)。
