# 設定檔 (Configuration Files)

[English](../CONFIG-FILES.md) | [繁體中文](CONFIG-FILES.md)

`oasdiff` 指令可從檔案中讀取設定。  
這對於複雜的參數設定或重複執行的使用情境非常方便。

## 預設搜尋順序

預設情況下，`oasdiff` 會在執行指令的目錄中依序尋找設定檔：

1. `.oasdiff.{json,yaml,yml,toml,hcl}` — 優先推薦（點開頭的 Dotfile 慣例）
2. `oasdiff.{json,yaml,yml,toml,hcl}` — 舊版向後相容備用檔名

新專案建議使用 `.oasdiff.*`。既有使用 `oasdiff.*` 的使用者無需遷移——兩種檔名皆可正常運作。

範例請參閱 [.oasdiff.yaml](../../examples/.oasdiff.yaml)。

## 明確指定覆寫

若要使用不同的檔名或路徑，可傳入 `--config <path>` 或設定 `OASDIFF_CONFIG` 環境變數。當兩者之一被設定時，系統會跳過預設搜尋，且指定路徑的檔案必須存在（若檔案不存在或格式損壞將回報錯誤）。

優先權順序：`--config <path>` > `OASDIFF_CONFIG` > 預設 `.oasdiff.*` 搜尋。

```sh
# 明確傳入旗標（單次執行）
oasdiff diff --config ./my-config.yaml base.yaml revision.yaml

# 環境變數（為 Shell 或 CI 工作流程設定一次）
export OASDIFF_CONFIG=./my-config.yaml
oasdiff diff base.yaml revision.yaml
```

設定檔支援命令列所支援的所有參數旗標，唯一例外為 `--open`。`--open` 會將比對資料上傳並在瀏覽器中開啟互動式左右對照審查介面，此為僅限命令列的互動操作，不會從設定檔讀取（設定檔通常會簽入版控並被共用，若支援此選項將導致每次執行皆自動上傳與開啟瀏覽器）。

附註：
1. 命令列旗標的優先權高於設定檔中的設定。
2. **布林值旗標**：在命令列上若要將布林旗標設為 `false`，必須使用 `=` 語法：`--flag=false`。若未使用 `=`，Cobra 解析器會將其後的文字當作位置引數：`--flag false` ≠ `--flag=false`。在設定檔中，請使用標準的 YAML/JSON 布林語法：`flag: false`。範例：`--allow-external-refs=false` (CLI) 與 `allow-external-refs: false` (設定檔)。
3. 部分旗標定義了指向其他設定檔的路徑：
    - `err-ignore`: 忽略錯誤（ERRORS）的設定檔
    - `severity-levels`: 自訂嚴重度層級的設定檔
    - `warn-ignore`: 忽略警告（WARNINGS）的設定檔
    - `template`: 產生變更日誌時使用的自訂 Go 範本檔案

   **這些旗標中的相對路徑是以設定檔所在的目錄進行解析**，而非當前工作目錄。因此當您在 `path/to/.oasdiff.yaml` 中撰寫 `err-ignore: rules.txt` 時，oasdiff 會讀取 `path/to/rules.txt`。絕對路徑與直接透過 CLI 旗標傳入的路徑則不會被重寫。

## 與 GitHub Action 搭配使用

[oasdiff-action](https://github.com/oasdiff/oasdiff-action)（包含 `breaking`、`changelog`、`diff`、`pr-comment`）會自動讀取儲存庫根目錄下的 `.oasdiff.yaml`，因此您在本機使用的相同設定檔亦能在 CI 中直接生效。Action 的 `with:` 輸入參數可覆寫設定檔中的數值。若需使用非預設路徑，可在 workflow 的 `env:` 區塊中設定 `OASDIFF_CONFIG`。

相關 Action 專屬細節與範例，請參閱 oasdiff-action README 中的 [Configuring with `.oasdiff.yaml`](https://github.com/oasdiff/oasdiff-action#configuring-with-oasdiffyaml) 區段。
