import {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  DeferStatement,
  Expression,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  IdentifierExpression,
  IfStatement,
  IncDecStatement,
  IndexExpression,
  MapLiteralExpression,
  ProgramAst,
  SelectorExpression,
  SpreadsheetRangeExpression,
  Statement,
  SwitchStatement,
  TypeNode,
  UnaryExpression
} from "./ast.js";
import { cstToAst } from "./cstToAst.js";
import { Diagnostic } from "./diagnostics.js";
import { parseGoJunior } from "./parser.js";

export type RuntimeValue =
  | null
  | boolean
  | string
  | number
  | bigint
  | RuntimeValue[]
  | RuntimeMap
  | RuntimeObject
  | RuntimeCallable
  | GoJuniorFunction
  | SheetBinding;

export interface RuntimeObject {
  [name: string]: RuntimeValue;
}

export interface RuntimeCallable {
  kind: "HostCallable";
  name: string;
  call(args: RuntimeValue[], context: EvaluationContext): RuntimeValue;
}

export interface GoJuniorFunction {
  kind: "GoJuniorFunction";
  name: string;
  declaration?: FunctionDecl;
  call(args: RuntimeValue[], context: EvaluationContext): RuntimeValue;
}

export interface SheetData {
  [cell: string]: RuntimeValue;
}

export interface EvaluationOptions {
  packages?: Record<string, RuntimeObject>;
  sheet?: SheetData;
  sheets?: Record<string, SheetData>;
  currentSheetName?: string;
  stdout?: (text: string) => void;
  maxLoopIterations?: number;
}

export interface EvaluationResult {
  diagnostics: Diagnostic[];
  output: string[];
  ast?: ProgramAst;
  value?: RuntimeValue;
  values?: RuntimeValue[];
  incomplete?: boolean;
}

type Completion =
  | { kind: "normal"; value?: RuntimeValue }
  | { kind: "return"; values: RuntimeValue[] }
  | { kind: "break"; label?: string }
  | { kind: "continue"; label?: string }
  | { kind: "fallthrough" }
  | { kind: "goto"; label: string };

interface Binding {
  value: RuntimeValue;
  mutable: boolean;
}

export class GoJuniorRuntimeError extends Error {
  public constructor(message: string) {
    super(message);
    this.name = "GoJuniorRuntimeError";
  }
}

export class GoJuniorPanic extends Error {
  public readonly value: RuntimeValue;

  public constructor(value: RuntimeValue) {
    super(`panic: ${formatValue(value)}`);
    this.name = "GoJuniorPanic";
    this.value = value;
  }
}

export class EvaluationContext {
  public readonly output: string[] = [];
  private readonly rootScope = new Scope();
  private currentScope = this.rootScope;
  private readonly deferFrames: Array<Array<() => RuntimeValue>> = [[]];
  private readonly maxLoopIterations: number;
  private readonly stdout: ((text: string) => void) | undefined;

  public constructor(private readonly options: EvaluationOptions = {}) {
    this.maxLoopIterations = options.maxLoopIterations ?? 100_000;
    this.stdout = options.stdout;
    installBuiltins(this);
    installAutomaticImports(this);
  }

  public declare(name: string, value: RuntimeValue, mutable = true): void {
    this.currentScope.declare(name, value, mutable);
  }

  public declareRoot(name: string, value: RuntimeValue, mutable = true): void {
    this.rootScope.declare(name, value, mutable);
  }

  public declareOrAssignRoot(name: string, value: RuntimeValue, mutable = true): void {
    if (this.rootScope.hasLocal(name)) {
      this.rootScope.assign(name, value);
      return;
    }
    this.rootScope.declare(name, value, mutable);
  }

  public assign(name: string, value: RuntimeValue): void {
    this.currentScope.assign(name, value);
  }

  public lookup(name: string): RuntimeValue {
    return this.currentScope.lookup(name);
  }

  public outputFrom(offset: number): string[] {
    return this.output.slice(offset);
  }

  public childScope<T>(body: () => T): T {
    const previous = this.currentScope;
    this.currentScope = new Scope(previous);
    try {
      return body();
    } finally {
      this.currentScope = previous;
    }
  }

  public captureScope(): Scope {
    return this.currentScope;
  }

  public withScope<T>(scope: Scope, body: () => T): T {
    const previous = this.currentScope;
    this.currentScope = scope;
    try {
      return body();
    } finally {
      this.currentScope = previous;
    }
  }

  public pushDefer(callback: () => RuntimeValue): void {
    this.currentDeferFrame().push(callback);
  }

  public runDefers(): void {
    const frame = this.currentDeferFrame();
    while (frame.length > 0) {
      frame.pop()?.();
    }
  }

  public deferScope<T>(body: () => T): T {
    this.deferFrames.push([]);
    try {
      return body();
    } finally {
      try {
        this.runDefers();
      } finally {
        this.deferFrames.pop();
      }
    }
  }

  public write(text: string): void {
    this.output.push(text);
    this.stdout?.(text);
  }

  public loopLimit(): number {
    return this.maxLoopIterations;
  }

  public packages(): Record<string, RuntimeObject> {
    return this.options.packages ?? {};
  }

  public currentSheetName(): string {
    return this.options.currentSheetName ?? "sheet";
  }

  public sheetData(name: string): SheetData {
    if (name === "sheet") return this.options.sheet ?? this.options.sheets?.[this.currentSheetName()] ?? {};
    return this.options.sheets?.[name] ?? {};
  }

  private currentDeferFrame(): Array<() => RuntimeValue> {
    const frame = this.deferFrames[this.deferFrames.length - 1];
    if (!frame) throw new GoJuniorRuntimeError("internal error: missing defer frame");
    return frame;
  }
}

class Scope {
  private readonly bindings = new Map<string, Binding>();

  public constructor(private readonly parent?: Scope) {}

  public declare(name: string, value: RuntimeValue, mutable: boolean): void {
    if (this.bindings.has(name)) {
      throw new GoJuniorRuntimeError(`${name} already declared`);
    }
    this.bindings.set(name, { value, mutable });
  }

  public hasLocal(name: string): boolean {
    return this.bindings.has(name);
  }

  public assign(name: string, value: RuntimeValue): void {
    const binding = this.bindings.get(name);
    if (binding) {
      if (!binding.mutable) {
        throw new GoJuniorRuntimeError(`${name} is const`);
      }
      binding.value = value;
      return;
    }
    if (this.parent) {
      this.parent.assign(name, value);
      return;
    }
    throw new GoJuniorRuntimeError(`${name} is not declared`);
  }

  public lookup(name: string): RuntimeValue {
    const binding = this.bindings.get(name);
    if (binding) return binding.value;
    if (this.parent) return this.parent.lookup(name);
    throw new GoJuniorRuntimeError(`${name} is not declared`);
  }
}

class SheetBinding {
  public constructor(
    public readonly name: string,
    private readonly data: SheetData
  ) {}

  public get(cell: string): RuntimeValue {
    return this.data[normalizeCell(cell)] ?? null;
  }

  public set(cell: string, value: RuntimeValue): void {
    this.data[normalizeCell(cell)] = value;
  }

  public range(start: string, end: string): RuntimeValue[] {
    const startCell = parseCellAddress(start);
    const endCell = parseCellAddress(end);
    const values: RuntimeValue[] = [];
    const rowStep = startCell.row <= endCell.row ? 1 : -1;
    const colStep = startCell.col <= endCell.col ? 1 : -1;

    for (let row = startCell.row; row !== endCell.row + rowStep; row += rowStep) {
      const rowValues: RuntimeValue[] = [];
      for (let col = startCell.col; col !== endCell.col + colStep; col += colStep) {
        rowValues.push(this.get(formatCellAddress(col, row)));
      }
      values.push(rowValues);
    }

    return values;
  }
}

interface RuntimeMapEntry {
  key: RuntimeValue;
  value: RuntimeValue;
}

const objectMapKeyIds = new WeakMap<object, number>();
let nextObjectMapKeyId = 1;

export class RuntimeMap {
  private readonly entries = new Map<string, RuntimeMapEntry>();

  public constructor(
    public readonly keyType: string,
    public readonly valueType: string
  ) {}

  public get(key: RuntimeValue): RuntimeValue {
    const entry = this.entries.get(runtimeMapKeyId(key));
    return entry ? entry.value : zeroValueForMapValue(this.valueType);
  }

  public set(key: RuntimeValue, value: RuntimeValue): void {
    assertAssignableToType(key, this.keyType, "map key");
    assertAssignableToType(value, this.valueType, "map value");
    this.entries.set(runtimeMapKeyId(key), { key, value });
  }

  public orderedEntries(): Array<[RuntimeValue, RuntimeValue]> {
    return [...this.entries.values()].map((entry) => [entry.key, entry.value]);
  }
}

export function evaluateSource(source: string, options: EvaluationOptions = {}): EvaluationResult {
  const parsed = parseGoJunior(source);
  const ast = parsed.cst ? cstToAst(parsed.cst, parsed.diagnostics) : undefined;
  if (parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return {
      diagnostics: parsed.diagnostics,
      output: [],
      ...(ast ? { ast } : {})
    };
  }
  if (!ast) {
    return {
      diagnostics: parsed.diagnostics,
      output: []
    };
  }
  return evaluateProgram(ast, options);
}

export function evaluateProgram(ast: ProgramAst, options: EvaluationOptions = {}): EvaluationResult {
  const context = new EvaluationContext(options);
  try {
    installSheets(context, options);
    installImports(context, ast);

    for (const declaration of ast.functions) {
      context.declareRoot(declaration.name, functionValue(declaration), true);
    }

    if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
      const value = context.lookup(ast.functions[0].name);
      return {
        diagnostics: ast.diagnostics,
        output: context.output,
        ast,
        value
      };
    }

    const completion = executeTopLevelStatements(ast.body, context);

    if (completion.kind === "return") {
      return {
        diagnostics: ast.diagnostics,
        output: context.output,
        ast,
        values: completion.values,
        ...(completion.values.length === 1 ? { value: completion.values[0] } : {})
      };
    }

    if (completion.kind !== "normal") throw new GoJuniorRuntimeError(completionErrorMessage(completion));

    return {
      diagnostics: ast.diagnostics,
      output: context.output,
      ast,
      ...(completion.value !== undefined ? { value: completion.value } : {})
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      diagnostics: [
        ...ast.diagnostics,
        {
          code: error instanceof GoJuniorPanic ? "GJPANIC001" : "GJRUNTIME001",
          severity: "error",
          message
        }
      ],
      output: context.output,
      ast
    };
  }
}

export class GoJuniorSession {
  private readonly context: EvaluationContext;

  public constructor(private readonly options: EvaluationOptions = {}) {
    this.context = new EvaluationContext(options);
    installSheets(this.context, options);
  }

  public setSheet(sheet: SheetData): void {
    this.context.declareOrAssignRoot("sheet", new SheetBinding("sheet", sheet), true);
  }

  public setSheets(sheets: Record<string, SheetData>): void {
    for (const [name, data] of Object.entries(sheets)) {
      this.context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
    }
  }

  public evaluate(source: string): EvaluationResult {
    const parsed = parseGoJunior(source);
    const ast = parsed.cst ? cstToAst(parsed.cst, parsed.diagnostics) : undefined;
    const hasError = parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error");
    if (hasError) {
      return {
        diagnostics: parsed.diagnostics,
        output: [],
        incomplete: diagnosticsLookIncomplete(parsed.diagnostics),
        ...(ast ? { ast } : {})
      };
    }
    if (!ast) {
      return {
        diagnostics: parsed.diagnostics,
        output: []
      };
    }

    const outputStart = this.context.output.length;
    try {
      installImports(this.context, ast);

      for (const declaration of ast.functions) {
        this.context.declareOrAssignRoot(declaration.name, functionValue(declaration), true);
      }

      if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
        const value = this.context.lookup(ast.functions[0].name);
        return {
          diagnostics: ast.diagnostics,
          output: this.context.outputFrom(outputStart),
          ast,
          value
        };
      }

      const completion = executeTopLevelStatements(ast.body, this.context);
      return resultFromCompletion(ast, this.context.outputFrom(outputStart), completion);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        diagnostics: [
          ...ast.diagnostics,
          {
            code: error instanceof GoJuniorPanic ? "GJPANIC001" : "GJRUNTIME001",
            severity: "error",
            message
          }
        ],
        output: this.context.outputFrom(outputStart),
        ast
      };
    }
  }
}

function resultFromCompletion(ast: ProgramAst, output: string[], completion: Completion): EvaluationResult {
  if (completion.kind === "return") {
    return {
      diagnostics: ast.diagnostics,
      output,
      ast,
      values: completion.values,
      ...(completion.values.length === 1 ? { value: completion.values[0] } : {})
    };
  }

  if (completion.kind !== "normal") throw new GoJuniorRuntimeError(completionErrorMessage(completion));

  return {
    diagnostics: ast.diagnostics,
    output,
    ast,
    ...(completion.value !== undefined ? { value: completion.value } : {})
  };
}

function diagnosticsLookIncomplete(diagnostics: Diagnostic[]): boolean {
  const errors = diagnostics.filter((diagnostic) => diagnostic.severity === "error");
  return errors.length > 0 && errors.every((diagnostic) =>
    diagnostic.code === "GJPARSE001" &&
    (/found\s+-->\s*''\s*<--/.test(diagnostic.message) || /but found:\s*''/.test(diagnostic.message))
  );
}

function installBuiltins(context: EvaluationContext): void {
  context.declareRoot("panic", hostCallable("panic", (args) => {
    throw new GoJuniorPanic(args[0] ?? null);
  }));
  context.declareRoot("panicOn", hostCallable("panicOn", (args) => {
    const err = args[0] ?? null;
    if (err !== null && err !== false) {
      throw new GoJuniorPanic(err);
    }
    return null;
  }));
}

function installSheets(context: EvaluationContext, options: EvaluationOptions): void {
  context.declareOrAssignRoot("sheet", new SheetBinding("sheet", options.sheet ?? {}), true);
  for (const [name, data] of Object.entries(options.sheets ?? {})) {
    if (name !== "sheet") {
      context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
    }
  }
}

function installImports(context: EvaluationContext, ast: ProgramAst): void {
  const packages = availablePackages(context);
  for (const imported of ast.imports) {
    const defaultName = imported.path.split("/").filter(Boolean).at(-1) ?? imported.path;
    const name = imported.alias ?? defaultName;
    const pkg = packages[imported.path] ?? packages[defaultName];
    if (!pkg) {
      throw new GoJuniorRuntimeError(`package ${imported.path} is not available`);
    }
    context.declareOrAssignRoot(name, pkg, true);
  }
}

function installAutomaticImports(context: EvaluationContext): void {
  context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}

function availablePackages(context: EvaluationContext): Record<string, RuntimeObject> {
  return {
    fmt: fmtPackage(),
    ...context.packages()
  };
}

function fmtPackage(): RuntimeObject {
  return {
    Printf: hostCallable("fmt.Printf", (args, context) => {
      const format = toStringValue(args[0] ?? "");
      const text = sprintf(format, args.slice(1));
      context.write(text);
      return BigInt(text.length);
    }),
    Sprintf: hostCallable("fmt.Sprintf", (args) => {
      const format = toStringValue(args[0] ?? "");
      return sprintf(format, args.slice(1));
    }),
    Println: hostCallable("fmt.Println", (args, context) => {
      const text = `${args.map(formatValue).join(" ")}\n`;
      context.write(text);
      return BigInt(text.length);
    })
  };
}

function hostCallable(
  name: string,
  call: (args: RuntimeValue[], context: EvaluationContext) => RuntimeValue
): RuntimeCallable {
  return {
    kind: "HostCallable",
    name,
    call
  };
}

function functionValue(declaration: FunctionDecl): GoJuniorFunction {
  return goJuniorFunctionValue(declaration.name, declaration.signature, declaration.body, undefined, declaration);
}

function functionLiteralValue(expression: FunctionLiteralExpression, context: EvaluationContext): GoJuniorFunction {
  return goJuniorFunctionValue("<closure>", expression.signature, expression.body, context.captureScope());
}

function executeTopLevelStatements(statements: Statement[], context: EvaluationContext): Completion {
  try {
    return executeStatements(statements, context);
  } finally {
    context.runDefers();
  }
}

function goJuniorFunctionValue(
  name: string,
  signature: FunctionDecl["signature"],
  body: FunctionDecl["body"],
  closureScope?: Scope,
  declaration?: FunctionDecl
): GoJuniorFunction {
  return {
    kind: "GoJuniorFunction",
    name,
    ...(declaration ? { declaration } : {}),
    call(args, parentContext) {
      const context = parentContext;
      const invoke = () => context.childScope(() => context.deferScope(() => {
        for (const [index, parameter] of signature.parameters.entries()) {
          if (parameter.name) {
            const value = parameter.variadic ? args.slice(index) : args[index] ?? null;
            context.declare(parameter.name, value, true);
            if (parameter.variadic) break;
          }
        }
        for (const result of signature.results) {
          if (result.name) {
            context.declare(result.name, defaultValueForDeclarationType(result.type), true);
          }
        }
        const completion = executeBlock(body, context, false);
        context.runDefers();
        if (completion.kind === "return") {
          const values = completion.values.length === 0
            ? namedReturnValues(signature, context)
            : completion.values;
          return values.length === 1 ? values[0] ?? null : values;
        }
        return null;
      }));
      return closureScope ? context.withScope(closureScope, invoke) : invoke();
    }
  };
}

function namedReturnValues(signature: FunctionDecl["signature"], context: EvaluationContext): RuntimeValue[] {
  const namedResults = signature.results.filter((result) => result.name);
  if (namedResults.length === 0) return [];
  return namedResults.map((result) => result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type));
}

function executeStatements(statements: Statement[], context: EvaluationContext): Completion {
  const labels = statementLabels(statements);
  let lastValue: RuntimeValue | undefined;
  for (let pc = 0; pc < statements.length; pc += 1) {
    const statement = statements[pc];
    if (!statement) continue;
    const completion = executeStatement(statement, context);
    if (completion.kind === "goto") {
      const target = labels.get(completion.label);
      if (target !== undefined) {
        pc = target - 1;
        lastValue = undefined;
        continue;
      }
    }
    if (completion.kind !== "normal") return completion;
    if (completion.value !== undefined) lastValue = completion.value;
  }
  return lastValue === undefined ? { kind: "normal" } : { kind: "normal", value: lastValue };
}

function statementLabels(statements: Statement[]): Map<string, number> {
  const labels = new Map<string, number>();
  for (const [index, statement] of statements.entries()) {
    if (statement.kind === "LabeledStatement") {
      labels.set(statement.label, index);
    }
  }
  return labels;
}

function executeBlock(block: BlockStatement, context: EvaluationContext, createScope = true): Completion {
  if (!createScope) return executeStatements(block.statements, context);
  return context.childScope(() => executeStatements(block.statements, context));
}

function executeStatement(statement: Statement, context: EvaluationContext): Completion {
  switch (statement.kind) {
    case "BlockStatement":
      return executeBlock(statement, context);

    case "LabeledStatement":
      return executeLabeledStatement(statement, context);

    case "ConstDecl":
      for (const declaration of statement.declarations) {
        context.declare(
          declaration.name,
          declaration.value ? evaluateExpression(declaration.value, context) : defaultValueForDeclarationType(declaration.type),
          false
        );
      }
      return { kind: "normal" };

    case "VarDecl":
      for (const declaration of statement.declarations) {
        context.declare(
          declaration.name,
          declaration.value ? evaluateExpression(declaration.value, context) : defaultValueForDeclarationType(declaration.type),
          true
        );
      }
      return { kind: "normal" };

    case "TypeDecl":
      return { kind: "normal" };

    case "ReturnStatement":
      return {
        kind: "return",
        values: statement.values.map((expression) => evaluateExpression(expression, context))
      };

    case "IfStatement":
      return executeIf(statement, context);

    case "SwitchStatement":
      return executeSwitch(statement, context);

    case "ForStatement":
      return executeFor(statement, context);

    case "DeferStatement":
      executeDefer(statement, context);
      return { kind: "normal" };

    case "BranchStatement":
      return branchCompletion(statement);

    case "AssignStatement":
      assignTarget(statement, evaluateExpression(statement.value, context), context);
      return { kind: "normal" };

    case "ShortVarStatement":
      context.declare(statement.name, evaluateExpression(statement.value, context), true);
      return { kind: "normal" };

    case "IncDecStatement":
      executeIncDec(statement, context);
      return { kind: "normal" };

    case "ExpressionStatement":
      return { kind: "normal", value: evaluateExpression(statement.expression, context) };
  }
}

function executeLabeledStatement(statement: Extract<Statement, { kind: "LabeledStatement" }>, context: EvaluationContext): Completion {
  if (!statement.statement) return { kind: "normal" };
  if (statement.statement.kind === "ForStatement") {
    return executeFor(statement.statement, context, statement.label);
  }
  if (statement.statement.kind === "SwitchStatement") {
    return executeSwitch(statement.statement, context, statement.label);
  }
  return executeStatement(statement.statement, context);
}

function executeIf(statement: IfStatement, context: EvaluationContext): Completion {
  if (toBool(evaluateExpression(statement.condition, context))) {
    return executeBlock(statement.thenBlock, context);
  }
  if (!statement.elseBranch) return { kind: "normal" };
  return statement.elseBranch.kind === "IfStatement"
    ? executeIf(statement.elseBranch, context)
    : executeBlock(statement.elseBranch, context);
}

function executeSwitch(statement: SwitchStatement, context: EvaluationContext, label?: string): Completion {
  const switchValue = statement.expression ? evaluateExpression(statement.expression, context) : true;
  let matched = false;

  for (const clause of statement.clauses) {
    if (!matched) {
      matched = clause.default || clause.values.some((value) => valueEqual(switchValue, evaluateExpression(value, context)));
    }
    if (!matched) continue;

    const completion = executeStatements(clause.statements, context);
    if (completion.kind === "fallthrough") {
      matched = true;
      continue;
    }
    if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
    return completion;
  }

  return { kind: "normal" };
}

function executeFor(statement: ForStatement, context: EvaluationContext, label?: string): Completion {
  if (statement.range) {
    const source = evaluateExpression(statement.range.source, context);
    const entries = rangeEntries(source);
    for (const [index, value] of entries) {
      const completion = context.childScope(() => {
        if (statement.range?.keyName) {
          if (statement.range.define) context.declare(statement.range.keyName, index, true);
          else context.assign(statement.range.keyName, index);
        }
        if (statement.range?.valueName) {
          if (statement.range.define) context.declare(statement.range.valueName, value, true);
          else context.assign(statement.range.valueName, value);
        }
        return executeBlock(statement.body, context, false);
      });
      if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
      if (completion.kind === "continue" && labelMatches(completion.label, label)) continue;
      if (completion.kind !== "normal") return completion;
    }
    return { kind: "normal" };
  }

  return context.childScope(() => {
    if (statement.init) {
      expectNormalCompletion(executeStatement(statement.init, context), "for init statement");
    }

    for (let iteration = 0; iteration < context.loopLimit(); iteration += 1) {
      if (statement.condition && !toBool(evaluateExpression(statement.condition, context))) {
        return { kind: "normal" };
      }

      const completion = executeBlock(statement.body, context);
      if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
      if (completion.kind === "continue" && labelMatches(completion.label, label)) {
        if (statement.post) {
          expectNormalCompletion(executeStatement(statement.post, context), "for post statement");
        }
        continue;
      }
      if (completion.kind !== "normal") return completion;

      if (statement.post) {
        expectNormalCompletion(executeStatement(statement.post, context), "for post statement");
      }
    }

    throw new GoJuniorRuntimeError(`loop exceeded ${context.loopLimit()} iterations`);
  });
}

function expectNormalCompletion(completion: Completion, label: string): void {
  if (completion.kind !== "normal") {
    throw new GoJuniorRuntimeError(`${completionDescription(completion)} used in ${label}`);
  }
}

function labelMatches(completionLabel: string | undefined, activeLabel: string | undefined): boolean {
  return completionLabel === undefined || completionLabel === activeLabel;
}

function completionErrorMessage(completion: Exclude<Completion, { kind: "normal" | "return" }>): string {
  if (completion.kind === "goto") return `unresolved goto label ${completion.label}`;
  return `${completionDescription(completion)} used outside a matching loop or switch`;
}

function completionDescription(completion: Exclude<Completion, { kind: "normal" }>): string {
  if (completion.kind === "return") return "return";
  if (completion.kind === "fallthrough") return "fallthrough";
  return completion.label ? `${completion.kind} ${completion.label}` : completion.kind;
}

function executeDefer(statement: DeferStatement, context: EvaluationContext): void {
  if (statement.expression.kind === "CallExpression") {
    const callee = evaluateExpression(statement.expression.callee, context);
    const args = statement.expression.args.map((arg) => evaluateExpression(arg, context));
    context.pushDefer(() => callRuntime(callee, args, context));
    return;
  }

  context.pushDefer(() => evaluateExpression(statement.expression, context));
}

function branchCompletion(statement: BranchStatement): Completion {
  if (statement.branch === "break") {
    return statement.label ? { kind: "break", label: statement.label } : { kind: "break" };
  }
  if (statement.branch === "continue") {
    return statement.label ? { kind: "continue", label: statement.label } : { kind: "continue" };
  }
  if (statement.branch === "goto") return { kind: "goto", label: statement.label ?? "<missing>" };
  return { kind: "fallthrough" };
}

function assignTarget(statement: AssignStatement, value: RuntimeValue, context: EvaluationContext): void {
  assignExpressionTarget(statement.target, value, context);
}

function executeIncDec(statement: IncDecStatement, context: EvaluationContext): void {
  const current = evaluateExpression(statement.target, context);
  const next = statement.operator === "++"
    ? addNumbers(current, 1n)
    : subtractNumbers(current, 1n);
  assignExpressionTarget(statement.target, next, context);
}

function assignExpressionTarget(target: Expression, value: RuntimeValue, context: EvaluationContext): void {
  if (target.kind === "Identifier") {
    context.assign(target.name, value);
    return;
  }
  if (target.kind === "SelectorExpression") {
    setSelector(target, value, context);
    return;
  }
  if (target.kind === "IndexExpression") {
    setIndex(target, value, context);
    return;
  }
  throw new GoJuniorRuntimeError("unsupported assignment target");
}

function evaluateExpression(expression: Expression, context: EvaluationContext): RuntimeValue {
  switch (expression.kind) {
    case "Identifier":
      return evaluateIdentifier(expression, context);

    case "Literal":
      return expression.value;

    case "FunctionLiteralExpression":
      return functionLiteralValue(expression, context);

    case "MapLiteralExpression":
      return evaluateMapLiteral(expression, context);

    case "UnaryExpression":
      return evaluateUnary(expression, context);

    case "BinaryExpression":
      return evaluateBinary(expression, context);

    case "SelectorExpression":
      return getSelector(expression, context);

    case "CallExpression":
      return evaluateCall(expression, context);

    case "IndexExpression":
      return getIndex(evaluateExpression(expression.object, context), evaluateExpression(expression.index, context));

    case "SliceExpression":
      return getSlice(
        evaluateExpression(expression.object, context),
        expression.start ? evaluateExpression(expression.start, context) : undefined,
        expression.end ? evaluateExpression(expression.end, context) : undefined
      );

    case "SpreadsheetRangeExpression":
      return getSpreadsheetRange(expression, context);
  }
}

function evaluateMapLiteral(expression: MapLiteralExpression, context: EvaluationContext): RuntimeMap {
  const map = new RuntimeMap(expression.keyType.text, expression.valueType.text);
  for (const entry of expression.entries) {
    map.set(evaluateExpression(entry.key, context), evaluateExpression(entry.value, context));
  }
  return map;
}

function evaluateIdentifier(expression: IdentifierExpression, context: EvaluationContext): RuntimeValue {
  if (expression.name === "nil") return null;
  return context.lookup(expression.name);
}

function evaluateUnary(expression: UnaryExpression, context: EvaluationContext): RuntimeValue {
  const value = evaluateExpression(expression.operand, context);
  switch (expression.operator) {
    case "+":
      return numericIdentity(value);
    case "-":
      return negateNumber(value);
    case "!":
      return !toBool(value);
    case "&":
      throw new GoJuniorRuntimeError("address-of is parsed but not implemented in the runtime slice yet");
    case "*":
      throw new GoJuniorRuntimeError("pointer dereference is parsed but not implemented in the runtime slice yet");
  }
}

function evaluateBinary(expression: BinaryExpression, context: EvaluationContext): RuntimeValue {
  if (expression.operator === "&&") {
    return toBool(evaluateExpression(expression.left, context)) && toBool(evaluateExpression(expression.right, context));
  }
  if (expression.operator === "||") {
    return toBool(evaluateExpression(expression.left, context)) || toBool(evaluateExpression(expression.right, context));
  }

  const left = evaluateExpression(expression.left, context);
  const right = evaluateExpression(expression.right, context);

  switch (expression.operator) {
    case "==":
      return valueEqual(left, right);
    case "!=":
      return !valueEqual(left, right);
    case "<":
      return compareValues(left, right) < 0;
    case "<=":
      return compareValues(left, right) <= 0;
    case ">":
      return compareValues(left, right) > 0;
    case ">=":
      return compareValues(left, right) >= 0;
    case "+":
      if (typeof left === "string" || typeof right === "string") return `${formatValue(left)}${formatValue(right)}`;
      return addNumbers(left, right);
    case "-":
      return subtractNumbers(left, right);
    case "*":
      return multiplyNumbers(left, right);
    case "/":
      return divideNumbers(left, right);
    case "%":
      return moduloNumbers(left, right);
  }
}

function evaluateCall(expression: CallExpression, context: EvaluationContext): RuntimeValue {
  const callee = evaluateExpression(expression.callee, context);
  const args = expression.args.map((arg) => evaluateExpression(arg, context));
  return callRuntime(callee, expression.spreadLast ? spreadLastArgument(args) : args, context);
}

function spreadLastArgument(args: RuntimeValue[]): RuntimeValue[] {
  if (args.length === 0) return args;
  const last = args[args.length - 1];
  if (!Array.isArray(last)) {
    throw new GoJuniorRuntimeError(`${formatValue(last ?? null)} is not spreadable`);
  }
  return [...args.slice(0, -1), ...last];
}

function callRuntime(callee: RuntimeValue, args: RuntimeValue[], context: EvaluationContext): RuntimeValue {
  if (isRuntimeCallable(callee) || isGoJuniorFunction(callee)) {
    return callee.call(args, context);
  }
  throw new GoJuniorRuntimeError(`${formatValue(callee)} is not callable`);
}

function getSelector(expression: SelectorExpression, context: EvaluationContext): RuntimeValue {
  const object = evaluateExpression(expression.object, context);
  if (object instanceof SheetBinding) {
    return object.get(expression.field);
  }
  if (isRuntimeObject(object)) {
    const value = object[expression.field];
    if (value !== undefined) return value;
  }
  throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}

function setSelector(expression: SelectorExpression, value: RuntimeValue, context: EvaluationContext): void {
  const object = evaluateExpression(expression.object, context);
  if (object instanceof SheetBinding) {
    object.set(expression.field, value);
    return;
  }
  if (isRuntimeObject(object)) {
    object[expression.field] = value;
    return;
  }
  throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}

function getSpreadsheetRange(expression: SpreadsheetRangeExpression, context: EvaluationContext): RuntimeValue {
  const sheet = evaluateExpression(expression.start.object, context);
  if (!(sheet instanceof SheetBinding)) {
    throw new GoJuniorRuntimeError("spreadsheet ranges must start with a sheet namespace");
  }
  return sheet.range(expression.start.field, expression.endCell);
}

function getIndex(object: RuntimeValue, index: RuntimeValue): RuntimeValue {
  if (object instanceof RuntimeMap) return object.get(index);
  if (isRuntimeObject(object)) return object[String(index)] ?? null;
  const numericIndex = toNumber(index);
  if (Array.isArray(object)) return object[numericIndex] ?? null;
  if (typeof object === "string") return object[numericIndex] ?? "";
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
}

function setIndex(expression: IndexExpression, value: RuntimeValue, context: EvaluationContext): void {
  const object = evaluateExpression(expression.object, context);
  const index = evaluateExpression(expression.index, context);
  if (object instanceof RuntimeMap) {
    object.set(index, value);
    return;
  }
  if (isRuntimeObject(object)) {
    object[String(index)] = value;
    return;
  }
  const numericIndex = toNumber(index);
  if (Array.isArray(object)) {
    object[numericIndex] = value;
    return;
  }
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
}

function getSlice(object: RuntimeValue, start: RuntimeValue | undefined, end: RuntimeValue | undefined): RuntimeValue {
  const startIndex = start === undefined ? undefined : toNumber(start);
  const endIndex = end === undefined ? undefined : toNumber(end);
  if (Array.isArray(object)) return object.slice(startIndex, endIndex);
  if (typeof object === "string") return object.slice(startIndex, endIndex);
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not sliceable`);
}

function defaultValueForDeclarationType(type: TypeNode | undefined): RuntimeValue {
  const typeText = type?.text ?? "";
  const mapType = parseMapTypeText(typeText);
  if (mapType) return new RuntimeMap(mapType.keyType, mapType.valueType);
  return zeroValueForMapValue(typeText);
}

function zeroValueForMapValue(typeText: string): RuntimeValue {
  const type = normalizeTypeText(typeText);
  if (isIntegerType(type)) return 0n;
  if (isFloatType(type)) return 0;
  if (type === "string") return "";
  if (type === "bool") return false;
  return null;
}

function assertAssignableToType(value: RuntimeValue, typeText: string, role: string): void {
  const type = normalizeTypeText(typeText);
  if (!type || type === "<missing>" || type === "any" || type === "interface{}") return;

  if (value === null) {
    if (isNilAssignableType(type)) return;
    throw new GoJuniorRuntimeError(`${role} nil is not assignable to ${type}`);
  }

  if (type === "string") {
    if (typeof value !== "string") throwTypeError(value, type, role);
    return;
  }
  if (type === "bool") {
    if (typeof value !== "boolean") throwTypeError(value, type, role);
    return;
  }
  if (isIntegerType(type)) {
    if (typeof value !== "bigint" || !integerInRange(value, type)) throwTypeError(value, type, role);
    return;
  }
  if (isFloatType(type)) {
    if (typeof value !== "number") throwTypeError(value, type, role);
    return;
  }
  if (parseMapTypeText(type)) {
    if (!(value instanceof RuntimeMap)) throwTypeError(value, type, role);
    return;
  }
  if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type)) {
    if (!Array.isArray(value)) throwTypeError(value, type, role);
    return;
  }
  if (type.startsWith("func(")) {
    if (!isRuntimeCallable(value) && !isGoJuniorFunction(value)) throwTypeError(value, type, role);
    return;
  }
}

function throwTypeError(value: RuntimeValue, type: string, role: string): never {
  throw new GoJuniorRuntimeError(`${role} ${formatValue(value)} is not assignable to ${type}`);
}

function parseMapTypeText(typeText: string): { keyType: string; valueType: string } | undefined {
  const type = normalizeTypeText(typeText);
  if (!type.startsWith("map[")) return undefined;

  let depth = 0;
  for (let index = 3; index < type.length; index += 1) {
    const char = type[index];
    if (char === "[") {
      depth += 1;
      continue;
    }
    if (char === "]") {
      depth -= 1;
      if (depth === 0) {
        const keyType = type.slice(4, index);
        const valueType = type.slice(index + 1);
        if (!keyType || !valueType) return undefined;
        return { keyType, valueType };
      }
    }
  }
  return undefined;
}

function normalizeTypeText(typeText: string): string {
  return typeText.replace(/\s+/g, "");
}

function runtimeMapKeyId(key: RuntimeValue): string {
  if (key === null) return "nil";
  if (typeof key === "boolean") return `b:${key}`;
  if (typeof key === "string") return `s:${key}`;
  if (typeof key === "bigint") return `i:${key}`;
  if (typeof key === "number") return `f:${Object.is(key, -0) ? "-0" : String(key)}`;
  return `o:${objectIdentityId(key)}`;
}

function objectIdentityId(value: object): number {
  const existing = objectMapKeyIds.get(value);
  if (existing !== undefined) return existing;
  const id = nextObjectMapKeyId;
  nextObjectMapKeyId += 1;
  objectMapKeyIds.set(value, id);
  return id;
}

function isNilAssignableType(type: string): boolean {
  return type === "error" ||
    type === "any" ||
    type === "interface{}" ||
    type.startsWith("*") ||
    type.startsWith("[]") ||
    type.startsWith("map[") ||
    type.startsWith("func(");
}

function isIntegerType(type: string): boolean {
  return /^(?:u?int(?:8|16|32|64)?|byte|rune)$/.test(type);
}

function isFloatType(type: string): boolean {
  return type === "float32" || type === "float64";
}

function integerInRange(value: bigint, type: string): boolean {
  const bounds = integerTypeBounds(type);
  return !bounds || (value >= bounds.min && value <= bounds.max);
}

function integerTypeBounds(type: string): { min: bigint; max: bigint } | undefined {
  switch (type) {
    case "uint8":
    case "byte":
      return { min: 0n, max: 255n };
    case "int8":
      return { min: -128n, max: 127n };
    case "uint16":
      return { min: 0n, max: 65535n };
    case "int16":
      return { min: -32768n, max: 32767n };
    case "uint32":
      return { min: 0n, max: 4294967295n };
    case "int32":
    case "rune":
      return { min: -2147483648n, max: 2147483647n };
    case "uint":
    case "uint64":
      return { min: 0n, max: 18446744073709551615n };
    case "int":
    case "int64":
      return { min: -9223372036854775808n, max: 9223372036854775807n };
    default:
      return undefined;
  }
}

function addNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  if (typeof left === "bigint" && typeof right === "bigint") return left + right;
  return toFloat(left) + toFloat(right);
}

function subtractNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  if (typeof left === "bigint" && typeof right === "bigint") return left - right;
  return toFloat(left) - toFloat(right);
}

function multiplyNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  if (typeof left === "bigint" && typeof right === "bigint") return left * right;
  return toFloat(left) * toFloat(right);
}

function divideNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  if (typeof left === "bigint" && typeof right === "bigint") return left / right;
  return toFloat(left) / toFloat(right);
}

function moduloNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  if (typeof left === "bigint" && typeof right === "bigint") return left % right;
  return toFloat(left) % toFloat(right);
}

function negateNumber(value: RuntimeValue): RuntimeValue {
  if (typeof value === "bigint") return -value;
  return -toFloat(value);
}

function numericIdentity(value: RuntimeValue): RuntimeValue {
  if (typeof value === "bigint" || typeof value === "number") return value;
  throw new GoJuniorRuntimeError(`${formatValue(value)} is not numeric`);
}

function toBool(value: RuntimeValue): boolean {
  if (typeof value !== "boolean") {
    throw new GoJuniorRuntimeError(`${formatValue(value)} is not a bool`);
  }
  return value;
}

function toFloat(value: RuntimeValue): number {
  if (typeof value === "number") return value;
  if (typeof value === "bigint") return Number(value);
  throw new GoJuniorRuntimeError(`${formatValue(value)} is not numeric`);
}

function toNumber(value: RuntimeValue): number {
  const number = toFloat(value);
  if (!Number.isInteger(number)) {
    throw new GoJuniorRuntimeError(`${formatValue(value)} is not an integer index`);
  }
  return number;
}

function compareValues(left: RuntimeValue, right: RuntimeValue): number {
  if ((typeof left === "bigint" || typeof left === "number") && (typeof right === "bigint" || typeof right === "number")) {
    const leftNumber = toFloat(left);
    const rightNumber = toFloat(right);
    return leftNumber === rightNumber ? 0 : leftNumber < rightNumber ? -1 : 1;
  }
  if (typeof left === "string" && typeof right === "string") {
    return left === right ? 0 : left < right ? -1 : 1;
  }
  throw new GoJuniorRuntimeError(`cannot compare ${formatValue(left)} and ${formatValue(right)}`);
}

function valueEqual(left: RuntimeValue, right: RuntimeValue): boolean {
  if ((typeof left === "bigint" || typeof left === "number") && (typeof right === "bigint" || typeof right === "number")) {
    return toFloat(left) === toFloat(right);
  }
  return left === right;
}

function rangeEntries(source: RuntimeValue): Array<[RuntimeValue, RuntimeValue]> {
  if (Array.isArray(source)) {
    return source.map((value, index) => [BigInt(index), value]);
  }
  if (typeof source === "string") {
    return [...source].map((value, index) => [BigInt(index), value]);
  }
  if (source instanceof RuntimeMap) {
    return source.orderedEntries();
  }
  if (isRuntimeObject(source)) {
    return Object.entries(source).map(([key, value]) => [key, value]);
  }
  throw new GoJuniorRuntimeError(`${formatValue(source)} is not rangeable`);
}

function sprintf(format: string, args: RuntimeValue[]): string {
  let argIndex = 0;
  return format.replace(/%(#)?[%vdsft]/g, (match, alternate: string | undefined) => {
    if (match === "%%") return "%";
    const value = args[argIndex++] ?? null;
    const goSyntax = alternate === "#";
    switch (match) {
      case "%d":
        return String(toNumber(value));
      case "%f":
        return String(toFloat(value));
      case "%s":
        return toStringValue(value);
      case "%t":
        return String(toBool(value));
      case "%v":
        return formatValue(value);
      case "%#v":
        return formatGoSyntaxValue(value);
      default:
        return goSyntax ? formatGoSyntaxValue(value) : formatValue(value);
    }
  });
}

function toStringValue(value: RuntimeValue): string {
  if (typeof value === "string") return value;
  return formatValue(value);
}

export function formatValue(value: RuntimeValue): string {
  if (value === null) return "<nil>";
  if (typeof value === "bigint") return value.toString();
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "string") return String(value);
  if (Array.isArray(value)) return `[${value.map(formatValue).join(" ")}]`;
  if (value instanceof RuntimeMap) return formatRuntimeMap(value);
  if (isRuntimeCallable(value) || isGoJuniorFunction(value)) return `<func ${value.name}>`;
  if (value instanceof SheetBinding) return `<sheet ${value.name}>`;
  return `{${Object.entries(value).map(([key, item]) => `${key}:${formatValue(item)}`).join(" ")}}`;
}

function formatGoSyntaxValue(value: RuntimeValue): string {
  if (value === null) return "nil";
  if (typeof value === "bigint") return `${integerTypeName(value)}(${value.toString()})`;
  if (typeof value === "number") return `float64(${formatFloat(value)})`;
  if (typeof value === "boolean") return `bool(${value})`;
  if (typeof value === "string") return `string(${JSON.stringify(value)})`;
  if (Array.isArray(value)) return `[]interface{}{${value.map(formatGoSyntaxValue).join(", ")}}`;
  if (value instanceof RuntimeMap) return formatGoSyntaxMap(value);
  if (isRuntimeCallable(value) || isGoJuniorFunction(value)) return `<func ${value.name}>`;
  if (value instanceof SheetBinding) return `<sheet ${value.name}>`;
  return `map[string]interface{}{${Object.entries(value)
    .map(([key, item]) => `${JSON.stringify(key)}: ${formatGoSyntaxValue(item)}`)
    .join(", ")}}`;
}

function formatRuntimeMap(value: RuntimeMap): string {
  const entries = value.orderedEntries()
    .map(([key, item]) => `${formatValue(key)}:${formatValue(item)}`)
    .join(" ");
  return `map[${entries}]`;
}

function formatGoSyntaxMap(value: RuntimeMap): string {
  const entries = value.orderedEntries()
    .map(([key, item]) => `${formatGoSyntaxValue(key)}: ${formatGoSyntaxValue(item)}`)
    .join(", ");
  return `map[${value.keyType}]${value.valueType}{${entries}}`;
}

function integerTypeName(value: bigint): "int64" | "bigint" {
  return value >= -9223372036854775808n && value <= 9223372036854775807n ? "int64" : "bigint";
}

function formatFloat(value: number): string {
  if (Number.isNaN(value)) return "NaN";
  if (value === Infinity) return "+Inf";
  if (value === -Infinity) return "-Inf";
  return Number.isInteger(value) ? `${value}.0` : String(value);
}

function isRuntimeObject(value: RuntimeValue): value is RuntimeObject {
  return Boolean(
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    !(value instanceof RuntimeMap) &&
    !(value instanceof SheetBinding) &&
    !isRuntimeCallable(value) &&
    !isGoJuniorFunction(value)
  );
}

function isRuntimeCallable(value: RuntimeValue): value is RuntimeCallable {
  return Boolean(value && typeof value === "object" && "kind" in value && value.kind === "HostCallable");
}

function isGoJuniorFunction(value: RuntimeValue): value is GoJuniorFunction {
  return Boolean(value && typeof value === "object" && "kind" in value && value.kind === "GoJuniorFunction");
}

function normalizeCell(cell: string): string {
  return cell.replace(/\$/g, "").toUpperCase();
}

function parseCellAddress(cell: string): { col: number; row: number } {
  const normalized = normalizeCell(cell);
  const match = /^([A-Z]+)([1-9][0-9]*)$/.exec(normalized);
  if (!match) throw new GoJuniorRuntimeError(`${cell} is not a spreadsheet cell address`);
  return {
    col: columnNameToNumber(match[1] ?? "A"),
    row: Number(match[2] ?? "1")
  };
}

function columnNameToNumber(name: string): number {
  let value = 0;
  for (const char of name) {
    value = value * 26 + (char.charCodeAt(0) - 64);
  }
  return value;
}

function formatCellAddress(col: number, row: number): string {
  let name = "";
  let value = col;
  while (value > 0) {
    const remainder = (value - 1) % 26;
    name = String.fromCharCode(65 + remainder) + name;
    value = Math.floor((value - 1) / 26);
  }
  return `${name}${row}`;
}
