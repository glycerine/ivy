import { EmitterContext } from "./context.js";
import { checkedInWasmStencil } from "./stencils.js";
export const GOJR_STAGE1_BACKEND = "copy-patch-wasm-stage1";
const emptyExpressionEnv = { locals: new Map() };
export function emitStage1Package(artifact, ast) {
    const ctx = new EmitterContext({ artifact });
    const wasmFunctions = ast.functions.filter(isI64AddFunction);
    const usesWasm = wasmFunctions.length > 0;
    const wasmBase64 = usesWasm ? checkedInWasmStencil("i64.add.kernel").wasmBase64 : undefined;
    const bodyLines = [];
    for (const statement of ast.body) {
        if (statement.kind === "ConstDecl") {
            bodyLines.push(...emitConstDecl(ctx, statement));
            continue;
        }
        if (statement.kind === "VarDecl") {
            bodyLines.push(...emitVarDecl(ctx, statement));
            continue;
        }
        if (statement.kind === "TypeDecl")
            continue;
        ctx.emitError(`unsupported Stage 1 top-level statement ${statement.kind}`);
    }
    for (const fn of ast.functions) {
        bodyLines.push(...emitFunction(ctx, fn));
    }
    const javascript = stage1JavaScript(artifact, usesWasm, bodyLines);
    return {
        diagnostics: ctx.diagnostics,
        javascript,
        ...(wasmBase64 ? { wasmBase64 } : {})
    };
}
function emitConstDecl(ctx, statement) {
    return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "const", declaration));
}
function emitVarDecl(ctx, statement) {
    return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "var", declaration));
}
function emitDeclaration(ctx, _kind, declaration) {
    const value = declaration.value
        ? expressionToJs(ctx, declaration.value)
        : zeroValueForType(declaration.type?.text);
    if (!value) {
        ctx.emitError(`unsupported Stage 1 declaration for ${declaration.name}`);
        return [];
    }
    return [`  pkg[${JSON.stringify(declaration.name)}] = ${value};`];
}
function emitFunction(ctx, fn) {
    if (isI64AddFunction(fn)) {
        return [
            `  pkg[${JSON.stringify(fn.name)}] = async (a, b) => __gojrWasmExports.add_i64(BigInt(a), BigInt(b));`
        ];
    }
    const statements = fn.body.statements;
    if (statements.length === 0) {
        return [`  pkg[${JSON.stringify(fn.name)}] = async () => null;`];
    }
    if (statements.length === 1 && statements[0]?.kind === "ReturnStatement") {
        const parameters = functionParameterBindings(fn);
        const returned = returnExpression(ctx, statements[0], parameters.env);
        if (returned)
            return [`  pkg[${JSON.stringify(fn.name)}] = async (${parameters.params.join(", ")}) => ${returned};`];
    }
    ctx.emitError(`unsupported Stage 1 function body for ${fn.name}`);
    return [];
}
function returnExpression(ctx, statement, env) {
    if (statement.values.length === 0)
        return "null";
    if (statement.values.length !== 1)
        return undefined;
    return expressionToJs(ctx, statement.values[0], env);
}
function functionParameterBindings(fn) {
    const locals = new Map();
    const params = fn.signature.parameters.map((parameter, index) => {
        const goName = parameter.name && parameter.name !== "_" ? parameter.name : `__gojrArg${index}`;
        const jsName = safeLocalName(goName, index);
        if (parameter.name && parameter.name !== "_")
            locals.set(parameter.name, jsName);
        return jsName;
    });
    return { params, env: { locals } };
}
function isI64AddFunction(fn) {
    if (fn.signature.parameters.length !== 2 || fn.signature.results.length !== 1)
        return false;
    if (!fn.signature.parameters.every((param) => param.type.text === "int64"))
        return false;
    if (fn.signature.results[0]?.type.text !== "int64")
        return false;
    const statement = onlyReturn(fn.body);
    if (!statement || statement.values.length !== 1)
        return false;
    const value = statement.values[0];
    if (!value || value.kind !== "BinaryExpression" || value.operator !== "+")
        return false;
    return isIdentifier(value.left, fn.signature.parameters[0]?.name) &&
        isIdentifier(value.right, fn.signature.parameters[1]?.name);
}
function onlyReturn(body) {
    return body.statements.length === 1 && body.statements[0]?.kind === "ReturnStatement"
        ? body.statements[0]
        : undefined;
}
function isIdentifier(expression, name) {
    return Boolean(name && expression.kind === "Identifier" && expression.name === name);
}
function expressionToJs(ctx, expression, env = emptyExpressionEnv) {
    if (!expression)
        return undefined;
    switch (expression.kind) {
        case "Identifier":
            return identifierToJs(expression, env);
        case "Literal":
            return literalToJs(expression);
        case "ArrayLiteralExpression":
            return arrayLiteralToJs(ctx, expression, env);
        case "StructLiteralExpression":
            return structLiteralToJs(ctx, expression, env);
        case "MapLiteralExpression":
            return mapLiteralToJs(ctx, expression, env);
        case "UnaryExpression":
            return unaryExpressionToJs(ctx, expression, env);
        case "BinaryExpression":
            return binaryExpressionToJs(ctx, expression, env);
        case "SelectorExpression":
            return selectorExpressionToJs(ctx, expression, env);
        case "CallExpression":
            return callExpressionToJs(ctx, expression, env);
        case "IndexExpression":
            return indexExpressionToJs(ctx, expression, env);
        case "SliceExpression":
            return sliceExpressionToJs(ctx, expression, env);
        case "TypeAssertionExpression":
            return typeAssertionExpressionToJs(ctx, expression);
        default:
            ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
            return undefined;
    }
}
function identifierToJs(expression, env) {
    const local = env.locals.get(expression.name);
    if (local)
        return local;
    return `pkg[${JSON.stringify(expression.name)}]`;
}
function literalToJs(expression) {
    switch (expression.literalKind) {
        case "int":
        case "rune":
            if (typeof expression.value === "bigint")
                return `${expression.value.toString()}n`;
            if (typeof expression.value === "number")
                return `${Math.trunc(expression.value)}n`;
            return undefined;
        case "float":
            return typeof expression.value === "number" ? JSON.stringify(expression.value) : undefined;
        case "string":
            return typeof expression.value === "string" ? JSON.stringify(expression.value) : undefined;
        case "bool":
            return typeof expression.value === "boolean" ? String(expression.value) : undefined;
        case "nil":
            return "null";
        default:
            return undefined;
    }
}
function arrayLiteralToJs(ctx, expression, env) {
    if (expression.elements.some((element) => element.key !== undefined)) {
        ctx.emitError("unsupported Stage 3 keyed array literal");
        return undefined;
    }
    const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
    if (values.some((value) => value === undefined))
        return undefined;
    const rendered = values.filter((value) => value !== undefined);
    if (isByteSliceType(expression.type.text)) {
        const bytes = expression.elements.map((element) => byteLiteralValue(element.value));
        if (bytes.some((value) => value === undefined)) {
            ctx.emitError(`unsupported Stage 1 []byte literal element`);
            return undefined;
        }
        return `new Uint8Array([${bytes.join(", ")}])`;
    }
    return `[${rendered.join(", ")}]`;
}
function structLiteralToJs(ctx, expression, env) {
    const fields = [];
    for (const field of expression.fields) {
        if (!field.name) {
            ctx.emitError(`unsupported Stage 3 unkeyed struct literal ${expression.typeName}`);
            return undefined;
        }
        const value = expressionToJs(ctx, field.value, env);
        if (!value)
            return undefined;
        fields.push(`${JSON.stringify(field.name)}: ${value}`);
    }
    return `({ ${fields.join(", ")} })`;
}
function mapLiteralToJs(ctx, expression, env) {
    const entries = [];
    for (const entry of expression.entries) {
        const key = expressionToJs(ctx, entry.key, env);
        const value = expressionToJs(ctx, entry.value, env);
        if (!key || !value)
            return undefined;
        entries.push(`[${key}, ${value}]`);
    }
    return `new Map([${entries.join(", ")}])`;
}
function unaryExpressionToJs(ctx, expression, env) {
    const operand = expressionToJs(ctx, expression.operand, env);
    if (!operand)
        return undefined;
    switch (expression.operator) {
        case "+":
            return `(${operand})`;
        case "-":
            return `(-(${operand}))`;
        case "!":
            return `(!(${operand}))`;
        case "^":
            return `(~(${operand}))`;
        default:
            ctx.emitError(`unsupported Stage 3 unary operator ${expression.operator}`);
            return undefined;
    }
}
function binaryExpressionToJs(ctx, expression, env) {
    const left = expressionToJs(ctx, expression.left, env);
    const right = expressionToJs(ctx, expression.right, env);
    if (!left || !right)
        return undefined;
    if (expression.operator === "&&" || expression.operator === "||")
        return `((${left}) ${expression.operator} (${right}))`;
    if (expression.operator === "&^")
        return `((${left}) & ~(${right}))`;
    const operator = jsBinaryOperator(expression.operator);
    if (!operator) {
        ctx.emitError(`unsupported Stage 3 binary operator ${expression.operator}`);
        return undefined;
    }
    return `((${left}) ${operator} (${right}))`;
}
function selectorExpressionToJs(ctx, expression, env) {
    const object = expressionToJs(ctx, expression.object, env);
    if (!object)
        return undefined;
    return `(${object})[${JSON.stringify(expression.field)}]`;
}
function callExpressionToJs(ctx, expression, env) {
    if (expression.spreadLast) {
        ctx.emitError("unsupported Stage 3 spread call");
        return undefined;
    }
    if (expression.callee.kind === "TypeExpression")
        return conversionCallToJs(ctx, expression.callee.type.text, expression.args, env);
    if (expression.callee.kind === "Identifier" && primitiveTypeNames.has(expression.callee.name) && !env.locals.has(expression.callee.name)) {
        return conversionCallToJs(ctx, expression.callee.name, expression.args, env);
    }
    const args = expression.args.map((arg) => expressionToJs(ctx, arg, env));
    if (args.some((arg) => arg === undefined))
        return undefined;
    const renderedArgs = args.filter((arg) => arg !== undefined);
    if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
        return `(await pkg[${JSON.stringify(expression.callee.name)}](${renderedArgs.join(", ")}))`;
    }
    const callee = expressionToJs(ctx, expression.callee, env);
    if (!callee)
        return undefined;
    return `(await (${callee})(${renderedArgs.join(", ")}))`;
}
function conversionCallToJs(ctx, typeText, args, env) {
    if (args.length !== 1) {
        ctx.emitError(`unsupported Stage 3 conversion to ${typeText} with ${args.length} arguments`);
        return undefined;
    }
    const value = expressionToJs(ctx, args[0], env);
    if (!value)
        return undefined;
    if (isIntegerType(typeText))
        return `BigInt(${value})`;
    if (isFloatType(typeText))
        return `Number(${value})`;
    if (typeText === "string")
        return `String(${value})`;
    ctx.emitError(`unsupported Stage 3 conversion to ${typeText}`);
    return undefined;
}
function indexExpressionToJs(ctx, expression, env) {
    const object = expressionToJs(ctx, expression.object, env);
    const index = expressionToJs(ctx, expression.index, env);
    if (!object || !index)
        return undefined;
    const objectType = expression.object.typeText ?? "";
    if (objectType.startsWith("map[")) {
        return `((${object}).get(${index}) ?? ${zeroValueForType(expression.typeText) ?? "null"})`;
    }
    if (objectType === "string")
        return `BigInt(__gojrStringByteAt(${object}, Number(${index})))`;
    return `(${object})[Number(${index})]`;
}
function sliceExpressionToJs(ctx, expression, env) {
    if (expression.max) {
        ctx.emitError("unsupported Stage 3 three-index slice");
        return undefined;
    }
    const object = expressionToJs(ctx, expression.object, env);
    const start = expression.start ? expressionToJs(ctx, expression.start, env) : undefined;
    const end = expression.end ? expressionToJs(ctx, expression.end, env) : undefined;
    if (!object || (expression.start && !start) || (expression.end && !end))
        return undefined;
    if (!start && !end)
        return `(${object}).slice()`;
    if (!end)
        return `(${object}).slice(Number(${start}))`;
    return `(${object}).slice(${start ? `Number(${start})` : "undefined"}, Number(${end}))`;
}
function typeAssertionExpressionToJs(ctx, expression) {
    ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
    return undefined;
}
function byteLiteralValue(expression) {
    if (expression.kind !== "Literal")
        return undefined;
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
function zeroValueForType(typeText) {
    if (!typeText)
        return "null";
    switch (typeText) {
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
            return "0n";
        case "float32":
        case "float64":
            return "0";
        case "string":
            return "\"\"";
        case "bool":
            return "false";
        default:
            if (typeText.startsWith("[]"))
                return "[]";
            if (typeText.startsWith("map["))
                return "new Map()";
            return "null";
    }
}
function isByteSliceType(typeText) {
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
    "uint16",
    "uint32",
    "uint64",
    "uintptr",
    "float32",
    "float64"
]);
function isIntegerType(typeText) {
    return typeText === "int" ||
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
function isFloatType(typeText) {
    return typeText === "float32" || typeText === "float64";
}
function jsBinaryOperator(operator) {
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
function safeLocalName(name, index) {
    const candidate = name.replace(/[^A-Za-z0-9_$]/g, "_");
    if (/^[A-Za-z_$][A-Za-z0-9_$]*$/.test(candidate))
        return candidate;
    return `__gojrArg${index}`;
}
function stage1JavaScript(artifact, usesWasm, bodyLines) {
    const artifactHeader = { ...artifact };
    delete artifactHeader.runtime;
    return [
        "// Code generated by gojr Stage 1 copy-and-patch emitter; DO NOT EDIT.",
        `export const gojrPackageArtifact = ${JSON.stringify(artifactHeader, null, 2)};`,
        "export async function instantiateGoJrPackage(runtime = {}, options = {}) {",
        "  const pkg = Object.create(null);",
        "  const importsByPath = options.importsByPath || runtime.importsByPath || {};",
        "  void importsByPath;",
        usesWasm ? "  const wasmBase64 = options.wasmBase64 || runtime.wasmBase64;" : "",
        usesWasm ? "  if (typeof wasmBase64 !== \"string\" || wasmBase64.length === 0) throw new Error(\"gojr Stage 1 package requires options.wasmBase64 from _gojr.wasm\");" : "",
        usesWasm ? "  const __gojrWasmModule = new WebAssembly.Module(__gojrDecodeBase64(wasmBase64));" : "",
        usesWasm ? "  const __gojrWasmInstance = new WebAssembly.Instance(__gojrWasmModule, {});" : "",
        usesWasm ? "  const __gojrWasmExports = __gojrWasmInstance.exports;" : "  const __gojrWasmExports = {};",
        ...bodyLines,
        "  return { diagnostics: [], output: [], package: pkg, wasm: __gojrWasmExports };",
        "}",
        "function __gojrStringByteAt(value, index) {",
        "  const bytes = new TextEncoder().encode(String(value));",
        "  if (index < 0 || index >= bytes.length) throw new RangeError(\"string index out of range\");",
        "  return bytes[index];",
        "}",
        usesWasm ? "function __gojrDecodeBase64(base64) {" : "",
        usesWasm ? "  if (typeof Buffer !== \"undefined\") return new Uint8Array(Buffer.from(base64, \"base64\"));" : "",
        usesWasm ? "  const binary = atob(base64);" : "",
        usesWasm ? "  const out = new Uint8Array(binary.length);" : "",
        usesWasm ? "  for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);" : "",
        usesWasm ? "  return out;" : "",
        usesWasm ? "}" : "",
        "export default { artifact: gojrPackageArtifact, instantiateGoJrPackage };",
        ""
    ].filter((line) => line !== "").join("\n");
}
