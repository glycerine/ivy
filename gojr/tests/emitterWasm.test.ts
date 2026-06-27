import { describe, expect, test } from "./testHarness.js";
import {
  buildPackages,
  type BuildArtifactStore,
  checkedInWasmStencil,
  createStage1RuntimeCore,
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
  const module = await import(url) as MixedArtifactModule;
  return {
    ...module,
    instantiateGoJrPackage(runtime: Record<string, unknown> = {}, options: Record<string, unknown> = {}) {
      return module.instantiateGoJrPackage(createStage1RuntimeCore(runtime), options);
    }
  };
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

    const scalar = checkedInWasmStencil("i64.scalar.add");
    const scalarModule = new WebAssembly.Module(wasmBufferSource(scalar.wasmBytes));
    const scalarInstance = new WebAssembly.Instance(scalarModule);
    const scalarExports = scalarInstance.exports as Record<string, (...args: bigint[]) => bigint | number>;
    expect(Object.keys(scalarExports).sort()).toEqual([
      "gojr_i64_add",
      "gojr_i64_and",
      "gojr_i64_bitclear",
      "gojr_i64_div",
      "gojr_i64_eq",
      "gojr_i64_ge",
      "gojr_i64_gt",
      "gojr_i64_le",
      "gojr_i64_lt",
      "gojr_i64_mul",
      "gojr_i64_ne",
      "gojr_i64_or",
      "gojr_i64_rem",
      "gojr_i64_shl",
      "gojr_i64_shr",
      "gojr_i64_sub",
      "gojr_i64_xor"
    ]);
    expect(scalarExports.gojr_i64_add!(40n, 2n)).toBe(42n);
    expect(scalarExports.gojr_i64_mul!(6n, 7n)).toBe(42n);
    expect(scalarExports.gojr_i64_lt!(2n, 9n)).toBe(1);
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
func Sub(a, b int64) int64 { return a - b }
func Mul(a, b int64) int64 { return a * b }
func Div(a, b int64) int64 { return a / b }
func Rem(a, b int64) int64 { return a % b }
func And(a, b int64) int64 { return a & b }
func Or(a, b int64) int64 { return a | b }
func Xor(a, b int64) int64 { return a ^ b }
func Bitclear(a, b int64) int64 { return a &^ b }
func Shl(a, b int64) int64 { return a << b }
func Shr(a, b int64) int64 { return a >> b }
func Eq(a, b int64) bool { return a == b }
func Ne(a, b int64) bool { return a != b }
func Lt(a, b int64) bool { return a < b }
func Le(a, b int64) bool { return a <= b }
func Gt(a, b int64) bool { return a > b }
func Ge(a, b int64) bool { return a >= b }
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
    expect(await (pkg.Sub as (a: bigint, b: bigint) => Promise<bigint>)(40n, 2n)).toBe(38n);
    expect(await (pkg.Mul as (a: bigint, b: bigint) => Promise<bigint>)(6n, 7n)).toBe(42n);
    expect(await (pkg.Div as (a: bigint, b: bigint) => Promise<bigint>)(40n, 5n)).toBe(8n);
    expect(await (pkg.Rem as (a: bigint, b: bigint) => Promise<bigint>)(40n, 6n)).toBe(4n);
    expect(await (pkg.And as (a: bigint, b: bigint) => Promise<bigint>)(6n, 3n)).toBe(2n);
    expect(await (pkg.Or as (a: bigint, b: bigint) => Promise<bigint>)(4n, 1n)).toBe(5n);
    expect(await (pkg.Xor as (a: bigint, b: bigint) => Promise<bigint>)(6n, 3n)).toBe(5n);
    expect(await (pkg.Bitclear as (a: bigint, b: bigint) => Promise<bigint>)(7n, 3n)).toBe(4n);
    expect(await (pkg.Shl as (a: bigint, b: bigint) => Promise<bigint>)(3n, 2n)).toBe(12n);
    expect(await (pkg.Shr as (a: bigint, b: bigint) => Promise<bigint>)(8n, 1n)).toBe(4n);
    expect(await (pkg.Eq as (a: bigint, b: bigint) => Promise<boolean>)(3n, 3n)).toBe(true);
    expect(await (pkg.Ne as (a: bigint, b: bigint) => Promise<boolean>)(3n, 4n)).toBe(true);
    expect(await (pkg.Lt as (a: bigint, b: bigint) => Promise<boolean>)(3n, 4n)).toBe(true);
    expect(await (pkg.Le as (a: bigint, b: bigint) => Promise<boolean>)(4n, 4n)).toBe(true);
    expect(await (pkg.Gt as (a: bigint, b: bigint) => Promise<boolean>)(5n, 4n)).toBe(true);
    expect(await (pkg.Ge as (a: bigint, b: bigint) => Promise<boolean>)(5n, 5n)).toBe(true);
    expect(await (pkg.Answer as () => Promise<bigint>)()).toBe(42n);
    expect(await (pkg.Noop as () => Promise<null>)()).toBeNull();
  });

  test("emits byte literal supernodes instead of giant element AST-shaped payloads", async () => {
    const dataA = Array.from({ length: 256 }, (_, index) => index % 256);
    const dataB = Array.from({ length: 192 }, (_, index) => (255 - index) & 255);
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage1data",
      artifactRoot: "/tmp/gojr-stage1data",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage1data.go",
        source: `package stage1data

var Blob = []byte{${dataA.join(", ")}}
var Files = map[string][]byte{
	"a": []byte{${dataA.join(", ")}},
	"b": []byte{${dataB.join(", ")}},
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage1data/example.com/stage1data.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).toContain("__gojrBytesBase64(");
    expect(archive?.javascript).toContain("__gojrMapStringBytesBase64(");
    expect(archive?.javascript).not.toContain("runtime.ast");
    expect(archive?.javascript).not.toContain(dataA.slice(0, 32).join(", "));
    expect((archive?.members.find((member) => member.name === "__.PKGDEF")?.data.length ?? 0) < 8192).toBe(true);

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(Array.from(pkg.Blob as Uint8Array)).toEqual(dataA);
    expect(Array.from((pkg.Files as Map<string, Uint8Array>).get("a") ?? [])).toEqual(dataA);
    expect(Array.from((pkg.Files as Map<string, Uint8Array>).get("b") ?? [])).toEqual(dataB);
  });

  test("keeps shared Stage 1 helpers out of timezone-style package artifacts", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "4d63.com/tz",
      artifactRoot: "/tmp/gojr-stage1tz",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        time: [{
          filename: "time.go",
          source: `package time

type Location struct { name string }
var UTC = &Location{name: "UTC"}
func (l *Location) String() string { return l.name }
`
        }]
      },
      files: [{
        filename: "tz.go",
        source: `package tz

import "time"

var fileNames = []string{"UTC"}
var files = map[string][]byte{"UTC": []byte{85, 84, 67}}

func tzData(name string) ([]byte, bool) {
	data, ok := files[name]
	return data, ok
}

func LoadLocation(name string) (*time.Location, error) {
	_, _ = tzData(name)
	return time.UTC, nil
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage1tz/4d63.com/tz.a") ?? "");
    const timeArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage1tz/time.a") ?? "");
    if (!archive || !timeArchive) throw new Error("missing timezone-style archive");
    expect(archive.javascript.length < 100_000).toBe(true);
    expect(archive.javascript).toContain("pkg[\"fileNames\"]");
    expect(archive.javascript).toContain("pkg[\"files\"]");
    expect(archive.javascript).toContain("pkg[\"tzData\"]");
    expect(archive.javascript).toContain("pkg[\"LoadLocation\"]");
    expect(archive.javascript).not.toContain("function __gojrReflectDescriptorString");
    expect(archive.javascript).not.toContain("const __gojrReflectTypeMethods");
    expect(archive.javascript).not.toContain("function __gojrDecodeBase64");
    expect(archive.javascript).not.toContain("function __gojrBuiltinImport(path, importsByPath) {");

    const timeModule = await importArtifactJavaScript(timeArchive.javascript ?? "") as unknown as Stage1ArtifactModule;
    const timePackage = await timeModule.instantiateGoJrPackage();
    const module = await importArtifactJavaScript(archive.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage({}, { importsByPath: { time: timePackage } });
    expect(instantiated.diagnostics).toEqual([]);
    expect(Array.from(instantiated.package.fileNames as string[])).toEqual(["UTC"]);
    expect(Array.from((instantiated.package.files as Map<string, Uint8Array>).get("UTC") ?? [])).toEqual([85, 84, 67]);
    const [data, ok] = await (instantiated.package.tzData as (name: string) => Promise<[Uint8Array, boolean]>)("UTC");
    expect(Array.from(data)).toEqual([85, 84, 67]);
    expect(ok).toBe(true);
    const [location, err] = await (instantiated.package.LoadLocation as (name: string) => Promise<[unknown, unknown]>)("UTC");
    expect(Boolean(err && typeof err === "object" && (err as { __gojrInterface?: boolean; value?: unknown }).__gojrInterface === true &&
      (err as { value?: unknown }).value === null)).toBe(true);
    expect(location).toBe(timePackage.package.UTC);
  });

  test("forwards caller constraint element type into local generic calls", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/genericforward",
      artifactRoot: "/tmp/gojr-genericforward",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "genericforward.go",
        source: `package genericforward

func SortFunc[S ~[]E, E any](x S, cmp func(a, b E) int) {
	n := len(x)
	pdqsortCmpFunc(x, 0, n, cmp)
}

func pdqsortCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) {
	if a < b {
		_ = data[a]
	}
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-genericforward/example.com/genericforward.a") ?? "");
    if (!archive) throw new Error("missing genericforward archive");
    expect(archive.javascript).toContain(
      'pkg["pdqsortCmpFunc"](({ "E": __gojrSubstituteType("E", __gojrTypeArgs) })'
    );
    expect(archive.javascript).not.toContain('pkg["pdqsortCmpFunc"](({ "E": "any" })');
  });

  test("lowers Stage 3 concrete expressions without interpreter expression calls", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage3expr",
      artifactRoot: "/tmp/gojr-stage3expr",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage3expr.go",
        source: `package stage3expr

type Point struct {
	X int64
	Name string
}
type errString struct {
	s string
}
type MyInts []int64

var Numbers = []int64{4, 5, 6}
var Labels = map[string]int64{"a": 11}
var P = Point{X: 7, Name: "ada"}

func Mul(a, b int64) int64 { return a * b }
func Mix(a, b int64) int64 { return (a + b) * 2 - 3 }
func Cat(a, b string) string { return a + ":" + b }
func Pick(i int) int64 { return Numbers[i] }
func Tail() []int64 { return Numbers[1:] }
func ByteAt(s string, i int) uint8 { return s[i] }
func Lookup() int64 { return Labels["a"] }
func Missing() int64 { return Labels["missing"] }
func Field() string { return P.Name }
func Convert(a int64) float64 { return float64(a) + 0.5 }
func Pair(a, b int64) (int64, int64) { return a, b }
func ComplexOps() complex128 { return complex(1, 2) * complex(3, 4) + 5i }
func RealPart() float64 { return real(ComplexOps()) }
func ImagPart() float64 { return imag(ComplexOps()) }
func Bytes(s string) []byte { return []byte(s) }
func StringFromBytes() string { return string([]byte{104, 105}) }
func StringFromDynamicByte(b byte) string { return string([]byte{b}) }
func ErrText(s string) string { return (&errString{s}).s }
func AssertInt(x any) int64 { return x.(int64) }
func AssertIntOk(x any) (int64, bool) { v, ok := x.(int64); return v, ok }
func InterfaceConvert() string {
	x := interface{}(P)
	return x.(Point).Name
}
func BuiltinSliceOps() (int, int, int64) {
	xs := []int64{1, 2}
	xs = append(xs, 3, 4)
	return len(xs), cap(xs), xs[2]
}
func BuiltinCopyDelete() (int, bool, uint8) {
	dst := []byte{0, 0, 0}
	n := copy(dst, []byte{7, 8})
	m := map[string]int64{"a": 1}
	delete(m, "a")
	_, ok := m["a"]
	return n, ok, dst[1]
}
func KeyedSlice() (int, string, string) {
	xs := []string{2: "two", 4: "four"}
	return len(xs), xs[0], xs[4]
}
func ConvertNamedSlice() []int64 {
	xs := MyInts{8, 9}
	return []int64(xs)
}
func ThreeIndexCap() (int, int) {
	xs := []int64{1, 2, 3, 4, 5}
	ys := xs[1:3:4]
	return len(ys), cap(ys)
}
func BuiltinPanic() { panic("boom") }
func (p Point) Sum(delta int64) int64 { return p.X + delta }
func MethodCall(delta int64) int64 { return P.Sum(delta) }
func MethodValue() func(int64) int64 { return P.Sum }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage3expr/example.com/stage3expr.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js", "_gojr.wasm"]);
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");
    expect(archive?.javascript).not.toContain("runtime.binary");
    expect(archive?.javascript).toContain("pkg[\"Mul\"] = async");
    expect(archive?.javascript).toContain("gojr_i64_mul");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage({}, { wasmBase64: archive?.wasmBase64 });
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.Mul as (a: bigint, b: bigint) => Promise<bigint>)(6n, 7n)).toBe(42n);
    expect(await (pkg.Mix as (a: bigint, b: bigint) => Promise<bigint>)(3n, 4n)).toBe(11n);
    expect(await (pkg.Cat as (a: string, b: string) => Promise<string>)("go", "jr")).toBe("go:jr");
    expect(await (pkg.Pick as (i: bigint) => Promise<bigint>)(2n)).toBe(6n);
    expect(await (pkg.Tail as () => Promise<bigint[]>)()).toEqual([5n, 6n]);
    expect(await (pkg.ByteAt as (s: string, i: bigint) => Promise<bigint>)("Aπ", 1n)).toBe(207n);
    expect(await (pkg.Lookup as () => Promise<bigint>)()).toBe(11n);
    expect(await (pkg.Missing as () => Promise<bigint>)()).toBe(0n);
    expect(await (pkg.Field as () => Promise<string>)()).toBe("ada");
    expect(await (pkg.Convert as (a: bigint) => Promise<number>)(4n)).toBe(4.5);
    expect(await (pkg.Pair as (a: bigint, b: bigint) => Promise<[bigint, bigint]>)(7n, 9n)).toEqual([7n, 9n]);
    expect(await (pkg.ComplexOps as () => Promise<{ real: number; imag: number }>)()).toEqual({ real: -5, imag: 15 });
    expect(await (pkg.RealPart as () => Promise<number>)()).toBe(-5);
    expect(await (pkg.ImagPart as () => Promise<number>)()).toBe(15);
    expect(Array.from(await (pkg.Bytes as (s: string) => Promise<Uint8Array>)("Aπ"))).toEqual([65, 207, 128]);
    expect(await (pkg.StringFromBytes as () => Promise<string>)()).toBe("hi");
    expect(await (pkg.StringFromDynamicByte as (b: bigint) => Promise<string>)(33n)).toBe("!");
    expect(await (pkg.ErrText as (s: string) => Promise<string>)("boom")).toBe("boom");
    expect(await (pkg.AssertInt as (x: unknown) => Promise<bigint>)(42n)).toBe(42n);
    expect(await (pkg.AssertIntOk as (x: unknown) => Promise<[bigint, boolean]>)(42n)).toEqual([42n, true]);
    expect(await (pkg.AssertIntOk as (x: unknown) => Promise<[bigint, boolean]>)("nope")).toEqual([0n, false]);
    expect(await (pkg.InterfaceConvert as () => Promise<string>)()).toBe("ada");
    expect(await (pkg.BuiltinSliceOps as () => Promise<[bigint, bigint, bigint]>)()).toEqual([4n, 4n, 3n]);
    expect(await (pkg.BuiltinCopyDelete as () => Promise<[bigint, boolean, bigint]>)()).toEqual([2n, false, 8n]);
    expect(await (pkg.KeyedSlice as () => Promise<[bigint, string, string]>)()).toEqual([5n, "", "four"]);
    expect(await (pkg.ConvertNamedSlice as () => Promise<bigint[]>)()).toEqual([8n, 9n]);
    expect(await (pkg.ThreeIndexCap as () => Promise<[bigint, bigint]>)()).toEqual([2n, 3n]);
    let panicMessage = "";
    try {
      await (pkg.BuiltinPanic as () => Promise<null>)();
    } catch (error) {
      panicMessage = error instanceof Error ? error.message : String(error);
    }
    expect(panicMessage).toContain("panic: boom");
    expect(await (pkg["Point.Sum"] as (p: { X: bigint; Name: string }, delta: bigint) => Promise<bigint>)(pkg.P as { X: bigint; Name: string }, 4n)).toBe(11n);
    expect(await (pkg.MethodCall as (delta: bigint) => Promise<bigint>)(5n)).toBe(12n);
    const methodValue = await (pkg.MethodValue as () => Promise<(delta: bigint) => Promise<bigint>>)();
    expect(await methodValue(8n)).toBe(15n);
  });

  test("lowers Stage 4 structured statements without interpreter statement calls", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage4stmt",
      artifactRoot: "/tmp/gojr-stage4stmt",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage4stmt.go",
        source: `package stage4stmt

var Labels = map[string]int64{"a": 11, "b": 31}
var Log string

func Pair(a, b int64) (int64, int64) { return a, b }
func note(s string) { Log += s }
func sendValue(c chan int64, v int64) { c <- v }

func Control(n int64) int64 {
	sum := int64(0)
	for i := int64(0); i < n; i++ {
		if i == 2 {
			continue
		}
		sum += i
	}
	return sum
}

func RangeSlice() int64 {
	sum := int64(0)
	for _, v := range []int64{1, 2, 3} {
		sum += v
	}
	return sum
}

func RangeByteModulo() byte {
	var out byte
	for _, b := range []byte{65} {
		out = b % 16
	}
	return out
}

func RangeMap() int64 {
	sum := int64(0)
	for _, v := range Labels {
		sum += v
	}
	return sum
}

func SwitchIt(x int64) string {
	switch x {
	case 1:
		return "one"
	case 2:
		fallthrough
	case 3:
		return "few"
	default:
		return "many"
	}
}

func MapOk(key string) (int64, bool) {
	v, ok := Labels[key]
	return v, ok
}

func TupleUse() int64 {
	a, b := Pair(2, 3)
	return a + b
}

func AssignSwap() int64 {
	a, b := int64(1), int64(2)
	a, b = b, a
	return a*10 + b
}

type Flag bool

func ForInitMultiAndBool() int64 {
	sum := int64(0)
	for i, j := int64(0), int64(2); i < 3; i++ {
		if bool(Flag(i < j)) {
			sum += j
		}
	}
	return sum
}

func IfInit(x int64) string {
	if y := x + 1; y > 4 {
		return "big"
	} else if y == 4 {
		return "four"
	}
	return "small"
}

func Named() (a, b int64) {
	a = 5
	b = 7
	return
}

type Counter int64

func NamedIntegerForPost() Counter {
	var i Counter
	for ; i < 3; i++ {
	}
	return i
}

func Labeled() int64 {
	sum := int64(0)
outer:
	for i := int64(0); i < 3; i++ {
		for j := int64(0); j < 3; j++ {
			if i == 1 && j == 1 {
				break outer
			}
			if j == 0 {
				continue
			}
			sum += i + j
		}
	}
	return sum
}

func DeferOrder() string {
	Log = ""
	defer note("a")
	defer note("b")
	return "body"
}

func ChannelBuffered() int64 {
	c := make(chan int64, 1)
	c <- 42
	return <-c
}

func ChannelClose() bool {
	c := make(chan int64)
	close(c)
	_, ok := <-c
	return ok
}

func ChannelGo() int64 {
	c := make(chan int64)
	go sendValue(c, 9)
	v, ok := <-c
	if ok {
		return v
	}
	return 0
}

func SelectDefault() int64 {
	c := make(chan int64)
	out := int64(0)
	select {
	case out = <-c:
	default:
		out = 7
	}
	return out
}

func SelectRecv() int64 {
	c := make(chan int64, 1)
	c <- 5
	out := int64(0)
	select {
	case out = <-c:
	default:
		out = 7
	}
	return out
}

func SelectSend() int64 {
	c := make(chan int64, 1)
	select {
	case c <- 6:
	default:
	}
	return <-c
}

func SelectBreakInLoop() int64 {
	c := make(chan int64, 1)
	c <- 1
	out := int64(0)
	for i := int64(0); i < 3; i++ {
		select {
		case <-c:
			out += 10
			break
			out += 100
		default:
			out += 1
		}
	}
	return out
}

func TypeSwitch(x any) string {
	switch v := x.(type) {
	case int64:
		return "int"
	case string:
		return v
	default:
		return "other"
	}
}

func GotoSum() int64 {
	i := int64(0)
	sum := int64(0)
loop:
	if i >= 4 {
		goto done
	}
	sum += i
	i++
	goto loop
done:
	return sum
}

func LocalForwardGoto(n int64) int64 {
	sum := int64(0)
	for i := int64(0); i < n; i++ {
		for j := int64(0); j < 3; j++ {
			if j == 1 {
				goto next
			}
			sum += 10
		}
		sum += 100
	next:
		sum += i
	}
	return sum
}

func NestedBackwardGoto(n int64) int64 {
	sum := int64(0)
	for i := int64(0); i < n; i++ {
		j := int64(0)
		goto Skip
	CheckAndLoop:
		if j >= 3 {
			continue
		}
		sum += j
		j++
		goto CheckAndLoop
	Skip:
		j++
		goto CheckAndLoop
	}
	return sum
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage4stmt/example.com/stage4stmt.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js"]);
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");
    expect(archive?.javascript).not.toContain("executeStatement");
    expect(archive?.javascript).not.toContain("evaluateExpression");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.Control as (n: bigint) => Promise<bigint>)(5n)).toBe(8n);
    expect(await (pkg.RangeSlice as () => Promise<bigint>)()).toBe(6n);
    expect(await (pkg.RangeByteModulo as () => Promise<bigint>)()).toBe(1n);
    expect(await (pkg.RangeMap as () => Promise<bigint>)()).toBe(42n);
    expect(await (pkg.SwitchIt as (x: bigint) => Promise<string>)(1n)).toBe("one");
    expect(await (pkg.SwitchIt as (x: bigint) => Promise<string>)(2n)).toBe("few");
    expect(await (pkg.SwitchIt as (x: bigint) => Promise<string>)(9n)).toBe("many");
    expect(await (pkg.MapOk as (key: string) => Promise<[bigint, boolean]>)("a")).toEqual([11n, true]);
    expect(await (pkg.MapOk as (key: string) => Promise<[bigint, boolean]>)("missing")).toEqual([0n, false]);
    expect(await (pkg.TupleUse as () => Promise<bigint>)()).toBe(5n);
    expect(await (pkg.AssignSwap as () => Promise<bigint>)()).toBe(21n);
    expect(await (pkg.ForInitMultiAndBool as () => Promise<bigint>)()).toBe(4n);
    expect(await (pkg.IfInit as (x: bigint) => Promise<string>)(3n)).toBe("four");
    expect(await (pkg.IfInit as (x: bigint) => Promise<string>)(4n)).toBe("big");
    expect(await (pkg.IfInit as (x: bigint) => Promise<string>)(1n)).toBe("small");
    expect(await (pkg.Named as () => Promise<[bigint, bigint]>)()).toEqual([5n, 7n]);
    expect(await (pkg.NamedIntegerForPost as () => Promise<bigint>)()).toBe(3n);
    expect(await (pkg.Labeled as () => Promise<bigint>)()).toBe(3n);
    expect(await (pkg.DeferOrder as () => Promise<string>)()).toBe("body");
    expect(pkg.Log).toBe("ba");
    expect(await (pkg.ChannelBuffered as () => Promise<bigint>)()).toBe(42n);
    expect(await (pkg.ChannelClose as () => Promise<boolean>)()).toBe(false);
    expect(await (pkg.ChannelGo as () => Promise<bigint>)()).toBe(9n);
    expect(await (pkg.SelectDefault as () => Promise<bigint>)()).toBe(7n);
    expect(await (pkg.SelectRecv as () => Promise<bigint>)()).toBe(5n);
    expect(await (pkg.SelectSend as () => Promise<bigint>)()).toBe(6n);
    expect(await (pkg.SelectBreakInLoop as () => Promise<bigint>)()).toBe(12n);
    expect(await (pkg.TypeSwitch as (x: unknown) => Promise<string>)(3n)).toBe("int");
    expect(await (pkg.TypeSwitch as (x: unknown) => Promise<string>)("hi")).toBe("hi");
    expect(await (pkg.TypeSwitch as (x: unknown) => Promise<string>)(true)).toBe("other");
    expect(await (pkg.GotoSum as () => Promise<bigint>)()).toBe(6n);
    expect(await (pkg.LocalForwardGoto as (n: bigint) => Promise<bigint>)(3n)).toBe(33n);
    expect(await (pkg.NestedBackwardGoto as (n: bigint) => Promise<bigint>)(2n)).toBe(6n);
  });

  test("emits labeled switch expressions as valid JavaScript blocks", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage4labeledswitch",
      artifactRoot: "/tmp/gojr-stage4labeledswitch",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage4labeledswitch.go",
        source: `package stage4labeledswitch

func LabeledSwitchWithExpression(x int64) int64 {
Switch:
	switch x {
	case 1:
		for i := int64(0); i < 3; i++ {
			if i == 1 {
				break Switch
			}
		}
		return 99
	default:
		return 7
	}
	return 11
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage4labeledswitch/example.com/stage4labeledswitch.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).toContain("Switch: {");
    expect(archive?.javascript).not.toContain("Switch:\n  const ");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    const pkg = instantiated.package;
    expect(await (pkg.LabeledSwitchWithExpression as (x: bigint) => Promise<bigint>)(1n)).toBe(11n);
    expect(await (pkg.LabeledSwitchWithExpression as (x: bigint) => Promise<bigint>)(2n)).toBe(7n);
  });

  test("runs generated init functions after package variable initialization in source order", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage5init",
      artifactRoot: "/tmp/gojr-stage5init",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage5init.go",
        source: `package stage5init

var Log = seed()

func seed() string { return "seed" }
func init() { Log += ":a" }
func init() { Log += ":b" }
func Value() string { return Log }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage5init/example.com/stage5init.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("pkg[\"init\"]");
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(pkg.Log).toBe("seed:a:b");
    expect(await (pkg.Value as () => Promise<string>)()).toBe("seed:a:b");
  });

  test("lowers generated function literals with lexical captures", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage5closures",
      artifactRoot: "/tmp/gojr-stage5closures",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage5closures.go",
        source: `package stage5closures

var PackageBase int64 = 100

func ClosureSum() int64 {
	base := int64(10)
	add := func(x int64) int64 {
		base += x
		return base
	}
	return add(5) + add(1)
}

func ReturnClosure() func(int64) int64 {
	base := int64(3)
	return func(x int64) int64 {
		return PackageBase + base + x
	}
}

func Fact(n int64) int64 {
	if n <= 1 {
		return 1
	}
	return n * Fact(n-1)
}

func Even(n int64) bool {
	if n == 0 {
		return true
	}
	return Odd(n - 1)
}

func Odd(n int64) bool {
	if n == 0 {
		return false
	}
	return Even(n - 1)
}

func Sum(prefix int64, vals ...int64) int64 {
	sum := prefix
	for _, v := range vals {
		sum += v
	}
	return sum
}

func VariadicCalls() (int64, int64) {
	xs := []int64{2, 3}
	return Sum(1, 2, 3), Sum(1, xs...)
}

func VariadicLiteral() int64 {
	f := func(vals ...int64) int64 {
		sum := int64(0)
		for _, v := range vals {
			sum += v
		}
		return sum
	}
	return f(4, 5)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage5closures/example.com/stage5closures.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ClosureSum as () => Promise<bigint>)()).toBe(31n);
    const closure = await (pkg.ReturnClosure as () => Promise<(x: bigint) => Promise<bigint>>)();
    expect(await closure(4n)).toBe(107n);
    expect(await (pkg.Fact as (n: bigint) => Promise<bigint>)(5n)).toBe(120n);
    expect(await (pkg.Even as (n: bigint) => Promise<boolean>)(8n)).toBe(true);
    expect(await (pkg.Odd as (n: bigint) => Promise<boolean>)(8n)).toBe(false);
    expect(await (pkg.VariadicCalls as () => Promise<[bigint, bigint]>)()).toEqual([6n, 6n]);
    expect(await (pkg.VariadicLiteral as () => Promise<bigint>)()).toBe(9n);
  });

  test("calls generated dependency packages through importsByPath", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5imports",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/dep": [{
          filename: "dep.go",
          source: `package dep

var Count int64

func init() { Count = 40 }
func Value(delta int64) int64 { return Count + delta }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import d "example.com/dep"

func Call() int64 { return d.Value(2) }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5imports/example.com/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5imports/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(rootArchive?.javascript).not.toContain("runtime.ast");

    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/dep": dep.package } });
    expect(dep.package.Count).toBe(40n);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(42n);
  });

  test("preserves imported function-typed parameter signatures", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedfuncparam",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/dep": [{
          filename: "dep.go",
          source: `package dep

type Hash uint

const SHA Hash = 10

var Seen Hash
var Callback func() Hash

func Register(h Hash, f func() Hash) {
	Seen = h
	Callback = f
}

func Sum() Hash { return Seen + Callback() }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import dep "example.com/dep"

func Call() dep.Hash {
	dep.Register(dep.SHA, func() dep.Hash { return dep.Hash(2) })
	return dep.Sum()
}
`
      }]
    }, store);
    const otherResult = buildPackages({
      importPath: "example.com/other",
      artifactRoot: "/tmp/gojr-stage5importedfuncparam",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "other.go",
        source: `package other

type Hash struct { Name string }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(otherResult.diagnostics).toEqual([]);
    expect(otherResult.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedfuncparam/example.com/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedfuncparam/example.com/root.a") ?? "");
    const otherArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedfuncparam/example.com/other.a") ?? "");
    const otherModule = await importArtifactJavaScript(otherArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const other = await otherModule.instantiateGoJrPackage();
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/other": other.package, "example.com/dep": dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(12n);
  });

  test("resolves qualified imported parameter types through the declaring package import paths", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5qualifiedparampath",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/right/same": [{
          filename: "right.go",
          source: `package same

type Hash uint

const SHA Hash = 10
`
        }],
        "example.com/dep": [{
          filename: "dep.go",
          source: `package dep

import same "example.com/right/same"

var Seen same.Hash
var Callback func() same.Hash

func Register(h same.Hash, f func() same.Hash) {
	Seen = h
	Callback = f
}

func Sum() same.Hash { return Seen + Callback() }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import (
	dep "example.com/dep"
	same "example.com/right/same"
)

func Call() same.Hash {
	dep.Register(same.SHA, func() same.Hash { return same.Hash(2) })
	return dep.Sum()
}
`
      }]
    }, store);
    const wrongResult = buildPackages({
      importPath: "example.com/wrong/same",
      artifactRoot: "/tmp/gojr-stage5qualifiedparampath",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "wrong.go",
        source: `package same

type Hash interface { M() int }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(wrongResult.diagnostics).toEqual([]);
    expect(wrongResult.ok).toBe(true);
    const wrongArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedparampath/example.com/wrong/same.a") ?? "");
    const rightArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedparampath/example.com/right/same.a") ?? "");
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedparampath/example.com/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedparampath/example.com/root.a") ?? "");
    const wrongModule = await importArtifactJavaScript(wrongArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const wrong = await wrongModule.instantiateGoJrPackage();
    const rightModule = await importArtifactJavaScript(rightArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const right = await rightModule.instantiateGoJrPackage();
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage({}, {
      importsByPath: {
        "example.com/wrong/same": wrong.package,
        "example.com/right/same": right.package
      }
    });
    expect((dep.package as { __gojrImportPathsByQualifier?: Record<string, string> }).__gojrImportPathsByQualifier?.same).toBe("example.com/right/same");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, {
      importsByPath: {
        "example.com/wrong/same": wrong.package,
        "example.com/right/same": right.package,
        "example.com/dep": dep.package
      }
    });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(12n);
    expect(Boolean((dep.package.Seen as { __gojrInterface?: boolean }).__gojrInterface)).toBe(false);
  });

  test("retags pointer conversions to imported named wrapper types by package path", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/wrap",
      artifactRoot: "/tmp/gojr-stage5pointerretag",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/base": [{
          filename: "base.go",
          source: `package base

type Setting struct { Name string }

func New(name string) *Setting { return &Setting{Name: name} }
func (s *Setting) Value() string { return "base:" + s.Name }
`
        }]
      },
      files: [{
        filename: "wrap.go",
        source: `package wrap

import base "example.com/base"

type Setting base.Setting

func New(name string) *Setting {
	return (*Setting)(base.New(name))
}

func (s *Setting) Value() string {
	return (*base.Setting)(s).Value()
}

func RoundTrip() string { return New("ok").Value() }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const baseArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5pointerretag/example.com/base.a") ?? "");
    const wrapArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5pointerretag/example.com/wrap.a") ?? "");
    expect(wrapArchive?.javascript).toContain("__gojrConvertPointer(\"*base.Setting\"");

    const baseModule = await importArtifactJavaScript(baseArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const base = await baseModule.instantiateGoJrPackage();
    const wrapModule = await importArtifactJavaScript(wrapArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const wrap = await wrapModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/base": base.package } });
    expect(wrap.diagnostics).toEqual([]);
    expect(await (wrap.package.RoundTrip as () => Promise<string>)()).toBe("base:ok");
  });

  test("tags new imported named type pointers with declaring package path", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5newimportedpointer",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/dep": [{
          filename: "dep.go",
          source: `package dep

type Big struct { N int64 }

func (b *Big) SetInt64(n int64) *Big {
	b.N = n
	return b
}

func (b *Big) Value() int64 { return b.N }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import dep "example.com/dep"

var one = new(dep.Big).SetInt64(1)

func Call() int64 { return one.Value() }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5newimportedpointer/example.com/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5newimportedpointer/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).toContain("__gojrPointerValue(\"dep.Big\"");
    expect(rootArchive?.javascript).toContain("\"example.com/dep\"");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/dep": dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(1n);
  });

  test("preserves imported named result types when local package has same type name", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5qualifiedtypetext",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

type Type struct { N int64 }

func TypeOf(v any) Type { return Type{N: 7} }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

type Type byte

var depType = dep.TypeOf(nil)

func Value() int64 { return depType.N }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedtypetext/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5qualifiedtypetext/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).not.toContain("pkg[\"depType\"] = __gojrIntegerFrom");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Value as () => Promise<bigint>)()).toBe(7n);
  });

  test("converts to imported named types using imported package descriptors", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedconversion",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

type Type int64
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

type Type string

func Converted(x int64) dep.Type { return dep.Type(x) }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedconversion/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedconversion/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).toContain("__gojrConvertDynamicType(\"dep.Type\"");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Converted as (x: bigint) => Promise<bigint>)(12n)).toBe(12n);
  });

  test("infers type arguments for generated imported generic functions", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedgenericinfer",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

func Id[T any](value T) T { return value }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

type Detail struct {
	N int64
}

func Call() int64 {
	v := dep.Id(Detail{N: 9})
	return v.N
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericinfer/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericinfer/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).toContain("__gojrCallImportedFunction(\"dep\", \"Id\"");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(9n);
  });

  test("infers imported generic constraint element types through named slices before conversion", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedgenericnamed",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

func First[S ~[]E, E any](value S) E { return value[0] }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

type Detail struct {
	N int64
}
type Details []Detail

func Call() int64 {
	v := dep.First(Details{{N: 17}})
	return v.N
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericnamed/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericnamed/example.com/root.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(17n);
  });

  test("runs imported generic sort-style calls with function-typed element parameters", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedgenericsort",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

func SortFunc[S ~[]E, E any](x S, cmp func(a, b E) int) {
	n := len(x)
	pdqsortCmpFunc(x, 0, n, cmp)
}

func pdqsortCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) {
	if a < b {
		_ = cmp(data[0], data[0])
	}
}
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

type Detail struct {
	N int
}
type Details []Detail

func Call() {
	values := Details{
		{N: 17}, {N: 29}, {N: 3}, {N: 41}, {N: 5},
		{N: 23}, {N: 7}, {N: 19}, {N: 11}, {N: 13},
		{N: 31}, {N: 37}, {N: 43}, {N: 47}, {N: 53},
		{N: 59}, {N: 61}, {N: 67}, {N: 71}, {N: 73},
	}
	dep.SortFunc(values, func(a, b Detail) int {
		return a.N - b.N
	})
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericsort/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericsort/example.com/root.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    expect(root.diagnostics).toEqual([]);
    await (root.package.Call as () => Promise<void>)();
  });

  test("forwards generated generic type parameters through imported instantiations", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5importedgenericforward",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

import "unsafe"

func SizeOf[T any]() uintptr {
	var value T
	return unsafe.Sizeof(value)
}
`
        }],
        mid: [{
          filename: "mid.go",
          source: `package mid

import "dep"

func Size[T any]() uintptr { return dep.SizeOf[T]() }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "mid"

func Call() uintptr { return mid.Size[byte]() }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericforward/dep.a") ?? "");
    const midArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericforward/mid.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5importedgenericforward/example.com/root.a") ?? "");
    expect(midArchive?.javascript).toContain("__gojrSubstituteType(\"T\", __gojrTypeArgs)");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const midModule = await importArtifactJavaScript(midArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const mid = await midModule.instantiateGoJrPackage({}, { importsByPath: { dep: dep.package } });
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { mid: mid.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<bigint>)()).toBe(1n);
  });

  test("substitutes package-local generic type parameters in generated generic calls", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5localgenericforward",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

type item struct {
	Name string
}

func identity[T any](value T) T {
	return value
}

func first[S ~[]E, E any](values S) E {
	return identity[E](values[0])
}

func Call() string {
	values := []item{{Name: "ok"}}
	return first(values).Name
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5localgenericforward/example.com/root.a") ?? "");
    expect(archive?.javascript).toContain("\"T\": __gojrSubstituteType(\"E\", __gojrTypeArgs)");
    expect(archive?.javascript).toContain("\"E\": \"item\"");
    expect(archive?.javascript).not.toContain("\"E\": \"E\"");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await module.instantiateGoJrPackage();
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<string>)()).toBe("ok");
  });

  test("emits internal abi TypeFor as a generated intrinsic", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage5internalabitypefor",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "internal/abi": [{
          filename: "type.go",
          source: `package abi

type Kind uint8

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
)

type Type struct {
	Size_ uintptr
	Kind_ Kind
}

func TypeFor[T any]() *Type { return nil }
func (t *Type) Size() uintptr { return t.Size_ }
func (t *Type) Kind() Kind { return t.Kind_ }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "internal/abi"

func Call() (uintptr, abi.Kind) {
	t := abi.TypeFor[byte]()
	return t.Size(), t.Kind()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const abiArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5internalabitypefor/internal/abi.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage5internalabitypefor/example.com/root.a") ?? "");
    expect(abiArchive?.javascript).toContain("pkg[\"TypeFor\"] = async (__gojrTypeArgs) => __gojrInternalAbiTypePointer");
    const abiModule = await importArtifactJavaScript(abiArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const abi = await abiModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "internal/abi": abi.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Call as () => Promise<[bigint, bigint]>)()).toEqual([1n, 8n]);
  });

  test("emits internal abi MapType hasher as a generated intrinsic", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-internal-abi-maptype",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "internal/abi": [{
          filename: "abi.go",
          source: `package abi

import "unsafe"

type Type struct {
	Size_ uintptr
}

type MapType struct {
	Type
	Hasher func(unsafe.Pointer, uintptr) uintptr
}

func TypeOf(v any) *Type
func (t *Type) MapType() *MapType
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import (
	"internal/abi"
	"unsafe"
)

func Call() (uintptr, bool) {
	var m map[string]int
	mt := abi.TypeOf(m).MapType()
	key := "ivy"
	hash := mt.Hasher(unsafe.Pointer(&key), 5)
	return hash, hash != 0
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const abiArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-internal-abi-maptype/internal/abi.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-internal-abi-maptype/example.com/root.a") ?? "");
    const abiModule = await importArtifactJavaScript(abiArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const abi = await abiModule.instantiateGoJrPackage();
    expect(abi.diagnostics).toEqual([]);
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "internal/abi": abi.package } });
    expect(root.diagnostics).toEqual([]);
    const resultValue = await (root.package.Call as () => Promise<unknown>)();
    expect(Array.isArray(resultValue)).toBe(true);
    expect((resultValue as unknown[])[1]).toBe(true);
  });

  test("dispatches generated interface method calls through concrete type descriptors", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6iface",
      artifactRoot: "/tmp/gojr-stage6iface",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6iface.go",
        source: `package stage6iface

type Stringer interface {
	String() string
}

type Labeller interface {
	Label() string
}

type Thing struct {
	Name string
}

func (t Thing) String() string { return "thing:" + t.Name }
func (t *Thing) Label() string {
	if t == nil {
		return "nil thing"
	}
	return "ptr:" + t.Name
}
func Use(s Stringer) string { return s.String() }
func UseLabel(l Labeller) string { return l.Label() }
func InterfaceCall() string { return Use(Thing{Name: "ivy"}) }
func PointerValueCall() string {
	t := Thing{Name: "bee"}
	return t.Label()
}
func TypedNilInterfaceIsNil() bool {
	var t *Thing
	var l Labeller = t
	return l == nil
}
func NilInterfaceIsNil() bool {
	var l Labeller
	return l == nil
}
func TypedNilInterfaceCall() string {
	var t *Thing
	return UseLabel(t)
}
func AssertThing(x any) (string, bool) {
	t, ok := x.(Thing)
	if ok {
		return t.Name, ok
	}
	return "", ok
}
func AssertThingCall() (string, bool) { return AssertThing(Thing{Name: "ok"}) }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage6iface/example.com/stage6iface.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.InterfaceCall as () => Promise<string>)()).toBe("thing:ivy");
    expect(await (pkg.PointerValueCall as () => Promise<string>)()).toBe("ptr:bee");
    expect(await (pkg.TypedNilInterfaceIsNil as () => Promise<boolean>)()).toBe(false);
    expect(await (pkg.NilInterfaceIsNil as () => Promise<boolean>)()).toBe(true);
    expect(await (pkg.TypedNilInterfaceCall as () => Promise<string>)()).toBe("nil thing");
    expect(await (pkg.AssertThingCall as () => Promise<[string, boolean]>)()).toEqual(["ok", true]);
  });

  test("lowers method calls on dereferenced pointer selector receivers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6ptrselector",
      artifactRoot: "/tmp/gojr-stage6ptrselector",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6ptrselector.go",
        source: `package stage6ptrselector

type node struct {
	v int
	right *node
}

func (x *node) next() *node {
	return x.right
}

func Call() int {
	root := &node{right: &node{v: 7}}
	pos := &root
	return (*pos).next().v
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6ptrselector/example.com/stage6ptrselector.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<bigint>)()).toBe(7n);
  });

  test("generates package-local type descriptors for reflect-style metadata", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6reflect",
      artifactRoot: "/tmp/gojr-stage6reflect",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Kind int
type StructTag string

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Ptr = Pointer
	Slice
	String
	Struct
	UnsafePointer
)

type Type interface {
	String() string
	Name() string
	PkgPath() string
	Kind() Kind
	NumField() int
	Field(i int) StructField
	Elem() Type
	Implements(u Type) bool
}

type StructField struct {
	Name string
	Type Type
	Tag StructTag
	Index []int
	Anonymous bool
}

func TypeOf(i any) Type
`
        }]
      },
      files: [{
        filename: "stage6reflect.go",
        source: `package stage6reflect

import "reflect"

type Stringer interface {
	String() string
}

type Thing struct {
	Name string
	Count int64
}

func (t Thing) String() string { return t.Name }

func ReflectThing() (string, string, int, string, string, bool) {
	t := reflect.TypeOf(Thing{Name: "ivy", Count: 3})
	f := t.Field(1)
	return t.String(), t.Name(), t.NumField(), f.Name, f.Type.String(), t.Kind() == reflect.Struct
}

func ReflectPointer() (string, string, bool) {
	var t *Thing
	typ := reflect.TypeOf(t)
	return typ.String(), typ.Elem().Name(), typ.Kind() == reflect.Pointer
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage6reflect/example.com/stage6reflect.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).toContain("__gojrTypeDescriptors[\"Thing\"]");
    expect(archive?.javascript).not.toContain("function __gojrReflectPackage");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(Object.keys(pkg.__gojrTypeDescriptors as Record<string, unknown>)).toContain("Thing");
    expect(await (pkg.ReflectThing as () => Promise<[string, string, bigint, string, string, boolean]>)()).toEqual([
      "Thing",
      "Thing",
      2n,
      "Count",
      "int64",
      true
    ]);
    expect(await (pkg.ReflectPointer as () => Promise<[string, string, boolean]>)()).toEqual(["*Thing", "Thing", true]);
  });

  test("generated artifacts prefer intrinsic unsafe over supplied import objects", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6unsafe",
      artifactRoot: "/tmp/gojr-stage6unsafe",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6unsafe.go",
        source: `package stage6unsafe

import "unsafe"

func IntrinsicLen() int {
	buf := []byte{1, 2, 3}
	return len(unsafe.Slice(unsafe.SliceData(buf), 2))
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6unsafe/example.com/stage6unsafe.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage({}, {
      importsByPath: {
        unsafe: {
          SliceData: async () => [9, 9, 9, 9],
          Slice: async () => Array.from({ length: 99 }, () => 9)
        }
      }
    });
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.IntrinsicLen as () => Promise<bigint>)()).toBe(2n);
  });

  test("generated pointer uintptr conversions preserve pointer cells through xor zero", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6ptruintptr",
      artifactRoot: "/tmp/gojr-stage6ptruintptr",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6ptruintptr.go",
        source: `package stage6ptruintptr

import "unsafe"

func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func RoundTrip() int {
	var v int
	p := unsafe.Pointer(&v)
	q := (*int)(NoEscape(p))
	*q = 42
	return v
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6ptruintptr/example.com/stage6ptruintptr.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.RoundTrip as () => Promise<bigint>)()).toBe(42n);
  });

  test("generated unsafe pointer conversions recover zero-offset struct owners", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6ptrfirstfield",
      artifactRoot: "/tmp/gojr-stage6ptrfirstfield",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6ptrfirstfield.go",
        source: `package stage6ptrfirstfield

import "unsafe"

type node struct {
	isEntry bool
}

type embeddedEntry struct {
	node
	key int
}

type namedEntry struct {
	first node
	key int
}

func embedded(n *node) *embeddedEntry {
	return (*embeddedEntry)(unsafe.Pointer(n))
}

func named(n *node) *namedEntry {
	return (*namedEntry)(unsafe.Pointer(n))
}

func RoundTrip() (bool, int, bool, int) {
	e := &embeddedEntry{node: node{isEntry: true}, key: 7}
	n := &namedEntry{first: node{isEntry: true}, key: 11}
	ee := embedded(&e.node)
	nn := named(&n.first)
	return ee.isEntry, ee.key, nn.first.isEntry, nn.key
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6ptrfirstfield/example.com/stage6ptrfirstfield.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.RoundTrip as () => Promise<[boolean, bigint, boolean, bigint]>)()).toEqual([true, 7n, true, 11n]);
  });

  test("generated unsafe pointer conversions expose scalar bytes through fixed byte array pointers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6ptrbytes",
      artifactRoot: "/tmp/gojr-stage6ptrbytes",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6ptrbytes.go",
        source: `package stage6ptrbytes

import "unsafe"

func Probe() (byte, byte, int, uint32) {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	first := b[0]
	b[0] = 2
	s := b[:]
	return first, s[0], len(s), i
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6ptrbytes/example.com/stage6ptrbytes.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Probe as () => Promise<[bigint, bigint, bigint, bigint]>)()).toEqual([1n, 2n, 4n, 2n]);
  });

  test("generated defer and recover follow Go direct deferred-call semantics", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6recover",
      artifactRoot: "/tmp/gojr-stage6recover",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6recover.go",
        source: `package stage6recover

func Recovered() (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = r.(string)
		}
	}()
	panic("caught")
}

func Unnamed() int {
	defer func() { recover() }()
	panic("unnamed")
}

func helper() any { return recover() }

func HelperFails() {
	defer func() { _ = helper() }()
	panic("boom")
}

func DirectRecoverFails() {
	defer recover()
	panic("boom")
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6recover/example.com/stage6recover.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.Recovered as () => Promise<string>)()).toBe("caught");
    expect(await (pkg.Unnamed as () => Promise<bigint>)()).toBe(0n);
    let helperPanic = "";
    try {
      await (pkg.HelperFails as () => Promise<null>)();
    } catch (error) {
      helperPanic = error instanceof Error ? error.message : String(error);
    }
    expect(helperPanic).toContain("panic: boom");
    let directPanic = "";
    try {
      await (pkg.DirectRecoverFails as () => Promise<null>)();
    } catch (error) {
      directPanic = error instanceof Error ? error.message : String(error);
    }
    expect(directPanic).toContain("panic: boom");
  });

  test("generated modern builtins include min max and clear", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6builtins",
      artifactRoot: "/tmp/gojr-stage6builtins",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6builtins.go",
        source: `package stage6builtins

func Builtins() (int, string, float64, int, int, bool, int) {
	lo := min(3, 1, 2)
	hi := max("a", "c", "b")
	f := min(2.5, 1.5)
	xs := []int{4, 5}
	clear(xs)
	m := map[string]int{"a": 1}
	clear(m)
	_, ok := m["a"]
	return lo, hi, f, xs[0], len(m), ok, xs[1]
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6builtins/example.com/stage6builtins.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Builtins as () => Promise<[bigint, string, number, bigint, bigint, boolean, bigint]>)()).toEqual([1n, "c", 1.5, 0n, 0n, false, 0n]);
  });

  test("generated uintptr arguments preserve pointer tokens", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6uintptrarg",
      artifactRoot: "/tmp/gojr-stage6uintptrarg",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6uintptrarg.go",
        source: `package stage6uintptrarg

import "unsafe"

func Take(x uintptr) unsafe.Pointer {
	return unsafe.Pointer(x)
}

func RoundTrip() int {
	var v int
	p := Take(uintptr(unsafe.Pointer(&v)))
	q := (*int)(p)
	*q = 99
	return v
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6uintptrarg/example.com/stage6uintptrarg.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.RoundTrip as () => Promise<bigint>)()).toBe(99n);
  });

  test("generated package globals zero anonymous struct values", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6anonstruct",
      artifactRoot: "/tmp/gojr-stage6anonstruct",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6anonstruct.go",
        source: `package stage6anonstruct

var PPC64 struct {
	IsPOWER8 bool
	IsPOWER9 bool
	Count int
	Names []string
}

func Initial() (bool, bool, int, int) {
	return PPC64.IsPOWER8, PPC64.IsPOWER9, PPC64.Count, len(PPC64.Names)
}

func Set() bool {
	PPC64.IsPOWER9 = true
	PPC64.Count = 7
	return PPC64.IsPOWER9
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6anonstruct/example.com/stage6anonstruct.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Initial as () => Promise<[boolean, boolean, bigint, bigint]>)()).toEqual([false, false, 0n, 0n]);
    expect(await (instantiated.package.Set as () => Promise<boolean>)()).toBe(true);
    expect(await (instantiated.package.Initial as () => Promise<[boolean, boolean, bigint, bigint]>)()).toEqual([false, true, 7n, 0n]);
  });

  test("generated const declarations evaluate iota and inherited expressions", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6iota",
      artifactRoot: "/tmp/gojr-stage6iota",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6iota.go",
        source: `package stage6iota

const (
	A = 1 << iota
	B
	C
)

func Values() (int, int, int) {
	const (
		X = 10 + iota
		Y
	)
	return A, B + X, C + Y
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6iota/example.com/stage6iota.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Values as () => Promise<[bigint, bigint, bigint]>)()).toEqual([1n, 12n, 15n]);
  });

  test("preserves named scalar identity for inherited typed constants across package calls", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/hashuser",
      artifactRoot: "/tmp/gojr-stage6namedconst",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/hashdep": [{
          filename: "hashdep.go",
          source: `package hashdep

type Hash uint

const (
	MD4 Hash = 1 + iota
	MD5
	SHA1
	SHA224
	SHA256
)

var sizes = []uint8{
	MD4:    16,
	SHA256: 32,
}

func (h Hash) Size() int {
	return int(sizes[h])
}
`
        }]
      },
      files: [{
        filename: "hashuser.go",
        source: `package hashuser

import "example.com/hashdep"

func Size() int {
	return hashdep.SHA256.Size()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6namedconst/example.com/hashdep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6namedconst/example.com/hashuser.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/hashdep": dep.package } });
    expect(dep.package.SHA256).toMatchObject({ __gojrNamed: true, __gojrType: "Hash", __gojrPkgPath: "example.com/hashdep" });
    expect(await (root.package.Size as () => Promise<bigint>)()).toBe(32n);
  });

  test("generated package declarations initialize in dependency order", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6declorder",
      artifactRoot: "/tmp/gojr-stage6declorder",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6declorder.go",
        source: `package stage6declorder

const A = 1 + B
const C = A + B
const B = 2

var X = Y + 3
var Y = C

func Values() (int, int, int, int, int) {
	return A, B, C, X, Y
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6declorder/example.com/stage6declorder.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Values as () => Promise<[bigint, bigint, bigint, bigint, bigint]>)()).toEqual([3n, 2n, 5n, 8n, 5n]);
  });

  test("generated package declaration order includes globals referenced through called functions", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6declfuncdeps",
      artifactRoot: "/tmp/gojr-stage6declfuncdeps",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6declfuncdeps.go",
        source: `package stage6declfuncdeps

var ready = alignOf(0)

func alignOf(n int) int {
	align := sizeofPtr
	if n == 0 {
		return align
	}
	return n + align
}

const sizeofPtr = 8

func Ready() int {
	return ready
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6declfuncdeps/example.com/stage6declfuncdeps.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Ready as () => Promise<bigint>)()).toBe(8n);
  });

  test("generated package declaration order includes globals referenced through called methods", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6declmethoddeps",
      artifactRoot: "/tmp/gojr-stage6declmethoddeps",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6declmethoddeps.go",
        source: `package stage6declmethoddeps

type T struct { N int }

var A = new(T).F()
var B = &T{N: 41}

func (t *T) F() *T {
	return B
}

func Ready() int {
	return A.N + 1
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6declmethoddeps/example.com/stage6declmethoddeps.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Ready as () => Promise<bigint>)()).toBe(42n);
  });

  test("emitted packages intrinsicify crypto internal constanttime boolToUint8", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "crypto/internal/constanttime",
      artifactRoot: "/tmp/gojr-stage6-constanttime-intrinsic",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "/usr/local/go/src/crypto/internal/constanttime/constant_time.go",
        source: `package constanttime

func Select(v, x, y int) int {
	v = int(boolToUint8(v != 0))
	return ^(v-1)&x | (v-1)&y
}

func ByteEq(x, y uint8) int {
	return int(boolToUint8(x == y))
}

func Eq(x, y int32) int {
	return int(boolToUint8(x == y))
}

func LessOrEq(x, y int) int {
	return int(boolToUint8(x <= y))
}

func boolToUint8(b bool) uint8 {
	panic("unreachable; must be intrinsicified")
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-constanttime-intrinsic/crypto/internal/constanttime.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.ByteEq as (x: bigint, y: bigint) => Promise<bigint>)(7n, 7n)).toBe(1n);
    expect(await (instantiated.package.ByteEq as (x: bigint, y: bigint) => Promise<bigint>)(7n, 8n)).toBe(0n);
    expect(await (instantiated.package.Eq as (x: bigint, y: bigint) => Promise<bigint>)(9n, 9n)).toBe(1n);
    expect(await (instantiated.package.LessOrEq as (x: bigint, y: bigint) => Promise<bigint>)(9n, 8n)).toBe(0n);
    expect(await (instantiated.package.Select as (v: bigint, x: bigint, y: bigint) => Promise<bigint>)(1n, 41n, 99n)).toBe(41n);
    expect(await (instantiated.package.Select as (v: bigint, x: bigint, y: bigint) => Promise<bigint>)(0n, 41n, 99n)).toBe(99n);
  });

  test("top-level multi-result var declarations split tuples once", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6topvarmultivalue",
      artifactRoot: "/tmp/gojr-stage6-topvarmultivalue",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6topvarmultivalue.go",
        source: `package stage6topvarmultivalue

type T struct { N int }

var Calls int
var A, _ = makeT()
var B = A.N + Calls

func makeT() (*T, error) {
	Calls = Calls + 1
	return &T{N: 41}, nil
}

func Ready() (int, int, int) {
	return A.N, Calls, B
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-topvarmultivalue/example.com/stage6topvarmultivalue.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Ready as () => Promise<[bigint, bigint, bigint]>)()).toEqual([41n, 1n, 42n]);
  });

  test("generated float constants render integer literals as numbers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6floatconst",
      artifactRoot: "/tmp/gojr-stage6floatconst",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6floatconst.go",
        source: `package stage6floatconst

const Ln2 = 0.5
const Log2E = 1 / Ln2
const Max = 2 * (1 + (1 - 0.25))

func Values() (float64, float64) {
	return Log2E, Max
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6floatconst/example.com/stage6floatconst.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Values as () => Promise<[number, number]>)()).toEqual([2, 3.5]);
  });

  test("generated typed integer constants normalize float-containing initializers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6intfloatconst",
      artifactRoot: "/tmp/gojr-stage6intfloatconst",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6intfloatconst.go",
        source: `package stage6intfloatconst

const (
	base = 4
	scale = 10
	mixed int64 = (base*1.5 + 2) * scale
)

func Value() int64 {
	return mixed
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6intfloatconst/example.com/stage6intfloatconst.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Value as () => Promise<bigint>)()).toBe(80n);
  });

  test("generated functions rename JavaScript reserved parameter names", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6reserved",
      artifactRoot: "/tmp/gojr-stage6reserved",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage6reserved.go",
        source: `package stage6reserved

func Swap(old, new int) int {
	return old + new
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6reserved/example.com/stage6reserved.a") ?? "");
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Swap as (oldValue: bigint, newValue: bigint) => Promise<bigint>)(2n, 5n)).toBe(7n);
  });

  test("generated zero values use exact package paths for aliased imported fields", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/sync",
      artifactRoot: "/tmp/gojr-stage6pkgpath-zero",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/internal/sync": [{
          filename: "internal_sync.go",
          source: `package sync

type Mutex struct {
	State int
}
`
        }]
      },
      files: [{
        filename: "sync.go",
        source: `package sync

import isync "example.com/internal/sync"

type Mutex struct {
	mu isync.Mutex
}

var M Mutex

func State() int {
	return M.mu.State
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const internalArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6pkgpath-zero/example.com/internal/sync.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6pkgpath-zero/example.com/sync.a") ?? "");
    const internalModule = await importArtifactJavaScript(internalArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const importsByPath: Record<string, unknown> = {};
    const internalInstantiated = await internalModule.instantiateGoJrPackage({}, { importsByPath });
    expect(internalInstantiated.diagnostics).toEqual([]);
    importsByPath["example.com/internal/sync"] = internalInstantiated.package;
    const rootInstantiated = await rootModule.instantiateGoJrPackage({}, { importsByPath });
    expect(rootInstantiated.diagnostics).toEqual([]);
    importsByPath["example.com/sync"] = rootInstantiated.package;
    expect(await (rootInstantiated.package.State as () => Promise<bigint>)()).toBe(0n);
    const currentMutex = rootInstantiated.package.M as { mu?: { State?: bigint } };
    expect(Boolean(currentMutex.mu)).toBe(true);
    expect(currentMutex.mu?.State).toBe(0n);
    expect(Object.getOwnPropertyDescriptor(currentMutex.mu, "__gojrPkgPath")?.value).toBe("example.com/internal/sync");
  });

  test("generated internal sync HashTrieMap zero values dispatch intrinsic methods", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-internal-sync-hashtrie",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "internal/sync": [{
          filename: "hashtriemap.go",
          source: `package sync

type HashTrieMap[K comparable, V any] struct{}

func (m *HashTrieMap[K, V]) Load(key K) (V, bool) {
	panic("gojr test fallback")
}

func (m *HashTrieMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	panic("gojr test fallback")
}
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import isync "internal/sync"

var m isync.HashTrieMap[string, int]

func Call() (bool, int, bool, int, bool) {
	_, ok1 := m.Load("x")
	stored, loaded := m.LoadOrStore("x", 7)
	second, ok2 := m.Load("x")
	return ok1, stored, loaded, second, ok2
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const internalArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-internal-sync-hashtrie/internal/sync.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-internal-sync-hashtrie/example.com/root.a") ?? "");
    const internalModule = await importArtifactJavaScript(internalArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const importsByPath: Record<string, unknown> = {};
    const internalInstantiated = await internalModule.instantiateGoJrPackage({}, { importsByPath });
    expect(internalInstantiated.diagnostics).toEqual([]);
    importsByPath["internal/sync"] = internalInstantiated.package;
    const rootInstantiated = await rootModule.instantiateGoJrPackage({}, { importsByPath });
    expect(rootInstantiated.diagnostics).toEqual([]);

    expect(await (rootInstantiated.package.Call as () => Promise<unknown>)()).toEqual([false, 7n, false, 7n, true]);
  });

  test("calls function-valued struct fields without method dispatch", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-function-field-call",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

type hasher struct {
	hash func(int, int) int
}

func add(a, b int) int {
	return a + b
}

func Call() int {
	m := &hasher{hash: add}
	return m.hash(3, 4)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-function-field-call/example.com/root.a") ?? "");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(7n);
  });

  test("normalizes package constant array lengths in generated zero values", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-array-length-const-zero",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

type cell struct {
	Value int
}

type table struct {
	cells [n]cell
}

const (
	shift = 2
	n     = 1 << shift
)

func Call() int {
	var t table
	return len(t.cells)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-array-length-const-zero/example.com/root.a") ?? "");
    expect(rootArchive?.javascript).toContain("\"type\":\"[4]cell\"");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(4n);
  });

  test("dispatches intrinsic methods through pointers to fixed-array elements", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-array-atomic-pointer-method",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "sync/atomic": [{
          filename: "atomic.go",
          source: `package atomic

type Pointer[T any] struct{}

func (p *Pointer[T]) Load() *T { panic("gojr test fallback") }
func (p *Pointer[T]) Store(value *T) { panic("gojr test fallback") }
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "sync/atomic"

const n = 2

type node struct {
	Value int
}

type table struct {
	ptrs [n]atomic.Pointer[node]
}

func Call() bool {
	var t table
	slot := &t.ptrs[1]
	if slot.Load() != nil {
		return false
	}
	next := &node{Value: 7}
	slot.Store(next)
	return slot.Load() == next
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-array-atomic-pointer-method/example.com/root.a") ?? "");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(true);
  });

  test("emits weak Pointer as a generated intrinsic", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-weak-intrinsic",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

import "weak"

type node struct {
	Value int
}

func Call() (bool, bool, int) {
	next := &node{Value: 9}
	wp := weak.Make(next)
	zero := weak.Pointer[node]{}
	return wp.Value() == next, zero.Value() == nil, wp.Value().Value
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-weak-intrinsic/example.com/root.a") ?? "");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toEqual([true, true, 9n]);
  });

  test("preserves weak Pointer methods in generic struct fields", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-weak-generic-field",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

import "weak"

type entry[T comparable] struct {
	key weak.Pointer[T]
}

func newEntry[T comparable](key T) *entry[T] {
	k := new(T)
	*k = key
	wp := weak.Make(k)
	return &entry[T]{key: wp}
}

func (e *entry[T]) lookup(key T) bool {
	s := e.key.Value()
	return s != nil && *s == key
}

func Call() bool {
	return newEntry[string]("zone").lookup("zone")
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-weak-generic-field/example.com/root.a") ?? "");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(true);
  });

  test("preserves weak Pointer methods in imported generic struct fields", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-weak-imported-generic-field",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        dep: [{
          filename: "dep.go",
          source: `package dep

import "weak"

type Entry[T comparable] struct {
	key weak.Pointer[T]
}

func NewEntry[T comparable](key T) *Entry[T] {
	k := new(T)
	*k = key
	wp := weak.Make(k)
	return &Entry[T]{key: wp}
}

func (e *Entry[T]) Lookup(key T) bool {
	s := e.key.Value()
	return s != nil && *s == key
}
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "dep"

func Call() bool {
	return dep.NewEntry[string]("zone").Lookup("zone")
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-weak-imported-generic-field/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-weak-imported-generic-field/example.com/root.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const depPackage = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage({}, { importsByPath: { dep: depPackage.package } });
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(true);
  });

  test("promotes generic embedded pointer method receivers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-generic-promoted-pointer-method",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "root.go",
        source: `package root

type inner[T any] struct {
	seed uintptr
	hash func(uintptr) uintptr
}

func newInner[T any]() *inner[T] {
	m := new(inner[T])
	m.seed = 9
	m.hash = func(seed uintptr) uintptr { return seed + 4 }
	return m
}

func (m *inner[T]) Load() uintptr {
	return m.hash(m.seed)
}

type outer[T any] struct {
	*inner[T]
}

func Call() uintptr {
	o := &outer[int]{inner: newInner[int]()}
	return o.Load()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-generic-promoted-pointer-method/example.com/root.a") ?? "");
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await rootModule.instantiateGoJrPackage();
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.Call as () => Promise<unknown>)()).toBe(13n);
  });

  test("generated imported struct literals dispatch methods through receiver package paths", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-imported-method",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/internal/poll": [{
          filename: "fd.go",
          source: `package poll

type FD struct {
	Sysfd int
	IsStream bool
}

func (fd *FD) Init(name string, pollable bool) error {
	if pollable {
		fd.Sysfd = 7
	}
	return nil
}
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "example.com/internal/poll"

type file struct {
	pfd poll.FD
}

type File struct {
	*file
}

var F = &File{file: &file{pfd: poll.FD{Sysfd: 1, IsStream: true}}}

func Run() int {
	F.pfd.Init("file", true)
	return F.pfd.Sysfd
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-imported-method/example.com/internal/poll.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-imported-method/example.com/root.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const importsByPath: Record<string, unknown> = {};
    const dep = await depModule.instantiateGoJrPackage({}, { importsByPath });
    expect(dep.diagnostics).toEqual([]);
    importsByPath["example.com/internal/poll"] = dep.package;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.Run as () => Promise<bigint>)()).toBe(7n);
  });

  test("generates make for imported named slice types through package descriptors", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/root",
      artifactRoot: "/tmp/gojr-stage6-imported-make-slice",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/net": [{
          filename: "ip.go",
          source: `package net

type IP []byte
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package root

import "example.com/net"

func MakeIP() int {
	ip := make(net.IP, 4)
	ip[0] = 12
	return int(ip[0]) + len(ip)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-imported-make-slice/example.com/net.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6-imported-make-slice/example.com/root.a") ?? "");
    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const importsByPath: Record<string, unknown> = {};
    const dep = await depModule.instantiateGoJrPackage({}, { importsByPath });
    expect(dep.diagnostics).toEqual([]);
    importsByPath["example.com/net"] = dep.package;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.MakeIP as () => Promise<bigint>)()).toBe(16n);
  });

  test("resolves imported package type descriptors for reflect-style metadata", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6reflectroot",
      artifactRoot: "/tmp/gojr-stage6reflect-imported",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        "example.com/dep": [{
          filename: "dep.go",
          source: `package dep

type RemoteThing struct {
	Name string
	Count int64
}

func Make() RemoteThing { return RemoteThing{Name: "dep", Count: 9} }
`
        }],
        "example.com/other": [{
          filename: "other.go",
          source: `package other

type RemoteThing struct {
	Other string
}

func Make() RemoteThing { return RemoteThing{Other: "other"} }
`
        }],
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Kind int
type StructTag string

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Ptr = Pointer
	Slice
	String
	Struct
	UnsafePointer
)

type Type interface {
	String() string
	Name() string
	PkgPath() string
	Kind() Kind
	NumField() int
	Field(i int) StructField
	Elem() Type
	Implements(u Type) bool
}

type StructField struct {
	Name string
	Type Type
	Tag StructTag
	Index []int
	Anonymous bool
}

func TypeOf(i any) Type
`
        }]
      },
      files: [{
        filename: "root.go",
        source: `package stage6reflectroot

import (
	d "example.com/dep"
	o "example.com/other"
	"reflect"
)

type RemoteThing struct {
	Local string
}

type Container struct {
	Item d.RemoteThing
	Other o.RemoteThing
	Local RemoteThing
}

func ReflectImported() (string, string, string, int, string, string, bool) {
	t := reflect.TypeOf(d.Make())
	f := t.Field(1)
	return t.String(), t.Name(), t.PkgPath(), t.NumField(), f.Name, f.Type.String(), t.Kind() == reflect.Struct
}

func ReflectImportedField() (string, string, string, string, string, string) {
	t := reflect.TypeOf(Container{Item: d.Make(), Other: o.Make(), Local: RemoteThing{Local: "root"}})
	imported := t.Field(0)
	other := t.Field(1)
	local := t.Field(2)
	return imported.Type.String(), imported.Type.PkgPath(), other.Type.String(), other.Type.PkgPath(), local.Type.String(), local.Type.PkgPath()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const depArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6reflect-imported/example.com/dep.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6reflect-imported/example.com/stage6reflectroot.a") ?? "");
    expect(rootArchive?.pkgdef.runtime).toBeUndefined();
    expect(rootArchive?.javascript).toContain("__gojrTypeDescriptors[\"RemoteThing\"]");
    expect(depArchive?.javascript).toContain("__gojrTypeDescriptors[\"RemoteThing\"]");
    expect(rootArchive?.javascript).not.toContain("runtime.ast");

    const depModule = await importArtifactJavaScript(depArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const dep = await depModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/dep": dep.package } });
    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.ReflectImported as () => Promise<[string, string, string, bigint, string, string, boolean]>)()).toEqual([
      "dep.RemoteThing",
      "RemoteThing",
      "example.com/dep",
      2n,
      "Count",
      "int64",
      true
    ]);
    const otherArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage6reflect-imported/example.com/other.a") ?? "");
    const otherModule = await importArtifactJavaScript(otherArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const other = await otherModule.instantiateGoJrPackage();
    const rootWithBothImports = await rootModule.instantiateGoJrPackage({}, { importsByPath: { "example.com/dep": dep.package, "example.com/other": other.package } });
    expect(await (rootWithBothImports.package.ReflectImportedField as () => Promise<[string, string, string, string, string, string]>)()).toEqual([
      "dep.RemoteThing",
      "example.com/dep",
      "other.RemoteThing",
      "example.com/other",
      "RemoteThing",
      "example.com/stage6reflectroot"
    ]);
  });

  test("generates composite named type descriptors for reflect-style metadata", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6reflectcomposite",
      artifactRoot: "/tmp/gojr-stage6reflect-composite",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Kind int
type StructTag string

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Ptr = Pointer
	Slice
	String
	Struct
	UnsafePointer
)

type Type interface {
	String() string
	Name() string
	PkgPath() string
	Kind() Kind
	NumField() int
	Field(i int) StructField
	Elem() Type
	Key() Type
	Len() int
	NumIn() int
	In(i int) Type
	NumOut() int
	Out(i int) Type
	Implements(u Type) bool
}

type StructField struct {
	Name string
	Type Type
	Tag StructTag
	Index []int
	Anonymous bool
}

func TypeOf(i any) Type
`
        }]
      },
      files: [{
        filename: "composite.go",
        source: `package stage6reflectcomposite

import "reflect"

type Names []string
type Counts map[string]int64
type Pair [2]int64
type Updates chan int64
type Callback func(int64) string

func ReflectComposites() (string, bool, string, bool, string, string, bool, int, string, bool, string, bool, int, string, int, string) {
	var names *Names
	var counts *Counts
	var pair *Pair
	var updates *Updates
	var callback *Callback
	nt := reflect.TypeOf(names).Elem()
	mt := reflect.TypeOf(counts).Elem()
	at := reflect.TypeOf(pair).Elem()
	ct := reflect.TypeOf(updates).Elem()
	ft := reflect.TypeOf(callback).Elem()
	return nt.Name(), nt.Kind() == reflect.Slice, nt.Elem().String(),
		mt.Kind() == reflect.Map, mt.Key().String(), mt.Elem().String(),
		at.Kind() == reflect.Array, at.Len(), at.Elem().String(),
		ct.Kind() == reflect.Chan, ct.Elem().String(),
		ft.Kind() == reflect.Func, ft.NumIn(), ft.In(0).String(), ft.NumOut(), ft.Out(0).String()
}

func ReflectNilableZeros() (bool, bool, bool, bool, string, string, string, string, int64, bool, int, int) {
	var names Names
	var counts Counts
	var updates Updates
	var callback Callback
	missing, ok := counts["missing"]
	return names == nil, counts == nil, updates == nil, callback == nil,
		reflect.TypeOf(names).Name(), reflect.TypeOf(counts).Name(), reflect.TypeOf(updates).Name(), reflect.TypeOf(callback).Name(),
		missing, ok, len(names), cap(names)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage6reflect-composite/example.com/stage6reflectcomposite.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).toContain("__gojrTypeDescriptors[\"Names\"]");
    expect(archive?.javascript).toContain("\"elem\":\"string\"");
    expect(archive?.javascript).toContain("\"key\":\"string\"");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ReflectComposites as () => Promise<[string, boolean, string, boolean, string, string, boolean, bigint, string, boolean, string, boolean, bigint, string, bigint, string]>)()).toEqual([
      "Names",
      true,
      "string",
      true,
      "string",
      "int64",
      true,
      2n,
      "int64",
      true,
      "int64",
      true,
      1n,
      "int64",
      1n,
      "string"
    ]);
    expect(await (pkg.ReflectNilableZeros as () => Promise<[boolean, boolean, boolean, boolean, string, string, string, string, bigint, boolean, bigint, bigint]>)()).toEqual([
      true,
      true,
      true,
      true,
      "Names",
      "Counts",
      "Updates",
      "Callback",
      0n,
      false,
      0n,
      0n
    ]);
  });

  test("keeps generated named map zero values nil until make or literal", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6nilmap",
      artifactRoot: "/tmp/gojr-stage6nilmap",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "nilmap.go",
        source: `package stage6nilmap

type Counts map[string]int64

func NilMapReadRange() (bool, int64, bool, int64) {
	var counts Counts
	var loops int64
	for range counts {
		loops++
	}
	missing, ok := counts["missing"]
	return counts == nil, missing, ok, loops
}

func NilMapWrite() {
	var counts Counts
	counts["missing"] = 1
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage6nilmap/example.com/stage6nilmap.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.NilMapReadRange as () => Promise<[boolean, bigint, boolean, bigint]>)()).toEqual([true, 0n, false, 0n]);
    let error: unknown;
    try {
      await (pkg.NilMapWrite as () => Promise<null>)();
    } catch (err) {
      error = err;
    }
    expect(Boolean(error)).toBe(true);
    expect(String(error)).toContain("assignment to entry in nil map");
  });

  test("generates real pointer cells for new address-of dereference and pointer receivers", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7pointers",
      artifactRoot: "/tmp/gojr-stage7pointers",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "pointers.go",
        source: `package stage7pointers

type Point struct {
	X int64
	Name string
}

func (p *Point) Bump(delta int64) {
	p.X += delta
}

func PointerBasics() (int64, string, int64, string, int64, bool, string) {
	p := new(Point)
	(*p).X = 7
	q := &Point{X: 9, Name: "ok"}
	*q = Point{X: (*q).X + 1, Name: q.Name + "!"}
	field := &q.X
	*field += 2
	v := Point{X: 3}
	v.Bump(4)
	key := new("marker")
	return p.X, p.Name, q.X, q.Name, v.X, p != q, *key
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7pointers/example.com/stage7pointers.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.PointerBasics as () => Promise<[bigint, string, bigint, string, bigint, boolean, string]>)()).toEqual([
      7n,
      "",
      12n,
      "ok!",
      7n,
      true,
      "marker"
    ]);
  });

  test("generates pointer conversions for predeclared interface types", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7predeclaredptr",
      artifactRoot: "/tmp/gojr-stage7predeclaredptr",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "predeclaredptr.go",
        source: `package stage7predeclaredptr

import "unsafe"

var errorPointer = (*error)(nil)
var anyPointer = (*any)(nil)

func PointerToPredeclaredTypes() (bool, bool, string) {
	b := []byte("ok")
	return errorPointer == nil, anyPointer == nil, unsafe.String(&b[0], len(b))
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7predeclaredptr/example.com/stage7predeclaredptr.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("__gojrDeref(pkg[\"error\"])");
    expect(archive?.javascript).not.toContain("__gojrDeref(pkg[\"any\"])");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.PointerToPredeclaredTypes as () => Promise<[boolean, boolean, string]>)()).toEqual([true, true, "ok"]);
  });

  test("keeps pointers to predeclared interfaces as pointers for reflectlite type init", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7reflectliteerrorptr",
      artifactRoot: "/tmp/gojr-stage7reflectliteerrorptr",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "reflectliteerrorptr.go",
        source: `package stage7reflectliteerrorptr

import "internal/reflectlite"

var errorType = reflectlite.TypeOf((*error)(nil)).Elem()

func ErrorTypeInfo() (bool, bool, string) {
	return errorType.Kind() == reflectlite.Interface, errorType.Comparable(), errorType.String()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7reflectliteerrorptr/example.com/stage7reflectliteerrorptr.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("__gojrToInterface(null, \"*error\"");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ErrorTypeInfo as () => Promise<[boolean, boolean, string]>)()).toEqual([true, true, "error"]);
  });

  test("zeros generated fixed byte array fields as indexable byte storage", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7fixedbytearrayzero",
      artifactRoot: "/tmp/gojr-stage7fixedbytearrayzero",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "fixedbytearrayzero.go",
        source: `package stage7fixedbytearrayzero

type Encoding struct {
	decodeMap [256]uint8
	outbuf    [1024 / 4 * 3]byte
}

func FixedByteArrayZero() (uint8, int, int) {
	e := new(Encoding)
	e.decodeMap[65] = 7
	e.outbuf[12] = 9
	return e.decodeMap[65], len(e.decodeMap), int(e.outbuf[12])
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7fixedbytearrayzero/example.com/stage7fixedbytearrayzero.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.FixedByteArrayZero as () => Promise<[bigint, bigint, bigint]>)()).toEqual([7n, 256n, 9n]);
  });

  test("emits empty fixed byte array literals as zero-filled byte storage", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7emptyfixedbytearray",
      artifactRoot: "/tmp/gojr-stage7emptyfixedbytearray",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "emptyfixedbytearray.go",
        source: `package stage7emptyfixedbytearray

func EmptyFixedByteArray() (int, byte, byte) {
	a := [4]byte{}
	a[3] = 9
	return len(a), a[0], a[3]
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7emptyfixedbytearray/example.com/stage7emptyfixedbytearray.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.EmptyFixedByteArray as () => Promise<[bigint, bigint, bigint]>)()).toEqual([4n, 0n, 9n]);
  });

  test("indexes generated Go string byte tables as bytes", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7stringbytetable",
      artifactRoot: "/tmp/gojr-stage7stringbytetable",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stringbytetable.go",
        source: `package stage7stringbytetable

const tab = "" + "\\x00\\x80\\x40\\xc0"

func Pick(x uint16) uint16 {
	return uint16(tab[x>>8]) | uint16(tab[x&3])<<8
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7stringbytetable/example.com/stage7stringbytetable.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.Pick as (x: bigint) => Promise<bigint>)(3n)).toBe(0xc000n);
  });

  test("lowers function-local type aliases in generated conversions", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7localtypealias",
      artifactRoot: "/tmp/gojr-stage7localtypealias",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "localtypealias.go",
        source: `package stage7localtypealias

import "unsafe"

func AliasSizeAndConversion() (uintptr, uint64, float64) {
	type F = float32
	type U uint64
	return unsafe.Sizeof(F(0)), uint64(U(3)), float64(F(2))
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7localtypealias/example.com/stage7localtypealias.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("unsupported Stage 3 conversion to F");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.AliasSizeAndConversion as () => Promise<[bigint, bigint, number]>)()).toEqual([4n, 3n, 2]);
  });

  test("lowers byte and rune as predeclared conversions in generated packages", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7predeclaredaliases",
      artifactRoot: "/tmp/gojr-stage7predeclaredaliases",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "predeclaredaliases.go",
        source: `package stage7predeclaredaliases

func ByteAndRuneConversions(x int) (byte, rune) {
	return byte(x), rune(x + 1)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7predeclaredaliases/example.com/stage7predeclaredaliases.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain(`pkg["byte"]`);
    expect(archive?.javascript).not.toContain(`pkg["rune"]`);

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ByteAndRuneConversions as (x: bigint) => Promise<[bigint, bigint]>)(65n)).toEqual([65n, 66n]);
  });

  test("checks generated reflect Implements with descriptor method sets", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7reflectimplements",
      artifactRoot: "/tmp/gojr-stage7reflectimplements",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Type interface {
	Implements(Type) bool
}

func TypeFor[T any]() Type { return nil }
func PointerTo(Type) Type { return nil }
`
        }]
      },
      files: [{
        filename: "reflectimplements.go",
        source: `package stage7reflectimplements

import "reflect"

type Encoder interface {
	Encode() string
}

type Value struct{}
func (Value) Encode() string { return "" }

type PtrOnly struct{}
func (*PtrOnly) Encode() string { return "" }

type Plain struct{}

func ImplementsCheck() (bool, bool, bool, bool) {
	encoder := reflect.TypeFor[Encoder]()
	return reflect.TypeFor[Value]().Implements(encoder),
		reflect.TypeFor[Plain]().Implements(encoder),
		reflect.TypeFor[PtrOnly]().Implements(encoder),
		reflect.PointerTo(reflect.TypeFor[PtrOnly]()).Implements(encoder)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7reflectimplements/example.com/stage7reflectimplements.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ImplementsCheck as () => Promise<[boolean, boolean, boolean, boolean]>)()).toEqual([true, false, false, true]);
  });

  test("unwraps imported reflect Type interface arguments for PointerTo", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7reflectpointertointerface",
      artifactRoot: "/tmp/gojr-stage7reflectpointertointerface",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Type interface {
	Implements(Type) bool
}

func TypeFor[T any]() Type { return nil }
func PointerTo(Type) Type { return nil }
`
        }]
      },
      files: [{
        filename: "reflectpointertointerface.go",
        source: `package stage7reflectpointertointerface

import "reflect"

type Encoder interface {
	Encode() string
}

type PtrOnly struct{}
func (*PtrOnly) Encode() string { return "" }

func PointerCheck() bool {
	encoder := reflect.TypeFor[Encoder]()
	return reflect.PointerTo(reflect.TypeFor[PtrOnly]()).Implements(encoder)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const reflectArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage7reflectpointertointerface/reflect.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage7reflectpointertointerface/example.com/stage7reflectpointertointerface.a") ?? "");
    const reflectModule = await importArtifactJavaScript(reflectArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const reflectPkg = await reflectModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { reflect: reflectPkg.package } });

    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.PointerCheck as () => Promise<boolean>)()).toBe(true);
  });

  test("copies raw Go string bytes through fixed byte array slices", async () => {
    const store = new MemoryArtifactStore();
    const rawFill = "\\xff".repeat(66);
    const result = buildPackages({
      importPath: "example.com/stage7rawbytecopy",
      artifactRoot: "/tmp/gojr-stage7rawbytecopy",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "rawbytecopy.go",
        source: `package stage7rawbytecopy

type Encoding struct {
	decodeMap [256]uint8
}

const decodeMapInitialize = "${rawFill}"

func CopyRawBytesIntoFixedArray() (uint8, uint8, int) {
	e := new(Encoding)
	n := copy(e.decodeMap[:], decodeMapInitialize)
	return e.decodeMap[0], e.decodeMap[65], n
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7rawbytecopy/example.com/stage7rawbytecopy.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.CopyRawBytesIntoFixedArray as () => Promise<[bigint, bigint, bigint]>)()).toEqual([255n, 255n, 66n]);
  });

  test("merges intrinsic reflect overrides with generated generic exports", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7reflectmerge",
      artifactRoot: "/tmp/gojr-stage7reflectmerge",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

func TypeFor[T any]() string { return "compiled TypeFor" }
func TypeOf(v any) string { return "compiled TypeOf" }
func ExtraFor[T any]() string { return "compiled ExtraFor" }
`
        }]
      },
      files: [{
        filename: "reflectmerge.go",
        source: `package stage7reflectmerge

import "reflect"

func GenericReflectExport() string {
	return reflect.ExtraFor[int]()
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const reflectArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage7reflectmerge/reflect.a") ?? "");
    const rootArchive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gojr-stage7reflectmerge/example.com/stage7reflectmerge.a") ?? "");
    const reflectModule = await importArtifactJavaScript(reflectArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const reflectPkg = await reflectModule.instantiateGoJrPackage();
    const rootModule = await importArtifactJavaScript(rootArchive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const root = await rootModule.instantiateGoJrPackage({}, { importsByPath: { reflect: reflectPkg.package } });

    expect(root.diagnostics).toEqual([]);
    expect(await (root.package.GenericReflectExport as () => Promise<string>)()).toBe("compiled ExtraFor");
  });

  test("defers intrinsic selector methods through generated method dispatch", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7deferintrinsicmethod",
      artifactRoot: "/tmp/gojr-stage7deferintrinsicmethod",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "deferintrinsicmethod.go",
        source: `package stage7deferintrinsicmethod

import "sync"

func DeferUnlock() int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	return 7
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7deferintrinsicmethod/example.com/stage7deferintrinsicmethod.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();

    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.DeferUnlock as () => Promise<bigint>)()).toBe(7n);
  });

  test("dispatches promoted embedded pointer receiver methods", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7promotedmethods",
      artifactRoot: "/tmp/gojr-stage7promotedmethods",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "promotedmethods.go",
        source: `package stage7promotedmethods

type Common struct {
	Id int
}

func (c *Common) id() int { return c.Id }
func (c *Common) setId(id int) { c.Id = id }

type Outer struct {
	Common
}

type hasId interface {
	id() int
	setId(int)
}

func use(x hasId) (int, int) {
	before := x.id()
	x.setId(9)
	return before, x.id()
}

func PromotedMethods() (int, int) {
	o := &Outer{Common: Common{Id: 7}}
	return use(o)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7promotedmethods/example.com/stage7promotedmethods.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();

    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.PromotedMethods as () => Promise<[bigint, bigint]>)()).toEqual([7n, 9n]);
  });

  test("lowers unsigned bitwise complement with Go integer width", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7unsignednotwidth",
      artifactRoot: "/tmp/gojr-stage7unsignednotwidth",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "unsignednotwidth.go",
        source: `package stage7unsignednotwidth

func WordBits() (int, int, uint8) {
	return 32 << (^uint(0) >> 63), 32 << (^uintptr(0) >> 63), ^uint8(0)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7unsignednotwidth/example.com/stage7unsignednotwidth.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();

    expect(instantiated.diagnostics).toEqual([]);
    expect(await (instantiated.package.WordBits as () => Promise<[bigint, bigint, bigint]>)()).toEqual([64n, 64n, 255n]);
  });

  test("generates reflect.Value host methods for basic generated values", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage6reflectvalue",
      artifactRoot: "/tmp/gojr-stage6reflect-value",
      backend: GOJR_STAGE1_BACKEND,
      packageSources: {
        reflect: [{
          filename: "reflect.go",
          source: `package reflect

type Kind int
type StructTag string

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Ptr = Pointer
	Slice
	String
	Struct
	UnsafePointer
)

type Type interface {
	String() string
	Name() string
	PkgPath() string
	Kind() Kind
	NumField() int
	Field(i int) StructField
	Elem() Type
	Implements(u Type) bool
}

type StructField struct {
	Name string
	Type Type
	Tag StructTag
	Index []int
	Anonymous bool
}

type Value struct{}

func TypeOf(i any) Type
func ValueOf(i any) Value
func (v Value) IsValid() bool
func (v Value) IsNil() bool
func (v Value) Kind() Kind
func (v Value) Type() Type
func (v Value) Field(i int) Value
func (v Value) Interface() any
func (v Value) String() string
func (v Value) Int() int64
func (v Value) Bool() bool
`
        }]
      },
      files: [{
        filename: "value.go",
        source: `package stage6reflectvalue

import "reflect"

type Thing struct {
	Name string
	Count int64
	Flag bool
}

func ReflectValueBasics() (bool, bool, bool, string, int64, bool, string, bool) {
	var names []string
	invalid := reflect.ValueOf(nil)
	nilSlice := reflect.ValueOf(names)
	v := reflect.ValueOf(Thing{Name: "ivy", Count: 7, Flag: true})
	name := v.Field(0)
	count := v.Field(1)
	flag := v.Field(2)
	return invalid.IsValid(), nilSlice.IsNil(), v.Kind() == reflect.Struct, v.Type().Name(), count.Int(), flag.Bool(), name.String(), name.Interface() == "ivy"
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage6reflect-value/example.com/stage6reflectvalue.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.pkgdef.runtime).toBeUndefined();
    expect(archive?.javascript).toContain("__gojrReflectValueMethods");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.ReflectValueBasics as () => Promise<[boolean, boolean, boolean, string, bigint, boolean, string, boolean]>)()).toEqual([
      false,
      true,
      true,
      "Thing",
      7n,
      true,
      "ivy",
      true
    ]);
  });

  test("erases generated generic function instantiations to reusable functions", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7generic",
      artifactRoot: "/tmp/gojr-stage7generic",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage7generic.go",
        source: `package stage7generic

func Identity[T any](x T) T { return x }
func Choose[A, B any](a A, b B) B { return b }
func First[T any](xs []T) T { return xs[0] }
func PtrValue[T any](p *T) T { return *p }
func GenericInt() int64 { return Identity[int64](42) }
func GenericString() string { return Choose[int64, string](7, "seven") }
func GenericInferredInt() int64 { return Identity(int64(43)) }
func GenericInferredString() string { return Choose(int64(8), "eight") }
func GenericInferredSlice() string { return First([]string{"alpha", "beta"}) }
func GenericInferredPointer() int64 {
	x := int64(11)
	return PtrValue(&x)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7generic/example.com/stage7generic.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("evaluatePackageArtifact");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.GenericInt as () => Promise<bigint>)()).toBe(42n);
    expect(await (pkg.GenericString as () => Promise<string>)()).toBe("seven");
    expect(await (pkg.GenericInferredInt as () => Promise<bigint>)()).toBe(43n);
    expect(await (pkg.GenericInferredString as () => Promise<string>)()).toBe("eight");
    expect(await (pkg.GenericInferredSlice as () => Promise<string>)()).toBe("alpha");
    expect(await (pkg.GenericInferredPointer as () => Promise<bigint>)()).toBe(11n);
  });

  test("lowers generated generic named type methods by generic receiver base", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7genericmethod",
      artifactRoot: "/tmp/gojr-stage7genericmethod",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage7genericmethod.go",
        source: `package stage7genericmethod

type Box[T any] struct {
	Value T
}

func (b Box[T]) Get() T { return b.Value }
func (b *Box[T]) Set(v T) { b.Value = v }
func (b Box[T]) Zero() T {
	var z T
	return z
}

func GenericBox() (int64, string, int64, string) {
	bi := Box[int64]{Value: 7}
	bs := Box[string]{Value: "ok"}
	bi.Set(9)
	return bi.Get(), bs.Get(), bi.Zero(), bs.Zero()
}

func GenericNamedZero() (int64, string) {
	var bi Box[int64]
	var bs Box[string]
	return bi.Value, bs.Value
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7genericmethod/example.com/stage7genericmethod.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.GenericBox as () => Promise<[bigint, string, bigint, string]>)()).toEqual([9n, "ok", 0n, ""]);
    expect(await (pkg.GenericNamedZero as () => Promise<[bigint, string]>)()).toEqual([0n, ""]);
  });

  test("passes generated generic function type dictionaries for zero values", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7genericdict",
      artifactRoot: "/tmp/gojr-stage7genericdict",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage7genericdict.go",
        source: `package stage7genericdict

func Zero[T any]() T {
	var z T
	return z
}

func Defaults[T any]() []T {
	return make([]T, 2)
}

func GenericDefaults() (int64, string, int64, string) {
	ints := Defaults[int64]()
	strings := Defaults[string]()
	return Zero[int64](), Zero[string](), ints[1], strings[1]
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7genericdict/example.com/stage7genericdict.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).toContain("__gojrTypeArg");
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.GenericDefaults as () => Promise<[bigint, string, bigint, string]>)()).toEqual([0n, "", 0n, ""]);
  });

  test("lowers generated generic type-parameter conversions for unsafe sizeof", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/stage7genericsizeof",
      artifactRoot: "/tmp/gojr-stage7genericsizeof",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "stage7genericsizeof.go",
        source: `package stage7genericsizeof

import "unsafe"

type unrounded uint64

func genericSize[F float32 | float64]() uintptr {
	return unsafe.Sizeof(F(0))
}

func bool2[T ~int | ~uint32 | ~uint64](b bool) T {
	if b {
		return T(1)
	}
	return T(0)
}

func GenericSizeofAndConvert() (uintptr, uintptr, unrounded) {
	return genericSize[float32](), genericSize[float64](), bool2[unrounded](true)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const source = store.writes.get("/tmp/gojr-stage7genericsizeof/example.com/stage7genericsizeof.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    expect(archive?.javascript).not.toContain("runtime.ast");

    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    const pkg = instantiated.package;
    expect(instantiated.diagnostics).toEqual([]);
    expect(await (pkg.GenericSizeofAndConvert as () => Promise<[bigint, bigint, bigint]>)()).toEqual([4n, 8n, 1n]);
  });

  test("lowers deferred spread calls in generated packages", async () => {
    const output: string[] = [];
    const result = buildPackages({
      importPath: "example.com/stage1defer",
      artifactRoot: "/tmp/gojr-stage1defer",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "defer.go",
        source: `package stage1defer

func Hard(f func(...int64), xs []int64) {
	defer f(xs...)
}
`
      }]
    }, new MemoryArtifactStore());

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const store = new MemoryArtifactStore();
    buildPackages({
      importPath: "example.com/stage1defer",
      artifactRoot: "/tmp/gojr-stage1defer-run",
      backend: GOJR_STAGE1_BACKEND,
      files: [{
        filename: "defer.go",
        source: `package stage1defer

func Hard(f func(...int64), xs []int64) {
	defer f(xs...)
}
`
      }]
    }, store);
    const source = store.writes.get("/tmp/gojr-stage1defer-run/example.com/stage1defer.a") ?? "";
    const archive = parseGoJuniorPackageArchive(source);
    const module = await importArtifactJavaScript(archive?.javascript ?? "") as unknown as Stage1ArtifactModule;
    const instantiated = await module.instantiateGoJrPackage();
    await (instantiated.package.Hard as (f: (...values: bigint[]) => Promise<void>, xs: bigint[]) => Promise<null>)(
      async (...values: bigint[]) => { output.push(values.join(",")); },
      [1n, 2n, 3n]
    );
    expect(output).toEqual(["1,2,3"]);
  });
});
