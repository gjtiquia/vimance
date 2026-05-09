package tui

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type InputType string

const (
	InputTypeNone   InputType = "none"
	InputTypeText   InputType = "text"
	InputTypeList   InputType = "list"
	InputTypeRecord InputType = "record"
	InputTypeQuery  InputType = "query"
)

type Model struct {
	database    *sql.DB
	service     *service.Service
	history     []string
	inputChain  []string
	InputType   InputType
	textInput   textinput.Model
	listInput   list.Model
	RecordInput RecordModel
	QueryInput  QueryModel
	Width       int
	Height      int
}

func NewModel(database *sql.DB) Model {
	header := "vimance\n"
	history := []string{header}

	textInput := textinput.New()
	listInput := NewUnstyledList()
	svc := service.New(database)
	recordInput := NewRecordModel(svc)

	m := Model{
		database:    database,
		service:     svc,
		history:     history,
		textInput:   textInput,
		listInput:   listInput,
		RecordInput: recordInput,
		QueryInput:  NewQueryModel(svc),
	}

	m, _ = m.EnterListInput()
	return m
}

func (m Model) Init() tea.Cmd {
	if m.InputType == InputTypeText {
		return textinput.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.InputType == InputTypeQuery {
			m.QueryInput.SetPageSize(m.Height)
		}
	}

	switch m.InputType {
	case InputTypeText:
		return m.UpdateTextInput(msg)
	case InputTypeList:
		return m.UpdateListInput(msg)
	case InputTypeRecord:
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			if m.RecordInput.State == RecordStateEditing && !m.RecordInput.isInInsertMode() {
				return m.routeBackFromRecord()
			}
			if m.RecordInput.State == RecordStateSuccess {
				return m.routeBackFromRecord()
			}
		}
		return m.UpdateRecordInput(msg)
	case InputTypeQuery:
		return m.UpdateQueryInput(msg)
	}

	return m, nil
}

func (m Model) View() tea.View {
	var sb strings.Builder

	for _, s := range m.history {
		sb.WriteString(s)
	}

	switch m.InputType {
	case InputTypeText:
		sb.WriteString(m.textInput.View())
	case InputTypeList:
		sb.WriteString(m.listInput.View())
	case InputTypeRecord:
		sb.WriteString(m.RecordInput.View())
	case InputTypeQuery:
		if m.QueryInput.State == QueryStateMenu {
			if m.QueryInput.ErrorMsg != "" {
				sb.WriteString(fmt.Sprintf("Error: %s\n\n", m.QueryInput.ErrorMsg))
			}
			sb.WriteString(m.listInput.View())
		} else {
			sb.WriteString(m.QueryInput.View())
		}
	}

	return tea.NewView(sb.String())
}

func (m Model) routeBackFromRecord() (Model, tea.Cmd) {
	if m.RecordInput.Origin == RecordOriginQuery {
		m.RecordInput = NewRecordModel(m.service)
		m.InputType = InputTypeQuery
		m.QueryInput.RefreshResults()
		return m, nil
	}
	m.RecordInput = NewRecordModel(m.service)
	return m.EnterListInput()
}

func (m Model) EnterQueryInput() (Model, tea.Cmd) {
	m.InputType = InputTypeQuery
	m.QueryInput = NewQueryModel(m.service)
	if m.Height > 0 {
		m.QueryInput.SetPageSize(m.Height)
	}

	m.listInput.SetItems([]list.Item{
		NewListItem("new", "create a new query", "n"),
		NewListItem("saved", "use a saved query", "s"),
	})
	m.listInput.Title = "query:"
	m.listInput.ResetSelected()
	listHeight := 2 + 3 + 2 + 1 + 3
	m.listInput.SetHeight(listHeight)
	m.listInput.SetFilterText("")
	m.listInput.SetFilterState(list.Filtering)

	return m, nil
}

func (m Model) UpdateQueryInput(msg tea.Msg) (Model, tea.Cmd) {
	if m.QueryInput.State == QueryStateMenu {
		m.QueryInput.ErrorMsg = ""

		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				if m.listInput.FilterState() != list.Filtering {
					m.QueryInput = NewQueryModel(m.service)
					return m.EnterListInput()
				}
				m.listInput.SetFilterState(list.FilterApplied)
				m.listInput.Help.ShowAll = true
				return m, nil
			case "enter":
				if item, ok := visibleItem(m.listInput); ok {
					li := item.(ListItem)
					if li.title == "new" {
						m.QueryInput.State = QueryStateFilterForm
						m.QueryInput.ActiveField = FilterDateFrom
						m.QueryInput.FocusActiveField()
						return m, nil
					}
					if li.title == "saved" {
						m.QueryInput.loadSavedQueries()
						return m, nil
					}
				}
			case "up":
				if m.listInput.FilterState() == list.Filtering {
					m.listInput.CursorUp()
				}
				return m, nil
			case "down":
				if m.listInput.FilterState() == list.Filtering {
					m.listInput.CursorDown()
				}
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.listInput, cmd = m.listInput.Update(msg)
		m.listInput = clampListCursor(m.listInput)
		return m, cmd
	}

	var cmd tea.Cmd
	m.QueryInput, cmd = m.QueryInput.Update(msg)

	if m.QueryInput.SelectedID != 0 {
		id := m.QueryInput.SelectedID
		m.QueryInput.SelectedID = 0

		full, err := m.service.GetRecordFull(context.Background(), id)
		if err != nil {
			m.QueryInput.setError(fmt.Sprintf("Failed to load record: %v", err))
			return m, nil
		}
		m.InputType = InputTypeRecord
		m.RecordInput = NewEditRecordModel(m.service, full, RecordOriginQuery)
		return m, nil
	}

	return m, cmd
}
