/** Mirrors engine mode so keydown can preventDefault before the same key is applied as text input. */

export type ClientMode = "n" | "i" | "v";

let clientMode: ClientMode = "n";

export function setClientMode(mode: string): void {
    if (mode === "n" || mode === "i" || mode === "v") {
        clientMode = mode;
    }
}

export function getClientMode(): ClientMode {
    return clientMode;
}
