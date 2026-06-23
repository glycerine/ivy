import { REPL_FILENAME } from "./diagnostics.js";
import { analyzeEffects } from "./effects.js";
class AsyncEmitter {
    program;
    diagnostics = [];
    functions = new Map();
    nextLocal = 0;
    constructor(program) {
        this.program = program;
        for (const declaration of program.functions) {
            if (declaration.receiver) {
                this.unsupported("method emission is not in the async emitter slice yet", declaration.span);
                continue;
            }
            this.functions.set(declaration.name, {
                declaration,
                jsName: generatedFunctionName(declaration.name)
            });
        }
    }
    emit() {
        const effects = analyzeEffects(this.program);
        const source = this.diagnostics.length > 0 ? "" : this.emitModule(effects);
        return {
            diagnostics: this.diagnostics,
            effects,
            source: this.diagnostics.length > 0 ? "" : source
        };
    }
    emitModule(effects) {
        const lines = [];
        lines.push(`"use strict";`);
        lines.push(`const __gojr = ((__runtime = {}) => {`);
        lines.push(`  const __scheduler = __runtime.scheduler;`);
        lines.push(`  const __AsyncGoChannel = __runtime.AsyncGoChannel;`);
        lines.push(`  const __asyncSelect = __runtime.asyncSelect;`);
        lines.push(`  const __makeChannel = (capacity, zeroValue) => new __AsyncGoChannel(__scheduler, Number(capacity ?? 0), zeroValue);`);
        lines.push(`  const __functions = Object.create(null);`);
        lines.push(`  const __effects = ${JSON.stringify(effects)};`);
        for (const info of this.functions.values()) {
            lines.push(this.indent(this.emitFunction(info), 1));
            lines.push(`  __functions[${JSON.stringify(info.declaration.name)}] = ${info.jsName};`);
        }
        lines.push(this.indent(this.emitMain(), 1));
        lines.push(`  return { functions: __functions, main: __main, effects: __effects };`);
        lines.push(`})(arguments[0] ?? {});`);
        lines.push(`return __gojr;`);
        return lines.join("\n");
    }
    emitFunction(info) {
        const scope = this.rootScope();
        const params = [];
        for (const [index, parameter] of info.declaration.signature.parameters.entries()) {
            const sourceName = parameter.name ?? `_arg${index}`;
            const jsName = this.declare(scope, sourceName);
            params.push(parameter.variadic ? `...${jsName}` : jsName);
        }
        const body = this.emitBlockStatements(info.declaration.body, scope, false);
        return [
            `async function ${info.jsName}(${params.join(", ")}) {`,
            this.indent(body || `return null;`, 1),
            `}`
        ].join("\n");
    }
    emitMain() {
        const scope = this.rootScope();
        const body = this.emitStatements(this.program.body, scope);
        return [
            `async function __main() {`,
            this.indent(body || `return null;`, 1),
            `}`
        ].join("\n");
    }
    emitBlockStatements(block, scope, createScope) {
        return this.emitStatements(block.statements, createScope ? this.childScope(scope) : scope);
    }
    emitStatements(statements, scope) {
        const lines = [];
        for (const statement of statements) {
            const emitted = this.emitStatement(statement, scope);
            if (emitted)
                lines.push(emitted);
        }
        return lines.join("\n");
    }
    emitStatement(statement, scope) {
        switch (statement.kind) {
            case "BlockStatement":
                return `{\n${this.indent(this.emitBlockStatements(statement, scope, true), 1)}\n}`;
            case "VarDecl":
                return statement.declarations.map((declaration) => {
                    const jsName = this.declare(scope, declaration.name);
                    const value = declaration.value ? this.emitExpression(declaration.value, scope) : "null";
                    return `let ${jsName} = ${value};`;
                }).join("\n");
            case "ShortVarStatement":
                return statement.names.map((name, index) => {
                    const jsName = this.declare(scope, name);
                    const value = statement.values[index] ? this.emitExpression(statement.values[index], scope) : "null";
                    return `let ${jsName} = ${value};`;
                }).join("\n");
            case "AssignStatement":
                return this.emitAssign(statement, scope);
            case "ReturnStatement":
                return this.emitReturn(statement, scope);
            case "ExpressionStatement":
                return `${this.emitExpression(statement.expression, scope)};`;
            case "GoStatement":
                return this.emitGo(statement, scope);
            case "SendStatement":
                return this.emitSend(statement, scope);
            case "SelectStatement":
                return this.emitSelect(statement, scope);
            case "IfStatement": {
                const condition = this.emitExpression(statement.condition, scope);
                const thenBody = this.emitBlockStatements(statement.thenBlock, this.childScope(scope), false);
                const elseBody = statement.elseBranch
                    ? ` else ${statement.elseBranch.kind === "IfStatement"
                        ? this.emitStatement(statement.elseBranch, scope)
                        : `{\n${this.indent(this.emitBlockStatements(statement.elseBranch, this.childScope(scope), false), 1)}\n}`}`
                    : "";
                return `if (${condition}) {\n${this.indent(thenBody, 1)}\n}${elseBody}`;
            }
            default:
                this.unsupported(`unsupported statement in async emitter: ${statement.kind}`, statement.span);
                return "";
        }
    }
    emitAssign(statement, scope) {
        if (statement.operator !== "=") {
            this.unsupported(`unsupported assignment operator in async emitter: ${statement.operator}`, statement.span);
            return "";
        }
        return statement.targets.map((target, index) => {
            const value = statement.values[index] ? this.emitExpression(statement.values[index], scope) : "null";
            return `${this.emitAssignableExpression(target, scope)} = ${value};`;
        }).join("\n");
    }
    emitReturn(statement, scope) {
        const values = statement.values.map((value) => this.emitExpression(value, scope));
        if (values.length === 0)
            return `return null;`;
        if (values.length === 1)
            return `return ${values[0]};`;
        return `return [${values.join(", ")}];`;
    }
    emitGo(statement, scope) {
        return `__scheduler.go(async () => { await ${this.emitInvocation(statement.call, scope)}; });`;
    }
    emitSend(statement, scope) {
        return `await ${this.emitExpression(statement.channel, scope)}.send(${this.emitExpression(statement.value, scope)});`;
    }
    emitSelect(statement, scope) {
        const resultName = `_sel${this.nextLocal++}`;
        const cases = statement.clauses.map((clause) => this.emitSelectCaseObject(clause, scope));
        const lines = [];
        lines.push(`{`);
        lines.push(`  const ${resultName} = await __asyncSelect(__scheduler, [${cases.join(", ")}]);`);
        lines.push(`  switch (${resultName}.index) {`);
        for (const [index, clause] of statement.clauses.entries()) {
            const caseScope = this.childScope(scope);
            const prelude = this.emitSelectCasePrelude(clause, resultName, caseScope);
            const body = this.emitStatements(clause.statements, caseScope);
            lines.push(`    case ${index}: {`);
            const caseLines = [prelude, body, `break;`].filter(Boolean).join("\n");
            lines.push(this.indent(caseLines, 3));
            lines.push(`    }`);
        }
        lines.push(`  }`);
        lines.push(`}`);
        return lines.join("\n");
    }
    emitExpression(expression, scope) {
        switch (expression.kind) {
            case "Identifier":
                return this.lookup(scope, expression.name, expression.span);
            case "Literal":
                return this.emitLiteral(expression);
            case "FunctionLiteralExpression":
                return this.emitFunctionLiteral(expression, scope);
            case "UnaryExpression":
                return this.emitUnary(expression, scope);
            case "BinaryExpression":
                return this.emitBinary(expression, scope);
            case "CallExpression":
                return this.emitCall(expression, scope);
            case "TypeAssertionExpression":
                return this.emitTypeAssertion(expression, scope);
            default:
                this.unsupported(`unsupported expression in async emitter: ${expression.kind}`, expression.span);
                return "undefined";
        }
    }
    emitAssignableExpression(expression, scope) {
        if (expression.kind === "Identifier")
            return this.lookup(scope, expression.name, expression.span);
        this.unsupported(`unsupported assignment target in async emitter: ${expression.kind}`, expression.span);
        return "undefined";
    }
    emitLiteral(expression) {
        switch (expression.literalKind) {
            case "int":
            case "rune":
                return `${String(expression.value)}n`;
            case "float":
                return JSON.stringify(expression.value);
            case "string":
                return JSON.stringify(expression.value);
            case "bool":
                return expression.value ? "true" : "false";
            case "nil":
                return "null";
            case "imag":
                return `{ real: 0, imag: ${JSON.stringify(expression.value.imag)} }`;
        }
    }
    emitUnary(expression, scope) {
        const operand = this.emitExpression(expression.operand, scope);
        if (expression.operator === "<-") {
            return `(await ${operand}.receive())[0]`;
        }
        return `(${expression.operator}${operand})`;
    }
    emitBinary(expression, scope) {
        const left = this.emitExpression(expression.left, scope);
        const right = this.emitExpression(expression.right, scope);
        return `(${left} ${expression.operator} ${right})`;
    }
    emitCall(expression, scope) {
        if (expression.callee.kind === "Identifier" && expression.callee.name === "make") {
            return this.emitMake(expression, scope);
        }
        return `(await ${this.emitInvocation(expression, scope)})`;
    }
    emitInvocation(expression, scope) {
        if (expression.callee.kind !== "Identifier") {
            if (expression.callee.kind === "FunctionLiteralExpression") {
                const callee = this.emitFunctionLiteral(expression.callee, scope);
                const args = expression.args.map((arg) => this.emitExpression(arg, scope)).join(", ");
                return `${callee}(${args})`;
            }
            this.unsupported("only direct function calls are supported in this async emitter slice", expression.span);
            return "undefined";
        }
        const target = this.functions.get(expression.callee.name);
        if (!target) {
            this.unsupported(`unknown function call in async emitter: ${expression.callee.name}`, expression.callee.span);
            return "undefined";
        }
        const args = expression.args.map((arg) => this.emitExpression(arg, scope)).join(", ");
        return `${target.jsName}(${args})`;
    }
    emitMake(expression, scope) {
        const typeArg = expression.args[0];
        if (!typeArg || typeArg.kind !== "TypeExpression" || !isChannelType(typeArg)) {
            this.unsupported("only make(chan T[, n]) is supported in this async emitter slice", expression.span);
            return "undefined";
        }
        const capacity = expression.args[1] ? this.emitExpression(expression.args[1], scope) : "0";
        return `__makeChannel(${capacity}, ${zeroValueFactoryForType(channelElementType(typeArg.type.text))})`;
    }
    emitFunctionLiteral(expression, scope) {
        const literalScope = this.childScope(scope);
        const params = [];
        for (const [index, parameter] of expression.signature.parameters.entries()) {
            const sourceName = parameter.name ?? `_arg${index}`;
            const jsName = this.declare(literalScope, sourceName);
            params.push(parameter.variadic ? `...${jsName}` : jsName);
        }
        const body = this.emitBlockStatements(expression.body, literalScope, false);
        return `(async function(${params.join(", ")}) {\n${this.indent(body || `return null;`, 1)}\n})`;
    }
    emitSelectCaseObject(clause, scope) {
        if (clause.default || !clause.comm)
            return `{ op: "default" }`;
        if (clause.comm.kind === "SendStatement") {
            return `{ op: "send", channel: ${this.emitExpression(clause.comm.channel, scope)}, value: ${this.emitExpression(clause.comm.value, scope)} }`;
        }
        const receive = receiveChannelFromStatement(clause.comm);
        if (receive)
            return `{ op: "receive", channel: ${this.emitExpression(receive, scope)} }`;
        this.unsupported(`unsupported select communication clause in async emitter: ${clause.comm.kind}`, clause.comm.span);
        return `{ op: "default" }`;
    }
    emitSelectCasePrelude(clause, resultName, scope) {
        if (!clause.comm)
            return "";
        if (clause.comm.kind === "ShortVarStatement") {
            return this.emitReceiveShortVarPrelude(clause.comm, resultName, scope);
        }
        if (clause.comm.kind === "AssignStatement") {
            return this.emitReceiveAssignPrelude(clause.comm, resultName, scope);
        }
        return "";
    }
    emitReceiveShortVarPrelude(statement, resultName, scope) {
        if (!receiveChannelFromStatement(statement))
            return "";
        return statement.names.map((name, index) => {
            const jsName = this.declare(scope, name);
            const property = index === 1 ? "ok" : "value";
            return `let ${jsName} = ${resultName}.${property};`;
        }).join("\n");
    }
    emitReceiveAssignPrelude(statement, resultName, scope) {
        if (!receiveChannelFromStatement(statement))
            return "";
        return statement.targets.map((target, index) => {
            const property = index === 1 ? "ok" : "value";
            return `${this.emitAssignableExpression(target, scope)} = ${resultName}.${property};`;
        }).join("\n");
    }
    emitTypeAssertion(expression, scope) {
        this.unsupported("type assertion emission is not in the first async emitter slice", expression.span);
        return this.emitExpression(expression.expression, scope);
    }
    rootScope() {
        return { bindings: new Map() };
    }
    childScope(parent) {
        return { parent, bindings: new Map() };
    }
    declare(scope, name) {
        const jsName = `_v${this.nextLocal++}_${sanitizeIdentifierPart(name)}`;
        scope.bindings.set(name, jsName);
        return jsName;
    }
    lookup(scope, name, span) {
        let cursor = scope;
        while (cursor) {
            const found = cursor.bindings.get(name);
            if (found)
                return found;
            cursor = cursor.parent;
        }
        const functionInfo = this.functions.get(name);
        if (functionInfo)
            return functionInfo.jsName;
        this.unsupported(`unknown identifier in async emitter: ${name}`, span);
        return "undefined";
    }
    unsupported(message, span) {
        this.diagnostics.push({
            filename: span?.filename ?? REPL_FILENAME,
            code: "GOJR_EMIT001",
            severity: "error",
            message,
            ...(span ? { span } : {})
        });
    }
    indent(text, level) {
        if (!text)
            return "";
        const prefix = "  ".repeat(level);
        return text.split("\n").map((line) => line ? `${prefix}${line}` : line).join("\n");
    }
}
export function emitAsyncJavaScript(program) {
    return new AsyncEmitter(program).emit();
}
function generatedFunctionName(name) {
    return `_fn_${sanitizeIdentifierPart(name)}`;
}
function sanitizeIdentifierPart(name) {
    const sanitized = name.replace(/[^A-Za-z0-9_$]/g, "_");
    if (/^[A-Za-z_$]/.test(sanitized))
        return sanitized || "_";
    return `_${sanitized}`;
}
function receiveChannelFromStatement(statement) {
    if (statement.kind === "ExpressionStatement") {
        return receiveChannelFromExpression(statement.expression);
    }
    if (statement.kind === "ShortVarStatement" && statement.values.length === 1) {
        return receiveChannelFromExpression(statement.values[0]);
    }
    if (statement.kind === "AssignStatement" && statement.values.length === 1) {
        return receiveChannelFromExpression(statement.values[0]);
    }
    return undefined;
}
function receiveChannelFromExpression(expression) {
    if (expression.kind === "UnaryExpression" && expression.operator === "<-") {
        return expression.operand;
    }
    return undefined;
}
function isChannelType(expression) {
    return normalizeTypeText(expression.type.text).startsWith("chan ")
        || normalizeTypeText(expression.type.text).startsWith("<-chan ")
        || normalizeTypeText(expression.type.text).startsWith("chan<- ");
}
function channelElementType(typeText) {
    const normalized = normalizeTypeText(typeText);
    if (normalized.startsWith("<-chan "))
        return normalized.slice("<-chan ".length).trim();
    if (normalized.startsWith("chan<- "))
        return normalized.slice("chan<- ".length).trim();
    if (normalized.startsWith("chan "))
        return normalized.slice("chan ".length).trim();
    return "any";
}
function zeroValueFactoryForType(typeText) {
    const normalized = normalizeTypeText(typeText);
    if (/^(?:u?int(?:8|16|32|64)?|uintptr|byte|rune)$/.test(normalized))
        return `() => 0n`;
    if (/^(?:float32|float64)$/.test(normalized))
        return `() => 0`;
    if (/^(?:complex64|complex128)$/.test(normalized))
        return `() => ({ real: 0, imag: 0 })`;
    if (normalized === "string")
        return `() => ""`;
    if (normalized === "bool")
        return `() => false`;
    return `() => null`;
}
function normalizeTypeText(typeText) {
    return typeText.replace(/\s+/g, " ").trim();
}
