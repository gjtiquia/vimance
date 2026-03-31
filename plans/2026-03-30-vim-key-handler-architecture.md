# Vim key handler architecture

**Status:** Phases 0–2 done; **Phase 3a** — [2026-03-30-phase3a-operators.md](./2026-03-30-phase3a-operators.md). **Engine buffer:** [2026-03-30-engine-buffer-datasource.md](./2026-03-30-engine-buffer-datasource.md). **Phase 3b:** [2026-03-30-phase3b-operator-motion.md](./2026-03-30-phase3b-operator-motion.md). **Phase 3c** (undo/redo): [2026-03-31-phase3c-undo-redo.md](./2026-03-31-phase3c-undo-redo.md). **Not started:** text objects, keymap, full visual integration.

## Problem

Normal-mode input was a flat `switch` on single keys in `Engine.KeyPress`, which cannot support multi-key sequences (`gg`), composable operators (`d$`), counts (`3j`), or user keymaps. The browser layer duplicated vim knowledge (`ClientMode` mirror + hardcoded `preventDefault` list).

## Decision

1. **Grammar-based key handling in Go** — Motions, simple commands, and (later) operators/text-objects are registered in tries. A `KeyHandler` parses input with a pending buffer for multi-key sequences.

2. **Sync JS→Go RPC for keydown** — `preventDefault` must run synchronously. Async RPC deadlocks when Go calls back into JS during the same keydown stack. **Fix:** buffer `OnModeChanged` / `OnCursorMoved` during the handler and return them in the JSON-RPC result; TypeScript dispatches `CustomEvent`s from that payload. Pointer RPCs (`setCursor`, `setCursorAndEdit`) drain the same buffer and flush via `goEngineEventsSync` (sync JS entry point).

3. **Capture flag** — After each `KeyPress`, `Engine.LastKeyCaptured()` is true iff the key was consumed (executed command/motion, valid incomplete prefix like `g` before `gg`, or Escape in insert/visual). Ignored keys do not prevent default.

4. **Rollback** — Revert by git to the commit before this plan. Public engine API (`KeyPress`, `SetCursor`, `SetCursorAndEdit`) is unchanged.

## Grammar (target)

```
[count] operator [count] motion | text-object
```

**Counts (Phase 2):** Leading digits (`3j`, `5G`, `5gg`, `10l`, …). Lone `0` is still the “first column” motion; other digits form a count until the next command. `G` without a count goes to the last row; with a count, to that 1-based line. **Not yet:** operators, text objects, keymaps.

## Sync RPC rationale

See `.cursor/plans` overview: Go must not `AwaitGlobalPromise` during a synchronous keydown from JS. Events are returned in the response instead.

## Files

| Area | Files |
|------|--------|
| Engine | `internal/engine/trie.go`, `keybuffer.go`, `motion.go`, `command.go`, `keyhandler.go`, `engine.go` |
| WASM | `web/wasm/main.go` |
| TS | `web/src/wasm/rpc.ts`, `web/src/engine/input.ts` |

## Next phases

- **Phase 2:** Done — leading digit counts (`3j`, `5G`, `5gg`, etc.).
- **Phase 3a:** Done — linewise doubled operators, register, paste, clipboard — [2026-03-30-phase3a-operators.md](./2026-03-30-phase3a-operators.md).
- **Phase 3b:** Done — operator + motion — [2026-03-30-phase3b-operator-motion.md](./2026-03-30-phase3b-operator-motion.md).
- **Phase 3c:** Done — undo/redo (linear stack) — [2026-03-31-phase3c-undo-redo.md](./2026-03-31-phase3c-undo-redo.md). Undo tree TBD.
- **Phase 4:** Text objects (`iw`, …).
- **Phase 5:** `Keymap` / `:nmap`-style remapping.
- **Phase 6:** Visual mode motions as selection extend/shrink.
