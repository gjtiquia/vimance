// Go WASM runtime from wasm_exec.js (loaded as script before this module).
declare const Go: new () => {
    importObject: WebAssembly.Imports;
    run(instance: WebAssembly.Instance): Promise<number>; // returns exit code
};

type Wasm = WebAssembly.Instance;

type JsonRpcNotification = {
    jsonrpc: "2.0";
    method: string;
    params?: unknown;
};

export let wasm: Wasm | undefined = undefined;

export async function initAsync() {
    const go = new Go();

    // polyfill if browsers do not support WebAssembly.instantiateStreaming
    if (!WebAssembly.instantiateStreaming) {
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    // fetch wasm and run main.wasm
    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch("/public/main.wasm"),
            go.importObject,
        );

        wasm = result.instance as Wasm;

        console.log("running main.wasm...");
        const exitCode = await go.run(wasm); // runs main()
        console.log("main.wasm exit code:", exitCode);
    } catch (err) {
        console.error("wasm.initAsync: error");
        console.error(err);
    }
}
