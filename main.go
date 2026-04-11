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

	fmt.Println("[exiting gracefully...]")
}

type InputType int

const (
	InputTypeNone InputType = iota
	InputTypeText
	InputTypeList
)

// tea.Model interface super simple: Init, Update, View
// - Init: what command to run on init
// - Update: how the Model updates on message
// - View: how to render based on the Model
// everything else is just deciding how to store state
type Model struct {
	history       []string
	userInputType InputType
	userTextInput textinput.Model
	userListInput list.Model
}

func NewModel() Model {
	header := "vimance\n"
	history := []string{header}

	userTextInput := textinput.New()

	// TODO : this should be sub command and also updates the keybinds dynamically
	// TODO : https://github.com/charmbracelet/bubbletea/blob/main/examples/help/main.go
	userListInput := NewUnstyledList([]list.Item{
		NewListItem("create", "create a new record", "c", "new", "n"),
		NewListItem("query", "query existing records", "q", "list", "ls", "l"),
		NewListItem("test", "test", "t", "e"),
	})

	m := Model{
		history:       history,
		userTextInput: userTextInput,
		userListInput: userListInput,
	}

	m = m.EnterListInput()
	// m = m.EnterTextInput()
	return m
}

// tea.Model INTERFACE
func (m Model) Init() tea.Cmd {
	if m.userInputType == InputTypeText {
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

	switch m.userInputType {
	case InputTypeText:
		return m.UpdateTextInput(msg)
	case InputTypeList:
		return m.UpdateListInput(msg)
	}

	return m, nil
}

// tea.Model INTERFACE
func (m Model) View() tea.View {
	var sb strings.Builder

	for _, s := range m.history {
		sb.WriteString(s)
	}

	switch m.userInputType {
	case InputTypeText:
		sb.WriteString(m.userTextInput.View())
	case InputTypeList:
		sb.WriteString(m.userListInput.View())
	}

	// pass in a string to create a view
	return tea.NewView(sb.String())
}
