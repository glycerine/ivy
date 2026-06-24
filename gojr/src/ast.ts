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
  typeParameters?: string[];
  signature: Signature;
  body: BlockStatement;
  source?: string;
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
  | SelectStatement
  | ForStatement
  | DeferStatement
  | GoStatement
  | SendStatement
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
  valueGroup?: number;
  iotaIndex?: number;
  valueIndex?: number;
  valueCount?: number;
  groupNameCount?: number;
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
  typeParameters?: string[];
  type: TypeNode;
  structFields?: StructFieldDecl[];
  interfaceMethods?: InterfaceMethodDecl[];
  interfaceEmbeds?: TypeNode[];
}

export interface StructFieldDecl {
  name: string;
  type: TypeNode;
  embedded?: boolean;
  tag?: string;
}

export interface InterfaceMethodDecl {
  name: string;
  signature: Signature;
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
  init?: Statement;
  condition: Expression;
  thenBlock: BlockStatement;
  elseBranch?: IfStatement | BlockStatement;
  span?: SourceSpan;
}

export interface SwitchClause {
  kind: "SwitchClause";
  values: Expression[];
  typeValues?: TypeNode[];
  default: boolean;
  statements: Statement[];
}

export interface TypeSwitchGuard {
  name?: string;
  define: boolean;
  expression: Expression;
}

export interface SwitchStatement {
  kind: "SwitchStatement";
  init?: Statement;
  expression?: Expression;
  typeSwitch?: TypeSwitchGuard;
  clauses: SwitchClause[];
  span?: SourceSpan;
}

export interface CommClause {
  kind: "CommClause";
  comm?: Statement;
  default: boolean;
  statements: Statement[];
}

export interface SelectStatement {
  kind: "SelectStatement";
  clauses: CommClause[];
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
  keyTarget?: Expression;
  valueName?: string;
  valueTarget?: Expression;
  define: boolean;
  source: Expression;
}

export interface DeferStatement {
  kind: "DeferStatement";
  expression: Expression;
  span?: SourceSpan;
}

export interface GoStatement {
  kind: "GoStatement";
  call: CallExpression;
  span?: SourceSpan;
}

export interface SendStatement {
  kind: "SendStatement";
  channel: Expression;
  value: Expression;
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
  targets: Expression[];
  operator: "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^=" | "&^=" | "<<=" | ">>=";
  values: Expression[];
  span?: SourceSpan;
}

export interface ShortVarStatement {
  kind: "ShortVarStatement";
  names: string[];
  values: Expression[];
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
  | TypeExpression
  | LiteralExpression
  | FunctionLiteralExpression
  | ArrayLiteralExpression
  | StructLiteralExpression
  | MapLiteralExpression
  | UnaryExpression
  | BinaryExpression
  | TypeAssertionExpression
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

export interface TypeExpression {
  kind: "TypeExpression";
  type: TypeNode;
  span?: SourceSpan;
}

export interface ComplexLiteralValue {
  real: number;
  imag: number;
}

export type LiteralValue = bigint | number | ComplexLiteralValue | string | boolean | null;

export interface LiteralExpression {
  kind: "Literal";
  literalKind: "int" | "float" | "imag" | "rune" | "string" | "bool" | "nil";
  value: LiteralValue;
  raw: string;
  span?: SourceSpan;
}

export interface FunctionLiteralExpression {
  kind: "FunctionLiteralExpression";
  signature: Signature;
  body: BlockStatement;
  source?: string;
  span?: SourceSpan;
}

export interface ArrayLiteralExpression {
  kind: "ArrayLiteralExpression";
  type: TypeNode;
  elements: ArrayLiteralElement[];
  span?: SourceSpan;
}

export interface ArrayLiteralElement {
  key?: Expression;
  value: Expression;
}

export interface StructLiteralField {
  name?: string;
  key?: Expression;
  value: Expression;
}

export interface StructLiteralExpression {
  kind: "StructLiteralExpression";
  typeName: string;
  fields: StructLiteralField[];
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
  operator: "+" | "-" | "!" | "^" | "&" | "*" | "<-";
  operand: Expression;
  span?: SourceSpan;
}

export interface BinaryExpression {
  kind: "BinaryExpression";
  operator: "||" | "&&" | "==" | "!=" | "<" | "<=" | ">" | ">=" | "+" | "-" | "|" | "^" | "*" | "/" | "%" | "<<" | ">>" | "&" | "&^";
  left: Expression;
  right: Expression;
  span?: SourceSpan;
}

export interface TypeAssertionExpression {
  kind: "TypeAssertionExpression";
  expression: Expression;
  type: TypeNode;
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
  max?: Expression;
  span?: SourceSpan;
}

export interface SpreadsheetRangeExpression {
  kind: "SpreadsheetRangeExpression";
  start: SelectorExpression;
  endCell: string;
  span?: SourceSpan;
}
