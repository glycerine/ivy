import { SourceSpan } from "../diagnostics.js";
import { TokenKind } from "./token.js";

export interface Node {
  kind: NodeKind;
  span?: SourceSpan;
}

export type NodeKind =
  | "BadExpr"
  | "Ident"
  | "Ellipsis"
  | "BasicLit"
  | "FuncLit"
  | "CompositeLit"
  | "ParenExpr"
  | "SelectorExpr"
  | "IndexExpr"
  | "SliceExpr"
  | "TypeAssertExpr"
  | "CallExpr"
  | "StarExpr"
  | "UnaryExpr"
  | "BinaryExpr"
  | "KeyValueExpr"
  | "CellRefExpr"
  | "RangeRefExpr"
  | "ArrayType"
  | "StructType"
  | "FuncType"
  | "InterfaceType"
  | "MapType"
  | "ChanType"
  | "BadStmt"
  | "DeclStmt"
  | "EmptyStmt"
  | "LabeledStmt"
  | "ExprStmt"
  | "AssignStmt"
  | "IncDecStmt"
  | "ReturnStmt"
  | "BranchStmt"
  | "BlockStmt"
  | "IfStmt"
  | "CaseClause"
  | "CommClause"
  | "SwitchStmt"
  | "TypeSwitchStmt"
  | "SelectStmt"
  | "ForStmt"
  | "RangeStmt"
  | "DeferStmt"
  | "GoStmt"
  | "SendStmt"
  | "UnsupportedStmt"
  | "ImportSpec"
  | "ValueSpec"
  | "TypeSpec"
  | "BadDecl"
  | "GenDecl"
  | "FuncDecl"
  | "Field"
  | "FieldList"
  | "File"
  | "Package"
  | "Comment"
  | "CommentGroup";

export interface Ident extends Node {
  kind: "Ident";
  name: string;
}

export interface BadExpr extends Node {
  kind: "BadExpr";
}

export interface Ellipsis extends Node {
  kind: "Ellipsis";
  element?: Expr;
}

export interface BasicLit extends Node {
  kind: "BasicLit";
  token: TokenKind.IntLiteral | TokenKind.FloatLiteral | TokenKind.ImagLiteral | TokenKind.RuneLiteral | TokenKind.StringLiteral;
  value: string;
}

export interface FuncLit extends Node {
  kind: "FuncLit";
  type: FuncType;
  body: BlockStmt;
}

export interface CompositeLit extends Node {
  kind: "CompositeLit";
  type?: Expr;
  elements: Expr[];
}

export interface ParenExpr extends Node {
  kind: "ParenExpr";
  expr: Expr;
}

export interface SelectorExpr extends Node {
  kind: "SelectorExpr";
  object: Expr;
  selector: Ident;
}

export interface IndexExpr extends Node {
  kind: "IndexExpr";
  object: Expr;
  index: Expr;
}

export interface SliceExpr extends Node {
  kind: "SliceExpr";
  object: Expr;
  low?: Expr;
  high?: Expr;
  max?: Expr;
}

export interface TypeAssertExpr extends Node {
  kind: "TypeAssertExpr";
  object: Expr;
  type?: Expr;
  typeSwitch: boolean;
}

export interface CallExpr extends Node {
  kind: "CallExpr";
  fun: Expr;
  args: Expr[];
  ellipsis: boolean;
}

export interface StarExpr extends Node {
  kind: "StarExpr";
  expr: Expr;
}

export interface UnaryExpr extends Node {
  kind: "UnaryExpr";
  op: TokenKind.Plus | TokenKind.Minus | TokenKind.Bang | TokenKind.Caret | TokenKind.Amp | TokenKind.Arrow;
  expr: Expr;
}

export interface BinaryExpr extends Node {
  kind: "BinaryExpr";
  left: Expr;
  op: BinaryOperator;
  right: Expr;
}

export type BinaryOperator =
  | TokenKind.OrOr
  | TokenKind.AndAnd
  | TokenKind.Equal
  | TokenKind.NotEqual
  | TokenKind.Less
  | TokenKind.LessEqual
  | TokenKind.Greater
  | TokenKind.GreaterEqual
  | TokenKind.Plus
  | TokenKind.Minus
  | TokenKind.Or
  | TokenKind.Caret
  | TokenKind.Star
  | TokenKind.Slash
  | TokenKind.Percent
  | TokenKind.Shl
  | TokenKind.Shr
  | TokenKind.Amp
  | TokenKind.BitClear;

export interface KeyValueExpr extends Node {
  kind: "KeyValueExpr";
  key: Expr;
  value: Expr;
}

export interface CellRefExpr extends Node {
  kind: "CellRefExpr";
  namespace: Ident;
  address: CellAddress;
}

export interface RangeRefExpr extends Node {
  kind: "RangeRefExpr";
  namespace: Ident;
  start: CellAddress;
  end: CellAddress;
}

export interface CellAddress {
  raw: string;
  column: string;
  row: number;
  absoluteColumn: boolean;
  absoluteRow: boolean;
}

export interface ArrayType extends Node {
  kind: "ArrayType";
  length?: Expr;
  element: Expr;
  inferredLength: boolean;
}

export interface StructType extends Node {
  kind: "StructType";
  fields: FieldList;
}

export interface FuncType extends Node {
  kind: "FuncType";
  params: FieldList;
  results?: FieldList;
}

export interface InterfaceType extends Node {
  kind: "InterfaceType";
  methods: FieldList;
}

export interface MapType extends Node {
  kind: "MapType";
  key: Expr;
  value: Expr;
}

export interface ChanType extends Node {
  kind: "ChanType";
  direction: "send" | "receive" | "both";
  value: Expr;
}

export type Expr =
  | BadExpr
  | Ident
  | Ellipsis
  | BasicLit
  | FuncLit
  | CompositeLit
  | ParenExpr
  | SelectorExpr
  | IndexExpr
  | SliceExpr
  | TypeAssertExpr
  | CallExpr
  | StarExpr
  | UnaryExpr
  | BinaryExpr
  | KeyValueExpr
  | CellRefExpr
  | RangeRefExpr
  | ArrayType
  | StructType
  | FuncType
  | InterfaceType
  | MapType
  | ChanType;

export interface BadStmt extends Node {
  kind: "BadStmt";
}

export interface DeclStmt extends Node {
  kind: "DeclStmt";
  decl: Decl;
}

export interface EmptyStmt extends Node {
  kind: "EmptyStmt";
  implicit: boolean;
}

export interface LabeledStmt extends Node {
  kind: "LabeledStmt";
  label: Ident;
  stmt: Stmt;
}

export interface ExprStmt extends Node {
  kind: "ExprStmt";
  expr: Expr;
}

export interface AssignStmt extends Node {
  kind: "AssignStmt";
  lhs: Expr[];
  token:
    | TokenKind.Assign
    | TokenKind.Define
    | TokenKind.PlusAssign
    | TokenKind.MinusAssign
    | TokenKind.StarAssign
    | TokenKind.SlashAssign
    | TokenKind.PercentAssign
    | TokenKind.AmpAssign
    | TokenKind.OrAssign
    | TokenKind.CaretAssign
    | TokenKind.ShlAssign
    | TokenKind.ShrAssign
    | TokenKind.BitClearAssign;
  rhs: Expr[];
}

export interface IncDecStmt extends Node {
  kind: "IncDecStmt";
  expr: Expr;
  token: TokenKind.PlusPlus | TokenKind.MinusMinus;
}

export interface ReturnStmt extends Node {
  kind: "ReturnStmt";
  results: Expr[];
}

export interface BranchStmt extends Node {
  kind: "BranchStmt";
  token: TokenKind.Break | TokenKind.Continue | TokenKind.Goto | TokenKind.Fallthrough;
  label?: Ident;
}

export interface BlockStmt extends Node {
  kind: "BlockStmt";
  statements: Stmt[];
}

export interface IfStmt extends Node {
  kind: "IfStmt";
  init?: Stmt;
  condition: Expr;
  body: BlockStmt;
  else?: Stmt;
}

export interface CaseClause extends Node {
  kind: "CaseClause";
  list: Expr[];
  body: Stmt[];
  default: boolean;
}

export interface CommClause extends Node {
  kind: "CommClause";
  comm?: Stmt;
  body: Stmt[];
  default: boolean;
}

export interface SwitchStmt extends Node {
  kind: "SwitchStmt";
  init?: Stmt;
  tag?: Expr;
  body: CaseClause[];
}

export interface TypeSwitchStmt extends Node {
  kind: "TypeSwitchStmt";
  init?: Stmt;
  assign: Stmt;
  body: CaseClause[];
}

export interface SelectStmt extends Node {
  kind: "SelectStmt";
  body: CommClause[];
}

export interface ForStmt extends Node {
  kind: "ForStmt";
  init?: Stmt;
  condition?: Expr;
  post?: Stmt;
  body: BlockStmt;
}

export interface RangeStmt extends Node {
  kind: "RangeStmt";
  key?: Expr;
  value?: Expr;
  token: TokenKind.Assign | TokenKind.Define;
  source: Expr;
  body: BlockStmt;
}

export interface DeferStmt extends Node {
  kind: "DeferStmt";
  call: CallExpr;
}

export interface GoStmt extends Node {
  kind: "GoStmt";
  call: CallExpr;
}

export interface SendStmt extends Node {
  kind: "SendStmt";
  channel: Expr;
  value: Expr;
}

export interface UnsupportedStmt extends Node {
  kind: "UnsupportedStmt";
  token: TokenKind.Go | TokenKind.Select;
  reason: string;
}

export type Stmt =
  | BadStmt
  | DeclStmt
  | EmptyStmt
  | LabeledStmt
  | ExprStmt
  | AssignStmt
  | IncDecStmt
  | ReturnStmt
  | BranchStmt
  | BlockStmt
  | IfStmt
  | CaseClause
  | CommClause
  | SwitchStmt
  | TypeSwitchStmt
  | SelectStmt
  | ForStmt
  | RangeStmt
  | DeferStmt
  | GoStmt
  | SendStmt
  | UnsupportedStmt;

export interface Field extends Node {
  kind: "Field";
  names: Ident[];
  type: Expr;
  tag?: BasicLit;
}

export interface FieldList extends Node {
  kind: "FieldList";
  fields: Field[];
}

export interface ImportSpec extends Node {
  kind: "ImportSpec";
  name?: Ident;
  path: BasicLit;
}

export interface ValueSpec extends Node {
  kind: "ValueSpec";
  names: Ident[];
  type?: Expr;
  values: Expr[];
}

export interface TypeSpec extends Node {
  kind: "TypeSpec";
  name: Ident;
  type: Expr;
  alias: boolean;
}

export type Spec = ImportSpec | ValueSpec | TypeSpec;

export interface BadDecl extends Node {
  kind: "BadDecl";
}

export interface GenDecl extends Node {
  kind: "GenDecl";
  token: TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var;
  specs: Spec[];
  grouped: boolean;
}

export interface FuncDecl extends Node {
  kind: "FuncDecl";
  receiver?: FieldList;
  name: Ident;
  type: FuncType;
  body?: BlockStmt;
}

export type Decl = BadDecl | GenDecl | FuncDecl;

export interface File extends Node {
  kind: "File";
  name?: Ident;
  declarations: Decl[];
  imports: ImportSpec[];
  unresolved: Ident[];
  comments: CommentGroup[];
}

export interface Package extends Node {
  kind: "Package";
  name: string;
  files: File[];
}

export interface Comment extends Node {
  kind: "Comment";
  text: string;
}

export interface CommentGroup extends Node {
  kind: "CommentGroup";
  list: Comment[];
}

export type AstNode =
  | Expr
  | Stmt
  | Decl
  | Spec
  | Field
  | FieldList
  | File
  | Package
  | Comment
  | CommentGroup;

export function ident(name: string, span?: SourceSpan): Ident {
  return { kind: "Ident", name, ...(span ? { span } : {}) };
}

export function parseCellAddress(raw: string): CellAddress | undefined {
  const match = /^(\$?)([A-Z]{1,3})(\$?)([1-9][0-9]*)$/.exec(raw);
  if (!match) return undefined;
  const [, columnAbs, column, rowAbs, row] = match;
  if (!column || !row) return undefined;
  return {
    raw,
    column,
    row: Number(row),
    absoluteColumn: columnAbs === "$",
    absoluteRow: rowAbs === "$"
  };
}

export function walk(node: AstNode, visit: (node: AstNode) => void): void {
  visit(node);
  for (const child of childNodes(node)) {
    walk(child, visit);
  }
}

export function childNodes(node: AstNode): AstNode[] {
  switch (node.kind) {
    case "File":
      return [...(node.name ? [node.name] : []), ...node.declarations, ...node.imports, ...node.unresolved, ...node.comments];
    case "Package":
      return node.files;
    case "GenDecl":
      return node.specs;
    case "FuncDecl":
      return [...(node.receiver ? [node.receiver] : []), node.name, node.type, ...(node.body ? [node.body] : [])];
    case "FieldList":
      return node.fields;
    case "Field":
      return [...node.names, node.type, ...(node.tag ? [node.tag] : [])];
    case "ImportSpec":
      return [...(node.name ? [node.name] : []), node.path];
    case "ValueSpec":
      return [...node.names, ...(node.type ? [node.type] : []), ...node.values];
    case "TypeSpec":
      return [node.name, node.type];
    case "FuncType":
      return [node.params, ...(node.results ? [node.results] : [])];
    case "BlockStmt":
      return node.statements;
    case "DeclStmt":
      return [node.decl];
    case "LabeledStmt":
      return [node.label, node.stmt];
    case "ExprStmt":
      return [node.expr];
    case "AssignStmt":
      return [...node.lhs, ...node.rhs];
    case "IncDecStmt":
      return [node.expr];
    case "ReturnStmt":
      return node.results;
    case "BranchStmt":
      return node.label ? [node.label] : [];
    case "IfStmt":
      return [...(node.init ? [node.init] : []), node.condition, node.body, ...(node.else ? [node.else] : [])];
    case "CaseClause":
      return [...node.list, ...node.body];
    case "CommClause":
      return [...(node.comm ? [node.comm] : []), ...node.body];
    case "SwitchStmt":
      return [...(node.init ? [node.init] : []), ...(node.tag ? [node.tag] : []), ...node.body];
    case "TypeSwitchStmt":
      return [...(node.init ? [node.init] : []), node.assign, ...node.body];
    case "SelectStmt":
      return [...node.body];
    case "ForStmt":
      return [...(node.init ? [node.init] : []), ...(node.condition ? [node.condition] : []), ...(node.post ? [node.post] : []), node.body];
    case "RangeStmt":
      return [...(node.key ? [node.key] : []), ...(node.value ? [node.value] : []), node.source, node.body];
    case "DeferStmt":
      return [node.call];
    case "GoStmt":
      return [node.call];
    case "SendStmt":
      return [node.channel, node.value];
    case "Ellipsis":
      return node.element ? [node.element] : [];
    case "FuncLit":
      return [node.type, node.body];
    case "CompositeLit":
      return [...(node.type ? [node.type] : []), ...node.elements];
    case "ParenExpr":
      return [node.expr];
    case "SelectorExpr":
      return [node.object, node.selector];
    case "IndexExpr":
      return [node.object, node.index];
    case "SliceExpr":
      return [node.object, ...(node.low ? [node.low] : []), ...(node.high ? [node.high] : []), ...(node.max ? [node.max] : [])];
    case "TypeAssertExpr":
      return [node.object, ...(node.type ? [node.type] : [])];
    case "CallExpr":
      return [node.fun, ...node.args];
    case "StarExpr":
      return [node.expr];
    case "UnaryExpr":
      return [node.expr];
    case "BinaryExpr":
      return [node.left, node.right];
    case "KeyValueExpr":
      return [node.key, node.value];
    case "CellRefExpr":
      return [node.namespace];
    case "RangeRefExpr":
      return [node.namespace];
    case "ArrayType":
      return [...(node.length ? [node.length] : []), node.element];
    case "StructType":
      return [node.fields];
    case "InterfaceType":
      return [node.methods];
    case "MapType":
      return [node.key, node.value];
    case "ChanType":
      return [node.value];
    default:
      return [];
  }
}
