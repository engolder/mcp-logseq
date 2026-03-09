package logseqsvc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

type SearchSvc interface {
	SearchBlocks(query string, limit, offset int) (*logseq.SearchResult, error)
}

type searchSvc struct {
	client *logseq.Client
}

func NewSearchSvc(client *logseq.Client) SearchSvc {
	return &searchSvc{client: client}
}


func (s *searchSvc) SearchBlocks(query string, limit, offset int) (*logseq.SearchResult, error) {
	conditions := []string{
		"[?b :block/content ?c]",
		"[?b :block/page ?p]",
		"[?p :block/name ?pname]",
	}

	if query != "" {
		conditions = append(conditions, fmt.Sprintf("[(clojure.string/includes? ?c %q)]", query))
	}

	dq := "[:find (pull ?b [:block/uuid :block/content]) (pull ?p [:block/original-name :block/name :block/journal? :block/journal-day]) :where " +
		strings.Join(conditions, " ") + "]"

	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{dq})
	if err != nil {
		return nil, err
	}

	var rows [][]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	type blockData struct {
		UUID    string `json:"uuid"`
		Content string `json:"content"`
	}
	type pageData struct {
		OriginalName string `json:"original-name"`
		Name         string `json:"name"`
		JournalDay   int    `json:"journal-day"`
	}

	total := len(rows)

	if offset >= total {
		return &logseq.SearchResult{Total: total, Blocks: nil}, nil
	}
	rows = rows[offset:]
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}

	blocks := make([]logseq.SearchBlock, 0, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		var bd blockData
		if err := json.Unmarshal(row[0], &bd); err != nil {
			continue
		}
		var pd pageData
		if err := json.Unmarshal(row[1], &pd); err != nil {
			continue
		}

		pageName := pd.OriginalName
		if pageName == "" {
			pageName = pd.Name
		}

		blocks = append(blocks, logseq.SearchBlock{
			UUID:       bd.UUID,
			Content:    bd.Content,
			PageName:   pageName,
			JournalDay: pd.JournalDay,
		})
	}

	return &logseq.SearchResult{Total: total, Blocks: blocks}, nil
}
