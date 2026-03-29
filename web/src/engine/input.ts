import { sendRpcAsync } from "../wasm";

export function subscribeToKeyDownEvent() {
    document.addEventListener("keydown", (e) => {
        // console.log("keydown:", e);
        sendRpcAsync("keydown", {
            key: e.key,
        });
    });
}
