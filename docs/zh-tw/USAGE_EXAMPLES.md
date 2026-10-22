# 使用範例 (Usage Examples)

[English](../USAGE_EXAMPLES.md) | [繁體中文](USAGE_EXAMPLES.md)

## 本機 YAML 檔案的 OpenAPI Diff
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml
```
預設 diff 輸出格式為 YAML。  
沒有任何輸出表示 diff 為空，亦即兩份規格之間沒有任何變更。

## 本機檔案的 Text/Markdown 格式 OpenAPI Diff
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f text
```
Text/Markdown diff 報告提供簡化且局部的變更檢視。  
若需檢視完整細節，請使用預設格式：YAML。  
若希望在 text/markdown 報告中加入額外細節，歡迎提交 [Feature Request](https://github.com/oasdiff/oasdiff/issues/new?assignees=&labels=&template=feature_request.md&title=)。

## 本機檔案的 HTML 格式 OpenAPI Diff
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f html 
```
HTML diff 報告提供簡化且局部的變更檢視。  
若需檢視完整細節，請使用預設格式：YAML。  
若希望在 HTML 報告中加入額外細節，歡迎提交 [Feature Request](https://github.com/oasdiff/oasdiff/issues/new?assignees=&labels=&template=feature_request.md&title=)。

## 透過 HTTP/S 比較遠端檔案
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

## 檢測破壞性變更 (Breaking changes)
```bash
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 以 YAML 格式輸出破壞性變更
```bash
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f yaml
```

## 使用萬用字元（Glob）跨多個規格檢測破壞性變更
```bash
oasdiff breaking "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c
```

## 若發現 ERR 層級破壞性變更則以代碼 1 失敗
```bash
oasdiff breaking "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c -o ERR
```

## 若發現任何變更則以代碼 1 失敗
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text -o
```

## OpenAPI 變更日誌 (Changelog)
```bash
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 使用自訂範本輸出 OpenAPI 變更日誌
```bash
oasdiff changelog base.yaml revision.yaml --template my-template.md -f markdown
oasdiff changelog base.yaml revision.yaml --template my-template.html -f html
```
完整的說明與範例請參閱 [CHANGELOG-TEMPLATE.md](CHANGELOG-TEMPLATE.md)。

## 在瀏覽器中進行左右對照審查 (`--open`)
```bash
oasdiff changelog HEAD~1:openapi.yaml HEAD:openapi.yaml --open
oasdiff breaking HEAD~1:openapi.yaml HEAD:openapi.yaml --open
```
在印出變更日誌後，此指令會在您的本機加密比對內容，僅將加密密文上傳至 oasdiff.com，並於瀏覽器中開啟呈現左右對照的視覺化審查介面。無需登入：解密金鑰保留在審查連結的 `#` 錨點部分，絕不傳送至伺服器。連結有效期限為 7 天且可分享給任何人（金鑰包含於網址中）。完整工作流程請參閱 [BREAKING-CHANGES.md](BREAKING-CHANGES.md#在瀏覽器中進行左右對照審查---open)。

## 篩選路徑中包含 "/api" 的端點 Diff
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text -p "/api"
```
篩選條件會遞迴套用於所有層級。例如若某個路徑包含 [Callback](https://swagger.io/docs/specification/callbacks/)，篩選條件會同時套用於路徑本身與 Callback 路徑。若要包含此種嵌套變更，請使用包含兩者的正規表示式，例如 `--match-path "path|callback-path"`。

## 排除包含擴充欄位 "x-beta" 的路徑與操作
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text --filter-extension "x-beta"
``` 
附註：
1. [OpenAPI 擴充欄位](https://swagger.io/docs/specification/openapi-extensions/) 可以定義於 [路徑層級（path）](https://swagger.io/docs/specification/paths-and-operations/) 或 [操作層級（operation）](https://swagger.io/docs/specification/paths-and-operations/)。兩者皆會被此旗標比對並排除。
2. 若某個路徑或操作僅在其中一份規格中具備相符的擴充欄位，而在另一份中沒有，它將以新增（Added）或刪除（Deleted）呈現。

## 忽略描述 (description) 與範例 (examples) 的變更
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test3.yaml --exclude-elements description,examples -f text
``` 

## 顯示變更摘要統計 (Summary)
```bash
oasdiff summary https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 使用 Docker 執行 OpenAPI Diff
使用 Docker 執行時，只需將 `oasdiff` 指令替換為 `docker run --rm -t tufin/oasdiff`，例如：

```bash
docker run --rm -t tufin/oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

## 使用 Docker 檢測破壞性變更
```bash
docker run --rm -t tufin/oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 使用 Docker 比較本機檔案
```bash
docker run --rm -t -v $(pwd)/data:/data:ro tufin/oasdiff diff /data/openapi-test1.yaml /data/openapi-test3.yaml
```

將 `$(pwd)/data` 替換為包含您檔案的本機目錄路徑。  
請注意容器內的規格路徑必須以 `/` 開頭。

## 從腳本或 CI 傳遞參數

當 `base` 與 `revision` 來自變數而非手動輸入時，請在所有旗標後方加上 `--`，以確保變數值始終被當作位置引數解析，絕不會被誤判為旗標：

```bash
oasdiff breaking $flags -- "$base" "$revision"
```

請注意順序：`--` 必須位在所有旗標**之後**。`--` 之後的所有內容都會被視為位置引數，因此任何放在 `--` 後面的旗標都會被當作規格路徑讀取。

此規則適用於所有輸入類型——本機檔案、URL 或 Git 修訂版本皆然。如果沒有使用 `--`，當變數值以 `-` 開頭時會被解析為 oasdiff 旗標；若剛好與某個旗標名稱相符，將會意外改變指令的行為。
