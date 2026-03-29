import { Request } from "./request";

export interface Response {
    jsonrpc: "2.0";
    result?: any;
    error?: { code: number; message: string; data?: any };
    id: number | string | null;
}

export function newResponse(result: any, id: number | string | null): Response {
    return {
        jsonrpc: "2.0",
        result,
        id,
    };
}

export function newSuccessResponse(id: number | string | null): Response {
    return newResponse({ success: true }, id);
}

export function decodeResponse(jsonString: string): Response {
    const obj = JSON.parse(jsonString);

    // for simplicity, we assume the response is always valid and does not contain an error
    return obj as Response;
}

export function newParseErrorResponse(
    message: string,
    id: number | string | null,
): Response {
    return newErrorResponse(-32700, `Parse error: ${message}`, id);
}

export function newInvalidRequestResponse(
    message: string,
    id: number | string | null,
): Response {
    return newErrorResponse(-32600, `Invalid request: ${message}`, id);
}

export function newMethodNotFoundResponse(request: Request): Response {
    return newErrorResponse(
        -32601,
        `Method not found: ${request.method}`,
        request.id,
    );
}

export function newInvalidParamsResponse(
    message: string,
    id: number | string | null,
): Response {
    return newErrorResponse(-32602, `Invalid params: ${message}`, id);
}

export function newInternalErrorResponse(
    message: string,
    id: number | string | null,
): Response {
    return newErrorResponse(-32603, `Internal error: ${message}`, id);
}

export function newServerErrorResponse(
    code: number,
    message: string,
    id: number | string | null,
    data?: any,
): Response {
    if (code < -32099 || code > -32000) {
        throw new Error("Server error code must be between -32099 and -32000");
    }
    return newErrorResponse(code, `Server error: ${message}`, id, data);
}

export function newErrorResponse(
    code: number,
    message: string,
    id: number | string | null,
    data?: any,
): Response {
    return {
        jsonrpc: "2.0",
        error: { code, message, data },
        id,
    };
}
