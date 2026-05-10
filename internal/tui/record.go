package tui

import (
	"context"
	"fmt"
	"os"
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
	FieldCurrency
	FieldTags
	FieldAmount
	FieldLinks
	FieldNotes
)

var fieldOrder = []ActiveField{
	FieldDateYear, FieldDateMonth, FieldDateDay,
	FieldCurrency, FieldTags, FieldAmount,
	FieldLinks, FieldNotes,
}

func (m RecordModel) nextField() ActiveField {
	for i, f := range fieldOrder {
		if f == m.ActiveField && i < len(fieldOrder)-1 {
			return fieldOrder[i+1]
		}
	}
	return m.ActiveField
}

func (m RecordModel) prevField() ActiveField {
	for i, f := range fieldOrder {
		if f == m.ActiveField && i > 0 {
			return fieldOrder[i-1]
		}
	}
	return m.ActiveField
}

func isDateField(f ActiveField) bool {
	return f == FieldDateYear || f == FieldDateMonth || f == FieldDateDay
}

func (m *RecordModel) fillCurrentDateDefault() {
	switch m.ActiveField {
	case FieldDateYear:
		if m.DateYearInput.Value() == "" {
			m.DateYearInput.SetValue(m.DateYearInput.Placeholder)
		}
	case FieldDateMonth:
		if m.DateMonthInput.Value() == "" {
			m.DateMonthInput.SetValue(m.DateMonthInput.Placeholder)
		}
	case FieldDateDay:
		if m.DateDayInput.Value() == "" {
			m.DateDayInput.SetValue(m.DateDayInput.Placeholder)
		}
	}
}

func (m RecordModel) isInInsertMode() bool {
	switch m.ActiveField {
	case FieldCurrency:
		return m.CurrencyInput.Mode == CurrencyModeInsert
	case FieldTags:
		return m.TagsInput.Mode == TagModeInsert
	case FieldLinks:
		return m.LinksInput.Mode == LinkModeInsert
	}
	return false
}

type RecordModel struct {
	State          RecordState
	ActiveField    ActiveField
	DateYearInput  textinput.Model
	DateMonthInput textinput.Model
	DateDayInput   textinput.Model

	TagsInput     TagsModel
	CurrencyInput CurrencyModel
	AmountInput   textinput.Model
	LinksInput    LinksModel
	NotesInput    textinput.Model

	ConfirmModel   ConfirmModel
	SuccessModel   SuccessModel
	FatalErr       error
	InlineErrors   map[string]string
	InlineWarnings map[string]string
	service        *service.Service
	Origin         RecordOrigin
	EditRecordID   int64
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

	linksInput := NewLinksModel(svc)

	notesInput := textinput.New()
	notesInput.Prompt = "Notes: "

	m := RecordModel{
		State:          RecordStateEditing,
		ActiveField:    FieldDateYear,
		DateYearInput:  yearInput,
		DateMonthInput: monthInput,
		DateDayInput:   dayInput,
		TagsInput:      tagsInput,
		CurrencyInput:  currencyInput,
		AmountInput:    amountInput,
		LinksInput:     linksInput,
		NotesInput:     notesInput,
		ConfirmModel:   NewConfirmModel(),
		SuccessModel:   NewSuccessModel(),
		InlineErrors:   make(map[string]string),
		InlineWarnings: make(map[string]string),
		service:        svc,
	}

	m.focusActiveField()
	return m
}

func NewEditRecordModel(svc *service.Service, full *service.RecordFull, origin RecordOrigin) RecordModel {
	m := NewRecordModel(svc)
	m.Origin = origin
	m.EditRecordID = full.Record.ID

	// pre-fill date
	parts := strings.Split(full.Record.Date, "-")
	if len(parts) == 3 {
		m.DateYearInput.SetValue(parts[0])
		m.DateMonthInput.SetValue(parts[1])
		m.DateDayInput.SetValue(parts[2])
	}

	// pre-fill amount
	m.AmountInput.SetValue(service.FormatCents(full.Record.AmountCents))

	// pre-fill notes
	m.NotesInput.SetValue(full.Record.Notes)

	// pre-fill currency
	m.CurrencyInput.LoadCurrencies(context.Background())
	for i := range m.CurrencyInput.AllCurrencies {
		if m.CurrencyInput.AllCurrencies[i].ID == full.Record.CurrencyID {
			m.CurrencyInput.Selected = &m.CurrencyInput.AllCurrencies[i]
			break
		}
	}

	// pre-fill tags
	m.TagsInput.LoadTags(context.Background())
	for _, t := range full.Tags {
		m.TagsInput.addTag(t.Name)
	}

	// pre-fill links
	if len(full.Parents) > 0 {
		year := m.DateYearInput.Value()
		month := m.DateMonthInput.Value()
		m.LinksInput.SetDateRange(year, month)
		if m.CurrencyInput.Selected != nil {
			m.LinksInput.SetCurrencyID(m.CurrencyInput.Selected.ID)
		}
		m.LinksInput.SelectedParents = make([]LinkedRecord, len(full.Parents))
		for i, p := range full.Parents {
			m.LinksInput.SelectedParents[i] = LinkedRecord{
				ID:          p.ID,
				Date:        p.Date,
				AmountCents: p.AmountCents,
				CurrencyID:  p.CurrencyID,
				Notes:       p.Notes,
				TagNames:    p.TagNames,
			}
		}
	}

	return m
}

func (m *RecordModel) focusActiveField() {
	m.DateYearInput.Blur()
	m.DateMonthInput.Blur()
	m.DateDayInput.Blur()
	m.TagsInput.SearchInput.Blur()
	m.CurrencyInput.SearchInput.Blur()
	m.AmountInput.Blur()
	m.LinksInput.SearchInput.Blur()
	m.NotesInput.Blur()

	switch m.ActiveField {
	case FieldDateYear:
		m.DateYearInput.Focus()
	case FieldDateMonth:
		m.DateMonthInput.Focus()
	case FieldDateDay:
		m.DateDayInput.Focus()
	case FieldTags:
		m.TagsInput.SearchInput.Focus()
	case FieldCurrency:
		m.CurrencyInput.SearchInput.Focus()
	case FieldAmount:
		m.AmountInput.Focus()
	case FieldLinks:
		m.LinksInput.SearchInput.Focus()
	case FieldNotes:
		m.NotesInput.Focus()
	}
}

func (m *RecordModel) validateCurrentField() {
	switch m.ActiveField {
	case FieldDateYear, FieldDateMonth, FieldDateDay:
		if err := ValidateDate(m.DateYearInput.Value(), m.DateMonthInput.Value(), m.DateDayInput.Value()); err != nil {
			m.InlineErrors["date"] = err.Error()
		} else {
			delete(m.InlineErrors, "date")
		}
	case FieldCurrency:
		if m.CurrencyInput.Selected == nil {
			m.InlineErrors["currency"] = "currency is required"
		} else {
			delete(m.InlineErrors, "currency")
		}
	case FieldAmount:
		if err := ValidateAmount(m.AmountInput.Value()); err != nil {
			m.InlineErrors["amount"] = err.Error()
		} else {
			delete(m.InlineErrors, "amount")
		}
	case FieldTags:
		if len(m.TagsInput.SelectedTags) == 0 {
			m.InlineWarnings["tags"] = "no tags selected"
		} else {
			delete(m.InlineWarnings, "tags")
		}
	}
}

func (m *RecordModel) setActiveField(field ActiveField) {
	if m.ActiveField == field {
		return
	}

	m.validateCurrentField()
	m.CurrencyInput.ShouldAdvance = false

	switch field {
	case FieldTags:
		m.TagsInput.Mode = TagModeInsert
		m.TagsInput.LoadTags(context.Background())
	case FieldCurrency:
		m.CurrencyInput.Mode = CurrencyModeInsert
		m.CurrencyInput.LoadCurrencies(context.Background())
	case FieldLinks:
		m.LinksInput.Mode = LinkModeInsert
		m.LinksInput.SetDateRange(m.DateYearInput.Value(), m.DateMonthInput.Value())
		if m.CurrencyInput.Selected != nil {
			m.LinksInput.SetCurrencyID(m.CurrencyInput.Selected.ID)
		}
		m.LinksInput.LoadCandidates(context.Background(), m.EditRecordID)
	}

	m.ActiveField = field
	m.focusActiveField()
}

func (m Model) EnterRecordInput() (Model, tea.Cmd) {
	m.InputType = InputTypeRecord
	m.RecordInput.Origin = RecordOriginCreate
	m.RecordInput.setActiveField(FieldDateYear)
	return m, nil
}

func (m Model) UpdateRecordInput(msg tea.Msg) (Model, tea.Cmd) {
	var recordCmd tea.Cmd
	m.RecordInput, recordCmd = m.RecordInput.Update(msg)
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
	case RecordStateFatal:
		return m.updateFatal(msg)
	}
	return m, nil
}

func (m RecordModel) updateEditing(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			if isDateField(m.ActiveField) {
				m.fillCurrentDateDefault()
			}
			next := m.nextField()
			if next != m.ActiveField {
				m.setActiveField(next)
			}
			return m, nil

		case "shift+tab":
			prev := m.prevField()
			if prev != m.ActiveField {
				m.setActiveField(prev)
			}
			return m, nil

		case "enter":
			switch m.ActiveField {
			case FieldDateYear, FieldDateMonth, FieldDateDay:
				m.fillCurrentDateDefault()
				m.setActiveField(m.nextField())
				return m, nil
			case FieldCurrency:
				m.CurrencyInput, _ = m.CurrencyInput.Update(msg)
				if m.CurrencyInput.ShouldAdvance {
					m.CurrencyInput.ShouldAdvance = false
					m.setActiveField(FieldTags)
				}
				return m, nil
			case FieldTags:
				m.TagsInput, _ = m.TagsInput.Update(msg)
				return m, nil
			case FieldAmount:
				m.setActiveField(m.nextField())
				return m, nil
			case FieldLinks:
				if len(m.LinksInput.FilteredCandidates) == 0 {
					m.setActiveField(m.nextField())
					return m, nil
				}
				m.LinksInput, _ = m.LinksInput.Update(msg)
				return m, nil
			case FieldNotes:
				m.State = RecordStateConfirm
				m.ConfirmModel.Errors = m.Validate()
				m.ConfirmModel.Warnings = m.GetWarnings()
				return m, nil
			}
		}
	}

	var yearCmd, monthCmd, dayCmd tea.Cmd
	m.DateYearInput, yearCmd = m.DateYearInput.Update(msg)
	m.DateMonthInput, monthCmd = m.DateMonthInput.Update(msg)
	m.DateDayInput, dayCmd = m.DateDayInput.Update(msg)

	var amountCmd, notesCmd tea.Cmd
	m.AmountInput, amountCmd = m.AmountInput.Update(msg)
	m.NotesInput, notesCmd = m.NotesInput.Update(msg)

	var tagsCmd, currencyCmd, linksCmd tea.Cmd
	switch m.ActiveField {
	case FieldTags:
		m.TagsInput, tagsCmd = m.TagsInput.Update(msg)
	case FieldCurrency:
		m.CurrencyInput, currencyCmd = m.CurrencyInput.Update(msg)
	case FieldLinks:
		m.LinksInput, linksCmd = m.LinksInput.Update(msg)
	}

	if m.CurrencyInput.ShouldAdvance {
		m.CurrencyInput.ShouldAdvance = false
		m.setActiveField(FieldTags)
		return m, tea.Batch(yearCmd, monthCmd, dayCmd, tagsCmd, currencyCmd, amountCmd, linksCmd, notesCmd)
	}

	return m, tea.Batch(yearCmd, monthCmd, dayCmd, tagsCmd, currencyCmd, amountCmd, linksCmd, notesCmd)
}

func (m RecordModel) updateConfirm(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "1":
			m.State = RecordStateEditing
			m.setActiveField(FieldDateYear)
			return m, nil
		case "2":
			m.State = RecordStateEditing
			m.setActiveField(FieldCurrency)
			return m, nil
		case "3":
			m.State = RecordStateEditing
			m.setActiveField(FieldTags)
			return m, nil
		case "4":
			m.State = RecordStateEditing
			m.setActiveField(FieldAmount)
			return m, nil
		case "5":
			m.State = RecordStateEditing
			m.setActiveField(FieldLinks)
			return m, nil
		case "6":
			m.State = RecordStateEditing
			m.setActiveField(FieldNotes)
			return m, nil
		case "esc":
			m.State = RecordStateEditing
			m.setActiveField(FieldNotes)
			return m, nil
		case "enter":
			if !m.ConfirmModel.Errors.HasErrors() {
			err := m.SaveRecord(context.Background(), 1)
			if err != nil {
				m.FatalErr = err
				fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
				m.State = RecordStateFatal
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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.Origin == RecordOriginCreate {
				return NewRecordModel(m.service), nil
			}
		case "esc":
			// handled by app.go for origin-aware routing
			return m, nil
		}
	}

	return m, nil
}

func (m RecordModel) updateFatal(msg tea.Msg) (RecordModel, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m RecordModel) SaveRecord(ctx context.Context, userID int64) error {
	if m.Origin == RecordOriginQuery {
		return m.updateRecord(ctx, userID)
	}
	return m.CreateRecord(ctx, userID)
}

func (m RecordModel) updateRecord(ctx context.Context, userID int64) error {
	date := FormatDate(m.DateYearInput.Value(), m.DateMonthInput.Value(), m.DateDayInput.Value())

	amountCents, err := ParseAmountToCents(m.AmountInput.Value())
	if err != nil {
		return err
	}

	currency := m.CurrencyInput.Selected
	if currency == nil {
		return fmt.Errorf("currency is required")
	}

	if currency.IsNew {
		createdCurrency, _, err := m.service.GetOrCreateCurrency(ctx, currency.Code)
		if err != nil {
			return err
		}
		currency.ID = createdCurrency.ID
	}

	tagIDs := make([]int64, 0, len(m.TagsInput.SelectedTags))
	for _, tag := range m.TagsInput.SelectedTags {
		if tag.IsNew {
			createdTag, _, err := m.service.GetOrCreateTag(ctx, tag.Name, userID)
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, createdTag.ID)
		} else {
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	parentIDs := make([]int64, 0, len(m.LinksInput.SelectedParents))
	for _, p := range m.LinksInput.SelectedParents {
		parentIDs = append(parentIDs, p.ID)
	}

	_, err = m.service.UpdateRecordWithTagsAndLinks(ctx, m.EditRecordID, date, amountCents, currency.ID, m.NotesInput.Value(), userID, tagIDs, parentIDs)
	return err
}

func (m RecordModel) CreateRecord(ctx context.Context, userID int64) error {
	date := FormatDate(m.DateYearInput.Value(), m.DateMonthInput.Value(), m.DateDayInput.Value())

	amountCents, err := ParseAmountToCents(m.AmountInput.Value())
	if err != nil {
		return err
	}

	currency := m.CurrencyInput.Selected
	if currency == nil {
		return fmt.Errorf("currency is required")
	}

	if currency.IsNew {
		createdCurrency, _, err := m.service.GetOrCreateCurrency(ctx, currency.Code)
		if err != nil {
			return err
		}
		currency.ID = createdCurrency.ID
	}

	tagIDs := make([]int64, 0, len(m.TagsInput.SelectedTags))
	for _, tag := range m.TagsInput.SelectedTags {
		if tag.IsNew {
			createdTag, _, err := m.service.GetOrCreateTag(ctx, tag.Name, userID)
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, createdTag.ID)
		} else {
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	parentIDs := make([]int64, 0, len(m.LinksInput.SelectedParents))
	for _, p := range m.LinksInput.SelectedParents {
		parentIDs = append(parentIDs, p.ID)
	}

	_, err = m.service.CreateRecordWithTagsAndLinks(ctx, date, amountCents, currency.ID, m.NotesInput.Value(), userID, tagIDs, parentIDs)
	return err
}

func (m RecordModel) viewFatal() string {
	return fmt.Sprintf("\nFatal error: %v\n\nPress any key to exit.\n", m.FatalErr)
}

func (m RecordModel) View() string {
	switch m.State {
	case RecordStateEditing:
		return m.viewEditing()
	case RecordStateConfirm:
		return m.ConfirmModel.View(m)
	case RecordStateSuccess:
		return m.SuccessModel.View(m.Origin)
	case RecordStateFatal:
		return m.viewFatal()
	}
	return ""
}

func (m RecordModel) viewEditing() string {
	var sb strings.Builder
	for _, f := range fieldOrder {
		switch f {
		case FieldDateYear:
			if isDateField(m.ActiveField) {
				prefixY, prefixM, prefixD := "  ", "  ", "  "
				switch m.ActiveField {
				case FieldDateYear:
					prefixY = "> "
				case FieldDateMonth:
					prefixM = "> "
				case FieldDateDay:
					prefixD = "> "
				}
				sb.WriteString(prefixY)
				sb.WriteString(m.DateYearInput.View())
				sb.WriteString("\n")
				sb.WriteString(prefixM)
				sb.WriteString(m.DateMonthInput.View())
				sb.WriteString("\n")
				sb.WriteString(prefixD)
				sb.WriteString(m.DateDayInput.View())
				sb.WriteString("\n")
			} else {
				sb.WriteString("  Date: ")
				year := m.DateYearInput.Value()
				month := m.DateMonthInput.Value()
				day := m.DateDayInput.Value()
				if year != "" && month != "" && day != "" {
					sb.WriteString(FormatDate(year, month, day))
				} else {
					sb.WriteString("(empty)")
				}
				if errMsg := m.InlineErrors["date"]; errMsg != "" {
					sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
				}
				sb.WriteString("\n")
			}
		case FieldDateMonth, FieldDateDay:
			continue
		case FieldCurrency:
			if m.ActiveField == FieldCurrency {
				sb.WriteString("> ")
				sb.WriteString(m.CurrencyInput.View())
			} else {
				sb.WriteString("  Currency: ")
				if m.CurrencyInput.Selected != nil {
					sb.WriteString(m.CurrencyInput.Selected.Code)
					if m.CurrencyInput.Selected.IsNew {
						sb.WriteString("*")
					}
				} else {
					sb.WriteString("(empty)")
				}
				if errMsg := m.InlineErrors["currency"]; errMsg != "" {
					sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
				}
				sb.WriteString("\n")
			}
		case FieldTags:
			if m.ActiveField == FieldTags {
				sb.WriteString("> ")
				sb.WriteString(m.TagsInput.View())
			} else {
				sb.WriteString("  Tags: ")
				if len(m.TagsInput.SelectedTags) > 0 {
					for i, t := range m.TagsInput.SelectedTags {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(t.Name)
						if t.IsNew {
							sb.WriteString("*")
						}
					}
				} else {
					sb.WriteString("(none)")
				}
				if warnMsg := m.InlineWarnings["tags"]; warnMsg != "" {
					sb.WriteString(fmt.Sprintf("  ⚠ %s", warnMsg))
				}
				sb.WriteString("\n")
			}
		case FieldAmount:
			if m.ActiveField == FieldAmount {
				sb.WriteString("> ")
				sb.WriteString(m.AmountInput.View())
				sb.WriteString("\n")
			} else {
				sb.WriteString("  Amount: ")
				val := strings.TrimSpace(m.AmountInput.Value())
				if val == "" {
					sb.WriteString("(empty)")
				} else {
					sb.WriteString(val)
				}
				if errMsg := m.InlineErrors["amount"]; errMsg != "" {
					sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
				}
				sb.WriteString("\n")
			}
		case FieldLinks:
			if m.ActiveField == FieldLinks {
				sb.WriteString("> ")
				sb.WriteString(m.LinksInput.View())
			} else {
				sb.WriteString("  Links: ")
				if len(m.LinksInput.SelectedParents) > 0 {
					for i, p := range m.LinksInput.SelectedParents {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("[%s]", Truncate(p.Notes, 20)))
					}
				} else {
					sb.WriteString("(none)")
				}
				sb.WriteString("\n")
			}
		case FieldNotes:
			if m.ActiveField == FieldNotes {
				sb.WriteString("> ")
				sb.WriteString(m.NotesInput.View())
				sb.WriteString("\n")
			} else {
				sb.WriteString("  Notes: ")
				val := strings.TrimSpace(m.NotesInput.Value())
				if val == "" {
					sb.WriteString("(empty)")
				} else {
					sb.WriteString(val)
				}
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}
