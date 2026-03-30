init();

export function init() {
    const elements = document.body.querySelectorAll(
        "[data-engine-debug-console]",
    );

    document.body.addEventListener("engine:onEventTriggered", async (event) => {
        const customEvent = event as CustomEvent;

        // console.log("js: engine:onEventTriggered:", {
        //     type: customEvent.type,
        //     detail: customEvent.detail,
        // });

        elements.forEach((element) => {
            const container = element as HTMLElement;

            const eventName = customEvent.detail.eventName;
            const params = customEvent.detail.params;

            const log = document.createElement("p");
            log.textContent = `${eventName}: ${JSON.stringify(params)}`;

            // must call BEFORE appending child
            const stickToBottom = isScrolledToBottom(container);

            container.appendChild(log);

            if (stickToBottom) {
                container.scrollTop = container.scrollHeight;
            }
        });
    });
}

/** Pixels from the bottom to still count as "at bottom" (subpixel / rounding). */
const BOTTOM_THRESHOLD_PX = 4;

function isScrolledToBottom(el: HTMLElement): boolean {
    return (
        el.scrollHeight - el.scrollTop - el.clientHeight <= BOTTOM_THRESHOLD_PX
    );
}

