import { describe, expect, test } from "./testHarness.js";
import { checkFrontSource } from "../src/front/checker.js";
import {
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
    expect(range?.typeString()).toBe("[]float64");
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
});
