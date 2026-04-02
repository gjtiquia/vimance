# Phase 6: Visual mode (v, V, Ctrl+v)

## Summary

Implemented full visual mode with three sub-modes, rectangular / linewise selection, operators on selection, `o` to swap anchor, `gv` to re-select last visual, `OnSelectionChanged` for UI highlighting, and WASM/TS wiring.

## Behavior

- **Modes:** `ModeVisual` (`v`), `ModeVisualLine` (`V`), `ModeVisualBlock` (`Ctrl+v`). `v` and `Ctrl+v` both use cell-wise rectangular selection; linewise `V` selects full rows (`startX=0`, `endX=cols-1`).
- **Selection:** Anchor at enter; cursor moves with motions; `GetVisualSelection()` returns inclusive bounds.
- **Operators:** `d` / `y` / `c` / `x` (as `d`) run immediately on the selection; `c` enters insert afterward.
- **Cell-wise:** `deleteRect` / `yankRect` / `changeRect` (multi-row register, non-linewise). **`deleteRect` / `changeRect` never write to row 0** (header): mutation uses `mutStartY = max(startY, 1)`; header-only selections are a no-op for `d` and exit to normal for `c` without insert. `yankRect` may still read header cells.
- **Linewise:** Reuses `deleteRowsRange` / `yankRowsRange` / `changeRowsRange` with `startY = max(minY, 1)` for `d`/`c` (header protection).
- **gv:** Special case in `feedWithPending` when `["g","v"]` is `MatchNone` — calls `RestoreLastVisualSelection()`.
- **Events:** `EventListener.OnSelectionChanged(startX, startY, endX, endY, cursorX, cursorY)`.

## Files touched

| File | Role |
|------|------|
| `internal/engine/engine.go` | Visual state, `enterVisualMode` / `exitVisualMode`, `ExecuteVisualOperator`, rect + linewise helpers, `feedKeyMode`, `keymapTableForMode`, `SetCursor*`, `Unmap` |
| `internal/engine/visualkeyhandler.go` | Visual key parsing (toggle, `o`, counts, motions, operators) |
| `internal/engine/command.go` | `v` / `V` / `Ctrl+v` → `enterVisualMode` |
| `internal/engine/keyhandler.go` | `gv` in `feedWithPending` |
| `web/wasm/main.go` | `Ctrl+v` normalization, `OnSelectionChanged` RPC event |
| `web/src/table.ts` | `handleOnSelectionChanged`, visual cleanup on `mode === "n"` |
| `internal/engine/engine_test.go` | Mock `OnSelectionChanged`, Phase 6 tests |

## Rollback

```bash
git revert <commit>
```

Restore `EventListener` to five methods (remove `OnSelectionChanged`) and revert the files above.

## References

- Roadmap: [2026-03-30-vim-key-handler-architecture.md](./2026-03-30-vim-key-handler-architecture.md)
- Previous phase: [2026-03-31-phase5-keymap.md](./2026-03-31-phase5-keymap.md)
