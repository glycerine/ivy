import { describe, expect, test } from "./testHarness.js";
import { parseFrontSource } from "../src/front/parser.js";
import { TokenKind } from "../src/front/token.js";
import { walk, type AstNode, type Expr, type Field, type FuncDecl, type GenDecl } from "../src/front/ast.js";

const TEST_FILENAME = "front-parser-test.go";

function parseOk(source: string) {
  const result = parseFrontSource(source, TEST_FILENAME);
  expect(result.diagnostics).toEqual([]);
  expect(result.file).toBeDefined();
  return result.file!;
}

function collectKinds(source: string): string[] {
  const file = parseOk(source);
  const kinds: string[] = [];
  walk(file, (node) => kinds.push(node.kind));
  return kinds;
}

function collectNodes(source: string): AstNode[] {
  const file = parseOk(source);
  const nodes: AstNode[] = [];
  walk(file, (node) => nodes.push(node));
  return nodes;
}

describe("Go-junior TypeScript front parser", () => {
  test("parses packages, imports, grouped parameters, named results, and varargs", () => {
    const file = parseOk(`
package stats

import (
  f "fmt"
  "math"
)

func PairAndLog(a, b, c int, rest ...string) (left, right int) {
  f.Printf("%#v", rest...)
  return a + b*c, c
}
`);

    expect(file.name?.name).toBe("stats");
    expect(file.imports.map((item) => [item.name?.name, item.path.value])).toEqual([
      ["f", "\"fmt\""],
      [undefined, "\"math\""]
    ]);

    const fn = file.declarations.find((decl): decl is FuncDecl => decl.kind === "FuncDecl");
    expect(fn?.name.name).toBe("PairAndLog");
    expect(fn?.type.params.fields).toHaveLength(2);
    expect(fn?.type.params.fields[0]?.names.map((name) => name.name)).toEqual(["a", "b", "c"]);
    expect(fn?.type.params.fields[1]?.names.map((name) => name.name)).toEqual(["rest"]);
    expect(fn?.type.params.fields[1]?.type.kind).toBe("Ellipsis");
    expect(fn?.type.results?.fields[0]?.names.map((name) => name.name)).toEqual(["left", "right"]);
  });

  test("parses structs, interfaces, maps, function literals, and composite literals", () => {
    const kinds = collectKinds(`
package model

type Point struct {
  X, Y float64
}

type Stringer interface {
  String() string
}

var counts = map[string]int64{"a": 1, "b": 2}
var f = func(a, b, c int) (d, e, f int) { return b, c, a }
`);

    expect(kinds).toContain("StructType");
    expect(kinds).toContain("InterfaceType");
    expect(kinds).toContain("MapType");
    expect(kinds).toContain("CompositeLit");
    expect(kinds).toContain("FuncLit");
    expect(kinds.filter((kind) => kind === "KeyValueExpr")).toHaveLength(2);
  });

  test("parses spreadsheet cell and range references as explicit AST nodes", () => {
    const nodes = collectNodes(`
package workbook

var x = sheet.$A$1
var r = Data.A1:B10
`);

    const cell = nodes.find((node) => node.kind === "CellRefExpr");
    const range = nodes.find((node) => node.kind === "RangeRefExpr");

    expect(cell).toMatchObject({
      kind: "CellRefExpr",
      namespace: { name: "sheet" },
      address: { raw: "$A$1", absoluteColumn: true, absoluteRow: true }
    });
    expect(range).toMatchObject({
      kind: "RangeRefExpr",
      namespace: { name: "Data" },
      start: { raw: "A1" },
      end: { raw: "B10" }
    });
  });

  test("keeps operator precedence and parses control flow labels and range loops", () => {
    const file = parseOk(`
package loops

func Sum(xs []int) int {
top:
  for i, v := range xs {
    if i > 10 {
      break top
    }
    _ = v + i * 2
  }
  return 0
}
`);

    const nodes: AstNode[] = [];
    walk(file, (node) => nodes.push(node));
    expect(nodes.some((node) => node.kind === "LabeledStmt")).toBe(true);
    expect(nodes.some((node) => node.kind === "RangeStmt")).toBe(true);

    const assign = nodes.find((node) => node.kind === "AssignStmt");
    const rhs = assign?.kind === "AssignStmt" ? assign.rhs[0] : undefined;
    expect(rhs?.kind).toBe("BinaryExpr");
    const right = rhs?.kind === "BinaryExpr" ? rhs.right : undefined;
    expect(right?.kind).toBe("BinaryExpr");
    expect(right?.kind === "BinaryExpr" ? right.op : undefined).toBe(TokenKind.Star);
  });

  test("parses value switches and type switches with case clauses", () => {
    const nodes = collectNodes(`
package switches

func Classify(x int, any interface{}) string {
  switch x {
  case 1, 2:
    fallthrough
  default:
    fmt.Printf("other")
  }

  switch v := any.(type) {
  case int:
    return "int"
  case string:
    return v
  default:
    return "unknown"
  }
}
`);

    const valueSwitch = nodes.find((node) => node.kind === "SwitchStmt");
    const typeSwitch = nodes.find((node) => node.kind === "TypeSwitchStmt");
    const cases = nodes.filter((node) => node.kind === "CaseClause");

    expect(valueSwitch).toMatchObject({ kind: "SwitchStmt" });
    expect(typeSwitch).toMatchObject({ kind: "TypeSwitchStmt" });
    expect(cases).toHaveLength(5);
    expect(cases.filter((node) => node.kind === "CaseClause" && node.default)).toHaveLength(2);
    expect(nodes.some((node) => node.kind === "BranchStmt" && node.token === TokenKind.Fallthrough)).toBe(true);
  });

  test("reports unsupported go and select statements", () => {
    const result = parseFrontSource(`
package unsupported

func F() {
  go F()
  select {}
}
`, TEST_FILENAME);

    expect(result.diagnostics.map((item) => item.code)).toEqual([
      "GOJR_PARSE_UNSUPPORTED",
      "GOJR_PARSE_UNSUPPORTED"
    ]);
  });
});
