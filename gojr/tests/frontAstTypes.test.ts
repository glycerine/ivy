import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, test } from "./testHarness.js";
import {
  childNodes,
  commentGroupText,
  CommentMap,
  ident,
  IsExported,
  IsGenerated,
  FileExports,
  FilterFuncDuplicates,
  FilterImportDuplicates,
  Fprint,
  collapse,
  EndOf,
  filterCompositeLit,
  MergePackageFiles,
  NewCommentMap,
  NewIdent,
  NewObj,
  NewPackage,
  NewScope,
  NotNilFilter,
  ObjKind,
  ParseDirective,
  Pkg,
  SortImports,
  summary,
  Var as AstVar,
  importComment,
  importName,
  importPath,
  parseCellAddress,
  Inspect,
  Ident as AstIdent,
  PosOf,
  Preorder,
  PreorderStack,
  Unparen,
  Walk,
  walk,
  type AstNode,
  type CommentGroup,
  type CompositeLit,
  type FuncDecl,
  type ImportSpec,
  type Package,
  type File
} from "../src/go/ast/index.js";
import { parseFrontSource } from "../src/front/parser.js";
import { NewFileSet, Token as GoToken } from "../src/go/token/index.js";
import {
	  AssignableTo,
	  Byte,
	  Builtin,
  ChanDir,
  Const,
  Func as GoTypesFunc,
  Int64,
  Float64,
  Map as GoTypesMap,
  NewChan,
  NewField,
  NewFunc,
  NewInterfaceType,
  NewMap,
  NewNamed,
  NewPackage as NewGoTypesPackage,
  NewPointer,
  NewScope as NewGoTypesScope,
  NewSignatureType,
  NewSlice,
	  NewStruct,
	  NewTerm,
	  NewTuple,
	  NewTypeParam,
	  NewTypeName,
	  NewUnion,
	  NewVar,
  Nil,
  NoPos,
  String as GoTypesString,
  Typ,
  Universe,
  UntypedInt,
  UntypedNil,
  ensureUniverseInitialized,
  type Type
} from "../src/go/types/index.js";
import { TokenKind } from "../src/front/token.js";

describe("Go-junior Go-style AST", () => {
  test("keeps every upstream go/ast node struct field in the TypeScript AST", () => {
    const upstream = readFileSync("/usr/local/go1.27rc1/src/go/ast/ast.go", "utf8");
    const translated = readFileSync(join(process.cwd(), "src", "go", "ast", "index.ts"), "utf8");
    const upstreamStructs = parseGoStructFields(upstream);
    const translatedInterfaces = parseTypeScriptInterfaceFields(translated);
    const mismatches: string[] = [];

    for (const [name, fields] of upstreamStructs) {
      const translatedFields = translatedInterfaces.get(name);
      if (translatedFields === undefined) {
        mismatches.push(`${name}: missing interface`);
        continue;
      }
      const missing = fields.filter((field) => !translatedFields.has(field));
      if (missing.length > 0) {
        mismatches.push(`${name}: missing ${missing.join(", ")}`);
      }
    }

    expect(mismatches).toEqual([]);
  });

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
    expect(NewIdent("Thing").Name).toBe("Thing");
    expect(NewIdent("Thing").NamePos).toBe(0);
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

  test("parsed nodes expose standard go/ast field names", () => {
    const parsed = parseFrontSource("package main\nfunc F(x int) int { return x + 1 }\n", "std-fields.go");
    expect(parsed.file === undefined).toBe(false);
    const file = parsed.file!;
    expect(file.Name?.Name).toBe("main");
    expect(file.Imports).toEqual([]);
    expect(file.Decls).toHaveLength(1);
    expect(typeof file.Pos).toBe("function");
    expect(typeof file.End).toBe("function");

    const decl = file.Decls![0] as Extract<NonNullable<File["Decls"]>[number], { kind: "FuncDecl" }>;
    expect(decl.Name?.Name).toBe("F");
    expect(decl.Type?.Params?.List).toHaveLength(1);
    expect(decl.Type?.Results?.List).toHaveLength(1);
    expect(decl.Body?.List).toHaveLength(1);
    const returnStmt = decl.Body?.List?.[0] as { Results?: Array<{ Op?: GoToken; Y?: { Kind?: GoToken } }> };
    expect(returnStmt.Results?.[0]?.Op).toBe(GoToken.ADD);
    expect(returnStmt.Results?.[0]?.Y?.Kind).toBe(GoToken.INT);
    expect(globalThis.Object.keys(parsed.file!)).not.toContain("Decls");
  });

  test("walks and positions standard-shaped go/ast nodes", () => {
    const file = {
      kind: "File",
      Name: { kind: "Ident", NamePos: 9, Name: "main" },
      Decls: [{
        kind: "GenDecl",
        TokPos: 14,
        Tok: GoToken.VAR,
        Specs: [{
          kind: "ValueSpec",
          Names: [{ kind: "Ident", NamePos: 18, Name: "x" }],
          Type: { kind: "Ident", NamePos: 20, Name: "int" },
          Values: [{
            kind: "BasicLit",
            ValuePos: 26,
            Kind: GoToken.INT,
            Value: "3"
          }]
        }]
      }],
      Imports: [],
      Unresolved: [],
      Comments: [],
      FileStart: 1,
      Package: 1,
      FileEnd: 27
    } as unknown as File;

    const names: string[] = [];
    Inspect(file, (node) => {
      if (node?.kind === "Ident") names.push(AstIdent.String(node));
      return true;
    });
    expect(names).toEqual(["main", "x", "int"]);
    expect(childNodes(file)).toHaveLength(2);
    expect(PosOf(file)).toBe(1);
    expect(EndOf(file)).toBe(27);
    expect(PosOf(file.Name)).toBe(9);
    expect(EndOf(file.Name)).toBe(13);

    expect(EndOf({
      kind: "BasicLit",
      ValuePos: 30,
      ValueEnd: 37,
      Kind: GoToken.STRING,
      Value: "`a\r\nb`"
    } as unknown as AstNode)).toBe(37);
    expect(EndOf({
      kind: "BlockStmt",
      Lbrace: 40,
      List: [],
      Rbrace: 44
    } as unknown as AstNode)).toBe(45);
    expect(EndOf({
      kind: "BranchStmt",
      TokPos: 50,
      Tok: GoToken.BREAK
    } as unknown as AstNode)).toBe(55);
    expect(EndOf({
      kind: "ImportSpec",
      Path: {
        kind: "BasicLit",
        ValuePos: 60,
        ValueEnd: 65,
        Kind: GoToken.STRING,
        Value: "\"fmt\""
      },
      EndPos: 70
    } as unknown as AstNode)).toBe(70);
    expect(EndOf({
      kind: "File",
      Name: { kind: "Ident", NamePos: 80, Name: "main" },
      Decls: [],
      FileEnd: 999
    } as unknown as AstNode)).toBe(84);
    expect(childNodes({
      kind: "Package",
      Name: "p",
      Files: new Map([
        ["a.go", file]
      ])
    } as unknown as AstNode)).toEqual([file]);
  });

  test("Walk follows standard go/ast File children", () => {
    const doc: CommentGroup = {
      kind: "CommentGroup",
      List: [{ kind: "Comment", Text: "// package doc", text: "// package doc" }],
      list: [{ kind: "Comment", Text: "// package doc", text: "// package doc" }]
    };
    const imp: ImportSpec = {
      kind: "ImportSpec",
      Path: { kind: "BasicLit", ValuePos: 30, Kind: GoToken.STRING, Value: "\"fmt\"", token: TokenKind.StringLiteral, value: "\"fmt\"" },
      path: { kind: "BasicLit", ValuePos: 30, Kind: GoToken.STRING, Value: "\"fmt\"", token: TokenKind.StringLiteral, value: "\"fmt\"" }
    };
    const unresolved = { kind: "Ident" as const, NamePos: 90, Name: "Missing" };
    const looseComment: CommentGroup = {
      kind: "CommentGroup",
      List: [{ kind: "Comment", Text: "// loose", text: "// loose" }],
      list: [{ kind: "Comment", Text: "// loose", text: "// loose" }]
    };
    const file = {
      kind: "File" as const,
      Doc: doc,
      Package: 1,
      Name: { kind: "Ident" as const, NamePos: 9, Name: "p" },
      Decls: [{
        kind: "GenDecl" as const,
        TokPos: 20,
        Tok: GoToken.IMPORT,
        Specs: [imp]
      }],
      Imports: [imp],
      Unresolved: [unresolved],
      Comments: [looseComment]
    } as unknown as File;

    const seen: string[] = [];
    Walk({
      Visit(node: AstNode | undefined) {
        if (node) seen.push(node.kind);
        return this;
      }
    }, file);

    expect(seen).toContain("CommentGroup");
    expect(seen).toContain("GenDecl");
    expect(seen).toContain("ImportSpec");
    expect(seen).not.toContain("Missing");
    expect(seen.filter((kind) => kind === "CommentGroup")).toHaveLength(1);
    expect(childNodes(file).map((node) => node.kind)).toEqual(["CommentGroup", "Ident", "GenDecl"]);
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

  test("transliterates go/ast comment map helpers", () => {
    const target = ident("x", { filename: "commentmap.go", offset: 10, length: 1, line: 1, column: 11 });
    const replacement = ident("y", { filename: "commentmap.go", offset: 40, length: 1, line: 4, column: 1 });
    const trailing: CommentGroup = {
      kind: "CommentGroup",
      list: [{
        kind: "Comment",
        text: "// trailing",
        span: { filename: "commentmap.go", offset: 12, length: 11, line: 1, column: 13 }
      }]
    };
    const earlier: CommentGroup = {
      kind: "CommentGroup",
      list: [{
        kind: "Comment",
        text: "/* earlier\ncomment */",
        span: { filename: "commentmap.go", offset: 2, length: 21, line: 1, column: 3 }
      }]
    };

    const cmap = NewCommentMap(undefined, target, [trailing]);
    expect(cmap?.get(target)).toEqual([trailing]);
    expect(cmap?.Comments()).toEqual([trailing]);
    expect(cmap?.Update(target, replacement)).toBe(replacement);
    expect(cmap?.get(target)).toBeUndefined();
    expect(cmap?.get(replacement)).toEqual([trailing]);

    const manual = new CommentMap();
    manual.addComment(replacement, trailing);
    manual.addComment(replacement, earlier);
    expect(manual.Comments()).toEqual([earlier, trailing]);
    expect(manual.Filter(replacement).get(replacement)).toEqual([trailing, earlier]);
    expect(summary([earlier])).toBe("/* earlier comment */");
    expect(manual.String()).toContain("CommentMap");
    expect(NewCommentMap(undefined, target, [])).toBeUndefined();
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
    expect(typeDecl.specs[0].type.Incomplete).toBe(true);

    const lit: CompositeLit = {
      kind: "CompositeLit",
      elements: [
        { kind: "KeyValueExpr", key: ident("Hidden"), value: ident("x") },
        { kind: "KeyValueExpr", key: ident("Public"), value: ident("y") }
      ]
    };
    filterCompositeLit(lit, (name) => name === "Public", true);
    expect(lit.elements).toHaveLength(1);
    expect(lit.Incomplete).toBe(true);

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
          imports: [
            { kind: "ImportSpec", path: { kind: "BasicLit", token: TokenKind.StringLiteral, value: "\"fmt\"" } },
            { kind: "ImportSpec", path: { kind: "BasicLit", token: TokenKind.StringLiteral, value: "\"fmt\"" } }
          ],
          unresolved: [],
          comments: []
        }
      ]
    }, FilterFuncDuplicates | FilterImportDuplicates);

    expect(merged.declarations.filter((decl) => decl.kind === "FuncDecl" && decl.name.name === "Exported")).toHaveLength(1);
    expect(merged.imports.map((spec) => spec.path.value)).toEqual(["\"fmt\""]);
  });

  test("MergePackageFiles preserves package docs, file range, comments, and filename order", () => {
    const docA: CommentGroup = { kind: "CommentGroup", List: [{ kind: "Comment", Text: "// doc a", text: "// doc a" }], list: [{ kind: "Comment", Text: "// doc a", text: "// doc a" }] };
    const docB: CommentGroup = { kind: "CommentGroup", List: [{ kind: "Comment", Text: "// doc b", text: "// doc b" }], list: [{ kind: "Comment", Text: "// doc b", text: "// doc b" }] };
    const commentA: CommentGroup = { kind: "CommentGroup", List: [{ kind: "Comment", Text: "// comment a", text: "// comment a" }], list: [{ kind: "Comment", Text: "// comment a", text: "// comment a" }] };
    const commentB: CommentGroup = { kind: "CommentGroup", List: [{ kind: "Comment", Text: "// comment b", text: "// comment b" }], list: [{ kind: "Comment", Text: "// comment b", text: "// comment b" }] };
    const fmtImport: ImportSpec = {
      kind: "ImportSpec",
      Path: { kind: "BasicLit", Value: "\"fmt\"", token: TokenKind.StringLiteral, value: "\"fmt\"" },
      path: { kind: "BasicLit", Value: "\"fmt\"", token: TokenKind.StringLiteral, value: "\"fmt\"" }
    };
    const stringsImport: ImportSpec = {
      kind: "ImportSpec",
      Path: { kind: "BasicLit", Value: "\"strings\"", token: TokenKind.StringLiteral, value: "\"strings\"" },
      path: { kind: "BasicLit", Value: "\"strings\"", token: TokenKind.StringLiteral, value: "\"strings\"" }
    };
    const declA: FuncDecl = {
      kind: "FuncDecl",
      Name: { kind: "Ident", Name: "A", NamePos: 40 },
      Type: { kind: "FuncType", Params: { kind: "FieldList", List: [] } },
      name: ident("A"),
      type: { kind: "FuncType", params: { kind: "FieldList", fields: [] } }
    } as unknown as FuncDecl;
    const declB: FuncDecl = {
      kind: "FuncDecl",
      Name: { kind: "Ident", Name: "B", NamePos: 20 },
      Type: { kind: "FuncType", Params: { kind: "FieldList", List: [] } },
      name: ident("B"),
      type: { kind: "FuncType", params: { kind: "FieldList", fields: [] } }
    } as unknown as FuncDecl;
    const merged = MergePackageFiles({
      kind: "Package",
      Name: "p",
      Files: new Map([
        ["b.go", {
          kind: "File",
          Doc: docB,
          Package: 10,
          Name: { kind: "Ident", Name: "p", NamePos: 18 },
          Decls: [declB],
          FileStart: 10,
          FileEnd: 30,
          Imports: [stringsImport, fmtImport],
          Comments: [commentB]
        } as unknown as File],
        ["a.go", {
          kind: "File",
          Doc: docA,
          Package: 50,
          Name: { kind: "Ident", Name: "p", NamePos: 58 },
          Decls: [declA],
          FileStart: 5,
          FileEnd: 90,
          Imports: [fmtImport],
          Comments: [commentA]
        } as unknown as File]
      ])
    } as unknown as Package, FilterImportDuplicates);

    expect(merged.Package).toBe(50);
    expect(merged.FileStart).toBe(5);
    expect(merged.FileEnd).toBe(90);
    expect(merged.Doc?.List?.map((comment) => comment.Text)).toEqual(["// doc a", "//", "// doc b"]);
    expect(merged.Decls?.map((decl) => decl.kind === "FuncDecl" ? decl.Name?.Name : "")).toEqual(["A", "B"]);
    expect(merged.Imports?.map((spec) => spec.Path?.Value)).toEqual(["\"fmt\"", "\"strings\""]);
    expect(merged.Comments).toEqual([commentA, commentB]);
    expect(merged.Name?.Name).toBe("p");
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

  test("SortImports uses translated token.FileSet line operations", () => {
    const source = "package p\nimport (\n\t\"z\"\n\t\"a\"\n\t\"z\"\n)\n";
    const fset = NewFileSet();
    const tokenFile = fset.AddFile("imports.go", -1, source.length);
    tokenFile.SetLinesForContent(source);
    const lit = (value: string, nth: number): ImportSpec => {
      let at = -1;
      let from = -1;
      for (let i = 0; i <= nth; i += 1) {
        from = source.indexOf(value, from + 1);
        at = from;
      }
      const start = tokenFile.Pos(at);
      const path = {
        kind: "BasicLit" as const,
        ValuePos: start,
        ValueEnd: start + value.length,
        Kind: GoToken.STRING,
        Value: value,
        token: TokenKind.StringLiteral as const,
        value
      };
      return {
        kind: "ImportSpec",
        Path: path,
        path
      };
    };
    const z1 = lit("\"z\"", 0);
    const a = lit("\"a\"", 0);
    const z2 = lit("\"z\"", 1);
    const decl = {
      kind: "GenDecl" as const,
      TokPos: tokenFile.Pos(source.indexOf("import")),
      Tok: GoToken.IMPORT,
      Lparen: tokenFile.Pos(source.indexOf("(")),
      Specs: [z1, a, z2],
      Rparen: tokenFile.Pos(source.lastIndexOf(")"))
    };
    const file = {
      kind: "File" as const,
      Package: tokenFile.Pos(source.indexOf("package")),
      Name: { kind: "Ident" as const, NamePos: tokenFile.Pos(source.indexOf("p")), Name: "p" },
      Decls: [decl],
      Imports: [z1, a, z2],
      Unresolved: [],
      Comments: []
    } as unknown as File;
    const beforeLines = tokenFile.LineCount();

    SortImports(fset, file);

    expect(decl.Specs?.map((spec) => spec.kind === "ImportSpec" ? importPath(spec) : "")).toEqual(["a", "z"]);
    expect(file.Imports?.map(importPath)).toEqual(["a", "z"]);
    expect(beforeLines - tokenFile.LineCount()).toBeGreaterThan(0);
    expect(a.Path?.ValuePos).toBe(z1.Path?.ValuePos);
    expect(z2.EndPos).toBeDefined();
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

function parseGoStructFields(source: string): Map<string, string[]> {
  const stripped = stripGoComments(source);
  const structs = new Map<string, string[]>();
  for (const block of stripped.matchAll(/^type\s*\(([\s\S]*?)^\)/gm)) {
    for (const match of block[1]!.matchAll(/^\s*(\w+)\s+struct\s*\{([\s\S]*?)^\s*\}/gm)) {
      structs.set(match[1]!, parseGoStructBodyFields(match[2]!));
    }
  }
  for (const match of stripped.matchAll(/^type\s+(\w+)\s+struct\s*\{([\s\S]*?)^\}/gm)) {
    structs.set(match[1]!, parseGoStructBodyFields(match[2]!));
  }
  return structs;
}

function parseGoStructBodyFields(body: string): string[] {
  const fields: string[] = [];
  for (let line of body.split("\n")) {
    line = line.trim().replace(/`[^`]*`/g, "").trim();
    if (line === "") continue;
    const match = /^([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s+/.exec(line);
    if (match) {
      fields.push(...match[1]!.split(/\s*,\s*/));
    }
  }
  return fields;
}

function parseTypeScriptInterfaceFields(source: string): Map<string, Set<string>> {
  const interfaces = new Map<string, Set<string>>();
  for (const match of source.matchAll(/^export\s+interface\s+(\w+)\s+(?:extends\s+\w+\s+)?\{([\s\S]*?)^\}/gm)) {
    const fields = new Set<string>();
    for (const line of match[2]!.split("\n")) {
      const field = /^\s*([A-Za-z_]\w*)\??\s*:/.exec(line);
      if (field) fields.add(field[1]!);
    }
    interfaces.set(match[1]!, fields);
  }
  return interfaces;
}

function stripGoComments(source: string): string {
  return source.replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/.*$/gm, "");
}

  describe("Go-junior Go-style types", () => {
    test("builds a universe scope with predeclared types, constants, nil, and builtins", () => {
    ensureUniverseInitialized();

    expect(Universe.Lookup("int64")?.constructor.name).toBe("TypeName");
    expect(Universe.Lookup("true")).toBeInstanceOf(Const);
    expect(Universe.Lookup("nil")).toBeInstanceOf(Nil);
    expect(Universe.Lookup("panic")).toBeInstanceOf(Builtin);
    expect(Universe.Lookup("recover")).toBeInstanceOf(Builtin);
    expect(Universe.Names()).toContain("append");
  });

  test("supports nested scopes and duplicate detection", () => {
    ensureUniverseInitialized();
    const child = NewGoTypesScope(Universe, NoPos, NoPos, "function");
    const first = NewVar(NoPos, null, "x", Typ[Int64]!);
    const duplicate = NewVar(NoPos, null, "x", Typ[GoTypesString]!);

    expect(child.Insert(first)).toBeNull();
    expect(child.Insert(duplicate)).toBe(first);
    expect(child.LookupParent("x", NoPos)[1]).toBe(first);
    expect(child.LookupParent("string", NoPos)[1]?.constructor.name).toBe("TypeName");
  });

	  test("formats signatures and compound types", () => {
    ensureUniverseInitialized();
    const signature = NewSignatureType(
      null,
      null,
      null,
      NewTuple(NewVar(NoPos, null, "format", Typ[GoTypesString]!), NewVar(NoPos, null, "args", NewSlice(Universe.Lookup("any")!.Type()!))),
      NewTuple(NewVar(NoPos, null, "", Typ[Int64]!), NewVar(NoPos, null, "", Universe.Lookup("error")!.Type()!)),
      true
    );

    expect(signature.String()).toBe("func(format string, args ...any) (int64, error)");
    expect(NewMap(Typ[GoTypesString]!, Typ[Int64]!).String()).toBe("map[string]int64");
    expect(NewChan(ChanDir.SendRecv, Typ[Int64]!).String()).toBe("chan int64");
    expect(NewChan(ChanDir.SendOnly, Typ[GoTypesString]!).String()).toBe("chan<- string");
	    expect(NewChan(ChanDir.RecvOnly, Typ[Float64]!).String()).toBe("<-chan float64");
	  });

	  test("validates variadic signatures over type parameter type sets", () => {
	    ensureUniverseInitialized();
	    const pkg = NewGoTypesPackage("workbook/generic", "generic");
	    const stringOrBytes = NewInterfaceType(null, [
	      NewUnion([
	        NewTerm(false, Typ[GoTypesString]!),
	        NewTerm(false, NewSlice(Typ[Byte]!))
	      ])
	    ]).Complete();
	    const t = NewTypeParam(NewTypeName(NoPos, pkg, "T", null), stringOrBytes);

	    expect(() => NewSignatureType(
	      null,
	      null,
	      [t],
	      NewTuple(NewVar(NoPos, pkg, "x", t)),
	      null,
	      true
	    )).not.toThrow();

	    const intParam = NewTypeParam(NewTypeName(NoPos, pkg, "I", null), Typ[Int64]!);
	    expect(() => NewSignatureType(
	      null,
	      null,
	      [intParam],
	      NewTuple(NewVar(NoPos, pkg, "x", intParam)),
	      null,
	      true
	    )).toThrow(/variadic parameter of slice or string type/);
	  });

	  test("checks assignability for untyped constants, nil, named types, and method-set interfaces", () => {
    ensureUniverseInitialized();
    const pkg = NewGoTypesPackage("workbook/geom", "geom");
    const pointName = NewTypeName(NoPos, pkg, "Point", null);
    const point = NewNamed(pointName, NewStruct([
      NewField(NoPos, pkg, "X", Typ[Float64]!, false),
      NewField(NoPos, pkg, "Y", Typ[Float64]!, false)
    ], null), null);

    const lenSig = NewSignatureType(
      NewVar(NoPos, pkg, "p", point),
      null,
      null,
      null,
      NewTuple(NewVar(NoPos, pkg, "", Typ[Float64]!)),
      false
    );
    const scaleSig = NewSignatureType(
      NewVar(NoPos, pkg, "p", NewPointer(point)),
      null,
      null,
      NewTuple(NewVar(NoPos, pkg, "k", Typ[Float64]!)),
      null,
      false
    );
    point.AddMethod(NewFunc(NoPos, pkg, "Len2", lenSig));
    point.AddMethod(NewFunc(NoPos, pkg, "Scale", scaleSig));

    const hasLen = NewInterfaceType([NewFunc(NoPos, pkg, "Len2", NewSignatureType(
      null,
      null,
      null,
      null,
      NewTuple(NewVar(NoPos, pkg, "", Typ[Float64]!)),
      false
    ))], null).Complete();
    const hasScale = NewInterfaceType([NewFunc(NoPos, pkg, "Scale", NewSignatureType(
      null,
      null,
      null,
      NewTuple(NewVar(NoPos, pkg, "k", Typ[Float64]!)),
      null,
      false
    ))], null).Complete();

    expect(AssignableTo(Typ[UntypedInt]!, Typ[Int64]!)).toBe(true);
    expect(AssignableTo(Typ[UntypedNil]!, NewSlice(Typ[GoTypesString]!))).toBe(true);
    expect(AssignableTo(point, hasLen)).toBe(true);
    expect(AssignableTo(point, hasScale)).toBe(false);
    expect(AssignableTo(NewPointer(point), hasScale)).toBe(true);
  });
});
