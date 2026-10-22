# 路徑前綴修改 (Path Prefix Modification)

[English](../PATH-PREFIX.md) | [繁體中文](PATH-PREFIX.md)

有時候 API 的路徑前綴需要進行調整，例如在發布新版本時：
- `/api/v1/...`
- `/api/v2/...`

Oasdiff 允許透過移除（Stripping）與/或前綴附加（Prepending）路徑前綴來比較具備不同前綴的 API 規格。  
在上述範例中，您可以按如下方式比較檔案：
```bash
oasdiff diff original.yaml new.yaml --strip-prefix-base /api/v1 --prefix-base /api/v2
```
或是：
```bash
oasdiff diff original.yaml new.yaml --strip-prefix-base /api/v1 --strip-prefix-revision /api/v2
```
請注意，移除（Stripping）操作會優先於前綴附加（Prepending）操作執行。
