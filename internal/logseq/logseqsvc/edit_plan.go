package logseqsvc

import (
	"fmt"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

// editPlan describes the block operations needed for an edit_page call.
type editPlan struct {
	// RemoveUUIDs are the top-level block UUIDs to remove (children removed automatically).
	RemoveUUIDs []string
	// AnchorUUID is the reference block for insertion.
	AnchorUUID string
	// AnchorPosition is "before", "after", or "child".
	AnchorPosition string
	// NewNodes are the parsed outline nodes to insert.
	NewNodes []*logseq.OutlineNode
}

// computeEditPlan determines which blocks to remove and where to insert new ones.
// This is a pure function that takes the block tree and edit content, enabling unit testing
// without Logseq API calls.
func computeEditPlan(blocks []logseq.Block, oldContent, newContent string) (*editPlan, error) {
	// Step 1: Render page with annotations
	rendered, mappings := logseq.RenderTreeAnnotated(blocks)

	// Step 2: Find oldContent as substring (must be unique)
	idx := strings.Index(rendered, oldContent)
	if idx == -1 {
		return nil, fmt.Errorf("old_content not found in page")
	}
	if strings.Contains(rendered[idx+1:], oldContent) {
		return nil, fmt.Errorf("old_content matches multiple locations")
	}

	// Step 3: Compute affected line range
	startLine := strings.Count(rendered[:idx], "\n")
	endLine := startLine + strings.Count(oldContent, "\n")
	if !strings.HasSuffix(oldContent, "\n") {
		endLine++
	}

	// Step 4: Identify affected blocks
	var affected []logseq.BlockLineMapping
	for _, m := range mappings {
		if m.StartLine < endLine && m.EndLine > startLine {
			affected = append(affected, m)
		}
	}
	if len(affected) == 0 {
		return nil, fmt.Errorf("no blocks found in affected range")
	}

	// Step 5: Find anchor
	anchorUUID, anchorPosition := findAnchor(mappings, affected)

	// Step 6: Determine which blocks to remove (top-level only)
	var removeUUIDs []string
	for _, m := range affected {
		isChildOfRemoved := false
		for _, other := range affected {
			if other.Block.UUID == m.Block.UUID {
				continue
			}
			if isDescendant(mappings, m.Block.UUID, other.Block.UUID) {
				isChildOfRemoved = true
				break
			}
		}
		if !isChildOfRemoved {
			removeUUIDs = append(removeUUIDs, m.Block.UUID)
		}
	}

	// Step 7: Parse newContent
	var newNodes []*logseq.OutlineNode
	trimmed := strings.TrimRight(newContent, "\n")
	if trimmed != "" {
		var err error
		newNodes, err = logseq.ParseOutline(trimmed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse new_content: %w", err)
		}
	}

	return &editPlan{
		RemoveUUIDs:    removeUUIDs,
		AnchorUUID:     anchorUUID,
		AnchorPosition: anchorPosition,
		NewNodes:       newNodes,
	}, nil
}

func findAnchor(mappings []logseq.BlockLineMapping, affected []logseq.BlockLineMapping) (string, string) {
	firstAffected := affected[0]
	lastAffected := affected[len(affected)-1]

	// Try: find first unaffected sibling AFTER affected region → insert "before"
	for i, m := range mappings {
		if m.Block.UUID == lastAffected.Block.UUID {
			for j := i + 1; j < len(mappings); j++ {
				next := mappings[j]
				if next.Indent <= lastAffected.Indent {
					if next.Indent == firstAffected.Indent {
						return next.Block.UUID, "before"
					}
					break
				}
			}
			break
		}
	}

	// Fallback: find block just before the first affected block
	for i, m := range mappings {
		if m.Block.UUID == firstAffected.Block.UUID {
			if i > 0 {
				prev := mappings[i-1]
				if prev.Indent < firstAffected.Indent {
					return prev.Block.UUID, "child"
				}
				return prev.Block.UUID, "after"
			}
			break
		}
	}

	return "", ""
}

// isDescendant checks if childUUID is a descendant of parentUUID in the mapping order.
func isDescendant(mappings []logseq.BlockLineMapping, childUUID, parentUUID string) bool {
	parentIdx := -1
	for i, m := range mappings {
		if m.Block.UUID == parentUUID {
			parentIdx = i
			break
		}
	}
	if parentIdx == -1 {
		return false
	}
	parentIndent := mappings[parentIdx].Indent
	for i := parentIdx + 1; i < len(mappings); i++ {
		if mappings[i].Indent <= parentIndent {
			break
		}
		if mappings[i].Block.UUID == childUUID {
			return true
		}
	}
	return false
}
