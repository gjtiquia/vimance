# Plan 006: Aggregation + Targets (TDD)

## Philosophy

Aggregation and targets are the missing read-layer. Records + tags can represent everything; we just need SUM/GROUP BY on top of them. Build bottom-up: service tests first, then implementation, then TUI.

## Active Primitives (revised)

| Primitive | Status | Purpose |
|-----------|--------|---------|
| Record    | ✅     | Data point (transaction, snapshot) |
| Tag       | ✅     | Horizontal grouping |
| Saved query | ✅  | "Which records?" (filter combo) |
| Link      | ✅ Dormant | Directional relationship, hidden from UI |
| Aggregation | ❌ NEXT | Computed view of a query (SUM, GROUP BY) |
| Target    | ❌ After aggregation | "This query's aggregate should be X" |

## Implementation Phases

### Phase 1: Aggregation Service (TDD)

**Tests first.** New files: `internal/service/aggregation.go`, `internal/service/aggregation_test.go`

#### Types

```go
type AggregationResult struct {
    TotalAmount int64
    IncomeSum   int64    // sum of positive amounts
    ExpenseSum  int64    // sum of negative amounts (stored negative)
    RecordCount int
    HasData     bool     // false when no records matched
    ByTag       []TagSum
}

type TagSum struct {
    TagName string
    Amount  int64
    Count   int
}

type PeriodAggregationResult struct {
    TotalAmount int64
    IncomeSum   int64
    ExpenseSum  int64
    RecordCount int
    HasData     bool
    Periods     []PeriodSum
}

type PeriodSum struct {
    Period    string    // "2026-01", "2026-W05", "2026", "2026-01-15"
    Amount    int64
    IncomeSum int64
    ExpenseSum int64
    Count     int
}

type PeriodGrouping string

const (
    PeriodByDay    PeriodGrouping = "day"
    PeriodByWeek   PeriodGrouping = "week"
    PeriodByMonth  PeriodGrouping = "month"
    PeriodByYear   PeriodGrouping = "year"
)
```

#### Method signatures

```go
func (s *Service) Aggregate(
    ctx context.Context,
    dateFrom, dateTo string,
    currencyID *int64,
    tagIDs []int64,
    fuzzy string,
) (*AggregationResult, error)

func (s *Service) AggregateByPeriod(
    ctx context.Context,
    dateFrom, dateTo string,
    currencyID *int64,
    tagIDs []int64,
    fuzzy string,
    grouping PeriodGrouping,
) (*PeriodAggregationResult, error)
```

Both call `s.QueryRecords()` internally, then aggregate in Go. No new SQL queries. Personal finance datasets are small. Optimize later if needed.

#### Nil vs Zero

- `HasData = false` → no records matched. Target shows `???`.
- `HasData = true` and `TotalAmount = 0` → records matched but sum to $0.

#### Multi-tag semantics

Record tagged `#food` + `#dining` contributes to BOTH tag sums. Tag sums can exceed total. Untagged records appear under `(untagged)`.

#### Period grouping

- day: "2026-01-15" (full date)
- week: "2026-W02" (ISO week, computed)
- month: "2026-01" (first 7 chars)
- year: "2026" (first 4 chars)

Sorted chronologically.

#### Integration tests (in order)

1. **TestAggregateTotals** — 5 records (mix income/expense), assert TotalAmount, IncomeSum, ExpenseSum, RecordCount
2. **TestAggregateByTag** — records with tags, assert ByTag breakdown
3. **TestAggregateByTag_Untagged** — records without tags appear under "(untagged)"
4. **TestAggregateByTag_MultiTag** — record with 2 tags contributes to both
5. **TestAggregateByTag_SameAmountDifferentSigns** — positive and negative in same tag
6. **TestAggregateEmptyResult** — no match → HasData=false, zero values, empty ByTag
7. **TestAggregateWithFilters** — date range, currency, tags, fuzzy work with aggregation
8. **TestAggregateSingleRecord** — one record gives valid aggregation
9. **TestAggregateByPeriod_Monthly** — 3 months, group by month
10. **TestAggregateByPeriod_Weekly** — group by ISO week
11. **TestAggregateByPeriod_Yearly** — group by year
12. **TestAggregateByPeriod_Daily** — group by day
13. **TestAggregateByPeriod_Empty** — no records → HasData=false

---

### Phase 2: TUI Query Enhancement — Aggregation in Results

**Modify:** `internal/tui/query.go`

#### V1 approach (simple)

Always show 2-line aggregation summary above record list:
```
Total: -$1,245.00 | Income: +$3,500.00 | Expense: -$4,745.00 | 5 records
food -645.00 (2) | rent -2,100.00 (1) | salary +3,500.00 (1) | transport -200.00 (1)
────────────────────────────────────────────────────────────────
1) 2026-05-13  -50.00  USD  lunch  [food]
...
```

Key `v` toggles to trend view (period grouping). First press → monthly, cycles: daily → weekly → monthly → yearly → back to list.

#### Trend view

```
Monthly Trend
───────────────────────────────────
2026-03  -$200 (4)
2026-04  -$600 (8)
2026-05  +$900 (6)
───────────────────────────────────
v: cycle grouping | esc: back to list
```

#### Key `T` (shift+t) creates target from current query

If query is already saved, prompt for name + amount. If not saved, save first then create target.

#### Implementation

- `QueryModel` gains: `Aggregation *service.AggregationResult`, `PeriodAgg *service.PeriodAggregationResult`, `ViewMode` (list/trend), `PeriodGrouping`
- After executing query, also call `Aggregate()` and store result
- `viewResults()` checks ViewMode, renders accordingly
- Key `v` toggles view mode; first press triggers `AggregateByPeriod(monthly)` and caches it

#### TUI tests

14. **TestQueryResultsShowAggregation** — query returns results, view contains total/income/expense
15. **TestQueryResultsTrendView** — press `v`, view contains period breakdown

---

### Phase 3: Targets Schema + Service (TDD)

#### New migration: `db/migrations/20260513000000_targets.sql`

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE targets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    saved_query_id INTEGER NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    target_cents INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_by INTEGER NOT NULL REFERENCES users(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS targets;

-- +goose StatementEnd
```

#### New sqlc queries: `db/queries/targets.sql`

```sql
-- name: CreateTarget :one
INSERT INTO targets (name, saved_query_id, target_cents, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTarget :one
SELECT * FROM targets WHERE id = ?;

-- name: ListTargets :many
SELECT * FROM targets ORDER BY created_at ASC;

-- name: UpdateTarget :one
UPDATE targets
SET name = ?, saved_query_id = ?, target_cents = ?, updated_at = ?, updated_by = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTarget :exec
DELETE FROM targets WHERE id = ?;
```

Then `sqlc generate`.

#### Service: `internal/service/targets.go`

```go
type TargetWithActual struct {
    Target       db.Target
    SavedQuery   SavedQueryItem
    ActualAmount *int64  // nil when HasData=false
    HasData      bool
}

func (s *Service) CreateTarget(ctx, name, savedQueryID, targetCents, createdBy) (db.Target, error)
func (s *Service) GetTarget(ctx, id) (db.Target, error)
func (s *Service) ListTargets(ctx) ([]db.Target, error)
func (s *Service) UpdateTarget(ctx, id, name, savedQueryID, targetCents, updatedBy) (db.Target, error)
func (s *Service) DeleteTarget(ctx, id) error
func (s *Service) GetTargetWithActual(ctx, id) (*TargetWithActual, error)
func (s *Service) ListTargetsWithActuals(ctx) ([]TargetWithActual, error)
```

`GetTargetWithActual`: load target → load saved query → build filters → call Aggregate() → return target + actual.

`ListTargetsWithActuals`: same but for all targets.

#### Target tests: `internal/service/targets_test.go`

16. **TestTargetCRUD** — create, get, list, update, delete
17. **TestTargetWithActual_HasData** — target + matching records → actual amount populated
18. **TestTargetWithActual_NoData** — target + no matching records → actual is nil, HasData=false
19. **TestTargetWithActual_TotalZero** — records match but sum to 0 → HasData=true, ActualAmount points to 0
20. **TestTargetCascadeDelete** — delete saved query → target cascade-deleted
21. **TestListTargetsWithActuals** — multiple targets with different actuals

---

### Phase 4: TUI Targets + Home Menu

#### Modify `internal/tui/app.go`

Add `InputTypeTargets`. Menu becomes:
```
commands:
  1) create   - create a new record
  2) query    - query existing records
  3) targets  - view targets vs actuals
```

#### New file: `internal/tui/targets.go`

`TargetsModel` with states:
- `TargetsStateList` — show all targets with actual vs planned
- `TargetsStateCreate` — create target (pick saved query, enter name + amount)
- `TargetsStateDeleteConfirm` — confirm deletion

#### Target list view

```
Targets
─────────────────────────────────────────────
  1) food budget      -$320 of -$500  ✓
  2) transport         -$80 of -$200  ✓
  3) rent             ??? of -$1,500  ⚠ no data
  4) savings        +$3,000 of +$10,000
─────────────────────────────────────────────
  j/k: move | enter: view query | a: add | d: delete | esc: back
```

- `enter` on target → navigate to query results with that target's saved query filters
- `a` → create target (pick from saved queries, enter name + amount)
- `d` → confirm deletion

#### Create target flow

1. List saved queries, pick one (reuse `FilteredListModel`)
2. Enter target name
3. Enter target amount (uses ParseAmountToCents)
4. Save → back to list

---

### Phase 5: Links Cleanup (verification)

Links are already dormant (not in `fieldOrder`, never used in TUI). Verify no active references. Keep `links.go` file, keep service methods, keep DB schema. No migration needed.

---

## File changes summary

### New files
- `internal/service/aggregation.go`
- `internal/service/aggregation_test.go`
- `internal/service/targets.go`
- `internal/service/targets_test.go`
- `db/migrations/20260513000000_targets.sql`
- `db/queries/targets.sql`
- `internal/tui/targets.go`

### Modified files
- `internal/tui/query.go` — ViewMode, aggregation display, trend toggle, target creation
- `internal/tui/app.go` — InputTypeTargets, menu item, routing
- `internal/db/*.go` — regenerated by sqlc
- `README.md` — update todos

### Generated by sqlc
- `internal/db/targets.sql.go`
- `internal/db/models.go` — add Target struct

---

## Test execution order

Within each phase: write test → run (fail) → implement → run (pass) → refactor → next test.

1. Phase 1: tests 1-13 (aggregation service)
2. Phase 2: tests 14-15 (TUI aggregation display)
3. Phase 3: tests 16-21 (targets service)
4. Phase 4: TUI targets (integration tests)
5. Phase 5: verify links dormant

---

## Open questions (decide during implementation)

- **Multi-currency aggregation**: V1 just sums cents regardless of currency. Add warning in UI if mixed currencies. Future: per-currency aggregation.
- **Dashboard as home screen**: V1 keeps menu as home. V2 can make targets view the landing screen.
- **Record soft-delete in TUI**: Not in scope. Service layer supports it.
- **`amount_cents` naming**: Not renaming in this phase. Could become `amount` in a future refactor.