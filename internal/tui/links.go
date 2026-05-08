package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type LinkMode string

const (
	LinkModeInsert LinkMode = "insert"
	LinkModeNormal LinkMode = "normal"
)

type LinkedRecord struct {
	ID          int64
	Date        string
	AmountCents int64
	CurrencyID  int64
	Notes       string
	TagNames    []string
}

type LinksModel struct {
	SelectedParents   []LinkedRecord
	SearchInput       textinput.Model
	AllCandidates     []LinkedRecord
	FilteredCandidates []LinkedRecord
	CursorIndex       int
	Mode              LinkMode
	DateFrom          string
	DateTo            string
	CurrencyID        int64
	service           *service.Service
	loaded            bool
}

func NewLinksModel(svc *service.Service) LinksModel {
	textInput := textinput.New()
	textInput.Prompt = "Type: "

	return LinksModel{
		SelectedParents:    make([]LinkedRecord, 0),
		SearchInput:        textInput,
		AllCandidates:      make([]LinkedRecord, 0),
		FilteredCandidates: make([]LinkedRecord, 0),
		CursorIndex:        0,
		Mode:               LinkModeInsert,
		DateFrom:           "",
		DateTo:             "",
		CurrencyID:         0,
		service:            svc,
		loaded:             false,
	}
}

func (m *LinksModel) SetDateRange(year, month string) {
	if year == "" || month == "" {
		return
	}
	m.DateFrom = fmt.Sprintf("%s-%s-01", year, month)
	m.DateTo = fmt.Sprintf("%s-%s-31", year, month)
}

func (m *LinksModel) SetCurrencyID(id int64) {
	m.CurrencyID = id
}

func (m *LinksModel) LoadCandidates(ctx context.Context) error {
	if m.DateFrom == "" || m.DateTo == "" || m.CurrencyID == 0 {
		m.AllCandidates = make([]LinkedRecord, 0)
		m.FilteredCandidates = make([]LinkedRecord, 0)
		m.loaded = true
		return nil
	}

	candidates, err := m.service.SearchLinkCandidates(ctx, m.DateFrom, m.DateTo, m.CurrencyID, 0)
	if err != nil {
		return err
	}

	m.AllCandidates = make([]LinkedRecord, len(candidates))
	for i, c := range candidates {
		m.AllCandidates[i] = LinkedRecord{
			ID:          c.ID,
			Date:        c.Date,
			AmountCents: c.AmountCents,
			CurrencyID:  c.CurrencyID,
			Notes:       c.Notes,
			TagNames:    c.TagNames,
		}
	}
	m.FilteredCandidates = m.filterCandidates()
	m.loaded = true
	return nil
}

func (m *LinksModel) filterCandidates() []LinkedRecord {
	input := strings.TrimSpace(m.SearchInput.Value())
	if input == "" {
		return m.AllCandidates
	}

	inputLower := strings.ToLower(input)

	var filtered []LinkedRecord
	for _, c := range m.AllCandidates {
		searchStr := strings.ToLower(c.Notes)
		for _, tag := range c.TagNames {
			searchStr += " " + strings.ToLower(tag)
		}

		if strings.Contains(searchStr, inputLower) {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

func (m *LinksModel) addParent(record LinkedRecord) {
	for _, p := range m.SelectedParents {
		if p.ID == record.ID {
			return
		}
	}

	m.SelectedParents = append(m.SelectedParents, record)
	m.SearchInput.SetValue("")
	m.FilteredCandidates = m.filterCandidates()
}

func (m *LinksModel) removeLastParent() {
	if len(m.SelectedParents) > 0 {
		m.SelectedParents = m.SelectedParents[:len(m.SelectedParents)-1]
	}
}

func (m LinksModel) Update(msg tea.Msg) (LinksModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Mode {
		case LinkModeInsert:
			switch msg.String() {
			case "esc":
				m.Mode = LinkModeNormal
				m.SearchInput.Blur()
				if len(m.FilteredCandidates) > 0 {
					m.CursorIndex = 0
				}
				return m, nil
			case "enter":
				input := strings.TrimSpace(m.SearchInput.Value())
				if input != "" {
					filtered := m.FilteredCandidates
					if len(filtered) > 0 && m.CursorIndex < len(filtered) {
						m.addParent(filtered[m.CursorIndex])
					}
				}
				return m, nil
			case "up":
				if m.CursorIndex > 0 {
					m.CursorIndex--
				}
				return m, nil
			case "down":
				if len(m.FilteredCandidates) > 0 && m.CursorIndex < len(m.FilteredCandidates)-1 {
					m.CursorIndex++
				}
				return m, nil
			case "ctrl+z":
				m.removeLastParent()
				return m, nil
			case "tab":
				return m, nil
			}
		case LinkModeNormal:
			switch msg.String() {
			case "i", "a":
				m.Mode = LinkModeInsert
				m.SearchInput.Focus()
				m.FilteredCandidates = m.filterCandidates()
				return m, nil
			case "j", "down":
				if len(m.FilteredCandidates) > 0 && m.CursorIndex < len(m.FilteredCandidates)-1 {
					m.CursorIndex++
				}
				return m, nil
			case "k", "up":
				if m.CursorIndex > 0 {
					m.CursorIndex--
				}
				return m, nil
			case "enter":
				if len(m.FilteredCandidates) > 0 && m.CursorIndex < len(m.FilteredCandidates) {
					m.addParent(m.FilteredCandidates[m.CursorIndex])
				}
				return m, nil
			case "ctrl+z":
				m.removeLastParent()
				return m, nil
			case "tab":
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	m.FilteredCandidates = m.filterCandidates()
	return m, cmd
}

func (m LinksModel) View() string {
	var sb strings.Builder

	sb.WriteString("Links: ")
	if len(m.SelectedParents) == 0 {
		sb.WriteString("(none)")
	} else {
		for i, p := range m.SelectedParents {
			if i > 0 {
				sb.WriteString(", ")
			}
			notes := p.Notes
			if len(notes) > 20 {
				notes = notes[:20] + "..."
			}
			sb.WriteString(fmt.Sprintf("[%s]", notes))
		}
	}
	sb.WriteString("\n")

	if m.loaded && m.DateFrom != "" && m.DateTo != "" {
		sb.WriteString(fmt.Sprintf("(showing records from %s to %s)\n", m.DateFrom, m.DateTo))
	}

	sb.WriteString(m.SearchInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(m.filteredListView())

	return sb.String()
}

func (m LinksModel) filteredListView() string {
	filtered := m.FilteredCandidates

	if !m.loaded {
		return "  (enter date and currency first to see candidates)\n"
	}

	if len(filtered) == 0 {
		return "  (no records found in this date range and currency)\n"
	}

	var sb strings.Builder
	for i, c := range filtered {
		cursor := " "
		if i == m.CursorIndex {
			cursor = ">"
		}

		cents := c.AmountCents
		dollars := cents / 100
		centsRemainder := cents % 100
		amountStr := fmt.Sprintf("%d.%02d", dollars, centsRemainder)

		notes := c.Notes
		if len(notes) > 30 {
			notes = notes[:30] + "..."
		}

		tags := strings.Join(c.TagNames, ", ")

		sb.WriteString(fmt.Sprintf("%s %d) %s  $%s  \"%s\"", cursor, i+1, c.Date, amountStr, notes))
		if tags != "" {
			sb.WriteString(fmt.Sprintf("  [%s]", tags))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
