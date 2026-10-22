# 合併 AllOf Schemas [Beta] (Merging AllOf Schemas)

[English](../ALLOF.md) | [繁體中文](ALLOF.md)

OpenAPI 3.0 提供了數個可用於組合 Schema 的關鍵字。  
您可以使用這些關鍵字建立複雜的資料模型或針對多項準則驗證數值：  
- `oneOf` – 驗證數值必須剛好符合其中一個子 Schema
- `allOf` – 驗證數值必須符合所有的子 Schema
- `anyOf` – 驗證數值必須符合任一個（一或多個）子 Schema

使用這些關鍵字有助於描述複雜的資料結構，但同時也會增加破壞性變更檢測的複雜度。  
以比較以下兩份 OpenAPI 規格為例：
```bash
oasdiff breaking data/allof/simple.yaml data/allof/revision.yaml 
```

比對結果顯示了一項破壞性變更，原因是 `allOf` 下新增了一個子 Schema。然而，這個新子 Schema 實際上並沒有增加任何新限制，因為它與先前既有的子 Schema 完全相同；因此在語意上這並非破壞性變更。  
您可以透過一般的檔案 diff 比對來驗證這點：
```bash
diff --side-by-side data/allof/simple.yaml data/allof/revision.yaml
```

為了減少此類誤判（False-positives），oasdiff 支援在比對規格前先將 `allOf` 替換為合併後的等價 Schema：

```bash
oasdiff breaking data/allof/simple.yaml data/allof/revision.yaml --flatten-allof
```
在這種情況下，oasdiff 能正確回報無任何破壞性變更。  
`--flatten-allof` 旗標亦支援 `diff`、`changelog` 與 `summary` 指令。

若要檢視 oasdiff 如何將 `allOf` 合併，可以使用專屬的 `flatten` 指令：
```bash
oasdiff flatten data/allof/simple.yaml
```

## 原始檔行號定位限制
使用 `--flatten-allof` 時，扁平化後的 Schema 中所偵測到的變更無法提供來源位置（檔案、行號、欄位）。這是因為合併後的 Schema 是新產生的結構，無法對應至原始檔案中的單一位置。

以下 Schema 欄位不會被合併：
- Extensions（擴充欄位）
- Example
- ExternalDocs
- AllowEmptyValue
- Deprecated
- XML
- Discriminator
- `if` / `then` / `else` (OpenAPI 3.1)
- `dependentSchemas` (OpenAPI 3.1)
- `unevaluatedProperties` / `unevaluatedItems` (OpenAPI 3.1)
- `patternProperties` (OpenAPI 3.1)
- `prefixItems` (OpenAPI 3.1)
- `contentSchema` (OpenAPI 3.1)
- `examples` (OpenAPI 3.1)
- `$id` / `$anchor` / `$dynamicRef` / `$dynamicAnchor` / `$schema` / `$comment` (OpenAPI 3.1)
- `$defs` (OpenAPI 3.1) — 在扁平化輸出中被刻意捨棄。`$defs` 是一個可重複使用 Schema 的命名空間，作為 `$ref` 指標的目標。在 `--flatten-allof` 執行完成後，合併後的 Schema 已無任何未解析的 `$ref`，因此該命名空間對語意已無貢獻。保留它只會增加雜訊（且當兩個 `allOf` 子 Schema 在相同的 `$defs` 鍵下定義了不同內容時，有靜默衝突的風險）。

上述除 `$defs` 外的所有欄位皆取自外部 Schema，因此沒有 `allOf` 的 Schema 在進出處理後維持不變。但若此類欄位出現在 `allOf` **子 Schema** 上，則會遺失：因為目前沒有合併兩個 `if` 或兩個 `unevaluatedProperties` 的規則，攜帶這些欄位的子 Schema 會將其丟失（[#878](https://github.com/oasdiff/oasdiff/issues/878)）。對於 3.1 關鍵字而言，這並非完全中性的損失，因為 `unevaluatedProperties: false` 會將封閉的 Schema 變成開放式，且條件分支是完整的驗證邏輯。

## 無效輸入處理

合併器對規格違規採取防禦性處理——會導致合併崩潰或產生無意義結果的數值會被靜默跳過，而非中斷回報。例如，等於零或負數的 `multipleOf` 值（在 OpenAPI / JSON Schema 規範中不被允許）會從合併結果中捨棄。若要在規格中抓出無效數值而非依賴合併器吞掉它們，請先單獨對規格執行 `validate`。

## 遞迴 Schema (Recursive schemas)

Schema 只能透過 `$ref` 表達遞迴，因此扁平化會確保所有循環皆能完整表達：

- 與額外條件合併的遞迴元件會將其遞迴保留為 `$ref`。當合併後的結果本身沒有名稱時（即在遞迴元件上進行行內 `allOf`），oasdiff 會根據被合併的元件名稱建構新名稱（如 `AllOfMerged_NodeA_NodeB`）將其加入 `components.schemas`，並將遞迴引用指向它。該名稱僅取決於合併的內容，因此相同 Schema 在 API 的各個修訂版本中都會保持相同名稱。
- 當遞迴引用在同一處與進一步的條件結合時，扁平化輸出會將它們保留為該具名引用與合併條件的 `allOf`。該特定位置會保持未完全展平，但不會遺失任何資訊。

例如，`NodeA` 具有遞迴性（`child` 指回 `NodeA`），且覆蓋層在 `child` 上增加了限制（`Plain` 為 `{type: object, properties: {leaf: {type: string}}}`）：

```yaml
tree:
  allOf:
    - $ref: '#/components/schemas/NodeA'
    - type: object
      properties:
        child:
          $ref: '#/components/schemas/Plain'
```

展平後，`tree` 本身的欄位被合併，而 `child` 同時要求遞迴結構與覆蓋層的限制：

```yaml
tree:
  type: object
  properties:
    child:
      allOf:
        - $ref: '#/components/schemas/AllOfMerged_NodeA'
        - type: object
          properties:
            leaf:
              type: string
```

此處保留 `allOf` 是因為 `child` 在語意上同時代表兩者，若將限制直接折疊至共用的遞迴元件中，將會改變它在其他所有地方的使用行為。

因此，展平後的文件可能會包含您未曾宣告的元件，以及圍繞遞迴點的殘留 `allOf` 項目；兩者皆是為了讓遞迴結構在扁平化後依然成立。

## 當子 Schema 不同時 `contains` 會過度受限 (OpenAPI 3.1)

`allOf` 對於 `contains` 的嚴格語意是 *「≥1 個陣列元素符合 X 且 ≥1 個元素符合 Y」*——這兩個符合的元素可以是不同的元素。我們無法將其表示為單一子 Schema，因此當多個 `allOf` 同級項目各自設定了 `contains` 時，oasdiff 會遞迴合併子 Schema（如同包在 `allOf` 中）並輸出單一 `contains: Merge(X, Y)`。這比原始規格**更嚴格**：它要求必須有一個項目同時滿足 X 與 Y。原始規格接受的陣列（X 與 Y 分別由不同項目滿足）在展平後可能會被拒絕；但原始規格拒絕的陣列絕不會被意外接受。在實務上這鮮少造成影響，因為多數規格在每個 `allOf` 鏈中最多只有一個 `contains`。

歡迎提供意見回饋與問題回報以協助我們改善此功能。
