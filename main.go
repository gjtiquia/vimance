package main

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
)

func main() {
	initialModel := createModel()
	p := tea.NewProgram(initialModel)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("error: %v", err)
		return;
	}

	finalView := finalModel.View().Content;
	fmt.Println(finalView)
	fmt.Println("\nexiting gracefully...")
}

// tea.Model interface super simple: Init, Update, View
// - Init: what command to run on init
// - Update: how the model updates on message
// - View: how to render based on the model
// everything else is just deciding how to store state
type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{} // using map like a set of chosen indices
}

func createModel() model {
	return model{
		choices:  []string{"first", "second", "third"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// initial command to run
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {

	s := "hello bubbletea"

	// pass in a string to create a view
	return tea.NewView(s)
}
