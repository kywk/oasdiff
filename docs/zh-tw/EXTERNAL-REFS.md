# 安全性 (Security)

[English](../EXTERNAL-REFS.md) | [繁體中文](EXTERNAL-REFS.md)

## 透過外部 $ref 發動的 SSRF 攻擊

OpenAPI 規格中可以包含指向外部 URL 或檔案的 `$ref`。  
預設情況下，oasdiff 會追蹤並載入這些外部引用以完整解析規格。

當 oasdiff 處理**不可信任的規格**時（例如在 CI 流程中處理外部貢獻者提交的 Pull Request），這可能會帶來安全性風險。  
惡意的規格可能會包含如下的 `$ref`：

```yaml
$ref: 'http://169.254.169.254/latest/meta-data/iam/security-credentials/'
```

當 oasdiff 在雲端託管的 CI 環境中對此規格執行比對時，它會發出請求獲取該 URL，進而可能外洩雲端虛擬機實例的中繼資料與 IAM 憑證。

## 防範緩解措施

使用 `--allow-external-refs=false` 可阻止 oasdiff 抓取規格中所引用的任何外部 URL 或檔案：

```bash
oasdiff breaking base.yaml revision.yaml --allow-external-refs=false
```

附註：在命令列中將布林旗標設為 `false` 時必須使用 `=` 語法——`--allow-external-refs false` 無法如預期運作（請參閱[設定檔](CONFIG-FILES.md)說明）。

或者，亦可在 [oasdiff 設定檔](CONFIG-FILES.md)中進行設定：
```yaml
allow-external-refs: false
```

此旗標適用於 `diff`、`breaking`、`changelog`、`summary` 與 `flatten` 指令。

當外部引用被停用時，若規格中包含外部 `$ref`，oasdiff 會直接回傳錯誤，而不會向外發出網路請求。
結束代碼為 123，CI 步驟可据此區分被拒的 `$ref` 與其他載入失敗。

## CI 環境安全建議

若您的 CI 管線會針對來自不可信任來源的規格執行 oasdiff（例如外部社群提交的 PR），建議：

1. 若規格皆為自包含（Self-contained），請啟用 `--allow-external-refs=false`
2. 在隔離的沙盒環境中執行 oasdiff，禁止存取雲端中繼資料端點（如 169.254.169.254）或內部網路服務
