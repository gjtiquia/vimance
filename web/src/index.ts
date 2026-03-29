import {} from "./test-button";
import * as wasm from "./wasm";

async function initAsync() {
    console.log("js: running...");
    await wasm.initAsync();
}

initAsync();
