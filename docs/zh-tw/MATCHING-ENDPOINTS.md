# oasdiff 如何比對端點 (How oasdiff compares endpoints)

[English](../MATCHING-ENDPOINTS.md) | [繁體中文](MATCHING-ENDPOINTS.md)

Oasdiff 會將相互配對的端點（Matching endpoints）進行比較。  
預設情況下，比對演算法**會忽略**路徑參數的名稱。  
例如，以下兩個端點會被互相比較，因為它們僅在路徑參數名稱上有差異：
- GET /pet/{id}
- GET /pet/{petId}

此特性讓 oasdiff 即使在路徑參數被重新命名時，依然能夠正確比對相應的端點。

# 重複端點 (Duplicate Endpoints)
因為 oasdiff 會配對相應的端點進行比較，因此它預期在被比較的每份規格（或在[跨多檔案模式](COMPOSED.md)的集合）中，每個端點僅出現單一實例。

在某些情況下，您的規格可能包含重複的相符端點，這會導致 oasdiff 回報錯誤，例如：
```
✗ oasdiff diff data/duplicate_endpoints/base.yaml data/duplicate_endpoints/revision.yaml
Error: diff failed: duplicate endpoint: GET /pet/{petId2} in data/duplicate_endpoints/base.yaml and GET /pet/{petId3} in data/duplicate_endpoints/base.yaml. You may add the x-since-date extension to specify order, or include path parameter names in endpoint matching with --include-path-params
```

有兩種方式可以解決此問題：
1. 若重複是因為重命名路徑參數所造成，可加入 `--include-path-params` 旗標指示 oasdiff 在端點比對演算法中納入路徑參數名稱
2. 使用 [`x-since-date`](#重複端點與-x-since-date)

# 重複端點與 `x-since-date`
若在任一被比對的規格（或[跨多檔案模式](COMPOSED.md)中的規格集合）中發現重複的端點，oasdiff 會採用具有最新 `x-since-date` 數值的端點。

- `x-since-date` 擴充欄位可設定於 Path 或 Operation 層級。
- 設定於 Operation 層級的 `x-since-date` 會覆寫 Path 層級的值。
- 若端點未設定 `x-since-date` 擴充欄位，其數值預設為："2000-01-01"。
- 具有相同 `x-since-date` 數值的重複端點將會觸發錯誤。
- `x-since-date` 的格式必須符合 RFC3339 full-date 格式（例如 YYYY-MM-DD）。

`x-since-date` 使用範例：
```yaml
/api/test:
  get:
    x-since-date: "2023-01-11"
```
