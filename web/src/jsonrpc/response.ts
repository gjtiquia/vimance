export interface JsonRpcResponse {
    jsonrpc: "2.0";
    result?: any;
    error?: { code: number; message: string; data?: any };
    id: number | string | null;
}

export function newJsonRpcResponse(
    result: any,
    id: number | string | null,
): JsonRpcResponse {
    return {
        jsonrpc: "2.0",
        result,
        id,
    };
}

export function decodeJsonRpcResponse(jsonString: string): JsonRpcResponse {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the response is always valid and does not contain an error
    return obj as JsonRpcResponse;
}
