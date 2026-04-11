package main

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func NewUnstyledList(items []string) list.Model {

	options := make([]list.Item, 0, len(items))

	for _, item := range items {
		options = append(options, ListItem(item))
	}

	const listWidth = 20 // arbitrary

	// title
	// status bar
	// pagination newline
	// pagination dot
	// help = 1 + expanded buffer 3
	listHeight := len(options) + 3 + 1 + 3

	l := list.New(options, ListItemDelegate{}, listWidth, listHeight)
	l.Title = "Commands:"
	l.Styles = list.Styles{}  // reset styling (see list.DefaultStyles)
	l.SetShowStatusBar(false) // shows item count
	// l.SetShowPagination(false) // we will make sure all is shown anyways
	// l.SetShowHelp(false)

	// show all by default
	l.Help.ShowAll = true

	return l
}

type ListItem string

// list.Item INTERFACE
func (i ListItem) FilterValue() string { return string(i) }

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

	str := fmt.Sprintf("%s %d) %s", cursor, index+1, item)
	fmt.Fprint(w, str)
}

func (m Model) UpdateListInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			if m.userListInput.FilterState() != list.Filtering {
				return m, tea.Quit
			}

		case "enter":
			if m.userListInput.FilterState() == list.Filtering {
				break
			}

			item := m.userListInput.Items()[m.userListInput.GlobalIndex()].(ListItem)

			itemRender := m.userTextInput.Prompt + string(item) + "\n"
			m.history = append(m.history, itemRender)

			// TODO : change to cmd...?
			m.userListInput.ResetSelected()
			m.userListInput.ResetFilter()

			// TODO : for now, swap between
			m.userInputType = InputTypeText
			m.userTextInput.Focus() // TODO : this should be an OnEnter thing
		}
	}

	var cmd tea.Cmd
	m.userListInput, cmd = m.userListInput.Update(msg)
	return m, cmd
}
