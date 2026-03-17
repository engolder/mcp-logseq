# mcp-logseq

Logseq MCP server. Exposes Logseq graph data and editing capabilities as MCP tools.

## Tools

- `delete_page` — deletes a page by name; returns an error if not found
- `read_page` — returns the full outline text of a page (2-space indent, `- ` prefix)
- `write_page` — overwrites a page with outline text, or creates it if new
- `edit_page` — finds a block by content and replaces its subtree; `old_content` must uniquely match one block
- `search_pages` — searches non-journal pages by name; omit query to list all pages; supports pagination via `limit` and `offset`
- `list_journal_pages` — lists journal (daily note) pages, newest first; optionally filter by `start_date`/`end_date` in `YYYYMMDD` format; supports pagination via `limit` and `offset`
- `search` — searches blocks by keyword across all pages; returns page name and content (no UUIDs); supports pagination via `limit` and `offset`

### Outline text format

```
- Block 1
  - Child 1
    - Grandchild
  - Child 2
- Block 2
```

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
