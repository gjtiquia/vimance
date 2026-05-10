package tui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/gjtiquia/vimance/internal/tui"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

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

func TestQueryModelResultsEscRouting(t *testing.T) {
	tests := []struct {
		name     string
		origin   tui.QueryState
		expected tui.QueryState
	}{
		{"filter form origin", tui.QueryStateFilterForm, tui.QueryStateFilterForm},
		{"saved list origin", tui.QueryStateSavedList, tui.QueryStateSavedList},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := setupQueryModel(t)
			m.State = tui.QueryStateResults
			m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
			m.ResultsOrigin = tc.origin

			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
			if m.State != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, m.State)
			}
		})
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
	tests := []struct {
		name     string
		origin   tui.QueryState
		expected tui.QueryState
	}{
		{"filter form origin", tui.QueryStateFilterForm, tui.QueryStateFilterForm},
		{"saved list origin", tui.QueryStateSavedList, tui.QueryStateSavedList},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := setupQueryModel(t)
			m.State = tui.QueryStateResults
			m.Results = []service.QueryResult{}
			m.ResultsOrigin = tc.origin

			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
			if m.State != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, m.State)
			}
		})
	}
}

func TestQueryModelResultsErrorDismissal(t *testing.T) {
	tests := []struct {
		name         string
		origin       tui.QueryState
		expected     tui.QueryState
	}{
		{"saved list origin", tui.QueryStateSavedList, tui.QueryStateSavedList},
		{"filter form origin", tui.QueryStateFilterForm, tui.QueryStateFilterForm},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := setupQueryModel(t)
			m.State = tui.QueryStateResults
			m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
			m.ResultsOrigin = tc.origin
			m.ErrorMsg = "some error"

			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if m.State != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, m.State)
			}
			if m.ErrorMsg != "" {
				t.Error("expected ErrorMsg cleared after dismissal")
			}
		})
	}
}

func TestQueryModelInactiveSubModelsIgnoreKeys(t *testing.T) {
	m := setupQueryModel(t)
	m.State = tui.QueryStateFilterForm
	m.ActiveField = tui.FilterDateTo
	m.FocusActiveField()

	m.Currency.AllCurrencies = []tui.CurrencyItem{
		{ID: 1, Code: "USD"},
		{ID: 2, Code: "EUR"},
	}
	m.Currency.CursorIndex = 0
	m.Currency.Mode = tui.CurrencyModeNormal

	// Part 1: inactive field — down on DateTo should NOT move currency cursor
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.Currency.CursorIndex != 0 {
		t.Errorf("down on DateTo should not move currency cursor: got %d", m.Currency.CursorIndex)
	}

	// Part 2: active field — set to Currency and verify down DOES move cursor
	m.ActiveField = tui.FilterCurrency
	m.Currency.Mode = tui.CurrencyModeNormal
	m.Currency.CursorIndex = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.Currency.CursorIndex != 1 {
		t.Errorf("down on Currency should move cursor: expected 1, got %d", m.Currency.CursorIndex)
	}
}

// --- Picker mode tests ---

func setupPickerModel(t *testing.T) tui.QueryModel {
	t.Helper()
	m := setupQueryModel(t)
	m.PickerMode = true
	m.State = tui.QueryStatePickerMenu
	return m
}

func TestPickerMenuInitialState(t *testing.T) {
	m := setupPickerModel(t)
	if m.State != tui.QueryStatePickerMenu {
		t.Errorf("expected QueryStatePickerMenu, got %v", m.State)
	}
	if item, ok := m.PickerMenu.SelectedItem(); !ok || item.Title != "new query" {
		t.Errorf("expected first item 'new query', got %q", item.Title)
	}
}

func TestPickerMenuSelectNew(t *testing.T) {
	m := setupPickerModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.QueryStateFilterForm {
		t.Errorf("expected QueryStateFilterForm, got %v", m.State)
	}
}

func TestPickerMenuSelectSaved(t *testing.T) {
	m := setupPickerModel(t)
	// navigate to "saved queries"
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.QueryStateSavedList {
		t.Errorf("expected QueryStateSavedList, got %v", m.State)
	}
}

func TestPickerMenuNavigation(t *testing.T) {
	m := setupPickerModel(t)

	// down → saved queries
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if item, ok := m.PickerMenu.SelectedItem(); !ok || item.Title != "saved queries" {
		t.Errorf("expected 'saved queries', got %q", item.Title)
	}

	// j → saved queries
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if item, ok := m.PickerMenu.SelectedItem(); !ok || item.Title != "saved queries" {
		t.Errorf("expected 'saved queries' after j, got %q", item.Title)
	}

	// up → back to new query
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if item, ok := m.PickerMenu.SelectedItem(); !ok || item.Title != "new query" {
		t.Errorf("expected 'new query' after up, got %q", item.Title)
	}

	// k → new query (already at top)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if item, ok := m.PickerMenu.SelectedItem(); !ok || item.Title != "new query" {
		t.Errorf("expected 'new query' after k, got %q", item.Title)
	}
}

func TestPickerMenuEscCancels(t *testing.T) {
	m := setupPickerModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if !m.PickerDone {
		t.Error("expected PickerDone set after esc")
	}
	if !m.PickerCancelled {
		t.Error("expected PickerCancelled set after esc")
	}
}

func TestPickerMenuNumberKeys(t *testing.T) {
	m := setupPickerModel(t)

	// "2" selects saved queries
	m, _ = m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if m.State != tui.QueryStateSavedList {
		t.Errorf("expected QueryStateSavedList after 2, got %v", m.State)
	}
}

func TestPickerFilterFormEscToMenu(t *testing.T) {
	m := setupPickerModel(t)
	// enter → filter form
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	// esc from filter form → back to picker menu
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStatePickerMenu {
		t.Errorf("expected QueryStatePickerMenu after esc from filter, got %v", m.State)
	}
}

func TestPickerSavedListEscToMenu(t *testing.T) {
	m := setupPickerModel(t)
	// navigate to saved, enter
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	// esc from saved list → back to picker menu
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.State != tui.QueryStatePickerMenu {
		t.Errorf("expected QueryStatePickerMenu after esc from saved list, got %v", m.State)
	}
}

func TestPickerResultsMultiSelect(t *testing.T) {
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{
		{ID: 1, Date: "2026-05-01", Notes: "a"},
		{ID: 2, Date: "2026-05-02", Notes: "b"},
		{ID: 3, Date: "2026-05-03", Notes: "c"},
	}
	m.CursorIndex = 0

	// space toggles selection on cursor
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if len(m.PickerSelected) != 1 {
		t.Fatalf("expected 1 selected, got %d", len(m.PickerSelected))
	}
	if m.PickerSelected[0].ID != 1 {
		t.Errorf("expected selected ID 1, got %d", m.PickerSelected[0].ID)
	}

	// move down and select another
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if len(m.PickerSelected) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(m.PickerSelected))
	}

	// toggle off first one
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if len(m.PickerSelected) != 1 {
		t.Fatalf("expected 1 selected after toggle, got %d", len(m.PickerSelected))
	}
	if m.PickerSelected[0].ID != 2 {
		t.Errorf("expected remaining selected ID 2, got %d", m.PickerSelected[0].ID)
	}
}

func TestPickerResultsEnterConfirms(t *testing.T) {
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{
		{ID: 1, Date: "2026-05-01", Notes: "a"},
	}
	m.CursorIndex = 0
	m.PickerSelected = []service.QueryResult{{ID: 1, Date: "2026-05-01", Notes: "a"}}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.PickerDone {
		t.Error("expected PickerDone after enter in picker results")
	}
	if m.PickerCancelled {
		t.Error("expected PickerCancelled=false after enter in picker results")
	}
	if len(m.PickerSelected) != 1 {
		t.Errorf("expected 1 selected record, got %d", len(m.PickerSelected))
	}
}

func TestPickerResultsEnterEmptyConfirmsEmpty(t *testing.T) {
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{
		{ID: 1, Date: "2026-05-01", Notes: "a"},
	}
	m.CursorIndex = 0
	// no selections

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.PickerDone {
		t.Error("expected PickerDone after enter with empty selection")
	}
	if m.PickerCancelled {
		t.Error("expected PickerCancelled=false")
	}
	if len(m.PickerSelected) != 0 {
		t.Errorf("expected 0 selected, got %d", len(m.PickerSelected))
	}
}

func TestPickerResultsEnterDoesNotSetSelectedID(t *testing.T) {
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{{ID: 42, Date: "2026-05-01"}}
	m.CursorIndex = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.SelectedID != 0 {
		t.Errorf("expected SelectedID=0 in picker mode, got %d", m.SelectedID)
	}
}

func TestPickerResultsEscRouting(t *testing.T) {
	tests := []struct {
		name     string
		origin   tui.QueryState
		expected tui.QueryState
	}{
		{"filter form origin", tui.QueryStateFilterForm, tui.QueryStateFilterForm},
		{"saved list origin", tui.QueryStateSavedList, tui.QueryStateSavedList},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := setupPickerModel(t)
			m.State = tui.QueryStateResults
			m.PickerMode = true
			m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
			m.ResultsOrigin = tc.origin

			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
			if m.State != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, m.State)
			}
		})
	}
}

func TestPickerViewShowsSelectionIndicators(t *testing.T) {
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{
		{ID: 1, Date: "2026-05-01", Notes: "a"},
		{ID: 2, Date: "2026-05-02", Notes: "b"},
	}
	m.CursorIndex = 0
	m.PickerSelected = []service.QueryResult{{ID: 1, Date: "2026-05-01", Notes: "a"}}

	view := m.View()
	if !contains(view, "[x]") {
		t.Errorf("expected [x] for selected record in picker view, got:\n%s", view)
	}
	if !contains(view, "[ ]") {
		t.Errorf("expected [ ] for unselected record in picker view, got:\n%s", view)
	}
}

func TestPickerModeSavedQuerySavesWorks(t *testing.T) {
	// pressing s in picker results should still open save name
	m := setupPickerModel(t)
	m.State = tui.QueryStateResults
	m.PickerMode = true
	m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}

	m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if m.State != tui.QueryStateSaveName {
		t.Errorf("expected QueryStateSaveName after s in picker mode, got %v", m.State)
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
