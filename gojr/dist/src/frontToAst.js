import { parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
export function frontSourceToAst(source, filename) {
    const parsed = parseFrontSource(source, filename);
    const ast = parsed.file ? frontToProgramAst(parsed.file, parsed.diagnostics, parsed.statements) : undefined;
    return {
        diagnostics: parsed.diagnostics,
        parsed,
        ...(ast ? { ast } : {})
    };
}
export function frontSourceFilesToAst(files) {
    const parsed = parseFrontSourceFiles(files);
    const ast = parsed.files.length > 0 ? frontFilesToProgramAst(parsed.files, parsed.diagnostics, parsed.statements) : undefined;
    return {
        diagnostics: parsed.diagnostics,
        parsed,
        ...(ast ? { ast } : {})
    };
}
export function frontToProgramAst(file, diagnostics = [], statements = []) {
    return frontFilesToProgramAst([file], diagnostics, statements);
}
export function frontFilesToProgramAst(files, diagnostics = [], statements = []) {
    const body = [
        ...files.flatMap((file) => file.declarations.flatMap(declarationToBodyStatement)),
        ...statements.map(statementToAst)
    ];
    const functions = files.flatMap((file) => file.declarations).flatMap((declaration) => declaration.kind === "FuncDecl" ? [functionDeclToAst(declaration)] : []);
    return {
        kind: functions.length > 0 && body.length === 0 ? "function" : "script",
        imports: files.flatMap((file) => file.imports.map(importSpecToAst)),
        diagnostics,
        body,
        functions
    };
}
function declarationToBodyStatement(declaration) {
    if (declaration.kind !== "GenDecl")
        return [];
    const statement = genDeclToStatement(declaration);
    return statement ? [statement] : [];
}
function importSpecToAst(spec) {
    const path = unquote(spec.path.value);
    return {
        path,
        ...(spec.name ? { alias: spec.name.name } : {})
    };
}
function genDeclToStatement(declaration) {
    if (declaration.token === TokenKind.Const) {
        return withSpan({
            kind: "ConstDecl",
            declarations: declaration.specs.flatMap(valueSpecToDeclarations)
        }, declaration.span);
    }
    if (declaration.token === TokenKind.Var) {
        return withSpan({
            kind: "VarDecl",
            declarations: declaration.specs.flatMap(valueSpecToDeclarations)
        }, declaration.span);
    }
    if (declaration.token === TokenKind.Type) {
        return withSpan({
            kind: "TypeDecl",
            declarations: declaration.specs.flatMap(typeSpecToAst)
        }, declaration.span);
    }
    return undefined;
}
function valueSpecToDeclarations(spec) {
    if (spec.kind !== "ValueSpec")
        return [];
    return spec.names.map((name, index) => ({
        name: name.name,
        ...(spec.type ? { type: typeNode(spec.type) } : {}),
        ...(spec.values[index] ? { value: expressionToAst(spec.values[index]) } : {})
    }));
}
function typeSpecToAst(spec) {
    if (spec.kind !== "TypeSpec")
        return [];
    return [{
            name: spec.name.name,
            type: typeNode(spec.type),
            ...structFieldsFromType(spec),
            ...interfaceMethodsFromType(spec)
        }];
}
function structFieldsFromType(spec) {
    if (spec.type.kind !== "StructType")
        return {};
    return {
        structFields: spec.type.fields.fields.flatMap((field) => {
            const tag = field.tag ? unquote(field.tag.value) : undefined;
            if (field.names.length === 0) {
                return [{
                        name: embeddedFieldName(field.type),
                        type: typeNode(field.type),
                        embedded: true,
                        ...(tag !== undefined ? { tag } : {})
                    }];
            }
            return field.names.map((name) => ({
                name: name.name,
                type: typeNode(field.type),
                ...(tag !== undefined ? { tag } : {})
            }));
        })
    };
}
function interfaceMethodsFromType(spec) {
    if (spec.type.kind !== "InterfaceType")
        return {};
    const interfaceMethods = [];
    const interfaceEmbeds = [];
    for (const field of spec.type.methods.fields) {
        const methodType = field.type;
        if (methodType.kind !== "FuncType") {
            interfaceEmbeds.push(typeNode(methodType));
            continue;
        }
        for (const name of field.names) {
            interfaceMethods.push({
                name: name.name,
                signature: signatureToAst(methodType)
            });
        }
    }
    return {
        interfaceMethods,
        ...(interfaceEmbeds.length > 0 ? { interfaceEmbeds } : {})
    };
}
function functionDeclToAst(declaration) {
    return withSpan({
        kind: "FunctionDecl",
        name: declaration.name.name,
        ...(declaration.receiver ? { receiver: receiverToAst(declaration.receiver) } : {}),
        signature: signatureToAst(declaration.type),
        body: declaration.body ? blockToAst(declaration.body) : { kind: "BlockStatement", statements: [] }
    }, declaration.span);
}
function receiverToAst(list) {
    const field = list.fields[0];
    return {
        ...(field?.names[0] ? { name: field.names[0].name } : {}),
        type: field ? typeNode(field.type) : { text: "<missing>" }
    };
}
function signatureToAst(type) {
    return {
        parameters: parametersFromFields(type.params),
        results: type.results ? parametersFromFields(type.results) : []
    };
}
function parametersFromFields(list) {
    return list.fields.flatMap((field) => {
        const fieldType = field.type.kind === "Ellipsis" && field.type.element ? field.type.element : field.type;
        const base = {
            type: typeNode(fieldType),
            variadic: field.type.kind === "Ellipsis"
        };
        if (field.names.length === 0)
            return [base];
        return field.names.map((name) => ({
            name: name.name,
            ...base
        }));
    });
}
function blockToAst(block) {
    return withSpan({
        kind: "BlockStatement",
        statements: block.statements.map(statementToAst)
    }, block.span);
}
function statementToAst(statement) {
    switch (statement.kind) {
        case "DeclStmt": {
            if (statement.decl.kind === "GenDecl")
                return genDeclToStatement(statement.decl) ?? expressionStatement(missingExpression(), statement.span);
            return expressionStatement(missingExpression(), statement.span);
        }
        case "BlockStmt":
            return blockToAst(statement);
        case "LabeledStmt":
            return withSpan({
                kind: "LabeledStatement",
                label: statement.label.name,
                statement: statementToAst(statement.stmt)
            }, statement.span);
        case "ExprStmt":
            return expressionStatement(expressionToAst(statement.expr), statement.span);
        case "AssignStmt":
            if (statement.token === TokenKind.Define) {
                return withSpan({
                    kind: "ShortVarStatement",
                    names: statement.lhs.map(shortVarName),
                    values: statement.rhs.map(expressionToAst)
                }, statement.span);
            }
            return withSpan({
                kind: "AssignStatement",
                targets: statement.lhs.map(expressionToAst),
                operator: assignmentOperator(statement.token),
                values: statement.rhs.map(expressionToAst)
            }, statement.span);
        case "IncDecStmt":
            return withSpan({
                kind: "IncDecStatement",
                target: expressionToAst(statement.expr),
                operator: statement.token === TokenKind.PlusPlus ? "++" : "--"
            }, statement.span);
        case "ReturnStmt":
            return withSpan({
                kind: "ReturnStatement",
                values: statement.results.map(expressionToAst)
            }, statement.span);
        case "BranchStmt":
            return branchToAst(statement);
        case "IfStmt":
            return withSpan({
                kind: "IfStatement",
                ...(statement.init ? { init: statementToAst(statement.init) } : {}),
                condition: expressionToAst(statement.condition),
                thenBlock: blockToAst(statement.body),
                ...(statement.else ? { elseBranch: elseBranchToAst(statement.else) } : {})
            }, statement.span);
        case "ForStmt":
            return withSpan({
                kind: "ForStatement",
                ...(statement.init ? { init: statementToAst(statement.init) } : {}),
                ...(statement.condition ? { condition: expressionToAst(statement.condition) } : {}),
                ...(statement.post ? { post: statementToAst(statement.post) } : {}),
                body: blockToAst(statement.body)
            }, statement.span);
        case "RangeStmt":
            return rangeStmtToAst(statement);
        case "SwitchStmt":
            return switchStmtToAst(statement);
        case "TypeSwitchStmt":
            return typeSwitchStmtToAst(statement);
        case "SelectStmt":
            return selectStmtToAst(statement);
        case "DeferStmt":
            return withSpan({
                kind: "DeferStatement",
                expression: expressionToAst(statement.call)
            }, statement.span);
        case "GoStmt":
            return withSpan({
                kind: "GoStatement",
                call: expressionToAst(statement.call)
            }, statement.span);
        case "SendStmt":
            return withSpan({
                kind: "SendStatement",
                channel: expressionToAst(statement.channel),
                value: expressionToAst(statement.value)
            }, statement.span);
        default:
            return expressionStatement(missingExpression(), statement.span);
    }
}
function elseBranchToAst(statement) {
    if (statement.kind === "IfStmt")
        return statementToAst(statement);
    if (statement.kind === "BlockStmt")
        return blockToAst(statement);
    return { kind: "BlockStatement", statements: [statementToAst(statement)] };
}
function branchToAst(statement) {
    const branch = statement.token === TokenKind.Break
        ? "break"
        : statement.token === TokenKind.Continue
            ? "continue"
            : statement.token === TokenKind.Goto
                ? "goto"
                : "fallthrough";
    return withSpan({
        kind: "BranchStatement",
        branch,
        ...(statement.label ? { label: statement.label.name } : {})
    }, statement.span);
}
function rangeStmtToAst(statement) {
    return withSpan({
        kind: "ForStatement",
        range: {
            ...(statement.key?.kind === "Ident" ? { keyName: statement.key.name } : {}),
            ...(statement.value?.kind === "Ident" ? { valueName: statement.value.name } : {}),
            define: statement.token === TokenKind.Define,
            source: expressionToAst(statement.source)
        },
        body: blockToAst(statement.body)
    }, statement.span);
}
function switchStmtToAst(statement) {
    return withSpan({
        kind: "SwitchStatement",
        ...(statement.init ? { init: statementToAst(statement.init) } : {}),
        ...(statement.tag ? { expression: expressionToAst(statement.tag) } : {}),
        clauses: statement.body.map((clause) => caseClauseToAst(clause, false))
    }, statement.span);
}
function typeSwitchStmtToAst(statement) {
    return withSpan({
        kind: "SwitchStatement",
        ...(statement.init ? { init: statementToAst(statement.init) } : {}),
        typeSwitch: typeSwitchGuardToAst(statement.assign),
        clauses: statement.body.map((clause) => caseClauseToAst(clause, true))
    }, statement.span);
}
function selectStmtToAst(statement) {
    return withSpan({
        kind: "SelectStatement",
        clauses: statement.body.map(commClauseToAst)
    }, statement.span);
}
function commClauseToAst(clause) {
    return {
        kind: "CommClause",
        ...(clause.comm ? { comm: statementToAst(clause.comm) } : {}),
        default: clause.default,
        statements: clause.body.map(statementToAst)
    };
}
function typeSwitchGuardToAst(statement) {
    if (statement.kind === "AssignStmt") {
        const assertion = statement.rhs[0];
        return {
            ...(statement.lhs[0]?.kind === "Ident" ? { name: statement.lhs[0].name } : {}),
            define: statement.token === TokenKind.Define,
            expression: assertion?.kind === "TypeAssertExpr" ? expressionToAst(assertion.object) : missingExpression()
        };
    }
    if (statement.kind === "ExprStmt" && statement.expr.kind === "TypeAssertExpr") {
        return {
            define: false,
            expression: expressionToAst(statement.expr.object)
        };
    }
    return {
        define: false,
        expression: missingExpression()
    };
}
function caseClauseToAst(clause, typeSwitch) {
    return {
        kind: "SwitchClause",
        values: typeSwitch ? [] : clause.list.map(expressionToAst),
        ...(typeSwitch ? { typeValues: clause.list.map(typeNode) } : {}),
        default: clause.default,
        statements: clause.body.map(statementToAst)
    };
}
function expressionStatement(expression, span) {
    return withSpan({
        kind: "ExpressionStatement",
        expression
    }, span);
}
function expressionToAst(expr) {
    switch (expr.kind) {
        case "Ident":
            if (expr.name === "true" || expr.name === "false")
                return literal(expr.name === "true", "bool", expr.name, expr.span);
            if (expr.name === "nil")
                return literal(null, "nil", "nil", expr.span);
            return withSpan({ kind: "Identifier", name: expr.name }, expr.span);
        case "BasicLit":
            return basicLitToAst(expr);
        case "FuncLit":
            return withSpan({
                kind: "FunctionLiteralExpression",
                signature: signatureToAst(expr.type),
                body: blockToAst(expr.body)
            }, expr.span);
        case "CompositeLit":
            return compositeLitToAst(expr);
        case "ParenExpr":
            return expressionToAst(expr.expr);
        case "SelectorExpr":
            return withSpan({
                kind: "SelectorExpression",
                object: expressionToAst(expr.object),
                field: expr.selector.name
            }, expr.span);
        case "CellRefExpr":
            return cellRefToSelector(expr);
        case "RangeRefExpr":
            return rangeRefToAst(expr);
        case "IndexExpr":
            return withSpan({
                kind: "IndexExpression",
                object: expressionToAst(expr.object),
                index: expressionToAst(expr.index)
            }, expr.span);
        case "SliceExpr":
            return withSpan({
                kind: "SliceExpression",
                object: expressionToAst(expr.object),
                ...(expr.low ? { start: expressionToAst(expr.low) } : {}),
                ...(expr.high ? { end: expressionToAst(expr.high) } : {}),
                ...(expr.max ? { max: expressionToAst(expr.max) } : {})
            }, expr.span);
        case "TypeAssertExpr":
            return withSpan({
                kind: "TypeAssertionExpression",
                expression: expressionToAst(expr.object),
                type: expr.type ? typeNode(expr.type) : { text: "type" }
            }, expr.span);
        case "CallExpr":
            return withSpan({
                kind: "CallExpression",
                callee: expressionToAst(expr.fun),
                args: expr.args.map(expressionToAst),
                spreadLast: expr.ellipsis
            }, expr.span);
        case "StarExpr":
            return withSpan({
                kind: "UnaryExpression",
                operator: "*",
                operand: expressionToAst(expr.expr)
            }, expr.span);
        case "UnaryExpr":
            return withSpan({
                kind: "UnaryExpression",
                operator: unaryOperator(expr.op),
                operand: expressionToAst(expr.expr)
            }, expr.span);
        case "BinaryExpr":
            return withSpan({
                kind: "BinaryExpression",
                operator: binaryOperator(expr.op),
                left: expressionToAst(expr.left),
                right: expressionToAst(expr.right)
            }, expr.span);
        case "ArrayType":
        case "MapType":
        case "StructType":
        case "InterfaceType":
        case "FuncType":
            return withSpan({
                kind: "TypeExpression",
                type: typeNode(expr)
            }, expr.span);
        case "ChanType":
            return withSpan({
                kind: "TypeExpression",
                type: typeNode(expr)
            }, expr.span);
        default:
            return missingExpression(expr.span);
    }
}
function basicLitToAst(expr) {
    if (expr.token === TokenKind.IntLiteral)
        return literal(parseGoIntLiteral(expr.value), "int", expr.value, expr.span);
    if (expr.token === TokenKind.FloatLiteral)
        return literal(parseGoFloatLiteral(expr.value), "float", expr.value, expr.span);
    if (expr.token === TokenKind.ImagLiteral) {
        const raw = expr.value.slice(0, -1);
        const imag = /[.eEpP]/.test(raw) ? parseGoFloatLiteral(raw) : Number(parseGoIntLiteral(raw));
        return literal({ real: 0, imag }, "imag", expr.value, expr.span);
    }
    if (expr.token === TokenKind.RuneLiteral)
        return literal(parseGoRuneLiteral(expr.value), "rune", expr.value, expr.span);
    return literal(unquote(expr.value), "string", expr.value, expr.span);
}
function literal(value, literalKind, raw, span) {
    return withSpan({
        kind: "Literal",
        literalKind,
        value,
        raw
    }, span);
}
function compositeLitToAst(expr) {
    const type = expr.type;
    if (type?.kind === "MapType") {
        return withSpan({
            kind: "MapLiteralExpression",
            keyType: typeNode(type.key),
            valueType: typeNode(type.value),
            entries: expr.elements.flatMap(mapEntryToAst)
        }, expr.span);
    }
    if (type?.kind === "ArrayType") {
        return withSpan({
            kind: "ArrayLiteralExpression",
            type: typeNode(type),
            elements: expr.elements.map(elementValueToAst)
        }, expr.span);
    }
    return withSpan({
        kind: "StructLiteralExpression",
        typeName: type ? typeText(type) : "<missing>",
        fields: expr.elements.map(structFieldToAst)
    }, expr.span);
}
function mapEntryToAst(expr) {
    if (expr.kind !== "KeyValueExpr")
        return [];
    return [{ key: expressionToAst(expr.key), value: expressionToAst(expr.value) }];
}
function structFieldToAst(expr) {
    if (expr.kind === "KeyValueExpr") {
        return {
            ...(expr.key.kind === "Ident" ? { name: expr.key.name } : {}),
            value: expressionToAst(expr.value)
        };
    }
    return { value: expressionToAst(expr) };
}
function elementValueToAst(expr) {
    return expr.kind === "KeyValueExpr" ? expressionToAst(expr.value) : expressionToAst(expr);
}
function cellRefToSelector(expr) {
    return withSpan({
        kind: "SelectorExpression",
        object: { kind: "Identifier", name: expr.namespace.name },
        field: expr.address.raw
    }, expr.span);
}
function rangeRefToAst(expr) {
    return withSpan({
        kind: "SpreadsheetRangeExpression",
        start: withSpan({
            kind: "SelectorExpression",
            object: { kind: "Identifier", name: expr.namespace.name },
            field: expr.start.raw
        }, expr.span),
        endCell: expr.end.raw
    }, expr.span);
}
function shortVarName(expr) {
    return expr.kind === "Ident" ? expr.name : "<invalid>";
}
function typeNode(expr) {
    return withSpan({ text: typeText(expr) }, expr.span);
}
function typeText(expr) {
    switch (expr.kind) {
        case "Ident":
            return expr.name;
        case "SelectorExpr":
            return `${typeText(expr.object)}.${expr.selector.name}`;
        case "IndexExpr":
            return `${typeText(expr.object)}[${typeText(expr.index)}]`;
        case "StarExpr":
            return `*${typeText(expr.expr)}`;
        case "UnaryExpr":
            if (expr.op === TokenKind.Tilde)
                return `~${typeText(expr.expr)}`;
            return `${unaryOperator(expr.op)}${typeText(expr.expr)}`;
        case "BinaryExpr":
            if (expr.op === TokenKind.Or)
                return `${typeText(expr.left)} | ${typeText(expr.right)}`;
            return `${typeText(expr.left)} ${binaryOperator(expr.op)} ${typeText(expr.right)}`;
        case "ArrayType":
            return `${arrayLengthText(expr)}${typeText(expr.element)}`;
        case "MapType":
            return `map[${typeText(expr.key)}]${typeText(expr.value)}`;
        case "ChanType":
            if (expr.direction === "send")
                return `chan<- ${typeText(expr.value)}`;
            if (expr.direction === "receive")
                return `<-chan ${typeText(expr.value)}`;
            return `chan ${typeText(expr.value)}`;
        case "StructType":
            return `struct{${fieldsText(expr.fields)}}`;
        case "InterfaceType":
            return `interface{${interfaceText(expr.methods)}}`;
        case "FuncType":
            return `func(${paramsText(expr.params)})${resultsText(expr.results)}`;
        case "Ellipsis":
            return `...${expr.element ? typeText(expr.element) : ""}`;
        case "BasicLit":
            return expr.value;
        default:
            return "<missing>";
    }
}
function arrayLengthText(expr) {
    if (expr.inferredLength)
        return "[...]";
    return expr.length ? `[${expressionText(expr.length)}]` : "[]";
}
function fieldsText(fields) {
    return fields.fields.map((field) => {
        const names = field.names.map((name) => name.name).join(", ");
        const tag = field.tag ? ` ${field.tag.value}` : "";
        return `${names ? `${names} ` : ""}${typeText(field.type)}${tag}`;
    }).join("; ");
}
function interfaceText(fields) {
    return fields.fields.map((field) => {
        const names = field.names.map((name) => name.name).join(", ");
        if (!names && field.type.kind !== "FuncType")
            return typeText(field.type);
        return `${names}${field.type.kind === "FuncType" ? `(${paramsText(field.type.params)})${resultsText(field.type.results)}` : ` ${typeText(field.type)}`}`;
    }).join("; ");
}
function paramsText(fields) {
    return fields.fields.map(fieldText).join(", ");
}
function fieldText(field) {
    const names = field.names.map((name) => name.name).join(", ");
    return `${names ? `${names} ` : ""}${typeText(field.type)}`;
}
function resultsText(results) {
    if (!results || results.fields.length === 0)
        return "";
    if (results.fields.length === 1 && results.fields[0]?.names.length === 0)
        return ` ${typeText(results.fields[0].type)}`;
    return ` (${paramsText(results)})`;
}
function expressionText(expr) {
    if (expr.kind === "BasicLit")
        return expr.value;
    if (expr.kind === "Ident")
        return expr.name;
    return typeText(expr);
}
function unaryOperator(kind) {
    if (kind === TokenKind.Arrow)
        return "<-";
    if (kind === TokenKind.Minus)
        return "-";
    if (kind === TokenKind.Bang)
        return "!";
    if (kind === TokenKind.Caret)
        return "^";
    if (kind === TokenKind.Amp)
        return "&";
    if (kind === TokenKind.Star)
        return "*";
    return "+";
}
function binaryOperator(kind) {
    switch (kind) {
        case TokenKind.OrOr: return "||";
        case TokenKind.AndAnd: return "&&";
        case TokenKind.Equal: return "==";
        case TokenKind.NotEqual: return "!=";
        case TokenKind.Less: return "<";
        case TokenKind.LessEqual: return "<=";
        case TokenKind.Greater: return ">";
        case TokenKind.GreaterEqual: return ">=";
        case TokenKind.Minus: return "-";
        case TokenKind.Or: return "|";
        case TokenKind.Caret: return "^";
        case TokenKind.Star: return "*";
        case TokenKind.Slash: return "/";
        case TokenKind.Percent: return "%";
        case TokenKind.Shl: return "<<";
        case TokenKind.Shr: return ">>";
        case TokenKind.Amp: return "&";
        case TokenKind.BitClear: return "&^";
        default: return "+";
    }
}
function assignmentOperator(kind) {
    switch (kind) {
        case TokenKind.PlusAssign: return "+=";
        case TokenKind.MinusAssign: return "-=";
        case TokenKind.StarAssign: return "*=";
        case TokenKind.SlashAssign: return "/=";
        case TokenKind.PercentAssign: return "%=";
        case TokenKind.AmpAssign: return "&=";
        case TokenKind.OrAssign: return "|=";
        case TokenKind.CaretAssign: return "^=";
        case TokenKind.BitClearAssign: return "&^=";
        case TokenKind.ShlAssign: return "<<=";
        case TokenKind.ShrAssign: return ">>=";
        default: return "=";
    }
}
function parseGoIntLiteral(value) {
    const text = value.replace(/_/g, "");
    if (/^0[0-7]+$/.test(text))
        return BigInt(`0o${text.slice(1)}`);
    return BigInt(text);
}
function parseGoFloatLiteral(value) {
    const text = value.replace(/_/g, "");
    const hex = /^0[xX]([0-9a-fA-F]*)(?:\.([0-9a-fA-F]*))?[pP]([+-]?[0-9]+)$/.exec(text);
    if (!hex)
        return Number(text);
    const whole = hex[1] || "0";
    const frac = hex[2] || "";
    const exponent = Number(hex[3]);
    const wholeValue = Number.parseInt(whole, 16);
    let fracValue = 0;
    for (let index = 0; index < frac.length; index += 1) {
        fracValue += Number.parseInt(frac[index] ?? "0", 16) / 16 ** (index + 1);
    }
    return (wholeValue + fracValue) * 2 ** exponent;
}
function parseGoRuneLiteral(value) {
    const body = value.slice(1, -1);
    const decoded = decodeGoEscaped(body);
    return BigInt([...decoded][0]?.codePointAt(0) ?? 0);
}
function embeddedFieldName(expr) {
    if (expr.kind === "Ident")
        return expr.name;
    if (expr.kind === "SelectorExpr")
        return expr.selector.name;
    if (expr.kind === "StarExpr")
        return embeddedFieldName(expr.expr);
    return "";
}
function missingExpression(span) {
    return withSpan({ kind: "Identifier", name: "<missing>" }, span);
}
function withSpan(node, span) {
    return span ? { ...node, span } : node;
}
function unquote(value) {
    if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
        return value.slice(1, -1).replace(/\r/g, "");
    }
    if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
        return decodeGoEscaped(value.slice(1, -1));
    }
    return value;
}
function decodeGoEscaped(value) {
    let decoded = "";
    for (let index = 0; index < value.length; index += 1) {
        const char = value[index] ?? "";
        if (char !== "\\") {
            decoded += char;
            continue;
        }
        const next = value[index + 1] ?? "";
        index += 1;
        switch (next) {
            case "a":
                decoded += "\x07";
                break;
            case "b":
                decoded += "\b";
                break;
            case "f":
                decoded += "\f";
                break;
            case "n":
                decoded += "\n";
                break;
            case "r":
                decoded += "\r";
                break;
            case "t":
                decoded += "\t";
                break;
            case "v":
                decoded += "\x0b";
                break;
            case "\\":
            case "\"":
            case "'":
                decoded += next;
                break;
            case "x": {
                decoded += codePointFromEscape(value.slice(index + 1, index + 3), 16);
                index += 2;
                break;
            }
            case "u": {
                decoded += codePointFromEscape(value.slice(index + 1, index + 5), 16);
                index += 4;
                break;
            }
            case "U": {
                decoded += codePointFromEscape(value.slice(index + 1, index + 9), 16);
                index += 8;
                break;
            }
            default:
                if (/^[0-7]$/.test(next)) {
                    const digits = next + value.slice(index + 1, index + 3);
                    decoded += codePointFromEscape(digits, 8);
                    index += 2;
                }
                else {
                    decoded += next;
                }
                break;
        }
    }
    return decoded;
}
function codePointFromEscape(digits, radix) {
    const value = Number.parseInt(digits, radix);
    if (!Number.isFinite(value))
        return "";
    try {
        return String.fromCodePoint(value);
    }
    catch {
        return "";
    }
}
