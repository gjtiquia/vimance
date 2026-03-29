# Wasm build caching (planned)

**Status:** not implemented yet — execute when ready.

Skip TinyGo wasm recompilation when no Go source files have changed, using timestamp-based `find -newer` checking. Extract `build:wasm` into its own script with the caching logic.

## Problem

The current `build:assets` script runs TinyGo wasm compilation on every invocation, even when only TS/CSS files changed. TinyGo is slow, so this wastes time during frontend-only iteration.

## Approach

Use `find -newer` to compare source file timestamps against `web/public/main.wasm`. If no Go source file is newer than the output, skip the build.

**Wasm source dependencies:**

- `web/wasm/main.go`
- `internal/engine/*.go`
- `internal/js/*.go`
- `internal/jsonrpc/*.go`
- `go.mod`, `go.sum`

## Changes

### 1. Create `scripts/build-wasm.sh`

A small shell script that:

1. Checks if `web/public/main.wasm` exists
2. Uses `find` on all Go source dirs + `go.mod`/`go.sum` with `-newer web/public/main.wasm`
3. If no files are newer, prints `wasm: up to date` and exits
4. Otherwise, runs the TinyGo build

```bash
#!/usr/bin/env bash
set -euo pipefail

OUT="web/public/main.wasm"
DEPS="web/wasm internal/engine internal/js internal/jsonrpc"

if [ -f "$OUT" ] && [ -z "$(find $DEPS go.mod go.sum -newer "$OUT" 2>/dev/null)" ]; then
    echo "wasm: up to date"
    exit 0
fi

GOOS=js GOARCH=wasm tinygo build -o "$OUT" ./web/wasm
```

Make the script executable: `chmod +x scripts/build-wasm.sh` (optional if invoking via `bash scripts/build-wasm.sh`).

### 2. Update `package.json`

- Add `build:wasm` calling the script
- Point `build:assets` at `bun run build:wasm` instead of the raw TinyGo command

```json
"build:assets": "bunx tsc --noEmit && bunx tailwindcss -i ./web/src/input.css -o ./web/public/styles.css && bun build ./web/src/index.ts --outdir=./web/public && templ generate && bun run build:wasm",
"build:wasm": "bash scripts/build-wasm.sh",
```

### 3. Update `.air.toml` comment (optional)

The `[build]` comment in `.air.toml` currently points at `package.json` for the asset pipeline. After implementation, you can mention that wasm is conditionally skipped via `scripts/build-wasm.sh` so readers know where the logic lives.

No change is strictly required: `.air.toml` already runs `bun run build:assets`.

## Expected behavior

- **First build or after Go changes:** TinyGo runs, `main.wasm` is produced
- **Subsequent builds with only TS/CSS changes:** `wasm: up to date`, TinyGo skipped
- **After `git checkout` that bumps Go file timestamps:** TinyGo runs (false positive, but harmless and rare)

## Verification checklist

- [ ] Run `bun run build:assets` twice; second run should print `wasm: up to date` after wasm step
- [ ] Touch a file under `internal/engine/` and run again; TinyGo should run
- [ ] README: optional note that wasm is cached by timestamps (only if you want it documented for contributors)
