package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type TagItem struct {
	ID       int64
	Name     string
	IsPinned bool
	IsNew    bool
}

type TagMode string

const (
	TagModeInsert TagMode = "insert"
	TagModeNormal TagMode = "normal"
)

type TagsModel struct {
	SelectedTags []TagItem
	SearchInput  textinput.Model
	AllTags      []TagItem
	CursorIndex  int
	Mode         TagMode
	service      *service.Service
}

func NewTagsModel(svc *service.Service) TagsModel {
	textInput := textinput.New()
	textInput.Prompt = "Type: "

	return TagsModel{
		SelectedTags: make([]TagItem, 0),
		SearchInput:  textInput,
		AllTags:      make([]TagItem, 0),
		CursorIndex:  0,
		Mode:         TagModeInsert,
		service:      svc,
	}
}

func (m *TagsModel) LoadTags(ctx context.Context) error {
	activeTags, err := m.service.ListActiveTags(ctx)
	if err != nil {
		return err
	}

	pinnedTags, err := m.service.ListPinnedTags(ctx)
	if err != nil {
		return err
	}

	pinnedIDs := make(map[int64]bool)
	for _, t := range pinnedTags {
		pinnedIDs[t.ID] = true
	}

	m.AllTags = make([]TagItem, 0)

	for _, t := range pinnedTags {
		m.AllTags = append(m.AllTags, TagItem{
			ID:       t.ID,
			Name:     t.Name,
			IsPinned: true,
			IsNew:    false,
		})
	}

	for _, t := range activeTags {
		if !pinnedIDs[t.ID] {
			m.AllTags = append(m.AllTags, TagItem{
				ID:       t.ID,
				Name:     t.Name,
				IsPinned: false,
				IsNew:    false,
			})
		}
	}

	return nil
}

func (m TagsModel) getFilteredTags() []TagItem {
	input := strings.TrimSpace(m.SearchInput.Value())
	if input == "" {
		return m.AllTags
	}

	var filtered []TagItem
	for _, tag := range m.AllTags {
		if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(input)) {
			filtered = append(filtered, tag)
		}
	}

	for _, selected := range m.SelectedTags {
		if selected.IsNew && strings.EqualFold(selected.Name, input) {
			return filtered
		}
	}

	return filtered
}

func (m *TagsModel) addTag(name string) {
	for _, t := range m.SelectedTags {
		if strings.EqualFold(t.Name, name) {
			return
		}
	}

	for _, t := range m.AllTags {
		if strings.EqualFold(t.Name, name) {
			m.SelectedTags = append(m.SelectedTags, t)
			m.SearchInput.SetValue("")
			return
		}
	}

	m.SelectedTags = append(m.SelectedTags, TagItem{
		ID:       0,
		Name:     name,
		IsPinned: false,
		IsNew:    true,
	})
	m.SearchInput.SetValue("")
}

func (m *TagsModel) removeLastTag() {
	if len(m.SelectedTags) > 0 {
		m.SelectedTags = m.SelectedTags[:len(m.SelectedTags)-1]
	}
}

func (m TagsModel) Update(msg tea.Msg) (TagsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Mode {
		case TagModeInsert:
			switch msg.String() {
			case "esc":
				m.Mode = TagModeNormal
				m.SearchInput.Blur()
				if len(m.getFilteredTags()) > 0 {
					m.CursorIndex = 0
				}
				return m, nil
			case "enter":
				input := strings.TrimSpace(m.SearchInput.Value())
				if input != "" {
					m.addTag(input)
				}
				return m, nil
			case "ctrl+z":
				m.removeLastTag()
				return m, nil
			}
		case TagModeNormal:
			switch msg.String() {
			case "i", "a":
				m.Mode = TagModeInsert
				m.SearchInput.Focus()
				return m, nil
			case "j", "down":
				filtered := m.getFilteredTags()
				if len(filtered) > 0 && m.CursorIndex < len(filtered)-1 {
					m.CursorIndex++
				}
				return m, nil
			case "k", "up":
				if m.CursorIndex > 0 {
					m.CursorIndex--
				}
				return m, nil
			case "enter":
				filtered := m.getFilteredTags()
				if len(filtered) > 0 && m.CursorIndex < len(filtered) {
					m.addTag(filtered[m.CursorIndex].Name)
				}
				return m, nil
			case "ctrl+z":
				m.removeLastTag()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	return m, cmd
}

func (m TagsModel) View() string {
	var sb strings.Builder

	sb.WriteString("Tags: ")
	for i, t := range m.SelectedTags {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.Name)
		if t.IsNew {
			sb.WriteString("*")
		}
	}
	sb.WriteString("\n")

	sb.WriteString(m.SearchInput.View())
	if m.Mode == TagModeNormal {
		sb.WriteString(" [NORMAL]")
	}
	sb.WriteString("\n\n")

	sb.WriteString(m.filteredTagsView())

	return sb.String()
}

func (m TagsModel) filteredTagsView() string {
	filtered := m.getFilteredTags()

	var sb strings.Builder

	for i, tag := range filtered {
		cursor := " "
		if m.Mode == TagModeNormal && i == m.CursorIndex {
			cursor = ">"
		}

		pinned := ""
		if tag.IsPinned {
			pinned = " [pinned]"
		}

		sb.WriteString(fmt.Sprintf("%s %d) %s%s\n", cursor, i+1, tag.Name, pinned))
	}

	return sb.String()
}
