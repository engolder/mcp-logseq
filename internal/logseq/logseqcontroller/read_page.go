package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ReadPageTool struct {
	svc logseqsvc.PageSvc
}

func NewReadPageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &ReadPageTool{svc: svc}
}

func (t *ReadPageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: read_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_page",
		Description: "Returns the full outline text of a Logseq page. Blocks are rendered with 2-space indents and '- ' prefixes.",
	}, t.handle)
}

func (t *ReadPageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ReadPageInput],
) (*mcp.CallToolResultFor[any], error) {
	text, exists, err := t.svc.ReadPage(params.Arguments.Name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "page not found: " + params.Arguments.Name}},
		}, nil
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

type ReadPageInput struct {
	Name string `json:"name"`
}
