import { sendRpcSync } from "./wasm";

const CELL_BASE =
    "border border-stone-50/25 px-2 py-1 h-8 min-w-0 truncate ";
const HEADER_CELL = CELL_BASE + "text-left font-bold text-stone-50/70";
const TD_NORMAL = CELL_BASE + "bg-stone-50/30";
const TD_DEFAULT = CELL_BASE;
const TD_VISUAL = CELL_BASE + "bg-blue-50/10 text-stone-50/70";

init();

export function init() {
    const tables = document.body.querySelectorAll("[data-table]");

    document.body.addEventListener("engine:onEventTriggered", async (event) => {
        const customEvent = event as CustomEvent;

        const eventName = customEvent.detail.eventName;
        const params = customEvent.detail.params;

        const handler = getEventHandler(eventName);
        if (!handler) {
            console.warn(`js: table: No handler found for event: ${eventName}`);
            return;
        }

        tables.forEach((element) => {
            const table = element as HTMLTableElement;
            handler(table, params);
        });
    });
}

/** Call after WASM is running. Fills the table from engine state (getGrid). */
export function hydrateTableFromEngine(): void {
    const tables = document.body.querySelectorAll("[data-table]");
    const response = sendRpcSync("getGrid", {});
    if (response.error || !response.result) {
        console.error("js: table: getGrid failed", response.error);
        return;
    }
    const result = response.result as {
        cells: string[][];
        cursorX: number;
        cursorY: number;
    };
    const { cells, cursorX, cursorY } = result;
    if (!Array.isArray(cells) || cells.length === 0) {
        console.error("js: table: getGrid returned empty cells");
        return;
    }

    tables.forEach((element) => {
        const table = element as HTMLTableElement;
        const tbody = table.querySelector("[data-table-tbody]");
        if (!tbody) {
            console.error("js: table: no [data-table-tbody]");
            return;
        }
        tbody.replaceChildren();

        for (let y = 0; y < cells.length; y++) {
            const row = cells[y];
            const tr = document.createElement("tr");
            tr.setAttribute("data-row-y", String(y));
            for (let x = 0; x < row.length; x++) {
                const variant =
                    x === cursorX && y === cursorY ? "normal" : "default";
                const td = createDataCell(
                    table,
                    x,
                    y,
                    row[x] ?? "",
                    variant,
                    y === 0,
                );
                tr.appendChild(td);
            }
            tbody.appendChild(tr);
        }
    });
}

function cellClassName(
    variant: "normal" | "default",
    isHeaderRow: boolean,
): string {
    if (isHeaderRow) {
        return variant === "normal"
            ? HEADER_CELL + " bg-stone-50/30"
            : HEADER_CELL;
    }
    return variant === "normal" ? TD_NORMAL : TD_DEFAULT;
}

function createDataCell(
    table: HTMLTableElement,
    x: number,
    y: number,
    value: string,
    variant: "normal" | "default",
    isHeaderRow: boolean,
): HTMLTableCellElement {
    const template = table.querySelector(
        `template[data-cell-template="${variant}"]`,
    ) as HTMLTemplateElement | null;
    if (!template?.content.firstElementChild) {
        const td = document.createElement("td");
        td.setAttribute("data-cell-variant", variant);
        td.setAttribute("data-cell-x", String(x));
        td.setAttribute("data-cell-y", String(y));
        td.className = cellClassName(variant, isHeaderRow);
        td.textContent = value;
        return td;
    }

    const td = template.content.firstElementChild.cloneNode(
        true,
    ) as HTMLTableCellElement;
    td.setAttribute("data-cell-variant", variant);
    td.setAttribute("data-cell-x", String(x));
    td.setAttribute("data-cell-y", String(y));
    td.className = cellClassName(variant, isHeaderRow);

    const input = td.querySelector("input");
    if (input) {
        input.value = value;
    } else {
        td.textContent = value;
    }
    return td;
}

function getEventHandler(eventName: string) {
    switch (eventName) {
        case "OnModeChanged":
            return handleOnModeChanged;

        case "OnCursorMoved":
            return handleOnCursorMoved;

        case "OnBufferChanged":
            return handleOnBufferChanged;

        case "OnClipboardWrite":
            return handleOnClipboardWrite;

        case "OnSelectionChanged":
            return handleOnSelectionChanged;

        default:
            return null;
    }
}

function getCellDisplayValue(cell: HTMLTableCellElement): string {
    const input = cell.querySelector("input");
    if (input) {
        return input.value;
    }
    return cell.textContent?.trim() ?? "";
}

function handleOnModeChanged(table: HTMLTableElement, params: any) {
    console.log("js: table: handleOnModeChanged:", params);

    const mode = params.mode;

    if (mode === "i") {
        const normalCell = table.querySelector(
            "[data-cell-variant='normal']",
        ) as HTMLTableCellElement | null;
        if (!normalCell) {
            console.error(
                "js: table: handleOnModeChanged: No normal cell found!",
            );
            return;
        }

        const value = normalCell.textContent?.trim() || "";
        const newCell = replaceCell(table, normalCell, "input", value);
        if (!newCell) {
            return;
        }

        const input = newCell.querySelector("input");
        if (!input) {
            return;
        }

        input.focus();
        const insertPosition = params.insertPosition as string | undefined;
        if (insertPosition === "after") {
            const len = input.value.length;
            input.setSelectionRange(len, len);
        } else if (insertPosition === "highlight") {
            const len = input.value.length;
            input.setSelectionRange(0, len);
        } else {
            input.setSelectionRange(0, 0);
        }
    } else if (mode === "n") {
        table.querySelectorAll("[data-cell-variant='visual']").forEach((el) => {
            const cell = el as HTMLTableCellElement;
            replaceCell(table, cell, "default", getCellDisplayValue(cell));
        });

        const inputCell = table.querySelector(
            "[data-cell-variant='input']",
        ) as HTMLTableCellElement | null;
        if (!inputCell) {
            return;
        }

        const input = inputCell.querySelector("input");
        const value =
            input?.value?.trim() ?? inputCell.textContent?.trim() ?? "";

        const xs = inputCell.getAttribute("data-cell-x");
        const ys = inputCell.getAttribute("data-cell-y");
        if (xs !== null && ys !== null) {
            const x = parseInt(xs, 10);
            const y = parseInt(ys, 10);
            if (!Number.isNaN(x) && !Number.isNaN(y)) {
                try {
                    sendRpcSync("setCellValue", { x, y, value });
                } catch (e) {
                    console.error("js: table: setCellValue failed", e);
                }
            }
        }

        replaceCell(table, inputCell, "normal", value);
    }
}

function handleOnBufferChanged(table: HTMLTableElement, _params: unknown) {
    void table;
    hydrateTableFromEngine();
}

function handleOnSelectionChanged(table: HTMLTableElement, params: any) {
    table.querySelectorAll("[data-cell-variant='visual']").forEach((el) => {
        const cell = el as HTMLTableCellElement;
        replaceCell(table, cell, "default", getCellDisplayValue(cell));
    });

    const sx = params.startX as number;
    const sy = params.startY as number;
    const ex = params.endX as number;
    const ey = params.endY as number;
    const cursorX = params.cursorX as number;
    const cursorY = params.cursorY as number;

    for (let y = sy; y <= ey; y++) {
        for (let x = sx; x <= ex; x++) {
            const cell = table.querySelector(
                `[data-cell-x="${x}"][data-cell-y="${y}"]`,
            ) as HTMLTableCellElement | null;
            if (
                cell &&
                cell.getAttribute("data-cell-variant") !== "input"
            ) {
                replaceCell(table, cell, "visual", getCellDisplayValue(cell));
            }
        }
    }

    const cursorCell = table.querySelector(
        `[data-cell-x="${cursorX}"][data-cell-y="${cursorY}"]`,
    ) as HTMLTableCellElement | null;
    if (cursorCell) {
        replaceCell(
            table,
            cursorCell,
            "normal",
            getCellDisplayValue(cursorCell),
        );
    }
}

function handleOnClipboardWrite(_table: HTMLTableElement, params: { text?: string }) {
    const text = params.text ?? "";
    if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
        void navigator.clipboard.writeText(text).catch((err) => {
            console.warn("js: table: clipboard write failed", err);
        });
    }
}

function handleOnCursorMoved(table: HTMLTableElement, params: any) {
    const x = params.x;
    const y = params.y;

    const targetCell = table.querySelector(
        `[data-cell-x="${x}"][data-cell-y="${y}"]`,
    ) as HTMLTableCellElement | null;
    if (!targetCell) {
        return;
    }

    const normalCell = table.querySelector(
        "[data-cell-variant='normal']",
    ) as HTMLTableCellElement | null;
    if (!normalCell) {
        return;
    }

    if (normalCell === targetCell) {
        return;
    }

    const fromValue = getCellDisplayValue(normalCell);
    const toValue = getCellDisplayValue(targetCell);

    replaceCell(table, normalCell, "default", fromValue);

    const newTarget = table.querySelector(
        `[data-cell-x="${x}"][data-cell-y="${y}"]`,
    ) as HTMLTableCellElement | null;
    if (!newTarget) {
        return;
    }

    replaceCell(table, newTarget, "normal", toValue);
}

function replaceCell(
    table: HTMLTableElement,
    oldCell: HTMLTableCellElement,
    variant: string,
    value: string,
): HTMLTableCellElement | null {
    const template = table.querySelector(
        `template[data-cell-template="${variant}"]`,
    ) as HTMLTemplateElement | null;
    if (!template) {
        console.error(
            `js: table: replaceCell: No template found for variant: ${variant}`,
        );
        return null;
    }

    const newCell = template.content.firstElementChild!.cloneNode(
        true,
    ) as HTMLTableCellElement;

    const x = oldCell.getAttribute("data-cell-x");
    const y = oldCell.getAttribute("data-cell-y");
    if (x !== null) newCell.setAttribute("data-cell-x", x);
    if (y !== null) newCell.setAttribute("data-cell-y", y);

    const yNum = y !== null ? parseInt(y, 10) : -1;
    const isHeaderRow = yNum === 0;
    let baseClass: string;
    if (variant === "input") {
        baseClass = CELL_BASE;
    } else if (variant === "visual") {
        baseClass = isHeaderRow
            ? HEADER_CELL + " bg-blue-50/10 text-stone-50/70"
            : TD_VISUAL;
    } else if (isHeaderRow) {
        baseClass =
            variant === "normal"
                ? HEADER_CELL + " bg-stone-50/30"
                : HEADER_CELL;
    } else {
        baseClass = variant === "normal" ? TD_NORMAL : TD_DEFAULT;
    }
    newCell.className = baseClass;

    const input = newCell.querySelector("input");
    if (input) {
        input.value = value;
    } else {
        newCell.textContent = value;
    }

    newCell.setAttribute("data-cell-variant", variant);

    oldCell.replaceWith(newCell);
    return newCell;
}
