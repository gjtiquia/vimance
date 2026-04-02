# Phase 5 — Keymap / remapping

**Status:** implemented.

## Summary

- **Go API** (no `:` command-line yet): `Nmap` / `Nnoremap` (normal), `Imap` / `Inoremap` (insert), `Vmap` / `Vnoremap` (visual), `Unmap(mode, lhs)`.
- **`ParseKeys`**: character runes plus `<Name>` tokens (e.g. `<Escape>`, `<Tab>`, `<Ctrl+r>`).
- **`KeymapTable`**: trie-backed LHS → `KeymapEntry` (RHS `[]string`, `Recursive`).
- **`Engine.feedKey`**: resolves keymap pending buffers, single-key lookups, then `feedKeyMode`. **`KeyPress`** resets `lastKeyCaptured` and `keymapDepth` (not `keymapPending`).
- **Recursive vs non-recursive**: RHS replay uses `feedKey(k, entry.Recursive)`. `Nnoremap` sets `Recursive` false so RHS does not re-enter keymaps.
- **Depth guard**: `maxKeymapDepth` (50); at limit, RHS keys are fed with `checkKeymap=false`.
- **Drain on `MatchNone`**: full buffered sequence is replayed with **no** keymap on any key (fixes `gg` after a `gx`-style prefix).
- **Mode / cursor**: `keymapPending` cleared when mode changes (`setMode`) or on `SetCursor` / `SetCursorAndEdit`.

## Files

| Area | Files |
|------|--------|
| Engine | `keymap.go`, `trie.go` (`Delete`), `engine.go` |
| Tests | `keymap_test.go`, `engine_test.go` |

## Rollback

Git revert this change set.

## Next

- **Phase 6:** Visual mode — [2026-03-30-vim-key-handler-architecture.md](./2026-03-30-vim-key-handler-architecture.md)
- **Future:** `:` command-line for `:nmap`, undo tree
