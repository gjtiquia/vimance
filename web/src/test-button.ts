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

        console.log("awaited response from go wasm:", response);
    });
}

init();

// ===== send RPC =====

async function sendRpcAsync(method: string, params: any): Promise<string> {
    const request = newJsonRpcRequest(method, params);
    const requestJson = JSON.stringify(request);

    return (globalThis as GoGlobal).jsToGoJsonRpcAsync.call(requestJson);
}

type GoGlobal = typeof globalThis & {
    jsToGoJsonRpcAsync: GoFunc;
};

type GoFunc = {
    call: (message: string) => Promise<string>;
};

// ===== receive RPC =====

async function onReceiveJsonRpcAsync(jsonString: string): Promise<string> {
    const request = decodeJsonRpcRequest(jsonString);

    console.log("js.onReceiveJsonRpcAsync:", request);

    // TODO : route to different handlers based on the method in the request, and return appropriate responses
    // for now its assuming "echo"
    // params should be structured data, but for simplicity we are just using a string here for now

    const response = newJsonRpcResponse(
        { message: `js echoooooo ${request.params}` },
        request.id,
    );

    const responseJson = JSON.stringify(response);
    return responseJson;
}

Object.defineProperty(globalThis, "goToJsJsonRpcAsync", {
    value: onReceiveJsonRpcAsync,
    writable: false,
    configurable: false,
    enumerable: false,
});

// ====== JSON RPC types ======

interface JsonRpcRequest {
    jsonrpc: "2.0";
    method: string;
    params: any;
    id: number | string | null;
}

let requestIdCounter = 0;

function newJsonRpcRequest(method: string, params: any): JsonRpcRequest {
    requestIdCounter++;
    return {
        jsonrpc: "2.0",
        method,
        params,
        id: requestIdCounter,
    };
}

// decode json rpc request
function decodeJsonRpcRequest(jsonString: string): JsonRpcRequest {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the request is always valid and does not contain an error
    return obj as JsonRpcRequest;
}

interface JsonRpcResponse {
    jsonrpc: "2.0";
    result?: any;
    error?: { code: number; message: string; data?: any };
    id: number | string | null;
}

function newJsonRpcResponse(
    result: any,
    id: number | string | null,
): JsonRpcResponse {
    return {
        jsonrpc: "2.0",
        result,
        id,
    };
}

function decodeJsonRpcResponse(jsonString: string): JsonRpcResponse {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the response is always valid and does not contain an error
    return obj as JsonRpcResponse;
}
