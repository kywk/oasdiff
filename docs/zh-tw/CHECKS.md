# 檢查規則 (Checks)

[English](../CHECKS.md) | [繁體中文](CHECKS.md)

`oasdiff checks` 指令用於顯示 oasdiff 所套用的規則清單。每個規則集有獨立的子指令：

- `oasdiff checks changelog` — 比較兩份規格時，`oasdiff breaking` 與 `oasdiff changelog` 所套用的檢查規則。本頁主要介紹此規則集。
- `oasdiff checks validate` — 針對單一規格時，`oasdiff validate` 所回報的驗證規則。

此指令通常用於探索 oasdiff 能偵測哪些變更，或是查詢特定的檢查規則 ID，以便進行規則忽略或自訂嚴重度層級。

單獨執行 `oasdiff checks` 會印出這兩個子指令：規則清單始終需要指明規則集名稱。

## 範例：顯示所有檢查規則
```bash
oasdiff checks changelog
```

## 輸出格式
預設輸出格式為 `text`。  
可透過 `--format` 旗標產生其他格式：
- text: 包含 ID、說明與嚴重度層級的人類易讀表格（預設）
- yaml: 適合進一步自動化處理的機器可讀格式
- json: 適合進一步自動化處理的機器可讀格式

## 依嚴重度篩選
使用 `--severity` 僅顯示特定層級的檢查規則：
```bash
oasdiff checks changelog --severity error
oasdiff checks changelog --severity warn
oasdiff checks changelog --severity info
```

檢查規則分為三個嚴重度等級：
- `error` — 確定的破壞性變更，應予以避免
- `warn` — 潛在的破壞性變更，無法透過程式完全確認
- `info` — 非破壞性變更

執行 `oasdiff checks changelog` 檢視當前清單，或至 [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes) 瀏覽完整目錄。

## 分類規則 (Categorization)
每個檢查項目皆依據獨立的維度進行分類，並輸出於 `json` 與 `yaml` 的對應欄位中：

- `area` — 檢查所關注的 OpenAPI 物件，與 OpenAPI 規範的物件模型一致：`schema`、`parameters`、`requestBody`、`responses`、`paths`、`headers`、`security`、`tags`、`components`。
- `kind` — API 契約變動的面向：`existence`（元素新增或刪除）、`requiredness`（必填 / 選填 / nullable）、`mutability`（唯讀 read-only / 唯寫 write-only）、`type`（資料型別或格式）、`constraints`（限制邊界如 min/max、長度、正則 pattern、項目數）、`values`（enum、const、預設值 default）、`structure`（組合與條件關鍵字如 allOf/anyOf/oneOf、discriminator、if/then/else、contains），以及 `lifecycle`（廢棄 deprecation、日落 sunset、穩定度 stability）。
- `actions` — 檢查所涵蓋的語法編輯動作：`add`、`remove`、`change`、`increase`、`decrease`、`set`、`unset`。
- `effect` — 該檢查對於契約所接受之有效酬載集合的判定：`widens`（放寬）、`narrows`（縮緊）、`incomparable`（既拒絕了原本合法的酬載又接受了原本非法的酬載）、`unknown`（無法判定）、`violation`（違反了廢棄/穩定度契約，而非資料傳輸契約），或 `none`（中繼資料，對接受的酬載無影響）。`effect` 與 `direction` 共同決定預設嚴重度：縮緊請求或放寬回應會破壞客戶端相容性。
- `direction` — 傳輸方向：`request`、`response` 或 `none`。
- `guards` — 具備保護條件的規範情境，在此情境下常規判定不適用（顯示於文字輸出的 GUARDS 欄位）：`read-only` / `write-only`（屬性不會出現在該側，因此限制條件不會造成破壞，並以 `info` 回報附帶說明）、`sanctioned`（被刪除的元素先前已宣告廢棄且日落期限已被遵守）、`non-success`（非成功回應狀態，responses 映射表並不保證窮舉）、`has-default`（元素宣告了預設值）、`negotiated`（客戶端所選擇的元素，如媒體型別或回應狀態碼，嚴重度以請求側方式推導），以及 `required-in-requests`（OpenAPI 3.1+ 中必填的 read-only 屬性，`required` 亦適用於請求，因此客戶端已經送出）。請參閱[當 oasdiff 回報的層級低於檢查規則所宣告的層級時](BREAKING-CHANGES.md#當-oasdiff-回報的層級低於檢查規則所宣告的層級時)。
- `generated` — 該檢查是由規則產生器自動生成而非手寫函式（顯示於文字輸出的 GENERATED 欄位）。請參閱 [../CUSTOMIZING-CHECKS.md](../CUSTOMIZING-CHECKS.md#second-is-the-check-generated)。

`--id` 依 ID 顯示單一檢查項目；`--location` 則僅保留位置路徑中包含指定字串的檢查項目：
```bash
oasdiff checks changelog --id request-body-max-set
oasdiff checks changelog --location requestBody.content.*.schema.maximum
```

## 依標籤篩選 (Filtering by Tag)
使用 `--tags` 僅顯示特定 area、kind、action、effect 或 direction 的檢查規則：
```bash
oasdiff checks changelog --tags request,parameters
oasdiff checks changelog --tags schema,constraints
```

各維度可用的標籤：

- direction: `request`, `response`
- action: `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`
- effect: `widens`, `narrows`
- area: `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`
- kind: `existence`, `requiredness`, `mutability`, `type`, `constraints`, `values`, `structure`, `lifecycle`
- guard: `read-only`, `write-only`, `sanctioned`, `non-success`, `has-default`, `negotiated`, `required-in-requests`
- provenance: `generated`, `hand-written` — 每個檢查項目必屬兩者之一

相同維度的值以 OR 結合，不同維度的值以 AND 結合：`--tags request,response,add` 會選取「（request 或 response）且為 add」的檢查項目。

## 覆蓋率對應表 (Coverage Map)
`oasdiff checks changelog coverage` 列出 OpenAPI 文件的所有可能編輯操作及其稽核結論，每種編輯一行：

- `covered` — 有相應檢查規則覆蓋該編輯
- `waived` — 免除檢查；`category` 欄位會說明原因：`open`（尚未實作的檢查項目，附有原因與建議 ID）、`resolved-at-usage`（元件定義是在被引用的操作處進行比較，各自擁有獨立行），或 `covered-as`（相同的文件編輯操作已在其他動作下回報）
- `uncovered` — 無檢查規則且未免除（建置在此狀態下會失敗，因此該清單通常為空）
- `non-contract` — 該編輯無法改變哪些酬載為有效，因此不預期有檢查規則

各維度標籤：

- status: `covered`, `uncovered`, `waived`, `non-contract`
- polarity: `request`, `response`, `document`（非傳輸方向）, `shared`（元件，其方向取決於引用位置）
- action: `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`

如同變更日誌清單，相同維度使用 OR 結合，不同維度使用 AND 結合。`--location` 保留包含指定字串的位置，`--id` 保留指定檢查項目所主張的編輯，`--format text|json|yaml` 選擇輸出格式。

```bash
oasdiff checks changelog coverage --tags waived,request
oasdiff checks changelog coverage --tags covered,add,remove
oasdiff checks changelog coverage --location paths.*.*.parameters.*.style
oasdiff checks changelog coverage --id request-body-max-set
```

`--patterns` 以摘要方式匯總相同的統計結果而非列出所有行：每個免除（waiver）或非契約（non-contract）項目一行，並顯示其涵蓋的編輯數量，讓您無需閱讀數千行即可迅速掌握未受檢編輯的分佈原因。此參數不接受 `--tags`。

```bash
oasdiff checks changelog coverage --patterns

KIND         PATTERN                                       EDITS REASON
waiver       webhooks.**                                   2353  webhooks are diffed (WebhooksDiff) but checkers only ...
non-contract **.servers.**                                 84    server URLs are deployment metadata, not part of the ...
```

相同的會計審計機制已由單元測試強制驗證，因此清單始終保持最新狀態。在 `checks changelog --format json|yaml` 中的 `locations` 欄位則顯示反向對應：各檢查項目所聲稱覆蓋的編輯位置。

## 多國語言在地化 (Localization)
使用 `--lang` 以支援的語言檢視檢查規則說明：
```bash
oasdiff checks changelog --lang ru
```
支援語言：`en`（預設）、`ru`、`pt-br`、`es`。

請注意僅變更日誌清單支援 `--lang`；驗證（validate）規則說明目前未進行本地化。

## Validate 檢查規則
`oasdiff checks validate` 列出 `oasdiff validate` 所回報的規則，支援相同的 `--format` 與 `--severity` 旗標：
```bash
oasdiff checks validate
oasdiff checks validate --severity error --format json
```
它不接受 `--tags`（validate 規則不帶標籤）亦不接受 `--lang`。

## 使用檢查規則 ID
每個檢查項目擁有唯一的 ID（例如 `api-path-removed-without-deprecation`），可用於：
- [忽略特定的破壞性變更](BREAKING-CHANGES.md#忽略特定的破壞性變更)
- [自訂嚴重度層級](BREAKING-CHANGES.md#自訂嚴重度層級)
- [撰寫自訂檢查規則](../CUSTOMIZING-CHECKS.md)
