import type {
  AstNode,
  ArrayType,
  BasicLit,
  BinaryOperator,
  BlockStmt,
  BranchStmt,
  CallExpr,
  CaseClause,
  ChanType,
  CompositeLit,
  Decl,
  Expr,
  Field,
  FieldList,
  File,
  FuncDecl,
  FuncLit,
  FuncType,
  GenDecl,
  ImportSpec,
  KeyValueExpr,
  RangeStmt,
  Spec,
  Stmt,
  TypeSpec,
  ValueSpec
} from "../front/ast.js";
import { parseFrontSource } from "../front/parser.js";
import { TokenKind } from "../front/token.js";

const INDENT = "\t";

export function Source(source: string, filename = "gojr-format.go"): string {
  const trimmed = source.trim();
  if (!trimmed) return "";
  const parsed = parseFrontSource(trimmed, filename);
  if (parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error") || !parsed.file) {
    return ensureTrailingNewline(trimmed);
  }
  const formatted = formatFile(parsed.file, parsed.statements);
  return ensureTrailingNewline(formatted || trimmed);
}

export function Node(node: AstNode): string {
  switch (node.kind) {
    case "File":
      return formatFile(node, "");
    case "FuncDecl":
      return formatFuncDecl(node, 0);
    case "FuncLit":
      return formatFuncLit(node, 0);
    case "GenDecl":
      return formatGenDecl(node, 0);
    case "BlockStmt":
      return formatBlock(node, 0);
    default:
      if (isStmt(node)) return formatStmt(node, 0);
      if (isExpr(node)) return formatExpr(node);
      return "";
  }
}

function formatFile(file: File, statements: Stmt[] | ""): string {
  const parts: string[] = [];
  if (file.name) parts.push(`package ${file.name.name}`);
  for (const declaration of file.declarations) parts.push(formatDecl(declaration, 0));
  if (Array.isArray(statements)) {
    for (const statement of statements) parts.push(formatStmt(statement, 0));
  }
  return parts.filter(Boolean).join("\n\n");
}

function formatDecl(declaration: Decl, depth: number): string {
  switch (declaration.kind) {
    case "FuncDecl":
      return formatFuncDecl(declaration, depth);
    case "GenDecl":
      return formatGenDecl(declaration, depth);
    default:
      return "";
  }
}

function formatFuncDecl(declaration: FuncDecl, depth: number): string {
  const receiver = declaration.receiver ? `(${formatFieldListContent(declaration.receiver)}) ` : "";
  const header = `func ${receiver}${declaration.name.name}${formatFuncSignature(declaration.type)}`;
  return declaration.body ? `${header} ${formatBlock(declaration.body, depth)}` : header;
}

function formatFuncLit(expression: FuncLit, depth: number): string {
  return `func${formatFuncSignature(expression.type)} ${formatBlock(expression.body, depth)}`;
}

function formatFuncSignature(type: FuncType): string {
  const typeParams = type.typeParams ? `[${formatFieldListContent(type.typeParams)}]` : "";
  return `${typeParams}(${formatFieldListContent(type.params)})${formatResults(type.results)}`;
}

function formatResults(results: FieldList | undefined): string {
  if (!results || results.fields.length === 0) return "";
  if (results.fields.length === 1 && results.fields[0]?.names.length === 0) {
    return ` ${formatExpr(results.fields[0].type)}`;
  }
  return ` (${formatFieldListContent(results)})`;
}

function formatGenDecl(declaration: GenDecl, depth: number): string {
  const keyword = tokenKeyword(declaration.token);
  if (!declaration.grouped && declaration.specs.length === 1 && declaration.specs[0]) {
    return `${keyword} ${formatSpec(declaration.specs[0])}`;
  }
  const specs = declaration.specs.map((spec) => `${indent(depth + 1)}${formatSpec(spec)}`).join("\n");
  return `${keyword} (\n${specs}\n${indent(depth)})`;
}

function formatSpec(spec: Spec): string {
  switch (spec.kind) {
    case "ImportSpec":
      return formatImportSpec(spec);
    case "TypeSpec":
      return formatTypeSpec(spec);
    case "ValueSpec":
      return formatValueSpec(spec);
  }
}

function formatImportSpec(spec: ImportSpec): string {
  return `${spec.name ? `${spec.name.name} ` : ""}${spec.path.value}`;
}

function formatTypeSpec(spec: TypeSpec): string {
  const typeParams = spec.typeParams ? `[${formatFieldListContent(spec.typeParams)}]` : "";
  return `${spec.name.name}${typeParams}${spec.alias ? " =" : ""} ${formatExpr(spec.type)}`;
}

function formatValueSpec(spec: ValueSpec): string {
  const names = spec.names.map((name) => name.name).join(", ");
  const type = spec.type ? ` ${formatExpr(spec.type)}` : "";
  const values = spec.values.length > 0 ? ` = ${spec.values.map((value) => formatExpr(value)).join(", ")}` : "";
  return `${names}${type}${values}`;
}

function formatBlock(block: BlockStmt, depth: number): string {
  if (block.statements.length === 0) return "{}";
  const statements = block.statements
    .map((statement) => formatStmt(statement, depth + 1))
    .filter((line) => line.length > 0)
    .map((line) => `${indent(depth + 1)}${line}`)
    .join("\n");
  return `{\n${statements}\n${indent(depth)}}`;
}

function formatStmt(statement: Stmt, depth: number): string {
  switch (statement.kind) {
    case "DeclStmt":
      return formatDecl(statement.decl, depth);
    case "EmptyStmt":
      return "";
    case "LabeledStmt": {
      const body = formatStmt(statement.stmt, depth);
      return `${statement.label.name}:\n${indent(depth)}${body}`;
    }
    case "ExprStmt":
      return formatExpr(statement.expr);
    case "AssignStmt":
      return `${statement.lhs.map((expr) => formatExpr(expr)).join(", ")} ${assignmentOperator(statement.token)} ${statement.rhs.map((expr) => formatExpr(expr)).join(", ")}`;
    case "IncDecStmt":
      return `${formatExpr(statement.expr)}${statement.token === TokenKind.PlusPlus ? "++" : "--"}`;
    case "ReturnStmt":
      return statement.results.length === 0 ? "return" : `return ${statement.results.map((expr) => formatExpr(expr)).join(", ")}`;
    case "BranchStmt":
      return `${branchKeyword(statement)}${statement.label ? ` ${statement.label.name}` : ""}`;
    case "BlockStmt":
      return formatBlock(statement, depth);
    case "IfStmt":
      return formatIfStmt(statement, depth);
    case "ForStmt":
      return formatForStmt(statement, depth);
    case "RangeStmt":
      return formatRangeStmt(statement, depth);
    case "SwitchStmt":
      return formatSwitchStmt(statement, depth);
    case "TypeSwitchStmt":
      return formatTypeSwitchStmt(statement, depth);
    case "SelectStmt":
      return `select ${formatClauses(statement.body, depth, formatCommClause)}`;
    case "DeferStmt":
      return `defer ${formatExpr(statement.call)}`;
    case "GoStmt":
      return `go ${formatExpr(statement.call)}`;
    case "SendStmt":
      return `${formatExpr(statement.channel)} <- ${formatExpr(statement.value)}`;
    case "UnsupportedStmt":
      return statement.reason;
    default:
      return "";
  }
}

function formatIfStmt(statement: Extract<Stmt, { kind: "IfStmt" }>, depth: number): string {
  const init = statement.init ? `${formatSimpleStmt(statement.init)}; ` : "";
  const elsePart = statement.else
    ? statement.else.kind === "IfStmt"
      ? ` else ${formatIfStmt(statement.else, depth)}`
      : ` else ${formatBlock(statement.else.kind === "BlockStmt" ? statement.else : { kind: "BlockStmt", statements: [statement.else] }, depth)}`
    : "";
  return `if ${init}${formatExpr(statement.condition)} ${formatBlock(statement.body, depth)}${elsePart}`;
}

function formatForStmt(statement: Extract<Stmt, { kind: "ForStmt" }>, depth: number): string {
  if (!statement.init && !statement.condition && !statement.post) return `for ${formatBlock(statement.body, depth)}`;
  if (!statement.init && !statement.post) return `for ${formatExpr(statement.condition!)} ${formatBlock(statement.body, depth)}`;
  const init = statement.init ? formatSimpleStmt(statement.init) : "";
  const condition = statement.condition ? formatExpr(statement.condition) : "";
  const post = statement.post ? formatSimpleStmt(statement.post) : "";
  return `for ${init}; ${condition}; ${post} ${formatBlock(statement.body, depth)}`;
}

function formatRangeStmt(statement: RangeStmt, depth: number): string {
  const lhs = [statement.key, statement.value].filter((expr): expr is Expr => Boolean(expr)).map((expr) => formatExpr(expr)).join(", ");
  const assign = statement.token === TokenKind.Define ? ":=" : "=";
  const header = lhs ? `${lhs} ${assign} range ${formatExpr(statement.source)}` : `range ${formatExpr(statement.source)}`;
  return `for ${header} ${formatBlock(statement.body, depth)}`;
}

function formatSwitchStmt(statement: Extract<Stmt, { kind: "SwitchStmt" }>, depth: number): string {
  const init = statement.init ? `${formatSimpleStmt(statement.init)}; ` : "";
  const tag = statement.tag ? formatExpr(statement.tag) : "";
  const header = `${init}${tag}`.trim();
  return `switch${header ? ` ${header}` : ""} ${formatClauses(statement.body, depth, formatCaseClause)}`;
}

function formatTypeSwitchStmt(statement: Extract<Stmt, { kind: "TypeSwitchStmt" }>, depth: number): string {
  const init = statement.init ? `${formatSimpleStmt(statement.init)}; ` : "";
  return `switch ${init}${formatSimpleStmt(statement.assign)} ${formatClauses(statement.body, depth, formatCaseClause)}`;
}

function formatClauses<T>(clauses: T[], depth: number, formatClause: (clause: T, depth: number) => string): string {
  if (clauses.length === 0) return "{}";
  return `{\n${clauses.map((clause) => formatClause(clause, depth + 1)).join("\n")}\n${indent(depth)}}`;
}

function formatCaseClause(clause: CaseClause, depth: number): string {
  const header = clause.default ? "default:" : `case ${clause.list.map((expr) => formatExpr(expr)).join(", ")}:`;
  return formatClauseBody(header, clause.body, depth);
}

function formatCommClause(clause: Extract<Stmt, { kind: "CommClause" }>, depth: number): string {
  const header = clause.default ? "default:" : `case ${clause.comm ? formatSimpleStmt(clause.comm) : ""}:`;
  return formatClauseBody(header, clause.body, depth);
}

function formatClauseBody(header: string, body: Stmt[], depth: number): string {
  if (body.length === 0) return `${indent(depth)}${header}`;
  return `${indent(depth)}${header}\n${body.map((stmt) => `${indent(depth + 1)}${formatStmt(stmt, depth + 1)}`).join("\n")}`;
}

function formatSimpleStmt(statement: Stmt): string {
  const formatted = formatStmt(statement, 0);
  return formatted.replace(/\n+/g, " ").trim();
}

function formatExpr(expr: Expr, parentPrecedence = 0): string {
  switch (expr.kind) {
    case "BadExpr":
      return "<bad>";
    case "Ident":
      return expr.name;
    case "BasicLit":
      return expr.value;
    case "Ellipsis":
      return `...${expr.element ? formatExpr(expr.element) : ""}`;
    case "FuncLit":
      return formatFuncLit(expr, 0);
    case "CompositeLit":
      return formatCompositeLit(expr);
    case "ParenExpr":
      return `(${formatExpr(expr.expr)})`;
    case "SelectorExpr":
      return `${formatExpr(expr.object, 8)}.${expr.selector.name}`;
    case "IndexExpr":
      return `${formatExpr(expr.object, 8)}[${formatExpr(expr.index)}]`;
    case "IndexListExpr":
      return `${formatExpr(expr.object, 8)}[${expr.indices.map((index) => formatExpr(index)).join(", ")}]`;
    case "SliceExpr":
      return `${formatExpr(expr.object, 8)}[${expr.low ? formatExpr(expr.low) : ""}:${expr.high ? formatExpr(expr.high) : ""}${expr.max ? `:${formatExpr(expr.max)}` : ""}]`;
    case "TypeAssertExpr":
      return `${formatExpr(expr.object, 8)}.(${expr.type ? formatExpr(expr.type) : "type"})`;
    case "CallExpr":
      return formatCallExpr(expr);
    case "StarExpr":
      return `*${formatExpr(expr.expr, 7)}`;
    case "UnaryExpr":
      return `${unaryOperator(expr.op)}${formatExpr(expr.expr, 7)}`;
    case "BinaryExpr":
      return formatBinaryExpr(expr, parentPrecedence);
    case "KeyValueExpr":
      return `${formatExpr(expr.key)}: ${formatExpr(expr.value)}`;
    case "CellRefExpr":
      return `${expr.namespace.name}.${expr.address.raw}`;
    case "RangeRefExpr":
      return `${expr.namespace.name}.${expr.start.raw}:${expr.end.raw}`;
    case "ArrayType":
      return formatArrayType(expr);
    case "StructType":
      return `struct{${formatFieldListContent(expr.fields)}}`;
    case "FuncType":
      return `func${formatFuncSignature(expr)}`;
    case "InterfaceType":
      return `interface{${formatInterfaceFields(expr.methods)}}`;
    case "MapType":
      return `map[${formatExpr(expr.key)}]${formatExpr(expr.value)}`;
    case "ChanType":
      return formatChanType(expr);
  }
}

function formatCompositeLit(expr: CompositeLit): string {
  const type = expr.type ? formatExpr(expr.type) : "";
  return `${type}{${expr.elements.map((element) => formatExpr(element)).join(", ")}}`;
}

function formatCallExpr(expr: CallExpr): string {
  return `${formatExpr(expr.fun, 8)}(${expr.args.map((arg, index) =>
    `${formatExpr(arg)}${expr.ellipsis && index === expr.args.length - 1 ? "..." : ""}`
  ).join(", ")})`;
}

function formatBinaryExpr(expr: Extract<Expr, { kind: "BinaryExpr" }>, parentPrecedence: number): string {
  const precedence = binaryPrecedence(expr.op);
  const text = `${formatExpr(expr.left, precedence)} ${binaryOperator(expr.op)} ${formatExpr(expr.right, precedence + 1)}`;
  return precedence < parentPrecedence ? `(${text})` : text;
}

function formatArrayType(expr: ArrayType): string {
  const length = expr.inferredLength ? "..." : expr.length ? formatExpr(expr.length) : "";
  return `[${length}]${formatExpr(expr.element)}`;
}

function formatChanType(expr: ChanType): string {
  if (expr.direction === "send") return `chan<- ${formatExpr(expr.value)}`;
  if (expr.direction === "receive") return `<-chan ${formatExpr(expr.value)}`;
  return `chan ${formatExpr(expr.value)}`;
}

function formatFieldListContent(list: FieldList): string {
  return list.fields.map(formatField).join(", ");
}

function formatInterfaceFields(list: FieldList): string {
  return list.fields.map((field) => {
    const fieldType = field.type;
    if (fieldType.kind === "FuncType" && field.names.length > 0) {
      return field.names.map((name) => `${name.name}(${formatFieldListContent(fieldType.params)})${formatResults(fieldType.results)}`).join("; ");
    }
    return formatField(field);
  }).join("; ");
}

function formatField(field: Field): string {
  const names = field.names.map((name) => name.name).join(", ");
  const tag = field.tag ? ` ${field.tag.value}` : "";
  return `${names ? `${names} ` : ""}${formatExpr(field.type)}${tag}`;
}

function tokenKeyword(token: GenDecl["token"]): string {
  switch (token) {
    case TokenKind.Import:
      return "import";
    case TokenKind.Const:
      return "const";
    case TokenKind.Type:
      return "type";
    case TokenKind.Var:
      return "var";
  }
}

function branchKeyword(statement: BranchStmt): string {
  switch (statement.token) {
    case TokenKind.Break:
      return "break";
    case TokenKind.Continue:
      return "continue";
    case TokenKind.Goto:
      return "goto";
    case TokenKind.Fallthrough:
      return "fallthrough";
  }
}

function assignmentOperator(token: Extract<Stmt, { kind: "AssignStmt" }>["token"]): string {
  switch (token) {
    case TokenKind.Define:
      return ":=";
    case TokenKind.PlusAssign:
      return "+=";
    case TokenKind.MinusAssign:
      return "-=";
    case TokenKind.StarAssign:
      return "*=";
    case TokenKind.SlashAssign:
      return "/=";
    case TokenKind.PercentAssign:
      return "%=";
    case TokenKind.AmpAssign:
      return "&=";
    case TokenKind.OrAssign:
      return "|=";
    case TokenKind.CaretAssign:
      return "^=";
    case TokenKind.ShlAssign:
      return "<<=";
    case TokenKind.ShrAssign:
      return ">>=";
    case TokenKind.BitClearAssign:
      return "&^=";
    default:
      return "=";
  }
}

function unaryOperator(token: TokenKind): string {
  switch (token) {
    case TokenKind.Arrow:
      return "<-";
    case TokenKind.Minus:
      return "-";
    case TokenKind.Bang:
      return "!";
    case TokenKind.Caret:
      return "^";
    case TokenKind.Tilde:
      return "~";
    case TokenKind.Amp:
      return "&";
    case TokenKind.Star:
      return "*";
    default:
      return "+";
  }
}

function binaryOperator(token: BinaryOperator): string {
  switch (token) {
    case TokenKind.OrOr:
      return "||";
    case TokenKind.AndAnd:
      return "&&";
    case TokenKind.Equal:
      return "==";
    case TokenKind.NotEqual:
      return "!=";
    case TokenKind.Less:
      return "<";
    case TokenKind.LessEqual:
      return "<=";
    case TokenKind.Greater:
      return ">";
    case TokenKind.GreaterEqual:
      return ">=";
    case TokenKind.Minus:
      return "-";
    case TokenKind.Or:
      return "|";
    case TokenKind.Caret:
      return "^";
    case TokenKind.Star:
      return "*";
    case TokenKind.Slash:
      return "/";
    case TokenKind.Percent:
      return "%";
    case TokenKind.Shl:
      return "<<";
    case TokenKind.Shr:
      return ">>";
    case TokenKind.Amp:
      return "&";
    case TokenKind.BitClear:
      return "&^";
    default:
      return "+";
  }
}

function binaryPrecedence(token: BinaryOperator): number {
  switch (token) {
    case TokenKind.OrOr:
      return 1;
    case TokenKind.AndAnd:
      return 2;
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
  }
}

function indent(depth: number): string {
  return INDENT.repeat(depth);
}

function ensureTrailingNewline(text: string): string {
  return text.endsWith("\n") ? text : `${text}\n`;
}

function isStmt(node: AstNode): node is Stmt {
  return [
    "BadStmt",
    "DeclStmt",
    "EmptyStmt",
    "LabeledStmt",
    "ExprStmt",
    "AssignStmt",
    "IncDecStmt",
    "ReturnStmt",
    "BranchStmt",
    "BlockStmt",
    "IfStmt",
    "CaseClause",
    "CommClause",
    "SwitchStmt",
    "TypeSwitchStmt",
    "SelectStmt",
    "ForStmt",
    "RangeStmt",
    "DeferStmt",
    "GoStmt",
    "SendStmt",
    "UnsupportedStmt"
  ].includes(node.kind);
}

function isExpr(node: AstNode): node is Expr {
  return [
    "BadExpr",
    "Ident",
    "Ellipsis",
    "BasicLit",
    "FuncLit",
    "CompositeLit",
    "ParenExpr",
    "SelectorExpr",
    "IndexExpr",
    "IndexListExpr",
    "SliceExpr",
    "TypeAssertExpr",
    "CallExpr",
    "StarExpr",
    "UnaryExpr",
    "BinaryExpr",
    "KeyValueExpr",
    "CellRefExpr",
    "RangeRefExpr",
    "ArrayType",
    "StructType",
    "FuncType",
    "InterfaceType",
    "MapType",
    "ChanType"
  ].includes(node.kind);
}
