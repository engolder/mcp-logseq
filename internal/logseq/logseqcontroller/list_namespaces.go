package logseqcontroller

import (
	"context"
	"log"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListNamespacesTool struct {
	svc logseqsvc.SearchSvc
}

func NewListNamespacesTool(svc logseqsvc.SearchSvc) mcpext.ToolRegistrar {
	return &ListNamespacesTool{svc: svc}
}

func (t *ListNamespacesTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: list_namespaces")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_namespaces",
		Description: "Lists available namespaces in the Logseq graph (e.g. journal, project). Use before search_blocks to determine search scope.",
	}, t.handle)
}

func (t *ListNamespacesTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ListNamespacesInput],
) (*mcp.CallToolResultFor[any], error) {
	namespaces, err := t.svc.ListNamespaces()
	if err != nil {
		return nil, err
	}

	text := strings.Join(namespaces, "\n")
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

type ListNamespacesInput struct{}
