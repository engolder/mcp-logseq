package logseqsvc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

type PageSvc interface {
	ReadPage(name string) (string, bool, error)
	WritePage(name string, nodes []*logseq.OutlineNode) error
	EditPage(name string, oldContent string, newNodes []*logseq.OutlineNode) error
	DeletePage(name string) error
	SearchPages(query string, limit, offset int) (*logseq.PageResult, error)
	ListJournalPages(startDate, endDate string, limit, offset int) (*logseq.JournalPageResult, error)
}

type pageSvc struct {
	client   *logseq.Client
	blockSvc BlockSvc
}

func NewPageSvc(client *logseq.Client, blockSvc BlockSvc) PageSvc {
	return &pageSvc{client: client, blockSvc: blockSvc}
}

func (s *pageSvc) getPageBlocks(name string) ([]logseq.Block, bool, error) {
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



func (s *pageSvc) ReadPage(name string) (string, bool, error) {
	blocks, exists, err := s.getPageBlocks(name)
	if err != nil {
		return "", false, err
	}
	if !exists {
		return "", false, nil
	}
	var sb strings.Builder
	for _, block := range blocks {
		sb.WriteString(logseq.RenderTree(&block, 0))
	}
	return sb.String(), true, nil
}

func (s *pageSvc) WritePage(name string, nodes []*logseq.OutlineNode) error {
	// Delete existing page if present
	result, err := s.client.DoAPI("logseq.Editor.getPage", []any{name})
	if err != nil {
		return err
	}
	if string(result) != "null" {
		if _, err := s.client.DoAPI("logseq.Editor.deletePage", []any{name}); err != nil {
			return err
		}
	}

	if len(nodes) == 0 {
		_, err := s.client.DoAPI("logseq.Editor.createPage", []any{name, map[string]any{}, map[string]any{"createFirstBlock": false, "redirect": false}})
		return err
	}

	if _, err := s.client.DoAPI("logseq.Editor.createPage", []any{name, map[string]any{}, map[string]any{"createFirstBlock": true, "redirect": false}}); err != nil {
		return err
	}

	raw, err := s.client.DoAPI("logseq.Editor.getPageBlocksTree", []any{name})
	if err != nil {
		return err
	}
	var blocks []logseq.Block
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return err
	}
	if len(blocks) == 0 {
		return fmt.Errorf("page has no blocks after creation")
	}

	firstUUID := blocks[0].UUID
	if err := s.blockSvc.UpdateBlock(firstUUID, nodes[0].Content); err != nil {
		return err
	}
	if len(nodes[0].Children) > 0 {
		if err := s.blockSvc.InsertTree(firstUUID, nodes[0].Children); err != nil {
			return err
		}
	}

	prevUUID := firstUUID
	for _, node := range nodes[1:] {
		uuid, err := s.blockSvc.InsertBlock(prevUUID, node.Content, "after")
		if err != nil {
			return err
		}
		if len(node.Children) > 0 {
			if err := s.blockSvc.InsertTree(uuid, node.Children); err != nil {
				return err
			}
		}
		prevUUID = uuid
	}
	return nil
}

func (s *pageSvc) EditPage(name string, oldContent string, newNodes []*logseq.OutlineNode) error {
	blocks, exists, err := s.getPageBlocks(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("page not found: %s", name)
	}

	var matches []logseq.Block
	var dfs func(bs []logseq.Block)
	dfs = func(bs []logseq.Block) {
		for _, b := range bs {
			if logseq.CleanContent(b.Content) == oldContent {
				matches = append(matches, b)
			}
			dfs(b.Children)
		}
	}
	dfs(blocks)

	if len(matches) == 0 {
		return fmt.Errorf("block not found: %q", oldContent)
	}
	if len(matches) > 1 {
		return fmt.Errorf("ambiguous match: %d blocks match %q", len(matches), oldContent)
	}

	matched := matches[0]

	newRootContent := ""
	if len(newNodes) > 0 {
		newRootContent = newNodes[0].Content
	}
	if err := s.blockSvc.UpdateBlock(matched.UUID, newRootContent); err != nil {
		return err
	}

	for _, child := range matched.Children {
		if _, err := s.client.DoAPI("logseq.Editor.removeBlock", []any{child.UUID}); err != nil {
			return err
		}
	}

	var newChildren []*logseq.OutlineNode
	if len(newNodes) > 0 {
		newChildren = newNodes[0].Children
	}
	if len(newChildren) > 0 {
		if err := s.blockSvc.InsertTree(matched.UUID, newChildren); err != nil {
			return err
		}
	}

	prevUUID := matched.UUID
	for _, node := range newNodes[1:] {
		uuid, err := s.blockSvc.InsertBlock(prevUUID, node.Content, "after")
		if err != nil {
			return err
		}
		if len(node.Children) > 0 {
			if err := s.blockSvc.InsertTree(uuid, node.Children); err != nil {
				return err
			}
		}
		prevUUID = uuid
	}
	return nil
}

func (s *pageSvc) DeletePage(name string) error {
	result, err := s.client.DoAPI("logseq.Editor.getPage", []any{name})
	if err != nil {
		return err
	}
	if string(result) == "null" {
		return fmt.Errorf("page not found: %s", name)
	}
	_, err = s.client.DoAPI("logseq.Editor.deletePage", []any{name})
	return err
}

func (s *pageSvc) SearchPages(query string, limit, offset int) (*logseq.PageResult, error) {
	conditions := []string{
		"[?p :block/original-name ?name]",
		"(not [?p :block/journal? true])",
		// only pages backed by an actual file (matches Logseq's All Pages view)
		"[?p :block/file _]",
	}
	if query != "" {
		conditions = append(conditions, fmt.Sprintf("[(clojure.string/includes? ?name %q)]", query))
	}
	dq := "[:find ?name :where " + strings.Join(conditions, " ") + "]"

	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{dq})
	if err != nil {
		return nil, err
	}
	var rows [][]string
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	// trim + lowercase-based dedup (matches :block/name normalization in Logseq)
	seen := make(map[string]bool)
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row) > 0 {
			name := strings.TrimSpace(row[0])
			if name != "" && !seen[strings.ToLower(name)] {
				seen[strings.ToLower(name)] = true
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)

	total := len(names)
	if offset >= total {
		return &logseq.PageResult{Total: total, Pages: []string{}}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return &logseq.PageResult{Total: total, Pages: names[offset:end]}, nil
}

func (s *pageSvc) ListJournalPages(startDate, endDate string, limit, offset int) (*logseq.JournalPageResult, error) {
	conditions := []string{
		"[?p :block/original-name ?name]",
		"[?p :block/journal? true]",
		"[?p :block/journal-day ?jday]",
	}
	if startDate != "" {
		conditions = append(conditions, fmt.Sprintf("[(>= ?jday %s)]", startDate))
	}
	if endDate != "" {
		conditions = append(conditions, fmt.Sprintf("[(< ?jday %s)]", endDate))
	}
	dq := "[:find ?name ?jday :where " + strings.Join(conditions, " ") + "]"

	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{dq})
	if err != nil {
		return nil, err
	}
	var rows [][]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	type entry struct {
		name string
		day  int
	}
	entries := make([]entry, 0, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		var name string
		var day int
		if err := json.Unmarshal(row[0], &name); err != nil {
			continue
		}
		if err := json.Unmarshal(row[1], &day); err != nil {
			continue
		}
		entries = append(entries, entry{name, day})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].day > entries[j].day // newest first
	})

	total := len(entries)
	if offset >= total {
		return &logseq.JournalPageResult{Total: total, Pages: []string{}}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	names := make([]string, end-offset)
	for i, e := range entries[offset:end] {
		names[i] = e.name
	}
	return &logseq.JournalPageResult{Total: total, Pages: names}, nil
}
