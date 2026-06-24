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
  InterfaceMethodDecl,
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
import {
  checkGoJuniorSourceFiles,
  GOJR_SYNTHETIC_CHECK_PREFIX,
  isGoJuniorSyntheticCheckName,
  type GoJuniorCheckConfig,
  type GoJuniorCheckResult
} from "./typecheck.js";
import {
  ensureUniverseInitialized,
  NewPackage,
  type GoJuniorSheetNamespace,
  type Object as GoTypesObject,
  type Package as GoTypesPackage
} from "./go/types/index.js";
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
import { cellDependency, rangeDependency } from "./spreadsheet.js";
import type { SpreadsheetDependency } from "./spreadsheet.js";

export type RuntimeValue =
  | null
  | boolean
  | string
  | RuntimeGoString
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

export class RuntimeGoString {
  public constructor(public readonly bytes: Uint8Array) {}

  public static fromUtf8Text(text: string): RuntimeGoString {
    return new RuntimeGoString(new TextEncoder().encode(text));
  }

  public text(): string {
    return new TextDecoder("utf-8").decode(this.bytes);
  }

  public byteLength(): number {
    return this.bytes.length;
  }

  public byteAt(index: number): bigint {
    return BigInt(this.bytes[index] ?? 0);
  }

  public slice(start?: number, end?: number): RuntimeGoString {
    return new RuntimeGoString(this.bytes.slice(start, end));
  }

  public concat(other: RuntimeGoString): RuntimeGoString {
    const bytes = new Uint8Array(this.bytes.length + other.bytes.length);
    bytes.set(this.bytes, 0);
    bytes.set(other.bytes, this.bytes.length);
    return new RuntimeGoString(bytes);
  }
}

export interface GoJuniorFunction {
  kind: "GoJuniorFunction";
  name: string;
  signature?: FunctionDecl["signature"];
  declaration?: FunctionDecl;
  source?: string;
  call(args: RuntimeValue[], context: EvaluationContext): MaybePromise<RuntimeValue>;
}

export interface SheetData {
  [cell: string]: RuntimeValue;
}

export interface EvaluationOptions {
  filename?: string;
  packages?: Record<string, RuntimeObject>;
  packageInfos?: Record<string, GoTypesPackage>;
  env?: Record<string, string>;
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
  observedDeps?: SpreadsheetDependency[];
}

export interface PackageEvaluationOptions extends EvaluationOptions {
  importPath?: string;
  packageName?: string;
}

export interface PackageEvaluationResult {
  diagnostics: Diagnostic[];
  output: string[];
  package?: RuntimeObject;
  packageInfo?: GoTypesPackage;
}

export interface SourcePackageProvider {
  load(importPath: string): SourceFile[] | undefined;
}

export interface SourcePackageSpec {
  importPath: string;
  files: SourceFile[];
  packageName?: string;
}

export interface SourcePackageGraphOptions extends EvaluationOptions {
  sourcePackageProvider?: SourcePackageProvider;
}

export interface SourcePackageGraphEvaluationResult {
  diagnostics: Diagnostic[];
  output: string[];
  packages: Record<string, RuntimeObject>;
  packageInfos: Record<string, GoTypesPackage>;
  initializedImportPaths: string[];
}

type Completion =
  | { kind: "normal"; value?: RuntimeValue }
  | { kind: "return"; values: RuntimeValue[] }
  | { kind: "break"; label?: string }
  | { kind: "continue"; label?: string }
  | { kind: "fallthrough" }
  | { kind: "goto"; label: string };

type MaybePromise<T> = T | Promise<T>;

const RECOVERED_PANIC = Symbol("recovered panic");
type RecoveredPanic = typeof RECOVERED_PANIC;

interface PackageScopeReplacement {
  name: string;
  object: GoTypesObject;
}

interface Binding {
  value: RuntimeValue;
  mutable: boolean;
  typeText?: string;
}

interface StructTypeDef {
  name: string;
  fields: StructFieldDecl[];
}

const anonymousStructTypeCache = new Map<string, StructTypeDef>();

interface InterfaceTypeDef {
  name: string;
  methods: NonNullable<TypeSpec["interfaceMethods"]>;
  embeds: TypeNode[];
}

const anonymousInterfaceTypeCache = new Map<string, InterfaceTypeDef>();

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
  observedDeps: Map<string, SpreadsheetDependency>;
}

export class EvaluationContext {
  private readonly shared: EvaluationSharedState;
  private currentScope: Scope;
  private readonly deferFrames: Array<Array<() => MaybePromise<RuntimeValue>>> = [[]];
  private callDepth = 0;
  private activeRecoverPanic: GoJuniorPanic | undefined;
  private recoverCallDepth: number | undefined;
  private recoveredActivePanic = false;

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
      scheduler: new AsyncGoScheduler(options.randomSeed === undefined ? {} : { randomSeed: options.randomSeed }),
      observedDeps: new Map()
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
    if (name === "_") return;
    const typeText = bindingTypeText(type);
    const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
    this.currentScope.declare(name, stored, mutable, typeText ? resolvedDeclaredTypeText(typeText, this) : undefined);
  }

  public declareRoot(name: string, value: RuntimeValue, mutable = true, type?: TypeNode | string): void {
    if (name === "_") return;
    const typeText = bindingTypeText(type);
    const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
    this.shared.rootScope.declare(name, stored, mutable, typeText ? resolvedDeclaredTypeText(typeText, this) : undefined);
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

  public isUntypedConstant(name: string): boolean {
    return this.currentScope.isUntypedConstant(name);
  }

  public hasLocal(name: string): boolean {
    return this.currentScope.hasLocal(name);
  }

  public hasBinding(name: string): boolean {
    return this.currentScope.hasBinding(name);
  }

  public pointerToBinding(name: string, typeName: string): RuntimePointer {
    const scope = this.currentScope;
    return new RuntimePointer(typeName, () => scope.lookup(name), (next) => scope.assign(name, next, this.assignmentChecker()));
  }

  public outputFrom(offset: number): string[] {
    return this.output.slice(offset);
  }

  public clearObservedDeps(): void {
    this.shared.observedDeps.clear();
  }

  public observedDeps(): SpreadsheetDependency[] {
    return [...this.shared.observedDeps.values()];
  }

  public getenv(name: string): string {
    return this.options.env?.[name] ?? "";
  }

  public recordSheetCellRead(sheet: string, cell: string): void {
    const dependency = cellDependency({ sheet, cell });
    this.shared.observedDeps.set(dependencyKey(dependency), dependency);
  }

  public recordSheetRangeRead(sheet: string, start: string, end: string): void {
    const dependency = rangeDependency(sheet, start, end);
    this.shared.observedDeps.set(dependencyKey(dependency), dependency);
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

  public async functionCallAsync<T>(body: () => Promise<T>): Promise<T> {
    this.callDepth += 1;
    try {
      return await body();
    } finally {
      this.callDepth -= 1;
    }
  }

  public recover(): RuntimeValue {
    if (
      this.activeRecoverPanic &&
      this.recoverCallDepth === this.callDepth &&
      !this.recoveredActivePanic
    ) {
      this.recoveredActivePanic = true;
      return this.activeRecoverPanic.value;
    }
    return null;
  }

  public runDefers(): void {
    const frame = this.currentDeferFrame();
    while (frame.length > 0) {
      frame.pop()?.();
    }
  }

  public async runDefersAsync(panic?: GoJuniorPanic): Promise<GoJuniorPanic | undefined> {
    const frame = this.currentDeferFrame();
    let activePanic = panic;
    while (frame.length > 0) {
      const callback = frame.pop();
      if (!callback) continue;
      const previousRecoverPanic = this.activeRecoverPanic;
      const previousRecoverCallDepth = this.recoverCallDepth;
      const previousRecoveredActivePanic = this.recoveredActivePanic;
      if (activePanic) {
        this.activeRecoverPanic = activePanic;
        this.recoverCallDepth = this.callDepth + 1;
        this.recoveredActivePanic = false;
      } else {
        this.activeRecoverPanic = undefined;
        this.recoverCallDepth = undefined;
        this.recoveredActivePanic = false;
      }
      try {
        await callback();
      } catch (error) {
        const normalized = normalizeAsyncRuntimeError(error);
        if (normalized instanceof GoJuniorPanic) {
          activePanic = normalized;
        } else {
          throw normalized;
        }
      } finally {
        if (this.recoveredActivePanic) {
          activePanic = undefined;
        }
        this.activeRecoverPanic = previousRecoverPanic;
        this.recoverCallDepth = previousRecoverCallDepth;
        this.recoveredActivePanic = previousRecoveredActivePanic;
      }
    }
    return activePanic;
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

  public async deferScopeAsync<T>(body: () => Promise<T>): Promise<T | RecoveredPanic> {
    this.deferFrames.push([]);
    try {
      const value = await body();
      const panic = await this.runDefersAsync();
      if (panic) throw panic;
      return value;
    } catch (error) {
      const normalized = normalizeAsyncRuntimeError(error);
      if (normalized instanceof GoJuniorPanic) {
        const panic = await this.runDefersAsync(normalized);
        if (panic) throw panic;
        return RECOVERED_PANIC;
      }
      const panic = await this.runDefersAsync();
      if (panic) throw panic;
      throw normalized;
    } finally {
      this.deferFrames.pop();
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
    return this.shared.types.get(name) ?? this.shared.types.get(genericBaseTypeName(name));
  }

  public interfaceDef(name: string): InterfaceTypeDef | undefined {
    return this.shared.interfaces.get(name) ?? this.shared.interfaces.get(genericBaseTypeName(name));
  }

  public aliasType(name: string): string | undefined {
    return this.shared.aliases.get(name) ?? this.shared.aliases.get(genericBaseTypeName(name));
  }

  public isKnownType(name: string): boolean {
    const type = normalizeTypeText(name);
    const genericBase = genericBaseTypeName(type);
    return isPredeclaredType(type) ||
      this.shared.aliases.has(type) ||
      this.shared.types.has(type) ||
      this.shared.interfaces.has(type) ||
      this.shared.aliases.has(genericBase) ||
      this.shared.types.has(genericBase) ||
      this.shared.interfaces.has(genericBase) ||
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
    return this.shared.methods.get(methodKey(typeName, methodName)) ??
      this.shared.methods.get(methodKey(genericBaseTypeName(typeName), methodName));
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

  public hasBinding(name: string): boolean {
    return this.bindings.has(name) || Boolean(this.parent?.hasBinding(name));
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

  public isUntypedConstant(name: string): boolean {
    const binding = this.bindings.get(name);
    if (binding) return !binding.mutable && binding.typeText === undefined;
    return this.parent?.isUntypedConstant(name) ?? false;
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
const arrayTypeTexts = new WeakMap<RuntimeValue[], string>();
const arrayViews = new WeakMap<RuntimeValue[], { source: RuntimeValue[]; offset: number }>();
let nextObjectMapKeyId = 1;
let nextNonReflexiveMapKeyId = 1;

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
    key = prepareAssignableToType(key, this.keyType, "map key", this.context);
    const entry = this.entries.get(runtimeMapKeyId(key));
    return entry ? entry.value : zeroValueForMapValue(this.valueType);
  }

  public getWithPresence(key: RuntimeValue): [RuntimeValue, boolean] {
    key = prepareAssignableToType(key, this.keyType, "map key", this.context);
    const entry = this.entries.get(runtimeMapKeyId(key));
    return entry ? [entry.value, true] : [zeroValueForMapValue(this.valueType), false];
  }

  public set(key: RuntimeValue, value: RuntimeValue): void {
    key = prepareAssignableToType(key, this.keyType, "map key", this.context);
    assertComparableValue(key, "map key");
    value = prepareAssignableToType(value, this.valueType, "map value", this.context);
    this.entries.set(runtimeMapKeyId(key, { freshNonReflexive: true }), { key, value });
  }

  public delete(key: RuntimeValue): void {
    key = prepareAssignableToType(key, this.keyType, "map key", this.context);
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

export async function evaluatePackageSourceFiles(files: SourceFile[], options: PackageEvaluationOptions = {}): Promise<PackageEvaluationResult> {
  const parsed = frontSourceFilesToAst(files);
  const ast = parsed.ast;
  if (parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return {
      diagnostics: parsed.diagnostics,
      output: []
    };
  }
  if (!ast) {
    return {
      diagnostics: parsed.diagnostics,
      output: []
    };
  }

  const packageName = options.packageName ?? packageNameFromSourceFiles(files) ?? "main";
  const checked = checkGoJuniorSourceFiles(files, {
    ...typeCheckConfig(options),
    packageName,
    packagePath: options.importPath ?? packageName
  });
  if (checked.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return {
      diagnostics: checked.diagnostics,
      output: []
    };
  }

  const context = new EvaluationContext(options);
  try {
    const pkg = await context.scheduler().runRoot(async () => {
      installImports(context, ast);
      for (const declaration of ast.functions) installPackageFunctionDeclaration(context, declaration);
      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      if (statements.length > 0) {
        throw new GoJuniorRuntimeError("package source cannot contain top-level executable statements");
      }
      predeclareTopLevelTypes(declarations, context);
      const declarationCompletion = await executeTopLevelStatements(declarations, context);
      expectNormalCompletion(declarationCompletion, "package declarations");
      await runInitFunctions(ast.functions, context);
      return exportedRuntimePackageObject(packageScopeObjects(checked.pkg), context);
    });
    return {
      diagnostics: checked.diagnostics,
      output: context.output,
      package: pkg,
      packageInfo: checked.pkg
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      diagnostics: [
        ...checked.diagnostics,
        runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
      ],
      output: context.output
    };
  }
}

export async function evaluateSourcePackageGraph(
  specs: SourcePackageSpec[],
  options: SourcePackageGraphOptions = {}
): Promise<SourcePackageGraphEvaluationResult> {
  const evaluator = new SourcePackageGraphEvaluator(specs, options);
  return evaluator.evaluate();
}

class SourcePackageGraphEvaluator {
  private readonly specsByPath = new Map<string, SourcePackageSpec>();
  private readonly importsByPath = new Map<string, string[]>();
  private readonly diagnostics: Diagnostic[] = [];
  private readonly output: string[] = [];
  private readonly packages: Record<string, RuntimeObject>;
  private readonly packageInfos: Record<string, GoTypesPackage>;
  private readonly initialized = new Set<string>();
  private readonly initializedImportPaths: string[] = [];

  public constructor(
    specs: SourcePackageSpec[],
    private readonly options: SourcePackageGraphOptions
  ) {
    this.packages = { ...(options.packages ?? {}) };
    this.packageInfos = { ...(options.packageInfos ?? {}) };
    for (const spec of specs) {
      if (!spec.importPath) {
        this.diagnostics.push(packageGraphDiagnostic(REPL_FILENAME, "source package spec is missing importPath"));
        continue;
      }
      if (this.specsByPath.has(spec.importPath)) {
        this.diagnostics.push(packageGraphDiagnostic(spec.files[0]?.filename ?? REPL_FILENAME, `duplicate source package spec for ${spec.importPath}`));
        continue;
      }
      this.specsByPath.set(spec.importPath, spec);
    }
  }

  public async evaluate(): Promise<SourcePackageGraphEvaluationResult> {
    this.discoverSourcePackages();
    if (this.hasErrors()) return this.result();

    const sortedImportPaths = () => [...this.specsByPath.keys()].sort();
    while (this.initialized.size < this.specsByPath.size) {
      let progressed = false;
      for (const importPath of sortedImportPaths()) {
        if (this.initialized.has(importPath)) continue;
        const imports = this.importsByPath.get(importPath) ?? [];
        const pendingSourceImports = imports.filter((dependency) =>
          this.specsByPath.has(dependency) && !this.initialized.has(dependency)
        );
        if (pendingSourceImports.length > 0) continue;

        await this.initializePackage(importPath);
        progressed = true;
        break;
      }

      if (this.hasErrors()) return this.result();
      if (!progressed) {
        const remaining = sortedImportPaths().filter((importPath) => !this.initialized.has(importPath));
        this.diagnostics.push(packageGraphDiagnostic(
          this.specsByPath.get(remaining[0] ?? "")?.files[0]?.filename ?? REPL_FILENAME,
          `package initialization cycle detected among source packages: ${remaining.join(", ")}`
        ));
        return this.result();
      }
    }

    return this.result();
  }

  private discoverSourcePackages(): void {
    for (const importPath of [...this.specsByPath.keys()].sort()) {
      this.discoverOne(importPath, []);
      if (this.hasErrors()) return;
    }
  }

  private discoverOne(importPath: string, stack: string[]): void {
    if (this.importsByPath.has(importPath) || this.hasErrors()) return;
    if (stack.includes(importPath)) {
      this.diagnostics.push(packageGraphDiagnostic(
        this.specsByPath.get(importPath)?.files[0]?.filename ?? REPL_FILENAME,
        `package import cycle detected: ${[...stack, importPath].join(" -> ")}`
      ));
      return;
    }

    const spec = this.specsByPath.get(importPath);
    if (!spec) return;
    const parsed = frontSourceFilesToAst(spec.files);
    this.diagnostics.push(...parsed.diagnostics);
    if (this.hasErrors()) return;

    const imports = uniqueSortedSourceImports(parsed.ast?.imports.map((imported) => imported.path) ?? []);
    this.importsByPath.set(importPath, imports);
    for (const dependency of imports) {
      if (isRuntimeBuiltinImport(dependency) || this.packages[dependency]) continue;
      if (!this.specsByPath.has(dependency)) {
        const loaded = this.options.sourcePackageProvider?.load(dependency);
        if (loaded) {
          this.specsByPath.set(dependency, { importPath: dependency, files: loaded });
        } else {
          this.diagnostics.push(packageGraphDiagnostic(
            spec.files[0]?.filename ?? REPL_FILENAME,
            `package ${dependency} is not available to gojr package initialization; provide it as a source package`
          ));
          return;
        }
      }
      this.discoverOne(dependency, [...stack, importPath]);
      if (this.hasErrors()) return;
    }
  }

  private async initializePackage(importPath: string): Promise<void> {
    const spec = this.specsByPath.get(importPath);
    if (!spec) return;
    const result = await evaluatePackageSourceFiles(spec.files, {
      ...this.options,
      importPath,
      ...(spec.packageName ? { packageName: spec.packageName } : {}),
      packages: this.packages,
      packageInfos: this.packageInfos
    });
    this.output.push(...result.output);
    this.diagnostics.push(...result.diagnostics);
    if (this.hasErrors()) return;

    const pkg = result.package ?? {};
    this.packages[importPath] = pkg;
    const defaultName = sourcePackageDefaultName(importPath);
    if (this.packages[defaultName] === undefined) {
      this.packages[defaultName] = pkg;
    }
    if (result.packageInfo) {
      this.packageInfos[importPath] = result.packageInfo;
    }
    this.initialized.add(importPath);
    this.initializedImportPaths.push(importPath);
  }

  private hasErrors(): boolean {
    return this.diagnostics.some((diagnostic) => diagnostic.severity === "error");
  }

  private result(): SourcePackageGraphEvaluationResult {
    return {
      diagnostics: this.diagnostics,
      output: this.output,
      packages: this.packages,
      packageInfos: this.packageInfos,
      initializedImportPaths: this.initializedImportPaths
    };
  }
}

function uniqueSortedSourceImports(values: string[]): string[] {
  return [...new Set(values)].sort();
}

function sourcePackageDefaultName(importPath: string): string {
  return importPath.split("/").filter(Boolean).at(-1) ?? importPath;
}

function isRuntimeBuiltinImport(importPath: string): boolean {
  return importPath === "cmp" ||
    importPath === "fmt" ||
    importPath === "testing" ||
    importPath === "unsafe" ||
    importPath === "math" ||
    importPath === "strconv" ||
    importPath === "os";
}

function packageGraphDiagnostic(filename: string, message: string): Diagnostic {
  return {
    filename,
    code: "GOJR_RUNTIME001",
    severity: "error",
    message
  };
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

  const checked = checkGoJuniorSourceFiles(files, typeCheckConfig(options));
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

function packageNameFromSourceFiles(files: SourceFile[]): string | undefined {
  for (const file of files) {
    const match = /^\s*package\s+([A-Za-z_]\w*)/m.exec(file.source);
    if (match?.[1]) return match[1];
  }
  return undefined;
}

export async function evaluateProgram(ast: ProgramAst, options: EvaluationOptions = {}): Promise<EvaluationResult> {
  const context = new EvaluationContext(options);
  context.clearObservedDeps();
  try {
    const result = await context.scheduler().runRoot(async () => {
      installSheets(context, options);
      installImports(context, ast);

      for (const declaration of ast.functions) {
        installFunctionDeclaration(context, declaration);
      }

      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      predeclareTopLevelTypes(declarations, context);
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
    return withObservedDeps(result, context);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return withObservedDeps({
      diagnostics: [
        ...ast.diagnostics,
        runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
      ],
      output: context.output,
      ast
    }, context);
  }
}

function exportedRuntimePackageObject(objects: GoTypesObject[], context: EvaluationContext): RuntimeObject {
  const pkg: RuntimeObject = {};
  for (const object of objects) {
    if (!object.Exported()) continue;
    try {
      pkg[object.Name()] = context.lookup(object.Name());
    } catch {
      // Type-only exports have no runtime value in the interpreter package object.
    }
  }
  return pkg;
}

function packageScopeObjects(pkg: GoTypesPackage): GoTypesObject[] {
  return pkg.Scope().Names().flatMap((name) => {
    if (isGoJuniorSyntheticCheckName(name) || name === "fmt") return [];
    const object = pkg.Scope().Lookup(name);
    if (object === null || object.constructor.name === "PkgName") return [];
    if (object.Pkg() !== pkg) return [];
    return [object];
  });
}

async function testProgram(ast: ProgramAst, baseDiagnostics: Diagnostic[], options: EvaluationOptions): Promise<EvaluationResult> {
  const context = new EvaluationContext(options);
  context.clearObservedDeps();
  const diagnostics = [...baseDiagnostics];
  try {
    const result = await context.scheduler().runRoot(async () => {
      installSheets(context, options);
      installImports(context, ast);

      for (const declaration of ast.functions) {
        installFunctionDeclaration(context, declaration);
      }

      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      predeclareTopLevelTypes(declarations, context);
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
    return withObservedDeps(result, context);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return withObservedDeps({
      diagnostics: [
        ...diagnostics,
        runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
      ],
      output: context.output,
      ast
    }, context);
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
  return sprint(args);
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
  private readonly checkerPackage: GoTypesPackage;
  private sessionSheet: SheetData | undefined;
  private sessionSheets: Record<string, SheetData> | undefined;
  private checkSequence = 0;

  public constructor(private readonly options: EvaluationOptions = {}) {
    this.sessionSheet = options.sheet;
    this.sessionSheets = options.sheets ? { ...options.sheets } : undefined;
    ensureUniverseInitialized();
    this.checkerPackage = NewPackage("main", "main");
    this.context = new EvaluationContext(options);
    installSheets(this.context, options);
  }

  public setSheet(sheet: SheetData): void {
    this.sessionSheet = sheet;
    this.context.declareOrAssignRoot("sheet", new SheetBinding("sheet", sheet), true);
  }

  public setSheets(sheets: Record<string, SheetData>): void {
    this.sessionSheets = {
      ...(this.sessionSheets ?? {}),
      ...sheets
    };
    for (const [name, data] of Object.entries(sheets)) {
      this.context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
    }
  }

  public async evaluate(source: string): Promise<EvaluationResult> {
    this.context.clearObservedDeps();
    const sourceFile = sourceFileFromSource(source, { ...this.options, filename: this.options.filename ?? REPL_FILENAME });
    const parsed = frontSourceToAst(sourceFile.source, sourceFile.filename);
    const ast = parsed.ast;
    const hasError = parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error");
    if (hasError) {
      return withObservedDeps({
        diagnostics: parsed.diagnostics,
        output: [],
        incomplete: diagnosticsLookIncomplete(parsed.diagnostics),
        ...(ast ? { ast } : {})
      }, this.context);
    }
    if (!ast) {
      return withObservedDeps({
        diagnostics: parsed.diagnostics,
        output: []
      }, this.context);
    }

    const preparedRedeclarations = this.prepareFunctionRedeclarations(ast, sourceFile.filename);
    if (preparedRedeclarations.diagnostics.length > 0) {
      return withObservedDeps({
        diagnostics: preparedRedeclarations.diagnostics,
        output: [],
        ast
      }, this.context);
    }

    const checked = this.checkSource(sourceFile);
    if (checked.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
      this.restoreFunctionRedeclarations(preparedRedeclarations.replacements);
      return withObservedDeps({
        diagnostics: checked.diagnostics,
        output: [],
        ast
      }, this.context);
    }

    const outputStart = this.context.output.length;
    try {
      const { declarations, statements } = splitTopLevelDeclarations(ast.body);
      const declarationCompletion = await this.context.scheduler().runRoot(async () => {
        installImports(this.context, ast);

        for (const declaration of ast.functions) {
          installFunctionDeclaration(this.context, declaration);
        }

        predeclareTopLevelTypes(declarations, this.context);
        return executeTopLevelStatements(declarations, this.context);
      });
      expectNormalCompletion(declarationCompletion, "top-level declarations");

      if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
        const value = installedFunctionValue(this.context, ast.functions[0]);
        this.acceptTypeInfo(checked);
        return withObservedDeps({
          diagnostics: ast.diagnostics,
          output: this.context.outputFrom(outputStart),
          ast,
          value
        }, this.context);
      }

      const completion = await this.context.scheduler().runRoot(async () => {
        await runInitFunctions(ast.functions, this.context);
        return executeTopLevelStatements(statements, this.context);
      });
      const result = withObservedDeps(resultFromCompletion(ast, this.context.outputFrom(outputStart), completion), this.context);
      this.acceptTypeInfo(checked);
      return result;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      return withObservedDeps({
        diagnostics: [
          ...ast.diagnostics,
          runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message)
        ],
        output: this.context.outputFrom(outputStart),
        ast
      }, this.context);
    }
  }

  private checkSource(source: SourceFile): GoJuniorCheckResult {
    const checked = checkGoJuniorSourceFiles([ensureTrailingNewlineSourceFile(source)], {
      ...typeCheckConfig(this.currentOptions()),
      packageInstance: this.checkerPackage,
      syntheticFunctionName: `${GOJR_SYNTHETIC_CHECK_PREFIX}_${++this.checkSequence}`
    });
    return checked;
  }

  private acceptTypeInfo(_checked: GoJuniorCheckResult): void {
    // The session owns a persistent go/types.Package, so successful checks have
    // already extended its package scope. Keeping this hook makes the call sites
    // spell out when checked declarations become visible to later evaluations.
  }

  private prepareFunctionRedeclarations(ast: ProgramAst, filename: string): { diagnostics: Diagnostic[]; replacements: PackageScopeReplacement[] } {
    const diagnostics: Diagnostic[] = [];
    const replacements: PackageScopeReplacement[] = [];
    const scope = this.checkerPackage.Scope();
    for (const declaration of ast.functions) {
      if (declaration.receiver || declaration.name === "init") continue;
      const existingObject = scope.Lookup(declaration.name);
      if (existingObject === null) continue;
      const existingValue = this.context.hasBinding(declaration.name) ? this.context.lookup(declaration.name) : undefined;
      if (
        existingValue === undefined ||
        !isGoJuniorFunction(existingValue) ||
        existingValue.signature === undefined ||
        !signaturesCompatible(existingValue.signature, declaration.signature)
      ) {
        diagnostics.push({
          filename,
          code: "GOJR_TYPE001",
          severity: "error",
          message: `cannot redeclare ${declaration.name} with different signature`,
          ...(declaration.span ? { span: declaration.span } : {})
        });
        continue;
      }
      scope.elems.delete(declaration.name);
      replacements.push({ name: declaration.name, object: existingObject });
    }
    return { diagnostics, replacements };
  }

  private restoreFunctionRedeclarations(replacements: PackageScopeReplacement[]): void {
    const scope = this.checkerPackage.Scope();
    for (const replacement of replacements) {
      if (scope.Lookup(replacement.name) === null) {
        scope.insert(replacement.name, replacement.object);
      }
    }
  }

  private currentOptions(): EvaluationOptions {
    return {
      ...this.options,
      ...(this.sessionSheet ? { sheet: this.sessionSheet } : {}),
      ...(this.sessionSheets ? { sheets: this.sessionSheets } : {})
    };
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

export function typeCheckConfig(options: EvaluationOptions): GoJuniorCheckConfig {
  return {
    sheetNamespaces: sheetNamespacesForOptions(options),
    ...(options.packageInfos ? {
      importer: {
        import: (path: string) => options.packageInfos?.[path]
      }
    } : {})
  };
}

function sheetNamespacesForOptions(options: EvaluationOptions): Record<string, GoJuniorSheetNamespace> {
  const namespaces: Record<string, GoJuniorSheetNamespace> = {
    sheet: sheetNamespaceForData(options.sheet ?? options.sheets?.[options.currentSheetName ?? "sheet"] ?? {})
  };
  for (const [name, data] of Object.entries(options.sheets ?? {})) {
    namespaces[name] = sheetNamespaceForData(data);
  }
  return namespaces;
}

function sheetNamespaceForData(data: SheetData): GoJuniorSheetNamespace {
  void data;
  return {};
}

function resultFromCompletion(ast: ProgramAst, output: string[], completion: Completion): EvaluationResult {
  if (completion.kind === "return") {
    const values = completion.values.map(externalizeRuntimeValue);
    return {
      diagnostics: ast.diagnostics,
      output,
      ast,
      values,
      ...(values.length === 1 ? { value: values[0] } : {})
    };
  }

  if (completion.kind !== "normal") throw new GoJuniorRuntimeError(completionErrorMessage(completion));

  return {
    diagnostics: ast.diagnostics,
    output,
    ast,
    ...(completion.value !== undefined ? { value: externalizeRuntimeValue(completion.value) } : {})
  };
}

function withObservedDeps(result: EvaluationResult, context: EvaluationContext): EvaluationResult {
  return {
    ...result,
    observedDeps: context.observedDeps()
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
  context.declareRoot("recover", hostCallable("recover", (_args, context) => {
    return context.recover();
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
    clearValue(args[0] ?? null, context);
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

function installPackageFunctionDeclaration(context: EvaluationContext, declaration: FunctionDecl): void {
  if (declaration.receiver) {
    context.registerMethod(declaration);
    return;
  }
  context.declareOrAssignRoot(
    declaration.name,
    goJuniorFunctionValue(declaration.name, declaration.signature, declaration.body, context.captureScope(), declaration),
    true
  );
}

function installedFunctionValue(context: EvaluationContext, declaration: FunctionDecl): RuntimeValue {
  return declaration.receiver ? functionValue(declaration) : context.lookup(declaration.name);
}

function bindingTypeText(type: TypeNode | string | undefined): string | undefined {
  const text = typeof type === "string" ? type : type?.text;
  const normalized = text ? normalizeTypeText(text) : "";
  return normalized && normalized !== "<missing>" ? normalized : undefined;
}

function resolvedDeclaredTypeText(typeText: string, context: EvaluationContext): string {
  return parseArrayOrSliceTypeText(typeText, context)?.typeText ?? typeText;
}

function installAutomaticImports(context: EvaluationContext): void {
  context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}

function availablePackages(context: EvaluationContext): Record<string, RuntimeObject> {
  return {
    cmp: cmpPackage(),
    fmt: fmtPackage(),
    math: mathPackage(),
    os: osPackage(),
    strconv: strconvPackage(),
    testing: testingPackage(),
    unsafe: unsafePackage(),
    ...context.packages()
  };
}

function cmpPackage(): RuntimeObject {
  return {
    Compare: hostCallable("cmp.Compare", (args) => BigInt(cmpCompare(args[0] ?? null, args[1] ?? null))),
    Less: hostCallable("cmp.Less", (args) => cmpLess(args[0] ?? null, args[1] ?? null)),
    Or: hostCallable("cmp.Or", (args) => args.find((arg) => !isRuntimeZeroValue(arg)) ?? zeroValueLike(args[0] ?? null))
  };
}

function fmtPackage(): RuntimeObject {
  return {
    Printf: hostCallable("fmt.Printf", async (args, context) => {
      const format = await toStringValueAsync(args[0] ?? "", context);
      const text = await sprintfAsync(format, args.slice(1), context);
      context.write(text);
      return BigInt(text.length);
    }),
    Sprintf: hostCallable("fmt.Sprintf", async (args, context) => {
      const format = await toStringValueAsync(args[0] ?? "", context);
      return sprintfAsync(format, args.slice(1), context);
    }),
    Sprint: hostCallable("fmt.Sprint", (args, context) => {
      return sprintAsync(args, context);
    }),
    Sprintln: hostCallable("fmt.Sprintln", (args, context) => {
      return sprintlnAsync(args, context);
    }),
    Println: hostCallable("fmt.Println", async (args, context) => {
      const text = await sprintlnAsync(args, context);
      context.write(text);
      return BigInt(text.length);
    })
  };
}

function mathPackage(): RuntimeObject {
  return {
    NaN: hostCallable("math.NaN", () => Number.NaN),
    Inf: hostCallable("math.Inf", (args) => toFloat(args[0] ?? 1) < 0 ? -Infinity : Infinity),
    IsNaN: hostCallable("math.IsNaN", (args) => Number.isNaN(toFloat(args[0] ?? 0))),
    Float32bits: hostCallable("math.Float32bits", (args) => float32Bits(toFloat(args[0] ?? 0))),
    Float32frombits: hostCallable("math.Float32frombits", (args) => float32FromBits(toBigInt(args[0] ?? 0n))),
    Float64bits: hostCallable("math.Float64bits", (args) => float64Bits(toFloat(args[0] ?? 0))),
    Float64frombits: hostCallable("math.Float64frombits", (args) => float64FromBits(toBigInt(args[0] ?? 0n)))
  };
}

function strconvPackage(): RuntimeObject {
  return {
    Itoa: hostCallable("strconv.Itoa", (args) => toBigInt(args[0] ?? 0n).toString())
  };
}

function testingPackage(): RuntimeObject {
  return {
    Short: hostCallable("testing.Short", () => false),
    Verbose: hostCallable("testing.Verbose", () => false)
  };
}

function unsafePackage(): RuntimeObject {
  return {
    Add: hostCallable("unsafe.Add", (args) => args[0] ?? null),
    Alignof: hostCallable("unsafe.Alignof", (args) => BigInt(runtimeAlignof(args[0] ?? null))),
    Offsetof: hostCallable("unsafe.Offsetof", () => 0n),
    Sizeof: hostCallable("unsafe.Sizeof", (args) => BigInt(runtimeSizeof(args[0] ?? null))),
    Slice: hostCallable("unsafe.Slice", (args) => {
      const length = toNonNegativeLength(args[1] ?? 0n, "unsafe.Slice length");
      if ((args[0] ?? null) === null) {
        if (length === 0) return null;
        throw new GoJuniorRuntimeError("unsafe.Slice: ptr is nil and len is not zero");
      }
      return Array.from({ length }, () => null);
    }),
    SliceData: hostCallable("unsafe.SliceData", (args) => {
      const slice = unwrapNamed(args[0] ?? null);
      if (slice === null) return null;
      if (!Array.isArray(slice)) throw new GoJuniorRuntimeError(`${formatValue(args[0] ?? null)} is not a slice`);
      return slice.length > 0 ? new RuntimePointer(pointerTypeName(slice[0] ?? null), () => slice[0] ?? null, (next) => {
        slice[0] = next;
      }) : null;
    }),
    String: hostCallable("unsafe.String", () => {
      throw new GoJuniorRuntimeError("unsafe.String is not supported by the Go-junior runtime");
    }),
    StringData: hostCallable("unsafe.StringData", () => {
      throw new GoJuniorRuntimeError("unsafe.StringData is not supported by the Go-junior runtime");
    })
  };
}

function osPackage(): RuntimeObject {
  return {
    Exit: hostCallable("os.Exit", (args) => {
      const code = toNumber(args[0] ?? 0);
      throw new GoJuniorRuntimeError(`os.Exit(${code})`);
    }),
    Getenv: hostCallable("os.Getenv", (args, context) => {
      return context.getenv(toStringValue(args[0] ?? ""));
    })
  };
}

function hostCallable(
  name: string,
  call: (args: RuntimeValue[], context: EvaluationContext) => MaybePromise<RuntimeValue>
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
  return goJuniorFunctionValue("<closure>", expression.signature, expression.body, context.captureScope(), undefined, undefined, expression.source);
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

function predeclareTopLevelTypes(statements: Statement[], context: EvaluationContext): void {
  for (const statement of statements) {
    if (statement.kind !== "TypeDecl") continue;
    for (const declaration of statement.declarations) {
      context.registerType(declaration);
    }
  }
}

async function runInitFunctions(functions: FunctionDecl[], context: EvaluationContext): Promise<void> {
  for (const declaration of functions) {
    if (declaration.name !== "init" || declaration.receiver) continue;
    await callRuntime(functionValue(declaration), [], context);
  }
}

async function executeTopLevelStatements(statements: Statement[], context: EvaluationContext): Promise<Completion> {
  const outcome = await context.deferScopeAsync(() => executeStatements(statements, context));
  return outcome === RECOVERED_PANIC ? { kind: "normal" } : outcome;
}

function goJuniorFunctionValue(
  name: string,
  signature: FunctionDecl["signature"],
  body: FunctionDecl["body"],
  closureScope?: Scope,
  declaration?: FunctionDecl,
  boundReceiver?: RuntimeValue,
  source?: string
): GoJuniorFunction {
  const formattedSource = source ?? declaration?.source;
  return {
    kind: "GoJuniorFunction",
    name,
    signature,
    ...(declaration ? { declaration } : {}),
    ...(formattedSource ? { source: formattedSource } : {}),
    async call(args, parentContext) {
      const context = parentContext;
      const invoke = () => context.functionCallAsync(() => context.childScopeAsync(async () => {
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
        const outcome = await context.deferScopeAsync(() => executeBlock(body, context, false));
        const completion = outcome === RECOVERED_PANIC
          ? { kind: "return", values: namedReturnValuesOrZero(signature, context) } satisfies Completion
          : outcome;
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

function namedReturnValuesOrZero(signature: FunctionDecl["signature"], context: EvaluationContext): RuntimeValue[] {
  if (signature.results.length === 0) return [];
  return signature.results.map((result) =>
    result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type, context)
  );
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

function methodExpressionValue(method: MethodDef, receiverTypeText: string): GoJuniorFunction {
  const receiverType: TypeNode = { text: receiverTypeText };
  return {
    kind: "GoJuniorFunction",
    name: `${receiverTypeText}.${method.declaration.name}`,
    signature: {
      parameters: [
        { type: receiverType, variadic: false },
        ...method.declaration.signature.parameters
      ],
      results: method.declaration.signature.results
    },
    declaration: method.declaration,
    ...(method.declaration.source ? { source: method.declaration.source } : {}),
    async call(args, context) {
      const receiver = args[0] ?? null;
      const methodArgs = args.slice(1);
      return callRuntime(boundMethodValue(method, receiver), methodArgs, context);
    }
  };
}

function dynamicMethodExpressionValue(
  receiverTypeText: string,
  methodName: string,
  methodSignature?: FunctionDecl["signature"]
): GoJuniorFunction {
  const receiverType: TypeNode = { text: receiverTypeText };
  return {
    kind: "GoJuniorFunction",
    name: `${receiverTypeText}.${methodName}`,
    ...(methodSignature ? {
      signature: {
        parameters: [
          { type: receiverType, variadic: false },
          ...methodSignature.parameters
        ],
        results: methodSignature.results
      }
    } : {}),
    async call(args, context) {
      const receiver = args[0] ?? null;
      const method = methodForValue(receiver, methodName, context);
      if (!method) throw new GoJuniorRuntimeError(`${formatValue(receiver)} has no method ${methodName}`);
      return callRuntime(boundMethodValue(method.method, method.receiver), args.slice(1), context);
    }
  };
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

function pointerToReceiver(receiver: RuntimeValue, typeName: string): RuntimeValue {
  if (receiver instanceof RuntimePointer) return receiver;
  if (receiver instanceof RuntimeTypedNilValue && runtimeTypeAssignableMatch(receiver.typeName, `*${typeName}`)) return receiver;
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
      let previousConstValues: Expression[] = [];
      let previousConstType = undefined as TypeNode | undefined;
      for (const [index, declaration] of statement.declarations.entries()) {
        const valueExpression = declaration.value ?? previousConstValues[declaration.valueIndex ?? 0];
        const type = declaration.type ?? previousConstType;
        const value = valueExpression
          ? await evaluateConstExpression(valueExpression, BigInt(declaration.iotaIndex ?? index), context)
          : defaultValueForDeclarationType(type, context);
        const typedValue = type ? convertValueToType(value, type.text, context, expressionDeclaredTypeText(valueExpression, context, value)) : value;
        context.declare(
          declaration.name,
          typedValue,
          false,
          type
        );
        if (declaration.value) previousConstValues[declaration.valueIndex ?? 0] = declaration.value;
        if (declaration.type) previousConstType = declaration.type;
      }
      return { kind: "normal" };

    case "VarDecl":
      for (const declaration of statement.declarations) {
        const value = declaration.value
          ? prepareValueForTargetType(
            await evaluateExpression(declaration.value, context),
            declaration.type?.text,
            declaration.value,
            context,
            `variable ${declaration.name}`
          )
          : defaultValueForDeclarationType(declaration.type, context);
        const typeText = declaration.type ??
          (declaration.value
            ? expressionDeclaredTypeText(declaration.value, context, value) ?? inferredTypeText(value)
            : undefined);
        context.declare(
          declaration.name,
          value,
          true,
          typeText
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
  const iotaShadowed = context.hasLocal("iota");
  return context.childScopeAsync(async () => {
    if (!iotaShadowed) context.declare("iota", iotaValue, false, "int64");
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
  const start = await selectValueSwitchClause(statement, switchValue, context);
  if (start === undefined) return { kind: "normal" };

  for (let index = start; index < statement.clauses.length; index += 1) {
    const clause = statement.clauses[index];
    if (!clause) continue;
    const completion = await executeStatements(clause.statements, context);
    if (completion.kind === "fallthrough") {
      continue;
    }
    if (completion.kind === "break" && labelMatches(completion.label, label)) return { kind: "normal" };
    return completion;
  }

  return { kind: "normal" };
}

async function selectValueSwitchClause(
  statement: SwitchStatement,
  switchValue: RuntimeValue,
  context: EvaluationContext
): Promise<number | undefined> {
  let defaultIndex: number | undefined;
  for (const [index, clause] of statement.clauses.entries()) {
    if (clause.default) {
      defaultIndex ??= index;
      continue;
    }
    for (const value of clause.values) {
      if (valueEqual(switchValue, await evaluateExpression(value, context))) return index;
    }
  }
  return defaultIndex;
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
  const start = selectTypeSwitchClause(statement, switchValue, context);
  if (start === undefined) return { kind: "normal" };

  for (let index = start; index < statement.clauses.length; index += 1) {
    const clause = statement.clauses[index];
    if (!clause) continue;
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

function selectTypeSwitchClause(
  statement: SwitchStatement,
  switchValue: RuntimeValue,
  context: EvaluationContext
): number | undefined {
  let defaultIndex: number | undefined;
  for (const [index, clause] of statement.clauses.entries()) {
    if (clause.default) {
      defaultIndex ??= index;
      continue;
    }
    if ((clause.typeValues ?? []).some((type) => runtimeValueMatchesType(switchValue, type.text, context))) return index;
  }
  return defaultIndex;
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
    const keyTypeText = integerRangeKeyTypeText(statement.range.source, source, context);
    const entries = await rangeEntries(source, context);
    for (const [index, value] of entries) {
      const completion = await context.childScopeAsync(async () => {
        if (statement.range?.keyName) {
          if (statement.range.keyName !== "_") {
            if (statement.range.define) context.declare(statement.range.keyName, index, true, keyTypeText ?? inferredTypeText(index));
            else context.assign(statement.range.keyName, index);
          }
        } else if (statement.range?.keyTarget) {
          await assignExpressionTarget(statement.range.keyTarget, index, context);
        }
        if (statement.range?.valueName) {
          if (statement.range.valueName !== "_") {
            if (statement.range.define) context.declare(statement.range.valueName, value, true, inferredTypeText(value));
            else context.assign(statement.range.valueName, value);
          }
        } else if (statement.range?.valueTarget) {
          await assignExpressionTarget(statement.range.valueTarget, value, context);
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
    const args = await evaluateCallArguments(callee, statement.expression.args, statement.expression.spreadLast, context);
    context.pushDefer(() => callRuntime(callee, args, context));
    return;
  }

  context.pushDefer(() => evaluateExpression(statement.expression, context));
}

async function executeGo(statement: GoStatement, context: EvaluationContext): Promise<void> {
  const callee = await evaluateExpression(statement.call.callee, context);
  const args = await evaluateCallArguments(callee, statement.call.args, statement.call.spreadLast, context);
  const goroutineContext = context.fork();
  context.scheduler().go(async () => {
    await callRuntime(callee, args, goroutineContext);
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
  const completion = await context.childScopeAsync(async () => {
    await applySelectResult(preparedClause.clause.comm, selected, context);
    return executeStatements(preparedClause.clause.statements, context);
  });
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
    const source = statement.values.length === statement.targets.length ? statement.values[index] : undefined;
    let value = prepareValueForAssignmentTarget(values[index] ?? null, target, source, context);
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
  declareOrAssignShortVars(statement.names, values, context, statement.values);
}

function declareOrAssignShortVars(names: string[], values: RuntimeValue[], context: EvaluationContext, sources: Expression[] = []): void {
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
      context.assign(name, prepareValueForTargetType(value, context.lookupTypeText(name), sources[index], context, `variable ${name}`));
    } else {
      const typeText = expressionDeclaredTypeText(sources[index], context, value) ?? inferredTypeText(value);
      context.declare(name, value, true, typeText);
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
  const object = unwrapNamed(await evaluateExpression(expression.object, context));
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
      if (expression.literalKind === "string") return goStringFromLiteralRaw(expression.raw);
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

    case "IndexExpression": {
      const object = await evaluateExpression(expression.object, context);
      if ((isRuntimeCallable(object) || isGoJuniorFunction(object)) && isTypeArgumentExpression(expression.index, context)) {
        return object;
      }
      return getIndex(object, await evaluateExpression(expression.index, context));
    }

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

async function evaluateCallArguments(
  callee: RuntimeValue,
  expressions: Expression[],
  spreadLast: boolean | undefined,
  context: EvaluationContext
): Promise<RuntimeValue[]> {
  const raw = await evaluateExpressionList(expressions, context);
  const args = spreadLast ? spreadLastArgument(raw) : raw;
  if (!isGoJuniorFunction(callee) || !callee.signature) return args;

  const prepared = [...args];
  for (const [index, parameter] of callee.signature.parameters.entries()) {
    if (parameter.variadic) {
      for (let argIndex = index; argIndex < prepared.length; argIndex += 1) {
        const expression = spreadLast && argIndex >= expressions.length - 1
          ? expressions[expressions.length - 1]
          : expressions[argIndex];
        prepared[argIndex] = prepareValueForParameterType(
          prepared[argIndex] ?? null,
          parameter.type.text,
          expression,
          context,
          `argument ${argIndex + 1}`
        );
      }
      break;
    }
    if (index >= prepared.length) break;
    prepared[index] = prepareValueForParameterType(prepared[index] ?? null, parameter.type.text, expressions[index], context, `argument ${index + 1}`);
  }
  return prepared;
}

async function evaluateArrayLiteral(expression: ArrayLiteralExpression, context: EvaluationContext): Promise<RuntimeValue[]> {
  const type = parseArrayOrSliceTypeText(expression.type.text, context);
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
  markArrayType(values, type.typeText);
  return values;
}

function makeRuntimeSlice(elementType: string, length: number, capacity: number, context: EvaluationContext, typeText = `[]${elementType}`): RuntimeValue[] {
  if (capacity < length) throw new GoJuniorRuntimeError("slice capacity is smaller than length");
  const backing = Array.from({ length: capacity }, () => defaultValueForTypeText(elementType, context));
  markArrayType(backing, typeText);
  if (length === capacity) {
    arrayCapacities.set(backing, capacity);
    return backing;
  }
  const values = Array.from({ length }, (_item, index) => getArrayElement(backing, index));
  arrayCapacities.set(values, capacity);
  markArrayType(values, typeText);
  markArrayView(values, backing, 0);
  return values;
}

async function evaluateStructLiteral(expression: StructLiteralExpression, context: EvaluationContext): Promise<RuntimeValue> {
  const typeDef = context.typeDef(expression.typeName) ?? parseAnonymousStructTypeText(expression.typeName);
  if (!typeDef) {
    const alias = context.aliasType(expression.typeName);
    const arrayType = alias ? parseArrayOrSliceTypeText(alias, context) : undefined;
    if (arrayType) {
      const values = await evaluateArrayLiteralFields(expression, arrayType, context);
      return new RuntimeNamedValue(expression.typeName, values);
    }
    const mapType = alias ? parseMapTypeText(alias) : undefined;
    if (mapType) {
      return new RuntimeNamedValue(expression.typeName, await evaluateNamedMapLiteralFields(expression, mapType, context));
    }
    throw new GoJuniorRuntimeError(`${expression.typeName} is not a declared struct type`);
  }

  const struct = new RuntimeStruct(typeDef.name);
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

async function evaluateArrayLiteralFields(
  expression: StructLiteralExpression,
  type: NonNullable<ReturnType<typeof parseArrayOrSliceTypeText>>,
  context: EvaluationContext
): Promise<RuntimeValue[]> {
  const values: RuntimeValue[] = [];
  for (const field of expression.fields) {
    if (field.name) throw new GoJuniorRuntimeError(`${expression.typeName} array literal uses unsupported keyed field ${field.name}`);
    values.push(prepareAssignableToType(await evaluateExpression(field.value, context), type.elementType, "array element", context));
  }
  if (type.length !== undefined && values.length > type.length) {
    throw new GoJuniorRuntimeError(`array literal has ${values.length} elements but type ${expression.typeName} has length ${type.length}`);
  }
  if (type.length !== undefined && !type.inferLength) {
    while (values.length < type.length) {
      values.push(defaultValueForTypeText(type.elementType, context));
    }
  }
  markArrayType(values, expression.typeName);
  return values;
}

async function evaluateNamedMapLiteralFields(
  expression: StructLiteralExpression,
  type: NonNullable<ReturnType<typeof parseMapTypeText>>,
  context: EvaluationContext
): Promise<RuntimeMap> {
  const map = new RuntimeMap(type.keyType, type.valueType, context);
  for (const field of expression.fields) {
    if (!field.key) throw new GoJuniorRuntimeError(`${expression.typeName} map literal requires keyed elements`);
    const key = prepareAssignableToType(await evaluateExpression(field.key, context), type.keyType, "map key", context);
    const value = prepareAssignableToType(await evaluateExpression(field.value, context), type.valueType, "map value", context);
    map.set(key, value);
  }
  return map;
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
  let value: RuntimeValue;
  switch (expression.operator) {
    case "+":
      value = numericIdentity(await evaluateExpression(expression.operand, context));
      return materializeValueForExpressionType(value, expression, context);
    case "-":
      value = negateNumber(await evaluateExpression(expression.operand, context));
      return materializeValueForExpressionType(value, expression, context);
    case "!":
      return !toBool(await evaluateExpression(expression.operand, context));
    case "^":
      value = bitwiseComplement(await evaluateExpression(expression.operand, context));
      return materializeValueForExpressionType(value, expression, context);
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
      return materializeValueForExpressionType(addValues(left, right), expression, context);
    case "-":
      return materializeValueForExpressionType(subtractNumbers(left, right), expression, context);
    case "*":
      return materializeValueForExpressionType(multiplyNumbers(left, right), expression, context);
    case "/":
      return materializeValueForExpressionType(divideNumbers(left, right), expression, context);
    case "%":
      return materializeValueForExpressionType(moduloNumbers(left, right), expression, context);
    case "|":
      return materializeValueForExpressionType(bitwiseOr(left, right), expression, context);
    case "^":
      return materializeValueForExpressionType(bitwiseXor(left, right), expression, context);
    case "&":
      return materializeValueForExpressionType(bitwiseAnd(left, right), expression, context);
    case "&^":
      return materializeValueForExpressionType(bitClear(left, right), expression, context);
    case "<<":
      return materializeValueForExpressionType(shiftLeft(left, right), expression, context);
    case ">>":
      return materializeValueForExpressionType(shiftRight(left, right), expression, context);
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
    return await evaluateNew(expression, context);
  }
  const conversionType = conversionTargetType(expression.callee, context);
  if (conversionType) {
    if (expression.args.length !== 1 || expression.spreadLast) {
      throw new GoJuniorRuntimeError(`conversion to ${conversionType} expects exactly one argument`);
    }
    const argument = expression.args[0]!;
    return convertValueToType(
      await evaluateExpression(argument, context),
      conversionType,
      context,
      expressionDeclaredTypeText(argument, context)
    );
  }
  const callee = await evaluateExpression(expression.callee, context);
  const args = await evaluateCallArguments(callee, expression.args, expression.spreadLast, context);
  return callRuntime(callee, args, context);
}

function conversionTargetType(callee: Expression, context: EvaluationContext): string | undefined {
  const typeText = conversionTargetTypeText(callee, context);
  return typeText && runtimeKnowsTypeText(typeText, context) ? typeText : undefined;
}

function conversionTargetTypeText(callee: Expression, context: EvaluationContext): string | undefined {
  if (callee.kind === "TypeExpression") return callee.type.text;
  if (callee.kind === "Identifier") {
    if (context.hasBinding(callee.name)) return undefined;
    return context.isKnownType(callee.name) ? callee.name : undefined;
  }
  if (callee.kind === "UnaryExpression" && callee.operator === "*") {
    const operand = conversionTargetTypeText(callee.operand, context);
    return operand ? `*${operand}` : undefined;
  }
  const typeText = typeArgumentText(callee);
  return typeText && context.isKnownType(typeText) ? typeText : undefined;
}

async function evaluateNew(expression: CallExpression, context: EvaluationContext): Promise<RuntimeValue> {
  if (expression.spreadLast || expression.args.length !== 1) throw new GoJuniorRuntimeError("new expects exactly one argument");
  const argument = expression.args[0]!;
  const typeArgument = newTypeArgumentText(argument, context);
  const evaluated = typeArgument ? undefined : await evaluateExpression(argument, context);
  const typeText = typeArgument ?? expressionDeclaredTypeText(argument, context, evaluated);
  if (!typeText) throw new GoJuniorRuntimeError("new could not infer argument type");
  let value = typeArgument
    ? defaultValueForTypeText(typeText, context)
    : copyValueForNew(prepareAssignableToType(evaluated!, typeText, "new argument", context), typeText, context);
  return new RuntimePointer(normalizeTypeText(typeText), () => value, (next) => {
    assertAssignableToType(next, typeText, `*${typeText}`, context);
    value = next;
  });
}

function newTypeArgumentText(expression: Expression, context: EvaluationContext): string | undefined {
  if (expression.kind === "TypeExpression") return expression.type.text;
  if (expression.kind === "Identifier") {
    if (context.hasBinding(expression.name)) return undefined;
    return context.isKnownType(expression.name) ? expression.name : undefined;
  }
  const typeText = typeArgumentText(expression);
  return typeText && context.isKnownType(typeText) ? typeText : undefined;
}

function copyValueForNew(value: RuntimeValue, typeText: string, context: EvaluationContext): RuntimeValue {
  const type = normalizeTypeText(typeText);
  if (value instanceof RuntimeNamedValue) {
    const underlying = context.aliasType(value.typeName) ?? type;
    return new RuntimeNamedValue(value.typeName, copyValueForNew(value.value, underlying, context));
  }
  if (value instanceof RuntimeInterfaceValue) {
    return new RuntimeInterfaceValue(value.interfaceType, value.value);
  }
  if (value instanceof RuntimeStruct) {
    const struct = new RuntimeStruct(value.typeName);
    const typeDef = context.typeDef(value.typeName);
    for (const [fieldName, fieldValue] of value.orderedFields()) {
      const fieldType = typeDef?.fields.find((field) => field.name === fieldName)?.type.text;
      struct.set(fieldName, fieldType ? copyValueForNew(fieldValue, fieldType, context) : fieldValue);
    }
    return struct;
  }
  const arrayType = parseArrayOrSliceTypeText(type, context);
  if (arrayType?.length !== undefined && !arrayType.inferLength && Array.isArray(value)) {
    const copy = value.map((item) => copyValueForNew(item, arrayType.elementType, context));
    arrayCapacities.set(copy, sliceCapacity(value));
    markArrayType(copy, arrayType.typeText);
    return copy;
  }
  return value;
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
  const arrayType = parseArrayOrSliceTypeText(typeText, context);
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
    return makeRuntimeSlice(arrayType.elementType, length, capacity, context, arrayType.typeText);
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
  const typeText = Array.isArray(target) ? arrayTypeTexts.get(target) : undefined;
  if (typeText) markArrayType(appended, typeText);
  return appended;
}

function clearValue(target: RuntimeValue, context?: EvaluationContext): void {
  target = unwrapNamed(target);
  if (target instanceof RuntimeMap) {
    target.clear();
    return;
  }
  if (Array.isArray(target)) {
    const type = parseArrayOrSliceTypeText(arrayTypeTexts.get(target) ?? "", context);
    for (let index = 0; index < target.length; index += 1) {
      const zero = type ? defaultValueForTypeText(type.elementType, context) : null;
      setArrayElement(target, index, zero);
    }
    return;
  }
  throw new GoJuniorRuntimeError("clear expects a map or slice");
}

function copyValues(target: RuntimeValue, source: RuntimeValue): number {
  target = unwrapNamed(target);
  source = unwrapNamed(source);
  if (!Array.isArray(target)) throw new GoJuniorRuntimeError("copy destination must be a slice");
  const sourceValues = isRuntimeString(source)
    ? [...goStringBytes(source)].map((value) => BigInt(value))
    : source;
  if (!Array.isArray(sourceValues)) throw new GoJuniorRuntimeError("copy source must be a slice or string");
  const count = Math.min(target.length, sourceValues.length);
  for (let index = 0; index < count; index += 1) {
    const value = isRuntimeString(source) ? sourceValues[index] ?? null : getArrayElement(sourceValues, index);
    setArrayElement(target, index, value);
  }
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
  const methodExpression = methodExpressionForSelector(expression, context);
  if (methodExpression) return methodExpression;

  const object = await evaluateExpression(expression.object, context);
  if (object instanceof SheetBinding) {
    context.recordSheetCellRead(object.name, expression.field);
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

function methodExpressionForSelector(expression: SelectorExpression, context: EvaluationContext): GoJuniorFunction | undefined {
  const receiver = methodExpressionReceiverType(expression.object, context);
  if (!receiver) return undefined;
  const method = context.methodFor(receiver.baseType, expression.field);
  if (method) {
    if (method.pointerReceiver && !receiver.pointer) return undefined;
    return methodExpressionValue(method, receiver.argumentType);
  }

  const interfaceMethod = interfaceMethodForType(receiver.baseType, expression.field, context);
  if (interfaceMethod) return dynamicMethodExpressionValue(receiver.argumentType, expression.field, interfaceMethod.signature);
  if (promotedMethodAvailableForType(receiver.baseType, expression.field, context, receiver.pointer)) {
    return dynamicMethodExpressionValue(receiver.argumentType, expression.field);
  }
  return undefined;
}

function methodExpressionReceiverType(
  expression: Expression,
  context: EvaluationContext
): { baseType: string; argumentType: string; pointer: boolean } | undefined {
  if (expression.kind === "Identifier") {
    if (context.hasBinding(expression.name) || !context.isKnownType(expression.name)) return undefined;
    const type = normalizeTypeText(expression.name);
    return { baseType: type, argumentType: type, pointer: false };
  }
  if (expression.kind === "TypeExpression") {
    return methodExpressionReceiverTypeText(expression.type.text, context);
  }
  if (expression.kind === "UnaryExpression" && expression.operator === "*") {
    const base = methodExpressionReceiverType(expression.operand, context);
    if (!base) return undefined;
    return { baseType: base.baseType, argumentType: `*${base.baseType}`, pointer: true };
  }
  return undefined;
}

function methodExpressionReceiverTypeText(
  typeText: string,
  context: EvaluationContext
): { baseType: string; argumentType: string; pointer: boolean } | undefined {
  const type = normalizeTypeText(typeText);
  if (!type) return undefined;
  if (type.startsWith("*")) {
    const base = type.slice(1);
    if (!runtimeKnowsTypeText(base, context)) return undefined;
    return { baseType: base, argumentType: type, pointer: true };
  }
  if (!runtimeKnowsTypeText(type, context)) return undefined;
  return { baseType: type, argumentType: type, pointer: false };
}

function runtimeKnowsTypeText(type: string, context: EvaluationContext): boolean {
  return context.isKnownType(type) ||
    parseAnonymousStructTypeText(type) !== undefined ||
    parseAnonymousInterfaceTypeText(type) !== undefined;
}

function interfaceMethodForType(
  typeText: string,
  methodName: string,
  context: EvaluationContext
): InterfaceMethodDecl | undefined {
  const interfaceType = interfaceTarget(typeText, context);
  return interfaceType?.methods.find((method) => method.name === methodName);
}

function promotedMethodAvailableForType(
  typeText: string,
  methodName: string,
  context: EvaluationContext,
  addressable = false,
  seen = new Set<string>()
): boolean {
  const receiver = normalizeReceiverType(typeText);
  const baseType = receiver.baseType;
  if (seen.has(baseType)) return false;
  seen.add(baseType);
  const typeDef = context.typeDef(baseType) ?? parseAnonymousStructTypeText(baseType);
  if (!typeDef) return false;
  for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
    const embedded = normalizeReceiverType(field.type.text);
    const direct = context.methodFor(embedded.baseType, methodName);
    const embeddedAddressable = addressable || receiver.pointer || embedded.pointer;
    if (direct && (!direct.pointerReceiver || embeddedAddressable)) return true;
    if (promotedMethodAvailableForType(embedded.baseType, methodName, context, embeddedAddressable, seen)) return true;
  }
  return false;
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
  context.recordSheetRangeRead(sheet.name, expression.start.field, expression.endCell);
  return sheet.range(expression.start.field, expression.endCell);
}

function getIndex(object: RuntimeValue, index: RuntimeValue): RuntimeValue {
  object = unwrapNamed(object);
  if (object instanceof RuntimeMap) return object.get(index);
  const numericIndex = toNumber(index);
  if (isRuntimeString(object)) return goStringByteAt(object, numericIndex);
  if (isRuntimeObject(object)) return object[String(index)] ?? null;
  if (object instanceof RuntimePointer) object = object.get();
  object = unwrapNamed(object);
  if (Array.isArray(object)) return getArrayElement(object, numericIndex);
  if (isRuntimeString(object)) return goStringByteAt(object, numericIndex);
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
}

async function setIndex(expression: IndexExpression, value: RuntimeValue, context: EvaluationContext): Promise<void> {
  let object = await evaluateExpression(expression.object, context);
  const index = await evaluateExpression(expression.index, context);
  object = unwrapNamed(object);
  if (object instanceof RuntimeMap) {
    object.set(index, value);
    return;
  }
  if (isRuntimeObject(object)) {
    object[String(index)] = value;
    return;
  }
  if (object instanceof RuntimePointer) {
    const target = unwrapNamed(object.get());
    if (!Array.isArray(target)) throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
    setArrayElement(target, toNumber(index), value);
    return;
  }
  const numericIndex = toNumber(index);
  if (Array.isArray(object)) {
    setArrayElement(object, numericIndex, value);
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
    const typeName = context.lookupTypeText(expression.name) ?? pointerTypeName(value);
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
    const typeName = pointerTypeName(getArrayElement(object, numericIndex));
    return new RuntimePointer(typeName, () => getArrayElement(object, numericIndex), (next) => {
      setArrayElement(object, numericIndex, next);
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
  const typeDef = structTypeDefFor(struct, context);
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
  const typeDef = structTypeDefFor(struct, context);
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

function structTypeDefFor(struct: RuntimeStruct, context: EvaluationContext): StructTypeDef | undefined {
  return context.typeDef(struct.typeName) ?? parseAnonymousStructTypeText(struct.typeName);
}

function methodForValue(value: RuntimeValue, methodName: string, context: EvaluationContext): MethodLookup | undefined {
  if (value instanceof RuntimeInterfaceValue) {
    if (value.value === null) return undefined;
    return methodForValue(value.value, methodName, context);
  }
  const addressable = value instanceof RuntimePointer;
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
  if (actual instanceof RuntimeStruct) return promotedMethodForStruct(actual, methodName, context, addressable);
  return undefined;
}

function promotedMethodForStruct(
  struct: RuntimeStruct,
  methodName: string,
  context: EvaluationContext,
  addressable = false,
  seen = new Set<string>()
): MethodLookup | undefined {
  if (seen.has(struct.typeName)) return undefined;
  seen.add(struct.typeName);
  const typeDef = structTypeDefFor(struct, context);
  if (!typeDef) return undefined;

  const matches: MethodLookup[] = [];
  for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
    const receiver = struct.get(field.name);
    if (receiver === undefined || receiver === null) continue;
    if (receiver instanceof RuntimeInterfaceValue) {
      const promoted = methodForValue(receiver, methodName, context);
      if (promoted) {
        matches.push(promoted);
        continue;
      }
    }
    const embeddedReceiver = normalizeReceiverType(field.type.text);
    const receiverType = receiverTypeName(receiver) ?? embeddedReceiver.baseType;
    if (receiverType) {
      const direct = context.methodFor(receiverType, methodName);
      if (direct) {
        if (!direct.pointerReceiver || receiver instanceof RuntimePointer || isTypedNilPointer(receiver)) {
          matches.push({ method: direct, receiver });
          continue;
        }
        if (addressable) {
          matches.push({ method: direct, receiver: pointerToStructField(struct, field, context) });
        }
        continue;
      }
    }
    const embedded = structFromValue(receiver);
    if (!embedded) continue;
    const embeddedAddressable = addressable || receiver instanceof RuntimePointer || isTypedNilPointer(receiver);
    const promoted = promotedMethodForStruct(embedded, methodName, context, embeddedAddressable, seen);
    if (promoted) matches.push(promoted);
  }
  if (matches.length > 1) throw new GoJuniorRuntimeError(`ambiguous promoted method ${methodName}`);
  return matches[0];
}

function pointerToStructField(struct: RuntimeStruct, field: StructFieldDecl, context: EvaluationContext): RuntimePointer {
  const typeName = normalizeReceiverType(field.type.text).baseType;
  return new RuntimePointer(typeName, () => struct.get(field.name) ?? defaultValueForDeclarationType(field.type, context), (next) => {
    struct.set(field.name, prepareAssignableToType(next, field.type.text, `field ${field.name}`, context));
  });
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
  if (isRuntimeString(actual)) return "string";
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
  if (isRuntimeString(value)) return "string";
  if (typeof value === "boolean") return "bool";
  if (Array.isArray(value)) return arrayTypeTexts.get(value);
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return `map[${value.keyType}]${value.valueType}`;
  if (value instanceof RuntimeStruct) return value.typeName;
  if (value instanceof RuntimePointer) return `*${value.typeName}`;
  return undefined;
}

function getSlice(object: RuntimeValue, start: RuntimeValue | undefined, end: RuntimeValue | undefined, max: RuntimeValue | undefined): RuntimeValue {
  if (object instanceof RuntimePointer) object = object.get();
  object = unwrapNamed(object);
  const startIndex = start === undefined ? undefined : toNumber(start);
  const endIndex = end === undefined ? undefined : toNumber(end);
  const maxIndex = max === undefined ? undefined : toNumber(max);
  if (Array.isArray(object)) {
    const low = startIndex ?? 0;
    const high = endIndex ?? object.length;
    const capEnd = maxIndex ?? sliceCapacity(object);
    if (capEnd < high) throw new GoJuniorRuntimeError("slice max is smaller than high bound");
    const result = Array.from({ length: Math.max(0, high - low) }, (_item, index) => getArrayElement(object, low + index));
    arrayCapacities.set(result, Math.max(0, capEnd - low));
    const view = arrayViews.get(object);
    markArrayView(result, view?.source ?? object, (view?.offset ?? 0) + low);
    const sourceType = parseArrayOrSliceTypeText(arrayTypeTexts.get(object) ?? "");
    if (sourceType) markArrayType(result, `[]${sourceType.elementType}`);
    return result;
  }
  if (max !== undefined) throw new GoJuniorRuntimeError("three-index slicing is only supported for arrays and slices");
  if (isRuntimeString(object)) return goStringSlice(object, startIndex, endIndex);
  throw new GoJuniorRuntimeError(`${formatValue(object)} is not sliceable`);
}

function defaultValueForDeclarationType(type: TypeNode | undefined, context?: EvaluationContext): RuntimeValue {
  return defaultValueForTypeText(type?.text ?? "", context);
}

function defaultValueForTypeText(typeText: string, context?: EvaluationContext): RuntimeValue {
  const mapType = parseMapTypeText(typeText);
  if (mapType) return new RuntimeMap(mapType.keyType, mapType.valueType, context);
  if (parseChanTypeText(typeText)) return null;
  const arrayType = parseArrayOrSliceTypeText(typeText, context);
  if (arrayType) {
    const values = arrayType.length === undefined || arrayType.inferLength
      ? []
      : Array.from({ length: arrayType.length }, () => defaultValueForTypeText(arrayType.elementType, context));
    markArrayType(values, arrayType.typeText);
    return values;
  }
  const type = normalizeTypeText(typeText);
  const interfaceType = interfaceTarget(type, context);
  if (interfaceType) return new RuntimeInterfaceValue(type, null);
  if (type.startsWith("*")) return new RuntimeTypedNilValue(type);
  const anonymousStruct = parseAnonymousStructTypeText(typeText);
  if (anonymousStruct) {
    const struct = new RuntimeStruct(anonymousStruct.name);
    for (const field of anonymousStruct.fields) {
      struct.set(field.name, defaultValueForTypeText(field.type.text, context));
    }
    return struct;
  }
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

function convertValueToType(
  value: RuntimeValue,
  typeText: string,
  context: EvaluationContext,
  sourceTypeText?: string
): RuntimeValue {
  const type = normalizeTypeText(typeText);
  const targetInterface = interfaceTarget(type, context);
  if (targetInterface) return prepareInterfaceAssignment(value, type, targetInterface, "conversion", context, sourceTypeText);
  const alias = context.aliasType(type);
  if (alias && alias !== type) {
    return new RuntimeNamedValue(type, convertValueToType(value, alias, context, sourceTypeText));
  }
  const actual = unwrapNamed(value);
  if (type === "any" || type === "interface{}") return actual;
  if (type === "bool") {
    if (typeof actual !== "boolean") throwTypeError(actual, type, "conversion");
    return actual;
  }
  if (isIntegerType(type)) {
    if (typeof actual === "bigint") {
      return convertIntegerToType(actual, type);
    }
    if (typeof actual === "number") {
      const converted = BigInt(Math.trunc(actual));
      return convertIntegerToType(converted, type);
    }
    if (isComplexValue(actual) && actual.imag === 0) {
      const converted = BigInt(Math.trunc(actual.real));
      return convertIntegerToType(converted, type);
    }
    throwTypeError(actual, type, "conversion");
  }
  if (isFloatType(type)) {
    if (isComplexValue(actual)) {
      if (actual.imag !== 0) throwTypeError(actual, type, "conversion");
      return roundFloatForType(actual.real, type);
    }
    return roundFloatForType(toFloat(actual), type);
  }
  if (isComplexType(type)) {
    return roundComplexForType(toComplex(actual), type);
  }
  if (type === "string") {
    if (isRuntimeString(actual)) return ensureRuntimeGoString(actual);
    if (typeof actual === "bigint") return goStringFromRune(actual);
    if (Array.isArray(actual)) {
      return stringFromIntegerArray(actual, sourceTypeText, context);
    }
    throwTypeError(actual, type, "conversion");
  }
  const arrayType = parseArrayOrSliceTypeText(type, context);
  if (arrayType) {
    if (actual === null && arrayType.length === undefined && !arrayType.inferLength) {
      return new RuntimeTypedNilValue(type);
    }
    if (isRuntimeString(actual)) {
      const values = integerArrayFromString(actual, type, context);
      if (values) {
        markArrayType(values, arrayType.typeText);
        return values;
      }
    }
    if (!Array.isArray(actual)) throwTypeError(actual, type, "conversion");
    if (arrayType.length !== undefined && !arrayType.inferLength && actual.length !== arrayType.length) {
      throwTypeError(actual, type, "conversion");
    }
    for (const item of actual) {
      assertAssignableToType(item, arrayType.elementType, "conversion element", context);
    }
    markArrayType(actual, arrayType.typeText);
    return actual;
  }
  if (context.typeDef(type) && actual instanceof RuntimeStruct && actual.typeName === type) return actual;
  if (actual === null && type.startsWith("*")) return new RuntimeTypedNilValue(type);
  if (actual === null && isNilAssignableType(type)) return null;
  throw new GoJuniorRuntimeError(`unsupported conversion to ${type}`);
}

function integerArrayFromString(source: RuntimeValue, targetTypeText: string, context?: EvaluationContext): RuntimeValue[] | undefined {
  const arrayType = parseArrayOrSliceTypeText(targetTypeText, context);
  if (!arrayType || arrayType.length !== undefined) return undefined;
  const elementType = integerSequenceElementType(targetTypeText, context);
  if (elementType === "byte") {
    return [...goStringBytes(source)].map((value) => BigInt(value));
  }
  if (elementType === "rune") {
    return goStringRuneEntries(source).map((entry) => entry.rune);
  }
  return undefined;
}

function stringFromIntegerArray(values: RuntimeValue[], sourceTypeText: string | undefined, context?: EvaluationContext): RuntimeGoString {
  const elementType = integerSequenceElementType(sourceTypeText, context);
  if (elementType !== "byte" && elementType !== "rune") {
    throwTypeError(values, "string", "conversion");
  }
  const integers = values.map((value) => {
    const actual = unwrapNamed(value);
    if (typeof actual !== "bigint") throwTypeError(value, "string", "conversion");
    return actual;
  });
  if (elementType === "byte") {
    const bytes = integers.map((value) => Number(BigInt.asUintN(8, value)));
    return new RuntimeGoString(new Uint8Array(bytes));
  }
  return goStringFromRunes(integers);
}

function integerSequenceElementType(typeText: string | undefined, context?: EvaluationContext): "byte" | "rune" | undefined {
  if (!typeText) return undefined;
  let type = normalizeTypeText(typeText);
  const seen = new Set<string>();
  while (context?.aliasType(type) && !seen.has(type)) {
    seen.add(type);
    type = normalizeTypeText(context.aliasType(type) ?? type);
  }
  const arrayType = parseArrayOrSliceTypeText(type, context);
  const elementType = normalizeTypeText(arrayType?.elementType ?? "");
  if (elementType === "byte" || elementType === "uint8") return "byte";
  if (elementType === "rune" || elementType === "int32") return "rune";
  return undefined;
}

function goStringFromRune(value: bigint): RuntimeGoString {
  const codePoint = Number(value);
  if (!Number.isInteger(codePoint) || codePoint < 0 || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff)) {
    return RuntimeGoString.fromUtf8Text("\ufffd");
  }
  return RuntimeGoString.fromUtf8Text(String.fromCodePoint(codePoint));
}

function prepareValueForAssignmentTarget(
  value: RuntimeValue,
  target: Expression,
  source: Expression | undefined,
  context: EvaluationContext
): RuntimeValue {
  if (target.kind === "Identifier") {
    return prepareValueForTargetType(value, context.lookupTypeText(target.name), source, context, `variable ${target.name}`);
  }
  if (target.kind === "IndexExpression") {
    const objectType = expressionDeclaredTypeText(target.object, context);
    const targetType = indexResultTypeText(objectType);
    return prepareValueForTargetType(value, targetType, source, context, "index assignment");
  }
  return value;
}

function prepareValueForTargetType(
  value: RuntimeValue,
  targetTypeText: string | undefined,
  source: Expression | undefined,
  context: EvaluationContext,
  role = "interface assignment"
): RuntimeValue {
  const targetType = targetTypeText ? normalizeTypeText(targetTypeText) : "";
  const interfaceType = targetType ? interfaceTarget(targetType, context) : undefined;
  if (!interfaceType) return value;
  const dynamicType = expressionDynamicTypeText(source, context, value);
  return prepareInterfaceAssignment(value, targetType, interfaceType, role, context, dynamicType);
}

function prepareValueForParameterType(
  value: RuntimeValue,
  targetTypeText: string | undefined,
  source: Expression | undefined,
  context: EvaluationContext,
  role: string
): RuntimeValue {
  const targetType = targetTypeText ? normalizeTypeText(targetTypeText) : "";
  if (!targetType) return value;
  if (interfaceTarget(targetType, context)) {
    return prepareValueForTargetType(value, targetType, source, context, role);
  }
  return prepareAssignableToType(value, targetType, role, context);
}

function expressionDynamicTypeText(
  expression: Expression | undefined,
  context: EvaluationContext,
  value?: RuntimeValue
): string | undefined {
  if (!expression) return inferredConcreteDynamicType(value);
  switch (expression.kind) {
    case "Identifier":
      if (expression.name === "nil") return undefined;
      if (context.isUntypedConstant(expression.name)) return undefined;
      return context.lookupTypeText(expression.name) ?? inferredConcreteDynamicType(value);
    case "Literal":
      switch (expression.literalKind) {
        case "int": return "int";
        case "float": return "float64";
        case "imag": return "complex128";
        case "rune": return "rune";
        case "string": return "string";
        case "bool": return "bool";
        case "nil": return undefined;
      }
      return inferredConcreteDynamicType(value);
    case "CallExpression": {
      if (expression.callee.kind === "Identifier" && expression.callee.name === "make" && expression.args[0]?.kind === "TypeExpression") {
        return normalizeTypeText(expression.args[0].type.text);
      }
      const conversionType = conversionTargetType(expression.callee, context);
      return conversionType ? normalizeTypeText(conversionType) : inferredConcreteDynamicType(value);
    }
    case "TypeAssertionExpression":
      return normalizeTypeText(expression.type.text);
    case "UnaryExpression":
      if (expression.operator === "&") {
        const operandType = expressionDynamicTypeText(expression.operand, context);
        return operandType ? `*${operandType}` : inferredConcreteDynamicType(value);
      }
      if (expression.operator === "<-") {
        const operandType = expressionDynamicTypeText(expression.operand, context);
        return channelElementTypeText(operandType) ?? inferredConcreteDynamicType(value);
      }
      if (expression.operator !== "!") return expressionDeclaredTypeText(expression, context, value) ?? inferredConcreteDynamicType(value);
      return inferredConcreteDynamicType(value);
    case "BinaryExpression":
      return binaryExpressionTypeText(expression, context, value) ?? inferredConcreteDynamicType(value);
    default:
      return inferredConcreteDynamicType(value);
  }
}

function expressionDeclaredTypeText(
  expression: Expression | undefined,
  context: EvaluationContext,
  value?: RuntimeValue
): string | undefined {
  if (!expression) return inferredConcreteDynamicType(value);
  switch (expression.kind) {
    case "Identifier":
      if (expression.name === "nil") return undefined;
      return context.lookupTypeText(expression.name) ?? inferredConcreteDynamicType(value);
    case "ArrayLiteralExpression":
      return expression.type.text;
    case "MapLiteralExpression":
      return `map[${expression.keyType.text}]${expression.valueType.text}`;
    case "StructLiteralExpression":
      return expression.typeName;
    case "SliceExpression": {
      const objectType = expressionDeclaredTypeText(expression.object, context);
      const sliced = sliceResultTypeText(objectType);
      return sliced ?? inferredConcreteDynamicType(value);
    }
    case "IndexExpression": {
      const objectType = expressionDeclaredTypeText(expression.object, context);
      const indexed = indexResultTypeText(objectType);
      return indexed ?? inferredConcreteDynamicType(value);
    }
    case "UnaryExpression":
      if (expression.operator === "*") {
        const operandType = expressionDeclaredTypeText(expression.operand, context);
        return operandType?.startsWith("*") ? operandType.slice(1) : inferredConcreteDynamicType(value);
      }
      if (expression.operator === "&") {
        const operandType = expressionDeclaredTypeText(expression.operand, context);
        return operandType ? `*${operandType}` : inferredConcreteDynamicType(value);
      }
      if (expression.operator === "!") return "bool";
      if (expression.operator === "<-") {
        const operandType = expressionDeclaredTypeText(expression.operand, context);
        return channelElementTypeText(operandType) ?? inferredConcreteDynamicType(value);
      }
      return expressionDeclaredTypeText(expression.operand, context, value) ?? inferredConcreteDynamicType(value);
    case "BinaryExpression":
      return binaryExpressionTypeText(expression, context, value);
    case "CallExpression": {
      const conversionType = conversionTargetType(expression.callee, context);
      return conversionType ? normalizeTypeText(conversionType) : inferredConcreteDynamicType(value);
    }
    case "TypeAssertionExpression":
      return normalizeTypeText(expression.type.text);
    default:
      return inferredConcreteDynamicType(value);
  }
}

function materializeValueForExpressionType(value: RuntimeValue, expression: Expression, context: EvaluationContext): RuntimeValue {
  if (isUntypedConstantExpression(expression)) return value;
  return materializeValueForType(value, expressionDeclaredTypeText(expression, context, value), context);
}

function materializeValueForType(value: RuntimeValue, typeText: string | undefined, context?: EvaluationContext): RuntimeValue {
  const type = normalizeTypeText(typeText ?? "");
  if (!type) return value;
  const base = numericAliasBaseType(type, context);
  if (isIntegerType(base)) {
    const actual = unwrapNamed(value);
    const converted = typeof actual === "bigint"
      ? convertIntegerToType(actual, base)
      : typeof actual === "number" && Number.isInteger(actual)
        ? convertIntegerToType(BigInt(actual), base)
        : undefined;
    if (converted !== undefined) return namedNumericResult(type, base, converted, context);
  }
  if (isFloatType(base)) {
    const rounded = roundFloatForType(toFloat(value), base);
    return namedNumericResult(type, base, rounded, context);
  }
  if (isComplexType(base)) {
    const rounded = roundComplexForType(toComplex(value), base);
    return namedNumericResult(type, base, rounded, context);
  }
  return value;
}

function namedNumericResult(type: string, base: string, value: RuntimeValue, context?: EvaluationContext): RuntimeValue {
  void type;
  void base;
  void context;
  return value;
}

function binaryExpressionTypeText(
  expression: BinaryExpression,
  context: EvaluationContext,
  value?: RuntimeValue
): string | undefined {
  if (expression.operator === "&&" || expression.operator === "||" || comparisonOperators.has(expression.operator)) return "bool";

  const leftType = expressionDeclaredTypeText(expression.left, context);
  const rightType = expressionDeclaredTypeText(expression.right, context);
  const leftUntyped = isUntypedConstantExpression(expression.left);
  const rightUntyped = isUntypedConstantExpression(expression.right);

  if (leftUntyped && rightUntyped) return undefined;
  if (expression.operator === "<<" || expression.operator === ">>") return leftUntyped ? undefined : leftType;

  if (leftType && rightUntyped && isNumericTypeText(leftType, context)) return leftType;
  if (rightType && leftUntyped && isNumericTypeText(rightType, context)) return rightType;
  if (leftType && rightType && normalizeTypeText(leftType) === normalizeTypeText(rightType)) return leftType;
  return undefined;
}

const comparisonOperators = new Set<BinaryExpression["operator"]>(["==", "!=", "<", "<=", ">", ">="]);

function isUntypedConstantExpression(expression: Expression): boolean {
  switch (expression.kind) {
    case "Literal":
      return expression.literalKind !== "string" || expression.raw.startsWith("\"") || expression.raw.startsWith("`");
    case "UnaryExpression":
      return expression.operator !== "<-" && expression.operator !== "&" && expression.operator !== "*" &&
        isUntypedConstantExpression(expression.operand);
    case "BinaryExpression":
      return isUntypedConstantExpression(expression.left) && isUntypedConstantExpression(expression.right);
    default:
      return false;
  }
}

function isNumericTypeText(typeText: string, context?: EvaluationContext): boolean {
  const base = numericAliasBaseType(typeText, context);
  return isIntegerType(base) || isFloatType(base) || isComplexType(base);
}

function numericAliasBaseType(typeText: string, context?: EvaluationContext): string {
  let type = normalizeTypeText(typeText);
  const seen = new Set<string>();
  while (context?.aliasType(type) && !seen.has(type)) {
    seen.add(type);
    type = normalizeTypeText(context.aliasType(type) ?? type);
  }
  return type;
}

function channelElementTypeText(typeText: string | undefined): string | undefined {
  if (!typeText) return undefined;
  return parseChanTypeText(typeText)?.elementType;
}

function roundFloatForType(value: number, typeText: string): number {
  return normalizeTypeText(typeText) === "float32" ? Math.fround(value) : value;
}

function roundComplexForType(value: ComplexLiteralValue, typeText: string): ComplexLiteralValue {
  return normalizeTypeText(typeText) === "complex64"
    ? complexValue(Math.fround(value.real), Math.fround(value.imag))
    : value;
}

function sliceResultTypeText(typeText: string | undefined): string | undefined {
  const type = normalizeTypeText(typeText ?? "");
  if (!type) return undefined;
  if (type === "string") return "string";
  const dereferenced = type.startsWith("*") ? type.slice(1) : type;
  const arrayType = parseArrayOrSliceTypeText(dereferenced);
  if (!arrayType) return undefined;
  return `[]${arrayType.elementType}`;
}

function indexResultTypeText(typeText: string | undefined): string | undefined {
  const type = normalizeTypeText(typeText ?? "");
  if (!type) return undefined;
  if (type === "string") return "byte";
  const dereferenced = type.startsWith("*") ? type.slice(1) : type;
  const arrayType = parseArrayOrSliceTypeText(dereferenced);
  if (arrayType) return arrayType.elementType;
  const mapType = parseMapTypeText(dereferenced);
  if (mapType) return mapType.valueType;
  return undefined;
}

function inferredConcreteDynamicType(value: RuntimeValue | undefined): string | undefined {
  if (value === undefined || value === null) return undefined;
  if (value instanceof RuntimeNamedValue) return value.typeName;
  if (value instanceof RuntimeInterfaceValue) return value.interfaceType;
  if (value instanceof RuntimeTypedNilValue) return value.typeName;
  if (typeof value === "bigint") return "int64";
  if (typeof value === "number") return "float64";
  if (isComplexValue(value)) return "complex128";
  if (isRuntimeString(value)) return "string";
  if (typeof value === "boolean") return "bool";
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return `map[${value.keyType}]${value.valueType}`;
  if (value instanceof RuntimeStruct) return value.typeName;
  if (value instanceof RuntimePointer) return `*${value.typeName}`;
  return undefined;
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
    if (!isRuntimeString(value)) throwTypeError(value, type, role);
    return ensureRuntimeGoString(value);
  }
  if (type === "bool") {
    if (typeof value !== "boolean") throwTypeError(value, type, role);
    return value;
  }
  if (isIntegerType(type)) {
    if (typeof value === "number" && Number.isInteger(value)) value = BigInt(value);
    if (typeof value !== "bigint" || !integerInRange(value, type)) throwTypeError(value, type, role);
    return value;
  }
  if (isFloatType(type)) {
    if (typeof value === "bigint") return roundFloatForType(Number(value), type);
    if (typeof value !== "number") throwTypeError(value, type, role);
    return roundFloatForType(value, type);
  }
  if (isComplexType(type)) {
    if (typeof value === "bigint" || typeof value === "number") return roundComplexForType(toComplex(value), type);
    if (!isComplexValue(value)) throwTypeError(value, type, role);
    return roundComplexForType(value, type);
  }
  const mapType = parseMapTypeText(type);
  if (mapType) {
    if (!(value instanceof RuntimeMap)) throwTypeError(value, type, role);
    if (!runtimeTypeAssignableMatch(value.keyType, mapType.keyType) ||
      !runtimeTypeAssignableMatch(value.valueType, mapType.valueType)) {
      throwTypeError(value, type, role);
    }
    return value;
  }
  const chanType = parseChanTypeText(type);
  if (chanType) {
    if (!(value instanceof RuntimeChannel) || !runtimeTypeAssignableMatch(value.elementType, chanType.elementType)) {
      throwTypeError(value, type, role);
    }
    return value;
  }
  if (type.startsWith("*")) {
    if (!(value instanceof RuntimePointer) || !runtimeTypeAssignableMatch(value.typeName, type.slice(1))) throwTypeError(value, type, role);
    return value;
  }
  const structType = context?.typeDef(type);
  if (structType) {
    if (!(value instanceof RuntimeStruct) || value.typeName !== structType.name) throwTypeError(value, type, role);
    return value;
  }
  const anonymousStruct = parseAnonymousStructTypeText(type);
  if (anonymousStruct) {
    if (!(value instanceof RuntimeStruct) || value.typeName !== anonymousStruct.name) throwTypeError(value, type, role);
    return value;
  }
  if (value instanceof RuntimeStruct) {
    if (value.typeName !== type) throwTypeError(value, type, role);
    return value;
  }
  const arrayType = parseArrayOrSliceTypeText(type, context);
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
  return context?.interfaceDef(type) ?? parseAnonymousInterfaceTypeText(type);
}

function prepareInterfaceAssignment(
  value: RuntimeValue,
  type: string,
  interfaceType: InterfaceTypeDef,
  role: string,
  context?: EvaluationContext,
  dynamicType?: string
): RuntimeInterfaceValue {
  if (value instanceof RuntimeInterfaceValue) {
    const dynamicValue = value.value;
    if (dynamicValue === null) return new RuntimeInterfaceValue(type, null);
    if (context && !valueImplementsInterface(dynamicValue, interfaceType, context)) throwTypeError(value, type, role);
    return new RuntimeInterfaceValue(type, dynamicValue);
  }
  if (value === null) return new RuntimeInterfaceValue(type, null);
  if (context && !valueImplementsInterface(value, interfaceType, context)) throwTypeError(value, type, role);
  return new RuntimeInterfaceValue(type, boxDynamicInterfaceValue(value, dynamicType));
}

function boxDynamicInterfaceValue(value: RuntimeValue, dynamicType: string | undefined): RuntimeValue {
  if (!dynamicType || value instanceof RuntimeNamedValue || value instanceof RuntimeTypedNilValue) return value;
  const type = canonicalRuntimeTypeName(normalizeTypeText(dynamicType));
  if (!type || type === "nil" || type === "any" || type === "interface{}") return value;
  if (value instanceof RuntimeStruct || value instanceof RuntimePointer || value instanceof RuntimeMap || value instanceof RuntimeChannel) return value;
  if (isRuntimeCallable(value) || isGoJuniorFunction(value) || value instanceof SheetBinding) return value;
  return new RuntimeNamedValue(type, value);
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
    if (
      candidate.method.pointerReceiver &&
      !(candidate.receiver instanceof RuntimePointer) &&
      !isTypedNilPointer(candidate.receiver)
    ) return false;
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
  if (!type || type === "<missing>") return false;
  if (type === "any" || type === "interface{}") return true;
  const interfaceType = context?.interfaceDef(type);
  if (interfaceType && context) return valueImplementsInterface(value, interfaceType, context);
  if (value instanceof RuntimeNamedValue) {
    return canonicalRuntimeTypeName(normalizeTypeText(value.typeName)) === canonicalRuntimeTypeName(type);
  }
  if (value instanceof RuntimeTypedNilValue) {
    if (type.startsWith("*")) return runtimeTypeAssignableMatch(value.typeName, type);
  }
  if (type === "string") return isRuntimeString(value);
  if (type === "bool") return typeof value === "boolean";
  if (isIntegerType(type)) return type === "int64" && typeof value === "bigint" && integerInRange(value, type);
  if (isFloatType(type)) return typeof value === "number";
  if (isComplexType(type)) return isComplexValue(value);
  if (type.startsWith("*")) return value instanceof RuntimePointer && runtimeTypeAssignableMatch(value.typeName, type.slice(1));
  if (value instanceof RuntimeStruct) return value.typeName === type;
  if (value instanceof RuntimeMap) {
    const mapType = parseMapTypeText(type);
    return Boolean(mapType &&
      runtimeTypeAssignableMatch(value.keyType, mapType.keyType) &&
      runtimeTypeAssignableMatch(value.valueType, mapType.valueType));
  }
  if (value instanceof RuntimeChannel) {
    const chanType = parseChanTypeText(type);
    return Boolean(chanType && runtimeTypeAssignableMatch(value.elementType, chanType.elementType));
  }
  if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type)) return Array.isArray(value);
  if (type.startsWith("func(")) return isRuntimeCallable(value) || isGoJuniorFunction(value);
  return false;
}

function canonicalRuntimeTypeName(type: string): string {
  switch (type) {
    case "byte":
      return "uint8";
    case "rune":
      return "int32";
    default:
      return type;
  }
}

function runtimeTypeAssignableMatch(actual: string, expected: string): boolean {
  return assignmentRuntimeTypeName(actual) === assignmentRuntimeTypeName(expected);
}

function assignmentRuntimeTypeName(type: string): string {
  const normalized = canonicalRuntimeTypeName(normalizeTypeText(type));
  if (normalized === "int") return "int64";
  if (normalized.startsWith("*")) return `*${assignmentRuntimeTypeName(normalized.slice(1))}`;
  const mapType = parseMapTypeText(normalized);
  if (mapType) return `map[${assignmentRuntimeTypeName(mapType.keyType)}]${assignmentRuntimeTypeName(mapType.valueType)}`;
  const chanType = parseChanTypeText(normalized);
  if (chanType) return `${chanType.direction === "receive" ? "<-" : chanType.direction === "send" ? "chan<-" : "chan"}${assignmentRuntimeTypeName(chanType.elementType)}`;
  return normalized;
}

function valueLength(value: RuntimeValue): number {
  value = unwrapNamed(value);
  if (value instanceof RuntimeTypedNilValue) {
    const type = normalizeTypeText(value.typeName);
    if (type.startsWith("*")) {
      const arrayType = parseArrayOrSliceTypeText(type.slice(1));
      if (arrayType?.length !== undefined && !arrayType.inferLength) return arrayType.length;
    }
    if (type.startsWith("[]") || type.startsWith("map[") || parseChanTypeText(type)) return 0;
  }
  if (value instanceof RuntimePointer) {
    const arrayType = parseArrayOrSliceTypeText(value.typeName);
    if (arrayType?.length !== undefined && !arrayType.inferLength) return arrayType.length;
  }
  if (isRuntimeString(value)) return ensureRuntimeGoString(value).byteLength();
  if (Array.isArray(value)) return value.length;
  if (value instanceof RuntimeMap) return value.size();
  if (value instanceof RuntimeChannel) return value.len();
  throw new GoJuniorRuntimeError(`${formatValue(value)} has no len`);
}

function valueCapacity(value: RuntimeValue): number {
  value = unwrapNamed(value);
  if (value instanceof RuntimeTypedNilValue) {
    const type = normalizeTypeText(value.typeName);
    if (type.startsWith("*")) {
      const arrayType = parseArrayOrSliceTypeText(type.slice(1));
      if (arrayType?.length !== undefined && !arrayType.inferLength) return arrayType.length;
    }
    if (type.startsWith("[]") || parseChanTypeText(type)) return 0;
  }
  if (value instanceof RuntimePointer) {
    const arrayType = parseArrayOrSliceTypeText(value.typeName);
    if (arrayType?.length !== undefined && !arrayType.inferLength) return arrayType.length;
  }
  if (Array.isArray(value)) return sliceCapacity(value);
  if (value instanceof RuntimeChannel) return value.cap();
  throw new GoJuniorRuntimeError(`${formatValue(value)} has no cap`);
}

function sliceCapacity(value: RuntimeValue[]): number {
  return arrayCapacities.get(value) ?? value.length;
}

function markArrayType(value: RuntimeValue[], typeText: string): void {
  arrayTypeTexts.set(value, normalizeTypeText(typeText));
}

function markArrayView(value: RuntimeValue[], source: RuntimeValue[], offset: number): void {
  arrayViews.set(value, { source, offset });
}

function getArrayElement(value: RuntimeValue[], index: number): RuntimeValue {
  const view = arrayViews.get(value);
  if (view) return getArrayElement(view.source, view.offset + index);
  return value[index] ?? null;
}

function setArrayElement(value: RuntimeValue[], index: number, item: RuntimeValue): void {
  value[index] = item;
  const view = arrayViews.get(value);
  if (view) setArrayElement(view.source, view.offset + index, item);
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

interface ParsedArrayOrSliceType {
  elementType: string;
  length?: number;
  inferLength: boolean;
  typeText: string;
}

function parseArrayOrSliceTypeText(typeText: string, context?: EvaluationContext): ParsedArrayOrSliceType | undefined {
  const type = normalizeTypeText(typeText);
  if (!type.startsWith("[")) return undefined;

  const close = type.indexOf("]");
  if (close < 0) return undefined;
  const lengthText = type.slice(1, close);
  const elementType = type.slice(close + 1);
  if (!elementType) return undefined;

  if (lengthText === "") {
    return { elementType, inferLength: false, typeText: `[]${elementType}` };
  }
  if (lengthText === "...") {
    return { elementType, inferLength: true, typeText: `[...]${elementType}` };
  }
  const length = resolveArrayLengthText(lengthText, context);
  if (length === undefined) return undefined;
  return { elementType, length, inferLength: false, typeText: `[${length}]${elementType}` };
}

function resolveArrayLengthText(lengthText: string, context?: EvaluationContext): number | undefined {
  const constant = evaluateArrayLengthConstant(lengthText, context);
  if (constant === undefined) return undefined;
  if (constant < 0n || constant > BigInt(Number.MAX_SAFE_INTEGER)) return undefined;
  return Number(constant);
}

function evaluateArrayLengthConstant(text: string, context?: EvaluationContext): bigint | undefined {
  const parser = new ArrayLengthConstantParser(text, context);
  try {
    return parser.parse();
  } catch {
    return undefined;
  }
}

class ArrayLengthConstantParser {
  private offset = 0;

  public constructor(
    private readonly text: string,
    private readonly context?: EvaluationContext
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
      left = applyArrayLengthConstantOperator(operator.text, left, right);
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
      if (!this.consume(")")) return undefined;
      return value;
    }

    const integer = this.consumeIntegerLiteral();
    if (integer !== undefined) return integer;

    const identifier = this.consumeIdentifier();
    if (identifier) return this.lookupIdentifier(identifier);

    return undefined;
  }

  private lookupIdentifier(name: string): bigint | undefined {
    if (!this.context) return undefined;
    const value = unwrapNamed(this.context.lookup(name));
    if (typeof value === "bigint") return value;
    if (typeof value === "number" && Number.isSafeInteger(value)) return BigInt(value);
    return undefined;
  }

  private consumeIntegerLiteral(): bigint | undefined {
    const start = this.offset;
    if (this.text.startsWith("0x", start) || this.text.startsWith("0X", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[0-9a-fA-F_]/.test(this.text[this.offset]!)) this.offset += 1;
      return this.integerFromDigits(start, 16, 2);
    }
    if (this.text.startsWith("0b", start) || this.text.startsWith("0B", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[01_]/.test(this.text[this.offset]!)) this.offset += 1;
      return this.integerFromDigits(start, 2, 2);
    }
    if (this.text.startsWith("0o", start) || this.text.startsWith("0O", start)) {
      this.offset += 2;
      while (this.offset < this.text.length && /[0-7_]/.test(this.text[this.offset]!)) this.offset += 1;
      return this.integerFromDigits(start, 8, 2);
    }
    while (this.offset < this.text.length && /[0-9_]/.test(this.text[this.offset]!)) this.offset += 1;
    if (this.offset === start) return undefined;
    const raw = this.text.slice(start, this.offset).replace(/_/g, "");
    if (raw.length > 1 && raw.startsWith("0")) {
      if (!/^[0-7]+$/.test(raw)) return undefined;
      return BigInt(`0o${raw.slice(1) || "0"}`);
    }
    return BigInt(raw);
  }

  private integerFromDigits(start: number, radix: number, prefixLength: number): bigint | undefined {
    const raw = this.text.slice(start + prefixLength, this.offset).replace(/_/g, "");
    if (!raw) return undefined;
    if (radix === 16) return BigInt(`0x${raw}`);
    if (radix === 8) return BigInt(`0o${raw}`);
    return BigInt(`0b${raw}`);
  }

  private consumeIdentifier(): string | undefined {
    this.skipSpace();
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

function applyArrayLengthConstantOperator(operator: string, left: bigint, right: bigint): bigint | undefined {
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

function parseAnonymousStructTypeText(typeText: string): StructTypeDef | undefined {
  const name = normalizeTypeText(typeText);
  const cached = anonymousStructTypeCache.get(name);
  if (cached) return cached;

  const match = /^struct\s*\{([\s\S]*)\}$/.exec(typeText.trim());
  if (!match) return undefined;
  const body = match[1]?.trim() ?? "";
  const fields = body === ""
    ? []
    : splitTopLevel(body, ";").flatMap(parseAnonymousStructField);
  const typeDef = { name, fields };
  anonymousStructTypeCache.set(name, typeDef);
  return typeDef;
}

function parseAnonymousStructField(text: string): StructFieldDecl[] {
  const field = text.trim();
  if (!field) return [];
  const withoutTag = stripStructFieldTag(field);
  const split = lastTopLevelWhitespace(withoutTag);
  if (split < 0) {
    return [{
      name: embeddedFieldNameFromTypeText(withoutTag),
      type: { text: withoutTag },
      embedded: true
    }];
  }
  const namesText = withoutTag.slice(0, split).trim();
  const typeText = withoutTag.slice(split).trim();
  return splitTopLevel(namesText, ",")
    .map((name) => name.trim())
    .filter(Boolean)
    .map((name) => ({ name, type: { text: typeText } }));
}

function parseAnonymousInterfaceTypeText(typeText: string): InterfaceTypeDef | undefined {
  const name = normalizeTypeText(typeText);
  const cached = anonymousInterfaceTypeCache.get(name);
  if (cached) return cached;

  const match = /^interface\s*\{([\s\S]*)\}$/.exec(typeText.trim());
  if (!match) return undefined;
  const body = match[1]?.trim() ?? "";
  const methods: InterfaceMethodDecl[] = [];
  const embeds: TypeNode[] = [];
  for (const item of body === "" ? [] : splitTopLevel(body, ";")) {
    const text = item.trim();
    if (!text) continue;
    const open = text.indexOf("(");
    if (open > 0 && /^[A-Za-z_]\w*$/.test(text.slice(0, open).trim())) {
      const close = matchingParen(text, open);
      if (close > open) {
        methods.push({
          name: text.slice(0, open).trim(),
          signature: {
            parameters: parseSignatureParameterListText(text.slice(open + 1, close)),
            results: parseSignatureResultText(text.slice(close + 1).trim())
          }
        });
        continue;
      }
    }
    embeds.push({ text });
  }
  const typeDef = { name, methods, embeds };
  anonymousInterfaceTypeCache.set(name, typeDef);
  return typeDef;
}

function parseSignatureParameterListText(text: string): FunctionDecl["signature"]["parameters"] {
  const trimmed = text.trim();
  if (!trimmed) return [];
  return splitTopLevel(trimmed, ",").flatMap(parseSignatureParameterText);
}

function parseSignatureParameterText(text: string): FunctionDecl["signature"]["parameters"] {
  const parameter = text.trim();
  if (!parameter) return [];
  const split = lastTopLevelWhitespace(parameter);
  const namesText = split < 0 ? "" : parameter.slice(0, split).trim();
  let typeText = split < 0 ? parameter : parameter.slice(split).trim();
  const variadic = typeText.startsWith("...");
  if (variadic) typeText = typeText.slice(3);
  const type = { text: typeText };
  if (!namesText) return [{ type, variadic }];
  return splitTopLevel(namesText, ",")
    .map((name) => name.trim())
    .filter(Boolean)
    .map((name) => ({ name, type, variadic }));
}

function parseSignatureResultText(text: string): FunctionDecl["signature"]["results"] {
  const trimmed = text.trim();
  if (!trimmed) return [];
  if (trimmed.startsWith("(")) {
    const close = matchingParen(trimmed, 0);
    if (close === trimmed.length - 1) return parseSignatureParameterListText(trimmed.slice(1, -1));
  }
  return parseSignatureParameterText(trimmed);
}

function matchingParen(text: string, openIndex: number): number {
  let depth = 0;
  for (let index = openIndex; index < text.length; index++) {
    const char = text[index] ?? "";
    if (char === "(") depth++;
    if (char === ")") {
      depth--;
      if (depth === 0) return index;
    }
  }
  return -1;
}

function stripStructFieldTag(field: string): string {
  let depth = 0;
  for (let index = 0; index < field.length; index++) {
    const char = field[index] ?? "";
    if (char === "[" || char === "(" || char === "{") depth++;
    if (char === "]" || char === ")" || char === "}") depth--;
    if (depth === 0 && char === "`") return field.slice(0, index).trim();
  }
  return field;
}

function lastTopLevelWhitespace(text: string): number {
  let depth = 0;
  let result = -1;
  for (let index = 0; index < text.length; index++) {
    const char = text[index] ?? "";
    if (char === "[" || char === "(" || char === "{") depth++;
    if (char === "]" || char === ")" || char === "}") depth--;
    if (depth === 0 && /\s/.test(char)) result = index;
  }
  return result;
}

function splitTopLevel(text: string, delimiter: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let start = 0;
  for (let index = 0; index < text.length; index++) {
    const char = text[index] ?? "";
    if (char === "[" || char === "(" || char === "{") depth++;
    if (char === "]" || char === ")" || char === "}") depth--;
    if (depth === 0 && char === delimiter) {
      parts.push(text.slice(start, index));
      start = index + 1;
    }
  }
  parts.push(text.slice(start));
  return parts;
}

function embeddedFieldNameFromTypeText(typeText: string): string {
  const normalized = normalizeTypeText(typeText);
  const base = normalized.startsWith("*") ? normalized.slice(1) : normalized;
  const selector = base.lastIndexOf(".");
  return selector >= 0 ? base.slice(selector + 1) : base;
}

function normalizeTypeText(typeText: string): string {
  return typeText.replace(/\s+/g, "");
}

function isRuntimeString(value: RuntimeValue): value is string | RuntimeGoString {
  return typeof value === "string" || value instanceof RuntimeGoString;
}

function ensureRuntimeGoString(value: string | RuntimeGoString): RuntimeGoString {
  return value instanceof RuntimeGoString ? value : RuntimeGoString.fromUtf8Text(value);
}

function goStringBytes(value: RuntimeValue): Uint8Array {
  if (value instanceof RuntimeGoString) return value.bytes;
  if (typeof value === "string") return new TextEncoder().encode(value);
  throwTypeError(value, "string", "conversion");
}

function goStringText(value: string | RuntimeGoString): string {
  return value instanceof RuntimeGoString ? value.text() : value;
}

function goStringByteAt(value: string | RuntimeGoString, index: number): bigint {
  return ensureRuntimeGoString(value).byteAt(index);
}

function goStringSlice(value: string | RuntimeGoString, start?: number, end?: number): RuntimeGoString {
  return ensureRuntimeGoString(value).slice(start, end);
}

function goStringKey(value: string | RuntimeGoString): string {
  return [...goStringBytes(value)].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

function compareGoStrings(left: string | RuntimeGoString, right: string | RuntimeGoString): number {
  const leftBytes = goStringBytes(left);
  const rightBytes = goStringBytes(right);
  const count = Math.min(leftBytes.length, rightBytes.length);
  for (let index = 0; index < count; index += 1) {
    const leftByte = leftBytes[index] ?? 0;
    const rightByte = rightBytes[index] ?? 0;
    if (leftByte !== rightByte) return leftByte < rightByte ? -1 : 1;
  }
  return leftBytes.length === rightBytes.length ? 0 : leftBytes.length < rightBytes.length ? -1 : 1;
}

function goStringFromRunes(values: bigint[]): RuntimeGoString {
  return RuntimeGoString.fromUtf8Text(values.map((value) => goStringFromRune(value).text()).join(""));
}

function goStringRuneEntries(value: RuntimeValue): Array<{ index: number; rune: bigint }> {
  const bytes = goStringBytes(value);
  const entries: Array<{ index: number; rune: bigint }> = [];
  for (let index = 0; index < bytes.length;) {
    const decoded = decodeUtf8Rune(bytes, index);
    entries.push({ index, rune: BigInt(decoded.rune) });
    index += decoded.width;
  }
  return entries;
}

function decodeUtf8Rune(bytes: Uint8Array, index: number): { rune: number; width: number } {
  const first = bytes[index] ?? 0;
  if (first < 0x80) return { rune: first, width: 1 };
  const replacement = { rune: 0xfffd, width: 1 };
  if (first < 0xc2) return replacement;
  if (first < 0xe0) {
    const b1 = bytes[index + 1];
    if (b1 === undefined || (b1 & 0xc0) !== 0x80) return replacement;
    return { rune: ((first & 0x1f) << 6) | (b1 & 0x3f), width: 2 };
  }
  if (first < 0xf0) {
    const b1 = bytes[index + 1];
    const b2 = bytes[index + 2];
    if (b1 === undefined || b2 === undefined || (b1 & 0xc0) !== 0x80 || (b2 & 0xc0) !== 0x80) return replacement;
    if (first === 0xe0 && b1 < 0xa0) return replacement;
    if (first === 0xed && b1 >= 0xa0) return replacement;
    return { rune: ((first & 0x0f) << 12) | ((b1 & 0x3f) << 6) | (b2 & 0x3f), width: 3 };
  }
  if (first < 0xf5) {
    const b1 = bytes[index + 1];
    const b2 = bytes[index + 2];
    const b3 = bytes[index + 3];
    if (b1 === undefined || b2 === undefined || b3 === undefined || (b1 & 0xc0) !== 0x80 || (b2 & 0xc0) !== 0x80 || (b3 & 0xc0) !== 0x80) return replacement;
    if (first === 0xf0 && b1 < 0x90) return replacement;
    if (first === 0xf4 && b1 >= 0x90) return replacement;
    return { rune: ((first & 0x07) << 18) | ((b1 & 0x3f) << 12) | ((b2 & 0x3f) << 6) | (b3 & 0x3f), width: 4 };
  }
  return replacement;
}

function goStringFromLiteralRaw(raw: string): RuntimeGoString {
  if (raw.length >= 2 && raw.startsWith("`") && raw.endsWith("`")) {
    return RuntimeGoString.fromUtf8Text(raw.slice(1, -1).replace(/\r/g, ""));
  }
  if (raw.length >= 2 && raw.startsWith("\"") && raw.endsWith("\"")) {
    return new RuntimeGoString(decodeGoStringLiteralBytes(raw.slice(1, -1)));
  }
  return RuntimeGoString.fromUtf8Text(raw);
}

function decodeGoStringLiteralBytes(value: string): Uint8Array {
  const bytes: number[] = [];
  for (let index = 0; index < value.length; index += 1) {
    const char = value[index] ?? "";
    if (char !== "\\") {
      bytes.push(...new TextEncoder().encode(char));
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
        bytes.push(parseInt(value.slice(index + 1, index + 3), 16));
        index += 2;
        break;
      case "u":
        bytes.push(...new TextEncoder().encode(String.fromCodePoint(parseInt(value.slice(index + 1, index + 5), 16))));
        index += 4;
        break;
      case "U":
        bytes.push(...new TextEncoder().encode(String.fromCodePoint(parseInt(value.slice(index + 1, index + 9), 16))));
        index += 8;
        break;
      default:
        if (/[0-7]/.test(next)) {
          bytes.push(parseInt(next + value.slice(index + 1, index + 3), 8));
          index += 2;
        } else {
          bytes.push(...new TextEncoder().encode(next));
        }
    }
  }
  return new Uint8Array(bytes.map((byte) => byte & 0xff));
}

function externalizeRuntimeValue(value: RuntimeValue): RuntimeValue {
  if (value instanceof RuntimeGoString) return value.text();
  if (value instanceof RuntimeNamedValue) return new RuntimeNamedValue(value.typeName, externalizeRuntimeValue(value.value));
  if (value instanceof RuntimeInterfaceValue) return new RuntimeInterfaceValue(value.interfaceType, value.value === null ? null : externalizeRuntimeValue(value.value));
  if (Array.isArray(value)) return value.map(externalizeRuntimeValue);
  if (value instanceof RuntimeStruct) {
    const struct = new RuntimeStruct(value.typeName);
    for (const [name, item] of value.orderedFields()) struct.set(name, externalizeRuntimeValue(item));
    return struct;
  }
  return value;
}

function genericBaseTypeName(typeText: string): string {
  const type = normalizeTypeText(typeText);
  if (type.startsWith("[") || type.startsWith("map[") || !type.endsWith("]")) return type;
  const bracket = type.indexOf("[");
  return bracket > 0 ? type.slice(0, bracket) : type;
}

function isTypeArgumentExpression(expression: Expression, context: EvaluationContext): boolean {
  const typeText = typeArgumentText(expression);
  if (typeText === undefined) return false;
  const parts = splitTopLevelTypeArguments(typeText);
  return parts.length > 1
    ? parts.every((part) => context.isKnownType(part))
    : context.isKnownType(typeText);
}

function splitTopLevelTypeArguments(typeText: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let start = 0;
  for (let index = 0; index < typeText.length; index += 1) {
    const char = typeText[index];
    if (char === "[" || char === "(" || char === "{") depth += 1;
    else if (char === "]" || char === ")" || char === "}") depth -= 1;
    else if (char === "," && depth === 0) {
      parts.push(typeText.slice(start, index).trim());
      start = index + 1;
    }
  }
  parts.push(typeText.slice(start).trim());
  return parts.filter((part) => part.length > 0);
}

function typeArgumentText(expression: Expression): string | undefined {
  switch (expression.kind) {
    case "TypeExpression":
      return expression.type.text;
    case "Identifier":
      return expression.name;
    case "SelectorExpression": {
      const object = typeArgumentText(expression.object);
      return object ? `${object}.${expression.field}` : undefined;
    }
    case "IndexExpression": {
      const object = typeArgumentText(expression.object);
      const index = typeArgumentText(expression.index);
      return object && index ? `${object}[${index}]` : undefined;
    }
    default:
      return undefined;
  }
}

function runtimeMapKeyId(key: RuntimeValue, options: { freshNonReflexive?: boolean } = {}): string {
  const actual = unwrapNamed(key);
  if (actual === null) return "nil";
  if (actual instanceof RuntimeInterfaceValue) {
    return actual.value === null
      ? `iface:${actual.interfaceType}:nil`
      : `iface:${actual.interfaceType}:${runtimeMapKeyId(actual.value, options)}`;
  }
  if (actual instanceof RuntimeTypedNilValue) return `typednil:${actual.typeName}`;
  if (typeof actual === "boolean") return `b:${actual}`;
  if (isRuntimeString(actual)) return `s:${goStringKey(actual)}`;
  if (typeof actual === "bigint") return `i:${actual}`;
  if (typeof actual === "number") return `f:${mapNumberKeyId(actual, options)}`;
  if (isComplexValue(actual)) return `c:${mapNumberKeyId(actual.real, options)}:${mapNumberKeyId(actual.imag, options)}`;
  if (Array.isArray(actual)) return `a:[${actual.map((item) => runtimeMapKeyId(item, options)).join(",")}]`;
  if (actual instanceof RuntimeStruct) {
    return `st:${actual.typeName}{${actual.orderedFields().map(([name, item]) => `${name}:${runtimeMapKeyId(item, options)}`).join(",")}}`;
  }
  return `o:${objectIdentityId(actual as object)}`;
}

function mapNumberKeyId(value: number, options: { freshNonReflexive?: boolean }): string {
  if (Number.isNaN(value)) {
    if (options.freshNonReflexive) {
      const id = nextNonReflexiveMapKeyId;
      nextNonReflexiveMapKeyId += 1;
      return `NaN#${id}`;
    }
    return "NaN";
  }
  return String(value);
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
  const arrayType = parseArrayOrSliceTypeText(type, context);
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
    return;
  }
  const anonymousStruct = parseAnonymousStructTypeText(type);
  if (anonymousStruct) {
    for (const field of anonymousStruct.fields) assertComparableType(field.type.text, context);
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

function convertIntegerToType(value: bigint, type: string): bigint {
  const normalized = canonicalRuntimeTypeName(normalizeTypeText(type));
  switch (normalized) {
    case "uint8":
      return BigInt.asUintN(8, value);
    case "int8":
      return BigInt.asIntN(8, value);
    case "uint16":
      return BigInt.asUintN(16, value);
    case "int16":
      return BigInt.asIntN(16, value);
    case "uint32":
      return BigInt.asUintN(32, value);
    case "int32":
      return BigInt.asIntN(32, value);
    case "uint":
    case "uint64":
    case "uintptr":
      return BigInt.asUintN(64, value);
    case "int":
    case "int64":
      return BigInt.asIntN(64, value);
    default:
      return value;
  }
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
  if (isRuntimeString(actualLeft) || isRuntimeString(actualRight)) {
    if (isRuntimeString(actualLeft) && isRuntimeString(actualRight)) return ensureRuntimeGoString(actualLeft).concat(ensureRuntimeGoString(actualRight));
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

function float32Bits(value: number): bigint {
  const view = new DataView(new ArrayBuffer(4));
  view.setFloat32(0, value, true);
  return BigInt(view.getUint32(0, true));
}

function float32FromBits(bits: bigint): number {
  const view = new DataView(new ArrayBuffer(4));
  view.setUint32(0, Number(BigInt.asUintN(32, bits)), true);
  return view.getFloat32(0, true);
}

function float64Bits(value: number): bigint {
  const view = new DataView(new ArrayBuffer(8));
  view.setFloat64(0, value, true);
  return view.getBigUint64(0, true);
}

function float64FromBits(bits: bigint): number {
  const view = new DataView(new ArrayBuffer(8));
  view.setBigUint64(0, BigInt.asUintN(64, bits), true);
  return view.getFloat64(0, true);
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
    return compareNumericValues(left, right);
  }
  if (isRuntimeString(left) && isRuntimeString(right)) {
    return compareGoStrings(left, right);
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
    return compareNumericValues(left, right) === 0;
  }
  if (isComplexValue(left) || isComplexValue(right)) {
    const a = toComplex(left);
    const b = toComplex(right);
    return a.real === b.real && a.imag === b.imag;
  }
  if (isRuntimeString(left) || isRuntimeString(right)) {
    return isRuntimeString(left) && isRuntimeString(right) && compareGoStrings(left, right) === 0;
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

function compareNumericValues(left: bigint | number, right: bigint | number): number {
  if (typeof left === "bigint" && typeof right === "bigint") {
    return left === right ? 0 : left < right ? -1 : 1;
  }
  const leftNumber = typeof left === "bigint" ? Number(left) : left;
  const rightNumber = typeof right === "bigint" ? Number(right) : right;
  if (Number.isNaN(leftNumber) || Number.isNaN(rightNumber)) return Number.NaN;
  return leftNumber === rightNumber ? 0 : leftNumber < rightNumber ? -1 : 1;
}

function cmpCompare(left: RuntimeValue, right: RuntimeValue): number {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  const leftNaN = isRuntimeNaN(left);
  const rightNaN = isRuntimeNaN(right);
  if (leftNaN) return rightNaN ? 0 : -1;
  if (rightNaN) return 1;
  const compared = compareValues(left, right);
  if (Number.isNaN(compared)) return 0;
  return compared < 0 ? -1 : compared > 0 ? 1 : 0;
}

function cmpLess(left: RuntimeValue, right: RuntimeValue): boolean {
  left = unwrapNamed(left);
  right = unwrapNamed(right);
  return (isRuntimeNaN(left) && !isRuntimeNaN(right)) || compareValues(left, right) < 0;
}

function isRuntimeNaN(value: RuntimeValue): boolean {
  value = unwrapNamed(value);
  return typeof value === "number" && Number.isNaN(value);
}

function isRuntimeZeroValue(value: RuntimeValue): boolean {
  value = unwrapNamed(value);
  if (value === null) return true;
  if (value instanceof RuntimeInterfaceValue) return value.value === null;
  if (value instanceof RuntimeTypedNilValue) return true;
  if (typeof value === "boolean") return !value;
  if (typeof value === "bigint") return value === 0n;
  if (typeof value === "number") return Object.is(value, 0) || Object.is(value, -0);
  if (isRuntimeString(value)) return goStringBytes(value).length === 0;
  if (isComplexValue(value)) return value.real === 0 && value.imag === 0;
  return false;
}

function zeroValueLike(value: RuntimeValue): RuntimeValue {
  value = unwrapNamed(value);
  if (typeof value === "boolean") return false;
  if (typeof value === "bigint") return 0n;
  if (typeof value === "number") return 0;
  if (isRuntimeString(value)) return "";
  if (isComplexValue(value)) return complexValue(0, 0);
  return null;
}

function runtimeAlignof(value: RuntimeValue): number {
  return Math.max(1, Math.min(8, runtimeSizeof(value)));
}

function runtimeSizeof(value: RuntimeValue): number {
  value = unwrapNamed(value);
  if (value === null) return 0;
  if (typeof value === "boolean") return 1;
  if (typeof value === "bigint" || typeof value === "number") return 8;
  if (isComplexValue(value)) return 16;
  if (isRuntimeString(value)) return 16;
  if (value instanceof RuntimePointer || value instanceof RuntimeChannel || value instanceof RuntimeMap) return 8;
  if (isRuntimeCallable(value)) return 8;
  if (value instanceof RuntimeInterfaceValue) return 16;
  if (value instanceof RuntimeTypedNilValue) return value.typeName.startsWith("*") ? 8 : 0;
  if (Array.isArray(value)) return 24;
  if (value instanceof RuntimeStruct) {
    let size = 0;
    for (const [, field] of value.orderedFields()) size += runtimeSizeof(field);
    return size;
  }
  if (typeof value === "object") return 8;
  return 0;
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
    if (source < 0) return [];
    const count = toNonNegativeLength(source, "range count");
    return Array.from({ length: count }, (_, index) => [BigInt(index), BigInt(index)]);
  }
  if (isRuntimeCallable(source) || isGoJuniorFunction(source)) {
    return iteratorFunctionEntries(source, context);
  }
  if (Array.isArray(source)) {
    return source.map((value, index) => [BigInt(index), value]);
  }
  if (isRuntimeString(source)) {
    return goStringRuneEntries(source).map((entry) => [BigInt(entry.index), entry.rune]);
  }
  if (source instanceof RuntimeMap) {
    return source.orderedEntries();
  }
  if (isRuntimeObject(source)) {
    return Object.entries(source).map(([key, value]) => [key, value]);
  }
  throw new GoJuniorRuntimeError(`${formatValue(source)} is not rangeable`);
}

function integerRangeKeyTypeText(sourceExpression: Expression, source: RuntimeValue, context: EvaluationContext): string | undefined {
  const actual = unwrapNamed(source);
  if (typeof actual !== "bigint" && typeof actual !== "number") return undefined;
  return expressionDynamicTypeText(sourceExpression, context, source) ?? inferredTypeText(source);
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

async function sprintfAsync(format: string, args: RuntimeValue[], context: EvaluationContext): Promise<string> {
  let argIndex = 0;
  let text = "";
  const pattern = /%(#)?[%vdsft]/g;
  let lastIndex = 0;
  for (const match of format.matchAll(pattern)) {
    text += format.slice(lastIndex, match.index);
    lastIndex = (match.index ?? 0) + match[0].length;
    if (match[0] === "%%") {
      text += "%";
      continue;
    }
    const value = args[argIndex++] ?? null;
    text += await formatFmtValue(match[0], match[1] === "#", value, context);
  }
  return text + format.slice(lastIndex);
}

async function formatFmtValue(verb: string, goSyntax: boolean, value: RuntimeValue, context: EvaluationContext): Promise<string> {
  if (!goSyntax) {
    const stringer = await fmtStringerValue(value, context);
    if (stringer !== undefined && (verb === "%v" || verb === "%s")) return stringer;
  }
  switch (verb) {
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
}

async function fmtStringerValue(value: RuntimeValue, context: EvaluationContext): Promise<string | undefined> {
  const method = methodForValue(value, "String", context);
  if (!method || method.method.declaration.signature.parameters.length !== 0) return undefined;
  const results = method.method.declaration.signature.results;
  if (results.length !== 1 || normalizeTypeText(results[0]?.type.text ?? "") !== "string") return undefined;
  const result = await callRuntime(boundMethodValue(method.method, method.receiver), [], context);
  return toStringValue(result);
}

function sprint(args: RuntimeValue[]): string {
  let text = "";
  let previousWasString = false;
  for (const arg of args) {
    const currentIsString = runtimeValueIsString(arg);
    if (text && !previousWasString && !currentIsString) text += " ";
    text += formatValue(arg);
    previousWasString = currentIsString;
  }
  return text;
}

async function sprintAsync(args: RuntimeValue[], context: EvaluationContext): Promise<string> {
  let text = "";
  let previousWasString = false;
  for (const arg of args) {
    const currentIsString = runtimeValueIsString(arg);
    if (text && !previousWasString && !currentIsString) text += " ";
    text += await formatFmtValue("%v", false, arg, context);
    previousWasString = currentIsString;
  }
  return text;
}

function sprintln(args: RuntimeValue[]): string {
  return `${args.map(formatValue).join(" ")}\n`;
}

async function sprintlnAsync(args: RuntimeValue[], context: EvaluationContext): Promise<string> {
  return `${(await Promise.all(args.map((arg) => formatFmtValue("%v", false, arg, context)))).join(" ")}\n`;
}

function runtimeValueIsString(value: RuntimeValue): boolean {
  if (value instanceof RuntimeNamedValue) return runtimeValueIsString(value.value);
  if (value instanceof RuntimeInterfaceValue) return value.value !== null && runtimeValueIsString(value.value);
  return isRuntimeString(value);
}

function toStringValue(value: RuntimeValue): string {
  if (isRuntimeString(value)) return goStringText(value);
  return formatValue(value);
}

async function toStringValueAsync(value: RuntimeValue, context: EvaluationContext): Promise<string> {
  if (isRuntimeString(value)) return goStringText(value);
  return await fmtStringerValue(value, context) ?? formatValue(value);
}

export function formatValue(value: RuntimeValue): string {
  if (value instanceof RuntimeNamedValue) return formatValue(value.value);
  if (value instanceof RuntimeInterfaceValue) return value.value === null ? "<nil>" : formatValue(value.value);
  if (value instanceof RuntimeTypedNilValue) return "<nil>";
  if (value === null) return "<nil>";
  if (typeof value === "bigint") return value.toString();
  if (isComplexValue(value)) return `(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  if (isRuntimeString(value)) return goStringText(value);
  if (Array.isArray(value)) return `[${value.map(formatValue).join(" ")}]`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatRuntimeMap(value);
  if (value instanceof RuntimeStruct) return formatRuntimeStruct(value);
  if (value instanceof RuntimePointer) return `&${formatValue(value.get())}`;
  if (isGoJuniorFunction(value)) return formatFunctionValue(value);
  if (isRuntimeCallable(value)) return `<func ${value.name}>`;
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
  if (isRuntimeString(value)) return formatReplString(goStringText(value));
  if (Array.isArray(value)) return `[${value.map(formatReplValue).join(" ")}]`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatReplMap(value);
  if (value instanceof RuntimeStruct) return formatReplStruct(value);
  if (value instanceof RuntimePointer) return `&${formatReplValue(value.get())}`;
  if (isGoJuniorFunction(value)) return formatFunctionValue(value);
  if (isRuntimeCallable(value)) return `<func ${value.name}>`;
  if (value instanceof SheetBinding) return `<sheet ${value.name}>`;
  return `{${Object.entries(value).map(([key, item]) => `${key}:${formatReplValue(item)}`).join(" ")}}`;
}

function formatGoSyntaxValue(value: RuntimeValue): string {
  if (value instanceof RuntimeNamedValue) {
    return value.value instanceof RuntimeTypedNilValue
      ? `${value.typeName}(nil)`
      : `${value.typeName}(${formatGoSyntaxValue(value.value)})`;
  }
  if (value instanceof RuntimeInterfaceValue) {
    return value.value === null ? `${value.interfaceType}(nil)` : formatGoSyntaxValue(value.value);
  }
  if (value instanceof RuntimeTypedNilValue) return `${value.typeName}(nil)`;
  if (value === null) return "nil";
  if (typeof value === "bigint") return `${integerTypeName(value)}(${value.toString()})`;
  if (typeof value === "number") return `float64(${formatFloat(value)})`;
  if (isComplexValue(value)) return `complex128(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
  if (typeof value === "boolean") return `bool(${value})`;
  if (isRuntimeString(value)) return `string(${JSON.stringify(goStringText(value))})`;
  if (Array.isArray(value)) return `[]interface{}{${value.map(formatGoSyntaxValue).join(", ")}}`;
  if (value instanceof RuntimeChannel) return `chan ${value.elementType}`;
  if (value instanceof RuntimeMap) return formatGoSyntaxMap(value);
  if (value instanceof RuntimeStruct) return formatGoSyntaxStruct(value);
  if (value instanceof RuntimePointer) return `&${formatGoSyntaxValue(value.get())}`;
  if (isGoJuniorFunction(value)) return formatFunctionValue(value);
  if (isRuntimeCallable(value)) return `<func ${value.name}>`;
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

function formatFunctionValue(value: GoJuniorFunction): string {
  return value.source?.trimEnd() || `<func ${value.name}>`;
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

function dependencyKey(dependency: SpreadsheetDependency): string {
  if (dependency.kind === "cell") return `cell:${dependency.sheet}:${dependency.cell}`;
  return `range:${dependency.sheet}:${dependency.start}:${dependency.end}`;
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
