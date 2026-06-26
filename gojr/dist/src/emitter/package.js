import { EmitterContext } from "./context.js";
import { checkedInWasmStencil } from "./stencils.js";
import { GENERATED_ARTIFACT_INTRINSIC_IMPORTS } from "../intrinsicPackages.js";
export const GOJR_STAGE1_BACKEND = "copy-patch-wasm-stage1";
const emptyPackageEmitFacts = {
    methodKeys: new Map(),
    resultCounts: new Map(),
    packageTypes: new Map(),
    typeUnderlyings: new Map(),
    imports: new Map(),
    interfaceTypes: new Set(),
    functionParamTypes: new Map(),
    functionResultTypes: new Map(),
    functionTypeParameters: new Map(),
    methodPointerReceivers: new Map()
};
const emptyExpressionEnv = { locals: new Map(), localTypes: new Map(), facts: emptyPackageEmitFacts };
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
export function emitStage1Package(artifact, ast) {
    const ctx = new EmitterContext({ artifact });
    const facts = packageEmitFacts(ast, artifact);
    const wasmLowerings = new Map();
    for (const fn of ast.functions) {
        const lowering = wasmScalarLoweringForFunction(fn);
        if (lowering)
            wasmLowerings.set(fn, lowering);
    }
    const usesWasm = wasmLowerings.size > 0;
    const wasmBase64 = usesWasm ? checkedInWasmStencil("i64.scalar.add").wasmBase64 : undefined;
    const declarationItems = [];
    const functionLines = [];
    const initNames = [];
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
        if (statement.kind === "TypeDecl")
            continue;
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
        ...orderTopLevelDeclarationEmissions(declarationItems).flatMap((item) => item.lines),
        ...initNames.map((name) => `  await ${name}();`)
    ];
    const javascript = stage1JavaScript(artifact, usesWasm, bodyLines, typeDescriptorLines);
    return {
        diagnostics: ctx.diagnostics,
        javascript,
        ...(wasmBase64 ? { wasmBase64 } : {})
    };
}
function topLevelValueDeclarationNames(ast) {
    const names = new Set();
    for (const statement of ast.body) {
        if (statement.kind !== "ConstDecl" && statement.kind !== "VarDecl")
            continue;
        for (const declaration of statement.declarations)
            names.add(declaration.name);
    }
    return names;
}
function emitConstDecl(ctx, statement, facts, topLevelNames, functionDependencies, nextOrder) {
    const items = [];
    let inheritedValues = [];
    for (const group of declarationGroups(statement.declarations)) {
        const groupValues = declarationValueExpressions(group);
        if (groupValues.some(Boolean))
            inheritedValues = groupValues;
        for (const declaration of group) {
            const effectiveValue = declaration.value ?? inheritedValues[declaration.valueIndex ?? 0];
            items.push(emitDeclaration(ctx, declaration, facts, topLevelNames, functionDependencies, nextOrder(), effectiveValue));
        }
    }
    return items;
}
function emitVarDecl(ctx, statement, facts, topLevelNames, functionDependencies, nextOrder) {
    return statement.declarations.map((declaration) => emitDeclaration(ctx, declaration, facts, topLevelNames, functionDependencies, nextOrder()));
}
function emitDeclaration(ctx, declaration, facts, topLevelNames, functionDependencies, order, effectiveValue = declaration.value) {
    const env = {
        ...emptyExpressionEnv,
        facts,
        ...(declaration.iotaIndex !== undefined ? { iotaValue: declaration.iotaIndex } : {})
    };
    const rawValue = effectiveValue
        ? expressionToJs(ctx, effectiveValue, env)
        : zeroValueForType(declaration.type?.text, facts);
    if (!rawValue) {
        ctx.emitError(`unsupported Stage 1 declaration for ${declaration.name}`);
        return { name: declaration.name, order, lines: [], dependencies: new Set() };
    }
    const targetType = declarationTypeText(declaration, effectiveValue);
    const value = valueForTargetType(rawValue, targetType, expressionTypeText(effectiveValue, env), env);
    return {
        name: declaration.name,
        order,
        lines: [`  pkg[${JSON.stringify(declaration.name)}] = ${value};`],
        dependencies: expressionDependencies(effectiveValue, topLevelNames, facts, declaration.name, functionDependencies)
    };
}
function orderTopLevelDeclarationEmissions(items) {
    const remaining = [...items].sort((left, right) => left.order - right.order);
    const ordered = [];
    while (remaining.length > 0) {
        const remainingNames = new Set(remaining.map((item) => item.name));
        const readyIndex = remaining.findIndex((item) => [...item.dependencies].every((dependency) => !remainingNames.has(dependency)));
        if (readyIndex < 0) {
            ordered.push(...remaining);
            break;
        }
        const [ready] = remaining.splice(readyIndex, 1);
        if (ready)
            ordered.push(ready);
    }
    return ordered;
}
function expressionDependencies(expression, topLevelNames, facts, selfName, functionDependencies = new Map()) {
    const dependencies = new Set();
    const namesAndFunctions = functionDependencies.size > 0
        ? new Set([...topLevelNames, ...functionDependencies.keys()])
        : topLevelNames;
    collectExpressionDependencies(expression, namesAndFunctions, facts, dependencies);
    expandFunctionDependencies(dependencies, functionDependencies);
    dependencies.delete(selfName);
    return dependencies;
}
function topLevelFunctionDependencyMap(ast, topLevelNames, facts) {
    const functionNames = new Set();
    const functions = new Map();
    for (const fn of ast.functions) {
        if (fn.receiver || fn.name === "init")
            continue;
        functionNames.add(fn.name);
        functions.set(fn.name, fn);
    }
    const namesAndFunctions = new Set([...topLevelNames, ...functionNames]);
    const direct = new Map();
    for (const [name, fn] of functions.entries()) {
        const dependencies = new Set();
        for (const statement of fn.body.statements)
            collectStatementDependencies(statement, namesAndFunctions, facts, dependencies);
        dependencies.delete(name);
        direct.set(name, dependencies);
    }
    const expanded = new Map();
    const expand = (name, visiting = new Set()) => {
        const cached = expanded.get(name);
        if (cached)
            return cached;
        if (visiting.has(name))
            return new Set();
        visiting.add(name);
        const out = new Set();
        for (const dependency of direct.get(name) ?? []) {
            if (functionNames.has(dependency)) {
                for (const transitive of expand(dependency, visiting))
                    out.add(transitive);
            }
            else if (topLevelNames.has(dependency)) {
                out.add(dependency);
            }
        }
        visiting.delete(name);
        expanded.set(name, out);
        return out;
    };
    for (const name of functionNames)
        expand(name);
    return expanded;
}
function expandFunctionDependencies(dependencies, functionDependencies) {
    const functionNames = [...dependencies].filter((dependency) => functionDependencies.has(dependency));
    for (const name of functionNames) {
        dependencies.delete(name);
        for (const dependency of functionDependencies.get(name) ?? [])
            dependencies.add(dependency);
    }
}
function collectExpressionDependencies(expression, topLevelNames, facts, dependencies) {
    if (!expression)
        return;
    switch (expression.kind) {
        case "Identifier":
            if (topLevelNames.has(expression.name))
                dependencies.add(expression.name);
            return;
        case "Literal":
        case "TypeExpression":
            return;
        case "SelectorExpression":
            if (expression.object.kind === "Identifier" && facts.imports.has(expression.object.name))
                return;
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
            for (const arg of expression.args)
                collectExpressionDependencies(arg, topLevelNames, facts, dependencies);
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
            for (const statement of expression.body.statements)
                collectStatementDependencies(statement, topLevelNames, facts, dependencies);
            return;
        default:
            return;
    }
}
function collectStatementDependencies(statement, topLevelNames, facts, dependencies) {
    if (!statement)
        return;
    switch (statement.kind) {
        case "ExpressionStatement":
            collectExpressionDependencies(statement.expression, topLevelNames, facts, dependencies);
            return;
        case "LabeledStatement":
            collectStatementDependencies(statement.statement, topLevelNames, facts, dependencies);
            return;
        case "AssignStatement":
            for (const target of statement.targets)
                collectExpressionDependencies(target, topLevelNames, facts, dependencies);
            for (const value of statement.values)
                collectExpressionDependencies(value, topLevelNames, facts, dependencies);
            return;
        case "ShortVarStatement":
            for (const value of statement.values)
                collectExpressionDependencies(value, topLevelNames, facts, dependencies);
            return;
        case "ReturnStatement":
            for (const value of statement.values)
                collectExpressionDependencies(value, topLevelNames, facts, dependencies);
            return;
        case "BlockStatement":
            for (const child of statement.statements)
                collectStatementDependencies(child, topLevelNames, facts, dependencies);
            return;
        case "ConstDecl":
        case "VarDecl":
            for (const declaration of statement.declarations)
                collectExpressionDependencies(declaration.value, topLevelNames, facts, dependencies);
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
                for (const value of clause.values)
                    collectExpressionDependencies(value, topLevelNames, facts, dependencies);
                for (const child of clause.statements)
                    collectStatementDependencies(child, topLevelNames, facts, dependencies);
            }
            return;
        case "SelectStatement":
            for (const clause of statement.clauses) {
                collectStatementDependencies(clause.comm, topLevelNames, facts, dependencies);
                for (const child of clause.statements)
                    collectStatementDependencies(child, topLevelNames, facts, dependencies);
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
function declarationGroups(declarations) {
    const groups = [];
    let currentGroup = [];
    let currentKey;
    for (const declaration of declarations) {
        const key = declaration.valueGroup;
        if (currentGroup.length > 0 && key !== currentKey) {
            groups.push(currentGroup);
            currentGroup = [];
        }
        currentKey = key;
        currentGroup.push(declaration);
    }
    if (currentGroup.length > 0)
        groups.push(currentGroup);
    return groups;
}
function declarationValueExpressions(group) {
    const out = [];
    for (const declaration of group) {
        if (declaration.value)
            out[declaration.valueIndex ?? 0] = declaration.value;
    }
    return out;
}
function emitFunction(ctx, fn, facts, options = {}) {
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
    if (hasDefer)
        bodyEnv.deferName = ctx.symbol("defer");
    const resultLines = emitNamedResultDeclarations(fn, bodyEnv);
    const statementLines = statementsRequireGotoStateMachine(fn.body.statements)
        ? emitGotoStateMachineStatements(ctx, fn, bodyEnv, "    ")
        : emitStatements(ctx, fn.body.statements, bodyEnv, "    ");
    if (!statementLines)
        return [];
    const hasExplicitReturn = functionBodyAlwaysReturns(fn.body);
    const target = options.localName ? `const ${options.localName}` : `pkg[${JSON.stringify(functionPackageKey(fn))}]`;
    const firstLine = hasDefer
        ? `  ${target} = async (${parameters.params.join(", ")}) => __gojrDeferScope(async (${bodyEnv.deferName}) => {`
        : `  ${target} = async (${parameters.params.join(", ")}) => {`;
    const lastLine = hasDefer ? "  });" : "  };";
    return [
        firstLine,
        ...resultLines.map((line) => `    ${line}`),
        ...statementLines,
        ...(hasExplicitReturn ? [] : [`    return ${defaultFunctionReturn(fn, bodyEnv)};`]),
        lastLine
    ];
}
function returnExpression(ctx, statement, env) {
    if (statement.values.length === 0)
        return namedResultReturn(env) ?? "null";
    if (statement.values.length > 1) {
        const values = statement.values.map((value, index) => {
            const rendered = expressionToJs(ctx, value, env);
            return rendered ? valueForTargetType(rendered, env.expectedReturnTypes?.[index], expressionTypeText(value, env), env) : undefined;
        });
        if (values.some((value) => value === undefined))
            return undefined;
        return `__gojrTuple([${values.filter((value) => value !== undefined).join(", ")}])`;
    }
    const rendered = expressionToJs(ctx, statement.values[0], env);
    if (!rendered)
        return undefined;
    if ((env.expectedReturnTypes?.length ?? 0) > 1) {
        return returnTupleExpression(rendered, env.expectedReturnTypes ?? [], env);
    }
    return valueForTargetType(rendered, env.expectedReturnTypes?.[0], expressionTypeText(statement.values[0], env), env);
}
function returnTupleExpression(rendered, expectedTypes, env) {
    const valuesName = "__gojrReturnValues";
    const converted = expectedTypes.map((typeText, index) => valueForTargetType(`${valuesName}[${index}]`, typeText, undefined, env));
    return `__gojrTuple(((${valuesName}) => [${converted.join(", ")}])(__gojrTupleValues(${rendered})))`;
}
function emitNamedResultDeclarations(fn, env) {
    const namedResults = [];
    const lines = [];
    for (const [index, result] of fn.signature.results.entries()) {
        if (!result.name || result.name === "_")
            continue;
        const name = safeLocalName(result.name, index);
        env.locals.set(result.name, name);
        env.localTypes.set(result.name, result.type.text);
        namedResults.push(name);
        lines.push(`let ${name} = ${zeroValueForTypeInEnv(result.type.text, env) ?? "null"};`);
    }
    if (namedResults.length > 0)
        env.namedResults = namedResults;
    return lines;
}
function namedResultReturn(env) {
    if (!env.namedResults || env.namedResults.length === 0)
        return undefined;
    if (env.namedResults.length === 1)
        return env.namedResults[0];
    return `__gojrTuple([${env.namedResults.join(", ")}])`;
}
function defaultFunctionReturn(fn, env) {
    const named = namedResultReturn(env);
    if (named)
        return named;
    if (fn.signature.results.length === 0)
        return "null";
    if (fn.signature.results.length === 1)
        return zeroValueForTypeInEnv(fn.signature.results[0]?.type.text, env) ?? "null";
    return `__gojrTuple([${fn.signature.results.map((result) => zeroValueForTypeInEnv(result.type.text, env) ?? "null").join(", ")}])`;
}
function functionBodyAlwaysReturns(body) {
    const last = body.statements[body.statements.length - 1];
    return last?.kind === "ReturnStatement";
}
function statementsContainDefer(statements) {
    return statements.some(statementContainsDefer);
}
function statementContainsDefer(statement) {
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
function statementContainsDeferInIf(statement) {
    if (statementContainsDefer(statement.thenBlock))
        return true;
    if (!statement.elseBranch)
        return false;
    return statement.elseBranch.kind === "IfStatement"
        ? statementContainsDeferInIf(statement.elseBranch)
        : statementContainsDefer(statement.elseBranch);
}
function statementsContainGoto(statements) {
    return statements.some(statementContainsGoto);
}
function statementsRequireGotoStateMachine(statements) {
    return statementsContainGoto(statements) && statements.some((statement) => statement.kind === "LabeledStatement");
}
function statementContainsGoto(statement) {
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
function statementContainsGotoTo(statement, label) {
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
function statementsContainGotoTo(statements, label) {
    return statements.some((statement) => statementContainsGotoTo(statement, label));
}
function emitGotoStateMachineStatements(ctx, fn, env, indent) {
    const labels = new Map();
    for (const [index, statement] of fn.body.statements.entries()) {
        if (statement.kind === "LabeledStatement")
            labels.set(statement.label, index);
    }
    const hoisted = hoistedGotoLocals(fn.body.statements);
    for (const [name, typeText] of hoisted.entries()) {
        env.locals.set(name, safeLocalName(name, 0));
        if (typeText)
            env.localTypes.set(name, typeText);
    }
    const pc = ctx.symbol("pc");
    env.gotoLabels = labels;
    env.gotoPcName = pc;
    const lines = [];
    for (const name of hoisted.keys()) {
        lines.push(`${indent}let ${env.locals.get(name)};`);
    }
    lines.push(`${indent}let ${pc} = 0;`);
    lines.push(`${indent}__gojrGotoLoop: while (true) {`);
    lines.push(`${indent}  switch (${pc}) {`);
    for (const [index, statement] of fn.body.statements.entries()) {
        const active = statement.kind === "LabeledStatement" && statement.statement ? statement.statement : statement;
        const emitted = emitStatement(ctx, active, env, `${indent}      `);
        if (!emitted)
            return undefined;
        lines.push(`${indent}    case ${index}:`);
        lines.push(...emitted);
        lines.push(`${indent}      ${pc} = ${index + 1};`);
        lines.push(`${indent}      continue __gojrGotoLoop;`);
    }
    lines.push(`${indent}    default:`);
    lines.push(`${indent}      return ${defaultFunctionReturn(fn, env)};`);
    lines.push(`${indent}  }`);
    lines.push(`${indent}}`);
    return lines;
}
function hoistedGotoLocals(statements) {
    const out = new Map();
    for (const statement of statements) {
        if (statement.kind === "ShortVarStatement") {
            for (const [index, name] of statement.names.entries()) {
                if (name !== "_")
                    out.set(name, expressionTypeText(statement.values[index], emptyExpressionEnv));
            }
        }
        if (statement.kind === "VarDecl") {
            for (const declaration of statement.declarations) {
                if (declaration.name !== "_")
                    out.set(declaration.name, declarationTypeText(declaration));
            }
        }
        if (statement.kind === "LabeledStatement" && statement.statement?.kind === "ShortVarStatement") {
            for (const [index, name] of statement.statement.names.entries()) {
                if (name !== "_")
                    out.set(name, expressionTypeText(statement.statement.values[index], emptyExpressionEnv));
            }
        }
    }
    return out;
}
function cloneExpressionEnv(env) {
    return {
        locals: new Map(env.locals),
        localTypes: new Map(env.localTypes),
        facts: env.facts,
        ...(env.namedResults ? { namedResults: [...env.namedResults] } : {}),
        ...(env.expectedReturnTypes ? { expectedReturnTypes: [...env.expectedReturnTypes] } : {}),
        ...(env.typeParameters ? { typeParameters: new Set(env.typeParameters) } : {}),
        ...(env.deferName ? { deferName: env.deferName } : {}),
        ...(env.gotoLabels ? { gotoLabels: env.gotoLabels } : {}),
        ...(env.gotoPcName ? { gotoPcName: env.gotoPcName } : {}),
        ...(env.gotoBreakLabels ? { gotoBreakLabels: new Map(env.gotoBreakLabels) } : {}),
        ...(env.iotaValue !== undefined ? { iotaValue: env.iotaValue } : {})
    };
}
function emitStatements(ctx, statements, env, indent) {
    const localForward = findLocalForwardGoto(statements);
    if (localForward)
        return emitLocalForwardGotoStatements(ctx, statements, env, indent, localForward);
    const lines = [];
    for (const statement of statements) {
        const emitted = emitStatement(ctx, statement, env, indent);
        if (!emitted)
            return undefined;
        lines.push(...emitted);
    }
    return lines;
}
function findLocalForwardGoto(statements) {
    for (const [index, statement] of statements.entries()) {
        if (statement.kind !== "LabeledStatement")
            continue;
        const prior = statements.slice(0, index);
        if (statementsContainGotoTo(prior, statement.label))
            return { label: statement.label, index };
    }
    return undefined;
}
function emitLocalForwardGotoStatements(ctx, statements, env, indent, target) {
    const jsLabel = ctx.symbol(`label_${target.label}`);
    const beforeEnv = cloneExpressionEnv(env);
    beforeEnv.gotoBreakLabels = new Map(beforeEnv.gotoBreakLabels);
    beforeEnv.gotoBreakLabels.set(target.label, jsLabel);
    const before = emitStatements(ctx, statements.slice(0, target.index), beforeEnv, `${indent}  `);
    if (!before)
        return undefined;
    const labeled = statements[target.index];
    const labelStatement = labeled.statement ? emitStatement(ctx, labeled.statement, env, indent) : [];
    if (!labelStatement)
        return undefined;
    const after = emitStatements(ctx, statements.slice(target.index + 1), env, indent);
    if (!after)
        return undefined;
    return [
        `${indent}${jsLabel}: {`,
        ...before,
        `${indent}}`,
        ...labelStatement,
        ...after
    ];
}
function emitStatement(ctx, statement, env, indent) {
    switch (statement.kind) {
        case "BlockStatement":
            return emitBlockStatement(ctx, statement, env, indent);
        case "LabeledStatement":
            return emitLabeledStatement(ctx, statement, env, indent);
        case "ConstDecl":
        case "VarDecl":
            return emitLocalDeclarationStatement(ctx, statement, env, indent);
        case "TypeDecl":
            return [];
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
            ctx.emitError(`unsupported Stage 4 statement ${statement.kind}`);
            return undefined;
    }
}
function emitBlockStatement(ctx, statement, env, indent) {
    const blockEnv = cloneExpressionEnv(env);
    const body = emitStatements(ctx, statement.statements, blockEnv, `${indent}  `);
    if (!body)
        return undefined;
    return [
        `${indent}{`,
        ...body,
        `${indent}}`
    ];
}
function emitLabeledStatement(ctx, statement, env, indent) {
    if (!statement.statement)
        return [`${indent}${statement.label}: {}`];
    const emitted = emitStatement(ctx, statement.statement, env, indent);
    if (!emitted)
        return undefined;
    return [`${indent}${statement.label}:`, ...emitted];
}
function emitLocalDeclarationStatement(ctx, statement, env, indent) {
    const lines = [];
    const keyword = statement.kind === "ConstDecl" ? "const" : "let";
    let inheritedValues = [];
    let localIndex = 0;
    for (const group of declarationGroups(statement.declarations)) {
        const groupValues = statement.kind === "ConstDecl" ? declarationValueExpressions(group) : [];
        if (groupValues.some(Boolean))
            inheritedValues = groupValues;
        for (const declaration of group) {
            if (declaration.name === "_")
                continue;
            const name = safeLocalName(declaration.name, localIndex++);
            const declarationEnv = {
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
            env.locals.set(declaration.name, name);
            const typeText = declarationTypeText(declaration, effectiveValue);
            if (typeText)
                env.localTypes.set(declaration.name, typeText);
            const value = valueForTargetType(rawValue, typeText, expressionTypeText(effectiveValue, declarationEnv), declarationEnv);
            lines.push(`${indent}${keyword} ${name} = ${value};`);
        }
    }
    return lines;
}
function emitReturnStatement(ctx, statement, env, indent) {
    const returned = returnExpression(ctx, statement, env);
    if (!returned)
        return undefined;
    return [`${indent}return ${returned};`];
}
function emitIfStatement(ctx, statement, env, indent) {
    const scoped = Boolean(statement.init);
    const ifEnv = cloneExpressionEnv(env);
    const lines = [];
    if (scoped)
        lines.push(`${indent}{`);
    const activeIndent = scoped ? `${indent}  ` : indent;
    if (statement.init) {
        const initLines = emitStatement(ctx, statement.init, ifEnv, activeIndent);
        if (!initLines)
            return undefined;
        lines.push(...initLines);
    }
    const condition = expressionToJs(ctx, statement.condition, ifEnv);
    if (!condition)
        return undefined;
    const thenEnv = cloneExpressionEnv(ifEnv);
    const thenLines = emitStatements(ctx, statement.thenBlock.statements, thenEnv, `${activeIndent}  `);
    if (!thenLines)
        return undefined;
    lines.push(`${activeIndent}if (${condition}) {`);
    lines.push(...thenLines);
    if (statement.elseBranch) {
        if (statement.elseBranch.kind === "IfStatement") {
            const elseLines = emitIfStatement(ctx, statement.elseBranch, cloneExpressionEnv(ifEnv), activeIndent);
            if (!elseLines)
                return undefined;
            const [first, ...rest] = elseLines;
            if (!first)
                return undefined;
            lines.push(`${activeIndent}} else ${first.trimStart()}`);
            lines.push(...rest);
        }
        else {
            const elseEnv = cloneExpressionEnv(ifEnv);
            const elseLines = emitStatements(ctx, statement.elseBranch.statements, elseEnv, `${activeIndent}  `);
            if (!elseLines)
                return undefined;
            lines.push(`${activeIndent}} else {`);
            lines.push(...elseLines);
            lines.push(`${activeIndent}}`);
        }
    }
    else {
        lines.push(`${activeIndent}}`);
    }
    if (scoped)
        lines.push(`${indent}}`);
    return lines;
}
function emitForStatement(ctx, statement, env, indent) {
    if (statement.range)
        return emitRangeStatement(ctx, statement, env, indent);
    const loopEnv = cloneExpressionEnv(env);
    const init = statement.init ? emitForHeaderStatement(ctx, statement.init, loopEnv, "init") : "";
    const condition = statement.condition ? expressionToJs(ctx, statement.condition, loopEnv) : "";
    const post = statement.post ? emitForHeaderStatement(ctx, statement.post, loopEnv, "post") : "";
    if (init === undefined || condition === undefined || post === undefined)
        return undefined;
    const bodyEnv = cloneExpressionEnv(loopEnv);
    const body = emitStatements(ctx, statement.body.statements, bodyEnv, `${indent}  `);
    if (!body)
        return undefined;
    return [
        `${indent}for (${init}; ${condition}; ${post}) {`,
        ...body,
        `${indent}}`
    ];
}
function emitRangeStatement(ctx, statement, env, indent) {
    if (!statement.range)
        return undefined;
    const source = expressionToJs(ctx, statement.range.source, env);
    if (!source)
        return undefined;
    const key = ctx.symbol("key");
    const value = ctx.symbol("value");
    const loopEnv = cloneExpressionEnv(env);
    const bindLines = [];
    const bindRangePart = (name, target, part, index) => {
        if (name !== undefined) {
            if (name === "_")
                return [];
            const jsName = safeLocalName(name, index);
            if (statement.range?.define || !loopEnv.locals.has(name)) {
                loopEnv.locals.set(name, jsName);
                const rangeTypes = rangeIterationTypeTexts(statement.range?.source, env);
                if (name === statement.range?.keyName && rangeTypes.key)
                    loopEnv.localTypes.set(name, rangeTypes.key);
                if (name === statement.range?.valueName && rangeTypes.value)
                    loopEnv.localTypes.set(name, rangeTypes.value);
                return [`${indent}  let ${jsName} = ${part};`];
            }
            return [`${indent}  ${loopEnv.locals.get(name)} = ${part};`];
        }
        if (!target)
            return [];
        return assignExpressionTargetLines(ctx, target, part, loopEnv, `${indent}  `);
    };
    const keyLines = bindRangePart(statement.range.keyName, statement.range.keyTarget, key, 0);
    const valueLines = bindRangePart(statement.range.valueName, statement.range.valueTarget, value, 1);
    if (!keyLines || !valueLines)
        return undefined;
    bindLines.push(...keyLines, ...valueLines);
    const body = emitStatements(ctx, statement.body.statements, cloneExpressionEnv(loopEnv), `${indent}  `);
    if (!body)
        return undefined;
    return [
        `${indent}for (const [${key}, ${value}] of __gojrRangeEntries(${source})) {`,
        ...bindLines,
        ...body,
        `${indent}}`
    ];
}
function emitSwitchStatement(ctx, statement, env, indent) {
    if (statement.typeSwitch) {
        return emitTypeSwitchStatement(ctx, statement, env, indent);
    }
    const scoped = Boolean(statement.init);
    const switchEnv = cloneExpressionEnv(env);
    const lines = [];
    if (scoped)
        lines.push(`${indent}{`);
    const activeIndent = scoped ? `${indent}  ` : indent;
    if (statement.init) {
        const initLines = emitStatement(ctx, statement.init, switchEnv, activeIndent);
        if (!initLines)
            return undefined;
        lines.push(...initLines);
    }
    const switchValue = ctx.symbol("switch");
    const value = statement.expression ? expressionToJs(ctx, statement.expression, switchEnv) : "true";
    if (!value)
        return undefined;
    lines.push(`${activeIndent}const ${switchValue} = ${value};`);
    lines.push(`${activeIndent}switch (true) {`);
    for (const clause of statement.clauses) {
        if (clause.default) {
            lines.push(`${activeIndent}  default: {`);
        }
        else {
            if (clause.values.length === 0) {
                ctx.emitError("unsupported Stage 4 switch clause without values");
                return undefined;
            }
            for (const expression of clause.values) {
                const clauseValue = expressionToJs(ctx, expression, switchEnv);
                if (!clauseValue)
                    return undefined;
                lines.push(`${activeIndent}  case __gojrEqual(${switchValue}, ${clauseValue}):`);
            }
            lines.push(`${activeIndent}  {`);
        }
        const fallthrough = clause.statements[clause.statements.length - 1]?.kind === "BranchStatement" &&
            clause.statements[clause.statements.length - 1].branch === "fallthrough";
        if (clause.statements.some((item, index) => item.kind === "BranchStatement" && item.branch === "fallthrough" && index !== clause.statements.length - 1)) {
            ctx.emitError("fallthrough must be the final statement in a switch clause");
            return undefined;
        }
        const bodyStatements = fallthrough ? clause.statements.slice(0, -1) : clause.statements;
        const body = emitStatements(ctx, bodyStatements, cloneExpressionEnv(switchEnv), `${activeIndent}    `);
        if (!body)
            return undefined;
        lines.push(...body);
        if (!fallthrough)
            lines.push(`${activeIndent}    break;`);
        lines.push(`${activeIndent}  }`);
    }
    lines.push(`${activeIndent}}`);
    if (scoped)
        lines.push(`${indent}}`);
    return lines;
}
function emitTypeSwitchStatement(ctx, statement, env, indent) {
    if (!statement.typeSwitch)
        return undefined;
    const scoped = Boolean(statement.init);
    const switchEnv = cloneExpressionEnv(env);
    const lines = [];
    if (scoped)
        lines.push(`${indent}{`);
    const activeIndent = scoped ? `${indent}  ` : indent;
    if (statement.init) {
        const initLines = emitStatement(ctx, statement.init, switchEnv, activeIndent);
        if (!initLines)
            return undefined;
        lines.push(...initLines);
    }
    const switchValue = ctx.symbol("typeswitch");
    const value = expressionToJs(ctx, statement.typeSwitch.expression, switchEnv);
    if (!value)
        return undefined;
    lines.push(`${activeIndent}const ${switchValue} = ${value};`);
    lines.push(`${activeIndent}switch (true) {`);
    for (const clause of statement.clauses) {
        if (clause.statements.some((item) => item.kind === "BranchStatement" && item.branch === "fallthrough")) {
            ctx.emitError("fallthrough is not allowed in type switches");
            return undefined;
        }
        if (clause.default) {
            lines.push(`${activeIndent}  default: {`);
        }
        else {
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
        if (!binding)
            return undefined;
        const body = emitStatements(ctx, clause.statements, clauseEnv, `${activeIndent}    `);
        if (!body)
            return undefined;
        lines.push(...binding, ...body, `${activeIndent}    break;`, `${activeIndent}  }`);
    }
    lines.push(`${activeIndent}}`);
    if (scoped)
        lines.push(`${indent}}`);
    return lines;
}
function emitTypeSwitchBinding(statement, clause, env, switchValue, indent) {
    const guard = statement.typeSwitch;
    if (!guard?.name || guard.name === "_")
        return [];
    const singleType = !clause.default && (clause.typeValues?.length ?? 0) === 1 ? clause.typeValues?.[0]?.text : undefined;
    const value = singleType ? `__gojrTypeAssert(${switchValue}, ${JSON.stringify(singleType)})` : switchValue;
    if (guard.define || !env.locals.has(guard.name)) {
        const local = safeLocalName(guard.name, 0);
        env.locals.set(guard.name, local);
        if (singleType)
            env.localTypes.set(guard.name, singleType);
        return [`${indent}let ${local} = ${value};`];
    }
    return [`${indent}${env.locals.get(guard.name)} = ${value};`];
}
function emitBranchStatement(ctx, statement, env, indent) {
    if (statement.branch === "goto") {
        const breakLabel = statement.label ? env.gotoBreakLabels?.get(statement.label) : undefined;
        if (breakLabel)
            return [`${indent}break ${breakLabel};`];
        const target = statement.label ? env.gotoLabels?.get(statement.label) : undefined;
        if (!env.gotoPcName || target === undefined) {
            ctx.emitError(`unsupported Stage 4 unresolved goto ${statement.label ?? "<missing>"}`);
            return undefined;
        }
        return [
            `${indent}${env.gotoPcName} = ${target};`,
            `${indent}continue __gojrGotoLoop;`
        ];
    }
    if (statement.branch === "fallthrough")
        return [`${indent}/* fallthrough */`];
    return [`${indent}${statement.branch}${statement.label ? ` ${statement.label}` : ""};`];
}
function emitAssignStatement(ctx, statement, env, indent) {
    if (statement.operator !== "=") {
        if (statement.targets.length !== 1 || statement.values.length !== 1) {
            ctx.emitError(`unsupported Stage 4 compound assignment shape`);
            return undefined;
        }
        const current = expressionToJs(ctx, statement.targets[0], env);
        const value = expressionToJs(ctx, statement.values[0], env);
        if (!current || !value)
            return undefined;
        const compound = compoundAssignmentExpression(ctx, statement.operator, current, value);
        if (!compound)
            return undefined;
        return assignExpressionTargetLines(ctx, statement.targets[0], compound, env, indent, expressionTypeText(statement.values[0], env));
    }
    if (statement.values.length === 1 && statement.targets.length > 1) {
        const source = statement.values[0];
        if (!source)
            return undefined;
        const tuple = tupleSourceExpressionToJs(ctx, source, statement.targets.length, env);
        if (!tuple)
            return undefined;
        const temp = ctx.symbol("assign");
        const lines = [`${indent}const ${temp} = ${tuple};`];
        for (const [index, target] of statement.targets.entries()) {
            const assigned = assignExpressionTargetLines(ctx, target, `${temp}[${index}]`, env, indent, tupleSourceTypeTexts(source, statement.targets.length, env)[index]);
            if (!assigned)
                return undefined;
            lines.push(...assigned);
        }
        return lines;
    }
    if (statement.values.length !== statement.targets.length) {
        ctx.emitError(`unsupported Stage 4 assignment count ${statement.targets.length} targets and ${statement.values.length} values`);
        return undefined;
    }
    const values = statement.values.map((value) => expressionToJs(ctx, value, env));
    if (values.some((value) => value === undefined))
        return undefined;
    const sourceTypes = statement.values.map((value) => expressionTypeText(value, env));
    const temp = ctx.symbol("assign");
    const lines = [`${indent}const ${temp} = [${values.filter((value) => value !== undefined).join(", ")}];`];
    for (const [index, target] of statement.targets.entries()) {
        const assigned = assignExpressionTargetLines(ctx, target, `${temp}[${index}]`, env, indent, sourceTypes[index]);
        if (!assigned)
            return undefined;
        lines.push(...assigned);
    }
    return lines;
}
function emitShortVarStatement(ctx, statement, env, indent) {
    if (statement.values.length === 1 && statement.names.length > 1) {
        const source = statement.values[0];
        if (!source)
            return undefined;
        const tuple = tupleSourceExpressionToJs(ctx, source, statement.names.length, env);
        if (!tuple)
            return undefined;
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
    if (values.some((value) => value === undefined))
        return undefined;
    const temp = ctx.symbol("short");
    const sourceTypes = statement.values.map((value) => expressionTypeText(value, env));
    return [
        `${indent}const ${temp} = [${values.filter((value) => value !== undefined).join(", ")}];`,
        ...bindShortNames(statement.names, temp, env, indent, sourceTypes)
    ];
}
function emitIncDecStatement(ctx, statement, env, indent) {
    const current = expressionToJs(ctx, statement.target, env);
    if (!current)
        return undefined;
    const one = isIntegerType(statement.target.typeText ?? "") ? "1n" : "1";
    const next = statement.operator === "++" ? `((${current}) + ${one})` : `((${current}) - ${one})`;
    return assignExpressionTargetLines(ctx, statement.target, next, env, indent);
}
function emitExpressionStatement(ctx, statement, env, indent) {
    const expression = expressionToJs(ctx, statement.expression, env);
    if (!expression)
        return undefined;
    return [`${indent}${expression};`];
}
function emitDeferStatement(ctx, statement, env, indent) {
    if (!env.deferName) {
        ctx.emitError("defer used outside generated function defer scope");
        return undefined;
    }
    if (statement.expression.kind === "CallExpression") {
        const callee = deferredCalleeToJs(ctx, statement.expression, env);
        if (!callee)
            return undefined;
        if (statement.expression.spreadLast) {
            ctx.emitError("unsupported Stage 4 deferred spread call");
            return undefined;
        }
        const args = statement.expression.args.map((arg) => expressionToJs(ctx, arg, env));
        if (args.some((arg) => arg === undefined))
            return undefined;
        const temp = ctx.symbol("deferArgs");
        return [
            `${indent}const ${temp} = [${args.filter((arg) => arg !== undefined).join(", ")}];`,
            `${indent}${env.deferName}(async () => await (${callee})(...${temp}));`
        ];
    }
    const expression = expressionToJs(ctx, statement.expression, env);
    if (!expression)
        return undefined;
    return [`${indent}${env.deferName}(async () => ${expression});`];
}
function emitGoStatement(ctx, statement, env, indent) {
    const call = callExpressionToJs(ctx, statement.call, env);
    if (!call)
        return undefined;
    return [`${indent}__gojrGo(async () => { ${call}; });`];
}
function emitSendStatement(ctx, statement, env, indent) {
    const channel = expressionToJs(ctx, statement.channel, env);
    const value = expressionToJs(ctx, statement.value, env);
    if (!channel || !value)
        return undefined;
    return [`${indent}await __gojrChanSend(${channel}, ${value});`];
}
function emitSelectStatement(ctx, statement, env, indent) {
    const cases = [];
    for (const clause of statement.clauses) {
        const clauseEnv = cloneExpressionEnv(env);
        const selected = ctx.symbol("selected");
        const bodyLines = [];
        const prepared = emitSelectComm(ctx, clause.comm, clause.default, clauseEnv, `${indent}      `, selected);
        if (!prepared)
            return undefined;
        bodyLines.push(...prepared.prefixLines);
        const emittedBody = emitStatements(ctx, clause.statements, clauseEnv, `${indent}      `);
        if (!emittedBody)
            return undefined;
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
function emitSelectComm(ctx, statement, isDefault, env, indent, selected) {
    if (isDefault || !statement)
        return { op: "default", prefixLines: [] };
    if (statement.kind === "SendStatement") {
        const channel = expressionToJs(ctx, statement.channel, env);
        const value = expressionToJs(ctx, statement.value, env);
        if (!channel || !value)
            return undefined;
        return { op: "send", channel, value, prefixLines: [] };
    }
    const receive = selectReceiveOperand(statement);
    if (!receive) {
        ctx.emitError(`unsupported Stage 4 select communication statement ${statement.kind}`);
        return undefined;
    }
    const channel = expressionToJs(ctx, receive, env);
    if (!channel)
        return undefined;
    if (statement.kind === "ExpressionStatement")
        return { op: "recv", channel, prefixLines: [] };
    if (statement.kind === "ShortVarStatement") {
        return {
            op: "recv",
            channel,
            prefixLines: bindShortNames(statement.names, selected, env, indent, selectReceiveTypeTexts(receive, statement.names.length, env))
        };
    }
    if (statement.kind === "AssignStatement") {
        const prefixLines = [];
        for (const [index, target] of statement.targets.entries()) {
            const assigned = assignExpressionTargetLines(ctx, target, `${selected}[${index}]`, env, indent);
            if (!assigned)
                return undefined;
            prefixLines.push(...assigned);
        }
        return { op: "recv", channel, prefixLines };
    }
    ctx.emitError(`unsupported Stage 4 select receive statement ${statement.kind}`);
    return undefined;
}
function selectReceiveOperand(statement) {
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
function selectReceiveTypeTexts(channel, targetCount, env) {
    const elementType = chanElementTypeText(expressionTypeText(channel, env));
    return targetCount === 2 ? [elementType, "bool"] : [elementType];
}
function deferredCalleeToJs(ctx, expression, env) {
    if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
        return `pkg[${JSON.stringify(expression.callee.name)}]`;
    }
    if (expression.callee.kind === "SelectorExpression") {
        const methodKey = methodPackageKeyForSelector(expression.callee, env);
        if (methodKey) {
            const receiver = expressionToJs(ctx, expression.callee.object, env);
            if (!receiver)
                return undefined;
            return `(async (...__gojrMethodArgs) => await pkg[${JSON.stringify(methodKey)}](${receiver}, ...__gojrMethodArgs))`;
        }
    }
    return expressionToJs(ctx, expression.callee, env);
}
function emitForHeaderStatement(ctx, statement, env, position) {
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
                    if (!source)
                        return undefined;
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
                        if (sourceType)
                            env.localTypes.set(goName, sourceType);
                    }
                    return `let [${names.join(", ")}] = ${tuple}`;
                }
                if (statement.names.length !== statement.values.length) {
                    ctx.emitError("unsupported Stage 4 for init short declaration");
                    return undefined;
                }
                const declarations = [];
                for (const [index, goName] of statement.names.entries()) {
                    const expression = statement.values[index];
                    if (!goName || !expression)
                        return undefined;
                    const value = expressionToJs(ctx, expression, env);
                    if (!value)
                        return undefined;
                    const name = safeLocalName(goName, index);
                    env.locals.set(goName, name);
                    const typeText = expressionTypeText(expression, env);
                    if (typeText)
                        env.localTypes.set(goName, typeText);
                    declarations.push(`${name} = ${value}`);
                }
                return `let ${declarations.join(", ")}`;
            }
        case "AssignStatement":
            if (statement.values.length === 1 && statement.targets.length > 1) {
                const source = statement.values[0];
                if (!source)
                    return undefined;
                const tuple = tupleSourceExpressionToJs(ctx, source, statement.targets.length, env);
                if (!tuple) {
                    ctx.emitError("unsupported Stage 4 for header assignment");
                    return undefined;
                }
                const targets = statement.targets.map((target) => expressionToJs(ctx, target, env));
                if (targets.some((target) => target === undefined))
                    return undefined;
                return `([${targets.filter((target) => target !== undefined).join(", ")}] = ${tuple})`;
            }
            if (statement.targets.length !== statement.values.length) {
                ctx.emitError("unsupported Stage 4 for header assignment");
                return undefined;
            }
            {
                const parts = [];
                for (const [index, targetExpression] of statement.targets.entries()) {
                    const target = expressionToJs(ctx, targetExpression, env);
                    const valueExpression = statement.values[index];
                    const value = valueExpression ? expressionToJs(ctx, valueExpression, env) : undefined;
                    if (!target || !value)
                        return undefined;
                    if (statement.operator === "=") {
                        parts.push(`${target} = ${value}`);
                        continue;
                    }
                    if (statement.targets.length !== 1) {
                        ctx.emitError("unsupported Stage 4 compound multi-assignment in for header");
                        return undefined;
                    }
                    const compound = compoundAssignmentExpression(ctx, statement.operator, target, value);
                    if (!compound)
                        return undefined;
                    parts.push(`${target} = ${compound}`);
                }
                return parts.join(", ");
            }
        case "IncDecStatement":
            {
                const target = expressionToJs(ctx, statement.target, env);
                if (!target)
                    return undefined;
                const one = isIntegerType(statement.target.typeText ?? "") ? "1n" : "1";
                return `${target} ${statement.operator === "++" ? "+=" : "-="} ${one}`;
            }
        case "ExpressionStatement":
            return expressionToJs(ctx, statement.expression, env);
        default:
            ctx.emitError(`unsupported Stage 4 for ${position} statement ${statement.kind}`);
            return undefined;
    }
}
function tupleSourceExpressionToJs(ctx, expression, targetCount, env) {
    if (targetCount === 2 && expression.kind === "IndexExpression") {
        const object = expressionToJs(ctx, expression.object, env);
        const index = expressionToJs(ctx, expression.index, env);
        if (!object || !index)
            return undefined;
        const objectType = expressionTypeText(expression.object, env) ?? packageExportTypeText(ctx, expression.object);
        const zero = zeroValueForTypeInEnv(mapValueTypeText(objectType, env.facts) ?? expression.typeText, env) ??
            "null";
        return `__gojrMapGetOk(${object}, ${index}, ${zero})`;
    }
    if (targetCount === 2 && expression.kind === "TypeAssertionExpression") {
        const value = expressionToJs(ctx, expression.expression, env);
        if (!value)
            return undefined;
        return `__gojrTypeAssertOk(${value}, ${JSON.stringify(expression.type.text)})`;
    }
    if (targetCount === 2 && expression.kind === "UnaryExpression" && expression.operator === "<-") {
        const channel = expressionToJs(ctx, expression.operand, env);
        if (!channel)
            return undefined;
        return `(await __gojrChanRecv(${channel}))`;
    }
    const value = expressionToJs(ctx, expression, env);
    if (!value)
        return undefined;
    return value;
}
function bindShortNames(names, tuple, env, indent, sourceTypes = []) {
    const lines = [];
    for (const [index, name] of names.entries()) {
        if (name === "_")
            continue;
        const existing = env.locals.get(name);
        if (existing) {
            lines.push(`${indent}${existing} = ${tuple}[${index}];`);
            continue;
        }
        const local = safeLocalName(name, index);
        env.locals.set(name, local);
        const source = sourceTypes[index];
        if (source)
            env.localTypes.set(name, source);
        lines.push(`${indent}let ${local} = ${tuple}[${index}];`);
    }
    return lines;
}
function assignExpressionTargetLines(ctx, target, value, env, indent, sourceType) {
    if (!target)
        return undefined;
    switch (target.kind) {
        case "Identifier": {
            if (target.name === "_")
                return [];
            const local = env.locals.get(target.name);
            const targetType = env.localTypes.get(target.name) ?? env.facts.packageTypes.get(target.name);
            return [`${indent}${local ?? `pkg[${JSON.stringify(target.name)}]`} = ${valueForTargetType(value, targetType, sourceType, env)};`];
        }
        case "SelectorExpression": {
            const object = expressionToJs(ctx, target.object, env);
            if (!object)
                return undefined;
            return [`${indent}__gojrSetField(${object}, ${JSON.stringify(target.field)}, ${value});`];
        }
        case "IndexExpression": {
            const object = expressionToJs(ctx, target.object, env);
            const index = expressionToJs(ctx, target.index, env);
            if (!object || !index)
                return undefined;
            const objectType = expressionTypeText(target.object, env) ?? packageExportTypeText(ctx, target.object) ?? "";
            if (parseMapTypeText(resolveUnderlyingTypeText(objectType, env.facts)))
                return [`${indent}__gojrMapSet(${object}, ${index}, ${value});`];
            return [`${indent}(${object})[Number(${index})] = ${value};`];
        }
        case "UnaryExpression": {
            if (target.operator !== "*") {
                ctx.emitError(`unsupported Stage 4 assignment target ${target.operator}${target.kind}`);
                return undefined;
            }
            const pointer = expressionToJs(ctx, target.operand, env);
            if (!pointer)
                return undefined;
            return [`${indent}__gojrSetDeref(${pointer}, ${value});`];
        }
        default:
            ctx.emitError(`unsupported Stage 4 assignment target ${target.kind}`);
            return undefined;
    }
}
function compoundAssignmentExpression(ctx, operator, left, right) {
    if (operator === "=")
        return right;
    const binary = operator.slice(0, -1);
    if (binary === "&^")
        return `((${left}) & ~(${right}))`;
    const jsOperator = jsBinaryOperator(binary);
    if (!jsOperator) {
        ctx.emitError(`unsupported Stage 4 compound operator ${operator}`);
        return undefined;
    }
    return `((${left}) ${jsOperator} (${right}))`;
}
function functionParameterBindings(fn, facts) {
    const locals = new Map();
    const localTypes = new Map();
    const params = [];
    const typeParameters = fn.typeParameters && fn.typeParameters.length > 0
        ? new Set(fn.typeParameters)
        : undefined;
    if (typeParameters)
        params.push("__gojrTypeArgs");
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
    return { params, env: { locals, localTypes, facts, ...(typeParameters ? { typeParameters } : {}) } };
}
function wasmScalarLoweringForFunction(fn) {
    if (fn.receiver)
        return undefined;
    if (fn.typeParameters && fn.typeParameters.length > 0)
        return undefined;
    if (fn.signature.parameters.length !== 2 || fn.signature.results.length !== 1)
        return undefined;
    if (!fn.signature.parameters.every((param) => param.name && param.type.text === "int64"))
        return undefined;
    const statement = onlyReturn(fn.body);
    if (!statement || statement.values.length !== 1)
        return undefined;
    const value = statement.values[0];
    if (!value || value.kind !== "BinaryExpression")
        return undefined;
    if (!isIdentifier(value.left, fn.signature.parameters[0]?.name))
        return undefined;
    if (!isIdentifier(value.right, fn.signature.parameters[1]?.name))
        return undefined;
    const resultType = fn.signature.results[0]?.type.text;
    const lowering = i64ScalarLowerings[value.operator];
    if (!lowering || lowering.resultType !== resultType)
        return undefined;
    return lowering;
}
const i64ScalarLowerings = {
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
function onlyReturn(body) {
    return body.statements.length === 1 && body.statements[0]?.kind === "ReturnStatement"
        ? body.statements[0]
        : undefined;
}
function isIdentifier(expression, name) {
    return Boolean(name && expression.kind === "Identifier" && expression.name === name);
}
function packageEmitFacts(ast, artifact) {
    const methodKeys = new Map();
    const resultCounts = new Map();
    const packageTypes = new Map();
    const typeUnderlyings = new Map();
    const imports = new Map();
    const interfaceTypes = new Set(["any", "interface{}", "error"]);
    const functionParamTypes = new Map();
    const functionResultTypes = new Map();
    const functionTypeParameters = new Map();
    const methodPointerReceivers = new Map();
    for (const item of ast.imports) {
        if (item.alias === "_" || item.alias === ".")
            continue;
        imports.set(item.alias ?? defaultImportName(item.path), item.path);
    }
    for (const exported of artifact.exports) {
        if (exported.typeText)
            packageTypes.set(exported.name, exported.typeText);
    }
    for (const statement of ast.body) {
        if (statement.kind === "TypeDecl") {
            for (const declaration of statement.declarations) {
                typeUnderlyings.set(declaration.name, declaration.type.text);
                if (isInterfaceTypeSpec(declaration))
                    interfaceTypes.add(declaration.name);
            }
            continue;
        }
        if (statement.kind === "ConstDecl" || statement.kind === "VarDecl") {
            for (const declaration of statement.declarations) {
                const typeText = declarationTypeText(declaration);
                if (typeText)
                    packageTypes.set(declaration.name, typeText);
            }
        }
    }
    for (const fn of ast.functions) {
        const key = functionPackageKey(fn);
        resultCounts.set(key, fn.signature.results.length);
        functionParamTypes.set(key, fn.signature.parameters.map((parameter) => parameter.type.text));
        functionResultTypes.set(key, fn.signature.results.map((result) => result.type.text));
        if (fn.typeParameters && fn.typeParameters.length > 0)
            functionTypeParameters.set(key, fn.typeParameters);
        if (!fn.receiver) {
            resultCounts.set(fn.name, fn.signature.results.length);
            functionParamTypes.set(fn.name, fn.signature.parameters.map((parameter) => parameter.type.text));
            functionResultTypes.set(fn.name, fn.signature.results.map((result) => result.type.text));
            if (fn.typeParameters && fn.typeParameters.length > 0)
                functionTypeParameters.set(fn.name, fn.typeParameters);
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
        imports,
        interfaceTypes,
        functionParamTypes,
        functionResultTypes,
        functionTypeParameters,
        methodPointerReceivers
    };
}
function isInterfaceTypeSpec(spec) {
    return Boolean(spec.interfaceMethods || spec.interfaceEmbeds || spec.type.text.trim().startsWith("interface{"));
}
function emitTypeDescriptorLines(ast, artifact, facts) {
    const descriptors = packageTypeDescriptors(ast, artifact, facts);
    if (descriptors.length === 0)
        return [];
    return descriptors.map((descriptor) => `__gojrTypeDescriptors[${JSON.stringify(descriptor.type)}] = ${JSON.stringify(descriptor)};`);
}
function packageTypeDescriptors(ast, artifact, facts) {
    const descriptors = new Map();
    for (const statement of ast.body) {
        if (statement.kind !== "TypeDecl")
            continue;
        for (const spec of statement.declarations) {
            const descriptor = typeSpecDescriptor(spec, artifact.importPath ?? "", artifact.packageName, facts);
            descriptors.set(descriptor.type, descriptor);
        }
    }
    for (const fn of ast.functions) {
        if (!fn.receiver)
            continue;
        const receiver = receiverBaseType(fn.receiver.type.text);
        const descriptor = descriptors.get(receiver);
        if (!descriptor)
            continue;
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
function typeSpecDescriptor(spec, pkgPath, pkgName, facts) {
    const type = spec.name;
    const kind = typeDescriptorKind(spec);
    const descriptor = {
        type,
        name: type,
        string: type,
        kind,
        pkgPath,
        pkgName: pkgName ?? defaultImportName(pkgPath),
        underlying: spec.type.text
    };
    Object.assign(descriptor, compositeTypeDescriptorFields(spec.type.text));
    if (spec.structFields) {
        descriptor.fields = spec.structFields.map((field) => {
            const fieldPkgPath = directSelectorTypePackagePath(field.type.text, facts);
            return {
                name: field.name,
                type: field.type.text,
                embedded: Boolean(field.embedded),
                ...(fieldPkgPath ? { pkgPath: fieldPkgPath } : {}),
                ...(field.tag ? { tag: field.tag } : {})
            };
        });
    }
    if (spec.typeParameters && spec.typeParameters.length > 0)
        descriptor.typeParameters = [...spec.typeParameters];
    if (spec.interfaceMethods) {
        descriptor.interfaceMethods = spec.interfaceMethods.map((method) => ({
            name: method.name,
            params: method.signature.parameters.map((parameter) => parameter.variadic ? `...${parameter.type.text}` : parameter.type.text),
            results: method.signature.results.map((result) => result.type.text)
        }));
    }
    if (spec.interfaceEmbeds)
        descriptor.interfaceEmbeds = spec.interfaceEmbeds.map((embed) => embed.text);
    return descriptor;
}
function typeDescriptorKind(spec) {
    if (spec.structFields)
        return "struct";
    if (isInterfaceTypeSpec(spec))
        return "interface";
    return typeTextKind(spec.type.text);
}
function typeTextKind(typeText) {
    const trimmed = typeText.trim();
    if (trimmed.startsWith("*"))
        return "ptr";
    if (trimmed.startsWith("[]"))
        return "slice";
    if (/^\[[^\]]+\]/.test(trimmed))
        return "array";
    if (trimmed.startsWith("map["))
        return "map";
    if (trimmed.startsWith("chan ") || trimmed.startsWith("<-chan") || trimmed.startsWith("chan<-"))
        return "chan";
    if (trimmed.startsWith("func("))
        return "func";
    if (trimmed.startsWith("interface{"))
        return "interface";
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
function compositeTypeDescriptorFields(typeText) {
    const trimmed = typeText.trim();
    const mapType = parseMapTypeText(trimmed);
    if (mapType)
        return { key: mapType.keyType, elem: mapType.valueType };
    const arrayType = parseArrayOrSliceTypeText(trimmed);
    if (arrayType) {
        const length = parseArrayLengthTypeText(trimmed);
        return {
            elem: arrayType.elementType,
            ...(length === undefined ? {} : { len: length })
        };
    }
    const chanElem = chanElementTypeText(trimmed);
    if (chanElem)
        return { elem: chanElem };
    const fn = parseFuncTypeText(trimmed);
    if (fn)
        return { params: fn.params, results: fn.results };
    return {};
}
function parseArrayLengthTypeText(typeText) {
    if (!typeText?.startsWith("[") || typeText.startsWith("[]"))
        return undefined;
    const close = typeText.indexOf("]");
    if (close < 0)
        return undefined;
    const raw = typeText.slice(1, close).trim();
    if (!/^\d+$/.test(raw))
        return undefined;
    const length = Number(raw);
    return Number.isSafeInteger(length) ? length : undefined;
}
function parseFuncTypeText(typeText) {
    const trimmed = typeText?.trim();
    if (!trimmed?.startsWith("func("))
        return undefined;
    const paramsEnd = matchingParenIndex(trimmed, 4);
    if (paramsEnd < 0)
        return undefined;
    const params = splitTopLevelTypes(trimmed.slice(5, paramsEnd));
    const resultText = trimmed.slice(paramsEnd + 1).trim();
    if (!resultText)
        return { params, results: [] };
    if (resultText.startsWith("(")) {
        const resultsEnd = matchingParenIndex(resultText, 0);
        if (resultsEnd !== resultText.length - 1)
            return undefined;
        return { params, results: splitTopLevelTypes(resultText.slice(1, resultsEnd)) };
    }
    return { params, results: [resultText] };
}
function matchingParenIndex(text, openIndex) {
    let depth = 0;
    for (let index = openIndex; index < text.length; index += 1) {
        const char = text[index];
        if (char === "(")
            depth += 1;
        if (char === ")") {
            depth -= 1;
            if (depth === 0)
                return index;
        }
    }
    return -1;
}
function splitTopLevelTypes(text) {
    const out = [];
    let start = 0;
    let parenDepth = 0;
    let bracketDepth = 0;
    for (let index = 0; index < text.length; index += 1) {
        const char = text[index];
        if (char === "(")
            parenDepth += 1;
        else if (char === ")")
            parenDepth -= 1;
        else if (char === "[")
            bracketDepth += 1;
        else if (char === "]")
            bracketDepth -= 1;
        else if (char === "," && parenDepth === 0 && bracketDepth === 0) {
            const item = text.slice(start, index).trim();
            if (item)
                out.push(item);
            start = index + 1;
        }
    }
    const last = text.slice(start).trim();
    if (last)
        out.push(last);
    return out;
}
function directSelectorTypePackagePath(typeText, facts) {
    let text = typeText.trim();
    while (text.startsWith("*"))
        text = text.slice(1).trim();
    const match = /^([A-Za-z_][A-Za-z0-9_]*)\.[A-Za-z_][A-Za-z0-9_]*(?:\[.*\])?$/.exec(text);
    if (!match)
        return undefined;
    return facts.imports.get(match[1] ?? "");
}
function defaultImportName(path) {
    const parts = path.split("/").filter(Boolean);
    return parts[parts.length - 1] ?? path;
}
function functionPackageKey(fn) {
    return fn.receiver ? `${receiverBaseType(fn.receiver.type.text)}.${fn.name}` : fn.name;
}
function receiverBaseType(typeText) {
    const trimmed = typeText.trim();
    const dereferenced = trimmed.startsWith("*") ? trimmed.slice(1).trim() : trimmed;
    return genericBaseTypeText(dereferenced);
}
function dereferencedTypeText(typeText) {
    const trimmed = typeText?.trim();
    return trimmed?.startsWith("*") ? trimmed.slice(1).trim() : undefined;
}
function genericBaseTypeText(typeText) {
    const trimmed = typeText.trim();
    if (trimmed.startsWith("[") || trimmed.startsWith("map["))
        return trimmed;
    const open = topLevelGenericOpenBracket(trimmed);
    if (open === undefined)
        return trimmed;
    const close = matchingTypeBracket(trimmed, open);
    return close === trimmed.length - 1 ? trimmed.slice(0, open).trim() : trimmed;
}
function topLevelGenericOpenBracket(typeText) {
    let depth = 0;
    for (let index = 0; index < typeText.length; index += 1) {
        const char = typeText[index];
        if (char === "[") {
            if (depth === 0)
                return index;
            depth += 1;
            continue;
        }
        if (char === "]")
            depth = Math.max(0, depth - 1);
    }
    return undefined;
}
function matchingTypeBracket(typeText, open) {
    let depth = 0;
    for (let index = open; index < typeText.length; index += 1) {
        const char = typeText[index];
        if (char === "[")
            depth += 1;
        if (char === "]") {
            depth -= 1;
            if (depth === 0)
                return index;
        }
    }
    return undefined;
}
function expressionTypeText(expression, env) {
    if (!expression)
        return undefined;
    if (expression.typeText)
        return expression.typeText;
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
    if (expression.kind === "Identifier")
        return env.localTypes.get(expression.name) ?? env.facts.packageTypes.get(expression.name);
    if (expression.kind === "IndexExpression")
        return mapValueTypeText(expressionTypeText(expression.object, env), env.facts);
    if (expression.kind === "MapLiteralExpression")
        return `map[${expression.keyType.text}]${expression.valueType.text}`;
    if (expression.kind === "ArrayLiteralExpression")
        return expression.type.text;
    if (expression.kind === "StructLiteralExpression")
        return expression.typeName;
    if (expression.kind === "FunctionLiteralExpression")
        return expression.typeText;
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
    if (expression.kind === "CallExpression" && expression.callee.kind === "TypeExpression")
        return expression.callee.type.text;
    if (expression.kind === "CallExpression" &&
        expression.callee.kind === "Identifier" &&
        primitiveTypeNames.has(expression.callee.name)) {
        return expression.callee.name;
    }
    return undefined;
}
function packageExportTypeText(ctx, expression) {
    if (expression?.kind !== "Identifier")
        return undefined;
    return ctx.options.artifact.exportIndex[expression.name]?.typeText ??
        ctx.options.artifact.exports.find((item) => item.name === expression.name)?.typeText;
}
function declarationTypeText(declaration, effectiveValue = declaration.value) {
    if (declaration.type?.text)
        return declaration.type.text;
    if (effectiveValue)
        return expressionTypeText(effectiveValue, emptyExpressionEnv);
    return undefined;
}
function tupleSourceTypeTexts(expression, targetCount, env) {
    if (targetCount === 2 && expression.kind === "IndexExpression") {
        return [mapValueTypeText(expressionTypeText(expression.object, env)), "bool"];
    }
    if (targetCount === 2 && expression.kind === "TypeAssertionExpression") {
        return [expression.type.text, "bool"];
    }
    if (targetCount === 2 && expression.kind === "UnaryExpression" && expression.operator === "<-") {
        return [chanElementTypeText(expressionTypeText(expression.operand, env)), "bool"];
    }
    if (targetCount === 1)
        return [expressionTypeText(expression, env)];
    return [];
}
function valueForTargetType(value, targetType, sourceType, env) {
    const resolvedTargetType = targetType ? resolveUnderlyingTypeText(targetType, env.facts) : undefined;
    if (resolvedTargetType && isIntegerType(resolvedTargetType) && !(resolvedTargetType === "uintptr" && isPointerLikeTypeText(sourceType))) {
        return `__gojrIntegerFrom(${value})`;
    }
    if (resolvedTargetType && isFloatType(resolvedTargetType))
        return `Number(${value})`;
    if (!isInterfaceTypeText(targetType, env.facts))
        return value;
    return `__gojrToInterface(${value}, ${JSON.stringify(targetType)}, ${JSON.stringify(sourceType)})`;
}
function isInterfaceTypeText(typeText, facts) {
    if (!typeText)
        return false;
    const normalized = typeText.trim();
    if (normalized === "any" || normalized === "interface{}" || normalized === "error")
        return true;
    if (normalized.startsWith("interface{"))
        return true;
    return facts.interfaceTypes.has(normalized) || facts.interfaceTypes.has(receiverBaseType(normalized));
}
function rangeIterationTypeTexts(expression, env) {
    const typeText = expressionTypeText(expression, env);
    if (!typeText)
        return {};
    const resolved = resolveUnderlyingTypeText(typeText, env.facts);
    if (resolved === "string")
        return { key: "int", value: "rune" };
    const mapType = parseMapTypeText(resolved);
    if (mapType)
        return { key: mapType.keyType, value: mapType.valueType };
    const sliceType = parseArrayOrSliceTypeText(resolved);
    if (sliceType)
        return { key: "int", value: sliceType.elementType };
    if (isIntegerType(resolved))
        return { key: resolved, value: resolved };
    return {};
}
function expressionToJs(ctx, expression, env = emptyExpressionEnv) {
    if (!expression)
        return undefined;
    const expressionEnv = expression.typeText && isFloatType(expression.typeText) && env.numericLiteralKind !== "float"
        ? { ...env, numericLiteralKind: "float" }
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
function identifierToJs(expression, env) {
    if (expression.name === "iota" && env.iotaValue !== undefined)
        return `${env.iotaValue}n`;
    const local = env.locals.get(expression.name);
    if (local)
        return local;
    const importPath = env.facts.imports.get(expression.name);
    if (importPath)
        return `__gojrImport(${JSON.stringify(importPath)})`;
    return `pkg[${JSON.stringify(expression.name)}]`;
}
function literalToJs(expression, env = emptyExpressionEnv) {
    switch (expression.literalKind) {
        case "int":
        case "rune":
            if (env.numericLiteralKind === "float") {
                if (typeof expression.value === "bigint")
                    return expression.value.toString();
                if (typeof expression.value === "number")
                    return String(Math.trunc(expression.value));
                return undefined;
            }
            if (typeof expression.value === "bigint")
                return `${expression.value.toString()}n`;
            if (typeof expression.value === "number")
                return `${Math.trunc(expression.value)}n`;
            return undefined;
        case "float":
            return typeof expression.value === "number" ? JSON.stringify(expression.value) : undefined;
        case "imag":
            return isComplexLiteralValue(expression.value) ? `__gojrComplex(${JSON.stringify(expression.value.real)}, ${JSON.stringify(expression.value.imag)})` : undefined;
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
    const arrayType = parseArrayOrSliceTypeText(expression.type.text);
    if (expression.elements.some((element) => element.key !== undefined)) {
        if (!arrayType) {
            ctx.emitError(`unsupported Stage 3 keyed array literal ${expression.type.text}`);
            return undefined;
        }
        const entries = [];
        for (const element of expression.elements) {
            const key = element.key ? expressionToJs(ctx, element.key, env) : "null";
            const value = expressionToJs(ctx, element.value, env);
            if (!key || !value)
                return undefined;
            entries.push(`[${key}, ${value}]`);
        }
        return `__gojrArrayLiteral(${JSON.stringify(expression.type.text)}, ${JSON.stringify(arrayType.elementType)}, [${entries.join(", ")}])`;
    }
    if (isByteSliceType(expression.type.text)) {
        const bytes = bytesFromArrayLiteral(expression);
        if (bytes) {
            if (bytes.length > 64)
                return `__gojrBytesBase64(${JSON.stringify(bytesToBase64(bytes))})`;
            return `new Uint8Array([${bytes.join(", ")}])`;
        }
        const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
        if (values.some((value) => value === undefined))
            return undefined;
        return `Uint8Array.from([${values.filter((value) => value !== undefined).join(", ")}], (item) => Number(item) & 255)`;
    }
    const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
    if (values.some((value) => value === undefined))
        return undefined;
    const rendered = values.filter((value) => value !== undefined);
    return `[${rendered.join(", ")}]`;
}
function structLiteralToJs(ctx, expression, env) {
    const underlying = resolveUnderlyingTypeText(expression.typeName, env.facts);
    const arrayType = parseArrayOrSliceTypeText(underlying);
    if (arrayType) {
        const entries = [];
        for (const field of expression.fields) {
            if (field.name) {
                ctx.emitError(`unsupported Stage 3 named array or slice literal field ${field.name}`);
                return undefined;
            }
            const key = field.key ? expressionToJs(ctx, field.key, env) : "null";
            const value = expressionToJs(ctx, field.value, env);
            if (!key || !value)
                return undefined;
            entries.push(`[${key}, ${value}]`);
        }
        return `__gojrArrayLiteral(${JSON.stringify(underlying)}, ${JSON.stringify(arrayType.elementType)}, [${entries.join(", ")}])`;
    }
    if (expression.fields.some((field) => !field.name)) {
        const values = [];
        for (const field of expression.fields) {
            if (field.name) {
                ctx.emitError(`unsupported Stage 3 mixed keyed and unkeyed struct literal ${expression.typeName}`);
                return undefined;
            }
            const value = expressionToJs(ctx, field.value, env);
            if (!value)
                return undefined;
            values.push(value);
        }
        return `__gojrStructFromValues(${JSON.stringify(expression.typeName)}, [${values.join(", ")}])`;
    }
    const fields = [];
    for (const field of expression.fields) {
        const value = expressionToJs(ctx, field.value, env);
        if (!value)
            return undefined;
        fields.push(`${JSON.stringify(field.name)}: ${value}`);
    }
    const pkgPath = packagePathForTypeText(expression.typeName, env);
    return `__gojrStruct(${JSON.stringify(expression.typeName)}, { ${fields.join(", ")} }${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
}
function mapLiteralToJs(ctx, expression, env) {
    const mapStringBytes = mapStringBytesLiteralEntries(expression);
    if (mapStringBytes) {
        return `__gojrMapStringBytesBase64([${mapStringBytes.map(([key, bytes]) => `[${JSON.stringify(key)}, ${JSON.stringify(bytesToBase64(bytes))}]`).join(", ")}])`;
    }
    const entries = [];
    for (const entry of expression.entries) {
        const key = expressionToJs(ctx, entry.key, env);
        const value = expressionToJs(ctx, entry.value, env);
        if (!key || !value)
            return undefined;
        entries.push(`[${key}, ${value}]`);
    }
    return `__gojrMap(new Map([${entries.join(", ")}]), ${zeroValueForType(expression.valueType.text, env.facts) ?? "null"})`;
}
function functionLiteralToJs(ctx, expression, env) {
    const literalEnv = cloneExpressionEnv(env);
    literalEnv.expectedReturnTypes = expression.signature.results.map((result) => result.type.text);
    const params = [];
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
    if (hasDefer)
        literalEnv.deferName = ctx.symbol("defer");
    const namedResults = [];
    const resultLines = [];
    for (const [index, result] of expression.signature.results.entries()) {
        if (!result.name || result.name === "_")
            continue;
        const name = safeLocalName(result.name, index);
        literalEnv.locals.set(result.name, name);
        literalEnv.localTypes.set(result.name, result.type.text);
        namedResults.push(name);
        resultLines.push(`  let ${name} = ${zeroValueForTypeInEnv(result.type.text, literalEnv) ?? "null"};`);
    }
    if (namedResults.length > 0)
        literalEnv.namedResults = namedResults;
    const statementLines = emitStatements(ctx, expression.body.statements, literalEnv, "  ");
    if (!statementLines)
        return undefined;
    const defaultReturn = functionLiteralDefaultReturn(expression, literalEnv);
    const body = [
        ...resultLines,
        ...statementLines,
        ...(functionBodyAlwaysReturns(expression.body) ? [] : [`  return ${defaultReturn};`])
    ].join("\n");
    if (hasDefer) {
        return `(async (${params.join(", ")}) => __gojrDeferScope(async (${literalEnv.deferName}) => {\n${body}\n}))`;
    }
    return `(async (${params.join(", ")}) => {\n${body}\n})`;
}
function functionLiteralDefaultReturn(expression, env) {
    const named = namedResultReturn(env);
    if (named)
        return named;
    if (expression.signature.results.length === 0)
        return "null";
    if (expression.signature.results.length === 1)
        return zeroValueForTypeInEnv(expression.signature.results[0]?.type.text, env) ?? "null";
    return `__gojrTuple([${expression.signature.results.map((result) => zeroValueForTypeInEnv(result.type.text, env) ?? "null").join(", ")}])`;
}
function unaryExpressionToJs(ctx, expression, env) {
    if (expression.operator === "&")
        return addressOfExpressionToJs(ctx, expression.operand, env);
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
        case "*":
            return `__gojrDeref(${operand})`;
        case "<-":
            return `(await __gojrChanRecv(${operand}))[0]`;
        default:
            ctx.emitError(`unsupported Stage 3 unary operator ${expression.operator}`);
            return undefined;
    }
}
function addressOfExpressionToJs(ctx, expression, env) {
    const typeText = expressionTypeText(expression, env) ?? "any";
    switch (expression.kind) {
        case "Identifier": {
            const target = identifierToJs(expression, env);
            const targetType = env.localTypes.get(expression.name) ?? env.facts.packageTypes.get(expression.name);
            const pkgPath = packagePathForTypeText(typeText, env);
            return `__gojrPointer(${JSON.stringify(typeText)}, () => ${target}, (next) => { ${target} = ${valueForTargetType("next", targetType, undefined, env)}; }${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
        }
        case "SelectorExpression": {
            const object = expressionToJs(ctx, expression.object, env);
            if (!object)
                return undefined;
            const pkgPath = packagePathForTypeText(typeText, env);
            return `__gojrAddressField(${object}, ${JSON.stringify(expression.field)}, ${JSON.stringify(typeText)}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
        }
        case "IndexExpression": {
            const objectType = expressionTypeText(expression.object, env) ?? "";
            if (parseMapTypeText(resolveUnderlyingTypeText(objectType, env.facts))) {
                ctx.emitError("unsupported Stage 4 address-of map index");
                return undefined;
            }
            const object = expressionToJs(ctx, expression.object, env);
            const index = expressionToJs(ctx, expression.index, env);
            if (!object || !index)
                return undefined;
            const pkgPath = packagePathForTypeText(typeText, env);
            return `__gojrAddressIndex(${object}, ${index}, ${JSON.stringify(typeText)}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
        }
        case "StructLiteralExpression":
        case "ArrayLiteralExpression":
        case "MapLiteralExpression": {
            const value = expressionToJs(ctx, expression, env);
            if (!value)
                return undefined;
            const pkgPath = packagePathForTypeText(typeText, env);
            return `__gojrPointerValue(${JSON.stringify(typeText)}, ${value}${pkgPath ? `, ${JSON.stringify(pkgPath)}` : ""})`;
        }
        default:
            ctx.emitError(`unsupported Stage 4 address-of ${expression.kind}`);
            return undefined;
    }
}
function binaryExpressionToJs(ctx, expression, env) {
    const left = expressionToJs(ctx, expression.left, env);
    const right = expressionToJs(ctx, expression.right, env);
    if (!left || !right)
        return undefined;
    if (isComplexType(expression.typeText) || isComplexType(expression.left.typeText) || isComplexType(expression.right.typeText)) {
        return complexBinaryExpressionToJs(ctx, expression, left, right);
    }
    const usesFloatArithmetic = expressionPrefersNumberArithmetic(expression, env);
    if (usesFloatArithmetic && numberArithmeticOperators.has(expression.operator))
        return `((Number(${left})) ${expression.operator} (Number(${right})))`;
    if (usesFloatArithmetic && numberComparisonOperators.has(expression.operator))
        return `((Number(${left})) ${expression.operator} (Number(${right})))`;
    if (expression.operator === "&&" || expression.operator === "||")
        return `((${left}) ${expression.operator} (${right}))`;
    if (expression.operator === "==")
        return `__gojrEqual(${left}, ${right})`;
    if (expression.operator === "!=")
        return `(!__gojrEqual(${left}, ${right}))`;
    if (expression.operator === "&^")
        return `((${left}) & ~(${right}))`;
    if (expression.operator === "^")
        return `__gojrBitwiseXor(${left}, ${right})`;
    const operator = jsBinaryOperator(expression.operator);
    if (!operator) {
        ctx.emitError(`unsupported Stage 3 binary operator ${expression.operator}`);
        return undefined;
    }
    return `((${left}) ${operator} (${right}))`;
}
function complexBinaryExpressionToJs(ctx, expression, left, right) {
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
function selectorExpressionToJs(ctx, expression, env) {
    if (isImportedPackageSelector(expression, env) && expression.object.kind === "Identifier") {
        const importPath = env.facts.imports.get(expression.object.name);
        return `__gojrPackageSelector(${JSON.stringify(importPath)}, ${JSON.stringify(expression.field)})`;
    }
    const methodKey = methodPackageKeyForSelector(expression, env);
    if (methodKey) {
        const receiver = methodReceiverToJs(ctx, expression, methodKey, env);
        if (!receiver)
            return undefined;
        const typeArgs = methodReceiverTypeArgObject(expression, methodKey, env, receiver);
        return `(async (...__gojrMethodArgs) => await pkg[${JSON.stringify(methodKey)}](${[...(typeArgs ? [typeArgs] : []), receiver, "...__gojrMethodArgs"].join(", ")}))`;
    }
    const object = expressionToJs(ctx, expression.object, env);
    if (!object)
        return undefined;
    return `__gojrGetField(${object}, ${JSON.stringify(expression.field)})`;
}
function methodPackageKeyForSelector(expression, env) {
    const receiverType = expressionTypeText(expression.object, env);
    if (!receiverType)
        return undefined;
    return env.facts.methodKeys.get(`${receiverBaseType(receiverType)}.${expression.field}`);
}
function methodReceiverToJs(ctx, expression, methodKey, env) {
    const receiverType = expressionTypeText(expression.object, env);
    if (env.facts.methodPointerReceivers.get(methodKey) && !receiverType?.trim().startsWith("*")) {
        const addressed = addressOfExpressionToJs(ctx, expression.object, env);
        if (addressed)
            return addressed;
    }
    return expressionToJs(ctx, expression.object, env);
}
function methodReceiverTypeArgObject(expression, methodKey, env, receiverJs) {
    const parameterNames = env.facts.functionTypeParameters.get(methodKey);
    if (!parameterNames || parameterNames.length === 0)
        return undefined;
    const receiverType = expressionTypeText(expression.object, env);
    const typeArgs = receiverType ? genericTypeArgumentsFromTypeText(receiverType) : undefined;
    if (typeArgs && typeArgs.length > 0) {
        const fields = parameterNames.map((name, index) => `${JSON.stringify(name)}: ${JSON.stringify(typeArgs[index] ?? "any")}`);
        return `({ ${fields.join(", ")} })`;
    }
    return `__gojrTypeArgsFromReceiver(${receiverJs}, ${JSON.stringify(parameterNames)})`;
}
function genericTypeArgumentsFromTypeText(typeText) {
    let trimmed = typeText.trim();
    if (trimmed.startsWith("*"))
        trimmed = trimmed.slice(1).trim();
    const open = topLevelGenericOpenBracket(trimmed);
    if (open === undefined)
        return undefined;
    const close = matchingTypeBracket(trimmed, open);
    if (close !== trimmed.length - 1 || close === undefined)
        return undefined;
    return splitTopLevelTypes(trimmed.slice(open + 1, close));
}
function packagePathForTypeText(typeText, env) {
    let trimmed = typeText?.trim();
    if (!trimmed)
        return undefined;
    while (trimmed.startsWith("*"))
        trimmed = trimmed.slice(1).trim();
    const base = genericBaseTypeText(trimmed);
    const dot = base.indexOf(".");
    if (dot <= 0)
        return undefined;
    return env.facts.imports.get(base.slice(0, dot));
}
function methodCallToJs(ctx, expression, env, methodKey) {
    const callee = expression.callee;
    const receiver = methodReceiverToJs(ctx, callee, methodKey, env);
    if (!receiver)
        return undefined;
    const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(methodKey), expression.spreadLast);
    if (!args)
        return undefined;
    const typeArgs = methodReceiverTypeArgObject(callee, methodKey, env, receiver);
    return `(await pkg[${JSON.stringify(methodKey)}](${[...(typeArgs ? [typeArgs] : []), receiver, ...args].join(", ")}))`;
}
function callExpressionToJs(ctx, expression, env) {
    if (expression.callee.kind === "Identifier" && expression.callee.name === "new" && !env.locals.has("new")) {
        return newCallToJs(ctx, expression, env);
    }
    if (expression.callee.kind === "Identifier" && expression.callee.name === "make" && !env.locals.has("make")) {
        return makeCallToJs(ctx, expression, env);
    }
    const conversionType = conversionCalleeTypeText(expression.callee, env);
    if (conversionType) {
        return conversionCallToJs(ctx, conversionType, expression.args, env);
    }
    const builtin = builtinCallToJs(ctx, expression, env);
    if (builtin)
        return builtin;
    const instantiatedSelectorCall = instantiatedImportedSelectorCallToJs(ctx, expression, env);
    if (instantiatedSelectorCall)
        return instantiatedSelectorCall;
    if (expression.callee.kind === "SelectorExpression") {
        const methodKey = methodPackageKeyForSelector(expression.callee, env);
        if (methodKey)
            return methodCallToJs(ctx, expression, env, methodKey);
        if (isImportedPackageSelector(expression.callee, env) && expression.callee.object.kind === "Identifier") {
            const importPath = env.facts.imports.get(expression.callee.object.name);
            const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
            if (!renderedArgs)
                return undefined;
            return `(await (__gojrPackageCallableSelector(${JSON.stringify(importPath)}, ${JSON.stringify(expression.callee.field)}))(${renderedArgs.join(", ")}))`;
        }
        if (!isImportedPackageSelector(expression.callee, env))
            return dynamicMethodCallToJs(ctx, expression, env);
    }
    const instantiatedName = instantiatedFunctionName(expression.callee);
    if (instantiatedName && !env.locals.has(instantiatedName)) {
        const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(instantiatedName), expression.spreadLast);
        if (!args)
            return undefined;
        const typeArgs = instantiatedFunctionTypeArgObject(expression.callee, instantiatedName, env);
        return `(await pkg[${JSON.stringify(instantiatedName)}](${[...(typeArgs ? [typeArgs] : []), ...args].join(", ")}))`;
    }
    if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
        const renderedArgs = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(expression.callee.name), expression.spreadLast);
        if (!renderedArgs)
            return undefined;
        const typeArgs = inferredFunctionTypeArgObject(expression.callee.name, expression.args, env);
        return `(await pkg[${JSON.stringify(expression.callee.name)}](${[...(typeArgs ? [typeArgs] : []), ...renderedArgs].join(", ")}))`;
    }
    const callee = expressionToJs(ctx, expression.callee, env);
    if (!callee)
        return undefined;
    const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
    if (!renderedArgs)
        return undefined;
    return `(await (${callee})(${renderedArgs.join(", ")}))`;
}
function instantiatedImportedSelectorCallToJs(ctx, expression, env) {
    if (expression.callee.kind !== "IndexExpression")
        return undefined;
    const selector = expression.callee.object;
    if (selector.kind !== "SelectorExpression" || selector.object.kind !== "Identifier")
        return undefined;
    const importPath = env.facts.imports.get(selector.object.name);
    if (!importPath)
        return undefined;
    const typeArgText = typeArgumentExpressionText(expression.callee.index);
    if (!typeArgText)
        return undefined;
    const args = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
    if (!args)
        return undefined;
    const imported = `__gojrImport(${JSON.stringify(importPath)})`;
    const typeArgs = `__gojrImportedTypeArgs(${imported}, ${JSON.stringify(selector.field)}, [${splitTopLevelTypes(typeArgText).map((typeArg) => JSON.stringify(typeArg)).join(", ")}])`;
    return `(await (__gojrGetField(${imported}, ${JSON.stringify(selector.field)}))(${[typeArgs, ...args].join(", ")}))`;
}
function conversionCalleeTypeText(callee, env) {
    if (callee.kind === "IndexExpression" &&
        callee.object.kind === "SelectorExpression" &&
        isImportedPackageSelector(callee.object, env)) {
        return undefined;
    }
    if (callee.kind === "TypeExpression")
        return callee.type.text;
    if (callee.kind === "Identifier") {
        if (env.locals.has(callee.name))
            return undefined;
        if (primitiveTypeNames.has(callee.name) || env.facts.typeUnderlyings.has(callee.name))
            return callee.name;
        return undefined;
    }
    const typeText = conversionTypeExpressionText(callee, env);
    if (!typeText)
        return undefined;
    const base = genericBaseTypeText(typeText);
    if (isCompositeTypeText(typeText) || env.facts.typeUnderlyings.has(base) || typeText === "unsafe.Pointer")
        return typeText;
    return undefined;
}
function conversionTypeExpressionText(expression, env) {
    switch (expression.kind) {
        case "TypeExpression":
            return expression.type.text;
        case "Identifier":
            if (env.locals.has(expression.name))
                return undefined;
            return primitiveTypeNames.has(expression.name) || env.facts.typeUnderlyings.has(expression.name)
                ? expression.name
                : undefined;
        case "SelectorExpression":
            if (expression.object.kind !== "Identifier" || !env.facts.imports.has(expression.object.name))
                return undefined;
            return `${expression.object.name}.${expression.field}`;
        case "IndexExpression": {
            const object = conversionTypeExpressionText(expression.object, env);
            const index = typeArgumentExpressionText(expression.index);
            return object && index ? `${object}[${index}]` : undefined;
        }
        case "UnaryExpression": {
            if (expression.operator !== "*")
                return undefined;
            const operand = conversionTypeExpressionText(expression.operand, env);
            return operand ? `*${operand}` : undefined;
        }
        default:
            return undefined;
    }
}
function renderCallArgs(ctx, args, env, targetTypes = [], spreadLast = false) {
    const rendered = [];
    for (const [index, arg] of args.entries()) {
        const value = expressionToJs(ctx, arg, env);
        if (!value)
            return undefined;
        const lowered = valueForTargetType(value, targetTypes[index], expressionTypeText(arg, env), env);
        rendered.push(spreadLast && index === args.length - 1 ? `...__gojrSpread(${lowered})` : lowered);
    }
    return rendered;
}
function instantiatedFunctionName(expression) {
    if (expression.kind !== "IndexExpression")
        return undefined;
    if (expression.object.kind === "Identifier")
        return expression.object.name;
    return undefined;
}
function instantiatedFunctionTypeArgObject(expression, functionName, env) {
    if (expression.kind !== "IndexExpression")
        return undefined;
    const parameterNames = env.facts.functionTypeParameters.get(functionName);
    if (!parameterNames || parameterNames.length === 0)
        return undefined;
    const typeArgText = typeArgumentExpressionText(expression.index);
    if (!typeArgText)
        return undefined;
    const typeArgs = splitTopLevelTypes(typeArgText);
    const fields = parameterNames.map((name, index) => `${JSON.stringify(name)}: ${JSON.stringify(typeArgs[index] ?? "any")}`);
    return `({ ${fields.join(", ")} })`;
}
function inferredFunctionTypeArgObject(functionName, args, env) {
    const parameterNames = env.facts.functionTypeParameters.get(functionName);
    if (!parameterNames || parameterNames.length === 0)
        return undefined;
    const parameterSet = new Set(parameterNames);
    const parameterTypes = env.facts.functionParamTypes.get(functionName) ?? [];
    const inferred = new Map();
    for (const [index, parameterType] of parameterTypes.entries()) {
        const actualType = expressionTypeText(args[index], env);
        if (!actualType)
            continue;
        inferGenericTypeArguments(parameterType, actualType, parameterSet, inferred);
    }
    const fields = parameterNames.map((name) => `${JSON.stringify(name)}: ${JSON.stringify(inferred.get(name) ?? "any")}`);
    return `({ ${fields.join(", ")} })`;
}
function inferGenericTypeArguments(pattern, actual, parameters, out) {
    const lhs = pattern?.trim();
    const rhs = actual?.trim();
    if (!lhs || !rhs)
        return;
    if (parameters.has(lhs)) {
        if (!out.has(lhs) || out.get(lhs) === "any")
            out.set(lhs, rhs);
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
        for (const [index, lhsArg] of lhsArgs.entries())
            inferGenericTypeArguments(lhsArg, rhsArgs[index], parameters, out);
    }
}
function typeArgumentExpressionText(expression) {
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
            if (expression.operator !== "*")
                return undefined;
            const operand = typeArgumentExpressionText(expression.operand);
            return operand ? `*${operand}` : undefined;
        }
        default:
            return undefined;
    }
}
function isImportedPackageSelector(expression, env) {
    return expression.object.kind === "Identifier" && env.facts.imports.has(expression.object.name);
}
function dynamicMethodCallToJs(ctx, expression, env) {
    if (expression.callee.kind !== "SelectorExpression")
        return undefined;
    const receiver = expressionToJs(ctx, expression.callee.object, env);
    if (!receiver)
        return undefined;
    const args = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
    if (!args)
        return undefined;
    return `(await __gojrCallMethod(${receiver}, ${JSON.stringify(expression.callee.field)}, [${args.join(", ")}], pkg))`;
}
function newCallToJs(ctx, expression, env) {
    const typeArg = expression.args[0];
    const typeText = typeArg ? typeArgumentExpressionText(typeArg) : undefined;
    if (!typeText) {
        ctx.emitError("unsupported Stage 4 new without type argument");
        return undefined;
    }
    if (expression.args.length !== 1) {
        ctx.emitError("unsupported Stage 4 new arity");
        return undefined;
    }
    return `__gojrPointerValue(${JSON.stringify(typeText)}, ${zeroValueForTypeInEnv(typeText, env) ?? `__gojrZero(${JSON.stringify(typeText)})`})`;
}
function makeCallToJs(ctx, expression, env) {
    const typeArg = expression.args[0];
    const typeText = typeArg ? typeArgumentExpressionText(typeArg) : undefined;
    if (!typeText) {
        ctx.emitError("unsupported Stage 4 make without type argument");
        return undefined;
    }
    if (env.typeParameters?.has(typeText)) {
        const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
        const capacity = expression.args[2] ? expressionToJs(ctx, expression.args[2], env) : length;
        if (!length || !capacity)
            return undefined;
        return `__gojrMake(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(typeText)}), Number(${length}), Number(${capacity}))`;
    }
    const resolvedTypeText = resolveUnderlyingTypeText(typeText, env.facts);
    if (chanElementTypeText(resolvedTypeText)) {
        const capacity = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
        if (!capacity)
            return undefined;
        return `__gojrMakeChan(${JSON.stringify(chanElementTypeText(resolvedTypeText) ?? "any")}, Number(${capacity}))`;
    }
    if (parseMapTypeText(resolvedTypeText)) {
        return `__gojrMap(new Map(), ${zeroValueForTypeInEnv(mapValueTypeText(typeText), env) ?? "null"})`;
    }
    const slice = parseArrayOrSliceTypeText(resolvedTypeText);
    if (slice && resolvedTypeText.startsWith("[]")) {
        const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
        if (!length)
            return undefined;
        return `Array.from({ length: Number(${length}) }, () => ${zeroValueForTypeInEnv(slice.elementType, env) ?? "null"})`;
    }
    ctx.emitError(`unsupported Stage 4 make(${typeText})`);
    return undefined;
}
function builtinCallToJs(ctx, expression, env) {
    if (expression.callee.kind !== "Identifier" || env.locals.has(expression.callee.name))
        return undefined;
    const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
    if (!renderedArgs)
        return undefined;
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
            return `__gojrDelete(${renderedArgs[0]}, ${renderedArgs[1]})`;
        case "panic":
            if (renderedArgs.length !== 1) {
                ctx.emitError("unsupported Stage 3 panic call arity");
                return undefined;
            }
            return `__gojrPanic(${renderedArgs[0]})`;
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
function conversionCallToJs(ctx, typeText, args, env) {
    if (args.length !== 1) {
        ctx.emitError(`unsupported Stage 3 conversion to ${typeText} with ${args.length} arguments`);
        return undefined;
    }
    const value = expressionToJs(ctx, args[0], env);
    if (!value)
        return undefined;
    const sourceType = expressionTypeText(args[0], env);
    const resolvedType = resolveUnderlyingTypeText(typeText, env.facts);
    if (isByteSliceType(resolvedType))
        return `__gojrBytesFrom(${value})`;
    if (typeText.trim() === "unsafe.Pointer" || resolvedType.trim() === "unsafe.Pointer")
        return `__gojrConvertPointer(${JSON.stringify(typeText)}, ${value})`;
    if (resolvedType === "uintptr" && isPointerLikeTypeText(sourceType))
        return `__gojrPointerToUintptr(${value})`;
    if (isIntegerType(resolvedType))
        return `BigInt(${value})`;
    if (isFloatType(resolvedType))
        return `Number(${value})`;
    if (isComplexType(resolvedType))
        return `__gojrToComplex(${value})`;
    if (resolvedType === "bool")
        return `Boolean(${value})`;
    if (resolvedType === "string")
        return `__gojrStringFrom(${value})`;
    {
        const arrayTarget = parseArrayOrSliceTypeText(resolvedType);
        if (arrayTarget && !resolvedType.startsWith("[]"))
            return `__gojrConvertArray(${JSON.stringify(typeText)}, ${JSON.stringify(arrayTarget.elementType)}, ${value})`;
    }
    if (resolvedType.startsWith("[]"))
        return `__gojrConvertSlice(${JSON.stringify(typeText)}, ${value})`;
    if (resolvedType.startsWith("map["))
        return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
    if (resolvedType.startsWith("chan ") || resolvedType.startsWith("<-chan") || resolvedType.startsWith("chan<-"))
        return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
    if (resolvedType.startsWith("func("))
        return `__gojrConvertNilable(${JSON.stringify(typeText)}, ${value})`;
    if (typeText.trim().startsWith("*") || resolvedType.trim().startsWith("*"))
        return `__gojrConvertPointer(${JSON.stringify(typeText)}, ${value})`;
    if (env.facts.typeUnderlyings.has(typeText))
        return value;
    ctx.emitError(`unsupported Stage 3 conversion to ${typeText}`);
    return undefined;
}
function indexExpressionToJs(ctx, expression, env) {
    const instantiated = instantiatedImportedSelectorValueToJs(ctx, expression, env);
    if (instantiated)
        return instantiated;
    const object = expressionToJs(ctx, expression.object, env);
    const index = expressionToJs(ctx, expression.index, env);
    if (!object || !index)
        return undefined;
    const objectType = expressionTypeText(expression.object, env) ?? packageExportTypeText(ctx, expression.object) ?? "";
    const resolvedObjectType = resolveUnderlyingTypeText(objectType, env.facts);
    if (parseMapTypeText(resolvedObjectType)) {
        const zero = zeroValueForTypeInEnv(mapValueTypeText(objectType, env.facts) ?? expression.typeText, env) ?? "null";
        return `__gojrMapGetOk(${object}, ${index}, ${zero})[0]`;
    }
    if (resolvedObjectType === "string")
        return `BigInt(__gojrStringByteAt(${object}, Number(${index})))`;
    const elementType = expression.typeText ?? parseArrayOrSliceTypeText(resolvedObjectType)?.elementType;
    return `__gojrIndex(${object}, Number(${index}), ${isIntegerType(elementType ?? "") ? "true" : "false"})`;
}
function instantiatedImportedSelectorValueToJs(_ctx, expression, env) {
    const selector = expression.object;
    if (selector.kind !== "SelectorExpression" || selector.object.kind !== "Identifier")
        return undefined;
    const importPath = env.facts.imports.get(selector.object.name);
    if (!importPath)
        return undefined;
    const typeArgText = typeArgumentExpressionText(expression.index);
    if (!typeArgText)
        return undefined;
    const imported = `__gojrImport(${JSON.stringify(importPath)})`;
    const typeArgs = `__gojrImportedTypeArgs(${imported}, ${JSON.stringify(selector.field)}, [${splitTopLevelTypes(typeArgText).map((typeArg) => JSON.stringify(typeArg)).join(", ")}])`;
    return `(async (...__gojrGenericArgs) => await (__gojrGetField(${imported}, ${JSON.stringify(selector.field)}))(${typeArgs}, ...__gojrGenericArgs))`;
}
function sliceExpressionToJs(ctx, expression, env) {
    const object = expressionToJs(ctx, expression.object, env);
    const start = expression.start ? expressionToJs(ctx, expression.start, env) : undefined;
    const end = expression.end ? expressionToJs(ctx, expression.end, env) : undefined;
    const max = expression.max ? expressionToJs(ctx, expression.max, env) : undefined;
    if (!object || (expression.start && !start) || (expression.end && !end) || (expression.max && !max))
        return undefined;
    return `__gojrSlice(${object}, ${start ? `Number(${start})` : "null"}, ${end ? `Number(${end})` : "null"}, ${max ? `Number(${max})` : "null"})`;
}
function typeAssertionExpressionToJs(ctx, expression, env = emptyExpressionEnv) {
    if (expression.kind !== "TypeAssertionExpression") {
        ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
        return undefined;
    }
    const value = expressionToJs(ctx, expression.expression, env);
    if (!value)
        return undefined;
    return `__gojrTypeAssert(${value}, ${JSON.stringify(expression.type.text)})`;
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
function bytesFromArrayLiteral(expression) {
    const bytes = expression.elements.map((element) => byteLiteralValue(element.value));
    return bytes.some((value) => value === undefined) ? undefined : bytes;
}
function mapStringBytesLiteralEntries(expression) {
    if (expression.keyType.text !== "string" || !isByteSliceType(expression.valueType.text))
        return undefined;
    const entries = [];
    for (const entry of expression.entries) {
        if (entry.key.kind !== "Literal" || entry.key.literalKind !== "string" || typeof entry.key.value !== "string")
            return undefined;
        if (entry.value.kind !== "ArrayLiteralExpression" || !isByteSliceType(entry.value.type.text))
            return undefined;
        const bytes = bytesFromArrayLiteral(entry.value);
        if (!bytes)
            return undefined;
        entries.push([entry.key.value, bytes]);
    }
    return entries;
}
function bytesToBase64(bytes) {
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
function zeroValueForType(typeText, facts = emptyPackageEmitFacts) {
    const staticType = typeText?.trim();
    if (!staticType)
        return "null";
    if (isNilAssignableConcreteTypeText(staticType))
        return `__gojrTypedNil(${JSON.stringify(staticType)})`;
    const valueType = resolveUnderlyingTypeText(staticType, facts);
    if (isNilAssignableConcreteTypeText(valueType))
        return `__gojrTypedNil(${JSON.stringify(staticType)})`;
    if (facts.typeUnderlyings.has(staticType))
        return `__gojrZero(${JSON.stringify(staticType)})`;
    const genericBase = genericBaseTypeText(staticType);
    if (genericBase !== staticType && facts.typeUnderlyings.has(genericBase))
        return `__gojrZero(${JSON.stringify(staticType)})`;
    const arrayType = parseArrayOrSliceTypeText(valueType);
    if (arrayType && !valueType.startsWith("[]")) {
        const length = parseArrayLengthTypeText(valueType);
        if (length !== undefined)
            return `Array.from({ length: ${length} }, () => ${zeroValueForType(arrayType.elementType, facts) ?? "null"})`;
    }
    const anonymousStructFields = parseAnonymousStructFields(valueType);
    if (anonymousStructFields) {
        const fields = anonymousStructFields.map((field) => `${JSON.stringify(field.name)}: ${zeroValueForType(field.type, facts) ?? "null"}`);
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
function parseAnonymousStructFields(typeText) {
    const match = /^struct\s*\{([\s\S]*)\}$/.exec(typeText.trim());
    if (!match)
        return undefined;
    const body = match[1]?.trim() ?? "";
    if (body === "")
        return [];
    return splitTopLevel(body, ";").flatMap(parseAnonymousStructField);
}
function parseAnonymousStructField(text) {
    const field = stripStructFieldTag(text.trim());
    if (!field)
        return [];
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
function stripStructFieldTag(field) {
    let depth = 0;
    for (let index = 0; index < field.length; index += 1) {
        const char = field[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth += 1;
        else if (char === "]" || char === ")" || char === "}")
            depth -= 1;
        else if (char === "`" && depth === 0)
            return field.slice(0, index).trim();
    }
    return field;
}
function firstAnonymousStructFieldNameTypeSplit(field) {
    let depth = 0;
    for (let index = 0; index < field.length; index += 1) {
        const char = field[index] ?? "";
        if (char === "[" || char === "(" || char === "{")
            depth += 1;
        else if (char === "]" || char === ")" || char === "}")
            depth -= 1;
        if (depth !== 0 || !/\s/.test(char))
            continue;
        const namesText = field.slice(0, index).trim();
        if (!anonymousStructFieldNameList(namesText))
            continue;
        let next = index;
        while (next < field.length && /\s/.test(field[next] ?? ""))
            next += 1;
        if (next < field.length)
            return index;
    }
    return -1;
}
function anonymousStructFieldNameList(namesText) {
    const names = splitTopLevel(namesText, ",").map((name) => name.trim());
    return names.length > 0 && names.every((name) => /^[A-Za-z_]\w*$/.test(name));
}
function embeddedFieldNameFromTypeText(typeText) {
    let text = typeText.trim();
    while (text.startsWith("*"))
        text = text.slice(1).trim();
    const generic = text.indexOf("[");
    if (generic >= 0)
        text = text.slice(0, generic).trim();
    const dot = text.lastIndexOf(".");
    return dot >= 0 ? text.slice(dot + 1) : text;
}
function splitTopLevel(text, separator) {
    const out = [];
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
        if (inBacktick)
            continue;
        if (char === "(")
            parenDepth += 1;
        else if (char === ")")
            parenDepth -= 1;
        else if (char === "[")
            bracketDepth += 1;
        else if (char === "]")
            bracketDepth -= 1;
        else if (char === "{")
            braceDepth += 1;
        else if (char === "}")
            braceDepth -= 1;
        else if (char === separator && parenDepth === 0 && bracketDepth === 0 && braceDepth === 0) {
            const item = text.slice(start, index).trim();
            if (item)
                out.push(item);
            start = index + 1;
        }
    }
    const last = text.slice(start).trim();
    if (last)
        out.push(last);
    return out;
}
function zeroValueForTypeInEnv(typeText, env) {
    const staticType = typeText?.trim();
    if (staticType && env.typeParameters?.has(staticType)) {
        return `__gojrZero(__gojrTypeArg(__gojrTypeArgs, ${JSON.stringify(staticType)}))`;
    }
    return zeroValueForType(typeText, env.facts);
}
function resolveUnderlyingTypeText(typeText, facts) {
    let current = typeText.trim();
    const seen = new Set();
    while (!seen.has(current)) {
        seen.add(current);
        const underlying = facts.typeUnderlyings.get(current);
        if (!underlying)
            break;
        current = underlying.trim();
    }
    return current;
}
function isNilAssignableConcreteTypeText(typeText) {
    return typeText.startsWith("*") ||
        typeText.startsWith("[]") ||
        typeText.startsWith("map[") ||
        typeText.startsWith("chan ") ||
        typeText.startsWith("<-chan") ||
        typeText.startsWith("chan<-") ||
        typeText.startsWith("func(");
}
function isCompositeTypeText(typeText) {
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
function mapValueTypeText(typeText, facts = emptyPackageEmitFacts) {
    if (!typeText)
        return undefined;
    return parseMapTypeText(resolveUnderlyingTypeText(typeText, facts))?.valueType;
}
function parseMapTypeText(typeText) {
    if (!typeText?.startsWith("map["))
        return undefined;
    let depth = 0;
    for (let index = 4; index < typeText.length; index += 1) {
        const char = typeText[index];
        if (char === "[")
            depth += 1;
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
function parseArrayOrSliceTypeText(typeText) {
    if (!typeText?.startsWith("["))
        return undefined;
    const close = typeText.indexOf("]");
    if (close < 0)
        return undefined;
    return { elementType: typeText.slice(close + 1).trim() };
}
function chanElementTypeText(typeText) {
    if (!typeText)
        return undefined;
    const trimmed = typeText.trim();
    if (trimmed.startsWith("chan<-"))
        return trimmed.slice("chan<-".length).trim();
    if (trimmed.startsWith("<-chan"))
        return trimmed.slice("<-chan".length).trim();
    if (trimmed.startsWith("chan "))
        return trimmed.slice("chan ".length).trim();
    return undefined;
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
    "float64",
    "complex64",
    "complex128"
]);
function isIntegerType(typeText) {
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
function isFloatType(typeText) {
    return typeText === "float32" || typeText === "float64" || typeText === "untyped float";
}
function isComplexType(typeText) {
    return typeText === "complex64" || typeText === "complex128" || typeText === "untyped complex";
}
function isPointerLikeTypeText(typeText) {
    const text = typeText?.trim();
    return Boolean(text && (text === "unsafe.Pointer" || text.startsWith("*")));
}
const numberArithmeticOperators = new Set(["+", "-", "*", "/"]);
const numberComparisonOperators = new Set(["==", "!=", "<", "<=", ">", ">="]);
function expressionPrefersNumberArithmetic(expression, env) {
    if (!expression)
        return false;
    const typeText = expressionTypeText(expression, env);
    if (typeText && isFloatType(resolveUnderlyingTypeText(typeText, env.facts)))
        return true;
    switch (expression.kind) {
        case "Literal":
            return expression.literalKind === "float";
        case "UnaryExpression":
            return expressionPrefersNumberArithmetic(expression.operand, env);
        case "BinaryExpression":
            return expressionPrefersNumberArithmetic(expression.left, env) || expressionPrefersNumberArithmetic(expression.right, env);
        case "CallExpression":
            if (expression.callee.kind === "TypeExpression")
                return isFloatType(resolveUnderlyingTypeText(expression.callee.type.text, env.facts));
            return isFloatType(typeText ?? "");
        default:
            return false;
    }
}
function isComplexLiteralValue(value) {
    return Boolean(value && typeof value === "object" && "real" in value && "imag" in value);
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
    if (/^[A-Za-z_$][A-Za-z0-9_$]*$/.test(candidate) && !JS_RESERVED_WORDS.has(candidate))
        return candidate;
    return `__gojrArg${index}`;
}
function stage1JavaScript(artifact, usesWasm, bodyLines, typeDescriptorLines) {
    const artifactHeader = { ...artifact };
    delete artifactHeader.runtime;
    const generatedBuiltinImports = JSON.stringify(GENERATED_ARTIFACT_INTRINSIC_IMPORTS);
    return [
        "// Code generated by gojr Stage 1 copy-and-patch emitter; DO NOT EDIT.",
        `export const gojrPackageArtifact = ${JSON.stringify(artifactHeader, null, 2)};`,
        "const __gojrTypeDescriptors = Object.create(null);",
        ...typeDescriptorLines,
        "let __gojrActiveImportsByPath = {};",
        "let __gojrActiveRuntimeOptions = {};",
        "export async function instantiateGoJrPackage(runtime = {}, options = {}) {",
        "  const pkg = Object.create(null);",
        "  const importsByPath = options.importsByPath || runtime.importsByPath || options.packages || runtime.packages || {};",
        "  __gojrActiveImportsByPath = importsByPath;",
        "  const __gojrStdout = options.stdout || runtime.stdout || (() => {});",
        "  __gojrActiveRuntimeOptions = { ...runtime, ...options, stdout: __gojrStdout };",
        "  const __gojrImport = (path) => {",
        "    const builtin = __gojrBuiltinImport(path, importsByPath);",
        "    if (__gojrBuiltinImportOverrides(path) && builtin) return builtin;",
        "    const imported = importsByPath[path];",
        "    if (imported && typeof imported === \"object\" && \"package\" in imported) return imported.package;",
        "    return imported || builtin || {};",
        "  };",
        "  const __gojrPackageSelector = (path, field) => __gojrSelectPackageField(importsByPath, path, field);",
        "  const __gojrPackageCallableSelector = (path, field) => __gojrSelectPackageCallableField(importsByPath, path, field);",
        usesWasm ? "  const wasmBase64 = options.wasmBase64 || runtime.wasmBase64;" : "",
        usesWasm ? "  if (typeof wasmBase64 !== \"string\" || wasmBase64.length === 0) throw new Error(\"gojr Stage 1 package requires options.wasmBase64 from _gojr.wasm\");" : "",
        usesWasm ? "  const __gojrWasmModule = new WebAssembly.Module(__gojrDecodeBase64(wasmBase64));" : "",
        usesWasm ? "  const __gojrWasmInstance = new WebAssembly.Instance(__gojrWasmModule, {});" : "",
        usesWasm ? "  const __gojrWasmExports = __gojrWasmInstance.exports;" : "  const __gojrWasmExports = {};",
        ...bodyLines,
        "  Object.defineProperty(pkg, \"__gojrTypeDescriptors\", { value: __gojrTypeDescriptors });",
        "  Object.defineProperty(pkg, \"__gojrExportIndex\", { value: gojrPackageArtifact.exportIndex });",
        "  return { diagnostics: [], output: [], package: pkg, wasm: __gojrWasmExports };",
        "}",
        "function __gojrStringByteAt(value, index) {",
        "  const bytes = new TextEncoder().encode(String(value));",
        "  if (index < 0 || index >= bytes.length) throw new RangeError(\"string index out of range\");",
        "  return bytes[index];",
        "}",
        "function __gojrStringFrom(value) {",
        "  if (value instanceof Uint8Array) return new TextDecoder().decode(value);",
        "  if (ArrayBuffer.isView(value)) return new TextDecoder().decode(new Uint8Array(value.buffer, value.byteOffset, value.byteLength));",
        "  if (Array.isArray(value)) return new TextDecoder().decode(Uint8Array.from(value, (item) => Number(item) & 255));",
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
        "function __gojrEqual(left, right) {",
        "  if (left && left.__gojrInterface === true && right === null) return left.value === null;",
        "  if (right && right.__gojrInterface === true && left === null) return right.value === null;",
        "  if (left && left.__gojrTypedNil === true && right === null) return true;",
        "  if (right && right.__gojrTypedNil === true && left === null) return true;",
        "  if (left && left.__gojrInterface === true) left = left.value;",
        "  if (right && right.__gojrInterface === true) right = right.value;",
        "  if (left && left.__gojrTypedNil === true && right && right.__gojrTypedNil === true) return left.__gojrType === right.__gojrType;",
        "  if (left && typeof left === \"object\" && \"real\" in left && \"imag\" in left) {",
        "    right = __gojrToComplex(right);",
        "    return left.real === right.real && left.imag === right.imag;",
        "  }",
        "  if (right && typeof right === \"object\" && \"real\" in right && \"imag\" in right) {",
        "    left = __gojrToComplex(left);",
        "    return left.real === right.real && left.imag === right.imag;",
        "  }",
        "  return left === right;",
        "}",
        "function __gojrMapGetOk(map, key, zero) {",
        "  if (!(map instanceof Map)) return __gojrTuple([zero, false]);",
        "  if (zero === null && Object.prototype.hasOwnProperty.call(map, \"__gojrValueZero\")) zero = map.__gojrValueZero;",
        "  return map.has(key) ? __gojrTuple([map.get(key), true]) : __gojrTuple([zero, false]);",
        "}",
        "function __gojrMapSet(map, key, value) {",
        "  if (map && map.__gojrTypedNil === true && __gojrDescriptorForTypeName(map.__gojrType)?.kind === \"map\") throw new Error(\"GOJR_RUNTIME001: assignment to entry in nil map\");",
        "  if (!(map instanceof Map)) throw new Error(\"GOJR_RUNTIME001: assignment target is not a map\");",
        "  map.set(key, value);",
        "}",
        "function __gojrMap(map, valueZero) {",
        "  Object.defineProperty(map, \"__gojrValueZero\", { value: valueZero });",
        "  return map;",
        "}",
        "function __gojrBytesBase64(base64) {",
        "  return __gojrDecodeBase64(base64);",
        "}",
        "function __gojrBytesFrom(value) {",
        "  if (typeof value === \"string\") return new TextEncoder().encode(value);",
        "  if (value instanceof Uint8Array) return new Uint8Array(value);",
        "  if (ArrayBuffer.isView(value)) return new Uint8Array(value.buffer.slice(value.byteOffset, value.byteOffset + value.byteLength));",
        "  if (Array.isArray(value)) return Uint8Array.from(value, (item) => Number(item) & 255);",
        "  return new Uint8Array();",
        "}",
        "function __gojrIsByteSliceTypeName(typeName) {",
        "  return typeName === \"[]byte\" || typeName === \"[]uint8\";",
        "}",
        "function __gojrMapStringBytesBase64(entries) {",
        "  const map = new Map();",
        "  for (const [key, base64] of entries) map.set(key, __gojrBytesBase64(base64));",
        "  return __gojrMap(map, new Uint8Array());",
        "}",
        "function __gojrLen(value) {",
        "  if (value == null) return 0;",
        "  if (typeof value === \"string\" || Array.isArray(value) || ArrayBuffer.isView(value)) return value.length;",
        "  if (value instanceof Map) return value.size;",
        "  if (value.__gojrChannel === true) return value.queue.length;",
        "  return 0;",
        "}",
        "function __gojrCap(value) {",
        "  if (value == null) return 0;",
        "  if (typeof value.__gojrCap === \"number\") return value.__gojrCap;",
        "  if (Array.isArray(value) || ArrayBuffer.isView(value)) return value.length;",
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
        "  if (__gojrIsByteSliceTypeName(typeName)) {",
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
        "  if (__gojrIntegerTypeName(elemType) && value !== null && value !== undefined) return BigInt(value);",
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
        "function __gojrFixedArrayLength(typeName) {",
        "  const match = /^\\[(\\d+)\\]/.exec(String(typeName || \"\"));",
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
        "  const sliced = typeof target === \"string\" ? target.slice(low, high) : (typeof target.slice === \"function\" ? target.slice(low, high) : []);",
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
        "  if (text.startsWith(\"map[\")) return __gojrMap(new Map(), __gojrZero(__gojrMapValueType(text) || \"any\"));",
        "  if (text.startsWith(\"chan \")) return __gojrMakeChan(text.slice(5).trim(), capacity);",
        "  if (text.startsWith(\"chan<-\")) return __gojrMakeChan(text.slice(6).trim(), capacity);",
        "  if (text.startsWith(\"<-chan\")) return __gojrMakeChan(text.slice(6).trim(), capacity);",
        "  const elem = __gojrSliceElementType(text);",
        "  if (elem !== undefined) return __gojrSetCap(Array.from({ length }, () => __gojrZero(elem)), capacity);",
        "  const descriptor = __gojrDescriptorForTypeName(text);",
        "  if (descriptor && descriptor.kind === \"slice\") return __gojrSetCap(Array.from({ length }, () => __gojrZero(descriptor.elem || \"any\")), capacity);",
        "  if (descriptor && descriptor.kind === \"map\") return __gojrMap(new Map(), __gojrZero(descriptor.elem || \"any\"));",
        "  if (descriptor && descriptor.kind === \"chan\") return __gojrMakeChan(descriptor.elem || \"any\", capacity);",
        "  throw new Error(`GOJR_RUNTIME001: cannot make ${text}`);",
        "}",
        "function __gojrSliceElementType(typeName) {",
        "  return String(typeName || \"\").startsWith(\"[]\") ? String(typeName).slice(2).trim() : undefined;",
        "}",
        "function __gojrMapValueType(typeName) {",
        "  const text = String(typeName || \"\");",
        "  if (!text.startsWith(\"map[\")) return undefined;",
        "  let depth = 0;",
        "  for (let index = 4; index < text.length; index += 1) {",
        "    const char = text[index];",
        "    if (char === \"[\") depth += 1;",
        "    if (char === \"]\") {",
        "      if (depth === 0) return text.slice(index + 1).trim();",
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
        "  if (typeof src === \"string\") src = new TextEncoder().encode(src);",
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
        "  if (map instanceof Map) map.delete(key);",
        "  return null;",
        "}",
        "function __gojrPanic(value) {",
        "  throw value instanceof Error ? value : new Error(`panic: ${String(value)}`);",
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
        "  const value = object[index];",
        "  if (integerResult || object instanceof Uint8Array || object instanceof Uint8ClampedArray || object instanceof Int8Array || object instanceof Uint16Array || object instanceof Int16Array || object instanceof Uint32Array || object instanceof Int32Array || object instanceof BigInt64Array || object instanceof BigUint64Array) return BigInt(value);",
        "  return value;",
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
        "function __gojrBuiltinImport(path, importsByPath) {",
        "  if (path === \"internal/reflectlite\") return __gojrReflectlitePackage(importsByPath);",
        "  if (path === \"os\") return __gojrOsPackage();",
        "  if (path === \"reflect\") return __gojrReflectPackage(importsByPath);",
        "  if (path === \"runtime\") return __gojrRuntimePackage();",
        "  if (path === \"unsafe\") return __gojrUnsafePackage();",
        "  return undefined;",
        "}",
        `const __gojrGeneratedBuiltinImportPaths = new Set(${generatedBuiltinImports});`,
        "function __gojrBuiltinImportOverrides(path) {",
        "  return __gojrGeneratedBuiltinImportPaths.has(path);",
        "}",
        "function __gojrUnsafePackage() {",
        "  return {",
        "    Add: (ptr, _offset) => ptr,",
        "    Alignof: (_value) => 8n,",
        "    Offsetof: (_value) => 0n,",
        "    Sizeof: (_value) => 8n,",
        "    Slice: (ptr, len) => {",
        "      const length = Number(len || 0);",
        "      if (ptr instanceof Uint8Array) return ptr.subarray(0, length);",
        "      if (ArrayBuffer.isView(ptr)) return Array.from(ptr).slice(0, length);",
        "      return Array.from({ length }, () => 0n);",
        "    },",
        "    SliceData: (slice) => slice,",
        "    String: (ptr, len) => {",
        "      const length = Number(len || 0);",
        "      if (ptr instanceof Uint8Array) return new TextDecoder().decode(ptr.subarray(0, length));",
        "      if (Array.isArray(ptr)) return new TextDecoder().decode(Uint8Array.from(ptr.slice(0, length), (item) => Number(item) & 255));",
        "      return \"\";",
        "    },",
        "    StringData: (value) => value",
        "  };",
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
        "    O_RDONLY: 0n, O_WRONLY: 1n, O_RDWR: 2n, O_APPEND: 8n, O_CREATE: 512n, O_EXCL: 2048n, O_SYNC: 128n, O_TRUNC: 1024n,",
        "    ModeDir: 2147483648n, ModeAppend: 1073741824n, ModeExclusive: 536870912n, ModeTemporary: 268435456n, ModeSymlink: 134217728n, ModeDevice: 67108864n, ModeNamedPipe: 33554432n, ModeSocket: 16777216n, ModeSetuid: 8388608n, ModeSetgid: 4194304n, ModeCharDevice: 2097152n, ModeSticky: 1048576n, ModeIrregular: 524288n, ModeType: 2399666176n, ModePerm: 511n,",
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
        "    WriteString: async (_self, s) => { const bytes = new TextEncoder().encode(String(s ?? \"\")); return __gojrTuple([__gojrOsWrite(fd, bytes), null]); }",
        "  };",
        "  return pointer;",
        "}",
        "function __gojrOsRead(fd, target) {",
        "  const value = __gojrDerefIfPointer(target);",
        "  const length = __gojrLen(value);",
        "  if (length <= 0) return 0n;",
        "  const buffer = new Uint8Array(length);",
        "  const hostRead = globalThis.__gojrReadSync;",
        "  const count = typeof hostRead === \"function\" ? Number(hostRead(fd, buffer, 0, length, null) || 0) : 0;",
        "  for (let index = 0; index < count; index += 1) value[index] = BigInt(buffer[index] || 0);",
        "  return BigInt(count);",
        "}",
        "function __gojrOsWrite(fd, source) {",
        "  const value = __gojrDerefIfPointer(source);",
        "  const bytes = __gojrBytesFrom(value);",
        "  const hostWrite = globalThis.__gojrWriteSync;",
        "  if (typeof hostWrite === \"function\") return BigInt(Number(hostWrite(fd, bytes, 0, bytes.length, null) || bytes.length));",
        "  if (fd === 1 || fd === 2) __gojrActiveRuntimeOptions.stdout?.(new TextDecoder().decode(bytes));",
        "  return BigInt(bytes.length);",
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
        "    ValueOf: async (value) => __gojrReflectValueOf(value, importsByPath)",
        "  };",
        "}",
        "function __gojrReflectTypeOf(value, importsByPath) {",
        "  const actual = value && value.__gojrInterface === true ? value.value : value;",
        "  if (actual === null || actual === undefined) return null;",
        "  return __gojrReflectTypeForTypeName(__gojrRuntimeTypeName(actual), importsByPath, __gojrRuntimeTypePackagePath(actual));",
        "}",
        "function __gojrReflectTypeForTypeName(typeName, importsByPath = {}, pkgPath = undefined) {",
        "  const descriptor = __gojrDescriptorForTypeName(typeName, importsByPath, pkgPath);",
        "  if (!descriptor) return null;",
        "  return { __gojrReflectType: true, __gojrDescriptor: descriptor, __gojrImportsByPath: importsByPath, __gojrMethods: __gojrReflectTypeMethods };",
        "}",
        "function __gojrDescriptorForTypeName(typeName, importsByPath = __gojrActiveImportsByPath, pkgPath = undefined) {",
        "  if (typeName === undefined || typeName === null || typeName === \"nil\") return undefined;",
        "  const text = String(typeName).trim();",
        "  if ((!pkgPath || pkgPath === gojrPackageArtifact.importPath) && __gojrTypeDescriptors[text]) return __gojrTypeDescriptors[text];",
        "  const local = __gojrReceiverBaseType(text);",
        "  if ((!pkgPath || pkgPath === gojrPackageArtifact.importPath) && __gojrTypeDescriptors[local]) return __gojrTypeDescriptors[local];",
        "  if (text.startsWith(\"*\")) {",
        "    const elem = __gojrDescriptorForTypeName(text.slice(1).trim(), importsByPath, pkgPath);",
        "    return { type: text, name: \"\", string: `*${elem ? __gojrReflectDescriptorString(elem) : __gojrReceiverBaseType(text.slice(1).trim())}`, kind: \"ptr\", pkgPath: elem ? elem.pkgPath : \"\", pkgName: elem ? elem.pkgName : \"\", elem: elem ? elem.type : text.slice(1).trim(), elemPkgPath: elem ? elem.pkgPath : undefined };",
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
        "  Implements: async () => true",
        "};",
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
        "  if (typeName.startsWith(\"[]\")) return \"slice\";",
        "  if (typeName.startsWith(\"map[\")) return \"map\";",
        "  if (typeName.startsWith(\"chan \") || typeName.startsWith(\"<-chan\") || typeName.startsWith(\"chan<-\")) return \"chan\";",
        "  if (typeName.startsWith(\"func(\")) return \"func\";",
        "  return typeName;",
        "}",
        "function __gojrReflectKindValue(kind) {",
        "  switch (kind) {",
        "    case \"bool\": return 1n; case \"int\": return 2n; case \"int8\": return 3n; case \"int16\": return 4n; case \"int32\": return 5n; case \"int64\": return 6n;",
        "    case \"uint\": return 7n; case \"uint8\": return 8n; case \"uint16\": return 9n; case \"uint32\": return 10n; case \"uint64\": return 11n; case \"uintptr\": return 12n;",
        "    case \"float32\": return 13n; case \"float64\": return 14n; case \"complex64\": return 15n; case \"complex128\": return 16n;",
        "    case \"array\": return 17n; case \"chan\": return 18n; case \"func\": return 19n; case \"interface\": return 20n; case \"map\": return 21n; case \"ptr\": return 22n; case \"slice\": return 23n; case \"string\": return 24n; case \"struct\": return 25n;",
        "    default: return 0n;",
        "  }",
        "}",
        "function __gojrStruct(typeName, value, pkgPath = gojrPackageArtifact.importPath) {",
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
        "function __gojrPointer(typeName, get, set, pkgPath = gojrPackageArtifact.importPath) {",
        "  return { __gojrPointer: true, __gojrType: `*${typeName}`, __gojrElemType: typeName, __gojrPkgPath: pkgPath || gojrPackageArtifact.importPath, __gojrGet: get, __gojrSet: set };",
        "}",
        "function __gojrPointerValue(typeName, value, pkgPath = gojrPackageArtifact.importPath) {",
        "  let cell = value;",
        "  return __gojrPointer(typeName, () => cell, (next) => { cell = next; }, pkgPath);",
        "}",
        "function __gojrConvertPointer(typeName, value) {",
        "  if (value === null || value === undefined) return __gojrTypedNil(typeName);",
        "  if (value && value.__gojrTypedNil === true) return __gojrTypedNil(typeName);",
        "  if (value && value.__gojrPointer === true) {",
        "    const elemType = String(typeName).startsWith(\"*\") ? String(typeName).slice(1).trim() : String(typeName);",
        "    return __gojrPointer(elemType, value.__gojrGet, value.__gojrSet);",
        "  }",
        "  return value;",
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
        "function __gojrBitwiseXor(left, right) {",
        "  if (__gojrPointerLike(left) && __gojrIsZeroInteger(right)) return left;",
        "  if (__gojrPointerLike(right) && __gojrIsZeroInteger(left)) return right;",
        "  return BigInt(left) ^ BigInt(right);",
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
        "  return __gojrPointer(typeName, () => __gojrGetField(object, field), (next) => { __gojrSetField(object, field, next); }, pkgPath);",
        "}",
        "function __gojrAddressIndex(object, index, typeName, pkgPath = gojrPackageArtifact.importPath) {",
        "  const target = __gojrDerefIfPointer(object);",
        "  const numericIndex = Number(index);",
        "  if (target == null) throw new Error(\"GOJR_RUNTIME001: cannot take address of nil index target\");",
        "  return __gojrPointer(typeName, () => target[numericIndex], (next) => { target[numericIndex] = next; }, pkgPath);",
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
        "  if (typeof value === \"string\") return \"string\";",
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
        "  const typeName = actual && (actual.__gojrType || __gojrRuntimeTypeName(actual));",
        "  const methodKey = typeName ? `${__gojrReceiverBaseType(typeName)}.${method}` : undefined;",
        "  let fn = methodKey ? pkg[methodKey] : undefined;",
        "  if (typeof fn !== \"function\" && methodKey) {",
        "    const receiverPkgPath = __gojrRuntimeTypePackagePath(actual);",
        "    if (receiverPkgPath && receiverPkgPath !== gojrPackageArtifact.importPath) {",
        "      const importedFn = __gojrSelectPackageField(__gojrActiveImportsByPath, receiverPkgPath, methodKey);",
        "      if (typeof importedFn === \"function\") fn = importedFn;",
        "    }",
        "  }",
        "  if (typeof fn !== \"function\") throw new TypeError(`method ${method} not found`);",
        "  return await fn(actual, ...args);",
        "}",
        "async function __gojrDeferScope(body) {",
        "  const stack = [];",
        "  const push = (fn) => { stack.push(fn); };",
        "  let result;",
        "  let thrown;",
        "  try {",
        "    result = await body(push);",
        "  } catch (error) {",
        "    thrown = error;",
        "  }",
        "  while (stack.length > 0) {",
        "    try {",
        "      await stack.pop()();",
        "    } catch (error) {",
        "      thrown = error;",
        "    }",
        "  }",
        "  if (thrown !== undefined) throw thrown;",
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
        "  return { __gojrChannel: true, elementType, capacity: Math.max(0, Number(capacity)), queue: [], receivers: [], senders: [] };",
        "}",
        "async function __gojrChanSend(channel, value) {",
        "  if (!channel || channel.__gojrChannel !== true) throw new TypeError(\"send on non-channel\");",
        "  if (channel.receivers.length > 0) {",
        "    const receiver = channel.receivers.shift();",
        "    receiver(__gojrTuple([value, true]));",
        "    return null;",
        "  }",
        "  if (channel.queue.length < channel.capacity) {",
        "    channel.queue.push(value);",
        "    return null;",
        "  }",
        "  return await new Promise((resolve) => { channel.senders.push({ value, resolve }); });",
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
        "  return Boolean(channel && channel.__gojrChannel === true && (channel.queue.length > 0 || channel.senders.length > 0));",
        "}",
        "function __gojrChanCanSend(channel) {",
        "  return Boolean(channel && channel.__gojrChannel === true && (channel.receivers.length > 0 || channel.queue.length < channel.capacity));",
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
        "  if (typeof source[Symbol.iterator] === \"function\") return Array.from(source);",
        "  throw new TypeError(\"GOJR_RUNTIME001: spread argument is not a slice\");",
        "}",
        "function __gojrRangeEntries(source) {",
        "  if (source == null) return [];",
        "  if (source && source.__gojrTypedNil === true) return [];",
        "  if (typeof source === \"bigint\" || typeof source === \"number\") {",
        "    const count = Number(source);",
        "    const entries = [];",
        "    for (let i = 0; i < count; i += 1) entries.push([BigInt(i), BigInt(i)]);",
        "    return entries;",
        "  }",
        "  if (typeof source === \"string\") {",
        "    const entries = [];",
        "    let byteIndex = 0;",
        "    for (const rune of source) {",
        "      entries.push([BigInt(byteIndex), BigInt(rune.codePointAt(0) ?? 0)]);",
        "      byteIndex += new TextEncoder().encode(rune).length;",
        "    }",
        "    return entries;",
        "  }",
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
        "  if (String(typeText || \"\").startsWith(\"*\")) return Boolean((value && value.__gojrPointer === true && value.__gojrType === typeText) || (value && value.__gojrTypedNil === true && value.__gojrType === typeText));",
        "  switch (typeText) {",
        "    case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return typeof value === \"bigint\";",
        "    case \"float32\": case \"float64\": return typeof value === \"number\";",
        "    case \"string\": return typeof value === \"string\";",
        "    case \"bool\": return typeof value === \"boolean\";",
        "    case \"complex64\": case \"complex128\": return Boolean(value && typeof value === \"object\" && \"real\" in value && \"imag\" in value);",
        "    default: return __gojrReceiverBaseType(__gojrRuntimeTypeName(value)) === __gojrReceiverBaseType(typeText);",
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
        "      if (kind === \"array\" && typeof descriptor.len === \"number\") return Array.from({ length: descriptor.len }, () => __gojrZero(descriptor.elem, descriptor.elemPkgPath, importsByPath));",
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
}
