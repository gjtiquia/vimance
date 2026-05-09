package tui

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gjtiquia/vimance/internal/service"
)

type QueryState string

const (
	QueryStateMenu          QueryState = "menu"
	QueryStateFilterForm    QueryState = "filter"
	QueryStateConfirm       QueryState = "confirm"
	QueryStateSavedList     QueryState = "saved"
	QueryStateSaveName      QueryState = "save_name"
	QueryStateDeleteConfirm QueryState = "delete_confirm"
	QueryStateResults       QueryState = "results"
)

type FilterField int

const (
	FilterDateFrom FilterField = iota
	FilterDateTo
	FilterCurrency
	FilterTags
	FilterFuzzy
)

var filterFieldOrder = []FilterField{FilterDateFrom, FilterDateTo, FilterCurrency, FilterTags, FilterFuzzy}

type SavedQueryItem struct {
	ID         int64
	Name       string
	DateFrom   string
	DateTo     string
	CurrencyID *int64
	FuzzyText  string
	TagIDs     []int64
}

func (i SavedQueryItem) FilterValue() string { return i.Name }

type QueryModel struct {
	State         QueryState
	svc           *service.Service
	err           error

	menuInput     list.Model

	DateFrom      textinput.Model
	DateTo        textinput.Model
	Currency      CurrencyModel
	Tags          TagsModel
	Fuzzy         textinput.Model
	ActiveField   FilterField
	dateToManual  bool

	savedList    list.Model
	savedQueries []SavedQueryItem

	deleteTarget SavedQueryItem

	saveNameInput textinput.Model

	results      []service.QueryResult
	cursorIndex  int
	pageSize     int
	needsRefresh bool

	selectedID int64
	errorMsg   string
}

func NewQueryModel(svc *service.Service) QueryModel {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	firstOfMonth := fmt.Sprintf("%04d-%02d-01", currentYear, currentMonth)
	lastOfMonth := fmt.Sprintf("%04d-%02d-%02d", currentYear, currentMonth, daysInMonth(currentYear, currentMonth))

	dateFrom := textinput.New()
	dateFrom.Prompt = "Date From: "
	dateFrom.Placeholder = firstOfMonth
	dateFrom.SetValue(firstOfMonth)
	dateFrom.CharLimit = 10
	dateFrom.SetWidth(10)

	dateTo := textinput.New()
	dateTo.Prompt = "Date To: "
	dateTo.Placeholder = lastOfMonth
	dateTo.SetValue(lastOfMonth)
	dateTo.CharLimit = 10
	dateTo.SetWidth(10)

	fuzzy := textinput.New()
	fuzzy.Prompt = "Fuzzy: "
	fuzzy.Placeholder = "optional"

	saveNameInput := textinput.New()
	saveNameInput.Prompt = "Name: "
	saveNameInput.CharLimit = 64

	return QueryModel{
		State:         QueryStateMenu,
		svc:           svc,
		menuInput:     newQueryMenuList(),
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		Currency:      NewCurrencyModel(svc),
		Tags:          NewTagsModel(svc),
		Fuzzy:         fuzzy,
		ActiveField:   FilterDateFrom,
		savedList:     newSavedQueryList(),
		saveNameInput: saveNameInput,
		pageSize:      10,
	}
}

func newQueryMenuList() list.Model {
	const listWidth = 20

	items := []list.Item{
		NewListItem("new", "create a new query", "n"),
		NewListItem("saved", "use a saved query", "s"),
	}

	l := list.New(items, ListItemDelegate{}, listWidth, 0)
	l.Styles = list.Styles{}
	l.Title = "query:"
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(1, 0)
	l.SetShowStatusBar(false)
	l.FilterInput.Prompt = "type command: "
	l.SetShowHelp(false)
	l.Help.ShowAll = true
	l.KeyMap = CustomKeyMap()

	listHeight := len(items) + 3 + 2 + 1 + 3
	l.SetHeight(listHeight)

	l.SetFilterText("")
	l.SetFilterState(list.Filtering)

	return l
}

func newSavedQueryList() list.Model {
	const listWidth = 20

	l := list.New(make([]list.Item, 0), SavedQueryListItemDelegate{}, listWidth, 0)
	l.Styles = list.Styles{}
	l.Title = "saved queries:"
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(1, 0)
	l.SetShowStatusBar(false)
	l.FilterInput.Prompt = "type to filter: "
	l.SetShowHelp(false)
	l.Help.ShowAll = true
	l.KeyMap = CustomKeyMap()

	return l
}

type SavedQueryListItem struct {
	ID       int64
	Name     string
	DateFrom string
	DateTo   string
}

func (i SavedQueryListItem) FilterValue() string { return i.Name }

type SavedQueryListItemDelegate struct{}

func (d SavedQueryListItemDelegate) Height() int                             { return 1 }
func (d SavedQueryListItemDelegate) Spacing() int                            { return 0 }
func (d SavedQueryListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d SavedQueryListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(SavedQueryListItem)
	if !ok {
		return
	}

	var cursor string
	if index == m.Index() {
		cursor = ">"
	} else {
		cursor = " "
	}

	fmt.Fprintf(w, "%s %s  |  %s → %s\n", cursor, item.Name, item.DateFrom, item.DateTo)
}

func (m *QueryModel) SetPageSize(height int) {
	if height > 8 {
		m.pageSize = height - 4
	} else {
		m.pageSize = 5
	}
}

func (m *QueryModel) RefreshResults() {
	if m.State != QueryStateResults {
		return
	}

	params := m.currentFilterParams()
	results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Query error: %v", err)
		return
	}
	m.results = results
	m.cursorIndex = 0
}

func (m *QueryModel) setError(msg string) {
	m.errorMsg = msg
}

type filterParams struct {
	dateFrom   string
	dateTo     string
	currencyID *int64
	tagIDs     []int64
	fuzzy      string
}

func (m *QueryModel) currentFilterParams() filterParams {
	var currencyID *int64
	if m.Currency.Selected != nil && !m.Currency.Selected.IsNew {
		currencyID = &m.Currency.Selected.ID
	}

	tagIDs := make([]int64, 0, len(m.Tags.SelectedTags))
	for _, t := range m.Tags.SelectedTags {
		if !t.IsNew {
			tagIDs = append(tagIDs, t.ID)
		}
	}

	return filterParams{
		dateFrom:   m.DateFrom.Value(),
		dateTo:     m.DateTo.Value(),
		currencyID: currencyID,
		tagIDs:     tagIDs,
		fuzzy:      strings.TrimSpace(m.Fuzzy.Value()),
	}
}

func (m QueryModel) Update(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch m.State {
	case QueryStateMenu:
		return m.updateMenu(msg)
	case QueryStateFilterForm:
		return m.updateFilterForm(msg)
	case QueryStateConfirm:
		return m.updateConfirm(msg)
	case QueryStateSavedList:
		return m.updateSavedList(msg)
	case QueryStateSaveName:
		return m.updateSaveName(msg)
	case QueryStateDeleteConfirm:
		return m.updateDeleteConfirm(msg)
	case QueryStateResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m QueryModel) updateMenu(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			visibleItems := m.menuInput.VisibleItems()
			visibleIndex := m.menuInput.Index()
			if len(visibleItems) > 0 {
				item := visibleItems[visibleIndex].(ListItem)
				if item.title == "new" {
					m.State = QueryStateFilterForm
					m.ActiveField = FilterDateFrom
					m.focusActiveField()
					return m, nil
				}
				if item.title == "saved" {
					m.loadSavedQueries()
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.menuInput, cmd = m.menuInput.Update(msg)
	return m, cmd
}

func (m *QueryModel) loadSavedQueries() {
	saved, err := m.svc.ListSavedQueries(context.Background())
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to load saved queries: %v", err)
		return
	}

	m.savedQueries = make([]SavedQueryItem, len(saved))
	items := make([]list.Item, len(saved))

	for i, sq := range saved {
		item := SavedQueryItem{
			ID:        sq.ID,
			Name:      sq.Name,
			DateFrom:  sq.DateFrom,
			DateTo:    sq.DateTo,
			FuzzyText: sq.FuzzyText,
		}
		if sq.CurrencyID.Valid {
			v := sq.CurrencyID.Int64
			item.CurrencyID = &v
		}

		tags, err := m.svc.GetSavedQueryTags(context.Background(), sq.ID)
		if err == nil {
			item.TagIDs = make([]int64, len(tags))
			for j, t := range tags {
				item.TagIDs[j] = t.ID
			}
		}

		m.savedQueries[i] = item
		items[i] = SavedQueryListItem{
			ID:       sq.ID,
			Name:     sq.Name,
			DateFrom: sq.DateFrom,
			DateTo:   sq.DateTo,
		}
	}

	m.savedList.SetItems(items)
	listHeight := len(items) + 3 + 2 + 1 + 3
	m.savedList.SetHeight(listHeight)
	m.State = QueryStateSavedList
}

func (m QueryModel) updateSavedList(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			visibleItems := m.savedList.VisibleItems()
			visibleIndex := m.savedList.Index()
			if len(visibleItems) > 0 {
				sqItem := visibleItems[visibleIndex].(SavedQueryListItem)
				m.executeSavedQuery(sqItem.ID)
			}
			return m, nil
		case "d":
			visibleItems := m.savedList.VisibleItems()
			visibleIndex := m.savedList.Index()
			if len(visibleItems) > 0 {
				sqItem := visibleItems[visibleIndex].(SavedQueryListItem)
				for _, sq := range m.savedQueries {
					if sq.ID == sqItem.ID {
						m.deleteTarget = sq
						m.State = QueryStateDeleteConfirm
						return m, nil
					}
				}
			}
		case "esc":
			m.State = QueryStateMenu
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.savedList, cmd = m.savedList.Update(msg)
	return m, cmd
}

func (m *QueryModel) executeSavedQuery(id int64) {
	var target SavedQueryItem
	found := false
	for _, sq := range m.savedQueries {
		if sq.ID == id {
			target = sq
			found = true
			break
		}
	}
	if !found {
		m.errorMsg = "Saved query not found"
		return
	}

	m.DateFrom.SetValue(target.DateFrom)
	m.DateTo.SetValue(target.DateTo)
	m.Fuzzy.SetValue(target.FuzzyText)

	if target.CurrencyID != nil {
		m.Currency.LoadCurrencies(context.Background())
		for _, c := range m.Currency.AllCurrencies {
			if c.ID == *target.CurrencyID {
				m.Currency.Selected = &c
				break
			}
		}
	}

	params := m.currentFilterParams()
	results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Query error: %v", err)
		return
	}

	m.results = results
	m.cursorIndex = 0
	m.State = QueryStateResults
}

func (m QueryModel) updateDeleteConfirm(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if err := m.svc.DeleteSavedQuery(context.Background(), m.deleteTarget.ID); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to delete: %v", err)
			}
			m.loadSavedQueries()
			return m, nil
		case "n", "N", "esc":
			m.State = QueryStateSavedList
			return m, nil
		}
	}
	return m, nil
}

func (m QueryModel) updateSaveName(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.saveNameInput.Value())
			if name == "" {
				return m, nil
			}

			params := m.currentFilterParams()

			_, err := m.svc.CreateSavedQueryWithTags(context.Background(), name, params.dateFrom, params.dateTo, params.currencyID, params.fuzzy, 1, params.tagIDs)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save query: %v", err)
				m.State = QueryStateResults
				return m, nil
			}

			m.saveNameInput.SetValue("")
			m.State = QueryStateResults
			return m, nil
		case "esc":
			m.saveNameInput.SetValue("")
			m.State = QueryStateResults
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.saveNameInput, cmd = m.saveNameInput.Update(msg)
	return m, cmd
}

func (m QueryModel) updateFilterForm(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			if m.ActiveField == FilterFuzzy {
				m.State = QueryStateConfirm
				return m, nil
			}

			if isFilterDateField(m.ActiveField) {
				m.fillCurrentDateDefault()
			}

			if m.ActiveField == FilterDateFrom {
				m.autoShiftDateTo()
			}

			next := filterFieldOrder[m.ActiveField+1]
			m.setActiveFilterField(next)
			return m, nil

		case "shift+tab":
			if m.ActiveField > 0 {
				prev := m.ActiveField - 1
				m.setActiveFilterField(prev)
			}
			return m, nil

		case "esc":
			m.State = QueryStateMenu
			return m, nil
		}
	}

	var fromCmd, toCmd, fuzzyCmd tea.Cmd
	m.DateFrom, fromCmd = m.DateFrom.Update(msg)
	m.DateTo, toCmd = m.DateTo.Update(msg)
	m.Currency, _ = m.Currency.Update(msg)
	m.Tags, _ = m.Tags.Update(msg)
	m.Fuzzy, fuzzyCmd = m.Fuzzy.Update(msg)

	return m, tea.Batch(fromCmd, toCmd, fuzzyCmd)
}

func (m *QueryModel) autoShiftDateTo() {
	fromVal := m.DateFrom.Value()
	if !isValidDate(fromVal) {
		return
	}

	parts := strings.Split(fromVal, "-")
	if len(parts) != 3 {
		return
	}

	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	lastDay := daysInMonth(year, time.Month(month))
	m.DateTo.SetValue(fmt.Sprintf("%04d-%02d-%02d", year, month, lastDay))
}

func isFilterDateField(f FilterField) bool {
	return f == FilterDateFrom || f == FilterDateTo
}

func (m *QueryModel) fillCurrentDateDefault() {}

func isValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (m *QueryModel) setActiveFilterField(field FilterField) {
	m.DateFrom.Blur()
	m.DateTo.Blur()
	m.Currency.SearchInput.Blur()
	m.Tags.SearchInput.Blur()
	m.Fuzzy.Blur()

	m.ActiveField = field

	switch field {
	case FilterDateFrom:
		m.DateFrom.Focus()
	case FilterDateTo:
		m.DateTo.Focus()
	case FilterCurrency:
		m.Currency.Mode = CurrencyModeInsert
		m.Currency.LoadCurrencies(context.Background())
		m.Currency.SearchInput.Focus()
	case FilterTags:
		m.Tags.Mode = TagModeInsert
		m.Tags.LoadTags(context.Background())
		m.Tags.SearchInput.Focus()
	case FilterFuzzy:
		m.Fuzzy.Focus()
	}
}

func (m *QueryModel) focusActiveField() {
	m.setActiveFilterField(m.ActiveField)
}

func (m QueryModel) updateConfirm(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			params := m.currentFilterParams()
			results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Query error: %v", err)
				return m, nil
			}
			m.results = results
			m.cursorIndex = 0
			m.State = QueryStateResults
			return m, nil
		case "esc":
			m.State = QueryStateFilterForm
			return m, nil
		case "1":
			m.setActiveFilterField(FilterDateFrom)
			m.State = QueryStateFilterForm
			return m, nil
		case "2":
			m.setActiveFilterField(FilterCurrency)
			m.State = QueryStateFilterForm
			return m, nil
		case "3":
			m.setActiveFilterField(FilterTags)
			m.State = QueryStateFilterForm
			return m, nil
		case "4":
			m.setActiveFilterField(FilterFuzzy)
			m.State = QueryStateFilterForm
			return m, nil
		}
	}
	return m, nil
}

func (m QueryModel) updateResults(msg tea.Msg) (QueryModel, tea.Cmd) {
	if m.errorMsg != "" {
		switch msg.(type) {
		case tea.KeyMsg:
			m.errorMsg = ""
			m.State = QueryStateFilterForm
			return m, nil
		}
	}

	if m.needsRefresh {
		m.needsRefresh = false
		params := m.currentFilterParams()
		results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Query error: %v", err)
			return m, nil
		}
		m.results = results
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursorIndex < len(m.results)-1 {
				m.cursorIndex++
			}
			return m, nil
		case "k", "up":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}
			return m, nil
		case "n":
			m.cursorIndex += m.pageSize
			if m.cursorIndex >= len(m.results) {
				m.cursorIndex = len(m.results) - 1
			}
			return m, nil
		case "p":
			m.cursorIndex -= m.pageSize
			if m.cursorIndex < 0 {
				m.cursorIndex = 0
			}
			return m, nil
		case "g":
			m.cursorIndex = 0
			return m, nil
		case "G":
			m.cursorIndex = len(m.results) - 1
			return m, nil
		case "enter":
			if len(m.results) > 0 {
				m.selectedID = m.results[m.cursorIndex].ID
			}
			return m, nil
		case "s":
			if len(m.results) > 0 {
				m.saveNameInput.SetValue("")
				m.State = QueryStateSaveName
				m.saveNameInput.Focus()
			}
			return m, nil
		case "esc":
			m.State = QueryStateFilterForm
			m.focusActiveField()
			return m, nil
		}
	}
	return m, nil
}

func (m *QueryModel) View() string {
	switch m.State {
	case QueryStateMenu:
		return m.viewMenu()
	case QueryStateFilterForm:
		return m.viewFilterForm()
	case QueryStateConfirm:
		return m.viewConfirm()
	case QueryStateSavedList:
		return m.viewSavedList()
	case QueryStateSaveName:
		return m.viewSaveName()
	case QueryStateDeleteConfirm:
		return m.viewDeleteConfirm()
	case QueryStateResults:
		return m.viewResults()
	}
	return ""
}

func (m *QueryModel) viewMenu() string {
	return m.menuInput.View()
}

func (m *QueryModel) viewFilterForm() string {
	var sb strings.Builder

	sb.WriteString(m.DateFrom.View())
	sb.WriteString("\n")
	sb.WriteString(m.DateTo.View())
	sb.WriteString("\n")
	sb.WriteString(m.Currency.View())
	sb.WriteString(m.Tags.View())
	sb.WriteString(m.Fuzzy.View())
	sb.WriteString("\n\n")
	sb.WriteString("tab: next field  |  shift+tab: previous  |  enter: confirm  |  esc: back\n")

	return sb.String()
}

func (m *QueryModel) viewConfirm() string {
	var sb strings.Builder

	sb.WriteString("Query Parameters:\n")
	sb.WriteString("─────────────────────\n")
	sb.WriteString(fmt.Sprintf("  1) Date: %s → %s\n", m.DateFrom.Value(), m.DateTo.Value()))

	if m.Currency.Selected != nil {
		sb.WriteString(fmt.Sprintf("  2) Currency: %s", m.Currency.Selected.Code))
		if m.Currency.Selected.IsNew {
			sb.WriteString("*")
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("  2) Currency: (any)\n")
	}

	sb.WriteString("  3) Tags: ")
	if len(m.Tags.SelectedTags) == 0 {
		sb.WriteString("(any)")
	} else {
		for i, t := range m.Tags.SelectedTags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(t.Name)
		}
	}
	sb.WriteString("\n")

	fuzzy := strings.TrimSpace(m.Fuzzy.Value())
	if fuzzy == "" {
		sb.WriteString("  4) Fuzzy: (none)\n")
	} else {
		sb.WriteString(fmt.Sprintf("  4) Fuzzy: %s\n", fuzzy))
	}

	sb.WriteString("\nEnter: execute query\n")
	sb.WriteString("Esc: back to edit\n")
	sb.WriteString("Press number to edit field\n")

	return sb.String()
}

func (m *QueryModel) viewSavedList() string {
	var sb strings.Builder

	visibleItems := m.savedList.VisibleItems()
	if len(visibleItems) == 0 && len(m.savedQueries) == 0 {
		sb.WriteString("No saved queries yet.\n\n")
		sb.WriteString("Press esc to go back.\n")
		return sb.String()
	}

	sb.WriteString(m.savedList.View())
	return sb.String()
}

func (m *QueryModel) viewSaveName() string {
	var sb strings.Builder

	sb.WriteString(m.saveNameInput.View())
	sb.WriteString("\n\n")

	sb.WriteString("Filter:\n")
	sb.WriteString(fmt.Sprintf("  Date: %s → %s\n", m.DateFrom.Value(), m.DateTo.Value()))
	if m.Currency.Selected != nil {
		sb.WriteString(fmt.Sprintf("  Currency: %s\n", m.Currency.Selected.Code))
	} else {
		sb.WriteString("  Currency: (any)\n")
	}
	if len(m.Tags.SelectedTags) > 0 {
		sb.WriteString("  Tags: ")
		for i, t := range m.Tags.SelectedTags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(t.Name)
		}
		sb.WriteString("\n")
	}
	fuzzy := strings.TrimSpace(m.Fuzzy.Value())
	if fuzzy != "" {
		sb.WriteString(fmt.Sprintf("  Fuzzy: %s\n", fuzzy))
	}

	return sb.String()
}

func (m *QueryModel) viewDeleteConfirm() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Delete saved query \"%s\"? (y/n)\n", m.deleteTarget.Name))
	return sb.String()
}

func (m *QueryModel) viewResults() string {
	var sb strings.Builder

	if m.errorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n", m.errorMsg))
		sb.WriteString("Press any key to go back.\n")
		return sb.String()
	}

	if len(m.results) == 0 {
		sb.WriteString("No records match the current filters.\n\n")
		sb.WriteString("Press esc to go back to filter form.\n")
		return sb.String()
	}

	totalPages := (len(m.results) + m.pageSize - 1) / m.pageSize
	currentPage := m.cursorIndex / m.pageSize
	sb.WriteString(fmt.Sprintf("Page %d/%d  (%d records)\n\n", currentPage+1, totalPages, len(m.results)))

	pageStart := currentPage * m.pageSize
	pageEnd := pageStart + m.pageSize
	if pageEnd > len(m.results) {
		pageEnd = len(m.results)
	}

	for i := pageStart; i < pageEnd; i++ {
		r := m.results[i]
		cursor := " "
		if i == m.cursorIndex {
			cursor = ">"
		}

		sb.WriteString(fmt.Sprintf("%s %d) %s  %s  %s  %s",
			cursor, i+1, r.Date, service.FormatCents(r.AmountCents), r.CurrencyCode, truncate(r.Notes, 30)))

		if len(r.TagNames) > 0 {
			tagsToShow := filterOutFilterTags(r.TagNames, m.Tags.SelectedTags)
			if len(tagsToShow) > 0 {
				sb.WriteString(fmt.Sprintf("  [%s]", strings.Join(tagsToShow, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nj/k: move  |  n/p: page  |  g/G: top/bottom  |  enter: edit  |  s: save query  |  esc: back\n")

	return sb.String()
}

func filterOutFilterTags(tagNames []string, filterTags []TagItem) []string {
	filterSet := make(map[string]bool)
	for _, t := range filterTags {
		filterSet[t.Name] = true
	}

	var result []string
	for _, name := range tagNames {
		if !filterSet[name] {
			result = append(result, name)
		}
	}
	return result
}
