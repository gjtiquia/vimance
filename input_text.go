package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

const InputTypeText InputType = "text"

func (m Model) EnterTextInput() Model {
	m.inputType = InputTypeText
	m.textInput.SetValue("")
	m.textInput.Focus()
	return m
}

func (m Model) UpdateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc", "q":
			if m.textInput.Focused() {
				m.textInput.Blur()
			} else {
				return m, tea.Quit
			}

		case "enter":
			value := strings.TrimSpace(m.textInput.Value())
			if value == "" {
				break
			}

			// blur focus after enter
			m.textInput.Blur()

			textinputRender := m.textInput.Prompt + value + "\n"
			m.history = append(m.history, textinputRender)

			// TODO : for now, swap between
			return m.EnterListInput()
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
