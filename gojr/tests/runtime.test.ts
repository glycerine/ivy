import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { describe, expect, test } from "./testHarness.js";
import {
  cellDependency,
  evaluatePackageSourceFiles,
  evaluationResultToHostPayload,
  evaluateSource,
  evaluateSourceFilesWithPackagesOnNode,
  evaluateSourceFiles,
  evaluateSourcePackageGraph,
  formatGoNode,
  formatGoSource,
  formatReplValue,
  GoJuniorSession,
  parseProgram,
  rangeDependency,
  runMainSourcePackageFiles,
  testSource,
  testSourceFiles,
  testSourceFilesWithPackagesOnNode
} from "../src/index.js";
import { createNodeSourcePackageProvider } from "../src/nodeHost.js";

async function expectRuns(source: string, options = {}) {
  const result = await evaluateSource(source, options);
  expect(result.diagnostics).toEqual([]);
  return result;
}

async function withHostInput<T>(text: string, body: () => Promise<T>): Promise<T> {
  const host = globalThis as typeof globalThis & {
    __gojrReadSync?: (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null) => number;
    __gojrWriteSync?: (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null) => number;
  };
  const previousRead = host.__gojrReadSync;
  const previousWrite = host.__gojrWriteSync;
  const input = new TextEncoder().encode(text);
  let inputOffset = 0;
  host.__gojrReadSync = (fd, buffer, offset, length) => {
    if (fd !== 0) return 0;
    const count = Math.min(length, input.length - inputOffset);
    buffer.set(input.slice(inputOffset, inputOffset + count), offset);
    inputOffset += count;
    return count;
  };
  host.__gojrWriteSync = (_fd, _buffer, _offset, length) => length;
  try {
    return await body();
  } finally {
    if (previousRead) host.__gojrReadSync = previousRead;
    else delete host.__gojrReadSync;
    if (previousWrite) host.__gojrWriteSync = previousWrite;
    else delete host.__gojrWriteSync;
  }
}

describe("Go-junior runtime slice", () => {
  test("evaluates sheet arithmetic with explicit dynamic type assertions", async () => {
    const result = await expectRuns(`
var x = sheet.A1.(int64) + sheet.B1.(int64) * 2
if x > 10 {
  return x
}
return x * 2
`, {
      sheet: {
        A1: 4n,
        B1: 3n
      }
    });

    expect(result.value).toBe(20n);
    expect(result.values).toEqual([20n]);
  });

  test("resolves absolute refs, cross-sheet namespaces, and fmt aliases", async () => {
    const result = await expectRuns(`
import f "fmt"

f.Printf("A1=%v\\n", sheet.$A$1)
return Budget.B2.(int64) + sheet.$A$1.(int64)
`, {
      sheet: {
        A1: 2n
      },
      sheets: {
        Budget: {
          B2: 5n
        }
      }
    });

    expect(result.output).toEqual(["A1=2\n"]);
    expect(result.value).toBe(7n);
  });

  test("evaluates spreadsheet ranges as row-major two-dimensional arrays", async () => {
    const result = await expectRuns(`
rows := sheet.A1:B2
return rows[0][0].(int64) + rows[1][1].(int64), rows
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      }
    });

    expect(result.values?.[0]).toBe(5n);
    expect(result.values?.[1]).toEqual([
      [1n, 2n],
      [3n, 4n]
    ]);
  });

  test("records observed spreadsheet cell and range dependencies", async () => {
    const result = await expectRuns(`
_ = sheet.$A$1
_ = Budget.B2
return sheet.A1:B2
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      },
      sheets: {
        Budget: {
          B2: 5n
        }
      }
    });

    expect(result.observedDeps).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" }),
      cellDependency({ sheet: "Budget", cell: "B2" }),
      rangeDependency("sheet", "A1", "B2")
    ]);
  });

  test("runs defers after the script body in reverse order", async () => {
    const result = await expectRuns(`
import "fmt"

defer fmt.Printf("third")
defer fmt.Printf(" second ")
fmt.Printf("first")
`);

    expect(result.output).toEqual(["first", " second ", "third"]);
  });

  test("keeps function defer stacks scoped and last-in-first-out", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func inner() {
  defer fmt.Printf("inner defer\\n")
  fmt.Printf("inner body\\n")
}
`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func outer() {
  defer fmt.Printf("outer first\\n")
  defer fmt.Printf("outer second\\n")
  inner()
  fmt.Printf("after inner\\n")
}
`)).diagnostics).toEqual([]);

    const result = await session.evaluate("outer()");
    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual([
      "inner body\n",
      "inner defer\n",
      "after inner\n",
      "outer second\n",
      "outer first\n"
    ]);
  });

  test("runs deferred closures before reading named return values", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func f() (x int) {
  defer func() {
    x = 2
  }()
  x = 1
  return
}
`);
    expect(define.diagnostics).toEqual([]);

    const result = await session.evaluate("f()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("runs function defers while panicking", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func boom() {
  defer fmt.Printf("cleanup\\n")
  panic("bad")
}
`)).diagnostics).toEqual([]);

    const result = await session.evaluate("boom()");
    expect(result.output).toEqual(["cleanup\n"]);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_PANIC001");
  });

  test("supports Go recover only from direct deferred function calls", async () => {
    const recovered = await expectRuns(`
func f() (out string) {
  defer func() {
    if r := recover(); r != nil {
      out = r.(string)
    }
  }()
  panic("caught")
}
return f(), recover() == nil
`);

    expect(recovered.values).toEqual(["caught", true]);

    const helper = await evaluateSource(`
func helper() interface{} { return recover() }
func f() {
  defer func() { _ = helper() }()
  panic("boom")
}
f()
`);
    expect(helper.diagnostics).toHaveLength(1);
    expect(helper.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(helper.diagnostics[0]?.message).toContain("boom");

    const directDefer = await evaluateSource(`
func f() {
  defer recover()
  panic("boom")
}
f()
`);
    expect(directDefer.diagnostics).toHaveLength(1);
    expect(directDefer.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(directDefer.diagnostics[0]?.message).toContain("boom");
  });

  test("returns named values and zero unnamed values after recover", async () => {
    const result = await expectRuns(`
func named() (x int) {
  defer func() {
    recover()
    x = 7
  }()
  panic("named")
}

func unnamed() int {
  defer func() { recover() }()
  panic("unnamed")
}

return named(), unnamed()
`);

    expect(result.values).toEqual([7n, 0n]);
  });

  test("formats Go-junior values with fmt %#v", async () => {
    const result = await expectRuns(`
import "fmt"

fmt.Printf("i = %#v\\n", 7)
return fmt.Sprintf("%#v %#v %#v", "x", 1.5, true)
`);

    expect(result.output).toEqual(["i = int64(7)\n"]);
    expect(result.value).toBe(`string("x") float64(1.5) bool(true)`);
  });

  test("formats fmt.Sprint with Go operand spacing", async () => {
    const result = await expectRuns(`
return fmt.Sprint("signed ", 1),
  fmt.Sprint(1, 2),
  fmt.Sprint("a", "b"),
  fmt.Sprint(1, "a", 2),
  fmt.Sprintln("a", 1, "b"),
  string(fmt.Append([]byte("x"), "=", 1)),
  string(fmt.Appendf(nil, "n=%v", 2)),
  string(fmt.Appendln(nil, "a", 1))
`);

    expect(result.values).toEqual(["signed 1", "1 2", "ab", "1a2", "a 1 b\n", "x=1", "n=2", "a 1\n"]);
  });

  test("formats cyclic values without unbounded recursion", async () => {
    const result = await expectRuns(`
type Node struct {
  Next *Node
  Label string
}

n := &Node{Label: "root"}
n.Next = n
return fmt.Sprint(n), fmt.Sprintf("%#v", n), n
`);

    expect(result.values?.[0]).toBe(`&Node{Next:<cycle> Label:root}`);
    expect(result.values?.[1]).toBe(`&Node{Next: <cycle>, Label: string("root")}`);
    expect(formatReplValue(result.values?.[2] ?? null)).toBe(`&Node{Next:<cycle>, Label:"root"}`);
  });

  test("formats fmt.Fprintf to io.Writer-shaped values", async () => {
    const result = await expectRuns(`
type writer struct {
  text string
}

func (w *writer) Write(p []byte) (int, error) {
  w.text += string(p)
  return len(p), nil
}

var w writer
n, err := fmt.Fprintf(&w, "x=%v", 7)
return n, err == nil, w.text
`);

    expect(result.values).toEqual([3n, true, "x=7"]);
  });

  test("formats fmt.Errorf as a Go error value", async () => {
    const result = await expectRuns(`
var e error = fmt.Errorf("bad %v", 3)
err := fmt.Errorf("worse %v", 4)
return e.Error(), fmt.Sprint(e), err.Error()
`);

    expect(result.values).toEqual(["bad 3", "bad 3", "worse 4"]);
  });

  test("formats Go source through the TypeScript go/format port", () => {
    const expected = "func f(a int) int {\n\treturn a + 1\n}";
    expect(formatGoSource("func f(a int)int{return a+1}").trimEnd()).toBe(expected);

    const parsed = parseProgram("func f(a int)int{return a+1}", "format-node.go");
    const declaration = parsed.parsed.file?.declarations[0];
    expect(declaration ? formatGoNode(declaration) : "").toBe(expected);
  });

  test("displays saved formatted source for Go-junior function values", async () => {
    const session = new GoJuniorSession();
    const define = await session.evaluate("func f(a int)int{return a+1}");
    expect(define.diagnostics).toEqual([]);

    const lookup = await session.evaluate("f");
    expect(lookup.diagnostics).toEqual([]);
    expect(formatReplValue(lookup.value ?? null)).toBe("func f(a int) int {\n\treturn a + 1\n}");

    const closure = await expectRuns("f := func(a,b int)int{return a+b}\nreturn f");
    expect(formatReplValue(closure.value ?? null)).toBe("func(a, b int) int {\n\treturn a + b\n}");
  });

  test("automatically imports fmt in scripts and REPL sessions", async () => {
    const script = await expectRuns(`
fmt.Printf("auto %v\\n", 7)
return fmt.Sprintf("ok %v", 8)
`);
    expect(script.output).toEqual(["auto 7\n"]);
    expect(script.value).toBe("ok 8");

    const session = new GoJuniorSession();
    const define = await session.evaluate(`f := func() { fmt.Printf("hiya\\n") }`);
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f()");
    expect(call.diagnostics).toEqual([]);
    expect(call.output).toEqual(["hiya\n"]);
    expect(call.value).toBeNull();
  });

  test("supports importing the testing package", async () => {
    const script = await expectRuns(`
import "testing"

return testing.Short(), testing.Verbose()
`);
    expect(script.values).toEqual([false, false]);

    const session = new GoJuniorSession();
    const define = await session.evaluate(`
import "testing"

func TestThing(t *testing.T) {
  t.Helper()
  t.Logf("x=%v", 1)
  if testing.Short() {
    t.SkipNow()
  }
}
`);
    expect(define.diagnostics).toEqual([]);
  });

  test("supports grouped dot imports of fmt in package source", async () => {
    const pkg = await evaluatePackageSourceFiles([{
      filename: "/workspace/app/app.go",
      source: `package app

import (
  . "fmt"
)

var Message = Sprintf("x=%v", 7)
`
    }], { importPath: "example.com/app" });

    expect(pkg.diagnostics).toEqual([]);
    expect(pkg.package?.Message).toBe("x=7");
  });

  test("supports importing cmp and evaluating its ordering helpers", async () => {
    const script = await expectRuns(`
import "cmp"
import "math"

return cmp.Compare(1, 2),
  cmp.Compare(2, 1),
  cmp.Compare("a", "a"),
  cmp.Less(math.NaN(), 0.0),
  cmp.Compare(math.NaN(), math.NaN()),
  cmp.Or("", "fallback")
`);

    expect(script.values).toEqual([-1n, 1n, 0n, true, 0n, "fallback"]);

    const invalid = await evaluateSource(`
import "cmp"

return cmp.Compare([]int{1}, []int{2})
`);
    expect(invalid.diagnostics.some((diagnostic) =>
      diagnostic.severity === "error" && diagnostic.message.includes("cannot compare")
    )).toBe(true);
  });

  test("REPL mode inspects imported package signatures", async () => {
    const session = new GoJuniorSession();
    const imported = await session.evaluate(`import "iter"`);
    expect(imported.diagnostics).toEqual([]);

    const inspected = await session.evaluate("iter");
    expect(inspected.diagnostics).toEqual([]);
    const text = formatReplValue(inspected.value ?? null);
    expect(text).toContain("package iter");
    expect(text).toContain("type Seq");
    expect(text).toContain("func Pull");
    expect(text).toContain("func Pull2");
  });

  test("supports importing unsafe and evaluating size/alignment helpers", async () => {
    const script = await expectRuns(`
import "unsafe"

var x int64
type box struct {
  x int
}
b := &box{x: 7}
p := unsafe.Pointer(b)
roundTrip := (*box)(p)
var zero unsafe.Pointer
var nilBox *box
nilPointer := unsafe.Pointer(nilBox)
zeroBox := (*box)(unsafe.Pointer(uintptr(0)))
opaqueBox := (*box)(unsafe.Pointer(uintptr(42)))
return unsafe.Sizeof(x), unsafe.Alignof(x), p != nil, roundTrip.x, uintptr(zero), uintptr(p) != 0, uintptr(nilPointer), zeroBox == nil, opaqueBox != nil
`);

    expect(script.values).toEqual([8n, 8n, true, 7n, 0n, true, 0n, true, true]);
  });

  test("supports unsafe string and byte-slice pointer intrinsics", async () => {
    const script = await expectRuns(`
import "unsafe"

buf := []byte{103, 111, 106, 114}
s := unsafe.String(&buf[0], len(buf))
p := unsafe.StringData("gojr")
round := unsafe.String(p, 4)
var nilByte *byte
empty := unsafe.String(nilByte, 0)
return s, round, empty
`);

    expect(script.values?.map(formatReplValue)).toEqual([`"gojr"`, `"gojr"`, `""`]);
  });

  test("supports unsafe pointer reinterpretation used by reflect headers", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/unsafeheaders/main.go",
      source: `package main

import "reflect"
import "unsafe"

type unsafeIntf struct {
  typ unsafe.Pointer
  ptr unsafe.Pointer
}

type unsafeReflectValue struct {
  unsafeIntf
  flag uintptr
}

func header(a any) unsafeIntf {
  eface := *(*unsafeIntf)(unsafe.Pointer(&a))
  return eface
}

func rv4iptr(i any) (v reflect.Value) {
  uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
  uv.unsafeIntf = *(*unsafeIntf)(unsafe.Pointer(&i))
  uv.flag = uintptr(reflect.Ptr)
  return
}

func main() {
  var a int = 12
  h := header(&a)
  v := rv4iptr(&a).Elem()
  if h.typ == nil { panic("nil interface type") }
  if h.ptr == nil { panic("nil interface data") }
  if v.Kind().String() != "int" { panic("bad reflect kind") }
  if v.Int() != 12 { panic("bad reflect int") }
  print("ok\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ok\n"]);
  });

  test("supports importing runtime and evaluating stack/caller helpers", async () => {
    const script = await expectRuns(`
import "runtime"

buf := make([]byte, 64)
n := runtime.Stack(buf, false)
pc, file, line, ok := runtime.Caller(0)
pcs := make([]uintptr, 2)
calls := runtime.Callers(0, pcs)
frames := runtime.CallersFrames(pcs[:calls])
frame, more := frames.Next()
fn := runtime.FuncForPC(0)
return runtime.GOOS, runtime.GOARCH, n > 0, pc, file, line, ok, calls, frame.Function, more, fn.Name()
`);

    expect(script.values).toEqual(["js", "gojr", true, 0n, "", 0n, false, 0n, "", false, ""]);
  });

  test("supports intrinsic internal/reflectlite for standard-library error initialization", async () => {
    const script = await expectRuns(`
import "internal/reflectlite"

t := reflectlite.TypeOf((*error)(nil)).Elem()
xs := []int{3, 1}
swap := reflectlite.Swapper(xs)
swap(0, 1)
return t.Kind() == reflectlite.Interface, t.Comparable(), t.String(), reflectlite.ValueOf(xs).Len(), xs[0], xs[1]
`);

    expect(script.values).toEqual([true, true, "error", 2n, 1n, 3n]);

    const graph = await evaluateSourcePackageGraph([{
      importPath: "errors",
      files: [{
        filename: "/usr/local/go/src/errors/errors.go",
        source: `package errors

import "internal/reflectlite"

func New(text string) error { return &errorString{text} }

type errorString struct { s string }

func (e *errorString) Error() string { return e.s }

var errorType = reflectlite.TypeOf((*error)(nil)).Elem()
`
      }]
    }]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.initializedImportPaths).toEqual(["errors"]);
    expect(Object.keys(graph.packages.errors ?? {})).toContain("New");
  });

  test("preserves imported package wrapper identity for cyclic structs", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "main.go",
      source: `package main

import "example.com/lib"

func main() {
  n := lib.New()
  print(n)
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/lib",
        files: [{
          filename: "lib.go",
          source: `package lib

type Node struct {
  Next *Node
  V int
}

func New() *Node {
  n := &Node{V: 1}
  n.Next = n
  return n
}
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("&example.com/lib.Node{");
    expect(result.output.join("")).toContain("<cycle>");
  });

  test("supports host-resolved syscall/js without running source package files", async () => {
    const script = await expectRuns(`
import "syscall/js"

g := js.Global()
g.Set("__gojr_syscall_js_test", 7)
v := g.Get("__gojr_syscall_js_test")
src := js.ValueOf([]any{65, 66, 67})
dst := make([]byte, 2)
n := js.CopyBytesToGo(dst, src)
return v.Int(), v.Type() == js.TypeNumber, js.TypeNumber.String(), js.ValueOf("hi").String(), n, dst[0], dst[1]
`);

    expect(script.values).toEqual([7n, true, "number", "hi", 2n, 65n, 66n]);
  });

  test("reads host stdin through shared host stdio hooks", async () => {
    await withHostInput("xy\n", async () => {
      const script = await expectRuns(`
import "syscall/js"

buf := js.Global().Get("Uint8Array").New(3)
var got int
cb := js.FuncOf(func(this js.Value, args []js.Value) any {
  got = args[1].Int()
  return nil
})
js.Global().Get("fs").Call("read", 0, buf, 0, 3, nil, cb)
var dst []byte
dst = make([]byte, 3)
js.CopyBytesToGo(dst, buf)
return got, dst[0], dst[1], dst[2]
`);

      expect(script.values).toEqual([3n, 120n, 121n, 10n]);
    });

    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    await withHostInput("gojr\n", async () => {
      const result = await runMainSourcePackageFiles([{
        filename: "/workspace/stdin-syscall/main.go",
        source: `package main

import "syscall"

func main() {
  buf := make([]byte, 5)
  n, err := syscall.Read(0, buf)
  if err != nil {
    panic(err)
  }
  if n != 5 || buf[0] != 'g' || buf[1] != 'o' || buf[2] != 'j' || buf[3] != 'r' || buf[4] != '\\n' {
    panic("bad stdin read")
  }
}
`
      }], { sourcePackageProvider });

      expect(result.diagnostics).toEqual([]);
    });

    await withHostInput("alpha\n", async () => {
      const result = await runMainSourcePackageFiles([{
        filename: "/workspace/stdin-bufio/main.go",
        source: `package main

import (
  "bufio"
  "os"
)

func main() {
  line, err := bufio.NewReader(os.Stdin).ReadString('\\n')
  if err != nil {
    panic(err)
  }
  if line != "alpha\\n" {
    panic(line)
  }
}
`
      }], { sourcePackageProvider });

      expect(result.diagnostics).toEqual([]);
    });
  });

  test("initializes imported sync values and compiled atomic named zero values", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync",
        files: [{
          filename: "/usr/local/go/src/sync/sync.go",
          source: `package sync

type Map struct{}
func (m *Map) Load(key any) (any, bool) { return nil, false }
func (m *Map) LoadOrStore(key, value any) (any, bool) { return value, false }
func (m *Map) Store(key, value any) {}
func (m *Map) Range(f func(key, value any) bool) {}

type Once struct {
  done bool
}
func (o *Once) Do(f func()) {
  if !o.done {
    o.done = true
    f()
  }
}

type Mutex struct{}
func (m *Mutex) Lock() {}
func (m *Mutex) Unlock() {}

type Pool struct{ New func() any }
func (p *Pool) Get() any { return nil }
func (p *Pool) Put(x any) {}
`
        }]
      },
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

func LoadUint32(addr *uint32) uint32
func StoreUint32(addr *uint32, val uint32)
func SwapUint32(addr *uint32, new uint32) uint32
func CompareAndSwapUint32(addr *uint32, old, new uint32) bool

func b32(b bool) uint32 {
  if b {
    return 1
  }
  return 0
}

type Bool struct{ v uint32 }
func (x *Bool) CompareAndSwap(old, new bool) bool { return CompareAndSwapUint32(&x.v, b32(old), b32(new)) }
func (x *Bool) Load() bool { return LoadUint32(&x.v) != 0 }
func (x *Bool) Store(val bool) { StoreUint32(&x.v, b32(val)) }
func (x *Bool) Swap(new bool) bool { return SwapUint32(&x.v, b32(new)) != 0 }

func LoadInt32(addr *int32) int32
func StoreInt32(addr *int32, val int32)
func SwapInt32(addr *int32, new int32) int32
func CompareAndSwapInt32(addr *int32, old, new int32) bool
func AddInt32(addr *int32, delta int32) int32

type Int32 struct{ v int32 }
func (x *Int32) Add(delta int32) int32 { return AddInt32(&x.v, delta) }
func (x *Int32) CompareAndSwap(old, new int32) bool { return CompareAndSwapInt32(&x.v, old, new) }
func (x *Int32) Load() int32 { return LoadInt32(&x.v) }
func (x *Int32) Store(val int32) { StoreInt32(&x.v, val) }
func (x *Int32) Swap(new int32) int32 { return SwapInt32(&x.v, new) }
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "sync"
import "sync/atomic"

var cache sync.Map
var once sync.Once
var mu sync.Mutex
var pool = sync.Pool{New: func() any { return 9 }}
var flag atomic.Bool
var hits atomic.Int32
var Seen any
var Flag bool
var HitCount int32

func init() {
  mu.Lock()
  mu.Unlock()
  flag.Store(true)
  Flag = flag.Load()
  hits.Add(1)
  hits.Store(hits.Add(2))
  _ = hits.CompareAndSwap(3, 5)
  _ = hits.Swap(hits.Add(1))
  HitCount = hits.Load()
  once.Do(func() { cache.Store("x", 4) })
  pool.Put(7)
  _ = pool.Get()
  _ = pool.Get()
  Seen, _ = cache.Load("x")
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(formatReplValue(graph.packages.app?.Seen ?? null)).toBe("4");
    expect(graph.packages.app?.Flag).toBe(true);
    expect(graph.packages.app?.HitCount).toBe(6n);
  });

  test("supports runtime cleanup handles as no-op host values", async () => {
    const script = await expectRuns(`
import "runtime"

type file struct {
  name string
}

f := file{name: "tmp"}
cleanup := runtime.AddCleanup(&f, func(name string) {}, f.name)
cleanup.Stop()
return true
`);

  expect(script.values).toEqual([true]);
  });

  test("panics when unimplemented runtime profiling stubs are called", async () => {
    const script = await expectRuns(`
import "runtime"

func call() (caught any) {
  defer func() {
    caught = recover()
  }()
  runtime.SetCPUProfileRate(100)
  return "not reached"
}

return call().(string) + ""
`);

    expect(script.values).toEqual(["gojr error: runtime.SetCPUProfileRate not implemented"]);
  });

  test("supports importing os.Exit without ambient process authority", async () => {
    const script = await expectRuns(`
import "os"

if false {
  os.Exit(1)
}
return 7
`);
    expect(script.value).toBe(7n);

    const exit = await evaluateSource(`
import "os"

os.Exit(3)
`);
    expect(exit.diagnostics).toEqual([]);
    expect(exit.exitCode).toBe(3);
  });

  test("supports explicit deterministic os.Getenv bindings", async () => {
    const empty = await expectRuns(`
import "os"
return os.Getenv("MISSING")
`);
    expect(empty.value).toBe("");

    const configured = await expectRuns(`
import "os"
return os.Getenv("GOJR_MODE")
`, {
      env: {
        GOJR_MODE: "test"
      }
    });
    expect(configured.value).toBe("test");
  });

  test("supports os.Args for ambient and source-built os packages", async () => {
    const ambient = await expectRuns(`
import "os"
return os.Args[0], os.Args[1]
`, {
      argv: ["gojr-test", "-ambient"]
    });
    expect(ambient.values).toEqual(["gojr-test", "-ambient"]);

    const sourceBuilt = await runMainSourcePackageFiles([{
      filename: "/workspace/cmd/app/main.go",
      source: `package main

import "os"

func main() {
  print(os.Args[0] + "|" + os.Args[1] + "\\n")
}
`
    }], {
      importPath: "example.com/app",
      argv: ["app", "-source"],
      sourcePackages: [{
        importPath: "os",
        files: [{
          filename: "/workspace/os/os.go",
          source: "package os\n\nvar Args []string\n"
        }]
      }]
    });

    expect(sourceBuilt.diagnostics).toEqual([]);
    expect(sourceBuilt.output).toEqual(["app|-source\n"]);
  });

  test("supports importing math.NaN for float map keys and clear", async () => {
    const script = await expectRuns(`
import "math"

m := make(map[float64]int)
m[math.NaN()] = 1
m[math.NaN()] = 2
before := len(m)
clear(m)
return before, len(m), math.Floor(math.Exp(math.Log(3.5)))
`);

    expect(script.values).toEqual([2n, 0n, 3]);
  });

  test("supports strconv.Itoa and math float bit helpers", async () => {
    const script = await expectRuns(`
import "math"
import "strconv"

f32 := math.Float32frombits(1 << 31)
f64 := math.Float64frombits(1 << 63)
quoted, quoteErr := strconv.Unquote(\`"hi\\n"\`)
raw, rawErr := strconv.Unquote("\`there\`")
_, badErr := strconv.Unquote("not quoted")
return strconv.Itoa(-12), strconv.Itoa(34), quoted, quoteErr == nil, raw, rawErr == nil, badErr != nil, math.Float32bits(f32), math.Float64bits(f64), math.IsNaN(math.NaN()), math.MaxFloat32 > 1e38, math.MaxFloat64 > 1e300, math.MaxInt32, math.MaxUint16, math.MaxUint64
`);

    expect(script.values).toEqual(["-12", "34", "hi\n", true, "there", true, true, 2147483648n, 9223372036854775808n, true, true, true, 2147483647n, 65535n, 18446744073709551615n]);
  });

  test("matches Go float map-key semantics for signed zero and NaN", async () => {
    const script = await expectRuns(`
import "math"

positiveZero := 0.0
negativeZero := math.Float64frombits(1 << 63)
m := map[float64]string{positiveZero: "+0"}
before := m[negativeZero]
m[negativeZero] = "-0"

nanA := math.NaN()
nanB := math.Float64frombits(math.Float64bits(nanA) ^ 2)
m[nanA] = "nan-a"
m[nanB] = "nan-b"
_, okA := m[nanA]
_, okB := m[nanB]
return before, m[positiveZero], len(m), okA, okB, math.IsNaN(nanB)
`);

    expect(script.values).toEqual(["+0", "-0", 3n, false, false, true]);
  });

  test("converts map lookup and delete keys to the declared key type", async () => {
    const script = await expectRuns(`
mf := map[float64]string{0: "zero", 1.0: "one"}
before := mf[0]
_, okBefore := mf[1]
delete(mf, 1)
_, okAfter := mf[1.0]

mi := map[int]string{0.0: "int-zero"}
return before, okBefore, okAfter, mi[0.0]
`);

    expect(script.values).toEqual(["zero", true, false, "int-zero"]);
  });

  test("compares NaN values with Go ordered-comparison semantics", async () => {
    const script = await expectRuns(`
import "math"

nan := math.NaN()
f := 1.0
return nan == nan, nan != nan, nan < nan, nan <= nan, nan > nan, nan >= nan,
  f < nan, f <= nan, f > nan, f >= nan,
  nan < f, nan <= f, nan > f, nan >= f
`);

    expect(script.values).toEqual([
      false, true, false, false, false, false,
      false, false, false, false,
      false, false, false, false
    ]);
  });

  test("evaluates source packages and lets formulas import their exported runtime values", async () => {
    const pkg = await evaluatePackageSourceFiles([{
      filename: "counter.go",
      source: `package counter

var Count int

func init() {
  Count = 40
}

func Next() int {
  Count++
  return Count
}
`
    }], { importPath: "example.com/counter" });

    expect(pkg.diagnostics).toEqual([]);
    expect(pkg.package).toBeDefined();

    const first = await expectRuns(`
import counter "example.com/counter"
return counter.Next()
`, {
      packages: {
        "example.com/counter": pkg.package ?? {}
      }
    });
    const second = await expectRuns(`
import counter "example.com/counter"
return counter.Next()
`, {
      packages: {
        "example.com/counter": pkg.package ?? {}
      }
    });

    expect(first.value).toBe(41n);
    expect(second.value).toBe(42n);
  });

  test("source package graph resolves Go-junior host stubs before provider source", async () => {
    const loaded: string[] = [];
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/app/app.go",
        source: `package app

import "github.com/gopherjs/gopherjs/js"

var Object = js.Global.Get("Object")
var Keys = js.Keys(Object)
`
      }]
    }], {
      sourcePackageProvider: {
        load(importPath: string) {
          loaded.push(importPath);
          if (importPath === "github.com/gopherjs/gopherjs/js") {
            return [{
              filename: "/workspace/gopherjs/js/js.go",
              source: "package js\n\nconst _ = 1 / 0\n"
            }];
          }
          return undefined;
        }
      }
    });

    expect(graph.diagnostics).toEqual([]);
    expect(loaded).toEqual([]);
    expect(graph.packages["github.com/gopherjs/gopherjs/js"]).toBeDefined();
    expect(graph.packages["example.com/app"]).toBeDefined();
  });

  test("zeroes cached imported time.Time values without package contexts", async () => {
    const timePkg = await evaluatePackageSourceFiles([{
      filename: "/workspace/time/time.go",
      source: `package time

type Location struct{}

type Time struct {
  wall uint64
  ext int64
  loc *Location
}

type Duration int64

func (t Time) Unix() int64 { return 0 }
func (t Time) UnixNano() int64 { return 0 }
func (t Time) UTC() Time { return t }
func (t Time) In(loc *Location) Time { return t }
func (t Time) Sub(u Time) Duration { return 0 }
func (t Time) Format(layout string) string { return "" }
`
    }], { importPath: "time" });
    expect(timePkg.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "time"

var T time.Time
var Zero = time.Time{}
var Ptr = new(time.Time)

_ = Zero.Unix()
_ = Zero.UnixNano()
_ = Zero.UTC()
_ = Zero.In(nil)
_ = Zero.Sub(Zero)
_ = Zero.Format("")
_ = Ptr.Unix()

return true
`, {
      packages: { time: timePkg.package ?? {} },
      packageInfos: { time: timePkg.packageInfo! }
    });

    expect(result.values).toEqual([true]);
  });

  test("keys runtime packages and type identity by full import path when package names collide", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/leftpkg",
        files: [{
          filename: "/workspace/left/left.go",
          source: `package dup

type T struct { N int }

func New(n int) *T { return &T{N: n} }

func (t *T) Value() int { return t.N }
`
        }]
      },
      {
        importPath: "example.org/rightpkg",
        files: [{
          filename: "/workspace/right/right.go",
          source: `package dup

type T struct { S string }

func New(s string) *T { return &T{S: s} }

func (t *T) Value() string { return t.S }
`
        }]
      },
      {
        importPath: "example.net/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import (
  left "example.com/leftpkg"
  right "example.org/rightpkg"
)

type Holder struct {
  L *left.T
  R *right.T
}

var H = Holder{L: left.New(7), R: right.New("ok")}

func Values() (int, string) {
  return H.L.Value(), H.R.Value()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.dup).toBeUndefined();
    expect(Object.keys(graph.packages).sort()).toEqual([
      "example.com/leftpkg",
      "example.net/app",
      "example.org/rightpkg"
    ]);

    const app = await expectRuns(`
import app "example.net/app"
x, y := app.Values()
return x, y
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(app.values).toEqual([7n, "ok"]);

    const packageClauseBinding = await expectRuns(`
import "example.com/leftpkg"
return dup.New(3).Value()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(packageClauseBinding.value).toBe(3n);
  });

  test("uses declaring package type context for imported function parameters", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/syscallish",
        files: [{
          filename: "/workspace/syscallish/stat.go",
          source: `package syscallish

type Stat_t struct { N int }

func Stat(name string, st *Stat_t) error {
  st.N = len(name)
  return nil
}
`
        }]
      },
      {
        importPath: "example.com/osish",
        files: [{
          filename: "/workspace/osish/os.go",
          source: `package osish

import sc "example.com/syscallish"

func Read() int {
  var st sc.Stat_t
  err := sc.Stat("hello", &st)
  if err != nil {
    return -1
  }
  return st.N
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import osish "example.com/osish"
return osish.Read()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(5n);
  });

  test("uses declaring package type context for imported function tuple results", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/props",
        files: [{
          filename: "/workspace/props/props.go",
          source: `package props

type Properties struct { Entry uint8 }

func Lookup() (Properties, int) {
  return Properties{Entry: 11}, 1
}
`
        }]
      },
      {
        importPath: "example.com/rules",
        files: [{
          filename: "/workspace/rules/rules.go",
          source: `package rules

import "example.com/props"

func Entry() int {
  p, _ := props.Lookup()
  var q props.Properties = p
  return int(q.Entry)
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/rules"
return rules.Entry()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(11n);
  });

  test("preserves caller-owned local types through imported generic callbacks", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/helper",
        files: [{
          filename: "/workspace/helper/helper.go",
          source: `package helper

func Visit[S ~[]E, E any](items S, fn func(E)) {
  fn(items[0])
}
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "example.com/helper"

type Flag struct { Name string }

func Run() string {
  items := []*Flag{{Name: "ok"}}
  got := ""
  helper.Visit(items, func(flag *Flag) {
    got = flag.Name
  })
  return got
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/app"
return app.Run()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe("ok");
  });

  test("preserves imported uint64 tuple result types above int64 range", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/words",
        files: [{
          filename: "/workspace/words/words.go",
          source: `package words

func Pair() (uint64, uint64) {
  return uint64(128), uint64(10254876495507714224)
}
`
        }]
      },
      {
        importPath: "example.com/user",
        files: [{
          filename: "/workspace/user/user.go",
          source: `package user

import "example.com/words"

func Values() (uint64, uint64) {
  hi, lo := words.Pair()
  return hi, lo
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/user"
hi, lo := user.Values()
return hi, lo
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([128n, 10254876495507714224n]);
  });

  test("constructs imported struct aliases with composite literals", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/base",
        files: [{
          filename: "/workspace/base/base.go",
          source: `package base

type PathError struct {
  Op string
  Path string
}

func (e *PathError) Error() string {
  return e.Op + " " + e.Path
}
`
        }]
      },
      {
        importPath: "example.com/alias",
        files: [{
          filename: "/workspace/alias/alias.go",
          source: `package alias

import "example.com/base"

type PathError = base.PathError

func New() *PathError {
  return &PathError{Op: "open", Path: "file"}
}

func Message() string {
  var err error = New()
  return err.Error()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/alias"
err := alias.New()
return err.Op, err.Path, alias.Message()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual(["open", "file", "open file"]);
  });

  test("retags concrete type assertion results across package boundaries", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/internal/strconv",
        files: [{
          filename: "/workspace/example.com/internal/strconv/atoi.go",
          source: `package strconv

type NumError struct {
  Num string
}

func (e *NumError) Error() string {
  return e.Num
}

func ParseInt(s string) (int, error) {
  return 0, &NumError{Num: s}
}
`
        }]
      },
      {
        importPath: "example.com/strconv",
        files: [{
          filename: "/workspace/example.com/strconv/atoi.go",
          source: `package strconv

import strconv "example.com/internal/strconv"

func Check() string {
  _, err := strconv.ParseInt("567")
  e := err.(*strconv.NumError)
  return e.Num
}

func CheckSwitch() string {
  _, err := strconv.ParseInt("890")
  switch e := err.(type) {
  case *strconv.NumError:
    return e.Num
  default:
    return "bad"
  }
}

func CheckPresence() (string, bool) {
  _, err := strconv.ParseInt("321")
  e, ok := err.(*strconv.NumError)
  return e.Num, ok
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/strconv"
s, ok := strconv.CheckPresence()
return strconv.Check(), strconv.CheckSwitch(), s, ok
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual(["567", "890", "321", true]);
  });

  test("matches imported pointer receiver methods against interface alias result types", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/fs",
        files: [{
          filename: "/workspace/fs/fs.go",
          source: `package fs

type FileMode uint32

type FileInfo interface {
  Mode() FileMode
}
`
        }]
      },
      {
        importPath: "example.com/os",
        files: [{
          filename: "/workspace/os/os.go",
          source: `package os

import fs "example.com/fs"

type FileMode = fs.FileMode
type FileInfo = fs.FileInfo

type fileStat struct {
  mode FileMode
}

func (f *fileStat) Mode() FileMode {
  return f.mode
}

func Stat() FileInfo {
  fi := &fileStat{mode: 7}
  return fi
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/os"
fi := os.Stat()
return fi.Mode()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(7n);
  });

  test("matches package-info-only pointer receiver methods against interface alias result types", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/fs",
        files: [{
          filename: "/workspace/fs/fs.go",
          source: `package fs

type FileMode uint32

type FileInfo interface {
  Mode() FileMode
}
`
        }]
      },
      {
        importPath: "example.com/os",
        files: [{
          filename: "/workspace/os/os.go",
          source: `package os

import fs "example.com/fs"

type FileMode = fs.FileMode
type FileInfo = fs.FileInfo

type fileStat struct {
  mode FileMode
}

func (f *fileStat) Mode() FileMode {
  return f.mode
}

func Stat() FileInfo {
  fi := &fileStat{mode: 7}
  return fi
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/os"
fi := os.Stat()
return fi.Mode()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });

    expect(result.value).toBe(7n);
  });

  test("externalizes imported map element types without copying map identity", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/tree",
        files: [{
          filename: "/workspace/tree/tree.go",
          source: `package tree

type Tree struct {
  Name string
}

var Saved map[string]*Tree

func Make() map[string]*Tree {
  Saved = map[string]*Tree{"one": &Tree{Name: "one"}}
  return Saved
}

func Count() int {
  return len(Saved)
}

func Lookup(name string) *Tree {
  return Saved[name]
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/tree"

var trees map[string]*tree.Tree
trees = tree.Make()
trees["two"] = &tree.Tree{Name: "two"}
return trees["one"].Name, trees["two"].Name, tree.Count(), tree.Lookup("two").Name
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual(["one", "two", 2n, "two"]);
  });

  test("keeps named array pointer types for same-package calls", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/namedarray",
        files: [{
          filename: "/workspace/namedarray/table.go",
          source: `package namedarray

type Table [4]uint32

func Fill(t *Table) {
  t[2] = 7
}

func Make() *Table {
  t := new(Table)
  Fill(t)
  return t
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import namedarray "example.com/namedarray"
return namedarray.Make()[2]
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(7n);
  });

  test("recognizes imported qualified type names in new calls", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/field",
        files: [{
          filename: "/workspace/field/field.go",
          source: `package field

type Element struct { N int }

func (v *Element) One() *Element {
  v.N = 1
  return v
}
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import f "example.com/field"

var X = new(f.Element).One()

func Value() int {
  return X.N
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import app "example.com/app"
return app.Value()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(1n);
  });

  test("distributes grouped var initializers and keeps package init dependencies", async () => {
    const local = await expectRuns(`
func pair() (int, string) { return 3, "ok" }
var a, b = pair()
return a, b
`);
    expect(local.values).toEqual([3n, "ok"]);

    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/initvars",
        files: [{
          filename: "/workspace/initvars/initvars.go",
          source: `package initvars

type T struct { N int }

var A, _ = new(T).F()
var B = &T{N: 41}

func (t *T) F() (*T, error) {
  return B, nil
}

func Value() int {
  return A.N + 1
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import initvars "example.com/initvars"
return initvars.Value()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(42n);
  });

  test("intrinsicifies crypto internal constanttime boolToUint8", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "crypto/internal/constanttime",
        files: [{
          filename: "/usr/local/go/src/crypto/internal/constanttime/constant_time.go",
          source: `package constanttime

func ByteEq(x, y uint8) int {
  return int(boolToUint8(x == y))
}

func boolToUint8(b bool) uint8 {
  panic("unreachable; must be intrinsicified")
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import constanttime "crypto/internal/constanttime"
return constanttime.ByteEq(7, 7), constanttime.ByteEq(7, 8)
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([1n, 0n]);
  });

  test("intrinsicifies internal abi EscapeNonString as an escape-analysis no-op", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/escape.go",
          source: `package abi

func Touch[T any](v T) T {
  EscapeNonString(v)
  return EscapeToResultNonString(v)
}

func EscapeNonString[T any](v T) {
  panic("intrinsic")
}

func EscapeToResultNonString[T any](v T) T {
  EscapeNonString(v)
  return *new(T)
}

func Escape[T any](v T) T {
  return v
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import abi "internal/abi"
type detail struct {
  ok bool
}
value := abi.EscapeToResultNonString(detail{ok: true})
escaped := abi.Escape(&value)
return abi.Touch("ok"), value.ok, escaped.ok
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual(["ok", true, true]);
  });

  test("preserves caller type identity through internal abi generic result pointers", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/escape.go",
          source: `package abi

func EscapeToResultNonString[T any](v T) T {
  panic("intrinsic")
}
`
        }]
      },
      {
        importPath: "example.com/canon",
        files: [{
          filename: "/workspace/canon/canon.go",
source: `package canon

import (
  abi "internal/abi"
  "unsafe"
)

type canonMap[T comparable] struct{}
type cloneSeq struct {
  stringOffsets []uintptr
}
var singleStringClone = cloneSeq{stringOffsets: []uintptr{0}}

func (m *canonMap[T]) Load(key T) *T {
  return nil
}

func (m *canonMap[T]) LoadOrStore(key T) *T {
  return &key
}

func clone[T comparable](value T) T {
  for _, offset := range singleStringClone.stringOffsets {
    ps := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&value)) + offset))
    *ps = *ps
  }
  return abi.EscapeToResultNonString(value)
}

func Make[T comparable](value T) *T {
  m := &canonMap[T]{}
  ptr := m.Load(value)
  if ptr == nil {
    ptr = m.LoadOrStore(clone(value))
  }
  return ptr
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/canon"
type detail struct {
  zone string
}
ptr := canon.Make(detail{zone: "z"})
return ptr.zone
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.values).toEqual(["z"]);
  });

  test("intrinsicifies internal abi NoEscape as pointer identity", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/escape.go",
          source: `package abi

import "unsafe"

func NoEscape(p unsafe.Pointer) unsafe.Pointer {
  x := uintptr(p)
  return unsafe.Pointer(x ^ 0)
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import abi "internal/abi"
import "unsafe"

type Builder struct {
  addr *Builder
  buf []byte
}

func (b *Builder) copyCheck() {
  if b.addr == nil {
    b.addr = (*Builder)(abi.NoEscape(unsafe.Pointer(b)))
  } else if b.addr != b {
    panic("copied")
  }
}

func (b *Builder) WriteByte(c byte) {
  b.copyCheck()
  b.buf = append(b.buf, c)
}

var b Builder
b.WriteByte(1)
b.WriteByte(2)
return len(b.buf)
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(2n);
  });

  test("runs multiple package init functions in source file sequence order", async () => {
    const pkg = await evaluatePackageSourceFiles([
      {
        filename: "a.go",
        source: `package initseq

var Events []string

func init() {
  Events = append(Events, "a1")
}

func init() {
  Events = append(Events, "a2")
}
`
      },
      {
        filename: "b.go",
        source: `package initseq

func init() {
  Events = append(Events, "b1")
}
`
      }
    ], { importPath: "example.com/initseq" });

    expect(pkg.diagnostics).toEqual([]);
    expect(formatReplValue(pkg.package?.Events ?? null)).toBe(`["a1" "a2" "b1"]`);

    const result = await expectRuns(`
import initseq "example.com/initseq"
return initseq.Events
`, {
      packages: {
        "example.com/initseq": pkg.package ?? {}
      },
      packageInfos: {
        "example.com/initseq": pkg.packageInfo!
      }
    });

    expect(formatReplValue(result.value ?? null)).toBe(`["a1" "a2" "b1"]`);
  });

  test("initializes source package graph in Go spec order and initializes shared imports only once", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/right",
        files: [{
          filename: "right.go",
          source: `package right

import (
  shared "example.com/shared"
  trace "example.com/trace"
)

var Value = shared.Value + 2

func init() {
  _ = shared.Value
  trace.Add("right")
}
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "app.go",
          source: `package app

import (
  left "example.com/left"
  right "example.com/right"
  trace "example.com/trace"
)

func init() {
  _ = left.Value
  _ = right.Value
  trace.Add("app")
}
`
        }]
      },
      {
        importPath: "example.com/trace",
        files: [{
          filename: "trace.go",
          source: `package trace

var Events []string

func init() {
  Events = append(Events, "trace")
}

func Add(event string) {
  Events = append(Events, event)
}

func Snapshot() []string {
  return Events
}
`
        }]
      },
      {
        importPath: "example.com/left",
        files: [{
          filename: "left.go",
          source: `package left

import (
  shared "example.com/shared"
  trace "example.com/trace"
)

var Value = shared.Value + 1

func init() {
  trace.Add("left")
}
`
        }]
      },
      {
        importPath: "example.com/shared",
        files: [{
          filename: "shared.go",
          source: `package shared

import trace "example.com/trace"

var Value = 40

func init() {
  trace.Add("shared")
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.initializedImportPaths).toEqual([
      "example.com/trace",
      "example.com/shared",
      "example.com/left",
      "example.com/right",
      "example.com/app"
    ]);

    const result = await expectRuns(`
import trace "example.com/trace"
return trace.Snapshot()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });

    expect(formatReplValue(result.value ?? null)).toBe(`["trace" "shared" "left" "right" "app"]`);
  });

  test("supports bodyless runtime intrinsics during package initialization", async () => {
    const result = await evaluateSourcePackageGraph([{
      importPath: "clock",
      files: [{
        filename: "clock.go",
        source: `package clock

func runtimeNano() int64
func runtimeNow() (sec int64, nsec int32, mono int64)

var Start = runtimeNano() - 1

func Grab() (int64, int32, int64) {
  return runtimeNow()
}
`
      }]
    }]);

    expect(result.diagnostics).toEqual([]);
    expect(result.packages.clock?.Start).toBe(0n);

    const script = await expectRuns(`
import "clock"
sec, nsec, mono := clock.Grab()
return sec, nsec, mono
`, {
      packages: result.packages,
      packageInfos: result.packageInfos
    });
    expect(script.values).toEqual([0n, 0n, 1n]);
  });

  test("supports bodyless internal bytealg intrinsics selected by native standard-library files", async () => {
    const result = await evaluateSourcePackageGraph([{
      importPath: "internal/bytealg",
      files: [{
        filename: "bytealg.go",
        source: `package bytealg

func IndexByte(b []byte, c byte) int
func IndexByteString(s string, c byte) int
func Index(a, b []byte) int
func IndexString(a, b string) int
func Count(b []byte, c byte) int
func CountString(s string, c byte) int
func Compare(a, b []byte) int
func abigen_runtime_cmpstring(a, b string) int
func MakeNoZero(n int) []byte

func CompareString(a, b string) int {
  return abigen_runtime_cmpstring(a, b)
}

var Results = []int{
  IndexByte([]byte("abc"), 'b'),
  IndexByteString("abc", 'c'),
  Index([]byte("banana"), []byte("na")),
  IndexString("banana", "nan"),
  Count([]byte("banana"), 'a'),
  CountString("banana", 'n'),
  Compare([]byte("a"), []byte("b")),
  CompareString("b", "a"),
  len(MakeNoZero(3)),
}
`
      }]
    }]);

    expect(result.diagnostics).toEqual([]);
    expect(result.packages["internal/bytealg"]?.Results).toEqual([
      1n, 2n, 2n, 2n, 3n, 2n, -1n, 1n, 3n
    ]);
  });

  test("time.LoadLocation uses deterministic loaded timezone package data", async () => {
    const result = await evaluateSourcePackageGraph([{
      importPath: "time",
      files: [{
        filename: "time.go",
        source: `package time

type Location struct {
  name string
}

var UTC = &Location{name: "UTC"}
var Local = UTC

func LoadLocation(name string) (*Location, error) {
  panic("compiled time.LoadLocation body should not run")
}

func (l *Location) String() string {
  return l.name
}
`
      }]
    }, {
      importPath: "4d63.com/tz",
      files: [{
        filename: "tz.go",
        source: `package tz

import "time"

func LoadLocation(name string) (*time.Location, error) {
  return time.UTC, nil
}
`
      }]
    }, {
      importPath: "app",
      files: [{
        filename: "app.go",
        source: `package app

import "time"

var Name string
var UTCName string

func init() {
  loc, err := time.LoadLocation("America/New_York")
  if err != nil {
    Name = err.Error()
    return
  }
  Name = loc.String()
  loc, err = time.LoadLocation("UTC")
  if err != nil {
    UTCName = err.Error()
    return
  }
  UTCName = loc.String()
}
`
      }]
    }]);

    expect(result.diagnostics).toEqual([]);
    expect(result.packages.app?.Name).toBe("UTC");
    expect(result.packages.app?.UTCName).toBe("UTC");
  });

  test("package methods capture sibling package functions", async () => {
    const result = await evaluatePackageSourceFiles([{
      filename: "/workspace/p/p.go",
      source: `package p

type T struct{}

func helper() int {
  return 7
}

func (T) M() int {
  return helper()
}

var Out = T{}.M()
`
    }], { importPath: "example.com/p" });

    expect(result.diagnostics).toEqual([]);
    expect(result.package?.Out).toBe(7n);
  });

  test("imported closures keep their defining package method context", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/once",
      files: [{
        filename: "/workspace/once/once.go",
        source: `package once

type Once struct {
  done bool
}

func (o *Once) Do(f func()) {
  if !o.done {
    o.done = true
    f()
  }
}

func Make() func() int {
  d := struct {
    once Once
    result int
  }{}
  return func() int {
    d.once.Do(func() {
      d.result = 9
    })
    return d.result
  }
}
`
      }]
    }]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import once "example.com/once"
f := once.Make()
return f(), f()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.values).toEqual([9n, 9n]);
  });

  test("uses host mutex primitives for internal sync fields", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/sync",
        files: [{
          filename: "/usr/local/go/src/internal/sync/mutex.go",
          source: `package sync

type Mutex struct{}

func (m *Mutex) Lock() {}
func (m *Mutex) Unlock() {}
`
        }]
      },
      {
        importPath: "sync",
        files: [{
          filename: "/usr/local/go/src/sync/once.go",
          source: `package sync

import isync "internal/sync"

type Mutex struct {
  mu isync.Mutex
}

func (m *Mutex) Lock() {
  m.mu.Lock()
}

func (m *Mutex) Unlock() {
  m.mu.Unlock()
}

type Once struct {
  done bool
  m Mutex
}

func (o *Once) Do(f func()) {
  o.m.Lock()
  defer o.m.Unlock()
  if !o.done {
    defer func() { o.done = true }()
    f()
  }
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "sync"

var once sync.Once
hits := 0
once.Do(func() { hits++ })
once.Do(func() { hits++ })
return hits
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toBe(1n);
  });

  test("imported package methods externalize private helper return types", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/base",
        files: [{
          filename: "/workspace/base/base.go",
          source: `package base

type hidden struct { X int }

type T struct { h *hidden }

func New() *T { return &T{} }

func lookup() *hidden { return &hidden{X: 9} }

func (t *T) Init() { t.h = lookup() }

func (t *T) Value() int { return t.h.X }
`
        }]
      },
      {
        importPath: "example.com/wrap",
        files: [{
          filename: "/workspace/wrap/wrap.go",
          source: `package wrap

import "example.com/base"

var V = base.New()

func init() { V.Init() }

func Value() int { return V.Value() }
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/wrap"
return wrap.Value()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(9n);
  });

  test("infers imported generic function parameters from caller-private struct arguments", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/uniq",
        files: [{
          filename: "/workspace/uniq/uniq.go",
          source: `package uniq

type Handle[T comparable] struct {
  Value T
}

func Make[T comparable](value T) Handle[T] {
  return Handle[T]{Value: value}
}
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "example.com/uniq"

type detail struct {
  ok bool
}

var Got = uniq.Make(detail{ok: true}).Value.ok
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages["example.com/app"]?.Got).toBe(true);
  });

  test("preserves imported slice metadata for string conversions", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/tzdata",
        files: [{
          filename: "/workspace/tzdata/tzdata.go",
          source: `package tzdata

type reader struct {
  p []byte
}

func (r *reader) read(n int) []byte {
  p := r.p[0:n]
  r.p = r.p[n:]
  return p
}

func Magic() []byte {
  r := reader{p: []byte{'T', 'Z', 'i', 'f'}}
  return r.read(4)
}
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "example.com/tzdata"

var Got = string(tzdata.Magic())
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages["example.com/app"]?.Got).toBe("TZif");
  });

  test("typechecks and evaluates source packages that import other source packages", async () => {
    const lib = await evaluatePackageSourceFiles([{
      filename: "lib.go",
      source: `package lib

func One() int {
  return 1
}
`
    }], { importPath: "example.com/lib" });
    expect(lib.diagnostics).toEqual([]);
    expect(lib.package).toBeDefined();
    expect(lib.packageInfo).toBeDefined();

    const app = await evaluatePackageSourceFiles([{
      filename: "app.go",
      source: `package app

import lib "example.com/lib"

func Two() int {
  return lib.One() + 1
}
`
    }], {
      importPath: "example.com/app",
      packages: {
        "example.com/lib": lib.package ?? {}
      },
      packageInfos: {
        "example.com/lib": lib.packageInfo!
      }
    });
    expect(app.diagnostics).toEqual([]);
    expect(app.package).toBeDefined();

    const result = await expectRuns(`
import app "example.com/app"
return app.Two()
`, {
      packages: {
        "example.com/app": app.package ?? {},
        "example.com/lib": lib.package ?? {}
      }
    });
    expect(result.value).toBe(2n);
  });

  test("loads provider-supplied standard source packages instead of ambient fallbacks", async () => {
    const loaded: string[] = [];
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "math"

var V = math.Twice(21)
`
      }]
    }], {
      sourcePackageProvider: {
        load(importPath) {
          loaded.push(importPath);
          if (importPath !== "math") return undefined;
          return [{
            filename: "/workspace/math/math.go",
            source: `package math

func Twice(x int) int { return x * 2 }
`
          }];
        }
      }
    });

    expect(graph.diagnostics).toEqual([]);
    expect(loaded).toEqual(["math"]);
    expect(graph.initializedImportPaths).toEqual(["math", "example.com/app"]);
    expect(graph.packages["example.com/app"]?.V).toBe(42n);
  });

  test("keeps package import paths distinct when package clauses share a name", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "example.com/strconv"

func Run() string {
  return strconv.Unquote()
}
`
      }]
    }], {
      sourcePackageProvider: {
        load(importPath) {
          if (importPath === "example.com/internal/strconv") {
            return [{
              filename: "/workspace/example.com/internal/strconv/strconv.go",
              source: `package strconv

type Error string

func Atoi() string { return "internal" }
`
            }];
          }
          if (importPath === "example.com/strconv") {
            return [{
              filename: "/workspace/example.com/strconv/strconv.go",
              source: `package strconv

import "example.com/internal/strconv"

type NumError struct {
  Err strconv.Error
}

func Unquote() string { return "public" }
`
            }];
          }
          return undefined;
        }
      }
    });
    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/strconv"
return strconv.Unquote()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toBe("public");
  });

  test("keeps imports file scoped across same-package source files", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/example.com/app/public.go",
        source: `package app

import strconv "example.com/public/strconv"

func Pub() string { return strconv.Unquote() }
`
      }, {
        filename: "/workspace/example.com/app/internal.go",
        source: `package app

import strconv "example.com/internal/strconv"

func Int() string { return strconv.Atoi() }
`
      }, {
        filename: "/workspace/example.com/app/run.go",
        source: `package app

func Run() string { return Pub() + ":" + Int() }
`
      }]
    }], {
      sourcePackageProvider: {
        load(importPath) {
          if (importPath === "example.com/public/strconv") {
            return [{
              filename: "/workspace/example.com/public/strconv/strconv.go",
              source: `package strconv

func Unquote() string { return "public" }
`
            }];
          }
          if (importPath === "example.com/internal/strconv") {
            return [{
              filename: "/workspace/example.com/internal/strconv/strconv.go",
              source: `package strconv

func Atoi() string { return "internal" }
`
            }];
          }
          return undefined;
        }
      }
    });
    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/app"
return app.Run()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toBe("public:internal");
  });

  test("resolves imported constants in array literal lengths", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "example.com/lib"

func Run() int {
  xs := [lib.K]int{0: 7, 2: 9}
  return len(xs)
}
`
      }]
    }], {
      sourcePackageProvider: {
        load(importPath) {
          if (importPath === "example.com/lib") {
            return [{
              filename: "/workspace/example.com/lib/lib.go",
              source: `package lib

const K = 3
`
            }];
          }
          return undefined;
        }
      }
    });
    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/app"
return app.Run()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toBe(3n);
  });

  test("runs source-built fmt string formatting through package graph", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/fmtmain/main.go",
      source: `package main

import "fmt"

func main() {
  print(fmt.Sprint("zygo") + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["zygo\n"]);
  });

  test("runs source-built fmt printf through os stdout file writes", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/fmtprint/main.go",
      source: `package main

import "fmt"

func main() {
  fmt.Printf("hello %s!\\n", "gojr")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["hello gojr!\n"]);
  });

  test("runs source-built fmt sharp-v formatting over slices", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/fmtslice/main.go",
      source: `package main

import "fmt"

func main() {
  args := []string{"zygo", "-h"}
  fmt.Printf("args=%#v\\n", args)
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual([`args=[]string{"zygo", "-h"}\n`]);
  });

  test("runs source-built fmt formatting through error interfaces", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/fmterror/main.go",
      source: `package main

import "fmt"

type customError struct { msg string }
func (e *customError) Error() string { return "ERR:" + e.msg }

func main() {
  var err error = &customError{msg:"boom"}
  fmt.Printf("%v|%s\\n", err, err)
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ERR:boom|ERR:boom\n"]);
  });

  test("initializes source-built base64 byte-string constants", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/base64main/main.go",
      source: `package main

import "encoding/base64"

func main() {
  _ = base64.StdEncoding
  print("base64 ok\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["base64 ok\n"]);
  });

  test("runs source-built reflect.New for interface-held pointer values", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/refliface/main.go",
      source: `package main

import "reflect"

type Value interface{ String() string }
type stringValue string

func (s *stringValue) String() string { return string(*s) }

func main() {
  var sv stringValue
  var v Value = &sv
  typ := reflect.TypeOf(v)
  z := reflect.New(typ.Elem())
  print(typ.Kind().String() + "," + typ.Elem().Kind().String() + "," + z.Type().Kind().String() + "\\n")
  print(z.Interface().(Value).String() + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ptr,string,ptr\n", "\n"]);
  });

  test("runs source-built reflect Type Elem through runtime descriptors", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflelem/main.go",
      source: `package main

import "reflect"

func main() {
  print(reflect.TypeOf(map[string]int{}).Elem().Kind().String() + "\\n")
  print(reflect.TypeOf([2]string{}).Elem().Kind().String() + "\\n")
  print(reflect.TypeOf([]float64{}).Elem().Kind().String() + "\\n")
  print(reflect.TypeOf((*error)(nil)).Elem().Kind().String() + "\\n")
  print(reflect.TypeOf(make(chan byte)).Elem().Kind().String() + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["int\n", "string\n", "float64\n", "interface\n", "uint8\n"]);
  });

  test("source-built reflect sees caller struct fields through any parameters", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/inspector",
      files: [{
        filename: "/workspace/inspector/inspector.go",
        source: `package inspector

import "reflect"

func Inspect(v any) (int, string, string, string) {
  typ := reflect.ValueOf(v).Elem().Type()
  field := typ.Field(0)
  return typ.NumField(), field.Name, string(field.Tag), field.Tag.Get("json")
}
`
      }]
    }, {
      importPath: "example.com/app",
      files: [{
        filename: "/workspace/app/app.go",
        source: `package app

import "example.com/inspector"

type Table struct {
  Headers []string   ` + "`json:\"headers\" msg:\"headers\"`" + `
  Rows    [][]string ` + "`json:\"rows\" msg:\"rows\"`" + `
}

func Run() (int, string, string, string) {
  return inspector.Inspect(&Table{})
}
`
      }]
    }], {
      sourcePackageProvider
    });
    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/app"
return app.Run()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toEqual([
      2n,
      "Headers",
      `json:"headers" msg:"headers"`,
      "headers"
    ]);
  });

  test("source-built reflect field Addr Interface preserves field identity", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflfieldaddr/main.go",
      source: `package main

import "reflect"

type Table struct {
  Headers []string
}

func fill(v any) {
  ptr := reflect.ValueOf(v).Elem().Field(0).Addr().Interface().(*[]string)
  *ptr = []string{"wood", "metal"}
}

func main() {
  table := &Table{}
  fill(table)
  print(table.Headers[0] + "," + table.Headers[1] + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["wood,metal\n"]);
  });

  test("source-built reflect rtype methods dispatch after Value.Type", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflrtype/main.go",
      source: `package main

import "reflect"

type Table struct {
  Headers []string
}

func main() {
  fieldType := reflect.ValueOf(&Table{}).Elem().Field(0).Type()
  print(fieldType.Kind().String() + "," + fieldType.Elem().Kind().String() + "\\n")
  ptrElem := reflect.TypeOf(&[]string{}).Elem()
  print(ptrElem.Kind().String() + "," + ptrElem.Elem().Kind().String() + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["slice,string\n", "slice,string\n"]);
  });

  test("source-built reflect MakeSlice Append and Set mutate slice fields", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflmakeslice/main.go",
      source: `package main

import "reflect"

type Table struct {
  Headers []string
}

func fill(v any) {
  fld := reflect.ValueOf(v).Elem().Field(0)
  slc := reflect.MakeSlice(fld.Type(), 0, 2)
  elem := reflect.New(fld.Type().Elem())
  elem.Elem().SetString("wood")
  slc = reflect.Append(slc, elem.Elem())
  elem = reflect.New(fld.Type().Elem())
  elem.Elem().SetString("metal")
  slc = reflect.Append(slc, elem.Elem())
  fld.Set(slc)
}

func main() {
  table := &Table{}
  fill(table)
  print(table.Headers[0] + "," + table.Headers[1] + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["wood,metal\n"]);
  });

  test("source-built reflect sets embedded interface fields from concrete pointer values", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflinterfacefield/main.go",
      source: `package main

import "reflect"

type Flyer interface {
  Fly() string
}

type Plane struct {
  Chld Flyer
}

type Snoopy struct {
  Plane
}

type Hellcat struct {
  Speed int
}

func (h *Hellcat) Fly() string {
  return "hellcat"
}

func fill(target any) {
  targVa := reflect.ValueOf(target)
  fld := targVa.Elem().Field(0).Field(0)
  ptrFld := fld.Addr()
  factOutputVal := reflect.ValueOf(&Hellcat{Speed: 567})
  ifacePtr := reflect.ValueOf(ptrFld.Interface())
  if ifacePtr.Type().Elem().Kind() == reflect.Interface && factOutputVal.Type().Implements(ifacePtr.Type().Elem()) {
    ifacePtr.Elem().Set(factOutputVal)
  }
}

func main() {
  var sn Snoopy
  fill(&sn)
  print(sn.Chld.Fly() + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["hellcat\n"]);
  });

  test("calls nil-safe pointer receiver methods on typed nil arguments", async () => {
    const result = await expectRuns(`
type PrintState struct {
  Indent int
}

func (ps *PrintState) GetIndent() int {
  if ps == nil {
    return 0
  }
  return ps.Indent
}

func show(ps *PrintState) int {
  return ps.GetIndent()
}

return show(nil)
`);

    expect(result.values).toEqual([0n]);
  });

  test("uses source-built reflect Type identity as a map key", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/refltypemap/main.go",
      source: `package main

import "reflect"

type wireType struct {
  A *int
}

func main() {
  a := reflect.TypeFor[wireType]()
  b := reflect.TypeFor[wireType]()
  c := reflect.TypeOf((*wireType)(nil)).Elem()
  m := make(map[reflect.Type]int)
  m[a] = 11
  firstB := m[b]
  firstC := m[c]
  m[b] = 12
  m[c] = 13
  print(len(m), ":", firstB, ":", firstC, ":", m[a], ":", m[b], ":", m[c], "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toEqual("1:11:11:13:13:13\n");
  });

  test("runs source-built flag default printing through reflect.New", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/flagmain/main.go",
      source: `package main

import "flag"

func main() {
  fs := flag.NewFlagSet("x", flag.ContinueOnError)
  fs.String("memprofile", "", "write mem profile to file")
  fs.PrintDefaults()
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
  });

  test("runs source-built flag ExitOnError help through os.Exit", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/flaghelp/main.go",
      source: `package main

import "flag"

func main() {
  fs := flag.NewFlagSet("zygo", flag.ExitOnError)
  fs.String("memprofile", "", "write mem profile to file")
  fs.Parse([]string{"-h"})
  print("not reached\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.exitCode).toBe(0);
    expect(result.output.join("")).toContain("Usage of zygo:");
    expect(result.output.join("")).not.toContain("not reached");
  });

  test("runs source-built flag StringVar into struct fields", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/flagstringvar/main.go",
      source: `package main

import "flag"

type Config struct {
  Command string
}

func main() {
  var cfg Config
  fs := flag.NewFlagSet("zygo", flag.ContinueOnError)
  fs.StringVar(&cfg.Command, "c", "", "expressions to evaluate")
  err := fs.Parse([]string{"-c", "(+ 1 2)"})
  if err != nil {
    panic(err)
  }
  print(cfg.Command + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toBe("(+ 1 2)\n");
  });

  test("writes through pointers to struct fields", async () => {
    const result = await expectRuns(`
type Config struct {
  Command string
}
func set(p *string) { *p = "ok" }
var cfg Config
set(&cfg.Command)
cfg.Command
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ok");
  });

  test("assigns converted pointers with named scalar receiver methods to interfaces", async () => {
    const result = await expectRuns(`
type Value interface {
  Set(string) error
  String() string
}
type stringValue string
func (s *stringValue) Set(v string) error { *s = stringValue(v); return nil }
func (s *stringValue) String() string { return string(*s) }
var target string
var v Value = (*stringValue)(&target)
panicOn(v.Set("ok"))
target
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ok");
  });

  test("formats source-level error interfaces with Error methods", async () => {
    const result = await expectRuns(`
import "fmt"

type customError struct { msg string }
func (e *customError) Error() string { return "ERR:" + e.msg }
var err error = &customError{msg:"boom"}
fmt.Printf("%v|%s\\n", err, err)
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ERR:boom|ERR:boom\n"]);
  });

  test("asserts predeclared error from an any interface value", async () => {
    const result = await expectRuns(`
type customError struct { msg string }
func (e *customError) Error() string { return "ERR:" + e.msg }
var err error = &customError{msg:"boom"}
var a any = err
got, ok := a.(error)
if !ok { panic("missing error assertion") }
got.Error()
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ERR:boom");
  });

  test("mutates maps reached through interface slice type assertions", async () => {
    const result = await expectRuns(`
type Elem interface { IsElem() }
type Value interface { Value() string }
type Item struct { name string }
func (i *Item) IsElem() {}
func (i *Item) Value() string { return i.name }
type Scope struct { Map map[int]Value }
func (s *Scope) IsElem() {}
type Stack struct { elements []Elem }
func NewStack() *Stack { return &Stack{elements: make([]Elem, 0)} }
func (s *Stack) Push(e Elem) { s.elements = append(s.elements, e) }
type Env struct { stack *Stack }
func NewEnv() *Env {
  env := &Env{stack: NewStack()}
  env.stack.Push(&Scope{Map: make(map[int]Value)})
  return env
}
func (env *Env) Add(k int, v Value) {
  env.stack.elements[0].(*Scope).Map[k] = v
}
env := NewEnv()
env.Add(3, &Item{name:"ok"})
env.stack.elements[0].(*Scope).Map[3].Value()
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ok");
  });

  test("mutates package-local global scope maps through constructor methods", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/envbugmain/main.go",
      source: `package main

import "example.com/envbug"

func main() {
  print(envbug.Run() + "\\n")
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/envbug",
        files: [{
          filename: "/workspace/envbug/envbug.go",
          source: `package envbug

type Elem interface { IsElem() }
type Value interface { Value() string }
type Item struct { name string }
func (i *Item) IsElem() {}
func (i *Item) Value() string { return i.name }
type Scope struct { Map map[int]Value }
func (s *Scope) IsElem() {}
type Stack struct { tos int; elements []Elem }
func NewStack() *Stack { return &Stack{tos: -1, elements: make([]Elem, 0)} }
func (s *Stack) Push(e Elem) { s.tos++; s.elements = append(s.elements, e) }
type Env struct { stack *Stack; symbols map[string]int; next int }
func NewEnv() *Env {
  env := new(Env)
  env.stack = NewStack()
  env.stack.Push(&Scope{Map: make(map[int]Value)})
  env.symbols = make(map[string]int)
  env.next = 1
  env.Add("ok", &Item{name:"ok"})
  return env
}
func (env *Env) Symbol(name string) int {
  n, ok := env.symbols[name]
  if ok { return n }
  n = env.next
  env.next++
  env.symbols[name] = n
  return n
}
func (env *Env) Add(name string, value Value) {
  sym := env.Symbol(name)
  env.stack.elements[0].(*Scope).Map[sym] = value
}
func (env *Env) Find(name string) (Value, bool) {
  sym := env.Symbol(name)
  v, ok := env.stack.elements[0].(*Scope).Map[sym]
  return v, ok
}
func Run() string {
  env := NewEnv()
  v, ok := env.Find("ok")
  if !ok { return "missing" }
  return v.Value()
}
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ok\n"]);
  });

  test("treats nil error results as nil after multi-return lookup", async () => {
    const result = await expectRuns(`
type Value interface { Value() string }
type Item struct { name string }
func (i *Item) Value() string { return i.name }
type Scope struct{}
func lookup() (Value, error, *Scope) {
  return &Item{name:"ok"}, nil, &Scope{}
}
v, err, _ := lookup()
if err != nil { return "bad-if" }
switch err {
case nil:
  return v.Value()
default:
  return "bad-switch"
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ok");
  });

  test("finds map entries after interface-returning stack lookup type switches", async () => {
    const result = await expectRuns(`
type StackElem interface { IsStackElem() }
type Sexp interface { SexpString() string }
type Item struct { name string }
func (i *Item) SexpString() string { return i.name }
type Scope struct { Map map[int]Sexp }
func (s Scope) IsStackElem() {}
type Stack struct { tos int; elements []StackElem }
func NewStack() *Stack { return &Stack{tos: -1, elements: make([]StackElem, 0)} }
func (s *Stack) Push(e StackElem) { s.tos++; s.elements = append(s.elements, e) }
func (s *Stack) Get(i int) (StackElem, error) { return s.elements[i], nil }
func (s *Stack) Lookup(k int) (Sexp, error, *Scope) {
  for i := 0; i <= s.tos; i++ {
    elem, err := s.Get(i)
    if err != nil { return nil, err, nil }
    switch scope := elem.(type) {
    case (*Scope):
      value, ok := scope.Map[k]
      if ok { return value, nil, scope }
    }
  }
  return nil, fmt.Errorf("not found"), nil
}
stack := NewStack()
stack.Push(&Scope{Map: make(map[int]Sexp)})
stack.elements[0].(*Scope).Map[17] = &Item{name:"ok"}
value, err, _ := stack.Lookup(17)
if err != nil { return "missing" }
return value.SexpString()
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("ok");
  });

  test("keeps imported receiver map identity across exported method calls", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/envboundarymain/main.go",
      source: `package main

import "example.com/envboundary"

func main() {
  env := envboundary.NewEnv()
  env.Add("ok", envboundary.NewItem("ok"))
  value, found := env.Find("ok")
  if !found {
    print("missing\\n")
    return
  }
  print(value.SexpString() + "\\n")
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/envboundary",
        files: [{
          filename: "/workspace/envboundary/envboundary.go",
          source: `package envboundary

import "fmt"

type StackElem interface { IsStackElem() }
type Sexp interface { SexpString() string }
type Item struct { name string }
func NewItem(name string) *Item { return &Item{name:name} }
func (i *Item) SexpString() string { return i.name }
type Scope struct { Map map[int]Sexp }
func (s Scope) IsStackElem() {}
type Stack struct { tos int; elements []StackElem }
func NewStack() *Stack { return &Stack{tos: -1, elements: make([]StackElem, 0)} }
func (s *Stack) Push(e StackElem) { s.tos++; s.elements = append(s.elements, e) }
func (s *Stack) Get(i int) (StackElem, error) { return s.elements[i], nil }
func (s *Stack) Lookup(k int) (Sexp, error, *Scope) {
  for i := 0; i <= s.tos; i++ {
    elem, err := s.Get(i)
    if err != nil { return nil, err, nil }
    switch scope := elem.(type) {
    case (*Scope):
      value, ok := scope.Map[k]
      if ok { return value, nil, scope }
    }
  }
  return nil, fmt.Errorf("not found"), nil
}
type Env struct { stack *Stack; symbols map[string]int; next int }
func NewEnv() *Env {
  env := &Env{stack: NewStack(), symbols: make(map[string]int), next: 1}
  env.stack.Push(&Scope{Map: make(map[int]Sexp)})
  return env
}
func (env *Env) Symbol(name string) int {
  n, ok := env.symbols[name]
  if ok { return n }
  n = env.next
  env.next++
  env.symbols[name] = n
  return n
}
func (env *Env) Add(name string, value Sexp) {
  env.stack.elements[0].(*Scope).Map[env.Symbol(name)] = value
}
func (env *Env) Find(name string) (Sexp, bool) {
  value, err, _ := env.stack.Lookup(env.Symbol(name))
  if err != nil { return nil, false }
  return value, true
}
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ok\n"]);
  });

  test("passes imported package-owned pointers back into their methods without recursive wrappers", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/wrapmain/main.go",
      source: `package main

import "example.com/wrap"

func main() {
  env := wrap.NewEnv()
  env.Add("ok")
  print(env.Last() + "\\n")
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/wrap",
        files: [{
          filename: "/workspace/wrap/wrap.go",
          source: `package wrap

type Env struct { values []string }

func NewEnv() *Env { return &Env{} }
func (env *Env) Add(value string) { env.values = append(env.values, value) }
func (env *Env) Last() string { return env.values[len(env.values)-1] }
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ok\n"]);
  });

  test("passes same-package pointer values to same-package interface parameters inside imported methods", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/listifacemain/main.go",
      source: `package main

import "example.com/listiface"

func main() {
  env := listiface.NewEnv()
  got := env.FilterList(&listiface.Pair{})
  print(got + "\\n")
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/listiface",
        files: [{
          filename: "/workspace/listiface/listiface.go",
          source: `package listiface

type PrintState struct{}
type RegisteredType struct{}

type Sexp interface {
  SexpString(ps *PrintState) string
  Type() *RegisteredType
}

type Pair struct {
  Head Sexp
  Tail Sexp
}

func (p *Pair) SexpString(ps *PrintState) string { return "pair" }
func (p *Pair) Type() *RegisteredType { return nil }

func ListToArray(expr Sexp) string {
  return expr.SexpString(nil)
}

type Env struct{}
func NewEnv() *Env { return &Env{} }
func (env *Env) FilterList(h *Pair) string { return ListToArray(h) }
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["pair\n"]);
  });

  test("type switches imported interface elements pushed through exported stacks", async () => {
    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/stackboundarymain/main.go",
      source: `package main

import "example.com/stackboundary"

func main() {
  env := stackboundary.NewEnv()
  stack := env.NewStack()
  scope := env.NewNamedScope("probe")
  sym := env.MakeSymbol("direct")
  scope.Map[sym.Number()] = stackboundary.SexpNull
  _, directOK := scope.Map[sym.Number()]
  stack.Push(scope)
  _, err, _ := stack.LookupSymbol(sym)
  if !directOK {
    print("direct-missing\\n")
    return
  }
  if err != nil {
    print("lookup-missing\\n")
    return
  }
  print("ok\\n")
}
`
    }], {
      sourcePackages: [{
        importPath: "example.com/stackboundary",
        files: [{
          filename: "/workspace/stackboundary/stackboundary.go",
          source: `package stackboundary

import "fmt"

type StackElem interface { IsStackElem() }
type Sexp interface { SexpString() string }
type SexpSentinel struct{}
func (s *SexpSentinel) SexpString() string { return "nil" }
var SexpNull Sexp = &SexpSentinel{}
type Symbol struct { name string; number int }
func (s *Symbol) Number() int { return s.number }
type Scope struct { Map map[int]Sexp; Name string }
func (s Scope) IsStackElem() {}
type Stack struct { tos int; elements []StackElem }
func (s *Stack) Push(e StackElem) { s.tos++; s.elements = append(s.elements, e) }
func (s *Stack) Get(i int) (StackElem, error) { return s.elements[i], nil }
func (s *Stack) LookupSymbol(sym *Symbol) (Sexp, error, *Scope) {
  for i := 0; i <= s.tos; i++ {
    elem, err := s.Get(i)
    if err != nil { return SexpNull, err, nil }
    switch scope := elem.(type) {
    case (*Scope):
      value, ok := scope.Map[sym.number]
      if ok { return value, nil, scope }
    }
  }
  return SexpNull, fmt.Errorf("not found"), nil
}
type Env struct { symtable map[string]int; next int }
func NewEnv() *Env { return &Env{symtable: make(map[string]int), next: 1} }
func (env *Env) NewStack() *Stack { return &Stack{tos: -1, elements: make([]StackElem, 0)} }
func (env *Env) NewNamedScope(name string) *Scope { return &Scope{Map: make(map[int]Sexp), Name: name} }
func (env *Env) MakeSymbol(name string) *Symbol {
  n, ok := env.symtable[name]
  if !ok {
    n = env.next
    env.next++
    env.symtable[name] = n
  }
  return &Symbol{name: name, number: n}
}
`
        }]
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ok\n"]);
  });

  test("preserves imported source-built slice identity through bytes.Buffer", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/bytesmain/main.go",
      source: `package main

import "bytes"

func main() {
  var b bytes.Buffer
  b.WriteString("ab")
  print(b.String() + "\\n")
  b.WriteByte('!')
  print(b.String() + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["ab\n", "ab!\n"]);
  });

  test("shares standard iter type identity across source-built maps and slices", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/mapslices/main.go",
      source: `package main

import (
  "maps"
  "slices"
)

func main() {
  param := map[string]string{"b": "2", "a": "1"}
  keys := slices.Sorted(maps.Keys(param))
  print(keys[0] + "," + keys[1] + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["a,b\n"]);
  });

  test("runs standard unique.Make through source-built hashtriemap dependencies", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/uniquemain/main.go",
      source: `package main

import "unique"

type detail struct {
  isV6 bool
  zoneV6 string
}

func main() {
  h := unique.Make(detail{isV6: true, zoneV6: "z"})
  print(h.Value().zoneV6 + "\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["z\n"]);
  });

  test("runs Go-junior tests that import source package metadata", async () => {
    const app = await evaluatePackageSourceFiles([{
      filename: "app.go",
      source: `package app

func Two() int {
  return 2
}
`
    }], { importPath: "example.com/app" });
    expect(app.diagnostics).toEqual([]);
    expect(app.package).toBeDefined();
    expect(app.packageInfo).toBeDefined();

    const result = await testSourceFiles([{
      filename: "use_app_test.go",
      source: `package useapp

import (
  app "example.com/app"
  "testing"
)

func TestTwo(t *testing.T) {
  if app.Two() != 2 {
    t.Fatalf("bad Two")
  }
}
`
    }], {
      packages: {
        "example.com/app": app.package ?? {}
      },
      packageInfos: {
        "example.com/app": app.packageInfo!
      }
    });
    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("PASS");
  });

  test("runs package tests with package-owned callback closure type context", async () => {
    const result = await testSourceFilesWithPackagesOnNode({
      importPath: "example.com/p",
      packageName: "p",
      files: [{
        filename: "p_test.go",
        source: `package p

import (
  invoker "example.com/invoker"
  "testing"
)

type Local struct{}

func TestImportedCallback(t *testing.T) {
  invoker.Call(func() {
    _ = &Local{}
  })
}
`
      }],
      packages: [{
        importPath: "example.com/invoker",
        files: [{
          filename: "invoker.go",
          source: `package invoker

func Call(fn func()) {
  fn()
}
`
        }]
      }],
      testVerbose: true
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("PASS");
  });

  test("expands cached standard-library interface method tuple results in short declarations", async () => {
    const packageCacheParent = mkdtempSync(join(tmpdir(), "gojr-runtime-test-cache-"));
    try {
      const result = await testSourceFilesWithPackagesOnNode({
        importPath: "example.com/readrune",
        packageName: "readrune",
        files: [{
          filename: "readrune_test.go",
          source: `package readrune

import (
  "bytes"
  "io"
  "testing"
)

type Holder struct {
  stream io.RuneScanner
}

func TestReadRuneTuple(t *testing.T) {
  h := Holder{stream: bytes.NewBuffer([]byte("a"))}
  r, _, err := h.stream.ReadRune()
  if r != 97 || err != nil {
    t.Fatalf("ReadRune() = %v, %v", r, err)
  }
}
`
        }],
        packageCacheParent,
        testVerbose: true
      });

      expect(result.diagnostics).toEqual([]);
      expect(result.output.join("")).toContain("PASS");
    } finally {
      rmSync(packageCacheParent, { recursive: true, force: true });
    }
  });

  test("runs Go-junior tests with testing.T", async () => {
    const result = await testSource(`
import "testing"

func Add(a, b int) int { return a + b }

func TestAdd(t *testing.T) {
  t.Helper()
  t.Logf("sum=%v", Add(2, 3))
  if Add(2, 3) != 5 {
    t.Fatalf("bad sum")
  }
}

func TestNoArg() {}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("=== RUN   TestAdd\n");
    expect(result.output.join("")).toContain("    sum=5\n");
    expect(result.output.join("")).toContain("--- PASS: TestAdd\n");
    expect(result.output.join("")).toContain("--- PASS: TestNoArg\n");
    expect(result.output.join("")).toContain("PASS\n");
  });

  test("runs testing subtests and cleanups", async () => {
    const result = await testSource(`
import "testing"

var log string

func TestSubtests(t *testing.T) {
  t.Cleanup(func() { log += "cleanup;" })
  if !t.Run("child", func(t *testing.T) {
    t.Cleanup(func() { log += "child-cleanup;" })
    log += "child;"
  }) {
    t.Fatalf("subtest failed")
  }
  log += "parent;"
}

func TestAfter(t *testing.T) {
  if log != "child;child-cleanup;parent;cleanup;" {
    t.Fatalf("bad log: %s", log)
  }
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("PASS\n");
  });

  test("typechecks testing benchmarks", async () => {
    const result = await testSource(`
import "testing"

func BenchmarkThing(b *testing.B) {
  b.ReportAllocs()
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
  }
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("testing: warning: no tests to run\nPASS\n");
  });

  test("honors testing.Verbose from test options", async () => {
    const source = `
import "testing"

func TestVerbose(t *testing.T) {
  if !testing.Verbose() {
    t.Fatalf("not verbose")
  }
}
`;
    const quiet = await testSource(source, { testVerbose: false });
    expect(quiet.diagnostics).toHaveLength(1);
    expect(quiet.output.join("")).not.toContain("=== RUN   TestVerbose\n");
    expect(quiet.output.join("")).toContain("--- FAIL: TestVerbose\n");

    const verbose = await testSource(source, { testVerbose: true });
    expect(verbose.diagnostics).toEqual([]);
    expect(verbose.output.join("")).toContain("=== RUN   TestVerbose\n");
    expect(verbose.output.join("")).toContain("--- PASS: TestVerbose\n");
  });

  test("reports failing and skipped Go-junior tests", async () => {
    const result = await testSource(`
import "testing"

func TestFail(t *testing.T) {
  t.Errorf("bad %v", 3)
}

func TestSkip(t *testing.T) {
  t.Skipf("skip %v", 4)
}
`);

    const output = result.output.join("");
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_TEST001");
    expect(output).toContain("=== RUN   TestFail\n");
    expect(output).toContain("    bad 3\n");
    expect(output).toContain("--- FAIL: TestFail\n");
    expect(output).toContain("=== RUN   TestSkip\n");
    expect(output).toContain("    skip 4\n");
    expect(output).toContain("--- SKIP: TestSkip\n");
    expect(output).toContain("FAIL\n");
  });

  test("validates Go-junior test signatures", async () => {
    const result = await testSource(`
func TestBad(t int) {}
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_TEST001");
    expect(result.diagnostics[0]?.message).toContain("func TestX(t *testing.T)");
    expect(result.output.join("")).toContain("--- FAIL: TestBad\n");
  });

  test("supports made maps with typed string and integer keys", async () => {
    const stringKeyed = await expectRuns(`
m := make(map[string]int)
m["hi"] = 3
return m["hi"]
`);
    expect(stringKeyed.value).toBe(3n);

    const intKeyed = await expectRuns(`
mm := make(map[int]string)
mm[3] = "hi"
return mm[3]
`);
    expect(intKeyed.value).toBe("hi");
  });

  test("keeps declared map zero values nil and rejects writes before make", async () => {
    const read = await expectRuns(`
var m map[string]int
v, ok := m["hi"]
return m == nil, v, ok, len(m)
`);
    expect(read.values).toEqual([true, 0n, false, 0n]);

    const write = await evaluateSource(`
var m map[string]int
m["hi"] = 3
`);
    expect(write.diagnostics).toHaveLength(1);
    expect(write.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(write.diagnostics[0]?.message).toContain("assignment to entry in nil map");
  });

  test("supports Go two-value map lookups for key presence", async () => {
    const missing = await expectRuns(`
var m map[int]int
a, ok := m[3]
return a, ok
`);
    expect(missing.values).toEqual([0n, false]);

    const present = await expectRuns(`
m := make(map[int]string)
m[3] = "hi"
a, ok := m[3]
return a, ok
`);
    expect(present.values).toEqual(["hi", true]);
  });

  test("supports multi-result method calls in short declarations", async () => {
    const result = await expectRuns(`
type S struct{}
func (s *S) bounds() (int, int, bool, error) { return 1, 2, true, nil }
x := &S{}
start, end, ok, err := x.bounds()
return start, end, ok, err == nil
`);
    expect(result.values).toEqual([1n, 2n, true, true]);
  });

  test("supports imported multi-result functions and methods in short declarations", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/multis",
        files: [{
          filename: "/workspace/multis/multis.go",
          source: `package multis

type S struct{}

func New() *S { return &S{} }
func Pair() (int, int) { return 1, 2 }
func (s *S) Triple() (int, int, int) { return 3, 4, 5 }
`
        }]
      },
      {
        importPath: "example.com/app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "example.com/multis"

func Run() (int, int, int, int, int) {
  a, b := multis.Pair()
  s := multis.New()
  c, d, e := s.Triple()
  return a, b, c, d, e
}
`
        }]
      }
    ]);
    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/app"
a, b, c, d, e := app.Run()
return a, b, c, d, e
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });
    expect(result.values).toEqual([1n, 2n, 3n, 4n, 5n]);
  });

  test("supports named multi-result methods in short declarations", async () => {
    const result = await expectRuns(`
type S struct{}
func (s *S) Lookup() (a int, b int, c bool) {
  a = 6
  b = 7
  c = true
  return
}
s := &S{}
a, b, c := s.Lookup()
return a, b, c
`);
    expect(result.values).toEqual([6n, 7n, true]);
  });

  test("supports standard Go make for maps and slices", async () => {
    const mapResult = await expectRuns(`
m := make(map[int]int)
missing, missingOK := m[3]
m[3] = 9
present, presentOK := m[3]
return missing, missingOK, present, presentOK
`);
    expect(mapResult.values).toEqual([0n, false, 9n, true]);

    const sliceResult = await expectRuns(`
slc := make([]int, 20, 50)
slc = append(slc, 7)
slc2 := make([]string, 3)
return len(slc), cap(slc), slc[0], slc[20], len(slc2), cap(slc2), slc2[0]
`);
    expect(sliceResult.values).toEqual([21n, 50n, 0n, 7n, 3n, 3n, ""]);
  });

  test("zero-fills made byte slices through append and assignment", async () => {
    const result = await expectRuns(`
b := make([]byte, 3, 5)
c := append(b, byte(7))
var d []byte
d = c
return d[0], d[1], d[2], d[3], len(d), cap(d)
`);
    expect(result.values).toEqual([0n, 0n, 0n, 7n, 4n, 5n]);
  });

  test("externalizes sparse zero arrays with materialized zero elements", async () => {
    const result = await expectRuns(`
var b [2048]byte
return b
`);
    const values = result.value as unknown[];
    expect(Array.isArray(values)).toBe(true);
    expect(values.length).toBe(2048);
    expect(values[0]).toBe(0n);
    expect(values[1024]).toBe(0n);
    expect(values[2047]).toBe(0n);
  });

  test("supports standard Go make for named map slice and channel types", async () => {
    const result = await expectRuns(`
type Bytes []byte
type Counts map[string]int
type Ints chan int
b := make(Bytes, 2, 4)
b[1] = 7
m := make(Counts)
m["x"] = 9
ch := make(Ints, 1)
ch <- 11
return len(b), cap(b), b[0], b[1], m["x"], <-ch
`);
    expect(result.values).toEqual([2n, 4n, 0n, 7n, 9n, 11n]);
  });

  test("respects ordinary shadowing of predeclared make", async () => {
    const result = await expectRuns(`
make := func(x int) int { return x + 1 }
return make(10)
`);
    expect(result.value).toBe(11n);
  });

  test("reports ordinary runtime failures with GoJr-prefixed diagnostic codes", async () => {
    const result = await evaluateSource(`
missingName
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(result.diagnostics[0]?.message).toContain("missingName is not declared");
  });

  test("keeps filenames on source-file diagnostics", async () => {
    const result = await evaluateSourceFiles([
      {
        filename: "pkg/bad.go",
        source: `
package pkg

func Bad() {
  return )
}
`
      }
    ]);

    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.filename).toBe("pkg/bad.go");
    expect(result.diagnostics[0]?.span?.filename).toBe("pkg/bad.go");
    expect(result.diagnostics[0]?.span?.line).toBe(5);
  });

  test("keeps selector runtime diagnostics on the selector span", async () => {
    const result = await evaluateSource(`
type S struct{}
s := S{}
_ = s.missing
`, { filename: "pkg/select.go" });

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.filename).toBe("pkg/select.go");
    expect(result.diagnostics[0]?.span?.filename).toBe("pkg/select.go");
    expect(result.diagnostics[0]?.span?.line).toBe(4);
    expect(result.diagnostics[0]?.message).toContain("has no selector missing");
  });

  test("predeclares top-level types before variable initializers", async () => {
    const result = await expectRuns(`
var p = S1{
  F1: complex(float64(2.5), float64(-0.25)),
  F2: S2{F1: 9},
  F3: 103050709,
}

var d Derived = Box{X: 5}

type S1 struct {
  F1 complex128
  F2 S2
  F3 uint64
}

type S2 struct {
  F1 uint64
  F2 empty
}

type empty struct{}

type Derived interface { Base }
type Base interface { String() string }

type Box struct { X int }
func (b Box) String() string { return fmt.Sprintf("Box(%v)", b.X) }

return p.F2.F1, p.F3, d.String()
`);

    expect(result.values).toEqual([9n, 103050709n, "Box(5)"]);
  });

  test("enforces declared and inferred local variable types", async () => {
    const declared = await evaluateSource(`
var x int
x = "bad"
`);
    expect(declared.diagnostics).toHaveLength(1);
    expect(declared.diagnostics[0]?.message).toContain("variable x bad is not assignable to int");

    const inferred = await evaluateSource(`
x := 1
x = "bad"
`);
    expect(inferred.diagnostics).toHaveLength(1);
    expect(inferred.diagnostics[0]?.message).toContain("variable x bad is not assignable to int64");
  });

  test("rejects mixed string and numeric addition", async () => {
    const direct = await evaluateSource(`
d := \` hi there\`
a := 10
return d + a
`);
    expect(direct.diagnostics).toHaveLength(1);
    expect(direct.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(direct.diagnostics[0]?.message).toContain("invalid operation: string + int64");

    const session = new GoJuniorSession();
    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("d := ` hi there`")).diagnostics).toEqual([]);
    const mixed = await session.evaluate("d + a");
    expect(mixed.diagnostics).toHaveLength(1);
    expect(mixed.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(mixed.diagnostics[0]?.message).toContain("mismatched types string and int");

    const strings = await expectRuns(`
s := "hi"
return s + " there"
`);
    expect(strings.value).toBe("hi there");
  });

  test("compares huge untyped integer constants without losing precision", async () => {
    const result = await expectRuns(`
const (
  chuge = 1 << 100
  chuge_1 = chuge - 1
  c1 = chuge >> 100
)
return chuge > chuge_1, chuge == chuge_1, chuge_1 + 1 == chuge, c1 == 1, c1
`);

    expect(result.values).toEqual([true, false, true, true, 1n]);
  });

  test("supports Go-style const groups with iota and repeated expressions", async () => {
    const result = await expectRuns(`
const Single = iota
const (
  A = iota
  B
  C int = iota
  D
)
const (
  F float32 = 2 * iota
  G complex128 = iota
)
const (
  abit, amask = 1 << iota, 1<<iota - 1
  bbit, bmask
)
const (
  PackageX = 2
)
func shadowConst() (int, int, int, int, int, int) {
  const (
    First = iota
    iota = iota
    ShadowedA
    ShadowedB
  )
  const (
    PackageX = PackageX + PackageX
    LocalY
    LocalZ = iota
  )
  return First, ShadowedA, ShadowedB, PackageX, LocalY, LocalZ
}
First, ShadowedA, ShadowedB, LocalX, LocalY, LocalZ := shadowConst()
return Single, A, B, C, D, F, G, abit, amask, bbit, bmask, First, ShadowedA, ShadowedB, LocalX, LocalY, LocalZ
`);

    expect(result.values).toEqual([0n, 0n, 1n, 2n, 3n, 0, { real: 1, imag: 0 }, 1n, 0n, 2n, 1n, 0n, 1n, 1n, 4n, 8n, 1n]);

    const immutable = await evaluateSource(`
const X = 1
X = 2
`);
    expect(immutable.diagnostics).toHaveLength(1);
    expect(immutable.diagnostics[0]?.message).toContain("X is const");
  });

  test("enforces function parameter and typed collection assignments", async () => {
    const badParam = await evaluateSource(`
func f(x int) int { return x }
return f("bad")
`);
    expect(badParam.diagnostics).toHaveLength(1);
    expect(badParam.diagnostics[0]?.message).toContain("bad is not assignable to int");

    const badArray = await evaluateSource(`
var xs []int
xs = []string{"bad"}
`);
    expect(badArray.diagnostics).toHaveLength(1);
    expect(badArray.diagnostics[0]?.message).toContain("variable xs element bad is not assignable to int");

    const badMap = await evaluateSource(`
var m map[string]int
m = map[int]string{1: "bad"}
`);
    expect(badMap.diagnostics).toHaveLength(1);
    expect(badMap.diagnostics[0]?.message).toContain("variable m");
  });

  test("reports map key and value type mismatches without numeric-index coercion", async () => {
    const badKey = await evaluateSource(`
mm := make(map[int]string)
mm["hi"] = 3
`);
    expect(badKey.diagnostics).toHaveLength(1);
    expect(badKey.diagnostics[0]?.message).toContain("map key hi is not assignable to int");
    expect(badKey.diagnostics[0]?.message).not.toContain("numeric");

    const badValue = await evaluateSource(`
m := make(map[string]int)
m["hi"] = "three"
`);
    expect(badValue.diagnostics).toHaveLength(1);
    expect(badValue.diagnostics[0]?.message).toContain("map value three is not assignable to int");
  });

  test("evaluates map literals, missing-key zero values, and insertion-order range", async () => {
    const result = await expectRuns(`
counts := map[string]int64{"a": 1, "b": 2}
out := ""
for k, v := range counts {
  out = out + k + ":" + fmt.Sprintf("%v", v) + ";"
}
return counts["a"] + counts["b"] + counts["missing"], out
`);

    expect(result.values).toEqual([3n, "a:1;b:2;"]);
  });

  test("returns zero structs for missing struct-valued map keys", async () => {
    const result = await expectRuns(`
type charGroup struct {
  sign int
  class []rune
}
groups := map[string]charGroup{}
g := groups["missing"]
return g.sign, g.class == nil
`);

    expect(result.values).toEqual([0n, true]);
  });

  test("reports zero length for missing map and channel values", async () => {
    const result = await expectRuns(`
maps := map[string]map[int]int{}
chans := map[string]chan int{}
return len(maps["missing"]), len(chans["missing"]), cap(chans["missing"])
`);

    expect(result.values).toEqual([0n, 0n, 0n]);
  });

  test("evaluates named map composite literals", async () => {
    const result = await expectRuns(`
type M map[int]int
m := M{0: 10, 1: 20}
m[2] = 30
v, ok := m[1]
missing, missingOK := m[3]
return m[0] + v + m[2] + missing, ok, missingOK, len(m)
`);

    expect(result.values).toEqual([60n, true, false, 3n]);
  });

  test("formats typed maps with fmt verbs", async () => {
    const result = await expectRuns(`
import "fmt"

counts := map[string]int64{"a": 1, "b": 2}
floats := make(map[int]float64)
empty := fmt.Sprintf("%v", floats)
floats[39] = 3.2
return fmt.Sprintf("%v | %#v | %v | %v", counts, counts, empty, floats)
`);

    expect(result.value).toBe(`map[string]int64{a:1 b:2} | map[string]int64{string("a"): int64(1), string("b"): int64(2)} | map[int]float64{} | map[int]float64{39:3.2}`);
  });

  test("formats REPL map string values as Go literals", async () => {
    const result = await expectRuns(`
m := make(map[int]string)
m[3] = "hi"
m[5] = "there"
var empty map[int]string
multi := make(map[int]string)
multi[9] = "hello\\nthere"
quoted := make(map[int]string)
quoted[1] = "he said \\"hi\\""
quoted[2] = "tick \` and \\"quote\\""
return empty, m, multi, quoted
`);

    expect(result.values?.map(formatReplValue)).toEqual([
      `map[int]string(nil)`,
      `map[int]string{3:"hi", 5:"there"}`,
      "map[int]string{9:`hello\nthere`}",
      "map[int]string{1:`he said \"hi\"`, 2:`tick \\` and \"quote\"`}"
    ]);
  });

  test("evaluates struct literals, zero values, field mutation, and fmt verbs", async () => {
    const result = await expectRuns(`
type Point struct {
  X, Y int
  Name string
}

a := Point{X: 1, Y: 2, Name: "home"}
b := Point{3, 4, "away"}
c := Point{}
a.X = a.X + b.Y
return a.X, c.Y, fmt.Sprintf("%v | %#v", a, a)
`);

    expect(result.values).toEqual([
      5n,
      0n,
      `Point{X:5 Y:2 Name:home} | Point{X: int64(5), Y: int64(2), Name: string("home")}`
    ]);
  });

  test("evaluates elided composite literals in typed array and map literals", async () => {
    const result = await expectRuns(`
type Point struct{ X, Y int }
rows := []struct {
  Name string
  Pos Point
}{
  {"a", Point{1, 2}},
  {"b", {Y: 4}},
}
lookup := map[Point]Point{
  {X: 1}: {Y: 2},
}
type Node struct{ X int }
ptrs := map[int]*Node{
  0: {X: 7},
  1: {},
}
return rows[0].Name, rows[1].Pos.Y, lookup[Point{X: 1}].Y, ptrs[0].X, ptrs[1].X
`);

    expect(result.values).toEqual(["a", 4n, 2n, 7n, 0n]);
  });

  test("evaluates elided composite literals in named slice aliases", async () => {
    const result = await expectRuns(`
type Entry struct {
  X int
}
type Table []Entry

items := Table{
  {X: 1},
  {X: 2},
}
return items[0].X, items[1].X
`);

    expect(result.values).toEqual([1n, 2n]);
  });

  test("supports value and pointer receiver methods with Go selector syntax", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`
type Point struct {
  X, Y int
}
`)).diagnostics).toEqual([]);

    const valueMethod = await session.evaluate(`
func (p Point) Sum() int {
  return p.X + p.Y
}
`);
    expect(valueMethod.diagnostics).toEqual([]);

    const pointerMethod = await session.evaluate(`
func (p *Point) Scale(k int) {
  p.X = p.X * k
  p.Y = p.Y * k
}
`);
    expect(pointerMethod.diagnostics).toEqual([]);

    const call = await session.evaluate(`
p := Point{2, 3}
before := p.Sum()
p.Scale(4)
return before, p.X, p.Y, p.Sum()
`);
    expect(call.diagnostics).toEqual([]);
    expect(call.values).toEqual([5n, 8n, 12n, 20n]);
  });

  test("calls pointer receiver methods through addressable struct selectors", async () => {
    const result = await expectRuns(`
type Counter struct {
  N int
}

func (c *Counter) Inc() {
  c.N++
}

holder := struct {
  C Counter
}{}
holder.C.Inc()
holder.C.Inc()
return holder.C.N
`);
    expect(result.value).toBe(2n);
  });

  test("supports Go method expressions on named receiver types", async () => {
    const result = await expectRuns(`
type T []int

func (t T) Len() int {
  return len(t)
}

type Counter int

func (c *Counter) IncBy(k int) {
  *c = *c + Counter(k)
}

var t T = T{0, 1, 2, 3, 4}
var c Counter
f := T.Len
g := (*T).Len
h := (*Counter).IncBy
h(&c, 3)
return T.Len(t), f(t), g(&t), c
`);

    expect(result.values).toEqual([5n, 5n, 5n, 3n]);
  });

  test("supports interface and promoted method expressions", async () => {
    const result = await expectRuns(`
got := ""

type I interface {
  m()
}

type S struct{}

func (S) m() {
  got += "m;"
}

func (S) m1(s string) {
  got += "m1(" + s + ");"
}

type T int

func (T) m2() {
  got += "m2;"
}

type Outer struct { *Inner }
type Inner struct { s string }

func (i Inner) M() string {
  return i.s
}

I.m(S{})
f := interface{ m1(string) }.m1
f(S{}, "a")
interface{ m1(string) }.m1(S{}, "b")
g := struct{ T }.m2
g(struct{ T }{})
h := (*Outer).M
return got, h(&Outer{&Inner{"hello"}})
`);

    expect(result.values).toEqual(["m;m1(a);m1(b);m2;", "hello"]);
  });

  test("supports promoted pointer receiver method expressions", async () => {
    const result = await expectRuns(`
type Scalar int

func (s *Scalar) M(a int, x [2]int, b float64, y [2]float64) (Scalar, int, [2]int, float64, [2]float64) {
  return *s, a, x, b, y
}

type Wrapper struct {
  Scalar
}

var scalar Scalar = 42
var wrapper = &Wrapper{Scalar: scalar}
fn := (*Wrapper).M
s1, a1, x1, b1, y1 := fn(wrapper, 123, [2]int{456, 789}, 1.2, [2]float64{3.4, 5.6})
return s1 == scalar && a1 == 123 && x1 == [2]int{456, 789} && b1 == 1.2 && y1 == [2]float64{3.4, 5.6}
`);

    expect(result.value).toBe(true);
  });

  test("REPL sessions typecheck interface method expressions from earlier declarations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`got := ""`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`type I interface { m() }`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`type S struct{}`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`func (S) m() { got += "m" }`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`I.m(S{})`)).diagnostics).toEqual([]);

    const result = await session.evaluate("got");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("m");
  });

  test("keeps value selectors ahead of type method expressions when a type name is shadowed", async () => {
    const result = await expectRuns(`
type T struct { X int }

func (t T) XPlus(k int) int {
  return t.X + k
}

T := T{X: 7}
return T.X
`);

    expect(result.value).toBe(7n);
  });

  test("supports address-of and dereference assignment for structs and fields", async () => {
    const result = await expectRuns(`
type Box struct {
  X int
}

b := Box{X: 1}
p := &b
(*p).X = 9
q := &b.X
*q = *q + 1
return b.X, (*p).X
`);

    expect(result.values).toEqual([10n, 10n]);
  });

  test("uses declared field and element types for address expressions", async () => {
    const result = await expectRuns(`
func compareAndSwapInt32(addr *int32, old, next int32) bool {
  if *addr == old {
    *addr = next
    return true
  }
  return false
}

type Int32 struct {
  v int32
}

func (x *Int32) CompareAndSwap(old, next int32) bool {
  return compareAndSwapInt32(&x.v, old, next)
}

var x Int32
ok := x.CompareAndSwap(0, 1)
items := []int32{3}
p := &items[0]
*p = *p + 4
return ok, x.v, items[0]
`);

    expect(result.values).toEqual([true, 1n, 7n]);
  });

  test("executes Go type switches over structs, pointers, nil, and basic values", async () => {
    const result = await expectRuns(`
type Point struct {
  X int
}

describe := func(x interface{}) string {
  switch v := x.(type) {
  case nil:
    return "nil"
  case *Point:
    return fmt.Sprintf("ptr:%v", v.X)
  case Point:
    return fmt.Sprintf("point:%v", v.X)
  case int:
    return fmt.Sprintf("int:%v", v)
  default:
    return "other"
  }
}

p := Point{X: 7}
return describe(p), describe(&p), describe(nil), describe(3), describe("x")
`);

    expect(result.values).toEqual(["point:7", "ptr:7", "nil", "int:3", "other"]);
  });

  test("supports typed nil pointer conversions", async () => {
    const result = await expectRuns(`
var p *int
var five = 5
accept := func(p *int) bool { return p != nil }
return (*int)(nil) == p, fmt.Sprintf("%#v", (*int)(nil)), accept(&five)
`);

    expect(result.values).toEqual([true, "*int(nil)", true]);
  });

  test("supports pointer conversions between named types after typechecking", async () => {
    const result = await expectRuns(`
type A struct {
  name string
}
type B A

func (b *B) ok() bool {
  return b != nil
}

a := &A{name: "x"}
b := (*B)(a)
return b != nil, b.ok()
`);

    expect(result.values).toEqual([true, true]);
  });

  test("converts nil unsafe pointers back to typed pointers", async () => {
    const result = await expectRuns(`
import "unsafe"

type A struct{}
type B struct{}

p := (*B)(unsafe.Pointer((*A)(nil)))
return p == nil, fmt.Sprintf("%#v", p)
`);

    expect(result.values).toEqual([true, "*B(nil)"]);
  });

  test("allows methods with pointer receivers on typed nil pointers", async () => {
    const result = await expectRuns(`
type T []T

func (*T) Sum(args ...int) int {
  s := 0
  for _, v := range args {
    s += v
  }
  return s
}

return ((*T)(nil)).Sum(1, 3, 5, 7), (*T).Sum(nil, 1, 3, 5, 6)
`);

    expect(result.values).toEqual([16n, 15n]);
  });

  test("preserves concrete dynamic numeric types inside interfaces", async () => {
    const result = await expectRuns(`
type Duration int

describe := func(x interface{}) string {
  switch x.(type) {
  case int:
    return "int"
  case int64:
    return "int64"
  case uint:
    return "uint"
  case Duration:
    return "duration"
  default:
    return "other"
  }
}

var d interface{} = Duration(5)
sameDuration, okDuration := d.(Duration)
_, okInt := d.(int)

return describe(1), describe(int64(1)), describe(uint(1)), describe(Duration(1)), okDuration, int(sameDuration), okInt
`);

    expect(result.values).toEqual(["int", "int64", "uint", "duration", true, 5n, false]);
  });

  test("preserves interface dynamic types through index assignments", async () => {
    const result = await expectRuns(`
type Duration int

var xs [3]interface{}
xs[0] = 1
xs[1] = int64(2)
xs[2] = Duration(3)
_, xs0Int := xs[0].(int)
_, xs0Int64 := xs[0].(int64)
_, xs1Int64 := xs[1].(int64)
_, xs2Duration := xs[2].(Duration)

slc := make([]interface{}, 1)
slc[0] = 4
_, slcInt := slc[0].(int)

m := make(map[string]interface{})
m["x"] = 5
m["y"] = int64(6)
_, mapInt := m["x"].(int)
_, mapInt64 := m["y"].(int64)

return xs0Int, xs0Int64, xs1Int64, xs2Duration, slcInt, mapInt, mapInt64
`);

    expect(result.values).toEqual([true, false, true, true, true, true, true]);
  });

  test("treats byte and rune as aliases in interface type switches", async () => {
    const result = await expectRuns(`
var x interface{}
x = byte(1)
byteIsUint8 := false
switch x.(type) {
case uint8:
  byteIsUint8 = true
}
x = uint8(2)
uint8IsByte := false
switch x.(type) {
case byte:
  uint8IsByte = true
}
x = rune(3)
runeIsInt32 := false
switch x.(type) {
case int32:
  runeIsInt32 = true
}
x = int32(4)
int32IsRune := false
switch x.(type) {
case rune:
  int32IsRune = true
}
return byteIsUint8, uint8IsByte, runeIsInt32, int32IsRune
`);

    expect(result.values).toEqual([true, true, true, true]);
  });

  test("executes unbound type switches and rejects fallthrough", async () => {
    const ok = await expectRuns(`
x := "hello"
out := ""
switch x.(type) {
case string:
  out = "string"
default:
  out = "other"
}
return out
`);
    expect(ok.value).toBe("string");

    const bad = await evaluateSource(`
var x interface{} = 1
switch x.(type) {
case int:
  fallthrough
default:
  return "bad"
}
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("fallthrough is not allowed in type switches");
  });

  test("runs switch default only after checking non-default cases", async () => {
    const result = await expectRuns(`
value := ""
switch 2 {
default:
  value = "default"
case 2:
  value = "case"
}

typed := ""
var x interface{} = 1
switch x.(type) {
default:
  typed = "default"
case int:
  typed = "int"
}

return value, typed
`);

    expect(result.values).toEqual(["case", "int"]);
  });

  test("evaluates Go type assertions and reports mismatches", async () => {
    const ok = await expectRuns(`
type Point struct {
  X int
}

asPoint := func(x interface{}) int {
  return x.(Point).X
}

p := Point{X: 11}
ptr := &p
return asPoint(p), ptr.(*Point).X
`);
    expect(ok.values).toEqual([11n, 11n]);

    const bad = await evaluateSource(`
x := "hello"
return x.(int)
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("does not have dynamic type int");
  });

  test("enforces named interface method sets on typed variables", async () => {
    const ok = await expectRuns(`
type Stringer interface {
  String() string
}

type Point struct {
  X int
}

func (p Point) String() string {
  return fmt.Sprintf("Point(%v)", p.X)
}

var s Stringer
s = Point{X: 5}
return s.String()
`);
    expect(ok.value).toBe("Point(5)");

    const pointerOnly = await evaluateSource(`
type Mutator interface {
  Mutate()
}

type Box struct { X int }
func (b *Box) Mutate() { b.X++ }

var m Mutator
b := Box{X: 1}
m = b
`);
    expect(pointerOnly.diagnostics).toHaveLength(1);
    expect(pointerOnly.diagnostics[0]?.message).toContain("variable m");

    const pointerOk = await expectRuns(`
type Mutator interface {
  Mutate()
}

type Box struct { X int }
func (b *Box) Mutate() { b.X++ }

var m Mutator
b := Box{X: 1}
m = &b
m.Mutate()
return b.X
`);
    expect(pointerOk.value).toBe(2n);

    const aliasSignatureOk = await expectRuns(`
type BaseMode uint32
type FileMode = BaseMode

type FileInfo interface {
  Mode() BaseMode
}

type fileStat struct{}
func (*fileStat) Mode() FileMode { return 7 }

var info FileInfo = &fileStat{}
return info.Mode()
`);
    expect(aliasSignatureOk.value).toBe(7n);

    const promotedPointerOk = await expectRuns(`
type Summable interface {
  Sum(...int) int
}

type T []T
func (*T) Sum(args ...int) int {
  total := 0
  for _, v := range args {
    total += v
  }
  return total
}

type U struct { *T }

var u U
var s Summable = u
var holder struct { Summable }
holder.Summable = &u
return s.Sum(2, 3, 5, 6), holder.Sum(2, 3, 5, 8)
`);
    expect(promotedPointerOk.values).toEqual([16n, 18n]);
  });

  test("assigns package-defined value receiver types to predeclared error", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/deadline",
      files: [{
        filename: "/workspace/deadline/deadline.go",
        source: `package deadline

var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string { return "context deadline exceeded" }
`
      }]
    }]);

    expect(graph.diagnostics).toEqual([]);
  });

  test("assigns typed named integer constants to predeclared error", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/errno",
      files: [{
        filename: "/workspace/errno/errno.go",
        source: `package errno

type Errno uintptr

const EAGAIN Errno = 11

var errEAGAIN error = EAGAIN

func (e Errno) Error() string { return "try again" }
`
      }]
    }]);

    expect(graph.diagnostics).toEqual([]);
  });

  test("assigns imported error interface values to error fields without rechecking unexported concrete methods", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "errors",
        files: [{
          filename: "/usr/local/go/src/errors/errors.go",
          source: `package errors

func New(text string) error { return &errorString{s: text} }

type errorString struct { s string }

func (e *errorString) Error() string { return e.s }
`
        }]
      },
      {
        importPath: "example.com/hpack",
        files: [{
          filename: "/workspace/hpack/hpack.go",
          source: `package hpack

import "errors"

type DecodingError struct {
  Err error
}

var ErrVarintOverflow = DecodingError{errors.New("varint integer overflow")}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
  });

  test("keeps callee-owned concrete methods for imported interface returns", async () => {
    const result = await evaluateSourceFilesWithPackagesOnNode({
      importPath: "example.com/main",
      packageName: "main",
      files: [{
        filename: "main.go",
        source: `package main

import "example.com/dep"

h := dep.New()
n, err := h.Write([]byte{1, 2, 3})
if err != nil { return -1 }
return n
`
      }],
      packages: [
        {
          importPath: "example.com/hashlike",
          files: [{
            filename: "hashlike.go",
            source: `package hashlike

type Writer interface {
  Write([]byte) (int, error)
}
`
          }]
        },
        {
          importPath: "example.com/dep",
          files: [{
            filename: "dep.go",
            source: `package dep

import "example.com/hashlike"

type digest struct { total int }

func (d *digest) Write(p []byte) (int, error) {
  d.total += len(p)
  return d.total, nil
}

func New() hashlike.Writer {
  return &digest{}
}
`
          }]
        }
      ]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(3n);
  });

  test("assigns imported named integer constants to local interfaces", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "syscall",
        files: [{
          filename: "/usr/local/go/src/syscall/syscall_js.go",
          source: `package syscall

type Signal int

const (
  _ Signal = iota
  SIGCHLD
  SIGINT
)

func (s Signal) Signal() {}
`
        }]
      },
      {
        importPath: "os",
        files: [{
          filename: "/usr/local/go/src/os/exec_posix.go",
          source: `package os

import "syscall"

type Signal interface {
  Signal()
}

var Interrupt Signal = syscall.SIGINT
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
  });

  test("evaluates composite literals for imported struct types", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/poll",
        files: [{
          filename: "/usr/local/go/src/internal/poll/fd_unix.go",
          source: `package poll

type fdMutex struct{}

type FD struct {
  fdmu fdMutex
  Sysfd int
  IsStream bool
}

func (fd *FD) Init() {
  fd.Sysfd = 2
}
`
        }]
      },
      {
        importPath: "os",
        files: [{
          filename: "/usr/local/go/src/os/file_unix.go",
          source: `package os

import "internal/poll"

var PFD = poll.FD{Sysfd: 1, IsStream: true}

func init() {
  PFD.Init()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
  });

  test("represents interfaces as typed runtime values with dynamic nil state", async () => {
    const nilInterface = await expectRuns(`
type I interface {
  fun()
}

var i I
return i, i == nil
`);

    expect(nilInterface.values?.map(formatReplValue)).toEqual(["I(nil)", "true"]);

    const typedNilPointer = await expectRuns(`
type I interface {
  fun()
}

type Box struct { X int }
func (b *Box) fun() {}

var p *Box
var i I = p
q, ok := i.(*Box)
return i == nil, p == nil, q == nil, ok, i
`);

    expect(typedNilPointer.values?.slice(0, 4)).toEqual([false, true, true, true]);
    expect(formatReplValue(typedNilPointer.values?.[4] ?? null)).toBe("*Box(nil)");

    const empty = await expectRuns(`
type I interface {
  fun()
}

var i I
var x interface{} = i
return x, x == nil
`);

    expect(empty.values?.map(formatReplValue)).toEqual(["interface{}(nil)", "true"]);
  });

  test("evaluates array and slice literals with indexing, slicing, and range", async () => {
    const result = await expectRuns(`
xs := []int{1, 2, 3}
ys := [4]int{4, 5}
zs := [...]string{"a", "b", "c"}
sum := 0
for _, v := range xs {
  sum = sum + v
}
return xs[1], xs[1:3], ys, len(zs), sum
`);

    expect(result.values).toEqual([
      2n,
      [2n, 3n],
      [4n, 5n, 0n, 0n],
      3n,
      6n
    ]);
  });

  test("evaluates keyed array literals with named integer keys and inferred length", async () => {
    const result = await expectRuns(`
type Token int
const (
  zero Token = iota
  one
  two
)
tokens := [...]string{
  two: "two",
  zero: "zero",
}
return len(tokens), tokens[0], tokens[1], tokens[2]
`);

    expect(result.values).toEqual([3n, "zero", "", "two"]);
  });

  test("supports len, cap, append, and default slice/array declarations", async () => {
    const result = await expectRuns(`
var xs []int
var ys [3]string
nilBefore := xs == nil
lenBefore := len(xs)
capBefore := cap(xs)
xs = append(xs, 1, 2)
more := []int{3, 4}
xs = append(xs, more...)
return nilBefore, lenBefore, capBefore, xs, len(xs), cap(xs), len(ys), ys[0]
`);

    expect(result.values).toEqual([true, 0n, 0n, [1n, 2n, 3n, 4n], 4n, 4n, 3n, ""]);
  });

  test("supports reslicing make slices up to capacity with zero-filled backing storage", async () => {
    const result = await expectRuns(`
s := make([]int, 2, 5)
s[0] = 7
grown := s[0:5]
grown[3] = 11
again := s[0:5]
return len(s), cap(s), len(grown), cap(grown), grown[0], grown[1], grown[3], again[3]
`);

    expect(result.values).toEqual([2n, 5n, 5n, 5n, 7n, 0n, 11n, 11n]);
  });

  test("supports len and cap on pointers to arrays", async () => {
    const result = await expectRuns(`
p := new([4]int)
var nilp *[3]string
return len(p), cap(p), len(nilp), cap(nilp)
`);

    expect(result.values).toEqual([4n, 4n, 3n, 3n]);
  });

  test("supports constant identifiers in array lengths", async () => {
    const result = await expectRuns(`
const size = 4
var a [size]byte
for k := range a {
  a[k] = byte(k + 1)
}
return len(a), a[0], a[3]
`);

    expect(result.values).toEqual([4n, 1n, 4n]);
  });

  test("initializes large scalar arrays sparsely with zero values", async () => {
    const result = await expectRuns(`
type ScratchBuffer [1 << 25]byte
var memory ScratchBuffer
memory[7] = 3
return len(memory), memory[0], memory[7], memory[(1<<25)-1]
`);

    expect(result.values).toEqual([33554432n, 0n, 3n, 0n]);
  });

  test("predeclares package arrays with unexported constant lengths", async () => {
    const pkg = await evaluatePackageSourceFiles([{
      filename: "/workspace/jump/jump.go",
      source: `package jump

type Type int

const (
  InvalidType Type = iota
  StrType
  _maxtype
)

type writer interface { Write([]byte) (int, error) }
type Reader struct{}

var defuns [_maxtype]func(writer, *Reader) (int, error)

func noop(writer, *Reader) (int, error) { return 0, nil }

func init() {
  defuns = [_maxtype]func(writer, *Reader) (int, error){
    StrType: noop,
  }
}

func Len() int { return len(defuns) }
`
    }], { importPath: "example.com/jump" });

    expect(pkg.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "example.com/jump"
return jump.Len()
`, {
      packages: { "example.com/jump": pkg.package ?? {} },
      packageInfos: { "example.com/jump": pkg.packageInfo! }
    });

    expect(result.value).toBe(2n);
  });

  test("resolves imported constants in array lengths without path-qualifying the expression", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/runtime.go",
          source: `package abi

type Word uintptr

const ZeroValSize = 4
`
        }]
      },
      {
        importPath: "internal/runtime/maps",
        files: [{
          filename: "/usr/local/go/src/internal/runtime/maps/runtime.go",
          source: `package maps

import "internal/abi"

var zeroVal [abi.ZeroValSize]abi.Word

func Len() int { return len(zeroVal) }
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import maps "internal/runtime/maps"
return maps.Len()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.value).toBe(4n);
  });

  test("supports internal abi function entry intrinsics as uintptr values", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/funcpc.go",
          source: `package abi

func FuncPCABI0(f any) uintptr
func FuncPCABIInternal(f any) uintptr
`
        }]
      },
      {
        importPath: "example.com/trap",
        files: [{
          filename: "/workspace/trap/trap.go",
          source: `package trap

import "internal/abi"

func trampoline()

var P0 = abi.FuncPCABI0(trampoline)
var P1 = abi.FuncPCABIInternal(trampoline)
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import trap "example.com/trap"
return trap.P0 != 0, trap.P1 != 0
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.values).toEqual([true, true]);
  });

  test("synthesizes internal abi composite type descriptors with element metadata", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/type.go",
          source: `package abi

import "unsafe"

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
  Slice
  String
  Struct
  UnsafePointer
)

type Type struct {
  Kind_ Kind
}

func TypeOf(a any) *Type
func TypeFor[T any]() *Type { return nil }

type ArrayType struct {
  Type
  Elem *Type
  Slice *Type
  Len uintptr
}

type ChanDir int

const (
  RecvDir ChanDir = 1 << iota
  SendDir
  BothDir = RecvDir | SendDir
)

type ChanType struct {
  Type
  Elem *Type
  Dir ChanDir
}

type MapType struct {
  Type
  Key *Type
  Elem *Type
  Group *Type
  Hasher func(unsafe.Pointer, uintptr) uintptr
  GroupSize uintptr
  KeysOff uintptr
  KeyStride uintptr
  ElemsOff uintptr
  ElemStride uintptr
  ElemOff uintptr
  Flags uint32
}

type SliceType struct {
  Type
  Elem *Type
}

type PtrType struct {
  Type
  Elem *Type
}

func (t *Type) Kind() Kind { return t.Kind_ }

func (t *Type) Elem() *Type {
  switch t.Kind() {
  case Array:
    return (*ArrayType)(unsafe.Pointer(t)).Elem
  case Chan:
    return (*ChanType)(unsafe.Pointer(t)).Elem
  case Map:
    return (*MapType)(unsafe.Pointer(t)).Elem
  case Pointer:
    return (*PtrType)(unsafe.Pointer(t)).Elem
  case Slice:
    return (*SliceType)(unsafe.Pointer(t)).Elem
  }
  return nil
}

func (t *Type) Key() *Type {
  if t.Kind() == Map {
    return (*MapType)(unsafe.Pointer(t)).Key
  }
  return nil
}

func (t *Type) Len() int {
  if t.Kind() == Array {
    return int((*ArrayType)(unsafe.Pointer(t)).Len)
  }
  return 0
}

func (t *Type) ChanDir() ChanDir {
  if t.Kind() == Chan {
    return (*ChanType)(unsafe.Pointer(t)).Dir
  }
  return 0
}
`
        }]
      },
      {
        importPath: "reflect",
        files: [{
          filename: "/usr/local/go/src/reflect/type.go",
          source: `package reflect

import "internal/abi"

func TypeFor[T any]() *abi.Type {
  return abi.TypeFor[T]()
}
`
        }]
      },
      {
        importPath: "example.com/model",
        files: [{
          filename: "/workspace/model/model.go",
          source: `package model

type Local struct {
  X int
}

type Node struct {
  Next *Node
  Kids []*Node
  Value any
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "internal/abi"
import r "reflect"
import "example.com/model"

s := []int{1}
a := [2]string{"x", "y"}
m := map[string]int{"x": 1}
c := make(chan bool, 1)
p := new(int)
pt := r.TypeFor[*model.Local]()
node := r.TypeFor[model.Node]()
nodePtr := r.TypeFor[*model.Node]()

return abi.TypeOf(s).Kind(), abi.TypeOf(s).Elem().Kind(),
  abi.TypeOf(a).Kind(), abi.TypeOf(a).Elem().Kind(), abi.TypeOf(a).Len(),
  abi.TypeOf(m).Kind(), abi.TypeOf(m).Key().Kind(), abi.TypeOf(m).Elem().Kind(),
  abi.TypeOf(c).Kind(), abi.TypeOf(c).Elem().Kind(), abi.TypeOf(c).ChanDir(),
  abi.TypeOf(p).Kind(), abi.TypeOf(p).Elem().Kind(),
  r.TypeFor[model.Local]().Kind(), pt.Kind(), pt.Elem().Kind(),
  node.Kind(), nodePtr.Kind(), nodePtr.Elem().Kind()
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([
      23n, 2n,
      17n, 24n, 2n,
      21n, 24n, 2n,
      18n, 1n, 3n,
      22n, 2n,
      25n, 22n, 25n,
      25n, 22n, 25n
    ]);
  });

  test("preserves abi descriptor metadata through reflect rtype reinterpretation", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/type.go",
          source: `package abi

import "unsafe"

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
  Slice
  String
  Struct
  UnsafePointer
)

type Type struct {
  Kind_ Kind
}

type PtrType struct {
  Type
  Elem *Type
}

func TypeOf(a any) *Type
func TypeFor[T any]() *Type { return nil }
func (t *Type) Kind() Kind { return t.Kind_ }
func (t *Type) Elem() *Type {
  if t.Kind() == Pointer {
    return (*PtrType)(unsafe.Pointer(t)).Elem
  }
  return nil
}
`
        }]
      },
      {
        importPath: "reflect",
        files: [{
          filename: "/usr/local/go/src/reflect/type.go",
          source: `package reflect

import (
  "internal/abi"
  "unsafe"
)

type Kind = abi.Kind

const (
  Invalid = abi.Invalid
  Interface = abi.Interface
  Pointer = abi.Pointer
)

type Type interface {
  Kind() Kind
  Elem() Type
}

type rtype struct {
  t abi.Type
}

func (t *rtype) common() *abi.Type { return &t.t }
func toRType(t *abi.Type) *rtype { return (*rtype)(unsafe.Pointer(t)) }
func toType(t *abi.Type) Type {
  if t == nil {
    return nil
  }
  return toRType(t)
}
func (t *rtype) Kind() Kind { return Kind(t.t.Kind()) }
func (t *rtype) Elem() Type { return toType(t.common().Elem()) }
func TypeOf(i any) Type { return toType(abi.TypeOf(i)) }
func TypeFor[T any]() Type { return toRType(abi.TypeFor[T]()) }
`
        }]
      },
      {
        importPath: "example.com/goblike",
        files: [{
          filename: "/workspace/goblike/type.go",
          source: `package goblike

import "reflect"

type GobEncoder interface {
  GobEncode() ([]byte, error)
}

var EncoderType = reflect.TypeFor[GobEncoder]()
var EncoderKind = EncoderType.Kind()
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "reflect"
import g "example.com/goblike"

t := reflect.TypeOf((*error)(nil))
return int(t.Kind()), int(t.Elem().Kind()), int(g.EncoderKind)
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([22n, 20n, 20n]);
  });

  test("uses stable reflect Type identity as a map key", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/type.go",
          source: `package abi

import "unsafe"

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
  Slice
  String
  Struct
  UnsafePointer
)

type Type struct {
  Kind_ Kind
}

type PtrType struct {
  Type
  Elem *Type
}

func TypeOf(a any) *Type
func TypeFor[T any]() *Type { return nil }
func (t *Type) Kind() Kind { return t.Kind_ }
func (t *Type) Elem() *Type {
  if t.Kind() == Pointer {
    return (*PtrType)(unsafe.Pointer(t)).Elem
  }
  return nil
}
`
        }]
      },
      {
        importPath: "reflect",
        files: [{
          filename: "/usr/local/go/src/reflect/type.go",
          source: `package reflect

import (
  "internal/abi"
  "unsafe"
)

type Kind = abi.Kind

const (
  Invalid = abi.Invalid
  Pointer = abi.Pointer
)

type Type interface {
  Kind() Kind
  Elem() Type
}

type rtype struct {
  t abi.Type
}

func (t *rtype) common() *abi.Type { return &t.t }
func toRType(t *abi.Type) *rtype { return (*rtype)(unsafe.Pointer(t)) }
func toType(t *abi.Type) Type {
  if t == nil {
    return nil
  }
  return toRType(t)
}
func (t *rtype) Kind() Kind { return Kind(t.t.Kind()) }
func (t *rtype) Elem() Type { return toType(t.common().Elem()) }
func TypeOf(i any) Type { return toType(abi.TypeOf(i)) }
func TypeFor[T any]() Type { return toRType(abi.TypeFor[T]()) }
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "reflect"

type wireType struct {
  A *int
}

a := reflect.TypeFor[wireType]()
b := reflect.TypeFor[wireType]()
c := reflect.TypeOf((*wireType)(nil)).Elem()
m := make(map[reflect.Type]int)
m[a] = 11
firstB := m[b]
firstC := m[c]
m[b] = 12
m[c] = 13
return firstB, firstC, len(m), m[a], m[b], m[c]
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([11n, 11n, 1n, 13n, 13n, 13n]);
  });

  test("reflect descriptors preserve named underlying scalar and slice kinds", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/type.go",
          source: `package abi

import "unsafe"

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
  Slice
  String
  Struct
  UnsafePointer
)

type Type struct {
  Kind_ Kind
}

type SliceType struct {
  Type
  Elem *Type
}

func TypeFor[T any]() *Type { return nil }
func TypeOf(a any) *Type { return nil }
func (t *Type) Kind() Kind { return t.Kind_ }
func (t *Type) Elem() *Type {
  if t.Kind() == Slice {
    return (*SliceType)(unsafe.Pointer(t)).Elem
  }
  return nil
}
func (t *Type) HasName() bool { return false }
`
        }]
      },
      {
        importPath: "reflect",
        files: [{
          filename: "/usr/local/go/src/reflect/type.go",
          source: `package reflect

import (
  "internal/abi"
  "unsafe"
)

type Kind = abi.Kind

const (
  Invalid = abi.Invalid
  Slice = abi.Slice
)

type Type interface {
  Kind() Kind
  Name() string
  String() string
  Elem() Type
}

type rtype struct {
  t abi.Type
}

func (t *rtype) common() *abi.Type { return &t.t }
func toRType(t *abi.Type) *rtype { return (*rtype)(unsafe.Pointer(t)) }
func toType(t *abi.Type) Type {
  if t == nil {
    return nil
  }
  return toRType(t)
}
func (t *rtype) Kind() Kind { return Kind(t.t.Kind()) }
func (t *rtype) Name() string { return "" }
func (t *rtype) String() string { return "" }
func (t *rtype) Elem() Type { return toType(t.common().Elem()) }
func TypeFor[T any]() Type { return toRType(abi.TypeFor[T]()) }
`
        }]
      },
      {
        importPath: "example.com/named",
        files: [{
          filename: "/workspace/named/named.go",
          source: `package named

import "reflect"

type typeId int32
type idSlice []typeId

func Inspect() (reflect.Kind, string, string, reflect.Kind, string, string, reflect.Kind, string, string) {
  tid := reflect.TypeFor[typeId]()
  sid := reflect.TypeFor[idSlice]()
  elem := sid.Elem()
  return tid.Kind(), tid.Name(), tid.String(), sid.Kind(), sid.Name(), sid.String(), elem.Kind(), elem.Name(), elem.String()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import n "example.com/named"

tidKind, tidName, tidText, sidKind, sidName, sidText, elemKind, elemName, elemText := n.Inspect()
return int(tidKind), tidName, tidText, int(sidKind), sidName, sidText, int(elemKind), elemName, elemText
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([5n, "typeId", "named.typeId", 23n, "idSlice", "named.idSlice", 5n, "typeId", "named.typeId"]);
  });

  test("exposes runtime abi struct fields through reflect Type", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "internal/abi",
        files: [{
          filename: "/usr/local/go/src/internal/abi/type.go",
          source: `package abi

import "unsafe"

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
  Slice
  String
  Struct
  UnsafePointer
)

type Type struct {
  Kind_ Kind
}

type Name struct {
  Bytes *byte
}

type StructField struct {
  Name Name
  Typ *Type
  Offset uintptr
}

type ArrayType struct {
  Type
  Elem *Type
  Slice *Type
  Len uintptr
}

type MapType struct {
  Type
  Key *Type
  Elem *Type
}

type SliceType struct {
  Type
  Elem *Type
}

type PtrType struct {
  Type
  Elem *Type
}

type StructType struct {
  Type
  PkgPath Name
  Fields []StructField
}

func TypeFor[T any]() *Type { return nil }
func TypeOf(a any) *Type { return nil }
func (t *Type) Kind() Kind { return t.Kind_ }
func (t *Type) Elem() *Type {
  switch t.Kind() {
  case Array:
    return (*ArrayType)(unsafe.Pointer(t)).Elem
  case Map:
    return (*MapType)(unsafe.Pointer(t)).Elem
  case Pointer:
    return (*PtrType)(unsafe.Pointer(t)).Elem
  case Slice:
    return (*SliceType)(unsafe.Pointer(t)).Elem
  }
  return nil
}
func (t *Type) Key() *Type {
  if t.Kind() == Map {
    return (*MapType)(unsafe.Pointer(t)).Key
  }
  return nil
}
func (t *Type) Len() int {
  if t.Kind() == Array {
    return int((*ArrayType)(unsafe.Pointer(t)).Len)
  }
  return 0
}
func (n Name) Name() string { return "" }
func (n Name) Tag() string { return "" }
func (n Name) IsExported() bool { return false }
func (n Name) IsEmbedded() bool { return false }
func (f *StructField) Embedded() bool { return f.Name.IsEmbedded() }

var _ unsafe.Pointer
`
        }]
      },
      {
        importPath: "reflect",
        files: [{
          filename: "/usr/local/go/src/reflect/type.go",
          source: `package reflect

import (
  "internal/abi"
  "unsafe"
)

type Kind = abi.Kind

const (
  Invalid = abi.Invalid
  Int = abi.Int
  Int32 = abi.Int32
  Array = abi.Array
  Map = abi.Map
  Pointer = abi.Pointer
  Struct = abi.Struct
  Slice = abi.Slice
  String = abi.String
)

type Type interface {
  Kind() Kind
  Elem() Type
  Key() Type
  Len() int
  Name() string
  String() string
  PkgPath() string
  NumField() int
  Field(int) StructField
}

type StructTag string

type StructField struct {
  Name string
  Type Type
  Tag StructTag
  Offset uintptr
  Index []int
  Anonymous bool
}

type rtype struct {
  t abi.Type
}

type structField = abi.StructField

type structType struct {
  abi.StructType
}

func toRType(t *abi.Type) *rtype { return (*rtype)(unsafe.Pointer(t)) }
func toType(t *abi.Type) Type {
  if t == nil {
    return nil
  }
  return toRType(t)
}
func (t *rtype) Kind() Kind { return Kind(t.t.Kind()) }
func (t *rtype) Elem() Type { return toType(t.t.Elem()) }
func (t *rtype) Key() Type { return toType(t.t.Key()) }
func (t *rtype) Len() int { return t.t.Len() }
func (t *rtype) Name() string { return "" }
func (t *rtype) String() string { return "" }
func (t *rtype) PkgPath() string { return "" }
func (t *rtype) NumField() int {
  if t.Kind() != Struct {
    panic("NumField of non-struct")
  }
  tt := (*structType)(unsafe.Pointer(t))
  return len(tt.Fields)
}
func (t *rtype) Field(i int) StructField {
  if t.Kind() != Struct {
    panic("Field of non-struct")
  }
  tt := (*structType)(unsafe.Pointer(t))
  return tt.Field(i)
}
func (t *structType) Field(i int) (f StructField) {
  p := &t.Fields[i]
  f.Type = toType(p.Typ)
  f.Name = p.Name.Name()
  f.Anonymous = p.Embedded()
  if tag := p.Name.Tag(); tag != "" {
    f.Tag = StructTag(tag)
  }
  f.Offset = p.Offset
  f.Index = []int{i}
  return
}
func TypeFor[T any]() Type { return toRType(abi.TypeFor[T]()) }
func TypeOf(i any) Type { return toType(abi.TypeOf(i)) }
`
        }]
      },
      {
        importPath: "example.com/goblike",
        files: [{
          filename: "/workspace/goblike/type.go",
          source: `package goblike

import "reflect"

type fieldType struct {
  Name string
  Id int32
}

type structType struct {
  CommonType
  Field []fieldType
}

type CommonType struct {
  Name string
}

func Inspect() (int, string, bool, string, reflect.Kind, reflect.Kind, string, string, string) {
  t := reflect.TypeFor[structType]()
  f0 := t.Field(0)
  f1 := t.Field(1)
  return t.NumField(), f0.Name, f0.Anonymous, f1.Name, f1.Type.Kind(), f1.Type.Elem().Kind(), t.Name(), t.String(), t.PkgPath()
}
`
        }]
      },
      {
        importPath: "example.com/other",
        files: [{
          filename: "/workspace/other/type.go",
          source: `package other

import "reflect"

type structType struct {
  Only int
}

func Inspect() (int, string, string, string) {
  t := reflect.TypeFor[structType]()
  f := t.Field(0)
  return t.NumField(), f.Name, t.Name(), t.PkgPath()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import g "example.com/goblike"
import o "example.com/other"

n, first, anonymous, second, kind, elemKind, name, text, pkgPath := g.Inspect()
otherN, otherFirst, otherName, otherPkgPath := o.Inspect()
return n, first, anonymous, second, int(kind), int(elemKind), name, text, pkgPath, otherN, otherFirst, otherName, otherPkgPath
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.values).toEqual([
      2n,
      "CommonType",
      true,
      "Field",
      23n,
      25n,
      "structType",
      "goblike.structType",
      "example.com/goblike",
      1n,
      "Only",
      "structType",
      "example.com/other"
    ]);
  });

  test("initializes standard encoding/gob bootstrap descriptors", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/gobmain/main.go",
      source: `package main

import _ "encoding/gob"

func main() {}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
  });

  test("source-built reflect preserves TypeOf for typed nil maps", async () => {
    const sourcePackageProvider = createNodeSourcePackageProvider([]);
    if (!sourcePackageProvider) throw new Error("node source package provider is unavailable");

    const result = await runMainSourcePackageFiles([{
      filename: "/workspace/reflectmap/main.go",
      source: `package main

import "reflect"

func main() {
  t := reflect.TypeOf(map[string]interface{}(nil))
  if t == nil {
    panic("typed nil map lost its reflect type")
  }
  print(t.String() + "\\n")
  print(reflect.ValueOf(t).Pointer() != 0)
  print("\\n")
}
`
    }], {
      sourcePackageProvider
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual(["map[string]interface{}\n", "true", "\n"]);
  });

  test("gojr test root package types stay visible to source-built reflect", async () => {
    const result = await testSourceFilesWithPackagesOnNode({
      importPath: "example.com/p",
      packageName: "p",
      testVerbose: true,
      testRun: "TestReflectFieldFromRootTestPackage",
      files: [{
        filename: "/workspace/p/table_test.go",
        source: `package p

import (
  "reflect"
  "testing"
)

type Table struct {
  Headers []string \`json:"headers" msg:"headers"\`
  Rows [][]string
}

type RegisteredType struct {
  Factory func() (interface{}, error)
}

var R = &RegisteredType{Factory: func() (interface{}, error) {
  return &Table{}, nil
}}

func TestReflectFieldFromRootTestPackage(t *testing.T) {
  rs, err := R.Factory()
  if err != nil {
    t.Fatal(err)
  }
  tye := reflect.ValueOf(rs).Elem().Type()
  if got := tye.Kind().String(); got != "struct" {
    t.Fatalf("kind = %q", got)
  }
  if got := tye.NumField(); got != 2 {
    t.Fatalf("NumField = %d", got)
  }
  fld := tye.Field(0)
  if fld.Name != "Headers" {
    t.Fatalf("field name = %q", fld.Name)
  }
  if got := string(fld.Tag); got != "json:\\"headers\\" msg:\\"headers\\"" {
    t.Fatalf("field tag = %q", got)
  }
  if got := fld.Type.String(); got != "[]string" {
    t.Fatalf("field type = %q", got)
  }
}
`
      }]
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.output).toContain("--- PASS: TestReflectFieldFromRootTestPackage\n");
  });

  test("bodyless package functions return declared zero result values", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/bodyless",
        files: [{
          filename: "/workspace/bodyless/bodyless.go",
          source: `package bodyless

type Errno uintptr

func rawSyscall(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)

func Call() (uintptr, uintptr, Errno) {
  return rawSyscall(1, 2, 3, 4)
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import bodyless "example.com/bodyless"
r1, r2, err := bodyless.Call()
return r1, r2, uintptr(err), err == 0
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });
    expect(result.values).toEqual([0n, 0n, 0n, true]);
  });

  test("reports typed array and slice literal element mismatches", async () => {
    const result = await evaluateSource(`
xs := []int{1, "bad"}
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("array element bad is not assignable to int");
  });

  test("supports len on strings and maps", async () => {
    const result = await expectRuns(`
m := map[string]int{"a": 1, "b": 2}
return len("hiya"), len(m)
`);

    expect(result.values).toEqual([4n, 2n]);
  });

  test("reports panicOn failures as runtime diagnostics", async () => {
    const result = await evaluateSource(`
panicOn("bad")
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_PANIC001");
  });

  test("keeps REPL session locals across eager evaluations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    const result = await session.evaluate("a + 2");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(12n);
  });

  test("checks REPL const groups with scoped iota", async () => {
    const session = new GoJuniorSession();

    const result = await session.evaluate(`
const (
  abit, amask = 1 << iota, 1<<iota - 1
  bbit, bmask
)
return abit, amask, bbit, bmask
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.values).toEqual([1n, 0n, 2n, 1n]);
  });

  test("tracks REPL sheet dependencies when sheet data changes", async () => {
    const session = new GoJuniorSession();

    session.setSheet({ A1: 40n });
    const result = await session.evaluate("sheet.A1.(int64) + 2");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(42n);
    expect(result.observedDeps).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" })
    ]);

    const next = await session.evaluate("sheet.A1:B1");
    expect(next.diagnostics).toEqual([]);
    expect(next.observedDeps).toEqual([
      rangeDependency("sheet", "A1", "B1")
    ]);
  });

  test("evaluates Go raw string literals", async () => {
    const result = await expectRuns("a := `hi\nthere`\nreturn a");

    expect(result.value).toBe("hi\nthere");
  });

  test("supports Go-style increment and decrement statements in sessions", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("a++")).diagnostics).toEqual([]);
    expect((await session.evaluate("a")).value).toBe(11n);
    expect((await session.evaluate("a--")).diagnostics).toEqual([]);
    expect((await session.evaluate("a")).value).toBe(10n);
  });

  test("accepts top-level semicolons in REPL session input", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10;")).diagnostics).toEqual([]);
    const result = await session.evaluate("a");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(10n);
  });

  test("accepts import-only REPL input and does not replay output", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`fmt.Printf("hi")`)).output).toEqual(["hi"]);
    expect((await session.evaluate(`import "cmp"`)).diagnostics).toEqual([]);
    const compare = await session.evaluate(`cmp.Compare(1, 2)`);
    expect(compare.diagnostics).toEqual([]);
    expect(compare.value).toBe(-1n);
    expect((await session.evaluate("a := 1")).output).toEqual([]);
  });

  test("defines and calls functions with grouped names in REPL sessions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate("func f(a, b, c int) (x, y, z int) { return a, b, c }");
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f(1, 2, 3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([1n, 2n, 3n]);
  });

  test("supports named result variables and naked returns in REPL functions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func swap(a, b int) (left, right int) {
  left = b
  right = a
  return
}
`);
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("swap(1, 2)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([2n, 1n]);
  });

  test("supports variadic parameters and spread calls in REPL functions", async () => {
    const session = new GoJuniorSession({
      sheet: {
        A1: [4n, 5n, 6n]
      }
    });

    const define = await session.evaluate(`
func sum(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
    expect(define.diagnostics).toEqual([]);

    const direct = await session.evaluate("sum(1, 2, 3)");
    expect(direct.diagnostics).toEqual([]);
    expect(direct.value).toBe(6n);

    const spread = await session.evaluate("sum(sheet.A1.([]int)...)");
    expect(spread.diagnostics).toEqual([]);
    expect(spread.value).toBe(15n);
  });

  test("spreads typed nil slices as zero variadic arguments", async () => {
    const result = await expectRuns(`
func count(vals ...int) int {
  return len(vals)
}
var xs []int
return count(xs...), count(99, xs...)
`);

    expect(result.values).toEqual([0n, 1n]);
  });

  test("supports multiple short declarations and assignments", async () => {
    const result = await expectRuns(`
a, b := 1, 2
a, b = b, a
return a, b
`);

    expect(result.values).toEqual([2n, 1n]);
  });

  test("supports destructuring multiple function returns", async () => {
    const result = await expectRuns(`
func divmod(x, y int) (int, int) {
  return x / y, x % y
}

q, r := divmod(17, 5)
_, onlyR := divmod(19, 5)
return q, r, onlyR
`);

    expect(result.values).toEqual([3n, 2n, 4n]);
  });

  test("expands a single multi-result call in return statements", async () => {
    const result = await expectRuns(`
func pair() (int, error) {
  return 7, nil
}

func forward() (int, error) {
  return pair()
}

v, err := forward()
return v, err == nil
`);

    expect(result.values).toEqual([7n, true]);
  });

  test("infers tuple result slot types for short and var declarations", async () => {
    const result = await expectRuns(`
func words() (uint64, uint64) {
  return 7, uint64(10254876495507714224)
}

hi, lo := words()
var a, b = words()
return hi, lo, a, b
`);

    expect(result.values).toEqual([7n, 10254876495507714224n, 7n, 10254876495507714224n]);
  });

  test("infers tuple result slot types from method signatures", async () => {
    const result = await expectRuns(`
type reader struct{}

func (reader) words() (uint64, bool) {
  return uint64(18446744070991911616), true
}

r := reader{}
n8, ok := r.words()
n := int64(n8)
return n8, ok, n
`);

    expect(result.values).toEqual([18446744070991911616n, true, -2717640000n]);
  });

  test("expands a single multi-result call into another call argument list", async () => {
    const result = await expectRuns(`
func pair() (int, int) {
  return 3, 4
}

func add(a, b int) int {
  return a + b
}

return add(pair())
`);

    expect(result.value).toBe(7n);
  });

  test("supports multi-assignment to index targets after evaluating rhs", async () => {
    const result = await expectRuns(`
xs := []int{1, 2}
xs[0], xs[1] = xs[1], xs[0]
return xs
`);

    expect(result.value).toEqual([2n, 1n]);
  });

  test("defines and calls function literals with grouped names in REPL sessions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate("f := func(a, b, c int) (d, e, f int) { return b, c, a }");
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f(1, 2, 3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([2n, 3n, 1n]);
  });

  test("closures capture lexical variables by reference", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("base := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("addBase := func(x int) int { return base + x }")).diagnostics).toEqual([]);
    expect((await session.evaluate("base = 20")).diagnostics).toEqual([]);

    const call = await session.evaluate("addBase(2)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toBe(22n);
  });

  test("keeps REPL top-level channel variables visible to later function declarations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("func f() { for i := range 5 { c <- i } }")).diagnostics).toEqual([]);
    expect((await session.evaluate("go f()")).diagnostics).toEqual([]);

    const first = await session.evaluate("<-c");
    const second = await session.evaluate("<-c");

    expect(first.diagnostics).toEqual([]);
    expect(second.diagnostics).toEqual([]);
    expect(first.value).toBe(0n);
    expect(second.value).toBe(1n);
  });

  test("returned closures keep their defining function scope alive", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func makeAdder(base int) func(int) int {
  return func(x int) int {
    return base + x
  }
}
`);
    expect(define.diagnostics).toEqual([]);
    expect((await session.evaluate("add5 := makeAdder(5)")).diagnostics).toEqual([]);

    const call = await session.evaluate("add5(3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toBe(8n);
  });

  test("function literals support variadic parameters and spread calls", async () => {
    const session = new GoJuniorSession({
      sheet: {
        A1: [7n, 8n, 9n]
      }
    });

    const define = await session.evaluate(`
sum := func(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
    expect(define.diagnostics).toEqual([]);

    const direct = await session.evaluate("sum(1, 2, 3)");
    expect(direct.diagnostics).toEqual([]);
    expect(direct.value).toBe(6n);

    const spread = await session.evaluate("sum(sheet.A1.([]int)...)");
    expect(spread.diagnostics).toEqual([]);
    expect(spread.value).toBe(24n);
  });

  test("marks incomplete REPL input without executing it", async () => {
    const session = new GoJuniorSession();
    const incomplete = await session.evaluate("if true {");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.line).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.column).toBeGreaterThan(0);

    const complete = await session.evaluate(`
if true {
  return 1
}
`);

    expect(complete.diagnostics).toEqual([]);
    expect(complete.value).toBe(1n);
  });

  test("keeps incomplete raw strings pending in REPL input", async () => {
    const session = new GoJuniorSession();
    const incomplete = await session.evaluate("a := ` hi there");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.message).toContain("unterminated raw string");
  });

  test("keeps for blocks with increment statements pending until closed", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("for {\n  a++");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.span?.line).not.toBeNaN();
    expect(result.diagnostics[0]?.span?.column).not.toBeNaN();
  });

  test("keeps Go-style for clauses pending until the block is closed", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("for b := 0; b < 10; b++ {");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.message).not.toContain("':='");
  });

  test("keeps bare labels pending until their statement is entered", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("top:");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
  });

  test("executes Go-style for init condition and post clauses", async () => {
    const result = await expectRuns(`
sum := 0
for b := 0; b < 10; b++ {
  sum = sum + b
}
return sum
`);

    expect(result.value).toBe(45n);
  });

  test("executes Go-style for clauses with omitted init statements", async () => {
    const result = await expectRuns(`
sum := 0
b := 0
for ; b < 5; b++ {
  if b == 2 {
    continue
  }
  sum += b
}
for ; ; b-- {
  if b == 0 {
    break
  }
  sum++
}
return sum
`);

    expect(result.value).toBe(13n);
  });

  test("runs for post clause after continue", async () => {
    const result = await expectRuns(`
sum := 0
for b := 0; b < 5; b++ {
  if b == 2 {
    continue
  }
  sum = sum + b
}
return sum
`);

    expect(result.value).toBe(8n);
  });

  test("supports REPL entry of labeled nested loops with labeled break", async () => {
    const session = new GoJuniorSession();
    const lines = [
      "top:",
      "for i := 0; i < 5; i++ {",
      "  inner:",
      "  for j := 0; j < 10; j++ {",
      "    fmt.Printf(\"i=%v j=%v\\n\", i, j)",
      "    if j == 0 { continue inner }",
      "    break top",
      "  }",
      "}",
    ];
    let source = "";
    for (const [index, line] of lines.entries()) {
      source += `${line}\n`;
      const result = await session.evaluate(source);
      if (index < lines.length - 1) {
        expect(result.incomplete).toBe(true);
      } else {
        expect(result.diagnostics).toEqual([]);
        expect(result.output).toEqual(["i=0 j=0\n", "i=0 j=1\n"]);
      }
    }
  });

  test("supports goto to forward and backward labels", async () => {
    const forward = await expectRuns(`
i := 0
goto Done
i = 99
Done:
return i
`);
    expect(forward.value).toBe(0n);

    const backward = await expectRuns(`
i := 0
Loop:
i++
if i < 3 {
  goto Loop
}
return i
`);
    expect(backward.value).toBe(3n);
  });

  test("supports labeled break out of nested loops", async () => {
    const result = await expectRuns(`
count := 0
Outer:
for i := 0; i < 3; i++ {
  for j := 0; j < 3; j++ {
    count = count + 1
    break Outer
  }
}
return count
`);

    expect(result.value).toBe(1n);
  });

  test("supports labeled continue from inside a switch to an outer for", async () => {
    const result = await expectRuns(`
sum := 0
Outer:
for i := 0; i < 4; i++ {
  switch i {
  case 2:
    continue Outer
  }
  sum = sum + i
}
return sum
`);

    expect(result.value).toBe(4n);
  });

  test("keeps unlabeled break scoped to switch inside for", async () => {
    const result = await expectRuns(`
sum := 0
for i := 0; i < 3; i++ {
  switch i {
  case 1:
    break
  }
  sum = sum + 1
}
return sum
`);

    expect(result.value).toBe(3n);
  });

  test("lets unlabeled continue inside switch continue the containing for", async () => {
    const result = await expectRuns(`
sum := 0
for i := 0; i < 3; i++ {
  switch i {
  case 1:
    continue
  }
  sum = sum + 1
}
return sum
`);

    expect(result.value).toBe(2n);
  });

  test("supports Go conversions, numeric literals, rune literals, imaginary literals, and complex builtins", async () => {
    const result = await expectRuns(`
a := int(0b1010)
b := int64(0x10)
c := rune('A')
d := byte(0o7)
z := complex(float64(a), 2.5)
return a, b, c, d, real(z), imag(z), 3i + 2i
`);

    expect(result.values).toEqual([10n, 16n, 65n, 7n, 10, 2.5, { real: 0, imag: 5 }]);
  });

  test("uses named bool values in control flow", async () => {
    const result = await expectRuns(`
type Truth bool

var t Truth = true
if t {
  return "yes"
}
return "no"
`);

    expect(formatReplValue(result.value ?? null)).toBe(`"yes"`);
  });

  test("rounds float32 assignments, conversions, and expression results like Go", async () => {
    const result = await expectRuns(`
func f32(v float64) float32 { return float32(v) }

var f09 float32 = 1e-10
var f10 float32 = 1e+10
var f13 float32 = .1e-10
var f14 float32 = .1e+10
var c64 complex64 = complex(1.1, .1e-10)
return f13 == f09/10.0, f14 == f10/10.0, float64(float32(1.1)) == float64(f32(1.1)), real(c64) == f32(1.1)
`);

    expect(result.values).toEqual([true, true, true, true]);
  });

  test("wraps typed integer expression results like Go", async () => {
    const result = await expectRuns(`
func f8(x, y int8) (int8, int8) {
  return x / y, x % y
}
func f16(x, y int16) (int16, int16) {
  return x / y, x % y
}

q8, r8 := f8(-1<<7, -1)
q16, r16 := f16(-1<<15, -1)
return q8, r8, q16, r16
`);

    expect(result.values).toEqual([-128n, 0n, -32768n, 0n]);
  });

  test("infers selector-based uint64 shift expressions from field types", async () => {
    const result = await expectRuns(`
type T struct {
  U uint64
  V uint64
}

t := T{U: 2251799813685249, V: 1}
u0 := t.U<<51 | t.V
return u0
`);

    expect(result.value).toBe(2251799813685249n);
  });

  test("infers short var uint64 types from package array element indexes", async () => {
    const result = await expectRuns(`
var iv = [2]uint64{1, 13503953896175478587}

func grab() uint64 {
  v9 := iv[1]
  return v9
}

return grab()
`);

    expect(result.value).toBe(13503953896175478587n);
  });

  test("wraps typed unsigned compound assignment and incdec like Go", async () => {
    const result = await expectRuns(`
var a uint64 = 18446744073709551615
a += 1
var b uint64 = 1
b <<= 64
var c uint8 = 255
c++
var d uint8
d--
return a, b, c, d
`);

    expect(result.values).toEqual([0n, 0n, 0n, 255n]);
  });

  test("keeps exact untyped exponent constants in typed integer operations", async () => {
    const result = await expectRuns(`
func f(u uint64) (uint64, uint32, uint32) {
  rem := uint32(u % 1e8)
  u /= 1e8
  x := uint32(u)
  return u, x, rem
}

a, b, c := f(567000000)
return a, b, c
`);

    expect(result.values).toEqual([5n, 5n, 67000000n]);
  });

  test("declares range values using array and slice element types", async () => {
    const result = await expectRuns(`
var words = [2]uint64{1, 10231371594470170519}
var got uint64
for _, s := range words[:] {
  got = s
}
return got
`);

    expect(result.value).toBe(10231371594470170519n);
  });

  test("predeclares package variables with inferred checked types across source files", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/u64",
      files: [
        {
          filename: "/workspace/u64/data.go",
          source: `package u64

var iv = [2]uint64{1, 13503953896175478587}
`
        },
        {
          filename: "/workspace/u64/grab.go",
          source: `package u64

func Grab() uint64 {
  v9 := iv[1]
  return v9
}

var Got = Grab()
`
        }
      ]
    }]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages["example.com/u64"]?.Got).toBe(13503953896175478587n);
  });

  test("infers short var types from single-result function signatures", async () => {
    const result = await expectRuns(`
func mask64Bits(cond int) uint64 {
  return ^(uint64(cond) - 1)
}

m := mask64Bits(1)
return m
`);

    expect(result.value).toBe(18446744073709551615n);
  });

  test("preserves named numeric expression result identity", async () => {
    const result = await expectRuns(`
type A int
var a A = 1

_, plusA := interface{}(+a).(A)
_, addA := interface{}(a + 0).(A)
_, addInt := interface{}(a + 0).(int)
return plusA, addA, addInt
`);

    expect(result.values).toEqual([true, true, false]);
  });

  test("wraps explicit integer conversions like Go", async () => {
    const result = await expectRuns(`
a := int8(-32668)
b := uint8(-1)
c := int16(65535)
d := int32(uint32(0xffffffff))
e := uint64(-1)
return a, b, c, d, e
`);

    expect(result.values).toEqual([100n, 255n, -1n, -1n, 18446744073709551615n]);

    const assignment = await evaluateSource(`
var x int8 = 128
`);
    expect(assignment.diagnostics).toHaveLength(1);
    expect(assignment.diagnostics[0]?.message).toContain("not assignable to int8");
  });

  test("supports Go string conversions from byte and rune slices", async () => {
    const result = await expectRuns(`
bs := []byte{0xe1, 0x88, 0xb4}
rs := []rune{'a', '\\u1234', 'c'}
p := new([3]byte)
p[0] = 'x'
p[1] = 'y'
p[2] = 'z'
return string(bs), string(rs), string(p[0:])
`);

    expect(result.values).toEqual(["\u1234", "a\u1234c", "xyz"]);
  });

  test("converts nil byte and rune slices to empty strings", async () => {
    const result = await expectRuns(`
type Bytes []byte
var bs []byte
var rs []rune
var named Bytes
return string(bs), string(rs), string(named)
`);

    expect(result.values).toEqual(["", "", ""]);
  });

  test("preserves package string constant byte escapes during var initialization", async () => {
    const result = await evaluatePackageSourceFiles([{
      filename: "p.go",
      source: `package p

const encodedBytes = "\\xff\\xff"

type Box struct {
  data [2]uint8
}

func NewBox() *Box {
  b := new(Box)
  copy(b.data[:], encodedBytes)
  return b
}

var B = NewBox()
var First = B.data[0]
`
    }], { importPath: "example.com/p" });

    expect(result.diagnostics).toEqual([]);
    expect(result.package?.First).toBe(255n);
  });

  test("treats Go strings as byte sequences for len index slice and escapes", async () => {
    const result = await expectRuns(`
s := "aä本☺"
largest := string(0x10ffff)
encoded := "\\xf4\\x8f\\xbf\\xbf"
bad := "\\xff\\xff"
return len(s), s[1], s[1:3], largest == encoded, []rune(bad), string([]byte(bad))
`);

    expect(result.values).toEqual([9n, 195n, "ä", true, [65533n, 65533n], "\ufffd\ufffd"]);
  });

  test("supports Go byte and rune slice conversions from strings", async () => {
    const result = await expectRuns(`
type Bytes []byte
type Runes []rune
s := "aä本☺"
bs := []byte(s)
rs := []rune(s)
nbs := Bytes(s)
nrs := Runes(s)
return bs, rs, string(bs), string(rs), string(nbs), string(nrs)
`);

    expect(result.values).toEqual([
      [97n, 195n, 164n, 230n, 156n, 172n, 226n, 152n, 186n],
      [97n, 228n, 26412n, 9786n],
      "aä本☺",
      "aä本☺",
      "aä本☺",
      "aä本☺"
    ]);
  });

  test("supports append of strings into byte slices with spread", async () => {
    const result = await expectRuns(`
type Bytes []byte
type Buffer []byte
func (b *Buffer) writeString(s string) {
  *b = append(*b, s...)
}
b := []byte("go")
b = append(b, "jr"...)
named := Bytes("a")
named = append(named, "ä"...)
var nilBytes []byte
nilBytes = append(nilBytes, "Type"...)
buffer := Buffer("fmt")
buffer.writeString(" path")
return string(b), named, string(nilBytes), string(buffer)
`);

    expect(result.values).toEqual(["gojr", [97n, 195n, 164n], "Type", "fmt path"]);
  });

  test("supports append of strings into imported package byte-slice aliases", async () => {
    const graph = await evaluateSourcePackageGraph([{
      importPath: "example.com/textbuf",
      files: [{
        filename: "textbuf/buf.go",
        source: `package textbuf

type buffer []byte

func (b *buffer) writeString(s string) {
  *b = append(*b, s...)
}

func Build(s string) string {
  var b buffer
  b.writeString(s)
  return string(b)
}
`
      }]
    }]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import textbuf "example.com/textbuf"
return textbuf.Build("zygo")
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe("zygo");
  });

  test("supports typed nil slice conversions with Go len and cap", async () => {
    const result = await expectRuns(`
type Ints []int
s := []int(nil)
named := Ints(nil)
return len(s), cap(s), len(named), cap(named), fmt.Sprintf("%v", s), fmt.Sprintf("%#v", named)
`);

    expect(result.values).toEqual([0n, 0n, 0n, 0n, "<nil>", "Ints(nil)"]);
  });

  test("slices typed nil slices at zero bounds", async () => {
    const result = await expectRuns(`
var s []int
t := s[:]
u := s[:0:0]
return len(t), cap(t), len(u), cap(u), t == nil, u == nil
`);

    expect(result.values).toEqual([0n, 0n, 0n, 0n, true, true]);
  });

  test("keeps inferred var declaration types for later conversions", async () => {
    const result = await expectRuns(`
var bs = make([]uint8, 3)
bs[0] = 'x'
bs[1] = 'y'
bs[2] = 'z'
return string(bs)
`);

    expect(result.value).toBe("xyz");
  });

  test("supports conversions to named slice types with matching underlying type", async () => {
    const result = await expectRuns(`
type Bytes []uint8
var bs = make([]uint8, 2)
named := Bytes(bs)
named[0] = 'o'
named[1] = 'k'
return string(bs), string(named)
`);

    expect(result.values).toEqual(["ok", "ok"]);
  });

  test("slices share backing storage for index assignment and copy", async () => {
    const result = await expectRuns(`
xs := []int{1, 2, 3, 4}
ys := xs[1:3]
ys[0] = 20
n := copy(xs[2:], []int{30, 40})
return xs, ys[0], n
`);

    expect(result.values).toEqual([[1n, 20n, 30n, 40n], 20n, 2n]);
  });

  test("copy from strings writes UTF-8 bytes into byte slices", async () => {
    const result = await expectRuns(`
buf := make([]uint8, 4)
n := copy(buf, "abc")
return n, buf, string(buf[:3])
`);

    expect(result.values).toEqual([3n, [97n, 98n, 99n, 0n], "abc"]);
  });

  test("copy into a slice view mutates the backing slice", async () => {
    const result = await expectRuns(`
b := make([]byte, 2, 64)
v := b[0:]
n := copy(v, "ab")
return string(b), string(v), n
`);

    expect(result.values).toEqual(["ab", "ab", 2n]);
  });

  test("supports bitwise, shift, unary complement, and compound assignment operators", async () => {
    const result = await expectRuns(`
x := 0b1010
x |= 0b0101
y := x & 0b1100
y ^= 0b0010
y &^= 0b0100
y <<= 2
y >>= 1
s := "hi"
s += "!"
return x, y, ^0, s
`);

    expect(result.values).toEqual([15n, 20n, -1n, "hi!"]);
  });

  test("supports if init statements and Go-style builtins new delete clear copy min max print println", async () => {
    const result = await expectRuns(`
xs := []int{1, 2, 3}
ys := make([]int, 3)
n := copy(ys, xs)
clear(ys[1:3])
m := map[string]int{"a": 1, "b": 2}
delete(m, "a")
before := len(m)
clear(m)
p := new(int)
*p = max(3, min(7, 4))
if v := *p; v == 4 {
  print("v=", v)
  println(" ok")
}
return n, ys[0], ys[1], ys[2], before, len(m), *p
`);

    expect(result.output).toEqual(["v=", "4", " ok\n"]);
    expect(result.values).toEqual([3n, 1n, 0n, 0n, 1n, 0n, 4n]);
  });

  test("supports Go 1.26 new with expression arguments", async () => {
    const result = await expectRuns(`
p := new(123)
x := [2]int{123, 456}
q := new(x)
x[0] = 999
i := 0
next := func() int { i++; return i }
r := new(next())
b := new(i > 10)
return *p, (*q)[0], (*q)[1], *r, i, *b
`);

    expect(result.values).toEqual([123n, 123n, 456n, 1n, 1n, false]);
  });

  test("REPL typechecker treats star expressions as dereferences when the operand is a value", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("x := 42")).diagnostics).toEqual([]);
    expect((await session.evaluate("p := new(x)")).diagnostics).toEqual([]);
    let result = await session.evaluate("*p");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(42n);

    expect((await session.evaluate("arr := [2]int{1, 2}")).diagnostics).toEqual([]);
    expect((await session.evaluate("q := new(arr)")).diagnostics).toEqual([]);
    expect((await session.evaluate("arr[0] = 9")).diagnostics).toEqual([]);
    result = await session.evaluate("(*q)[0]");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("supports buffered channels, close, len cap, and receive ok values", async () => {
    const result = await expectRuns(`
ch := make(chan int, 2)
ch <- 7
ch <- 8
a := <-ch
b, ok := <-ch
close(ch)
c, ok2 := <-ch
return a, b, ok, c, ok2, len(ch), cap(ch)
`);

    expect(result.values).toEqual([7n, 8n, true, 0n, false, 0n, 2n]);
  });

  test("keeps goroutine channel sends parked across REPL evaluations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("go func() { c <- 1 }()")).diagnostics).toEqual([]);
    expect((await session.evaluate("a := <-c")).diagnostics).toEqual([]);

    const result = await session.evaluate("a");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("runs for range loops inside REPL goroutine function literals", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("go func() { for i := range 5 { c <- i } }()")).diagnostics).toEqual([]);

    for (const expected of [0n, 1n, 2n, 3n, 4n]) {
      const result = await session.evaluate("<-c");
      expect(result.diagnostics).toEqual([]);
      expect(result.value).toBe(expected);
    }
  });

  test("lets later REPL function declarations use earlier top-level variables", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("func f() { for i := range 5 { c <- i } }")).diagnostics).toEqual([]);
    expect((await session.evaluate("go f()")).diagnostics).toEqual([]);

    for (const expected of [0n, 1n, 2n, 3n, 4n]) {
      const result = await session.evaluate("<-c");
      expect(result.diagnostics).toEqual([]);
      expect(result.value).toBe(expected);
    }
  });

  test("compiled REPL functions observe later top-level variable value changes", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("x := 1")).diagnostics).toEqual([]);
    expect((await session.evaluate("func readX() int { return x }")).diagnostics).toEqual([]);

    let result = await session.evaluate("readX()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);

    expect((await session.evaluate("x = 7")).diagnostics).toEqual([]);
    result = await session.evaluate("readX()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(7n);
  });

  test("compatible REPL function redefinition updates existing callers through the function slot", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("func f() int { return 1 }")).diagnostics).toEqual([]);
    expect((await session.evaluate("func g() int { return f() }")).diagnostics).toEqual([]);

    let result = await session.evaluate("g()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);

    expect((await session.evaluate("func f() int { return 2 }")).diagnostics).toEqual([]);
    result = await session.evaluate("g()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("REPL function redefinition may replace an existing binding with a different signature", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("func f() int { return 1 }")).diagnostics).toEqual([]);
    expect((await session.evaluate("func f(x int) int { return x }")).diagnostics).toEqual([]);

    const result = await session.evaluate("f(7)");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(7n);
  });

  test("failed REPL function definitions leave no accepted type or runtime binding", async () => {
    const session = new GoJuniorSession();

    const bad = await session.evaluate("func f(a, b int) { return a * b }");
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(bad.diagnostics[0]?.message).toContain("too many return values");

    const missing = await session.evaluate("f");
    expect(missing.diagnostics).toHaveLength(1);
    expect(missing.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(missing.diagnostics[0]?.message).toContain("undefined: f");

    expect((await session.evaluate("func f(a, b int) int { return a * b }")).diagnostics).toEqual([]);

    const result = await session.evaluate("f(6, 7)");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(42n);
  });

  test("failed REPL function replacement keeps the old accepted binding", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("func f() int { return 1 }")).diagnostics).toEqual([]);

    const bad = await session.evaluate("func f(x int) { return x }");
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(bad.diagnostics[0]?.message).toContain("too many return values");

    const result = await session.evaluate("f()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("REPL short declarations may replace an existing top-level variable", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate(`a := "hi"`)).diagnostics).toEqual([]);

    const result = await session.evaluate("a");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("hi");
  });

  test("REPL var declarations may replace an existing top-level variable with a different type", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("var v string")).diagnostics).toEqual([]);
    expect((await session.evaluate("var v int")).diagnostics).toEqual([]);

    const result = await session.evaluate("v");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(0n);
  });

  test("REPL type errors render source line context in shared host diagnostics", async () => {
    const session = new GoJuniorSession();
    const source = `func f() { a := 10; var a string = "hi"; println(a) }`;

    const result = await session.evaluate(source);
    expect(result.diagnostics).toHaveLength(2);
    expect(result.diagnostics[0]?.sourceLine).toBe(source);
    expect(result.diagnostics[1]?.sourceLine).toBe(source);
    expect(result.diagnostics[0]?.span?.column).toBe(25);
    expect(result.diagnostics[1]?.span?.column).toBe(12);

    const payload = evaluationResultToHostPayload(result);
    expect(payload.diagnostics[0]).toContain(`a redeclared in this block\n${source}\n`);
    expect(payload.diagnostics[0]).toContain(`${" ".repeat(24)}^`);
    expect(payload.diagnostics[1]).toContain(`other declaration of a\n${source}\n`);
    expect(payload.diagnostics[1]).toContain(`${" ".repeat(11)}^`);
  });

  test("failed REPL runtime execution rolls back top-level declarations", async () => {
    const session = new GoJuniorSession();

    const failed = await session.evaluate(`a := 10; panic("boom")`);
    expect(failed.diagnostics).toHaveLength(1);
    expect(failed.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(failed.diagnostics[0]?.message).toContain("boom");

    const missing = await session.evaluate("a");
    expect(missing.diagnostics).toHaveLength(1);
    expect(missing.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(missing.diagnostics[0]?.message).toContain("undefined: a");
  });

  test("reports REPL deadlocks and keeps the session usable without zombie receives", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("r := 0")).diagnostics).toEqual([]);

    let result = await session.evaluate("r = <-c");
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(result.diagnostics[0]?.message).toContain("receive from channel would block");

    expect((await session.evaluate("go func() { c <- 2 }()")).diagnostics).toEqual([]);

    result = await session.evaluate("r");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(0n);

    expect((await session.evaluate("r = <-c")).diagnostics).toEqual([]);
    result = await session.evaluate("r");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("supports deterministic select choice from the configured runtime seed", async () => {
    const source = `
ch1 := make(chan int, 2)
ch2 := make(chan int, 2)
ch1 <- 1
ch1 <- 3
ch2 <- 2
ch2 <- 4
a := 0
b := 0
select {
case a = <-ch1:
case a = <-ch2:
}
select {
case b = <-ch1:
case b = <-ch2:
}
return a, b
`;

    const first = await expectRuns(source, { randomSeed: "gojr-select-seed" });
    const second = await expectRuns(source, { randomSeed: "gojr-select-seed" });
    const other = await expectRuns(source, { randomSeed: "gojr-other-select-seed" });

    expect(first.values).toEqual(second.values);
    expect(first.values).toEqual([2n, 1n]);
    expect(other.values).toEqual([1n, 2n]);
  });

  test("supports select default and reports would-block channel operations", async () => {
    const result = await expectRuns(`
ch := make(chan int, 1)
out := 3
select {
case out = <-ch:
default:
  out = 9
}
return out
`);
    expect(result.value).toBe(9n);

    const send = await evaluateSource(`
ch := make(chan int)
ch <- 1
`);
    expect(send.diagnostics).toHaveLength(1);
    expect(send.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(send.diagnostics[0]?.message).toContain("send on channel would block");

    const receive = await evaluateSource(`
ch := make(chan int)
return <-ch
`);
    expect(receive.diagnostics).toHaveLength(1);
    expect(receive.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(receive.diagnostics[0]?.message).toContain("receive from channel would block");
  });

  test("supports nil channel blocking and disabled nil-channel select cases", async () => {
    const selectedDefault = await expectRuns(`
var ch chan int
side := 0
next := func() int {
  side++
  return 1
}
select {
case ch <- next():
  side = 100
case v := <-ch:
  side = 200 + v
default:
  side += 10
}
return side
`);
    expect(selectedDefault.value).toBe(11n);

    const send = await evaluateSource(`
var ch chan int
ch <- 1
`);
    expect(send.diagnostics).toHaveLength(1);
    expect(send.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(send.diagnostics[0]?.message).toContain("send on nil channel would block");

    const receive = await evaluateSource(`
var ch chan int
return <-ch
`);
    expect(receive.diagnostics).toHaveLength(1);
    expect(receive.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(receive.diagnostics[0]?.message).toContain("receive from nil channel would block");

    const blockedSelect = await evaluateSource(`
var ch chan int
select {
case <-ch:
}
`);
    expect(blockedSelect.diagnostics).toHaveLength(1);
    expect(blockedSelect.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(blockedSelect.diagnostics[0]?.message).toContain("select would block");

    const closeNil = await evaluateSource(`
var ch chan int
close(ch)
`);
    expect(closeNil.diagnostics).toHaveLength(1);
    expect(closeNil.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(closeNil.diagnostics[0]?.message).toContain("close of nil channel");
  });

  test("supports methods on named non-struct types and two-value type assertions", async () => {
    const result = await expectRuns(`
type Duration int

func (d Duration) Double() Duration {
  return d + d
}

var x interface{} = Duration(5)
v, ok := x.(Duration)
bad, badOK := x.(string)
return ok, v.Double(), bad, badOK
`);

  expect(result.values).toEqual([true, 10n, "", false]);
  });

  test("supports erased generic functions and generic struct declarations", async () => {
    const result = await expectRuns(`
type Box[T any] struct {
  Value T
}

type Number interface {
  ~int | ~float64
}

func Identity[T any](value T) T {
  var copy T = value
  return copy
}

func Pick[T Number](value T) T {
  return value
}

func Add[T Number](left, right T) T {
  return left + right
}

func Unbox[T any](box Box[T]) T {
  return box.Value
}

a := Identity[int](42)
b := Identity[string]("hi")
c := Unbox[int](Box[int]{Value: 7})
d := Pick[float64](2.5)
e := Add[int](3, 4)
f := Add[float64](1.25, 2.5)
return a, b, c, d, e, f
`);

    expect(result.values).toEqual([42n, "hi", 7n, 2.5, 7n, 3.75]);
  });

  test("supports erased generic functions with multiple type arguments", async () => {
    const result = await expectRuns(`
type Pair[A, B any] struct {
  A A
  B B
}

func Make[A, B any](a A, b B) Pair[A, B] {
  return Pair[A, B]{A: a, B: b}
}

p := Make[int, string](7, "seven")
return p.A, p.B
`);

    expect(result.values).toEqual([7n, "seven"]);
  });

  test("binds generic type parameters for new and zero values", async () => {
    const result = await expectRuns(`
func Ptr[T any](value T) *T {
  p := new(T)
  *p = value
  return p
}

func Zero[T any]() T {
  var zero T
  return zero
}

i := Ptr[int](42)
s := Ptr[string]("hi")
return *i, *s, Zero[int]()
`);

    expect(result.values).toEqual([42n, "hi", 0n]);
  });

  test("infers generic type parameters from function argument signatures", async () => {
    const result = await expectRuns(`
func do[T any](fn func() (T, error)) (T, error) {
  return fn()
}

func noerr() (int, error) {
  return 7, nil
}

v, err := do(noerr)
return v, err == nil
`);

    expect(result.values).toEqual([7n, true]);
  });

  test("supports methods on erased generic receiver types", async () => {
    const result = await expectRuns(`
type Box[T any] struct {
  value T
}

func (b *Box[T]) Set(value T) {
  b.value = value
}

func (b *Box[T]) Get() T {
  return b.value
}

var box Box[int]
box.Set(9)
return box.Get()
`);

    expect(result.value).toBe(9n);
  });

  test("instantiates generic receiver method parameters from imported named receiver types", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/atomic",
        files: [{
          filename: "/workspace/atomic/atomic.go",
          source: `package atomic

type Pointer[T any] struct {
  v *T
}

func (x *Pointer[T]) Load() *T {
  return x.v
}

func (x *Pointer[T]) CompareAndSwap(old, new *T) bool {
  if x.v == old {
    x.v = new
    return true
  }
  return false
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "example.com/atomic"

type dirInfo struct {
  dir int
}

type file struct {
  dirinfo atomic.Pointer[dirInfo]
}

var F file
var Swapped bool
var Dir int

func init() {
  next := &dirInfo{dir: 7}
  Swapped = F.dirinfo.CompareAndSwap(nil, next)
  Dir = F.dirinfo.Load().dir
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Swapped).toBe(true);
    expect(graph.packages.app?.Dir).toBe(7n);
  });

  test("compares named unsafe typed nils in atomic pointer compare-and-swap", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

import "unsafe"

type Pointer[T any] struct {
  v unsafe.Pointer
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) bool

func (x *Pointer[T]) Load() *T {
  return (*T)(LoadPointer(&x.v))
}

func (x *Pointer[T]) CompareAndSwap(old, new *T) bool {
  return CompareAndSwapPointer(&x.v, unsafe.Pointer(old), unsafe.Pointer(new))
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "sync/atomic"

type D struct {
  v int
}

var P atomic.Pointer[D]
var OK bool
var Seen int

func init() {
  d := &D{v: 9}
  OK = P.CompareAndSwap(nil, d)
  Seen = P.Load().v
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.OK).toBe(true);
    expect(graph.packages.app?.Seen).toBe(9n);
  });

  test("keeps caller package identity for generic atomic pointer method results", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

import "unsafe"

type Pointer[T any] struct {
  v unsafe.Pointer
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer

func (x *Pointer[T]) Load() *T {
  return (*T)(LoadPointer(&x.v))
}

func (x *Pointer[T]) Store(val *T) {
  StorePointer(&x.v, unsafe.Pointer(val))
}

func (x *Pointer[T]) Swap(new *T) *T {
  return (*T)(SwapPointer(&x.v, unsafe.Pointer(new)))
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "sync/atomic"

type dirInfo struct {
  dir uintptr
}

func (d *dirInfo) close() {
  Closed = true
}

type file struct {
  dirinfo atomic.Pointer[dirInfo]
}

var F file
var Closed bool

func init() {
  F.dirinfo.Store(&dirInfo{dir: 7})
  old := F.dirinfo.Swap(nil)
  old.close()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Closed).toBe(true);
  });

  test("preserves caller-owned struct values through imported generic receiver methods", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "box",
        files: [{
          filename: "/workspace/box/box.go",
          source: `package box

type Box[T any] struct{}

func (b *Box[T]) Id(x T) T {
  return x
}
`
        }]
      },
      {
        importPath: "g",
        files: [{
          filename: "/workspace/g/g.go",
          source: `package g

import "box"

type value struct {
  text string
}

var b box.Box[value]

func Run() string {
  v := b.Id(value{text:"ok"})
  return v.text
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "g"

var Seen string

func init() {
  Seen = g.Run()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe("ok");
  });

  test("preserves unsafe pointer payloads for imported generic atomic pointers", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

import "unsafe"

type Pointer[T any] struct {
  _ [0]*T
  v unsafe.Pointer
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)

func (x *Pointer[T]) Load() *T {
  return (*T)(LoadPointer(&x.v))
}

func (x *Pointer[T]) Store(val *T) {
  StorePointer(&x.v, unsafe.Pointer(val))
}
`
        }]
      },
      {
        importPath: "g",
        files: [{
          filename: "/workspace/g/g.go",
          source: `package g

import "sync/atomic"

type setting struct {
  value atomic.Pointer[value]
}

type value struct {
  text string
}

var empty = value{text:"ok"}

func Run() string {
  s := new(setting)
  s.value.Store(&empty)
  return (*s.value.Load()).text
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "g"

var Seen string

func init() {
  Seen = g.Run()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe("ok");
  });

  test("preserves caller package identity through imported generic atomic pointer loads", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

import "unsafe"

type Pointer[T any] struct {
  _ [0]*T
  v unsafe.Pointer
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)

func (x *Pointer[T]) Load() *T {
  return (*T)(LoadPointer(&x.v))
}

func (x *Pointer[T]) Store(val *T) {
  StorePointer(&x.v, unsafe.Pointer(val))
}
`
        }]
      },
      {
        importPath: "internal/sync",
        files: [{
          filename: "/usr/local/go/src/internal/sync/hashtriemap.go",
          source: `package sync

import "sync/atomic"

type Mutex struct {
  state int32
  sema uint32
}

func (m *Mutex) Lock() {
  m.state = 1
}

type node[K comparable, V any] struct {
  isEntry bool
}

type indirect[K comparable, V any] struct {
  node[K, V]
  mu atomiclessMutex
}

type atomiclessMutex = Mutex

type HashTrieMap[K comparable, V any] struct {
  root atomic.Pointer[indirect[K, V]]
}

func newIndirectNode[K comparable, V any]() *indirect[K, V] {
  return &indirect[K, V]{node: node[K, V]{isEntry: false}}
}

func (h *HashTrieMap[K, V]) Init() {
  h.root.Store(newIndirectNode[K, V]())
}

func (h *HashTrieMap[K, V]) Use() int32 {
  h.Init()
  i := h.root.Load()
  i.mu.Lock()
  return i.mu.state
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import isync "internal/sync"

var Seen int32

func init() {
  var h isync.HashTrieMap[int, string]
  Seen = h.Use()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe(1n);
  });

  test("initializes imported generic hashtrie zero values with instantiated array fields", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sync/atomic",
        files: [{
          filename: "/usr/local/go/src/sync/atomic/type.go",
          source: `package atomic

import "unsafe"

type Pointer[T any] struct {
  _ [0]*T
  v unsafe.Pointer
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)

func (x *Pointer[T]) Load() *T {
  return (*T)(LoadPointer(&x.v))
}

func (x *Pointer[T]) Store(val *T) {
  StorePointer(&x.v, unsafe.Pointer(val))
}
`
        }]
      },
      {
        importPath: "internal/sync",
        files: [{
          filename: "/usr/local/go/src/internal/sync/hashtriemap.go",
          source: `package sync

import "sync/atomic"

const (
  nChildrenLog2 = 4
  nChildren = 1 << nChildrenLog2
)

type node[K comparable, V any] struct {
  isEntry bool
}

type indirect[K comparable, V any] struct {
  node[K, V]
  children [nChildren]atomic.Pointer[node[K, V]]
}

type HashTrieMap[K comparable, V any] struct {
  root atomic.Pointer[indirect[K, V]]
}

func newIndirectNode[K comparable, V any]() *indirect[K, V] {
  return &indirect[K, V]{node: node[K, V]{isEntry: false}}
}

func (h *HashTrieMap[K, V]) Init() {
  h.root.Store(newIndirectNode[K, V]())
}

func (h *HashTrieMap[K, V]) EmptySlot() bool {
  h.Init()
  i := h.root.Load()
  n := i.children[0].Load()
  return n == nil
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import isync "internal/sync"

var Seen bool

func init() {
  var h isync.HashTrieMap[int, string]
  Seen = h.EmptySlot()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe(true);
  });

  test("preserves interface dynamic package identity across imported calls", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "box",
        files: [{
          filename: "/workspace/box/box.go",
          source: `package box

func RoundTrip(v any) any {
  return v
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "box"

type hidden struct {
  value int
}

var Seen int

func init() {
  h := &hidden{value: 7}
  Seen = box.RoundTrip(h).(*hidden).value
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe(7n);
  });

  test("passes caller concrete methods through imported interface parameters", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sortlike",
        files: [{
          filename: "/workspace/sortlike/sortlike.go",
          source: `package sortlike

type Interface interface {
  Len() int
}

func Sort(data Interface) int {
  return data.Len()
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "sortlike"

type fastpathAslice struct{}

func (fastpathAslice) Len() int {
  return 56
}

var Seen = sortlike.Sort(fastpathAslice{})
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe(56n);
  });

  test("REPL checker accepts keyed generic struct literals", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("type Box[T any] struct { Value T }")).diagnostics).toEqual([]);
    expect((await session.evaluate("func Unbox[T any](box Box[T]) T { return box.Value }")).diagnostics).toEqual([]);

    const result = await session.evaluate("Unbox[int](Box[int]{Value: 7})");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(7n);
  });

  test("evaluates generic anonymous structs with function and result-typed fields", async () => {
    const result = await expectRuns(`
func OnceValue[T any](f func() T) func() T {
  d := struct {
    f      func() T
    result T
  }{
    f: f,
  }
  return func() T {
    d.result = d.f()
    return d.result
  }
}

f := OnceValue[int](func() int { return 7 })
return f()
`);
    expect(result.value).toBe(7n);
  });

  test("ranges typed nil slices as zero iterations inside once closures", async () => {
    const result = await expectRuns(`
var envs []string
var env map[string]int

onceFunc := func(f func()) func() {
  done := false
  return func() {
    if !done {
      done = true
      f()
    }
  }
}

copyenv := onceFunc(func() {
  env = make(map[string]int)
  for i, s := range envs {
    env[s] = i
  }
})
copyenv()
_, ok := env["x"]
return ok, len(env)
`);
    expect(result.values).toEqual([false, 0n]);
  });

  test("supports address-of composite literals and three-index slicing capacity", async () => {
    const result = await expectRuns(`
type Point struct { X int }
p := &Point{X: 7}
xs := make([]int, 5, 8)
ys := xs[1:3:4]
return p.X, len(ys), cap(ys)
`);

    expect(result.values).toEqual([7n, 2n, 3n]);
  });

  test("enforces comparable map keys and supports array and struct comparability", async () => {
    const ok = await expectRuns(`
type Point struct { X int; Y string }
a := [2]int{1, 2}
b := [2]int{1, 2}
p := Point{X: 1, Y: "a"}
q := Point{X: 1, Y: "a"}
return a == b, p == q
`);
    expect(ok.values).toEqual([true, true]);

    const bad = await evaluateSource(`
m := map[[]int]int{}
_ = m
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("supports anonymous empty struct zero values, literals, and array range assignment", async () => {
    const result = await expectRuns(`
var xs [3]struct{}
for i := range xs {
  xs[i] = struct{}{}
}
return xs[0] == struct{}{}, xs == [3]struct{}{}
`);

    expect(result.values).toEqual([true, true]);
  });

  test("supports non-empty anonymous struct zero values and pointer selectors", async () => {
    const result = await expectRuns(`
type x2 struct { a, b, c int; d int }
var g1 x2
var g2 struct { a, b, c int; d x2 }

s1 := &g1
s2 := &g2
s1.a = 1
s1.b = 2
s1.c = 3
s1.d = 5
s2.a = 7
s2.b = 11
s2.c = 13
s2.d.a = 17
s2.d.b = 19
s2.d.c = 23
s2.d.d = 20
return s2.d.c, g2.d.c, s1.a + s1.b + s1.c + s1.d + s2.a + s2.b + s2.c + s2.d.a + s2.d.b + s2.d.c + s2.d.d
`);

    expect(result.values).toEqual([23n, 23n, 121n]);
  });

  test("supports blank imports, dot imports, init functions, and range over integers and iterator functions", async () => {
    const result = await expectRuns(`
import . "fmt"
import _ "fmt"

var total int

func init() {
  total = 2
}

for i := range 4 {
  total += i
}

iter := func(yield func(int, string) bool) {
  if !yield(10, "a") {
    return
  }
  yield(20, "b")
}

out := ""
for k, v := range iter {
  out += Sprintf("%v:%v;", k, v)
}

xs := [2]int{}
q := 0
for xs[func() int {
  q++
  return 0
}()] = range [2]int{} {
}

yieldOnce := func(yield func(int) bool) {
  yield(1)
}

for _ = range yieldOnce {
  total++
}

neg := 0
for i := range -1 {
  neg += i
  total += 1000
}

runeCount := 0
var last rune
for i := range 'a' {
  var _ *rune = &i
  last = i
  runeCount++
}

return total, out, q, xs[0], neg, runeCount, last
`);

    expect(result.values).toEqual([9n, "10:a;20:b;", 2n, 1n, 0n, 97n, 96n]);
  });

  test("supports standard iter.Pull and iter.Pull2", async () => {
    const result = await expectRuns(`
import "iter"

seq := func(yield func(int) bool) {
  if !yield(7) {
    return
  }
  yield(11)
}

next, stop := iter.Pull[int](seq)
defer stop()
a, okA := next()
b, okB := next()
c, okC := next()

seq2 := func(yield func(string, int) bool) {
  if !yield("a", 1) {
    return
  }
  yield("b", 2)
}

next2, stop2 := iter.Pull2[string, int](seq2)
defer stop2()
k1, v1, ok1 := next2()
k2, v2, ok2 := next2()
k3, v3, ok3 := next2()

return a, okA, b, okB, c, okC, k1, v1, ok1, k2, v2, ok2, k3, v3, ok3
`);

    expect(result.values).toEqual([
      7n, true, 11n, true, 0n, false,
      "a", 1n, true, "b", 2n, true, "", 0n, false
    ]);
  });

  test("supports stopping standard iter.Pull producers early", async () => {
    const result = await expectRuns(`
import "iter"

count := 0
seq := func(yield func(int) bool) {
  count++
  if !yield(1) {
    return
  }
  count += 10
}

next, stop := iter.Pull[int](seq)
a, ok := next()
stop()
b, ok2 := next()
return a, ok, b, ok2, count
`);

    expect(result.values).toEqual([1n, true, 0n, false, 1n]);
  });

  test("keeps caller assignment scope across iter.Pull function fields", async () => {
    const result = await expectRuns(`
import "iter"

type parser struct {
  next func() (int, bool)
  stop func()
}

func seq(yield func(int) bool) {
  yield(3)
  yield(5)
}

func (p *parser) run() int {
  if p.next == nil {
    p.next, p.stop = iter.Pull[int](seq)
  }
  defer p.stop()
  var reply int
  ok := true
  for ok {
    reply, ok = p.next()
  }
  return reply
}

p := &parser{}
return p.run()
`);

    expect(result.value).toBe(0n);
  });

  test("keeps iter.Pull pointer results ordered with ok through function fields", async () => {
    const result = await expectRuns(`
import "iter"

type reply struct {
  n int
}

type parser struct {
  next func() (*reply, bool)
  stop func()
}

func seq(yield func(*reply) bool) {
  yield(&reply{n: 2595})
}

func (p *parser) run() (int, bool, bool) {
  if p.next == nil {
    p.next, p.stop = iter.Pull[*reply](seq)
  }
  defer p.stop()
  got, ok := p.next()
  empty, ok2 := p.next()
  return got.n, ok, empty == nil && !ok2
}

p := &parser{}
a, b, c := p.run()
return a, b, c
`);

    expect(result.values).toEqual([2595n, true, true]);
  });

  test("converts ordinary functions to named function types", async () => {
    const result = await expectRuns(`
type completer func(line string, pos int) (head string, completions []string, tail string)

func complete(line string, pos int) (string, []string, string) {
  return line, []string{"a", "b"}, line[pos:]
}

c := completer(complete)
head, choices, tail := c("hello", 2)
return head, choices[1], tail
`);

    expect(result.values).toEqual(["hello", "b", "llo"]);
  });

  test("supports keyed literals for embedded instantiated generic fields", async () => {
    const result = await expectRuns(`
type node[K comparable, V any] struct {
  isEntry bool
}

type indirect[K comparable, V any] struct {
  node[K, V]
  parent *indirect[K, V]
}

func newIndirect[K comparable, V any](parent *indirect[K, V]) *indirect[K, V] {
  return &indirect[K, V]{node: node[K, V]{isEntry: false}, parent: parent}
}

x := newIndirect[int, string](nil)
return x.node.isEntry, x.parent == nil
`);
    expect(result.values).toEqual([false, true]);
  });

  test("promotes methods from embedded instantiated pointer fields after any round trips", async () => {
    const result = await expectRuns(`
type canonMap[T comparable] struct {
  value T
}

func (m *canonMap[T]) Load(key T) *T {
  return &m.value
}

type uniqueMap[T comparable] struct {
  *canonMap[T]
}

func round(v any) any {
  return v
}

u := &uniqueMap[int]{canonMap: &canonMap[int]{value: 7}}
m := round(u).(*uniqueMap[int])
return *m.Load(1)
`);
    expect(result.value).toBe(7n);
  });

  test("keeps standard iter intrinsic ahead of cached package payloads", async () => {
    const result = await expectRuns(`
import "iter"

seq := func(yield func(int) bool) {
  yield(3)
}

next, stop := iter.Pull[int](seq)
defer stop()
value, ok := next()
return value, ok
`, {
      packages: {
        iter: {
          Pull: 123n
        }
      }
    });

    expect(result.values).toEqual([3n, true]);
  });

  test("supports embedded fields, promoted methods, interface embedding, and struct tags", async () => {
    const result = await expectRuns(`
type Inner struct {
  X int
}

func (i Inner) Double() int {
  return i.X * 2
}

type Doubler interface {
  Double() int
}

type NamedDoubler interface {
  Doubler
}

type Outer struct {
  Inner \`json:"inner"\`
  Name string \`json:"name"\`
}

o := Outer{Inner: Inner{X: 3}, Name: "n"}
o.X = 4
var d NamedDoubler
d = o
return o.X, o.Double(), d.Double()
`);

    expect(result.values).toEqual([4n, 8n, 8n]);
  });

  test("assigns promoted fields through imported embedded struct zero values", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/codec",
        files: [{
          filename: "/workspace/codec/codec.go",
          source: `package codec

type DecodeOptions struct {
  MapType int
}

type BasicHandle struct {
  DecodeOptions
}

type MsgpackHandle struct {
  BasicHandle
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import codec "example.com/codec"

var mh codec.MsgpackHandle
mh.MapType = 7
return mh.MapType
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(7n);
  });

  test("assigns promoted fields through package-info-only imported embedded struct zero values", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "example.com/codec",
        files: [{
          filename: "/workspace/codec/codec.go",
          source: `package codec

type DecodeOptions struct {
  MapType int
}

type BasicHandle struct {
  DecodeOptions
}

type MsgpackHandle struct {
  BasicHandle
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import codec "example.com/codec"

var mh codec.MsgpackHandle
mh.MapType = 7
return mh.MapType
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });

    expect(result.value).toBe(7n);
  });

  test("assigns local named slices to package-info-only imported interfaces", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sort",
        files: [{
          filename: "/workspace/sort/sort.go",
          source: `package sort

type Interface interface {
  Len() int
  Less(i, j int) bool
  Swap(i, j int)
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "sort"

type Sorter []int

func (s Sorter) Len() int { return len(s) }
func (s Sorter) Less(i, j int) bool { return s[i] < s[j] }
func (s Sorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

var s Sorter
var x sort.Interface = s
return x != nil
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });

    expect(result.value).toBe(true);
  });

  test("preserves converted named slice method sets across interface package calls", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sort",
        files: [{
          filename: "/workspace/sort/sort.go",
          source: `package sort

type Interface interface {
  Len() int
  Less(i, j int) bool
  Swap(i, j int)
}

func Sort(data Interface) int {
  return data.Len()
}
`
        }]
      },
      {
        importPath: "app",
        files: [{
          filename: "/workspace/app/app.go",
          source: `package app

import "sort"

type Sorter []int

func (s Sorter) Len() int { return len(s) }
func (s Sorter) Less(i, j int) bool { return s[i] < s[j] }
func (s Sorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

var Seen = sort.Sort(Sorter([]int{1, 2, 3}))
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);
    expect(graph.packages.app?.Seen).toBe(3n);
  });

  test("preserves REPL-local converted named slice method sets across interface package calls", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "sort",
        files: [{
          filename: "/workspace/sort/sort.go",
          source: `package sort

type Interface interface {
  Len() int
  Less(i, j int) bool
  Swap(i, j int)
}

func Sort(data Interface) int {
  return data.Len()
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "sort"

type Sorter []int

func (s Sorter) Len() int { return len(s) }
func (s Sorter) Less(i, j int) bool { return s[i] < s[j] }
func (s Sorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

return sort.Sort(Sorter([]int{1, 2, 3}))
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos,
      packageRuntimes: graph.packageRuntimes
    });

    expect(result.value).toBe(3n);
  });

  test("assigns imported pointer receiver values to package-info-only imported interfaces", async () => {
    const graph = await evaluateSourcePackageGraph([
      {
        importPath: "io",
        files: [{
          filename: "/workspace/io/io.go",
          source: `package io

type Writer interface {
  Write(p []byte) (n int, err error)
}
`
        }]
      },
      {
        importPath: "os",
        files: [{
          filename: "/workspace/os/os.go",
          source: `package os

type File struct{}

var Stdout = &File{}

func (f *File) Write(p []byte) (n int, err error) {
  return len(p), nil
}
`
        }]
      }
    ]);

    expect(graph.diagnostics).toEqual([]);

    const result = await expectRuns(`
import "io"
import "os"

var w io.Writer = os.Stdout
return w != nil
`, {
      packages: graph.packages,
      packageInfos: graph.packageInfos
    });

    expect(result.value).toBe(true);
  });

  test("prefers direct methods over deeper promoted fields", async () => {
    const result = await expectRuns(`
type Tree struct {
  Name string
}

type Template struct {
  *Tree
}

func (t *Template) Name() string {
  return "method:" + t.Tree.Name
}

t := &Template{Tree: &Tree{Name: "field"}}
return t.Name()
`);

    expect(result.value).toBe("method:field");
  });

  test("promotes methods from embedded interface fields when satisfying broader interfaces", async () => {
    const result = await expectRuns(`
type Reader interface {
  Read([]byte) (int, error)
}

type Closer interface {
  Close() error
}

type ReadCloser interface {
  Reader
  Closer
}

type nopCloser struct {
  Reader
}

func (nopCloser) Close() error {
  return nil
}

type bytesReader struct{}

func (bytesReader) Read([]byte) (int, error) {
  return 7, nil
}

func makeReadCloser(r Reader) ReadCloser {
  return nopCloser{r}
}

rc := makeReadCloser(bytesReader{})
n, readErr := rc.Read(nil)
closeErr := rc.Close()
var nilReader Reader
nilRC := makeReadCloser(nilReader)
nilCloseErr := nilRC.Close()
return n, readErr == nil, closeErr == nil, nilCloseErr == nil
`);

    expect(result.values).toEqual([7n, true, true, true]);
  });

  test("supports pointer receivers on named scalar types", async () => {
    const result = await expectRuns(`
type Counter int

func (c *Counter) Inc() {
  *c = *c + 1
}

var c Counter
c.Inc()
c.Inc()
return c
`);

    expect(result.value).toBe(2n);
  });

  test("supports switch init statements and short redeclarations", async () => {
    const result = await expectRuns(`
x := 1
x, y := 2, 3
out := 0
switch z := x + y; z {
case 5:
  out = z
default:
  out = 99
}
return x, y, out
`);

    expect(result.values).toEqual([2n, 3n, 5n]);
  });

  test("scopes statement init and clause short declarations like Go", async () => {
    const result = await expectRuns(`
total := 0
for i := 0; i < 2; i++ {
  total += i
}
for i := 0; i < 2; i++ {
  total += i * 10
}

if _, ok := map[int]int{}[1]; !ok {
  total += 100
}
if _, ok := map[int]int{1: 1}[1]; ok {
  total += 1000
}

switch x := 1; x {
case 1:
  y := 2
  total += y
case 2:
  y := 3
  total += y
}

ch := make(chan int)
select {
case v := <-ch:
  total += v
default:
  v := 7
  total += v
}
select {
case v := <-ch:
  total += v
default:
  v := 8
  total += v
}

return total
`);

    expect(result.value).toBe(1128n);
  });

  test("rejects invalid short declarations, for posts, fallthrough, and gotos over variables", async () => {
    const repeatedSite = await evaluateSource(`
i := 0
sum := 0
Loop:
inst := i
sum = sum + inst
i++
if i < 3 {
	goto Loop
}
sum
`);
    expect(repeatedSite.diagnostics).toEqual([]);
    expect(repeatedSite.value).toBe(3n);

    const noNew = await evaluateSource(`
x := 1
x := 2
`);
    expect(noNew.diagnostics).toHaveLength(1);
    expect(noNew.diagnostics[0]?.message).toContain("short declaration has no new variables");

    const badPost = await evaluateSource(`
for i := 0; i < 2; i := i + 1 {
}
`);
    expect(badPost.diagnostics).toHaveLength(1);
    expect(badPost.diagnostics[0]?.message).toContain("for post");

    const badFallthroughMiddle = await evaluateSource(`
switch 1 {
case 1:
  fallthrough
  fmt.Printf("nope")
default:
}
`);
    expect(badFallthroughMiddle.diagnostics).toHaveLength(1);
    expect(badFallthroughMiddle.diagnostics[0]?.message).toContain("fallthrough must be the final statement");

    const badFallthroughFinal = await evaluateSource(`
switch 1 {
case 1:
  fallthrough
}
`);
    expect(badFallthroughFinal.diagnostics).toHaveLength(1);
    expect(badFallthroughFinal.diagnostics[0]?.message).toContain("final switch clause");

    const badGoto = await evaluateSource(`
goto Done
x := 1
Done:
return x
`);
    expect(badGoto.diagnostics).toHaveLength(1);
    expect(badGoto.diagnostics[0]?.message).toContain("jumps over variable declaration");
  });

  test("decodes Go string and rune escapes", async () => {
    const result = await expectRuns(`
s := "\\x41\\101\\u0042\\U00000043"
r := '\\n'
return s, r
`);

    expect(result.values).toEqual(["AABC", 10n]);
  });

  test("enforces recursive map-key comparability for arrays and structs", async () => {
    const ok = await expectRuns(`
type Key struct { A [2]int; B string }
m := map[Key]int{}
m[Key{A: [2]int{1, 2}, B: "x"}] = 7
return m[Key{A: [2]int{1, 2}, B: "x"}]
`);
    expect(ok.value).toBe(7n);

    const badStruct = await evaluateSource(`
type Bad struct { A []int }
_ = map[Bad]int{}
`);
    expect(badStruct.diagnostics).toHaveLength(1);
    expect(badStruct.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("reports unresolved goto labels", async () => {
    const result = await evaluateSource(`
goto Missing
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("unresolved goto label Missing");
  });

  test("reports loop-limit diagnostics at the for statement span", async () => {
    const result = await evaluateSource(`
x := 1
for {
  x++
}
`, { filename: "loop.go", maxLoopIterations: 2 });

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.filename).toBe("loop.go");
    expect(result.diagnostics[0]?.span?.filename).toBe("loop.go");
    expect(result.diagnostics[0]?.span?.line).toBe(3);
    expect(result.diagnostics[0]?.message).toContain("loop exceeded 2 iterations");
  });
});
