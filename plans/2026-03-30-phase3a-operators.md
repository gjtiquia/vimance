# Phase 3a — Linewise operators, register, paste, clipboard

**Status:** implemented.

## Summary

- **Operators `d` / `y` / `c`**: First key sets operator-pending; second **same** key runs **linewise doubled** (`dd`, `yy`, `cc`) with optional leading count (`3dd`, etc.).
- **Register** (`internal/engine/register.go`): Unnamed register holds `[][]string` + `Linewise`; used by delete/yank/change and by `p` / `P` / `x`.
- **Grid mutations**: `dd` removes rows (grid shrinks). **Row 0 (header) is never deleted**; `dd` on row 0 is a no-op.
- **`yy`**: Copies rows to register; **does not** fire `OnBufferChanged`; fires `OnClipboardWrite` with space-delimited rows (configurable via engine `clipboardDelimiter`).
- **`dd` / `cc`**: Also emit `OnClipboardWrite` after updating the register.
- **`cc`**: Clears cell text in the affected rows (not row 0 alone — `cc` on header is no-op); enters insert with highlight; fires `OnBufferChanged`.
- **`p` / `P`**: Linewise paste inserts full rows; **`P` on row 0** inserts at row 1 (never above header). Character-wise paste (`p`/`P` after `x`) replaces the current cell.
- **`x`**: Clears current cell; stores old value in register (non-linewise).
- **Events**: `EventListener` adds `OnBufferChanged()` and `OnClipboardWrite(text string)`. WASM maps them to `engine.OnBufferChanged` and `engine.OnClipboardWrite`. TS: `OnBufferChanged` → `hydrateTableFromEngine()`; `OnClipboardWrite` → `navigator.clipboard.writeText`.

## Files

| Area | Files |
|------|--------|
| Engine | `register.go`, `operator.go`, `engine.go`, `keyhandler.go`, `command.go` |
| Tests | `engine_test.go`, `keyhandler_test.go` (invalid-prefix test uses `gz` not `gx` because `x` is a command) |
| WASM | `web/wasm/main.go` |
| UI | `web/src/table.ts` |

## Rollback

Git revert this change set. Restore `EventListener` to two methods only if reverting; restore `keyhandler`/`command`/`engine` accordingly.

## Insert-mode navigation (added post-3a)

- **`Tab` in insert mode**: Exits insert → moves cursor right → re-enters insert with highlight. Effectively Esc-l-Enter as an atomic operation. At last column, cursor stays (still re-enters insert).
- **`Enter` in insert mode**: Exits insert → moves cursor down → re-enters insert with highlight. At last row, cursor stays.
- Both set `captured = true` so `preventDefault` fires (no browser tab-focus or newline).
- **Shift+Enter** (newline): Deferred — cells are `<input>` (single-line), so no newline behavior exists yet. When multiline cells are needed, pass modifiers in the keydown RPC.
- Event chain per keypress: `OnModeChanged("n")` (TS saves cell value) → `OnCursorMoved` → `OnModeChanged("i", "highlight")` (TS creates input on new cell). Cell value is saved before cursor moves.
- Tests: `TestTabInInsertModeMovesRight`, `TestTabAtLastColumnStaysInInsert`, `TestEnterInInsertModeMovesDown`, `TestEnterAtLastRowStaysInInsert`.

## Next

- **Phase 3b**: Done — see [2026-03-30-phase3b-operator-motion.md](./2026-03-30-phase3b-operator-motion.md).
- **Phase 3c**: Undo/redo (linear stack first, undo tree later).
- **Phases 4–6**: Text objects, keymap, visual mode (see main architecture doc).
