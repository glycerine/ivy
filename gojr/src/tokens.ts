import { createToken, Lexer } from "chevrotain";

export const WhiteSpace = createToken({
  name: "WhiteSpace",
  pattern: /[ \t\r\n]+/,
  group: Lexer.SKIPPED
});

export const LineComment = createToken({
  name: "LineComment",
  pattern: /\/\/[^\n\r]*/,
  group: Lexer.SKIPPED
});

export const BlockComment = createToken({
  name: "BlockComment",
  pattern: /\/\*[^]*?\*\//,
  group: Lexer.SKIPPED
});

export const IdentifierLike = createToken({
  name: "IdentifierLike",
  pattern: Lexer.NA
});

export const Identifier = createToken({
  name: "Identifier",
  pattern: /[A-Za-z_][A-Za-z0-9_]*/,
  categories: IdentifierLike
});

function keyword(name: string, text: string) {
  return createToken({
    name,
    pattern: new RegExp(text),
    longer_alt: Identifier
  });
}

export const Import = keyword("Import", "import");
export const Func = keyword("Func", "func");
export const Return = keyword("Return", "return");
export const If = keyword("If", "if");
export const Else = keyword("Else", "else");
export const Switch = keyword("Switch", "switch");
export const Case = keyword("Case", "case");
export const Default = keyword("Default", "default");
export const Fallthrough = keyword("Fallthrough", "fallthrough");
export const Goto = keyword("Goto", "goto");
export const For = keyword("For", "for");
export const Range = keyword("Range", "range");
export const Break = keyword("Break", "break");
export const Continue = keyword("Continue", "continue");
export const Defer = keyword("Defer", "defer");
export const Var = keyword("Var", "var");
export const Const = keyword("Const", "const");
export const Type = keyword("Type", "type");
export const Struct = keyword("Struct", "struct");
export const Interface = keyword("Interface", "interface");
export const MapTok = keyword("MapTok", "map");
export const True = keyword("True", "true");
export const False = keyword("False", "false");
export const Nil = keyword("Nil", "nil");

export const StringLiteral = createToken({
  name: "StringLiteral",
  pattern: /"(?:[^"\\]|\\.)*"/
});

export const FloatLiteral = createToken({
  name: "FloatLiteral",
  pattern: /(?:[0-9]+\.[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?/
});

export const IntLiteral = createToken({
  name: "IntLiteral",
  pattern: /0|[1-9][0-9]*/
});

export const CellAddress = createToken({
  name: "CellAddress",
  pattern: /\$?[A-Z]{1,3}\$?[1-9][0-9]*/,
  categories: IdentifierLike
});

export const Ellipsis = createToken({ name: "Ellipsis", pattern: /\.\.\./ });
export const Define = createToken({ name: "Define", pattern: /:=/ });
export const Equal = createToken({ name: "Equal", pattern: /==/ });
export const NotEqual = createToken({ name: "NotEqual", pattern: /!=/ });
export const LessEqual = createToken({ name: "LessEqual", pattern: /<=/ });
export const GreaterEqual = createToken({ name: "GreaterEqual", pattern: />=/ });
export const AndAnd = createToken({ name: "AndAnd", pattern: /&&/ });
export const OrOr = createToken({ name: "OrOr", pattern: /\|\|/ });
export const Arrow = createToken({ name: "Arrow", pattern: /->/ });
export const PlusPlus = createToken({ name: "PlusPlus", pattern: /\+\+/ });
export const MinusMinus = createToken({ name: "MinusMinus", pattern: /--/ });
export const Assign = createToken({ name: "Assign", pattern: /=/ });
export const Less = createToken({ name: "Less", pattern: /</ });
export const Greater = createToken({ name: "Greater", pattern: />/ });
export const Plus = createToken({ name: "Plus", pattern: /\+/ });
export const Minus = createToken({ name: "Minus", pattern: /-/ });
export const Star = createToken({ name: "Star", pattern: /\*/ });
export const Slash = createToken({ name: "Slash", pattern: /\// });
export const Percent = createToken({ name: "Percent", pattern: /%/ });
export const Bang = createToken({ name: "Bang", pattern: /!/ });
export const Amp = createToken({ name: "Amp", pattern: /&/ });
export const Dot = createToken({ name: "Dot", pattern: /\./ });
export const Comma = createToken({ name: "Comma", pattern: /,/ });
export const Colon = createToken({ name: "Colon", pattern: /:/ });
export const Semicolon = createToken({ name: "Semicolon", pattern: /;/ });
export const LParen = createToken({ name: "LParen", pattern: /\(/ });
export const RParen = createToken({ name: "RParen", pattern: /\)/ });
export const LBrace = createToken({ name: "LBrace", pattern: /\{/ });
export const RBrace = createToken({ name: "RBrace", pattern: /\}/ });
export const LBracket = createToken({ name: "LBracket", pattern: /\[/ });
export const RBracket = createToken({ name: "RBracket", pattern: /\]/ });

export const allTokens = [
  WhiteSpace,
  LineComment,
  BlockComment,
  Import,
  Func,
  Return,
  If,
  Else,
  Switch,
  Case,
  Default,
  Fallthrough,
  Goto,
  For,
  Range,
  Break,
  Continue,
  Defer,
  Var,
  Const,
  Type,
  Struct,
  Interface,
  MapTok,
  True,
  False,
  Nil,
  StringLiteral,
  FloatLiteral,
  IntLiteral,
  CellAddress,
  Identifier,
  Ellipsis,
  Define,
  Equal,
  NotEqual,
  LessEqual,
  GreaterEqual,
  AndAnd,
  OrOr,
  Arrow,
  PlusPlus,
  MinusMinus,
  Assign,
  Less,
  Greater,
  Plus,
  Minus,
  Star,
  Slash,
  Percent,
  Bang,
  Amp,
  Dot,
  Comma,
  Colon,
  Semicolon,
  LParen,
  RParen,
  LBrace,
  RBrace,
  LBracket,
  RBracket
];

export const goJuniorLexer = new Lexer(allTokens, {
  positionTracking: "full"
});
