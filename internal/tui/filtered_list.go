package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type FilterListItem struct {
	ID    int64
	Title string
	Desc  string
}

type FilteredListModel struct {
	Title       string
	Items       []FilterListItem
	FilterInput textinput.Model
	Cursor      int
}

func NewFilteredListModel(title string) FilteredListModel {
	fi := textinput.New()
	fi.Prompt = "type to filter: "
	fi.Focus()

	return FilteredListModel{
		Title:       title,
		FilterInput: fi,
	}
}

func (m *FilteredListModel) SetItems(items []FilterListItem) {
	m.Items = items
	m.Cursor = 0
	m.FilterInput.SetValue("")
}

func (m *FilteredListModel) filteredItems() []FilterListItem {
	filter := strings.ToLower(strings.TrimSpace(m.FilterInput.Value()))
	if filter == "" {
		return m.Items
	}
	var result []FilterListItem
	for _, item := range m.Items {
		if strings.Contains(strings.ToLower(item.Title), filter) {
			result = append(result, item)
		}
	}
	return result
}

func (m *FilteredListModel) CursorUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

func (m *FilteredListModel) CursorDown() {
	items := m.filteredItems()
	if m.Cursor < len(items)-1 {
		m.Cursor++
	}
}

func (m *FilteredListModel) SelectedItem() (FilterListItem, bool) {
	items := m.filteredItems()
	if len(items) == 0 {
		return FilterListItem{}, false
	}
	if m.Cursor >= len(items) {
		m.Cursor = len(items) - 1
	}
	return items[m.Cursor], true
}

func (m FilteredListModel) Update(msg tea.Msg) (FilteredListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k", "shift+tab":
			if m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil
		case "down", "j", "tab":
			items := m.filteredItems()
			if m.Cursor < len(items)-1 {
				m.Cursor++
			}
			return m, nil
		default:
			oldVal := m.FilterInput.Value()
			var cmd tea.Cmd
			m.FilterInput, cmd = m.FilterInput.Update(msg)
			if m.FilterInput.Value() != oldVal {
				items := m.filteredItems()
				if m.Cursor >= len(items) {
					if len(items) > 0 {
						m.Cursor = len(items) - 1
					} else {
						m.Cursor = 0
					}
				}
			}
			return m, cmd
		}
	}
	return m, nil
}

func (m FilteredListModel) View() string {
	var sb strings.Builder

	if m.Title != "" {
		sb.WriteString(m.Title)
		sb.WriteString("\n")
	}

	sb.WriteString(m.FilterInput.View())
	sb.WriteString("\n")

	items := m.filteredItems()
	if len(items) == 0 {
		if len(m.Items) > 0 {
			sb.WriteString("(no matches)\n")
		}
		return sb.String()
	}

	cursor := m.Cursor
	if cursor >= len(items) {
		cursor = len(items) - 1
	}

	for i, item := range items {
		c := " "
		if i == cursor {
			c = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %d) %s\n", c, i+1, item.Title))
		if item.Desc != "" {
			sb.WriteString(fmt.Sprintf("%s    %s\n", c, item.Desc))
		}
	}

	return sb.String()
}
