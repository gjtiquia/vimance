export function init() {
    document.body.addEventListener("engine:onEventTriggered", async (event) => {
        const customEvent = event as CustomEvent;

        console.log("js: engine:onEventTriggered:", {
            type: customEvent.type,
            detail: customEvent.detail,
        });

        // TODO : append the <pre> element with "{}: {params}"
    });
}

init();
