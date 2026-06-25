import {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  CommClause,
  ConstDeclStatement,
  DeclarationSpec,
  DeferStatement,
  Expression,
  ExpressionStatement,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  GoStatement,
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
  SelectStatement,
  SendStatement,
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
import { Diagnostic, SourceFile, SourceSpan } from "./diagnostics.js";
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
  StructType,
  TypeSpec as FrontTypeSpec,
  ValueSpec
} from "./front/ast.js";
import { ParseFrontFilesResult, ParseFrontResult, parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { Node as formatNode } from "./go/format.js";

export function frontSourceToAst(source: string, filename: string): { ast?: ProgramAst; diagnostics: Diagnostic[]; parsed: ParseFrontResult } {
  const parsed = parseFrontSource(source, filename);
  const ast = parsed.file ? frontToProgramAst(parsed.file, parsed.diagnostics, parsed.statements) : undefined;
  return {
    diagnostics: parsed.diagnostics,
    parsed,
    ...(ast ? { ast } : {})
  };
}

export function frontSourceFilesToAst(files: SourceFile[]): { ast?: ProgramAst; diagnostics: Diagnostic[]; parsed: ParseFrontFilesResult } {
  const parsed = parseFrontSourceFiles(files);
  const ast = parsed.files.length > 0 ? frontFilesToProgramAst(parsed.files, parsed.diagnostics, parsed.statements) : undefined;
  return {
    diagnostics: parsed.diagnostics,
    parsed,
    ...(ast ? { ast } : {})
  };
}

export function frontToProgramAst(file: File, diagnostics: Diagnostic[] = [], statements: Stmt[] = []): ProgramAst {
  return frontFilesToProgramAst([file], diagnostics, statements);
}

export function frontFilesToProgramAst(files: File[], diagnostics: Diagnostic[] = [], statements: Stmt[] = []): ProgramAst {
  const body: Statement[] = [
    ...files.flatMap((file) => file.declarations.flatMap(declarationToBodyStatement)),
    ...statements.map(statementToAst)
  ];
  const functions = files.flatMap((file) => file.declarations).flatMap((declaration) =>
    declaration.kind === "FuncDecl" ? [functionDeclToAst(declaration)] : []
  );
  return {
    kind: functions.length > 0 && body.length === 0 ? "function" : "script",
    imports: files.flatMap((file) => file.imports.map(importSpecToAst)),
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
    ...(spec.name ? { alias: spec.name.name } : {}),
    ...(spec.span ? { span: spec.span } : {})
  };
}

function genDeclToStatement(declaration: GenDecl): ConstDeclStatement | VarDeclStatement | TypeDeclStatement | undefined {
  if (declaration.token === TokenKind.Const) {
    return withSpan({
      kind: "ConstDecl",
      declarations: declaration.specs.flatMap((spec, index) => valueSpecToDeclarations(spec, index, index))
    } satisfies ConstDeclStatement, declaration.span);
  }
  if (declaration.token === TokenKind.Var) {
    return withSpan({
      kind: "VarDecl",
      declarations: declaration.specs.flatMap((spec, index) => valueSpecToDeclarations(spec, undefined, index))
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

function valueSpecToDeclarations(spec: Spec, iotaIndex?: number, valueGroup?: number): DeclarationSpec[] {
  if (spec.kind !== "ValueSpec") return [];
  const values = spec.values.map((value) => expressionToAst(value));
  const sharedSingleValue = values.length === 1 && spec.names.length > 1 ? values[0] : undefined;
  return spec.names.map((name, index) => ({
    name: name.name,
    ...(spec.type ? { type: typeNode(spec.type) } : {}),
    ...((values[index] ?? sharedSingleValue) ? { value: (values[index] ?? sharedSingleValue)! } : {}),
    ...(valueGroup !== undefined ? { valueGroup } : {}),
    ...(iotaIndex !== undefined ? { iotaIndex } : {}),
    valueIndex: index,
    valueCount: values.length,
    groupNameCount: spec.names.length
  }));
}

function typeSpecToAst(spec: Spec): TypeSpec[] {
  if (spec.kind !== "TypeSpec") return [];
  const typeParameters = typeSpecTypeParameterNames(spec);
  return [{
    name: spec.name.name,
    ...(spec.alias ? { alias: true } : {}),
    ...(typeParameters.length > 0 ? { typeParameters } : {}),
    type: typeNode(spec.type),
    ...structFieldsFromType(spec),
    ...interfaceMethodsFromType(spec)
  }];
}

function typeSpecTypeParameterNames(spec: FrontTypeSpec): string[] {
  const names = new Set<string>();
  addFieldListNames(spec.typeParams, names);
  return [...names];
}

function structFieldsFromType(spec: FrontTypeSpec): { structFields: StructFieldDecl[] } | {} {
  if (spec.type.kind !== "StructType") return {};
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

function interfaceMethodsFromType(spec: FrontTypeSpec): Pick<TypeSpec, "interfaceMethods" | "interfaceEmbeds"> | {} {
  if (spec.type.kind !== "InterfaceType") return {};
  const interfaceMethods: NonNullable<TypeSpec["interfaceMethods"]> = [];
  const interfaceEmbeds: TypeNode[] = [];
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

function functionDeclToAst(declaration: FuncDecl): FunctionDecl {
  const typeParameters = functionTypeParameterNames(declaration);
  return withSpan({
    kind: "FunctionDecl",
    name: declaration.name.name,
    ...(declaration.receiver ? { receiver: receiverToAst(declaration.receiver) } : {}),
    ...(typeParameters.length > 0 ? { typeParameters } : {}),
    signature: signatureToAst(declaration.type),
    body: declaration.body ? blockToAst(declaration.body) : { kind: "BlockStatement", statements: [] },
    source: formatNode(declaration).trimEnd()
  } satisfies FunctionDecl, declaration.span);
}

function functionTypeParameterNames(declaration: FuncDecl): string[] {
  const names = new Set<string>();
  addFieldListNames(declaration.type.typeParams, names);
  if (declaration.receiver) addReceiverTypeParameterNames(declaration.receiver, names);
  return [...names];
}

function addFieldListNames(fields: FieldList | undefined, names: Set<string>): void {
  if (!fields) return;
  for (const field of fields.fields) {
    for (const name of field.names) {
      if (name.name !== "_") names.add(name.name);
    }
  }
}

function addReceiverTypeParameterNames(receiver: FieldList, names: Set<string>): void {
  const field = receiver.fields[0];
  if (!field) return;
  collectReceiverTypeParameterNames(field.type, names);
}

function collectReceiverTypeParameterNames(expr: Expr, names: Set<string>): void {
  switch (expr.kind) {
    case "StarExpr":
      collectReceiverTypeParameterNames(expr.expr, names);
      return;
    case "IndexExpr":
      addReceiverTypeArgumentName(expr.index, names);
      return;
    case "IndexListExpr":
      for (const index of expr.indices) addReceiverTypeArgumentName(index, names);
      return;
    case "ParenExpr":
      collectReceiverTypeParameterNames(expr.expr, names);
      return;
    default:
      return;
  }
}

function addReceiverTypeArgumentName(expr: Expr, names: Set<string>): void {
  if (expr.kind === "Ident" && expr.name !== "_") names.add(expr.name);
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
        operator: assignmentOperator(statement.token),
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
        ...(statement.init ? { init: statementToAst(statement.init) } : {}),
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
    case "SelectStmt":
      return selectStmtToAst(statement);
    case "DeferStmt":
      return withSpan({
        kind: "DeferStatement",
        expression: expressionToAst(statement.call)
      } satisfies DeferStatement, statement.span);
    case "GoStmt":
      return withSpan({
        kind: "GoStatement",
        call: expressionToAst(statement.call) as CallExpression
      } satisfies GoStatement, statement.span);
    case "SendStmt":
      return withSpan({
        kind: "SendStatement",
        channel: expressionToAst(statement.channel),
        value: expressionToAst(statement.value)
      } satisfies SendStatement, statement.span);
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
      ...(statement.key && statement.key.kind !== "Ident" ? { keyTarget: expressionToAst(statement.key) } : {}),
      ...(statement.value?.kind === "Ident" ? { valueName: statement.value.name } : {}),
      ...(statement.value && statement.value.kind !== "Ident" ? { valueTarget: expressionToAst(statement.value) } : {}),
      define: statement.token === TokenKind.Define,
      source: expressionToAst(statement.source)
    },
    body: blockToAst(statement.body)
  } satisfies ForStatement, statement.span);
}

function switchStmtToAst(statement: Extract<Stmt, { kind: "SwitchStmt" }>): SwitchStatement {
  return withSpan({
    kind: "SwitchStatement",
    ...(statement.init ? { init: statementToAst(statement.init) } : {}),
    ...(statement.tag ? { expression: expressionToAst(statement.tag) } : {}),
    clauses: statement.body.map((clause) => caseClauseToAst(clause, false))
  } satisfies SwitchStatement, statement.span);
}

function typeSwitchStmtToAst(statement: Extract<Stmt, { kind: "TypeSwitchStmt" }>): SwitchStatement {
  return withSpan({
    kind: "SwitchStatement",
    ...(statement.init ? { init: statementToAst(statement.init) } : {}),
    typeSwitch: typeSwitchGuardToAst(statement.assign),
    clauses: statement.body.map((clause) => caseClauseToAst(clause, true))
  } satisfies SwitchStatement, statement.span);
}

function selectStmtToAst(statement: Extract<Stmt, { kind: "SelectStmt" }>): SelectStatement {
  return withSpan({
    kind: "SelectStatement",
    clauses: statement.body.map(commClauseToAst)
  } satisfies SelectStatement, statement.span);
}

function commClauseToAst(clause: Extract<Stmt, { kind: "CommClause" }>): CommClause {
  return {
    kind: "CommClause",
    ...(clause.comm ? { comm: statementToAst(clause.comm) } : {}),
    default: clause.default,
    statements: clause.body.map(statementToAst)
  };
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
        body: blockToAst(expr.body),
        source: formatNode(expr).trimEnd()
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
    case "IndexListExpr":
      return withSpan({
        kind: "IndexExpression",
        object: expressionToAst(expr.object),
        index: withSpan({
          kind: "TypeExpression",
          type: { text: expr.indices.map(typeText).join(", ") }
        } satisfies TypeExpression, expr.span)
      } satisfies IndexExpression, expr.span);
    case "SliceExpr":
      return withSpan({
        kind: "SliceExpression",
        object: expressionToAst(expr.object),
        ...(expr.low ? { start: expressionToAst(expr.low) } : {}),
        ...(expr.high ? { end: expressionToAst(expr.high) } : {}),
        ...(expr.max ? { max: expressionToAst(expr.max) } : {})
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
    case "ChanType":
      return withSpan({
        kind: "TypeExpression",
        type: typeNode(expr)
      } satisfies TypeExpression, expr.span);
    default:
      return missingExpression(expr.span);
  }
}

function basicLitToAst(expr: BasicLit): LiteralExpression {
  if (expr.token === TokenKind.IntLiteral) return literal(parseGoIntLiteral(expr.value), "int", expr.value, expr.span);
  if (expr.token === TokenKind.FloatLiteral) return literal(parseGoFloatLiteral(expr.value), "float", expr.value, expr.span);
  if (expr.token === TokenKind.ImagLiteral) {
    const raw = expr.value.slice(0, -1);
    const imag = /[.eEpP]/.test(raw) ? parseGoFloatLiteral(raw) : Number(parseGoIntLiteral(raw));
    return literal({ real: 0, imag }, "imag", expr.value, expr.span);
  }
  if (expr.token === TokenKind.RuneLiteral) return literal(parseGoRuneLiteral(expr.value), "rune", expr.value, expr.span);
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

function compositeLitToAst(expr: CompositeLit, expectedType?: Expr): Expression {
  const type = expr.type ?? expectedType;
  if (type?.kind === "MapType") {
    return withSpan({
      kind: "MapLiteralExpression",
      keyType: typeNode(type.key),
      valueType: typeNode(type.value),
      entries: expr.elements.flatMap((element) => mapEntryToAst(element, type.key, type.value))
    } satisfies MapLiteralExpression, expr.span);
  }
  if (type?.kind === "ArrayType") {
    return withSpan({
      kind: "ArrayLiteralExpression",
      type: typeNode(type),
      elements: expr.elements.map((element) => arrayElementToAst(element, type.element))
    }, expr.span);
  }
  const structFieldTypes = type?.kind === "StructType" ? expandedStructFieldTypes(type) : [];
  return withSpan({
    kind: "StructLiteralExpression",
    typeName: type ? typeText(type) : "<missing>",
    fields: expr.elements.map((element, index) => structFieldToAst(element, structFieldTypeForElement(element, index, structFieldTypes)))
  } satisfies StructLiteralExpression, expr.span);
}

function mapEntryToAst(expr: Expr, keyType?: Expr, valueType?: Expr): MapEntryExpression[] {
  if (expr.kind !== "KeyValueExpr") return [];
  return [{ key: expressionToAstWithExpectedType(expr.key, keyType), value: expressionToAstWithExpectedType(expr.value, valueType) }];
}

function structFieldToAst(expr: Expr, expectedType?: Expr): StructLiteralField {
  if (expr.kind === "KeyValueExpr") {
    return {
      ...(expr.key.kind === "Ident" ? { name: expr.key.name } : {}),
      key: expressionToAst(expr.key),
      value: expressionToAstWithExpectedType(expr.value, expectedType)
    };
  }
  return { value: expressionToAstWithExpectedType(expr, expectedType) };
}

function elementValueToAst(expr: Expr, expectedType?: Expr): Expression {
  return expr.kind === "KeyValueExpr" ? expressionToAstWithExpectedType(expr.value, expectedType) : expressionToAstWithExpectedType(expr, expectedType);
}

function arrayElementToAst(expr: Expr, expectedType?: Expr) {
  if (expr.kind === "KeyValueExpr") {
    return {
      key: expressionToAst(expr.key),
      value: expressionToAstWithExpectedType(expr.value, expectedType)
    };
  }
  return { value: expressionToAstWithExpectedType(expr, expectedType) };
}

function expressionToAstWithExpectedType(expr: Expr, expectedType?: Expr): Expression {
  return expr.kind === "CompositeLit" && !expr.type ? compositeLitToAst(expr, expectedType) : expressionToAst(expr);
}

function expandedStructFieldTypes(type: StructType): Array<{ name?: string; type: Expr }> {
  const fields: Array<{ name?: string; type: Expr }> = [];
  for (const field of type.fields.fields) {
    if (field.names.length === 0) {
      fields.push({ type: field.type });
      continue;
    }
    for (const name of field.names) {
      fields.push({ name: name.name, type: field.type });
    }
  }
  return fields;
}

function structFieldTypeForElement(expr: Expr, index: number, fields: Array<{ name?: string; type: Expr }>): Expr | undefined {
  if (expr.kind === "KeyValueExpr" && expr.key.kind === "Ident") {
    const keyName = expr.key.name;
    return fields.find((field) => field.name === keyName)?.type;
  }
  return fields[index]?.type;
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
    case "IndexExpr":
      return `${typeText(expr.object)}[${typeText(expr.index)}]`;
    case "IndexListExpr":
      return `${typeText(expr.object)}[${expr.indices.map(typeText).join(", ")}]`;
    case "StarExpr":
      return `*${typeText(expr.expr)}`;
    case "UnaryExpr":
      if (expr.op === TokenKind.Tilde) return `~${typeText(expr.expr)}`;
      return `${unaryOperator(expr.op)}${typeText(expr.expr)}`;
    case "BinaryExpr":
      if (expr.op === TokenKind.Or) return `${typeText(expr.left)} | ${typeText(expr.right)}`;
      return `${typeText(expr.left)} ${binaryOperator(expr.op)} ${typeText(expr.right)}`;
    case "ArrayType":
      return `${arrayLengthText(expr)}${typeText(expr.element)}`;
    case "MapType":
      return `map[${typeText(expr.key)}]${typeText(expr.value)}`;
    case "ChanType":
      if (expr.direction === "send") return `chan<- ${typeText(expr.value)}`;
      if (expr.direction === "receive") return `<-chan ${typeText(expr.value)}`;
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
    case "ParenExpr":
      return `(${typeText(expr.expr)})`;
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
    const tag = field.tag ? ` ${field.tag.value}` : "";
    return `${names ? `${names} ` : ""}${typeText(field.type)}${tag}`;
  }).join("; ");
}

function interfaceText(fields: FieldList): string {
  return fields.fields.map((field) => {
    const names = field.names.map((name) => name.name).join(", ");
    if (!names && field.type.kind !== "FuncType") return typeText(field.type);
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
  if (expr.kind === "ParenExpr") return `(${expressionText(expr.expr)})`;
  return typeText(expr);
}

function unaryOperator(kind: TokenKind): UnaryExpression["operator"] {
  if (kind === TokenKind.Arrow) return "<-";
  if (kind === TokenKind.Minus) return "-";
  if (kind === TokenKind.Bang) return "!";
  if (kind === TokenKind.Caret) return "^";
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

function assignmentOperator(kind: TokenKind): AssignStatement["operator"] {
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

function parseGoIntLiteral(value: string): bigint {
  const text = value.replace(/_/g, "");
  if (/^0[0-7]+$/.test(text)) return BigInt(`0o${text.slice(1)}`);
  return BigInt(text);
}

function parseGoFloatLiteral(value: string): number {
  const text = value.replace(/_/g, "");
  const hex = /^0[xX]([0-9a-fA-F]*)(?:\.([0-9a-fA-F]*))?[pP]([+-]?[0-9]+)$/.exec(text);
  if (!hex) return Number(text);
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

function parseGoRuneLiteral(value: string): bigint {
  const body = value.slice(1, -1);
  const decoded = decodeGoEscaped(body);
  return BigInt([...decoded][0]?.codePointAt(0) ?? 0);
}

function embeddedFieldName(expr: Expr): string {
  if (expr.kind === "Ident") return expr.name;
  if (expr.kind === "SelectorExpr") return expr.selector.name;
  if (expr.kind === "IndexExpr") return embeddedFieldName(expr.object);
  if (expr.kind === "IndexListExpr") return embeddedFieldName(expr.object);
  if (expr.kind === "StarExpr") return embeddedFieldName(expr.expr);
  return "";
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
    return decodeGoEscaped(value.slice(1, -1));
  }
  return value;
}

function decodeGoEscaped(value: string): string {
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
      case "a": decoded += "\x07"; break;
      case "b": decoded += "\b"; break;
      case "f": decoded += "\f"; break;
      case "n": decoded += "\n"; break;
      case "r": decoded += "\r"; break;
      case "t": decoded += "\t"; break;
      case "v": decoded += "\x0b"; break;
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
        } else {
          decoded += next;
        }
        break;
    }
  }
  return decoded;
}

function codePointFromEscape(digits: string, radix: number): string {
  const value = Number.parseInt(digits, radix);
  if (!Number.isFinite(value)) return "";
  try {
    return String.fromCodePoint(value);
  } catch {
    return "";
  }
}
