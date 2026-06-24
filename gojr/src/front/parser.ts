import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan, diagnosticFilename } from "../diagnostics.js";
import {
  ArrayType,
  AssignStmt,
  AstNode,
  Bad,
  BasicLit,
  BinaryOperator,
  BlockStmt,
  BranchStmt,
  CallExpr,
  CaseClause,
  CellAddress,
  CompositeLit,
  Decl,
  DeferStmt,
  Expr,
  Field,
  FieldList,
  File,
  FuncDecl,
  FuncLit,
  FuncType,
  Fun,
  GenDecl,
  Ident,
  ImportSpec,
  KeyValueExpr,
  Lbl,
  NewObj,
  NewScope,
  NormalizeAst,
  Object as AstObject,
  Package,
  Pos,
  Scope,
  parseCellAddress,
  RangeStmt,
  Spec,
  Stmt,
  Typ,
  TypeSpec,
  Unparen,
  ValueSpec,
  Var,
  Visitor,
  Walk,
  Con,
  ident
} from "./ast.js";
import { scanSource } from "./scanner.js";
import { FrontToken, isAssignmentToken, isIdentifierLike, TokenKind } from "./token.js";

export interface ParseFrontResult {
  file?: File;
  statements: Stmt[];
  diagnostics: Diagnostic[];
  tokens: FrontToken[];
}

export interface ParseFrontFilesResult {
  files: File[];
  statements: Stmt[];
  diagnostics: Diagnostic[];
  results: ParseFrontResult[];
}

type SimpleStmtMode = "basic" | "labelOk" | "rangeOk";

export const basic: SimpleStmtMode = "basic";
export const labelOk: SimpleStmtMode = "labelOk";
export const rangeOk: SimpleStmtMode = "rangeOk";
export const maxNestLev = 1e5;

export const declStart = new Set<TokenKind>([TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func]);
export const stmtStart = new Set<TokenKind>([
  TokenKind.Break,
  TokenKind.Const,
  TokenKind.Continue,
  TokenKind.Defer,
  TokenKind.Fallthrough,
  TokenKind.For,
  TokenKind.Go,
  TokenKind.Goto,
  TokenKind.If,
  TokenKind.Return,
  TokenKind.Select,
  TokenKind.Switch,
  TokenKind.Type,
  TokenKind.Var
]);
export const exprEnd = new Set<TokenKind>([TokenKind.Comma, TokenKind.Semicolon, TokenKind.Colon, TokenKind.RParen, TokenKind.RBracket, TokenKind.RBrace]);

export class bailout {
  public constructor(
    public pos?: SourceSpan,
    public msg = ""
  ) {}
}

export interface field {
  name?: Ident;
  typ?: Expr;
}

export type parseSpecFunction = (doc: unknown, keyword: TokenKind, iota: number) => Spec;

export function assert(condition: boolean, message = "assertion failed"): void {
  if (!condition) {
    throw new Error(message);
  }
}

export function trace(p: parser, msg: string): parser {
  p.printTrace(msg, "(");
  p.indent += 1;
  return p;
}

export function un(p: parser): void {
  p.indent -= 1;
  p.printTrace(")");
}

export function incNestLev(p: parser): parser {
  p.nestLev += 1;
  if (p.nestLev > maxNestLev) {
    p.error("exceeded max nesting depth", p.currentSpan());
    throw new bailout();
  }
  return p;
}

export function decNestLev(p: parser): void {
  p.nestLev -= 1;
}

export function isTypeSwitchAssert(x: Expr): boolean {
  return x.kind === "TypeAssertExpr" && x.typeSwitch;
}

export function packIndexExpr(x: Expr, _lbrack: SourceSpan | undefined, exprs: Expr[], rbrack: SourceSpan | undefined): Expr {
  switch (exprs.length) {
    case 0:
      throw new Error("internal error: packIndexExpr with empty expr slice");
    case 1:
      return {
        kind: "IndexExpr",
        object: x,
        index: exprs[0]!,
        span: mergeSpans(x.span, rbrack ?? exprs[0]?.span)
      };
    default:
      return {
        kind: "IndexListExpr",
        object: x,
        indices: exprs,
        span: mergeSpans(x.span, rbrack ?? exprs[exprs.length - 1]?.span)
      };
  }
}

export type Mode = number;

export const PackageClauseOnly: Mode = 1 << 0;
export const ImportsOnly: Mode = 1 << 1;
export const ParseComments: Mode = 1 << 2;
export const Trace: Mode = 1 << 3;
export const DeclarationErrors: Mode = 1 << 4;
export const SpuriousErrors: Mode = 1 << 5;
export const SkipObjectResolution: Mode = 1 << 6;
export const AllErrors: Mode = SpuriousErrors;

export const debugResolve = false;
export const maxScopeDepth = 1e3;
export const unresolved = NewObj(Bad, "");

type declarationErrorHandler = (pos: Pos, message: string) => void;

export function resolveFile(file: File, handle: unknown, declErr?: declarationErrorHandler): void {
  const pkgScope = NewScope();
  const r = new resolver(handle, declErr, pkgScope, pkgScope, 1);

  for (const decl of file.declarations) {
    Walk(r, decl);
  }

  r.closeScope();
  assert(r.topScope === undefined, "unbalanced scopes");
  assert(r.labelScope === undefined, "unbalanced label scopes");

  let i = 0;
  for (const identNode of r.unresolved) {
    assert(identNode.Obj === unresolved, "object already resolved");
    const obj = r.pkgScope.Lookup(identNode.name);
    if (obj === undefined) {
      delete identNode.Obj;
      r.unresolved[i] = identNode;
      i += 1;
    } else if (debugResolve) {
      identNode.Obj = obj;
      r.trace("resolved %s@%v to package object %v", identNode.name, identNode.span?.offset ?? 0, identNode.Obj.Pos());
    } else {
      identNode.Obj = obj;
    }
  }
  file.scope = r.pkgScope;
  file.unresolved = r.unresolved.slice(0, i);
}

export class resolver implements Visitor {
  public unresolved: Ident[] = [];
  public labelScope: Scope | undefined;
  public targetStack: Ident[][] = [];

  public constructor(
    public handle: unknown,
    public declErr: declarationErrorHandler | undefined,
    public topScope: Scope | undefined,
    public pkgScope: Scope,
    public depth: number
  ) {}

  public trace(format: string, ...args: unknown[]): void {
    globalThis.console?.log(`${". ".repeat(this.depth)}${this.sprintf(format, ...args)}`);
  }

  public sprintf(format: string, ...args: unknown[]): string {
    return resolverSprintf(format, args);
  }

  public openScope(pos: Pos): void {
    this.depth += 1;
    if (this.depth > maxScopeDepth) {
      throw new bailout(posSpanForResolver(this.handle, pos), "exceeded max scope depth during object resolution");
    }
    if (debugResolve) {
      this.trace("opening scope @%v", pos);
    }
    this.topScope = NewScope(this.topScope);
  }

  public closeScope(): void {
    this.depth -= 1;
    if (debugResolve) {
      this.trace("closing scope");
    }
    this.topScope = this.topScope?.Outer;
  }

  public openLabelScope(): void {
    this.labelScope = NewScope(this.labelScope);
    this.targetStack.push([]);
  }

  public closeLabelScope(): void {
    const n = this.targetStack.length - 1;
    const scope = this.labelScope;
    for (const identNode of this.targetStack[n] ?? []) {
      const obj = scope?.Lookup(identNode.name);
      if (obj === undefined && this.declErr) {
        delete identNode.Obj;
        this.declErr(identPos(identNode), `label ${identNode.name} undefined`);
      } else if (obj !== undefined) {
        identNode.Obj = obj;
      }
    }
    this.targetStack.length = Math.max(0, n);
    this.labelScope = this.labelScope?.Outer;
  }

  public declare(decl: unknown, data: unknown, scope: Scope | undefined, kind: number, ...idents: Ident[]): void {
    if (scope === undefined) {
      return;
    }
    for (const identNode of idents) {
      if (identNode.Obj !== undefined) {
        throw new globalThis.Error(`${identPos(identNode)}: identifier ${identNode.name} already declared or resolved`);
      }
      const obj = NewObj(kind, identNode.name);
      obj.Decl = decl;
      obj.Data = data;
      if (!isIdentNode(decl)) {
        identNode.Obj = obj;
      }
      if (identNode.name !== "_") {
        if (debugResolve) {
          this.trace("declaring %s@%v", identNode.name, identPos(identNode));
        }
        const alt = scope.Insert(obj);
        if (alt !== undefined && this.declErr) {
          let prevDecl = "";
          const pos = alt.Pos();
          if (pos !== 0) {
            prevDecl = this.sprintf("\n\tprevious declaration at %v", pos);
          }
          this.declErr(identPos(identNode), `${identNode.name} redeclared in this block${prevDecl}`);
        }
      }
    }
  }

  public shortVarDecl(decl: AssignStmt): void {
    let n = 0;
    for (const x of decl.lhs) {
      if (x.kind === "Ident") {
        assert(x.Obj === undefined, "identifier already declared or resolved");
        const obj = NewObj(Var, x.name);
        obj.Decl = decl;
        x.Obj = obj;
        if (x.name !== "_") {
          if (debugResolve) {
            this.trace("declaring %s@%v", x.name, identPos(x));
          }
          const alt = this.topScope?.Insert(obj);
          if (alt !== undefined) {
            x.Obj = alt;
          } else {
            n += 1;
          }
        }
      }
    }
    if (n === 0 && this.declErr && decl.lhs[0]) {
      this.declErr(nodePosForResolver(decl.lhs[0]), "no new variables on left side of :=");
    }
  }

  public resolve(identNode: Ident, collectUnresolved: boolean): void {
    if (identNode.Obj !== undefined) {
      throw new globalThis.Error(this.sprintf("%v: identifier %s already declared or resolved", identPos(identNode), identNode.name));
    }
    if (identNode.name === "_") {
      return;
    }
    for (let s = this.topScope; s !== undefined; s = s.Outer) {
      const obj = s.Lookup(identNode.name);
      if (obj !== undefined) {
        if (debugResolve) {
          this.trace("resolved %v:%s to %v", identPos(identNode), identNode.name, obj.Name);
        }
        assert(obj.Name !== "", "obj with no name");
        if (!isIdentNode(obj.Decl)) {
          identNode.Obj = obj;
        }
        return;
      }
    }
    if (collectUnresolved) {
      identNode.Obj = unresolved;
      this.unresolved.push(identNode);
    }
  }

  public walkExprs(list: Expr[]): void {
    for (const node of list) {
      Walk(this, node);
    }
  }

  public walkLHS(list: Expr[]): void {
    for (const expr of list) {
      const node = Unparen(expr);
      if (node.kind !== "Ident") {
        Walk(this, node);
      }
    }
  }

  public walkStmts(list: Stmt[]): void {
    for (const stmt of list) {
      Walk(this, stmt);
    }
  }

  public Visit(node: AstNode | undefined): Visitor | undefined {
    if (debugResolve && node !== undefined) {
      this.trace("node %s@%v", node.kind, nodePosForResolver(node));
    }

    if (node === undefined) {
      return undefined;
    }

    switch (node.kind) {
      case "Ident":
        this.resolve(node, true);
        break;

      case "FuncLit":
        this.openScope(nodePosForResolver(node));
        this.walkFuncType(node.type);
        this.walkBody(node.body);
        this.closeScope();
        break;

      case "SelectorExpr":
        Walk(this, node.object);
        break;

      case "StructType":
        this.openScope(nodePosForResolver(node));
        this.walkFieldList(node.fields, Var);
        this.closeScope();
        break;

      case "FuncType":
        this.openScope(nodePosForResolver(node));
        this.walkFuncType(node);
        this.closeScope();
        break;

      case "CompositeLit":
        if (node.type !== undefined) {
          Walk(this, node.type);
        }
        for (const element of node.elements) {
          if (element.kind === "KeyValueExpr") {
            if (element.key.kind === "Ident") {
              this.resolve(element.key, false);
            } else {
              Walk(this, element.key);
            }
            Walk(this, element.value);
          } else {
            Walk(this, element);
          }
        }
        break;

      case "InterfaceType":
        this.openScope(nodePosForResolver(node));
        this.walkFieldList(node.methods, Fun);
        this.closeScope();
        break;

      case "LabeledStmt":
        this.declare(node, undefined, this.labelScope, Lbl, node.label);
        Walk(this, node.stmt);
        break;

      case "AssignStmt":
        this.walkExprs(node.rhs);
        if (node.token === TokenKind.Define) {
          this.shortVarDecl(node);
        } else {
          this.walkExprs(node.lhs);
        }
        break;

      case "BranchStmt":
        if (node.token !== TokenKind.Fallthrough && node.label !== undefined && this.targetStack.length > 0) {
          this.targetStack[this.targetStack.length - 1]?.push(node.label);
        }
        break;

      case "BlockStmt":
        this.openScope(nodePosForResolver(node));
        this.walkStmts(node.statements);
        this.closeScope();
        break;

      case "IfStmt":
        this.openScope(nodePosForResolver(node));
        if (node.init !== undefined) {
          Walk(this, node.init);
        }
        Walk(this, node.condition);
        Walk(this, node.body);
        if (node.else !== undefined) {
          Walk(this, node.else);
        }
        this.closeScope();
        break;

      case "CaseClause":
        this.walkExprs(node.list);
        this.openScope(nodePosForResolver(node));
        this.walkStmts(node.body);
        this.closeScope();
        break;

      case "SwitchStmt":
        this.openScope(nodePosForResolver(node));
        if (node.init !== undefined) {
          Walk(this, node.init);
        }
        if (node.tag !== undefined) {
          if (node.init !== undefined) {
            this.openScope(nodePosForResolver(node.tag));
            Walk(this, node.tag);
            this.closeScope();
          } else {
            Walk(this, node.tag);
          }
        }
        this.walkStmts(node.body);
        this.closeScope();
        break;

      case "TypeSwitchStmt":
        if (node.init !== undefined) {
          this.openScope(nodePosForResolver(node));
          Walk(this, node.init);
          this.closeScope();
        }
        this.openScope(nodePosForResolver(node.assign));
        Walk(this, node.assign);
        this.walkStmts(node.body);
        this.closeScope();
        break;

      case "CommClause":
        this.openScope(nodePosForResolver(node));
        if (node.comm !== undefined) {
          Walk(this, node.comm);
        }
        this.walkStmts(node.body);
        this.closeScope();
        break;

      case "SelectStmt":
        this.walkStmts(node.body);
        break;

      case "ForStmt":
        this.openScope(nodePosForResolver(node));
        if (node.init !== undefined) {
          Walk(this, node.init);
        }
        if (node.condition !== undefined) {
          Walk(this, node.condition);
        }
        if (node.post !== undefined) {
          Walk(this, node.post);
        }
        Walk(this, node.body);
        this.closeScope();
        break;

      case "RangeStmt": {
        this.openScope(nodePosForResolver(node));
        Walk(this, node.source);
        const lhs = [node.key, node.value].filter((expr): expr is Expr => expr !== undefined);
        if (lhs.length > 0) {
          if (node.token === TokenKind.Define) {
            const as: AssignStmt = {
              kind: "AssignStmt",
              lhs,
              token: TokenKind.Define,
              rhs: [node.source],
              ...(node.span ? { span: node.span } : {})
            };
            this.walkLHS(lhs);
            this.shortVarDecl(as);
          } else {
            this.walkExprs(lhs);
          }
        }
        Walk(this, node.body);
        this.closeScope();
        break;
      }

      case "GenDecl":
        switch (node.token) {
          case TokenKind.Const:
          case TokenKind.Var:
            for (let i = 0; i < node.specs.length; i += 1) {
              const spec = node.specs[i];
              if (spec?.kind !== "ValueSpec") {
                continue;
              }
              const kind = node.token === TokenKind.Var ? Var : Con;
              this.walkExprs(spec.values);
              if (spec.type !== undefined) {
                Walk(this, spec.type);
              }
              this.declare(spec, i, this.topScope, kind, ...spec.names);
            }
            break;
          case TokenKind.Type:
            for (const spec of node.specs) {
              if (spec.kind !== "TypeSpec") {
                continue;
              }
              this.declare(spec, undefined, this.topScope, Typ, spec.name);
              if (spec.typeParams !== undefined) {
                this.openScope(nodePosForResolver(spec));
                this.walkTParams(spec.typeParams);
                this.closeScope();
              }
              Walk(this, spec.type);
            }
            break;
        }
        break;

      case "FuncDecl":
        this.openScope(nodePosForResolver(node));
        this.walkRecv(node.receiver);
        if (node.type.typeParams !== undefined) {
          this.walkTParams(node.type.typeParams);
        }
        this.resolveList(node.type.params);
        this.resolveList(node.type.results);
        this.declareList(node.receiver, Var);
        this.declareList(node.type.params, Var);
        this.declareList(node.type.results, Var);
        this.walkBody(node.body);
        if (node.receiver === undefined && node.name.name !== "init") {
          this.declare(node, undefined, this.pkgScope, Fun, node.name);
        }
        this.closeScope();
        break;

      default:
        return this;
    }

    return undefined;
  }

  public walkFuncType(typ: FuncType): void {
    this.resolveList(typ.params);
    this.resolveList(typ.results);
    this.declareList(typ.params, Var);
    this.declareList(typ.results, Var);
  }

  public resolveList(list: FieldList | undefined): void {
    if (list === undefined) {
      return;
    }
    for (const field of list.fields) {
      Walk(this, field.type);
    }
  }

  public declareList(list: FieldList | undefined, kind: number): void {
    if (list === undefined) {
      return;
    }
    for (const field of list.fields) {
      this.declare(field, undefined, this.topScope, kind, ...field.names);
    }
  }

  public walkRecv(recv: FieldList | undefined): void {
    if (recv === undefined || recv.fields.length === 0) {
      return;
    }
    let typ = recv.fields[0]?.type;
    if (typ?.kind === "StarExpr") {
      typ = typ.expr;
    }

    let declareExprs: Expr[] = [];
    let resolveExprs: Expr[] = [];
    switch (typ?.kind) {
      case "IndexExpr":
        declareExprs = [typ.index];
        resolveExprs.push(typ.object);
        break;
      case "IndexListExpr":
        declareExprs = typ.indices;
        resolveExprs.push(typ.object);
        break;
      default:
        if (typ !== undefined) {
          resolveExprs.push(typ);
        }
        break;
    }

    for (const expr of declareExprs) {
      if (expr.kind === "Ident") {
        this.declare(expr, undefined, this.topScope, Typ, expr);
      } else {
        resolveExprs.push(expr);
      }
    }
    for (const expr of resolveExprs) {
      Walk(this, expr);
    }
    for (const field of recv.fields.slice(1)) {
      Walk(this, field.type);
    }
  }

  public walkFieldList(list: FieldList | undefined, kind: number): void {
    if (list === undefined) {
      return;
    }
    this.resolveList(list);
    this.declareList(list, kind);
  }

  public walkTParams(list: FieldList): void {
    this.declareList(list, Typ);
    this.resolveList(list);
  }

  public walkBody(body: BlockStmt | undefined): void {
    if (body === undefined) {
      return;
    }
    this.openLabelScope();
    this.walkStmts(body.statements);
    this.closeLabelScope();
  }
}

function resolverSprintf(format: string, args: unknown[]): string {
  let i = 0;
  return format.replace(/%[sv]/g, () => String(args[i++]));
}

function posSpanForResolver(handle: unknown, pos: Pos): SourceSpan {
  const position = (handle as { Position?: (pos: Pos) => { filename?: string; Filename?: string; offset?: number; Offset?: number; line?: number; Line?: number; column?: number; Column?: number } } | undefined)?.Position?.(pos);
  return {
    filename: position?.filename ?? position?.Filename ?? REPL_FILENAME,
    offset: position?.offset ?? position?.Offset ?? pos,
    length: 0,
    line: position?.line ?? position?.Line ?? 1,
    column: position?.column ?? position?.Column ?? 1
  };
}

function nodePosForResolver(node: AstNode): Pos {
  return node.span?.offset ?? 0;
}

function identPos(node: Ident): Pos {
  return node.span?.offset ?? 0;
}

function isIdentNode(value: unknown): value is Ident {
  return typeof value === "object" && value !== null && "kind" in value && value.kind === "Ident";
}

interface SimpleStmtResult {
  statement: Stmt;
  isRange: boolean;
}

interface ParamDecl {
  name?: Ident;
  type?: Expr;
}

export function parseFrontSource(source: string, filename: string): ParseFrontResult {
  const scanned = scanSource(source, filename);
  const p = new parser(scanned.tokens, scanned.diagnostics, filename);
  const result = p.parseFile();
  if (result.file) NormalizeAst(result.file);
  for (const statement of result.statements) NormalizeAst(statement);
  return result;
}

export function parseFrontSourceFiles(files: SourceFile[]): ParseFrontFilesResult {
  const results = files.map((file) => parseFrontSource(file.source, file.filename));
  return {
    files: results.flatMap((result) => result.file ? [result.file] : []),
    statements: results.flatMap((result) => result.statements),
    diagnostics: results.flatMap((result) => result.diagnostics),
    results
  };
}

export function readSource(filename: string, src: unknown): [string | undefined, Error | undefined] {
  if (src !== null && src !== undefined) {
    if (typeof src === "string") {
      return [src, undefined];
    }
    if (src instanceof Uint8Array) {
      return [new TextDecoder().decode(src), undefined];
    }
    if (src instanceof ArrayBuffer) {
      return [new TextDecoder().decode(src), undefined];
    }
    if (hasBytes(src)) {
      return [new TextDecoder().decode(src.Bytes()), undefined];
    }
    if (hasRead(src)) {
      const text = src.Read();
      if (typeof text === "string") {
        return [text, undefined];
      }
      if (text instanceof Uint8Array) {
        return [new TextDecoder().decode(text), undefined];
      }
    }
    return [undefined, new Error("invalid source")];
  }
  const reader = hostReadFile();
  if (!reader) {
    return [undefined, new Error(`could not read ${filename}: no host file reader available`)];
  }
  try {
    return [reader(filename), undefined];
  } catch (err) {
    return [undefined, err instanceof Error ? err : new Error(String(err))];
  }
}

export function ParseFile(_fset: unknown, filename: string, src: unknown, mode: Mode = 0): [File | undefined, Error | undefined] {
  if (_fset === null || _fset === undefined) {
    throw new Error("parser.ParseFile: no token.FileSet provided (fset == nil)");
  }

  const [text, readErr] = readSource(filename, src);
  if (readErr || text === undefined) {
    return [undefined, readErr];
  }

  const result = parseFrontSource(text, filename);
  let file = result.file;
  if (!file) {
    file = {
      kind: "File",
      name: ident(""),
      declarations: [],
      imports: [],
      unresolved: [],
      comments: []
    };
  }
  if ((mode & PackageClauseOnly) !== 0) {
    file = { ...file, declarations: [], imports: [], unresolved: [] };
  } else if ((mode & ImportsOnly) !== 0) {
    const declarations = file.declarations.filter((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import);
    file = { ...file, declarations, imports: declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs.filter((spec): spec is ImportSpec => spec.kind === "ImportSpec") : []) };
  }

  if ((mode & SkipObjectResolution) === 0 && (mode & PackageClauseOnly) === 0) {
    const declErr = (mode & DeclarationErrors) !== 0
      ? (pos: Pos, message: string) => {
        const span = posSpanForResolver(_fset, pos);
        result.diagnostics.push({
          filename: diagnosticFilename(span, filename),
          code: "GOJR_PARSE_FRONT001",
          severity: "error",
          message,
          span
        });
      }
      : undefined;
    resolveFile(file, _fset, declErr);
  }

  return [file, diagnosticAsError(result.diagnostics[0])];
}

export function ParseDir(fset: unknown, path: string, filter: ((info: { name: string; isFile: boolean }) => boolean) | undefined, mode: Mode = 0): [Map<string, Package> | undefined, Error | undefined] {
  const reader = hostReadDir();
  if (!reader) {
    return [undefined, new Error(`could not read directory ${path}: no host directory reader available`)];
  }

  const packages = new Map<string, Package>();
  let first: Error | undefined;
  for (const entry of reader(path)) {
    if (!entry.isFile || !entry.name.endsWith(".go")) {
      continue;
    }
    if (filter && !filter(entry)) {
      continue;
    }
    const filename = `${path.replace(/\/$/u, "")}/${entry.name}`;
    const [src, err] = ParseFile(fset, filename, undefined, mode);
    if (src && !err) {
      const name = src.name?.name ?? "";
      let pkg = packages.get(name);
      if (!pkg) {
        pkg = { kind: "Package", name, files: [] };
        packages.set(name, pkg);
      }
      if (Array.isArray(pkg.files)) {
        pkg.files.push(src);
      }
    } else if (!first) {
      first = err;
    }
  }
  return [packages, first];
}

export function ParseExprFrom(fset: unknown, filename: string, src: unknown, mode: Mode = 0): [Expr | undefined, Error | undefined] {
  if (fset === null || fset === undefined) {
    throw new Error("parser.ParseExprFrom: no token.FileSet provided (fset == nil)");
  }

  const [text, readErr] = readSource(filename, src);
  if (readErr || text === undefined) {
    return [undefined, readErr];
  }

  const result = parseFrontSource(`package p\nvar _ = ${text}`, filename || REPL_FILENAME);
  const declaration = result.file?.declarations.find((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Var);
  const spec = declaration?.kind === "GenDecl" ? declaration.specs[0] : undefined;
  const expr = spec?.kind === "ValueSpec" ? spec.values[0] : undefined;
  return [expr, diagnosticAsError(result.diagnostics[0])];
}

export function ParseExpr(x: string): [Expr | undefined, Error | undefined] {
  return ParseExprFrom({}, "", x, 0);
}

class parser {
  private index = 0;
  public indent = 0;
  public nestLev = 0;
  // Faithful port of go/parser.parser.exprLev:
  // exprLev < 0 means we are parsing an if/for/switch control clause, where
  // a following "{" may be the statement body rather than a composite literal.
  // Parenthesized/call/index subexpressions increment it back into expression
  // context, exactly like the standard parser.
  private exprLev = 0;
  // Faithful port of go/parser.parser.inRhs: while parsing right-hand-side
  // expression lists, "=" is treated as equality for tolerant parsing.
  private inRhs = false;
  private allowBareIdentifierComposite = true;
  private allowSpreadsheetRanges = true;
  private readonly diagnostics: Diagnostic[];

  public constructor(
    private readonly tokens: FrontToken[],
    diagnostics: Diagnostic[],
    private readonly filename: string
  ) {
    this.diagnostics = [...diagnostics];
  }

  public parseFile(): ParseFrontResult {
    let name: Ident | undefined;
    if (this.match(TokenKind.Package)) {
      name = this.parseIdent("expected package name");
      this.consumeSemi();
    }

    const declarations: Decl[] = [];
    const imports: ImportSpec[] = [];
    const statements: Stmt[] = [];
    while (!this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.EOF)) break;
      if (this.startsDecl()) {
        const declaration = this.parseDecl();
        declarations.push(declaration);
        if (declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
          imports.push(...declaration.specs.filter((spec): spec is ImportSpec => spec.kind === "ImportSpec"));
        }
      } else {
        statements.push(this.parseStatement());
      }
      this.consumeSemi();
    }

    return {
      tokens: this.tokens,
      diagnostics: this.diagnostics,
      statements,
      file: {
        kind: "File",
        ...(name ? { name } : {}),
        declarations,
        imports,
        unresolved: [],
        comments: [],
        span: mergeSpans(name?.span, declarations[declarations.length - 1]?.span)
      }
    };
  }

  private parseDecl(): Decl {
    if (this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var)) return this.parseGenDecl();
    if (this.at(TokenKind.Func)) return this.parseFuncDecl();
    const token = this.peek();
    this.error(`expected declaration, found ${token.lexeme || token.kind}`, token.span);
    this.advance();
    return { kind: "BadDecl", span: token.span };
  }

  private startsDecl(): boolean {
    return this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func);
  }

  private parseGenDecl(): GenDecl {
    const start = this.advance();
    const token = start.kind as TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var;
    const specs: Spec[] = [];
    let grouped = false;

    if (this.match(TokenKind.LParen)) {
      grouped = true;
      while (!this.at(TokenKind.RParen) && !this.at(TokenKind.EOF)) {
        this.skipSemis();
        if (this.at(TokenKind.RParen)) break;
        specs.push(this.parseSpec(token));
        this.consumeSemi();
      }
      this.expect(TokenKind.RParen, "expected ')' after declaration group");
    } else {
      specs.push(this.parseSpec(token));
    }

    return {
      kind: "GenDecl",
      token,
      specs,
      grouped,
      span: mergeSpans(start.span, specs[specs.length - 1]?.span)
    };
  }

  private parseSpec(token: TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var): Spec {
    if (token === TokenKind.Import) return this.parseImportSpec();
    if (token === TokenKind.Type) return this.parseTypeSpec();
    return this.parseValueSpec();
  }

  private parseImportSpec(): ImportSpec {
    const start = this.peek();
    let name: Ident | undefined;
    if (this.at(TokenKind.Identifier) && this.peek(1).kind === TokenKind.StringLiteral) {
      name = this.parseIdent("expected import alias");
    } else if (this.at(TokenKind.Dot) && this.peek(1).kind === TokenKind.StringLiteral) {
      const dot = this.advance();
      name = ident(".", dot.span);
    }
    const path = this.expectBasicLit(TokenKind.StringLiteral, "expected import path string");
    return {
      kind: "ImportSpec",
      ...(name ? { name } : {}),
      path,
      span: mergeSpans(start.span, path.span)
    };
  }

  private parseTypeSpec(): TypeSpec {
    const name = this.parseIdent("expected type name");
    let typeParams: FieldList | undefined;
    let alias = false;
    let type: Expr;

    if (this.match(TokenKind.LBracket)) {
      const open = this.previous();
      if (isIdentifierLike(this.peek().kind)) {
        const firstName = this.parseIdent("expected type parameter name or array length");
        let expression: Expr = firstName;
        if (!this.at(TokenKind.LBracket)) {
          this.withExpressionLevel(() => {
            expression = this.parseBinaryExpression(this.parsePrimaryFrom(expression), 1);
          });
        }
        const { name: paramName, type: paramType } = extractName(expression, this.at(TokenKind.Comma));
        if (paramName && (paramType || !this.at(TokenKind.RBracket))) {
          const spec: TypeSpec = {
            kind: "TypeSpec",
            name,
            type: badExpr(open.span),
            alias: false
          };
          this.parseGenericType(spec, open, paramName, paramType);
          return {
            ...spec,
            span: mergeSpans(name.span, spec.type.span)
          };
        } else {
          type = this.parseArrayTypeAfterOpen(open, expression);
        }
      } else {
        type = this.parseArrayTypeAfterOpen(open);
      }
    } else {
      alias = this.match(TokenKind.Assign);
      type = this.parseType();
    }
    return {
      kind: "TypeSpec",
      name,
      ...(typeParams ? { typeParams } : {}),
      type,
      alias,
      span: mergeSpans(name.span, type.span)
    };
  }

  private parseGenericType(spec: TypeSpec, open: FrontToken, name0: Ident, typ0?: Expr): void {
    spec.typeParams = this.parseTypeParameterListAfterOpen(open, name0, typ0);
    spec.alias = this.match(TokenKind.Assign);
    spec.type = this.parseType();
  }

  private parseValueSpec(): ValueSpec {
    const names = this.parseIdentList();
    let type: Expr | undefined;
    let values: Expr[] = [];
    if (!this.atAny(TokenKind.Assign, TokenKind.Semicolon, TokenKind.RParen, TokenKind.EOF)) {
      type = this.parseType();
    }
    if (this.match(TokenKind.Assign)) {
      values = this.parseExpressionList(true);
    }
    return {
      kind: "ValueSpec",
      names,
      ...(type ? { type } : {}),
      values,
      span: mergeSpans(names[0]?.span, values[values.length - 1]?.span ?? type?.span ?? names[names.length - 1]?.span)
    };
  }

  private parseFuncDecl(): FuncDecl {
    const start = this.expect(TokenKind.Func, "expected func");
    let receiver: FieldList | undefined;
    if (this.at(TokenKind.LParen) && this.looksLikeReceiver()) {
      receiver = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    }
    const name = this.parseIdent("expected function name");
    const typeParams = this.at(TokenKind.LBracket) ? this.parseTypeParamList() : undefined;
    const type = this.parseSignature(start.span, typeParams);
    const body = this.at(TokenKind.LBrace) ? this.parseBlock() : undefined;
    return {
      kind: "FuncDecl",
      ...(receiver ? { receiver } : {}),
      name,
      type,
      ...(body ? { body } : {}),
      span: mergeSpans(start.span, body?.span ?? type.span)
    };
  }

  private parseStatement(): Stmt {
    this.skipSemis();
    if (this.at(TokenKind.RBrace)) {
      return { kind: "EmptyStmt", implicit: true, span: this.peek().span };
    }
    if (this.at(TokenKind.LBrace)) return this.parseBlock();
    if (this.atAny(TokenKind.Const, TokenKind.Type, TokenKind.Var)) {
      return { kind: "DeclStmt", decl: this.parseGenDecl() };
    }
    if (this.match(TokenKind.Return)) {
      const start = this.previous();
      const results = this.atAny(TokenKind.Semicolon, TokenKind.RBrace, TokenKind.EOF) ? [] : this.parseExpressionList(true);
      return { kind: "ReturnStmt", results, span: mergeSpans(start.span, results[results.length - 1]?.span ?? start.span) };
    }
    if (this.atAny(TokenKind.Break, TokenKind.Continue, TokenKind.Goto, TokenKind.Fallthrough)) {
      const token = this.advance();
      const acceptsLabel = token.kind === TokenKind.Goto ||
        ((token.kind === TokenKind.Break || token.kind === TokenKind.Continue) && this.at(TokenKind.Identifier));
      const label = acceptsLabel ? this.parseIdent("expected label") : undefined;
      return {
        kind: "BranchStmt",
        token: token.kind as BranchStmt["token"],
        ...(label ? { label } : {}),
        span: mergeSpans(token.span, label?.span)
      };
    }
    if (this.match(TokenKind.Defer)) {
      const start = this.previous();
      const expression = this.parseRhsExpression();
      const call = expression.kind === "CallExpr"
        ? expression
        : ({ kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) } satisfies CallExpr);
      return { kind: "DeferStmt", call, span: mergeSpans(start.span, call.span) } satisfies DeferStmt;
    }
    if (this.at(TokenKind.Go)) return this.parseGoStmt();
    if (this.at(TokenKind.Select)) return this.parseSelectStmt();
    if (this.at(TokenKind.For)) return this.parseForStmt();
    if (this.at(TokenKind.If)) return this.parseIfStmt();
    if (this.at(TokenKind.Switch)) return this.parseSwitchStmt();

    return this.parseSimpleStmt("labelOk");
  }

  private parseSimpleStmt(mode: SimpleStmtMode = "basic"): Stmt {
    return this.parseSimpleStmtWithRange(mode).statement;
  }

  private parseSimpleStmtWithRange(mode: SimpleStmtMode): SimpleStmtResult {
    const lhs = this.parseExpressionList(false);
    if (this.match(TokenKind.Arrow)) {
      if (lhs.length !== 1) {
        this.error("send statement expects one channel expression", lhs[1]?.span ?? lhs[0]?.span);
      }
      const value = this.parseRhsExpression();
      return {
        statement: {
        kind: "SendStmt",
        channel: lhs[0] ?? badExpr(value.span),
        value,
        span: mergeSpans(lhs[0]?.span, value.span)
        },
        isRange: false
      };
    }
    if (isAssignmentToken(this.peek().kind)) {
      const token = this.advance();
      let rhs: Expr[];
      let isRange = false;
      if ((token.kind === TokenKind.Assign || token.kind === TokenKind.Define) && mode === "rangeOk" && this.match(TokenKind.Range)) {
        rhs = [this.parseRhsExpression()];
        isRange = true;
      } else {
        rhs = this.parseExpressionList(true);
      }
      return {
        statement: {
        kind: "AssignStmt",
        lhs,
        token: token.kind as Extract<Stmt, { kind: "AssignStmt" }>["token"],
        rhs,
        span: mergeSpans(lhs[0]?.span, rhs[rhs.length - 1]?.span)
        },
        isRange
      };
    }
    if (this.match(TokenKind.Colon)) {
      const colon = this.previous();
      if (mode === "labelOk" && lhs[0]?.kind === "Ident" && lhs.length === 1) {
        const stmt = this.parseStatement();
        return {
          statement: {
            kind: "LabeledStmt",
            label: lhs[0],
            stmt,
            span: mergeSpans(lhs[0].span, stmt.span)
          },
          isRange: false
        };
      }
      this.error("illegal label declaration", colon.span);
      return { statement: { kind: "BadStmt", span: mergeSpans(lhs[0]?.span, colon.span) }, isRange: false };
    }
    if (this.at(TokenKind.PlusPlus) || this.at(TokenKind.MinusMinus)) {
      const token = this.advance();
      return {
        statement: {
        kind: "IncDecStmt",
        expr: lhs[0] ?? badExpr(token.span),
        token: token.kind as TokenKind.PlusPlus | TokenKind.MinusMinus,
        span: mergeSpans(lhs[0]?.span, token.span)
        },
        isRange: false
      };
    }
    if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
    return {
      statement: { kind: "ExprStmt", expr: lhs[0] ?? badExpr(this.peek().span), span: mergeSpans(lhs[0]?.span, lhs[0]?.span) },
      isRange: false
    };
  }

  private parseGoStmt(): Stmt {
    const start = this.expect(TokenKind.Go, "expected go");
    const expression = this.parseRhsExpression();
    if (expression.kind !== "CallExpr") {
      this.error("go statement requires function call", expression.span);
      return {
        kind: "GoStmt",
        call: { kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) },
        span: mergeSpans(start.span, expression.span)
      };
    }
    return { kind: "GoStmt", call: expression, span: mergeSpans(start.span, expression.span) };
  }

  private parseIfHeader(): { init?: Stmt; condition: Expr } {
    let init: Stmt | undefined;
    let conditionStatement: Stmt | undefined;
    if (this.at(TokenKind.LBrace)) {
      this.error("missing condition in if statement", this.peek().span);
      return { condition: badExpr(this.peek().span) };
    }
    if (!this.at(TokenKind.Semicolon)) {
      if (this.match(TokenKind.Var)) this.error("var declaration not allowed in if initializer", this.previous().span);
      init = this.parseSimpleStmt("basic");
    }
    if (!this.at(TokenKind.LBrace)) {
      this.expect(TokenKind.Semicolon, "expected ';' after if init statement");
      if (!this.at(TokenKind.LBrace)) conditionStatement = this.parseSimpleStmt("basic");
    } else {
      conditionStatement = init;
      init = undefined;
    }
    return {
      ...(init ? { init } : {}),
      condition: this.statementExpression(conditionStatement, "boolean expression") ?? badExpr(this.peek().span)
    };
  }

  private parseIfStmt(): Stmt {
    const start = this.expect(TokenKind.If, "expected if");
    const { init, condition } = this.withControlClause(() => this.parseIfHeader());
    const body = this.parseBlock();
    let elseStmt: Stmt | undefined;
    if (this.match(TokenKind.Else)) {
      elseStmt = this.at(TokenKind.If) ? this.parseIfStmt() : this.parseBlock();
    }
    return {
      kind: "IfStmt",
      ...(init ? { init } : {}),
      condition,
      body,
      ...(elseStmt ? { else: elseStmt } : {}),
      span: mergeSpans(start.span, elseStmt?.span ?? body.span)
    };
  }

  private parseForStmt(): Stmt {
    const start = this.expect(TokenKind.For, "expected for");
    const header = this.withControlClause(() => {
      let first: Stmt | undefined;
      let second: Stmt | undefined;
      let third: Stmt | undefined;
      let isRange = false;

      if (!this.at(TokenKind.LBrace)) {
        if (!this.at(TokenKind.Semicolon)) {
          if (this.match(TokenKind.Range)) {
            const source = this.parseRhsExpression();
            second = { kind: "AssignStmt", lhs: [], token: TokenKind.Assign, rhs: [source], span: source.span ?? start.span };
            isRange = true;
          } else {
            const parsed = this.parseSimpleStmtWithRange("rangeOk");
            second = parsed.statement;
            isRange = parsed.isRange;
          }
        }
        if (!isRange && this.match(TokenKind.Semicolon)) {
          first = second;
          second = undefined;
          if (!this.at(TokenKind.Semicolon)) second = this.parseSimpleStmt("basic");
          this.expect(TokenKind.Semicolon, "expected ';' in for clause");
          if (!this.at(TokenKind.LBrace)) third = this.parseSimpleStmt("basic");
        }
      }
      return { first, second, third, isRange };
    });
    const body = this.parseBlock();
    if (header.isRange) {
      const assign = header.second?.kind === "AssignStmt" ? header.second : undefined;
      const lhs = assign?.lhs ?? [];
      if (lhs.length > 2) this.error("expected at most 2 expressions", lhs[2]?.span ?? assign?.span);
      const source = assign?.rhs[0] ?? badExpr(header.second?.span ?? start.span);
      return {
        kind: "RangeStmt",
        ...(lhs[0] ? { key: lhs[0] } : {}),
        ...(lhs[1] ? { value: lhs[1] } : {}),
        token: assign?.token === TokenKind.Define ? TokenKind.Define : TokenKind.Assign,
        source,
        body,
        span: mergeSpans(start.span, body.span)
      };
    }
    const condition = this.statementExpression(header.second, "boolean or range expression");
    return {
      kind: "ForStmt",
      ...(header.first ? { init: header.first } : {}),
      ...(condition ? { condition } : {}),
      ...(header.third ? { post: header.third } : {}),
      body,
      span: mergeSpans(start.span, body.span)
    };
  }

  private parseSwitchStmt(): Stmt {
    const start = this.expect(TokenKind.Switch, "expected switch");
    const { init, tag, assign, typeSwitch } = this.withControlClause(() => {
      let init: Stmt | undefined;
      let tag: Expr | undefined;
      let assign: Stmt | undefined;
      let typeSwitch = false;

      if (!this.at(TokenKind.LBrace)) {
        const first = this.at(TokenKind.Semicolon) ? undefined : this.parseSimpleStmt("basic");
        if (this.match(TokenKind.Semicolon)) {
          init = first;
          if (!this.at(TokenKind.LBrace)) {
            const second = this.parseSimpleStmt("basic");
            if (this.isTypeSwitchGuard(second)) {
              assign = second;
              typeSwitch = true;
            } else if (second.kind === "ExprStmt") {
              tag = second.expr;
            } else {
              this.error("expected switch expression or type switch guard after ';'", second.span);
            }
          }
        } else if (first && this.isTypeSwitchGuard(first)) {
          assign = first;
          typeSwitch = true;
        } else if (first?.kind === "ExprStmt") {
          tag = first.expr;
        } else if (first) {
          this.error("expected ';' after switch init statement", first.span);
          init = first;
        }
      }
      return { init, tag, assign, typeSwitch };
    });

    const { clauses, span } = this.parseSwitchBody();
    if (typeSwitch) {
      return {
        kind: "TypeSwitchStmt",
        ...(init ? { init } : {}),
        assign: assign ?? { kind: "BadStmt", span: start.span },
        body: clauses,
        span: mergeSpans(start.span, span)
      };
    }
    return {
      kind: "SwitchStmt",
      ...(init ? { init } : {}),
      ...(tag ? { tag } : {}),
      body: clauses,
      span: mergeSpans(start.span, span)
    };
  }

  private parseSwitchBody(): { clauses: CaseClause[]; span: SourceSpan } {
    const start = this.expect(TokenKind.LBrace, "expected '{' after switch");
    const clauses: CaseClause[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
        clauses.push(this.parseCaseClause());
      } else {
        const token = this.peek();
        this.error("expected case or default in switch body", token.span);
        this.advance();
      }
    }
    const end = this.expect(TokenKind.RBrace, "expected '}' after switch body");
    return { clauses, span: mergeSpans(start.span, end.span) };
  }

  private parseSelectStmt(): Stmt {
    const start = this.expect(TokenKind.Select, "expected select");
    const open = this.expect(TokenKind.LBrace, "expected '{' after select");
    const clauses: Extract<Stmt, { kind: "CommClause" }>[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
        clauses.push(this.parseCommClause());
      } else {
        const token = this.peek();
        this.error("expected case or default in select body", token.span);
        this.advance();
      }
    }
    const close = this.expect(TokenKind.RBrace, "expected '}' after select body");
    return {
      kind: "SelectStmt",
      body: clauses,
      span: mergeSpans(start.span, close.span ?? open.span)
    };
  }

  private parseCommClause(): Extract<Stmt, { kind: "CommClause" }> {
    const start = this.advance();
    const isDefault = start.kind === TokenKind.Default;
    let comm: Stmt | undefined;
    if (!isDefault && !this.at(TokenKind.Colon)) {
      const lhs = this.parseExpressionList(false);
      if (this.match(TokenKind.Arrow)) {
        if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
        const value = this.parseRhsExpression();
        comm = {
          kind: "SendStmt",
          channel: lhs[0] ?? badExpr(value.span),
          value,
          span: mergeSpans(lhs[0]?.span, value.span)
        };
      } else if (this.at(TokenKind.Assign) || this.at(TokenKind.Define)) {
        if (lhs.length > 2) this.error("expected 1 or 2 expressions", lhs[2]?.span ?? lhs[0]?.span);
        const token = this.advance();
        const rhs = this.parseRhsExpression();
        comm = {
          kind: "AssignStmt",
          lhs,
          token: token.kind as TokenKind.Assign | TokenKind.Define,
          rhs: [rhs],
          span: mergeSpans(lhs[0]?.span, rhs.span)
        };
      } else {
        if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
        comm = {
          kind: "ExprStmt",
          expr: lhs[0] ?? badExpr(start.span),
          span: mergeSpans(lhs[0]?.span, lhs[0]?.span)
        };
      }
    }
    this.expect(TokenKind.Colon, "expected ':' after select case");
    const body: Stmt[] = [];
    while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
      this.skipSemis();
      if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) break;
      body.push(this.parseStatement());
      this.consumeSemi();
    }
    return {
      kind: "CommClause",
      ...(comm ? { comm } : {}),
      body,
      default: isDefault,
      span: mergeSpans(start.span, body[body.length - 1]?.span ?? comm?.span ?? start.span)
    };
  }

  private parseCaseClause(): CaseClause {
    const start = this.advance();
    const isDefault = start.kind === TokenKind.Default;
    const list = isDefault ? [] : this.withSpreadsheetRanges(false, () => this.parseExpressionList(true));
    this.expect(TokenKind.Colon, "expected ':' after switch case");
    const body: Stmt[] = [];
    while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
      this.skipSemis();
      if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) break;
      body.push(this.parseStatement());
      this.consumeSemi();
    }
    return {
      kind: "CaseClause",
      list,
      body,
      default: isDefault,
      span: mergeSpans(start.span, body[body.length - 1]?.span ?? list[list.length - 1]?.span ?? start.span)
    };
  }

  private isTypeSwitchGuard(statement: Stmt): boolean {
    if (statement.kind === "ExprStmt") {
      return statement.expr.kind === "TypeAssertExpr" && statement.expr.typeSwitch;
    }
    if (statement.kind !== "AssignStmt" || statement.rhs.length !== 1) return false;
    const rhs = statement.rhs[0];
    return rhs?.kind === "TypeAssertExpr" && rhs.typeSwitch;
  }

  private parseBlock(): BlockStmt {
    const start = this.expect(TokenKind.LBrace, "expected '{'");
    const statements: Stmt[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      statements.push(this.parseStatement());
      this.consumeSemi();
    }
    const end = this.expect(TokenKind.RBrace, "expected '}'");
    return {
      kind: "BlockStmt",
      statements,
      span: mergeSpans(start.span, end.span)
    };
  }

  private parseExpressionList(inRhs = false): Expr[] {
    return this.withRhs(inRhs, () => {
      const expressions = [this.parseExpression()];
      while (this.match(TokenKind.Comma)) {
        expressions.push(this.parseExpression());
      }
      return expressions;
    });
  }

  private parseRhsExpression(): Expr {
    return this.withRhs(true, () => this.parseExpression());
  }

  private parseExpression(minPrecedence = 1): Expr {
    const left = this.parseUnary();
    return this.parseBinaryExpression(left, minPrecedence);
  }

  private parseBinaryExpression(leftOperand: Expr, minPrecedence = 1): Expr {
    let left = leftOperand;
    while (true) {
      const operatorKind = this.inRhs && this.peek().kind === TokenKind.Assign
        ? TokenKind.Equal
        : this.peek().kind;
      const precedence = binaryPrecedence(operatorKind);
      if (precedence < minPrecedence) break;
      const operator = this.advance();
      const right = this.parseExpression(precedence + 1);
      left = {
        kind: "BinaryExpr",
        left,
        op: operatorKind as BinaryOperator,
        right,
        span: mergeSpans(left.span, right.span)
      };
    }
    return left;
  }

  private statementExpression(statement: Stmt | undefined, want: string): Expr | undefined {
    if (!statement) return undefined;
    if (statement.kind === "ExprStmt") return statement.expr;
    const found = statement.kind === "AssignStmt" ? "assignment" : "simple statement";
    this.error(`expected ${want}, found ${found} (missing parentheses around composite literal?)`, statement.span);
    return badExpr(statement.span);
  }

  private parseUnary(): Expr {
    if (this.atAny(TokenKind.Plus, TokenKind.Minus, TokenKind.Bang, TokenKind.Caret, TokenKind.Amp, TokenKind.Arrow)) {
      const operator = this.advance();
      const expr = this.parseUnary();
      return {
        kind: "UnaryExpr",
        op: operator.kind as TokenKind.Plus | TokenKind.Minus | TokenKind.Bang | TokenKind.Caret | TokenKind.Amp | TokenKind.Arrow,
        expr,
        span: mergeSpans(operator.span, expr.span)
      };
    }
    if (this.match(TokenKind.Star)) {
      const start = this.previous();
      const expr = this.parseUnary();
      return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
    }
    return this.parsePrimary();
  }

  private parsePrimary(): Expr {
    return this.parsePrimaryFrom(this.parseOperand());
  }

  private parsePrimaryFrom(start: Expr): Expr {
    let expression = start;
    while (true) {
      if (this.match(TokenKind.Dot)) {
        const dot = this.previous();
        if (this.match(TokenKind.LParen)) {
          if (this.match(TokenKind.Type)) {
            const close = this.expect(TokenKind.RParen, "expected ')' after type switch guard");
            expression = {
              kind: "TypeAssertExpr",
              object: expression,
              typeSwitch: true,
              span: mergeSpans(expression.span, close.span)
            };
          } else {
            const type = this.parseType();
            const close = this.expect(TokenKind.RParen, "expected ')' after type assertion");
            expression = {
              kind: "TypeAssertExpr",
              object: expression,
              type,
              typeSwitch: false,
              span: mergeSpans(expression.span, close.span)
            };
          }
          continue;
        }

        const cellCandidate = this.peek();
        const cellStart = (cellCandidate.kind === TokenKind.CellAddress || cellCandidate.kind === TokenKind.Identifier)
          ? parseCellAddress(cellCandidate.lexeme)
          : undefined;
        const rangeEndCandidate = this.peek(2);
        const hasSpreadsheetRangeEnd = this.allowSpreadsheetRanges &&
          this.peek(1).kind === TokenKind.Colon &&
          (rangeEndCandidate.kind === TokenKind.CellAddress || rangeEndCandidate.kind === TokenKind.Identifier) &&
          !!parseCellAddress(rangeEndCandidate.lexeme);
        if (cellStart && expression.kind === "Ident" && (cellCandidate.kind === TokenKind.CellAddress || hasSpreadsheetRangeEnd)) {
          const cellToken = this.advance();
          if (this.match(TokenKind.Colon)) {
            const { token: endToken, address: end } = this.expectCellAddress("expected cell address after ':'");
            if (!end) {
              this.error("malformed spreadsheet range end", endToken.span);
              expression = badExpr(endToken.span);
            } else {
              expression = {
                kind: "RangeRefExpr",
                namespace: expression,
                start: cellStart,
                end,
                span: mergeSpans(expression.span, endToken.span)
              };
            }
          } else {
            expression = {
              kind: "CellRefExpr",
              namespace: expression,
              address: cellStart,
              span: mergeSpans(expression.span, cellToken.span)
            };
          }
          continue;
        }

        const selector = this.parseIdent("expected selector after '.'");
        expression = {
          kind: "SelectorExpr",
          object: expression,
          selector,
          span: mergeSpans(expression.span, selector.span ?? dot.span)
        };
        continue;
      }

      if (this.match(TokenKind.LParen)) {
        const args: Expr[] = [];
        let ellipsis = false;
        if (!this.at(TokenKind.RParen)) {
          args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
          if (this.match(TokenKind.Ellipsis)) ellipsis = true;
          while (this.match(TokenKind.Comma) && !this.at(TokenKind.RParen)) {
            args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
            if (this.match(TokenKind.Ellipsis)) ellipsis = true;
          }
        }
        const close = this.expect(TokenKind.RParen, "expected ')' after arguments");
        expression = {
          kind: "CallExpr",
          fun: expression,
          args,
          ellipsis,
          span: mergeSpans(expression.span, close.span)
        };
        continue;
      }

      if (this.match(TokenKind.LBracket)) {
        const open = this.previous();
        const low = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
          ? undefined
          : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
        if (this.match(TokenKind.Colon)) {
          const high = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
            ? undefined
            : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
          const max = this.match(TokenKind.Colon)
            ? this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()))
            : undefined;
          const close = this.expect(TokenKind.RBracket, "expected ']' after slice");
          expression = {
            kind: "SliceExpr",
            object: expression,
            ...(low ? { low } : {}),
            ...(high ? { high } : {}),
            ...(max ? { max } : {}),
            span: mergeSpans(expression.span, close.span)
          };
        } else if (this.match(TokenKind.Comma)) {
          const indices = [low ?? badExpr(open.span)];
          while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
            indices.push(this.withExpressionLevel(() => this.parseType()));
            if (!this.match(TokenKind.Comma)) break;
          }
          const close = this.expect(TokenKind.RBracket, "expected ']' after type arguments");
          expression = indices.length === 1
            ? {
              kind: "IndexExpr",
              object: expression,
              index: indices[0]!,
              span: mergeSpans(expression.span, close.span)
            }
            : {
              kind: "IndexListExpr",
              object: expression,
              indices,
              span: mergeSpans(expression.span, close.span)
            };
        } else {
          const close = this.expect(TokenKind.RBracket, "expected ']' after index");
          expression = {
            kind: "IndexExpr",
            object: expression,
            index: low ?? badExpr(close.span),
            span: mergeSpans(expression.span, close.span)
          };
        }
        continue;
      }

      if (this.at(TokenKind.LBrace) && this.canUseCompositeLiteralType(expression)) {
        this.advance();
        expression = this.finishCompositeLiteral(expression);
        continue;
      }

      break;
    }

    return expression;
  }

  private parseOperand(): Expr {
    const token = this.peek();
    if (this.at(TokenKind.Identifier) || this.at(TokenKind.CellAddress)) return this.parseIdent("expected identifier");
    if (this.at(TokenKind.True) || this.at(TokenKind.False) || this.at(TokenKind.Nil)) {
      const keyword = this.advance();
      return ident(keyword.lexeme, keyword.span);
    }
    if (this.at(TokenKind.IntLiteral) || this.at(TokenKind.FloatLiteral) || this.at(TokenKind.ImagLiteral) || this.at(TokenKind.RuneLiteral) || this.at(TokenKind.StringLiteral)) {
      return this.expectBasicLit(this.peek().kind as BasicLit["token"], "expected literal");
    }
    if (this.match(TokenKind.LParen)) {
      const start = this.previous();
      const expr = this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
      const close = this.expect(TokenKind.RParen, "expected ')'");
      return { kind: "ParenExpr", expr, span: mergeSpans(start.span, close.span) };
    }
    if (this.atAny(TokenKind.LBracket, TokenKind.Map, TokenKind.Struct, TokenKind.Interface, TokenKind.Chan)) return this.parseType();
    if (this.at(TokenKind.Func)) return this.parseFuncTypeOrLit();

    this.error(`expected expression, found ${token.lexeme || token.kind}`, token.span);
    this.advance();
    return badExpr(token.span);
  }

  private canUseCompositeLiteralType(expression: Expr): boolean {
    if (expression.kind === "ArrayType" || expression.kind === "MapType" || expression.kind === "StructType") {
      return true;
    }
    if (
      (expression.kind === "Ident" && !["true", "false", "nil"].includes(expression.name)) ||
      expression.kind === "SelectorExpr" ||
      expression.kind === "IndexExpr" ||
      expression.kind === "IndexListExpr"
    ) {
      return this.exprLev >= 0 && this.allowBareIdentifierComposite;
    }
    return false;
  }

  private finishCompositeLiteral(type?: Expr, startSpan: SourceSpan | undefined = type?.span): CompositeLit {
    const elements: Expr[] = [];
    this.withExpressionLevel(() => {
      while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
        this.skipSemis();
        if (this.at(TokenKind.RBrace)) break;
        const first = this.parseCompositeLiteralElement();
        if (this.match(TokenKind.Colon)) {
          const value = this.parseCompositeLiteralElement();
          elements.push({
            kind: "KeyValueExpr",
            key: first,
            value,
            span: mergeSpans(first.span, value.span)
          } satisfies KeyValueExpr);
        } else {
          elements.push(first);
        }
        this.match(TokenKind.Comma);
        this.consumeSemi();
      }
    });
    const close = this.expect(TokenKind.RBrace, "expected '}' after composite literal");
    return {
      kind: "CompositeLit",
      ...(type ? { type } : {}),
      elements,
      span: mergeSpans(startSpan, close.span)
    };
  }

  private parseCompositeLiteralElement(): Expr {
    if (this.match(TokenKind.LBrace)) {
      const start = this.previous();
      return this.finishCompositeLiteral(undefined, start.span);
    }
    return this.parseRhsExpression();
  }

  private parseType(): Expr {
    let left = this.parseTypeTerm();
    while (this.match(TokenKind.Or)) {
      const operator = this.previous();
      const right = this.parseTypeTerm();
      left = {
        kind: "BinaryExpr",
        left,
        op: operator.kind as BinaryOperator,
        right,
        span: mergeSpans(left.span, right.span)
      };
    }
    return left;
  }

  private parseTypeTerm(): Expr {
    const start = this.peek();
    if (this.match(TokenKind.Tilde)) {
      const expr = this.parseTypeTerm();
      return {
        kind: "UnaryExpr",
        op: TokenKind.Tilde,
        expr,
        span: mergeSpans(start.span, expr.span)
      };
    }
    if (this.match(TokenKind.Arrow)) {
      const chan = this.expect(TokenKind.Chan, "expected chan after '<-' in channel type");
      const value = this.parseType();
      return { kind: "ChanType", direction: "receive", value, span: mergeSpans(start.span, value.span ?? chan.span) };
    }
    if (this.match(TokenKind.Star)) {
      const expr = this.parseTypeTerm();
      return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
    }
    if (this.match(TokenKind.LBracket)) {
      return this.parseArrayTypeAfterOpen(this.previous());
    }
    if (this.match(TokenKind.Map)) {
      this.expect(TokenKind.LBracket, "expected '[' after map");
      const key = this.parseType();
      this.expect(TokenKind.RBracket, "expected ']' after map key type");
      const value = this.parseType();
      return { kind: "MapType", key, value, span: mergeSpans(start.span, value.span) };
    }
    if (this.match(TokenKind.Chan)) {
      const sendOnly = this.match(TokenKind.Arrow);
      const value = this.parseType();
      return {
        kind: "ChanType",
        direction: sendOnly ? "send" : "both",
        value,
        span: mergeSpans(start.span, value.span)
      };
    }
    if (this.match(TokenKind.Struct)) return this.parseStructType(start.span);
    if (this.match(TokenKind.Interface)) return this.parseInterfaceType(start.span);
    if (this.at(TokenKind.Func)) return this.parseFuncType();
    if (this.match(TokenKind.LParen)) {
      const type = this.parseType();
      const close = this.expect(TokenKind.RParen, "expected ')' after type");
      return { kind: "ParenExpr", expr: type, span: mergeSpans(start.span, close.span) };
    }
    return this.parseTypeName();
  }

  private parseTypeName(): Expr {
    let expression: Expr = this.parseIdent("expected type name");
    while (this.match(TokenKind.Dot)) {
      const selector = this.parseIdent("expected selector in qualified type");
      expression = {
        kind: "SelectorExpr",
        object: expression,
        selector,
        span: mergeSpans(expression.span, selector.span)
      };
    }
    if (this.match(TokenKind.LBracket)) {
      const { indices, close } = this.parseTypeArgumentList();
      expression = indices.length === 1 ? {
        kind: "IndexExpr",
        object: expression,
        index: indices[0]!,
        span: mergeSpans(expression.span, close.span)
      } : {
        kind: "IndexListExpr",
        object: expression,
        indices,
        span: mergeSpans(expression.span, close.span)
      };
    }
    return expression;
  }

  private parseArrayTypeAfterOpen(open: FrontToken, parsedLength?: Expr): ArrayType {
    let length = parsedLength;
    let inferredLength = false;
    if (!length) {
      this.withExpressionLevel(() => {
        if (this.match(TokenKind.Ellipsis)) {
          inferredLength = true;
        } else if (!this.at(TokenKind.RBracket)) {
          length = this.parseRhsExpression();
        }
      });
    }
    if (this.match(TokenKind.Comma)) {
      this.error("unexpected comma; expecting ]", this.previous().span);
    }
    this.expect(TokenKind.RBracket, "expected ']' in array or slice type");
    const element = this.parseTypeTerm();
    return {
      kind: "ArrayType",
      ...(length ? { length } : {}),
      element,
      inferredLength,
      span: mergeSpans(open.span, element.span)
    } satisfies ArrayType;
  }

  private parseTypeArgumentList(): { indices: Expr[]; close: FrontToken } {
    const indices: Expr[] = [];
    while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
      indices.push(this.withExpressionLevel(() => this.parseType()));
      if (!this.match(TokenKind.Comma)) break;
    }
    return {
      indices: indices.length > 0 ? indices : [badExpr(this.peek().span)],
      close: this.expect(TokenKind.RBracket, "expected ']' after type arguments")
    };
  }

  private parseStructType(start: SourceSpan): Expr {
    const fields = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
    return { kind: "StructType", fields, span: mergeSpans(start, fields.span) };
  }

  private parseInterfaceType(start: SourceSpan): Expr {
    const methods = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
    return { kind: "InterfaceType", methods, span: mergeSpans(start, methods.span) };
  }

  private parseSignature(start: SourceSpan, typeParams?: FieldList): FuncType {
    const params = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    let results: FieldList | undefined;
    if (this.at(TokenKind.LParen)) {
      results = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    } else if (this.startsType()) {
      const type = this.parseType();
      results = {
        kind: "FieldList",
        fields: [{ kind: "Field", names: [], type, span: mergeSpans(type.span, type.span) }],
        span: mergeSpans(type.span, type.span)
      };
    }
    return {
      kind: "FuncType",
      ...(typeParams ? { typeParams } : {}),
      params,
      ...(results ? { results } : {}),
      span: mergeSpans(start, results?.span ?? params.span)
    };
  }

  private parseFuncType(): FuncType {
    const start = this.expect(TokenKind.Func, "expected func");
    if (this.at(TokenKind.LBracket)) {
      const typeParams = this.parseTypeParamList();
      this.error("function type must have no type parameters", typeParams.span);
    }
    return this.parseSignature(start.span);
  }

  private parseFuncTypeOrLit(): Expr {
    const type = this.parseFuncType();
    if (!this.at(TokenKind.LBrace)) {
      return type;
    }

    const body = this.withExpressionLevel(() => this.parseBlock());
    return { kind: "FuncLit", type, body, span: mergeSpans(type.span, body.span) } satisfies FuncLit;
  }

  private parseTypeParamList(): FieldList {
    return this.parseFieldList(TokenKind.LBracket, TokenKind.RBracket);
  }

  private parseTypeParameterListAfterOpen(open: FrontToken, firstName?: Ident, firstType?: Expr): FieldList {
    const fields = this.parseParameterList(TokenKind.RBracket, firstName, firstType, false);
    const close = this.expect(TokenKind.RBracket, "expected ']'");
    if (fields.length === 0) this.error("empty type parameter list", close.span);
    return { kind: "FieldList", fields, span: mergeSpans(open.span, close.span) };
  }

  private parseParameterList(closing: TokenKind, firstName?: Ident, firstType?: Expr, allowEllipsis = false): Field[] {
    const typeParams = closing === TokenKind.RBracket;
    const params: ParamDecl[] = [];
    let name0 = firstName;
    let type0 = firstType;
    let named = 0;
    let typed = 0;

    while (name0 || (!this.at(closing) && !this.at(TokenKind.EOF))) {
      let param: ParamDecl;
      if (type0) {
        param = { ...(name0 ? { name: name0 } : {}), type: typeParams ? this.embeddedElem(type0) : type0 };
      } else {
        param = this.parseParamDecl(name0, typeParams);
      }
      name0 = undefined;
      type0 = undefined;
      if (param.name || param.type) {
        params.push(param);
        if (param.name && param.type) named += 1;
        if (param.type) typed += 1;
      }
      if (this.at(closing) || this.at(TokenKind.EOF)) break;
      if (!this.match(TokenKind.Comma)) break;
    }

    if (params.length === 0) return [];
    if (named === 0) {
      for (const param of params) {
        if (param.name && !param.type) {
          param.type = param.name;
          delete param.name;
        }
      }
      if (typeParams) {
        const message = named === typed ? "missing type constraint" : `missing type parameter name${params.length === 1 ? " or invalid array length" : ""}`;
        this.error(message, this.peek().span);
      }
    } else if (named !== params.length) {
      let type: Expr | undefined;
      let errorSpan: SourceSpan | undefined;
      for (let index = params.length - 1; index >= 0; index -= 1) {
        const param = params[index]!;
        if (param.type) {
          type = param.type;
          if (!param.name) {
            errorSpan = param.type.span;
            param.name = ident("_", errorSpan);
          }
        } else if (type) {
          param.type = type;
        } else {
          errorSpan = param.name?.span;
          param.type = badExpr(errorSpan);
        }
      }
      if (errorSpan) {
        const message = named === typed
          ? (typeParams ? "missing type constraint" : "missing parameter type")
          : (typeParams ? `missing type parameter name${params.length === 1 ? " or invalid array length" : ""}` : "missing parameter name");
        this.error(message, errorSpan);
      }
    }

    let reportedEllipsis = false;
    for (let index = 0; index < params.length; index += 1) {
      const param = params[index]!;
      if (param.type?.kind === "Ellipsis" && (!allowEllipsis || index + 1 < params.length)) {
        if (!reportedEllipsis) {
          this.error(allowEllipsis ? "can only use ... with final parameter" : "invalid use of ...", param.type.span);
          reportedEllipsis = true;
        }
        param.type = badExpr(param.type.span);
      }
    }

    if (named === 0) {
      return params.map((param) => ({
        kind: "Field",
        names: [],
        type: param.type ?? badExpr(param.name?.span),
        span: mergeSpans(param.type?.span ?? param.name?.span, param.type?.span ?? param.name?.span)
      }));
    }

    const fields: Field[] = [];
    let names: Ident[] = [];
    let currentType: Expr | undefined;
    const flush = () => {
      if (!currentType || names.length === 0) return;
      fields.push({
        kind: "Field",
        names,
        type: currentType,
        span: mergeSpans(names[0]?.span, currentType.span)
      });
      names = [];
    };
    for (const param of params) {
      if (param.type !== currentType) {
        flush();
        currentType = param.type;
      }
      names.push(param.name ?? ident("_", param.type?.span));
    }
    flush();
    return fields;
  }

  private parseParamDecl(name0: Ident | undefined, typeSetsOK: boolean): ParamDecl {
    let name: Ident | undefined = name0;
    let type: Expr | undefined;

    if (name || isIdentifierLike(this.peek().kind)) {
      if (!name) name = this.parseIdent("expected parameter name or type");
      if (this.startsType() || this.at(TokenKind.LParen)) {
        type = this.parseType();
      } else if (this.match(TokenKind.Ellipsis)) {
        const dots = this.previous();
        type = { kind: "Ellipsis", element: this.parseType(), span: dots.span };
      } else if (this.match(TokenKind.Dot)) {
        const selector = this.parseIdent("expected selector in qualified type");
        type = {
          kind: "SelectorExpr",
          object: name,
          selector,
          span: mergeSpans(name.span, selector.span)
        };
        name = undefined;
      } else if (typeSetsOK && this.at(TokenKind.Or)) {
        type = this.embeddedElem(name);
        name = undefined;
      }
    } else if (this.startsType() || this.at(TokenKind.LParen)) {
      type = this.parseType();
    } else if (this.match(TokenKind.Ellipsis)) {
      const dots = this.previous();
      type = { kind: "Ellipsis", element: this.parseType(), span: dots.span };
    } else {
      this.error(`expected parameter name or type, found ${this.peek().lexeme || this.peek().kind}`, this.peek().span);
      this.advance();
      return { type: badExpr(this.previous().span) };
    }

    if (typeSetsOK && type && this.at(TokenKind.Or)) {
      type = this.embeddedElem(type);
    }
    return { ...(name ? { name } : {}), ...(type ? { type } : {}) };
  }

  private embeddedElem(initial: Expr): Expr {
    let expr = initial;
    while (this.match(TokenKind.Or)) {
      const operator = this.previous();
      const right = this.parseTypeTerm();
      expr = {
        kind: "BinaryExpr",
        left: expr,
        op: operator.kind as BinaryOperator,
        right,
        span: mergeSpans(expr.span, right.span)
      };
    }
    return expr;
  }

  private parseFieldList(
    open: TokenKind.LParen | TokenKind.LBrace | TokenKind.LBracket,
    close: TokenKind.RParen | TokenKind.RBrace | TokenKind.RBracket
  ): FieldList {
    const start = this.expect(open, `expected '${tokenDisplay(open)}'`);
    const fields: Field[] = [];
    while (!this.at(close) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(close)) break;
      fields.push(this.parseField(close));
      if (!this.match(TokenKind.Comma)) this.consumeSemi();
    }
    const end = this.expect(close, `expected '${tokenDisplay(close)}'`);
    return { kind: "FieldList", fields, span: mergeSpans(start.span, end.span) };
  }

  private parseField(close: TokenKind): Field {
    const start = this.peek();
    if (isIdentifierLike(this.peek().kind) && this.peek(1).kind === TokenKind.LParen) {
      const name = this.parseIdent("expected method name");
      const type = this.parseSignature(name.span ?? start.span);
      return {
        kind: "Field",
        names: [name],
        type,
        span: mergeSpans(name.span, type.span)
      };
    }

    if (isIdentifierLike(this.peek().kind) && this.fieldHasExplicitNames(close)) {
      const names = this.parseIdentList();
      const type = this.match(TokenKind.Ellipsis)
        ? ({ kind: "Ellipsis", element: this.parseType(), span: start.span } as Expr)
        : this.parseType();
      const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
      return {
        kind: "Field",
        names,
        type,
        ...(tag ? { tag } : {}),
        span: mergeSpans(names[0]?.span, tag?.span ?? type.span)
      };
    }

    const type = this.match(TokenKind.Ellipsis)
      ? ({ kind: "Ellipsis", element: this.parseType(), span: start.span } as Expr)
      : this.parseType();
    const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
    return { kind: "Field", names: [], type, ...(tag ? { tag } : {}), span: mergeSpans(type.span, tag?.span ?? type.span) };
  }

  private fieldHasExplicitNames(close: TokenKind): boolean {
    let offset = 0;
    if (!isIdentifierLike(this.peek(offset).kind)) return false;
    offset += 1;
    while (this.peek(offset).kind === TokenKind.Comma && isIdentifierLike(this.peek(offset + 1).kind)) {
      offset += 2;
    }
    const afterNames = this.peek(offset).kind;
    if (offset === 1 && afterNames === TokenKind.LBracket) {
      const afterBracket = this.kindAfterBalancedBrackets(offset);
      if (
        afterBracket === close ||
        afterBracket === TokenKind.Semicolon ||
        afterBracket === TokenKind.RBrace ||
        afterBracket === TokenKind.RParen ||
        afterBracket === TokenKind.RBracket ||
        afterBracket === TokenKind.StringLiteral ||
        afterBracket === TokenKind.Comma ||
        afterBracket === TokenKind.EOF
      ) {
        return false;
      }
    }
    if (afterNames === close || afterNames === TokenKind.Semicolon || afterNames === TokenKind.RBrace || afterNames === TokenKind.RParen || afterNames === TokenKind.RBracket) {
      return false;
    }
    return this.startsType(afterNames) || afterNames === TokenKind.Ellipsis;
  }

  private kindAfterBalancedBrackets(openOffset: number): TokenKind {
    let depth = 0;
    for (let offset = openOffset; ; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.EOF) return TokenKind.EOF;
      if (kind === TokenKind.LBracket) depth += 1;
      if (kind === TokenKind.RBracket) {
        depth -= 1;
        if (depth === 0) return this.peek(offset + 1).kind;
      }
    }
  }

  private parseIdentList(): Ident[] {
    const names = [this.parseIdent("expected identifier")];
    while (this.match(TokenKind.Comma)) {
      names.push(this.parseIdent("expected identifier after ','"));
    }
    return names;
  }

  private parseIdent(message: string): Ident {
    const token = this.peek();
    if (isIdentifierLike(token.kind)) {
      this.advance();
      return ident(token.lexeme, token.span);
    }
    this.error(message, token.span);
    this.advance();
    return ident("<missing>", token.span);
  }

  private expectCellAddress(message: string): { token: FrontToken; address?: CellAddress } {
    const token = this.peek();
    if (token.kind === TokenKind.CellAddress || token.kind === TokenKind.Identifier) {
      this.advance();
      const address = parseCellAddress(token.lexeme);
      if (address) return { token, address };
      this.error(message, token.span);
      return { token };
    }
    this.error(message, token.span);
    this.advance();
    return { token };
  }

  private expectBasicLit(kind: BasicLit["token"], message: string): BasicLit {
    const token = this.peek();
    const end = this.end(token);
    if (this.match(kind)) {
      return { kind: "BasicLit", token: kind, value: token.lexeme, span: mergeSpans(token.span, end) };
    }
    this.error(message, token.span);
    return { kind: "BasicLit", token: kind, value: "", span: token.span };
  }

  private looksLikeReceiver(): boolean {
    let depth = 0;
    for (let offset = 0; offset < 32; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.LParen) depth += 1;
      if (kind === TokenKind.RParen) {
        depth -= 1;
        if (depth === 0) return this.peek(offset + 1).kind === TokenKind.Identifier;
      }
      if (kind === TokenKind.EOF || kind === TokenKind.LBrace) return false;
    }
    return false;
  }

  private looksLikeRangeClause(): boolean {
    let parens = 0;
    let brackets = 0;
    let braces = 0;
    for (let offset = 0; offset < 64; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.LParen) parens += 1;
      else if (kind === TokenKind.RParen && parens > 0) parens -= 1;
      else if (kind === TokenKind.LBracket) brackets += 1;
      else if (kind === TokenKind.RBracket && brackets > 0) brackets -= 1;
      else if (kind === TokenKind.LBrace) {
        if (parens === 0 && brackets === 0 && braces === 0) return false;
        braces += 1;
      } else if (kind === TokenKind.RBrace && braces > 0) braces -= 1;
      if (kind === TokenKind.Range) return true;
      if ((kind === TokenKind.Semicolon || kind === TokenKind.EOF) && parens === 0 && brackets === 0 && braces === 0) return false;
    }
    return false;
  }

  private startsType(kind = this.peek().kind): boolean {
    return isIdentifierLike(kind) ||
      kind === TokenKind.Star ||
      kind === TokenKind.Tilde ||
      kind === TokenKind.LBracket ||
      kind === TokenKind.Map ||
      kind === TokenKind.Chan ||
      kind === TokenKind.Arrow ||
      kind === TokenKind.Struct ||
      kind === TokenKind.Interface ||
      kind === TokenKind.Func ||
      kind === TokenKind.LParen;
  }

  private withBareIdentifierComposites<T>(enabled: boolean, fn: () => T): T {
    const previous = this.allowBareIdentifierComposite;
    this.allowBareIdentifierComposite = enabled;
    try {
      return fn();
    } finally {
      this.allowBareIdentifierComposite = previous;
    }
  }

  private withControlClause<T>(fn: () => T): T {
    const previous = this.exprLev;
    this.exprLev = -1;
    try {
      return fn();
    } finally {
      this.exprLev = previous;
    }
  }

  private withExpressionLevel<T>(fn: () => T): T {
    const previous = this.exprLev;
    this.exprLev += 1;
    try {
      return fn();
    } finally {
      this.exprLev = previous;
    }
  }

  private withRhs<T>(enabled: boolean, fn: () => T): T {
    const previous = this.inRhs;
    this.inRhs = enabled;
    try {
      return fn();
    } finally {
      this.inRhs = previous;
    }
  }

  private withSpreadsheetRanges<T>(enabled: boolean, fn: () => T): T {
    const previous = this.allowSpreadsheetRanges;
    this.allowSpreadsheetRanges = enabled;
    try {
      return fn();
    } finally {
      this.allowSpreadsheetRanges = previous;
    }
  }

  public printTrace(...args: unknown[]): void {
    const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . ";
    const prefix = dots.slice(0, Math.min(dots.length, 2 * this.indent));
    globalThis.console?.log(`${this.peek().span?.line ?? 0}:${this.peek().span?.column ?? 0}: ${prefix}${args.join(" ")}`);
  }

  public init(_file: unknown, _text: string | Uint8Array, _mode: Mode): void {
  }

  private next0(): void {
    this.advance();
  }

  private next(): void {
    this.advance();
  }

  private lineFor(pos: SourceSpan | undefined): number {
    return pos?.line ?? 0;
  }

  private atComma(_context: string, _follow: TokenKind): boolean {
    if (this.at(TokenKind.Comma)) {
      return true;
    }
    if (this.at(TokenKind.Semicolon)) {
      this.error("missing ',' before newline in composite literal", this.peek().span);
      return true;
    }
    return false;
  }

  private errorExpected(pos: SourceSpan | undefined, msg: string): void {
    this.error(`expected ${msg}`, pos);
  }

  private expect2(kind: TokenKind): [SourceSpan | undefined, string] {
    const token = this.expect(kind, `expected ${tokenDisplay(kind)}`);
    return [token.span, token.lexeme];
  }

  private expectClosing(kind: TokenKind, context: string): SourceSpan | undefined {
    return this.expect(kind, `expected ${tokenDisplay(kind)} closing ${context}`).span;
  }

  private expectSemi(): void {
    switch (this.peek().kind) {
      case TokenKind.Semicolon:
      case TokenKind.RParen:
      case TokenKind.RBrace:
        this.consumeSemi();
        return;
      default:
        this.error("expected ';'", this.peek().span);
    }
  }

  private consumeComment(): void {
  }

  private consumeCommentGroup(): void {
  }

  private tokPrec(): [TokenKind, number] {
    let kind = this.peek().kind;
    if (this.inRhs && kind === TokenKind.Assign) {
      kind = TokenKind.Equal;
    }
    return [kind, binaryPrecedence(kind)];
  }

  private makeExpr(statement: Stmt | undefined, want: string): Expr | undefined {
    return this.statementExpression(statement, want);
  }

  private parseStmt(): Stmt {
    return this.parseStatement();
  }

  private parseStmtList(): Stmt[] {
    const list: Stmt[] = [];
    while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
      list.push(this.parseStatement());
      this.consumeSemi();
    }
    return list;
  }

  private parseBlockStmt(): BlockStmt {
    return this.parseBlock();
  }

  private parseBody(): BlockStmt {
    return this.parseBlock();
  }

  private parseReturnStmt(): Stmt {
    if (!this.at(TokenKind.Return)) {
      this.error("expected return", this.peek().span);
    }
    return this.parseStatement();
  }

  private parseBranchStmt(): Stmt {
    return this.parseStatement();
  }

  private parseDeferStmt(): Stmt {
    return this.parseStatement();
  }

  private parseGoStmtAlias(): Stmt {
    return this.parseGoStmt();
  }

  private parseExpr(): Expr {
    return this.parseExpression();
  }

  private parseRhs(): Expr {
    return this.parseRhsExpression();
  }

  private parseBinaryExpr(x?: Expr, prec1 = 1): Expr {
    return this.parseBinaryExpression(x ?? this.parseUnary(), prec1);
  }

  private parseUnaryExpr(): Expr {
    return this.parseUnary();
  }

  private parsePrimaryExpr(x?: Expr): Expr {
    return x ? this.parsePrimaryFrom(x) : this.parsePrimary();
  }

  private parseCallExpr(fun: Expr): Expr {
    return this.parsePrimaryFrom(fun);
  }

  private parseCallOrConversion(fun: Expr): Expr {
    return this.parsePrimaryFrom(fun);
  }

  private parseSelector(x: Expr): Expr {
    if (!this.at(TokenKind.Dot)) {
      return x;
    }
    const dot = this.advance();
    const selector = this.parseIdent("expected selector");
    return { kind: "SelectorExpr", object: x, selector, span: mergeSpans(x.span, selector.span ?? dot.span) };
  }

  private parseTypeAssertion(x: Expr): Expr {
    return this.parsePrimaryFrom(x);
  }

  private parseIndexOrSliceOrInstance(x: Expr): Expr {
    return this.parsePrimaryFrom(x);
  }

  private parseTypeInstance(x: Expr): Expr {
    return this.parsePrimaryFrom(x);
  }

  private parseElement(): Expr {
    return this.parseCompositeLiteralElement();
  }

  private parseElementList(): Expr[] {
    const list: Expr[] = [];
    while (!this.atAny(TokenKind.RBrace, TokenKind.EOF)) {
      list.push(this.parseElement());
      if (!this.match(TokenKind.Comma)) {
        break;
      }
    }
    return list;
  }

  private parseLiteralValue(type?: Expr): CompositeLit {
    const open = this.expect(TokenKind.LBrace, "expected literal value");
    const elements = this.parseElementList();
    const close = this.expect(TokenKind.RBrace, "expected '}' after literal value");
    return { kind: "CompositeLit", ...(type ? { type } : {}), elements, span: mergeSpans(open.span, close.span) };
  }

  private parseValue(): Expr {
    return this.parseRhsExpression();
  }

  private parseExprList(): Expr[] {
    return this.parseExpressionList(false);
  }

  private parseList(inRhs: boolean): Expr[] {
    return this.parseExpressionList(inRhs);
  }

  private parsePointerType(): Expr {
    const star = this.expect(TokenKind.Star, "expected '*'");
    const expr = this.parseType();
    return { kind: "StarExpr", expr, span: mergeSpans(star.span, expr.span) };
  }

  private parseDotsType(): Expr {
    const dots = this.expect(TokenKind.Ellipsis, "expected '...'");
    const element = this.parseType();
    return { kind: "Ellipsis", element, span: mergeSpans(dots.span, element.span) };
  }

  private parseArrayType(): ArrayType {
    const open = this.expect(TokenKind.LBracket, "expected '['");
    return this.parseArrayTypeAfterOpen(open);
  }

  private parseArrayFieldOrTypeInstance(name: Ident): [Ident | undefined, Expr] {
    const open = this.expect(TokenKind.LBracket, "expected '['");
    const type = this.parseArrayTypeAfterOpen(open, name);
    return [name, type];
  }

  private parseMapType(): Expr {
    const start = this.expect(TokenKind.Map, "expected map");
    this.expect(TokenKind.LBracket, "expected '[' after map");
    const key = this.parseType();
    this.expect(TokenKind.RBracket, "expected ']' after map key");
    const value = this.parseType();
    return { kind: "MapType", key, value, span: mergeSpans(start.span, value.span) };
  }

  private parseChanType(): Expr {
    return this.parseTypeTerm();
  }

  private parseQualifiedIdent(name?: Ident): Expr {
    const object = name ?? this.parseIdent("expected identifier");
    if (this.match(TokenKind.Dot)) {
      const selector = this.parseIdent("expected selector");
      return { kind: "SelectorExpr", object, selector, span: mergeSpans(object.span, selector.span) };
    }
    return object;
  }

  private tryIdentOrType(): Expr | undefined {
    if (isIdentifierLike(this.peek().kind)) {
      return this.parseTypeName();
    }
    if (this.startsType()) {
      return this.parseType();
    }
    return undefined;
  }

  private embeddedTerm(): Expr {
    return this.parseTypeTerm();
  }

  private parseFieldDecl(): Field {
    return this.parseField(TokenKind.RBrace);
  }

  private parseMethodSpec(): Field {
    return this.parseField(TokenKind.RBrace);
  }

  private parseParameters(acceptTParams = false): FieldList | undefined {
    if (this.at(TokenKind.LParen)) {
      return this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    }
    if (acceptTParams && this.at(TokenKind.LBracket)) {
      return this.parseFieldList(TokenKind.LBracket, TokenKind.RBracket);
    }
    return undefined;
  }

  private parseTypeParameters(): FieldList | undefined {
    return this.at(TokenKind.LBracket) ? this.parseTypeParamList() : undefined;
  }

  private consumeSemi(): void {
    this.match(TokenKind.Semicolon);
  }

  private skipSemis(): void {
    while (this.match(TokenKind.Semicolon)) {
      // consume all empty statements between declarations/statements
    }
  }

  private match(kind: TokenKind): boolean {
    if (!this.at(kind)) return false;
    this.advance();
    return true;
  }

  private expect(kind: TokenKind, message: string): FrontToken {
    if (this.at(kind)) return this.advance();
    const token = this.peek();
    this.error(message, token.span);
    return token;
  }

  private expectAny(kinds: TokenKind[], message: string): FrontToken {
    if (kinds.includes(this.peek().kind)) return this.advance();
    const token = this.peek();
    this.error(message, token.span);
    return token;
  }

  private at(kind: TokenKind): boolean {
    return this.peek().kind === kind;
  }

  private atAny(...kinds: TokenKind[]): boolean {
    const current = this.peek().kind;
    return kinds.includes(current);
  }

  private advance(): FrontToken {
    const token = this.peek();
    if (!this.at(TokenKind.EOF)) this.index += 1;
    return token;
  }

  private previous(): FrontToken {
    return this.tokens[Math.max(0, this.index - 1)] ?? this.peek();
  }

  private peek(ahead = 0): FrontToken {
    return this.tokens[this.index + ahead] ?? this.tokens[this.tokens.length - 1] ?? eofToken();
  }

  public currentSpan(): SourceSpan | undefined {
    return this.peek().span;
  }

  private end(token = this.peek()): SourceSpan | undefined {
    if (!token.span) return undefined;
    return {
      ...token.span,
      offset: token.span.offset + token.span.length,
      length: 0,
      column: token.span.column + token.span.length
    };
  }

  public error(message: string, span: SourceSpan | undefined, code = "GOJR_PARSE_FRONT001"): void {
    this.diagnostics.push({
      filename: diagnosticFilename(span, this.filename),
      code,
      severity: "error",
      message,
      ...(span ? { span } : {})
    });
  }
}

function binaryPrecedence(kind: TokenKind): number {
  switch (kind) {
    case TokenKind.OrOr: return 1;
    case TokenKind.AndAnd: return 2;
    case TokenKind.Equal:
    case TokenKind.NotEqual:
    case TokenKind.Less:
    case TokenKind.LessEqual:
    case TokenKind.Greater:
    case TokenKind.GreaterEqual:
      return 3;
    case TokenKind.Plus:
    case TokenKind.Minus:
    case TokenKind.Or:
    case TokenKind.Caret:
      return 4;
    case TokenKind.Star:
    case TokenKind.Slash:
    case TokenKind.Percent:
    case TokenKind.Shl:
    case TokenKind.Shr:
    case TokenKind.Amp:
    case TokenKind.BitClear:
      return 5;
    default:
      return 0;
  }
}

function extractName(expr: Expr, force: boolean): { name?: Ident; type?: Expr } {
  if (expr.kind === "Ident") {
    return { name: expr };
  }
  if (expr.kind === "BinaryExpr") {
    if (expr.op === TokenKind.Star && expr.left.kind === "Ident" && (force || isTypeElem(expr.right))) {
      return {
        name: expr.left,
        type: {
          kind: "StarExpr",
          expr: expr.right,
          span: mergeSpans(expr.left.span, expr.right.span)
        }
      };
    }
    if (expr.op === TokenKind.Or) {
      const split = extractName(expr.left, force || isTypeElem(expr.right));
      if (split.name && split.type) {
        return {
          name: split.name,
          type: {
            kind: "BinaryExpr",
            left: split.type,
            op: TokenKind.Or,
            right: expr.right,
            span: mergeSpans(split.type.span, expr.right.span)
          }
        };
      }
    }
  }
  if (expr.kind === "CallExpr" && expr.fun.kind === "Ident" && expr.args.length === 1 && !expr.ellipsis && (force || isTypeElem(expr.args[0]!))) {
    return {
      name: expr.fun,
      type: {
        kind: "ParenExpr",
        expr: expr.args[0]!,
        span: mergeSpans(expr.fun.span, expr.span)
      }
    };
  }
  return { type: expr };
}

function isTypeElem(expr: Expr): boolean {
  switch (expr.kind) {
    case "ArrayType":
    case "StructType":
    case "FuncType":
    case "InterfaceType":
    case "MapType":
    case "ChanType":
      return true;
    case "BinaryExpr":
      return isTypeElem(expr.left) || isTypeElem(expr.right);
    case "UnaryExpr":
      return expr.op === TokenKind.Tilde;
    case "ParenExpr":
      return isTypeElem(expr.expr);
    default:
      return false;
  }
}

function badExpr(span?: SourceSpan): Expr {
  return { kind: "BadExpr", ...(span ? { span } : {}) };
}

function mergeSpans(start: SourceSpan | undefined, end: SourceSpan | undefined): SourceSpan {
  if (!start && !end) return eofToken().span;
  if (!start) return end ?? eofToken().span;
  if (!end) return start;
  const endOffset = end.offset + end.length;
  return {
    filename: start.filename,
    offset: start.offset,
    length: Math.max(0, endOffset - start.offset),
    line: start.line,
    column: start.column
  };
}

function tokenDisplay(kind: TokenKind): string {
  switch (kind) {
    case TokenKind.LParen:
      return "(";
    case TokenKind.RParen:
      return ")";
    case TokenKind.LBrace:
      return "{";
    case TokenKind.RBrace:
      return "}";
    case TokenKind.LBracket:
      return "[";
    case TokenKind.RBracket:
      return "]";
    default:
      return kind;
  }
}

function eofToken(): FrontToken {
  return {
    kind: TokenKind.EOF,
    lexeme: "",
    span: { filename: REPL_FILENAME, offset: 0, length: 0, line: 1, column: 1 }
  };
}

function hasBytes(value: unknown): value is { Bytes(): Uint8Array } {
  return typeof value === "object" && value !== null && "Bytes" in value && typeof value.Bytes === "function";
}

function hasRead(value: unknown): value is { Read(): string | Uint8Array } {
  return typeof value === "object" && value !== null && "Read" in value && typeof value.Read === "function";
}

function diagnosticAsError(diagnostic: Diagnostic | undefined): Error | undefined {
  if (!diagnostic) {
    return undefined;
  }
  return new Error(`${diagnostic.filename}:${diagnostic.span?.line ?? 0}:${diagnostic.span?.column ?? 0}: ${diagnostic.message}`);
}

function hostReadFile(): ((filename: string) => string) | undefined {
  const host = globalThis as typeof globalThis & { __gojrReadFile?: (filename: string) => string };
  return host.__gojrReadFile;
}

function hostReadDir(): ((path: string) => Array<{ name: string; isFile: boolean }>) | undefined {
  const host = globalThis as typeof globalThis & { __gojrReadDir?: (path: string) => Array<{ name: string; isFile: boolean }> };
  return host.__gojrReadDir;
}
