# 破壞性變更與變更日誌 (Breaking Changes and Changelog)

[English](../BREAKING-CHANGES.md) | [繁體中文](BREAKING-CHANGES.md)

隨著 API 的演進，規格會歷經各種變更。其中某些變更是「破壞性（Breaking）」的，而其他則不是。  
`oasdiff breaking` 指令用於顯示會破壞既有 API 客戶端的變更。  
`oasdiff changelog` 指令用於顯示所有可能影響 API 使用者的變更（包含破壞性與非破壞性變更）。純文件修改（例如 descriptions 描述與 examples 範例）不會包含在內；若需檢視 API 定義的每一處差異，請使用 `oasdiff diff`。  
這些指令常用於 CI 流程中，以回報或阻止破壞性變更發布。

> **附註：** `breaking` 與 `changelog` 執行於 [DIFF.md](DIFF.md) 中所介紹的 diff 引擎之上。該文件中的參數與概念——擴充欄位追蹤（extension tracking）、路徑比對（path matching）、`allOf` 展平（`allOf` flattening）、路徑前綴系列旗標（path-prefix）、標頭大小寫、`--allow-external-refs`、`--fail-on-diff`——在此皆完全適用，並外加下方專屬於破壞性變更的選項。

> **團隊協作功能：** [oasdiff.com](https://www.oasdiff.com) 將這些指令包裝為 PR 留言逐項呈現變更、提供單鍵批准/拒絕按鈕，以及 commit-status 檢查——讓審查者能直接在 Pull Request 中對每一項破壞性變更做出裁決。

## 範例：顯示破壞性變更
```bash
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 範例：顯示變更日誌
```bash
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 在瀏覽器中進行左右對照審查 (`--open`)
`breaking` 與 `changelog` 皆支援 `--open` 旗標。在終端機輸出一般結果後，CLI 會將比對結果上傳至 [oasdiff.com](https://www.oasdiff.com) 並於瀏覽器中開啟呈現左右對照的視覺化審查介面，畫面會精確定位至各項變更發生的位置。畫面中會顯示批准/拒絕按鈕（免費版為唯讀停用狀態；僅在 [oasdiff Pro](https://www.oasdiff.com/pricing) 方案中啟用）。

```bash
oasdiff changelog HEAD~1:openapi.yaml HEAD:openapi.yaml --open
```

比對資料在從您的機器上傳前已完成端對端加密，一次性金鑰僅存在於審查連結的 Hash 部分（`#` 之後的字串），絕不會傳送至我們的伺服器，因此我們儲存的資料伺服器端完全無法讀取。無需登入也無需建立帳號。

產生的 URL 有效期為 7 天。您可以將其貼在 PR 留言或 Slack 中，審查者無需自行安裝 CLI 即可開啟審查畫面。任何取得完整連結的人都能開啟（因為 `#` 後方即為解密金鑰），因此請妥善保管此連結如同保護機密。

## 檢查規則（Checks）
Oasdiff 支援數百項檢查規則（執行 `oasdiff checks changelog` 可檢視目前清單，或至 [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes) 瀏覽完整目錄），分為三個嚴重度層級：  
- `ERR` - 錯誤：確定的破壞性變更，應予以避免
- `WARN` - 警告：潛在的破壞性變更，開發者應特別注意，但無法完全透過程式確認是否必然造成破壞
- `INFO` - 資訊：非破壞性變更

`oasdiff breaking` 僅偵測層級為 `ERR` 與 `WARN` 的變更。  
`oasdiff changelog` 則偵測層級大於或等於 `--level` 參數所指定層級的變更。預設層級為 `INFO`，會包含所有檢查項。

若需檢視完整的檢查項目及其說明，請執行：
```bash
oasdiff checks changelog
```
另請參閱 [自訂嚴重度層級](#自訂嚴重度層級)

## oasdiff 如何判定何為破壞性變更
oasdiff 是根據您的 OpenAPI 定義所宣告的「API 契約（Contract）」來評判變更，而非根據特定後端伺服器湊巧能接受什麼輸入。只要依循舊契約運作的客戶端在新契約下可能無法正常工作，該變更即為破壞性變更。

這點至關重要：因為 OpenAPI 定義宣告了哪些請求與回應是合法的，但多數伺服器在執行階段並不會嚴格執行契約驗證，伺服器可能會默默接受契約明定無效的請求。oasdiff 仍會將此視為破壞性變更，因為同一份契約的其他消費者「會」嚴格執行：例如 API Gateway 與驗證中介軟體會阻擋不符合規範的請求，而由規格產生的客戶端 SDK 會將契約轉換為型別化程式碼，導致編譯錯誤。您的後端伺服器是否寬鬆處理是您自身的選擇，但並不代表 API 契約未曾改變。

值得注意的兩項實務推論：

- 使原本合法的請求變成無效的變更屬於破壞性變更，即使您的後端伺服器仍願意接受。例如：新增一個必填（required）的請求屬性即屬破壞性變更——遺漏該屬性的請求在新契約下即屬非法，無論該屬性是否定義了預設值（default）。預設值僅是伺服器端的備用值（fallback），並不會使遺漏必填屬性的請求變得符合規範。
- `WARN` 僅保留給定義中確實缺乏足夠資訊以做出絕對判斷的變更（例如對物件參數的變更，而該規格未指明其序列化格式）。它不用於「在某些環境會壞、某些環境不會壞」的變更；那類變更屬於錯誤（ERR）。

若檢查項目的預設嚴重度不符合您的 API 相容性政策，可使用 [自訂嚴重度層級](#自訂嚴重度層級) 進行調整。

## 阻止破壞性變更進入
oasdiff 常見的用法是作為 CI/CD 管線中的檢查關卡。  
為了阻止破壞性變更發布，可設定 oasdiff 在偵測到特定層級以上的變更時以錯誤代碼結束：
- 若在發現 ERR 層級變更時以代碼 1 結束，加入 `--fail-on ERR` 旗標。  
- 若在發現 ERR 或 WARN 層級變更時以代碼 1 結束，加入 `--fail-on WARN` 旗標。
- 若在發現任何變更時以代碼 1 結束，加入 `--fail-on INFO` 旗標。

範例：
```bash
oasdiff breaking --fail-on ERR data/openapi-test1.yaml data/openapi-test3.yaml
```

## 一次檢查多個規格
若需比對一組已變更的規格與特定 Git ref 中的對應版本（每個規格獨立比對），請參閱 [BREAKING-FILES.md](BREAKING-FILES.md)。這也是 pre-commit hook 所採用的機制。

## 輸出格式
預設情況下，oasdiff 會以人類易讀的[上色](#顏色色彩)文字格式輸出變更。  
可透過 `--format` 旗標指定其他輸出格式：
- json
- yaml
- githubactions: 適合整合至 GitHub Actions 註解
- junit: 適合整合至 GitLab CI 或測試報表
- html: [檢視範例](https://html-preview.github.io/?url=https://github.com/oasdiff/oasdiff/blob/main/examples/changelog.html)
- markdown: [檢視範例](../../examples/changelog.md)
- text: 預設格式，人類易讀
- singleline: 單行顯示每個變更，便於編寫[忽略清單檔案](#忽略特定的破壞性變更)

範例：
```bash
oasdiff breaking -f yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### JSON Schema
`oasdiff schema` 會印出 `breaking` 與 `changelog` 的 `json` 輸出結構所對應的 [JSON Schema](https://json-schema.org/)（`yaml` 輸出具有完全相同的結構）：
```bash
oasdiff schema
```

## 顏色色彩
當在 Unix 終端機中輸出變更時，oasdiff 會自動使用 ANSI 顏色跳脫字元為文字著色。  
若輸出被導向（pipe）至其他處理程序或重新導向至檔案，oasdiff 會自動停用顏色。  
若需手動控制色彩，可使用 `--color` 旗標傳入 `always` 或 `never`。

## API 穩定度等級 (API Stability Levels)
為 API 指派[穩定度等級](STABILITY.md)允許您依據 API 的成熟度細緻控制允許進行何種變更。

## 廢棄 API (Deprecating APIs)
在刪除端點之前，建議先以「廢棄（deprecation）」方式給予消費者預告期。  
Oasdiff 允許您[優雅地廢棄 API](DEPRECATION.md) 而不觸發破壞性變更錯誤。

## 版本號遞增檢查 (Version Bumps)
若您採用語意化版本（Semantic Versioning），oasdiff 可以回報在未遞增主版本號（Major Version）時釋出的破壞性變更。  
請參閱 [版本號遞增與破壞性變更](VERSIONING.md)。

## 可為 Null 性變更 (Nullability Changes)
Schema 可以透過三種等價的方式允許 `null`，而 nullability 變更是否構成破壞性變更取決於它出現在請求（Request）還是回應（Response）中。  
請參閱 [Nullability 變更](NULLABILITY.md)。

## 於多處共用的 Schema (Schemas Used in Several Places)

當同一個酬載中的多個屬性引用同一個 Schema 時，每個檢查項目只會回報一次，並顯示於抵達該 Schema 的其中一個屬性上。報告中會指出該共用 Schema 的名稱並列出受變更影響的其他屬性：

```
the `customerId` response's property pattern `^[0-9]+$` was added for the status `200` (shared schema: `Id`, also at `userId`)
```

這代表變更是發生在 `components/schemas/Id` 之下，並且同時影響了 `customerId` 與 `userId` 兩個屬性。相同的比對始終會在同一個屬性上回報該變更。

文字輸出中最多列出三個屬性並統計其餘數量。當共用 Schema 彼此嵌套時，每一種組合都會被計算在內。假設 `Inner` 擁有屬性 `id`，`Outer` 的屬性 `a` 與 `b` 皆引用 `Inner`，而酬載的屬性 `p` 與 `q` 皆引用 `Outer`。若為 `id` 新增 `pattern`，將會影響四個屬性路徑：`p/a/id`、`q/a/id`、`p/b/id` 與 `q/b/id`：

```
the `p/a/id` response's property pattern `^[0-9]+$` was added for the status `200` (shared schema: `Inner`, also at `q/a/id`, `p/b/id` and 1 more)
```

JSON 與 YAML 輸出中會提供相同的清單與完整計數：

```json
"sharedSchema": {
  "name": "Inner",
  "properties": ["p/a/id", "q/a/id", "p/b/id"],
  "count": 4
}
```

為什麼只報告一次：變更所影響的屬性數量會隨嵌套層級呈指數級倍增。在 Microsoft Graph API 規格（v1.0，來自 APIs.guru）中，單一請求或回應主體中可透過 164,769 個屬性路徑抵達同一個 Schema。如果每個屬性路徑都產生一條報告，變更日誌將完全無法閱讀且比對過程會極度緩慢；因此 oasdiff 僅回報一次變更並統計其他影響路徑。

當共用 Schema 互相嵌套時，報告會指名距離變更最近的 Schema：修改 `id` 會指名 `Inner`，而在 `Outer` 新增屬性則會指名 `Outer`。

共用 Schema 不會改變變更項目的指紋（Fingerprint），指紋是依據變更的參數計算，而非輸出文字。

以下三件事仍保持獨立，各自享有獨立的輸出條目：

- **每個檢查規則（Check）：** 每個規則各自獨立回報。因此若一個共用 Schema 既新增了屬性又加上了 `pattern`，會分別產生一條 `response-optional-property-added` 與一條 `response-property-pattern-added`。
- **每個變更項目：** 在同一個共用 Schema 中新增兩個屬性代表兩個獨立變更，皆會在同一個代表屬性上回報。
- **每個操作（Operation）：** 每個 Operation 是獨立的契約，因此抵達同一個變更 Schema 的不同 Operation 都會各自回報。

## 忽略特定的破壞性變更
有時您需要允許某些特定的破壞性變更（例如規格與後端服務暫時不同步，需要修正規格）。  
Oasdiff 允許您在設定檔中定義欲忽略的破壞性變更。  
您可以在 oasdiff 命令列中使用 `--warn-ignore` 指定忽略 WARNINGS 的檔案，或使用 `--err-ignore` 指定忽略 ERRORS 的檔案。  
設定檔中的每一行應包含兩個部分：
1. 方法與路徑（該行中第一個以斜線 `/` 開頭的欄位）以忽略特定端點的變更；或使用關鍵字 `components` 忽略元件層級的變更
2. 該破壞性變更的文字描述

例如：
```
GET /api/{domain}/{project}/badges/security-score removed the success response with the status '200'
```
或是針對元件變更：
```
components removed the schema 'rules'
```

必填欄位可以任何順序出現，大小寫皆可，且設定行中可以包含額外文字（例如日期或說明），例如：
```
 - 12.01.2023 In GET /api/{domain}/{project}/badges/security-score we removed the success response with the status '200'
 - 31.10.2023 Removed the schema 'network-policies' from components
```

設定檔可以是任何文字格式（如 Markdown），方便您直接以此檔案記錄破壞性變更與其他重要變動歷史。

## Enum 列舉值的破壞性變更
Oasdiff 透過 `x-extensible-enum` 擴充欄位支援列舉變更的特殊規則。  
此方法允許在回應（Responses）中所使用的 enum 中加入新值，這在許多情境下非常實用，但要求客戶端在收到未知列舉值時必須具備降級至預設邏輯的處理能力。  
`x-extensible-enum` 最早由 [Zalando](https://opensource.zalando.com/restful-api-guidelines/#112) 提出並被 OpenAPI 社群廣泛採用。技術上它可以用 `anyOf` + 傳統 enum 取代，但 `x-extensible-enum` 更加直觀明確。  
在多數情況下，`x-extensible-enum` 與一般 enum 類似，但它允許在傳送給客戶端的訊息（回應或回呼）中新增列舉值。  
若您的 OpenAPI 規格未使用 `x-extensible-enum`，一切維持原樣；若有使用，oasdiff 會針對 `x-extensible-enum` 參數與屬性進行相應的破壞性變更檢驗。

## 本地化與多國語言 (Localization)
若要以其他語言顯示變更訊息，請使用 `--lang` 旗標。  
目前支援英文 (en)、俄文 (ru) 與巴西葡萄牙文 (pt-br)。  
[歡迎協助將您的語言加入 oasdiff](https://github.com/oasdiff/oasdiff/issues/383)。

## 自訂嚴重度層級 (Customizing Severity Levels)
Oasdiff 允許您依據自身需求修改預設的嚴重度層級。  
例如，`api-security-scope-removed` 檢查項目的預設層級為 `INFO`，您可以透過 `oasdiff checks changelog` 確認。  
若要將 `api-security-scope-removed` 的嚴重度層級改為 `ERR`，請執行：
```bash
oasdiff changelog data/checker/api_security_updated_base.yaml data/checker/api_security_updated_revision.yaml --severity-levels oasdiff-levels.txt
```
其中 `oasdiff-levels.txt` 檔案包含一行設定：
```
api-security-scope-removed    err
```

可自訂的層級如下表：
| 自訂層級 | 檢查規則狀態 |
|---|---|
| err | 啟用，層級為 ERR |
| warn | 啟用，層級為 WARN |
| info | 啟用，層級為 INFO |
| none | 停用此檢查規則 |

### 當 oasdiff 回報的層級低於檢查規則所宣告的層級時
檢查規則所宣告的層級（即 `oasdiff checks changelog` 所列出的層級）代表該變更在一般通常情況下的含義。對於特定變更，oasdiff 可能會以較低的層級回報，並附帶註解說明原因。這通常發生在兩種情況：

- **比對不精確：** 發生在 `allOf` 內部的變更最為常見：若未使用 [`--flatten-allof`](ALLOF.md)，分支是逐一比對的，因此某個分支刪除的條件可能仍在另一個分支中得到保證。此類變更即使規則宣告為 `ERR`，也會以 `WARN` 回報，註解中會說明缺失了什麼以及應執行什麼參數以取得精確結果。
- **規格證明該變更不會對任何人造成破壞：** 標記為 `readOnly` 的屬性絕不會出現在請求中，因此縮緊該屬性的限制（加入 `pattern`、設定 `maximum`）絕不會使請求無效；同理，回應中的 `writeOnly` 屬性亦然。此類變更即使規則宣告為 `ERR`，也會以 `INFO` 回報並說明原因。各項規則所支援的條件保護列於 `oasdiff checks changelog` 的 GUARDS 欄位中（請參閱 [CHECKS.md](CHECKS.md#分類規則)）。

**自行設定層級將關閉該檢查項目的自動降級機制。** 您所設定的層級即為最終層級，無論上述情況是否存在。這在兩個方向上皆有影響：將規則釘選為 `err` 意味著即使 oasdiff 原本會將其降級，它依然會回報 `ERR`；而將規則釘選為其原有的預設層級也足以關閉自動降級。

## 自訂破壞性變更檢查規則
若您遇到未被回報的變更，您可以：
1. 執行 `oasdiff checks changelog` 確認該規則是否存在，並[依需求自訂嚴重度層級](#自訂嚴重度層級)。
2. 新增[自訂檢查規則](../CUSTOMIZING-CHECKS.md)

## 已知限制
- 尚未針對請求參數使用 `context` 代替 `schema` 提供檢查
- 尚未針對 `callback` 提供檢查
