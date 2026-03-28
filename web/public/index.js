// web/src/wasm/rpc.ts
var globalScope = globalThis;
var inboundHandlers = new Map;
function installJsonRpcReceiver() {
  globalScope.vimanceHandleJsonRpc = function(payload) {
    let message;
    try {
      message = JSON.parse(payload);
    } catch (error) {
      console.error("json-rpc: failed to parse inbound payload", error);
      return;
    }
    if (message.jsonrpc !== "2.0" || typeof message.method !== "string") {
      console.warn("json-rpc: ignored malformed inbound message", message);
      return;
    }
    const handler = inboundHandlers.get(message.method);
    if (!handler) {
      console.warn("json-rpc: no handler registered for", message.method);
      return;
    }
    handler(message.params);
  };
}
function registerJsonRpcHandler(method, handler) {
  inboundHandlers.set(method, handler);
}
function sendJsonRpcNotification(method, params) {
  const sendToGo = globalScope.vimanceDeliverJsonRpc;
  if (!sendToGo) {
    console.warn("json-rpc: go runtime is not ready yet");
    return;
  }
  const payload = {
    jsonrpc: "2.0",
    method,
    params
  };
  sendToGo(JSON.stringify(payload));
}

// web/src/test-button.ts
registerJsonRpcHandler("bridge.testButtonResult", (params) => {
  console.log("bridge.testButtonResult", params);
});
function init() {
  document.body.addEventListener("click", async (event) => {
    if (!(event.target instanceof HTMLElement))
      return;
    const button = event.target;
    if (!button.matches("[data-test-button]"))
      return;
    console.log("button pressed");
    sendJsonRpcNotification("bridge.testButtonPressed", {
      source: "test-button"
    });
  });
}
init();

// web/src/wasm/wasm.ts
var wasm = undefined;
async function initAsync() {
  const go = new Go;
  if (!WebAssembly.instantiateStreaming) {
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer();
      return await WebAssembly.instantiate(source, importObject);
    };
  }
  try {
    const result = await WebAssembly.instantiateStreaming(fetch("/public/main.wasm"), go.importObject);
    wasm = result.instance;
    console.log("running main.wasm...");
    const exitCode = await go.run(wasm);
    console.log("main.wasm exit code:", exitCode);
  } catch (err) {
    console.error("wasm.initAsync: error");
    console.error(err);
  }
}
// web/src/index.ts
registerJsonRpcHandler("bridge.goReady", (params) => {
  console.log("bridge.goReady", params);
});
async function initAsync2() {
  console.log("hello world from js");
  installJsonRpcReceiver();
  await initAsync();
}
initAsync2();
