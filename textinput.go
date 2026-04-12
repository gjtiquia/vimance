package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

const InputTypeText InputType = "text"

func (m Model) EnterTextInput() Model {
	m.userInputType = InputTypeText
	m.userTextInput.SetValue("")
	m.userTextInput.Focus()
	return m
}

func (m Model) UpdateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc", "q":
			if m.userTextInput.Focused() {
				m.userTextInput.Blur()
			} else {
				return m, tea.Quit
			}

		case "enter":
			value := strings.TrimSpace(m.userTextInput.Value())
			if value == "" {
				break
			}

			// blur focus after enter
			m.userTextInput.Blur()

			textinputRender := m.userTextInput.Prompt + value + "\n"
			m.history = append(m.history, textinputRender)

			// TODO : for now, swap between
			return m.EnterListInput()
		}
	}

	var cmd tea.Cmd
	m.userTextInput, cmd = m.userTextInput.Update(msg)
	return m, cmd
}
