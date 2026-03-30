import {} from "./test-button";
import {} from "./engine-debug-console";
import { hydrateTableFromEngine } from "./table";
import * as engine from "./engine";
import * as wasm from "./wasm";

async function initAsync() {
    console.log("js: running...");

    await wasm.initAsync();

    engine.init();

    // must run this AFTER wasm is running
    hydrateTableFromEngine();
}

initAsync();
