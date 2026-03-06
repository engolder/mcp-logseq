package logseq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Client struct {
	apiURL string
	token  string
	http   *http.Client
}

func NewClient() *Client {
	return &Client{
		apiURL: getEnv("LOGSEQ_API_URL", "http://localhost:12315"),
		token:  os.Getenv("LOGSEQ_API_TOKEN"),
		http:   &http.Client{},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type apiRequest struct {
	Method string `json:"method"`
	Args   []any  `json:"args"`
}

type Block struct {
	UUID    string  `json:"uuid"`
	Content string  `json:"content"`
	Level   int     `json:"level"`
	Children []Block `json:"children"`
}

var idPropRe = regexp.MustCompile(`\nid:: [a-f0-9-]+`)

func cleanContent(content string) string {
	return strings.TrimSpace(idPropRe.ReplaceAllString(content, ""))
}

func (c *Client) GetBlock(uuid string) (*Block, error) {
	body, err := json.Marshal(apiRequest{
		Method: "logseq.Editor.getBlock",
		Args:   []any{uuid, map[string]bool{"includeChildren": true}},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.apiURL+"/api", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("logseq api returned %d", resp.StatusCode)
	}

	var block Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, err
	}
	return &block, nil
}

func RenderTree(block *Block, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent) + "- "
	sb.WriteString(prefix + cleanContent(block.Content) + "\n")
	for _, child := range block.Children {
		sb.WriteString(RenderTree(&child, indent+1))
	}
	return sb.String()
}
