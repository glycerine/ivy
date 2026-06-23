export function analyzeEffects(program) {
    const functionKeys = new Map();
    for (const declaration of program.functions) {
        functionKeys.set(declaration.name, functionEffectKey(declaration));
    }
    const summaries = new Map();
    for (const declaration of program.functions) {
        const key = functionEffectKey(declaration);
        summaries.set(key, analyzeBlock(key, declaration.name, declaration.body, functionKeys));
    }
    const topLevel = analyzeStatements("<top-level>", "<top-level>", program.body, functionKeys);
    let changed = true;
    while (changed) {
        changed = false;
        for (const summary of summaries.values()) {
            if (summary.maySuspend)
                continue;
            for (const target of summary.calls) {
                if (summaries.get(target)?.maySuspend) {
                    summary.maySuspend = true;
                    changed = true;
                    break;
                }
            }
        }
        if (!topLevel.maySuspend) {
            for (const target of topLevel.calls) {
                if (summaries.get(target)?.maySuspend) {
                    topLevel.maySuspend = true;
                    changed = true;
                    break;
                }
            }
        }
    }
    const functions = {};
    for (const [key, summary] of summaries) {
        functions[key] = freezeSummary(summary, summaries);
    }
    return {
        topLevel: freezeSummary(topLevel, summaries),
        functions
    };
}
export function functionEffectKey(declaration) {
    if (!declaration.receiver)
        return declaration.name;
    return `${declaration.receiver.type.text}.${declaration.name}`;
}
function analyzeBlock(key, name, block, functionKeys) {
    return analyzeStatements(key, name, block.statements, functionKeys);
}
function analyzeStatements(key, name, statements, functionKeys) {
    const summary = {
        key,
        name,
        intrinsicReasons: [],
        calls: new Set(),
        maySuspend: false
    };
    for (const statement of statements)
        analyzeStatement(statement, summary, functionKeys);
    summary.maySuspend = summary.intrinsicReasons.length > 0;
    return summary;
}
function analyzeStatement(statement, summary, functionKeys) {
    if (!statement)
        return;
    switch (statement.kind) {
        case "BlockStatement":
            analyzeStatementsInto(statement.statements, summary, functionKeys);
            return;
        case "LabeledStatement":
            analyzeStatement(statement.statement, summary, functionKeys);
            return;
        case "ConstDecl":
        case "VarDecl":
            for (const declaration of statement.declarations)
                analyzeExpression(declaration.value, summary, functionKeys);
            return;
        case "TypeDecl":
        case "BranchStatement":
        case "IncDecStatement":
            if (statement.kind === "IncDecStatement")
                analyzeExpression(statement.target, summary, functionKeys);
            return;
        case "ReturnStatement":
            for (const value of statement.values)
                analyzeExpression(value, summary, functionKeys);
            return;
        case "IfStatement":
            analyzeIf(statement, summary, functionKeys);
            return;
        case "SwitchStatement":
            analyzeSwitch(statement, summary, functionKeys);
            return;
        case "SelectStatement":
            addReason(summary, { kind: "select" });
            analyzeSelect(statement, summary, functionKeys);
            return;
        case "ForStatement":
            analyzeFor(statement, summary, functionKeys);
            return;
        case "DeferStatement":
            analyzeDefer(statement, summary, functionKeys);
            return;
        case "GoStatement":
            analyzeGo(statement, summary, functionKeys);
            return;
        case "SendStatement":
            analyzeSend(statement, summary, functionKeys);
            return;
        case "AssignStatement":
            analyzeAssign(statement, summary, functionKeys);
            return;
        case "ShortVarStatement":
            analyzeShortVar(statement, summary, functionKeys);
            return;
        case "ExpressionStatement":
            analyzeExpression(statement.expression, summary, functionKeys);
            return;
    }
}
function analyzeStatementsInto(statements, summary, functionKeys) {
    for (const statement of statements)
        analyzeStatement(statement, summary, functionKeys);
}
function analyzeIf(statement, summary, functionKeys) {
    analyzeStatement(statement.init, summary, functionKeys);
    analyzeExpression(statement.condition, summary, functionKeys);
    analyzeStatement(statement.thenBlock, summary, functionKeys);
    analyzeStatement(statement.elseBranch, summary, functionKeys);
}
function analyzeSwitch(statement, summary, functionKeys) {
    analyzeStatement(statement.init, summary, functionKeys);
    analyzeExpression(statement.expression, summary, functionKeys);
    analyzeExpression(statement.typeSwitch?.expression, summary, functionKeys);
    for (const clause of statement.clauses) {
        for (const value of clause.values)
            analyzeExpression(value, summary, functionKeys);
        analyzeStatementsInto(clause.statements, summary, functionKeys);
    }
}
function analyzeSelect(statement, summary, functionKeys) {
    for (const clause of statement.clauses)
        analyzeCommClause(clause, summary, functionKeys);
}
function analyzeCommClause(clause, summary, functionKeys) {
    analyzeStatement(clause.comm, summary, functionKeys);
    analyzeStatementsInto(clause.statements, summary, functionKeys);
}
function analyzeFor(statement, summary, functionKeys) {
    analyzeStatement(statement.init, summary, functionKeys);
    analyzeExpression(statement.condition, summary, functionKeys);
    analyzeStatement(statement.post, summary, functionKeys);
    analyzeExpression(statement.range?.source, summary, functionKeys);
    analyzeStatement(statement.body, summary, functionKeys);
}
function analyzeDefer(statement, summary, functionKeys) {
    analyzeExpression(statement.expression, summary, functionKeys);
}
function analyzeGo(statement, summary, functionKeys) {
    addReason(summary, { kind: "go" });
    analyzeCall(statement.call, summary, functionKeys);
}
function analyzeSend(statement, summary, functionKeys) {
    addReason(summary, { kind: "channel-send" });
    analyzeExpression(statement.channel, summary, functionKeys);
    analyzeExpression(statement.value, summary, functionKeys);
}
function analyzeAssign(statement, summary, functionKeys) {
    for (const target of statement.targets)
        analyzeExpression(target, summary, functionKeys);
    for (const value of statement.values)
        analyzeExpression(value, summary, functionKeys);
}
function analyzeShortVar(statement, summary, functionKeys) {
    for (const value of statement.values)
        analyzeExpression(value, summary, functionKeys);
}
function analyzeExpression(expression, summary, functionKeys) {
    if (!expression)
        return;
    switch (expression.kind) {
        case "Identifier":
        case "TypeExpression":
        case "Literal":
            return;
        case "FunctionLiteralExpression":
            analyzeFunctionLiteral(expression, summary, functionKeys);
            return;
        case "ArrayLiteralExpression":
            for (const element of expression.elements)
                analyzeExpression(element, summary, functionKeys);
            return;
        case "StructLiteralExpression":
            for (const field of expression.fields) {
                if (field.key)
                    analyzeExpression(field.key, summary, functionKeys);
                analyzeExpression(field.value, summary, functionKeys);
            }
            return;
        case "MapLiteralExpression":
            analyzeMapLiteral(expression, summary, functionKeys);
            return;
        case "UnaryExpression":
            analyzeUnary(expression, summary, functionKeys);
            return;
        case "BinaryExpression":
            analyzeBinary(expression, summary, functionKeys);
            return;
        case "TypeAssertionExpression":
            analyzeExpression(expression.expression, summary, functionKeys);
            return;
        case "SelectorExpression":
            analyzeExpression(expression.object, summary, functionKeys);
            return;
        case "CallExpression":
            analyzeCall(expression, summary, functionKeys);
            return;
        case "IndexExpression":
            analyzeExpression(expression.object, summary, functionKeys);
            analyzeExpression(expression.index, summary, functionKeys);
            return;
        case "SliceExpression":
            analyzeSlice(expression, summary, functionKeys);
            return;
        case "SpreadsheetRangeExpression":
            analyzeExpression(expression.start, summary, functionKeys);
            return;
    }
}
function analyzeFunctionLiteral(expression, summary, functionKeys) {
    // A function literal value does not execute where it appears. Its body is
    // analyzed when the literal is immediately called or launched.
    void expression;
    void summary;
    void functionKeys;
}
function analyzeMapLiteral(expression, summary, functionKeys) {
    for (const entry of expression.entries) {
        analyzeExpression(entry.key, summary, functionKeys);
        analyzeExpression(entry.value, summary, functionKeys);
    }
}
function analyzeUnary(expression, summary, functionKeys) {
    if (expression.operator === "<-")
        addReason(summary, { kind: "channel-receive" });
    analyzeExpression(expression.operand, summary, functionKeys);
}
function analyzeBinary(expression, summary, functionKeys) {
    analyzeExpression(expression.left, summary, functionKeys);
    analyzeExpression(expression.right, summary, functionKeys);
}
function analyzeSlice(expression, summary, functionKeys) {
    analyzeExpression(expression.object, summary, functionKeys);
    analyzeExpression(expression.start, summary, functionKeys);
    analyzeExpression(expression.end, summary, functionKeys);
    analyzeExpression(expression.max, summary, functionKeys);
}
function analyzeCall(expression, summary, functionKeys) {
    if (expression.callee.kind === "Identifier") {
        const target = functionKeys.get(expression.callee.name);
        if (target)
            summary.calls.add(target);
    }
    else if (expression.callee.kind === "FunctionLiteralExpression") {
        const literalSummary = analyzeBlock("<function-literal>", "<function-literal>", expression.callee.body, functionKeys);
        for (const reason of literalSummary.intrinsicReasons)
            addReason(summary, reason);
        for (const call of literalSummary.calls)
            summary.calls.add(call);
    }
    else {
        analyzeExpression(expression.callee, summary, functionKeys);
    }
    for (const arg of expression.args)
        analyzeExpression(arg, summary, functionKeys);
}
function addReason(summary, reason) {
    if (summary.intrinsicReasons.some((item) => item.kind === reason.kind && item.target === reason.target))
        return;
    summary.intrinsicReasons.push(reason);
}
function freezeSummary(summary, summaries) {
    const reasons = [...summary.intrinsicReasons];
    for (const target of summary.calls) {
        if (summaries.get(target)?.maySuspend && !reasons.some((item) => item.kind === "call" && item.target === target)) {
            reasons.push({ kind: "call", target });
        }
    }
    const maySuspend = summary.maySuspend || reasons.length > 0;
    return {
        key: summary.key,
        name: summary.name,
        direct: !maySuspend,
        maySuspend,
        calls: [...summary.calls].sort(),
        reasons
    };
}
