package logseqcontroller

import (
	"fmt"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

// --- outline builder ---

// block builds an outline node with optional children.
//
//	block("Parent",
//	    block("Child 1"),
//	    block("Child 2"),
//	)
func block(content string, children ...string) string {
	lines := "- " + content + "\n"
	for _, child := range children {
		for _, line := range strings.Split(strings.TrimRight(child, "\n"), "\n") {
			lines += "  " + line + "\n"
		}
	}
	return lines
}

// outline joins blocks into a full page outline string.
func outline(blocks ...string) string {
	return strings.Join(blocks, "")
}

// renderNodes converts OutlineNodes back to outline text for assertion.
func renderNodes(nodes []*logseq.OutlineNode) string {
	var sb strings.Builder
	var render func(nodes []*logseq.OutlineNode, indent int)
	render = func(nodes []*logseq.OutlineNode, indent int) {
		for _, n := range nodes {
			sb.WriteString(strings.Repeat("  ", indent))
			sb.WriteString("- " + n.Content + "\n")
			render(n.Children, indent+1)
		}
	}
	render(nodes, 0)
	return sb.String()
}

// --- mockPageSvc ---

// mockPageSvc is a simple in-memory page store.
// Tests observe state via m.pages after operations.
type mockPageSvc struct {
	pages map[string]string // name -> outline text
}

func newMockPageSvc() *mockPageSvc {
	return &mockPageSvc{pages: make(map[string]string)}
}

func (m *mockPageSvc) ReadPage(name string) (string, bool, error) {
	text, ok := m.pages[name]
	return text, ok, nil
}

func (m *mockPageSvc) WritePage(name string, nodes []*logseq.OutlineNode) error {
	m.pages[name] = renderNodes(nodes)
	return nil
}

func (m *mockPageSvc) EditPage(name string, oldContent string, newContent string) error {
	text, ok := m.pages[name]
	if !ok {
		return fmt.Errorf("page not found: %s", name)
	}
	if !strings.Contains(text, oldContent) {
		return fmt.Errorf("old_content not found in page")
	}
	m.pages[name] = strings.Replace(text, oldContent, newContent, 1)
	return nil
}

func (m *mockPageSvc) DeletePage(name string) error {
	if _, ok := m.pages[name]; !ok {
		return fmt.Errorf("page not found: %s", name)
	}
	delete(m.pages, name)
	return nil
}

func (m *mockPageSvc) SearchPages(query string, limit, offset int) (*logseq.PageResult, error) {
	var matched []string
	for name := range m.pages {
		if query == "" || strings.Contains(name, query) {
			matched = append(matched, name)
		}
	}
	return &logseq.PageResult{Total: len(matched), Pages: matched}, nil
}

func (m *mockPageSvc) ListJournalPages(startDate, endDate string, limit, offset int) (*logseq.JournalPageResult, error) {
	return &logseq.JournalPageResult{Total: 1, Pages: []string{"Apr 6th, 2026"}}, nil
}

// --- mockBlockSvc ---

type mockBlockSvc struct {
	blocks map[string]*logseq.Block
}

func newMockBlockSvc() *mockBlockSvc {
	return &mockBlockSvc{blocks: make(map[string]*logseq.Block)}
}

func (m *mockBlockSvc) GetBlock(uuid string) (*logseq.Block, error) {
	b, ok := m.blocks[uuid]
	if !ok {
		return nil, fmt.Errorf("block not found: %s", uuid)
	}
	return b, nil
}

func (m *mockBlockSvc) InsertBlock(targetUUID, content, position string) (string, error) {
	return "new-uuid", nil
}

func (m *mockBlockSvc) UpdateBlock(uuid, content string) error { return nil }
func (m *mockBlockSvc) RemoveBlock(uuid string) error         { return nil }
func (m *mockBlockSvc) InsertTree(parentUUID string, nodes []*logseq.OutlineNode) error {
	return nil
}

// --- mockSearchSvc ---

type mockSearchSvc struct{}

func (m *mockSearchSvc) SearchBlocks(query string, limit, offset int) (*logseq.SearchResult, error) {
	return &logseq.SearchResult{
		Total: 1,
		Blocks: []logseq.SearchBlock{
			{UUID: "u1", Content: "matched: " + query, PageName: "Test Page"},
		},
	}, nil
}
