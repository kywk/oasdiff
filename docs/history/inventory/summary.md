# Inventory Feature — 需求摘要

## 背景

公司依規定需定期盤點 API 清冊。新增 `oasdiff inventory` 子命令群組，提供 API 盤點表的產製、檢視、差異比對、更新及合併功能。

## 子命令一覽

| 子命令 | 用途 |
|--------|------|
| `generate` | 依 config + OpenAPI spec 產出盤點 CSV/Excel |
| `review` | 依 config + OpenAPI spec 摘要 API 資訊（含自定義屬性統計） |
| `diff` | 比對 OpenAPI spec 與前一份盤點表的差異 |
| `apply` | 將差異套用至盤點表（新增/修改/刪除） |
| `merge` | 合併兩份盤點表（保留基礎表編號） |

## Config 設定檔

格式：YAML（同時支援 JSON）

範例結構：
```yaml
prefix: "API"              # 流水號前綴 → API-001, API-002...
output: "csv"              # csv | excel
columns:
  - source: "method"
    header: "HTTP方法"
  - source: "path"
    header: "API路徑"
  - source: "operationId"
    header: "操作ID"
  - source: "summary"
    header: "摘要說明"
  - source: "description"
    header: "詳細描述"
  - source: "tags"
    header: "標籤"
  - source: "parameters"
    header: "參數"
  - source: "requestBody"     # 僅 Excel 支援
    header: "請求主體"
  - source: "responses"       # 僅 Excel 支援
    header: "回應"
  - source: "x-transaction"
    header: "交易類別"
    type: "boolean"
```

## 固定欄位（前三欄）

| 欄位 | 說明 |
|------|------|
| 編號 | 自定義前綴 + 流水號（三碼，破千自動增長）。例：API-001 |
| 日期 | 產製/異動時間 YYYY-MM-DD |
| 異動 | 新增 / 修改 / 刪除 |

## 核心規則

### API 識別鍵
- 以 `method + path` 作為唯一識別鍵

### 異動判斷
- 僅比對 config 中定義的欄位，有任一欄位值變更即視為「修改」

### 流水號規則
- 三碼起始（001），破千自動增長為四碼
- 已刪除的 API 編號永不重用
- 新增 API 接續最大流水號 +1
- 即使相同 endpoint 重新出現，仍需分配新編號

### 刪除處理
- 刪除的 API 保留在盤點表中
- 第二欄更新為刪除日期
- 第三欄標註「刪除」
- API 不可起死回生

### 輸出格式
- CSV：支援 method / path / operationId / summary / description / tags / parameters / 自定義 x- 屬性
- Excel：額外支援 requestBody schema、responses（因複雜結構不適合 CSV）

### Merge 規則
- 以 endpoint (method + path) 比對差異
- 遵循 apply 相同的新增/修改/刪除規則
- 不可異動基礎文件的既有編號
- 資訊衝突時以更新時間最新者為準

### Composed Mode
- 支援多個 OpenAPI 檔案 glob 合併（沿用專案既有功能）
