import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan, diagnosticFilename } from "../diagnostics.js";
import {
  ArrayType as ArrayTypeNode,
  AstNode,
  BasicLit,
  BinaryExpr,
  BlockStmt,
  CallExpr,
  CellRefExpr,
  ChanType as ChanTypeNode,
  CompositeLit,
  Decl,
  Expr,
  Field,
  FieldList,
  File,
  FuncDecl,
  FuncLit,
  FuncType,
  GenDecl,
  Ident,
  IndexExpr,
  InterfaceType as InterfaceTypeNode,
  MapType as MapTypeNode,
  RangeRefExpr,
  SelectorExpr,
  Stmt,
  StructType as StructTypeNode,
  TypeAssertExpr,
  TypeSpec,
  UnaryExpr,
  ValueSpec
} from "./ast.js";
import { parseFrontSource, parseFrontSourceFiles } from "./parser.js";
import { TokenKind } from "./token.js";
import {
  ArrayType,
  BasicKind,
  BasicType,
  ChanType,
  BuiltinObject,
  ConstObject,
  FuncObject,
  InterfaceType,
  MapType,
  NamedType,
  newUniverse,
  ObjectKind,
  PackageInfo,
  PackageNameObject,
  PointerType,
  Scope,
  SignatureType,
  SliceType,
  StructType,
  Type,
  TypeParamType,
  TypeNameObject,
  TypeObject,
  TypeSetInfo,
  Universe,
  VarObject,
  assignableTo,
  emptyTypeSet,
  methodSet,
  tuple,
  varOf
} from "./types.js";

export type TypeMode = "invalid" | "constant" | "value" | "type" | "builtin" | "nil" | "package";

export interface TypeAndValue {
  mode: TypeMode;
  type: Type;
  value?: unknown;
}

export interface CheckInfo {
  types: Map<Expr, TypeAndValue>;
  defs: Map<Ident, TypeObject | undefined>;
  uses: Map<Ident, TypeObject>;
  scopes: Map<AstNode, Scope>;
  selections: Map<SelectorExpr, TypeObject>;
}

export interface Importer {
  import(path: string): PackageInfo | undefined;
}

export interface SheetNamespace {
  name?: string;
  cells?: Record<string, Type>;
  defaultType?: Type;
}

export interface CheckConfig {
  universe?: Universe;
  packagePath?: string;
  packageName?: string;
  importer?: Importer;
  sheetNamespaces?: Record<string, SheetNamespace>;
  predeclaredPackageObjects?: TypeObject[];
}

export interface CheckResult {
  pkg: PackageInfo;
  info: CheckInfo;
  diagnostics: Diagnostic[];
  universe: Universe;
}

export function checkFrontSource(source: string, filename: string, config: CheckConfig = {}): CheckResult & { file?: File } {
  const parsed = parseFrontSource(source, filename);
  const result = checkFrontFiles(parsed.file ? [parsed.file] : [], config, parsed.diagnostics, parsed.statements);
  return {
    ...result,
    ...(parsed.file ? { file: parsed.file } : {})
  };
}

export function checkFrontSourceFiles(sourceFiles: SourceFile[], config: CheckConfig = {}): CheckResult & { files: File[] } {
  const parsed = parseFrontSourceFiles(sourceFiles);
  const result = checkFrontFiles(parsed.files, config, parsed.diagnostics, parsed.statements);
  return {
    ...result,
    files: parsed.files
  };
}

export function checkFrontFiles(
  files: File[],
  config: CheckConfig = {},
  parserDiagnostics: Diagnostic[] = [],
  statements: Stmt[] = []
): CheckResult {
  const checker = new FrontChecker(config, parserDiagnostics);
  return checker.check(files, statements);
}

class FrontChecker {
  private readonly universe: Universe;
  private readonly diagnostics: Diagnostic[];
  private readonly info: CheckInfo = {
    types: new Map(),
    defs: new Map(),
    uses: new Map(),
    scopes: new Map(),
    selections: new Map()
  };
  private readonly predeclaredPackageObjectNames = new Set<string>();
  private readonly redeclaredPredeclaredFunctions = new Set<string>();
  private pkg: PackageInfo;

  public constructor(
    private readonly config: CheckConfig,
    parserDiagnostics: Diagnostic[]
  ) {
    this.universe = config.universe ?? newUniverse();
    this.diagnostics = [...parserDiagnostics];
    this.pkg = new PackageInfo(config.packagePath ?? "", config.packageName ?? "main", this.universe.scope);
  }

  public check(files: File[], statements: Stmt[] = []): CheckResult {
    const firstName = files.find((file) => file.name)?.name?.name;
    const packageName = this.config.packageName ?? firstName ?? "main";
    const packagePath = this.config.packagePath ?? packageName;
    this.pkg = new PackageInfo(packagePath, packageName, this.universe.scope);
    for (const object of this.config.predeclaredPackageObjects ?? []) {
      this.pkg.scope.insert(object);
      this.predeclaredPackageObjectNames.add(object.name);
    }

    for (const file of files) {
      this.info.scopes.set(file, this.pkg.scope);
      this.declareImports(file);
    }
    this.declareAutomaticImports();
    for (const file of files) this.declareTopLevelTypeNames(file);
    for (const file of files) this.declareTopLevelObjects(file);
    for (const file of files) this.checkTopLevelBodies(file);
    for (const statement of statements) this.checkStmt(statement, this.pkg.scope);

    return {
      pkg: this.pkg,
      info: this.info,
      diagnostics: this.diagnostics,
      universe: this.universe
    };
  }

  private declareImports(file: File): void {
    for (const spec of file.imports) {
      const path = unquote(spec.path.value);
      const imported = this.config.importer?.import(path) ?? standardPackageInfo(path, this.universe) ?? new PackageInfo(path, importName(path), this.universe.scope);
      const localName = spec.name?.name === "." ? imported.name : spec.name?.name ?? imported.name;
      const existing = this.pkg.scope.lookup(localName);
      if (existing instanceof PackageNameObject && existing.imported.path === imported.path) {
        if (spec.name) this.info.defs.set(spec.name, existing);
        continue;
      }
      const object = new PackageNameObject(localName, this.universe.basic.any, imported, this.pkg.scope, this.pkg);
      this.insert(this.pkg.scope, object, spec.name ?? spec.path);
      if (spec.name) this.info.defs.set(spec.name, object);
    }
  }

  private declareAutomaticImports(): void {
    if (this.pkg.scope.lookup("fmt")) return;
    const fmt = fmtPackageInfo(this.universe);
    this.insert(this.pkg.scope, new PackageNameObject("fmt", this.universe.basic.any, fmt, this.pkg.scope, this.pkg));
  }

  private declareTopLevelTypeNames(file: File): void {
    for (const declaration of file.declarations) {
      if (declaration.kind !== "GenDecl" || declaration.token !== TokenKind.Type) continue;
      for (const spec of declaration.specs) {
        if (spec.kind !== "TypeSpec") continue;
        const placeholder = new TypeNameObject(spec.name.name, this.universe.basic.invalid, this.pkg.scope, this.pkg);
        const named = new NamedType(placeholder, this.universe.basic.invalid);
        placeholder.setType(named);
        this.insert(this.pkg.scope, placeholder, spec.name);
      }
    }
  }

  private declareTopLevelObjects(file: File): void {
    for (const declaration of file.declarations) {
      if (declaration.kind === "GenDecl") {
        this.declareGenDecl(declaration, this.pkg.scope, true);
      } else if (declaration.kind === "FuncDecl") {
        this.declareFuncDecl(declaration);
      }
    }
  }

  private declareGenDecl(declaration: GenDecl, scope: Scope, topLevel: boolean): void {
    for (const spec of declaration.specs) {
      if (spec.kind === "TypeSpec") {
        this.bindTypeSpec(spec, scope);
      } else if (spec.kind === "ValueSpec") {
        this.bindValueSpec(spec, scope, declaration.token, topLevel);
      }
    }
  }

  private bindTypeSpec(spec: TypeSpec, scope: Scope): void {
    const object = scope.lookup(spec.name.name);
    const typeName = object instanceof TypeNameObject
      ? object
      : new TypeNameObject(spec.name.name, this.universe.basic.invalid, scope, this.pkg);
    if (!(object instanceof TypeNameObject)) this.insert(scope, typeName, spec.name);

    const typeScope = spec.typeParams ? this.scopeWithTypeParams(spec.typeParams, scope) : scope;
    const underlying = this.resolveType(spec.type, typeScope);
    if (typeName.type instanceof NamedType && !spec.alias) {
      typeName.type.setUnderlying(underlying);
    } else {
      typeName.setType(spec.alias ? underlying : new NamedType(typeName, underlying));
    }
    this.info.defs.set(spec.name, typeName);
  }

  private bindValueSpec(spec: ValueSpec, scope: Scope, token: TokenKind, topLevel: boolean): void {
    const explicitType = spec.type ? this.resolveType(spec.type, scope) : undefined;
    for (const [index, name] of spec.names.entries()) {
      const valueType = spec.values[index] ? this.checkExpr(spec.values[index]!, scope).type : undefined;
      const type = explicitType ?? (token === TokenKind.Const ? valueType : defaultType(valueType, this.universe)) ?? this.universe.basic.invalid;
      const object = token === TokenKind.Const
        ? new ConstObject(name.name, type, undefined, scope, this.pkg)
        : new VarObject(name.name, type, false, scope, this.pkg);
      this.insert(scope, object, name);
      if (!topLevel && spec.values[index]) {
        this.checkAssignable(spec.values[index]!, type, scope);
      }
    }
  }

  private declareFuncDecl(declaration: FuncDecl): void {
    const receiver = declaration.receiver ? this.receiverVar(declaration.receiver) : undefined;
    const signature = this.signatureFromFuncType(declaration.type, this.pkg.scope, receiver);
    const object = new FuncObject(declaration.name.name, signature, this.pkg.scope, this.pkg);
    if (receiver) {
      const named = this.receiverNamedType(receiver.type);
      if (named) named.addMethod(object);
      this.info.defs.set(declaration.name, object);
      return;
    }
    const existing = this.pkg.scope.lookup(declaration.name.name);
    if (existing) {
      const canReplacePredeclaredFunction = this.predeclaredPackageObjectNames.has(declaration.name.name) &&
        !this.redeclaredPredeclaredFunctions.has(declaration.name.name) &&
        existing instanceof FuncObject &&
        sameSignature(existing.signature, signature);
      if (canReplacePredeclaredFunction) {
        this.redeclaredPredeclaredFunctions.add(declaration.name.name);
        this.info.defs.set(declaration.name, existing);
        return;
      }
      this.info.defs.set(declaration.name, undefined);
      if (this.predeclaredPackageObjectNames.has(declaration.name.name) && existing instanceof FuncObject) {
        this.error(`cannot redeclare ${declaration.name.name} with different signature`, declaration.name.span);
        return;
      }
      this.error(`${declaration.name.name} already declared`, declaration.name.span);
      return;
    }
    this.insert(this.pkg.scope, object, declaration.name);
  }

  private checkTopLevelBodies(file: File): void {
    for (const declaration of file.declarations) {
      if (declaration.kind === "FuncDecl" && declaration.body) {
        const receiver = declaration.receiver ? this.receiverVar(declaration.receiver) : undefined;
        const parent = declaration.type.typeParams ? this.scopeWithTypeParams(declaration.type.typeParams, this.pkg.scope) : this.pkg.scope;
        const signature = this.signatureFromFuncType(declaration.type, parent, receiver);
        const scope = new Scope(parent, `func ${declaration.name.name}`);
        this.info.scopes.set(declaration.body, scope);
        if (receiver && declaration.receiver) this.insert(scope, receiver, firstName(declaration.receiver));
        for (const param of signature.params.variables) this.insert(scope, param);
        for (const result of signature.results.variables) {
          if (result.name) this.insert(scope, result);
        }
        this.checkBlock(declaration.body, scope, signature);
      }
    }
  }

  private checkBlock(block: BlockStmt, parent: Scope, signature?: SignatureType): void {
    const scope = this.info.scopes.get(block) ?? new Scope(parent, "block");
    this.info.scopes.set(block, scope);
    for (const statement of block.statements) this.checkStmt(statement, scope, signature);
  }

  private checkStmt(statement: Stmt, scope: Scope, signature?: SignatureType): void {
    switch (statement.kind) {
      case "DeclStmt":
        if (statement.decl.kind === "GenDecl") this.declareGenDecl(statement.decl, scope, false);
        break;
      case "BlockStmt":
        this.checkBlock(statement, new Scope(scope, "block"), signature);
        break;
      case "LabeledStmt":
        this.checkStmt(statement.stmt, scope, signature);
        break;
      case "ExprStmt":
        this.checkExpr(statement.expr, scope);
        break;
      case "AssignStmt":
        this.checkAssign(statement, scope);
        break;
      case "IncDecStmt":
        this.checkExpr(statement.expr, scope);
        break;
      case "ReturnStmt":
        this.checkReturn(statement.results, scope, signature);
        break;
      case "IfStmt":
        if (statement.init) this.checkStmt(statement.init, scope, signature);
        this.checkExpr(statement.condition, scope);
        this.checkBlock(statement.body, new Scope(scope, "if"), signature);
        if (statement.else) this.checkStmt(statement.else, scope, signature);
        break;
      case "ForStmt":
        if (statement.init) this.checkStmt(statement.init, scope, signature);
        if (statement.condition) this.checkExpr(statement.condition, scope);
        if (statement.post) this.checkStmt(statement.post, scope, signature);
        this.checkBlock(statement.body, new Scope(scope, "for"), signature);
        break;
      case "RangeStmt":
        this.checkRange(statement, scope, signature);
        break;
      case "SwitchStmt":
        if (statement.init) this.checkStmt(statement.init, scope, signature);
        if (statement.tag) this.checkExpr(statement.tag, scope);
        for (const clause of statement.body) {
          for (const item of clause.list) this.checkExpr(item, scope);
          for (const item of clause.body) this.checkStmt(item, scope, signature);
        }
        break;
      case "TypeSwitchStmt":
        if (statement.init) this.checkStmt(statement.init, scope, signature);
        this.checkStmt(statement.assign, scope, signature);
        for (const clause of statement.body) {
          for (const item of clause.list) this.resolveType(item, scope);
          for (const item of clause.body) this.checkStmt(item, scope, signature);
        }
        break;
      case "DeferStmt":
        this.checkExpr(statement.call, scope);
        break;
      case "GoStmt":
        this.checkExpr(statement.call, scope);
        break;
      case "SendStmt":
        this.checkSend(statement, scope);
        break;
      case "SelectStmt":
        for (const clause of statement.body) {
          if (clause.comm) this.checkStmt(clause.comm, scope, signature);
          for (const item of clause.body) this.checkStmt(item, scope, signature);
        }
        break;
      default:
        break;
    }
  }

  private checkSend(statement: Extract<Stmt, { kind: "SendStmt" }>, scope: Scope): void {
    const channel = this.checkExpr(statement.channel, scope).type.underlying();
    const value = this.checkExpr(statement.value, scope).type;
    if (!(channel instanceof ChanType)) {
      const hint = value.underlying() instanceof ChanType
        ? "; `<-` between expressions is a send, use `x = <-ch` to receive into an existing variable"
        : "";
      this.error(`cannot send to non-channel ${channel.typeString()}${hint}`, statement.channel.span);
      return;
    }
    if (channel.direction === "receive") {
      this.error(`cannot send to receive-only channel ${channel.typeString()}`, statement.channel.span);
      return;
    }
    if (!assignableTo(value, channel.element)) {
      this.error(`cannot send ${value.typeString()} as ${channel.element.typeString()}`, statement.value.span);
    }
  }

  private checkAssign(statement: Extract<Stmt, { kind: "AssignStmt" }>, scope: Scope): void {
    const rhs = statement.rhs.map((expr) => this.checkExpr(expr, scope).type);
    for (const [index, lhs] of statement.lhs.entries()) {
      if (statement.token === TokenKind.Define && lhs.kind === "Ident") {
        if (lhs.name === "_") continue;
        const type = defaultType(rhs[index] ?? rhs[0], this.universe) ?? this.universe.basic.invalid;
        this.insert(scope, new VarObject(lhs.name, type, false, scope, this.pkg), lhs);
        continue;
      }
      if (lhs.kind === "Ident" && lhs.name === "_") continue;
      const lhsType = this.checkExpr(lhs, scope).type;
      if (statement.rhs[index]) this.checkAssignable(statement.rhs[index]!, lhsType, scope);
    }
  }

  private checkRange(statement: Extract<Stmt, { kind: "RangeStmt" }>, scope: Scope, signature?: SignatureType): void {
    const source = this.checkExpr(statement.source, scope).type.underlying();
    const keyType = this.universe.basic.int64;
    let valueType: Type = this.universe.basic.any;
    if (source instanceof SliceType || source instanceof ArrayType) valueType = source.element;
    if (source instanceof MapType) valueType = source.value;
    if (source instanceof BasicType && source.basicKind === BasicKind.String) valueType = this.universe.basic.string;

    if (statement.token === TokenKind.Define) {
      if (statement.key?.kind === "Ident" && statement.key.name !== "_") this.insert(scope, new VarObject(statement.key.name, keyType, false, scope, this.pkg), statement.key);
      if (statement.value?.kind === "Ident" && statement.value.name !== "_") this.insert(scope, new VarObject(statement.value.name, valueType, false, scope, this.pkg), statement.value);
    } else {
      if (statement.key) this.checkExpr(statement.key, scope);
      if (statement.value) this.checkExpr(statement.value, scope);
    }
    this.checkBlock(statement.body, new Scope(scope, "range"), signature);
  }

  private checkReturn(results: Expr[], scope: Scope, signature?: SignatureType): void {
    const expected = signature?.results.variables ?? [];
    for (const [index, expr] of results.entries()) {
      const actual = this.checkExpr(expr, scope).type;
      const target = expected[index]?.type;
      if (target && !assignableTo(actual, target)) {
        this.error(`cannot use ${actual.typeString()} as ${target.typeString()} in return`, expr.span);
      }
    }
  }

  private checkExpr(expr: Expr, scope: Scope): TypeAndValue {
    const cached = this.info.types.get(expr);
    if (cached) return cached;

    let result: TypeAndValue;
    switch (expr.kind) {
      case "Ident":
        result = this.checkIdent(expr, scope);
        break;
      case "BasicLit":
        result = this.checkBasicLit(expr);
        break;
      case "CellRefExpr":
        result = { mode: "value", type: this.cellType(expr) };
        break;
      case "RangeRefExpr":
        result = { mode: "value", type: new SliceType(this.rangeElementType(expr)) };
        break;
      case "ParenExpr":
        result = this.checkExpr(expr.expr, scope);
        break;
      case "UnaryExpr":
        result = this.checkUnary(expr, scope);
        break;
      case "BinaryExpr":
        result = this.checkBinary(expr, scope);
        break;
      case "SelectorExpr":
        result = this.checkSelector(expr, scope);
        break;
      case "CallExpr":
        result = this.checkCall(expr, scope);
        break;
      case "IndexExpr":
        result = this.checkIndex(expr, scope);
        break;
      case "SliceExpr":
        result = { mode: "value", type: this.checkExpr(expr.object, scope).type };
        if (expr.low) this.checkExpr(expr.low, scope);
        if (expr.high) this.checkExpr(expr.high, scope);
        if (expr.max) this.checkExpr(expr.max, scope);
        break;
      case "TypeAssertExpr":
        result = this.checkTypeAssert(expr, scope);
        break;
      case "CompositeLit":
        result = this.checkCompositeLit(expr, scope);
        break;
      case "FuncLit":
        result = this.checkFuncLit(expr, scope);
        break;
      case "ArrayType":
      case "StructType":
      case "FuncType":
      case "InterfaceType":
      case "MapType":
      case "ChanType":
      case "StarExpr":
      case "Ellipsis":
        result = { mode: "type", type: this.resolveType(expr, scope) };
        break;
      default:
        result = { mode: "invalid", type: this.universe.basic.invalid };
        break;
    }
    this.info.types.set(expr, result);
    return result;
  }

  private checkIdent(expr: Ident, scope: Scope): TypeAndValue {
    const found = scope.lookupParent(expr.name);
    if (!found) {
      this.error(`${expr.name} is not declared`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    this.info.uses.set(expr, found.object);
    switch (found.object.kind) {
      case ObjectKind.TypeName:
        return { mode: "type", type: found.object.type };
      case ObjectKind.Const:
        return { mode: "constant", type: found.object.type, value: found.object instanceof ConstObject ? found.object.value : undefined };
      case ObjectKind.Builtin:
        return { mode: "builtin", type: found.object.type };
      case ObjectKind.Nil:
        return { mode: "nil", type: found.object.type };
      case ObjectKind.PackageName:
        return { mode: "package", type: found.object.type };
      default:
        return { mode: "value", type: found.object.type };
    }
  }

  private checkBasicLit(expr: BasicLit): TypeAndValue {
    if (expr.token === TokenKind.IntLiteral) return { mode: "constant", type: this.universe.basic.untypedInt, value: parseGoIntLiteral(expr.value) };
    if (expr.token === TokenKind.FloatLiteral) return { mode: "constant", type: this.universe.basic.untypedFloat, value: parseGoFloatLiteral(expr.value) };
    if (expr.token === TokenKind.ImagLiteral) return { mode: "constant", type: this.universe.basic.untypedComplex, value: expr.value };
    if (expr.token === TokenKind.RuneLiteral) return { mode: "constant", type: this.universe.basic.untypedInt, value: parseGoRuneLiteral(expr.value) };
    return { mode: "constant", type: this.universe.basic.untypedString, value: unquote(expr.value) };
  }

  private checkUnary(expr: UnaryExpr, scope: Scope): TypeAndValue {
    const operand = this.checkExpr(expr.expr, scope);
    if (expr.op === TokenKind.Amp) return { mode: "value", type: new PointerType(operand.type) };
    if (expr.op === TokenKind.Arrow) {
      const channel = operand.type.underlying();
      if (!(channel instanceof ChanType)) {
        this.error(`cannot receive from non-channel ${channel.typeString()}`, expr.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
      }
      if (channel.direction === "send") {
        this.error(`cannot receive from send-only channel ${channel.typeString()}`, expr.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
      }
      return { mode: "value", type: channel.element };
    }
    return { mode: operand.mode === "constant" ? "constant" : "value", type: operand.type };
  }

  private checkBinary(expr: BinaryExpr, scope: Scope): TypeAndValue {
    const left = this.checkExpr(expr.left, scope).type;
    const right = this.checkExpr(expr.right, scope).type;
    const typeParam = sameTypeParam(left, right);
    if (typeParam) return this.checkTypeParamBinary(expr, typeParam);
    if ([TokenKind.Equal, TokenKind.NotEqual, TokenKind.Less, TokenKind.LessEqual, TokenKind.Greater, TokenKind.GreaterEqual].includes(expr.op)) {
      return { mode: "value", type: this.universe.basic.bool };
    }
    if (expr.op === TokenKind.AndAnd || expr.op === TokenKind.OrOr) return { mode: "value", type: this.universe.basic.bool };
    if (expr.op === TokenKind.Plus) {
      if (isStringLike(left) || isStringLike(right)) {
        if (isStringLike(left) && isStringLike(right)) {
          return { mode: "value", type: isUntyped(left) && isUntyped(right) ? this.universe.basic.untypedString : this.universe.basic.string };
        }
        this.error(`invalid operation: ${left.typeString()} + ${right.typeString()} (mismatched types ${left.typeString()} and ${right.typeString()})`, expr.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
      }
      if (!isNumericLike(left) || !isNumericLike(right)) {
        this.error(`invalid operation: ${left.typeString()} + ${right.typeString()} (operator + not defined for those types)`, expr.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
      }
    }
    return { mode: isUntyped(left) && isUntyped(right) ? "constant" : "value", type: promoteNumeric(left, right, this.universe) };
  }

  private checkTypeParamBinary(expr: BinaryExpr, type: TypeParamType): TypeAndValue {
    const constraint = type.constraint;
    if ([TokenKind.Equal, TokenKind.NotEqual].includes(expr.op)) {
      if (constraint.comparable) return { mode: "value", type: this.universe.basic.bool };
      this.error(`invalid operation: ${type.typeString()} ${operatorText(expr.op)} ${type.typeString()} (type parameter is not comparable)`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    if ([TokenKind.Less, TokenKind.LessEqual, TokenKind.Greater, TokenKind.GreaterEqual].includes(expr.op)) {
      if (constraint.ordered) return { mode: "value", type: this.universe.basic.bool };
      this.error(`invalid operation: ${type.typeString()} ${operatorText(expr.op)} ${type.typeString()} (type parameter is not ordered)`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    if (expr.op === TokenKind.Plus) {
      if (constraint.addable) return { mode: "value", type };
      this.error(`invalid operation: ${type.typeString()} + ${type.typeString()} (operator + not permitted by constraint)`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    if ([TokenKind.Minus, TokenKind.Star, TokenKind.Slash].includes(expr.op)) {
      if (constraint.numeric) return { mode: "value", type };
      this.error(`invalid operation: ${type.typeString()} ${operatorText(expr.op)} ${type.typeString()} (operator not permitted by constraint)`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    if ([TokenKind.Percent, TokenKind.Or, TokenKind.Caret, TokenKind.Amp, TokenKind.BitClear, TokenKind.Shl, TokenKind.Shr].includes(expr.op)) {
      if (constraint.integer) return { mode: "value", type };
      this.error(`invalid operation: ${type.typeString()} ${operatorText(expr.op)} ${type.typeString()} (integer operator not permitted by constraint)`, expr.span);
      return { mode: "invalid", type: this.universe.basic.invalid };
    }
    return { mode: "value", type };
  }

  private checkSelector(expr: SelectorExpr, scope: Scope): TypeAndValue {
    if (expr.object.kind === "Ident") {
      const object = scope.lookupParent(expr.object.name)?.object;
      if (object instanceof PackageNameObject) {
        this.info.uses.set(expr.object, object);
        const selected = object.imported.scope.lookup(expr.selector.name);
        if (selected) {
          this.info.uses.set(expr.selector, selected);
          this.info.selections.set(expr, selected);
          return { mode: selected.kind === ObjectKind.TypeName ? "type" : "value", type: selected.type };
        }
        if (object.imported.scope.names().length === 0) return { mode: "value", type: this.universe.basic.any };
      }
    }

    const base = this.checkExpr(expr.object, scope).type;
    const field = lookupFieldOrMethod(base, expr.selector.name);
    if (field) {
      this.info.uses.set(expr.selector, field);
      this.info.selections.set(expr, field);
      return { mode: field.kind === ObjectKind.TypeName ? "type" : "value", type: field.type };
    }

    this.error(`${expr.selector.name} is not a field or method`, expr.selector.span);
    return { mode: "invalid", type: this.universe.basic.invalid };
  }

  private checkCall(expr: CallExpr, scope: Scope): TypeAndValue {
    const callee = this.checkExpr(expr.fun, scope);
    for (const arg of expr.args) this.checkExpr(arg, scope);
    if (callee.mode === "type") return { mode: "value", type: callee.type };
    const builtin = expr.fun.kind === "Ident" ? expr.fun.name : undefined;
    const special = builtin ? this.checkBuiltinCall(builtin, expr, scope) : undefined;
    if (special) return special;
    const fun = callee.type;
    if (fun instanceof SignatureType) {
      const results = fun.results.variables;
      if (results.length === 0) return { mode: "value", type: this.universe.basic.untypedNil };
      if (results.length === 1) return { mode: "value", type: results[0]?.type ?? this.universe.basic.invalid };
      return { mode: "value", type: fun.results };
    }
    if (fun instanceof BasicType && fun.basicKind === BasicKind.Any) return { mode: "value", type: this.universe.basic.any };
    this.error("cannot call non-function value", expr.fun.span);
    return { mode: "invalid", type: this.universe.basic.invalid };
  }

  private checkBuiltinCall(name: string, expr: CallExpr, scope: Scope): TypeAndValue | undefined {
    switch (name) {
      case "make":
        return { mode: "value", type: expr.args[0] ? this.resolveType(expr.args[0], scope) : this.universe.basic.invalid };
      case "new":
        return { mode: "value", type: new PointerType(expr.args[0] ? this.resolveType(expr.args[0], scope) : this.universe.basic.invalid) };
      case "append":
        return { mode: "value", type: expr.args[0] ? this.checkExpr(expr.args[0], scope).type : this.universe.basic.invalid };
      case "len":
      case "cap":
      case "copy":
        return { mode: "value", type: this.universe.basic.int64 };
      case "complex":
        return { mode: "value", type: this.universe.basic.complex128 };
      case "real":
      case "imag":
        return { mode: "value", type: this.universe.basic.float64 };
      case "recover":
        return { mode: "value", type: this.universe.basic.any };
      case "min":
      case "max": {
        const first = expr.args[0] ? this.checkExpr(expr.args[0], scope).type : this.universe.basic.invalid;
        const second = expr.args[1] ? this.checkExpr(expr.args[1], scope).type : first;
        return { mode: "value", type: defaultType(promoteNumeric(first, second, this.universe), this.universe) ?? this.universe.basic.invalid };
      }
      case "panic":
      case "panicOn":
      case "print":
      case "println":
      case "delete":
      case "clear":
      case "close":
        return { mode: "value", type: this.universe.basic.untypedNil };
      default:
        return undefined;
    }
  }

  private checkIndex(expr: IndexExpr, scope: Scope): TypeAndValue {
    const objectValue = this.checkExpr(expr.object, scope);
    const object = objectValue.type.underlying();
    if (object instanceof SignatureType) {
      const typeArg = this.resolveType(expr.index, scope);
      if (typeArg instanceof BasicType && typeArg.basicKind === BasicKind.Invalid) {
        return { mode: "invalid", type: this.universe.basic.invalid };
      }
      return { mode: "value", type: objectValue.type };
    }
    this.checkExpr(expr.index, scope);
    if (object instanceof ArrayType || object instanceof SliceType) return { mode: "value", type: object.element };
    if (object instanceof MapType) return { mode: "value", type: object.value };
    if (object instanceof BasicType && object.basicKind === BasicKind.String) return { mode: "value", type: this.universe.basic.string };
    this.error("cannot index value", expr.span);
    return { mode: "invalid", type: this.universe.basic.invalid };
  }

  private checkTypeAssert(expr: TypeAssertExpr, scope: Scope): TypeAndValue {
    this.checkExpr(expr.object, scope);
    if (expr.typeSwitch) return { mode: "value", type: this.universe.basic.any };
    const type = expr.type ? this.resolveType(expr.type, scope) : this.universe.basic.any;
    return { mode: "value", type };
  }

  private checkCompositeLit(expr: CompositeLit, scope: Scope): TypeAndValue {
    const type = expr.type ? this.resolveType(expr.type, scope) : this.universe.basic.invalid;
    const literalType = type.underlying();
    for (const element of expr.elements) {
      if (element.kind === "KeyValueExpr") {
        if (!(literalType instanceof StructType && element.key.kind === "Ident")) {
          this.checkExpr(element.key, scope);
        }
        this.checkExpr(element.value, scope);
      } else {
        this.checkExpr(element, scope);
      }
    }
    return { mode: "value", type };
  }

  private checkFuncLit(expr: FuncLit, scope: Scope): TypeAndValue {
    const signature = this.signatureFromFuncType(expr.type, scope);
    const child = new Scope(scope, "func literal");
    for (const param of signature.params.variables) this.insert(child, param);
    for (const result of signature.results.variables) {
      if (result.name) this.insert(child, result);
    }
    this.checkBlock(expr.body, child, signature);
    return { mode: "value", type: signature };
  }

  private resolveType(expr: Expr, scope: Scope): Type {
    switch (expr.kind) {
      case "Ident": {
        const object = scope.lookupParent(expr.name)?.object;
        if (object?.kind === ObjectKind.TypeName) {
          this.info.uses.set(expr, object);
          return object.type;
        }
        this.error(`${expr.name} is not a type`, expr.span);
        return this.universe.basic.invalid;
      }
      case "SelectorExpr": {
        const selected = this.checkSelector(expr, scope);
        return selected.type;
      }
      case "IndexExpr":
        this.resolveType(expr.index, scope);
        return this.resolveType(expr.object, scope);
      case "UnaryExpr":
        if (expr.op === TokenKind.Tilde) return this.resolveType(expr.expr, scope);
        this.error("expected type", expr.span);
        return this.universe.basic.invalid;
      case "BinaryExpr":
        if (expr.op === TokenKind.Or) {
          this.resolveType(expr.left, scope);
          this.resolveType(expr.right, scope);
          return this.universe.basic.any;
        }
        this.error("expected type", expr.span);
        return this.universe.basic.invalid;
      case "StarExpr":
        return new PointerType(this.resolveType(expr.expr, scope));
      case "ArrayType":
        return this.resolveArrayType(expr, scope);
      case "MapType":
        return new MapType(this.resolveType(expr.key, scope), this.resolveType(expr.value, scope));
      case "ChanType":
        return this.resolveChanType(expr, scope);
      case "StructType":
        return this.resolveStructType(expr, scope);
      case "InterfaceType":
        return this.resolveInterfaceType(expr, scope);
      case "FuncType":
        return this.signatureFromFuncType(expr, scope);
      case "Ellipsis":
        return new SliceType(expr.element ? this.resolveType(expr.element, scope) : this.universe.basic.invalid);
      default:
        this.error("expected type", expr.span);
        return this.universe.basic.invalid;
    }
  }

  private resolveArrayType(expr: ArrayTypeNode, scope: Scope): Type {
    const element = this.resolveType(expr.element, scope);
    if (!expr.length) return new SliceType(element);
    if (expr.length.kind === "BasicLit" && expr.length.token === TokenKind.IntLiteral) {
      return new ArrayType(Number(expr.length.value), element);
    }
    this.checkExpr(expr.length, scope);
    return new ArrayType(expr.inferredLength ? -1 : 0, element);
  }

  private resolveChanType(expr: ChanTypeNode, scope: Scope): Type {
    return new ChanType(this.resolveType(expr.value, scope), expr.direction);
  }

  private resolveStructType(expr: StructTypeNode, scope: Scope): Type {
    const fields = expr.fields.fields.flatMap((field) => {
      const type = this.resolveType(field.type, scope);
      const tag = field.tag ? unquote(field.tag.value) : undefined;
      if (field.names.length === 0) return [{ name: embeddedFieldName(field.type), type, embedded: true, ...(tag ? { tag } : {}) }];
      return field.names.map((name) => ({ name: name.name, type, embedded: false, ...(tag ? { tag } : {}) }));
    });
    return new StructType(fields);
  }

  private resolveInterfaceType(expr: InterfaceTypeNode, scope: Scope): Type {
    const typeSets: TypeSetInfo[] = [];
    const methods = expr.methods.fields.flatMap((field): FuncObject[] => {
      if (field.type.kind !== "FuncType") {
        typeSets.push(this.typeSetInfoForExpr(field.type, scope));
        const embedded = this.resolveType(field.type, scope).underlying();
        return embedded instanceof InterfaceType ? embedded.methods : [];
      }
      return field.names.map((name) => new FuncObject(name.name, this.signatureFromFuncType(field.type as FuncType, scope), scope, this.pkg));
    });
    return new InterfaceType(methods, intersectTypeSets(typeSets)).complete();
  }

  private signatureFromFuncType(expr: FuncType, scope: Scope, receiver?: VarObject): SignatureType {
    const typeScope = expr.typeParams ? this.scopeWithTypeParams(expr.typeParams, scope) : scope;
    const params = this.fieldListTuple(expr.params, typeScope, true);
    const results = expr.results ? this.fieldListTuple(expr.results, typeScope, false) : tuple();
    const variadic = params.variables[params.variables.length - 1]?.type instanceof SliceType &&
      expr.params.fields[expr.params.fields.length - 1]?.type.kind === "Ellipsis";
    return new SignatureType(receiver, params, results, variadic);
  }

  private scopeWithTypeParams(typeParams: FieldList, parent: Scope): Scope {
    const scope = new Scope(parent, "type parameters");
    for (const field of typeParams.fields) {
      const constraint = this.typeSetInfoForExpr(field.type, parent);
      for (const name of field.names) {
        this.insert(scope, new TypeNameObject(name.name, new TypeParamType(name.name, constraint), scope, this.pkg), name);
      }
    }
    return scope;
  }

  private fieldListTuple(list: FieldList, scope: Scope, params: boolean): ReturnType<typeof tuple> {
    const variables: VarObject[] = [];
    for (const field of list.fields) {
      const type = this.fieldType(field, scope);
      if (field.names.length === 0) {
        variables.push(varOf("", type));
      } else {
        for (const name of field.names) {
          variables.push(varOf(name.name, type));
        }
      }
    }
    if (params) return tuple(...variables);
    return tuple(...variables);
  }

  private fieldType(field: Field, scope: Scope): Type {
    if (field.type.kind === "Ellipsis") {
      return new SliceType(field.type.element ? this.resolveType(field.type.element, scope) : this.universe.basic.invalid);
    }
    return this.resolveType(field.type, scope);
  }

  private typeSetInfoForExpr(expr: Expr, scope: Scope): TypeSetInfo {
    if (expr.kind === "UnaryExpr" && expr.op === TokenKind.Tilde) return this.typeSetInfoForExpr(expr.expr, scope);
    if (expr.kind === "BinaryExpr" && expr.op === TokenKind.Or) {
      return intersectTypeSets([
        this.typeSetInfoForExpr(expr.left, scope),
        this.typeSetInfoForExpr(expr.right, scope)
      ]);
    }
    return this.typeSetInfoForType(this.resolveType(expr, scope));
  }

  private typeSetInfoForType(type: Type): TypeSetInfo {
    if (type instanceof TypeParamType) return type.constraint;
    if (type instanceof NamedType) return this.typeSetInfoForType(type.underlying());
    if (type instanceof InterfaceType) return type.typeSet;
    if (type instanceof BasicType) return basicTypeSetInfo(type);
    return emptyTypeSet();
  }

  private receiverVar(list: FieldList): VarObject | undefined {
    const field = list.fields[0];
    if (!field) return undefined;
    const name = field.names[0]?.name ?? "";
    return new VarObject(name, this.resolveType(field.type, this.pkg.scope), false, this.pkg.scope, this.pkg);
  }

  private receiverNamedType(type: Type): NamedType | undefined {
    if (type instanceof NamedType) return type;
    if (type instanceof PointerType && type.base instanceof NamedType) return type.base;
    return undefined;
  }

  private cellType(expr: CellRefExpr): Type {
    const namespace = this.config.sheetNamespaces?.[expr.namespace.name];
    return namespace?.cells?.[expr.address.raw] ?? namespace?.defaultType ?? this.universe.basic.any;
  }

  private rangeElementType(expr: RangeRefExpr): Type {
    const namespace = this.config.sheetNamespaces?.[expr.namespace.name];
    return namespace?.cells?.[expr.start.raw] ?? namespace?.defaultType ?? this.universe.basic.any;
  }

  private checkAssignable(expr: Expr, target: Type, scope: Scope): void {
    const actual = this.checkExpr(expr, scope).type;
    if (!assignableTo(actual, target)) {
      this.error(`cannot use ${actual.typeString()} as ${target.typeString()}`, expr.span);
    }
  }

  private insert(scope: Scope, object: TypeObject, identNode?: Ident | BasicLit): void {
    const existing = scope.insert(object);
    if (identNode && identNode.kind === "Ident") this.info.defs.set(identNode, existing ? undefined : object);
    if (existing && identNode) this.error(`${object.name} already declared`, identNode.span);
  }

  private error(message: string, span?: SourceSpan, code = "GOJR_TYPE001"): void {
    this.diagnostics.push({
      filename: diagnosticFilename(span, REPL_FILENAME),
      code,
      severity: "error",
      message,
      ...(span ? { span } : {})
    });
  }
}

function unquote(value: string): string {
  if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
    return value.slice(1, -1).replace(/\r/g, "");
  }
  if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
    return decodeGoEscaped(value.slice(1, -1));
  }
  return value;
}

function parseGoIntLiteral(value: string): bigint {
  const text = value.replace(/_/g, "");
  if (/^0[0-7]+$/.test(text)) return BigInt(`0o${text.slice(1)}`);
  return BigInt(text);
}

function parseGoFloatLiteral(value: string): number {
  const text = value.replace(/_/g, "");
  const hex = /^0[xX]([0-9a-fA-F]*)(?:\.([0-9a-fA-F]*))?[pP]([+-]?[0-9]+)$/.exec(text);
  if (!hex) return Number(text);
  const whole = hex[1] || "0";
  const frac = hex[2] || "";
  const exponent = Number(hex[3]);
  const wholeValue = Number.parseInt(whole, 16);
  let fracValue = 0;
  for (let index = 0; index < frac.length; index += 1) {
    fracValue += Number.parseInt(frac[index] ?? "0", 16) / 16 ** (index + 1);
  }
  return (wholeValue + fracValue) * 2 ** exponent;
}

function parseGoRuneLiteral(value: string): bigint {
  const decoded = decodeGoEscaped(value.slice(1, -1));
  return BigInt([...decoded][0]?.codePointAt(0) ?? 0);
}

function decodeGoEscaped(value: string): string {
  let decoded = "";
  for (let index = 0; index < value.length; index += 1) {
    const char = value[index] ?? "";
    if (char !== "\\") {
      decoded += char;
      continue;
    }
    const next = value[index + 1] ?? "";
    index += 1;
    switch (next) {
      case "a": decoded += "\x07"; break;
      case "b": decoded += "\b"; break;
      case "f": decoded += "\f"; break;
      case "n": decoded += "\n"; break;
      case "r": decoded += "\r"; break;
      case "t": decoded += "\t"; break;
      case "v": decoded += "\x0b"; break;
      case "\\":
      case "\"":
      case "'":
        decoded += next;
        break;
      case "x":
        decoded += codePointFromEscape(value.slice(index + 1, index + 3), 16);
        index += 2;
        break;
      case "u":
        decoded += codePointFromEscape(value.slice(index + 1, index + 5), 16);
        index += 4;
        break;
      case "U":
        decoded += codePointFromEscape(value.slice(index + 1, index + 9), 16);
        index += 8;
        break;
      default:
        if (/^[0-7]$/.test(next)) {
          const digits = next + value.slice(index + 1, index + 3);
          decoded += codePointFromEscape(digits, 8);
          index += 2;
        } else {
          decoded += next;
        }
        break;
    }
  }
  return decoded;
}

function codePointFromEscape(digits: string, radix: number): string {
  const value = Number.parseInt(digits, radix);
  if (!Number.isFinite(value)) return "";
  try {
    return String.fromCodePoint(value);
  } catch {
    return "";
  }
}

function importName(path: string): string {
  const parts = path.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

function firstName(list: FieldList): Ident | undefined {
  return list.fields[0]?.names[0];
}

function sameSignature(left: SignatureType, right: SignatureType): boolean {
  return sameVarTuple(left.params, right.params) &&
    sameVarTuple(left.results, right.results) &&
    left.variadic === right.variadic &&
    typeString(left.receiver?.type) === typeString(right.receiver?.type);
}

function sameVarTuple(left: { variables: VarObject[] }, right: { variables: VarObject[] }): boolean {
  return left.variables.length === right.variables.length &&
    left.variables.every((variable, index) => typeString(variable.type) === typeString(right.variables[index]?.type));
}

function typeString(type: Type | undefined): string {
  return type?.typeString() ?? "";
}

function sameTypeParam(left: Type, right: Type): TypeParamType | undefined {
  return left instanceof TypeParamType && right instanceof TypeParamType && left.name === right.name ? left : undefined;
}

function operatorText(kind: TokenKind): string {
  switch (kind) {
    case TokenKind.Equal: return "==";
    case TokenKind.NotEqual: return "!=";
    case TokenKind.Less: return "<";
    case TokenKind.LessEqual: return "<=";
    case TokenKind.Greater: return ">";
    case TokenKind.GreaterEqual: return ">=";
    case TokenKind.Plus: return "+";
    case TokenKind.Minus: return "-";
    case TokenKind.Star: return "*";
    case TokenKind.Slash: return "/";
    case TokenKind.Percent: return "%";
    case TokenKind.Or: return "|";
    case TokenKind.Caret: return "^";
    case TokenKind.Amp: return "&";
    case TokenKind.BitClear: return "&^";
    case TokenKind.Shl: return "<<";
    case TokenKind.Shr: return ">>";
    default: return kind;
  }
}

function intersectTypeSets(typeSets: TypeSetInfo[]): TypeSetInfo {
  if (typeSets.length === 0) return emptyTypeSet();
  return {
    addable: typeSets.every((set) => set.addable),
    numeric: typeSets.every((set) => set.numeric),
    integer: typeSets.every((set) => set.integer),
    ordered: typeSets.every((set) => set.ordered),
    comparable: typeSets.every((set) => set.comparable)
  };
}

function basicTypeSetInfo(type: BasicType): TypeSetInfo {
  const actual = defaultBasicKind(type.basicKind);
  const integer = isIntegerBasicKind(actual);
  const float = actual === BasicKind.Float32 || actual === BasicKind.Float64;
  const complex = actual === BasicKind.Complex64 || actual === BasicKind.Complex128;
  const numeric = integer || float || complex;
  const string = actual === BasicKind.String;
  const bool = actual === BasicKind.Bool;
  return {
    addable: numeric || string,
    numeric,
    integer,
    ordered: integer || float || string,
    comparable: numeric || string || bool
  };
}

function defaultBasicKind(kind: BasicKind): BasicKind {
  switch (kind) {
    case BasicKind.UntypedInt:
      return BasicKind.Int64;
    case BasicKind.UntypedFloat:
      return BasicKind.Float64;
    case BasicKind.UntypedComplex:
      return BasicKind.Complex128;
    case BasicKind.UntypedString:
      return BasicKind.String;
    case BasicKind.UntypedBool:
      return BasicKind.Bool;
    default:
      return kind;
  }
}

function isIntegerBasicKind(kind: BasicKind): boolean {
  return [
    BasicKind.Int,
    BasicKind.Int8,
    BasicKind.Int16,
    BasicKind.Int32,
    BasicKind.Int64,
    BasicKind.Uint,
    BasicKind.Uint8,
    BasicKind.Uint16,
    BasicKind.Uint32,
    BasicKind.Uint64,
    BasicKind.Uintptr,
    BasicKind.Byte,
    BasicKind.Rune
  ].includes(kind);
}

function embeddedFieldName(expr: Expr): string {
  if (expr.kind === "Ident") return expr.name;
  if (expr.kind === "SelectorExpr") return expr.selector.name;
  if (expr.kind === "StarExpr") return embeddedFieldName(expr.expr);
  return "";
}

function lookupFieldOrMethod(type: Type, name: string): TypeObject | undefined {
  const actual = type instanceof PointerType ? type.base.underlying() : type.underlying();
  if (actual instanceof StructType) {
    const field = actual.fields.find((item) => item.name === name);
    if (field) return new VarObject(field.name, field.type, field.embedded);
    const promoted = promotedFieldOrMethod(actual, name);
    if (promoted) return promoted;
  }
  const directMethod = methodSet(type).find((method) => method.name === name);
  if (directMethod) return directMethod;
  if (type instanceof NamedType) return type.methodSet().find((method) => method.name === name);
  return undefined;
}

function promotedFieldOrMethod(type: StructType, name: string, seen = new Set<StructType>()): TypeObject | undefined {
  if (seen.has(type)) return undefined;
  seen.add(type);
  const matches: TypeObject[] = [];
  for (const field of type.fields.filter((item) => item.embedded)) {
    const fieldType = field.type instanceof PointerType ? field.type.base : field.type;
    const directMethod = methodSet(field.type).find((method) => method.name === name);
    if (directMethod) {
      matches.push(directMethod);
      continue;
    }
    const underlying = fieldType.underlying();
    if (!(underlying instanceof StructType)) continue;
    const directField = underlying.fields.find((item) => item.name === name);
    if (directField) {
      matches.push(new VarObject(directField.name, directField.type, directField.embedded));
      continue;
    }
    const promoted = promotedFieldOrMethod(underlying, name, seen);
    if (promoted) matches.push(promoted);
  }
  return matches.length === 1 ? matches[0] : undefined;
}

function isStringLike(type: Type): boolean {
  return type instanceof BasicType && (type.basicKind === BasicKind.String || type.basicKind === BasicKind.UntypedString);
}

function isNumericLike(type: Type): boolean {
  return type instanceof BasicType && type.info.has("numeric");
}

function isUntyped(type: Type): boolean {
  return type instanceof BasicType && type.info.has("untyped");
}

function defaultType(type: Type | undefined, universe: Universe): Type | undefined {
  if (!(type instanceof BasicType)) return type;
  switch (type.basicKind) {
    case BasicKind.UntypedBool:
      return universe.basic.bool;
    case BasicKind.UntypedInt:
      return universe.basic.int64;
    case BasicKind.UntypedFloat:
      return universe.basic.float64;
    case BasicKind.UntypedComplex:
      return universe.basic.complex128;
    case BasicKind.UntypedString:
      return universe.basic.string;
    default:
      return type;
  }
}

function fmtPackageInfo(universe: Universe): PackageInfo {
  const pkg = new PackageInfo("fmt", "fmt", universe.scope);
  const scope = pkg.scope;
  scope.insert(new FuncObject(
    "Printf",
    new SignatureType(
      undefined,
      tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))),
      tuple(varOf("", universe.basic.int64)),
      true
    ),
    scope,
    pkg
  ));
  scope.insert(new FuncObject(
    "Sprintf",
    new SignatureType(
      undefined,
      tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))),
      tuple(varOf("", universe.basic.string)),
      true
    ),
    scope,
    pkg
  ));
  scope.insert(new FuncObject(
    "Println",
    new SignatureType(
      undefined,
      tuple(varOf("args", new SliceType(universe.basic.any))),
      tuple(varOf("", universe.basic.int64)),
      true
    ),
    scope,
    pkg
  ));
  return pkg;
}

function testingPackageInfo(universe: Universe): PackageInfo {
  const pkg = new PackageInfo("testing", "testing", universe.scope);
  const scope = pkg.scope;
  const typeName = new TypeNameObject("T", universe.basic.invalid, scope, pkg);
  const tType = new NamedType(typeName, new StructType([]));
  typeName.setType(tType);
  scope.insert(typeName);

  const receiver = varOf("", new PointerType(tType));
  const noResults = tuple();
  const addMethod = (name: string, params = tuple(), results = noResults, variadic = false): void => {
    tType.addMethod(new FuncObject(name, new SignatureType(receiver, params, results, variadic), scope, pkg));
  };
  const anyArgs = tuple(varOf("args", new SliceType(universe.basic.any)));
  const formattedArgs = tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any)));
  addMethod("Fail");
  addMethod("FailNow");
  addMethod("Failed", tuple(), tuple(varOf("", universe.basic.bool)));
  addMethod("Fatal", anyArgs, noResults, true);
  addMethod("Fatalf", formattedArgs, noResults, true);
  addMethod("Error", anyArgs, noResults, true);
  addMethod("Errorf", formattedArgs, noResults, true);
  addMethod("Log", anyArgs, noResults, true);
  addMethod("Logf", formattedArgs, noResults, true);
  addMethod("Name", tuple(), tuple(varOf("", universe.basic.string)));
  addMethod("Helper");
  addMethod("Skip", anyArgs, noResults, true);
  addMethod("Skipf", formattedArgs, noResults, true);
  addMethod("SkipNow");
  addMethod("Skipped", tuple(), tuple(varOf("", universe.basic.bool)));

  scope.insert(new FuncObject("Short", new SignatureType(undefined, tuple(), tuple(varOf("", universe.basic.bool))), scope, pkg));
  scope.insert(new FuncObject("Verbose", new SignatureType(undefined, tuple(), tuple(varOf("", universe.basic.bool))), scope, pkg));
  return pkg;
}

function standardPackageInfo(path: string, universe: Universe): PackageInfo | undefined {
  if (path === "fmt") return fmtPackageInfo(universe);
  if (path === "testing") return testingPackageInfo(universe);
  return undefined;
}

function promoteNumeric(left: Type, right: Type, universe: Universe): Type {
  if (left instanceof BasicType && (left.basicKind === BasicKind.Complex64 || left.basicKind === BasicKind.Complex128)) return left;
  if (right instanceof BasicType && (right.basicKind === BasicKind.Complex64 || right.basicKind === BasicKind.Complex128)) return right;
  if (left instanceof BasicType && left.basicKind === BasicKind.UntypedComplex) return left;
  if (right instanceof BasicType && right.basicKind === BasicKind.UntypedComplex) return right;
  if (left instanceof BasicType && left.basicKind === BasicKind.Float64) return left;
  if (right instanceof BasicType && right.basicKind === BasicKind.Float64) return right;
  if (left instanceof BasicType && left.basicKind === BasicKind.UntypedFloat) return left;
  if (right instanceof BasicType && right.basicKind === BasicKind.UntypedFloat) return right;
  if (left instanceof BasicType && left.basicKind === BasicKind.Int64) return left;
  if (right instanceof BasicType && right.basicKind === BasicKind.Int64) return right;
  if (left instanceof BasicType && left.basicKind === BasicKind.UntypedInt && right instanceof BasicType && right.basicKind === BasicKind.UntypedInt) {
    return universe.basic.untypedInt;
  }
  return left instanceof BasicType ? left : universe.basic.invalid;
}
