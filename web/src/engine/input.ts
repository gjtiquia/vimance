import { sendRpcAsync, sendRpcSync } from "../wasm";
import * as jsonrpc from "../jsonrpc";

function dispatchKeydownEvents(
    result: jsonrpc.Response["result"],
): void {
    if (!result || typeof result !== "object") {
        return;
    }
    const r = result as {
        events?: { method: string; params: Record<string, unknown> }[];
    };
    if (!Array.isArray(r.events)) {
        return;
    }
    for (const e of r.events) {
        const parts = e.method.split(".");
        const eventName = parts.length >= 2 ? parts[1] : e.method;
        document.body.dispatchEvent(
            new CustomEvent("engine:onEventTriggered", {
                detail: {
                    eventName,
                    params: e.params,
                },
            }),
        );
    }
}

export function subscribeToKeyDownEvent() {
    document.addEventListener("keydown", (e) => {
        const response = sendRpcSync("keydown", {
            key: e.key,
            shiftKey: e.shiftKey,
        });
        if (response.error) {
            return;
        }
        const result = response.result as
            | { captured?: boolean; events?: unknown }
            | undefined;
        if (result?.captured) {
            e.preventDefault();
        }
        dispatchKeydownEvents(response.result);
    });
}

const DOUBLE_TAP_MS = 300;

/** Browsers often fire a synthetic mouse click after touchend; ignore those for grid RPCs. */
const GHOST_CLICK_AFTER_TOUCH_MS = 450;

let lastTouchCellKey: string | null = null;
let lastTouchTime = 0;

let lastTouchOnGridCellMs = 0;

function getCellCoordsFromEventTarget(target: EventTarget | null): {
    x: number;
    y: number;
} | null {
    if (!target || !(target instanceof Element)) {
        return null;
    }
    const cell = target.closest(
        "td[data-cell-x][data-cell-y]",
    ) as HTMLTableCellElement | null;
    if (!cell) {
        return null;
    }
    const xs = cell.getAttribute("data-cell-x");
    const ys = cell.getAttribute("data-cell-y");
    if (xs === null || ys === null) {
        return null;
    }
    const x = parseInt(xs, 10);
    const y = parseInt(ys, 10);
    if (Number.isNaN(x) || Number.isNaN(y)) {
        return null;
    }
    return { x, y };
}

function isGhostMouseEventAfterTouch(): boolean {
    return Date.now() - lastTouchOnGridCellMs < GHOST_CLICK_AFTER_TOUCH_MS;
}

export function subscribeToPointerEvents() {
    document.addEventListener("click", (e) => {
        const coords = getCellCoordsFromEventTarget(e.target);
        if (!coords) {
            return;
        }
        if (isGhostMouseEventAfterTouch()) {
            e.preventDefault();
            return;
        }
        void sendRpcAsync("setCursor", coords);
    });

    document.addEventListener("dblclick", (e) => {
        const coords = getCellCoordsFromEventTarget(e.target);
        if (!coords) {
            return;
        }
        if (isGhostMouseEventAfterTouch()) {
            e.preventDefault();
            return;
        }
        e.preventDefault();
        void sendRpcAsync("setCursorAndEdit", coords);
    });

    document.addEventListener("touchend", (e) => {
        const coords = getCellCoordsFromEventTarget(e.target);
        if (!coords) {
            return;
        }
        lastTouchOnGridCellMs = Date.now();
        const key = `${coords.x},${coords.y}`;
        const now = lastTouchOnGridCellMs;
        const isDoubleTap =
            key === lastTouchCellKey && now - lastTouchTime < DOUBLE_TAP_MS;
        lastTouchCellKey = key;
        lastTouchTime = now;
        if (isDoubleTap) {
            void sendRpcAsync("setCursorAndEdit", coords);
        } else {
            void sendRpcAsync("setCursor", coords);
        }
    });
}
