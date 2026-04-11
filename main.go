package main

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"fmt"
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

	return model{
		history:   make([]string, 0),
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
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {

	s := "(this is a title)\n"

	if !m.quitting {
		s += m.textInput.View()
	} else {
		s += m.textInput.Prompt + m.textInput.Value() + "\n" // must add newline or else bubbletea wont render it if its the last line
	}

	// pass in a string to create a view
	return tea.NewView(s)
}
