package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
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

type QueryModel struct {
	State         QueryState
	svc           *service.Service
	err           error

	DateFrom      textinput.Model
	DateTo        textinput.Model
	Currency      CurrencyModel
	Tags          TagsModel
	Fuzzy         textinput.Model
	ActiveField   FilterField
	dateToManual  bool

	savedList    FilteredListModel
	savedQueries []SavedQueryItem

	saveNameInput textinput.Model

	Results       []service.QueryResult
	CursorIndex   int
	pageSize      int

	SelectedID    int64
	ErrorMsg      string
	ResultsOrigin QueryState
	DeleteTarget  SavedQueryItem
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
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		Currency:      NewCurrencyModel(svc),
		Tags:          NewTagsModel(svc),
		Fuzzy:         fuzzy,
		ActiveField:   FilterDateFrom,
		savedList:     NewFilteredListModel("saved queries:"),
		saveNameInput: saveNameInput,
		pageSize:      10,
	}
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
		m.ErrorMsg = fmt.Sprintf("Query error: %v", err)
		return
	}
	m.Results = results
	m.CursorIndex = 0
}

func (m *QueryModel) setError(msg string) {
	m.ErrorMsg = msg
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

func (m QueryModel) loadSavedQueries() (QueryModel, tea.Cmd) {
	saved, err := m.svc.ListSavedQueries(context.Background())
	if err != nil {
		m.ErrorMsg = fmt.Sprintf("Failed to load saved queries: %v", err)
		m.State = QueryStateMenu
		return m, nil
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
	m.State = QueryStateSavedList
	return m, textinput.Blink
}

func (m QueryModel) updateSavedList(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.savedList.SelectedItem(); ok {
				m.executeSavedQuery(item.ID)
			}
			return m, nil
		case "d":
			if item, ok := m.savedList.SelectedItem(); ok {
				for _, sq := range m.savedQueries {
					if sq.ID == item.ID {
						m.DeleteTarget = sq
						m.State = QueryStateDeleteConfirm
						return m, nil
					}
				}
			}
			return m, nil
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
		m.ErrorMsg = "Saved query not found"
		return
	}

	m.DateFrom.SetValue(target.DateFrom)
	m.DateTo.SetValue(target.DateTo)
	m.Fuzzy.SetValue(target.FuzzyText)
	m.dateToManual = true

	if target.CurrencyID != nil {
		m.Currency.LoadCurrencies(context.Background())
		for i := range m.Currency.AllCurrencies {
			if m.Currency.AllCurrencies[i].ID == *target.CurrencyID {
				m.Currency.Selected = &m.Currency.AllCurrencies[i]
				break
			}
		}
	}

	if len(target.TagIDs) > 0 {
		m.Tags.LoadTags(context.Background())
		m.Tags.SelectedTags = make([]TagItem, 0)
		for _, tid := range target.TagIDs {
			for _, t := range m.Tags.AllTags {
				if t.ID == tid {
					m.Tags.SelectedTags = append(m.Tags.SelectedTags, t)
					break
				}
			}
		}
	}

	params := m.currentFilterParams()
	results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
	if err != nil {
		m.ErrorMsg = fmt.Sprintf("Query error: %v", err)
		return
	}

	m.Results = results
	m.CursorIndex = 0
	m.ResultsOrigin = QueryStateSavedList
	m.State = QueryStateResults
}

func (m QueryModel) updateDeleteConfirm(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			if err := m.svc.DeleteSavedQuery(context.Background(), m.DeleteTarget.ID); err != nil {
				m.ErrorMsg = fmt.Sprintf("Failed to delete: %v", err)
			}
			var cmd tea.Cmd
			m, cmd = m.loadSavedQueries()
			return m, cmd
		case "n", "N", "esc":
			m.State = QueryStateSavedList
			return m, nil
		}
	}
	return m, nil
}

func (m QueryModel) updateSaveName(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.saveNameInput.Value())
			if name == "" {
				return m, nil
			}

			params := m.currentFilterParams()

			_, err := m.svc.CreateSavedQueryWithTags(context.Background(), name, params.dateFrom, params.dateTo, params.currencyID, params.fuzzy, 1, params.tagIDs)
			if err != nil {
				m.ErrorMsg = fmt.Sprintf("Failed to save query: %v", err)
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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
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

		case "enter":
			if m.ActiveField == FilterCurrency || m.ActiveField == FilterTags {
				// let sub-widget handle enter (select item / add tag)
			} else if m.ActiveField == FilterFuzzy {
				m.State = QueryStateConfirm
				return m, nil
			} else {
				if isFilterDateField(m.ActiveField) {
					m.fillCurrentDateDefault()
				}
				if m.ActiveField == FilterDateFrom {
					m.autoShiftDateTo()
				}
				next := filterFieldOrder[m.ActiveField+1]
				m.setActiveFilterField(next)
				return m, nil
			}

		case "shift+tab":
			if m.ActiveField > 0 {
				prev := m.ActiveField - 1
				m.setActiveFilterField(prev)
			}
			return m, nil

		case "esc":
			if m.ActiveField == FilterCurrency && m.Currency.Mode == CurrencyModeInsert {
				// let sub-widget handle (toggle mode)
			} else if m.ActiveField == FilterTags && m.Tags.Mode == TagModeInsert {
				// let sub-widget handle (toggle mode)
			} else {
				m.State = QueryStateMenu
				return m, nil
			}
		}
	}

	var fromCmd, toCmd, fuzzyCmd, currencyCmd, tagsCmd tea.Cmd
	oldFromVal := m.DateFrom.Value()
	m.DateFrom, fromCmd = m.DateFrom.Update(msg)
	if oldFromVal != m.DateFrom.Value() {
		m.dateToManual = false
	}

	oldToVal := m.DateTo.Value()
	m.DateTo, toCmd = m.DateTo.Update(msg)
	if oldToVal != m.DateTo.Value() {
		m.dateToManual = true
	}

	switch m.ActiveField {
	case FilterCurrency:
		m.Currency, currencyCmd = m.Currency.Update(msg)
		if m.Currency.ShouldAdvance {
			m.Currency.ShouldAdvance = false
			m.setActiveFilterField(FilterTags)
			return m, tea.Batch(fromCmd, toCmd, currencyCmd, tagsCmd, fuzzyCmd)
		}
	case FilterTags:
		m.Tags, tagsCmd = m.Tags.Update(msg)
	case FilterFuzzy:
		m.Fuzzy, fuzzyCmd = m.Fuzzy.Update(msg)
	}

	return m, tea.Batch(fromCmd, toCmd, currencyCmd, tagsCmd, fuzzyCmd)
}

func (m *QueryModel) autoShiftDateTo() {
	if m.dateToManual {
		return
	}

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

func (m *QueryModel) fillCurrentDateDefault() {
	switch m.ActiveField {
	case FilterDateFrom:
		if m.DateFrom.Value() == "" {
			m.DateFrom.SetValue(m.DateFrom.Placeholder)
		}
	case FilterDateTo:
		if m.DateTo.Value() == "" {
			m.DateTo.SetValue(m.DateTo.Placeholder)
		}
	}
}

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

func (m *QueryModel) FocusActiveField() {
	m.setActiveFilterField(m.ActiveField)
}

func (m QueryModel) updateConfirm(msg tea.Msg) (QueryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			df := m.DateFrom.Value()
			dt := m.DateTo.Value()
			if !isValidDate(df) {
				m.ErrorMsg = "Invalid Date From format (use YYYY-MM-DD)"
				m.State = QueryStateFilterForm
				return m, nil
			}
			if !isValidDate(dt) {
				m.ErrorMsg = "Invalid Date To format (use YYYY-MM-DD)"
				m.State = QueryStateFilterForm
				return m, nil
			}

			params := m.currentFilterParams()
			results, err := m.svc.QueryRecords(context.Background(), params.dateFrom, params.dateTo, params.currencyID, params.tagIDs, params.fuzzy)
			if err != nil {
				m.ErrorMsg = fmt.Sprintf("Query error: %v", err)
				return m, nil
			}
			m.Results = results
			m.CursorIndex = 0
			m.ResultsOrigin = QueryStateFilterForm
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
	if m.ErrorMsg != "" {
		switch msg.(type) {
		case tea.KeyPressMsg:
			m.ErrorMsg = ""
			if m.ResultsOrigin == QueryStateSavedList {
				m.State = QueryStateSavedList
			} else {
				m.State = QueryStateFilterForm
			}
			return m, nil
		}
	}

	if len(m.Results) == 0 {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				if m.ResultsOrigin == QueryStateSavedList {
					m.State = QueryStateSavedList
				} else {
					m.State = QueryStateFilterForm
				}
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.CursorIndex < len(m.Results)-1 {
				m.CursorIndex++
			}
			return m, nil
		case "k", "up":
			if m.CursorIndex > 0 {
				m.CursorIndex--
			}
			return m, nil
		case "n":
			if len(m.Results) > 0 {
				m.CursorIndex += m.pageSize
				if m.CursorIndex >= len(m.Results) {
					m.CursorIndex = len(m.Results) - 1
				}
			}
			return m, nil
		case "p":
			if len(m.Results) > 0 {
				m.CursorIndex -= m.pageSize
				if m.CursorIndex < 0 {
					m.CursorIndex = 0
				}
			}
			return m, nil
		case "g":
			m.CursorIndex = 0
			return m, nil
		case "G":
			if len(m.Results) > 0 {
				m.CursorIndex = len(m.Results) - 1
			}
			return m, nil
		case "enter":
			if len(m.Results) > 0 {
				m.SelectedID = m.Results[m.CursorIndex].ID
			}
			return m, nil
		case "s":
			if len(m.Results) > 0 {
				m.saveNameInput.SetValue("")
				m.State = QueryStateSaveName
				m.saveNameInput.Focus()
			}
			return m, nil
		case "esc":
			if m.ResultsOrigin == QueryStateSavedList {
				m.State = QueryStateSavedList
			} else {
				m.State = QueryStateFilterForm
				m.FocusActiveField()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *QueryModel) View() string {
	switch m.State {
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
	sb.WriteString("tab: next field  |  shift+tab: previous  |  esc: back\n")

	return sb.String()
}

func (m *QueryModel) viewConfirm() string {
	var sb strings.Builder

	sb.WriteString("Query Parameters:\n")
	sb.WriteString("─────────────────────\n")
	sb.WriteString(fmt.Sprintf("  1) Date: %s \u2192 %s\n", m.DateFrom.Value(), m.DateTo.Value()))

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

	if m.ErrorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n\n", m.ErrorMsg))
	}

	if len(m.savedList.Items) == 0 && len(m.savedQueries) == 0 {
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
	sb.WriteString(fmt.Sprintf("  Date: %s \u2192 %s\n", m.DateFrom.Value(), m.DateTo.Value()))
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

	sb.WriteString("\n")
	sb.WriteString("Enter: save query  |  Esc: cancel\n")

	return sb.String()
}

func (m *QueryModel) viewDeleteConfirm() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Delete saved query \"%s\"? (y/n)\n", m.DeleteTarget.Name))
	return sb.String()
}

func (m *QueryModel) viewResults() string {
	var sb strings.Builder

	if m.ErrorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n", m.ErrorMsg))
		sb.WriteString("Press any key to go back.\n")
		return sb.String()
	}

	if len(m.Results) == 0 {
		sb.WriteString("No records match the current filters.\n\n")
		sb.WriteString("Press esc to go back to filter form.\n")
		return sb.String()
	}

	totalPages := (len(m.Results) + m.pageSize - 1) / m.pageSize
	currentPage := m.CursorIndex / m.pageSize
	sb.WriteString(fmt.Sprintf("Page %d/%d  (%d records)\n\n", currentPage+1, totalPages, len(m.Results)))

	pageStart := currentPage * m.pageSize
	pageEnd := pageStart + m.pageSize
	if pageEnd > len(m.Results) {
		pageEnd = len(m.Results)
	}

	for i := pageStart; i < pageEnd; i++ {
		r := m.Results[i]
		cursor := " "
		if i == m.CursorIndex {
			cursor = ">"
		}

		sb.WriteString(fmt.Sprintf("%s %d) %s  %s  %s  %s",
			cursor, i+1, r.Date, service.FormatCents(r.AmountCents), r.CurrencyCode, Truncate(r.Notes, 30)))

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
