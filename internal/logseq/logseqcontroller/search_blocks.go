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

type SearchBlocksTool struct {
	svc logseqsvc.SearchSvc
}

func NewSearchBlocksTool(svc logseqsvc.SearchSvc) mcpext.ToolRegistrar {
	return &SearchBlocksTool{svc: svc}
}

func (t *SearchBlocksTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: search_blocks")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_blocks",
		Description: "Searches blocks by keyword and/or date range. Use list_namespaces first to determine available namespaces.",
	}, t.handle)
}

func (t *SearchBlocksTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[SearchBlocksInput],
) (*mcp.CallToolResultFor[any], error) {
	input := params.Arguments
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	result, err := t.svc.SearchBlocks(input.Query, input.Namespace, input.StartDate, input.EndDate, limit, input.Offset)
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
		fmt.Fprintf(&sb, "uuid: %s\n", block.UUID)
		fmt.Fprintf(&sb, "content: %s\n", block.Content)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

type SearchBlocksInput struct {
	Query     string `json:"query"`
	Namespace string `json:"namespace,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}
