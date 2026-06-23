import { Diagnostic, SourceSpan } from "../diagnostics.js";
import { FrontToken, keywordKind, tokenCanEndStatement, TokenKind } from "./token.js";

export interface ScanResult {
  tokens: FrontToken[];
  diagnostics: Diagnostic[];
}

interface Position {
  offset: number;
  line: number;
  column: number;
}

export function scanSource(source: string): ScanResult {
  return new Scanner(source).scan();
}

class Scanner {
  private offset = 0;
  private line = 1;
  private column = 1;
  private insertSemi = false;
  private emittedEOF = false;
  private readonly tokens: FrontToken[] = [];
  private readonly diagnostics: Diagnostic[] = [];

  public constructor(private readonly source: string) {}

  public scan(): ScanResult {
    while (!this.emittedEOF) {
      this.scanOne();
    }
    return {
      tokens: this.tokens,
      diagnostics: this.diagnostics
    };
  }

  private scanOne(): void {
    const skipped = this.skipTrivia();
    if (skipped) return;

    if (this.isEOF()) {
      if (this.insertSemi) {
        this.emitSyntheticSemicolon(this.position(), "eof");
        this.insertSemi = false;
        return;
      }
      this.emit(TokenKind.EOF, "", this.position(), this.position());
      this.emittedEOF = true;
      return;
    }

    const start = this.position();
    const char = this.peek();

    if (char === "\"" ) {
      this.scanString(start);
      return;
    }
    if (isDigit(char) || (char === "." && isDigit(this.peek(1)))) {
      this.scanNumber(start);
      return;
    }
    if (char === "$" || isIdentifierStart(char)) {
      this.scanIdentifierOrCell(start);
      return;
    }

    this.scanOperatorOrDelimiter(start);
  }

  private skipTrivia(): boolean {
    while (!this.isEOF()) {
      const char = this.peek();
      if (char === " " || char === "\t" || char === "\f") {
        this.advance();
        continue;
      }
      if (char === "\r" || char === "\n") {
        const newlineStart = this.position();
        this.advanceNewline();
        if (this.insertSemi) {
          this.emitSyntheticSemicolon(newlineStart, "newline");
          this.insertSemi = false;
          return true;
        }
        continue;
      }
      if (char === "/" && this.peek(1) === "/") {
        this.skipLineComment();
        continue;
      }
      if (char === "/" && this.peek(1) === "*") {
        const commentStart = this.position();
        const hadNewline = this.skipBlockComment(commentStart);
        if (hadNewline && this.insertSemi) {
          this.emitSyntheticSemicolon(this.position(), "newline");
          this.insertSemi = false;
          return true;
        }
        continue;
      }
      return false;
    }
    return false;
  }

  private scanString(start: Position): void {
    this.advance();
    while (!this.isEOF()) {
      const char = this.peek();
      if (char === "\"") {
        this.advance();
        this.emit(TokenKind.StringLiteral, this.sliceFrom(start), start, this.position());
        return;
      }
      if (char === "\n" || char === "\r") {
        this.error("GJSCAN001", "unterminated string literal", start, this.position());
        this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
        return;
      }
      if (char === "\\") {
        this.advance();
        if (!this.isEOF()) this.advance();
        continue;
      }
      this.advance();
    }
    this.error("GJSCAN001", "unterminated string literal", start, this.position());
    this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
  }

  private scanNumber(start: Position): void {
    let kind = TokenKind.IntLiteral;

    if (this.peek() === ".") {
      kind = TokenKind.FloatLiteral;
      this.advance();
      this.consumeDigits();
    } else {
      this.consumeDigits();
      if (this.peek() === ".") {
        kind = TokenKind.FloatLiteral;
        this.advance();
        this.consumeDigits();
      }
    }

    if (this.peek().toLowerCase() === "e") {
      kind = TokenKind.FloatLiteral;
      this.advance();
      if (this.peek() === "+" || this.peek() === "-") this.advance();
      if (!isDigit(this.peek())) {
        this.error("GJSCAN002", "malformed exponent in numeric literal", start, this.position());
      } else {
        this.consumeDigits();
      }
    }

    this.emit(kind, this.sliceFrom(start), start, this.position());
  }

  private scanIdentifierOrCell(start: Position): void {
    const cell = this.matchCellAddressAt(this.offset);
    if (cell) {
      this.advanceMany(cell.length);
      this.emit(TokenKind.CellAddress, cell, start, this.position());
      return;
    }

    if (this.peek() === "$") {
      this.advance();
      this.error("GJSCAN003", "malformed spreadsheet cell address", start, this.position());
      this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
      return;
    }

    this.advance();
    while (isIdentifierPart(this.peek())) this.advance();
    const text = this.sliceFrom(start);
    this.emit(keywordKind(text) ?? TokenKind.Identifier, text, start, this.position());
  }

  private scanOperatorOrDelimiter(start: Position): void {
    const two = this.source.slice(this.offset, this.offset + 2);
    const three = this.source.slice(this.offset, this.offset + 3);
    if (three === "...") {
      this.advanceMany(3);
      this.emit(TokenKind.Ellipsis, three, start, this.position());
      return;
    }

    const twoKind = twoCharToken(two);
    if (twoKind) {
      this.advanceMany(2);
      this.emit(twoKind, two, start, this.position());
      if (two === "<-") {
        this.error("GJSCAN004", "channels and channel operations are not supported", start, this.position());
      } else if (two === "->") {
        this.error("GJSCAN006", "C/C++ pointer selector syntax is not supported; use Go-style '.' method calls", start, this.position());
      }
      return;
    }

    const one = this.peek();
    const oneKind = oneCharToken(one);
    if (oneKind) {
      this.advance();
      this.emit(oneKind, one, start, this.position());
      return;
    }

    this.advance();
    this.error("GJSCAN005", `unexpected character ${JSON.stringify(one)}`, start, this.position());
    this.emit(TokenKind.Illegal, one, start, this.position());
  }

  private consumeDigits(): void {
    while (isDigit(this.peek())) this.advance();
  }

  private skipLineComment(): void {
    while (!this.isEOF() && this.peek() !== "\n" && this.peek() !== "\r") {
      this.advance();
    }
  }

  private skipBlockComment(start: Position): boolean {
    this.advanceMany(2);
    let hadNewline = false;
    while (!this.isEOF()) {
      if (this.peek() === "*" && this.peek(1) === "/") {
        this.advanceMany(2);
        return hadNewline;
      }
      if (this.peek() === "\n" || this.peek() === "\r") {
        hadNewline = true;
        this.advanceNewline();
        continue;
      }
      this.advance();
    }
    this.error("GJSCAN006", "unterminated block comment", start, this.position());
    return hadNewline;
  }

  private emit(kind: TokenKind, lexeme: string, start: Position, end: Position, inserted?: "newline" | "eof"): void {
    const token: FrontToken = {
      kind,
      lexeme,
      span: spanBetween(start, end),
      ...(inserted ? { inserted } : {})
    };
    this.tokens.push(token);
    this.insertSemi = tokenCanEndStatement(kind);
  }

  private emitSyntheticSemicolon(position: Position, inserted: "newline" | "eof"): void {
    this.emit(TokenKind.Semicolon, ";", position, position, inserted);
  }

  private error(code: string, message: string, start: Position, end: Position): void {
    this.diagnostics.push({
      code,
      severity: "error",
      message,
      span: spanBetween(start, end)
    });
  }

  private matchCellAddressAt(offset: number): string | undefined {
    const rest = this.source.slice(offset);
    const match = /^\$?[A-Z]{1,3}\$?[1-9][0-9]*/.exec(rest);
    if (!match) return undefined;
    const text = match[0];
    const next = rest[text.length] ?? "";
    return isIdentifierPart(next) || next === "$" ? undefined : text;
  }

  private position(): Position {
    return {
      offset: this.offset,
      line: this.line,
      column: this.column
    };
  }

  private sliceFrom(start: Position): string {
    return this.source.slice(start.offset, this.offset);
  }

  private peek(ahead = 0): string {
    return this.source[this.offset + ahead] ?? "";
  }

  private isEOF(): boolean {
    return this.offset >= this.source.length;
  }

  private advance(): void {
    this.offset += 1;
    this.column += 1;
  }

  private advanceMany(count: number): void {
    for (let index = 0; index < count; index += 1) this.advance();
  }

  private advanceNewline(): void {
    if (this.peek() === "\r" && this.peek(1) === "\n") {
      this.offset += 2;
    } else {
      this.offset += 1;
    }
    this.line += 1;
    this.column = 1;
  }
}

function spanBetween(start: Position, end: Position): SourceSpan {
  return {
    offset: start.offset,
    length: Math.max(0, end.offset - start.offset),
    line: start.line,
    column: start.column
  };
}

function twoCharToken(text: string): TokenKind | undefined {
  switch (text) {
    case ":=": return TokenKind.Define;
    case "==": return TokenKind.Equal;
    case "!=": return TokenKind.NotEqual;
    case "<=": return TokenKind.LessEqual;
    case ">=": return TokenKind.GreaterEqual;
    case "&&": return TokenKind.AndAnd;
    case "||": return TokenKind.OrOr;
    case "->": return TokenKind.Arrow;
    case "++": return TokenKind.PlusPlus;
    case "--": return TokenKind.MinusMinus;
    case "<-": return TokenKind.Illegal;
    default: return undefined;
  }
}

function oneCharToken(text: string): TokenKind | undefined {
  switch (text) {
    case "=": return TokenKind.Assign;
    case "<": return TokenKind.Less;
    case ">": return TokenKind.Greater;
    case "+": return TokenKind.Plus;
    case "-": return TokenKind.Minus;
    case "*": return TokenKind.Star;
    case "/": return TokenKind.Slash;
    case "%": return TokenKind.Percent;
    case "!": return TokenKind.Bang;
    case "&": return TokenKind.Amp;
    case ".": return TokenKind.Dot;
    case ",": return TokenKind.Comma;
    case ":": return TokenKind.Colon;
    case ";": return TokenKind.Semicolon;
    case "(": return TokenKind.LParen;
    case ")": return TokenKind.RParen;
    case "{": return TokenKind.LBrace;
    case "}": return TokenKind.RBrace;
    case "[": return TokenKind.LBracket;
    case "]": return TokenKind.RBracket;
    default: return undefined;
  }
}

function isIdentifierStart(text: string): boolean {
  return /^[A-Za-z_]$/.test(text);
}

function isIdentifierPart(text: string): boolean {
  return /^[A-Za-z0-9_]$/.test(text);
}

function isDigit(text: string): boolean {
  return /^[0-9]$/.test(text);
}
