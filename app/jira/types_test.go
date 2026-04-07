package jira

import (
	"testing"
)

func TestToTickets(t *testing.T) {
	issues := []Issue{
		{
			Key: "PROJ-1",
			Fields: IssueFields{
				Summary:  "First issue",
				Status:   NameField{Name: "Open"},
				Assignee: &NameField{DisplayName: "Alice"},
				Priority: &NameField{Name: "High"},
				Type:     NameField{Name: "Bug"},
				Reporter: &NameField{DisplayName: "Bob"},
				Components: []NameField{
					{Name: "Backend"},
					{Name: "API"},
				},
				Description: "Some description",
				Created:     "2026-03-15T10:30:00.000+0000",
				Updated:     "2026-03-16T14:00:00.000+0000",
			},
		},
		{
			Key: "PROJ-2",
			Fields: IssueFields{
				Summary: "Unassigned issue",
				Status:  NameField{Name: "Closed"},
				Type:    NameField{Name: "Task"},
			},
		},
	}

	tickets := ToTickets(issues)

	if len(tickets) != 2 {
		t.Fatalf("expected 2 tickets, got %d", len(tickets))
	}

	// First ticket
	tk := tickets[0]
	if tk.IssueKey != "PROJ-1" {
		t.Errorf("key = %q, want PROJ-1", tk.IssueKey)
	}
	if tk.AssigneeName != "Alice" {
		t.Errorf("assignee = %q, want Alice", tk.AssigneeName)
	}
	if !tk.IsAssigned {
		t.Error("expected IsAssigned=true")
	}
	if tk.Components != "Backend, API" {
		t.Errorf("components = %q, want 'Backend, API'", tk.Components)
	}
	if tk.Created.Year() != 2026 {
		t.Errorf("created year = %d, want 2026", tk.Created.Year())
	}

	// Second ticket — unassigned
	tk2 := tickets[1]
	if tk2.IsAssigned {
		t.Error("expected IsAssigned=false for nil assignee")
	}
	if tk2.AssigneeName != "" {
		t.Errorf("assignee = %q, want empty", tk2.AssigneeName)
	}
}

func TestNameField_GetName(t *testing.T) {
	tests := []struct {
		name  string
		field *NameField
		want  string
	}{
		{"nil", nil, ""},
		{"display name", &NameField{Name: "n", DisplayName: "d"}, "d"},
		{"name only", &NameField{Name: "n"}, "n"},
		{"empty", &NameField{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.field.GetName()
			if got != tt.want {
				t.Errorf("GetName() = %q, want %q", got, tt.want)
			}
		})
	}
}
