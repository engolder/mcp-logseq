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
