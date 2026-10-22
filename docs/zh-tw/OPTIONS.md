# 為端點新增 OPTIONS 方法 (Add OPTIONS Method for CORS Preflight)

[English](../OPTIONS.md) | [繁體中文](OPTIONS.md)

`oasdiff options`（亦支援別名 `add-options`）子指令用於檢查 OpenAPI 規格中的端點，並自動為缺少 `OPTIONS` 的端點補上標準的 `OPTIONS` 方法與 `200`/`204` 回應。

這解決了 API Gateway 與 APIM 平台常見的整合痛點：Spring Boot 等後端框架預設在 Filter / Servlet 層級處理跨來源（CORS）Preflight 預檢請求，Controller 中並無對應的 Handler 映射，導致自動產出的 OpenAPI 規格中缺乏 `OPTIONS` 定義。當規格匯入嚴格依賴路徑定義的 APIM 平台時，瀏覽器的 CORS 預檢請求常因平台判定無此端點而遭攔截。

## 使用方式

```bash
# 輸出修改後的規格至標準輸出（可重新導向至檔案或 pipe 處理）
oasdiff options openapi.yaml > openapi-with-options.yaml

# 直接儲存至指定新檔案
oasdiff options openapi.yaml -o openapi-with-options.yaml

# 原地覆寫原檔案（僅限本機檔案）
oasdiff options openapi.yaml -i

# 為所有尚未定義 OPTIONS 的端點皆補上 OPTIONS（不限 HTTP 方法）
oasdiff options openapi.yaml --all

# 僅針對特定 HTTP 方法的端點補上 OPTIONS
oasdiff options openapi.yaml --methods POST,PUT,DELETE,PATCH

# 輸出為 JSON 格式
oasdiff options openapi.yaml -f json -o openapi-with-options.json
```

## 預設行為

預設情況下，`oasdiff options` 會：
1. 掃描規格中的所有路徑（Paths）。
2. 檢查該端點是否已定義 `OPTIONS` 方法；若已有定義則保持原狀、不作修改。
3. 檢查端點是否包含具狀態變更特性的 HTTP 方法（`POST`、`PUT`、`DELETE`、`PATCH`）。
4. 若符合條件，則注入標準的 `OPTIONS` 操作：
   - `summary`: "CORS preflight"
   - `description`: "CORS preflight response for APIM"
   - `responses`:
     - `200`: "CORS preflight successful"
     - `204`: "No Content"

## 參數旗標 (Flags)

| 參數旗標 | 預設值 | 說明 |
|---|---|---|
| `-o, --output` | (stdout) | 輸出檔案路徑 |
| `-i, --in-place` | `false` | 原地覆寫輸入檔案（僅限本機檔案） |
| `-m, --methods` | `POST,PUT,DELETE,PATCH` | 觸發補上 OPTIONS 的 HTTP 方法清單（逗號分隔） |
| `--all` | `false` | 為所有尚未定義 OPTIONS 的端點補上 OPTIONS（忽略方法限制） |
| `-f, --format` | `yaml` | 輸出格式：`yaml` 或 `json` |
| `--summary` | `CORS preflight` | 產生的 OPTIONS 操作之 summary 摘要文字 |
| `--description` | `CORS preflight response for APIM` | 產生的 OPTIONS 操作之 description 說明文字 |
| `--status-codes` | `200,204` | 產生的回應 HTTP 狀態碼清單 |
| `--allow-external-refs` | `true` | 是否允許規格中的外部 `$ref` |
