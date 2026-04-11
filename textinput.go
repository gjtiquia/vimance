package main

import tea "charm.land/bubbletea/v2"

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
			textinputRender := m.userTextInput.Prompt + m.userTextInput.Value() + "\n"
			m.history = append(m.history, textinputRender)
			m.userTextInput.SetValue("") // TODO : this should be on enter

			// TODO : for now, swap between
			m.userInputType = InputTypeList
		}
	}

	var cmd tea.Cmd
	m.userTextInput, cmd = m.userTextInput.Update(msg)
	return m, cmd
}
