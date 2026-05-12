package tui_test

import (
	"strings"
	"testing"

	"github.com/gjtiquia/vimance/internal/tui"
)

func TestLinksModelInitialState(t *testing.T) {
	m := tui.NewLinksModel()
	if len(m.SelectedParents) != 0 {
		t.Errorf("expected no parents, got %d", len(m.SelectedParents))
	}
}

func TestLinksModelAddParent(t *testing.T) {
	m := tui.NewLinksModel()
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "test record", AmountCents: 1000})
	if len(m.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(m.SelectedParents))
	}
	if m.SelectedParents[0].ID != 1 {
		t.Errorf("expected parent ID=1, got %d", m.SelectedParents[0].ID)
	}
}

func TestLinksModelDuplicateParentIgnored(t *testing.T) {
	m := tui.NewLinksModel()
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "test"})
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "test"})
	if len(m.SelectedParents) != 1 {
		t.Errorf("expected 1 parent (duplicate ignored), got %d", len(m.SelectedParents))
	}
}

func TestLinksModelRemoveLastParent(t *testing.T) {
	m := tui.NewLinksModel()
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01"})
	m.AddParent(tui.LinkedRecord{ID: 2, Date: "2026-05-02"})
	m.RemoveLastParent()
	if len(m.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent after remove, got %d", len(m.SelectedParents))
	}
	if m.SelectedParents[0].ID != 1 {
		t.Errorf("expected remaining parent ID=1, got %d", m.SelectedParents[0].ID)
	}
}

func TestLinksModelRemoveFromEmpty(t *testing.T) {
	m := tui.NewLinksModel()
	m.RemoveLastParent()
	if len(m.SelectedParents) != 0 {
		t.Errorf("expected 0 parents after remove from empty, got %d", len(m.SelectedParents))
	}
}

func TestLinksModelViewNoParents(t *testing.T) {
	m := tui.NewLinksModel()
	view := m.View()
	if !strings.Contains(view, "(none)") {
		t.Errorf("expected '(none)' in view, got: %s", view)
	}
}

func TestLinksModelViewWithParents(t *testing.T) {
	m := tui.NewLinksModel()
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "coffee beans"})
	m.AddParent(tui.LinkedRecord{ID: 2, Date: "2026-05-02", Notes: "lunch out"})
	view := m.View()
	if !strings.Contains(view, "[coffee beans]") {
		t.Errorf("expected '[coffee beans]' in view, got: %s", view)
	}
	if !strings.Contains(view, "[lunch out]") {
		t.Errorf("expected '[lunch out]' in view, got: %s", view)
	}
}

func TestLinksModelViewTruncatesLongNotes(t *testing.T) {
	longNotes := "this is a very long note that should be truncated to twenty characters"
	m := tui.NewLinksModel()
	m.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: longNotes})
	view := m.View()
	if !strings.Contains(view, "[this is a very long ...]") {
		t.Errorf("expected truncated note in view, got: %s", view)
	}
}
