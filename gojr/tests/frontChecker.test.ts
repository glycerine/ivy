import { describe, expect, test } from "./testHarness.js";
import { checkFrontSource } from "../src/front/checker.js";
import {
  ArrayType,
  FuncObject,
  NamedType,
  PackageInfo,
  PointerType,
  SignatureType,
  SliceType,
  TypeKind,
  VarObject,
  methodSet,
  newUniverse,
  tuple,
  varOf
} from "../src/front/types.js";
import type { CheckConfig } from "../src/front/checker.js";

const TEST_FILENAME = "front-checker-test.go";

function check(source: string, config: CheckConfig = {}) {
  return checkFrontSource(source, TEST_FILENAME, config);
}

describe("Go-junior TypeScript front checker", () => {
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
    const identity = result.pkg.scope.lookup("Identity");
    expect(identity?.type.typeString()).toBe("func(value T) T");
    const box = result.pkg.scope.lookup("Box");
    expect(box?.type.typeString()).toBe("generic.Box");
  });

  test("rejects type parameter operators not permitted by the constraint", () => {
    const result = check(`
package generic

func Bad[T any](left, right T) T {
  return left + right
}
`);

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toContain("invalid operation: T + T (operator + not permitted by constraint)");
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

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toContain("notAType is not a type");
  });

  test("builds package scopes, imports, methods, local scopes, and range variable types", () => {
    const universe = newUniverse();
    const fmt = new PackageInfo("fmt", "fmt", universe.scope);
    fmt.scope.insert(new FuncObject(
      "Printf",
      new SignatureType(
        undefined,
        tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))),
        tuple(varOf("", universe.basic.int64)),
        true
      ),
      fmt.scope,
      fmt
    ));

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
`, {
      universe,
      importer: {
        import(path) {
          return path === "fmt" ? fmt : undefined;
        }
      }
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.scope.lookup("Point")?.kind).toBe("TypeName");
    expect(result.pkg.scope.lookup("Counts")?.kind).toBe("Var");
    expect(result.pkg.scope.lookup("Sum")?.type.typeString()).toBe("func(xs []int64) int64");

    const point = result.pkg.scope.lookup("Point")?.type;
    expect(point).toBeInstanceOf(NamedType);
    expect(methodSet(new PointerType(point as NamedType)).map((method) => method.name)).toEqual(["Scale"]);

    const vDef = [...result.info.defs.entries()].find(([ident]) => ident.name === "v")?.[1];
    expect(vDef).toBeInstanceOf(VarObject);
    expect(vDef?.type.typeString()).toBe("int64");
  });

  test("types spreadsheet cell and range references through configured namespaces", () => {
    const universe = newUniverse();
    const result = check(`
package workbook

var a = sheet.A1 + sheet.B1
var r = Data.A1:B2
var first = r[0][0]
`, {
      universe,
      sheetNamespaces: {
        sheet: {
          cells: {
            A1: universe.basic.int64,
            B1: universe.basic.int64
          }
        },
        Data: {
          defaultType: universe.basic.float64
        }
      }
    });

    expect(result.diagnostics).toEqual([]);
    expect(result.pkg.scope.lookup("a")?.type.typeString()).toBe("int64");

    const range = result.pkg.scope.lookup("r")?.type;
    expect(range?.kind).toBe(TypeKind.Slice);
    expect(range?.typeString()).toBe("[][]float64");
    expect(result.pkg.scope.lookup("first")?.type.typeString()).toBe("float64");
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
    const defs = [...result.info.defs.entries()].filter(([ident]) => ident.name === "i");
    expect(defs.map(([, object]) => object?.type.typeString())).toEqual(["int64", "rune"]);
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
    expect(result.pkg.scope.lookup("b")?.type.typeString()).toBe("byte");
    expect(result.pkg.scope.lookup("s")?.type.typeString()).toBe("[]byte");
    expect(result.pkg.scope.lookup("text")?.type.typeString()).toBe("string");
  });

  test("resolves constant identifiers in array lengths", () => {
    const result = check(`
package arraylen

const size = 16
var a [size]byte
`);

    expect(result.diagnostics).toEqual([]);
    const aType = result.pkg.scope.lookup("a")?.type;
    expect(aType).toBeInstanceOf(ArrayType);
    expect(aType?.typeString()).toBe("[16]byte");
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
    expect(result.diagnostics[0]?.message).toContain("invalid operation: string + int64");
    expect(result.pkg.scope.lookup("good")?.type.typeString()).toBe("string");
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
      "cannot send to receive-only channel <-chan int",
      "cannot receive from send-only channel chan<- int",
      "cannot send to non-channel int",
      "cannot send to non-channel int; `<-` between expressions is a send, use `x = <-ch` to receive into an existing variable",
      "cannot send untyped string as int"
    ]);
  });
});
