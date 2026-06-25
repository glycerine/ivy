import { REPL_FILENAME } from "./diagnostics.js";
import { blake3RawBytes } from "./blake3.js";
import { checkGoJuniorSourceFiles, GOJR_SYNTHETIC_CHECK_PREFIX, isGoJuniorSyntheticCheckName, standardTypePackage } from "./typecheck.js";
import { Const as GoTypesConst, Var as GoTypesVar, ensureUniverseInitialized, NewPackage, NewPkgName, NoPos, RelativeTo as GoTypesRelativeTo, TypeString as GoTypesTypeString } from "./go/types/index.js";
import { frontSourceFilesToAst, frontSourceToAst } from "./frontToAst.js";
import { isHostResolvedSourceImport } from "./intrinsicPackages.js";
import { DeterministicPrng } from "./prng.js";
import { AsyncGoChannel, AsyncGoDeadlockError, AsyncGoPanic, AsyncGoScheduler, asyncSelect } from "./asyncRuntime.js";
import { cellDependency, rangeDependency } from "./spreadsheet.js";
const runtimeHashEncoder = new TextEncoder();
export class RuntimeGoString {
    bytes;
    constructor(bytes) {
        this.bytes = bytes;
    }
    static fromUtf8Text(text) {
        return new RuntimeGoString(new TextEncoder().encode(text));
    }
    static fromBytes(bytes) {
        return new RuntimeGoString(Uint8Array.from(bytes));
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
            trueAliases: new Map(),
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
    testingVerbose() {
        return this.options.testVerbose ?? false;
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
        const bindingScope = scope.bindingScope(name) ?? scope;
        return new RuntimePointer(typeName, () => scope.lookup(name), (next) => scope.assign(name, next, this.assignmentChecker()), `binding:${objectIdentityId(bindingScope)}:${name}`);
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
    argv(defaultName = "gojr") {
        const argv = this.options.argv;
        return argv && argv.length > 0 ? argv.map(String) : [defaultName];
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
    packageName() {
        return this.options.packageName;
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
        const target = this.resolveImportedTypeText(spec.type.text);
        this.shared.aliases.set(spec.name, target);
        if (spec.alias)
            this.shared.trueAliases.set(spec.name, target);
        if (spec.structFields) {
            this.shared.types.set(spec.name, {
                name: spec.name,
                ...(spec.typeParameters ? { typeParameters: spec.typeParameters } : {}),
                declaringContext: this,
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
        const localName = this.options.importPath
            ? unqualifyLocalRuntimeTypeName(name, this.options.importPath)
            : name;
        const local = this.shared.types.get(name) ??
            this.shared.types.get(genericBaseTypeName(name)) ??
            this.shared.types.get(localName) ??
            this.shared.types.get(genericBaseTypeName(localName));
        if (local)
            return local;
        const qualified = packageQualifiedRuntimeTypeNameParts(name);
        const packageContext = qualified ? this.options.packageContexts?.[qualified.importPath] : undefined;
        const foreign = packageContext?.shared.types.get(qualified.localName) ??
            packageContext?.shared.types.get(qualified.baseName);
        return foreign && qualified ? qualifyRuntimeStructTypeDef(foreign, qualified.importPath) : undefined;
    }
    interfaceDef(name) {
        const localName = this.options.importPath
            ? unqualifyLocalRuntimeTypeName(name, this.options.importPath)
            : name;
        const local = this.shared.interfaces.get(name) ??
            this.shared.interfaces.get(genericBaseTypeName(name)) ??
            this.shared.interfaces.get(localName) ??
            this.shared.interfaces.get(genericBaseTypeName(localName));
        if (local)
            return local;
        const qualified = packageQualifiedRuntimeTypeNameParts(name);
        const packageContext = qualified ? this.options.packageContexts?.[qualified.importPath] : undefined;
        const foreign = packageContext?.shared.interfaces.get(qualified.localName) ??
            packageContext?.shared.interfaces.get(qualified.baseName);
        return foreign && qualified ? qualifyRuntimeInterfaceDef(foreign, qualified.importPath) : undefined;
    }
    aliasType(name) {
        const local = this.currentScope.lookupTypeAlias(name) ??
            this.currentScope.lookupTypeAlias(genericBaseTypeName(name)) ??
            this.shared.aliases.get(name) ??
            this.shared.aliases.get(genericBaseTypeName(name));
        if (local)
            return local;
        const qualified = packageQualifiedRuntimeTypeNameParts(name);
        const packageContext = qualified ? this.options.packageContexts?.[qualified.importPath] : undefined;
        const foreign = packageContext?.shared.aliases.get(qualified.localName) ??
            packageContext?.shared.aliases.get(qualified.baseName);
        return foreign && qualified ? qualifyLocalRuntimeTypeName(foreign, qualified.importPath) : undefined;
    }
    trueAliasType(name) {
        return this.shared.trueAliases.get(name) ??
            this.shared.trueAliases.get(genericBaseTypeName(name));
    }
    scopedAliasType(name) {
        return this.currentScope.lookupTypeAlias(name) ?? this.currentScope.lookupTypeAlias(genericBaseTypeName(name));
    }
    registerErasedTypeParameter(name) {
        if (!name || name === "_")
            return;
        this.currentScope.declareTypeAlias(name, name);
    }
    declareTypeAlias(name, target) {
        if (!name || name === "_")
            return;
        this.currentScope.declareTypeAlias(name, target);
    }
    isKnownType(name) {
        const type = normalizeTypeText(name);
        const resolved = normalizeTypeText(this.resolveImportedTypeText(type));
        const candidates = resolved === type ? [type] : [type, resolved];
        return candidates.some((candidate) => {
            const genericBase = genericBaseTypeName(candidate);
            return isPredeclaredType(candidate) ||
                isUnsafePointerType(candidate) ||
                this.currentScope.lookupTypeAlias(candidate) !== undefined ||
                this.shared.aliases.has(candidate) ||
                this.shared.types.has(candidate) ||
                this.shared.interfaces.has(candidate) ||
                this.currentScope.lookupTypeAlias(genericBase) !== undefined ||
                this.shared.aliases.has(genericBase) ||
                this.shared.types.has(genericBase) ||
                this.shared.interfaces.has(genericBase) ||
                candidate.startsWith("*") ||
                candidate.startsWith("[]") ||
                /^\[[0-9.]*\]/.test(candidate) ||
                candidate.startsWith("map[") ||
                parseChanTypeText(candidate) !== undefined ||
                candidate.startsWith("func(");
        });
    }
    registerMethod(declaration, intrinsic) {
        if (!declaration.receiver)
            return;
        const resolved = resolveRuntimeFunctionDeclarationTypeImports(declaration, this);
        const receiver = normalizeReceiverType(resolved.receiver?.type.text ?? declaration.receiver.type.text);
        const receiverBaseType = genericBaseTypeName(receiver.baseType);
        this.shared.methods.set(methodKey(receiverBaseType, resolved.name), {
            declaration: resolved,
            receiverType: receiverBaseType,
            pointerReceiver: receiver.pointer,
            closureScope: this.captureScope(),
            closureContext: this,
            ...(intrinsic ? { intrinsic } : {})
        });
    }
    importPackageMethods(importPath, source) {
        for (const method of source.shared.methods.values()) {
            const receiverType = qualifyLocalRuntimeTypeName(method.receiverType, importPath);
            this.shared.methods.set(methodKey(receiverType, method.declaration.name), {
                declaration: method.declaration,
                receiverType,
                pointerReceiver: method.pointerReceiver,
                closureScope: method.closureScope,
                closureContext: method.closureContext,
                ...(method.intrinsic ? { intrinsic: method.intrinsic } : {})
            });
        }
    }
    importPackageDefinitions(importPath, source) {
        for (const [name, target] of source.shared.aliases.entries()) {
            this.shared.aliases.set(qualifyLocalRuntimeTypeName(name, importPath), qualifyLocalRuntimeTypeName(target, importPath));
        }
        for (const [name, target] of source.shared.trueAliases.entries()) {
            this.shared.trueAliases.set(qualifyLocalRuntimeTypeName(name, importPath), qualifyLocalRuntimeTypeName(target, importPath));
        }
        for (const [name, typeDef] of source.shared.types.entries()) {
            this.shared.types.set(qualifyLocalRuntimeTypeName(name, importPath), qualifyRuntimeStructTypeDef(typeDef, importPath));
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
        const candidates = [];
        const add = (candidate) => {
            const type = normalizeTypeText(candidate ?? "");
            if (type && !candidates.includes(type))
                candidates.push(type);
        };
        const addType = (candidate) => {
            add(candidate);
            add(genericBaseTypeName(candidate));
            if (this.options.importPath) {
                const local = unqualifyLocalRuntimeTypeName(candidate, this.options.importPath);
                add(local);
                add(genericBaseTypeName(local));
            }
        };
        addType(typeName);
        for (const aliasType of runtimeTrueAliasTypeChain(typeName, this))
            addType(aliasType);
        for (const candidate of candidates) {
            const method = this.shared.methods.get(methodKey(candidate, methodName));
            if (method)
                return method;
            const qualified = packageQualifiedRuntimeTypeNameParts(candidate);
            const packageContext = qualified ? this.options.packageContexts?.[qualified.importPath] : undefined;
            const foreign = packageContext?.shared.methods.get(methodKey(qualified.localName, methodName)) ??
                packageContext?.shared.methods.get(methodKey(qualified.baseName, methodName));
            if (foreign && qualified) {
                return {
                    ...foreign,
                    receiverType: qualifyLocalRuntimeTypeName(foreign.receiverType, qualified.importPath)
                };
            }
        }
        return undefined;
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
    typeAliases = new Map();
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
    bindingScope(name) {
        if (this.bindings.has(name))
            return this;
        return this.parent?.bindingScope(name);
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
    declareTypeAlias(name, target) {
        this.typeAliases.set(name, target);
    }
    lookupTypeAlias(name) {
        return this.typeAliases.get(name) ?? this.parent?.lookupTypeAlias(name);
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
        const clone = new RuntimeStruct(this.typeName, this.fields.entries());
        copyRuntimeStructMetadata(this, clone);
        return clone;
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
        this.source.set(field, internalizePackageRuntimeValue(value, this.packageName));
    }
    clone() {
        const clone = new RuntimeStruct(this.typeName, this.orderedFields());
        copyRuntimeStructMetadata(this.source, clone);
        copyRuntimeStructMetadata(this, clone);
        return clone;
    }
    orderedFields() {
        return this.source.orderedFields().map(([name, value]) => [
            name,
            externalizePackageRuntimeValue(value, this.packageName)
        ]);
    }
}
class RuntimeInternalizedStruct extends RuntimeStruct {
    source;
    packageName;
    constructor(source, packageName) {
        super(unqualifyLocalRuntimeTypeName(source.typeName, packageName));
        this.source = source;
        this.packageName = packageName;
    }
    get(field) {
        const value = this.source.get(field);
        return value === undefined ? undefined : internalizePackageRuntimeValue(value, this.packageName);
    }
    set(field, value) {
        this.source.set(field, externalizePackageRuntimeValue(value, this.packageName));
    }
    clone() {
        const clone = new RuntimeStruct(this.typeName, this.orderedFields());
        copyRuntimeStructMetadata(this.source, clone);
        copyRuntimeStructMetadata(this, clone);
        return clone;
    }
    orderedFields() {
        return this.source.orderedFields().map(([name, value]) => [
            name,
            internalizePackageRuntimeValue(value, this.packageName)
        ]);
    }
}
export class RuntimePointer {
    typeName;
    getValue;
    setValue;
    stableIdentity;
    sequence;
    constructor(typeName, getValue, setValue, stableIdentity, sequence) {
        this.typeName = typeName;
        this.getValue = getValue;
        this.setValue = setValue;
        this.stableIdentity = stableIdentity;
        this.sequence = sequence;
    }
    get() {
        return this.getValue();
    }
    set(value) {
        this.setValue(value);
    }
    identityKey() {
        return this.stableIdentity;
    }
    sequenceInfo() {
        return this.sequence;
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
const sparseArrayDefaultElementTypes = new WeakMap();
const tupleValues = new WeakSet();
const runtimePackageInspections = new WeakMap();
const reflectValueInfos = new WeakMap();
const runtimeUintptrIds = new Map();
let nextObjectMapKeyId = 1;
let nextNonReflexiveMapKeyId = 1;
let nextRuntimeUintptrId = 1n;
const SPARSE_ZERO_ARRAY_LENGTH = 1_000_000;
function markTupleValues(values) {
    tupleValues.add(values);
    return values;
}
function isTupleValues(value) {
    return Array.isArray(value) && tupleValues.has(value);
}
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
class RuntimeExternalizedMap extends RuntimeMap {
    source;
    packageName;
    constructor(source, packageName) {
        super(qualifyLocalRuntimeTypeName(source.keyType, packageName), qualifyLocalRuntimeTypeName(source.valueType, packageName));
        this.source = source;
        this.packageName = packageName;
    }
    get(key) {
        return externalizePackageRuntimeValue(this.source.get(internalizePackageRuntimeValue(key, this.packageName)), this.packageName);
    }
    getWithPresence(key) {
        const [value, ok] = this.source.getWithPresence(internalizePackageRuntimeValue(key, this.packageName));
        return [externalizePackageRuntimeValue(value, this.packageName), ok];
    }
    set(key, value) {
        this.source.set(internalizePackageRuntimeValue(key, this.packageName), internalizePackageRuntimeValue(value, this.packageName));
    }
    delete(key) {
        this.source.delete(internalizePackageRuntimeValue(key, this.packageName));
    }
    clear() {
        this.source.clear();
    }
    orderedEntries() {
        return this.source.orderedEntries().map(([key, value]) => [
            externalizePackageRuntimeValue(key, this.packageName),
            externalizePackageRuntimeValue(value, this.packageName)
        ]);
    }
    size() {
        return this.source.size();
    }
}
class RuntimeInternalizedMap extends RuntimeMap {
    source;
    packageName;
    constructor(source, packageName) {
        super(unqualifyLocalRuntimeTypeName(source.keyType, packageName), unqualifyLocalRuntimeTypeName(source.valueType, packageName));
        this.source = source;
        this.packageName = packageName;
    }
    get(key) {
        return internalizePackageRuntimeValue(this.source.get(externalizePackageRuntimeValue(key, this.packageName)), this.packageName);
    }
    getWithPresence(key) {
        const [value, ok] = this.source.getWithPresence(externalizePackageRuntimeValue(key, this.packageName));
        return [internalizePackageRuntimeValue(value, this.packageName), ok];
    }
    set(key, value) {
        this.source.set(externalizePackageRuntimeValue(key, this.packageName), externalizePackageRuntimeValue(value, this.packageName));
    }
    delete(key) {
        this.source.delete(externalizePackageRuntimeValue(key, this.packageName));
    }
    clear() {
        this.source.clear();
    }
    orderedEntries() {
        return this.source.orderedEntries().map(([key, value]) => [
            internalizePackageRuntimeValue(key, this.packageName),
            internalizePackageRuntimeValue(value, this.packageName)
        ]);
    }
    size() {
        return this.source.size();
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
        packagePath: options.importPath ?? packageName,
        autoImportFmt: false
    });
    if (checked.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return {
            diagnostics: checked.diagnostics,
            output: []
        };
    }
    const context = new EvaluationContext(options);
    if (options.importPath && options.packageContexts) {
        options.packageContexts[options.importPath] = context;
    }
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
            predeclarePackageVariables(declarations, checked.pkg, context);
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
        this.progress({
            action: "initializing",
            importPath,
            fileCount: spec.files.length,
            dependencyCount: this.importsByPath.get(importPath)?.length ?? 0
        });
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
        this.progress({
            action: "initialized",
            importPath,
            fileCount: spec.files.length,
            dependencyCount: this.importsByPath.get(importPath)?.length ?? 0
        });
    }
    progress(event) {
        this.options.onProgress?.(event);
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
        installMainProgramArgs(graph, options, importPath);
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
    const testOptions = { testVerbose: true, ...options };
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
    const checked = checkGoJuniorSourceFiles(files, typeCheckConfig(testOptions));
    if (checked.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return {
            diagnostics: checked.diagnostics,
            output: [],
            ast
        };
    }
    return testProgram(ast, checked.diagnostics, testOptions);
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
function installMainProgramArgs(graph, options, mainImportPath) {
    const args = runtimeStringSlice(runtimeArgv(options, mainImportPath));
    const osContext = graph.packageContexts.os;
    if (osContext)
        osContext.declareOrAssignRoot("Args", args, true, "[]string");
    const osPackageObject = graph.packages.os;
    if (osPackageObject)
        osPackageObject.Args = args;
}
function runtimeArgv(options, defaultName = "gojr") {
    return options.argv && options.argv.length > 0 ? options.argv.map(String) : [defaultName];
}
function runtimeStringSlice(values) {
    const slice = values.map((value) => RuntimeGoString.fromUtf8Text(value));
    markArrayType(slice, "[]string");
    return slice;
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
                const testFailed = await runOneTest(declaration, context, diagnostics, options.testVerbose ?? false);
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
async function runOneTest(declaration, context, diagnostics, verbose) {
    if (verbose)
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
    const state = testingT?.state;
    if (testingT && (verbose || runtimeFailure || state?.failed || state?.skipped))
        emitTestingLogs(context, testingT.state);
    if (runtimeFailure && !testingT) {
        context.write(indentTestingLog(runtimeFailure));
    }
    if (runtimeFailure || state?.failed) {
        context.write(`--- FAIL: ${declaration.name}\n`);
        diagnostics.push(testDiagnostic(declaration, runtimeFailure ? `${declaration.name}: ${runtimeFailure}` : `${declaration.name} failed`));
        return true;
    }
    if (state?.skipped) {
        context.write(`--- SKIP: ${declaration.name}\n`);
        return false;
    }
    if (verbose)
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
            replMode: this.options.replMode ?? true,
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
        ...(options.replMode ? { allowPackageInspection: true } : {}),
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
        markRuntimePackageObject(pkg, imported.path, context.packageInfo(imported.path) ?? standardTypePackage(imported.path));
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
        context.registerMethod(resolved, packageMethodIntrinsic(resolved, context.importPath()));
        return;
    }
    const intrinsic = packageFunctionIntrinsic(resolved, context.importPath()) ??
        bodylessPackageFunctionIntrinsic(resolved, context.importPath());
    context.declareOrAssignRoot(resolved.name, intrinsic ?? goJuniorFunctionValue(resolved.name, resolved.signature, resolved.body, context.captureScope(), resolved, undefined, undefined, context), true);
}
function packageFunctionIntrinsic(declaration, importPath) {
    let intrinsic;
    if (importPath === "reflect" && declaration.name === "ValueOf") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => reflectValueOf(args[0] ?? null, context));
    }
    if (importPath === "reflect" && (declaration.name === "PointerTo" || declaration.name === "PtrTo")) {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => {
            const sourceType = reflectRtypeRuntimeTypeText(args[0] ?? null);
            return sourceType ? reflectRtypePointerForTypeText(`*${sourceType}`, context) : null;
        });
    }
    if (importPath === "internal/abi" && declaration.name === "TypeOf") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => internalAbiTypeOf(args[0] ?? null, context));
    }
    if (importPath === "internal/abi" && declaration.name === "TypeFor") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (_args, context, typeArguments) => internalAbiTypeDescriptor(typeArguments?.[0] ?? "interface{}", context));
    }
    if (importPath === "internal/abi" && declaration.name === "NoEscape") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => args[0] ?? null);
    }
    if (importPath === "internal/abi" && declaration.name === "EscapeNonString") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, () => null);
    }
    if (importPath === "internal/abi" && declaration.name === "EscapeToResultNonString") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => args[0] ?? null);
    }
    if (importPath === "crypto/internal/constanttime" && declaration.name === "boolToUint8") {
        intrinsic = intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => toBool(args[0] ?? false) ? 1n : 0n);
    }
    return intrinsic
        ? {
            ...intrinsic,
            ...(declaration.source ? { source: declaration.source } : {}),
            declaration
        }
        : undefined;
}
function packageMethodIntrinsic(declaration, importPath) {
    if (!declaration.receiver)
        return undefined;
    const receiver = normalizeReceiverType(declaration.receiver.type.text);
    if (importPath === "reflect" && (receiver.baseType === "Value" || receiver.baseType === "reflect.Value")) {
        switch (declaration.name) {
            case "Kind":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueKind(args[0] ?? null));
            case "IsValid":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueIsValid(args[0] ?? null));
            case "Type":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => reflectValueType(args[0] ?? null, context));
            case "Pointer":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValuePointer(args[0] ?? null));
            case "Elem":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => reflectValueElem(args[0] ?? null, context));
            case "Interface":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueInterface(args[0] ?? null));
            case "IsNil":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueIsNil(args[0] ?? null));
            case "Len":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueLen(args[0] ?? null));
            case "CanAddr":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => Boolean(reflectValueInfo(args[0] ?? null)?.set));
            case "Set":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => reflectValueSet(args[0] ?? null, args[1] ?? null));
            default:
                return undefined;
        }
    }
    if (importPath === "reflect" && (receiver.baseType === "rtype" || receiver.baseType === "reflect.rtype")) {
        switch (declaration.name) {
            case "Name":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => RuntimeGoString.fromUtf8Text(reflectRtypeDescriptorText(args[0] ?? null, "GoJrName")));
            case "String":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => RuntimeGoString.fromUtf8Text(reflectRtypeDescriptorText(args[0] ?? null, "GoJrString")));
            case "PkgPath":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => RuntimeGoString.fromUtf8Text(reflectRtypeDescriptorText(args[0] ?? null, "GoJrPkgPath")));
            case "Implements":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args, context) => reflectRtypeImplements(args[0] ?? null, args[1] ?? null, context));
            default:
                return undefined;
        }
    }
    if (importPath !== "internal/abi")
        return undefined;
    if (receiver.baseType === "Type" || receiver.baseType === "internal/abi.Type") {
        switch (declaration.name) {
            case "HasName":
                return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => (internalAbiTypeFlag(args[0] ?? null) & internalAbiTFlagNamed) !== 0n);
            default:
                return undefined;
        }
    }
    if (receiver.baseType !== "Name" && receiver.baseType !== "internal/abi.Name")
        return undefined;
    switch (declaration.name) {
        case "Name":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => RuntimeGoString.fromUtf8Text(internalAbiNameText(args[0] ?? null)));
        case "Tag":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => RuntimeGoString.fromUtf8Text(internalAbiNameTag(args[0] ?? null)));
        case "IsExported":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => internalAbiNameBool(args[0] ?? null, "Exported"));
        case "HasTag":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => internalAbiNameTag(args[0] ?? null) !== "");
        case "IsEmbedded":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => internalAbiNameBool(args[0] ?? null, "Embedded"));
        case "IsBlank":
            return intrinsicGoJuniorFunction(declaration.name, declaration.signature, (args) => internalAbiNameText(args[0] ?? null) === "_");
        default:
            return undefined;
    }
}
function intrinsicGoJuniorFunction(name, signature, call) {
    return {
        kind: "GoJuniorFunction",
        name,
        signature,
        async call(args, context, typeArguments) {
            return call(args, context, typeArguments);
        }
    };
}
function bodylessPackageFunctionIntrinsic(declaration, importPath) {
    if (!isBodylessFunctionDeclaration(declaration))
        return undefined;
    const intrinsic = bodylessBytealgIntrinsic(importPath, declaration.name, declaration.signature) ??
        bodylessAtomicIntrinsic(importPath, declaration.name, declaration.signature) ??
        bodylessAbiIntrinsic(importPath, declaration.name, declaration.signature) ??
        bodylessIterCoroutineIntrinsic(importPath, declaration.name, declaration.signature) ??
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
function bodylessIterCoroutineIntrinsic(importPath, name, signature) {
    if (importPath !== "iter")
        return undefined;
    if (name !== "newcoro" && name !== "coroswitch")
        return undefined;
    return intrinsicGoJuniorFunction(`${importPath}.${name}`, signature, () => {
        throw new GoJuniorPanic(`gojr error: ${importPath}.${name} not implemented`);
    });
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
const internalAbiTypeDescriptorCache = new Map();
const internalAbiTFlagNamed = 1n << 2n;
function internalAbiTypeOf(value, context) {
    if (value instanceof RuntimeInterfaceValue) {
        if (value.value === null)
            return null;
        value = value.value;
    }
    if (value === null)
        return null;
    return internalAbiTypeDescriptor(reflectliteTypeTextOfValue(value), context);
}
function isInternalAbiTypePointer(value, context) {
    const type = normalizeTypeText(value.typeName);
    return type === "internal/abi.Type" || (type === "Type" && context.importPath() === "internal/abi");
}
function internalAbiTypeDescriptor(typeText, context) {
    const type = canonicalRuntimeTypeName(normalizeTypeText(resolveRuntimeCompositeAliases(internalAbiCacheTypeText(typeText, context), context)));
    const names = internalAbiRuntimeTypeNames(context);
    const cacheKey = `${Object.values(names).join(":")}:${type}`;
    const cached = internalAbiTypeDescriptorCache.get(cacheKey);
    if (cached)
        return cached;
    const layout = internalAbiDescriptorStruct(type, names, context);
    const pointer = new RuntimePointer(names.type, () => layout, (next) => {
        const struct = structFromValue(next);
        if (!struct)
            throwTypeError(next, names.type, `*${names.type}`);
        for (const [field, value] of struct.orderedFields())
            layout.set(field, value);
    }, `internal/abi.Type:${cacheKey}`);
    internalAbiTypeDescriptorCache.set(cacheKey, pointer);
    return pointer;
}
function internalAbiCacheTypeText(typeText, context) {
    const resolved = context.resolveImportedTypeText(typeText);
    const importPath = context.importPath();
    return importPath ? qualifyLocalRuntimeTypeName(resolved, importPath) : resolved;
}
function internalAbiRuntimeTypeNames(context) {
    return {
        type: internalAbiRuntimeTypeName("Type", context),
        arrayType: internalAbiRuntimeTypeName("ArrayType", context),
        chanType: internalAbiRuntimeTypeName("ChanType", context),
        mapType: internalAbiRuntimeTypeName("MapType", context),
        name: internalAbiRuntimeTypeName("Name", context),
        ptrType: internalAbiRuntimeTypeName("PtrType", context),
        sliceType: internalAbiRuntimeTypeName("SliceType", context),
        structField: internalAbiRuntimeTypeName("StructField", context),
        structType: internalAbiRuntimeTypeName("StructType", context)
    };
}
function internalAbiDescriptorStruct(typeText, names, context) {
    const underlyingType = internalAbiUnderlyingTypeText(typeText, context);
    const arrayType = parseArrayOrSliceTypeText(underlyingType, context);
    if (arrayType) {
        return arrayType.length === undefined || arrayType.inferLength
            ? internalAbiSliceTypeStruct(typeText, arrayType.elementType, names, context)
            : internalAbiArrayTypeStruct(typeText, arrayType.elementType, arrayType.length, names, context);
    }
    const mapType = parseMapTypeText(underlyingType);
    if (mapType)
        return internalAbiMapTypeStruct(typeText, mapType.keyType, mapType.valueType, names, context);
    const chanType = parseChanTypeText(underlyingType);
    if (chanType)
        return internalAbiChanTypeStruct(typeText, chanType.elementType, chanType.direction, names, context);
    if (underlyingType.startsWith("*"))
        return internalAbiPtrTypeStruct(typeText, underlyingType.slice(1), names, context);
    const structType = context.typeDef(typeText) ?? parseAnonymousStructTypeText(typeText);
    if (structType)
        return internalAbiStructTypeStruct(typeText, structType, names, context);
    return internalAbiTypeStruct(typeText, names.type, context);
}
function internalAbiArrayTypeStruct(typeText, elementType, length, names, context) {
    const struct = new RuntimeStruct(names.arrayType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    struct.set("Elem", internalAbiTypeDescriptor(elementType, context));
    struct.set("Slice", internalAbiTypeDescriptor(`[]${elementType}`, context));
    struct.set("Len", BigInt(length));
    return struct;
}
function internalAbiChanTypeStruct(typeText, elementType, direction, names, context) {
    const struct = new RuntimeStruct(names.chanType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    struct.set("Elem", internalAbiTypeDescriptor(elementType, context));
    struct.set("Dir", direction === "receive" ? 1n : direction === "send" ? 2n : 3n);
    return struct;
}
function internalAbiMapTypeStruct(typeText, keyType, valueType, names, context) {
    const struct = new RuntimeStruct(names.mapType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    const keySize = internalAbiSizeForType(keyType, context);
    const valueSize = internalAbiSizeForType(valueType, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    struct.set("Key", internalAbiTypeDescriptor(keyType, context));
    struct.set("Elem", internalAbiTypeDescriptor(valueType, context));
    struct.set("Group", internalAbiTypeDescriptor("struct{}", context));
    struct.set("Hasher", hostCallable("internal/abi.MapType.Hasher", (args) => BigInt(internalAbiTypeHash(`${keyType}:${formatValue(args[0] ?? null)}:${String(args[1] ?? 0n)}`))));
    struct.set("GroupSize", 0n);
    struct.set("KeysOff", 0n);
    struct.set("KeyStride", keySize);
    struct.set("ElemsOff", 0n);
    struct.set("ElemStride", valueSize);
    struct.set("ElemOff", 0n);
    struct.set("Flags", (keySize > 128n ? 4n : 0n) | (valueSize > 128n ? 8n : 0n));
    return struct;
}
function internalAbiPtrTypeStruct(typeText, elementType, names, context) {
    const struct = new RuntimeStruct(names.ptrType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    struct.set("Elem", internalAbiTypeDescriptor(elementType, context));
    return struct;
}
function internalAbiSliceTypeStruct(typeText, elementType, names, context) {
    const struct = new RuntimeStruct(names.sliceType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    struct.set("Elem", internalAbiTypeDescriptor(elementType, context));
    return struct;
}
function internalAbiStructTypeStruct(typeText, typeDef, names, context) {
    const struct = new RuntimeStruct(names.structType);
    const base = internalAbiTypeStruct(typeText, names.type, context);
    internalAbiSetEmbeddedTypeFields(struct, base);
    const ownerPackage = runtimeQualifiedTypePackagePath(typeText) ?? runtimeQualifiedTypePackagePath(typeDef.name);
    struct.set("PkgPath", internalAbiNameStruct(ownerPackage ?? "", "", false, false, names));
    const fields = [];
    let offset = 0n;
    for (const field of instantiateStructFields(typeDef, typeText)) {
        const resolvedField = ownerPackage ? {
            ...field,
            type: {
                ...field.type,
                text: qualifyLocalRuntimeTypeName(field.type.text, ownerPackage)
            }
        } : field;
        const fieldSize = internalAbiSizeForType(resolvedField.type.text, context);
        fields.push(internalAbiStructFieldStruct(resolvedField, offset, names, context));
        offset += fieldSize;
    }
    markArrayType(fields, `[]${names.structField}`);
    struct.set("Fields", fields);
    return struct;
}
function internalAbiSetEmbeddedTypeFields(struct, base) {
    struct.set("Type", base);
    for (const [field, value] of base.orderedFields()) {
        struct.set(field, value);
    }
}
function internalAbiStructFieldStruct(field, offset, names, context) {
    return new RuntimeStruct(names.structField, [
        ["Name", internalAbiNameStruct(field.name, field.tag ?? "", isExportedRuntimeName(field.name), Boolean(field.embedded), names)],
        ["Typ", internalAbiTypeDescriptor(field.type.text, context)],
        ["Offset", offset]
    ]);
}
function internalAbiNameStruct(name, tag, exported, embedded, names) {
    return new RuntimeStruct(names.name, [
        ["Bytes", new RuntimeTypedNilValue("*byte")],
        ["Text", RuntimeGoString.fromUtf8Text(name)],
        ["TagText", RuntimeGoString.fromUtf8Text(tag)],
        ["Exported", exported],
        ["Embedded", embedded]
    ]);
}
function internalAbiNameText(value) {
    const struct = structFromValue(value);
    const text = struct?.get("Text");
    return text !== undefined && isRuntimeString(text) ? goStringText(text) : "";
}
function internalAbiNameTag(value) {
    const struct = structFromValue(value);
    const tag = struct?.get("TagText");
    return tag !== undefined && isRuntimeString(tag) ? goStringText(tag) : "";
}
function internalAbiNameBool(value, field) {
    const struct = structFromValue(value);
    return struct?.get(field) === true;
}
function internalAbiTypeStruct(typeText, typeName, context) {
    const display = internalAbiTypeDisplayInfo(typeText, context);
    const kind = internalAbiKindForType(typeText, context);
    const size = internalAbiSizeForType(typeText, context);
    const ptrBytes = typeText.startsWith("*") || typeText.startsWith("map[") || typeText.startsWith("chan ") || typeText.startsWith("<-chan") || typeText.startsWith("chan<-") || typeText.startsWith("func(")
        ? 8n
        : 0n;
    return new RuntimeStruct(typeName, [
        ["Size_", size],
        ["PtrBytes", ptrBytes],
        ["Hash", BigInt(internalAbiTypeHash(typeText))],
        ["TFlag", display.named ? internalAbiTFlagNamed : 0n],
        ["Align_", BigInt(Math.min(8, Math.max(1, Number(size || 1n))))],
        ["FieldAlign_", BigInt(Math.min(8, Math.max(1, Number(size || 1n))))],
        ["Kind_", kind],
        ["Equal", hostCallable("internal/abi.Type.Equal", () => true)],
        ["GCData", new RuntimeTypedNilValue("*byte")],
        ["Str", 0n],
        ["PtrToThis", 0n],
        ["GoJrTypeText", RuntimeGoString.fromUtf8Text(display.typeText)],
        ["GoJrName", RuntimeGoString.fromUtf8Text(display.name)],
        ["GoJrString", RuntimeGoString.fromUtf8Text(display.string)],
        ["GoJrPkgPath", RuntimeGoString.fromUtf8Text(display.pkgPath)]
    ]);
}
function internalAbiTypeDisplayInfo(typeText, context) {
    const type = normalizeTypeText(typeText);
    const canonical = canonicalRuntimeTypeName(type);
    if (canonical === "any" || canonical === "interface{}") {
        return {
            typeText: canonical,
            name: "",
            string: "interface{}",
            pkgPath: "",
            named: false
        };
    }
    if (isPredeclaredType(canonical)) {
        return {
            typeText: canonical,
            name: canonical,
            string: canonical,
            pkgPath: "",
            named: true
        };
    }
    const typeDef = context.typeDef(type);
    const interfaceDef = context.interfaceDef(type);
    const namedAliasTarget = internalAbiNamedAliasTarget(type, context);
    const named = Boolean(typeDef || interfaceDef || namedAliasTarget);
    const namedType = typeDef?.name ?? interfaceDef?.name ?? type;
    const pkgPath = named
        ? runtimeQualifiedTypePackagePath(namedType) ?? runtimeQualifiedTypePackagePath(type) ?? context.importPath() ?? ""
        : "";
    const name = named ? runtimeDisplayLocalTypeName(namedType, context) : "";
    return {
        typeText: type,
        name,
        string: runtimeDisplayTypeString(type, context),
        pkgPath,
        named
    };
}
function runtimeDisplayTypeString(typeText, context) {
    const type = normalizeTypeText(typeText);
    const canonical = canonicalRuntimeTypeName(type);
    if (canonical === "any" || canonical === "interface{}")
        return "interface{}";
    if (isPredeclaredType(canonical))
        return canonical;
    if (type.startsWith("*"))
        return `*${runtimeDisplayTypeString(type.slice(1), context)}`;
    if (type.startsWith("[]"))
        return `[]${runtimeDisplayTypeString(type.slice(2), context)}`;
    const arrayType = parseArrayOrSliceTypeText(type, context);
    if (arrayType?.length !== undefined && !arrayType.inferLength) {
        return `[${arrayType.length}]${runtimeDisplayTypeString(arrayType.elementType, context)}`;
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${runtimeDisplayTypeString(mapType.keyType, context)}]${runtimeDisplayTypeString(mapType.valueType, context)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const prefix = chanType.direction === "receive" ? "<-chan " : chanType.direction === "send" ? "chan<- " : "chan ";
        return `${prefix}${runtimeDisplayTypeString(chanType.elementType, context)}`;
    }
    const generic = genericTypeArguments(type);
    if (generic) {
        return `${runtimeDisplayTypeString(generic.base, context)}[${generic.args.map((arg) => runtimeDisplayTypeString(arg, context)).join(",")}]`;
    }
    const typeDef = context.typeDef(type);
    const interfaceDef = context.interfaceDef(type);
    const namedAliasTarget = internalAbiNamedAliasTarget(type, context);
    if (typeDef || interfaceDef || namedAliasTarget) {
        const namedType = typeDef?.name ?? interfaceDef?.name ?? type;
        const pkgPath = runtimeQualifiedTypePackagePath(namedType) ?? runtimeQualifiedTypePackagePath(type) ?? context.importPath();
        const localName = runtimeDisplayLocalTypeName(namedType, context);
        return pkgPath ? `${runtimePackageDisplayName(pkgPath, context)}.${localName}` : localName;
    }
    return type;
}
function runtimeDisplayLocalTypeName(typeText, context) {
    const type = normalizeTypeText(typeText);
    const generic = genericTypeArguments(type);
    if (generic) {
        return `${runtimeDisplayLocalTypeName(generic.base, context)}[${generic.args.map((arg) => runtimeDisplayTypeString(arg, context)).join(",")}]`;
    }
    const base = genericBaseTypeName(type);
    const dot = base.lastIndexOf(".");
    return dot >= 0 ? base.slice(dot + 1) : base;
}
function runtimePackageDisplayName(importPath, context) {
    return context.packageInfo(importPath)?.Name() ??
        (context.importPath() === importPath ? context.packageName() : undefined) ??
        standardTypePackage(importPath)?.Name() ??
        importDefaultName(importPath);
}
function reflectRtypeDescriptorText(value, field) {
    const descriptor = reflectRtypeDescriptor(value);
    if (!descriptor)
        return "";
    return runtimeDescriptorText(descriptor, field);
}
function reflectRtypeDescriptor(value) {
    const actual = value instanceof RuntimeInterfaceValue ? value.value : value;
    const rtype = structFromValue(actual);
    const descriptor = rtype?.get("t");
    return descriptor instanceof RuntimeStruct ? descriptor : undefined;
}
function reflectRtypeRuntimeTypeText(value) {
    const descriptor = reflectRtypeDescriptor(value);
    return descriptor ? runtimeDescriptorText(descriptor, "GoJrTypeText") : "";
}
function reflectRtypePointerForTypeText(typeText, context) {
    const abiPointer = internalAbiTypeDescriptor(typeText, context);
    return reinterpretInternalAbiTypeAsReflectRtype(abiPointer, "reflect.rtype", context) ??
        new RuntimePointer("reflect.rtype", () => new RuntimeStruct("reflect.rtype", [["t", abiPointer.get()]]), () => {
            throw new GoJuniorRuntimeError("reflect.rtype descriptors are read-only");
        }, `reflect.rtype:${runtimePointerIdentityKey(abiPointer)}`);
}
function reflectRtypeImplements(source, target, context) {
    const sourceType = reflectRtypeRuntimeTypeText(source);
    const targetType = reflectRtypeRuntimeTypeText(target);
    if (!sourceType || !targetType)
        return false;
    const targetInterface = interfaceTarget(targetType, context);
    if (!targetInterface)
        return false;
    return runtimeTypeTextImplementsInterface(sourceType, targetInterface, context);
}
function copyRuntimeStructMetadata(source, target) {
    const reflectInfo = reflectValueInfos.get(source);
    if (reflectInfo)
        reflectValueInfos.set(target, reflectInfo);
}
function reflectTypeOf(value, context) {
    const actual = reflectDynamicValue(value);
    if (actual === null)
        return new RuntimeInterfaceValue(reflectTypeInterfaceName(context), null);
    return reflectTypeInterfaceForTypeText(reflectliteTypeTextOfValue(actual), context);
}
function reflectValueOf(value, context) {
    const actual = reflectDynamicValue(value);
    if (actual === null)
        return reflectZeroValue(context);
    return reflectValueStruct(actual, context);
}
function reflectDynamicValue(value) {
    if (value instanceof RuntimeInterfaceValue) {
        return value.value === null ? null : value.value;
    }
    return value;
}
function reflectValueStruct(value, context, typeText = reflectliteTypeTextOfValue(value), set) {
    const kind = internalAbiKindForType(typeText, context);
    const struct = new RuntimeStruct(reflectValueRuntimeTypeName(context), [
        ["typ_", internalAbiTypeDescriptor(typeText, context)],
        ["ptr", reflectValuePointerPayload(value, typeText, context, set)],
        ["flag", kind]
    ]);
    reflectValueInfos.set(struct, { value, typeText, kind, ...(set ? { set } : {}) });
    return struct;
}
function reflectZeroValue(context) {
    const struct = new RuntimeStruct(reflectValueRuntimeTypeName(context), [
        ["typ_", new RuntimeTypedNilValue("*internal/abi.Type")],
        ["ptr", new RuntimeTypedNilValue("unsafe.Pointer")],
        ["flag", 0n]
    ]);
    reflectValueInfos.set(struct, { value: null, typeText: "", kind: 0n });
    return struct;
}
function reflectValueRuntimeTypeName(context) {
    return context.importPath() === "reflect" ? "Value" : "reflect.Value";
}
function reflectTypeInterfaceName(context) {
    return context.importPath() === "reflect" ? "Type" : "reflect.Type";
}
function reflectTypeInterfaceForTypeText(typeText, context) {
    const interfaceName = reflectTypeInterfaceName(context);
    const pointer = reflectRtypePointerForTypeText(typeText, context);
    return new RuntimeInterfaceValue(interfaceName, pointer);
}
function reflectValuePointerPayload(value, typeText, context, set) {
    const actual = unwrapNamed(value);
    if (actual instanceof RuntimePointer || actual instanceof RuntimeTypedNilValue)
        return actual;
    const pointerType = normalizeTypeText(typeText);
    return new RuntimePointer(pointerType, () => value, (next) => {
        if (!set)
            throw new GoJuniorRuntimeError("reflect.Value data is read-only");
        set(prepareAssignableToType(next, pointerType, "reflect.Value data", context));
    }, `reflect.Value:${runtimeUintptrForValue(value)}`);
}
function reflectValueInfo(value) {
    const struct = structFromValue(value);
    if (!struct)
        return undefined;
    return reflectValueInfos.get(struct) ?? reflectValueInfoFromFields(struct);
}
function reflectValueInfoFromFields(struct) {
    const flag = struct.get("flag");
    const kind = typeof flag === "bigint" ? flag & 31n : 0n;
    const typ = struct.get("typ_");
    const descriptor = typ instanceof RuntimePointer ? structFromValue(typ.get()) : structFromValue(typ ?? null);
    const typeText = descriptor ? runtimeDescriptorText(descriptor, "GoJrTypeText") : "";
    if (kind === 0n && !typeText)
        return { value: null, typeText: "", kind: 0n };
    const ptr = struct.get("ptr");
    return { value: ptr ?? null, typeText, kind };
}
function reflectValueKind(value) {
    return reflectValueInfo(value)?.kind ?? 0n;
}
function reflectValueIsValid(value) {
    return reflectValueKind(value) !== 0n;
}
function reflectValueType(value, context) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n || !info.typeText) {
        throw new GoJuniorPanic("panic: reflect: call of reflect.Value.Type on zero Value");
    }
    return reflectTypeInterfaceForTypeText(info.typeText, context);
}
function reflectValuePointer(value) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n) {
        throw new GoJuniorPanic("panic: &ValueError{Method:reflect.Value.Pointer Kind:0}");
    }
    switch (info.kind) {
        case 18n:
        case 19n:
        case 21n:
        case 22n:
        case 23n:
        case 24n:
        case 26n:
            return runtimeUintptrForValue(info.value);
        default:
            throw new GoJuniorPanic(`panic: &ValueError{Method:reflect.Value.Pointer Kind:${info.kind}}`);
    }
}
function reflectValueElem(value, context) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n) {
        throw new GoJuniorPanic("panic: reflect: call of reflect.Value.Elem on zero Value");
    }
    const actual = unwrapNamed(info.value);
    if (info.kind === 22n) {
        if (actual instanceof RuntimeTypedNilValue)
            return reflectZeroValue(context);
        if (actual instanceof RuntimePointer) {
            return reflectValueStruct(actual.get(), context, reflectElemTypeText(info.typeText, context), (next) => actual.set(next));
        }
    }
    if (info.kind === 20n) {
        const dynamic = reflectDynamicValue(info.value);
        return dynamic === null ? reflectZeroValue(context) : reflectValueStruct(dynamic, context);
    }
    throw new GoJuniorPanic(`panic: &ValueError{Method:reflect.Value.Elem Kind:${info.kind}}`);
}
function reflectValueInterface(value) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n) {
        throw new GoJuniorPanic("panic: reflect: call of reflect.Value.Interface on zero Value");
    }
    return info.value;
}
function reflectValueIsNil(value) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n) {
        throw new GoJuniorPanic("panic: reflect: call of reflect.Value.IsNil on zero Value");
    }
    return reflectliteIsNil(info.value);
}
function reflectValueLen(value) {
    const info = reflectValueInfo(value);
    if (!info || info.kind === 0n) {
        throw new GoJuniorPanic("panic: reflect: call of reflect.Value.Len on zero Value");
    }
    return BigInt(valueLength(info.value));
}
function reflectValueSet(target, source) {
    const info = reflectValueInfo(target);
    if (!info || !info.set)
        throw new GoJuniorPanic("panic: reflect: reflect.Value.Set using unaddressable value");
    info.set(reflectValuePayload(source));
    return null;
}
function reflectValuePayload(value) {
    return reflectValueInfo(value)?.value ?? value;
}
function reflectElemTypeText(typeText, context) {
    const type = internalAbiUnderlyingTypeText(typeText, context);
    if (type.startsWith("*"))
        return type.slice(1);
    if (type.startsWith("[]"))
        return type.slice(2);
    const array = parseArrayOrSliceTypeText(type, context);
    if (array)
        return array.elementType;
    const mapType = parseMapTypeText(type);
    if (mapType)
        return mapType.valueType;
    const chanType = parseChanTypeText(type);
    if (chanType)
        return chanType.elementType;
    return "interface{}";
}
function runtimeUintptrForValue(value) {
    const actual = unwrapNamed(value);
    if (actual === null)
        return 0n;
    if (actual instanceof RuntimeInterfaceValue)
        return actual.value === null ? 0n : runtimeUintptrForValue(actual.value);
    if (actual instanceof RuntimeTypedNilValue)
        return 0n;
    if (actual instanceof RuntimePointer)
        return runtimeUintptrForKey(`ptr:${runtimePointerIdentityKey(actual)}`);
    if (Array.isArray(actual))
        return runtimeUintptrForKey(`array:${objectIdentityId(actual)}`);
    if (actual instanceof RuntimeGoString)
        return actual.byteLength() === 0 ? 0n : runtimeUintptrForKey(`gostring:${objectIdentityId(actual)}`);
    if (actual instanceof RuntimeMap)
        return runtimeUintptrForKey(`map:${objectIdentityId(actual)}`);
    if (actual instanceof RuntimeChannel)
        return runtimeUintptrForKey(`chan:${objectIdentityId(actual)}`);
    if (actual instanceof RuntimeStruct)
        return runtimeUintptrForKey(`struct:${objectIdentityId(actual)}`);
    if (isGoJuniorFunction(actual) || isRuntimeCallable(actual))
        return runtimeUintptrForKey(`func:${objectIdentityId(actual)}`);
    if (typeof actual === "object")
        return runtimeUintptrForKey(`object:${objectIdentityId(actual)}`);
    if (typeof actual === "string")
        return actual.length === 0 ? 0n : runtimeUintptrForKey(`string:${actual}`);
    if (typeof actual === "bigint")
        return BigInt.asUintN(64, actual);
    if (typeof actual === "number")
        return BigInt.asUintN(64, BigInt(Math.trunc(actual)));
    return runtimeUintptrForKey(`primitive:${String(actual)}`);
}
function runtimeUintptrForKey(key) {
    const existing = runtimeUintptrIds.get(key);
    if (existing !== undefined)
        return existing;
    const next = nextRuntimeUintptrId++;
    runtimeUintptrIds.set(key, next);
    return next;
}
function internalAbiTypeFlag(value) {
    const descriptor = structFromValue(value);
    const tflag = descriptor?.get("TFlag");
    return typeof tflag === "bigint" ? tflag : 0n;
}
function runtimeDescriptorText(descriptor, field) {
    const value = descriptor.get(field);
    return value !== undefined && isRuntimeString(value) ? goStringText(value) : "";
}
function internalAbiUnderlyingTypeText(typeText, context, seen = new Set()) {
    const type = normalizeTypeText(typeText);
    if (seen.has(type))
        return type;
    seen.add(type);
    if (context.typeDef(type) || context.interfaceDef(type) || parseAnonymousStructTypeText(type) || parseAnonymousInterfaceTypeText(type))
        return type;
    const alias = context.aliasType(type);
    if (alias && alias !== type)
        return internalAbiUnderlyingTypeText(alias, context, seen);
    return type;
}
function internalAbiNamedAliasTarget(typeText, context) {
    const type = normalizeTypeText(typeText);
    if (context.typeDef(type) || context.interfaceDef(type))
        return undefined;
    const alias = context.aliasType(type);
    if (!alias || alias === type)
        return undefined;
    if (context.trueAliasType(type))
        return undefined;
    return alias;
}
function internalAbiRuntimeTypeName(name, context) {
    if (context.typeDef(name) || context.aliasType(name) || context.interfaceDef(name))
        return name;
    const qualified = `internal/abi.${name}`;
    if (context.typeDef(qualified) || context.aliasType(qualified) || context.interfaceDef(qualified))
        return qualified;
    return context.importPath() === "internal/abi" ? name : qualified;
}
function internalAbiKindForType(typeText, context) {
    const type = context ? internalAbiUnderlyingTypeText(typeText, context) : normalizeTypeText(typeText);
    if (type === "bool")
        return 1n;
    if (type === "int")
        return 2n;
    if (type === "int8")
        return 3n;
    if (type === "int16")
        return 4n;
    if (type === "int32" || type === "rune")
        return 5n;
    if (type === "int64")
        return 6n;
    if (type === "uint")
        return 7n;
    if (type === "uint8" || type === "byte")
        return 8n;
    if (type === "uint16")
        return 9n;
    if (type === "uint32")
        return 10n;
    if (type === "uint64")
        return 11n;
    if (type === "uintptr")
        return 12n;
    if (type === "float32")
        return 13n;
    if (type === "float64")
        return 14n;
    if (type === "complex64")
        return 15n;
    if (type === "complex128")
        return 16n;
    if (type.startsWith("[]"))
        return 23n;
    if (/^\[(?:[0-9.]+|\.\.\.)\]/.test(type))
        return 17n;
    if (parseChanTypeText(type))
        return 18n;
    if (type.startsWith("func("))
        return 19n;
    if (type === "error" || type === "any" || type === "interface{}" || context?.interfaceDef(type) || parseAnonymousInterfaceTypeText(type))
        return 20n;
    if (type.startsWith("map["))
        return 21n;
    if (type.startsWith("*"))
        return 22n;
    if (type === "string")
        return 24n;
    if (context?.typeDef(type) || parseAnonymousStructTypeText(type))
        return 25n;
    if (isUnsafePointerType(type))
        return 26n;
    return 0n;
}
function internalAbiSizeForType(typeText, context) {
    const type = context ? internalAbiUnderlyingTypeText(typeText, context) : normalizeTypeText(typeText);
    if (type === "bool" || type === "int8" || type === "uint8" || type === "byte")
        return 1n;
    if (type === "int16" || type === "uint16")
        return 2n;
    if (type === "int32" || type === "rune" || type === "uint32" || type === "float32")
        return 4n;
    if (type === "complex64")
        return 8n;
    if (type === "complex128")
        return 16n;
    if (type === "string" || type === "error" || type === "any" || type === "interface{}" || context?.interfaceDef(type))
        return 16n;
    if (type.startsWith("[]"))
        return 24n;
    if (type.startsWith("*") || type.startsWith("map[") || parseChanTypeText(type) || type.startsWith("func(") || isUnsafePointerType(type))
        return 8n;
    const array = parseArrayOrSliceTypeText(type, context);
    if (array?.length !== undefined && !array.inferLength)
        return BigInt(array.length) * internalAbiSizeForType(array.elementType, context);
    const structType = context?.typeDef(type) ?? parseAnonymousStructTypeText(type);
    if (structType) {
        return structType.fields.reduce((sum, field) => sum + internalAbiSizeForType(field.type.text, context), 0n);
    }
    return 8n;
}
function internalAbiTypeHash(typeText) {
    const digest = blake3RawBytes(runtimeHashEncoder.encode(typeText), 4);
    return (digest[0] |
        (digest[1] << 8) |
        (digest[2] << 16) |
        (digest[3] << 24)) >>> 0;
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
            return values.length === 0 ? null : values.length === 1 ? values[0] ?? null : markTupleValues(values);
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
function bodylessAtomicIntrinsic(importPath, name, signature) {
    if (importPath !== "sync/atomic")
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
        case "LoadInt32":
            return functionValue((args) => atomicLoadInteger(args, "int32", name));
        case "LoadUint32":
            return functionValue((args) => atomicLoadInteger(args, "uint32", name));
        case "LoadUint64":
            return functionValue((args) => atomicLoadInteger(args, "uint64", name));
        case "LoadUintptr":
            return functionValue((args) => atomicLoadInteger(args, "uintptr", name));
        case "StoreInt32":
            return functionValue((args, context) => atomicStoreInteger(args, "int32", name, context));
        case "StoreUint32":
            return functionValue((args, context) => atomicStoreInteger(args, "uint32", name, context));
        case "StoreUint64":
            return functionValue((args, context) => atomicStoreInteger(args, "uint64", name, context));
        case "StoreUintptr":
            return functionValue((args, context) => atomicStoreInteger(args, "uintptr", name, context));
        case "SwapInt32":
            return functionValue((args, context) => atomicSwapInteger(args, "int32", name, context));
        case "SwapUint32":
            return functionValue((args, context) => atomicSwapInteger(args, "uint32", name, context));
        case "SwapUint64":
            return functionValue((args, context) => atomicSwapInteger(args, "uint64", name, context));
        case "SwapUintptr":
            return functionValue((args, context) => atomicSwapInteger(args, "uintptr", name, context));
        case "CompareAndSwapInt32":
            return functionValue((args, context) => atomicCompareAndSwapInteger(args, "int32", name, context));
        case "CompareAndSwapUint32":
            return functionValue((args, context) => atomicCompareAndSwapInteger(args, "uint32", name, context));
        case "CompareAndSwapUint64":
            return functionValue((args, context) => atomicCompareAndSwapInteger(args, "uint64", name, context));
        case "CompareAndSwapUintptr":
            return functionValue((args, context) => atomicCompareAndSwapInteger(args, "uintptr", name, context));
        case "AddInt32":
            return functionValue((args, context) => atomicAddInteger(args, "int32", name, context));
        case "AddUint32":
            return functionValue((args, context) => atomicAddInteger(args, "uint32", name, context));
        case "AddUint64":
            return functionValue((args, context) => atomicAddInteger(args, "uint64", name, context));
        case "AddUintptr":
            return functionValue((args, context) => atomicAddInteger(args, "uintptr", name, context));
        case "AndInt32":
            return functionValue((args, context) => atomicBitwiseInteger(args, "int32", name, "&", context));
        case "AndUint32":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uint32", name, "&", context));
        case "AndUint64":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uint64", name, "&", context));
        case "AndUintptr":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uintptr", name, "&", context));
        case "OrInt32":
            return functionValue((args, context) => atomicBitwiseInteger(args, "int32", name, "|", context));
        case "OrUint32":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uint32", name, "|", context));
        case "OrUint64":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uint64", name, "|", context));
        case "OrUintptr":
            return functionValue((args, context) => atomicBitwiseInteger(args, "uintptr", name, "|", context));
        case "LoadPointer":
            return functionValue((args) => atomicPointerArg(args[0] ?? null, name).get() ?? new RuntimeTypedNilValue("unsafe.Pointer"));
        case "StorePointer":
            return functionValue((args) => {
                atomicPointerArg(args[0] ?? null, name).set(args[1] ?? new RuntimeTypedNilValue("unsafe.Pointer"));
                return null;
            });
        case "SwapPointer":
            return functionValue((args) => {
                const ptr = atomicPointerArg(args[0] ?? null, name);
                const previous = ptr.get() ?? new RuntimeTypedNilValue("unsafe.Pointer");
                ptr.set(args[1] ?? new RuntimeTypedNilValue("unsafe.Pointer"));
                return previous;
            });
        case "CompareAndSwapPointer":
            return functionValue((args) => {
                const ptr = atomicPointerArg(args[0] ?? null, name);
                const current = ptr.get() ?? new RuntimeTypedNilValue("unsafe.Pointer");
                if (!valueEqual(current, args[1] ?? new RuntimeTypedNilValue("unsafe.Pointer")))
                    return false;
                ptr.set(args[2] ?? new RuntimeTypedNilValue("unsafe.Pointer"));
                return true;
            });
        default:
            return undefined;
    }
}
function atomicPointerArg(value, name) {
    if (!(value instanceof RuntimePointer))
        throwTypeError(value, "pointer", `sync/atomic.${name} address`);
    return value;
}
function atomicLoadInteger(args, type, name) {
    return atomicPreparedInteger(atomicPointerArg(args[0] ?? null, name).get() ?? 0n, type, `sync/atomic.${name}`);
}
function atomicStoreInteger(args, type, name, context) {
    atomicPointerArg(args[0] ?? null, name).set(prepareAssignableToType(args[1] ?? 0n, type, `sync/atomic.${name} value`, context));
    return null;
}
function atomicSwapInteger(args, type, name, context) {
    const ptr = atomicPointerArg(args[0] ?? null, name);
    const previous = atomicPreparedInteger(ptr.get() ?? 0n, type, `sync/atomic.${name}`);
    ptr.set(prepareAssignableToType(args[1] ?? 0n, type, `sync/atomic.${name} value`, context));
    return previous;
}
function atomicCompareAndSwapInteger(args, type, name, context) {
    const ptr = atomicPointerArg(args[0] ?? null, name);
    const current = atomicPreparedInteger(ptr.get() ?? 0n, type, `sync/atomic.${name}`);
    const oldValue = prepareAssignableToType(args[1] ?? 0n, type, `sync/atomic.${name} old value`, context);
    if (!valueEqual(current, oldValue))
        return false;
    ptr.set(prepareAssignableToType(args[2] ?? 0n, type, `sync/atomic.${name} new value`, context));
    return true;
}
function atomicAddInteger(args, type, name, context) {
    const ptr = atomicPointerArg(args[0] ?? null, name);
    const current = toBigInt(atomicPreparedInteger(ptr.get() ?? 0n, type, `sync/atomic.${name}`));
    const delta = toBigInt(prepareAssignableToType(args[1] ?? 0n, type, `sync/atomic.${name} delta`, context));
    const next = prepareAssignableToType(current + delta, type, `sync/atomic.${name} result`, context);
    ptr.set(next);
    return next;
}
function atomicBitwiseInteger(args, type, name, operator, context) {
    const ptr = atomicPointerArg(args[0] ?? null, name);
    const previous = atomicPreparedInteger(ptr.get() ?? 0n, type, `sync/atomic.${name}`);
    const mask = toBigInt(prepareAssignableToType(args[1] ?? 0n, type, `sync/atomic.${name} mask`, context));
    const previousInt = toBigInt(previous);
    const next = prepareAssignableToType(operator === "&" ? previousInt & mask : previousInt | mask, type, `sync/atomic.${name} result`, context);
    ptr.set(next);
    return previous;
}
function atomicPreparedInteger(value, type, role) {
    return prepareAssignableToType(value, type, role);
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
    const resolved = context.resolveImportedTypeText(typeText);
    return parseArrayOrSliceTypeText(resolved, context)?.typeText ?? resolved;
}
function installAutomaticImports(context) {
    context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}
function availablePackages(context) {
    const sourcePackages = context.packages();
    const packages = {
        cmp: cmpPackage(),
        fmt: fmtPackage(),
        math: mathPackage(),
        os: osPackage(context),
        "runtime/pprof": runtimePprofPackage(),
        strconv: strconvPackage(),
        testing: testingPackage(context),
        ...sourcePackages,
        iter: iterPackage(),
        "internal/reflectlite": reflectlitePackage(),
        runtime: runtimePackage(),
        "syscall/js": syscallJSPackage(),
        unsafe: unsafePackage()
    };
    for (const [path, pkg] of Object.entries(packages)) {
        markRuntimePackageObject(pkg, path, context.packageInfo(path) ?? standardTypePackage(path));
    }
    return packages;
}
function cmpPackage() {
    return {
        Compare: hostCallable("cmp.Compare", (args) => BigInt(cmpCompare(args[0] ?? null, args[1] ?? null))),
        Less: hostCallable("cmp.Less", (args) => cmpLess(args[0] ?? null, args[1] ?? null)),
        Or: hostCallable("cmp.Or", (args) => args.find((arg) => !isRuntimeZeroValue(arg)) ?? zeroValueLike(args[0] ?? null))
    };
}
function iterPackage() {
    return {
        Pull: hostCallable("iter.Pull", (args, context, typeArguments) => iterPull(args[0] ?? null, context, 1, typeArguments), { tupleResult: true }),
        Pull2: hostCallable("iter.Pull2", (args, context, typeArguments) => iterPull(args[0] ?? null, context, 2, typeArguments), { tupleResult: true })
    };
}
function fmtPackage() {
    return {
        Printf: hostCallable("fmt.Printf", async (args, context) => {
            const format = await toStringValueAsync(args[0] ?? "", context);
            const text = await sprintfAsync(format, args.slice(1), context);
            context.write(text);
            return [BigInt(text.length), null];
        }, { tupleResult: true }),
        Fprintf: hostCallable("fmt.Fprintf", async (args, context) => {
            const writer = args[0] ?? null;
            const format = await toStringValueAsync(args[1] ?? "", context);
            return fmtWriteToWriter(writer, await sprintfAsync(format, args.slice(2), context), context);
        }, { tupleResult: true }),
        Print: hostCallable("fmt.Print", async (args, context) => {
            const text = await sprintAsync(args, context);
            context.write(text);
            return [BigInt(text.length), null];
        }, { tupleResult: true }),
        Fprint: hostCallable("fmt.Fprint", async (args, context) => {
            return fmtWriteToWriter(args[0] ?? null, await sprintAsync(args.slice(1), context), context);
        }, { tupleResult: true }),
        Fprintln: hostCallable("fmt.Fprintln", async (args, context) => {
            return fmtWriteToWriter(args[0] ?? null, await sprintlnAsync(args.slice(1), context), context);
        }, { tupleResult: true }),
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
            return [BigInt(text.length), null];
        }, { tupleResult: true })
    };
}
async function fmtWriteToWriter(writer, text, context) {
    const bytes = [...new TextEncoder().encode(text)].map((byte) => BigInt(byte));
    markArrayType(bytes, "[]byte");
    const writeMethod = methodForValue(writer, "Write", context);
    if (writeMethod) {
        const result = await callRuntime(boundMethodValue(writeMethod.method, writeMethod.receiver, context), [bytes], context);
        if (isTupleValues(result))
            return result;
    }
    return [BigInt(bytes.length), null];
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
        MaxInt: 9223372036854775807n,
        MinInt: -9223372036854775808n,
        MaxInt8: 127n,
        MinInt8: -128n,
        MaxInt16: 32767n,
        MinInt16: -32768n,
        MaxInt32: 2147483647n,
        MinInt32: -2147483648n,
        MaxInt64: 9223372036854775807n,
        MinInt64: -9223372036854775808n,
        MaxUint: 18446744073709551615n,
        MaxUint8: 255n,
        MaxUint16: 65535n,
        MaxUint32: 4294967295n,
        MaxUint64: 18446744073709551615n,
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
        GOOS: "js",
        GOARCH: "gojr",
        Compiler: "gojr",
        MemProfileRate: 0n,
        AddCleanup: hostCallable("runtime.AddCleanup", () => runtimeCleanupValue()),
        BlockProfile: gojrNotImplementedCallable("runtime.BlockProfile"),
        Caller: hostCallable("runtime.Caller", () => [0n, "", 0n, false], { tupleResult: true }),
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
        FileLine: hostCallable("runtime.(*Func).FileLine", () => ["", 0n], { tupleResult: true }),
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
        }, { tupleResult: true })
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
function testingPackage(context) {
    return {
        Short: hostCallable("testing.Short", () => false),
        Verbose: hostCallable("testing.Verbose", () => context.testingVerbose())
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
            const pointer = unsafePointerPayload(args[0] ?? null);
            if (isNilUnsafePointerPayload(pointer)) {
                if (length === 0)
                    return null;
                throw new GoJuniorRuntimeError("unsafe.Slice: ptr is nil and len is not zero");
            }
            if (!(pointer instanceof RuntimePointer))
                throw new GoJuniorRuntimeError(`${formatValue(args[0] ?? null)} is not a pointer`);
            const sequence = pointer.sequenceInfo();
            const values = sequence
                ? Array.from({ length }, (_item, index) => getArrayElement(sequence.values, sequence.index + index))
                : Array.from({ length }, (_item, index) => index === 0 ? pointer.get() : null);
            markArrayType(values, `[]${pointer.typeName}`);
            if (sequence)
                markArrayView(values, sequence.values, sequence.index);
            arrayCapacities.set(values, length);
            return values;
        }),
        SliceData: hostCallable("unsafe.SliceData", (args) => {
            const slice = unwrapNamed(args[0] ?? null);
            if (slice === null)
                return null;
            if (!Array.isArray(slice))
                throw new GoJuniorRuntimeError(`${formatValue(args[0] ?? null)} is not a slice`);
            if (slice.length === 0)
                return null;
            const sequence = arrayPointerSequence(slice, 0);
            const typeName = arrayElementTypeText(slice) ?? pointerTypeName(getArrayElement(sequence.values, sequence.index));
            return new RuntimePointer(typeName, () => getArrayElement(sequence.values, sequence.index), (next) => {
                setArrayElement(sequence.values, sequence.index, next);
            }, undefined, sequence);
        }),
        String: hostCallable("unsafe.String", (args) => {
            const length = toNonNegativeLength(args[1] ?? 0n, "unsafe.String length");
            const pointer = unsafePointerPayload(args[0] ?? null);
            if (isNilUnsafePointerPayload(pointer)) {
                if (length === 0)
                    return RuntimeGoString.fromUtf8Text("");
                throw new GoJuniorRuntimeError("unsafe.String: ptr is nil and len is not zero");
            }
            if (!(pointer instanceof RuntimePointer))
                throw new GoJuniorRuntimeError(`${formatValue(args[0] ?? null)} is not a pointer`);
            return RuntimeGoString.fromBytes(unsafePointerBytes(pointer, length));
        }),
        StringData: hostCallable("unsafe.StringData", (args) => {
            const value = unwrapNamed(args[0] ?? null);
            if (!isRuntimeString(value))
                throwTypeError(value, "string", "unsafe.StringData argument");
            const bytes = [...goStringBytes(value)].map((byte) => BigInt(byte));
            if (bytes.length === 0)
                return null;
            return new RuntimePointer("byte", () => bytes[0] ?? 0n, (next) => {
                bytes[0] = toBigInt(next);
            }, undefined, { values: bytes, index: 0 });
        })
    };
}
function unsafePointerPayload(value) {
    return unwrapNamed(value);
}
function isNilUnsafePointerPayload(value) {
    return value === null ||
        (value instanceof RuntimeTypedNilValue && (value.typeName.startsWith("*") || isUnsafePointerType(value.typeName)));
}
function unsafePointerBytes(pointer, length) {
    if (length === 0)
        return new Uint8Array();
    const sequence = pointer.sequenceInfo();
    if (!sequence) {
        if (length !== 1)
            throw new GoJuniorRuntimeError("unsafe.String: pointer does not reference a byte sequence");
        return Uint8Array.from([bytealgByte(pointer.get())]);
    }
    const bytes = new Uint8Array(length);
    for (let index = 0; index < length; index += 1) {
        bytes[index] = bytealgByte(getArrayElement(sequence.values, sequence.index + index));
    }
    return bytes;
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
function osPackage(context) {
    return {
        Args: runtimeStringSlice(context?.argv() ?? ["gojr"]),
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
function hostCallable(name, call, options = {}) {
    return {
        kind: "HostCallable",
        name,
        ...(options.tupleResult ? { tupleResult: true } : {}),
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
function predeclarePackageVariables(declarations, pkg, context) {
    const varsByName = new Map();
    for (const object of packageScopeObjects(pkg)) {
        if (object instanceof GoTypesVar)
            varsByName.set(object.Name(), object);
    }
    for (const statement of declarations) {
        if (statement.kind !== "VarDecl")
            continue;
        for (const declaration of statement.declarations) {
            if (declaration.name === "_")
                continue;
            const checkedType = declaration.type ? varsByName.get(declaration.name)?.Type() : null;
            const typeText = checkedType
                ? GoTypesTypeString(checkedType, GoTypesRelativeTo(pkg))
                : declaration.type?.text;
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
    const groups = packageVarDeclarationGroups(declarations);
    const groupsByName = new Map();
    for (const group of groups) {
        for (const declaration of group.declarations) {
            if (declaration.name !== "_")
                groupsByName.set(declaration.name, group);
        }
    }
    const initialized = new Set();
    for (const initializer of initOrder ?? []) {
        const names = initializer.Lhs.map((object) => object.Name()).filter((name) => name !== "_");
        for (const name of names) {
            const group = groupsByName.get(name);
            if (!group || initialized.has(group))
                continue;
            await assignPackageVarDeclarationGroup(group, context);
            initialized.add(group);
        }
    }
    for (const group of groups) {
        if (initialized.has(group) || group.valueExpressions.length === 0)
            continue;
        await assignPackageVarDeclarationGroup(group, context);
    }
}
function packageVarDeclarationGroups(declarations) {
    const groups = [];
    for (const statement of declarations) {
        if (statement.kind !== "VarDecl")
            continue;
        groups.push(...varDeclarationGroups(statement.declarations));
    }
    return groups;
}
function varDeclarationGroups(declarations) {
    const groups = new Map();
    const ordered = [];
    let syntheticGroup = declarations.length;
    for (const declaration of declarations) {
        const groupKey = declaration.valueGroup ?? syntheticGroup++;
        let group = groups.get(groupKey);
        if (!group) {
            group = { declarations: [], valueExpressions: [] };
            groups.set(groupKey, group);
            ordered.push(group);
        }
        group.declarations.push(declaration);
        if (declaration.value &&
            declaration.valueIndex !== undefined &&
            declaration.valueIndex < (declaration.valueCount ?? 1) &&
            group.valueExpressions[declaration.valueIndex] === undefined) {
            group.valueExpressions[declaration.valueIndex] = declaration.value;
        }
    }
    for (const group of ordered) {
        group.valueExpressions = group.valueExpressions.filter((expression) => expression !== undefined);
    }
    return ordered;
}
async function assignPackageVarDeclarationGroup(group, context) {
    const values = await evaluateVarDeclarationGroupValues(group, context);
    for (const [index, declaration] of group.declarations.entries()) {
        if (declaration.name === "_")
            continue;
        const source = declarationSourceExpression(group, index);
        const value = prepareValueForTargetType(values[index] ?? null, declaration.type?.text, source, context, `variable ${declaration.name}`);
        if (context.hasLocal(declaration.name)) {
            context.assign(declaration.name, value);
            continue;
        }
        context.declare(declaration.name, value, true, declarationRuntimeTypeText(declaration, source, value, group, context));
    }
}
async function evaluateVarDeclarationGroupValues(group, context) {
    if (group.valueExpressions.length === 0) {
        return group.declarations.map((declaration) => defaultValueForDeclarationType(declaration.type, context));
    }
    const values = await evaluateAssignmentValues(group.valueExpressions, group.declarations.length, context);
    if (values.length !== group.declarations.length) {
        throw new GoJuniorRuntimeError(`var declaration count mismatch: ${group.declarations.length} names but ${values.length} values`);
    }
    return values;
}
function declarationSourceExpression(group, index) {
    return group.valueExpressions.length === group.declarations.length
        ? group.valueExpressions[index]
        : group.valueExpressions[0];
}
function declarationRuntimeTypeText(declaration, source, value, group, context) {
    if (declaration.type)
        return declaration.type;
    const index = group.declarations.indexOf(declaration);
    const tupleType = index >= 0
        ? tupleCallResultTypeText(group.valueExpressions, index, group.declarations.length, context)
        : undefined;
    if (tupleType)
        return tupleType;
    if (source && group.valueExpressions.length === group.declarations.length) {
        return expressionDeclaredTypeText(source, context, value) ?? inferredTypeText(value);
    }
    return inferredTypeText(value);
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
function goJuniorFunctionValue(name, signature, body, closureScope, declaration, boundReceiver, source, closureContext, typeArgumentBindings) {
    const formattedSource = source ?? declaration?.source;
    return {
        kind: "GoJuniorFunction",
        name,
        signature,
        ...(declaration ? { declaration } : {}),
        ...(formattedSource ? { source: formattedSource } : {}),
        ...(closureContext ? { callContext: closureContext } : {}),
        ...(typeArgumentBindings && Object.keys(typeArgumentBindings).length > 0 ? { typeArgumentBindings } : {}),
        async call(args, parentContext, typeArguments) {
            const context = closureContext ?? parentContext;
            const packageName = closureContext?.importPath();
            const crossesPackageBoundary = packageName !== undefined && packageName !== parentContext.importPath();
            const boundaryTypeArgumentBindings = preliminaryFunctionTypeArgumentBindings(declaration, signature, args, context, typeArguments, typeArgumentBindings);
            const boundarySignature = instantiateRuntimeSignatureTypeArguments(signature, boundaryTypeArgumentBindings);
            const boundaryReceiverType = declaration?.receiver
                ? instantiateRuntimeTypeNodeTypeArguments(declaration.receiver.type, boundaryTypeArgumentBindings)
                : undefined;
            const callArgs = crossesPackageBoundary
                ? args.map((arg, index) => packageBoundaryArgumentValue(arg, parameterTypeForArgument(boundarySignature, index), packageName, parentContext.importPath(), context))
                : args;
            const callReceiver = crossesPackageBoundary
                ? packageBoundaryArgumentValue(boundReceiver ?? null, boundaryReceiverType?.text, packageName, parentContext.importPath(), context)
                : boundReceiver;
            const invoke = () => context.functionCallAsync(() => context.childScopeAsync(async () => {
                const activeTypeArgumentBindings = bindFunctionTypeParameters(declaration, signature, callArgs, context, typeArguments, typeArgumentBindings);
                const activeSignature = instantiateRuntimeSignatureTypeArguments(signature, activeTypeArgumentBindings);
                if (declaration?.receiver?.name) {
                    const receiverType = instantiateRuntimeTypeNodeTypeArguments(declaration.receiver.type, activeTypeArgumentBindings);
                    context.declare(declaration.receiver.name, prepareReceiverBinding(receiverType.text, callReceiver), true, receiverType);
                }
                for (const [index, parameter] of activeSignature.parameters.entries()) {
                    if (parameter.name) {
                        const value = parameter.variadic ? callArgs.slice(index) : callArgs[index] ?? null;
                        context.declare(parameter.name, value, true, parameter.variadic ? `[]${parameter.type.text}` : parameter.type);
                        if (parameter.variadic)
                            break;
                    }
                }
                for (const result of activeSignature.results) {
                    if (result.name) {
                        context.declare(result.name, defaultValueForDeclarationType(result.type, context), true, result.type);
                    }
                }
                const outcome = await context.deferScopeAsync(() => executeBlock(body, context, false));
                const completion = outcome === RECOVERED_PANIC
                    ? { kind: "return", values: namedReturnValuesOrZero(activeSignature, context) }
                    : outcome;
                if (completion.kind === "return") {
                    const rawValues = completion.values.length === 0
                        ? namedReturnValues(activeSignature, context)
                        : completion.values;
                    const values = prepareFunctionReturnValues(activeSignature, expandFunctionReturnValues(activeSignature, rawValues), completion.sources, context);
                    if (closureContext && closureContext !== parentContext && closureContext.importPath()) {
                        const resultValues = values.map((value, index) => packageBoundaryResultValue(value ?? null, activeSignature.results[index]?.type.text, closureContext.importPath(), parentContext.importPath(), context));
                        return resultValues.length === 0 ? null : resultValues.length === 1 ? resultValues[0] ?? null : markTupleValues(resultValues);
                    }
                    const result = values.length === 0 ? null : values.length === 1 ? values[0] ?? null : markTupleValues(values);
                    return result;
                }
                return null;
            }));
            return closureScope ? context.withScopeAsync(closureScope, invoke) : invoke();
        }
    };
}
function preliminaryFunctionTypeArgumentBindings(declaration, signature, args, context, explicitTypeArguments, seedBindings) {
    const names = declaration?.typeParameters ?? [];
    if (names.length === 0)
        return seedBindings;
    const inferred = inferFunctionTypeArguments(names, signature, args, explicitTypeArguments, context, seedBindings);
    const bindings = { ...(seedBindings ?? {}) };
    const importPath = context.importPath();
    for (const name of names) {
        const type = inferred.get(name) ?? seedBindings?.[name];
        if (!type)
            continue;
        bindings[name] = type;
        if (importPath)
            bindings[`${importPath}.${name}`] = type;
    }
    return Object.keys(bindings).length > 0 ? bindings : undefined;
}
function parameterTypeForArgument(signature, index) {
    const parameter = signature.parameters[index] ?? (() => {
        for (let candidateIndex = signature.parameters.length - 1; candidateIndex >= 0; candidateIndex -= 1) {
            const candidate = signature.parameters[candidateIndex];
            if (candidate?.variadic && index >= candidateIndex)
                return candidate;
        }
        return undefined;
    })();
    return parameter?.type.text;
}
function packageBoundaryArgumentValue(value, typeText, calleePackageName, callerPackageName, calleeContext) {
    const interfaceBoundary = isDynamicInterfaceBoundaryType(typeText, calleeContext);
    const owner = typeText ? runtimeTypeOwnerPackagePath(typeText) : undefined;
    const genericOwner = !owner && callerPackageName && callerPackageName !== calleePackageName
        ? runtimeTypeOwnerPackagePath(qualifyBareGenericRuntimeTypeArguments(typeText, callerPackageName))
        : undefined;
    const dynamicOwner = interfaceBoundary
        ? runtimeDynamicValueOwnerPackagePath(value) ?? runtimeValueLocalOwnerFallback(value, callerPackageName)
        : undefined;
    const effectiveOwner = dynamicOwner ?? owner ?? genericOwner;
    return effectiveOwner && effectiveOwner !== calleePackageName
        ? externalizePackageRuntimeValue(value, effectiveOwner)
        : internalizePackageRuntimeValue(value, calleePackageName);
}
function qualifyBareGenericRuntimeTypeArguments(typeText, packageName) {
    const type = normalizeTypeText(typeText ?? "");
    if (!type)
        return type;
    if (type.startsWith("*"))
        return `*${qualifyBareGenericRuntimeTypeArguments(type.slice(1), packageName)}`;
    if (type.startsWith("[]"))
        return `[]${qualifyBareGenericRuntimeTypeArguments(type.slice(2), packageName)}`;
    const generic = genericTypeArguments(type);
    if (!generic)
        return type;
    return `${generic.base}[${generic.args.map((arg) => {
        const normalized = normalizeTypeText(arg);
        if (!normalized || runtimeTypeOwnerPackagePath(normalized) || isPredeclaredType(normalized))
            return normalized;
        return qualifyLocalRuntimeTypeName(normalized, packageName);
    }).join(", ")}]`;
}
function packageBoundaryResultValue(value, typeText, calleePackageName, callerPackageName, calleeContext) {
    const interfaceBoundary = isDynamicInterfaceBoundaryType(typeText, calleeContext);
    const owner = runtimeTypeOwnerPackagePath(typeText);
    const dynamicOwner = interfaceBoundary ? runtimeDynamicValueOwnerPackagePath(value) : undefined;
    const effectiveOwner = dynamicOwner ?? owner;
    if (effectiveOwner && callerPackageName === effectiveOwner)
        return internalizePackageRuntimeValue(value, effectiveOwner);
    return externalizePackageRuntimeValue(value, effectiveOwner ?? calleePackageName);
}
function isDynamicInterfaceBoundaryType(typeText, context) {
    const type = normalizeTypeText(typeText ?? "");
    return type === "any" || type === "interface{}" || type.startsWith("interface{") ||
        Boolean(type && context?.interfaceDef(type));
}
function runtimeDynamicValueOwnerPackagePath(value) {
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null ? undefined : runtimeDynamicValueOwnerPackagePath(value.value);
    return runtimeValueOwnerPackagePath(value);
}
function runtimeValueOwnerPackagePath(value) {
    if (value instanceof RuntimeInterfaceValue) {
        return runtimeValueOwnerPackagePath(value.value) ?? runtimeTypeOwnerPackagePath(value.interfaceType);
    }
    if (value instanceof RuntimeNamedValue) {
        return runtimeTypeOwnerPackagePath(value.typeName) ?? runtimeValueOwnerPackagePath(value.value);
    }
    if (value instanceof RuntimeTypedNilValue)
        return runtimeTypeOwnerPackagePath(value.typeName);
    if (value instanceof RuntimePointer)
        return runtimeTypeOwnerPackagePath(value.typeName) ?? runtimeValueOwnerPackagePath(value.get());
    if (value instanceof RuntimeStruct)
        return runtimeTypeOwnerPackagePath(value.typeName);
    if (value instanceof RuntimeMap)
        return runtimeTypeOwnerPackagePath(`map[${value.keyType}]${value.valueType}`);
    if (value instanceof RuntimeChannel)
        return runtimeTypeOwnerPackagePath(`chan ${value.elementType}`);
    return undefined;
}
function runtimeValueLocalOwnerFallback(value, callerPackageName) {
    if (!callerPackageName)
        return undefined;
    const type = runtimeValueConcreteTypeText(value);
    return type && runtimeTypeNeedsPackageOwner(type) ? callerPackageName : undefined;
}
function runtimeValueConcreteTypeText(value) {
    if (value instanceof RuntimeInterfaceValue)
        return runtimeValueConcreteTypeText(value.value);
    if (value instanceof RuntimeNamedValue)
        return value.typeName;
    if (value instanceof RuntimeTypedNilValue)
        return value.typeName;
    if (value instanceof RuntimePointer)
        return `*${value.typeName}`;
    if (value instanceof RuntimeStruct)
        return value.typeName;
    if (value instanceof RuntimeMap)
        return `map[${value.keyType}]${value.valueType}`;
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (Array.isArray(value))
        return arrayTypeTexts.get(value);
    return undefined;
}
function runtimeTypeNeedsPackageOwner(typeText) {
    const type = normalizeTypeText(typeText);
    if (!type || runtimeTypeOwnerPackagePath(type))
        return false;
    if (isPredeclaredType(type) || isUnsafePointerType(type))
        return false;
    if (type.startsWith("*"))
        return runtimeTypeNeedsPackageOwner(type.slice(1));
    if (type.startsWith("[]"))
        return runtimeTypeNeedsPackageOwner(type.slice(2));
    const arrayType = parseArrayOrSliceTypeText(type);
    if (arrayType)
        return runtimeTypeNeedsPackageOwner(arrayType.elementType);
    const mapType = parseMapTypeText(type);
    if (mapType)
        return runtimeTypeNeedsPackageOwner(mapType.keyType) || runtimeTypeNeedsPackageOwner(mapType.valueType);
    const chanType = parseChanTypeText(type);
    if (chanType)
        return runtimeTypeNeedsPackageOwner(chanType.elementType);
    const generic = genericTypeArguments(type);
    if (generic)
        return runtimeTypeNeedsPackageOwner(generic.base) || generic.args.some(runtimeTypeNeedsPackageOwner);
    return /^[A-Za-z_]\w*$/.test(type);
}
function expandFunctionReturnValues(signature, values) {
    if (values.length !== 1 || signature.results.length <= 1)
        return values;
    const first = values[0];
    return first !== undefined && isTupleValues(first) ? first : values;
}
function bindFunctionTypeParameters(declaration, signature, args, context, explicitTypeArguments, seedBindings) {
    const names = declaration?.typeParameters ?? [];
    if (names.length === 0)
        return seedBindings;
    const inferred = inferFunctionTypeArguments(names, signature, args, explicitTypeArguments, context, seedBindings);
    const bindings = {};
    const importPath = context.importPath();
    for (const [name, type] of Object.entries(seedBindings ?? {})) {
        bindings[name] = type;
    }
    for (const name of names) {
        const type = inferred.get(name) ?? seedBindings?.[name] ?? name;
        context.declareTypeAlias(name, type);
        bindings[name] = type;
        if (importPath)
            bindings[`${importPath}.${name}`] = type;
    }
    return Object.keys(bindings).length > 0 ? bindings : undefined;
}
function instantiateRuntimeSignatureTypeArguments(signature, bindings) {
    if (!bindings)
        return signature;
    return {
        parameters: signature.parameters.map((parameter) => ({
            ...parameter,
            type: instantiateRuntimeTypeNodeTypeArguments(parameter.type, bindings)
        })),
        results: signature.results.map((result) => ({
            ...result,
            type: instantiateRuntimeTypeNodeTypeArguments(result.type, bindings)
        }))
    };
}
function instantiateRuntimeTypeNodeTypeArguments(type, bindings) {
    if (!bindings)
        return type;
    return { ...type, text: substituteTypeArgumentBindings(type.text, bindings) };
}
function inferFunctionTypeArguments(names, signature, args, explicitTypeArguments, context, seedBindings) {
    const typeNames = new Set(names);
    const inferred = new Map();
    for (const [name, type] of Object.entries(seedBindings ?? {})) {
        const localName = name.includes(".") ? name.slice(name.lastIndexOf(".") + 1) : name;
        if (typeNames.has(localName))
            inferred.set(localName, normalizeTypeText(context.resolveImportedTypeText(type)));
    }
    for (const [index, name] of names.entries()) {
        const explicit = explicitTypeArguments?.[index];
        if (explicit)
            inferred.set(name, normalizeTypeText(context.resolveImportedTypeText(explicit)));
    }
    for (const [index, parameter] of signature.parameters.entries()) {
        const arg = parameter.variadic ? args.slice(index) : args[index];
        inferTypeArgumentsFromRuntimeValue(parameter.variadic ? `[]${parameter.type.text}` : parameter.type.text, arg ?? null, typeNames, inferred, context);
        if (parameter.variadic)
            break;
    }
    return inferred;
}
function inferTypeArgumentsFromRuntimeValue(patternType, value, typeNames, inferred, context) {
    if (isGoJuniorFunction(value) && value.signature) {
        inferTypeArgumentsFromTypeText(patternType, signatureTypeText(value.signature), typeNames, inferred, context);
        return;
    }
    const actual = inferredTypeText(value);
    if (actual)
        inferTypeArgumentsFromTypeText(patternType, actual, typeNames, inferred, context);
}
function inferTypeArgumentsFromTypeText(patternType, actualType, typeNames, inferred, context) {
    const pattern = normalizeTypeText(context?.resolveImportedTypeText(patternType) ?? patternType);
    const actual = normalizeTypeText(context?.resolveImportedTypeText(actualType) ?? actualType);
    if (typeNames.has(pattern)) {
        inferred.set(pattern, actual);
        return;
    }
    if (pattern.startsWith("*") && actual.startsWith("*")) {
        inferTypeArgumentsFromTypeText(pattern.slice(1), actual.slice(1), typeNames, inferred, context);
        return;
    }
    if (pattern.startsWith("[]") && actual.startsWith("[]")) {
        inferTypeArgumentsFromTypeText(pattern.slice(2), actual.slice(2), typeNames, inferred, context);
        return;
    }
    const patternArray = parseArrayOrSliceTypeText(pattern, context);
    const actualArray = parseArrayOrSliceTypeText(actual, context);
    if (patternArray && actualArray) {
        inferTypeArgumentsFromTypeText(patternArray.elementType, actualArray.elementType, typeNames, inferred, context);
        return;
    }
    const patternMap = parseMapTypeText(pattern);
    const actualMap = parseMapTypeText(actual);
    if (patternMap && actualMap) {
        inferTypeArgumentsFromTypeText(patternMap.keyType, actualMap.keyType, typeNames, inferred, context);
        inferTypeArgumentsFromTypeText(patternMap.valueType, actualMap.valueType, typeNames, inferred, context);
        return;
    }
    const patternFunc = parseFunctionTypeText(pattern);
    const actualFunc = parseFunctionTypeText(actual);
    if (patternFunc && actualFunc) {
        inferTypeArgumentsFromSignatures(patternFunc, actualFunc, typeNames, inferred, context);
        return;
    }
    const patternIndex = genericTypeArguments(pattern);
    const actualIndex = genericTypeArguments(actual);
    if (patternIndex && actualIndex && patternIndex.base === actualIndex.base && patternIndex.args.length === actualIndex.args.length) {
        for (let index = 0; index < patternIndex.args.length; index += 1) {
            inferTypeArgumentsFromTypeText(patternIndex.args[index], actualIndex.args[index], typeNames, inferred, context);
        }
    }
}
function inferTypeArgumentsFromSignatures(pattern, actual, typeNames, inferred, context) {
    for (let index = 0; index < Math.min(pattern.parameters.length, actual.parameters.length); index += 1) {
        inferTypeArgumentsFromTypeText(pattern.parameters[index].type.text, actual.parameters[index].type.text, typeNames, inferred, context);
    }
    for (let index = 0; index < Math.min(pattern.results.length, actual.results.length); index += 1) {
        inferTypeArgumentsFromTypeText(pattern.results[index].type.text, actual.results[index].type.text, typeNames, inferred, context);
    }
}
function signatureTypeText(signature) {
    return `func(${signature.parameters.map(parameterTypeText).join(", ")})${signatureResultTypeText(signature.results)}`;
}
function parameterTypeText(parameter) {
    return `${parameter.variadic ? "..." : ""}${parameter.type.text}`;
}
function signatureResultTypeText(results) {
    if (results.length === 0)
        return "";
    if (results.length === 1 && !results[0]?.name)
        return ` ${results[0].type.text}`;
    return ` (${results.map(parameterTypeText).join(", ")})`;
}
function parseFunctionTypeText(typeText) {
    const text = normalizeTypeText(typeText);
    if (!text.startsWith("func("))
        return undefined;
    const open = text.indexOf("(");
    const close = matchingParen(text, open);
    if (close < 0)
        return undefined;
    return {
        parameters: parseSignatureParameterListText(text.slice(open + 1, close)),
        results: parseSignatureResultText(text.slice(close + 1).trim())
    };
}
function genericTypeArguments(typeText) {
    const text = normalizeTypeText(typeText);
    if (!text.endsWith("]") || text.startsWith("[") || text.startsWith("map["))
        return undefined;
    const bracket = text.indexOf("[");
    if (bracket <= 0)
        return undefined;
    return {
        base: text.slice(0, bracket),
        args: splitTopLevelTypeArguments(text.slice(bracket + 1, -1))
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
function boundMethodValue(method, receiver, callerContext) {
    const receiverType = receiverTypeName(receiver) ?? method.receiverType;
    const boundReceiver = method.pointerReceiver
        ? pointerToReceiver(receiver, receiverType)
        : valueReceiver(receiver);
    if (method.intrinsic) {
        return {
            kind: "GoJuniorFunction",
            name: `${receiverType}.${method.declaration.name}`,
            signature: method.declaration.signature,
            declaration: method.declaration,
            ...(method.declaration.source ? { source: method.declaration.source } : {}),
            ...(method.closureContext ? { callContext: method.closureContext } : {}),
            async call(args, context, typeArguments) {
                return method.intrinsic.call([boundReceiver, ...args], context, typeArguments);
            }
        };
    }
    const typeArgumentBindings = methodReceiverTypeArgumentBindings(method, receiverType, callerContext);
    return goJuniorFunctionValue(method.declaration.name, method.declaration.signature, method.declaration.body, method.closureScope, method.declaration, boundReceiver, undefined, method.closureContext, typeArgumentBindings);
}
function methodReceiverTypeArgumentBindings(method, receiverType, callerContext) {
    const names = method.declaration.typeParameters ?? [];
    if (names.length === 0 || !method.declaration.receiver)
        return undefined;
    const pattern = methodReceiverPatternType(method);
    const actual = method.pointerReceiver ? `*${receiverType}` : receiverType;
    const inferred = new Map();
    inferTypeArgumentsFromTypeText(pattern, actual, new Set(names), inferred, method.closureContext);
    if (inferred.size === 0)
        return undefined;
    const bindings = {};
    const importPath = method.closureContext?.importPath();
    for (const [name, type] of inferred) {
        const resolvedType = qualifyInferredReceiverTypeArgument(type, method, callerContext);
        bindings[name] = resolvedType;
        if (importPath)
            bindings[`${importPath}.${name}`] = resolvedType;
    }
    return bindings;
}
function qualifyInferredReceiverTypeArgument(type, method, callerContext) {
    const callerImportPath = callerContext?.importPath();
    const methodImportPath = method.closureContext?.importPath();
    if (!callerImportPath || callerImportPath === methodImportPath)
        return type;
    return qualifyLocalRuntimeTypeName(type, callerImportPath);
}
function methodReceiverPatternType(method) {
    const receiver = normalizeReceiverType(method.declaration.receiver?.type.text ?? method.receiverType);
    const args = genericTypeArguments(receiver.baseType)?.args;
    const base = args && args.length > 0
        ? `${method.receiverType}[${args.join(", ")}]`
        : method.receiverType;
    return receiver.pointer ? `*${base}` : base;
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
        ...(method.closureContext ? { callContext: method.closureContext } : {}),
        async call(args, context) {
            const receiver = args[0] ?? null;
            const methodArgs = args.slice(1);
            return callRuntime(boundMethodValue(method, receiver, context), methodArgs, context);
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
            return callRuntime(boundMethodValue(method.method, method.receiver, context), args.slice(1), context);
        }
    };
}
function boundDynamicInterfaceMethodValue(receiver, receiverTypeText, methodName, methodSignature) {
    return {
        kind: "GoJuniorFunction",
        name: `${receiverTypeText}.${methodName}`,
        signature: methodSignature,
        async call(args, context, typeArguments) {
            const method = methodForValue(receiver, methodName, context);
            if (!method) {
                const fmtError = methodName === "Error" ? fmtErrorMessage(receiver) : undefined;
                if (fmtError !== undefined)
                    return fmtError;
                throw new GoJuniorRuntimeError(`${formatValue(receiver)} has no method ${methodName}`);
            }
            return callRuntime(boundMethodValue(method.method, method.receiver, context), args, context, typeArguments);
        }
    };
}
function interfaceMethodDef(receiverTypeText, method) {
    const receiverType = normalizeReceiverType(receiverTypeText).baseType;
    const declaration = {
        kind: "FunctionDecl",
        name: method.name,
        receiver: { type: { text: receiverTypeText } },
        signature: method.signature,
        body: { kind: "BlockStatement", statements: [] }
    };
    return {
        declaration,
        receiverType,
        pointerReceiver: false,
        intrinsic: {
            kind: "GoJuniorFunction",
            name: `${receiverTypeText}.${method.name}`,
            signature: method.signature,
            declaration,
            async call(args, context, typeArguments) {
                const receiver = args[0] ?? null;
                const lookup = receiver instanceof RuntimeInterfaceValue && receiver.value === null
                    ? undefined
                    : methodForValue(receiver, method.name, context);
                if (!lookup)
                    throw new GoJuniorPanic(`runtime error: nil embedded interface method ${method.name}`);
                return callRuntime(boundMethodValue(lookup.method, lookup.receiver, context), args.slice(1), context, typeArguments);
            }
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
    const receiverType = receiver instanceof RuntimeNamedValue
        ? receiver.typeName
        : receiver instanceof RuntimeStruct
            ? receiver.typeName
            : typeName;
    const receiverStruct = unwrapNamed(receiver);
    if (receiverStruct instanceof RuntimeStruct && runtimeTypeAssignableMatch(receiverType, typeName)) {
        return new RuntimePointer(receiverType, () => receiver, (value) => {
            const valueType = value instanceof RuntimeNamedValue
                ? value.typeName
                : value instanceof RuntimeStruct
                    ? value.typeName
                    : "";
            const valueStruct = unwrapNamed(value);
            if (!(valueStruct instanceof RuntimeStruct) || !runtimeTypeAssignableMatch(valueType, receiverType)) {
                throw new GoJuniorRuntimeError(`${formatValue(value)} is not assignable to *${typeName}`);
            }
            receiverStruct.fields.clear();
            for (const [field, item] of valueStruct.orderedFields())
                receiverStruct.set(field, item);
        }, `struct:${objectIdentityId(receiverStruct)}`);
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
            for (const group of varDeclarationGroups(statement.declarations)) {
                const values = await evaluateVarDeclarationGroupValues(group, context);
                for (const [index, declaration] of group.declarations.entries()) {
                    if (declaration.name === "_")
                        continue;
                    const source = declarationSourceExpression(group, index);
                    const value = prepareValueForTargetType(values[index] ?? null, declaration.type?.text, source, context, `variable ${declaration.name}`);
                    context.declare(declaration.name, value, true, declarationRuntimeTypeText(declaration, source, value, group, context));
                }
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
        throw new GoJuniorRuntimeError(`loop exceeded ${context.loopLimit()} iterations`, statement.span);
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
        const typeArguments = callTypeArguments(statement.expression.callee, context);
        const callee = await evaluateExpression(statement.expression.callee, context);
        const args = await evaluateCallArguments(callee, statement.expression.args, statement.expression.spreadLast, context, typeArguments);
        context.pushDefer(() => callRuntime(callee, args, context, typeArguments));
        return;
    }
    context.pushDefer(() => evaluateExpression(statement.expression, context));
}
async function executeGo(statement, context) {
    const typeArguments = callTypeArguments(statement.call.callee, context);
    const callee = await evaluateExpression(statement.call.callee, context);
    const args = await evaluateCallArguments(callee, statement.call.args, statement.call.spreadLast, context, typeArguments);
    const goroutineContext = context.fork();
    context.scheduler().go(async () => {
        await callRuntime(callee, args, goroutineContext, typeArguments);
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
    const expectedTypes = statement.values.length === statement.targets.length
        ? statement.targets.map((target) => assignmentTargetExpectedTypeText(target, context))
        : [];
    const expectedRoles = statement.values.length === statement.targets.length
        ? statement.targets.map((target) => assignmentTargetRole(target))
        : [];
    const values = await evaluateAssignmentValues(statement.values, statement.targets.length, context, expectedTypes, expectedRoles);
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
            await assignExpressionTarget(target, value, context, source);
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
        const source = assignmentSourceExpression(sources, index, names.length);
        if (context.hasLocal(name)) {
            context.assign(name, prepareValueForTargetType(value, context.lookupTypeText(name), source, context, `variable ${name}`));
        }
        else {
            const typeText = assignmentDeclaredTypeText(sources, index, names.length, context, value);
            context.declare(name, value, true, typeText);
        }
    }
}
function assignmentSourceExpression(sources, index, targetCount) {
    if (sources.length === targetCount)
        return sources[index];
    if (sources.length === 1 && targetCount === 1)
        return sources[0];
    if (sources.length === 1 && index === 0 && isMultiValueExpression(sources[0]))
        return sources[0];
    return undefined;
}
function assignmentDeclaredTypeText(sources, index, targetCount, context, value) {
    const tupleType = tupleCallResultTypeText(sources, index, targetCount, context);
    if (tupleType)
        return tupleType;
    const source = assignmentSourceExpression(sources, index, targetCount);
    return expressionDeclaredTypeText(source, context, value) ?? inferredTypeText(value);
}
function tupleCallResultTypeText(sources, index, targetCount, context) {
    if (targetCount <= 1 || sources.length !== 1)
        return undefined;
    const source = sources[0];
    if (!source)
        return undefined;
    if (source.kind === "CallExpression")
        return callExpressionResultTypeText(source, context, index);
    if (targetCount === 2 && source.kind === "IndexExpression") {
        if (index === 1)
            return "bool";
        const objectType = expressionDeclaredTypeText(source.object, context);
        return indexResultTypeText(objectType);
    }
    if (targetCount === 2 && source.kind === "TypeAssertionExpression") {
        return index === 1 ? "bool" : normalizeTypeText(source.type.text);
    }
    if (targetCount === 2 && source.kind === "UnaryExpression" && source.operator === "<-") {
        if (index === 1)
            return "bool";
        const operandType = expressionDeclaredTypeText(source.operand, context);
        return channelElementTypeText(operandType);
    }
    return undefined;
}
function isMultiValueExpression(expression) {
    return expression?.kind === "CallExpression" ||
        expression?.kind === "IndexExpression" ||
        expression?.kind === "TypeAssertionExpression" ||
        (expression?.kind === "UnaryExpression" && expression.operator === "<-");
}
async function evaluateAssignmentValues(expressions, targetCount, context, expectedTypes = [], expectedRoles = []) {
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
    const values = expectedTypes.length === expressions.length
        ? await evaluateExpressionListWithExpectedTypes(expressions, context, expectedTypes, expectedRoles)
        : await evaluateExpressionList(expressions, context);
    if (targetCount > 1 && values.length === 1) {
        const first = values[0];
        if (first !== undefined && isTupleValues(first)) {
            return first;
        }
    }
    return values;
}
async function evaluateExpressionListWithExpectedTypes(expressions, context, expectedTypes, expectedRoles) {
    const values = [];
    for (const [index, expression] of expressions.entries()) {
        values.push(await evaluateExpressionWithExpectedType(expression, context, expectedTypes[index], expectedRoles[index]));
    }
    return values;
}
async function evaluateExpressionWithExpectedType(expression, context, expectedTypeText, expectedRole) {
    if (expectedTypeText && expression.kind === "ArrayLiteralExpression") {
        return evaluateArrayLiteral(expression, context, expectedTypeText, expectedRole);
    }
    if (expectedTypeText && expression.kind === "StructLiteralExpression" && normalizeTypeText(expression.typeName) === "<missing>") {
        const typeText = resolveRuntimeCompositeAliases(context.resolveImportedTypeText(expectedTypeText), context);
        return evaluateStructLiteral({ ...expression, typeName: typeText }, context);
    }
    return evaluateExpression(expression, context);
}
function assignmentTargetExpectedTypeText(target, context) {
    if (target.kind === "Identifier")
        return context.lookupTypeText(target.name);
    return undefined;
}
function assignmentTargetRole(target) {
    if (target.kind === "Identifier")
        return `variable ${target.name}`;
    if (target.kind === "IndexExpression")
        return "index assignment";
    return undefined;
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
async function assignExpressionTarget(target, value, context, source) {
    if (target.kind === "Identifier") {
        if (target.name === "_")
            return;
        context.assign(target.name, value);
        return;
    }
    if (target.kind === "SelectorExpression") {
        await setSelector(target, value, context, source);
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
            if ((isRuntimeCallable(object) || isGoJuniorFunction(object)) && isTypeArgumentExpression(expression.index, context, true)) {
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
async function evaluateCallArguments(callee, expressions, spreadLast, context, typeArguments) {
    const raw = expandSingleMultiReturnCallArgument(await evaluateExpressionList(expressions, context), expressions, callee, spreadLast);
    const calleeName = runtimeCallableName(callee);
    const args = spreadLast
        ? spreadLastArgument(raw, context, calleeName, expressions.at(-1)?.span, expressionDeclaredTypeText(expressions[0], context))
        : raw;
    if (!isGoJuniorFunction(callee) || !callee.signature)
        return args;
    const prepared = [...args];
    const typeArgumentBindings = inferRuntimeCallTypeArgumentBindings(callee, args, typeArguments, context, expressions);
    for (const [index, parameter] of callee.signature.parameters.entries()) {
        const parameterType = calleeParameterTypeText(callee, parameter.type.text, context, typeArgumentBindings);
        if (parameter.variadic) {
            for (let argIndex = index; argIndex < prepared.length; argIndex += 1) {
                const expression = spreadLast && argIndex >= expressions.length - 1
                    ? expressions[expressions.length - 1]
                    : expressions[argIndex];
                prepared[argIndex] = prepareValueForParameterType(prepared[argIndex] ?? null, parameterType, expression, context, `argument ${argIndex + 1} to ${calleeName}`, context);
            }
            break;
        }
        if (index >= prepared.length)
            break;
        prepared[index] = prepareValueForParameterType(prepared[index] ?? null, parameterType, expressions[index], context, `argument ${index + 1} to ${calleeName}`, context);
    }
    return prepared;
}
function expandSingleMultiReturnCallArgument(values, expressions, callee, spreadLast) {
    if (spreadLast)
        return values;
    if (expressions.length !== 1 || values.length !== 1)
        return values;
    const first = values[0];
    if (first === undefined || !isTupleValues(first))
        return values;
    if (!isGoJuniorFunction(callee) || !callee.signature)
        return values;
    if (callee.signature.parameters.length === 1 && !callee.signature.parameters[0]?.variadic)
        return values;
    return first;
}
function calleeParameterTypeText(callee, typeText, callerContext, typeArgumentBindings) {
    const packageName = callee.callContext?.importPath();
    const bindings = packageName
        ? qualifyRuntimeTypeArgumentBindings(typeArgumentBindings, packageName)
        : typeArgumentBindings;
    const resolved = packageName && packageName !== callerContext.importPath()
        ? qualifyLocalRuntimeTypeName(typeText, packageName)
        : typeText;
    return substituteTypeArgumentBindings(resolved, bindings);
}
function qualifyRuntimeTypeArgumentBindings(bindings, packageName) {
    if (!bindings)
        return undefined;
    const qualified = { ...bindings };
    for (const [name, type] of Object.entries(bindings)) {
        if (!name.includes("."))
            qualified[`${packageName}.${name}`] = type;
    }
    return qualified;
}
function inferRuntimeCallTypeArgumentBindings(callee, args, explicitTypeArguments, context, expressions) {
    const names = callee.declaration?.typeParameters ?? [];
    const seedBindings = callee.typeArgumentBindings;
    if (names.length === 0 && !seedBindings)
        return undefined;
    const inferred = names.length > 0 && callee.signature
        ? inferFunctionTypeArguments(names, callee.signature, args, explicitTypeArguments, context, seedBindings)
        : new Map();
    if (names.length > 0 && callee.signature) {
        inferRuntimeCallTypeArgumentsFromCaller(callee, args, expressions ?? [], names, inferred, context, seedBindings);
    }
    for (const [name, type] of Object.entries(seedBindings ?? {})) {
        if (!inferred.has(name))
            inferred.set(name, type);
    }
    if (inferred.size === 0)
        return undefined;
    const bindings = {};
    const importPath = callee.callContext?.importPath();
    for (const [name, type] of inferred) {
        bindings[name] = type;
        if (importPath && !name.includes("."))
            bindings[`${importPath}.${name}`] = type;
    }
    return bindings;
}
function inferRuntimeCallTypeArgumentsFromCaller(callee, args, expressions, names, inferred, callerContext, seedBindings) {
    const importPath = callee.callContext?.importPath();
    const typeNames = new Set(names);
    if (importPath) {
        for (const name of names)
            typeNames.add(`${importPath}.${name}`);
    }
    for (const [index, parameter] of callee.signature?.parameters.entries() ?? []) {
        const parameterType = calleeParameterTypeText(callee, parameter.type.text, callerContext, seedBindings);
        if (parameter.variadic) {
            for (let argIndex = index; argIndex < args.length; argIndex += 1) {
                const expression = expressions[argIndex];
                const actualType = expressionDeclaredTypeText(expression, callerContext, args[argIndex] ?? null) ?? inferredTypeText(args[argIndex] ?? null);
                if (actualType)
                    inferTypeArgumentsFromTypeText(parameterType, actualType, typeNames, inferred, callerContext);
            }
            break;
        }
        if (index >= args.length)
            break;
        const actualType = expressionDeclaredTypeText(expressions[index], callerContext, args[index] ?? null) ?? inferredTypeText(args[index] ?? null);
        if (actualType)
            inferTypeArgumentsFromTypeText(parameterType, actualType, typeNames, inferred, callerContext);
    }
}
function substituteTypeArgumentBindings(typeText, bindings) {
    if (!bindings)
        return typeText;
    let text = typeText;
    const entries = Object.entries(bindings).sort((left, right) => right[0].length - left[0].length);
    for (const [name, type] of entries) {
        text = replaceRuntimeTypeName(text, name, type);
    }
    return text;
}
function replaceRuntimeTypeName(text, name, replacement) {
    if (!name)
        return text;
    let output = "";
    for (let index = 0; index < text.length;) {
        if (text.startsWith(name, index) &&
            !isRuntimeTypeNameChar(text[index - 1]) &&
            !isRuntimeTypeNameChar(text[index + name.length])) {
            output += replacement;
            index += name.length;
            continue;
        }
        output += text[index] ?? "";
        index += 1;
    }
    return output;
}
function isRuntimeTypeNameChar(char) {
    return Boolean(char && /[A-Za-z0-9_./]/.test(char));
}
function runtimeCallableName(callee) {
    if (isGoJuniorFunction(callee) || isRuntimeCallable(callee))
        return callee.name;
    return "call";
}
async function evaluateArrayLiteral(expression, context, expectedTypeText, expectedRole) {
    const typeText = resolveRuntimeCompositeAliases(context.resolveImportedTypeText(expectedTypeText ?? expression.type.text), context);
    const type = parseArrayOrSliceTypeText(typeText, context);
    if (!type) {
        throw new GoJuniorRuntimeError(`${typeText} is not an array or slice literal type`);
    }
    const values = await evaluateArrayLiteralEntries(expression.elements, type.elementType, type.length, Boolean(type.inferLength), typeText, context, expectedRole);
    markArrayType(values, type.typeText);
    return values;
}
async function evaluateArrayLiteralEntries(elements, elementType, length, inferLength, displayType, context, expectedRole) {
    const values = [];
    let nextIndex = 0;
    for (const element of elements) {
        const index = element.key ? toNumber(await evaluateExpression(element.key, context)) : nextIndex;
        if (!Number.isInteger(index) || index < 0)
            throw new GoJuniorRuntimeError(`${displayType} array literal has invalid index ${index}`);
        if (length !== undefined && !inferLength && index >= length) {
            throw new GoJuniorRuntimeError(`${displayType} array literal index ${index} out of bounds`);
        }
        const elementRole = expectedRole ? `${expectedRole} element` : "array element";
        values[index] = prepareAssignableToType(await evaluateExpressionWithExpectedType(element.value, context, elementType, elementRole), elementType, elementRole, context);
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
    let typeDef = context.typeDef(typeName) ?? parseAnonymousStructTypeText(typeName);
    if (!typeDef) {
        const alias = context.aliasType(typeName);
        const aliasType = alias ? normalizeTypeText(context.resolveImportedTypeText(alias)) : undefined;
        typeDef = aliasType ? context.typeDef(aliasType) ?? parseAnonymousStructTypeText(aliasType) : undefined;
    }
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
    const fields = instantiateStructFields(typeDef, typeName);
    const struct = new RuntimeStruct(typeDef.name);
    const fieldDefaultContext = typeDef.declaringContext ?? context;
    for (const field of fields) {
        struct.set(field.name, defaultValueForTypeText(field.type.text, fieldDefaultContext));
    }
    for (const [index, field] of expression.fields.entries()) {
        const declared = field.name
            ? fields.find((candidate) => candidate.name === field.name)
            : fields[index];
        if (!declared) {
            throw new GoJuniorRuntimeError(field.name
                ? `${typeName} has no field ${field.name}`
                : `${typeName} literal has too many values`);
        }
        const value = await evaluateExpression(field.value, context);
        struct.set(declared.name, prepareValueForTargetType(value, declared.type.text, field.value, context, `field ${field.name ?? declared.name}`));
    }
    return typeName !== typeDef.name ? new RuntimeNamedValue(typeName, struct) : struct;
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
    const keyType = context.resolveImportedTypeText(expression.keyType.text);
    const valueType = context.resolveImportedTypeText(expression.valueType.text);
    const map = new RuntimeMap(keyType, valueType, context);
    for (const entry of expression.entries) {
        map.set(await evaluateExpressionWithExpectedType(entry.key, context, keyType, "map key"), await evaluateExpressionWithExpectedType(entry.value, context, valueType, "map value"));
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
    if (expression.callee.kind === "Identifier" && expression.callee.name === "make" && !context.hasBinding("make")) {
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
    const typeArguments = callTypeArguments(expression.callee, context);
    const callee = await evaluateExpression(expression.callee, context);
    const args = await evaluateCallArguments(callee, expression.args, expression.spreadLast, context, typeArguments);
    return callRuntime(callee, args, context, typeArguments);
}
function callTypeArguments(callee, context) {
    if (callee.kind !== "IndexExpression" || !isTypeArgumentExpression(callee.index, context, true))
        return undefined;
    const typeText = typeArgumentText(callee.index, true);
    if (typeText === undefined)
        return undefined;
    return splitTopLevelTypeArguments(typeText).map((type) => {
        const imported = context.resolveImportedTypeText(type);
        const scopedAlias = context.scopedAliasType(imported) ?? context.scopedAliasType(type);
        const resolved = resolveRuntimeCompositeAliases(scopedAlias ?? imported, context);
        const importPath = context.importPath();
        return importPath && scopedAlias === undefined ? qualifyLocalRuntimeTypeName(resolved, importPath) : resolved;
    });
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
    const rawTypeText = typeArgument ?? expressionDeclaredTypeText(argument, context, evaluated);
    if (!rawTypeText)
        throw new GoJuniorRuntimeError("new could not infer argument type");
    const typeText = resolveRuntimeCompositeAliases(context.resolveImportedTypeText(rawTypeText), context);
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
    const rawTypeText = typeArg ? makeTypeArgumentText(typeArg, context) : undefined;
    if (!rawTypeText) {
        throw new GoJuniorRuntimeError("make expects a map or slice type as its first argument");
    }
    const resultTypeText = normalizeTypeText(context.resolveImportedTypeText(rawTypeText));
    const typeText = makeUnderlyingTypeText(resultTypeText, context);
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
        return makeRuntimeSlice(arrayType.elementType, length, capacity, context, resultTypeText);
    }
    throw new GoJuniorRuntimeError(`cannot make ${typeText}`);
}
function makeTypeArgumentText(expression, context) {
    if (expression.kind === "TypeExpression")
        return expression.type.text;
    const typeText = typeArgumentText(expression, true);
    return typeText && context.isKnownType(typeText) ? typeText : undefined;
}
function makeUnderlyingTypeText(typeText, context) {
    const imported = normalizeTypeText(context.resolveImportedTypeText(typeText));
    const alias = context.aliasType(imported);
    return resolveRuntimeCompositeAliases(alias && alias !== imported ? alias : imported, context);
}
function spreadLastArgument(args, context, calleeName, span, appendTargetType) {
    if (args.length === 0)
        return args;
    const last = args[args.length - 1] ?? null;
    if (last instanceof RuntimeTypedNilValue && parseArrayOrSliceTypeText(last.typeName)) {
        return args.slice(0, -1);
    }
    if (calleeName === "append" && isRuntimeString(last) && appendTargetAcceptsStringSpread(args[0] ?? null, context, appendTargetType)) {
        return [...args.slice(0, -1), ...[...goStringBytes(last)].map((value) => BigInt(value))];
    }
    if (!Array.isArray(last)) {
        throw new GoJuniorRuntimeError(`${formatValue(last ?? null)} is not spreadable`, span);
    }
    return [...args.slice(0, -1), ...last];
}
function appendTargetAcceptsStringSpread(target, context, staticTypeText) {
    const actual = unwrapNamed(target);
    const runtimeTypeText = actual instanceof RuntimeTypedNilValue
        ? actual.typeName
        : Array.isArray(actual)
            ? arrayTypeTexts.get(actual)
            : undefined;
    for (const typeText of [staticTypeText, runtimeTypeText]) {
        if (!typeText)
            continue;
        const sliceType = parseArrayOrSliceTypeText(resolveRuntimeCompositeAliases(typeText, context), context);
        if (canonicalRuntimeTypeName(sliceType?.elementType ?? "") === "uint8")
            return true;
    }
    return false;
}
function appendValues(target, values) {
    const actualTarget = unwrapNamed(target);
    const typedNilSlice = actualTarget instanceof RuntimeTypedNilValue && parseArrayOrSliceTypeText(actualTarget.typeName);
    if (actualTarget !== null && !typedNilSlice && !Array.isArray(actualTarget)) {
        throw new GoJuniorRuntimeError(`${formatValue(target)} is not appendable`);
    }
    const base = Array.isArray(actualTarget) ? actualTarget : [];
    const appended = [...base, ...values];
    const previousCapacity = sliceCapacity(base);
    const needed = appended.length;
    arrayCapacities.set(appended, Math.max(previousCapacity, needed));
    const typeText = Array.isArray(actualTarget) ? arrayTypeTexts.get(actualTarget) : typedNilSlice && actualTarget instanceof RuntimeTypedNilValue ? actualTarget.typeName : undefined;
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
async function callRuntime(callee, args, context, typeArguments) {
    if (isRuntimeCallable(callee) || isGoJuniorFunction(callee)) {
        const result = await callee.call(args, context, typeArguments);
        if (Array.isArray(result) && ((isRuntimeCallable(callee) && callee.tupleResult === true) ||
            (isGoJuniorFunction(callee) && (callee.signature?.results.length ?? 0) > 1))) {
            return markTupleValues(result);
        }
        return result;
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
    const staticObjectType = expressionDeclaredTypeText(expression.object, context, object);
    const interfaceMethod = staticObjectType ? interfaceMethodForType(staticObjectType, expression.field, context) : undefined;
    if (interfaceMethod) {
        return boundDynamicInterfaceMethodValue(object, staticObjectType, expression.field, interfaceMethod.signature);
    }
    const staticFieldType = staticObjectType ? structFieldTypeText(staticObjectType, expression.field, context) : undefined;
    const runtimeObjectForSelector = unwrapNamed(dereferenceIfPointer(object));
    const runtimeObjectSelector = isRuntimeObject(runtimeObjectForSelector)
        ? runtimeObjectForSelector[expression.field]
        : undefined;
    if (staticObjectType && !staticFieldType && runtimeObjectSelector === undefined) {
        const staticMethod = await staticMethodSelectorValue(expression, object, staticObjectType, context);
        if (staticMethod)
            return staticMethod;
    }
    if (object instanceof RuntimeInterfaceValue) {
        const method = methodForValue(object, expression.field, context);
        if (method)
            return boundMethodValue(method.method, method.receiver, context);
        const fmtError = expression.field === "Error" ? fmtErrorMessage(object) : undefined;
        if (fmtError !== undefined) {
            return hostCallable("fmt.errorString.Error", () => fmtError);
        }
    }
    if (object instanceof RuntimePointer && isInternalAbiTypePointer(object, context)) {
        const method = methodForValue(object, expression.field, context);
        if (method)
            return boundMethodValue(method.method, method.receiver, context);
    }
    const struct = structFromValue(object);
    if (struct) {
        const fieldValue = struct.get(expression.field);
        if (fieldValue !== undefined)
            return fieldValue;
    }
    if (isRuntimeObject(runtimeObjectForSelector)) {
        const value = runtimeObjectForSelector[expression.field];
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
        return boundMethodValue(method.method, receiver, context);
    }
    const namedType = expression.object.kind === "Identifier" ? context.lookupTypeText(expression.object.name) : undefined;
    if (namedType && expression.object.kind === "Identifier") {
        const method = context.methodFor(normalizeTypeText(namedType), expression.field);
        if (method) {
            const receiver = method.pointerReceiver
                ? context.pointerToBinding(expression.object.name, normalizeTypeText(namedType))
                : object;
            return boundMethodValue(method, receiver, context);
        }
    }
    if (struct) {
        const promotedField = promotedFieldAccessor(struct, expression.field, context);
        if (promotedField)
            return promotedField.get() ?? null;
    }
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`, expression.span);
}
async function staticMethodSelectorValue(expression, object, staticTypeText, context) {
    const receiver = normalizeReceiverType(staticTypeText);
    const method = context.methodFor(receiver.baseType, expression.field);
    if (!method)
        return undefined;
    if (!method.pointerReceiver)
        return boundMethodValue(method, object, context);
    if (object instanceof RuntimePointer || isTypedNilPointer(object)) {
        return boundMethodValue(method, object, context);
    }
    const pointer = await pointerToExpressionIfAddressable(expression.object, context);
    return pointer ? boundMethodValue(method, pointer, context) : undefined;
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
        if (interfaceMethodForType(embedded.baseType, methodName, context))
            return true;
        const direct = context.methodFor(embedded.baseType, methodName);
        const embeddedAddressable = addressable || receiver.pointer || embedded.pointer;
        if (direct && (!direct.pointerReceiver || embeddedAddressable))
            return true;
        if (promotedMethodAvailableForType(embedded.baseType, methodName, context, embeddedAddressable, seen))
            return true;
    }
    return false;
}
async function setSelector(expression, value, context, source) {
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
        field.set(prepareValueForTargetType(value, field.type.type.text, source, context, `field ${expression.field}`));
        return;
    }
    if (isRuntimeObject(object)) {
        object[expression.field] = value;
        return;
    }
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`, expression.span);
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
        }, `field:${objectIdentityId(struct)}:${expression.field}`);
    }
    if (expression.kind === "IndexExpression") {
        const object = await evaluateExpression(expression.object, context);
        const index = await evaluateExpression(expression.index, context);
        if (!Array.isArray(object))
            throw new GoJuniorRuntimeError("address-of index requires an array or slice value");
        const numericIndex = toNumber(index);
        const sequence = arrayPointerSequence(object, numericIndex);
        const typeName = indexResultTypeText(expressionDeclaredTypeText(expression.object, context)) ??
            pointerTypeName(getArrayElement(object, numericIndex));
        return new RuntimePointer(typeName, () => getArrayElement(sequence.values, sequence.index), (next) => {
            setArrayElement(sequence.values, sequence.index, next);
        }, `array:${objectIdentityId(sequence.values)}:${sequence.index}`, sequence);
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
    const actual = unwrapNamed(dereferenceIfPointer(value));
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
    const declaredReceiverType = receiverTypeName(value);
    if (declaredReceiverType) {
        const direct = context.methodFor(declaredReceiverType, methodName);
        if (direct)
            return { method: direct, receiver: value };
    }
    const dereferenced = dereferenceIfPointer(value);
    const actual = unwrapNamed(dereferenced);
    const receiverType = dereferenced instanceof RuntimeNamedValue
        ? dereferenced.typeName
        : actual instanceof RuntimeStruct
            ? actual.typeName
            : actual instanceof RuntimeTypedNilValue
                ? nilReceiverTypeName(actual)
                : undefined;
    if (receiverType && receiverType !== declaredReceiverType) {
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
        const embeddedReceiver = normalizeReceiverType(field.type.text);
        if (receiver === undefined || receiver === null) {
            const staticInterfaceMethod = interfaceMethodForType(embeddedReceiver.baseType, methodName, context);
            if (staticInterfaceMethod) {
                matches.push({
                    method: interfaceMethodDef(embeddedReceiver.baseType, staticInterfaceMethod),
                    receiver: new RuntimeInterfaceValue(embeddedReceiver.baseType, null)
                });
            }
            continue;
        }
        if (receiver instanceof RuntimeInterfaceValue) {
            const promoted = methodForValue(receiver, methodName, context);
            if (promoted) {
                matches.push(promoted);
                continue;
            }
            const staticInterfaceMethod = interfaceMethodForType(embeddedReceiver.baseType, methodName, context);
            if (staticInterfaceMethod) {
                matches.push({ method: interfaceMethodDef(embeddedReceiver.baseType, staticInterfaceMethod), receiver });
                continue;
            }
        }
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
    }, `field:${objectIdentityId(struct)}:${field.name}`);
}
function receiverTypeName(value) {
    if (value instanceof RuntimeInterfaceValue)
        return value.value === null ? undefined : receiverTypeName(value.value);
    if (value instanceof RuntimePointer)
        return value.typeName;
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
    const resolvedTypeText = resolveRuntimeCompositeAliases(context?.resolveImportedTypeText(typeText) ?? typeText, context);
    const mapType = parseMapTypeText(resolvedTypeText);
    if (mapType)
        return new RuntimeMap(mapType.keyType, mapType.valueType, context);
    if (parseChanTypeText(resolvedTypeText))
        return null;
    const arrayType = parseArrayOrSliceTypeText(resolvedTypeText, context);
    if (arrayType) {
        if (arrayType.length === undefined && !arrayType.inferLength)
            return new RuntimeTypedNilValue(arrayType.typeText);
        if (shouldUseSparseZeroArray(arrayType)) {
            return makeSparseZeroArray(arrayType);
        }
        const values = arrayType.length === undefined || arrayType.inferLength
            ? []
            : Array.from({ length: arrayType.length }, () => defaultValueForTypeText(arrayType.elementType, context));
        markArrayType(values, arrayType.typeText);
        return values;
    }
    const type = normalizeTypeText(resolvedTypeText);
    const typeDef = context?.typeDef(type);
    if (typeDef && prefersCompiledZeroValue(type)) {
        const struct = defaultStructValueForTypeDef(typeDef, context, type);
        return type !== typeDef.name ? new RuntimeNamedValue(type, struct) : struct;
    }
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
    if (typeDef) {
        const struct = defaultStructValueForTypeDef(typeDef, context, type);
        return type !== typeDef.name ? new RuntimeNamedValue(type, struct) : struct;
    }
    const scopedAlias = context?.scopedAliasType(type);
    if (scopedAlias && scopedAlias !== type) {
        return defaultValueForTypeText(scopedAlias, context);
    }
    const alias = context?.aliasType(type);
    if (alias && alias !== type) {
        const base = defaultValueForTypeText(alias, context);
        return new RuntimeNamedValue(type, base);
    }
    return zeroValueForMapValue(type);
}
function defaultStructValueForTypeDef(typeDef, context, typeText = typeDef.name) {
    const fields = instantiateStructFields(typeDef, typeText);
    const struct = new RuntimeStruct(typeDef.name);
    const fieldDefaultContext = typeDef.declaringContext ?? context;
    for (const field of fields) {
        struct.set(field.name, defaultValueForTypeText(field.type.text, fieldDefaultContext));
    }
    return struct;
}
function shouldUseSparseZeroArray(type) {
    return type.length !== undefined &&
        !type.inferLength &&
        type.length >= SPARSE_ZERO_ARRAY_LENGTH &&
        isSparseZeroArrayElementType(type.elementType);
}
function isSparseZeroArrayElementType(typeText) {
    const type = normalizeTypeText(typeText);
    return isIntegerType(type) || isFloatType(type) || type === "string" || type === "bool";
}
function makeSparseZeroArray(type) {
    const values = [];
    values.length = type.length ?? 0;
    markArrayType(values, type.typeText);
    arrayCapacities.set(values, values.length);
    sparseArrayDefaultElementTypes.set(values, normalizeTypeText(type.elementType));
    return values;
}
function instantiateStructFields(typeDef, typeText) {
    const bindings = typeDefinitionTypeArgumentBindings(typeDef, typeText);
    if (!bindings)
        return typeDef.fields;
    return typeDef.fields.map((field) => ({
        ...field,
        type: instantiateRuntimeTypeNodeTypeArguments(field.type, bindings)
    }));
}
function typeDefinitionTypeArgumentBindings(typeDef, typeText) {
    const params = typeDef.typeParameters ?? [];
    const args = genericTypeArguments(typeText)?.args;
    if (params.length === 0 || !args || args.length === 0)
        return undefined;
    const bindings = {};
    const importPath = runtimeQualifiedTypePackagePath(typeDef.name);
    for (const [index, name] of params.entries()) {
        const arg = args[index];
        if (arg) {
            bindings[name] = arg;
            if (importPath)
                bindings[`${importPath}.${name}`] = arg;
        }
    }
    return Object.keys(bindings).length > 0 ? bindings : undefined;
}
function runtimeQualifiedTypePackagePath(typeName) {
    const base = genericBaseTypeName(normalizeReceiverType(typeName).baseType);
    const dot = base.lastIndexOf(".");
    if (dot <= 0)
        return undefined;
    return base.slice(0, dot);
}
function runtimeTypeOwnerPackagePath(typeText) {
    const type = normalizeTypeText(typeText ?? "");
    if (!type)
        return undefined;
    if (type.startsWith("*"))
        return runtimeTypeOwnerPackagePath(type.slice(1));
    if (type.startsWith("[]"))
        return runtimeTypeOwnerPackagePath(type.slice(2));
    const arrayType = parseArrayOrSliceTypeText(type);
    if (arrayType)
        return runtimeTypeOwnerPackagePath(arrayType.elementType);
    const chanType = parseChanTypeText(type);
    if (chanType)
        return runtimeTypeOwnerPackagePath(chanType.elementType);
    const mapType = parseMapTypeText(type);
    if (mapType)
        return runtimeTypeOwnerPackagePath(mapType.valueType) ?? runtimeTypeOwnerPackagePath(mapType.keyType);
    const owner = runtimeQualifiedTypePackagePath(type);
    if (owner)
        return owner;
    const generic = genericTypeArguments(type);
    return generic?.args.map(runtimeTypeOwnerPackagePath).find((candidate) => Boolean(candidate));
}
function prefersCompiledZeroValue(type) {
    return type === "sync.Once" ||
        type === "sync.Mutex" ||
        type === "sync.RWMutex" ||
        type === "sync/atomic.Bool" ||
        type === "sync/atomic.Int32" ||
        type === "sync/atomic.Uint32" ||
        type === "sync/atomic.Uint64" ||
        type === "sync/atomic.Uintptr" ||
        /^sync\/atomic\.Pointer(?:\[[\s\S]*\])?$/.test(type);
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
        return new RuntimeNamedValue(type, atomicPointerValue(atomicPointerElementType(type)));
    return undefined;
}
function syncMapValue() {
    const entries = new Map();
    return {
        Load: hostCallable("sync.Map.Load", (args) => {
            const key = args[0] ?? null;
            const entry = entries.get(runtimeMapKeyId(key));
            return entry ? [entry.value, true] : [null, false];
        }, { tupleResult: true }),
        LoadOrStore: hostCallable("sync.Map.LoadOrStore", (args) => {
            const key = args[0] ?? null;
            const value = args[1] ?? null;
            const id = runtimeMapKeyId(key);
            const existing = entries.get(id);
            if (existing)
                return [existing.value, true];
            entries.set(id, { key, value });
            return [value, false];
        }, { tupleResult: true }),
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
function atomicPointerElementType(type) {
    const generic = genericTypeArguments(type);
    return generic?.args[0];
}
function atomicPointerValue(elementType) {
    let value = null;
    const pointerType = elementType ? `*${elementType}` : undefined;
    const prepare = (next, context, role) => pointerType ? prepareAssignableToType(next, pointerType, role, context) : next;
    const load = (context) => value === null ? null : prepare(value, context, "sync/atomic.Pointer.Load result");
    return {
        Load: hostCallable("sync/atomic.Pointer.Load", (_args, context) => load(context)),
        Store: hostCallable("sync/atomic.Pointer.Store", (args, context) => {
            value = prepare(args[0] ?? null, context, "sync/atomic.Pointer.Store argument");
            return null;
        }),
        Swap: hostCallable("sync/atomic.Pointer.Swap", (args, context) => {
            const previous = load(context);
            value = prepare(args[0] ?? null, context, "sync/atomic.Pointer.Swap argument");
            return previous;
        }),
        CompareAndSwap: hostCallable("sync/atomic.Pointer.CompareAndSwap", (args, context) => {
            const oldValue = prepare(args[0] ?? null, context, "sync/atomic.Pointer.CompareAndSwap old argument");
            if (valueEqual(load(context), oldValue)) {
                value = prepare(args[1] ?? null, context, "sync/atomic.Pointer.CompareAndSwap new argument");
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
    const type = resolveRuntimeCompositeAliases(typeText, context);
    const sourceType = sourceTypeText ? resolveRuntimeCompositeAliases(sourceTypeText, context) : undefined;
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
    const mapType = parseMapTypeText(type);
    if (mapType) {
        if (actual === null)
            return new RuntimeTypedNilValue(type);
        if (!(actual instanceof RuntimeMap))
            throwTypeError(actual, type, "conversion");
        if (!runtimeTypeAssignableMatchInContext(actual.keyType, mapType.keyType, context) ||
            !runtimeTypeAssignableMatchInContext(actual.valueType, mapType.valueType, context)) {
            throwTypeError(actual, type, "conversion");
        }
        return actual;
    }
    if (context.typeDef(type) && actual instanceof RuntimeStruct && actual.typeName === type)
        return actual;
    if (type.startsWith("*")) {
        if (actual === null)
            return new RuntimeTypedNilValue(type);
        if (actual instanceof RuntimeTypedNilValue && isUnsafePointerType(actual.typeName))
            return new RuntimeTypedNilValue(type);
        if (sourceIsUnsafePointer && actual instanceof RuntimeTypedNilValue && actual.typeName.startsWith("*"))
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
    const reflectStructView = reinterpretReflectRtypeAsReflectStructType(pointer, targetType, context);
    if (reflectStructView)
        return reflectStructView;
    const abiReflectView = reinterpretInternalAbiTypeAsReflectRtype(pointer, targetType, context);
    if (abiReflectView)
        return abiReflectView;
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
    }, runtimePointerIdentityKey(pointer), pointer.sequenceInfo());
}
function reinterpretReflectRtypeAsReflectStructType(pointer, targetType, context) {
    if (!isReflectRtypeName(pointer.typeName, context) || !isReflectStructTypeName(targetType, context))
        return undefined;
    let structTypeView;
    const currentValue = () => {
        const rtype = structFromValue(pointer.get());
        const descriptor = rtype?.get("t");
        if (!(descriptor instanceof RuntimeStruct))
            return defaultValueForTypeText(targetType, context);
        if (!structTypeView)
            structTypeView = new RuntimeStruct(targetType);
        structTypeView.set("StructType", descriptor);
        for (const [field, value] of descriptor.orderedFields()) {
            structTypeView.set(field, value);
        }
        return structTypeView;
    };
    return new RuntimePointer(targetType, currentValue, (next) => {
        const nextStruct = structFromValue(next);
        if (!nextStruct)
            throwTypeError(next, targetType, `*${targetType}`);
        const rtype = structFromValue(pointer.get());
        const descriptor = rtype?.get("t");
        if (descriptor instanceof RuntimeStruct) {
            for (const [field, value] of nextStruct.orderedFields())
                descriptor.set(field, value);
        }
        structTypeView = nextStruct;
    }, `reflect.structType:${runtimePointerIdentityKey(pointer)}`);
}
function reinterpretInternalAbiTypeAsReflectRtype(pointer, targetType, context) {
    if (!isInternalAbiTypePointer(pointer, context) || !isReflectRtypeName(targetType, context))
        return undefined;
    let rtypeStruct;
    const currentValue = () => {
        const descriptor = pointer.get();
        if (!rtypeStruct) {
            rtypeStruct = new RuntimeStruct(targetType);
        }
        rtypeStruct.set("t", descriptor);
        return rtypeStruct;
    };
    return new RuntimePointer(targetType, currentValue, (next) => {
        const struct = structFromValue(next);
        if (!struct)
            throwTypeError(next, targetType, `*${targetType}`);
        const descriptor = struct.get("t");
        if (descriptor !== undefined)
            pointer.set(descriptor);
        rtypeStruct = struct;
    }, `reflect.rtype:${runtimePointerIdentityKey(pointer)}`);
}
function isReflectRtypeName(typeText, context) {
    const type = normalizeTypeText(typeText);
    return type === "reflect.rtype" || (type === "rtype" && context.importPath() === "reflect");
}
function isReflectStructTypeName(typeText, context) {
    const type = normalizeTypeText(typeText);
    return type === "reflect.structType" || (type === "structType" && context.importPath() === "reflect");
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
function prepareValueForParameterType(value, targetTypeText, source, context, role, sourceContext = context) {
    const targetType = targetTypeText ? normalizeTypeText(targetTypeText) : "";
    if (!targetType)
        return value;
    const interfaceType = interfaceTarget(targetType, context);
    if (interfaceType) {
        const dynamicType = expressionDynamicTypeText(source, sourceContext, value);
        return prepareInterfaceAssignment(value, targetType, interfaceType, role, context, dynamicType);
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
            return conversionType
                ? normalizeTypeText(conversionType)
                : callExpressionResultTypeText(expression, context) ?? inferredConcreteDynamicType(value);
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
        case "SelectorExpression":
            return selectorFieldTypeText(expression, context) ?? inferredConcreteDynamicType(value);
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
        case "SelectorExpression":
            return selectorFieldTypeText(expression, context) ?? inferredConcreteDynamicType(value);
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
            return conversionType
                ? normalizeTypeText(conversionType)
                : callExpressionResultTypeText(expression, context) ?? inferredConcreteDynamicType(value);
        }
        case "TypeAssertionExpression":
            return normalizeTypeText(expression.type.text);
        default:
            return inferredConcreteDynamicType(value);
    }
}
function selectorFieldTypeText(expression, context) {
    const objectType = expressionDeclaredTypeText(expression.object, context);
    return structFieldTypeText(objectType, expression.field, context);
}
function structFieldTypeText(typeText, fieldName, context, seen = new Set()) {
    if (!typeText)
        return undefined;
    const receiver = normalizeReceiverType(typeText);
    const baseType = receiver.baseType;
    if (seen.has(baseType))
        return undefined;
    seen.add(baseType);
    const typeDef = context.typeDef(baseType) ?? parseAnonymousStructTypeText(baseType);
    if (!typeDef)
        return undefined;
    const direct = typeDef.fields.find((field) => field.name === fieldName);
    if (direct)
        return normalizeTypeText(direct.type.text);
    const matches = typeDef.fields
        .filter((field) => field.embedded)
        .map((field) => structFieldTypeText(field.type.text, fieldName, context, seen))
        .filter((type) => type !== undefined);
    return matches.length === 1 ? matches[0] : undefined;
}
function callExpressionResultTypeText(expression, context, resultIndex = 0) {
    const callee = staticCalleeRuntimeValue(expression.callee, context);
    if (callee === undefined)
        return undefined;
    if (!isGoJuniorFunction(callee))
        return undefined;
    const result = callee.signature?.results[resultIndex];
    if (!result)
        return undefined;
    const packageName = callee.callContext?.importPath();
    const resultTypeText = packageName && packageName !== context.importPath()
        ? qualifyLocalRuntimeTypeName(result.type.text, packageName)
        : result.type.text;
    const typeArgumentBindings = inferRuntimeCallTypeArgumentBindings(callee, [], callTypeArguments(expression.callee, context), context);
    const typeText = normalizeTypeText(substituteTypeArgumentBindings(resultTypeText, typeArgumentBindings));
    return typeTextContainsAnyTypeParameter(typeText, callee.declaration?.typeParameters ?? []) ? undefined : typeText;
}
function typeTextContainsAnyTypeParameter(typeText, names) {
    if (names.length === 0)
        return false;
    return names.some((name) => new RegExp(`(^|[^A-Za-z0-9_])${escapeRegExp(name)}([^A-Za-z0-9_]|$)`).test(typeText));
}
function escapeRegExp(text) {
    return text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
function staticCalleeRuntimeValue(expression, context) {
    try {
        if (expression.kind === "Identifier")
            return context.lookup(expression.name);
        if (expression.kind === "IndexExpression")
            return staticCalleeRuntimeValue(expression.object, context);
        if (expression.kind === "SelectorExpression") {
            const object = staticCalleeRuntimeValue(expression.object, context);
            if (object === undefined)
                return undefined;
            const runtimeObject = unwrapNamed(dereferenceIfPointer(object));
            if (isRuntimeObject(runtimeObject))
                return runtimeObject[expression.field];
        }
    }
    catch {
        return undefined;
    }
    return undefined;
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
    const type = resolveRuntimeCompositeAliases(context?.resolveImportedTypeText(typeText) ?? typeText, context);
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
        if (value.typeName === type || isNilAssignableType(type) || runtimeTypeAssignableMatchInContext(value.typeName, type, context))
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
        if (!runtimeTypeAssignableMatchInContext(value.keyType, mapType.keyType, context) ||
            !runtimeTypeAssignableMatchInContext(value.valueType, mapType.valueType, context)) {
            throwTypeError(value, type, role);
        }
        return value;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        if (!(value instanceof RuntimeChannel) || !runtimeTypeAssignableMatchInContext(value.elementType, chanType.elementType, context)) {
            throwTypeError(value, type, role);
        }
        return value;
    }
    if (type.startsWith("*")) {
        const targetType = type.slice(1);
        if (!(value instanceof RuntimePointer) || !runtimeTypeAssignableMatchInContext(value.typeName, targetType, context))
            throwTypeError(value, type, role);
        return value.typeName === targetType ? value : retagRuntimePointer(value, targetType);
    }
    const structType = context?.typeDef(type);
    if (structType) {
        if (!(value instanceof RuntimeStruct) || !runtimeTypeAssignableMatchInContext(value.typeName, structType.name, context)) {
            throwTypeError(value, type, role);
        }
        return value.typeName === type ? value : retagRuntimeStruct(value, type);
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
        if (sourceInterface && interfaceDefinitionImplementsInterface(sourceInterface, interfaceType, context)) {
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
function interfaceDefinitionImplementsInterface(source, target, context) {
    for (const method of target.methods) {
        const candidate = source.methods.find((sourceMethod) => sourceMethod.name === method.name);
        if (!candidate || !signaturesCompatible(candidate.signature, method.signature, context))
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
        if (!signaturesCompatible(candidate.method.declaration.signature, method.signature, context))
            return false;
    }
    return true;
}
function runtimeTypeTextImplementsInterface(typeText, interfaceType, context) {
    const sourceInterface = interfaceTarget(typeText, context);
    if (sourceInterface)
        return interfaceDefinitionImplementsInterface(sourceInterface, interfaceType, context);
    return valueImplementsInterface(defaultValueForTypeText(typeText, context), interfaceType, context);
}
function signaturesCompatible(actual, expected, context) {
    if (actual.parameters.length !== expected.parameters.length || actual.results.length !== expected.results.length)
        return false;
    return actual.parameters.every((param, index) => {
        const expectedParam = expected.parameters[index];
        if (!expectedParam)
            return false;
        return signatureTypeCompatible(param.type.text, expectedParam.type.text, context) &&
            param.variadic === expectedParam.variadic;
    }) && actual.results.every((result, index) => {
        const expectedResult = expected.results[index];
        if (!expectedResult)
            return false;
        return signatureTypeCompatible(result.type.text, expectedResult.type.text, context) &&
            result.variadic === expectedResult.variadic;
    });
}
function signatureTypeCompatible(actual, expected, context) {
    return runtimeTypeAssignableMatch(resolveRuntimeSignatureAliasTypeText(actual, context), resolveRuntimeSignatureAliasTypeText(expected, context));
}
function resolveRuntimeAliasTypeText(typeText, context) {
    return runtimeAliasTypeChain(typeText, context).at(-1) ?? normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
}
function runtimeAliasTypeChain(typeText, context) {
    const chain = [];
    let type = normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
    const seen = new Set();
    while (context?.aliasType(type) && !seen.has(type)) {
        seen.add(type);
        type = normalizeTypeText(context.resolveImportedTypeText(context.aliasType(type) ?? type));
        chain.push(type);
    }
    return chain;
}
function resolveRuntimeSignatureAliasTypeText(typeText, context) {
    const raw = context?.resolveImportedTypeText(typeText) ?? typeText;
    if (/^\s*(?:struct|interface)\s*\{/.test(raw) || /^\s*func\s*\(/.test(raw))
        return raw.trim();
    const type = normalizeTypeText(raw);
    const aliasChain = runtimeTrueAliasTypeChain(type, context);
    if (aliasChain.length > 0)
        return resolveRuntimeSignatureAliasTypeText(aliasChain[aliasChain.length - 1], context);
    if (type.startsWith("*"))
        return `*${resolveRuntimeSignatureAliasTypeText(type.slice(1), context)}`;
    if (type.startsWith("[]"))
        return `[]${resolveRuntimeSignatureAliasTypeText(type.slice(2), context)}`;
    const arrayType = parseArrayOrSliceTypeText(type, context);
    if (arrayType?.length !== undefined && !arrayType.inferLength) {
        return `[${arrayType.length}]${resolveRuntimeSignatureAliasTypeText(arrayType.elementType, context)}`;
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${resolveRuntimeSignatureAliasTypeText(mapType.keyType, context)}]${resolveRuntimeSignatureAliasTypeText(mapType.valueType, context)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const value = resolveRuntimeSignatureAliasTypeText(chanType.elementType, context);
        if (chanType.direction === "receive")
            return `<-chan ${value}`;
        if (chanType.direction === "send")
            return `chan<- ${value}`;
        return `chan ${value}`;
    }
    const generic = genericTypeArguments(type);
    if (generic) {
        return `${resolveRuntimeSignatureAliasTypeText(generic.base, context)}[${generic.args.map((arg) => resolveRuntimeSignatureAliasTypeText(arg, context)).join(", ")}]`;
    }
    return type;
}
function runtimeTrueAliasTypeChain(typeText, context) {
    const chain = [];
    let type = normalizeTypeText(context?.resolveImportedTypeText(typeText) ?? typeText);
    const seen = new Set();
    while (context?.trueAliasType(type) && !seen.has(type)) {
        seen.add(type);
        type = normalizeTypeText(context.resolveImportedTypeText(context.trueAliasType(type) ?? type));
        chain.push(type);
    }
    return chain;
}
function resolveRuntimeCompositeAliases(typeText, context) {
    const raw = context?.resolveImportedTypeText(typeText) ?? typeText;
    if (/^\s*(?:struct|interface)\s*\{/.test(raw) || /^\s*func\s*\(/.test(raw))
        return raw.trim();
    const type = normalizeTypeText(raw);
    const alias = context?.scopedAliasType(type);
    if (alias && alias !== type)
        return resolveRuntimeCompositeAliases(alias, context);
    if (type.startsWith("*"))
        return `*${resolveRuntimeCompositeAliases(type.slice(1), context)}`;
    if (type.startsWith("[]"))
        return `[]${resolveRuntimeCompositeAliases(type.slice(2), context)}`;
    const arrayType = parseArrayOrSliceTypeText(type, context);
    if (arrayType?.length !== undefined && !arrayType.inferLength) {
        return `[${arrayType.length}]${resolveRuntimeCompositeAliases(arrayType.elementType, context)}`;
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${resolveRuntimeCompositeAliases(mapType.keyType, context)}]${resolveRuntimeCompositeAliases(mapType.valueType, context)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const value = resolveRuntimeCompositeAliases(chanType.elementType, context);
        if (chanType.direction === "receive")
            return `<-chan ${value}`;
        if (chanType.direction === "send")
            return `chan<- ${value}`;
        return `chan ${value}`;
    }
    return type;
}
function runtimeValueMatchesType(value, typeText, context) {
    const type = resolveRuntimeTrueAliases(resolveRuntimeCompositeAliases(context?.resolveImportedTypeText(typeText) ?? typeText, context), context);
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
        return runtimeTypeIdentityMatchInContext(value.typeName, type, context);
    }
    if (value instanceof RuntimeTypedNilValue) {
        if (isUnsafePointerType(type))
            return isUnsafePointerType(value.typeName) || value.typeName.startsWith("*");
        if (type.startsWith("*"))
            return runtimeTypeIdentityMatchInContext(value.typeName, type, context);
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
        return value instanceof RuntimePointer && runtimeTypeIdentityMatchInContext(value.typeName, type.slice(1), context);
    if (value instanceof RuntimeStruct)
        return runtimeTypeIdentityMatchInContext(value.typeName, type, context);
    if (value instanceof RuntimeMap) {
        const mapType = parseMapTypeText(type);
        return Boolean(mapType &&
            runtimeTypeIdentityMatchInContext(value.keyType, mapType.keyType, context) &&
            runtimeTypeIdentityMatchInContext(value.valueType, mapType.valueType, context));
    }
    if (value instanceof RuntimeChannel) {
        const chanType = parseChanTypeText(type);
        return Boolean(chanType && runtimeTypeIdentityMatchInContext(value.elementType, chanType.elementType, context));
    }
    if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type))
        return Array.isArray(value);
    if (type.startsWith("func("))
        return isRuntimeCallable(value) || isGoJuniorFunction(value);
    return false;
}
function canonicalRuntimeTypeName(type) {
    const normalized = normalizeTypeText(type);
    if (normalized.startsWith("*"))
        return `*${canonicalRuntimeTypeName(normalized.slice(1))}`;
    const arrayType = parseArrayOrSliceTypeText(normalized);
    if (arrayType) {
        const element = canonicalRuntimeTypeName(arrayType.elementType);
        return arrayType.length === undefined || arrayType.inferLength
            ? `[]${element}`
            : `[${arrayType.length}]${element}`;
    }
    const mapType = parseMapTypeText(normalized);
    if (mapType)
        return `map[${canonicalRuntimeTypeName(mapType.keyType)}]${canonicalRuntimeTypeName(mapType.valueType)}`;
    const chanType = parseChanTypeText(normalized);
    if (chanType) {
        const element = canonicalRuntimeTypeName(chanType.elementType);
        if (chanType.direction === "receive")
            return `<-chan ${element}`;
        if (chanType.direction === "send")
            return `chan<- ${element}`;
        return `chan ${element}`;
    }
    const generic = genericTypeArguments(normalized);
    if (generic) {
        return `${generic.base}[${generic.args.map(canonicalRuntimeTypeName).join(", ")}]`;
    }
    switch (normalized) {
        case "byte":
            return "uint8";
        case "rune":
            return "int32";
        default:
            return normalized;
    }
}
function runtimeTypeAssignableMatch(actual, expected) {
    return assignmentRuntimeTypeName(actual) === assignmentRuntimeTypeName(expected);
}
function runtimeTypeAssignableMatchInContext(actual, expected, context) {
    return runtimeTypeAssignableMatch(runtimeTypeNameInContext(actual, context), runtimeTypeNameInContext(expected, context));
}
function runtimeTypeIdentityMatchInContext(actual, expected, context) {
    return canonicalRuntimeTypeName(runtimeTypeNameInContext(actual, context)) ===
        canonicalRuntimeTypeName(runtimeTypeNameInContext(expected, context));
}
function retagRuntimePointer(pointer, typeName) {
    return new RuntimePointer(typeName, () => pointer.get(), (next) => pointer.set(next), runtimePointerIdentityKey(pointer), pointer.sequenceInfo());
}
function retagRuntimeStruct(struct, typeName) {
    return new RuntimeStruct(typeName, struct.orderedFields());
}
function runtimeTypeNameInContext(typeText, context) {
    const resolved = resolveRuntimeTrueAliases(resolveRuntimeCompositeAliases(typeText, context), context);
    const importPath = context?.importPath();
    const localized = importPath ? unqualifyLocalRuntimeTypeName(resolved, importPath) : resolved;
    return resolveRuntimeTrueAliases(localized, context);
}
function resolveRuntimeTrueAliases(typeText, context, seen = new Set()) {
    const type = normalizeTypeText(typeText);
    if (!type || !context)
        return type;
    if (seen.has(type))
        return type;
    seen.add(type);
    const alias = context.trueAliasType(type);
    if (alias && alias !== type) {
        return resolveRuntimeTrueAliases(context.resolveImportedTypeText(alias), context, seen);
    }
    if (type.startsWith("*"))
        return `*${resolveRuntimeTrueAliases(type.slice(1), context, new Set(seen))}`;
    if (type.startsWith("[]"))
        return `[]${resolveRuntimeTrueAliases(type.slice(2), context, new Set(seen))}`;
    const arrayType = parseArrayOrSliceTypeText(type);
    if (arrayType) {
        const element = resolveRuntimeTrueAliases(arrayType.elementType, context, new Set(seen));
        return arrayType.length === undefined || arrayType.inferLength ? `[]${element}` : `[${arrayType.length}]${element}`;
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${resolveRuntimeTrueAliases(mapType.keyType, context, new Set(seen))}]${resolveRuntimeTrueAliases(mapType.valueType, context, new Set(seen))}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const element = resolveRuntimeTrueAliases(chanType.elementType, context, new Set(seen));
        if (chanType.direction === "receive")
            return `<-chan ${element}`;
        if (chanType.direction === "send")
            return `chan<- ${element}`;
        return `chan ${element}`;
    }
    const generic = genericTypeArguments(type);
    if (generic) {
        return `${resolveRuntimeTrueAliases(generic.base, context, new Set(seen))}[${generic.args.map((arg) => resolveRuntimeTrueAliases(arg, context, new Set(seen))).join(", ")}]`;
    }
    return type;
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
    return genericBaseTypeName(normalized);
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
function arrayPointerSequence(value, index) {
    const view = arrayViews.get(value);
    return {
        values: view?.source ?? value,
        index: (view?.offset ?? 0) + index
    };
}
function arrayElementTypeText(value) {
    return parseArrayOrSliceTypeText(arrayTypeTexts.get(value) ?? "")?.elementType;
}
function getArrayElement(value, index) {
    const view = arrayViews.get(value);
    if (view)
        return getArrayElement(view.source, view.offset + index);
    if (Object.prototype.hasOwnProperty.call(value, index))
        return value[index] ?? null;
    const defaultType = sparseArrayDefaultElementTypes.get(value);
    if (defaultType)
        return zeroValueForMapValue(defaultType);
    return null;
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
    const key = normalizeTypeText(typeText);
    const cached = anonymousStructTypeCache.get(key);
    if (cached)
        return cached;
    const match = /^struct\s*\{([\s\S]*)\}$/.exec(typeText.trim());
    if (!match)
        return undefined;
    const name = typeText.trim();
    const body = match[1]?.trim() ?? "";
    const fields = body === ""
        ? []
        : splitTopLevel(body, ";").flatMap(parseAnonymousStructField);
    const typeDef = { name, fields };
    anonymousStructTypeCache.set(key, typeDef);
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
    const generic = genericTypeArguments(base);
    const named = generic?.base ?? base;
    const selector = named.lastIndexOf(".");
    return selector >= 0 ? named.slice(selector + 1) : named;
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
    if (Array.isArray(value)) {
        const mapped = value.map(externalizeRuntimeValue);
        return isTupleValues(value) ? markTupleValues(mapped) : mapped;
    }
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
        const signature = value.signature
            ? qualifyRuntimeSignature(value.signature, packageName, functionDeclarationRuntimeTypeParameters(value.declaration))
            : undefined;
        return {
            ...value,
            ...(signature ? { signature } : {}),
            async call(args, context, typeArguments) {
                const result = await value.call(args, context, typeArguments);
                return value.callContext ? result : externalizePackageRuntimeValue(result, packageName);
            }
        };
    }
    if (value instanceof RuntimeNamedValue && isUnsafePointerType(value.typeName)) {
        return new RuntimeNamedValue(value.typeName, externalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeNamedValue) {
        return new RuntimeNamedValue(qualifyLocalRuntimeTypeName(value.typeName, packageName), externalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeInterfaceValue) {
        return new RuntimeInterfaceValue(qualifyLocalRuntimeTypeName(value.interfaceType, packageName), value.value === null ? null : externalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimePointer) {
        return new RuntimePointer(qualifyLocalRuntimeTypeName(value.typeName, packageName), () => externalizePackageRuntimeValue(value.get(), packageName), (next) => value.set(internalizePackageRuntimeValue(next, packageName)), runtimePointerIdentityKey(value), value.sequenceInfo());
    }
    if (Array.isArray(value)) {
        const mapped = value.map((item) => externalizePackageRuntimeValue(item, packageName));
        return isTupleValues(value) ? markTupleValues(mapped) : mapped;
    }
    if (value instanceof RuntimeMap)
        return new RuntimeExternalizedMap(value, packageName);
    if (value instanceof RuntimeStruct) {
        return new RuntimeExternalizedStruct(value, packageName);
    }
    return value;
}
function internalizePackageRuntimeValue(value, packageName) {
    if (value instanceof RuntimeGoString)
        return value;
    if (isGoJuniorFunction(value) || isRuntimeCallable(value) || value instanceof SheetBinding)
        return value;
    if (value instanceof RuntimeNamedValue && isUnsafePointerType(value.typeName)) {
        return new RuntimeNamedValue(value.typeName, internalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeNamedValue) {
        return new RuntimeNamedValue(unqualifyLocalRuntimeTypeName(value.typeName, packageName), internalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeInterfaceValue) {
        return new RuntimeInterfaceValue(unqualifyLocalRuntimeTypeName(value.interfaceType, packageName), value.value === null ? null : internalizePackageRuntimeValue(value.value, packageName));
    }
    if (value instanceof RuntimeTypedNilValue) {
        return new RuntimeTypedNilValue(unqualifyLocalRuntimeTypeName(value.typeName, packageName));
    }
    if (value instanceof RuntimePointer) {
        return new RuntimePointer(unqualifyLocalRuntimeTypeName(value.typeName, packageName), () => internalizePackageRuntimeValue(value.get(), packageName), (next) => value.set(externalizePackageRuntimeValue(next, packageName)), runtimePointerIdentityKey(value), value.sequenceInfo());
    }
    if (Array.isArray(value)) {
        const mapped = value.map((item) => internalizePackageRuntimeValue(item, packageName));
        return isTupleValues(value) ? markTupleValues(mapped) : mapped;
    }
    if (value instanceof RuntimeMap)
        return new RuntimeInternalizedMap(value, packageName);
    if (value instanceof RuntimeStruct)
        return new RuntimeInternalizedStruct(value, packageName);
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
function qualifyLocalRuntimeTypeName(typeName, packageName, localTypeParameters = new Set()) {
    const type = normalizeTypeText(typeName);
    if (localTypeParameters.has(type))
        return type;
    const generic = genericTypeArguments(type);
    if (!type || !packageName || isPredeclaredType(type) || (!generic && /^[A-Za-z_]\w*\.[A-Za-z_]\w*$/.test(type)))
        return type;
    if (type.startsWith("*"))
        return `*${qualifyLocalRuntimeTypeName(type.slice(1), packageName, localTypeParameters)}`;
    if (type.startsWith("[]"))
        return `[]${qualifyLocalRuntimeTypeName(type.slice(2), packageName, localTypeParameters)}`;
    const array = /^\[([0-9.]*)\]([\s\S]+)$/.exec(type);
    if (array)
        return `[${array[1] ?? ""}]${qualifyLocalRuntimeTypeName(array[2] ?? "", packageName, localTypeParameters)}`;
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${qualifyLocalRuntimeTypeName(mapType.keyType, packageName, localTypeParameters)}]${qualifyLocalRuntimeTypeName(mapType.valueType, packageName, localTypeParameters)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const prefix = chanType.direction === "receive" ? "<-chan " : chanType.direction === "send" ? "chan<- " : "chan ";
        return `${prefix}${qualifyLocalRuntimeTypeName(chanType.elementType, packageName, localTypeParameters)}`;
    }
    if (generic) {
        return `${qualifyLocalRuntimeTypeName(generic.base, packageName, localTypeParameters)}[${generic.args.map((arg) => qualifyLocalRuntimeTypeName(arg, packageName, localTypeParameters)).join(", ")}]`;
    }
    if (type.startsWith("func(") || type.startsWith("interface{") || type.startsWith("struct{"))
        return type;
    return /^[A-Za-z_]\w*$/.test(type) ? `${packageName}.${type}` : type;
}
function unqualifyLocalRuntimeTypeName(typeName, packageName) {
    const type = normalizeTypeText(typeName);
    if (!type || !packageName || isPredeclaredType(type))
        return type;
    if (type.startsWith("*"))
        return `*${unqualifyLocalRuntimeTypeName(type.slice(1), packageName)}`;
    if (type.startsWith("[]"))
        return `[]${unqualifyLocalRuntimeTypeName(type.slice(2), packageName)}`;
    const array = /^\[([0-9.]*)\]([\s\S]+)$/.exec(type);
    if (array)
        return `[${array[1] ?? ""}]${unqualifyLocalRuntimeTypeName(array[2] ?? "", packageName)}`;
    const mapType = parseMapTypeText(type);
    if (mapType) {
        return `map[${unqualifyLocalRuntimeTypeName(mapType.keyType, packageName)}]${unqualifyLocalRuntimeTypeName(mapType.valueType, packageName)}`;
    }
    const chanType = parseChanTypeText(type);
    if (chanType) {
        const prefix = chanType.direction === "receive" ? "<-chan " : chanType.direction === "send" ? "chan<- " : "chan ";
        return `${prefix}${unqualifyLocalRuntimeTypeName(chanType.elementType, packageName)}`;
    }
    const generic = genericTypeArguments(type);
    if (generic) {
        return `${unqualifyLocalRuntimeTypeName(generic.base, packageName)}[${generic.args.map((arg) => unqualifyLocalRuntimeTypeName(arg, packageName)).join(", ")}]`;
    }
    const prefix = `${packageName}.`;
    if (type.startsWith(prefix))
        return type.slice(prefix.length);
    return type;
}
function qualifyRuntimeSignature(signature, packageName, typeParameters) {
    const localTypeParameters = new Set(typeParameters ?? []);
    return {
        parameters: signature.parameters.map((parameter) => ({
            ...parameter,
            type: { text: qualifyLocalRuntimeTypeName(parameter.type.text, packageName, localTypeParameters) }
        })),
        results: signature.results.map((result) => ({
            ...result,
            type: { text: qualifyLocalRuntimeTypeName(result.type.text, packageName, localTypeParameters) }
        }))
    };
}
function qualifyRuntimeStructTypeDef(typeDef, packageName) {
    const localTypeParameters = new Set(typeDef.typeParameters ?? []);
    return {
        name: qualifyLocalRuntimeTypeName(typeDef.name, packageName),
        ...(typeDef.typeParameters ? { typeParameters: typeDef.typeParameters } : {}),
        ...(typeDef.declaringContext ? { declaringContext: typeDef.declaringContext } : {}),
        fields: typeDef.fields.map((field) => ({
            ...field,
            type: { text: qualifyLocalRuntimeTypeName(field.type.text, packageName, localTypeParameters) }
        }))
    };
}
function qualifyRuntimeInterfaceDef(interfaceDef, packageName) {
    return {
        name: qualifyLocalRuntimeTypeName(interfaceDef.name, packageName),
        methods: interfaceDef.methods.map((method) => ({
            ...method,
            signature: qualifyRuntimeSignature(method.signature, packageName)
        })),
        embeds: interfaceDef.embeds.map((embed) => ({ text: qualifyLocalRuntimeTypeName(embed.text, packageName) }))
    };
}
function qualifyRuntimeMethodDeclaration(declaration, packageName) {
    const localTypeParameters = functionDeclarationRuntimeTypeParameters(declaration);
    return {
        ...declaration,
        signature: qualifyRuntimeSignature(declaration.signature, packageName, localTypeParameters),
        ...(declaration.receiver ? {
            receiver: {
                ...declaration.receiver,
                type: { text: qualifyLocalRuntimeTypeName(declaration.receiver.type.text, packageName, localTypeParameters) }
            }
        } : {})
    };
}
function functionDeclarationRuntimeTypeParameters(declaration) {
    const names = new Set(declaration?.typeParameters ?? []);
    const receiver = declaration?.receiver ? normalizeReceiverType(declaration.receiver.type.text).baseType : "";
    const generic = genericTypeArguments(receiver);
    for (const arg of generic?.args ?? []) {
        const name = normalizeTypeText(arg);
        if (/^[A-Za-z_]\w*$/.test(name))
            names.add(name);
    }
    return names;
}
function genericBaseTypeName(typeText) {
    const type = normalizeTypeText(typeText);
    if (type.startsWith("[") || type.startsWith("map[") || !type.endsWith("]"))
        return type;
    const bracket = type.indexOf("[");
    return bracket > 0 ? type.slice(0, bracket) : type;
}
function packageQualifiedRuntimeTypeNameParts(typeText) {
    const baseName = genericBaseTypeName(typeText);
    const dot = baseName.lastIndexOf(".");
    if (dot <= 0 || dot === baseName.length - 1)
        return undefined;
    return {
        importPath: baseName.slice(0, dot),
        localName: baseName.slice(dot + 1),
        baseName
    };
}
function isTypeArgumentExpression(expression, context, allowPointer = false) {
    const typeText = typeArgumentText(expression, allowPointer);
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
function typeArgumentText(expression, allowPointer = false) {
    switch (expression.kind) {
        case "TypeExpression":
            return expression.type.text;
        case "Identifier":
            return expression.name;
        case "SelectorExpression": {
            const object = typeArgumentText(expression.object, allowPointer);
            return object ? `${object}.${expression.field}` : undefined;
        }
        case "IndexExpression": {
            const object = typeArgumentText(expression.object, allowPointer);
            const index = typeArgumentText(expression.index, allowPointer);
            return object && index ? `${object}[${index}]` : undefined;
        }
        case "UnaryExpression": {
            if (!allowPointer || expression.operator !== "*")
                return undefined;
            const operand = typeArgumentText(expression.operand, allowPointer);
            return operand ? `*${operand}` : undefined;
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
    if (actual instanceof RuntimePointer)
        return `ptr:${runtimePointerIdentityKey(actual)}`;
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
function runtimePointerIdentityKey(pointer) {
    return pointer.identityKey() ?? `ptr-object:${objectIdentityId(pointer)}`;
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
    left = unwrapNamed(left);
    right = unwrapNamed(right);
    if (left instanceof RuntimeTypedNilValue || right instanceof RuntimeTypedNilValue) {
        if (left === null || right === null)
            return true;
        return left instanceof RuntimeTypedNilValue &&
            right instanceof RuntimeTypedNilValue &&
            (left.typeName === right.typeName || (isNilPointerType(left.typeName) && isNilPointerType(right.typeName)));
    }
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
    if (left instanceof RuntimePointer && right instanceof RuntimePointer) {
        return runtimePointerIdentityKey(left) === runtimePointerIdentityKey(right);
    }
    return left === right;
}
function isNilPointerType(typeName) {
    return typeName.startsWith("*") || isUnsafePointerType(typeName);
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
    if (source instanceof RuntimeTypedNilValue) {
        if (parseArrayOrSliceTypeText(source.typeName, context) || parseMapTypeText(source.typeName))
            return [];
    }
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
function iterPull(seq, context, arity, typeArguments) {
    const source = unwrapNamed(seq);
    if (!isRuntimeCallable(source) && !isGoJuniorFunction(source)) {
        throw new GoJuniorRuntimeError(`iter.Pull expects an iterator function, got ${formatValue(seq)}`);
    }
    const name = arity === 1 ? "iter.Pull" : "iter.Pull2";
    const producerContext = context.fork();
    let started = false;
    let done = false;
    let stopped = false;
    let producer;
    let producerError;
    let waitingNext;
    let resumeYield;
    let lastYielded = [];
    const doneValues = () => {
        const values = Array.from({ length: arity }, (_item, index) => iterPullZeroValue(index, typeArguments, lastYielded, context));
        values.push(false);
        return values;
    };
    const finishDone = () => {
        done = true;
        const waiter = waitingNext;
        waitingNext = undefined;
        waiter?.resolve(doneValues());
    };
    const finishError = (error) => {
        producerError = normalizeAsyncRuntimeError(error);
        done = true;
        const waiter = waitingNext;
        waitingNext = undefined;
        waiter?.reject(producerError);
    };
    const yieldFn = hostCallable(`${name}.yield`, async (args) => {
        if (stopped || done)
            return false;
        const waiter = waitingNext;
        if (!waiter) {
            throw new GoJuniorPanic(`${name}: yield called before next`);
        }
        waitingNext = undefined;
        lastYielded = args.slice(0, arity).map((arg) => arg ?? null);
        waiter.resolve([...lastYielded, true]);
        return await new Promise((resolve) => {
            resumeYield = resolve;
        });
    });
    const runProducer = async () => {
        try {
            await callRuntime(source, [yieldFn], producerContext);
            finishDone();
        }
        catch (error) {
            finishError(error);
        }
    };
    const next = hostCallable(`${name}.next`, async () => {
        if (producerError)
            throw producerError;
        if (done || stopped)
            return doneValues();
        if (waitingNext) {
            throw new GoJuniorPanic(`${name}: next called before previous next returned`);
        }
        const result = new Promise((resolve, reject) => {
            waitingNext = { resolve, reject };
        });
        if (!started) {
            started = true;
            producer = runProducer();
        }
        else if (resumeYield) {
            const resume = resumeYield;
            resumeYield = undefined;
            resume(true);
        }
        return await result;
    }, { tupleResult: true });
    const stop = hostCallable(`${name}.stop`, async () => {
        if (producerError)
            throw producerError;
        if (done || stopped)
            return null;
        stopped = true;
        const waiter = waitingNext;
        waitingNext = undefined;
        waiter?.resolve(doneValues());
        if (resumeYield) {
            const resume = resumeYield;
            resumeYield = undefined;
            resume(false);
        }
        if (producer) {
            await producer;
            if (producerError)
                throw producerError;
        }
        else {
            done = true;
        }
        return null;
    });
    return [next, stop];
}
function iterPullZeroValue(index, typeArguments, lastYielded, context) {
    const typeText = typeArguments?.[index];
    if (typeText)
        return defaultValueForTypeText(typeText, context);
    return zeroValueLike(lastYielded[index] ?? null);
}
function markRuntimePackageObject(value, importPath, packageInfo) {
    runtimePackageInspections.set(value, {
        importPath,
        ...(packageInfo ? { packageInfo } : {})
    });
}
function formatRuntimePackageObject(value) {
    const inspection = runtimePackageInspections.get(value);
    if (!inspection)
        return undefined;
    const packageName = inspection.packageInfo?.Name() ?? importDefaultName(inspection.importPath);
    const lines = [`package ${packageName}`];
    const signatures = inspection.packageInfo
        ? packageInspectionSignatures(inspection.packageInfo)
        : runtimeObjectInspectionSignatures(value);
    if (signatures.length > 0)
        lines.push(...signatures);
    return lines.join("\n");
}
function packageInspectionSignatures(pkg) {
    return pkg.Scope().Names()
        .flatMap((name) => {
        const object = pkg.Scope().Lookup(name);
        if (object === null || object.constructor.name === "PkgName" || !object.Exported())
            return [];
        const signature = packageObjectSignature(object, pkg);
        return signature ? [signature] : [];
    });
}
function packageObjectSignature(object, pkg) {
    const name = object.Name();
    const typeText = goTypesObjectTypeText(object, pkg);
    switch (object.constructor.name) {
        case "Func":
            return formatPackageFunctionSignature(name, typeText);
        case "TypeName":
            return formatPackageTypeSignature(name, typeText, object, pkg);
        case "Const":
            return typeText ? `const ${name} ${typeText}` : `const ${name}`;
        case "Var":
            return typeText ? `var ${name} ${typeText}` : `var ${name}`;
        default:
            return undefined;
    }
}
function formatPackageFunctionSignature(name, typeText) {
    if (!typeText)
        return `func ${name}`;
    return typeText.startsWith("func")
        ? `func ${name}${typeText.slice("func".length)}`
        : `func ${name} ${typeText}`;
}
function formatPackageTypeSignature(name, typeText, object, pkg) {
    const header = typeText ?? name;
    const underlying = goTypesObjectUnderlyingTypeText(object, pkg);
    return underlying && underlying !== header
        ? `type ${header} ${underlying}`
        : `type ${header}`;
}
function goTypesObjectTypeText(object, pkg) {
    const type = object.Type();
    if (type === null)
        return undefined;
    return GoTypesTypeString(type, GoTypesRelativeTo(pkg));
}
function goTypesObjectUnderlyingTypeText(object, pkg) {
    const type = object.Type();
    const underlying = type?.Underlying?.();
    if (underlying === undefined || underlying === null || underlying === type)
        return undefined;
    return GoTypesTypeString(underlying, GoTypesRelativeTo(pkg));
}
function runtimeObjectInspectionSignatures(value) {
    return Object.entries(value)
        .filter(([name]) => isExportedRuntimeName(name))
        .map(([name, item]) => runtimeObjectMemberSignature(name, item));
}
function runtimeObjectMemberSignature(name, value) {
    if (isGoJuniorFunction(value) && value.signature)
        return formatPackageFunctionSignature(name, signatureTypeText(value.signature));
    if (isRuntimeCallable(value))
        return `func ${name}`;
    return `var ${name} ${runtimeTypeText(value)}`;
}
function isExportedRuntimeName(name) {
    const first = name.codePointAt(0);
    return first !== undefined && first >= 65 && first <= 90;
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
    const result = await callRuntime(boundMethodValue(method.method, method.receiver, context), [], context);
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
function newValueFormatState() {
    return { active: new Set() };
}
function withValueFormatCycleGuard(state, value, body) {
    if (state.active.has(value))
        return "<cycle>";
    state.active.add(value);
    try {
        return body();
    }
    finally {
        state.active.delete(value);
    }
}
export function formatValue(value) {
    return formatValueInner(value, newValueFormatState());
}
function formatValueInner(value, state) {
    if (value instanceof RuntimeNamedValue)
        return withValueFormatCycleGuard(state, value, () => formatValueInner(value.value, state));
    if (value instanceof RuntimeInterfaceValue) {
        return withValueFormatCycleGuard(state, value, () => value.value === null ? "<nil>" : formatValueInner(value.value, state));
    }
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
    if (Array.isArray(value)) {
        return withValueFormatCycleGuard(state, value, () => `[${value.map((item) => formatValueInner(item, state)).join(" ")}]`);
    }
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatRuntimeMap(value, state);
    if (value instanceof RuntimeStruct)
        return formatRuntimeStruct(value, state);
    if (value instanceof RuntimePointer)
        return withValueFormatCycleGuard(state, value, () => `&${formatValueInner(value.get(), state)}`);
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    const packageText = formatRuntimePackageObject(value);
    if (packageText !== undefined)
        return packageText;
    return withValueFormatCycleGuard(state, value, () => `{${Object.entries(value).map(([key, item]) => `${key}:${formatValueInner(item, state)}`).join(" ")}}`);
}
export function formatReplValue(value) {
    return formatReplValueInner(value, newValueFormatState());
}
function formatReplValueInner(value, state) {
    if (value instanceof RuntimeNamedValue)
        return withValueFormatCycleGuard(state, value, () => formatReplValueInner(value.value, state));
    if (value instanceof RuntimeInterfaceValue) {
        return withValueFormatCycleGuard(state, value, () => value.value === null ? `${value.interfaceType}(nil)` : formatReplValueInner(value.value, state));
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
    if (Array.isArray(value)) {
        return withValueFormatCycleGuard(state, value, () => `[${value.map((item) => formatReplValueInner(item, state)).join(" ")}]`);
    }
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatReplMap(value, state);
    if (value instanceof RuntimeStruct)
        return formatReplStruct(value, state);
    if (value instanceof RuntimePointer)
        return withValueFormatCycleGuard(state, value, () => `&${formatReplValueInner(value.get(), state)}`);
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    const packageText = formatRuntimePackageObject(value);
    if (packageText !== undefined)
        return packageText;
    return withValueFormatCycleGuard(state, value, () => `{${Object.entries(value).map(([key, item]) => `${key}:${formatReplValueInner(item, state)}`).join(" ")}}`);
}
function formatGoSyntaxValue(value) {
    return formatGoSyntaxValueInner(value, newValueFormatState());
}
function formatGoSyntaxValueInner(value, state) {
    if (value instanceof RuntimeNamedValue) {
        return withValueFormatCycleGuard(state, value, () => value.value instanceof RuntimeTypedNilValue
            ? `${value.typeName}(nil)`
            : `${value.typeName}(${formatGoSyntaxValueInner(value.value, state)})`);
    }
    if (value instanceof RuntimeInterfaceValue) {
        return withValueFormatCycleGuard(state, value, () => value.value === null ? `${value.interfaceType}(nil)` : formatGoSyntaxValueInner(value.value, state));
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
    if (Array.isArray(value)) {
        return withValueFormatCycleGuard(state, value, () => `[]interface{}{${value.map((item) => formatGoSyntaxValueInner(item, state)).join(", ")}}`);
    }
    if (value instanceof RuntimeChannel)
        return `chan ${value.elementType}`;
    if (value instanceof RuntimeMap)
        return formatGoSyntaxMap(value, state);
    if (value instanceof RuntimeStruct)
        return formatGoSyntaxStruct(value, state);
    if (value instanceof RuntimePointer)
        return withValueFormatCycleGuard(state, value, () => `&${formatGoSyntaxValueInner(value.get(), state)}`);
    if (isGoJuniorFunction(value))
        return formatFunctionValue(value);
    if (isRuntimeCallable(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    return withValueFormatCycleGuard(state, value, () => `map[string]interface{}{${Object.entries(value)
        .map(([key, item]) => `${JSON.stringify(key)}: ${formatGoSyntaxValueInner(item, state)}`)
        .join(", ")}}`);
}
function formatRuntimeMap(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const entries = value.orderedEntries()
            .map(([key, item]) => `${formatValueInner(key, state)}:${formatValueInner(item, state)}`)
            .join(" ");
        return `map[${value.keyType}]${value.valueType}{${entries}}`;
    });
}
function formatFunctionValue(value) {
    return value.source?.trimEnd() || `<func ${value.name}>`;
}
function formatRuntimeStruct(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const fields = value.orderedFields()
            .map(([key, item]) => `${key}:${formatValueInner(item, state)}`)
            .join(" ");
        return `${value.typeName}{${fields}}`;
    });
}
function formatReplMap(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const entries = value.orderedEntries()
            .map(([key, item]) => `${formatReplValueInner(key, state)}:${formatReplValueInner(item, state)}`)
            .join(", ");
        return `map[${value.keyType}]${value.valueType}{${entries}}`;
    });
}
function formatReplStruct(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const fields = value.orderedFields()
            .map(([key, item]) => `${key}:${formatReplValueInner(item, state)}`)
            .join(", ");
        return `${value.typeName}{${fields}}`;
    });
}
function formatReplString(value) {
    if (value.includes("\n") || value.includes("\""))
        return `\`${value.replace(/`/g, "\\`")}\``;
    return JSON.stringify(value);
}
function formatGoSyntaxMap(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const entries = value.orderedEntries()
            .map(([key, item]) => `${formatGoSyntaxValueInner(key, state)}: ${formatGoSyntaxValueInner(item, state)}`)
            .join(", ");
        return `map[${value.keyType}]${value.valueType}{${entries}}`;
    });
}
function formatGoSyntaxStruct(value, state) {
    return withValueFormatCycleGuard(state, value, () => {
        const fields = value.orderedFields()
            .map(([key, item]) => `${key}: ${formatGoSyntaxValueInner(item, state)}`)
            .join(", ");
        return `${value.typeName}{${fields}}`;
    });
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
