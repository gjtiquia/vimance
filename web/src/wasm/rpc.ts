import { newJsonRpcRequest, decodeJsonRpcRequest } from "../jsonrpc/request";
import {
    JsonRpcResponse,
    decodeJsonRpcResponse,
    newJsonRpcResponse,
} from "../jsonrpc/response";

type GoGlobal = typeof globalThis & {
    goToJsJsonRpcAsync: (json: string) => Promise<string>;
    jsToGoJsonRpcAsync: GoFunc;
};

type GoFunc = {
    call: (json: string) => Promise<string>;
};

(globalThis as GoGlobal).goToJsJsonRpcAsync = onReceiveJsonRpcAsync;

export async function sendRpcAsync(
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
