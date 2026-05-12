# Gap Analysis & Primitive Design

## Summary of conversation

We evaluated vimance from a "user testing" perspective: assuming the product is "complete," can it answer the high-level questions in README.md? The answer was **no** — while data entry works, the query/read side is too weak to answer any of them.

## High-level questions (from README)

1. How long to save up to X amount of money?
2. How much can I save per month?
3. How much can I invest per month, without being broke, while still being able to travel?
4. How much to allocate/envelope into different categories each month?
5. Can I evaluate daily if my spending is within my envelope amount?
6. Can I evaluate the performance of my investments?
7. Can I see the overall trend of my finances? Is it growing?

## Gap analysis per question

| Question | Current capability | Gap |
|----------|------------------|-----|
| 1 (save up to X) | None. No aggregation. | Need monthly net calculation, projection |
| 2 (save per month) | Can query individual records | No sum/aggregation |
| 3 (invest vs travel vs broke) | Can tag records | No sum by tag, no comparison |
| 4 (envelope budgeting) | Tags exist but no budget amounts | No planned vs actual per tag |
| 5 (daily evaluation) | None | No running totals, no daily view |
| 6 (investment performance) | Can create investment records | No ROI, no time-series |
| 7 (overall trend) | None | No monthly aggregation, no trend view |

**Critical insight**: Every question requires aggregation. Zero aggregate queries exist. The data model is fine — the read layer is missing.

## Is the generic record+tag model sufficient?

**Yes.** The problem isn't the data model. It's that we have a spreadsheet with only FILTER and no SUM/GROUP BY. Records + tags + amount_cents (signed) already can represent everything needed. We just need to compute over them.

### Sign convention (no schema change)

- Positive amounts = money in (income)
- Negative amounts = money out (expense)

This is not "baking in" finance — it's just math. Enables: `SUM(amount_cents) WHERE tag='food' AND month=5` → spending on food. `SUM(amount_cents) WHERE amount_cents > 0 AND month=5` → total income.

## Why links are dormant

Links (parent→child relationships between records) were analyzed for use cases:
- **Japan trip grouping**: Better served by tags (#japan-trip). A query on tags gives you the same aggregation.
- **Budget/container**: A "parent" with budget amount → better served by a target linked to a saved query.
- **Snapshot anchor**: Could work, but convention-based (#snapshot tag) achieves same without extra complexity.
- **Transfer tracking**: The only legitimate unique use case — linking a debit to a credit says "these are the same event." But this is bottom-up, not top-down, and not needed for V1.

**Decision**: Links stay in schema (no migration needed), but are hidden from the TUI. They'll be re-activated when transfer tracking becomes a real need.

## The revised primitive set

| Primitive | Status | Purpose |
|-----------|--------|---------|
| Record | ✅ Exists | A data point (transaction, snapshot) |
| Tag | ✅ Exists | Horizontal grouping (category, account, type) |
| Saved query | ✅ Exists | "Which records?" (filter combination) |
| Link | ✅ Dormant | Directional relationship (transfers), hidden from UI |
| Aggregation | ❌ Next | Computed view of a query (SUM, COUNT, GROUP BY tag/month) |
| Target | ❌ Future | "This query's aggregate should be X" (budget, savings goal) |

## Aggregation design (V1)

Queries should return not just a list of records, but also summary data:

```
┌─ Query Results ──────────────────────────────┐
│ 5 records matched                             │
│                                               │
│  Total:      -$1,245.00                       │
│  Income:     +$3,500.00  (2 records)         │
│  Expense:    -$4,745.00  (3 records)         │
│                                               │
│  By tag:                                      │
│   food        -$645.00  (2)                  │
│   rent       -$2,100.00  (1)                 │
│   salary     +$3,500.00  (1)                 │
│   transport   -$200.00   (1)                 │
│                                               │
│  j/k navigate │ g/G top/bot │ s save │ enter  │
└───────────────────────────────────────────────┘
```

### Aggregation modes

| Mode | Shows | Answers |
|------|-------|---------|
| **List** (current) | Individual records | "What did I buy on May 3?" |
| **Summary by tag** | SUM per tag, COUNT per tag | "How much on food vs transport?" |
| **Summary by month** | SUM per month | "Is my spending growing?" |
| **Trend** | Month-over-month delta (+/−) | "Am I saving more or less?" |
| **Budget** (future) | Tag vs planned amount vs actual | "Am I within my envelope?" |

First 4 need **zero schema changes**. They're just SQL `GROUP BY` on existing data.

## Targets (the budget/envelope primitive)

Targets replace the original idea of allocations. A target is:

```sql
CREATE TABLE targets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    saved_query_id INTEGER NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    target_cents INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
```

Why targets > allocations (monthly_budget_cents on tags):
- Allocations say "tag X gets $Y/month" — rigid, monthly-only
- Targets say "the aggregate of *this arbitrary query* should be X" — generic
- The time scope comes from the query's date range, not baked in
- A target with no matching data shows `actual: ???` — progressive disclosure

Target examples:
| Name | Linked query | Amount | Meaning |
|------|-------------|--------|---------|
| Monthly net | `tags: none, date: this month` | +$500 | "I want to save $500/mo" |
| Food budget | `tags: food, date: this month` | -$500 | "Food under $500/mo" |
| Savings goal | `tags: savings, date: all time` | +$10,000 | "I want $10k saved" |

## Progressive disclosure (top-down philosophy)

The app should work at every level of data richness, revealing what's missing:

**Level 0 — No data**: Dashboard says "Add a record."

**Level 1 — Just records**: Queries show lists + aggregation (monthly sums). Already answers: "how much did I spend?" and "is my net positive or negative?"

**Level 2 — Targets**: Dashboard shows planned vs actual. Gaps like `???` reveal missing data ("you planned $200 for transport but have no records — did you forget?").

**Level 3 — Snapshots**: Tag records `#snapshot` with absolute balance values. Aggregation detects snapshots and shows: `Balance (May 1): $5,000 ← LAST(snapshot)`. Activity since: -$1,200 ← SUM(transactions after snapshot). Implied balance: $3,800.

## Implementation priority

1. **Aggregation in queries** (no schema change) — SUM by tag, income vs expense. Unlocks answering most questions immediately.
2. **Dashboard/home screen** — Show current month summary + targets vs actual.
3. **Targets table** — Thin new entity. Enables budget tracking.
4. **Trend view** — Monthly table over time. Just aggregation grouped differently.

## Implementation note: Links dormant

When implementing this plan, links were hidden from the TUI:
- `FieldLinks` removed from record form field order
- `LinksModel` removed from record model
- `LinkPickerQuery` / picker mode removed from query model
- `RecordStateLinkPicker` removed from confirm states
- Links field removed from confirm screen
- Service layer (`CreateRecordWithTagsAndLinks`, etc.) unchanged — still accepts `nil` for parentIDs
- Database schema unchanged — record_links table stays
- All link-related TUI tests removed or updated