package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type RecordState string

const (
	RecordStateEditing   RecordState = "editing"
	RecordStateConfirm   RecordState = "confirm"
	RecordStateSuccess   RecordState = "success"
	RecordStateFatal     RecordState = "fatal"
)

type ConfirmModel struct {
	Errors   ValidationErrors
	Warnings ValidationErrors
}

func NewConfirmModel() ConfirmModel {
	return ConfirmModel{
		Errors:   make(ValidationErrors, 0),
		Warnings: make(ValidationErrors, 0),
	}
}

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	return m, nil
}

func (m ConfirmModel) View(record RecordModel) string {
	var sb strings.Builder

	sb.WriteString("Review Record:\n")
	sb.WriteString("─────────────────────\n")

	sb.WriteString(fmt.Sprintf("1) Date: %s", formatDate(record.DateYearInput.Value(), record.DateMonthInput.Value(), record.DateDayInput.Value())))
	if errMsg := m.Errors.Get("date"); errMsg != "" {
		sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
	}
	sb.WriteString("\n")

	sb.WriteString("2) Currency: ")
	if record.CurrencyInput.Selected != nil {
		sb.WriteString(record.CurrencyInput.Selected.Code)
		if record.CurrencyInput.Selected.IsNew {
			sb.WriteString("*")
		}
	} else {
		sb.WriteString("(empty)")
		if errMsg := m.Errors.Get("currency"); errMsg != "" {
			sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("3) Tags: ")
	if len(record.TagsInput.SelectedTags) == 0 {
		sb.WriteString("(none)")
		if warnMsg := m.Warnings.Get("tags"); warnMsg != "" {
			sb.WriteString(fmt.Sprintf("  ⚠ %s", warnMsg))
		}
	} else {
		for i, t := range record.TagsInput.SelectedTags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(t.Name)
			if t.IsNew {
				sb.WriteString("*")
			}
		}
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("4) Amount: %s", strings.TrimSpace(record.AmountInput.Value())))
	if errMsg := m.Errors.Get("amount"); errMsg != "" {
		sb.WriteString(fmt.Sprintf("  ← %s", errMsg))
	}
	sb.WriteString("\n")

	sb.WriteString("5) Links: ")
	if len(record.LinksInput.SelectedParents) == 0 {
		sb.WriteString("(none)")
	} else {
		for i, p := range record.LinksInput.SelectedParents {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("[%s]", truncate(p.Notes, 30)))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("6) Notes: %s\n", strings.TrimSpace(record.NotesInput.Value())))

	sb.WriteString("\n")

	hasNew := false
	for _, t := range record.TagsInput.SelectedTags {
		if t.IsNew {
			hasNew = true
			break
		}
	}
	if record.CurrencyInput.Selected != nil && record.CurrencyInput.Selected.IsNew {
		hasNew = true
	}

	if hasNew {
		sb.WriteString("* = new (will be created)\n\n")
	}

	if m.Errors.HasErrors() {
		sb.WriteString("Press number to edit, esc to go back\n")
	} else {
		sb.WriteString("Press number to edit, enter to confirm, esc to go back\n")
	}

	return sb.String()
}

type SuccessModel struct{}

func NewSuccessModel() SuccessModel {
	return SuccessModel{}
}

func (m SuccessModel) Update(msg tea.Msg) (SuccessModel, tea.Cmd) {
	return m, nil
}

func (m SuccessModel) View() string {
	var sb strings.Builder

	sb.WriteString("✓ Record created successfully\n\n")
	sb.WriteString("Press enter to add another record, esc to return to menu\n")

	return sb.String()
}
