package logseqcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListPagesTool struct {
	svc logseqsvc.PageSvc
}

func NewListPagesTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &ListPagesTool{svc: svc}
}

func (t *ListPagesTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: list_pages")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pages",
		Description: `Lists non-journal pages. "parent" controls scope: omit for all pages, null for root-only pages (no namespace), or a namespace name (e.g. "project") for its direct children. Each entry includes has_children to indicate whether further navigation is possible.`,
	}, t.handle)
}

func (t *ListPagesTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ListPagesInput],
) (*mcp.CallToolResultFor[any], error) {
	input := params.Arguments

	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	parent, err := input.parseParent()
	if err != nil {
		return nil, err
	}

	result, err := t.svc.ListPages(parent, limit, input.Offset)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d\n\n", result.Total)
	for _, entry := range result.Pages {
		sb.WriteString(formatPageEntry(entry))
		sb.WriteByte('\n')
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

func formatPageEntry(e logseq.PageEntry) string {
	if e.HasChildren {
		return e.Name + "/"
	}
	return e.Name
}

type ListPagesInput struct {
	// Parent: omit = all pages, null = root pages, "name" = children of that namespace.
	RawParent json.RawMessage `json:"parent,omitempty"`
	Limit     int             `json:"limit,omitempty"`
	Offset    int             `json:"offset,omitempty"`
}

// parseParent returns nil (all pages), &"" (root pages), or &"name" (namespace children).
func (i *ListPagesInput) parseParent() (*string, error) {
	if len(i.RawParent) == 0 {
		return nil, nil // omitted = all pages
	}
	if string(i.RawParent) == "null" {
		empty := ""
		return &empty, nil // JSON null = root pages only
	}
	var s string
	if err := json.Unmarshal(i.RawParent, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
