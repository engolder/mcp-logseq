package logseq

type SearchBlock struct {
	UUID       string
	Content    string
	PageName   string
	JournalDay int // 0 if not a journal page
}

type SearchResult struct {
	Total  int
	Blocks []SearchBlock
}
