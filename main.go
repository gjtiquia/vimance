package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func main() {
	m := createModel()
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Println("[exiting gracefully...]")
}

// tea.Model interface super simple: Init, Update, View
// - Init: what command to run on init
// - Update: how the model updates on message
// - View: how to render based on the model
// everything else is just deciding how to store state
type model struct {
	history   []string
	textInput textinput.Model
	quitting  bool
}

func createModel() model {
	ti := textinput.New()
	ti.Focus() // must focus or else will not accept any user input

	s := "(this is a title)\n"
	h := []string{s}

	return model{
		history:   h,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	// initial command to run
	return textinput.Blink // starts the blink timer
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			textinputRender := m.textInput.Prompt + m.textInput.Value() + "\n"
			m.history = append(m.history, textinputRender)
			m.textInput.SetValue("")
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {

	var sb strings.Builder

	for _, s := range m.history {
		sb.WriteString(s)
	}

	sb.WriteString(m.textInput.View())

	// pass in a string to create a view
	return tea.NewView(sb.String())
}
