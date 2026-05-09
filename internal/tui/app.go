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
	inputType   InputType
	textInput   textinput.Model
	listInput   list.Model
	recordInput RecordModel
	queryInput  QueryModel
	width       int
	height      int
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
		recordInput: recordInput,
		queryInput:  NewQueryModel(svc),
	}

	m, _ = m.EnterListInput()
	return m
}

func (m Model) Init() tea.Cmd {
	if m.inputType == InputTypeText {
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
		m.width = msg.Width
		m.height = msg.Height
		if m.inputType == InputTypeQuery {
			m.queryInput.SetPageSize(m.height)
		}
	}

	switch m.inputType {
	case InputTypeText:
		return m.UpdateTextInput(msg)
	case InputTypeList:
		return m.UpdateListInput(msg)
	case InputTypeRecord:
		if m.recordInput.State == RecordStateSuccess {
			if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
				if m.recordInput.Origin == RecordOriginQuery {
					m.recordInput = NewRecordModel(m.service)
					m.inputType = InputTypeQuery
					m.queryInput.RefreshResults()
					return m, nil
				}
				m.recordInput = NewRecordModel(m.service)
				return m.EnterListInput()
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

	switch m.inputType {
	case InputTypeText:
		sb.WriteString(m.textInput.View())
	case InputTypeList:
		sb.WriteString(m.listInput.View())
	case InputTypeRecord:
		sb.WriteString(m.recordInput.View())
	case InputTypeQuery:
		if m.queryInput.State == QueryStateMenu {
			if m.queryInput.errorMsg != "" {
				sb.WriteString(fmt.Sprintf("Error: %s\n\n", m.queryInput.errorMsg))
			}
			sb.WriteString(m.listInput.View())
		} else {
			sb.WriteString(m.queryInput.View())
		}
	}

	return tea.NewView(sb.String())
}

func (m Model) EnterQueryInput() (Model, tea.Cmd) {
	m.inputType = InputTypeQuery
	m.queryInput = NewQueryModel(m.service)
	if m.height > 0 {
		m.queryInput.SetPageSize(m.height)
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
	if m.queryInput.State == QueryStateMenu {
		m.queryInput.errorMsg = ""

		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				if m.listInput.FilterState() != list.Filtering {
					m.queryInput = NewQueryModel(m.service)
					return m.EnterListInput()
				}
				m.listInput.SetFilterState(list.FilterApplied)
				m.listInput.Help.ShowAll = true
				return m, nil
			case "enter":
				visibleItems := m.listInput.VisibleItems()
				visibleIndex := m.listInput.Index()
				if len(visibleItems) > 0 {
					item := visibleItems[visibleIndex].(ListItem)
					if item.title == "new" {
						m.queryInput.State = QueryStateFilterForm
						m.queryInput.ActiveField = FilterDateFrom
						m.queryInput.focusActiveField()
						return m, nil
					}
					if item.title == "saved" {
						m.queryInput.loadSavedQueries()
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
		return m, cmd
	}

	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)

	if m.queryInput.selectedID != 0 {
		id := m.queryInput.selectedID
		m.queryInput.selectedID = 0

		full, err := m.service.GetRecordFull(context.Background(), id)
		if err != nil {
			m.queryInput.setError(fmt.Sprintf("Failed to load record: %v", err))
			return m, nil
		}
		m.inputType = InputTypeRecord
		m.recordInput = NewEditRecordModel(m.service, full, RecordOriginQuery)
		return m, nil
	}

	return m, cmd
}
