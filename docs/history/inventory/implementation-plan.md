# Inventory Feature — Implementation Plan

## 技術架構

### 專案整合方式

遵循現有 oasdiff 架構模式：
- CLI 層：`internal/inventory.go` — Cobra 子命令定義
- 核心邏輯：`inventory/` package — 獨立於 CLI 的業務邏輯
- 設定解析：`inventory/config.go` — Config 結構與載入
- 盤點表模型：`inventory/model.go` — 資料結構定義
- 輸出格式：`inventory/writer.go` — CSV / Excel 寫入

### 新增依賴

| 套件 | 用途 |
|------|------|
| `github.com/xuri/excelize/v2` | Excel (.xlsx) 讀寫 |

CSV 使用 Go 標準庫 `encoding/csv`。

### 目錄結構

```
oasdiff/
├── inventory/
│   ├── config.go          # Config 結構、YAML/JSON 載入
│   ├── config_test.go
│   ├── model.go           # InventoryRecord, InventorySheet 資料模型
│   ├── model_test.go
│   ├── generate.go        # generate 核心邏輯
│   ├── generate_test.go
│   ├── review.go          # review 核心邏輯
│   ├── review_test.go
│   ├── diff.go            # diff 核心邏輯
│   ├── diff_test.go
│   ├── apply.go           # apply 核心邏輯
│   ├── apply_test.go
│   ├── merge.go           # merge 核心邏輯
│   ├── merge_test.go
│   ├── reader.go          # CSV/Excel 讀取（解析既有盤點表）
│   ├── reader_test.go
│   ├── writer.go          # CSV/Excel 寫入
│   ├── writer_test.go
│   ├── extractor.go       # 從 OpenAPI spec 擷取欄位值
│   ├── extractor_test.go
│   ├── serial.go          # 流水號管理邏輯
│   └── serial_test.go
├── internal/
│   ├── inventory.go       # Cobra 子命令註冊 (generate/review/diff/apply/merge)
│   └── inventory_flags.go # inventory 相關 flags 定義
├── data/
│   └── inventory/         # 測試用 fixture 檔案
│       ├── config.yaml
│       ├── sample-spec.yaml
│       └── sample-inventory.csv
```

## 資料模型設計

### Config

```go
type Config struct {
    Prefix  string         `yaml:"prefix" json:"prefix"`
    Output  string         `yaml:"output" json:"output"`   // "csv" | "excel"
    Columns []ColumnConfig `yaml:"columns" json:"columns"`
}

type ColumnConfig struct {
    Source string `yaml:"source" json:"source"` // OpenAPI 欄位名或 x- 屬性名
    Header string `yaml:"header" json:"header"` // 盤點表顯示的欄位標題
    Type   string `yaml:"type" json:"type"`     // 可選：boolean, string, array 等
}
```

### InventoryRecord

```go
type InventoryRecord struct {
    Serial     string            // 編號 (e.g. "API-001")
    Date       string            // YYYY-MM-DD
    ChangeType string            // 新增 / 修改 / 刪除
    Method     string            // HTTP method (identity key part 1)
    Path       string            // API path (identity key part 2)
    Values     map[string]string // config 定義的各欄位值
}
```

### InventorySheet

```go
type InventorySheet struct {
    Config     *Config
    Records    []InventoryRecord
    MaxSerial  int               // 目前最大流水號（含已刪除）
}
```

## 各子命令流程

### generate

1. 載入 config
2. 載入 OpenAPI spec（支援 glob/composed mode）
3. 遍歷所有 operations，依 config 欄位擷取值
4. 分配流水號 (prefix + 001 起)
5. 所有 record 的 ChangeType = "新增"，Date = today
6. 寫出 CSV 或 Excel

### review

1. 載入 config
2. 載入 OpenAPI spec
3. 統計各欄位資訊（總 API 數、各 tag 數量等）
4. 特別統計自定義 x- 屬性（例：x-transaction=true 共 N 支 API）
5. 輸出摘要報告至 stdout

### diff

1. 載入 config
2. 載入 OpenAPI spec
3. 載入前一份盤點 CSV/Excel
4. 以 method+path 比對：
   - 盤點表有、spec 沒有 → 標記待刪除
   - spec 有、盤點表沒有 → 標記待新增
   - 兩者都有但欄位值不同 → 標記待修改
5. 輸出差異清單至 stdout

### apply

1. 執行 diff 邏輯取得差異
2. 讀取既有盤點表
3. 套用變更：
   - 新增：MaxSerial + 1，ChangeType = "新增"，Date = today
   - 修改：更新欄位值，ChangeType = "修改"，Date = today
   - 刪除：保留 record，ChangeType = "刪除"，Date = today
4. 寫出更新後的盤點表

### merge

1. 載入基礎盤點表 A
2. 載入待合併盤點表 B
3. 以 method+path 比對：
   - B 有 A 沒有 → 新增至 A（分配新流水號）
   - 兩者都有但內容不同 → 以更新時間新者為準，標記修改
   - A 有 B 沒有 → 保持不動（不刪除，因為來源不同）
4. 不可更動 A 的既有編號
5. 寫出合併結果

## CLI 介面設計

```
oasdiff inventory generate --config <path> <openapi-spec> [--output csv|excel] [-o output-file]
oasdiff inventory review   --config <path> <openapi-spec>
oasdiff inventory diff     --config <path> <openapi-spec> --inventory <csv/excel>
oasdiff inventory apply    --config <path> <openapi-spec> --inventory <csv/excel> [-o output-file]
oasdiff inventory merge    --base <csv/excel> --patch <csv/excel> [-o output-file]
```

## 錯誤處理

- Config 缺少必要欄位 → 明確錯誤訊息
- OpenAPI spec 載入失敗 → 沿用現有 load package 錯誤處理
- CSV/Excel 格式不符 → 提示欄位對不上 config
- requestBody/responses 欄位在 CSV 模式下 → 警告並跳過
- 流水號衝突 → 不應發生（以 MaxSerial 管理），若發生則 fatal error
