package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type ActiveField int

const (
	FieldDateYear ActiveField = iota
	FieldDateMonth
	FieldDateDay
	FieldTags
	FieldCurrency
	FieldAmount
	FieldNotes
)

type RecordModel struct {
	State          RecordState
	ActiveField    ActiveField
	DateYearInput  textinput.Model
	DateMonthInput textinput.Model
	DateDayInput   textinput.Model

	TagsInput     TagsModel
	CurrencyInput CurrencyModel
	AmountInput   textinput.Model
	NotesInput    textinput.Model

	ConfirmModel ConfirmModel
	SuccessModel SuccessModel
	service      *service.Service
}

func NewRecordModel(svc *service.Service) RecordModel {

	yearInput := textinput.New()
	yearInput.Prompt = "Year: "
	yearInput.Placeholder = time.Now().Format("2006")
	yearInput.CharLimit = 4
	yearInput.SetWidth(4)

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

	tagsInput := NewTagsModel(svc)
	currencyInput := NewCurrencyModel(svc)

	amountInput := textinput.New()
	amountInput.Prompt = "Amount: "
	amountInput.Placeholder = "0.00"

	notesInput := textinput.New()
	notesInput.Prompt = "Notes: "

	m := RecordModel{
		State:         RecordStateEditing,
		ActiveField:   FieldDateYear,
		DateYearInput: yearInput,
		DateMonthInput: monthInput,
		DateDayInput:   dayInput,
		TagsInput:      tagsInput,
		CurrencyInput:  currencyInput,
		AmountInput:    amountInput,
		NotesInput:     notesInput,
		ConfirmModel:   NewConfirmModel(),
		SuccessModel:   NewSuccessModel(),
		service:        svc,
	}

	m.focusActiveField()
	return m
}

func (m *RecordModel) focusActiveField() {
	m.DateYearInput.Blur()
	m.DateMonthInput.Blur()
	m.DateDayInput.Blur()
	m.TagsInput.SearchInput.Blur()
	m.CurrencyInput.SearchInput.Blur()
	m.AmountInput.Blur()
	m.NotesInput.Blur()

	switch m.ActiveField {
	case FieldDateYear:
		m.DateYearInput.Focus()
	case FieldDateMonth:
		m.DateMonthInput.Focus()
	case FieldDateDay:
		m.DateDayInput.Focus()
	case FieldTags:
		m.TagsInput.Mode = TagModeInsert
		m.TagsInput.SearchInput.Focus()
	case FieldCurrency:
		m.CurrencyInput.Mode = CurrencyModeInsert
		m.CurrencyInput.SearchInput.Focus()
	case FieldAmount:
		m.AmountInput.Focus()
	case FieldNotes:
		m.NotesInput.Focus()
	}
}

func (m *RecordModel) setActiveField(field ActiveField) {
	m.ActiveField = field
	m.focusActiveField()
}

func (m Model) EnterRecordInput() (Model, tea.Cmd) {
	m.inputType = InputTypeRecord
	m.recordInput.setActiveField(FieldDateYear)
	return m, nil
}

func (m Model) UpdateRecordInput(msg tea.Msg) (Model, tea.Cmd) {
	var recordCmd tea.Cmd
	m.recordInput, recordCmd = m.recordInput.Update(msg)
	return m, recordCmd
}

func (m RecordModel) Update(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch m.State {
	case RecordStateEditing:
		return m.updateEditing(msg)
	case RecordStateConfirm:
		return m.updateConfirm(msg)
	case RecordStateSuccess:
		return m.updateSuccess(msg)
	}
	return m, nil
}

func (m RecordModel) updateEditing(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			switch m.ActiveField {
			case FieldDateYear:
				if m.DateYearInput.Value() == "" {
					m.DateYearInput.SetValue(m.DateYearInput.Placeholder)
				}
				m.setActiveField(FieldDateMonth)
				return m, nil
			case FieldDateMonth:
				if m.DateMonthInput.Value() == "" {
					m.DateMonthInput.SetValue(m.DateMonthInput.Placeholder)
				}
				m.setActiveField(FieldDateDay)
				return m, nil
			case FieldDateDay:
				if m.DateDayInput.Value() == "" {
					m.DateDayInput.SetValue(m.DateDayInput.Placeholder)
				}
				m.setActiveField(FieldTags)
				return m, nil
			case FieldTags:
				m.setActiveField(FieldCurrency)
				return m, nil
			case FieldCurrency:
				m.setActiveField(FieldAmount)
				return m, nil
			case FieldAmount:
				m.setActiveField(FieldNotes)
				return m, nil
			}
			return m, nil

		case "shift+tab":
			switch m.ActiveField {
			case FieldNotes:
				m.setActiveField(FieldAmount)
				return m, nil
			case FieldAmount:
				m.setActiveField(FieldCurrency)
				return m, nil
			case FieldCurrency:
				m.setActiveField(FieldTags)
				return m, nil
			case FieldTags:
				m.setActiveField(FieldDateDay)
				return m, nil
			case FieldDateDay:
				m.setActiveField(FieldDateMonth)
				return m, nil
			case FieldDateMonth:
				m.setActiveField(FieldDateYear)
				return m, nil
			}
			return m, nil

		case "enter":
			switch m.ActiveField {
			case FieldDateYear:
				if m.DateYearInput.Value() == "" {
					m.DateYearInput.SetValue(m.DateYearInput.Placeholder)
				}
				m.setActiveField(FieldDateMonth)
				return m, nil
			case FieldDateMonth:
				if m.DateMonthInput.Value() == "" {
					m.DateMonthInput.SetValue(m.DateMonthInput.Placeholder)
				}
				m.setActiveField(FieldDateDay)
				return m, nil
			case FieldDateDay:
				if m.DateDayInput.Value() == "" {
					m.DateDayInput.SetValue(m.DateDayInput.Placeholder)
				}
				m.setActiveField(FieldTags)
				return m, nil
			case FieldAmount:
				m.setActiveField(FieldNotes)
				return m, nil
			case FieldNotes:
				m.State = RecordStateConfirm
				m.ConfirmModel.Errors = m.Validate()
				m.ConfirmModel.Warnings = m.GetWarnings()
				return m, nil
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

	var currencyCmd tea.Cmd
	m.CurrencyInput, currencyCmd = m.CurrencyInput.Update(msg)

	var amountCmd tea.Cmd
	m.AmountInput, amountCmd = m.AmountInput.Update(msg)

	var notesCmd tea.Cmd
	m.NotesInput, notesCmd = m.NotesInput.Update(msg)

	if m.CurrencyInput.ShouldAdvance {
		m.CurrencyInput.ShouldAdvance = false
		m.setActiveField(FieldAmount)
		return m, tea.Batch(yearCmd, monthCmd, dayCmd, tagsCmd, currencyCmd, amountCmd, notesCmd)
	}

	return m, tea.Batch(yearCmd, monthCmd, dayCmd, tagsCmd, currencyCmd, amountCmd, notesCmd)
}

func (m RecordModel) updateConfirm(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.State = RecordStateEditing
			m.setActiveField(FieldDateYear)
			return m, nil
		case "2":
			m.State = RecordStateEditing
			m.setActiveField(FieldTags)
			return m, nil
		case "3":
			m.State = RecordStateEditing
			m.setActiveField(FieldCurrency)
			return m, nil
		case "4":
			m.State = RecordStateEditing
			m.setActiveField(FieldAmount)
			return m, nil
		case "5":
			m.State = RecordStateEditing
			m.setActiveField(FieldNotes)
			return m, nil
		case "esc":
			m.State = RecordStateEditing
			m.setActiveField(FieldNotes)
			return m, nil
		case "enter":
			if !m.ConfirmModel.Errors.HasErrors() {
				err := m.CreateRecord(context.Background(), 1)
				if err != nil {
					m.ConfirmModel.Errors = append(m.ConfirmModel.Errors, ValidationError{
						Field:   "record",
						Message: err.Error(),
					})
					return m, nil
				}
				m.State = RecordStateSuccess
				return m, nil
			}
		}
	}

	return m, nil
}

func (m RecordModel) updateSuccess(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return NewRecordModel(m.service), nil
		}
	}

	return m, nil
}

func (m RecordModel) CreateRecord(ctx context.Context, userID int64) error {
	date := formatDate(m.DateYearInput.Value(), m.DateMonthInput.Value(), m.DateDayInput.Value())

	amountCents, err := parseAmountToCents(m.AmountInput.Value())
	if err != nil {
		return err
	}

	currency := m.CurrencyInput.Selected
	if currency == nil {
		return fmt.Errorf("currency is required")
	}

	if currency.IsNew {
		createdCurrency, err := m.service.CreateCurrency(ctx, currency.Code)
		if err != nil {
			return err
		}
		currency.ID = createdCurrency.ID
	}

	tagIDs := make([]int64, 0, len(m.TagsInput.SelectedTags))
	for _, tag := range m.TagsInput.SelectedTags {
		if tag.IsNew {
			createdTag, err := m.service.CreateTag(ctx, tag.Name, "", "", userID)
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, createdTag.ID)
		} else {
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	_, err = m.service.CreateRecordWithTags(ctx, date, amountCents, currency.ID, m.NotesInput.Value(), userID, tagIDs)
	return err
}

func (m RecordModel) View() string {
	switch m.State {
	case RecordStateEditing:
		return m.viewEditing()
	case RecordStateConfirm:
		return m.ConfirmModel.View(m)
	case RecordStateSuccess:
		return m.SuccessModel.View()
	}
	return ""
}

func (m RecordModel) viewEditing() string {
	var sb strings.Builder
	sb.WriteString(m.DateYearInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateMonthInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateDayInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.TagsInput.View())
	sb.WriteString(m.CurrencyInput.View())
	sb.WriteString(m.AmountInput.View())
	sb.WriteString("\n")
	sb.WriteString(m.NotesInput.View())
	return sb.String()
}
