import { frontSourceToAst } from "./frontToAst.js";
export class GoJuniorRuntimeError extends Error {
    constructor(message) {
        super(message);
        this.name = "GoJuniorRuntimeError";
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
export class EvaluationContext {
    options;
    output = [];
    rootScope = new Scope();
    currentScope = this.rootScope;
    deferFrames = [[]];
    types = new Map();
    interfaces = new Map();
    methods = new Map();
    maxLoopIterations;
    stdout;
    constructor(options = {}) {
        this.options = options;
        this.maxLoopIterations = options.maxLoopIterations ?? 100_000;
        this.stdout = options.stdout;
        installBuiltins(this);
        installAutomaticImports(this);
    }
    declare(name, value, mutable = true, type) {
        const typeText = bindingTypeText(type);
        if (typeText)
            assertAssignableToType(value, typeText, `variable ${name}`, this);
        this.currentScope.declare(name, value, mutable, typeText);
    }
    declareRoot(name, value, mutable = true, type) {
        const typeText = bindingTypeText(type);
        if (typeText)
            assertAssignableToType(value, typeText, `variable ${name}`, this);
        this.rootScope.declare(name, value, mutable, typeText);
    }
    declareOrAssignRoot(name, value, mutable = true, type) {
        if (this.rootScope.hasLocal(name)) {
            this.rootScope.assign(name, value, this.assignmentChecker());
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
    pointerToBinding(name, typeName) {
        const scope = this.currentScope;
        return new RuntimePointer(typeName, () => scope.lookup(name), (next) => scope.assign(name, next, this.assignmentChecker()));
    }
    outputFrom(offset) {
        return this.output.slice(offset);
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
    pushDefer(callback) {
        this.currentDeferFrame().push(callback);
    }
    runDefers() {
        const frame = this.currentDeferFrame();
        while (frame.length > 0) {
            frame.pop()?.();
        }
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
    write(text) {
        this.output.push(text);
        this.stdout?.(text);
    }
    loopLimit() {
        return this.maxLoopIterations;
    }
    packages() {
        return this.options.packages ?? {};
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
        if (spec.structFields) {
            this.types.set(spec.name, {
                name: spec.name,
                fields: spec.structFields
            });
        }
        if (spec.interfaceMethods) {
            this.interfaces.set(spec.name, {
                name: spec.name,
                methods: spec.interfaceMethods
            });
        }
    }
    typeDef(name) {
        return this.types.get(name);
    }
    interfaceDef(name) {
        return this.interfaces.get(name);
    }
    registerMethod(declaration) {
        if (!declaration.receiver)
            return;
        const receiver = normalizeReceiverType(declaration.receiver.type.text);
        this.methods.set(methodKey(receiver.baseType, declaration.name), {
            declaration,
            receiverType: receiver.baseType,
            pointerReceiver: receiver.pointer
        });
    }
    methodFor(typeName, methodName) {
        return this.methods.get(methodKey(typeName, methodName));
    }
    currentDeferFrame() {
        const frame = this.deferFrames[this.deferFrames.length - 1];
        if (!frame)
            throw new GoJuniorRuntimeError("internal error: missing defer frame");
        return frame;
    }
    assignmentChecker() {
        return (value, typeText, role) => assertAssignableToType(value, typeText, role, this);
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
    assign(name, value, checker) {
        const binding = this.bindings.get(name);
        if (binding) {
            if (!binding.mutable) {
                throw new GoJuniorRuntimeError(`${name} is const`);
            }
            if (binding.typeText) {
                checker?.(value, binding.typeText, `variable ${name}`);
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
const objectMapKeyIds = new WeakMap();
const arrayCapacities = new WeakMap();
let nextObjectMapKeyId = 1;
export class RuntimeMap {
    keyType;
    valueType;
    entries = new Map();
    constructor(keyType, valueType) {
        this.keyType = keyType;
        this.valueType = valueType;
    }
    get(key) {
        const entry = this.entries.get(runtimeMapKeyId(key));
        return entry ? entry.value : zeroValueForMapValue(this.valueType);
    }
    getWithPresence(key) {
        const entry = this.entries.get(runtimeMapKeyId(key));
        return entry ? [entry.value, true] : [zeroValueForMapValue(this.valueType), false];
    }
    set(key, value) {
        assertAssignableToType(key, this.keyType, "map key");
        assertAssignableToType(value, this.valueType, "map value");
        this.entries.set(runtimeMapKeyId(key), { key, value });
    }
    orderedEntries() {
        return [...this.entries.values()].map((entry) => [entry.key, entry.value]);
    }
    size() {
        return this.entries.size;
    }
}
export function evaluateSource(source, options = {}) {
    const parsed = frontSourceToAst(source);
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
export function evaluateProgram(ast, options = {}) {
    const context = new EvaluationContext(options);
    try {
        installSheets(context, options);
        installImports(context, ast);
        for (const declaration of ast.functions) {
            installFunctionDeclaration(context, declaration);
        }
        if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
            const value = installedFunctionValue(context, ast.functions[0]);
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
        if (completion.kind !== "normal")
            throw new GoJuniorRuntimeError(completionErrorMessage(completion));
        return {
            diagnostics: ast.diagnostics,
            output: context.output,
            ast,
            ...(completion.value !== undefined ? { value: completion.value } : {})
        };
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
            diagnostics: [
                ...ast.diagnostics,
                {
                    code: error instanceof GoJuniorPanic ? "GJPANIC001" : "GOJR_RUNTIME001",
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
    options;
    context;
    constructor(options = {}) {
        this.options = options;
        this.context = new EvaluationContext(options);
        installSheets(this.context, options);
    }
    setSheet(sheet) {
        this.context.declareOrAssignRoot("sheet", new SheetBinding("sheet", sheet), true);
    }
    setSheets(sheets) {
        for (const [name, data] of Object.entries(sheets)) {
            this.context.declareOrAssignRoot(name, new SheetBinding(name, data), true);
        }
    }
    evaluate(source) {
        const parsed = frontSourceToAst(source);
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
        const outputStart = this.context.output.length;
        try {
            installImports(this.context, ast);
            for (const declaration of ast.functions) {
                installFunctionDeclaration(this.context, declaration);
            }
            if (ast.kind === "function" && ast.functions[0] && ast.body.length === 0) {
                const value = installedFunctionValue(this.context, ast.functions[0]);
                return {
                    diagnostics: ast.diagnostics,
                    output: this.context.outputFrom(outputStart),
                    ast,
                    value
                };
            }
            const completion = executeTopLevelStatements(ast.body, this.context);
            return resultFromCompletion(ast, this.context.outputFrom(outputStart), completion);
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            return {
                diagnostics: [
                    ...ast.diagnostics,
                    {
                        code: error instanceof GoJuniorPanic ? "GJPANIC001" : "GOJR_RUNTIME001",
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
function resultFromCompletion(ast, output, completion) {
    if (completion.kind === "return") {
        return {
            diagnostics: ast.diagnostics,
            output,
            ast,
            values: completion.values,
            ...(completion.values.length === 1 ? { value: completion.values[0] } : {})
        };
    }
    if (completion.kind !== "normal")
        throw new GoJuniorRuntimeError(completionErrorMessage(completion));
    return {
        diagnostics: ast.diagnostics,
        output,
        ast,
        ...(completion.value !== undefined ? { value: completion.value } : {})
    };
}
function diagnosticsLookIncomplete(diagnostics) {
    const errors = diagnostics.filter((diagnostic) => diagnostic.severity === "error");
    return errors.length > 0 && errors.every((diagnostic) => {
        if (diagnostic.code === "GJPARSE001") {
            return /found\s+-->\s*''\s*<--/.test(diagnostic.message) || /but found:\s*''/.test(diagnostic.message);
        }
        if (diagnostic.code !== "GJPARSE_FRONT001")
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
    context.declareRoot("len", hostCallable("len", (args) => {
        return BigInt(valueLength(args[0] ?? null));
    }));
    context.declareRoot("cap", hostCallable("cap", (args) => {
        return BigInt(valueCapacity(args[0] ?? null));
    }));
    context.declareRoot("append", hostCallable("append", (args) => {
        return appendValues(args[0] ?? null, args.slice(1));
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
        const defaultName = imported.path.split("/").filter(Boolean).at(-1) ?? imported.path;
        const name = imported.alias ?? defaultName;
        const pkg = packages[imported.path] ?? packages[defaultName];
        if (!pkg) {
            throw new GoJuniorRuntimeError(`package ${imported.path} is not available`);
        }
        context.declareOrAssignRoot(name, pkg, true);
    }
}
function installFunctionDeclaration(context, declaration) {
    if (declaration.receiver) {
        context.registerMethod(declaration);
        return;
    }
    context.declareOrAssignRoot(declaration.name, functionValue(declaration), true);
}
function installedFunctionValue(context, declaration) {
    return declaration.receiver ? functionValue(declaration) : context.lookup(declaration.name);
}
function bindingTypeText(type) {
    const text = typeof type === "string" ? type : type?.text;
    const normalized = text ? normalizeTypeText(text) : "";
    return normalized && normalized !== "<missing>" ? normalized : undefined;
}
function installAutomaticImports(context) {
    context.declareRoot("fmt", availablePackages(context).fmt ?? fmtPackage(), true);
}
function availablePackages(context) {
    return {
        fmt: fmtPackage(),
        ...context.packages()
    };
}
function fmtPackage() {
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
function hostCallable(name, call) {
    return {
        kind: "HostCallable",
        name,
        call
    };
}
function functionValue(declaration) {
    return goJuniorFunctionValue(declaration.name, declaration.signature, declaration.body, undefined, declaration);
}
function functionLiteralValue(expression, context) {
    return goJuniorFunctionValue("<closure>", expression.signature, expression.body, context.captureScope());
}
function executeTopLevelStatements(statements, context) {
    try {
        return executeStatements(statements, context);
    }
    finally {
        context.runDefers();
    }
}
function goJuniorFunctionValue(name, signature, body, closureScope, declaration, boundReceiver) {
    return {
        kind: "GoJuniorFunction",
        name,
        ...(declaration ? { declaration } : {}),
        call(args, parentContext) {
            const context = parentContext;
            const invoke = () => context.childScope(() => context.deferScope(() => {
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
function namedReturnValues(signature, context) {
    const namedResults = signature.results.filter((result) => result.name);
    if (namedResults.length === 0)
        return [];
    return namedResults.map((result) => result.name ? context.lookup(result.name) : defaultValueForDeclarationType(result.type, context));
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
    return goJuniorFunctionValue(method.declaration.name, method.declaration.signature, method.declaration.body, undefined, method.declaration, boundReceiver);
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
function executeStatements(statements, context) {
    const labels = statementLabels(statements);
    let lastValue;
    for (let pc = 0; pc < statements.length; pc += 1) {
        const statement = statements[pc];
        if (!statement)
            continue;
        const completion = executeStatement(statement, context);
        if (completion.kind === "goto") {
            const target = labels.get(completion.label);
            if (target !== undefined) {
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
function statementLabels(statements) {
    const labels = new Map();
    for (const [index, statement] of statements.entries()) {
        if (statement.kind === "LabeledStatement") {
            labels.set(statement.label, index);
        }
    }
    return labels;
}
function executeBlock(block, context, createScope = true) {
    if (!createScope)
        return executeStatements(block.statements, context);
    return context.childScope(() => executeStatements(block.statements, context));
}
function executeStatement(statement, context) {
    switch (statement.kind) {
        case "BlockStatement":
            return executeBlock(statement, context);
        case "LabeledStatement":
            return executeLabeledStatement(statement, context);
        case "ConstDecl":
            let previousConstValue = undefined;
            let previousConstType = undefined;
            for (const [index, declaration] of statement.declarations.entries()) {
                const valueExpression = declaration.value ?? previousConstValue;
                const type = declaration.type ?? previousConstType;
                const value = valueExpression
                    ? evaluateConstExpression(valueExpression, BigInt(index), context)
                    : defaultValueForDeclarationType(type, context);
                context.declare(declaration.name, value, false, type);
                if (declaration.value)
                    previousConstValue = declaration.value;
                if (declaration.type)
                    previousConstType = declaration.type;
            }
            return { kind: "normal" };
        case "VarDecl":
            for (const declaration of statement.declarations) {
                context.declare(declaration.name, declaration.value ? evaluateExpression(declaration.value, context) : defaultValueForDeclarationType(declaration.type, context), true, declaration.type);
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
            executeAssign(statement, context);
            return { kind: "normal" };
        case "ShortVarStatement":
            executeShortVar(statement, context);
            return { kind: "normal" };
        case "IncDecStatement":
            executeIncDec(statement, context);
            return { kind: "normal" };
        case "ExpressionStatement":
            return { kind: "normal", value: evaluateExpression(statement.expression, context) };
    }
}
function executeLabeledStatement(statement, context) {
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
function evaluateConstExpression(expression, iotaValue, context) {
    return context.childScope(() => {
        context.declare("iota", iotaValue, false, "int64");
        return evaluateExpression(expression, context);
    });
}
function executeIf(statement, context) {
    if (toBool(evaluateExpression(statement.condition, context))) {
        return executeBlock(statement.thenBlock, context);
    }
    if (!statement.elseBranch)
        return { kind: "normal" };
    return statement.elseBranch.kind === "IfStatement"
        ? executeIf(statement.elseBranch, context)
        : executeBlock(statement.elseBranch, context);
}
function executeSwitch(statement, context, label) {
    if (statement.typeSwitch)
        return executeTypeSwitch(statement, context, label);
    const switchValue = statement.expression ? evaluateExpression(statement.expression, context) : true;
    let matched = false;
    for (const clause of statement.clauses) {
        if (!matched) {
            matched = clause.default || clause.values.some((value) => valueEqual(switchValue, evaluateExpression(value, context)));
        }
        if (!matched)
            continue;
        const completion = executeStatements(clause.statements, context);
        if (completion.kind === "fallthrough") {
            matched = true;
            continue;
        }
        if (completion.kind === "break" && labelMatches(completion.label, label))
            return { kind: "normal" };
        return completion;
    }
    return { kind: "normal" };
}
function executeTypeSwitch(statement, context, label) {
    if (!statement.typeSwitch)
        return { kind: "normal" };
    const switchValue = evaluateExpression(statement.typeSwitch.expression, context);
    for (const clause of statement.clauses) {
        const matched = clause.default || (clause.typeValues ?? []).some((type) => runtimeValueMatchesType(switchValue, type.text));
        if (!matched)
            continue;
        const completion = context.childScope(() => {
            if (statement.typeSwitch?.name) {
                if (statement.typeSwitch.define) {
                    context.declare(statement.typeSwitch.name, switchValue, true);
                }
                else {
                    context.assign(statement.typeSwitch.name, switchValue);
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
function executeFor(statement, context, label) {
    if (statement.range) {
        const source = evaluateExpression(statement.range.source, context);
        const entries = rangeEntries(source);
        for (const [index, value] of entries) {
            const completion = context.childScope(() => {
                if (statement.range?.keyName) {
                    if (statement.range.define)
                        context.declare(statement.range.keyName, index, true, inferredTypeText(index));
                    else
                        context.assign(statement.range.keyName, index);
                }
                if (statement.range?.valueName) {
                    if (statement.range.define)
                        context.declare(statement.range.valueName, value, true, inferredTypeText(value));
                    else
                        context.assign(statement.range.valueName, value);
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
    return context.childScope(() => {
        if (statement.init) {
            expectNormalCompletion(executeStatement(statement.init, context), "for init statement");
        }
        for (let iteration = 0; iteration < context.loopLimit(); iteration += 1) {
            if (statement.condition && !toBool(evaluateExpression(statement.condition, context))) {
                return { kind: "normal" };
            }
            const completion = executeBlock(statement.body, context);
            if (completion.kind === "break" && labelMatches(completion.label, label))
                return { kind: "normal" };
            if (completion.kind === "continue" && labelMatches(completion.label, label)) {
                if (statement.post) {
                    expectNormalCompletion(executeStatement(statement.post, context), "for post statement");
                }
                continue;
            }
            if (completion.kind !== "normal")
                return completion;
            if (statement.post) {
                expectNormalCompletion(executeStatement(statement.post, context), "for post statement");
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
function executeDefer(statement, context) {
    if (statement.expression.kind === "CallExpression") {
        const callee = evaluateExpression(statement.expression.callee, context);
        const args = statement.expression.args.map((arg) => evaluateExpression(arg, context));
        context.pushDefer(() => callRuntime(callee, args, context));
        return;
    }
    context.pushDefer(() => evaluateExpression(statement.expression, context));
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
function executeAssign(statement, context) {
    const values = evaluateAssignmentValues(statement.values, statement.targets.length, context);
    if (values.length !== statement.targets.length) {
        throw new GoJuniorRuntimeError(`assignment count mismatch: ${statement.targets.length} targets but ${values.length} values`);
    }
    for (const [index, target] of statement.targets.entries()) {
        assignExpressionTarget(target, values[index] ?? null, context);
    }
}
function executeShortVar(statement, context) {
    const values = evaluateAssignmentValues(statement.values, statement.names.length, context);
    if (values.length !== statement.names.length) {
        throw new GoJuniorRuntimeError(`short declaration count mismatch: ${statement.names.length} names but ${values.length} values`);
    }
    for (const [index, name] of statement.names.entries()) {
        if (name === "_")
            continue;
        if (name === "<invalid>")
            throw new GoJuniorRuntimeError("non-identifier used in short declaration");
        const value = values[index] ?? null;
        context.declare(name, value, true, inferredTypeText(value));
    }
}
function evaluateAssignmentValues(expressions, targetCount, context) {
    if (targetCount === 2 && expressions.length === 1 && expressions[0]?.kind === "IndexExpression") {
        const lookup = evaluateMapLookupWithPresence(expressions[0], context);
        if (lookup)
            return lookup;
    }
    const values = expressions.map((expression) => evaluateExpression(expression, context));
    if (targetCount > 1 && values.length === 1 && Array.isArray(values[0])) {
        return values[0];
    }
    return values;
}
function evaluateMapLookupWithPresence(expression, context) {
    const object = evaluateExpression(expression.object, context);
    if (!(object instanceof RuntimeMap))
        return undefined;
    const index = evaluateExpression(expression.index, context);
    const [value, ok] = object.getWithPresence(index);
    return [value, ok];
}
function executeIncDec(statement, context) {
    const current = evaluateExpression(statement.target, context);
    const next = statement.operator === "++"
        ? addNumbers(current, 1n)
        : subtractNumbers(current, 1n);
    assignExpressionTarget(statement.target, next, context);
}
function assignExpressionTarget(target, value, context) {
    if (target.kind === "Identifier") {
        if (target.name === "_")
            return;
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
    if (target.kind === "UnaryExpression" && target.operator === "*") {
        const pointer = evaluateExpression(target.operand, context);
        if (!(pointer instanceof RuntimePointer)) {
            throw new GoJuniorRuntimeError(`${formatValue(pointer)} is not a pointer`);
        }
        pointer.set(value);
        return;
    }
    throw new GoJuniorRuntimeError("unsupported assignment target");
}
function evaluateExpression(expression, context) {
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
            return getIndex(evaluateExpression(expression.object, context), evaluateExpression(expression.index, context));
        case "SliceExpression":
            return getSlice(evaluateExpression(expression.object, context), expression.start ? evaluateExpression(expression.start, context) : undefined, expression.end ? evaluateExpression(expression.end, context) : undefined);
        case "SpreadsheetRangeExpression":
            return getSpreadsheetRange(expression, context);
    }
}
function evaluateArrayLiteral(expression, context) {
    const type = parseArrayOrSliceTypeText(expression.type.text);
    if (!type) {
        throw new GoJuniorRuntimeError(`${expression.type.text} is not an array or slice literal type`);
    }
    const values = expression.elements.map((element) => evaluateExpression(element, context));
    for (const value of values) {
        assertAssignableToType(value, type.elementType, "array element", context);
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
function makeRuntimeSlice(elementType, length, capacity, context) {
    const values = Array.from({ length }, () => defaultValueForTypeText(elementType, context));
    arrayCapacities.set(values, capacity);
    return values;
}
function evaluateStructLiteral(expression, context) {
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
        const value = evaluateExpression(field.value, context);
        assertAssignableToType(value, declared.type.text, `field ${field.name ?? declared.name}`, context);
        struct.set(declared.name, value);
    }
    return struct;
}
function evaluateMapLiteral(expression, context) {
    const map = new RuntimeMap(expression.keyType.text, expression.valueType.text);
    for (const entry of expression.entries) {
        map.set(evaluateExpression(entry.key, context), evaluateExpression(entry.value, context));
    }
    return map;
}
function evaluateIdentifier(expression, context) {
    if (expression.name === "nil")
        return null;
    return context.lookup(expression.name);
}
function evaluateUnary(expression, context) {
    switch (expression.operator) {
        case "+":
            return numericIdentity(evaluateExpression(expression.operand, context));
        case "-":
            return negateNumber(evaluateExpression(expression.operand, context));
        case "!":
            return !toBool(evaluateExpression(expression.operand, context));
        case "&":
            return pointerToExpression(expression.operand, context);
        case "*":
            return dereference(evaluateExpression(expression.operand, context));
    }
}
function evaluateBinary(expression, context) {
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
            if (typeof left === "string" || typeof right === "string")
                return `${formatValue(left)}${formatValue(right)}`;
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
function evaluateTypeAssertion(expression, context) {
    const value = evaluateExpression(expression.expression, context);
    if (!runtimeValueMatchesType(value, expression.type.text)) {
        throw new GoJuniorRuntimeError(`${formatValue(value)} does not have dynamic type ${normalizeTypeText(expression.type.text)}`);
    }
    return value;
}
function evaluateCall(expression, context) {
    if (expression.callee.kind === "Identifier" && expression.callee.name === "make") {
        return evaluateMake(expression, context);
    }
    const callee = evaluateExpression(expression.callee, context);
    const args = expression.args.map((arg) => evaluateExpression(arg, context));
    return callRuntime(callee, expression.spreadLast ? spreadLastArgument(args) : args, context);
}
function evaluateMake(expression, context) {
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
            toNonNegativeLength(evaluateExpression(expression.args[1], context), "map size hint");
        return new RuntimeMap(mapType.keyType, mapType.valueType);
    }
    const arrayType = parseArrayOrSliceTypeText(typeText);
    if (arrayType) {
        if (arrayType.length !== undefined)
            throw new GoJuniorRuntimeError(`cannot make array type ${typeText}; use a slice type`);
        if (expression.args.length < 2 || expression.args.length > 3) {
            throw new GoJuniorRuntimeError("make slice expects length and optional capacity");
        }
        const length = toNonNegativeLength(evaluateExpression(expression.args[1], context), "slice length");
        const capacity = expression.args[2]
            ? toNonNegativeLength(evaluateExpression(expression.args[2], context), "slice capacity")
            : length;
        if (capacity < length)
            throw new GoJuniorRuntimeError("slice capacity is smaller than length");
        return makeRuntimeSlice(arrayType.elementType, length, capacity, context);
    }
    throw new GoJuniorRuntimeError(`cannot make ${typeText}`);
}
function spreadLastArgument(args) {
    if (args.length === 0)
        return args;
    const last = args[args.length - 1];
    if (!Array.isArray(last)) {
        throw new GoJuniorRuntimeError(`${formatValue(last ?? null)} is not spreadable`);
    }
    return [...args.slice(0, -1), ...last];
}
function appendValues(target, values) {
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
function callRuntime(callee, args, context) {
    if (isRuntimeCallable(callee) || isGoJuniorFunction(callee)) {
        return callee.call(args, context);
    }
    throw new GoJuniorRuntimeError(`${formatValue(callee)} is not callable`);
}
function getSelector(expression, context) {
    const object = evaluateExpression(expression.object, context);
    if (object instanceof SheetBinding) {
        return object.get(expression.field);
    }
    const struct = structFromValue(object);
    if (struct) {
        const fieldValue = struct.get(expression.field);
        if (fieldValue !== undefined)
            return fieldValue;
        const method = context.methodFor(struct.typeName, expression.field);
        if (method)
            return boundMethodValue(method, object);
    }
    if (isRuntimeObject(object)) {
        const value = object[expression.field];
        if (value !== undefined)
            return value;
    }
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}
function setSelector(expression, value, context) {
    const object = evaluateExpression(expression.object, context);
    if (object instanceof SheetBinding) {
        object.set(expression.field, value);
        return;
    }
    const struct = structFromValue(object);
    if (struct) {
        assertAssignableToStructField(struct, expression.field, value, context);
        struct.set(expression.field, value);
        return;
    }
    if (isRuntimeObject(object)) {
        object[expression.field] = value;
        return;
    }
    throw new GoJuniorRuntimeError(`${formatValue(object)} has no selector ${expression.field}`);
}
function getSpreadsheetRange(expression, context) {
    const sheet = evaluateExpression(expression.start.object, context);
    if (!(sheet instanceof SheetBinding)) {
        throw new GoJuniorRuntimeError("spreadsheet ranges must start with a sheet namespace");
    }
    return sheet.range(expression.start.field, expression.endCell);
}
function getIndex(object, index) {
    if (object instanceof RuntimeMap)
        return object.get(index);
    if (isRuntimeObject(object))
        return object[String(index)] ?? null;
    const numericIndex = toNumber(index);
    if (Array.isArray(object))
        return object[numericIndex] ?? null;
    if (typeof object === "string")
        return object[numericIndex] ?? "";
    throw new GoJuniorRuntimeError(`${formatValue(object)} is not indexable`);
}
function setIndex(expression, value, context) {
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
function pointerToExpression(expression, context) {
    if (expression.kind === "Identifier") {
        const value = context.lookup(expression.name);
        const typeName = pointerTypeName(value);
        return context.pointerToBinding(expression.name, typeName);
    }
    if (expression.kind === "SelectorExpression") {
        const object = evaluateExpression(expression.object, context);
        const struct = structFromValue(object);
        if (!struct)
            throw new GoJuniorRuntimeError("address-of selector requires a struct value");
        const current = struct.get(expression.field);
        const typeName = pointerTypeName(current ?? null);
        return new RuntimePointer(typeName, () => struct.get(expression.field) ?? null, (next) => {
            assertAssignableToStructField(struct, expression.field, next, context);
            struct.set(expression.field, next);
        });
    }
    if (expression.kind === "IndexExpression") {
        const object = evaluateExpression(expression.object, context);
        const index = evaluateExpression(expression.index, context);
        if (!Array.isArray(object))
            throw new GoJuniorRuntimeError("address-of index requires an array or slice value");
        const numericIndex = toNumber(index);
        const typeName = pointerTypeName(object[numericIndex] ?? null);
        return new RuntimePointer(typeName, () => object[numericIndex] ?? null, (next) => {
            object[numericIndex] = next;
        });
    }
    throw new GoJuniorRuntimeError("expression is not addressable");
}
function dereference(value) {
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
function assertAssignableToStructField(struct, fieldName, value, context) {
    const typeDef = context.typeDef(struct.typeName);
    const field = typeDef?.fields.find((candidate) => candidate.name === fieldName);
    if (!field)
        throw new GoJuniorRuntimeError(`${struct.typeName} has no field ${fieldName}`);
    assertAssignableToType(value, field.type.text, `field ${fieldName}`, context);
}
function pointerTypeName(value) {
    const actual = dereferenceIfPointer(value);
    if (actual instanceof RuntimeStruct)
        return actual.typeName;
    if (typeof actual === "bigint")
        return "int";
    if (typeof actual === "number")
        return "float64";
    if (typeof actual === "string")
        return "string";
    if (typeof actual === "boolean")
        return "bool";
    return "interface{}";
}
function inferredTypeText(value) {
    if (value === null)
        return undefined;
    if (typeof value === "bigint")
        return "int64";
    if (typeof value === "number")
        return "float64";
    if (typeof value === "string")
        return "string";
    if (typeof value === "boolean")
        return "bool";
    if (Array.isArray(value))
        return undefined;
    if (value instanceof RuntimeMap)
        return `map[${value.keyType}]${value.valueType}`;
    if (value instanceof RuntimeStruct)
        return value.typeName;
    if (value instanceof RuntimePointer)
        return `*${value.typeName}`;
    return undefined;
}
function getSlice(object, start, end) {
    const startIndex = start === undefined ? undefined : toNumber(start);
    const endIndex = end === undefined ? undefined : toNumber(end);
    if (Array.isArray(object))
        return object.slice(startIndex, endIndex);
    if (typeof object === "string")
        return object.slice(startIndex, endIndex);
    throw new GoJuniorRuntimeError(`${formatValue(object)} is not sliceable`);
}
function defaultValueForDeclarationType(type, context) {
    return defaultValueForTypeText(type?.text ?? "", context);
}
function defaultValueForTypeText(typeText, context) {
    const mapType = parseMapTypeText(typeText);
    if (mapType)
        return new RuntimeMap(mapType.keyType, mapType.valueType);
    const arrayType = parseArrayOrSliceTypeText(typeText);
    if (arrayType) {
        if (arrayType.length === undefined || arrayType.inferLength)
            return [];
        return Array.from({ length: arrayType.length }, () => defaultValueForTypeText(arrayType.elementType, context));
    }
    const type = normalizeTypeText(typeText);
    const typeDef = context?.typeDef(type);
    if (typeDef) {
        const struct = new RuntimeStruct(typeDef.name);
        for (const field of typeDef.fields) {
            struct.set(field.name, defaultValueForTypeText(field.type.text, context));
        }
        return struct;
    }
    return zeroValueForMapValue(typeText);
}
function zeroValueForMapValue(typeText) {
    const type = normalizeTypeText(typeText);
    if (isIntegerType(type))
        return 0n;
    if (isFloatType(type))
        return 0;
    if (type === "string")
        return "";
    if (type === "bool")
        return false;
    return null;
}
function assertAssignableToType(value, typeText, role, context) {
    const type = normalizeTypeText(typeText);
    if (!type || type === "<missing>" || type === "any" || type === "interface{}")
        return;
    const interfaceType = context?.interfaceDef(type);
    if (value === null) {
        if (isNilAssignableType(type) || interfaceType)
            return;
        throw new GoJuniorRuntimeError(`${role} nil is not assignable to ${type}`);
    }
    if (type === "string") {
        if (typeof value !== "string")
            throwTypeError(value, type, role);
        return;
    }
    if (type === "bool") {
        if (typeof value !== "boolean")
            throwTypeError(value, type, role);
        return;
    }
    if (isIntegerType(type)) {
        if (typeof value !== "bigint" || !integerInRange(value, type))
            throwTypeError(value, type, role);
        return;
    }
    if (isFloatType(type)) {
        if (typeof value !== "number")
            throwTypeError(value, type, role);
        return;
    }
    const mapType = parseMapTypeText(type);
    if (mapType) {
        if (!(value instanceof RuntimeMap))
            throwTypeError(value, type, role);
        if (normalizeTypeText(value.keyType) !== normalizeTypeText(mapType.keyType) ||
            normalizeTypeText(value.valueType) !== normalizeTypeText(mapType.valueType)) {
            throwTypeError(value, type, role);
        }
        return;
    }
    if (type.startsWith("*")) {
        if (!(value instanceof RuntimePointer) || value.typeName !== type.slice(1))
            throwTypeError(value, type, role);
        return;
    }
    const structType = context?.typeDef(type);
    if (structType) {
        if (!(value instanceof RuntimeStruct) || value.typeName !== structType.name)
            throwTypeError(value, type, role);
        return;
    }
    if (interfaceType && context) {
        if (!valueImplementsInterface(value, interfaceType, context))
            throwTypeError(value, type, role);
        return;
    }
    if (value instanceof RuntimeStruct) {
        if (value.typeName !== type)
            throwTypeError(value, type, role);
        return;
    }
    const arrayType = parseArrayOrSliceTypeText(type);
    if (arrayType) {
        if (!Array.isArray(value))
            throwTypeError(value, type, role);
        if (arrayType.length !== undefined && !arrayType.inferLength && value.length !== arrayType.length) {
            throwTypeError(value, type, role);
        }
        for (const item of value) {
            assertAssignableToType(item, arrayType.elementType, `${role} element`, context);
        }
        return;
    }
    if (type.startsWith("func(")) {
        if (!isRuntimeCallable(value) && !isGoJuniorFunction(value))
            throwTypeError(value, type, role);
        return;
    }
}
function valueImplementsInterface(value, interfaceType, context) {
    const receiver = dereferenceIfPointer(value);
    if (!(receiver instanceof RuntimeStruct))
        return interfaceType.methods.length === 0;
    for (const method of interfaceType.methods) {
        const candidate = context.methodFor(receiver.typeName, method.name);
        if (!candidate)
            return false;
        if (candidate.pointerReceiver && !(value instanceof RuntimePointer))
            return false;
        if (!signaturesCompatible(candidate.declaration.signature, method.signature))
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
function runtimeValueMatchesType(value, typeText) {
    const type = normalizeTypeText(typeText);
    if (type === "nil")
        return value === null;
    if (value === null)
        return false;
    if (!type || type === "<missing>")
        return false;
    if (type === "any" || type === "interface{}")
        return true;
    if (type === "string")
        return typeof value === "string";
    if (type === "bool")
        return typeof value === "boolean";
    if (isIntegerType(type))
        return typeof value === "bigint" && integerInRange(value, type);
    if (isFloatType(type))
        return typeof value === "number";
    if (type.startsWith("*"))
        return value instanceof RuntimePointer && value.typeName === type.slice(1);
    if (value instanceof RuntimeStruct)
        return value.typeName === type;
    if (value instanceof RuntimeMap) {
        const mapType = parseMapTypeText(type);
        return Boolean(mapType &&
            normalizeTypeText(value.keyType) === normalizeTypeText(mapType.keyType) &&
            normalizeTypeText(value.valueType) === normalizeTypeText(mapType.valueType));
    }
    if (type.startsWith("[]") || /^\[[0-9]*\]/.test(type))
        return Array.isArray(value);
    if (type.startsWith("func("))
        return isRuntimeCallable(value) || isGoJuniorFunction(value);
    return false;
}
function valueLength(value) {
    if (typeof value === "string" || Array.isArray(value))
        return value.length;
    if (value instanceof RuntimeMap)
        return value.size();
    throw new GoJuniorRuntimeError(`${formatValue(value)} has no len`);
}
function valueCapacity(value) {
    if (Array.isArray(value))
        return sliceCapacity(value);
    throw new GoJuniorRuntimeError(`${formatValue(value)} has no cap`);
}
function sliceCapacity(value) {
    return arrayCapacities.get(value) ?? value.length;
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
function parseArrayOrSliceTypeText(typeText) {
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
        return { elementType, inferLength: false };
    }
    if (lengthText === "...") {
        return { elementType, inferLength: true };
    }
    if (!/^(0|[1-9][0-9]*)$/.test(lengthText))
        return undefined;
    return { elementType, length: Number(lengthText), inferLength: false };
}
function normalizeTypeText(typeText) {
    return typeText.replace(/\s+/g, "");
}
function runtimeMapKeyId(key) {
    if (key === null)
        return "nil";
    if (typeof key === "boolean")
        return `b:${key}`;
    if (typeof key === "string")
        return `s:${key}`;
    if (typeof key === "bigint")
        return `i:${key}`;
    if (typeof key === "number")
        return `f:${Object.is(key, -0) ? "-0" : String(key)}`;
    return `o:${objectIdentityId(key)}`;
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
        type.startsWith("*") ||
        type.startsWith("[]") ||
        type.startsWith("map[") ||
        type.startsWith("func(");
}
function isIntegerType(type) {
    return /^(?:u?int(?:8|16|32|64)?|byte|rune)$/.test(type);
}
function isFloatType(type) {
    return type === "float32" || type === "float64";
}
function integerInRange(value, type) {
    const bounds = integerTypeBounds(type);
    return !bounds || (value >= bounds.min && value <= bounds.max);
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
            return { min: 0n, max: 18446744073709551615n };
        case "int":
        case "int64":
            return { min: -9223372036854775808n, max: 9223372036854775807n };
        default:
            return undefined;
    }
}
function addNumbers(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint")
        return left + right;
    return toFloat(left) + toFloat(right);
}
function subtractNumbers(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint")
        return left - right;
    return toFloat(left) - toFloat(right);
}
function multiplyNumbers(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint")
        return left * right;
    return toFloat(left) * toFloat(right);
}
function divideNumbers(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint")
        return left / right;
    return toFloat(left) / toFloat(right);
}
function moduloNumbers(left, right) {
    if (typeof left === "bigint" && typeof right === "bigint")
        return left % right;
    return toFloat(left) % toFloat(right);
}
function negateNumber(value) {
    if (typeof value === "bigint")
        return -value;
    return -toFloat(value);
}
function numericIdentity(value) {
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
function compareValues(left, right) {
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
function valueEqual(left, right) {
    if ((typeof left === "bigint" || typeof left === "number") && (typeof right === "bigint" || typeof right === "number")) {
        return toFloat(left) === toFloat(right);
    }
    return left === right;
}
function rangeEntries(source) {
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
function toStringValue(value) {
    if (typeof value === "string")
        return value;
    return formatValue(value);
}
export function formatValue(value) {
    if (value === null)
        return "<nil>";
    if (typeof value === "bigint")
        return value.toString();
    if (typeof value === "number" || typeof value === "boolean" || typeof value === "string")
        return String(value);
    if (Array.isArray(value))
        return `[${value.map(formatValue).join(" ")}]`;
    if (value instanceof RuntimeMap)
        return formatRuntimeMap(value);
    if (value instanceof RuntimeStruct)
        return formatRuntimeStruct(value);
    if (value instanceof RuntimePointer)
        return `&${formatValue(value.get())}`;
    if (isRuntimeCallable(value) || isGoJuniorFunction(value))
        return `<func ${value.name}>`;
    if (value instanceof SheetBinding)
        return `<sheet ${value.name}>`;
    return `{${Object.entries(value).map(([key, item]) => `${key}:${formatValue(item)}`).join(" ")}}`;
}
function formatGoSyntaxValue(value) {
    if (value === null)
        return "nil";
    if (typeof value === "bigint")
        return `${integerTypeName(value)}(${value.toString()})`;
    if (typeof value === "number")
        return `float64(${formatFloat(value)})`;
    if (typeof value === "boolean")
        return `bool(${value})`;
    if (typeof value === "string")
        return `string(${JSON.stringify(value)})`;
    if (Array.isArray(value))
        return `[]interface{}{${value.map(formatGoSyntaxValue).join(", ")}}`;
    if (value instanceof RuntimeMap)
        return formatGoSyntaxMap(value);
    if (value instanceof RuntimeStruct)
        return formatGoSyntaxStruct(value);
    if (value instanceof RuntimePointer)
        return `&${formatGoSyntaxValue(value.get())}`;
    if (isRuntimeCallable(value) || isGoJuniorFunction(value))
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
function formatRuntimeStruct(value) {
    const fields = value.orderedFields()
        .map(([key, item]) => `${key}:${formatValue(item)}`)
        .join(" ");
    return `${value.typeName}{${fields}}`;
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
        !(value instanceof SheetBinding) &&
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
