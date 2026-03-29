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
        const normalCell = table.querySelector("[data-cell-variant='normal']");
        if (!normalCell) {
            console.error("js: table: handleOnModeChanged: No normal cell found! unable to change mode to insert mode!");
        }

        // TODO :
        // - save cell content
        // - change cell to input
        // - add cell content to input
    }
}
