package logseq

import (
	"fmt"
	"strings"
)

type BlockPage struct {
	Name string `json:"name"`
}

type Block struct {
	UUID     string    `json:"uuid"`
	Content  string    `json:"content"`
	Level    int       `json:"level"`
	Children []Block   `json:"children"`
	Page     BlockPage `json:"page"`
}

type OutlineNode struct {
	Content  string
	Children []*OutlineNode
}

// CleanContent strips properties and trims whitespace from block content.
func CleanContent(content string) string {
	cleaned, _ := StripProperties(strings.TrimSpace(content))
	return cleaned
}

// RenderTree renders a block tree to outline text with properties stripped.
func RenderTree(block *Block, indent int) string {
	var sb strings.Builder
	renderBlock(&sb, block, indent)
	return sb.String()
}

func renderBlock(sb *strings.Builder, block *Block, indent int) {
	prefix := strings.Repeat("  ", indent) + "- "
	content := CleanContent(block.Content)
	if content == "" {
		sb.WriteString(prefix + "\n")
	} else {
		contentLines := strings.Split(content, "\n")
		sb.WriteString(prefix + contentLines[0] + "\n")
		continuation := strings.Repeat("  ", indent) + "  "
		for _, cl := range contentLines[1:] {
			sb.WriteString(continuation + cl + "\n")
		}
	}
	for _, child := range block.Children {
		renderBlock(sb, &child, indent+1)
	}
}

// BlockLineMapping tracks which block produced which lines in the rendered outline.
type BlockLineMapping struct {
	Block     *Block
	StartLine int // 0-indexed line number
	EndLine   int // exclusive
	Indent    int
	PropLines []string // extracted properties for restoration
}

// RenderTreeAnnotated renders blocks to outline text (properties stripped)
// and returns per-block line mappings.
func RenderTreeAnnotated(blocks []Block) (string, []BlockLineMapping) {
	var sb strings.Builder
	var mappings []BlockLineMapping
	lineNum := 0
	for i := range blocks {
		renderBlockAnnotated(&sb, &blocks[i], 0, &lineNum, &mappings)
	}
	return sb.String(), mappings
}

func renderBlockAnnotated(sb *strings.Builder, block *Block, indent int, lineNum *int, mappings *[]BlockLineMapping) {
	startLine := *lineNum
	prefix := strings.Repeat("  ", indent) + "- "
	cleaned, props := StripProperties(strings.TrimSpace(block.Content))

	if cleaned == "" {
		sb.WriteString(prefix + "\n")
		*lineNum++
	} else {
		contentLines := strings.Split(cleaned, "\n")
		sb.WriteString(prefix + contentLines[0] + "\n")
		*lineNum++
		continuation := strings.Repeat("  ", indent) + "  "
		for _, cl := range contentLines[1:] {
			sb.WriteString(continuation + cl + "\n")
			*lineNum++
		}
	}

	*mappings = append(*mappings, BlockLineMapping{
		Block:     block,
		StartLine: startLine,
		EndLine:   *lineNum,
		Indent:    indent,
		PropLines: props,
	})

	for i := range block.Children {
		renderBlockAnnotated(sb, &block.Children[i], indent+1, lineNum, mappings)
	}
}

// ParseOutline parses outline text (2-space indent, "- " prefix) into OutlineNodes.
func ParseOutline(text string) ([]*OutlineNode, error) {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	var roots []*OutlineNode
	type stackEntry struct {
		node  *OutlineNode
		level int
	}
	var stack []stackEntry
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed := strings.TrimLeft(line, " ")
		spaces := len(line) - len(trimmed)
		if !strings.HasPrefix(trimmed, "- ") {
			// Continuation line: append to the last node's content
			if len(stack) > 0 {
				last := stack[len(stack)-1].node
				last.Content += "\n" + trimmed
				continue
			}
			return nil, fmt.Errorf("line missing '- ' prefix: %q", line)
		}
		if spaces%2 != 0 {
			return nil, fmt.Errorf("invalid indent (odd spaces): %q", line)
		}
		level := spaces / 2
		content := trimmed[2:]
		node := &OutlineNode{Content: content}
		for len(stack) > 0 && stack[len(stack)-1].level >= level {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			parent := stack[len(stack)-1].node
			parent.Children = append(parent.Children, node)
		}
		stack = append(stack, stackEntry{node, level})
	}
	return roots, nil
}
