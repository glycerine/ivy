import { SourceSpan } from "../../diagnostics.js";
import { TokenKind } from "../../front/token.js";
import { Token as GoToken, type File as TokenFile, type Pos as TokenPos } from "../token/index.js";

export type Pos = TokenPos;
export type ChanDir = number;

export const SEND: ChanDir = 1 << 0;
export const RECV: ChanDir = 1 << 1;

export interface Node {
  kind: NodeKind;
  span?: SourceSpan;
  Pos?: () => Pos;
  End?: () => Pos;
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
  NamePos?: Pos;
  Name?: string;
  name: string;
  Obj?: Object;
}

export interface BadExpr extends Node {
  kind: "BadExpr";
  From?: Pos;
  To?: Pos;
}

export interface Ellipsis extends Node {
  kind: "Ellipsis";
  Ellipsis?: Pos;
  Elt?: Expr;
  element?: Expr;
}

export interface BasicLit extends Node {
  kind: "BasicLit";
  ValuePos?: Pos;
  ValueEnd?: Pos;
  Kind?: GoToken;
  Value?: string;
  token: TokenKind.IntLiteral | TokenKind.FloatLiteral | TokenKind.ImagLiteral | TokenKind.RuneLiteral | TokenKind.StringLiteral;
  value: string;
}

export interface FuncLit extends Node {
  kind: "FuncLit";
  Type?: FuncType;
  Body?: BlockStmt;
  type: FuncType;
  body: BlockStmt;
}

export interface CompositeLit extends Node {
  kind: "CompositeLit";
  Type?: Expr;
  Lbrace?: Pos;
  Elts?: Expr[];
  Rbrace?: Pos;
  Incomplete?: boolean;
  type?: Expr;
  elements: Expr[];
}

export interface ParenExpr extends Node {
  kind: "ParenExpr";
  Lparen?: Pos;
  X?: Expr;
  Rparen?: Pos;
  expr: Expr;
}

export interface SelectorExpr extends Node {
  kind: "SelectorExpr";
  X?: Expr;
  Sel?: Ident;
  object: Expr;
  selector: Ident;
}

export interface IndexExpr extends Node {
  kind: "IndexExpr";
  X?: Expr;
  Lbrack?: Pos;
  Index?: Expr;
  Rbrack?: Pos;
  object: Expr;
  index: Expr;
}

export interface IndexListExpr extends Node {
  kind: "IndexListExpr";
  X?: Expr;
  Lbrack?: Pos;
  Indices?: Expr[];
  Rbrack?: Pos;
  object: Expr;
  indices: Expr[];
}

export interface SliceExpr extends Node {
  kind: "SliceExpr";
  X?: Expr;
  Lbrack?: Pos;
  Low?: Expr;
  High?: Expr;
  Max?: Expr;
  Slice3?: boolean;
  Rbrack?: Pos;
  object: Expr;
  low?: Expr;
  high?: Expr;
  max?: Expr;
}

export interface TypeAssertExpr extends Node {
  kind: "TypeAssertExpr";
  X?: Expr;
  Lparen?: Pos;
  Type?: Expr;
  Rparen?: Pos;
  object: Expr;
  type?: Expr;
  typeSwitch: boolean;
}

export interface CallExpr extends Node {
  kind: "CallExpr";
  Fun?: Expr;
  Lparen?: Pos;
  Args?: Expr[];
  Ellipsis?: boolean | Pos;
  Rparen?: Pos;
  fun: Expr;
  args: Expr[];
  ellipsis: boolean;
}

export interface StarExpr extends Node {
  kind: "StarExpr";
  Star?: Pos;
  X?: Expr;
  expr: Expr;
}

export interface UnaryExpr extends Node {
  kind: "UnaryExpr";
  OpPos?: Pos;
  Op?: GoToken;
  X?: Expr;
  op: TokenKind.Plus | TokenKind.Minus | TokenKind.Bang | TokenKind.Caret | TokenKind.Tilde | TokenKind.Amp | TokenKind.Arrow;
  expr: Expr;
}

export interface BinaryExpr extends Node {
  kind: "BinaryExpr";
  X?: Expr;
  OpPos?: Pos;
  Op?: GoToken;
  Y?: Expr;
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
  Key?: Expr;
  Colon?: Pos;
  Value?: Expr;
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
  Lbrack?: Pos;
  Len?: Expr;
  Elt?: Expr;
  length?: Expr;
  element: Expr;
  inferredLength: boolean;
}

export interface StructType extends Node {
  kind: "StructType";
  Struct?: Pos;
  Fields?: FieldList;
  Incomplete?: boolean;
  fields: FieldList;
}

export interface FuncType extends Node {
  kind: "FuncType";
  Func?: Pos;
  TypeParams?: FieldList;
  Params?: FieldList;
  Results?: FieldList;
  typeParams?: FieldList;
  params: FieldList;
  results?: FieldList;
}

export interface InterfaceType extends Node {
  kind: "InterfaceType";
  Interface?: Pos;
  Methods?: FieldList;
  Incomplete?: boolean;
  methods: FieldList;
}

export interface MapType extends Node {
  kind: "MapType";
  Map?: Pos;
  Key?: Expr;
  Value?: Expr;
  key: Expr;
  value: Expr;
}

export interface ChanType extends Node {
  kind: "ChanType";
  Begin?: Pos;
  Arrow?: Pos;
  Dir?: ChanDir | string;
  Value?: Expr;
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
  From?: Pos;
  To?: Pos;
}

export interface DeclStmt extends Node {
  kind: "DeclStmt";
  Decl?: Decl;
  decl: Decl;
}

export interface EmptyStmt extends Node {
  kind: "EmptyStmt";
  Semicolon?: Pos;
  Implicit?: boolean;
  implicit: boolean;
}

export interface LabeledStmt extends Node {
  kind: "LabeledStmt";
  Label?: Ident;
  Colon?: Pos;
  Stmt?: Stmt;
  label: Ident;
  stmt: Stmt;
}

export interface ExprStmt extends Node {
  kind: "ExprStmt";
  X?: Expr;
  expr: Expr;
}

export interface AssignStmt extends Node {
  kind: "AssignStmt";
  Lhs?: Expr[];
  TokPos?: Pos;
  Tok?: GoToken;
  Rhs?: Expr[];
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
  X?: Expr;
  TokPos?: Pos;
  Tok?: GoToken;
  expr: Expr;
  token: TokenKind.PlusPlus | TokenKind.MinusMinus;
}

export interface ReturnStmt extends Node {
  kind: "ReturnStmt";
  Return?: Pos;
  Results?: Expr[];
  results: Expr[];
}

export interface BranchStmt extends Node {
  kind: "BranchStmt";
  TokPos?: Pos;
  Tok?: GoToken;
  Label?: Ident;
  token: TokenKind.Break | TokenKind.Continue | TokenKind.Goto | TokenKind.Fallthrough;
  label?: Ident;
}

export interface BlockStmt extends Node {
  kind: "BlockStmt";
  Lbrace?: Pos;
  List?: Stmt[];
  Rbrace?: Pos;
  statements: Stmt[];
}

export interface IfStmt extends Node {
  kind: "IfStmt";
  If?: Pos;
  Init?: Stmt;
  Cond?: Expr;
  Body?: BlockStmt;
  Else?: Stmt;
  init?: Stmt;
  condition: Expr;
  body: BlockStmt;
  else?: Stmt;
}

export interface CaseClause extends Node {
  kind: "CaseClause";
  Case?: Pos;
  List?: Expr[];
  Colon?: Pos;
  Body?: Stmt[];
  list: Expr[];
  body: Stmt[];
  default: boolean;
}

export interface CommClause extends Node {
  kind: "CommClause";
  Case?: Pos;
  Comm?: Stmt;
  Colon?: Pos;
  Body?: Stmt[];
  comm?: Stmt;
  body: Stmt[];
  default: boolean;
}

export interface SwitchStmt extends Node {
  kind: "SwitchStmt";
  Switch?: Pos;
  Init?: Stmt;
  Tag?: Expr;
  Body?: BlockStmt;
  init?: Stmt;
  tag?: Expr;
  body: CaseClause[];
}

export interface TypeSwitchStmt extends Node {
  kind: "TypeSwitchStmt";
  Switch?: Pos;
  Init?: Stmt;
  Assign?: Stmt;
  Body?: BlockStmt;
  init?: Stmt;
  assign: Stmt;
  body: CaseClause[];
}

export interface SelectStmt extends Node {
  kind: "SelectStmt";
  Select?: Pos;
  Body?: BlockStmt;
  body: CommClause[];
}

export interface ForStmt extends Node {
  kind: "ForStmt";
  For?: Pos;
  Init?: Stmt;
  Cond?: Expr;
  Post?: Stmt;
  Body?: BlockStmt;
  init?: Stmt;
  condition?: Expr;
  post?: Stmt;
  body: BlockStmt;
}

export interface RangeStmt extends Node {
  kind: "RangeStmt";
  For?: Pos;
  Key?: Expr;
  Value?: Expr;
  TokPos?: Pos;
  Tok?: GoToken;
  Range?: Pos;
  X?: Expr;
  Body?: BlockStmt;
  key?: Expr;
  value?: Expr;
  token: TokenKind.Assign | TokenKind.Define;
  source: Expr;
  body: BlockStmt;
}

export interface DeferStmt extends Node {
  kind: "DeferStmt";
  Defer?: Pos;
  Call?: CallExpr;
  call: CallExpr;
}

export interface GoStmt extends Node {
  kind: "GoStmt";
  Go?: Pos;
  Call?: CallExpr;
  call: CallExpr;
}

export interface SendStmt extends Node {
  kind: "SendStmt";
  Chan?: Expr;
  Arrow?: Pos;
  Value?: Expr;
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
  Doc?: CommentGroup;
  Names?: Ident[];
  Type?: Expr;
  Tag?: BasicLit;
  Comment?: CommentGroup;
  doc?: CommentGroup;
  names: Ident[];
  type: Expr;
  tag?: BasicLit;
  comment?: CommentGroup;
}

export interface FieldList extends Node {
  kind: "FieldList";
  Opening?: Pos;
  List?: Field[];
  Closing?: Pos;
  fields: Field[];
}

export interface ImportSpec extends Node {
  kind: "ImportSpec";
  Doc?: CommentGroup;
  Name?: Ident;
  Path?: BasicLit;
  EndPos?: Pos;
  Comment?: CommentGroup;
  doc?: CommentGroup;
  name?: Ident;
  path: BasicLit;
  comment?: CommentGroup;
}

export interface ValueSpec extends Node {
  kind: "ValueSpec";
  Doc?: CommentGroup;
  Names?: Ident[];
  Type?: Expr;
  Values?: Expr[];
  Comment?: CommentGroup;
  doc?: CommentGroup;
  names: Ident[];
  type?: Expr;
  values: Expr[];
  comment?: CommentGroup;
}

export interface TypeSpec extends Node {
  kind: "TypeSpec";
  Doc?: CommentGroup;
  Name?: Ident;
  TypeParams?: FieldList;
  Assign?: Pos;
  Type?: Expr;
  Comment?: CommentGroup;
  doc?: CommentGroup;
  name: Ident;
  typeParams?: FieldList;
  type: Expr;
  alias: boolean;
  comment?: CommentGroup;
}

export type Spec = ImportSpec | ValueSpec | TypeSpec;

export interface BadDecl extends Node {
  kind: "BadDecl";
  From?: Pos;
  To?: Pos;
}

export interface GenDecl extends Node {
  kind: "GenDecl";
  Doc?: CommentGroup;
  TokPos?: Pos;
  Tok?: GoToken;
  Lparen?: Pos;
  Specs?: Spec[];
  Rparen?: Pos;
  doc?: CommentGroup;
  token: TokenKind.Import | TokenKind.Const | TokenKind.Type | TokenKind.Var;
  specs: Spec[];
  grouped: boolean;
}

export interface FuncDecl extends Node {
  kind: "FuncDecl";
  Doc?: CommentGroup;
  Recv?: FieldList;
  Name?: Ident;
  Type?: FuncType;
  Body?: BlockStmt;
  doc?: CommentGroup;
  receiver?: FieldList;
  name: Ident;
  type: FuncType;
  body?: BlockStmt;
}

export type Decl = BadDecl | GenDecl | FuncDecl;

export interface File extends Node {
  kind: "File";
  Doc?: CommentGroup;
  Package?: Pos;
  Name?: Ident;
  Decls?: Decl[];
  FileStart?: Pos;
  FileEnd?: Pos;
  Scope?: Scope;
  Imports?: ImportSpec[];
  Unresolved?: Ident[];
  Comments?: CommentGroup[];
  GoVersion?: string;
  doc?: CommentGroup;
  name?: Ident;
  declarations: Decl[];
  imports: ImportSpec[];
  unresolved: Ident[];
  comments: CommentGroup[];
  scope?: Scope;
}

export interface Package extends Node {
  kind: "Package";
  Name?: string;
  Scope?: Scope;
  Imports?: Map<string, Object>;
  Files?: Map<string, File> | Record<string, File> | File[];
  name: string;
  files: Map<string, File> | Record<string, File> | File[];
  scope?: Scope;
  imports?: Map<string, Object>;
}

export interface Comment extends Node {
  kind: "Comment";
  Slash?: Pos;
  Text?: string;
  text: string;
}

export interface CommentGroup extends Node {
  kind: "CommentGroup";
  List?: Comment[];
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
    const list = group ? astList<Comment>(group, "List", "list") : [];
    return PosOf(list[0]);
  }

  export function End(group: CommentGroup | undefined): Pos {
    const list = group ? astList<Comment>(group, "List", "list") : [];
    return EndOf(list[list.length - 1]);
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
  const id = ident(name);
  NormalizeAst(id);
  return id;
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
    return nodePos(id) + identNameValue(id).length;
  }

  export function exprNode(_id: Ident | undefined): void {
  }

  export function IsExported(id: Ident | undefined): boolean {
    return id ? astIsExported(identNameValue(id)) : false;
  }

  export function String(id: Ident | undefined): string {
    if (id) {
      return identNameValue(id);
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
  export function End(x: Ellipsis): Pos {
    const element = astField<Expr>(x, "Elt", "element");
    return element ? EndOf(element) : nodeEnd(x) || nodePos(x) + 3;
  }
  export function exprNode(_x: Ellipsis): void {}
}

export namespace BasicLit {
  export function Pos(x: BasicLit): Pos { return nodePos(x); }
  export function End(x: BasicLit): Pos {
    return nodeEnd(x) || nodePos(x) + (x.Value ?? x.value).length;
  }
  export function exprNode(_x: BasicLit): void {}
}

export namespace FuncLit {
  export function Pos(x: FuncLit): Pos { return PosOf(astField<FuncType>(x, "Type", "type")); }
  export function End(x: FuncLit): Pos { return EndOf(astField<BlockStmt>(x, "Body", "body")); }
  export function exprNode(_x: FuncLit): void {}
}

export namespace CompositeLit {
  export function Pos(x: CompositeLit): Pos {
    const typ = astField<Expr>(x, "Type", "type");
    return typ ? PosOf(typ) : nodePos(x);
  }
  export function End(x: CompositeLit): Pos { return nodeEnd(x); }
  export function exprNode(_x: CompositeLit): void {}
}

export namespace ParenExpr {
  export function Pos(x: ParenExpr): Pos { return nodePos(x); }
  export function End(x: ParenExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: ParenExpr): void {}
}

export namespace SelectorExpr {
  export function Pos(x: SelectorExpr): Pos { return PosOf(astField<Expr>(x, "X", "object")); }
  export function End(x: SelectorExpr): Pos { return EndOf(astField<Ident>(x, "Sel", "selector")); }
  export function exprNode(_x: SelectorExpr): void {}
}

export namespace IndexExpr {
  export function Pos(x: IndexExpr): Pos { return PosOf(astField<Expr>(x, "X", "object")); }
  export function End(x: IndexExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: IndexExpr): void {}
}

export namespace IndexListExpr {
  export function Pos(x: IndexListExpr): Pos { return PosOf(astField<Expr>(x, "X", "object")); }
  export function End(x: IndexListExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: IndexListExpr): void {}
}

export namespace SliceExpr {
  export function Pos(x: SliceExpr): Pos { return PosOf(astField<Expr>(x, "X", "object")); }
  export function End(x: SliceExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: SliceExpr): void {}
}

export namespace TypeAssertExpr {
  export function Pos(x: TypeAssertExpr): Pos { return PosOf(astField<Expr>(x, "X", "object")); }
  export function End(x: TypeAssertExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: TypeAssertExpr): void {}
}

export namespace CallExpr {
  export function Pos(x: CallExpr): Pos { return PosOf(astField<Expr>(x, "Fun", "fun")); }
  export function End(x: CallExpr): Pos { return nodeEnd(x); }
  export function exprNode(_x: CallExpr): void {}
}

export namespace StarExpr {
  export function Pos(x: StarExpr): Pos { return nodePos(x); }
  export function End(x: StarExpr): Pos { return EndOf(astField<Expr>(x, "X", "expr")); }
  export function exprNode(_x: StarExpr): void {}
}

export namespace UnaryExpr {
  export function Pos(x: UnaryExpr): Pos { return nodePos(x); }
  export function End(x: UnaryExpr): Pos { return EndOf(astField<Expr>(x, "X", "expr")); }
  export function exprNode(_x: UnaryExpr): void {}
}

export namespace BinaryExpr {
  export function Pos(x: BinaryExpr): Pos { return PosOf(astField<Expr>(x, "X", "left")); }
  export function End(x: BinaryExpr): Pos { return EndOf(astField<Expr>(x, "Y", "right")); }
  export function exprNode(_x: BinaryExpr): void {}
}

export namespace KeyValueExpr {
  export function Pos(x: KeyValueExpr): Pos { return PosOf(astField<Expr>(x, "Key", "key")); }
  export function End(x: KeyValueExpr): Pos { return EndOf(astField<Expr>(x, "Value", "value")); }
  export function exprNode(_x: KeyValueExpr): void {}
}

export namespace ArrayType {
  export function Pos(x: ArrayType): Pos { return nodePos(x); }
  export function End(x: ArrayType): Pos { return EndOf(astField<Expr>(x, "Elt", "element")); }
  export function exprNode(_x: ArrayType): void {}
}

export namespace StructType {
  export function Pos(x: StructType): Pos { return nodePos(x); }
  export function End(x: StructType): Pos { return FieldList.End(astField<FieldList>(x, "Fields", "fields")); }
  export function exprNode(_x: StructType): void {}
}

export namespace FuncType {
  export function Pos(x: FuncType): Pos { return nodePos(x) || FieldList.Pos(astField<FieldList>(x, "Params", "params")); }
  export function End(x: FuncType): Pos {
    const results = astField<FieldList>(x, "Results", "results");
    return results ? FieldList.End(results) : FieldList.End(astField<FieldList>(x, "Params", "params"));
  }
  export function exprNode(_x: FuncType): void {}
}

export namespace InterfaceType {
  export function Pos(x: InterfaceType): Pos { return nodePos(x); }
  export function End(x: InterfaceType): Pos { return FieldList.End(astField<FieldList>(x, "Methods", "methods")); }
  export function exprNode(_x: InterfaceType): void {}
}

export namespace MapType {
  export function Pos(x: MapType): Pos { return nodePos(x); }
  export function End(x: MapType): Pos { return EndOf(astField<Expr>(x, "Value", "value")); }
  export function exprNode(_x: MapType): void {}
}

export namespace ChanType {
  export function Pos(x: ChanType): Pos { return nodePos(x); }
  export function End(x: ChanType): Pos { return EndOf(astField<Expr>(x, "Value", "value")); }
  export function exprNode(_x: ChanType): void {}
}

export namespace BadStmt {
  export function Pos(x: BadStmt): Pos { return nodePos(x); }
  export function End(x: BadStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: BadStmt): void {}
}

export namespace DeclStmt {
  export function Pos(x: DeclStmt): Pos { return PosOf(astField<Decl>(x, "Decl", "decl")); }
  export function End(x: DeclStmt): Pos { return EndOf(astField<Decl>(x, "Decl", "decl")); }
  export function stmtNode(_x: DeclStmt): void {}
}

export namespace EmptyStmt {
  export function Pos(x: EmptyStmt): Pos { return nodePos(x); }
  export function End(x: EmptyStmt): Pos { return x.implicit ? nodePos(x) : nodeEnd(x) || nodePos(x) + 1; }
  export function stmtNode(_x: EmptyStmt): void {}
}

export namespace LabeledStmt {
  export function Pos(x: LabeledStmt): Pos { return PosOf(astField<Ident>(x, "Label", "label")); }
  export function End(x: LabeledStmt): Pos { return EndOf(astField<Stmt>(x, "Stmt", "stmt")); }
  export function stmtNode(_x: LabeledStmt): void {}
}

export namespace ExprStmt {
  export function Pos(x: ExprStmt): Pos { return PosOf(astField<Expr>(x, "X", "expr")); }
  export function End(x: ExprStmt): Pos { return EndOf(astField<Expr>(x, "X", "expr")); }
  export function stmtNode(_x: ExprStmt): void {}
}

export namespace SendStmt {
  export function Pos(x: SendStmt): Pos { return PosOf(astField<Expr>(x, "Chan", "channel")); }
  export function End(x: SendStmt): Pos { return EndOf(astField<Expr>(x, "Value", "value")); }
  export function stmtNode(_x: SendStmt): void {}
}

export namespace IncDecStmt {
  export function Pos(x: IncDecStmt): Pos { return PosOf(astField<Expr>(x, "X", "expr")); }
  export function End(x: IncDecStmt): Pos { return nodeEnd(x) || EndOf(astField<Expr>(x, "X", "expr")) + 2; }
  export function stmtNode(_x: IncDecStmt): void {}
}

export namespace AssignStmt {
  export function Pos(x: AssignStmt): Pos { return PosOf(astList<Expr>(x, "Lhs", "lhs")[0]); }
  export function End(x: AssignStmt): Pos {
    const rhs = astList<Expr>(x, "Rhs", "rhs");
    return EndOf(rhs[rhs.length - 1]);
  }
  export function stmtNode(_x: AssignStmt): void {}
}

export namespace GoStmt {
  export function Pos(x: GoStmt): Pos { return nodePos(x); }
  export function End(x: GoStmt): Pos { return EndOf(astField<CallExpr>(x, "Call", "call")); }
  export function stmtNode(_x: GoStmt): void {}
}

export namespace DeferStmt {
  export function Pos(x: DeferStmt): Pos { return nodePos(x); }
  export function End(x: DeferStmt): Pos { return EndOf(astField<CallExpr>(x, "Call", "call")); }
  export function stmtNode(_x: DeferStmt): void {}
}

export namespace ReturnStmt {
  export function Pos(x: ReturnStmt): Pos { return nodePos(x); }
  export function End(x: ReturnStmt): Pos {
    const results = astList<Expr>(x, "Results", "results");
    return results.length > 0 ? EndOf(results[results.length - 1]) : nodeEnd(x) || nodePos(x) + 6;
  }
  export function stmtNode(_x: ReturnStmt): void {}
}

export namespace BranchStmt {
  export function Pos(x: BranchStmt): Pos { return nodePos(x); }
  export function End(x: BranchStmt): Pos {
    const label = astField<Ident>(x, "Label", "label");
    if (label) return EndOf(label);
    const tok = (x as unknown as { Tok?: GoToken }).Tok;
    if (typeof tok === "number") {
      return nodePos(x) + GoToken.String(tok).length;
    }
    return nodeEnd(x);
  }
  export function stmtNode(_x: BranchStmt): void {}
}

export namespace BlockStmt {
  export function Pos(x: BlockStmt): Pos { return nodePos(x); }
  export function End(x: BlockStmt): Pos {
    const list = astList<Stmt>(x, "List", "statements");
    return nodeEnd(x) || (list.length > 0 ? EndOf(list[list.length - 1]) : nodePos(x) + 1);
  }
  export function stmtNode(_x: BlockStmt): void {}
}

export namespace IfStmt {
  export function Pos(x: IfStmt): Pos { return nodePos(x); }
  export function End(x: IfStmt): Pos {
    const elseStmt = astField<Stmt>(x, "Else", "else");
    return elseStmt ? EndOf(elseStmt) : EndOf(astField<BlockStmt>(x, "Body", "body"));
  }
  export function stmtNode(_x: IfStmt): void {}
}

export namespace CaseClause {
  export function Pos(x: CaseClause): Pos { return nodePos(x); }
  export function End(x: CaseClause): Pos {
    const body = astList<Stmt>(x, "Body", "body");
    return body.length > 0 ? EndOf(body[body.length - 1]) : nodeEnd(x);
  }
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
  export function End(x: CommClause): Pos {
    const body = astList<Stmt>(x, "Body", "body");
    return body.length > 0 ? EndOf(body[body.length - 1]) : nodeEnd(x);
  }
  export function stmtNode(_x: CommClause): void {}
}

export namespace SelectStmt {
  export function Pos(x: SelectStmt): Pos { return nodePos(x); }
  export function End(x: SelectStmt): Pos { return nodeEnd(x); }
  export function stmtNode(_x: SelectStmt): void {}
}

export namespace ForStmt {
  export function Pos(x: ForStmt): Pos { return nodePos(x); }
  export function End(x: ForStmt): Pos { return EndOf(astField<BlockStmt>(x, "Body", "body")); }
  export function stmtNode(_x: ForStmt): void {}
}

export namespace RangeStmt {
  export function Pos(x: RangeStmt): Pos { return nodePos(x); }
  export function End(x: RangeStmt): Pos { return EndOf(astField<BlockStmt>(x, "Body", "body")); }
  export function stmtNode(_x: RangeStmt): void {}
}

export namespace Comment {
  export function Pos(x: Comment): Pos { return nodePos(x); }
  export function End(x: Comment): Pos { return nodeEnd(x) || nodePos(x) + x.text.length; }
}

export namespace Field {
  export function Pos(x: Field): Pos {
    const names = astList<Ident>(x, "Names", "names");
    return names.length > 0 ? PosOf(names[0]) : PosOf(astField<Expr>(x, "Type", "type"));
  }
  export function End(x: Field): Pos {
    const tag = astField<BasicLit>(x, "Tag", "tag");
    if (tag) return EndOf(tag);
    const typ = astField<Expr>(x, "Type", "type");
    if (typ) return EndOf(typ);
    const names = astList<Ident>(x, "Names", "names");
    return EndOf(names[names.length - 1]);
  }
}

export namespace FieldList {
  export function Pos(x: FieldList | undefined): Pos { return nodePos(x) || PosOf(x ? astList<Field>(x, "List", "fields")[0] : undefined); }
  export function End(x: FieldList | undefined): Pos {
    if (!x) return 0;
    const list = astList<Field>(x, "List", "fields");
    return nodeEnd(x) || EndOf(list[list.length - 1]);
  }
  export function NumFields(x: FieldList | undefined): number {
    if (!x) return 0;
    let n = 0;
    for (const field of astList<Field>(x, "List", "fields")) {
      n += astList<Ident>(field, "Names", "names").length || 1;
    }
    return n;
  }
}

export namespace ImportSpec {
  export function Pos(x: ImportSpec): Pos {
    const name = astField<Ident>(x, "Name", "name");
    return name ? PosOf(name) : PosOf(astField<BasicLit>(x, "Path", "path"));
  }
  export function End(x: ImportSpec): Pos {
    const endPos = numberField(x, "EndPos");
    if (endPos !== 0) return endPos;
    return EndOf(astField<BasicLit>(x, "Path", "path"));
  }
  export function specNode(_x: ImportSpec): void {}
}

export namespace ValueSpec {
  export function Pos(x: ValueSpec): Pos { return PosOf(astList<Ident>(x, "Names", "names")[0]); }
  export function End(x: ValueSpec): Pos {
    const values = astList<Expr>(x, "Values", "values");
    if (values.length > 0) return EndOf(values[values.length - 1]);
    const typ = astField<Expr>(x, "Type", "type");
    if (typ) return EndOf(typ);
    const names = astList<Ident>(x, "Names", "names");
    return EndOf(names[names.length - 1]);
  }
  export function specNode(_x: ValueSpec): void {}
}

export namespace TypeSpec {
  export function Pos(x: TypeSpec): Pos { return PosOf(astField<Ident>(x, "Name", "name")); }
  export function End(x: TypeSpec): Pos { return EndOf(astField<Expr>(x, "Type", "type")); }
  export function specNode(_x: TypeSpec): void {}
}

export namespace BadDecl {
  export function Pos(x: BadDecl): Pos { return nodePos(x); }
  export function End(x: BadDecl): Pos { return nodeEnd(x); }
  export function declNode(_x: BadDecl): void {}
}

export namespace GenDecl {
  export function Pos(x: GenDecl): Pos { return nodePos(x); }
  export function End(x: GenDecl): Pos {
    const rparen = numberField(x, "Rparen");
    if (rparen !== 0) return rparen + 1;
    const specs = astList<Spec>(x, "Specs", "specs");
    return EndOf(specs[0]);
  }
  export function declNode(_x: GenDecl): void {}
}

export namespace FuncDecl {
  export function Pos(x: FuncDecl): Pos { return PosOf(astField<FuncType>(x, "Type", "type")); }
  export function End(x: FuncDecl): Pos {
    const body = astField<BlockStmt>(x, "Body", "body");
    return body ? EndOf(body) : EndOf(astField<FuncType>(x, "Type", "type"));
  }
  export function declNode(_x: FuncDecl): void {}
}

export namespace File {
  export function Pos(x: File): Pos { return numberField(x, "Package"); }
  export function End(x: File): Pos {
    const decls = astList<Decl>(x, "Decls", "declarations");
    return decls.length > 0 ? EndOf(decls[decls.length - 1]) : EndOf(astField<Ident>(x, "Name", "name"));
  }
}

export namespace Package {
  export function Pos(_x: Package): Pos { return 0; }
  export function End(_x: Package): Pos { return 0; }
}

export function PosOf(node: AstNode | undefined): Pos {
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

export function EndOf(node: AstNode | undefined): Pos {
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
  const packageOffset = numberField(file, "Package") || PosOf(astField<Ident>(file, "Name", "name")) || Number.POSITIVE_INFINITY;
  for (const group of astList<CommentGroup>(file, "Comments", "comments")) {
    for (const comment of astList<Comment>(group, "List", "list")) {
      if (PosOf(comment) > packageOffset) {
        break;
      }
      const prefix = "// Code generated ";
      const text = comment.Text ?? comment.text;
      if (text.includes(prefix)) {
        for (const line of text.split("\n")) {
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
    expr = astField<Expr>(expr, "X", "expr")!;
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
    if (f(identNameValue(x))) {
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
      if (astField<Expr>(x, "X", "object")?.kind === "Ident") {
        return astField<Ident>(x, "Sel", "selector");
      }
      break;
    case "StarExpr":
      return fieldName(astField<Expr>(x, "X", "expr")!);
  }
  return undefined;
}

export function filterFieldList(fields: FieldList | undefined, filter: Filter, exportOnly: boolean): boolean {
  if (!fields) {
    return false;
  }
  const list = astList<Field>(fields, "List", "fields");
  let j = 0;
  let removedFields = false;
  for (const field of list) {
    let keepField = false;
    const names = astList<Ident>(field, "Names", "names");
    const typ = astField<Expr>(field, "Type", "type");
    if (names.length === 0) {
      const name = typ ? fieldName(typ) : undefined;
      keepField = name !== undefined && filter(identNameValue(name));
    } else {
      const n = names.length;
      field.names = filterIdentList(names, filter);
      if (field.Names !== undefined) {
        field.Names = field.names;
      }
      if (field.names.length < n) {
        removedFields = true;
      }
      keepField = field.names.length > 0;
    }
    if (keepField) {
      if (exportOnly) {
        filterType(typ, filter, exportOnly);
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
  const n = astList<Expr>(lit, "Elts", "elements").length;
  lit.elements = filterExprList(astList<Expr>(lit, "Elts", "elements"), filter, exportOnly);
  if (lit.Elts !== undefined) {
    lit.Elts = lit.elements;
  }
  if (lit.elements.length < n) {
    lit.Incomplete = true;
  }
}

export function filterExprList(list: Expr[], filter: Filter, exportOnly: boolean): Expr[] {
  let j = 0;
  for (const exp of list) {
    switch (exp.kind) {
      case "CompositeLit":
        filterCompositeLit(exp, filter, exportOnly);
        break;
      case "KeyValueExpr":
        {
          const key = astField<Expr>(exp, "Key", "key");
          const value = astField<Expr>(exp, "Value", "value");
          if (key?.kind === "Ident" && !filter(identNameValue(key))) {
            continue;
          }
          if (value?.kind === "CompositeLit") {
            filterCompositeLit(value, filter, exportOnly);
          }
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
  for (const field of astList<Field>(fields, "List", "fields")) {
    if (filterType(astField<Expr>(field, "Type", "type"), filter, exportOnly)) {
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
      return f(identNameValue(typ));
    case "ParenExpr":
      return filterType(astField<Expr>(typ, "X", "expr"), f, exportOnly);
    case "ArrayType":
      return filterType(astField<Expr>(typ, "Elt", "element"), f, exportOnly);
    case "StructType": {
      const fields = astField<FieldList>(typ, "Fields", "fields");
      if (filterFieldList(fields, f, exportOnly)) {
        typ.Incomplete = true;
      }
      return fields ? astList<Field>(fields, "List", "fields").length > 0 : false;
    }
    case "FuncType": {
      const b1 = filterParamList(astField<FieldList>(typ, "Params", "params"), f, exportOnly);
      const b2 = filterParamList(astField<FieldList>(typ, "Results", "results"), f, exportOnly);
      return b1 || b2;
    }
    case "InterfaceType": {
      const methods = astField<FieldList>(typ, "Methods", "methods");
      if (filterFieldList(methods, f, exportOnly)) {
        typ.Incomplete = true;
      }
      return methods ? astList<Field>(methods, "List", "fields").length > 0 : false;
    }
    case "MapType": {
      const b1 = filterType(astField<Expr>(typ, "Key", "key"), f, exportOnly);
      const b2 = filterType(astField<Expr>(typ, "Value", "value"), f, exportOnly);
      return b1 || b2;
    }
    case "ChanType":
      return filterType(astField<Expr>(typ, "Value", "value"), f, exportOnly);
  }
  return false;
}

export function filterSpec(spec: Spec, f: Filter, exportOnly: boolean): boolean {
  switch (spec.kind) {
    case "ValueSpec": {
      spec.names = filterIdentList(astList<Ident>(spec, "Names", "names"), f);
      if (spec.Names !== undefined) {
        spec.Names = spec.names;
      }
      spec.values = filterExprList(astList<Expr>(spec, "Values", "values"), f, exportOnly);
      if (spec.Values !== undefined) {
        spec.Values = spec.values;
      }
      if (spec.names.length > 0) {
        if (exportOnly) {
          filterType(astField<Expr>(spec, "Type", "type"), f, exportOnly);
        }
        return true;
      }
      break;
    }
    case "TypeSpec": {
      if (f(identNameValue(astField<Ident>(spec, "Name", "name")))) {
        if (exportOnly) {
          filterType(astField<Expr>(spec, "Type", "type"), f, exportOnly);
        }
        return true;
      }
      if (!exportOnly) {
        return filterType(astField<Expr>(spec, "Type", "type"), f, exportOnly);
      }
      break;
    }
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
      decl.specs = filterSpecList(astList<Spec>(decl, "Specs", "specs"), f, exportOnly);
      if (decl.Specs !== undefined) {
        decl.Specs = decl.specs;
      }
      return decl.specs.length > 0;
    case "FuncDecl":
      return f(identNameValue(astField<Ident>(decl, "Name", "name")));
  }
  return false;
}

export function FilterFile(src: File, f: Filter): boolean {
  return filterFile(src, f, false);
}

export function filterFile(src: File, f: Filter, exportOnly: boolean): boolean {
  let j = 0;
  const declarations = astList<Decl>(src, "Decls", "declarations");
  for (const declaration of declarations) {
    if (filterDecl(declaration, f, exportOnly)) {
      declarations[j] = declaration;
      j += 1;
    }
  }
  declarations.length = j;
  return j > 0;
}

export function FilterPackage(pkg: Package, f: Filter): boolean {
  return filterPackage(pkg, f, false);
}

export function filterPackage(pkg: Package, f: Filter, exportOnly: boolean): boolean {
  let hasDecls = false;
  for (const src of fileValues((pkg.Files ?? pkg.files) as Map<string, File> | Record<string, File> | File[])) {
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

export const separator: Comment = { kind: "Comment", Slash: 0, Text: "//", text: "//" };

export function nameOf(f: FuncDecl): string {
  const receiver = astField<FieldList>(f, "Recv", "receiver");
  const fields = receiver ? astList<Field>(receiver, "List", "fields") : [];
  if (fields.length === 1) {
    let typ = astField<Expr>(fields[0]!, "Type", "type");
    if (typ?.kind === "StarExpr") {
      typ = astField<Expr>(typ, "X", "expr");
    }
    if (typ?.kind === "Ident") {
      return `${identNameValue(typ)}.${identNameValue(astField<Ident>(f, "Name", "name"))}`;
    }
  }
  return identNameValue(astField<Ident>(f, "Name", "name"));
}

export function MergePackageFiles(pkg: Package, mode: MergeMode): File {
  const entries = fileEntries((pkg.Files ?? pkg.files) as Map<string, File> | Record<string, File> | File[])
    .sort(([left], [right]) => left.localeCompare(right));

  let ndocs = 0;
  let ncomments = 0;
  let ndecls = 0;
  let minPos = 0;
  let maxPos = 0;
  for (let i = 0; i < entries.length; i += 1) {
    const file = entries[i]![1];
    const doc = astField<CommentGroup>(file, "Doc", "doc");
    if (doc) {
      ndocs += astList<Comment>(doc, "List", "list").length + 1;
    }
    ncomments += astList<CommentGroup>(file, "Comments", "comments").length;
    ndecls += astList<Decl>(file, "Decls", "declarations").length;
    const fileStart = numberField(file, "FileStart") || PosOf(file);
    const fileEnd = numberField(file, "FileEnd") || EndOf(file);
    if (i === 0 || fileStart < minPos) {
      minPos = fileStart;
    }
    if (i === 0 || fileEnd > maxPos) {
      maxPos = fileEnd;
    }
  }

  let doc: CommentGroup | undefined;
  let pos = 0;
  if (ndocs > 0) {
    const list: Comment[] = [];
    for (const [, file] of entries) {
      const fileDoc = astField<CommentGroup>(file, "Doc", "doc");
      if (!fileDoc) {
        continue;
      }
      if (list.length > 0) {
        list.push(separator);
      }
      list.push(...astList<Comment>(fileDoc, "List", "list"));
      const packagePos = numberField(file, "Package");
      if (packagePos > pos) {
        pos = packagePos;
      }
    }
    doc = { kind: "CommentGroup", List: list, list };
  }

  let declarations: Decl[] = [];
  if (ndecls > 0) {
    const funcs = new Map<string, number>();
    const maybeDecls: Array<Decl | undefined> = [];
    for (const [, file] of entries) {
      for (let declaration of astList<Decl>(file, "Decls", "declarations")) {
        if ((mode & FilterFuncDuplicates) !== 0 && declaration.kind === "FuncDecl") {
          const name = nameOf(declaration);
          const index = funcs.get(name);
          if (index !== undefined) {
            const existing = maybeDecls[index];
            if (existing?.kind === "FuncDecl" && astField<CommentGroup>(existing, "Doc", "doc") === undefined) {
              maybeDecls[index] = undefined;
            } else {
              declaration = undefined as unknown as Decl;
            }
          } else {
            funcs.set(name, maybeDecls.length);
          }
        }
        maybeDecls.push(declaration);
      }
    }
    declarations = maybeDecls.filter((decl): decl is Decl => decl !== undefined);
  }

  let imports: ImportSpec[] = [];
  if ((mode & FilterImportDuplicates) !== 0) {
    const seen = new Set<string>();
    for (const [, file] of entries) {
      for (const imp of astList<ImportSpec>(file, "Imports", "imports")) {
        const path = astField<BasicLit>(imp, "Path", "path");
        const key = path?.Value ?? path?.value ?? "";
        if (!seen.has(key)) {
          imports.push(imp);
          seen.add(key);
        }
      }
    }
  } else {
    for (const [, file] of entries) {
      imports.push(...astList<ImportSpec>(file, "Imports", "imports"));
    }
  }

  let comments: CommentGroup[] = [];
  if ((mode & FilterUnassociatedComments) === 0) {
    comments = new Array<CommentGroup>(ncomments);
    let i = 0;
    for (const [, file] of entries) {
      for (const comment of astList<CommentGroup>(file, "Comments", "comments")) {
        comments[i] = comment;
        i += 1;
      }
    }
  }

  const name = NewIdent(pkg.Name ?? pkg.name);
  const scope = pkg.Scope ?? pkg.scope;
  const file: File = {
    kind: "File",
    ...(doc ? { Doc: doc, doc } : {}),
    Package: pos,
    Name: name,
    Decls: declarations,
    FileStart: minPos,
    FileEnd: maxPos,
    ...(scope ? { Scope: scope, scope } : {}),
    Imports: imports,
    Unresolved: [],
    Comments: comments,
    GoVersion: "",
    name,
    declarations,
    imports,
    unresolved: [],
    comments
  };
  return file;
}

interface commentPosition {
  Offset: number;
  Line: number;
}

interface positionFileSet {
  Position(pos: Pos): { Offset?: number; offset?: number; Line?: number; line?: number };
}

interface importFileSet {
  PositionFor?: (pos: Pos, adjusted: boolean) => { Line?: number; line?: number };
  Position?: (pos: Pos) => { Line?: number; line?: number };
  File?: (pos: Pos) => TokenFile | undefined;
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
      const s = node.kind === "Ident" ? identNameValue(node) : node.kind;
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
  for (const decl of astList<Decl>(f, "Decls", "declarations")) {
    if (decl.kind !== "GenDecl" || genDeclTokForAst(decl) !== GoToken.IMPORT) {
      break;
    }

    if (!genDeclGroupedForAst(decl)) {
      continue;
    }

    let i = 0;
    const specs: Spec[] = [];
    const declSpecs = astList<Spec>(decl, "Specs", "specs");
    for (let j = 0; j < declSpecs.length; j += 1) {
      const spec = declSpecs[j];
      const prev = declSpecs[j - 1];
      if (spec && prev && j > i && lineAt(fset, PosOf(spec)) > 1 + lineAt(fset, EndOf(prev))) {
        specs.push(...sortSpecs(fset, f, decl, declSpecs.slice(i, j)));
        i = j;
      }
    }
    specs.push(...sortSpecs(fset, f, decl, declSpecs.slice(i)));
    decl.specs = specs;
    if (decl.Specs !== undefined) {
      decl.Specs = specs;
    }

    if (specs.length > 0 && decl.Rparen !== undefined && decl.Rparen !== 0) {
      const lastSpec = specs[specs.length - 1];
      const lastLine = lineAt(fset, PosOf(lastSpec));
      let rparenLine = lineAt(fset, decl.Rparen);
      while (rparenLine > lastLine + 1) {
        rparenLine -= 1;
        mergeLine(fset, decl.Rparen, rparenLine);
      }
    }
  }

  const imports = astList<ImportSpec>(f, "Imports", "imports");
  imports.length = 0;
  for (const decl of astList<Decl>(f, "Decls", "declarations")) {
    if (decl.kind === "GenDecl" && genDeclTokForAst(decl) === GoToken.IMPORT) {
      for (const spec of astList<Spec>(decl, "Specs", "specs")) {
        if (spec.kind === "ImportSpec") {
          imports.push(spec);
        }
      }
    }
  }
}

export function lineAt(fset: unknown, pos: Pos | SourceSpan | AstNode | undefined): number {
  if (pos && typeof pos === "object" && "kind" in pos) {
    return lineAt(fset, PosOf(pos));
  }
  if (typeof pos === "number") {
    if (!Number.isFinite(pos) || pos === 0) return 0;
    const fileSet = fset as importFileSet | undefined;
    const positioned = fileSet?.PositionFor?.(pos, false) ?? fileSet?.Position?.(pos);
    return positioned?.Line ?? positioned?.line ?? 0;
  }
  return pos?.line ?? 0;
}

export function importPath(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  const path = astField<BasicLit>(s, "Path", "path");
  try {
    return unquoteGoString(path?.Value ?? path?.value ?? "");
  } catch {
    return "";
  }
}

export function importName(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  return identNameValue(astField<Ident>(s, "Name", "name"));
}

export function importComment(s: Spec): string {
  if (s.kind !== "ImportSpec") {
    return "";
  }
  return CommentGroup.Text(astField<CommentGroup>(s, "Comment", "comment"));
}

export function collapse(prev: Spec, next: Spec): boolean {
  if (importPath(next) !== importPath(prev) || importName(next) !== importName(prev)) {
    return false;
  }
  return prev.kind === "ImportSpec" && astField<CommentGroup>(prev, "Comment", "comment") === undefined;
}

export function sortSpecs(fset: unknown, f: File, d: GenDecl, specs: Spec[]): Spec[] {
  if (specs.length <= 1) {
    return specs;
  }

  const pos = new Array<posSpan>(specs.length);
  for (let i = 0; i < specs.length; i += 1) {
    const spec = specs[i];
    pos[i] = new posSpan(PosOf(spec), EndOf(spec));
  }

  const begSpecs = pos[0]?.Start ?? 0;
  const endSpecs = pos[pos.length - 1]?.End ?? 0;
  const beg = lineStart(fset, f, lineAt(fset, begSpecs));
  const endLine = lineAt(fset, endSpecs);
  let end: Pos;
  const endFile = fileForPos(fset, endSpecs);
  if (endLine === maxLine(fset, f, endFile)) {
    end = endSpecs;
  } else {
    end = lineStart(fset, f, endLine + 1, endFile);
  }
  const fileComments = astList<CommentGroup>(f, "Comments", "comments");
  let first = fileComments.length;
  let last = -1;
  for (let i = 0; i < fileComments.length; i += 1) {
    const group = fileComments[i];
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
    comments = fileComments.slice(first, last + 1);
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
      lineAt(fset, pos[specIndex]?.Start) + 1 === lineAt(fset, PosOf(group))
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
      const specStart = PosOf(spec);
      if (lineAt(fset, specStart) !== lineAt(fset, d.Rparen)) {
        mergeLine(fset, specStart, lineAt(fset, specStart));
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
    const name = astField<Ident>(spec, "Name", "name");
    const path = astField<BasicLit>(spec, "Path", "path");
    if (name?.span) {
      name.span = moveSpan(name.span, span.Start);
    }
    if (path) updateBasicLitPos(path, span.Start);
    spec.EndPos = span.End;
    for (const group of importComments.get(spec) ?? []) {
      for (const comment of astList<Comment>(group.cg, "List", "list")) {
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
  const length = EndOf(lit) - PosOf(lit);
  lit.ValuePos = pos;
  if (lit.ValueEnd !== undefined && lit.ValueEnd !== 0) {
    lit.ValueEnd = pos + length;
  }
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
      for (const n of astList<Ident>(decl, "Names", "names")) {
        if (identNameValue(n) === name) {
          return nodePos(n);
        }
      }
    } else if (isImportSpec(decl)) {
      const importNameNode = astField<Ident>(decl, "Name", "name");
      if (importNameNode && identNameValue(importNameNode) === name) {
        return nodePos(importNameNode);
      }
      return nodePos(astField<BasicLit>(decl, "Path", "path"));
    } else if (isValueSpec(decl)) {
      for (const n of astList<Ident>(decl, "Names", "names")) {
        if (identNameValue(n) === name) {
          return nodePos(n);
        }
      }
    } else if (isTypeSpec(decl)) {
      const typeNameNode = astField<Ident>(decl, "Name", "name");
      if (identNameValue(typeNameNode) === name) {
        return nodePos(typeNameNode);
      }
    } else if (isFuncDecl(decl)) {
      const funcNameNode = astField<Ident>(decl, "Name", "name");
      if (identNameValue(funcNameNode) === name) {
        return nodePos(funcNameNode);
      }
    } else if (isLabeledStmt(decl)) {
      const labelNode = astField<Ident>(decl, "Label", "label");
      if (identNameValue(labelNode) === name) {
        return nodePos(labelNode);
      }
    } else if (isAssignStmt(decl)) {
      for (const x of astList<Expr>(decl, "Lhs", "lhs")) {
        if (x.kind === "Ident" && identNameValue(x) === name) {
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
    const obj = scope.Lookup(identNameValue(id));
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
    const fileName = astField<Ident>(file, "Name", "name");
    const name = identNameValue(fileName);
    switch (true) {
      case pkgName === "":
        pkgName = name;
        break;
      case name !== pkgName:
        p.errorf(nodePos(fileName), "package %s; expected %s", name, pkgName);
        continue;
    }

    for (const obj of (file.Scope ?? file.scope)?.Objects.values() ?? []) {
      p.declare(pkgScope, undefined, obj);
    }
  }

  const imports = new Map<string, Object>();

  for (const file of fileValues(files)) {
    if (identNameValue(astField<Ident>(file, "Name", "name")) !== pkgName) {
      continue;
    }

    let importErrors = false;
    const fileScope = NewScope(pkgScope);
    for (const spec of astList<ImportSpec>(file, "Imports", "imports")) {
      if (importer === undefined) {
        importErrors = true;
        continue;
      }
      const path = importPath(spec);
      const [pkg, err] = importer(imports, path);
      if (err !== undefined || pkg === undefined) {
        p.errorf(nodePos(astField<BasicLit>(spec, "Path", "path")), "could not import %s (%s)", path, err?.message ?? "unknown error");
        importErrors = true;
        continue;
      }

      let name = pkg.Name;
      const importNameNode = astField<Ident>(spec, "Name", "name");
      if (importNameNode !== undefined) {
        name = identNameValue(importNameNode);
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
    const unresolved = astList<Ident>(file, "Unresolved", "unresolved");
    for (const id of unresolved) {
      if (!resolve(fileScope, id)) {
        p.errorf(nodePos(id), "undeclared name: %s", identNameValue(id));
        unresolved[i] = id;
        i += 1;
      }
    }
    unresolved.length = i;
    pkgScope.Outer = universe;
  }

  p.errors.sort((a, b) => a.message.localeCompare(b.message));
  return [{ kind: "Package", name: pkgName, scope: pkgScope, imports, files }, p.errors[0]];
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

function fileValues(files: Map<string, File> | Record<string, File> | File[]): File[] {
  if (Array.isArray(files)) {
    return files;
  }
  if (files instanceof Map) {
    return [...files.values()];
  }
  return globalThis.Object.values(files);
}

function fileEntries(files: Map<string, File> | Record<string, File> | File[]): Array<[string, File]> {
  if (Array.isArray(files)) {
    return files.map((file, index) => [String(index), file]);
  }
  if (files instanceof Map) {
    return [...files.entries()];
  }
  return globalThis.Object.entries(files);
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

function identNameValue(id: Ident | undefined): string {
  return id?.name ?? id?.Name ?? "";
}

function numberField(node: unknown, ...names: string[]): Pos {
  const object = node as Record<string, unknown> | null | undefined;
  if (!object) return 0;
  for (const name of names) {
    const value = object[name];
    if (typeof value === "number") return value;
  }
  return 0;
}

function genDeclTokForAst(decl: GenDecl): GoToken {
  if (typeof decl.Tok === "number") return decl.Tok;
  if (decl.token !== undefined) return goToken(decl.token);
  return GoToken.ILLEGAL;
}

function genDeclGroupedForAst(decl: GenDecl): boolean {
  return numberField(decl, "Lparen") !== 0 || decl.grouped === true;
}

function nodePos(node: Node | CommentGroup | undefined): Pos {
  return node?.span?.offset ??
    numberField(
      node,
      "NamePos",
      "From",
      "Ellipsis",
      "ValuePos",
      "Lbrace",
      "Lparen",
      "Lbrack",
      "Star",
      "OpPos",
      "Struct",
      "Func",
      "Interface",
      "Map",
      "Begin",
      "Semicolon",
      "TokPos",
      "If",
      "Case",
      "Switch",
      "Select",
      "For",
      "Defer",
      "Go",
      "Return",
      "Opening",
      "Package",
      "FileStart",
      "Slash"
    );
}

function nodeEnd(node: Node | undefined): Pos {
  if (node?.span) {
    return node.span.offset + node.span.length;
  }
  const direct = numberField(node, "To", "ValueEnd", "FileEnd", "EndPos");
  if (direct !== 0) {
    return direct;
  }
  const closing = numberField(node, "Rbrace", "Rparen", "Rbrack", "Closing");
  if (closing !== 0) {
    return closing + 1;
  }
  const object = node as Record<string, unknown> | undefined;
  if (object?.kind === "Ident") {
    return nodePos(node) + identNameValue(node as Ident).length;
  }
  if (object?.kind === "BasicLit" && typeof object.Value === "string") {
    return nodePos(node) + object.Value.length;
  }
  if (object?.kind === "Comment" && typeof object.Text === "string") {
    return nodePos(node) + object.Text.length;
  }
  return 0;
}

function moveSpan(span: SourceSpan, offset: Pos): SourceSpan {
  return {
    ...span,
    offset
  };
}

function fileForPos(fset: unknown, pos: Pos): TokenFile | undefined {
  return (fset as importFileSet | undefined)?.File?.(pos);
}

function lineStart(fset: unknown, file: File, line: number, tokenFile?: TokenFile): Pos {
  const start = tokenFile?.LineStart(line);
  if (start !== undefined) {
    return start;
  }
  const fileSetFile = fileForPos(fset, PosOf(file));
  const fileSetStart = fileSetFile?.LineStart(line);
  if (fileSetStart !== undefined) {
    return fileSetStart;
  }
  for (const node of childNodes(file)) {
    const span = node.span;
    if (span && span.line === line) {
      return span.offset - Math.max(0, span.column - 1);
    }
  }
  return line;
}

function maxLine(fset: unknown, file: File, tokenFile?: TokenFile): number {
  const fileLineCount = tokenFile?.LineCount();
  if (fileLineCount !== undefined) {
    return fileLineCount;
  }
  const fileSetFile = fileForPos(fset, PosOf(file));
  const fileSetLineCount = fileSetFile?.LineCount();
  if (fileSetLineCount !== undefined) {
    return fileSetLineCount;
  }
  let line = 0;
  walk(file, (node) => {
    if (node.span) {
      line = Math.max(line, node.span.line);
    }
  });
  return line;
}

function mergeLine(fset: unknown, pos: Pos, line: number): void {
  fileForPos(fset, pos)?.MergeLine(line);
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

function defineGoField<T extends AstNode>(node: T, name: string, value: unknown): void {
  globalThis.Object.defineProperty(node, name, {
    configurable: true,
    enumerable: false,
    writable: true,
    value
  });
}

function spanPos(node: Node | undefined): Pos {
  return node?.span?.offset ?? 0;
}

function spanEnd(node: Node | undefined): Pos {
  return node?.span ? node.span.offset + node.span.length : spanPos(node);
}

function closingSpanPos(node: Node | undefined): Pos {
  const end = spanEnd(node);
  const start = spanPos(node);
  return end > start ? end - 1 : end;
}

function normalizeAstNode(node: AstNode | undefined): void {
  if (node) NormalizeAst(node);
}

function normalizeAstNodes(nodes: readonly AstNode[] | undefined): void {
  if (!nodes) return;
  for (const node of nodes) NormalizeAst(node);
}

function goToken(kind: TokenKind): GoToken {
  switch (kind) {
    case TokenKind.Illegal:
      return GoToken.ILLEGAL;
    case TokenKind.EOF:
      return GoToken.EOF;
    case TokenKind.Identifier:
    case TokenKind.CellAddress:
      return GoToken.IDENT;
    case TokenKind.IntLiteral:
      return GoToken.INT;
    case TokenKind.FloatLiteral:
      return GoToken.FLOAT;
    case TokenKind.ImagLiteral:
      return GoToken.IMAG;
    case TokenKind.RuneLiteral:
      return GoToken.CHAR;
    case TokenKind.StringLiteral:
      return GoToken.STRING;
    case TokenKind.Import:
      return GoToken.IMPORT;
    case TokenKind.Package:
      return GoToken.PACKAGE;
    case TokenKind.Func:
      return GoToken.FUNC;
    case TokenKind.Return:
      return GoToken.RETURN;
    case TokenKind.If:
      return GoToken.IF;
    case TokenKind.Else:
      return GoToken.ELSE;
    case TokenKind.Switch:
      return GoToken.SWITCH;
    case TokenKind.Case:
      return GoToken.CASE;
    case TokenKind.Default:
      return GoToken.DEFAULT;
    case TokenKind.Fallthrough:
      return GoToken.FALLTHROUGH;
    case TokenKind.Goto:
      return GoToken.GOTO;
    case TokenKind.For:
      return GoToken.FOR;
    case TokenKind.Range:
      return GoToken.RANGE;
    case TokenKind.Break:
      return GoToken.BREAK;
    case TokenKind.Continue:
      return GoToken.CONTINUE;
    case TokenKind.Defer:
      return GoToken.DEFER;
    case TokenKind.Var:
      return GoToken.VAR;
    case TokenKind.Const:
      return GoToken.CONST;
    case TokenKind.Type:
      return GoToken.TYPE;
    case TokenKind.Struct:
      return GoToken.STRUCT;
    case TokenKind.Interface:
      return GoToken.INTERFACE;
    case TokenKind.Map:
      return GoToken.MAP;
    case TokenKind.Chan:
      return GoToken.CHAN;
    case TokenKind.Go:
      return GoToken.GO;
    case TokenKind.Select:
      return GoToken.SELECT;
    case TokenKind.Ellipsis:
      return GoToken.ELLIPSIS;
    case TokenKind.Define:
      return GoToken.DEFINE;
    case TokenKind.Assign:
      return GoToken.ASSIGN;
    case TokenKind.PlusAssign:
      return GoToken.ADD_ASSIGN;
    case TokenKind.MinusAssign:
      return GoToken.SUB_ASSIGN;
    case TokenKind.StarAssign:
      return GoToken.MUL_ASSIGN;
    case TokenKind.SlashAssign:
      return GoToken.QUO_ASSIGN;
    case TokenKind.PercentAssign:
      return GoToken.REM_ASSIGN;
    case TokenKind.AmpAssign:
      return GoToken.AND_ASSIGN;
    case TokenKind.OrAssign:
      return GoToken.OR_ASSIGN;
    case TokenKind.CaretAssign:
      return GoToken.XOR_ASSIGN;
    case TokenKind.ShlAssign:
      return GoToken.SHL_ASSIGN;
    case TokenKind.ShrAssign:
      return GoToken.SHR_ASSIGN;
    case TokenKind.BitClearAssign:
      return GoToken.AND_NOT_ASSIGN;
    case TokenKind.Equal:
      return GoToken.EQL;
    case TokenKind.NotEqual:
      return GoToken.NEQ;
    case TokenKind.Less:
      return GoToken.LSS;
    case TokenKind.LessEqual:
      return GoToken.LEQ;
    case TokenKind.Greater:
      return GoToken.GTR;
    case TokenKind.GreaterEqual:
      return GoToken.GEQ;
    case TokenKind.AndAnd:
      return GoToken.LAND;
    case TokenKind.OrOr:
      return GoToken.LOR;
    case TokenKind.Arrow:
      return GoToken.ARROW;
    case TokenKind.Plus:
      return GoToken.ADD;
    case TokenKind.PlusPlus:
      return GoToken.INC;
    case TokenKind.Minus:
      return GoToken.SUB;
    case TokenKind.MinusMinus:
      return GoToken.DEC;
    case TokenKind.Star:
      return GoToken.MUL;
    case TokenKind.Slash:
      return GoToken.QUO;
    case TokenKind.Percent:
      return GoToken.REM;
    case TokenKind.Or:
      return GoToken.OR;
    case TokenKind.Caret:
      return GoToken.XOR;
    case TokenKind.Tilde:
      return GoToken.TILDE;
    case TokenKind.Shl:
      return GoToken.SHL;
    case TokenKind.Shr:
      return GoToken.SHR;
    case TokenKind.BitClear:
      return GoToken.AND_NOT;
    case TokenKind.Bang:
      return GoToken.NOT;
    case TokenKind.Amp:
      return GoToken.AND;
    case TokenKind.Dot:
      return GoToken.PERIOD;
    case TokenKind.Comma:
      return GoToken.COMMA;
    case TokenKind.Colon:
      return GoToken.COLON;
    case TokenKind.Semicolon:
      return GoToken.SEMICOLON;
    case TokenKind.LParen:
      return GoToken.LPAREN;
    case TokenKind.RParen:
      return GoToken.RPAREN;
    case TokenKind.LBrace:
      return GoToken.LBRACE;
    case TokenKind.RBrace:
      return GoToken.RBRACE;
    case TokenKind.LBracket:
      return GoToken.LBRACK;
    case TokenKind.RBracket:
      return GoToken.RBRACK;
    case TokenKind.True:
    case TokenKind.False:
    case TokenKind.Nil:
      return GoToken.IDENT;
  }
}

function chanDir(direction: "send" | "receive" | "both"): ChanDir {
  switch (direction) {
    case "send":
      return SEND;
    case "receive":
      return RECV;
    case "both":
      return SEND | RECV;
  }
}

// NormalizeAst mirrors the standard go/ast exported struct field names onto
// the existing GoJr node objects. The aliases are non-enumerable so code that
// already serializes or compares the lower-case compatibility fields keeps the
// same behavior while callers can use the standard Go field names.
export function NormalizeAst(node: AstNode | undefined): void {
  if (!node) return;

  defineGoField(node, "Pos", () => PosOf(node));
  defineGoField(node, "End", () => EndOf(node));

  switch (node.kind) {
    case "Ident":
      defineGoField(node, "NamePos", spanPos(node));
      defineGoField(node, "Name", node.name);
      defineGoField(node, "Obj", node.Obj);
      break;
    case "BadExpr":
      defineGoField(node, "From", spanPos(node));
      defineGoField(node, "To", spanEnd(node));
      break;
    case "Ellipsis":
      normalizeAstNode(node.element);
      defineGoField(node, "Ellipsis", spanPos(node));
      defineGoField(node, "Elt", node.element);
      break;
    case "BasicLit":
      defineGoField(node, "ValuePos", spanPos(node));
      defineGoField(node, "ValueEnd", spanEnd(node));
      defineGoField(node, "Kind", goToken(node.token));
      defineGoField(node, "Value", node.value);
      break;
    case "FuncLit":
      normalizeAstNode(node.type);
      normalizeAstNode(node.body);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Body", node.body);
      break;
    case "CompositeLit":
      normalizeAstNode(node.type);
      normalizeAstNodes(node.elements);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Lbrace", spanPos(node));
      defineGoField(node, "Elts", node.elements);
      defineGoField(node, "Rbrace", closingSpanPos(node));
      defineGoField(node, "Incomplete", false);
      break;
    case "ParenExpr":
      normalizeAstNode(node.expr);
      defineGoField(node, "Lparen", spanPos(node));
      defineGoField(node, "X", node.expr);
      defineGoField(node, "Rparen", closingSpanPos(node));
      break;
    case "SelectorExpr":
      normalizeAstNode(node.object);
      normalizeAstNode(node.selector);
      defineGoField(node, "X", node.object);
      defineGoField(node, "Sel", node.selector);
      break;
    case "IndexExpr":
      normalizeAstNode(node.object);
      normalizeAstNode(node.index);
      defineGoField(node, "X", node.object);
      defineGoField(node, "Lbrack", spanPos(node.index));
      defineGoField(node, "Index", node.index);
      defineGoField(node, "Rbrack", closingSpanPos(node));
      break;
    case "IndexListExpr":
      normalizeAstNode(node.object);
      normalizeAstNodes(node.indices);
      defineGoField(node, "X", node.object);
      defineGoField(node, "Lbrack", node.indices[0] ? spanPos(node.indices[0]) : spanPos(node));
      defineGoField(node, "Indices", node.indices);
      defineGoField(node, "Rbrack", closingSpanPos(node));
      break;
    case "SliceExpr":
      normalizeAstNode(node.object);
      normalizeAstNode(node.low);
      normalizeAstNode(node.high);
      normalizeAstNode(node.max);
      defineGoField(node, "X", node.object);
      defineGoField(node, "Lbrack", spanPos(node));
      defineGoField(node, "Low", node.low);
      defineGoField(node, "High", node.high);
      defineGoField(node, "Max", node.max);
      defineGoField(node, "Slice3", node.max !== undefined);
      defineGoField(node, "Rbrack", closingSpanPos(node));
      break;
    case "TypeAssertExpr":
      normalizeAstNode(node.object);
      normalizeAstNode(node.type);
      defineGoField(node, "X", node.object);
      defineGoField(node, "Lparen", spanPos(node));
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Rparen", closingSpanPos(node));
      break;
    case "CallExpr":
      normalizeAstNode(node.fun);
      normalizeAstNodes(node.args);
      defineGoField(node, "Fun", node.fun);
      defineGoField(node, "Lparen", spanPos(node));
      defineGoField(node, "Args", node.args);
      defineGoField(node, "Ellipsis", node.ellipsis ? spanEnd(node) : 0);
      defineGoField(node, "Rparen", closingSpanPos(node));
      break;
    case "StarExpr":
      normalizeAstNode(node.expr);
      defineGoField(node, "Star", spanPos(node));
      defineGoField(node, "X", node.expr);
      break;
    case "UnaryExpr":
      normalizeAstNode(node.expr);
      defineGoField(node, "OpPos", spanPos(node));
      defineGoField(node, "Op", goToken(node.op));
      defineGoField(node, "X", node.expr);
      break;
    case "BinaryExpr":
      normalizeAstNode(node.left);
      normalizeAstNode(node.right);
      defineGoField(node, "X", node.left);
      defineGoField(node, "OpPos", spanPos(node));
      defineGoField(node, "Op", goToken(node.op));
      defineGoField(node, "Y", node.right);
      break;
    case "KeyValueExpr":
      normalizeAstNode(node.key);
      normalizeAstNode(node.value);
      defineGoField(node, "Key", node.key);
      defineGoField(node, "Colon", spanPos(node));
      defineGoField(node, "Value", node.value);
      break;
    case "CellRefExpr":
      normalizeAstNode(node.namespace);
      break;
    case "RangeRefExpr":
      normalizeAstNode(node.namespace);
      break;
    case "ArrayType":
      normalizeAstNode(node.length);
      normalizeAstNode(node.element);
      defineGoField(node, "Lbrack", spanPos(node));
      defineGoField(node, "Len", node.length);
      defineGoField(node, "Elt", node.element);
      break;
    case "StructType":
      normalizeAstNode(node.fields);
      defineGoField(node, "Struct", spanPos(node));
      defineGoField(node, "Fields", node.fields);
      defineGoField(node, "Incomplete", false);
      break;
    case "FuncType":
      normalizeAstNode(node.typeParams);
      normalizeAstNode(node.params);
      normalizeAstNode(node.results);
      defineGoField(node, "Func", spanPos(node));
      defineGoField(node, "TypeParams", node.typeParams);
      defineGoField(node, "Params", node.params);
      defineGoField(node, "Results", node.results);
      break;
    case "InterfaceType":
      normalizeAstNode(node.methods);
      defineGoField(node, "Interface", spanPos(node));
      defineGoField(node, "Methods", node.methods);
      defineGoField(node, "Incomplete", false);
      break;
    case "MapType":
      normalizeAstNode(node.key);
      normalizeAstNode(node.value);
      defineGoField(node, "Map", spanPos(node));
      defineGoField(node, "Key", node.key);
      defineGoField(node, "Value", node.value);
      break;
    case "ChanType":
      normalizeAstNode(node.value);
      defineGoField(node, "Begin", spanPos(node));
      defineGoField(node, "Arrow", spanPos(node));
      defineGoField(node, "Dir", chanDir(node.direction));
      defineGoField(node, "Value", node.value);
      break;
    case "BadStmt":
      defineGoField(node, "From", spanPos(node));
      defineGoField(node, "To", spanEnd(node));
      break;
    case "DeclStmt":
      normalizeAstNode(node.decl);
      defineGoField(node, "Decl", node.decl);
      break;
    case "EmptyStmt":
      defineGoField(node, "Semicolon", spanPos(node));
      defineGoField(node, "Implicit", node.implicit);
      break;
    case "LabeledStmt":
      normalizeAstNode(node.label);
      normalizeAstNode(node.stmt);
      defineGoField(node, "Label", node.label);
      defineGoField(node, "Colon", spanPos(node));
      defineGoField(node, "Stmt", node.stmt);
      break;
    case "ExprStmt":
      normalizeAstNode(node.expr);
      defineGoField(node, "X", node.expr);
      break;
    case "SendStmt":
      normalizeAstNode(node.channel);
      normalizeAstNode(node.value);
      defineGoField(node, "Chan", node.channel);
      defineGoField(node, "Arrow", spanPos(node));
      defineGoField(node, "Value", node.value);
      break;
    case "IncDecStmt":
      normalizeAstNode(node.expr);
      defineGoField(node, "X", node.expr);
      defineGoField(node, "TokPos", spanPos(node));
      defineGoField(node, "Tok", goToken(node.token));
      break;
    case "AssignStmt":
      normalizeAstNodes(node.lhs);
      normalizeAstNodes(node.rhs);
      defineGoField(node, "Lhs", node.lhs);
      defineGoField(node, "TokPos", spanPos(node));
      defineGoField(node, "Tok", goToken(node.token));
      defineGoField(node, "Rhs", node.rhs);
      break;
    case "GoStmt":
      normalizeAstNode(node.call);
      defineGoField(node, "Go", spanPos(node));
      defineGoField(node, "Call", node.call);
      break;
    case "DeferStmt":
      normalizeAstNode(node.call);
      defineGoField(node, "Defer", spanPos(node));
      defineGoField(node, "Call", node.call);
      break;
    case "ReturnStmt":
      normalizeAstNodes(node.results);
      defineGoField(node, "Return", spanPos(node));
      defineGoField(node, "Results", node.results);
      break;
    case "BranchStmt":
      normalizeAstNode(node.label);
      defineGoField(node, "TokPos", spanPos(node));
      defineGoField(node, "Tok", goToken(node.token));
      defineGoField(node, "Label", node.label);
      break;
    case "BlockStmt":
      normalizeAstNodes(node.statements);
      defineGoField(node, "Lbrace", spanPos(node));
      defineGoField(node, "List", node.statements);
      defineGoField(node, "Rbrace", closingSpanPos(node));
      break;
    case "IfStmt":
      normalizeAstNode(node.init);
      normalizeAstNode(node.condition);
      normalizeAstNode(node.body);
      normalizeAstNode(node.else);
      defineGoField(node, "If", spanPos(node));
      defineGoField(node, "Init", node.init);
      defineGoField(node, "Cond", node.condition);
      defineGoField(node, "Body", node.body);
      defineGoField(node, "Else", node.else);
      break;
    case "CaseClause":
      normalizeAstNodes(node.list);
      normalizeAstNodes(node.body);
      defineGoField(node, "Case", spanPos(node));
      defineGoField(node, "List", node.default ? undefined : node.list);
      defineGoField(node, "Colon", spanPos(node));
      defineGoField(node, "Body", node.body);
      break;
    case "CommClause":
      normalizeAstNode(node.comm);
      normalizeAstNodes(node.body);
      defineGoField(node, "Case", spanPos(node));
      defineGoField(node, "Comm", node.default ? undefined : node.comm);
      defineGoField(node, "Colon", spanPos(node));
      defineGoField(node, "Body", node.body);
      break;
    case "SwitchStmt":
      normalizeAstNode(node.init);
      normalizeAstNode(node.tag);
      normalizeAstNodes(node.body);
      defineGoField(node, "Switch", spanPos(node));
      defineGoField(node, "Init", node.init);
      defineGoField(node, "Tag", node.tag);
      defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
      normalizeAstNode((node as unknown as { Body?: AstNode }).Body);
      break;
    case "TypeSwitchStmt":
      normalizeAstNode(node.init);
      normalizeAstNode(node.assign);
      normalizeAstNodes(node.body);
      defineGoField(node, "Switch", spanPos(node));
      defineGoField(node, "Init", node.init);
      defineGoField(node, "Assign", node.assign);
      defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
      normalizeAstNode((node as unknown as { Body?: AstNode }).Body);
      break;
    case "SelectStmt":
      normalizeAstNodes(node.body);
      defineGoField(node, "Select", spanPos(node));
      defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
      normalizeAstNode((node as unknown as { Body?: AstNode }).Body);
      break;
    case "ForStmt":
      normalizeAstNode(node.init);
      normalizeAstNode(node.condition);
      normalizeAstNode(node.post);
      normalizeAstNode(node.body);
      defineGoField(node, "For", spanPos(node));
      defineGoField(node, "Init", node.init);
      defineGoField(node, "Cond", node.condition);
      defineGoField(node, "Post", node.post);
      defineGoField(node, "Body", node.body);
      break;
    case "RangeStmt":
      normalizeAstNode(node.key);
      normalizeAstNode(node.value);
      normalizeAstNode(node.source);
      normalizeAstNode(node.body);
      defineGoField(node, "For", spanPos(node));
      defineGoField(node, "Key", node.key);
      defineGoField(node, "Value", node.value);
      defineGoField(node, "TokPos", spanPos(node));
      defineGoField(node, "Tok", goToken(node.token));
      defineGoField(node, "Range", spanPos(node.source));
      defineGoField(node, "X", node.source);
      defineGoField(node, "Body", node.body);
      break;
    case "UnsupportedStmt":
      break;
    case "Field":
      normalizeAstNode(node.doc);
      normalizeAstNodes(node.names);
      normalizeAstNode(node.type);
      normalizeAstNode(node.tag);
      normalizeAstNode(node.comment);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Names", node.names);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Tag", node.tag);
      defineGoField(node, "Comment", node.comment);
      break;
    case "FieldList":
      normalizeAstNodes(node.fields);
      defineGoField(node, "Opening", spanPos(node));
      defineGoField(node, "List", node.fields);
      defineGoField(node, "Closing", closingSpanPos(node));
      break;
    case "ImportSpec":
      normalizeAstNode(node.doc);
      normalizeAstNode(node.name);
      normalizeAstNode(node.path);
      normalizeAstNode(node.comment);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Name", node.name);
      defineGoField(node, "Path", node.path);
      defineGoField(node, "EndPos", spanEnd(node));
      defineGoField(node, "Comment", node.comment);
      break;
    case "ValueSpec":
      normalizeAstNode(node.doc);
      normalizeAstNodes(node.names);
      normalizeAstNode(node.type);
      normalizeAstNodes(node.values);
      normalizeAstNode(node.comment);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Names", node.names);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Values", node.values);
      defineGoField(node, "Comment", node.comment);
      break;
    case "TypeSpec":
      normalizeAstNode(node.doc);
      normalizeAstNode(node.name);
      normalizeAstNode(node.typeParams);
      normalizeAstNode(node.type);
      normalizeAstNode(node.comment);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Name", node.name);
      defineGoField(node, "TypeParams", node.typeParams);
      defineGoField(node, "Assign", node.alias ? spanPos(node) : 0);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Comment", node.comment);
      break;
    case "BadDecl":
      defineGoField(node, "From", spanPos(node));
      defineGoField(node, "To", spanEnd(node));
      break;
    case "GenDecl":
      normalizeAstNode(node.doc);
      normalizeAstNodes(node.specs);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "TokPos", spanPos(node));
      defineGoField(node, "Tok", goToken(node.token));
      defineGoField(node, "Lparen", node.grouped ? spanPos(node) : 0);
      defineGoField(node, "Specs", node.specs);
      defineGoField(node, "Rparen", node.grouped ? closingSpanPos(node) : 0);
      break;
    case "FuncDecl":
      normalizeAstNode(node.doc);
      normalizeAstNode(node.receiver);
      normalizeAstNode(node.name);
      normalizeAstNode(node.type);
      normalizeAstNode(node.body);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Recv", node.receiver);
      defineGoField(node, "Name", node.name);
      defineGoField(node, "Type", node.type);
      defineGoField(node, "Body", node.body);
      break;
    case "File":
      normalizeAstNode(node.doc);
      normalizeAstNode(node.name);
      normalizeAstNodes(node.declarations);
      normalizeAstNodes(node.imports);
      normalizeAstNodes(node.unresolved);
      normalizeAstNodes(node.comments);
      defineGoField(node, "Doc", node.doc);
      defineGoField(node, "Package", spanPos(node.name));
      defineGoField(node, "Name", node.name);
      defineGoField(node, "Decls", node.declarations);
      defineGoField(node, "FileStart", spanPos(node));
      defineGoField(node, "FileEnd", spanEnd(node));
      defineGoField(node, "Scope", node.scope);
      defineGoField(node, "Imports", node.imports);
      defineGoField(node, "Unresolved", node.unresolved);
      defineGoField(node, "Comments", node.comments);
      defineGoField(node, "GoVersion", "");
      break;
    case "Package":
      normalizeAstNodes(fileValues(node.files));
      defineGoField(node, "Name", node.name);
      defineGoField(node, "Scope", node.scope);
      defineGoField(node, "Imports", node.imports);
      defineGoField(node, "Files", node.files);
      break;
    case "Comment":
      defineGoField(node, "Slash", spanPos(node));
      defineGoField(node, "Text", node.text);
      break;
    case "CommentGroup":
      normalizeAstNodes(node.list);
      defineGoField(node, "List", node.list);
      break;
  }
}

function astField<T extends AstNode>(node: AstNode, standard: string, compat?: string): T | undefined {
  const object = node as unknown as Record<string, unknown>;
  return (object[standard] ?? (compat ? object[compat] : undefined)) as T | undefined;
}

function astList<T extends AstNode>(node: AstNode, standard: string, compat?: string): T[] {
  const object = node as unknown as Record<string, unknown>;
  const value = object[standard] ?? (compat ? object[compat] : undefined);
  return Array.isArray(value) ? value as T[] : [];
}

function astMaybe<T extends AstNode>(value: T | undefined): T[] {
  return value ? [value] : [];
}

export function childNodes(node: AstNode): AstNode[] {
  switch (node.kind) {
    case "File":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astMaybe(astField<Ident>(node, "Name", "name")),
        ...astList<Decl>(node, "Decls", "declarations")
      ];
    case "Package":
      return fileValues((node.Files ?? node.files) as Map<string, File> | Record<string, File> | File[]);
    case "GenDecl":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astList<Spec>(node, "Specs", "specs")
      ];
    case "FuncDecl":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astMaybe(astField<FieldList>(node, "Recv", "receiver")),
        ...astMaybe(astField<Ident>(node, "Name", "name")),
        ...astMaybe(astField<FuncType>(node, "Type", "type")),
        ...astMaybe(astField<BlockStmt>(node, "Body", "body"))
      ];
    case "FieldList":
      return astList<Field>(node, "List", "fields");
    case "Field":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astList<Ident>(node, "Names", "names"),
        ...astMaybe(astField<Expr>(node, "Type", "type")),
        ...astMaybe(astField<BasicLit>(node, "Tag", "tag")),
        ...astMaybe(astField<CommentGroup>(node, "Comment", "comment"))
      ];
    case "ImportSpec":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astMaybe(astField<Ident>(node, "Name", "name")),
        ...astMaybe(astField<BasicLit>(node, "Path", "path")),
        ...astMaybe(astField<CommentGroup>(node, "Comment", "comment"))
      ];
    case "ValueSpec":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astList<Ident>(node, "Names", "names"),
        ...astMaybe(astField<Expr>(node, "Type", "type")),
        ...astList<Expr>(node, "Values", "values"),
        ...astMaybe(astField<CommentGroup>(node, "Comment", "comment"))
      ];
    case "TypeSpec":
      return [
        ...astMaybe(astField<CommentGroup>(node, "Doc", "doc")),
        ...astMaybe(astField<Ident>(node, "Name", "name")),
        ...astMaybe(astField<FieldList>(node, "TypeParams", "typeParams")),
        ...astMaybe(astField<Expr>(node, "Type", "type")),
        ...astMaybe(astField<CommentGroup>(node, "Comment", "comment"))
      ];
    case "FuncType":
      return [
        ...astMaybe(astField<FieldList>(node, "TypeParams", "typeParams")),
        ...astMaybe(astField<FieldList>(node, "Params", "params")),
        ...astMaybe(astField<FieldList>(node, "Results", "results"))
      ];
    case "BlockStmt":
      return astList<Stmt>(node, "List", "statements");
    case "DeclStmt":
      return astMaybe(astField<Decl>(node, "Decl", "decl"));
    case "LabeledStmt":
      return [
        ...astMaybe(astField<Ident>(node, "Label", "label")),
        ...astMaybe(astField<Stmt>(node, "Stmt", "stmt"))
      ];
    case "ExprStmt":
      return astMaybe(astField<Expr>(node, "X", "expr"));
    case "AssignStmt":
      return [...astList<Expr>(node, "Lhs", "lhs"), ...astList<Expr>(node, "Rhs", "rhs")];
    case "IncDecStmt":
      return astMaybe(astField<Expr>(node, "X", "expr"));
    case "ReturnStmt":
      return astList<Expr>(node, "Results", "results");
    case "BranchStmt":
      return astMaybe(astField<Ident>(node, "Label", "label"));
    case "IfStmt":
      return [
        ...astMaybe(astField<Stmt>(node, "Init", "init")),
        ...astMaybe(astField<Expr>(node, "Cond", "condition")),
        ...astMaybe(astField<BlockStmt>(node, "Body", "body")),
        ...astMaybe(astField<Stmt>(node, "Else", "else"))
      ];
    case "CaseClause":
      return [...astList<Expr>(node, "List", "list"), ...astList<Stmt>(node, "Body", "body")];
    case "CommClause":
      return [...astMaybe(astField<Stmt>(node, "Comm", "comm")), ...astList<Stmt>(node, "Body", "body")];
    case "SwitchStmt":
      return [
        ...astMaybe(astField<Stmt>(node, "Init", "init")),
        ...astMaybe(astField<Expr>(node, "Tag", "tag")),
        ...childNodes(astField<BlockStmt>(node, "Body") ?? { kind: "BlockStmt", statements: node.body } as BlockStmt)
      ];
    case "TypeSwitchStmt":
      return [
        ...astMaybe(astField<Stmt>(node, "Init", "init")),
        ...astMaybe(astField<Stmt>(node, "Assign", "assign")),
        ...childNodes(astField<BlockStmt>(node, "Body") ?? { kind: "BlockStmt", statements: node.body } as BlockStmt)
      ];
    case "SelectStmt":
      return childNodes(astField<BlockStmt>(node, "Body") ?? { kind: "BlockStmt", statements: node.body } as BlockStmt);
    case "ForStmt":
      return [
        ...astMaybe(astField<Stmt>(node, "Init", "init")),
        ...astMaybe(astField<Expr>(node, "Cond", "condition")),
        ...astMaybe(astField<Stmt>(node, "Post", "post")),
        ...astMaybe(astField<BlockStmt>(node, "Body", "body"))
      ];
    case "RangeStmt":
      return [
        ...astMaybe(astField<Expr>(node, "Key", "key")),
        ...astMaybe(astField<Expr>(node, "Value", "value")),
        ...astMaybe(astField<Expr>(node, "X", "source")),
        ...astMaybe(astField<BlockStmt>(node, "Body", "body"))
      ];
    case "DeferStmt":
      return astMaybe(astField<CallExpr>(node, "Call", "call"));
    case "GoStmt":
      return astMaybe(astField<CallExpr>(node, "Call", "call"));
    case "SendStmt":
      return [
        ...astMaybe(astField<Expr>(node, "Chan", "channel")),
        ...astMaybe(astField<Expr>(node, "Value", "value"))
      ];
    case "Ellipsis":
      return astMaybe(astField<Expr>(node, "Elt", "element"));
    case "FuncLit":
      return [
        ...astMaybe(astField<FuncType>(node, "Type", "type")),
        ...astMaybe(astField<BlockStmt>(node, "Body", "body"))
      ];
    case "CompositeLit":
      return [
        ...astMaybe(astField<Expr>(node, "Type", "type")),
        ...astList<Expr>(node, "Elts", "elements")
      ];
    case "ParenExpr":
      return astMaybe(astField<Expr>(node, "X", "expr"));
    case "SelectorExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "object")),
        ...astMaybe(astField<Ident>(node, "Sel", "selector"))
      ];
    case "IndexExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "object")),
        ...astMaybe(astField<Expr>(node, "Index", "index"))
      ];
    case "IndexListExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "object")),
        ...astList<Expr>(node, "Indices", "indices")
      ];
    case "SliceExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "object")),
        ...astMaybe(astField<Expr>(node, "Low", "low")),
        ...astMaybe(astField<Expr>(node, "High", "high")),
        ...astMaybe(astField<Expr>(node, "Max", "max"))
      ];
    case "TypeAssertExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "object")),
        ...astMaybe(astField<Expr>(node, "Type", "type"))
      ];
    case "CallExpr":
      return [
        ...astMaybe(astField<Expr>(node, "Fun", "fun")),
        ...astList<Expr>(node, "Args", "args")
      ];
    case "StarExpr":
      return astMaybe(astField<Expr>(node, "X", "expr"));
    case "UnaryExpr":
      return astMaybe(astField<Expr>(node, "X", "expr"));
    case "BinaryExpr":
      return [
        ...astMaybe(astField<Expr>(node, "X", "left")),
        ...astMaybe(astField<Expr>(node, "Y", "right"))
      ];
    case "KeyValueExpr":
      return [
        ...astMaybe(astField<Expr>(node, "Key", "key")),
        ...astMaybe(astField<Expr>(node, "Value", "value"))
      ];
    case "CellRefExpr":
      return [node.namespace];
    case "RangeRefExpr":
      return [node.namespace];
    case "ArrayType":
      return [
        ...astMaybe(astField<Expr>(node, "Len", "length")),
        ...astMaybe(astField<Expr>(node, "Elt", "element"))
      ];
    case "StructType":
      return astMaybe(astField<FieldList>(node, "Fields", "fields"));
    case "InterfaceType":
      return astMaybe(astField<FieldList>(node, "Methods", "methods"));
    case "MapType":
      return [
        ...astMaybe(astField<Expr>(node, "Key", "key")),
        ...astMaybe(astField<Expr>(node, "Value", "value"))
      ];
    case "ChanType":
      return astMaybe(astField<Expr>(node, "Value", "value"));
    default:
      return [];
  }
}
