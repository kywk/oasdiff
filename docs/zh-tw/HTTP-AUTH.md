# 遠端規格的 HTTP 認證 (HTTP Authentication for Remote Specs)

[English](../HTTP-AUTH.md) | [繁體中文](HTTP-AUTH.md)

當從需要驗證身分的 HTTP(S) URL 載入 OpenAPI 規格時，您可以使用 `--header` 與 `--query` 旗標將認證憑證注入至 oasdiff 所發出的每個 HTTP 請求中——包含解析外部 `$ref` 所發出的請求。

## 使用方式

### HTTP 標頭 (Headers)

使用 `--header key=value` 新增一或多個 HTTP 標頭（可重複使用此旗標）：

```bash
# Bearer Token
oasdiff diff https://api.example.com/v1/openapi.yaml https://api.example.com/v2/openapi.yaml \
  --header "Authorization=Bearer eyJhbGciOi..."

# 多個標頭
oasdiff changelog https://api.example.com/base.yaml https://api.example.com/revision.yaml \
  --header "Authorization=Bearer token" \
  --header "X-API-Key=my-key"

# Basic Auth
oasdiff breaking https://api.example.com/v1.yaml https://api.example.com/v2.yaml \
  --header "Authorization=Basic dXNlcjpwYXNz"
```

### 查詢參數 (Query Parameters)

使用 `--query key=value` 新增查詢參數（可重複使用此旗標）：

```bash
# 將 API Key 作為查詢參數
oasdiff diff https://api.example.com/openapi.yaml?version=1 \
             https://api.example.com/openapi.yaml?version=2 \
  --query "api_key=secret123"

# 多個查詢參數
oasdiff changelog https://api.example.com/spec.yaml https://api.example.com/spec.yaml \
  --query "token=abc" \
  --query "format=full"
```

### 兩者結合使用

標頭與查詢參數可合併同時使用：

```bash
oasdiff diff https://api.example.com/v1/spec.yaml https://api.example.com/v2/spec.yaml \
  --header "Authorization=Bearer token" \
  --query "env=staging"
```

## 運作行為

- **所有 HTTP 請求**：標頭與查詢參數會被加入至每個 HTTP 請求中，包含透過 URL 引用其他檔案解析外部 `$ref` 時所發出的請求。
- **既有查詢參數**：規格 URL 中若已存在任何查詢參數皆會完整保留；`--query` 傳入的值會附加於其後。
- **鍵值格式**：請使用 `key=value` 格式。若數值中包含 `=` 能被正確處理（僅第一個 `=` 會被視為分隔符號）。
- **不包含 `=` 的項目**：會被靜默忽略。
- **非 HTTP 來源**：這些旗標對本機檔案或 Git 修訂版本來源無任何影響，僅在規格 URL 採用 `http` 或 `https` 協定時生效。

## 支援的指令

`--header` 與 `--query` 旗標適用於所有載入規格的指令：

- `oasdiff diff`
- `oasdiff changelog`
- `oasdiff breaking`
- `oasdiff summary`
- `oasdiff flatten`
- `oasdiff upgrade`
- `oasdiff validate`

## 設定檔支援

這些旗標亦可在 [oasdiff 設定檔](CONFIG-FILES.md) 中進行配置：

```yaml
header:
  - "Authorization=Bearer token"
  - "X-Custom-Header=value"
query:
  - "api_key=secret"
```

## 安全性考量

- 避免在共用環境（例如 CI 日誌）中直接於命令列傳入機密金鑰。建議改用環境變數展開或設定受權限保護的設定檔。
- 當設定 `--allow-external-refs=false` 時，無論是否配置了認證標頭，oasdiff 皆不會追蹤外部 `$ref`，藉此有效防範 SSRF 攻擊。
