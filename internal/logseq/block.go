package logseq

import (
	"regexp"
	"strings"
)

type Block struct {
	UUID     string  `json:"uuid"`
	Content  string  `json:"content"`
	Level    int     `json:"level"`
	Children []Block `json:"children"`
}

var idPropRe = regexp.MustCompile(`\nid:: [a-f0-9-]+`)

func cleanContent(content string) string {
	return strings.TrimSpace(idPropRe.ReplaceAllString(content, ""))
}

func RenderTree(block *Block, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent) + "- "
	sb.WriteString(prefix + cleanContent(block.Content) + "\n")
	for _, child := range block.Children {
		sb.WriteString(RenderTree(&child, indent+1))
	}
	return sb.String()
}
