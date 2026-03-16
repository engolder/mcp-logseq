package logseqcontroller

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchTool struct {
	svc logseqsvc.SearchSvc
}

func NewSearchTool(svc logseqsvc.SearchSvc) mcpext.ToolRegistrar {
	return &SearchTool{svc: svc}
}

func (t *SearchTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: search")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Searches blocks by keyword across all pages. Returns page name and block content. Supports pagination via limit and offset.",
	}, t.handle)
}

func (t *SearchTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[SearchInput],
) (*mcp.CallToolResultFor[any], error) {
	input := params.Arguments
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	result, err := t.svc.SearchBlocks(input.Query, limit, input.Offset)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d\n", result.Total)
	for _, block := range result.Blocks {
		sb.WriteString("\n")
		if block.JournalDay != 0 {
			fmt.Fprintf(&sb, "[%s] (journal: %d)\n", block.PageName, block.JournalDay)
		} else {
			fmt.Fprintf(&sb, "[%s]\n", block.PageName)
		}
		fmt.Fprintf(&sb, "content: %s\n", logseq.CleanContent(block.Content))
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

type SearchInput struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
