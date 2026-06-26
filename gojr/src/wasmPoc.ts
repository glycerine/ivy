import type { GoJuniorBenchmarkPhaseSummary } from "./bench.js";

export interface GoJuniorWasmPocOptions {
  iterations?: number;
  warmupIterations?: number;
  workItems?: number;
  fuel?: number;
}

export interface GoJuniorWasmPocReport {
  ok: boolean;
  iterations: number;
  warmupIterations: number;
  workItems: number;
  fuel: number;
  phases: GoJuniorBenchmarkPhaseSummary[];
  metrics: GoJuniorWasmPocMetrics;
  diagnostics: string[];
}

export interface GoJuniorWasmPocMetrics {
  i64ModuleBytes: number;
  f64ModuleBytes: number;
  f64ChunkModuleBytes: number;
  f64MemoryBytes: number;
  chunkCountMax: number;
  jsNumberResult: number;
  jsBigIntResult: string;
  wasmI64Result: string;
  jsF64Result: number;
  wasmF64Result: number;
  wasmChunkedF64Result: number;
}

class PhaseAccumulator {
  private readonly samples = new Map<string, number[]>();

  public add(name: string, durationMs: number): void {
    const samples = this.samples.get(name);
    if (samples) {
      samples.push(durationMs);
      return;
    }
    this.samples.set(name, [durationMs]);
  }

  public summaries(): GoJuniorBenchmarkPhaseSummary[] {
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

export async function runGoJuniorWasmPocBenchmark(options: GoJuniorWasmPocOptions = {}): Promise<GoJuniorWasmPocReport> {
  const iterations = positiveInteger(options.iterations, 1);
  const warmupIterations = nonNegativeInteger(options.warmupIterations, 0);
  const workItems = positiveInteger(options.workItems, 100_000);
  const fuel = positiveInteger(options.fuel, 8192);
  const totalIterations = iterations + warmupIterations;
  const phases = new PhaseAccumulator();
  const diagnostics: string[] = [];
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

    const f64Memory = f64Instance.exports.memory as WebAssembly.Memory;
    const f64ChunkMemory = f64ChunkInstance.exports.memory as WebAssembly.Memory;
    f64MemoryBytes = f64Memory.buffer.byteLength;
    const f64Input = makeF64Input(workItems);
    writeF64Input(f64Memory, f64Input);
    writeF64Input(f64ChunkMemory, f64Input);
    initializeChunkState(f64ChunkMemory);

    jsNumberResult = timed(phases, "js-number-loop", collect, () => jsNumberSum(workItems));
    jsBigIntResult = timed(phases, "js-bigint-loop", collect, () => jsBigIntSum(workItems));
    wasmI64Result = timed(phases, "wasm-i64-whole-loop", collect, () =>
      (i64Instance.exports.sum_i64 as (n: bigint) => bigint)(BigInt(workItems)));
    jsF64Result = timed(phases, "js-f64array-loop", collect, () => jsF64ArraySum(f64Input));
    wasmF64Result = timed(phases, "wasm-f64array-whole-loop", collect, () =>
      (f64Instance.exports.sum_f64 as (ptr: number, length: number) => number)(DATA_PTR, workItems));
    const chunked = timed(phases, "wasm-f64array-fuel-loop", collect, () =>
      runChunkedF64Loop(f64ChunkInstance, workItems, fuel));
    wasmChunkedF64Result = chunked.sum;
    chunkCountMax = Math.max(chunkCountMax, chunked.chunks);

    if (collect) {
      if (jsBigIntResult !== expectedI64) diagnostics.push(`js bigint sum ${jsBigIntResult} did not match ${expectedI64}`);
      if (wasmI64Result !== expectedI64) diagnostics.push(`wasm i64 sum ${wasmI64Result} did not match ${expectedI64}`);
      if (jsNumberResult !== Number(expectedI64)) diagnostics.push(`js number sum ${jsNumberResult} did not match ${Number(expectedI64)}`);
      if (!sameF64(jsF64Result, wasmF64Result)) diagnostics.push(`wasm f64 sum ${wasmF64Result} did not match js ${jsF64Result}`);
      if (!sameF64(jsF64Result, wasmChunkedF64Result)) diagnostics.push(`chunked wasm f64 sum ${wasmChunkedF64Result} did not match js ${jsF64Result}`);
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

export function formatGoJuniorWasmPocReport(report: GoJuniorWasmPocReport): string {
  const lines: string[] = [];
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

function emitI64SumModule(): Uint8Array {
  const typeSection = vector([
    functionType([ValType.i64], [ValType.i64])
  ]);
  const functionSection = vector([u32(0)]);
  const exportSection = vector([
    exportEntry("sum_i64", ExportKind.func, 0)
  ]);
  const body = functionBody(
    [localDecl(2, ValType.i64)],
    [
      Op.i64Const, ...sleb(0), Op.localSet, 1,
      Op.i64Const, ...sleb(0), Op.localSet, 2,
      Op.block, BlockType.empty,
      Op.loop, BlockType.empty,
      Op.localGet, 2, Op.localGet, 0, Op.i64GeU, Op.brIf, 1,
      Op.localGet, 1, Op.localGet, 2, Op.i64Add, Op.localSet, 1,
      Op.localGet, 2, Op.i64Const, ...sleb(1), Op.i64Add, Op.localSet, 2,
      Op.br, 0,
      Op.end,
      Op.end,
      Op.localGet, 1,
      Op.end
    ]
  );
  return moduleBytes([
    section(Section.type, typeSection),
    section(Section.func, functionSection),
    section(Section.export, exportSection),
    section(Section.code, vector([body]))
  ]);
}

function emitF64MemorySumModule(length: number): Uint8Array {
  const pages = pagesForData(length);
  const typeSection = vector([
    functionType([ValType.i32, ValType.i32], [ValType.f64])
  ]);
  const functionSection = vector([u32(0)]);
  const memorySection = vector([[0x00, ...u32(pages)]]);
  const exportSection = vector([
    exportEntry("memory", ExportKind.memory, 0),
    exportEntry("sum_f64", ExportKind.func, 0)
  ]);
  const body = functionBody(
    [localDecl(1, ValType.i32), localDecl(1, ValType.f64)],
    [
      Op.i32Const, ...sleb(0), Op.localSet, 2,
      Op.f64Const, ...f64Bytes(0), Op.localSet, 3,
      Op.block, BlockType.empty,
      Op.loop, BlockType.empty,
      Op.localGet, 2, Op.localGet, 1, Op.i32GeU, Op.brIf, 1,
      Op.localGet, 3,
      Op.localGet, 0, Op.localGet, 2, Op.i32Const, ...sleb(3), Op.i32Shl, Op.i32Add, Op.f64Load, ...memarg(3, 0),
      Op.f64Add, Op.localSet, 3,
      Op.localGet, 2, Op.i32Const, ...sleb(1), Op.i32Add, Op.localSet, 2,
      Op.br, 0,
      Op.end,
      Op.end,
      Op.localGet, 3,
      Op.end
    ]
  );
  return moduleBytes([
    section(Section.type, typeSection),
    section(Section.func, functionSection),
    section(Section.memory, memorySection),
    section(Section.export, exportSection),
    section(Section.code, vector([body]))
  ]);
}

function emitF64ChunkedMemorySumModule(length: number): Uint8Array {
  const pages = pagesForData(length);
  const typeSection = vector([
    functionType([ValType.i32, ValType.i32, ValType.i32, ValType.i32], [ValType.i32])
  ]);
  const functionSection = vector([u32(0)]);
  const memorySection = vector([[0x00, ...u32(pages)]]);
  const exportSection = vector([
    exportEntry("memory", ExportKind.memory, 0),
    exportEntry("sum_f64_chunk", ExportKind.func, 0)
  ]);
  const body = functionBody(
    [localDecl(2, ValType.i32), localDecl(1, ValType.f64)],
    [
      Op.localGet, 0, Op.i32Load, ...memarg(2, 0), Op.localSet, 4,
      Op.localGet, 0, Op.f64Load, ...memarg(3, 8), Op.localSet, 6,
      Op.localGet, 3, Op.localSet, 5,
      Op.block, BlockType.empty,
      Op.loop, BlockType.empty,
      Op.localGet, 5, Op.i32Eqz, Op.brIf, 1,
      Op.localGet, 4, Op.localGet, 2, Op.i32GeU, Op.brIf, 1,
      Op.localGet, 6,
      Op.localGet, 1, Op.localGet, 4, Op.i32Const, ...sleb(3), Op.i32Shl, Op.i32Add, Op.f64Load, ...memarg(3, 0),
      Op.f64Add, Op.localSet, 6,
      Op.localGet, 4, Op.i32Const, ...sleb(1), Op.i32Add, Op.localSet, 4,
      Op.localGet, 5, Op.i32Const, ...sleb(1), Op.i32Sub, Op.localSet, 5,
      Op.br, 0,
      Op.end,
      Op.end,
      Op.localGet, 0, Op.localGet, 4, Op.i32Store, ...memarg(2, 0),
      Op.localGet, 0, Op.localGet, 6, Op.f64Store, ...memarg(3, 8),
      Op.localGet, 4, Op.localGet, 2, Op.i32GeU,
      Op.end
    ]
  );
  return moduleBytes([
    section(Section.type, typeSection),
    section(Section.func, functionSection),
    section(Section.memory, memorySection),
    section(Section.export, exportSection),
    section(Section.code, vector([body]))
  ]);
}

function functionType(params: number[], results: number[]): number[] {
  return [0x60, ...vector(params), ...vector(results)];
}

function functionBody(locals: number[][], instructions: number[]): number[] {
  return sized([...vector(locals), ...instructions]);
}

function localDecl(count: number, type: number): number[] {
  return [...u32(count), type];
}

function exportEntry(name: string, kind: number, index: number): number[] {
  return [...nameBytes(name), kind, ...u32(index)];
}

function section(id: number, data: number[]): number[] {
  return [id, ...sized(data)];
}

function moduleBytes(sections: number[][]): Uint8Array {
  return new Uint8Array([
    0x00, 0x61, 0x73, 0x6d,
    0x01, 0x00, 0x00, 0x00,
    ...sections.flat()
  ]);
}

function vector(items: number[][] | number[]): number[] {
  if (items.length === 0) return [0];
  if (typeof items[0] === "number") return [...u32(items.length), ...(items as number[])];
  const nested = items as number[][];
  return [...u32(nested.length), ...nested.flat()];
}

function sized(data: number[]): number[] {
  return [...u32(data.length), ...data];
}

function nameBytes(name: string): number[] {
  const bytes = [...new TextEncoder().encode(name)];
  return [...u32(bytes.length), ...bytes];
}

function u32(value: number): number[] {
  const out: number[] = [];
  let current = value >>> 0;
  do {
    let byte = current & 0x7f;
    current >>>= 7;
    if (current !== 0) byte |= 0x80;
    out.push(byte);
  } while (current !== 0);
  return out;
}

function sleb(value: number): number[] {
  const out: number[] = [];
  let current = value | 0;
  let more = true;
  while (more) {
    let byte = current & 0x7f;
    current >>= 7;
    const signBit = (byte & 0x40) !== 0;
    more = !((current === 0 && !signBit) || (current === -1 && signBit));
    if (more) byte |= 0x80;
    out.push(byte);
  }
  return out;
}

function memarg(align: number, offset: number): number[] {
  return [...u32(align), ...u32(offset)];
}

function f64Bytes(value: number): number[] {
  const bytes = new Uint8Array(8);
  new DataView(bytes.buffer).setFloat64(0, value, true);
  return [...bytes];
}

function pagesForData(length: number): number {
  return Math.max(1, Math.ceil((DATA_PTR + length * 8) / PAGE_SIZE));
}

function makeF64Input(length: number): Float64Array {
  const values = new Float64Array(length);
  for (let index = 0; index < length; index += 1) {
    values[index] = (index % 17) + 0.25;
  }
  return values;
}

function writeF64Input(memory: WebAssembly.Memory, values: Float64Array): void {
  new Float64Array(memory.buffer, DATA_PTR, values.length).set(values);
}

function initializeChunkState(memory: WebAssembly.Memory): void {
  const view = new DataView(memory.buffer);
  view.setInt32(STATE_PTR, 0, true);
  view.setFloat64(STATE_PTR + 8, 0, true);
}

function runChunkedF64Loop(instance: WebAssembly.Instance, length: number, fuel: number): { sum: number; chunks: number } {
  const step = instance.exports.sum_f64_chunk as (statePtr: number, dataPtr: number, length: number, fuel: number) => number;
  let done = 0;
  let chunks = 0;
  while (done === 0) {
    done = step(STATE_PTR, DATA_PTR, length, fuel);
    chunks += 1;
  }
  const memory = instance.exports.memory as WebAssembly.Memory;
  return {
    sum: new DataView(memory.buffer).getFloat64(STATE_PTR + 8, true),
    chunks
  };
}

function jsNumberSum(length: number): number {
  let sum = 0;
  for (let index = 0; index < length; index += 1) sum += index;
  return sum;
}

function jsBigIntSum(length: number): bigint {
  let sum = 0n;
  for (let index = 0n; index < BigInt(length); index += 1n) sum += index;
  return sum;
}

function jsF64ArraySum(values: Float64Array): number {
  let sum = 0;
  for (let index = 0; index < values.length; index += 1) sum += values[index] ?? 0;
  return sum;
}

function sameF64(left: number, right: number): boolean {
  return Object.is(left, right) || Math.abs(left - right) <= 1e-9 * Math.max(1, Math.abs(left), Math.abs(right));
}

function wasmBufferSource(bytes: Uint8Array): BufferSource {
  return bytes as unknown as BufferSource;
}

function timed<T>(phases: PhaseAccumulator, name: string, collect: boolean, fn: () => T): T {
  const start = nowMs();
  try {
    return fn();
  } finally {
    if (collect) phases.add(name, nowMs() - start);
  }
}

function nowMs(): number {
  return typeof performance !== "undefined" && typeof performance.now === "function"
    ? performance.now()
    : Date.now();
}

function formatMs(value: number): string {
  if (!Number.isFinite(value)) return "NaN";
  return value.toFixed(value >= 10 ? 2 : 3);
}

function positiveInteger(value: number | undefined, fallback: number): number {
  const number = Math.trunc(Number(value));
  return Number.isFinite(number) && number > 0 ? number : fallback;
}

function nonNegativeInteger(value: number | undefined, fallback: number): number {
  const number = Math.trunc(Number(value));
  return Number.isFinite(number) && number >= 0 ? number : fallback;
}

const enum Section {
  type = 1,
  func = 3,
  memory = 5,
  export = 7,
  code = 10
}

const enum ExportKind {
  func = 0,
  memory = 2
}

const enum ValType {
  i32 = 0x7f,
  i64 = 0x7e,
  f64 = 0x7c
}

const enum BlockType {
  empty = 0x40
}

const enum Op {
  block = 0x02,
  loop = 0x03,
  br = 0x0c,
  brIf = 0x0d,
  end = 0x0b,
  localGet = 0x20,
  localSet = 0x21,
  i32Load = 0x28,
  f64Load = 0x2b,
  i32Store = 0x36,
  f64Store = 0x39,
  i32Const = 0x41,
  i64Const = 0x42,
  f64Const = 0x44,
  i32Eqz = 0x45,
  i32GeU = 0x4f,
  i64GeU = 0x5a,
  i32Add = 0x6a,
  i32Sub = 0x6b,
  i32Shl = 0x74,
  i64Add = 0x7c,
  f64Add = 0xa0
}
