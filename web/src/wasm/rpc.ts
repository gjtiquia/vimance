export type JsonRpcNotification = {
    jsonrpc: "2.0";
    method: string;
    params?: unknown;
};

type JsonRpcHandler = (params: unknown) => void;

type BridgeGlobals = typeof globalThis & {
    vimanceDeliverJsonRpc?: (payload: string) => void;
    vimanceHandleJsonRpc?: (payload: string) => void;
};

const globalScope = globalThis as BridgeGlobals;
const inboundHandlers = new Map<string, JsonRpcHandler>();

export function installJsonRpcReceiver() {
    globalScope.vimanceHandleJsonRpc = function (payload: string) {
        let message: JsonRpcNotification;

        try {
            message = JSON.parse(payload) as JsonRpcNotification;
        } catch (error) {
            console.error("json-rpc: failed to parse inbound payload", error);
            return;
        }

        if (message.jsonrpc !== "2.0" || typeof message.method !== "string") {
            console.warn("json-rpc: ignored malformed inbound message", message);
            return;
        }

        const handler = inboundHandlers.get(message.method);
        if (!handler) {
            console.warn("json-rpc: no handler registered for", message.method);
            return;
        }

        handler(message.params);
    };
}

export function registerJsonRpcHandler(
    method: string,
    handler: JsonRpcHandler,
) {
    inboundHandlers.set(method, handler);
}

export function sendJsonRpcNotification(method: string, params?: unknown) {
    const sendToGo = globalScope.vimanceDeliverJsonRpc;

    if (!sendToGo) {
        console.warn("json-rpc: go runtime is not ready yet");
        return;
    }

    const payload: JsonRpcNotification = {
        jsonrpc: "2.0",
        method,
        params,
    };

    sendToGo(JSON.stringify(payload));
}
