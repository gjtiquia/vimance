package tui_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/gjtiquia/vimance/internal/tui"
)

func TestModelInitialState(t *testing.T) {
	m := newTestModel(t)
	if m.InputType != tui.InputTypeList {
		t.Errorf("expected InputTypeList, got %v", m.InputType)
	}
}

func TestModelCtrlCQuits(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModelListToCreate(t *testing.T) {
	m := newTestModel(t)
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)

	if m.InputType != tui.InputTypeRecord {
		t.Errorf("expected InputTypeRecord, got %v", m.InputType)
	}
	if m.RecordInput.State != tui.RecordStateEditing {
		t.Errorf("expected RecordStateEditing, got %v", m.RecordInput.State)
	}
}

func TestModelListToQueryAndBack(t *testing.T) {
	m := newTestModel(t)

	r, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = r.(tui.Model)
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)

	if m.InputType != tui.InputTypeQuery {
		t.Errorf("expected InputTypeQuery, got %v", m.InputType)
	}
	if m.QueryInput.State != tui.QueryStateMenu {
		t.Errorf("expected QueryStateMenu, got %v", m.QueryInput.State)
	}

	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = r.(tui.Model)
	if m.InputType != tui.InputTypeList {
		t.Errorf("expected InputTypeList after esc, got %v", m.InputType)
	}
}

func TestModelWindowSize(t *testing.T) {
	m := newTestModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = result.(tui.Model)
	if m.Width != 100 || m.Height != 50 {
		t.Errorf("expected (100,50), got (%d,%d)", m.Width, m.Height)
	}
}

func TestModelRecordCreateIntegration(t *testing.T) {
	db := setupTestDB(t)
	seedTestDB(t, db)
	m := tui.NewModel(db)

	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)
	if m.InputType != tui.InputTypeRecord {
		t.Fatalf("expected InputTypeRecord, got %v", m.InputType)
	}

	for i := 0; i < 3; i++ {
		result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = result.(tui.Model)
	}
	if m.RecordInput.ActiveField != tui.FieldCurrency {
		t.Fatalf("expected FieldCurrency, got %v", m.RecordInput.ActiveField)
	}

	result, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)
	if m.RecordInput.CurrencyInput.Selected == nil || m.RecordInput.CurrencyInput.Selected.Code != "USD" {
		t.Fatal("expected USD selected")
	}

	// Auto-advanced to Tags on enter, one tab to Amount
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = result.(tui.Model)
	if m.RecordInput.ActiveField != tui.FieldAmount {
		t.Fatalf("expected FieldAmount, got %v", m.RecordInput.ActiveField)
	}

	result, _ = m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: '.', Text: "."})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m = result.(tui.Model)

	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = result.(tui.Model)
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = result.(tui.Model)
	if m.RecordInput.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.RecordInput.ActiveField)
	}

	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)
	if m.RecordInput.State != tui.RecordStateConfirm {
		t.Fatalf("expected RecordStateConfirm, got %v", m.RecordInput.State)
	}
	if m.RecordInput.ConfirmModel.Errors.HasErrors() {
		t.Fatalf("expected no errors, got %v", m.RecordInput.ConfirmModel.Errors)
	}

	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)
	if m.RecordInput.State != tui.RecordStateSuccess {
		t.Fatalf("expected RecordStateSuccess, got %v", m.RecordInput.State)
	}

	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = result.(tui.Model)
	if m.RecordInput.State != tui.RecordStateEditing {
		t.Errorf("expected RecordStateEditing after Enter in success, got %v", m.RecordInput.State)
	}
	if m.RecordInput.ActiveField != tui.FieldDateYear {
		t.Errorf("expected FieldDateYear after reset, got %v", m.RecordInput.ActiveField)
	}
}

func TestModelEditFromQueryIntegration(t *testing.T) {
	db := setupTestDB(t)
	seedTestDB(t, db)
	svc := service.New(db)

	_, err := svc.CreateRecordWithTagsAndLinks(context.Background(), "2026-05-01", 1000, 1, "test record", 1, nil, nil)
	if err != nil {
		t.Fatalf("create record: %v", err)
	}

	m := tui.NewModel(db)

	// Navigate to query menu
	r, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = r.(tui.Model)
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.InputType != tui.InputTypeQuery {
		t.Fatalf("expected InputTypeQuery, got %v", m.InputType)
	}

	// Select "new" from query menu
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.QueryInput.State != tui.QueryStateFilterForm {
		t.Fatalf("expected QueryStateFilterForm, got %v", m.QueryInput.State)
	}

	// Set date range to cover our record
	m.QueryInput.DateFrom.SetValue("2026-05-01")
	m.QueryInput.DateTo.SetValue("2026-05-31")

	// Tab through to confirm: DateFrom → DateTo → Currency → Tags → Fuzzy → Confirm
	for i := 0; i < 5; i++ {
		r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = r.(tui.Model)
	}
	if m.QueryInput.State != tui.QueryStateConfirm {
		t.Fatalf("expected QueryStateConfirm, got %v", m.QueryInput.State)
	}

	// Execute query
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.QueryInput.State != tui.QueryStateResults {
		t.Fatalf("expected QueryStateResults, got %v", m.QueryInput.State)
	}
	if len(m.QueryInput.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(m.QueryInput.Results))
	}
	if m.QueryInput.Results[0].Notes != "test record" {
		t.Errorf("expected notes 'test record', got %q", m.QueryInput.Results[0].Notes)
	}

	// Enter on the record to edit
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.InputType != tui.InputTypeRecord {
		t.Fatalf("expected InputTypeRecord, got %v", m.InputType)
	}
	if m.RecordInput.Origin != tui.RecordOriginQuery {
		t.Errorf("expected RecordOriginQuery, got %v", m.RecordInput.Origin)
	}
	if m.RecordInput.AmountInput.Value() != "10.00" {
		t.Errorf("expected amount 10.00, got %q", m.RecordInput.AmountInput.Value())
	}
	if m.RecordInput.NotesInput.Value() != "test record" {
		t.Errorf("expected notes 'test record', got %q", m.RecordInput.NotesInput.Value())
	}

	// Tab to Notes and confirm
	for i := 0; i < 7; i++ {
		r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = r.(tui.Model)
	}
	if m.RecordInput.ActiveField != tui.FieldNotes {
		t.Fatalf("expected FieldNotes, got %v", m.RecordInput.ActiveField)
	}

	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.RecordInput.State != tui.RecordStateConfirm {
		t.Fatalf("expected RecordStateConfirm, got %v", m.RecordInput.State)
	}
	if m.RecordInput.ConfirmModel.Errors.HasErrors() {
		t.Fatalf("expected no errors, got %v", m.RecordInput.ConfirmModel.Errors)
	}

	// Enter to save
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = r.(tui.Model)
	if m.RecordInput.State != tui.RecordStateSuccess {
		t.Fatalf("expected RecordStateSuccess, got %v", m.RecordInput.State)
	}

	// Esc to return to results
	r, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = r.(tui.Model)
	if m.InputType != tui.InputTypeQuery {
		t.Fatalf("expected InputTypeQuery, got %v", m.InputType)
	}
	if m.QueryInput.State != tui.QueryStateResults {
		t.Errorf("expected QueryStateResults, got %v", m.QueryInput.State)
	}
	if len(m.QueryInput.Results) != 1 {
		t.Errorf("expected 1 result after refresh, got %d", len(m.QueryInput.Results))
	}
}
