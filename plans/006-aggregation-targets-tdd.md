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

## User Journey Flows (acceptance criteria)

These are the end-to-end scenarios we identified during gap analysis. Each represents a real user question our app should answer.

### First-time / Onboarding

1. **"I just installed, now what?"** — 0 records, query returns HasData=false, empty target list. App must handle 0→1 gracefully.
2. **"I only know my bank balance"** — create 1 record (+5000 #checking #snapshot), aggregation valid with count=1.

### Basic Aggregation

3. **"How much did I spend this month?"** — mixed income/expense records, query May, assert TotalAmount, IncomeSum, ExpenseSum, RecordCount, ByTag.
4. **"Where did my money go?"** — aggregation breakdown by tag, each tag shows correct sum and count.
5. **"Is my spending growing?"** — records across months, AggregateByPeriod monthly, period sums show trend.
6. **"Show me weekly spending"** — AggregateByPeriod weekly, ISO week grouping.
7. **"Compare this month vs last month"** — wide date range, group by month, see two periods side by side.

### Aggregation Edge Cases

8. **"Single record aggregation"** — 1 record, aggregation valid (count=1, sum=amount).
9. **"All records same tag"** — 10 records all #food, ByTag has one entry with correct sum/count.
10. **"Record with multiple tags"** — 1 record -$50 tagged #food #dining, both tags show -$50, total is -$50 (no double-counting).
11. **"No tags on records"** — records with no tags appear under "(untagged)" in ByTag.
12. **"Same tag, mixed signs"** — #food has +$50 (refund) and -$200 (groceries), tag sum = -$150.
13. **"Empty query result"** — no records match, HasData=false, TotalAmount=0, empty ByTag.

### Targets — Core

14. **"I want to budget $500/month for food"** — create query (tag: food, date: May), save it, create target (amount: -$500), ListTargetsWithActuals shows food budget actual vs planned.
15. **"Track a category I haven't started"** — target for #investment but no investment records → HasData=false, ActualAmount=nil, shows "???".
16. **"Multiple targets at once"** — food, transport, rent, savings targets. Some with data, some without. ListTargetsWithActuals returns all with correct actuals.
17. **"I want to save $10,000 total"** — target with all-time date range, cumulative sum, not monthly. The query scope defines the period, not the target.
18. **"Target amount changed mid-month"** — update target amount, previous records unchanged, new actual vs planned shown.

### Targets — Progressive Disclosure

19. **"Target reveals missing data"** — target for #rent shows HasData=false, prompts "you haven't logged rent yet".
20. **"Add first record to tracked category"** — target was HasData=false, add matching record, now HasData=true, ActualAmount populated.

### Target ↔ Query Drill-Down

21. **"From target to query results"** — target → load its saved query → navigate to query results with those filters → see detailed records.
22. **"From query results create target"** — execute query → see aggregation → press T → create target from that query.

### Snapshots (Convention-Based)

23. **"Snapshot + transactions"** — record tagged #checking #snapshot +$5000 (May 1), plus May transactions tagged #checking. Query #checking for May. Aggregation includes both. (V1: no special snapshot handling, just summed. Future: detect #snapshot and show differently.)

### Multi-Currency

24. **"Mixed currencies in aggregation"** — $50 USD + 2000 PHP in same query. V1: just sums cents. Future: warn or group by currency.

---

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

#### Design decisions

- **Nil vs Zero**: `HasData = false` → no records matched. Target shows `???`. `HasData = true` and `TotalAmount = 0` → records matched but sum to $0.
- **Multi-tag**: record tagged #food + #dining contributes to BOTH tag sums. Tag sums can exceed total. `(untagged)` for records with no tags.
- **Period grouping**: day = full date, week = ISO week "2026-W02", month = "2026-01", year = "2026". Sorted chronologically.
- **Mixed currencies**: V1 just sums cents. Future: warn or group by currency.

#### Service unit tests (TDD order)

1. **TestAggregate_EmptyResult** — no records match, HasData=false, zero values, empty ByTag (flows 1, 13)
2. **TestAggregate_SingleRecord** — one record, valid aggregation with count=1 (flow 8)
3. **TestAggregate_Totals_IncomeExpense** — 5 records mix income/expense, assert TotalAmount, IncomeSum, ExpenseSum, RecordCount (flows 1→3)
4. **TestAggregate_ByTag_Basic** — records with tags, correct breakdown (flow 4)
5. **TestAggregate_ByTag_SameTag** — all records same tag, one entry (flow 9)
6. **TestAggregate_ByTag_MultiTag** — record with 2 tags, both get amount, total not double-counted (flow 10)
7. **TestAggregate_Untagged** — records without tags appear under "(untagged)" (flow 11)
8. **TestAggregate_ByTag_MixedSigns** — positive and negative in same tag (flow 12)
9. **TestAggregate_WithFilters** — date range, currency, tags, fuzzy all work (flow variant)
10. **TestAggregateByPeriod_Monthly** — 3 months, group by month (flows 5, 7)
11. **TestAggregateByPeriod_Weekly** — ISO week grouping (flow 6)
12. **TestAggregateByPeriod_Yearly** — group by year
13. **TestAggregateByPeriod_Daily** — group by day
14. **TestAggregateByPeriod_Empty** — no records → HasData=false (flow 1 variant)

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

Key `v` toggles to trend view (period grouping). Cycles: daily → weekly → monthly → yearly → back to list.

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

#### Additional keys in results view

- `v` — toggle between list view and trend view (cycles grouping period)
- `T` (shift+t) — create target from current query (save query if not saved, then prompt name + amount)

#### Implementation

- `QueryModel` gains: `Aggregation *service.AggregationResult`, `PeriodAgg *service.PeriodAggregationResult`, `ViewMode` (list/trend), `PeriodGrouping`
- After executing query, also call `Aggregate()` and store result
- `viewResults()` checks ViewMode, renders accordingly
- Key `v` toggles view mode; first press triggers `AggregateByPeriod(monthly)` and caches it

#### TUI tests

15. **TestQueryResultsShowAggregation** — query returns results, view contains total/income/expense
16. **TestQueryResultsTrendView** — press `v`, view contains period breakdown

---

### Phase 3: Targets Schema + Service (TDD)

#### Append to existing migration: `db/migrations/20260423000000_initial_schema.sql`

Add targets table to the existing migration file (no prod DB, can modify in place):

```sql
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
```

Also add to the Down section: `DROP TABLE IF EXISTS targets;`

Add index: `CREATE INDEX idx_targets_saved_query_id ON targets(saved_query_id);`

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

`GetTargetWithActual`: load target → load saved query → build filters from saved query → call Aggregate() → return target + actual.

`ListTargetsWithActuals`: same for all targets.

Note: saved queries store tag IDs via `saved_query_tags`. When resolving a target's query for aggregation, need to load the saved query's tags and pass them as `tagIDs` to `Aggregate()`.

#### Target service tests: `internal/service/targets_test.go`

17. **TestTarget_CRUD** — create, get, list, update, delete
18. **TestTarget_WithActual_HasData** — target + matching records → actual amount populated (flow 14)
19. **TestTarget_WithActual_NoData** — target + no matching records → ActualAmount=nil, HasData=false (flow 15)
20. **TestTarget_WithActual_TotalZero** — records match but sum to 0 → HasData=true, ActualAmount points to 0
21. **TestTarget_CascadeDelete** — delete saved query → target cascade-deleted (flow variant)
22. **TestTarget_ListWithActuals** — multiple targets with different actuals, some with data, some without (flow 16)
23. **TestTarget_CumulativeScope** — target with all-time date range sums all matching records (flow 17)
24. **TestTarget_UpdateAmount** — update target amount, verify new amount (flow 18)

---

### Phase 4: Journey / Integration Tests

These test end-to-end user workflows at the service level, combining aggregation + targets + queries. They're the ultimate acceptance criteria.

New file: `internal/service/journey_test.go`

25. **TestJourney_FirstRecord** — empty DB → query returns HasData=false → create record → query returns HasData=true with valid aggregation (flow 1→2)
26. **TestJourney_MonthlySpending** — create May records → query May → assert aggregation matches expected totals by tag (flow 3)
27. **TestJourney_BudgetTracking** — create food records → save query → create target → ListTargetsWithActuals shows actual vs planned → add more food → actual updates (flow 14 full)
28. **TestJourney_MonthlyTrend** — records across 3 months → AggregateByPeriod(monthly) → assert each month's sum (flow 5)
29. **TestJourney_ProgressiveDisclosure** — create target for empty category → HasData=false → add record → HasData=true (flows 15, 19, 20)
30. **TestJourney_MultiTag** — record with 2 tags → aggregation shows amount in both tags, total not double-counted (flow 10)
31. **TestJourney_MixedCurrencies** — records in USD + PHP → aggregation sums cents regardless of currency (flow 24, V1 behavior)
32. **TestJourney_TargetFromQuery** — execute query → save query → create target → GetTargetWithActual matches aggregation results (flow 22)
33. **TestJourney_SnapshotConvention** — record tagged #snapshot + other records → aggregation includes both (flow 23, V1: no special handling)

---

### Phase 5: TUI Targets + Home Menu

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

#### TUI tests

34. **TestTUI_TargetList** — targets shown with actual vs planned
35. **TestTUI_TargetCreation** — create target from saved query
36. **TestTUI_TargetDrillDown** — enter on target navigates to query results (flow 21)

---

### Phase 6: Links Cleanup (verification)

Links are already dormant (not in `fieldOrder`, never used in TUI). Verify no active references. Keep `links.go` file, keep service methods, keep DB schema. No migration needed. Likely a no-op.

---

## File changes summary

### New files
- `internal/service/aggregation.go` — AggregationResult, TagSum, PeriodSum types + Aggregate(), AggregateByPeriod()
- `internal/service/aggregation_test.go` — tests 1-14
- `internal/service/targets.go` — TargetWithActual, CRUD + GetTargetWithActual, ListTargetsWithActuals
- `internal/service/targets_test.go` — tests 17-24
- `internal/service/journey_test.go` — tests 25-33
- `db/queries/targets.sql` — sqlc queries for targets
- `internal/tui/targets.go` — TargetsModel

### Modified files
- `db/migrations/20260423000000_initial_schema.sql` — append targets table
- `internal/tui/query.go` — ViewMode, aggregation display, trend toggle, target creation shortcut
- `internal/tui/app.go` — InputTypeTargets, menu item, routing
- `internal/db/*.go` — regenerated by sqlc
- `README.md` — update todos

### Generated by sqlc
- `internal/db/targets.sql.go`
- `internal/db/models.go` — add Target struct

---

## Test execution order

TDD strictly. Write test → run (fail) → implement → run (pass) → refactor → next test.

1. Phase 1: tests 1-14 (aggregation service)
2. Phase 2: tests 15-16 (TUI aggregation display)
3. Phase 3: tests 17-24 (targets service)
4. Phase 4: tests 25-33 (journey tests — acceptance criteria)
5. Phase 5: tests 34-36 (TUI targets)
6. Phase 6: verify links dormant (no tests needed)

---

## Open questions (decide during implementation)

- **Multi-currency aggregation**: V1 just sums cents regardless of currency. Future: per-currency aggregation or conversion.
- **Dashboard as home screen**: V1 keeps menu as home. V2 can make targets view the landing screen.
- **Record soft-delete in TUI**: Not in scope. Service layer supports it.
- **`amount_cents` naming**: Not renaming in this phase. Could become `amount` in a future refactor.
- **Snapshot handling**: V1 treats snapshots same as regular records in aggregation. Future: detect #snapshot tag and display differently.