package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func NewUnstyledList(items []list.Item) list.Model {
	const listWidth = 20 // arbitrary

	// title bar 3
	// status bar
	// pagination newline
	// pagination dot
	// help = 1 + expanded buffer 3
	listHeight := len(items) + 3 + 2 + 1 + 3

	l := list.New(items, ListItemDelegate{}, listWidth, listHeight)
	l.Styles = list.Styles{} // reset styling (see list.DefaultStyles)

	l.Title = "commands:"
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(1, 0) // (y, x)

	l.SetShowStatusBar(false) // shows item count
	// l.SetShowPagination(false) // we will make sure all is shown anyways
	// l.SetShowHelp(false)

	// show all by default
	l.Help.ShowAll = true
	// TODO : should... customize this...?

	l.FilterInput.Prompt = "type command: "

	return l
}

func (m Model) EnterListInput() Model {
	m.userInputType = InputTypeList

	m.userListInput.ResetSelected()

	// enter filtering immediately
	m.userListInput.SetFilterText("")
	m.userListInput.SetFilterState(list.Filtering)

	return m
}

func (m Model) UpdateListInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			if m.userListInput.FilterState() != list.Filtering {
				return m, tea.Quit
			}

		// TODO : keybindings (like j and k handling!)
		case "i", "a":
			if m.userListInput.FilterState() != list.Filtering {
				m.userListInput.SetFilterText("")
				m.userListInput.SetFilterState(list.Filtering)
				return m, nil
			}

		case "enter":
			// always submit even if in filtering state

			item := m.userListInput.Items()[m.userListInput.GlobalIndex()].(ListItem)

			itemRender := m.userTextInput.Prompt + string(item.title) + "\n"
			m.history = append(m.history, itemRender)

			// TODO : for now, swap between
			return m.EnterTextInput(), nil
		}
	}

	var cmd tea.Cmd
	m.userListInput, cmd = m.userListInput.Update(msg)
	return m, cmd
}

type ListItem struct {
	title string
	desc  string
	alias []string
}

func NewListItem(title, desc string, alias ...string) ListItem {
	return ListItem{title, desc, alias}
}

// list.Item INTERFACE
func (i ListItem) FilterValue() string { return string(i.title + " " + strings.Join(i.alias, " ")) }

type ListItemDelegate struct{}

// list.ItemDelegate INTERFACE
func (d ListItemDelegate) Height() int                             { return 1 }
func (d ListItemDelegate) Spacing() int                            { return 0 }
func (d ListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ListItem)
	if !ok {
		return
	}

	var cursor string
	if index == m.Index() {
		cursor = ">"
	} else {
		cursor = " "
	}

	str := fmt.Sprintf("%s %d) %s\n", cursor, index+1, item.title)
	str += fmt.Sprintf("%s    %s\n", cursor, item.desc)
	str += fmt.Sprintf("%s    alias: [%s]\n", cursor, strings.Join(item.alias, ", "))
	fmt.Fprint(w, str)
}
