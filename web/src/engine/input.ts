import { sendRpcAsync } from "../wasm";
import { getClientMode } from "./mode";

function shouldPreventDefaultForVim(key: string): boolean {
    const mode = getClientMode();
    if (mode === "n") {
        return [
            "i",
            "a",
            "v",
            "h",
            "j",
            "k",
            "l",
            "w",
            "e",
            "b",
        ].includes(key);
    }
    if (mode === "i" && key === "Escape") {
        return true;
    }
    if (mode === "v" && key === "Escape") {
        return true;
    }
    return false;
}

export function subscribeToKeyDownEvent() {
    document.addEventListener("keydown", (e) => {
        if (shouldPreventDefaultForVim(e.key)) {
            e.preventDefault();
        }
        sendRpcAsync("keydown", {
            key: e.key,
        });
    });
}
