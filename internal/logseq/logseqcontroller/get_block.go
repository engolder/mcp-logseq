package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetBlockTool struct {
	client *logseq.Client
}

func NewGetBlockTool(client *logseq.Client) mcpext.ToolRegistrar {
	return &GetBlockTool{client: client}
}

func (t *GetBlockTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: get_block")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_block",
		Description: "Get a Logseq block and its children by UUID. Use this for ((uuid)) block references.",
	}, t.handle)
}

func (t *GetBlockTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetBlockInput],
) (*mcp.CallToolResultFor[any], error) {
	block, err := t.client.GetBlock(params.Arguments.UUID)
	if err != nil {
		return nil, err
	}

	text := logseq.RenderTree(block, 0)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

type GetBlockInput struct {
	UUID string `json:"uuid" jsonschema:"Block UUID to retrieve (e.g. '69a9a12e-7ff3-4129-bf85-5801c1d76994')"`
}
