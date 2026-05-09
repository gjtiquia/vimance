package tui_test

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/gjtiquia/vimance/internal/tui"
)

func TestLinksModelInitialState(t *testing.T) {
	m := tui.NewLinksModel(nil)
	if m.Mode != tui.LinkModeInsert {
		t.Errorf("expected LinkModeInsert, got %v", m.Mode)
	}
	if len(m.SelectedParents) != 0 {
		t.Errorf("expected no parents, got %d", len(m.SelectedParents))
	}
}

func TestLinksModelAddParent(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.FilteredCandidates = []tui.LinkedRecord{
		{ID: 1, Date: "2026-05-01", Notes: "test record", AmountCents: 1000},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(m.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(m.SelectedParents))
	}
	if m.SelectedParents[0].ID != 1 {
		t.Errorf("expected parent ID=1, got %d", m.SelectedParents[0].ID)
	}
}

func TestLinksModelDuplicateParentIgnored(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.FilteredCandidates = []tui.LinkedRecord{
		{ID: 1, Date: "2026-05-01", Notes: "test"},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(m.SelectedParents) != 1 {
		t.Errorf("expected 1 parent (duplicate ignored), got %d", len(m.SelectedParents))
	}
}

func TestLinksModelRemoveLastParent(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.SelectedParents = []tui.LinkedRecord{
		{ID: 1, Date: "2026-05-01"},
		{ID: 2, Date: "2026-05-02"},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent after ctrl+z, got %d", len(m.SelectedParents))
	}
	if m.SelectedParents[0].ID != 1 {
		t.Errorf("expected remaining parent ID=1, got %d", m.SelectedParents[0].ID)
	}
}

func TestLinksModelModeToggle(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.Mode = tui.LinkModeInsert

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.Mode != tui.LinkModeNormal {
		t.Errorf("expected LinkModeNormal after esc, got %v", m.Mode)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	if m.Mode != tui.LinkModeInsert {
		t.Errorf("expected LinkModeInsert after i, got %v", m.Mode)
	}
}

func TestLinksModelUpDown(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.FilteredCandidates = []tui.LinkedRecord{
		{ID: 1, Date: "2026-05-01"},
		{ID: 2, Date: "2026-05-02"},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CursorIndex != 1 {
		t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 0 {
		t.Errorf("expected CursorIndex=0, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 0 {
		t.Errorf("expected CursorIndex=0 at top, got %d", m.CursorIndex)
	}
}

func TestLinksModelLoadCandidatesWithoutPrereqs(t *testing.T) {
	m := tui.NewLinksModel(nil)
	m.LoadCandidates(context.Background(), 0)
	view := m.View()
	if strings.Contains(view, "no records found") {
		t.Error("should not say 'no records found' when date/currency not set")
	}
	if !strings.Contains(view, "enter date and currency first") {
		t.Error("should hint to enter date and currency first when prereqs not set")
	}
}

func TestLinksModelLoadCandidatesWithError(t *testing.T) {
	db := setupTestDB(t)
	seedTestDB(t, db)
	svc := service.New(db)
	m := tui.NewLinksModel(svc)
	m.SetDateRange("2026", "05")
	m.SetCurrencyID(1)
	db.Close()

	err := m.LoadCandidates(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error from closed DB")
	}
	if m.LoadErr == "" {
		t.Error("expected LoadErr to be set after error")
	}
	if m.Loaded {
		t.Error("expected Loaded=false after DB error")
	}
	view := m.View()
	if strings.Contains(view, "no records found") {
		t.Error("should not say 'no records found' when error occurred")
	}
	if !strings.Contains(view, "error") {
		t.Error("view should show error message")
	}
}
