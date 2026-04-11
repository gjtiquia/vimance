package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
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
	history     []string
	userInput   textinput.Model
	userOptions list.Model
}

type Option string

// list.Item INTERFACE
func (i Option) FilterValue() string { return string(i) }

type OptionDelegate struct{}

// list.ItemDelegate INTERFACE
func (d OptionDelegate) Height() int                             { return 1 }
func (d OptionDelegate) Spacing() int                            { return 0 }
func (d OptionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d OptionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Option)
	if !ok {
		return
	}

	var cursor string
	if index == m.Index() {
		cursor = ">"
	} else {
		cursor = " "
	}

	str := fmt.Sprintf("%s %d) %s", cursor, index+1, item)
	fmt.Fprint(w, str)
}

func createModel() model {
	s := "(this is a title)\n"
	h := []string{s}

	ti := textinput.New()
	ti.Focus() // must focus or else will not accept any user input

	options := []list.Item{
		Option("first"),
		Option("second"),
		Option("third"),
	}

	const listWidth = 20
	listHeight := len(options) + 4 // title + status bar + pagiation (2) + help

	optionsList := list.New(options, OptionDelegate{}, listWidth, listHeight)
	optionsList.Title = "Commands:"
	optionsList.Styles = list.Styles{}  // reset styling (see list.DefaultStyles)
	optionsList.SetShowStatusBar(false) // shows item count
	// optionsList.SetShowPagination(false) // we will make sure all is shown anyways
	// optionsList.SetShowHelp(false)

	return model{
		history:     h,
		userInput:   ti,
		userOptions: optionsList,
	}
}

// tea.Model INTERFACE
func (m model) Init() tea.Cmd {
	// initial command to run
	return textinput.Blink // starts the blink timer
}

// tea.Model INTERFACE
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if !m.userInput.Focused() && m.userOptions.FilterState() != list.Filtering {
				return m, tea.Quit
			}

			if m.userInput.Focused() {
				m.userInput.Blur()
			}

		case "enter":
			textinputRender := m.userInput.Prompt + m.userInput.Value() + "\n"
			m.history = append(m.history, textinputRender)
			m.userInput.SetValue("")

		}
	}

	var inputCmd tea.Cmd
	m.userInput, inputCmd = m.userInput.Update(msg)

	var optionsCmd tea.Cmd
	m.userOptions, optionsCmd = m.userOptions.Update(msg)

	return m, tea.Batch(inputCmd, optionsCmd)
}

// tea.Model INTERFACE
func (m model) View() tea.View {

	var sb strings.Builder

	for _, s := range m.history {
		sb.WriteString(s)
	}

	sb.WriteString(m.userInput.View())
	sb.WriteString("\n")

	sb.WriteString(m.userOptions.View())

	// pass in a string to create a view
	return tea.NewView(sb.String())
}
