package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type TargetsState string

const (
	TargetsStateList         TargetsState = "list"
	TargetsStateCreate       TargetsState = "create"
	TargetsStateDeleteConfirm TargetsState = "delete_confirm"
)

type TargetsModel struct {
	State    TargetsState
	svc      *service.Service
	err      error
	targets  []service.TargetWithActual
	cursor   int

	// Create flow
	savedQueries []SavedQueryItem
	savedList    FilteredListModel
	selectedSQ   *SavedQueryItem
	nameInput    textinput.Model
	amountInput  textinput.Model
	createStage  int // 0=pick query, 1=enter name, 2=enter amount
}

func NewTargetsModel(svc *service.Service) TargetsModel {
	nameInput := textinput.New()
	nameInput.Prompt = "Name: "
	nameInput.CharLimit = 64

	amountInput := textinput.New()
	amountInput.Prompt = "Target amount: "
	amountInput.Placeholder = "0.00"

	return TargetsModel{
		State:       TargetsStateList,
		svc:         svc,
		nameInput:   nameInput,
		amountInput: amountInput,
		savedList:   NewFilteredListModel("saved queries:"),
	}
}

func (m *TargetsModel) loadTargets() tea.Cmd {
	targets, err := m.svc.ListTargetsWithActuals(context.Background())
	if err != nil {
		m.err = err
		m.targets = nil
		return nil
	}
	m.targets = targets
	return nil
}

func (m *TargetsModel) loadSavedQueries() tea.Cmd {
	saved, err := m.svc.ListSavedQueries(context.Background())
	if err != nil {
		m.err = err
		return nil
	}

	m.savedQueries = make([]SavedQueryItem, len(saved))
	items := make([]FilterListItem, len(saved))

	for i, sq := range saved {
		sqItem := SavedQueryItem{
			ID:        sq.ID,
			Name:      sq.Name,
			DateFrom:  sq.DateFrom,
			DateTo:    sq.DateTo,
			FuzzyText: sq.FuzzyText,
		}
		if sq.CurrencyID.Valid {
			v := sq.CurrencyID.Int64
			sqItem.CurrencyID = &v
		}

		tags, err := m.svc.GetSavedQueryTags(context.Background(), sq.ID)
		if err == nil {
			sqItem.TagIDs = make([]int64, len(tags))
			for j, t := range tags {
				sqItem.TagIDs[j] = t.ID
			}
		}

		m.savedQueries[i] = sqItem
		items[i] = FilterListItem{
			ID:    sq.ID,
			Title: sq.Name,
			Desc:  fmt.Sprintf("%s \u2192 %s", sq.DateFrom, sq.DateTo),
		}
	}

	m.savedList.SetItems(items)
	return textinput.Blink
}

func (m TargetsModel) Update(msg tea.Msg) (TargetsModel, tea.Cmd) {
	switch m.State {
	case TargetsStateList:
		return m.updateList(msg)
	case TargetsStateCreate:
		return m.updateCreate(msg)
	case TargetsStateDeleteConfirm:
		return m.updateDeleteConfirm(msg)
	}
	return m, nil
}

func (m TargetsModel) updateList(msg tea.Msg) (TargetsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.targets)-1 {
				m.cursor++
			}
			return m, nil
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "a":
			m.State = TargetsStateCreate
			m.createStage = 0
			m.selectedSQ = nil
			cmd := m.loadSavedQueries()
			return m, cmd
		case "d":
			if len(m.targets) > 0 {
				m.State = TargetsStateDeleteConfirm
				return m, nil
			}
			return m, nil
		case "enter":
			// TODO: navigate to query results with target's saved query filters
			return m, nil
		}
	}
	return m, nil
}

func (m TargetsModel) updateCreate(msg tea.Msg) (TargetsModel, tea.Cmd) {
	switch m.createStage {
	case 0:
		// Pick saved query
		var cmd tea.Cmd
		m.savedList, cmd = m.savedList.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "enter":
				if item, ok := m.savedList.SelectedItem(); ok {
					for i := range m.savedQueries {
						if m.savedQueries[i].ID == item.ID {
							sq := m.savedQueries[i]
							m.selectedSQ = &sq
							m.createStage = 1
							m.nameInput.SetValue(sq.Name)
							m.nameInput.Focus()
							return m, textinput.Blink
						}
					}
				}
			case "esc":
				m.State = TargetsStateList
				cmd := m.loadTargets()
				return m, cmd
			}
		}
		return m, cmd
	case 1:
		// Enter name
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.nameInput.Value())
				if name == "" {
					return m, nil
				}
				m.createStage = 2
				m.nameInput.Blur()
				m.amountInput.Focus()
				return m, textinput.Blink
			case "esc":
				m.State = TargetsStateList
				cmd := m.loadTargets()
				return m, cmd
			}
		}
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	case 2:
		// Enter amount
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "enter":
				amountStr := strings.TrimSpace(m.amountInput.Value())
				if amountStr == "" {
					return m, nil
				}
				amountCents, err := ParseAmountToCents(amountStr)
				if err != nil {
					m.err = fmt.Errorf("invalid amount: %v", err)
					return m, nil
				}

				user, err := m.svc.ListActiveUsers(context.Background())
				if err != nil || len(user) == 0 {
					m.err = fmt.Errorf("no user found")
					return m, nil
				}

				_, err = m.svc.CreateTarget(context.Background(), strings.TrimSpace(m.nameInput.Value()), m.selectedSQ.ID, amountCents, user[0].ID)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.nameInput.SetValue("")
				m.amountInput.SetValue("")
				m.amountInput.Blur()
				m.State = TargetsStateList
				m.err = nil
				cmd := m.loadTargets()
				return m, cmd
			case "esc":
				m.State = TargetsStateList
				cmd := m.loadTargets()
				return m, cmd
			}
		}
		var cmd tea.Cmd
		m.amountInput, cmd = m.amountInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m TargetsModel) updateDeleteConfirm(msg tea.Msg) (TargetsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			if len(m.targets) > 0 && m.cursor < len(m.targets) {
				err := m.svc.DeleteTarget(context.Background(), m.targets[m.cursor].Target.ID)
				if err != nil {
					m.err = err
				}
			}
			m.State = TargetsStateList
			cmd := m.loadTargets()
			return m, cmd
		case "n", "N", "esc":
			m.State = TargetsStateList
			return m, nil
		}
	}
	return m, nil
}

func (m TargetsModel) View() string {
	switch m.State {
	case TargetsStateList:
		return m.viewList()
	case TargetsStateCreate:
		return m.viewCreate()
	case TargetsStateDeleteConfirm:
		return m.viewDeleteConfirm()
	}
	return ""
}

func (m TargetsModel) viewList() string {
	var sb strings.Builder

	sb.WriteString("Targets\n")
	sb.WriteString("\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")

	if m.err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	if len(m.targets) == 0 {
		sb.WriteString("No targets yet.\n\n")
		sb.WriteString("Press a to add a target, esc to go back.\n")
		return sb.String()
	}

	maxNameLen := 4
	if len(m.targets) == 0 {
		maxNameLen = 4
	} else {
		for _, t := range m.targets {
			if len(t.Target.Name) > maxNameLen {
				maxNameLen = len(t.Target.Name)
			}
		}
	}

	for i, t := range m.targets {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		actualStr := "???"
		if t.HasData && t.ActualAmount != nil {
			actualStr = service.FormatCents(*t.ActualAmount)
		}
		targetStr := service.FormatCents(t.Target.TargetCents)

		status := " "
		if t.HasData && t.ActualAmount != nil {
			if t.Target.TargetCents < 0 {
				if *t.ActualAmount <= t.Target.TargetCents {
					status = "\u2713"
				} else {
					status = "\u2717"
				}
			} else {
				if *t.ActualAmount >= t.Target.TargetCents {
					status = "\u2713"
				} else {
					status = "\u2717"
				}
			}
		} else {
			status = "\u26A0"
		}

		padFmt := fmt.Sprintf("%%-%ds", maxNameLen)
		sb.WriteString(fmt.Sprintf("%s %s  %s of %s  %s\n",
			cursor,
			fmt.Sprintf(padFmt, t.Target.Name),
			actualStr,
			targetStr,
			status,
		))
	}

	sb.WriteString("\n")
	sb.WriteString("j/k: move | a: add | d: delete | esc: back\n")
	return sb.String()
}

func (m TargetsModel) viewCreate() string {
	var sb strings.Builder

	sb.WriteString("Create Target\n")
	sb.WriteString("\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")

	if m.err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	switch m.createStage {
	case 0:
		sb.WriteString("Select a saved query:\n\n")
		sb.WriteString(m.savedList.View())
		sb.WriteString("\nenter: select | esc: back\n")
	case 1:
		sb.WriteString("Selected query: ")
		if m.selectedSQ != nil {
			sb.WriteString(fmt.Sprintf("%s (%s \u2192 %s)\n\n", m.selectedSQ.Name, m.selectedSQ.DateFrom, m.selectedSQ.DateTo))
		}
		sb.WriteString(m.nameInput.View())
		sb.WriteString("\n\nenter: next | esc: cancel\n")
	case 2:
		sb.WriteString(fmt.Sprintf("Name: %s\n", m.nameInput.Value()))
		sb.WriteString("Selected query: ")
		if m.selectedSQ != nil {
			sb.WriteString(fmt.Sprintf("%s\n", m.selectedSQ.Name))
		}
		sb.WriteString("\n")
		sb.WriteString(m.amountInput.View())
		sb.WriteString("\n\nNegative amounts = expense budget (e.g. -500.00)\n")
		sb.WriteString("Positive amounts = savings/income goal (e.g. 10000.00)\n\n")
		sb.WriteString("enter: create | esc: cancel\n")
	}

	return sb.String()
}

func (m TargetsModel) viewDeleteConfirm() string {
	var sb strings.Builder
	if len(m.targets) > 0 && m.cursor < len(m.targets) {
		sb.WriteString(fmt.Sprintf("Delete target \"%s\"? (y/n)\n", m.targets[m.cursor].Target.Name))
	} else {
		sb.WriteString("No target to delete.\n")
	}
	return sb.String()
}