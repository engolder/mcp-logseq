package logseqsvc

import (
	"encoding/json"
	"github.com/engolder/mcp-logseq/internal/logseq"
)

type PageSvc interface {
	GetPageBlocks(name string) ([]logseq.Block, bool, error)
	CreatePage(name, content string) (string, error)
}

type pageSvc struct {
	client *logseq.Client
}

func NewPageSvc(client *logseq.Client) PageSvc {
	return &pageSvc{client: client}
}

func (s *pageSvc) GetPageBlocks(name string) ([]logseq.Block, bool, error) {
	result, err := s.client.DoAPI("logseq.Editor.getPage", []any{name})
	if err != nil {
		return nil, false, err
	}
	if string(result) == "null" {
		return nil, false, nil
	}

	result, err = s.client.DoAPI("logseq.Editor.getPageBlocksTree", []any{name})
	if err != nil {
		return nil, false, err
	}
	var blocks []logseq.Block
	if err := json.Unmarshal(result, &blocks); err != nil {
		return nil, false, err
	}
	return blocks, true, nil
}

func (s *pageSvc) CreatePage(name, content string) (string, error) {
	result, err := s.client.DoAPI("logseq.Editor.createPage", []any{name, map[string]any{}, map[string]any{"createFirstBlock": true, "redirect": false}})
	if err != nil {
		return "", err
	}
	var page struct {
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(result, &page); err != nil {
		return "", err
	}

	result, err = s.client.DoAPI("logseq.Editor.getPageBlocksTree", []any{name})
	if err != nil {
		return "", err
	}
	var blocks []logseq.Block
	if err := json.Unmarshal(result, &blocks); err != nil {
		return "", err
	}

	if len(blocks) == 0 {
		return page.UUID, nil
	}

	if _, err := s.client.DoAPI("logseq.Editor.updateBlock", []any{blocks[0].UUID, content}); err != nil {
		return "", err
	}
	return page.UUID, nil
}
