# 廢棄 API、參數與屬性 (Deprecating APIs, Parameters and Properties)

[English](../DEPRECATION.md) | [繁體中文](DEPRECATION.md)

有時候某些[資源](#支援廢棄機制的資源)需要從 API 中移除，例如當以新版本取代舊端點，或是某個參數不再支援時。  
作為 API 擁有者，我們希望有一套流程能盡可能平順地淘汰舊資源，並將對業務的衝擊降到最低。

OpenAPI 規範支援 `deprecated` 旗標，可用於將資源標記為已廢棄。  
將資源標記為廢棄**不被視為**破壞性變更，因為它不會破壞客戶端的功能，僅作為未來預計移除的意向宣告。  
在廢棄該資源後，即可將其移除而不觸發破壞性變更錯誤，因為客戶端早已知曉該資源即將退場。

## 無日落日期的廢棄 (Deprecation without a sunset date)
Oasdiff 允許您透過以下步驟平順移除資源而不產生破壞性變更錯誤：
1. 將該資源標記為 `deprecated`:
   ```yaml
   /api/test:
     get:
       deprecated: true
   ```
2. 隨後，即可直接移除該資源，而不會觸發破壞性變更錯誤。

## 具備日落日期的廢棄 (Deprecation with a sunset date)
更成熟的廢棄流程會包含日落日期（Sunset date），明確告知 API 客戶端在退場前還有多長的時間可以繼續使用該資源：
1. 將資源標記為 `deprecated`，並加上[擴充欄位](https://swagger.io/specification/#specification-extensions) `x-sunset` 宣告該資源正式移除的日期：
   ```yaml
   /api/test:
     get:
       deprecated: true
       x-sunset: "2022-12-31"
   ```
   請注意，日落日期必須符合 [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339) 格式。
2. 在日落當天或之後的任何時間，即可移除該資源而不觸發破壞性變更錯誤。若早於日落日期移除則會被視為破壞性變更。

附註：
1. 在此模式下，oasdiff 將 `x-sunset` 擴充欄位視為**選填（optional）**，允許開發者自由選擇是否附帶 `x-sunset`。
2. 一旦指定了 `x-sunset` 擴充欄位，將日期提前至更早的日期會觸發破壞性變更。

## 強制執行日落寬限期 (Enforcing Sunset Grace Period)
Oasdiff 允許您強制要求日落寬限期：即宣告廢棄與實際移除之間所需的最少天數。  
在此模式下，oasdiff 將 `x-sunset` 擴充欄位視為**必填（mandatory）**，廢棄任何資源時必須同時附帶 `x-sunset`，否則會被視為破壞性變更。

例如，以下指令要求廢棄資源時必須附帶至少 180 天以後的 `x-sunset` 日期，否則該廢棄宣告本身即被視為破壞性變更：
```bash
oasdiff breaking data/deprecation/base.yaml data/deprecation/deprecated-past.yaml --deprecation-days-stable=180
```

若您採用了 [API 穩定度等級](STABILITY.md)，您可以分別為不同等級定義日落寬限期：
- 穩定（Stable）API：使用 `--deprecation-days-stable`
- Beta API：使用 `--deprecation-days-beta`

附註：
1. 寬限天數可設定為非負整數。
2. 將寬限天數設為 0 會關閉強制寬限期檢查，並回復為[具備日落日期的廢棄](#具備日落日期的廢棄-deprecation-with-a-sunset-date)行為。
3. 指定 `x-sunset` 擴充欄位後，後續僅能將其修改為相對於修改當日仍滿足寬限期要求的未來日期。

## 支援廢棄機制的資源
OpenAPI 3 針對 `Operations`、`Parameters`、`Headers` 與 `Schemas` 皆支援 `deprecated` 欄位。  
Oasdiff 目前針對 `Operations` 與 `Parameters` 完整支援上述的完整廢棄流程：一旦廢棄（且超過設定的日落日期），便可安全移除而不產生破壞性變更。

對於請求與回應主體中的屬性（`Properties`），oasdiff 能偵測到廢棄標記並檢查日落日期，但目前移除已廢棄屬性仍會如同移除一般屬性一樣回報（例如 `response-required-property-removed`）。對移除已廢棄屬性的完整支援目前在 [#858](https://github.com/oasdiff/oasdiff/issues/858) 追蹤中。
