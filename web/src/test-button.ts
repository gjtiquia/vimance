export function init() {
    document.body.addEventListener("click", async (event) => {
        const button = event.target as HTMLElement;
        if (!button.matches("[data-test-button]")) return;

        console.log("button pressed");

        // TODO : refactor this

        // TODO : should send a json rpc that follows the spec
        // - https://www.jsonrpc.org/specification

        const response = await sendRpcAsync({
            message: "hello world from button",
        });

        console.log("response from go wasm:", response);

        // TODO : go wasm side should also send a delayed message back
        // - go wasm side should be able to send/receive according to json rpc spec
        // - go json blog https://go.dev/blog/json
    });
}

init();

async function sendRpcAsync(obj: object): Promise<string> {
    return (globalThis as GoGlobal).goWasmJsonRpcAsync.call(
        JSON.stringify(obj),
    );
}

type GoGlobal = typeof globalThis & {
    goWasmJsonRpcAsync: GoFunc;
};

type GoFunc = {
    call: (message: string) => Promise<string>;
};
