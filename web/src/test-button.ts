export function init() {
    document.body.addEventListener("click", async (event) => {
        const button = event.target as HTMLElement;
        if (!button.matches("[data-test-button]")) return;

        console.log("button pressed");

        // TODO : refactor this

        // TODO : should send a json rpc that follows the spec
        // - https://www.jsonrpc.org/specification

        const response = await sendRpcAsync("echo", {
            message: "helloooooo from js",
        });

        console.log("response from go wasm:", response);

        // TODO : go wasm side should also send a delayed message back
        // - go wasm side should be able to send/receive according to json rpc spec
        // - go json blog https://go.dev/blog/json
    });
}

init();

// ===== send RPC =====

let requestIdCounter = 0;

async function sendRpcAsync(method: string, params: object): Promise<string> {
    requestIdCounter++;
    return (globalThis as GoGlobal).jsToGoJsonRpcAsync.call(
        JSON.stringify({
            jsonrpc: "2.0",
            method,
            params,
            id: requestIdCounter,
        }),
    );
}

type GoGlobal = typeof globalThis & {
    jsToGoJsonRpcAsync: GoFunc;
};

type GoFunc = {
    call: (message: string) => Promise<string>;
};

// ===== receive RPC =====

function onReceiveJsonRpc(message: string): string {
    console.log("received json rpc from go wasm:", message);

    // TODO : json rpc spec
    return "test response from js"
}

Object.defineProperty(globalThis, "goToJsJsonRpcAsync", {
    value: onReceiveJsonRpc,
    writable: false,
    configurable: false,
    enumerable: false,
});
