import { describe, expect, test } from "./testHarness.js";
import {
  HighestPrec,
  IsExported,
  IsIdentifier,
  IsKeyword,
  IsKeywordToken,
  IsLiteral,
  IsOperator,
  Lookup,
  NewFileSet,
  NoPos,
  Position,
  PosIsValid,
  Precedence,
  Token,
  TokenString
} from "../src/go/token/index.js";

describe("Go-junior go/token", () => {
  test("transliterates token strings, lookup, predicates, and precedence", () => {
    expect(TokenString(Token.ADD)).toBe("+");
    expect(Token.String(Token.IDENT)).toBe("IDENT");
    expect(TokenString(999 as Token)).toBe("token(999)");
    expect(Lookup("func")).toBe(Token.FUNC);
    expect(Lookup("ordinary")).toBe(Token.IDENT);
    expect(IsLiteral(Token.STRING)).toBe(true);
    expect(Token.IsLiteral(Token.FLOAT)).toBe(true);
    expect(IsOperator(Token.TILDE)).toBe(true);
    expect(Token.IsOperator(Token.ADD_ASSIGN)).toBe(true);
    expect(IsKeywordToken(Token.RETURN)).toBe(true);
    expect(Token.IsKeyword(Token.RETURN)).toBe(true);
    expect(IsKeyword("return")).toBe(true);
    expect(IsIdentifier("return")).toBe(false);
    expect(IsIdentifier("_thing9")).toBe(true);
    expect(IsExported("Thing")).toBe(true);
    expect(IsExported("thing")).toBe(false);
    expect(Precedence(Token.LOR)).toBe(1);
    expect(Token.Precedence(Token.MUL)).toBe(5);
    expect(HighestPrec).toBe(7);
  });

  test("transliterates positions, files, and file sets", () => {
    const invalid = new Position();
    expect(invalid.IsValid()).toBe(false);
    expect(invalid.String()).toBe("-");
    expect(PosIsValid(NoPos)).toBe(false);

    const fset = NewFileSet();
    const file = fset.AddFile("a.go", -1, 12);
    file.SetLinesForContent("ab\nc\nlast");
    expect(file.Name()).toBe("a.go");
    expect(file.Base()).toBe(1);
    expect(file.End()).toBe(13);
    expect(file.LineCount()).toBe(3);
    expect(file.LineStart(2)).toBe(file.Pos(3));
    expect(file.Position(file.Pos(4)).String()).toBe("a.go:2:2");

    file.AddLineColumnInfo(3, "mapped.go", 20, 4);
    expect(file.Position(file.Pos(4)).String()).toBe("mapped.go:20:5");
    expect(file.PositionFor(file.Pos(4), false).String()).toBe("a.go:2:2");
    expect(fset.File(file.Pos(5))).toBe(file);
    expect(fset.Position(file.Pos(5)).String()).toBe("mapped.go:21:1");
  });
});
