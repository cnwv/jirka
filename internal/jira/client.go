package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CountJQL executes a JQL query with maxResults=0 to validate and get the total count.
func (c *Client) CountJQL(ctx context.Context, jql string) (int, error) {
	body := map[string]any{
		"jql":        jql,
		"fields":     []string{},
		"startAt":    0,
		"maxResults": 0,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/rest/api/2/search", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("jira request: %w", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("jira returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}
	return result.Total, nil
}

var requestedFields = []string{
	"summary", "status", "assignee", "priority", "components",
	"created", "updated", "issuetype", "reporter", "description", "comment",
}

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SearchJQL(ctx context.Context, jql string) ([]Issue, error) {
	body := map[string]any{
		"jql":        jql,
		"fields":     requestedFields,
		"startAt":    0,
		"maxResults": 50,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/rest/api/2/search", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira returned %d: %s", resp.StatusCode, string(respBody))
	}

	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return searchResp.Issues, nil
}
