package logseqcontroller

import (
	"context"
	"fmt"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type InsertBlockTool struct {
	svc logseqsvc.BlockSvc
}

func NewInsertBlockTool(svc logseqsvc.BlockSvc) mcpext.ToolRegistrar {
	return &InsertBlockTool{svc: svc}
}

func (t *InsertBlockTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: insert_block")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "insert_block",
		Description: "Inserts a new block relative to a target block. position must be one of: before, after, child.",
	}, t.handle)
}

func (t *InsertBlockTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[InsertBlockInput],
) (*mcp.CallToolResultFor[any], error) {
	position := params.Arguments.Position
	if position != "before" && position != "after" && position != "child" {
		return nil, fmt.Errorf("invalid position %q: must be one of: before, after, child", position)
	}

	uuid, err := t.svc.InsertBlock(params.Arguments.TargetUUID, params.Arguments.Content, position)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: uuid}},
	}, nil
}

type InsertBlockInput struct {
	TargetUUID string `json:"target_uuid"`
	Content    string `json:"content"`
	Position   string `json:"position"` // "before" | "after" | "child"
}
