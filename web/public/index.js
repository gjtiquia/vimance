// web/src/test-button.ts
function init() {
  document.body.addEventListener("click", async (event) => {
    const button = event.target;
    if (!button.matches("[data-test-button]"))
      return;
    console.log("button pressed");
    globalThis.goWasmJsonRpc?.call(JSON.stringify({ message: "Hello from the button!" }));
  });
}
init();

// web/src/wasm/exports.ts
var textDecoder = new TextDecoder;
function createExports() {
  return {
    notify: function(eventId) {
      if (!wasm)
        return;
      const slicePtr = wasm.exports.getCanvasCellsPtr();
      const sliceHeader = new Uint32Array(wasm.exports.memory.buffer, slicePtr, 3);
      const ptr = sliceHeader[0];
      const len = sliceHeader[1];
      const cap = sliceHeader[2];
    }
  };
}

// web/src/wasm/wasm.ts
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
async function initAsync2() {
  console.log("hello world from js");
  await initAsync();
}
initAsync2();
