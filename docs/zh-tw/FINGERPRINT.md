# 變更指紋 (Change Fingerprints)

[English](../FINGERPRINT.md) | [繁體中文](FINGERPRINT.md)

oasdiff 回報的每條變更日誌項目中皆包含一個 `fingerprint`（指紋）——由該變更的實質內容所衍生出的簡短且穩定的唯一識別碼。

## 何謂變更指紋

指紋是由以下內容計算出的 12 字元十六進位字串：

```
SHA256("{id}:{operation}:{path}:{args}")[:12]
```

其中 `id` 為檢查規則 ID（例如 `response-success-status-removed`），`operation` 為 HTTP 方法，`path` 為 API 路徑，`args` 則是建構變更訊息時所使用的參數值（以 `;` 分隔連接）。渲染後的文字本身不作為輸入，因此當變更日誌訊息潤飾文字或被翻譯成其他語言時，該變更的指紋仍維持完全不變。

## 為什麼需要指紋

變更指紋具備**跨 Commit 的穩定性**——在一個 PR 中，無論是由哪一個 Commit 引入了某項破壞性變更，該變更皆會獲得相同的指紋。這讓系統能夠實現：

- **跨 Commit 追蹤審查決策**：若審查者在 Commit A 批准了某個破壞性變更，當推送 Commit B 且該變更依然存在時，先前的批准狀態可以自動沿用。
- **多次執行結果去重**：比對多次不同執行的輸出結果時能輕鬆去除重複項。
- **在外部系統中精準引用特定變更**：在 CI、審查工具、稽核日誌中引用特定變更項目，而無需儲存冗長的變更文字。

## 當基準規格（Base）改變時

只要 Base 規格保持不變，指紋就會保持一致。當 Base 規格發生變更時（例如 PR 與更新後的 `main` 分支進行 Rebase 或 Merge），相同的變更可能會產生新的指紋，此時與舊指紋關聯的審查決策將不再匹配：

- **取自 Base 的參數**：某些訊息中包含來自 Base 規格的值。型別變更訊息為 `changed from integer to string`，因此若 Base 中的型別改變了，指紋也會隨之改變，儘管 PR 依然是將其改為 `string`。
- **相對位置**：非 `$ref` 的 `oneOf`、`anyOf` 或 `allOf` 分支是依其位置命名的（如 `oneOf[subschema #4]`）。若 Base 在其前方新增了一個分支，該分支內的變更將獲得新的路徑。
- **重新命名**：若 Base 重新命名了 API 路徑或屬性，其下的每個變更皆會獲得新指紋。
- **多屬性共用的 Schema**：此類 Schema 中的變更會回報在字母順序最前面的第一個屬性上。若 Base 新增了一個字母順序更靠前的屬性且同樣引用了該 Schema，變更將改為在該新屬性上回報。

在上述每種情況下，該變更會重新回到「未審查」狀態，避免沿用可能已不再適用的舊審查決策。

## 輸出格式

指紋會出現在 JSON 與 YAML 輸出中：

### JSON 格式 (`-f json`)

```json
{
  "id": "response-success-status-removed",
  "text": "removed the success response with status '200'",
  "level": 3,
  "operation": "GET",
  "path": "/users/{id}",
  "section": "paths",
  "fingerprint": "a3f8c21b9d04"
}
```

### YAML 格式 (`-f yaml`)

```yaml
- id: response-success-status-removed
  text: removed the success response with status '200'
  level: 3
  operation: GET
  path: /users/{id}
  section: paths
  fingerprint: a3f8c21b9d04
```

## 在 Go 程式中使用

當將 oasdiff 作為 Go 套件使用時，每個 `formatters.Change` 皆可直接存取該指紋：

```go
import (
    "github.com/oasdiff/oasdiff/formatters"
    "github.com/oasdiff/oasdiff/checker"
)

changes := formatters.NewChanges(checkerChanges, localizer)
for _, c := range changes {
    fmt.Printf("change %s fingerprint: %s\n", c.Id, c.Fingerprint)
}
```

## 碰撞機率

指紋為 12 個十六進位字元（48 位元）。對於包含數十個破壞性變更的一般 PR 而言，碰撞機率微乎其微。若極端罕見情況下兩項變更剛好共享同一個指紋，它們會被視為同一個變更——在真實世界的 API diff 中幾乎不可能發生。
