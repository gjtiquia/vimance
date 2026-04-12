package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const InputTypeRecord InputType = "record"

type RecordModel struct {
	DateYearInput  textinput.Model
	DateMonthInput textinput.Model
	DateDayInput   textinput.Model

	// TODO :
	// - tags
	// - currency
	// - amount
	// - notes
}

// TODO : maybe can add "category" support in the future, but its independent of the records, its more of a "collection of tags", but i think "queries" can handle that tho

func NewRecordModel() RecordModel {

	yearInput := textinput.New()
	yearInput.Prompt = "Year: "
	yearInput.Placeholder = "2026" // TODO : this should auto-default to today's date
	yearInput.CharLimit = 4
	yearInput.SetWidth(4) // required or else placeholder gets truncated to width(0) + 1 = 1 char

	monthInput := textinput.New()
	monthInput.Prompt = "Month: "
	monthInput.Placeholder = "04" // TODO : this should auto-default to today's date
	monthInput.CharLimit = 2
	monthInput.SetWidth(2)

	dayInput := textinput.New()
	dayInput.Prompt = "Day: "
	dayInput.Placeholder = "12" // TODO : this should auto-default to today's date
	dayInput.CharLimit = 2
	dayInput.SetWidth(2)

	return RecordModel{
		DateYearInput:  yearInput,
		DateMonthInput: monthInput,
		DateDayInput:   dayInput,
	}
}

func (m Model) EnterRecordInput() (Model, tea.Cmd) {
	m.inputType = InputTypeRecord
	m.recordInput.DateYearInput.Focus()
	return m, nil
}

func (m Model) UpdateRecordInput(msg tea.Msg) (Model, tea.Cmd) {
	var recordCmd tea.Cmd
	m.recordInput, recordCmd = m.recordInput.Update(msg)
	return m, recordCmd
}

func (m RecordModel) Update(msg tea.Msg) (RecordModel, tea.Cmd) {

	// TODO : switch focus on tab / shift tab / enter
	// TODO : or even... escape to normal mode and vim keys
	// TODO : to support normal mode, the input prompt should change as well so we know what is hovered (different from simply what is in input mode)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			if m.DateYearInput.Focused() {
				if m.DateYearInput.Value() == "" {
					m.DateYearInput.SetValue(m.DateYearInput.Placeholder)
				}
				m.DateYearInput.Blur()
				m.DateMonthInput.Focus()
				break
			}

			if m.DateMonthInput.Focused() {
				if m.DateMonthInput.Value() == "" {
					m.DateMonthInput.SetValue(m.DateMonthInput.Placeholder)
				}
				m.DateMonthInput.Blur()
				m.DateDayInput.Focus()
				break
			}

			if m.DateDayInput.Focused() {
				if m.DateDayInput.Value() == "" {
					m.DateDayInput.SetValue(m.DateDayInput.Placeholder)
				}
				m.DateDayInput.Blur()
				break
			}

		case "shift+tab":
			if m.DateDayInput.Focused() {
				m.DateDayInput.Blur()
				m.DateMonthInput.Focus()
				break
			}

			if m.DateMonthInput.Focused() {
				m.DateMonthInput.Blur()
				m.DateYearInput.Focus()
				break
			}

			if m.DateYearInput.Focused() {
				m.DateYearInput.Blur()
				m.DateDayInput.Focus()
				break
			}
		}

	}

	var yearCmd tea.Cmd
	m.DateYearInput, yearCmd = m.DateYearInput.Update(msg)

	var monthCmd tea.Cmd
	m.DateMonthInput, monthCmd = m.DateMonthInput.Update(msg)

	var dayCmd tea.Cmd
	m.DateDayInput, dayCmd = m.DateDayInput.Update(msg)

	return m, tea.Batch(yearCmd, monthCmd, dayCmd)
}

func (m RecordModel) View() string {
	var sb strings.Builder
	sb.WriteString(m.DateYearInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateMonthInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateDayInput.View())
	return sb.String()
}
