import * as wasm from "./wasm"

async function initAsync() {
    console.log("hello world from js");

    await wasm.initAsync();
}

initAsync();
