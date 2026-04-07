package logseqcontroller

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReadPage_returnsOutline(t *testing.T) {
	svc := newMockPageSvc()
	svc.pages["My Notes"] = outline(
		block("Block 1",
			block("Child"),
		),
		block("Block 2"),
	)

	tool := &ReadPageTool{svc: svc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[ReadPageInput]{
		Arguments: ReadPageInput{Name: "My Notes"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	want := outline(
		block("Block 1",
			block("Child"),
		),
		block("Block 2"),
	)
	if text != want {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", text, want)
	}
}

func TestReadPage_notFound(t *testing.T) {
	svc := newMockPageSvc()

	tool := &ReadPageTool{svc: svc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[ReadPageInput]{
		Arguments: ReadPageInput{Name: "Nonexistent"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "page not found") {
		t.Errorf("expected 'page not found', got: %s", text)
	}
}

func TestWritePage_createsPage(t *testing.T) {
	svc := newMockPageSvc()

	tool := &WritePageTool{svc: svc}
	_, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[WritePageInput]{
		Arguments: WritePageInput{
			Name: "New Page",
			Content: outline(
				block("Hello",
					block("World"),
				),
			),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := outline(
		block("Hello",
			block("World"),
		),
	)
	if svc.pages["New Page"] != want {
		t.Errorf("page snapshot:\ngot:  %q\nwant: %q", svc.pages["New Page"], want)
	}
}

func TestEditPage_findAndReplace(t *testing.T) {
	svc := newMockPageSvc()
	svc.pages["Daily"] = outline(
		block("Task A"),
		block("Task B",
			block("Detail"),
		),
		block("Task C"),
	)

	tool := &EditPageTool{svc: svc}
	_, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[EditPageInput]{
		Arguments: EditPageInput{
			Name: "Daily",
			OldContent: outline(
				block("Task B",
					block("Detail"),
				),
			),
			NewContent: outline(
				block("Task B (done)",
					block("Detail"),
					block("Added note"),
				),
			),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := outline(
		block("Task A"),
		block("Task B (done)",
			block("Detail"),
			block("Added note"),
		),
		block("Task C"),
	)
	if svc.pages["Daily"] != want {
		t.Errorf("page snapshot:\ngot:  %q\nwant: %q", svc.pages["Daily"], want)
	}
}

func TestEditPage_notFoundError(t *testing.T) {
	svc := newMockPageSvc()
	svc.pages["Daily"] = outline(
		block("Task A"),
	)

	tool := &EditPageTool{svc: svc}
	_, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[EditPageInput]{
		Arguments: EditPageInput{
			Name:       "Daily",
			OldContent: outline(block("Nonexistent")),
			NewContent: outline(block("Replaced")),
		},
	})
	if err == nil {
		t.Fatal("expected error for non-matching old_content")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %s", err)
	}
}

func TestDeletePage_removesPage(t *testing.T) {
	svc := newMockPageSvc()
	svc.pages["Old Page"] = outline(block("content"))

	tool := &DeletePageTool{svc: svc}
	_, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[DeletePageInput]{
		Arguments: DeletePageInput{Name: "Old Page"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := svc.pages["Old Page"]; exists {
		t.Error("expected page to be deleted")
	}
}

func TestDeletePage_notFound(t *testing.T) {
	svc := newMockPageSvc()

	tool := &DeletePageTool{svc: svc}
	_, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[DeletePageInput]{
		Arguments: DeletePageInput{Name: "Ghost"},
	})
	if err == nil {
		t.Fatal("expected error for nonexistent page")
	}
}

func TestSearchPages_filtersAndFormats(t *testing.T) {
	svc := newMockPageSvc()
	svc.pages["Go Notes"] = ""
	svc.pages["Rust Notes"] = ""
	svc.pages["Daily Log"] = ""

	tool := &SearchPagesTool{svc: svc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[SearchPagesInput]{
		Arguments: SearchPagesInput{Query: "Notes"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Total: 2") {
		t.Errorf("expected Total: 2, got:\n%s", text)
	}
	if !strings.Contains(text, "Go Notes") || !strings.Contains(text, "Rust Notes") {
		t.Errorf("expected matching pages, got:\n%s", text)
	}
}

func TestListJournalPages_formats(t *testing.T) {
	svc := newMockPageSvc()

	tool := &ListJournalPagesTool{svc: svc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[ListJournalPagesInput]{
		Arguments: ListJournalPagesInput{Limit: 10},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Apr 6th, 2026") {
		t.Errorf("expected journal page, got:\n%s", text)
	}
}

func TestSearch_formatsResults(t *testing.T) {
	svc := &mockSearchSvc{}

	tool := &SearchTool{svc: svc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[SearchInput]{
		Arguments: SearchInput{Query: "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Total: 1") {
		t.Errorf("expected Total: 1, got:\n%s", text)
	}
	if !strings.Contains(text, "[Test Page]") {
		t.Errorf("expected page name, got:\n%s", text)
	}
	if !strings.Contains(text, "matched: hello") {
		t.Errorf("expected block content, got:\n%s", text)
	}
}

func TestGetBlock_returnsStructuredJSON(t *testing.T) {
	pageSvc := newMockPageSvc()
	pageSvc.pages["My Notes"] = outline(
		block("Block A"),
		block("Target Block",
			block("Child"),
		),
		block("Block C"),
	)

	blockSvc := newMockBlockSvc()
	blockSvc.blocks["target-uuid"] = &logseq.Block{
		UUID:    "target-uuid",
		Content: "Target Block",
		Page:    logseq.BlockPage{Name: "My Notes"},
		Children: []logseq.Block{
			{UUID: "child-uuid", Content: "Child"},
		},
	}

	tool := &GetBlockTool{blockSvc: blockSvc, pageSvc: pageSvc}
	result, err := tool.handle(context.Background(), nil, &mcp.CallToolParamsFor[GetBlockInput]{
		Arguments: GetBlockInput{UUID: "target-uuid"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var got getBlockResult
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("expected JSON, got: %s", text)
	}

	if got.Page != "My Notes" {
		t.Errorf("page = %q, want %q", got.Page, "My Notes")
	}

	wantOutline := outline(
		block("Block A"),
		block("Target Block",
			block("Child"),
		),
		block("Block C"),
	)
	if got.Outline != wantOutline {
		t.Errorf("outline mismatch:\ngot:  %q\nwant: %q", got.Outline, wantOutline)
	}

	// target_start/target_end should point to the "- Target Block\n  - Child\n" substring
	targetSubstr := got.Outline[got.TargetStart:got.TargetEnd]
	wantTarget := outline(
		block("Target Block",
			block("Child"),
		),
	)
	if targetSubstr != wantTarget {
		t.Errorf("outline[target_start:target_end]:\ngot:  %q\nwant: %q", targetSubstr, wantTarget)
	}
}
