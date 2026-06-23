import {
  AssignStatement,
  ArrayLiteralExpression,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  ComplexLiteralValue,
  DeferStatement,
  Expression,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  GoStatement,
  IdentifierExpression,
  IfStatement,
  IncDecStatement,
  IndexExpression,
  MapLiteralExpression,
  ProgramAst,
  SelectorExpression,
  SelectStatement,
  SendStatement,
  ShortVarStatement,
  SpreadsheetRangeExpression,
  Statement,
  StructFieldDecl,
  StructLiteralExpression,
  SwitchStatement,
  TypeAssertionExpression,
  TypeExpression,
  TypeNode,
  TypeSpec,
  UnaryExpression
} from "./ast.js";
import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan } from "./diagnostics.js";
import { checkFrontSource, checkFrontSourceFiles, CheckConfig, SheetNamespace } from "./front/checker.js";
import {
  BasicKind,
  BasicType,
  MapType as CheckerMapType,
  SliceType as CheckerSliceType,
  Type as CheckerType,
  Universe,
  newUniverse
} from "./front/types.js";
import { frontSourceFilesToAst, frontSourceToAst } from "./frontToAst.js";
import { DeterministicPrng } from "./prng.js";
import {
  AsyncGoChannel,
  AsyncGoDeadlockError,
  AsyncGoPanic,
  AsyncGoScheduler,
  asyncSelect
} from "./asyncRuntime.js";
import type { AsyncSelectCase, AsyncSelectResult } from "./asyncRuntime.js";

export type RuntimeValue =
  | null
  | boolean
  | string
  | number
  | bigint
  | ComplexLiteralValue
  | RuntimeValue[]
  | RuntimeChannel
  | RuntimeMap
  | RuntimeStruct
  | RuntimePointer
  | RuntimeNamedValue
  | RuntimeInterfaceValue
  | RuntimeTypedNilValue
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
  call(args: RuntimeValue[], context: EvaluationContext): MaybePromise<RuntimeValue>;
}

export interface GoJuniorFunction {
  kind: "GoJuniorFunction";
  name: string;
  declaration?: FunctionDecl;
  call(args: RuntimeValue[], context: EvaluationContext): MaybePromise<RuntimeValue>;
}

export interface SheetData {
  [cell: string]: RuntimeValue;
}

export interface EvaluationOptions {
  filename?: string;
  packages?: Record<string, RuntimeObject>;
  sheet?: SheetData;
  sheets?: Record<string, SheetData>;
  currentSheetName?: string;
  stdout?: (text: string) => void;
  maxLoopIterations?: number;
  randomSeed?: number | string | bigint;
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

type MaybePromise<T> = T | Promise<T>;

interface Binding {
  value: RuntimeValue;
  mutable: boolean;
  typeText?: string;
}

interface StructTypeDef {
  name: string;
  fields: StructFieldDecl[];
}

interface InterfaceTypeDef {
  name: string;
  methods: NonNullable<TypeSpec["interfaceMethods"]>;
  embeds: TypeNode[];
}

interface MethodDef {
  declaration: FunctionDecl;
  receiverType: string;
  pointerReceiver: boolean;
}

export class GoJuniorRuntimeError extends Error {
  public constructor(message: string) {
    super(message);
    this.name = "GoJuniorRuntimeError";
  }
}

export class GoJuniorDeadlockError extends Error {
  public constructor(message: string) {
    super(message);
    this.name = "GoJuniorDeadlockError";
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

class GoJuniorTestStop extends Error {
  public constructor(public readonly action: "failNow" | "skipNow") {
    super(action);
    this.name = "GoJuniorTestStop";
  }
}

interface TestingTState {
  name: string;
  failed: boolean;
  skipped: boolean;
  logs: string[];
}

interface EvaluationSharedState {
  output: string[];
  rootScope: Scope;
  types: Map<string, StructTypeDef>;
  interfaces: Map<string, InterfaceTypeDef>;
  aliases: Map<string, string>;
  methods: Map<string, MethodDef>;
  maxLoopIterations: number;
  stdout?: (text: string) => void;
  random: DeterministicPrng;
  scheduler: AsyncGoScheduler;
}

export class EvaluationContext {
  private readonly shared: EvaluationSharedState;
  private currentScope: Scope;
  private readonly deferFrames: Array<Array<() => MaybePromise<RuntimeValue>>> = [[]];

  public constructor(
    private readonly options: EvaluationOptions = {},
    shared?: EvaluationSharedState,
    currentScope?: Scope
  ) {
    if (shared) {
      this.shared = shared;
      this.currentScope = currentScope ?? shared.rootScope;
      return;
    }
    this.shared = {
      output: [],
      rootScope: new Scope(),
      types: new Map(),
      interfaces: new Map(),
      aliases: new Map(),
      methods: new Map(),
      maxLoopIterations: options.maxLoopIterations ?? 100_000,
      ...(options.stdout ? { stdout: options.stdout } : {}),
      random: new DeterministicPrng(options.randomSeed),
      scheduler: new AsyncGoScheduler(options.randomSeed === undefined ? {} : { randomSeed: options.randomSeed })
    };
    this.currentScope = this.shared.rootScope;
    installBuiltins(this);
    installAutomaticImports(this);
  }

  public get output(): string[] {
    return this.shared.output;
  }

  public scheduler(): AsyncGoScheduler {
    return this.shared.scheduler;
  }

  public fork(currentScope: Scope = this.shared.rootScope): EvaluationContext {
    return new EvaluationContext(this.options, this.shared, currentScope);
  }

  public declare(name: string, value: RuntimeValue, mutable = true, type?: TypeNode | string): void {
    const typeText = bindingTypeText(type);
    const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
    this.currentScope.declare(name, stored, mutable, typeText);
  }

  public declareRoot(name: string, value: RuntimeValue, mutable = true, type?: TypeNode | string): void {
    const typeText = bindingTypeText(type);
    const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
    this.shared.rootScope.declare(name, stored, mutable, typeText);
  }

  public declareOrAssignRoot(name: string, value: RuntimeValue, mutable = true, type?: TypeNode | string): void {
    if (this.shared.rootScope.hasLocal(name)) {
      this.shared.rootScope.assign(name, value, this.assignmentChecker());
      return;
    }
    this.declareRoot(name, value, mutable, type);
  }

  public assign(name: string, value: RuntimeValue): void {
    this.currentScope.assign(name, value, this.assignmentChecker());
  }

  public lookup(name: string): RuntimeValue {
    return this.currentScope.lookup(name);
  }

  public lookupTypeText(name: string): string | undefined {
    return this.currentScope.lookupTypeText(name);
  }

  public hasLocal(name: string): boolean {
    return this.currentScope.hasLocal(name);
  }

  public pointerToBinding(name: string, typeName: string): RuntimePointer {
    const scope = this.currentScope;
    return new RuntimePointer(typeName, () => scope.lookup(name), (next) => scope.assign(name, next, this.assignmentChecker()));
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

  public async childScopeAsync<T>(body: () => Promise<T>): Promise<T> {
    const previous = this.currentScope;
    this.currentScope = new Scope(previous);
    try {
      return await body();
    } finally {
      this.currentScope = previous;
    }
  }

  public async withScopeAsync<T>(scope: Scope, body: () => Promise<T>): Promise<T> {
    const previous = this.currentScope;
    this.currentScope = scope;
    try {
      return await body();
    } finally {
      this.currentScope = previous;
    }
  }

  public pushDefer(callback: () => MaybePromise<RuntimeValue>): void {
    this.currentDeferFrame().push(callback);
  }

  public runDefers(): void {
    const frame = this.currentDeferFrame();
    while (frame.length > 0) {
      frame.pop()?.();
    }
  }

  public async runDefersAsync(): Promise<void> {
    const frame = this.currentDeferFrame();
    while (frame.length > 0) {
      await frame.pop()?.();
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

  public async deferScopeAsync<T>(body: () => Promise<T>): Promise<T> {
    this.deferFrames.push([]);
    try {
      return await body();
    } finally {
      try {
        await this.runDefersAsync();
      } finally {
        this.deferFrames.pop();
      }
    }
  }

  public write(text: string): void {
    this.shared.output.push(text);
    this.shared.stdout?.(text);
  }

  public loopLimit(): number {
    return this.shared.maxLoopIterations;
  }

  public randomIndex(length: number): number {
    return this.shared.random.nextIndex(length);
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

  public registerType(spec: TypeSpec): void {
    this.shared.aliases.set(spec.name, spec.type.text);
    if (spec.structFields) {
      this.shared.types.set(spec.name, {
        name: spec.name,
        fields: spec.structFields
      });
    }
    if (spec.interfaceMethods || spec.interfaceEmbeds) {
      this.shared.interfaces.set(spec.name, {
        name: spec.name,
        methods: this.flattenInterfaceMethods(spec.interfaceMethods ?? [], spec.interfaceEmbeds ?? []),
        embeds: spec.interfaceEmbeds ?? []
      });
    }
  }

  public typeDef(name: string): StructTypeDef | undefined {
    return this.shared.types.get(name);
  }

  public interfaceDef(name: string): InterfaceTypeDef | undefined {
    return this.shared.interfaces.get(name);
  }

  public aliasType(name: string): string | undefined {
    return this.shared.aliases.get(name);
  }

  public isKnownType(name: string): boolean {
    const type = normalizeTypeText(name);
    return isPredeclaredType(type) ||
      this.shared.aliases.has(type) ||
      this.shared.types.has(type) ||
      this.shared.interfaces.has(type) ||
      type.startsWith("*") ||
      type.startsWith("[]") ||
      /^\[[0-9.]*\]/.test(type) ||
      type.startsWith("map[") ||
      parseChanTypeText(type) !== undefined ||
      type.startsWith("func(");
  }

  public registerMethod(declaration: FunctionDecl): void {
    if (!declaration.receiver) return;
    const receiver = normalizeReceiverType(declaration.receiver.type.text);
    this.shared.methods.set(methodKey(receiver.baseType, declaration.name), {
      declaration,
      receiverType: receiver.baseType,
      pointerReceiver: receiver.pointer
    });
  }

  public methodFor(typeName: string, methodName: string): MethodDef | undefined {
    return this.shared.methods.get(methodKey(typeName, methodName));
  }

  private flattenInterfaceMethods(methods: NonNullable<TypeSpec["interfaceMethods"]>, embeds: TypeNode[]): NonNullable<TypeSpec["interfaceMethods"]> {
    const flattened = [...methods];
    for (const embed of embeds) {
      const embedded = this.interfaceDef(embed.text);
      if (!embedded) continue;
      for (const method of embedded.methods) {
        if (!flattened.some((candidate) => candidate.name === method.name)) flattened.push(method);
      }
    }
    return flattened;
  }

  private currentDeferFrame(): Array<() => MaybePromise<RuntimeValue>> {
    const frame = this.deferFrames[this.deferFrames.length - 1];
    if (!frame) throw new GoJuniorRuntimeError("internal error: missing defer frame");
    return frame;
  }

  private assignmentChecker(): AssignmentChecker {
    return (value, typeText, role) => prepareAssignableToType(value, typeText, role, this);
  }
}

type AssignmentChecker = (value: RuntimeValue, typeText: string, role: string) => RuntimeValue;

class Scope {
  private readonly bindings = new Map<string, Binding>();

  public constructor(private readonly parent?: Scope) {}

  public declare(name: string, value: RuntimeValue, mutable: boolean, typeText?: string): void {
    if (this.bindings.has(name)) {
      throw new GoJuniorRuntimeError(`${name} already declared`);
    }
    this.bindings.set(name, { value, mutable, ...(typeText ? { typeText } : {}) });
  }

  public hasLocal(name: string): boolean {
    return this.bindings.has(name);
  }

  public assign(name: string, value: RuntimeValue, checker?: AssignmentChecker): void {
    const binding = this.bindings.get(name);
    if (binding) {
      if (!binding.mutable) {
        throw new GoJuniorRuntimeError(`${name} is const`);
      }
      if (binding.typeText) {
        value = checker ? checker(value, binding.typeText, `variable ${name}`) : value;
      }
      binding.value = value;
      return;
    }
    if (this.parent) {
      this.parent.assign(name, value, checker);
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

  public lookupTypeText(name: string): string | undefined {
    const binding = this.bindings.get(name);
    if (binding) return binding.typeText;
    return this.parent?.lookupTypeText(name);
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

export class RuntimeStruct {
  public readonly fields = new Map<string, RuntimeValue>();

  public constructor(
    public readonly typeName: string,
    initialFields: Iterable<[string, RuntimeValue]> = []
  ) {
    for (const [name, value] of initialFields) {
      this.fields.set(name, value);
    }
  }

  public get(field: string): RuntimeValue | undefined {
    return this.fields.get(field);
  }

  public set(field: string, value: RuntimeValue): void {
    this.fields.set(field, value);
  }

  public clone(): RuntimeStruct {
    return new RuntimeStruct(this.typeName, this.fields.entries());
  }

  public orderedFields(): Array<[string, RuntimeValue]> {
    return [...this.fields.entries()];
  }
}

export class RuntimePointer {
  public constructor(
    public readonly typeName: string,
    private readonly getValue: () => RuntimeValue,
    private readonly setValue: (value: RuntimeValue) => void
  ) {}

  public get(): RuntimeValue {
    return this.getValue();
  }

  public set(value: RuntimeValue): void {
    this.setValue(value);
  }
}

export class RuntimeNamedValue {
  public constructor(
    public readonly typeName: string,
    public readonly value: RuntimeValue
  ) {}
}

export class RuntimeInterfaceValue {
  public constructor(
    public readonly interfaceType: string,
    public readonly value: RuntimeValue
  ) {}

  public isNil(): boolean {
    return this.value === null;
  }
}

export class RuntimeTypedNilValue {
  public constructor(public readonly typeName: string) {}
}

interface RuntimeMapEntry {
  key: RuntimeValue;
  value: RuntimeValue;
}

const objectMapKeyIds = new WeakMap<object, number>();
const arrayCapacities = new WeakMap<RuntimeValue[], number>();
let nextObjectMapKeyId = 1;

export class RuntimeMap {
  private readonly entries = new Map<string, RuntimeMapEntry>();

  public constructor(
    public readonly keyType: string,
    public readonly valueType: string,
    private readonly context?: EvaluationContext
  ) {
    assertComparableType(keyType, context);
  }

  public get(key: RuntimeValue): RuntimeValue {
    const entry = this.entries.get(runtimeMapKeyId(key));
    return entry ? entry.value : zeroValueForMapValue(this.valueType);
  }

  public getWithPresence(key: RuntimeValue): [RuntimeValue, boolean] {
    const entry = this.entries.get(runtimeMapKeyId(key));
    return entry ? [entry.value, true] : [zeroValueForMapValue(this.valueType), false];
  }

  public set(key: RuntimeValue, value: RuntimeValue): void {
    key = prepareAssignableToType(key, this.keyType, "map key", this.context);
    assertComparableValue(key, "map key");
    value = prepareAssignableToType(value, this.valueType, "map value", this.context);
    this.entries.set(runtimeMapKeyId(key), { key, value });
  }

  public delete(key: RuntimeValue): void {
    this.entries.delete(runtimeMapKeyId(key));
  }

  public clear(): void {
    this.entries.clear();
  }

  public orderedEntries(): Array<[RuntimeValue, RuntimeValue]> {
    return [...this.entries.values()].map((entry) => [entry.key, entry.value]);
  }

  public size(): number {
    return this.entries.size;
  }
}

export class RuntimeChannel {
  private readonly channel: AsyncGoChannel<RuntimeValue>;

  public constructor(
    public readonly elementType: string,
    public readonly capacity: number,
    private readonly context?: EvaluationContext
  ) {
    this.channel = new AsyncGoChannel(
      context?.scheduler() ?? new AsyncGoScheduler(),
      capacity,
      () => defaultValueForTypeText(elementType, context)
    );
  }

  public send(value: RuntimeValue): void {
    try {
      if (!this.channel.trySend(prepareAssignableToType(value, this.elementType, "channel send", this.context))) {
        throw new GoJuniorDeadlockError("send on channel would block");
      }
    } catch (error) {
      throw normalizeAsyncRuntimeError(error);
    }
  }

  public async sendAsync(value: RuntimeValue): Promise<void> {
    try {
      await this.channel.send(prepareAssignableToType(value, this.elementType, "channel send", this.context));
    } catch (error) {
      if (isDeadlockError(error)) throw new GoJuniorDeadlockError("send on channel would block");
      throw normalizeAsyncRuntimeError(error);
    }
  }

  public receive(): [RuntimeValue, boolean] {
    try {
      const ready = this.channel.tryReceive();
      if (!ready) throw new GoJuniorDeadlockError("receive from channel would block");
      return ready;
    } catch (error) {
      throw normalizeAsyncRuntimeError(error);
    }
  }

  public async receiveAsync(): Promise<[RuntimeValue, boolean]> {
    try {
      return await this.channel.receive();
    } catch (error) {
      if (isDeadlockError(error)) throw new GoJuniorDeadlockError("receive from channel would block");
      throw normalizeAsyncRuntimeError(error);
    }
  }

  public canSend(): boolean {
    return this.channel.canSendNow();
  }

  public canReceive(): boolean {
    return this.channel.canReceiveNow();
  }

  public close(): void {
    try {
      this.channel.close();
    } catch (error) {
      throw normalizeAsyncRuntimeError(error);
    }
  }

  public len(): number {
    return this.channel.len();
  }

  public cap(): number {
    return this.capacity;
  }

  public asyncChannel(): AsyncGoChannel<RuntimeValue> {
    return this.channel;
  }
}

export async function evaluateSource(source: string, options: EvaluationOptions = {}): Promise<EvaluationResult> {
  return evaluateSourceFiles([sourceFileFromSource(source, options)], options);
}

export async function evaluateSourceFiles(files: SourceFile[], options: EvaluationOptions = {}): Promise<EvaluationResult> {
  const parsed = frontSourceFilesToAst(files);
  const ast = parsed.ast;
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

export async function testSource(source: string, options: EvaluationOptions = {}): Promise<EvaluationResult> {
  return testSourceFiles([sourceFileFromSource(source, options)], options);
}

export async function testSourceFiles(files: SourceFile[], options: EvaluationOptions = {}): Promise<EvaluationResult> {
  const parsed = frontSourceFilesToAst(files);
  const ast = parsed.ast;
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

  const checked = checkFrontSourceFiles(files, typeCheckConfig(options));
  if (checked.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return {
      diagnostics: checked.diagnostics,
      output: [],
      ast
    };
  }

  return testProgram(ast, checked.diagnostics, options);
}

function sourceFileFromSource(source: string, options: EvaluationOptions = {}): SourceFile {
  return {
    filename: options.filename ?? REPL_FILENAME,
    source
  };
}

export async function evaluateProgram(ast: ProgramAst, options: EvaluationOptions = {}): Promise<EvaluationResult> {
  const context = new EvaluationContext(options);
  try {
    return await context.scheduler().runRoot(async () => {
      installSheets(context, options);
      installImports(context, ast);

      for (const declaration of ast.functions) {
        installFunctionDeclaration(context, declaration);
      }

      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      const declarationCompletion = await executeTopLevelStatements(declarations, context);
      expectNormalCompletion(declarationCompletion, "top-level declarations");

      if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
        const value = installedFunctionValue(context, ast.functions[0]);
        return {
          diagnostics: ast.diagnostics,
          output: context.output,
          ast,
          value
        };
      }

      await runInitFunctions(ast.functions, context);
      const completion = await executeTopLevelStatements(statements, context);
      return resultFromCompletion(ast, context.output, completion);
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      diagnostics: [
        ...ast.diagnostics,
        runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
      ],
      output: context.output,
      ast
    };
  }
}

async function testProgram(ast: ProgramAst, baseDiagnostics: Diagnostic[], options: EvaluationOptions): Promise<EvaluationResult> {
  const context = new EvaluationContext(options);
  const diagnostics = [...baseDiagnostics];
  try {
    return await context.scheduler().runRoot(async () => {
      installSheets(context, options);
      installImports(context, ast);

      for (const declaration of ast.functions) {
        installFunctionDeclaration(context, declaration);
      }

      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      const declarationCompletion = await executeTopLevelStatements(declarations, context);
      expectNormalCompletion(declarationCompletion, "top-level declarations");

      await runInitFunctions(ast.functions, context);
      const topLevelCompletion = await executeTopLevelStatements(statements, context);
      expectNormalCompletion(topLevelCompletion, "top-level statements");

      const tests = ast.functions.filter(isTestFunctionDecl);
      if (tests.length === 0) {
        context.write("testing: warning: no tests to run\nPASS\n");
        return { diagnostics, output: context.output, ast };
      }

      let failed = false;
      for (const declaration of tests) {
        const testFailed = await runOneTest(declaration, context, diagnostics);
        failed ||= testFailed;
      }
      context.write(failed ? "FAIL\n" : "PASS\n");
      return { diagnostics, output: context.output, ast };
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      diagnostics: [
        ...diagnostics,
        runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
      ],
      output: context.output,
      ast
    };
  }
}

function isTestFunctionDecl(declaration: FunctionDecl): boolean {
  return !declaration.receiver && /^Test($|[^a-z])/.test(declaration.name);
}

async function runOneTest(declaration: FunctionDecl, context: EvaluationContext, diagnostics: Diagnostic[]): Promise<boolean> {
  context.write(`=== RUN   ${declaration.name}\n`);
  const signatureError = testSignatureError(declaration);
  if (signatureError) {
    context.write(`    ${signatureError}\n`);
    context.write(`--- FAIL: ${declaration.name}\n`);
    diagnostics.push(testDiagnostic(declaration, `${declaration.name}: ${signatureError}`));
    return true;
  }

  const testingT = declaration.signature.parameters.length === 0 ? undefined : makeTestingT(declaration.name);
  let runtimeFailure: string | undefined;
  try {
    const callee = context.lookup(declaration.name);
    await callRuntime(callee, testingT ? [testingT.value] : [], context);
  } catch (error) {
    if (error instanceof GoJuniorTestStop) {
      // The testing.T state already records whether this was FailNow or SkipNow.
    } else {
      runtimeFailure = error instanceof Error ? error.message : String(error);
      if (testingT) {
        testingT.state.failed = true;
        testingT.state.logs.push(runtimeFailure);
      }
    }
  }

  if (testingT) emitTestingLogs(context, testingT.state);
  if (runtimeFailure && !testingT) {
    context.write(indentTestingLog(runtimeFailure));
  }

  const state = testingT?.state;
  if (runtimeFailure || state?.failed) {
    context.write(`--- FAIL: ${declaration.name}\n`);
    diagnostics.push(testDiagnostic(declaration, runtimeFailure ? `${declaration.name}: ${runtimeFailure}` : `${declaration.name} failed`));
    return true;
  }
  if (state?.skipped) {
    context.write(`--- SKIP: ${declaration.name}\n`);
    return false;
  }
  context.write(`--- PASS: ${declaration.name}\n`);
  return false;
}

function testSignatureError(declaration: FunctionDecl): string | undefined {
  if (declaration.signature.results.length !== 0) return "test function must not return values";
  const parameters = declaration.signature.parameters;
  if (parameters.length === 0) return undefined;
  if (parameters.length === 1 && !parameters[0]?.variadic && normalizeTypeText(parameters[0]?.type.text ?? "") === "*testing.T") {
    return undefined;
  }
  return "test function must have signature func TestX() or func TestX(t *testing.T)";
}

function testDiagnostic(declaration: FunctionDecl, message: string): Diagnostic {
  return {
    filename: declaration.span?.filename ?? REPL_FILENAME,
    code: "GOJR_TEST001",
    severity: "error",
    message,
    ...(declaration.span ? { span: declaration.span } : {})
  };
}

function runtimeDiagnostic(ast: ProgramAst, code: string, message: string): Diagnostic {
  const span = firstProgramSpan(ast);
  return {
    filename: span?.filename ?? REPL_FILENAME,
    code,
    severity: "error",
    message,
    ...(span ? { span } : {})
  };
}

function runtimeDiagnosticCode(error: unknown): string {
  if (error instanceof GoJuniorPanic) return "GOJR_PANIC001";
  if (error instanceof AsyncGoPanic) return "GOJR_PANIC001";
  if (error instanceof GoJuniorDeadlockError || error instanceof AsyncGoDeadlockError) return "GOJR_DEADLOCK001";
  return "GOJR_RUNTIME001";
}

function normalizeAsyncRuntimeError(error: unknown): unknown {
  if (error instanceof AsyncGoPanic) return new GoJuniorPanic(error.value as RuntimeValue);
  if (error instanceof AsyncGoDeadlockError) return new GoJuniorDeadlockError(error.message);
  return error;
}

function isDeadlockError(error: unknown): boolean {
  return error instanceof GoJuniorDeadlockError || error instanceof AsyncGoDeadlockError;
}

function firstProgramSpan(ast: ProgramAst): SourceSpan | undefined {
  return ast.body[0]?.span ?? ast.functions[0]?.span;
}

function makeTestingT(name: string): { value: RuntimePointer; state: TestingTState } {
  const state: TestingTState = {
    name,
    failed: false,
    skipped: false,
    logs: []
  };
  const object: RuntimeObject = {
    Fail: hostCallable("testing.(*T).Fail", () => {
      state.failed = true;
      return null;
    }),
    FailNow: hostCallable("testing.(*T).FailNow", () => {
      state.failed = true;
      throw new GoJuniorTestStop("failNow");
    }),
    Failed: hostCallable("testing.(*T).Failed", () => state.failed),
    Fatal: hostCallable("testing.(*T).Fatal", (args) => {
      appendTestingLog(state, testingSprint(args));
      state.failed = true;
      throw new GoJuniorTestStop("failNow");
    }),
    Fatalf: hostCallable("testing.(*T).Fatalf", (args) => {
      appendTestingLog(state, testingSprintf(args));
      state.failed = true;
      throw new GoJuniorTestStop("failNow");
    }),
    Error: hostCallable("testing.(*T).Error", (args) => {
      appendTestingLog(state, testingSprint(args));
      state.failed = true;
      return null;
    }),
    Errorf: hostCallable("testing.(*T).Errorf", (args) => {
      appendTestingLog(state, testingSprintf(args));
      state.failed = true;
      return null;
    }),
    Log: hostCallable("testing.(*T).Log", (args) => {
      appendTestingLog(state, testingSprint(args));
      return null;
    }),
    Logf: hostCallable("testing.(*T).Logf", (args) => {
      appendTestingLog(state, testingSprintf(args));
      return null;
    }),
    Name: hostCallable("testing.(*T).Name", () => state.name),
    Helper: hostCallable("testing.(*T).Helper", () => null),
    Skip: hostCallable("testing.(*T).Skip", (args) => {
      appendTestingLog(state, testingSprint(args));
      state.skipped = true;
      throw new GoJuniorTestStop("skipNow");
    }),
    Skipf: hostCallable("testing.(*T).Skipf", (args) => {
      appendTestingLog(state, testingSprintf(args));
      state.skipped = true;
      throw new GoJuniorTestStop("skipNow");
    }),
    SkipNow: hostCallable("testing.(*T).SkipNow", () => {
      state.skipped = true;
      throw new GoJuniorTestStop("skipNow");
    }),
    Skipped: hostCallable("testing.(*T).Skipped", () => state.skipped)
  };
  return {
    value: new RuntimePointer("testing.T", () => object, () => {
      throw new GoJuniorRuntimeError("cannot assign to testing.T");
    }),
    state
  };
}

function testingSprint(args: RuntimeValue[]): string {
  return args.map(formatValue).join(" ");
}

function testingSprintf(args: RuntimeValue[]): string {
  return sprintf(toStringValue(args[0] ?? ""), args.slice(1));
}

function appendTestingLog(state: TestingTState, text: string): void {
  state.logs.push(text.endsWith("\n") ? text : `${text}\n`);
}

function emitTestingLogs(context: EvaluationContext, state: TestingTState): void {
  for (const log of state.logs) {
    context.write(indentTestingLog(log));
  }
}

function indentTestingLog(text: string): string {
  const lineText = text.endsWith("\n") ? text.slice(0, -1) : text;
  if (lineText === "") return "    \n";
  return lineText.split("\n").map((line) => `    ${line}\n`).join("");
}

export class GoJuniorSession {
  private readonly context: EvaluationContext;
  private readonly acceptedSources: SourceFile[] = [];

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

  public async evaluate(source: string): Promise<EvaluationResult> {
    const sourceFile = sourceFileFromSource(source, { ...this.options, filename: this.options.filename ?? REPL_FILENAME });
    const parsed = frontSourceToAst(sourceFile.source, sourceFile.filename);
    const ast = parsed.ast;
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

    const typeDiagnostics = this.checkSource(sourceFile);
    if (typeDiagnostics.some((diagnostic) => diagnostic.severity === "error")) {
      return {
        diagnostics: typeDiagnostics,
        output: [],
        ast
      };
    }

    const outputStart = this.context.output.length;
    try {
      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      const declarationCompletion = await this.context.scheduler().runRoot(async () => {
        installImports(this.context, ast);

        for (const declaration of ast.functions) {
          installFunctionDeclaration(this.context, declaration);
        }

        return executeTopLevelStatements(declarations, this.context);
      });
      expectNormalCompletion(declarationCompletion, "top-level declarations");

      if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
        const value = installedFunctionValue(this.context, ast.functions[0]);
        this.acceptedSources.push(ensureTrailingNewlineSourceFile(sourceFile));
        return {
          diagnostics: ast.diagnostics,
          output: this.context.outputFrom(outputStart),
          ast,
          value
        };
      }

      const completion = await this.context.scheduler().runRoot(async () => {
        await runInitFunctions(ast.functions, this.context);
        return executeTopLevelStatements(statements, this.context);
      });
      const result = resultFromCompletion(ast, this.context.outputFrom(outputStart), completion);
      this.acceptedSources.push(ensureTrailingNewlineSourceFile(sourceFile));
      return result;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        diagnostics: [
          ...ast.diagnostics,
          runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
        ],
        output: this.context.outputFrom(outputStart),
        ast
      };
    }
  }

  private checkSource(source: SourceFile): Diagnostic[] {
    const checked = checkFrontSourceFiles([...this.acceptedSources, ensureTrailingNewlineSourceFile(source)], typeCheckConfig(this.options));
    return checked.diagnostics;
  }
}

function ensureTrailingNewline(source: string): string {
  return source.endsWith("\n") ? source : `${source}\n`;
}

function ensureTrailingNewlineSourceFile(file: SourceFile): SourceFile {
  return {
    filename: file.filename,
    source: ensureTrailingNewline(file.source)
  };
}

function typeCheckConfig(options: EvaluationOptions): CheckConfig {
  const universe = newUniverse();
  return {
    universe,
    sheetNamespaces: sheetNamespacesForOptions(options, universe)
  };
}

function sheetNamespacesForOptions(options: EvaluationOptions, universe: Universe): Record<string, SheetNamespace> {
  const namespaces: Record<string, SheetNamespace> = {
    sheet: sheetNamespaceForData(options.sheet ?? options.sheets?.[options.currentSheetName ?? "sheet"] ?? {}, universe)
  };
  for (const [name, data] of Object.entries(options.sheets ?? {})) {
    namespaces[name] = sheetNamespaceForData(data, universe);
  }
  return namespaces;
}

function sheetNamespaceForData(data: SheetData, universe: Universe): SheetNamespace {
  const cells: Record<string, CheckerType> = {};
  for (const [cell, value] of Object.entries(data)) {
    cells[normalizeCell(cell)] = checkerTypeForRuntimeValue(value, universe);
  }
  return {
    cells,
    defaultType: universe.basic.any
  };
}

function checkerTypeForRuntimeValue(value: RuntimeValue, universe: Universe): CheckerType {
  const actual = unwrapNamed(value);
  if (typeof actual === "bigint") return universe.basic.int64;
  if (typeof actual === "number") return universe.basic.float64;
  if (typeof actual === "string") return universe.basic.string;
  if (typeof actual === "boolean") return universe.basic.bool;
  if (isComplexValue(actual)) return universe.basic.complex128;
  if (Array.isArray(actual)) return new CheckerSliceType(universe.basic.any);
  if (actual instanceof RuntimeMap) {
    return new CheckerMapType(checkerTypeFromText(actual.keyType, universe), checkerTypeFromText(actual.valueType, universe));
  }
  return universe.basic.any;
}

function checkerTypeFromText(typeText: string, universe: Universe): CheckerType {
  const type = normalizeTypeText(typeText);
  switch (type) {
    case "bool": return universe.basic.bool;
    case "string": return universe.basic.string;
    case "float32":
    case "float64": return universe.basic.float64;
    case "complex64":
    case "complex128": return universe.basic.complex128;
    case "int":
    case "int8":
    case "int16":
    case "int32":
    case "int64":
    case "uint":
    case "uint8":
    case "uint16":
    case "uint32":
    case "uint64":
    case "uintptr":
    case "byte":
    case "rune":
      return universe.basic.int64;
    default:
      if (type.startsWith("[]")) return new CheckerSliceType(checkerTypeFromText(type.slice(2), universe));
      return new BasicType(BasicKind.Any);
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
  return errors.length > 0 && errors.every((diagnostic) => {
    if (diagnostic.code === "GOJR_PARSE001") {
      return /found\s+-->\s*''\s*<--/.test(diagnostic.message) || /but found:\s*''/.test(diagnostic.message);
    }
    if (diagnostic.code === "GOJR_SCAN001") {
      return /unterminated .*string literal/i.test(diagnostic.message);
    }
    if (diagnostic.code !== "GOJR_PARSE_FRONT001") return false;
    return diagnostic.span?.length === 0 &&
      (/expected/i.test(diagnostic.message) || /found EOF/i.test(diagnostic.message));
  });
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
  context.declareRoot("len", hostCallable("len", (args) => {
    return BigInt(valueLength(args[0] ?? null));
  }));
  context.declareRoot("cap", hostCallable("cap", (args) => {
    return BigInt(valueCapacity(args[0] ?? null));
  }));
  context.declareRoot("append", hostCallable("append", (args) => {
    return appendValues(args[0] ?? null, args.slice(1));
  }));
  context.declareRoot("new", hostCallable("new", () => {
    throw new GoJuniorRuntimeError("new expects a type argument");
  }));
  context.declareRoot("delete", hostCallable("delete", (args) => {
    const target = args[0] ?? null;
    if (!(target instanceof RuntimeMap)) throw new GoJuniorRuntimeError("delete expects a map");
    target.delete(args[1] ?? null);
    return null;
  }));
  context.declareRoot("close", hostCallable("close", (args) => {
    const target = args[0] ?? null;
    if (target === null) throw new GoJuniorPanic("close of nil channel");
    if (!(target instanceof RuntimeChannel)) throw new GoJuniorRuntimeError("close expects a channel");
    target.close();
    return null;
  }));
  context.declareRoot("clear", hostCallable("clear", (args) => {
    clearValue(args[0] ?? null);
    return null;
  }));
  context.declareRoot("copy", hostCallable("copy", (args) => {
    return BigInt(copyValues(args[0] ?? null, args[1] ?? null));
  }));
  context.declareRoot("min", hostCallable("min", (args) => minMaxValues(args, "min")));
  context.declareRoot("max", hostCallable("max", (args) => minMaxValues(args, "max")));
  context.declareRoot("complex", hostCallable("complex", (args) => complexValue(toFloat(unwrapNamed(args[0] ?? 0)), toFloat(unwrapNamed(args[1] ?? 0)))));
  context.declareRoot("real", hostCallable("real", (args) => {
    const value = unwrapNamed(args[0] ?? 0);
    return isComplexValue(value) ? value.real : toFloat(value);
  }));
  context.declareRoot("imag", hostCallable("imag", (args) => {
    const value = unwrapNamed(args[0] ?? 0);
    return isComplexValue(value) ? value.imag : 0;
  }));
  context.declareRoot("print", hostCallable("print", (args, context) => {
    for (const arg of args) context.write(formatValue(arg));
    return null;
  }));
  context.declareRoot("println", hostCallable("println", (args, context) => {
    context.write(`${args.map(formatValue).join(" ")}\n`);
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
    if (name === "_") continue;
    if (name === ".") {
      for (const [exportName, value] of Object.entries(pkg)) {
        context.declareOrAssignRoot(exportName, value, true);
      }
      continue;
    }
    context.declareOrAssignRoot(name, pkg, true);
  }
}

function installFunctionDeclaration(context: EvaluationContext, declaration: FunctionDecl): void {
  if (declaration.receiver) {
    context.registerMethod(declaration);
    return;
  }
  context.declareOrAssignRoot(declaration.name, functionValue(declaration), true);
}

function installedFunctionValue(context: EvaluationContext, declaration: FunctionDecl): RuntimeValue {
  return declaration.receiver ? functionValue(declaration) : context.lookup(declaration.name);
}

function bindingTypeText(type: TypeNode | string | undefined): string | undefined {
  const text = typeof type === "string" ? type : type?.text;
  const normalized = text ? normalizeTypeText(text) : "";
  return normalized && normalized !== "<missing>" ? normalized : undefined;
}

function installAutomaticImports(context: EvaluationContext): void {
  context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}

function availablePackages(context: EvaluationContext): Record<string, RuntimeObject> {
  return {
    fmt: fmtPackage(),
    testing: testingPackage(),
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

function testingPackage(): RuntimeObject {
  return {
    Short: hostCallable("testing.Short", () => false),
    Verbose: hostCallable("testing.Verbose", () => false)
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

function splitTopLevelDeclarations(statements: Statement[]): { declarations: Statement[]; statements: Statement[] } {
  const declarations: Statement[] = [];
  const executable: Statement[] = [];
  for (const statement of statements) {
    if (statement.kind === "ConstDecl" || statement.kind === "VarDecl" || statement.kind === "TypeDecl") {
      declarations.push(statement);
    } else {
      executable.push(statement);
    }
  }
  return { declarations, statements: executable };
}

async function runInitFunctions(functions: FunctionDecl[], context: EvaluationContext): Promise<void> {
  for (const declaration of functions) {
    if (declaration.name !== "init" || declaration.receiver) continue;
    await callRuntime(functionValue(declaration), [], context);
  }
}

async function executeTopLevelStatements(statements: Statement[], context: EvaluationContext): Promise<Completion> {
  try {
    return await executeStatements(statements, context);
  } finally {
    await context.runDefersAsync();
  }
}

function goJuniorFunctionValue(
  name: string,
  signature: FunctionDecl["signature"],
  body: FunctionDecl["body"],
  closureScope?: Scope,
  declaration?: FunctionDecl,
  boundReceiver?: RuntimeValue
): GoJuniorFunction {
  return {
    kind: "GoJuniorFunction",
    name,
    ...(declaration ? { declaration } : {}),
    async call(args, parentContext) {
      const context = parentContext;
      const invoke = () => context.childScopeAsync(() => context.deferScopeAsync(async () => {
        if (declaration?.receiver?.name) {
          context.declare(
            declaration.receiver.name,
            prepareReceiverBinding(declaration.receiver.type.text, boundReceiver),
            true,
            declaration.receiver.type
          );
        }
        for (const [index, parameter] of signature.parameters.entries()) {
          if (parameter.name) {
            const value = parameter.variadic ? args.slice(index) : args[index] ?? null;
            context.declare(parameter.name, value, true, parameter.variadic ? `[]${parameter.type.text}` : parameter.type);
            if (parameter.variadic) break;
          }
        }
        for (const result of signature.results) {
          if (result.name) {
            context.declare(result.name, defaultValueForDeclarationType(result.type, context), true, result.type);
          }
        }
        const completion = await executeBlock(body, context, false);
        await context.runDefersAsync();
        if (completion.kind === "return") {
          const values = completion.values.length === 0
            ? namedReturnValues(signature, context)
            : completion.values;
          return values.length === 1 ? values[0] ?? null : values;
        }
        return null;
      }));
      return closureScope ? context.withScopeAsync(closureScope, invoke) : invoke();
    }
  };
}

function namedReturnValues(signature: FunctionDecl["signature"], context: EvaluationContext): RuntimeValue[] {
  const namedResults = signature.results.filter((result) => result.name);
  if (namedResults.length === 0) return [];
  return namedResults.map((result) => result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type, context));
}

function normalizeReceiverType(typeText: string): { baseType: string; pointer: boolean } {
  const type = normalizeTypeText(typeText);
  if (type.startsWith("*")) return { baseType: type.slice(1), pointer: true };
  return { baseType: type, pointer: false };
}

function methodKey(typeName: string, methodName: string): string {
  return `${typeName}.${methodName}`;
}

function boundMethodValue(method: MethodDef, receiver: RuntimeValue): GoJuniorFunction {
  const boundReceiver = method.pointerReceiver
    ? pointerToReceiver(receiver, method.receiverType)
    : valueReceiver(receiver);
  return goJuniorFunctionValue(
    method.declaration.name,
    method.declaration.signature,
    method.declaration.body,
    undefined,
    method.declaration,
    boundReceiver
  );
}

function valueReceiver(receiver: RuntimeValue): RuntimeValue {
  const value = dereferenceIfPointer(receiver);
  return value instanceof RuntimeStruct ? value.clone() : value;
}

function prepareReceiverBinding(receiverTypeText: string, receiver: RuntimeValue | undefined): RuntimeValue {
  if (receiver === undefined) return null;
  const receiverType = normalizeReceiverType(receiverTypeText);
  if (receiverType.pointer) return pointerToReceiver(receiver, receiverType.baseType);
  return valueReceiver(receiver);
}

function pointerToReceiver(receiver: RuntimeValue, typeName: string): RuntimePointer {
  if (receiver instanceof RuntimePointer) return receiver;
  if (receiver instanceof RuntimeStruct && receiver.typeName === typeName) {
    return new RuntimePointer(typeName, () => receiver, (value) => {
      if (!(value instanceof RuntimeStruct) || value.typeName !== typeName) {
        throw new GoJuniorRuntimeError(`${formatValue(value)} is not assignable to *${typeName}`);
      }
      receiver.fields.clear();
      for (const [field, item] of value.orderedFields()) receiver.set(field, item);
    });
  }
  throw new GoJuniorRuntimeError(`${formatValue(receiver)} is not addressable as *${typeName}`);
}

async function executeStatements(statements: Statement[], context: EvaluationContext): Promise<Completion> {
  const labels = statementLabels(statements);
  let lastValue: RuntimeValue | undefined;
  for (let pc = 0; pc < statements.length; pc += 1) {
    const statement = statements[pc];
    if (!statement) continue;
    const completion = await executeStatement(statement, context);
    if (completion.kind === "goto") {
      const target = labels.get(completion.label);
      if (target !== undefined) {
        validateGotoTarget(statements, pc, target, completion.label);
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

function validateGotoTarget(statements: Statement[], sourceIndex: number, targetIndex: number, label: string): void {
  if (targetIndex <= sourceIndex) return;
  for (let index = sourceIndex + 1; index < targetIndex; index += 1) {
    if (statementDeclaresVariables(statements[index])) {
      throw new GoJuniorRuntimeError(`goto ${label} jumps over variable declaration`);
    }
  }
}

function statementDeclaresVariables(statement: Statement | undefined): boolean {
  if (!statement) return false;
  if (statement.kind === "VarDecl" || statement.kind === "ShortVarStatement") return true;
  if (statement.kind === "LabeledStatement") return statementDeclaresVariables(statement.statement);
  return false;
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

async function executeBlock(block: BlockStatement, context: EvaluationContext, createScope = true): Promise<Completion> {
  if (!createScope) return executeStatements(block.statements, context);
  return context.childScopeAsync(() => executeStatements(block.statements, context));
}

async function executeStatement(statement: Statement, context: EvaluationContext): Promise<Completion> {
  switch (statement.kind) {
    case "BlockStatement":
      return executeBlock(statement, context);

    case "LabeledStatement":
      return executeLabeledStatement(statement, context);

    case "ConstDecl":
      let previousConstValue = undefined as Expression | undefined;
      let previousConstType = undefined as TypeNode | undefined;
      for (const [index, declaration] of statement.declarations.entries()) {
        const valueExpression = declaration.value ?? previousConstValue;
        const type = declaration.type ?? previousConstType;
        const value = valueExpression
          ? await evaluateConstExpression(valueExpression, BigInt(index), context)
          : defaultValueForDeclarationType(type, context);
        context.declare(
          declaration.name,
          value,
          false,
          type
        );
        if (declaration.value) previousConstValue = declaration.value;
        if (declaration.type) previousConstType = declaration.type;
      }
      return { kind: "normal" };

    case "VarDecl":
      for (const declaration of statement.declarations) {
        context.declare(
          declaration.name,
          declaration.value ? await evaluateExpression(declaration.value, context) : defaultValueForDeclarationType(declaration.type, context),
          true,
          declaration.type
        );
      }
      return { kind: "normal" };

    case "TypeDecl":
      for (const declaration of statement.declarations) {
        context.registerType(declaration);
      }
      return { kind: "normal" };

    case "ReturnStatement":
      return {
        kind: "return",
        values: await evaluateExpressionList(statement.values, context)
      };

    case "IfStatement":
      return executeIf(statement, context);

    case "SwitchStatement":
      return executeSwitch(statement, context);

    case "SelectStatement":
      return executeSelect(statement, context);

    case "ForStatement":
      return executeFor(statement, context);

    case "DeferStatement":
      await executeDefer(statement, context);
      return { kind: "normal" };

    case "GoStatement":
      await executeGo(statement, context);
      return { kind: "normal" };

    case "SendStatement":
      await executeSend(statement, context);
      return { kind: "normal" };

    case "BranchStatement":
      return branchCompletion(statement);

    case "AssignStatement":
      await executeAssign(statement, context);
      return { kind: "normal" };

    case "ShortVarStatement":
      await executeShortVar(statement, context);
      return { kind: "normal" };

    case "IncDecStatement":
      await executeIncDec(statement, context);
      return { kind: "normal" };

    case "ExpressionStatement":
      return { kind: "normal", value: await evaluateExpression(statement.expression, context) };
  }
}

async function executeLabeledStatement(statement: Extract<Statement, { kind: "LabeledStatement" }>, context: EvaluationContext): Promise<Completion> {
  if (!statement.statement) return { kind: "normal" };
  if (statement.statement.kind === "ForStatement") {
    return executeFor(statement.statement, context, statement.label);
  }
  if (statement.statement.kind === "SwitchStatement") {
    return executeSwitch(statement.statement, context, statement.label);
  }
  return executeStatement(statement.statement, context);
}

async function evaluateConstExpression(expression: Expression, iotaValue: bigint, context: EvaluationContext): Promise<RuntimeValue> {
  return context.childScopeAsync(async () => {
    context.declare("iota", iotaValue, false, "int64");
    return evaluateExpression(expression, context);
  });
}

async function executeIf(statement: IfStatement, context: EvaluationContext): Promise<Completion> {
  return context.childScopeAsync(async () => {
    if (statement.init) {
      expectNormalCompletion(await executeStatement(statement.init, context), "if init statement");
    }
    if (toBool(await evaluateExpression(statement.condition, context))) {
      return executeBlock(statement.thenBlock, context);
    }
    if (!statement.elseBranch) return { kind: "normal" };
    return statement.elseBranch.kind === "IfStatement"
      ? executeIf(statement.elseBranch, context)
      : executeBlock(statement.elseBranch, context);
  });
}

async function executeSwitch(statement: SwitchStatement, context: EvaluationContext, label?: string): Promise<Completion> {
  return context.childScopeAsync(async () => {
    if (statement.init) {
      expectNormalCompletion(await executeStatement(statement.init, context), "switch init statement");
    }
    if (statement.typeSwitch) return executeTypeSwitch(statement, context, label);
    validateValueSwitchFallthrough(statement);
    return executeValueSwitch(statement, context, label);
  });
}

async function executeValueSwitch(statement: SwitchStatement, context: EvaluationContext, label?: string): Promise<Completion> {
  const switchValue = statement.expression ? await evaluateExpression(statement.expression, context) : true;
  let matched = false;

  for (const clause of statement.clauses) {
    if (!matched) {
      matched = clause.default;
      for (const value of clause.values) {
        if (matched) break;
        matched = valueEqual(switchValue, await evaluateExpression(value, context));
      }
    }
    if (!matched) continue;

    const completion = await executeStatements(clause.statements, context);
    if (completion.kind === "fallthrough") {
      matched = true;
      continue;
    }
    if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
    return completion;
  }

  return { kind: "normal" };
}

function validateValueSwitchFallthrough(statement: SwitchStatement): void {
  for (const [index, clause] of statement.clauses.entries()) {
    const fallthroughIndex = clause.statements.findIndex((item) =>
      item.kind === "BranchStatement" && item.branch === "fallthrough"
    );
    if (fallthroughIndex < 0) continue;
    if (index === statement.clauses.length - 1) {
      throw new GoJuniorRuntimeError("fallthrough cannot appear in the final switch clause");
    }
    if (fallthroughIndex !== clause.statements.length - 1) {
      throw new GoJuniorRuntimeError("fallthrough must be the final statement in a switch clause");
    }
  }
}

async function executeTypeSwitch(statement: SwitchStatement, context: EvaluationContext, label?: string): Promise<Completion> {
  if (!statement.typeSwitch) return { kind: "normal" };
  const switchValue = await evaluateExpression(statement.typeSwitch.expression, context);

  for (const clause of statement.clauses) {
    const matched = clause.default || (clause.typeValues ?? []).some((type) => runtimeValueMatchesType(switchValue, type.text, context));
    if (!matched) continue;

    const completion = await context.childScopeAsync(async () => {
      if (statement.typeSwitch?.name) {
        const bindingValue = typeSwitchBindingValue(switchValue, clause, context);
        if (statement.typeSwitch.define) {
          context.declare(statement.typeSwitch.name, bindingValue, true);
        } else {
          context.assign(statement.typeSwitch.name, bindingValue);
        }
      }
      return executeStatements(clause.statements, context);
    });
    if (completion.kind === "fallthrough") {
      throw new GoJuniorRuntimeError("fallthrough is not allowed in type switches");
    }
    if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
    return completion;
  }

  return { kind: "normal" };
}

function typeSwitchBindingValue(switchValue: RuntimeValue, clause: SwitchStatement["clauses"][number], context: EvaluationContext): RuntimeValue {
  if (clause.default || (clause.typeValues?.length ?? 0) !== 1) return switchValue;
  const typeText = clause.typeValues?.[0]?.text;
  if (!typeText) return switchValue;
  if (normalizeTypeText(typeText) === "nil") return null;
  return assertedRuntimeValue(switchValue, typeText, context);
}

async function executeFor(statement: ForStatement, context: EvaluationContext, label?: string): Promise<Completion> {
  if (statement.range) {
    const source = await evaluateExpression(statement.range.source, context);
    const entries = await rangeEntries(source, context);
    for (const [index, value] of entries) {
      const completion = await context.childScopeAsync(async () => {
        if (statement.range?.keyName) {
          if (statement.range.define) context.declare(statement.range.keyName, index, true, inferredTypeText(index));
          else context.assign(statement.range.keyName, index);
        }
        if (statement.range?.valueName) {
          if (statement.range.define) context.declare(statement.range.valueName, value, true, inferredTypeText(value));
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

  return context.childScopeAsync(async () => {
    if (statement.init) {
      expectNormalCompletion(await executeStatement(statement.init, context), "for init statement");
    }
    if (statement.post?.kind === "ShortVarStatement") {
      throw new GoJuniorRuntimeError("short variable declaration is not allowed in a for post statement");
    }

    for (let iteration = 0; iteration < context.loopLimit(); iteration += 1) {
      if (statement.condition && !toBool(await evaluateExpression(statement.condition, context))) {
        return { kind: "normal" };
      }

      const completion = await executeBlock(statement.body, context);
      if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
      if (completion.kind === "continue" && labelMatches(completion.label, label)) {
        if (statement.post) {
          expectNormalCompletion(await executeStatement(statement.post, context), "for post statement");
        }
        continue;
      }
      if (completion.kind !== "normal") return completion;

      if (statement.post) {
        expectNormalCompletion(await executeStatement(statement.post, context), "for post statement");
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

async function executeDefer(statement: DeferStatement, context: EvaluationContext): Promise<void> {
  if (statement.expression.kind === "CallExpression") {
    const callee = await evaluateExpression(statement.expression.callee, context);
    const args = await evaluateExpressionList(statement.expression.args, context);
    context.pushDefer(() => callRuntime(callee, args, context));
    return;
  }

  context.pushDefer(() => evaluateExpression(statement.expression, context));
}

async function executeGo(statement: GoStatement, context: EvaluationContext): Promise<void> {
  const callee = await evaluateExpression(statement.call.callee, context);
  const args = await evaluateExpressionList(statement.call.args, context);
  const goroutineContext = context.fork();
  context.scheduler().go(async () => {
    await callRuntime(callee, statement.call.spreadLast ? spreadLastArgument(args) : args, goroutineContext);
  });
}

async function executeSend(statement: SendStatement, context: EvaluationContext): Promise<void> {
  const channel = await evaluateExpression(statement.channel, context);
  if (channel === null) throw new GoJuniorDeadlockError("send on nil channel would block");
  if (!(channel instanceof RuntimeChannel)) throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
  await channel.sendAsync(await evaluateExpression(statement.value, context));
}

async function executeSelect(statement: SelectStatement, context: EvaluationContext): Promise<Completion> {
  const prepared = await Promise.all(statement.clauses.map((clause) => prepareSelectClause(clause, context)));
  let selected: AsyncSelectResult<RuntimeValue>;
  try {
    selected = await asyncSelect(context.scheduler(), prepared.map((item) => item.selectCase));
  } catch (error) {
    if (isDeadlockError(error)) throw new GoJuniorDeadlockError("select would block");
    throw normalizeAsyncRuntimeError(error);
  }
  const preparedClause = prepared[selected.index];
  if (!preparedClause) return { kind: "normal" };
  await applySelectResult(preparedClause.clause.comm, selected, context);
  const completion = await executeStatements(preparedClause.clause.statements, context);
  if (completion.kind === "break" && completion.label === undefined) return { kind: "normal" };
  return completion;
}

interface PreparedSelectClause {
  clause: SelectStatement["clauses"][number];
  selectCase: AsyncSelectCase<RuntimeValue>;
}

async function prepareSelectClause(
  clause: SelectStatement["clauses"][number],
  context: EvaluationContext
): Promise<PreparedSelectClause> {
  const comm = clause.comm;
  if (clause.default || !comm) return { clause, selectCase: { op: "default" } };
  if (comm.kind === "SendStatement") {
    const channel = await evaluateExpression(comm.channel, context);
    const value = await evaluateExpression(comm.value, context);
    if (channel === null) return { clause, selectCase: { op: "receive", channel: undefined } };
    if (!(channel instanceof RuntimeChannel)) throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
    return { clause, selectCase: { op: "send", channel: channel.asyncChannel(), value } };
  }
  const receive = selectReceiveExpression(comm);
  if (receive) {
    const channel = await evaluateExpression(receive, context);
    if (channel === null) return { clause, selectCase: { op: "receive", channel: undefined } };
    if (!(channel instanceof RuntimeChannel)) throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
    return { clause, selectCase: { op: "receive", channel: channel.asyncChannel() } };
  }
  await executeStatement(comm, context).then((completion) => {
    expectNormalCompletion(completion, "select communication clause");
  });
  return { clause, selectCase: { op: "default" } };
}

async function applySelectResult(
  statement: Statement | undefined,
  result: AsyncSelectResult<RuntimeValue>,
  context: EvaluationContext
): Promise<void> {
  if (result.op !== "receive" || !statement) return;
  await assignSelectReceive(statement, [result.value, result.ok], context);
}

function selectReceiveExpression(statement: Statement): Expression | undefined {
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

async function assignSelectReceive(statement: Statement, received: [RuntimeValue, boolean], context: EvaluationContext): Promise<void> {
  if (statement.kind === "ExpressionStatement") return;
  if (statement.kind === "ShortVarStatement") {
    const values = selectReceiveValues(received, statement.names.length);
    declareOrAssignShortVars(statement.names, values, context);
    return;
  }
  if (statement.kind === "AssignStatement") {
    const values = selectReceiveValues(received, statement.targets.length);
    if (values.length !== statement.targets.length) {
      throw new GoJuniorRuntimeError(`assignment count mismatch: ${statement.targets.length} targets but ${values.length} values`);
    }
    for (const [index, target] of statement.targets.entries()) {
      await assignExpressionTarget(target, values[index] ?? null, context);
    }
  }
}

function selectReceiveValues(received: [RuntimeValue, boolean], targetCount: number): RuntimeValue[] {
  if (targetCount === 2) return [received[0], received[1]];
  return [received[0]];
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

async function executeAssign(statement: AssignStatement, context: EvaluationContext): Promise<void> {
  const values = await evaluateAssignmentValues(statement.values, statement.targets.length, context);
  if (values.length !== statement.targets.length) {
    throw new GoJuniorRuntimeError(`assignment count mismatch: ${statement.targets.length} targets but ${values.length} values`);
  }
  for (const [index, target] of statement.targets.entries()) {
    const value = values[index] ?? null;
    if (statement.operator && statement.operator !== "=") {
      const current = await evaluateExpression(target, context);
      await assignExpressionTarget(target, applyCompoundAssignment(statement.operator, current, value), context);
    } else {
      await assignExpressionTarget(target, value, context);
    }
  }
}

async function executeShortVar(statement: ShortVarStatement, context: EvaluationContext): Promise<void> {
  const values = await evaluateAssignmentValues(statement.values, statement.names.length, context);
  declareOrAssignShortVars(statement.names, values, context);
}

function declareOrAssignShortVars(names: string[], values: RuntimeValue[], context: EvaluationContext): void {
  if (values.length !== names.length) {
    throw new GoJuniorRuntimeError(`short declaration count mismatch: ${names.length} names but ${values.length} values`);
  }
  const hasNewName = names.some((name) => name !== "_" && name !== "<invalid>" && !context.hasLocal(name));
  if (!hasNewName) {
    throw new GoJuniorRuntimeError("short declaration has no new variables");
  }
  for (const [index, name] of names.entries()) {
    if (name === "_") continue;
    if (name === "<invalid>") throw new GoJuniorRuntimeError("non-identifier used in short declaration");
    const value = values[index] ?? null;
    if (context.hasLocal(name)) {
      context.assign(name, value);
    } else {
      context.declare(name, value, true, inferredTypeText(value));
    }
  }
}

async function evaluateAssignmentValues(expressions: Expression[], targetCount: number, context: EvaluationContext): Promise<RuntimeValue[]> {
  if (targetCount === 2 && expressions.length === 1 && expressions[0]?.kind === "IndexExpression") {
    const lookup = await evaluateMapLookupWithPresence(expressions[0], context);
    if (lookup) return lookup;
  }
  if (targetCount === 2 && expressions.length === 1 && expressions[0]?.kind === "TypeAssertionExpression") {
    return evaluateTypeAssertionWithPresence(expressions[0], context);
  }
  if (expressions.length === 1 && expressions[0]?.kind === "UnaryExpression" && expressions[0].operator === "<-") {
    const received = await receiveFromChannel(await evaluateExpression(expressions[0].operand, context));
    return targetCount === 2 ? [received[0], received[1]] : [received[0]];
  }
  const values = await evaluateExpressionList(expressions, context);
  if (targetCount > 1 && values.length === 1 && Array.isArray(values[0])) {
    return values[0];
  }
  return values;
}

function applyCompoundAssignment(operator: NonNullable<AssignStatement["operator"]>, left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  switch (operator) {
    case "+=": return addValues(left, right);
    case "-=": return subtractNumbers(left, right);
    case "*=": return multiplyNumbers(left, right);
    case "/=": return divideNumbers(left, right);
    case "%=": return moduloNumbers(left, right);
    case "&=": return bitwiseAnd(left, right);
    case "|=": return bitwiseOr(left, right);
    case "^=": return bitwiseXor(left, right);
    case "&^=": return bitClear(left, right);
    case "<<=": return shiftLeft(left, right);
    case ">>=": return shiftRight(left, right);
    default: return right;
  }
}

async function evaluateMapLookupWithPresence(expression: IndexExpression, context: EvaluationContext): Promise<RuntimeValue[] | undefined> {
  const object = await evaluateExpression(expression.object, context);
  if (!(object instanceof RuntimeMap)) return undefined;
  const index = await evaluateExpression(expression.index, context);
  const [value, ok] = object.getWithPresence(index);
  return [value, ok];
}

async function executeIncDec(statement: IncDecStatement, context: EvaluationContext): Promise<void> {
  const current = await evaluateExpression(statement.target, context);
  const next = statement.operator === "++"
    ? addNumbers(current, 1n)
    : subtractNumbers(current, 1n);
  await assignExpressionTarget(statement.target, next, context);
}

async function assignExpressionTarget(target: Expression, value: RuntimeValue, context: EvaluationContext): Promise<void> {
  if (target.kind === "Identifier") {
    if (target.name === "_") return;
    context.assign(target.name, value);
    return;
  }
  if (target.kind === "SelectorExpression") {
    await setSelector(target, value, context);
    return;
  }
  if (target.kind === "IndexExpression") {
    await setIndex(target, value, context);
    return;
  }
  if (target.kind === "UnaryExpression" && target.operator === "*") {
    const pointer = await evaluateExpression(target.operand, context);
    if (!(pointer instanceof RuntimePointer)) {
      throw new GoJuniorRuntimeError(`${formatValue(pointer)} is not a pointer`);
    }
    pointer.set(value);
    return;
  }
  throw new GoJuniorRuntimeError("unsupported assignment target");
}

async function evaluateExpression(expression: Expression, context: EvaluationContext): Promise<RuntimeValue> {
  switch (expression.kind) {
    case "Identifier":
      return evaluateIdentifier(expression, context);

    case "TypeExpression":
      throw new GoJuniorRuntimeError(`${expression.type.text} is a type, not a value`);

    case "Literal":
      return expression.value;

    case "FunctionLiteralExpression":
      return functionLiteralValue(expression, context);

    case "ArrayLiteralExpression":
      return evaluateArrayLiteral(expression, context);

    case "StructLiteralExpression":
      return evaluateStructLiteral(expression, context);

    case "MapLiteralExpression":
      return evaluateMapLiteral(expression, context);

    case "UnaryExpression":
      return evaluateUnary(expression, context);

    case "BinaryExpression":
      return evaluateBinary(expression, context);

    case "TypeAssertionExpression":
      return evaluateTypeAssertion(expression, context);

    case "SelectorExpression":
      return getSelector(expression, context);

    case "CallExpression":
      return evaluateCall(expression, context);

    case "IndexExpression":
      return getIndex(await evaluateExpression(expression.object, context), await evaluateExpression(expression.index, context));

    case "SliceExpression":
      return getSlice(
        await evaluateExpression(expression.object, context),
        expression.start ? await evaluateExpression(expression.start, context) : undefined,
        expression.end ? await evaluateExpression(expression.end, context) : undefined,
        expression.max ? await evaluateExpression(expression.max, context) : undefined
      );

    case "SpreadsheetRangeExpression":
      return getSpreadsheetRange(expression, context);
  }
}

async function evaluateExpressionList(expressions: Expression[], context: EvaluationContext): Promise<RuntimeValue[]> {
  const values: RuntimeValue[] = [];
  for (const expression of expressions) values.push(await evaluateExpression(expression, context));
  return values;
}

async function evaluateArrayLiteral(expression: ArrayLiteralExpression, context: EvaluationContext): Promise<RuntimeValue[]> {
  const type = parseArrayOrSliceTypeText(expression.type.text);
  if (!type) {
    throw new GoJuniorRuntimeError(`${expression.type.text} is not an array or slice literal type`);
  }
  const values: RuntimeValue[] = [];
  for (const element of expression.elements) {
    values.push(prepareAssignableToType(await evaluateExpression(element, context), type.elementType, "array element", context));
  }
  if (type.length !== undefined && values.length > type.length) {
    throw new GoJuniorRuntimeError(`array literal has ${values.length} elements but type ${expression.type.text} has length ${type.length}`);
  }
  if (type.length !== undefined && !type.inferLength) {
    while (values.length < type.length) {
      values.push(defaultValueForTypeText(type.elementType, context));
    }
  }
  return values;
}

function makeRuntimeSlice(elementType: string, length: number, capacity: number, context: EvaluationContext): RuntimeValue[] {
  const values = Array.from({ length }, () => defaultValueForTypeText(elementType, context));
  arrayCapacities.set(values, capacity);
  return values;
}

async function evaluateStructLiteral(expression: StructLiteralExpression, context: EvaluationContext): Promise<RuntimeStruct> {
  const typeDef = context.typeDef(expression.typeName);
  if (!typeDef) {
    throw new GoJuniorRuntimeError(`${expression.typeName} is not a declared struct type`);
  }

  const struct = new RuntimeStruct(expression.typeName);
  for (const field of typeDef.fields) {
    struct.set(field.name, defaultValueForTypeText(field.type.text, context));
  }
  for (const [index, field] of expression.fields.entries()) {
    const declared = field.name
      ? typeDef.fields.find((candidate) => candidate.name === field.name)
      : typeDef.fields[index];
    if (!declared) {
      throw new GoJuniorRuntimeError(field.name
        ? `${expression.typeName} has no field ${field.name}`
        : `${expression.typeName} literal has too many values`);
    }
    const value = await evaluateExpression(field.value, context);
    struct.set(declared.name, prepareAssignableToType(value, declared.type.text, `field ${field.name ?? declared.name}`, context));
  }
  return struct;
}

async function evaluateMapLiteral(expression: MapLiteralExpression, context: EvaluationContext): Promise<RuntimeMap> {
  const map = new RuntimeMap(expression.keyType.text, expression.valueType.text, context);
  for (const entry of expression.entries) {
    map.set(await evaluateExpression(entry.key, context), await evaluateExpression(entry.value, context));
  }
  return map;
}

function evaluateIdentifier(expression: IdentifierExpression, context: EvaluationContext): RuntimeValue {
  if (expression.name === "nil") return null;
  return context.lookup(expression.name);
}

async function evaluateUnary(expression: UnaryExpression, context: EvaluationContext): Promise<RuntimeValue> {
  switch (expression.operator) {
    case "+":
      return numericIdentity(await evaluateExpression(expression.operand, context));
    case "-":
      return negateNumber(await evaluateExpression(expression.operand, context));
    case "!":
      return !toBool(await evaluateExpression(expression.operand, context));
    case "^":
      return bitwiseComplement(await evaluateExpression(expression.operand, context));
    case "&":
      return pointerToExpression(expression.operand, context);
    case "*":
      return dereference(await evaluateExpression(expression.operand, context));
    case "<-":
      return (await receiveFromChannel(await evaluateExpression(expression.operand, context)))[0];
  }
}

async function receiveFromChannel(value: RuntimeValue): Promise<[RuntimeValue, boolean]> {
  if (value === null) throw new GoJuniorDeadlockError("receive from nil channel would block");
  if (!(value instanceof RuntimeChannel)) throw new GoJuniorRuntimeError(`${formatValue(value)} is not a channel`);
  return value.receiveAsync();
}

async function evaluateBinary(expression: BinaryExpression, context: EvaluationContext): Promise<RuntimeValue> {
  if (expression.operator === "&&") {
    return toBool(await evaluateExpression(expression.left, context)) && toBool(await evaluateExpression(expression.right, context));
  }
  if (expression.operator === "||") {
    return toBool(await evaluateExpression(expression.left, context)) || toBool(await evaluateExpression(expression.right, context));
  }

  const left = await evaluateExpression(expression.left, context);
  const right = await evaluateExpression(expression.right, context);

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
      return addValues(left, right);
    case "-":
      return subtractNumbers(left, right);
    case "*":
      return multiplyNumbers(left, right);
    case "/":
      return divideNumbers(left, right);
    case "%":
      return moduloNumbers(left, right);
    case "|":
      return bitwiseOr(left, right);
    case "^":
      return bitwiseXor(left, right);
    case "&":
      return bitwiseAnd(left, right);
    case "&^":
      return bitClear(left, right);
    case "<<":
      return shiftLeft(left, right);
    case ">>":
      return shiftRight(left, right);
  }
}

async function evaluateTypeAssertion(expression: TypeAssertionExpression, context: EvaluationContext): Promise<RuntimeValue> {
  const value = await evaluateExpression(expression.expression, context);
  if (!runtimeValueMatchesType(value, expression.type.text, context)) {
    throw new GoJuniorRuntimeError(`${formatValue(value)} does not have dynamic type ${normalizeTypeText(expression.type.text)}`);
  }
  return assertedRuntimeValue(value, expression.type.text, context);
}

async function evaluateTypeAssertionWithPresence(expression: TypeAssertionExpression, context: EvaluationContext): Promise<RuntimeValue[]> {
  const value = await evaluateExpression(expression.expression, context);
  const typeText = normalizeTypeText(expression.type.text);
  if (runtimeValueMatchesType(value, typeText, context)) return [assertedRuntimeValue(value, typeText, context), true];
  return [defaultValueForTypeText(typeText, context), false];
}

function assertedRuntimeValue(value: RuntimeValue, typeText: string, context: EvaluationContext): RuntimeValue {
  const type = normalizeTypeText(typeText);
  const dynamicValue = value instanceof RuntimeInterfaceValue ? value.value : value;
  const targetInterface = interfaceTarget(type, context);
  if (targetInterface) return prepareInterfaceAssignment(dynamicValue, type, targetInterface, "type assertion", context);
  return dynamicValue;
}

async function evaluateCall(expression: CallExpression, context: EvaluationContext): Promise<RuntimeValue> {
  if (expression.callee.kind === "Identifier" && expression.callee.name === "make") {
    return evaluateMake(expression, context);
  }
  if (expression.callee.kind === "Identifier" && expression.callee.name === "new") {
    return evaluateNew(expression, context);
  }
  const conversionType = conversionTargetType(expression.callee, context);
  if (conversionType) {
    if (expression.args.length !== 1 || expression.spreadLast) {
      throw new GoJuniorRuntimeError(`conversion to ${conversionType} expects exactly one argument`);
    }
    return convertValueToType(await evaluateExpression(expression.args[0]!, context), conversionType, context);
  }
  const callee = await evaluateExpression(expression.callee, context);
  const args = await evaluateExpressionList(expression.args, context);
  return callRuntime(callee, expression.spreadLast ? spreadLastArgument(args) : args, context);
}

function conversionTargetType(callee: Expression, context: EvaluationContext): string | undefined {
  if (callee.kind === "TypeExpression") return callee.type.text;
  if (callee.kind === "Identifier" && context.isKnownType(callee.name)) return callee.name;
  return undefined;
}

function evaluateNew(expression: CallExpression, context: EvaluationContext): RuntimeValue {
  if (expression.spreadLast || expression.args.length !== 1) throw new GoJuniorRuntimeError("new expects exactly one type argument");
  const typeArg = expression.args[0];
  const typeText = typeArg?.kind === "TypeExpression"
    ? typeArg.type.text
    : typeArg?.kind === "Identifier" && context.isKnownType(typeArg.name)
      ? typeArg.name
      : undefined;
  if (!typeText) throw new GoJuniorRuntimeError("new expects a type argument");
  let value = defaultValueForTypeText(typeText, context);
  return new RuntimePointer(normalizeTypeText(typeText), () => value, (next) => {
    assertAssignableToType(next, typeText, `*${typeText}`, context);
    value = next;
  });
}

async function evaluateMake(expression: CallExpression, context: EvaluationContext): Promise<RuntimeValue> {
  if (expression.spreadLast) throw new GoJuniorRuntimeError("make does not accept spread arguments");
  const typeArg = expression.args[0];
  if (!typeArg || typeArg.kind !== "TypeExpression") {
    throw new GoJuniorRuntimeError("make expects a map or slice type as its first argument");
  }
  const typeText = normalizeTypeText(typeArg.type.text);
  const mapType = parseMapTypeText(typeText);
  if (mapType) {
    if (expression.args.length > 2) throw new GoJuniorRuntimeError("make map accepts at most one size hint");
    if (expression.args[1]) toNonNegativeLength(await evaluateExpression(expression.args[1], context), "map size hint");
    return new RuntimeMap(mapType.keyType, mapType.valueType, context);
  }
  const chanType = parseChanTypeText(typeText);
  if (chanType) {
    if (chanType.direction !== "both") throw new GoJuniorRuntimeError(`cannot make directional channel type ${typeText}`);
    if (expression.args.length > 2) throw new GoJuniorRuntimeError("make channel accepts at most one buffer size");
    const capacity = expression.args[1]
      ? toNonNegativeLength(await evaluateExpression(expression.args[1], context), "channel buffer size")
      : 0;
    return new RuntimeChannel(chanType.elementType, capacity, context);
  }
  const arrayType = parseArrayOrSliceTypeText(typeText);
  if (arrayType) {
    if (arrayType.length !== undefined) throw new GoJuniorRuntimeError(`cannot make array type ${typeText}; use a slice type`);
    if (expression.args.length < 2 || expression.args.length > 3) {
      throw new GoJuniorRuntimeError("make slice expects length and optional capacity");
    }
    const length = toNonNegativeLength(await evaluateExpression(expression.args[1]!, context), "slice length");
    const capacity = expression.args[2]
      ? toNonNegativeLength(await evaluateExpression(expression.args[2], context), "slice capacity")
      : length;
    if (capacity < length) throw new GoJuniorRuntimeError("slice capacity is smaller than length");
    return makeRuntimeSlice(arrayType.elementType, length, capacity, context);
  }
  throw new GoJuniorRuntimeError(`cannot make ${typeText}`);
}

function spreadLastArgument(args: RuntimeValue[]): RuntimeValue[] {
  if (args.length === 0) return args;
  const last = args[args.length - 1];
  if (!Array.isArray(last)) {
    throw new GoJuniorRuntimeError(`${formatValue(last ?? null)} is not spreadable`);
  }
  return [...args.slice(0, -1), ...last];
}

function appendValues(target: RuntimeValue, values: RuntimeValue[]): RuntimeValue[] {
  if (target !== null && !Array.isArray(target)) {
    throw new GoJuniorRuntimeError(`${formatValue(target)} is not appendable`);
  }
  const base = target ?? [];
  const appended = [...base, ...values];
  const previousCapacity = sliceCapacity(base);
  const needed = appended.length;
  arrayCapacities.set(appended, Math.max(previousCapacity, needed));
  return appended;
}

function clearValue(target: RuntimeValue): void {
  target = unwrapNamed(target);
  if (target instanceof RuntimeMap) {
    target.clear();
    return;
  }
  if (Array.isArray(target)) {
    for (let index = 0; index < target.length; index += 1) target[index] = null;
    return;
  }
  throw new GoJuniorRuntimeError("clear expects a map or slice");
}

function copyValues(target: RuntimeValue, source: RuntimeValue): number {
  target = unwrapNamed(target);
  source = unwrapNamed(source);
  if (!Array.isArray(target)) throw new GoJuniorRuntimeError("copy destination must be a slice");
  const sourceValues = typeof source === "string" ? [...source] : source;
  if (!Array.isArray(sourceValues)) throw new GoJuniorRuntimeError("copy source must be a slice or string");
  const count = Math.min(target.length, sourceValues.length);
  for (let index = 0; index < count; index += 1) target[index] = sourceValues[index] ?? null;
  return count;
}

function minMaxValues(args: RuntimeValue[], mode: "min" | "max"): RuntimeValue {
  if (args.length === 0) throw new GoJuniorRuntimeError(`${mode} expects at least one argument`);
  let best = unwrapNamed(args[0] ?? null);
  for (const arg of args.slice(1)) {
    const candidate = unwrapNamed(arg);
    const comparison = compareValues(candidate, best);
    if ((mode === "min" && comparison < 0) || (mode === "max" && comparison > 0)) best = candidate;
  }
  return best;
}

async function callRuntime(callee: RuntimeValue, args: RuntimeValue[], context: EvaluationContext): Promise<RuntimeValue> {
  if (isRuntimeCallable(callee) || isGoJuniorFunction(callee)) {
    return await callee.call(args, context);
  }
  throw new GoJuniorRuntimeError(`${formatValue(callee)} is not callable`);
}

async function getSelector(expression: SelectorExpression, context: EvaluationContext): Promise<RuntimeValue> {
  const object = await evaluateExpression(expression.object, context);
  if (object instanceof SheetBinding) {
    return object.get(expression.field);
  }
  const struct = structFromValue(object);
  if (struct) {
    const fieldValue = struct.get(expression.field);
    if (fieldValue !== undefined) return fieldValue;
    const promotedField = promotedFieldAccessor(struct, expression.field, context);
    if (promotedField) return promotedField.get() ?? null;
  }
  const method = methodForValue(object, expression.field, context);
  if (method) {
    let receiver = method.receiver;
    if (method.method.pointerReceiver && expression.object.kind === "Identifier") {
      const receiverType = receiverTypeName(receiver);
      const bindingType = context.lookupTypeText(expression.object.name);
      if (receiverType === method.method.receiverType && normalizeTypeText(bindingType ?? "") === receiverType) {
        receiver = context.pointerToBinding(expression.object.name, receiverType);
      }
    }
    return boundMethodValue(method.method, receiver);
  }
  const namedType = expression.object.kind === "Identifier" ? context.lookupTypeText(expression.object.name) : undefined;
  if (namedType && expression.object.kind === "Identifier") {
    const method = context.methodFor(normalizeTypeText(namedType), expression.field);
    if (method) {
      const receiver = method.pointerReceiver
        ? context.pointerToBinding(expression.object.name, normalizeTypeText(namedType))
        : object;
      return boundMethodValue(method, receiver);
    }
  }
  const runtimeObject = dereferenceIfPointer(object);
  if (isRuntimeObject(runtimeObject)) {
    const value = runtimeObject[expression.field];
    if (value !== undefined) return value;
  }
  throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}

async function setSelector(expression: SelectorExpression, value: RuntimeValue, context: EvaluationContext): Promise<void> {
  const object = await evaluateExpression(expression.object, context);
  if (object instanceof SheetBinding) {
    object.set(expression.field, value);
    return;
  }
  const struct = structFromValue(object);
  if (struct) {
    const field = structFieldAccessor(struct, expression.field, context);
    if (!field) throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${expression.field}`);
    field.set(prepareAssignableToType(value, field.type.type.text, `field ${expression.field}`, context));
    return;
  }
  if (isRuntimeObject(object)) {
    object[expression.field] = value;
    return;
  }
  throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}

async function getSpreadsheetRange(expression: SpreadsheetRangeExpression, context: EvaluationContext): Promise<RuntimeValue> {
  const sheet = await evaluateExpression(expression.start.object, context);
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

async function setIndex(expression: IndexExpression, value: RuntimeValue, context: EvaluationContext): Promise<void> {
  const object = await evaluateExpression(expression.object, context);
  const index = await evaluateExpression(expression.index, context);
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

async function pointerToExpression(expression: Expression, context: EvaluationContext): Promise<RuntimePointer> {
  if (expression.kind === "StructLiteralExpression" || expression.kind === "ArrayLiteralExpression" || expression.kind === "MapLiteralExpression") {
    let value = await evaluateExpression(expression, context);
    return new RuntimePointer(pointerTypeName(value), () => value, (next) => {
      value = next;
    });
  }
  if (expression.kind === "Identifier") {
    const value = context.lookup(expression.name);
    const typeName = pointerTypeName(value);
    return context.pointerToBinding(expression.name, typeName);
  }
  if (expression.kind === "SelectorExpression") {
    const object = await evaluateExpression(expression.object, context);
    const struct = structFromValue(object);
    if (!struct) throw new GoJuniorRuntimeError("address-of selector requires a struct value");
    const field = structFieldAccessor(struct, expression.field, context);
    if (!field) throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${expression.field}`);
    const current = field.get();
    const typeName = pointerTypeName(current ?? null);
    return new RuntimePointer(typeName, () => field.get() ?? null, (next) => {
      field.set(prepareAssignableToType(next, field.type.type.text, `field ${expression.field}`, context));
    });
  }
  if (expression.kind === "IndexExpression") {
    const object = await evaluateExpression(expression.object, context);
    const index = await evaluateExpression(expression.index, context);
    if (!Array.isArray(object)) throw new GoJuniorRuntimeError("address-of index requires an array or slice value");
    const numericIndex = toNumber(index);
    const typeName = pointerTypeName(object[numericIndex] ?? null);
    return new RuntimePointer(typeName, () => object[numericIndex] ?? null, (next) => {
      object[numericIndex] = next;
    });
  }
  throw new GoJuniorRuntimeError("expression is not addressable");
}

function dereference(value: RuntimeValue): RuntimeValue {
  if (value instanceof RuntimeTypedNilValue && value.typeName.startsWith("*")) {
    throw new GoJuniorRuntimeError("nil pointer dereference");
  }
  if (!(value instanceof RuntimePointer)) throw new GoJuniorRuntimeError(`${formatValue(value)} is not a pointer`);
  return value.get();
}

function dereferenceIfPointer(value: RuntimeValue): RuntimeValue {
  return value instanceof RuntimePointer ? value.get() : value;
}

function structFromValue(value: RuntimeValue): RuntimeStruct | undefined {
  const actual = dereferenceIfPointer(value);
  return actual instanceof RuntimeStruct ? actual : undefined;
}

interface StructFieldAccessor {
  type: StructFieldDecl;
  get(): RuntimeValue | undefined;
  set(value: RuntimeValue): void;
}

interface MethodLookup {
  method: MethodDef;
  receiver: RuntimeValue;
}

function structFieldAccessor(struct: RuntimeStruct, fieldName: string, context: EvaluationContext): StructFieldAccessor | undefined {
  const direct = directStructFieldAccessor(struct, fieldName, context);
  if (direct) return direct;
  return promotedFieldAccessor(struct, fieldName, context);
}

function directStructFieldAccessor(struct: RuntimeStruct, fieldName: string, context: EvaluationContext): StructFieldAccessor | undefined {
  const typeDef = context.typeDef(struct.typeName);
  const field = typeDef?.fields.find((candidate) => candidate.name === fieldName);
  if (!field) return undefined;
  return {
    type: field,
    get: () => struct.get(fieldName),
    set: (value) => struct.set(fieldName, value)
  };
}

function promotedFieldAccessor(
  struct: RuntimeStruct,
  fieldName: string,
  context: EvaluationContext,
  seen = new Set<string>()
): StructFieldAccessor | undefined {
  if (seen.has(struct.typeName)) return undefined;
  seen.add(struct.typeName);
  const typeDef = context.typeDef(struct.typeName);
  if (!typeDef) return undefined;

  const matches: StructFieldAccessor[] = [];
  for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
    const embedded = structFromValue(struct.get(field.name) ?? null);
    if (!embedded) continue;
    const direct = directStructFieldAccessor(embedded, fieldName, context);
    if (direct) {
      matches.push(direct);
      continue;
    }
    const promoted = promotedFieldAccessor(embedded, fieldName, context, seen);
    if (promoted) matches.push(promoted);
  }
  if (matches.length > 1) throw new GoJuniorRuntimeError(`ambiguous promoted field ${fieldName}`);
  return matches[0];
}

function methodForValue(value: RuntimeValue, methodName: string, context: EvaluationContext): MethodLookup | undefined {
  if (value instanceof RuntimeInterfaceValue) {
    if (value.value === null) return undefined;
    return methodForValue(value.value, methodName, context);
  }
  const actual = dereferenceIfPointer(value);
  const receiverType = actual instanceof RuntimeStruct
    ? actual.typeName
    : actual instanceof RuntimeNamedValue
      ? actual.typeName
      : actual instanceof RuntimeTypedNilValue
        ? nilReceiverTypeName(actual)
      : undefined;
  if (receiverType) {
    const direct = context.methodFor(receiverType, methodName);
    if (direct) return { method: direct, receiver: value };
  }
  if (actual instanceof RuntimeStruct) return promotedMethodForStruct(actual, methodName, context);
  return undefined;
}

function promotedMethodForStruct(
  struct: RuntimeStruct,
  methodName: string,
  context: EvaluationContext,
  seen = new Set<string>()
): MethodLookup | undefined {
  if (seen.has(struct.typeName)) return undefined;
  seen.add(struct.typeName);
  const typeDef = context.typeDef(struct.typeName);
  if (!typeDef) return undefined;

  const matches: MethodLookup[] = [];
  for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
    const receiver = struct.get(field.name);
    if (receiver === undefined || receiver === null) continue;
    const receiverType = receiverTypeName(receiver);
    if (receiverType) {
      const direct = context.methodFor(receiverType, methodName);
      if (direct) {
        matches.push({ method: direct, receiver });
        continue;
      }
    }
    const embedded = structFromValue(receiver);
    if (!embedded) continue;
    const promoted = promotedMethodForStruct(embedded, methodName, context, seen);
    if (promoted) matches.push(promoted);
  }
  if (matches.length > 1) throw new GoJuniorRuntimeError(`ambiguous promoted method ${methodName}`);
  return matches[0];
}

function receiverTypeName(value: RuntimeValue): string | undefined {
  if (value instanceof RuntimeInterfaceValue) return value.value === null ? undefined : receiverTypeName(value.value);
  const actual = dereferenceIfPointer(value);
  if (actual instanceof RuntimeStruct) return actual.typeName;
  if (actual instanceof RuntimeNamedValue) return actual.typeName;
  if (actual instanceof RuntimeTypedNilValue) return nilReceiverTypeName(actual);
  return undefined;
}

function nilReceiverTypeName(value: RuntimeTypedNilValue): string {
  return value.typeName.startsWith("*") ? value.typeName.slice(1) : value.typeName;
}

function isTypedNilPointer(value: RuntimeValue): boolean {
  return value instanceof RuntimeTypedNilValue && value.typeName.startsWith("*");
}

function assertAssignableToStructField(struct: RuntimeStruct, fieldName: string, value: RuntimeValue, context: EvaluationContext): void {
  const field = structFieldAccessor(struct, fieldName, context);
  if (!field) throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${fieldName}`);
  assertAssignableToType(value, field.type.type.text, `field ${fieldName}`, context);
}

function pointerTypeName(value: RuntimeValue): string {
  const actual = dereferenceIfPointer(value);
  if (actual instanceof RuntimeInterfaceValue) return actual.interfaceType;
  if (actual instanceof RuntimeTypedNilValue) return actual.typeName;
  if (actual instanceof RuntimeNamedValue) return actual.typeName;
  if (actual instanceof RuntimeStruct) return actual.typeName;
  if (isComplexValue(actual)) return "complex128";
  if (typeof actual === "bigint") return "int";
  if (typeof actual === "number") return "float64";
  if (typeof actual === "string") return "string";
  if (typeof actual === "boolean") return "bool";
  return "interface{}";
}

function inferredTypeText(value: RuntimeValue): string | undefined {
  if (value === null) return undefined;
  if (value instanceof RuntimeNamedValue) return value.typeName;
  if (value instanceof RuntimeInterfaceValue) return value.interfaceType;
  if (value instanceof RuntimeTypedNilValue) return value.typeName;
  if (typeof value === "bigint") return "int64";
  if (typeof value === "number") return "float64";
  if (isComplexValue(value)) return "complex128";
  if (typeof value === "string") return "string";
  if (typeof value === "boolean") return "bool";
  if (Array.isArray(value)) return undefined;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return `map[${value.keyType}]${value.valueType}`;
  if (value instanceof RuntimeStruct) return value.typeName;
  if (value instanceof RuntimePointer) return `*${value.typeName}`;
  return undefined;
}

function getSlice(object: RuntimeValue, start: RuntimeValue | undefined, end: RuntimeValue | undefined, max: RuntimeValue | undefined): RuntimeValue {
  const startIndex = start === undefined ? undefined : toNumber(start);
  const endIndex = end === undefined ? undefined : toNumber(end);
  const maxIndex = max === undefined ? undefined : toNumber(max);
  if (Array.isArray(object)) {
    const low = startIndex ?? 0;
    const high = endIndex ?? object.length;
    const capEnd = maxIndex ?? sliceCapacity(object);
    if (capEnd < high) throw new GoJuniorRuntimeError("slice max is smaller than high bound");
    const result = object.slice(low, high);
    arrayCapacities.set(result, Math.max(0, capEnd - low));
    return result;
  }
  if (max !== undefined) throw new GoJuniorRuntimeError("three-index slicing is only supported for arrays and slices");
  if (typeof object === "string") return object.slice(startIndex, endIndex);
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not sliceable`);
}

function defaultValueForDeclarationType(type: TypeNode | undefined, context?: EvaluationContext): RuntimeValue {
  return defaultValueForTypeText(type?.text ?? "", context);
}

function defaultValueForTypeText(typeText: string, context?: EvaluationContext): RuntimeValue {
  const mapType = parseMapTypeText(typeText);
  if (mapType) return new RuntimeMap(mapType.keyType, mapType.valueType, context);
  if (parseChanTypeText(typeText)) return null;
  const arrayType = parseArrayOrSliceTypeText(typeText);
  if (arrayType) {
    if (arrayType.length === undefined || arrayType.inferLength) return [];
    return Array.from({ length: arrayType.length }, () => defaultValueForTypeText(arrayType.elementType, context));
  }
  const type = normalizeTypeText(typeText);
  const interfaceType = interfaceTarget(type, context);
  if (interfaceType) return new RuntimeInterfaceValue(type, null);
  if (type.startsWith("*")) return new RuntimeTypedNilValue(type);
  const typeDef = context?.typeDef(type);
  if (typeDef) {
    const struct = new RuntimeStruct(typeDef.name);
    for (const field of typeDef.fields) {
      struct.set(field.name, defaultValueForTypeText(field.type.text, context));
    }
    return struct;
  }
  const alias = context?.aliasType(type);
  if (alias && alias !== type) {
    const base = defaultValueForTypeText(alias, context);
    return new RuntimeNamedValue(type, base);
  }
  return zeroValueForMapValue(typeText);
}

function zeroValueForMapValue(typeText: string): RuntimeValue {
  const type = normalizeTypeText(typeText);
  if (isIntegerType(type)) return 0n;
  if (isFloatType(type)) return 0;
  if (isComplexType(type)) return complexValue(0, 0);
  if (type === "string") return "";
  if (type === "bool") return false;
  return null;
}

function convertValueToType(value: RuntimeValue, typeText: string, context: EvaluationContext): RuntimeValue {
  const type = normalizeTypeText(typeText);
  const targetInterface = interfaceTarget(type, context);
  if (targetInterface) return prepareInterfaceAssignment(value, type, targetInterface, "conversion", context);
  const alias = context.aliasType(type);
  if (alias && alias !== type) {
    return new RuntimeNamedValue(type, convertValueToType(value, alias, context));
  }
  const actual = unwrapNamed(value);
  if (type === "any" || type === "interface{}") return actual;
  if (type === "bool") {
    if (typeof actual !== "boolean") throwTypeError(actual, type, "conversion");
    return actual;
  }
  if (isIntegerType(type)) {
    if (typeof actual === "bigint") {
      if (!integerInRange(actual, type)) throwTypeError(actual, type, "conversion");
      return actual;
    }
    if (typeof actual === "number") {
      const converted = BigInt(Math.trunc(actual));
      if (!integerInRange(converted, type)) throwTypeError(actual, type, "conversion");
      return converted;
    }
    if (isComplexValue(actual) && actual.imag === 0) {
      const converted = BigInt(Math.trunc(actual.real));
      if (!integerInRange(converted, type)) throwTypeError(actual, type, "conversion");
      return converted;
    }
    throwTypeError(actual, type, "conversion");
  }
  if (isFloatType(type)) {
    if (isComplexValue(actual)) {
      if (actual.imag !== 0) throwTypeError(actual, type, "conversion");
      return actual.real;
    }
    return toFloat(actual);
  }
  if (isComplexType(type)) {
    return toComplex(actual);
  }
  if (type === "string") {
    if (typeof actual === "string") return actual;
    if (typeof actual === "bigint") return String.fromCodePoint(Number(actual));
    throwTypeError(actual, type, "conversion");
  }
  if (context.typeDef(type) && actual instanceof RuntimeStruct && actual.typeName === type) return actual;
  if (actual === null && type.startsWith("*")) return new RuntimeTypedNilValue(type);
  if (actual === null && isNilAssignableType(type)) return null;
  throw new GoJuniorRuntimeError(`unsupported conversion to ${type}`);
}

function prepareAssignableToType(value: RuntimeValue, typeText: string, role: string, context?: EvaluationContext): RuntimeValue {
  const type = normalizeTypeText(typeText);
  if (!type || type === "<missing>") return value;
  const interfaceType = interfaceTarget(type, context);
  if (interfaceType) return prepareInterfaceAssignment(value, type, interfaceType, role, context);
  if (value instanceof RuntimeNamedValue) {
    if (value.typeName === type) return value;
    value = value.value;
  }
  const alias = context?.aliasType(type);
  if (alias && alias !== type && !context?.typeDef(type) && !context?.interfaceDef(type)) {
    prepareAssignableToType(value, alias, role, context);
    return value;
  }
  if (value === null) {
    if (type.startsWith("*")) return new RuntimeTypedNilValue(type);
    if (isNilAssignableType(type)) return value;
    throw new GoJuniorRuntimeError(`${role} nil is not assignable to ${type}`);
  }
  if (value instanceof RuntimeTypedNilValue) {
    if (value.typeName === type || isNilAssignableType(type)) return value;
    throwTypeError(value, type, role);
  }

  if (type === "string") {
    if (typeof value !== "string") throwTypeError(value, type, role);
    return value;
  }
  if (type === "bool") {
    if (typeof value !== "boolean") throwTypeError(value, type, role);
    return value;
  }
  if (isIntegerType(type)) {
    if (typeof value !== "bigint" || !integerInRange(value, type)) throwTypeError(value, type, role);
    return value;
  }
  if (isFloatType(type)) {
    if (typeof value !== "number") throwTypeError(value, type, role);
    return value;
  }
  if (isComplexType(type)) {
    if (!isComplexValue(value)) throwTypeError(value, type, role);
    return value;
  }
  const mapType = parseMapTypeText(type);
  if (mapType) {
    if (!(value instanceof RuntimeMap)) throwTypeError(value, type, role);
    if (normalizeTypeText(value.keyType) !== normalizeTypeText(mapType.keyType) ||
      normalizeTypeText(value.valueType) !== normalizeTypeText(mapType.valueType)) {
      throwTypeError(value, type, role);
    }
    return value;
  }
  const chanType = parseChanTypeText(type);
  if (chanType) {
    if (!(value instanceof RuntimeChannel) || normalizeTypeText(value.elementType) !== normalizeTypeText(chanType.elementType)) {
      throwTypeError(value, type, role);
    }
    return value;
  }
  if (type.startsWith("*")) {
    if (!(value instanceof RuntimePointer) || value.typeName !== type.slice(1)) throwTypeError(value, type, role);
    return value;
  }
  const structType = context?.typeDef(type);
  if (structType) {
    if (!(value instanceof RuntimeStruct) || value.typeName !== structType.name) throwTypeError(value, type, role);
    return value;
  }
  if (value instanceof RuntimeStruct) {
    if (value.typeName !== type) throwTypeError(value, type, role);
    return value;
  }
  const arrayType = parseArrayOrSliceTypeText(type);
  if (arrayType) {
    if (!Array.isArray(value)) throwTypeError(value, type, role);
    if (arrayType.length !== undefined && !arrayType.inferLength && value.length !== arrayType.length) {
      throwTypeError(value, type, role);
    }
    for (const item of value) {
      assertAssignableToType(item, arrayType.elementType, `${role} element`, context);
    }
    return value;
  }
  if (type.startsWith("func(")) {
    if (!isRuntimeCallable(value) && !isGoJuniorFunction(value)) throwTypeError(value, type, role);
    return value;
  }
  return value;
}

function assertAssignableToType(value: RuntimeValue, typeText: string, role: string, context?: EvaluationContext): void {
  prepareAssignableToType(value, typeText, role, context);
}

function interfaceTarget(type: string, context?: EvaluationContext): InterfaceTypeDef | undefined {
  if (type === "any" || type === "interface{}") return { name: type, methods: [], embeds: [] };
  return context?.interfaceDef(type);
}

function prepareInterfaceAssignment(
  value: RuntimeValue,
  type: string,
  interfaceType: InterfaceTypeDef,
  role: string,
  context?: EvaluationContext
): RuntimeInterfaceValue {
  if (value instanceof RuntimeInterfaceValue) {
    const dynamicValue = value.value;
    if (dynamicValue === null) return new RuntimeInterfaceValue(type, null);
    if (context && !valueImplementsInterface(dynamicValue, interfaceType, context)) throwTypeError(value, type, role);
    return new RuntimeInterfaceValue(type, dynamicValue);
  }
  if (value === null) return new RuntimeInterfaceValue(type, null);
  if (context && !valueImplementsInterface(value, interfaceType, context)) throwTypeError(value, type, role);
  return new RuntimeInterfaceValue(type, value);
}

function valueImplementsInterface(value: RuntimeValue, interfaceType: InterfaceTypeDef, context: EvaluationContext): boolean {
  if (value instanceof RuntimeInterfaceValue) {
    return value.value === null || valueImplementsInterface(value.value, interfaceType, context);
  }
  const receiver = dereferenceIfPointer(value);
  const receiverType = receiver instanceof RuntimeStruct
    ? receiver.typeName
    : receiver instanceof RuntimeNamedValue
      ? receiver.typeName
      : receiver instanceof RuntimeTypedNilValue
        ? nilReceiverTypeName(receiver)
      : undefined;
  if (!receiverType) return interfaceType.methods.length === 0;

  for (const method of interfaceType.methods) {
    const candidate = methodForValue(value, method.name, context);
    if (!candidate) return false;
    if (candidate.method.pointerReceiver && !(value instanceof RuntimePointer) && !isTypedNilPointer(value)) return false;
    if (!signaturesCompatible(candidate.method.declaration.signature, method.signature)) return false;
  }
  return true;
}

function signaturesCompatible(actual: FunctionDecl["signature"], expected: FunctionDecl["signature"]): boolean {
  if (actual.parameters.length !== expected.parameters.length || actual.results.length !== expected.results.length) return false;
  return actual.parameters.every((param, index) => {
    const expectedParam = expected.parameters[index];
    if (!expectedParam) return false;
    return normalizeTypeText(param.type.text) === normalizeTypeText(expectedParam.type.text) &&
      param.variadic === expectedParam.variadic;
  }) && actual.results.every((result, index) => {
    const expectedResult = expected.results[index];
    if (!expectedResult) return false;
    return normalizeTypeText(result.type.text) === normalizeTypeText(expectedResult.type.text) &&
      result.variadic === expectedResult.variadic;
  });
}

function runtimeValueMatchesType(value: RuntimeValue, typeText: string, context?: EvaluationContext): boolean {
  const type = normalizeTypeText(typeText);
  if (value instanceof RuntimeInterfaceValue) {
    if (type === "nil") return value.value === null;
    if (value.value === null) return false;
    return runtimeValueMatchesType(value.value, type, context);
  }
  if (type === "nil") return value === null;
  if (value === null) return false;
  if (value instanceof RuntimeNamedValue) {
    return value.typeName === type || runtimeValueMatchesType(value.value, type, context);
  }
  if (value instanceof RuntimeTypedNilValue) {
    if (type.startsWith("*")) return value.typeName === type;
  }
  if (!type || type === "<missing>") return false;
  if (type === "any" || type === "interface{}") return true;
  const interfaceType = context?.interfaceDef(type);
  if (interfaceType && context) return valueImplementsInterface(value, interfaceType, context);
  if (type === "string") return typeof value === "string";
  if (type === "bool") return typeof value === "boolean";
  if (isIntegerType(type)) return typeof value === "bigint" && integerInRange(value, type);
  if (isFloatType(type)) return typeof value === "number";
  if (isComplexType(type)) return isComplexValue(value);
  if (type.startsWith("*")) return value instanceof RuntimePointer && value.typeName === type.slice(1);
  if (value instanceof RuntimeStruct) return value.typeName === type;
  if (value instanceof RuntimeMap) {
    const mapType = parseMapTypeText(type);
    return Boolean(mapType &&
      normalizeTypeText(value.keyType) === normalizeTypeText(mapType.keyType) &&
      normalizeTypeText(value.valueType) === normalizeTypeText(mapType.valueType));
  }
  if (value instanceof RuntimeChannel) {
    const chanType = parseChanTypeText(type);
    return Boolean(chanType && normalizeTypeText(value.elementType) === normalizeTypeText(chanType.elementType));
  }
  if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type)) return Array.isArray(value);
  if (type.startsWith("func(")) return isRuntimeCallable(value) || isGoJuniorFunction(value);
  return false;
}

function valueLength(value: RuntimeValue): number {
  if (typeof value === "string" || Array.isArray(value)) return value.length;
  if (value instanceof RuntimeMap) return value.size();
  if (value instanceof RuntimeChannel) return value.len();
  throw new GoJuniorRuntimeError(`${formatValue(value)} has no len`);
}

function valueCapacity(value: RuntimeValue): number {
  if (Array.isArray(value)) return sliceCapacity(value);
  if (value instanceof RuntimeChannel) return value.cap();
  throw new GoJuniorRuntimeError(`${formatValue(value)} has no cap`);
}

function sliceCapacity(value: RuntimeValue[]): number {
  return arrayCapacities.get(value) ?? value.length;
}

function toNonNegativeLength(value: RuntimeValue, role: string): number {
  const length = toNumber(value);
  if (length < 0) throw new GoJuniorRuntimeError(`${role} must be non-negative`);
  return length;
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

function parseChanTypeText(typeText: string): { elementType: string; direction: "send" | "receive" | "both" } | undefined {
  const type = normalizeTypeText(typeText);
  if (type.startsWith("chan<-")) {
    const elementType = type.slice("chan<-".length);
    return elementType ? { elementType, direction: "send" } : undefined;
  }
  if (type.startsWith("<-chan")) {
    const elementType = type.slice("<-chan".length);
    return elementType ? { elementType, direction: "receive" } : undefined;
  }
  if (type.startsWith("chan") && type.length > "chan".length) {
    return { elementType: type.slice("chan".length), direction: "both" };
  }
  return undefined;
}

function parseArrayOrSliceTypeText(typeText: string): { elementType: string; length?: number; inferLength: boolean } | undefined {
  const type = normalizeTypeText(typeText);
  if (!type.startsWith("[")) return undefined;

  const close = type.indexOf("]");
  if (close < 0) return undefined;
  const lengthText = type.slice(1, close);
  const elementType = type.slice(close + 1);
  if (!elementType) return undefined;

  if (lengthText === "") {
    return { elementType, inferLength: false };
  }
  if (lengthText === "...") {
    return { elementType, inferLength: true };
  }
  if (!/^(0|[1-9][0-9]*)$/.test(lengthText)) return undefined;
  return { elementType, length: Number(lengthText), inferLength: false };
}

function normalizeTypeText(typeText: string): string {
  return typeText.replace(/\s+/g, "");
}

function runtimeMapKeyId(key: RuntimeValue): string {
  const actual = unwrapNamed(key);
  if (actual === null) return "nil";
  if (actual instanceof RuntimeInterfaceValue) {
    return actual.value === null
      ? `iface:${actual.interfaceType}:nil`
      : `iface:${actual.interfaceType}:${runtimeMapKeyId(actual.value)}`;
  }
  if (actual instanceof RuntimeTypedNilValue) return `typednil:${actual.typeName}`;
  if (typeof actual === "boolean") return `b:${actual}`;
  if (typeof actual === "string") return `s:${actual}`;
  if (typeof actual === "bigint") return `i:${actual}`;
  if (typeof actual === "number") return `f:${Object.is(actual, -0) ? "-0" : String(actual)}`;
  if (isComplexValue(actual)) return `c:${actual.real}:${actual.imag}`;
  if (Array.isArray(actual)) return `a:[${actual.map(runtimeMapKeyId).join(",")}]`;
  if (actual instanceof RuntimeStruct) {
    return `st:${actual.typeName}{${actual.orderedFields().map(([name, item]) => `${name}:${runtimeMapKeyId(item)}`).join(",")}}`;
  }
  return `o:${objectIdentityId(actual as object)}`;
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
    parseChanTypeText(type) !== undefined ||
    type.startsWith("func(");
}

function assertComparableType(typeText: string, context?: EvaluationContext): void {
  const type = normalizeTypeText(typeText);
  if (type.startsWith("[]") || type.startsWith("map[") || type.startsWith("func(")) {
    throw new GoJuniorRuntimeError(`map key type ${type} is not comparable`);
  }
  const arrayType = parseArrayOrSliceTypeText(type);
  if (arrayType) {
    if (arrayType.length === undefined || arrayType.inferLength) {
      throw new GoJuniorRuntimeError(`map key type ${type} is not comparable`);
    }
    assertComparableType(arrayType.elementType, context);
    return;
  }
  const alias = context?.aliasType(type);
  if (alias && alias !== type && !context?.typeDef(type) && !context?.interfaceDef(type)) {
    assertComparableType(alias, context);
    return;
  }
  const structType = context?.typeDef(type);
  if (structType) {
    for (const field of structType.fields) assertComparableType(field.type.text, context);
  }
}

function assertComparableValue(value: RuntimeValue, role: string): void {
  value = unwrapNamed(value);
  if (value instanceof RuntimeInterfaceValue) {
    if (value.value !== null) assertComparableValue(value.value, role);
    return;
  }
  if (value instanceof RuntimeTypedNilValue) return;
  if (value instanceof RuntimeMap || isRuntimeCallable(value) || isGoJuniorFunction(value)) {
    throw new GoJuniorRuntimeError(`${role} ${formatValue(value)} is not comparable`);
  }
  if (Array.isArray(value)) {
    for (const item of value) assertComparableValue(item, `${role} element`);
    return;
  }
  if (value instanceof RuntimeStruct) {
    for (const [field, item] of value.orderedFields()) assertComparableValue(item, `${role} field ${field}`);
  }
}

function isIntegerType(type: string): boolean {
  return /^(?:u?int(?:8|16|32|64)?|uintptr|byte|rune)$/.test(type);
}

function isFloatType(type: string): boolean {
  return type === "float32" || type === "float64";
}

function isComplexType(type: string): boolean {
  return type === "complex64" || type === "complex128";
}

function isPredeclaredType(type: string): boolean {
  return type === "bool" ||
    type === "string" ||
    type === "error" ||
    type === "any" ||
    type === "interface{}" ||
    isIntegerType(type) ||
    isFloatType(type) ||
    isComplexType(type);
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
    case "uintptr":
      return { min: 0n, max: 18446744073709551615n };
    case "int":
    case "int64":
      return { min: -9223372036854775808n, max: 9223372036854775807n };
    default:
      return undefined;
  }
}

function addValues(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  const actualLeft = unwrapNamed(left);
  const actualRight = unwrapNamed(right);
  if (typeof actualLeft === "string" || typeof actualRight === "string") {
    if (typeof actualLeft === "string" && typeof actualRight === "string") return actualLeft + actualRight;
    throw new GoJuniorRuntimeError(`invalid operation: ${runtimeTypeText(actualLeft)} + ${runtimeTypeText(actualRight)} (mismatched types ${runtimeTypeText(actualLeft)} and ${runtimeTypeText(actualRight)})`);
  }
  return addNumbers(actualLeft, actualRight);
}

function runtimeTypeText(value: RuntimeValue): string {
  return inferredTypeText(value) ?? pointerTypeName(value);
}

function addNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    return complexValue(a.real + b.real, a.imag + b.imag);
  }
  if (typeof left === "bigint" && typeof right === "bigint") return left + right;
  return toFloat(left) + toFloat(right);
}

function subtractNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    return complexValue(a.real - b.real, a.imag - b.imag);
  }
  if (typeof left === "bigint" && typeof right === "bigint") return left - right;
  return toFloat(left) - toFloat(right);
}

function multiplyNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    return complexValue(a.real * b.real - a.imag * b.imag, a.real * b.imag + a.imag * b.real);
  }
  if (typeof left === "bigint" && typeof right === "bigint") return left * right;
  return toFloat(left) * toFloat(right);
}

function divideNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    const denominator = b.real * b.real + b.imag * b.imag;
    return complexValue((a.real * b.real + a.imag * b.imag) / denominator, (a.imag * b.real - a.real * b.imag) / denominator);
  }
  if (typeof left === "bigint" && typeof right === "bigint") return left / right;
  return toFloat(left) / toFloat(right);
}

function moduloNumbers(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if (typeof left === "bigint" && typeof right === "bigint") return left % right;
  return toFloat(left) % toFloat(right);
}

function negateNumber(value: RuntimeValue): RuntimeValue {
  value = unwrapNamed(value);
  if (isComplexValue(value)) return complexValue(-value.real, -value.imag);
  if (typeof value === "bigint") return -value;
  return -toFloat(value);
}

function numericIdentity(value: RuntimeValue): RuntimeValue {
  value = unwrapNamed(value);
  if (isComplexValue(value)) return value;
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
  value = unwrapNamed(value);
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

function bitwiseComplement(value: RuntimeValue): RuntimeValue {
  return ~toBigInt(value);
}

function bitwiseAnd(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  return toBigInt(left) & toBigInt(right);
}

function bitwiseOr(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  return toBigInt(left) | toBigInt(right);
}

function bitwiseXor(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  return toBigInt(left) ^ toBigInt(right);
}

function bitClear(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  return toBigInt(left) & ~toBigInt(right);
}

function shiftLeft(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  const shift = toBigInt(right);
  if (shift < 0n) throw new GoJuniorRuntimeError("negative shift count");
  return toBigInt(left) << shift;
}

function shiftRight(left: RuntimeValue, right: RuntimeValue): RuntimeValue {
  const shift = toBigInt(right);
  if (shift < 0n) throw new GoJuniorRuntimeError("negative shift count");
  return toBigInt(left) >> shift;
}

function toBigInt(value: RuntimeValue): bigint {
  value = unwrapNamed(value);
  if (typeof value === "bigint") return value;
  if (typeof value === "number" && Number.isInteger(value)) return BigInt(value);
  throw new GoJuniorRuntimeError(`${formatValue(value)} is not an integer`);
}

function complexValue(real: number, imag: number): ComplexLiteralValue {
  return { real, imag };
}

function isComplexValue(value: RuntimeValue): value is ComplexLiteralValue {
  return Boolean(
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    !(value instanceof RuntimeMap) &&
    !(value instanceof RuntimeChannel) &&
    !(value instanceof RuntimeStruct) &&
    !(value instanceof RuntimePointer) &&
    !(value instanceof RuntimeNamedValue) &&
    !(value instanceof RuntimeInterfaceValue) &&
    !(value instanceof RuntimeTypedNilValue) &&
    "real" in value &&
    "imag" in value &&
    typeof value.real === "number" &&
    typeof value.imag === "number"
  );
}

function toComplex(value: RuntimeValue): ComplexLiteralValue {
  value = unwrapNamed(value);
  if (isComplexValue(value)) return value;
  return complexValue(toFloat(value), 0);
}

function unwrapNamed(value: RuntimeValue): RuntimeValue {
  return value instanceof RuntimeNamedValue ? value.value : value;
}

function compareValues(left: RuntimeValue, right: RuntimeValue): number {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
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
  if (left instanceof RuntimeInterfaceValue || right instanceof RuntimeInterfaceValue) {
    return interfaceAwareEqual(left, right);
  }
  if (left instanceof RuntimeTypedNilValue || right instanceof RuntimeTypedNilValue) {
    if (left === null || right === null) return true;
    return left instanceof RuntimeTypedNilValue &&
      right instanceof RuntimeTypedNilValue &&
      left.typeName === right.typeName;
  }
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  if ((typeof left === "bigint" || typeof left === "number") && (typeof right === "bigint" || typeof right === "number")) {
    return toFloat(left) === toFloat(right);
  }
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    return a.real === b.real && a.imag === b.imag;
  }
  if (Array.isArray(left) && Array.isArray(right)) {
    return left.length === right.length && left.every((item, index) => valueEqual(item, right[index] ?? null));
  }
  if (left instanceof RuntimeStruct && right instanceof RuntimeStruct) {
    if (left.typeName !== right.typeName) return false;
    const leftFields = left.orderedFields();
    const rightFields = right.orderedFields();
    return leftFields.length === rightFields.length &&
      leftFields.every(([name, value], index) => {
        const rightField = rightFields[index];
        return rightField?.[0] === name && valueEqual(value, rightField[1]);
      });
  }
  return left === right;
}

function interfaceAwareEqual(left: RuntimeValue, right: RuntimeValue): boolean {
  if (left instanceof RuntimeInterfaceValue && right === null) return left.value === null;
  if (right instanceof RuntimeInterfaceValue && left === null) return right.value === null;
  if (left instanceof RuntimeInterfaceValue && right instanceof RuntimeInterfaceValue) {
    if (left.value === null || right.value === null) return left.value === null && right.value === null;
    return valueEqual(left.value, right.value);
  }
  if (left instanceof RuntimeInterfaceValue) {
    if (left.value === null) return false;
    return valueEqual(left.value, right);
  }
  if (right instanceof RuntimeInterfaceValue) {
    if (right.value === null) return false;
    return valueEqual(left, right.value);
  }
  return false;
}

async function rangeEntries(source: RuntimeValue, context: EvaluationContext): Promise<Array<[RuntimeValue, RuntimeValue]>> {
  source = unwrapNamed(source);
  if (typeof source === "bigint" || typeof source === "number") {
    const count = toNonNegativeLength(source, "range count");
    return Array.from({ length: count }, (_, index) => [BigInt(index), BigInt(index)]);
  }
  if (isRuntimeCallable(source) || isGoJuniorFunction(source)) {
    return iteratorFunctionEntries(source, context);
  }
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

async function iteratorFunctionEntries(source: RuntimeCallable | GoJuniorFunction, context: EvaluationContext): Promise<Array<[RuntimeValue, RuntimeValue]>> {
  const entries: Array<[RuntimeValue, RuntimeValue]> = [];
  let index = 0n;
  const yieldFn = hostCallable("yield", (args) => {
    if (args.length === 0) {
      entries.push([index, index]);
      index += 1n;
    } else if (args.length === 1) {
      entries.push([args[0] ?? null, args[0] ?? null]);
    } else {
      entries.push([args[0] ?? null, args[1] ?? null]);
    }
    return true;
  });
  await callRuntime(source, [yieldFn], context);
  return entries;
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
  if (value instanceof RuntimeNamedValue) return formatValue(value.value);
  if (value instanceof RuntimeInterfaceValue) return value.value === null ? "<nil>" : formatValue(value.value);
  if (value instanceof RuntimeTypedNilValue) return "<nil>";
  if (value === null) return "<nil>";
  if (typeof value === "bigint") return value.toString();
  if (isComplexValue(value)) return `(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "string") return String(value);
  if (Array.isArray(value)) return `[${value.map(formatValue).join(" ")}]`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatRuntimeMap(value);
  if (value instanceof RuntimeStruct) return formatRuntimeStruct(value);
  if (value instanceof RuntimePointer) return `&${formatValue(value.get())}`;
  if (isRuntimeCallable(value) || isGoJuniorFunction(value)) return `<func ${value.name}>`;
  if (value instanceof SheetBinding) return `<sheet ${value.name}>`;
  return `{${Object.entries(value).map(([key, item]) => `${key}:${formatValue(item)}`).join(" ")}}`;
}

export function formatReplValue(value: RuntimeValue): string {
  if (value instanceof RuntimeNamedValue) return formatReplValue(value.value);
  if (value instanceof RuntimeInterfaceValue) {
    return value.value === null ? `${value.interfaceType}(nil)` : formatReplValue(value.value);
  }
  if (value instanceof RuntimeTypedNilValue) return `${value.typeName}(nil)`;
  if (value === null) return "<nil>";
  if (typeof value === "bigint") return value.toString();
  if (isComplexValue(value)) return `(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  if (typeof value === "string") return formatReplString(value);
  if (Array.isArray(value)) return `[${value.map(formatReplValue).join(" ")}]`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatReplMap(value);
  if (value instanceof RuntimeStruct) return formatReplStruct(value);
  if (value instanceof RuntimePointer) return `&${formatReplValue(value.get())}`;
  if (isRuntimeCallable(value) || isGoJuniorFunction(value)) return `<func ${value.name}>`;
  if (value instanceof SheetBinding) return `<sheet ${value.name}>`;
  return `{${Object.entries(value).map(([key, item]) => `${key}:${formatReplValue(item)}`).join(" ")}}`;
}

function formatGoSyntaxValue(value: RuntimeValue): string {
  if (value instanceof RuntimeNamedValue) return `${value.typeName}(${formatGoSyntaxValue(value.value)})`;
  if (value instanceof RuntimeInterfaceValue) {
    return value.value === null ? `${value.interfaceType}(nil)` : formatGoSyntaxValue(value.value);
  }
  if (value instanceof RuntimeTypedNilValue) return `${value.typeName}(nil)`;
  if (value === null) return "nil";
  if (typeof value === "bigint") return `${integerTypeName(value)}(${value.toString()})`;
  if (typeof value === "number") return `float64(${formatFloat(value)})`;
  if (isComplexValue(value)) return `complex128(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
  if (typeof value === "boolean") return `bool(${value})`;
  if (typeof value === "string") return `string(${JSON.stringify(value)})`;
  if (Array.isArray(value)) return `[]interface{}{${value.map(formatGoSyntaxValue).join(", ")}}`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatGoSyntaxMap(value);
  if (value instanceof RuntimeStruct) return formatGoSyntaxStruct(value);
  if (value instanceof RuntimePointer) return `&${formatGoSyntaxValue(value.get())}`;
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
  return `map[${value.keyType}]${value.valueType}{${entries}}`;
}

function formatRuntimeStruct(value: RuntimeStruct): string {
  const fields = value.orderedFields()
    .map(([key, item]) => `${key}:${formatValue(item)}`)
    .join(" ");
  return `${value.typeName}{${fields}}`;
}

function formatReplMap(value: RuntimeMap): string {
  const entries = value.orderedEntries()
    .map(([key, item]) => `${formatReplValue(key)}:${formatReplValue(item)}`)
    .join(", ");
  return `map[${value.keyType}]${value.valueType}{${entries}}`;
}

function formatReplStruct(value: RuntimeStruct): string {
  const fields = value.orderedFields()
    .map(([key, item]) => `${key}:${formatReplValue(item)}`)
    .join(", ");
  return `${value.typeName}{${fields}}`;
}

function formatReplString(value: string): string {
  if (value.includes("\n") || value.includes("\"")) return `\`${value.replace(/`/g, "\\`")}\``;
  return JSON.stringify(value);
}

function formatGoSyntaxMap(value: RuntimeMap): string {
  const entries = value.orderedEntries()
    .map(([key, item]) => `${formatGoSyntaxValue(key)}: ${formatGoSyntaxValue(item)}`)
    .join(", ");
  return `map[${value.keyType}]${value.valueType}{${entries}}`;
}

function formatGoSyntaxStruct(value: RuntimeStruct): string {
  const fields = value.orderedFields()
    .map(([key, item]) => `${key}: ${formatGoSyntaxValue(item)}`)
    .join(", ");
  return `${value.typeName}{${fields}}`;
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
    !(value instanceof RuntimeStruct) &&
    !(value instanceof RuntimePointer) &&
    !(value instanceof RuntimeNamedValue) &&
    !(value instanceof RuntimeInterfaceValue) &&
    !(value instanceof RuntimeTypedNilValue) &&
    !(value instanceof SheetBinding) &&
    !isComplexValue(value) &&
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
