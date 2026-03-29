import { setClientMode } from "./engine/mode";

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

function getEventHandler(eventName: string) {
    switch (eventName) {
        case "OnModeChanged":
            return handleOnModeChanged;

        case "OnCursorMoved":
            return handleOnCursorMoved;

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

    setClientMode(params.mode);

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
        } else {
            input.setSelectionRange(0, 0);
        }
    } else if (mode === "n") {
        const inputCell = table.querySelector(
            "[data-cell-variant='input']",
        ) as HTMLTableCellElement | null;
        if (!inputCell) {
            return;
        }

        const input = inputCell.querySelector("input");
        const value =
            input?.value?.trim() ?? inputCell.textContent?.trim() ?? "";
        replaceCell(table, inputCell, "normal", value);
    }
}

function handleOnCursorMoved(table: HTMLTableElement, params: any) {
    const x = params.x;
    const y = params.y;

    const targetCell = table.querySelector(
        `td[data-cell-x="${x}"][data-cell-y="${y}"]`,
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
        `td[data-cell-x="${x}"][data-cell-y="${y}"]`,
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

    const input = newCell.querySelector("input");
    if (input) {
        input.value = value;
    } else {
        newCell.textContent = value;
    }

    oldCell.replaceWith(newCell);
    return newCell;
}
