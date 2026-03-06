package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type UpdateBlockTool struct {
	svc logseqsvc.BlockSvc
}

func NewUpdateBlockTool(svc logseqsvc.BlockSvc) mcpext.ToolRegistrar {
	return &UpdateBlockTool{svc: svc}
}

func (t *UpdateBlockTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: update_block")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_block",
		Description: "Replaces the content of a block by UUID.",
	}, t.handle)
}

func (t *UpdateBlockTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[UpdateBlockInput],
) (*mcp.CallToolResultFor[any], error) {
	if err := t.svc.UpdateBlock(params.Arguments.UUID, params.Arguments.Content); err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
	}, nil
}

type UpdateBlockInput struct {
	UUID    string `json:"uuid"`
	Content string `json:"content"`
}
