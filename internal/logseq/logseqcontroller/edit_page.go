package logseqcontroller

import (
	"context"
	"log"

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
		Description: "Find and replace text in a Logseq page. old_content must be an exact substring of the page outline (as returned by read_page or get_block). new_content is the replacement in the same outline format. Both can span multiple blocks. Indentation must match exactly.",
	}, t.handle)
}

func (t *EditPageTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[EditPageInput],
) (*mcp.CallToolResultFor[any], error) {
	if err := t.svc.EditPage(params.Arguments.Name, params.Arguments.OldContent, params.Arguments.NewContent); err != nil {
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
