package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/tui"
)

func setupRecordModel(t *testing.T) tui.RecordModel {
	t.Helper()
	return tui.NewRecordModel(newTestService(t))
}

func TestRecordModelInitialState(t *testing.T) {
	m := setupRecordModel(t)
	if m.State != tui.RecordStateEditing {
		t.Errorf("expected RecordStateEditing, got %v", m.State)
	}
	if m.ActiveField != tui.FieldDateYear {
		t.Errorf("expected FieldDateYear, got %v", m.ActiveField)
	}
}

func TestRecordModelFieldNavigation(t *testing.T) {
	m := setupRecordModel(t)
	fields := []tui.ActiveField{
		tui.FieldDateYear, tui.FieldDateMonth, tui.FieldDateDay,
		tui.FieldCurrency, tui.FieldTags, tui.FieldAmount,
		tui.FieldLinks, tui.FieldNotes,
	}

	for i, expected := range fields {
		if m.ActiveField != expected {
			t.Errorf("step %d: expected %v, got %v", i, expected, m.ActiveField)
		}
		if i < len(fields)-1 {
			m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		}
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldNotes {
		t.Errorf("tab at end should stay at FieldNotes, got %v", m.ActiveField)
	}

	for i := len(fields) - 2; i >= 0; i-- {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		if m.ActiveField != fields[i] {
			t.Errorf("shift+tab step %d: expected %v, got %v", i, fields[i], m.ActiveField)
		}
	}
}

func TestRecordModelEnterOnDateFillsDefault(t *testing.T) {
	m := setupRecordModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.ActiveField != tui.FieldDateMonth {
		t.Errorf("expected FieldDateMonth after enter on year, got %v", m.ActiveField)
	}
	if m.DateYearInput.Value() == "" {
		t.Error("expected year to be filled with placeholder default")
	}
}

func TestRecordModelEnterOnNotesGoesToConfirm(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateConfirm {
		t.Errorf("expected RecordStateConfirm, got %v", m.State)
	}
}

func TestRecordModelConfirmNumberKeys(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	tests := []struct {
		key      string
		expected tui.ActiveField
	}{
		{"1", tui.FieldDateYear},
		{"2", tui.FieldCurrency},
		{"3", tui.FieldTags},
		{"4", tui.FieldAmount},
		{"5", tui.FieldLinks},
		{"6", tui.FieldNotes},
	}
	for _, tc := range tests {
		copyM := m
		copyM, _ = copyM.Update(tea.KeyPressMsg{Code: rune(tc.key[0]), Text: tc.key})
		if copyM.State != tui.RecordStateEditing {
			t.Errorf("key %q: expected RecordStateEditing, got %v", tc.key, copyM.State)
		}
		if copyM.ActiveField != tc.expected {
			t.Errorf("key %q: expected %v, got %v", tc.key, tc.expected, copyM.ActiveField)
		}
	}
}

func TestRecordModelConfirmEscReturnsToEditing(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if m.State != tui.RecordStateEditing {
		t.Errorf("expected RecordStateEditing after esc, got %v", m.State)
	}
	if m.ActiveField != tui.FieldNotes {
		t.Errorf("expected FieldNotes after esc from confirm, got %v", m.ActiveField)
	}
}

func TestRecordModelConfirmEnterWithErrors(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.ConfirmModel.Errors.HasErrors() {
		t.Error("expected validation errors (no currency, no amount)")
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateConfirm {
		t.Errorf("should stay in confirm when errors exist, got %v", m.State)
	}
}

func TestRecordModelCurrencyAdvanceResetsOnFieldChange(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.ActiveField)
	}

	// Type "USD" and enter to select
	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.CurrencyInput.Selected == nil || m.CurrencyInput.Selected.Code != "USD" {
		t.Fatal("expected USD selected")
	}

	// Enter auto-advances to FieldTags and clears ShouldAdvance
	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected FieldTags after enter, got %v", m.ActiveField)
	}
	if m.CurrencyInput.ShouldAdvance {
		t.Error("ShouldAdvance should be false after auto-advance")
	}
}

func TestRecordModelConfirmEnterNoErrors(t *testing.T) {
	m := setupRecordModel(t)

	// Tab through date fields (fills defaults): year → month → day → currency
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.ActiveField)
	}

	// Type "USD" and enter to select
	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.CurrencyInput.Selected == nil || m.CurrencyInput.Selected.Code != "USD" {
		t.Fatal("expected USD selected")
	}

	// Auto-advances to Tags on enter
	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected auto-advance to FieldTags, got %v", m.ActiveField)
	}

	// Tab to Amount
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}

	// Type "10.00"
	m, _ = m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '.', Text: "."})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})

	// Tab to Links
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldLinks {
		t.Fatalf("expected FieldLinks, got %v", m.ActiveField)
	}

	// Tab to Notes
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	// Enter on Notes → Confirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateConfirm {
		t.Fatalf("expected RecordStateConfirm, got %v", m.State)
	}
	if m.ConfirmModel.Errors.HasErrors() {
		t.Fatalf("expected no validation errors, got %v", m.ConfirmModel.Errors)
	}

	// Enter to save → Success
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateSuccess {
		t.Errorf("expected RecordStateSuccess, got %v", m.State)
	}
}

func TestRecordModelInactiveSubModelsIgnoreKeys(t *testing.T) {
	m := setupRecordModel(t)

	m.TagsInput.SelectedTags = []tui.TagItem{{ID: 1, Name: "food"}}
	m.CurrencyInput.AllCurrencies = []tui.CurrencyItem{
		{ID: 1, Code: "USD"}, {ID: 2, Code: "EUR"},
	}
	m.CurrencyInput.CursorIndex = 0
	m.CurrencyInput.Mode = tui.CurrencyModeNormal

	// Tab from FieldDateYear to FieldAmount
	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}

	// ctrl+z on Amount should NOT remove tag from TagsInput
	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.TagsInput.SelectedTags) != 1 {
		t.Errorf("ctrl+z on Amount should not remove tags: got %d", len(m.TagsInput.SelectedTags))
	}

	// down arrow on Amount should NOT move currency cursor
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CurrencyInput.CursorIndex != 0 {
		t.Errorf("down arrow on Amount should not move currency cursor: got %d", m.CurrencyInput.CursorIndex)
	}

	// Navigate to Tags and verify ctrl+z DOES work on the active field
	// Tab from Amount → Links → Notes, then shift+tab back to Tags
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // → FieldLinks
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // → FieldNotes
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}) // → FieldLinks
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}) // → FieldAmount
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}) // → FieldTags
	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected FieldTags, got %v", m.ActiveField)
	}

	// ctrl+z on Tags SHOULD remove tag
	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.TagsInput.SelectedTags) != 0 {
		t.Errorf("ctrl+z on Tags should remove tag: got %d", len(m.TagsInput.SelectedTags))
	}
}
