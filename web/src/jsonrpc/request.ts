export interface JsonRpcRequest {
    jsonrpc: "2.0";
    method: string;
    params: any;
    id: number | string | null;
}

let requestIdCounter = 0;

export function newJsonRpcRequest(method: string, params: any): JsonRpcRequest {
    requestIdCounter++;
    return {
        jsonrpc: "2.0",
        method,
        params,
        id: requestIdCounter,
    };
}

export function decodeJsonRpcRequest(jsonString: string): JsonRpcRequest {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the request is always valid and does not contain an error
    return obj as JsonRpcRequest;
}
