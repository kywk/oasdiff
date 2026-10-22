# 可為 Null 性變更 (Nullability Changes)

[English](../NULLABILITY.md) | [繁體中文](NULLABILITY.md)

在 Schema 中可以透過三種等價的方式允許 `null` 值：

1. `nullable` 關鍵字（OpenAPI 3.0）：
   ```yaml
   type: string
   nullable: true
   ```
2. 在型別陣列中包含 `"null"` 項目（OpenAPI 3.1）：
   ```yaml
   type: [string, "null"]
   ```
3. 帶有 null 分支的 `oneOf`。這是讓透過 `$ref` 引用的 Schema 具備 nullable 特性的常見做法，因為 `$ref` 本身無法攜帶型別陣列：
   ```yaml
   oneOf:
     - type: "null"
     - $ref: '#/components/schemas/Address'
   ```

oasdiff 將上述三種形式視為同一件事：加入任一種形式會被回報為 `became-nullable`（變為可為 null），移除任一種形式則會被回報為 `became-not-nullable`（變為不可為 null）。

這對於 `oneOf` 形式尤為重要。將 Schema 包裝在 `oneOf` 中會使其向下移動一層，未經處理的比對工具會誤判為型別變更、enum 移除與 `oneOf` 新增，並將其中部分視為破壞性變更。oasdiff 能辨識出被包裝的 Schema 與原本等價，因此只會精確回報一條 nullability 變更：

```
oasdiff changelog data/checker/nullable_wrap_base.yaml data/checker/nullable_wrap_revision.yaml

info    [request-property-became-nullable]
        in API POST /test
                the request property `optionalEnum` became nullable
```

## 是否構成破壞性變更？
nullability 變更是否構成破壞性變更取決於**資料傳輸的方向**：

- **請求（Request）** 由客戶端送出。變為 nullable 的請求 Schema 除了原本接受的所有值之外，還額外接受 `null`：**非破壞性**。變為 not nullable 的請求 Schema 會拒絕原本合法的 `null` 值：**破壞性變更**。
- **回應（Response）** 由伺服器回傳。變為 nullable 的回應 Schema 可能會回傳 `null` 給未預期的客戶端：**破壞性變更**。變為 not nullable 的回應 Schema 則代表伺服器提供了比以往更嚴格的保證：**非破壞性**。

| 變更行為 | 請求 (Request) | 回應 (Response) |
|---|---|---|
| became nullable (變為可為 null) | info | error |
| became not nullable (變為不可為 null) | error | info |

例如，從上述請求 Schema 中移除包裝即構成破壞性變更：

```
oasdiff changelog data/checker/nullable_wrap_revision.yaml data/checker/nullable_wrap_base.yaml

error   [request-property-became-not-nullable]
        in API POST /test
                the request property `optionalEnum` became not nullable
```

參數遵循請求欄位的邏輯：`request-parameter-became-nullable` (info) 與 `request-parameter-became-not-nullable` (error)，物件參數屬性則對應 `request-parameter-property-*`。

## 限制
oasdiff 僅在 `oneOf` 屬於純包裝時進行辨識：必須剛好有兩個分支，其中一個僅為 `{type: "null"}`，另一個與原始 Schema 等價。當編輯操作超越單純的 nullability 變更時，oasdiff 會退回至回報底層變更，秉持「寧可多報不可漏報破壞性變更」的原則：

- **包裝原本就已 nullable 的 Schema** 不算 nullability 變更。在 `oneOf` 下，數值必須剛好符合一個分支；若 Schema 本身已允許 `null`，則 `null` 值會同時符合兩個分支而被拒絕。
- **在包裝的同時修改了 Schema**（例如在加入包裝的同時刪除了一個 enum 值），會依底層變更（enum 刪除）進行回報，而非回報為 nullability 變更。
- **在已有數個分支的 `oneOf` 中新增 null 分支**，會被回報為型別清單變更（針對簡單型別）或 `oneOf` 變更（針對物件），而非 nullability 變更。
- **回應標頭（Response headers）** 目前尚未檢查 Schema 變更（包括 nullability，[#1094](https://github.com/oasdiff/oasdiff/issues/1094)）。請求標頭屬於參數（`in: header`），已由參數檢查規則覆蓋。
