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
	pageSvc := logseqsvc.NewPageSvc(client, blockSvc)

	// list_pages (all, first 5)
	fmt.Println("\n=== list_pages (all, first 5) ===")
	pageResult, err := pageSvc.ListPages(5, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total: %d\n", pageResult.Total)
	for _, name := range pageResult.Pages {
		fmt.Println(name)
	}

	// search_blocks
	fmt.Println("\n=== search_blocks: ipc in journal ===")
	result, err := searchSvc.SearchBlocks("ipc", 3, 0)
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

	// read_page
	fmt.Println("=== read_page: knowledge-note-design ===")
	text, exists, err := pageSvc.ReadPage("knowledge-note-design")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exists: %v\n%s", exists, text)
}
