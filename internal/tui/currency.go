package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type CurrencyItem struct {
	ID    int64
	Code  string
	IsNew bool
}

type CurrencyMode string

const (
	CurrencyModeInsert CurrencyMode = "insert"
	CurrencyModeNormal CurrencyMode = "normal"
)

type CurrencyModel struct {
	Selected      *CurrencyItem
	SearchInput   textinput.Model
	AllCurrencies []CurrencyItem
	CursorIndex   int
	Mode          CurrencyMode
	ShouldAdvance bool
	service       *service.Service
}

func NewCurrencyModel(svc *service.Service) CurrencyModel {
	textInput := textinput.New()
	textInput.Prompt = "Currency: "

	return CurrencyModel{
		Selected:      nil,
		SearchInput:   textInput,
		AllCurrencies: make([]CurrencyItem, 0),
		CursorIndex:   0,
		Mode:          CurrencyModeInsert,
		ShouldAdvance: false,
		service:       svc,
	}
}

func (m *CurrencyModel) LoadCurrencies(ctx context.Context) error {
	currencies, err := m.service.ListCurrencies(ctx)
	if err != nil {
		return err
	}

	m.AllCurrencies = make([]CurrencyItem, 0, len(currencies))
	for _, c := range currencies {
		m.AllCurrencies = append(m.AllCurrencies, CurrencyItem{
			ID:    c.ID,
			Code:  c.Code,
			IsNew: false,
		})
	}

	return nil
}

func (m CurrencyModel) getFilteredCurrencies() []CurrencyItem {
	input := strings.TrimSpace(m.SearchInput.Value())
	if input == "" {
		return m.AllCurrencies
	}

	var filtered []CurrencyItem
	for _, currency := range m.AllCurrencies {
		if strings.Contains(strings.ToUpper(currency.Code), strings.ToUpper(input)) {
			filtered = append(filtered, currency)
		}
	}

	return filtered
}

func (m *CurrencyModel) selectCurrency(code string) {
	upperCode := strings.ToUpper(code)

	for _, c := range m.AllCurrencies {
		if c.Code == upperCode {
			m.Selected = &c
			m.SearchInput.SetValue("")
			return
		}
	}

	m.Selected = &CurrencyItem{
		ID:    0,
		Code:  upperCode,
		IsNew: true,
	}
	m.SearchInput.SetValue("")
}

func (m *CurrencyModel) clearSelection() {
	m.Selected = nil
}

func (m CurrencyModel) Update(msg tea.Msg) (CurrencyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Mode {
		case CurrencyModeInsert:
			switch msg.String() {
			case "esc":
				m.Mode = CurrencyModeNormal
				m.SearchInput.Blur()
				if len(m.getFilteredCurrencies()) > 0 {
					m.CursorIndex = 0
				}
				return m, nil
			case "enter":
				input := strings.TrimSpace(m.SearchInput.Value())
				if input != "" {
					m.selectCurrency(input)
					m.ShouldAdvance = true
				} else {
					filtered := m.getFilteredCurrencies()
					if len(filtered) > 0 && m.CursorIndex < len(filtered) {
						m.selectCurrency(filtered[m.CursorIndex].Code)
						m.ShouldAdvance = true
					}
				}
				return m, nil
			case "up", "k":
				if m.CursorIndex > 0 {
					m.CursorIndex--
				}
				return m, nil
			case "down", "j":
				filtered := m.getFilteredCurrencies()
				if len(filtered) > 0 && m.CursorIndex < len(filtered)-1 {
					m.CursorIndex++
				}
				return m, nil
			}
		case CurrencyModeNormal:
			switch msg.String() {
			case "i", "a":
				m.Mode = CurrencyModeInsert
				m.SearchInput.Focus()
				return m, nil
			case "up", "k":
				if m.CursorIndex > 0 {
					m.CursorIndex--
				}
				return m, nil
			case "down", "j":
				filtered := m.getFilteredCurrencies()
				if len(filtered) > 0 && m.CursorIndex < len(filtered)-1 {
					m.CursorIndex++
				}
				return m, nil
			case "enter":
				filtered := m.getFilteredCurrencies()
				if len(filtered) > 0 && m.CursorIndex < len(filtered) {
					m.selectCurrency(filtered[m.CursorIndex].Code)
					m.ShouldAdvance = true
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	return m, cmd
}

func (m CurrencyModel) View() string {
	var sb strings.Builder

	sb.WriteString("Currency: ")
	if m.Selected != nil {
		sb.WriteString(m.Selected.Code)
		if m.Selected.IsNew {
			sb.WriteString("*")
		}
	}
	sb.WriteString("\n")

	sb.WriteString(m.SearchInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(m.filteredCurrenciesView())

	return sb.String()
}

func (m CurrencyModel) filteredCurrenciesView() string {
	filtered := m.getFilteredCurrencies()

	var sb strings.Builder

	if m.Selected != nil {
		sb.WriteString(fmt.Sprintf("Selected: %s", m.Selected.Code))
		if m.Selected.IsNew {
			sb.WriteString(" (new)")
		}
		sb.WriteString("\n\n")
	}

	if len(filtered) == 0 {
		sb.WriteString("  (no currencies found - type to create new)\n")
		return sb.String()
	}

	for i, currency := range filtered {
		cursor := " "
		if i == m.CursorIndex {
			cursor = ">"
		}

		sb.WriteString(fmt.Sprintf("%s %d) %s\n", cursor, i+1, currency.Code))
	}

	return sb.String()
}
