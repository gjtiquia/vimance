package tui

import (
	"database/sql"
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
	}

	switch m.inputType {
	case InputTypeText:
		return m.UpdateTextInput(msg)
	case InputTypeList:
		return m.UpdateListInput(msg)
	case InputTypeRecord:
		if m.recordInput.State == RecordStateSuccess {
			if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
				m.recordInput = NewRecordModel(m.service)
				return m.EnterListInput()
			}
		}
		return m.UpdateRecordInput(msg)
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
	}

	return tea.NewView(sb.String())
}
