package tui_test

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/tui"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func cleanView(v string) string {
	return ansiRe.ReplaceAllString(v, "")
}

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
		tui.FieldNotes,
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
	for i := 0; i < 6; i++ {
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
	for i := 0; i < 6; i++ {
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
		{"5", tui.FieldNotes},
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
	for i := 0; i < 6; i++ {
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
	for i := 0; i < 6; i++ {
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

	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.CurrencyInput.Selected == nil || m.CurrencyInput.Selected.Code != "USD" {
		t.Fatal("expected USD selected")
	}

	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected FieldTags after enter, got %v", m.ActiveField)
	}
	if m.CurrencyInput.ShouldAdvance {
		t.Error("ShouldAdvance should be false after auto-advance")
	}
}

func TestRecordModelConfirmEnterNoErrors(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.CurrencyInput.Selected == nil || m.CurrencyInput.Selected.Code != "USD" {
		t.Fatal("expected USD selected")
	}

	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected auto-advance to FieldTags, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '.', Text: "."})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateConfirm {
		t.Fatalf("expected RecordStateConfirm, got %v", m.State)
	}
	if m.ConfirmModel.Errors.HasErrors() {
		t.Fatalf("expected no validation errors, got %v", m.ConfirmModel.Errors)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateSuccess {
		t.Errorf("expected RecordStateSuccess, got %v", m.State)
	}
}

func TestRecordViewInitialState_DateExpandedOthersCollapsed(t *testing.T) {
	m := setupRecordModel(t)
	cv := cleanView(m.View())

	if !strings.Contains(cv, "> Year:") {
		t.Error("active Year should have > caret")
	}
	if !strings.Contains(cv, "  Month:") {
		t.Error("unfocused Month should show without caret")
	}
	if !strings.Contains(cv, "  Day:") {
		t.Error("unfocused Day should show without caret")
	}
	if strings.Contains(cv, "  Date:") {
		t.Error("collapsed date line should NOT appear when date is focused")
	}

	if !strings.Contains(cv, "  Currency:") {
		t.Error("collapsed Currency should appear")
	}
	if !strings.Contains(cv, "  Tags:") {
		t.Error("collapsed Tags should appear")
	}
	if !strings.Contains(cv, "  Amount:") {
		t.Error("collapsed Amount should appear")
	}
	if !strings.Contains(cv, "  Notes:") {
		t.Error("collapsed Notes should appear")
	}

	if strings.Contains(cv, "←") {
		t.Error("no inline errors should appear at initial state")
	}
	if strings.Contains(cv, "⚠") {
		t.Error("no inline warnings should appear at initial state")
	}
}

func TestRecordView_DateCaretMovesWithinGroup(t *testing.T) {
	m := setupRecordModel(t)

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	cv := cleanView(m.View())

	if !strings.Contains(cv, "  Year:") {
		t.Error("Year should lose caret after tab")
	}
	if !strings.Contains(cv, "> Month:") {
		t.Error("Month should gain caret after tab")
	}
	if !strings.Contains(cv, "  Day:") {
		t.Error("Day should still show without caret")
	}
}

func TestRecordView_DateCollapsesAfterLeaving(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())

	if !strings.Contains(cv, "  Date:") {
		t.Error("collapsed date line should appear after leaving date fields")
	}
	if strings.Contains(cv, "> Year:") || strings.Contains(cv, "> Month:") || strings.Contains(cv, "> Day:") {
		t.Error("no date fields should have caret after leaving date group")
	}

	if !strings.Contains(cv, "> Currency:") {
		t.Error("active Currency should have > caret")
	}
}

func TestRecordView_TagsExpandedWhenActive(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected FieldTags, got %v", m.ActiveField)
	}

	view := m.View()
	if !strings.Contains(view, "> Tags:") {
		t.Error("active Tags should have > caret")
	}
	if !strings.Contains(view, "Type:") {
		t.Error("expanded Tags should show search input prompt 'Type:'")
	}
}

func TestRecordView_AmountShowsCaretWhenActive(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Amount:") {
		t.Error("active Amount should have > caret")
	}
}

func TestRecordView_CollapsedCurrencyShowsInlineError(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "←") {
		t.Error("expected inline error for currency left empty")
	}
}

func TestRecordView_CollapsedTagsShowsInlineWarning(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "⚠") {
		t.Error("expected inline warning for empty tags")
	}
}

func TestRecordView_ActiveFieldChangesCaret(t *testing.T) {
	m := setupRecordModel(t)

	for i, expected := range []string{"> Month:", "> Day:", "> Currency:", "> Tags:", "> Amount:", "> Notes:"} {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		cv := cleanView(m.View())
		if !strings.Contains(cv, expected) {
			t.Errorf("step %d: expected %q in view", i, expected)
		}
	}
}

func TestRecordView_CollapsedAmountShowsEmpty(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "  Amount: (empty)") {
		t.Error("collapsed empty Amount should show '(empty)', got view:\n" + cv)
	}
}

func TestRecordView_CollapsedAmountShowsValue(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: '5', Text: "5"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "  Amount: 50") {
		t.Error("collapsed Amount should show value '50', got view:\n" + cv)
	}
	if strings.Contains(cv, "0.00") {
		t.Error("collapsed Amount should not show placeholder '0.00', got view:\n" + cv)
	}
}

func TestRecordView_AmountInlineErrorOnlyWhenCollapsed(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "← amount is required") {
		t.Error("collapsed Amount should show inline error when empty, got:\n" + cv)
	}
	if !strings.Contains(cv, "  Amount: (empty)") {
		t.Error("collapsed Amount should show (empty) with error, got:\n" + cv)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount back, got %v", m.ActiveField)
	}

	cv = cleanView(m.View())
	if strings.Contains(cv, "← amount is required") {
		t.Error("active Amount should NOT show inline error when focused, got:\n" + cv)
	}
	if !strings.Contains(cv, "> Amount:") {
		t.Error("active Amount should show > caret, got:\n" + cv)
	}
}

func TestRecordView_CollapsedNotesShowsEmpty(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Notes:") {
		t.Error("active Notes should show > caret, got:\n" + cv)
	}

	if strings.Contains(cv, "  Notes:") {
		t.Error("active Notes should NOT show collapsed format, got:\n" + cv)
	}
}

func TestRecordView_CollapsedNotesSummary(t *testing.T) {
	m := setupRecordModel(t)

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Notes: test") {
		t.Error("active Notes should show typed value with caret, got:\n" + cv)
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

	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.TagsInput.SelectedTags) != 1 {
		t.Errorf("ctrl+z on Amount should not remove tags: got %d", len(m.TagsInput.SelectedTags))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CurrencyInput.CursorIndex != 0 {
		t.Errorf("down arrow on Amount should not move currency cursor: got %d", m.CurrencyInput.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.ActiveField != tui.FieldTags {
		t.Fatalf("expected FieldTags, got %v", m.ActiveField)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.TagsInput.SelectedTags) != 0 {
		t.Errorf("ctrl+z on Tags should remove tag: got %d", len(m.TagsInput.SelectedTags))
	}
}