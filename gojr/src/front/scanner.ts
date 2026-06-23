import { Diagnostic, REPL_FILENAME, SourceSpan } from "../diagnostics.js";
import { FrontToken, keywordKind, tokenCanEndStatement, TokenKind } from "./token.js";

export interface ScanResult {
  tokens: FrontToken[];
  diagnostics: Diagnostic[];
}

export interface Position {
  filename: string;
  offset: number;
  line: number;
  column: number;
}

// Error mirrors go/scanner.Error.
export class Error {
  public constructor(
    public Pos: Position,
    public Msg: string
  ) {}

  public Error(): string {
    if (this.Pos.filename !== "" || positionIsValid(this.Pos)) {
      return `${positionString(this.Pos)}: ${this.Msg}`;
    }
    return this.Msg;
  }

  public toString(): string {
    return this.Error();
  }
}

// ErrorList mirrors go/scanner.ErrorList.
export class ErrorList extends Array<Error> {
  public Add(pos: Position, msg: string): void {
    this.push(new Error(pos, msg));
  }

  public Reset(): void {
    this.length = 0;
  }

  public Len(): number {
    return this.length;
  }

  public Swap(i: number, j: number): void {
    const item = this[i];
    this[i] = this[j]!;
    this[j] = item!;
  }

  public Less(i: number, j: number): boolean {
    const e = this[i]!.Pos;
    const f = this[j]!.Pos;
    if (e.filename !== f.filename) {
      return e.filename < f.filename;
    }
    if (e.line !== f.line) {
      return e.line < f.line;
    }
    if (e.column !== f.column) {
      return e.column < f.column;
    }
    return this[i]!.Msg < this[j]!.Msg;
  }

  public Sort(): void {
    this.sort((left, right) => compareScannerErrors(left, right));
  }

  public RemoveMultiples(): void {
    this.Sort();
    let last: Position = { filename: "", offset: 0, line: 0, column: 0 };
    let index = 0;
    for (const error of this) {
      if (error.Pos.filename !== last.filename || error.Pos.line !== last.line) {
        last = error.Pos;
        this[index] = error;
        index += 1;
      }
    }
    this.length = index;
  }

  public Error(): string {
    switch (this.length) {
      case 0:
        return "no errors";
      case 1:
        return this[0]!.Error();
      default:
        return `${this[0]} (and ${this.length - 1} more errors)`;
    }
  }

  public Err(): ErrorList | undefined {
    if (this.length === 0) {
      return undefined;
    }
    return this;
  }

  public override toString(): string {
    return this.Error();
  }
}

export interface ScannerErrorWriter {
  write(text: string): unknown;
}

export function PrintError(w: ScannerErrorWriter | ((text: string) => unknown), err: unknown): void {
  if (err instanceof ErrorList) {
    for (const item of err) {
      writeScannerError(w, `${item}\n`);
    }
  } else if (err !== undefined && err !== null) {
    writeScannerError(w, `${String(err)}\n`);
  }
}

export type ErrorHandler = (pos: Position, msg: string) => void;
export type Mode = number;

export const ScanComments: Mode = 1 << 0;
export const dontInsertSemis: Mode = 1 << 1;

const bom = 0xFEFF;
const eof = -1;
const prefix = "line ";

function init(): void {
  return;
}

function compareScannerErrors(left: Error, right: Error): number {
  if (left.Pos.filename !== right.Pos.filename) return left.Pos.filename < right.Pos.filename ? -1 : 1;
  if (left.Pos.line !== right.Pos.line) return left.Pos.line - right.Pos.line;
  if (left.Pos.column !== right.Pos.column) return left.Pos.column - right.Pos.column;
  if (left.Msg === right.Msg) return 0;
  return left.Msg < right.Msg ? -1 : 1;
}

function positionIsValid(position: Position): boolean {
  return position.line > 0;
}

function positionString(position: Position): string {
  const file = position.filename;
  const line = position.line;
  const column = position.column;
  if (file !== "") {
    if (line > 0) {
      if (column > 0) return `${file}:${line}:${column}`;
      return `${file}:${line}`;
    }
    return file;
  }
  if (line > 0) {
    if (column > 0) return `${line}:${column}`;
    return String(line);
  }
  return "-";
}

function writeScannerError(w: ScannerErrorWriter | ((text: string) => unknown), text: string): void {
  if (typeof w === "function") {
    w(text);
  } else {
    w.write(text);
  }
}

export function scanSource(source: string, filename: string): ScanResult {
  return new Scanner(source, filename).Scan();
}

export class Scanner {
  private offset = 0;
  private line = 1;
  private column = 1;
  private insertSemi = false;
  private emittedEOF = false;
  private tokens: FrontToken[] = [];
  private diagnostics: Diagnostic[] = [];
  private err: ErrorHandler | undefined;
  private mode: Mode = 0;
  public ErrorCount = 0;

  public constructor(
    private source: string,
    private filename: string
  ) {}

  public Init(filename: string, source: string, err?: ErrorHandler, mode: Mode = 0): void {
    this.filename = filename;
    this.source = source;
    this.err = err;
    this.mode = mode;
    this.offset = 0;
    this.line = 1;
    this.column = 1;
    this.insertSemi = false;
    this.emittedEOF = false;
    this.tokens = [];
    this.diagnostics = [];
    this.ErrorCount = 0;
    if (this.offset === 0 && this.currentChar() === String.fromCodePoint(bom)) this.advanceChar();
  }

  public Scan(): ScanResult {
    return this.scan();
  }

  public scan(): ScanResult {
    if (this.offset === 0 && this.currentChar() === String.fromCodePoint(bom)) this.advanceChar();
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
    const char = this.currentChar();

    if (char === "\u0000") {
      this.advanceChar();
      this.error("GOJR_SCAN005", "illegal character NUL", start, this.position());
      this.emit(TokenKind.Illegal, char, start, this.position());
      return;
    }
    if (char === "\uFEFF") {
      this.advanceChar();
      this.error("GOJR_SCAN005", "illegal byte order mark", start, this.position());
      this.emit(TokenKind.Illegal, char, start, this.position());
      return;
    }

    if (char === "\"" ) {
      this.scanString(start);
      return;
    }
    if (char === "'") {
      this.scanRune(start);
      return;
    }
    if (char === "`") {
      this.scanRawString(start);
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

  private next(): void {
    if (this.peek() === "\r" || this.peek() === "\n") {
      this.advanceNewline();
    } else {
      this.advanceChar();
    }
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
        this.scanComment();
        continue;
      }
      if (char === "/" && this.peek(1) === "*") {
        const commentStart = this.position();
        const comment = this.scanComment();
        if (comment.nlOffset > 0 && this.insertSemi) {
          this.emitSyntheticSemicolon(this.positionAt(comment.nlOffset), "newline");
          this.insertSemi = false;
          return true;
        }
        if (comment.text === "" && this.offset === commentStart.offset) this.skipBlockComment(commentStart);
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
        this.error("GOJR_SCAN001", "unterminated string literal", start, this.position());
        this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
        return;
      }
      if (char === "\\") {
        this.advance();
        this.scanEscape("\"");
        continue;
      }
      this.advance();
    }
    this.error("GOJR_SCAN001", "unterminated string literal", start, this.position());
    this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
  }

  private scanRawString(start: Position): void {
    this.advance();
    while (!this.isEOF()) {
      const char = this.peek();
      if (char === "`") {
        this.advance();
        this.emit(TokenKind.StringLiteral, this.sliceFrom(start), start, this.position());
        return;
      }
      if (char === "\n" || char === "\r") {
        this.advanceNewline();
        continue;
      }
      this.advance();
    }
    this.error("GOJR_SCAN001", "unterminated raw string literal", start, this.position());
    this.emit(TokenKind.StringLiteral, this.sliceFrom(start), start, this.position());
  }

  private scanRune(start: Position): void {
    this.advance();
    let sawContent = false;
    while (!this.isEOF()) {
      const char = this.peek();
      if (char === "'") {
        this.advance();
        this.emit(TokenKind.RuneLiteral, this.sliceFrom(start), start, this.position());
        return;
      }
      if (char === "\n" || char === "\r") {
        this.error("GOJR_SCAN001", "unterminated rune literal", start, this.position());
        this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
        return;
      }
      sawContent = true;
      if (char === "\\") {
        this.advance();
        this.scanEscape("'");
        continue;
      }
      this.advance();
    }
    this.error("GOJR_SCAN001", sawContent ? "unterminated rune literal" : "empty rune literal", start, this.position());
    this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
  }

  private scanNumber(start: Position): void {
    let kind = TokenKind.Illegal;
    let base = 10;
    let prefix = "";
    let digsep = 0;
    const invalid = { offset: -1 };

    if (this.peek() !== ".") {
      kind = TokenKind.IntLiteral;
      if (this.peek() === "0") {
        this.advance();
        switch (lower(this.peek())) {
          case "x":
            this.advance();
            base = 16;
            prefix = "x";
            break;
          case "o":
            this.advance();
            base = 8;
            prefix = "o";
            break;
          case "b":
            this.advance();
            base = 2;
            prefix = "b";
            break;
          default:
            base = 8;
            prefix = "0";
            digsep = 1;
            break;
        }
      }
      digsep |= this.digits(base, invalid);
    }

    if (this.peek() === ".") {
      kind = TokenKind.FloatLiteral;
      if (prefix === "o" || prefix === "b") {
        this.errorAt("GOJR_SCAN002", `invalid radix point in ${litname(prefix)}`, this.offset);
      }
      this.advance();
      digsep |= this.digits(base, invalid);
    }

    if ((digsep & 1) === 0) {
      this.errorAt("GOJR_SCAN002", `${litname(prefix)} has no digits`, this.offset);
    }

    const exponent = lower(this.peek());
    if (exponent === "e" || exponent === "p") {
      if (exponent === "e" && prefix !== "" && prefix !== "0") {
        this.errorAt("GOJR_SCAN002", `'${this.peek()}' exponent requires decimal mantissa`, this.offset);
      } else if (exponent === "p" && prefix !== "x") {
        this.errorAt("GOJR_SCAN002", `'${this.peek()}' exponent requires hexadecimal mantissa`, this.offset);
      }
      this.advance();
      kind = TokenKind.FloatLiteral;
      if (this.peek() === "+" || this.peek() === "-") this.advance();
      const exponentDigits = this.digits(10, undefined);
      digsep |= exponentDigits;
      if ((exponentDigits & 1) === 0) {
        this.errorAt("GOJR_SCAN002", "exponent has no digits", this.offset);
      }
    } else if (prefix === "x" && kind === TokenKind.FloatLiteral) {
      this.errorAt("GOJR_SCAN002", "hexadecimal mantissa requires a 'p' exponent", this.offset);
    }

    if (this.peek() === "i") {
      kind = TokenKind.ImagLiteral;
      this.advance();
    }

    const text = this.sliceFrom(start);
    if (kind === TokenKind.IntLiteral && invalid.offset >= 0) {
      this.errorAt("GOJR_SCAN002", `invalid digit '${this.source[invalid.offset]}' in ${litname(prefix)}`, invalid.offset);
    }
    if ((digsep & 2) !== 0) {
      const badSeparator = invalidSep(text);
      if (badSeparator >= 0) {
        this.errorAt("GOJR_SCAN002", "'_' must separate successive digits", start.offset + badSeparator);
      }
    }
    this.emit(kind, text, start, this.position());
  }

  private scanIdentifierOrCell(start: Position): void {
    if (this.peek() === "$") {
      const cell = this.matchCellAddressAt(this.offset);
      if (cell) {
        this.advanceMany(cell.length);
        this.emit(TokenKind.CellAddress, cell, start, this.position());
        return;
      }
      this.advance();
      this.error("GOJR_SCAN003", "malformed spreadsheet cell address", start, this.position());
      this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
      return;
    }

    const text = this.scanIdentifier();
    this.emit(keywordKind(text) ?? TokenKind.Identifier, text, start, this.position());
  }

  private scanIdentifier(): string {
    const start = this.position();
    this.advanceChar();
    while (isIdentifierPart(this.currentChar())) this.advanceChar();
    return this.sliceFrom(start);
  }

  private scanOperatorOrDelimiter(start: Position): void {
    const three = this.source.slice(this.offset, this.offset + 3);
    if (three === "...") {
      this.advanceMany(3);
      this.emit(TokenKind.Ellipsis, three, start, this.position());
      return;
    }
    const threeKind = threeCharToken(three);
    if (threeKind) {
      this.advanceMany(3);
      this.emit(threeKind, three, start, this.position());
      return;
    }

    const two = this.source.slice(this.offset, this.offset + 2);
    const twoKind = twoCharToken(two);
    if (twoKind) {
      this.advanceMany(2);
      this.emit(twoKind, two, start, this.position());
      if (two === "->") {
        this.error("GOJR_SCAN006", "C/C++ pointer selector syntax is not supported; use Go-style '.' method calls", start, this.position());
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
    this.error("GOJR_SCAN005", `unexpected character ${JSON.stringify(one)}`, start, this.position());
    this.emit(TokenKind.Illegal, one, start, this.position());
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
    this.error("GOJR_SCAN006", "unterminated block comment", start, this.position());
    return hadNewline;
  }

  private scanComment(): { text: string; nlOffset: number } {
    const offs = this.offset;
    let next = -1;
    let numCR = 0;
    let nlOffset = 0;

    if (this.peek() !== "/" || (this.peek(1) !== "/" && this.peek(1) !== "*")) {
      return { text: "", nlOffset: 0 };
    }

    const block = this.peek(1) === "*";
    this.advanceMany(2);
    if (!block) {
      while (!this.isEOF() && this.peek() !== "\n" && this.peek() !== "\r") {
        if (this.peek() === "\r") numCR += 1;
        this.advance();
      }
      next = this.offset;
      if (this.peek() === "\n" || this.peek() === "\r") {
        next += this.peek() === "\r" && this.peek(1) === "\n" ? 2 : 1;
      }
    } else {
      while (!this.isEOF()) {
        const char = this.peek();
        if (char === "\r") {
          numCR += 1;
        } else if (char === "\n" && nlOffset === 0) {
          nlOffset = this.offset;
        }
        if (char === "\n" || char === "\r") {
          this.advanceNewline();
        } else {
          this.advance();
        }
        if (char === "*" && this.peek() === "/") {
          this.advance();
          next = this.offset;
          break;
        }
      }
      if (next < 0) {
        this.errorAt("GOJR_SCAN006", "comment not terminated", offs);
      }
    }

    let lit = this.source.slice(offs, this.offset);
    if (numCR > 0 && lit.length >= 2 && lit[1] === "/" && lit.endsWith("\r")) {
      lit = lit.slice(0, -1);
      numCR -= 1;
    }
    if (next >= 0 && (lit[1] === "*" || offs === 0 || this.source[offs - 1] === "\n" || this.source[offs - 1] === "\r") && lit.slice(2).startsWith(prefix)) {
      this.updateLineInfo(next, offs, lit);
    }
    if (numCR > 0) {
      lit = stripCR(lit, lit[1] === "*");
    }
    return { text: lit, nlOffset };
  }

  private updateLineInfo(_next: number, offs: number, text: string): void {
    let body = text;
    if (body[1] === "*") {
      body = body.slice(0, -2);
    }
    body = body.slice(7);
    offs += 7;

    const trailing = trailingDigits(body);
    if (trailing.index === 0) {
      return;
    }
    if (!trailing.ok) {
      this.errorAt("GOJR_SCAN006", `invalid line number: ${body.slice(trailing.index)}`, offs + trailing.index);
      return;
    }

    const maxLineCol = 1 << 30;
    let line = trailing.value;
    let column = 0;
    const columnTrailing = trailingDigits(body.slice(0, trailing.index - 1));
    if (columnTrailing.ok) {
      line = columnTrailing.value;
      column = trailing.value;
      if (column === 0 || column > maxLineCol) {
        this.errorAt("GOJR_SCAN006", `invalid column number: ${body.slice(trailing.index)}`, offs + trailing.index);
        return;
      }
    }
    if (line === 0 || line > maxLineCol) {
      this.errorAt("GOJR_SCAN006", `invalid line number: ${body.slice(columnTrailing.ok ? columnTrailing.index : trailing.index)}`, offs + (columnTrailing.ok ? columnTrailing.index : trailing.index));
    }
  }

  private scanEscape(quote: string): boolean {
    const offs = this.offset;
    let n = 0;
    let base = 0;
    let max = 0;
    switch (this.peek()) {
      case "a":
      case "b":
      case "f":
      case "n":
      case "r":
      case "t":
      case "v":
      case "\\":
      case quote:
        this.advance();
        return true;
      case "0":
      case "1":
      case "2":
      case "3":
      case "4":
      case "5":
      case "6":
      case "7":
        n = 3;
        base = 8;
        max = 255;
        break;
      case "x":
        this.advance();
        n = 2;
        base = 16;
        max = 255;
        break;
      case "u":
        this.advance();
        n = 4;
        base = 16;
        max = 0x10ffff;
        break;
      case "U":
        this.advance();
        n = 8;
        base = 16;
        max = 0x10ffff;
        break;
      default:
        this.errorAt("GOJR_SCAN001", this.isEOF() ? "escape sequence not terminated" : "unknown escape sequence", offs);
        return false;
    }

    let value = 0;
    while (n > 0) {
      const digit = digitVal(this.peek());
      if (digit >= base) {
        this.errorAt("GOJR_SCAN001", this.isEOF() ? "escape sequence not terminated" : `illegal character ${JSON.stringify(this.peek())} in escape sequence`, this.offset);
        return false;
      }
      value = value * base + digit;
      this.advance();
      n -= 1;
    }

    if (value > max || (0xd800 <= value && value < 0xe000)) {
      this.errorAt("GOJR_SCAN001", "escape sequence is invalid Unicode code point", offs);
      return false;
    }
    return true;
  }

  private skipWhitespace(): void {
    while (this.peek() === " " || this.peek() === "\t" || (this.peek() === "\n" && !this.insertSemi) || this.peek() === "\r") {
      this.next();
    }
  }

  private switch2(tok0: TokenKind, tok1: TokenKind): TokenKind {
    if (this.peek() === "=") {
      this.advance();
      return tok1;
    }
    return tok0;
  }

  private switch3(tok0: TokenKind, tok1: TokenKind, ch2: string, tok2: TokenKind): TokenKind {
    if (this.peek() === "=") {
      this.advance();
      return tok1;
    }
    if (this.peek() === ch2) {
      this.advance();
      return tok2;
    }
    return tok0;
  }

  private switch4(tok0: TokenKind, tok1: TokenKind, ch2: string, tok2: TokenKind, tok3: TokenKind): TokenKind {
    if (this.peek() === "=") {
      this.advance();
      return tok1;
    }
    if (this.peek() === ch2) {
      this.advance();
      if (this.peek() === "=") {
        this.advance();
        return tok3;
      }
      return tok2;
    }
    return tok0;
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
      filename: start.filename,
      code,
      severity: "error",
      message,
      span: spanBetween(start, end)
    });
    this.err?.(start, message);
    this.ErrorCount += 1;
  }

  private errorf(offset: number, format: string, ...args: unknown[]): void {
    let index = 0;
    const message = format.replace(/%[qsdv]/g, (verb) => {
      const value = args[index++];
      if (verb === "%q") return JSON.stringify(value);
      return String(value);
    });
    this.errorAt("GOJR_SCAN002", message, offset);
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
      filename: this.filename,
      offset: this.offset,
      line: this.line,
      column: this.column
    };
  }

  private positionAt(offset: number): Position {
    let line = 1;
    let column = 1;
    for (let index = 0; index < offset && index < this.source.length;) {
      const char = this.source[index] ?? "";
      if (char === "\r") {
        if (this.source[index + 1] === "\n") index += 2;
        else index += 1;
        line += 1;
        column = 1;
        continue;
      }
      if (char === "\n") {
        index += 1;
        line += 1;
        column = 1;
        continue;
      }
      const codePoint = this.source.codePointAt(index);
      const text = codePoint === undefined ? "" : String.fromCodePoint(codePoint);
      index += Math.max(1, text.length);
      column += 1;
    }
    return {
      filename: this.filename,
      offset,
      line,
      column
    };
  }

  private errorAt(code: string, message: string, offset: number): void {
    this.error(code, message, this.positionAt(offset), this.positionAt(Math.min(this.source.length, offset + 1)));
  }

  private sliceFrom(start: Position): string {
    return this.source.slice(start.offset, this.offset);
  }

  private peek(ahead = 0): string {
    return this.source[this.offset + ahead] ?? "";
  }

  private currentChar(ahead = 0): string {
    const codePoint = this.source.codePointAt(this.offset + ahead);
    return codePoint === undefined ? "" : String.fromCodePoint(codePoint);
  }

  private isEOF(): boolean {
    return this.offset >= this.source.length;
  }

  private advance(): void {
    this.offset += 1;
    this.column += 1;
  }

  private advanceChar(): void {
    const char = this.currentChar();
    this.offset += Math.max(1, char.length);
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

  private digits(base: number, invalid: { offset: number } | undefined): number {
    let digsep = 0;
    if (base <= 10) {
      const max = String.fromCharCode("0".charCodeAt(0) + base);
      while (isDecimal(this.peek()) || this.peek() === "_") {
        let ds = 1;
        if (this.peek() === "_") {
          ds = 2;
        } else if (this.peek() >= max && invalid && invalid.offset < 0) {
          invalid.offset = this.offset;
        }
        digsep |= ds;
        this.advance();
      }
    } else {
      while (isHex(this.peek()) || this.peek() === "_") {
        const ds = this.peek() === "_" ? 2 : 1;
        digsep |= ds;
        this.advance();
      }
    }
    return digsep;
  }
}

function spanBetween(start: Position, end: Position): SourceSpan {
  return {
    filename: start.filename || end.filename || REPL_FILENAME,
    offset: start.offset,
    length: Math.max(0, end.offset - start.offset),
    line: start.line,
    column: start.column
  };
}

function twoCharToken(text: string): TokenKind | undefined {
  switch (text) {
    case ":=": return TokenKind.Define;
    case "+=": return TokenKind.PlusAssign;
    case "-=": return TokenKind.MinusAssign;
    case "*=": return TokenKind.StarAssign;
    case "/=": return TokenKind.SlashAssign;
    case "%=": return TokenKind.PercentAssign;
    case "&=": return TokenKind.AmpAssign;
    case "|=": return TokenKind.OrAssign;
    case "^=": return TokenKind.CaretAssign;
    case "<<": return TokenKind.Shl;
    case ">>": return TokenKind.Shr;
    case "&^": return TokenKind.BitClear;
    case "==": return TokenKind.Equal;
    case "!=": return TokenKind.NotEqual;
    case "<=": return TokenKind.LessEqual;
    case ">=": return TokenKind.GreaterEqual;
    case "&&": return TokenKind.AndAnd;
    case "||": return TokenKind.OrOr;
    case "->": return TokenKind.Arrow;
    case "++": return TokenKind.PlusPlus;
    case "--": return TokenKind.MinusMinus;
    case "<-": return TokenKind.Arrow;
    default: return undefined;
  }
}

function threeCharToken(text: string): TokenKind | undefined {
  switch (text) {
    case "<<=": return TokenKind.ShlAssign;
    case ">>=": return TokenKind.ShrAssign;
    case "&^=": return TokenKind.BitClearAssign;
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
    case "|": return TokenKind.Or;
    case "^": return TokenKind.Caret;
    case "~": return TokenKind.Tilde;
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
  return text === "_" || /^\p{L}$/u.test(text);
}

function isIdentifierPart(text: string): boolean {
  return text === "_" || /^\p{L}$/u.test(text) || /^\p{Nd}$/u.test(text);
}

function isDigit(text: string): boolean {
  return /^[0-9]$/.test(text);
}

function isLetter(text: string): boolean {
  const lowered = lower(text);
  return ("a" <= lowered && lowered <= "z") || text === "_" || /^\p{L}$/u.test(text);
}

function digitVal(text: string): number {
  if ("0" <= text && text <= "9") return text.charCodeAt(0) - "0".charCodeAt(0);
  const lowered = lower(text);
  if ("a" <= lowered && lowered <= "f") return lowered.charCodeAt(0) - "a".charCodeAt(0) + 10;
  return 16;
}

function lower(text: string): string {
  if (text.length === 0) return "";
  const code = text.charCodeAt(0);
  return code >= 65 && code <= 90 ? String.fromCharCode(code + 32) : text[0] ?? "";
}

function isDecimal(text: string): boolean {
  return text.length === 1 && "0" <= text && text <= "9";
}

function isHex(text: string): boolean {
  const lowered = lower(text);
  return isDecimal(text) || ("a" <= lowered && lowered <= "f");
}

function litname(prefix: string): string {
  switch (prefix) {
    case "x": return "hexadecimal literal";
    case "o":
    case "0": return "octal literal";
    case "b": return "binary literal";
    default: return "decimal literal";
  }
}

function invalidSep(text: string): number {
  let prefix = " ";
  let digit = ".";
  let index = 0;

  if (text.length >= 2 && text[0] === "0") {
    prefix = lower(text[1] ?? "");
    if (prefix === "x" || prefix === "o" || prefix === "b") {
      digit = "0";
      index = 2;
    }
  }

  for (; index < text.length; index += 1) {
    const previous = digit;
    digit = text[index] ?? "";
    if (digit === "_") {
      if (previous !== "0") return index;
    } else if (isDecimal(digit) || (prefix === "x" && isHex(digit))) {
      digit = "0";
    } else {
      if (previous === "_") return index - 1;
      digit = ".";
    }
  }
  if (digit === "_") return text.length - 1;
  return -1;
}

function stripCR(text: string, comment: boolean): string {
  let result = "";
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index] ?? "";
    if (char !== "\r" || (comment && result.length > "/*".length && result[result.length - 1] === "*" && index + 1 < text.length && text[index + 1] === "/")) {
      result += char;
    }
  }
  return result;
}

function trailingDigits(text: string): { index: number; value: number; ok: boolean } {
  const index = text.lastIndexOf(":");
  if (index < 0) {
    return { index: 0, value: 0, ok: false };
  }
  const digits = text.slice(index + 1);
  if (!/^[0-9]+$/.test(digits)) {
    return { index: index + 1, value: 0, ok: false };
  }
  return { index: index + 1, value: Number.parseInt(digits, 10), ok: true };
}

void init;
void eof;
void isLetter;
