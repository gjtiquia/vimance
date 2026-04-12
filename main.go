package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func main() {
	m := NewModel()
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Println("\n[exiting gracefully...]")
}

type InputType string

const InputTypeNone InputType = "none"

// tea.Model interface super simple: Init, Update, View
// - Init: what command to run on init
// - Update: how the Model updates on message
// - View: how to render based on the Model
// everything else is just deciding how to store state
type Model struct {
	history     []string
	inputChain  []string
	inputType   InputType
	textInput   textinput.Model
	listInput   list.Model
	recordInput RecordModel
}

func NewModel() Model {
	header := "vimance\n"
	history := []string{header}

	textInput := textinput.New()
	listInput := NewUnstyledList()
	recordInput := NewRecordModel()

	m := Model{
		history:     history,
		textInput:   textInput,
		listInput:   listInput,
		recordInput: recordInput,
	}

	m, _ = m.EnterListInput()
	return m
}

// tea.Model INTERFACE
func (m Model) Init() tea.Cmd {
	if m.inputType == InputTypeText {
		return textinput.Blink // starts the blink timer
	}
	return nil
}

// tea.Model INTERFACE
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
		return m.UpdateRecordInput(msg)
	}

	return m, nil
}

// tea.Model INTERFACE
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

	// pass in a string to create a view
	return tea.NewView(sb.String())
}
