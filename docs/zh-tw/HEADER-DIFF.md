# 不區分大小寫的標頭比較 (Case-Insensitive Header Comparison)

[English](../HEADER-DIFF.md) | [繁體中文](HEADER-DIFF.md)

oasdiff 預設以不區分大小寫的方式比較標頭名稱，這符合 HTTP 網路傳輸規則中「標頭名稱不區分大小寫」的標準（RFC 7230, RFC 9110）。`Content-Type` 與 `content-type` 指的是同一個標頭，在 diff 中會被視為相同標頭處理。此行為適用於請求標頭參數（`in: header`）、回應標頭（response headers），以及 `apiKey` 安全架構中所傳送的標頭（`in: header`）。查詢參數（Query parameter）與 Cookie 名稱則維持區分大小寫。

若需恢復區分大小寫的比對（符合 OpenAPI 通用規則「規範中的所有欄位名稱皆區分大小寫」），可設定 `--case-insensitive-headers=false`：

```bash
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers=false
```
