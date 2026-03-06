package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetPageTool struct {
	svc logseqsvc.PageSvc
}

func NewGetPageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &GetPageTool{svc: svc}
}

func (t *GetPageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: get_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_page",
		Description: "Fetches all blocks of a Logseq page by name. Returns exists:false if the page does not exist.",
	}, t.handle)
}

func (t *GetPageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetPageInput],
) (*mcp.CallToolResultFor[any], error) {
	blocks, exists, err := t.svc.GetPageBlocks(params.Arguments.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "page not found: " + params.Arguments.Name}},
		}, nil
	}

	var text string
	for _, block := range blocks {
		text += logseq.RenderTree(&block, 0)
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

type GetPageInput struct {
	Name string `json:"name"`
}
