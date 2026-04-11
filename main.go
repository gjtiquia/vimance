package main

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"fmt"
)

func main() {
	initialModel := createModel()
	p := tea.NewProgram(initialModel)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	// dont clear screen
	finalView := finalModel.View().Content
	fmt.Println(finalView)

	fmt.Println("\nexiting gracefully...")
}

// tea.Model interface super simple: Init, Update, View
// - Init: what command to run on init
// - Update: how the model updates on message
// - View: how to render based on the model
// everything else is just deciding how to store state
type model struct {
	history   []string
	textInput textinput.Model
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
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {

	s := m.textInput.View()

	// pass in a string to create a view
	return tea.NewView(s)
}
