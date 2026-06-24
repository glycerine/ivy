import { describe, expect, test } from "./testHarness.js";
import { checkGoJuniorSource } from "../src/typecheck.js";
import {
  Array as GoTypesArray,
  Named as GoTypesNamed,
  NewMethodSet,
  NewPointer,
  Slice as GoTypesSlice,
  Var as GoTypesVar,
  type GoJuniorSheetNamespace
} from "../src/go/types/index.js";
import type { GoJuniorCheckConfig } from "../src/typecheck.js";

const TEST_FILENAME = "front-checker-test.go";

function check(source: string, config: GoJuniorCheckConfig = {}) {
  return checkGoJuniorSource(source, TEST_FILENAME, config);
}

describe("Go-junior TypeScript go/types checker", () => {
  test("imports cmp through the static checker", () => {
    const result = check(`import "cmp"\n`);

    expect(result.diagnostics).toEqual([]);
  });

  test("imports runtime error and cleanup APIs through the static checker", () => {
    const result = check(`
package runtimeuser

import "runtime"

func Use(e any) bool {
  _, ok := e.(runtime.Error)
  value := ""
  runtime.AddCleanup(&value, func(name string) {}, value).Stop()
  return ok
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("records named-type constant declarations", () => {
    const result = check(`
package constants

type MyInt int

const X MyInt = 1
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("X")?.Type()?.String()).toBe("MyInt");
  });

  test("scopes generic type parameters for functions and type declarations", () => {
    const result = check(`
package generic

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

var A = Identity[int](1)
var B = Pick[int](2)
var C = Add[int](3, 4)
var D = Add[float64](1.25, 2.5)
`);

    expect(result.diagnostics).toEqual([]);
    const identity = result.pkg.Scope().Lookup("Identity");
    expect(identity?.Type()?.String()).toBe("func[T any](value T) T");
    const box = result.pkg.Scope().Lookup("Box");
    expect(box?.Type()?.String()).toBe("Box");
  });

  test("attaches methods declared on generic receiver instantiations", () => {
    const result = check(`
package generic

type Box[T any] struct {
  Value T
}

func (b *Box[T]) Get() T {
  return b.Value
}

func UseConcrete(b *Box[int]) int {
  return b.Get()
}

func UseParameterized[T any](b *Box[T]) T {
  return b.Get()
}
`);

    expect(result.diagnostics).toEqual([]);
    const box = result.pkg.Scope().Lookup("Box")?.Type();
    expect(box).toBeInstanceOf(GoTypesNamed);
    const methodSet = NewMethodSet(NewPointer(box as GoTypesNamed));
    expect(Array.from({ length: methodSet.Len() }, (_, index) => methodSet.At(index).Obj().Name())).toEqual(["Get"]);
  });

  test("resolves later generic constraints while collecting type parameters", () => {
    const result = check(`
package generic

type Curve[P Point[P]] struct {
  Value P
}

type Point[P any] interface {
  Set(P) P
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("rejects type parameter operators not permitted by the constraint", () => {
    const result = check(`
package generic

func Bad[T any](left, right T) T {
  return left + right
}
`);

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toContain("invalid operation: operator Plus not defined on T");
  });

  test("indexes type parameters constrained to strings and byte slices", () => {
    const result = check(`
package generic

func HashStr[T string | []byte](sep T) uint32 {
  hash := uint32(0)
  for i := 0; i < len(sep); i++ {
    hash = hash*16777619 + uint32(sep[i])
  }
  return hash
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("accepts named type parameters constrained to byte slices or strings", () => {
    const result = check(`
package generic

func IsDigit[bytes []byte | string](s bytes, i int) bool {
  return s[i] >= '0' && s[i] <= '9'
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("allows conversions between struct types that differ only by tags", () => {
    const result = check(`
package conversions

type A struct { X int \`json:"x"\` }
type B struct { X int \`xml:"x"\` }

var _ = A(B{})
var _ = (*A)(new(B))
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("checks constant conversions for representability", () => {
    const result = check(`
package conversions

var _ = int8(1000)
var _ = bool(1)
`);

    const messages = result.diagnostics.map((diagnostic) => diagnostic.message);
    expect(messages.some((message) => message.includes("constant 1000 overflows int8"))).toBe(true);
    expect(messages.some((message) => message.includes("cannot convert") && message.includes("type bool"))).toBe(true);
  });

  test("keeps integer-valued float literals classified as floats", () => {
    const result = check(`
package floats

const MaxFloat64 = 0x1p1023 * (1 + (1 - 0x1p-52))

var _ float64 = MaxFloat64
var _ float64 = 1.79769313486231570814527423731704357e+308
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("permits integer-valued untyped float constants in integer contexts", () => {
    const result = check(`
package floats

const starvationThresholdNs = 1e6

func Starving(waitStartTime int64, runtimeNanotime int64) bool {
  return runtimeNanotime-waitStartTime > starvationThresholdNs
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("keeps exact numeric constants representable in uint64 contexts", () => {
    const result = check(`
package constants

var uint64pow10 = [...]uint64{
  1, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
  1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("uses standard sizes for unsafe Sizeof constants without explicit config sizes", () => {
    const result = check(`
package constants

import "unsafe"

const wordSize = unsafe.Sizeof(uintptr(0))

func Words(n int) uintptr {
  return uintptr(n) / wordSize
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("keeps single basic type parameter constraints precise", () => {
    const good = check(`
package generic

func Inc[T int](x T) T {
  return x + 1
}
`);
    expect(good.diagnostics).toEqual([]);

    const bad = check(`
package generic

func Bad[T int](x T) T {
  return x + "no"
}
`);
    expect(bad.diagnostics.map((diagnostic) => diagnostic.message).some((message) => message.includes("mismatched types"))).toBe(true);
  });

  test("records comma-ok map index expressions", () => {
    const result = check(`
package maps

var M map[int]int

func Has(k int) bool {
  if _, ok := M[k]; ok {
    return true
  }
  return false
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("rejects generic function instantiation with non-type arguments", () => {
    const result = check(`
package generic

func Identity[T any](value T) T {
  return value
}

var notAType = 3
var A = Identity[notAType](1)
`);

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toContain("notAType (package-level variable) is not a type");
  });

  test("builds package scopes, imports, methods, local scopes, and range variable types", () => {
    const result = check(`
package model

import f "fmt"

type Point struct {
  X, Y float64
}

func (p *Point) Scale(k float64) {
  p.X = p.X * k
}

var Counts = map[string]int64{"a": 1}

func Sum(xs []int64) int64 {
  var total int64
  for _, v := range xs {
    total = total + v
  }
  f.Printf("%v", total)
  return total
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("Point")?.constructor.name).toBe("TypeName");
    expect(result.pkg.Scope().Lookup("Counts")?.constructor.name).toBe("Var");
    expect(result.pkg.Scope().Lookup("Sum")?.Type()?.String()).toBe("func(xs []int64) int64");

    const point = result.pkg.Scope().Lookup("Point")?.Type();
    expect(point).toBeInstanceOf(GoTypesNamed);
    const methodSet = NewMethodSet(NewPointer(point as GoTypesNamed));
    expect(Array.from({ length: methodSet.Len() }, (_, index) => methodSet.At(index).Obj().Name())).toEqual(["Scale"]);

    const vDef = [...(result.info.Defs ?? new Map()).entries()].find(([ident]) => (ident as { name?: string }).name === "v")?.[1];
    expect(vDef).toBeInstanceOf(GoTypesVar);
    expect(vDef?.Type()?.String()).toBe("int64");
  });

  test("types spreadsheet cell and range references through configured namespaces", () => {
    const sheetNamespaces: Record<string, GoJuniorSheetNamespace> = {
      sheet: {},
      Data: {}
    };
    const result = check(`
package workbook

var a = sheet.A1
var b = sheet.B1
var r = Data.A1:B2
var first = r[0][0]
`, {
      sheetNamespaces
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("a")?.Type()?.String()).toBe("any");
    expect(result.pkg.Scope().Lookup("b")?.Type()?.String()).toBe("any");

    const range = result.pkg.Scope().Lookup("r")?.Type();
    expect(range).toBeInstanceOf(GoTypesSlice);
    expect(range?.String()).toBe("[][]any");
    expect(result.pkg.Scope().Lookup("first")?.Type()?.String()).toBe("any");
  });

  test("loads predeclared any before checking ordinary packages", () => {
    const result = check(`
package builtins

var x any

func Echo(v any) any {
  return v
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("x")?.Type()?.String()).toBe("any");
    expect(result.pkg.Scope().Lookup("Echo")?.Type()?.String()).toBe("func(v any) any");
  });

  test("accepts Go init functions without declaring them in package scope", () => {
    const result = check(`
package initpkg

var x int

func init() {
  x = 1
}

func Value() int {
  return x
}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("init")).toBeNull();
    expect(result.pkg.Scope().Lookup("Value")?.Type()?.String()).toBe("func() int");
  });

  test("scopes range variables to the for statement and preserves integer range key types", () => {
    const result = check(`
package rangescope

func R() {
  for i := range -1 {
    _ = i
  }
  for i := range 'a' {
    var _ *rune = &i
  }
}
`);

    expect(result.diagnostics).toEqual([]);
    const defs = [...(result.info.Defs ?? new Map()).entries()].filter(([ident]) => (ident as { name?: string }).name === "i");
    expect(defs.map(([, object]) => object?.Type()?.String())).toEqual(["int", "rune"]);
  });

  test("types array pointer indexing and slicing like Go", () => {
    const result = check(`
package arrayptr

var p = new([3]byte)
var b = p[0]
var s = p[0:]
var text = string(s)
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("b")?.Type()?.String()).toBe("byte");
    expect(result.pkg.Scope().Lookup("s")?.Type()?.String()).toBe("[]byte");
    expect(result.pkg.Scope().Lookup("text")?.Type()?.String()).toBe("string");
  });

  test("resolves constant identifiers in array lengths", () => {
    const result = check(`
package arraylen

const size = 16
var a [size]byte
`);

    expect(result.diagnostics).toEqual([]);
    const aType = result.pkg.Scope().Lookup("a")?.Type();
    expect(aType).toBeInstanceOf(GoTypesArray);
    expect(aType?.String()).toBe("[16]byte");
  });

  test("rejects mixed string and numeric addition", () => {
    const result = check(`
package workbook

var a = 10
var d = "hi"
var good = d + " there"
var bad = d + a
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("invalid operation: d + a (mismatched types string and int)");
    expect(result.pkg.Scope().Lookup("good")?.Type()?.String()).toBe("string");
  });

  test("types channel directions, sends, receives, goroutines, and select clauses", () => {
    const result = check(`
package concurrent

func Worker(ch chan int, send chan<- int, recv <-chan int, done chan struct{}) {
  go Worker(ch, send, recv, done)
  ch <- 1
  send <- <-recv
  select {
  case ch <- 2:
  case value := <-recv:
    _ = value
  case <-done:
  default:
  }
}
`);

    expect(result.diagnostics).toEqual([]);
  });

  test("rejects invalid channel sends and receives", () => {
    const result = check(`
package concurrent

func Bad(send chan<- int, recv <-chan int, n int) {
  recv <- 1
  _ = <-send
  n <- 2
  n <- recv
  send <- "bad"
}
`);

    expect(result.diagnostics.map((item) => item.message)).toEqual([
      "invalid operation: cannot send to receive-only channel <-chan int <-chan int",
      "invalid operation: cannot receive from send-only channel chan<- int chan<- int",
      "invalid operation: cannot send to non-channel int int",
      "invalid operation: cannot send to non-channel int int",
      "cannot use untyped string as int value in send"
    ]);
  });
});
