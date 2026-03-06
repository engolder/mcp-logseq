package main

import (
	"fmt"
	"log"

	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
)

func main() {
	client := logseq.NewClient()
	searchSvc := logseqsvc.NewSearchSvc(client)
	blockSvc := logseqsvc.NewBlockSvc(client)
	pageSvc := logseqsvc.NewPageSvc(client)

	// list_namespaces
	fmt.Println("=== list_namespaces ===")
	namespaces, err := searchSvc.ListNamespaces()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(namespaces)

	// search_blocks
	fmt.Println("\n=== search_blocks: toss-pos-bridge in journal ===")
	result, err := searchSvc.SearchBlocks("toss-pos-bridge", "journal", "", "", 3, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total: %d\n", result.Total)
	for _, b := range result.Blocks {
		fmt.Printf("[%s] (journal: %d)\nuuid: %s\ncontent: %s\n\n", b.PageName, b.JournalDay, b.UUID, b.Content)
	}

	// get_block
	if len(result.Blocks) > 0 {
		fmt.Println("=== get_block ===")
		block, err := blockSvc.GetBlock(result.Blocks[0].UUID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(logseq.RenderTree(block, 0))
	}

	// get_page
	fmt.Println("=== get_page: knowledge-note-design ===")
	blocks, exists, err := pageSvc.GetPageBlocks("knowledge-note-design")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exists: %v, blocks: %d\n", exists, len(blocks))
}
