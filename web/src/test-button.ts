export function init() {
    document.body.addEventListener("click", async (event) => {
        const button = event.target as HTMLElement;
        if (!button.matches("[data-test-button]")) return;

        const response = await sendRpcAsync("echo", {
            message: "helloooooo from js",
        });

        if (response.error) {
            console.error(
                "js: echo.response.error.message:",
                response.error.message,
            );
            return;
        }

        console.log(
            "js: echo.response.result.message:",
            response.result.message,
        );
    });
}

init();

// ===== send RPC =====

async function sendRpcAsync(
    method: string,
    params: any,
): Promise<JsonRpcResponse> {
    const request = newJsonRpcRequest(method, params);
    const requestJson = JSON.stringify(request);

    const responseJson = await (globalThis as GoGlobal).jsToGoJsonRpcAsync.call(
        requestJson,
    );
    return decodeJsonRpcResponse(responseJson);
}

type GoGlobal = typeof globalThis & {
    goToJsJsonRpcAsync: (json: string) => Promise<string>;
    jsToGoJsonRpcAsync: GoFunc;
};

type GoFunc = {
    call: (json: string) => Promise<string>;
};

// ===== receive RPC =====

(globalThis as GoGlobal).goToJsJsonRpcAsync = onReceiveJsonRpcAsync;

async function onReceiveJsonRpcAsync(jsonString: string): Promise<string> {
    const request = decodeJsonRpcRequest(jsonString);

    // TODO : route to different handlers based on the method in the request, and return appropriate responses
    // for now its assuming "echo"

    type EchoParams = { message: string };
    const echoParams = request.params as EchoParams; // TODO : should use some sort of runtime validation eg. zod

    console.log(
        `js: ${request.method}.request.params.message:`,
        echoParams.message,
    );

    const response = newJsonRpcResponse(
        { message: `js echoooooo ${echoParams.message}` },
        request.id,
    );

    const responseJson = JSON.stringify(response);
    return responseJson;
}

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
