import { EmitterContext } from "./context.js";
import { checkedInWasmStencil } from "./stencils.js";
export const GOJR_STAGE1_BACKEND = "copy-patch-wasm-stage1";
const emptyPackageEmitFacts = {
    methodKeys: new Map(),
    resultCounts: new Map(),
    packageTypes: new Map(),
    imports: new Map(),
    interfaceTypes: new Set(),
    functionParamTypes: new Map(),
    functionResultTypes: new Map()
};
const emptyExpressionEnv = { locals: new Map(), localTypes: new Map(), facts: emptyPackageEmitFacts };
export function emitStage1Package(artifact, ast) {
    const ctx = new EmitterContext({ artifact });
    const facts = packageEmitFacts(ast, artifact);
    const wasmFunctions = ast.functions.filter(isI64AddFunction);
    const usesWasm = wasmFunctions.length > 0;
    const wasmBase64 = usesWasm ? checkedInWasmStencil("i64.add.kernel").wasmBase64 : undefined;
    const declarationLines = [];
    const functionLines = [];
    const initNames = [];
    for (const statement of ast.body) {
        if (statement.kind === "ConstDecl") {
            declarationLines.push(...emitConstDecl(ctx, statement, facts));
            continue;
        }
        if (statement.kind === "VarDecl") {
            declarationLines.push(...emitVarDecl(ctx, statement, facts));
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
        functionLines.push(...emitFunction(ctx, fn, facts));
    }
    const bodyLines = [
        ...functionLines,
        ...declarationLines,
        ...initNames.map((name) => `  await ${name}();`)
    ];
    const javascript = stage1JavaScript(artifact, usesWasm, bodyLines);
    return {
        diagnostics: ctx.diagnostics,
        javascript,
        ...(wasmBase64 ? { wasmBase64 } : {})
    };
}
function emitConstDecl(ctx, statement, facts) {
    return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "const", declaration, facts));
}
function emitVarDecl(ctx, statement, facts) {
    return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "var", declaration, facts));
}
function emitDeclaration(ctx, _kind, declaration, facts) {
    const rawValue = declaration.value
        ? expressionToJs(ctx, declaration.value)
        : zeroValueForType(declaration.type?.text);
    if (!rawValue) {
        ctx.emitError(`unsupported Stage 1 declaration for ${declaration.name}`);
        return [];
    }
    const targetType = declarationTypeText(declaration);
    const value = valueForTargetType(rawValue, targetType, expressionTypeText(declaration.value, emptyExpressionEnv), { ...emptyExpressionEnv, facts });
    return [`  pkg[${JSON.stringify(declaration.name)}] = ${value};`];
}
function emitFunction(ctx, fn, facts, options = {}) {
    if (isI64AddFunction(fn)) {
        if (options.localName) {
            return [
                `  const ${options.localName} = async (a, b) => __gojrWasmExports.add_i64(BigInt(a), BigInt(b));`
            ];
        }
        return [
            `  pkg[${JSON.stringify(functionPackageKey(fn))}] = async (a, b) => __gojrWasmExports.add_i64(BigInt(a), BigInt(b));`
        ];
    }
    const parameters = functionParameterBindings(fn, facts);
    const bodyEnv = cloneExpressionEnv(parameters.env);
    bodyEnv.expectedReturnTypes = fn.signature.results.map((result) => result.type.text);
    const hasDefer = statementsContainDefer(fn.body.statements);
    if (hasDefer)
        bodyEnv.deferName = ctx.symbol("defer");
    const resultLines = emitNamedResultDeclarations(fn, bodyEnv);
    const statementLines = statementsContainGoto(fn.body.statements)
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
    return rendered ? valueForTargetType(rendered, env.expectedReturnTypes?.[0], expressionTypeText(statement.values[0], env), env) : undefined;
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
        lines.push(`let ${name} = ${zeroValueForType(result.type.text) ?? "null"};`);
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
        return zeroValueForType(fn.signature.results[0]?.type.text) ?? "null";
    return `__gojrTuple([${fn.signature.results.map((result) => zeroValueForType(result.type.text) ?? "null").join(", ")}])`;
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
        ...(env.deferName ? { deferName: env.deferName } : {}),
        ...(env.gotoLabels ? { gotoLabels: env.gotoLabels } : {}),
        ...(env.gotoPcName ? { gotoPcName: env.gotoPcName } : {})
    };
}
function emitStatements(ctx, statements, env, indent) {
    const lines = [];
    for (const statement of statements) {
        const emitted = emitStatement(ctx, statement, env, indent);
        if (!emitted)
            return undefined;
        lines.push(...emitted);
    }
    return lines;
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
    for (const [index, declaration] of statement.declarations.entries()) {
        if (declaration.name === "_")
            continue;
        const name = safeLocalName(declaration.name, index);
        const rawValue = declaration.value
            ? expressionToJs(ctx, declaration.value, env)
            : zeroValueForType(declaration.type?.text);
        if (!rawValue) {
            ctx.emitError(`unsupported Stage 4 declaration for ${declaration.name}`);
            return undefined;
        }
        env.locals.set(declaration.name, name);
        const typeText = declarationTypeText(declaration);
        if (typeText)
            env.localTypes.set(declaration.name, typeText);
        const value = valueForTargetType(rawValue, typeText, expressionTypeText(declaration.value, env), env);
        lines.push(`${indent}${keyword} ${name} = ${value};`);
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
    const range = ctx.symbol("range");
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
        `${indent}const ${range} = ${source};`,
        `${indent}for (const [${key}, ${value}] of __gojrRangeEntries(${range})) {`,
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
            if (statement.names.length !== 1 || statement.values.length !== 1 || statement.names[0] === "_") {
                ctx.emitError("unsupported Stage 4 for init short declaration");
                return undefined;
            }
            {
                const goName = statement.names[0];
                const expression = statement.values[0];
                if (!goName || !expression)
                    return undefined;
                const value = expressionToJs(ctx, expression, env);
                if (!value)
                    return undefined;
                const name = safeLocalName(goName, 0);
                env.locals.set(goName, name);
                return `let ${name} = ${value}`;
            }
        case "AssignStatement":
            if (statement.targets.length !== 1 || statement.values.length !== 1) {
                ctx.emitError("unsupported Stage 4 for header assignment");
                return undefined;
            }
            {
                const target = expressionToJs(ctx, statement.targets[0], env);
                const value = expressionToJs(ctx, statement.values[0], env);
                if (!target || !value)
                    return undefined;
                if (statement.operator === "=")
                    return `${target} = ${value}`;
                const compound = compoundAssignmentExpression(ctx, statement.operator, target, value);
                return compound ? `${target} = ${compound}` : undefined;
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
        const zero = (expression.typeText ? zeroValueForType(expression.typeText) : undefined) ??
            zeroValueForType(mapValueTypeText(objectType)) ??
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
            return [`${indent}(${object})[${JSON.stringify(target.field)}] = ${value};`];
        }
        case "IndexExpression": {
            const object = expressionToJs(ctx, target.object, env);
            const index = expressionToJs(ctx, target.index, env);
            if (!object || !index)
                return undefined;
            if (((expressionTypeText(target.object, env) ?? packageExportTypeText(ctx, target.object)) ?? "").startsWith("map["))
                return [`${indent}(${object}).set(${index}, ${value});`];
            return [`${indent}(${object})[Number(${index})] = ${value};`];
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
    return { params, env: { locals, localTypes, facts } };
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
function packageEmitFacts(ast, artifact) {
    const methodKeys = new Map();
    const resultCounts = new Map();
    const packageTypes = new Map();
    const imports = new Map();
    const interfaceTypes = new Set(["any", "interface{}", "error"]);
    const functionParamTypes = new Map();
    const functionResultTypes = new Map();
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
        if (!fn.receiver) {
            resultCounts.set(fn.name, fn.signature.results.length);
            functionParamTypes.set(fn.name, fn.signature.parameters.map((parameter) => parameter.type.text));
            functionResultTypes.set(fn.name, fn.signature.results.map((result) => result.type.text));
            continue;
        }
        const receiver = receiverBaseType(fn.receiver.type.text);
        methodKeys.set(`${receiver}.${fn.name}`, key);
    }
    return { methodKeys, resultCounts, packageTypes, imports, interfaceTypes, functionParamTypes, functionResultTypes };
}
function isInterfaceTypeSpec(spec) {
    return Boolean(spec.interfaceMethods || spec.interfaceEmbeds || spec.type.text.trim().startsWith("interface{"));
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
    return trimmed.startsWith("*") ? trimmed.slice(1).trim() : trimmed;
}
function expressionTypeText(expression, env) {
    if (!expression)
        return undefined;
    if (expression.typeText)
        return expression.typeText;
    if (expression.kind === "Identifier")
        return env.localTypes.get(expression.name) ?? env.facts.packageTypes.get(expression.name);
    if (expression.kind === "IndexExpression")
        return mapValueTypeText(expressionTypeText(expression.object, env));
    if (expression.kind === "MapLiteralExpression")
        return `map[${expression.keyType.text}]${expression.valueType.text}`;
    if (expression.kind === "ArrayLiteralExpression")
        return expression.type.text;
    if (expression.kind === "StructLiteralExpression")
        return expression.typeName;
    if (expression.kind === "FunctionLiteralExpression")
        return expression.typeText;
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
function declarationTypeText(declaration) {
    if (declaration.type?.text)
        return declaration.type.text;
    if (declaration.value)
        return expressionTypeText(declaration.value, emptyExpressionEnv);
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
    if (typeText === "string")
        return { key: "int", value: "rune" };
    const mapType = parseMapTypeText(typeText);
    if (mapType)
        return { key: mapType.keyType, value: mapType.valueType };
    const sliceType = parseArrayOrSliceTypeText(typeText);
    if (sliceType)
        return { key: "int", value: sliceType.elementType };
    if (isIntegerType(typeText))
        return { key: typeText, value: typeText };
    return {};
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
        case "FunctionLiteralExpression":
            return functionLiteralToJs(ctx, expression, env);
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
            return typeAssertionExpressionToJs(ctx, expression, env);
        default:
            ctx.emitError(`unsupported Stage 3 expression ${expression.kind}`);
            return undefined;
    }
}
function identifierToJs(expression, env) {
    const local = env.locals.get(expression.name);
    if (local)
        return local;
    const importPath = env.facts.imports.get(expression.name);
    if (importPath)
        return `__gojrImport(${JSON.stringify(importPath)})`;
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
    if (expression.elements.some((element) => element.key !== undefined)) {
        ctx.emitError("unsupported Stage 3 keyed array literal");
        return undefined;
    }
    if (isByteSliceType(expression.type.text)) {
        const bytes = bytesFromArrayLiteral(expression);
        if (!bytes) {
            ctx.emitError(`unsupported Stage 1 []byte literal element`);
            return undefined;
        }
        if (bytes.length > 64)
            return `__gojrBytesBase64(${JSON.stringify(bytesToBase64(bytes))})`;
        return `new Uint8Array([${bytes.join(", ")}])`;
    }
    const values = expression.elements.map((element) => expressionToJs(ctx, element.value, env));
    if (values.some((value) => value === undefined))
        return undefined;
    const rendered = values.filter((value) => value !== undefined);
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
    return `__gojrStruct(${JSON.stringify(expression.typeName)}, { ${fields.join(", ")} })`;
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
    return `__gojrMap(new Map([${entries.join(", ")}]), ${zeroValueForType(expression.valueType.text) ?? "null"})`;
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
        resultLines.push(`  let ${name} = ${zeroValueForType(result.type.text) ?? "null"};`);
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
        return zeroValueForType(expression.signature.results[0]?.type.text) ?? "null";
    return `__gojrTuple([${expression.signature.results.map((result) => zeroValueForType(result.type.text) ?? "null").join(", ")}])`;
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
        case "<-":
            return `(await __gojrChanRecv(${operand}))[0]`;
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
    if (isComplexType(expression.typeText) || isComplexType(expression.left.typeText) || isComplexType(expression.right.typeText)) {
        return complexBinaryExpressionToJs(ctx, expression, left, right);
    }
    if (expression.operator === "&&" || expression.operator === "||")
        return `((${left}) ${expression.operator} (${right}))`;
    if (expression.operator === "==")
        return `__gojrEqual(${left}, ${right})`;
    if (expression.operator === "!=")
        return `(!__gojrEqual(${left}, ${right}))`;
    if (expression.operator === "&^")
        return `((${left}) & ~(${right}))`;
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
    const methodKey = methodPackageKeyForSelector(expression, env);
    if (methodKey) {
        const receiver = expressionToJs(ctx, expression.object, env);
        if (!receiver)
            return undefined;
        return `(async (...__gojrMethodArgs) => await pkg[${JSON.stringify(methodKey)}](${receiver}, ...__gojrMethodArgs))`;
    }
    const object = expressionToJs(ctx, expression.object, env);
    if (!object)
        return undefined;
    return `(${object})[${JSON.stringify(expression.field)}]`;
}
function methodPackageKeyForSelector(expression, env) {
    const receiverType = expressionTypeText(expression.object, env);
    if (!receiverType)
        return undefined;
    return env.facts.methodKeys.get(`${receiverBaseType(receiverType)}.${expression.field}`);
}
function methodCallToJs(ctx, expression, env, methodKey) {
    const callee = expression.callee;
    const receiver = expressionToJs(ctx, callee.object, env);
    if (!receiver)
        return undefined;
    const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(methodKey), expression.spreadLast);
    if (!args)
        return undefined;
    return `(await pkg[${JSON.stringify(methodKey)}](${[receiver, ...args].join(", ")}))`;
}
function callExpressionToJs(ctx, expression, env) {
    if (expression.callee.kind === "Identifier" && expression.callee.name === "make" && !env.locals.has("make")) {
        return makeCallToJs(ctx, expression, env);
    }
    if (expression.callee.kind === "TypeExpression")
        return conversionCallToJs(ctx, expression.callee.type.text, expression.args, env);
    if (expression.callee.kind === "Identifier" && primitiveTypeNames.has(expression.callee.name) && !env.locals.has(expression.callee.name)) {
        return conversionCallToJs(ctx, expression.callee.name, expression.args, env);
    }
    const builtin = builtinCallToJs(ctx, expression, env);
    if (builtin)
        return builtin;
    if (expression.callee.kind === "SelectorExpression") {
        const methodKey = methodPackageKeyForSelector(expression.callee, env);
        if (methodKey)
            return methodCallToJs(ctx, expression, env, methodKey);
        if (!isImportedPackageSelector(expression.callee, env))
            return dynamicMethodCallToJs(ctx, expression, env);
    }
    const instantiatedName = instantiatedFunctionName(expression.callee);
    if (instantiatedName && !env.locals.has(instantiatedName)) {
        const args = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(instantiatedName), expression.spreadLast);
        if (!args)
            return undefined;
        return `(await pkg[${JSON.stringify(instantiatedName)}](${args.join(", ")}))`;
    }
    if (expression.callee.kind === "Identifier" && !env.locals.has(expression.callee.name)) {
        const renderedArgs = renderCallArgs(ctx, expression.args, env, env.facts.functionParamTypes.get(expression.callee.name), expression.spreadLast);
        if (!renderedArgs)
            return undefined;
        return `(await pkg[${JSON.stringify(expression.callee.name)}](${renderedArgs.join(", ")}))`;
    }
    const callee = expressionToJs(ctx, expression.callee, env);
    if (!callee)
        return undefined;
    const renderedArgs = renderCallArgs(ctx, expression.args, env, [], expression.spreadLast);
    if (!renderedArgs)
        return undefined;
    return `(await (${callee})(${renderedArgs.join(", ")}))`;
}
function renderCallArgs(ctx, args, env, targetTypes = [], spreadLast = false) {
    const rendered = [];
    for (const [index, arg] of args.entries()) {
        const value = expressionToJs(ctx, arg, env);
        if (!value)
            return undefined;
        const lowered = valueForTargetType(value, targetTypes[index], expressionTypeText(arg, env), env);
        rendered.push(spreadLast && index === args.length - 1 ? `...${lowered}` : lowered);
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
function makeCallToJs(ctx, expression, env) {
    const typeArg = expression.args[0];
    if (!typeArg || typeArg.kind !== "TypeExpression") {
        ctx.emitError("unsupported Stage 4 make without type argument");
        return undefined;
    }
    const typeText = typeArg.type.text;
    if (typeText.startsWith("chan ")) {
        const capacity = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
        if (!capacity)
            return undefined;
        return `__gojrMakeChan(${JSON.stringify(chanElementTypeText(typeText) ?? "any")}, Number(${capacity}))`;
    }
    if (typeText.startsWith("map[")) {
        return `__gojrMap(new Map(), ${zeroValueForType(mapValueTypeText(typeText)) ?? "null"})`;
    }
    const slice = parseArrayOrSliceTypeText(typeText);
    if (slice && typeText.startsWith("[]")) {
        const length = expression.args[1] ? expressionToJs(ctx, expression.args[1], env) : "0n";
        if (!length)
            return undefined;
        return `Array.from({ length: Number(${length}) }, () => ${zeroValueForType(slice.elementType) ?? "null"})`;
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
    if (isIntegerType(typeText))
        return `BigInt(${value})`;
    if (isFloatType(typeText))
        return `Number(${value})`;
    if (isComplexType(typeText))
        return `__gojrToComplex(${value})`;
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
    const objectType = expressionTypeText(expression.object, env) ?? packageExportTypeText(ctx, expression.object) ?? "";
    if (objectType.startsWith("map[")) {
        return `((${object}).get(${index}) ?? ${zeroValueForType(expression.typeText) ?? "null"})`;
    }
    if (objectType === "string")
        return `BigInt(__gojrStringByteAt(${object}, Number(${index})))`;
    const elementType = expression.typeText ?? parseArrayOrSliceTypeText(objectType)?.elementType;
    return `__gojrIndex(${object}, Number(${index}), ${isIntegerType(elementType ?? "") ? "true" : "false"})`;
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
        case "complex64":
        case "complex128":
            return "__gojrComplex(0, 0)";
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
function mapValueTypeText(typeText) {
    return parseMapTypeText(typeText)?.valueType;
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
function isComplexType(typeText) {
    return typeText === "complex64" || typeText === "complex128" || typeText === "untyped complex";
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
        "  const __gojrImport = (path) => {",
        "    const imported = importsByPath[path];",
        "    if (imported && typeof imported === \"object\" && \"package\" in imported) return imported.package;",
        "    return imported || {};",
        "  };",
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
        "function __gojrTuple(values) {",
        "  Object.defineProperty(values, \"__gojrTuple\", { value: true });",
        "  return values;",
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
        "function __gojrMap(map, valueZero) {",
        "  Object.defineProperty(map, \"__gojrValueZero\", { value: valueZero });",
        "  return map;",
        "}",
        "function __gojrBytesBase64(base64) {",
        "  return __gojrDecodeBase64(base64);",
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
        "  if (Array.isArray(value) || ArrayBuffer.isView(value)) return value.length;",
        "  if (value.__gojrChannel === true) return value.capacity;",
        "  return __gojrLen(value);",
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
        "function __gojrIndex(object, index, integerResult) {",
        "  const value = object[index];",
        "  if (integerResult || object instanceof Uint8Array || object instanceof Uint8ClampedArray || object instanceof Int8Array || object instanceof Uint16Array || object instanceof Int16Array || object instanceof Uint32Array || object instanceof Int32Array || object instanceof BigInt64Array || object instanceof BigUint64Array) return BigInt(value);",
        "  return value;",
        "}",
        "function __gojrStruct(typeName, value) {",
        "  Object.defineProperty(value, \"__gojrType\", { value: typeName });",
        "  return value;",
        "}",
        "function __gojrTypedNil(typeName) {",
        "  return { __gojrTypedNil: true, __gojrType: typeName };",
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
        "  if (value && value.__gojrTypedNil === true) return value.__gojrType;",
        "  if (value && typeof value === \"object\" && typeof value.__gojrType === \"string\") return value.__gojrType;",
        "  if (typeof value === \"bigint\") return \"int\";",
        "  if (typeof value === \"number\") return \"float64\";",
        "  if (typeof value === \"string\") return \"string\";",
        "  if (typeof value === \"boolean\") return \"bool\";",
        "  return value === null || value === undefined ? \"nil\" : typeof value;",
        "}",
        "function __gojrReceiverBaseType(typeName) {",
        "  let name = String(typeName ?? \"\");",
        "  if (name.startsWith(\"*\")) name = name.slice(1).trim();",
        "  const dot = name.lastIndexOf(\".\");",
        "  return dot >= 0 ? name.slice(dot + 1) : name;",
        "}",
        "async function __gojrCallMethod(receiver, method, args, pkg) {",
        "  const actual = receiver && receiver.__gojrInterface === true ? receiver.value : receiver;",
        "  const typeName = actual && (actual.__gojrType || __gojrRuntimeTypeName(actual));",
        "  const fn = typeName ? pkg[`${__gojrReceiverBaseType(typeName)}.${method}`] : undefined;",
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
        "function __gojrRangeEntries(source) {",
        "  if (source == null) return [];",
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
        "  switch (typeText) {",
        "    case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return typeof value === \"bigint\";",
        "    case \"float32\": case \"float64\": return typeof value === \"number\";",
        "    case \"string\": return typeof value === \"string\";",
        "    case \"bool\": return typeof value === \"boolean\";",
        "    case \"complex64\": case \"complex128\": return Boolean(value && typeof value === \"object\" && \"real\" in value && \"imag\" in value);",
        "    default: return __gojrReceiverBaseType(__gojrRuntimeTypeName(value)) === __gojrReceiverBaseType(typeText);",
        "  }",
        "}",
        "function __gojrZero(typeText) {",
        "  switch (typeText) {",
        "    case \"int\": case \"int8\": case \"int16\": case \"int32\": case \"int64\": case \"uint\": case \"uint8\": case \"uint16\": case \"uint32\": case \"uint64\": case \"uintptr\": return 0n;",
        "    case \"float32\": case \"float64\": return 0;",
        "    case \"string\": return \"\";",
        "    case \"bool\": return false;",
        "    case \"complex64\": case \"complex128\": return __gojrComplex(0, 0);",
        "    default: return null;",
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
