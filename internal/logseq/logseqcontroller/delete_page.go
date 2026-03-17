package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeletePageTool struct {
	svc logseqsvc.PageSvc
}

func NewDeletePageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &DeletePageTool{svc: svc}
}

func (t *DeletePageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: delete_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_page",
		Description: "Deletes a Logseq page by name. Returns an error if the page does not exist.",
	}, t.handle)
}

func (t *DeletePageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[DeletePageInput],
) (*mcp.CallToolResultFor[any], error) {
	if err := t.svc.DeletePage(params.Arguments.Name); err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
	}, nil
}

type DeletePageInput struct {
	Name string `json:"name"`
}
