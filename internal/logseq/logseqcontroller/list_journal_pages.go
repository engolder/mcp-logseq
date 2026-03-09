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

type ListJournalPagesTool struct {
	svc logseqsvc.PageSvc
}

func NewListJournalPagesTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &ListJournalPagesTool{svc: svc}
}

func (t *ListJournalPagesTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: list_journal_pages")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_journal_pages",
		Description: "Lists journal (daily note) pages, newest first. Optionally filter by date range using start_date and end_date in YYYYMMDD integer format (e.g. 20250101).",
	}, t.handle)
}

func (t *ListJournalPagesTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ListJournalPagesInput],
) (*mcp.CallToolResultFor[any], error) {
	input := params.Arguments

	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	result, err := t.svc.ListJournalPages(input.StartDate, input.EndDate, limit, input.Offset)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d\n\n", result.Total)
	sb.WriteString(strings.Join(result.Pages, "\n"))

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil
}

type ListJournalPagesInput struct {
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}
