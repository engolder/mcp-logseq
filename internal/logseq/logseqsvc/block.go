package logseqsvc

import (
	"encoding/json"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

type BlockSvc interface {
	GetBlock(uuid string) (*logseq.Block, error)
	InsertBlock(targetUUID, content, position string) (string, error)
	UpdateBlock(uuid, content string) error
}

type blockSvc struct {
	client *logseq.Client
}

func NewBlockSvc(client *logseq.Client) BlockSvc {
	return &blockSvc{client: client}
}

func (s *blockSvc) GetBlock(uuid string) (*logseq.Block, error) {
	raw, err := s.client.DoAPI("logseq.Editor.getBlock", []any{uuid, map[string]any{"includeChildren": true}})
	if err != nil {
		return nil, err
	}
	var block logseq.Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (s *blockSvc) InsertBlock(targetUUID, content, position string) (string, error) {
	opts := map[string]any{"before": false, "sibling": true}
	switch position {
	case "before":
		opts["before"] = true
	case "child":
		opts["sibling"] = false
	}
	raw, err := s.client.DoAPI("logseq.Editor.insertBlock", []any{targetUUID, content, opts})
	if err != nil {
		return "", err
	}
	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	return result.UUID, nil
}

func (s *blockSvc) UpdateBlock(uuid, content string) error {
	_, err := s.client.DoAPI("logseq.Editor.updateBlock", []any{uuid, content})
	return err
}
