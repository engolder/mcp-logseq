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
		Description: "Searches blocks by keyword across all pages. Supports pagination via limit and offset.",
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
		fmt.Fprintf(&sb, "uuid: %s\n", block.UUID)
		fmt.Fprintf(&sb, "content: %s\n", block.Content)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

type SearchBlocksInput struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
