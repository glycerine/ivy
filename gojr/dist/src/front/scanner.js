import { keywordKind, tokenCanEndStatement, TokenKind } from "./token.js";
export function scanSource(source) {
    return new Scanner(source).scan();
}
class Scanner {
    source;
    offset = 0;
    line = 1;
    column = 1;
    insertSemi = false;
    emittedEOF = false;
    tokens = [];
    diagnostics = [];
    constructor(source) {
        this.source = source;
    }
    scan() {
        while (!this.emittedEOF) {
            this.scanOne();
        }
        return {
            tokens: this.tokens,
            diagnostics: this.diagnostics
        };
    }
    scanOne() {
        const skipped = this.skipTrivia();
        if (skipped)
            return;
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
        if (char === "\"") {
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
    skipTrivia() {
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
    scanString(start) {
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
                if (!this.isEOF())
                    this.advance();
                continue;
            }
            this.advance();
        }
        this.error("GOJR_SCAN001", "unterminated string literal", start, this.position());
        this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
    }
    scanRawString(start) {
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
    scanRune(start) {
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
                if (!this.isEOF())
                    this.advance();
                continue;
            }
            this.advance();
        }
        this.error("GOJR_SCAN001", sawContent ? "unterminated rune literal" : "empty rune literal", start, this.position());
        this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
    }
    scanNumber(start) {
        const match = goNumberPattern.exec(this.source.slice(this.offset));
        if (!match) {
            this.advance();
            this.error("GOJR_SCAN002", "malformed numeric literal", start, this.position());
            this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
            return;
        }
        const text = match[0];
        this.advanceMany(text.length);
        const core = text.endsWith("i") ? text.slice(0, -1) : text;
        const kind = text.endsWith("i")
            ? TokenKind.ImagLiteral
            : /[.eEpP]/.test(core)
                ? TokenKind.FloatLiteral
                : TokenKind.IntLiteral;
        this.emit(kind, text, start, this.position());
    }
    scanIdentifierOrCell(start) {
        const cell = this.matchCellAddressAt(this.offset);
        if (cell) {
            this.advanceMany(cell.length);
            this.emit(TokenKind.CellAddress, cell, start, this.position());
            return;
        }
        if (this.peek() === "$") {
            this.advance();
            this.error("GOJR_SCAN003", "malformed spreadsheet cell address", start, this.position());
            this.emit(TokenKind.Illegal, this.sliceFrom(start), start, this.position());
            return;
        }
        this.advance();
        while (isIdentifierPart(this.peek()))
            this.advance();
        const text = this.sliceFrom(start);
        this.emit(keywordKind(text) ?? TokenKind.Identifier, text, start, this.position());
    }
    scanOperatorOrDelimiter(start) {
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
            if (two === "<-") {
                this.error("GOJR_SCAN004", "channels and channel operations are not supported", start, this.position());
            }
            else if (two === "->") {
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
    skipLineComment() {
        while (!this.isEOF() && this.peek() !== "\n" && this.peek() !== "\r") {
            this.advance();
        }
    }
    skipBlockComment(start) {
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
    emit(kind, lexeme, start, end, inserted) {
        const token = {
            kind,
            lexeme,
            span: spanBetween(start, end),
            ...(inserted ? { inserted } : {})
        };
        this.tokens.push(token);
        this.insertSemi = tokenCanEndStatement(kind);
    }
    emitSyntheticSemicolon(position, inserted) {
        this.emit(TokenKind.Semicolon, ";", position, position, inserted);
    }
    error(code, message, start, end) {
        this.diagnostics.push({
            code,
            severity: "error",
            message,
            span: spanBetween(start, end)
        });
    }
    matchCellAddressAt(offset) {
        const rest = this.source.slice(offset);
        const match = /^\$?[A-Z]{1,3}\$?[1-9][0-9]*/.exec(rest);
        if (!match)
            return undefined;
        const text = match[0];
        const next = rest[text.length] ?? "";
        return isIdentifierPart(next) || next === "$" ? undefined : text;
    }
    position() {
        return {
            offset: this.offset,
            line: this.line,
            column: this.column
        };
    }
    sliceFrom(start) {
        return this.source.slice(start.offset, this.offset);
    }
    peek(ahead = 0) {
        return this.source[this.offset + ahead] ?? "";
    }
    isEOF() {
        return this.offset >= this.source.length;
    }
    advance() {
        this.offset += 1;
        this.column += 1;
    }
    advanceMany(count) {
        for (let index = 0; index < count; index += 1)
            this.advance();
    }
    advanceNewline() {
        if (this.peek() === "\r" && this.peek(1) === "\n") {
            this.offset += 2;
        }
        else {
            this.offset += 1;
        }
        this.line += 1;
        this.column = 1;
    }
}
function spanBetween(start, end) {
    return {
        offset: start.offset,
        length: Math.max(0, end.offset - start.offset),
        line: start.line,
        column: start.column
    };
}
function twoCharToken(text) {
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
        case "<-": return TokenKind.Illegal;
        default: return undefined;
    }
}
function threeCharToken(text) {
    switch (text) {
        case "<<=": return TokenKind.ShlAssign;
        case ">>=": return TokenKind.ShrAssign;
        case "&^=": return TokenKind.BitClearAssign;
        default: return undefined;
    }
}
function oneCharToken(text) {
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
const decimalDigits = String.raw `(?:[0-9](?:_?[0-9])*)`;
const binaryDigits = String.raw `(?:[01](?:_?[01])*)`;
const octalDigits = String.raw `(?:[0-7](?:_?[0-7])*)`;
const hexDigits = String.raw `(?:[0-9A-Fa-f](?:_?[0-9A-Fa-f])*)`;
const decimalFloat = String.raw `(?:(?:${decimalDigits}\.${decimalDigits}?|${decimalDigits}\.|\.(?:${decimalDigits}))(?:[eE][+-]?${decimalDigits})?|${decimalDigits}[eE][+-]?${decimalDigits})`;
const hexMantissa = String.raw `(?:${hexDigits}(?:\.${hexDigits}?)?|\.${hexDigits})`;
const hexFloat = String.raw `(?:0[xX]${hexMantissa}[pP][+-]?${decimalDigits})`;
const integer = String.raw `(?:0[bB]${binaryDigits}|0[oO]${octalDigits}|0[xX]${hexDigits}|${decimalDigits})`;
const goNumberPattern = new RegExp(`^(?:${hexFloat}|${decimalFloat}|${integer})(?:i)?`);
function isIdentifierStart(text) {
    return /^[A-Za-z_]$/.test(text);
}
function isIdentifierPart(text) {
    return /^[A-Za-z0-9_]$/.test(text);
}
function isDigit(text) {
    return /^[0-9]$/.test(text);
}
