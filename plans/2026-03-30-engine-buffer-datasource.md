# Engine buffer and DataSource

**Status:** implemented.

## What changed

- **`DataSource` interface** (`internal/engine/datasource.go`): `Load() [][]string`, `Save(cells [][]string) error`.
- **`StubDataSource`**: hardcoded 5×6 finance-style sample grid (header row + 4 data rows). `Save` is a no-op until you add persistence.
- **`StaticDataSource`**: holds a grid in memory; `Save` replaces it — used in tests and for in-memory round-trips.
- **`Engine`** owns **`cells [][]string`**, loaded from `DataSource` in `New(ds)`. Exposes `CellValue`, `SetCellValue`, `CellsSnapshot`, `Cols`, `Rows`, `SaveBuffer` (calls `DataSource.Save`).
- **RPC (WASM):** `getGrid` → `{ cols, rows, cells, cursorX, cursorY }`; `setCellValue` → `{ x, y, value }`; `saveBuffer` → persists via `Save` (for future `:w`).
- **UI:** `TableComponentShell()` in templ — empty `<tbody data-table-tbody>`. **`hydrateTableFromEngine()`** in `web/src/table.ts` runs after **`wasm.initAsync()`** in `web/src/index.ts`, builds all rows from `getGrid`. Row `y === 0` uses header styling (bold) on `<td>`s.
- **`OnModeChanged`** WASM payload now includes **`x`**, **`y`** (cursor). Leaving insert mode calls **`setCellValue`** so the engine stays the source of truth.

## Init order

Keep **`engine.init()`** before **`await wasm.initAsync()`** (see comment in `index.ts`). **`hydrateTableFromEngine()`** runs **after** WASM so `getGrid` / `jsToGoJsonRpcSync` exist.

## Next steps

- Replace `StubDataSource` with SQLite (or HTTP) loading the same `[][]string` shape.
- Wire **`:w`** to `saveBuffer` RPC (and optional server persistence).
- Phase 3a operators (`dd`, `yy`, `cc`, `p`/`P`) mutate `cells` / clipboard; `OnBufferChanged` + `hydrateTableFromEngine` refresh the table — see [2026-03-30-phase3a-operators.md](./2026-03-30-phase3a-operators.md).

## Rollback

Git revert this change set; restore `TableComponent(...)` in `page_home.templ` and `engine.New(cols, rows)` if rolling back entirely.
