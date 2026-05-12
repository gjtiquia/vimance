package tui

import (
	"fmt"
	"strings"
)

type LinkedRecord struct {
	ID          int64
	Date        string
	AmountCents int64
	CurrencyID  int64
	Notes       string
	TagNames    []string
}

type LinksModel struct {
	SelectedParents []LinkedRecord
}

func NewLinksModel() LinksModel {
	return LinksModel{
		SelectedParents: make([]LinkedRecord, 0),
	}
}

func (m *LinksModel) AddParent(record LinkedRecord) {
	for _, p := range m.SelectedParents {
		if p.ID == record.ID {
			return
		}
	}
	m.SelectedParents = append(m.SelectedParents, record)
}

func (m *LinksModel) RemoveLastParent() {
	if len(m.SelectedParents) > 0 {
		m.SelectedParents = m.SelectedParents[:len(m.SelectedParents)-1]
	}
}

func (m LinksModel) View() string {
	var sb strings.Builder

	sb.WriteString("Links: ")
	if len(m.SelectedParents) == 0 {
		sb.WriteString("(none)")
	} else {
		for i, p := range m.SelectedParents {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("[%s]", Truncate(p.Notes, 20)))
		}
	}
	sb.WriteString("\n")

	return sb.String()
}
