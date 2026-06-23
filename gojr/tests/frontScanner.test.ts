import { describe, expect, test } from "./testHarness.js";
import { scanSource } from "../src/front/scanner.js";
import { TokenKind } from "../src/front/token.js";

const TEST_FILENAME = "front-scanner-test.go";

function scan(source: string) {
  return scanSource(source, TEST_FILENAME);
}

function kinds(source: string): TokenKind[] {
  return scan(source).tokens.map((token) => token.kind);
}

function lexemes(source: string): string[] {
  return scan(source).tokens.map((token) => token.lexeme);
}

describe("Go-junior TypeScript front scanner", () => {
  test("scans Go-like declarations, keywords, and operators", () => {
    expect(kinds(`func f(x int) int { return x + 1 }`)).toEqual([
      TokenKind.Func,
      TokenKind.Identifier,
      TokenKind.LParen,
      TokenKind.Identifier,
      TokenKind.Identifier,
      TokenKind.RParen,
      TokenKind.Identifier,
      TokenKind.LBrace,
      TokenKind.Return,
      TokenKind.Identifier,
      TokenKind.Plus,
      TokenKind.IntLiteral,
      TokenKind.RBrace,
      TokenKind.Semicolon,
      TokenKind.EOF
    ]);
  });

  test("inserts Go-style semicolons at newlines and EOF", () => {
    const result = scan("a := 1\nb := 2\nx := 1 +\n2");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens
      .filter((token) => token.kind === TokenKind.Semicolon)
      .map((token) => token.inserted)).toEqual(["newline", "newline", "eof"]);
    expect(lexemes("a := 1\nb := 2").filter((item) => item === ";")).toHaveLength(2);
  });

  test("scans spreadsheet cell and range tokens without confusing identifiers", () => {
    const result = scan(`Data.A1:B10\nsheet.$A$1:$B$10\nA1foo`);

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
      [TokenKind.Identifier, "Data"],
      [TokenKind.Dot, "."],
      [TokenKind.Identifier, "A1"],
      [TokenKind.Colon, ":"],
      [TokenKind.Identifier, "B10"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "sheet"],
      [TokenKind.Dot, "."],
      [TokenKind.CellAddress, "$A$1"],
      [TokenKind.Colon, ":"],
      [TokenKind.CellAddress, "$B$10"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "A1foo"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.EOF, ""]
    ]);
  });

  test("keeps Go identifiers that look like relative cells as identifiers", () => {
    const result = scan("func (S) M1(x I1) S2 { return S2{} }");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.filter((token) => ["S", "M1", "I1", "S2"].includes(token.lexeme)).map((token) => token.kind))
      .toEqual([TokenKind.Identifier, TokenKind.Identifier, TokenKind.Identifier, TokenKind.Identifier, TokenKind.Identifier]);
  });

  test("scans true false and nil as predeclared identifiers, not keywords", () => {
    const result = scan("var true = false\nvar nil = 1");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.filter((token) => ["true", "false", "nil"].includes(token.lexeme)).map((token) => token.kind))
      .toEqual([TokenKind.Identifier, TokenKind.Identifier, TokenKind.Identifier]);
  });

  test("scans literals and skips comments while preserving spans", () => {
    const result = scan("x := 1.5e2 // comment\ns := \"hi\\nthere\"");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.map((token) => token.lexeme)).toEqual([
      "x", ":=", "1.5e2", ";", "s", ":=", "\"hi\\nthere\"", ";", ""
    ]);
    expect(result.tokens[4]?.span).toMatchObject({ line: 2, column: 1 });
  });

  test("classifies hex integers with e digits separately from hex floats", () => {
    const result = scan("a := 0xe1\nb := 0xE1\nc := 0x1p2\nd := 0x_1\ne := 0o_7\nf := 0b_1");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
      [TokenKind.Identifier, "a"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "0xe1"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "b"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "0xE1"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "c"],
      [TokenKind.Define, ":="],
      [TokenKind.FloatLiteral, "0x1p2"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "d"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "0x_1"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "e"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "0o_7"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "f"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "0b_1"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.EOF, ""]
    ]);
  });

  test("scans Go raw string literals across newlines", () => {
    const result = scan("s := `hi\nthere`\nx := 1");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
      [TokenKind.Identifier, "s"],
      [TokenKind.Define, ":="],
      [TokenKind.StringLiteral, "`hi\nthere`"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "x"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "1"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.EOF, ""]
    ]);
    expect(result.tokens[4]?.span).toMatchObject({ line: 3, column: 1 });
  });

  test("scans Unicode identifiers and ignores a leading BOM like go/scanner", () => {
    const result = scan("\uFEFFπ := 3\n变量 := π");

    expect(result.diagnostics).toEqual([]);
    expect(result.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
      [TokenKind.Identifier, "π"],
      [TokenKind.Define, ":="],
      [TokenKind.IntLiteral, "3"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.Identifier, "变量"],
      [TokenKind.Define, ":="],
      [TokenKind.Identifier, "π"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.EOF, ""]
    ]);
  });

  test("reports NUL and non-leading BOM characters", () => {
    const result = scan("a\u0000b\nc\uFEFFd");

    expect(result.diagnostics.map((diagnostic) => diagnostic.message)).toEqual([
      "illegal character NUL",
      "illegal byte order mark"
    ]);
    expect(result.tokens.some((token) => token.kind === TokenKind.Illegal)).toBe(true);
  });

  test("scans channel arrows and malformed strings", () => {
    const channel = scan("x <- y");
    expect(channel.diagnostics).toEqual([]);
    expect(channel.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
      [TokenKind.Identifier, "x"],
      [TokenKind.Arrow, "<-"],
      [TokenKind.Identifier, "y"],
      [TokenKind.Semicolon, ";"],
      [TokenKind.EOF, ""]
    ]);

    const string = scan("\"unterminated\nnext");
    expect(string.diagnostics).toHaveLength(1);
    expect(string.diagnostics[0]?.message).toContain("unterminated string");

    const raw = scan("a := `unterminated");
    expect(raw.diagnostics).toHaveLength(1);
    expect(raw.diagnostics[0]?.message).toContain("unterminated raw string");
  });
});
