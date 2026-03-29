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
// web/src/jsonrpc/request.ts
var requestIdCounter = 0;
function newRequest(method, params) {
  requestIdCounter++;
  return {
    jsonrpc: "2.0",
    method,
    params,
    id: requestIdCounter
  };
}
function decodeRequest(jsonString) {
  const obj = JSON.parse(jsonString);
  return obj;
}
// web/src/jsonrpc/response.ts
function newResponse(result, id) {
  return {
    jsonrpc: "2.0",
    result,
    id
  };
}
function newSuccessResponse(id) {
  return newResponse({ success: true }, id);
}
function decodeResponse(jsonString) {
  const obj = JSON.parse(jsonString);
  return obj;
}
function newMethodNotFoundResponse(request) {
  return newErrorResponse(-32601, `Method not found: ${request.method}`, request.id);
}
function newErrorResponse(code, message, id, data) {
  return {
    jsonrpc: "2.0",
    error: { code, message, data },
    id
  };
}
// web/src/wasm/rpc.ts
globalThis.goToJsJsonRpcAsync = onReceiveJsonRpcAsync;
async function sendRpcAsync(method, params) {
  const request2 = newRequest(method, params);
  const requestJson = JSON.stringify(request2);
  const responseJson = await globalThis.jsToGoJsonRpcAsync.call(requestJson);
  return decodeResponse(responseJson);
}
async function onReceiveJsonRpcAsync(jsonString) {
  const request2 = decodeRequest(jsonString);
  const { ok, responseJson } = tryHandleEngineEvents(request2);
  if (ok) {
    return responseJson;
  }
  switch (request2.method) {
    case "echo":
      return handleEcho(request2);
    default:
      const response2 = newMethodNotFoundResponse(request2);
      console.error(`js: ${response2.error?.message}`);
      return JSON.stringify(response2);
  }
}
async function handleEcho(request2) {
  const echoParams = request2.params;
  const response2 = newResponse({ message: `js echoooooo ${echoParams.message}` }, request2.id);
  const responseJson = JSON.stringify(response2);
  return responseJson;
}
function tryHandleEngineEvents(request2) {
  if (!request2.method.startsWith("engine.") || request2.method.split(".").length !== 2) {
    return { ok: false };
  }
  const eventName = request2.method.split(".")[1];
  document.body.dispatchEvent(new CustomEvent("engine:onEventTriggered", {
    detail: {
      eventName,
      params: request2.params
    }
  }));
  const response2 = newSuccessResponse(request2.id);
  return { ok: true, responseJson: JSON.stringify(response2) };
}
// web/src/test-button.ts
function init() {
  document.body.addEventListener("click", async (event) => {
    const button = event.target;
    if (!button.matches("[data-test-button]"))
      return;
    const response2 = await sendRpcAsync("echo", {
      message: "helloooooo from js"
    });
    if (response2.error) {
      console.error("js: echo.response.error:", response2.error);
      return;
    }
    console.log("js: echo.response.result.message:", response2.result.message);
  });
}
init();

// web/src/engine-debug-console.ts
function init2() {
  const elements = document.body.querySelectorAll("[data-engine-debug-console]");
  document.body.addEventListener("engine:onEventTriggered", async (event) => {
    const customEvent = event;
    elements.forEach((parent) => {
      const eventName = customEvent.detail.eventName;
      const params = customEvent.detail.params;
      const child = document.createElement("p");
      child.textContent = `${eventName}: ${JSON.stringify(params)}`;
      parent.appendChild(child);
    });
  });
}
init2();

// web/src/engine/input.ts
function subscribeToKeyDownEvent() {
  document.addEventListener("keydown", (e) => {
    sendRpcAsync("keydown", {
      key: e.key
    });
  });
}

// web/src/engine/index.ts
function init3() {
  subscribeToKeyDownEvent();
}

// web/src/index.ts
async function initAsync2() {
  console.log("js: running...");
  init3();
  await initAsync();
}
initAsync2();
