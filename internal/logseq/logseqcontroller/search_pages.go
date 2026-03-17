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

type SearchPagesTool struct {
	svc logseqsvc.PageSvc
}

func NewSearchPagesTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &SearchPagesTool{svc: svc}
}

func (t *SearchPagesTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: search_pages")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_pages",
		Description: "Searches non-journal pages by name (substring match). Returns only pages backed by an actual file (excludes auto-created tag/reference pages). Omit or leave query empty to list all pages. Supports pagination via limit and offset.",
	}, t.handle)
}

func (t *SearchPagesTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[SearchPagesInput],
) (*mcp.CallToolResultFor[any], error) {
	input := params.Arguments
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	result, err := t.svc.SearchPages(input.Query, limit, input.Offset)
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

type SearchPagesInput struct {
	Query  string `json:"query,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
