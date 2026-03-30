import { createExports } from "./exports";
import { type Imports } from "./imports";

const WASM_URL = "/public/main.wasm";

// Go WASM runtime from wasm_exec.js (loaded as script before this module).
declare const Go: new () => {
    importObject: WebAssembly.Imports;
    run(instance: WebAssembly.Instance): Promise<number>; // returns exit code
};

// Exports from our TinyGo WASM module (main.wasm).
interface WasmExports extends WebAssembly.Exports, Imports {
    memory: WebAssembly.Memory;
}

type Wasm = WebAssembly.Instance & { exports: WasmExports };

export let wasm: Wasm | undefined = undefined;

export async function initAsync() {
    const go = new Go();

    // import functions for main.wasm to use
    go.importObject.env = createExports();

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
            fetch(WASM_URL),
            go.importObject,
        );

        wasm = result.instance as Wasm;

        console.log("js: running main.wasm...");

        // const exitCode = await go.run(wasm); // runs main()
        // console.log("js: main.wasm exit code:", exitCode);

        // dont await it or else it will be blocking
        void go.run(wasm)
    } catch (err) {
        console.error("js: wasm.initAsync: error");
        console.error(err);
    }
}
