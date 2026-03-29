import {} from "./test-button";
import * as engine from "./engine";
import * as wasm from "./wasm";

async function initAsync() {
    console.log("js: running...");
    engine.init();
    await wasm.initAsync();
}

initAsync();
