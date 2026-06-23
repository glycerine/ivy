import { describe, expect, test } from "./testHarness.js";
import {
  childNodes,
  commentGroupText,
  ident,
  IsExported,
  IsGenerated,
  FileExports,
  FilterFuncDuplicates,
  FilterImportDuplicates,
  Fprint,
  collapse,
  MergePackageFiles,
  NewIdent,
  NewObj,
  NewPackage,
  NewScope,
  NotNilFilter,
  ObjKind,
  ParseDirective,
  Pkg,
  SortImports,
  Var as AstVar,
  importComment,
  importName,
  importPath,
  parseCellAddress,
  Inspect,
  Preorder,
  PreorderStack,
  Unparen,
  Walk,
  walk,
  type AstNode,
  type ImportSpec,
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

    const visitorSeen: string[] = [];
    Walk({
      Visit(node: AstNode | undefined) {
        visitorSeen.push(node?.kind ?? "<nil>");
        return node?.kind === "FuncDecl" ? undefined : this;
      }
    }, file);
    expect(visitorSeen).toContain("File");
    expect(visitorSeen).toContain("FuncDecl");
    expect(visitorSeen).toContain("<nil>");
    expect(visitorSeen).not.toContain("ReturnStmt");

    const inspected: string[] = [];
    Inspect(file, (node) => {
      inspected.push(node?.kind ?? "<nil>");
      return node?.kind !== "FuncDecl";
    });
    expect(inspected).toEqual(visitorSeen);

    expect([...Preorder(file)].map((node) => node.kind).slice(0, 3)).toEqual(["File", "Ident", "FuncDecl"]);

    const stacks: string[] = [];
    PreorderStack(file, [], (node, stack) => {
      stacks.push(`${node.kind}:${stack.map((entry) => entry.kind).join("/")}`);
      return true;
    });
    expect(stacks[0]).toBe("File:");
    expect(stacks).toContain("FuncDecl:File");
  });

  test("transliterates go/ast comment and identifier helpers", () => {
    expect(NewIdent("Thing")).toEqual({ kind: "Ident", name: "Thing" });
    expect(IsExported("Thing")).toBe(true);
    expect(IsExported("thing")).toBe(false);

    expect(commentGroupText({
      kind: "CommentGroup",
      list: [
        { kind: "Comment", text: "// first  " },
        { kind: "Comment", text: "//go:noinline" },
        { kind: "Comment", text: "/*\nsecond\n\n*/" }
      ]
    })).toEqual("first\n\nsecond\n");

    const wrapped = {
      kind: "ParenExpr" as const,
      expr: {
        kind: "ParenExpr" as const,
        expr: ident("x")
      }
    };
    expect(Unparen(wrapped)).toEqual(ident("x"));
  });

  test("detects generated source comments before the package clause", () => {
    const file: File = {
      kind: "File",
      name: ident("main", { filename: "generated.go", offset: 100, length: 4, line: 4, column: 9 }),
      declarations: [],
      imports: [],
      unresolved: [],
      comments: [{
        kind: "CommentGroup",
        list: [{
          kind: "Comment",
          text: "// Code generated gojr-test DO NOT EDIT.",
          span: { filename: "generated.go", offset: 0, length: 40, line: 1, column: 1 }
        }]
      }]
    };
    expect(IsGenerated(file)).toBe(true);
  });

  test("transliterates go/ast directive parsing", () => {
    const [directive, ok] = ParseDirective(10, "//go:generate stringer -type Op `raw arg` \"quoted arg\"");
    expect(ok).toBe(true);
    expect(directive.Tool).toBe("go");
    expect(directive.Name).toBe("generate");
    expect(directive.Args).toBe("stringer -type Op `raw arg` \"quoted arg\"");
    expect(directive.Pos()).toBe(10);
    expect(directive.End()).toBe(10 + "//go:generate ".length + directive.Args.length);

    const [args, err] = directive.ParseArgs();
    expect(err).toBeUndefined();
    expect(args.map((arg) => [arg.Arg, arg.Pos])).toEqual([
      ["stringer", 24],
      ["-type", 33],
      ["Op", 39],
      ["raw arg", 42],
      ["quoted arg", 52]
    ]);

    expect(ParseDirective(0, "//not a directive")[1]).toBe(false);
    expect(ParseDirective(0, "/*go:generate nope*/")[1]).toBe(false);
  });

  test("transliterates go/ast export filtering and package merging", () => {
    const file: File = {
      kind: "File",
      name: ident("p"),
      declarations: [
        {
          kind: "FuncDecl",
          name: ident("Exported"),
          type: { kind: "FuncType", params: { kind: "FieldList", fields: [] } }
        },
        {
          kind: "FuncDecl",
          name: ident("hidden"),
          type: { kind: "FuncType", params: { kind: "FieldList", fields: [] } }
        },
        {
          kind: "GenDecl",
          token: TokenKind.Type,
          grouped: false,
          specs: [{
            kind: "TypeSpec",
            name: ident("Thing"),
            alias: false,
            type: {
              kind: "StructType",
              fields: {
                kind: "FieldList",
                fields: [
                  { kind: "Field", names: [ident("Public")], type: ident("int") },
                  { kind: "Field", names: [ident("private")], type: ident("int") }
                ]
              }
            }
          }]
        }
      ],
      imports: [],
      unresolved: [],
      comments: []
    };

    expect(FileExports(file)).toBe(true);
    expect(file.declarations.map((decl) => decl.kind === "FuncDecl" ? decl.name.name : "type")).toEqual(["Exported", "type"]);
    const typeDecl = file.declarations[1];
    if (typeDecl?.kind !== "GenDecl" || typeDecl.specs[0]?.kind !== "TypeSpec" || typeDecl.specs[0].type.kind !== "StructType") {
      throw new Error("expected filtered struct type");
    }
    expect(typeDecl.specs[0].type.fields.fields.map((field) => field.names[0]?.name)).toEqual(["Public"]);

    const merged = MergePackageFiles({
      kind: "Package",
      name: "p",
      files: [
        file,
        {
          kind: "File",
          name: ident("p"),
          declarations: [
            {
              kind: "FuncDecl",
              name: ident("Exported"),
              type: { kind: "FuncType", params: { kind: "FieldList", fields: [] } }
            },
            {
              kind: "GenDecl",
              token: TokenKind.Import,
              grouped: false,
              specs: [
                { kind: "ImportSpec", path: { kind: "BasicLit", token: TokenKind.StringLiteral, value: "\"fmt\"" } },
                { kind: "ImportSpec", path: { kind: "BasicLit", token: TokenKind.StringLiteral, value: "\"fmt\"" } }
              ]
            }
          ],
          imports: [],
          unresolved: [],
          comments: []
        }
      ]
    }, FilterFuncDuplicates | FilterImportDuplicates);

    expect(merged.declarations.filter((decl) => decl.kind === "FuncDecl" && decl.name.name === "Exported")).toHaveLength(1);
    expect(merged.imports.map((spec) => spec.path.value)).toEqual(["\"fmt\""]);
  });

  test("transliterates go/ast import sorting helpers", () => {
    const spec = (value: string, line: number, offset: number, name?: string, comment?: string): ImportSpec => ({
      kind: "ImportSpec",
      ...(name ? { name: ident(name, { filename: "imports.go", offset, length: name.length, line, column: 8 }) } : {}),
      path: {
        kind: "BasicLit",
        token: TokenKind.StringLiteral,
        value,
        span: { filename: "imports.go", offset, length: value.length, line, column: name ? 14 : 8 }
      },
      ...(comment ? { comment: { kind: "CommentGroup", list: [{ kind: "Comment", text: comment }] } } : {}),
      span: { filename: "imports.go", offset, length: value.length + (name?.length ?? 0), line, column: 2 }
    });

    const z1 = spec("\"z\"", 2, 20);
    const a = spec("\"a\"", 3, 30, "alias");
    const z2 = spec("\"z\"", 4, 40);
    const m = spec("`m`", 5, 50);
    const withComment = spec("\"z\"", 6, 60, undefined, "// keep");
    const decl = {
      kind: "GenDecl" as const,
      token: TokenKind.Import as const,
      grouped: true,
      specs: [z1, a, z2, m, withComment],
      span: { filename: "imports.go", offset: 10, length: 70, line: 1, column: 1 }
    };
    const file: File = {
      kind: "File",
      name: ident("p"),
      declarations: [decl],
      imports: [z1, a, z2, m, withComment],
      unresolved: [],
      comments: []
    };

    expect(importPath(a)).toBe("a");
    expect(importPath(m)).toBe("m");
    expect(importName(a)).toBe("alias");
    expect(importComment(withComment)).toBe("keep\n");
    expect(collapse(z1, z2)).toBe(true);
    expect(collapse(withComment, z2)).toBe(false);

    SortImports(undefined, file);

    expect(decl.specs.map((s) => s.kind === "ImportSpec" ? [importName(s), importPath(s), importComment(s)] : [])).toEqual([
      ["alias", "a", ""],
      ["", "m", ""],
      ["", "z", "keep\n"]
    ]);
    expect(file.imports).toEqual(decl.specs);
    expect(a.path.span?.offset).toBe(20);
    expect(m.path.span?.offset).toBe(30);
    expect(withComment.path.span?.offset).toBe(40);
  });

  test("transliterates go/ast scope and package resolution helpers", () => {
    const scope = NewScope();
    const xName = ident("X", { filename: "p.go", offset: 12, length: 1, line: 2, column: 5 });
    const xSpec = { kind: "ValueSpec" as const, names: [xName], values: [] };
    const xObj = NewObj(AstVar, "X");
    xObj.Decl = xSpec;

    expect(scope.Lookup("X")).toBeUndefined();
    expect(scope.Insert(xObj)).toBeUndefined();
    expect(scope.Lookup("X")).toBe(xObj);
    expect(scope.Insert(NewObj(AstVar, "X"))).toBe(xObj);
    expect(xObj.Pos()).toBe(12);
    expect(ObjKind.String(AstVar)).toBe("var");
    expect(scope.String()).toContain("var X");

    const fileScope = NewScope();
    fileScope.Insert(xObj);
    const unresolvedX = ident("X", { filename: "p.go", offset: 30, length: 1, line: 4, column: 2 });
    const unresolvedFmt = ident("fmt", { filename: "p.go", offset: 40, length: 3, line: 5, column: 2 });
    const unresolvedMissing = ident("Missing", { filename: "p.go", offset: 50, length: 7, line: 6, column: 2 });
    const fmtImport: ImportSpec = {
      kind: "ImportSpec",
      path: { kind: "BasicLit", token: TokenKind.StringLiteral, value: "\"fmt\"", span: { filename: "p.go", offset: 20, length: 5, line: 3, column: 8 } }
    };
    const file: File = {
      kind: "File",
      name: ident("p"),
      declarations: [],
      imports: [fmtImport],
      unresolved: [unresolvedX, unresolvedFmt, unresolvedMissing],
      comments: [],
      scope: fileScope
    };
    const fmtPkg = NewObj(Pkg, "fmt");
    fmtPkg.Data = NewScope();

    const [pkg, err] = NewPackage(undefined, new Map([["p.go", file]]), () => [fmtPkg, undefined], NewScope());

    expect(pkg.name).toBe("p");
    expect(err?.message).toContain("undeclared name: Missing");
    expect(unresolvedX.Obj).toBe(xObj);
    expect(unresolvedFmt.Obj?.Kind).toBe(Pkg);
    expect(file.unresolved.map((id) => id.name)).toEqual(["Missing"]);
  });

  test("transliterates go/ast print helpers", () => {
    const chunks: string[] = [];
    const err = Fprint({
      write(data: string): void {
        chunks.push(data);
      }
    }, undefined, {
      kind: "Ident",
      name: "x",
      missing: undefined,
      nested: [ident("y")]
    }, NotNilFilter);

    expect(err).toBeUndefined();
    const printed = chunks.join("");
    expect(printed).toContain("Ident {");
    expect(printed).toContain("kind: \"Ident\"");
    expect(printed).toContain("name: \"x\"");
    expect(printed).toContain("nested: Array (len = 1)");
    expect(printed).not.toContain("missing:");

    const nilChunks: string[] = [];
    expect(Fprint({ write: (data: string) => nilChunks.push(data) }, undefined, undefined, NotNilFilter)).toBeUndefined();
    expect(nilChunks.join("")).toContain("nil");

    const writeErr = Fprint({
      write(): void {
        throw new Error("write failed");
      }
    }, undefined, ident("z"), NotNilFilter);
    expect(writeErr?.message).toBe("write failed");
  });
});

describe("Go-junior Go-style types", () => {
  test("builds a universe scope with predeclared types, constants, nil, and builtins", () => {
    const universe = newUniverse();

    expect(universe.scope.lookup("int64")?.kind).toBe(ObjectKind.TypeName);
    expect(universe.scope.lookup("true")?.kind).toBe(ObjectKind.Const);
    expect(universe.scope.lookup("nil")?.kind).toBe(ObjectKind.Nil);
    expect(universe.scope.lookup("panic")?.kind).toBe(ObjectKind.Builtin);
    expect(universe.scope.lookup("recover")?.kind).toBe(ObjectKind.Builtin);
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
