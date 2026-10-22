# OpenAPI 3.1 支援 (OpenAPI 3.1 Support)

[English](../OPENAPI-31.md) | [繁體中文](OPENAPI-31.md)

oasdiff 的所有指令皆完整支援 OpenAPI 3.1 規格。  
同時亦可讀取 3.2 規格；詳見下方的 [OpenAPI 3.2](#openapi-32)。

OpenAPI 3.1 支援自 `v1.15.0` 正式提供（GA）。先前的 Beta 標籤（`v1.15.0-openapi31.beta.*` 與 `v2.2.0-openapi31.beta.*`）已被取代。

## 支援功能

### 規格載入
- OpenAPI 3.1.0 文件解析
- JSON Schema 2020-12 關鍵字
- Webhooks

### Diff 差異比較
能偵測所有 3.1 專屬欄位的變更：
- 數值形式的 `exclusiveMinimum`/`exclusiveMaximum`
- `const`, `prefixItems`, `contains`, `minContains`, `maxContains`
- `if`/`then`/`else` 條件 Schema
- `dependentRequired`, `dependentSchemas`
- `patternProperties`, `propertyNames`
- `unevaluatedItems`, `unevaluatedProperties`（支援布林值與 Schema 形式）
- `contentSchema`, `contentMediaType`, `contentEncoding`
- `$defs`, `$schema`, `$comment`
- Webhooks（新增 / 刪除 / 修改）
- Info `summary`, License `identifier`

### Validate 規格驗證
`validate` 清楚掌握 3.1 引入的所有新欄位，因此在宣告舊版本的文件中若出現了 3.1 專屬欄位會被主動回報而非靜默忽略。共有 29 條驗證規則覆蓋各欄位（`const`、`$anchor`、`$comment`、`contains`、`contentEncoding` 等）。

### Flatten 扁平化
`--flatten-allof` 會在 `allOf` 同級項目中合併 3.1 關鍵字：`const`、`contains`/`minContains`/`maxContains`、`contentMediaType`/`contentEncoding`、`dependentRequired`、`propertyNames` 以及多型別陣列。在正式依賴此功能前，有兩項行為值得注意（皆詳述於 [ALLOF.md](ALLOF.md)）：當同級項目的 `contains` 不一致時會合併為比原規格更嚴格的限制，且 `$defs` 會在展平輸出中被移除。

### 破壞性變更與變更日誌
新增 162 項規則 ID 涵蓋：
- **Nullable 型別陣列**：`type: ["string", "null"]` 會被正確辨識為 nullability 變更（而非型別變更的誤報）
- **Webhooks**：新增/刪除偵測 + 所有既有的操作層級檢查會自動套用於被修改的 Webhook 操作
- **const**：請求與回應主體/屬性的新增/刪除/修改
- **exclusiveMinimum/exclusiveMaximum**：請求主體/屬性/參數與回應主體/屬性的增加/減少/設定
- **prefixItems**：請求與回應的新增/刪除
- **if/then/else**：請求與回應的新增/刪除
- **contains/minContains/maxContains**：主體、屬性、參數與回應標頭層級的新增/刪除/增加/減少/邊界設定與移除
- **dependentRequired**：新增/刪除/修改
- **dependentSchemas**：新增/刪除
- **patternProperties**：新增/刪除
- **propertyNames**：新增/刪除
- **unevaluatedItems/unevaluatedProperties**：新增/刪除
- **contentSchema/contentMediaType/contentEncoding**：新增/刪除/修改

## OpenAPI 3.2

oasdiff 支援讀取 3.2 規格。OpenAPI Initiative 保證 3.x 未來版本（3.2.x, 3.3.x, ...）嚴格向後相容，因此能正確讀取 3.1 的解析器也能讀取 3.2，上述所有章節皆完全適用。

若 3.2 的語法結構被錯誤回報，或變更完全未被回報，請參閱下方的[意見反饋](#意見反饋)，並在標題註明 `[3.2]`。

## 從 3.0 遷移至 3.1

OpenAPI 3.1 改變了若干語法結構的形式：

| 3.0 | 3.1 |
|---|---|
| `type: string` + `nullable: true` | `type: ["string", "null"]` |
| `exclusiveMinimum: true` + `minimum: 0` | `exclusiveMinimum: 0`（數值） |
| `example: 7` | `examples: [7]` |

從 3.0 遷移至 3.1 的團隊面臨兩個問題，oasdiff 皆能提供協助：產出遷移後的規格，以及驗證遷移未破壞客戶端相容性。

### 使用 `oasdiff upgrade` 轉換規格

`upgrade` 子指令可將 3.0 規格重寫為最新的 3.x 規範形式（目前為 3.2.1）。轉換操作具備等冪性（Idempotent），對已符合規範的規格重複執行只會遞增版本字串。

目標鎖定最新的 3.x 而非 3.1，是因為上述的相容性保證，因此單次遷移即可涵蓋後續數個版本。詳情請參閱 [Why the target is 3.2](https://www.oasdiff.com/docs/openapi-31-migration)。

```bash
oasdiff upgrade old-spec.yaml > new-spec.yaml
oasdiff upgrade old-spec.yaml --format json > new-spec.json
cat old-spec.yaml | oasdiff upgrade -
```

輸出會寫入至 stdout；重新導向至檔案即可儲存。預設輸出格式為 `yaml`；傳入 `--format json` 則可輸出 JSON。

`upgrade` 僅處理 3.0 → 3.x。Swagger 2.0 → 3.0 不在範圍內。

自 `v1.16.0` 起提供。

### 使用 `--auto-upgrade` 跨版本比較

`diff`、`breaking`、`changelog` 與 `summary` 支援 `--auto-upgrade` 旗標，會在載入規格後於記憶體中直接對兩份規格執行相同的標準化轉換。這讓跨版本比較（例如 3.0 基準版本 vs 3.1 修訂版本）能夠產生有意義的結果，避免比對結果被大量的語法方言差異雜訊淹沒。

```bash
# old.yaml 為 3.0，new.yaml 為 3.1
oasdiff breaking  old.yaml new.yaml --auto-upgrade
oasdiff changelog old.yaml new.yaml --auto-upgrade
oasdiff diff      old.yaml new.yaml --auto-upgrade
oasdiff summary   old.yaml new.yaml --auto-upgrade
```

若未傳入此旗標，diff 會將 3.0→3.1 的方言改寫（如 `nullable` 變為 `type: ["string", "null"]` 等）回報為 Schema 變更。加上此旗標後，兩側規格皆先被標準化，因此只會保留實質的 Schema 層級差異。

此旗標預設為關閉；需在呼叫時主動啟用。即使兩份規格已經是相同版本，設定此旗標也是完全安全的：轉換操作對已標準化的輸入具備等冪性。

自 `v1.16.0` 起提供。

## 注意事項與限制

以下 3.1 功能目前在 [`kin-openapi`](https://github.com/getkin/kin-openapi)（oasdiff 所使用的解析器）中尚未完全支援，因此不會出現在 oasdiff 的 diff 結果中：

- **`$dynamicRef` / `$dynamicAnchor`**：能被解析但載入時尚未被引用解鎖。使用動態引用實現遞迴定義的 Schema 將無法深入追蹤。
- **`components` 中的 `pathItems`**：未體現在解析器的資料模型中。在 components 中宣告可重複使用 path items 的規格會被靜默忽略。

若您遭遇上述問題，請[建立 Issue 回報](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20)。

## 意見反饋

發現問題？請[在此建立 Issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20)，標題請註明 `[3.1]`；若是上述 3.2 功能請至[此處](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.2]%20)並在標題註明 `[3.2]`。3.2 支援仍在持續完善中，因此回報某語法尚未支援與回報處理錯誤具有同等的價值。
