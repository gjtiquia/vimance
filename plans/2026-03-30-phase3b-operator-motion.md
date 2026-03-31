# Phase 3b — Operator + motion composition

**Status:** implemented.

## Summary

- **Motions** now return `MotionResult{TargetX, TargetY}` without moving the cursor. `MotionInfo` adds `Linewise` and `Inclusive` (for horizontal non-linewise ranges).
- **Standalone motion**: after `minfo.Fn(...)`, `moveCursorTo(target)` preserves prior behavior.
- **Operator + motion**: with `d`/`y`/`c` pending, the next keys resolve a motion (including multi-key `gg`). Digits between operator and motion form **count2**; effective motion count = `pendingOpCount * motionBase` (same pattern as leading count × motion count).
- **Linewise** (`j`, `k`, `gg`, `G`, arrows vertical): range is rows `min(startY, targetY)` … `max(startY, targetY)` inclusive; **delete/change** clamp the range to start at row **≥ 1** (header never deleted).
- **Non-linewise** (`h`/`l`/`w`/`e`/`b`, `$`, `0`, horizontal arrows): column range uses **inclusive** for `$` and `0`, **exclusive** for `h`/`l`/etc. (so `dl` clears one cell like vim).
- **Escape** while operator is pending: cancels pending operator / motion buffer; returns **captured** so the key does not bubble.
- **Refactor**: `deleteRowsRange`, `yankRowsRange`, `changeRowsRange` shared by `ExecuteLinewiseDoubled` and linewise operator+motion. Non-linewise uses `deleteCellsInRowRange` / `yankCellsInRowRange` / `changeCellsInRowRange`.

## Files

| Area | Files |
|------|--------|
| Engine | `motion.go`, `keyhandler.go`, `engine.go` |
| Tests | `engine_test.go` |

## Rollback

Git revert this change set.

## Next

- **Phase 3c**: Done — [2026-03-31-phase3c-undo-redo.md](./2026-03-31-phase3c-undo-redo.md).
- **Phase 4+**: Text objects, keymap, visual mode.
