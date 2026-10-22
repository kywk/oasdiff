# 原始碼行號定位追蹤 (Source Location Tracking)

[English](../SOURCE-LOCATOR.md) | [繁體中文](SOURCE-LOCATOR.md)

oasdiff 能將破壞性變更與變更日誌項目精確關聯至 OpenAPI 規格檔案中發生變更的確切行號（Line）與欄號（Column）。  
這能讓 GitHub Pull Request 顯示行內註解，直接將審查者引導至發生變動的程式碼行。

原始碼行號定位需要 YAML 格式的規格檔案。請參閱[限制](#限制)。

## 功能說明

啟用原始檔位置追蹤後，回報的變更項目會包含：

- **BaseSource**：在基準規格（舊版）中的檔案、行號與欄號
- **RevisionSource**：在修訂規格（新版）中的檔案、行號與欄號

代表刪除的變更（例如端點刪除、屬性刪除）只會具備 `BaseSource`，因為被刪除的元素在修訂版本檔案中已不存在。  
同理，代表新增的變更（例如端點新增、新 enum 值）只會具備 `RevisionSource`，因為新增的元素在基準版本檔案中不存在。  
代表修改的變更（例如 maxLength 增加）則同時具備 `BaseSource` 與 `RevisionSource`。

### 跨多檔案規格 (Multi-file specs)

當規格使用 `$ref` 從其他 YAML 檔案引入 Schema 時，來源位置會精確指向發生變更的元素所在的實際檔案及其精準行號與欄號。`BaseSource` 與 `RevisionSource` 的 `file` 欄位會反映所引入的子檔案路徑，而非僅顯示最外層規格的進入點路徑，GitHub Actions 的行內註解亦會採用該路徑。

引入 JSON 檔案的行為有所不同，請參閱[限制](#限制)。

## 限制

### JSON 格式規格
來源位置追蹤僅適用於由 oasdiff 解析為 YAML 的規格。被解析為 JSON 的規格不攜帶位置資訊，因此：

- 在 `-f json` 與 `-f yaml` 輸出中，`baseSource` 與 `revisionSource` 通常會被完全省略；若存在，也僅包含 `file` 而無 `line` 或 `column`。
- GitHub Actions 註解產生時不會帶有 `line` 與 `col`，因此 GitHub 會將註解掛載於檔案層級而非特定的變更行。
- `oasdiff validate` 回報違規時僅有檔名而無線號與欄號。

此判定取決於檔案的**實質內容**而非副檔名。即使檔名是 `.yaml` 但內部包含 JSON 格式內容，仍會被視為 JSON 解析，行為相同。

在跨多檔案規格中，引入的 JSON 檔案內發生的變更並非單純失去位置資訊，而是會回退定位至最接近且具備位置資訊的外層包裹元素，通常是父層檔案中的 `$ref` 所在行。因此回報的 `file` 與 `line` 可能會指向與實際變更不同的檔案。

若需取得精確的行號定位，目前請將規格轉換為 YAML 格式。若您需要 JSON 規格的來源位置支援，歡迎至 [Issue #1128](https://github.com/oasdiff/oasdiff/issues/1128) 按 👍 讓我們評估社群需求。

## 輸出格式

`-f` 旗標決定 oasdiff 產出報告的格式，這與規格本身採用的格式無關。

原始檔位置支援以下格式：
- **JSON** (`-f json`) 與 **YAML** (`-f yaml`)：輸出中每項變更包含 `baseSource` 與 `revisionSource` 欄位
- **GitHub Actions** (`-f githubactions`)：輸出包含 `revisionSource`，可直接在 PR 的「Files changed」標籤頁作為行內註解呈現。請注意 `baseSource` 不會顯示，因為 GitHub 僅能在檔案的最新修訂版本（Head）上顯示註解。

在上述格式中，數值僅在規格被解析為 YAML 時才會填入。請參閱[限制](#限制)。

### 精確度層級

來源位置會盡可能以最精確的層級回報：
- **欄位層級 (Field-level)**：指向發生變更的確切欄位（例如型別改變時指向 `type:` 所在行）
- **序列項目層級 (Sequence item-level)**：指向清單中的特定項目（例如特定的 enum 值或必填屬性名稱）
- **子物件層級 (Sub-object-level)**：指向發生變更的 Schema、參數或回應物件

## 範例輸出

### GitHub Actions 行內註解

新增與修改會以行內註解顯示：
```
::error title=new-required-request-property,file=openapi.yaml,line=30,col=13::in API POST /users added the new required request property 'email'
```

刪除項目則無行內註解，因為 GitHub 僅將 Actions 註解附加至修訂版（Head）檔案上，而非基準版檔案。`BaseSource` 依然會輸出於 `-f json` 與 `-f yaml` 中，並指向被刪除元素在基準規格中的確切行號。

```
::error title=request-property-removed::in API POST /users removed the request property 'name'
```

### JSON 輸出 (`-f json`)

```json
{
  "id": "request-property-type-changed",
  "level": 3,
  "operation": "POST",
  "path": "/users",
  "baseSource": {"file": "base.yaml", "line": 42, "column": 15},
  "revisionSource": {"file": "revision.yaml", "line": 42, "column": 15}
}
```

### YAML 輸出 (`-f yaml`)

```yaml
- id: request-property-type-changed
  level: 3
  operation: POST
  path: /users
  baseSource:
    file: base.yaml
    line: 42
    column: 15
  revisionSource:
    file: revision.yaml
    line: 42
    column: 15
```

## 演示

**線上體驗：** [oasdiff.com/diff](https://oasdiff.com/diff) 能在瀏覽器中視覺化呈現任何兩個 OpenAPI 規格的左右對照比較，無需安裝任何工具。
