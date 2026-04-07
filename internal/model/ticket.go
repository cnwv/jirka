package model

import "time"

type Ticket struct {
	IssueKey     string
	Summary      string
	StatusName   string
	AssigneeName string
	IsAssigned   bool
	Priority     string
	IssueType    string
	Components   string
	Description  string
	ReporterName string
	Created      time.Time
	Updated      time.Time
}
