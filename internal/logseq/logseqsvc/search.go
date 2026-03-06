package logseqsvc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/engolder/mcp-logseq/internal/logseq"
)

type SearchSvc interface {
	ListNamespaces() ([]string, error)
	SearchBlocks(query, namespace, startDate, endDate string, limit, offset int) (*logseq.SearchResult, error)
}

type searchSvc struct {
	client *logseq.Client
}

func NewSearchSvc(client *logseq.Client) SearchSvc {
	return &searchSvc{client: client}
}

func (s *searchSvc) ListNamespaces() ([]string, error) {
	raw, err := s.client.DoAPI("logseq.DB.datascriptQuery", []any{
		"[:find ?name :where [?child :block/namespace ?parent] [?parent :block/original-name ?name]]",
	})
	if err != nil {
		return nil, err
	}
	var nsRows [][]string
	if err := json.Unmarshal(raw, &nsRows); err != nil {
		return nil, err
	}

	raw, err = s.client.DoAPI("logseq.DB.datascriptQuery", []any{
		"[:find ?p :where [?p :block/journal? true]]",
	})
	if err != nil {
		return nil, err
	}
	var journalRows [][]any
	if err := json.Unmarshal(raw, &journalRows); err != nil {
		return nil, err
	}

	var result []string
	if len(journalRows) > 0 {
		result = append(result, "journal")
	}
	for _, row := range nsRows {
		if len(row) > 0 {
			result = append(result, row[0])
		}
	}
	return result, nil
}

func (s *searchSvc) SearchBlocks(query, namespace, startDate, endDate string, limit, offset int) (*logseq.SearchResult, error) {
	conditions := []string{
		"[?b :block/content ?c]",
		"[?b :block/page ?p]",
		"[?p :block/name ?pname]",
	}

	if query != "" {
		conditions = append(conditions, fmt.Sprintf("[(clojure.string/includes? ?c %q)]", query))
	}

	switch {
	case namespace == "journal":
		conditions = append(conditions, "[?p :block/journal? true]")
		if startDate != "" {
			conditions = append(conditions, fmt.Sprintf("[?p :block/journal-day ?jd][(>= ?jd %s)]", startDate))
		}
		if endDate != "" {
			conditions = append(conditions, fmt.Sprintf("[?p :block/journal-day ?jd][(< ?jd %s)]", endDate))
		}
	case namespace != "":
		conditions = append(conditions, fmt.Sprintf("[(clojure.string/starts-with? ?pname %q)]", namespace+"/"))
	default:
		if startDate != "" || endDate != "" {
			conditions = append(conditions, "[?p :block/journal? true]")
			conditions = append(conditions, "[?p :block/journal-day ?jd]")
			if startDate != "" {
				conditions = append(conditions, fmt.Sprintf("[(>= ?jd %s)]", startDate))
			}
			if endDate != "" {
				conditions = append(conditions, fmt.Sprintf("[(< ?jd %s)]", endDate))
			}
		}
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
