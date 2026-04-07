package logseqsvc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

// --- outline builder (same DSL as controller tests) ---

func block(content string, children ...string) string {
	lines := "- " + content + "\n"
	for _, child := range children {
		for _, line := range strings.Split(strings.TrimRight(child, "\n"), "\n") {
			lines += "  " + line + "\n"
		}
	}
	return lines
}

func outline(blocks ...string) string {
	return strings.Join(blocks, "")
}

// blocksFromOutline builds a []logseq.Block tree from outline text with auto-generated UUIDs.
// UUID format: "u0", "u1", "u2", ... in DFS order.
func blocksFromOutline(text string) []logseq.Block {
	nodes, err := logseq.ParseOutline(text)
	if err != nil {
		panic(err)
	}
	counter := 0
	var convert func(nodes []*logseq.OutlineNode) []logseq.Block
	convert = func(nodes []*logseq.OutlineNode) []logseq.Block {
		var blocks []logseq.Block
		for _, n := range nodes {
			id := fmt.Sprintf("u%d", counter)
			counter++
			b := logseq.Block{
				UUID:     id,
				Content:  n.Content,
				Children: convert(n.Children),
			}
			blocks = append(blocks, b)
		}
		return blocks
	}
	return convert(nodes)
}

// --- Tests ---

func TestComputeEditPlan_replaceSiblingPreservesOrder(t *testing.T) {
	// Page:
	//   - Parent
	//     - First child    ← edit target
	//     - Second child
	//
	// Editing "First child" should anchor "before" Second child, not "child" of Parent.
	blocks := blocksFromOutline(outline(
		block("Parent",
			block("First child"),
			block("Second child"),
		),
	))

	// old/new content must match the rendered indent (First child is at indent 1)
	plan, err := computeEditPlan(blocks,
		"  - First child\n",
		"  - First child (edited)\n",
	)
	if err != nil {
		t.Fatal(err)
	}

	if plan.AnchorPosition != "before" {
		t.Errorf("anchor position = %q, want \"before\"", plan.AnchorPosition)
	}
	// Anchor should be "Second child" block
	if plan.AnchorUUID != blocks[0].Children[1].UUID {
		t.Errorf("anchor UUID = %q, want %q (Second child)", plan.AnchorUUID, blocks[0].Children[1].UUID)
	}
	if len(plan.RemoveUUIDs) != 1 {
		t.Fatalf("remove count = %d, want 1", len(plan.RemoveUUIDs))
	}
	if plan.RemoveUUIDs[0] != blocks[0].Children[0].UUID {
		t.Errorf("removed = %q, want %q (First child)", plan.RemoveUUIDs[0], blocks[0].Children[0].UUID)
	}
}

func TestComputeEditPlan_replaceLastChild(t *testing.T) {
	// Page:
	//   - Parent
	//     - First child
	//     - Last child    ← edit target
	//
	// No sibling after → anchor should be "First child" with "after".
	blocks := blocksFromOutline(outline(
		block("Parent",
			block("First child"),
			block("Last child"),
		),
	))

	plan, err := computeEditPlan(blocks,
		"  - Last child\n",
		"  - Last child (edited)\n",
	)
	if err != nil {
		t.Fatal(err)
	}

	if plan.AnchorPosition != "after" {
		t.Errorf("anchor position = %q, want \"after\"", plan.AnchorPosition)
	}
	if plan.AnchorUUID != blocks[0].Children[0].UUID {
		t.Errorf("anchor UUID = %q, want %q (First child)", plan.AnchorUUID, blocks[0].Children[0].UUID)
	}
}

func TestComputeEditPlan_replaceOnlyChild(t *testing.T) {
	// Page:
	//   - Parent
	//     - Only child    ← edit target
	//
	// No sibling before or after → anchor should be Parent with "child".
	blocks := blocksFromOutline(outline(
		block("Parent",
			block("Only child"),
		),
	))

	plan, err := computeEditPlan(blocks,
		"  - Only child\n",
		"  - Only child (edited)\n",
	)
	if err != nil {
		t.Fatal(err)
	}

	if plan.AnchorPosition != "child" {
		t.Errorf("anchor position = %q, want \"child\"", plan.AnchorPosition)
	}
	if plan.AnchorUUID != blocks[0].UUID {
		t.Errorf("anchor UUID = %q, want %q (Parent)", plan.AnchorUUID, blocks[0].UUID)
	}
}

func TestComputeEditPlan_replaceRootBlock(t *testing.T) {
	// Page:
	//   - Block A         ← edit target
	//   - Block B
	//
	// First root block → anchor should be "before" Block B.
	blocks := blocksFromOutline(outline(
		block("Block A"),
		block("Block B"),
	))

	plan, err := computeEditPlan(blocks,
		"- Block A\n",
		"- Block A (edited)\n",
	)
	if err != nil {
		t.Fatal(err)
	}

	if plan.AnchorPosition != "before" {
		t.Errorf("anchor position = %q, want \"before\"", plan.AnchorPosition)
	}
	if plan.AnchorUUID != blocks[1].UUID {
		t.Errorf("anchor UUID = %q, want %q (Block B)", plan.AnchorUUID, blocks[1].UUID)
	}
}

func TestComputeEditPlan_replaceWithSubtree(t *testing.T) {
	// Page:
	//   - Parent
	//     - Child          ← edit target (with its subtree)
	//       - Grandchild
	//     - Sibling
	//
	// Should remove only "Child" (grandchild removed automatically).
	blocks := blocksFromOutline(outline(
		block("Parent",
			block("Child",
				block("Grandchild"),
			),
			block("Sibling"),
		),
	))

	plan, err := computeEditPlan(blocks,
		"  - Child\n    - Grandchild\n",
		"  - Child (edited)\n",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.RemoveUUIDs) != 1 {
		t.Fatalf("remove count = %d, want 1 (parent removal handles grandchild)", len(plan.RemoveUUIDs))
	}
	if plan.AnchorPosition != "before" {
		t.Errorf("anchor position = %q, want \"before\"", plan.AnchorPosition)
	}
}

func TestComputeEditPlan_deleteBlocks(t *testing.T) {
	blocks := blocksFromOutline(outline(
		block("Keep"),
		block("Delete me"),
	))

	plan, err := computeEditPlan(blocks,
		"- Delete me\n",
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.RemoveUUIDs) != 1 {
		t.Fatalf("remove count = %d, want 1", len(plan.RemoveUUIDs))
	}
	if len(plan.NewNodes) != 0 {
		t.Errorf("new nodes = %d, want 0", len(plan.NewNodes))
	}
}

func TestComputeEditPlan_oldContentNotFound(t *testing.T) {
	blocks := blocksFromOutline(outline(
		block("Existing"),
	))

	_, err := computeEditPlan(blocks, "- Nonexistent\n", "- Replaced\n")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err)
	}
}
