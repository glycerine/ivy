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

  test("parses generic function and type parameter lists", () => {
    const file = parseOk(`
package generic

type Box[T any] struct {
  Value T
}

type Number interface {
  ~int | ~float64
}

func Identity[T any](value T) T {
  return value
}

func Pick[T Number](value T) T {
  return value
}

var answer = Identity[int](42)
`);

    const typeDecl = file.declarations.find((decl): decl is GenDecl => decl.kind === "GenDecl");
    const typeSpec = typeDecl?.specs[0];
    expect(typeSpec?.kind).toBe("TypeSpec");
    if (typeSpec?.kind === "TypeSpec") {
      expect(typeSpec.typeParams?.fields[0]?.names.map((name) => name.name)).toEqual(["T"]);
      expect((typeSpec.typeParams?.fields[0]?.type as Expr | undefined)?.kind).toBe("Ident");
    }

    const fn = file.declarations.find((decl): decl is FuncDecl => decl.kind === "FuncDecl");
    expect(fn?.type.typeParams?.fields[0]?.names.map((name) => name.name)).toEqual(["T"]);
    expect(fn?.type.params.fields[0]?.type.kind).toBe("Ident");
    expect(fn?.type.results?.fields[0]?.type.kind).toBe("Ident");
    const numberSpec = file.declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs : [])
      .find((spec) => spec.kind === "TypeSpec" && spec.name.name === "Number");
    const term = numberSpec?.kind === "TypeSpec" && numberSpec.type.kind === "InterfaceType"
      ? numberSpec.type.methods.fields[0]?.type
      : undefined;
    expect(term?.kind).toBe("BinaryExpr");
    if (term?.kind === "BinaryExpr") {
      expect(term.left.kind).toBe("UnaryExpr");
      expect(term.right.kind).toBe("UnaryExpr");
    }
    const varDecl = file.declarations.find((decl): decl is GenDecl =>
      decl.kind === "GenDecl" && decl.specs.some((spec) => spec.kind === "ValueSpec")
    );
    const valueSpec = varDecl?.specs.find((spec) => spec.kind === "ValueSpec");
    const call = valueSpec?.kind === "ValueSpec" ? valueSpec.values[0] : undefined;
    expect(call?.kind).toBe("CallExpr");
    if (call?.kind === "CallExpr") expect(call.fun.kind).toBe("IndexExpr");
  });

  test("parses multiple type arguments as Go ast IndexListExpr", () => {
    const nodes = collectNodes(`
package generic

type Pair[A, B any] struct {
  A A
  B B
}

func Make[A, B any](a A, b B) Pair[A, B] {
  return Pair[A, B]{A: a, B: b}
}

var value = Make[int, string](1, "one")
`);

    expect(nodes.some((node) => node.kind === "IndexListExpr")).toBe(true);
  });

  test("parses embedded instantiated generic fields", () => {
    const nodes = collectNodes(`
package generic

type A[T any] struct {
  B[T]
}

type B[T any] struct {
  val T
}
`);

    expect(nodes.some((node) => node.kind === "IndexExpr")).toBe(true);
    const fields = nodes.filter((node): node is Field => node.kind === "Field");
    expect(fields.some((field) => field.names.length === 0 && field.type.kind === "IndexExpr")).toBe(true);
  });

  test("parses parenthesized types in channel and function result types", () => {
    const nodes = collectNodes(`
package types

var C chan<- (chan int)
var F func() (func())
`);

    expect(nodes.some((node) => node.kind === "ParenExpr")).toBe(true);
    expect(nodes.filter((node) => node.kind === "ChanType")).toHaveLength(2);
  });

  test("keeps slice and array type declarations distinct from type parameter lists", () => {
    const file = parseOk(`
package slices

type Slice []int
type Array [3]int
type UnsafeArray [unsafe.Sizeof(byte(0))]*byte
type Box[T any] struct { Value T }
`);

    const specs = file.declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs : []);
    const slice = specs.find((spec) => spec.kind === "TypeSpec" && spec.name.name === "Slice");
    const array = specs.find((spec) => spec.kind === "TypeSpec" && spec.name.name === "Array");
    const unsafeArray = specs.find((spec) => spec.kind === "TypeSpec" && spec.name.name === "UnsafeArray");
    const box = specs.find((spec) => spec.kind === "TypeSpec" && spec.name.name === "Box");

    expect(slice?.kind === "TypeSpec" ? slice.type.kind : undefined).toBe("ArrayType");
    expect(array?.kind === "TypeSpec" ? array.type.kind : undefined).toBe("ArrayType");
    expect(unsafeArray?.kind === "TypeSpec" ? unsafeArray.type.kind : undefined).toBe("ArrayType");
    expect(box?.kind === "TypeSpec" ? box.typeParams?.fields[0]?.names.map((name) => name.name) : undefined).toEqual(["T"]);
  });

  test("parses structs, interfaces, maps, function literals, and composite literals", () => {
    const kinds = collectKinds(`
package model

type Point struct {
  X, Y float64
  F0 [0]struct{}
  F1 float32
  F2 [0]struct{}
}

type Stringer interface {
  String() string
  F1() int
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

  test("parses elided composite literals inside typed composite literals", () => {
    const nodes = collectNodes(`
package model

var rows = []struct {
  name string
  vals []int
}{
  {"a", []int{1, 2}},
  {"b", nil},
}

type Point struct{ X, Y int }
var lookup = map[Point]Point{
  {X: 1}: {Y: 2},
}
`);

    const literals = nodes.filter((node) => node.kind === "CompositeLit");
    expect(literals.length).toBeGreaterThan(5);
    expect(literals.some((node) => node.kind === "CompositeLit" && !node.type)).toBe(true);
  });

  test("parses composite literal operands inside if header indexes and calls", () => {
    const nodes = collectNodes(`
package headers

type T [1]byte

func ok(T) bool { return true }

func F(m map[T][1]byte) {
  if x, y := m[T{}][0], m[T{1}][0]; x != y {
  }
  if ok(T{}) {
  }
  if (T{}) == (T{}) {
  }
}
`);

    expect(nodes.filter((node) => node.kind === "IfStmt")).toHaveLength(3);
    expect(nodes.filter((node) => node.kind === "CompositeLit").length).toBeGreaterThan(4);
    expect(nodes.filter((node) => node.kind === "IndexExpr").length).toBeGreaterThan(3);
  });

  test("parses spreadsheet cell and range references as explicit AST nodes", () => {
    const nodes = collectNodes(`
package workbook

var x = sheet.$A$1
var y = sheet.A1
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
    expect(nodes.some((node) => node.kind === "SelectorExpr" && node.selector.name === "A1")).toBe(true);
  });

  test("does not confuse ordinary Go selectors with spreadsheet cells", () => {
    const nodes = collectNodes(`
package selectors

func F(i1 one.I1) {
  switch v := i1.(type) {
  case two.S2:
    one.F1(v)
  }
}
`);

    expect(nodes.filter((node) => node.kind === "RangeRefExpr")).toHaveLength(0);
    expect(nodes.filter((node) => node.kind === "CellRefExpr")).toHaveLength(0);
    expect(nodes.some((node) => node.kind === "SelectorExpr" && node.selector.name === "F1")).toBe(true);
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
  var a [2]int
  q := 0
  for a[func() int {
    q++
    return 0
  }()] = range [2]int{} {
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
    expect(nodes.filter((node) => node.kind === "RangeStmt")).toHaveLength(2);
  });

  test("parses Go for clauses with omitted init statements", () => {
    const nodes = collectNodes(`
package control

func F() int {
  i := 0
  sum := 0
  for ; i < 5; i++ {
    sum += i
  }
  for ; ; i-- {
    if i == 0 {
      break
    }
  }
  return sum
}
`);

    const loops = nodes.filter((node) => node.kind === "ForStmt");
    expect(loops).toHaveLength(2);
    expect(loops[0]).toMatchObject({ kind: "ForStmt" });
    expect(loops[0]?.kind === "ForStmt" ? loops[0].init : undefined).toBeUndefined();
    expect(loops[0]?.kind === "ForStmt" ? loops[0].condition?.kind : undefined).toBe("BinaryExpr");
    expect(loops[0]?.kind === "ForStmt" ? loops[0].post?.kind : undefined).toBe("IncDecStmt");
    expect(loops[1]).toMatchObject({ kind: "ForStmt" });
    expect(loops[1]?.kind === "ForStmt" ? loops[1].init : undefined).toBeUndefined();
    expect(loops[1]?.kind === "ForStmt" ? loops[1].condition : undefined).toBeUndefined();
    expect(loops[1]?.kind === "ForStmt" ? loops[1].post?.kind : undefined).toBe("IncDecStmt");
  });

  test("parses standard for-range headers without lookahead specialization", () => {
    const nodes = collectNodes(`
package control

func F(xs []int) {
  for range xs {
  }
  for i := range xs {
    _ = i
  }
  for i, v = range xs {
    _, _ = i, v
  }
}
`);

    const ranges = nodes.filter((node) => node.kind === "RangeStmt");
    expect(ranges).toHaveLength(3);
    expect(ranges[0]?.kind === "RangeStmt" ? ranges[0].key : undefined).toBeUndefined();
    expect(ranges[1]?.kind === "RangeStmt" ? ranges[1].token : undefined).toBe(TokenKind.Define);
    expect(ranges[2]?.kind === "RangeStmt" ? ranges[2].token : undefined).toBe(TokenKind.Assign);
  });

  test("parses labels through simple statement mode", () => {
    const nodes = collectNodes(`
package labels

func F() {
top:
  for {
  inner:
    for {
      break inner
    }
    continue top
  }
}
`);

    expect(nodes.filter((node) => node.kind === "LabeledStmt")).toHaveLength(2);
    expect(nodes.filter((node) => node.kind === "BranchStmt")).toHaveLength(2);
  });

  test("parses labels immediately before closing braces", () => {
    const nodes = collectNodes(`
package labels

func F() {
  goto done
done:
}
`);

    expect(nodes.some((node) => node.kind === "LabeledStmt")).toBe(true);
    expect(nodes.some((node) => node.kind === "EmptyStmt")).toBe(true);
  });

  test("treats '=' as equality while parsing right-hand-side lists", () => {
    const nodes = collectNodes(`
package rhs

func F(x, y int) {
  switch {
  case x = y:
  }
}
`);

    const binary = nodes.find((node) => node.kind === "BinaryExpr");
    expect(binary?.kind === "BinaryExpr" ? binary.op : undefined).toBe(TokenKind.Equal);
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

  test("parses switches with an empty init statement before semicolon", () => {
    const nodes = collectNodes(`
package switches

func F() {
  switch ; {
  case true:
    return
  default:
    return
  }
}
`);

    expect(nodes.some((node) => node.kind === "SwitchStmt")).toBe(true);
    expect(nodes.filter((node) => node.kind === "CaseClause")).toHaveLength(2);
  });

  test("keeps control-clause block braces out of composite literal parsing", () => {
    const nodes = collectNodes(`
package switches

func source(item interface{}) error {
  var err error
  var sourceItem func(item Sexp) error

  sourceItem = func(item Sexp) error {
    switch t := item.(type) {
    case *SexpArray:
      for _, v := range t.Val {
        if err := sourceItem(v); err != nil {
          return err
        }
      }
    case *SexpPair:
      expr := item
      for expr != SexpNull {
        list := expr.(*SexpPair)
        if err := sourceItem(list.Head); err != nil {
          return err
        }
        expr = list.Tail
      }
    case *SexpStr:
      _ = t
    default:
      _ = err
    }
    return nil
  }
  return sourceItem(item)
}

type Sexp interface{}
type SexpArray struct{ Val []Sexp }
type SexpPair struct{ Head Sexp; Tail Sexp }
type SexpStr struct{ S string }
var SexpNull Sexp
`);

    const typeSwitch = nodes.find((node) => node.kind === "TypeSwitchStmt");
    const range = nodes.find((node) => node.kind === "RangeStmt");
    const cases = nodes.filter((node) => node.kind === "CaseClause");

    expect(typeSwitch).toMatchObject({ kind: "TypeSwitchStmt" });
    expect(range).toMatchObject({ kind: "RangeStmt" });
    expect(cases).toHaveLength(4);
  });

  test("parses goroutines, channel types, sends, receives, and select clauses", () => {
    const nodes = collectNodes(`
package concurrent

func F(ch chan int, send chan<- int, recv <-chan int, done chan struct{}) {
  go F(ch, send, recv, done)
  ch <- 1
  x := <-recv
  _ = x
  select {
  case send <- x:
  case y, ok := <-ch:
    _, _ = y, ok
  case <-done:
    return
  default:
  }
}
`);

    expect(nodes.filter((node) => node.kind === "ChanType")).toHaveLength(4);
    expect(nodes.some((node) => node.kind === "GoStmt")).toBe(true);
    expect(nodes.some((node) => node.kind === "SendStmt")).toBe(true);
    expect(nodes.some((node) => node.kind === "SelectStmt")).toBe(true);
    expect(nodes.filter((node) => node.kind === "CommClause")).toHaveLength(4);
    expect(nodes.some((node) => node.kind === "UnaryExpr" && node.op === TokenKind.Arrow)).toBe(true);
  });
});
