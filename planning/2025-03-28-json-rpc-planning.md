# JSON-RPC Bridge Plan

Date: 2025-03-28

## Goal

Prototype a tiny notification-only bridge between the frontend JS runtime and Go wasm.

The bridge should support:

- Fire-and-forget messages from JS to Go
- Fire-and-forget messages from Go back to JS
- Routing by `method`
- No promises
- No `id` for now
- A shape that can later grow correlation metadata if needed

## Design

Treat JSON-RPC as a message envelope, not a strict request/response system.

### Outbound direction: JS -> Go

JS sends a notification-style payload into Go wasm.

Example shape:

```json
{
  "jsonrpc": "2.0",
  "method": "bridge.testButtonPressed",
  "params": {
    "source": "test-button"
  }
}
```

### Inbound direction: Go -> JS

Go emits a separate notification-style payload back to JS when it wants to report something.

Example shape:

```json
{
  "jsonrpc": "2.0",
  "method": "bridge.testButtonResult",
  "params": {
    "message": "button handled"
  }
}
```

## Implementation Shape

### JS side

- Add one generic sender into Go wasm.
- Add one generic receiver from Go wasm.
- Route inbound messages by `method`.
- Keep `web/src/test-button.ts` as the first smoke test.

### Go side

- Expose one generic handler that accepts a JSON string.
- Parse the envelope.
- Dispatch by `method` internally.
- Emit outbound notifications back to JS through one generic callback.

## Routing Rules

- `method` is the primary routing key.
- No request/response pairing exists in the first version.
- If multiple events ever need correlation, add metadata later in `params`.

## Success Criteria

- Clicking the test button sends a notification into Go wasm.
- Go processes the message without needing a promise or response object.
- Go can send a follow-up notification back to JS.
- The bridge stays small and easy to extend.

## Notes

- This is intentionally notification-only.
- A real JSON-RPC response only applies when the message carries an `id`.
- We are not using that part yet.

