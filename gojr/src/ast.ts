import { Diagnostic, SourceSpan } from "./diagnostics.js";

export interface ImportDecl {
  path: string;
  alias?: string;
}

export type ProgramKind = "script" | "function";

export interface ProgramAst {
  kind: ProgramKind;
  imports: ImportDecl[];
  diagnostics: Diagnostic[];
  body: Statement[];
  functions: FunctionDecl[];
}

export interface FunctionDecl {
  kind: "FunctionDecl";
  name: string;
  receiver?: ReceiverDecl;
  signature: Signature;
  body: BlockStatement;
  span?: SourceSpan;
}

export interface ReceiverDecl {
  name?: string;
  type: TypeNode;
}

export interface Signature {
  parameters: ParameterDecl[];
  results: ParameterDecl[];
}

export interface ParameterDecl {
  name?: string;
  type: TypeNode;
  variadic: boolean;
}

export interface TypeNode {
  text: string;
  span?: SourceSpan;
}

export type Statement =
  | BlockStatement
  | LabeledStatement
  | ConstDeclStatement
  | VarDeclStatement
  | TypeDeclStatement
  | ReturnStatement
  | IfStatement
  | SwitchStatement
  | ForStatement
  | DeferStatement
  | BranchStatement
  | AssignStatement
  | ShortVarStatement
  | IncDecStatement
  | ExpressionStatement;

export interface BlockStatement {
  kind: "BlockStatement";
  statements: Statement[];
  span?: SourceSpan;
}

export interface LabeledStatement {
  kind: "LabeledStatement";
  label: string;
  statement?: Statement;
  span?: SourceSpan;
}

export interface DeclarationSpec {
  name: string;
  type?: TypeNode;
  value?: Expression;
}

export interface ConstDeclStatement {
  kind: "ConstDecl";
  declarations: DeclarationSpec[];
  span?: SourceSpan;
}

export interface VarDeclStatement {
  kind: "VarDecl";
  declarations: DeclarationSpec[];
  span?: SourceSpan;
}

export interface TypeSpec {
  name: string;
  type: TypeNode;
}

export interface TypeDeclStatement {
  kind: "TypeDecl";
  declarations: TypeSpec[];
  span?: SourceSpan;
}

export interface ReturnStatement {
  kind: "ReturnStatement";
  values: Expression[];
  span?: SourceSpan;
}

export interface IfStatement {
  kind: "IfStatement";
  condition: Expression;
  thenBlock: BlockStatement;
  elseBranch?: IfStatement | BlockStatement;
  span?: SourceSpan;
}

export interface SwitchClause {
  kind: "SwitchClause";
  values: Expression[];
  default: boolean;
  statements: Statement[];
}

export interface SwitchStatement {
  kind: "SwitchStatement";
  expression?: Expression;
  clauses: SwitchClause[];
  span?: SourceSpan;
}

export interface ForStatement {
  kind: "ForStatement";
  init?: Statement;
  condition?: Expression;
  post?: Statement;
  range?: RangeClause;
  body: BlockStatement;
  span?: SourceSpan;
}

export interface RangeClause {
  keyName?: string;
  valueName?: string;
  define: boolean;
  source: Expression;
}

export interface DeferStatement {
  kind: "DeferStatement";
  expression: Expression;
  span?: SourceSpan;
}

export interface BranchStatement {
  kind: "BranchStatement";
  branch: "break" | "continue" | "fallthrough" | "goto";
  label?: string;
  span?: SourceSpan;
}

export interface AssignStatement {
  kind: "AssignStatement";
  target: Expression;
  value: Expression;
  span?: SourceSpan;
}

export interface ShortVarStatement {
  kind: "ShortVarStatement";
  name: string;
  value: Expression;
  span?: SourceSpan;
}

export interface IncDecStatement {
  kind: "IncDecStatement";
  target: Expression;
  operator: "++" | "--";
  span?: SourceSpan;
}

export interface ExpressionStatement {
  kind: "ExpressionStatement";
  expression: Expression;
  span?: SourceSpan;
}

export type Expression =
  | IdentifierExpression
  | LiteralExpression
  | FunctionLiteralExpression
  | MapLiteralExpression
  | UnaryExpression
  | BinaryExpression
  | SelectorExpression
  | CallExpression
  | IndexExpression
  | SliceExpression
  | SpreadsheetRangeExpression;

export interface IdentifierExpression {
  kind: "Identifier";
  name: string;
  span?: SourceSpan;
}

export type LiteralValue = bigint | number | string | boolean | null;

export interface LiteralExpression {
  kind: "Literal";
  literalKind: "int" | "float" | "string" | "bool" | "nil";
  value: LiteralValue;
  raw: string;
  span?: SourceSpan;
}

export interface FunctionLiteralExpression {
  kind: "FunctionLiteralExpression";
  signature: Signature;
  body: BlockStatement;
  span?: SourceSpan;
}

export interface MapEntryExpression {
  key: Expression;
  value: Expression;
}

export interface MapLiteralExpression {
  kind: "MapLiteralExpression";
  keyType: TypeNode;
  valueType: TypeNode;
  entries: MapEntryExpression[];
  span?: SourceSpan;
}

export interface UnaryExpression {
  kind: "UnaryExpression";
  operator: "+" | "-" | "!" | "&" | "*";
  operand: Expression;
  span?: SourceSpan;
}

export interface BinaryExpression {
  kind: "BinaryExpression";
  operator: "||" | "&&" | "==" | "!=" | "<" | "<=" | ">" | ">=" | "+" | "-" | "*" | "/" | "%";
  left: Expression;
  right: Expression;
  span?: SourceSpan;
}

export interface SelectorExpression {
  kind: "SelectorExpression";
  object: Expression;
  field: string;
  span?: SourceSpan;
}

export interface CallExpression {
  kind: "CallExpression";
  callee: Expression;
  args: Expression[];
  spreadLast: boolean;
  span?: SourceSpan;
}

export interface IndexExpression {
  kind: "IndexExpression";
  object: Expression;
  index: Expression;
  span?: SourceSpan;
}

export interface SliceExpression {
  kind: "SliceExpression";
  object: Expression;
  start?: Expression;
  end?: Expression;
  span?: SourceSpan;
}

export interface SpreadsheetRangeExpression {
  kind: "SpreadsheetRangeExpression";
  start: SelectorExpression;
  endCell: string;
  span?: SourceSpan;
}
