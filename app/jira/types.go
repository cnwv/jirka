package jira

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cnwv/jirka/app/model"
)

type SearchResponse struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary        string          `json:"summary"`
	Status         NameField       `json:"status"`
	Assignee       *NameField      `json:"assignee"`
	Priority       *NameField      `json:"priority"`
	Type           NameField       `json:"issuetype"`
	Reporter       *NameField      `json:"reporter"`
	Components     []NameField     `json:"components"`
	Created        string          `json:"created"`
	Updated        string          `json:"updated"`
	RawDescription json.RawMessage `json:"description"`
}

// Description returns the description as a plain string.
// Handles both v2 (string) and v3 (ADF JSON object) formats.
func (f *IssueFields) Description() string {
	if len(f.RawDescription) == 0 {
		return ""
	}
	// Try string first (v2)
	var s string
	if json.Unmarshal(f.RawDescription, &s) == nil {
		return s
	}
	// Try ADF object (v3)
	var doc adfDoc
	if json.Unmarshal(f.RawDescription, &doc) == nil {
		return adfToText(doc.Content)
	}
	return ""
}

// adfDoc is a minimal representation of Atlassian Document Format.
type adfDoc struct {
	Type    string    `json:"type"`
	Content []adfNode `json:"content"`
}

type adfNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text"`
	Content []adfNode `json:"content"`
}

// adfToText extracts plain text from ADF nodes.
func adfToText(nodes []adfNode) string {
	var b strings.Builder
	for i, node := range nodes {
		switch node.Type {
		case "text":
			b.WriteString(node.Text)
		case "hardBreak":
			b.WriteByte('\n')
		default:
			if len(node.Content) > 0 {
				b.WriteString(adfToText(node.Content))
			}
		}
		if node.Type == "paragraph" || node.Type == "heading" || node.Type == "bulletList" ||
			node.Type == "orderedList" || node.Type == "codeBlock" || node.Type == "blockquote" {
			if i < len(nodes)-1 {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

type NameField struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func (n *NameField) GetName() string {
	if n == nil {
		return ""
	}
	if n.DisplayName != "" {
		return n.DisplayName
	}
	return n.Name
}

func ToTickets(issues []Issue) []model.Ticket {
	tickets := make([]model.Ticket, 0, len(issues))

	for _, issue := range issues {
		created, _ := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Created)
		updated, _ := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Updated)

		var components []string
		for _, c := range issue.Fields.Components {
			components = append(components, c.Name)
		}

		t := model.Ticket{
			IssueKey:     issue.Key,
			Summary:      issue.Fields.Summary,
			StatusName:   issue.Fields.Status.Name,
			IsAssigned:   issue.Fields.Assignee != nil,
			AssigneeName: issue.Fields.Assignee.GetName(),
			Priority:     issue.Fields.Priority.GetName(),
			IssueType:    issue.Fields.Type.Name,
			Components:   strings.Join(components, ", "),
			Description:  issue.Fields.Description(),
			ReporterName: issue.Fields.Reporter.GetName(),
			Created:      created,
			Updated:      updated,
		}
		tickets = append(tickets, t)
	}
	return tickets
}
