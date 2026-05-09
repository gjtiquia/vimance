# Query View Implementation Plan

## Overview

Build a query view for vimance that lets users search, browse, and edit existing records.
Follows BubbleTea state machine pattern established by RecordModel.

## Flow

```
App menu ─── "create" → create record flow (existing)
           └── "query" → QueryModel
                 ├── "new query" → filter form → confirm → results
                 └── "saved queries" → saved list → pick → results (immediate)

Results → enter → RecordModel edit mode (pre-filled) → save → back to results
        → s → save name prompt → save → stay in results
        → esc → back to filter form (pre-filled)

Filter form → esc → back to query menu
Saved list  → esc → back to query menu
Saved list  → d → delete confirm → y/n → refresh list
```

## Phase 1: DB Schema

Note: all new tables use hard delete (no soft delete). Saved queries are user convenience config, not financial data. Easy to recreate.

Append to `db/migrations/20260423000000_initial_schema.sql`:

```sql
CREATE TABLE saved_queries (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    date_from TEXT NOT NULL,
    date_to TEXT NOT NULL,
    currency_id INTEGER REFERENCES currencies(id),
    fuzzy_text TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL,
    updated_by INTEGER NOT NULL REFERENCES users(id)
);

CREATE TABLE saved_query_tags (
    query_id INTEGER NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (query_id, tag_id)
);
```

Add Down section — must be placed BEFORE existing table drops (before `DROP TABLE IF EXISTS pinned_tags;`) since `saved_query_tags` references `tags` and `saved_queries` references `currencies/users`:

```sql
DROP TABLE IF EXISTS saved_query_tags;
DROP TABLE IF EXISTS saved_queries;
```

Note: append only. User deletes `vimance.db` to recreate from scratch.

## Phase 2: sqlc Queries

### `db/queries/saved_queries.sql`

```sql
-- name: CreateSavedQuery :one
INSERT INTO saved_queries (name, date_from, date_to, currency_id, fuzzy_text, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSavedQuery :one
SELECT * FROM saved_queries WHERE id = ?;

-- name: ListSavedQueries :many
SELECT * FROM saved_queries ORDER BY name;

-- name: UpdateSavedQuery :one
UPDATE saved_queries
SET name = ?, date_from = ?, date_to = ?, currency_id = ?, fuzzy_text = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSavedQuery :exec
DELETE FROM saved_queries WHERE id = ?;
```

### `db/queries/saved_query_tags.sql`

```sql
-- name: AddSavedQueryTag :exec
INSERT INTO saved_query_tags (query_id, tag_id) VALUES (?, ?);

-- name: RemoveSavedQueryTag :exec
DELETE FROM saved_query_tags WHERE query_id = ? AND tag_id = ?;

-- name: RemoveAllSavedQueryTags :exec
DELETE FROM saved_query_tags WHERE query_id = ?;

-- name: ListSavedQueryTags :many
SELECT t.* FROM tags t
INNER JOIN saved_query_tags sqt ON t.id = sqt.tag_id
WHERE sqt.query_id = ?
ORDER BY t.name;
```

### `db/queries/record_links.sql` — add query

```sql
-- name: RemoveAllRecordLinks :exec
DELETE FROM record_links WHERE child_id = ?;
```

### `db/queries/saved_queries.sql` — note

`ListSavedQueries` returns `[]db.SavedQuery` (no tags). TUI maps to internal `SavedQueryItem` with name + date range display (see Phase 7). Tags optionally fetched per query for display.

`currency_id` is nullable → sqlc generates `sql.NullInt64` in struct fields (CreateSavedQueryParams, UpdateSavedQueryParams, etc.). Service layer converts between `*int64` and `sql.NullInt64`.

### Saved query loading into filter form

When a saved query is picked from list:

1. `GetSavedQuery(id)` → get row as `db.SavedQuery` (handle `sql.NullInt64` for currency_id)
2. `GetSavedQueryTags(id)` → get `[]db.Tag` → extract `[]int64` tag IDs
3. Lookup currency code: if `currency_id` is set, call `GetCurrency(id)` for the code
4. Set `QueryModel.DateFrom`, `.DateTo`, `.Currency.Selected`, `.Tags.SelectedTags`, `.Fuzzy`
5. Call `svc.QueryRecords(dateFrom, dateTo, currencyID, tagIDs, fuzzyText)` → populate `results`
6. Show results immediately
7. Esc from results → filter form (pre-filled from step 4)

Need a method on `QueryModel`:

```go
func (m *QueryModel) LoadSavedQuery(sq db.SavedQuery, tagIDs []int64, currencyCode string)
```

This sets all filter form fields. `CurrencyModel` must load currencies first before we can set `.Selected`.

### Saved query list display

Each item shows: `name` + `date_from → date_to` (e.g., "Monthly groceries  |  2026-05-01 → 2026-05-31").
Tags not shown in list (to keep items compact). Full details visible on save confirm.

`ListItem` lacks an `ID` field. For saved queries, create a `SavedQueryListItem` type implementing `list.Item` with an `ID` field so selected queries can be identified by ID, not by name.

### `db/queries/records_tags.sql` — note

Already has `GetRecordTagsByIDs` (in `record_links.sql`). Used by service for batch tag enrichment. No changes needed.

Run `sqlc generate` after all sql files written.

## Phase 3: Service Layer

### `internal/service/query_records.go`

Query result struct + combined filter method:

```go
type QueryResult struct {
    ID           int64
    Date         string
    AmountCents  int64
    AmountStr    string
    CurrencyCode string
    Notes        string
    TagNames     []string
}

func (s *Service) QueryRecords(ctx, dateFrom, dateTo string, currencyID *int64, tagIDs []int64, fuzzy string) ([]QueryResult, error)
```

Implementation:
1. `ListActiveRecordsByDateRange(ctx, dateFrom, dateTo)` — DB fetch
2. Filter by currencyID (if non-nil)
3. Batch-fetch tags for all records via `GetRecordTagsByIDs`
4. Filter by ALL tags match (Go: check all selected tag IDs present)
5. Filter by fuzzy (substring match on notes + tag names, case-insensitive)
6. Enrich: load all currencies into `map[int64]string`, attach code to each result
7. Sort: `date DESC, created_at DESC` (Go sort, re-sort after DB)
8. Return `[]QueryResult`

### `internal/service/saved_queries.go`

```go
func (s *Service) CreateSavedQuery(ctx, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, createdBy int64) (db.SavedQuery, error)

func (s *Service) CreateSavedQueryWithTags(ctx, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, createdBy int64, tagIDs []int64) (db.SavedQuery, error)

func (s *Service) GetSavedQuery(ctx, id int64) (db.SavedQuery, error)

func (s *Service) ListSavedQueries(ctx) ([]db.SavedQuery, error)

func (s *Service) GetSavedQueryTags(ctx, queryID int64) ([]db.Tag, error)

func (s *Service) UpdateSavedQuery(ctx, id int64, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, updatedBy int64) (db.SavedQuery, error)

func (s *Service) DeleteSavedQuery(ctx, id int64) error
```

Use inline tx pattern with `WithTransaction` for `CreateSavedQueryWithTags` (insert query + tags in tx).

### `internal/service/records.go` — additions

```go
type RecordFull struct {
    Record       db.Record
    CurrencyCode string
    Tags         []db.Tag    // tags for this record
    Parents      []LinkCandidate
}
```

### Tests — service layer

Write tests alongside service files:

- `query_records_test.go` — `TestQueryRecords`:
  - Date range only (happy path)
  - Date range + currency filter
  - Date range + ALL tags filter (single tag, multiple tags, partial match → excluded)
  - Date range + fuzzy text (notes match, tag names match, no match)
  - All filters combined
  - Empty results
- `saved_queries_test.go` — CRUD + tag association
- `records_test.go` additions:
  - `TestGetRecordFull` — record + tags + currency + parents
  - `TestUpdateRecordWithTagsAndLinks` — verify links updated (old removed, new added)
- `validation_test.go` — `TestFormatCents` (0, 100, 150, 1500, -500)

`GetRecordFull`:
1. Get record by ID → `db.Record`
2. Get tags for this record → `[]db.Tag` directly (no need for custom TagItem here)
3. Get currency code: `GetCurrency(record.CurrencyID)` → extract `Code`
4. Get parent records: `GetRecordParents(id)` → `[]db.ActiveRecord`
5. Get parents' tags: collect parent IDs → `GetRecordTagsByIDs(parentIDs)` → map of `recordID → []string`
6. Build `[]LinkCandidate` for parents (with tag names, dates, amounts, notes)

`RecordFull.Parents` uses `[]LinkCandidate` from `record_links.go` (already has tag names).

`UpdateRecordWithTagsAndLinks`: wraps `UpdateRecord` + `RemoveAllRecordTags` + re-add tags + `RemoveAllRecordLinks` + re-add links in a single tx.

### `internal/service/record_links.go` — add passthrough

```go
func (s *Service) RemoveAllRecordLinks(ctx, childID int64) error
```

## Phase 4: TUI — Refactor

### `FormatCents` — shared utility

`FormatCents` needed by both service (to build `QueryResult.AmountStr`) and TUI (links.go refactor).

Add as method on `Service` in `internal/service/currencies.go` (or new `internal/service/format.go`):

```go
func (s *Service) FormatCents(cents int64) string
```

Service calls `s.FormatCents()` when building `QueryResult`. TUI calls `s.service.FormatCents()` for links.go.

Test in `currencies_test.go` or new `format_test.go`: covers 0, <100, exact 100, >100, negative, large values.

### `internal/tui/links.go` — use `s.service.FormatCents`

Replace inline cents formatting with call to `s.service.FormatCents()`.

## Phase 5: TUI — app.go changes

```go
const (
    // ... existing
    InputTypeQuery InputType = "query"
)

type Model struct {
    // ... existing
    queryInput  QueryModel
    width       int
    height      int
}

func (m Model) EnterQueryInput() (Model, tea.Cmd)
```

Add `tea.WindowSizeMsg` handler:
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    // propagate to queryInput if active
```

Wire `Update`/`View` for `InputTypeQuery`.

In `Update()` for `InputTypeQuery`, after calling `queryInput.Update(msg)`:

```go
case InputTypeQuery:
    # handle esc from query menu → app menu
    if m.queryInput.State == QueryStateMenu {
        if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
            m.queryInput = NewQueryModel(m.service)  // reset
            return m.EnterListInput()
        }
    }

    # handle record edit request from results
    if m.queryInput.selectedID != 0 {
        id := m.queryInput.selectedID
        m.queryInput.selectedID = 0  // clear
        full, err := m.service.GetRecordFull(ctx.Background(), id)
        if err != nil {
            m.queryInput.errorMsg = fmt.Sprintf("Failed to load record: %v", err)
            return m, nil
        }
        m.inputType = InputTypeRecord
        m.recordInput = NewEditRecordModel(m.service, full, RecordOriginQuery)
        return m, nil
    }

    var queryCmd tea.Cmd
    m.queryInput, queryCmd = m.queryInput.Update(msg)
    return m, queryCmd
```

Wire `RecordModel` success with `RecordOriginQuery`:
- In `app.go Update()`, modify existing `RecordStateSuccess` handler (currently always goes to list menu):

```go
case InputTypeRecord:
    if m.recordInput.State == RecordStateSuccess {
        if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
            if m.recordInput.Origin == RecordOriginQuery {
                m.recordInput = NewRecordModel(m.service)
                m.inputType = InputTypeQuery
                m.queryInput.RefreshResults()
                return m, nil
            }
            // existing: back to list menu
            m.recordInput = NewRecordModel(m.service)
            return m.EnterListInput()
        }
    }
    return m.UpdateRecordInput(msg)
```

## Phase 6: TUI — list.go changes

- Remove "test" menu item
- Wire "query" → `m.EnterQueryInput()`

## Phase 7: TUI — QueryModel

### `internal/tui/query.go`

State machine:

```go
type QueryState string
const (
    QueryStateMenu         QueryState = "menu"
    QueryStateFilterForm   QueryState = "filter"
    QueryStateConfirm      QueryState = "confirm"
    QueryStateSavedList    QueryState = "saved"
    QueryStateSaveName     QueryState = "save_name"
    QueryStateDeleteConfirm QueryState = "delete_confirm"
    QueryStateResults      QueryState = "results"
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
    CurrencyID sql.NullInt64   // nullable
    FuzzyText  string
    TagIDs     []int64
}

func (i SavedQueryItem) FilterValue() string { return i.Name }

type QueryModel struct {
    State         QueryState
    svc           *service.Service
    err           error

    // menu
    menuInput     list.Model

    // filter form
    DateFrom      textinput.Model
    DateTo        textinput.Model
    Currency      CurrencyModel
    Tags          TagsModel
    Fuzzy         textinput.Model
    ActiveField   FilterField
    dateToManual  bool   // user manually changed DateTo

    // saved queries
    savedList     list.Model
    savedQueries  []SavedQueryItem

    // delete
    deleteTarget  SavedQueryItem

    // save name
    saveNameInput textinput.Model
    queryParams   string  // filter summary for save confirm

    // results
    results         []service.QueryResult
    cursorIndex     int
    pageSize        int
    needsRefresh    bool   // results stale, re-query needed

    // record editing (signal to app.go)
    selectedID      int64  // non-zero when user picked a record for editing

    // error state
    errorMsg        string  // displayed in results area, dismissed on any key
}
```

State transitions:

| Current | Input | Next |
|---------|-------|------|
| Menu | enter "new" | FilterForm |
| Menu | enter "saved" | SavedList |
| Menu | esc | Return to app menu |
| FilterForm | enter on last field | Confirm |
| FilterForm | tab/shift+tab | Next/prev field |
| FilterForm | esc | Menu |
| Confirm | enter | QueryRecords → Results (or ErrorMsg on failure) |
| Confirm | esc | FilterForm |
| Confirm | 1-4 | FilterForm (jump to field; 1=DateFrom, 2=Currency, 3=Tags, 4=Fuzzy) |
| SavedList | enter | Load query data → QueryRecords → Results (or ErrorMsg) |
| SavedList | d | DeleteConfirm |
| SavedList | esc | Menu |
| SavedList (empty) | — | Show "No saved queries" message |
| DeleteConfirm | y | Delete → SavedList (refreshed) |
| DeleteConfirm | n/esc | SavedList |
| SaveName | enter | Save → Results |
| SaveName | esc | Results |
| Results | enter | Set selectedID → app.go picks up and opens edit |
| Results | s | SaveName |
| Results | esc | FilterForm (pre-filled) |
| Results | j/k | Cursor up/down (absolute index, page derived) |
| Results | n/p | Page next/prev (cursor ± pageSize) |
| Results | g/G | Top/bottom |
| Results (0 results) | — | Show "No records match filters" message |
| ErrorMsg | any key | Return to filter form or saved list (context-dependent) |

Filter form fields (sequential, tab/enter to advance, shift+tab to go back):

| Index | Field | Widget | Default | Skip |
|-------|-------|--------|---------|------|
| 0 | Date From | textinput YYYY-MM-DD | Start of current month | tab |
| 1 | Date To | textinput YYYY-MM-DD | End of current month | tab |
| 2 | Currency | CurrencyModel | Any (empty = no filter) | tab through |
| 3 | Tags | TagsModel | Any (empty = no filter) | tab through |
| 4 | Fuzzy | textinput | "" | tab or enter empty |

Date auto-shift: when tab/enter leaves DateFrom, recalculate DateTo to last day of DateFrom's month (unless user manually changed DateTo — tracked by `dateToManual` flag).

Confirm step field mapping: confirm view shows 4 items (Date combined, Currency, Tags, Fuzzy).
- Press 1 → jump to FilterDateFrom (user tabs to DateTo if needed)
- Press 2 → jump to FilterCurrency
- Press 3 → jump to FilterTags
- Press 4 → jump to FilterFuzzy

Results pagination: `cursorIndex` is absolute (0 to total-1). Page is derived:
- `pageSize = max(height - <header_footer_lines>, 1)`
- Current visible page: index `cursorIndex - (cursorIndex % pageSize)` to that + pageSize
- `n`: `cursorIndex += pageSize` (clamped)
- `p`: `cursorIndex -= pageSize` (clamped)
- `g`: `cursorIndex = 0`
- `G`: `cursorIndex = len(results) - 1`

Key functions:
- `NewQueryModel(svc) QueryModel`
- `(m *QueryModel) Reset()` — factory reset to menu
- `(m *QueryModel) RefreshResults()` — re-execute current query params, called by app.go after edit save
- `(m *QueryModel) Update(msg) QueryModel`
- `(m *QueryModel) View() string`

### Filter form confirm view

```
Query Parameters:
  1) Date: 2026-05-01 → 2026-05-31
  2) Currency: USD
  3) Tags: food, drink
  4) Fuzzy: coffee

Enter: execute query
Esc: back to edit
Press number to edit field
```

### Save name view

```
Name: [________]

Filter:
  Date: 2026-05-01 → 2026-05-31
  Currency: USD
  Tags: food, drink
  Fuzzy: coffee
```

### Empty saved queries list

```
No saved queries yet.
```

### Delete confirm view

```
Delete saved query "monthly report"? (y/n)
```

## Phase 8: TUI — Results Table

### `internal/tui/query_results.go`

Custom render function (part of QueryModel, or QueryModel has a method):

```
 1) 2026-05-09  42.00 USD  groceries           [food, essential]
 2) 2026-05-08  12.50 USD  coffee              [food, drink]
 3) 2026-04-30  200.00 USD monthly rent        [housing]
```

- Columns: `cursor #  Date       Amount  Curr  Notes             Tags`
- Tags already selected in filter hidden from tag column
- Lines truncated to terminal width
- Cursor `>` on current row
- Page size from `pageSize` (based on terminal height)
- No lipgloss except truncation

Empty state (no results):
```
No records match the current filters.
Press esc to go back to filter form.
```

Error state (query failed):
```
Error: <error message>
Press any key to go back.
```

Pagination:
- `cursorIndex` is absolute (0 to total-1)
- Visible page derived from `cursorIndex - (cursorIndex % pageSize)`
- `n` = next page (cursor + pageSize), `p` = prev page (cursor - pageSize)
- `g` = first page, `G` = last page

## Phase 9: TUI — RecordModel edit support

### `internal/tui/record.go` changes

```go
type RecordOrigin string
const (
    RecordOriginCreate RecordOrigin = "create"
    RecordOriginQuery  RecordOrigin = "query"
)

type RecordModel struct {
    // ... existing fields
    Origin       RecordOrigin
    EditRecordID int64
}
```

New constructor:
```go
func NewEditRecordModel(svc, full *service.RecordFull, origin RecordOrigin) RecordModel
```

Pre-fills:
- Date: split `full.Record.Date` (YYYY-MM-DD) into year/month/day inputs
- Currency: load all currencies, find by `full.Record.CurrencyID`, set Selected
- Tags: load all tags, find matching IDs from `full.Tags`, set SelectedTags
- Amount: `FormatCents(full.Record.AmountCents)` into AmountInput
- Links: set SelectedParents from `full.Parents` (already have structured data with dates, amounts, notes, tags)
- Notes: set NotesInput from `full.Record.Notes`

Confirm behavior:
- If `Origin == RecordOriginQuery` → call `UpdateRecordWithTagsAndLinks`
- If `Origin == RecordOriginCreate` → call `CreateRecordWithTagsAndLinks` (existing)

Success view: different message for edit vs create.
- Edit: "Record updated successfully" + "Press esc to return to results"
- Create: "Record created successfully" + "Press enter to add another, esc to return"

On esc from success:
- OriginQuery → signal app to switch to `InputTypeQuery` + `RefreshResults()`
- OriginCreate → existing behavior (new record prompt)

When entering edit from results, the flow is:
1. QueryModel sets `selectedID = record.ID` in its Update method
2. app.go detects `selectedID != 0`, calls `GetRecordFull`, creates `NewEditRecordModel`, sets `InputTypeRecord`
3. User edits record, confirms, sees success message
4. User presses esc → app.go checks `Origin == RecordOriginQuery` → switches back to `InputTypeQuery` → calls `queryInput.RefreshResults()`

## Implementation Order

1. DB schema (append migration)
2. sqlc queries (3 new files + 1 modify)
3. `sqlc generate`
4. `rm vimance.db`
5. `go test ./...` — verify existing tests still pass
6. Service: saved_queries.go + saved_queries_test.go
7. Service: query_records.go + query_records_test.go
8. Service: records.go additions (GetRecordFull, UpdateRecordWithTagsAndLinks) + tests
9. `go test ./...` — verify all new service tests pass
10. Service: FormatCents utility + test (used by both service and TUI)
11. TUI: links.go (use FormatCents via service)
12. TUI: app.go (InputTypeQuery, window size, wiring)
13. TUI: list.go (remove test, add query)
14. TUI: record.go (edit mode, pre-fill, origin tracking)
15. TUI: query.go (main state machine)
16. TUI: query_results.go (or as part of query.go)
17. `go build`
18. `go test ./...`
19. User runs app and tests manually
