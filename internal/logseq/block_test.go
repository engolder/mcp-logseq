package logseq

import (
	"testing"
)

func TestParseOutlineRoundTrip(t *testing.T) {
	input := "- Block 1\n  - Child 1\n    - Grandchild\n  - Child 2\n- Block 2\n"

	nodes, err := ParseOutline(input)
	if err != nil {
		t.Fatalf("ParseOutline error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(nodes))
	}

	// Reconstruct outline text from nodes
	var render func(nodes []*OutlineNode, indent int) string
	render = func(nodes []*OutlineNode, indent int) string {
		var out string
		for _, n := range nodes {
			prefix := ""
			for i := 0; i < indent; i++ {
				prefix += "  "
			}
			out += prefix + "- " + n.Content + "\n"
			out += render(n.Children, indent+1)
		}
		return out
	}

	got := render(nodes, 0)
	if got != input {
		t.Errorf("round-trip mismatch:\nwant: %q\ngot:  %q", input, got)
	}
}

func TestParseOutlineEmpty(t *testing.T) {
	nodes, err := ParseOutline("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestParseOutlineInvalidPrefix(t *testing.T) {
	_, err := ParseOutline("no prefix\n")
	if err == nil {
		t.Error("expected error for missing '- ' prefix")
	}
}

func TestParseOutlineOddIndent(t *testing.T) {
	_, err := ParseOutline("- root\n   - odd\n")
	if err == nil {
		t.Error("expected error for odd indentation")
	}
}

func TestParseOutlineContinuationLine(t *testing.T) {
	input := "- Block 1\n  continuation\n- Block 2\n"
	nodes, err := ParseOutline(input)
	if err != nil {
		t.Fatalf("ParseOutline error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Content != "Block 1\ncontinuation" {
		t.Errorf("expected multiline content, got %q", nodes[0].Content)
	}
}

func TestStripProperties(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantClean string
		wantProps []string
	}{
		{
			name:      "no properties",
			input:     "just text",
			wantClean: "just text",
			wantProps: nil,
		},
		{
			name:      "id property only",
			input:     "id:: 69d33009-8585-4782-9572-fbc3bfa4a37b",
			wantClean: "",
			wantProps: []string{"id:: 69d33009-8585-4782-9572-fbc3bfa4a37b"},
		},
		{
			name:      "text with properties",
			input:     "some text\nid:: abc-123\ncollapsed:: true",
			wantClean: "some text",
			wantProps: []string{"id:: abc-123", "collapsed:: true"},
		},
		{
			name:      "text between properties",
			input:     "line one\nid:: abc\nline two",
			wantClean: "line one\nline two",
			wantProps: []string{"id:: abc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned, props := StripProperties(tt.input)
			if cleaned != tt.wantClean {
				t.Errorf("cleaned = %q, want %q", cleaned, tt.wantClean)
			}
			if len(props) != len(tt.wantProps) {
				t.Errorf("props len = %d, want %d", len(props), len(tt.wantProps))
				return
			}
			for i, p := range props {
				if p != tt.wantProps[i] {
					t.Errorf("props[%d] = %q, want %q", i, p, tt.wantProps[i])
				}
			}
		})
	}
}

func TestRestoreProperties(t *testing.T) {
	content := "some text"
	props := []string{"id:: abc-123", "collapsed:: true"}
	got := RestoreProperties(content, props)
	want := "some text\nid:: abc-123\ncollapsed:: true"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRestorePropertiesEmpty(t *testing.T) {
	got := RestoreProperties("text", nil)
	if got != "text" {
		t.Errorf("got %q, want %q", got, "text")
	}
}

func TestRenderTreeAnnotated(t *testing.T) {
	blocks := []Block{
		{
			UUID:    "u1",
			Content: "Block A\nid:: aaa",
			Children: []Block{
				{UUID: "u2", Content: "Child 1\ncollapsed:: true"},
				{UUID: "u3", Content: "Child 2"},
			},
		},
		{UUID: "u4", Content: "Block B"},
	}

	rendered, mappings := RenderTreeAnnotated(blocks)

	// Properties should be stripped
	wantRendered := "- Block A\n  - Child 1\n  - Child 2\n- Block B\n"
	if rendered != wantRendered {
		t.Errorf("rendered:\ngot:  %q\nwant: %q", rendered, wantRendered)
	}

	if len(mappings) != 4 {
		t.Fatalf("expected 4 mappings, got %d", len(mappings))
	}

	// Block A: line 0
	if mappings[0].StartLine != 0 || mappings[0].EndLine != 1 {
		t.Errorf("Block A lines: %d-%d, want 0-1", mappings[0].StartLine, mappings[0].EndLine)
	}
	if len(mappings[0].PropLines) != 1 || mappings[0].PropLines[0] != "id:: aaa" {
		t.Errorf("Block A props: %v", mappings[0].PropLines)
	}

	// Child 1: line 1, has collapsed property
	if mappings[1].StartLine != 1 || mappings[1].EndLine != 2 {
		t.Errorf("Child 1 lines: %d-%d, want 1-2", mappings[1].StartLine, mappings[1].EndLine)
	}
	if len(mappings[1].PropLines) != 1 || mappings[1].PropLines[0] != "collapsed:: true" {
		t.Errorf("Child 1 props: %v", mappings[1].PropLines)
	}

	// Child 2: line 2
	if mappings[2].StartLine != 2 || mappings[2].EndLine != 3 {
		t.Errorf("Child 2 lines: %d-%d, want 2-3", mappings[2].StartLine, mappings[2].EndLine)
	}

	// Block B: line 3
	if mappings[3].StartLine != 3 || mappings[3].EndLine != 4 {
		t.Errorf("Block B lines: %d-%d, want 3-4", mappings[3].StartLine, mappings[3].EndLine)
	}
}

func TestRenderTreeStripsProperties(t *testing.T) {
	block := &Block{
		UUID:    "u1",
		Content: "some text\nid:: 69d33009-8585-4782-9572-fbc3bfa4a37b",
	}
	got := RenderTree(block, 0)
	want := "- some text\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTreePropertyOnlyBlock(t *testing.T) {
	block := &Block{
		UUID:    "u1",
		Content: "id:: 69d33009-8585-4782-9572-fbc3bfa4a37b",
	}
	got := RenderTree(block, 0)
	want := "- \n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
