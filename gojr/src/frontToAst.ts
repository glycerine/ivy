import {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  ConstDeclStatement,
  DeclarationSpec,
  DeferStatement,
  Expression,
  ExpressionStatement,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  IdentifierExpression,
  IfStatement,
  IncDecStatement,
  ImportDecl,
  IndexExpression,
  LiteralExpression,
  MapEntryExpression,
  MapLiteralExpression,
  ParameterDecl,
  ProgramAst,
  ReceiverDecl,
  ReturnStatement,
  SelectorExpression,
  ShortVarStatement,
  Signature,
  SliceExpression,
  SpreadsheetRangeExpression,
  Statement,
  StructFieldDecl,
  StructLiteralExpression,
  StructLiteralField,
  SwitchClause,
  SwitchStatement,
  TypeAssertionExpression,
  TypeExpression,
  TypeDeclStatement,
  TypeNode,
  TypeSpec,
  UnaryExpression,
  VarDeclStatement
} from "./ast.js";
import { Diagnostic, SourceSpan } from "./diagnostics.js";
import {
  ArrayType,
  BasicLit,
  BinaryOperator,
  BlockStmt,
  BranchStmt,
  CallExpr,
  CaseClause,
  CellRefExpr,
  CompositeLit,
  Decl,
  Expr,
  Field,
  FieldList,
  File,
  FuncDecl,
  FuncType,
  GenDecl,
  ImportSpec,
  KeyValueExpr,
  RangeRefExpr,
  RangeStmt,
  Spec,
  Stmt,
  TypeSpec as FrontTypeSpec,
  ValueSpec
} from "./front/ast.js";
import { ParseFrontResult, parseFrontSource } from "./front/parser.js";
import { TokenKind } from "./front/token.js";

export function frontSourceToAst(source: string): { ast?: ProgramAst; diagnostics: Diagnostic[]; parsed: ParseFrontResult } {
  const parsed = parseFrontSource(source);
  const ast = parsed.file ? frontToProgramAst(parsed.file, parsed.diagnostics, parsed.statements) : undefined;
  return {
    diagnostics: parsed.diagnostics,
    parsed,
    ...(ast ? { ast } : {})
  };
}

export function frontToProgramAst(file: File, diagnostics: Diagnostic[] = [], statements: Stmt[] = []): ProgramAst {
  const body: Statement[] = [
    ...file.declarations.flatMap(declarationToBodyStatement),
    ...statements.map(statementToAst)
  ];
  const functions = file.declarations.flatMap((declaration) =>
    declaration.kind === "FuncDecl" ? [functionDeclToAst(declaration)] : []
  );
  return {
    kind: functions.length > 0 && body.length === 0 ? "function" : "script",
    imports: file.imports.map(importSpecToAst),
    diagnostics,
    body,
    functions
  };
}

function declarationToBodyStatement(declaration: Decl): Statement[] {
  if (declaration.kind !== "GenDecl") return [];
  const statement = genDeclToStatement(declaration);
  return statement ? [statement] : [];
}

function importSpecToAst(spec: ImportSpec): ImportDecl {
  const path = unquote(spec.path.value);
  return {
    path,
    ...(spec.name ? { alias: spec.name.name } : {})
  };
}

function genDeclToStatement(declaration: GenDecl): ConstDeclStatement | VarDeclStatement | TypeDeclStatement | undefined {
  if (declaration.token === TokenKind.Const) {
    return withSpan({
      kind: "ConstDecl",
      declarations: declaration.specs.flatMap(valueSpecToDeclarations)
    } satisfies ConstDeclStatement, declaration.span);
  }
  if (declaration.token === TokenKind.Var) {
    return withSpan({
      kind: "VarDecl",
      declarations: declaration.specs.flatMap(valueSpecToDeclarations)
    } satisfies VarDeclStatement, declaration.span);
  }
  if (declaration.token === TokenKind.Type) {
    return withSpan({
      kind: "TypeDecl",
      declarations: declaration.specs.flatMap(typeSpecToAst)
    } satisfies TypeDeclStatement, declaration.span);
  }
  return undefined;
}

function valueSpecToDeclarations(spec: Spec): DeclarationSpec[] {
  if (spec.kind !== "ValueSpec") return [];
  return spec.names.map((name, index) => ({
    name: name.name,
    ...(spec.type ? { type: typeNode(spec.type) } : {}),
    ...(spec.values[index] ? { value: expressionToAst(spec.values[index]!) } : {})
  }));
}

function typeSpecToAst(spec: Spec): TypeSpec[] {
  if (spec.kind !== "TypeSpec") return [];
  return [{
    name: spec.name.name,
    type: typeNode(spec.type),
    ...structFieldsFromType(spec),
    ...interfaceMethodsFromType(spec)
  }];
}

function structFieldsFromType(spec: FrontTypeSpec): { structFields: StructFieldDecl[] } | {} {
  if (spec.type.kind !== "StructType") return {};
  return {
    structFields: spec.type.fields.fields.flatMap((field) =>
      field.names.map((name) => ({
        name: name.name,
        type: typeNode(field.type)
      }))
    )
  };
}

function interfaceMethodsFromType(spec: FrontTypeSpec): { interfaceMethods: NonNullable<TypeSpec["interfaceMethods"]> } | {} {
  if (spec.type.kind !== "InterfaceType") return {};
  return {
    interfaceMethods: spec.type.methods.fields.flatMap((field) => {
      const methodType = field.type;
      if (methodType.kind !== "FuncType") return [];
      return field.names.map((name) => ({
        name: name.name,
        signature: signatureToAst(methodType)
      }));
    })
  };
}

function functionDeclToAst(declaration: FuncDecl): FunctionDecl {
  return withSpan({
    kind: "FunctionDecl",
    name: declaration.name.name,
    ...(declaration.receiver ? { receiver: receiverToAst(declaration.receiver) } : {}),
    signature: signatureToAst(declaration.type),
    body: declaration.body ? blockToAst(declaration.body) : { kind: "BlockStatement", statements: [] }
  } satisfies FunctionDecl, declaration.span);
}

function receiverToAst(list: FieldList): ReceiverDecl {
  const field = list.fields[0];
  return {
    ...(field?.names[0] ? { name: field.names[0].name } : {}),
    type: field ? typeNode(field.type) : { text: "<missing>" }
  };
}

function signatureToAst(type: FuncType): Signature {
  return {
    parameters: parametersFromFields(type.params),
    results: type.results ? parametersFromFields(type.results) : []
  };
}

function parametersFromFields(list: FieldList): ParameterDecl[] {
  return list.fields.flatMap((field) => {
    const fieldType = field.type.kind === "Ellipsis" && field.type.element ? field.type.element : field.type;
    const base = {
      type: typeNode(fieldType),
      variadic: field.type.kind === "Ellipsis"
    };
    if (field.names.length === 0) return [base];
    return field.names.map((name) => ({
      name: name.name,
      ...base
    }));
  });
}

function blockToAst(block: BlockStmt): BlockStatement {
  return withSpan({
    kind: "BlockStatement",
    statements: block.statements.map(statementToAst)
  } satisfies BlockStatement, block.span);
}

function statementToAst(statement: Stmt): Statement {
  switch (statement.kind) {
    case "DeclStmt": {
      if (statement.decl.kind === "GenDecl") return genDeclToStatement(statement.decl) ?? expressionStatement(missingExpression(), statement.span);
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
        } satisfies ShortVarStatement, statement.span);
      }
      return withSpan({
        kind: "AssignStatement",
        targets: statement.lhs.map(expressionToAst),
        values: statement.rhs.map(expressionToAst)
      } satisfies AssignStatement, statement.span);
    case "IncDecStmt":
      return withSpan({
        kind: "IncDecStatement",
        target: expressionToAst(statement.expr),
        operator: statement.token === TokenKind.PlusPlus ? "++" : "--"
      } satisfies IncDecStatement, statement.span);
    case "ReturnStmt":
      return withSpan({
        kind: "ReturnStatement",
        values: statement.results.map(expressionToAst)
      } satisfies ReturnStatement, statement.span);
    case "BranchStmt":
      return branchToAst(statement);
    case "IfStmt":
      return withSpan({
        kind: "IfStatement",
        condition: expressionToAst(statement.condition),
        thenBlock: blockToAst(statement.body),
        ...(statement.else ? { elseBranch: elseBranchToAst(statement.else) } : {})
      } satisfies IfStatement, statement.span);
    case "ForStmt":
      return withSpan({
        kind: "ForStatement",
        ...(statement.init ? { init: statementToAst(statement.init) } : {}),
        ...(statement.condition ? { condition: expressionToAst(statement.condition) } : {}),
        ...(statement.post ? { post: statementToAst(statement.post) } : {}),
        body: blockToAst(statement.body)
      } satisfies ForStatement, statement.span);
    case "RangeStmt":
      return rangeStmtToAst(statement);
    case "SwitchStmt":
      return switchStmtToAst(statement);
    case "TypeSwitchStmt":
      return typeSwitchStmtToAst(statement);
    case "DeferStmt":
      return withSpan({
        kind: "DeferStatement",
        expression: expressionToAst(statement.call)
      } satisfies DeferStatement, statement.span);
    default:
      return expressionStatement(missingExpression(), statement.span);
  }
}

function elseBranchToAst(statement: Stmt): IfStatement | BlockStatement {
  if (statement.kind === "IfStmt") return statementToAst(statement) as IfStatement;
  if (statement.kind === "BlockStmt") return blockToAst(statement);
  return { kind: "BlockStatement", statements: [statementToAst(statement)] };
}

function branchToAst(statement: BranchStmt): BranchStatement {
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
  } satisfies BranchStatement, statement.span);
}

function rangeStmtToAst(statement: RangeStmt): ForStatement {
  return withSpan({
    kind: "ForStatement",
    range: {
      ...(statement.key?.kind === "Ident" ? { keyName: statement.key.name } : {}),
      ...(statement.value?.kind === "Ident" ? { valueName: statement.value.name } : {}),
      define: statement.token === TokenKind.Define,
      source: expressionToAst(statement.source)
    },
    body: blockToAst(statement.body)
  } satisfies ForStatement, statement.span);
}

function switchStmtToAst(statement: Extract<Stmt, { kind: "SwitchStmt" }>): SwitchStatement {
  return withSpan({
    kind: "SwitchStatement",
    ...(statement.tag ? { expression: expressionToAst(statement.tag) } : {}),
    clauses: statement.body.map((clause) => caseClauseToAst(clause, false))
  } satisfies SwitchStatement, statement.span);
}

function typeSwitchStmtToAst(statement: Extract<Stmt, { kind: "TypeSwitchStmt" }>): SwitchStatement {
  return withSpan({
    kind: "SwitchStatement",
    typeSwitch: typeSwitchGuardToAst(statement.assign),
    clauses: statement.body.map((clause) => caseClauseToAst(clause, true))
  } satisfies SwitchStatement, statement.span);
}

function typeSwitchGuardToAst(statement: Stmt) {
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

function caseClauseToAst(clause: CaseClause, typeSwitch: boolean): SwitchClause {
  return {
    kind: "SwitchClause",
    values: typeSwitch ? [] : clause.list.map(expressionToAst),
    ...(typeSwitch ? { typeValues: clause.list.map(typeNode) } : {}),
    default: clause.default,
    statements: clause.body.map(statementToAst)
  };
}

function expressionStatement(expression: Expression, span?: SourceSpan): ExpressionStatement {
  return withSpan({
    kind: "ExpressionStatement",
    expression
  } satisfies ExpressionStatement, span);
}

function expressionToAst(expr: Expr): Expression {
  switch (expr.kind) {
    case "Ident":
      if (expr.name === "true" || expr.name === "false") return literal(expr.name === "true", "bool", expr.name, expr.span);
      if (expr.name === "nil") return literal(null, "nil", "nil", expr.span);
      return withSpan({ kind: "Identifier", name: expr.name } satisfies IdentifierExpression, expr.span);
    case "BasicLit":
      return basicLitToAst(expr);
    case "FuncLit":
      return withSpan({
        kind: "FunctionLiteralExpression",
        signature: signatureToAst(expr.type),
        body: blockToAst(expr.body)
      } satisfies FunctionLiteralExpression, expr.span);
    case "CompositeLit":
      return compositeLitToAst(expr);
    case "ParenExpr":
      return expressionToAst(expr.expr);
    case "SelectorExpr":
      return withSpan({
        kind: "SelectorExpression",
        object: expressionToAst(expr.object),
        field: expr.selector.name
      } satisfies SelectorExpression, expr.span);
    case "CellRefExpr":
      return cellRefToSelector(expr);
    case "RangeRefExpr":
      return rangeRefToAst(expr);
    case "IndexExpr":
      return withSpan({
        kind: "IndexExpression",
        object: expressionToAst(expr.object),
        index: expressionToAst(expr.index)
      } satisfies IndexExpression, expr.span);
    case "SliceExpr":
      return withSpan({
        kind: "SliceExpression",
        object: expressionToAst(expr.object),
        ...(expr.low ? { start: expressionToAst(expr.low) } : {}),
        ...(expr.high ? { end: expressionToAst(expr.high) } : {})
      } satisfies SliceExpression, expr.span);
    case "TypeAssertExpr":
      return withSpan({
        kind: "TypeAssertionExpression",
        expression: expressionToAst(expr.object),
        type: expr.type ? typeNode(expr.type) : { text: "type" }
      } satisfies TypeAssertionExpression, expr.span);
    case "CallExpr":
      return withSpan({
        kind: "CallExpression",
        callee: expressionToAst(expr.fun),
        args: expr.args.map(expressionToAst),
        spreadLast: expr.ellipsis
      } satisfies CallExpression, expr.span);
    case "StarExpr":
      return withSpan({
        kind: "UnaryExpression",
        operator: "*",
        operand: expressionToAst(expr.expr)
      } satisfies UnaryExpression, expr.span);
    case "UnaryExpr":
      return withSpan({
        kind: "UnaryExpression",
        operator: unaryOperator(expr.op),
        operand: expressionToAst(expr.expr)
      } satisfies UnaryExpression, expr.span);
    case "BinaryExpr":
      return withSpan({
        kind: "BinaryExpression",
        operator: binaryOperator(expr.op),
        left: expressionToAst(expr.left),
        right: expressionToAst(expr.right)
      } satisfies BinaryExpression, expr.span);
    case "ArrayType":
    case "MapType":
    case "StructType":
    case "InterfaceType":
    case "FuncType":
      return withSpan({
        kind: "TypeExpression",
        type: typeNode(expr)
      } satisfies TypeExpression, expr.span);
    default:
      return missingExpression(expr.span);
  }
}

function basicLitToAst(expr: BasicLit): LiteralExpression {
  if (expr.token === TokenKind.IntLiteral) return literal(BigInt(expr.value), "int", expr.value, expr.span);
  if (expr.token === TokenKind.FloatLiteral) return literal(Number(expr.value), "float", expr.value, expr.span);
  return literal(unquote(expr.value), "string", expr.value, expr.span);
}

function literal(value: LiteralExpression["value"], literalKind: LiteralExpression["literalKind"], raw: string, span?: SourceSpan): LiteralExpression {
  return withSpan({
    kind: "Literal",
    literalKind,
    value,
    raw
  } satisfies LiteralExpression, span);
}

function compositeLitToAst(expr: CompositeLit): Expression {
  const type = expr.type;
  if (type?.kind === "MapType") {
    return withSpan({
      kind: "MapLiteralExpression",
      keyType: typeNode(type.key),
      valueType: typeNode(type.value),
      entries: expr.elements.flatMap(mapEntryToAst)
    } satisfies MapLiteralExpression, expr.span);
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
  } satisfies StructLiteralExpression, expr.span);
}

function mapEntryToAst(expr: Expr): MapEntryExpression[] {
  if (expr.kind !== "KeyValueExpr") return [];
  return [{ key: expressionToAst(expr.key), value: expressionToAst(expr.value) }];
}

function structFieldToAst(expr: Expr): StructLiteralField {
  if (expr.kind === "KeyValueExpr") {
    return {
      ...(expr.key.kind === "Ident" ? { name: expr.key.name } : {}),
      value: expressionToAst(expr.value)
    };
  }
  return { value: expressionToAst(expr) };
}

function elementValueToAst(expr: Expr): Expression {
  return expr.kind === "KeyValueExpr" ? expressionToAst(expr.value) : expressionToAst(expr);
}

function cellRefToSelector(expr: CellRefExpr): SelectorExpression {
  return withSpan({
    kind: "SelectorExpression",
    object: { kind: "Identifier", name: expr.namespace.name },
    field: expr.address.raw
  } satisfies SelectorExpression, expr.span);
}

function rangeRefToAst(expr: RangeRefExpr): SpreadsheetRangeExpression {
  return withSpan({
    kind: "SpreadsheetRangeExpression",
    start: withSpan({
      kind: "SelectorExpression",
      object: { kind: "Identifier", name: expr.namespace.name },
      field: expr.start.raw
    } satisfies SelectorExpression, expr.span),
    endCell: expr.end.raw
  } satisfies SpreadsheetRangeExpression, expr.span);
}

function shortVarName(expr: Expr): string {
  return expr.kind === "Ident" ? expr.name : "<invalid>";
}

function typeNode(expr: Expr): TypeNode {
  return withSpan({ text: typeText(expr) }, expr.span);
}

function typeText(expr: Expr): string {
  switch (expr.kind) {
    case "Ident":
      return expr.name;
    case "SelectorExpr":
      return `${typeText(expr.object)}.${expr.selector.name}`;
    case "StarExpr":
      return `*${typeText(expr.expr)}`;
    case "ArrayType":
      return `${arrayLengthText(expr)}${typeText(expr.element)}`;
    case "MapType":
      return `map[${typeText(expr.key)}]${typeText(expr.value)}`;
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

function arrayLengthText(expr: ArrayType): string {
  if (expr.inferredLength) return "[...]";
  return expr.length ? `[${expressionText(expr.length)}]` : "[]";
}

function fieldsText(fields: FieldList): string {
  return fields.fields.map((field) => {
    const names = field.names.map((name) => name.name).join(", ");
    return `${names ? `${names} ` : ""}${typeText(field.type)}`;
  }).join("; ");
}

function interfaceText(fields: FieldList): string {
  return fields.fields.map((field) => {
    const names = field.names.map((name) => name.name).join(", ");
    return `${names}${field.type.kind === "FuncType" ? `(${paramsText(field.type.params)})${resultsText(field.type.results)}` : ` ${typeText(field.type)}`}`;
  }).join("; ");
}

function paramsText(fields: FieldList): string {
  return fields.fields.map(fieldText).join(", ");
}

function fieldText(field: Field): string {
  const names = field.names.map((name) => name.name).join(", ");
  return `${names ? `${names} ` : ""}${typeText(field.type)}`;
}

function resultsText(results: FieldList | undefined): string {
  if (!results || results.fields.length === 0) return "";
  if (results.fields.length === 1 && results.fields[0]?.names.length === 0) return ` ${typeText(results.fields[0].type)}`;
  return ` (${paramsText(results)})`;
}

function expressionText(expr: Expr): string {
  if (expr.kind === "BasicLit") return expr.value;
  if (expr.kind === "Ident") return expr.name;
  return typeText(expr);
}

function unaryOperator(kind: TokenKind): UnaryExpression["operator"] {
  if (kind === TokenKind.Minus) return "-";
  if (kind === TokenKind.Bang) return "!";
  if (kind === TokenKind.Amp) return "&";
  if (kind === TokenKind.Star) return "*";
  return "+";
}

function binaryOperator(kind: BinaryOperator): BinaryExpression["operator"] {
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
    case TokenKind.Star: return "*";
    case TokenKind.Slash: return "/";
    case TokenKind.Percent: return "%";
    default: return "+";
  }
}

function missingExpression(span?: SourceSpan): IdentifierExpression {
  return withSpan({ kind: "Identifier", name: "<missing>" }, span);
}

function withSpan<T extends object>(node: T, span?: SourceSpan): T {
  return span ? { ...node, span } : node;
}

function unquote(value: string): string {
  if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
    return value.slice(1, -1).replace(/\r/g, "");
  }
  if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
    try {
      return JSON.parse(value) as string;
    } catch {
      return value.slice(1, -1);
    }
  }
  return value;
}
