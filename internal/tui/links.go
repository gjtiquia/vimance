package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	SelectedParents    []LinkedRecord
	SearchInput        textinput.Model
	AllCandidates      []LinkedRecord
	FilteredCandidates []LinkedRecord
	CursorIndex        int
	Mode               LinkMode
	DateFrom           string
	DateTo             string
	CurrencyID         int64
	Loaded             bool
	LoadErr            string
	service            *service.Service
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
		Loaded:             false,
		LoadErr:            "",
	}
}

func (m *LinksModel) SetDateRange(year, month string) {
	if year == "" || month == "" {
		return
	}
	if len(year) != 4 || len(month) < 1 || len(month) > 2 {
		return
	}
	m.DateFrom = fmt.Sprintf("%s-%s-01", year, month)

	y, errY := strconv.Atoi(year)
	mo, errM := strconv.Atoi(month)
	if errY != nil || errM != nil || y < 1 || mo < 1 || mo > 12 {
		return
	}
	lastDay := time.Date(y, time.Month(mo)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	m.DateTo = fmt.Sprintf("%s-%s-%02d", year, month, lastDay)
}

func (m *LinksModel) SetCurrencyID(id int64) {
	m.CurrencyID = id
}

func (m *LinksModel) LoadCandidates(ctx context.Context, excludeID int64) error {
	if m.DateFrom == "" || m.DateTo == "" || m.CurrencyID == 0 {
		m.AllCandidates = make([]LinkedRecord, 0)
		m.FilteredCandidates = make([]LinkedRecord, 0)
		m.Loaded = false
		m.LoadErr = ""
		return nil
	}

	candidates, err := m.service.SearchLinkCandidates(ctx, m.DateFrom, m.DateTo, m.CurrencyID, excludeID)
	if err != nil {
		m.AllCandidates = make([]LinkedRecord, 0)
		m.FilteredCandidates = make([]LinkedRecord, 0)
		m.Loaded = false
		m.LoadErr = err.Error()
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
	m.Loaded = true
	m.LoadErr = ""
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
	case tea.KeyPressMsg:
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
				filtered := m.FilteredCandidates
				if len(filtered) > 0 && m.CursorIndex < len(filtered) {
					m.addParent(filtered[m.CursorIndex])
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
			sb.WriteString(fmt.Sprintf("[%s]", Truncate(p.Notes, 20)))
		}
	}
	sb.WriteString("\n")

	if m.Loaded && m.DateFrom != "" && m.DateTo != "" {
		sb.WriteString(fmt.Sprintf("(showing records from %s to %s)\n", m.DateFrom, m.DateTo))
	}

	sb.WriteString(m.SearchInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(m.filteredListView())

	return sb.String()
}

func (m LinksModel) filteredListView() string {
	filtered := m.FilteredCandidates

	if m.LoadErr != "" {
		return fmt.Sprintf("  (error loading candidates: %s)\n", m.LoadErr)
	}

	if !m.Loaded {
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

		amountStr := service.FormatCents(c.AmountCents)

		tags := strings.Join(c.TagNames, ", ")

		sb.WriteString(fmt.Sprintf("%s %d) %s  $%s  \"%s\"", cursor, i+1, c.Date, amountStr, Truncate(c.Notes, 30)))
		if tags != "" {
			sb.WriteString(fmt.Sprintf("  [%s]", tags))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
