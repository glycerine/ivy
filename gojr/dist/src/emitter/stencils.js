import { extractWasmFunction } from "./wasm/module.js";
const I64_SCALAR_KERNELS_BASE64 = "AGFzbQEAAAABDQJgAn5+AX5gAn5+AX8DEhEAAAAAAAAAAAAAAAEBAQEBAQf+AREMZ29qcl9pNjRfYWRkAAAMZ29qcl9pNjRfc3ViAAEMZ29qcl9pNjRfbXVsAAIMZ29qcl9pNjRfZGl2AAMMZ29qcl9pNjRfcmVtAAQMZ29qcl9pNjRfYW5kAAULZ29qcl9pNjRfb3IABgxnb2pyX2k2NF94b3IABxFnb2pyX2k2NF9iaXRjbGVhcgAIDGdvanJfaTY0X3NobAAJDGdvanJfaTY0X3NocgAKC2dvanJfaTY0X2VxAAsLZ29qcl9pNjRfbmUADAtnb2pyX2k2NF9sdAANC2dvanJfaTY0X2xlAA4LZ29qcl9pNjRfZ3QADwtnb2pyX2k2NF9nZQAQCowBEQcAIAAgAXwLBwAgACABfQsHACAAIAF+CwcAIAAgAX8LBwAgACABgQsHACAAIAGDCwcAIAAgAYQLBwAgACABhQsKACAAIAFCf4WDCwcAIAAgAYYLBwAgACABhwsHACAAIAFRCwcAIAAgAVILBwAgACABUwsHACAAIAFXCwcAIAAgAVULBwAgACABWQs=";
const i64ScalarStencilSpecs = [
    ["i64.scalar.add", "gojr_i64_add", "a + b"],
    ["i64.scalar.sub", "gojr_i64_sub", "a - b"],
    ["i64.scalar.mul", "gojr_i64_mul", "a * b"],
    ["i64.scalar.div", "gojr_i64_div", "a / b"],
    ["i64.scalar.rem", "gojr_i64_rem", "a % b"],
    ["i64.scalar.and", "gojr_i64_and", "a & b"],
    ["i64.scalar.or", "gojr_i64_or", "a | b"],
    ["i64.scalar.xor", "gojr_i64_xor", "a ^ b"],
    ["i64.scalar.bitclear", "gojr_i64_bitclear", "a &^ b"],
    ["i64.scalar.shl", "gojr_i64_shl", "a << b"],
    ["i64.scalar.shr", "gojr_i64_shr", "a >> b"],
    ["i64.scalar.eq", "gojr_i64_eq", "a == b"],
    ["i64.scalar.ne", "gojr_i64_ne", "a != b"],
    ["i64.scalar.lt", "gojr_i64_lt", "a < b"],
    ["i64.scalar.le", "gojr_i64_le", "a <= b"],
    ["i64.scalar.gt", "gojr_i64_gt", "a > b"],
    ["i64.scalar.ge", "gojr_i64_ge", "a >= b"]
];
const i64ScalarStencilEntries = i64ScalarStencilSpecs.map(([name, exportName, operation]) => ({
    name,
    exportName,
    sourceLanguage: "c",
    importMemory: false,
    description: `Package-lowerable scalar int64 kernel for ${operation}.`,
    cSource: `// Extracted from the checked-in scalar kernel bundle: ${operation}\n`,
    wasmBase64: I64_SCALAR_KERNELS_BASE64
}));
export const CHECKED_IN_WASM_STENCILS = [
    {
        name: "i64.add.kernel",
        exportName: "add_i64",
        sourceLanguage: "c",
        importMemory: false,
        description: "Small scalar i64 arithmetic stencil used to validate Clang-to-Wasm extraction.",
        cSource: "__attribute__((visibility(\"default\"))) long long add_i64(long long a, long long b) { return a + b; }\n",
        wasmBase64: "AGFzbQEAAAABBwFgAn5+AX4DAgEABwsBB2FkZF9pNjQAAAoJAQcAIAAgAXwLAC8JcHJvZHVjZXJzAQxwcm9jZXNzZWQtYnkBDkhvbWVicmV3IGNsYW5nBjIxLjEuNACUAQ90YXJnZXRfZmVhdHVyZXMIKw9tdXRhYmxlLWdsb2JhbHMrE25vbnRyYXBwaW5nLWZwdG9pbnQrC2J1bGstbWVtb3J5KwhzaWduLWV4dCsPcmVmZXJlbmNlLXR5cGVzKwptdWx0aXZhbHVlKw9idWxrLW1lbW9yeS1vcHQrFmNhbGwtaW5kaXJlY3Qtb3Zlcmxvbmc="
    },
    ...i64ScalarStencilEntries,
    {
        name: "f64.slice.sum.kernel",
        exportName: "sum_f64",
        sourceLanguage: "c",
        importMemory: true,
        description: "Dense f64 memory loop stencil over JS/GoJr-owned WebAssembly.Memory.",
        cSource: "double sum_f64(double *xs, int n) { double s = 0; for (int i = 0; i < n; i++) { s += xs[i]; } return s; }\n",
        wasmBase64: "AGFzbQEAAAABBwFgAn9/AXwCDwEDZW52Bm1lbW9yeQIAAgMCAQAHCwEHc3VtX2Y2NAAACqcBAaQBAgF8A38gAUEATARARAAAAAAAAAAADwsgAUEDcSEEAkAgAUEESQRADAELIAFB/P///wdxIQUgACEBA0AgAiABKwMAoCABQQhqKwMAoCABQRBqKwMAoCABQRhqKwMAoCECIAFBIGohASAFIANBBGoiA0cNAAsLIAQEQCAAIANBA3RqIQEDQCACIAErAwCgIQIgAUEIaiEBIARBAWsiBA0ACwsgAgsALwlwcm9kdWNlcnMBDHByb2Nlc3NlZC1ieQEOSG9tZWJyZXcgY2xhbmcGMjEuMS40AJQBD3RhcmdldF9mZWF0dXJlcwgrD211dGFibGUtZ2xvYmFscysTbm9udHJhcHBpbmctZnB0b2ludCsLYnVsay1tZW1vcnkrCHNpZ24tZXh0Kw9yZWZlcmVuY2UtdHlwZXMrCm11bHRpdmFsdWUrD2J1bGstbWVtb3J5LW9wdCsWY2FsbC1pbmRpcmVjdC1vdmVybG9uZw=="
    },
    {
        name: "f64.slice.sum.chunk.kernel",
        exportName: "sum_f64_chunk",
        sourceLanguage: "c",
        importMemory: true,
        description: "Fuel-bounded dense f64 memory loop stencil with JS-owned preemption state.",
        cSource: [
            "struct SumF64State {",
            "  int index;",
            "  double sum;",
            "};",
            "",
            "int sum_f64_chunk(struct SumF64State *state, double *xs, int n, int fuel) {",
            "  int i = state->index;",
            "  double sum = state->sum;",
            "  int end = i + fuel;",
            "  if (end > n) end = n;",
            "  for (; i < end; i++) {",
            "    sum += xs[i];",
            "  }",
            "  state->index = i;",
            "  state->sum = sum;",
            "  return i >= n;",
            "}",
            ""
        ].join("\n"),
        wasmBase64: "AGFzbQEAAAABCQFgBH9/f38BfwIPAQNlbnYGbWVtb3J5AgACAwIBAAcRAQ1zdW1fZjY0X2NodW5rAAAK4AEB3QECBH8BfCAAKwMIIQggACgCACIEIAMgBGoiAyACIAIgA0obIgdIBEACQCAHIARrQQNxIgVFBEAgBCEGDAELIAEgBEEDdGohAyAEIQYDQCAGQQFqIQYgCCADKwMAoCEIIANBCGohAyAFQQFrIgUNAAsLIAQgB2tBfE0EQCAHIAZrIQUgASAGQQN0aiEDA0AgCCADKwMAoCADQQhqKwMAoCADQRBqKwMAoCADQRhqKwMAoCEIIANBIGohAyAFQQRrIgUNAAsLIAchBAsgACAIOQMIIAAgBDYCACACIARMCwAvCXByb2R1Y2VycwEMcHJvY2Vzc2VkLWJ5AQ5Ib21lYnJldyBjbGFuZwYyMS4xLjQAlAEPdGFyZ2V0X2ZlYXR1cmVzCCsPbXV0YWJsZS1nbG9iYWxzKxNub250cmFwcGluZy1mcHRvaW50KwtidWxrLW1lbW9yeSsIc2lnbi1leHQrD3JlZmVyZW5jZS10eXBlcysKbXVsdGl2YWx1ZSsPYnVsay1tZW1vcnktb3B0KxZjYWxsLWluZGlyZWN0LW92ZXJsb25n"
    }
];
export function checkedInWasmStencil(name) {
    const stencil = CHECKED_IN_WASM_STENCILS.find((item) => item.name === name || item.exportName === name);
    if (!stencil)
        throw new Error(`unknown GoJr Wasm stencil ${name}`);
    const wasmBytes = decodeBase64(stencil.wasmBase64);
    return {
        ...stencil,
        wasmBytes,
        extracted: extractWasmFunction(wasmBytes, stencil.exportName, {
            allowImportedMemory: stencil.importMemory
        })
    };
}
export function decodeBase64(base64) {
    if (typeof Buffer !== "undefined")
        return new Uint8Array(Buffer.from(base64, "base64"));
    const atobFn = globalThis.atob;
    if (!atobFn)
        throw new Error("base64 decoding is not available in this JavaScript environment");
    const binary = atobFn(base64);
    const out = new Uint8Array(binary.length);
    for (let index = 0; index < binary.length; index += 1)
        out[index] = binary.charCodeAt(index);
    return out;
}
export function encodeBase64(bytes) {
    if (typeof Buffer !== "undefined")
        return Buffer.from(bytes).toString("base64");
    let binary = "";
    for (const byte of bytes)
        binary += String.fromCharCode(byte);
    const btoaFn = globalThis.btoa;
    if (!btoaFn)
        throw new Error("base64 encoding is not available in this JavaScript environment");
    return btoaFn(binary);
}
