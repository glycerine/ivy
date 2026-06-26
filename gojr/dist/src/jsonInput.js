export function parseSheetJson(json) {
    const value = parseRuntimeJson(json);
    if (!isRuntimeObject(value)) {
        throw new Error("--sheet-json must be a JSON object");
    }
    return value;
}
export function parseSheetsJson(json) {
    const value = parseRuntimeJson(json);
    if (!isRuntimeObject(value)) {
        throw new Error("--sheets-json must be a JSON object");
    }
    const sheets = {};
    for (const [sheetName, sheetValue] of Object.entries(value)) {
        if (!isRuntimeObject(sheetValue)) {
            throw new Error(`sheet ${sheetName} must be a JSON object`);
        }
        sheets[sheetName] = sheetValue;
    }
    return sheets;
}
export function parseRuntimeJson(json) {
    return new RuntimeJsonParser(json).parse();
}
class RuntimeJsonParser {
    source;
    offset = 0;
    constructor(source) {
        this.source = source;
    }
    parse() {
        this.skipWhitespace();
        const value = this.parseValue();
        this.skipWhitespace();
        if (!this.eof()) {
            this.fail("unexpected trailing input");
        }
        return value;
    }
    parseValue() {
        this.skipWhitespace();
        const char = this.peek();
        if (char === "{")
            return this.parseObject();
        if (char === "[")
            return this.parseArray();
        if (char === "\"")
            return this.parseString();
        if (char === "t")
            return this.parseKeyword("true", true);
        if (char === "f")
            return this.parseKeyword("false", false);
        if (char === "n")
            return this.parseKeyword("null", null);
        if (char === "-" || isDigit(char))
            return this.parseNumber();
        this.fail("expected JSON value");
    }
    parseObject() {
        const result = {};
        this.expect("{");
        this.skipWhitespace();
        if (this.consumeIf("}"))
            return result;
        while (true) {
            this.skipWhitespace();
            if (this.peek() !== "\"") {
                this.fail("expected object key");
            }
            const key = this.parseString();
            this.skipWhitespace();
            this.expect(":");
            result[key] = this.parseValue();
            this.skipWhitespace();
            if (this.consumeIf("}"))
                return result;
            this.expect(",");
        }
    }
    parseArray() {
        const result = [];
        this.expect("[");
        this.skipWhitespace();
        if (this.consumeIf("]"))
            return result;
        while (true) {
            result.push(this.parseValue());
            this.skipWhitespace();
            if (this.consumeIf("]"))
                return result;
            this.expect(",");
        }
    }
    parseString() {
        const start = this.offset;
        this.expect("\"");
        while (!this.eof()) {
            const char = this.advance();
            if (char === "\"") {
                const raw = this.source.slice(start, this.offset);
                return JSON.parse(raw);
            }
            if (char === "\\") {
                this.advanceEscape();
            }
            if (char !== undefined && char < " ") {
                this.fail("control character in string");
            }
        }
        this.fail("unterminated string");
    }
    advanceEscape() {
        const escaped = this.advance();
        if (!escaped)
            this.fail("unterminated escape");
        if ("\"\\/bfnrt".includes(escaped))
            return;
        if (escaped === "u") {
            for (let count = 0; count < 4; count += 1) {
                if (!isHexDigit(this.advance())) {
                    this.fail("invalid unicode escape");
                }
            }
            return;
        }
        this.fail("invalid escape");
    }
    parseKeyword(keyword, value) {
        if (this.source.slice(this.offset, this.offset + keyword.length) !== keyword) {
            this.fail(`expected ${keyword}`);
        }
        this.offset += keyword.length;
        return value;
    }
    parseNumber() {
        const start = this.offset;
        this.consumeIf("-");
        this.parseIntegerPart();
        let floating = false;
        if (this.consumeIf(".")) {
            floating = true;
            if (!isDigit(this.peek()))
                this.fail("expected digit after decimal point");
            while (isDigit(this.peek()))
                this.offset += 1;
        }
        if (this.peek() === "e" || this.peek() === "E") {
            floating = true;
            this.offset += 1;
            if (this.peek() === "+" || this.peek() === "-")
                this.offset += 1;
            if (!isDigit(this.peek()))
                this.fail("expected exponent digit");
            while (isDigit(this.peek()))
                this.offset += 1;
        }
        const raw = this.source.slice(start, this.offset);
        if (floating) {
            const value = Number(raw);
            if (!Number.isFinite(value))
                this.fail("number is outside float64 range");
            return value;
        }
        return BigInt(raw);
    }
    parseIntegerPart() {
        if (this.peek() === "0") {
            this.offset += 1;
            if (isDigit(this.peek()))
                this.fail("leading zero in number");
            return;
        }
        if (!isDigitOneToNine(this.peek())) {
            this.fail("expected digit");
        }
        while (isDigit(this.peek()))
            this.offset += 1;
    }
    skipWhitespace() {
        while (/\s/.test(this.peek() ?? ""))
            this.offset += 1;
    }
    expect(char) {
        if (!this.consumeIf(char)) {
            this.fail(`expected ${char}`);
        }
    }
    consumeIf(char) {
        if (this.peek() !== char)
            return false;
        this.offset += 1;
        return true;
    }
    advance() {
        if (this.eof())
            return undefined;
        const char = this.source[this.offset];
        this.offset += 1;
        return char;
    }
    peek() {
        return this.source[this.offset];
    }
    eof() {
        return this.offset >= this.source.length;
    }
    fail(message) {
        throw new Error(`${message} at JSON offset ${this.offset}`);
    }
}
function isRuntimeObject(value) {
    return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function isDigit(char) {
    return Boolean(char && char >= "0" && char <= "9");
}
function isDigitOneToNine(char) {
    return Boolean(char && char >= "1" && char <= "9");
}
function isHexDigit(char) {
    return Boolean(char && /[0-9a-fA-F]/.test(char));
}
