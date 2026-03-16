package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WritePageTool struct {
	svc logseqsvc.PageSvc
}

func NewWritePageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &WritePageTool{svc: svc}
}

func (t *WritePageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: write_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_page",
		Description: "Overwrites the full content of a Logseq page (creates it if new). content is outline text with 2-space indents and '- ' prefixes.",
	}, t.handle)
}

func (t *WritePageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[WritePageInput],
) (*mcp.CallToolResultFor[any], error) {
	nodes, err := logseq.ParseOutline(params.Arguments.Content)
	if err != nil {
		return nil, err
	}
	if err := t.svc.WritePage(params.Arguments.Name, nodes); err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
	}, nil
}

type WritePageInput struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
