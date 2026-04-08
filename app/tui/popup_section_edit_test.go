package tui

import (
	"testing"

	"github.com/cnwv/jirka/app/config"
)

func TestSectionEditPopup_Build(t *testing.T) {
	sec := config.SectionConfig{Name: "BUGS", JQL: "type = Bug", Color: "38;5;203"}
	p := newSectionEditPopup(0, 0, sec)

	got := p.build()
	if got.Name != "BUGS" {
		t.Errorf("Name = %q, want BUGS", got.Name)
	}
	if got.JQL != "type = Bug" {
		t.Errorf("JQL = %q, want 'type = Bug'", got.JQL)
	}
}

func TestSectionEditPopup_NewSection(t *testing.T) {
	p := newSectionEditPopup(1, -1, config.SectionConfig{})
	if p.sectionIdx != -1 {
		t.Errorf("sectionIdx = %d, want -1", p.sectionIdx)
	}
	if p.panelIdx != 1 {
		t.Errorf("panelIdx = %d, want 1", p.panelIdx)
	}
}

func TestSectionEditPopup_EscCloses(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{Name: "X"})
	action, _ := p.handleKey("esc")
	if action != "close" {
		t.Errorf("action = %q, want close", action)
	}
}

func TestSectionEditPopup_CtrlSSaves(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{Name: "X", JQL: "q"})
	action, sec := p.handleKey("ctrl+s")
	if action != "save" {
		t.Errorf("action = %q, want save", action)
	}
	if sec.Name != "X" {
		t.Errorf("Name = %q, want X", sec.Name)
	}
}

func TestSectionEditPopup_TabCyclesFields(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	if p.focusField != 0 {
		t.Fatalf("initial focus = %d, want 0", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 1 {
		t.Errorf("after tab: focus = %d, want 1", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 2 {
		t.Errorf("after tab: focus = %d, want 2", p.focusField)
	}
	p.handleKey("tab")
	if p.focusField != 0 {
		t.Errorf("after tab: focus = %d, want 0 (wrap)", p.focusField)
	}
}

func TestSectionEditPopup_ShiftTabReverseCycles(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.handleKey("shift+tab")
	if p.focusField != 2 {
		t.Errorf("shift+tab from 0: focus = %d, want 2", p.focusField)
	}
}

func TestSectionEditPopup_TestJQL(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{JQL: "project = X"})
	p.focusField = 1
	action, _ := p.handleKey("enter")
	if action != "test_jql" {
		t.Errorf("action = %q, want test_jql", action)
	}
}

func TestSectionEditPopup_EnterOnNameAdvances(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.focusField = 0
	action, _ := p.handleKey("enter")
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
	if p.focusField != 1 {
		t.Errorf("focusField = %d, want 1", p.focusField)
	}
}

func TestSectionEditPopup_ColorSelector(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.focusField = 2
	initial := p.colorIdx

	p.handleKey("right")
	if p.colorIdx != initial+1 {
		t.Errorf("colorIdx = %d, want %d", p.colorIdx, initial+1)
	}
	p.handleKey("left")
	if p.colorIdx != initial {
		t.Errorf("colorIdx = %d, want %d", p.colorIdx, initial)
	}
}

func TestSectionEditPopup_ColorClampAtBounds(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.focusField = 2
	p.colorIdx = 0
	p.handleKey("left")
	if p.colorIdx != 0 {
		t.Errorf("colorIdx = %d, want 0 (clamped)", p.colorIdx)
	}

	p.colorIdx = len(config.ColorNames) - 1
	p.handleKey("right")
	if p.colorIdx != len(config.ColorNames)-1 {
		t.Errorf("colorIdx = %d, want %d (clamped)", p.colorIdx, len(config.ColorNames)-1)
	}
}

func TestSectionEditPopup_SetTestResult(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.setTestResult(true, "5 tickets")
	if p.testStatus == "" {
		t.Error("testStatus should not be empty after setTestResult")
	}

	p.setTestResult(false, "bad JQL")
	if p.testStatus == "" {
		t.Error("testStatus should not be empty after failure")
	}
}

func TestSectionEditPopup_Paste(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{})
	p.focusField = 1
	p.handlePaste("project = TEST")
	if p.jqlInput.Value != "project = TEST" {
		t.Errorf("jqlInput = %q, want 'project = TEST'", p.jqlInput.Value)
	}
	if p.testStatus != "" {
		t.Error("testStatus should be cleared after paste")
	}
}

func TestSectionEditPopup_CurrentJQL(t *testing.T) {
	p := newSectionEditPopup(0, 0, config.SectionConfig{JQL: "  project = X  "})
	got := p.currentJQL()
	if got != "project = X" {
		t.Errorf("currentJQL = %q, want 'project = X'", got)
	}
}
