# mcp-logseq

Logseq MCP server. Exposes Logseq graph data and editing capabilities as MCP tools.

## Tools

- `get_block` — fetches a block and its full child tree by UUID
- `list_pages` — lists non-journal pages; `parent` controls scope: omit for all, `null` for root-only, or a namespace name for its direct children; each entry includes `has_children`; supports pagination via `limit` and `offset`
- `list_journal_pages` — lists journal (daily note) pages, newest first; optionally filter by `start_date`/`end_date` in `YYYYMMDD` format; supports pagination via `limit` and `offset`
- `search_blocks` — searches blocks by keyword across all pages; supports pagination via `limit` and `offset`
- `get_page` — fetches all blocks of a page by name
- `create_page` — creates a new page with given content
- `insert_block` — inserts a new block relative to a target block (`before`, `after`, or `child`)
- `update_block` — replaces the content of a block by UUID

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
