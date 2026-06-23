import { describe, expect, test } from "vitest";
import { scanSource } from "../src/front/scanner.js";
import { TokenKind } from "../src/front/token.js";
function kinds(source) {
    return scanSource(source).tokens.map((token) => token.kind);
}
function lexemes(source) {
    return scanSource(source).tokens.map((token) => token.lexeme);
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
        const result = scanSource("a := 1\nb := 2\nx := 1 +\n2");
        expect(result.diagnostics).toEqual([]);
        expect(result.tokens
            .filter((token) => token.kind === TokenKind.Semicolon)
            .map((token) => token.inserted)).toEqual(["newline", "newline", "eof"]);
        expect(lexemes("a := 1\nb := 2").filter((item) => item === ";")).toHaveLength(2);
    });
    test("scans spreadsheet cell and range tokens without confusing identifiers", () => {
        const result = scanSource(`Data.A1:B10\nsheet.$A$1:$B$10\nA1foo`);
        expect(result.diagnostics).toEqual([]);
        expect(result.tokens.map((token) => [token.kind, token.lexeme])).toEqual([
            [TokenKind.Identifier, "Data"],
            [TokenKind.Dot, "."],
            [TokenKind.CellAddress, "A1"],
            [TokenKind.Colon, ":"],
            [TokenKind.CellAddress, "B10"],
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
    test("scans literals and skips comments while preserving spans", () => {
        const result = scanSource("x := 1.5e2 // comment\ns := \"hi\\nthere\"");
        expect(result.diagnostics).toEqual([]);
        expect(result.tokens.map((token) => token.lexeme)).toEqual([
            "x", ":=", "1.5e2", ";", "s", ":=", "\"hi\\nthere\"", ";", ""
        ]);
        expect(result.tokens[4]?.span).toMatchObject({ line: 2, column: 1 });
    });
    test("scans Go raw string literals across newlines", () => {
        const result = scanSource("s := `hi\nthere`\nx := 1");
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
    test("reports unsupported channels and malformed strings", () => {
        const channel = scanSource("x <- y");
        expect(channel.diagnostics).toHaveLength(1);
        expect(channel.diagnostics[0]?.message).toContain("channels");
        const string = scanSource("\"unterminated\nnext");
        expect(string.diagnostics).toHaveLength(1);
        expect(string.diagnostics[0]?.message).toContain("unterminated string");
        const raw = scanSource("a := `unterminated");
        expect(raw.diagnostics).toHaveLength(1);
        expect(raw.diagnostics[0]?.message).toContain("unterminated raw string");
    });
});
