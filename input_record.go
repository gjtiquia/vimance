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
}

func NewRecordModel() RecordModel {

	yearInput := textinput.New()
	yearInput.Prompt = "Year: "
	yearInput.Placeholder = "2026" // TODO : this should auto-default to today's date

	monthInput := textinput.New()
	monthInput.Prompt = "Month: "
	monthInput.Placeholder = "04" // TODO : this should auto-default to today's date

	dayInput := textinput.New()
	dayInput.Prompt = "Day: "
	dayInput.Placeholder = "12" // TODO : this should auto-default to today's date

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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			// TODO : 
			return m, nil
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
