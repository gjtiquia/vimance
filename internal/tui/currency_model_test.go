package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/tui"
)

func TestCurrencyModelInitialState(t *testing.T) {
	m := tui.NewCurrencyModel(nil)
	if m.Mode != tui.CurrencyModeInsert {
		t.Errorf("expected CurrencyModeInsert, got %v", m.Mode)
	}
	if m.Selected != nil {
		t.Error("expected no selected currency")
	}
	if m.ShouldAdvance {
		t.Error("expected ShouldAdvance=false initially")
	}
}

func TestCurrencyModelEnterNewCurrency(t *testing.T) {
	m := tui.NewCurrencyModel(nil)
	m.SearchInput.Focus()

	m, _ = m.Update(tea.KeyPressMsg{Code: 'E', Text: "E"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Selected == nil {
		t.Fatal("expected selected currency after typing and entering")
	}
	if m.Selected.Code != "EUR" {
		t.Errorf("expected EUR, got %s", m.Selected.Code)
	}
	if !m.Selected.IsNew {
		t.Error("expected IsNew=true for typed currency")
	}
	if !m.ShouldAdvance {
		t.Error("expected ShouldAdvance=true after enter")
	}
}

func TestCurrencyModelSelectFromList(t *testing.T) {
	m := tui.NewCurrencyModel(nil)
	m.AllCurrencies = []tui.CurrencyItem{
		{ID: 1, Code: "USD"},
		{ID: 2, Code: "EUR"},
	}
	m.Mode = tui.CurrencyModeNormal

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Selected == nil {
		t.Fatal("expected selected currency")
	}
	if m.Selected.Code != "EUR" {
		t.Errorf("expected EUR, got %s", m.Selected.Code)
	}
	if m.Selected.IsNew {
		t.Error("expected IsNew=false for existing currency")
	}
}

func TestCurrencyModelModeToggle(t *testing.T) {
	m := tui.NewCurrencyModel(nil)
	m.Mode = tui.CurrencyModeInsert

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.Mode != tui.CurrencyModeNormal {
		t.Errorf("expected CurrencyModeNormal after esc, got %v", m.Mode)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	if m.Mode != tui.CurrencyModeInsert {
		t.Errorf("expected CurrencyModeInsert after i, got %v", m.Mode)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if m.Mode != tui.CurrencyModeInsert {
		t.Errorf("expected CurrencyModeInsert after a, got %v", m.Mode)
	}
}

func TestCurrencyModelCursorUpDown(t *testing.T) {
	m := tui.NewCurrencyModel(nil)
	m.AllCurrencies = []tui.CurrencyItem{
		{ID: 1, Code: "USD"},
		{ID: 2, Code: "EUR"},
		{ID: 3, Code: "GBP"},
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CursorIndex != 1 {
		t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CursorIndex != 2 {
		t.Errorf("expected CursorIndex=2, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.CursorIndex != 2 {
		t.Errorf("expected CursorIndex=2 at bottom, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 1 {
		t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.CursorIndex != 0 {
		t.Errorf("expected CursorIndex=0 at top, got %d", m.CursorIndex)
	}
}
