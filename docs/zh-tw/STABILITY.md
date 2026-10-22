# API 穩定度等級 (API Stability Levels)

[English](../STABILITY.md) | [繁體中文](STABILITY.md)

當推出新的 API 時，您可能希望允許開發者自由調整其行為，而不觸發破壞性變更錯誤。  
您可以使用 `x-stability-level` 擴充欄位來定義端點的穩定度等級。  
共分為四個穩定度等級：`draft`（草案）-> `alpha` -> `beta` -> `stable`（穩定）。

### 穩定度門檻 (`--stability-level`)
預設情況下，oasdiff 採用 **beta** 作為門檻：標記為 `draft` 或 `alpha` 的端點會被排除在破壞性變更檢查之外，而 `beta` 與 `stable` 的端點則會進行檢查。

您可以透過 `--stability-level` 旗標覆寫此門檻：

| 旗標設定值 | 納入檢查的端點等級 |
|---|---|
| `draft` | 全部（`draft`, `alpha`, `beta`, `stable`） |
| `alpha` | `alpha`, `beta`, `stable` |
| `beta` *(預設值)* | `beta`, `stable` |
| `stable` | 僅限 `stable` |

範例：
```bash
# 在破壞性變更檢查中納入 draft 與 alpha 端點
oasdiff changelog base.yaml revision.yaml --stability-level draft
```

**未標記** `x-stability-level` 的端點預設視為 `stable`，無論設定的門檻為何皆始終納入檢查。

無效的數值（如 `--stability-level=banana`）會在解析命令列旗標時被直接拒絕。

### 雙向穩定度變更偵測
oasdiff 能**雙向**偵測端點 `x-stability-level` 的變更：

- **降級**（`stable`→`beta`, `beta`→`alpha` 等）— 回報為 `api-stability-decreased`
- **晉級**（`draft`→`alpha`, `alpha`→`beta` 等）— 回報為 `api-stability-increased`

相同的偵測機制亦套用於請求與回應的屬性：
- `request-property-stability-decreased` / `request-property-stability-increased`
- `response-property-stability-decreased` / `response-property-stability-increased`

只有當基準穩定度（原本離開的等級）符合設定的門檻時，這些變更才會被回報。從受檢等級降級會被回報，而原本就低於門檻的元素發生等級變動則不會回報。

### 程式化使用
將 `WithStabilityLevel` 作為選項傳入 `NewConfig`：
```go
config := checker.NewConfig(
    checker.GetAllChecks(),
    checker.WithOptionalChecks(optionalChecks),
    checker.WithStabilityLevel("alpha"),
)
```

### 設定範例
```yaml
/api/test:
  post:
    x-stability-level: "alpha"
```
