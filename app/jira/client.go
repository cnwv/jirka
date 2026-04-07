package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	searchPath     = "/rest/api/2/search"
	maxResults     = 50
	contentTypeJSON = "application/json"
)

var requestedFields = []string{
	"summary", "status", "assignee", "priority", "components",
	"created", "updated", "issuetype", "reporter", "description", "comment",
}

// Client is a Jira REST API client that uses Bearer token authentication.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new Jira client with the given base URL and bearer token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

// CountJQL executes a JQL query with maxResults=0 to validate and get the total count.
func (c *Client) CountJQL(ctx context.Context, jql string) (int, error) {
	body := map[string]any{
		"jql":        jql,
		"fields":     []string{},
		"startAt":    0,
		"maxResults": 0,
	}

	respBody, err := c.doSearch(ctx, body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}
	return result.Total, nil
}

// SearchJQL fetches up to maxResults tickets matching the JQL query.
func (c *Client) SearchJQL(ctx context.Context, jql string) ([]Issue, error) {
	body := map[string]any{
		"jql":        jql,
		"fields":     requestedFields,
		"startAt":    0,
		"maxResults": maxResults,
	}

	respBody, err := c.doSearch(ctx, body)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return searchResp.Issues, nil
}

// doSearch sends a POST request to the Jira search API and returns the response body.
func (c *Client) doSearch(ctx context.Context, body map[string]any) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+searchPath, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
