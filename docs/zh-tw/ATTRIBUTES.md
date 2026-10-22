# 在變更日誌中加入 OpenAPI 擴充欄位 (Add OpenAPI Extensions to the Changelog)

[English](../ATTRIBUTES.md) | [繁體中文](ATTRIBUTES.md)

OpenAPI 規格中可以包含自訂的 `x-*` 擴充欄位，為各操作附加中繼資料（Metadata）。例如：
```yaml
/restapi/oauth/token:
  post:
    operationId: getToken
    x-audience: Public
    summary: ...
    requestBody:
        ...
    responses:
        ...
```

使用 `--attributes` 旗標可在 JSON 或 YAML 格式的變更日誌項目中包含這些數值：

```bash
❯ oasdiff changelog base.yaml revision.yaml -f yaml --attributes x-audience
- id: new-optional-request-property
  text: added the new optional request property ivr_pin
  level: 1
  operation: POST
  operationId: getToken
  path: /restapi/oauth/token
  source: new-revision.yaml
  section: paths
  attributes:
    x-audience: Public
```
