# 錯誤處理 (Errors)

[English](../ERRORS.md) | [繁體中文](ERRORS.md)

## 無效的規格檔案 (Invalid Specs)

Oasdiff 預期輸入合法的 OpenAPI 3 規格檔案。  
規格可以撰寫為 JSON 或 YAML 格式。  
當傳入無效的規格時，oasdiff 可能會回傳錯誤，例如：
```
Error: failed to load base spec from "spec.yaml": error converting YAML to JSON: yaml: line 2: mapping values are not allowed in this context
```
發生此錯誤的原因是底層使用的解析函式庫 [kin-openapi](https://github.com/getkin/kin-openapi) 在解析前會先將 YAML 規格轉換為 JSON。

當規格載入失敗時，oasdiff 會以代碼 `102` 結束（若傳入無效旗標為 `101`，若萬用字元 Glob 未比對到任何規格為 `103`）。

## 無效規格 vs. 規範違規 (Invalid specs vs. spec violations)

`diff`、`breaking`、`changelog` 與 `summary` 僅在無法載入或無法解析規格時才會中斷失敗（如上述的解析錯誤）。其餘情況下表現較為寬鬆：即使規格成功載入但在語法上違反了 OpenAPI 或 JSON Schema 的部分規則，仍然會進行比對。

若要檢查單一規格是否符合規範（例如無效的型別、缺少必填欄位、格式錯誤的正規表示式、無法解析的 `$ref` 等），請使用 [`oasdiff validate`](VALIDATE.md)。它會針對每項違規回報穩定的規則 ID、嚴重度層級與原始碼位置，並在規格完全無法載入時以代碼 102 結束。
