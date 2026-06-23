import { REPL_FILENAME } from "../diagnostics.js";
import { keywordKind, tokenCanEndStatement, TokenKind } from "./token.js";
export function scanSource(source, filename) {
    return new Scanner(source, filename).scan();
}
class Scanner {
    source;
    filename;
    offset = 0;
    line = 1;
    column = 1;
    insertSemi = false;
    emittedEOF = false;
    tokens = [];
    diagnostics = [];
    constructor(source, filename) {
        this.source = source;
        this.filename = filename;
    }
    scan() {
        if (this.offset === 0 && this.currentChar() === "\uFEFF")
            this.advanceChar();
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
            }
            else if (exponent === "p" && prefix !== "x") {
                this.errorAt("GOJR_SCAN002", `'${this.peek()}' exponent requires hexadecimal mantissa`, this.offset);
            }
            this.advance();
            kind = TokenKind.FloatLiteral;
            if (this.peek() === "+" || this.peek() === "-")
                this.advance();
            const exponentDigits = this.digits(10, undefined);
            digsep |= exponentDigits;
            if ((exponentDigits & 1) === 0) {
                this.errorAt("GOJR_SCAN002", "exponent has no digits", this.offset);
            }
        }
        else if (prefix === "x" && kind === TokenKind.FloatLiteral) {
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
    scanIdentifierOrCell(start) {
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
        this.advanceChar();
        while (isIdentifierPart(this.currentChar()))
            this.advanceChar();
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
            filename: start.filename,
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
            filename: this.filename,
            offset: this.offset,
            line: this.line,
            column: this.column
        };
    }
    positionAt(offset) {
        let line = 1;
        let column = 1;
        for (let index = 0; index < offset && index < this.source.length;) {
            const char = this.source[index] ?? "";
            if (char === "\r") {
                if (this.source[index + 1] === "\n")
                    index += 2;
                else
                    index += 1;
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
    errorAt(code, message, offset) {
        this.error(code, message, this.positionAt(offset), this.positionAt(Math.min(this.source.length, offset + 1)));
    }
    sliceFrom(start) {
        return this.source.slice(start.offset, this.offset);
    }
    peek(ahead = 0) {
        return this.source[this.offset + ahead] ?? "";
    }
    currentChar(ahead = 0) {
        const codePoint = this.source.codePointAt(this.offset + ahead);
        return codePoint === undefined ? "" : String.fromCodePoint(codePoint);
    }
    isEOF() {
        return this.offset >= this.source.length;
    }
    advance() {
        this.offset += 1;
        this.column += 1;
    }
    advanceChar() {
        const char = this.currentChar();
        this.offset += Math.max(1, char.length);
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
    digits(base, invalid) {
        let digsep = 0;
        if (base <= 10) {
            const max = String.fromCharCode("0".charCodeAt(0) + base);
            while (isDecimal(this.peek()) || this.peek() === "_") {
                let ds = 1;
                if (this.peek() === "_") {
                    ds = 2;
                }
                else if (this.peek() >= max && invalid && invalid.offset < 0) {
                    invalid.offset = this.offset;
                }
                digsep |= ds;
                this.advance();
            }
        }
        else {
            while (isHex(this.peek()) || this.peek() === "_") {
                const ds = this.peek() === "_" ? 2 : 1;
                digsep |= ds;
                this.advance();
            }
        }
        return digsep;
    }
}
function spanBetween(start, end) {
    return {
        filename: start.filename || end.filename || REPL_FILENAME,
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
        case "<-": return TokenKind.Arrow;
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
function isIdentifierStart(text) {
    return text === "_" || /^\p{L}$/u.test(text);
}
function isIdentifierPart(text) {
    return text === "_" || /^\p{L}$/u.test(text) || /^\p{Nd}$/u.test(text);
}
function isDigit(text) {
    return /^[0-9]$/.test(text);
}
function digitVal(text) {
    if ("0" <= text && text <= "9")
        return text.charCodeAt(0) - "0".charCodeAt(0);
    const lowered = lower(text);
    if ("a" <= lowered && lowered <= "f")
        return lowered.charCodeAt(0) - "a".charCodeAt(0) + 10;
    return 16;
}
function lower(text) {
    if (text.length === 0)
        return "";
    const code = text.charCodeAt(0);
    return code >= 65 && code <= 90 ? String.fromCharCode(code + 32) : text[0] ?? "";
}
function isDecimal(text) {
    return text.length === 1 && "0" <= text && text <= "9";
}
function isHex(text) {
    const lowered = lower(text);
    return isDecimal(text) || ("a" <= lowered && lowered <= "f");
}
function litname(prefix) {
    switch (prefix) {
        case "x": return "hexadecimal literal";
        case "o":
        case "0": return "octal literal";
        case "b": return "binary literal";
        default: return "decimal literal";
    }
}
function invalidSep(text) {
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
            if (previous !== "0")
                return index;
        }
        else if (isDecimal(digit) || (prefix === "x" && isHex(digit))) {
            digit = "0";
        }
        else {
            if (previous === "_")
                return index - 1;
            digit = ".";
        }
    }
    if (digit === "_")
        return text.length - 1;
    return -1;
}
