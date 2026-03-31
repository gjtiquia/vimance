# Phase 3c — Undo/redo (linear stack)

**Status:** implemented.

## Summary

- **Linear undo/redo** with snapshots of `cells`, `cursorX`, and `cursorY` (`UndoEntry` in [`internal/engine/undostack.go`](../internal/engine/undostack.go)).
- **`u`** — pops undo, pushes current state to redo, restores previous snapshot.
- **`Ctrl+r`** — pops redo, pushes current to undo, restores. Browser reload is **prevented** when the key is captured (same pattern as other normal-mode keys); users can still hard-reload via the URL bar or Ctrl+Shift+R.
- **New mutations** call `pushUndoCheckpoint()` before changing the grid, which appends a pre-change snapshot to the undo stack and **clears the redo stack**.
- **`yy` / yanks** do not push undo (no buffer mutation).
- **Insert-mode cell saves** use `SetCellValueUndoable` from WASM (`setCellValue` RPC); skips checkpoint when the value is unchanged.
- **Restore** emits `OnBufferChanged()` and `OnCursorMoved()` so the TS layer can rehydrate the table.

## Checkpoints

Undo checkpoints are pushed in:

- `deleteRowsRange`, `changeRowsRange`, `deleteCellsInRowRange`, `changeCellsInRowRange` (covers `dd`, operator+motion `d`/`c`, etc.)
- `DeleteCharUnderCursor` (`x`)
- `PasteAfter` / `PasteBefore` (when a paste actually runs)
- `SetCellValueUndoable` (when the cell value changes)

`SetCellValue` remains non-undoable for internal use after a checkpoint (e.g. `DeleteCharUnderCursor`).

## Granularity

`cc` then typing then Esc produces **two** undo steps if applicable: one for the change operator clearing rows, one for `setCellValue` — acceptable for v1 (see plan).

## Files

| Area | Files |
|------|--------|
| Engine | `undostack.go`, `engine.go`, `command.go` |
| WASM | `web/wasm/main.go` (`ctrlKey` → `Ctrl+r`, `setCellValue` → `SetCellValueUndoable`) |
| TS | `web/src/engine/input.ts` |
| Tests | `engine_test.go` |

## Rollback

Git revert this change set.

## Next

- **Phase 4:** Text objects (`iw`, …)
- **Phase 5:** Keymap / remapping
- **Phase 6:** Visual mode
- **Future:** Undo tree (branching history)
