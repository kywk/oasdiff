[![CI](https://github.com/oasdiff/oasdiff/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/oasdiff/oasdiff/actions)
[![codecov](https://codecov.io/gh/oasdiff/oasdiff/branch/main/graph/badge.svg?token=Y8BM6X77JY)](https://codecov.io/gh/oasdiff/oasdiff)
[![Go Reference](https://pkg.go.dev/badge/github.com/oasdiff/oasdiff.svg)](https://pkg.go.dev/github.com/oasdiff/oasdiff)
[![Docker Image Version](https://img.shields.io/docker/v/tufin/oasdiff?sort=semver)](https://hub.docker.com/r/tufin/oasdiff/tags)

[English](README.md) | [繁體中文](README_zh-tw.md)

![oasdiff banner](https://github.com/yonatanmgr/oasdiff/assets/31913495/ac9b154e-72d1-4969-bc3b-f527bbe7751d)


用於比較與檢測 OpenAPI 規格中破壞性變更（Breaking Changes）的命令列工具。

可在本機執行、透過 [GitHub Action](https://github.com/oasdiff/oasdiff-action) 於 CI 流程中整合，或使用 [oasdiff.com](https://www.oasdiff.com) 託管的 PR 審查工作流程，透過 CI commit status 批准或拒絕各項變更。

## 30 秒快速上手

無需安裝——可直接使用 Docker 針對兩個範例規格進行測試：

```bash
docker run --rm -t tufin/oasdiff changelog \
  https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml \
  https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test5.yaml
```

上述指令會以人類易讀的格式輸出可能影響 API 使用者的變更日誌（包含破壞性與非破壞性變更）。將 `changelog` 替換為 `breaking` 可僅檢視會破壞既有 API 客戶端的變更；替換為 `diff` 則可輸出 API 定義的完整機器可讀 diff（包含純文件修改）。

## 安裝方式

### 使用 Go 安裝
```bash
go install github.com/oasdiff/oasdiff@latest
```

### macOS 使用 Homebrew 安裝
```bash
brew install oasdiff
```

### macOS 與 Linux 使用 curl 安裝

[install.sh](../install.sh) 指令碼會下載最新的 oasdiff 二進位檔至 `/usr/local/bin`：

```bash
curl -fsSL https://raw.githubusercontent.com/oasdiff/oasdiff/main/install.sh | sh
```

若需指定特定版本，可設定 `version` 環境變數：

```bash
curl -fsSL https://raw.githubusercontent.com/oasdiff/oasdiff/main/install.sh | version=1.11.7 sh
```

### 使用 asdf 安裝

https://github.com/oasdiff/asdf-oasdiff

### 下載二進位檔
macOS、Linux 與 Windows（皆支援 x86_64 與 arm64）的預建二進位檔可至 [Releases 頁面](https://github.com/oasdiff/oasdiff/releases/) 下載。

## 文件導覽

依您要執行的任務分門別類。初次接觸 oasdiff？建議從 **指令（Commands）** 開始。

### 指令（Commands）
最上層的子指令，分類方式與 `oasdiff --help` 相同。

**比較兩個規格**

- [`breaking`](zh-tw/BREAKING-CHANGES.md) — 僅檢測會破壞既有 API 客戶端的變更
- [`changelog`](zh-tw/BREAKING-CHANGES.md) — 以人類易讀格式呈現所有可能影響 API 使用者的變更（無論是否具破壞性）
- [`diff`](zh-tw/DIFF.md) — API 定義的完整差異比較，包含純文件修改（輸出格式：html、json、markdown、markup、text 或 yaml — 預設為 yaml）
- [`summary`](zh-tw/DIFF.md) — 兩個規格之間的高階變更摘要統計（基於 diff 引擎；共用相同選項）

**處理單一規格**

- [`validate`](zh-tw/VALIDATE.md) — 依 RFC 規範檢查規格是否違規（無效型別、缺少必填欄位、不合法的正規表示式、未解析的 `$ref`）
- [`upgrade`](zh-tw/OPENAPI-31.md#使用-oasdiff-upgrade-轉換規格) — 將 OpenAPI 3.0 規格正規化轉換為最新的 3.x
- [`flatten`](zh-tw/ALLOF.md) — 將 `allOf` 結構取代為合併後的等價 Schema
- [`options`](zh-tw/OPTIONS.md) — 為端點自動補上 OPTIONS 方法以支援 APIM 與瀏覽器 CORS 預檢請求

**Git 整合**

- [`breaking-files`](zh-tw/BREAKING-FILES.md) — 將變更的規格清單與指定 Git ref 中的版本逐一比較；適用於 [pre-commit hook](zh-tw/BREAKING-FILES.md#pre-commit-hook-整合)
- [`git-diff-driver`](zh-tw/GIT-DIFF-DRIVER.md) — 作為 Git 外部 diff 驅動程式，讓 `git log --patch` 直接以行內方式呈現 OpenAPI 變更日誌

**API 清單盤點**

- [`inventory`](zh-tw/inventory.md) — 定期審計的 API 清單管理（產生、審查、比對、套用及合併盤點表）

**參考資訊**

- [`checks changelog`](zh-tw/CHECKS.md) — 列出 `breaking` 與 `changelog` 用於分類變更的規則（[自訂規則](CUSTOMIZING-CHECKS.md)）
- [`checks validate`](zh-tw/CHECKS.md#validate-檢查規則) — 列出 `validate` 報告的驗證規則
- [`checks changelog coverage`](zh-tw/CHECKS.md#覆蓋率對應表-coverage-map) — 列出 OpenAPI 文件的所有可能編輯操作及其對應的檢查規則
- [`schema`](zh-tw/BREAKING-CHANGES.md#json-schema) — 輸出 `breaking`/`changelog` 的 JSON 輸出結構 Schema

### 輸入來源（Inputs）
規格檔案的來源方式。

- [Git 修訂版本](zh-tw/GIT-REVISION.md) — 與分支、標籤（tag）或 Commit 進行比較
- 本機檔案、http/s URL、YAML 或 JSON — 皆能透明解析與處理

### 規格比較（Comparison）
oasdiff 如何配對基準版本（Base）與修訂版本（Revision），以及何者視為差異。

- [端點比對](zh-tw/MATCHING-ENDPOINTS.md) — 處理重新命名的路徑參數與[重複端點](zh-tw/MATCHING-ENDPOINTS.md#重複端點-duplicate-endpoints)
- [可為 Null 性變更（Nullability）](zh-tw/NULLABILITY.md) — 三種等價的 nullable 形式及何時變更會構成破壞性變更
- [跨多檔案 API 比較（Composed Mode）](zh-tw/COMPOSED.md) — 例如 API Gateway 每個微服務擁有獨立規格
- [檢查多個獨立規格](zh-tw/BREAKING-FILES.md) — 針對每個規格分別進行獨立比較，而非像 Composed 模式合併處理
- [篩選端點](zh-tw/FILTERING-ENDPOINTS.md) — 將 diff 範圍縮小至特定端點子集

### 正規化（Normalization）
在比對前對齊規格，使等價內容相互對應。

- [合併 `allOf` Schemas](zh-tw/ALLOF.md)
- [合併共用（路徑層級）參數](zh-tw/COMMON-PARAMS.md)
- [路徑前綴修改](zh-tw/PATH-PREFIX.md) — 移除或新增前綴以比對路徑變更後的 API
- [不區分大小寫的標頭比較](zh-tw/HEADER-DIFF.md) — 將 `Content-Type` 與 `content-type` 視為相同標頭

### API 生命週期（API Lifecycle）
在版本演進中傳達設計意圖。

- [API 與參數廢棄標記](zh-tw/DEPRECATION.md)
- [API 穩定度等級](zh-tw/STABILITY.md)（draft / alpha / beta / stable）
- [版本號遞增檢查](zh-tw/VERSIONING.md) — 檢查在未遞增主版本號（Major Version）時釋出的破壞性變更

### 變更篩選（Filtering Changes）
選擇要報告哪些類型的差異。

- [排除特定類型的變更](zh-tw/DIFF.md#排除特定類型的變更) — 例如忽略純描述（description）修改
- [追蹤 OpenAPI Extensions 擴充欄位](zh-tw/DIFF.md#openapi-extensions-擴充欄位) — 啟用回報 `x-*` 欄位的變更
- [排除特定的擴充欄位名稱](zh-tw/DIFF.md#排除特定名稱的擴充欄位) — 在追蹤 Extensions 時跳過不感興趣的欄位

### 輸出自訂（Output）
調整與豐富產出的報告內容。

- [自訂 HTML 與 Markdown 變更日誌範本](zh-tw/CHANGELOG-TEMPLATE.md)
- [在變更日誌項目中加入 OpenAPI-extension 屬性](zh-tw/ATTRIBUTES.md)
- [原始檔位置追蹤（Source Locator）](zh-tw/SOURCE-LOCATOR.md)
- [變更指紋（Change Fingerprints）](zh-tw/FINGERPRINT.md) — 跨 Commit 的穩定變更識別碼
- [錯誤報告與結束代碼](zh-tw/ERRORS.md)
- 多國語言在地化：en, ru, pt-br, es

### 執行方式（How to run）
- [Docker](zh-tw/DOCKER.md)
- [設定檔（Configuration file）](zh-tw/CONFIG-FILES.md)
- [嵌入至 Go 程式](GO.md)
- 用於 CI 的 [GitHub Action](https://github.com/oasdiff/oasdiff-action) — 以及適用於團隊的 [oasdiff.com](https://www.oasdiff.com)，支援每個變更的 PR 留言審批與 commit-status 檢查
- [MCP 伺服器](zh-tw/MCP.md) — 透過位於 `https://api.oasdiff.com/mcp` 的託管伺服器，自 AI 助手（Claude、Cursor 等）呼叫 oasdiff

### 參考文件（Reference）
- [OpenAPI 3.1 與 3.2 支援](zh-tw/OPENAPI-31.md) — 支援狀況與功能
- [安全性：控制外部 `$ref` 載入以防範 SSRF](zh-tw/SECURITY.md)
- [使用範例與指引](zh-tw/USAGE_EXAMPLES.md) — 常見情境操作配方
- [參與貢獻](CONTRIB.md)

## 演示

<img src="./demo.svg">

## 鳴謝
本專案仰賴優秀的 Go 語言 OpenAPI 3.0 與 3.1 實作套件：[kin-openapi](https://github.com/getkin/kin-openapi)。

## 意見反饋
歡迎提供回饋！  
若您對改進有任何建議或對 API 相關功能有其他需求，歡迎[提出討論](https://github.com/oasdiff/oasdiff/discussions/new?category=ideas)。
