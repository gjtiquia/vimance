// web/src/test-button.ts
function init() {
  document.body.addEventListener("click", async (event) => {
    const button = event.target;
    if (!button.matches("[data-test-button]"))
      return;
    console.log("button pressed");
    const response = await sendRpcAsync("echo", "helloooooo from js");
    console.log("awaited response from go wasm:", response);
  });
}
init();
var requestIdCounter = 0;
async function sendRpcAsync(method, params) {
  requestIdCounter++;
  return globalThis.jsToGoJsonRpcAsync.call(JSON.stringify({
    jsonrpc: "2.0",
    method,
    params,
    id: requestIdCounter
  }));
}
async function onReceiveJsonRpcAsync(message) {
  console.log("received json rpc from go wasm:", message);
  return "test response from js";
}
Object.defineProperty(globalThis, "goToJsJsonRpcAsync", {
  value: onReceiveJsonRpcAsync,
  writable: false,
  configurable: false,
  enumerable: false
});

// web/src/wasm/exports.ts
function createExports() {
  return {};
}

// web/src/wasm/wasm.ts
var WASM_URL = "/public/main.wasm";
var wasm = undefined;
async function initAsync() {
  const go = new Go;
  go.importObject.env = createExports();
  if (!WebAssembly.instantiateStreaming) {
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer();
      return await WebAssembly.instantiate(source, importObject);
    };
  }
  try {
    const result = await WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject);
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
async function initAsync2() {
  console.log("hello world from js");
  await initAsync();
}
initAsync2();
