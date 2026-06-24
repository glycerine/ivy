import { REPL_FILENAME } from "./diagnostics.js";
import { checkGoJuniorSourceFiles, GOJR_SYNTHETIC_CHECK_PREFIX, isGoJuniorSyntheticCheckName, standardTypePackage } from "./typecheck.js";
import { Const as GoTypesConst, ensureUniverseInitialized, NewPackage, NewPkgName, NoPos, RelativeTo as GoTypesRelativeTo, TypeString as GoTypesTypeString } from "./go/types/index.js";
import { frontSourceFilesToAst, frontSourceToAst } from "./frontToAst.js";
import { isHostResolvedSourceImport } from "./intrinsicPackages.js";
import { DeterministicPrng } from "./prng.js";
import { AsyncGoChannel, AsyncGoDeadlockError, AsyncGoPanic, AsyncGoScheduler, asyncSelect } from "./asyncRuntime.js";
import { cellDependency, rangeDependency } from "./spreadsheet.js";
export class RuntimeGoString {
    bytes;
    constructor(bytes) {
        this.bytes = bytes;
    }
    static fromUtf8Text(text) {
        return new RuntimeGoString(new TextEncoder().encode(text));
    }
    text() {
        return new TextDecoder("utf-8").decode(this.bytes);
    }
    byteLength() {
        return this.bytes.length;
    }
    byteAt(index) {
        return BigInt(this.bytes[index] ?? 0);
    }
    slice(start, end) {
        return new RuntimeGoString(this.bytes.slice(start, end));
    }
    concat(other) {
        const bytes = new Uint8Array(this.bytes.length + other.bytes.length);
        bytes.set(this.bytes, 0);
        bytes.set(other.bytes, this.bytes.length);
        return new RuntimeGoString(bytes);
    }
}
const RECOVERED_PANIC = Symbol("recovered panic");
const anonymousStructTypeCache = new Map();
const anonymousInterfaceTypeCache = new Map();
const errorInterfaceType = {
    name: "error",
    methods: [{
            name: "Error",
            signature: {
                parameters: [],
                results: [{ type: { text: "string" }, variadic: false }]
            }
        }],
    embeds: []
};
const fmtErrorStringType = "fmt.errorString";
export class GoJuniorRuntimeError extends Error {
    span;
    constructor(message, span) {
        super(message);
        this.name = "GoJuniorRuntimeError";
        this.span = span;
    }
}
export class GoJuniorDeadlockError extends Error {
    constructor(message) {
        super(message);
        this.name = "GoJuniorDeadlockError";
    }
}
export class GoJuniorPanic extends Error {
    value;
    constructor(value) {
        super(`panic: ${formatValue(value)}`);
        this.name = "GoJuniorPanic";
        this.value = value;
    }
}
class GoJuniorTestStop extends Error {
    action;
    constructor(action) {
        super(action);
        this.action = action;
        this.name = "GoJuniorTestStop";
    }
}
export class EvaluationContext {
    options;
    shared;
    currentScope;
    deferFrames = [[]];
    callDepth = 0;
    activeRecoverPanic;
    recoverCallDepth;
    recoveredActivePanic = false;
    constructor(options = {}, shared, currentScope) {
        this.options = options;
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
            importPathsByLocalName: new Map(),
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
    get output() {
        return this.shared.output;
    }
    scheduler() {
        return this.shared.scheduler;
    }
    fork(currentScope = this.shared.rootScope) {
        return new EvaluationContext(this.options, this.shared, currentScope);
    }
    declare(name, value, mutable = true, type) {
        if (name === "_")
            return;
        const typeText = bindingTypeText(type);
        const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
        this.currentScope.declare(name, stored, mutable, typeText ? resolvedDeclaredTypeText(typeText, this) : undefined);
    }
    declareRoot(name, value, mutable = true, type) {
        if (name === "_")
            return;
        const typeText = bindingTypeText(type);
        const stored = typeText ? prepareAssignableToType(value, typeText, `variable ${name}`, this) : value;
        this.shared.rootScope.declare(name, stored, mutable, typeText ? resolvedDeclaredTypeText(typeText, this) : undefined);
    }
    declareOrAssignRoot(name, value, mutable = true, type) {
        if (this.shared.rootScope.hasLocal(name)) {
            this.shared.rootScope.assign(name, value, this.assignmentChecker());
            return;
        }
        this.declareRoot(name, value, mutable, type);
    }
    assign(name, value) {
        this.currentScope.assign(name, value, this.assignmentChecker());
    }
    lookup(name) {
        return this.currentScope.lookup(name);
    }
    lookupTypeText(name) {
        return this.currentScope.lookupTypeText(name);
    }
    isUntypedConstant(name) {
        return this.currentScope.isUntypedConstant(name);
    }
    hasLocal(name) {
        return this.currentScope.hasLocal(name);
    }
    hasBinding(name) {
        return this.currentScope.hasBinding(name);
    }
    pointerToBinding(name, typeName) {
        const scope = this.currentScope;
        return new RuntimePointer(typeName, () => scope.lookup(name), (next) => scope.assign(name, next, this.assignmentChecker()));
    }
    outputFrom(offset) {
        return this.output.slice(offset);
    }
    clearObservedDeps() {
        this.shared.observedDeps.clear();
    }
    observedDeps() {
        return [...this.shared.observedDeps.values()];
    }
    getenv(name) {
        return this.options.env?.[name] ?? "";
    }
    recordSheetCellRead(sheet, cell) {
        const dependency = cellDependency({ sheet, cell });
        this.shared.observedDeps.set(dependencyKey(dependency), dependency);
    }
    recordSheetRangeRead(sheet, start, end) {
        const dependency = rangeDependency(sheet, start, end);
        this.shared.observedDeps.set(dependencyKey(dependency), dependency);
    }
    childScope(body) {
        const previous = this.currentScope;
        this.currentScope = new Scope(previous);
        try {
            return body();
        }
        finally {
            this.currentScope = previous;
        }
    }
    captureScope() {
        return this.currentScope;
    }
    withScope(scope, body) {
        const previous = this.currentScope;
        this.currentScope = scope;
        try {
            return body();
        }
        finally {
            this.currentScope = previous;
        }
    }
    async childScopeAsync(body) {
        const previous = this.currentScope;
        this.currentScope = new Scope(previous);
        try {
            return await body();
        }
        finally {
            this.currentScope = previous;
        }
    }
    async withScopeAsync(scope, body) {
        const previous = this.currentScope;
        this.currentScope = scope;
        try {
            return await body();
        }
        finally {
            this.currentScope = previous;
        }
    }
    pushDefer(callback) {
        this.currentDeferFrame().push(callback);
    }
    async functionCallAsync(body) {
        this.callDepth += 1;
        try {
            return await body();
        }
        finally {
            this.callDepth -= 1;
        }
    }
    recover() {
        if (this.activeRecoverPanic &&
            this.recoverCallDepth === this.callDepth &&
            !this.recoveredActivePanic) {
            this.recoveredActivePanic = true;
            return this.activeRecoverPanic.value;
        }
        return null;
    }
    runDefers() {
        const frame = this.currentDeferFrame();
        while (frame.length > 0) {
            frame.pop()?.();
        }
    }
    async runDefersAsync(panic) {
        const frame = this.currentDeferFrame();
        let activePanic = panic;
        while (frame.length > 0) {
            const callback = frame.pop();
            if (!callback)
                continue;
            const previousRecoverPanic = this.activeRecoverPanic;
            const previousRecoverCallDepth = this.recoverCallDepth;
            const previousRecoveredActivePanic = this.recoveredActivePanic;
            if (activePanic) {
                this.activeRecoverPanic = activePanic;
                this.recoverCallDepth = this.callDepth + 1;
                this.recoveredActivePanic = false;
            }
            else {
                this.activeRecoverPanic = undefined;
                this.recoverCallDepth = undefined;
                this.recoveredActivePanic = false;
            }
            try {
                await callback();
            }
            catch (error) {
                const normalized = normalizeAsyncRuntimeError(error);
                if (normalized instanceof GoJuniorPanic) {
                    activePanic = normalized;
                }
                else {
                    throw normalized;
                }
            }
            finally {
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
    deferScope(body) {
        this.deferFrames.push([]);
        try {
            return body();
        }
        finally {
            try {
                this.runDefers();
            }
            finally {
                this.deferFrames.pop();
            }
        }
    }
    async deferScopeAsync(body) {
        this.deferFrames.push([]);
        try {
            const value = await body();
            const panic = await this.runDefersAsync();
            if (panic)
                throw panic;
            return value;
        }
        catch (error) {
            const normalized = normalizeAsyncRuntimeError(error);
            if (normalized instanceof GoJuniorPanic) {
                const panic = await this.runDefersAsync(normalized);
                if (panic)
                    throw panic;
                return RECOVERED_PANIC;
            }
            const panic = await this.runDefersAsync();
            if (panic)
                throw panic;
            throw normalized;
        }
        finally {
            this.deferFrames.pop();
        }
    }
    write(text) {
        this.shared.output.push(text);
        this.shared.stdout?.(text);
    }
    loopLimit() {
        return this.shared.maxLoopIterations;
    }
    randomIndex(length) {
        return this.shared.random.nextIndex(length);
    }
    packages() {
        return this.options.packages ?? {};
    }
    packageInfo(importPath) {
        return this.options.packageInfos?.[importPath];
    }
    importPath() {
        return this.options.importPath;
    }
    currentSheetName() {
        return this.options.currentSheetName ?? "sheet";
    }
    sheetData(name) {
        if (name === "sheet")
            return this.options.sheet ?? this.options.sheets?.[this.currentSheetName()] ?? {};
        return this.options.sheets?.[name] ?? {};
    }
    registerType(spec) {
        this.shared.aliases.set(spec.name, this.resolveImportedTypeText(spec.type.text));
        if (spec.structFields) {
            this.shared.types.set(spec.name, {
                name: spec.name,
                fields: spec.structFields.map((field) => ({
                    ...field,
                    type: { text: this.resolveImportedTypeText(field.type.text) }
                }))
            });
        }
        if (spec.interfaceMethods || spec.interfaceEmbeds) {
            this.shared.interfaces.set(spec.name, {
                name: spec.name,
                methods: this.flattenInterfaceMethods((spec.interfaceMethods ?? []).map((method) => ({
                    ...method,
                    signature: resolveRuntimeSignatureTypeImports(method.signature, this)
                })), (spec.interfaceEmbeds ?? []).map((embed) => ({ text: this.resolveImportedTypeText(embed.text) }))),
                embeds: (spec.interfaceEmbeds ?? []).map((embed) => ({ text: this.resolveImportedTypeText(embed.text) }))
            });
        }
    }
    typeDef(name) {
        return this.shared.types.get(name) ?? this.shared.types.get(genericBaseTypeName(name));
    }
    interfaceDef(name) {
        return this.shared.interfaces.get(name) ?? this.shared.interfaces.get(genericBaseTypeName(name));
    }
    aliasType(name) {
        return this.shared.aliases.get(name) ?? this.shared.aliases.get(genericBaseTypeName(name));
    }
    isKnownType(name) {
        const type = normalizeTypeText(name);
        const genericBase = genericBaseTypeName(type);
        return isPredeclaredType(type) ||
            isUnsafePointerType(type) ||
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
    registerMethod(declaration) {
        if (!declaration.receiver)
            return;
        const resolved = resolveRuntimeFunctionDeclarationTypeImports(declaration, this);
        const receiver = normalizeReceiverType(resolved.receiver?.type.text ?? declaration.receiver.type.text);
        this.shared.methods.set(methodKey(receiver.baseType, resolved.name), {
            declaration: resolved,
            receiverType: receiver.baseType,
            pointerReceiver: receiver.pointer,
            closureScope: this.captureScope()
        });
    }
    importPackageMethods(importPath, source) {
        for (const method of source.shared.methods.values()) {
            const receiverType = qualifyLocalRuntimeTypeName(method.receiverType, importPath);
            const declaration = qualifyRuntimeMethodDeclaration(method.declaration, importPath);
            this.shared.methods.set(methodKey(receiverType, method.declaration.name), {
                declaration,
                receiverType,
                pointerReceiver: method.pointerReceiver,
                closureScope: method.closureScope
            });
        }
    }
    importPackageDefinitions(importPath, source) {
        for (const [name, target] of source.shared.aliases.entries()) {
            this.shared.aliases.set(qualifyLocalRuntimeTypeName(name, importPath), qualifyLocalRuntimeTypeName(target, importPath));
        }
        for (const [name, typeDef] of source.shared.types.entries()) {
            this.shared.types.set(qualifyLocalRuntimeTypeName(name, importPath), {
                name: qualifyLocalRuntimeTypeName(typeDef.name, importPath),
                fields: typeDef.fields.map((field) => ({
                    ...field,
                    type: { text: qualifyLocalRuntimeTypeName(field.type.text, importPath) }
                }))
            });
        }
        for (const [name, interfaceDef] of source.shared.interfaces.entries()) {
            this.shared.interfaces.set(qualifyLocalRuntimeTypeName(name, importPath), {
                name: qualifyLocalRuntimeTypeName(interfaceDef.name, importPath),
                methods: interfaceDef.methods.map((method) => ({
                    ...method,
                    signature: qualifyRuntimeSignature(method.signature, importPath)
                })),
                embeds: interfaceDef.embeds.map((embed) => ({ text: qualifyLocalRuntimeTypeName(embed.text, importPath) }))
            });
        }
    }
    packageContext(importPath) {
        return this.options.packageContexts?.[importPath];
    }
    registerImportBinding(localName, importPath) {
        if (localName === "_" || localName === ".")
            return;
        this.shared.importPathsByLocalName.set(localName, importPath);
    }
    resolveImportedTypeText(typeText) {
        return rewriteRuntimeTypeText(typeText, (qualifier) => this.shared.importPathsByLocalName.get(qualifier));
    }
    methodFor(typeName, methodName) {
        return this.shared.methods.get(methodKey(typeName, methodName)) ??
            this.shared.methods.get(methodKey(genericBaseTypeName(typeName), methodName));
    }
    flattenInterfaceMethods(methods, embeds) {
        const flattened = [...methods];
        for (const embed of embeds) {
            const embedded = this.interfaceDef(embed.text);
            if (!embedded)
                continue;
            for (const method of embedded.methods) {
                if (!flattened.some((candidate) => candidate.name === method.name))
                    flattened.push(method);
            }
        }
        return flattened;
    }
    currentDeferFrame() {
        const frame = this.deferFrames[this.deferFrames.length - 1];
        if (!frame)
            throw new GoJuniorRuntimeError("internal error: missing defer frame");
        return frame;
    }
    assignmentChecker() {
        return (value, typeText, role) => prepareAssignableToType(value, typeText, role, this);
    }
}
class Scope {
    parent;
    bindings = new Map();
    constructor(parent) {
        this.parent = parent;
    }
    declare(name, value, mutable, typeText) {
        if (this.bindings.has(name)) {
            throw new GoJuniorRuntimeError(`${name} already declared`);
        }
        this.bindings.set(name, { value, mutable, ...(typeText ? { typeText } : {}) });
    }
    hasLocal(name) {
        return this.bindings.has(name);
    }
    hasBinding(name) {
        return this.bindings.has(name) || Boolean(this.parent?.hasBinding(name));
    }
    assign(name, value, checker) {
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
    lookup(name) {
        const binding = this.bindings.get(name);
        if (binding)
            return binding.value;
        if (this.parent)
            return this.parent.lookup(name);
        throw new GoJuniorRuntimeError(`${name} is not declared`);
    }
    lookupTypeText(name) {
        const binding = this.bindings.get(name);
        if (binding)
            return binding.typeText;
        return this.parent?.lookupTypeText(name);
    }
    isUntypedConstant(name) {
        const binding = this.bindings.get(name);
        if (binding)
            return !binding.mutable && binding.typeText === undefined;
        return this.parent?.isUntypedConstant(name) ?? false;
    }
}
class SheetBinding {
    name;
    data;
    constructor(name, data) {
        this.name = name;
        this.data = data;
    }
    get(cell) {
        return this.data[normalizeCell(cell)] ?? null;
    }
    set(cell, value) {
        this.data[normalizeCell(cell)] = value;
    }
    range(start, end) {
        const startCell = parseCellAddress(start);
        const endCell = parseCellAddress(end);
        const values = [];
        const rowStep = startCell.row <= endCell.row ? 1 : -1;
        const colStep = startCell.col <= endCell.col ? 1 : -1;
        for (let row = startCell.row; row !== endCell.row + rowStep; row += rowStep) {
            const rowValues = [];
            for (let col = startCell.col; col !== endCell.col + colStep; col += colStep) {
                rowValues.push(this.get(formatCellAddress(col, row)));
            }
            values.push(rowValues);
        }
        return values;
    }
}
export class RuntimeStruct {
    typeName;
    fields = new Map();
    constructor(typeName, initialFields = []) {
        this.typeName = typeName;
        for (const [name, value] of initialFields) {
            this.fields.set(name, value);
        }
    }
    get(field) {
        return this.fields.get(field);
    }
    set(field, value) {
        this.fields.set(field, value);
    }
    clone() {
        return new RuntimeStruct(this.typeName, this.fields.entries());
    }
    orderedFields() {
        return [...this.fields.entries()];
    }
}
class RuntimeExternalizedStruct extends RuntimeStruct {
    source;
    packageName;
    constructor(source, packageName) {
        super(qualifyLocalRuntimeTypeName(source.typeName, packageName));
        this.source = source;
        this.packageName = packageName;
    }
    get(field) {
        const value = this.source.get(field);
        return value === undefined ? undefined : externalizePackageRuntimeValue(value, this.packageName);
    }
    set(field, value) {
        this.source.set(field, value);
    }
    clone() {
        return new RuntimeStruct(this.typeName, this.orderedFields());
    }
    orderedFields() {
        return this.source.orderedFields().map(([name, value]) => [
            name,
            externalizePackageRuntimeValue(value, this.packageName)
        ]);
    }
}
export class RuntimePointer {
    typeName;
    getValue;
    setValue;
    constructor(typeName, getValue, setValue) {
        this.typeName = typeName;
        this.getValue = getValue;
        this.setValue = setValue;
    }
    get() {
        return this.getValue();
    }
    set(value) {
        this.setValue(value);
    }
}
export class RuntimeNamedValue {
    typeName;
    value;
    constructor(typeName, value) {
        this.typeName = typeName;
        this.value = value;
    }
}
export class RuntimeInterfaceValue {
    interfaceType;
    value;
    constructor(interfaceType, value) {
        this.interfaceType = interfaceType;
        this.value = value;
    }
    isNil() {
        return this.value === null;
    }
}
export class RuntimeTypedNilValue {
    typeName;
    constructor(typeName) {
        this.typeName = typeName;
    }
}
const objectMapKeyIds = new WeakMap();
const arrayCapacities = new WeakMap();
const arrayTypeTexts = new WeakMap();
const arrayViews = new WeakMap();
let nextObjectMapKeyId = 1;
let nextNonReflexiveMapKeyId = 1;
export class RuntimeMap {
    keyType;
    valueType;
    context;
    entries = new Map();
    constructor(keyType, valueType, context) {
        this.keyType = keyType;
        this.valueType = valueType;
        this.context = context;
        assertComparableType(keyType, context);
    }
    get(key) {
        key = prepareAssignableToType(key, this.keyType, "map key", this.context);
        const entry = this.entries.get(runtimeMapKeyId(key));
        return entry ? entry.value : zeroValueForMapValue(this.valueType);
    }
    getWithPresence(key) {
        key = prepareAssignableToType(key, this.keyType, "map key", this.context);
        const entry = this.entries.get(runtimeMapKeyId(key));
        return entry ? [entry.value, true] : [zeroValueForMapValue(this.valueType), false];
    }
    set(key, value) {
        key = prepareAssignableToType(key, this.keyType, "map key", this.context);
        assertComparableValue(key, "map key");
        value = prepareAssignableToType(value, this.valueType, "map value", this.context);
        this.entries.set(runtimeMapKeyId(key, { freshNonReflexive: true }), { key, value });
    }
    delete(key) {
        key = prepareAssignableToType(key, this.keyType, "map key", this.context);
        this.entries.delete(runtimeMapKeyId(key));
    }
    clear() {
        this.entries.clear();
    }
    orderedEntries() {
        return [...this.entries.values()].map((entry) => [entry.key, entry.value]);
    }
    size() {
        return this.entries.size;
    }
}
export class RuntimeChannel {
    elementType;
    capacity;
    context;
    channel;
    constructor(elementType, capacity, context) {
        this.elementType = elementType;
        this.capacity = capacity;
        this.context = context;
        this.channel = new AsyncGoChannel(context?.scheduler() ?? new AsyncGoScheduler(), capacity, () => defaultValueForTypeText(elementType, context));
    }
    send(value) {
        try {
            if (!this.channel.trySend(prepareAssignableToType(value, this.elementType, "channel send", this.context))) {
                throw new GoJuniorDeadlockError("send on channel would block");
            }
        }
        catch (error) {
            throw normalizeAsyncRuntimeError(error);
        }
    }
    async sendAsync(value) {
        try {
            await this.channel.send(prepareAssignableToType(value, this.elementType, "channel send", this.context));
        }
        catch (error) {
            if (isDeadlockError(error))
                throw new GoJuniorDeadlockError("send on channel would block");
            throw normalizeAsyncRuntimeError(error);
        }
    }
    receive() {
        try {
            const ready = this.channel.tryReceive();
            if (!ready)
                throw new GoJuniorDeadlockError("receive from channel would block");
            return ready;
        }
        catch (error) {
            throw normalizeAsyncRuntimeError(error);
        }
    }
    async receiveAsync() {
        try {
            return await this.channel.receive();
        }
        catch (error) {
            if (isDeadlockError(error))
                throw new GoJuniorDeadlockError("receive from channel would block");
            throw normalizeAsyncRuntimeError(error);
        }
    }
    canSend() {
        return this.channel.canSendNow();
    }
    canReceive() {
        return this.channel.canReceiveNow();
    }
    close() {
        try {
            this.channel.close();
        }
        catch (error) {
            throw normalizeAsyncRuntimeError(error);
        }
    }
    len() {
        return this.channel.len();
    }
    cap() {
        return this.capacity;
    }
    asyncChannel() {
        return this.channel;
    }
}
export async function evaluateSource(source, options = {}) {
    return evaluateSourceFiles([sourceFileFromSource(source, options)], options);
}
export async function evaluateSourceFiles(files, options = {}) {
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
export async function evaluatePackageSourceFiles(files, options = {}) {
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
    const packageName = options.packageName ?? packageNameFromParsedFiles(parsed.parsed.files) ?? "main";
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
            for (const declaration of ast.functions)
                installPackageFunctionDeclaration(context, declaration);
            const { declarations, statements } = splitTopLevelDeclarations(ast.body);
            if (statements.length > 0) {
                throw new GoJuniorRuntimeError("package source cannot contain top-level executable statements");
            }
            predeclareTopLevelTypes(declarations, context);
            predeclarePackageConstants(checked.pkg, declarations, context);
            predeclarePackageVariables(declarations, context);
            const runtimeDeclarations = declarations.filter((declaration) => declaration.kind === "TypeDecl");
            const declarationCompletion = await executeTopLevelStatements(runtimeDeclarations, context);
            expectNormalCompletion(declarationCompletion, "package declarations");
            await executePackageVarInitializers(declarations, checked.info.InitOrder, context);
            await runInitFunctions(ast.functions, context);
            return exportedRuntimePackageObject(packageScopeObjects(checked.pkg), context, options.importPath ?? packageName);
        });
        return {
            diagnostics: checked.diagnostics,
            output: context.output,
            package: pkg,
            packageInfo: checked.pkg,
            context
        };
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
            diagnostics: [
                ...checked.diagnostics,
                runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message, error)
            ],
            output: context.output
        };
    }
}
export async function evaluateSourcePackageGraph(specs, options = {}) {
    const evaluator = new SourcePackageGraphEvaluator(specs, options);
    return evaluator.evaluate();
}
class SourcePackageGraphEvaluator {
    options;
    specsByPath = new Map();
    importsByPath = new Map();
    diagnostics = [];
    output = [];
    packages;
    packageInfos;
    packageContexts = {};
    initialized = new Set();
    initializedImportPaths = [];
    constructor(specs, options) {
        this.options = options;
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
    async evaluate() {
        this.discoverSourcePackages();
        if (this.hasErrors())
            return this.result();
        const sortedImportPaths = () => [...this.specsByPath.keys()].sort();
        while (this.initialized.size < this.specsByPath.size) {
            let progressed = false;
            for (const importPath of sortedImportPaths()) {
                if (this.initialized.has(importPath))
                    continue;
                const imports = this.importsByPath.get(importPath) ?? [];
                const pendingSourceImports = imports.filter((dependency) => this.specsByPath.has(dependency) && !this.initialized.has(dependency));
                if (pendingSourceImports.length > 0)
                    continue;
                await this.initializePackage(importPath);
                progressed = true;
                break;
            }
            if (this.hasErrors())
                return this.result();
            if (!progressed) {
                const remaining = sortedImportPaths().filter((importPath) => !this.initialized.has(importPath));
                this.diagnostics.push(packageGraphDiagnostic(this.specsByPath.get(remaining[0] ?? "")?.files[0]?.filename ?? REPL_FILENAME, `package initialization cycle detected among source packages: ${remaining.join(", ")}`));
                return this.result();
            }
        }
        return this.result();
    }
    discoverSourcePackages() {
        for (const importPath of [...this.specsByPath.keys()].sort()) {
            this.discoverOne(importPath, []);
            if (this.hasErrors())
                return;
        }
    }
    discoverOne(importPath, stack) {
        if (this.importsByPath.has(importPath) || this.hasErrors())
            return;
        if (stack.includes(importPath)) {
            this.diagnostics.push(packageGraphDiagnostic(this.specsByPath.get(importPath)?.files[0]?.filename ?? REPL_FILENAME, `package import cycle detected: ${[...stack, importPath].join(" -> ")}`));
            return;
        }
        const spec = this.specsByPath.get(importPath);
        if (!spec)
            return;
        const parsed = frontSourceFilesToAst(spec.files);
        this.diagnostics.push(...parsed.diagnostics);
        if (this.hasErrors())
            return;
        const imports = uniqueSortedSourceImports(parsed.ast?.imports.map((imported) => imported.path) ?? []);
        this.importsByPath.set(importPath, imports);
        for (const dependency of imports) {
            if (isHostResolvedSourceImport(dependency))
                continue;
            if (this.packages[dependency])
                continue;
            if (!this.specsByPath.has(dependency)) {
                let loaded;
                try {
                    loaded = this.options.sourcePackageProvider?.load(dependency);
                }
                catch (error) {
                    const message = error instanceof Error ? error.message : String(error);
                    this.diagnostics.push(packageGraphDiagnostic(spec.files[0]?.filename ?? REPL_FILENAME, `could not load package ${dependency}: ${message}`));
                    return;
                }
                if (loaded) {
                    this.specsByPath.set(dependency, { importPath: dependency, files: loaded });
                }
            }
            if (!this.specsByPath.has(dependency))
                continue;
            this.discoverOne(dependency, [...stack, importPath]);
            if (this.hasErrors())
                return;
        }
    }
    async initializePackage(importPath) {
        const spec = this.specsByPath.get(importPath);
        if (!spec)
            return;
        const result = await evaluatePackageSourceFiles(spec.files, {
            ...this.options,
            importPath,
            ...(spec.packageName ? { packageName: spec.packageName } : {}),
            packages: this.packages,
            packageInfos: this.packageInfos,
            packageContexts: this.packageContexts
        });
        this.output.push(...result.output);
        this.diagnostics.push(...result.diagnostics);
        if (this.hasErrors())
            return;
        const pkg = result.package ?? {};
        this.packages[importPath] = pkg;
        if (result.packageInfo) {
            this.packageInfos[importPath] = result.packageInfo;
        }
        if (result.context) {
            this.packageContexts[importPath] = result.context;
        }
        this.initialized.add(importPath);
        this.initializedImportPaths.push(importPath);
    }
    hasErrors() {
        return this.diagnostics.some((diagnostic) => diagnostic.severity === "error");
    }
    result() {
        return {
            diagnostics: this.diagnostics,
            output: this.output,
            packages: this.packages,
            packageInfos: this.packageInfos,
            packageContexts: this.packageContexts,
            initializedImportPaths: this.initializedImportPaths
        };
    }
}
function uniqueSortedSourceImports(values) {
    return [...new Set(values)].sort();
}
function packageGraphDiagnostic(filename, message) {
    return {
        filename,
        code: "GOJR_RUNTIME001",
        severity: "error",
        message
    };
}
export async function runMainSourcePackageFiles(files, options = {}) {
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
            diagnostics: [packageGraphDiagnostic(files[0]?.filename ?? REPL_FILENAME, "no package source files supplied")],
            output: []
        };
    }
    const packageName = options.packageName ?? packageNameFromParsedFiles(parsed.parsed.files) ?? "main";
    if (packageName !== "main") {
        return {
            diagnostics: [packageGraphDiagnostic(files[0]?.filename ?? REPL_FILENAME, `gojr run requires package main, found package ${packageName}`)],
            output: [],
            ast
        };
    }
    const importPath = options.importPath ?? packageName;
    const sourcePackages = (options.sourcePackages ?? []).filter((spec) => spec.importPath !== importPath);
    const { importPath: _rootImportPath, packageName: _rootPackageName, sourcePackages: _sourcePackages, ...graphOptions } = options;
    const graph = await evaluateSourcePackageGraph([
        ...sourcePackages,
        { importPath, packageName, files }
    ], graphOptions);
    if (graph.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return {
            diagnostics: graph.diagnostics,
            output: graph.output,
            ast
        };
    }
    const context = graph.packageContexts[importPath];
    if (!context) {
        return {
            diagnostics: [packageGraphDiagnostic(files[0]?.filename ?? REPL_FILENAME, `package ${importPath} did not produce a runtime context`)],
            output: graph.output,
            ast
        };
    }
    const outputOffset = context.output.length;
    try {
        const main = context.lookup("main");
        await context.scheduler().runRoot(async () => {
            await callRuntime(main, [], context);
        });
        return {
            diagnostics: graph.diagnostics,
            output: [...graph.output, ...context.output.slice(outputOffset)],
            ast
        };
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
            diagnostics: [
                ...graph.diagnostics,
                runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message, error)
            ],
            output: [...graph.output, ...context.output.slice(outputOffset)],
            ast
        };
    }
}
export async function testSource(source, options = {}) {
    return testSourceFiles([sourceFileFromSource(source, options)], options);
}
export async function testSourceFiles(files, options = {}) {
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
function sourceFileFromSource(source, options = {}) {
    return {
        filename: options.filename ?? REPL_FILENAME,
        source
    };
}
function packageNameFromParsedFiles(files) {
    return files.find((file) => file.name)?.name?.name;
}
export async function evaluateProgram(ast, options = {}) {
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
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return withObservedDeps({
            diagnostics: [
                ...ast.diagnostics,
                runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message, error)
            ],
            output: context.output,
            ast
        }, context);
    }
}
function exportedRuntimePackageObject(objects, context, importPath) {
    const pkg = {};
    for (const object of objects) {
        if (!object.Exported())
            continue;
        try {
            pkg[object.Name()] = externalizePackageRuntimeValue(context.lookup(object.Name()), importPath);
        }
        catch {
            // Type-only exports have no runtime value in the interpreter package object.
        }
    }
    return pkg;
}
function packageScopeObjects(pkg) {
    return pkg.Scope().Names().flatMap((name) => {
        if (isGoJuniorSyntheticCheckName(name) || name === "fmt")
            return [];
        const object = pkg.Scope().Lookup(name);
        if (object === null || object.constructor.name === "PkgName")
            return [];
        if (object.Pkg() !== pkg)
            return [];
        return [object];
    });
}
async function testProgram(ast, baseDiagnostics, options) {
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
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return withObservedDeps({
            diagnostics: [
                ...diagnostics,
                runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message, error)
            ],
            output: context.output,
            ast
        }, context);
    }
}
function isTestFunctionDecl(declaration) {
    return !declaration.receiver && /^Test($|[^a-z])/.test(declaration.name);
}
async function runOneTest(declaration, context, diagnostics) {
    context.write(`=== RUN   ${declaration.name}\n`);
    const signatureError = testSignatureError(declaration);
    if (signatureError) {
        context.write(`    ${signatureError}\n`);
        context.write(`--- FAIL: ${declaration.name}\n`);
        diagnostics.push(testDiagnostic(declaration, `${declaration.name}: ${signatureError}`));
        return true;
    }
    const testingT = declaration.signature.parameters.length === 0 ? undefined : makeTestingT(declaration.name);
    let runtimeFailure;
    try {
        const callee = context.lookup(declaration.name);
        await callRuntime(callee, testingT ? [testingT.value] : [], context);
    }
    catch (error) {
        if (error instanceof GoJuniorTestStop) {
            // The testing.T state already records whether this was FailNow or SkipNow.
        }
        else {
            runtimeFailure = error instanceof Error ? error.message : String(error);
            if (testingT) {
                testingT.state.failed = true;
                testingT.state.logs.push(runtimeFailure);
            }
        }
    }
    if (testingT)
        emitTestingLogs(context, testingT.state);
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
function testSignatureError(declaration) {
    if (declaration.signature.results.length !== 0)
        return "test function must not return values";
    const parameters = declaration.signature.parameters;
    if (parameters.length === 0)
        return undefined;
    if (parameters.length === 1 && !parameters[0]?.variadic && normalizeTypeText(parameters[0]?.type.text ?? "") === "*testing.T") {
        return undefined;
    }
    return "test function must have signature func TestX() or func TestX(t *testing.T)";
}
function testDiagnostic(declaration, message) {
    return {
        filename: declaration.span?.filename ?? REPL_FILENAME,
        code: "GOJR_TEST001",
        severity: "error",
        message,
        ...(declaration.span ? { span: declaration.span } : {})
    };
}
function runtimeDiagnostic(ast, code, message, error) {
    const span = error instanceof GoJuniorRuntimeError && error.span
        ? error.span
        : firstProgramSpan(ast);
    return {
        filename: span?.filename ?? REPL_FILENAME,
        code,
        severity: "error",
        message,
        ...(error instanceof Error && error.stack ? { stack: error.stack } : {}),
        ...(span ? { span } : {})
    };
}
function runtimeDiagnosticCode(error) {
    if (error instanceof GoJuniorPanic)
        return "GOJR_PANIC001";
    if (error instanceof AsyncGoPanic)
        return "GOJR_PANIC001";
    if (error instanceof GoJuniorDeadlockError || error instanceof AsyncGoDeadlockError)
        return "GOJR_DEADLOCK001";
    return "GOJR_RUNTIME001";
}
function normalizeAsyncRuntimeError(error) {
    if (error instanceof AsyncGoPanic)
        return new GoJuniorPanic(error.value);
    if (error instanceof AsyncGoDeadlockError)
        return new GoJuniorDeadlockError(error.message);
    return error;
}
function isDeadlockError(error) {
    return error instanceof GoJuniorDeadlockError || error instanceof AsyncGoDeadlockError;
}
function firstProgramSpan(ast) {
    return ast.body[0]?.span ?? ast.functions[0]?.span;
}
function makeTestingT(name) {
    const state = {
        name,
        failed: false,
        skipped: false,
        logs: []
    };
    const object = {
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
function testingSprint(args) {
    return sprint(args);
}
function testingSprintf(args) {
    return sprintf(toStringValue(args[0] ?? ""), args.slice(1));
}
function appendTestingLog(state, text) {
    state.logs.push(text.endsWith("\n") ? text : `${text}\n`);
}
function emitTestingLogs(context, state) {
    for (const log of state.logs) {
        context.write(indentTestingLog(log));
    }
}
function indentTestingLog(text) {
    const lineText = text.endsWith("\n") ? text.slice(0, -1) : text;
    if (lineText === "")
        return "    \n";
    return lineText.split("\n").map((line) => `    ${line}\n`).join("");
}
export class GoJuniorSession {
    options;
    context;
    checkerPackage;
    sessionSheet;
    sessionSheets;
    checkSequence = 0;
    constructor(options = {}) {
        this.options = options;
        this.sessionSheet = options.sheet;
        this.sessionSheets = options.sheets ? { ...options.sheets } : undefined;
        ensureUniverseInitialized();
        this.checkerPackage = NewPackage("main", "main");
        this.context = new EvaluationContext(options);
        installSheets(this.context, options);
    }
    setSheet(sheet) {
        this.sessionSheet = sheet;
        this.context.declareOrAssignRoot("sheet", new SheetBinding("sheet", sheet), true);
    }
    setSheets(sheets) {
        this.sessionSheets = {
            ...(this.sessionSheets ?? {}),
            ...sheets
        };
        for (const [name, data] of Object.entries(sheets)) {
            this.context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
        }
    }
    async evaluate(source) {
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
                this.persistCheckerImports(ast);
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
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            return withObservedDeps({
                diagnostics: [
                    ...ast.diagnostics,
                    runtimeDiagnostic(ast, runtimeDiagnosticCode(error), message, error)
                ],
                output: this.context.outputFrom(outputStart),
                ast
            }, this.context);
        }
    }
    checkSource(source) {
        const checked = checkGoJuniorSourceFiles([ensureTrailingNewlineSourceFile(source)], {
            ...typeCheckConfig(this.currentOptions()),
            packageInstance: this.checkerPackage,
            syntheticFunctionName: `${GOJR_SYNTHETIC_CHECK_PREFIX}_${++this.checkSequence}`
        });
        return checked;
    }
    acceptTypeInfo(_checked) {
        // The session owns a persistent go/types.Package, so successful checks have
        // already extended its package scope. Keeping this hook makes the call sites
        // spell out when checked declarations become visible to later evaluations.
    }
    persistCheckerImports(ast) {
        for (const imported of ast.imports) {
            const name = importBindingName(imported, this.context);
            if (name === "_" || name === ".")
                continue;
            if (this.checkerPackage.Scope().Lookup(name) !== null)
                continue;
            const pkg = this.options.packageInfos?.[imported.path]
                ?? standardTypePackage(imported.path);
            if (pkg !== undefined) {
                this.checkerPackage.Scope().Insert(NewPkgName(NoPos, this.checkerPackage, name, pkg));
            }
        }
    }
    prepareFunctionRedeclarations(ast, filename) {
        const diagnostics = [];
        const replacements = [];
        const scope = this.checkerPackage.Scope();
        for (const declaration of ast.functions) {
            if (declaration.receiver || declaration.name === "init")
                continue;
            const existingObject = scope.Lookup(declaration.name);
            if (existingObject === null)
                continue;
            const existingValue = this.context.hasBinding(declaration.name) ? this.context.lookup(declaration.name) : undefined;
            if (existingValue === undefined ||
                !isGoJuniorFunction(existingValue) ||
                existingValue.signature === undefined ||
                !signaturesCompatible(existingValue.signature, declaration.signature)) {
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
    restoreFunctionRedeclarations(replacements) {
        const scope = this.checkerPackage.Scope();
        for (const replacement of replacements) {
            if (scope.Lookup(replacement.name) === null) {
                scope.insert(replacement.name, replacement.object);
            }
        }
    }
    currentOptions() {
        return {
            ...this.options,
            ...(this.sessionSheet ? { sheet: this.sessionSheet } : {}),
            ...(this.sessionSheets ? { sheets: this.sessionSheets } : {})
        };
    }
}
function ensureTrailingNewline(source) {
    return source.endsWith("\n") ? source : `${source}\n`;
}
function ensureTrailingNewlineSourceFile(file) {
    return {
        filename: file.filename,
        source: ensureTrailingNewline(file.source)
    };
}
export function typeCheckConfig(options) {
    return {
        sheetNamespaces: sheetNamespacesForOptions(options),
        ...(options.packageInfos ? {
            importer: {
                import: (path) => options.packageInfos?.[path]
            }
        } : {})
    };
}
function sheetNamespacesForOptions(options) {
    const namespaces = {
        sheet: sheetNamespaceForData(options.sheet ?? options.sheets?.[options.currentSheetName ?? "sheet"] ?? {})
    };
    for (const [name, data] of Object.entries(options.sheets ?? {})) {
        namespaces[name] = sheetNamespaceForData(data);
    }
    return namespaces;
}
function sheetNamespaceForData(data) {
    void data;
    return {};
}
function resultFromCompletion(ast, output, completion) {
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
    if (completion.kind !== "normal")
        throw new GoJuniorRuntimeError(completionErrorMessage(completion));
    return {
        diagnostics: ast.diagnostics,
        output,
        ast,
        ...(completion.value !== undefined ? { value: externalizeRuntimeValue(completion.value) } : {})
    };
}
function withObservedDeps(result, context) {
    return {
        ...result,
        observedDeps: context.observedDeps()
    };
}
function diagnosticsLookIncomplete(diagnostics) {
    const errors = diagnostics.filter((diagnostic) => diagnostic.severity === "error");
    return errors.length > 0 && errors.every((diagnostic) => {
        if (diagnostic.code === "GOJR_PARSE001") {
            return /found\s+-->\s*''\s*<--/.test(diagnostic.message) || /but found:\s*''/.test(diagnostic.message);
        }
        if (diagnostic.code === "GOJR_SCAN001") {
            return /unterminated .*string literal/i.test(diagnostic.message);
        }
        if (diagnostic.code !== "GOJR_PARSE_FRONT001")
            return false;
        return diagnostic.span?.length === 0 &&
            (/expected/i.test(diagnostic.message) || /found EOF/i.test(diagnostic.message));
    });
}
function installBuiltins(context) {
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
        if (!(target instanceof RuntimeMap))
            throw new GoJuniorRuntimeError("delete expects a map");
        target.delete(args[1] ?? null);
        return null;
    }));
    context.declareRoot("close", hostCallable("close", (args) => {
        const target = args[0] ?? null;
        if (target === null)
            throw new GoJuniorPanic("close of nil channel");
        if (!(target instanceof RuntimeChannel))
            throw new GoJuniorRuntimeError("close expects a channel");
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
        for (const arg of args)
            context.write(formatValue(arg));
        return null;
    }));
    context.declareRoot("println", hostCallable("println", (args, context) => {
        context.write(`${args.map(formatValue).join(" ")}\n`);
        return null;
    }));
}
function installSheets(context, options) {
    context.declareOrAssignRoot("sheet", new SheetBinding("sheet", options.sheet ?? {}), true);
    for (const [name, data] of Object.entries(options.sheets ?? {})) {
        if (name !== "sheet") {
            context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
        }
    }
}
function installImports(context, ast) {
    const packages = availablePackages(context);
    for (const imported of ast.imports) {
        const name = importBindingName(imported, context);
        const pkg = packages[imported.path];
        if (!pkg) {
            throw new GoJuniorRuntimeError(`package ${imported.path} is not available`);
        }
        context.registerImportBinding(name, imported.path);
        const importedContext = context.packageContext(imported.path);
        if (importedContext) {
            context.importPackageDefinitions(imported.path, importedContext);
            context.importPackageMethods(imported.path, importedContext);
        }
        if (name === "_")
            continue;
        if (name === ".") {
            for (const [exportName, value] of Object.entries(pkg)) {
                context.declareOrAssignRoot(exportName, value, true);
            }
            continue;
        }
        context.declareOrAssignRoot(name, pkg, true);
    }
}
function importDefaultName(path) {
    return path.split("/").filter(Boolean).at(-1) ?? path;
}
function importBindingName(imported, context) {
    return imported.alias
        ?? context?.packageInfo(imported.path)?.Name()
        ?? standardTypePackage(imported.path)?.Name()
        ?? importDefaultName(imported.path);
}
function installFunctionDeclaration(context, declaration) {
    if (declaration.receiver) {
        context.registerMethod(resolveRuntimeFunctionDeclarationTypeImports(declaration, context));
        return;
    }
    const resolved = resolveRuntimeFunctionDeclarationTypeImports(declaration, context);
    context.declareOrAssignRoot(resolved.name, functionValue(resolved), true);
}
function installPackageFunctionDeclaration(context, declaration) {
    const resolved = resolveRuntimeFunctionDeclarationTypeImports(declaration, context);
    if (declaration.receiver) {
        context.registerMethod(resolved);
        return;
    }
    const intrinsic = bodylessPackageFunctionIntrinsic(resolved, context.importPath());
    context.declareOrAssignRoot(resolved.name, intrinsic ?? goJuniorFunctionValue(resolved.name, resolved.signature, resolved.body, context.captureScope(), resolved, undefined, undefined, context), true);
}
function bodylessPackageFunctionIntrinsic(declaration, importPath) {
    if (!isBodylessFunctionDeclaration(declaration))
        return undefined;
    const intrinsic = bodylessBytealgIntrinsic(importPath, declaration.name, declaration.signature) ??
        bodylessAbiIntrinsic(importPath, declaration.name, declaration.signature) ??
        bodylessRuntimeIntrinsic(declaration.name, declaration.signature);
    return intrinsic
        ? {
            ...intrinsic,
            ...(declaration.source ? { source: declaration.source } : {}),
            declaration
        }
        : bodylessZeroResultFunction(declaration, importPath);
}
function isBodylessFunctionDeclaration(declaration) {
    return declaration.body.statements.length === 0 && declaration.source !== undefined && !declaration.source.includes("{");
}
function bodylessRuntimeIntrinsic(name, signature) {
    const functionValue = (result, intrinsicName = name) => ({
        kind: "GoJuniorFunction",
        name: intrinsicName,
        signature,
        async call() {
            return result;
        }
    });
    switch (name) {
        case "now":
        case "runtimeNow":
            return functionValue([0n, 0n, 1n]);
        case "runtimeNano":
            return functionValue(1n);
        case "runtimeIsBubbled":
            return functionValue(false);
        case "Sleep":
            return functionValue(null);
        case "newTimer":
            return functionValue(null);
        case "stopTimer":
        case "resetTimer":
            return functionValue(false);
        default:
            return undefined;
    }
}
function bodylessAbiIntrinsic(importPath, name, signature) {
    if (importPath !== "internal/abi")
        return undefined;
    if (name !== "FuncPCABI0" && name !== "FuncPCABIInternal")
        return undefined;
    return {
        kind: "GoJuniorFunction",
        name: `${importPath}.${name}`,
        signature,
        async call(args) {
            const value = unwrapNamed(args[0] ?? null);
            if (value === null)
                return 0n;
            if (typeof value === "object")
                return BigInt(objectIdentityId(value));
            return BigInt(runtimeMapKeyId(value).length);
        }
    };
}
function bodylessZeroResultFunction(declaration, importPath) {
    return {
        kind: "GoJuniorFunction",
        name: importPath ? `${importPath}.${declaration.name}` : declaration.name,
        signature: declaration.signature,
        ...(declaration.source ? { source: declaration.source } : {}),
        declaration,
        async call(_args, context) {
            const values = declaration.signature.results.map((result) => defaultValueForDeclarationType(result.type, context));
            return values.length === 0 ? null : values.length === 1 ? values[0] ?? null : values;
        }
    };
}
function bodylessBytealgIntrinsic(importPath, name, signature) {
    if (importPath !== "internal/bytealg")
        return undefined;
    const functionValue = (call) => ({
        kind: "GoJuniorFunction",
        name: `${importPath}.${name}`,
        signature,
        async call(args, context) {
            return call(args, context);
        }
    });
    switch (name) {
        case "IndexByte":
        case "IndexByteString":
            return functionValue((args) => BigInt(byteIndex(bytealgBytes(args[0] ?? null), bytealgByte(args[1] ?? 0n))));
        case "Index":
        case "IndexString":
            return functionValue((args) => BigInt(byteSequenceIndex(bytealgBytes(args[0] ?? null), bytealgBytes(args[1] ?? null))));
        case "Count":
        case "CountString":
            return functionValue((args) => BigInt(byteCount(bytealgBytes(args[0] ?? null), bytealgByte(args[1] ?? 0n))));
        case "Compare":
        case "CompareString":
        case "abigen_runtime_cmpstring":
            return functionValue((args) => BigInt(byteSequenceCompare(bytealgBytes(args[0] ?? null), bytealgBytes(args[1] ?? null))));
        case "MakeNoZero":
            return functionValue((args, context) => {
                const length = toNonNegativeLength(args[0] ?? 0n, "internal/bytealg.MakeNoZero length");
                return makeRuntimeSlice("byte", length, length, context, "[]byte");
            });
        case "abigen_runtime_memequal":
            return functionValue((args) => toBigInt(args[2] ?? 0n) === 0n || valueEqual(args[0] ?? null, args[1] ?? null));
        case "abigen_runtime_memequal_varlen":
            return functionValue((args) => valueEqual(args[0] ?? null, args[1] ?? null));
        default:
            return undefined;
    }
}
function bytealgBytes(value) {
    value = unwrapNamed(value);
    if (isRuntimeString(value))
        return goStringBytes(value);
    if (value instanceof RuntimeTypedNilValue && parseArrayOrSliceTypeText(value.typeName))
        return new Uint8Array();
    if (Array.isArray(value)) {
        return Uint8Array.from(value.map((item) => Number(BigInt.asUintN(8, toBigInt(item)))));
    }
    throwTypeError(value, "string or []byte", "internal/bytealg argument");
}
function bytealgByte(value) {
    return Number(BigInt.asUintN(8, toBigInt(value)));
}
function byteIndex(values, needle) {
    for (let index = 0; index < values.length; index += 1) {
        if (values[index] === needle)
            return index;
    }
    return -1;
}
function byteCount(values, needle) {
    let count = 0;
    for (const value of values) {
        if (value === needle)
            count += 1;
    }
    return count;
}
function byteSequenceIndex(values, needle) {
    if (needle.length === 0)
        return 0;
    if (needle.length > values.length)
        return -1;
    const lastStart = values.length - needle.length;
    for (let start = 0; start <= lastStart; start += 1) {
        let matched = true;
        for (let offset = 0; offset < needle.length; offset += 1) {
            if (values[start + offset] !== needle[offset]) {
                matched = false;
                break;
            }
        }
        if (matched)
            return start;
    }
    return -1;
}
function byteSequenceCompare(left, right) {
    const count = Math.min(left.length, right.length);
    for (let index = 0; index < count; index += 1) {
        const leftByte = left[index] ?? 0;
        const rightByte = right[index] ?? 0;
        if (leftByte !== rightByte)
            return leftByte < rightByte ? -1 : 1;
    }
    return left.length === right.length ? 0 : left.length < right.length ? -1 : 1;
}
function installedFunctionValue(context, declaration) {
    return declaration.receiver ? functionValue(declaration) : context.lookup(declaration.name);
}
function bindingTypeText(type) {
    const text = typeof type === "string" ? type : type?.text;
    const normalized = text ? normalizeTypeText(text) : "";
    return normalized && normalized !== "<missing>" ? normalized : undefined;
}
function resolvedDeclaredTypeText(typeText, context) {
    return parseArrayOrSliceTypeText(typeText, context)?.typeText ?? typeText;
}
function installAutomaticImports(context) {
    context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}
function availablePackages(context) {
    return {
        cmp: cmpPackage(),
        fmt: fmtPackage(),
        "internal/reflectlite": reflectlitePackage(),
        math: mathPackage(),
        os: osPackage(),
        runtime: runtimePackage(),
        "runtime/pprof": runtimePprofPackage(),
        strconv: strconvPackage(),
        "syscall/js": syscallJSPackage(),
        testing: testingPackage(),
        unsafe: unsafePackage(),
        ...context.packages()
    };
}
function cmpPackage() {
    return {
        Compare: hostCallable("cmp.Compare", (args) => BigInt(cmpCompare(args[0] ?? null, args[1] ?? null))),
        Less: hostCallable("cmp.Less", (args) => cmpLess(args[0] ?? null, args[1] ?? null)),
        Or: hostCallable("cmp.Or", (args) => args.find((arg) => !isRuntimeZeroValue(arg)) ?? zeroValueLike(args[0] ?? null))
    };
}
function fmtPackage() {
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
        Errorf: hostCallable("fmt.Errorf", async (args, context) => {
            const format = await toStringValueAsync(args[0] ?? "", context);
            return fmtErrorValue(await sprintfAsync(format, args.slice(1), context));
        }),
        Sprint: hostCallable("fmt.Sprint", (args, context) => {
            return sprintAsync(args, context);
        }),
        Sprintln: hostCallable("fmt.Sprintln", (args, context) => {
            return sprintlnAsync(args, context);
        }),
        Append: hostCallable("fmt.Append", async (args, context) => {
            return appendFormattedBytes(args[0] ?? null, await sprintAsync(args.slice(1), context));
        }),
        Appendf: hostCallable("fmt.Appendf", async (args, context) => {
            const format = await toStringValueAsync(args[1] ?? "", context);
            return appendFormattedBytes(args[0] ?? null, await sprintfAsync(format, args.slice(2), context));
        }),
        Appendln: hostCallable("fmt.Appendln", async (args, context) => {
            return appendFormattedBytes(args[0] ?? null, await sprintlnAsync(args.slice(1), context));
        }),
        Println: hostCallable("fmt.Println", async (args, context) => {
            const text = await sprintlnAsync(args, context);
            context.write(text);
            return BigInt(text.length);
        })
    };
}
function fmtErrorValue(message) {
    return new RuntimeInterfaceValue("error", new RuntimeNamedValue(fmtErrorStringType, RuntimeGoString.fromUtf8Text(message)));
}
function fmtErrorMessage(value) {
    if (value instanceof RuntimeInterfaceValue) {
        return value.value === null ? undefined : fmtErrorMessage(value.value);
    }
    if (!(value instanceof RuntimeNamedValue) || value.typeName !== fmtErrorStringType)
        return undefined;
    const actual = unwrapNamed(value);
    return isRuntimeString(actual) ? goStringText(actual) : formatValue(actual);
}
function mathPackage() {
    return {
        MaxInt32: 2147483647n,
        MaxUint16: 65535n,
        MaxFloat32: 3.4028234663852886e38,
        MaxFloat64: Number.MAX_VALUE,
        NaN: hostCallable("math.NaN", () => Number.NaN),
        Inf: hostCallable("math.Inf", (args) => toFloat(args[0] ?? 1) < 0 ? -Infinity : Infinity),
        Exp: hostCallable("math.Exp", (args) => Math.exp(toFloat(args[0] ?? 0))),
        Floor: hostCallable("math.Floor", (args) => Math.floor(toFloat(args[0] ?? 0))),
        IsNaN: hostCallable("math.IsNaN", (args) => Number.isNaN(toFloat(args[0] ?? 0))),
        Log: hostCallable("math.Log", (args) => Math.log(toFloat(args[0] ?? 0))),
        Float32bits: hostCallable("math.Float32bits", (args) => float32Bits(toFloat(args[0] ?? 0))),
        Float32frombits: hostCallable("math.Float32frombits", (args) => float32FromBits(toBigInt(args[0] ?? 0n))),
        Float64bits: hostCallable("math.Float64bits", (args) => float64Bits(toFloat(args[0] ?? 0))),
        Float64frombits: hostCallable("math.Float64frombits", (args) => float64FromBits(toBigInt(args[0] ?? 0n)))
    };
}
function strconvPackage() {
    return {
        Itoa: hostCallable("strconv.Itoa", (args) => toBigInt(args[0] ?? 0n).toString())
    };
}
function appendFormattedBytes(base, text) {
    const actual = unwrapNamed(base);
    const values = Array.isArray(actual) ? [...actual] : [];
    for (const byte of new TextEncoder().encode(text))
        values.push(BigInt(byte));
    markArrayType(values, "[]byte");
    return values;
}
function runtimePackage() {
    return {
        GOOS: "gojr",
        GOARCH: "js",
        Compiler: "gojr",
        MemProfileRate: 0n,
        AddCleanup: hostCallable("runtime.AddCleanup", () => runtimeCleanupValue()),
        BlockProfile: gojrNotImplementedCallable("runtime.BlockProfile"),
        Caller: hostCallable("runtime.Caller", () => [0n, "", 0n, false]),
        Callers: hostCallable("runtime.Callers", () => 0n),
        CallersFrames: hostCallable("runtime.CallersFrames", () => runtimeFramesValue()),
        FuncForPC: hostCallable("runtime.FuncForPC", () => runtimeFuncValue()),
        GC: hostCallable("runtime.GC", () => null),
        GOMAXPROCS: hostCallable("runtime.GOMAXPROCS", () => 1n),
        GOROOT: hostCallable("runtime.GOROOT", () => ""),
        Goexit: hostCallable("runtime.Goexit", () => null),
        Gosched: hostCallable("runtime.Gosched", () => null),
        KeepAlive: hostCallable("runtime.KeepAlive", () => null),
        GoroutineProfile: gojrNotImplementedCallable("runtime.GoroutineProfile"),
        MemProfile: gojrNotImplementedCallable("runtime.MemProfile"),
        MutexProfile: gojrNotImplementedCallable("runtime.MutexProfile"),
        NumCPU: hostCallable("runtime.NumCPU", () => 1n),
        NumGoroutine: hostCallable("runtime.NumGoroutine", () => 1n),
        ReadMemStats: hostCallable("runtime.ReadMemStats", () => null),
        SetBlockProfileRate: gojrNotImplementedCallable("runtime.SetBlockProfileRate"),
        SetCPUProfileRate: gojrNotImplementedCallable("runtime.SetCPUProfileRate"),
        SetFinalizer: hostCallable("runtime.SetFinalizer", () => null),
        SetMutexProfileFraction: gojrNotImplementedCallable("runtime.SetMutexProfileFraction"),
        Stack: hostCallable("runtime.Stack", (args) => {
            const buffer = unwrapNamed(args[0] ?? null);
            if (!Array.isArray(buffer))
                return 0n;
            const text = "goroutine 1 [running]:\nruntime.Stack(...)\n";
            const bytes = new TextEncoder().encode(text);
            const count = Math.min(buffer.length, bytes.length);
            for (let index = 0; index < count; index += 1) {
                setArrayElement(buffer, index, BigInt(bytes[index] ?? 0));
            }
            return BigInt(count);
        }),
        ThreadCreateProfile: gojrNotImplementedCallable("runtime.ThreadCreateProfile"),
        Version: hostCallable("runtime.Version", () => "gojr")
    };
}
function runtimePprofPackage() {
    return {
        NewProfile: gojrNotImplementedCallable("runtime/pprof.NewProfile"),
        Lookup: gojrNotImplementedCallable("runtime/pprof.Lookup"),
        Profiles: gojrNotImplementedCallable("runtime/pprof.Profiles"),
        WriteHeapProfile: gojrNotImplementedCallable("runtime/pprof.WriteHeapProfile"),
        StartCPUProfile: gojrNotImplementedCallable("runtime/pprof.StartCPUProfile"),
        StopCPUProfile: gojrNotImplementedCallable("runtime/pprof.StopCPUProfile"),
        WithLabels: gojrNotImplementedCallable("runtime/pprof.WithLabels"),
        Labels: gojrNotImplementedCallable("runtime/pprof.Labels"),
        Label: gojrNotImplementedCallable("runtime/pprof.Label"),
        ForLabels: gojrNotImplementedCallable("runtime/pprof.ForLabels"),
        SetGoroutineLabels: gojrNotImplementedCallable("runtime/pprof.SetGoroutineLabels"),
        Do: gojrNotImplementedCallable("runtime/pprof.Do")
    };
}
function runtimeCleanupValue() {
    return {
        Stop: hostCallable("runtime.Cleanup.Stop", () => null)
    };
}
function runtimeFuncValue() {
    return {
        Entry: hostCallable("runtime.(*Func).Entry", () => 0n),
        FileLine: hostCallable("runtime.(*Func).FileLine", () => ["", 0n]),
        Name: hostCallable("runtime.(*Func).Name", () => "")
    };
}
function runtimeFramesValue() {
    let consumed = false;
    return {
        Next: hostCallable("runtime.(*Frames).Next", () => {
            if (consumed)
                return [runtimeFrameValue(), false];
            consumed = true;
            return [runtimeFrameValue(), false];
        })
    };
}
function runtimeFrameValue() {
    return new RuntimeStruct("runtime.Frame", [
        ["PC", 0n],
        ["Func", null],
        ["Function", ""],
        ["File", ""],
        ["Line", 0n],
        ["Entry", 0n]
    ]);
}
function testingPackage() {
    return {
        Short: hostCallable("testing.Short", () => false),
        Verbose: hostCallable("testing.Verbose", () => false)
    };
}
function unsafePackage() {
    return {
        Add: hostCallable("unsafe.Add", (args) => args[0] ?? null),
        Alignof: hostCallable("unsafe.Alignof", (args) => BigInt(runtimeAlignof(args[0] ?? null))),
        Offsetof: hostCallable("unsafe.Offsetof", () => 0n),
        Sizeof: hostCallable("unsafe.Sizeof", (args) => BigInt(runtimeSizeof(args[0] ?? null))),
        Slice: hostCallable("unsafe.Slice", (args) => {
            const length = toNonNegativeLength(args[1] ?? 0n, "unsafe.Slice length");
            if ((args[0] ?? null) === null) {
                if (length === 0)
                    return null;
                throw new GoJuniorRuntimeError("unsafe.Slice: ptr is nil and len is not zero");
            }
            return Array.from({ length }, () => null);
        }),
        SliceData: hostCallable("unsafe.SliceData", (args) => {
            const slice = unwrapNamed(args[0] ?? null);
            if (slice === null)
                return null;
            if (!Array.isArray(slice))
                throw new GoJuniorRuntimeError(`${formatValue(args[0] ?? null)} is not a slice`);
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
const REFLECTLITE_TYPE_INFO = Symbol("gojr reflectlite type");
const REFLECTLITE_VALUE_INFO = Symbol("gojr reflectlite value");
const reflectliteKind = {
    Invalid: 0n,
    Interface: 20n,
    Ptr: 22n
};
function reflectlitePackage() {
    return {
        Invalid: reflectliteKind.Invalid,
        Interface: reflectliteKind.Interface,
        Ptr: reflectliteKind.Ptr,
        TypeOf: hostCallable("internal/reflectlite.TypeOf", (args, context) => reflectliteTypeOf(args[0] ?? null, context)),
        Swapper: hostCallable("internal/reflectlite.Swapper", (args) => reflectliteSwapper(args[0] ?? null)),
        ValueOf: hostCallable("internal/reflectlite.ValueOf", (args, context) => reflectliteValueOf(args[0] ?? null, context))
    };
}
function reflectliteTypeOf(value, context) {
    if (value === null)
        return null;
    return reflectliteType(reflectliteTypeTextOfValue(value), context);
}
function reflectliteValueOf(value, context) {
    return reflectliteValue(value, context);
}
function reflectliteType(typeText, context) {
    const type = normalizeTypeText(typeText);
    const object = {};
    const info = {
        typeText: type,
        kind: reflectliteKindForType(type, context)
    };
    if (type.startsWith("*")) {
        info.elem = reflectliteType(type.slice(1), context);
    }
    object[REFLECTLITE_TYPE_INFO] = info;
    object.Kind = hostCallable("internal/reflectlite.Type.Kind", () => info.kind);
    object.String = hostCallable("internal/reflectlite.Type.String", () => info.typeText);
    object.Elem = hostCallable("internal/reflectlite.Type.Elem", () => {
        if (!info.elem)
            throw new GoJuniorRuntimeError(`reflectlite: Elem of ${type}`);
        return info.elem;
    });
    object.Comparable = hostCallable("internal/reflectlite.Type.Comparable", () => true);
    object.AssignableTo = hostCallable("internal/reflectlite.Type.AssignableTo", (args) => {
        const target = reflectliteTypeInfo(args[0] ?? null);
        return Boolean(target && reflectliteAssignableTo(info, target, context));
    });
    object.Implements = hostCallable("internal/reflectlite.Type.Implements", (args) => {
        const target = reflectliteTypeInfo(args[0] ?? null);
        return Boolean(target && reflectliteAssignableTo(info, target, context));
    });
    return object;
}
function reflectliteValue(value, context, set) {
    const type = reflectliteType(reflectliteTypeTextOfValue(value), context);
    const object = {};
    const info = { value, type, ...(set ? { set } : {}) };
    object[REFLECTLITE_VALUE_INFO] = info;
    object.Type = hostCallable("internal/reflectlite.Value.Type", () => info.type);
    object.Kind = hostCallable("internal/reflectlite.Value.Kind", () => reflectliteTypeInfo(info.type)?.kind ?? reflectliteKind.Invalid);
    object.Len = hostCallable("internal/reflectlite.Value.Len", () => BigInt(valueLength(reflectliteValuePayload(info.value))));
    object.IsNil = hostCallable("internal/reflectlite.Value.IsNil", () => reflectliteIsNil(info.value));
    object.Elem = hostCallable("internal/reflectlite.Value.Elem", () => {
        const actual = info.value instanceof RuntimeInterfaceValue && info.value.value !== null ? info.value.value : info.value;
        if (actual instanceof RuntimePointer) {
            return reflectliteValue(actual.get(), context, (next) => actual.set(next));
        }
        if (actual instanceof RuntimeTypedNilValue && actual.typeName.startsWith("*")) {
            return reflectliteValue(defaultValueForTypeText(actual.typeName.slice(1), context), context);
        }
        throw new GoJuniorRuntimeError(`reflectlite: Elem of ${formatValue(info.value)}`);
    });
    object.Set = hostCallable("internal/reflectlite.Value.Set", (args) => {
        if (!info.set)
            throw new GoJuniorRuntimeError("reflectlite: Set using unaddressable value");
        info.set(reflectliteValuePayload(args[0] ?? null));
        return null;
    });
    return object;
}
function reflectliteSwapper(value) {
    const actual = reflectliteValuePayload(value);
    if (!Array.isArray(actual))
        throw new GoJuniorRuntimeError("reflectlite.Swapper expects a slice or array");
    return {
        kind: "GoJuniorFunction",
        name: "internal/reflectlite.Swapper.func",
        async call(args) {
            const i = toNumber(args[0] ?? 0n);
            const j = toNumber(args[1] ?? 0n);
            const tmp = actual[i] ?? null;
            actual[i] = actual[j] ?? null;
            actual[j] = tmp;
            return null;
        }
    };
}
function reflectliteTypeTextOfValue(value) {
    if (value instanceof RuntimeInterfaceValue && value.value !== null) {
        return reflectliteTypeTextOfValue(value.value);
    }
    return inferredConcreteDynamicType(value) ?? inferredTypeText(value) ?? "interface{}";
}
function reflectliteKindForType(typeText, context) {
    const type = normalizeTypeText(typeText);
    if (type.startsWith("*"))
        return reflectliteKind.Ptr;
    if (type === "error" || type === "any" || type === "interface{}" || context?.interfaceDef(type) || parseAnonymousInterfaceTypeText(type)) {
        return reflectliteKind.Interface;
    }
    return reflectliteKind.Invalid;
}
function reflectliteAssignableTo(source, target, context) {
    if (runtimeTypeAssignableMatch(source.typeText, target.typeText))
        return true;
    if (target.typeText === "any" || target.typeText === "interface{}")
        return true;
    if (target.typeText === "error")
        return true;
    const targetInterface = interfaceTarget(target.typeText, context);
    if (targetInterface)
        return true;
    return false;
}
function reflectliteIsNil(value) {
    if (value === null)
        return true;
    if (value instanceof RuntimeTypedNilValue)
        return true;
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null;
    return false;
}
function reflectliteTypeInfo(value) {
    if (isRuntimeObject(value)) {
        const typeObject = value;
        return typeObject[REFLECTLITE_TYPE_INFO];
    }
    return undefined;
}
function reflectliteValuePayload(value) {
    if (isRuntimeObject(value)) {
        const valueObject = value;
        const info = valueObject[REFLECTLITE_VALUE_INFO];
        if (info)
            return info.value;
    }
    return value;
}
function osPackage() {
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
const SYSCALL_JS_VALUE_INFO = Symbol("gojr syscall/js value");
const syscallJSTypeNames = [
    "undefined",
    "null",
    "boolean",
    "number",
    "string",
    "symbol",
    "object",
    "function"
];
function syscallJSPackage() {
    return {
        TypeUndefined: syscallJSTypeValue(0n),
        TypeNull: syscallJSTypeValue(1n),
        TypeBoolean: syscallJSTypeValue(2n),
        TypeNumber: syscallJSTypeValue(3n),
        TypeString: syscallJSTypeValue(4n),
        TypeSymbol: syscallJSTypeValue(5n),
        TypeObject: syscallJSTypeValue(6n),
        TypeFunction: syscallJSTypeValue(7n),
        CopyBytesToGo: hostCallable("syscall/js.CopyBytesToGo", (args) => syscallJSCopyBytesToGo(args[0] ?? null, args[1] ?? null)),
        CopyBytesToJS: hostCallable("syscall/js.CopyBytesToJS", (args) => syscallJSCopyBytesToJS(args[0] ?? null, args[1] ?? null)),
        FuncOf: hostCallable("syscall/js.FuncOf", (args, context) => syscallJSFuncOf(args[0] ?? null, context)),
        Global: hostCallable("syscall/js.Global", () => syscallJSValue(syscallJSGlobalObject())),
        Null: hostCallable("syscall/js.Null", () => syscallJSValue(null)),
        Undefined: hostCallable("syscall/js.Undefined", () => syscallJSValue(undefined)),
        ValueOf: hostCallable("syscall/js.ValueOf", (args) => syscallJSValue(syscallJSRawFromRuntime(args[0] ?? null)))
    };
}
function syscallJSValue(raw) {
    const object = {};
    object[SYSCALL_JS_VALUE_INFO] = raw;
    object.Bool = hostCallable("syscall/js.Value.Bool", () => {
        if (typeof raw !== "boolean")
            throw new GoJuniorRuntimeError(`syscall/js: call of Value.Bool on ${syscallJSTypeName(raw)}`);
        return raw;
    });
    object.Call = hostCallable("syscall/js.Value.Call", async (args, context) => {
        const property = toStringValue(args[0] ?? "");
        const target = syscallJSObject(raw, "Value.Call");
        const fn = target[property];
        if (typeof fn !== "function")
            throw new GoJuniorRuntimeError(`syscall/js: Value.Call: property ${property} is not a function`);
        return syscallJSValue(await Reflect.apply(fn, target, syscallJSRawArgs(args.slice(1), context)));
    });
    object.Delete = hostCallable("syscall/js.Value.Delete", (args) => {
        const target = syscallJSObject(raw, "Value.Delete");
        delete target[toStringValue(args[0] ?? "")];
        return null;
    });
    object.Equal = hostCallable("syscall/js.Value.Equal", (args) => Object.is(raw, syscallJSRawValue(args[0] ?? null)));
    object.Float = hostCallable("syscall/js.Value.Float", () => {
        if (typeof raw !== "number")
            throw new GoJuniorRuntimeError(`syscall/js: call of Value.Float on ${syscallJSTypeName(raw)}`);
        return raw;
    });
    object.Get = hostCallable("syscall/js.Value.Get", (args) => {
        const target = syscallJSObject(raw, "Value.Get");
        return syscallJSValue(target[toStringValue(args[0] ?? "")]);
    });
    object.Index = hostCallable("syscall/js.Value.Index", (args) => {
        const target = syscallJSObject(raw, "Value.Index");
        return syscallJSValue(target[toNumber(args[0] ?? 0n)]);
    });
    object.InstanceOf = hostCallable("syscall/js.Value.InstanceOf", (args) => {
        const ctor = syscallJSRawValue(args[0] ?? null);
        return typeof ctor === "function" && raw instanceof ctor;
    });
    object.Int = hostCallable("syscall/js.Value.Int", () => {
        if (typeof raw !== "number")
            throw new GoJuniorRuntimeError(`syscall/js: call of Value.Int on ${syscallJSTypeName(raw)}`);
        return BigInt(Math.trunc(raw));
    });
    object.Invoke = hostCallable("syscall/js.Value.Invoke", async (args, context) => {
        if (typeof raw !== "function")
            throw new GoJuniorRuntimeError(`syscall/js: call of Value.Invoke on ${syscallJSTypeName(raw)}`);
        return syscallJSValue(await raw(...syscallJSRawArgs(args, context)));
    });
    object.IsNaN = hostCallable("syscall/js.Value.IsNaN", () => typeof raw === "number" && Number.isNaN(raw));
    object.IsNull = hostCallable("syscall/js.Value.IsNull", () => raw === null);
    object.IsUndefined = hostCallable("syscall/js.Value.IsUndefined", () => raw === undefined);
    object.Length = hostCallable("syscall/js.Value.Length", () => {
        const target = syscallJSObject(raw, "Value.Length");
        return BigInt(Number(target.length ?? 0));
    });
    object.New = hostCallable("syscall/js.Value.New", (args, context) => {
        if (typeof raw !== "function")
            throw new GoJuniorRuntimeError(`syscall/js: call of Value.New on ${syscallJSTypeName(raw)}`);
        return syscallJSValue(Reflect.construct(raw, syscallJSRawArgs(args, context)));
    });
    object.Set = hostCallable("syscall/js.Value.Set", (args, context) => {
        const target = syscallJSObject(raw, "Value.Set");
        target[toStringValue(args[0] ?? "")] = syscallJSRawFromRuntime(args[1] ?? null, context);
        return null;
    });
    object.SetIndex = hostCallable("syscall/js.Value.SetIndex", (args, context) => {
        const target = syscallJSObject(raw, "Value.SetIndex");
        target[toNumber(args[0] ?? 0n)] = syscallJSRawFromRuntime(args[1] ?? null, context);
        return null;
    });
    object.String = hostCallable("syscall/js.Value.String", () => syscallJSString(raw));
    object.Truthy = hostCallable("syscall/js.Value.Truthy", () => Boolean(raw));
    object.Type = hostCallable("syscall/js.Value.Type", () => syscallJSTypeValue(syscallJSType(raw)));
    return new RuntimeNamedValue("js.Value", object);
}
function syscallJSFuncOf(fn, context) {
    const raw = (...args) => {
        if (!context)
            return undefined;
        const goArgs = args.map((arg) => syscallJSValue(arg));
        void callRuntime(fn, [syscallJSValue(undefined), goArgs], context);
        return undefined;
    };
    const value = syscallJSValue(raw);
    const object = {
        Value: value,
        Release: hostCallable("syscall/js.Func.Release", () => null)
    };
    const valueObject = unwrapNamed(value);
    if (isRuntimeObject(valueObject)) {
        for (const [name, method] of Object.entries(valueObject))
            object[name] = method;
    }
    object.Invoke = hostCallable("syscall/js.Func.Invoke", async (args, context) => {
        const jsArgs = args.map((arg) => syscallJSValue(syscallJSRawFromRuntime(arg, context)));
        return syscallJSValue(syscallJSRawFromRuntime(await callRuntime(fn, [syscallJSValue(undefined), jsArgs], context), context));
    });
    return new RuntimeNamedValue("js.Func", object);
}
function syscallJSTypeSelector(value, field) {
    if (!(value instanceof RuntimeNamedValue) || value.typeName !== "js.Type")
        return undefined;
    if (field === "String") {
        return hostCallable("syscall/js.Type.String", () => syscallJSTypeNames[Number(toBigInt(value.value))] ?? "unknown");
    }
    return undefined;
}
function syscallJSTypeValue(kind) {
    return new RuntimeNamedValue("js.Type", kind);
}
function syscallJSType(raw) {
    if (raw === undefined)
        return 0n;
    if (raw === null)
        return 1n;
    switch (typeof raw) {
        case "boolean":
            return 2n;
        case "number":
        case "bigint":
            return 3n;
        case "string":
            return 4n;
        case "symbol":
            return 5n;
        case "function":
            return 7n;
        default:
            return 6n;
    }
}
function syscallJSTypeName(raw) {
    return syscallJSTypeNames[Number(syscallJSType(raw))] ?? "unknown";
}
function syscallJSString(raw) {
    switch (syscallJSType(raw)) {
        case 0n:
            return "<undefined>";
        case 1n:
            return "<null>";
        case 2n:
            return `<boolean: ${String(raw)}>`;
        case 3n:
            return `<number: ${String(raw)}>`;
        case 4n:
            return String(raw);
        case 5n:
            return "<symbol>";
        case 6n:
            return "<object>";
        case 7n:
            return "<function>";
        default:
            return "<unknown>";
    }
}
function syscallJSObject(raw, method) {
    if ((typeof raw !== "object" && typeof raw !== "function") || raw === null) {
        throw new GoJuniorRuntimeError(`syscall/js: call of ${method} on ${syscallJSTypeName(raw)}`);
    }
    return raw;
}
function syscallJSRawArgs(args, context) {
    return args.map((arg) => syscallJSRawFromRuntime(arg, context));
}
function syscallJSRawValue(value) {
    value = unwrapNamed(value);
    if (isRuntimeObject(value)) {
        const object = value;
        if (SYSCALL_JS_VALUE_INFO in object)
            return object[SYSCALL_JS_VALUE_INFO];
        const field = value.Value;
        if (field !== undefined)
            return syscallJSRawValue(field);
    }
    return undefined;
}
function syscallJSRawFromRuntime(value, context) {
    value = unwrapNamed(value);
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null ? null : syscallJSRawFromRuntime(value.value, context);
    if (value instanceof RuntimeTypedNilValue || value === null)
        return null;
    if (isRuntimeObject(value)) {
        const object = value;
        if (SYSCALL_JS_VALUE_INFO in object)
            return object[SYSCALL_JS_VALUE_INFO];
        const field = value.Value;
        if (field !== undefined)
            return syscallJSRawValue(field);
    }
    if (isRuntimeString(value))
        return goStringText(value);
    if (typeof value === "bigint")
        return Number(value);
    if (typeof value === "number" || typeof value === "boolean")
        return value;
    if (Array.isArray(value))
        return value.map((item) => syscallJSRawFromRuntime(item, context));
    if (value instanceof RuntimeMap) {
        const object = {};
        for (const [key, item] of value.orderedEntries()) {
            object[toStringValue(key)] = syscallJSRawFromRuntime(item, context);
        }
        return object;
    }
    if (value instanceof RuntimeStruct) {
        const object = {};
        for (const [name, item] of value.orderedFields())
            object[name] = syscallJSRawFromRuntime(item, context);
        return object;
    }
    return value;
}
let syscallJSGlobalCache;
function syscallJSGlobalObject() {
    if (syscallJSGlobalCache)
        return syscallJSGlobalCache;
    const base = globalThis;
    const globalObject = Object.create(base);
    globalObject.process = base.process ?? syscallJSProcessObject();
    globalObject.path = base.path ?? syscallJSPathObject();
    globalObject.fs = base.fs ?? syscallJSFSObject();
    globalObject.Uint8Array = base.Uint8Array ?? Uint8Array;
    globalObject.Array = base.Array ?? Array;
    globalObject.Object = base.Object ?? Object;
    globalObject.Function = base.Function ?? Function;
    globalObject.Symbol = base.Symbol ?? Symbol;
    globalObject.eval = base.eval ?? ((source) => {
        throw new GoJuniorRuntimeError(`syscall/js eval is not available: ${source}`);
    });
    syscallJSGlobalCache = globalObject;
    return globalObject;
}
function syscallJSProcessObject() {
    const processLike = globalThis.process;
    let cwd = "/";
    return {
        argv: processLike?.argv ?? ["gojr"],
        cwd: () => processLike?.cwd?.() ?? cwd,
        chdir: (path) => {
            if (processLike?.chdir)
                processLike.chdir(path);
            else
                cwd = syscallJSResolvePath(cwd, path);
        }
    };
}
function syscallJSPathObject() {
    return {
        resolve: (...parts) => {
            const processObject = syscallJSGlobalObject().process;
            const cwd = processObject.cwd?.() ?? "/";
            return parts.reduce((current, part) => syscallJSResolvePath(current, String(part)), cwd);
        }
    };
}
function syscallJSResolvePath(cwd, path) {
    const raw = path.startsWith("/") ? path : `${cwd.replace(/\/+$/, "")}/${path}`;
    const stack = [];
    for (const part of raw.split("/")) {
        if (part === "" || part === ".")
            continue;
        if (part === "..")
            stack.pop();
        else
            stack.push(part);
    }
    return `/${stack.join("/")}`;
}
function syscallJSFSObject() {
    const constants = {
        O_WRONLY: 1,
        O_RDWR: 2,
        O_CREAT: 64,
        O_TRUNC: 512,
        O_APPEND: 1024,
        O_EXCL: 128,
        O_DIRECTORY: 65536
    };
    const statObject = (directory = false) => ({
        dev: 0,
        ino: 1,
        mode: directory ? 0o040755 : 0o100644,
        nlink: 1,
        uid: 0,
        gid: 0,
        rdev: 0,
        size: 0,
        blksize: 4096,
        blocks: 0,
        atimeMs: 0,
        mtimeMs: 0,
        ctimeMs: 0,
        isDirectory: () => directory
    });
    const callback = (args) => {
        const last = args[args.length - 1];
        return typeof last === "function" ? last : undefined;
    };
    const ok = (args, value) => {
        callback(args)?.(null, value);
    };
    return {
        constants,
        open: (...args) => ok(args, 3),
        close: (...args) => ok(args),
        mkdir: (...args) => ok(args),
        fstat: (...args) => ok(args, statObject(false)),
        stat: (...args) => ok(args, statObject(false)),
        lstat: (...args) => ok(args, statObject(false)),
        readdir: (...args) => ok(args, []),
        unlink: (...args) => ok(args),
        rmdir: (...args) => ok(args),
        chmod: (...args) => ok(args),
        fchmod: (...args) => ok(args),
        chown: (...args) => ok(args),
        fchown: (...args) => ok(args),
        lchown: (...args) => ok(args),
        utimes: (...args) => ok(args),
        rename: (...args) => ok(args),
        truncate: (...args) => ok(args),
        ftruncate: (...args) => ok(args),
        readlink: (...args) => ok(args, ""),
        link: (...args) => ok(args),
        symlink: (...args) => ok(args),
        fsync: (...args) => ok(args),
        read: (...args) => ok(args, 0),
        write: (...args) => ok(args, syscallJSWriteLength(args))
    };
}
function syscallJSWriteLength(args) {
    const length = args[3];
    return typeof length === "number" ? length : 0;
}
function syscallJSCopyBytesToGo(dst, src) {
    const target = unwrapNamed(dst);
    if (!Array.isArray(target))
        throw new GoJuniorRuntimeError("syscall/js: CopyBytesToGo: expected dst to be a byte slice");
    const source = syscallJSRawValue(src);
    const bytes = syscallJSBytes(source);
    const count = Math.min(target.length, bytes.length);
    for (let index = 0; index < count; index += 1)
        target[index] = BigInt(bytes[index] ?? 0);
    return BigInt(count);
}
function syscallJSCopyBytesToJS(dst, src) {
    const target = syscallJSRawValue(dst);
    const bytes = unwrapNamed(src);
    if (!Array.isArray(bytes))
        throw new GoJuniorRuntimeError("syscall/js: CopyBytesToJS: expected src to be a byte slice");
    if (!syscallJSWritableBytes(target)) {
        throw new GoJuniorRuntimeError("syscall/js: CopyBytesToJS: expected dst to be a Uint8Array or Uint8ClampedArray");
    }
    const count = Math.min(target.length, bytes.length);
    for (let index = 0; index < count; index += 1)
        target[index] = Number(toBigInt(bytes[index] ?? 0n)) & 0xff;
    return BigInt(count);
}
function syscallJSBytes(value) {
    if (value instanceof Uint8Array || value instanceof Uint8ClampedArray)
        return value;
    if (Array.isArray(value) && value.every((item) => typeof item === "number"))
        return value;
    throw new GoJuniorRuntimeError("syscall/js: CopyBytesToGo: expected src to be a Uint8Array or Uint8ClampedArray");
}
function syscallJSWritableBytes(value) {
    return value instanceof Uint8Array ||
        value instanceof Uint8ClampedArray ||
        (Array.isArray(value) && value.every((item) => typeof item === "number"));
}
function hostCallable(name, call) {
    return {
        kind: "HostCallable",
        name,
        call
    };
}
function gojrNotImplementedCallable(name) {
    return hostCallable(name, () => {
        throw new GoJuniorPanic(`gojr error: ${name} not implemented`);
    });
}
function functionValue(declaration) {
    return goJuniorFunctionValue(declaration.name, declaration.signature, declaration.body, undefined, declaration);
}
function functionLiteralValue(expression, context) {
    const closureContext = context.importPath() ? context : undefined;
    return goJuniorFunctionValue("<closure>", resolveRuntimeSignatureTypeImports(expression.signature, context), expression.body, context.captureScope(), undefined, undefined, expression.source, closureContext);
}
function splitTopLevelDeclarations(statements) {
    const declarations = [];
    const executable = [];
    for (const statement of statements) {
        if (statement.kind === "ConstDecl" || statement.kind === "VarDecl" || statement.kind === "TypeDecl") {
            declarations.push(statement);
        }
        else {
            executable.push(statement);
        }
    }
    return { declarations, statements: executable };
}
function predeclareTopLevelTypes(statements, context) {
    for (const statement of statements) {
        if (statement.kind !== "TypeDecl")
            continue;
        for (const declaration of statement.declarations) {
            context.registerType(declaration);
        }
    }
}
function predeclarePackageConstants(pkg, declarations, context) {
    const sourceStringConstants = sourceStringConstantValues(declarations);
    for (const object of packageScopeObjects(pkg)) {
        if (!(object instanceof GoTypesConst))
            continue;
        const typeText = checkedConstantTypeText(object, pkg);
        context.declareRoot(object.Name(), typedCheckedConstantValue(object.Val(), typeText, context, sourceStringConstants.get(object.Name())), false, typeText);
    }
}
function checkedConstantTypeText(object, pkg) {
    const type = object.Type();
    if (!type)
        return undefined;
    const text = GoTypesTypeString(type, GoTypesRelativeTo(pkg));
    return text.startsWith("untyped ") ? undefined : text;
}
function typedCheckedConstantValue(value, typeText, context, sourceValue) {
    const runtimeValue = sourceValue ?? runtimeValueFromCheckedConstant(value);
    if (!typeText)
        return runtimeValue;
    const type = normalizeTypeText(typeText);
    const alias = context.aliasType(type);
    if (alias && alias !== type && !context.typeDef(type) && !context.interfaceDef(type)) {
        return new RuntimeNamedValue(type, convertValueToType(runtimeValue, alias, context));
    }
    return materializeValueForType(runtimeValue, type, context);
}
function sourceStringConstantValues(declarations) {
    const values = new Map();
    for (const statement of declarations) {
        if (statement.kind !== "ConstDecl")
            continue;
        let previousConstValues = [];
        for (const declaration of statement.declarations) {
            const valueExpression = declaration.value ?? previousConstValues[declaration.valueIndex ?? 0];
            const value = sourceStringConstantValue(valueExpression);
            if (value !== undefined)
                values.set(declaration.name, value);
            if (declaration.value)
                previousConstValues[declaration.valueIndex ?? 0] = declaration.value;
        }
    }
    return values;
}
function sourceStringConstantValue(expression) {
    if (!expression)
        return undefined;
    if (expression.kind === "Literal" && expression.literalKind === "string") {
        return goStringFromLiteralRaw(expression.raw);
    }
    if (expression.kind === "BinaryExpression" && expression.operator === "+") {
        const left = sourceStringConstantValue(expression.left);
        const right = sourceStringConstantValue(expression.right);
        return left && right ? left.concat(right) : undefined;
    }
    if (expression.kind === "UnaryExpression" && expression.operator === "+") {
        return sourceStringConstantValue(expression.operand);
    }
    return undefined;
}
function predeclarePackageVariables(declarations, context) {
    for (const statement of declarations) {
        if (statement.kind !== "VarDecl")
            continue;
        for (const declaration of statement.declarations) {
            if (declaration.name === "_")
                continue;
            const typeText = declaration.type?.text;
            context.declareRoot(declaration.name, typeText ? defaultValueForTypeText(typeText, context) : null, true, typeText);
        }
    }
}
function runtimeValueFromCheckedConstant(value) {
    if (value === null ||
        typeof value === "boolean" ||
        typeof value === "string" ||
        typeof value === "number" ||
        typeof value === "bigint") {
        return value;
    }
    return value;
}
async function executePackageVarInitializers(declarations, initOrder, context) {
    const declarationsByName = packageVarDeclarationsByName(declarations);
    const initialized = new Set();
    for (const initializer of initOrder ?? []) {
        const names = initializer.Lhs.map((object) => object.Name()).filter((name) => name !== "_");
        for (const name of names) {
            const declaration = declarationsByName.get(name);
            if (!declaration)
                continue;
            await assignPackageVarDeclaration(declaration, context);
            initialized.add(name);
        }
    }
    for (const [name, declaration] of declarationsByName) {
        if (initialized.has(name) || !declaration.value)
            continue;
        await assignPackageVarDeclaration(declaration, context);
    }
}
function packageVarDeclarationsByName(declarations) {
    const vars = new Map();
    for (const statement of declarations) {
        if (statement.kind !== "VarDecl")
            continue;
        for (const declaration of statement.declarations) {
            if (declaration.name !== "_")
                vars.set(declaration.name, declaration);
        }
    }
    return vars;
}
async function assignPackageVarDeclaration(declaration, context) {
    const value = declaration.value
        ? prepareValueForTargetType(await evaluateExpression(declaration.value, context), declaration.type?.text, declaration.value, context, `variable ${declaration.name}`)
        : defaultValueForDeclarationType(declaration.type, context);
    if (context.hasLocal(declaration.name)) {
        context.assign(declaration.name, value);
        return;
    }
    const typeText = declaration.type ??
        (declaration.value
            ? expressionDeclaredTypeText(declaration.value, context, value) ?? inferredTypeText(value)
            : undefined);
    context.declare(declaration.name, value, true, typeText);
}
async function runInitFunctions(functions, context) {
    for (const declaration of functions) {
        if (declaration.name !== "init" || declaration.receiver)
            continue;
        await callRuntime(functionValue(declaration), [], context);
    }
}
async function executeTopLevelStatements(statements, context) {
    const outcome = await context.deferScopeAsync(() => executeStatements(statements, context));
    return outcome === RECOVERED_PANIC ? { kind: "normal" } : outcome;
}
function goJuniorFunctionValue(name, signature, body, closureScope, declaration, boundReceiver, source, closureContext) {
    const formattedSource = source ?? declaration?.source;
    return {
        kind: "GoJuniorFunction",
        name,
        signature,
        ...(declaration ? { declaration } : {}),
        ...(formattedSource ? { source: formattedSource } : {}),
        async call(args, parentContext) {
            const context = closureContext ?? parentContext;
            const invoke = () => context.functionCallAsync(() => context.childScopeAsync(async () => {
                if (declaration?.receiver?.name) {
                    context.declare(declaration.receiver.name, prepareReceiverBinding(declaration.receiver.type.text, boundReceiver), true, declaration.receiver.type);
                }
                for (const [index, parameter] of signature.parameters.entries()) {
                    if (parameter.name) {
                        const value = parameter.variadic ? args.slice(index) : args[index] ?? null;
                        context.declare(parameter.name, value, true, parameter.variadic ? `[]${parameter.type.text}` : parameter.type);
                        if (parameter.variadic)
                            break;
                    }
                }
                for (const result of signature.results) {
                    if (result.name) {
                        context.declare(result.name, defaultValueForDeclarationType(result.type, context), true, result.type);
                    }
                }
                const outcome = await context.deferScopeAsync(() => executeBlock(body, context, false));
                const completion = outcome === RECOVERED_PANIC
                    ? { kind: "return", values: namedReturnValuesOrZero(signature, context) }
                    : outcome;
                if (completion.kind === "return") {
                    const rawValues = completion.values.length === 0
                        ? namedReturnValues(signature, context)
                        : completion.values;
                    const values = prepareFunctionReturnValues(signature, rawValues, completion.sources, context);
                    const result = values.length === 1 ? values[0] ?? null : values;
                    if (closureContext && closureContext !== parentContext && closureContext.importPath()) {
                        return externalizePackageRuntimeValue(result, closureContext.importPath());
                    }
                    return result;
                }
                return null;
            }));
            return closureScope ? context.withScopeAsync(closureScope, invoke) : invoke();
        }
    };
}
function prepareFunctionReturnValues(signature, values, sources, context) {
    if (signature.results.length === 0)
        return values;
    return values.map((value, index) => {
        const result = signature.results[index];
        if (!result)
            return value;
        return prepareValueForTargetType(value, result.type.text, sources?.[index], context, "return value");
    });
}
function namedReturnValues(signature, context) {
    const namedResults = signature.results.filter((result) => result.name);
    if (namedResults.length === 0)
        return [];
    return namedResults.map((result) => result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type, context));
}
function namedReturnValuesOrZero(signature, context) {
    if (signature.results.length === 0)
        return [];
    return signature.results.map((result) => result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type, context));
}
function normalizeReceiverType(typeText) {
    const type = normalizeTypeText(typeText);
    if (type.startsWith("*"))
        return { baseType: type.slice(1), pointer: true };
    return { baseType: type, pointer: false };
}
function methodKey(typeName, methodName) {
    return `${typeName}.${methodName}`;
}
function boundMethodValue(method, receiver) {
    const boundReceiver = method.pointerReceiver
        ? pointerToReceiver(receiver, method.receiverType)
        : valueReceiver(receiver);
    return goJuniorFunctionValue(method.declaration.name, method.declaration.signature, method.declaration.body, method.closureScope, method.declaration, boundReceiver);
}
function methodExpressionValue(method, receiverTypeText) {
    const receiverType = { text: receiverTypeText };
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
function dynamicMethodExpressionValue(receiverTypeText, methodName, methodSignature) {
    const receiverType = { text: receiverTypeText };
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
            if (!method)
                throw new GoJuniorRuntimeError(`${formatValue(receiver)} has no method ${methodName}`);
            return callRuntime(boundMethodValue(method.method, method.receiver), args.slice(1), context);
        }
    };
}
function valueReceiver(receiver) {
    const value = dereferenceIfPointer(receiver);
    return value instanceof RuntimeStruct ? value.clone() : value;
}
function prepareReceiverBinding(receiverTypeText, receiver) {
    if (receiver === undefined)
        return null;
    const receiverType = normalizeReceiverType(receiverTypeText);
    if (receiverType.pointer)
        return pointerToReceiver(receiver, receiverType.baseType);
    return valueReceiver(receiver);
}
function pointerToReceiver(receiver, typeName) {
    if (receiver instanceof RuntimePointer)
        return receiver;
    if (receiver instanceof RuntimeTypedNilValue && runtimeTypeAssignableMatch(receiver.typeName, `*${typeName}`))
        return receiver;
    if (receiver instanceof RuntimeStruct && receiver.typeName === typeName) {
        return new RuntimePointer(typeName, () => receiver, (value) => {
            if (!(value instanceof RuntimeStruct) || value.typeName !== typeName) {
                throw new GoJuniorRuntimeError(`${formatValue(value)} is not assignable to *${typeName}`);
            }
            receiver.fields.clear();
            for (const [field, item] of value.orderedFields())
                receiver.set(field, item);
        });
    }
    throw new GoJuniorRuntimeError(`${formatValue(receiver)} is not addressable as *${typeName}`);
}
async function executeStatements(statements, context) {
    const labels = statementLabels(statements);
    let lastValue;
    for (let pc = 0; pc < statements.length; pc += 1) {
        const statement = statements[pc];
        if (!statement)
            continue;
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
        if (completion.kind !== "normal")
            return completion;
        if (completion.value !== undefined)
            lastValue = completion.value;
    }
    return lastValue === undefined ? { kind: "normal" } : { kind: "normal", value: lastValue };
}
function validateGotoTarget(statements, sourceIndex, targetIndex, label) {
    if (targetIndex <= sourceIndex)
        return;
    for (let index = sourceIndex + 1; index < targetIndex; index += 1) {
        if (statementDeclaresVariables(statements[index])) {
            throw new GoJuniorRuntimeError(`goto ${label} jumps over variable declaration`);
        }
    }
}
function statementDeclaresVariables(statement) {
    if (!statement)
        return false;
    if (statement.kind === "VarDecl" || statement.kind === "ShortVarStatement")
        return true;
    if (statement.kind === "LabeledStatement")
        return statementDeclaresVariables(statement.statement);
    return false;
}
function statementLabels(statements) {
    const labels = new Map();
    for (const [index, statement] of statements.entries()) {
        if (statement.kind === "LabeledStatement") {
            labels.set(statement.label, index);
        }
    }
    return labels;
}
async function executeBlock(block, context, createScope = true) {
    if (!createScope)
        return executeStatements(block.statements, context);
    return context.childScopeAsync(() => executeStatements(block.statements, context));
}
async function executeStatement(statement, context) {
    switch (statement.kind) {
        case "BlockStatement":
            return executeBlock(statement, context);
        case "LabeledStatement":
            return executeLabeledStatement(statement, context);
        case "ConstDecl":
            let previousConstValues = [];
            let previousConstType = undefined;
            for (const [index, declaration] of statement.declarations.entries()) {
                const valueExpression = declaration.value ?? previousConstValues[declaration.valueIndex ?? 0];
                const type = declaration.type ?? previousConstType;
                const value = valueExpression
                    ? await evaluateConstExpression(valueExpression, BigInt(declaration.iotaIndex ?? index), context)
                    : defaultValueForDeclarationType(type, context);
                const typedValue = type ? convertValueToType(value, type.text, context, expressionDeclaredTypeText(valueExpression, context, value)) : value;
                context.declare(declaration.name, typedValue, false, type);
                if (declaration.value)
                    previousConstValues[declaration.valueIndex ?? 0] = declaration.value;
                if (declaration.type)
                    previousConstType = declaration.type;
            }
            return { kind: "normal" };
        case "VarDecl":
            for (const declaration of statement.declarations) {
                const value = declaration.value
                    ? prepareValueForTargetType(await evaluateExpression(declaration.value, context), declaration.type?.text, declaration.value, context, `variable ${declaration.name}`)
                    : defaultValueForDeclarationType(declaration.type, context);
                const typeText = declaration.type ??
                    (declaration.value
                        ? expressionDeclaredTypeText(declaration.value, context, value) ?? inferredTypeText(value)
                        : undefined);
                context.declare(declaration.name, value, true, typeText);
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
                values: await evaluateExpressionList(statement.values, context),
                sources: statement.values
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
async function executeLabeledStatement(statement, context) {
    if (!statement.statement)
        return { kind: "normal" };
    if (statement.statement.kind === "ForStatement") {
        return executeFor(statement.statement, context, statement.label);
    }
    if (statement.statement.kind === "SwitchStatement") {
        return executeSwitch(statement.statement, context, statement.label);
    }
    return executeStatement(statement.statement, context);
}
async function evaluateConstExpression(expression, iotaValue, context) {
    const iotaShadowed = context.hasLocal("iota");
    return context.childScopeAsync(async () => {
        if (!iotaShadowed)
            context.declare("iota", iotaValue, false, "int64");
        return evaluateExpression(expression, context);
    });
}
async function executeIf(statement, context) {
    return context.childScopeAsync(async () => {
        if (statement.init) {
            expectNormalCompletion(await executeStatement(statement.init, context), "if init statement");
        }
        if (toBool(await evaluateExpression(statement.condition, context))) {
            return executeBlock(statement.thenBlock, context);
        }
        if (!statement.elseBranch)
            return { kind: "normal" };
        return statement.elseBranch.kind === "IfStatement"
            ? executeIf(statement.elseBranch, context)
            : executeBlock(statement.elseBranch, context);
    });
}
async function executeSwitch(statement, context, label) {
    return context.childScopeAsync(async () => {
        if (statement.init) {
            expectNormalCompletion(await executeStatement(statement.init, context), "switch init statement");
        }
        if (statement.typeSwitch)
            return executeTypeSwitch(statement, context, label);
        validateValueSwitchFallthrough(statement);
        return executeValueSwitch(statement, context, label);
    });
}
async function executeValueSwitch(statement, context, label) {
    const switchValue = statement.expression ? await evaluateExpression(statement.expression, context) : true;
    const start = await selectValueSwitchClause(statement, switchValue, context);
    if (start === undefined)
        return { kind: "normal" };
    for (let index = start; index < statement.clauses.length; index += 1) {
        const clause = statement.clauses[index];
        if (!clause)
            continue;
        const completion = await executeStatements(clause.statements, context);
        if (completion.kind === "fallthrough") {
            continue;
        }
        if (completion.kind === "break" && labelMatches(completion.label, label))
            return { kind: "normal" };
        return completion;
    }
    return { kind: "normal" };
}
async function selectValueSwitchClause(statement, switchValue, context) {
    let defaultIndex;
    for (const [index, clause] of statement.clauses.entries()) {
        if (clause.default) {
            defaultIndex ??= index;
            continue;
        }
        for (const value of clause.values) {
            if (valueEqual(switchValue, await evaluateExpression(value, context)))
                return index;
        }
    }
    return defaultIndex;
}
function validateValueSwitchFallthrough(statement) {
    for (const [index, clause] of statement.clauses.entries()) {
        const fallthroughIndex = clause.statements.findIndex((item) => item.kind === "BranchStatement" && item.branch === "fallthrough");
        if (fallthroughIndex < 0)
            continue;
        if (index === statement.clauses.length - 1) {
            throw new GoJuniorRuntimeError("fallthrough cannot appear in the final switch clause");
        }
        if (fallthroughIndex !== clause.statements.length - 1) {
            throw new GoJuniorRuntimeError("fallthrough must be the final statement in a switch clause");
        }
    }
}
async function executeTypeSwitch(statement, context, label) {
    if (!statement.typeSwitch)
        return { kind: "normal" };
    const switchValue = await evaluateExpression(statement.typeSwitch.expression, context);
    const start = selectTypeSwitchClause(statement, switchValue, context);
    if (start === undefined)
        return { kind: "normal" };
    for (let index = start; index < statement.clauses.length; index += 1) {
        const clause = statement.clauses[index];
        if (!clause)
            continue;
        const completion = await context.childScopeAsync(async () => {
            if (statement.typeSwitch?.name) {
                const bindingValue = typeSwitchBindingValue(switchValue, clause, context);
                if (statement.typeSwitch.define) {
                    context.declare(statement.typeSwitch.name, bindingValue, true);
                }
                else {
                    context.assign(statement.typeSwitch.name, bindingValue);
                }
            }
            return executeStatements(clause.statements, context);
        });
        if (completion.kind === "fallthrough") {
            throw new GoJuniorRuntimeError("fallthrough is not allowed in type switches");
        }
        if (completion.kind === "break" && labelMatches(completion.label, label))
            return { kind: "normal" };
        return completion;
    }
    return { kind: "normal" };
}
function selectTypeSwitchClause(statement, switchValue, context) {
    let defaultIndex;
    for (const [index, clause] of statement.clauses.entries()) {
        if (clause.default) {
            defaultIndex ??= index;
            continue;
        }
        if ((clause.typeValues ?? []).some((type) => runtimeValueMatchesType(switchValue, type.text, context)))
            return index;
    }
    return defaultIndex;
}
function typeSwitchBindingValue(switchValue, clause, context) {
    if (clause.default || (clause.typeValues?.length ?? 0) !== 1)
        return switchValue;
    const typeText = clause.typeValues?.[0]?.text;
    if (!typeText)
        return switchValue;
    if (normalizeTypeText(typeText) === "nil")
        return null;
    return assertedRuntimeValue(switchValue, typeText, context);
}
async function executeFor(statement, context, label) {
    if (statement.range) {
        const source = await evaluateExpression(statement.range.source, context);
        const keyTypeText = integerRangeKeyTypeText(statement.range.source, source, context);
        const entries = await rangeEntries(source, context);
        for (const [index, value] of entries) {
            const completion = await context.childScopeAsync(async () => {
                if (statement.range?.keyName) {
                    if (statement.range.keyName !== "_") {
                        if (statement.range.define)
                            context.declare(statement.range.keyName, index, true, keyTypeText ?? inferredTypeText(index));
                        else
                            context.assign(statement.range.keyName, index);
                    }
                }
                else if (statement.range?.keyTarget) {
                    await assignExpressionTarget(statement.range.keyTarget, index, context);
                }
                if (statement.range?.valueName) {
                    if (statement.range.valueName !== "_") {
                        if (statement.range.define)
                            context.declare(statement.range.valueName, value, true, inferredTypeText(value));
                        else
                            context.assign(statement.range.valueName, value);
                    }
                }
                else if (statement.range?.valueTarget) {
                    await assignExpressionTarget(statement.range.valueTarget, value, context);
                }
                return executeBlock(statement.body, context, false);
            });
            if (completion.kind === "break" && labelMatches(completion.label, label))
                return { kind: "normal" };
            if (completion.kind === "continue" && labelMatches(completion.label, label))
                continue;
            if (completion.kind !== "normal")
                return completion;
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
            if (completion.kind === "break" && labelMatches(completion.label, label))
                return { kind: "normal" };
            if (completion.kind === "continue" && labelMatches(completion.label, label)) {
                if (statement.post) {
                    expectNormalCompletion(await executeStatement(statement.post, context), "for post statement");
                }
                continue;
            }
            if (completion.kind !== "normal")
                return completion;
            if (statement.post) {
                expectNormalCompletion(await executeStatement(statement.post, context), "for post statement");
            }
        }
        throw new GoJuniorRuntimeError(`loop exceeded ${context.loopLimit()} iterations`);
    });
}
function expectNormalCompletion(completion, label) {
    if (completion.kind !== "normal") {
        throw new GoJuniorRuntimeError(`${completionDescription(completion)} used in ${label}`);
    }
}
function labelMatches(completionLabel, activeLabel) {
    return completionLabel === undefined || completionLabel === activeLabel;
}
function completionErrorMessage(completion) {
    if (completion.kind === "goto")
        return `unresolved goto label ${completion.label}`;
    return `${completionDescription(completion)} used outside a matching loop or switch`;
}
function completionDescription(completion) {
    if (completion.kind === "return")
        return "return";
    if (completion.kind === "fallthrough")
        return "fallthrough";
    return completion.label ? `${completion.kind} ${completion.label}` : completion.kind;
}
async function executeDefer(statement, context) {
    if (statement.expression.kind === "CallExpression") {
        const callee = await evaluateExpression(statement.expression.callee, context);
        const args = await evaluateCallArguments(callee, statement.expression.args, statement.expression.spreadLast, context);
        context.pushDefer(() => callRuntime(callee, args, context));
        return;
    }
    context.pushDefer(() => evaluateExpression(statement.expression, context));
}
async function executeGo(statement, context) {
    const callee = await evaluateExpression(statement.call.callee, context);
    const args = await evaluateCallArguments(callee, statement.call.args, statement.call.spreadLast, context);
    const goroutineContext = context.fork();
    context.scheduler().go(async () => {
        await callRuntime(callee, args, goroutineContext);
    });
}
async function executeSend(statement, context) {
    const channel = await evaluateExpression(statement.channel, context);
    if (channel === null)
        throw new GoJuniorDeadlockError("send on nil channel would block");
    if (!(channel instanceof RuntimeChannel))
        throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
    await channel.sendAsync(await evaluateExpression(statement.value, context));
}
async function executeSelect(statement, context) {
    const prepared = await Promise.all(statement.clauses.map((clause) => prepareSelectClause(clause, context)));
    let selected;
    try {
        selected = await asyncSelect(context.scheduler(), prepared.map((item) => item.selectCase));
    }
    catch (error) {
        if (isDeadlockError(error))
            throw new GoJuniorDeadlockError("select would block");
        throw normalizeAsyncRuntimeError(error);
    }
    const preparedClause = prepared[selected.index];
    if (!preparedClause)
        return { kind: "normal" };
    const completion = await context.childScopeAsync(async () => {
        await applySelectResult(preparedClause.clause.comm, selected, context);
        return executeStatements(preparedClause.clause.statements, context);
    });
    if (completion.kind === "break" && completion.label === undefined)
        return { kind: "normal" };
    return completion;
}
async function prepareSelectClause(clause, context) {
    const comm = clause.comm;
    if (clause.default || !comm)
        return { clause, selectCase: { op: "default" } };
    if (comm.kind === "SendStatement") {
        const channel = await evaluateExpression(comm.channel, context);
        const value = await evaluateExpression(comm.value, context);
        if (channel === null)
            return { clause, selectCase: { op: "receive", channel: undefined } };
        if (!(channel instanceof RuntimeChannel))
            throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
        return { clause, selectCase: { op: "send", channel: channel.asyncChannel(), value } };
    }
    const receive = selectReceiveExpression(comm);
    if (receive) {
        const channel = await evaluateExpression(receive, context);
        if (channel === null)
            return { clause, selectCase: { op: "receive", channel: undefined } };
        if (!(channel instanceof RuntimeChannel))
            throw new GoJuniorRuntimeError(`${formatValue(channel)} is not a channel`);
        return { clause, selectCase: { op: "receive", channel: channel.asyncChannel() } };
    }
    await executeStatement(comm, context).then((completion) => {
        expectNormalCompletion(completion, "select communication clause");
    });
    return { clause, selectCase: { op: "default" } };
}
async function applySelectResult(statement, result, context) {
    if (result.op !== "receive" || !statement)
        return;
    await assignSelectReceive(statement, [result.value, result.ok], context);
}
function selectReceiveExpression(statement) {
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
async function assignSelectReceive(statement, received, context) {
    if (statement.kind === "ExpressionStatement")
        return;
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
function selectReceiveValues(received, targetCount) {
    if (targetCount === 2)
        return [received[0], received[1]];
    return [received[0]];
}
function branchCompletion(statement) {
    if (statement.branch === "break") {
        return statement.label ? { kind: "break", label: statement.label } : { kind: "break" };
    }
    if (statement.branch === "continue") {
        return statement.label ? { kind: "continue", label: statement.label } : { kind: "continue" };
    }
    if (statement.branch === "goto")
        return { kind: "goto", label: statement.label ?? "<missing>" };
    return { kind: "fallthrough" };
}
async function executeAssign(statement, context) {
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
        }
        else {
            await assignExpressionTarget(target, value, context);
        }
    }
}
async function executeShortVar(statement, context) {
    const values = await evaluateAssignmentValues(statement.values, statement.names.length, context);
    declareOrAssignShortVars(statement.names, values, context, statement.values);
}
function declareOrAssignShortVars(names, values, context, sources = []) {
    if (values.length !== names.length) {
        throw new GoJuniorRuntimeError(`short declaration count mismatch: ${names.length} names but ${values.length} values`);
    }
    const hasNewName = names.some((name) => name !== "_" && name !== "<invalid>" && !context.hasLocal(name));
    if (!hasNewName) {
        throw new GoJuniorRuntimeError("short declaration has no new variables");
    }
    for (const [index, name] of names.entries()) {
        if (name === "_")
            continue;
        if (name === "<invalid>")
            throw new GoJuniorRuntimeError("non-identifier used in short declaration");
        const value = values[index] ?? null;
        if (context.hasLocal(name)) {
            context.assign(name, prepareValueForTargetType(value, context.lookupTypeText(name), sources[index], context, `variable ${name}`));
        }
        else {
            const typeText = expressionDeclaredTypeText(sources[index], context, value) ?? inferredTypeText(value);
            context.declare(name, value, true, typeText);
        }
    }
}
async function evaluateAssignmentValues(expressions, targetCount, context) {
    if (targetCount === 2 && expressions.length === 1 && expressions[0]?.kind === "IndexExpression") {
        const lookup = await evaluateMapLookupWithPresence(expressions[0], context);
        if (lookup)
            return lookup;
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
function applyCompoundAssignment(operator, left, right) {
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
async function evaluateMapLookupWithPresence(expression, context) {
    const object = unwrapNamed(await evaluateExpression(expression.object, context));
    if (!(object instanceof RuntimeMap))
        return undefined;
    const index = await evaluateExpression(expression.index, context);
    const [value, ok] = object.getWithPresence(index);
    return [value, ok];
}
async function executeIncDec(statement, context) {
    const current = await evaluateExpression(statement.target, context);
    const next = statement.operator === "++"
        ? addNumbers(current, 1n)
        : subtractNumbers(current, 1n);
    await assignExpressionTarget(statement.target, next, context);
}
async function assignExpressionTarget(target, value, context) {
    if (target.kind === "Identifier") {
        if (target.name === "_")
            return;
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
async function evaluateExpression(expression, context) {
    switch (expression.kind) {
        case "Identifier":
            return evaluateIdentifier(expression, context);
        case "TypeExpression":
            throw new GoJuniorRuntimeError(`${expression.type.text} is a type, not a value`);
        case "Literal":
            if (expression.literalKind === "string")
                return goStringFromLiteralRaw(expression.raw);
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
            return getSlice(await evaluateExpression(expression.object, context), expression.start ? await evaluateExpression(expression.start, context) : undefined, expression.end ? await evaluateExpression(expression.end, context) : undefined, expression.max ? await evaluateExpression(expression.max, context) : undefined);
        case "SpreadsheetRangeExpression":
            return getSpreadsheetRange(expression, context);
    }
}
async function evaluateExpressionList(expressions, context) {
    const values = [];
    for (const expression of expressions)
        values.push(await evaluateExpression(expression, context));
    return values;
}
async function evaluateCallArguments(callee, expressions, spreadLast, context) {
    const raw = await evaluateExpressionList(expressions, context);
    const args = spreadLast ? spreadLastArgument(raw) : raw;
    if (!isGoJuniorFunction(callee) || !callee.signature)
        return args;
    const prepared = [...args];
    const calleeName = runtimeCallableName(callee);
    for (const [index, parameter] of callee.signature.parameters.entries()) {
        if (parameter.variadic) {
            for (let argIndex = index; argIndex < prepared.length; argIndex += 1) {
                const expression = spreadLast && argIndex >= expressions.length - 1
                    ? expressions[expressions.length - 1]
                    : expressions[argIndex];
                prepared[argIndex] = prepareValueForParameterType(prepared[argIndex] ?? null, parameter.type.text, expression, context, `argument ${argIndex + 1} to ${calleeName}`);
            }
            break;
        }
        if (index >= prepared.length)
            break;
        prepared[index] = prepareValueForParameterType(prepared[index] ?? null, parameter.type.text, expressions[index], context, `argument ${index + 1} to ${calleeName}`);
    }
    return prepared;
}
function runtimeCallableName(callee) {
    if (isGoJuniorFunction(callee) || isRuntimeCallable(callee))
        return callee.name;
    return "call";
}
async function evaluateArrayLiteral(expression, context) {
    const typeText = context.resolveImportedTypeText(expression.type.text);
    const type = parseArrayOrSliceTypeText(typeText, context);
    if (!type) {
        throw new GoJuniorRuntimeError(`${typeText} is not an array or slice literal type`);
    }
    const values = await evaluateArrayLiteralEntries(expression.elements, type.elementType, type.length, Boolean(type.inferLength), typeText, context);
    markArrayType(values, type.typeText);
    return values;
}
async function evaluateArrayLiteralEntries(elements, elementType, length, inferLength, displayType, context) {
    const values = [];
    let nextIndex = 0;
    for (const element of elements) {
        const index = element.key ? toNumber(await evaluateExpression(element.key, context)) : nextIndex;
        if (!Number.isInteger(index) || index < 0)
            throw new GoJuniorRuntimeError(`${displayType} array literal has invalid index ${index}`);
        if (length !== undefined && !inferLength && index >= length) {
            throw new GoJuniorRuntimeError(`${displayType} array literal index ${index} out of bounds`);
        }
        values[index] = prepareAssignableToType(await evaluateExpression(element.value, context), elementType, "array element", context);
        nextIndex = index + 1;
    }
    if (length !== undefined && values.length > length) {
        throw new GoJuniorRuntimeError(`array literal has ${values.length} elements but type ${displayType} has length ${length}`);
    }
    const targetLength = length !== undefined && !inferLength ? length : values.length;
    for (let index = 0; index < targetLength; index += 1) {
        if (values[index] === undefined)
            values[index] = defaultValueForTypeText(elementType, context);
    }
    return values;
}
function makeRuntimeSlice(elementType, length, capacity, context, typeText = `[]${elementType}`) {
    if (capacity < length)
        throw new GoJuniorRuntimeError("slice capacity is smaller than length");
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
async function evaluateStructLiteral(expression, context) {
    const typeName = context.resolveImportedTypeText(expression.typeName);
    const pointerTarget = normalizeTypeText(typeName).startsWith("*")
        ? normalizeTypeText(typeName).slice(1)
        : undefined;
    if (pointerTarget) {
        let value = await evaluateStructLiteral({ ...expression, typeName: pointerTarget }, context);
        return new RuntimePointer(pointerTarget, () => value, (next) => {
            value = prepareAssignableToType(next, pointerTarget, `*${pointerTarget} literal`, context);
        });
    }
    const resolvedExpression = { ...expression, typeName };
    const intrinsic = await evaluateIntrinsicNamedStructLiteral(resolvedExpression, context);
    if (intrinsic)
        return intrinsic;
    const typeDef = context.typeDef(typeName) ?? parseAnonymousStructTypeText(typeName);
    if (!typeDef) {
        const alias = context.aliasType(typeName);
        const arrayType = alias ? parseArrayOrSliceTypeText(alias, context) : undefined;
        if (arrayType) {
            const values = await evaluateArrayLiteralFields(resolvedExpression, arrayType, context);
            return new RuntimeNamedValue(typeName, values);
        }
        const mapType = alias ? parseMapTypeText(alias) : undefined;
        if (mapType) {
            return new RuntimeNamedValue(typeName, await evaluateNamedMapLiteralFields(resolvedExpression, mapType, context));
        }
        throw new GoJuniorRuntimeError(`${typeName} is not a declared struct type`);
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
                ? `${typeName} has no field ${field.name}`
                : `${typeName} literal has too many values`);
        }
        const value = await evaluateExpression(field.value, context);
        struct.set(declared.name, prepareAssignableToType(value, declared.type.text, `field ${field.name ?? declared.name}`, context));
    }
    return struct;
}
async function evaluateIntrinsicNamedStructLiteral(expression, context) {
    const type = normalizeTypeText(expression.typeName);
    if (type !== "sync.Pool")
        return undefined;
    let newFn;
    for (const field of expression.fields) {
        if (field.name === "New") {
            newFn = await evaluateExpression(field.value, context);
        }
    }
    return new RuntimeNamedValue(type, syncPoolValue(newFn));
}
async function evaluateArrayLiteralFields(expression, type, context) {
    const values = await evaluateArrayLiteralEntries(expression.fields.map((field) => ({ ...(field.key ? { key: field.key } : {}), value: field.value })), type.elementType, type.length, Boolean(type.inferLength), expression.typeName, context);
    markArrayType(values, expression.typeName);
    return values;
}
async function evaluateNamedMapLiteralFields(expression, type, context) {
    const map = new RuntimeMap(type.keyType, type.valueType, context);
    for (const field of expression.fields) {
        if (!field.key)
            throw new GoJuniorRuntimeError(`${expression.typeName} map literal requires keyed elements`);
        const key = prepareAssignableToType(await evaluateExpression(field.key, context), type.keyType, "map key", context);
        const value = prepareAssignableToType(await evaluateExpression(field.value, context), type.valueType, "map value", context);
        map.set(key, value);
    }
    return map;
}
async function evaluateMapLiteral(expression, context) {
    const map = new RuntimeMap(context.resolveImportedTypeText(expression.keyType.text), context.resolveImportedTypeText(expression.valueType.text), context);
    for (const entry of expression.entries) {
        map.set(await evaluateExpression(entry.key, context), await evaluateExpression(entry.value, context));
    }
    return map;
}
function evaluateIdentifier(expression, context) {
    if (expression.name === "nil")
        return null;
    return context.lookup(expression.name);
}
async function evaluateUnary(expression, context) {
    let value;
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
async function receiveFromChannel(value) {
    if (value === null)
        throw new GoJuniorDeadlockError("receive from nil channel would block");
    if (!(value instanceof RuntimeChannel))
        throw new GoJuniorRuntimeError(`${formatValue(value)} is not a channel`);
    return value.receiveAsync();
}
async function evaluateBinary(expression, context) {
    if (expression.operator === "&&") {
        return toBool(await evaluateExpression(expression.left, context)) && toBool(await evaluateExpression(expression.right, context));
    }
    if (expression.operator === "||") {
        return toBool(await evaluateExpression(expression.left, context)) || toBool(await evaluateExpression(expression.right, context));
    }
    const left = await evaluateExpression(expression.left, context);
    const right = await evaluateExpression(expression.right, context);
    try {
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
    catch (error) {
        if (error instanceof GoJuniorRuntimeError && error.span === undefined && expression.span) {
            throw new GoJuniorRuntimeError(error.message, expression.span);
        }
        throw error;
    }
}
async function evaluateTypeAssertion(expression, context) {
    const value = await evaluateExpression(expression.expression, context);
    if (!runtimeValueMatchesType(value, expression.type.text, context)) {
        throw new GoJuniorRuntimeError(`${formatValue(value)} does not have dynamic type ${normalizeTypeText(expression.type.text)}`);
    }
    return assertedRuntimeValue(value, expression.type.text, context);
}
async function evaluateTypeAssertionWithPresence(expression, context) {
    const value = await evaluateExpression(expression.expression, context);
    const typeText = normalizeTypeText(expression.type.text);
    if (runtimeValueMatchesType(value, typeText, context))
        return [assertedRuntimeValue(value, typeText, context), true];
    return [defaultValueForTypeText(typeText, context), false];
}
function assertedRuntimeValue(value, typeText, context) {
    const type = normalizeTypeText(typeText);
    const dynamicValue = value instanceof RuntimeInterfaceValue ? value.value : value;
    const targetInterface = interfaceTarget(type, context);
    if (targetInterface)
        return prepareInterfaceAssignment(dynamicValue, type, targetInterface, "type assertion", context);
    return dynamicValue;
}
async function evaluateCall(expression, context) {
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
        const argument = expression.args[0];
        const value = await evaluateExpression(argument, context);
        return convertValueToType(value, conversionType, context, expressionDeclaredTypeText(argument, context, value));
    }
    const callee = await evaluateExpression(expression.callee, context);
    const args = await evaluateCallArguments(callee, expression.args, expression.spreadLast, context);
    return callRuntime(callee, args, context);
}
function conversionTargetType(callee, context) {
    const typeText = conversionTargetTypeText(callee, context);
    return typeText && runtimeKnowsTypeText(typeText, context) ? typeText : undefined;
}
function conversionTargetTypeText(callee, context) {
    if (callee.kind === "TypeExpression")
        return callee.type.text;
    if (callee.kind === "Identifier") {
        if (context.hasBinding(callee.name))
            return undefined;
        return context.isKnownType(callee.name) ? callee.name : undefined;
    }
    if (callee.kind === "UnaryExpression" && callee.operator === "*") {
        const operand = conversionTargetTypeText(callee.operand, context);
        return operand ? `*${operand}` : undefined;
    }
    const typeText = typeArgumentText(callee);
    return typeText && context.isKnownType(typeText) ? typeText : undefined;
}
async function evaluateNew(expression, context) {
    if (expression.spreadLast || expression.args.length !== 1)
        throw new GoJuniorRuntimeError("new expects exactly one argument");
    const argument = expression.args[0];
    const typeArgument = newTypeArgumentText(argument, context);
    const evaluated = typeArgument ? undefined : await evaluateExpression(argument, context);
    const typeText = typeArgument ?? expressionDeclaredTypeText(argument, context, evaluated);
    if (!typeText)
        throw new GoJuniorRuntimeError("new could not infer argument type");
    let value = typeArgument
        ? defaultValueForTypeText(typeText, context)
        : copyValueForNew(prepareAssignableToType(evaluated, typeText, "new argument", context), typeText, context);
    return new RuntimePointer(normalizeTypeText(typeText), () => value, (next) => {
        assertAssignableToType(next, typeText, `*${typeText}`, context);
        value = next;
    });
}
function newTypeArgumentText(expression, context) {
    if (expression.kind === "TypeExpression")
        return expression.type.text;
    if (expression.kind === "Identifier") {
        if (context.hasBinding(expression.name))
            return undefined;
        return context.isKnownType(expression.name) ? expression.name : undefined;
    }
    const typeText = typeArgumentText(expression);
    return typeText && context.isKnownType(typeText) ? typeText : undefined;
}
function copyValueForNew(value, typeText, context) {
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
async function evaluateMake(expression, context) {
    if (expression.spreadLast)
        throw new GoJuniorRuntimeError("make does not accept spread arguments");
    const typeArg = expression.args[0];
    if (!typeArg || typeArg.kind !== "TypeExpression") {
        throw new GoJuniorRuntimeError("make expects a map or slice type as its first argument");
    }
    const typeText = normalizeTypeText(typeArg.type.text);
    const mapType = parseMapTypeText(typeText);
    if (mapType) {
        if (expression.args.length > 2)
            throw new GoJuniorRuntimeError("make map accepts at most one size hint");
        if (expression.args[1])
            toNonNegativeLength(await evaluateExpression(expression.args[1], context), "map size hint");
        return new RuntimeMap(mapType.keyType, mapType.valueType, context);
    }
    const chanType = parseChanTypeText(typeText);
    if (chanType) {
        if (chanType.direction !== "both")
            throw new GoJuniorRuntimeError(`cannot make directional channel type ${typeText}`);
        if (expression.args.length > 2)
            throw new GoJuniorRuntimeError("make channel accepts at most one buffer size");
        const capacity = expression.args[1]
            ? toNonNegativeLength(await evaluateExpression(expression.args[1], context), "channel buffer size")
            : 0;
        return new RuntimeChannel(chanType.elementType, capacity, context);
    }
    const arrayType = parseArrayOrSliceTypeText(typeText, context);
    if (arrayType) {
        if (arrayType.length !== undefined)
            throw new GoJuniorRuntimeError(`cannot make array type ${typeText}; use a slice type`);
        if (expression.args.length < 2 || expression.args.length > 3) {
            throw new GoJuniorRuntimeError("make slice expects length and optional capacity");
        }
        const length = toNonNegativeLength(await evaluateExpression(expression.args[1], context), "slice length");
        const capacity = expression.args[2]
            ? toNonNegativeLength(await evaluateExpression(expression.args[2], context), "slice capacity")
            : length;
        if (capacity < length)
            throw new GoJuniorRuntimeError("slice capacity is smaller than length");
        return makeRuntimeSlice(arrayType.elementType, length, capacity, context, arrayType.typeText);
    }
    throw new GoJuniorRuntimeError(`cannot make ${typeText}`);
}
function spreadLastArgument(args) {
    if (args.length === 0)
        return args;
    const last = args[args.length - 1];
    if (last instanceof RuntimeTypedNilValue && parseArrayOrSliceTypeText(last.typeName)) {
        return args.slice(0, -1);
    }
    if (!Array.isArray(last)) {
        throw new GoJuniorRuntimeError(`${formatValue(last ?? null)} is not spreadable`);
    }
    return [...args.slice(0, -1), ...last];
}
function appendValues(target, values) {
    const typedNilSlice = target instanceof RuntimeTypedNilValue && parseArrayOrSliceTypeText(target.typeName);
    if (target !== null && !typedNilSlice && !Array.isArray(target)) {
        throw new GoJuniorRuntimeError(`${formatValue(target)} is not appendable`);
    }
    const base = Array.isArray(target) ? target : [];
    const appended = [...base, ...values];
    const previousCapacity = sliceCapacity(base);
    const needed = appended.length;
    arrayCapacities.set(appended, Math.max(previousCapacity, needed));
    const typeText = Array.isArray(target) ? arrayTypeTexts.get(target) : typedNilSlice ? target.typeName : undefined;
    if (typeText)
        markArrayType(appended, typeText);
    return appended;
}
function clearValue(target, context) {
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
function copyValues(target, source) {
    target = unwrapNamed(target);
    source = unwrapNamed(source);
    if (!Array.isArray(target))
        throw new GoJuniorRuntimeError("copy destination must be a slice");
    const sourceValues = isRuntimeString(source)
        ? [...goStringBytes(source)].map((value) => BigInt(value))
        : source;
    if (!Array.isArray(sourceValues))
        throw new GoJuniorRuntimeError("copy source must be a slice or string");
    const count = Math.min(target.length, sourceValues.length);
    for (let index = 0; index < count; index += 1) {
        const value = isRuntimeString(source) ? sourceValues[index] ?? null : getArrayElement(sourceValues, index);
        setArrayElement(target, index, value);
    }
    return count;
}
function minMaxValues(args, mode) {
    if (args.length === 0)
        throw new GoJuniorRuntimeError(`${mode} expects at least one argument`);
    let best = unwrapNamed(args[0] ?? null);
    for (const arg of args.slice(1)) {
        const candidate = unwrapNamed(arg);
        const comparison = compareValues(candidate, best);
        if ((mode === "min" && comparison < 0) || (mode === "max" && comparison > 0))
            best = candidate;
    }
    return best;
}
async function callRuntime(callee, args, context) {
    if (isRuntimeCallable(callee) || isGoJuniorFunction(callee)) {
        return await callee.call(args, context);
    }
    throw new GoJuniorRuntimeError(`${formatValue(callee)} is not callable`);
}
async function getSelector(expression, context) {
    const methodExpression = methodExpressionForSelector(expression, context);
    if (methodExpression)
        return methodExpression;
    const object = await evaluateExpression(expression.object, context);
    if (object instanceof SheetBinding) {
        context.recordSheetCellRead(object.name, expression.field);
        return object.get(expression.field);
    }
    const syscallJSTypeMethod = syscallJSTypeSelector(object, expression.field);
    if (syscallJSTypeMethod)
        return syscallJSTypeMethod;
    const struct = structFromValue(object);
    if (struct) {
        const fieldValue = struct.get(expression.field);
        if (fieldValue !== undefined)
            return fieldValue;
        const promotedField = promotedFieldAccessor(struct, expression.field, context);
        if (promotedField)
            return promotedField.get() ?? null;
    }
    const runtimeObject = unwrapNamed(dereferenceIfPointer(object));
    if (isRuntimeObject(runtimeObject)) {
        const value = runtimeObject[expression.field];
        if (value !== undefined)
            return value;
    }
    const fmtError = expression.field === "Error" ? fmtErrorMessage(object) : undefined;
    if (fmtError !== undefined) {
        return hostCallable("fmt.errorString.Error", () => fmtError);
    }
    let method = methodForValue(object, expression.field, context);
    if (!method) {
        const pointer = await pointerToExpressionIfAddressable(expression.object, context);
        if (pointer)
            method = methodForValue(pointer, expression.field, context);
    }
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
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}
async function pointerToExpressionIfAddressable(expression, context) {
    if (expression.kind !== "Identifier" &&
        expression.kind !== "SelectorExpression" &&
        expression.kind !== "IndexExpression") {
        return undefined;
    }
    try {
        return await pointerToExpression(expression, context);
    }
    catch {
        return undefined;
    }
}
function methodExpressionForSelector(expression, context) {
    const receiver = methodExpressionReceiverType(expression.object, context);
    if (!receiver)
        return undefined;
    const method = context.methodFor(receiver.baseType, expression.field);
    if (method) {
        if (method.pointerReceiver && !receiver.pointer)
            return undefined;
        return methodExpressionValue(method, receiver.argumentType);
    }
    const interfaceMethod = interfaceMethodForType(receiver.baseType, expression.field, context);
    if (interfaceMethod)
        return dynamicMethodExpressionValue(receiver.argumentType, expression.field, interfaceMethod.signature);
    if (promotedMethodAvailableForType(receiver.baseType, expression.field, context, receiver.pointer)) {
        return dynamicMethodExpressionValue(receiver.argumentType, expression.field);
    }
    return undefined;
}
function methodExpressionReceiverType(expression, context) {
    if (expression.kind === "Identifier") {
        if (context.hasBinding(expression.name) || !context.isKnownType(expression.name))
            return undefined;
        const type = normalizeTypeText(expression.name);
        return { baseType: type, argumentType: type, pointer: false };
    }
    if (expression.kind === "TypeExpression") {
        return methodExpressionReceiverTypeText(expression.type.text, context);
    }
    if (expression.kind === "UnaryExpression" && expression.operator === "*") {
        const base = methodExpressionReceiverType(expression.operand, context);
        if (!base)
            return undefined;
        return { baseType: base.baseType, argumentType: `*${base.baseType}`, pointer: true };
    }
    return undefined;
}
function methodExpressionReceiverTypeText(typeText, context) {
    const type = normalizeTypeText(typeText);
    if (!type)
        return undefined;
    if (type.startsWith("*")) {
        const base = type.slice(1);
        if (!runtimeKnowsTypeText(base, context))
            return undefined;
        return { baseType: base, argumentType: type, pointer: true };
    }
    if (!runtimeKnowsTypeText(type, context))
        return undefined;
    return { baseType: type, argumentType: type, pointer: false };
}
function runtimeKnowsTypeText(type, context) {
    return context.isKnownType(type) ||
        parseAnonymousStructTypeText(type) !== undefined ||
        parseAnonymousInterfaceTypeText(type) !== undefined;
}
function interfaceMethodForType(typeText, methodName, context) {
    const interfaceType = interfaceTarget(typeText, context);
    return interfaceType?.methods.find((method) => method.name === methodName);
}
function promotedMethodAvailableForType(typeText, methodName, context, addressable = false, seen = new Set()) {
    const receiver = normalizeReceiverType(typeText);
    const baseType = receiver.baseType;
    if (seen.has(baseType))
        return false;
    seen.add(baseType);
    const typeDef = context.typeDef(baseType) ?? parseAnonymousStructTypeText(baseType);
    if (!typeDef)
        return false;
    for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
        const embedded = normalizeReceiverType(field.type.text);
        const direct = context.methodFor(embedded.baseType, methodName);
        const embeddedAddressable = addressable || receiver.pointer || embedded.pointer;
        if (direct && (!direct.pointerReceiver || embeddedAddressable))
            return true;
        if (promotedMethodAvailableForType(embedded.baseType, methodName, context, embeddedAddressable, seen))
            return true;
    }
    return false;
}
async function setSelector(expression, value, context) {
    const object = await evaluateExpression(expression.object, context);
    if (object instanceof SheetBinding) {
        object.set(expression.field, value);
        return;
    }
    const struct = structFromValue(object);
    if (struct) {
        const field = structFieldAccessor(struct, expression.field, context);
        if (!field)
            throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${expression.field}`);
        field.set(prepareAssignableToType(value, field.type.type.text, `field ${expression.field}`, context));
        return;
    }
    if (isRuntimeObject(object)) {
        object[expression.field] = value;
        return;
    }
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}
async function getSpreadsheetRange(expression, context) {
    const sheet = await evaluateExpression(expression.start.object, context);
    if (!(sheet instanceof SheetBinding)) {
        throw new GoJuniorRuntimeError("spreadsheet ranges must start with a sheet namespace");
    }
    context.recordSheetRangeRead(sheet.name, expression.start.field, expression.endCell);
    return sheet.range(expression.start.field, expression.endCell);
}
function getIndex(object, index) {
    object = unwrapNamed(object);
    if (object instanceof RuntimeMap)
        return object.get(index);
    const numericIndex = toNumber(index);
    if (isRuntimeString(object))
        return goStringByteAt(object, numericIndex);
    if (isRuntimeObject(object))
        return object[String(index)] ?? null;
    if (object instanceof RuntimePointer)
        object = object.get();
    object = unwrapNamed(object);
    if (Array.isArray(object))
        return getArrayElement(object, numericIndex);
    if (isRuntimeString(object))
        return goStringByteAt(object, numericIndex);
    throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
}
async function setIndex(expression, value, context) {
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
        if (!Array.isArray(target))
            throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
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
async function pointerToExpression(expression, context) {
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
        if (!struct)
            throw new GoJuniorRuntimeError("address-of selector requires a struct value");
        const field = structFieldAccessor(struct, expression.field, context);
        if (!field)
            throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${expression.field}`);
        const typeName = normalizeTypeText(field.type.type.text);
        return new RuntimePointer(typeName, () => field.get() ?? null, (next) => {
            field.set(prepareAssignableToType(next, field.type.type.text, `field ${expression.field}`, context));
        });
    }
    if (expression.kind === "IndexExpression") {
        const object = await evaluateExpression(expression.object, context);
        const index = await evaluateExpression(expression.index, context);
        if (!Array.isArray(object))
            throw new GoJuniorRuntimeError("address-of index requires an array or slice value");
        const numericIndex = toNumber(index);
        const typeName = indexResultTypeText(expressionDeclaredTypeText(expression.object, context)) ??
            pointerTypeName(getArrayElement(object, numericIndex));
        return new RuntimePointer(typeName, () => getArrayElement(object, numericIndex), (next) => {
            setArrayElement(object, numericIndex, next);
        });
    }
    throw new GoJuniorRuntimeError("expression is not addressable");
}
function dereference(value) {
    if (value instanceof RuntimeTypedNilValue && value.typeName.startsWith("*")) {
        throw new GoJuniorRuntimeError("nil pointer dereference");
    }
    if (!(value instanceof RuntimePointer))
        throw new GoJuniorRuntimeError(`${formatValue(value)} is not a pointer`);
    return value.get();
}
function dereferenceIfPointer(value) {
    return value instanceof RuntimePointer ? value.get() : value;
}
function structFromValue(value) {
    const actual = dereferenceIfPointer(value);
    return actual instanceof RuntimeStruct ? actual : undefined;
}
function structFieldAccessor(struct, fieldName, context) {
    const direct = directStructFieldAccessor(struct, fieldName, context);
    if (direct)
        return direct;
    return promotedFieldAccessor(struct, fieldName, context);
}
function directStructFieldAccessor(struct, fieldName, context) {
    const typeDef = structTypeDefFor(struct, context);
    const field = typeDef?.fields.find((candidate) => candidate.name === fieldName);
    if (!field)
        return undefined;
    return {
        type: field,
        get: () => struct.get(fieldName),
        set: (value) => struct.set(fieldName, value)
    };
}
function promotedFieldAccessor(struct, fieldName, context, seen = new Set()) {
    if (seen.has(struct.typeName))
        return undefined;
    seen.add(struct.typeName);
    const typeDef = structTypeDefFor(struct, context);
    if (!typeDef)
        return undefined;
    const matches = [];
    for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
        const embedded = structFromValue(struct.get(field.name) ?? null);
        if (!embedded)
            continue;
        const direct = directStructFieldAccessor(embedded, fieldName, context);
        if (direct) {
            matches.push(direct);
            continue;
        }
        const promoted = promotedFieldAccessor(embedded, fieldName, context, seen);
        if (promoted)
            matches.push(promoted);
    }
    if (matches.length > 1)
        throw new GoJuniorRuntimeError(`ambiguous promoted field ${fieldName}`);
    return matches[0];
}
function structTypeDefFor(struct, context) {
    return context.typeDef(struct.typeName) ?? parseAnonymousStructTypeText(struct.typeName);
}
function methodForValue(value, methodName, context) {
    if (value instanceof RuntimeInterfaceValue) {
        if (value.value === null)
            return undefined;
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
        if (direct)
            return { method: direct, receiver: value };
    }
    if (actual instanceof RuntimeStruct)
        return promotedMethodForStruct(actual, methodName, context, addressable);
    return undefined;
}
function promotedMethodForStruct(struct, methodName, context, addressable = false, seen = new Set()) {
    if (seen.has(struct.typeName))
        return undefined;
    seen.add(struct.typeName);
    const typeDef = structTypeDefFor(struct, context);
    if (!typeDef)
        return undefined;
    const matches = [];
    for (const field of typeDef.fields.filter((candidate) => candidate.embedded)) {
        const receiver = struct.get(field.name);
        if (receiver === undefined || receiver === null)
            continue;
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
        if (!embedded)
            continue;
        const embeddedAddressable = addressable || receiver instanceof RuntimePointer || isTypedNilPointer(receiver);
        const promoted = promotedMethodForStruct(embedded, methodName, context, embeddedAddressable, seen);
        if (promoted)
            matches.push(promoted);
    }
    if (matches.length > 1)
        throw new GoJuniorRuntimeError(`ambiguous promoted method ${methodName}`);
    return matches[0];
}
function pointerToStructField(struct, field, context) {
    const typeName = normalizeReceiverType(field.type.text).baseType;
    return new RuntimePointer(typeName, () => struct.get(field.name) ?? defaultValueForDeclarationType(field.type, context), (next) => {
        struct.set(field.name, prepareAssignableToType(next, field.type.text, `field ${field.name}`, context));
    });
}
function receiverTypeName(value) {
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null ? undefined : receiverTypeName(value.value);
    const actual = dereferenceIfPointer(value);
    if (actual instanceof RuntimeStruct)
        return actual.typeName;
    if (actual instanceof RuntimeNamedValue)
        return actual.typeName;
    if (actual instanceof RuntimeTypedNilValue)
        return nilReceiverTypeName(actual);
    return undefined;
}
function nilReceiverTypeName(value) {
    return value.typeName.startsWith("*") ? value.typeName.slice(1) : value.typeName;
}
function isTypedNilPointer(value) {
    return value instanceof RuntimeTypedNilValue && value.typeName.startsWith("*");
}
function assertAssignableToStructField(struct, fieldName, value, context) {
    const field = structFieldAccessor(struct, fieldName, context);
    if (!field)
        throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${fieldName}`);
    assertAssignableToType(value, field.type.type.text, `field ${fieldName}`, context);
}
function pointerTypeName(value) {
    const actual = dereferenceIfPointer(value);
    if (actual instanceof RuntimeInterfaceValue)
        return actual.interfaceType;
    if (actual instanceof RuntimeTypedNilValue)
        return actual.typeName;
    if (actual instanceof RuntimeNamedValue)
        return actual.typeName;
    if (actual instanceof RuntimeStruct)
        return actual.typeName;
    if (isComplexValue(actual))
        return "complex128";
    if (typeof actual === "bigint")
        return "int";
    if (typeof actual === "number")
        return "float64";
    if (isRuntimeString(actual))
        return "string";
    if (typeof actual === "boolean")
        return "bool";
    return "interface{}";
}
function inferredTypeText(value) {
    if (value === null)
        return undefined;
    if (value instanceof RuntimeNamedValue)
        return value.typeName;
    if (value instanceof RuntimeInterfaceValue)
        return value.interfaceType;
    if (value instanceof RuntimeTypedNilValue)
        return value.typeName;
    if (typeof value === "bigint")
        return "int64";
    if (typeof value === "number")
        return "float64";
    if (isComplexValue(value))
        return "complex128";
    if (isRuntimeString(value))
        return "string";
    if (typeof value === "boolean")
        return "bool";
    if (Array.isArray(value))
        return arrayTypeTexts.get(value);
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return `map[${value.keyType}]${value.valueType}`;
    if (value instanceof RuntimeStruct)
        return value.typeName;
    if (value instanceof RuntimePointer)
        return `*${value.typeName}`;
    return undefined;
}
function getSlice(object, start, end, max) {
    if (object instanceof RuntimePointer)
        object = object.get();
    object = unwrapNamed(object);
    const startIndex = start === undefined ? undefined : toNumber(start);
    const endIndex = end === undefined ? undefined : toNumber(end);
    const maxIndex = max === undefined ? undefined : toNumber(max);
    if (Array.isArray(object)) {
        const low = startIndex ?? 0;
        const high = endIndex ?? object.length;
        const capEnd = maxIndex ?? sliceCapacity(object);
        if (capEnd < high) {
            throw new GoJuniorRuntimeError(`slice max is smaller than high bound (len=${object.length}, cap=${sliceCapacity(object)}, high=${high}, max=${capEnd}, type=${arrayTypeTexts.get(object) ?? "[]"})`);
        }
        const result = Array.from({ length: Math.max(0, high - low) }, (_item, index) => getArrayElement(object, low + index));
        arrayCapacities.set(result, Math.max(0, capEnd - low));
        const view = arrayViews.get(object);
        markArrayView(result, view?.source ?? object, (view?.offset ?? 0) + low);
        const sourceType = parseArrayOrSliceTypeText(arrayTypeTexts.get(object) ?? "");
        if (sourceType)
            markArrayType(result, `[]${sourceType.elementType}`);
        return result;
    }
    if (max !== undefined)
        throw new GoJuniorRuntimeError("three-index slicing is only supported for arrays and slices");
    if (isRuntimeString(object))
        return goStringSlice(object, startIndex, endIndex);
    throw new GoJuniorRuntimeError(`${formatValue(object)} is not sliceable`);
}
function defaultValueForDeclarationType(type, context) {
    return defaultValueForTypeText(type?.text ?? "", context);
}
function defaultValueForTypeText(typeText, context) {
    const resolvedTypeText = context?.resolveImportedTypeText(typeText) ?? typeText;
    const mapType = parseMapTypeText(resolvedTypeText);
    if (mapType)
        return new RuntimeMap(mapType.keyType, mapType.valueType, context);
    if (parseChanTypeText(resolvedTypeText))
        return null;
    const arrayType = parseArrayOrSliceTypeText(resolvedTypeText, context);
    if (arrayType) {
        if (arrayType.length === undefined && !arrayType.inferLength)
            return new RuntimeTypedNilValue(arrayType.typeText);
        const values = arrayType.length === undefined || arrayType.inferLength
            ? []
            : Array.from({ length: arrayType.length }, () => defaultValueForTypeText(arrayType.elementType, context));
        markArrayType(values, arrayType.typeText);
        return values;
    }
    const type = normalizeTypeText(resolvedTypeText);
    const intrinsic = defaultIntrinsicNamedValue(type, context);
    if (intrinsic)
        return intrinsic;
    const interfaceType = interfaceTarget(type, context);
    if (interfaceType)
        return new RuntimeInterfaceValue(type, null);
    if (isUnsafePointerType(type))
        return new RuntimeTypedNilValue(type);
    if (type.startsWith("*"))
        return new RuntimeTypedNilValue(type);
    const anonymousStruct = parseAnonymousStructTypeText(resolvedTypeText);
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
    return zeroValueForMapValue(type);
}
function defaultIntrinsicNamedValue(type, context) {
    if (type === "js.Value" || type === "syscall/js.Value")
        return syscallJSValue(undefined);
    if (type === "js.Func" || type === "syscall/js.Func")
        return syscallJSFuncOf(null);
    if (type === "js.Type" || type === "syscall/js.Type")
        return syscallJSTypeValue(0n);
    if (type === "sync.Map")
        return new RuntimeNamedValue(type, syncMapValue());
    if (type === "sync.Once")
        return new RuntimeNamedValue(type, syncOnceValue());
    if (type === "sync.Pool")
        return new RuntimeNamedValue(type, syncPoolValue());
    if (type === "sync.Mutex" || type === "sync.RWMutex" || type === "internal/sync.Mutex")
        return new RuntimeNamedValue(type, syncMutexValue(type));
    if (type === "atomic.Bool" || type === "sync/atomic.Bool")
        return new RuntimeNamedValue(type, atomicBoolValue());
    if (type === "atomic.Int32" || type === "sync/atomic.Int32")
        return new RuntimeNamedValue(type, atomicInt32Value());
    if (type === "atomic.Uint64" || type === "sync/atomic.Uint64")
        return new RuntimeNamedValue(type, atomicUint64Value());
    const atomicPointer = /^(?:atomic|sync\/atomic)\.Pointer(?:\[[\s\S]*\])?$/.exec(type);
    if (atomicPointer)
        return new RuntimeNamedValue(type, atomicPointerValue());
    return undefined;
}
function syncMapValue() {
    const entries = new Map();
    return {
        Load: hostCallable("sync.Map.Load", (args) => {
            const key = args[0] ?? null;
            const entry = entries.get(runtimeMapKeyId(key));
            return entry ? [entry.value, true] : [null, false];
        }),
        LoadOrStore: hostCallable("sync.Map.LoadOrStore", (args) => {
            const key = args[0] ?? null;
            const value = args[1] ?? null;
            const id = runtimeMapKeyId(key);
            const existing = entries.get(id);
            if (existing)
                return [existing.value, true];
            entries.set(id, { key, value });
            return [value, false];
        }),
        Store: hostCallable("sync.Map.Store", (args) => {
            const key = args[0] ?? null;
            entries.set(runtimeMapKeyId(key), { key, value: args[1] ?? null });
            return null;
        }),
        Range: hostCallable("sync.Map.Range", async (args, context) => {
            const fn = args[0] ?? null;
            for (const entry of entries.values()) {
                if (!toBool(await callRuntime(fn, [entry.key, entry.value], context)))
                    break;
            }
            return null;
        }),
        Delete: hostCallable("sync.Map.Delete", (args) => {
            entries.delete(runtimeMapKeyId(args[0] ?? null));
            return null;
        })
    };
}
function syncOnceValue() {
    let done = false;
    return {
        Do: hostCallable("sync.Once.Do", async (args, context) => {
            if (!done) {
                done = true;
                await callRuntime(args[0] ?? null, [], context);
            }
            return null;
        })
    };
}
function syncPoolValue(newFn) {
    const values = [];
    return {
        Get: hostCallable("sync.Pool.Get", async (_args, context) => {
            const value = values.pop();
            if (value !== undefined)
                return value;
            if (newFn !== undefined)
                return callRuntime(newFn, [], context);
            return null;
        }),
        Put: hostCallable("sync.Pool.Put", (args) => {
            const value = args[0] ?? null;
            if (value !== null)
                values.push(value);
            return null;
        })
    };
}
function syncMutexValue(name) {
    return {
        Lock: hostCallable(`${name}.Lock`, () => null),
        Unlock: hostCallable(`${name}.Unlock`, () => null),
        RLock: hostCallable(`${name}.RLock`, () => null),
        RUnlock: hostCallable(`${name}.RUnlock`, () => null)
    };
}
function atomicPointerValue() {
    let value = null;
    return {
        Load: hostCallable("sync/atomic.Pointer.Load", () => value),
        Store: hostCallable("sync/atomic.Pointer.Store", (args) => {
            value = args[0] ?? null;
            return null;
        }),
        Swap: hostCallable("sync/atomic.Pointer.Swap", (args) => {
            const previous = value;
            value = args[0] ?? null;
            return previous;
        }),
        CompareAndSwap: hostCallable("sync/atomic.Pointer.CompareAndSwap", (args) => {
            if (valueEqual(value, args[0] ?? null)) {
                value = args[1] ?? null;
                return true;
            }
            return false;
        })
    };
}
function atomicBoolValue() {
    let value = false;
    return {
        CompareAndSwap: hostCallable("sync/atomic.Bool.CompareAndSwap", (args) => {
            const oldValue = toBool(args[0] ?? false);
            if (value === oldValue) {
                value = toBool(args[1] ?? false);
                return true;
            }
            return false;
        }),
        Load: hostCallable("sync/atomic.Bool.Load", () => value),
        Store: hostCallable("sync/atomic.Bool.Store", (args) => {
            value = toBool(args[0] ?? false);
            return null;
        }),
        Swap: hostCallable("sync/atomic.Bool.Swap", (args) => {
            const previous = value;
            value = toBool(args[0] ?? false);
            return previous;
        })
    };
}
function atomicInt32Value() {
    let value = 0n;
    return {
        Add: hostCallable("sync/atomic.Int32.Add", (args) => {
            value = BigInt.asIntN(32, value + toBigInt(args[0] ?? 0n));
            return value;
        }),
        CompareAndSwap: hostCallable("sync/atomic.Int32.CompareAndSwap", (args) => {
            const oldValue = BigInt.asIntN(32, toBigInt(args[0] ?? 0n));
            if (value === oldValue) {
                value = BigInt.asIntN(32, toBigInt(args[1] ?? 0n));
                return true;
            }
            return false;
        }),
        Load: hostCallable("sync/atomic.Int32.Load", () => value),
        Store: hostCallable("sync/atomic.Int32.Store", (args) => {
            value = BigInt.asIntN(32, toBigInt(args[0] ?? 0n));
            return null;
        }),
        Swap: hostCallable("sync/atomic.Int32.Swap", (args) => {
            const previous = value;
            value = BigInt.asIntN(32, toBigInt(args[0] ?? 0n));
            return previous;
        })
    };
}
function atomicUint64Value() {
    let value = 0n;
    return {
        Add: hostCallable("sync/atomic.Uint64.Add", (args) => {
            value = BigInt.asUintN(64, value + toBigInt(args[0] ?? 0n));
            return value;
        }),
        Load: hostCallable("sync/atomic.Uint64.Load", () => value),
        Store: hostCallable("sync/atomic.Uint64.Store", (args) => {
            value = BigInt.asUintN(64, toBigInt(args[0] ?? 0n));
            return null;
        })
    };
}
function zeroValueForMapValue(typeText) {
    const type = normalizeTypeText(typeText);
    if (isIntegerType(type))
        return 0n;
    if (isFloatType(type))
        return 0;
    if (isComplexType(type))
        return complexValue(0, 0);
    if (type === "string")
        return "";
    if (type === "bool")
        return false;
    return null;
}
function convertValueToType(value, typeText, context, sourceTypeText) {
    const type = normalizeTypeText(context.resolveImportedTypeText(typeText));
    const sourceType = sourceTypeText ? context.resolveImportedTypeText(sourceTypeText) : undefined;
    const targetInterface = interfaceTarget(type, context);
    if (targetInterface)
        return prepareInterfaceAssignment(value, type, targetInterface, "conversion", context, sourceType);
    const alias = context.aliasType(type);
    if (alias && alias !== type) {
        return new RuntimeNamedValue(type, convertValueToType(value, alias, context, sourceTypeText));
    }
    const sourceIsUnsafePointer = value instanceof RuntimeNamedValue && isUnsafePointerType(value.typeName) ||
        Boolean(sourceType && isUnsafePointerType(sourceType));
    const actual = unwrapNamed(value);
    if (isUnsafePointerType(type)) {
        if (actual === null)
            return new RuntimeTypedNilValue(type);
        if (actual instanceof RuntimeTypedNilValue) {
            if (actual.typeName.startsWith("*") || isUnsafePointerType(actual.typeName))
                return new RuntimeNamedValue(type, actual);
            throwTypeError(actual, type, "conversion");
        }
        if (actual instanceof RuntimePointer || typeof actual === "bigint")
            return new RuntimeNamedValue(type, actual);
        throwTypeError(actual, type, "conversion");
    }
    if (type === "any" || type === "interface{}")
        return actual;
    if (type === "bool") {
        if (typeof actual !== "boolean")
            throwTypeError(actual, type, "conversion");
        return actual;
    }
    if (type === "uintptr" && sourceIsUnsafePointer) {
        return uintptrFromUnsafePointer(actual);
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
            if (actual.imag !== 0)
                throwTypeError(actual, type, "conversion");
            return roundFloatForType(actual.real, type);
        }
        return roundFloatForType(toFloat(actual), type);
    }
    if (isComplexType(type)) {
        return roundComplexForType(toComplex(actual), type);
    }
    if (type === "string") {
        if (isRuntimeString(actual))
            return ensureRuntimeGoString(actual);
        if (typeof actual === "bigint")
            return goStringFromRune(actual);
        if (Array.isArray(actual)) {
            return stringFromIntegerArray(actual, sourceType, context);
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
        if (!Array.isArray(actual))
            throwTypeError(actual, type, "conversion");
        if (arrayType.length !== undefined && !arrayType.inferLength && actual.length !== arrayType.length) {
            throwTypeError(actual, type, "conversion");
        }
        for (const item of actual) {
            assertAssignableToType(item, arrayType.elementType, "conversion element", context);
        }
        markArrayType(actual, arrayType.typeText);
        return actual;
    }
    if (context.typeDef(type) && actual instanceof RuntimeStruct && actual.typeName === type)
        return actual;
    if (type.startsWith("*")) {
        if (actual === null)
            return new RuntimeTypedNilValue(type);
        if (actual instanceof RuntimeTypedNilValue && isUnsafePointerType(actual.typeName))
            return new RuntimeTypedNilValue(type);
        if (actual instanceof RuntimePointer) {
            const targetType = type.slice(1);
            if (runtimeTypeAssignableMatch(actual.typeName, targetType))
                return actual;
            return reinterpretUnsafePointer(actual, targetType, context);
        }
        if (sourceIsUnsafePointer && typeof actual === "bigint") {
            if (actual === 0n)
                return new RuntimeTypedNilValue(type);
            return opaqueUnsafeAddressPointer(type.slice(1), context);
        }
    }
    if (actual === null && isNilAssignableType(type))
        return null;
    throw new GoJuniorRuntimeError(`unsupported conversion to ${type}`);
}
function uintptrFromUnsafePointer(value) {
    if (value === null)
        return 0n;
    if (value instanceof RuntimeTypedNilValue && (isUnsafePointerType(value.typeName) || value.typeName.startsWith("*")))
        return 0n;
    if (typeof value === "bigint")
        return BigInt.asUintN(64, value);
    if (value instanceof RuntimePointer)
        return BigInt(objectIdentityId(value));
    throwTypeError(value, "uintptr", "conversion");
}
function reinterpretUnsafePointer(pointer, targetType, context) {
    let fallback;
    const currentValue = () => {
        const value = pointer.get();
        if (runtimeValueMatchesType(value, targetType, context))
            return value;
        if (fallback === undefined)
            fallback = defaultValueForTypeText(targetType, context);
        return fallback;
    };
    return new RuntimePointer(targetType, currentValue, (next) => {
        fallback = prepareAssignableToType(next, targetType, `*${targetType}`, context);
    });
}
function opaqueUnsafeAddressPointer(targetType, context) {
    let value = defaultValueForTypeText(targetType, context);
    return new RuntimePointer(targetType, () => value, (next) => {
        value = prepareAssignableToType(next, targetType, `*${targetType}`, context);
    });
}
function integerArrayFromString(source, targetTypeText, context) {
    const arrayType = parseArrayOrSliceTypeText(targetTypeText, context);
    if (!arrayType || arrayType.length !== undefined)
        return undefined;
    const elementType = integerSequenceElementType(targetTypeText, context);
    if (elementType === "byte") {
        return [...goStringBytes(source)].map((value) => BigInt(value));
    }
    if (elementType === "rune") {
        return goStringRuneEntries(source).map((entry) => entry.rune);
    }
    return undefined;
}
function stringFromIntegerArray(values, sourceTypeText, context) {
    const elementType = integerSequenceElementType(sourceTypeText ?? arrayTypeTexts.get(values), context);
    if (elementType !== "byte" && elementType !== "rune") {
        throwTypeError(values, "string", "conversion");
    }
    const integers = values.map((value) => {
        const actual = unwrapNamed(value);
        if (typeof actual !== "bigint")
            throwTypeError(value, "string", "conversion");
        return actual;
    });
    if (elementType === "byte") {
        const bytes = integers.map((value) => Number(BigInt.asUintN(8, value)));
        return new RuntimeGoString(new Uint8Array(bytes));
    }
    return goStringFromRunes(integers);
}
function integerSequenceElementType(typeText, context) {
    if (!typeText)
        return undefined;
    let type = normalizeTypeText(typeText);
    const seen = new Set();
    while (context?.aliasType(type) && !seen.has(type)) {
        seen.add(type);
        type = normalizeTypeText(context.aliasType(type) ?? type);
    }
    const arrayType = parseArrayOrSliceTypeText(type, context);
    const elementType = normalizeTypeText(arrayType?.elementType ?? "");
    if (elementType === "byte" || elementType === "uint8")
        return "byte";
    if (elementType === "rune" || elementType === "int32")
        return "rune";
    return undefined;
}
function goStringFromRune(value) {
    const codePoint = Number(value);
    if (!Number.isInteger(codePoint) || codePoint < 0 || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff)) {
        return RuntimeGoString.fromUtf8Text("\ufffd");
    }
    return RuntimeGoString.fromUtf8Text(String.fromCodePoint(codePoint));
}
function prepareValueForAssignmentTarget(value, target, source, context) {
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
function prepareValueForTargetType(value, targetTypeText, source, context, role = "interface assignment") {
    const targetType = targetTypeText ? normalizeTypeText(targetTypeText) : "";
    const interfaceType = targetType ? interfaceTarget(targetType, context) : undefined;
    if (!interfaceType)
        return value;
    const dynamicType = expressionDynamicTypeText(source, context, value);
    return prepareInterfaceAssignment(value, targetType, interfaceType, role, context, dynamicType);
}
function prepareValueForParameterType(value, targetTypeText, source, context, role) {
    const targetType = targetTypeText ? normalizeTypeText(targetTypeText) : "";
    if (!targetType)
        return value;
    if (interfaceTarget(targetType, context)) {
        return prepareValueForTargetType(value, targetType, source, context, role);
    }
    return prepareAssignableToType(value, targetType, role, context);
}
function expressionDynamicTypeText(expression, context, value) {
    if (!expression)
        return inferredConcreteDynamicType(value);
    switch (expression.kind) {
        case "Identifier":
            if (expression.name === "nil")
                return undefined;
            if (context.isUntypedConstant(expression.name))
                return undefined;
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
            if (expression.operator !== "!")
                return expressionDeclaredTypeText(expression, context, value) ?? inferredConcreteDynamicType(value);
            return inferredConcreteDynamicType(value);
        case "BinaryExpression":
            return binaryExpressionTypeText(expression, context, value) ?? inferredConcreteDynamicType(value);
        default:
            return inferredConcreteDynamicType(value);
    }
}
function expressionDeclaredTypeText(expression, context, value) {
    if (!expression)
        return inferredConcreteDynamicType(value);
    switch (expression.kind) {
        case "Identifier":
            if (expression.name === "nil")
                return undefined;
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
            if (expression.operator === "!")
                return "bool";
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
function materializeValueForExpressionType(value, expression, context) {
    if (isUntypedConstantExpression(expression))
        return value;
    return materializeValueForType(value, expressionDeclaredTypeText(expression, context, value), context);
}
function materializeValueForType(value, typeText, context) {
    const type = normalizeTypeText(typeText ?? "");
    if (!type)
        return value;
    const base = numericAliasBaseType(type, context);
    if (isIntegerType(base)) {
        const actual = unwrapNamed(value);
        const converted = typeof actual === "bigint"
            ? convertIntegerToType(actual, base)
            : typeof actual === "number" && Number.isInteger(actual)
                ? convertIntegerToType(BigInt(actual), base)
                : undefined;
        if (converted !== undefined)
            return namedNumericResult(type, base, converted, context);
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
function namedNumericResult(type, base, value, context) {
    void type;
    void base;
    void context;
    return value;
}
function binaryExpressionTypeText(expression, context, value) {
    if (expression.operator === "&&" || expression.operator === "||" || comparisonOperators.has(expression.operator))
        return "bool";
    const leftType = expressionDeclaredTypeText(expression.left, context);
    const rightType = expressionDeclaredTypeText(expression.right, context);
    const leftUntyped = isUntypedConstantExpression(expression.left);
    const rightUntyped = isUntypedConstantExpression(expression.right);
    if (leftUntyped && rightUntyped)
        return undefined;
    if (expression.operator === "<<" || expression.operator === ">>")
        return leftUntyped ? undefined : leftType;
    if (leftType && rightUntyped && isNumericTypeText(leftType, context))
        return leftType;
    if (rightType && leftUntyped && isNumericTypeText(rightType, context))
        return rightType;
    if (leftType && rightType && normalizeTypeText(leftType) === normalizeTypeText(rightType))
        return leftType;
    return undefined;
}
const comparisonOperators = new Set(["==", "!=", "<", "<=", ">", ">="]);
function isUntypedConstantExpression(expression) {
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
function isNumericTypeText(typeText, context) {
    const base = numericAliasBaseType(typeText, context);
    return isIntegerType(base) || isFloatType(base) || isComplexType(base);
}
function numericAliasBaseType(typeText, context) {
    let type = normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
    const seen = new Set();
    while (context?.aliasType(type) && !seen.has(type)) {
        seen.add(type);
        type = normalizeTypeText(context.aliasType(type) ?? type);
    }
    return type;
}
function channelElementTypeText(typeText) {
    if (!typeText)
        return undefined;
    return parseChanTypeText(typeText)?.elementType;
}
function roundFloatForType(value, typeText) {
    return normalizeTypeText(typeText) === "float32" ? Math.fround(value) : value;
}
function roundComplexForType(value, typeText) {
    return normalizeTypeText(typeText) === "complex64"
        ? complexValue(Math.fround(value.real), Math.fround(value.imag))
        : value;
}
function sliceResultTypeText(typeText) {
    const type = normalizeTypeText(typeText ?? "");
    if (!type)
        return undefined;
    if (type === "string")
        return "string";
    const dereferenced = type.startsWith("*") ? type.slice(1) : type;
    const arrayType = parseArrayOrSliceTypeText(dereferenced);
    if (!arrayType)
        return undefined;
    return `[]${arrayType.elementType}`;
}
function indexResultTypeText(typeText) {
    const type = normalizeTypeText(typeText ?? "");
    if (!type)
        return undefined;
    if (type === "string")
        return "byte";
    const dereferenced = type.startsWith("*") ? type.slice(1) : type;
    const arrayType = parseArrayOrSliceTypeText(dereferenced);
    if (arrayType)
        return arrayType.elementType;
    const mapType = parseMapTypeText(dereferenced);
    if (mapType)
        return mapType.valueType;
    return undefined;
}
function inferredConcreteDynamicType(value) {
    if (value === undefined || value === null)
        return undefined;
    if (value instanceof RuntimeNamedValue)
        return value.typeName;
    if (value instanceof RuntimeInterfaceValue)
        return value.interfaceType;
    if (value instanceof RuntimeTypedNilValue)
        return value.typeName;
    if (typeof value === "bigint")
        return "int64";
    if (typeof value === "number")
        return "float64";
    if (isComplexValue(value))
        return "complex128";
    if (isRuntimeString(value))
        return "string";
    if (typeof value === "boolean")
        return "bool";
    if (Array.isArray(value))
        return arrayTypeTexts.get(value);
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return `map[${value.keyType}]${value.valueType}`;
    if (value instanceof RuntimeStruct)
        return value.typeName;
    if (value instanceof RuntimePointer)
        return `*${value.typeName}`;
    return undefined;
}
function prepareAssignableToType(value, typeText, role, context) {
    const type = normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
    if (!type || type === "<missing>")
        return value;
    const interfaceType = interfaceTarget(type, context);
    if (interfaceType)
        return prepareInterfaceAssignment(value, type, interfaceType, role, context);
    if (value instanceof RuntimeNamedValue) {
        if (value.typeName === type)
            return value;
        value = value.value;
    }
    const alias = context?.aliasType(type);
    if (alias && alias !== type && !context?.typeDef(type) && !context?.interfaceDef(type)) {
        prepareAssignableToType(value, alias, role, context);
        return value;
    }
    if (value === null) {
        if (type.startsWith("*"))
            return new RuntimeTypedNilValue(type);
        if (isNilAssignableType(type))
            return value;
        throw new GoJuniorRuntimeError(`${role} nil is not assignable to ${type}`);
    }
    if (value instanceof RuntimeTypedNilValue) {
        if (value.typeName === type || isNilAssignableType(type))
            return value;
        throwTypeError(value, type, role);
    }
    if (type === "string") {
        if (!isRuntimeString(value))
            throwTypeError(value, type, role);
        return ensureRuntimeGoString(value);
    }
    if (type === "bool") {
        if (typeof value !== "boolean")
            throwTypeError(value, type, role);
        return value;
    }
    if (isIntegerType(type)) {
        if (typeof value === "number" && Number.isInteger(value))
            value = BigInt(value);
        if (typeof value !== "bigint" || !integerInRange(value, type))
            throwTypeError(value, type, role);
        return value;
    }
    if (isFloatType(type)) {
        if (typeof value === "bigint")
            return roundFloatForType(Number(value), type);
        if (typeof value !== "number")
            throwTypeError(value, type, role);
        return roundFloatForType(value, type);
    }
    if (isComplexType(type)) {
        if (typeof value === "bigint" || typeof value === "number")
            return roundComplexForType(toComplex(value), type);
        if (!isComplexValue(value))
            throwTypeError(value, type, role);
        return roundComplexForType(value, type);
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        if (!(value instanceof RuntimeMap))
            throwTypeError(value, type, role);
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
        if (!(value instanceof RuntimePointer) || !runtimeTypeAssignableMatch(value.typeName, type.slice(1)))
            throwTypeError(value, type, role);
        return value;
    }
    const structType = context?.typeDef(type);
    if (structType) {
        if (!(value instanceof RuntimeStruct) || value.typeName !== structType.name)
            throwTypeError(value, type, role);
        return value;
    }
    const anonymousStruct = parseAnonymousStructTypeText(type);
    if (anonymousStruct) {
        if (!(value instanceof RuntimeStruct) || value.typeName !== anonymousStruct.name)
            throwTypeError(value, type, role);
        return value;
    }
    if (value instanceof RuntimeStruct) {
        if (value.typeName !== type)
            throwTypeError(value, type, role);
        return value;
    }
    const arrayType = parseArrayOrSliceTypeText(type, context);
    if (arrayType) {
        if (!Array.isArray(value))
            throwTypeError(value, type, role);
        if (arrayType.length !== undefined && !arrayType.inferLength && value.length !== arrayType.length) {
            throwTypeError(value, type, role);
        }
        for (const item of value) {
            assertAssignableToType(item, arrayType.elementType, `${role} element`, context);
        }
        return value;
    }
    if (type.startsWith("func(")) {
        if (!isRuntimeCallable(value) && !isGoJuniorFunction(value))
            throwTypeError(value, type, role);
        return value;
    }
    return value;
}
function assertAssignableToType(value, typeText, role, context) {
    prepareAssignableToType(value, typeText, role, context);
}
function interfaceTarget(type, context) {
    if (type === "any" || type === "interface{}")
        return { name: type, methods: [], embeds: [] };
    if (type === "error")
        return errorInterfaceType;
    return context?.interfaceDef(type) ?? parseAnonymousInterfaceTypeText(type);
}
function prepareInterfaceAssignment(value, type, interfaceType, role, context, dynamicType) {
    if (value instanceof RuntimeInterfaceValue) {
        const dynamicValue = value.value;
        if (dynamicValue === null)
            return new RuntimeInterfaceValue(type, null);
        const sourceInterface = interfaceTarget(value.interfaceType, context);
        if (sourceInterface && interfaceDefinitionImplementsInterface(sourceInterface, interfaceType)) {
            return new RuntimeInterfaceValue(type, dynamicValue);
        }
        if (context && !valueImplementsInterface(dynamicValue, interfaceType, context))
            throwTypeError(value, type, role);
        return new RuntimeInterfaceValue(type, dynamicValue);
    }
    if (value === null)
        return new RuntimeInterfaceValue(type, null);
    const dynamicValue = boxDynamicInterfaceValue(value, dynamicType);
    if (context && !valueImplementsInterface(dynamicValue, interfaceType, context))
        throwTypeError(value, type, role);
    return new RuntimeInterfaceValue(type, dynamicValue);
}
function interfaceDefinitionImplementsInterface(source, target) {
    for (const method of target.methods) {
        const candidate = source.methods.find((sourceMethod) => sourceMethod.name === method.name);
        if (!candidate || !signaturesCompatible(candidate.signature, method.signature))
            return false;
    }
    return true;
}
function boxDynamicInterfaceValue(value, dynamicType) {
    if (!dynamicType || value instanceof RuntimeNamedValue || value instanceof RuntimeTypedNilValue)
        return value;
    const type = canonicalRuntimeTypeName(normalizeTypeText(dynamicType));
    if (!type || type === "nil" || type === "any" || type === "interface{}")
        return value;
    if (value instanceof RuntimeStruct || value instanceof RuntimePointer || value instanceof RuntimeMap || value instanceof RuntimeChannel)
        return value;
    if (isRuntimeCallable(value) || isGoJuniorFunction(value) || value instanceof SheetBinding)
        return value;
    return new RuntimeNamedValue(type, value);
}
function valueImplementsInterface(value, interfaceType, context) {
    if (value instanceof RuntimeInterfaceValue) {
        return value.value === null || valueImplementsInterface(value.value, interfaceType, context);
    }
    if (fmtErrorMessage(value) !== undefined) {
        return interfaceType.methods.every((method) => method.name === "Error" &&
            method.signature.parameters.length === 0 &&
            method.signature.results.length === 1 &&
            normalizeTypeText(method.signature.results[0]?.type.text ?? "") === "string");
    }
    const receiver = dereferenceIfPointer(value);
    const receiverType = receiver instanceof RuntimeStruct
        ? receiver.typeName
        : receiver instanceof RuntimeNamedValue
            ? receiver.typeName
            : receiver instanceof RuntimeTypedNilValue
                ? nilReceiverTypeName(receiver)
                : undefined;
    if (!receiverType)
        return interfaceType.methods.length === 0;
    for (const method of interfaceType.methods) {
        const candidate = methodForValue(value, method.name, context);
        if (!candidate)
            return false;
        if (candidate.method.pointerReceiver &&
            !(candidate.receiver instanceof RuntimePointer) &&
            !isTypedNilPointer(candidate.receiver))
            return false;
        if (!signaturesCompatible(candidate.method.declaration.signature, method.signature))
            return false;
    }
    return true;
}
function signaturesCompatible(actual, expected) {
    if (actual.parameters.length !== expected.parameters.length || actual.results.length !== expected.results.length)
        return false;
    return actual.parameters.every((param, index) => {
        const expectedParam = expected.parameters[index];
        if (!expectedParam)
            return false;
        return normalizeTypeText(param.type.text) === normalizeTypeText(expectedParam.type.text) &&
            param.variadic === expectedParam.variadic;
    }) && actual.results.every((result, index) => {
        const expectedResult = expected.results[index];
        if (!expectedResult)
            return false;
        return normalizeTypeText(result.type.text) === normalizeTypeText(expectedResult.type.text) &&
            result.variadic === expectedResult.variadic;
    });
}
function runtimeValueMatchesType(value, typeText, context) {
    const type = normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
    if (value instanceof RuntimeInterfaceValue) {
        if (type === "nil")
            return value.value === null;
        if (value.value === null)
            return false;
        return runtimeValueMatchesType(value.value, type, context);
    }
    if (type === "nil")
        return value === null;
    if (value === null)
        return false;
    if (!type || type === "<missing>")
        return false;
    if (type === "any" || type === "interface{}")
        return true;
    const interfaceType = context?.interfaceDef(type);
    if (interfaceType && context)
        return valueImplementsInterface(value, interfaceType, context);
    if (value instanceof RuntimeNamedValue) {
        return canonicalRuntimeTypeName(normalizeTypeText(value.typeName)) === canonicalRuntimeTypeName(type);
    }
    if (value instanceof RuntimeTypedNilValue) {
        if (isUnsafePointerType(type))
            return isUnsafePointerType(value.typeName) || value.typeName.startsWith("*");
        if (type.startsWith("*"))
            return runtimeTypeAssignableMatch(value.typeName, type);
    }
    if (isUnsafePointerType(type))
        return value instanceof RuntimeNamedValue && isUnsafePointerType(value.typeName);
    if (type === "string")
        return isRuntimeString(value);
    if (type === "bool")
        return typeof value === "boolean";
    if (isIntegerType(type))
        return type === "int64" && typeof value === "bigint" && integerInRange(value, type);
    if (isFloatType(type))
        return typeof value === "number";
    if (isComplexType(type))
        return isComplexValue(value);
    if (type.startsWith("*"))
        return value instanceof RuntimePointer && runtimeTypeAssignableMatch(value.typeName, type.slice(1));
    if (value instanceof RuntimeStruct)
        return value.typeName === type;
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
    if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type))
        return Array.isArray(value);
    if (type.startsWith("func("))
        return isRuntimeCallable(value) || isGoJuniorFunction(value);
    return false;
}
function canonicalRuntimeTypeName(type) {
    switch (type) {
        case "byte":
            return "uint8";
        case "rune":
            return "int32";
        default:
            return type;
    }
}
function runtimeTypeAssignableMatch(actual, expected) {
    return assignmentRuntimeTypeName(actual) === assignmentRuntimeTypeName(expected);
}
function assignmentRuntimeTypeName(type) {
    const normalized = canonicalRuntimeTypeName(normalizeTypeText(type));
    if (normalized === "int")
        return "int64";
    if (normalized.startsWith("*"))
        return `*${assignmentRuntimeTypeName(normalized.slice(1))}`;
    const mapType = parseMapTypeText(normalized);
    if (mapType)
        return `map[${assignmentRuntimeTypeName(mapType.keyType)}]${assignmentRuntimeTypeName(mapType.valueType)}`;
    const chanType = parseChanTypeText(normalized);
    if (chanType)
        return `${chanType.direction === "receive" ? "<-" : chanType.direction === "send" ? "chan<-" : "chan"}${assignmentRuntimeTypeName(chanType.elementType)}`;
    return normalized;
}
function valueLength(value) {
    value = unwrapNamed(value);
    if (value instanceof RuntimeTypedNilValue) {
        const type = normalizeTypeText(value.typeName);
        if (type.startsWith("*")) {
            const arrayType = parseArrayOrSliceTypeText(type.slice(1));
            if (arrayType?.length !== undefined && !arrayType.inferLength)
                return arrayType.length;
        }
        if (type.startsWith("[]") || type.startsWith("map[") || parseChanTypeText(type))
            return 0;
    }
    if (value instanceof RuntimePointer) {
        const arrayType = parseArrayOrSliceTypeText(value.typeName);
        if (arrayType?.length !== undefined && !arrayType.inferLength)
            return arrayType.length;
    }
    if (isRuntimeString(value))
        return ensureRuntimeGoString(value).byteLength();
    if (Array.isArray(value))
        return value.length;
    if (value instanceof RuntimeMap)
        return value.size();
    if (value instanceof RuntimeChannel)
        return value.len();
    throw new GoJuniorRuntimeError(`${formatValue(value)} has no len`);
}
function valueCapacity(value) {
    value = unwrapNamed(value);
    if (value instanceof RuntimeTypedNilValue) {
        const type = normalizeTypeText(value.typeName);
        if (type.startsWith("*")) {
            const arrayType = parseArrayOrSliceTypeText(type.slice(1));
            if (arrayType?.length !== undefined && !arrayType.inferLength)
                return arrayType.length;
        }
        if (type.startsWith("[]") || parseChanTypeText(type))
            return 0;
    }
    if (value instanceof RuntimePointer) {
        const arrayType = parseArrayOrSliceTypeText(value.typeName);
        if (arrayType?.length !== undefined && !arrayType.inferLength)
            return arrayType.length;
    }
    if (Array.isArray(value))
        return sliceCapacity(value);
    if (value instanceof RuntimeChannel)
        return value.cap();
    throw new GoJuniorRuntimeError(`${formatValue(value)} has no cap`);
}
function sliceCapacity(value) {
    return arrayCapacities.get(value) ?? value.length;
}
function markArrayType(value, typeText) {
    arrayTypeTexts.set(value, normalizeTypeText(typeText));
}
function markArrayView(value, source, offset) {
    arrayViews.set(value, { source, offset });
}
function getArrayElement(value, index) {
    const view = arrayViews.get(value);
    if (view)
        return getArrayElement(view.source, view.offset + index);
    return value[index] ?? null;
}
function setArrayElement(value, index, item) {
    value[index] = item;
    const view = arrayViews.get(value);
    if (view)
        setArrayElement(view.source, view.offset + index, item);
}
function toNonNegativeLength(value, role) {
    const length = toNumber(value);
    if (length < 0)
        throw new GoJuniorRuntimeError(`${role} must be non-negative`);
    return length;
}
function throwTypeError(value, type, role) {
    throw new GoJuniorRuntimeError(`${role} ${formatValue(value)} is not assignable to ${type}`);
}
function parseMapTypeText(typeText) {
    const type = normalizeTypeText(typeText);
    if (!type.startsWith("map["))
        return undefined;
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
                if (!keyType || !valueType)
                    return undefined;
                return { keyType, valueType };
            }
        }
    }
    return undefined;
}
function parseChanTypeText(typeText) {
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
function parseArrayOrSliceTypeText(typeText, context) {
    const type = normalizeTypeText(typeText);
    if (!type.startsWith("["))
        return undefined;
    const close = type.indexOf("]");
    if (close < 0)
        return undefined;
    const lengthText = type.slice(1, close);
    const elementType = type.slice(close + 1);
    if (!elementType)
        return undefined;
    if (lengthText === "") {
        return { elementType, inferLength: false, typeText: `[]${elementType}` };
    }
    if (lengthText === "...") {
        return { elementType, inferLength: true, typeText: `[...]${elementType}` };
    }
    const length = resolveArrayLengthText(lengthText, context);
    if (length === undefined)
        return undefined;
    return { elementType, length, inferLength: false, typeText: `[${length}]${elementType}` };
}
function resolveArrayLengthText(lengthText, context) {
    const constant = evaluateArrayLengthConstant(lengthText, context);
    if (constant === undefined)
        return undefined;
    if (constant < 0n || constant > BigInt(Number.MAX_SAFE_INTEGER))
        return undefined;
    return Number(constant);
}
function evaluateArrayLengthConstant(text, context) {
    const parser = new ArrayLengthConstantParser(text, context);
    try {
        return parser.parse();
    }
    catch {
        return undefined;
    }
}
class ArrayLengthConstantParser {
    text;
    context;
    offset = 0;
    constructor(text, context) {
        this.text = text;
        this.context = context;
    }
    parse() {
        const value = this.expression(1);
        this.skipSpace();
        return this.offset === this.text.length ? value : undefined;
    }
    expression(minPrecedence) {
        let left = this.unary();
        if (left === undefined)
            return undefined;
        while (true) {
            this.skipSpace();
            const operator = this.peekBinaryOperator();
            if (!operator || operator.precedence < minPrecedence)
                return left;
            this.offset += operator.text.length;
            const right = this.expression(operator.precedence + 1);
            if (right === undefined)
                return undefined;
            left = applyArrayLengthConstantOperator(operator.text, left, right);
            if (left === undefined)
                return undefined;
        }
    }
    unary() {
        this.skipSpace();
        if (this.consume("+"))
            return this.unary();
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
    primary() {
        this.skipSpace();
        if (this.consume("(")) {
            const value = this.expression(1);
            this.skipSpace();
            if (!this.consume(")"))
                return undefined;
            return value;
        }
        const integer = this.consumeIntegerLiteral();
        if (integer !== undefined)
            return integer;
        const identifier = this.consumeIdentifier();
        if (identifier)
            return this.lookupIdentifier(identifier);
        return undefined;
    }
    lookupIdentifier(name) {
        if (!this.context)
            return undefined;
        let value = unwrapNamed(this.context.lookup(name.split(".")[0] ?? name));
        for (const part of name.split(".").slice(1)) {
            if (!isRuntimeObject(value))
                return undefined;
            value = unwrapNamed(value[part] ?? null);
        }
        if (typeof value === "bigint")
            return value;
        if (typeof value === "number" && Number.isSafeInteger(value))
            return BigInt(value);
        return undefined;
    }
    consumeIntegerLiteral() {
        const start = this.offset;
        if (this.text.startsWith("0x", start) || this.text.startsWith("0X", start)) {
            this.offset += 2;
            while (this.offset < this.text.length && /[0-9a-fA-F_]/.test(this.text[this.offset]))
                this.offset += 1;
            return this.integerFromDigits(start, 16, 2);
        }
        if (this.text.startsWith("0b", start) || this.text.startsWith("0B", start)) {
            this.offset += 2;
            while (this.offset < this.text.length && /[01_]/.test(this.text[this.offset]))
                this.offset += 1;
            return this.integerFromDigits(start, 2, 2);
        }
        if (this.text.startsWith("0o", start) || this.text.startsWith("0O", start)) {
            this.offset += 2;
            while (this.offset < this.text.length && /[0-7_]/.test(this.text[this.offset]))
                this.offset += 1;
            return this.integerFromDigits(start, 8, 2);
        }
        while (this.offset < this.text.length && /[0-9_]/.test(this.text[this.offset]))
            this.offset += 1;
        if (this.offset === start)
            return undefined;
        const raw = this.text.slice(start, this.offset).replace(/_/g, "");
        if (raw.length > 1 && raw.startsWith("0")) {
            if (!/^[0-7]+$/.test(raw))
                return undefined;
            return BigInt(`0o${raw.slice(1) || "0"}`);
        }
        return BigInt(raw);
    }
    integerFromDigits(start, radix, prefixLength) {
        const raw = this.text.slice(start + prefixLength, this.offset).replace(/_/g, "");
        if (!raw)
            return undefined;
        if (radix === 16)
            return BigInt(`0x${raw}`);
        if (radix === 8)
            return BigInt(`0o${raw}`);
        return BigInt(`0b${raw}`);
    }
    consumeIdentifier() {
        this.skipSpace();
        const first = this.consumeIdentifierSegment();
        if (!first)
            return undefined;
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
    consumeIdentifierSegment() {
        const start = this.offset;
        if (start >= this.text.length || !/[A-Za-z_]/.test(this.text[start]))
            return undefined;
        this.offset += 1;
        while (this.offset < this.text.length && /[A-Za-z0-9_]/.test(this.text[this.offset]))
            this.offset += 1;
        return this.text.slice(start, this.offset);
    }
    peekBinaryOperator() {
        const operators = [
            ["&^", 5],
            ["<<", 5], [">>", 5],
            ["*", 5], ["/", 5], ["%", 5], ["&", 5],
            ["+", 4], ["-", 4], ["|", 4], ["^", 4]
        ];
        const operator = operators.find(([text]) => this.text.startsWith(text, this.offset));
        return operator ? { text: operator[0], precedence: operator[1] } : undefined;
    }
    consume(text) {
        this.skipSpace();
        if (!this.text.startsWith(text, this.offset))
            return false;
        this.offset += text.length;
        return true;
    }
    skipSpace() {
        while (this.offset < this.text.length && /\s/.test(this.text[this.offset]))
            this.offset += 1;
    }
}
function applyArrayLengthConstantOperator(operator, left, right) {
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
function parseAnonymousStructTypeText(typeText) {
    const name = normalizeTypeText(typeText);
    const cached = anonymousStructTypeCache.get(name);
    if (cached)
        return cached;
    const match = /^struct\s*\{([\s\S]*)\}$/.exec(typeText.trim());
    if (!match)
        return undefined;
    const body = match[1]?.trim() ?? "";
    const fields = body === ""
        ? []
        : splitTopLevel(body, ";").flatMap(parseAnonymousStructField);
    const typeDef = { name, fields };
    anonymousStructTypeCache.set(name, typeDef);
    return typeDef;
}
function parseAnonymousStructField(text) {
    const field = text.trim();
    if (!field)
        return [];
    const withoutTag = stripStructFieldTag(field);
    const split = firstAnonymousStructFieldNameTypeSplit(withoutTag);
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
function firstAnonymousStructFieldNameTypeSplit(field) {
    let depth = 0;
    for (let index = 0; index < field.length; index++) {
        const char = field[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth++;
        if (char === "]" || char === ")" || char === "}")
            depth--;
        if (depth !== 0 || !/\s/.test(char))
            continue;
        const namesText = field.slice(0, index).trim();
        if (!anonymousStructFieldNameList(namesText))
            continue;
        let next = index;
        while (next < field.length && /\s/.test(field[next]))
            next++;
        if (next < field.length)
            return index;
    }
    return -1;
}
function anonymousStructFieldNameList(namesText) {
    const names = splitTopLevel(namesText, ",").map((name) => name.trim());
    return names.length > 0 && names.every((name) => /^[A-Za-z_]\w*$/.test(name));
}
function parseAnonymousInterfaceTypeText(typeText) {
    const name = normalizeTypeText(typeText);
    const cached = anonymousInterfaceTypeCache.get(name);
    if (cached)
        return cached;
    const match = /^interface\s*\{([\s\S]*)\}$/.exec(typeText.trim());
    if (!match)
        return undefined;
    const body = match[1]?.trim() ?? "";
    const methods = [];
    const embeds = [];
    for (const item of body === "" ? [] : splitTopLevel(body, ";")) {
        const text = item.trim();
        if (!text)
            continue;
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
function parseSignatureParameterListText(text) {
    const trimmed = text.trim();
    if (!trimmed)
        return [];
    return splitTopLevel(trimmed, ",").flatMap(parseSignatureParameterText);
}
function parseSignatureParameterText(text) {
    const parameter = text.trim();
    if (!parameter)
        return [];
    const split = lastTopLevelWhitespace(parameter);
    const namesText = split < 0 ? "" : parameter.slice(0, split).trim();
    let typeText = split < 0 ? parameter : parameter.slice(split).trim();
    const variadic = typeText.startsWith("...");
    if (variadic)
        typeText = typeText.slice(3);
    const type = { text: typeText };
    if (!namesText)
        return [{ type, variadic }];
    return splitTopLevel(namesText, ",")
        .map((name) => name.trim())
        .filter(Boolean)
        .map((name) => ({ name, type, variadic }));
}
function parseSignatureResultText(text) {
    const trimmed = text.trim();
    if (!trimmed)
        return [];
    if (trimmed.startsWith("(")) {
        const close = matchingParen(trimmed, 0);
        if (close === trimmed.length - 1)
            return parseSignatureParameterListText(trimmed.slice(1, -1));
    }
    return parseSignatureParameterText(trimmed);
}
function matchingParen(text, openIndex) {
    let depth = 0;
    for (let index = openIndex; index < text.length; index++) {
        const char = text[index] ?? "";
        if (char === "(")
            depth++;
        if (char === ")") {
            depth--;
            if (depth === 0)
                return index;
        }
    }
    return -1;
}
function stripStructFieldTag(field) {
    let depth = 0;
    for (let index = 0; index < field.length; index++) {
        const char = field[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth++;
        if (char === "]" || char === ")" || char === "}")
            depth--;
        if (depth === 0 && char === "`")
            return field.slice(0, index).trim();
    }
    return field;
}
function lastTopLevelWhitespace(text) {
    let depth = 0;
    let result = -1;
    for (let index = 0; index < text.length; index++) {
        const char = text[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth++;
        if (char === "]" || char === ")" || char === "}")
            depth--;
        if (depth === 0 && /\s/.test(char))
            result = index;
    }
    return result;
}
function splitTopLevel(text, delimiter) {
    const parts = [];
    let depth = 0;
    let start = 0;
    for (let index = 0; index < text.length; index++) {
        const char = text[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth++;
        if (char === "]" || char === ")" || char === "}")
            depth--;
        if (depth === 0 && char === delimiter) {
            parts.push(text.slice(start, index));
            start = index + 1;
        }
    }
    parts.push(text.slice(start));
    return parts;
}
function embeddedFieldNameFromTypeText(typeText) {
    const normalized = normalizeTypeText(typeText);
    const base = normalized.startsWith("*") ? normalized.slice(1) : normalized;
    const selector = base.lastIndexOf(".");
    return selector >= 0 ? base.slice(selector + 1) : base;
}
function normalizeTypeText(typeText) {
    return typeText.replace(/\s+/g, "");
}
function isRuntimeString(value) {
    return typeof value === "string" || value instanceof RuntimeGoString;
}
function ensureRuntimeGoString(value) {
    return value instanceof RuntimeGoString ? value : RuntimeGoString.fromUtf8Text(value);
}
function goStringBytes(value) {
    if (value instanceof RuntimeGoString)
        return value.bytes;
    if (typeof value === "string")
        return new TextEncoder().encode(value);
    throwTypeError(value, "string", "conversion");
}
function goStringText(value) {
    return value instanceof RuntimeGoString ? value.text() : value;
}
function goStringByteAt(value, index) {
    return ensureRuntimeGoString(value).byteAt(index);
}
function goStringSlice(value, start, end) {
    return ensureRuntimeGoString(value).slice(start, end);
}
function goStringKey(value) {
    return [...goStringBytes(value)].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}
function compareGoStrings(left, right) {
    const leftBytes = goStringBytes(left);
    const rightBytes = goStringBytes(right);
    const count = Math.min(leftBytes.length, rightBytes.length);
    for (let index = 0; index < count; index += 1) {
        const leftByte = leftBytes[index] ?? 0;
        const rightByte = rightBytes[index] ?? 0;
        if (leftByte !== rightByte)
            return leftByte < rightByte ? -1 : 1;
    }
    return leftBytes.length === rightBytes.length ? 0 : leftBytes.length < rightBytes.length ? -1 : 1;
}
function goStringFromRunes(values) {
    return RuntimeGoString.fromUtf8Text(values.map((value) => goStringFromRune(value).text()).join(""));
}
function goStringRuneEntries(value) {
    const bytes = goStringBytes(value);
    const entries = [];
    for (let index = 0; index < bytes.length;) {
        const decoded = decodeUtf8Rune(bytes, index);
        entries.push({ index, rune: BigInt(decoded.rune) });
        index += decoded.width;
    }
    return entries;
}
function decodeUtf8Rune(bytes, index) {
    const first = bytes[index] ?? 0;
    if (first < 0x80)
        return { rune: first, width: 1 };
    const replacement = { rune: 0xfffd, width: 1 };
    if (first < 0xc2)
        return replacement;
    if (first < 0xe0) {
        const b1 = bytes[index + 1];
        if (b1 === undefined || (b1 & 0xc0) !== 0x80)
            return replacement;
        return { rune: ((first & 0x1f) << 6) | (b1 & 0x3f), width: 2 };
    }
    if (first < 0xf0) {
        const b1 = bytes[index + 1];
        const b2 = bytes[index + 2];
        if (b1 === undefined || b2 === undefined || (b1 & 0xc0) !== 0x80 || (b2 & 0xc0) !== 0x80)
            return replacement;
        if (first === 0xe0 && b1 < 0xa0)
            return replacement;
        if (first === 0xed && b1 >= 0xa0)
            return replacement;
        return { rune: ((first & 0x0f) << 12) | ((b1 & 0x3f) << 6) | (b2 & 0x3f), width: 3 };
    }
    if (first < 0xf5) {
        const b1 = bytes[index + 1];
        const b2 = bytes[index + 2];
        const b3 = bytes[index + 3];
        if (b1 === undefined || b2 === undefined || b3 === undefined || (b1 & 0xc0) !== 0x80 || (b2 & 0xc0) !== 0x80 || (b3 & 0xc0) !== 0x80)
            return replacement;
        if (first === 0xf0 && b1 < 0x90)
            return replacement;
        if (first === 0xf4 && b1 >= 0x90)
            return replacement;
        return { rune: ((first & 0x07) << 18) | ((b1 & 0x3f) << 12) | ((b2 & 0x3f) << 6) | (b3 & 0x3f), width: 4 };
    }
    return replacement;
}
function goStringFromLiteralRaw(raw) {
    if (raw.length >= 2 && raw.startsWith("`") && raw.endsWith("`")) {
        return RuntimeGoString.fromUtf8Text(raw.slice(1, -1).replace(/\r/g, ""));
    }
    if (raw.length >= 2 && raw.startsWith("\"") && raw.endsWith("\"")) {
        return new RuntimeGoString(decodeGoStringLiteralBytes(raw.slice(1, -1)));
    }
    return RuntimeGoString.fromUtf8Text(raw);
}
function decodeGoStringLiteralBytes(value) {
    const bytes = [];
    for (let index = 0; index < value.length; index += 1) {
        const char = value[index] ?? "";
        if (char !== "\\") {
            bytes.push(...new TextEncoder().encode(char));
            continue;
        }
        const next = value[index + 1] ?? "";
        index += 1;
        switch (next) {
            case "a":
                bytes.push(0x07);
                break;
            case "b":
                bytes.push(0x08);
                break;
            case "f":
                bytes.push(0x0c);
                break;
            case "n":
                bytes.push(0x0a);
                break;
            case "r":
                bytes.push(0x0d);
                break;
            case "t":
                bytes.push(0x09);
                break;
            case "v":
                bytes.push(0x0b);
                break;
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
                }
                else {
                    bytes.push(...new TextEncoder().encode(next));
                }
        }
    }
    return new Uint8Array(bytes.map((byte) => byte & 0xff));
}
function externalizeRuntimeValue(value) {
    if (value instanceof RuntimeGoString)
        return value.text();
    if (value instanceof RuntimeNamedValue)
        return new RuntimeNamedValue(value.typeName, externalizeRuntimeValue(value.value));
    if (value instanceof RuntimeInterfaceValue)
        return new RuntimeInterfaceValue(value.interfaceType, value.value === null ? null : externalizeRuntimeValue(value.value));
    if (Array.isArray(value))
        return value.map(externalizeRuntimeValue);
    if (value instanceof RuntimeStruct) {
        const struct = new RuntimeStruct(value.typeName);
        for (const [name, item] of value.orderedFields())
            struct.set(name, externalizeRuntimeValue(item));
        return struct;
    }
    return value;
}
function externalizePackageRuntimeValue(value, packageName) {
    if (value instanceof RuntimeGoString)
        return value.text();
    if (isGoJuniorFunction(value)) {
        return {
            ...value,
            async call(args, context) {
                return externalizePackageRuntimeValue(await value.call(args, context), packageName);
            }
        };
    }
    if (value instanceof RuntimeNamedValue) {
        return new RuntimeNamedValue(qualifyLocalRuntimeTypeName(value.typeName, packageName), externalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeInterfaceValue) {
        return new RuntimeInterfaceValue(qualifyLocalRuntimeTypeName(value.interfaceType, packageName), value.value === null ? null : externalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimePointer) {
        return new RuntimePointer(qualifyLocalRuntimeTypeName(value.typeName, packageName), () => externalizePackageRuntimeValue(value.get(), packageName), (next) => value.set(next));
    }
    if (Array.isArray(value))
        return value.map((item) => externalizePackageRuntimeValue(item, packageName));
    if (value instanceof RuntimeStruct) {
        return new RuntimeExternalizedStruct(value, packageName);
    }
    return value;
}
function resolveRuntimeFunctionDeclarationTypeImports(declaration, context) {
    return {
        ...declaration,
        signature: resolveRuntimeSignatureTypeImports(declaration.signature, context),
        ...(declaration.receiver ? {
            receiver: {
                ...declaration.receiver,
                type: resolveRuntimeTypeNodeImports(declaration.receiver.type, context)
            }
        } : {})
    };
}
function resolveRuntimeSignatureTypeImports(signature, context) {
    return {
        parameters: signature.parameters.map((parameter) => ({
            ...parameter,
            type: resolveRuntimeTypeNodeImports(parameter.type, context)
        })),
        results: signature.results.map((result) => ({
            ...result,
            type: resolveRuntimeTypeNodeImports(result.type, context)
        }))
    };
}
function resolveRuntimeTypeNodeImports(type, context) {
    return { text: context.resolveImportedTypeText(type.text) };
}
function rewriteRuntimeTypeText(typeText, resolveQualifier) {
    let output = "";
    let index = 0;
    while (index < typeText.length) {
        const char = typeText[index] ?? "";
        if (char === "[" && !isRuntimeTypePathOrIdentifierChar(typeText[index - 1] ?? "")) {
            const close = matchingSquareBracket(typeText, index);
            if (close > index) {
                output += typeText.slice(index, close + 1);
                index = close + 1;
                continue;
            }
        }
        if (isRuntimeTypeIdentifierStart(char) && !isRuntimeTypePathOrIdentifierChar(typeText[index - 1] ?? "")) {
            const qualifierStart = index;
            index += 1;
            while (index < typeText.length && isRuntimeTypeIdentifierPart(typeText[index] ?? ""))
                index += 1;
            const qualifier = typeText.slice(qualifierStart, index);
            if (typeText[index] === "." &&
                isRuntimeTypeIdentifierStart(typeText[index + 1] ?? "")) {
                const nameStart = index + 1;
                let nameEnd = nameStart + 1;
                while (nameEnd < typeText.length && isRuntimeTypeIdentifierPart(typeText[nameEnd] ?? ""))
                    nameEnd += 1;
                if (typeText[nameEnd] !== "/") {
                    const importPath = resolveQualifier(qualifier);
                    if (importPath) {
                        output += `${importPath}.${typeText.slice(nameStart, nameEnd)}`;
                        index = nameEnd;
                        continue;
                    }
                }
            }
            output += qualifier;
            continue;
        }
        output += char;
        index += 1;
    }
    return output;
}
function matchingSquareBracket(text, openIndex) {
    let depth = 0;
    for (let index = openIndex; index < text.length; index++) {
        const char = text[index] ?? "";
        if (char === "[")
            depth++;
        if (char === "]") {
            depth--;
            if (depth === 0)
                return index;
        }
    }
    return -1;
}
function isRuntimeTypeIdentifierStart(char) {
    return /[A-Za-z_]/.test(char);
}
function isRuntimeTypeIdentifierPart(char) {
    return /[A-Za-z0-9_]/.test(char);
}
function isRuntimeTypePathOrIdentifierChar(char) {
    return /[A-Za-z0-9_/.]/.test(char);
}
function qualifyLocalRuntimeTypeName(typeName, packageName) {
    const type = normalizeTypeText(typeName);
    if (!type || !packageName || isPredeclaredType(type) || /^[A-Za-z_]\w*\.[A-Za-z_]\w*(?:\[.*\])?$/.test(type))
        return type;
    if (type.startsWith("*"))
        return `*${qualifyLocalRuntimeTypeName(type.slice(1), packageName)}`;
    if (type.startsWith("[]"))
        return `[]${qualifyLocalRuntimeTypeName(type.slice(2), packageName)}`;
    const array = /^\[([0-9.]*)\]([\s\S]+)$/.exec(type);
    if (array)
        return `[${array[1] ?? ""}]${qualifyLocalRuntimeTypeName(array[2] ?? "", packageName)}`;
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${qualifyLocalRuntimeTypeName(mapType.keyType, packageName)}]${qualifyLocalRuntimeTypeName(mapType.valueType, packageName)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const prefix = chanType.direction === "receive" ? "<-chan " : chanType.direction === "send" ? "chan<- " : "chan ";
        return `${prefix}${qualifyLocalRuntimeTypeName(chanType.elementType, packageName)}`;
    }
    if (type.startsWith("func(") || type.startsWith("interface{") || type.startsWith("struct{"))
        return type;
    const generic = /^([A-Za-z_]\w*)\[(.*)\]$/.exec(type);
    if (generic)
        return `${packageName}.${generic[1]}[${generic[2]}]`;
    return /^[A-Za-z_]\w*$/.test(type) ? `${packageName}.${type}` : type;
}
function qualifyRuntimeSignature(signature, packageName) {
    return {
        parameters: signature.parameters.map((parameter) => ({
            ...parameter,
            type: { text: qualifyLocalRuntimeTypeName(parameter.type.text, packageName) }
        })),
        results: signature.results.map((result) => ({
            ...result,
            type: { text: qualifyLocalRuntimeTypeName(result.type.text, packageName) }
        }))
    };
}
function qualifyRuntimeMethodDeclaration(declaration, packageName) {
    return {
        ...declaration,
        signature: qualifyRuntimeSignature(declaration.signature, packageName),
        ...(declaration.receiver ? {
            receiver: {
                ...declaration.receiver,
                type: { text: qualifyLocalRuntimeTypeName(declaration.receiver.type.text, packageName) }
            }
        } : {})
    };
}
function genericBaseTypeName(typeText) {
    const type = normalizeTypeText(typeText);
    if (type.startsWith("[") || type.startsWith("map[") || !type.endsWith("]"))
        return type;
    const bracket = type.indexOf("[");
    return bracket > 0 ? type.slice(0, bracket) : type;
}
function isTypeArgumentExpression(expression, context) {
    const typeText = typeArgumentText(expression);
    if (typeText === undefined)
        return false;
    const parts = splitTopLevelTypeArguments(typeText);
    return parts.length > 1
        ? parts.every((part) => context.isKnownType(part))
        : context.isKnownType(typeText);
}
function splitTopLevelTypeArguments(typeText) {
    const parts = [];
    let depth = 0;
    let start = 0;
    for (let index = 0; index < typeText.length; index += 1) {
        const char = typeText[index];
        if (char === "[" || char === "(" || char === "{")
            depth += 1;
        else if (char === "]" || char === ")" || char === "}")
            depth -= 1;
        else if (char === "," && depth === 0) {
            parts.push(typeText.slice(start, index).trim());
            start = index + 1;
        }
    }
    parts.push(typeText.slice(start).trim());
    return parts.filter((part) => part.length > 0);
}
function typeArgumentText(expression) {
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
function runtimeMapKeyId(key, options = {}) {
    const actual = unwrapNamed(key);
    if (actual === null)
        return "nil";
    if (actual instanceof RuntimeInterfaceValue) {
        return actual.value === null
            ? `iface:${actual.interfaceType}:nil`
            : `iface:${actual.interfaceType}:${runtimeMapKeyId(actual.value, options)}`;
    }
    if (actual instanceof RuntimeTypedNilValue)
        return `typednil:${actual.typeName}`;
    if (typeof actual === "boolean")
        return `b:${actual}`;
    if (isRuntimeString(actual))
        return `s:${goStringKey(actual)}`;
    if (typeof actual === "bigint")
        return `i:${actual}`;
    if (typeof actual === "number")
        return `f:${mapNumberKeyId(actual, options)}`;
    if (isComplexValue(actual))
        return `c:${mapNumberKeyId(actual.real, options)}:${mapNumberKeyId(actual.imag, options)}`;
    if (Array.isArray(actual))
        return `a:[${actual.map((item) => runtimeMapKeyId(item, options)).join(",")}]`;
    if (actual instanceof RuntimeStruct) {
        return `st:${actual.typeName}{${actual.orderedFields().map(([name, item]) => `${name}:${runtimeMapKeyId(item, options)}`).join(",")}}`;
    }
    return `o:${objectIdentityId(actual)}`;
}
function mapNumberKeyId(value, options) {
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
function objectIdentityId(value) {
    const existing = objectMapKeyIds.get(value);
    if (existing !== undefined)
        return existing;
    const id = nextObjectMapKeyId;
    nextObjectMapKeyId += 1;
    objectMapKeyIds.set(value, id);
    return id;
}
function isNilAssignableType(type) {
    return type === "error" ||
        type === "any" ||
        type === "interface{}" ||
        isUnsafePointerType(type) ||
        type.startsWith("*") ||
        type.startsWith("[]") ||
        type.startsWith("map[") ||
        parseChanTypeText(type) !== undefined ||
        type.startsWith("func(");
}
function assertComparableType(typeText, context) {
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
        for (const field of structType.fields)
            assertComparableType(field.type.text, context);
        return;
    }
    const anonymousStruct = parseAnonymousStructTypeText(type);
    if (anonymousStruct) {
        for (const field of anonymousStruct.fields)
            assertComparableType(field.type.text, context);
    }
}
function assertComparableValue(value, role) {
    value = unwrapNamed(value);
    if (value instanceof RuntimeInterfaceValue) {
        if (value.value !== null)
            assertComparableValue(value.value, role);
        return;
    }
    if (value instanceof RuntimeTypedNilValue)
        return;
    if (value instanceof RuntimeMap || isRuntimeCallable(value) || isGoJuniorFunction(value)) {
        throw new GoJuniorRuntimeError(`${role} ${formatValue(value)} is not comparable`);
    }
    if (Array.isArray(value)) {
        for (const item of value)
            assertComparableValue(item, `${role} element`);
        return;
    }
    if (value instanceof RuntimeStruct) {
        for (const [field, item] of value.orderedFields())
            assertComparableValue(item, `${role} field ${field}`);
    }
}
function isIntegerType(type) {
    return /^(?:u?int(?:8|16|32|64)?|uintptr|byte|rune)$/.test(type);
}
function isFloatType(type) {
    return type === "float32" || type === "float64";
}
function isComplexType(type) {
    return type === "complex64" || type === "complex128";
}
function isUnsafePointerType(type) {
    return normalizeTypeText(type) === "unsafe.Pointer";
}
function isPredeclaredType(type) {
    return type === "bool" ||
        type === "string" ||
        type === "error" ||
        type === "any" ||
        type === "interface{}" ||
        isIntegerType(type) ||
        isFloatType(type) ||
        isComplexType(type);
}
function integerInRange(value, type) {
    const bounds = integerTypeBounds(type);
    return !bounds || (value >= bounds.min && value <= bounds.max);
}
function convertIntegerToType(value, type) {
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
function integerTypeBounds(type) {
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
function addValues(left, right) {
    const actualLeft = unwrapNamed(left);
    const actualRight = unwrapNamed(right);
    if (isRuntimeString(actualLeft) || isRuntimeString(actualRight)) {
        if (isRuntimeString(actualLeft) && isRuntimeString(actualRight))
            return ensureRuntimeGoString(actualLeft).concat(ensureRuntimeGoString(actualRight));
        throw new GoJuniorRuntimeError(`invalid operation: ${runtimeTypeText(actualLeft)} + ${runtimeTypeText(actualRight)} (mismatched types ${runtimeTypeText(actualLeft)} and ${runtimeTypeText(actualRight)})`);
    }
    return addNumbers(actualLeft, actualRight);
}
function runtimeTypeText(value) {
    return inferredTypeText(value) ?? pointerTypeName(value);
}
function addNumbers(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (isComplexValue(left) || isComplexValue(right)) {
        const a = toComplex(left);
        const b = toComplex(right);
        return complexValue(a.real + b.real, a.imag + b.imag);
    }
    if (typeof left === "bigint" && typeof right === "bigint")
        return left + right;
    return toFloat(left) + toFloat(right);
}
function subtractNumbers(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (isComplexValue(left) || isComplexValue(right)) {
        const a = toComplex(left);
        const b = toComplex(right);
        return complexValue(a.real - b.real, a.imag - b.imag);
    }
    if (typeof left === "bigint" && typeof right === "bigint")
        return left - right;
    return toFloat(left) - toFloat(right);
}
function multiplyNumbers(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (isComplexValue(left) || isComplexValue(right)) {
        const a = toComplex(left);
        const b = toComplex(right);
        return complexValue(a.real * b.real - a.imag * b.imag, a.real * b.imag + a.imag * b.real);
    }
    if (typeof left === "bigint" && typeof right === "bigint")
        return left * right;
    return toFloat(left) * toFloat(right);
}
function divideNumbers(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (isComplexValue(left) || isComplexValue(right)) {
        const a = toComplex(left);
        const b = toComplex(right);
        const denominator = b.real * b.real + b.imag * b.imag;
        return complexValue((a.real * b.real + a.imag * b.imag) / denominator, (a.imag * b.real - a.real * b.imag) / denominator);
    }
    if (typeof left === "bigint" && typeof right === "bigint")
        return left / right;
    return toFloat(left) / toFloat(right);
}
function moduloNumbers(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (typeof left === "bigint" && typeof right === "bigint")
        return left % right;
    return toFloat(left) % toFloat(right);
}
function negateNumber(value) {
    value = unwrapNamed(value);
    if (isComplexValue(value))
        return complexValue(-value.real, -value.imag);
    if (typeof value === "bigint")
        return -value;
    return -toFloat(value);
}
function numericIdentity(value) {
    value = unwrapNamed(value);
    if (isComplexValue(value))
        return value;
    if (typeof value === "bigint" || typeof value === "number")
        return value;
    throw new GoJuniorRuntimeError(`${formatValue(value)} is not numeric`);
}
function toBool(value) {
    if (typeof value !== "boolean") {
        throw new GoJuniorRuntimeError(`${formatValue(value)} is not a bool`);
    }
    return value;
}
function toFloat(value) {
    value = unwrapNamed(value);
    if (typeof value === "number")
        return value;
    if (typeof value === "bigint")
        return Number(value);
    throw new GoJuniorRuntimeError(`${formatValue(value)} is not numeric`);
}
function toNumber(value) {
    const number = toFloat(value);
    if (!Number.isInteger(number)) {
        throw new GoJuniorRuntimeError(`${formatValue(value)} is not an integer index`);
    }
    return number;
}
function bitwiseComplement(value) {
    return ~toBigInt(value);
}
function bitwiseAnd(left, right) {
    return toBigInt(left) & toBigInt(right);
}
function bitwiseOr(left, right) {
    return toBigInt(left) | toBigInt(right);
}
function bitwiseXor(left, right) {
    return toBigInt(left) ^ toBigInt(right);
}
function bitClear(left, right) {
    return toBigInt(left) & ~toBigInt(right);
}
function shiftLeft(left, right) {
    const shift = toBigInt(right);
    if (shift < 0n)
        throw new GoJuniorRuntimeError("negative shift count");
    return toBigInt(left) << shift;
}
function shiftRight(left, right) {
    const shift = toBigInt(right);
    if (shift < 0n)
        throw new GoJuniorRuntimeError("negative shift count");
    return toBigInt(left) >> shift;
}
function toBigInt(value) {
    value = unwrapNamed(value);
    if (typeof value === "bigint")
        return value;
    if (typeof value === "number" && Number.isInteger(value))
        return BigInt(value);
    throw new GoJuniorRuntimeError(`${formatValue(value)} is not an integer`);
}
function float32Bits(value) {
    const view = new DataView(new ArrayBuffer(4));
    view.setFloat32(0, value, true);
    return BigInt(view.getUint32(0, true));
}
function float32FromBits(bits) {
    const view = new DataView(new ArrayBuffer(4));
    view.setUint32(0, Number(BigInt.asUintN(32, bits)), true);
    return view.getFloat32(0, true);
}
function float64Bits(value) {
    const view = new DataView(new ArrayBuffer(8));
    view.setFloat64(0, value, true);
    return view.getBigUint64(0, true);
}
function float64FromBits(bits) {
    const view = new DataView(new ArrayBuffer(8));
    view.setBigUint64(0, BigInt.asUintN(64, bits), true);
    return view.getFloat64(0, true);
}
function complexValue(real, imag) {
    return { real, imag };
}
function isComplexValue(value) {
    return Boolean(value &&
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
        typeof value.imag === "number");
}
function toComplex(value) {
    value = unwrapNamed(value);
    if (isComplexValue(value))
        return value;
    return complexValue(toFloat(value), 0);
}
function unwrapNamed(value) {
    while (value instanceof RuntimeNamedValue)
        value = value.value;
    return value;
}
function compareValues(left, right) {
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
function valueEqual(left, right) {
    if (left instanceof RuntimeInterfaceValue || right instanceof RuntimeInterfaceValue) {
        return interfaceAwareEqual(left, right);
    }
    if (left instanceof RuntimeTypedNilValue || right instanceof RuntimeTypedNilValue) {
        if (left === null || right === null)
            return true;
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
        if (left.typeName !== right.typeName)
            return false;
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
function compareNumericValues(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint") {
        return left === right ? 0 : left < right ? -1 : 1;
    }
    const leftNumber = typeof left === "bigint" ? Number(left) : left;
    const rightNumber = typeof right === "bigint" ? Number(right) : right;
    if (Number.isNaN(leftNumber) || Number.isNaN(rightNumber))
        return Number.NaN;
    return leftNumber === rightNumber ? 0 : leftNumber < rightNumber ? -1 : 1;
}
function cmpCompare(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    const leftNaN = isRuntimeNaN(left);
    const rightNaN = isRuntimeNaN(right);
    if (leftNaN)
        return rightNaN ? 0 : -1;
    if (rightNaN)
        return 1;
    const compared = compareValues(left, right);
    if (Number.isNaN(compared))
        return 0;
    return compared < 0 ? -1 : compared > 0 ? 1 : 0;
}
function cmpLess(left, right) {
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    return (isRuntimeNaN(left) && !isRuntimeNaN(right)) || compareValues(left, right) < 0;
}
function isRuntimeNaN(value) {
    value = unwrapNamed(value);
    return typeof value === "number" && Number.isNaN(value);
}
function isRuntimeZeroValue(value) {
    value = unwrapNamed(value);
    if (value === null)
        return true;
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null;
    if (value instanceof RuntimeTypedNilValue)
        return true;
    if (typeof value === "boolean")
        return !value;
    if (typeof value === "bigint")
        return value === 0n;
    if (typeof value === "number")
        return Object.is(value, 0) || Object.is(value, -0);
    if (isRuntimeString(value))
        return goStringBytes(value).length === 0;
    if (isComplexValue(value))
        return value.real === 0 && value.imag === 0;
    return false;
}
function zeroValueLike(value) {
    value = unwrapNamed(value);
    if (typeof value === "boolean")
        return false;
    if (typeof value === "bigint")
        return 0n;
    if (typeof value === "number")
        return 0;
    if (isRuntimeString(value))
        return "";
    if (isComplexValue(value))
        return complexValue(0, 0);
    return null;
}
function runtimeAlignof(value) {
    return Math.max(1, Math.min(8, runtimeSizeof(value)));
}
function runtimeSizeof(value) {
    value = unwrapNamed(value);
    if (value === null)
        return 0;
    if (typeof value === "boolean")
        return 1;
    if (typeof value === "bigint" || typeof value === "number")
        return 8;
    if (isComplexValue(value))
        return 16;
    if (isRuntimeString(value))
        return 16;
    if (value instanceof RuntimePointer || value instanceof RuntimeChannel || value instanceof RuntimeMap)
        return 8;
    if (isRuntimeCallable(value))
        return 8;
    if (value instanceof RuntimeInterfaceValue)
        return 16;
    if (value instanceof RuntimeTypedNilValue)
        return value.typeName.startsWith("*") ? 8 : 0;
    if (Array.isArray(value))
        return 24;
    if (value instanceof RuntimeStruct) {
        let size = 0;
        for (const [, field] of value.orderedFields())
            size += runtimeSizeof(field);
        return size;
    }
    if (typeof value === "object")
        return 8;
    return 0;
}
function interfaceAwareEqual(left, right) {
    if (left instanceof RuntimeInterfaceValue && right === null)
        return left.value === null;
    if (right instanceof RuntimeInterfaceValue && left === null)
        return right.value === null;
    if (left instanceof RuntimeInterfaceValue && right instanceof RuntimeInterfaceValue) {
        if (left.value === null || right.value === null)
            return left.value === null && right.value === null;
        return valueEqual(left.value, right.value);
    }
    if (left instanceof RuntimeInterfaceValue) {
        if (left.value === null)
            return false;
        return valueEqual(left.value, right);
    }
    if (right instanceof RuntimeInterfaceValue) {
        if (right.value === null)
            return false;
        return valueEqual(left, right.value);
    }
    return false;
}
async function rangeEntries(source, context) {
    source = unwrapNamed(source);
    if (typeof source === "bigint" || typeof source === "number") {
        if (source < 0)
            return [];
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
function integerRangeKeyTypeText(sourceExpression, source, context) {
    const actual = unwrapNamed(source);
    if (typeof actual !== "bigint" && typeof actual !== "number")
        return undefined;
    return expressionDynamicTypeText(sourceExpression, context, source) ?? inferredTypeText(source);
}
async function iteratorFunctionEntries(source, context) {
    const entries = [];
    let index = 0n;
    const yieldFn = hostCallable("yield", (args) => {
        if (args.length === 0) {
            entries.push([index, index]);
            index += 1n;
        }
        else if (args.length === 1) {
            entries.push([args[0] ?? null, args[0] ?? null]);
        }
        else {
            entries.push([args[0] ?? null, args[1] ?? null]);
        }
        return true;
    });
    await callRuntime(source, [yieldFn], context);
    return entries;
}
function sprintf(format, args) {
    let argIndex = 0;
    return format.replace(/%(#)?[%vdsft]/g, (match, alternate) => {
        if (match === "%%")
            return "%";
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
async function sprintfAsync(format, args, context) {
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
async function formatFmtValue(verb, goSyntax, value, context) {
    if (!goSyntax) {
        const stringer = await fmtStringerValue(value, context);
        if (stringer !== undefined && (verb === "%v" || verb === "%s"))
            return stringer;
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
async function fmtStringerValue(value, context) {
    const method = methodForValue(value, "String", context);
    if (!method || method.method.declaration.signature.parameters.length !== 0)
        return undefined;
    const results = method.method.declaration.signature.results;
    if (results.length !== 1 || normalizeTypeText(results[0]?.type.text ?? "") !== "string")
        return undefined;
    const result = await callRuntime(boundMethodValue(method.method, method.receiver), [], context);
    return toStringValue(result);
}
function sprint(args) {
    let text = "";
    let previousWasString = false;
    for (const arg of args) {
        const currentIsString = runtimeValueIsString(arg);
        if (text && !previousWasString && !currentIsString)
            text += " ";
        text += formatValue(arg);
        previousWasString = currentIsString;
    }
    return text;
}
async function sprintAsync(args, context) {
    let text = "";
    let previousWasString = false;
    for (const arg of args) {
        const currentIsString = runtimeValueIsString(arg);
        if (text && !previousWasString && !currentIsString)
            text += " ";
        text += await formatFmtValue("%v", false, arg, context);
        previousWasString = currentIsString;
    }
    return text;
}
function sprintln(args) {
    return `${args.map(formatValue).join(" ")}\n`;
}
async function sprintlnAsync(args, context) {
    return `${(await Promise.all(args.map((arg) => formatFmtValue("%v", false, arg, context)))).join(" ")}\n`;
}
function runtimeValueIsString(value) {
    if (value instanceof RuntimeNamedValue)
        return runtimeValueIsString(value.value);
    if (value instanceof RuntimeInterfaceValue)
        return value.value !== null && runtimeValueIsString(value.value);
    return isRuntimeString(value);
}
function toStringValue(value) {
    if (isRuntimeString(value))
        return goStringText(value);
    return formatValue(value);
}
async function toStringValueAsync(value, context) {
    if (isRuntimeString(value))
        return goStringText(value);
    return await fmtStringerValue(value, context) ?? formatValue(value);
}
export function formatValue(value) {
    if (value instanceof RuntimeNamedValue)
        return formatValue(value.value);
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null ? "<nil>" : formatValue(value.value);
    if (value instanceof RuntimeTypedNilValue)
        return "<nil>";
    if (value === null)
        return "<nil>";
    if (typeof value === "bigint")
        return value.toString();
    if (isComplexValue(value))
        return `(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
    if (typeof value === "number" || typeof value === "boolean")
        return String(value);
    if (isRuntimeString(value))
        return goStringText(value);
    if (Array.isArray(value))
        return `[${value.map(formatValue).join(" ")}]`;
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatRuntimeMap(value);
    if (value instanceof RuntimeStruct)
        return formatRuntimeStruct(value);
    if (value instanceof RuntimePointer)
        return `&${formatValue(value.get())}`;
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    return `{${Object.entries(value).map(([key, item]) => `${key}:${formatValue(item)}`).join(" ")}}`;
}
export function formatReplValue(value) {
    if (value instanceof RuntimeNamedValue)
        return formatReplValue(value.value);
    if (value instanceof RuntimeInterfaceValue) {
        return value.value === null ? `${value.interfaceType}(nil)` : formatReplValue(value.value);
    }
    if (value instanceof RuntimeTypedNilValue)
        return `${value.typeName}(nil)`;
    if (value === null)
        return "<nil>";
    if (typeof value === "bigint")
        return value.toString();
    if (isComplexValue(value))
        return `(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
    if (typeof value === "number" || typeof value === "boolean")
        return String(value);
    if (isRuntimeString(value))
        return formatReplString(goStringText(value));
    if (Array.isArray(value))
        return `[${value.map(formatReplValue).join(" ")}]`;
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatReplMap(value);
    if (value instanceof RuntimeStruct)
        return formatReplStruct(value);
    if (value instanceof RuntimePointer)
        return `&${formatReplValue(value.get())}`;
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    return `{${Object.entries(value).map(([key, item]) => `${key}:${formatReplValue(item)}`).join(" ")}}`;
}
function formatGoSyntaxValue(value) {
    if (value instanceof RuntimeNamedValue) {
        return value.value instanceof RuntimeTypedNilValue
            ? `${value.typeName}(nil)`
            : `${value.typeName}(${formatGoSyntaxValue(value.value)})`;
    }
    if (value instanceof RuntimeInterfaceValue) {
        return value.value === null ? `${value.interfaceType}(nil)` : formatGoSyntaxValue(value.value);
    }
    if (value instanceof RuntimeTypedNilValue)
        return `${value.typeName}(nil)`;
    if (value === null)
        return "nil";
    if (typeof value === "bigint")
        return `${integerTypeName(value)}(${value.toString()})`;
    if (typeof value === "number")
        return `float64(${formatFloat(value)})`;
    if (isComplexValue(value))
        return `complex128(${formatFloat(value.real)}+${formatFloat(value.imag)}i)`;
    if (typeof value === "boolean")
        return `bool(${value})`;
    if (isRuntimeString(value))
        return `string(${JSON.stringify(goStringText(value))})`;
    if (Array.isArray(value))
        return `[]interface{}{${value.map(formatGoSyntaxValue).join(", ")}}`;
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatGoSyntaxMap(value);
    if (value instanceof RuntimeStruct)
        return formatGoSyntaxStruct(value);
    if (value instanceof RuntimePointer)
        return `&${formatGoSyntaxValue(value.get())}`;
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    return `map[string]interface{}{${Object.entries(value)
        .map(([key, item]) => `${JSON.stringify(key)}: ${formatGoSyntaxValue(item)}`)
        .join(", ")}}`;
}
function formatRuntimeMap(value) {
    const entries = value.orderedEntries()
        .map(([key, item]) => `${formatValue(key)}:${formatValue(item)}`)
        .join(" ");
    return `map[${value.keyType}]${value.valueType}{${entries}}`;
}
function formatFunctionValue(value) {
    return value.source?.trimEnd() || `<func ${value.name}>`;
}
function formatRuntimeStruct(value) {
    const fields = value.orderedFields()
        .map(([key, item]) => `${key}:${formatValue(item)}`)
        .join(" ");
    return `${value.typeName}{${fields}}`;
}
function formatReplMap(value) {
    const entries = value.orderedEntries()
        .map(([key, item]) => `${formatReplValue(key)}:${formatReplValue(item)}`)
        .join(", ");
    return `map[${value.keyType}]${value.valueType}{${entries}}`;
}
function formatReplStruct(value) {
    const fields = value.orderedFields()
        .map(([key, item]) => `${key}:${formatReplValue(item)}`)
        .join(", ");
    return `${value.typeName}{${fields}}`;
}
function formatReplString(value) {
    if (value.includes("\n") || value.includes("\""))
        return `\`${value.replace(/`/g, "\\`")}\``;
    return JSON.stringify(value);
}
function formatGoSyntaxMap(value) {
    const entries = value.orderedEntries()
        .map(([key, item]) => `${formatGoSyntaxValue(key)}: ${formatGoSyntaxValue(item)}`)
        .join(", ");
    return `map[${value.keyType}]${value.valueType}{${entries}}`;
}
function formatGoSyntaxStruct(value) {
    const fields = value.orderedFields()
        .map(([key, item]) => `${key}: ${formatGoSyntaxValue(item)}`)
        .join(", ");
    return `${value.typeName}{${fields}}`;
}
function integerTypeName(value) {
    return value >= -9223372036854775808n && value <= 9223372036854775807n ? "int64" : "bigint";
}
function formatFloat(value) {
    if (Number.isNaN(value))
        return "NaN";
    if (value === Infinity)
        return "+Inf";
    if (value === -Infinity)
        return "-Inf";
    return Number.isInteger(value) ? `${value}.0` : String(value);
}
function isRuntimeObject(value) {
    return Boolean(value &&
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
        !isGoJuniorFunction(value));
}
function isRuntimeCallable(value) {
    return Boolean(value && typeof value === "object" && "kind" in value && value.kind === "HostCallable");
}
function isGoJuniorFunction(value) {
    return Boolean(value && typeof value === "object" && "kind" in value && value.kind === "GoJuniorFunction");
}
function normalizeCell(cell) {
    return cell.replace(/\$/g, "").toUpperCase();
}
function dependencyKey(dependency) {
    if (dependency.kind === "cell")
        return `cell:${dependency.sheet}:${dependency.cell}`;
    return `range:${dependency.sheet}:${dependency.start}:${dependency.end}`;
}
function parseCellAddress(cell) {
    const normalized = normalizeCell(cell);
    const match = /^([A-Z]+)([1-9][0-9]*)$/.exec(normalized);
    if (!match)
        throw new GoJuniorRuntimeError(`${cell} is not a spreadsheet cell address`);
    return {
        col: columnNameToNumber(match[1] ?? "A"),
        row: Number(match[2] ?? "1")
    };
}
function columnNameToNumber(name) {
    let value = 0;
    for (const char of name) {
        value = value * 26 + (char.charCodeAt(0) - 64);
    }
    return value;
}
function formatCellAddress(col, row) {
    let name = "";
    let value = col;
    while (value > 0) {
        const remainder = (value - 1) % 26;
        name = String.fromCharCode(65 + remainder) + name;
        value = Math.floor((value - 1) / 26);
    }
    return `${name}${row}`;
}
