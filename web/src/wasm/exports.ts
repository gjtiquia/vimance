import { wasm } from "./wasm";

const textDecoder = new TextDecoder(); // optimization: cached for reuse

// functions exported to Go
export function createExports() {
    return {
        notify: function (eventId: number) {
            if (!wasm) return;

            // ptr returns the address of the Go slice header, not the byte data.
            // Slice header is [ptr: 4 bytes, len: 4 bytes, cap: 4 bytes].
            const slicePtr = wasm.exports.getCanvasCellsPtr();
            const sliceHeader = new Uint32Array(
                wasm.exports.memory.buffer,
                slicePtr,
                3,
            );

            const ptr = sliceHeader[0];
            const len = sliceHeader[1];
            const cap = sliceHeader[2];
        },
    };
}
