# mcp-logseq

Logseq MCP server. Retrieves Logseq blocks by UUID for use with `((uuid))` references.

## Tools

- `get_block` — fetches a block and its full child tree by UUID

## Setup

Logseq HTTP API must be enabled (Settings → Features → HTTP APIs).

```bash
go build -o build/mcp-logseq ./cmd/mcp-logseq/
```

Supports both stdio (default) and HTTP transports.

```bash
# stdio (default)
./build/mcp-logseq

# HTTP
./build/mcp-logseq -http=true -http-port=8080 -stdio=false
```

```json
{
  "mcpServers": {
    "logseq": {
      "command": "/path/to/build/mcp-logseq",
      "env": {
        "LOGSEQ_API_URL": "http://localhost:12315",
        "LOGSEQ_API_TOKEN": "your_token"
      }
    }
  }
}
```
