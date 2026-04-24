package main

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	database, err := initDB("vimance.db")
	if err != nil {
		fmt.Printf("error initializing database: %v\n", err)
		return
	}
	defer database.Close()

	m := NewModel(database)
	p := tea.NewProgram(m)

	_, err = p.Run()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Println("\n[exiting gracefully...]")
}

func initDB(dataSourceName string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(database, "db/migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

type InputType string

const InputTypeNone InputType = "none"

type Model struct {
	database    *sql.DB
	service     *service.Service
	history     []string
	inputChain  []string
	inputType   InputType
	textInput   textinput.Model
	listInput   list.Model
	recordInput RecordModel
	tagsInput   TagsModel
}

func NewModel(database *sql.DB) Model {
	header := "vimance\n"
	history := []string{header}

	textInput := textinput.New()
	listInput := NewUnstyledList()
	recordInput := NewRecordModel()

	m := Model{
		database:    database,
		service:     service.New(database),
		history:     history,
		textInput:   textInput,
		listInput:   listInput,
		recordInput: recordInput,
	}

	m, _ = m.EnterListInput()
	return m
}

func (m Model) Init() tea.Cmd {
	if m.inputType == InputTypeText {
		return textinput.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.inputType {
	case InputTypeText:
		return m.UpdateTextInput(msg)
	case InputTypeList:
		return m.UpdateListInput(msg)
	case InputTypeRecord:
		return m.UpdateRecordInput(msg)
	}

	return m, nil
}

func (m Model) View() tea.View {
	var sb strings.Builder

	for _, s := range m.history {
		sb.WriteString(s)
	}

	switch m.inputType {
	case InputTypeText:
		sb.WriteString(m.textInput.View())
	case InputTypeList:
		sb.WriteString(m.listInput.View())
	case InputTypeRecord:
		sb.WriteString(m.recordInput.View())
	}

	return tea.NewView(sb.String())
}
