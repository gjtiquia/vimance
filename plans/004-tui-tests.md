# TUI Tests Implementation Plan

## Overview

Add tests for the BubbleTea TUI layer (`internal/tui/`) that test state machine logic without coupling to rendering. Leverages ELM architecture where `Update(msg)` is pure-ish and `View()` is completely separate.

Design decisions:
- `package tui_test` (external tests) — black box, tests access model state via exported fields
- Export required fields on `Model` and `QueryModel` instead of accessor methods (internal package, no external consumers to protect)
- Sub-models (Currency, Tags, Links) manually populated with nil service — no DB needed
- RecordModel and QueryModel need real in-memory SQLite DB for field navigation (triggers `LoadCurrencies`/`LoadTags`)
- Never test `View()` output — treat rendering as freely changeable

## Key technical notes

### tea.Cmd comparison

`tea.Cmd` is `func() tea.Msg`. **Function values cannot be compared in Go** — `cmd != tea.Quit` will not compile.

To test for quit:
```go
_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
if cmd == nil {
    t.Fatal("expected non-nil cmd")
}
msg := cmd()
if _, ok := msg.(tea.QuitMsg); !ok {
    t.Errorf("expected tea.QuitMsg, got %T", msg)
}
```

### textinput.Model requires Focus()

`textinput.Model.Update(msg)` returns `(m, nil)` immediately when `m.focus == false`. Tests must ensure the textinput is focused before sending character keypresses.

- `NewRecordModel(svc)` calls `m.focusActiveField()` — FieldDateYear is focused. ✅
- `NewQueryModel(svc)` does NOT call `focusActiveField()` — initial field is NOT focused. Tests that set `m.State = QueryStateFilterForm` directly must manually focus the initial field: `m.DateFrom.Focus()`
- After the first `tab` keypress, `setActiveFilterField()` is called, which focuses the next field. So only the initial field needs manual focus setup.

### nil service safety

`NewCurrencyModel(nil)`, `NewTagsModel(nil)`, `NewLinksModel(nil)` are safe — `Update()` never calls `LoadCurrencies`/`LoadTags`/`LoadCandidates` directly. Those are only called from `setActiveField` (RecordModel) or `setActiveFilterField` (QueryModel), which use a real service.

### Model.Update returns tea.Model

Top-level `Model.Update` returns `(tea.Model, tea.Cmd)`. Tests must type-assert:
```go
result, _ := m.Update(msg)
m = result.(tui.Model)
```

Sub-model `Update` methods return their concrete type — no assertion needed.

## Files

```
internal/tui/
  testing_helper.go       # shared test helpers (package tui_test)
  model_test.go           # top-level Model smoke/integration
  record_model_test.go    # RecordModel state machine
  query_model_test.go     # QueryModel state machine
  currency_model_test.go  # CurrencyModel selection/filter/mode
  tags_model_test.go      # TagsModel add/remove/mode
  links_model_test.go     # LinksModel parent select/deselect
  validation_test.go      # pure functions
```

## Phase 1: Export fields + helpers

### Step 1.1 — Export Model fields

In `internal/tui/app.go`, rename these fields to exported:

| old | new | reason |
|---|---|---|
| `inputType` | `InputType` | top-level state routing |
| `recordInput` | `RecordInput` | inspect RecordModel from tests |
| `queryInput` | `QueryInput` | inspect QueryModel from tests |
| `width` | `Width` | test WindowSizeMsg handling |
| `height` | `Height` | test WindowSizeMsg handling |

Also rename all references throughout the file:
- `m.inputType` → `m.InputType` (in `Init`, `Update`, `View`, `EnterTextInput`, `UpdateTextInput`, `EnterListInput`, `EnterQueryInput`, `UpdateQueryInput`, `routeBackFromRecord`, `EnterRecordInput`, `UpdateRecordInput`)
- `m.recordInput` → `m.RecordInput` (in `NewModel`, `Update`, `View`, `routeBackFromRecord`, `EnterQueryInput`, `UpdateQueryInput`, `EnterRecordInput`, `UpdateRecordInput`)
- `m.queryInput` → `m.QueryInput` (in `NewModel`, `Update`, `View`, `routeBackFromRecord`, `EnterQueryInput`, `UpdateQueryInput`)
- `m.width` → `m.Width` (in `Update`)
- `m.height` → `m.Height` (in `Update`, `EnterQueryInput`)

### Step 1.2 — Export QueryModel fields

In `internal/tui/query.go`, rename these fields:

| old | new | reason |
|---|---|---|
| `cursorIndex` | `CursorIndex` | assert cursor position in results |
| `selectedID` | `SelectedID` | detect record selection |
| `errorMsg` | `ErrorMsg` | assert error state |
| `results` | `Results` | inspect result count/data |
| `resultsOrigin` | `ResultsOrigin` | esc routing from results (saved list vs filter form) |

Rename all references:
- `m.cursorIndex` → `m.CursorIndex` (in `RefreshResults`, `updateResults`, `View`)
- `m.selectedID` → `m.SelectedID` (declaration and usage in `UpdateQueryInput`)
- `m.errorMsg` → `m.ErrorMsg` (in `View`, `loadSavedQueries`, `updateSavedList`, `executeSavedQuery`, `updateDeleteConfirm`, `updateSaveName`, `updateConfirm`, `updateResults`, `setError`)
- `m.results` → `m.Results` (in `RefreshResults`, `executeSavedQuery`, `updateConfirm`, `updateResults`, `View`)
- `m.resultsOrigin` → `m.ResultsOrigin` (in `executeSavedQuery`, `updateConfirm`, `updateResults`)

### Step 1.3 — Export QueryModel.focusActiveField

Rename `focusActiveField` → `FocusActiveField` in `internal/tui/query.go`.

This method focuses/blurs text inputs for the active filter field. Tests that set `State = QueryStateFilterForm` directly must call `m.FocusActiveField()` for textinput keypresses to work.

Rename all internal references:
- `m.focusActiveField()` → `m.FocusActiveField()` (in `EnterQueryInput`, `updateResults`, `setActiveFilterField`)

### Step 1.4 — Export validation functions

In `internal/tui/validation.go`, rename these functions:

| old | new |
|---|---|
| `validateAmount` | `ValidateAmount` |
| `parseAmountToCents` | `ParseAmountToCents` |
| `validateDate` | `ValidateDate` |
| `formatDate` | `FormatDate` |
| `truncate` | `Truncate` |

Also rename all internal references:
- `validateAmount(` → `ValidateAmount(` (in `Validate`, `ParseAmountToCents`)
- `parseAmountToCents(` → `ParseAmountToCents(` (in `CreateRecord`, `updateRecord`)
- `validateDate(` → `ValidateDate(` (in `Validate`)
- `formatDate(` → `FormatDate(` (in `CreateRecord`, `updateRecord`, `viewConfirm`, `viewSaveName`, `viewResults`)
- `truncate(` → `Truncate(` (in `viewResults`, `filteredListView`, `ConfirmModel.View`)

`cleanAmount` and `amountRegex` stay unexported — they're internal implementation details.

### Step 1.5 — Create testing_helper.go

In `internal/tui/testing_helper.go`:

```go
package tui_test

import (
    "database/sql"
    "path/filepath"
    "testing"

    "github.com/gjtiquia/vimance/internal/service"
    "github.com/gjtiquia/vimance/internal/tui"
    "github.com/pressly/goose/v3"
    _ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    database, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("open test db: %v", err)
    }
    t.Cleanup(func() { database.Close() })

    _, err = database.Exec("PRAGMA foreign_keys = ON")
    if err != nil {
        t.Fatalf("enable foreign keys: %v", err)
    }

    if err := goose.SetDialect("sqlite3"); err != nil {
        t.Fatalf("set goose dialect: %v", err)
    }

    migrationsDir := filepath.Join("..", "..", "db", "migrations")
    if err := goose.Up(database, migrationsDir); err != nil {
        t.Fatalf("run migrations: %v", err)
    }

    return database
}

func seedTestDB(t *testing.T, db *sql.DB) {
    t.Helper()

    _, err := db.Exec("INSERT INTO users (id, name) VALUES (1, 'testuser')")
    if err != nil {
        t.Fatalf("seed user: %v", err)
    }
    _, err = db.Exec("INSERT INTO currencies (id, code) VALUES (1, 'USD')")
    if err != nil {
        t.Fatalf("seed currency: %v", err)
    }
    _, err = db.Exec("INSERT INTO tags (id, name) VALUES (1, 'food')")
    if err != nil {
        t.Fatalf("seed tag: %v", err)
    }
}

func newTestModel(t *testing.T) tui.Model {
    t.Helper()
    return tui.NewModel(setupTestDB(t))
}

func newTestService(t *testing.T) *service.Service {
    t.Helper()
    db := setupTestDB(t)
    seedTestDB(t, db)
    return service.New(db)
}
```

Note: `seedTestDB` creates user (id=1), currency (id=1, code="USD"), tag (id=1, name="food"). These are needed by `LoadCurrencies`, `LoadTags`, and `SaveRecord` (which references user_id=1).

---

## Phase 2: validation_test.go — pure function tests

No DB, no tea, fully parallel.

```go
package tui_test

import (
    "testing"

    "github.com/gjtiquia/vimance/internal/tui"
)

func TestValidateAmount(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name  string
        input string
        valid bool
    }{
        {"valid cents", "12.50", true},
        {"valid whole", "100", true},
        {"valid zero", "0", true},
        {"valid decimal", "0.99", true},
        {"empty", "", false},
        {"letters", "abc", false},
        {"three decimals", "12.123", false},
        {"trailing dot", "12.", false},
        {"leading dot", ".50", false},
        {"comma", "1,000", false},
        {"negative", "-50", false},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := tui.ValidateAmount(tc.input)
            if tc.valid && err != nil {
                t.Errorf("expected valid, got error: %v", err)
            }
            if !tc.valid && err == nil {
                t.Error("expected error, got nil")
            }
        })
    }
}

func TestParseAmountToCents(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name   string
        input  string
        expect int64
        valid  bool
    }{
        {"12.50", "12.50", 1250, true},
        {"100", "100", 10000, true},
        {"0", "0", 0, true},
        {"0.99", "0.99", 99, true},
        {"1.00", "1.00", 100, true},
        {"10.0", "10.0", 1000, true},
        {"invalid", "abc", 0, false},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            cents, err := tui.ParseAmountToCents(tc.input)
            if tc.valid && err != nil {
                t.Errorf("expected valid, got error: %v", err)
            }
            if !tc.valid && err == nil {
                t.Fatal("expected error")
            }
            if tc.valid && cents != tc.expect {
                t.Errorf("expected %d cents, got %d", tc.expect, cents)
            }
        })
    }
}

func TestValidateDate(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name              string
        year, month, day  string
        valid             bool
    }{
        {"valid", "2026", "05", "09", true},
        {"empty year", "", "05", "09", false},
        {"empty month", "2026", "", "09", false},
        {"empty day", "2026", "05", "", false},
        {"month 13", "2026", "13", "01", false},
        {"day 32", "2026", "01", "32", false},
        {"feb 29 non-leap", "2025", "02", "29", false},
        {"feb 29 leap", "2024", "02", "29", true},
        {"garbage", "abc", "def", "ghi", false},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := tui.ValidateDate(tc.year, tc.month, tc.day)
            if tc.valid && err != nil {
                t.Errorf("expected valid, got: %v", err)
            }
            if !tc.valid && err == nil {
                t.Error("expected error")
            }
        })
    }
}

func TestFormatDate(t *testing.T) {
    t.Parallel()
    got := tui.FormatDate("2026", "05", "09")
    want := "2026-05-09"
    if got != want {
        t.Errorf("FormatDate = %q, want %q", got, want)
    }
}

func TestTruncate(t *testing.T) {
    t.Parallel()
    tests := []struct {
        input string
        max   int
        want  string
    }{
        {"hello", 5, "hello"},
        {"hello world", 5, "hello..."},
        {"hi", 5, "hi"},
        {"", 3, ""},
    }
    for _, tc := range tests {
        got := tui.Truncate(tc.input, tc.max)
        if got != tc.want {
            t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
        }
    }
}

func TestValidationErrors(t *testing.T) {
    t.Parallel()
    var errors tui.ValidationErrors
    if errors.HasErrors() {
        t.Error("empty errors should not have errors")
    }

    errors = append(errors, tui.ValidationError{Field: "date", Message: "bad date"})
    if !errors.HasErrors() {
        t.Error("non-empty errors should have errors")
    }

    if msg := errors.Get("date"); msg != "bad date" {
        t.Errorf("Get('date') = %q, want 'bad date'", msg)
    }
    if msg := errors.Get("nonexistent"); msg != "" {
        t.Errorf("Get('nonexistent') = %q, want ''", msg)
    }
}
```

---

## Phase 3: record_model_test.go

### Setup

RecordModel's `setActiveField` calls `LoadCurrencies`/`LoadTags` when entering currency/tags fields. These require a real service with DB. Tests that tab past date fields will trigger these calls.

```go
package tui_test

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/gjtiquia/vimance/internal/tui"
)

func setupRecordModel(t *testing.T) tui.RecordModel {
    t.Helper()
    svc := newTestService(t)
    return tui.NewRecordModel(svc)
}
```

### Tests

```go
func TestRecordModelInitialState(t *testing.T) {
    m := setupRecordModel(t)
    if m.State != tui.RecordStateEditing {
        t.Errorf("expected RecordStateEditing, got %v", m.State)
    }
    if m.ActiveField != tui.FieldDateYear {
        t.Errorf("expected FieldDateYear, got %v", m.ActiveField)
    }
}

func TestRecordModelFieldNavigation(t *testing.T) {
    m := setupRecordModel(t)
    fields := []tui.ActiveField{
        tui.FieldDateYear, tui.FieldDateMonth, tui.FieldDateDay,
        tui.FieldCurrency, tui.FieldTags, tui.FieldAmount,
        tui.FieldLinks, tui.FieldNotes,
    }

    for i, expected := range fields {
        if m.ActiveField != expected {
            t.Errorf("step %d: expected %v, got %v", i, expected, m.ActiveField)
        }
        if i < len(fields)-1 {
            m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
        }
    }

    // Tab at end stays at FieldNotes (nextField returns same field)
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    if m.ActiveField != tui.FieldNotes {
        t.Errorf("tab at end should stay at FieldNotes, got %v", m.ActiveField)
    }

    // Shift+tab backward
    for i := len(fields) - 2; i >= 0; i-- {
        m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
        if m.ActiveField != fields[i] {
            t.Errorf("shift+tab step %d: expected %v, got %v", i, fields[i], m.ActiveField)
        }
    }
}

func TestRecordModelEnterOnDateFillsDefault(t *testing.T) {
    m := setupRecordModel(t)
    // Enter on FieldDateYear fills placeholder and advances
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.ActiveField != tui.FieldDateMonth {
        t.Errorf("expected FieldDateMonth after enter on year, got %v", m.ActiveField)
    }
    if m.DateYearInput.Value() == "" {
        t.Error("expected year to be filled with placeholder default")
    }
}

func TestRecordModelEnterOnNotesGoesToConfirm(t *testing.T) {
    m := setupRecordModel(t)
    // Tab 7 times to reach FieldNotes
    for i := 0; i < 7; i++ {
        m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    }
    if m.ActiveField != tui.FieldNotes {
        t.Fatalf("expected FieldNotes, got %v", m.ActiveField)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.RecordStateConfirm {
        t.Errorf("expected RecordStateConfirm, got %v", m.State)
    }
}

func TestRecordModelConfirmNumberKeys(t *testing.T) {
    m := setupRecordModel(t)
    // Tab to FieldNotes then enter to confirm
    for i := 0; i < 7; i++ {
        m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    }
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    numberFieldMap := map[string]tui.ActiveField{
        "1": tui.FieldDateYear,
        "2": tui.FieldCurrency,
        "3": tui.FieldTags,
        "4": tui.FieldAmount,
        "5": tui.FieldLinks,
        "6": tui.FieldNotes,
    }
    for key, expected := range numberFieldMap {
        mCopy := m // test each from the same confirm state
        mCopy, _ = mCopy.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
        if mCopy.State != tui.RecordStateEditing {
            t.Errorf("key %q: expected RecordStateEditing, got %v", key, mCopy.State)
        }
        if mCopy.ActiveField != expected {
            t.Errorf("key %q: expected %v, got %v", key, expected, mCopy.ActiveField)
        }
    }
}

func TestRecordModelConfirmEscReturnsToEditing(t *testing.T) {
    m := setupRecordModel(t)
    for i := 0; i < 7; i++ {
        m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    }
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

    if m.State != tui.RecordStateEditing {
        t.Errorf("expected RecordStateEditing after esc, got %v", m.State)
    }
    if m.ActiveField != tui.FieldNotes {
        t.Errorf("expected FieldNotes after esc from confirm, got %v", m.ActiveField)
    }
}

func TestRecordModelConfirmEnterWithErrors(t *testing.T) {
    m := setupRecordModel(t)
    // Tab past date fields (fills defaults) and through to notes
    for i := 0; i < 7; i++ {
        m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    }
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    // No currency selected, no amount typed → should have errors
    if !m.ConfirmModel.Errors.HasErrors() {
        t.Error("expected validation errors (no currency, no amount)")
    }

    // Enter on confirm with errors → should NOT transition to success
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.RecordStateConfirm {
        t.Errorf("should stay in confirm when errors exist, got %v", m.State)
    }
}

func TestRecordModelConfirmEnterNoErrors(t *testing.T) {
    // Full flow: fill all required fields, then confirm
    db := setupTestDB(t)
    seedTestDB(t, db)
    svc := newTestService(t)
    m := tui.NewRecordModel(svc)

    // Tab through date fields (fills defaults): year→month→day
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

    // At FieldCurrency — type "USD" and enter to select
    m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    // CurrencyModel.ShouldAdvance → tab advances to FieldTags
    // But ShouldAdvance is handled in updateEditing — after the batch update,
    // ShouldAdvance is checked. Need to verify this works correctly.
    // If CurrencyInput.ShouldAdvance is true, the next updateEditing call
    // sets active field to FieldTags.

    // Tab past Tags
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

    // At FieldAmount — type "10.00"
    m, _ = m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
    m, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // tab to Links

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // tab to Notes

    // Enter on Notes → confirm
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.RecordStateConfirm {
        t.Fatalf("expected RecordStateConfirm, got %v", m.State)
    }

    if m.ConfirmModel.Errors.HasErrors() {
        t.Fatalf("expected no errors, got: %v", m.ConfirmModel.Errors)
    }

    // Enter to save → success
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.RecordStateSuccess {
        t.Errorf("expected RecordStateSuccess, got %v", m.State)
    }
}
```

Note: `TestRecordModelConfirmEnterNoErrors` is the most complex test. The exact keypress flow for currency typing + selection may need debugging — `CurrencyModel.Update` processes character keys through `SearchInput`, and `ShouldAdvance` handling in `updateEditing` may interact with the batch update pattern.

---

## Phase 4: query_model_test.go

### Setup

```go
func setupQueryModel(t *testing.T) tui.QueryModel {
    t.Helper()
    svc := newTestService(t)
    m := tui.NewQueryModel(svc)
    return m
}

func setupQueryModelInFilterForm(t *testing.T) tui.QueryModel {
    t.Helper()
    m := setupQueryModel(t)
    m.State = tui.QueryStateFilterForm
    m.ActiveField = tui.FilterDateFrom
    m.FocusActiveField() // focus DateFrom so textinput processes keypresses
    return m
}
```

### Tests

```go
func TestQueryModelInitialState(t *testing.T) {
    m := setupQueryModel(t)
    if m.State != tui.QueryStateMenu {
        t.Errorf("expected QueryStateMenu, got %v", m.State)
    }
}

func TestQueryModelFilterFieldNavigation(t *testing.T) {
    m := setupQueryModelInFilterForm(t)
    fields := []tui.FilterField{
        tui.FilterDateFrom, tui.FilterDateTo,
        tui.FilterCurrency, tui.FilterTags,
        tui.FilterFuzzy,
    }

    for i, expected := range fields {
        if m.ActiveField != expected {
            t.Errorf("step %d: expected %v, got %v", i, expected, m.ActiveField)
        }
        if i < len(fields)-1 {
            m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
        }
    }

    // Tab on Fuzzy → QueryStateConfirm
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    if m.State != tui.QueryStateConfirm {
        t.Errorf("expected QueryStateConfirm after tab on Fuzzy, got %v", m.State)
    }
}

func TestQueryModelEscFromFilterForm(t *testing.T) {
    m := setupQueryModelInFilterForm(t)
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateMenu {
        t.Errorf("expected QueryStateMenu after esc, got %v", m.State)
    }
}

func TestQueryModelConfirmEsc(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateConfirm
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateFilterForm {
        t.Errorf("expected QueryStateFilterForm after esc, got %v", m.State)
    }
}

func TestQueryModelConfirmNumberKeys(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateConfirm

    numberFieldMap := map[string]tui.FilterField{
        "1": tui.FilterDateFrom,
        "2": tui.FilterCurrency,
        "3": tui.FilterTags,
        "4": tui.FilterFuzzy,
    }
    for key, expected := range numberFieldMap {
        mCopy := m
        mCopy, _ = mCopy.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
        if mCopy.State != tui.QueryStateFilterForm {
            t.Errorf("key %q: expected QueryStateFilterForm, got %v", key, mCopy.State)
        }
        if mCopy.ActiveField != expected {
            t.Errorf("key %q: expected %v, got %v", key, expected, mCopy.ActiveField)
        }
    }
}

func TestQueryModelConfirmEnterWithValidDates(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateConfirm
    // NewQueryModel pre-fills DateFrom/DateTo with current month defaults

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.QueryStateResults {
        t.Errorf("expected QueryStateResults, got %v", m.State)
    }
}

func TestQueryModelConfirmEnterInvalidDateFrom(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateConfirm
    m.DateFrom.SetValue("not-a-date")

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.QueryStateFilterForm {
        t.Errorf("expected QueryStateFilterForm on invalid date, got %v", m.State)
    }
    if m.ErrorMsg == "" {
        t.Error("expected error message for invalid date")
    }
}

func TestQueryModelConfirmEnterInvalidDateTo(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateConfirm
    m.DateTo.SetValue("bad")

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.State != tui.QueryStateFilterForm {
        t.Errorf("expected QueryStateFilterForm, got %v", m.State)
    }
}

func TestQueryModelResultsCursorNavigation(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{
        {ID: 1, Date: "2026-05-01", AmountCents: 1000},
        {ID: 2, Date: "2026-05-02", AmountCents: 2000},
        {ID: 3, Date: "2026-05-03", AmountCents: 3000},
        {ID: 4, Date: "2026-05-04", AmountCents: 4000},
        {ID: 5, Date: "2026-05-05", AmountCents: 5000},
    }
    m.CursorIndex = 0

    // j moves down
    m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
    if m.CursorIndex != 1 {
        t.Errorf("expected CursorIndex=1 after j, got %d", m.CursorIndex)
    }

    // down arrow moves down
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    if m.CursorIndex != 2 {
        t.Errorf("expected CursorIndex=2 after down, got %d", m.CursorIndex)
    }

    // k moves up
    m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
    if m.CursorIndex != 1 {
        t.Errorf("expected CursorIndex=1 after k, got %d", m.CursorIndex)
    }

    // up arrow moves up
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0 after up, got %d", m.CursorIndex)
    }

    // up at top stays at 0
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0 at top, got %d", m.CursorIndex)
    }

    // G goes to end
    m, _ = m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
    if m.CursorIndex != 4 {
        t.Errorf("expected CursorIndex=4 after G, got %d", m.CursorIndex)
    }

    // g goes to start
    m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0 after g, got %d", m.CursorIndex)
    }
}

func TestQueryModelResultsEnterSelectsRecord(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{
        {ID: 42, Date: "2026-05-01", AmountCents: 1000},
    }
    m.CursorIndex = 0

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if m.SelectedID != 42 {
        t.Errorf("expected SelectedID=42, got %d", m.SelectedID)
    }
}

func TestQueryModelResultsEscToFilterForm(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
    m.ResultsOrigin = tui.QueryStateFilterForm

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateFilterForm {
        t.Errorf("expected QueryStateFilterForm, got %v", m.State)
    }
}

func TestQueryModelResultsEscToSavedList(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}
    m.ResultsOrigin = tui.QueryStateSavedList

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateSavedList {
        t.Errorf("expected QueryStateSavedList, got %v", m.State)
    }
}

func TestQueryModelResultsSaveName(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{{ID: 1, Date: "2026-05-01"}}

    m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
    if m.State != tui.QueryStateSaveName {
        t.Errorf("expected QueryStateSaveName after s, got %v", m.State)
    }

    // Esc from save name → back to results
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateResults {
        t.Errorf("expected QueryStateResults after esc, got %v", m.State)
    }
}

func TestQueryModelResultsEmpty(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateResults
    m.Results = []service.QueryResult{}

    // Esc from empty results → back to filter form
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateFilterForm {
        t.Errorf("expected QueryStateFilterForm, got %v", m.State)
    }
}

func TestQueryModelDeleteConfirm(t *testing.T) {
    m := setupQueryModel(t)
    m.State = tui.QueryStateDeleteConfirm
    m.DeleteTarget = tui.SavedQueryItem{ID: 1, Name: "test query"}

    // "y" → deletes and reloads saved queries (requires DB with saved query)
    // This is hard to test without a real saved query in DB.
    // Skip for now — covered by integration tests.

    // "n" → back to saved list
    m, _ = m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
    if m.State != tui.QueryStateSavedList {
        t.Errorf("expected QueryStateSavedList after n, got %v", m.State)
    }

    // "esc" → back to saved list
    m.State = tui.QueryStateDeleteConfirm
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.State != tui.QueryStateSavedList {
        t.Errorf("expected QueryStateSavedList after esc, got %v", m.State)
    }
}
```

Note: `SavedQueryItem` and `DeleteTarget` need to be exported for the delete confirm test. Check export status during implementation — if `DeleteTarget` is unexported, add it to Step 1.2 or skip this test.

---

## Phase 5: currency_model_test.go

No DB needed. `NewCurrencyModel(nil)` safe — `Update()` never calls service.

```go
package tui_test

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/gjtiquia/vimance/internal/tui"
)

func TestCurrencyModelInitialState(t *testing.T) {
    m := tui.NewCurrencyModel(nil)
    if m.Mode != tui.CurrencyModeInsert {
        t.Errorf("expected CurrencyModeInsert, got %v", m.Mode)
    }
    if m.Selected != nil {
        t.Error("expected no selected currency")
    }
    if m.ShouldAdvance {
        t.Error("expected ShouldAdvance=false initially")
    }
}

func TestCurrencyModelEnterNewCurrency(t *testing.T) {
    m := tui.NewCurrencyModel(nil)
    m.SearchInput.Focus() // focus required for textinput to process keys

    m, _ = m.Update(tea.KeyPressMsg{Code: 'E', Text: "E"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    if m.Selected == nil {
        t.Fatal("expected selected currency after typing and entering")
    }
    if m.Selected.Code != "EUR" {
        t.Errorf("expected EUR, got %s", m.Selected.Code)
    }
    if !m.Selected.IsNew {
        t.Error("expected IsNew=true for typed currency")
    }
    if !m.ShouldAdvance {
        t.Error("expected ShouldAdvance=true after enter")
    }
}

func TestCurrencyModelSelectFromList(t *testing.T) {
    m := tui.NewCurrencyModel(nil)
    m.AllCurrencies = []tui.CurrencyItem{
        {ID: 1, Code: "USD"},
        {ID: 2, Code: "EUR"},
    }
    m.Mode = tui.CurrencyModeNormal

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    if m.Selected == nil {
        t.Fatal("expected selected currency")
    }
    if m.Selected.Code != "EUR" {
        t.Errorf("expected EUR, got %s", m.Selected.Code)
    }
    if m.Selected.IsNew {
        t.Error("expected IsNew=false for existing currency")
    }
}

func TestCurrencyModelModeToggle(t *testing.T) {
    m := tui.NewCurrencyModel(nil)
    m.Mode = tui.CurrencyModeInsert

    // esc → normal
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.Mode != tui.CurrencyModeNormal {
        t.Errorf("expected CurrencyModeNormal after esc, got %v", m.Mode)
    }

    // i → insert
    m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
    if m.Mode != tui.CurrencyModeInsert {
        t.Errorf("expected CurrencyModeInsert after i, got %v", m.Mode)
    }

    // a → also insert
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc}) // back to normal
    m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
    if m.Mode != tui.CurrencyModeInsert {
        t.Errorf("expected CurrencyModeInsert after a, got %v", m.Mode)
    }
}

func TestCurrencyModelCursorUpDown(t *testing.T) {
    m := tui.NewCurrencyModel(nil)
    m.AllCurrencies = []tui.CurrencyItem{
        {ID: 1, Code: "USD"},
        {ID: 2, Code: "EUR"},
        {ID: 3, Code: "GBP"},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    if m.CursorIndex != 1 {
        t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    if m.CursorIndex != 2 {
        t.Errorf("expected CursorIndex=2, got %d", m.CursorIndex)
    }

    // Down at bottom stays
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    if m.CursorIndex != 2 {
        t.Errorf("expected CursorIndex=2 at bottom, got %d", m.CursorIndex)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 1 {
        t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
    }

    // Up at top stays
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0 at top, got %d", m.CursorIndex)
    }
}
```

Note: `TestCurrencyModelEnterNewCurrency` calls `m.SearchInput.Focus()` because the textinput must be focused to process character keypresses. `CurrencyModel` in insert mode has `SearchInput` focused, but `NewCurrencyModel(nil)` doesn't call `Focus()` on it. However, `setActiveField` in RecordModel does focus it via `focusActiveField`. For standalone tests, we need to focus manually.

---

## Phase 6: tags_model_test.go

No DB needed. `NewTagsModel(nil)` safe — `Update()` never calls service.

```go
package tui_test

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/gjtiquia/vimance/internal/tui"
)

func TestTagsModelInitialState(t *testing.T) {
    m := tui.NewTagsModel(nil)
    if m.Mode != tui.TagModeInsert {
        t.Errorf("expected TagModeInsert, got %v", m.Mode)
    }
    if len(m.SelectedTags) != 0 {
        t.Errorf("expected no selected tags, got %d", len(m.SelectedTags))
    }
}

func TestTagsModelAddNewTag(t *testing.T) {
    m := tui.NewTagsModel(nil)
    m.SearchInput.Focus() // focus required for textinput

    m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    if len(m.SelectedTags) != 1 {
        t.Fatalf("expected 1 selected tag, got %d", len(m.SelectedTags))
    }
    if m.SelectedTags[0].Name != "foo" {
        t.Errorf("expected tag name 'foo', got %q", m.SelectedTags[0].Name)
    }
    if !m.SelectedTags[0].IsNew {
        t.Error("expected IsNew=true for typed tag")
    }
}

func TestTagsModelAddExistingTag(t *testing.T) {
    m := tui.NewTagsModel(nil)
    m.AllTags = []tui.TagItem{{ID: 1, Name: "food"}}
    m.Mode = tui.TagModeNormal

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if len(m.SelectedTags) != 1 {
        t.Fatalf("expected 1 selected tag, got %d", len(m.SelectedTags))
    }
    if m.SelectedTags[0].IsNew {
        t.Error("expected IsNew=false for existing tag")
    }
    if m.SelectedTags[0].Name != "food" {
        t.Errorf("expected 'food', got %q", m.SelectedTags[0].Name)
    }
}

func TestTagsModelDuplicateIgnored(t *testing.T) {
    m := tui.NewTagsModel(nil)
    m.SearchInput.Focus()

    m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    // Try adding same tag again
    m.SearchInput.Focus()
    m, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

    if len(m.SelectedTags) != 1 {
        t.Errorf("expected 1 tag (duplicate ignored), got %d", len(m.SelectedTags))
    }
}

func TestTagsModelRemoveLastTag(t *testing.T) {
    m := tui.NewTagsModel(nil)
    m.SelectedTags = []tui.TagItem{
        {ID: 1, Name: "food"},
        {ID: 2, Name: "drinks"},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
    if len(m.SelectedTags) != 1 {
        t.Fatalf("expected 1 tag after ctrl+z, got %d", len(m.SelectedTags))
    }
    if m.SelectedTags[0].Name != "food" {
        t.Errorf("expected remaining tag 'food', got %q", m.SelectedTags[0].Name)
    }

    // ctrl+z on empty is no-op
    m.SelectedTags = nil
    m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
    if len(m.SelectedTags) != 0 {
        t.Error("expected 0 tags after ctrl+z on empty")
    }
}

func TestTagsModelModeToggle(t *testing.T) {
    m := tui.NewTagsModel(nil)
    m.Mode = tui.TagModeInsert

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.Mode != tui.TagModeNormal {
        t.Errorf("expected TagModeNormal after esc, got %v", m.Mode)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
    if m.Mode != tui.TagModeInsert {
        t.Errorf("expected TagModeInsert after a, got %v", m.Mode)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
    if m.Mode != tui.TagModeInsert {
        t.Errorf("expected TagModeInsert after i, got %v", m.Mode)
    }
}
```

---

## Phase 7: links_model_test.go

No DB needed. `NewLinksModel(nil)` safe — `Update()` never calls service.

```go
package tui_test

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/gjtiquia/vimance/internal/tui"
)

func TestLinksModelInitialState(t *testing.T) {
    m := tui.NewLinksModel(nil)
    if m.Mode != tui.LinkModeInsert {
        t.Errorf("expected LinkModeInsert, got %v", m.Mode)
    }
    if len(m.SelectedParents) != 0 {
        t.Errorf("expected no parents, got %d", len(m.SelectedParents))
    }
}

func TestLinksModelAddParent(t *testing.T) {
    m := tui.NewLinksModel(nil)
    m.FilteredCandidates = []tui.LinkedRecord{
        {ID: 1, Date: "2026-05-01", Notes: "test record", AmountCents: 1000},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if len(m.SelectedParents) != 1 {
        t.Fatalf("expected 1 parent, got %d", len(m.SelectedParents))
    }
    if m.SelectedParents[0].ID != 1 {
        t.Errorf("expected parent ID=1, got %d", m.SelectedParents[0].ID)
    }
}

func TestLinksModelDuplicateParentIgnored(t *testing.T) {
    m := tui.NewLinksModel(nil)
    m.FilteredCandidates = []tui.LinkedRecord{
        {ID: 1, Date: "2026-05-01", Notes: "test"},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    // Select same again
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    if len(m.SelectedParents) != 1 {
        t.Errorf("expected 1 parent (duplicate ignored), got %d", len(m.SelectedParents))
    }
}

func TestLinksModelRemoveLastParent(t *testing.T) {
    m := tui.NewLinksModel(nil)
    m.SelectedParents = []tui.LinkedRecord{
        {ID: 1, Date: "2026-05-01"},
        {ID: 2, Date: "2026-05-02"},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
    if len(m.SelectedParents) != 1 {
        t.Fatalf("expected 1 parent after ctrl+z, got %d", len(m.SelectedParents))
    }
    if m.SelectedParents[0].ID != 1 {
        t.Errorf("expected remaining parent ID=1, got %d", m.SelectedParents[0].ID)
    }
}

func TestLinksModelModeToggle(t *testing.T) {
    m := tui.NewLinksModel(nil)
    m.Mode = tui.LinkModeInsert

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    if m.Mode != tui.LinkModeNormal {
        t.Errorf("expected LinkModeNormal after esc, got %v", m.Mode)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
    if m.Mode != tui.LinkModeInsert {
        t.Errorf("expected LinkModeInsert after i, got %v", m.Mode)
    }
}

func TestLinksModelUpDown(t *testing.T) {
    m := tui.NewLinksModel(nil)
    m.FilteredCandidates = []tui.LinkedRecord{
        {ID: 1, Date: "2026-05-01"},
        {ID: 2, Date: "2026-05-02"},
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    if m.CursorIndex != 1 {
        t.Errorf("expected CursorIndex=1, got %d", m.CursorIndex)
    }

    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0, got %d", m.CursorIndex)
    }

    // Up at top stays
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
    if m.CursorIndex != 0 {
        t.Errorf("expected CursorIndex=0 at top, got %d", m.CursorIndex)
    }
}
```

---

## Phase 8: model_test.go — top-level smoke tests

### Tests

```go
package tui_test

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/gjtiquia/vimance/internal/tui"
)

func TestModelInitialState(t *testing.T) {
    m := newTestModel(t)
    if m.InputType != tui.InputTypeList {
        t.Errorf("expected InputTypeList, got %v", m.InputType)
    }
}

func TestModelCtrlCQuits(t *testing.T) {
    m := newTestModel(t)
    _, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
    if cmd == nil {
        t.Fatal("expected non-nil cmd")
    }
    msg := cmd()
    if _, ok := msg.(tea.QuitMsg); !ok {
        t.Errorf("expected tea.QuitMsg, got %T", msg)
    }
}

func TestModelListToCreate(t *testing.T) {
    m := newTestModel(t)

    // First list item is "create" at index 0, enter selects it
    result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m = result.(tui.Model)

    if m.InputType != tui.InputTypeRecord {
        t.Errorf("expected InputTypeRecord, got %v", m.InputType)
    }
    if m.RecordInput.State != tui.RecordStateEditing {
        t.Errorf("expected RecordStateEditing, got %v", m.RecordInput.State)
    }
}

func TestModelListToQueryAndBack(t *testing.T) {
    m := newTestModel(t)

    // Down to select "query" (index 1), then enter
    m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
    result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m = result.(tui.Model)

    if m.InputType != tui.InputTypeQuery {
        t.Errorf("expected InputTypeQuery, got %v", m.InputType)
    }
    if m.QueryInput.State != tui.QueryStateMenu {
        t.Errorf("expected QueryStateMenu, got %v", m.QueryInput.State)
    }

    // Esc from query menu → back to list
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
    m = result.(tui.Model)
    if m.InputType != tui.InputTypeList {
        t.Errorf("expected InputTypeList after esc, got %v", m.InputType)
    }
}

func TestModelWindowSize(t *testing.T) {
    m := newTestModel(t)
    result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
    m = result.(tui.Model)
    if m.Width != 100 || m.Height != 50 {
        t.Errorf("expected (100,50), got (%d,%d)", m.Width, m.Height)
    }
}

func TestModelRecordCreateIntegration(t *testing.T) {
    db := setupTestDB(t)
    seedTestDB(t, db)
    m := tui.NewModel(db)

    // Select "create" from list
    result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m = result.(tui.Model)

    // Tab through date fields (fills defaults)
    for i := 0; i < 3; i++ {
        result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
        m = result.(tui.Model)
    }

    // At FieldCurrency — type "USD"
    result, _ = m.Update(tea.KeyPressMsg{Code: 'U', Text: "U"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: 'S', Text: "S"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m = result.(tui.Model)

    // ShouldAdvance → advances past currency
    // Tab past tags
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
    m = result.(tui.Model)

    // Type amount "10.00"
    result, _ = m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: '.', Text: "."})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
    m = result.(tui.Model)

    // Tab past links and notes — enter on notes
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // links
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // notes
    m = result.(tui.Model)
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // notes → confirm
    m = result.(tui.Model)

    if m.RecordInput.State != tui.RecordStateConfirm {
        t.Fatalf("expected RecordStateConfirm, got %v", m.RecordInput.State)
    }

    // Enter to save
    result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
    m = result.(tui.Model)

    if m.RecordInput.State != tui.RecordStateSuccess {
        t.Errorf("expected RecordStateSuccess, got %v", m.RecordInput.State)
    }
}
```

Note: `TestModelRecordCreateIntegration` is the most complex test. The currency typing + ShouldAdvance flow may need debugging. The `CurrencyModel.Update` processes character keys through `SearchInput` (which must be focused — `setActiveField` handles this). The `ShouldAdvance` flag is checked in `RecordModel.updateEditing` after the batch update.

---

## Phase 9: Run and fix

```bash
go test ./internal/tui/... -v -count=1
```

### Potential compile issues to check

- `tea.WindowSizeMsg` — verify exact type name (could be `tea.WindowSize`)
- `service.QueryResult` — verify all fields are exported (confirmed: ID, Date, AmountCents, AmountStr, CurrencyCode, Notes, TagNames all exported)
- `tui.SavedQueryItem` — verify fields are exported for delete confirm test
- `tui.DeleteTarget` — if unexported, add to Step 1.2 or skip `TestQueryModelDeleteConfirm`

### Potential runtime issues to debug

- `textinput.Model.Update` ignores keypresses when not focused — ensure all test setups call `.Focus()` where needed
- `CurrencyModel.ShouldAdvance` flag — when `enter` selects a currency in insert mode, `ShouldAdvance=true` is set. In `RecordModel.updateEditing`, this is checked AFTER the batch update of all sub-widgets. The flag is only checked if the currency was just selected.
- `list.Model` in top-level Model — the `UpdateListInput` method catches `enter` before passing to `listInput.Update`. The list must have visible items and be in `Filtering` state for `down`/`up` to work. After `EnterListInput()`, the list starts with all items visible and index 0 selected.

## Implementation order

| Step | What | Depends on | Risk |
|---|---|---|---|
| 1a | Export Model fields in `app.go` | — | low |
| 1b | Export QueryModel fields in `query.go` | — | low |
| 1c | Export `FocusActiveField` in `query.go` | — | low |
| 1d | Export validation funcs in `validation.go` | — | low |
| 1e | Create `testing_helper.go` | — | low |
| 2 | `validation_test.go` | 1d | low |
| 3 | `currency_model_test.go` | 1e | low |
| 4 | `tags_model_test.go` | 1e | low |
| 5 | `links_model_test.go` | 1e | low |
| 6 | `record_model_test.go` | 1e | medium |
| 7 | `query_model_test.go` | 1b, 1c, 1e | medium |
| 8 | `model_test.go` | 1a, 1e | medium-high |

Each step can be committed after passing tests.

## Commit plan

```
commit 1: export Model fields + QueryModel fields + FocusActiveField + validation funcs
commit 2: testing_helper.go + validation_test.go
commit 3: sub-model tests (currency, tags, links)
commit 4: record_model_test.go
commit 5: query_model_test.go
commit 6: model_test.go
```
