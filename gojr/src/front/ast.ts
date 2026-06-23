import { SourceSpan } from "../diagnostics.js";
import { TokenKind } from "./token.js";

export type Pos = number;
export type ChanDir = number;

export const SEND: ChanDir = 1 << 0;
export const RECV: ChanDir = 1 << 1;

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
  | "IndexListExpr"
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
  Obj?: Object;
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

export interface IndexListExpr extends Node {
  kind: "IndexListExpr";
  object: Expr;
  indices: Expr[];
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
  op: TokenKind.Plus | TokenKind.Minus | TokenKind.Bang | TokenKind.Caret | TokenKind.Tilde | TokenKind.Amp | TokenKind.Arrow;
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
  typeParams?: FieldList;
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
  | IndexListExpr
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
  doc?: CommentGroup;
  name?: Ident;
  path: BasicLit;
  comment?: CommentGroup;
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
  typeParams?: FieldList;
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
  scope?: Scope;
}

export interface Package extends Node {
  kind: "Package";
  name: string;
  files: File[];
  scope?: Scope;
  imports?: Map<string, Object>;
}

export interface Comment extends Node {
  kind: "Comment";
  text: string;
}

export interface CommentGroup extends Node {
  kind: "CommentGroup";
  list: Comment[];
}

export function isWhitespace(ch: string): boolean {
  return ch === " " || ch === "\t" || ch === "\n" || ch === "\r";
}

export function stripTrailingWhitespace(s: string): string {
  let i = s.length;
  while (i > 0 && isWhitespace(s[i - 1] ?? "")) {
    i -= 1;
  }
  return s.slice(0, i);
}

export function isDirective(c: string): boolean {
  if (c.startsWith("line ") || c.startsWith("extern ") || c.startsWith("export ")) {
    return true;
  }

  const colon = c.indexOf(":");
  if (colon <= 0 || colon + 1 >= c.length) {
    return false;
  }
  for (let i = 0; i <= colon + 1; i += 1) {
    if (i === colon) {
      continue;
    }
    const b = c[i] ?? "";
    if (!(("a" <= b && b <= "z") || ("0" <= b && b <= "9"))) {
      return false;
    }
  }
  return true;
}

export function commentGroupText(group: CommentGroup | undefined): string {
  if (!group) {
    return "";
  }
  const comments = group.list.map((comment) => comment.text);
  const lines: string[] = [];
  for (let comment of comments) {
    switch (comment[1]) {
      case "/":
        comment = comment.slice(2);
        if (comment.length === 0) {
          break;
        }
        if (comment[0] === " ") {
          comment = comment.slice(1);
          break;
        }
        if (isDirective(comment)) {
          continue;
        }
        break;
      case "*":
        comment = comment.slice(2, -2);
        break;
      default:
        break;
    }
    for (const line of comment.split("\n")) {
      lines.push(stripTrailingWhitespace(line));
    }
  }

  let n = 0;
  for (const line of lines) {
    if (line !== "" || (n > 0 && lines[n - 1] !== "")) {
      lines[n] = line;
      n += 1;
    }
  }
  lines.length = n;
  if (n > 0 && lines[n - 1] !== "") {
    lines.push("");
  }
  return lines.join("\n");
}

export namespace CommentGroup {
  export function Pos(group: CommentGroup | undefined): Pos {
    return PosOf(group?.list[0]);
  }

  export function End(group: CommentGroup | undefined): Pos {
    return EndOf(group?.list[group.list.length - 1]);
  }

  export function Text(group: CommentGroup | undefined): string {
    return commentGroupText(group);
  }
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

export function NewIdent(name: string): Ident {
  return ident(name);
}

export function IsExported(name: string): boolean {
  const first = Array.from(name)[0] ?? "";
  return /^\p{Lu}$/u.test(first);
}

export namespace Ident {
  export function Pos(id: Ident | undefined): Pos {
    return nodePos(id);
  }

  export function End(id: Ident | undefined): Pos {
    return nodePos(id) + (id?.name.length ?? 0);
  }

  export function exprNode(_id: Ident | undefined): void {
  }

  export function IsExported(id: Ident | undefined): boolean {
    return id ? astIsExported(id.name) : false;
  }

  export function String(id: Ident | undefined): string {
    if (id) {
      return id.name;
    }
    return "<nil>";
  }
}

export namespace BadExpr {
  export function Pos(x: BadExpr): Pos { return nodePos(x); }
  export function End(x: BadExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: BadExpr): void {}
}

export namespace Ellipsis {
  export function Pos(x: Ellipsis): Pos { return nodePos(x); }
  export function End(x: Ellipsis): Pos { return x.element ? EndOf(x.element) : nodeEnd(x) || nodePos(x) + 3; }
  export function exprNode(_x: Ellipsis): void {}
}

export namespace BasicLit {
  export function Pos(x: BasicLit): Pos { return nodePos(x); }
  export function End(x: BasicLit): Pos { return nodeEnd(x) || nodePos(x) + x.value.length; }
  export function exprNode(_x: BasicLit): void {}
}

export namespace FuncLit {
  export function Pos(x: FuncLit): Pos { return PosOf(x.type); }
  export function End(x: FuncLit): Pos { return EndOf(x.body); }
  export function exprNode(_x: FuncLit): void {}
}

export namespace CompositeLit {
  export function Pos(x: CompositeLit): Pos { return x.type ? PosOf(x.type) : nodePos(x); }
  export function End(x: CompositeLit): Pos { return nodeEnd(x); }
  export function exprNode(_x: CompositeLit): void {}
}

export namespace ParenExpr {
  export function Pos(x: ParenExpr): Pos { return nodePos(x); }
  export function End(x: ParenExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: ParenExpr): void {}
}

export namespace SelectorExpr {
  export function Pos(x: SelectorExpr): Pos { return PosOf(x.object); }
  export function End(x: SelectorExpr): Pos { return EndOf(x.selector); }
  export function exprNode(_x: SelectorExpr): void {}
}

export namespace IndexExpr {
  export function Pos(x: IndexExpr): Pos { return PosOf(x.object); }
  export function End(x: IndexExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: IndexExpr): void {}
}

export namespace IndexListExpr {
  export function Pos(x: IndexListExpr): Pos { return PosOf(x.object); }
  export function End(x: IndexListExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: IndexListExpr): void {}
}

export namespace SliceExpr {
  export function Pos(x: SliceExpr): Pos { return PosOf(x.object); }
  export function End(x: SliceExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: SliceExpr): void {}
}

export namespace TypeAssertExpr {
  export function Pos(x: TypeAssertExpr): Pos { return PosOf(x.object); }
  export function End(x: TypeAssertExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: TypeAssertExpr): void {}
}

export namespace CallExpr {
  export function Pos(x: CallExpr): Pos { return PosOf(x.fun); }
  export function End(x: CallExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: CallExpr): void {}
}

export namespace StarExpr {
  export function Pos(x: StarExpr): Pos { return nodePos(x); }
  export function End(x: StarExpr): Pos { return EndOf(x.expr); }
  export function exprNode(_x: StarExpr): void {}
}

export namespace UnaryExpr {
  export function Pos(x: UnaryExpr): Pos { return nodePos(x); }
  export function End(x: UnaryExpr): Pos { return EndOf(x.expr); }
  export function exprNode(_x: UnaryExpr): void {}
}

export namespace BinaryExpr {
  export function Pos(x: BinaryExpr): Pos { return PosOf(x.left); }
  export function End(x: BinaryExpr): Pos { return EndOf(x.right); }
  export function exprNode(_x: BinaryExpr): void {}
}

export namespace KeyValueExpr {
  export function Pos(x: KeyValueExpr): Pos { return PosOf(x.key); }
  export function End(x: KeyValueExpr): Pos { return EndOf(x.value); }
  export function exprNode(_x: KeyValueExpr): void {}
}

export namespace ArrayType {
  export function Pos(x: ArrayType): Pos { return nodePos(x); }
  export function End(x: ArrayType): Pos { return EndOf(x.element); }
  export function exprNode(_x: ArrayType): void {}
}

export namespace StructType {
  export function Pos(x: StructType): Pos { return nodePos(x); }
  export function End(x: StructType): Pos { return FieldList.End(x.fields); }
  export function exprNode(_x: StructType): void {}
}

export namespace FuncType {
  export function Pos(x: FuncType): Pos { return nodePos(x) || FieldList.Pos(x.params); }
  export function End(x: FuncType): Pos { return x.results ? FieldList.End(x.results) : FieldList.End(x.params); }
  export function exprNode(_x: FuncType): void {}
}

export namespace InterfaceType {
  export function Pos(x: InterfaceType): Pos { return nodePos(x); }
  export function End(x: InterfaceType): Pos { return FieldList.End(x.methods); }
  export function exprNode(_x: InterfaceType): void {}
}

export namespace MapType {
  export function Pos(x: MapType): Pos { return nodePos(x); }
  export function End(x: MapType): Pos { return EndOf(x.value); }
  export function exprNode(_x: MapType): void {}
}

export namespace ChanType {
  export function Pos(x: ChanType): Pos { return nodePos(x); }
  export function End(x: ChanType): Pos { return EndOf(x.value); }
  export function exprNode(_x: ChanType): void {}
}

export namespace BadStmt {
  export function Pos(x: BadStmt): Pos { return nodePos(x); }
  export function End(x: BadStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: BadStmt): void {}
}

export namespace DeclStmt {
  export function Pos(x: DeclStmt): Pos { return PosOf(x.decl); }
  export function End(x: DeclStmt): Pos { return EndOf(x.decl); }
  export function stmtNode(_x: DeclStmt): void {}
}

export namespace EmptyStmt {
  export function Pos(x: EmptyStmt): Pos { return nodePos(x); }
  export function End(x: EmptyStmt): Pos { return x.implicit ? nodePos(x) : nodeEnd(x) || nodePos(x) + 1; }
  export function stmtNode(_x: EmptyStmt): void {}
}

export namespace LabeledStmt {
  export function Pos(x: LabeledStmt): Pos { return PosOf(x.label); }
  export function End(x: LabeledStmt): Pos { return EndOf(x.stmt); }
  export function stmtNode(_x: LabeledStmt): void {}
}

export namespace ExprStmt {
  export function Pos(x: ExprStmt): Pos { return PosOf(x.expr); }
  export function End(x: ExprStmt): Pos { return EndOf(x.expr); }
  export function stmtNode(_x: ExprStmt): void {}
}

export namespace SendStmt {
  export function Pos(x: SendStmt): Pos { return PosOf(x.channel); }
  export function End(x: SendStmt): Pos { return EndOf(x.value); }
  export function stmtNode(_x: SendStmt): void {}
}

export namespace IncDecStmt {
  export function Pos(x: IncDecStmt): Pos { return PosOf(x.expr); }
  export function End(x: IncDecStmt): Pos { return nodeEnd(x) || EndOf(x.expr) + 2; }
  export function stmtNode(_x: IncDecStmt): void {}
}

export namespace AssignStmt {
  export function Pos(x: AssignStmt): Pos { return PosOf(x.lhs[0]); }
  export function End(x: AssignStmt): Pos { return EndOf(x.rhs[x.rhs.length - 1]); }
  export function stmtNode(_x: AssignStmt): void {}
}

export namespace GoStmt {
  export function Pos(x: GoStmt): Pos { return nodePos(x); }
  export function End(x: GoStmt): Pos { return EndOf(x.call); }
  export function stmtNode(_x: GoStmt): void {}
}

export namespace DeferStmt {
  export function Pos(x: DeferStmt): Pos { return nodePos(x); }
  export function End(x: DeferStmt): Pos { return EndOf(x.call); }
  export function stmtNode(_x: DeferStmt): void {}
}

export namespace ReturnStmt {
  export function Pos(x: ReturnStmt): Pos { return nodePos(x); }
  export function End(x: ReturnStmt): Pos { return x.results.length > 0 ? EndOf(x.results[x.results.length - 1]) : nodeEnd(x) || nodePos(x) + 6; }
  export function stmtNode(_x: ReturnStmt): void {}
}

export namespace BranchStmt {
  export function Pos(x: BranchStmt): Pos { return nodePos(x); }
  export function End(x: BranchStmt): Pos { return x.label ? EndOf(x.label) : nodeEnd(x); }
  export function stmtNode(_x: BranchStmt): void {}
}

export namespace BlockStmt {
  export function Pos(x: BlockStmt): Pos { return nodePos(x); }
  export function End(x: BlockStmt): Pos { return nodeEnd(x) || (x.statements.length > 0 ? EndOf(x.statements[x.statements.length - 1]) : nodePos(x) + 1); }
  export function stmtNode(_x: BlockStmt): void {}
}

export namespace IfStmt {
  export function Pos(x: IfStmt): Pos { return nodePos(x); }
  export function End(x: IfStmt): Pos { return x.else ? EndOf(x.else) : EndOf(x.body); }
  export function stmtNode(_x: IfStmt): void {}
}

export namespace CaseClause {
  export function Pos(x: CaseClause): Pos { return nodePos(x); }
  export function End(x: CaseClause): Pos { return x.body.length > 0 ? EndOf(x.body[x.body.length - 1]) : nodeEnd(x); }
  export function stmtNode(_x: CaseClause): void {}
}

export namespace SwitchStmt {
  export function Pos(x: SwitchStmt): Pos { return nodePos(x); }
  export function End(x: SwitchStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: SwitchStmt): void {}
}

export namespace TypeSwitchStmt {
  export function Pos(x: TypeSwitchStmt): Pos { return nodePos(x); }
  export function End(x: TypeSwitchStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: TypeSwitchStmt): void {}
}

export namespace CommClause {
  export function Pos(x: CommClause): Pos { return nodePos(x); }
  export function End(x: CommClause): Pos { return x.body.length > 0 ? EndOf(x.body[x.body.length - 1]) : nodeEnd(x); }
  export function stmtNode(_x: CommClause): void {}
}

export namespace SelectStmt {
  export function Pos(x: SelectStmt): Pos { return nodePos(x); }
  export function End(x: SelectStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: SelectStmt): void {}
}

export namespace ForStmt {
  export function Pos(x: ForStmt): Pos { return nodePos(x); }
  export function End(x: ForStmt): Pos { return EndOf(x.body); }
  export function stmtNode(_x: ForStmt): void {}
}

export namespace RangeStmt {
  export function Pos(x: RangeStmt): Pos { return nodePos(x); }
  export function End(x: RangeStmt): Pos { return EndOf(x.body); }
  export function stmtNode(_x: RangeStmt): void {}
}

export namespace Comment {
  export function Pos(x: Comment): Pos { return nodePos(x); }
  export function End(x: Comment): Pos { return nodeEnd(x) || nodePos(x) + x.text.length; }
}

export namespace Field {
  export function Pos(x: Field): Pos { return x.names.length > 0 ? PosOf(x.names[0]) : PosOf(x.type); }
  export function End(x: Field): Pos { return x.tag ? EndOf(x.tag) : EndOf(x.type); }
}

export namespace FieldList {
  export function Pos(x: FieldList | undefined): Pos { return nodePos(x) || PosOf(x?.fields[0]); }
  export function End(x: FieldList | undefined): Pos { return x ? nodeEnd(x) || EndOf(x.fields[x.fields.length - 1]) : 0; }
  export function NumFields(x: FieldList | undefined): number {
    if (!x) return 0;
    let n = 0;
    for (const field of x.fields) {
      n += field.names.length || 1;
    }
    return n;
  }
}

export namespace ImportSpec {
  export function Pos(x: ImportSpec): Pos { return x.name ? PosOf(x.name) : PosOf(x.path); }
  export function End(x: ImportSpec): Pos { return x.comment ? CommentGroup.End(x.comment) : EndOf(x.path); }
  export function specNode(_x: ImportSpec): void {}
}

export namespace ValueSpec {
  export function Pos(x: ValueSpec): Pos { return PosOf(x.names[0]); }
  export function End(x: ValueSpec): Pos { return x.values.length > 0 ? EndOf(x.values[x.values.length - 1]) : x.type ? EndOf(x.type) : EndOf(x.names[x.names.length - 1]); }
  export function specNode(_x: ValueSpec): void {}
}

export namespace TypeSpec {
  export function Pos(x: TypeSpec): Pos { return PosOf(x.name); }
  export function End(x: TypeSpec): Pos { return EndOf(x.type); }
  export function specNode(_x: TypeSpec): void {}
}

export namespace BadDecl {
  export function Pos(x: BadDecl): Pos { return nodePos(x); }
  export function End(x: BadDecl): Pos { return nodeEnd(x); }
  export function declNode(_x: BadDecl): void {}
}

export namespace GenDecl {
  export function Pos(x: GenDecl): Pos { return nodePos(x); }
  export function End(x: GenDecl): Pos { return nodeEnd(x) || EndOf(x.specs[x.specs.length - 1]); }
  export function declNode(_x: GenDecl): void {}
}

export namespace FuncDecl {
  export function Pos(x: FuncDecl): Pos { return nodePos(x) || PosOf(x.name); }
  export function End(x: FuncDecl): Pos { return x.body ? EndOf(x.body) : EndOf(x.type); }
  export function declNode(_x: FuncDecl): void {}
}

export namespace File {
  export function Pos(x: File): Pos { return nodePos(x) || PosOf(x.name); }
  export function End(x: File): Pos { return nodeEnd(x) || EndOf(x.declarations[x.declarations.length - 1]); }
}

export namespace Package {
  export function Pos(_x: Package): Pos { return 0; }
  export function End(_x: Package): Pos { return 0; }
}

function PosOf(node: AstNode | undefined): Pos {
  switch (node?.kind) {
    case "BadExpr": return BadExpr.Pos(node);
    case "Ident": return Ident.Pos(node);
    case "Ellipsis": return Ellipsis.Pos(node);
    case "BasicLit": return BasicLit.Pos(node);
    case "FuncLit": return FuncLit.Pos(node);
    case "CompositeLit": return CompositeLit.Pos(node);
    case "ParenExpr": return ParenExpr.Pos(node);
    case "SelectorExpr": return SelectorExpr.Pos(node);
    case "IndexExpr": return IndexExpr.Pos(node);
    case "IndexListExpr": return IndexListExpr.Pos(node);
    case "SliceExpr": return SliceExpr.Pos(node);
    case "TypeAssertExpr": return TypeAssertExpr.Pos(node);
    case "CallExpr": return CallExpr.Pos(node);
    case "StarExpr": return StarExpr.Pos(node);
    case "UnaryExpr": return UnaryExpr.Pos(node);
    case "BinaryExpr": return BinaryExpr.Pos(node);
    case "KeyValueExpr": return KeyValueExpr.Pos(node);
    case "ArrayType": return ArrayType.Pos(node);
    case "StructType": return StructType.Pos(node);
    case "FuncType": return FuncType.Pos(node);
    case "InterfaceType": return InterfaceType.Pos(node);
    case "MapType": return MapType.Pos(node);
    case "ChanType": return ChanType.Pos(node);
    case "BadStmt": return BadStmt.Pos(node);
    case "DeclStmt": return DeclStmt.Pos(node);
    case "EmptyStmt": return EmptyStmt.Pos(node);
    case "LabeledStmt": return LabeledStmt.Pos(node);
    case "ExprStmt": return ExprStmt.Pos(node);
    case "SendStmt": return SendStmt.Pos(node);
    case "IncDecStmt": return IncDecStmt.Pos(node);
    case "AssignStmt": return AssignStmt.Pos(node);
    case "GoStmt": return GoStmt.Pos(node);
    case "DeferStmt": return DeferStmt.Pos(node);
    case "ReturnStmt": return ReturnStmt.Pos(node);
    case "BranchStmt": return BranchStmt.Pos(node);
    case "BlockStmt": return BlockStmt.Pos(node);
    case "IfStmt": return IfStmt.Pos(node);
    case "CaseClause": return CaseClause.Pos(node);
    case "SwitchStmt": return SwitchStmt.Pos(node);
    case "TypeSwitchStmt": return TypeSwitchStmt.Pos(node);
    case "CommClause": return CommClause.Pos(node);
    case "SelectStmt": return SelectStmt.Pos(node);
    case "ForStmt": return ForStmt.Pos(node);
    case "RangeStmt": return RangeStmt.Pos(node);
    case "Comment": return Comment.Pos(node);
    case "CommentGroup": return CommentGroup.Pos(node);
    case "Field": return Field.Pos(node);
    case "FieldList": return FieldList.Pos(node);
    case "ImportSpec": return ImportSpec.Pos(node);
    case "ValueSpec": return ValueSpec.Pos(node);
    case "TypeSpec": return TypeSpec.Pos(node);
    case "BadDecl": return BadDecl.Pos(node);
    case "GenDecl": return GenDecl.Pos(node);
    case "FuncDecl": return FuncDecl.Pos(node);
    case "File": return File.Pos(node);
    case "Package": return Package.Pos(node);
    default: return nodePos(node);
  }
}

function EndOf(node: AstNode | undefined): Pos {
  switch (node?.kind) {
    case "BadExpr": return BadExpr.End(node);
    case "Ident": return Ident.End(node);
    case "Ellipsis": return Ellipsis.End(node);
    case "BasicLit": return BasicLit.End(node);
    case "FuncLit": return FuncLit.End(node);
    case "CompositeLit": return CompositeLit.End(node);
    case "ParenExpr": return ParenExpr.End(node);
    case "SelectorExpr": return SelectorExpr.End(node);
    case "IndexExpr": return IndexExpr.End(node);
    case "IndexListExpr": return IndexListExpr.End(node);
    case "SliceExpr": return SliceExpr.End(node);
    case "TypeAssertExpr": return TypeAssertExpr.End(node);
    case "CallExpr": return CallExpr.End(node);
    case "StarExpr": return StarExpr.End(node);
    case "UnaryExpr": return UnaryExpr.End(node);
    case "BinaryExpr": return BinaryExpr.End(node);
    case "KeyValueExpr": return KeyValueExpr.End(node);
    case "ArrayType": return ArrayType.End(node);
    case "StructType": return StructType.End(node);
    case "FuncType": return FuncType.End(node);
    case "InterfaceType": return InterfaceType.End(node);
    case "MapType": return MapType.End(node);
    case "ChanType": return ChanType.End(node);
    case "BadStmt": return BadStmt.End(node);
    case "DeclStmt": return DeclStmt.End(node);
    case "EmptyStmt": return EmptyStmt.End(node);
    case "LabeledStmt": return LabeledStmt.End(node);
    case "ExprStmt": return ExprStmt.End(node);
    case "SendStmt": return SendStmt.End(node);
    case "IncDecStmt": return IncDecStmt.End(node);
    case "AssignStmt": return AssignStmt.End(node);
    case "GoStmt": return GoStmt.End(node);
    case "DeferStmt": return DeferStmt.End(node);
    case "ReturnStmt": return ReturnStmt.End(node);
    case "BranchStmt": return BranchStmt.End(node);
    case "BlockStmt": return BlockStmt.End(node);
    case "IfStmt": return IfStmt.End(node);
    case "CaseClause": return CaseClause.End(node);
    case "SwitchStmt": return SwitchStmt.End(node);
    case "TypeSwitchStmt": return TypeSwitchStmt.End(node);
    case "CommClause": return CommClause.End(node);
    case "SelectStmt": return SelectStmt.End(node);
    case "ForStmt": return ForStmt.End(node);
    case "RangeStmt": return RangeStmt.End(node);
    case "Comment": return Comment.End(node);
    case "CommentGroup": return CommentGroup.End(node);
    case "Field": return Field.End(node);
    case "FieldList": return FieldList.End(node);
    case "ImportSpec": return ImportSpec.End(node);
    case "ValueSpec": return ValueSpec.End(node);
    case "TypeSpec": return TypeSpec.End(node);
    case "BadDecl": return BadDecl.End(node);
    case "GenDecl": return GenDecl.End(node);
    case "FuncDecl": return FuncDecl.End(node);
    case "File": return File.End(node);
    case "Package": return Package.End(node);
    default: return nodeEnd(node);
  }
}

export function astIsExported(name: string): boolean {
  return IsExported(name);
}

export function IsGenerated(file: File): boolean {
  return generator(file)[1];
}

export function generator(file: File): [string, boolean] {
  const packageOffset = file.name?.span?.offset ?? Number.POSITIVE_INFINITY;
  for (const group of file.comments) {
    for (const comment of group.list) {
      if ((comment.span?.offset ?? 0) > packageOffset) {
        break;
      }
      const prefix = "// Code generated ";
      if (comment.text.includes(prefix)) {
        for (const line of comment.text.split("\n")) {
          if (line.startsWith(prefix) && line.endsWith(" DO NOT EDIT.")) {
            return [line.slice(prefix.length, -" DO NOT EDIT.".length), true];
          }
        }
      }
    }
  }
  return ["", false];
}

export function Unparen(e: Expr): Expr {
  let expr = e;
  while (expr.kind === "ParenExpr") {
    expr = expr.expr;
  }
  return expr;
}

export function exportFilter(name: string): boolean {
  return IsExported(name);
}

export function FileExports(src: File): boolean {
  return filterFile(src, exportFilter, true);
}

export function PackageExports(pkg: Package): boolean {
  return filterPackage(pkg, exportFilter, true);
}

export type Filter = (name: string) => boolean;

export function filterIdentList(list: Ident[], f: Filter): Ident[] {
  let j = 0;
  for (const x of list) {
    if (f(x.name)) {
      list[j] = x;
      j += 1;
    }
  }
  list.length = j;
  return list;
}

export function fieldName(x: Expr): Ident | undefined {
  switch (x.kind) {
    case "Ident":
      return x;
    case "SelectorExpr":
      if (x.object.kind === "Ident") {
        return x.selector;
      }
      break;
    case "StarExpr":
      return fieldName(x.expr);
  }
  return undefined;
}

export function filterFieldList(fields: FieldList | undefined, filter: Filter, exportOnly: boolean): boolean {
  if (!fields) {
    return false;
  }
  const list = fields.fields;
  let j = 0;
  let removedFields = false;
  for (const field of list) {
    let keepField = false;
    if (field.names.length === 0) {
      const name = fieldName(field.type);
      keepField = name !== undefined && filter(name.name);
    } else {
      const n = field.names.length;
      field.names = filterIdentList(field.names, filter);
      if (field.names.length < n) {
        removedFields = true;
      }
      keepField = field.names.length > 0;
    }
    if (keepField) {
      if (exportOnly) {
        filterType(field.type, filter, exportOnly);
      }
      list[j] = field;
      j += 1;
    }
  }
  if (j < list.length) {
    removedFields = true;
  }
  list.length = j;
  return removedFields;
}

export function filterCompositeLit(lit: CompositeLit, filter: Filter, exportOnly: boolean): void {
  lit.elements = filterExprList(lit.elements, filter, exportOnly);
}

export function filterExprList(list: Expr[], filter: Filter, exportOnly: boolean): Expr[] {
  let j = 0;
  for (const exp of list) {
    switch (exp.kind) {
      case "CompositeLit":
        filterCompositeLit(exp, filter, exportOnly);
        break;
      case "KeyValueExpr":
        if (exp.key.kind === "Ident" && !filter(exp.key.name)) {
          continue;
        }
        if (exp.value.kind === "CompositeLit") {
          filterCompositeLit(exp.value, filter, exportOnly);
        }
        break;
    }
    list[j] = exp;
    j += 1;
  }
  list.length = j;
  return list;
}

export function filterParamList(fields: FieldList | undefined, filter: Filter, exportOnly: boolean): boolean {
  if (!fields) {
    return false;
  }
  let found = false;
  for (const field of fields.fields) {
    if (filterType(field.type, filter, exportOnly)) {
      found = true;
    }
  }
  return found;
}

export function filterType(typ: Expr | undefined, f: Filter, exportOnly: boolean): boolean {
  if (!typ) {
    return false;
  }
  switch (typ.kind) {
    case "Ident":
      return f(typ.name);
    case "ParenExpr":
      return filterType(typ.expr, f, exportOnly);
    case "ArrayType":
      return filterType(typ.element, f, exportOnly);
    case "StructType":
      filterFieldList(typ.fields, f, exportOnly);
      return typ.fields.fields.length > 0;
    case "FuncType": {
      const b1 = filterParamList(typ.params, f, exportOnly);
      const b2 = filterParamList(typ.results, f, exportOnly);
      return b1 || b2;
    }
    case "InterfaceType":
      filterFieldList(typ.methods, f, exportOnly);
      return typ.methods.fields.length > 0;
    case "MapType": {
      const b1 = filterType(typ.key, f, exportOnly);
      const b2 = filterType(typ.value, f, exportOnly);
      return b1 || b2;
    }
    case "ChanType":
      return filterType(typ.value, f, exportOnly);
  }
  return false;
}

export function filterSpec(spec: Spec, f: Filter, exportOnly: boolean): boolean {
  switch (spec.kind) {
    case "ValueSpec":
      spec.names = filterIdentList(spec.names, f);
      spec.values = filterExprList(spec.values, f, exportOnly);
      if (spec.names.length > 0) {
        if (exportOnly) {
          filterType(spec.type, f, exportOnly);
        }
        return true;
      }
      break;
    case "TypeSpec":
      if (f(spec.name.name)) {
        if (exportOnly) {
          filterType(spec.type, f, exportOnly);
        }
        return true;
      }
      if (!exportOnly) {
        return filterType(spec.type, f, exportOnly);
      }
      break;
  }
  return false;
}

export function filterSpecList(list: Spec[], f: Filter, exportOnly: boolean): Spec[] {
  let j = 0;
  for (const spec of list) {
    if (filterSpec(spec, f, exportOnly)) {
      list[j] = spec;
      j += 1;
    }
  }
  list.length = j;
  return list;
}

export function FilterDecl(decl: Decl, f: Filter): boolean {
  return filterDecl(decl, f, false);
}

export function filterDecl(decl: Decl, f: Filter, exportOnly: boolean): boolean {
  switch (decl.kind) {
    case "GenDecl":
      decl.specs = filterSpecList(decl.specs, f, exportOnly);
      return decl.specs.length > 0;
    case "FuncDecl":
      return f(decl.name.name);
  }
  return false;
}

export function FilterFile(src: File, f: Filter): boolean {
  return filterFile(src, f, false);
}

export function filterFile(src: File, f: Filter, exportOnly: boolean): boolean {
  let j = 0;
  for (const declaration of src.declarations) {
    if (filterDecl(declaration, f, exportOnly)) {
      src.declarations[j] = declaration;
      j += 1;
    }
  }
  src.declarations.length = j;
  return j > 0;
}

export function FilterPackage(pkg: Package, f: Filter): boolean {
  return filterPackage(pkg, f, false);
}

export function filterPackage(pkg: Package, f: Filter, exportOnly: boolean): boolean {
  let hasDecls = false;
  for (const src of pkg.files) {
    if (filterFile(src, f, exportOnly)) {
      hasDecls = true;
    }
  }
  return hasDecls;
}

export type MergeMode = number;

export const FilterFuncDuplicates: MergeMode = 1 << 0;
export const FilterUnassociatedComments: MergeMode = 1 << 1;
export const FilterImportDuplicates: MergeMode = 1 << 2;

export const separator: Comment = { kind: "Comment", text: "//" };

export function nameOf(f: FuncDecl): string {
  if (f.receiver && f.receiver.fields.length === 1) {
    let typ = f.receiver.fields[0]?.type;
    if (typ?.kind === "StarExpr") {
      typ = typ.expr;
    }
    if (typ?.kind === "Ident") {
      return `${typ.name}.${f.name.name}`;
    }
  }
  return f.name.name;
}

export function MergePackageFiles(pkg: Package, mode: MergeMode): File {
  const files = [...pkg.files].sort((left, right) => (left.name?.name ?? "").localeCompare(right.name?.name ?? ""));
  const declarations: Decl[] = [];
  const comments: CommentGroup[] = [];
  const funcs = new Set<string>();
  const imports = new Set<string>();

  for (const file of files) {
    if ((mode & FilterUnassociatedComments) === 0) {
      comments.push(...file.comments);
    }
    for (const declaration of file.declarations) {
      if ((mode & FilterFuncDuplicates) !== 0 && declaration.kind === "FuncDecl") {
        const name = nameOf(declaration);
        if (funcs.has(name)) {
          continue;
        }
        funcs.add(name);
      }
      if ((mode & FilterImportDuplicates) !== 0 && declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
        const before = declaration.specs.length;
        declaration.specs = declaration.specs.filter((spec) => {
          if (spec.kind !== "ImportSpec") return true;
          const key = spec.path.value;
          if (imports.has(key)) return false;
          imports.add(key);
          return true;
        });
        if (before > 0 && declaration.specs.length === 0) {
          continue;
        }
      }
      declarations.push(declaration);
    }
  }

  return {
    kind: "File",
    name: ident(pkg.name),
    declarations,
    imports: declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs.filter((spec): spec is ImportSpec => spec.kind === "ImportSpec") : []),
    unresolved: [],
    comments
  };
}

interface commentPosition {
  Offset: number;
  Line: number;
}

interface positionFileSet {
  Position(pos: Pos): { Offset?: number; offset?: number; Line?: number; line?: number };
}

function sortComments(list: CommentGroup[]): void {
  list.sort((a, b) => CommentGroup.Pos(a) - CommentGroup.Pos(b));
}

export class CommentMap extends Map<AstNode, CommentGroup[]> {
  public addComment(n: AstNode, c: CommentGroup): void {
    let list = this.get(n);
    if (list === undefined || list.length === 0) {
      list = [c];
    } else {
      list = [...list, c];
    }
    this.set(n, list);
  }

  public Update(old: AstNode, replacement: AstNode): AstNode {
    const list = this.get(old);
    if (list !== undefined && list.length > 0) {
      this.delete(old);
      this.set(replacement, [...(this.get(replacement) ?? []), ...list]);
    }
    return replacement;
  }

  public Filter(node: AstNode): CommentMap {
    const umap = new CommentMap();
    Inspect(node, (n) => {
      if (n !== undefined) {
        const groups = this.get(n);
        if (groups !== undefined && groups.length > 0) {
          umap.set(n, groups);
        }
      }
      return true;
    });
    return umap;
  }

  public Comments(): CommentGroup[] {
    const list: CommentGroup[] = [];
    for (const entry of this.values()) {
      list.push(...entry);
    }
    sortComments(list);
    return list;
  }

  public String(): string {
    const nodes = [...this.keys()];
    nodes.sort((a, b) => {
      const r = PosOf(a) - PosOf(b);
      if (r !== 0) {
        return r;
      }
      return EndOf(a) - EndOf(b);
    });

    let out = "CommentMap {\n";
    for (const node of nodes) {
      const comments = this.get(node) ?? [];
      const s = node.kind === "Ident" ? node.name : node.kind;
      out += `\t${commentMapNodeAddress(node)}  ${s.padStart(20)}:  ${summary(comments)}\n`;
    }
    out += "}\n";
    return out;
  }
}

export function nodeList(n: AstNode): AstNode[] {
  const list: AstNode[] = [];
  Inspect(n, (node) => {
    if (node === undefined) {
      return false;
    }
    switch (node.kind) {
      case "CommentGroup":
      case "Comment":
        return false;
      default:
        list.push(node);
        return true;
    }
  });
  return list;
}

export class commentListReader {
  public index = 0;
  public comment: CommentGroup | undefined;
  public pos: commentPosition = { Offset: 0, Line: 0 };
  public end: commentPosition = { Offset: 0, Line: 0 };

  public constructor(
    public fset: unknown,
    public list: CommentGroup[]
  ) {}

  public eol(): boolean {
    return this.index >= this.list.length;
  }

  public next(): void {
    if (!this.eol()) {
      const comment = this.list[this.index];
      if (comment === undefined) {
        return;
      }
      this.comment = comment;
      this.pos = sourcePosition(this.fset, this.comment, false);
      this.end = sourcePosition(this.fset, this.comment, true);
      this.index += 1;
    }
  }
}

export class nodeStack {
  private readonly list: AstNode[] = [];

  public push(n: AstNode): void {
    this.pop(PosOf(n));
    this.list.push(n);
  }

  public pop(pos: Pos): AstNode | undefined {
    let top: AstNode | undefined;
    let i = this.list.length;
    while (i > 0 && EndOf(this.list[i - 1]) <= pos) {
      top = this.list[i - 1];
      i -= 1;
    }
    this.list.length = i;
    return top;
  }
}

export function NewCommentMap(fset: unknown, node: AstNode, comments: CommentGroup[]): CommentMap | undefined {
  if (comments.length === 0) {
    return undefined;
  }

  const cmap = new CommentMap();
  const tmp = comments.slice();
  sortComments(tmp);
  const r = new commentListReader(fset, tmp);
  r.next();

  const nodes: Array<AstNode | undefined> = nodeList(node);
  nodes.push(undefined);

  let p: AstNode | undefined;
  let pend: commentPosition = { Offset: 0, Line: 0 };
  let pg: AstNode | undefined;
  let pgend: commentPosition = { Offset: 0, Line: 0 };
  const stack = new nodeStack();

  for (const q of nodes) {
    let qpos: commentPosition;
    if (q !== undefined) {
      qpos = sourcePosition(fset, q, false);
    } else {
      const infinity = 1 << 30;
      qpos = { Offset: infinity, Line: infinity };
    }

    while (r.comment !== undefined && r.end.Offset <= qpos.Offset) {
      const top = stack.pop(CommentGroup.Pos(r.comment));
      if (top !== undefined) {
        pg = top;
        pgend = sourcePosition(fset, pg, true);
      }

      let assoc: AstNode | undefined;
      if (
        pg !== undefined &&
        (
          pgend.Line === r.pos.Line ||
          (pgend.Line + 1 === r.pos.Line && r.end.Line + 1 < qpos.Line)
        )
      ) {
        assoc = pg;
      } else if (
        p !== undefined &&
        (
          pend.Line === r.pos.Line ||
          (pend.Line + 1 === r.pos.Line && r.end.Line + 1 < qpos.Line) ||
          q === undefined
        )
      ) {
        assoc = p;
      } else {
        if (q === undefined) {
          throw new globalThis.Error("internal error: no comments should be associated with sentinel");
        }
        assoc = q;
      }

      cmap.addComment(assoc, r.comment);
      if (r.eol()) {
        return cmap;
      }
      r.next();
    }

    if (q === undefined) {
      break;
    }
    p = q;
    pend = sourcePosition(fset, p, true);

    if (isCommentNodeGroup(q)) {
      stack.push(q);
    }
  }

  return cmap;
}

export function summary(list: CommentGroup[]): string {
  const maxLen = 40;
  let buf = "";

  commentLoop:
  for (const group of list) {
    for (const comment of group.list) {
      if (buf.length >= maxLen) {
        break commentLoop;
      }
      buf += comment.text;
    }
  }

  if (buf.length > maxLen) {
    buf = `${buf.slice(0, maxLen - 3)}...`;
  }

  return buf.replace(/[\t\n\r]/g, " ");
}

const commentMapIds = new WeakMap<AstNode, number>();
let nextCommentMapID = 1;

function commentMapNodeAddress(node: AstNode): string {
  let id = commentMapIds.get(node);
  if (id === undefined) {
    id = nextCommentMapID;
    nextCommentMapID += 1;
    commentMapIds.set(node, id);
  }
  return `0x${id.toString(16)}`;
}

function sourcePosition(fset: unknown, node: AstNode, end: boolean): commentPosition {
  const pos = end ? EndOf(node) : PosOf(node);
  const fsetPosition = (fset as Partial<positionFileSet> | undefined)?.Position;
  if (typeof fsetPosition === "function") {
    const got = fsetPosition.call(fset, pos);
    const offset = got.Offset ?? got.offset ?? pos;
    const line = got.Line ?? got.line ?? fallbackLine(node, end);
    return { Offset: offset, Line: line };
  }
  return { Offset: pos, Line: fallbackLine(node, end) };
}

function fallbackLine(node: AstNode, end: boolean): number {
  if (node.kind === "CommentGroup") {
    const comment = end ? node.list[node.list.length - 1] : node.list[0];
    return comment ? fallbackLine(comment, end) : 0;
  }
  if (node.kind === "Comment" && end) {
    return (node.span?.line ?? 0) + countLinesAfterFirst(node.text);
  }
  return node.span?.line ?? 0;
}

function countLinesAfterFirst(text: string): number {
  let n = 0;
  for (const ch of text) {
    if (ch === "\n") {
      n += 1;
    }
  }
  return n;
}

function isCommentNodeGroup(node: AstNode): boolean {
  switch (node.kind) {
    case "File":
    case "Field":
    case "ImportSpec":
    case "ValueSpec":
    case "TypeSpec":
    case "BadDecl":
    case "GenDecl":
    case "FuncDecl":
    case "BadStmt":
    case "DeclStmt":
    case "EmptyStmt":
    case "LabeledStmt":
    case "ExprStmt":
    case "AssignStmt":
    case "IncDecStmt":
    case "ReturnStmt":
    case "BranchStmt":
    case "BlockStmt":
    case "IfStmt":
    case "CaseClause":
    case "CommClause":
    case "SwitchStmt":
    case "TypeSwitchStmt":
    case "SelectStmt":
    case "ForStmt":
    case "RangeStmt":
    case "DeferStmt":
    case "GoStmt":
    case "SendStmt":
      return true;
    default:
      return false;
  }
}

export class posSpan {
  public constructor(
    public Start: Pos,
    public End: Pos
  ) {}
}

export class cgPos {
  public constructor(
    public left: boolean,
    public cg: CommentGroup
  ) {}
}

export function SortImports(fset: unknown, f: File): void {
  for (const decl of f.declarations) {
    if (decl.kind !== "GenDecl" || decl.token !== TokenKind.Import) {
      break;
    }

    if (!decl.grouped) {
      continue;
    }

    let i = 0;
    const specs: Spec[] = [];
    for (let j = 0; j < decl.specs.length; j += 1) {
      const spec = decl.specs[j];
      const prev = decl.specs[j - 1];
      if (spec && prev && j > i && lineAt(fset, spec) > 1 + lineAt(fset, prev)) {
        specs.push(...sortSpecs(fset, f, decl, decl.specs.slice(i, j)));
        i = j;
      }
    }
    specs.push(...sortSpecs(fset, f, decl, decl.specs.slice(i)));
    decl.specs = specs;
  }

  f.imports.length = 0;
  for (const decl of f.declarations) {
    if (decl.kind === "GenDecl" && decl.token === TokenKind.Import) {
      for (const spec of decl.specs) {
        if (spec.kind === "ImportSpec") {
          f.imports.push(spec);
        }
      }
    }
  }
}

export function lineAt(_fset: unknown, pos: Pos | SourceSpan | Node | undefined): number {
  if (typeof pos === "number") {
    return Number.isFinite(pos) ? pos : 0;
  }
  if (pos && "kind" in pos) {
    return pos.span?.line ?? 0;
  }
  return pos?.line ?? 0;
}

export function importPath(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  try {
    return unquoteGoString(s.path.value);
  } catch {
    return "";
  }
}

export function importName(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  return s.name?.name ?? "";
}

export function importComment(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  return CommentGroup.Text(s.comment);
}

export function collapse(prev: Spec, next: Spec): boolean {
  if (importPath(next) !== importPath(prev) || importName(next) !== importName(prev)) {
    return false;
  }
  return prev.kind === "ImportSpec" && prev.comment === undefined;
}

export function sortSpecs(fset: unknown, f: File, d: GenDecl, specs: Spec[]): Spec[] {
  if (specs.length <= 1) {
    return specs;
  }

  const pos = new Array<posSpan>(specs.length);
  for (let i = 0; i < specs.length; i += 1) {
    const spec = specs[i];
    pos[i] = new posSpan(nodePos(spec), nodeEnd(spec));
  }

  const begSpecs = pos[0]?.Start ?? 0;
  const endSpecs = pos[pos.length - 1]?.End ?? 0;
  const beg = lineStart(f, begSpecs);
  const endLine = lineAt(fset, specs[specs.length - 1]);
  let end: Pos;
  if (endLine === maxLine(f)) {
    end = endSpecs;
  } else {
    end = lineStart(f, endLine + 1);
  }
  let first = f.comments.length;
  let last = -1;
  for (let i = 0; i < f.comments.length; i += 1) {
    const group = f.comments[i];
    if (!group) {
      continue;
    }
    if (nodeEnd(group) >= end) {
      break;
    }
    if (beg <= nodePos(group)) {
      if (i < first) {
        first = i;
      }
      if (i > last) {
        last = i;
      }
    }
  }

  let comments: CommentGroup[] = [];
  if (last >= 0) {
    comments = f.comments.slice(first, last + 1);
  }

  const importComments = new Map<ImportSpec, cgPos[]>();
  let specIndex = 0;
  for (const group of comments) {
    while (specIndex + 1 < specs.length && (pos[specIndex + 1]?.Start ?? 0) <= nodePos(group)) {
      specIndex += 1;
    }
    let left = false;
    if (specIndex === 0 && (pos[specIndex]?.Start ?? 0) > nodePos(group)) {
      left = true;
    } else if (
      specIndex + 1 < specs.length &&
      lineAt(fset, specs[specIndex]) + 1 === lineAt(fset, group)
    ) {
      specIndex += 1;
      left = true;
    }
    const spec = specs[specIndex];
    if (spec?.kind === "ImportSpec") {
      const list = importComments.get(spec) ?? [];
      list.push(new cgPos(left, group));
      importComments.set(spec, list);
    }
  }

  specs.sort((a, b) => {
    let result = importPath(a).localeCompare(importPath(b));
    if (result !== 0) {
      return result;
    }
    result = importName(a).localeCompare(importName(b));
    if (result !== 0) {
      return result;
    }
    return importComment(a).localeCompare(importComment(b));
  });

  const deduped: Spec[] = [];
  for (let i = 0; i < specs.length; i += 1) {
    const spec = specs[i];
    const next = specs[i + 1];
    if (spec && (!next || !collapse(spec, next))) {
      deduped.push(spec);
    } else if (spec) {
      if (lineAt(fset, spec) !== lineAt(fset, d)) {
        mergeLine(f, lineAt(fset, spec));
      }
    }
  }
  specs = deduped;

  for (let i = 0; i < specs.length; i += 1) {
    const spec = specs[i];
    const span = pos[i];
    if (spec?.kind !== "ImportSpec" || !span) {
      continue;
    }
    if (spec.name?.span) {
      spec.name.span = moveSpan(spec.name.span, span.Start);
    }
    updateBasicLitPos(spec.path, span.Start);
    for (const group of importComments.get(spec) ?? []) {
      for (const comment of group.cg.list) {
        if (comment.span) {
          comment.span = moveSpan(comment.span, group.left ? span.Start - 1 : span.End);
        }
      }
    }
  }

  comments.sort((a, b) => nodePos(a) - nodePos(b));
  return specs;
}

export function updateBasicLitPos(lit: BasicLit, pos: Pos): void {
  if (!lit.span) {
    return;
  }
  lit.span = moveSpan(lit.span, pos);
}

export type ObjKind = number;

export const Bad: ObjKind = 0;
export const Pkg: ObjKind = 1;
export const Con: ObjKind = 2;
export const Typ: ObjKind = 3;
export const Var: ObjKind = 4;
export const Fun: ObjKind = 5;
export const Lbl: ObjKind = 6;

export const objKindStrings = ["bad", "package", "const", "type", "var", "func", "label"];

export class Scope {
  public Objects: Map<string, Object>;

  public constructor(
    public Outer?: Scope,
    objects?: Map<string, Object>
  ) {
    this.Objects = objects ?? new Map<string, Object>();
  }

  public Lookup(name: string): Object | undefined {
    return this.Objects.get(name);
  }

  public Insert(obj: Object): Object | undefined {
    const alt = this.Objects.get(obj.Name);
    if (alt === undefined) {
      this.Objects.set(obj.Name, obj);
    }
    return alt;
  }

  public String(): string {
    let buf = "scope {";
    if (this !== undefined && this.Objects.size > 0) {
      buf += "\n";
      for (const obj of this.Objects.values()) {
        buf += `\t${ObjKind.String(obj.Kind)} ${obj.Name}\n`;
      }
    }
    buf += "}\n";
    return buf;
  }
}

export function NewScope(outer?: Scope): Scope {
  return new Scope(outer, new Map<string, Object>());
}

export class Object {
  public Decl: unknown;
  public Data: unknown;
  public Type: unknown;

  public constructor(
    public Kind: ObjKind,
    public Name: string
  ) {}

  public Pos(): Pos {
    const name = this.Name;
    const decl = this.Decl;
    if (isField(decl)) {
      for (const n of decl.names) {
        if (n.name === name) {
          return nodePos(n);
        }
      }
    } else if (isImportSpec(decl)) {
      if (decl.name && decl.name.name === name) {
        return nodePos(decl.name);
      }
      return nodePos(decl.path);
    } else if (isValueSpec(decl)) {
      for (const n of decl.names) {
        if (n.name === name) {
          return nodePos(n);
        }
      }
    } else if (isTypeSpec(decl)) {
      if (decl.name.name === name) {
        return nodePos(decl.name);
      }
    } else if (isFuncDecl(decl)) {
      if (decl.name.name === name) {
        return nodePos(decl.name);
      }
    } else if (isLabeledStmt(decl)) {
      if (decl.label.name === name) {
        return nodePos(decl.label);
      }
    } else if (isAssignStmt(decl)) {
      for (const x of decl.lhs) {
        if (x.kind === "Ident" && x.name === name) {
          return nodePos(x);
        }
      }
    } else if (decl instanceof Scope) {
    }
    return 0;
  }
}

export function NewObj(kind: ObjKind, name: string): Object {
  return new Object(kind, name);
}

export namespace ObjKind {
  export function String(kind: ObjKind): string {
    return objKindStrings[kind] ?? "";
  }
}

export type Importer = (imports: Map<string, Object>, path: string) => [Object | undefined, Error | undefined];

export class pkgBuilder {
  public fset: unknown;
  public errors: Error[] = [];

  public error(pos: Pos, msg: string): void {
    this.errors.push(new globalThis.Error(`${pos}: ${msg}`));
  }

  public errorf(pos: Pos, format: string, ...args: unknown[]): void {
    this.error(pos, formatPrintf(format, args));
  }

  public declare(scope: Scope, altScope: Scope | undefined, obj: Object): void {
    let alt = scope.Insert(obj);
    if (alt === undefined && altScope !== undefined) {
      alt = altScope.Lookup(obj.Name);
    }
    if (alt !== undefined) {
      let prevDecl = "";
      const pos = alt.Pos();
      if (pos !== 0) {
        prevDecl = `\n\tprevious declaration at ${pos}`;
      }
      this.error(obj.Pos(), `${obj.Name} redeclared in this block${prevDecl}`);
    }
  }
}

export function resolve(scope: Scope | undefined, id: Ident): boolean {
  for (; scope !== undefined; scope = scope.Outer) {
    const obj = scope.Lookup(id.name);
    if (obj !== undefined) {
      id.Obj = obj;
      return true;
    }
  }
  return false;
}

export function NewPackage(_fset: unknown, files: Map<string, File> | Record<string, File>, importer?: Importer, universe?: Scope): [Package, Error | undefined] {
  const p = new pkgBuilder();
  p.fset = _fset;

  let pkgName = "";
  const pkgScope = NewScope(universe);
  for (const file of fileValues(files)) {
    const name = file.name?.name ?? "";
    switch (true) {
      case pkgName === "":
        pkgName = name;
        break;
      case name !== pkgName:
        p.errorf(nodePos(file.name), "package %s; expected %s", name, pkgName);
        continue;
    }

    for (const obj of file.scope?.Objects.values() ?? []) {
      p.declare(pkgScope, undefined, obj);
    }
  }

  const imports = new Map<string, Object>();

  for (const file of fileValues(files)) {
    if ((file.name?.name ?? "") !== pkgName) {
      continue;
    }

    let importErrors = false;
    const fileScope = NewScope(pkgScope);
    for (const spec of file.imports) {
      if (importer === undefined) {
        importErrors = true;
        continue;
      }
      const path = importPath(spec);
      const [pkg, err] = importer(imports, path);
      if (err !== undefined || pkg === undefined) {
        p.errorf(nodePos(spec.path), "could not import %s (%s)", path, err?.message ?? "unknown error");
        importErrors = true;
        continue;
      }

      let name = pkg.Name;
      if (spec.name !== undefined) {
        name = spec.name.name;
      }

      if (name === ".") {
        if (pkg.Data instanceof Scope) {
          for (const obj of pkg.Data.Objects.values()) {
            p.declare(fileScope, pkgScope, obj);
          }
        }
      } else if (name !== "_") {
        const obj = NewObj(Pkg, name);
        obj.Decl = spec;
        obj.Data = pkg.Data;
        p.declare(fileScope, pkgScope, obj);
      }
    }

    if (importErrors) {
      pkgScope.Outer = undefined;
    }
    let i = 0;
    for (const id of file.unresolved) {
      if (!resolve(fileScope, id)) {
        p.errorf(nodePos(id), "undeclared name: %s", id.name);
        file.unresolved[i] = id;
        i += 1;
      }
    }
    file.unresolved.length = i;
    pkgScope.Outer = universe;
  }

  p.errors.sort((a, b) => a.message.localeCompare(b.message));
  return [{ kind: "Package", name: pkgName, scope: pkgScope, imports, files: fileValues(files) }, p.errors[0]];
}

export type FieldFilter = (name: string, value: unknown) => boolean;

export function NotNilFilter(_name: string, value: unknown): boolean {
  switch (kindOf(value)) {
    case "chan":
    case "func":
    case "interface":
    case "map":
    case "pointer":
    case "slice":
      return value !== null && value !== undefined;
  }
  return true;
}

export interface Writer {
  write(data: string): unknown;
}

export function Fprint(w: Writer, fset: unknown, x: unknown, f?: FieldFilter): Error | undefined {
  return fprint(w, fset, x, f);
}

export function fprint(w: Writer, fset: unknown, x: unknown, f?: FieldFilter): Error | undefined {
  const p = new printer(w, fset, f);
  try {
    if (x === null || x === undefined) {
      p.printf("nil\n");
      return undefined;
    }
    p.print(x);
    p.printf("\n");
    return undefined;
  } catch (err) {
    if (err instanceof localError) {
      return err.err;
    }
    throw err;
  }
}

export function Print(fset: unknown, x: unknown): Error | undefined {
  return Fprint(defaultPrintWriter, fset, x, NotNilFilter);
}

export class printer {
  public ptrmap = new WeakMap<object, number>();
  public indent = 0;
  public last = "\n";
  public line = 0;

  public constructor(
    public output: Writer,
    public fset: unknown,
    public filter?: FieldFilter
  ) {}

  public Write(data: string | Uint8Array): [number, Error | undefined] {
    const text = typeof data === "string" ? data : new TextDecoder().decode(data);
    let n = 0;
    try {
      for (let i = 0; i < text.length; i += 1) {
        const b = text[i] ?? "";
        if (b === "\n") {
          this.writeRaw(text.slice(n, i + 1));
          n = i + 1;
          this.line += 1;
        } else if (this.last === "\n") {
          this.writeRaw(`${this.line.toString().padStart(6, " ")}  `);
          for (let j = this.indent; j > 0; j -= 1) {
            this.writeRaw(indent);
          }
        }
        this.last = b;
      }
      if (text.length > n) {
        this.writeRaw(text.slice(n));
        n = text.length;
      }
      return [n, undefined];
    } catch (err) {
      return [n, err instanceof Error ? err : new globalThis.Error(String(err))];
    }
  }

  public printf(format: string, ...args: unknown[]): void {
    const [, err] = this.Write(formatPrintf(format, args));
    if (err) {
      throw new localError(err);
    }
  }

  public print(x: unknown): void {
    if (!NotNilFilter("", x)) {
      this.printf("nil");
      return;
    }

    switch (kindOf(x)) {
      case "interface":
        this.print(x);
        return;

      case "map": {
        const map = x as Map<unknown, unknown>;
        this.printf("%s (len = %d) {", "Map", map.size);
        if (map.size > 0) {
          this.indent += 1;
          this.printf("\n");
          for (const [key, value] of map) {
            this.print(key);
            this.printf(": ");
            this.print(value);
            this.printf("\n");
          }
          this.indent -= 1;
        }
        this.printf("}");
        return;
      }

      case "pointer": {
        const ptr = x as object;
        this.printf("*");
        const line = this.ptrmap.get(ptr);
        if (line !== undefined) {
          this.printf("(obj @ %d)", line);
        } else {
          this.ptrmap.set(ptr, this.line);
          this.printObject(ptr);
        }
        return;
      }

      case "array":
      case "slice": {
        if (x instanceof Uint8Array) {
          this.printf("%#q", Array.from(x));
          return;
        }
        const list = x as readonly unknown[];
        this.printf("%s (len = %d) {", Array.isArray(x) ? "Array" : "Slice", list.length);
        if (list.length > 0) {
          this.indent += 1;
          this.printf("\n");
          for (let i = 0; i < list.length; i += 1) {
            this.printf("%d: ", i);
            this.print(list[i]);
            this.printf("\n");
          }
          this.indent -= 1;
        }
        this.printf("}");
        return;
      }

      case "struct":
        this.printObject(x as object);
        return;

      default:
        if (typeof x === "string") {
          this.printf("%q", x);
          return;
        }
        this.printf("%v", x);
    }
  }

  private writeRaw(data: string): void {
    this.output.write(data);
  }

  private printObject(x: object): void {
    const fields = globalThis.Object.entries(x);
    this.printf("%s {", objectTypeName(x));
    this.indent += 1;
    let first = true;
    for (const [name, value] of fields) {
      if (this.filter === undefined || this.filter(name, value)) {
        if (first) {
          this.printf("\n");
          first = false;
        }
        this.printf("%s: ", name);
        this.print(value);
        this.printf("\n");
      }
    }
    this.indent -= 1;
    this.printf("}");
  }
}

export const indent = ".  ";

export class localError extends Error {
  public constructor(public err: Error) {
    super(err.message);
  }
}

function kindOf(value: unknown): "array" | "chan" | "func" | "interface" | "map" | "pointer" | "slice" | "struct" | "other" {
  if (value === null || value === undefined) {
    return "interface";
  }
  if (value instanceof Map) {
    return "map";
  }
  if (value instanceof Uint8Array) {
    return "slice";
  }
  if (Array.isArray(value)) {
    return "array";
  }
  if (typeof value === "function") {
    return "func";
  }
  if (typeof value === "object") {
    return "struct";
  }
  return "other";
}

function formatPrintf(format: string, args: unknown[]): string {
  let index = 0;
  return format.replace(/%#q|%[sdvq%]/g, (verb) => {
    if (verb === "%%") {
      return "%";
    }
    const arg = args[index];
    index += 1;
    switch (verb) {
      case "%#q":
      case "%q":
        return quoteValue(arg);
      case "%d":
        return typeof arg === "bigint" ? arg.toString() : String(Math.trunc(Number(arg)));
      case "%s":
        return String(arg);
      case "%v":
      default:
        return formatValue(arg);
    }
  });
}

function quoteValue(value: unknown): string {
  if (typeof value === "string") {
    return JSON.stringify(value);
  }
  if (value instanceof Uint8Array) {
    return JSON.stringify(Array.from(value));
  }
  return JSON.stringify(value);
}

function formatValue(value: unknown): string {
  if (typeof value === "bigint") {
    return value.toString();
  }
  if (typeof value === "string") {
    return value;
  }
  if (value === undefined) {
    return "<undefined>";
  }
  return String(value);
}

function objectTypeName(value: object): string {
  if ("kind" in value && typeof value.kind === "string") {
    return value.kind;
  }
  const name = value.constructor?.name;
  return name && name !== "Object" ? name : "object";
}

const defaultPrintWriter: Writer = {
  write(data: string): void {
    const maybeProcess = globalThis as typeof globalThis & { process?: { stdout?: { write?: (chunk: string) => unknown } } };
    if (maybeProcess.process?.stdout?.write) {
      maybeProcess.process.stdout.write(data);
      return;
    }
    globalThis.console?.log(data);
  }
};

function fileValues(files: Map<string, File> | Record<string, File>): File[] {
  if (files instanceof Map) {
    return [...files.values()];
  }
  return globalThis.Object.values(files);
}

function isField(value: unknown): value is Field {
  return isNodeKind(value, "Field");
}

function isImportSpec(value: unknown): value is ImportSpec {
  return isNodeKind(value, "ImportSpec");
}

function isValueSpec(value: unknown): value is ValueSpec {
  return isNodeKind(value, "ValueSpec");
}

function isTypeSpec(value: unknown): value is TypeSpec {
  return isNodeKind(value, "TypeSpec");
}

function isFuncDecl(value: unknown): value is FuncDecl {
  return isNodeKind(value, "FuncDecl");
}

function isLabeledStmt(value: unknown): value is LabeledStmt {
  return isNodeKind(value, "LabeledStmt");
}

function isAssignStmt(value: unknown): value is AssignStmt {
  return isNodeKind(value, "AssignStmt");
}

function isNodeKind<T extends NodeKind>(value: unknown, kind: T): value is Extract<AstNode, { kind: T }> {
  return typeof value === "object" && value !== null && "kind" in value && value.kind === kind;
}

function nodePos(node: Node | CommentGroup | undefined): Pos {
  return node?.span?.offset ?? 0;
}

function nodeEnd(node: Node | undefined): Pos {
  if (!node?.span) {
    return 0;
  }
  return node.span.offset + node.span.length;
}

function moveSpan(span: SourceSpan, offset: Pos): SourceSpan {
  return {
    ...span,
    offset
  };
}

function lineStart(file: File, line: number): Pos {
  for (const node of childNodes(file)) {
    const span = node.span;
    if (span && span.line === line) {
      return span.offset - Math.max(0, span.column - 1);
    }
  }
  return line;
}

function maxLine(file: File): number {
  let line = 0;
  walk(file, (node) => {
    if (node.span) {
      line = Math.max(line, node.span.line);
    }
  });
  return line;
}

function mergeLine(_file: File, _line: number): void {
}

function unquoteGoString(value: string): string {
  if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
    return value.slice(1, -1).replace(/\r/g, "");
  }
  if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
    return decodeGoEscaped(value.slice(1, -1));
  }
  throw new globalThis.Error("invalid quoted string");
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
      case "x":
        decoded += codePointFromEscape(value.slice(index + 1, index + 3), 16);
        index += 2;
        break;
      case "u":
        decoded += codePointFromEscape(value.slice(index + 1, index + 5), 16);
        index += 4;
        break;
      case "U":
        decoded += codePointFromEscape(value.slice(index + 1, index + 9), 16);
        index += 8;
        break;
      default:
        if (/^[0-7]$/u.test(next)) {
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
  if (!Number.isFinite(value)) {
    return "";
  }
  try {
    return String.fromCodePoint(value);
  } catch {
    return "";
  }
}

export class Directive {
  public constructor(
    public Tool: string,
    public Name: string,
    public Args: string,
    public Slash: Pos,
    public ArgsPos: Pos
  ) {}

  public Pos(): Pos {
    return this.Slash;
  }

  public End(): Pos {
    return this.ArgsPos + this.Args.length;
  }

  public ParseArgs(): [DirectiveArg[], Error | undefined] {
    const args = new directiveScanner(this.Args, this.ArgsPos);
    const list: DirectiveArg[] = [];
    for (args.skipSpace(); args.str !== ""; args.skipSpace()) {
      let arg: string;
      const argPos = args.pos;
      switch (args.str[0]) {
        default:
          arg = args.takeNonSpace();
          break;
        case "`":
        case "\"": {
          const quoted = quotedPrefix(args.str);
          if (!quoted) {
            return [[], new globalThis.Error(`invalid quoted string in //${this.Tool}:${this.Name}: ${args.str}`)];
          }
          arg = unquoteDirectiveArg(args.take(quoted.length));
          if (args.str !== "" && !/^\s/u.test(args.str)) {
            return [[], new globalThis.Error(`invalid quoted string in //${this.Tool}:${this.Name}: ${args.str}`)];
          }
          break;
        }
      }
      list.push(new DirectiveArg(arg, argPos));
    }
    return [list, undefined];
  }
}

export class DirectiveArg {
  public constructor(
    public Arg: string,
    public Pos: Pos
  ) {}
}

export function ParseDirective(pos: Pos, c: string): [Directive, boolean] {
  if (!(c.length >= 3 && c[0] === "/" && c[1] === "/" && isalnum(c[2] ?? ""))) {
    return [new Directive("", "", "", 0, 0), false];
  }

  const buf = new directiveScanner(c, pos);
  buf.skip("//".length);

  const colon = buf.str.indexOf(":");
  if (colon <= 0 || colon + 1 >= buf.str.length) {
    return [new Directive("", "", "", 0, 0), false];
  }
  for (let i = 0; i <= colon + 1; i += 1) {
    if (i === colon) {
      continue;
    }
    if (!isalnum(buf.str[i] ?? "")) {
      return [new Directive("", "", "", 0, 0), false];
    }
  }
  const tool = buf.take(colon);
  buf.skip(":".length);

  const name = buf.takeNonSpace();
  buf.skipSpace();
  const argsPos = buf.pos;
  const args = trimRightUnicodeSpace(buf.str);

  return [new Directive(tool, name, args, pos, argsPos), true];
}

export function isalnum(b: string): boolean {
  return ("a" <= b && b <= "z") || ("0" <= b && b <= "9");
}

class directiveScanner {
  public constructor(
    public str: string,
    public pos: Pos
  ) {}

  public skip(n: number): void {
    this.pos += n;
    this.str = this.str.slice(n);
  }

  public take(n: number): string {
    const res = this.str.slice(0, n);
    this.skip(n);
    return res;
  }

  public takeNonSpace(): string {
    const match = /\s/u.exec(this.str);
    let i = match?.index ?? -1;
    if (i === -1) {
      i = this.str.length;
    }
    return this.take(i);
  }

  public skipSpace(): void {
    const trim = this.str.replace(/^\s+/u, "");
    this.skip(this.str.length - trim.length);
  }
}

function quotedPrefix(text: string): string | undefined {
  const quote = text[0];
  if (quote === "`") {
    const end = text.indexOf("`", 1);
    return end >= 0 ? text.slice(0, end + 1) : undefined;
  }
  if (quote !== "\"") return undefined;
  let escaped = false;
  for (let index = 1; index < text.length; index += 1) {
    const char = text[index] ?? "";
    if (escaped) {
      escaped = false;
      continue;
    }
    if (char === "\\") {
      escaped = true;
      continue;
    }
    if (char === "\"") {
      return text.slice(0, index + 1);
    }
  }
  return undefined;
}

function unquoteDirectiveArg(text: string): string {
  if (text.startsWith("`")) {
    return text.slice(1, -1);
  }
  return JSON.parse(text) as string;
}

function trimRightUnicodeSpace(text: string): string {
  return text.replace(/\s+$/u, "");
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

export interface Visitor {
  Visit(node: AstNode | undefined): Visitor | undefined;
}

export function walkList<N extends AstNode>(v: Visitor, list: N[]): void {
  for (const node of list) {
    Walk(v, node);
  }
}

export function Walk(v: Visitor, node: AstNode): void {
  v = v.Visit(node) as Visitor;
  if (v === undefined) {
    return;
  }

  for (const child of childNodes(node)) {
    Walk(v, child);
  }

  v.Visit(undefined);
}

export class inspector implements Visitor {
  public constructor(private readonly f: (node: AstNode | undefined) => boolean) {}

  public Visit(node: AstNode | undefined): Visitor | undefined {
    if (this.f(node)) {
      return this;
    }
    return undefined;
  }
}

export function Inspect(node: AstNode, f: (node: AstNode | undefined) => boolean): void {
  Walk(new inspector(f), node);
}

export function* Preorder(root: AstNode): IterableIterator<AstNode> {
  yield root;
  for (const child of childNodes(root)) {
    yield* Preorder(child);
  }
}

export function PreorderStack(root: AstNode, stack: AstNode[], f: (node: AstNode, stack: AstNode[]) => boolean): void {
  const before = stack.length;
  Inspect(root, (node) => {
    if (node !== undefined) {
      if (!f(node, stack)) {
        return false;
      }
      stack.push(node);
    } else {
      stack.length = stack.length - 1;
    }
    return true;
  });
  if (stack.length !== before) {
    throw new globalThis.Error("push/pop mismatch");
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
      return [...(node.doc ? [node.doc] : []), ...(node.name ? [node.name] : []), node.path, ...(node.comment ? [node.comment] : [])];
    case "ValueSpec":
      return [...node.names, ...(node.type ? [node.type] : []), ...node.values];
    case "TypeSpec":
      return [node.name, ...(node.typeParams ? [node.typeParams] : []), node.type];
    case "FuncType":
      return [...(node.typeParams ? [node.typeParams] : []), node.params, ...(node.results ? [node.results] : [])];
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
    case "IndexListExpr":
      return [node.object, ...node.indices];
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
