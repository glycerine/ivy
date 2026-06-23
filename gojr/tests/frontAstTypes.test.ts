import { describe, expect, test } from "./testHarness.js";
import {
  childNodes,
  ident,
  parseCellAddress,
  walk,
  type File
} from "../src/front/ast.js";
import {
  assignableTo,
  ChanType,
  FuncObject,
  InterfaceType,
  MapType,
  NamedType,
  newUniverse,
  ObjectKind,
  PackageInfo,
  PointerType,
  Scope,
  SignatureType,
  SliceType,
  StructType,
  tuple,
  TypeNameObject,
  varOf
} from "../src/front/types.js";
import { TokenKind } from "../src/front/token.js";

describe("Go-junior Go-style AST", () => {
  test("parses spreadsheet cell addresses into absolute row and column flags", () => {
    expect(parseCellAddress("$A$1")).toEqual({
      raw: "$A$1",
      column: "A",
      row: 1,
      absoluteColumn: true,
      absoluteRow: true
    });
    expect(parseCellAddress("BC23")).toMatchObject({
      column: "BC",
      row: 23,
      absoluteColumn: false,
      absoluteRow: false
    });
    expect(parseCellAddress("A0")).toBeUndefined();
  });

  test("walks Go-style AST nodes without depending on parser internals", () => {
    const file: File = {
      kind: "File",
      name: ident("main"),
      declarations: [{
        kind: "FuncDecl",
        name: ident("Double"),
        type: {
          kind: "FuncType",
          params: {
            kind: "FieldList",
            fields: [{
              kind: "Field",
              names: [ident("x")],
              type: ident("int64")
            }]
          },
          results: {
            kind: "FieldList",
            fields: [{
              kind: "Field",
              names: [],
              type: ident("int64")
            }]
          }
        },
        body: {
          kind: "BlockStmt",
          statements: [{
            kind: "ReturnStmt",
            results: [{
              kind: "BinaryExpr",
              left: ident("x"),
              op: TokenKind.Star,
              right: {
                kind: "BasicLit",
                token: TokenKind.IntLiteral,
                value: "2"
              }
            }]
          }]
        }
      }],
      imports: [],
      unresolved: [],
      comments: []
    };

    const seen: string[] = [];
    walk(file, (node) => seen.push(node.kind));

    expect(childNodes(file)).toHaveLength(2);
    expect(seen).toContain("FuncDecl");
    expect(seen).toContain("BinaryExpr");
    expect(seen.filter((kind) => kind === "Ident")).toHaveLength(6);
  });
});

describe("Go-junior Go-style types", () => {
  test("builds a universe scope with predeclared types, constants, nil, and builtins", () => {
    const universe = newUniverse();

    expect(universe.scope.lookup("int64")?.kind).toBe(ObjectKind.TypeName);
    expect(universe.scope.lookup("true")?.kind).toBe(ObjectKind.Const);
    expect(universe.scope.lookup("nil")?.kind).toBe(ObjectKind.Nil);
    expect(universe.scope.lookup("panic")?.kind).toBe(ObjectKind.Builtin);
    expect(universe.scope.names()).toContain("append");
  });

  test("supports nested scopes and duplicate detection", () => {
    const universe = newUniverse();
    const child = new Scope(universe.scope, "function");
    const first = varOf("x", universe.basic.int64);
    const duplicate = varOf("x", universe.basic.string);

    expect(child.insert(first)).toBeUndefined();
    expect(child.insert(duplicate)).toBe(first);
    expect(child.lookupParent("x")?.object).toBe(first);
    expect(child.lookupParent("string")?.object.kind).toBe(ObjectKind.TypeName);
  });

  test("formats signatures and compound types", () => {
    const universe = newUniverse();
    const signature = new SignatureType(
      undefined,
      tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))),
      tuple(varOf("", universe.basic.int64), varOf("", universe.basic.error)),
      true
    );

    expect(signature.typeString()).toBe("func(format string, args ...any) (int64, error)");
    expect(new MapType(universe.basic.string, universe.basic.int64).typeString()).toBe("map[string]int64");
    expect(new ChanType(universe.basic.int64).typeString()).toBe("chan int64");
    expect(new ChanType(universe.basic.string, "send").typeString()).toBe("chan<- string");
    expect(new ChanType(universe.basic.bool, "receive").typeString()).toBe("<-chan bool");
  });

  test("checks assignability for untyped constants, nil, named types, and method-set interfaces", () => {
    const universe = newUniverse();
    const pkg = new PackageInfo("workbook/geom", "geom");
    const pointName = new TypeNameObject("Point", universe.basic.invalid, pkg.scope, pkg);
    const point = new NamedType(pointName, new StructType([
      { name: "X", type: universe.basic.float64, embedded: false },
      { name: "Y", type: universe.basic.float64, embedded: false }
    ]));
    pointName.setType(point);

    const lenSig = new SignatureType(
      varOf("p", point),
      tuple(),
      tuple(varOf("", universe.basic.float64)),
      false
    );
    const scaleSig = new SignatureType(
      varOf("p", new PointerType(point)),
      tuple(varOf("k", universe.basic.float64)),
      tuple(),
      false
    );
    point.addMethod(new FuncObject("Len2", lenSig));
    point.addMethod(new FuncObject("Scale", scaleSig));

    const hasLen = new InterfaceType([new FuncObject("Len2", new SignatureType(
      undefined,
      tuple(),
      tuple(varOf("", universe.basic.float64)),
      false
    ))]).complete();
    const hasScale = new InterfaceType([new FuncObject("Scale", new SignatureType(
      undefined,
      tuple(varOf("k", universe.basic.float64)),
      tuple(),
      false
    ))]).complete();

    expect(assignableTo(universe.basic.untypedInt, universe.basic.int64)).toBe(true);
    expect(assignableTo(universe.basic.untypedNil, new SliceType(universe.basic.string))).toBe(true);
    expect(assignableTo(point, hasLen)).toBe(true);
    expect(assignableTo(point, hasScale)).toBe(false);
    expect(assignableTo(new PointerType(point), hasScale)).toBe(true);
  });
});
