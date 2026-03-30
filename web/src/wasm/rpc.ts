import * as jsonrpc from "../jsonrpc";

type GoGlobal = typeof globalThis & {
    goToJsJsonRpcAsync: (json: string) => Promise<string>;
    jsToGoJsonRpcAsync: GoFunc;
    /** Go wasm_exec passes the JSON request string as `this` (see js.NewSyncStringFunc). */
    jsToGoJsonRpcSync?: (this: string) => string;
};

type GoFunc = {
    call: (json: string) => Promise<string>;
};

(globalThis as GoGlobal).goToJsJsonRpcAsync = onReceiveJsonRpcAsync;

/** Synchronous engine events from Go (setCursor / setCursorAndEdit), same shape as keydown result.events. */
type EngineEventPayload = { method: string; params: Record<string, unknown> };

(globalThis as Record<string, unknown>).goEngineEventsSync = (
    json: string,
): void => {
    const events = JSON.parse(json) as EngineEventPayload[];
    for (const e of events) {
        const parts = e.method.split(".");
        const eventName = parts.length >= 2 ? parts[1] : e.method;
        document.body.dispatchEvent(
            new CustomEvent("engine:onEventTriggered", {
                detail: {
                    eventName,
                    params: e.params,
                },
            }),
        );
    }
};

export function sendRpcSync(method: string, params: unknown): jsonrpc.Response {
    const request = jsonrpc.newRequest(method, params);
    const requestJson = JSON.stringify(request);
    const fn = (globalThis as GoGlobal).jsToGoJsonRpcSync;
    if (typeof fn !== "function") {
        throw new Error("jsToGoJsonRpcSync is not registered (WASM not loaded?)");
    }
    const responseJson = fn.call(requestJson);
    return jsonrpc.decodeResponse(responseJson);
}

export async function sendRpcAsync(
    method: string,
    params: any,
): Promise<jsonrpc.Response> {
    const request = jsonrpc.newRequest(method, params);
    const requestJson = JSON.stringify(request);

    const responseJson = await (globalThis as GoGlobal).jsToGoJsonRpcAsync.call(
        requestJson,
    );

    return jsonrpc.decodeResponse(responseJson);
}

async function onReceiveJsonRpcAsync(jsonString: string): Promise<string> {
    const request = jsonrpc.decodeRequest(jsonString);

    // console.log(`js: ${request.method}.request.params:`, request.params);

    // TODO : refactor this such that wasm/rpc does not depend on engine
    const { ok, responseJson } = tryHandleEngineEvents(request);
    if (ok) {
        return responseJson;
    }

    switch (request.method) {
        case "echo":
            return handleEcho(request);

        default:
            const response = jsonrpc.newMethodNotFoundResponse(request);
            console.error(`js: ${response.error?.message}`);
            return JSON.stringify(response);
    }
}

async function handleEcho(request: jsonrpc.Request): Promise<string> {
    type EchoParams = { message: string };
    const echoParams = request.params as EchoParams; // TODO : should use some sort of runtime validation eg. zod

    const response = jsonrpc.newResponse(
        { message: `js echoooooo ${echoParams.message}` },
        request.id,
    );

    const responseJson = JSON.stringify(response);
    return responseJson;
}

// ===== engine router =====

function tryHandleEngineEvents(request: jsonrpc.Request) {
    if (
        !request.method.startsWith("engine.") ||
        request.method.split(".").length !== 2
    ) {
        return { ok: false } as const;
    }

    const eventName = request.method.split(".")[1];

    document.body.dispatchEvent(
        new CustomEvent("engine:onEventTriggered", {
            // bubbles: true,
            // composed: true,
            detail: {
                eventName,
                params: request.params,
            },
        }),
    );

    const response = jsonrpc.newSuccessResponse(request.id);
    return { ok: true, responseJson: JSON.stringify(response) } as const;
}
