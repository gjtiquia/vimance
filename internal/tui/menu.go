package tui

import (
	"fmt"
	"strings"
)

type MenuItem struct {
	Title string
	Desc  string
}

type MenuModel struct {
	Title  string
	Items  []MenuItem
	Cursor int
}

func NewMenuModel(title string, items []MenuItem) MenuModel {
	return MenuModel{
		Title:  title,
		Items:  items,
		Cursor: 0,
	}
}

func (m *MenuModel) CursorUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

func (m *MenuModel) CursorDown() {
	if m.Cursor < len(m.Items)-1 {
		m.Cursor++
	}
}

func (m *MenuModel) SelectedItem() (MenuItem, bool) {
	if len(m.Items) == 0 {
		return MenuItem{}, false
	}
	return m.Items[m.Cursor], true
}

func (m *MenuModel) SelectByIndex(n int) bool {
	if n >= 0 && n < len(m.Items) {
		m.Cursor = n
		return true
	}
	return false
}

func (m *MenuModel) Reset() {
	m.Cursor = 0
}

func (m MenuModel) View() string {
	var sb strings.Builder

	sb.WriteString(m.Title)
	sb.WriteString("\n")

	for i, item := range m.Items {
		cursor := " "
		if i == m.Cursor {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %d) %s\n", cursor, i+1, item.Title))
		sb.WriteString(fmt.Sprintf("%s    %s\n", cursor, item.Desc))
	}

	return sb.String()
}
