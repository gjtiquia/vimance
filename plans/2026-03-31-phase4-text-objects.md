# Phase 4 — Text objects (`iw` / `aw`)

**Status:** implemented.

## Summary

- **`iw` and `aw`** both resolve to the **current cell** (no intra-cell word model in the grid).
- Operators compose as **`ciw`**, **`diw`**, **`yiw`**, **`caw`**, **`daw`**, **`yaw`** — same semantics for `i` vs `a` pairs.
- **`TextObjectRegistry`** ([`internal/engine/textobject.go`](../internal/engine/textobject.go)) uses the shared [`Trie`](internal/engine/trie.go) with sequences `["i","w"]` and `["a","w"]`.
- **Key handler** ([`internal/engine/keyhandler.go`](../internal/engine/keyhandler.go)): when an operator is pending, a lone **`i` or `a`** is treated as a **text object prefix** (MatchPrefix), not as the insert command. The next key completes the object (`w`). **`Escape`** clears operator + text-object buffer.
- **`ExecuteOperatorWithTextObject`** ([`internal/engine/engine.go`](../internal/engine/engine.go)) delegates to `deleteCellsInRowRange` / `yankCellsInRowRange` / `changeCellsInRowRange` and moves the cursor after `d`/`c` like non-linewise operator+motion.

## Files

| Area | Files |
|------|--------|
| Engine | `textobject.go` (new), `keyhandler.go`, `engine.go` |
| Tests | `engine_test.go` |

## Rollback

Git revert this change set.

## Next

- **Phase 5:** Keymap / remapping — [2026-03-30-vim-key-handler-architecture.md](./2026-03-30-vim-key-handler-architecture.md)
- **Phase 6:** Visual mode
