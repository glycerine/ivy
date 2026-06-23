import { REPL_FILENAME, diagnosticFilename } from "../diagnostics.js";
import { parseCellAddress, ident } from "./ast.js";
import { scanSource } from "./scanner.js";
import { isAssignmentToken, isIdentifierLike, TokenKind } from "./token.js";
export function parseFrontSource(source, filename) {
    const scanned = scanSource(source, filename);
    const parser = new FrontParser(scanned.tokens, scanned.diagnostics, filename);
    return parser.parseFile();
}
export function parseFrontSourceFiles(files) {
    const results = files.map((file) => parseFrontSource(file.source, file.filename));
    return {
        files: results.flatMap((result) => result.file ? [result.file] : []),
        statements: results.flatMap((result) => result.statements),
        diagnostics: results.flatMap((result) => result.diagnostics),
        results
    };
}
class FrontParser {
    tokens;
    filename;
    index = 0;
    allowBareIdentifierComposite = true;
    diagnostics;
    constructor(tokens, diagnostics, filename) {
        this.tokens = tokens;
        this.filename = filename;
        this.diagnostics = [...diagnostics];
    }
    parseFile() {
        let name;
        if (this.match(TokenKind.Package)) {
            name = this.parseIdent("expected package name");
            this.consumeSemi();
        }
        const declarations = [];
        const imports = [];
        const statements = [];
        while (!this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.EOF))
                break;
            if (this.startsDecl()) {
                const declaration = this.parseDecl();
                declarations.push(declaration);
                if (declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
                    imports.push(...declaration.specs.filter((spec) => spec.kind === "ImportSpec"));
                }
            }
            else {
                statements.push(this.parseStatement());
            }
            this.consumeSemi();
        }
        return {
            tokens: this.tokens,
            diagnostics: this.diagnostics,
            statements,
            file: {
                kind: "File",
                ...(name ? { name } : {}),
                declarations,
                imports,
                unresolved: [],
                comments: [],
                span: mergeSpans(name?.span, declarations[declarations.length - 1]?.span)
            }
        };
    }
    parseDecl() {
        if (this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var))
            return this.parseGenDecl();
        if (this.at(TokenKind.Func))
            return this.parseFuncDecl();
        const token = this.peek();
        this.error(`expected declaration, found ${token.lexeme || token.kind}`, token.span);
        this.advance();
        return { kind: "BadDecl", span: token.span };
    }
    startsDecl() {
        return this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func);
    }
    parseGenDecl() {
        const start = this.advance();
        const token = start.kind;
        const specs = [];
        let grouped = false;
        if (this.match(TokenKind.LParen)) {
            grouped = true;
            while (!this.at(TokenKind.RParen) && !this.at(TokenKind.EOF)) {
                this.skipSemis();
                if (this.at(TokenKind.RParen))
                    break;
                specs.push(this.parseSpec(token));
                this.consumeSemi();
            }
            this.expect(TokenKind.RParen, "expected ')' after declaration group");
        }
        else {
            specs.push(this.parseSpec(token));
        }
        return {
            kind: "GenDecl",
            token,
            specs,
            grouped,
            span: mergeSpans(start.span, specs[specs.length - 1]?.span)
        };
    }
    parseSpec(token) {
        if (token === TokenKind.Import)
            return this.parseImportSpec();
        if (token === TokenKind.Type)
            return this.parseTypeSpec();
        return this.parseValueSpec();
    }
    parseImportSpec() {
        const start = this.peek();
        let name;
        if (this.at(TokenKind.Identifier) && this.peek(1).kind === TokenKind.StringLiteral) {
            name = this.parseIdent("expected import alias");
        }
        else if (this.at(TokenKind.Dot) && this.peek(1).kind === TokenKind.StringLiteral) {
            const dot = this.advance();
            name = ident(".", dot.span);
        }
        const path = this.expectBasicLit(TokenKind.StringLiteral, "expected import path string");
        return {
            kind: "ImportSpec",
            ...(name ? { name } : {}),
            path,
            span: mergeSpans(start.span, path.span)
        };
    }
    parseTypeSpec() {
        const name = this.parseIdent("expected type name");
        const alias = this.match(TokenKind.Assign);
        const type = this.parseType();
        return {
            kind: "TypeSpec",
            name,
            type,
            alias,
            span: mergeSpans(name.span, type.span)
        };
    }
    parseValueSpec() {
        const names = this.parseIdentList();
        let type;
        let values = [];
        if (!this.atAny(TokenKind.Assign, TokenKind.Semicolon, TokenKind.RParen, TokenKind.EOF)) {
            type = this.parseType();
        }
        if (this.match(TokenKind.Assign)) {
            values = this.parseExpressionList();
        }
        return {
            kind: "ValueSpec",
            names,
            ...(type ? { type } : {}),
            values,
            span: mergeSpans(names[0]?.span, values[values.length - 1]?.span ?? type?.span ?? names[names.length - 1]?.span)
        };
    }
    parseFuncDecl() {
        const start = this.expect(TokenKind.Func, "expected func");
        let receiver;
        if (this.at(TokenKind.LParen) && this.looksLikeReceiver()) {
            receiver = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        }
        const name = this.parseIdent("expected function name");
        const type = this.parseSignature(start.span);
        const body = this.at(TokenKind.LBrace) ? this.parseBlock() : undefined;
        return {
            kind: "FuncDecl",
            ...(receiver ? { receiver } : {}),
            name,
            type,
            ...(body ? { body } : {}),
            span: mergeSpans(start.span, body?.span ?? type.span)
        };
    }
    parseStatement() {
        this.skipSemis();
        if (this.at(TokenKind.LBrace))
            return this.parseBlock();
        if (this.atAny(TokenKind.Const, TokenKind.Type, TokenKind.Var)) {
            return { kind: "DeclStmt", decl: this.parseGenDecl() };
        }
        if (this.match(TokenKind.Return)) {
            const start = this.previous();
            const results = this.atAny(TokenKind.Semicolon, TokenKind.RBrace, TokenKind.EOF) ? [] : this.parseExpressionList();
            return { kind: "ReturnStmt", results, span: mergeSpans(start.span, results[results.length - 1]?.span ?? start.span) };
        }
        if (this.atAny(TokenKind.Break, TokenKind.Continue, TokenKind.Goto, TokenKind.Fallthrough)) {
            const token = this.advance();
            const label = this.at(TokenKind.Identifier) ? this.parseIdent("expected label") : undefined;
            return {
                kind: "BranchStmt",
                token: token.kind,
                ...(label ? { label } : {}),
                span: mergeSpans(token.span, label?.span)
            };
        }
        if (this.match(TokenKind.Defer)) {
            const start = this.previous();
            const expression = this.parseExpression();
            const call = expression.kind === "CallExpr"
                ? expression
                : { kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) };
            return { kind: "DeferStmt", call, span: mergeSpans(start.span, call.span) };
        }
        if (this.at(TokenKind.Go))
            return this.parseGoStmt();
        if (this.at(TokenKind.Select))
            return this.parseSelectStmt();
        if (this.at(TokenKind.For))
            return this.parseForStmt();
        if (this.at(TokenKind.If))
            return this.parseIfStmt();
        if (this.at(TokenKind.Switch))
            return this.parseSwitchStmt();
        if (isIdentifierLike(this.peek().kind) && this.peek(1).kind === TokenKind.Colon) {
            const label = this.parseIdent("expected label");
            this.expect(TokenKind.Colon, "expected ':' after label");
            return {
                kind: "LabeledStmt",
                label,
                stmt: this.parseStatement(),
                span: mergeSpans(label.span, label.span)
            };
        }
        return this.parseSimpleStmt();
    }
    parseSimpleStmt() {
        const lhs = this.parseExpressionList();
        if (this.match(TokenKind.Arrow)) {
            if (lhs.length !== 1) {
                this.error("send statement expects one channel expression", lhs[1]?.span ?? lhs[0]?.span);
            }
            const value = this.parseExpression();
            return {
                kind: "SendStmt",
                channel: lhs[0] ?? badExpr(value.span),
                value,
                span: mergeSpans(lhs[0]?.span, value.span)
            };
        }
        if (isAssignmentToken(this.peek().kind)) {
            const token = this.advance();
            const rhs = this.parseExpressionList();
            return {
                kind: "AssignStmt",
                lhs,
                token: token.kind,
                rhs,
                span: mergeSpans(lhs[0]?.span, rhs[rhs.length - 1]?.span)
            };
        }
        if (this.at(TokenKind.PlusPlus) || this.at(TokenKind.MinusMinus)) {
            const token = this.advance();
            return {
                kind: "IncDecStmt",
                expr: lhs[0] ?? badExpr(token.span),
                token: token.kind,
                span: mergeSpans(lhs[0]?.span, token.span)
            };
        }
        return { kind: "ExprStmt", expr: lhs[0] ?? badExpr(this.peek().span), span: mergeSpans(lhs[0]?.span, lhs[0]?.span) };
    }
    parseGoStmt() {
        const start = this.expect(TokenKind.Go, "expected go");
        const expression = this.parseExpression();
        if (expression.kind !== "CallExpr") {
            this.error("go statement requires function call", expression.span);
            return {
                kind: "GoStmt",
                call: { kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) },
                span: mergeSpans(start.span, expression.span)
            };
        }
        return { kind: "GoStmt", call: expression, span: mergeSpans(start.span, expression.span) };
    }
    parseIfStmt() {
        const start = this.expect(TokenKind.If, "expected if");
        const first = this.withBareIdentifierComposites(false, () => this.parseSimpleStmt());
        let init;
        let condition;
        if (this.match(TokenKind.Semicolon)) {
            init = first;
            condition = this.parseExpression();
        }
        else {
            condition = first.kind === "ExprStmt" ? first.expr : badExpr(first.span);
        }
        const body = this.parseBlock();
        let elseStmt;
        if (this.match(TokenKind.Else)) {
            elseStmt = this.at(TokenKind.If) ? this.parseIfStmt() : this.parseBlock();
        }
        return {
            kind: "IfStmt",
            ...(init ? { init } : {}),
            condition,
            body,
            ...(elseStmt ? { else: elseStmt } : {}),
            span: mergeSpans(start.span, elseStmt?.span ?? body.span)
        };
    }
    parseForStmt() {
        const start = this.expect(TokenKind.For, "expected for");
        if (this.at(TokenKind.LBrace)) {
            const body = this.parseBlock();
            return { kind: "ForStmt", body, span: mergeSpans(start.span, body.span) };
        }
        if (this.looksLikeRangeClause()) {
            let key;
            let value;
            let token = TokenKind.Assign;
            if (!this.at(TokenKind.Range)) {
                const lhs = this.parseExpressionList();
                const assign = this.expectAny([TokenKind.Assign, TokenKind.Define], "expected '=' or ':=' before range");
                token = assign.kind === TokenKind.Define ? TokenKind.Define : TokenKind.Assign;
                this.expect(TokenKind.Range, "expected range");
                key = lhs[0];
                value = lhs[1];
            }
            else {
                this.advance();
            }
            const source = this.withBareIdentifierComposites(false, () => this.parseExpression());
            const body = this.parseBlock();
            return {
                kind: "RangeStmt",
                ...(key ? { key } : {}),
                ...(value ? { value } : {}),
                token,
                source,
                body,
                span: mergeSpans(start.span, body.span)
            };
        }
        const first = this.parseSimpleStmt();
        if (this.match(TokenKind.Semicolon)) {
            const condition = this.at(TokenKind.Semicolon) ? undefined : this.parseExpression();
            this.expect(TokenKind.Semicolon, "expected ';' in for clause");
            const post = this.at(TokenKind.LBrace) ? undefined : this.parseSimpleStmt();
            const body = this.parseBlock();
            return {
                kind: "ForStmt",
                init: first,
                ...(condition ? { condition } : {}),
                ...(post ? { post } : {}),
                body,
                span: mergeSpans(start.span, body.span)
            };
        }
        const body = this.parseBlock();
        return {
            kind: "ForStmt",
            condition: first.kind === "ExprStmt" ? first.expr : badExpr(first.span),
            body,
            span: mergeSpans(start.span, body.span)
        };
    }
    parseSwitchStmt() {
        const start = this.expect(TokenKind.Switch, "expected switch");
        let init;
        let tag;
        let assign;
        let typeSwitch = false;
        if (!this.at(TokenKind.LBrace)) {
            const first = this.withBareIdentifierComposites(false, () => this.parseSimpleStmt());
            if (this.match(TokenKind.Semicolon)) {
                init = first;
                if (!this.at(TokenKind.LBrace)) {
                    const second = this.withBareIdentifierComposites(false, () => this.parseSimpleStmt());
                    if (this.isTypeSwitchGuard(second)) {
                        assign = second;
                        typeSwitch = true;
                    }
                    else if (second.kind === "ExprStmt") {
                        tag = second.expr;
                    }
                    else {
                        this.error("expected switch expression or type switch guard after ';'", second.span);
                    }
                }
            }
            else if (this.isTypeSwitchGuard(first)) {
                assign = first;
                typeSwitch = true;
            }
            else if (first.kind === "ExprStmt") {
                tag = first.expr;
            }
            else {
                this.error("expected ';' after switch init statement", first.span);
                init = first;
            }
        }
        const { clauses, span } = this.parseSwitchBody();
        if (typeSwitch) {
            return {
                kind: "TypeSwitchStmt",
                ...(init ? { init } : {}),
                assign: assign ?? { kind: "BadStmt", span: start.span },
                body: clauses,
                span: mergeSpans(start.span, span)
            };
        }
        return {
            kind: "SwitchStmt",
            ...(init ? { init } : {}),
            ...(tag ? { tag } : {}),
            body: clauses,
            span: mergeSpans(start.span, span)
        };
    }
    parseSwitchBody() {
        const start = this.expect(TokenKind.LBrace, "expected '{' after switch");
        const clauses = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
                clauses.push(this.parseCaseClause());
            }
            else {
                const token = this.peek();
                this.error("expected case or default in switch body", token.span);
                this.advance();
            }
        }
        const end = this.expect(TokenKind.RBrace, "expected '}' after switch body");
        return { clauses, span: mergeSpans(start.span, end.span) };
    }
    parseSelectStmt() {
        const start = this.expect(TokenKind.Select, "expected select");
        const open = this.expect(TokenKind.LBrace, "expected '{' after select");
        const clauses = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
                clauses.push(this.parseCommClause());
            }
            else {
                const token = this.peek();
                this.error("expected case or default in select body", token.span);
                this.advance();
            }
        }
        const close = this.expect(TokenKind.RBrace, "expected '}' after select body");
        return {
            kind: "SelectStmt",
            body: clauses,
            span: mergeSpans(start.span, close.span ?? open.span)
        };
    }
    parseCommClause() {
        const start = this.advance();
        const isDefault = start.kind === TokenKind.Default;
        const comm = isDefault || this.at(TokenKind.Colon) ? undefined : this.parseSimpleStmt();
        this.expect(TokenKind.Colon, "expected ':' after select case");
        const body = [];
        while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
            this.skipSemis();
            if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF))
                break;
            body.push(this.parseStatement());
            this.consumeSemi();
        }
        return {
            kind: "CommClause",
            ...(comm ? { comm } : {}),
            body,
            default: isDefault,
            span: mergeSpans(start.span, body[body.length - 1]?.span ?? comm?.span ?? start.span)
        };
    }
    parseCaseClause() {
        const start = this.advance();
        const isDefault = start.kind === TokenKind.Default;
        const list = isDefault ? [] : this.parseExpressionList();
        this.expect(TokenKind.Colon, "expected ':' after switch case");
        const body = [];
        while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
            this.skipSemis();
            if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF))
                break;
            body.push(this.parseStatement());
            this.consumeSemi();
        }
        return {
            kind: "CaseClause",
            list,
            body,
            default: isDefault,
            span: mergeSpans(start.span, body[body.length - 1]?.span ?? list[list.length - 1]?.span ?? start.span)
        };
    }
    isTypeSwitchGuard(statement) {
        if (statement.kind === "ExprStmt") {
            return statement.expr.kind === "TypeAssertExpr" && statement.expr.typeSwitch;
        }
        if (statement.kind !== "AssignStmt" || statement.rhs.length !== 1)
            return false;
        const rhs = statement.rhs[0];
        return rhs?.kind === "TypeAssertExpr" && rhs.typeSwitch;
    }
    parseBlock() {
        const start = this.expect(TokenKind.LBrace, "expected '{'");
        const statements = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            statements.push(this.parseStatement());
            this.consumeSemi();
        }
        const end = this.expect(TokenKind.RBrace, "expected '}'");
        return {
            kind: "BlockStmt",
            statements,
            span: mergeSpans(start.span, end.span)
        };
    }
    parseExpressionList() {
        const expressions = [this.parseExpression()];
        while (this.match(TokenKind.Comma)) {
            expressions.push(this.parseExpression());
        }
        return expressions;
    }
    parseExpression(minPrecedence = 1) {
        let left = this.parseUnary();
        while (true) {
            const precedence = binaryPrecedence(this.peek().kind);
            if (precedence < minPrecedence)
                break;
            const operator = this.advance();
            const right = this.parseExpression(precedence + 1);
            left = {
                kind: "BinaryExpr",
                left,
                op: operator.kind,
                right,
                span: mergeSpans(left.span, right.span)
            };
        }
        return left;
    }
    parseUnary() {
        if (this.atAny(TokenKind.Plus, TokenKind.Minus, TokenKind.Bang, TokenKind.Caret, TokenKind.Amp, TokenKind.Arrow)) {
            const operator = this.advance();
            const expr = this.parseUnary();
            return {
                kind: "UnaryExpr",
                op: operator.kind,
                expr,
                span: mergeSpans(operator.span, expr.span)
            };
        }
        if (this.match(TokenKind.Star)) {
            const start = this.previous();
            const expr = this.parseUnary();
            return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
        }
        return this.parsePrimary();
    }
    parsePrimary() {
        let expression = this.parseOperand();
        while (true) {
            if (this.match(TokenKind.Dot)) {
                const dot = this.previous();
                if (this.match(TokenKind.LParen)) {
                    if (this.match(TokenKind.Type)) {
                        const close = this.expect(TokenKind.RParen, "expected ')' after type switch guard");
                        expression = {
                            kind: "TypeAssertExpr",
                            object: expression,
                            typeSwitch: true,
                            span: mergeSpans(expression.span, close.span)
                        };
                    }
                    else {
                        const type = this.parseType();
                        const close = this.expect(TokenKind.RParen, "expected ')' after type assertion");
                        expression = {
                            kind: "TypeAssertExpr",
                            object: expression,
                            type,
                            typeSwitch: false,
                            span: mergeSpans(expression.span, close.span)
                        };
                    }
                    continue;
                }
                if (this.at(TokenKind.CellAddress) && expression.kind === "Ident") {
                    const cellToken = this.advance();
                    const start = parseCellAddress(cellToken.lexeme);
                    if (!start) {
                        this.error("malformed spreadsheet cell address", cellToken.span);
                        expression = badExpr(cellToken.span);
                        continue;
                    }
                    if (this.match(TokenKind.Colon)) {
                        const endToken = this.expect(TokenKind.CellAddress, "expected cell address after ':'");
                        const end = parseCellAddress(endToken.lexeme);
                        if (!end) {
                            this.error("malformed spreadsheet range end", endToken.span);
                            expression = badExpr(endToken.span);
                        }
                        else {
                            expression = {
                                kind: "RangeRefExpr",
                                namespace: expression,
                                start,
                                end,
                                span: mergeSpans(expression.span, endToken.span)
                            };
                        }
                    }
                    else {
                        expression = {
                            kind: "CellRefExpr",
                            namespace: expression,
                            address: start,
                            span: mergeSpans(expression.span, cellToken.span)
                        };
                    }
                    continue;
                }
                const selector = this.parseIdent("expected selector after '.'");
                expression = {
                    kind: "SelectorExpr",
                    object: expression,
                    selector,
                    span: mergeSpans(expression.span, selector.span ?? dot.span)
                };
                continue;
            }
            if (this.match(TokenKind.LParen)) {
                const args = [];
                let ellipsis = false;
                if (!this.at(TokenKind.RParen)) {
                    args.push(this.parseExpression());
                    if (this.match(TokenKind.Ellipsis))
                        ellipsis = true;
                    while (this.match(TokenKind.Comma) && !this.at(TokenKind.RParen)) {
                        args.push(this.parseExpression());
                        if (this.match(TokenKind.Ellipsis))
                            ellipsis = true;
                    }
                }
                const close = this.expect(TokenKind.RParen, "expected ')' after arguments");
                expression = {
                    kind: "CallExpr",
                    fun: expression,
                    args,
                    ellipsis,
                    span: mergeSpans(expression.span, close.span)
                };
                continue;
            }
            if (this.match(TokenKind.LBracket)) {
                const low = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket) ? undefined : this.parseExpression();
                if (this.match(TokenKind.Colon)) {
                    const high = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket) ? undefined : this.parseExpression();
                    const max = this.match(TokenKind.Colon) ? this.parseExpression() : undefined;
                    const close = this.expect(TokenKind.RBracket, "expected ']' after slice");
                    expression = {
                        kind: "SliceExpr",
                        object: expression,
                        ...(low ? { low } : {}),
                        ...(high ? { high } : {}),
                        ...(max ? { max } : {}),
                        span: mergeSpans(expression.span, close.span)
                    };
                }
                else {
                    const close = this.expect(TokenKind.RBracket, "expected ']' after index");
                    expression = {
                        kind: "IndexExpr",
                        object: expression,
                        index: low ?? badExpr(close.span),
                        span: mergeSpans(expression.span, close.span)
                    };
                }
                continue;
            }
            if (this.at(TokenKind.LBrace) && this.canUseCompositeLiteralType(expression)) {
                this.advance();
                expression = this.finishCompositeLiteral(expression);
                continue;
            }
            break;
        }
        return expression;
    }
    parseOperand() {
        const token = this.peek();
        if (this.at(TokenKind.Identifier) || this.at(TokenKind.CellAddress))
            return this.parseIdent("expected identifier");
        if (this.at(TokenKind.True) || this.at(TokenKind.False) || this.at(TokenKind.Nil)) {
            const keyword = this.advance();
            return ident(keyword.lexeme, keyword.span);
        }
        if (this.at(TokenKind.IntLiteral) || this.at(TokenKind.FloatLiteral) || this.at(TokenKind.ImagLiteral) || this.at(TokenKind.RuneLiteral) || this.at(TokenKind.StringLiteral)) {
            return this.expectBasicLit(this.peek().kind, "expected literal");
        }
        if (this.match(TokenKind.LParen)) {
            const start = this.previous();
            const expr = this.parseExpression();
            const close = this.expect(TokenKind.RParen, "expected ')'");
            return { kind: "ParenExpr", expr, span: mergeSpans(start.span, close.span) };
        }
        if (this.atAny(TokenKind.LBracket, TokenKind.Map, TokenKind.Struct, TokenKind.Interface, TokenKind.Chan))
            return this.parseType();
        if (this.match(TokenKind.Func)) {
            const start = this.previous();
            const type = this.parseSignature(start.span);
            if (this.at(TokenKind.LBrace)) {
                const body = this.parseBlock();
                return { kind: "FuncLit", type, body, span: mergeSpans(start.span, body.span) };
            }
            return type;
        }
        this.error(`expected expression, found ${token.lexeme || token.kind}`, token.span);
        this.advance();
        return badExpr(token.span);
    }
    canUseCompositeLiteralType(expression) {
        return expression.kind === "ArrayType" ||
            expression.kind === "MapType" ||
            expression.kind === "StructType" ||
            (this.allowBareIdentifierComposite && ((expression.kind === "Ident" && !["true", "false", "nil"].includes(expression.name)) ||
                expression.kind === "SelectorExpr"));
    }
    finishCompositeLiteral(type) {
        const elements = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            const first = this.parseExpression();
            if (this.match(TokenKind.Colon)) {
                const value = this.parseExpression();
                elements.push({
                    kind: "KeyValueExpr",
                    key: first,
                    value,
                    span: mergeSpans(first.span, value.span)
                });
            }
            else {
                elements.push(first);
            }
            this.match(TokenKind.Comma);
            this.consumeSemi();
        }
        const close = this.expect(TokenKind.RBrace, "expected '}' after composite literal");
        return {
            kind: "CompositeLit",
            type,
            elements,
            span: mergeSpans(type.span, close.span)
        };
    }
    parseType() {
        const start = this.peek();
        if (this.match(TokenKind.Arrow)) {
            const chan = this.expect(TokenKind.Chan, "expected chan after '<-' in channel type");
            const value = this.parseType();
            return { kind: "ChanType", direction: "receive", value, span: mergeSpans(start.span, value.span ?? chan.span) };
        }
        if (this.match(TokenKind.Star)) {
            const expr = this.parseType();
            return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
        }
        if (this.match(TokenKind.LBracket)) {
            let length;
            let inferredLength = false;
            if (this.match(TokenKind.Ellipsis)) {
                inferredLength = true;
            }
            else if (!this.at(TokenKind.RBracket)) {
                length = this.parseExpression();
            }
            this.expect(TokenKind.RBracket, "expected ']' in array or slice type");
            const element = this.parseType();
            return {
                kind: "ArrayType",
                ...(length ? { length } : {}),
                element,
                inferredLength,
                span: mergeSpans(start.span, element.span)
            };
        }
        if (this.match(TokenKind.Map)) {
            this.expect(TokenKind.LBracket, "expected '[' after map");
            const key = this.parseType();
            this.expect(TokenKind.RBracket, "expected ']' after map key type");
            const value = this.parseType();
            return { kind: "MapType", key, value, span: mergeSpans(start.span, value.span) };
        }
        if (this.match(TokenKind.Chan)) {
            const sendOnly = this.match(TokenKind.Arrow);
            const value = this.parseType();
            return {
                kind: "ChanType",
                direction: sendOnly ? "send" : "both",
                value,
                span: mergeSpans(start.span, value.span)
            };
        }
        if (this.match(TokenKind.Struct))
            return this.parseStructType(start.span);
        if (this.match(TokenKind.Interface))
            return this.parseInterfaceType(start.span);
        if (this.match(TokenKind.Func))
            return this.parseSignature(start.span);
        return this.parseTypeName();
    }
    parseTypeName() {
        let expression = this.parseIdent("expected type name");
        while (this.match(TokenKind.Dot)) {
            const selector = this.parseIdent("expected selector in qualified type");
            expression = {
                kind: "SelectorExpr",
                object: expression,
                selector,
                span: mergeSpans(expression.span, selector.span)
            };
        }
        return expression;
    }
    parseStructType(start) {
        const fields = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
        return { kind: "StructType", fields, span: mergeSpans(start, fields.span) };
    }
    parseInterfaceType(start) {
        const methods = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
        return { kind: "InterfaceType", methods, span: mergeSpans(start, methods.span) };
    }
    parseSignature(start) {
        const params = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        let results;
        if (this.at(TokenKind.LParen)) {
            results = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        }
        else if (this.startsType()) {
            const type = this.parseType();
            results = {
                kind: "FieldList",
                fields: [{ kind: "Field", names: [], type, span: mergeSpans(type.span, type.span) }],
                span: mergeSpans(type.span, type.span)
            };
        }
        return {
            kind: "FuncType",
            params,
            ...(results ? { results } : {}),
            span: mergeSpans(start, results?.span ?? params.span)
        };
    }
    parseFieldList(open, close) {
        const start = this.expect(open, `expected '${open === TokenKind.LParen ? "(" : "{"}'`);
        const fields = [];
        while (!this.at(close) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(close))
                break;
            fields.push(this.parseField(close));
            if (!this.match(TokenKind.Comma))
                this.consumeSemi();
        }
        const end = this.expect(close, `expected '${close === TokenKind.RParen ? ")" : "}"}'`);
        return { kind: "FieldList", fields, span: mergeSpans(start.span, end.span) };
    }
    parseField(close) {
        const start = this.peek();
        if (this.at(TokenKind.Identifier) && this.peek(1).kind === TokenKind.LParen) {
            const name = this.parseIdent("expected method name");
            const type = this.parseSignature(name.span ?? start.span);
            return {
                kind: "Field",
                names: [name],
                type,
                span: mergeSpans(name.span, type.span)
            };
        }
        if (this.at(TokenKind.Identifier) && this.fieldHasExplicitNames(close)) {
            const names = this.parseIdentList();
            const type = this.match(TokenKind.Ellipsis)
                ? { kind: "Ellipsis", element: this.parseType(), span: start.span }
                : this.parseType();
            const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
            return {
                kind: "Field",
                names,
                type,
                ...(tag ? { tag } : {}),
                span: mergeSpans(names[0]?.span, tag?.span ?? type.span)
            };
        }
        const type = this.match(TokenKind.Ellipsis)
            ? { kind: "Ellipsis", element: this.parseType(), span: start.span }
            : this.parseType();
        const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
        return { kind: "Field", names: [], type, ...(tag ? { tag } : {}), span: mergeSpans(type.span, tag?.span ?? type.span) };
    }
    fieldHasExplicitNames(close) {
        let offset = 0;
        if (!isIdentifierLike(this.peek(offset).kind))
            return false;
        offset += 1;
        while (this.peek(offset).kind === TokenKind.Comma && isIdentifierLike(this.peek(offset + 1).kind)) {
            offset += 2;
        }
        const afterNames = this.peek(offset).kind;
        if (afterNames === close || afterNames === TokenKind.Semicolon || afterNames === TokenKind.RBrace || afterNames === TokenKind.RParen) {
            return false;
        }
        return this.startsType(afterNames) || afterNames === TokenKind.Ellipsis;
    }
    parseIdentList() {
        const names = [this.parseIdent("expected identifier")];
        while (this.match(TokenKind.Comma)) {
            names.push(this.parseIdent("expected identifier after ','"));
        }
        return names;
    }
    parseIdent(message) {
        const token = this.peek();
        if (isIdentifierLike(token.kind)) {
            this.advance();
            return ident(token.lexeme, token.span);
        }
        this.error(message, token.span);
        this.advance();
        return ident("<missing>", token.span);
    }
    expectBasicLit(kind, message) {
        const token = this.peek();
        if (this.match(kind)) {
            return { kind: "BasicLit", token: kind, value: token.lexeme, span: token.span };
        }
        this.error(message, token.span);
        return { kind: "BasicLit", token: kind, value: "", span: token.span };
    }
    looksLikeReceiver() {
        let depth = 0;
        for (let offset = 0; offset < 32; offset += 1) {
            const kind = this.peek(offset).kind;
            if (kind === TokenKind.LParen)
                depth += 1;
            if (kind === TokenKind.RParen) {
                depth -= 1;
                if (depth === 0)
                    return this.peek(offset + 1).kind === TokenKind.Identifier;
            }
            if (kind === TokenKind.EOF || kind === TokenKind.LBrace)
                return false;
        }
        return false;
    }
    looksLikeRangeClause() {
        for (let offset = 0; offset < 64; offset += 1) {
            const kind = this.peek(offset).kind;
            if (kind === TokenKind.Range)
                return true;
            if (kind === TokenKind.LBrace || kind === TokenKind.Semicolon || kind === TokenKind.EOF)
                return false;
        }
        return false;
    }
    startsType(kind = this.peek().kind) {
        return isIdentifierLike(kind) ||
            kind === TokenKind.Star ||
            kind === TokenKind.LBracket ||
            kind === TokenKind.Map ||
            kind === TokenKind.Chan ||
            kind === TokenKind.Arrow ||
            kind === TokenKind.Struct ||
            kind === TokenKind.Interface ||
            kind === TokenKind.Func;
    }
    withBareIdentifierComposites(enabled, fn) {
        const previous = this.allowBareIdentifierComposite;
        this.allowBareIdentifierComposite = enabled;
        try {
            return fn();
        }
        finally {
            this.allowBareIdentifierComposite = previous;
        }
    }
    consumeSemi() {
        this.match(TokenKind.Semicolon);
    }
    skipSemis() {
        while (this.match(TokenKind.Semicolon)) {
            // consume all empty statements between declarations/statements
        }
    }
    match(kind) {
        if (!this.at(kind))
            return false;
        this.advance();
        return true;
    }
    expect(kind, message) {
        if (this.at(kind))
            return this.advance();
        const token = this.peek();
        this.error(message, token.span);
        return token;
    }
    expectAny(kinds, message) {
        if (kinds.includes(this.peek().kind))
            return this.advance();
        const token = this.peek();
        this.error(message, token.span);
        return token;
    }
    at(kind) {
        return this.peek().kind === kind;
    }
    atAny(...kinds) {
        const current = this.peek().kind;
        return kinds.includes(current);
    }
    advance() {
        const token = this.peek();
        if (!this.at(TokenKind.EOF))
            this.index += 1;
        return token;
    }
    previous() {
        return this.tokens[Math.max(0, this.index - 1)] ?? this.peek();
    }
    peek(ahead = 0) {
        return this.tokens[this.index + ahead] ?? this.tokens[this.tokens.length - 1] ?? eofToken();
    }
    error(message, span, code = "GOJR_PARSE_FRONT001") {
        this.diagnostics.push({
            filename: diagnosticFilename(span, this.filename),
            code,
            severity: "error",
            message,
            ...(span ? { span } : {})
        });
    }
}
function binaryPrecedence(kind) {
    switch (kind) {
        case TokenKind.OrOr: return 1;
        case TokenKind.AndAnd: return 2;
        case TokenKind.Equal:
        case TokenKind.NotEqual:
        case TokenKind.Less:
        case TokenKind.LessEqual:
        case TokenKind.Greater:
        case TokenKind.GreaterEqual:
            return 3;
        case TokenKind.Plus:
        case TokenKind.Minus:
        case TokenKind.Or:
        case TokenKind.Caret:
            return 4;
        case TokenKind.Star:
        case TokenKind.Slash:
        case TokenKind.Percent:
        case TokenKind.Shl:
        case TokenKind.Shr:
        case TokenKind.Amp:
        case TokenKind.BitClear:
            return 5;
        default:
            return 0;
    }
}
function badExpr(span) {
    return { kind: "BadExpr", ...(span ? { span } : {}) };
}
function mergeSpans(start, end) {
    if (!start && !end)
        return eofToken().span;
    if (!start)
        return end ?? eofToken().span;
    if (!end)
        return start;
    const endOffset = end.offset + end.length;
    return {
        filename: start.filename,
        offset: start.offset,
        length: Math.max(0, endOffset - start.offset),
        line: start.line,
        column: start.column
    };
}
function eofToken() {
    return {
        kind: TokenKind.EOF,
        lexeme: "",
        span: { filename: REPL_FILENAME, offset: 0, length: 0, line: 1, column: 1 }
    };
}
