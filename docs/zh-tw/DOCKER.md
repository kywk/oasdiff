# 透過 Docker 執行 oasdiff (Running oasdiff from docker)

[English](../DOCKER.md) | [繁體中文](DOCKER.md)

若要透過 Docker 執行 oasdiff，只需將 `oasdiff` 指令替換為 `docker run --rm -t tufin/oasdiff`（或您偏好的映像檔）。  
以下提供幾個常用範例：

## 使用 Docker 檢測破壞性變更
```bash
docker run --rm -t tufin/oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## 使用 Docker 比較本機檔案
```bash
docker run --rm -t -v $(pwd):/specs:ro -w /specs tufin/oasdiff changelog openapi-test1.yaml openapi-test3.yaml
```

將 `$(pwd)` 替換為包含您規格檔案的本機路徑。
