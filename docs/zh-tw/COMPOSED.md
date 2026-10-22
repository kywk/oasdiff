# 跨多檔案比較 API (Compare APIs Split Across Multiple Files)

[English](../COMPOSED.md) | [繁體中文](COMPOSED.md)

Composed 模式（跨多檔案模式）可比較兩組 OpenAPI 規格集合，而非預設模式下的單一規格對。  
規格集合是使用萬用字元 [Glob](https://en.wikipedia.org/wiki/Glob_(programming)) 來指定的。  
當您的 API 定義分散在多個檔案中時（例如多個微服務各自擁有獨立的規格，但統一置於 API Gateway 背後公開），您可以使用此模式一次檢查所有規格的變更。

附註：  
1. Composed 模式僅比較[路徑與端點](DIFF.md#路徑-paths-與端點-endpoints)；其他資源（如 components）只有在被路徑或端點引用時才會進行比對。
2. Composed 模式不支援[路徑前綴修改](PATH-PREFIX.md)。
3. 包含星號（*）的 Glob 字串必須以引號包覆或進行跳脫，避免被 Shell 先行展開。

範例：
```bash
oasdiff breaking --composed "data/composed/base/*.yaml" "data/composed/revision/*.yaml"
```
