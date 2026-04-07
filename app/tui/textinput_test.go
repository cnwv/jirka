package tui

import (
	"testing"
)

func TestTextInput_Insert(t *testing.T) {
	var ti textInput
	ti.handleKey("h")
	ti.handleKey("i")
	if ti.Value != "hi" {
		t.Errorf("Value = %q, want %q", ti.Value, "hi")
	}
	if ti.cursor != 2 {
		t.Errorf("cursor = %d, want 2", ti.cursor)
	}
}

func TestTextInput_Backspace(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.handleKey("backspace")
	if ti.Value != "ab" {
		t.Errorf("Value = %q, want %q", ti.Value, "ab")
	}
}

func TestTextInput_CursorMovement(t *testing.T) {
	var ti textInput
	ti.set("hello")

	ti.handleKey("left")
	if ti.cursor != 4 {
		t.Errorf("after left: cursor = %d, want 4", ti.cursor)
	}

	ti.handleKey("home")
	if ti.cursor != 0 {
		t.Errorf("after home: cursor = %d, want 0", ti.cursor)
	}

	ti.handleKey("end")
	if ti.cursor != 5 {
		t.Errorf("after end: cursor = %d, want 5", ti.cursor)
	}
}

func TestTextInput_MaxLen(t *testing.T) {
	ti := textInput{maxLen: 3}
	ti.handleKey("a")
	ti.handleKey("b")
	ti.handleKey("c")
	ti.handleKey("d") // should be ignored

	if ti.Value != "abc" {
		t.Errorf("Value = %q, want %q", ti.Value, "abc")
	}
}

func TestTextInput_DeleteForward(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.home()
	ti.handleKey("delete")
	if ti.Value != "bc" {
		t.Errorf("Value = %q, want %q", ti.Value, "bc")
	}
}

func TestTextInput_CtrlK(t *testing.T) {
	var ti textInput
	ti.set("hello world")
	ti.cursor = 5
	ti.handleKey("ctrl+k")
	if ti.Value != "hello" {
		t.Errorf("Value = %q, want %q", ti.Value, "hello")
	}
}

func TestTextInput_CtrlU(t *testing.T) {
	var ti textInput
	ti.set("hello world")
	ti.cursor = 6
	ti.handleKey("ctrl+u")
	if ti.Value != "world" {
		t.Errorf("Value = %q, want %q", ti.Value, "world")
	}
	if ti.cursor != 0 {
		t.Errorf("cursor = %d, want 0", ti.cursor)
	}
}

func TestTextInput_InsertString(t *testing.T) {
	ti := textInput{maxLen: 5}
	ti.insertString("abcdefgh")
	if ti.Value != "abcde" {
		t.Errorf("Value = %q, want %q (capped at maxLen)", ti.Value, "abcde")
	}
}

func TestTextInput_InsertStringSkipsControl(t *testing.T) {
	var ti textInput
	ti.insertString("a\x00b\x1fc")
	if ti.Value != "abc" {
		t.Errorf("Value = %q, want %q", ti.Value, "abc")
	}
}

func TestTextInput_View(t *testing.T) {
	var ti textInput
	ti.set("test")
	v := ti.view(0)
	if v == "" {
		t.Error("view should not be empty")
	}
	// Should contain the reverse video escape for cursor
	plain := stripAnsiStr(v)
	if plain != "test " { // space for cursor at end
		t.Errorf("plain view = %q, want %q", plain, "test ")
	}
}

func TestTextInput_ViewReadonly(t *testing.T) {
	var ti textInput
	ti.set("long text here")
	v := ti.viewReadonly(4)
	if v != "long" {
		t.Errorf("viewReadonly(4) = %q, want %q", v, "long")
	}
}

func TestTextInput_Set(t *testing.T) {
	ti := textInput{maxLen: 3}
	ti.set("toolong")
	if ti.Value != "too" {
		t.Errorf("Value = %q, want %q", ti.Value, "too")
	}
	if ti.cursor != 3 {
		t.Errorf("cursor = %d, want 3", ti.cursor)
	}
}
