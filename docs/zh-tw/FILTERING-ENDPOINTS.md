# 篩選端點 (Filtering Endpoints)

[English](../FILTERING-ENDPOINTS.md) | [繁體中文](FILTERING-ENDPOINTS.md)

您可以透過兩種方式篩選端點：

## 依路徑名稱篩選
使用 `--match-path` 選項排除不符合指定正規表示式的路徑。  
例如，以下 diff 僅包含路徑中含有 "/api" 的端點：
```bash
oasdiff diff --match-path "/api" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

使用 `--unmatch-path` 選項排除符合指定正規表示式的路徑。  
例如，以下 diff 排除路徑中含有 "beta" 的端點：
```bash
oasdiff diff --unmatch-path "beta" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

附註：  
若某個路徑包含 Callback（回呼），此篩選條件會同時套用於路徑本身與 Callback 路徑。  
若要同時包含路徑與 Callback，請使用針對各層級進行過濾的正規表示式，例如 `"path|callback-path"`。
   
## 依擴充欄位篩選
使用 `--filter-extension` 選項排除擁有與指定正規表示式相符之 OpenAPI 擴充欄位（Extension）的路徑與操作。  
例如，以下 diff 排除擁有擴充欄位 "x-beta" 的路徑與操作：
```bash
oasdiff diff --filter-extension "x-beta" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```
附註：
1. OpenAPI 擴充欄位可同時定義於路徑層級與操作層級。使用此旗標時，兩者皆會被比對並排除。
2. 若某個路徑或操作僅在其中一份規格中具有相符的擴充欄位，而在另一份規格中沒有，它將以新增（Added）或刪除（Deleted）呈現。
