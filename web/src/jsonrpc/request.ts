export interface Request {
    jsonrpc: "2.0";
    method: string;
    params: any;
    id: number | string | null;
}

let requestIdCounter = 0;

export function newRequest(method: string, params: any): Request {
    requestIdCounter++;
    return {
        jsonrpc: "2.0",
        method,
        params,
        id: requestIdCounter,
    };
}

export function decodeRequest(jsonString: string): Request {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the request is always valid and does not contain an error
    return obj as Request;
}
