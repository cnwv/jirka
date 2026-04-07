package jira

import (
	"buble_jira/internal/model"
	"strings"
	"time"
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
	Summary     string      `json:"summary"`
	Status      NameField   `json:"status"`
	Assignee    *NameField  `json:"assignee"`
	Priority    *NameField  `json:"priority"`
	Type        NameField   `json:"issuetype"`
	Reporter    *NameField  `json:"reporter"`
	Components  []NameField `json:"components"`
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	Description string      `json:"description"`
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
			Description:  issue.Fields.Description,
			ReporterName: issue.Fields.Reporter.GetName(),
			Created:      created,
			Updated:      updated,
		}
		tickets = append(tickets, t)
	}
	return tickets
}
