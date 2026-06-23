import { RuntimeObject, RuntimeValue, SheetData } from "./runtime.js";

export function parseSheetJson(json: string): SheetData {
  const value = parseRuntimeJson(json);
  if (!isRuntimeObject(value)) {
    throw new Error("--sheet-json must be a JSON object");
  }
  return value as SheetData;
}

export function parseSheetsJson(json: string): Record<string, SheetData> {
  const value = parseRuntimeJson(json);
  if (!isRuntimeObject(value)) {
    throw new Error("--sheets-json must be a JSON object");
  }

  const sheets: Record<string, SheetData> = {};
  for (const [sheetName, sheetValue] of Object.entries(value)) {
    if (!isRuntimeObject(sheetValue)) {
      throw new Error(`sheet ${sheetName} must be a JSON object`);
    }
    sheets[sheetName] = sheetValue as SheetData;
  }
  return sheets;
}

export function parseRuntimeJson(json: string): RuntimeValue {
  return new RuntimeJsonParser(json).parse();
}

class RuntimeJsonParser {
  private offset = 0;

  public constructor(private readonly source: string) {}

  public parse(): RuntimeValue {
    this.skipWhitespace();
    const value = this.parseValue();
    this.skipWhitespace();
    if (!this.eof()) {
      this.fail("unexpected trailing input");
    }
    return value;
  }

  private parseValue(): RuntimeValue {
    this.skipWhitespace();
    const char = this.peek();
    if (char === "{") return this.parseObject();
    if (char === "[") return this.parseArray();
    if (char === "\"") return this.parseString();
    if (char === "t") return this.parseKeyword("true", true);
    if (char === "f") return this.parseKeyword("false", false);
    if (char === "n") return this.parseKeyword("null", null);
    if (char === "-" || isDigit(char)) return this.parseNumber();
    this.fail("expected JSON value");
  }

  private parseObject(): RuntimeObject {
    const result: RuntimeObject = {};
    this.expect("{");
    this.skipWhitespace();
    if (this.consumeIf("}")) return result;

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
      if (this.consumeIf("}")) return result;
      this.expect(",");
    }
  }

  private parseArray(): RuntimeValue[] {
    const result: RuntimeValue[] = [];
    this.expect("[");
    this.skipWhitespace();
    if (this.consumeIf("]")) return result;

    while (true) {
      result.push(this.parseValue());
      this.skipWhitespace();
      if (this.consumeIf("]")) return result;
      this.expect(",");
    }
  }

  private parseString(): string {
    const start = this.offset;
    this.expect("\"");
    while (!this.eof()) {
      const char = this.advance();
      if (char === "\"") {
        const raw = this.source.slice(start, this.offset);
        return JSON.parse(raw) as string;
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

  private advanceEscape(): void {
    const escaped = this.advance();
    if (!escaped) this.fail("unterminated escape");
    if ("\"\\/bfnrt".includes(escaped)) return;
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

  private parseKeyword<T extends RuntimeValue>(keyword: string, value: T): T {
    if (this.source.slice(this.offset, this.offset + keyword.length) !== keyword) {
      this.fail(`expected ${keyword}`);
    }
    this.offset += keyword.length;
    return value;
  }

  private parseNumber(): RuntimeValue {
    const start = this.offset;
    this.consumeIf("-");
    this.parseIntegerPart();

    let floating = false;
    if (this.consumeIf(".")) {
      floating = true;
      if (!isDigit(this.peek())) this.fail("expected digit after decimal point");
      while (isDigit(this.peek())) this.offset += 1;
    }

    if (this.peek() === "e" || this.peek() === "E") {
      floating = true;
      this.offset += 1;
      if (this.peek() === "+" || this.peek() === "-") this.offset += 1;
      if (!isDigit(this.peek())) this.fail("expected exponent digit");
      while (isDigit(this.peek())) this.offset += 1;
    }

    const raw = this.source.slice(start, this.offset);
    if (floating) {
      const value = Number(raw);
      if (!Number.isFinite(value)) this.fail("number is outside float64 range");
      return value;
    }
    return BigInt(raw);
  }

  private parseIntegerPart(): void {
    if (this.peek() === "0") {
      this.offset += 1;
      if (isDigit(this.peek())) this.fail("leading zero in number");
      return;
    }
    if (!isDigitOneToNine(this.peek())) {
      this.fail("expected digit");
    }
    while (isDigit(this.peek())) this.offset += 1;
  }

  private skipWhitespace(): void {
    while (/\s/.test(this.peek() ?? "")) this.offset += 1;
  }

  private expect(char: string): void {
    if (!this.consumeIf(char)) {
      this.fail(`expected ${char}`);
    }
  }

  private consumeIf(char: string): boolean {
    if (this.peek() !== char) return false;
    this.offset += 1;
    return true;
  }

  private advance(): string | undefined {
    if (this.eof()) return undefined;
    const char = this.source[this.offset];
    this.offset += 1;
    return char;
  }

  private peek(): string | undefined {
    return this.source[this.offset];
  }

  private eof(): boolean {
    return this.offset >= this.source.length;
  }

  private fail(message: string): never {
    throw new Error(`${message} at JSON offset ${this.offset}`);
  }
}

function isRuntimeObject(value: RuntimeValue): value is RuntimeObject {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function isDigit(char: string | undefined): boolean {
  return Boolean(char && char >= "0" && char <= "9");
}

function isDigitOneToNine(char: string | undefined): boolean {
  return Boolean(char && char >= "1" && char <= "9");
}

function isHexDigit(char: string | undefined): boolean {
  return Boolean(char && /[0-9a-fA-F]/.test(char));
}
