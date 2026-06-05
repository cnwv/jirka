package tui

import (
	"testing"

	"github.com/cnwv/jirka/app/config"
)

func TestPanelEditPopup_BuildNoSections(t *testing.T) {
	pc := config.PanelConfig{Title: "Todo", JQL: "project = X", Color: "38;5;75"}
	p := newPanelEditPopup(0, pc)

	got := p.build()
	if got.Title != "Todo" {
		t.Errorf("Title = %q, want Todo", got.Title)
	}
	if got.JQL != "project = X" {
		t.Errorf("JQL = %q, want 'project = X'", got.JQL)
	}
	if len(got.Sections) != 0 {
		t.Errorf("expected no sections, got %d", len(got.Sections))
	}
}

func TestPanelEditPopup_PreservesSections(t *testing.T) {
	pc := config.PanelConfig{
		Title: "Mixed",
		Color: "38;5;75",
		Sections: []config.SectionConfig{
			{Name: "BUGS", JQL: "type = Bug", Color: "38;5;203"},
		},
	}
	p := newPanelEditPopup(0, pc)

	if !p.hasSections() {
		t.Fatal("expected hasSections=true")
	}
	got := p.build()
	if len(got.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got.Sections))
	}
	if got.Sections[0].Name != "BUGS" {
		t.Errorf("section name = %q, want BUGS", got.Sections[0].Name)
	}
	// JQL must be empty when sections exist
	if got.JQL != "" {
		t.Errorf("JQL must be empty when sections exist, got %q", got.JQL)
	}
}

func TestPanelEditPopup_EscCloses(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{Title: "X"})
	action, _ := p.handleKey("esc")
	if action != actionClose {
		t.Errorf("action = %q, want close", action)
	}
}

func TestPanelEditPopup_CtrlSSaves(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{Title: "X", JQL: "q"})
	action, pc := p.handleKey("ctrl+s")
	if action != actionSave {
		t.Errorf("action = %q, want save", action)
	}
	if pc.Title != "X" {
		t.Errorf("Title = %q, want X", pc.Title)
	}
}

func TestPanelEditPopup_TabCyclesFields(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{})
	if p.focusField != 0 {
		t.Fatalf("initial focus = %d, want 0", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 1 {
		t.Errorf("after 1 tab: focus = %d, want 1", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 2 {
		t.Errorf("after 2 tabs: focus = %d, want 2", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 0 {
		t.Errorf("after 3 tabs: focus = %d, want 0 (wrap)", p.focusField)
	}
}

func TestPanelEditPopup_TestJQL(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{JQL: "project = X"})
	p.focusField = 1
	action, _ := p.handleKey("enter")
	if action != actionTestJQL {
		t.Errorf("action = %q, want test_jql", action)
	}
}

func TestPanelEditPopup_DeleteSection(t *testing.T) {
	pc := config.PanelConfig{
		Title: "P",
		Sections: []config.SectionConfig{
			{Name: "A"},
			{Name: "B"},
		},
	}
	p := newPanelEditPopup(0, pc)
	p.focusField = 2
	p.sectionCursor = 0

	p.handleKey("d")

	if len(p.sections) != 1 {
		t.Fatalf("expected 1 section after delete, got %d", len(p.sections))
	}
	if p.sections[0].Name != "B" {
		t.Errorf("remaining section = %q, want B", p.sections[0].Name)
	}
}

func TestPanelEditPopup_AddSectionOpensPopup(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{})
	p.focusField = 2
	action, _ := p.handleKey("a")
	if action != actionAddSection {
		t.Errorf("action = %q, want add_section", action)
	}
}

func TestPanelEditPopup_EditSectionOpensPopup(t *testing.T) {
	pc := config.PanelConfig{
		Sections: []config.SectionConfig{{Name: "X"}},
	}
	p := newPanelEditPopup(0, pc)
	p.focusField = 2
	p.sectionCursor = 0

	action, _ := p.handleKey("enter")
	if action != actionEditSection {
		t.Errorf("action = %q, want edit_section", action)
	}
}

func TestPanelEditPopup_AddSection(t *testing.T) {
	p := newPanelEditPopup(0, config.PanelConfig{})
	sec := config.SectionConfig{Name: "BUGS", JQL: "type=Bug", Color: "38;5;203"}
	p.addSection(sec)

	if len(p.sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(p.sections))
	}
	if p.sections[0].Name != "BUGS" {
		t.Errorf("section name = %q, want BUGS", p.sections[0].Name)
	}
	if p.focusField != 2 {
		t.Errorf("focusField = %d, want 2 (sections)", p.focusField)
	}
}

func TestPanelEditPopup_SectionNavigate(t *testing.T) {
	pc := config.PanelConfig{
		Sections: []config.SectionConfig{{Name: "A"}, {Name: "B"}, {Name: "C"}},
	}
	p := newPanelEditPopup(0, pc)
	p.focusField = 2

	p.handleKey("down")
	if p.sectionCursor != 1 {
		t.Errorf("cursor = %d, want 1", p.sectionCursor)
	}
	p.handleKey("down")
	p.handleKey("down") // should not go past last
	if p.sectionCursor != 2 {
		t.Errorf("cursor = %d, want 2 (clamped)", p.sectionCursor)
	}
	p.handleKey("up")
	if p.sectionCursor != 1 {
		t.Errorf("cursor = %d, want 1", p.sectionCursor)
	}
}

func TestPanelEditPopup_UpdateSection(t *testing.T) {
	pc := config.PanelConfig{
		Sections: []config.SectionConfig{{Name: "Old", JQL: "old"}},
	}
	p := newPanelEditPopup(0, pc)
	p.updateSection(0, config.SectionConfig{Name: "New", JQL: "new"})

	if p.sections[0].Name != "New" {
		t.Errorf("name = %q, want New", p.sections[0].Name)
	}
}
