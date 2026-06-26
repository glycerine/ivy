import { describe, expect, test } from "./testHarness.js";
import {
  buildPackages,
  type BuildArtifactStore,
  checkedInWasmStencil,
  createChunkedMixedWasmArtifactFixture,
  createMixedWasmArtifactFixture,
  discoverClangForWasm,
  GOJR_STAGE1_BACKEND,
  parseGoJuniorPackageArchive
} from "../src/index.js";

class MemoryArtifactStore implements BuildArtifactStore {
  public readonly writes = new Map<string, string>();

  public read(path: string): string | undefined {
    return this.writes.get(path);
  }

  public writeAtomic(path: string, source: string): void {
    this.writes.set(path, source);
  }
}

interface MixedArtifactModule {
  instantiateGoJrPackage(
    runtime?: Record<string, unknown>,
    options?: Record<string, unknown>
  ): Promise<{
    diagnostics: unknown[];
    output: string[];
    package: {
      SumF64(ptr: number, n: number): number;
    };
    memory: WebAssembly.Memory;
    wasm: WebAssembly.Instance;
  }>;
}

interface Stage1ArtifactModule {
  instantiateGoJrPackage(
    runtime?: Record<string, unknown>,
    options?: Record<string, unknown>
  ): Promise<{
    diagnostics: unknown[];
    output: string[];
    package: Record<string, unknown>;
    wasm: WebAssembly.Exports;
  }>;
}

interface ChunkedMixedArtifactModule {
  instantiateGoJrPackage(
    runtime?: Record<string, unknown>,
    options?: Record<string, unknown>
  ): Promise<{
    diagnostics: unknown[];
    output: string[];
    package: {
      SumF64Chunked(ptr: number, n: number, fuel: number): Promise<number>;
    };
    memory: WebAssembly.Memory;
  }>;
}

async function importArtifactJavaScript(source: string): Promise<MixedArtifactModule> {
  const url = `data:text/javascript;base64,${Buffer.from(source, "utf8").toString("base64")}`;
  return await import(url) as MixedArtifactModule;
}

function wasmBufferSource(bytes: Uint8Array): ArrayBuffer {
  return new Uint8Array(bytes).buffer;
}

describe("GoJr copy-and-patch Wasm emitter infrastructure", () => {
  test("extracts checked-in Clang scalar Wasm stencils", () => {
    const stencil = checkedInWasmStencil("i64.add.kernel");

    expect(stencil.extracted.signature).toEqual({
      params: ["i64", "i64"],
      results: ["i64"]
    });
    expect(stencil.extracted.module.imports).toEqual([]);
    expect(stencil.extracted.bodyBytes.length).toBeGreaterThan(0);

    const module = new WebAssembly.Module(wasmBufferSource(stencil.wasmBytes));
    const instance = new WebAssembly.Instance(module);
    const add = instance.exports.add_i64 as (a: bigint, b: bigint) => bigint;
    expect(add(40n, 2n)).toBe(42n);
  });

  test("extracts checked-in Clang imported-memory Wasm stencils", () => {
    const stencil = checkedInWasmStencil("f64.slice.sum.kernel");

    expect(stencil.extracted.signature).toEqual({
      params: ["i32", "i32"],
      results: ["f64"]
    });
    expect(stencil.extracted.importedMemory?.module).toBe("env");
    expect(stencil.extracted.importedMemory?.name).toBe("memory");
    expect(stencil.extracted.bodyBytes.length).toBeGreaterThan(0);

    const memory = new WebAssembly.Memory({ initial: 2 });
    const values = new Float64Array(memory.buffer, 64, 4);
    values.set([1.25, 2.5, 3.75, 4.5]);
    const module = new WebAssembly.Module(wasmBufferSource(stencil.wasmBytes));
    const instance = new WebAssembly.Instance(module, { env: { memory } });
    const sum = instance.exports.sum_f64 as (ptr: number, n: number) => number;
    expect(sum(64, 4)).toBe(12);
  });

  test("runs checked-in fuel-chunked Wasm loop stencils with JS-owned state", () => {
    const stencil = checkedInWasmStencil("f64.slice.sum.chunk.kernel");

    expect(stencil.extracted.signature).toEqual({
      params: ["i32", "i32", "i32", "i32"],
      results: ["i32"]
    });
    expect(stencil.extracted.importedMemory?.module).toBe("env");
    expect(stencil.extracted.importedMemory?.name).toBe("memory");

    const memory = new WebAssembly.Memory({ initial: 2 });
    const view = new DataView(memory.buffer);
    const statePtr = 0;
    const dataPtr = 64;
    view.setInt32(statePtr, 0, true);
    view.setFloat64(statePtr + 8, 0, true);
    new Float64Array(memory.buffer, dataPtr, 5).set([1, 2, 3, 4, 5]);

    const module = new WebAssembly.Module(wasmBufferSource(stencil.wasmBytes));
    const instance = new WebAssembly.Instance(module, { env: { memory } });
    const chunk = instance.exports.sum_f64_chunk as (statePtr: number, dataPtr: number, n: number, fuel: number) => number;

    let done = 0;
    let chunks = 0;
    while (done === 0) {
      done = chunk(statePtr, dataPtr, 5, 2);
      chunks += 1;
      if (chunks > 10) throw new Error("chunked stencil did not finish");
    }

    expect(chunks).toBe(3);
    expect(view.getInt32(statePtr, true)).toBe(5);
    expect(view.getFloat64(statePtr + 8, true)).toBe(15);
  });

  test("discovers Clang through configurable candidates and probe", () => {
    const result = discoverClangForWasm({
      env: { GOJR_CLANG: "custom-clang" },
      candidates: ["fallback-clang"],
      probe(path) {
        return path === "custom-clang"
          ? { ok: true }
          : { ok: false, message: "nope" };
      }
    });

    expect(result.ok).toBe(true);
    expect(result.clangPath).toBe("custom-clang");
    expect(result.attempted).toEqual([{ path: "custom-clang", ok: true }]);
  });

  test("builds a mixed JS and Wasm package artifact fixture that executes the Wasm kernel", async () => {
    const fixture = createMixedWasmArtifactFixture();
    const archive = parseGoJuniorPackageArchive(fixture.artifactSource);

    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js", "_gojr.wasm"]);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");
    expect(archive?.wasmBase64).toBe(fixture.wasmBase64);

    const pkgdefMember = archive?.members.find((member) => member.name === "__.PKGDEF")?.data ?? "";
    const wasmMember = archive?.members.find((member) => member.name === "_gojr.wasm")?.data ?? "";
    expect(pkgdefMember.length < 4096).toBe(true);
    expect(wasmMember.length > 0).toBe(true);
    expect(fixture.artifactSource.length < 12000).toBe(true);

    const module = await importArtifactJavaScript(archive?.javascript ?? "");
    const result = await module.instantiateGoJrPackage({}, { wasmBase64: archive?.wasmBase64 });
    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual([]);
    const values = new Float64Array(result.memory.buffer, 64, 5);
    values.set([1, 2, 3, 4, 5]);
    expect(result.package.SumF64(64, 5)).toBe(15);
  });

  test("builds a chunked mixed Wasm fixture that drives fuel from JavaScript", async () => {
    const fixture = createChunkedMixedWasmArtifactFixture();
    const archive = parseGoJuniorPackageArchive(fixture.artifactSource);

    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js", "_gojr.wasm"]);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as ChunkedMixedArtifactModule;
    const result = await module.instantiateGoJrPackage({}, { wasmBase64: archive?.wasmBase64 });
    new Float64Array(result.memory.buffer, 64, 5).set([1, 2, 3, 4, 5]);

    expect(await result.package.SumF64Chunked(64, 5, 2)).toBe(15);
  });

  test("builds an opt-in Stage 1 package artifact with JS host code and a Wasm kernel", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage1",
      artifactRoot: "/tmp/gojr-stage1",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage1.go",
        source: `package stage1

const Greeting = "hi"
var Count int64
var Names = []string{"ada", "grace"}
var Data = []byte{1, 2, 3}
var Files = map[string][]byte{"a": []byte{4, 5}}

func Add(a, b int64) int64 { return a + b }
func Answer() int64 { return 42 }
func Noop() {}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage1/example.com/stage1.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js", "_gojr.wasm"]);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");
    expect(archive?.wasmBase64).toBeDefined();

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage({}, { wasmBase64: archive?.wasmBase64 });
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(pkg.Greeting).toBe("hi");
    expect(pkg.Count).toBe(0n);
    expect(pkg.Names).toEqual(["ada", "grace"]);
    expect(Array.from(pkg.Data as Uint8Array)).toEqual([1, 2, 3]);
    expect(Array.from((pkg.Files as Map<string, Uint8Array>).get("a") ?? [])).toEqual([4, 5]);
    expect(await (pkg.Add as (a: bigint, b: bigint) => Promise<bigint>)(40n, 2n)).toBe(42n);
    expect(await (pkg.Answer as () => Promise<bigint>)()).toBe(42n);
    expect(await (pkg.Noop as () => Promise<null>)()).toBeNull();
  });

  test("reports unsupported Stage 1 lowering as GOJR_EMIT001", () => {
    const result = buildPackages({
      importPath: "example.com/stage1bad",
      artifactRoot: "/tmp/gojr-stage1bad",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "bad.go",
        source: "package stage1bad\n\nfunc Mul(a, b int64) int64 { return a * b }\n"
      }]
    }, new MemoryArtifactStore());

    expect(result.ok).toBe(false);
    expect(result.diagnostics.map((diagnostic) => diagnostic.code)).toContain("GOJR_EMIT001");
  });
});
