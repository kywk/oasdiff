# 共用參數 (Common Parameters)

[English](../COMMON-PARAMS.md) | [繁體中文](COMMON-PARAMS.md)

## 共用參數定義

被某個路徑下所有操作共用的參數，可以定義在路徑（Path）層級而非各個操作（Operation）層級。  
路徑層級的參數會被該路徑下的所有操作繼承。  
典型的使用情境是透過路徑參數存取特定資源的 GET/PUT/PATCH/DELETE 操作。

## Diff 與共用參數

oasdiff 的 `diff` 子指令會分別回報兩個層級的參數變更：路徑層級與操作層級。  
但請注意，每個層級是分開檢查的。  
例如，若將某個參數從路徑層級移動至操作層級，它會被回報為一次刪除與一次新增：

params_in_path.yaml:
```yaml
paths:
  '/admin/v0/abc/{id}':
    parameters:
      - $ref: '#/components/parameters/tenant_id'
      - $ref: '#/components/parameters/id'
    get:
      ...
```

params_in_op.yaml:
```yaml
paths:
  '/admin/v0/abc/{id}':
    get:
      parameters:
        - $ref: '#/components/parameters/tenant_id'
        - $ref: '#/components/parameters/id'
      ...
```

命令列：
```bash
oasdiff diff data/common-params/params_in_path.yaml data/common-params/params_in_op.yaml --exclude-elements endpoints
```

輸出：
```yaml
paths:
    modified:
        /admin/v0/abc/{id}:
            operations:
                modified:
                    GET:
                        parameters:
                            added:
                                header:
                                    - tenant-id
                                path:
                                    - id
            parameters:
                deleted:
                    header:
                        - tenant-id
                    path:
                        - id
```

為了解決此限制，可以使用 `--flatten-params` 旗標，它會在執行比對前先將路徑層級的共用參數合併至各個操作層級中：
```bash
oasdiff diff data/common-params/params_in_path.yaml data/common-params/params_in_op.yaml --flatten-params
```
此時輸出將為空，表示未發現任何實質變更。

## 變更日誌（包含破壞性變更）與共用參數

`changelog` 與 `breaking` 子指令專注於操作層級的參數，預設會忽略多數路徑層級的參數變更。  
例如，修改路徑層級參數的 `maxItems`，一般情況下不會被回報：

```yaml
paths:
  /api/v1.0/groups:
    parameters:
      - in: query
        name: category
        schema:
          type: array
          items:
            type: string
            maxItems: 20
```

```yaml
paths:
  /api/v1.0/groups:
    parameters:
      - in: query
        name: category
        schema:
          type: array
          items:
            type: string
            maxItems: 10
```

但若加入 `--flatten-params` 旗標（在比對前將共用參數從路徑層級展平合併至操作層級），它就會被正確識別並回報為破壞性變更：

```bash
oasdiff changelog data/checker/common_request_parameter_max_items_updated_revision.yaml data/checker/common_request_parameter_max_items_updated_base.yaml --flatten-params
```

輸出：
```
1 changes: 1 error, 0 warning, 0 info
error	[request-parameter-max-items-decreased] at data/checker/common_request_parameter_max_items_updated_base.yaml
	in API POST /api/v1.0/groups
		for the `query` request parameter `category`, the maxItems was decreased from `20` to `10`
```

## 總結
建議使用 `--flatten-params` 旗標以提高共用參數比對的精確度。
