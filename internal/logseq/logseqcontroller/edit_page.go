package logseqcontroller

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EditPageTool struct {
	svc logseqsvc.PageSvc
}

func NewEditPageTool(svc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &EditPageTool{svc: svc}
}

func (t *EditPageTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: edit_page")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_page",
		Description: "Replaces a block subtree in a Logseq page. old_content is the plain block text to find (must be unique). new_content is the outline text replacement.",
	}, t.handle)
}

func (t *EditPageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[EditPageInput],
) (*mcp.CallToolResultFor[any], error) {
	newNodes, err := logseq.ParseOutline(params.Arguments.NewContent)
	if err != nil {
		return nil, err
	}
	if err := t.svc.EditPage(params.Arguments.Name, params.Arguments.OldContent, newNodes); err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
	}, nil
}

type EditPageInput struct {
	Name       string `json:"name"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
}
