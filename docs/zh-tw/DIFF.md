# 規格差異比較 (Diff)

[English](../DIFF.md) | [繁體中文](DIFF.md)

`oasdiff diff` 指令用於顯示兩個 OpenAPI 規格之間的完整差異。  
輸出包含完整的詳細資訊，通常為 YAML 或 JSON 格式，但亦支援 text、markdown 與 html 格式。  
此指令常用於產生結構化的 diff 報告供其他自動化工具解析，同時也適合人工閱讀。

> **附註：** `summary`、`breaking` 與 `changelog` 皆基於同一個 diff 引擎建置。本文介紹的大多數概念與參數旗標也適用於這些指令，例外之處會特別說明。

## 預設比較行為

oasdiff 採用了若干預設啟用的設定（Opt-out defaults），讓 diff 結果更符合真實世界 API 比較的預期，而非字面上的位元組層級規格比對。若您的使用情境需要更嚴格的比對行為，皆可透過 `--<flag>=false` 來關閉。

| 參數旗標 | 預設值 | 功能說明 | 詳細資訊 |
|---|---|---|---|
| `--match-inline-refs` | `true` | 將 `anyOf` / `oneOf` 下驗證等價的行內（inline）Schema 與 `$ref` 子 Schema 視為同一分支。程式碼生成重構若將行內 enum 擷取為具名元件，不會報告變更。 | [比對行內與 `$ref` 子 Schema](#比對行內與-ref-子-schema) |
| `--case-insensitive-headers` | `true` | 不區分大小寫比對標頭名稱。依據 RFC 7230 / 9110，`Content-Type` 與 `content-type` 視為同一標頭。 | [HEADER-DIFF.md](HEADER-DIFF.md) |
| `--allow-external-refs` | `true` | 允許解析器在載入規格時解析外部 `$ref`。處理不可信任的規格時可將其停用以防範 SSRF。 | (CLI 說明) |

其他比較微調旗標（如 `--flatten-allof`、`--flatten-params`、`--include-path-params`、`--auto-upgrade` 等）皆預設關閉（Opt-in，預設值為 `false`），因為它們會轉換輸入內容或改變比對語意，並非所有規格都適用。

## 輸出格式

預設的 diff 輸出格式為 `yaml`。  
可透過 `--format` 旗標產生其他格式：
- yaml: 包含所有 diff 詳細資訊
- json: 包含所有 diff 詳細資訊
- text: 專為人類閱讀設計，僅提供 diff 中最重要的部分（與 markdown 相同）
- markdown: 專為人類閱讀設計，僅提供 diff 中最重要的部分（與 text 相同）
- html: 專為人類閱讀設計，僅提供 diff 中最重要的部分（另請參閱 [HTML 格式變更日誌](BREAKING-CHANGES.md#輸出格式)）

附註：
- 空白的 `yaml` 或 `json` 結果表示 diff 為空，換言之即沒有任何變更。  
- `json` 格式預設會排除 `endpoints` 區段，以避免 [複雜映射鍵問題](#複雜映射鍵-complex-mapping-keys)。

## 阻止變更進入（Preventing Changes）

`oasdiff diff` 的常見用法是作為 CI/CD 管線中的檢查步驟以偵測變更。  
為了阻止變更合併，可設定 `oasdiff diff` 在發現任何變更時回傳錯誤代碼。  
若要在發現任何變更時以代碼 1 結束，請加入 `--fail-on-diff` 旗標。

## 路徑 (Paths) 與端點 (Endpoints)

OpenAPI 規範具備層級式的 [Paths（路徑）](https://swagger.io/specification/#paths-object) 與 [Operations（操作/HTTP 方法）](https://swagger.io/specification/#operation-object) 模型。  
Oasdiff 遵循此階層架構，輸出階層式的路徑 diff：包含新增 (added)、刪除 (deleted) 與修改 (modified)；而在「modified」區段下，又包含另一組操作變更（新增、刪除、修改）。例如：
```yaml
paths:
    deleted:
        - /register
        - /subscribe
    modified:
        /api/{domain}/{project}/badges/security-score:
            operations:
                modified:
                    GET:
```
Oasdiff 同時也提供每個「端點（Endpoint，即 Path + Operation 的組合）」的替代簡化 diff，例如：
```yaml
endpoints:
    deleted:
        - method: POST
          path: /subscribe
        - method: POST
          path: /register
    modified:
        ?   method: GET
            path: /api/{domain}/{project}/badges/security-score
        :   tags:
                deleted:
                    - security
```

## 複雜映射鍵 (Complex Mapping Keys)

修改過的端點（modified endpoints）區段在每個 Key 中包含兩個項目：method 與 path，這在 YAML 中被稱為[複雜映射鍵（Complex mapping key）](https://stackoverflow.com/questions/33987316/what-is-a-complex-mapping-key-in-yaml)。  
部分 YAML 解析函式庫不支援複雜映射鍵，例如：
- Python PyYAML: 請參閱 https://github.com/oasdiff/oasdiff/issues/94#issuecomment-1087468450
- Golang gopkg.in/yaml.v3 無法還原序列化（unmarshal）oasdiff 的輸出。此套件提供了因應方案：https://github.com/tliron/yamlkeys

為克服此限制，oasdiff 允許您加入 `--exclude-elements=endpoints` 旗標以排除端點區段。  
使用 `json` 輸出格式時，oasdiff 會自動排除 `endpoints`。

## OpenAPI Extensions 擴充欄位

Oasdiff 預設會追蹤 [OpenAPI 擴充欄位（Extensions）](https://swagger.io/docs/specification/openapi-extensions/) 的變更。若要停用，請參閱 [排除特定類型的變更](#排除特定類型的變更)。  
OpenAPI 擴充欄位的 diff 格式符合 [JavaScript Object Notation (JSON) Patch (RFC 6902)](https://datatracker.ietf.org/doc/html/rfc6902#section-4.4)，例如：
```yaml
endpoints:
    modified:
        ?   method: POST
            path: /example/callback
        :   extensions:
                modified:
                    x-amazon-apigateway-integration:
                        - oldValue: "201"
                          value: "200"
                          op: replace
                          from: ""
                          path: /responses/default/statusCode
                        - oldValue: http://api.example.com/v1/example/callback
                          value: http://api.example.com/v1/example/calllllllllback
                          op: replace
                          from: ""
                          path: /uri
```

## 排除特定類型的變更

目前適用於 `diff` 與 `summary` 指令。

您可以使用 `--exclude-elements` 旗標排除以下一或多個項目：
- 使用 `--exclude-elements examples` 排除 [Examples（範例）](https://swagger.io/specification/#example-object)
- 使用 `--exclude-elements extensions` 排除 [Extensions（擴充欄位）](https://swagger.io/specification/#specification-extensions)
- 使用 `--exclude-elements description` 排除 description 描述欄位
- 使用 `--exclude-elements title` 排除 title 標題欄位
- 使用 `--exclude-elements summary` 排除 summary 摘要欄位
- 使用 `--exclude-elements endpoints` 排除 [diff 中的 endpoints 端點區段](#路徑-paths-與端點-endpoints)

例如，以下 diff 會排除描述與範例欄位：
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test3.yaml --exclude-elements description,examples -f text
```

## 排除特定名稱的擴充欄位

若您希望依名稱排除特定的 [OpenAPI 擴充欄位](https://swagger.io/docs/specification/openapi-extensions/) 同時保留其他擴充欄位，請使用 `--exclude-extensions` 旗標。  
這與 `--exclude-elements extensions` 不同，後者會排除「所有」擴充欄位。

例如，若僅要排除 `x-internal` 與 `x-ignore` 擴充欄位並保留其他：
```bash
oasdiff diff base.yaml revision.yaml --exclude-extensions x-internal,x-ignore
```

當您有僅供內部或工具使用的擴充欄位（例如 `x-codegen-ignore`、`x-internal`），且不希望這些欄位的變更出現在 diff 報告中時，此功能非常實用。

## 比對行內與 `$ref` 子 Schema

在 `anyOf` 與 `oneOf` 下，oasdiff 預設將行內（inline）子 Schema 與驗證等價的 `$ref` 元件視為相同分支。這能捕捉常見的程式碼生成重構情境：工具將行內 Schema（通常是 enum）擷取為具名元件，但未改變所接受的有效酬載（Payload）：

```yaml
# base
role:
  anyOf:
    - type: string
      enum: [user, admin]
    - type: "null"

# revision (UserRole 為 { type: string, enum: [user, admin] })
role:
  anyOf:
    - $ref: "#/components/schemas/UserRole"
    - type: "null"
```

Diff 會將此 `anyOf` 清單報告為「無變更」，因為所接受酬載的聯集是完全相同的。此行為亦適用於基於同一 diff 引擎執行的 `breaking` 與 `changelog`。

若要恢復先前的行為（將行內分支報告為刪除，將 `$ref` 分支報告為新增），請傳入 `--match-inline-refs=false`。

純註解性質的差異（`title`、`description`、`default`、`example`、`examples`、external docs、`$comment`）在等價性檢查時會被忽略；但影響驗證的差異（包括 `deprecated`）則不會被忽略。元件 `$ref` 重新命名（兩側皆為 `$ref`，如 `UserRoleV1` 改為 `UserRole`）以及行內對行內的修改在設計上不在比對範圍內：僅行內轉為 `$ref` 形式的變更會被視為無變更。

比對是成對進行的：每個刪除的分支最多只會配對一個新增的分支。若新增的 `$ref` 恰好與基準規格中已存在的行內分支等價（即獨立新增且重疊了既有分支，而非重構），仍會被報告為新增。

## 使用範例

### 輸出 YAML 格式的 Diff
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml
```
預設 diff 輸出格式為 `yaml`。  
沒有任何輸出代表 diff 為空，亦即兩份規格沒有差異。

### Text/Markdown 格式的 Diff 報告
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f text
```
文字 diff 報告提供簡化且局部的變更檢視，同時相容於 Markdown。  
若需檢視完整的 diff 細節，請使用 `yaml` 或 `json` 格式。

### HTML 格式的 Diff 報告
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f html 
```
HTML diff 報告提供簡化且局部的變更檢視。  
若需檢視完整的 diff 細節，請使用 `yaml` 或 `json` 格式。

### 透過 HTTP/S 比較遠端檔案
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

### 使用萬用字元（Glob）跨多個規格比較
```bash
oasdiff diff "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c
```
