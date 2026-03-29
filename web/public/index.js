// web/src/test-button.ts
function init() {
  document.body.addEventListener("click", async (event) => {
    const button = event.target;
    if (!button.matches("[data-test-button]"))
      return;
    const response = await sendRpcAsync("echo", {
      message: "helloooooo from js"
    });
    if (response.error) {
      console.error("js: echo.response.error.message:", response.error.message);
      return;
    }
    console.log("js: echo.response.result.message:", response.result.message);
  });
}
init();
async function sendRpcAsync(method, params) {
  const request = newJsonRpcRequest(method, params);
  const requestJson = JSON.stringify(request);
  const responseJson = await globalThis.jsToGoJsonRpcAsync.call(requestJson);
  return decodeJsonRpcResponse(responseJson);
}
globalThis.goToJsJsonRpcAsync = onReceiveJsonRpcAsync;
async function onReceiveJsonRpcAsync(jsonString) {
  const request = decodeJsonRpcRequest(jsonString);
  const echoParams = request.params;
  console.log(`js: ${request.method}.request.params.message:`, echoParams.message);
  const response = newJsonRpcResponse({ message: `js echoooooo ${echoParams.message}` }, request.id);
  const responseJson = JSON.stringify(response);
  return responseJson;
}
var requestIdCounter = 0;
function newJsonRpcRequest(method, params) {
  requestIdCounter++;
  return {
    jsonrpc: "2.0",
    method,
    params,
    id: requestIdCounter
  };
}
function decodeJsonRpcRequest(jsonString) {
  const obj = JSON.parse(jsonString);
  return obj;
}
function newJsonRpcResponse(result, id) {
  return {
    jsonrpc: "2.0",
    result,
    id
  };
}
function decodeJsonRpcResponse(jsonString) {
  const obj = JSON.parse(jsonString);
  return obj;
}

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
    console.log("js: running main.wasm...");
    const exitCode = await go.run(wasm);
    console.log("js: main.wasm exit code:", exitCode);
  } catch (err) {
    console.error("js: wasm.initAsync: error");
    console.error(err);
  }
}
// web/src/index.ts
async function initAsync2() {
  console.log("js: running...");
  await initAsync();
}
initAsync2();
