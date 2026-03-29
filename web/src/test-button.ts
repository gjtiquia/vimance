import { sendRpcAsync } from "./wasm";

export function init() {
    document.body.addEventListener("click", async (event) => {
        const button = event.target as HTMLElement;
        if (!button.matches("[data-test-button]")) return;

        // TODO : rpc should be a low level thing, should replace with a simple high-level function like engine.echo
        const response = await sendRpcAsync("echo", {
            message: "helloooooo from js",
        });

        if (response.error) {
            console.error("js: echo.response.error:", response.error);
            return;
        }

        console.log(
            "js: echo.response.result.message:",
            response.result.message,
        );
    });
}

init();
