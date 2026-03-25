export function init() {
    document.body.addEventListener("click", async (event) => {
        const button = event.target as HTMLElement;
        if (!button.matches("[data-test-button]")) return;

        console.log("button pressed");

        // TODO : test sending a json rpc to go wasm, which calls back a json rpc on response
    });
}

init();
