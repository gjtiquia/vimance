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
	if !strings.Contains(cv, "  Links:") {
		t.Error("collapsed Links should appear")
	}
	if !strings.Contains(cv, "  Amount:") {
		t.Error("collapsed Amount should appear")
	}
	if !strings.Contains(cv, "  Notes:") {
		t.Error("collapsed Notes should appear")
	}

	// No inline errors at start (no field has been left yet)
	if strings.Contains(cv, "←") {
		t.Error("no inline errors should appear at initial state")
	}
	if strings.Contains(cv, "⚠") {
		t.Error("no inline warnings should appear at initial state")
	}
}

func TestRecordView_DateCaretMovesWithinGroup(t *testing.T) {
	m := setupRecordModel(t)

	// Tab from Year to Month
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

	// Tab through all date fields to Currency
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())

	// Date should be collapsed now
	if !strings.Contains(cv, "  Date:") {
		t.Error("collapsed date line should appear after leaving date fields")
	}
	if strings.Contains(cv, "> Year:") || strings.Contains(cv, "> Month:") || strings.Contains(cv, "> Day:") {
		t.Error("no date fields should have caret after leaving date group")
	}

	// Currency should be expanded
	if !strings.Contains(cv, "> Currency:") {
		t.Error("active Currency should have > caret")
	}
}

func TestRecordView_TagsExpandedWhenActive(t *testing.T) {
	m := setupRecordModel(t)

	// Tab through Year, Month, Day to Currency
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	// Type USD and Enter to auto-advance to Tags
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

	// Tab through Year, Month, Day, Currency, Tags to Amount
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

	// Tab through all date fields to Currency
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	// Tab through Currency to Tags (no selection = leaves currency empty)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	// Currency should show inline error since we left it without selecting
	if !strings.Contains(cv, "←") {
		t.Error("expected inline error for currency left empty")
	}
}

func TestRecordView_CollapsedTagsShowsInlineWarning(t *testing.T) {
	m := setupRecordModel(t)

	// Tab all the way to Notes
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	// Tags should show inline warning since no tags selected
	if !strings.Contains(cv, "⚠") {
		t.Error("expected inline warning for empty tags")
	}
}

func TestRecordView_LinksCollapsedWhenNotActive(t *testing.T) {
	m := setupRecordModel(t)

	// Tab all the way to Notes
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "  Links:") {
		t.Error("collapsed Links should appear with label")
	}
}

func TestRecordView_ActiveFieldChangesCaret(t *testing.T) {
	m := setupRecordModel(t)

	// Tab forward checking caret moves
	for i, expected := range []string{"> Month:", "> Day:", "> Currency:", "> Tags:", "> Amount:", "> Links:", "> Notes:"} {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		cv := cleanView(m.View())
		if !strings.Contains(cv, expected) {
			t.Errorf("step %d: expected %q in view", i, expected)
		}
	}
}

func TestRecordView_CollapsedAmountShowsEmpty(t *testing.T) {
	m := setupRecordModel(t)

	// Tab through Year, Month, Day, Currency, Tags to Amount
	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	// Tab to Links (Amount stays empty)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "  Amount: (empty)") {
		t.Error("collapsed empty Amount should show '(empty)', got view:\n" + cv)
	}
}

func TestRecordView_CollapsedAmountShowsValue(t *testing.T) {
	m := setupRecordModel(t)

	// Tab through Year, Month, Day, Currency, Tags to Amount
	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	// Type a value
	m, _ = m.Update(tea.KeyPressMsg{Code: '5', Text: "5"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	// Tab to Links (Amount has value now)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cv := cleanView(m.View())
	if !strings.Contains(cv, "  Amount: 50") {
		t.Error("collapsed Amount should show value '50', got view:\n" + cv)
	}
	// Should NOT contain placeholder
	if strings.Contains(cv, "0.00") {
		t.Error("collapsed Amount should not show placeholder '0.00', got view:\n" + cv)
	}
}

func TestRecordView_AmountInlineErrorOnlyWhenCollapsed(t *testing.T) {
	m := setupRecordModel(t)

	// Tab to Amount
	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.ActiveField)
	}
	// Tab to Links leaving Amount empty → inline error stored
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveField != tui.FieldLinks {
		t.Fatalf("expected FieldLinks, got %v", m.ActiveField)
	}

	// Collapsed Amount should show inline error
	cv := cleanView(m.View())
	if !strings.Contains(cv, "← amount is required") {
		t.Error("collapsed Amount should show inline error when empty, got:\n" + cv)
	}
	if !strings.Contains(cv, "  Amount: (empty)") {
		t.Error("collapsed Amount should show (empty) with error, got:\n" + cv)
	}

	// Tab back to Amount
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount back, got %v", m.ActiveField)
	}

	// Active Amount should NOT show inline error (just caret + textinput)
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

	// Tab all the way to Notes
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	// Notes is active → should show with caret
	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Notes:") {
		t.Error("active Notes should show > caret, got:\n" + cv)
	}

	// Can't tab past Notes (it's last), so check its active view instead
	// Active view should show textinput placeholder but not "(empty)"
	if strings.Contains(cv, "  Notes:") {
		t.Error("active Notes should NOT show collapsed format, got:\n" + cv)
	}
}

func TestRecordView_CollapsedNotesSummary(t *testing.T) {
	m := setupRecordModel(t)

	// Tab to Notes
	for i := 0; i < 7; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
	}

	// Type a note
	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})

	// Enter to confirm → Notes collapses in confirm state, but for edit view:
	// Can't tab past Notes to collapse it. But enter goes to confirm.
	// Instead, test that active Notes shows typed value
	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Notes: test") {
		t.Error("active Notes should show typed value with caret, got:\n" + cv)
	}
}

func TestRecordView_CollapsedLinksSummary(t *testing.T) {
	m := setupRecordModel(t)

	// Tab to Links
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldLinks {
		t.Fatalf("expected FieldLinks, got %v", m.ActiveField)
	}

	cv := cleanView(m.View())
	if !strings.Contains(cv, "> Links:") {
		t.Error("active Links should have > caret")
	}

	// Tab away to Notes
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	cv = cleanView(m.View())
	if !strings.Contains(cv, "  Links:") {
		t.Error("collapsed Links should appear with label")
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

// --- Link picker integration tests ---

func TestRecordModelLinkPickerOpensOnEnter(t *testing.T) {
	m := setupRecordModel(t)
	// tab to Links field
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // month
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // day
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // currency
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // tags
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // amount
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // links

	if m.ActiveField != tui.FieldLinks {
		t.Fatalf("expected FieldLinks, got %v", m.ActiveField)
	}

	// enter opens picker
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.State != tui.RecordStateLinkPicker {
		t.Fatalf("expected RecordStateLinkPicker, got %v", m.State)
	}
	if !m.LinkPickerQuery.PickerMode {
		t.Error("expected LinkPickerQuery.PickerMode true")
	}
	if m.LinkPickerQuery.State != tui.QueryStatePickerMenu {
		t.Errorf("expected QueryStatePickerMenu, got %v", m.LinkPickerQuery.State)
	}
}

func TestRecordModelLinkPickerCancelReturnsToLinks(t *testing.T) {
	m := setupRecordModel(t)
	// tab to Links, enter to open picker
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // open picker

	// esc from picker menu → cancels
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if m.State != tui.RecordStateEditing {
		t.Fatalf("expected RecordStateEditing after cancel, got %v", m.State)
	}
	if m.ActiveField != tui.FieldLinks {
		t.Errorf("expected FieldLinks after cancel, got %v", m.ActiveField)
	}
	if len(m.LinksInput.SelectedParents) != 0 {
		t.Errorf("expected no parents after cancel, got %d", len(m.LinksInput.SelectedParents))
	}
}

func TestRecordModelLinkPickerSelectAndConfirm(t *testing.T) {
	svc := newTestService(t)
	// seed data: user=1, currency=1 (USD), tag=1
	userID := int64(1)
	usdID := int64(1)
	m := tui.NewRecordModel(svc)

	// seed some records to pick from
	r1, _ := svc.CreateRecord(t.Context(), "2026-05-01", 1000, usdID, "first record", userID)
	r2, _ := svc.CreateRecord(t.Context(), "2026-05-02", 2000, usdID, "second record", userID)

	// pre-fill date so picker pre-fills date range
	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("05")

	// tab to Links, enter to open picker
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // open picker → PickerMenu

	// select "new query" → FilterForm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.LinkPickerQuery.State != tui.QueryStateFilterForm {
		t.Fatalf("expected FilterForm, got %v", m.LinkPickerQuery.State)
	}

	// tab through filter fields to reach confirm
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	// now on Fuzzy field, tab → Confirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.LinkPickerQuery.State != tui.QueryStateConfirm {
		t.Fatalf("expected Confirm, got %v", m.LinkPickerQuery.State)
	}

	// enter → execute query → Results
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.LinkPickerQuery.State != tui.QueryStateResults {
		t.Fatalf("expected Results, got %v", m.LinkPickerQuery.State)
	}
	if len(m.LinkPickerQuery.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(m.LinkPickerQuery.Results))
	}

	// select both records with space
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}) // select first
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})          // move down
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}) // select second

	if len(m.LinkPickerQuery.PickerSelected) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(m.LinkPickerQuery.PickerSelected))
	}

	// enter to confirm picker
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.State != tui.RecordStateEditing {
		t.Fatalf("expected RecordStateEditing after picker confirm, got %v", m.State)
	}
	if m.ActiveField != tui.FieldLinks {
		t.Errorf("expected FieldLinks after picker confirm, got %v", m.ActiveField)
	}
	if len(m.LinksInput.SelectedParents) != 2 {
		t.Fatalf("expected 2 parents after confirm, got %d", len(m.LinksInput.SelectedParents))
	}

	// verify the correct records were added
	ids := map[int64]bool{}
	for _, p := range m.LinksInput.SelectedParents {
		ids[p.ID] = true
	}
	if !ids[r1.ID] {
		t.Errorf("missing parent r1 (ID=%d)", r1.ID)
	}
	if !ids[r2.ID] {
		t.Errorf("missing parent r2 (ID=%d)", r2.ID)
	}
}

func TestRecordModelLinkPickerDoesNotDuplicateParents(t *testing.T) {
	svc := newTestService(t)
	userID := int64(1)
	usdID := int64(1)
	m := tui.NewRecordModel(svc)

	svc.CreateRecord(t.Context(), "2026-05-01", 1000, usdID, "record one", userID)

	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("05")

	// add one parent manually
	m.LinksInput.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "record one"})

	// tab to Links, enter to open picker
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// new query → filter → confirm → results
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // confirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // execute → results

	// select the record (same ID as existing parent)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // confirm

	// should still only have 1 parent (duplicate ignored by AddParent)
	if len(m.LinksInput.SelectedParents) != 1 {
		t.Errorf("expected 1 parent (duplicate ignored), got %d", len(m.LinksInput.SelectedParents))
	}
}

func TestRecordModelCtrlZRemovesLastParentOnLinks(t *testing.T) {
	m := setupRecordModel(t)
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if m.ActiveField != tui.FieldLinks {
		t.Fatalf("expected FieldLinks, got %v", m.ActiveField)
	}

	m.LinksInput.AddParent(tui.LinkedRecord{ID: 1, Date: "2026-05-01", Notes: "first"})
	m.LinksInput.AddParent(tui.LinkedRecord{ID: 2, Date: "2026-05-02", Notes: "second"})
	if len(m.LinksInput.SelectedParents) != 2 {
		t.Fatalf("expected 2 parents, got %d", len(m.LinksInput.SelectedParents))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.LinksInput.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent after ctrl+z, got %d", len(m.LinksInput.SelectedParents))
	}
	if m.LinksInput.SelectedParents[0].ID != 1 {
		t.Errorf("expected remaining parent ID=1, got %d", m.LinksInput.SelectedParents[0].ID)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.LinksInput.SelectedParents) != 0 {
		t.Errorf("expected 0 parents after second ctrl+z, got %d", len(m.LinksInput.SelectedParents))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if len(m.LinksInput.SelectedParents) != 0 {
		t.Errorf("expected 0 parents after ctrl+z on empty, got %d", len(m.LinksInput.SelectedParents))
	}
	if m.ActiveField != tui.FieldLinks {
		t.Errorf("should remain on FieldLinks after ctrl+z, got %v", m.ActiveField)
	}
}

func TestRecordModelLinkPickerConfirmEmptySelection(t *testing.T) {
	svc := newTestService(t)
	userID := int64(1)
	usdID := int64(1)
	m := tui.NewRecordModel(svc)

	svc.CreateRecord(t.Context(), "2026-05-01", 1000, usdID, "record one", userID)
	svc.CreateRecord(t.Context(), "2026-05-02", 2000, usdID, "record two", userID)

	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("05")

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.LinkPickerQuery.State != tui.QueryStateResults {
		t.Fatalf("expected Results, got %v", m.LinkPickerQuery.State)
	}
	if len(m.LinkPickerQuery.Results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(m.LinkPickerQuery.Results))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.State != tui.RecordStateEditing {
		t.Fatalf("expected RecordStateEditing, got %v", m.State)
	}
	if m.ActiveField != tui.FieldLinks {
		t.Errorf("expected FieldLinks, got %v", m.ActiveField)
	}
	if len(m.LinksInput.SelectedParents) != 0 {
		t.Errorf("expected 0 parents from empty selection, got %d", len(m.LinksInput.SelectedParents))
	}
}

func TestRecordModelLinkPickerPreFillsDates(t *testing.T) {
	svc := newTestService(t)
	m := tui.NewRecordModel(svc)

	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("07")

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.LinkPickerQuery.DateFrom.Value() != "2026-07-01" {
		t.Errorf("expected DateFrom 2026-07-01, got %q", m.LinkPickerQuery.DateFrom.Value())
	}
	if m.LinkPickerQuery.DateTo.Value() != "2026-07-31" {
		t.Errorf("expected DateTo 2026-07-31, got %q", m.LinkPickerQuery.DateTo.Value())
	}
}

func TestRecordModelLinkPickerPreservesExistingParents(t *testing.T) {
	svc := newTestService(t)
	userID := int64(1)
	usdID := int64(1)
	m := tui.NewRecordModel(svc)

	r, _ := svc.CreateRecord(t.Context(), "2026-05-02", 2000, usdID, "second record", userID)

	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("05")

	m.LinksInput.AddParent(tui.LinkedRecord{ID: 99, Date: "2026-04-01", Notes: "manual parent"})

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.LinkPickerQuery.Results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(m.LinkPickerQuery.Results))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.LinksInput.SelectedParents) != 2 {
		t.Fatalf("expected 2 parents (manual + picked), got %d", len(m.LinksInput.SelectedParents))
	}

	foundManual := false
	foundSeeded := false
	for _, p := range m.LinksInput.SelectedParents {
		if p.ID == 99 {
			foundManual = true
		}
		if p.ID == r.ID {
			foundSeeded = true
		}
	}
	if !foundManual {
		t.Error("manual parent (ID=99) missing after picker confirm")
	}
	if !foundSeeded {
		t.Error("seeded record missing after picker confirm")
	}
}

func TestRecordModelLinkPickerFieldMapping(t *testing.T) {
	svc := newTestService(t)
	userID := int64(1)
	usdID := int64(1)
	tagFoodID := int64(1)
	m := tui.NewRecordModel(svc)

	r, _ := svc.CreateRecordWithTagsAndLinks(t.Context(), "2026-05-15", 1500, usdID, "field mapping test", userID, []int64{tagFoodID}, nil)

	m.DateYearInput.SetValue("2026")
	m.DateMonthInput.SetValue("05")

	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.LinkPickerQuery.Results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(m.LinkPickerQuery.Results))
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(m.LinksInput.SelectedParents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(m.LinksInput.SelectedParents))
	}

	p := m.LinksInput.SelectedParents[0]
	if p.ID != r.ID {
		t.Errorf("expected ID %d, got %d", r.ID, p.ID)
	}
	if p.Date != "2026-05-15" {
		t.Errorf("expected Date 2026-05-15, got %q", p.Date)
	}
	if p.AmountCents != 1500 {
		t.Errorf("expected AmountCents 1500, got %d", p.AmountCents)
	}
	if p.CurrencyID != usdID {
		t.Errorf("expected CurrencyID %d, got %d", usdID, p.CurrencyID)
	}
	if p.Notes != "field mapping test" {
		t.Errorf("expected Notes 'field mapping test', got %q", p.Notes)
	}
	if len(p.TagNames) == 0 || p.TagNames[0] != "food" {
		t.Errorf("expected TagNames to include 'food', got %v", p.TagNames)
	}
}
