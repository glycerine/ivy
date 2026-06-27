import type {
  ArrayLiteralExpression,
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  ConstDeclStatement,
  DeclarationSpec,
  DeferStatement,
  Expression,
  ExpressionStatement,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  IdentifierExpression,
  IfStatement,
  IndexExpression,
  IncDecStatement,
  LabeledStatement,
  LiteralExpression,
  MapLiteralExpression,
  ProgramAst,
  ReturnStatement,
  SelectorExpression,
  ShortVarStatement,
  SelectStatement,
  SliceExpression,
  StructLiteralExpression,
  SwitchStatement,
  Statement,
  UnaryExpression,
  TypeSpec,
  VarDeclStatement
} from "../ast.js";
import type { GoJuniorPackageExportData } from "../build.js";
import type { Diagnostic } from "../diagnostics.js";
import { EmitterContext } from "./context.js";
import { checkedInWasmStencil } from "./stencils.js";
import { GENERATED_ARTIFACT_INTRINSIC_IMPORTS } from "../intrinsicPackages.js";

export const GOJR_STAGE1_BACKEND = "copy-patch-wasm-stage1";

export interface Stage1PackageEmitResult {
  diagnostics: Diagnostic[];
  javascript: string;
  wasmBase64?: string;
}

interface ExpressionEmitEnv {
  locals: Map<string, string>;
  localTypes: Map<string, string>;
  localTypeUnderlyings: Map<string, string>;
  facts: PackageEmitFacts;
  namedResults?: string[];
  expectedReturnTypes?: string[];
  typeParameters?: Set<string>;
  deferName?: string;
  gotoLabels?: Map<string, number>;
  gotoPcName?: string;
  gotoLoopName?: string;
  gotoBreakLabels?: Map<string, string>;
  loopContinueLabel?: string;
  iotaValue?: number;
  numericLiteralKind?: "float";
}

interface PackageEmitFacts {
  methodKeys: Map<string, string>;
  resultCounts: Map<string, number>;
  packageTypes: Map<string, string>;
  typeUnderlyings: Map<string, string>;
  constValues: Map<string, bigint>;
  structFields: Map<string, Array<{ name: string; type: string; embedded: boolean }>>;
  imports: Map<string, string>;
  interfaceTypes: Set<string>;
  functionParamTypes: Map<string, string[]>;
  functionResultTypes: Map<string, string[]>;
  functionTypeParameters: Map<string, string[]>;
  methodPointerReceivers: Map<string, boolean>;
}

interface WasmScalarLowering {
  exportName: string;
  resultType: "int64" | "bool";
}

interface GeneratedTypeDescriptor {
  type: string;
  name: string;
  string: string;
  kind: string;
  pkgPath: string;
  pkgName?: string;
  underlying?: string;
  elem?: string;
  key?: string;
  len?: number;
  params?: string[];
  results?: string[];
  fields?: Array<{
    name: string;
    type: string;
    embedded: boolean;
    pkgPath?: string;
    tag?: string;
  }>;
  typeParameters?: string[];
  methods?: Array<{
    name: string;
    receiver: string;
    pointerReceiver: boolean;
    params: string[];
    results: string[];
  }>;
  interfaceMethods?: Array<{
    name: string;
    params: string[];
    results: string[];
  }>;
  interfaceEmbeds?: string[];
}

const emptyPackageEmitFacts: PackageEmitFacts = {
  methodKeys: new Map(),
  resultCounts: new Map(),
  packageTypes: new Map(),
  typeUnderlyings: new Map(),
  constValues: new Map(),
  structFields: new Map(),
  imports: new Map(),
  interfaceTypes: new Set(),
  functionParamTypes: new Map(),
  functionResultTypes: new Map(),
  functionTypeParameters: new Map(),
  methodPointerReceivers: new Map()
};

const emptyExpressionEnv: ExpressionEmitEnv = {
  locals: new Map(),
  localTypes: new Map(),
  localTypeUnderlyings: new Map(),
  facts: emptyPackageEmitFacts
};
const JS_RESERVED_WORDS = new Set([
  "await",
  "break",
  "case",
  "catch",
  "class",
  "const",
  "continue",
  "debugger",
  "default",
  "delete",
  "do",
  "else",
  "enum",
  "export",
  "extends",
  "false",
  "finally",
  "for",
  "function",
  "if",
  "import",
  "in",
  "instanceof",
  "let",
  "new",
  "null",
  "return",
  "super",
  "switch",
  "this",
  "throw",
  "true",
  "try",
  "typeof",
  "undefined",
  "var",
  "void",
  "while",
  "with",
  "yield"
]);

export function emitStage1Package(artifact: GoJuniorPackageExportData, ast: ProgramAst): Stage1PackageEmitResult {
  const ctx = new EmitterContext({ artifact });
  const facts = packageEmitFacts(ast, artifact);
  const wasmLowerings = new Map<FunctionDecl, WasmScalarLowering>();
  for (const fn of ast.functions) {
    const lowering = wasmScalarLoweringForFunction(fn);
    if (lowering) wasmLowerings.set(fn, lowering);
  }
  const usesWasm = wasmLowerings.size > 0;
  const wasmBase64 = usesWasm ? checkedInWasmStencil("i64.scalar.add").wasmBase64 : undefined;
  const declarationItems: TopLevelDeclarationEmission[] = [];
  const functionLines: string[] = [];
  const initNames: string[] = [];
  const typeDescriptorLines = emitTypeDescriptorLines(ast, artifact, facts);
  const topLevelDeclarationNames = topLevelValueDeclarationNames(ast);
  const topLevelFunctionDependencies = topLevelFunctionDependencyMap(ast, topLevelDeclarationNames, facts);
  let declarationOrder = 0;

  for (const statement of ast.body) {
    if (statement.kind === "ConstDecl") {
      declarationItems.push(...emitConstDecl(ctx, statement, facts, topLevelDeclarationNames, topLevelFunctionDependencies, () => declarationOrder++));
      continue;
    }
    if (statement.kind === "VarDecl") {
      declarationItems.push(...emitVarDecl(ctx, statement, facts, topLevelDeclarationNames, topLevelFunctionDependencies, () => declarationOrder++));
      continue;
    }
    if (statement.kind === "TypeDecl") continue;
    ctx.emitError(`unsupported Stage 1 top-level statement ${statement.kind}`);
  }

  for (const fn of ast.functions) {
    if (!fn.receiver && fn.name === "init") {
      const initName = ctx.symbol("init");
      initNames.push(initName);
      functionLines.push(...emitFunction(ctx, fn, facts, { localName: initName }));
      continue;
    }
    const wasmLowering = wasmLowerings.get(fn);
    functionLines.push(...emitFunction(ctx, fn, facts, wasmLowering ? { wasmLowering } : {}));
  }

  const bodyLines = [
    ...functionLines,
    ...intrinsicPackageOverrideLines(artifact.importPath),
    ...orderTopLevelDeclarationEmissions(declarationItems).flatMap((item) => item.lines),
    ...initNames.map((name) => `  await ${name}();`)
  ];
  const javascript = stage1JavaScript(artifact, usesWasm, bodyLines, typeDescriptorLines, facts);
  return {
    diagnostics: ctx.diagnostics,
    javascript,
    ...(wasmBase64 ? { wasmBase64 } : {})
  };
}

function intrinsicPackageOverrideLines(importPath: string): string[] {
  const lines: string[] = [];
  if (importPath === "internal/abi") {
    lines.push(
      "  pkg[\"TypeFor\"] = async (__gojrTypeArgs) => __gojrInternalAbiTypePointer(__gojrTypeArg(__gojrTypeArgs, \"T\"));",
      "  pkg[\"TypeOf\"] = async (value) => __gojrInternalAbiTypeOf(value);",
      "  pkg[\"Type.MapType\"] = async (typePointer) => __gojrInternalAbiMapType(typePointer);",
      "  pkg[\"NoEscape\"] = async (value) => value;",
      "  pkg[\"Escape\"] = async (value) => value;",
      "  pkg[\"EscapeNonString\"] = async (value) => value;",
      "  pkg[\"EscapeToResultNonString\"] = async (value) => value;"
    );
  }
  if (importPath === "crypto/internal/constanttime") {
    lines.push("  pkg[\"boolToUint8\"] = async (value) => __gojrBool(value) ? 1n : 0n;");
  }
  return lines;
}

interface TopLevelDeclarationEmission {
  names: string[];
  order: number;
  lines: string[];
  dependencies: Set<string>;
}

function topLevelValueDeclarationNames(ast: ProgramAst): Set<string> {
  const names = new Set<string>();
  for (const statement of ast.body) {
    if (statement.kind !== "ConstDecl" && statement.kind !== "VarDecl") continue;
    for (const declaration of statement.declarations) {
      if (declaration.name !== "_") names.add(declaration.name);
    }
  }
  return names;
}

function emitConstDecl(
  ctx: EmitterContext,
  statement: ConstDeclStatement,
  facts: PackageEmitFacts,
  topLevelNames: Set<string>,
  functionDependencies: Map<string, Set<string>>,
  nextOrder: () => number
): TopLevelDeclarationEmission[] {
  const items: TopLevelDeclarationEmission[] = [];
  let inheritedValues: Array<Expression | undefined> = [];
  for (const group of declarationGroups(statement.declarations)) {
    const groupValues = declarationValueExpressions(group);
    if (groupValues.some(Boolean)) inheritedValues = groupValues;
    for (const declaration of group) {
      const effectiveValue = declaration.value ?? inheritedValues[declaration.valueIndex ?? 0];
      items.push(emitDeclaration(ctx, declaration, "const", facts, topLevelNames, functionDependencies, nextOrder(), effectiveValue));
    }
  }
  return items;
}

function emitVarDecl(
  ctx: EmitterContext,
  statement: VarDeclStatement,
  facts: PackageEmitFacts,
  topLevelNames: Set<string>,
  functionDependencies: Map<string, Set<string>>,
  nextOrder: () => number
): TopLevelDeclarationEmission[] {
  const items: TopLevelDeclarationEmission[] = [];
  for (const group of declarationGroups(statement.declarations)) {
    const shared = group.length > 1 && group.every((declaration) =>
      declaration.valueCount === 1 &&
      declaration.groupNameCount === group.length &&
      declaration.value === group[0]?.value
    );
    if (shared && group[0]?.value) {
      items.push(emitTupleVarDeclarationGroup(ctx, group, facts, topLevelNames, functionDependencies, nextOrder(), group[0].value));
      continue;
    }
    for (const declaration of group) {
      items.push(emitDeclaration(ctx, declaration, "var", facts, topLevelNames, functionDependencies, nextOrder()));
    }
  }
  return items;
}

function emitTupleVarDeclarationGroup(
  ctx: EmitterContext,
  group: DeclarationSpec[],
  facts: PackageEmitFacts,
  topLevelNames: Set<string>,
  functionDependencies: Map<string, Set<string>>,
  order: number,
  source: Expression
): TopLevelDeclarationEmission {
  const env: ExpressionEmitEnv = { ...emptyExpressionEnv, facts };
  const rendered = tupleSourceExpressionToJs(ctx, source, group.length, env);
  const names = group.map((declaration, index) => declaration.name === "_" ? `__gojrBlank${order}_${index}` : declaration.name);
  if (!rendered) {
    ctx.emitError(`unsupported Stage 1 declaration for ${group.map((declaration) => declaration.name).join(", ")}`);
    return { names, order, lines: [], dependencies: new Set() };
  }
  const temp = ctx.symbol("varTuple");
  const sourceTypes = tupleSourceTypeTexts(source, group.length, env);
  const lines = [`  const ${temp} = __gojrTupleValues(${rendered});`];
  for (const [index, declaration] of group.entries()) {
    const sourceType = sourceTypes[index];
    const targetType = declaration.type?.text ?? sourceType;
    const value = valueForTargetType(`${temp}[${index}]`, targetType, sourceType, env);
    if (declaration.name === "_") {
      lines.push(`  void (${value});`);
      continue;
    }
    lines.push(`  pkg[${JSON.stringify(declaration.name)}] = ${value};`);
  }
  const dependencies = expressionDependencies(source, topLevelNames, facts, "", functionDependencies);
  for (const declaration of group) dependencies.delete(declaration.name);
  return { names, order, lines, dependencies };
}

function emitDeclaration(
  ctx: EmitterContext,
  declaration: DeclarationSpec,
  kind: "const" | "var",
  facts: PackageEmitFacts,
  topLevelNames: Set<string>,
  functionDependencies: Map<string, Set<string>>,
  order: number,
  effectiveValue = declaration.value
): TopLevelDeclarationEmission {
  const env: ExpressionEmitEnv = {
    ...emptyExpressionEnv,
    facts,
    ...(declaration.iotaIndex !== undefined ? { iotaValue: declaration.iotaIndex } : {})
  };
  const itemName = declaration.name === "_" ? `__gojrBlank${order}` : declaration.name;
  if (declaration.name === "_" && kind === "const") {
    return { names: [itemName], order, lines: [], dependencies: new Set() };
  }
  if (declaration.name === "_" && kind === "var" && !effectiveValue) {
    return { names: [itemName], order, lines: [], dependencies: new Set() };
  }
  const rawValue = effectiveValue
    ? expressionToJs(ctx, effectiveValue, env)
    : zeroValueForType(declaration.type?.text, facts);
  if (!rawValue) {
    ctx.emitError(`unsupported Stage 1 declaration for ${declaration.name}`);
    return { names: [itemName], order, lines: [], dependencies: new Set() };
  }
  const targetType = declarationTypeText(declaration, effectiveValue);
  const value = valueForTargetType(rawValue, targetType, expressionTypeText(effectiveValue, env), env);
  if (declaration.name === "_") {
    return {
      names: [itemName],
      order,
      lines: [`  void (${value});`],
      dependencies: expressionDependencies(effectiveValue, topLevelNames, facts, itemName, functionDependencies)
    };
  }
  return {
    names: [declaration.name],
    order,
    lines: [`  pkg[${JSON.stringify(declaration.name)}] = ${value};`],
    dependencies: expressionDependencies(effectiveValue, topLevelNames, facts, declaration.name, functionDependencies)
  };
}

function orderTopLevelDeclarationEmissions(items: TopLevelDeclarationEmission[]): TopLevelDeclarationEmission[] {
  const remaining = [...items].sort((left, right) => left.order - right.order);
  const ordered: TopLevelDeclarationEmission[] = [];
  while (remaining.length > 0) {
    const remainingNames = new Set(remaining.flatMap((item) => item.names));
    const readyIndex = remaining.findIndex((item) =>
      [...item.dependencies].every((dependency) => !remainingNames.has(dependency))
    );
    if (readyIndex < 0) {
      ordered.push(...remaining);
      break;
    }
    const [ready] = remaining.splice(readyIndex, 1);
    if (ready) ordered.push(ready);
  }
  return ordered;
}

function expressionDependencies(
  expression: Expression | undefined,
  topLevelNames: Set<string>,
  facts: PackageEmitFacts,
  selfName: string,
  functionDependencies: Map<string, Set<string>> = new Map()
): Set<string> {
  const dependencies = new Set<string>();
  const namesAndFunctions = functionDependencies.size > 0
    ? new Set([...topLevelNames, ...functionDependencies.keys()])
    : topLevelNames;
  collectExpressionDependencies(expression, namesAndFunctions, facts, dependencies);
  expandFunctionDependencies(dependencies, functionDependencies);
  dependencies.delete(selfName);
  return dependencies;
}

function topLevelFunctionDependencyMap(ast: ProgramAst, topLevelNames: Set<string>, facts: PackageEmitFacts): Map<string, Set<string>> {
  const functionNames = new Set<string>();
  const functions = new Map<string, FunctionDecl>();
  for (const fn of ast.functions) {
    if (fn.name === "init") continue;
    const key = functionPackageKey(fn);
    functionNames.add(key);
    functions.set(key, fn);
  }
  const namesAndFunctions = new Set([...topLevelNames, ...functionNames]);
  const direct = new Map<string, Set<string>>();
  for (const [name, fn] of functions.entries()) {
    const dependencies = new Set<string>();
    for (const statement of fn.body.statements) collectStatementDependencies(statement, namesAndFunctions, facts, dependencies);
    dependencies.delete(name);
    direct.set(name, dependencies);
  }
  const expanded = new Map<string, Set<string>>();
  const expand = (name: string, visiting = new Set<string>()): Set<string> => {
    const cached = expanded.get(name);
    if (cached) return cached;
    if (visiting.has(name)) return new Set();
    visiting.add(name);
    const out = new Set<string>();
    for (const dependency of direct.get(name) ?? []) {
      if (functionNames.has(dependency)) {
        for (const transitive of expand(dependency, visiting)) out.add(transitive);
      } else if (topLevelNames.has(dependency)) {
        out.add(dependency);
      }
    }
    visiting.delete(name);
    expanded.set(name, out);
    return out;
  };
  for (const name of functionNames) expand(name);
  return expanded;
}

function expandFunctionDependencies(dependencies: Set<string>, functionDependencies: Map<string, Set<string>>): void {
  const functionNames = [...dependencies].filter((dependency) => functionDependencies.has(dependency));
  for (const name of functionNames) {
    dependencies.delete(name);
    for (const dependency of functionDependencies.get(name) ?? []) dependencies.add(dependency);
  }
}

function collectExpressionDependencies(
  expression: Expression | undefined,
  topLevelNames: Set<string>,
  facts: PackageEmitFacts,
  dependencies: Set<string>
): void {
  if (!expression) return;
  switch (expression.kind) {
    case "Identifier":
      if (topLevelNames.has(expression.name)) dependencies.add(expression.name);
      return;
    case "Literal":
    case "TypeExpression":
      return;
    case "SelectorExpression":
      if (expression.object.kind === "Identifier" && facts.imports.has(expression.object.name)) return;
      {
        const methodKey = methodPackageKeyForSelector(expression, { ...emptyExpressionEnv, facts });
        if (methodKey && topLevelNames.has(methodKey)) dependencies.add(methodKey);
      }
      collectExpressionDependencies(expression.object, topLevelNames, facts, dependencies);
      return;
    case "UnaryExpression":
      collectExpressionDependencies(expression.operand, topLevelNames, facts, dependencies);
      return;
    case "BinaryExpression":
      collectExpressionDependencies(expression.left, topLevelNames, facts, dependencies);
      collectExpressionDependencies(expression.right, topLevelNames, facts, dependencies);
      return;
    case "CallExpression":
      collectExpressionDependencies(expression.callee, topLevelNames, facts, dependencies);
      for (const arg of expression.args) collectExpressionDependencies(arg, topLevelNames, facts, dependencies);
      return;
    case "IndexExpression":
      collectExpressionDependencies(expression.object, topLevelNames, facts, dependencies);
      collectExpressionDependencies(expression.index, topLevelNames, facts, dependencies);
      return;
    case "SliceExpression":
      collectExpressionDependencies(expression.object, topLevelNames, facts, dependencies);
      collectExpressionDependencies(expression.start, topLevelNames, facts, dependencies);
      collectExpressionDependencies(expression.end, topLevelNames, facts, dependencies);
      collectExpressionDependencies(expression.max, topLevelNames, facts, dependencies);
      return;
    case "TypeAssertionExpression":
      collectExpressionDependencies(expression.expression, topLevelNames, facts, dependencies);
      return;
    case "ArrayLiteralExpression":
      for (const element of expression.elements) {
        collectExpressionDependencies(element.key, topLevelNames, facts, dependencies);
        collectExpressionDependencies(element.value, topLevelNames, facts, dependencies);
      }
      return;
    case "StructLiteralExpression":
      for (const field of expression.fields) {
        collectExpressionDependencies(field.key, topLevelNames, facts, dependencies);
        collectExpressionDependencies(field.value, topLevelNames, facts, dependencies);
      }
      return;
    case "MapLiteralExpression":
      for (const entry of expression.entries) {
        collectExpressionDependencies(entry.key, topLevelNames, facts, dependencies);
        collectExpressionDependencies(entry.value, topLevelNames, facts, dependencies);
      }
      return;
    case "FunctionLiteralExpression":
      for (const statement of expression.body.statements) collectStatementDependencies(statement, topLevelNames, facts, dependencies);
      return;
    default:
      return;
  }
}

function collectStatementDependencies(
  statement: Statement | undefined,
  topLevelNames: Set<string>,
  facts: PackageEmitFacts,
  dependencies: Set<string>
): void {
  if (!statement) return;
  switch (statement.kind) {
    case "ExpressionStatement":
      collectExpressionDependencies(statement.expression, topLevelNames, facts, dependencies);
      return;
    case "LabeledStatement":
      collectStatementDependencies(statement.statement, topLevelNames, facts, dependencies);
      return;
    case "AssignStatement":
      for (const target of statement.targets) collectExpressionDependencies(target, topLevelNames, facts, dependencies);
      for (const value of statement.values) collectExpressionDependencies(value, topLevelNames, facts, dependencies);
      return;
    case "ShortVarStatement":
      for (const value of statement.values) collectExpressionDependencies(value, topLevelNames, facts, dependencies);
      return;
    case "ReturnStatement":
      for (const value of statement.values) collectExpressionDependencies(value, topLevelNames, facts, dependencies);
      return;
    case "BlockStatement":
      for (const child of statement.statements) collectStatementDependencies(child, topLevelNames, facts, dependencies);
      return;
    case "ConstDecl":
    case "VarDecl":
      for (const declaration of statement.declarations) collectExpressionDependencies(declaration.value, topLevelNames, facts, dependencies);
      return;
    case "IfStatement":
      collectStatementDependencies(statement.init, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.condition, topLevelNames, facts, dependencies);
      collectStatementDependencies(statement.thenBlock, topLevelNames, facts, dependencies);
      collectStatementDependencies(statement.elseBranch, topLevelNames, facts, dependencies);
      return;
    case "SwitchStatement":
      collectStatementDependencies(statement.init, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.expression, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.typeSwitch?.expression, topLevelNames, facts, dependencies);
      for (const clause of statement.clauses) {
        for (const value of clause.values) collectExpressionDependencies(value, topLevelNames, facts, dependencies);
        for (const child of clause.statements) collectStatementDependencies(child, topLevelNames, facts, dependencies);
      }
      return;
    case "SelectStatement":
      for (const clause of statement.clauses) {
        collectStatementDependencies(clause.comm, topLevelNames, facts, dependencies);
        for (const child of clause.statements) collectStatementDependencies(child, topLevelNames, facts, dependencies);
      }
      return;
    case "ForStatement":
      collectStatementDependencies(statement.init, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.condition, topLevelNames, facts, dependencies);
      collectStatementDependencies(statement.post, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.range?.keyTarget, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.range?.valueTarget, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.range?.source, topLevelNames, facts, dependencies);
      collectStatementDependencies(statement.body, topLevelNames, facts, dependencies);
      return;
    case "DeferStatement":
      collectExpressionDependencies(statement.expression, topLevelNames, facts, dependencies);
      return;
    case "GoStatement":
      collectExpressionDependencies(statement.call, topLevelNames, facts, dependencies);
      return;
    case "SendStatement":
      collectExpressionDependencies(statement.channel, topLevelNames, facts, dependencies);
      collectExpressionDependencies(statement.value, topLevelNames, facts, dependencies);
      return;
    case "IncDecStatement":
      collectExpressionDependencies(statement.target, topLevelNames, facts, dependencies);
      return;
    default:
      return;
  }
}

function declarationGroups(declarations: DeclarationSpec[]): DeclarationSpec[][] {
  const groups: DeclarationSpec[][] = [];
  let currentGroup: DeclarationSpec[] = [];
  let currentKey: number | undefined;
  for (const declaration of declarations) {
    const key = declaration.valueGroup;
    if (currentGroup.length > 0 && key !== currentKey) {
      groups.push(currentGroup);
      currentGroup = [];
    }
    currentKey = key;
    currentGroup.push(declaration);
  }
  if (currentGroup.length > 0) groups.push(currentGroup);
  return groups;
}

function declarationValueExpressions(group: DeclarationSpec[]): Array<Expression | undefined> {
  const out: Array<Expression | undefined> = [];
  for (const declaration of group) {
    if (declaration.value) out[declaration.valueIndex ?? 0] = declaration.value;
  }
  return out;
}

interface FunctionEmitOptions {
  localName?: string;
  wasmLowering?: WasmScalarLowering;
}

function emitFunction(ctx: EmitterContext, fn: FunctionDecl, facts: PackageEmitFacts, options: FunctionEmitOptions = {}): string[] {
  if (options.wasmLowering) {
    const parameters = functionParameterBindings(fn, facts);
    const wasmCall = `__gojrWasmExports.${options.wasmLowering.exportName}(BigInt(${parameters.params[0]}), BigInt(${parameters.params[1]}))`;
    const result = options.wasmLowering.resultType === "bool" ? `Boolean(${wasmCall})` : wasmCall;
    if (options.localName) {
      return [
        `  const ${options.localName} = async (${parameters.params.join(", ")}) => ${result};`
      ];
    }
    return [
      `  pkg[${JSON.stringify(functionPackageKey(fn))}] = async (${parameters.params.join(", ")}) => ${result};`
    ];
  }

  const parameters = functionParameterBindings(fn, facts);
  const bodyEnv = cloneExpressionEnv(parameters.env);
  bodyEnv.expectedReturnTypes = fn.signature.results.map((result) => result.type.text);
  const hasDefer = statementsContainDefer(fn.body.statements);
  if (hasDefer) bodyEnv.deferName = ctx.symbol("defer");
  const resultLines = emitNamedResultDeclarations(fn, bodyEnv);
  const parameterCoercionLines = functionDeclarationParameterCoercionLines(fn, parameters.env, "    ");
  const statementLines = statementsRequireGotoStateMachine(fn.body.statements)
    ? emitGotoStateMachineStatements(ctx, fn.body.statements, bodyEnv, "    ", () => [`return ${defaultFunctionReturn(fn, bodyEnv)};`])
    : emitStatements(ctx, fn.body.statements, bodyEnv, "    ");
  if (!statementLines) return [];
  const hasExplicitReturn = functionBodyAlwaysReturns(fn.body);
  const target = options.localName ? `const ${options.localName}` : `pkg[${JSON.stringify(functionPackageKey(fn))}]`;
  if (hasDefer) {
    const defaultReturn = defaultFunctionReturn(fn, bodyEnv);
    return [
      `  ${target} = async (${parameters.params.join(", ")}) => __gojrCallScope(async () => {`,
      ...parameterCoercionLines,
      ...resultLines.map((line) => `    ${line}`),
      `    return await __gojrDeferScope(async (${bodyEnv.deferName}) => {`,
      ...statementLines,
      ...(hasExplicitReturn ? [] : [`      return ${defaultReturn};`]),
      `    }, () => ${defaultReturn});`,
      "  });"
    ];
  }
  return [
    `  ${target} = async (${parameters.params.join(", ")}) => __gojrCallScope(async () => {`,
    ...parameterCoercionLines,
    ...resultLines.map((line) => `    ${line}`),
    ...statementLines,
    ...(hasExplicitReturn ? [] : [`    return ${defaultFunctionReturn(fn, bodyEnv)};`]),
    "  });"
  ];
}

function returnExpression(ctx: EmitterContext, statement: ReturnStatement, env: ExpressionEmitEnv): string | undefined {
  if (statement.values.length === 0) return namedResultReturn(env) ?? "null";
  if (statement.values.length > 1) {
    const values = statement.values.map((value, index) => {
      const rendered = expressionToJs(ctx, value, env);
      return rendered ? valueForTargetType(rendered, env.expectedReturnTypes?.[index], expressionTypeText(value, env), env) : undefined;
    });
    if (values.some((value) => value === undefined)) return undefined;
    return `__gojrTuple([${values.filter((value): value is string => value !== undefined).join(", ")}])`;
  }
  const rendered = expressionToJs(ctx, statement.values[0], env);
  if (!rendered) return undefined;
  if ((env.expectedReturnTypes?.length ?? 0) > 1) {
    return returnTupleExpression(rendered, env.expectedReturnTypes ?? [], env);
  }
  return valueForTargetType(rendered, env.expectedReturnTypes?.[0], expressionTypeText(statement.values[0], env), env);
}

function returnTupleExpression(rendered: string, expectedTypes: string[], env: ExpressionEmitEnv): string {
  const valuesName = "__gojrReturnValues";
  const converted = expectedTypes.map((typeText, index) =>
    valueForTargetType(`${valuesName}[${index}]`, typeText, undefined, env)
  );
  return `__gojrTuple(((${valuesName}) => [${converted.join(", ")}])(__gojrTupleValues(${rendered})))`;
}

function emitNamedResultDeclarations(fn: FunctionDecl, env: ExpressionEmitEnv): string[] {
  const namedResults: string[] = [];
  const lines: string[] = [];
  for (const [index, result] of fn.signature.results.entries()) {
    if (!result.name || result.name === "_") continue;
    const name = safeLocalName(result.name, index);
    env.locals.set(result.name, name);
    env.localTypes.set(result.name, result.type.text);
    namedResults.push(name);
    lines.push(`let ${name} = ${zeroValueForTypeInEnv(result.type.text, env) ?? "null"};`);
  }
  if (namedResults.length > 0) env.namedResults = namedResults;
  return lines;
}

function namedResultReturn(env: ExpressionEmitEnv): string | undefined {
  if (!env.namedResults || env.namedResults.length === 0) return undefined;
  if (env.namedResults.length === 1) return env.namedResults[0];
  return `__gojrTuple([${env.namedResults.join(", ")}])`;
}

function defaultFunctionReturn(fn: FunctionDecl, env: ExpressionEmitEnv): string {
  const named = namedResultReturn(env);
  if (named) return named;
  if (fn.signature.results.length === 0) return "null";
  if (fn.signature.results.length === 1) return zeroValueForTypeInEnv(fn.signature.results[0]?.type.text, env) ?? "null";
  return `__gojrTuple([${fn.signature.results.map((result) => zeroValueForTypeInEnv(result.type.text, env) ?? "null").join(", ")}])`;
}

function functionBodyAlwaysReturns(body: BlockStatement): boolean {
  const last = body.statements[body.statements.length - 1];
  return last?.kind === "ReturnStatement";
}

function statementsContainDefer(statements: Statement[]): boolean {
  return statements.some(statementContainsDefer);
}

function statementContainsDefer(statement: Statement): boolean {
  switch (statement.kind) {
    case "DeferStatement":
      return true;
    case "BlockStatement":
      return statementsContainDefer(statement.statements);
    case "LabeledStatement":
      return statement.statement ? statementContainsDefer(statement.statement) : false;
    case "IfStatement":
      return statementContainsDeferInIf(statement);
    case "SwitchStatement":
      return statement.clauses.some((clause) => statementsContainDefer(clause.statements));
    case "SelectStatement":
      return statement.clauses.some((clause) => statementsContainDefer(clause.statements));
    case "ForStatement":
      return statementsContainDefer(statement.body.statements);
    default:
      return false;
  }
}

function statementContainsDeferInIf(statement: IfStatement): boolean {
  if (statementContainsDefer(statement.thenBlock)) return true;
  if (!statement.elseBranch) return false;
  return statement.elseBranch.kind === "IfStatement"
    ? statementContainsDeferInIf(statement.elseBranch)
    : statementContainsDefer(statement.elseBranch);
}

function statementsContainGoto(statements: Statement[]): boolean {
  return statements.some(statementContainsGoto);
}

function statementsRequireGotoStateMachine(statements: Statement[]): boolean {
  return statementsContainGoto(statements) && statements.some((statement) => statement.kind === "LabeledStatement");
}

function statementContainsGoto(statement: Statement): boolean {
  switch (statement.kind) {
    case "BranchStatement":
      return statement.branch === "goto";
    case "BlockStatement":
      return statementsContainGoto(statement.statements);
    case "LabeledStatement":
      return statement.statement ? statementContainsGoto(statement.statement) : false;
    case "IfStatement":
      return statementContainsGoto(statement.thenBlock) ||
        (statement.elseBranch ? statementContainsGoto(statement.elseBranch) : false) ||
        (statement.init ? statementContainsGoto(statement.init) : false);
    case "SwitchStatement":
      return statement.clauses.some((clause) => statementsContainGoto(clause.statements)) ||
        (statement.init ? statementContainsGoto(statement.init) : false);
    case "SelectStatement":
      return statement.clauses.some((clause) => statementsContainGoto(clause.statements) || (clause.comm ? statementContainsGoto(clause.comm) : false));
    case "ForStatement":
      return statementsContainGoto(statement.body.statements) ||
        (statement.init ? statementContainsGoto(statement.init) : false) ||
        (statement.post ? statementContainsGoto(statement.post) : false);
    default:
      return false;
  }
}

function statementContainsGotoTo(statement: Statement, label: string): boolean {
  switch (statement.kind) {
    case "BranchStatement":
      return statement.branch === "goto" && statement.label === label;
    case "BlockStatement":
      return statementsContainGotoTo(statement.statements, label);
    case "LabeledStatement":
      return statement.statement ? statementContainsGotoTo(statement.statement, label) : false;
    case "IfStatement":
      return statementContainsGotoTo(statement.thenBlock, label) ||
        (statement.elseBranch ? statementContainsGotoTo(statement.elseBranch, label) : false) ||
        (statement.init ? statementContainsGotoTo(statement.init, label) : false);
    case "SwitchStatement":
      return statement.clauses.some((clause) => statementsContainGotoTo(clause.statements, label)) ||
        (statement.init ? statementContainsGotoTo(statement.init, label) : false);
    case "SelectStatement":
      return statement.clauses.some((clause) => statementsContainGotoTo(clause.statements, label) || (clause.comm ? statementContainsGotoTo(clause.comm, label) : false));
    case "ForStatement":
      return statementsContainGotoTo(statement.body.statements, label) ||
        (statement.init ? statementContainsGotoTo(statement.init, label) : false) ||
        (statement.post ? statementContainsGotoTo(statement.post, label) : false);
    default:
      return false;
  }
}

function statementsContainGotoTo(statements: Statement[], label: string): boolean {
  return statements.some((statement) => statementContainsGotoTo(statement, label));
}

function emitGotoStateMachineStatements(
  ctx: EmitterContext,
  statements: Statement[],
  env: ExpressionEmitEnv,
  indent: string,
  defaultLines: (loopName: string) => string[]
): string[] | undefined {
  const labels = new Map<string, number>();
  for (const [index, statement] of statements.entries()) {
    if (statement.kind === "LabeledStatement") labels.set(statement.label, index);
  }
  const hoisted = hoistedGotoLocals(statements);
  for (const [name, typeText] of hoisted.entries()) {
    env.locals.set(name, safeLocalName(name, 0));
    if (typeText) env.localTypes.set(name, typeText);
  }
  const pc = ctx.symbol("pc");
  const loop = ctx.symbol("gotoLoop");
  env.gotoLabels = labels;
  env.gotoPcName = pc;
  env.gotoLoopName = loop;
  const lines: string[] = [];
  for (const name of hoisted.keys()) {
    lines.push(`${indent}let ${env.locals.get(name)};`);
  }
  lines.push(`${indent}let ${pc} = 0;`);
  lines.push(`${indent}${loop}: while (true) {`);
  lines.push(`${indent}  switch (${pc}) {`);
  for (const [index, statement] of statements.entries()) {
    const active = statement.kind === "LabeledStatement" && statement.statement ? statement.statement : statement;
    const emitted = emitStatement(ctx, active, env, `${indent}      `);
    if (!emitted) return undefined;
    lines.push(`${indent}    case ${index}:`);
    lines.push(...emitted);
    lines.push(`${indent}      ${pc} = ${index + 1};`);
    lines.push(`${indent}      continue ${loop};`);
  }
  lines.push(`${indent}    default:`);
  for (const line of defaultLines(loop)) {
    lines.push(`${indent}      ${line}`);
  }
  lines.push(`${indent}  }`);
  lines.push(`${indent}}`);
  return lines;
}

function hoistedGotoLocals(statements: Statement[]): Map<string, string | undefined> {
  const out = new Map<string, string | undefined>();
  for (const statement of statements) {
    if (statement.kind === "ShortVarStatement") {
      for (const [index, name] of statement.names.entries()) {
        if (name !== "_") out.set(name, expressionTypeText(statement.values[index], emptyExpressionEnv));
      }
    }
    if (statement.kind === "VarDecl") {
      for (const declaration of statement.declarations) {
        if (declaration.name !== "_") out.set(declaration.name, declarationTypeText(declaration));
      }
    }
    if (statement.kind === "LabeledStatement" && statement.statement?.kind === "ShortVarStatement") {
      for (const [index, name] of statement.statement.names.entries()) {
        if (name !== "_") out.set(name, expressionTypeText(statement.statement.values[index], emptyExpressionEnv));
      }
    }
  }
  return out;
}

function cloneExpressionEnv(env: ExpressionEmitEnv): ExpressionEmitEnv {
  return {
    locals: new Map(env.locals),
    localTypes: new Map(env.localTypes),
    localTypeUnderlyings: new Map(env.localTypeUnderlyings),
    facts: env.facts,
    ...(env.namedResults ? { namedResults: [...env.namedResults] } : {}),
    ...(env.expectedReturnTypes ? { expectedReturnTypes: [...env.expectedReturnTypes] } : {}),
    ...(env.typeParameters ? { typeParameters: new Set(env.typeParameters) } : {}),
    ...(env.deferName ? { deferName: env.deferName } : {}),
    ...(env.gotoLabels ? { gotoLabels: env.gotoLabels } : {}),
    ...(env.gotoPcName ? { gotoPcName: env.gotoPcName } : {}),
    ...(env.gotoLoopName ? { gotoLoopName: env.gotoLoopName } : {}),
    ...(env.gotoBreakLabels ? { gotoBreakLabels: new Map(env.gotoBreakLabels) } : {}),
    ...(env.loopContinueLabel ? { loopContinueLabel: env.loopContinueLabel } : {}),
    ...(env.iotaValue !== undefined ? { iotaValue: env.iotaValue } : {})
  };
}

function emitStatements(ctx: EmitterContext, statements: Statement[], env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (!env.gotoPcName && statementsRequireGotoStateMachine(statements)) {
    return emitGotoStateMachineStatements(ctx, statements, env, indent, (loopName) => [`break ${loopName};`]);
  }
  const localForward = findLocalForwardGoto(statements);
  if (localForward) return emitLocalForwardGotoStatements(ctx, statements, env, indent, localForward);
  const lines: string[] = [];
  for (const statement of statements) {
    const emitted = emitStatement(ctx, statement, env, indent);
    if (!emitted) return undefined;
    lines.push(...emitted);
  }
  return lines;
}

function findLocalForwardGoto(statements: Statement[]): { label: string; index: number } | undefined {
  for (const [index, statement] of statements.entries()) {
    if (statement.kind !== "LabeledStatement") continue;
    const prior = statements.slice(0, index);
    if (statementsContainGotoTo(prior, statement.label)) return { label: statement.label, index };
  }
  return undefined;
}

function emitLocalForwardGotoStatements(
  ctx: EmitterContext,
  statements: Statement[],
  env: ExpressionEmitEnv,
  indent: string,
  target: { label: string; index: number }
): string[] | undefined {
  const jsLabel = ctx.symbol(`label_${target.label}`);
  const beforeEnv = cloneExpressionEnv(env);
  beforeEnv.gotoBreakLabels = new Map(beforeEnv.gotoBreakLabels);
  beforeEnv.gotoBreakLabels.set(target.label, jsLabel);
  const before = emitStatements(ctx, statements.slice(0, target.index), beforeEnv, `${indent}  `);
  if (!before) return undefined;
  const labeled = statements[target.index] as LabeledStatement;
  const labelStatement = labeled.statement ? emitStatement(ctx, labeled.statement, env, indent) : [];
  if (!labelStatement) return undefined;
  const after = emitStatements(ctx, statements.slice(target.index + 1), env, indent);
  if (!after) return undefined;
  return [
    `${indent}${jsLabel}: {`,
    ...before,
    `${indent}}`,
    ...labelStatement,
    ...after
  ];
}

function emitStatement(ctx: EmitterContext, statement: Statement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  switch (statement.kind) {
    case "BlockStatement":
      return emitBlockStatement(ctx, statement, env, indent);
    case "LabeledStatement":
      return emitLabeledStatement(ctx, statement, env, indent);
    case "ConstDecl":
    case "VarDecl":
      return emitLocalDeclarationStatement(ctx, statement, env, indent);
    case "TypeDecl":
      return emitLocalTypeDeclarationStatement(statement, env);
    case "ReturnStatement":
      return emitReturnStatement(ctx, statement, env, indent);
    case "IfStatement":
      return emitIfStatement(ctx, statement, env, indent);
    case "SwitchStatement":
      return emitSwitchStatement(ctx, statement, env, indent);
    case "ForStatement":
      return emitForStatement(ctx, statement, env, indent);
    case "BranchStatement":
      return emitBranchStatement(ctx, statement, env, indent);
    case "AssignStatement":
      return emitAssignStatement(ctx, statement, env, indent);
    case "ShortVarStatement":
      return emitShortVarStatement(ctx, statement, env, indent);
    case "IncDecStatement":
      return emitIncDecStatement(ctx, statement, env, indent);
    case "ExpressionStatement":
      return emitExpressionStatement(ctx, statement, env, indent);
    case "DeferStatement":
      return emitDeferStatement(ctx, statement, env, indent);
    case "GoStatement":
      return emitGoStatement(ctx, statement, env, indent);
    case "SendStatement":
      return emitSendStatement(ctx, statement, env, indent);
    case "SelectStatement":
      return emitSelectStatement(ctx, statement, env, indent);
    default:
      ctx.emitError(`unsupported Stage 4 statement ${(statement as Statement).kind}`);
      return undefined;
  }
}

function emitLocalTypeDeclarationStatement(statement: Extract<Statement, { kind: "TypeDecl" }>, env: ExpressionEmitEnv): string[] {
  for (const declaration of statement.declarations) {
    env.localTypeUnderlyings.set(declaration.name, declaration.type.text);
  }
  return [];
}

function emitBlockStatement(ctx: EmitterContext, statement: BlockStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const blockEnv = cloneExpressionEnv(env);
  const body = emitStatements(ctx, statement.statements, blockEnv, `${indent}  `);
  if (!body) return undefined;
  return [
    `${indent}{`,
    ...body,
    `${indent}}`
  ];
}

function emitLabeledStatement(ctx: EmitterContext, statement: LabeledStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (!statement.statement) return [`${indent}${statement.label}: {}`];
  if (statement.statement.kind === "SwitchStatement") {
    const emitted = emitStatement(ctx, statement.statement, env, `${indent}  `);
    if (!emitted) return undefined;
    return [
      `${indent}${statement.label}: {`,
      ...emitted,
      `${indent}}`
    ];
  }
  const emitted = emitStatement(ctx, statement.statement, env, indent);
  if (!emitted) return undefined;
  return [`${indent}${statement.label}:`, ...emitted];
}

function emitLocalDeclarationStatement(
  ctx: EmitterContext,
  statement: ConstDeclStatement | VarDeclStatement,
  env: ExpressionEmitEnv,
  indent: string
): string[] | undefined {
  const lines: string[] = [];
  const keyword = statement.kind === "ConstDecl" ? "const" : "let";
  let inheritedValues: Array<Expression | undefined> = [];
  let localIndex = 0;
  for (const group of declarationGroups(statement.declarations)) {
    const groupValues = statement.kind === "ConstDecl" ? declarationValueExpressions(group) : [];
    if (groupValues.some(Boolean)) inheritedValues = groupValues;
    for (const declaration of group) {
      if (declaration.name === "_") continue;
      const existing = statement.kind === "VarDecl" ? env.locals.get(declaration.name) : undefined;
      const name = existing ?? safeLocalName(declaration.name, localIndex++);
      const declarationEnv: ExpressionEmitEnv = {
        ...env,
        ...(declaration.iotaIndex !== undefined ? { iotaValue: declaration.iotaIndex } : {})
      };
      const effectiveValue = declaration.value ?? (statement.kind === "ConstDecl" ? inheritedValues[declaration.valueIndex ?? 0] : undefined);
      const rawValue = effectiveValue
        ? expressionToJs(ctx, effectiveValue, declarationEnv)
        : zeroValueForTypeInEnv(declaration.type?.text, declarationEnv);
      if (!rawValue) {
        ctx.emitError(`unsupported Stage 4 declaration for ${declaration.name}`);
        return undefined;
      }
      const typeText = declarationTypeText(declaration, effectiveValue);
      if (typeText) env.localTypes.set(declaration.name, typeText);
      const value = valueForTargetType(rawValue, typeText, expressionTypeText(effectiveValue, declarationEnv), declarationEnv);
      if (existing) {
        lines.push(`${indent}${existing} = ${value};`);
      } else {
        env.locals.set(declaration.name, name);
        lines.push(`${indent}${keyword} ${name} = ${value};`);
      }
    }
  }
  return lines;
}

function emitReturnStatement(ctx: EmitterContext, statement: ReturnStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const returned = returnExpression(ctx, statement, env);
  if (!returned) return undefined;
  return [`${indent}return ${returned};`];
}

function emitIfStatement(ctx: EmitterContext, statement: IfStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const scoped = Boolean(statement.init);
  const ifEnv = cloneExpressionEnv(env);
  const lines: string[] = [];
  if (scoped) lines.push(`${indent}{`);
  const activeIndent = scoped ? `${indent}  ` : indent;
  if (statement.init) {
    const initLines = emitStatement(ctx, statement.init, ifEnv, activeIndent);
    if (!initLines) return undefined;
    lines.push(...initLines);
  }
  const condition = expressionToJs(ctx, statement.condition, ifEnv);
  if (!condition) return undefined;
  const thenEnv = cloneExpressionEnv(ifEnv);
  const thenLines = emitStatements(ctx, statement.thenBlock.statements, thenEnv, `${activeIndent}  `);
  if (!thenLines) return undefined;
  lines.push(`${activeIndent}if (${condition}) {`);
  lines.push(...thenLines);
  if (statement.elseBranch) {
    if (statement.elseBranch.kind === "IfStatement") {
      const elseLines = emitIfStatement(ctx, statement.elseBranch, cloneExpressionEnv(ifEnv), activeIndent);
      if (!elseLines) return undefined;
      const [first, ...rest] = elseLines;
      if (!first) return undefined;
      lines.push(`${activeIndent}} else ${first.trimStart()}`);
      lines.push(...rest);
    } else {
      const elseEnv = cloneExpressionEnv(ifEnv);
      const elseLines = emitStatements(ctx, statement.elseBranch.statements, elseEnv, `${activeIndent}  `);
      if (!elseLines) return undefined;
      lines.push(`${activeIndent}} else {`);
      lines.push(...elseLines);
      lines.push(`${activeIndent}}`);
    }
  } else {
    lines.push(`${activeIndent}}`);
  }
  if (scoped) lines.push(`${indent}}`);
  return lines;
}

function emitForStatement(ctx: EmitterContext, statement: ForStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (statement.range) return emitRangeStatement(ctx, statement, env, indent);
  const loopEnv = cloneExpressionEnv(env);
  const init = statement.init ? emitForHeaderStatement(ctx, statement.init, loopEnv, "init") : "";
  const condition = statement.condition ? expressionToJs(ctx, statement.condition, loopEnv) : "";
  const post = statement.post ? emitForHeaderStatement(ctx, statement.post, loopEnv, "post") : "";
  if (init === undefined || condition === undefined || post === undefined) return undefined;
  const bodyEnv = cloneExpressionEnv(loopEnv);
  const loopLabel = ctx.symbol("forLoop");
  bodyEnv.loopContinueLabel = loopLabel;
  const body = emitStatements(ctx, statement.body.statements, bodyEnv, `${indent}  `);
  if (!body) return undefined;
  return [
    `${indent}${loopLabel}: for (${init}; ${condition}; ${post}) {`,
    ...body,
    `${indent}}`
  ];
}

function emitRangeStatement(ctx: EmitterContext, statement: ForStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (!statement.range) return undefined;
  const source = expressionToJs(ctx, statement.range.source, env);
  if (!source) return undefined;
  const key = ctx.symbol("key");
  const value = ctx.symbol("value");
  const loopEnv = cloneExpressionEnv(env);
  const bindLines: string[] = [];
  const rangeTypes = rangeIterationTypeTexts(statement.range.source, env);
  const bindRangePart = (
    name: string | undefined,
    target: Expression | undefined,
    part: string,
    index: number,
    targetType: string | undefined
  ): string[] | undefined => {
    const assignedPart = targetType ? valueForTargetType(part, targetType, undefined, loopEnv) : part;
    if (name !== undefined) {
      if (name === "_") return [];
      const jsName = safeLocalName(name, index);
      if (statement.range?.define || !loopEnv.locals.has(name)) {
        loopEnv.locals.set(name, jsName);
        if (name === statement.range?.keyName && rangeTypes.key) loopEnv.localTypes.set(name, rangeTypes.key);
        if (name === statement.range?.valueName && rangeTypes.value) loopEnv.localTypes.set(name, rangeTypes.value);
        return [`${indent}  let ${jsName} = ${assignedPart};`];
      }
      return [`${indent}  ${loopEnv.locals.get(name)} = ${assignedPart};`];
    }
    if (!target) return [];
    return assignExpressionTargetLines(ctx, target, assignedPart, loopEnv, `${indent}  `, targetType);
  };
  const keyLines = bindRangePart(statement.range.keyName, statement.range.keyTarget, key, 0, rangeTypes.key);
  const valueLines = bindRangePart(statement.range.valueName, statement.range.valueTarget, value, 1, rangeTypes.value);
  if (!keyLines || !valueLines) return undefined;
  bindLines.push(...keyLines, ...valueLines);
  const bodyEnv = cloneExpressionEnv(loopEnv);
  const loopLabel = ctx.symbol("rangeLoop");
  bodyEnv.loopContinueLabel = loopLabel;
  const body = emitStatements(ctx, statement.body.statements, bodyEnv, `${indent}  `);
  if (!body) return undefined;
  return [
    `${indent}${loopLabel}: for (const [${key}, ${value}] of await __gojrRangeEntries(${source})) {`,
    ...bindLines,
    ...body,
    `${indent}}`
  ];
}

function emitSwitchStatement(ctx: EmitterContext, statement: SwitchStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (statement.typeSwitch) {
    return emitTypeSwitchStatement(ctx, statement, env, indent);
  }
  const scoped = Boolean(statement.init);
  const switchEnv = cloneExpressionEnv(env);
  const lines: string[] = [];
  if (scoped) lines.push(`${indent}{`);
  const activeIndent = scoped ? `${indent}  ` : indent;
  if (statement.init) {
    const initLines = emitStatement(ctx, statement.init, switchEnv, activeIndent);
    if (!initLines) return undefined;
    lines.push(...initLines);
  }
  const switchValue = ctx.symbol("switch");
  const value = statement.expression ? expressionToJs(ctx, statement.expression, switchEnv) : "true";
  if (!value) return undefined;
  lines.push(`${activeIndent}const ${switchValue} = ${value};`);
  lines.push(`${activeIndent}switch (true) {`);
  for (const clause of statement.clauses) {
    if (clause.default) {
      lines.push(`${activeIndent}  default: {`);
    } else {
      if (clause.values.length === 0) {
        ctx.emitError("unsupported Stage 4 switch clause without values");
        return undefined;
      }
      for (const expression of clause.values) {
        const clauseValue = expressionToJs(ctx, expression, switchEnv);
        if (!clauseValue) return undefined;
        lines.push(`${activeIndent}  case __gojrEqual(${switchValue}, ${clauseValue}):`);
      }
      lines.push(`${activeIndent}  {`);
    }
    const fallthrough = clause.statements[clause.statements.length - 1]?.kind === "BranchStatement" &&
      (clause.statements[clause.statements.length - 1] as BranchStatement).branch === "fallthrough";
    if (clause.statements.some((item, index) =>
      item.kind === "BranchStatement" && item.branch === "fallthrough" && index !== clause.statements.length - 1
    )) {
      ctx.emitError("fallthrough must be the final statement in a switch clause");
      return undefined;
    }
    const bodyStatements = fallthrough ? clause.statements.slice(0, -1) : clause.statements;
    const body = emitStatements(ctx, bodyStatements, cloneExpressionEnv(switchEnv), `${activeIndent}    `);
    if (!body) return undefined;
    lines.push(...body);
    if (!fallthrough) lines.push(`${activeIndent}    break;`);
    lines.push(`${activeIndent}  }`);
  }
  lines.push(`${activeIndent}}`);
  if (scoped) lines.push(`${indent}}`);
  return lines;
}

function emitTypeSwitchStatement(ctx: EmitterContext, statement: SwitchStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (!statement.typeSwitch) return undefined;
  const scoped = Boolean(statement.init);
  const switchEnv = cloneExpressionEnv(env);
  const lines: string[] = [];
  if (scoped) lines.push(`${indent}{`);
  const activeIndent = scoped ? `${indent}  ` : indent;
  if (statement.init) {
    const initLines = emitStatement(ctx, statement.init, switchEnv, activeIndent);
    if (!initLines) return undefined;
    lines.push(...initLines);
  }
  const switchValue = ctx.symbol("typeswitch");
  const value = expressionToJs(ctx, statement.typeSwitch.expression, switchEnv);
  if (!value) return undefined;
  lines.push(`${activeIndent}const ${switchValue} = ${value};`);
  lines.push(`${activeIndent}switch (true) {`);
  for (const clause of statement.clauses) {
    if (clause.statements.some((item) => item.kind === "BranchStatement" && item.branch === "fallthrough")) {
      ctx.emitError("fallthrough is not allowed in type switches");
      return undefined;
    }
    if (clause.default) {
      lines.push(`${activeIndent}  default: {`);
    } else {
      if (!clause.typeValues || clause.typeValues.length === 0) {
        ctx.emitError("unsupported Stage 4 type switch clause without type values");
        return undefined;
      }
      for (const type of clause.typeValues) {
        lines.push(`${activeIndent}  case __gojrValueMatchesType(${switchValue}, ${JSON.stringify(type.text)}):`);
      }
      lines.push(`${activeIndent}  {`);
    }
    const clauseEnv = cloneExpressionEnv(switchEnv);
    const binding = emitTypeSwitchBinding(statement, clause, clauseEnv, switchValue, `${activeIndent}    `);
    if (!binding) return undefined;
    const body = emitStatements(ctx, clause.statements, clauseEnv, `${activeIndent}    `);
    if (!body) return undefined;
    lines.push(...binding, ...body, `${activeIndent}    break;`, `${activeIndent}  }`);
  }
  lines.push(`${activeIndent}}`);
  if (scoped) lines.push(`${indent}}`);
  return lines;
}

function emitTypeSwitchBinding(
  statement: SwitchStatement,
  clause: SwitchStatement["clauses"][number],
  env: ExpressionEmitEnv,
  switchValue: string,
  indent: string
): string[] | undefined {
  const guard = statement.typeSwitch;
  if (!guard?.name || guard.name === "_") return [];
  const singleType = !clause.default && (clause.typeValues?.length ?? 0) === 1 ? clause.typeValues?.[0]?.text : undefined;
  const value = singleType ? `__gojrTypeAssert(${switchValue}, ${JSON.stringify(singleType)})` : switchValue;
  if (guard.define || !env.locals.has(guard.name)) {
    const local = safeLocalName(guard.name, 0);
    env.locals.set(guard.name, local);
    if (singleType) env.localTypes.set(guard.name, singleType);
    return [`${indent}let ${local} = ${value};`];
  }
  return [`${indent}${env.locals.get(guard.name)} = ${value};`];
}

function emitBranchStatement(ctx: EmitterContext, statement: BranchStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (statement.branch === "goto") {
    const breakLabel = statement.label ? env.gotoBreakLabels?.get(statement.label) : undefined;
    if (breakLabel) return [`${indent}break ${breakLabel};`];
    const target = statement.label ? env.gotoLabels?.get(statement.label) : undefined;
    if (!env.gotoPcName || target === undefined) {
      ctx.emitError(`unsupported Stage 4 unresolved goto ${statement.label ?? "<missing>"}`);
      return undefined;
    }
    return [
      `${indent}${env.gotoPcName} = ${target};`,
      `${indent}continue ${env.gotoLoopName ?? "__gojrGotoLoop"};`
    ];
  }
  if (statement.branch === "fallthrough") return [`${indent}/* fallthrough */`];
  if (statement.branch === "continue" && !statement.label && env.gotoPcName && env.loopContinueLabel) {
    return [`${indent}continue ${env.loopContinueLabel};`];
  }
  return [`${indent}${statement.branch}${statement.label ? ` ${statement.label}` : ""};`];
}

function emitAssignStatement(ctx: EmitterContext, statement: AssignStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (statement.operator !== "=") {
    if (statement.targets.length !== 1 || statement.values.length !== 1) {
      ctx.emitError(`unsupported Stage 4 compound assignment shape`);
      return undefined;
    }
    const current = expressionToJs(ctx, statement.targets[0], env);
    const value = expressionToJs(ctx, statement.values[0], env);
    if (!current || !value) return undefined;
    const compound = compoundAssignmentExpression(ctx, statement.operator, current, value, expressionTypeText(statement.targets[0], env));
    if (!compound) return undefined;
    return assignExpressionTargetLines(ctx, statement.targets[0], compound, env, indent, expressionTypeText(statement.values[0], env));
  }
  if (statement.values.length === 1 && statement.targets.length > 1) {
    const source = statement.values[0];
    if (!source) return undefined;
    const tuple = tupleSourceExpressionToJs(ctx, source, statement.targets.length, env);
    if (!tuple) return undefined;
    const temp = ctx.symbol("assign");
    const lines = [`${indent}const ${temp} = ${tuple};`];
    for (const [index, target] of statement.targets.entries()) {
      const assigned = assignExpressionTargetLines(ctx, target, `${temp}[${index}]`, env, indent, tupleSourceTypeTexts(source, statement.targets.length, env)[index]);
      if (!assigned) return undefined;
      lines.push(...assigned);
    }
    return lines;
  }
  if (statement.values.length !== statement.targets.length) {
    ctx.emitError(`unsupported Stage 4 assignment count ${statement.targets.length} targets and ${statement.values.length} values`);
    return undefined;
  }
  const values = statement.values.map((value) => expressionToJs(ctx, value, env));
  if (values.some((value) => value === undefined)) return undefined;
  const sourceTypes = statement.values.map((value) => expressionTypeText(value, env));
  const temp = ctx.symbol("assign");
  const lines = [`${indent}const ${temp} = [${values.filter((value): value is string => value !== undefined).join(", ")}];`];
  for (const [index, target] of statement.targets.entries()) {
    const assigned = assignExpressionTargetLines(ctx, target, `${temp}[${index}]`, env, indent, sourceTypes[index]);
    if (!assigned) return undefined;
    lines.push(...assigned);
  }
  return lines;
}

function emitShortVarStatement(ctx: EmitterContext, statement: ShortVarStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (statement.values.length === 1 && statement.names.length > 1) {
    const source = statement.values[0];
    if (!source) return undefined;
    const tuple = tupleSourceExpressionToJs(ctx, source, statement.names.length, env);
    if (!tuple) return undefined;
    const temp = ctx.symbol("short");
    const sourceTypes = tupleSourceTypeTexts(source, statement.names.length, env);
    return [
      `${indent}const ${temp} = ${tuple};`,
      ...bindShortNames(statement.names, temp, env, indent, sourceTypes)
    ];
  }
  if (statement.values.length !== statement.names.length) {
    ctx.emitError(`unsupported Stage 4 short declaration count ${statement.names.length} names and ${statement.values.length} values`);
    return undefined;
  }
  const values = statement.values.map((value) => expressionToJs(ctx, value, env));
  if (values.some((value) => value === undefined)) return undefined;
  const temp = ctx.symbol("short");
  const sourceTypes = statement.values.map((value) => expressionTypeText(value, env));
  return [
    `${indent}const ${temp} = [${values.filter((value): value is string => value !== undefined).join(", ")}];`,
    ...bindShortNames(statement.names, temp, env, indent, sourceTypes)
  ];
}

function emitIncDecStatement(ctx: EmitterContext, statement: IncDecStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const current = expressionToJs(ctx, statement.target, env);
  if (!current) return undefined;
  const one = incDecOneLiteral(statement.target, env);
  const targetType = expressionTypeText(statement.target, env);
  const next = statement.operator === "++"
    ? `__gojrAdd(${current}, ${one}, ${JSON.stringify(targetType ?? "")})`
    : `__gojrSub(${current}, ${one}, ${JSON.stringify(targetType ?? "")})`;
  return assignExpressionTargetLines(ctx, statement.target, next, env, indent);
}

function incDecOneLiteral(target: Expression, env: ExpressionEmitEnv): string {
  const typeText = expressionTypeText(target, env);
  const resolvedType = typeText ? resolveUnderlyingTypeTextInEnv(typeText, env) : "";
  return isIntegerType(resolvedType) ? "1n" : "1";
}

function emitExpressionStatement(ctx: EmitterContext, statement: ExpressionStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const expression = expressionToJs(ctx, statement.expression, env);
  if (!expression) return undefined;
  return [`${indent}${expression};`];
}

function emitDeferStatement(ctx: EmitterContext, statement: DeferStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  if (!env.deferName) {
    ctx.emitError("defer used outside generated function defer scope");
    return undefined;
  }
  if (statement.expression.kind === "CallExpression") {
    const callee = deferredCalleeToJs(ctx, statement.expression, env);
    if (!callee) return undefined;
    const args = renderCallArgs(ctx, statement.expression.args, env, [], statement.expression.spreadLast);
    if (!args) return undefined;
    const temp = ctx.symbol("deferArgs");
    return [
      `${indent}const ${temp} = [${args.filter((arg): arg is string => arg !== undefined).join(", ")}];`,
      `${indent}${env.deferName}(async () => await (${callee})(...${temp}));`
    ];
  }
  const expression = expressionToJs(ctx, statement.expression, env);
  if (!expression) return undefined;
  return [`${indent}${env.deferName}(async () => ${expression});`];
}

function emitGoStatement(ctx: EmitterContext, statement: Extract<Statement, { kind: "GoStatement" }>, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const call = callExpressionToJs(ctx, statement.call, env);
  if (!call) return undefined;
  return [`${indent}__gojrGo(async () => { ${call}; });`];
}

function emitSendStatement(ctx: EmitterContext, statement: Extract<Statement, { kind: "SendStatement" }>, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const channel = expressionToJs(ctx, statement.channel, env);
  const value = expressionToJs(ctx, statement.value, env);
  if (!channel || !value) return undefined;
  return [`${indent}await __gojrChanSend(${channel}, ${value});`];
}

function emitSelectStatement(ctx: EmitterContext, statement: SelectStatement, env: ExpressionEmitEnv, indent: string): string[] | undefined {
  const cases: string[] = [];
  for (const clause of statement.clauses) {
    const clauseEnv = cloneExpressionEnv(env);
    const selected = ctx.symbol("selected");
    const bodyLines: string[] = [];
    const prepared = emitSelectComm(ctx, clause.comm, clause.default, clauseEnv, `${indent}      `, selected);
    if (!prepared) return undefined;
    bodyLines.push(...prepared.prefixLines);
    const emittedBody = emitStatements(ctx, clause.statements, clauseEnv, `${indent}      `);
    if (!emittedBody) return undefined;
    bodyLines.push(...emittedBody);
    cases.push([
      `${indent}  {`,
      `${indent}    op: ${JSON.stringify(prepared.op)},`,
      ...(prepared.channel ? [`${indent}    channel: ${prepared.channel},`] : []),
      ...(prepared.value ? [`${indent}    value: ${prepared.value},`] : []),
      `${indent}    apply: async (${selected}) => {`,
      ...bodyLines,
      `${indent}    }`,
      `${indent}  }`
    ].join("\n"));
  }
  return [
    `${indent}await __gojrSelect([`,
    cases.join(",\n"),
    `${indent}]);`
  ];
}

function emitSelectComm(
  ctx: EmitterContext,
  statement: Statement | undefined,
  isDefault: boolean,
  env: ExpressionEmitEnv,
  indent: string,
  selected: string
): { op: "default" | "recv" | "send"; channel?: string; value?: string; prefixLines: string[] } | undefined {
  if (isDefault || !statement) return { op: "default", prefixLines: [] };
  if (statement.kind === "SendStatement") {
    const channel = expressionToJs(ctx, statement.channel, env);
    const value = expressionToJs(ctx, statement.value, env);
    if (!channel || !value) return undefined;
    return { op: "send", channel, value, prefixLines: [] };
  }
  const receive = selectReceiveOperand(statement);
  if (!receive) {
    ctx.emitError(`unsupported Stage 4 select communication statement ${statement.kind}`);
    return undefined;
  }
  const channel = expressionToJs(ctx, receive, env);
  if (!channel) return undefined;
  if (statement.kind === "ExpressionStatement") return { op: "recv", channel, prefixLines: [] };
  if (statement.kind === "ShortVarStatement") {
    return {
      op: "recv",
      channel,
      prefixLines: bindShortNames(statement.names, selected, env, indent, selectReceiveTypeTexts(receive, statement.names.length, env))
    };
  }
  if (statement.kind === "AssignStatement") {
    const prefixLines: string[] = [];
    for (const [index, target] of statement.targets.entries()) {
      const assigned = assignExpressionTargetLines(ctx, target, `${selected}[${index}]`, env, indent);
      if (!assigned) return undefined;
      prefixLines.push(...assigned);
    }
    return { op: "recv", channel, prefixLines };
  }
  ctx.emitError(`unsupported Stage 4 select receive statement ${statement.kind}`);
  return undefined;
}

function selectReceiveOperand(statement: Statement): Expression | undefined {
  if (statement.kind === "ExpressionStatement" &&
    statement.expression.kind === "UnaryExpression" &&
    statement.expression.operator === "<-") {
    return statement.expression.operand;
  }
  if ((statement.kind === "ShortVarStatement" || statement.kind === "AssignStatement") &&
    statement.values.length === 1 &&
    statement.values[0]?.kind === "UnaryExpression" &&
    statement.values[0].operator === "<-") {
    return statement.values[0].operand;
  }
  return undefined;
}

function selectReceiveTypeTexts(channel: Expression, targetCount: number, env: ExpressionEmitEnv): Array<string | undefined> {
  const elementType = chanElementTypeText(expressionTypeText(channel, env));
  return targetCount === 2 ? [elementType, "bool"] : [elementType];
}

function deferredCalleeToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
    return `pkg[${JSON.stringify(expression.callee.name)}]`;
  }
  if (expression.callee.kind === "SelectorExpression") {
    if (isImportedPackageSelector(expression.callee, env) && expression.callee.object.kind === "Identifier") {
      return `__gojrPackageCallableSelector(${JSON.stringify(env.facts.imports.get(expression.callee.object.name))}, ${JSON.stringify(expression.callee.field)})`;
    }
    const methodKey = methodPackageKeyForSelector(expression.callee, env);
    const receiver = methodKey
      ? methodReceiverToJs(ctx, expression.callee, methodKey, env)
      : expressionToJs(ctx, expression.callee.object, env);
    if (!receiver) return undefined;
    return `(async (...__gojrMethodArgs) => await __gojrCallMethod(${receiver}, ${JSON.stringify(expression.callee.field)}, __gojrMethodArgs, pkg))`;
  }
  return expressionToJs(ctx, expression.callee, env);
}

function emitForHeaderStatement(
  ctx: EmitterContext,
  statement: Statement,
  env: ExpressionEmitEnv,
  position: "init" | "post"
): string | undefined {
  switch (statement.kind) {
    case "ShortVarStatement":
      if (position === "post") {
        ctx.emitError("short variable declaration is not allowed in a for post statement");
        return undefined;
      }
      if (statement.names.some((name) => name === "_")) {
        ctx.emitError("unsupported Stage 4 for init short declaration");
        return undefined;
      }
      {
        if (statement.values.length === 1 && statement.names.length > 1) {
          const source = statement.values[0];
          if (!source) return undefined;
          const tuple = tupleSourceExpressionToJs(ctx, source, statement.names.length, env);
          if (!tuple) {
            ctx.emitError("unsupported Stage 4 for init short declaration");
            return undefined;
          }
          const names = statement.names.map((goName, index) => {
            const name = safeLocalName(goName, index);
            env.locals.set(goName, name);
            return name;
          });
          const sourceTypes = tupleSourceTypeTexts(source, statement.names.length, env);
          for (const [index, goName] of statement.names.entries()) {
            const sourceType = sourceTypes[index];
            if (sourceType) env.localTypes.set(goName, sourceType);
          }
          return `let [${names.join(", ")}] = ${tuple}`;
        }
        if (statement.names.length !== statement.values.length) {
          ctx.emitError("unsupported Stage 4 for init short declaration");
          return undefined;
        }
        const declarations: string[] = [];
        for (const [index, goName] of statement.names.entries()) {
          const expression = statement.values[index];
          if (!goName || !expression) return undefined;
          const value = expressionToJs(ctx, expression, env);
          if (!value) return undefined;
          const name = safeLocalName(goName, index);
          env.locals.set(goName, name);
          const typeText = expressionTypeText(expression, env);
          if (typeText) env.localTypes.set(goName, typeText);
          declarations.push(`${name} = ${value}`);
        }
        return `let ${declarations.join(", ")}`;
      }
    case "AssignStatement":
      if (statement.values.length === 1 && statement.targets.length > 1) {
        const source = statement.values[0];
        if (!source) return undefined;
        const tuple = tupleSourceExpressionToJs(ctx, source, statement.targets.length, env);
        if (!tuple) {
          ctx.emitError("unsupported Stage 4 for header assignment");
          return undefined;
        }
        const targets = statement.targets.map((target) => expressionToJs(ctx, target, env));
        if (targets.some((target) => target === undefined)) return undefined;
        return `([${targets.filter((target): target is string => target !== undefined).join(", ")}] = ${tuple})`;
      }
      if (statement.targets.length !== statement.values.length) {
        ctx.emitError("unsupported Stage 4 for header assignment");
        return undefined;
      }
      {
        const parts: string[] = [];
        for (const [index, targetExpression] of statement.targets.entries()) {
          const target = expressionToJs(ctx, targetExpression, env);
          const valueExpression = statement.values[index];
          const value = valueExpression ? expressionToJs(ctx, valueExpression, env) : undefined;
          if (!target || !value) return undefined;
          if (statement.operator === "=") {
            parts.push(`${target} = ${value}`);
            continue;
          }
          if (statement.targets.length !== 1) {
            ctx.emitError("unsupported Stage 4 compound multi-assignment in for header");
            return undefined;
          }
          const compound = compoundAssignmentExpression(ctx, statement.operator, target, value, expressionTypeText(targetExpression, env));
          if (!compound) return undefined;
          parts.push(`${target} = ${compound}`);
        }
        return parts.join(", ");
      }
    case "IncDecStatement":
      {
        const target = expressionToJs(ctx, statement.target, env);
        if (!target) return undefined;
        const one = incDecOneLiteral(statement.target, env);
        return `${target} = ${statement.operator === "++" ? "__gojrAdd" : "__gojrSub"}(${target}, ${one}, ${JSON.stringify(expressionTypeText(statement.target, env) ?? "")})`;
      }
    case "ExpressionStatement":
      return expressionToJs(ctx, statement.expression, env);
    default:
      ctx.emitError(`unsupported Stage 4 for ${position} statement ${statement.kind}`);
      return undefined;
  }
}

function tupleSourceExpressionToJs(ctx: EmitterContext, expression: Expression, targetCount: number, env: ExpressionEmitEnv): string | undefined {
  if (targetCount === 2 && expression.kind === "IndexExpression") {
    const object = expressionToJs(ctx, expression.object, env);
    const index = expressionToJs(ctx, expression.index, env);
    if (!object || !index) return undefined;
    const objectType = expressionTypeText(expression.object, env) ?? packageExportTypeText(ctx, expression.object);
    const mapType = parseMapTypeText(objectType ? resolveUnderlyingTypeTextInEnv(objectType, env) : undefined);
    const keyValue = mapType ? valueForTargetType(index, mapType.keyType, expressionTypeText(expression.index, env), env) : index;
    const zero = zeroValueForTypeInEnv(mapType?.valueType ?? mapValueTypeText(objectType, env.facts) ?? expression.typeText, env) ??
      "null";
    return `__gojrMapGetOk(${object}, ${keyValue}, ${zero})`;
  }
  if (targetCount === 2 && expression.kind === "TypeAssertionExpression") {
    const value = expressionToJs(ctx, expression.expression, env);
    if (!value) return undefined;
    return `__gojrTypeAssertOk(${value}, ${JSON.stringify(expression.type.text)})`;
  }
  if (targetCount === 2 && expression.kind === "UnaryExpression" && expression.operator === "<-") {
    const channel = expressionToJs(ctx, expression.operand, env);
    if (!channel) return undefined;
    return `(await __gojrChanRecv(${channel}))`;
  }
  const value = expressionToJs(ctx, expression, env);
  if (!value) return undefined;
  return value;
}

function bindShortNames(names: string[], tuple: string, env: ExpressionEmitEnv, indent: string, sourceTypes: Array<string | undefined> = []): string[] {
  const lines: string[] = [];
  for (const [index, name] of names.entries()) {
    if (name === "_") continue;
    const existing = env.locals.get(name);
    if (existing) {
      lines.push(`${indent}${existing} = ${valueForTargetType(`${tuple}[${index}]`, env.localTypes.get(name), sourceTypes[index], env)};`);
      continue;
    }
    const local = safeLocalName(name, index);
    env.locals.set(name, local);
    const source = sourceTypes[index];
    if (source) env.localTypes.set(name, source);
    lines.push(`${indent}let ${local} = ${source ? valueForTargetType(`${tuple}[${index}]`, source, source, env) : `${tuple}[${index}]`};`);
  }
  return lines;
}

function assignExpressionTargetLines(
  ctx: EmitterContext,
  target: Expression | undefined,
  value: string,
  env: ExpressionEmitEnv,
  indent: string,
  sourceType?: string
): string[] | undefined {
  if (!target) return undefined;
  switch (target.kind) {
    case "Identifier": {
      if (target.name === "_") return [];
      const local = env.locals.get(target.name);
      const targetType = env.localTypes.get(target.name) ?? env.facts.packageTypes.get(target.name);
      return [`${indent}${local ?? `pkg[${JSON.stringify(target.name)}]`} = ${valueForTargetType(value, targetType, sourceType, env)};`];
    }
    case "SelectorExpression": {
      const object = expressionToJs(ctx, target.object, env);
      if (!object) return undefined;
      return [`${indent}__gojrSetField(${object}, ${JSON.stringify(target.field)}, ${valueForTargetType(value, selectorFieldTypeText(target, env), sourceType, env)});`];
    }
    case "IndexExpression": {
      const object = expressionToJs(ctx, target.object, env);
      const index = expressionToJs(ctx, target.index, env);
      if (!object || !index) return undefined;
      const objectType = expressionTypeText(target.object, env) ?? packageExportTypeText(ctx, target.object) ?? "";
      const resolvedObjectType = resolveUnderlyingTypeTextInEnv(objectType, env);
      const mapType = parseMapTypeText(resolvedObjectType);
      if (mapType) {
        const keyValue = valueForTargetType(index, mapType.keyType, expressionTypeText(target.index, env), env);
        return [`${indent}__gojrMapSet(${object}, ${keyValue}, ${valueForTargetType(value, mapType.valueType, sourceType, env)});`];
      }
      const elementType = parseArrayOrSliceTypeText(resolvedObjectType)?.elementType;
      return [`${indent}__gojrSetIndex(${object}, ${index}, ${valueForTargetType(value, elementType, sourceType, env)});`];
    }
    case "UnaryExpression": {
      if (target.operator !== "*") {
        ctx.emitError(`unsupported Stage 4 assignment target ${target.operator}${target.kind}`);
        return undefined;
      }
      const pointer = expressionToJs(ctx, target.operand, env);
      if (!pointer) return undefined;
      return [`${indent}__gojrSetDeref(${pointer}, ${value});`];
    }
    default:
      ctx.emitError(`unsupported Stage 4 assignment target ${target.kind}`);
      return undefined;
  }
}

function compoundAssignmentExpression(
  ctx: EmitterContext,
  operator: AssignStatement["operator"],
  left: string,
  right: string,
  resultType?: string
): string | undefined {
  if (operator === "=") return right;
  const binary = operator.slice(0, -1) as BinaryExpression["operator"];
  const rendered = binaryOperationToJs(binary, left, right, resultType);
  if (!rendered) {
    ctx.emitError(`unsupported Stage 4 compound operator ${operator}`);
    return undefined;
  }
  return rendered;
}

function functionParameterBindings(fn: FunctionDecl, facts: PackageEmitFacts): { params: string[]; env: ExpressionEmitEnv } {
  const locals = new Map<string, string>();
  const localTypes = new Map<string, string>();
  const localTypeUnderlyings = new Map<string, string>();
  const params: string[] = [];
  const typeParameters = fn.typeParameters && fn.typeParameters.length > 0
    ? new Set(fn.typeParameters)
    : undefined;
  if (typeParameters) params.push("__gojrTypeArgs");
  if (fn.receiver) {
    const receiverName = fn.receiver.name && fn.receiver.name !== "_" ? fn.receiver.name : "__gojrRecv";
    const receiverJsName = safeLocalName(receiverName, -1);
    if (fn.receiver.name && fn.receiver.name !== "_") {
      locals.set(fn.receiver.name, receiverJsName);
      localTypes.set(fn.receiver.name, fn.receiver.type.text);
    }
    params.push(receiverJsName);
  }
  params.push(...fn.signature.parameters.map((parameter, index) => {
    const goName = parameter.name && parameter.name !== "_" ? parameter.name : `__gojrArg${index}`;
    const jsName = safeLocalName(goName, index);
    if (parameter.name && parameter.name !== "_") {
      locals.set(parameter.name, jsName);
      localTypes.set(parameter.name, parameter.variadic ? `[]${parameter.type.text}` : parameter.type.text);
    }
    return parameter.variadic ? `...${jsName}` : jsName;
  }));
  return { params, env: { locals, localTypes, localTypeUnderlyings, facts, ...(typeParameters ? { typeParameters } : {}) } };
}

function functionDeclarationParameterCoercionLines(fn: FunctionDecl, env: ExpressionEmitEnv, indent: string): string[] {
  const lines: string[] = [];
  if (fn.receiver) {
    const receiverName = fn.receiver.name && fn.receiver.name !== "_" ? fn.receiver.name : "__gojrRecv";
    const receiverJsName = safeLocalName(receiverName, -1);
    lines.push(`${indent}${receiverJsName} = ${valueForTargetType(receiverJsName, fn.receiver.type.text, fn.receiver.type.text, env)};`);
  }
  for (const [index, parameter] of fn.signature.parameters.entries()) {
    const goName = parameter.name && parameter.name !== "_" ? parameter.name : `__gojrArg${index}`;
    const jsName = safeLocalName(goName, index);
    if (parameter.variadic) {
      lines.push(`${indent}${jsName} = Array.from(${jsName}, (__gojrItem) => ${valueForTargetType("__gojrItem", parameter.type.text, parameter.type.text, env)});`);
    } else {
      lines.push(`${indent}${jsName} = ${valueForTargetType(jsName, parameter.type.text, parameter.type.text, env)};`);
    }
  }
  return lines;
}

function wasmScalarLoweringForFunction(fn: FunctionDecl): WasmScalarLowering | undefined {
  if (fn.receiver) return undefined;
  if (fn.typeParameters && fn.typeParameters.length > 0) return undefined;
  if (fn.signature.parameters.length !== 2 || fn.signature.results.length !== 1) return undefined;
  if (!fn.signature.parameters.every((param) => param.name && param.type.text === "int64")) return undefined;
  const statement = onlyReturn(fn.body);
  if (!statement || statement.values.length !== 1) return undefined;
  const value = statement.values[0];
  if (!value || value.kind !== "BinaryExpression") return undefined;
  if (!isIdentifier(value.left, fn.signature.parameters[0]?.name)) return undefined;
  if (!isIdentifier(value.right, fn.signature.parameters[1]?.name)) return undefined;
  const resultType = fn.signature.results[0]?.type.text;
  const lowering = i64ScalarLowerings[value.operator];
  if (!lowering || lowering.resultType !== resultType) return undefined;
  return lowering;
}

const i64ScalarLowerings: Partial<Record<BinaryExpression["operator"], WasmScalarLowering>> = {
  "+": { exportName: "gojr_i64_add", resultType: "int64" },
  "-": { exportName: "gojr_i64_sub", resultType: "int64" },
  "*": { exportName: "gojr_i64_mul", resultType: "int64" },
  "/": { exportName: "gojr_i64_div", resultType: "int64" },
  "%": { exportName: "gojr_i64_rem", resultType: "int64" },
  "&": { exportName: "gojr_i64_and", resultType: "int64" },
  "|": { exportName: "gojr_i64_or", resultType: "int64" },
  "^": { exportName: "gojr_i64_xor", resultType: "int64" },
  "&^": { exportName: "gojr_i64_bitclear", resultType: "int64" },
  "<<": { exportName: "gojr_i64_shl", resultType: "int64" },
  ">>": { exportName: "gojr_i64_shr", resultType: "int64" },
  "==": { exportName: "gojr_i64_eq", resultType: "bool" },
  "!=": { exportName: "gojr_i64_ne", resultType: "bool" },
  "<": { exportName: "gojr_i64_lt", resultType: "bool" },
  "<=": { exportName: "gojr_i64_le", resultType: "bool" },
  ">": { exportName: "gojr_i64_gt", resultType: "bool" },
  ">=": { exportName: "gojr_i64_ge", resultType: "bool" }
};

function onlyReturn(body: BlockStatement): ReturnStatement | undefined {
  return body.statements.length === 1 && body.statements[0]?.kind === "ReturnStatement"
    ? body.statements[0]
    : undefined;
}

function isIdentifier(expression: Expression, name: string | undefined): boolean {
  return Boolean(name && expression.kind === "Identifier" && expression.name === name);
}

function packageEmitFacts(ast: ProgramAst, artifact: GoJuniorPackageExportData): PackageEmitFacts {
  const methodKeys = new Map<string, string>();
  const resultCounts = new Map<string, number>();
  const packageTypes = new Map<string, string>();
  const typeUnderlyings = new Map<string, string>();
  const constValues = collectPackageConstValues(ast);
  const structFields = new Map<string, Array<{ name: string; type: string; embedded: boolean }>>();
  const imports = new Map<string, string>();
  const interfaceTypes = new Set<string>(["any", "interface{}", "error"]);
  const functionParamTypes = new Map<string, string[]>();
  const functionResultTypes = new Map<string, string[]>();
  const functionTypeParameters = new Map<string, string[]>();
  const methodPointerReceivers = new Map<string, boolean>();
  for (const item of ast.imports) {
    if (item.alias === "_" || item.alias === ".") continue;
    imports.set(item.alias ?? defaultImportName(item.path), item.path);
  }
  for (const exported of artifact.exports) {
    if (exported.typeText) packageTypes.set(exported.name, exported.typeText);
  }
  for (const statement of ast.body) {
    if (statement.kind === "TypeDecl") {
      for (const declaration of statement.declarations) {
        typeUnderlyings.set(declaration.name, declaration.type.text);
        if (declaration.structFields) {
          structFields.set(declaration.name, declaration.structFields.map((field) => ({
            name: field.name,
            type: field.type.text,
            embedded: Boolean(field.embedded)
          })));
        }
        if (isInterfaceTypeSpec(declaration)) interfaceTypes.add(declaration.name);
      }
      continue;
    }
    if (statement.kind === "ConstDecl" || statement.kind === "VarDecl") {
      for (const declaration of statement.declarations) {
        const typeText = declarationTypeText(declaration);
        if (typeText) packageTypes.set(declaration.name, typeText);
      }
    }
  }
  for (const fn of ast.functions) {
    const key = functionPackageKey(fn);
    resultCounts.set(key, fn.signature.results.length);
    functionParamTypes.set(key, fn.signature.parameters.map((parameter) => parameter.variadic ? `...${parameter.type.text}` : parameter.type.text));
    functionResultTypes.set(key, fn.signature.results.map((result) => result.type.text));
    if (fn.typeParameters && fn.typeParameters.length > 0) functionTypeParameters.set(key, fn.typeParameters);
    if (!fn.receiver) {
      resultCounts.set(fn.name, fn.signature.results.length);
      functionParamTypes.set(fn.name, fn.signature.parameters.map((parameter) => parameter.variadic ? `...${parameter.type.text}` : parameter.type.text));
      functionResultTypes.set(fn.name, fn.signature.results.map((result) => result.type.text));
      if (fn.typeParameters && fn.typeParameters.length > 0) functionTypeParameters.set(fn.name, fn.typeParameters);
      continue;
    }
    const receiver = receiverBaseType(fn.receiver.type.text);
    methodKeys.set(`${receiver}.${fn.name}`, key);
    methodPointerReceivers.set(key, fn.receiver.type.text.trim().startsWith("*"));
  }
  return {
    methodKeys,
    resultCounts,
    packageTypes,
    typeUnderlyings,
    constValues,
    structFields,
    imports,
    interfaceTypes,
    functionParamTypes,
    functionResultTypes,
    functionTypeParameters,
    methodPointerReceivers
  };
}

function collectPackageConstValues(ast: ProgramAst): Map<string, bigint> {
  const values = new Map<string, bigint>();
  for (const statement of ast.body) {
    if (statement.kind !== "ConstDecl") continue;
    let inheritedValues: Array<Expression | undefined> = [];
    for (const group of declarationGroups(statement.declarations)) {
      const groupValues = declarationValueExpressions(group);
      if (groupValues.some(Boolean)) inheritedValues = groupValues;
      for (const declaration of group) {
        if (declaration.name === "_") continue;
        const effectiveValue = declaration.value ?? inheritedValues[declaration.valueIndex ?? 0];
        const value = constantIntegerExpressionValue(
          effectiveValue,
          values,
          BigInt(declaration.iotaIndex ?? 0)
        );
        if (value !== undefined) values.set(declaration.name, value);
      }
    }
  }
  return values;
}

function constantIntegerExpressionValue(
  expression: Expression | undefined,
  constants: Map<string, bigint>,
  iotaValue: bigint
): bigint | undefined {
  if (!expression) return undefined;
  switch (expression.kind) {
    case "Literal":
      if (expression.literalKind !== "int" && expression.literalKind !== "rune") return undefined;
      if (typeof expression.value === "bigint") return expression.value;
      if (typeof expression.value === "number" && Number.isSafeInteger(expression.value)) return BigInt(expression.value);
      return parseIntegerLiteralText(expression.raw);
    case "Identifier":
      if (expression.name === "iota") return iotaValue;
      return constants.get(expression.name);
    case "UnaryExpression": {
      const value = constantIntegerExpressionValue(expression.operand, constants, iotaValue);
      if (value === undefined) return undefined;
      switch (expression.operator) {
        case "+":
          return value;
        case "-":
          return -value;
        case "^":
          return ~value;
        default:
          return undefined;
      }
    }
    case "BinaryExpression": {
      const left = constantIntegerExpressionValue(expression.left, constants, iotaValue);
      const right = constantIntegerExpressionValue(expression.right, constants, iotaValue);
      if (left === undefined || right === undefined) return undefined;
      return applyIntegerConstantOperator(expression.operator, left, right);
    }
    case "SelectorExpression":
      if (expression.object.kind !== "Identifier") return undefined;
      return constants.get(`${expression.object.name}.${expression.field}`);
    default:
      return undefined;
  }
}

function applyIntegerConstantOperator(operator: string, left: bigint, right: bigint): bigint | undefined {
  switch (operator) {
    case "+":
      return left + right;
    case "-":
      return left - right;
    case "*":
      return left * right;
    case "/":
      return right === 0n ? undefined : left / right;
    case "%":
      return right === 0n ? undefined : left % right;
    case "|":
      return left | right;
    case "^":
      return left ^ right;
    case "&":
      return left & right;
    case "&^":
      return left & ~right;
    case "<<":
      return right < 0n || right > 1024n ? undefined : left << right;
    case ">>":
      return right < 0n || right > 1024n ? undefined : left >> right;
    default:
      return undefined;
  }
}

function parseIntegerLiteralText(text: string): bigint | undefined {
  const raw = text.replace(/_/g, "");
  try {
    if (/^0[xX][0-9a-fA-F]+$/.test(raw)) return BigInt(raw);
    if (/^0[bB][01]+$/.test(raw)) return BigInt(`0b${raw.slice(2)}`);
    if (/^0[oO][0-7]+$/.test(raw)) return BigInt(`0o${raw.slice(2)}`);
    if (/^0[0-7]*$/.test(raw) && raw.length > 1) return BigInt(`0o${raw.slice(1) || "0"}`);
    if (/^[0-9]+$/.test(raw)) return BigInt(raw);
  } catch {
    return undefined;
  }
  return undefined;
}

function isInterfaceTypeSpec(spec: TypeSpec): boolean {
  return Boolean(spec.interfaceMethods || spec.interfaceEmbeds || spec.type.text.trim().startsWith("interface{"));
}

function emitTypeDescriptorLines(ast: ProgramAst, artifact: GoJuniorPackageExportData, facts: PackageEmitFacts): string[] {
  const descriptors = packageTypeDescriptors(ast, artifact, facts);
  if (descriptors.length === 0) return [];
  return descriptors.map((descriptor) =>
    `__gojrTypeDescriptors[${JSON.stringify(descriptor.type)}] = ${JSON.stringify(descriptor)};`
  );
}

function packageTypeDescriptors(ast: ProgramAst, artifact: GoJuniorPackageExportData, facts: PackageEmitFacts): GeneratedTypeDescriptor[] {
  const descriptors = new Map<string, GeneratedTypeDescriptor>();
  for (const statement of ast.body) {
    if (statement.kind !== "TypeDecl") continue;
    for (const spec of statement.declarations) {
      const descriptor = typeSpecDescriptor(spec, artifact.importPath ?? "", artifact.packageName, facts);
      descriptors.set(descriptor.type, descriptor);
    }
  }
  for (const fn of ast.functions) {
    if (!fn.receiver) continue;
    const receiver = receiverBaseType(fn.receiver.type.text);
    const descriptor = descriptors.get(receiver);
    if (!descriptor) continue;
    const methods = descriptor.methods ?? [];
    methods.push({
      name: fn.name,
      receiver: fn.receiver.type.text,
      pointerReceiver: fn.receiver.type.text.trim().startsWith("*"),
      params: fn.signature.parameters.map((parameter) => parameter.variadic ? `...${parameter.type.text}` : parameter.type.text),
      results: fn.signature.results.map((result) => result.type.text)
    });
    descriptor.methods = methods;
  }
  return [...descriptors.values()];
}

function typeSpecDescriptor(spec: TypeSpec, pkgPath: string, pkgName: string | undefined, facts: PackageEmitFacts): GeneratedTypeDescriptor {
  const type = spec.name;
  const kind = typeDescriptorKind(spec);
  const descriptor: GeneratedTypeDescriptor = {
    type,
    name: type,
    string: type,
    kind,
    pkgPath,
    pkgName: pkgName ?? defaultImportName(pkgPath),
    underlying: spec.type.text
  };
  Object.assign(descriptor, compositeTypeDescriptorFields(spec.type.text, facts));
  if (spec.structFields) {
    descriptor.fields = spec.structFields.map((field) => {
      const normalizedTypeText = normalizeFixedArrayLengthTypeText(field.type.text, facts);
      const fieldPkgPath = directSelectorTypePackagePath(normalizedTypeText, facts);
      return {
        name: field.name,
        type: normalizedTypeText,
        embedded: Boolean(field.embedded),
        ...(fieldPkgPath ? { pkgPath: fieldPkgPath } : {}),
        ...(field.tag ? { tag: field.tag } : {})
      };
    });
  }
  if (spec.typeParameters && spec.typeParameters.length > 0) descriptor.typeParameters = [...spec.typeParameters];
  if (spec.interfaceMethods) {
    descriptor.interfaceMethods = spec.interfaceMethods.map((method) => ({
      name: method.name,
      params: method.signature.parameters.map((parameter) => parameter.variadic ? `...${parameter.type.text}` : parameter.type.text),
      results: method.signature.results.map((result) => result.type.text)
    }));
  }
  if (spec.interfaceEmbeds) descriptor.interfaceEmbeds = spec.interfaceEmbeds.map((embed) => embed.text);
  return descriptor;
}

function typeDescriptorKind(spec: TypeSpec): string {
  if (spec.structFields) return "struct";
  if (isInterfaceTypeSpec(spec)) return "interface";
  return typeTextKind(spec.type.text);
}

function typeTextKind(typeText: string): string {
  const trimmed = typeText.trim();
  if (trimmed.startsWith("*")) return "ptr";
  if (trimmed.startsWith("[]")) return "slice";
  if (/^\[[^\]]+\]/.test(trimmed)) return "array";
  if (trimmed.startsWith("map[")) return "map";
  if (trimmed.startsWith("chan ") || trimmed.startsWith("<-chan") || trimmed.startsWith("chan<-")) return "chan";
  if (trimmed.startsWith("func(")) return "func";
  if (trimmed === "any" || trimmed === "error" || trimmed === "interface{}") return "interface";
  if (trimmed.startsWith("interface{")) return "interface";
  switch (trimmed) {
    case "bool":
      return "bool";
    case "int":
    case "int8":
    case "int16":
    case "int32":
    case "int64":
      return trimmed;
    case "uint":
    case "uint8":
    case "uint16":
    case "uint32":
    case "uint64":
    case "uintptr":
      return trimmed;
    case "float32":
    case "float64":
      return trimmed;
    case "complex64":
    case "complex128":
      return trimmed;
    case "string":
      return "string";
    default:
      return "invalid";
  }
}

function compositeTypeDescriptorFields(typeText: string, facts: PackageEmitFacts): Partial<GeneratedTypeDescriptor> {
  const trimmed = typeText.trim();
  const mapType = parseMapTypeText(trimmed);
  if (mapType) return { key: mapType.keyType, elem: mapType.valueType };
  const arrayType = parseArrayOrSliceTypeText(trimmed);
  if (arrayType) {
    const length = parseArrayLengthTypeText(trimmed, facts);
    return {
      elem: arrayType.elementType,
      ...(length === undefined ? {} : { len: length })
    };
  }
  const chanElem = chanElementTypeText(trimmed);
  if (chanElem) return { elem: chanElem };
  const fn = parseFuncTypeText(trimmed);
  if (fn) return { params: fn.params, results: fn.results };
  return {};
}

function parseArrayLengthTypeText(typeText: string | undefined, facts: PackageEmitFacts = emptyPackageEmitFacts): number | undefined {
  if (!typeText?.startsWith("[") || typeText.startsWith("[]")) return undefined;
  const close = typeText.indexOf("]");
  if (close < 0) return undefined;
  const raw = typeText.slice(1, close).trim();
  const constant = evaluateArrayLengthConstantText(raw, facts);
  if (constant === undefined || constant < 0n || constant > BigInt(Number.MAX_SAFE_INTEGER)) return undefined;
  const length = Number(constant);
  return Number.isSafeInteger(length) ? length : undefined;
}

function normalizeFixedArrayLengthTypeText(typeText: string, facts: PackageEmitFacts): string {
  const trimmed = typeText.trim();
  if (!trimmed.startsWith("[") || trimmed.startsWith("[]")) return typeText;
  const close = trimmed.indexOf("]");
  if (close < 0) return typeText;
  const length = parseArrayLengthTypeText(trimmed, facts);
  if (length === undefined) return typeText;
  return `[${length}]${trimmed.slice(close + 1).trim()}`;
}

function evaluateArrayLengthConstantText(text: string, facts: PackageEmitFacts): bigint | undefined {
  const parser = new EmitterArrayLengthConstantParser(text, facts);
  try {
    return parser.parse();
  } catch {
    return undefined;
  }
}

class EmitterArrayLengthConstantParser {
  private offset = 0;

  public constructor(
    private readonly text: string,
    private readonly facts: PackageEmitFacts
  ) {}

  public parse(): bigint | undefined {
    const value = this.expression(1);
    this.skipSpace();
    return this.offset === this.text.length ? value : undefined;
  }

  private expression(minPrecedence: number): bigint | undefined {
    let left = this.unary();
    if (left === undefined) return undefined;
    while (true) {
      this.skipSpace();
      const operator = this.peekBinaryOperator();
      if (!operator || operator.precedence < minPrecedence) return left;
      this.offset += operator.text.length;
      const right = this.expression(operator.precedence + 1);
      if (right === undefined) return undefined;
      left = applyIntegerConstantOperator(operator.text, left, right);
      if (left === undefined) return undefined;
    }
  }

  private unary(): bigint | undefined {
    this.skipSpace();
    if (this.consume("+")) return this.unary();
    if (this.consume("-")) {
      const value = this.unary();
      return value === undefined ? undefined : -value;
    }
    if (this.consume("^")) {
      const value = this.unary();
      return value === undefined ? undefined : ~value;
    }
    return this.primary();
  }

  private primary(): bigint | undefined {
    this.skipSpace();
    if (this.consume("(")) {
      const value = this.expression(1);
      this.skipSpace();
      return this.consume(")") ? value : undefined;
    }
    const integer = this.consumeIntegerLiteral();
    if (integer !== undefined) return integer;
    const identifier = this.consumeIdentifier();
    return identifier ? this.facts.constValues.get(identifier) : undefined;
  }

  private consumeIntegerLiteral(): bigint | undefined {
    const start = this.offset;
    if (this.text.startsWith("0x", start) || this.text.startsWith("0X", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[0-9a-fA-F_]/.test(this.text[this.offset]!)) this.offset += 1;
      return parseIntegerLiteralText(this.text.slice(start, this.offset));
    }
    if (this.text.startsWith("0b", start) || this.text.startsWith("0B", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[01_]/.test(this.text[this.offset]!)) this.offset += 1;
      return parseIntegerLiteralText(this.text.slice(start, this.offset));
    }
    if (this.text.startsWith("0o", start) || this.text.startsWith("0O", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[0-7_]/.test(this.text[this.offset]!)) this.offset += 1;
      return parseIntegerLiteralText(this.text.slice(start, this.offset));
    }
    while (this.offset < this.text.length && /[0-9_]/.test(this.text[this.offset]!)) this.offset += 1;
    return this.offset === start ? undefined : parseIntegerLiteralText(this.text.slice(start, this.offset));
  }

  private consumeIdentifier(): string | undefined {
    this.skipSpace();
    const first = this.consumeIdentifierSegment();
    if (!first) return undefined;
    let name = first;
    while (this.text[this.offset] === ".") {
      const dot = this.offset;
      this.offset += 1;
      const next = this.consumeIdentifierSegment();
      if (!next) {
        this.offset = dot;
        break;
      }
      name += `.${next}`;
    }
    return name;
  }

  private consumeIdentifierSegment(): string | undefined {
    const start = this.offset;
    if (start >= this.text.length || !/[A-Za-z_]/.test(this.text[start]!)) return undefined;
    this.offset += 1;
    while (this.offset < this.text.length && /[A-Za-z0-9_]/.test(this.text[this.offset]!)) this.offset += 1;
    return this.text.slice(start, this.offset);
  }

  private peekBinaryOperator(): { text: string; precedence: number } | undefined {
    const operators = [
      ["&^", 5],
      ["<<", 5], [">>", 5],
      ["*", 5], ["/", 5], ["%", 5], ["&", 5],
      ["+", 4], ["-", 4], ["|", 4], ["^", 4]
    ] as const;
    const operator = operators.find(([text]) => this.text.startsWith(text, this.offset));
    return operator ? { text: operator[0], precedence: operator[1] } : undefined;
  }

  private consume(text: string): boolean {
    this.skipSpace();
    if (!this.text.startsWith(text, this.offset)) return false;
    this.offset += text.length;
    return true;
  }

  private skipSpace(): void {
    while (this.offset < this.text.length && /\s/.test(this.text[this.offset]!)) this.offset += 1;
  }
}

function parseFuncTypeText(typeText: string | undefined): { params: string[]; results: string[] } | undefined {
  const trimmed = typeText?.trim();
  if (!trimmed?.startsWith("func(")) return undefined;
  const paramsEnd = matchingParenIndex(trimmed, 4);
  if (paramsEnd < 0) return undefined;
  const params = splitTopLevelTypes(trimmed.slice(5, paramsEnd));
  const resultText = trimmed.slice(paramsEnd + 1).trim();
  if (!resultText) return { params, results: [] };
  if (resultText.startsWith("(")) {
    const resultsEnd = matchingParenIndex(resultText, 0);
    if (resultsEnd !== resultText.length - 1) return undefined;
    return { params, results: splitTopLevelTypes(resultText.slice(1, resultsEnd)) };
  }
  return { params, results: [resultText] };
}

function matchingParenIndex(text: string, openIndex: number): number {
  let depth = 0;
  for (let index = openIndex; index < text.length; index += 1) {
    const char = text[index];
    if (char === "(") depth += 1;
    if (char === ")") {
      depth -= 1;
      if (depth === 0) return index;
    }
  }
  return -1;
}

function splitTopLevelTypes(text: string): string[] {
  const out: string[] = [];
  let start = 0;
  let parenDepth = 0;
  let bracketDepth = 0;
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index];
    if (char === "(") parenDepth += 1;
    else if (char === ")") parenDepth -= 1;
    else if (char === "[") bracketDepth += 1;
    else if (char === "]") bracketDepth -= 1;
    else if (char === "," && parenDepth === 0 && bracketDepth === 0) {
      const item = text.slice(start, index).trim();
      if (item) out.push(item);
      start = index + 1;
    }
  }
  const last = text.slice(start).trim();
  if (last) out.push(last);
  return out;
}

function directSelectorTypePackagePath(typeText: string, facts: PackageEmitFacts): string | undefined {
  let text = typeText.trim();
  while (text.startsWith("*")) text = text.slice(1).trim();
  const match = /^([A-Za-z_][A-Za-z0-9_]*)\.[A-Za-z_][A-Za-z0-9_]*(?:\[.*\])?$/.exec(text);
  if (!match) return undefined;
  return facts.imports.get(match[1] ?? "");
}

function defaultImportName(path: string): string {
  const parts = path.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

function functionPackageKey(fn: FunctionDecl): string {
  return fn.receiver ? `${receiverBaseType(fn.receiver.type.text)}.${fn.name}` : fn.name;
}

function receiverBaseType(typeText: string): string {
  const trimmed = typeText.trim();
  const dereferenced = trimmed.startsWith("*") ? trimmed.slice(1).trim() : trimmed;
  return genericBaseTypeText(dereferenced);
}

function dereferencedTypeText(typeText: string | undefined): string | undefined {
  const trimmed = typeText?.trim();
  return trimmed?.startsWith("*") ? trimmed.slice(1).trim() : undefined;
}

function genericBaseTypeText(typeText: string): string {
  const trimmed = typeText.trim();
  if (trimmed.startsWith("[") || trimmed.startsWith("map[")) return trimmed;
  const open = topLevelGenericOpenBracket(trimmed);
  if (open === undefined) return trimmed;
  const close = matchingTypeBracket(trimmed, open);
  return close === trimmed.length - 1 ? trimmed.slice(0, open).trim() : trimmed;
}

function topLevelGenericOpenBracket(typeText: string): number | undefined {
  let depth = 0;
  for (let index = 0; index < typeText.length; index += 1) {
    const char = typeText[index];
    if (char === "[") {
      if (depth === 0) return index;
      depth += 1;
      continue;
    }
    if (char === "]") depth = Math.max(0, depth - 1);
  }
  return undefined;
}

function matchingTypeBracket(typeText: string, open: number): number | undefined {
  let depth = 0;
  for (let index = open; index < typeText.length; index += 1) {
    const char = typeText[index];
    if (char === "[") depth += 1;
    if (char === "]") {
      depth -= 1;
      if (depth === 0) return index;
    }
  }
  return undefined;
}

function expressionTypeText(expression: Expression | undefined, env: ExpressionEmitEnv): string | undefined {
  if (!expression) return undefined;
  if (expression.typeText) return expression.typeText;
  if (expression.kind === "Literal") {
    switch (expression.literalKind) {
      case "int":
      case "rune":
        return "int";
      case "float":
        return "float64";
      case "imag":
        return "complex128";
      case "string":
        return "string";
      case "bool":
        return "bool";
      default:
        return undefined;
    }
  }
  if (expression.kind === "Identifier") return env.localTypes.get(expression.name) ?? env.facts.packageTypes.get(expression.name);
  if (expression.kind === "IndexExpression") return mapValueTypeText(expressionTypeText(expression.object, env), env.facts);
  if (expression.kind === "MapLiteralExpression") return `map[${expression.keyType.text}]${expression.valueType.text}`;
  if (expression.kind === "ArrayLiteralExpression") return expression.type.text;
  if (expression.kind === "StructLiteralExpression") return expression.typeName;
  if (expression.kind === "FunctionLiteralExpression") return expression.typeText;
  if (expression.kind === "UnaryExpression") {
    if (expression.operator === "&") {
      const operandType = expressionTypeText(expression.operand, env);
      return operandType ? `*${operandType}` : undefined;
    }
    if (expression.operator === "*") {
      return dereferencedTypeText(expressionTypeText(expression.operand, env));
    }
  }
  if (expression.kind === "CallExpression" &&
    expression.callee.kind === "Identifier" &&
    expression.callee.name === "new" &&
    expression.args[0]) {
    const typeText = typeArgumentExpressionText(expression.args[0]);
    return typeText ? `*${typeText}` : undefined;
  }
  if (expression.kind === "CallExpression" &&
    expression.callee.kind === "Identifier" &&
    expression.callee.name === "make" &&
    expression.args[0]?.kind === "TypeExpression") {
    return expression.args[0].type.text;
  }
  if (expression.kind === "CallExpression" && expression.callee.kind === "TypeExpression") return expression.callee.type.text;
  if (expression.kind === "CallExpression" &&
    expression.callee.kind === "Identifier" &&
    primitiveTypeNames.has(expression.callee.name)) {
    return expression.callee.name;
  }
  if (expression.kind === "CallExpression" &&
    expression.callee.kind === "Identifier" &&
    !env.locals.has(expression.callee.name) &&
    isKnownConversionTypeText(expression.callee.name, env)) {
    return expression.callee.name;
  }
  return undefined;
}

function packageExportTypeText(ctx: EmitterContext, expression: Expression | undefined): string | undefined {
  if (expression?.kind !== "Identifier") return undefined;
  return ctx.options.artifact.exportIndex[expression.name]?.typeText ??
    ctx.options.artifact.exports.find((item) => item.name === expression.name)?.typeText;
}

function declarationTypeText(declaration: DeclarationSpec, effectiveValue = declaration.value): string | undefined {
  if (declaration.type?.text) return declaration.type.text;
  if (effectiveValue) return expressionTypeText(effectiveValue, emptyExpressionEnv);
  return undefined;
}

function tupleSourceTypeTexts(expression: Expression, targetCount: number, env: ExpressionEmitEnv): Array<string | undefined> {
  if (targetCount === 2 && expression.kind === "IndexExpression") {
    return [mapValueTypeText(expressionTypeText(expression.object, env)), "bool"];
  }
  if (targetCount === 2 && expression.kind === "TypeAssertionExpression") {
    return [expression.type.text, "bool"];
  }
  if (targetCount === 2 && expression.kind === "UnaryExpression" && expression.operator === "<-") {
    return [chanElementTypeText(expressionTypeText(expression.operand, env)), "bool"];
  }
  if (targetCount === 1) return [expressionTypeText(expression, env)];
  return [];
}

function valueForTargetType(value: string, targetType: string | undefined, sourceType: string | undefined, env: ExpressionEmitEnv): string {
  const staticTargetType = targetType?.trim();
  if (staticTargetType && env.typeParameters?.has(staticTargetType)) {
    return `__gojrConvertDynamicType(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(staticTargetType)}), ${value})`;
  }
  const resolvedTargetType = targetType ? resolveUnderlyingTypeTextInEnv(targetType, env) : undefined;
  if (resolvedTargetType && isIntegerType(resolvedTargetType) && !(resolvedTargetType === "uintptr" && isPointerLikeTypeText(sourceType))) {
    return `__gojrIntegerConvert(${JSON.stringify(resolvedTargetType)}, ${value})`;
  }
  if (resolvedTargetType && isFloatType(resolvedTargetType)) return `Number(${value})`;
  if (!isInterfaceTypeText(targetType, env.facts)) return value;
  return `__gojrToInterface(${value}, ${JSON.stringify(targetType)}, ${JSON.stringify(sourceType)})`;
}

function isInterfaceTypeText(typeText: string | undefined, facts: PackageEmitFacts): boolean {
  if (!typeText) return false;
  const normalized = typeText.trim();
  if (normalized.startsWith("*")) return false;
  if (normalized === "any" || normalized === "interface{}" || normalized === "error") return true;
  if (normalized.startsWith("interface{")) return true;
  return facts.interfaceTypes.has(normalized) || facts.interfaceTypes.has(receiverBaseType(normalized));
}

function rangeIterationTypeTexts(expression: Expression | undefined, env: ExpressionEmitEnv): { key?: string; value?: string } {
  const typeText = expressionTypeText(expression, env);
  if (!typeText) return {};
  const resolved = resolveUnderlyingTypeTextInEnv(typeText, env);
  if (resolved === "string") return { key: "int", value: "rune" };
  const mapType = parseMapTypeText(resolved);
  if (mapType) return { key: mapType.keyType, value: mapType.valueType };
  const sliceType = parseArrayOrSliceTypeText(resolved);
  if (sliceType) return { key: "int", value: sliceType.elementType };
  if (isIntegerType(resolved)) return { key: resolved, value: resolved };
  return {};
}

function expressionToJs(ctx: EmitterContext, expression: Expression | undefined, env: ExpressionEmitEnv = emptyExpressionEnv): string | undefined {
  if (!expression) return undefined;
  const expressionEnv = expression.typeText && isFloatType(expression.typeText) && env.numericLiteralKind !== "float"
    ? { ...env, numericLiteralKind: "float" as const }
    : env;
  switch (expression.kind) {
    case "Identifier":
      return identifierToJs(expression, expressionEnv);
    case "Literal":
      return literalToJs(expression, expressionEnv);
    case "ArrayLiteralExpression":
      return arrayLiteralToJs(ctx, expression, expressionEnv);
    case "StructLiteralExpression":
      return structLiteralToJs(ctx, expression, expressionEnv);
    case "MapLiteralExpression":
      return mapLiteralToJs(ctx, expression, expressionEnv);
    case "FunctionLiteralExpression":
      return functionLiteralToJs(ctx, expression, expressionEnv);
    case "UnaryExpression":
      return unaryExpressionToJs(ctx, expression, expressionEnv);
    case "BinaryExpression":
      return binaryExpressionToJs(ctx, expression, expressionEnv);
    case "SelectorExpression":
      return selectorExpressionToJs(ctx, expression, expressionEnv);
    case "CallExpression":
      return callExpressionToJs(ctx, expression, expressionEnv);
    case "IndexExpression":
      return indexExpressionToJs(ctx, expression, expressionEnv);
    case "SliceExpression":
      return sliceExpressionToJs(ctx, expression, expressionEnv);
    case "TypeAssertionExpression":
      return typeAssertionExpressionToJs(ctx, expression, expressionEnv);
    default:
      ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
      return undefined;
  }
}

function identifierToJs(expression: IdentifierExpression, env: ExpressionEmitEnv): string {
  if (expression.name === "iota" && env.iotaValue !== undefined) return `${env.iotaValue}n`;
  const local = env.locals.get(expression.name);
  if (local) return local;
  const importPath = env.facts.imports.get(expression.name);
  if (importPath) return `__gojrImport(${JSON.stringify(importPath)})`;
  return `pkg[${JSON.stringify(expression.name)}]`;
}

function literalToJs(expression: LiteralExpression, env: ExpressionEmitEnv = emptyExpressionEnv): string | undefined {
  switch (expression.literalKind) {
    case "int":
    case "rune":
      if (env.numericLiteralKind === "float") {
        if (typeof expression.value === "bigint") return expression.value.toString();
        if (typeof expression.value === "number") return String(Math.trunc(expression.value));
        return undefined;
      }
      if (typeof expression.value === "bigint") return `${expression.value.toString()}n`;
      if (typeof expression.value === "number") return `${Math.trunc(expression.value)}n`;
      return undefined;
    case "float":
      return typeof expression.value === "number" ? JSON.stringify(expression.value) : undefined;
    case "imag":
      return isComplexLiteralValue(expression.value) ? `__gojrComplex(${JSON.stringify(expression.value.real)}, ${JSON.stringify(expression.value.imag)})` : undefined;
    case "string":
      if (typeof expression.value !== "string") return undefined;
      {
        const literalBytes = goStringLiteralBytes(expression.raw);
        if (literalBytes && !byteArraysEqual(literalBytes, utf8Bytes(expression.value))) {
          return `__gojrRawStringBase64(${JSON.stringify(bytesToBase64(literalBytes))})`;
        }
      }
      return JSON.stringify(expression.value);
    case "bool":
      return typeof expression.value === "boolean" ? String(expression.value) : undefined;
    case "nil":
      return "null";
    default:
      return undefined;
  }
}

function arrayLiteralToJs(ctx: EmitterContext, expression: ArrayLiteralExpression, env: ExpressionEmitEnv): string | undefined {
  const arrayType = parseArrayOrSliceTypeText(expression.type.text);
  const fixedArrayLiteral = arrayType && /^\[\d+\]/.test(expression.type.text.trim());
  if (fixedArrayLiteral || expression.elements.some((element) => element.key !== undefined)) {
    if (!arrayType) {
      ctx.emitError(`unsupported Stage 3 keyed array literal ${expression.type.text}`);
      return undefined;
    }
    const entries: string[] = [];
    for (const element of expression.elements) {
      const key = element.key ? expressionToJs(ctx, element.key, env) : "null";
      const value = expressionToJs(ctx, element.value, env);
      if (!key || !value) return undefined;
      entries.push(`[${key}, ${valueForTargetType(value, arrayType.elementType, expressionTypeText(element.value, env), env)}]`);
    }
    return `__gojrArrayLiteral(${JSON.stringify(expression.type.text)}, ${JSON.stringify(arrayType.elementType)}, [${entries.join(", ")}])`;
  }
  if (isByteSliceType(expression.type.text)) {
    const bytes = bytesFromArrayLiteral(expression);
    if (bytes) {
      if (bytes.length > 64) return `__gojrBytesBase64(${JSON.stringify(bytesToBase64(bytes))})`;
      return `new Uint8Array([${bytes.join(", ")}])`;
    }
    const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
    if (values.some((value) => value === undefined)) return undefined;
    return `Uint8Array.from([${values.filter((value): value is string => value !== undefined).join(", ")}], (item) => Number(item) & 255)`;
  }
  const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
  if (values.some((value) => value === undefined)) return undefined;
  const rendered = values
    .map((value, index) => value === undefined ? undefined : valueForTargetType(value, arrayType?.elementType, expressionTypeText(expression.elements[index]?.value, env), env))
    .filter((value): value is string => value !== undefined);
  return `[${rendered.join(", ")}]`;
}

function structLiteralToJs(ctx: EmitterContext, expression: StructLiteralExpression, env: ExpressionEmitEnv): string | undefined {
  const underlying = resolveUnderlyingTypeTextInEnv(expression.typeName, env);
  const arrayType = parseArrayOrSliceTypeText(underlying);
  if (arrayType) {
    const entries: string[] = [];
    for (const field of expression.fields) {
      if (field.name) {
        ctx.emitError(`unsupported Stage 3 named array or slice literal field ${field.name}`);
        return undefined;
      }
      const key = field.key ? expressionToJs(ctx, field.key, env) : "null";
      const value = expressionToJs(ctx, field.value, env);
      if (!key || !value) return undefined;
      entries.push(`[${key}, ${valueForTargetType(value, arrayType.elementType, expressionTypeText(field.value, env), env)}]`);
    }
    return `__gojrArrayLiteral(${JSON.stringify(underlying)}, ${JSON.stringify(arrayType.elementType)}, [${entries.join(", ")}])`;
  }
  if (expression.fields.some((field) => !field.name)) {
    const values: string[] = [];
    for (const field of expression.fields) {
      if (field.name) {
        ctx.emitError(`unsupported Stage 3 mixed keyed and unkeyed struct literal ${expression.typeName}`);
        return undefined;
      }
      const value = expressionToJs(ctx, field.value, env);
      if (!value) return undefined;
      values.push(value);
    }
    return `__gojrStructFromValues(${JSON.stringify(expression.typeName)}, [${values.join(", ")}])`;
  }
  const fields: string[] = [];
  for (const field of expression.fields) {
    const value = expressionToJs(ctx, field.value, env);
    if (!value) return undefined;
    fields.push(`${JSON.stringify(field.name)}: ${value}`);
  }
  const pkgPath = packagePathForTypeText(expression.typeName, env);
  return `__gojrStruct(${JSON.stringify(expression.typeName)}, { ${fields.join(", ")} }${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
}

function mapLiteralToJs(ctx: EmitterContext, expression: MapLiteralExpression, env: ExpressionEmitEnv): string | undefined {
  const mapStringBytes = mapStringBytesLiteralEntries(expression);
  if (mapStringBytes) {
    return `__gojrMapStringBytesBase64([${mapStringBytes.map(([key, bytes]) => `[${JSON.stringify(key)}, ${JSON.stringify(bytesToBase64(bytes))}]`).join(", ")}])`;
  }
  const entries: string[] = [];
  for (const entry of expression.entries) {
    const key = expressionToJs(ctx, entry.key, env);
    const value = expressionToJs(ctx, entry.value, env);
    if (!key || !value) return undefined;
    entries.push(`[${valueForTargetType(key, expression.keyType.text, expressionTypeText(entry.key, env), env)}, ${valueForTargetType(value, expression.valueType.text, expressionTypeText(entry.value, env), env)}]`);
  }
  return `__gojrMap(new Map([${entries.join(", ")}]), ${zeroValueForType(expression.valueType.text, env.facts) ?? "null"}, ${JSON.stringify(expression.keyType.text)}, ${JSON.stringify(expression.valueType.text)})`;
}

function functionLiteralToJs(ctx: EmitterContext, expression: FunctionLiteralExpression, env: ExpressionEmitEnv): string | undefined {
  const literalEnv = cloneExpressionEnv(env);
  literalEnv.expectedReturnTypes = expression.signature.results.map((result) => result.type.text);
  const params: string[] = [];
  for (const [index, parameter] of expression.signature.parameters.entries()) {
    const goName = parameter.name && parameter.name !== "_" ? parameter.name : `__gojrArg${index}`;
    const jsName = safeLocalName(goName, index);
    params.push(parameter.variadic ? `...${jsName}` : jsName);
    if (parameter.name && parameter.name !== "_") {
      literalEnv.locals.set(parameter.name, jsName);
      literalEnv.localTypes.set(parameter.name, parameter.variadic ? `[]${parameter.type.text}` : parameter.type.text);
    }
  }
  const hasDefer = statementsContainDefer(expression.body.statements);
  if (hasDefer) literalEnv.deferName = ctx.symbol("defer");
  const namedResults: string[] = [];
  const resultLines: string[] = [];
  for (const [index, result] of expression.signature.results.entries()) {
    if (!result.name || result.name === "_") continue;
    const name = safeLocalName(result.name, index);
    literalEnv.locals.set(result.name, name);
    literalEnv.localTypes.set(result.name, result.type.text);
    namedResults.push(name);
    resultLines.push(`  let ${name} = ${zeroValueForTypeInEnv(result.type.text, literalEnv) ?? "null"};`);
  }
  if (namedResults.length > 0) literalEnv.namedResults = namedResults;
  const parameterCoercionLines = functionLiteralParameterCoercionLines(expression, literalEnv, "  ");
  const statementLines = emitStatements(ctx, expression.body.statements, literalEnv, "  ");
  if (!statementLines) return undefined;
  const defaultReturn = functionLiteralDefaultReturn(expression, literalEnv);
  const body = [
    ...parameterCoercionLines,
    ...resultLines,
    ...statementLines,
    ...(functionBodyAlwaysReturns(expression.body) ? [] : [`  return ${defaultReturn};`])
  ].join("\n");
  if (hasDefer) {
    return `(async (${params.join(", ")}) => __gojrCallScope(async () => {\n${parameterCoercionLines.join("\n")}${parameterCoercionLines.length > 0 ? "\n" : ""}${resultLines.map((line) => `  ${line}`).join("\n")}${resultLines.length > 0 ? "\n" : ""}  return await __gojrDeferScope(async (${literalEnv.deferName}) => {\n${statementLines.join("\n")}\n${functionBodyAlwaysReturns(expression.body) ? "" : `  return ${defaultReturn};\n`}  }, () => ${defaultReturn});\n}))`;
  }
  return `(async (${params.join(", ")}) => __gojrCallScope(async () => {\n${body}\n}))`;
}

function functionLiteralParameterCoercionLines(expression: FunctionLiteralExpression, env: ExpressionEmitEnv, indent: string): string[] {
  const lines: string[] = [];
  for (const [index, parameter] of expression.signature.parameters.entries()) {
    const goName = parameter.name && parameter.name !== "_" ? parameter.name : `__gojrArg${index}`;
    const jsName = safeLocalName(goName, index);
    if (parameter.variadic) {
      lines.push(`${indent}${jsName} = Array.from(${jsName}, (__gojrItem) => ${valueForTargetType("__gojrItem", parameter.type.text, parameter.type.text, env)});`);
    } else {
      lines.push(`${indent}${jsName} = ${valueForTargetType(jsName, parameter.type.text, parameter.type.text, env)};`);
    }
  }
  return lines;
}

function functionLiteralDefaultReturn(expression: FunctionLiteralExpression, env: ExpressionEmitEnv): string {
  const named = namedResultReturn(env);
  if (named) return named;
  if (expression.signature.results.length === 0) return "null";
  if (expression.signature.results.length === 1) return zeroValueForTypeInEnv(expression.signature.results[0]?.type.text, env) ?? "null";
  return `__gojrTuple([${expression.signature.results.map((result) => zeroValueForTypeInEnv(result.type.text, env) ?? "null").join(", ")}])`;
}

function unaryExpressionToJs(ctx: EmitterContext, expression: UnaryExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.operator === "&") return addressOfExpressionToJs(ctx, expression.operand, env);
  const operand = expressionToJs(ctx, expression.operand, env);
  if (!operand) return undefined;
  switch (expression.operator) {
    case "+":
      return `__gojrNumericIdentity(${operand})`;
    case "-":
      return `__gojrNegate(${operand}, ${JSON.stringify(expression.typeText ?? expressionTypeText(expression.operand, env) ?? "")})`;
    case "!":
      return `(!__gojrBool(${operand}))`;
    case "^":
      return `__gojrBitwiseNot(${operand}, ${JSON.stringify(expression.typeText ?? expressionTypeText(expression.operand, env) ?? "")})`;
    case "*":
      return `__gojrDeref(${operand})`;
    case "<-":
      return `(await __gojrChanRecv(${operand}))[0]`;
    default:
      ctx.emitError(`unsupported Stage 3 unary operator ${expression.operator}`);
      return undefined;
  }
}

function addressOfExpressionToJs(ctx: EmitterContext, expression: Expression, env: ExpressionEmitEnv): string | undefined {
  const typeText = expressionTypeText(expression, env) ?? "any";
  switch (expression.kind) {
    case "Identifier": {
      const target = identifierToJs(expression, env);
      const targetType = env.localTypes.get(expression.name) ?? env.facts.packageTypes.get(expression.name);
      const pkgPath = packagePathForTypeText(typeText, env);
      const identity = env.locals.has(expression.name) ? undefined : `pkg:${ctx.options.artifact.importPath}:${expression.name}`;
      return `__gojrPointer(${JSON.stringify(typeText)}, () => ${target}, (next) => { ${target} = ${valueForTargetType("next", targetType, undefined, env)}; }, ${pkgPath ? JSON.stringify(pkgPath) : "undefined"}, undefined, undefined${identity ? `, ${JSON.stringify(identity)}` : ""})`;
    }
    case "SelectorExpression": {
      const object = expressionToJs(ctx, expression.object, env);
      if (!object) return undefined;
      const pkgPath = packagePathForTypeText(typeText, env);
      return `__gojrAddressField(${object}, ${JSON.stringify(expression.field)}, ${JSON.stringify(typeText)}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
    }
    case "IndexExpression": {
      const objectType = expressionTypeText(expression.object, env) ?? "";
      if (parseMapTypeText(resolveUnderlyingTypeTextInEnv(objectType, env))) {
        ctx.emitError("unsupported Stage 4 address-of map index");
        return undefined;
      }
      const object = expressionToJs(ctx, expression.object, env);
      const index = expressionToJs(ctx, expression.index, env);
      if (!object || !index) return undefined;
      const pkgPath = packagePathForTypeText(typeText, env);
      return `__gojrAddressIndex(${object}, ${index}, ${JSON.stringify(typeText)}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
    }
    case "StructLiteralExpression":
    case "ArrayLiteralExpression":
    case "MapLiteralExpression": {
      const value = expressionToJs(ctx, expression, env);
      if (!value) return undefined;
      const pkgPath = packagePathForTypeText(typeText, env);
      return `__gojrPointerValue(${JSON.stringify(typeText)}, ${value}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
    }
    default:
      ctx.emitError(`unsupported Stage 4 address-of ${expression.kind}`);
      return undefined;
  }
}

function binaryExpressionToJs(ctx: EmitterContext, expression: BinaryExpression, env: ExpressionEmitEnv): string | undefined {
  const left = expressionToJs(ctx, expression.left, env);
  const right = expressionToJs(ctx, expression.right, env);
  if (!left || !right) return undefined;
  if (isComplexType(expression.typeText) || isComplexType(expression.left.typeText) || isComplexType(expression.right.typeText)) {
    return complexBinaryExpressionToJs(ctx, expression, left, right);
  }
  const rendered = binaryOperationToJs(expression.operator, left, right, expression.typeText ?? expressionTypeText(expression.left, env));
  if (!rendered) {
    ctx.emitError(`unsupported Stage 3 binary operator ${expression.operator}`);
    return undefined;
  }
  return rendered;
}

function binaryOperationToJs(
  operator: BinaryExpression["operator"],
  left: string,
  right: string,
  resultType?: string
): string | undefined {
  const type = JSON.stringify(resultType ?? "");
  switch (operator) {
    case "&&":
      return `(__gojrBool(${left}) && __gojrBool(${right}))`;
    case "||":
      return `(__gojrBool(${left}) || __gojrBool(${right}))`;
    case "==":
      return `__gojrEqual(${left}, ${right})`;
    case "!=":
      return `(!__gojrEqual(${left}, ${right}))`;
    case "<":
    case "<=":
    case ">":
    case ">=":
      return `__gojrCompareOp(${left}, ${right}, ${JSON.stringify(operator)})`;
    case "+":
      return `__gojrAdd(${left}, ${right}, ${type})`;
    case "-":
      return `__gojrSub(${left}, ${right}, ${type})`;
    case "*":
      return `__gojrMul(${left}, ${right}, ${type})`;
    case "/":
      return `__gojrDiv(${left}, ${right}, ${type})`;
    case "%":
      return `__gojrMod(${left}, ${right}, ${type})`;
    case "&":
      return `__gojrBitwiseAnd(${left}, ${right}, ${type})`;
    case "|":
      return `__gojrBitwiseOr(${left}, ${right}, ${type})`;
    case "^":
      return `__gojrBitwiseXor(${left}, ${right}, ${type})`;
    case "&^":
      return `__gojrBitClear(${left}, ${right}, ${type})`;
    case "<<":
      return `__gojrShiftLeft(${left}, ${right}, ${type})`;
    case ">>":
      return `__gojrShiftRight(${left}, ${right}, ${type})`;
    default:
      return undefined;
  }
}

function complexBinaryExpressionToJs(ctx: EmitterContext, expression: BinaryExpression, left: string, right: string): string | undefined {
  switch (expression.operator) {
    case "+":
      return `__gojrComplexAdd(${left}, ${right})`;
    case "-":
      return `__gojrComplexSub(${left}, ${right})`;
    case "*":
      return `__gojrComplexMul(${left}, ${right})`;
    case "/":
      return `__gojrComplexDiv(${left}, ${right})`;
    case "==":
      return `__gojrComplexEq(${left}, ${right})`;
    case "!=":
      return `(!__gojrComplexEq(${left}, ${right}))`;
    default:
      ctx.emitError(`unsupported Stage 3 complex operator ${expression.operator}`);
      return undefined;
  }
}

function selectorExpressionToJs(ctx: EmitterContext, expression: SelectorExpression, env: ExpressionEmitEnv): string | undefined {
  if (isImportedPackageSelector(expression, env) && expression.object.kind === "Identifier") {
    const importPath = env.facts.imports.get(expression.object.name);
    return `__gojrPackageSelector(${JSON.stringify(importPath)}, ${JSON.stringify(expression.field)})`;
  }
  const methodKey = methodPackageKeyForSelector(expression, env);
  if (methodKey) {
    const receiver = methodReceiverToJs(ctx, expression, methodKey, env);
    if (!receiver) return undefined;
    const typeArgs = methodReceiverTypeArgObject(expression, methodKey, env, receiver);
    return `(async (...__gojrMethodArgs) => await pkg[${JSON.stringify(methodKey)}](${[...(typeArgs ? [typeArgs] : []), receiver, "...__gojrMethodArgs"].join(", ")}))`;
  }
  const object = expressionToJs(ctx, expression.object, env);
  if (!object) return undefined;
  return `__gojrGetField(${object}, ${JSON.stringify(expression.field)})`;
}

function methodPackageKeyForSelector(expression: SelectorExpression, env: ExpressionEmitEnv): string | undefined {
  const receiverType = expressionTypeText(expression.object, env);
  if (!receiverType) return undefined;
  return env.facts.methodKeys.get(`${receiverBaseType(receiverType)}.${expression.field}`);
}

function selectorFieldTypeText(expression: SelectorExpression, env: ExpressionEmitEnv): string | undefined {
  const objectType = expressionTypeText(expression.object, env);
  return structFieldTypeText(objectType, expression.field, env.facts);
}

function structFieldTypeText(
  typeText: string | undefined,
  fieldName: string,
  facts: PackageEmitFacts,
  seen = new Set<string>()
): string | undefined {
  if (!typeText) return undefined;
  const base = receiverBaseType(resolveUnderlyingTypeText(typeText, facts));
  if (!base || seen.has(base)) return undefined;
  seen.add(base);
  const fields = facts.structFields.get(base);
  if (!fields) return undefined;
  const direct = fields.find((field) => field.name === fieldName);
  if (direct) return direct.type;
  const promoted = fields
    .filter((field) => field.embedded)
    .map((field) => structFieldTypeText(field.type, fieldName, facts, seen))
    .filter((fieldType): fieldType is string => fieldType !== undefined);
  return promoted.length === 1 ? promoted[0] : undefined;
}

function methodReceiverToJs(ctx: EmitterContext, expression: SelectorExpression, methodKey: string, env: ExpressionEmitEnv): string | undefined {
  const receiverType = expressionTypeText(expression.object, env);
  if (env.facts.methodPointerReceivers.get(methodKey) && !receiverType?.trim().startsWith("*")) {
    const addressed = addressOfExpressionToJs(ctx, expression.object, env);
    if (addressed) return addressed;
  }
  return expressionToJs(ctx, expression.object, env);
}

function methodReceiverTypeArgObject(expression: SelectorExpression, methodKey: string, env: ExpressionEmitEnv, receiverJs: string): string | undefined {
  const parameterNames = env.facts.functionTypeParameters.get(methodKey);
  if (!parameterNames || parameterNames.length === 0) return undefined;
  const receiverType = expressionTypeText(expression.object, env);
  const typeArgs = receiverType ? genericTypeArgumentsFromTypeText(receiverType) : undefined;
  if (typeArgs && typeArgs.length > 0) {
    const fields = parameterNames.map((name, index) =>
      `${JSON.stringify(name)}: ${JSON.stringify(typeArgs[index] ?? "any")}`
    );
    return `({ ${fields.join(", ")} })`;
  }
  return `__gojrTypeArgsFromReceiver(${receiverJs}, ${JSON.stringify(parameterNames)})`;
}

function genericTypeArgumentsFromTypeText(typeText: string): string[] | undefined {
  let trimmed = typeText.trim();
  if (trimmed.startsWith("*")) trimmed = trimmed.slice(1).trim();
  const open = topLevelGenericOpenBracket(trimmed);
  if (open === undefined) return undefined;
  const close = matchingTypeBracket(trimmed, open);
  if (close !== trimmed.length - 1 || close === undefined) return undefined;
  return splitTopLevelTypes(trimmed.slice(open + 1, close));
}

function packagePathForTypeText(typeText: string | undefined, env: ExpressionEmitEnv): string | undefined {
  let trimmed = typeText?.trim();
  if (!trimmed) return undefined;
  while (trimmed.startsWith("*")) trimmed = trimmed.slice(1).trim();
  const base = genericBaseTypeText(trimmed);
  const dot = base.indexOf(".");
  if (dot <= 0) return undefined;
  return env.facts.imports.get(base.slice(0, dot));
}

function methodCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv, methodKey: string): string | undefined {
  const callee = expression.callee as SelectorExpression;
  const receiver = methodReceiverToJs(ctx, callee, methodKey, env);
  if (!receiver) return undefined;
  const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(methodKey), expression.spreadLast);
  if (!args) return undefined;
  const typeArgs = methodReceiverTypeArgObject(callee, methodKey, env, receiver);
  return `(await pkg[${JSON.stringify(methodKey)}](${[...(typeArgs ? [typeArgs] : []), receiver, ...args].join(", ")}))`;
}

function callExpressionToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind === "Identifier" && expression.callee.name === "new" && !env.locals.has("new")) {
    return newCallToJs(ctx, expression, env);
  }
  if (expression.callee.kind === "Identifier" && expression.callee.name === "make" && !env.locals.has("make")) {
    return makeCallToJs(ctx, expression, env);
  }
  const conversionType = conversionCalleeTypeText(expression.callee, env, expression.typeText);
  if (conversionType) {
    return conversionCallToJs(ctx, conversionType, expression.args, env);
  }
  const builtin = builtinCallToJs(ctx, expression, env);
  if (builtin) return builtin;
  const unsafeBuiltin = unsafePackageCallToJs(ctx, expression, env);
  if (unsafeBuiltin) return unsafeBuiltin;
  const instantiatedSelectorCall = instantiatedImportedSelectorCallToJs(ctx, expression, env);
  if (instantiatedSelectorCall) return instantiatedSelectorCall;
  if (expression.callee.kind === "SelectorExpression") {
    const functionFieldCall = functionFieldCallToJs(ctx, expression, env);
    if (functionFieldCall) return functionFieldCall;
    const methodKey = methodPackageKeyForSelector(expression.callee, env);
    if (methodKey) return methodCallToJs(ctx, expression, env, methodKey);
    if (isImportedPackageSelector(expression.callee, env) && expression.callee.object.kind === "Identifier") {
      const importPath = env.facts.imports.get(expression.callee.object.name);
      const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
      if (!renderedArgs) return undefined;
      const argTypes = expression.args.map((arg) => JSON.stringify(expressionTypeText(arg, env) ?? ""));
      return `(await __gojrCallImportedFunction(${JSON.stringify(importPath)}, ${JSON.stringify(expression.callee.field)}, [${renderedArgs.join(", ")}], [${argTypes.join(", ")}]))`;
    }
    if (!isImportedPackageSelector(expression.callee, env)) return dynamicMethodCallToJs(ctx, expression, env);
  }
  const instantiatedName = instantiatedFunctionName(expression.callee);
  if (instantiatedName && !env.locals.has(instantiatedName)) {
    const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(instantiatedName), expression.spreadLast);
    if (!args) return undefined;
    const typeArgs = instantiatedFunctionTypeArgObject(expression.callee, instantiatedName, env);
    return `(await pkg[${JSON.stringify(instantiatedName)}](${[...(typeArgs ? [typeArgs] : []), ...args].join(", ")}))`;
  }
  if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
    const renderedArgs = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(expression.callee.name), expression.spreadLast);
    if (!renderedArgs) return undefined;
    const typeArgs = inferredFunctionTypeArgObject(expression.callee.name, expression.args, env);
    return `(await pkg[${JSON.stringify(expression.callee.name)}](${[...(typeArgs ? [typeArgs] : []), ...renderedArgs].join(", ")}))`;
  }
  const callee = expressionToJs(ctx, expression.callee, env);
  if (!callee) return undefined;
  const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
  if (!renderedArgs) return undefined;
  return `(await (${callee})(${renderedArgs.join(", ")}))`;
}

function functionFieldCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind !== "SelectorExpression") return undefined;
  if (isImportedPackageSelector(expression.callee, env)) return undefined;
  const fieldType = selectorFieldTypeText(expression.callee, env);
  const signature = parseFuncTypeText(fieldType);
  if (!signature) return undefined;
  const receiver = expressionToJs(ctx, expression.callee.object, env);
  if (!receiver) return undefined;
  const args = renderCallArgs(ctx, expression.args, env, signature.params, expression.spreadLast);
  if (!args) return undefined;
  return `(await (__gojrGetField(${receiver}, ${JSON.stringify(expression.callee.field)}))(${args.join(", ")}))`;
}

function instantiatedImportedSelectorCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind !== "IndexExpression") return undefined;
  const selector = expression.callee.object;
  if (selector.kind !== "SelectorExpression" || selector.object.kind !== "Identifier") return undefined;
  const importPath = env.facts.imports.get(selector.object.name);
  if (!importPath) return undefined;
  const typeArgText = typeArgumentExpressionText(expression.callee.index);
  if (!typeArgText) return undefined;
  const args = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
  if (!args) return undefined;
  const imported = `__gojrImport(${JSON.stringify(importPath)})`;
  const typeArgs = `__gojrImportedTypeArgs(${imported}, ${JSON.stringify(selector.field)}, [${renderTypeArgumentValues(typeArgText, env).join(", ")}])`;
  return `(await (__gojrGetField(${imported}, ${JSON.stringify(selector.field)}))(${[typeArgs, ...args].join(", ")}))`;
}

function renderTypeArgumentValues(typeArgText: string, env: ExpressionEmitEnv): string[] {
  return splitTopLevelTypes(typeArgText).map((typeArg) => {
    const trimmed = typeArg.trim();
    if (env.typeParameters?.has(trimmed)) return `__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(trimmed)})`;
    return JSON.stringify(trimmed);
  });
}

function conversionCalleeTypeText(callee: Expression, env: ExpressionEmitEnv, callTypeText?: string): string | undefined {
  if (callee.kind === "IndexExpression" &&
    callee.object.kind === "SelectorExpression" &&
    isImportedPackageSelector(callee.object, env)) {
    return undefined;
  }
  if (callee.kind === "TypeExpression") return callee.type.text;
  if (callee.kind === "Identifier") {
    if (env.locals.has(callee.name)) return undefined;
    if (isKnownConversionTypeText(callee.name, env)) return callee.name;
    return undefined;
  }
  if (callee.kind === "SelectorExpression" && callee.object.kind === "Identifier") {
    const importPath = env.facts.imports.get(callee.object.name);
    if (importPath) {
      const candidate = `${callee.object.name}.${callee.field}`;
      return selectorConversionMatchesCallType(candidate, importPath, callTypeText) ? candidate : undefined;
    }
  }
  const typeText = conversionTypeExpressionText(callee, env);
  if (!typeText) return undefined;
  if (isKnownConversionTypeText(typeText, env)) return typeText;
  return undefined;
}

function selectorConversionMatchesCallType(candidate: string, importPath: string, callTypeText: string | undefined): boolean {
  const resultType = callTypeText?.trim();
  if (!resultType) return false;
  if (resultType === candidate) return true;
  const dot = candidate.indexOf(".");
  if (dot < 0) return false;
  const name = candidate.slice(dot + 1);
  return resultType === `${importPath}.${name}`;
}

function conversionTypeExpressionText(expression: Expression, env: ExpressionEmitEnv): string | undefined {
  switch (expression.kind) {
    case "TypeExpression":
      return expression.type.text;
    case "Identifier":
      if (env.locals.has(expression.name)) return undefined;
      return isKnownConversionTypeText(expression.name, env)
        ? expression.name
        : undefined;
    case "SelectorExpression":
      if (expression.object.kind !== "Identifier" || !env.facts.imports.has(expression.object.name)) return undefined;
      if (expression.typeText?.trim().startsWith("func(")) return undefined;
      return `${expression.object.name}.${expression.field}`;
    case "IndexExpression": {
      const object = conversionTypeExpressionText(expression.object, env);
      const index = typeArgumentExpressionText(expression.index);
      return object && index ? `${object}[${index}]` : undefined;
    }
    case "UnaryExpression": {
      if (expression.operator !== "*") return undefined;
      const operand = conversionTypeExpressionText(expression.operand, env);
      return operand ? `*${operand}` : undefined;
    }
    default:
      return undefined;
  }
}

function isKnownConversionTypeText(typeText: string, env: ExpressionEmitEnv): boolean {
  const trimmed = typeText.trim();
  if (!trimmed) return false;
  if (isCompositeTypeText(trimmed)) return true;
  if (trimmed === "unsafe.Pointer") return true;
  if (predeclaredTypeNames.has(trimmed)) return true;
  if (env.typeParameters?.has(trimmed)) return true;
  const base = genericBaseTypeText(trimmed);
  if (env.localTypeUnderlyings.has(base)) return true;
  if (env.facts.typeUnderlyings.has(base)) return true;
  if (directSelectorTypePackagePath(trimmed, env.facts)) return true;
  if (trimmed.startsWith("*")) return isKnownConversionTypeText(trimmed.slice(1).trim(), env);
  return false;
}

function renderCallArgs(ctx: EmitterContext, args: Expression[], env: ExpressionEmitEnv, targetTypes: string[] = [], spreadLast = false): string[] | undefined {
  const rendered: string[] = [];
  for (const [index, arg] of args.entries()) {
    const value = expressionToJs(ctx, arg, env);
    if (!value) return undefined;
    const targetType = callArgumentTargetType(targetTypes, index, spreadLast && index === args.length - 1);
    const lowered = valueForTargetType(value, targetType, expressionTypeText(arg, env), env);
    rendered.push(spreadLast && index === args.length - 1 ? `...__gojrSpread(${lowered})` : lowered);
  }
  return rendered;
}

function callArgumentTargetType(targetTypes: string[] | undefined, index: number, isSpread: boolean): string | undefined {
  if (!targetTypes || targetTypes.length === 0) return undefined;
  const direct = targetTypes[index];
  if (direct !== undefined) {
    return direct.startsWith("...")
      ? (isSpread ? `[]${direct.slice(3)}` : direct.slice(3))
      : direct;
  }
  const variadicIndex = targetTypes.findIndex((type) => type.startsWith("..."));
  if (variadicIndex >= 0 && index >= variadicIndex) return targetTypes[variadicIndex]?.slice(3);
  return undefined;
}

function instantiatedFunctionName(expression: Expression): string | undefined {
  if (expression.kind !== "IndexExpression") return undefined;
  if (expression.object.kind === "Identifier") return expression.object.name;
  return undefined;
}

function instantiatedFunctionTypeArgObject(expression: Expression, functionName: string, env: ExpressionEmitEnv): string | undefined {
  if (expression.kind !== "IndexExpression") return undefined;
  const parameterNames = env.facts.functionTypeParameters.get(functionName);
  if (!parameterNames || parameterNames.length === 0) return undefined;
  const typeArgText = typeArgumentExpressionText(expression.index);
  if (!typeArgText) return undefined;
  const typeArgs = splitTopLevelTypes(typeArgText);
  const fields = parameterNames.map((name, index) =>
    `${JSON.stringify(name)}: ${JSON.stringify(typeArgs[index] ?? "any")}`
  );
  return `({ ${fields.join(", ")} })`;
}

function inferredFunctionTypeArgObject(functionName: string, args: Expression[], env: ExpressionEmitEnv): string | undefined {
  const parameterNames = env.facts.functionTypeParameters.get(functionName);
  if (!parameterNames || parameterNames.length === 0) return undefined;
  const parameterSet = new Set(parameterNames);
  const parameterTypes = env.facts.functionParamTypes.get(functionName) ?? [];
  const inferred = new Map<string, string>();
  for (const [index, parameterType] of parameterTypes.entries()) {
    const actualType = expressionTypeText(args[index], env);
    if (!actualType) continue;
    inferGenericTypeArguments(parameterType.startsWith("...") ? parameterType.slice(3) : parameterType, actualType, parameterSet, inferred);
  }
  const fields = parameterNames.map((name) => `${JSON.stringify(name)}: ${JSON.stringify(inferred.get(name) ?? "any")}`);
  return `({ ${fields.join(", ")} })`;
}

function inferGenericTypeArguments(pattern: string | undefined, actual: string | undefined, parameters: Set<string>, out: Map<string, string>): void {
  let lhs = pattern?.trim();
  const rhs = actual?.trim();
  if (!lhs || !rhs) return;
  if (lhs.startsWith("...")) lhs = lhs.slice(3).trim();
  if (parameters.has(lhs)) {
    if (!out.has(lhs) || out.get(lhs) === "any") out.set(lhs, rhs);
    return;
  }
  const lhsDeref = lhs.startsWith("*") ? lhs.slice(1).trim() : undefined;
  const rhsDeref = rhs.startsWith("*") ? rhs.slice(1).trim() : undefined;
  if (lhsDeref && rhsDeref) {
    inferGenericTypeArguments(lhsDeref, rhsDeref, parameters, out);
    return;
  }
  if (lhs.startsWith("[]") && rhs.startsWith("[]")) {
    inferGenericTypeArguments(lhs.slice(2), rhs.slice(2), parameters, out);
    return;
  }
  const lhsMap = parseMapTypeText(lhs);
  const rhsMap = parseMapTypeText(rhs);
  if (lhsMap && rhsMap) {
    inferGenericTypeArguments(lhsMap.keyType, rhsMap.keyType, parameters, out);
    inferGenericTypeArguments(lhsMap.valueType, rhsMap.valueType, parameters, out);
    return;
  }
  const lhsArgs = genericTypeArgumentsFromTypeText(lhs);
  const rhsArgs = genericTypeArgumentsFromTypeText(rhs);
  if (lhsArgs && rhsArgs && genericBaseTypeText(lhs) === genericBaseTypeText(rhs)) {
    for (const [index, lhsArg] of lhsArgs.entries()) inferGenericTypeArguments(lhsArg, rhsArgs[index], parameters, out);
  }
}

function typeArgumentExpressionText(expression: Expression): string | undefined {
  switch (expression.kind) {
    case "TypeExpression":
      return expression.type.text;
    case "Identifier":
      return expression.name;
    case "SelectorExpression": {
      const object = typeArgumentExpressionText(expression.object);
      return object ? `${object}.${expression.field}` : undefined;
    }
    case "IndexExpression": {
      const object = typeArgumentExpressionText(expression.object);
      const index = typeArgumentExpressionText(expression.index);
      return object && index ? `${object}[${index}]` : undefined;
    }
    case "UnaryExpression": {
      if (expression.operator !== "*") return undefined;
      const operand = typeArgumentExpressionText(expression.operand);
      return operand ? `*${operand}` : undefined;
    }
    default:
      return undefined;
  }
}

function isImportedPackageSelector(expression: SelectorExpression, env: ExpressionEmitEnv): boolean {
  return expression.object.kind === "Identifier" && env.facts.imports.has(expression.object.name);
}

function dynamicMethodCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind !== "SelectorExpression") return undefined;
  const receiver = expressionToJs(ctx, expression.callee.object, env);
  if (!receiver) return undefined;
  const args = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
  if (!args) return undefined;
  return `(await __gojrCallMethod(${receiver}, ${JSON.stringify(expression.callee.field)}, [${args.join(", ")}], pkg))`;
}

function newCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  const typeArg = expression.args[0];
  const typeText = typeArg ? typeArgumentExpressionText(typeArg) : undefined;
  if (expression.args.length !== 1) {
    ctx.emitError("unsupported Stage 4 new arity");
    return undefined;
  }
  if (!typeText) {
    if (!typeArg) {
      ctx.emitError("unsupported Stage 4 new without argument");
      return undefined;
    }
    const value = expressionToJs(ctx, typeArg, env);
    const valueType = expressionTypeText(typeArg, env);
    if (!value || !valueType) {
      ctx.emitError("unsupported Stage 4 new without type argument");
      return undefined;
    }
    const pkgPath = packagePathForTypeText(valueType, env);
    return `__gojrPointerValue(${JSON.stringify(valueType)}, ${valueForTargetType(value, valueType, valueType, env)}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
  }
  const pkgPath = packagePathForTypeText(typeText, env);
  return `__gojrPointerValue(${JSON.stringify(typeText)}, ${zeroValueForTypeInEnv(typeText, env) ?? `__gojrZero(${JSON.stringify(typeText)})`}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
}

function makeCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  const typeArg = expression.args[0];
  const typeText = typeArg ? typeArgumentExpressionText(typeArg) : undefined;
  if (!typeText) {
    ctx.emitError("unsupported Stage 4 make without type argument");
    return undefined;
  }
  if (env.typeParameters?.has(typeText)) {
    const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
    const capacity = expression.args[2] ? expressionToJs(ctx, expression.args[2], env) : length;
    if (!length || !capacity) return undefined;
    return `__gojrMake(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(typeText)}), Number(${length}), Number(${capacity}))`;
  }
  const runtimeTypeText = runtimeTypeTextExpression(typeText, env);
  const resolvedTypeText = resolveUnderlyingTypeTextInEnv(typeText, env);
  if (typeTextReferencesTypeParameter(typeText, env)) {
    const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
    const capacity = expression.args[2] ? expressionToJs(ctx, expression.args[2], env) : length;
    if (!length || !capacity) return undefined;
    return `__gojrMake(${runtimeTypeText}, Number(${length}), Number(${capacity}))`;
  }
  if (chanElementTypeText(resolvedTypeText)) {
    const capacity = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
    if (!capacity) return undefined;
    return `__gojrMakeChan(${JSON.stringify(chanElementTypeText(resolvedTypeText) ?? "any")}, Number(${capacity}))`;
  }
  const mapType = parseMapTypeText(resolvedTypeText);
  if (mapType) {
    return `__gojrMap(new Map(), ${zeroValueForTypeInEnv(mapType.valueType, env) ?? "null"}, ${JSON.stringify(mapType.keyType)}, ${JSON.stringify(mapType.valueType)})`;
  }
  const slice = parseArrayOrSliceTypeText(resolvedTypeText);
  if (slice && resolvedTypeText.startsWith("[]")) {
    const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
    const capacity = expression.args[2] ? expressionToJs(ctx, expression.args[2], env) : length;
    if (!length || !capacity) return undefined;
    return `__gojrMake(${runtimeTypeText}, Number(${length}), Number(${capacity}))`;
  }
  if (typeText !== resolvedTypeText || typeText.includes(".")) {
    const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
    const capacity = expression.args[2] ? expressionToJs(ctx, expression.args[2], env) : length;
    if (!length) return undefined;
    if (!capacity) return undefined;
    return `__gojrMake(${runtimeTypeText}, Number(${length}), Number(${capacity}))`;
  }
  ctx.emitError(`unsupported Stage 4 make(${typeText})`);
  return undefined;
}

function runtimeTypeTextExpression(typeText: string, env: ExpressionEmitEnv): string {
  return typeTextReferencesTypeParameter(typeText, env)
    ? `__gojrSubstituteType(${JSON.stringify(typeText)}, __gojrTypeArgs)`
    : JSON.stringify(typeText);
}

function typeTextReferencesTypeParameter(typeText: string, env: ExpressionEmitEnv): boolean {
  if (!env.typeParameters || env.typeParameters.size === 0) return false;
  for (const parameter of env.typeParameters) {
    const escaped = parameter.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    if (new RegExp(`(^|[^A-Za-z0-9_])${escaped}([^A-Za-z0-9_]|$)`).test(typeText)) return true;
  }
  return false;
}

function builtinCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind !== "Identifier" || env.locals.has(expression.callee.name)) return undefined;
  const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
  if (!renderedArgs) return undefined;
  switch (expression.callee.name) {
    case "len":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 len call arity");
        return undefined;
      }
      return `BigInt(__gojrLen(${renderedArgs[0]}))`;
    case "cap":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 cap call arity");
        return undefined;
      }
      return `BigInt(__gojrCap(${renderedArgs[0]}))`;
    case "append":
      if (renderedArgs.length < 1) {
        ctx.emitError("unsupported Stage 3 append call arity");
        return undefined;
      }
      return `__gojrAppend(${renderedArgs.join(", ")})`;
    case "copy":
      if (renderedArgs.length !== 2) {
        ctx.emitError("unsupported Stage 3 copy call arity");
        return undefined;
      }
      return `BigInt(__gojrCopy(${renderedArgs[0]}, ${renderedArgs[1]}))`;
    case "delete":
      if (renderedArgs.length !== 2) {
        ctx.emitError("unsupported Stage 3 delete call arity");
        return undefined;
      }
      {
        const mapType = parseMapTypeText(resolveUnderlyingTypeTextInEnv(expressionTypeText(expression.args[0], env) ?? "", env));
        const mapArg = renderedArgs[0] ?? "";
        const rawKeyArg = renderedArgs[1] ?? "";
        const key = mapType ? valueForTargetType(rawKeyArg, mapType.keyType, expressionTypeText(expression.args[1], env), env) : rawKeyArg;
        return `__gojrDelete(${mapArg}, ${key})`;
      }
    case "clear":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 clear call arity");
        return undefined;
      }
      return `__gojrClear(${renderedArgs[0]})`;
    case "close":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 close call arity");
        return undefined;
      }
      return `__gojrClose(${renderedArgs[0]})`;
    case "min":
      if (renderedArgs.length < 1) {
        ctx.emitError("unsupported Stage 3 min call arity");
        return undefined;
      }
      return `__gojrMin(${renderedArgs.join(", ")})`;
    case "max":
      if (renderedArgs.length < 1) {
        ctx.emitError("unsupported Stage 3 max call arity");
        return undefined;
      }
      return `__gojrMax(${renderedArgs.join(", ")})`;
    case "panic":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 panic call arity");
        return undefined;
      }
      return `__gojrPanic(${renderedArgs[0]})`;
    case "recover":
      if (renderedArgs.length !== 0) {
        ctx.emitError("unsupported Stage 3 recover call arity");
        return undefined;
      }
      return "__gojrRecover()";
    case "print":
      return `__gojrPrint(__gojrStdout${renderedArgs.length > 0 ? `, ${renderedArgs.join(", ")}` : ""})`;
    case "println":
      return `__gojrPrintln(__gojrStdout${renderedArgs.length > 0 ? `, ${renderedArgs.join(", ")}` : ""})`;
    case "complex":
      if (renderedArgs.length !== 2) {
        ctx.emitError("unsupported Stage 3 complex call arity");
        return undefined;
      }
      return `__gojrComplex(Number(${renderedArgs[0]}), Number(${renderedArgs[1]}))`;
    case "real":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 real call arity");
        return undefined;
      }
      return `__gojrComplexReal(${renderedArgs[0]})`;
    case "imag":
      if (renderedArgs.length !== 1) {
        ctx.emitError("unsupported Stage 3 imag call arity");
        return undefined;
      }
      return `__gojrComplexImag(${renderedArgs[0]})`;
    default:
      return undefined;
  }
}

function unsafePackageCallToJs(ctx: EmitterContext, expression: CallExpression, env: ExpressionEmitEnv): string | undefined {
  if (expression.callee.kind !== "SelectorExpression") return undefined;
  if (expression.callee.object.kind !== "Identifier") return undefined;
  if (env.facts.imports.get(expression.callee.object.name) !== "unsafe") return undefined;
  switch (expression.callee.field) {
    case "Sizeof":
      if (expression.args.length !== 1) {
        ctx.emitError("unsupported Stage 3 unsafe.Sizeof arity");
        return undefined;
      }
      return unsafeSizeofExpressionToJs(ctx, expression.args[0], env);
    case "Alignof":
      if (expression.args.length !== 1) {
        ctx.emitError("unsupported Stage 3 unsafe.Alignof arity");
        return undefined;
      }
      return unsafeAlignofExpressionToJs(ctx, expression.args[0], env);
    case "Offsetof":
      if (expression.args.length !== 1) {
        ctx.emitError("unsupported Stage 3 unsafe.Offsetof arity");
        return undefined;
      }
      return unsafeOffsetofExpressionToJs(ctx, expression.args[0], env);
    default:
      return undefined;
  }
}

function unsafeSizeofExpressionToJs(ctx: EmitterContext, expression: Expression | undefined, env: ExpressionEmitEnv): string | undefined {
  const typeText = expressionTypeText(expression, env)?.trim();
  if (typeText) {
    if (env.typeParameters?.has(typeText)) return `__gojrSizeofType(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(typeText)}))`;
    return `__gojrSizeofType(${JSON.stringify(resolveUnderlyingTypeTextInEnv(typeText, env))})`;
  }
  const value = expressionToJs(ctx, expression, env);
  return value ? `__gojrSizeofValue(${value})` : undefined;
}

function unsafeAlignofExpressionToJs(ctx: EmitterContext, expression: Expression | undefined, env: ExpressionEmitEnv): string | undefined {
  const typeText = expressionTypeText(expression, env)?.trim();
  if (typeText) {
    if (env.typeParameters?.has(typeText)) return `__gojrAlignofType(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(typeText)}))`;
    return `__gojrAlignofType(${JSON.stringify(resolveUnderlyingTypeTextInEnv(typeText, env))})`;
  }
  const value = expressionToJs(ctx, expression, env);
  return value ? `__gojrAlignofValue(${value})` : undefined;
}

function unsafeOffsetofExpressionToJs(ctx: EmitterContext, expression: Expression | undefined, env: ExpressionEmitEnv): string | undefined {
  if (!expression) return undefined;
  if (expression.kind !== "SelectorExpression") {
    ctx.emitError("unsafe.Offsetof requires a selector expression");
    return undefined;
  }
  const object = expressionToJs(ctx, expression.object, env);
  if (!object) return undefined;
  return `__gojrOffsetof(${object}, ${JSON.stringify(expression.field)})`;
}

function conversionCallToJs(ctx: EmitterContext, typeText: string, args: Expression[], env: ExpressionEmitEnv): string | undefined {
  if (args.length !== 1) {
    ctx.emitError(`unsupported Stage 3 conversion to ${typeText} with ${args.length} arguments`);
    return undefined;
  }
  const value = expressionToJs(ctx, args[0], env);
  if (!value) return undefined;
  if (env.typeParameters?.has(typeText.trim())) {
    return `__gojrConvertDynamicType(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(typeText.trim())}), ${value})`;
  }
  const sourceType = expressionTypeText(args[0], env);
  const resolvedType = resolveUnderlyingTypeTextInEnv(typeText, env);
  const pointerPkgPath = pointerConversionPackagePath(typeText, env, ctx.options.artifact.importPath);
  if (isInterfaceTypeText(typeText, env.facts)) return valueForTargetType(value, typeText, sourceType, env);
  if (isByteSliceType(resolvedType)) return `__gojrBytesFrom(${value})`;
  if (typeText.trim() === "unsafe.Pointer" || resolvedType.trim() === "unsafe.Pointer") return `__gojrConvertPointer(${JSON.stringify(typeText)}, ${value}${pointerPkgPath ? `, ${JSON.stringify(pointerPkgPath)}` : ""})`;
  if (resolvedType === "uintptr" && isPointerLikeTypeText(sourceType)) return `__gojrPointerToUintptr(${value})`;
  if (isIntegerType(resolvedType)) return `__gojrIntegerConvert(${JSON.stringify(resolvedType)}, ${value})`;
  if (isFloatType(resolvedType)) return `Number(${value})`;
  if (isComplexType(resolvedType)) return `__gojrToComplex(${value})`;
  if (resolvedType === "bool") return `Boolean(${value})`;
  if (resolvedType === "string") return `__gojrStringFrom(${value})`;
  {
    const arrayTarget = parseArrayOrSliceTypeText(resolvedType);
    if (arrayTarget && !resolvedType.startsWith("[]")) return `__gojrConvertArray(${JSON.stringify(typeText)}, ${JSON.stringify(arrayTarget.elementType)}, ${value})`;
  }
  if (resolvedType.startsWith("[]")) return `__gojrConvertSlice(${JSON.stringify(typeText)}, ${value})`;
  if (resolvedType.startsWith("map[")) return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
  if (resolvedType.startsWith("chan ") || resolvedType.startsWith("<-chan") || resolvedType.startsWith("chan<-")) return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
  if (resolvedType.startsWith("func(")) return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
  if (typeText.trim().startsWith("*") || resolvedType.trim().startsWith("*")) return `__gojrConvertPointer(${JSON.stringify(typeText)}, ${value}${pointerPkgPath ? `, ${JSON.stringify(pointerPkgPath)}` : ""})`;
  const importedTypePkgPath = directSelectorTypePackagePath(typeText, env.facts);
  if (importedTypePkgPath) return `__gojrConvertDynamicType(${JSON.stringify(typeText)}, ${value}, __gojrActiveImportsByPath, ${JSON.stringify(importedTypePkgPath)})`;
  if (env.facts.typeUnderlyings.has(typeText)) return value;
  ctx.emitError(`unsupported Stage 3 conversion to ${typeText}`);
  return undefined;
}

function pointerConversionPackagePath(typeText: string, env: ExpressionEmitEnv, currentImportPath: string): string | undefined {
  const trimmed = typeText.trim();
  if (!trimmed.startsWith("*") && trimmed !== "unsafe.Pointer") return undefined;
  const imported = directSelectorTypePackagePath(trimmed, env.facts);
  if (imported) return imported;
  let elem = trimmed.startsWith("*") ? trimmed.slice(1).trim() : trimmed;
  while (elem.startsWith("*")) elem = elem.slice(1).trim();
  if (elem === "unsafe.Pointer") return "unsafe";
  if (primitiveTypeNames.has(elem) || elem.startsWith("[") || elem.startsWith("[]") || elem.startsWith("map[") || elem.startsWith("chan ") || elem.startsWith("<-chan") || elem.startsWith("chan<-") || elem.startsWith("func(") || elem.startsWith("interface{")) {
    return undefined;
  }
  return currentImportPath;
}

function indexExpressionToJs(ctx: EmitterContext, expression: IndexExpression, env: ExpressionEmitEnv): string | undefined {
  const instantiated = instantiatedImportedSelectorValueToJs(ctx, expression, env);
  if (instantiated) return instantiated;
  const object = expressionToJs(ctx, expression.object, env);
  const index = expressionToJs(ctx, expression.index, env);
  if (!object || !index) return undefined;
  const objectType = expressionTypeText(expression.object, env) ?? packageExportTypeText(ctx, expression.object) ?? "";
  const resolvedObjectType = resolveUnderlyingTypeTextInEnv(objectType, env);
  const mapType = parseMapTypeText(resolvedObjectType);
  if (mapType) {
    const zero = zeroValueForTypeInEnv(mapType.valueType ?? mapValueTypeText(objectType, env.facts) ?? expression.typeText, env) ?? "null";
    const keyValue = valueForTargetType(index, mapType.keyType, expressionTypeText(expression.index, env), env);
    return `__gojrMapGetOk(${object}, ${keyValue}, ${zero})[0]`;
  }
  if (resolvedObjectType === "string") return `BigInt(__gojrStringByteAt(${object}, Number(${index})))`;
  const elementType = expression.typeText ?? parseArrayOrSliceTypeText(resolvedObjectType)?.elementType;
  return `__gojrIndex(${object}, Number(${index}), ${isIntegerType(elementType ?? "") ? "true" : "false"})`;
}

function instantiatedImportedSelectorValueToJs(_ctx: EmitterContext, expression: IndexExpression, env: ExpressionEmitEnv): string | undefined {
  const selector = expression.object;
  if (selector.kind !== "SelectorExpression" || selector.object.kind !== "Identifier") return undefined;
  const importPath = env.facts.imports.get(selector.object.name);
  if (!importPath) return undefined;
  const typeArgText = typeArgumentExpressionText(expression.index);
  if (!typeArgText) return undefined;
  const imported = `__gojrImport(${JSON.stringify(importPath)})`;
  const typeArgs = `__gojrImportedTypeArgs(${imported}, ${JSON.stringify(selector.field)}, [${renderTypeArgumentValues(typeArgText, env).join(", ")}])`;
  return `(async (...__gojrGenericArgs) => await (__gojrGetField(${imported}, ${JSON.stringify(selector.field)}))(${typeArgs}, ...__gojrGenericArgs))`;
}

function sliceExpressionToJs(ctx: EmitterContext, expression: SliceExpression, env: ExpressionEmitEnv): string | undefined {
  const object = expressionToJs(ctx, expression.object, env);
  const start = expression.start ? expressionToJs(ctx, expression.start, env) : undefined;
  const end = expression.end ? expressionToJs(ctx, expression.end, env) : undefined;
  const max = expression.max ? expressionToJs(ctx, expression.max, env) : undefined;
  if (!object || (expression.start && !start) || (expression.end && !end) || (expression.max && !max)) return undefined;
  return `__gojrSlice(${object}, ${start ? `Number(${start})` : "null"}, ${end ? `Number(${end})` : "null"}, ${max ? `Number(${max})` : "null"})`;
}

function typeAssertionExpressionToJs(ctx: EmitterContext, expression: Expression, env: ExpressionEmitEnv = emptyExpressionEnv): string | undefined {
  if (expression.kind !== "TypeAssertionExpression") {
    ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
    return undefined;
  }
  const value = expressionToJs(ctx, expression.expression, env);
  if (!value) return undefined;
  return `__gojrTypeAssert(${value}, ${JSON.stringify(expression.type.text)})`;
}

function byteLiteralValue(expression: Expression): number | undefined {
  if (expression.kind !== "Literal") return undefined;
  if (typeof expression.value === "bigint") {
    const value = Number(expression.value);
    return Number.isInteger(value) && value >= 0 && value <= 255 ? value : undefined;
  }
  if (typeof expression.value === "number") {
    const value = Math.trunc(expression.value);
    return value >= 0 && value <= 255 ? value : undefined;
  }
  return undefined;
}

function bytesFromArrayLiteral(expression: ArrayLiteralExpression): number[] | undefined {
  const bytes = expression.elements.map((element) => byteLiteralValue(element.value));
  return bytes.some((value) => value === undefined) ? undefined : bytes as number[];
}

function utf8Bytes(value: string): number[] {
  return Array.from(new TextEncoder().encode(value));
}

function byteArraysEqual(left: number[], right: number[]): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}

function goStringLiteralBytes(raw: string | undefined): number[] | undefined {
  if (!raw) return undefined;
  if (raw.length >= 2 && raw.startsWith("`") && raw.endsWith("`")) {
    return utf8Bytes(raw.slice(1, -1).replace(/\r/g, ""));
  }
  if (raw.length < 2 || !raw.startsWith("\"") || !raw.endsWith("\"")) return undefined;
  return interpretedGoStringLiteralBytes(raw.slice(1, -1));
}

function interpretedGoStringLiteralBytes(value: string): number[] {
  const bytes: number[] = [];
  for (let index = 0; index < value.length; index += 1) {
    const char = value[index] ?? "";
    if (char !== "\\") {
      bytes.push(...utf8Bytes(char));
      continue;
    }
    const next = value[index + 1] ?? "";
    index += 1;
    switch (next) {
      case "a": bytes.push(0x07); break;
      case "b": bytes.push(0x08); break;
      case "f": bytes.push(0x0c); break;
      case "n": bytes.push(0x0a); break;
      case "r": bytes.push(0x0d); break;
      case "t": bytes.push(0x09); break;
      case "v": bytes.push(0x0b); break;
      case "\\":
      case "\"":
      case "'":
        bytes.push(next.charCodeAt(0));
        break;
      case "x":
        bytes.push(parseInt(value.slice(index + 1, index + 3), 16) & 255);
        index += 2;
        break;
      case "u":
        bytes.push(...utf8Bytes(String.fromCodePoint(parseInt(value.slice(index + 1, index + 5), 16))));
        index += 4;
        break;
      case "U":
        bytes.push(...utf8Bytes(String.fromCodePoint(parseInt(value.slice(index + 1, index + 9), 16))));
        index += 8;
        break;
      default:
        if (/^[0-7]$/.test(next)) {
          bytes.push(parseInt(next + value.slice(index + 1, index + 3), 8) & 255);
          index += 2;
        } else {
          bytes.push(...utf8Bytes(next));
        }
        break;
    }
  }
  return bytes;
}

function mapStringBytesLiteralEntries(expression: MapLiteralExpression): Array<[string, number[]]> | undefined {
  if (expression.keyType.text !== "string" || !isByteSliceType(expression.valueType.text)) return undefined;
  const entries: Array<[string, number[]]> = [];
  for (const entry of expression.entries) {
    if (entry.key.kind !== "Literal" || entry.key.literalKind !== "string" || typeof entry.key.value !== "string") return undefined;
    if (entry.value.kind !== "ArrayLiteralExpression" || !isByteSliceType(entry.value.type.text)) return undefined;
    const bytes = bytesFromArrayLiteral(entry.value);
    if (!bytes) return undefined;
    entries.push([entry.key.value, bytes]);
  }
  return entries;
}

function bytesToBase64(bytes: number[]): string {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
  let out = "";
  for (let index = 0; index < bytes.length; index += 3) {
    const a = bytes[index] ?? 0;
    const b = bytes[index + 1] ?? 0;
    const c = bytes[index + 2] ?? 0;
    const word = (a << 16) | (b << 8) | c;
    out += alphabet[(word >> 18) & 63];
    out += alphabet[(word >> 12) & 63];
    out += index + 1 < bytes.length ? alphabet[(word >> 6) & 63] : "=";
    out += index + 2 < bytes.length ? alphabet[word & 63] : "=";
  }
  return out;
}

function zeroValueForType(typeText: string | undefined, facts: PackageEmitFacts = emptyPackageEmitFacts): string | undefined {
  const staticType = typeText?.trim();
  if (!staticType) return "null";
  const weakType = weakIntrinsicTypeText(staticType, facts);
  if (weakType) return `__gojrIntrinsicZero(${JSON.stringify(weakType)})`;
  const syncType = syncIntrinsicTypeText(staticType, facts);
  if (syncType) return `__gojrIntrinsicZero(${JSON.stringify(syncType)})`;
  if (fsFileModeTypeText(staticType, facts)) return "0n";
  const atomicType = atomicIntrinsicTypeText(staticType, facts);
  if (atomicType) return `__gojrIntrinsicZero(${JSON.stringify(atomicType)})`;
  if (isNilAssignableConcreteTypeText(staticType)) return `__gojrTypedNil(${JSON.stringify(staticType)})`;
  const valueType = resolveUnderlyingTypeText(staticType, facts);
  if (isNilAssignableConcreteTypeText(valueType)) return `__gojrTypedNil(${JSON.stringify(staticType)})`;
  if (facts.typeUnderlyings.has(staticType)) return `__gojrZero(${JSON.stringify(staticType)})`;
  const genericBase = genericBaseTypeText(staticType);
  if (genericBase !== staticType && facts.typeUnderlyings.has(genericBase)) return `__gojrZero(${JSON.stringify(staticType)})`;
  const importedTypePkgPath = directSelectorTypePackagePath(staticType, facts);
  if (importedTypePkgPath) return `__gojrZero(${JSON.stringify(staticType)}, ${JSON.stringify(importedTypePkgPath)})`;
  const arrayType = parseArrayOrSliceTypeText(valueType);
  if (arrayType && !valueType.startsWith("[]")) {
    const length = parseArrayLengthTypeText(valueType, facts);
    if (length !== undefined) {
      if (arrayType.elementType === "byte" || arrayType.elementType === "uint8") return `__gojrSetCap(new Uint8Array(${length}), ${length})`;
      return `Array.from({ length: ${length} }, () => ${zeroValueForType(arrayType.elementType, facts) ?? "null"})`;
    }
  }
  const anonymousStructFields = parseAnonymousStructFields(valueType);
  if (anonymousStructFields) {
    const fields = anonymousStructFields.map((field) =>
      `${JSON.stringify(field.name)}: ${zeroValueForType(field.type, facts) ?? "null"}`
    );
    return `__gojrStruct(${JSON.stringify(staticType)}, { ${fields.join(", ")} })`;
  }
  switch (valueType) {
    case "int":
    case "rune":
    case "int8":
    case "int16":
    case "int32":
    case "int64":
    case "uint":
    case "byte":
    case "uint8":
    case "uint16":
    case "uint32":
    case "uint64":
    case "uintptr":
      return "0n";
    case "float32":
    case "float64":
      return "0";
    case "complex64":
    case "complex128":
      return "__gojrComplex(0, 0)";
    case "string":
      return "\"\"";
    case "bool":
      return "false";
    default:
      return "null";
  }
}

function atomicIntrinsicTypeText(typeText: string, facts: PackageEmitFacts): string | undefined {
  const trimmed = typeText.trim();
  if (/^sync\/atomic\.(?:Bool|Int32|Int64|Uint32|Uint64|Uintptr|Value|Pointer(?:\[[\s\S]*\])?)$/.test(trimmed)) {
    return trimmed;
  }
  const dot = trimmed.indexOf(".");
  if (dot <= 0) return undefined;
  const qualifier = trimmed.slice(0, dot);
  const rest = trimmed.slice(dot + 1);
  return facts.imports.get(qualifier) === "sync/atomic" &&
    /^(?:Bool|Int32|Int64|Uint32|Uint64|Uintptr|Value|Pointer(?:\[[\s\S]*\])?)$/.test(rest)
    ? `sync/atomic.${rest}`
    : undefined;
}

function weakIntrinsicTypeText(typeText: string, facts: PackageEmitFacts): string | undefined {
  const trimmed = typeText.trim();
  if (/^weak\.Pointer(?:\[[\s\S]*\])?$/.test(trimmed)) return trimmed;
  const dot = trimmed.indexOf(".");
  if (dot <= 0) return undefined;
  const qualifier = trimmed.slice(0, dot);
  const rest = trimmed.slice(dot + 1);
  return facts.imports.get(qualifier) === "weak" && /^Pointer(?:\[[\s\S]*\])?$/.test(rest)
    ? `weak.${rest}`
    : undefined;
}

function fsFileModeTypeText(typeText: string, facts: PackageEmitFacts): string | undefined {
  const trimmed = typeText.trim();
  if (trimmed === "io/fs.FileMode") return trimmed;
  const dot = trimmed.indexOf(".");
  if (dot <= 0) return undefined;
  const qualifier = trimmed.slice(0, dot);
  const rest = trimmed.slice(dot + 1);
  return facts.imports.get(qualifier) === "io/fs" && rest === "FileMode" ? "io/fs.FileMode" : undefined;
}

function syncIntrinsicTypeText(typeText: string, facts: PackageEmitFacts): string | undefined {
  const trimmed = typeText.trim();
  if (/^sync\.(?:Mutex|RWMutex|Once|Pool|WaitGroup|Cond|Map)$/.test(trimmed)) return trimmed;
  if (/^internal\/sync\.HashTrieMap(?:\[[\s\S]*\])?$/.test(trimmed)) return trimmed;
  const dot = trimmed.indexOf(".");
  if (dot <= 0) return undefined;
  const qualifier = trimmed.slice(0, dot);
  const rest = trimmed.slice(dot + 1);
  const importPath = facts.imports.get(qualifier);
  if (importPath === "sync" && /^(?:Mutex|RWMutex|Once|Pool|WaitGroup|Cond|Map)$/.test(rest)) {
    return `sync.${rest}`;
  }
  if (importPath === "internal/sync" && /^HashTrieMap(?:\[[\s\S]*\])?$/.test(rest)) {
    return `internal/sync.${rest}`;
  }
  return undefined;
}

function parseAnonymousStructFields(typeText: string): Array<{ name: string; type: string; embedded?: boolean }> | undefined {
  const match = /^struct\s*\{([\s\S]*)\}$/.exec(typeText.trim());
  if (!match) return undefined;
  const body = match[1]?.trim() ?? "";
  if (body === "") return [];
  return splitTopLevel(body, ";").flatMap(parseAnonymousStructField);
}

function parseAnonymousStructField(text: string): Array<{ name: string; type: string; embedded?: boolean }> {
  const field = stripStructFieldTag(text.trim());
  if (!field) return [];
  const split = firstAnonymousStructFieldNameTypeSplit(field);
  if (split < 0) {
    return [{ name: embeddedFieldNameFromTypeText(field), type: field, embedded: true }];
  }
  const namesText = field.slice(0, split).trim();
  const typeText = field.slice(split).trim();
  return splitTopLevel(namesText, ",")
    .map((name) => name.trim())
    .filter(Boolean)
    .map((name) => ({ name, type: typeText }));
}

function stripStructFieldTag(field: string): string {
  let depth = 0;
  for (let index = 0; index < field.length; index += 1) {
    const char = field[index] ?? "";
    if (char === "[" || char === "(" || char === "{") depth += 1;
    else if (char === "]" || char === ")" || char === "}") depth -= 1;
    else if (char === "`" && depth === 0) return field.slice(0, index).trim();
  }
  return field;
}

function firstAnonymousStructFieldNameTypeSplit(field: string): number {
  let depth = 0;
  for (let index = 0; index < field.length; index += 1) {
    const char = field[index] ?? "";
    if (char === "[" || char === "(" || char === "{") depth += 1;
    else if (char === "]" || char === ")" || char === "}") depth -= 1;
    if (depth !== 0 || !/\s/.test(char)) continue;
    const namesText = field.slice(0, index).trim();
    if (!anonymousStructFieldNameList(namesText)) continue;
    let next = index;
    while (next < field.length && /\s/.test(field[next] ?? "")) next += 1;
    if (next < field.length) return index;
  }
  return -1;
}

function anonymousStructFieldNameList(namesText: string): boolean {
  const names = splitTopLevel(namesText, ",").map((name) => name.trim());
  return names.length > 0 && names.every((name) => /^[A-Za-z_]\w*$/.test(name));
}

function embeddedFieldNameFromTypeText(typeText: string): string {
  let text = typeText.trim();
  while (text.startsWith("*")) text = text.slice(1).trim();
  const generic = text.indexOf("[");
  if (generic >= 0) text = text.slice(0, generic).trim();
  const dot = text.lastIndexOf(".");
  return dot >= 0 ? text.slice(dot + 1) : text;
}

function splitTopLevel(text: string, separator: string): string[] {
  const out: string[] = [];
  let start = 0;
  let parenDepth = 0;
  let bracketDepth = 0;
  let braceDepth = 0;
  let inBacktick = false;
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index] ?? "";
    if (char === "`") {
      inBacktick = !inBacktick;
      continue;
    }
    if (inBacktick) continue;
    if (char === "(") parenDepth += 1;
    else if (char === ")") parenDepth -= 1;
    else if (char === "[") bracketDepth += 1;
    else if (char === "]") bracketDepth -= 1;
    else if (char === "{") braceDepth += 1;
    else if (char === "}") braceDepth -= 1;
    else if (char === separator && parenDepth === 0 && bracketDepth === 0 && braceDepth === 0) {
      const item = text.slice(start, index).trim();
      if (item) out.push(item);
      start = index + 1;
    }
  }
  const last = text.slice(start).trim();
  if (last) out.push(last);
  return out;
}

function zeroValueForTypeInEnv(typeText: string | undefined, env: ExpressionEmitEnv): string | undefined {
  const staticType = typeText?.trim();
  if (staticType && env.typeParameters?.has(staticType)) {
    return `__gojrZero(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(staticType)}))`;
  }
  if (staticType && env.localTypeUnderlyings.has(staticType)) {
    const resolved = resolveUnderlyingTypeTextInEnv(staticType, env);
    if (isNilAssignableConcreteTypeText(resolved)) return `__gojrTypedNil(${JSON.stringify(staticType)})`;
    return zeroValueForType(resolved, env.facts);
  }
  return zeroValueForType(staticType ?? typeText, env.facts);
}

function resolveUnderlyingTypeTextInEnv(typeText: string, env: ExpressionEmitEnv): string {
  let current = typeText.trim();
  const seen = new Set<string>();
  while (current && !seen.has(current)) {
    seen.add(current);
    const underlying = env.localTypeUnderlyings.get(current) ?? env.facts.typeUnderlyings.get(current);
    if (!underlying) break;
    current = underlying.trim();
  }
  return current;
}

function resolveUnderlyingTypeText(typeText: string, facts: PackageEmitFacts): string {
  let current = typeText.trim();
  const seen = new Set<string>();
  while (!seen.has(current)) {
    seen.add(current);
    const underlying = facts.typeUnderlyings.get(current);
    if (!underlying) break;
    current = underlying.trim();
  }
  return current;
}

function isNilAssignableConcreteTypeText(typeText: string): boolean {
  return typeText.startsWith("*") ||
    typeText.startsWith("[]") ||
    typeText.startsWith("map[") ||
    typeText.startsWith("chan ") ||
    typeText.startsWith("<-chan") ||
    typeText.startsWith("chan<-") ||
    typeText.startsWith("func(");
}

function isCompositeTypeText(typeText: string): boolean {
  const trimmed = typeText.trim();
  return trimmed.startsWith("*") ||
    trimmed.startsWith("[") ||
    trimmed.startsWith("[]") ||
    trimmed.startsWith("map[") ||
    trimmed.startsWith("chan ") ||
    trimmed.startsWith("<-chan") ||
    trimmed.startsWith("chan<-") ||
    trimmed.startsWith("func(") ||
    trimmed.startsWith("struct{") ||
    trimmed.startsWith("interface{");
}

function mapValueTypeText(typeText: string | undefined, facts: PackageEmitFacts = emptyPackageEmitFacts): string | undefined {
  if (!typeText) return undefined;
  return parseMapTypeText(resolveUnderlyingTypeText(typeText, facts))?.valueType;
}

function parseMapTypeText(typeText: string | undefined): { keyType: string; valueType: string } | undefined {
  if (!typeText?.startsWith("map[")) return undefined;
  let depth = 0;
  for (let index = 4; index < typeText.length; index += 1) {
    const char = typeText[index];
    if (char === "[") depth += 1;
    if (char === "]") {
      if (depth === 0) {
        return {
          keyType: typeText.slice(4, index).trim(),
          valueType: typeText.slice(index + 1).trim()
        };
      }
      depth -= 1;
    }
  }
  return undefined;
}

function parseArrayOrSliceTypeText(typeText: string | undefined): { elementType: string } | undefined {
  if (!typeText?.startsWith("[")) return undefined;
  const close = typeText.indexOf("]");
  if (close < 0) return undefined;
  return { elementType: typeText.slice(close + 1).trim() };
}

function chanElementTypeText(typeText: string | undefined): string | undefined {
  if (!typeText) return undefined;
  const trimmed = typeText.trim();
  if (trimmed.startsWith("chan<-")) return trimmed.slice("chan<-".length).trim();
  if (trimmed.startsWith("<-chan")) return trimmed.slice("<-chan".length).trim();
  if (trimmed.startsWith("chan ")) return trimmed.slice("chan ".length).trim();
  return undefined;
}

function isByteSliceType(typeText: string): boolean {
  return typeText === "[]byte" || typeText === "[]uint8";
}

const primitiveTypeNames = new Set([
  "bool",
  "string",
  "int",
  "int8",
  "int16",
  "int32",
  "int64",
  "uint",
  "uint8",
  "byte",
  "uint16",
  "uint32",
  "uint64",
  "uintptr",
  "rune",
  "float32",
  "float64",
  "complex64",
  "complex128"
]);
const predeclaredTypeNames = new Set([...primitiveTypeNames, "any", "error"]);

function isIntegerType(typeText: string): boolean {
  return typeText === "int" ||
    typeText === "byte" ||
    typeText === "rune" ||
    typeText === "int8" ||
    typeText === "int16" ||
    typeText === "int32" ||
    typeText === "int64" ||
    typeText === "uint" ||
    typeText === "uint8" ||
    typeText === "uint16" ||
    typeText === "uint32" ||
    typeText === "uint64" ||
    typeText === "uintptr";
}

function isFloatType(typeText: string): boolean {
  return typeText === "float32" || typeText === "float64" || typeText === "untyped float";
}

function isComplexType(typeText: string | undefined): boolean {
  return typeText === "complex64" || typeText === "complex128" || typeText === "untyped complex";
}

function isPointerLikeTypeText(typeText: string | undefined): boolean {
  const text = typeText?.trim();
  return Boolean(text && (text === "unsafe.Pointer" || text.startsWith("*")));
}

const numberArithmeticOperators = new Set<BinaryExpression["operator"]>(["+", "-", "*", "/"]);
const numberComparisonOperators = new Set<BinaryExpression["operator"]>(["==", "!=", "<", "<=", ">", ">="]);

function expressionPrefersNumberArithmetic(expression: Expression | undefined, env: ExpressionEmitEnv): boolean {
  if (!expression) return false;
  const typeText = expressionTypeText(expression, env);
  if (typeText && isFloatType(resolveUnderlyingTypeTextInEnv(typeText, env))) return true;
  switch (expression.kind) {
    case "Literal":
      return expression.literalKind === "float";
    case "UnaryExpression":
      return expressionPrefersNumberArithmetic(expression.operand, env);
    case "BinaryExpression":
      return expressionPrefersNumberArithmetic(expression.left, env) || expressionPrefersNumberArithmetic(expression.right, env);
    case "CallExpression":
      if (expression.callee.kind === "TypeExpression") return isFloatType(resolveUnderlyingTypeTextInEnv(expression.callee.type.text, env));
      return isFloatType(typeText ?? "");
    default:
      return false;
  }
}

function isComplexLiteralValue(value: LiteralExpression["value"]): value is { real: number; imag: number } {
  return Boolean(value && typeof value === "object" && "real" in value && "imag" in value);
}

function jsBinaryOperator(operator: BinaryExpression["operator"]): string | undefined {
  switch (operator) {
    case "==":
      return "===";
    case "!=":
      return "!==";
    case "<":
    case "<=":
    case ">":
    case ">=":
    case "+":
    case "-":
    case "|":
    case "^":
    case "*":
    case "/":
    case "%":
    case "<<":
    case ">>":
    case "&":
      return operator;
    default:
      return undefined;
  }
}

function safeLocalName(name: string, index: number): string {
  if (name !== "_" && /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(name) && !JS_RESERVED_WORDS.has(name)) return name;
  const encoded = Array.from(name)
    .map((char) => char.codePointAt(0)?.toString(16) ?? "0")
    .join("_");
  const suffix = index < 0 ? "recv" : String(index);
  return `__gojrIdent_${encoded || "blank"}_${suffix}`;
}

function stage1JavaScript(artifact: GoJuniorPackageExportData, usesWasm: boolean, bodyLines: string[], typeDescriptorLines: string[], facts: PackageEmitFacts): string {
  const artifactHeader: GoJuniorPackageExportData = { ...artifact };
  delete artifactHeader.runtime;
  const generatedBuiltinImports = JSON.stringify(GENERATED_ARTIFACT_INTRINSIC_IMPORTS);
  const importPathsByQualifier = Object.fromEntries([...facts.imports.entries()].sort(([left], [right]) => left.localeCompare(right)));
  const source = [
    "// Code generated by gojr Stage 1 copy-and-patch emitter; DO NOT EDIT.",
    `export const gojrPackageArtifact = ${JSON.stringify(artifactHeader, null, 2)};`,
    `const __gojrImportPathsByQualifier = Object.freeze(${JSON.stringify(importPathsByQualifier, null, 2)});`,
    "const __gojrTypeDescriptors = Object.create(null);",
    ...typeDescriptorLines,
    "let __gojrActiveImportsByPath = {};",
    "let __gojrActivePackage = {};",
    "let __gojrActiveRuntimeOptions = {};",
    "export async function instantiateGoJrPackage(runtime = {}, options = {}) {",
    "  const pkg = Object.create(null);",
    "  const importsByPath = options.importsByPath || runtime.importsByPath || options.packages || runtime.packages || {};",
    "  __gojrActiveImportsByPath = importsByPath;",
    "  __gojrActivePackage = pkg;",
    "  const __gojrStdout = options.stdout || runtime.stdout || __gojrDefaultTextSink(\"stdout\") || (() => {});",
    "  const __gojrStderr = options.stderr || runtime.stderr || __gojrDefaultTextSink(\"stderr\") || (() => {});",
    "  __gojrActiveRuntimeOptions = { ...runtime, ...options, stdout: __gojrStdout, stderr: __gojrStderr };",
    "  const __gojrImport = (path) => {",
    "    const builtin = __gojrBuiltinImport(path, importsByPath);",
    "    const imported = importsByPath[path];",
    "    const packageObject = imported && typeof imported === \"object\" && \"package\" in imported ? imported.package : imported;",
    "    if (__gojrBuiltinImportOverrides(path) && builtin) return __gojrMergedBuiltinPackage(builtin, packageObject);",
    "    return packageObject || builtin || {};",
    "  };",
    "  const __gojrPackageSelector = (path, field) => __gojrSelectPackageField(importsByPath, path, field);",
    "  const __gojrPackageCallableSelector = (path, field) => __gojrSelectPackageCallableField(importsByPath, path, field);",
    usesWasm ? "  const wasmBase64 = options.wasmBase64 || runtime.wasmBase64;" : "",
    usesWasm ? "  if (typeof wasmBase64 !== \"string\" || wasmBase64.length === 0) throw new Error(\"gojr Stage 1 package requires options.wasmBase64 from _gojr.wasm\");" : "",
    usesWasm ? "  const __gojrWasmModule = new WebAssembly.Module(__gojrDecodeBase64(wasmBase64));" : "",
    usesWasm ? "  const __gojrWasmInstance = new WebAssembly.Instance(__gojrWasmModule, {});" : "",
    usesWasm ? "  const __gojrWasmExports = __gojrWasmInstance.exports;" : "  const __gojrWasmExports = {};",
    ...bodyLines,
    "  Object.defineProperty(pkg, \"__gojrImportPath\", { value: gojrPackageArtifact.importPath });",
    "  Object.defineProperty(pkg, \"__gojrPackageName\", { value: gojrPackageArtifact.packageName });",
    "  Object.defineProperty(pkg, \"__gojrTypeDescriptors\", { value: __gojrTypeDescriptors });",
    "  Object.defineProperty(pkg, \"__gojrExportIndex\", { value: gojrPackageArtifact.exportIndex });",
    "  Object.defineProperty(pkg, \"__gojrImportPathsByQualifier\", { value: __gojrImportPathsByQualifier });",
    "  return { diagnostics: [], output: [], package: pkg, wasm: __gojrWasmExports };",
    "}",
    "function __gojrStringByteAt(value, index) {",
    "  const bytes = __gojrStringBytes(value);",
    "  if (index < 0 || index >= bytes.length) throw new RangeError(\"string index out of range\");",
    "  return bytes[index];",
    "}",
    "function __gojrRawString(bytes) {",
    "  const rawBytes = bytes instanceof Uint8Array ? new Uint8Array(bytes) : Uint8Array.from(bytes ?? [], (item) => Number(item) & 255);",
    "  return { __gojrRawString: true, bytes: rawBytes, toString() { return new TextDecoder().decode(rawBytes); }, valueOf() { return new TextDecoder().decode(rawBytes); } };",
    "}",
    "function __gojrRawStringBase64(base64) {",
    "  return __gojrRawString(__gojrDecodeBase64(base64));",
    "}",
    "function __gojrIsString(value) {",
    "  return typeof value === \"string\" || Boolean(value && value.__gojrRawString === true);",
    "}",
    "function __gojrBytesEqual(left, right) {",
    "  if (left.length !== right.length) return false;",
    "  for (let index = 0; index < left.length; index += 1) if (left[index] !== right[index]) return false;",
    "  return true;",
    "}",
    "function __gojrStringBytes(value) {",
    "  if (value && value.__gojrRawString === true) return new Uint8Array(value.bytes);",
    "  return new TextEncoder().encode(String(value));",
    "}",
    "function __gojrStringFromBytes(bytes) {",
    "  const raw = bytes instanceof Uint8Array ? new Uint8Array(bytes) : Uint8Array.from(bytes ?? [], (item) => Number(item) & 255);",
    "  const text = new TextDecoder().decode(raw);",
    "  return __gojrBytesEqual(new TextEncoder().encode(text), raw) ? text : __gojrRawString(raw);",
    "}",
    "function __gojrConcatStrings(left, right) {",
    "  if (left && left.__gojrRawString === true || right && right.__gojrRawString === true) {",
    "    const leftBytes = __gojrStringBytes(left);",
    "    const rightBytes = __gojrStringBytes(right);",
    "    const out = new Uint8Array(leftBytes.length + rightBytes.length);",
    "    out.set(leftBytes, 0);",
    "    out.set(rightBytes, leftBytes.length);",
    "    return __gojrStringFromBytes(out);",
    "  }",
    "  return String(left) + String(right);",
    "}",
    "function __gojrStringFrom(value) {",
    "  if (value && value.__gojrRawString === true) return value;",
    "  if (value instanceof Uint8Array) return __gojrStringFromBytes(value);",
    "  if (ArrayBuffer.isView(value)) return __gojrStringFromBytes(new Uint8Array(value.buffer, value.byteOffset, value.byteLength));",
    "  if (Array.isArray(value)) return __gojrStringFromBytes(Uint8Array.from(value, (item) => Number(item) & 255));",
    "  return String(value);",
    "}",
    "function __gojrTuple(values) {",
    "  Object.defineProperty(values, \"__gojrTuple\", { value: true });",
    "  return values;",
    "}",
    "function __gojrTupleValues(value) {",
    "  if (Array.isArray(value) && value.__gojrTuple === true) return value;",
    "  if (Array.isArray(value)) return value;",
    "  return [value];",
    "}",
    "const __gojrObjectIds = new WeakMap();",
    "let __gojrNextObjectId = 1;",
    "let __gojrNextNonReflexiveMapKeyId = 1;",
    "function __gojrUnwrapInterface(value) {",
    "  while (value && value.__gojrInterface === true) value = value.value;",
    "  return value;",
    "}",
    "function __gojrEqual(left, right) {",
    "  if (left && left.__gojrInterface === true || right && right.__gojrInterface === true) {",
    "    if (left && left.__gojrInterface === true && right === null) return left.value === null;",
    "    if (right && right.__gojrInterface === true && left === null) return right.value === null;",
    "    left = __gojrUnwrapInterface(left);",
    "    right = __gojrUnwrapInterface(right);",
    "  }",
    "  if (left && left.__gojrTypedNil === true && right === null) return true;",
    "  if (right && right.__gojrTypedNil === true && left === null) return true;",
    "  if (left && left.__gojrTypedNil === true && right && right.__gojrTypedNil === true) return left.__gojrType === right.__gojrType || (__gojrNilPointerType(left.__gojrType) && __gojrNilPointerType(right.__gojrType));",
    "  if (left && left.__gojrWeakPointer === true && right && right.__gojrWeakPointer === true) return __gojrEqual(left.__gojrWeakValue, right.__gojrWeakValue);",
    "  if ((typeof left === \"bigint\" || typeof left === \"number\") && (typeof right === \"bigint\" || typeof right === \"number\")) return __gojrCompareValues(left, right) === 0;",
    "  if (left && typeof left === \"object\" && \"real\" in left && \"imag\" in left || right && typeof right === \"object\" && \"real\" in right && \"imag\" in right) return __gojrComplexEq(left, right);",
    "  if (__gojrIsString(left) || __gojrIsString(right)) return __gojrIsString(left) && __gojrIsString(right) && __gojrBytesEqual(__gojrStringBytes(left), __gojrStringBytes(right));",
    "  if (left && left.__gojrPointer === true && right && right.__gojrPointer === true) return __gojrPointerIdentityKey(left) === __gojrPointerIdentityKey(right);",
    "  if (Array.isArray(left) && Array.isArray(right)) return left.length === right.length && left.every((item, index) => __gojrEqual(item, right[index]));",
    "  if (ArrayBuffer.isView(left) && ArrayBuffer.isView(right)) {",
    "    if (left.length !== right.length) return false;",
    "    for (let index = 0; index < left.length; index += 1) if (!__gojrEqual(left[index], right[index])) return false;",
    "    return true;",
    "  }",
    "  if (left && right && typeof left === \"object\" && typeof right === \"object\" && left.__gojrType && right.__gojrType) {",
    "    if (left.__gojrType !== right.__gojrType) return false;",
    "    const fields = __gojrStructComparableFields(left);",
    "    return fields.every((field) => __gojrEqual(left[field], right[field]));",
    "  }",
    "  return left === right;",
    "}",
    "function __gojrNilPointerType(typeName) {",
    "  const text = String(typeName || \"\").trim();",
    "  return text === \"unsafe.Pointer\" || text.startsWith(\"*\");",
    "}",
    "function __gojrStructComparableFields(value) {",
    "  const descriptor = __gojrDescriptorForTypeName(__gojrRuntimeTypeName(value), __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(value));",
    "  if (descriptor && Array.isArray(descriptor.fields)) return descriptor.fields.map((field) => field.name);",
    "  return Object.keys(value).filter((key) => !key.startsWith(\"__gojr\"));",
    "}",
    "function __gojrObjectIdentityId(value) {",
    "  if (value === null || (typeof value !== \"object\" && typeof value !== \"function\")) return 0;",
    "  const existing = __gojrObjectIds.get(value);",
    "  if (existing !== undefined) return existing;",
    "  const id = __gojrNextObjectId++;",
    "  __gojrObjectIds.set(value, id);",
    "  return id;",
    "}",
    "function __gojrMapNumberKeyId(value, options = {}) {",
    "  if (Number.isNaN(value)) {",
    "    if (options.freshNonReflexive) return `NaN#${__gojrNextNonReflexiveMapKeyId++}`;",
    "    return \"NaN\";",
    "  }",
    "  return String(value);",
    "}",
    "function __gojrMapKeyId(key, options = {}) {",
    "  const actual = __gojrUnwrapInterface(key);",
    "  if (actual === null || actual === undefined) return \"nil\";",
    "  if (actual && actual.__gojrTypedNil === true) return `typednil:${actual.__gojrType}`;",
    "  if (typeof actual === \"boolean\") return `b:${actual}`;",
    "  if (__gojrIsString(actual)) return `s:${Array.from(__gojrStringBytes(actual)).map((byte) => byte.toString(16).padStart(2, \"0\")).join(\"\")}`;",
    "  if (typeof actual === \"bigint\") return `i:${actual}`;",
    "  if (typeof actual === \"number\") return `f:${__gojrMapNumberKeyId(actual, options)}`;",
    "  if (actual && typeof actual === \"object\" && \"real\" in actual && \"imag\" in actual) return `c:${__gojrMapNumberKeyId(actual.real, options)}:${__gojrMapNumberKeyId(actual.imag, options)}`;",
    "  if (actual && actual.__gojrPointer === true) return `ptr:${__gojrPointerIdentityKey(actual)}`;",
    "  if (Array.isArray(actual) || ArrayBuffer.isView(actual)) return `a:[${Array.from(actual).map((item) => __gojrMapKeyId(item, options)).join(\",\")}]`; ",
    "  if (actual && typeof actual === \"object\" && actual.__gojrType) return `st:${actual.__gojrType}{${__gojrStructComparableFields(actual).map((name) => `${name}:${__gojrMapKeyId(actual[name], options)}`).join(\",\")}}`; ",
    "  return `o:${__gojrObjectIdentityId(actual)}`;",
    "}",
    "function __gojrPointerIdentityKey(pointer) {",
    "  if (!pointer || pointer.__gojrPointer !== true) return `object:${__gojrObjectIdentityId(pointer)}`;",
    "  if (pointer.__gojrIdentity !== undefined) return String(pointer.__gojrIdentity);",
    "  if (pointer.__gojrSequence) return `seq:${__gojrObjectIdentityId(pointer.__gojrSequence.values)}:${Number(pointer.__gojrSequence.index || 0)}:${pointer.__gojrElemType || \"\"}`;",
    "  return `ptr:${__gojrObjectIdentityId(pointer)}`;",
    "}",
    "function __gojrMapGetOk(map, key, zero) {",
    "  if (!(map instanceof Map)) return __gojrTuple([zero, false]);",
    "  if (zero === null && Object.prototype.hasOwnProperty.call(map, \"__gojrValueZero\")) zero = map.__gojrValueZero;",
    "  if (map.__gojrGoMap === true) {",
    "    const entry = Map.prototype.get.call(map, __gojrMapKeyId(key));",
    "    return entry ? __gojrTuple([entry.value, true]) : __gojrTuple([zero, false]);",
    "  }",
    "  return map.has(key) ? __gojrTuple([map.get(key), true]) : __gojrTuple([zero, false]);",
    "}",
    "function __gojrMapSet(map, key, value) {",
    "  if (map && map.__gojrTypedNil === true && __gojrDescriptorForTypeName(map.__gojrType)?.kind === \"map\") throw new Error(\"GOJR_RUNTIME001: assignment to entry in nil map\");",
    "  if (!(map instanceof Map)) throw new Error(\"GOJR_RUNTIME001: assignment target is not a map\");",
    "  if (map.__gojrGoMap === true) Map.prototype.set.call(map, __gojrMapKeyId(key, { freshNonReflexive: true }), { key, value });",
    "  else map.set(key, value);",
    "}",
    "function __gojrMap(map, valueZero, keyType = \"any\", valueType = \"any\") {",
    "  const out = new Map();",
    "  Object.defineProperty(out, \"__gojrGoMap\", { value: true });",
    "  Object.defineProperty(out, \"__gojrValueZero\", { value: valueZero });",
    "  Object.defineProperty(out, \"__gojrKeyType\", { value: keyType });",
    "  Object.defineProperty(out, \"__gojrValueType\", { value: valueType });",
    "  Object.defineProperty(out, \"get\", { value(key) { const entry = Map.prototype.get.call(this, __gojrMapKeyId(key)); return entry ? entry.value : undefined; } });",
    "  Object.defineProperty(out, \"has\", { value(key) { return Map.prototype.has.call(this, __gojrMapKeyId(key)); } });",
    "  Object.defineProperty(out, \"delete\", { value(key) { return Map.prototype.delete.call(this, __gojrMapKeyId(key)); } });",
    "  Object.defineProperty(out, \"set\", { value(key, value) { __gojrMapSet(this, key, value); return this; } });",
    "  Object.defineProperty(out, \"entries\", { value: function* entries() { for (const entry of Map.prototype.values.call(this)) yield [entry.key, entry.value]; } });",
    "  Object.defineProperty(out, \"keys\", { value: function* keys() { for (const entry of Map.prototype.values.call(this)) yield entry.key; } });",
    "  Object.defineProperty(out, \"values\", { value: function* values() { for (const entry of Map.prototype.values.call(this)) yield entry.value; } });",
    "  Object.defineProperty(out, Symbol.iterator, { value: out.entries });",
    "  Object.defineProperty(out, \"forEach\", { value(callback, thisArg) { for (const [key, value] of this.entries()) callback.call(thisArg, value, key, this); } });",
    "  for (const [key, value] of map instanceof Map ? map.entries() : []) Map.prototype.set.call(out, __gojrMapKeyId(key, { freshNonReflexive: true }), { key, value });",
    "  return out;",
    "}",
    "function __gojrBytesBase64(base64) {",
    "  return __gojrDecodeBase64(base64);",
    "}",
    "function __gojrBytesFrom(value) {",
    "  if (__gojrIsString(value)) return __gojrStringBytes(value);",
    "  if (value && value.__gojrByteView === true) return value.subarray(0, value.length);",
    "  if (value instanceof Uint8Array) return new Uint8Array(value);",
    "  if (ArrayBuffer.isView(value)) return new Uint8Array(value.buffer.slice(value.byteOffset, value.byteOffset + value.byteLength));",
    "  if (Array.isArray(value)) return Uint8Array.from(value, (item) => Number(item) & 255);",
    "  return new Uint8Array();",
    "}",
    "function __gojrRawBytesFromGoString(value) {",
    "  return __gojrStringBytes(value);",
    "}",
    "function __gojrIsByteSliceTypeName(typeName) {",
    "  return typeName === \"[]byte\" || typeName === \"[]uint8\";",
    "}",
    "function __gojrMapStringBytesBase64(entries) {",
    "  const map = new Map();",
    "  for (const [key, base64] of entries) map.set(key, __gojrBytesBase64(base64));",
    "  return __gojrMap(map, new Uint8Array(), \"string\", \"[]byte\");",
    "}",
    "function __gojrLen(value) {",
    "  if (value == null) return 0;",
    "  if (__gojrIsString(value)) return __gojrStringBytes(value).length;",
    "  if (Array.isArray(value) || ArrayBuffer.isView(value)) return value.length;",
    "  if (value.__gojrByteView === true) return value.length;",
    "  if (value instanceof Map) return value.size;",
    "  if (value.__gojrChannel === true) return value.queue.length;",
    "  return 0;",
    "}",
    "function __gojrCap(value) {",
    "  if (value == null) return 0;",
    "  if (typeof value.__gojrCap === \"number\") return value.__gojrCap;",
    "  if (Array.isArray(value) || ArrayBuffer.isView(value)) return value.length;",
    "  if (value.__gojrByteView === true) return value.length;",
    "  if (value.__gojrChannel === true) return value.capacity;",
    "  return __gojrLen(value);",
    "}",
    "function __gojrSetCap(value, capacity) {",
    "  try { Object.defineProperty(value, \"__gojrCap\", { value: Math.max(0, Number(capacity)), configurable: true }); } catch { }",
    "  return value;",
    "}",
    "function __gojrArrayLiteral(typeName, elemType, entries) {",
    "  const resolved = [];",
    "  let nextIndex = 0;",
    "  let length = __gojrFixedArrayLength(typeName) ?? 0;",
    "  for (const [rawKey, value] of entries) {",
    "    const index = rawKey === null || rawKey === undefined ? nextIndex : Number(rawKey);",
    "    resolved.push([index, value]);",
    "    nextIndex = index + 1;",
    "    if (index + 1 > length) length = index + 1;",
    "  }",
    "  if (__gojrIsByteSliceTypeName(typeName) || (__gojrFixedByteArrayLength(typeName) !== undefined && (elemType === \"byte\" || elemType === \"uint8\"))) {",
    "    const out = new Uint8Array(length);",
    "    for (const [index, value] of resolved) out[index] = Number(value) & 255;",
    "    return __gojrSetCap(out, length);",
    "  }",
    "  const out = Array.from({ length }, () => __gojrZero(elemType));",
    "  for (const [index, value] of resolved) out[index] = value;",
    "  return __gojrSetCap(out, length);",
    "}",
    "function __gojrConvertArray(typeName, elemType, value) {",
    "  const length = __gojrFixedArrayLength(typeName) ?? __gojrLen(value);",
    "  if (__gojrIsByteSliceTypeName(`[]${elemType}`) || elemType === \"byte\" || elemType === \"uint8\") {",
    "    const out = new Uint8Array(length);",
    "    const source = value == null ? [] : value;",
    "    for (let index = 0; index < length && index < __gojrLen(source); index += 1) out[index] = Number(source[index]) & 255;",
    "    return __gojrSetCap(out, length);",
    "  }",
    "  const out = Array.from({ length }, () => __gojrZero(elemType));",
    "  const source = value == null ? [] : value;",
    "  for (let index = 0; index < length && index < __gojrLen(source); index += 1) out[index] = __gojrConvertArrayElement(elemType, source[index]);",
    "  return __gojrSetCap(out, length);",
    "}",
    "function __gojrConvertArrayElement(elemType, value) {",
    "  if (__gojrIntegerTypeName(elemType) && value !== null && value !== undefined) return __gojrIntegerConvert(elemType, value);",
    "  return value;",
    "}",
    "function __gojrIntegerTypeName(typeName) {",
    "  switch (String(typeName || \"\")) {",
    "    case \"byte\": case \"rune\": case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return true;",
    "    default: return false;",
    "  }",
    "}",
    "function __gojrIntegerFrom(value) {",
    "  if (typeof value === \"bigint\") return value;",
    "  if (typeof value === \"number\") return BigInt(Math.trunc(value));",
    "  if (__gojrPointerLike(value)) return value;",
    "  try { return BigInt(value); } catch (error) {",
    "    const typeName = value && typeof value === \"object\" ? (value.__gojrType || value.constructor && value.constructor.name || \"object\") : typeof value;",
    "    throw new TypeError(`cannot convert ${String(value)} (${typeName}) to Go integer`);",
    "  }",
    "}",
    "function __gojrIntegerConvert(typeName, value) {",
    "  const raw = __gojrIntegerFrom(value);",
    "  if (__gojrPointerLike(raw)) return raw;",
    "  switch (String(typeName || \"\").trim()) {",
    "    case \"uint8\": case \"byte\": return BigInt.asUintN(8, raw);",
    "    case \"int8\": return BigInt.asIntN(8, raw);",
    "    case \"uint16\": return BigInt.asUintN(16, raw);",
    "    case \"int16\": return BigInt.asIntN(16, raw);",
    "    case \"uint32\": return BigInt.asUintN(32, raw);",
    "    case \"int32\": case \"rune\": return BigInt.asIntN(32, raw);",
    "    case \"uint\": case \"uint64\": case \"uintptr\": return BigInt.asUintN(64, raw);",
    "    case \"int\": case \"int64\": return BigInt.asIntN(64, raw);",
    "    default: return raw;",
    "  }",
    "}",
    "function __gojrBool(value) {",
    "  value = __gojrUnwrapInterface(value);",
    "  if (typeof value !== \"boolean\") throw new TypeError(`GOJR_RUNTIME001: ${String(value)} is not a bool`);",
    "  return value;",
    "}",
    "function __gojrIsComplexValue(value) {",
    "  return Boolean(value && typeof value === \"object\" && \"real\" in value && \"imag\" in value);",
    "}",
    "function __gojrToFloat(value) {",
    "  value = __gojrUnwrapInterface(value);",
    "  if (typeof value === \"number\") return value;",
    "  if (typeof value === \"bigint\") return Number(value);",
    "  throw new TypeError(`GOJR_RUNTIME001: ${String(value)} is not numeric`);",
    "}",
    "function __gojrToBigInt(value) {",
    "  value = __gojrUnwrapInterface(value);",
    "  if (typeof value === \"bigint\") return value;",
    "  if (typeof value === \"number\" && Number.isInteger(value)) return BigInt(value);",
    "  throw new TypeError(`GOJR_RUNTIME001: ${String(value)} is not an integer`);",
    "}",
    "function __gojrWrapNumericResult(value, typeName) {",
    "  const type = String(typeName || \"\").trim();",
    "  if (__gojrIntegerTypeName(type)) return __gojrIntegerConvert(type, value);",
    "  if (type === \"float32\") return Math.fround(Number(value));",
    "  if (type === \"float64\") return Number(value);",
    "  if (type === \"complex64\" || type === \"complex128\") return __gojrToComplex(value);",
    "  return value;",
    "}",
    "function __gojrNumericIdentity(value) {",
    "  value = __gojrUnwrapInterface(value);",
    "  if (__gojrIsComplexValue(value) || typeof value === \"bigint\" || typeof value === \"number\") return value;",
    "  throw new TypeError(`GOJR_RUNTIME001: ${String(value)} is not numeric`);",
    "}",
    "function __gojrNegate(value, typeName = \"\") {",
    "  value = __gojrUnwrapInterface(value);",
    "  if (__gojrIsComplexValue(value)) return __gojrComplex(-value.real, -value.imag);",
    "  if (typeof value === \"bigint\") return __gojrWrapNumericResult(-value, typeName);",
    "  return __gojrWrapNumericResult(-__gojrToFloat(value), typeName);",
    "}",
    "function __gojrAdd(left, right, typeName = \"\") {",
    "  left = __gojrUnwrapInterface(left); right = __gojrUnwrapInterface(right);",
    "  if (__gojrIsString(left) || __gojrIsString(right)) {",
    "    if (__gojrIsString(left) && __gojrIsString(right)) return __gojrConcatStrings(left, right);",
    "    throw new TypeError(`GOJR_RUNTIME001: invalid operation: ${typeof left} + ${typeof right}`);",
    "  }",
    "  if (__gojrIsComplexValue(left) || __gojrIsComplexValue(right)) return __gojrComplexAdd(left, right);",
    "  if (typeof left === \"bigint\" && typeof right === \"bigint\") return __gojrWrapNumericResult(left + right, typeName);",
    "  return __gojrWrapNumericResult(__gojrToFloat(left) + __gojrToFloat(right), typeName);",
    "}",
    "function __gojrSub(left, right, typeName = \"\") {",
    "  left = __gojrUnwrapInterface(left); right = __gojrUnwrapInterface(right);",
    "  if (__gojrIsComplexValue(left) || __gojrIsComplexValue(right)) return __gojrComplexSub(left, right);",
    "  if (typeof left === \"bigint\" && typeof right === \"bigint\") return __gojrWrapNumericResult(left - right, typeName);",
    "  return __gojrWrapNumericResult(__gojrToFloat(left) - __gojrToFloat(right), typeName);",
    "}",
    "function __gojrMul(left, right, typeName = \"\") {",
    "  left = __gojrUnwrapInterface(left); right = __gojrUnwrapInterface(right);",
    "  if (__gojrIsComplexValue(left) || __gojrIsComplexValue(right)) return __gojrComplexMul(left, right);",
    "  if (typeof left === \"bigint\" && typeof right === \"bigint\") return __gojrWrapNumericResult(left * right, typeName);",
    "  return __gojrWrapNumericResult(__gojrToFloat(left) * __gojrToFloat(right), typeName);",
    "}",
    "function __gojrDiv(left, right, typeName = \"\") {",
    "  left = __gojrUnwrapInterface(left); right = __gojrUnwrapInterface(right);",
    "  if (__gojrIsComplexValue(left) || __gojrIsComplexValue(right)) return __gojrComplexDiv(left, right);",
    "  if (typeof left === \"bigint\" && typeof right === \"bigint\") return __gojrWrapNumericResult(left / right, typeName);",
    "  return __gojrWrapNumericResult(__gojrToFloat(left) / __gojrToFloat(right), typeName);",
    "}",
    "function __gojrMod(left, right, typeName = \"\") {",
    "  const result = __gojrToBigInt(left) % __gojrToBigInt(right);",
    "  return __gojrWrapNumericResult(result, typeName);",
    "}",
    "function __gojrBitwiseAnd(left, right, typeName = \"\") { return __gojrWrapNumericResult(__gojrToBigInt(left) & __gojrToBigInt(right), typeName); }",
    "function __gojrBitwiseOr(left, right, typeName = \"\") { return __gojrWrapNumericResult(__gojrToBigInt(left) | __gojrToBigInt(right), typeName); }",
    "function __gojrBitwiseXor(left, right, typeName = \"\") {",
    "  if (__gojrPointerLike(left) && __gojrIsZeroInteger(right)) return left;",
    "  if (__gojrPointerLike(right) && __gojrIsZeroInteger(left)) return right;",
    "  return __gojrWrapNumericResult(__gojrToBigInt(left) ^ __gojrToBigInt(right), typeName);",
    "}",
    "function __gojrBitClear(left, right, typeName = \"\") { return __gojrWrapNumericResult(__gojrToBigInt(left) & ~__gojrToBigInt(right), typeName); }",
    "function __gojrShiftLeft(left, right, typeName = \"\") {",
    "  const shift = __gojrToBigInt(right);",
    "  if (shift < 0n) throw new TypeError(\"GOJR_RUNTIME001: negative shift count\");",
    "  return __gojrWrapNumericResult(__gojrToBigInt(left) << shift, typeName);",
    "}",
    "function __gojrShiftRight(left, right, typeName = \"\") {",
    "  const shift = __gojrToBigInt(right);",
    "  if (shift < 0n) throw new TypeError(\"GOJR_RUNTIME001: negative shift count\");",
    "  return __gojrWrapNumericResult(__gojrToBigInt(left) >> shift, typeName);",
    "}",
    "function __gojrCompareValues(left, right) {",
    "  left = __gojrUnwrapInterface(left); right = __gojrUnwrapInterface(right);",
    "  if ((typeof left === \"bigint\" || typeof left === \"number\") && (typeof right === \"bigint\" || typeof right === \"number\")) {",
    "    const a = typeof left === \"bigint\" && typeof right === \"bigint\" ? left : Number(left);",
    "    const b = typeof left === \"bigint\" && typeof right === \"bigint\" ? right : Number(right);",
    "    if (Number.isNaN(a) || Number.isNaN(b)) return NaN;",
    "    return a === b ? 0 : a < b ? -1 : 1;",
    "  }",
    "  if (__gojrIsString(left) && __gojrIsString(right)) return __gojrByteSequenceCompare(__gojrStringBytes(left), __gojrStringBytes(right));",
    "  throw new TypeError(`GOJR_RUNTIME001: cannot compare ${String(left)} and ${String(right)}`);",
    "}",
    "function __gojrCompareOp(left, right, op) {",
    "  const compared = __gojrCompareValues(left, right);",
    "  switch (op) {",
    "    case \"<\": return compared < 0;",
    "    case \"<=\": return compared <= 0;",
    "    case \">\": return compared > 0;",
    "    case \">=\": return compared >= 0;",
    "    default: throw new TypeError(`GOJR_RUNTIME001: unsupported comparison ${op}`);",
    "  }",
    "}",
    "function __gojrUnderlyingRuntimeTypeName(typeName, importsByPath = __gojrActiveImportsByPath, pkgPath = undefined) {",
    "  let current = String(typeName || \"any\").trim();",
    "  const seen = new Set();",
    "  while (current && !seen.has(current)) {",
    "    seen.add(current);",
    "    const descriptor = __gojrDescriptorForTypeName(current, importsByPath, pkgPath);",
    "    if (!descriptor || !descriptor.underlying || descriptor.underlying === current) break;",
    "    pkgPath = descriptor.pkgPath || pkgPath;",
    "    current = String(descriptor.underlying).trim();",
    "  }",
    "  return current || \"any\";",
    "}",
    "function __gojrConvertDynamicType(typeName, value, importsByPath = __gojrActiveImportsByPath, pkgPath = undefined) {",
    "  const text = String(typeName || \"any\").trim();",
    "  const resolved = __gojrUnderlyingRuntimeTypeName(text, importsByPath, pkgPath);",
    "  if (__gojrIsByteSliceTypeName(resolved)) return __gojrBytesFrom(value);",
    "  if (__gojrIntegerTypeName(resolved)) return __gojrIntegerConvert(resolved, value);",
    "  if (resolved === \"float32\" || resolved === \"float64\") return Number(value);",
    "  if (resolved === \"complex64\" || resolved === \"complex128\") return __gojrToComplex(value);",
    "  if (resolved === \"bool\") return Boolean(value);",
    "  if (resolved === \"string\") return __gojrStringFrom(value);",
    "  if (resolved === \"unsafe.Pointer\") return __gojrConvertPointer(text, value, pkgPath);",
    "  if (resolved.startsWith(\"*\")) return __gojrConvertPointer(text, value, pkgPath);",
    "  if (resolved.startsWith(\"[]\")) return __gojrConvertSlice(text, value);",
    "  if (resolved.startsWith(\"map[\") || resolved.startsWith(\"chan \") || resolved.startsWith(\"<-chan\") || resolved.startsWith(\"chan<-\") || resolved.startsWith(\"func(\")) return __gojrConvertNilable(text, value);",
    "  return value;",
    "}",
    "function __gojrSizeofType(typeName, importsByPath = __gojrActiveImportsByPath) {",
    "  const text = String(typeName || \"any\").trim();",
    "  const resolved = __gojrUnderlyingRuntimeTypeName(text, importsByPath);",
    "  switch (resolved) {",
    "    case \"bool\": case \"int8\": case \"uint8\": case \"byte\": return 1n;",
    "    case \"int16\": case \"uint16\": return 2n;",
    "    case \"int32\": case \"uint32\": case \"rune\": case \"float32\": return 4n;",
    "    case \"int\": case \"uint\": case \"int64\": case \"uint64\": case \"uintptr\": case \"float64\": return 8n;",
    "    case \"complex64\": return 8n;",
    "    case \"complex128\": return 16n;",
    "    case \"string\": return 16n;",
    "  }",
    "  if (resolved.startsWith(\"*\") || resolved.startsWith(\"map[\") || resolved.startsWith(\"chan \") || resolved.startsWith(\"<-chan\") || resolved.startsWith(\"chan<-\") || resolved.startsWith(\"func(\")) return 8n;",
    "  if (resolved.startsWith(\"[]\")) return 24n;",
    "  const array = /^\\[(\\d+)\\](.+)$/.exec(resolved);",
    "  if (array) return BigInt(Number(array[1])) * __gojrSizeofType(array[2], importsByPath);",
    "  const descriptor = __gojrDescriptorForTypeName(text, importsByPath);",
    "  if (descriptor && descriptor.kind === \"struct\") {",
    "    let size = 0n;",
    "    for (const field of descriptor.fields || []) size += __gojrSizeofType(field.type, importsByPath);",
    "    return size;",
    "  }",
    "  return 8n;",
    "}",
    "function __gojrSizeofValue(value) {",
    "  if (value && value.__gojrInterface === true) value = value.value;",
    "  if (value === null || value === undefined) return 0n;",
    "  const typeName = value && (value.__gojrType || value.__gojrElemType);",
    "  if (typeName) return __gojrSizeofType(typeName);",
    "  if (typeof value === \"boolean\") return 1n;",
    "  if (typeof value === \"bigint\" || typeof value === \"number\") return 8n;",
    "  if (__gojrIsString(value)) return 16n;",
    "  if (value && typeof value === \"object\" && \"real\" in value && \"imag\" in value) return 16n;",
    "  if (Array.isArray(value) || ArrayBuffer.isView(value)) return 24n;",
    "  return 8n;",
    "}",
    "function __gojrAlignofType(typeName, importsByPath = __gojrActiveImportsByPath) {",
    "  const size = __gojrSizeofType(typeName, importsByPath);",
    "  if (size <= 1n) return 1n;",
    "  if (size <= 2n) return 2n;",
    "  if (size <= 4n) return 4n;",
    "  return 8n;",
    "}",
    "function __gojrAlignofValue(value) {",
    "  const size = __gojrSizeofValue(value);",
    "  if (size <= 1n) return 1n;",
    "  if (size <= 2n) return 2n;",
    "  if (size <= 4n) return 4n;",
    "  return 8n;",
    "}",
    "function __gojrOffsetof(object, field) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  const descriptor = __gojrDescriptorForTypeName(__gojrRuntimeTypeName(target), __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(target));",
    "  let offset = 0n;",
    "  for (const item of descriptor && descriptor.fields || []) {",
    "    if (item.name === field) return offset;",
    "    offset += __gojrSizeofType(item.type, __gojrActiveImportsByPath);",
    "  }",
    "  return 0n;",
    "}",
    "function __gojrFixedArrayLength(typeName) {",
    "  const match = /^\\[(\\d+)\\]/.exec(String(typeName || \"\"));",
    "  if (!match) return undefined;",
    "  const length = Number(match[1]);",
    "  return Number.isSafeInteger(length) ? length : undefined;",
    "}",
    "function __gojrFixedByteArrayLength(typeName) {",
    "  const match = /^\\[(\\d+)\\](?:byte|uint8)$/.exec(String(typeName || \"\").trim());",
    "  if (!match) return undefined;",
    "  const length = Number(match[1]);",
    "  return Number.isSafeInteger(length) ? length : undefined;",
    "}",
    "function __gojrSlice(object, start, end, max) {",
    "  if (object && object.__gojrTypedNil === true) object = [];",
    "  object = __gojrDerefIfPointer(object);",
    "  const length = __gojrLen(object);",
    "  const low = start === null || start === undefined ? 0 : Number(start);",
    "  const high = end === null || end === undefined ? length : Number(end);",
    "  const capHigh = max === null || max === undefined ? __gojrCap(object) : Number(max);",
    "  const target = object ?? [];",
    "  if (target && target.__gojrByteView === true) return __gojrSetCap(target.subarray(low, high), capHigh - low);",
    "  if (__gojrIsString(target)) return __gojrStringFromBytes(__gojrStringBytes(target).slice(low, high));",
    "  const sliced = typeof target === \"string\" ? target.slice(low, high) : (ArrayBuffer.isView(target) && typeof target.subarray === \"function\" ? target.subarray(low, high) : (typeof target.slice === \"function\" ? target.slice(low, high) : []));",
    "  return typeof sliced === \"string\" ? sliced : __gojrSetCap(sliced, capHigh - low);",
    "}",
    "function __gojrConvertSlice(typeName, value) {",
    "  if (value === null || value === undefined) return __gojrTypedNil(typeName);",
    "  if (value && value.__gojrTypedNil === true) return __gojrTypedNil(typeName);",
    "  return value;",
    "}",
    "function __gojrConvertNilable(typeName, value) {",
    "  if (value === null || value === undefined) return __gojrTypedNil(typeName);",
    "  if (value && value.__gojrTypedNil === true) return __gojrTypedNil(typeName);",
    "  return value;",
    "}",
    "function __gojrMake(typeName, length = 0, capacity = length) {",
    "  const text = String(typeName || \"\").trim();",
    "  if (text.startsWith(\"map[\")) { const kv = __gojrMapKeyValueTypes(text); return __gojrMap(new Map(), __gojrZero(kv && kv.valueType || \"any\"), kv && kv.keyType || \"any\", kv && kv.valueType || \"any\"); }",
    "  if (text.startsWith(\"chan \")) return __gojrMakeChan(text.slice(5).trim(), capacity);",
    "  if (text.startsWith(\"chan<-\")) return __gojrMakeChan(text.slice(6).trim(), capacity);",
    "  if (text.startsWith(\"<-chan\")) return __gojrMakeChan(text.slice(6).trim(), capacity);",
    "  const elem = __gojrSliceElementType(text);",
    "  if (elem !== undefined) return elem === \"byte\" || elem === \"uint8\" ? __gojrSetCap(new Uint8Array(length), capacity) : __gojrSetCap(Array.from({ length }, () => __gojrZero(elem)), capacity);",
    "  const descriptor = __gojrDescriptorForTypeName(text);",
    "  if (descriptor && descriptor.kind === \"slice\") return descriptor.elem === \"byte\" || descriptor.elem === \"uint8\" ? __gojrSetCap(new Uint8Array(length), capacity) : __gojrSetCap(Array.from({ length }, () => __gojrZero(descriptor.elem || \"any\")), capacity);",
    "  if (descriptor && descriptor.kind === \"map\") return __gojrMap(new Map(), __gojrZero(descriptor.elem || \"any\"), descriptor.key || \"any\", descriptor.elem || \"any\");",
    "  if (descriptor && descriptor.kind === \"chan\") return __gojrMakeChan(descriptor.elem || \"any\", capacity);",
    "  throw new Error(`GOJR_RUNTIME001: cannot make ${text}`);",
    "}",
    "function __gojrSliceElementType(typeName) {",
    "  return String(typeName || \"\").startsWith(\"[]\") ? String(typeName).slice(2).trim() : undefined;",
    "}",
    "function __gojrMapValueType(typeName) {",
    "  const mapType = __gojrMapKeyValueTypes(typeName);",
    "  return mapType ? mapType.valueType : undefined;",
    "}",
    "function __gojrMapKeyValueTypes(typeName) {",
    "  const text = String(typeName || \"\");",
    "  if (!text.startsWith(\"map[\")) return undefined;",
    "  let depth = 0;",
    "  for (let index = 4; index < text.length; index += 1) {",
    "    const char = text[index];",
    "    if (char === \"[\") depth += 1;",
    "    if (char === \"]\") {",
    "      if (depth === 0) return { keyType: text.slice(4, index).trim(), valueType: text.slice(index + 1).trim() };",
    "      depth -= 1;",
    "    }",
    "  }",
    "  return undefined;",
    "}",
    "function __gojrAppend(slice, ...values) {",
    "  if (slice instanceof Uint8Array) {",
    "    const out = new Uint8Array(slice.length + values.length);",
    "    out.set(slice, 0);",
    "    for (let index = 0; index < values.length; index += 1) out[slice.length + index] = Number(values[index]) & 255;",
    "    return out;",
    "  }",
    "  if (ArrayBuffer.isView(slice)) {",
    "    return Array.from(slice).concat(values);",
    "  }",
    "  return (Array.isArray(slice) ? slice : []).concat(values);",
    "}",
    "function __gojrCopy(dst, src) {",
    "  if (__gojrIsString(src)) src = __gojrRawBytesFromGoString(src);",
    "  const count = Math.min(__gojrLen(dst), __gojrLen(src));",
    "  if (dst instanceof Uint8Array) {",
    "    const view = src instanceof Uint8Array ? src : Uint8Array.from(Array.from(src ?? [], (value) => Number(value) & 255));",
    "    dst.set(view.subarray(0, count), 0);",
    "    return count;",
    "  }",
    "  for (let index = 0; index < count; index += 1) dst[index] = src[index];",
    "  return count;",
    "}",
    "function __gojrDelete(map, key) {",
    "  if (map instanceof Map && map.__gojrGoMap === true) Map.prototype.delete.call(map, __gojrMapKeyId(key));",
    "  else if (map instanceof Map) map.delete(key);",
    "  return null;",
    "}",
    "function __gojrClear(value) {",
    "  if (value instanceof Map) { value.clear(); return null; }",
    "  const target = __gojrDerefIfPointer(value);",
    "  if (target instanceof Map) { target.clear(); return null; }",
    "  if (target instanceof Uint8Array) { target.fill(0); return null; }",
    "  if (ArrayBuffer.isView(target) && typeof target.fill === \"function\") { target.fill(0); return null; }",
    "  if (Array.isArray(target)) {",
    "    for (let index = 0; index < target.length; index += 1) target[index] = __gojrZeroLike(target[index]);",
    "  }",
    "  return null;",
    "}",
    "function __gojrZeroLike(value) {",
    "  if (typeof value === \"bigint\") return 0n;",
    "  if (typeof value === \"number\") return 0;",
    "  if (typeof value === \"string\") return \"\";",
    "  if (typeof value === \"boolean\") return false;",
    "  if (value && typeof value === \"object\" && typeof value.__gojrType === \"string\") return __gojrZero(value.__gojrType, value.__gojrPkgPath);",
    "  return null;",
    "}",
    "function __gojrMin(first, ...rest) {",
    "  let best = first;",
    "  for (const value of rest) if (__gojrCompareValues(value, best) < 0) best = value;",
    "  return best;",
    "}",
    "function __gojrMax(first, ...rest) {",
    "  let best = first;",
    "  for (const value of rest) if (__gojrCompareValues(value, best) > 0) best = value;",
    "  return best;",
    "}",
    "class __GoJrPanic extends Error {",
    "  constructor(value) {",
    "    super(`panic: ${String(value)}`);",
    "    this.name = \"GoJuniorPanic\";",
    "    this.value = value;",
    "    this.__gojrPanic = true;",
    "  }",
    "}",
    "function __gojrPanic(value) {",
    "  throw value && value.__gojrPanic === true ? value : new __GoJrPanic(value);",
    "}",
    "function __gojrPrint(stdout, ...values) {",
    "  stdout(values.map((value) => String(value)).join(\"\"));",
    "  return null;",
    "}",
    "function __gojrPrintln(stdout, ...values) {",
    "  stdout(`${values.map((value) => String(value)).join(\" \")}\\n`);",
    "  return null;",
    "}",
    "function __gojrIndex(object, index, integerResult) {",
    "  object = __gojrDerefIfPointer(object);",
    "  if (__gojrIsString(object)) {",
    "    const value = __gojrStringByteAt(object, Number(index));",
    "    return integerResult ? BigInt(value) : value;",
    "  }",
    "  const value = object[index];",
    "  if (integerResult || object instanceof Uint8Array || object instanceof Uint8ClampedArray || object instanceof Int8Array || object instanceof Uint16Array || object instanceof Int16Array || object instanceof Uint32Array || object instanceof Int32Array || object instanceof BigInt64Array || object instanceof BigUint64Array) return BigInt(value);",
    "  return value;",
    "}",
    "function __gojrSetIndex(object, index, value) {",
    "  object = __gojrDerefIfPointer(object);",
    "  const numericIndex = Number(index);",
    "  if (object && object.__gojrByteView === true) {",
    "    object[numericIndex] = value;",
    "    return null;",
    "  }",
    "  if (object instanceof Uint8Array || object instanceof Uint8ClampedArray || object instanceof Int8Array || object instanceof Uint16Array || object instanceof Int16Array || object instanceof Uint32Array || object instanceof Int32Array) {",
    "    object[numericIndex] = Number(value);",
    "    return null;",
    "  }",
    "  if (object instanceof BigInt64Array || object instanceof BigUint64Array) {",
    "    object[numericIndex] = BigInt(value);",
    "    return null;",
    "  }",
    "  object[numericIndex] = value;",
    "  return null;",
    "}",
    "function __gojrSelectPackageField(importsByPath, path, field) {",
    "  const overridingBuiltin = __gojrBuiltinImportOverrides(path) ? __gojrBuiltinImport(path, importsByPath) : undefined;",
    "  if (overridingBuiltin && typeof overridingBuiltin === \"object\" && field in overridingBuiltin) return overridingBuiltin[field];",
    "  const imported = importsByPath[path];",
    "  const packageObject = imported && typeof imported === \"object\" && \"package\" in imported ? imported.package : imported;",
    "  if (packageObject && typeof packageObject === \"object\" && field in packageObject) return packageObject[field];",
    "  const builtin = __gojrBuiltinImport(path, importsByPath);",
    "  if (builtin && typeof builtin === \"object\" && field in builtin) return builtin[field];",
    "  return packageObject && typeof packageObject === \"object\" ? packageObject[field] : undefined;",
    "}",
    "function __gojrSelectPackageCallableField(importsByPath, path, field) {",
    "  const selected = __gojrSelectPackageField(importsByPath, path, field);",
    "  if (typeof selected === \"function\") return selected;",
    "  const builtin = __gojrBuiltinImport(path, importsByPath);",
    "  if (builtin && typeof builtin[field] === \"function\") return builtin[field];",
    "  throw new TypeError(`${path}.${field} is not a function`);",
    "}",
    "function __gojrMergedBuiltinPackage(builtin, packageObject) {",
    "  if (!packageObject || typeof packageObject !== \"object\") return builtin || {};",
    "  if (!builtin || typeof builtin !== \"object\") return packageObject;",
    "  const merged = Object.assign(Object.create(null), packageObject, builtin);",
    "  for (const key of [\"__gojrExportIndex\", \"__gojrTypeDescriptors\", \"__gojrImportPath\", \"__gojrPackageName\", \"__gojrImportPathsByQualifier\"]) {",
    "    if (Object.prototype.hasOwnProperty.call(packageObject, key) || packageObject[key] !== undefined) {",
    "      Object.defineProperty(merged, key, { value: packageObject[key] });",
    "    }",
    "  }",
    "  return merged;",
    "}",
    "function __gojrBuiltinImport(path, importsByPath) {",
    "  if (path === \"internal/bytealg\") return __gojrBytealgPackage();",
    "  if (path === \"internal/reflectlite\") return __gojrReflectlitePackage(importsByPath);",
    "  if (path === \"io/fs\") return __gojrFsPackage();",
    "  if (path === \"os\") return __gojrOsPackage();",
    "  if (path === \"reflect\") return __gojrReflectPackage(importsByPath);",
    "  if (path === \"runtime\") return __gojrRuntimePackage();",
    "  if (path === \"internal/sync\") return __gojrInternalSyncPackage();",
    "  if (path === \"sync\") return __gojrSyncPackage();",
    "  if (path === \"sync/atomic\") return __gojrAtomicPackage();",
    "  if (path === \"unsafe\") return __gojrUnsafePackage();",
    "  if (path === \"weak\") return __gojrWeakPackage();",
    "  return undefined;",
    "}",
    `const __gojrGeneratedBuiltinImportPaths = new Set(${generatedBuiltinImports});`,
    "function __gojrBuiltinImportOverrides(path) {",
    "  return __gojrGeneratedBuiltinImportPaths.has(path);",
    "}",
    "function __gojrBytealgPackage() {",
    "  return {",
    "    IndexByte: (b, c) => BigInt(__gojrByteIndex(__gojrBytesFrom(b), Number(c) & 255)),",
    "    IndexByteString: (s, c) => BigInt(__gojrByteIndex(__gojrBytesFrom(s), Number(c) & 255)),",
    "    Index: (a, b) => BigInt(__gojrByteSequenceIndex(__gojrBytesFrom(a), __gojrBytesFrom(b))),",
    "    IndexString: (a, b) => BigInt(__gojrByteSequenceIndex(__gojrBytesFrom(a), __gojrBytesFrom(b))),",
    "    Count: (b, c) => BigInt(__gojrByteCount(__gojrBytesFrom(b), Number(c) & 255)),",
    "    CountString: (s, c) => BigInt(__gojrByteCount(__gojrBytesFrom(s), Number(c) & 255)),",
    "    Compare: (a, b) => BigInt(__gojrByteSequenceCompare(__gojrBytesFrom(a), __gojrBytesFrom(b))),",
    "    CompareString: (a, b) => BigInt(__gojrByteSequenceCompare(__gojrBytesFrom(a), __gojrBytesFrom(b))),",
    "    Equal: (a, b) => __gojrByteSequenceCompare(__gojrBytesFrom(a), __gojrBytesFrom(b)) === 0,",
    "    MakeNoZero: (n) => __gojrSetCap(new Uint8Array(Number(n)), Number(n)),",
    "    abigen_runtime_cmpstring: (a, b) => BigInt(__gojrByteSequenceCompare(__gojrBytesFrom(a), __gojrBytesFrom(b))),",
    "    abigen_runtime_memequal: (a, b, size) => Number(size) === 0 || __gojrEqual(a, b),",
    "    abigen_runtime_memequal_varlen: (a, b) => __gojrEqual(a, b)",
    "  };",
    "}",
    "function __gojrByteIndex(values, needle) {",
    "  for (let index = 0; index < values.length; index += 1) if (values[index] === needle) return index;",
    "  return -1;",
    "}",
    "function __gojrByteCount(values, needle) {",
    "  let count = 0;",
    "  for (const value of values) if (value === needle) count += 1;",
    "  return count;",
    "}",
    "function __gojrByteSequenceIndex(values, needle) {",
    "  if (needle.length === 0) return 0;",
    "  if (needle.length > values.length) return -1;",
    "  const lastStart = values.length - needle.length;",
    "  for (let start = 0; start <= lastStart; start += 1) {",
    "    let matched = true;",
    "    for (let offset = 0; offset < needle.length; offset += 1) {",
    "      if (values[start + offset] !== needle[offset]) { matched = false; break; }",
    "    }",
    "    if (matched) return start;",
    "  }",
    "  return -1;",
    "}",
    "function __gojrByteSequenceCompare(left, right) {",
    "  const length = Math.min(left.length, right.length);",
    "  for (let index = 0; index < length; index += 1) {",
    "    if (left[index] !== right[index]) return left[index] < right[index] ? -1 : 1;",
    "  }",
    "  if (left.length === right.length) return 0;",
    "  return left.length < right.length ? -1 : 1;",
    "}",
    "function __gojrUnsafePackage() {",
    "  return {",
    "    Add: (ptr, offset) => __gojrUnsafeAdd(ptr, offset),",
    "    Alignof: (value) => __gojrAlignofValue(value),",
    "    Offsetof: (_value) => 0n,",
    "    Sizeof: (value) => __gojrSizeofValue(value),",
    "    Slice: (ptr, len) => {",
    "      const length = Number(len || 0);",
    "      if (__gojrNilUnsafePointer(ptr)) {",
    "        if (length === 0) return null;",
    "        throw new Error(\"GOJR_RUNTIME001: unsafe.Slice: ptr is nil and len is not zero\");",
    "      }",
    "      const sequence = __gojrPointerSequence(ptr);",
    "      if (sequence) return __gojrSliceFromSequence(sequence, length);",
    "      if (ptr instanceof Uint8Array) return ptr.subarray(0, length);",
    "      if (ArrayBuffer.isView(ptr)) return Array.from(ptr).slice(0, length);",
    "      if (Array.isArray(ptr)) return ptr.slice(0, length);",
    "      if (ptr && ptr.__gojrPointer === true) return __gojrSetCap(Array.from({ length }, (_item, index) => index === 0 ? __gojrDeref(ptr) : 0n), length);",
    "      return Array.from({ length }, () => 0n);",
    "    },",
    "    SliceData: (slice) => {",
    "      const target = __gojrDerefIfPointer(slice);",
    "      if (target == null || __gojrLen(target) === 0) return null;",
    "      return __gojrPointer(__gojrSliceElementRuntimeType(target), () => target[0], (next) => { target[0] = next; }, gojrPackageArtifact.importPath, { values: target, index: 0 });",
    "    },",
    "    String: (ptr, len) => {",
    "      const length = Number(len || 0);",
    "      if (__gojrNilUnsafePointer(ptr)) {",
    "        if (length === 0) return \"\";",
    "        throw new Error(\"GOJR_RUNTIME001: unsafe.String: ptr is nil and len is not zero\");",
    "      }",
    "      const sequence = __gojrPointerSequence(ptr);",
    "      if (sequence) return new TextDecoder().decode(__gojrBytesFromSequence(sequence, length));",
    "      if (ptr instanceof Uint8Array) return new TextDecoder().decode(ptr.subarray(0, length));",
    "      if (ArrayBuffer.isView(ptr)) return new TextDecoder().decode(__gojrBytesFromSequence({ values: ptr, index: 0 }, length));",
    "      if (Array.isArray(ptr)) return new TextDecoder().decode(Uint8Array.from(ptr.slice(0, length), (item) => Number(item) & 255));",
    "      if (ptr && ptr.__gojrPointer === true) return new TextDecoder().decode(Uint8Array.from([Number(__gojrDeref(ptr)) & 255]).subarray(0, length));",
    "      return \"\";",
    "    },",
    "    StringData: (value) => {",
    "      const bytes = __gojrStringBytes(value);",
    "      if (bytes.length === 0) return null;",
    "      return __gojrPointer(\"byte\", () => bytes[0], (next) => { bytes[0] = Number(next) & 255; }, gojrPackageArtifact.importPath, { values: bytes, index: 0 });",
    "    }",
    "  };",
    "}",
    "function __gojrSyncPackage() {",
    "  return {",
    "    __gojrTypeDescriptors: __gojrSyncDescriptors(),",
    "    NewCond: async (locker) => { const cond = __gojrSyncZero(\"sync.Cond\"); cond.L = locker ?? null; return __gojrPointerValue(\"sync.Cond\", cond, \"sync\"); },",
    "    OnceFunc: async (fn) => { const once = __gojrSyncZero(\"sync.Once\"); return async () => await __gojrSyncOnceDo(once, fn); },",
    "    OnceValue: async (fn) => { const once = __gojrSyncZero(\"sync.Once\"); let value; return async () => { await __gojrSyncOnceDo(once, async () => { value = await fn(); }); return value; }; },",
    "    OnceValues: async (fn) => { const once = __gojrSyncZero(\"sync.Once\"); let values = __gojrTuple([null, null]); return async () => { await __gojrSyncOnceDo(once, async () => { values = __gojrTupleValues(await fn()); }); return __gojrTuple(values); }; },",
    "    \"Mutex.Lock\": async () => null, \"Mutex.TryLock\": async () => true, \"Mutex.Unlock\": async () => null,",
    "    \"RWMutex.Lock\": async () => null, \"RWMutex.TryLock\": async () => true, \"RWMutex.Unlock\": async () => null,",
    "    \"RWMutex.RLock\": async () => null, \"RWMutex.TryRLock\": async () => true, \"RWMutex.RUnlock\": async () => null,",
    "    \"RWMutex.RLocker\": async () => __gojrSyncLocker(),",
    "    \"Once.Do\": async (receiver, fn) => __gojrSyncOnceDo(receiver, fn),",
    "    \"Pool.Put\": async (receiver, value) => __gojrSyncPoolPut(receiver, value),",
    "    \"Pool.Get\": async (receiver) => __gojrSyncPoolGet(receiver),",
    "    \"WaitGroup.Add\": async (receiver, delta) => { const state = __gojrSyncState(receiver); state.n = (state.n || 0n) + BigInt(delta ?? 0n); return null; },",
    "    \"WaitGroup.Done\": async (receiver) => { const state = __gojrSyncState(receiver); state.n = (state.n || 0n) - 1n; return null; },",
    "    \"WaitGroup.Wait\": async () => null,",
    "    \"Cond.Broadcast\": async () => null, \"Cond.Signal\": async () => null, \"Cond.Wait\": async () => null,",
    "    \"Map.Clear\": async (receiver) => { __gojrSyncMap(receiver).clear(); return null; },",
    "    \"Map.CompareAndDelete\": async (receiver, key, oldValue) => { const map = __gojrSyncMap(receiver); if (!map.has(key) || !__gojrEqual(map.get(key), oldValue)) return false; map.delete(key); return true; },",
    "    \"Map.CompareAndSwap\": async (receiver, key, oldValue, newValue) => { const map = __gojrSyncMap(receiver); if (!map.has(key) || !__gojrEqual(map.get(key), oldValue)) return false; map.set(key, newValue); return true; },",
    "    \"Map.Delete\": async (receiver, key) => { __gojrSyncMap(receiver).delete(key); return null; },",
    "    \"Map.Load\": async (receiver, key) => { const map = __gojrSyncMap(receiver); return map.has(key) ? __gojrTuple([map.get(key), true]) : __gojrTuple([null, false]); },",
    "    \"Map.LoadAndDelete\": async (receiver, key) => { const map = __gojrSyncMap(receiver); if (!map.has(key)) return __gojrTuple([null, false]); const value = map.get(key); map.delete(key); return __gojrTuple([value, true]); },",
    "    \"Map.LoadOrStore\": async (receiver, key, value) => { const map = __gojrSyncMap(receiver); if (map.has(key)) return __gojrTuple([map.get(key), true]); map.set(key, value); return __gojrTuple([value, false]); },",
    "    \"Map.Range\": async (receiver, fn) => { for (const [key, value] of __gojrSyncMap(receiver).entries()) { if (!await fn(key, value)) break; } return null; },",
    "    \"Map.Store\": async (receiver, key, value) => { __gojrSyncMap(receiver).set(key, value); return null; },",
    "    \"Map.Swap\": async (receiver, key, value) => { const map = __gojrSyncMap(receiver); const old = map.has(key) ? map.get(key) : null; const loaded = map.has(key); map.set(key, value); return __gojrTuple([old, loaded]); }",
    "  };",
    "}",
    "function __gojrInternalSyncPackage() {",
    "  return {",
    "    __gojrTypeDescriptors: __gojrInternalSyncDescriptors(),",
    "    \"HashTrieMap.Load\": async (receiver, key) => { const map = __gojrSyncMap(receiver); return map.has(key) ? __gojrTuple([map.get(key), true]) : __gojrTuple([null, false]); },",
    "    \"HashTrieMap.LoadOrStore\": async (receiver, key, value) => { const map = __gojrSyncMap(receiver); if (map.has(key)) return __gojrTuple([map.get(key), true]); map.set(key, value); return __gojrTuple([value, false]); },",
    "    \"HashTrieMap.LoadAndDelete\": async (receiver, key) => { const map = __gojrSyncMap(receiver); if (!map.has(key)) return __gojrTuple([null, false]); const value = map.get(key); map.delete(key); return __gojrTuple([value, true]); },",
    "    \"HashTrieMap.CompareAndDelete\": async (receiver, key, oldValue) => { const map = __gojrSyncMap(receiver); if (!map.has(key) || !__gojrEqual(map.get(key), oldValue)) return false; map.delete(key); return true; },",
    "    \"HashTrieMap.All\": async (receiver) => async (yieldFn) => { for (const [key, value] of __gojrSyncMap(receiver).entries()) { if (!await yieldFn(key, value)) break; } return null; }",
    "  };",
    "}",
    "function __gojrSyncDescriptors() {",
    "  const out = {};",
    "  const add = (name, fields = []) => { const descriptor = { type: `sync.${name}`, name, string: `sync.${name}`, kind: \"struct\", pkgPath: \"sync\", pkgName: \"sync\", fields }; out[name] = descriptor; out[`sync.${name}`] = descriptor; };",
    "  add(\"Mutex\"); add(\"RWMutex\"); add(\"Once\", [{ name: \"done\", type: \"bool\" }]); add(\"Pool\", [{ name: \"New\", type: \"func() any\" }]); add(\"WaitGroup\", [{ name: \"n\", type: \"int\" }]); add(\"Cond\", [{ name: \"L\", type: \"sync.Locker\" }]); add(\"Map\");",
    "  return out;",
    "}",
    "function __gojrSyncDescriptorForTypeName(typeName, pkgPath = undefined) {",
    "  const text = String(typeName || \"\").trim();",
    "  if (text.startsWith(\"internal/sync.HashTrieMap\") || ((pkgPath === \"internal/sync\" || gojrPackageArtifact.importPath === \"internal/sync\") && text.startsWith(\"HashTrieMap\"))) return __gojrInternalSyncDescriptors()[\"internal/sync.HashTrieMap\"];",
    "  const local = text.startsWith(\"sync.\") ? text.slice(\"sync.\".length) : text;",
    "  if (!text.startsWith(\"sync.\") && pkgPath !== \"sync\" && gojrPackageArtifact.importPath !== \"sync\") return undefined;",
    "  if (![\"Mutex\", \"RWMutex\", \"Once\", \"Pool\", \"WaitGroup\", \"Cond\", \"Map\"].includes(local)) return undefined;",
    "  return __gojrSyncDescriptors()[`sync.${local}`];",
    "}",
    "function __gojrInternalSyncDescriptors() {",
    "  const descriptor = { type: \"internal/sync.HashTrieMap\", name: \"HashTrieMap\", string: \"sync.HashTrieMap\", kind: \"struct\", pkgPath: \"internal/sync\", pkgName: \"sync\", fields: [] };",
    "  return { HashTrieMap: descriptor, \"internal/sync.HashTrieMap\": descriptor };",
    "}",
    "function __gojrSyncZero(typeText, pkgPath = undefined) {",
    "  const descriptor = __gojrSyncDescriptorForTypeName(typeText, pkgPath);",
    "  if (!descriptor) return undefined;",
    "  const value = { __gojrType: descriptor.type, __gojrPkgPath: descriptor.pkgPath || \"sync\" };",
    "  if (descriptor.name === \"Pool\") value.New = null;",
    "  if (descriptor.name === \"WaitGroup\") value.n = 0n;",
    "  if (descriptor.name === \"Cond\") value.L = null;",
    "  return value;",
    "}",
    "function __gojrSyncState(receiver) {",
    "  const target = __gojrDerefIfPointer(receiver);",
    "  if (target == null) throw new Error(\"GOJR_RUNTIME001: sync method on nil receiver\");",
    "  if (!Object.prototype.hasOwnProperty.call(target, \"__gojrSyncState\")) Object.defineProperty(target, \"__gojrSyncState\", { value: {}, configurable: true });",
    "  return target.__gojrSyncState;",
    "}",
    "async function __gojrSyncOnceDo(receiver, fn) {",
    "  const state = __gojrSyncState(receiver);",
    "  if (state.done) return null;",
    "  state.done = true;",
    "  if (typeof fn === \"function\") await fn();",
    "  return null;",
    "}",
    "function __gojrSyncPoolItems(receiver) {",
    "  const state = __gojrSyncState(receiver);",
    "  if (!state.items) state.items = [];",
    "  return state.items;",
    "}",
    "async function __gojrSyncPoolGet(receiver) {",
    "  const items = __gojrSyncPoolItems(receiver);",
    "  if (items.length > 0) return items.pop();",
    "  const target = __gojrDerefIfPointer(receiver);",
    "  const newFn = target && target.New;",
    "  return typeof newFn === \"function\" ? await newFn() : null;",
    "}",
    "function __gojrSyncPoolPut(receiver, value) {",
    "  if (value !== null && value !== undefined) __gojrSyncPoolItems(receiver).push(value);",
    "  return null;",
    "}",
    "function __gojrSyncMap(receiver) {",
    "  const state = __gojrSyncState(receiver);",
    "  if (!state.map) state.map = new Map();",
    "  return state.map;",
    "}",
    "function __gojrSyncLocker() {",
    "  return { __gojrMethods: { Lock: async () => null, Unlock: async () => null } };",
    "}",
    "function __gojrAtomicPackage() {",
    "  return {",
    "    __gojrTypeDescriptors: __gojrAtomicDescriptors(),",
    "    LoadInt32: async (addr) => __gojrAtomicLoad(addr, 32, true, \"LoadInt32\"),",
    "    LoadInt64: async (addr) => __gojrAtomicLoad(addr, 64, true, \"LoadInt64\"),",
    "    LoadUint32: async (addr) => __gojrAtomicLoad(addr, 32, false, \"LoadUint32\"),",
    "    LoadUint64: async (addr) => __gojrAtomicLoad(addr, 64, false, \"LoadUint64\"),",
    "    LoadUintptr: async (addr) => __gojrAtomicLoad(addr, 64, false, \"LoadUintptr\"),",
    "    StoreInt32: async (addr, val) => __gojrAtomicStore(addr, val, 32, true, \"StoreInt32\"),",
    "    StoreInt64: async (addr, val) => __gojrAtomicStore(addr, val, 64, true, \"StoreInt64\"),",
    "    StoreUint32: async (addr, val) => __gojrAtomicStore(addr, val, 32, false, \"StoreUint32\"),",
    "    StoreUint64: async (addr, val) => __gojrAtomicStore(addr, val, 64, false, \"StoreUint64\"),",
    "    StoreUintptr: async (addr, val) => __gojrAtomicStore(addr, val, 64, false, \"StoreUintptr\"),",
    "    SwapInt32: async (addr, val) => __gojrAtomicSwap(addr, val, 32, true, \"SwapInt32\"),",
    "    SwapInt64: async (addr, val) => __gojrAtomicSwap(addr, val, 64, true, \"SwapInt64\"),",
    "    SwapUint32: async (addr, val) => __gojrAtomicSwap(addr, val, 32, false, \"SwapUint32\"),",
    "    SwapUint64: async (addr, val) => __gojrAtomicSwap(addr, val, 64, false, \"SwapUint64\"),",
    "    SwapUintptr: async (addr, val) => __gojrAtomicSwap(addr, val, 64, false, \"SwapUintptr\"),",
    "    CompareAndSwapInt32: async (addr, old, val) => __gojrAtomicCompareAndSwap(addr, old, val, 32, true, \"CompareAndSwapInt32\"),",
    "    CompareAndSwapInt64: async (addr, old, val) => __gojrAtomicCompareAndSwap(addr, old, val, 64, true, \"CompareAndSwapInt64\"),",
    "    CompareAndSwapUint32: async (addr, old, val) => __gojrAtomicCompareAndSwap(addr, old, val, 32, false, \"CompareAndSwapUint32\"),",
    "    CompareAndSwapUint64: async (addr, old, val) => __gojrAtomicCompareAndSwap(addr, old, val, 64, false, \"CompareAndSwapUint64\"),",
    "    CompareAndSwapUintptr: async (addr, old, val) => __gojrAtomicCompareAndSwap(addr, old, val, 64, false, \"CompareAndSwapUintptr\"),",
    "    AddInt32: async (addr, delta) => __gojrAtomicAdd(addr, delta, 32, true, \"AddInt32\"),",
    "    AddInt64: async (addr, delta) => __gojrAtomicAdd(addr, delta, 64, true, \"AddInt64\"),",
    "    AddUint32: async (addr, delta) => __gojrAtomicAdd(addr, delta, 32, false, \"AddUint32\"),",
    "    AddUint64: async (addr, delta) => __gojrAtomicAdd(addr, delta, 64, false, \"AddUint64\"),",
    "    AddUintptr: async (addr, delta) => __gojrAtomicAdd(addr, delta, 64, false, \"AddUintptr\"),",
    "    AndInt32: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 32, true, \"&\", \"AndInt32\"),",
    "    AndInt64: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, true, \"&\", \"AndInt64\"),",
    "    AndUint32: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 32, false, \"&\", \"AndUint32\"),",
    "    AndUint64: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, false, \"&\", \"AndUint64\"),",
    "    AndUintptr: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, false, \"&\", \"AndUintptr\"),",
    "    OrInt32: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 32, true, \"|\", \"OrInt32\"),",
    "    OrInt64: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, true, \"|\", \"OrInt64\"),",
    "    OrUint32: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 32, false, \"|\", \"OrUint32\"),",
    "    OrUint64: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, false, \"|\", \"OrUint64\"),",
    "    OrUintptr: async (addr, mask) => __gojrAtomicBitwise(addr, mask, 64, false, \"|\", \"OrUintptr\"),",
    "    LoadPointer: async (addr) => __gojrAtomicPointerLoad(addr, \"LoadPointer\"),",
    "    StorePointer: async (addr, val) => __gojrAtomicPointerStore(addr, val, \"StorePointer\"),",
    "    SwapPointer: async (addr, val) => __gojrAtomicPointerSwap(addr, val, \"SwapPointer\"),",
    "    CompareAndSwapPointer: async (addr, old, val) => __gojrAtomicPointerCompareAndSwap(addr, old, val, \"CompareAndSwapPointer\")",
    "  };",
    "}",
    "function __gojrAtomicDescriptors() {",
    "  const out = {};",
    "  for (const name of [\"Bool\", \"Int32\", \"Int64\", \"Uint32\", \"Uint64\", \"Uintptr\", \"Value\", \"Pointer\"]) {",
    "    const descriptor = { type: `sync/atomic.${name}`, name, string: `atomic.${name}`, kind: \"struct\", pkgPath: \"sync/atomic\", pkgName: \"atomic\", fields: [] };",
    "    out[name] = descriptor;",
    "    out[`sync/atomic.${name}`] = descriptor;",
    "    out[`atomic.${name}`] = descriptor;",
    "  }",
    "  return out;",
    "}",
    "function __gojrAtomicDescriptorForTypeName(typeName) {",
    "  const text = String(typeName || \"\").trim();",
    "  let local = text.startsWith(\"sync/atomic.\") ? text.slice(\"sync/atomic.\".length) : text;",
    "  if (local.startsWith(\"atomic.\")) local = local.slice(\"atomic.\".length);",
    "  const base = local.includes(\"[\") ? local.slice(0, local.indexOf(\"[\")) : local;",
    "  if (![\"Bool\", \"Int32\", \"Int64\", \"Uint32\", \"Uint64\", \"Uintptr\", \"Value\", \"Pointer\"].includes(base)) return undefined;",
    "  return { type: text.startsWith(\"sync/atomic.\") ? text : `sync/atomic.${local}`, name: base, string: `atomic.${local}`, kind: \"struct\", pkgPath: \"sync/atomic\", pkgName: \"atomic\", fields: [] };",
    "}",
    "function __gojrAtomicZero(typeText) {",
    "  const descriptor = __gojrAtomicDescriptorForTypeName(typeText);",
    "  if (!descriptor) return undefined;",
    "  switch (descriptor.name) {",
    "    case \"Bool\": return __gojrAtomicBool(descriptor.type);",
    "    case \"Int32\": return __gojrAtomicIntegerBox(descriptor.type, 32, true);",
    "    case \"Int64\": return __gojrAtomicIntegerBox(descriptor.type, 64, true);",
    "    case \"Uint32\": return __gojrAtomicIntegerBox(descriptor.type, 32, false);",
    "    case \"Uint64\": return __gojrAtomicIntegerBox(descriptor.type, 64, false);",
    "    case \"Uintptr\": return __gojrAtomicIntegerBox(descriptor.type, 64, false);",
    "    case \"Pointer\": return __gojrAtomicPointerBox(descriptor.type);",
    "    case \"Value\": return __gojrAtomicValueBox(descriptor.type);",
    "    default: return undefined;",
    "  }",
    "}",
    "function __gojrAtomicCoerce(value, bits, signed) {",
    "  const raw = BigInt(value ?? 0n);",
    "  return signed ? BigInt.asIntN(bits, raw) : BigInt.asUintN(bits, raw);",
    "}",
    "function __gojrAtomicCell(addr, name) {",
    "  if (addr && addr.__gojrPointer === true) return addr;",
    "  throw new Error(`GOJR_RUNTIME001: sync/atomic.${name} address is not a pointer`);",
    "}",
    "function __gojrAtomicLoad(addr, bits, signed, name) {",
    "  return __gojrAtomicCoerce(__gojrAtomicCell(addr, name).__gojrGet() ?? 0n, bits, signed);",
    "}",
    "function __gojrAtomicStore(addr, value, bits, signed, name) {",
    "  __gojrAtomicCell(addr, name).__gojrSet(__gojrAtomicCoerce(value, bits, signed));",
    "  return null;",
    "}",
    "function __gojrAtomicSwap(addr, value, bits, signed, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const previous = __gojrAtomicCoerce(cell.__gojrGet() ?? 0n, bits, signed);",
    "  cell.__gojrSet(__gojrAtomicCoerce(value, bits, signed));",
    "  return previous;",
    "}",
    "function __gojrAtomicCompareAndSwap(addr, oldValue, value, bits, signed, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const current = __gojrAtomicCoerce(cell.__gojrGet() ?? 0n, bits, signed);",
    "  if (current !== __gojrAtomicCoerce(oldValue, bits, signed)) return false;",
    "  cell.__gojrSet(__gojrAtomicCoerce(value, bits, signed));",
    "  return true;",
    "}",
    "function __gojrAtomicAdd(addr, delta, bits, signed, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const next = __gojrAtomicCoerce(__gojrAtomicCoerce(cell.__gojrGet() ?? 0n, bits, signed) + BigInt(delta ?? 0n), bits, signed);",
    "  cell.__gojrSet(next);",
    "  return next;",
    "}",
    "function __gojrAtomicBitwise(addr, mask, bits, signed, operator, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const previous = __gojrAtomicCoerce(cell.__gojrGet() ?? 0n, bits, signed);",
    "  const next = __gojrAtomicCoerce(operator === \"&\" ? previous & BigInt(mask ?? 0n) : previous | BigInt(mask ?? 0n), bits, signed);",
    "  cell.__gojrSet(next);",
    "  return previous;",
    "}",
    "function __gojrAtomicPointerLoad(addr, name) {",
    "  return __gojrAtomicCell(addr, name).__gojrGet() ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\");",
    "}",
    "function __gojrAtomicPointerStore(addr, value, name) {",
    "  __gojrAtomicCell(addr, name).__gojrSet(value ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\"));",
    "  return null;",
    "}",
    "function __gojrAtomicPointerSwap(addr, value, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const previous = cell.__gojrGet() ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\");",
    "  cell.__gojrSet(value ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\"));",
    "  return previous;",
    "}",
    "function __gojrAtomicPointerCompareAndSwap(addr, oldValue, value, name) {",
    "  const cell = __gojrAtomicCell(addr, name);",
    "  const current = cell.__gojrGet() ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\");",
    "  if (!__gojrEqual(current, oldValue ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\"))) return false;",
    "  cell.__gojrSet(value ?? __gojrTypedNil(\"unsafe.Pointer\", \"unsafe\"));",
    "  return true;",
    "}",
    "function __gojrAtomicBool(typeName) {",
    "  let value = false;",
    "  const box = { __gojrType: typeName, __gojrPkgPath: \"sync/atomic\" };",
    "  box.__gojrMethods = {",
    "    CompareAndSwap: async (_self, oldValue, next) => { if (value !== Boolean(oldValue)) return false; value = Boolean(next); return true; },",
    "    Load: async () => value,",
    "    Store: async (_self, next) => { value = Boolean(next); return null; },",
    "    Swap: async (_self, next) => { const previous = value; value = Boolean(next); return previous; }",
    "  };",
    "  return box;",
    "}",
    "function __gojrAtomicIntegerBox(typeName, bits, signed) {",
    "  let value = 0n;",
    "  const coerce = (next) => __gojrAtomicCoerce(next, bits, signed);",
    "  const box = { __gojrType: typeName, __gojrPkgPath: \"sync/atomic\" };",
    "  box.__gojrMethods = {",
    "    Add: async (_self, delta) => { value = coerce(value + BigInt(delta ?? 0n)); return value; },",
    "    And: async (_self, mask) => { const previous = value; value = coerce(value & BigInt(mask ?? 0n)); return previous; },",
    "    CompareAndSwap: async (_self, oldValue, next) => { if (value !== coerce(oldValue)) return false; value = coerce(next); return true; },",
    "    Load: async () => value,",
    "    Or: async (_self, mask) => { const previous = value; value = coerce(value | BigInt(mask ?? 0n)); return previous; },",
    "    Store: async (_self, next) => { value = coerce(next); return null; },",
    "    Swap: async (_self, next) => { const previous = value; value = coerce(next); return previous; }",
    "  };",
    "  return box;",
    "}",
    "function __gojrAtomicPointerBox(typeName) {",
    "  let value = null;",
    "  const box = { __gojrType: typeName, __gojrPkgPath: \"sync/atomic\" };",
    "  box.__gojrMethods = {",
    "    CompareAndSwap: async (_self, oldValue, next) => { if (!__gojrEqual(value, oldValue ?? null)) return false; value = next ?? null; return true; },",
    "    Load: async () => value,",
    "    Store: async (_self, next) => { value = next ?? null; return null; },",
    "    Swap: async (_self, next) => { const previous = value; value = next ?? null; return previous; }",
    "  };",
    "  return box;",
    "}",
    "function __gojrAtomicValueBox(typeName) {",
    "  let value = null;",
    "  const box = { __gojrType: typeName, __gojrPkgPath: \"sync/atomic\" };",
    "  box.__gojrMethods = {",
    "    CompareAndSwap: async (_self, oldValue, next) => { if (!__gojrEqual(value, oldValue ?? null)) return false; value = next ?? null; return true; },",
    "    Load: async () => value,",
    "    Store: async (_self, next) => { value = next ?? null; return null; },",
    "    Swap: async (_self, next) => { const previous = value; value = next ?? null; return previous; }",
    "  };",
    "  return box;",
    "}",
    "function __gojrWeakPackage() {",
    "  return {",
    "    __gojrExportIndex: { Make: { typeText: \"func[T any](*T) Pointer[T]\" } },",
    "    __gojrTypeDescriptors: { Pointer: __gojrWeakDescriptorForTypeName(\"weak.Pointer[T]\") },",
    "    Make: async (__gojrTypeArgs, ptr) => __gojrWeakPointerBox(`weak.Pointer[${__gojrTypeArg(__gojrTypeArgs, \"T\")}]`, ptr),",
    "    \"Pointer.Value\": async (_typeArgs, receiver) => __gojrWeakPointerValue(receiver)",
    "  };",
    "}",
    "function __gojrWeakDescriptorForTypeName(typeName) {",
    "  const text = String(typeName || \"\").trim();",
    "  if (!text.startsWith(\"weak.\")) return undefined;",
    "  let local = text.slice(\"weak.\".length);",
    "  const base = local.includes(\"[\") ? local.slice(0, local.indexOf(\"[\")) : local;",
    "  if (base !== \"Pointer\") return undefined;",
    "  return { type: text.startsWith(\"weak.\") ? text : `weak.${local}`, name: \"Pointer\", string: `weak.${local}`, kind: \"struct\", pkgPath: \"weak\", pkgName: \"weak\", fields: [] };",
    "}",
    "function __gojrWeakZero(typeText) {",
    "  const descriptor = __gojrWeakDescriptorForTypeName(typeText);",
    "  return descriptor ? __gojrWeakPointerBox(descriptor.type, null) : undefined;",
    "}",
    "function __gojrWeakStruct(typeName, value, pkgPath) {",
    "  const descriptor = __gojrWeakDescriptorForTypeName(typeName);",
    "  if (!descriptor || (pkgPath && pkgPath !== \"weak\")) return undefined;",
    "  return __gojrWeakPointerBox(descriptor.type, value && Object.prototype.hasOwnProperty.call(value, \"u\") ? value.u : null);",
    "}",
    "function __gojrWeakPointerBox(typeName, ptr) {",
    "  const box = { __gojrWeakPointer: true, __gojrType: typeName, __gojrPkgPath: \"weak\", __gojrWeakValue: ptr ?? null };",
    "  box.__gojrMethods = { Value: async () => __gojrWeakPointerValue(box) };",
    "  return box;",
    "}",
    "function __gojrWeakPointerValue(value) {",
    "  if (value && value.__gojrWeakPointer === true) return value.__gojrWeakValue ?? null;",
    "  const structValue = __gojrDerefIfPointer(value);",
    "  if (structValue && structValue.__gojrWeakPointer === true) return structValue.__gojrWeakValue ?? null;",
    "  if (structValue && Object.prototype.hasOwnProperty.call(structValue, \"u\")) return structValue.u ?? null;",
    "  return null;",
    "}",
    "function __gojrRuntimePackage() {",
    "  return {",
    "    GOOS: \"js\",",
    "    GOARCH: \"gojr\",",
    "    Compiler: \"gojr\",",
    "    MemProfileRate: 0n,",
    "    AddCleanup: async () => ({ __gojrMethods: { Stop: async () => null } }),",
    "    BlockProfile: async () => { throw new Error(\"gojr error: runtime.BlockProfile not implemented\"); },",
    "    Caller: async () => __gojrTuple([0n, \"\", 0n, false]),",
    "    Callers: async () => 0n,",
    "    CallersFrames: async () => ({ __gojrMethods: { Next: async () => __gojrTuple([__gojrStruct(\"runtime.Frame\", { PC: 0n, Func: null, Function: \"\", File: \"\", Line: 0n, Entry: 0n }, \"runtime\"), false]) } }),",
    "    FuncForPC: async () => ({ __gojrMethods: { Entry: async () => 0n, FileLine: async () => __gojrTuple([\"\", 0n]), Name: async () => \"\" } }),",
    "    GC: async () => null,",
    "    GOMAXPROCS: async () => 1n,",
    "    GOROOT: async () => \"\",",
    "    Goexit: async () => null,",
    "    Gosched: async () => null,",
    "    KeepAlive: async () => null,",
    "    GoroutineProfile: async () => { throw new Error(\"gojr error: runtime.GoroutineProfile not implemented\"); },",
    "    MemProfile: async () => { throw new Error(\"gojr error: runtime.MemProfile not implemented\"); },",
    "    MutexProfile: async () => { throw new Error(\"gojr error: runtime.MutexProfile not implemented\"); },",
    "    NumCPU: async () => 1n,",
    "    NumGoroutine: async () => 1n,",
    "    ReadMemStats: async () => null,",
    "    ReadTrace: async () => null,",
    "    SetBlockProfileRate: async () => { throw new Error(\"gojr error: runtime.SetBlockProfileRate not implemented\"); },",
    "    SetCPUProfileRate: async () => { throw new Error(\"gojr error: runtime.SetCPUProfileRate not implemented\"); },",
    "    SetFinalizer: async () => null,",
    "    SetMutexProfileFraction: async () => { throw new Error(\"gojr error: runtime.SetMutexProfileFraction not implemented\"); },",
    "    StartTrace: async () => null,",
    "    StopTrace: async () => null,",
    "    Stack: async (buffer) => __gojrRuntimeStack(buffer),",
    "    ThreadCreateProfile: async () => { throw new Error(\"gojr error: runtime.ThreadCreateProfile not implemented\"); },",
    "    Version: async () => \"gojr\"",
    "  };",
    "}",
    "function __gojrRuntimeStack(buffer) {",
    "  const target = __gojrDerefIfPointer(buffer);",
    "  if (!Array.isArray(target) && !ArrayBuffer.isView(target)) return 0n;",
    "  const bytes = new TextEncoder().encode(\"goroutine 1 [running]:\\nruntime.Stack(...)\\n\");",
    "  const count = Math.min(target.length, bytes.length);",
    "  for (let index = 0; index < count; index += 1) target[index] = BigInt(bytes[index] || 0);",
    "  return BigInt(count);",
    "}",
    "function __gojrFsPackage() {",
    "  const fileInfoToDirEntry = (info) => ({ __gojrType: \"io/fs.dirEntry\", __gojrPkgPath: \"io/fs\", __gojrMethods: {",
    "    Name: async () => info && info.__gojrMethods && info.__gojrMethods.Name ? info.__gojrMethods.Name(info) : \"\",",
    "    IsDir: async () => info && info.__gojrMethods && info.__gojrMethods.IsDir ? info.__gojrMethods.IsDir(info) : false,",
    "    Type: async () => info && info.__gojrMethods && info.__gojrMethods.Mode ? info.__gojrMethods.Mode(info) : 0n,",
    "    Info: async () => __gojrTuple([info ?? null, null])",
    "  } });",
    "  return {",
    "    __gojrTypeDescriptors: { FileMode: { type: \"io/fs.FileMode\", name: \"FileMode\", string: \"fs.FileMode\", kind: \"uint32\", pkgPath: \"io/fs\", pkgName: \"fs\" }, \"io/fs.FileMode\": { type: \"io/fs.FileMode\", name: \"FileMode\", string: \"fs.FileMode\", kind: \"uint32\", pkgPath: \"io/fs\", pkgName: \"fs\" } },",
    "    ModeDir: 2147483648n, ModeAppend: 1073741824n, ModeExclusive: 536870912n, ModeTemporary: 268435456n, ModeSymlink: 134217728n, ModeDevice: 67108864n, ModeNamedPipe: 33554432n, ModeSocket: 16777216n, ModeSetuid: 8388608n, ModeSetgid: 4194304n, ModeCharDevice: 2097152n, ModeSticky: 1048576n, ModeIrregular: 524288n, ModeType: 2399666176n, ModePerm: 511n,",
    "    ErrInvalid: __gojrError(\"invalid argument\"), ErrPermission: __gojrError(\"permission denied\"), ErrExist: __gojrError(\"file already exists\"), ErrNotExist: __gojrError(\"file does not exist\"), ErrClosed: __gojrError(\"file already closed\"), SkipDir: __gojrError(\"skip this directory\"), SkipAll: __gojrError(\"skip everything and stop the walk\"),",
    "    FormatDirEntry: async () => \"\", FormatFileInfo: async () => \"\",",
    "    Glob: async () => __gojrTuple([[], null]), ReadDir: async () => __gojrTuple([[], null]), ReadFile: async () => __gojrTuple([new Uint8Array(), null]),",
    "    Stat: async () => __gojrTuple([null, null]), Sub: async (fsys) => __gojrTuple([fsys ?? null, null]), ValidPath: async () => true, WalkDir: async () => null,",
    "    FileInfoToDirEntry: async (info) => fileInfoToDirEntry(info)",
    "  };",
    "}",
    "function __gojrOsPackage() {",
    "  return {",
    "    Args: __gojrOsArgs(),",
    "    Stdin: __gojrOsFilePointer(0, \"/dev/stdin\"),",
    "    Stdout: __gojrOsFilePointer(1, \"/dev/stdout\"),",
    "    Stderr: __gojrOsFilePointer(2, \"/dev/stderr\"),",
    "    ErrInvalid: __gojrError(\"invalid argument\"),",
    "    ErrPermission: __gojrError(\"permission denied\"),",
    "    ErrExist: __gojrError(\"file already exists\"),",
    "    ErrNotExist: __gojrError(\"file does not exist\"),",
    "    ErrClosed: __gojrError(\"file already closed\"),",
    "    ErrDeadlineExceeded: __gojrError(\"i/o timeout\"),",
    "    ErrNoDeadline: __gojrError(\"file type does not support deadline\"),",
    "    ErrProcessDone: __gojrError(\"os: process already finished\"),",
    "    O_RDONLY: 0n, O_WRONLY: 1n, O_RDWR: 2n, O_APPEND: 8n, O_CREATE: 512n, O_EXCL: 2048n, O_SYNC: 128n, O_TRUNC: 1024n,",
    "    ModeDir: 2147483648n, ModeAppend: 1073741824n, ModeExclusive: 536870912n, ModeTemporary: 268435456n, ModeSymlink: 134217728n, ModeDevice: 67108864n, ModeNamedPipe: 33554432n, ModeSocket: 16777216n, ModeSetuid: 8388608n, ModeSetgid: 4194304n, ModeCharDevice: 2097152n, ModeSticky: 1048576n, ModeIrregular: 524288n, ModeType: 2399666176n, ModePerm: 511n,",
    "    PathSeparator: 47n, PathListSeparator: 58n,",
    "    DevNull: \"/dev/null\",",
    "    Exit: async (code) => { const error = new Error(`os.Exit(${Number(code || 0)})`); error.__gojrExitCode = Number(code || 0); throw error; },",
    "    Getenv: async (key) => __gojrOsGetenv(String(key ?? \"\")),",
    "    LookupEnv: async (key) => { const name = String(key ?? \"\"); const value = __gojrOsGetenv(name); return __gojrTuple([value, __gojrOsHasEnv(name)]); },",
    "    Environ: async () => Object.entries(__gojrOsEnv()).map(([key, value]) => `${key}=${value}`),",
    "    TempDir: async () => \"/tmp\",",
    "    Getwd: async () => __gojrTuple([\"/\", null]),",
    "    Hostname: async () => __gojrTuple([\"gojr\", null]),",
    "    Executable: async () => __gojrTuple([\"gojr\", null]),",
    "    UserCacheDir: async () => __gojrTuple([\"/tmp\", null]),",
    "    UserConfigDir: async () => __gojrTuple([\"/tmp\", null]),",
    "    UserHomeDir: async () => __gojrTuple([\"/\", null]),",
    "    Open: async (name) => __gojrTuple([__gojrOsFilePointer(-1, String(name ?? \"\")), null]),",
    "    OpenFile: async (name) => __gojrTuple([__gojrOsFilePointer(-1, String(name ?? \"\")), null]),",
    "    Create: async (name) => __gojrTuple([__gojrOsFilePointer(-1, String(name ?? \"\")), null]),",
    "    NewFile: async (fd, name) => __gojrOsFilePointer(Number(fd || 0), String(name ?? \"\")),",
    "    ReadFile: async () => __gojrTuple([new Uint8Array(), __gojrError(\"gojr error: os.ReadFile not implemented\")]),",
    "    WriteFile: async () => null,",
    "    Chdir: async () => null, Chmod: async () => null, Chown: async () => null, Chtimes: async () => null, Clearenv: async () => null,",
    "    CreateTemp: async (_dir, pattern) => __gojrTuple([__gojrOsFilePointer(-1, String(pattern ?? \"\")), null]),",
    "    DirFS: async (dir) => String(dir ?? \"\"),",
    "    Expand: async (s) => String(s ?? \"\"), ExpandEnv: async (s) => String(s ?? \"\"),",
    "    FindProcess: async () => __gojrTuple([__gojrStruct(\"os.Process\", {}, \"os\"), null]),",
    "    Getegid: async () => 0n, Geteuid: async () => 0n, Getgid: async () => 0n, Getgroups: async () => __gojrTuple([[], null]), Getpagesize: async () => 4096n, Getpid: async () => 1n, Getppid: async () => 0n, Getuid: async () => 0n,",
    "    Lchown: async () => null, Link: async () => null, Lstat: async () => __gojrTuple([null, null]),",
    "    Mkdir: async () => null, MkdirAll: async () => null, MkdirTemp: async (_dir, pattern) => __gojrTuple([String(pattern ?? \"\"), null]),",
    "    Pipe: async () => __gojrTuple([__gojrOsFilePointer(-1, \"pipe-r\"), __gojrOsFilePointer(-1, \"pipe-w\"), null]),",
    "    ReadDir: async () => __gojrTuple([[], null]), Readlink: async () => __gojrTuple([\"\", null]),",
    "    Remove: async () => null, RemoveAll: async () => null, Rename: async () => null, SameFile: async () => false, Setenv: async () => null,",
    "    StartProcess: async () => __gojrTuple([__gojrStruct(\"os.Process\", {}, \"os\"), null]),",
    "    Stat: async () => __gojrTuple([null, null]), Symlink: async () => null, Truncate: async () => null, Unsetenv: async () => null,",
    "    IsExist: async () => false, IsNotExist: async () => false, IsPermission: async () => false, IsTimeout: async () => false, IsPathSeparator: async (c) => Number(c) === 47",
    "  };",
    "}",
    "function __gojrOsArgs() {",
    "  const args = __gojrActiveRuntimeOptions.argv || __gojrActiveRuntimeOptions.args || [\"gojr\"];",
    "  return Array.isArray(args) ? args.map((item) => String(item)) : [\"gojr\"];",
    "}",
    "function __gojrOsEnv() {",
    "  const env = __gojrActiveRuntimeOptions.env || __gojrActiveRuntimeOptions.environment || {};",
    "  return env && typeof env === \"object\" ? env : {};",
    "}",
    "function __gojrOsHasEnv(key) {",
    "  const env = __gojrOsEnv();",
    "  return Object.prototype.hasOwnProperty.call(env, key);",
    "}",
    "function __gojrOsGetenv(key) {",
    "  const env = __gojrOsEnv();",
    "  return Object.prototype.hasOwnProperty.call(env, key) ? String(env[key]) : \"\";",
    "}",
    "function __gojrError(message) {",
    "  return { __gojrInterface: true, interfaceType: \"error\", value: { __gojrType: \"errorString\", message: String(message), __gojrMethods: { Error: async (self) => self.message } } };",
    "}",
    "function __gojrOsFilePointer(fd, name) {",
    "  const file = __gojrStruct(\"os.File\", { fd: BigInt(fd), name: String(name) }, \"os\");",
    "  const pointer = __gojrPointerValue(\"os.File\", file, \"os\");",
    "  pointer.__gojrMethods = {",
    "    Chdir: async () => null, Chmod: async () => null, Chown: async () => null, Close: async () => null, Fd: async () => BigInt(fd), Name: async () => String(name),",
    "    Read: async (_self, b) => __gojrTuple([__gojrOsRead(fd, b), null]),",
    "    ReadAt: async (_self, b) => __gojrTuple([__gojrOsRead(fd, b), null]),",
    "    ReadDir: async () => __gojrTuple([[], null]), Readdir: async () => __gojrTuple([[], null]), Readdirnames: async () => __gojrTuple([[], null]),",
    "    Seek: async () => __gojrTuple([0n, null]),",
    "    SetDeadline: async () => null, SetReadDeadline: async () => null, SetWriteDeadline: async () => null,",
    "    Stat: async () => __gojrTuple([null, null]),",
    "    Sync: async () => null, SyscallConn: async () => __gojrTuple([null, null]), Truncate: async () => null,",
    "    Write: async (_self, b) => __gojrTuple([__gojrOsWrite(fd, b), null]),",
    "    WriteAt: async (_self, b) => __gojrTuple([__gojrOsWrite(fd, b), null]),",
    "    WriteString: async (_self, s) => { const bytes = __gojrStringBytes(s ?? \"\"); return __gojrTuple([__gojrOsWrite(fd, bytes), null]); }",
    "  };",
    "  return pointer;",
    "}",
    "function __gojrOsRead(fd, target) {",
    "  const sequence = __gojrPointerSequence(target);",
    "  const value = sequence ? sequence.values : __gojrDerefIfPointer(target);",
    "  const start = sequence ? Number(sequence.index || 0) : 0;",
    "  const length = __gojrLen(value);",
    "  if (length <= 0) return 0n;",
    "  const buffer = new Uint8Array(length);",
    "  const hostRead = globalThis.__gojrReadSync;",
    "  const count = typeof hostRead === \"function\" ? Number(hostRead(fd, buffer, 0, length, null) || 0) : 0;",
    "  for (let index = 0; index < count; index += 1) {",
    "    const targetIndex = start + index;",
    "    if (value instanceof Uint8Array || ArrayBuffer.isView(value)) value[targetIndex] = Number(buffer[index] || 0);",
    "    else value[targetIndex] = BigInt(buffer[index] || 0);",
    "  }",
    "  return BigInt(count);",
    "}",
    "function __gojrOsWrite(fd, source) {",
    "  const value = __gojrDerefIfPointer(source);",
    "  const bytes = __gojrBytesFrom(value);",
    "  const standardCount = __gojrWriteStandardStream(fd, bytes);",
    "  if (standardCount !== undefined) return BigInt(standardCount);",
    "  const hostWrite = globalThis.__gojrWriteSync;",
    "  if (typeof hostWrite === \"function\") return BigInt(Number(hostWrite(fd, bytes, 0, bytes.length, null) || bytes.length));",
    "  return BigInt(bytes.length);",
    "}",
    "function __gojrWriteStandardStream(fd, bytes) {",
    "  if (fd !== 1 && fd !== 2) return undefined;",
    "  const text = new TextDecoder().decode(bytes);",
    "  if (fd === 2) __gojrActiveRuntimeOptions.stderr?.(text);",
    "  else __gojrActiveRuntimeOptions.stdout?.(text);",
    "  return bytes.length;",
    "}",
    "function __gojrDefaultTextSink(stream) {",
    "  if (typeof process !== \"undefined\") return undefined;",
    "  const target = stream === \"stderr\" ? globalThis.console?.error || globalThis.console?.log : globalThis.console?.log;",
    "  if (typeof target !== \"function\") return undefined;",
    "  return (text) => target.call(globalThis.console, text);",
    "}",
    "function __gojrReflectlitePackage(importsByPath = {}) {",
    "  return {",
    "    Invalid: 0n, Interface: 20n, Ptr: 22n,",
    "    TypeOf: async (value) => __gojrReflectTypeOf(value, importsByPath),",
    "    ValueOf: async (value) => __gojrReflectValueOf(value, importsByPath),",
    "    Swapper: async (value) => __gojrReflectliteSwapper(value)",
    "  };",
    "}",
    "function __gojrReflectPackage(importsByPath = {}) {",
    "  return {",
    "    Invalid: 0n, Bool: 1n, Int: 2n, Int8: 3n, Int16: 4n, Int32: 5n, Int64: 6n, Uint: 7n, Uint8: 8n, Uint16: 9n, Uint32: 10n, Uint64: 11n, Uintptr: 12n, Float32: 13n, Float64: 14n, Complex64: 15n, Complex128: 16n, Array: 17n, Chan: 18n, Func: 19n, Interface: 20n, Map: 21n, Pointer: 22n, Ptr: 22n, Slice: 23n, String: 24n, Struct: 25n, UnsafePointer: 26n,",
    "    TypeOf: async (value) => __gojrReflectTypeOf(value, importsByPath),",
    "    TypeFor: async (typeArgs) => __gojrReflectTypeForTypeName(__gojrTypeArg(typeArgs, \"T\"), importsByPath),",
    "    PointerTo: async (typeValue) => __gojrReflectPointerTo(typeValue, importsByPath),",
    "    TypeAssert: async (typeArgs, value) => __gojrTypeAssertOk(value && value.__gojrReflectValue === true ? value.__gojrValue : value, __gojrTypeArg(typeArgs, \"T\")),",
    "    ValueOf: async (value) => __gojrReflectValueOf(value, importsByPath)",
    "  };",
    "}",
    "const __gojrInternalAbiTypeCache = new Map();",
    "function __gojrInternalAbiTypeOf(value) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  if (actual === null || actual === undefined) return null;",
    "  return __gojrInternalAbiTypePointer(__gojrRuntimeTypeName(actual), __gojrRuntimeTypePackagePath(actual));",
    "}",
    "function __gojrInternalAbiTypePointer(typeName, pkgPath = undefined) {",
    "  const text = String(typeName || \"any\").trim();",
    "  const descriptor = __gojrDescriptorForTypeName(text, __gojrActiveImportsByPath, pkgPath);",
    "  const descriptorType = descriptor && descriptor.type ? descriptor.type : text;",
    "  const descriptorPkgPath = descriptor && descriptor.pkgPath ? descriptor.pkgPath : (pkgPath || \"\");",
    "  const cacheKey = `${descriptorPkgPath}\\u0000${descriptorType}`;",
    "  const cached = __gojrInternalAbiTypeCache.get(cacheKey);",
    "  if (cached) return cached;",
    "  const kind = descriptor && descriptor.kind ? descriptor.kind : __gojrKindNameForType(text);",
    "  const size = __gojrSizeofType(text, __gojrActiveImportsByPath);",
    "  const align = BigInt(Math.min(8, Math.max(1, Number(size || 1n))));",
    "  const ptrBytes = text.startsWith(\"*\") || text.startsWith(\"map[\") || text.startsWith(\"chan \") || text.startsWith(\"<-chan\") || text.startsWith(\"chan<-\") || text.startsWith(\"func(\") ? 8n : 0n;",
    "  const value = __gojrStruct(\"Type\", {",
    "    Size_: size, PtrBytes: ptrBytes, Hash: __gojrStableTypeHash(`${descriptorPkgPath}:${descriptorType}`),",
    "    TFlag: descriptor && descriptor.name ? 4n : 0n, Align_: align, FieldAlign_: align, Kind_: __gojrReflectKindValue(kind),",
    "    Equal: async () => true, GCData: null, Str: 0n, PtrToThis: 0n,",
    "    GoJrTypeText: descriptorType, GoJrName: descriptor && descriptor.name || \"\", GoJrString: descriptor && descriptor.string || text, GoJrPkgPath: descriptorPkgPath",
    "  }, \"internal/abi\");",
    "  const pointer = __gojrPointerValue(\"Type\", value, \"internal/abi\");",
    "  __gojrInternalAbiTypeCache.set(cacheKey, pointer);",
    "  return pointer;",
    "}",
    "function __gojrInternalAbiMapType(typePointer) {",
    "  const target = __gojrDerefIfPointer(typePointer);",
    "  const typeText = target && typeof target.GoJrTypeText === \"string\" ? target.GoJrTypeText : __gojrRuntimeTypeName(target);",
    "  const mapType = __gojrMapKeyValueTypes(typeText);",
    "  if (!mapType) return null;",
    "  const keyType = __gojrInternalAbiTypePointer(mapType.keyType);",
    "  const elemType = __gojrInternalAbiTypePointer(mapType.valueType);",
    "  const typeValue = __gojrDerefIfPointer(typePointer);",
    "  const hasher = async (ptr, seed) => __gojrStableValueHash(__gojrDerefIfPointer(ptr), seed);",
    "  const value = __gojrStruct(\"MapType\", {",
    "    Type: typeValue, Key: keyType, Elem: elemType, Group: null, Hasher: hasher,",
    "    GroupSize: 0n, KeysOff: 0n, KeyStride: 0n, ElemsOff: 0n, ElemStride: 0n, ElemOff: 0n, Flags: 0n",
    "  }, \"internal/abi\");",
    "  return __gojrPointerValue(\"MapType\", value, \"internal/abi\");",
    "}",
    "function __gojrStableTypeHash(text) {",
    "  let hash = 2166136261;",
    "  const source = String(text || \"\");",
    "  for (let index = 0; index < source.length; index += 1) { hash ^= source.charCodeAt(index); hash = Math.imul(hash, 16777619) >>> 0; }",
    "  return BigInt(hash);",
    "}",
    "function __gojrStableValueHash(value, seed = 0n) {",
    "  let hash = Number((2166136261n ^ (BigInt(seed || 0) & 0xffffffffn)) & 0xffffffffn);",
    "  const source = __gojrStableValueString(value);",
    "  for (let index = 0; index < source.length; index += 1) { hash ^= source.charCodeAt(index); hash = Math.imul(hash, 16777619) >>> 0; }",
    "  return BigInt(hash);",
    "}",
    "function __gojrStableValueString(value, seen = new Set()) {",
    "  if (value && value.__gojrInterface === true) return __gojrStableValueString(value.value, seen);",
    "  if (value && value.__gojrPointer === true) return `&${__gojrStableValueString(__gojrDerefIfPointer(value), seen)}`;",
    "  if (value && value.__gojrTypedNil === true) return `nil:${value.__gojrType || \"\"}`;",
    "  if (value === null || value === undefined) return \"nil\";",
    "  if (typeof value === \"bigint\") return `${value}n`;",
    "  if (typeof value === \"number\" || typeof value === \"boolean\") return JSON.stringify(value);",
    "  if (__gojrIsString(value)) return JSON.stringify(Array.from(__gojrStringBytes(value)).map((byte) => byte.toString(16).padStart(2, \"0\")).join(\"\"));",
    "  if (value instanceof Uint8Array) return `bytes:${Array.from(value).join(\",\")}`;",
    "  if (Array.isArray(value)) return `[${value.map((item) => __gojrStableValueString(item, seen)).join(\",\")}]`;",
    "  if (value instanceof Map) return `map:${Array.from(value.entries()).map(([key, item]) => `${__gojrStableValueString(key, seen)}:${__gojrStableValueString(item, seen)}`).sort().join(\",\")}`;",
    "  if (typeof value === \"object\") {",
    "    if (seen.has(value)) return \"<cycle>\";",
    "    seen.add(value);",
    "    const keys = Object.keys(value).filter((key) => !key.startsWith(\"__gojr\")).sort();",
    "    return `{${keys.map((key) => `${key}:${__gojrStableValueString(value[key], seen)}`).join(\",\")}}`;",
    "  }",
    "  return String(value);",
    "}",
    "const __gojrReflectTypeCache = new Map();",
    "function __gojrReflectTypeOf(value, importsByPath) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  if (actual === null || actual === undefined) return null;",
    "  return __gojrReflectTypeForTypeName(__gojrRuntimeTypeName(actual), importsByPath, __gojrRuntimeTypePackagePath(actual));",
    "}",
    "function __gojrReflectTypeForTypeName(typeName, importsByPath = {}, pkgPath = undefined) {",
    "  const descriptor = __gojrDescriptorForTypeName(typeName, importsByPath, pkgPath);",
    "  if (!descriptor) return null;",
    "  const key = `${descriptor.pkgPath || pkgPath || \"\"}\\u0000${descriptor.type || typeName}`;",
    "  const cached = __gojrReflectTypeCache.get(key);",
    "  if (cached) return cached;",
    "  const out = { __gojrReflectType: true, __gojrDescriptor: descriptor, __gojrImportsByPath: importsByPath, __gojrMethods: __gojrReflectTypeMethods };",
    "  __gojrReflectTypeCache.set(key, out);",
    "  return out;",
    "}",
    "function __gojrReflectPointerTo(typeValue, importsByPath = {}) {",
    "  typeValue = __gojrUnwrapInterface(typeValue);",
    "  if (!typeValue || typeValue.__gojrReflectType !== true || !typeValue.__gojrDescriptor) return null;",
    "  const descriptor = typeValue.__gojrDescriptor;",
    "  return __gojrReflectTypeForTypeName(`*${descriptor.type}`, typeValue.__gojrImportsByPath || importsByPath, descriptor.pkgPath);",
    "}",
    "function __gojrDescriptorForTypeName(typeName, importsByPath = __gojrActiveImportsByPath, pkgPath = undefined) {",
    "  if (typeName === undefined || typeName === null || typeName === \"nil\") return undefined;",
    "  const text = String(typeName).trim();",
    "  if ((!pkgPath || pkgPath === gojrPackageArtifact.importPath) && __gojrTypeDescriptors[text]) return __gojrTypeDescriptors[text];",
    "  const intrinsicDescriptor = __gojrIntrinsicDescriptor(text, pkgPath);",
    "  if (intrinsicDescriptor) return intrinsicDescriptor;",
    "  if (text.startsWith(\"*\")) {",
    "    const elem = __gojrDescriptorForTypeName(text.slice(1).trim(), importsByPath, pkgPath);",
    "    return { type: text, name: \"\", string: `*${elem ? __gojrReflectDescriptorString(elem) : __gojrReceiverBaseType(text.slice(1).trim())}`, kind: \"ptr\", pkgPath: elem ? elem.pkgPath : \"\", pkgName: elem ? elem.pkgName : \"\", elem: elem ? elem.type : text.slice(1).trim(), elemPkgPath: elem ? elem.pkgPath : undefined };",
    "  }",
    "  const local = __gojrReceiverBaseType(text);",
    "  if ((!pkgPath || pkgPath === gojrPackageArtifact.importPath) && __gojrTypeDescriptors[local]) return __gojrTypeDescriptors[local];",
    "  const fixedArray = /^\\[([^\\]]+)\\](.+)$/.exec(text);",
    "  if (fixedArray && !text.startsWith(\"[]\")) {",
    "    const len = __gojrArrayLengthConstant(fixedArray[1]);",
    "    const elem = fixedArray[2].trim();",
    "    return { type: text, name: \"\", string: text, kind: \"array\", pkgPath: \"\", elem, ...(len === undefined ? {} : { len }) };",
    "  }",
    "  if (text.startsWith(\"[]\")) return { type: text, name: \"\", string: text, kind: \"slice\", pkgPath: \"\", elem: text.slice(2).trim() };",
    "  if (text.startsWith(\"map[\")) return { type: text, name: \"\", string: text, kind: \"map\", pkgPath: \"\" };",
    "  if (text.startsWith(\"chan \") || text.startsWith(\"<-chan\") || text.startsWith(\"chan<-\")) return { type: text, name: \"\", string: text, kind: \"chan\", pkgPath: \"\" };",
    "  if (text.startsWith(\"func(\")) return { type: text, name: \"\", string: text, kind: \"func\", pkgPath: \"\" };",
    "  const importedExact = __gojrImportedDescriptorForTypeName(text, importsByPath, pkgPath);",
    "  if (importedExact) return importedExact;",
    "  const importedLocal = __gojrImportedDescriptorForTypeName(local, importsByPath, pkgPath);",
    "  if (importedLocal) return importedLocal;",
    "  return { type: text, name: __gojrBuiltinTypeName(text), string: text, kind: __gojrKindNameForType(text), pkgPath: \"\" };",
    "}",
    "function __gojrImportedDescriptorForTypeName(typeName, importsByPath = {}, pkgPath = undefined) {",
    "  const entries = Object.entries(importsByPath || {});",
    "  for (const [path, imported] of entries) {",
    "    if (pkgPath && path !== pkgPath) continue;",
    "    if (!pkgPath && path === gojrPackageArtifact.importPath) continue;",
    "    const importedPackage = imported && typeof imported === \"object\" && \"package\" in imported ? imported.package : imported;",
    "    const table = importedPackage && importedPackage.__gojrTypeDescriptors;",
    "    if (!table) continue;",
    "    if (table[typeName]) return table[typeName];",
    "    const local = __gojrReceiverBaseType(typeName);",
    "    if (table[local]) return table[local];",
    "  }",
    "  return undefined;",
    "}",
    "function __gojrArrayLengthConstant(text) {",
    "  const source = String(text || \"\").trim();",
    "  let offset = 0;",
    "  const skipSpace = () => { while (offset < source.length && /\\s/.test(source[offset])) offset += 1; };",
    "  const consume = (text) => { skipSpace(); if (!source.startsWith(text, offset)) return false; offset += text.length; return true; };",
    "  const parseInteger = () => {",
    "    skipSpace();",
    "    const start = offset;",
    "    if (source.startsWith(\"0x\", start) || source.startsWith(\"0X\", start)) { offset += 2; while (offset < source.length && /[0-9a-fA-F_]/.test(source[offset])) offset += 1; return __gojrParseArrayLengthInteger(source.slice(start, offset)); }",
    "    if (source.startsWith(\"0b\", start) || source.startsWith(\"0B\", start)) { offset += 2; while (offset < source.length && /[01_]/.test(source[offset])) offset += 1; return __gojrParseArrayLengthInteger(source.slice(start, offset)); }",
    "    if (source.startsWith(\"0o\", start) || source.startsWith(\"0O\", start)) { offset += 2; while (offset < source.length && /[0-7_]/.test(source[offset])) offset += 1; return __gojrParseArrayLengthInteger(source.slice(start, offset)); }",
    "    while (offset < source.length && /[0-9_]/.test(source[offset])) offset += 1;",
    "    return offset === start ? undefined : __gojrParseArrayLengthInteger(source.slice(start, offset));",
    "  };",
    "  const parseIdentifierSegment = () => {",
    "    skipSpace();",
    "    const start = offset;",
    "    if (start >= source.length || !/[A-Za-z_]/.test(source[start])) return undefined;",
    "    offset += 1;",
    "    while (offset < source.length && /[A-Za-z0-9_]/.test(source[offset])) offset += 1;",
    "    return source.slice(start, offset);",
    "  };",
    "  const parseIdentifier = () => {",
    "    const first = parseIdentifierSegment();",
    "    if (!first) return undefined;",
    "    let name = first;",
    "    while (source[offset] === \".\") {",
    "      const dot = offset;",
    "      offset += 1;",
    "      const next = parseIdentifierSegment();",
    "      if (!next) { offset = dot; break; }",
    "      name += `.${next}`;",
    "    }",
    "    return name;",
    "  };",
    "  const lookupIdentifier = (name) => {",
    "    const parts = String(name).split(\".\");",
    "    const root = parts[0];",
    "    let value = Object.prototype.hasOwnProperty.call(__gojrActivePackage, root) ? __gojrActivePackage[root] : undefined;",
    "    if (value === undefined && parts.length > 1) {",
    "      const imported = __gojrActiveImportsByPath[root];",
    "      const importedPackage = imported && typeof imported === \"object\" && \"package\" in imported ? imported.package : imported;",
    "      value = importedPackage;",
    "    }",
    "    for (const part of parts.slice(value === undefined ? 0 : 1)) {",
    "      if (value === undefined || value === null) return undefined;",
    "      value = value[part];",
    "    }",
    "    if (value && value.__gojrNamed !== undefined) value = value.value;",
    "    if (typeof value === \"bigint\") return value;",
    "    if (typeof value === \"number\" && Number.isSafeInteger(value)) return BigInt(value);",
    "    return undefined;",
    "  };",
    "  const apply = (op, left, right) => {",
    "    switch (op) {",
    "      case \"+\": return left + right; case \"-\": return left - right; case \"*\": return left * right;",
    "      case \"/\": return right === 0n ? undefined : left / right; case \"%\": return right === 0n ? undefined : left % right;",
    "      case \"|\": return left | right; case \"^\": return left ^ right; case \"&\": return left & right; case \"&^\": return left & ~right;",
    "      case \"<<\": return right < 0n || right > 1024n ? undefined : left << right; case \">>\": return right < 0n || right > 1024n ? undefined : left >> right;",
    "      default: return undefined;",
    "    }",
    "  };",
    "  const binaryOperator = () => {",
    "    const operators = [[\"&^\", 5], [\"<<\", 5], [\">>\", 5], [\"*\", 5], [\"/\", 5], [\"%\", 5], [\"&\", 5], [\"+\", 4], [\"-\", 4], [\"|\", 4], [\"^\", 4]];",
    "    const op = operators.find(([text]) => source.startsWith(text, offset));",
    "    return op ? { text: op[0], precedence: op[1] } : undefined;",
    "  };",
    "  const primary = () => {",
    "    skipSpace();",
    "    if (consume(\"(\")) { const value = expression(1); skipSpace(); return consume(\")\") ? value : undefined; }",
    "    const integer = parseInteger();",
    "    if (integer !== undefined) return integer;",
    "    const identifier = parseIdentifier();",
    "    return identifier ? lookupIdentifier(identifier) : undefined;",
    "  };",
    "  const unary = () => {",
    "    skipSpace();",
    "    if (consume(\"+\")) return unary();",
    "    if (consume(\"-\")) { const value = unary(); return value === undefined ? undefined : -value; }",
    "    if (consume(\"^\")) { const value = unary(); return value === undefined ? undefined : ~value; }",
    "    return primary();",
    "  };",
    "  const expression = (minPrecedence) => {",
    "    let left = unary();",
    "    if (left === undefined) return undefined;",
    "    while (true) {",
    "      skipSpace();",
    "      const op = binaryOperator();",
    "      if (!op || op.precedence < minPrecedence) return left;",
    "      offset += op.text.length;",
    "      const right = expression(op.precedence + 1);",
    "      if (right === undefined) return undefined;",
    "      left = apply(op.text, left, right);",
    "      if (left === undefined) return undefined;",
    "    }",
    "  };",
    "  const value = expression(1);",
    "  skipSpace();",
    "  return offset === source.length && value !== undefined && value >= 0n && value <= BigInt(Number.MAX_SAFE_INTEGER) ? Number(value) : undefined;",
    "}",
    "function __gojrParseArrayLengthInteger(text) {",
    "  const raw = String(text || \"\").replace(/_/g, \"\");",
    "  try {",
    "    if (/^0[xX][0-9a-fA-F]+$/.test(raw)) return BigInt(raw);",
    "    if (/^0[bB][01]+$/.test(raw)) return BigInt(`0b${raw.slice(2)}`);",
    "    if (/^0[oO][0-7]+$/.test(raw)) return BigInt(`0o${raw.slice(2)}`);",
    "    if (/^0[0-7]*$/.test(raw) && raw.length > 1) return BigInt(`0o${raw.slice(1) || \"0\"}`);",
    "    if (/^[0-9]+$/.test(raw)) return BigInt(raw);",
    "  } catch {",
    "    return undefined;",
    "  }",
    "  return undefined;",
    "}",
    "function __gojrReflectDescriptorString(descriptor) {",
    "  if (!descriptor) return \"\";",
    "  if (descriptor.name && descriptor.pkgPath && descriptor.pkgPath !== gojrPackageArtifact.importPath) return `${descriptor.pkgName || __gojrImportPathBase(descriptor.pkgPath)}.${descriptor.name}`;",
    "  return descriptor.string;",
    "}",
    "function __gojrImportPathBase(path) {",
    "  const parts = String(path || \"\").split(\"/\").filter(Boolean);",
    "  return parts[parts.length - 1] || \"\";",
    "}",
    "const __gojrReflectTypeMethods = {",
    "  String: async (self) => __gojrReflectDescriptorString(self.__gojrDescriptor),",
    "  Name: async (self) => self.__gojrDescriptor.name || \"\",",
    "  PkgPath: async (self) => self.__gojrDescriptor.pkgPath || \"\",",
    "  Kind: async (self) => __gojrReflectKindValue(self.__gojrDescriptor.kind),",
    "  NumField: async (self) => BigInt((self.__gojrDescriptor.fields || []).length),",
    "  Field: async (self, index) => {",
    "    const field = (self.__gojrDescriptor.fields || [])[Number(index)];",
    "    if (!field) throw new RangeError(\"reflect: Field index out of range\");",
    "    return __gojrStruct(\"reflect.StructField\", { Name: field.name, Type: __gojrReflectTypeForTypeName(field.type, self.__gojrImportsByPath, field.pkgPath), Tag: field.tag || \"\", Index: [BigInt(Number(index))], Anonymous: Boolean(field.embedded) });",
    "  },",
    "  Elem: async (self) => {",
    "    const elem = self.__gojrDescriptor.elem;",
    "    if (!elem) throw new TypeError(`reflect: Elem of invalid type ${self.__gojrDescriptor.string}`);",
    "    return __gojrReflectTypeForTypeName(elem, self.__gojrImportsByPath, self.__gojrDescriptor.elemPkgPath);",
    "  },",
    "  Key: async (self) => {",
    "    const key = self.__gojrDescriptor.key;",
    "    if (!key) throw new TypeError(`reflect: Key of invalid type ${self.__gojrDescriptor.string}`);",
    "    return __gojrReflectTypeForTypeName(key, self.__gojrImportsByPath, self.__gojrDescriptor.keyPkgPath);",
    "  },",
    "  Len: async (self) => {",
    "    if (typeof self.__gojrDescriptor.len !== \"number\") throw new TypeError(`reflect: Len of invalid type ${self.__gojrDescriptor.string}`);",
    "    return BigInt(self.__gojrDescriptor.len);",
    "  },",
    "  NumIn: async (self) => BigInt((self.__gojrDescriptor.params || []).length),",
    "  In: async (self, index) => {",
    "    const typeName = (self.__gojrDescriptor.params || [])[Number(index)];",
    "    if (!typeName) throw new RangeError(\"reflect: In index out of range\");",
    "    return __gojrReflectTypeForTypeName(typeName, self.__gojrImportsByPath);",
    "  },",
    "  NumOut: async (self) => BigInt((self.__gojrDescriptor.results || []).length),",
    "  Out: async (self, index) => {",
    "    const typeName = (self.__gojrDescriptor.results || [])[Number(index)];",
    "    if (!typeName) throw new RangeError(\"reflect: Out index out of range\");",
    "    return __gojrReflectTypeForTypeName(typeName, self.__gojrImportsByPath);",
    "  },",
    "  Comparable: async () => true,",
    "  AssignableTo: async () => true,",
    "  Implements: async (self, target) => __gojrReflectTypeImplements(self, target)",
    "};",
    "function __gojrReflectTypeImplements(self, target) {",
    "  self = __gojrUnwrapInterface(self);",
    "  target = __gojrUnwrapInterface(target);",
    "  if (!self || !target || self.__gojrReflectType !== true || target.__gojrReflectType !== true) return false;",
    "  const required = __gojrReflectInterfaceMethods(target.__gojrDescriptor, target.__gojrImportsByPath);",
    "  if (required === undefined) return false;",
    "  const available = __gojrReflectConcreteMethods(self.__gojrDescriptor, self.__gojrImportsByPath);",
    "  return required.every((method) => __gojrReflectMethodSetHas(available, method));",
    "}",
    "function __gojrReflectInterfaceMethods(descriptor, importsByPath = {}, seen = new Set()) {",
    "  if (!descriptor) return undefined;",
    "  const type = descriptor.type || descriptor.string || \"\";",
    "  if (type === \"any\" || type === \"interface{}\") return [];",
    "  if (type === \"error\") return [{ name: \"Error\", params: [], results: [\"string\"] }];",
    "  if (descriptor.kind !== \"interface\") return undefined;",
    "  const key = `${descriptor.pkgPath || \"\"}\\u0000${type}`;",
    "  if (seen.has(key)) return [];",
    "  seen.add(key);",
    "  const methods = [...(descriptor.interfaceMethods || [])];",
    "  for (const embedded of descriptor.interfaceEmbeds || []) {",
    "    const embeddedDescriptor = __gojrDescriptorForTypeName(embedded, importsByPath, descriptor.pkgPath);",
    "    const embeddedMethods = __gojrReflectInterfaceMethods(embeddedDescriptor, importsByPath, seen);",
    "    if (embeddedMethods) methods.push(...embeddedMethods);",
    "  }",
    "  return methods;",
    "}",
    "function __gojrReflectConcreteMethods(descriptor, importsByPath = {}, seen = new Set(), includePointerReceivers = false) {",
    "  if (!descriptor) return [];",
    "  if (descriptor.kind === \"interface\") return __gojrReflectInterfaceMethods(descriptor, importsByPath) || [];",
    "  if (descriptor.kind === \"ptr\") {",
    "    const elem = __gojrDescriptorForTypeName(descriptor.elem, importsByPath, descriptor.elemPkgPath || descriptor.pkgPath);",
    "    return __gojrReflectConcreteMethods(elem, importsByPath, seen, true);",
    "  }",
    "  const key = `${descriptor.pkgPath || \"\"}\\u0000${descriptor.type || descriptor.string || \"\"}`;",
    "  if (seen.has(key)) return [];",
    "  seen.add(key);",
    "  return (descriptor.methods || []).filter((method) => includePointerReceivers || !method.pointerReceiver);",
    "}",
    "function __gojrReflectMethodSetHas(methods, required) {",
    "  return methods.some((method) => method.name === required.name && __gojrReflectMethodSignatureMatches(method, required));",
    "}",
    "function __gojrReflectMethodSignatureMatches(actual, required) {",
    "  const actualParams = actual.params || [];",
    "  const requiredParams = required.params || [];",
    "  const actualResults = actual.results || [];",
    "  const requiredResults = required.results || [];",
    "  return __gojrReflectTypeListMatches(actualParams, requiredParams) && __gojrReflectTypeListMatches(actualResults, requiredResults);",
    "}",
    "function __gojrReflectTypeListMatches(left, right) {",
    "  if (left.length !== right.length) return false;",
    "  for (let index = 0; index < left.length; index += 1) {",
    "    if (String(left[index]).trim() !== String(right[index]).trim()) return false;",
    "  }",
    "  return true;",
    "}",
    "function __gojrReflectliteSwapper(value) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  if (!Array.isArray(actual) && !ArrayBuffer.isView(actual)) throw new TypeError(\"reflectlite.Swapper expects a slice or array\");",
    "  return async (i, j) => {",
    "    const left = Number(i);",
    "    const right = Number(j);",
    "    const tmp = actual[left];",
    "    actual[left] = actual[right];",
    "    actual[right] = tmp;",
    "    return null;",
    "  };",
    "}",
    "function __gojrReflectValueOf(value, importsByPath = {}) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  if (actual === null || actual === undefined) return { __gojrReflectValue: true, __gojrValid: false, __gojrValue: undefined, __gojrImportsByPath: importsByPath, __gojrMethods: __gojrReflectValueMethods };",
    "  return { __gojrReflectValue: true, __gojrValid: true, __gojrValue: actual, __gojrImportsByPath: importsByPath, __gojrMethods: __gojrReflectValueMethods };",
    "}",
    "function __gojrReflectValueType(self) {",
    "  if (!self.__gojrValid) return null;",
    "  return __gojrReflectTypeForTypeName(__gojrRuntimeTypeName(self.__gojrValue), self.__gojrImportsByPath, __gojrRuntimeTypePackagePath(self.__gojrValue));",
    "}",
    "const __gojrReflectValueMethods = {",
    "  IsValid: async (self) => Boolean(self.__gojrValid),",
    "  IsNil: async (self) => {",
    "    if (!self.__gojrValid) throw new TypeError(\"reflect: call of Value.IsNil on zero Value\");",
    "    const value = self.__gojrValue;",
    "    return value === null || value === undefined || Boolean(value && value.__gojrTypedNil === true);",
    "  },",
    "  Kind: async (self) => {",
    "    const typ = __gojrReflectValueType(self);",
    "    return typ ? __gojrReflectKindValue(typ.__gojrDescriptor.kind) : 0n;",
    "  },",
    "  Type: async (self) => {",
    "    const typ = __gojrReflectValueType(self);",
    "    if (!typ) throw new TypeError(\"reflect: call of Value.Type on zero Value\");",
    "    return typ;",
    "  },",
    "  Field: async (self, index) => {",
    "    const typ = __gojrReflectValueType(self);",
    "    const field = typ && (typ.__gojrDescriptor.fields || [])[Number(index)];",
    "    if (!field) throw new RangeError(\"reflect: Field index out of range\");",
    "    return __gojrReflectValueOf(self.__gojrValue[field.name], self.__gojrImportsByPath);",
    "  },",
    "  Interface: async (self) => {",
    "    if (!self.__gojrValid) throw new TypeError(\"reflect: call of Value.Interface on zero Value\");",
    "    return self.__gojrValue;",
    "  },",
    "  String: async (self) => {",
    "    if (!self.__gojrValid) return \"<invalid Value>\";",
    "    return String(self.__gojrValue);",
    "  },",
    "  Int: async (self) => {",
    "    if (!self.__gojrValid) throw new TypeError(\"reflect: call of Value.Int on zero Value\");",
    "    return BigInt(self.__gojrValue);",
    "  },",
    "  Bool: async (self) => {",
    "    if (!self.__gojrValid) throw new TypeError(\"reflect: call of Value.Bool on zero Value\");",
    "    return Boolean(self.__gojrValue);",
    "  }",
    "};",
    "function __gojrBuiltinTypeName(typeName) {",
    "  switch (typeName) {",
    "    case \"bool\": case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": case \"float32\": case \"float64\": case \"complex64\": case \"complex128\": case \"string\": return typeName;",
    "    default: return \"\";",
    "  }",
    "}",
    "function __gojrKindNameForType(typeName) {",
    "  if (typeName.startsWith(\"*\")) return \"ptr\";",
    "  if (/^\\[[^\\]]+\\]/.test(typeName) && !typeName.startsWith(\"[]\")) return \"array\";",
    "  if (typeName.startsWith(\"[]\")) return \"slice\";",
    "  if (typeName.startsWith(\"map[\")) return \"map\";",
    "  if (typeName.startsWith(\"chan \") || typeName.startsWith(\"<-chan\") || typeName.startsWith(\"chan<-\")) return \"chan\";",
    "  if (typeName.startsWith(\"func(\")) return \"func\";",
    "  if (typeName === \"any\" || typeName === \"error\" || typeName === \"interface{}\" || typeName.startsWith(\"interface{\")) return \"interface\";",
    "  return typeName;",
    "}",
    "function __gojrReflectKindValue(kind) {",
    "  switch (kind) {",
    "    case \"bool\": return 1n; case \"int\": return 2n; case \"int8\": return 3n; case \"int16\": return 4n; case \"int32\": return 5n; case \"int64\": return 6n;",
    "    case \"uint\": return 7n; case \"uint8\": case \"byte\": return 8n; case \"uint16\": return 9n; case \"uint32\": return 10n; case \"uint64\": return 11n; case \"uintptr\": return 12n;",
    "    case \"rune\": return 5n;",
    "    case \"float32\": return 13n; case \"float64\": return 14n; case \"complex64\": return 15n; case \"complex128\": return 16n;",
    "    case \"array\": return 17n; case \"chan\": return 18n; case \"func\": return 19n; case \"interface\": return 20n; case \"map\": return 21n; case \"ptr\": return 22n; case \"slice\": return 23n; case \"string\": return 24n; case \"struct\": return 25n;",
    "    default: return 0n;",
    "  }",
    "}",
    "function __gojrStruct(typeName, value, pkgPath = gojrPackageArtifact.importPath) {",
    "  const intrinsic = __gojrIntrinsicStruct(typeName, value, pkgPath);",
    "  if (intrinsic !== undefined) return intrinsic;",
    "  const descriptor = __gojrDescriptorForTypeName(typeName, __gojrActiveImportsByPath, pkgPath);",
    "  const fields = descriptor && Array.isArray(descriptor.fields) ? descriptor.fields : [];",
    "  for (const field of fields) {",
    "    if (!Object.prototype.hasOwnProperty.call(value, field.name)) value[field.name] = __gojrZero(field.type, field.pkgPath);",
    "  }",
    "  Object.defineProperty(value, \"__gojrType\", { value: typeName });",
    "  Object.defineProperty(value, \"__gojrPkgPath\", { value: pkgPath || gojrPackageArtifact.importPath });",
    "  return value;",
    "}",
    "function __gojrStructFromValues(typeName, values) {",
    "  const descriptor = __gojrDescriptorForTypeName(typeName);",
    "  const fields = descriptor && Array.isArray(descriptor.fields) ? descriptor.fields : [];",
    "  const out = {};",
    "  for (let index = 0; index < fields.length; index += 1) {",
    "    const field = fields[index];",
    "    out[field.name] = index < values.length ? values[index] : __gojrZero(field.type, field.pkgPath);",
    "  }",
    "  return __gojrStruct(typeName, out, descriptor && descriptor.pkgPath);",
    "}",
    "function __gojrPointer(typeName, get, set, pkgPath = gojrPackageArtifact.importPath, sequence = undefined, owners = undefined, identity = undefined) {",
    "  const pointer = { __gojrPointer: true, __gojrType: `*${typeName}`, __gojrElemType: typeName, __gojrPkgPath: pkgPath || gojrPackageArtifact.importPath, __gojrGet: get, __gojrSet: set };",
    "  if (sequence) Object.defineProperty(pointer, \"__gojrSequence\", { value: { values: sequence.values, index: Number(sequence.index || 0) } });",
    "  if (owners && owners.length) Object.defineProperty(pointer, \"__gojrOwners\", { value: owners });",
    "  if (identity !== undefined) Object.defineProperty(pointer, \"__gojrIdentity\", { value: String(identity) });",
    "  return pointer;",
    "}",
    "function __gojrPointerValue(typeName, value, pkgPath = gojrPackageArtifact.importPath) {",
    "  let cell = value;",
    "  return __gojrPointer(typeName, () => cell, (next) => { cell = next; }, pkgPath);",
    "}",
    "function __gojrConvertPointer(typeName, value, pkgPath = undefined) {",
    "  const targetPkgPath = pkgPath !== undefined ? pkgPath : gojrPackageArtifact.importPath;",
    "  if (value === null || value === undefined) return __gojrTypedNil(typeName, targetPkgPath);",
    "  if (value && value.__gojrTypedNil === true) return __gojrTypedNil(typeName, targetPkgPath);",
  "  if (value && value.__gojrPointer === true) {",
  "    const elemType = String(typeName).startsWith(\"*\") ? String(typeName).slice(1).trim() : String(typeName);",
    "    const sourceElemType = value.__gojrSourceElemType || value.__gojrElemType;",
    "    const owner = __gojrPointerOwnerForType(value, elemType, targetPkgPath);",
    "    const identity = value.__gojrIdentity !== undefined ? value.__gojrIdentity : __gojrPointerIdentityKey(value);",
    "    if (owner) return __gojrPointer(elemType, owner.get, owner.set, owner.pkgPath || targetPkgPath || value.__gojrPkgPath, undefined, owner.owners, identity);",
    "    const byteLength = __gojrFixedByteArrayLength(elemType);",
    "    if (byteLength !== undefined && __gojrIntegerTypeName(sourceElemType)) return __gojrPointerByteArrayView(elemType, value, sourceElemType, byteLength, identity);",
    "    const pointer = __gojrPointer(elemType, value.__gojrGet, value.__gojrSet, targetPkgPath || value.__gojrPkgPath, value.__gojrSequence, value.__gojrOwners, identity);",
    "    if (sourceElemType && elemType === \"unsafe.Pointer\") Object.defineProperty(pointer, \"__gojrSourceElemType\", { value: sourceElemType });",
    "    return pointer;",
  "  }",
  "  return value;",
  "}",
    "function __gojrPointerByteArrayView(typeName, sourcePointer, scalarType, length, identity = undefined) {",
    "  const view = __gojrIntegerByteView(sourcePointer, scalarType, length);",
    "  return __gojrPointer(typeName, () => view, (next) => {",
    "    let raw = 0n;",
    "    for (let index = 0; index < length; index += 1) raw |= (BigInt(Number(next && next[index] !== undefined ? next[index] : 0) & 255) << BigInt(index * 8));",
    "    sourcePointer.__gojrSet(__gojrIntegerConvert(scalarType, raw));",
    "  }, sourcePointer.__gojrPkgPath, { values: view, index: 0 }, sourcePointer.__gojrOwners, identity);",
    "}",
    "function __gojrIntegerByteView(sourcePointer, scalarType, length) {",
    "  const view = { __gojrByteView: true, length, __gojrCap: length };",
    "  const getByte = (index) => Number((__gojrToBigInt(sourcePointer.__gojrGet() ?? 0n) >> BigInt(index * 8)) & 255n);",
    "  const setByte = (index, value) => {",
    "    const shift = BigInt(index * 8);",
    "    const mask = 255n << shift;",
    "    const raw = (__gojrToBigInt(sourcePointer.__gojrGet() ?? 0n) & ~mask) | ((BigInt(Number(value) & 255) << shift) & mask);",
    "    sourcePointer.__gojrSet(__gojrIntegerConvert(scalarType, raw));",
    "  };",
    "  Object.defineProperty(view, \"subarray\", { value(start = 0, end = length) {",
    "    const low = Math.max(0, Number(start));",
    "    const high = Math.min(length, Number(end));",
    "    const out = new Uint8Array(Math.max(0, high - low));",
    "    for (let index = 0; index < out.length; index += 1) out[index] = getByte(low + index);",
    "    return __gojrSetCap(out, length - low);",
    "  } });",
    "  Object.defineProperty(view, \"slice\", { value: view.subarray });",
    "  for (let index = 0; index < length; index += 1) Object.defineProperty(view, index, {",
    "    enumerable: true, configurable: true,",
    "    get() { return getByte(index); },",
    "    set(value) { setByte(index, value); }",
    "  });",
    "  return view;",
    "}",
    "function __gojrPointerOwnerForType(pointer, elemType, pkgPath = undefined) {",
    "  const target = __gojrNormalizeTypeIdentity(elemType);",
    "  const targetPkgPath = pkgPath || \"\";",
    "  for (const owner of pointer && pointer.__gojrOwners || []) {",
    "    if (targetPkgPath && owner.pkgPath && owner.pkgPath !== targetPkgPath) continue;",
    "    if (__gojrNormalizeTypeIdentity(owner.typeName) === target) return owner;",
    "  }",
    "  return undefined;",
    "}",
    "function __gojrNormalizeTypeIdentity(typeName) {",
    "  let text = String(typeName || \"\").trim();",
    "  if (text.startsWith(\"*\")) text = text.slice(1).trim();",
    "  const dot = text.lastIndexOf(\".\");",
    "  if (dot >= 0) text = text.slice(dot + 1);",
    "  return text.replace(/\\s+/g, \"\");",
    "}",
    "function __gojrNilUnsafePointer(value) {",
    "  return value === null || value === undefined || (value && value.__gojrTypedNil === true && (String(value.__gojrType || \"\").startsWith(\"*\") || String(value.__gojrType || \"\") === \"unsafe.Pointer\"));",
    "}",
    "function __gojrUnsafeAdd(ptr, offset) {",
    "  if (__gojrNilUnsafePointer(ptr)) return ptr;",
    "  const bytes = Number(offset || 0);",
    "  const sequence = __gojrPointerSequence(ptr);",
    "  if (!sequence || !ptr || ptr.__gojrPointer !== true) return ptr;",
    "  const elemType = ptr.__gojrElemType || \"byte\";",
    "  const elemSize = Math.max(1, Number(__gojrSizeofType(elemType)));",
    "  const nextIndex = Number(sequence.index || 0) + Math.trunc(bytes / elemSize);",
    "  return __gojrPointer(elemType, () => sequence.values[nextIndex], (next) => { sequence.values[nextIndex] = next; }, ptr.__gojrPkgPath, { values: sequence.values, index: nextIndex }, ptr.__gojrOwners);",
    "}",
    "function __gojrPointerSequence(value) {",
    "  if (value && value.__gojrPointer === true && value.__gojrSequence) return { values: value.__gojrSequence.values, index: Number(value.__gojrSequence.index || 0) };",
    "  return undefined;",
    "}",
    "function __gojrBytesFromSequence(sequence, length) {",
    "  const values = sequence && sequence.values;",
    "  const start = Number(sequence && sequence.index || 0);",
    "  const count = Number(length || 0);",
    "  if (count <= 0) return new Uint8Array();",
    "  if (values instanceof Uint8Array) return values.subarray(start, start + count);",
    "  if (ArrayBuffer.isView(values)) return Uint8Array.from(Array.from(values).slice(start, start + count), (item) => Number(item) & 255);",
    "  if (Array.isArray(values)) return Uint8Array.from(values.slice(start, start + count), (item) => Number(item) & 255);",
    "  return new Uint8Array(count);",
    "}",
    "function __gojrSliceFromSequence(sequence, length) {",
    "  const values = sequence && sequence.values;",
    "  const start = Number(sequence && sequence.index || 0);",
    "  const count = Number(length || 0);",
    "  if (values instanceof Uint8Array) return values.subarray(start, start + count);",
    "  if (ArrayBuffer.isView(values)) return Array.from(values).slice(start, start + count);",
    "  if (Array.isArray(values)) return __gojrSetCap(values.slice(start, start + count), count);",
    "  return __gojrSetCap(Array.from({ length: count }, () => 0n), count);",
    "}",
    "function __gojrSliceElementRuntimeType(value) {",
    "  const typeName = value && value.__gojrType;",
    "  if (typeof typeName === \"string\" && typeName.startsWith(\"[]\")) return typeName.slice(2).trim();",
    "  if (value instanceof Uint8Array) return \"byte\";",
    "  if (Array.isArray(value) && value.length > 0) return __gojrRuntimeTypeName(value[0]);",
    "  return \"byte\";",
    "}",
    "function __gojrPointerLike(value) {",
    "  return Boolean(value && (value.__gojrPointer === true || value.__gojrTypedNil === true || String(value.__gojrType || \"\") === \"unsafe.Pointer\"));",
    "}",
    "function __gojrPointerToUintptr(value) {",
    "  return __gojrPointerLike(value) ? value : BigInt(value);",
    "}",
    "function __gojrIsZeroInteger(value) {",
    "  if (typeof value === \"bigint\") return value === 0n;",
    "  if (typeof value === \"number\") return value === 0;",
    "  return false;",
    "}",
    "function __gojrBitwiseNot(value, typeName) {",
    "  const bits = __gojrUnsignedIntegerBitWidth(typeName);",
    "  const raw = __gojrToBigInt(value);",
    "  return bits > 0 ? (((1n << BigInt(bits)) - 1n) ^ raw) : ~raw;",
    "}",
    "function __gojrUnsignedIntegerBitWidth(typeName) {",
    "  switch (String(typeName || \"\")) {",
    "    case \"uint8\": case \"byte\": return 8;",
    "    case \"uint16\": return 16;",
    "    case \"uint32\": return 32;",
    "    case \"uint\": case \"uint64\": case \"uintptr\": return 64;",
    "    default: return 0;",
    "  }",
    "}",
    "function __gojrNilPointerError() {",
    "  return new Error(\"GOJR_RUNTIME001: invalid memory address or nil pointer dereference\");",
    "}",
    "function __gojrDeref(pointer) {",
    "  if (pointer && pointer.__gojrPointer === true) return pointer.__gojrGet();",
    "  if (pointer && pointer.__gojrTypedNil === true && String(pointer.__gojrType || \"\").startsWith(\"*\")) throw __gojrNilPointerError();",
    "  throw new Error(\"GOJR_RUNTIME001: cannot dereference non-pointer\");",
    "}",
    "function __gojrSetDeref(pointer, value) {",
    "  if (pointer && pointer.__gojrPointer === true) { pointer.__gojrSet(value); return null; }",
    "  if (pointer && pointer.__gojrTypedNil === true && String(pointer.__gojrType || \"\").startsWith(\"*\")) throw __gojrNilPointerError();",
    "  throw new Error(\"GOJR_RUNTIME001: cannot assign through non-pointer\");",
    "}",
    "function __gojrDerefIfPointer(value) {",
    "  return value && value.__gojrPointer === true ? __gojrDeref(value) : value;",
    "}",
    "function __gojrOwnField(object, field) {",
    "  return object != null && Object.prototype.hasOwnProperty.call(object, field);",
    "}",
    "function __gojrPromotedFieldCell(object, field, seen = new Set()) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  if (target == null || typeof target !== \"object\" || seen.has(target)) return undefined;",
    "  seen.add(target);",
    "  const descriptor = __gojrDescriptorForTypeName(__gojrRuntimeTypeName(target), __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(target));",
    "  for (const descriptorField of descriptor && descriptor.fields || []) {",
    "    if (!descriptorField.embedded) continue;",
    "    const embedded = target[descriptorField.name];",
    "    const embeddedTarget = __gojrDerefIfPointer(embedded);",
    "    if (embeddedTarget != null && typeof embeddedTarget === \"object\" && __gojrOwnField(embeddedTarget, field)) {",
    "      return { get: () => embeddedTarget[field], set: (next) => { embeddedTarget[field] = next; } };",
    "    }",
    "    const nested = __gojrPromotedFieldCell(embedded, field, seen);",
    "    if (nested) return nested;",
    "  }",
    "  return undefined;",
    "}",
    "function __gojrGetField(object, field) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  if (target && target.__gojrTypedNil === true && String(target.__gojrType || \"\").startsWith(\"*\")) throw __gojrNilPointerError();",
    "  if (target == null) throw new Error(`GOJR_RUNTIME001: cannot select field ${field} on nil`);",
    "  if (__gojrOwnField(target, field)) return target[field];",
    "  const promoted = __gojrPromotedFieldCell(target, field);",
    "  if (promoted) return promoted.get();",
    "  return target[field];",
    "}",
    "function __gojrSetField(object, field, value) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  if (target == null) throw new Error(`GOJR_RUNTIME001: cannot set field ${field} on nil`);",
    "  if (!__gojrOwnField(target, field)) {",
    "    const promoted = __gojrPromotedFieldCell(target, field);",
    "    if (promoted) { promoted.set(value); return null; }",
    "  }",
    "  target[field] = value;",
    "  return null;",
    "}",
    "function __gojrAddressField(object, field, typeName, pkgPath = gojrPackageArtifact.importPath) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  const owners = __gojrZeroOffsetOwnersForField(object, target, field);",
    "  return __gojrPointer(typeName, () => __gojrGetField(object, field), (next) => { __gojrSetField(object, field, next); }, pkgPath, undefined, owners);",
    "}",
    "function __gojrZeroOffsetOwnersForField(object, target, field) {",
    "  if (target == null || typeof target !== \"object\") return undefined;",
    "  const typeName = __gojrRuntimeTypeName(target);",
    "  const pkgPath = __gojrRuntimeTypePackagePath(target);",
    "  const descriptor = __gojrDescriptorForTypeName(typeName, __gojrActiveImportsByPath, pkgPath);",
    "  const first = descriptor && Array.isArray(descriptor.fields) ? descriptor.fields[0] : undefined;",
    "  if (!first || first.name !== field) return object && object.__gojrOwners;",
    "  const owner = {",
    "    typeName,",
    "    pkgPath: pkgPath || gojrPackageArtifact.importPath,",
    "    get: () => target,",
    "    set: (next) => {",
    "      if (object && object.__gojrPointer === true) { __gojrSetDeref(object, next); return; }",
    "      if (next && typeof next === \"object\") {",
    "        for (const key of Object.keys(target)) delete target[key];",
    "        Object.assign(target, next);",
    "      }",
    "    }",
    "  };",
    "  return [owner, ...(object && object.__gojrOwners || [])];",
  "}",
    "function __gojrAddressIndex(object, index, typeName, pkgPath = gojrPackageArtifact.importPath) {",
    "  const target = __gojrDerefIfPointer(object);",
    "  const numericIndex = Number(index);",
    "  if (target == null) throw new Error(\"GOJR_RUNTIME001: cannot take address of nil index target\");",
    "  return __gojrPointer(typeName, () => target[numericIndex], (next) => { target[numericIndex] = next; }, pkgPath, { values: target, index: numericIndex });",
    "}",
    "function __gojrTypedNil(typeName, pkgPath = gojrPackageArtifact.importPath) {",
    "  return { __gojrTypedNil: true, __gojrType: typeName, __gojrPkgPath: pkgPath || gojrPackageArtifact.importPath };",
    "}",
    "function __gojrToInterface(value, interfaceType, dynamicType) {",
    "  if (value && value.__gojrInterface === true) return { __gojrInterface: true, interfaceType, value: value.value };",
    "  if (value === null && typeof dynamicType === \"string\" && __gojrIsNilAssignableType(dynamicType)) value = __gojrTypedNil(dynamicType);",
    "  return { __gojrInterface: true, interfaceType, value };",
    "}",
    "function __gojrIsNilAssignableType(typeText) {",
    "  return typeText.startsWith(\"*\") || typeText.startsWith(\"[]\") || typeText.startsWith(\"map[\") || typeText.startsWith(\"chan \") || typeText.startsWith(\"<-chan\") || typeText.startsWith(\"chan<-\") || typeText.startsWith(\"func(\");",
    "}",
    "function __gojrRuntimeTypeName(value) {",
    "  if (value && value.__gojrInterface === true) return __gojrRuntimeTypeName(value.value);",
    "  if (value && value.__gojrPointer === true) return value.__gojrType;",
    "  if (value && value.__gojrTypedNil === true) return value.__gojrType;",
    "  if (value && typeof value === \"object\" && typeof value.__gojrType === \"string\") return value.__gojrType;",
    "  if (typeof value === \"bigint\") return \"int\";",
    "  if (typeof value === \"number\") return \"float64\";",
    "  if (__gojrIsString(value)) return \"string\";",
    "  if (typeof value === \"boolean\") return \"bool\";",
    "  return value === null || value === undefined ? \"nil\" : typeof value;",
    "}",
    "function __gojrRuntimeTypePackagePath(value) {",
    "  if (value && value.__gojrInterface === true) return __gojrRuntimeTypePackagePath(value.value);",
    "  if (value && value.__gojrPointer === true) return value.__gojrPkgPath || \"\";",
    "  if (value && value.__gojrTypedNil === true) return value.__gojrPkgPath || \"\";",
    "  if (value && typeof value === \"object\" && typeof value.__gojrPkgPath === \"string\") return value.__gojrPkgPath;",
    "  return \"\";",
    "}",
    "function __gojrReceiverBaseType(typeName) {",
    "  let name = String(typeName ?? \"\");",
    "  if (name.startsWith(\"*\")) name = name.slice(1).trim();",
    "  const dot = name.lastIndexOf(\".\");",
    "  let base = dot >= 0 ? name.slice(dot + 1) : name;",
    "  const open = base.indexOf(\"[\");",
    "  if (open > 0 && base.endsWith(\"]\") && !base.startsWith(\"map[\") && !base.startsWith(\"[\")) base = base.slice(0, open);",
    "  return base;",
    "}",
    "async function __gojrCallMethod(receiver, method, args, pkg) {",
    "  const actual = receiver && receiver.__gojrInterface === true ? receiver.value : receiver;",
    "  if (actual && actual.__gojrMethods && typeof actual.__gojrMethods[method] === \"function\") return await actual.__gojrMethods[method](actual, ...args);",
    "  if (actual && actual.__gojrPointer === true) {",
    "    const pointed = __gojrDeref(actual);",
    "    if (pointed && pointed.__gojrMethods && typeof pointed.__gojrMethods[method] === \"function\") return await pointed.__gojrMethods[method](pointed, ...args);",
    "  }",
    "  const fn = __gojrMethodFunction(actual, method, pkg);",
    "  if (typeof fn === \"function\") return await __gojrInvokeMethodFunction(fn, __gojrMethodReceiverForCall(actual, method), args);",
    "  const promoted = __gojrPromotedMethodReceiver(actual, method, pkg);",
    "  if (promoted !== undefined) return await __gojrCallMethod(promoted, method, args, pkg);",
    "  const field = actual == null ? undefined : __gojrGetField(actual, method);",
    "  if (typeof field === \"function\") return await field(...args);",
    "  throw new TypeError(`method ${method} not found on ${__gojrDescribeMethodTarget(actual)}`);",
    "}",
    "function __gojrDescribeMethodTarget(actual) {",
    "  if (actual === null) return \"<null>\";",
    "  if (actual === undefined) return \"<undefined>\";",
    "  const parts = [];",
    "  try { parts.push(`type=${__gojrRuntimeTypeName(actual)}`); } catch { }",
    "  try { parts.push(`pkg=${__gojrRuntimeTypePackagePath(actual)}`); } catch { }",
    "  if (actual && actual.__gojrPointer === true) {",
    "    parts.push(\"pointer=true\");",
    "    try { const pointed = __gojrDeref(actual); parts.push(`pointsTo=${__gojrRuntimeTypeName(pointed)}`); parts.push(`pointedKeys=${Object.keys(pointed || {}).join(\",\")}`); if (pointed && pointed.__gojrMethods) parts.push(`pointedMethods=${Object.keys(pointed.__gojrMethods).join(\",\")}`); } catch (error) { parts.push(`pointedError=${error instanceof Error ? error.message : String(error)}`); }",
    "  }",
    "  if (actual && actual.__gojrWeakPointer === true) parts.push(\"weak=true\");",
    "  if (actual && actual.__gojrMethods) parts.push(`methods=${Object.keys(actual.__gojrMethods).join(\",\")}`);",
    "  if (actual && typeof actual === \"object\") parts.push(`keys=${Object.keys(actual).join(\",\")}`);",
    "  return parts.join(\" \") || String(actual);",
    "}",
    "async function __gojrInvokeMethodFunction(fn, actual, args) {",
    "  if (fn.length >= args.length + 2) return await fn(__gojrPositionalTypeArgs(__gojrGenericArgsFromTypeName(__gojrRuntimeTypeName(actual))), actual, ...args);",
    "  return await fn(actual, ...args);",
    "}",
    "function __gojrMethodReceiverForCall(actual, method) {",
    "  if (!actual || actual.__gojrPointer === true) return actual;",
    "  const descriptor = __gojrDescriptorForTypeName(__gojrRuntimeTypeName(actual), __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(actual));",
    "  const methodDescriptor = descriptor && Array.isArray(descriptor.methods) ? descriptor.methods.find((item) => item && item.name === method) : undefined;",
    "  if (!methodDescriptor || methodDescriptor.pointerReceiver !== true) return actual;",
    "  const typeName = __gojrRuntimeTypeName(actual);",
    "  const pkgPath = __gojrRuntimeTypePackagePath(actual);",
    "  const identity = `addr:${__gojrObjectIdentityId(actual)}`;",
    "  return __gojrPointer(typeName, () => actual, (next) => { __gojrReplaceObjectValue(actual, next); }, pkgPath, undefined, undefined, identity);",
    "}",
    "function __gojrReplaceObjectValue(target, next) {",
    "  if (!target || typeof target !== \"object\" || !next || typeof next !== \"object\") return;",
    "  for (const key of Object.keys(target)) if (!key.startsWith(\"__gojr\")) delete target[key];",
    "  for (const key of Object.keys(next)) if (!key.startsWith(\"__gojr\")) target[key] = next[key];",
    "}",
    "function __gojrMethodFunction(actual, method, pkg) {",
    "  const typeName = actual && (actual.__gojrType || __gojrRuntimeTypeName(actual));",
    "  const methodKey = typeName ? `${__gojrReceiverBaseType(typeName)}.${method}` : undefined;",
    "  if (!methodKey) return undefined;",
    "  const receiverPkgPath = __gojrRuntimeTypePackagePath(actual);",
    "  if (receiverPkgPath && receiverPkgPath !== gojrPackageArtifact.importPath) {",
    "    const importedFn = __gojrSelectPackageField(__gojrActiveImportsByPath, receiverPkgPath, methodKey);",
    "    if (typeof importedFn === \"function\") return importedFn;",
    "  }",
    "  const fn = pkg[methodKey];",
    "  return typeof fn === \"function\" ? fn : undefined;",
    "}",
    "function __gojrPromotedMethodReceiver(actual, method, pkg, seen = new Set()) {",
    "  if (actual === null || actual === undefined) return undefined;",
    "  const typeName = actual && (actual.__gojrType || __gojrRuntimeTypeName(actual));",
    "  if (!typeName || seen.has(typeName)) return undefined;",
    "  seen.add(typeName);",
    "  let descriptor = __gojrDescriptorForTypeName(typeName, __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(actual));",
    "  if (descriptor && descriptor.kind === \"ptr\") descriptor = __gojrDescriptorForTypeName(descriptor.elem, __gojrActiveImportsByPath, descriptor.elemPkgPath || descriptor.pkgPath);",
    "  if (!descriptor || descriptor.kind !== \"struct\") return undefined;",
    "  const structValue = __gojrDerefIfPointer(actual);",
    "  for (const field of descriptor.fields || []) {",
    "    if (!field.embedded) continue;",
    "    const embedded = __gojrGetField(structValue, field.name);",
    "    if (embedded === null || embedded === undefined) continue;",
    "    if (embedded.__gojrMethods && typeof embedded.__gojrMethods[method] === \"function\") return embedded;",
    "    if (__gojrMethodFunction(embedded, method, pkg)) return embedded;",
    "    const nested = __gojrPromotedMethodReceiver(embedded, method, pkg, seen);",
    "    if (nested !== undefined) return nested;",
    "  }",
    "  return undefined;",
    "}",
    "let __gojrCallDepth = 0;",
    "let __gojrActiveRecoverPanic = undefined;",
    "let __gojrRecoverCallDepth = undefined;",
    "let __gojrRecoveredActivePanic = false;",
    "async function __gojrCallScope(body) {",
    "  __gojrCallDepth += 1;",
    "  try { return await body(); } finally { __gojrCallDepth -= 1; }",
    "}",
    "function __gojrRecover() {",
    "  if (__gojrActiveRecoverPanic && __gojrRecoverCallDepth === __gojrCallDepth && !__gojrRecoveredActivePanic) {",
    "    __gojrRecoveredActivePanic = true;",
    "    return __gojrActiveRecoverPanic.value;",
    "  }",
    "  return null;",
    "}",
    "async function __gojrDeferScope(body, recoveredReturn = () => null) {",
    "  const stack = [];",
    "  const push = (fn) => { stack.push(fn); };",
    "  let result;",
    "  let thrown;",
    "  try {",
    "    result = await body(push);",
    "  } catch (error) {",
    "    thrown = error;",
    "  }",
    "  let activePanic = thrown && thrown.__gojrPanic === true ? thrown : undefined;",
    "  while (stack.length > 0) {",
    "    const previousRecoverPanic = __gojrActiveRecoverPanic;",
    "    const previousRecoverCallDepth = __gojrRecoverCallDepth;",
    "    const previousRecoveredActivePanic = __gojrRecoveredActivePanic;",
    "    if (activePanic) {",
    "      __gojrActiveRecoverPanic = activePanic;",
    "      __gojrRecoverCallDepth = __gojrCallDepth + 1;",
    "      __gojrRecoveredActivePanic = false;",
    "    } else {",
    "      __gojrActiveRecoverPanic = undefined;",
    "      __gojrRecoverCallDepth = undefined;",
    "      __gojrRecoveredActivePanic = false;",
    "    }",
    "    try {",
    "      await stack.pop()();",
    "    } catch (error) {",
    "      if (error && error.__gojrPanic === true) activePanic = error;",
    "      else thrown = error;",
    "    } finally {",
    "      if (__gojrRecoveredActivePanic) activePanic = undefined;",
    "      __gojrActiveRecoverPanic = previousRecoverPanic;",
    "      __gojrRecoverCallDepth = previousRecoverCallDepth;",
    "      __gojrRecoveredActivePanic = previousRecoveredActivePanic;",
    "    }",
    "  }",
    "  if (activePanic) throw activePanic;",
    "  if (thrown !== undefined) {",
    "    if (thrown.__gojrPanic === true) return recoveredReturn();",
    "    throw thrown;",
    "  }",
    "  return result;",
    "}",
    "function __gojrGo(fn) {",
    "  Promise.resolve().then(fn).catch((error) => { setTimeout(() => { throw error; }, 0); });",
    "}",
    "let __gojrSelectSeed = 1;",
    "function __gojrSelectIndex(length) {",
    "  if (length <= 1) return 0;",
    "  __gojrSelectSeed = ((__gojrSelectSeed * 1664525 + 1013904223) >>> 0);",
    "  return __gojrSelectSeed % length;",
    "}",
    "function __gojrMakeChan(elementType, capacity = 0) {",
    "  return { __gojrChannel: true, elementType, capacity: Math.max(0, Number(capacity)), queue: [], receivers: [], senders: [], closed: false };",
    "}",
    "function __gojrClose(channel) {",
    "  if (!channel) throw new Error(\"panic: close of nil channel\");",
    "  if (channel.__gojrChannel !== true) throw new TypeError(\"close of non-channel\");",
    "  if (channel.closed) throw new Error(\"panic: close of closed channel\");",
    "  channel.closed = true;",
    "  while (channel.receivers.length > 0) {",
    "    const receiver = channel.receivers.shift();",
    "    receiver(__gojrTuple([__gojrZero(channel.elementType), false]));",
    "  }",
    "  while (channel.senders.length > 0) {",
    "    const sender = channel.senders.shift();",
    "    if (sender.reject) sender.reject(new Error(\"panic: send on closed channel\"));",
    "    else sender.resolve(null);",
    "  }",
    "  return null;",
    "}",
    "async function __gojrChanSend(channel, value) {",
    "  if (!channel || channel.__gojrChannel !== true) throw new TypeError(\"send on non-channel\");",
    "  if (channel.closed) throw new Error(\"panic: send on closed channel\");",
    "  if (channel.receivers.length > 0) {",
    "    const receiver = channel.receivers.shift();",
    "    receiver(__gojrTuple([value, true]));",
    "    return null;",
    "  }",
    "  if (channel.queue.length < channel.capacity) {",
    "    channel.queue.push(value);",
    "    return null;",
    "  }",
    "  return await new Promise((resolve, reject) => { channel.senders.push({ value, resolve, reject }); });",
    "}",
    "async function __gojrChanRecv(channel) {",
    "  if (!channel || channel.__gojrChannel !== true) throw new TypeError(\"receive from non-channel\");",
    "  if (channel.queue.length > 0) {",
    "    const value = channel.queue.shift();",
    "    __gojrDrainChannelSenders(channel);",
    "    return __gojrTuple([value, true]);",
    "  }",
    "  if (channel.senders.length > 0) {",
    "    const sender = channel.senders.shift();",
    "    sender.resolve(null);",
    "    return __gojrTuple([sender.value, true]);",
    "  }",
    "  if (channel.closed) return __gojrTuple([__gojrZero(channel.elementType), false]);",
    "  return await new Promise((resolve) => { channel.receivers.push(resolve); });",
    "}",
    "function __gojrDrainChannelSenders(channel) {",
    "  while (channel.senders.length > 0 && channel.queue.length < channel.capacity) {",
    "    const sender = channel.senders.shift();",
    "    if (channel.receivers.length > 0) {",
    "      const receiver = channel.receivers.shift();",
    "      receiver(__gojrTuple([sender.value, true]));",
    "    } else {",
    "      channel.queue.push(sender.value);",
    "    }",
    "    sender.resolve(null);",
    "  }",
    "}",
    "function __gojrChanCanRecv(channel) {",
    "  return Boolean(channel && channel.__gojrChannel === true && (channel.queue.length > 0 || channel.senders.length > 0 || channel.closed));",
    "}",
    "function __gojrChanCanSend(channel) {",
    "  return Boolean(channel && channel.__gojrChannel === true && (channel.closed || channel.receivers.length > 0 || channel.queue.length < channel.capacity));",
    "}",
    "async function __gojrSelect(cases) {",
    "  const defaults = cases.filter((item) => item.op === \"default\");",
    "  const ready = cases.filter((item) =>",
    "    (item.op === \"recv\" && __gojrChanCanRecv(item.channel)) ||",
    "    (item.op === \"send\" && __gojrChanCanSend(item.channel))",
    "  );",
    "  let selected = ready.length > 0 ? ready[__gojrSelectIndex(ready.length)] : undefined;",
    "  if (!selected && defaults.length > 0) selected = defaults[0];",
    "  if (!selected) selected = cases[0];",
    "  if (!selected) return null;",
    "  if (selected.op === \"recv\") return selected.apply(await __gojrChanRecv(selected.channel));",
    "  if (selected.op === \"send\") { await __gojrChanSend(selected.channel, selected.value); return selected.apply(__gojrTuple([null, true])); }",
    "  return selected.apply(__gojrTuple([null, true]));",
    "}",
    "function __gojrSpread(source) {",
    "  if (source == null) return [];",
    "  if (source && source.__gojrInterface === true) return __gojrSpread(source.value);",
    "  if (source && source.__gojrPointer === true) return __gojrSpread(source.__gojrGet());",
    "  if (source && source.__gojrTypedNil === true) return [];",
    "  if (Array.isArray(source)) return source;",
    "  if (ArrayBuffer.isView(source)) return Array.from(source);",
    "  if (__gojrIsString(source)) return Array.from(__gojrStringBytes(source));",
    "  if (typeof source[Symbol.iterator] === \"function\") return Array.from(source);",
    "  throw new TypeError(\"GOJR_RUNTIME001: spread argument is not a slice\");",
    "}",
    "async function __gojrRangeEntries(source) {",
    "  source = __gojrUnwrapInterface(source);",
    "  if (source && source.__gojrPointer === true) source = __gojrDeref(source);",
    "  if (source == null) return [];",
    "  if (source && source.__gojrTypedNil === true) return [];",
    "  if (typeof source === \"bigint\" || typeof source === \"number\") {",
    "    const count = Number(source);",
    "    const entries = [];",
    "    for (let i = 0; i < count; i += 1) entries.push([BigInt(i), BigInt(i)]);",
    "    return entries;",
    "  }",
    "  if (__gojrIsString(source)) {",
    "    if (source && source.__gojrRawString === true) source = new TextDecoder().decode(source.bytes);",
    "    const entries = [];",
    "    let byteIndex = 0;",
    "    for (const rune of source) {",
    "      entries.push([BigInt(byteIndex), BigInt(rune.codePointAt(0) ?? 0)]);",
    "      byteIndex += new TextEncoder().encode(rune).length;",
    "    }",
    "    return entries;",
    "  }",
    "  if (typeof source === \"function\") {",
    "    const entries = [];",
    "    let index = 0n;",
    "    const yieldFn = async (...args) => {",
    "      if (args.length === 0) { entries.push([index, index]); index += 1n; }",
    "      else if (args.length === 1) entries.push([args[0] ?? null, args[0] ?? null]);",
    "      else entries.push([args[0] ?? null, args[1] ?? null]);",
    "      return true;",
    "    };",
    "    await source(yieldFn);",
    "    return entries;",
    "  }",
    "  if (source instanceof Map && source.__gojrGoMap === true) return Array.from(Map.prototype.values.call(source), (entry) => [entry.key, entry.value]);",
    "  if (source instanceof Map) return Array.from(source.entries());",
    "  if (Array.isArray(source) || ArrayBuffer.isView(source)) return Array.from(source, (value, index) => [BigInt(index), value]);",
    "  if (typeof source === \"object\") return Object.entries(source);",
    "  return [];",
    "}",
    "function __gojrComplex(real, imag) {",
    "  return { real: Number(real), imag: Number(imag) };",
    "}",
    "function __gojrToComplex(value) {",
    "  return value && typeof value === \"object\" && \"real\" in value && \"imag\" in value ? value : __gojrComplex(Number(value), 0);",
    "}",
    "function __gojrComplexAdd(left, right) {",
    "  left = __gojrToComplex(left); right = __gojrToComplex(right);",
    "  return __gojrComplex(left.real + right.real, left.imag + right.imag);",
    "}",
    "function __gojrComplexSub(left, right) {",
    "  left = __gojrToComplex(left); right = __gojrToComplex(right);",
    "  return __gojrComplex(left.real - right.real, left.imag - right.imag);",
    "}",
    "function __gojrComplexMul(left, right) {",
    "  left = __gojrToComplex(left); right = __gojrToComplex(right);",
    "  return __gojrComplex(left.real * right.real - left.imag * right.imag, left.real * right.imag + left.imag * right.real);",
    "}",
    "function __gojrComplexDiv(left, right) {",
    "  left = __gojrToComplex(left); right = __gojrToComplex(right);",
    "  const denom = right.real * right.real + right.imag * right.imag;",
    "  return __gojrComplex((left.real * right.real + left.imag * right.imag) / denom, (left.imag * right.real - left.real * right.imag) / denom);",
    "}",
    "function __gojrComplexEq(left, right) {",
    "  left = __gojrToComplex(left); right = __gojrToComplex(right);",
    "  return left.real === right.real && left.imag === right.imag;",
    "}",
    "function __gojrComplexReal(value) {",
    "  return __gojrToComplex(value).real;",
    "}",
    "function __gojrComplexImag(value) {",
    "  return __gojrToComplex(value).imag;",
    "}",
    "function __gojrTypeAssert(value, typeText) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  if (__gojrValueMatchesType(actual, typeText)) return actual;",
    "  throw new TypeError(`interface conversion: value is not ${typeText}`);",
    "}",
    "function __gojrTypeAssertOk(value, typeText) {",
    "  const actual = value && value.__gojrInterface === true ? value.value : value;",
    "  return __gojrValueMatchesType(actual, typeText) ? __gojrTuple([actual, true]) : __gojrTuple([__gojrZero(typeText), false]);",
    "}",
    "function __gojrValueMatchesType(value, typeText) {",
    "  if (value && value.__gojrInterface === true) value = value.value;",
    "  const text = String(typeText || \"\").trim();",
    "  if (text === \"any\" || text === \"interface{}\") return true;",
    "  if (text === \"nil\") return value === null || value === undefined || Boolean(value && value.__gojrTypedNil === true);",
    "  const descriptor = __gojrDescriptorForTypeName(text, __gojrActiveImportsByPath);",
    "  if (descriptor && descriptor.kind === \"interface\") {",
    "    if (value === null || value === undefined) return false;",
    "    const required = __gojrReflectInterfaceMethods(descriptor, __gojrActiveImportsByPath);",
    "    if (required === undefined) return false;",
    "    if (required.length === 0) return true;",
    "    const actualDescriptor = __gojrDescriptorForTypeName(__gojrRuntimeTypeName(value), __gojrActiveImportsByPath, __gojrRuntimeTypePackagePath(value));",
    "    const available = __gojrReflectConcreteMethods(actualDescriptor, __gojrActiveImportsByPath, new Set(), Boolean(value && value.__gojrPointer === true));",
    "    return required.every((method) => __gojrReflectMethodSetHas(available, method));",
    "  }",
    "  if (text.startsWith(\"*\")) return Boolean((value && value.__gojrPointer === true && __gojrNormalizeTypeIdentity(value.__gojrType) === __gojrNormalizeTypeIdentity(text)) || (value && value.__gojrTypedNil === true && __gojrNormalizeTypeIdentity(value.__gojrType) === __gojrNormalizeTypeIdentity(text)));",
    "  switch (text) {",
    "    case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return typeof value === \"bigint\";",
    "    case \"float32\": case \"float64\": return typeof value === \"number\";",
    "    case \"string\": return __gojrIsString(value);",
    "    case \"bool\": return typeof value === \"boolean\";",
    "    case \"complex64\": case \"complex128\": return Boolean(value && typeof value === \"object\" && \"real\" in value && \"imag\" in value);",
    "    default: return __gojrReceiverBaseType(__gojrRuntimeTypeName(value)) === __gojrReceiverBaseType(text);",
    "  }",
    "}",
    "function __gojrTypeArg(typeArgs, name) {",
    "  return typeArgs && typeof typeArgs[name] === \"string\" ? typeArgs[name] : \"any\";",
    "}",
    "function __gojrTypeArgsFromReceiver(receiver, names) {",
    "  const args = __gojrGenericArgsFromTypeName(__gojrRuntimeTypeName(receiver));",
    "  const out = {};",
    "  for (let index = 0; index < names.length; index += 1) out[names[index]] = args[index] || \"any\";",
    "  return out;",
    "}",
    "function __gojrImportedTypeArgs(importedPackage, functionName, typeArgs) {",
    "  const typeText = importedPackage && importedPackage.__gojrExportIndex && importedPackage.__gojrExportIndex[functionName] && importedPackage.__gojrExportIndex[functionName].typeText;",
    "  const names = __gojrFunctionTypeParameterNames(typeText);",
    "  const out = __gojrPositionalTypeArgs(typeArgs);",
    "  for (let index = 0; index < names.length; index += 1) out[names[index]] = typeArgs[index] || \"any\";",
    "  return out;",
    "}",
    "async function __gojrCallImportedFunction(importPath, functionName, args, argTypes) {",
    "  const importedPackage = __gojrImportedPackage(importPath);",
    "  const fn = __gojrSelectPackageCallableField(__gojrActiveImportsByPath, importPath, functionName);",
    "  const typeText = importedPackage && importedPackage.__gojrExportIndex && importedPackage.__gojrExportIndex[functionName] && importedPackage.__gojrExportIndex[functionName].typeText;",
    "  const parameterTypes = __gojrFunctionParameterTypes(typeText);",
    "  const convertedArgs = args.map((arg, index) => __gojrConvertRuntimeCallArg(arg, __gojrRuntimeCallArgumentTargetType(parameterTypes, index, false), importPath));",
    "  const typeArgs = __gojrInferImportedTypeArgs(importedPackage, functionName, argTypes);",
    "  return typeArgs ? await fn(typeArgs, ...convertedArgs) : await fn(...convertedArgs);",
    "}",
    "function __gojrRuntimeCallArgumentTargetType(targetTypes, index, isSpread) {",
    "  if (!Array.isArray(targetTypes) || targetTypes.length === 0) return undefined;",
    "  const direct = targetTypes[index];",
    "  if (direct !== undefined) return String(direct).startsWith(\"...\") ? (isSpread ? `[]${String(direct).slice(3)}` : String(direct).slice(3)) : direct;",
    "  const variadicIndex = targetTypes.findIndex((type) => String(type).startsWith(\"...\"));",
    "  return variadicIndex >= 0 && index >= variadicIndex ? String(targetTypes[variadicIndex]).slice(3) : undefined;",
    "}",
    "function __gojrConvertRuntimeCallArg(value, targetType, declaringImportPath = undefined) {",
    "  if (!targetType) return value;",
    "  const targetPkgPath = __gojrPackagePathForTypeName(targetType, declaringImportPath);",
    "  if (__gojrRuntimeInterfaceTypeName(targetType, targetPkgPath)) return __gojrToInterface(value, targetType, __gojrRuntimeTypeName(value));",
    "  return __gojrConvertDynamicType(targetType, value, __gojrActiveImportsByPath, targetPkgPath);",
    "}",
    "function __gojrRuntimeInterfaceTypeName(typeName, pkgPath = undefined) {",
    "  const text = String(typeName || \"\").trim();",
    "  if (text === \"any\" || text === \"interface{}\" || text === \"error\" || text.startsWith(\"interface{\")) return true;",
    "  return __gojrDescriptorForTypeName(text, __gojrActiveImportsByPath, pkgPath)?.kind === \"interface\";",
    "}",
    "function __gojrPackagePathForTypeName(typeName, declaringImportPath = undefined) {",
    "  const qualifier = __gojrLeadingTypeQualifier(typeName);",
    "  if (!qualifier) return declaringImportPath;",
    "  const declaringPackage = declaringImportPath ? __gojrImportedPackage(declaringImportPath) : undefined;",
    "  const importMap = declaringPackage && declaringPackage.__gojrImportPathsByQualifier;",
    "  if (importMap && typeof importMap[qualifier] === \"string\") return importMap[qualifier];",
    "  if (declaringImportPath === gojrPackageArtifact.importPath && typeof __gojrImportPathsByQualifier[qualifier] === \"string\") return __gojrImportPathsByQualifier[qualifier];",
    "  const builtinPath = __gojrBuiltinImportPathForQualifier(qualifier);",
    "  if (builtinPath) return builtinPath;",
    "  return declaringImportPath;",
    "}",
    "function __gojrLeadingTypeQualifier(typeName) {",
    "  let text = String(typeName || \"\").trim();",
    "  while (text.startsWith(\"...\")) text = text.slice(3).trim();",
    "  while (text.startsWith(\"*\")) text = text.slice(1).trim();",
    "  while (text.startsWith(\"[]\")) text = text.slice(2).trim();",
    "  while (/^\\[[^\\]]+\\]/.test(text) && !text.startsWith(\"[]\")) text = text.replace(/^\\[[^\\]]+\\]/, \"\").trim();",
    "  const match = /^([_A-Za-z$][_A-Za-z0-9$]*)\\./.exec(text);",
    "  return match ? match[1] : undefined;",
    "}",
    "function __gojrBuiltinImportPathForQualifier(qualifier) {",
    "  switch (qualifier) {",
    "    case \"unsafe\": return \"unsafe\";",
    "    case \"reflect\": return \"reflect\";",
    "    case \"runtime\": return \"runtime\";",
    "    case \"sync\": return \"sync\";",
    "    case \"atomic\": return \"sync/atomic\";",
    "    case \"weak\": return \"weak\";",
    "    default: return undefined;",
    "  }",
    "}",
    "function __gojrImportedPackage(importPath) {",
    "  const builtin = __gojrBuiltinImport(importPath, __gojrActiveImportsByPath);",
    "  const imported = __gojrActiveImportsByPath && __gojrActiveImportsByPath[importPath];",
    "  const packageObject = imported && typeof imported === \"object\" && \"package\" in imported ? imported.package : imported;",
    "  if (__gojrBuiltinImportOverrides(importPath) && builtin) return __gojrMergedBuiltinPackage(builtin, packageObject);",
    "  return packageObject || builtin || {};",
    "}",
    "function __gojrInferImportedTypeArgs(importedPackage, functionName, argTypes) {",
    "  const typeText = importedPackage && importedPackage.__gojrExportIndex && importedPackage.__gojrExportIndex[functionName] && importedPackage.__gojrExportIndex[functionName].typeText;",
    "  const names = __gojrFunctionTypeParameterNames(typeText);",
    "  if (names.length === 0) return undefined;",
    "  const parameterTypes = __gojrFunctionParameterTypes(typeText);",
    "  const parameters = new Set(names);",
    "  const inferred = Object.create(null);",
    "  for (let index = 0; index < parameterTypes.length; index += 1) __gojrInferTypeArgument(parameterTypes[index], argTypes[index], parameters, inferred);",
    "  const out = __gojrPositionalTypeArgs([]);",
    "  for (const name of names) out[name] = inferred[name] || \"any\";",
    "  return out;",
    "}",
    "function __gojrFunctionParameterTypes(typeText) {",
    "  const text = String(typeText || \"\").trim();",
    "  if (!text.startsWith(\"func\")) return [];",
    "  const typeParameterNames = new Set(__gojrFunctionTypeParameterNames(text));",
    "  let index = 4;",
    "  if (text[index] === \"[\") {",
    "    const close = __gojrMatchingBracket(text, index);",
    "    if (close < 0) return [];",
    "    index = close + 1;",
    "  }",
    "  while (/\\s/.test(text[index] || \"\")) index += 1;",
    "  if (text[index] !== \"(\") return [];",
    "  const close = __gojrMatchingParen(text, index);",
    "  if (close < 0) return [];",
    "  return __gojrParameterListTypes(text.slice(index + 1, close), typeParameterNames);",
    "}",
    "function __gojrParameterListTypes(text, typeParameterNames) {",
    "  const parts = __gojrSplitTopLevelTypes(text);",
    "  const out = [];",
    "  let pendingNames = 0;",
    "  for (const raw of parts) {",
    "    const part = raw.trim();",
    "    if (!part) continue;",
    "    const named = __gojrParameterNamedPart(part);",
    "    if (named) {",
    "      const count = Math.max(1, pendingNames + __gojrParameterNameCount(named.names));",
    "      for (let index = 0; index < count; index += 1) out.push(named.type);",
    "      pendingNames = 0;",
    "      continue;",
    "    }",
    "    if (pendingNames > 0) { out.push(part); pendingNames = 0; }",
    "    else if (typeParameterNames.has(part)) out.push(part);",
    "    else if (__gojrLooksLikeParameterName(part) && parts.length > 1) pendingNames += 1;",
    "    else out.push(part);",
    "  }",
    "  return out;",
    "}",
    "function __gojrParameterNamedPart(part) {",
    "  const boundary = __gojrFirstTopLevelWhitespace(part);",
    "  if (boundary < 0) return undefined;",
    "  const head = part.slice(0, boundary).trim();",
    "  const tail = part.slice(boundary).trim();",
    "  if (!head || !tail) return undefined;",
    "  if (!__gojrLooksLikeParameterName(head)) return undefined;",
    "  if (!__gojrLooksLikeTypeText(tail)) return undefined;",
    "  return { names: head, type: tail };",
    "}",
    "function __gojrFirstTopLevelWhitespace(text) {",
    "  let paren = 0, bracket = 0, brace = 0;",
    "  for (let index = 0; index < text.length; index += 1) {",
    "    const char = text[index];",
    "    if (char === \"(\") paren += 1;",
    "    else if (char === \")\") paren = Math.max(0, paren - 1);",
    "    else if (char === \"[\") bracket += 1;",
    "    else if (char === \"]\") bracket = Math.max(0, bracket - 1);",
    "    else if (char === \"{\") brace += 1;",
    "    else if (char === \"}\") brace = Math.max(0, brace - 1);",
    "    else if (paren === 0 && bracket === 0 && brace === 0 && /\\s/.test(char)) return index;",
    "  }",
    "  return -1;",
    "}",
    "function __gojrLooksLikeParameterName(text) {",
    "  const name = String(text || \"\").trim();",
    "  if (!/^[_A-Za-z$][_A-Za-z0-9$]*$/.test(name)) return false;",
    "  return !new Set([\"any\", \"bool\", \"byte\", \"chan\", \"complex64\", \"complex128\", \"error\", \"float32\", \"float64\", \"func\", \"int\", \"int8\", \"int16\", \"int32\", \"int64\", \"interface\", \"map\", \"rune\", \"string\", \"struct\", \"uint\", \"uint8\", \"uint16\", \"uint32\", \"uint64\", \"uintptr\"]).has(name);",
    "}",
    "function __gojrLooksLikeTypeText(text) {",
    "  const type = String(text || \"\").trim();",
    "  if (!type) return false;",
    "  if (type.startsWith(\"...\") || type.startsWith(\"*\") || type.startsWith(\"[]\") || type.startsWith(\"[\") || type.startsWith(\"<-\") || type.startsWith(\"chan\") || type.startsWith(\"func\") || type.startsWith(\"interface{\") || type.startsWith(\"struct{\") || type.startsWith(\"map[\")) return true;",
    "  return /^[_A-Za-z$][_A-Za-z0-9$]*(\\.[_A-Za-z$][_A-Za-z0-9$]*)?(\\[.*\\])?$/.test(type);",
    "}",
    "function __gojrParameterNameCount(text) {",
    "  return String(text || \"\").split(\",\").map((part) => part.trim()).filter(Boolean).length;",
    "}",
    "function __gojrMatchingParen(text, open) {",
    "  let depth = 0;",
    "  for (let index = open; index < text.length; index += 1) {",
    "    const char = text[index];",
    "    if (char === \"(\") depth += 1;",
    "    if (char === \")\") {",
    "      depth -= 1;",
    "      if (depth === 0) return index;",
    "    }",
    "  }",
    "  return -1;",
    "}",
    "function __gojrInferTypeArgument(pattern, actual, parameters, out) {",
    "  let lhs = String(pattern || \"\").trim();",
    "  const rhs = String(actual || \"\").trim();",
    "  if (lhs.startsWith(\"...\")) lhs = lhs.slice(3).trim();",
    "  if (!lhs || !rhs) return;",
    "  if (parameters.has(lhs)) { out[lhs] = rhs; return; }",
    "  if (lhs.startsWith(\"*\") && rhs.startsWith(\"*\")) { __gojrInferTypeArgument(lhs.slice(1), rhs.slice(1), parameters, out); return; }",
    "  if (lhs.startsWith(\"[]\") && rhs.startsWith(\"[]\")) { __gojrInferTypeArgument(lhs.slice(2), rhs.slice(2), parameters, out); return; }",
    "  const lhsArgs = __gojrGenericArgsFromTypeName(lhs);",
    "  const rhsArgs = __gojrGenericArgsFromTypeName(rhs);",
    "  if (lhsArgs.length > 0 && rhsArgs.length > 0 && __gojrGenericBaseTypeName(lhs) === __gojrGenericBaseTypeName(rhs)) {",
    "    for (let index = 0; index < lhsArgs.length; index += 1) __gojrInferTypeArgument(lhsArgs[index], rhsArgs[index], parameters, out);",
    "  }",
    "}",
    "function __gojrGenericBaseTypeName(typeName) {",
    "  const text = String(typeName || \"\").trim();",
    "  const open = text.indexOf(\"[\");",
    "  return open < 0 ? text : text.slice(0, open).trim();",
    "}",
    "function __gojrFunctionTypeParameterNames(typeText) {",
    "  const text = String(typeText || \"\");",
    "  if (!text.startsWith(\"func[\")) return [];",
    "  const close = __gojrMatchingBracket(text, 4);",
    "  if (close < 0) return [];",
    "  const names = [];",
    "  for (const part of __gojrSplitTopLevelTypes(text.slice(5, close))) {",
    "    const trimmed = part.trim();",
    "    if (!trimmed) continue;",
    "    const name = /^[_\\p{L}][_$\\p{L}\\p{N}]*/u.exec(trimmed)?.[0];",
    "    if (name) names.push(name);",
    "  }",
    "  return names;",
    "}",
    "function __gojrMatchingBracket(text, open) {",
    "  let depth = 0;",
    "  for (let index = open; index < text.length; index += 1) {",
    "    const char = text[index];",
    "    if (char === \"[\") depth += 1;",
    "    if (char === \"]\") {",
    "      depth -= 1;",
    "      if (depth === 0) return index;",
    "    }",
    "  }",
    "  return -1;",
    "}",
    "function __gojrPositionalTypeArgs(typeArgs) {",
    "  const out = {};",
    "  const aliases = [[\"T\", \"K\", \"E\", \"A\", \"T0\"], [\"U\", \"V\", \"B\", \"T1\"], [\"W\", \"C\", \"T2\"], [\"X\", \"D\", \"T3\"]];",
    "  for (let index = 0; index < typeArgs.length; index += 1) {",
    "    out[`T${index}`] = typeArgs[index] || \"any\";",
    "    for (const name of aliases[index] || []) out[name] = typeArgs[index] || \"any\";",
    "  }",
    "  return out;",
    "}",
    "function __gojrGenericArgsFromTypeName(typeName) {",
    "  const text = String(typeName || \"\").trim();",
    "  const open = text.indexOf(\"[\");",
    "  if (open < 0 || !text.endsWith(\"]\")) return [];",
    "  return __gojrSplitTopLevelTypes(text.slice(open + 1, -1));",
    "}",
    "function __gojrSplitTopLevelTypes(text) {",
    "  const parts = []; let depth = 0; let start = 0;",
    "  for (let index = 0; index < text.length; index += 1) {",
    "    const char = text[index];",
    "    if (char === \"[\" || char === \"(\" || char === \"{\") depth += 1;",
    "    else if (char === \"]\" || char === \")\" || char === \"}\") depth -= 1;",
    "    else if (char === \",\" && depth === 0) { parts.push(text.slice(start, index).trim()); start = index + 1; }",
    "  }",
    "  const last = text.slice(start).trim();",
    "  if (last) parts.push(last);",
    "  return parts;",
    "}",
    "function __gojrGenericTypeSubstitutions(typeText, descriptor) {",
    "  const params = descriptor && Array.isArray(descriptor.typeParameters) ? descriptor.typeParameters : [];",
    "  if (params.length === 0) return {};",
    "  const args = __gojrGenericArgsFromTypeName(typeText);",
    "  const out = {};",
    "  for (let index = 0; index < params.length; index += 1) out[params[index]] = args[index] || \"any\";",
    "  return out;",
    "}",
    "function __gojrSubstituteType(typeText, substitutions) {",
    "  const text = String(typeText || \"\").trim();",
    "  if (!text) return text;",
    "  if (Object.prototype.hasOwnProperty.call(substitutions, text)) return substitutions[text];",
    "  if (text.startsWith(\"*\")) return `*${__gojrSubstituteType(text.slice(1), substitutions)}`;",
    "  if (text.startsWith(\"[]\")) return `[]${__gojrSubstituteType(text.slice(2), substitutions)}`;",
    "  if (text.startsWith(\"map[\")) {",
    "    let depth = 0;",
    "    for (let index = 4; index < text.length; index += 1) {",
    "      const char = text[index];",
    "      if (char === \"[\") depth += 1;",
    "      if (char === \"]\") {",
    "        if (depth === 0) return `map[${__gojrSubstituteType(text.slice(4, index), substitutions)}]${__gojrSubstituteType(text.slice(index + 1), substitutions)}`;",
    "        depth -= 1;",
    "      }",
    "    }",
    "  }",
    "  const open = text.indexOf(\"[\");",
    "  if (open > 0 && text.endsWith(\"]\") && !text.startsWith(\"[\")) {",
    "    const args = __gojrSplitTopLevelTypes(text.slice(open + 1, -1)).map((arg) => __gojrSubstituteType(arg, substitutions));",
    "    return `${text.slice(0, open)}[${args.join(\", \")}]`;",
    "  }",
    "  return text;",
    "}",
    "function __gojrZero(typeText, pkgPath = undefined, importsByPath = __gojrActiveImportsByPath) {",
    "  const intrinsicZero = __gojrIntrinsicZero(typeText, pkgPath);",
    "  if (intrinsicZero !== undefined) return intrinsicZero;",
    "  switch (typeText) {",
    "    case \"byte\": case \"rune\": case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return 0n;",
    "    case \"float32\": case \"float64\": return 0;",
    "    case \"string\": return \"\";",
    "    case \"bool\": return false;",
    "    case \"complex64\": case \"complex128\": return __gojrComplex(0, 0);",
    "    default: {",
    "      const descriptor = __gojrDescriptorForTypeName(typeText, importsByPath, pkgPath);",
    "      const descriptorPkgPath = descriptor && descriptor.pkgPath ? descriptor.pkgPath : pkgPath;",
    "      const kind = descriptor && descriptor.kind;",
    "      switch (kind) {",
    "        case \"byte\": case \"rune\": case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return 0n;",
    "        case \"float32\": case \"float64\": return 0;",
    "        case \"string\": return \"\";",
    "        case \"bool\": return false;",
    "        case \"complex64\": case \"complex128\": return __gojrComplex(0, 0);",
    "      }",
    "      if (__gojrIsNilAssignableType(typeText) || kind === \"ptr\" || kind === \"slice\" || kind === \"map\" || kind === \"chan\" || kind === \"func\") return __gojrTypedNil(typeText, descriptorPkgPath);",
    "      if (kind === \"array\" && typeof descriptor.len === \"number\") {",
    "        if (descriptor.elem === \"byte\" || descriptor.elem === \"uint8\") return __gojrSetCap(new Uint8Array(descriptor.len), descriptor.len);",
    "        return Array.from({ length: descriptor.len }, () => __gojrZero(descriptor.elem, descriptor.elemPkgPath, importsByPath));",
    "      }",
    "      if (kind === \"struct\") {",
    "        const substitutions = __gojrGenericTypeSubstitutions(typeText, descriptor);",
    "        const fields = {};",
    "        for (const field of descriptor.fields || []) fields[field.name] = __gojrZero(__gojrSubstituteType(field.type, substitutions), field.pkgPath, importsByPath);",
    "        return __gojrStruct(typeText, fields, descriptorPkgPath);",
    "      }",
    "      if (descriptor && descriptor.underlying && descriptor.underlying !== typeText) return __gojrZero(descriptor.underlying, descriptorPkgPath, importsByPath);",
    "      return null;",
    "    }",
    "  }",
    "}",
    "function __gojrDecodeBase64(base64) {",
    "  if (typeof Buffer !== \"undefined\") return new Uint8Array(Buffer.from(base64, \"base64\"));",
    "  const binary = atob(base64);",
    "  const out = new Uint8Array(binary.length);",
    "  for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);",
    "  return out;",
    "}",
    "export default { artifact: gojrPackageArtifact, instantiateGoJrPackage };",
    ""
  ].filter((line) => line !== "").join("\n");
  return stripStage1IntrinsicSupport(source, generatedBuiltinImports);
}

const STAGE1_INTRINSIC_SUPPORT_START = "function __gojrBuiltinImport(path, importsByPath) {";
const STAGE1_INTRINSIC_SUPPORT_END = "const __gojrInternalAbiTypeCache = new Map();";
let cachedStage1IntrinsicSupportSource: string | undefined;

export function stage1IntrinsicSupportSource(): string {
  if (cachedStage1IntrinsicSupportSource) return cachedStage1IntrinsicSupportSource;
  stage1JavaScript(stage1IntrinsicSupportDummyArtifact(), false, [], [], emptyPackageEmitFacts);
  if (cachedStage1IntrinsicSupportSource) return cachedStage1IntrinsicSupportSource;
  throw new Error("could not extract GoJr Stage 1 intrinsic support source");
}

function stripStage1IntrinsicSupport(source: string, generatedBuiltinImports: string): string {
  const start = source.indexOf(STAGE1_INTRINSIC_SUPPORT_START);
  const end = source.indexOf(STAGE1_INTRINSIC_SUPPORT_END);
  if (start < 0 || end < 0 || end <= start) return source;
  cachedStage1IntrinsicSupportSource = source.slice(start, end).trimEnd();
  return [
    source.slice(0, start).trimEnd(),
    stage1IntrinsicSupportHookSource(generatedBuiltinImports),
    source.slice(end)
  ].join("\n");
}

function stage1IntrinsicSupportHookSource(generatedBuiltinImports: string): string {
  return [
    `const __gojrGeneratedBuiltinImportPaths = new Set(${generatedBuiltinImports});`,
    "function __gojrBuiltinImportOverrides(path) {",
    "  return __gojrGeneratedBuiltinImportPaths.has(path);",
    "}",
    "function __gojrBuiltinImport(path, importsByPath) {",
    "  const resolver = __gojrActiveRuntimeOptions.builtinImport || __gojrActiveRuntimeOptions.__gojrBuiltinImport;",
    "  if (typeof resolver !== \"function\") return undefined;",
    "  return resolver(path, importsByPath, __gojrIntrinsicOps());",
    "}",
    "function __gojrIntrinsicZero(typeText, pkgPath = undefined) {",
    "  const resolver = __gojrActiveRuntimeOptions.intrinsicZero || __gojrActiveRuntimeOptions.__gojrIntrinsicZero;",
    "  return typeof resolver === \"function\" ? resolver(typeText, pkgPath, __gojrIntrinsicOps()) : undefined;",
    "}",
    "function __gojrIntrinsicDescriptor(typeText, pkgPath = undefined) {",
    "  const resolver = __gojrActiveRuntimeOptions.intrinsicDescriptor || __gojrActiveRuntimeOptions.__gojrIntrinsicDescriptor;",
    "  return typeof resolver === \"function\" ? resolver(typeText, pkgPath, __gojrIntrinsicOps()) : undefined;",
    "}",
    "function __gojrIntrinsicStruct(typeName, value, pkgPath = undefined) {",
    "  const resolver = __gojrActiveRuntimeOptions.intrinsicStruct || __gojrActiveRuntimeOptions.__gojrIntrinsicStruct;",
    "  return typeof resolver === \"function\" ? resolver(typeName, value, pkgPath, __gojrIntrinsicOps()) : undefined;",
    "}",
    "function __gojrByteSequenceCompare(left, right) {",
    "  const length = Math.min(left.length, right.length);",
    "  for (let index = 0; index < length; index += 1) {",
    "    if (left[index] !== right[index]) return left[index] < right[index] ? -1 : 1;",
    "  }",
    "  if (left.length === right.length) return 0;",
    "  return left.length < right.length ? -1 : 1;",
    "}",
    "function __gojrDefaultTextSink(stream) {",
    "  if (typeof process !== \"undefined\") return undefined;",
    "  const target = stream === \"stderr\" ? globalThis.console?.error || globalThis.console?.log : globalThis.console?.log;",
    "  if (typeof target !== \"function\") return undefined;",
    "  return (text) => target.call(globalThis.console, text);",
    "}",
    "let __gojrCachedIntrinsicOps;",
    "function __gojrIntrinsicOps() {",
    "  if (__gojrCachedIntrinsicOps) return __gojrCachedIntrinsicOps;",
    "  const ops = {",
    "    gojrPackageArtifact,",
    "    get __gojrActiveRuntimeOptions() { return __gojrActiveRuntimeOptions; },",
    "    get __gojrActiveImportsByPath() { return __gojrActiveImportsByPath; },",
    "    get __gojrActivePackage() { return __gojrActivePackage; },",
    "    globalThis,",
    "    Array, ArrayBuffer, BigInt, Boolean, DataView, Error, JSON, Map, Math, Number, Object, Promise, RangeError, Reflect, RegExp, Set, String, Symbol, TextDecoder, TextEncoder, TypeError, Uint8Array, WeakMap, console,",
    "    setTimeout, clearTimeout,",
    "    __gojrSetCap, __gojrBytesFrom, __gojrBytesFromSequence, __gojrStringBytes, __gojrLen,",
    "    __gojrTuple, __gojrTupleValues, __gojrStruct, __gojrPointer, __gojrPointerValue, __gojrTypedNil,",
    "    __gojrDeref, __gojrDerefIfPointer, __gojrPointerSequence, __gojrSliceFromSequence, __gojrSliceElementRuntimeType,",
    "    __gojrUnsafeAdd, __gojrNilUnsafePointer, __gojrAlignofValue, __gojrSizeofValue,",
    "    __gojrEqual, __gojrTypeArg, __gojrRuntimeTypeName, __gojrRuntimeTypePackagePath,",
    "    __gojrReflectTypeOf, __gojrReflectValueOf, __gojrReflectliteSwapper, __gojrReflectTypeForTypeName,",
    "    __gojrReflectPointerTo, __gojrTypeAssertOk, __gojrDescriptorForTypeName, __gojrZero",
    "  };",
    "  __gojrCachedIntrinsicOps = ops;",
    "  return ops;",
    "}"
  ].join("\n");
}

function stage1IntrinsicSupportDummyArtifact(): GoJuniorPackageExportData {
  return {
    exportFormat: "gojr-iexport",
    exportVersion: 1,
    exportIndex: {},
    layoutVersion: "gojr-js-v4",
    compilerVersion: "gojr-dev",
    backend: GOJR_STAGE1_BACKEND,
    hostSpecVersion: "host-v0",
    capabilityPolicy: "default",
    goos: "js",
    goarch: "gojr",
    buildTags: [],
    standardLibrary: false,
    importPath: "gojr/intrinsic-support-probe",
    packageName: "probe",
    sourceHash: "",
    cacheKey: "",
    dependencies: [],
    dependencyCacheKeys: [],
    exports: [],
    sources: []
  };
}
