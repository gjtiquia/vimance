package tui

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type InputType string

const (
	InputTypeNone    InputType = "none"
	InputTypeText    InputType = "text"
	InputTypeList    InputType = "list"
	InputTypeRecord  InputType = "record"
	InputTypeQuery   InputType = "query"
	InputTypeTargets InputType = "targets"
)

type Model struct {
	database     *sql.DB
	service      *service.Service
	history      []string
	InputType     InputType
	textInput    textinput.Model
	menuInput    MenuModel
	RecordInput  RecordModel
	QueryInput   QueryModel
	TargetsInput TargetsModel
	Width        int
	Height       int
}

func NewModel(database *sql.DB) Model {
	header := "vimance\n"
	history := []string{header}

	textInput := textinput.New()
	svc := service.New(database)
	recordInput := NewRecordModel(svc)

	m := Model{
		database:     database,
		service:      svc,
		history:      history,
		textInput:    textInput,
		RecordInput:  recordInput,
		QueryInput:   NewQueryModel(svc),
		TargetsInput: NewTargetsModel(svc),
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
	case InputTypeTargets:
		var cmd tea.Cmd
		m.TargetsInput, cmd = m.TargetsInput.Update(msg)
		return m, cmd
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
		sb.WriteString(m.menuInput.View())
	case InputTypeRecord:
		sb.WriteString(m.RecordInput.View())
	case InputTypeQuery:
		if m.QueryInput.State == QueryStateMenu {
			if m.QueryInput.ErrorMsg != "" {
				sb.WriteString(fmt.Sprintf("Error: %s\n\n", m.QueryInput.ErrorMsg))
			}
			sb.WriteString(m.menuInput.View())
		} else {
			sb.WriteString(m.QueryInput.View())
		}
	case InputTypeTargets:
		sb.WriteString(m.TargetsInput.View())
	}

	return tea.NewView(sb.String())
}

func (m Model) EnterListInput() (Model, tea.Cmd) {
	m.InputType = InputTypeList
	m.menuInput = NewMenuModel("commands:", []MenuItem{
		{Title: "create", Desc: "create a new record"},
		{Title: "query", Desc: "query existing records"},
		{Title: "targets", Desc: "view targets vs actuals"},
	})
	return m, nil
}

func (m Model) UpdateListInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case "up", "k", "shift+tab":
			m.menuInput.CursorUp()
			return m, nil
		case "down", "j", "tab":
			m.menuInput.CursorDown()
			return m, nil
		case "enter":
			if item, ok := m.menuInput.SelectedItem(); ok {
				m.history = append(m.history, m.textInput.Prompt+item.Title+"\n")
				return m.routeListSelection(item.Title)
			}
		default:
			if n := numberKey(msg.String()); n >= 0 {
				if m.menuInput.SelectByIndex(n) {
					if item, ok := m.menuInput.SelectedItem(); ok {
						m.history = append(m.history, m.textInput.Prompt+item.Title+"\n")
						return m.routeListSelection(item.Title)
					}
				}
			}
		}
	}
	return m, nil
}

func (m Model) routeListSelection(title string) (Model, tea.Cmd) {
	switch title {
	case "create":
		return m.EnterRecordInput()
	case "query":
		return m.EnterQueryInput()
	case "targets":
		return m.EnterTargetsInput()
	}
	return m.EnterListInput()
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

	m.menuInput = NewMenuModel("query:", []MenuItem{
		{Title: "new", Desc: "create a new query"},
		{Title: "saved", Desc: "use a saved query"},
	})

	m.QueryInput.State = QueryStateMenu
	return m, nil
}

func (m Model) EnterTargetsInput() (Model, tea.Cmd) {
	m.InputType = InputTypeTargets
	m.TargetsInput = NewTargetsModel(m.service)
	return m, m.TargetsInput.loadTargets()
}

func (m Model) UpdateQueryInput(msg tea.Msg) (Model, tea.Cmd) {
	if m.QueryInput.State == QueryStateMenu {
		m.QueryInput.ErrorMsg = ""

		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.QueryInput = NewQueryModel(m.service)
				return m.EnterListInput()
			case "up", "k", "shift+tab":
				m.menuInput.CursorUp()
				return m, nil
			case "down", "j", "tab":
				m.menuInput.CursorDown()
				return m, nil
			case "enter":
				if item, ok := m.menuInput.SelectedItem(); ok {
					return m.routeQuerySelection(item.Title)
				}
			default:
				if n := numberKey(msg.String()); n >= 0 {
					if m.menuInput.SelectByIndex(n) {
						if item, ok := m.menuInput.SelectedItem(); ok {
							return m.routeQuerySelection(item.Title)
						}
					}
				}
			}
		}

		return m, nil
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

func (m Model) routeQuerySelection(title string) (Model, tea.Cmd) {
	if title == "new" {
		m.QueryInput.State = QueryStateFilterForm
		m.QueryInput.ActiveField = FilterDateFrom
		m.QueryInput.FocusActiveField()
		return m, nil
	}
	if title == "saved" {
		var cmd tea.Cmd
		m.QueryInput, cmd = m.QueryInput.loadSavedQueries()
		return m, cmd
	}
	return m, nil
}

func numberKey(s string) int {
	if len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
		return int(s[0] - '1')
	}
	return -1
}
