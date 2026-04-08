package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	searchPathV2    = "/rest/api/2/search"
	searchPathV3    = "/rest/api/3/search/jql"
	maxResults      = 50
	contentTypeJSON = "application/json"
)

var requestedFields = []string{
	"summary", "status", "assignee", "priority", "components",
	"created", "updated", "issuetype", "reporter", "description", "comment",
}

// Client is a Jira REST API client that supports Bearer and Basic authentication.
type Client struct {
	BaseURL    string
	Token      string
	Email      string
	AuthType   string // "bearer" or "basic"
	HTTPClient *http.Client
}

// NewClient creates a new Jira client with Bearer token authentication (Server/Data Center).
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		AuthType:   "bearer",
		HTTPClient: &http.Client{},
	}
}

// NewClientBasic creates a new Jira client with Basic authentication (Cloud).
func NewClientBasic(baseURL, email, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		Email:      email,
		AuthType:   "basic",
		HTTPClient: &http.Client{},
	}
}

// CountJQL executes a JQL query to validate and get the total count.
// For Cloud (v3 API), returns -1 on success since the total count is unavailable.
func (c *Client) CountJQL(ctx context.Context, jql string) (int, error) {
	if c.isCloud() {
		// v3 API requires maxResults >= 1 and doesn't return total
		body := map[string]any{
			"jql":        jql,
			"fields":     []string{"summary"},
			"maxResults": 1,
		}
		_, err := c.doSearch(ctx, body)
		if err != nil {
			return 0, err
		}
		return -1, nil
	}

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

// isCloud returns true if the client uses Cloud (basic auth / v3 API).
func (c *Client) isCloud() bool {
	return c.AuthType == "basic"
}

// SearchJQL fetches up to maxResults tickets matching the JQL query.
func (c *Client) SearchJQL(ctx context.Context, jql string) ([]Issue, error) {
	body := map[string]any{
		"jql":        jql,
		"fields":     requestedFields,
		"maxResults": maxResults,
	}
	if !c.isCloud() {
		body["startAt"] = 0
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

	searchURL := c.BaseURL + searchPathV2
	if c.isCloud() {
		searchURL = c.BaseURL + searchPathV3
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeJSON)
	switch c.AuthType {
	case "basic":
		creds := base64.StdEncoding.EncodeToString([]byte(c.Email + ":" + c.Token))
		req.Header.Set("Authorization", "Basic "+creds)
	default:
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

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
