package main

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const InputTypeRecord InputType = "record"

type RecordModel struct {
	DateYearInput  textinput.Model
	DateMonthInput textinput.Model
	DateDayInput   textinput.Model

	TagsInput TagsModel

	// TODO
	// - tags
	//   - here i would already need to implement my own list! might as well haha, but not a generic one yet
	//   - can consider... "tag templates" where one can define a set of tags that are commonly used together
	// - currency
	//   - likely this will re-use the tags list, but is constrained to one only
	//   - but with the amount of custom stuff in the tags, likely... no need to reuse yet
	// - amount
	// - notes
}

// TODO : maybe can add "category" support in the future, but its independent of the records, its more of a "collection of tags", but i think "queries" can handle that tho, or perhaps one can even just add a "tag" that is used as a category, cuz we can technically support belonging in more than one category

func NewRecordModel() RecordModel {

	yearInput := textinput.New()
	yearInput.Prompt = "Year: "
	yearInput.Placeholder = time.Now().Format("2006")
	yearInput.CharLimit = 4
	yearInput.SetWidth(4) // required or else placeholder gets truncated to width(0) + 1 = 1 char

	monthInput := textinput.New()
	monthInput.Prompt = "Month: "
	monthInput.Placeholder = time.Now().Format("01")
	monthInput.CharLimit = 2
	monthInput.SetWidth(2)

	dayInput := textinput.New()
	dayInput.Prompt = "Day: "
	dayInput.Placeholder = time.Now().Format("02")
	dayInput.CharLimit = 2
	dayInput.SetWidth(2)

	tagsInput := NewTagsModel()

	return RecordModel{
		DateYearInput:  yearInput,
		DateMonthInput: monthInput,
		DateDayInput:   dayInput,
		TagsInput:      tagsInput,
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
				m.TagsInput.SearchInput.Focus()
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

			if m.TagsInput.SearchInput.Focused() {
				m.TagsInput.SearchInput.Blur()
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

	var tagsCmd tea.Cmd
	m.TagsInput, tagsCmd = m.TagsInput.Update(msg)

	return m, tea.Batch(yearCmd, monthCmd, dayCmd, tagsCmd)
}

func (m RecordModel) View() string {
	var sb strings.Builder
	sb.WriteString(m.DateYearInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateMonthInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateDayInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.TagsInput.View())
	return sb.String()
}
