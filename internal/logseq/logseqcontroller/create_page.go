package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CreatePageTool struct {
	svc logseqsvc.PageSvc
}

func NewCreatePageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &CreatePageTool{svc: svc}
}

func (t *CreatePageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: create_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_page",
		Description: "Creates a new Logseq page with the given content. Content is written as-is with no template applied.",
	}, t.handle)
}

func (t *CreatePageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[CreatePageInput],
) (*mcp.CallToolResultFor[any], error) {
	uuid, err := t.svc.CreatePage(params.Arguments.Name, params.Arguments.Content)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: uuid}},
	}, nil
}

type CreatePageInput struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
