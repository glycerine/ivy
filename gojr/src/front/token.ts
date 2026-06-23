import { SourceSpan } from "../diagnostics.js";

export enum TokenKind {
  Illegal = "Illegal",
  EOF = "EOF",

  Identifier = "Identifier",
  CellAddress = "CellAddress",
  IntLiteral = "IntLiteral",
  FloatLiteral = "FloatLiteral",
  ImagLiteral = "ImagLiteral",
  RuneLiteral = "RuneLiteral",
  StringLiteral = "StringLiteral",

  Import = "Import",
  Package = "Package",
  Func = "Func",
  Return = "Return",
  If = "If",
  Else = "Else",
  Switch = "Switch",
  Case = "Case",
  Default = "Default",
  Fallthrough = "Fallthrough",
  Goto = "Goto",
  For = "For",
  Range = "Range",
  Break = "Break",
  Continue = "Continue",
  Defer = "Defer",
  Var = "Var",
  Const = "Const",
  Type = "Type",
  Struct = "Struct",
  Interface = "Interface",
  Map = "Map",
  Chan = "Chan",
  Go = "Go",
  Select = "Select",
  True = "True",
  False = "False",
  Nil = "Nil",

  Ellipsis = "Ellipsis",
  Define = "Define",
  Assign = "Assign",
  PlusAssign = "PlusAssign",
  MinusAssign = "MinusAssign",
  StarAssign = "StarAssign",
  SlashAssign = "SlashAssign",
  PercentAssign = "PercentAssign",
  AmpAssign = "AmpAssign",
  OrAssign = "OrAssign",
  CaretAssign = "CaretAssign",
  ShlAssign = "ShlAssign",
  ShrAssign = "ShrAssign",
  BitClearAssign = "BitClearAssign",
  Equal = "Equal",
  NotEqual = "NotEqual",
  Less = "Less",
  LessEqual = "LessEqual",
  Greater = "Greater",
  GreaterEqual = "GreaterEqual",
  AndAnd = "AndAnd",
  OrOr = "OrOr",
  Arrow = "Arrow",
  Plus = "Plus",
  PlusPlus = "PlusPlus",
  Minus = "Minus",
  MinusMinus = "MinusMinus",
  Star = "Star",
  Slash = "Slash",
  Percent = "Percent",
  Or = "Or",
  Caret = "Caret",
  Shl = "Shl",
  Shr = "Shr",
  BitClear = "BitClear",
  Bang = "Bang",
  Amp = "Amp",
  Dot = "Dot",
  Comma = "Comma",
  Colon = "Colon",
  Semicolon = "Semicolon",
  LParen = "LParen",
  RParen = "RParen",
  LBrace = "LBrace",
  RBrace = "RBrace",
  LBracket = "LBracket",
  RBracket = "RBracket"
}

export interface FrontToken {
  kind: TokenKind;
  lexeme: string;
  span: SourceSpan;
  inserted?: "newline" | "eof";
}

const keywords = new Map<string, TokenKind>([
  ["import", TokenKind.Import],
  ["package", TokenKind.Package],
  ["func", TokenKind.Func],
  ["return", TokenKind.Return],
  ["if", TokenKind.If],
  ["else", TokenKind.Else],
  ["switch", TokenKind.Switch],
  ["case", TokenKind.Case],
  ["default", TokenKind.Default],
  ["fallthrough", TokenKind.Fallthrough],
  ["goto", TokenKind.Goto],
  ["for", TokenKind.For],
  ["range", TokenKind.Range],
  ["break", TokenKind.Break],
  ["continue", TokenKind.Continue],
  ["defer", TokenKind.Defer],
  ["var", TokenKind.Var],
  ["const", TokenKind.Const],
  ["type", TokenKind.Type],
  ["struct", TokenKind.Struct],
  ["interface", TokenKind.Interface],
  ["map", TokenKind.Map],
  ["chan", TokenKind.Chan],
  ["go", TokenKind.Go],
  ["select", TokenKind.Select],
  ["true", TokenKind.True],
  ["false", TokenKind.False],
  ["nil", TokenKind.Nil]
]);

export function keywordKind(text: string): TokenKind | undefined {
  return keywords.get(text);
}

export function isIdentifierLike(kind: TokenKind): boolean {
  return kind === TokenKind.Identifier || kind === TokenKind.CellAddress;
}

export function tokenCanEndStatement(kind: TokenKind): boolean {
  return kind === TokenKind.Identifier ||
    kind === TokenKind.CellAddress ||
    kind === TokenKind.IntLiteral ||
    kind === TokenKind.FloatLiteral ||
    kind === TokenKind.ImagLiteral ||
    kind === TokenKind.RuneLiteral ||
    kind === TokenKind.StringLiteral ||
    kind === TokenKind.True ||
    kind === TokenKind.False ||
    kind === TokenKind.Nil ||
    kind === TokenKind.Return ||
    kind === TokenKind.Break ||
    kind === TokenKind.Continue ||
    kind === TokenKind.Fallthrough ||
    kind === TokenKind.PlusPlus ||
    kind === TokenKind.MinusMinus ||
    kind === TokenKind.RParen ||
    kind === TokenKind.RBracket ||
    kind === TokenKind.RBrace;
}

export type AssignmentToken =
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

export function isAssignmentToken(kind: TokenKind): kind is AssignmentToken {
  return kind === TokenKind.Assign ||
    kind === TokenKind.Define ||
    kind === TokenKind.PlusAssign ||
    kind === TokenKind.MinusAssign ||
    kind === TokenKind.StarAssign ||
    kind === TokenKind.SlashAssign ||
    kind === TokenKind.PercentAssign ||
    kind === TokenKind.AmpAssign ||
    kind === TokenKind.OrAssign ||
    kind === TokenKind.CaretAssign ||
    kind === TokenKind.ShlAssign ||
    kind === TokenKind.ShrAssign ||
    kind === TokenKind.BitClearAssign;
}
