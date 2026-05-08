# Phase 1: Record Links

## 1. Database Migration

New file: `db/migrations/<timestamp>_add_record_links.sql`

```sql
CREATE TABLE record_links (
    parent_id INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    child_id  INTEGER NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (parent_id, child_id)
);
```

- ON DELETE CASCADE: if a record is hard-deleted, its links vanish. soft-deleted records keep links (they still exist in the table).
- no `position` field — children ordered by date (chronological).

## 2. sqlc Queries

New file: `db/queries/record_links.sql`

- `AddRecordLink(ctx, parent_id, child_id)` — INSERT
- `RemoveRecordLink(ctx, parent_id, child_id)` — DELETE
- `GetRecordParents(ctx, child_id)` — SELECT parent_id, returns `[]RecordLink`
- `GetRecordChildren(ctx, parent_id)` — SELECT child_id, returns `[]RecordLink`
- `SearchLinkCandidates(ctx, {date_from, date_to, currency_id, query})` — full-text fuzzy search across notes + tag names, filtered by date range and currency, returns records.

## 3. Service Layer

New file: `internal/service/record_links.go`

- `LinkRecords(ctx, parentID, childID)`
- `UnlinkRecords(ctx, parentID, childID)`
- `GetRecordParents(ctx, recordID)`
- `GetRecordChildren(ctx, recordID)`
- `SearchLinkCandidates(ctx, dateFrom, dateTo, currencyID, query)` — wraps the sqlc query

Update: `records.go`
- Modify `CreateRecordWithTags` → `CreateRecordWithTagsAndLinks(ctx, ..., parentIDs []int64)` — inserts record, tags, AND links in a transaction.

## 4. TUI — LinksModel

New file: `internal/tui/links.go`

Pattern follows TagsModel / CurrencyModel structure (search + select + vim modes):

```
Links:
  [bank-balance-apr, credit-card-may]
  Search: _
  > 1) 2026-04-01  $5,000.00  "May bank balance"  [balance:checking]
    2) 2026-04-15  $1,200.00  "April credit card"  [expense:credit-card]
```

**Fields:**
- `SelectedParents []LinkedRecord` — chips of selected parent records
- `SearchInput textinput.Model` — type to fuzzy filter
- `AllCandidates []LinkedRecord` — loaded from service (date range + currency filter)
- `FilteredCandidates []LinkedRecord` — filtered by search input
- `CursorIndex int`
- `Mode LinkMode` — Insert (typing) / Normal (navigating)
- `service *service.Service`
- `DateFrom, DateTo string` — defaults to same month as the record's date
- `CurrencyID int64` — from the currency field

**LinkedRecord struct:**
```go
type LinkedRecord struct {
    ID          int64
    Date        string
    AmountCents int64
    CurrencyID  int64
    Notes       string
    TagNames    []string
}
```

**LoadCandidates(ctx)** — called when entering the Links field. queries `SearchLinkCandidates` using date range (derived from the date inputs on the form, same month) and currency from the currency field.

**Key behavior (same pattern as TagsModel):**
- Insert mode: type to filter, enter to select, ctrl+z to undo last selection
- Normal mode: j/k or up/down to navigate, enter to select, i/a to enter insert mode
- tab to advance to next field (notes)
- shift+tab to go back (amount)
- **No inline creation** — only pick from existing records

**Date range handling:**
- on field entry, derive `DateFrom` and `DateTo` from the record's date inputs (first to last day of that month)
- if date fields are empty, default to current month
- show a header like `[Showing records from 2026-04-01 to 2026-04-30 in USD]`

## 5. TUI — RecordModel Updates

**New field order:**
```
FieldDateYear → FieldDateMonth → FieldDateDay → FieldCurrency → FieldTags → FieldAmount → FieldLinks → FieldNotes
```

`RecordModel` gets a `LinksInput LinksModel` field.

`NewRecordModel` — initialize `LinksInput` with `NewLinksModel(svc)`.

`focusActiveField` — add `FieldLinks` case.

`setActiveField` — add `FieldLinks` case; on entering, call `LoadCandidates` with current date + currency.

`updateEditing` — add tab/shift+tab/enter transitions for the new field positions.

`viewEditing` — render `m.LinksInput.View()` between amount and notes.

## 6. Confirm Screen Update

`confirm.go` — add a "Links" line between Amount and Notes:

```
6) Links: bank-balance-apr, credit-card-may
   (or "(none)" with a warning for "no links — record is standalone")
```

## 7. Save Flow Update

`record.go` `CreateRecord(ctx, userID)`:
- after creating the record and tags, loop through `m.LinksInput.SelectedParents` and call `LinkRecords` for each parent ID
- wrap everything in the existing transaction

## 8. Not in Phase 1 (deferred to Phase 2)

- Browse/edit mode with vim-key navigation
- Editing existing records
- Removing links after creation
- Query builder UI
- Date range picker for link search (beyond default same-month)
