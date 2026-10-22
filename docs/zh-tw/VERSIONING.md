# 版本號遞增與破壞性變更 (Version Bumps and Breaking Changes)

[English](../VERSIONING.md) | [繁體中文](VERSIONING.md)

若您採用[語意化版本（Semantic Versioning, SemVer）](https://semver.org)管理 API 版本，破壞性變更理應伴隨著主版本號（Major Version）的遞增。  
Oasdiff 能夠在未遞增主版本號時主動提出警示，因為它同時掌握了比對的兩側資訊：變更的嚴重度，以及兩側規格中的 `info.version`。

## 檢查規則

此原則的核心只有一條規則：**破壞性變更必須遞增主版本號**。違反此原則的情形有三種，各自擁有獨立的檢查 ID：

| 檢查 ID | 回報情境（當偵測到破壞性變更且：） |
|---|---|
| `api-version-not-bumped` | `info.version` 完全未變更 |
| `api-version-decreased` | `info.version` 版本號倒退，例如從 `2.0.0` 降為 `1.0.0` |
| `api-major-version-not-bumped` | `info.version` 有變更，但主版本號未增加，例如從 `1.0.0` 改為 `1.1.0` |

輸出範例：
```
info    [api-major-version-not-bumped]
        in info
                a breaking change was detected but the major version did not increase, from '1.0.0' to '1.1.0'
```

這些規則的預設嚴重度層級為 `INFO`，因此預設情況下絕不會使 CI 建置失敗。

## 強制執行版本原則

若希望在違反原則時阻斷 CI 建置，可透過[自訂嚴重度層級](BREAKING-CHANGES.md#自訂嚴重度層級)將這些檢查提升至 `ERR`：
```bash
oasdiff breaking base.yaml revision.yaml --severity-levels oasdiff-levels.txt --fail-on ERR
```
其中 `oasdiff-levels.txt` 檔案內容如下：
```
api-version-not-bumped          err
api-version-decreased           err
api-major-version-not-bumped    err
```

## 關閉此原則檢查

若您完全不希望進行這些檢查，可在相同的設定檔中將它們設為 `none`：
```
api-version-not-bumped          none
api-version-decreased           none
api-major-version-not-bumped    none
```

## 不回報檢查的情境

只有在同時滿足以下兩個條件時，才會回報檢查結果：

1. **確實偵測到了破壞性變更**（依您設定的嚴重度等級為準）。若您已將某個檢查規則降級為 `INFO`，在此處它也不會被算作破壞性變更。沒有破壞性變更的發行版本絕不會被要求遞增主版本號。
2. **兩份規格的版本號皆符合語意化版本（SemVer）。** `1.0`、`v1`、`2026-06-01` 等其他版本格式沒有所謂的「主版本號」可供比對，因此會被直接跳過而非胡亂猜測。開頭帶有 `v` 是被允許的（如 `v1.2.3`）。先行編譯標籤（Prerelease）與建置中繼資料會被接受並忽略，因此 `1.1.0-rc.1` 會被視為次版本號（Minor）遞增。

在 `1.0.0` 以下的版本，SemVer 將次版本號（Minor）視為主版本號角色，因此從 `0.1.0` 升級至 `0.2.0` 允許包含破壞性變更，而從 `0.1.0` 升至 `0.1.1` 則不允許。

## 不在此檢查範圍內的項目
- 非破壞性變更應遞增哪一種版本號（只檢查破壞性變更）。
- 任何非 SemVer 的版本命名機制，包括 URL 路徑中所包含的版本號。
