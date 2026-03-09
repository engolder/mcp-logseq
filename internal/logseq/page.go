package logseq

type PageEntry struct {
	Name        string `json:"name"`
	HasChildren bool   `json:"has_children"`
}

type PageResult struct {
	Total int
	Pages []PageEntry
}

type JournalPageResult struct {
	Total int
	Pages []string
}
