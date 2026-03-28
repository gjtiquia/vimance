import {
    registerJsonRpcHandler,
    sendJsonRpcNotification,
} from "./wasm/rpc";

registerJsonRpcHandler("bridge.testButtonResult", (params) => {
    console.log("bridge.testButtonResult", params);
});

export function init() {
    document.body.addEventListener("click", async (event) => {
        if (!(event.target instanceof HTMLElement)) return;

        const button = event.target;
        if (!button.matches("[data-test-button]")) return;

        console.log("button pressed");

        sendJsonRpcNotification("bridge.testButtonPressed", {
            source: "test-button",
        });
    });
}

init();
