export function init() {
    document.body.addEventListener("click", async (event) => {
        const element = event.target as HTMLElement;
        if (!element.matches("[data-engine-debug-console]")) return;
    });
}

init();
