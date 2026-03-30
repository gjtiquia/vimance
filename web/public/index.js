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
globalThis.goEngineEventsSync = (json) => {
  const events = JSON.parse(json);
  for (const e of events) {
    const parts = e.method.split(".");
    const eventName = parts.length >= 2 ? parts[1] : e.method;
    document.body.dispatchEvent(new CustomEvent("engine:onEventTriggered", {
      detail: {
        eventName,
        params: e.params
      }
    }));
  }
};
function sendRpcSync(method, params) {
  const request2 = newRequest(method, params);
  const requestJson = JSON.stringify(request2);
  const fn = globalThis.jsToGoJsonRpcSync;
  if (typeof fn !== "function") {
    throw new Error("jsToGoJsonRpcSync is not registered (WASM not loaded?)");
  }
  const responseJson = fn.call(requestJson);
  return decodeResponse(responseJson);
}
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
init2();
function init2() {
  const elements = document.body.querySelectorAll("[data-engine-debug-console]");
  document.body.addEventListener("engine:onEventTriggered", async (event) => {
    const customEvent = event;
    elements.forEach((element) => {
      const container = element;
      const eventName = customEvent.detail.eventName;
      const params = customEvent.detail.params;
      const log = document.createElement("p");
      log.textContent = `${eventName}: ${JSON.stringify(params)}`;
      const stickToBottom = isScrolledToBottom(container);
      container.appendChild(log);
      if (stickToBottom) {
        container.scrollTop = container.scrollHeight;
      }
    });
  });
}
var BOTTOM_THRESHOLD_PX = 4;
function isScrolledToBottom(el) {
  return el.scrollHeight - el.scrollTop - el.clientHeight <= BOTTOM_THRESHOLD_PX;
}

// web/src/table.ts
init3();
function init3() {
  const tables = document.body.querySelectorAll("[data-table]");
  document.body.addEventListener("engine:onEventTriggered", async (event) => {
    const customEvent = event;
    const eventName = customEvent.detail.eventName;
    const params = customEvent.detail.params;
    const handler = getEventHandler(eventName);
    if (!handler) {
      console.warn(`js: table: No handler found for event: ${eventName}`);
      return;
    }
    tables.forEach((element) => {
      const table = element;
      handler(table, params);
    });
  });
}
function getEventHandler(eventName) {
  switch (eventName) {
    case "OnModeChanged":
      return handleOnModeChanged;
    case "OnCursorMoved":
      return handleOnCursorMoved;
    default:
      return null;
  }
}
function getCellDisplayValue(cell) {
  const input = cell.querySelector("input");
  if (input) {
    return input.value;
  }
  return cell.textContent?.trim() ?? "";
}
function handleOnModeChanged(table, params) {
  console.log("js: table: handleOnModeChanged:", params);
  const mode = params.mode;
  if (mode === "i") {
    const normalCell = table.querySelector("[data-cell-variant='normal']");
    if (!normalCell) {
      console.error("js: table: handleOnModeChanged: No normal cell found!");
      return;
    }
    const value = normalCell.textContent?.trim() || "";
    const newCell = replaceCell(table, normalCell, "input", value);
    if (!newCell) {
      return;
    }
    const input = newCell.querySelector("input");
    if (!input) {
      return;
    }
    input.focus();
    const insertPosition = params.insertPosition;
    if (insertPosition === "after") {
      const len = input.value.length;
      input.setSelectionRange(len, len);
    } else if (insertPosition === "highlight") {
      const len = input.value.length;
      input.setSelectionRange(0, len);
    } else {
      input.setSelectionRange(0, 0);
    }
  } else if (mode === "n") {
    const inputCell = table.querySelector("[data-cell-variant='input']");
    if (!inputCell) {
      return;
    }
    const input = inputCell.querySelector("input");
    const value = input?.value?.trim() ?? inputCell.textContent?.trim() ?? "";
    replaceCell(table, inputCell, "normal", value);
  }
}
function handleOnCursorMoved(table, params) {
  const x = params.x;
  const y = params.y;
  const targetCell = table.querySelector(`td[data-cell-x="${x}"][data-cell-y="${y}"]`);
  if (!targetCell) {
    return;
  }
  const normalCell = table.querySelector("[data-cell-variant='normal']");
  if (!normalCell) {
    return;
  }
  if (normalCell === targetCell) {
    return;
  }
  const fromValue = getCellDisplayValue(normalCell);
  const toValue = getCellDisplayValue(targetCell);
  replaceCell(table, normalCell, "default", fromValue);
  const newTarget = table.querySelector(`td[data-cell-x="${x}"][data-cell-y="${y}"]`);
  if (!newTarget) {
    return;
  }
  replaceCell(table, newTarget, "normal", toValue);
}
function replaceCell(table, oldCell, variant, value) {
  const template = table.querySelector(`template[data-cell-template="${variant}"]`);
  if (!template) {
    console.error(`js: table: replaceCell: No template found for variant: ${variant}`);
    return null;
  }
  const newCell = template.content.firstElementChild.cloneNode(true);
  const x = oldCell.getAttribute("data-cell-x");
  const y = oldCell.getAttribute("data-cell-y");
  if (x !== null)
    newCell.setAttribute("data-cell-x", x);
  if (y !== null)
    newCell.setAttribute("data-cell-y", y);
  const input = newCell.querySelector("input");
  if (input) {
    input.value = value;
  } else {
    newCell.textContent = value;
  }
  oldCell.replaceWith(newCell);
  return newCell;
}

// web/src/engine/input.ts
function dispatchKeydownEvents(result) {
  if (!result || typeof result !== "object") {
    return;
  }
  const r = result;
  if (!Array.isArray(r.events)) {
    return;
  }
  for (const e of r.events) {
    const parts = e.method.split(".");
    const eventName = parts.length >= 2 ? parts[1] : e.method;
    document.body.dispatchEvent(new CustomEvent("engine:onEventTriggered", {
      detail: {
        eventName,
        params: e.params
      }
    }));
  }
}
function subscribeToKeyDownEvent() {
  document.addEventListener("keydown", (e) => {
    const response2 = sendRpcSync("keydown", {
      key: e.key
    });
    if (response2.error) {
      return;
    }
    const result = response2.result;
    if (result?.captured) {
      e.preventDefault();
    }
    dispatchKeydownEvents(response2.result);
  });
}
var DOUBLE_TAP_MS = 300;
var GHOST_CLICK_AFTER_TOUCH_MS = 450;
var lastTouchCellKey = null;
var lastTouchTime = 0;
var lastTouchOnGridCellMs = 0;
function getCellCoordsFromEventTarget(target) {
  if (!target || !(target instanceof Element)) {
    return null;
  }
  const cell = target.closest("td[data-cell-x][data-cell-y]");
  if (!cell) {
    return null;
  }
  const xs = cell.getAttribute("data-cell-x");
  const ys = cell.getAttribute("data-cell-y");
  if (xs === null || ys === null) {
    return null;
  }
  const x = parseInt(xs, 10);
  const y = parseInt(ys, 10);
  if (Number.isNaN(x) || Number.isNaN(y)) {
    return null;
  }
  return { x, y };
}
function isGhostMouseEventAfterTouch() {
  return Date.now() - lastTouchOnGridCellMs < GHOST_CLICK_AFTER_TOUCH_MS;
}
function subscribeToPointerEvents() {
  document.addEventListener("click", (e) => {
    const coords = getCellCoordsFromEventTarget(e.target);
    if (!coords) {
      return;
    }
    if (isGhostMouseEventAfterTouch()) {
      e.preventDefault();
      return;
    }
    sendRpcAsync("setCursor", coords);
  });
  document.addEventListener("dblclick", (e) => {
    const coords = getCellCoordsFromEventTarget(e.target);
    if (!coords) {
      return;
    }
    if (isGhostMouseEventAfterTouch()) {
      e.preventDefault();
      return;
    }
    e.preventDefault();
    sendRpcAsync("setCursorAndEdit", coords);
  });
  document.addEventListener("touchend", (e) => {
    const coords = getCellCoordsFromEventTarget(e.target);
    if (!coords) {
      return;
    }
    lastTouchOnGridCellMs = Date.now();
    const key = `${coords.x},${coords.y}`;
    const now = lastTouchOnGridCellMs;
    const isDoubleTap = key === lastTouchCellKey && now - lastTouchTime < DOUBLE_TAP_MS;
    lastTouchCellKey = key;
    lastTouchTime = now;
    if (isDoubleTap) {
      sendRpcAsync("setCursorAndEdit", coords);
    } else {
      sendRpcAsync("setCursor", coords);
    }
  });
}

// web/src/engine/index.ts
function init4() {
  subscribeToKeyDownEvent();
  subscribeToPointerEvents();
}

// web/src/index.ts
async function initAsync2() {
  console.log("js: running...");
  init4();
  await initAsync();
}
initAsync2();
