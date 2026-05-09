package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/gjtiquia/vimance/internal/tui"
)

func setupQueryModel(t *testing.T) tui.QueryModel {
	t.Helper()
	return tui.NewQueryModel(newTestService(t))
}

func setupQueryModelInFilterForm(t *testing.T) tui.QueryModel {
	t.Helper()
	m := setupQueryModel(t)
	m.State = tui.QueryStateFilterForm
	m.ActiveField = tui.FilterDateFrom
	m.FocusActiveField()
	return m
}

func TestQueryModelInitialState(t *testing.T) {
	m := setupQueryModel(t)
	if m.State != tui.QueryStateMenu {
		t.Errorf("expected QueryStateMenu, got %v", m.State)
	}
}

func TestQueryModelFilterFieldNavigation(t *testing.T) {
	m := setupQueryModelInFilterForm(t)
	fields := []tui.FilterField{
		tui.FilterDateFrom, tui.FilterDateTo,
		tui.FilterCurrency, tui.FilterTags,
		tui.FilterFuzzy,
	}

	for i, expected := range fields {
		if m.ActiveField != expected {
			t.Errorf("step %d: expected %v, got %v", i, expected, m.ActiveField)
		}
		if i < len(fields)-1 {
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		}
	}

	// Tab on Fuzzy → QueryStateConfirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.State != tui.QueryStateConfirm {
		t.Errorf("expected QueryStateConfirm after tab on Fuzzy, got %v", m.State)
	}
}

func TestQueryModelEscFromFilterForm(t *testing.T) {
	m := setupQueryModelInFilterForm(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateMenu {
		t.Errorf("expected QueryStateMenu after esc, got %v", m.State)
	}
}

func TestQueryModelConfirmEsc(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateConfirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm after esc, got %v", m.State)
	}
}

func TestQueryModelConfirmNumberKeys(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateConfirm

	keyFieldMap := map[string]tui.FilterField{
		"1": tui.FilterDateFrom,
		"2": tui.FilterCurrency,
		"3": tui.FilterTags,
		"4": tui.FilterFuzzy,
	}
	for key, expected := range keyFieldMap {
		copyM := m
		copyM, _ = copyM.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
		if copyM.State != tui.QueryStateFilterForm {
			t.Errorf("key %q: expected QueryStateFilterForm, got %v", key, copyM.State)
		}
		if copyM.ActiveField != expected {
			t.Errorf("key %q: expected %v, got %v", key, expected, copyM.ActiveField)
		}
	}
}

func TestQueryModelConfirmEnterWithValidDates(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateConfirm

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.QueryStateResults {
		t.Errorf("expected QueryStateResults, got %v", m.State)
	}
}

func TestQueryModelConfirmEnterInvalidDateFrom(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateConfirm
	m.DateFrom.SetValue("not-a-date")

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm on invalid date, got %v", m.State)
	}
	if m.ErrorMsg == "" {
		t.Error("expected error message for invalid date")
	}
}

func TestQueryModelConfirmEnterInvalidDateTo(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateConfirm
	m.DateTo.SetValue("bad")

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm, got %v", m.State)
	}
	if m.ErrorMsg == "" {
		t.Error("expected error message for invalid date")
	}
}

func TestQueryModelResultsCursorNavigation(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateResults
	m.Results = []service.QueryResult{
		{ID: 1, Date: "2026-05-01"},
		{ID: 2, Date: "2026-05-02"},
		{ID: 3, Date: "2026-05-03"},
		{ID: 4, Date: "2026-05-04"},
		{ID: 5, Date: "2026-05-05"},
	}
	m.CursorIndex = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.CursorIndex != 1 {
		t.Errorf("expected 1 after j, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CursorIndex != 2 {
		t.Errorf("expected 2 after down, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if m.CursorIndex != 1 {
		t.Errorf("expected 1 after k, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 0 {
		t.Errorf("expected 0 after up, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 0 {
		t.Errorf("expected 0 at top, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	if m.CursorIndex != 4 {
		t.Errorf("expected 4 after G, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	if m.CursorIndex != 0 {
		t.Errorf("expected 0 after g, got %d", m.CursorIndex)
	}
}

func TestQueryModelResultsEnterSelectsRecord(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateResults
	m.Results = []service.QueryResult{
		{ID: 42, Date: "2026-05-01"},
	}
	m.CursorIndex = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.SelectedID != 42 {
		t.Errorf("expected SelectedID=42, got %d", m.SelectedID)
	}
}

func TestQueryModelResultsEscToFilterForm(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateResults
	m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
	m.ResultsOrigin = tui.QueryStateFilterForm

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm, got %v", m.State)
	}
}

func TestQueryModelResultsSaveName(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateResults
	m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}

	m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if m.State != tui.QueryStateSaveName {
		t.Errorf("expected QueryStateSaveName after s, got %v", m.State)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateResults {
		t.Errorf("expected QueryStateResults after esc, got %v", m.State)
	}
}

func TestQueryModelResultsEmpty(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateResults
	m.Results = []service.QueryResult{}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm, got %v", m.State)
	}
}

func TestQueryModelDeleteConfirm(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateDeleteConfirm
	m.DeleteTarget = tui.SavedQueryItem{ID: 1, Name: "test query"}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if m.State != tui.QueryStateSavedList {
		t.Errorf("expected QueryStateSavedList after n, got %v", m.State)
	}

	m.State = tui.QueryStateDeleteConfirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStateSavedList {
		t.Errorf("expected QueryStateSavedList after esc, got %v", m.State)
	}
}
