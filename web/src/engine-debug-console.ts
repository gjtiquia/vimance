export function init() {
    const elements = document.body.querySelectorAll("[data-engine-debug-console]");

    document.body.addEventListener("engine:onEventTriggered", async (event) => {
        const customEvent = event as CustomEvent;

        console.log("js: engine:onEventTriggered:", {
            type: customEvent.type,
            detail: customEvent.detail,
        });

        elements.forEach((parent) => {
            const eventName = customEvent.detail.eventName;
            const params = customEvent.detail.params;
            const child = document.createElement("p");
            child.textContent = `${eventName}: ${JSON.stringify(params)}`;
            parent.appendChild(child);
        });
    });
}

init();
