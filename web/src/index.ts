import {} from "./test-button";
import {
    installJsonRpcReceiver,
    registerJsonRpcHandler,
} from "./wasm/rpc";
import * as wasm from "./wasm";

registerJsonRpcHandler("bridge.goReady", (params) => {
    console.log("bridge.goReady", params);
});

async function initAsync() {
    console.log("hello world from js");

    installJsonRpcReceiver();
    await wasm.initAsync();
}

initAsync();
