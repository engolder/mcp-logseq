package logseq

import (
	"fmt"
	"regexp"
	"strings"
)

type Block struct {
	UUID     string  `json:"uuid"`
	Content  string  `json:"content"`
	Level    int     `json:"level"`
	Children []Block `json:"children"`
}

type OutlineNode struct {
	Content  string
	Children []*OutlineNode
}

var idPropRe = regexp.MustCompile(`\nid:: [a-f0-9-]+`)

func CleanContent(content string) string {
	return strings.TrimSpace(idPropRe.ReplaceAllString(content, ""))
}

func RenderTree(block *Block, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent) + "- "
	sb.WriteString(prefix + CleanContent(block.Content) + "\n")
	for _, child := range block.Children {
		sb.WriteString(RenderTree(&child, indent+1))
	}
	return sb.String()
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
		if spaces%2 != 0 {
			return nil, fmt.Errorf("invalid indent (odd spaces): %q", line)
		}
		level := spaces / 2
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, fmt.Errorf("line missing '- ' prefix: %q", line)
		}
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
