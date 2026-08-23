# Inventory Feature — Implementation Tasks

## Phase 0: 基礎建設

- [ ] T0.1 建立 `inventory/` package 目錄結構
- [ ] T0.2 新增 `github.com/xuri/excelize/v2` 依賴
- [ ] T0.3 建立測試用 fixture 檔案 (`data/inventory/`)

## Phase 1: Config 與資料模型

- [ ] T1.1 實作 `inventory/config.go` — Config 結構定義、YAML/JSON 載入、驗證
- [ ] T1.2 實作 `inventory/model.go` — InventoryRecord、InventorySheet 結構
- [ ] T1.3 實作 `inventory/serial.go` — 流水號產生與解析邏輯
- [ ] T1.4 撰寫 config / model / serial 單元測試

## Phase 2: OpenAPI 欄位擷取

- [ ] T2.1 實作 `inventory/extractor.go` — 從 kin-openapi T 結構擷取欄位值
- [ ] T2.2 支援標準欄位：method, path, operationId, summary, description, tags, parameters
- [ ] T2.3 支援自定義 x- 屬性擷取與型別判斷
- [ ] T2.4 支援 requestBody schema / responses 欄位（Excel 專用，結構序列化）
- [ ] T2.5 撰寫 extractor 單元測試

## Phase 3: 讀寫盤點表

- [ ] T3.1 實作 `inventory/writer.go` — CSV 寫入
- [ ] T3.2 實作 `inventory/writer.go` — Excel 寫入
- [ ] T3.3 實作 `inventory/reader.go` — CSV 讀取（解析回 InventorySheet）
- [ ] T3.4 實作 `inventory/reader.go` — Excel 讀取
- [ ] T3.5 撰寫 reader / writer 單元測試（round-trip 驗證）

## Phase 4: generate 子命令

- [ ] T4.1 實作 `inventory/generate.go` — 核心邏輯
- [ ] T4.2 實作 `internal/inventory.go` — Cobra 命令註冊（先做 generate）
- [ ] T4.3 在 `internal/run.go` 註冊 inventory 子命令群組
- [ ] T4.4 撰寫 generate 整合測試
- [ ] T4.5 手動測試：使用範例 spec 產出 CSV/Excel 確認格式正確

## Phase 5: review 子命令

- [ ] T5.1 實作 `inventory/review.go` — 摘要統計邏輯
- [ ] T5.2 整合 Cobra 命令
- [ ] T5.3 撰寫 review 單元測試

## Phase 6: diff 子命令

- [ ] T6.1 實作 `inventory/diff.go` — 比對邏輯（新增/修改/刪除偵測）
- [ ] T6.2 整合 Cobra 命令
- [ ] T6.3 撰寫 diff 單元測試（含邊界案例：空盤點表、全新 API、全刪除）

## Phase 7: apply 子命令

- [ ] T7.1 實作 `inventory/apply.go` — 套用變更邏輯
- [ ] T7.2 驗證流水號規則：刪除不重用、新增接續最大號
- [ ] T7.3 驗證刪除 API 不可復活規則
- [ ] T7.4 整合 Cobra 命令
- [ ] T7.5 撰寫 apply 單元測試

## Phase 8: merge 子命令

- [ ] T8.1 實作 `inventory/merge.go` — 合併邏輯
- [ ] T8.2 驗證基礎文件編號不被異動
- [ ] T8.3 驗證衝突解決（以更新時間為準）
- [ ] T8.4 整合 Cobra 命令
- [ ] T8.5 撰寫 merge 單元測試

## Phase 9: CLI 整合、文件與收尾

- [ ] T9.1 確認所有子命令 help text 完整（Long description + Examples）
- [ ] T9.2 支援 composed mode（glob 多檔合併）
- [ ] T9.3 補充 `--output` flag 在 generate/apply/merge 中的行為
- [ ] T9.4 端對端測試：完整流程 generate → review → diff → apply → merge
- [ ] T9.5 新增 `docs/inventory.md` 使用說明文件（含 config 範例、各子命令用法與範例輸出）
- [ ] T9.6 更新專案 README.md 加入 inventory 功能簡介與文件連結

## 預估工時

| Phase | 預估 |
|-------|------|
| Phase 0 | 0.5h |
| Phase 1 | 2h |
| Phase 2 | 3h |
| Phase 3 | 3h |
| Phase 4 | 2h |
| Phase 5 | 1.5h |
| Phase 6 | 2h |
| Phase 7 | 2.5h |
| Phase 8 | 3h |
| Phase 9 | 2h |
| **Total** | **~21.5h** |

## 開發順序建議

Phase 0 → 1 → 2 → 3 → 4（可交付第一個 milestone：generate）
→ 5 → 6 → 7（第二個 milestone：完整 single-spec 工作流）
→ 8 → 9（第三個 milestone：多源合併與收尾）
