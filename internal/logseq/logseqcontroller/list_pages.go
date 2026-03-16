package logseqcontroller

import (
	"context"
	"fmt"
	"log"
	"strings"

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
		Description: "Lists all non-journal pages. Namespace pages appear with their full name (e.g. \"project/sub\"). Supports pagination via limit and offset.",
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

	result, err := t.svc.ListPages(limit, input.Offset)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d\n\n", result.Total)
	for _, name := range result.Pages {
		sb.WriteString(name)
		sb.WriteByte('\n')
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

type ListPagesInput struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}
