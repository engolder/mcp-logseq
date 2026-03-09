package logseqsvc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

type PageSvc interface {
	GetPageBlocks(name string) ([]logseq.Block, bool, error)
	CreatePage(name, content string) (string, error)
	// ListPages lists non-journal pages.
	// parent==nil: all pages, parent==&"": root pages only, parent==&"project": direct children of "project".
	ListPages(parent *string, limit, offset int) (*logseq.PageResult, error)
	ListJournalPages(startDate, endDate string, limit, offset int) (*logseq.JournalPageResult, error)
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

func (s *pageSvc) ListPages(parent *string, limit, offset int) (*logseq.PageResult, error) {
	var query string
	switch {
	case parent == nil:
		query = `[:find ?name :where [?p :block/original-name ?name] (not [?p :block/journal? true])]`
	case *parent == "":
		query = `[:find ?name :where [?p :block/original-name ?name] (not [?p :block/namespace _]) (not [?p :block/journal? true])]`
	default:
		query = fmt.Sprintf(
			`[:find ?name :where [?p :block/original-name ?name] [?p :block/namespace ?par] [?par :block/name %q]]`,
			strings.ToLower(*parent),
		)
	}

	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{query})
	if err != nil {
		return nil, err
	}
	var rows [][]string
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row) > 0 {
			names = append(names, row[0])
		}
	}
	sort.Strings(names)

	total := len(names)
	if offset >= total {
		return &logseq.PageResult{Total: total, Pages: []logseq.PageEntry{}}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	names = names[offset:end]

	namespaceSet, err := s.fetchNamespaceSet()
	if err != nil {
		return nil, err
	}

	entries := make([]logseq.PageEntry, len(names))
	for i, name := range names {
		entries[i] = logseq.PageEntry{
			Name:        name,
			HasChildren: namespaceSet[name],
		}
	}
	return &logseq.PageResult{Total: total, Pages: entries}, nil
}

// fetchNamespaceSet returns a set of page original-names that have at least one child page.
func (s *pageSvc) fetchNamespaceSet() (map[string]bool, error) {
	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{
		`[:find ?name :where [_ :block/namespace ?p] [?p :block/original-name ?name]]`,
	})
	if err != nil {
		return nil, err
	}
	var rows [][]string
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(rows))
	for _, row := range rows {
		if len(row) > 0 {
			set[row[0]] = true
		}
	}
	return set, nil
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
	query := "[:find ?name ?jday :where " + strings.Join(conditions, " ") + "]"

	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{query})
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
