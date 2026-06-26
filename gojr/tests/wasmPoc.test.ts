import { describe, expect, test } from "./testHarness.js";
import {
  formatGoJuniorWasmPocReport,
  runGoJuniorWasmPocBenchmark,
  runGoJuniorBenchmark
} from "../src/index.js";

describe("Go-junior Wasm copy-and-patch POC", () => {
  test("runs JS-hosted Wasm numeric loop benchmarks with matching results", async () => {
    const report = await runGoJuniorWasmPocBenchmark({
      iterations: 1,
      workItems: 1000,
      fuel: 128
    });

    expect(report.ok).toBe(true);
    expect(report.diagnostics).toEqual([]);
    expect(report.metrics.i64ModuleBytes).toBeGreaterThan(0);
    expect(report.metrics.f64ModuleBytes).toBeGreaterThan(0);
    expect(report.metrics.f64ChunkModuleBytes).toBeGreaterThan(0);
    expect(report.metrics.jsBigIntResult).toBe("499500");
    expect(report.metrics.wasmI64Result).toBe("499500");
    expect(report.metrics.jsF64Result).toBe(report.metrics.wasmF64Result);
    expect(report.metrics.jsF64Result).toBe(report.metrics.wasmChunkedF64Result);
    expect(report.metrics.chunkCountMax).toBe(8);
    expect(report.phases.map((phase) => phase.name)).toContain("wasm-f64array-fuel-loop");
    expect(formatGoJuniorWasmPocReport(report)).toContain("gojr wasm-poc");
  });

  test("is available through the shared benchmark report", async () => {
    const report = await runGoJuniorBenchmark(undefined, {
      wasmPoc: true,
      iterations: 1,
      wasmWorkItems: 256,
      wasmFuel: 64
    });

    expect(report.ok).toBe(true);
    expect(report.cases).toEqual([]);
    expect(report.wasmPoc?.ok).toBe(true);
    expect(report.wasmPoc?.metrics.wasmI64Result).toBe("32640");
  });
});
