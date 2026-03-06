package logseq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

func (c *Client) DoAPI(method string, args []any) (json.RawMessage, error) {
	body, err := json.Marshal(apiRequest{Method: method, Args: args})
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
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}
