class PhaseAccumulator {
    samples = new Map();
    add(name, durationMs) {
        const samples = this.samples.get(name);
        if (samples) {
            samples.push(durationMs);
            return;
        }
        this.samples.set(name, [durationMs]);
    }
    summaries() {
        return [...this.samples.entries()].map(([name, values]) => {
            const totalMs = values.reduce((sum, value) => sum + value, 0);
            return {
                name,
                iterations: values.length,
                totalMs,
                meanMs: totalMs / values.length,
                minMs: Math.min(...values),
                maxMs: Math.max(...values)
            };
        });
    }
}
export async function runGoJuniorWasmPocBenchmark(options = {}) {
    const iterations = positiveInteger(options.iterations, 1);
    const warmupIterations = nonNegativeInteger(options.warmupIterations, 0);
    const workItems = positiveInteger(options.workItems, 100_000);
    const fuel = positiveInteger(options.fuel, 8192);
    const totalIterations = iterations + warmupIterations;
    const phases = new PhaseAccumulator();
    const diagnostics = [];
    const expectedI64 = BigInt(workItems) * BigInt(workItems - 1) / 2n;
    let i64ModuleBytes = 0;
    let f64ModuleBytes = 0;
    let f64ChunkModuleBytes = 0;
    let f64MemoryBytes = 0;
    let chunkCountMax = 0;
    let jsNumberResult = 0;
    let jsBigIntResult = 0n;
    let wasmI64Result = 0n;
    let jsF64Result = 0;
    let wasmF64Result = 0;
    let wasmChunkedF64Result = 0;
    for (let index = 0; index < totalIterations; index += 1) {
        const collect = index >= warmupIterations;
        const i64Bytes = timed(phases, "wasm-generate-i64-module", collect, () => emitI64SumModule());
        const f64Bytes = timed(phases, "wasm-generate-f64-memory-module", collect, () => emitF64MemorySumModule(workItems));
        const f64ChunkBytes = timed(phases, "wasm-generate-f64-chunk-module", collect, () => emitF64ChunkedMemorySumModule(workItems));
        i64ModuleBytes = i64Bytes.length;
        f64ModuleBytes = f64Bytes.length;
        f64ChunkModuleBytes = f64ChunkBytes.length;
        const i64Module = timed(phases, "wasm-compile-i64-module", collect, () => new WebAssembly.Module(wasmBufferSource(i64Bytes)));
        const f64Module = timed(phases, "wasm-compile-f64-memory-module", collect, () => new WebAssembly.Module(wasmBufferSource(f64Bytes)));
        const f64ChunkModule = timed(phases, "wasm-compile-f64-chunk-module", collect, () => new WebAssembly.Module(wasmBufferSource(f64ChunkBytes)));
        const i64Instance = timed(phases, "wasm-instantiate-i64-module", collect, () => new WebAssembly.Instance(i64Module));
        const f64Instance = timed(phases, "wasm-instantiate-f64-memory-module", collect, () => new WebAssembly.Instance(f64Module));
        const f64ChunkInstance = timed(phases, "wasm-instantiate-f64-chunk-module", collect, () => new WebAssembly.Instance(f64ChunkModule));
        const f64Memory = f64Instance.exports.memory;
        const f64ChunkMemory = f64ChunkInstance.exports.memory;
        f64MemoryBytes = f64Memory.buffer.byteLength;
        const f64Input = makeF64Input(workItems);
        writeF64Input(f64Memory, f64Input);
        writeF64Input(f64ChunkMemory, f64Input);
        initializeChunkState(f64ChunkMemory);
        jsNumberResult = timed(phases, "js-number-loop", collect, () => jsNumberSum(workItems));
        jsBigIntResult = timed(phases, "js-bigint-loop", collect, () => jsBigIntSum(workItems));
        wasmI64Result = timed(phases, "wasm-i64-whole-loop", collect, () => i64Instance.exports.sum_i64(BigInt(workItems)));
        jsF64Result = timed(phases, "js-f64array-loop", collect, () => jsF64ArraySum(f64Input));
        wasmF64Result = timed(phases, "wasm-f64array-whole-loop", collect, () => f64Instance.exports.sum_f64(DATA_PTR, workItems));
        const chunked = timed(phases, "wasm-f64array-fuel-loop", collect, () => runChunkedF64Loop(f64ChunkInstance, workItems, fuel));
        wasmChunkedF64Result = chunked.sum;
        chunkCountMax = Math.max(chunkCountMax, chunked.chunks);
        if (collect) {
            if (jsBigIntResult !== expectedI64)
                diagnostics.push(`js bigint sum ${jsBigIntResult} did not match ${expectedI64}`);
            if (wasmI64Result !== expectedI64)
                diagnostics.push(`wasm i64 sum ${wasmI64Result} did not match ${expectedI64}`);
            if (jsNumberResult !== Number(expectedI64))
                diagnostics.push(`js number sum ${jsNumberResult} did not match ${Number(expectedI64)}`);
            if (!sameF64(jsF64Result, wasmF64Result))
                diagnostics.push(`wasm f64 sum ${wasmF64Result} did not match js ${jsF64Result}`);
            if (!sameF64(jsF64Result, wasmChunkedF64Result))
                diagnostics.push(`chunked wasm f64 sum ${wasmChunkedF64Result} did not match js ${jsF64Result}`);
        }
    }
    return {
        ok: diagnostics.length === 0,
        iterations,
        warmupIterations,
        workItems,
        fuel,
        phases: phases.summaries(),
        metrics: {
            i64ModuleBytes,
            f64ModuleBytes,
            f64ChunkModuleBytes,
            f64MemoryBytes,
            chunkCountMax,
            jsNumberResult,
            jsBigIntResult: jsBigIntResult.toString(),
            wasmI64Result: wasmI64Result.toString(),
            jsF64Result,
            wasmF64Result,
            wasmChunkedF64Result
        },
        diagnostics
    };
}
export function formatGoJuniorWasmPocReport(report) {
    const lines = [];
    lines.push(`gojr wasm-poc: iterations=${report.iterations} warmup=${report.warmupIterations} work_items=${report.workItems} fuel=${report.fuel} ok=${report.ok}`);
    lines.push(`  module_bytes: i64=${report.metrics.i64ModuleBytes} f64=${report.metrics.f64ModuleBytes} f64_chunk=${report.metrics.f64ChunkModuleBytes} memory_bytes=${report.metrics.f64MemoryBytes}`);
    lines.push(`  results: js_number=${report.metrics.jsNumberResult} js_bigint=${report.metrics.jsBigIntResult} wasm_i64=${report.metrics.wasmI64Result} js_f64=${report.metrics.jsF64Result} wasm_f64=${report.metrics.wasmF64Result} wasm_chunked_f64=${report.metrics.wasmChunkedF64Result} chunks=${report.metrics.chunkCountMax}`);
    for (const phase of report.phases) {
        lines.push(`  ${phase.name}: total_ms=${formatMs(phase.totalMs)} mean_ms=${formatMs(phase.meanMs)} min_ms=${formatMs(phase.minMs)} max_ms=${formatMs(phase.maxMs)} n=${phase.iterations}`);
    }
    for (const diagnostic of report.diagnostics) {
        lines.push(`  error: ${diagnostic}`);
    }
    return `${lines.join("\n")}\n`;
}
const PAGE_SIZE = 65536;
const DATA_PTR = 64;
const STATE_PTR = 0;
function emitI64SumModule() {
    const typeSection = vector([
        functionType([126 /* ValType.i64 */], [126 /* ValType.i64 */])
    ]);
    const functionSection = vector([u32(0)]);
    const exportSection = vector([
        exportEntry("sum_i64", 0 /* ExportKind.func */, 0)
    ]);
    const body = functionBody([localDecl(2, 126 /* ValType.i64 */)], [
        66 /* Op.i64Const */, ...sleb(0), 33 /* Op.localSet */, 1,
        66 /* Op.i64Const */, ...sleb(0), 33 /* Op.localSet */, 2,
        2 /* Op.block */, 64 /* BlockType.empty */,
        3 /* Op.loop */, 64 /* BlockType.empty */,
        32 /* Op.localGet */, 2, 32 /* Op.localGet */, 0, 90 /* Op.i64GeU */, 13 /* Op.brIf */, 1,
        32 /* Op.localGet */, 1, 32 /* Op.localGet */, 2, 124 /* Op.i64Add */, 33 /* Op.localSet */, 1,
        32 /* Op.localGet */, 2, 66 /* Op.i64Const */, ...sleb(1), 124 /* Op.i64Add */, 33 /* Op.localSet */, 2,
        12 /* Op.br */, 0,
        11 /* Op.end */,
        11 /* Op.end */,
        32 /* Op.localGet */, 1,
        11 /* Op.end */
    ]);
    return moduleBytes([
        section(1 /* Section.type */, typeSection),
        section(3 /* Section.func */, functionSection),
        section(7 /* Section.export */, exportSection),
        section(10 /* Section.code */, vector([body]))
    ]);
}
function emitF64MemorySumModule(length) {
    const pages = pagesForData(length);
    const typeSection = vector([
        functionType([127 /* ValType.i32 */, 127 /* ValType.i32 */], [124 /* ValType.f64 */])
    ]);
    const functionSection = vector([u32(0)]);
    const memorySection = vector([[0x00, ...u32(pages)]]);
    const exportSection = vector([
        exportEntry("memory", 2 /* ExportKind.memory */, 0),
        exportEntry("sum_f64", 0 /* ExportKind.func */, 0)
    ]);
    const body = functionBody([localDecl(1, 127 /* ValType.i32 */), localDecl(1, 124 /* ValType.f64 */)], [
        65 /* Op.i32Const */, ...sleb(0), 33 /* Op.localSet */, 2,
        68 /* Op.f64Const */, ...f64Bytes(0), 33 /* Op.localSet */, 3,
        2 /* Op.block */, 64 /* BlockType.empty */,
        3 /* Op.loop */, 64 /* BlockType.empty */,
        32 /* Op.localGet */, 2, 32 /* Op.localGet */, 1, 79 /* Op.i32GeU */, 13 /* Op.brIf */, 1,
        32 /* Op.localGet */, 3,
        32 /* Op.localGet */, 0, 32 /* Op.localGet */, 2, 65 /* Op.i32Const */, ...sleb(3), 116 /* Op.i32Shl */, 106 /* Op.i32Add */, 43 /* Op.f64Load */, ...memarg(3, 0),
        160 /* Op.f64Add */, 33 /* Op.localSet */, 3,
        32 /* Op.localGet */, 2, 65 /* Op.i32Const */, ...sleb(1), 106 /* Op.i32Add */, 33 /* Op.localSet */, 2,
        12 /* Op.br */, 0,
        11 /* Op.end */,
        11 /* Op.end */,
        32 /* Op.localGet */, 3,
        11 /* Op.end */
    ]);
    return moduleBytes([
        section(1 /* Section.type */, typeSection),
        section(3 /* Section.func */, functionSection),
        section(5 /* Section.memory */, memorySection),
        section(7 /* Section.export */, exportSection),
        section(10 /* Section.code */, vector([body]))
    ]);
}
function emitF64ChunkedMemorySumModule(length) {
    const pages = pagesForData(length);
    const typeSection = vector([
        functionType([127 /* ValType.i32 */, 127 /* ValType.i32 */, 127 /* ValType.i32 */, 127 /* ValType.i32 */], [127 /* ValType.i32 */])
    ]);
    const functionSection = vector([u32(0)]);
    const memorySection = vector([[0x00, ...u32(pages)]]);
    const exportSection = vector([
        exportEntry("memory", 2 /* ExportKind.memory */, 0),
        exportEntry("sum_f64_chunk", 0 /* ExportKind.func */, 0)
    ]);
    const body = functionBody([localDecl(2, 127 /* ValType.i32 */), localDecl(1, 124 /* ValType.f64 */)], [
        32 /* Op.localGet */, 0, 40 /* Op.i32Load */, ...memarg(2, 0), 33 /* Op.localSet */, 4,
        32 /* Op.localGet */, 0, 43 /* Op.f64Load */, ...memarg(3, 8), 33 /* Op.localSet */, 6,
        32 /* Op.localGet */, 3, 33 /* Op.localSet */, 5,
        2 /* Op.block */, 64 /* BlockType.empty */,
        3 /* Op.loop */, 64 /* BlockType.empty */,
        32 /* Op.localGet */, 5, 69 /* Op.i32Eqz */, 13 /* Op.brIf */, 1,
        32 /* Op.localGet */, 4, 32 /* Op.localGet */, 2, 79 /* Op.i32GeU */, 13 /* Op.brIf */, 1,
        32 /* Op.localGet */, 6,
        32 /* Op.localGet */, 1, 32 /* Op.localGet */, 4, 65 /* Op.i32Const */, ...sleb(3), 116 /* Op.i32Shl */, 106 /* Op.i32Add */, 43 /* Op.f64Load */, ...memarg(3, 0),
        160 /* Op.f64Add */, 33 /* Op.localSet */, 6,
        32 /* Op.localGet */, 4, 65 /* Op.i32Const */, ...sleb(1), 106 /* Op.i32Add */, 33 /* Op.localSet */, 4,
        32 /* Op.localGet */, 5, 65 /* Op.i32Const */, ...sleb(1), 107 /* Op.i32Sub */, 33 /* Op.localSet */, 5,
        12 /* Op.br */, 0,
        11 /* Op.end */,
        11 /* Op.end */,
        32 /* Op.localGet */, 0, 32 /* Op.localGet */, 4, 54 /* Op.i32Store */, ...memarg(2, 0),
        32 /* Op.localGet */, 0, 32 /* Op.localGet */, 6, 57 /* Op.f64Store */, ...memarg(3, 8),
        32 /* Op.localGet */, 4, 32 /* Op.localGet */, 2, 79 /* Op.i32GeU */,
        11 /* Op.end */
    ]);
    return moduleBytes([
        section(1 /* Section.type */, typeSection),
        section(3 /* Section.func */, functionSection),
        section(5 /* Section.memory */, memorySection),
        section(7 /* Section.export */, exportSection),
        section(10 /* Section.code */, vector([body]))
    ]);
}
function functionType(params, results) {
    return [0x60, ...vector(params), ...vector(results)];
}
function functionBody(locals, instructions) {
    return sized([...vector(locals), ...instructions]);
}
function localDecl(count, type) {
    return [...u32(count), type];
}
function exportEntry(name, kind, index) {
    return [...nameBytes(name), kind, ...u32(index)];
}
function section(id, data) {
    return [id, ...sized(data)];
}
function moduleBytes(sections) {
    return new Uint8Array([
        0x00, 0x61, 0x73, 0x6d,
        0x01, 0x00, 0x00, 0x00,
        ...sections.flat()
    ]);
}
function vector(items) {
    if (items.length === 0)
        return [0];
    if (typeof items[0] === "number")
        return [...u32(items.length), ...items];
    const nested = items;
    return [...u32(nested.length), ...nested.flat()];
}
function sized(data) {
    return [...u32(data.length), ...data];
}
function nameBytes(name) {
    const bytes = [...new TextEncoder().encode(name)];
    return [...u32(bytes.length), ...bytes];
}
function u32(value) {
    const out = [];
    let current = value >>> 0;
    do {
        let byte = current & 0x7f;
        current >>>= 7;
        if (current !== 0)
            byte |= 0x80;
        out.push(byte);
    } while (current !== 0);
    return out;
}
function sleb(value) {
    const out = [];
    let current = value | 0;
    let more = true;
    while (more) {
        let byte = current & 0x7f;
        current >>= 7;
        const signBit = (byte & 0x40) !== 0;
        more = !((current === 0 && !signBit) || (current === -1 && signBit));
        if (more)
            byte |= 0x80;
        out.push(byte);
    }
    return out;
}
function memarg(align, offset) {
    return [...u32(align), ...u32(offset)];
}
function f64Bytes(value) {
    const bytes = new Uint8Array(8);
    new DataView(bytes.buffer).setFloat64(0, value, true);
    return [...bytes];
}
function pagesForData(length) {
    return Math.max(1, Math.ceil((DATA_PTR + length * 8) / PAGE_SIZE));
}
function makeF64Input(length) {
    const values = new Float64Array(length);
    for (let index = 0; index < length; index += 1) {
        values[index] = (index % 17) + 0.25;
    }
    return values;
}
function writeF64Input(memory, values) {
    new Float64Array(memory.buffer, DATA_PTR, values.length).set(values);
}
function initializeChunkState(memory) {
    const view = new DataView(memory.buffer);
    view.setInt32(STATE_PTR, 0, true);
    view.setFloat64(STATE_PTR + 8, 0, true);
}
function runChunkedF64Loop(instance, length, fuel) {
    const step = instance.exports.sum_f64_chunk;
    let done = 0;
    let chunks = 0;
    while (done === 0) {
        done = step(STATE_PTR, DATA_PTR, length, fuel);
        chunks += 1;
    }
    const memory = instance.exports.memory;
    return {
        sum: new DataView(memory.buffer).getFloat64(STATE_PTR + 8, true),
        chunks
    };
}
function jsNumberSum(length) {
    let sum = 0;
    for (let index = 0; index < length; index += 1)
        sum += index;
    return sum;
}
function jsBigIntSum(length) {
    let sum = 0n;
    for (let index = 0n; index < BigInt(length); index += 1n)
        sum += index;
    return sum;
}
function jsF64ArraySum(values) {
    let sum = 0;
    for (let index = 0; index < values.length; index += 1)
        sum += values[index] ?? 0;
    return sum;
}
function sameF64(left, right) {
    return Object.is(left, right) || Math.abs(left - right) <= 1e-9 * Math.max(1, Math.abs(left), Math.abs(right));
}
function wasmBufferSource(bytes) {
    return bytes;
}
function timed(phases, name, collect, fn) {
    const start = nowMs();
    try {
        return fn();
    }
    finally {
        if (collect)
            phases.add(name, nowMs() - start);
    }
}
function nowMs() {
    return typeof performance !== "undefined" && typeof performance.now === "function"
        ? performance.now()
        : Date.now();
}
function formatMs(value) {
    if (!Number.isFinite(value))
        return "NaN";
    return value.toFixed(value >= 10 ? 2 : 3);
}
function positiveInteger(value, fallback) {
    const number = Math.trunc(Number(value));
    return Number.isFinite(number) && number > 0 ? number : fallback;
}
function nonNegativeInteger(value, fallback) {
    const number = Math.trunc(Number(value));
    return Number.isFinite(number) && number >= 0 ? number : fallback;
}
