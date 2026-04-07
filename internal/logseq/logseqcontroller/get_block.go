package logseqcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetBlockTool struct {
	blockSvc logseqsvc.BlockSvc
	pageSvc  logseqsvc.PageSvc
}

func NewGetBlockTool(blockSvc logseqsvc.BlockSvc, pageSvc logseqsvc.PageSvc) mcpext.ToolRegistrar {
	return &GetBlockTool{blockSvc: blockSvc, pageSvc: pageSvc}
}

func (t *GetBlockTool) Register(server *mcp.Server) {
	log.Println("Registering MCP tool: get_block")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_block",
		Description: "Gets a Logseq block by UUID. Returns JSON {page, outline, target_start, target_end}. outline is the full page text (same format as read_page). target_start/target_end are character offsets locating the block in the outline. To edit this block, use edit_page with page name and a substring from this outline as old_content.",
	}, t.handle)
}

type getBlockResult struct {
	Page        string `json:"page"`
	TargetStart int    `json:"target_start"`
	TargetEnd   int    `json:"target_end"`
	Outline     string `json:"outline"`
}

func (t *GetBlockTool) handle(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetBlockInput],
) (*mcp.CallToolResultFor[any], error) {
	block, err := t.blockSvc.GetBlock(params.Arguments.UUID)
	if err != nil {
		return nil, err
	}

	pageName := block.Page.Name
	pageOutline, exists, err := t.pageSvc.ReadPage(pageName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("page not found: %s", pageName)
	}

	// Render the target block (with its subtree) to locate it in the page outline
	targetText := logseq.RenderTree(block, 0)

	// Find the target subtree's position in the outline
	// RenderTree renders at indent 0, but the block may be nested in the page.
	// Search for the first line to determine the actual indent, then match the full subtree.
	targetFirstLine := strings.SplitN(strings.TrimRight(targetText, "\n"), "\n", 2)[0]
	targetStart := -1
	lines := strings.SplitN(pageOutline, "\n", -1)
	pos := 0
	for _, line := range lines {
		if targetStart == -1 && strings.TrimSpace(line) == strings.TrimSpace(targetFirstLine) {
			targetStart = pos
		}
		pos += len(line) + 1
	}

	// Compute target_end by finding the rendered subtree length at the correct indent
	targetEnd := targetStart
	if targetStart != -1 {
		// Re-render at the actual indent level found in the page
		actualIndent := 0
		if targetStart < len(pageOutline) {
			lineAtStart := pageOutline[targetStart:]
			if idx := strings.Index(lineAtStart, "- "); idx >= 0 {
				actualIndent = idx / 2
			}
		}
		renderedAtIndent := logseq.RenderTree(block, actualIndent)
		targetEnd = targetStart + len(renderedAtIndent)
	}

	result := getBlockResult{
		Page:        pageName,
		TargetStart: targetStart,
		TargetEnd:   targetEnd,
		Outline:     pageOutline,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}

type GetBlockInput struct {
	UUID string `json:"uuid"`
}
