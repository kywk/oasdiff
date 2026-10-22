# MCP 伺服器 (MCP server)

[English](../MCP.md) | [繁體中文](MCP.md)

oasdiff 提供了託管式的 [Model Context Protocol (MCP)](https://modelcontextprotocol.io) 伺服器，讓 Claude 與 Cursor 等 AI 助手能直接為您執行比對。詢問「這項變更是否會破壞 API？」時，AI 助手便能自動針對您的規格呼叫 oasdiff 並讀取分析結果。

此為基於 Streamable HTTP 傳輸協定的遠端伺服器，託管於 `https://api.oasdiff.com/mcp`。無需安裝，亦無需 API 金鑰。

## 支援工具 (Tools)

- 檢測兩個 OpenAPI 規格之間的破壞性變更 (detect breaking changes)
- 產生變更日誌 (generate a changelog)
- 比較兩份規格的差異 (diff two specs)
- 驗證規格的合法性 (validate a spec)

規格由 AI 助手以純文字方式傳入；外部 `$ref` URL 不會被解析，因此規格必須是自包含（Self-contained）的。

## 連線設定 (Connect)

Claude Code:

```bash
claude mcp add --transport http oasdiff https://api.oasdiff.com/mcp
```

Cursor，設定於 `~/.cursor/mcp.json`（全域）或 `.cursor/mcp.json`（個別專案）：

```json
{
  "mcpServers": {
    "oasdiff": {
      "url": "https://api.oasdiff.com/mcp"
    }
  }
}
```

任何支援 Streamable HTTP 的 MCP 客戶端皆可直接連線：將端點指向 `https://api.oasdiff.com/mcp` 即可。

完整指南：[oasdiff.com/docs/mcp](https://www.oasdiff.com/docs/mcp)
