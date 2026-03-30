import { sendRpcAsync } from "../wasm";
import { getClientMode } from "./mode";

function shouldPreventDefaultForVim(key: string): boolean {
    const mode = getClientMode();
    if (mode === "n") {
        return [
            "i",
            "a",
            "v",
            "h",
            "j",
            "k",
            "l",
            "w",
            "e",
            "b",
            "ArrowLeft",
            "ArrowRight",
            "ArrowUp",
            "ArrowDown",
            "Enter",
        ].includes(key);
    }
    if (mode === "i" && key === "Escape") {
        return true;
    }
    if (mode === "v" && key === "Escape") {
        return true;
    }
    return false;
}

export function subscribeToKeyDownEvent() {
    document.addEventListener("keydown", (e) => {
        if (shouldPreventDefaultForVim(e.key)) {
            e.preventDefault();
        }
        sendRpcAsync("keydown", {
            key: e.key,
        });
    });
}

const DOUBLE_TAP_MS = 300;

let lastTouchCellKey: string | null = null;
let lastTouchTime = 0;

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

export function subscribeToPointerEvents() {
    document.addEventListener("click", (e) => {
        const coords = getCellCoordsFromEventTarget(e.target);
        if (!coords) {
            return;
        }
        void sendRpcAsync("setCursor", coords);
    });

    document.addEventListener("dblclick", (e) => {
        const coords = getCellCoordsFromEventTarget(e.target);
        if (!coords) {
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
        const key = `${coords.x},${coords.y}`;
        const now = Date.now();
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
