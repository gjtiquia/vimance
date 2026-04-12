package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type TagsModel struct {
	ExistingTags []string
	SearchInput  textinput.Model
	AllTags      []string
}

func NewTagsModel() TagsModel {

	tags := make([]string, 0)

	// TODO : get from db
	// TODO : support some level of description too and show that in the list
	// TODO : would be nice if tags can be pinned!
	allTags := []string{
		"food",
		"transport",
		"fun",
		"work",
		"shopping",
	}

	textInput := textinput.New()
	textInput.Prompt = "Type: "

	return TagsModel{
		ExistingTags: tags,
		SearchInput:  textInput,
		AllTags:      allTags,
	}
}

func (m TagsModel) Update(msg tea.Msg) (TagsModel, tea.Cmd) {

	// TODO : other msg handling

	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	return m, cmd
}

func (m TagsModel) View() string {
	var sb strings.Builder

	sb.WriteString("Tags: " + strings.Join(m.ExistingTags, ", ") + "\n")
	sb.WriteString(m.SearchInput.View())
	sb.WriteString("\n\n")

	// filtered
	sb.WriteString(m.FilteredTagsView())

	return sb.String()
}

func (m TagsModel) GetFilteredTags() []string {
	var filtered []string

	input := strings.TrimSpace(m.SearchInput.Value())
	if input == "" {
		return m.AllTags
	}

	for _, tag := range m.AllTags {
		if strings.Contains(tag, input) {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

func (m TagsModel) FilteredTagsView() string {
	filteredTags := m.GetFilteredTags()

	var sb strings.Builder

	for i, tag := range filteredTags {
		sb.WriteString(fmt.Sprintf("%d) %s\n", i+1, tag))
	}

	return sb.String()
}
