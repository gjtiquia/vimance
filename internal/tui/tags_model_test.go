package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/tui"
)

func TestTagsModelInitialState(t *testing.T) {
	m := tui.NewTagsModel(nil)
	if m.Mode != tui.TagModeInsert {
		t.Errorf("expected TagModeInsert, got %v", m.Mode)
	}
	if len(m.SelectedTags) != 0 {
		t.Errorf("expected no selected tags, got %d", len(m.SelectedTags))
	}
}

func TestTagsModelAddNewTag(t *testing.T) {
	m := tui.NewTagsModel(nil)
	m.SearchInput.Focus()

	m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.SelectedTags) != 1 {
		t.Fatalf("expected 1 selected tag, got %d", len(m.SelectedTags))
	}
	if m.SelectedTags[0].Name != "foo" {
		t.Errorf("expected tag name 'foo', got %q", m.SelectedTags[0].Name)
	}
	if !m.SelectedTags[0].IsNew {
		t.Error("expected IsNew=true for typed tag")
	}
}

func TestTagsModelAddExistingTag(t *testing.T) {
	m := tui.NewTagsModel(nil)
	m.AllTags = []tui.TagItem{{ID: 1, Name: "food"}}
	m.Mode = tui.TagModeNormal

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(m.SelectedTags) != 1 {
		t.Fatalf("expected 1 selected tag, got %d", len(m.SelectedTags))
	}
	if m.SelectedTags[0].IsNew {
		t.Error("expected IsNew=false for existing tag")
	}
	if m.SelectedTags[0].Name != "food" {
		t.Errorf("expected 'food', got %q", m.SelectedTags[0].Name)
	}
}

func TestTagsModelDuplicateIgnored(t *testing.T) {
	m := tui.NewTagsModel(nil)
	m.SearchInput.Focus()

	m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m.SearchInput.Focus()
	m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.SelectedTags) != 1 {
		t.Errorf("expected 1 tag (duplicate ignored), got %d", len(m.SelectedTags))
	}
}

func TestTagsModelRemoveLastTag(t *testing.T) {
	m := tui.NewTagsModel(nil)
	m.SelectedTags = []tui.TagItem{
		{ID: 1, Name: "food"},
		{ID: 2, Name: "drinks"},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.SelectedTags) != 1 {
		t.Fatalf("expected 1 tag after ctrl+z, got %d", len(m.SelectedTags))
	}
	if m.SelectedTags[0].Name != "food" {
		t.Errorf("expected remaining tag 'food', got %q", m.SelectedTags[0].Name)
	}

	m.SelectedTags = nil
	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.SelectedTags) != 0 {
		t.Error("expected 0 tags after ctrl+z on empty")
	}
}

func TestTagsModelModeToggle(t *testing.T) {
	m := tui.NewTagsModel(nil)
	m.Mode = tui.TagModeInsert

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.Mode != tui.TagModeNormal {
		t.Errorf("expected TagModeNormal after esc, got %v", m.Mode)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if m.Mode != tui.TagModeInsert {
		t.Errorf("expected TagModeInsert after a, got %v", m.Mode)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	if m.Mode != tui.TagModeInsert {
		t.Errorf("expected TagModeInsert after i, got %v", m.Mode)
	}
}
