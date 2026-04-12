package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const InputTypeList InputType = "list"

func NewUnstyledList() list.Model {
	const listWidth = 20 // arbitrary

	l := list.New(make([]list.Item, 0), ListItemDelegate{}, listWidth, 0)
	l.Styles = list.Styles{} // reset styling (see list.DefaultStyles)

	l.Title = "commands:"
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(1, 0) // (y, x)

	l.SetShowStatusBar(false) // shows item count

	l.FilterInput.Prompt = "type command: "

	// l.SetShowPagination(false) // we will make sure all is shown anyways
	l.SetShowHelp(false) // TODO : we can rewrite this list in the future anyways
	l.Help.ShowAll = true
	l.KeyMap = CustomKeyMap()

	return l
}

func (m Model) EnterListInput() (Model, tea.Cmd) {
	m.inputType = InputTypeList

	m.listInput.ResetSelected()
	m.listInput.ResetFilter()

	items := []list.Item{
		NewListItem("create", "create a new record", "c", "new", "n"),
		NewListItem("query", "query existing records", "q", "list", "ls", "l"),
		NewListItem("test", "test", "t", "e"),
	}

	cmd := m.listInput.SetItems(items)

	// title bar 3
	// status bar
	// pagination newline
	// pagination dot
	// help = 1 + expanded buffer 3
	listHeight := len(items) + 3 + 2 + 1 + 3
	m.listInput.SetHeight(listHeight)

	// enter filtering immediately
	m.listInput.SetFilterText("")
	m.listInput.SetFilterState(list.Filtering)

	return m, cmd
}

func (m Model) UpdateListInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	isFiltering := m.listInput.FilterState() == list.Filtering

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			if isFiltering {
				// this is still needed for the case of empty filter text
				m.listInput.SetFilterState(list.FilterApplied)

				// TODO : not sure why this doesnt work, probably would be better to create my own list component rather than fighting with the defaults
				m.listInput.Help.ShowAll = true
			} else {
				return m, tea.Quit
			}

		case "up":
			if isFiltering {
				m.listInput.CursorUp()
			}

		case "down":
			if isFiltering {
				m.listInput.CursorDown()
			}

		case "enter":
			// always submit even if in filtering state

			visibleItems := m.listInput.VisibleItems()
			visibleIndex := m.listInput.Index()
			if len(visibleItems) > 0 {
				item := visibleItems[visibleIndex].(ListItem)

				itemRender := m.textInput.Prompt + string(item.title) + "\n"
				m.history = append(m.history, itemRender)

				m.inputChain = append(m.inputChain, item.title)

				// TODO : not sure if there is a more elegant way to do this
				if item.title == "create" {
					return m.EnterRecordInput()
				}

				// re-enter for sub commands
				return m.EnterListInput()
			}
		}
	}

	var cmd tea.Cmd
	m.listInput, cmd = m.listInput.Update(msg)
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

func CustomKeyMap() list.KeyMap {
	return list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k", "shift+tab"),
			key.WithHelp("↑/shift+tab/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j", "tab"),
			key.WithHelp("↓/tab/j", "down"),
		),

		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown"),
			key.WithHelp("→/l/pgdn", "next page"),
		),

		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "go to end"),
		),

		Filter: key.NewBinding(
			key.WithKeys("/", "ctrl+f", "i", "a"),
			key.WithHelp("ctrl+f/i/a", "filter"),
		),

		ClearFilter: key.Binding{},

		// Filtering.
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys(""),
			key.WithHelp("↑/↓", "up/down"), // this is for visualization purposes only
		),

		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("enter", "select"), // this is for visualization purposes only
			// TODO : this help should be shown always, as well as esc
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
