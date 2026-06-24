import { describe, expect, test } from "./testHarness.js";
import { checkGoJuniorSource } from "../src/typecheck.js";
import {
  Array as GoTypesArray,
  Float64,
  Int64,
  Named as GoTypesNamed,
  NewMethodSet,
  NewPointer,
  Slice as GoTypesSlice,
  Typ,
  Var as GoTypesVar,
  type GoJuniorSheetNamespace
} from "../src/go/types/index.js";
import type { GoJuniorCheckConfig } from "../src/typecheck.js";

const TEST_FILENAME = "front-checker-test.go";

function check(source: string, config: GoJuniorCheckConfig = {}) {
  return checkGoJuniorSource(source, TEST_FILENAME, config);
}

describe("Go-junior TypeScript go/types checker", () => {
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

  test("rejects type parameter operators not permitted by the constraint", () => {
    const result = check(`
package generic

func Bad[T any](left, right T) T {
  return left + right
}
`);

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toContain("invalid operation: operator Plus not defined on T");
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
      sheet: {
        cells: {
          A1: Typ[Int64]!,
          B1: Typ[Int64]!
        }
      },
      Data: {
        defaultType: Typ[Float64]!
      }
    };
    const result = check(`
package workbook

var a = sheet.A1 + sheet.B1
var r = Data.A1:B2
var first = r[0][0]
`, {
      sheetNamespaces
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.Scope().Lookup("a")?.Type()?.String()).toBe("int64");

    const range = result.pkg.Scope().Lookup("r")?.Type();
    expect(range).toBeInstanceOf(GoTypesSlice);
    expect(range?.String()).toBe("[][]float64");
    expect(result.pkg.Scope().Lookup("first")?.Type()?.String()).toBe("float64");
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
