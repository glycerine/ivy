import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan, diagnosticFilename } from "../diagnostics.js";
import {
  ArrayType,
  BasicLit,
  BinaryOperator,
  BlockStmt,
  BranchStmt,
  CallExpr,
  CaseClause,
  CellAddress,
  CompositeLit,
  Decl,
  DeferStmt,
  Expr,
  Field,
  FieldList,
  File,
  FuncDecl,
  FuncLit,
  FuncType,
  GenDecl,
  Ident,
  ImportSpec,
  KeyValueExpr,
  parseCellAddress,
  RangeStmt,
  Spec,
  Stmt,
  TypeSpec,
  ValueSpec,
  ident
} from "./ast.js";
import { scanSource } from "./scanner.js";
import { FrontToken, isAssignmentToken, isIdentifierLike, TokenKind } from "./token.js";

export interface ParseFrontResult {
  file?: File;
  statements: Stmt[];
  diagnostics: Diagnostic[];
  tokens: FrontToken[];
}

export interface ParseFrontFilesResult {
  files: File[];
  statements: Stmt[];
  diagnostics: Diagnostic[];
  results: ParseFrontResult[];
}

type SimpleStmtMode = "basic" | "labelOk" | "rangeOk";

interface SimpleStmtResult {
  statement: Stmt;
  isRange: boolean;
}

export function parseFrontSource(source: string, filename: string): ParseFrontResult {
  const scanned = scanSource(source, filename);
  const parser = new FrontParser(scanned.tokens, scanned.diagnostics, filename);
  return parser.parseFile();
}

export function parseFrontSourceFiles(files: SourceFile[]): ParseFrontFilesResult {
  const results = files.map((file) => parseFrontSource(file.source, file.filename));
  return {
    files: results.flatMap((result) => result.file ? [result.file] : []),
    statements: results.flatMap((result) => result.statements),
    diagnostics: results.flatMap((result) => result.diagnostics),
    results
  };
}

class FrontParser {
  private index = 0;
  // Faithful port of go/parser.parser.exprLev:
  // exprLev < 0 means we are parsing an if/for/switch control clause, where
  // a following "{" may be the statement body rather than a composite literal.
  // Parenthesized/call/index subexpressions increment it back into expression
  // context, exactly like the standard parser.
  private exprLev = 0;
  // Faithful port of go/parser.parser.inRhs: while parsing right-hand-side
  // expression lists, "=" is treated as equality for tolerant parsing.
  private inRhs = false;
  private allowBareIdentifierComposite = true;
  private allowSpreadsheetRanges = true;
  private readonly diagnostics: Diagnostic[];

  public constructor(
    private readonly tokens: FrontToken[],
    diagnostics: Diagnostic[],
    private readonly filename: string
  ) {
    this.diagnostics = [...diagnostics];
  }

  public parseFile(): ParseFrontResult {
    let name: Ident | undefined;
    if (this.match(TokenKind.Package)) {
      name = this.parseIdent("expected package name");
      this.consumeSemi();
    }

    const declarations: Decl[] = [];
    const imports: ImportSpec[] = [];
    const statements: Stmt[] = [];
    while (!this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.EOF)) break;
      if (this.startsDecl()) {
        const declaration = this.parseDecl();
        declarations.push(declaration);
        if (declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
          imports.push(...declaration.specs.filter((spec): spec is ImportSpec => spec.kind === "ImportSpec"));
        }
      } else {
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

  private parseDecl(): Decl {
    if (this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var)) return this.parseGenDecl();
    if (this.at(TokenKind.Func)) return this.parseFuncDecl();
    const token = this.peek();
    this.error(`expected declaration, found ${token.lexeme || token.kind}`, token.span);
    this.advance();
    return { kind: "BadDecl", span: token.span };
  }

  private startsDecl(): boolean {
    return this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func);
  }

  private parseGenDecl(): GenDecl {
    const start = this.advance();
    const token = start.kind as TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var;
    const specs: Spec[] = [];
    let grouped = false;

    if (this.match(TokenKind.LParen)) {
      grouped = true;
      while (!this.at(TokenKind.RParen) && !this.at(TokenKind.EOF)) {
        this.skipSemis();
        if (this.at(TokenKind.RParen)) break;
        specs.push(this.parseSpec(token));
        this.consumeSemi();
      }
      this.expect(TokenKind.RParen, "expected ')' after declaration group");
    } else {
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

  private parseSpec(token: TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var): Spec {
    if (token === TokenKind.Import) return this.parseImportSpec();
    if (token === TokenKind.Type) return this.parseTypeSpec();
    return this.parseValueSpec();
  }

  private parseImportSpec(): ImportSpec {
    const start = this.peek();
    let name: Ident | undefined;
    if (this.at(TokenKind.Identifier) && this.peek(1).kind === TokenKind.StringLiteral) {
      name = this.parseIdent("expected import alias");
    } else if (this.at(TokenKind.Dot) && this.peek(1).kind === TokenKind.StringLiteral) {
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

  private parseTypeSpec(): TypeSpec {
    const name = this.parseIdent("expected type name");
    let typeParams: FieldList | undefined;
    let alias = false;
    let type: Expr;

    if (this.match(TokenKind.LBracket)) {
      const open = this.previous();
      if (isIdentifierLike(this.peek().kind)) {
        const firstName = this.parseIdent("expected type parameter name or array length");
        let expression: Expr = firstName;
        if (!this.at(TokenKind.LBracket)) {
          this.withExpressionLevel(() => {
            expression = this.parseBinaryExpression(this.parsePrimaryFrom(expression), 1);
          });
        }
        const { name: paramName, type: paramType } = extractName(expression, this.at(TokenKind.Comma));
        if (paramName && (paramType || !this.at(TokenKind.RBracket))) {
          typeParams = this.parseTypeParameterListAfterOpen(open, paramName, paramType);
          alias = this.match(TokenKind.Assign);
          type = this.parseType();
        } else {
          type = this.parseArrayTypeAfterOpen(open, expression);
        }
      } else {
        type = this.parseArrayTypeAfterOpen(open);
      }
    } else {
      alias = this.match(TokenKind.Assign);
      type = this.parseType();
    }
    return {
      kind: "TypeSpec",
      name,
      ...(typeParams ? { typeParams } : {}),
      type,
      alias,
      span: mergeSpans(name.span, type.span)
    };
  }

  private parseValueSpec(): ValueSpec {
    const names = this.parseIdentList();
    let type: Expr | undefined;
    let values: Expr[] = [];
    if (!this.atAny(TokenKind.Assign, TokenKind.Semicolon, TokenKind.RParen, TokenKind.EOF)) {
      type = this.parseType();
    }
    if (this.match(TokenKind.Assign)) {
      values = this.parseExpressionList(true);
    }
    return {
      kind: "ValueSpec",
      names,
      ...(type ? { type } : {}),
      values,
      span: mergeSpans(names[0]?.span, values[values.length - 1]?.span ?? type?.span ?? names[names.length - 1]?.span)
    };
  }

  private parseFuncDecl(): FuncDecl {
    const start = this.expect(TokenKind.Func, "expected func");
    let receiver: FieldList | undefined;
    if (this.at(TokenKind.LParen) && this.looksLikeReceiver()) {
      receiver = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    }
    const name = this.parseIdent("expected function name");
    const typeParams = this.at(TokenKind.LBracket) ? this.parseTypeParamList() : undefined;
    const type = this.parseSignature(start.span, typeParams);
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

  private parseStatement(): Stmt {
    this.skipSemis();
    if (this.at(TokenKind.RBrace)) {
      return { kind: "EmptyStmt", implicit: true, span: this.peek().span };
    }
    if (this.at(TokenKind.LBrace)) return this.parseBlock();
    if (this.atAny(TokenKind.Const, TokenKind.Type, TokenKind.Var)) {
      return { kind: "DeclStmt", decl: this.parseGenDecl() };
    }
    if (this.match(TokenKind.Return)) {
      const start = this.previous();
      const results = this.atAny(TokenKind.Semicolon, TokenKind.RBrace, TokenKind.EOF) ? [] : this.parseExpressionList(true);
      return { kind: "ReturnStmt", results, span: mergeSpans(start.span, results[results.length - 1]?.span ?? start.span) };
    }
    if (this.atAny(TokenKind.Break, TokenKind.Continue, TokenKind.Goto, TokenKind.Fallthrough)) {
      const token = this.advance();
      const acceptsLabel = token.kind === TokenKind.Goto ||
        ((token.kind === TokenKind.Break || token.kind === TokenKind.Continue) && this.at(TokenKind.Identifier));
      const label = acceptsLabel ? this.parseIdent("expected label") : undefined;
      return {
        kind: "BranchStmt",
        token: token.kind as BranchStmt["token"],
        ...(label ? { label } : {}),
        span: mergeSpans(token.span, label?.span)
      };
    }
    if (this.match(TokenKind.Defer)) {
      const start = this.previous();
      const expression = this.parseRhsExpression();
      const call = expression.kind === "CallExpr"
        ? expression
        : ({ kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) } satisfies CallExpr);
      return { kind: "DeferStmt", call, span: mergeSpans(start.span, call.span) } satisfies DeferStmt;
    }
    if (this.at(TokenKind.Go)) return this.parseGoStmt();
    if (this.at(TokenKind.Select)) return this.parseSelectStmt();
    if (this.at(TokenKind.For)) return this.parseForStmt();
    if (this.at(TokenKind.If)) return this.parseIfStmt();
    if (this.at(TokenKind.Switch)) return this.parseSwitchStmt();

    return this.parseSimpleStmt("labelOk");
  }

  private parseSimpleStmt(mode: SimpleStmtMode = "basic"): Stmt {
    return this.parseSimpleStmtWithRange(mode).statement;
  }

  private parseSimpleStmtWithRange(mode: SimpleStmtMode): SimpleStmtResult {
    const lhs = this.parseExpressionList(false);
    if (this.match(TokenKind.Arrow)) {
      if (lhs.length !== 1) {
        this.error("send statement expects one channel expression", lhs[1]?.span ?? lhs[0]?.span);
      }
      const value = this.parseRhsExpression();
      return {
        statement: {
        kind: "SendStmt",
        channel: lhs[0] ?? badExpr(value.span),
        value,
        span: mergeSpans(lhs[0]?.span, value.span)
        },
        isRange: false
      };
    }
    if (isAssignmentToken(this.peek().kind)) {
      const token = this.advance();
      let rhs: Expr[];
      let isRange = false;
      if ((token.kind === TokenKind.Assign || token.kind === TokenKind.Define) && mode === "rangeOk" && this.match(TokenKind.Range)) {
        rhs = [this.parseRhsExpression()];
        isRange = true;
      } else {
        rhs = this.parseExpressionList(true);
      }
      return {
        statement: {
        kind: "AssignStmt",
        lhs,
        token: token.kind as Extract<Stmt, { kind: "AssignStmt" }>["token"],
        rhs,
        span: mergeSpans(lhs[0]?.span, rhs[rhs.length - 1]?.span)
        },
        isRange
      };
    }
    if (this.match(TokenKind.Colon)) {
      const colon = this.previous();
      if (mode === "labelOk" && lhs[0]?.kind === "Ident" && lhs.length === 1) {
        const stmt = this.parseStatement();
        return {
          statement: {
            kind: "LabeledStmt",
            label: lhs[0],
            stmt,
            span: mergeSpans(lhs[0].span, stmt.span)
          },
          isRange: false
        };
      }
      this.error("illegal label declaration", colon.span);
      return { statement: { kind: "BadStmt", span: mergeSpans(lhs[0]?.span, colon.span) }, isRange: false };
    }
    if (this.at(TokenKind.PlusPlus) || this.at(TokenKind.MinusMinus)) {
      const token = this.advance();
      return {
        statement: {
        kind: "IncDecStmt",
        expr: lhs[0] ?? badExpr(token.span),
        token: token.kind as TokenKind.PlusPlus | TokenKind.MinusMinus,
        span: mergeSpans(lhs[0]?.span, token.span)
        },
        isRange: false
      };
    }
    if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
    return {
      statement: { kind: "ExprStmt", expr: lhs[0] ?? badExpr(this.peek().span), span: mergeSpans(lhs[0]?.span, lhs[0]?.span) },
      isRange: false
    };
  }

  private parseGoStmt(): Stmt {
    const start = this.expect(TokenKind.Go, "expected go");
    const expression = this.parseRhsExpression();
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

  private parseIfStmt(): Stmt {
    const start = this.expect(TokenKind.If, "expected if");
    const { init, condition } = this.withControlClause(() => {
      let init: Stmt | undefined;
      let conditionStatement: Stmt | undefined;
      if (this.at(TokenKind.LBrace)) {
        this.error("missing condition in if statement", this.peek().span);
        return { condition: badExpr(this.peek().span) };
      }
      if (!this.at(TokenKind.Semicolon)) {
        if (this.match(TokenKind.Var)) this.error("var declaration not allowed in if initializer", this.previous().span);
        init = this.parseSimpleStmt("basic");
      }
      if (!this.at(TokenKind.LBrace)) {
        this.expect(TokenKind.Semicolon, "expected ';' after if init statement");
        if (!this.at(TokenKind.LBrace)) conditionStatement = this.parseSimpleStmt("basic");
      } else {
        conditionStatement = init;
        init = undefined;
      }
      return {
        ...(init ? { init } : {}),
        condition: this.statementExpression(conditionStatement, "boolean expression") ?? badExpr(this.peek().span)
      };
    });
    const body = this.parseBlock();
    let elseStmt: Stmt | undefined;
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

  private parseForStmt(): Stmt {
    const start = this.expect(TokenKind.For, "expected for");
    const header = this.withControlClause(() => {
      let first: Stmt | undefined;
      let second: Stmt | undefined;
      let third: Stmt | undefined;
      let isRange = false;

      if (!this.at(TokenKind.LBrace)) {
        if (!this.at(TokenKind.Semicolon)) {
          if (this.match(TokenKind.Range)) {
            const source = this.parseRhsExpression();
            second = { kind: "AssignStmt", lhs: [], token: TokenKind.Assign, rhs: [source], span: source.span ?? start.span };
            isRange = true;
          } else {
            const parsed = this.parseSimpleStmtWithRange("rangeOk");
            second = parsed.statement;
            isRange = parsed.isRange;
          }
        }
        if (!isRange && this.match(TokenKind.Semicolon)) {
          first = second;
          second = undefined;
          if (!this.at(TokenKind.Semicolon)) second = this.parseSimpleStmt("basic");
          this.expect(TokenKind.Semicolon, "expected ';' in for clause");
          if (!this.at(TokenKind.LBrace)) third = this.parseSimpleStmt("basic");
        }
      }
      return { first, second, third, isRange };
    });
    const body = this.parseBlock();
    if (header.isRange) {
      const assign = header.second?.kind === "AssignStmt" ? header.second : undefined;
      const lhs = assign?.lhs ?? [];
      if (lhs.length > 2) this.error("expected at most 2 expressions", lhs[2]?.span ?? assign?.span);
      const source = assign?.rhs[0] ?? badExpr(header.second?.span ?? start.span);
      return {
        kind: "RangeStmt",
        ...(lhs[0] ? { key: lhs[0] } : {}),
        ...(lhs[1] ? { value: lhs[1] } : {}),
        token: assign?.token === TokenKind.Define ? TokenKind.Define : TokenKind.Assign,
        source,
        body,
        span: mergeSpans(start.span, body.span)
      };
    }
    const condition = this.statementExpression(header.second, "boolean or range expression");
    return {
      kind: "ForStmt",
      ...(header.first ? { init: header.first } : {}),
      ...(condition ? { condition } : {}),
      ...(header.third ? { post: header.third } : {}),
      body,
      span: mergeSpans(start.span, body.span)
    };
  }

  private parseSwitchStmt(): Stmt {
    const start = this.expect(TokenKind.Switch, "expected switch");
    const { init, tag, assign, typeSwitch } = this.withControlClause(() => {
      let init: Stmt | undefined;
      let tag: Expr | undefined;
      let assign: Stmt | undefined;
      let typeSwitch = false;

      if (!this.at(TokenKind.LBrace)) {
        const first = this.at(TokenKind.Semicolon) ? undefined : this.parseSimpleStmt("basic");
        if (this.match(TokenKind.Semicolon)) {
          init = first;
          if (!this.at(TokenKind.LBrace)) {
            const second = this.parseSimpleStmt("basic");
            if (this.isTypeSwitchGuard(second)) {
              assign = second;
              typeSwitch = true;
            } else if (second.kind === "ExprStmt") {
              tag = second.expr;
            } else {
              this.error("expected switch expression or type switch guard after ';'", second.span);
            }
          }
        } else if (first && this.isTypeSwitchGuard(first)) {
          assign = first;
          typeSwitch = true;
        } else if (first?.kind === "ExprStmt") {
          tag = first.expr;
        } else if (first) {
          this.error("expected ';' after switch init statement", first.span);
          init = first;
        }
      }
      return { init, tag, assign, typeSwitch };
    });

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

  private parseSwitchBody(): { clauses: CaseClause[]; span: SourceSpan } {
    const start = this.expect(TokenKind.LBrace, "expected '{' after switch");
    const clauses: CaseClause[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
        clauses.push(this.parseCaseClause());
      } else {
        const token = this.peek();
        this.error("expected case or default in switch body", token.span);
        this.advance();
      }
    }
    const end = this.expect(TokenKind.RBrace, "expected '}' after switch body");
    return { clauses, span: mergeSpans(start.span, end.span) };
  }

  private parseSelectStmt(): Stmt {
    const start = this.expect(TokenKind.Select, "expected select");
    const open = this.expect(TokenKind.LBrace, "expected '{' after select");
    const clauses: Extract<Stmt, { kind: "CommClause" }>[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
        clauses.push(this.parseCommClause());
      } else {
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

  private parseCommClause(): Extract<Stmt, { kind: "CommClause" }> {
    const start = this.advance();
    const isDefault = start.kind === TokenKind.Default;
    let comm: Stmt | undefined;
    if (!isDefault && !this.at(TokenKind.Colon)) {
      const lhs = this.parseExpressionList(false);
      if (this.match(TokenKind.Arrow)) {
        if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
        const value = this.parseRhsExpression();
        comm = {
          kind: "SendStmt",
          channel: lhs[0] ?? badExpr(value.span),
          value,
          span: mergeSpans(lhs[0]?.span, value.span)
        };
      } else if (this.at(TokenKind.Assign) || this.at(TokenKind.Define)) {
        if (lhs.length > 2) this.error("expected 1 or 2 expressions", lhs[2]?.span ?? lhs[0]?.span);
        const token = this.advance();
        const rhs = this.parseRhsExpression();
        comm = {
          kind: "AssignStmt",
          lhs,
          token: token.kind as TokenKind.Assign | TokenKind.Define,
          rhs: [rhs],
          span: mergeSpans(lhs[0]?.span, rhs.span)
        };
      } else {
        if (lhs.length > 1) this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
        comm = {
          kind: "ExprStmt",
          expr: lhs[0] ?? badExpr(start.span),
          span: mergeSpans(lhs[0]?.span, lhs[0]?.span)
        };
      }
    }
    this.expect(TokenKind.Colon, "expected ':' after select case");
    const body: Stmt[] = [];
    while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
      this.skipSemis();
      if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) break;
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

  private parseCaseClause(): CaseClause {
    const start = this.advance();
    const isDefault = start.kind === TokenKind.Default;
    const list = isDefault ? [] : this.withSpreadsheetRanges(false, () => this.parseExpressionList(true));
    this.expect(TokenKind.Colon, "expected ':' after switch case");
    const body: Stmt[] = [];
    while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
      this.skipSemis();
      if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) break;
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

  private isTypeSwitchGuard(statement: Stmt): boolean {
    if (statement.kind === "ExprStmt") {
      return statement.expr.kind === "TypeAssertExpr" && statement.expr.typeSwitch;
    }
    if (statement.kind !== "AssignStmt" || statement.rhs.length !== 1) return false;
    const rhs = statement.rhs[0];
    return rhs?.kind === "TypeAssertExpr" && rhs.typeSwitch;
  }

  private parseBlock(): BlockStmt {
    const start = this.expect(TokenKind.LBrace, "expected '{'");
    const statements: Stmt[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
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

  private parseExpressionList(inRhs = false): Expr[] {
    return this.withRhs(inRhs, () => {
      const expressions = [this.parseExpression()];
      while (this.match(TokenKind.Comma)) {
        expressions.push(this.parseExpression());
      }
      return expressions;
    });
  }

  private parseRhsExpression(): Expr {
    return this.withRhs(true, () => this.parseExpression());
  }

  private parseExpression(minPrecedence = 1): Expr {
    let left = this.parseUnary();
    while (true) {
      const operatorKind = this.inRhs && this.peek().kind === TokenKind.Assign
        ? TokenKind.Equal
        : this.peek().kind;
      const precedence = binaryPrecedence(operatorKind);
      if (precedence < minPrecedence) break;
      const operator = this.advance();
      const right = this.parseExpression(precedence + 1);
      left = {
        kind: "BinaryExpr",
        left,
        op: operatorKind as BinaryOperator,
        right,
        span: mergeSpans(left.span, right.span)
      };
    }
    return left;
  }

  private statementExpression(statement: Stmt | undefined, want: string): Expr | undefined {
    if (!statement) return undefined;
    if (statement.kind === "ExprStmt") return statement.expr;
    const found = statement.kind === "AssignStmt" ? "assignment" : "simple statement";
    this.error(`expected ${want}, found ${found} (missing parentheses around composite literal?)`, statement.span);
    return badExpr(statement.span);
  }

  private parseUnary(): Expr {
    if (this.atAny(TokenKind.Plus, TokenKind.Minus, TokenKind.Bang, TokenKind.Caret, TokenKind.Amp, TokenKind.Arrow)) {
      const operator = this.advance();
      const expr = this.parseUnary();
      return {
        kind: "UnaryExpr",
        op: operator.kind as TokenKind.Plus | TokenKind.Minus | TokenKind.Bang | TokenKind.Caret | TokenKind.Amp | TokenKind.Arrow,
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

  private parsePrimary(): Expr {
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
          } else {
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

        const cellCandidate = this.peek();
        const cellStart = (cellCandidate.kind === TokenKind.CellAddress || cellCandidate.kind === TokenKind.Identifier)
          ? parseCellAddress(cellCandidate.lexeme)
          : undefined;
        if (cellStart && expression.kind === "Ident" && (cellCandidate.kind === TokenKind.CellAddress || (this.allowSpreadsheetRanges && this.peek(1).kind === TokenKind.Colon))) {
          const cellToken = this.advance();
          if (this.match(TokenKind.Colon)) {
            const { token: endToken, address: end } = this.expectCellAddress("expected cell address after ':'");
            if (!end) {
              this.error("malformed spreadsheet range end", endToken.span);
              expression = badExpr(endToken.span);
            } else {
              expression = {
                kind: "RangeRefExpr",
                namespace: expression,
                start: cellStart,
                end,
                span: mergeSpans(expression.span, endToken.span)
              };
            }
          } else {
            expression = {
              kind: "CellRefExpr",
              namespace: expression,
              address: cellStart,
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
        const args: Expr[] = [];
        let ellipsis = false;
        if (!this.at(TokenKind.RParen)) {
          args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
          if (this.match(TokenKind.Ellipsis)) ellipsis = true;
          while (this.match(TokenKind.Comma) && !this.at(TokenKind.RParen)) {
            args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
            if (this.match(TokenKind.Ellipsis)) ellipsis = true;
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
        const open = this.previous();
        const low = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
          ? undefined
          : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
        if (this.match(TokenKind.Colon)) {
          const high = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
            ? undefined
            : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
          const max = this.match(TokenKind.Colon)
            ? this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()))
            : undefined;
          const close = this.expect(TokenKind.RBracket, "expected ']' after slice");
          expression = {
            kind: "SliceExpr",
            object: expression,
            ...(low ? { low } : {}),
            ...(high ? { high } : {}),
            ...(max ? { max } : {}),
            span: mergeSpans(expression.span, close.span)
          };
        } else if (this.match(TokenKind.Comma)) {
          const indices = [low ?? badExpr(open.span)];
          while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
            indices.push(this.withExpressionLevel(() => this.parseType()));
            if (!this.match(TokenKind.Comma)) break;
          }
          const close = this.expect(TokenKind.RBracket, "expected ']' after type arguments");
          expression = indices.length === 1
            ? {
              kind: "IndexExpr",
              object: expression,
              index: indices[0]!,
              span: mergeSpans(expression.span, close.span)
            }
            : {
              kind: "IndexListExpr",
              object: expression,
              indices,
              span: mergeSpans(expression.span, close.span)
            };
        } else {
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

  private parseOperand(): Expr {
    const token = this.peek();
    if (this.at(TokenKind.Identifier) || this.at(TokenKind.CellAddress)) return this.parseIdent("expected identifier");
    if (this.at(TokenKind.True) || this.at(TokenKind.False) || this.at(TokenKind.Nil)) {
      const keyword = this.advance();
      return ident(keyword.lexeme, keyword.span);
    }
    if (this.at(TokenKind.IntLiteral) || this.at(TokenKind.FloatLiteral) || this.at(TokenKind.ImagLiteral) || this.at(TokenKind.RuneLiteral) || this.at(TokenKind.StringLiteral)) {
      return this.expectBasicLit(this.peek().kind as BasicLit["token"], "expected literal");
    }
    if (this.match(TokenKind.LParen)) {
      const start = this.previous();
      const expr = this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
      const close = this.expect(TokenKind.RParen, "expected ')'");
      return { kind: "ParenExpr", expr, span: mergeSpans(start.span, close.span) };
    }
    if (this.atAny(TokenKind.LBracket, TokenKind.Map, TokenKind.Struct, TokenKind.Interface, TokenKind.Chan)) return this.parseType();
    if (this.match(TokenKind.Func)) {
      const start = this.previous();
      const type = this.parseSignature(start.span);
      if (this.at(TokenKind.LBrace)) {
        const body = this.parseBlock();
        return { kind: "FuncLit", type, body, span: mergeSpans(start.span, body.span) } satisfies FuncLit;
      }
      return type;
    }

    this.error(`expected expression, found ${token.lexeme || token.kind}`, token.span);
    this.advance();
    return badExpr(token.span);
  }

  private canUseCompositeLiteralType(expression: Expr): boolean {
    if (expression.kind === "ArrayType" || expression.kind === "MapType" || expression.kind === "StructType") {
      return true;
    }
    if (
      (expression.kind === "Ident" && !["true", "false", "nil"].includes(expression.name)) ||
      expression.kind === "SelectorExpr" ||
      expression.kind === "IndexExpr" ||
      expression.kind === "IndexListExpr"
    ) {
      return this.exprLev >= 0 && this.allowBareIdentifierComposite;
    }
    return false;
  }

  private finishCompositeLiteral(type?: Expr, startSpan: SourceSpan | undefined = type?.span): CompositeLit {
    const elements: Expr[] = [];
    while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(TokenKind.RBrace)) break;
      const first = this.parseCompositeLiteralElement();
      if (this.match(TokenKind.Colon)) {
        const value = this.parseCompositeLiteralElement();
        elements.push({
          kind: "KeyValueExpr",
          key: first,
          value,
          span: mergeSpans(first.span, value.span)
        } satisfies KeyValueExpr);
      } else {
        elements.push(first);
      }
      this.match(TokenKind.Comma);
      this.consumeSemi();
    }
    const close = this.expect(TokenKind.RBrace, "expected '}' after composite literal");
    return {
      kind: "CompositeLit",
      ...(type ? { type } : {}),
      elements,
      span: mergeSpans(startSpan, close.span)
    };
  }

  private parseCompositeLiteralElement(): Expr {
    if (this.match(TokenKind.LBrace)) {
      const start = this.previous();
      return this.finishCompositeLiteral(undefined, start.span);
    }
    return this.parseRhsExpression();
  }

  private parseType(): Expr {
    let left = this.parseTypeTerm();
    while (this.match(TokenKind.Or)) {
      const operator = this.previous();
      const right = this.parseTypeTerm();
      left = {
        kind: "BinaryExpr",
        left,
        op: operator.kind as BinaryOperator,
        right,
        span: mergeSpans(left.span, right.span)
      };
    }
    return left;
  }

  private parseTypeTerm(): Expr {
    const start = this.peek();
    if (this.match(TokenKind.Tilde)) {
      const expr = this.parseTypeTerm();
      return {
        kind: "UnaryExpr",
        op: TokenKind.Tilde,
        expr,
        span: mergeSpans(start.span, expr.span)
      };
    }
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
      let length: Expr | undefined;
      let inferredLength = false;
      if (this.match(TokenKind.Ellipsis)) {
        inferredLength = true;
      } else if (!this.at(TokenKind.RBracket)) {
        length = this.withExpressionLevel(() => this.parseRhsExpression());
      }
      this.expect(TokenKind.RBracket, "expected ']' in array or slice type");
      const element = this.parseType();
      return {
        kind: "ArrayType",
        ...(length ? { length } : {}),
        element,
        inferredLength,
        span: mergeSpans(start.span, element.span)
      } satisfies ArrayType;
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
    if (this.match(TokenKind.Struct)) return this.parseStructType(start.span);
    if (this.match(TokenKind.Interface)) return this.parseInterfaceType(start.span);
    if (this.match(TokenKind.Func)) return this.parseSignature(start.span);
    if (this.match(TokenKind.LParen)) {
      const type = this.parseType();
      const close = this.expect(TokenKind.RParen, "expected ')' after type");
      return { kind: "ParenExpr", expr: type, span: mergeSpans(start.span, close.span) };
    }
    return this.parseTypeName();
  }

  private parseTypeName(): Expr {
    let expression: Expr = this.parseIdent("expected type name");
    while (this.match(TokenKind.Dot)) {
      const selector = this.parseIdent("expected selector in qualified type");
      expression = {
        kind: "SelectorExpr",
        object: expression,
        selector,
        span: mergeSpans(expression.span, selector.span)
      };
    }
    if (this.match(TokenKind.LBracket)) {
      const { indices, close } = this.parseTypeArgumentList();
      expression = indices.length === 1 ? {
        kind: "IndexExpr",
        object: expression,
        index: indices[0]!,
        span: mergeSpans(expression.span, close.span)
      } : {
        kind: "IndexListExpr",
        object: expression,
        indices,
        span: mergeSpans(expression.span, close.span)
      };
    }
    return expression;
  }

  private parseTypeArgumentList(): { indices: Expr[]; close: FrontToken } {
    const indices: Expr[] = [];
    while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
      indices.push(this.withExpressionLevel(() => this.parseType()));
      if (!this.match(TokenKind.Comma)) break;
    }
    return {
      indices: indices.length > 0 ? indices : [badExpr(this.peek().span)],
      close: this.expect(TokenKind.RBracket, "expected ']' after type arguments")
    };
  }

  private parseStructType(start: SourceSpan): Expr {
    const fields = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
    return { kind: "StructType", fields, span: mergeSpans(start, fields.span) };
  }

  private parseInterfaceType(start: SourceSpan): Expr {
    const methods = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
    return { kind: "InterfaceType", methods, span: mergeSpans(start, methods.span) };
  }

  private parseSignature(start: SourceSpan, typeParams?: FieldList): FuncType {
    const params = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    let results: FieldList | undefined;
    if (this.at(TokenKind.LParen)) {
      results = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
    } else if (this.startsType()) {
      const type = this.parseType();
      results = {
        kind: "FieldList",
        fields: [{ kind: "Field", names: [], type, span: mergeSpans(type.span, type.span) }],
        span: mergeSpans(type.span, type.span)
      };
    }
    return {
      kind: "FuncType",
      ...(typeParams ? { typeParams } : {}),
      params,
      ...(results ? { results } : {}),
      span: mergeSpans(start, results?.span ?? params.span)
    };
  }

  private parseTypeParamList(): FieldList {
    return this.parseFieldList(TokenKind.LBracket, TokenKind.RBracket);
  }

  private parseFieldList(
    open: TokenKind.LParen | TokenKind.LBrace | TokenKind.LBracket,
    close: TokenKind.RParen | TokenKind.RBrace | TokenKind.RBracket
  ): FieldList {
    const start = this.expect(open, `expected '${tokenDisplay(open)}'`);
    const fields: Field[] = [];
    while (!this.at(close) && !this.at(TokenKind.EOF)) {
      this.skipSemis();
      if (this.at(close)) break;
      fields.push(this.parseField(close));
      if (!this.match(TokenKind.Comma)) this.consumeSemi();
    }
    const end = this.expect(close, `expected '${tokenDisplay(close)}'`);
    return { kind: "FieldList", fields, span: mergeSpans(start.span, end.span) };
  }

  private parseField(close: TokenKind): Field {
    const start = this.peek();
    if (isIdentifierLike(this.peek().kind) && this.peek(1).kind === TokenKind.LParen) {
      const name = this.parseIdent("expected method name");
      const type = this.parseSignature(name.span ?? start.span);
      return {
        kind: "Field",
        names: [name],
        type,
        span: mergeSpans(name.span, type.span)
      };
    }

    if (isIdentifierLike(this.peek().kind) && this.fieldHasExplicitNames(close)) {
      const names = this.parseIdentList();
      const type = this.match(TokenKind.Ellipsis)
        ? ({ kind: "Ellipsis", element: this.parseType(), span: start.span } as Expr)
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
      ? ({ kind: "Ellipsis", element: this.parseType(), span: start.span } as Expr)
      : this.parseType();
    const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
    return { kind: "Field", names: [], type, ...(tag ? { tag } : {}), span: mergeSpans(type.span, tag?.span ?? type.span) };
  }

  private fieldHasExplicitNames(close: TokenKind): boolean {
    let offset = 0;
    if (!isIdentifierLike(this.peek(offset).kind)) return false;
    offset += 1;
    while (this.peek(offset).kind === TokenKind.Comma && isIdentifierLike(this.peek(offset + 1).kind)) {
      offset += 2;
    }
    const afterNames = this.peek(offset).kind;
    if (offset === 1 && afterNames === TokenKind.LBracket) {
      const afterBracket = this.kindAfterBalancedBrackets(offset);
      if (
        afterBracket === close ||
        afterBracket === TokenKind.Semicolon ||
        afterBracket === TokenKind.RBrace ||
        afterBracket === TokenKind.RParen ||
        afterBracket === TokenKind.RBracket ||
        afterBracket === TokenKind.StringLiteral ||
        afterBracket === TokenKind.Comma ||
        afterBracket === TokenKind.EOF
      ) {
        return false;
      }
    }
    if (afterNames === close || afterNames === TokenKind.Semicolon || afterNames === TokenKind.RBrace || afterNames === TokenKind.RParen || afterNames === TokenKind.RBracket) {
      return false;
    }
    return this.startsType(afterNames) || afterNames === TokenKind.Ellipsis;
  }

  private kindAfterBalancedBrackets(openOffset: number): TokenKind {
    let depth = 0;
    for (let offset = openOffset; ; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.EOF) return TokenKind.EOF;
      if (kind === TokenKind.LBracket) depth += 1;
      if (kind === TokenKind.RBracket) {
        depth -= 1;
        if (depth === 0) return this.peek(offset + 1).kind;
      }
    }
  }

  private parseIdentList(): Ident[] {
    const names = [this.parseIdent("expected identifier")];
    while (this.match(TokenKind.Comma)) {
      names.push(this.parseIdent("expected identifier after ','"));
    }
    return names;
  }

  private parseIdent(message: string): Ident {
    const token = this.peek();
    if (isIdentifierLike(token.kind)) {
      this.advance();
      return ident(token.lexeme, token.span);
    }
    this.error(message, token.span);
    this.advance();
    return ident("<missing>", token.span);
  }

  private expectCellAddress(message: string): { token: FrontToken; address?: CellAddress } {
    const token = this.peek();
    if (token.kind === TokenKind.CellAddress || token.kind === TokenKind.Identifier) {
      this.advance();
      const address = parseCellAddress(token.lexeme);
      if (address) return { token, address };
      this.error(message, token.span);
      return { token };
    }
    this.error(message, token.span);
    this.advance();
    return { token };
  }

  private expectBasicLit(kind: BasicLit["token"], message: string): BasicLit {
    const token = this.peek();
    if (this.match(kind)) {
      return { kind: "BasicLit", token: kind, value: token.lexeme, span: token.span };
    }
    this.error(message, token.span);
    return { kind: "BasicLit", token: kind, value: "", span: token.span };
  }

  private looksLikeReceiver(): boolean {
    let depth = 0;
    for (let offset = 0; offset < 32; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.LParen) depth += 1;
      if (kind === TokenKind.RParen) {
        depth -= 1;
        if (depth === 0) return this.peek(offset + 1).kind === TokenKind.Identifier;
      }
      if (kind === TokenKind.EOF || kind === TokenKind.LBrace) return false;
    }
    return false;
  }

  private looksLikeRangeClause(): boolean {
    let parens = 0;
    let brackets = 0;
    let braces = 0;
    for (let offset = 0; offset < 64; offset += 1) {
      const kind = this.peek(offset).kind;
      if (kind === TokenKind.LParen) parens += 1;
      else if (kind === TokenKind.RParen && parens > 0) parens -= 1;
      else if (kind === TokenKind.LBracket) brackets += 1;
      else if (kind === TokenKind.RBracket && brackets > 0) brackets -= 1;
      else if (kind === TokenKind.LBrace) {
        if (parens === 0 && brackets === 0 && braces === 0) return false;
        braces += 1;
      } else if (kind === TokenKind.RBrace && braces > 0) braces -= 1;
      if (kind === TokenKind.Range) return true;
      if ((kind === TokenKind.Semicolon || kind === TokenKind.EOF) && parens === 0 && brackets === 0 && braces === 0) return false;
    }
    return false;
  }

  private startsType(kind = this.peek().kind): boolean {
    return isIdentifierLike(kind) ||
      kind === TokenKind.Star ||
      kind === TokenKind.Tilde ||
      kind === TokenKind.LBracket ||
      kind === TokenKind.Map ||
      kind === TokenKind.Chan ||
      kind === TokenKind.Arrow ||
      kind === TokenKind.Struct ||
      kind === TokenKind.Interface ||
      kind === TokenKind.Func;
  }

  private withBareIdentifierComposites<T>(enabled: boolean, fn: () => T): T {
    const previous = this.allowBareIdentifierComposite;
    this.allowBareIdentifierComposite = enabled;
    try {
      return fn();
    } finally {
      this.allowBareIdentifierComposite = previous;
    }
  }

  private withControlClause<T>(fn: () => T): T {
    const previous = this.exprLev;
    this.exprLev = -1;
    try {
      return fn();
    } finally {
      this.exprLev = previous;
    }
  }

  private withExpressionLevel<T>(fn: () => T): T {
    const previous = this.exprLev;
    this.exprLev += 1;
    try {
      return fn();
    } finally {
      this.exprLev = previous;
    }
  }

  private withRhs<T>(enabled: boolean, fn: () => T): T {
    const previous = this.inRhs;
    this.inRhs = enabled;
    try {
      return fn();
    } finally {
      this.inRhs = previous;
    }
  }

  private withSpreadsheetRanges<T>(enabled: boolean, fn: () => T): T {
    const previous = this.allowSpreadsheetRanges;
    this.allowSpreadsheetRanges = enabled;
    try {
      return fn();
    } finally {
      this.allowSpreadsheetRanges = previous;
    }
  }

  private consumeSemi(): void {
    this.match(TokenKind.Semicolon);
  }

  private skipSemis(): void {
    while (this.match(TokenKind.Semicolon)) {
      // consume all empty statements between declarations/statements
    }
  }

  private match(kind: TokenKind): boolean {
    if (!this.at(kind)) return false;
    this.advance();
    return true;
  }

  private expect(kind: TokenKind, message: string): FrontToken {
    if (this.at(kind)) return this.advance();
    const token = this.peek();
    this.error(message, token.span);
    return token;
  }

  private expectAny(kinds: TokenKind[], message: string): FrontToken {
    if (kinds.includes(this.peek().kind)) return this.advance();
    const token = this.peek();
    this.error(message, token.span);
    return token;
  }

  private at(kind: TokenKind): boolean {
    return this.peek().kind === kind;
  }

  private atAny(...kinds: TokenKind[]): boolean {
    const current = this.peek().kind;
    return kinds.includes(current);
  }

  private advance(): FrontToken {
    const token = this.peek();
    if (!this.at(TokenKind.EOF)) this.index += 1;
    return token;
  }

  private previous(): FrontToken {
    return this.tokens[Math.max(0, this.index - 1)] ?? this.peek();
  }

  private peek(ahead = 0): FrontToken {
    return this.tokens[this.index + ahead] ?? this.tokens[this.tokens.length - 1] ?? eofToken();
  }

  private error(message: string, span: SourceSpan | undefined, code = "GOJR_PARSE_FRONT001"): void {
    this.diagnostics.push({
      filename: diagnosticFilename(span, this.filename),
      code,
      severity: "error",
      message,
      ...(span ? { span } : {})
    });
  }
}

function binaryPrecedence(kind: TokenKind): number {
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

function badExpr(span?: SourceSpan): Expr {
  return { kind: "BadExpr", ...(span ? { span } : {}) };
}

function mergeSpans(start: SourceSpan | undefined, end: SourceSpan | undefined): SourceSpan {
  if (!start && !end) return eofToken().span;
  if (!start) return end ?? eofToken().span;
  if (!end) return start;
  const endOffset = end.offset + end.length;
  return {
    filename: start.filename,
    offset: start.offset,
    length: Math.max(0, endOffset - start.offset),
    line: start.line,
    column: start.column
  };
}

function tokenDisplay(kind: TokenKind): string {
  switch (kind) {
    case TokenKind.LParen:
      return "(";
    case TokenKind.RParen:
      return ")";
    case TokenKind.LBrace:
      return "{";
    case TokenKind.RBrace:
      return "}";
    case TokenKind.LBracket:
      return "[";
    case TokenKind.RBracket:
      return "]";
    default:
      return kind;
  }
}

function eofToken(): FrontToken {
  return {
    kind: TokenKind.EOF,
    lexeme: "",
    span: { filename: REPL_FILENAME, offset: 0, length: 0, line: 1, column: 1 }
  };
}
