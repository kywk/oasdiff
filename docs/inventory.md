# oasdiff inventory

API 清冊盤點工具，提供 API 盤點表的產製、檢視、差異比對、更新及合併功能。

## 概述

`oasdiff inventory` 子命令群組用於定期盤點 API 清冊。透過 OpenAPI 規格文件與設定檔，自動化產出及維護 API 盤點表（CSV / Excel 格式）。

## 子命令

| 子命令 | 用途 |
|--------|------|
| `generate` | 從 OpenAPI spec 產出全新盤點表 |
| `review` | 摘要統計 API 資訊（含自定義屬性分組） |
| `diff` | 比對 OpenAPI spec 與既有盤點表的差異 |
| `apply` | 將差異套用至盤點表 |
| `merge` | 合併兩份盤點表 |

## Config 設定檔

設定檔支援 YAML 與 JSON 格式。

### 範例 (YAML)

```yaml
prefix: "API"        # 流水號前綴
output: "csv"        # 預設輸出格式: csv | excel
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

### 欄位說明

**固定欄位（前三欄，自動產生）：**

| 欄位 | 說明 |
|------|------|
| 編號 | 前綴 + 流水號（三碼零填充，破千自動增長）例：API-001 |
| 日期 | YYYY-MM-DD 格式的產製/異動日期 |
| 異動 | 新增 / 修改 / 刪除 |

**可設定欄位 (source)：**

| source | 說明 |
|--------|------|
| `method` | HTTP 方法 (GET, POST, PUT, DELETE 等) |
| `path` | API 路徑 |
| `operationId` | Operation ID |
| `summary` | 操作摘要 |
| `description` | 操作描述 |
| `tags` | 標籤（逗號分隔） |
| `parameters` | 參數列表，格式：`name(in)*`（*表示必填） |
| `requestBody` | 請求主體 schema（僅 Excel） |
| `responses` | 回應 schema（僅 Excel） |
| `x-*` | 自定義 OpenAPI Extension |

**type 屬性（可選）：** `boolean`, `string`, `array`。影響 x- 屬性值的格式化方式。

## 使用範例

### generate — 產出盤點表

```bash
# 輸出 CSV 到 stdout
oasdiff inventory generate --inventory-config config.yaml openapi.yaml

# 輸出到檔案
oasdiff inventory generate --inventory-config config.yaml openapi.yaml -o inventory.csv

# 輸出 Excel
oasdiff inventory generate --inventory-config config.yaml openapi.yaml --format excel -o inventory.xlsx

# Composed mode（多檔合併）
oasdiff inventory generate --inventory-config config.yaml --composed "specs/*.yaml" -o inventory.csv
```

**輸出範例：**

```
編號,日期,異動,HTTP方法,API路徑,操作ID,摘要說明,標籤,交易類別
API-001,2024-01-15,新增,GET,/users,listUsers,List all users,users,false
API-002,2024-01-15,新增,POST,/users,createUser,Create a user,users,true
API-003,2024-01-15,新增,GET,/orders,listOrders,List orders,orders,false
```

### review — 摘要檢視

```bash
oasdiff inventory review --inventory-config config.yaml openapi.yaml
```

**輸出範例：**

```
=== API Inventory Review ===

Total APIs: 6

--- By Tag ---
  admin: 1 APIs
  orders: 2 APIs
  users: 4 APIs

--- By Method ---
  DELETE: 1 APIs
  GET: 3 APIs
  POST: 2 APIs

--- 交易類別 (x-transaction) ---
  x-transaction=false: 3 APIs
    - GET /users
    - GET /orders
    - GET /users/{id}
  x-transaction=true: 3 APIs
    - POST /users
    - POST /orders
    - DELETE /users/{id}
```

### diff — 差異比對

```bash
oasdiff inventory diff --inventory-config config.yaml --inventory inventory.csv openapi.yaml
```

**輸出範例：**

```
=== Inventory Diff ===

--- Added (1) ---
  + GET /products

--- Modified (1) ---
  ~ API-001 GET /users
      摘要說明: "List all users" → "List all users (paginated)"

--- Deleted (2) ---
  - API-003 GET /orders
  - API-004 POST /orders
```

### apply — 套用變更

```bash
# 更新到新檔案
oasdiff inventory apply --inventory-config config.yaml --inventory inventory.csv openapi.yaml -o inventory-updated.csv

# 直接覆寫原檔
oasdiff inventory apply --inventory-config config.yaml --inventory inventory.csv openapi.yaml
```

### merge — 合併盤點表

```bash
oasdiff inventory merge --inventory-config config.yaml --base main-inventory.csv --patch team-inventory.csv -o merged.csv
```

## 核心規則

### API 識別鍵
以 **method + path** 作為 API 的唯一識別。

### 流水號規則
- 三碼零填充（001-999），破千自動增長
- 已刪除的 API 編號**永不重用**
- 新增 API 接續當前最大流水號 + 1

### 異動判斷
僅比對 config 中定義的欄位值，任一欄位有變更即視為「修改」。

### 刪除處理
- 已刪除的 API **保留在盤點表中**
- 日期更新為刪除日期
- 異動欄標記「刪除」
- **API 不可起死回生**：即使相同 endpoint 重新出現，也視為全新 API 並分配新編號

### Merge 規則
- 基礎文件的編號**永不異動**
- 新 API 使用基礎文件的下一個流水號
- 資訊衝突時以**更新時間最新者**為準
- 遵循相同的「不可起死回生」規則
