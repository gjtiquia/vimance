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

        default:
            return null;
    }
}

function handleOnModeChanged(table: HTMLTableElement, params: any) {
    console.log("js: table: handleOnModeChanged:", params);

    const mode = params.mode;

    if (mode === "i") {
        const normalCell = table.querySelector("[data-cell-variant='normal']") as HTMLTableCellElement | null;
        if (!normalCell) {
            console.error("js: table: handleOnModeChanged: No normal cell found!");
            return;
        }

        const value = normalCell.textContent?.trim() || "";
        replaceCell(table, normalCell, "input", value);
    }
}

function replaceCell(
    table: HTMLTableElement,
    oldCell: HTMLTableCellElement,
    variant: string,
    value: string,
): HTMLTableCellElement | null {
    const template = table.querySelector(`template[data-cell-template="${variant}"]`) as HTMLTemplateElement | null;
    if (!template) {
        console.error(`js: table: replaceCell: No template found for variant: ${variant}`);
        return null;
    }

    const newCell = template.content.firstElementChild!.cloneNode(true) as HTMLTableCellElement;

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
