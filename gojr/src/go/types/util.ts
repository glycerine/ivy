// Mechanical TypeScript transliteration support for go/types/util.go fragments.

import { EndOf, PosOf, type AstNode, type CallExpr } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";
import { atPos, type positioner } from "./check.js";

export function assert(condition: boolean, message = "assertion failed"): asserts condition {
  if (!condition) {
    throw new Error(message);
  }
}

export function unreachable(): never {
  throw new Error("unreachable");
}

export const isTypes2 = false;

// hasDots reports whether the last argument in the call is followed by ...
export function hasDots(call: CallExpr): boolean {
  return call.ellipsis;
}

// dddErrPos returns the positioner for reporting an invalid ... use in a call.
export function dddErrPos(call: CallExpr): positioner {
  const last = call.args[call.args.length - 1];
  return new atPos(last !== undefined ? EndOf(last) : EndOf(call));
}

// isdddArray reports whether atyp is of the form [...]E.
export function isdddArray(atyp: unknown): boolean {
  const a = atyp as { kind?: string; inferredLength?: boolean; length?: unknown } | null;
  if (a !== null && a.kind === "ArrayType") {
    return a.inferredLength === true || ((a.length as { kind?: string; element?: unknown } | undefined)?.kind === "Ellipsis" && (a.length as { element?: unknown }).element === undefined);
  }
  return false;
}

// argErrPos returns positioner for reporting an invalid argument count.
export function argErrPos(call: CallExpr): positioner {
  return new atPos(EndOf(call));
}

// startPos returns the start position of node n.
export function startPos(n: AstNode): number {
  return PosOf(n);
}

// endPos returns the position of the first character immediately after node n.
export function endPos(n: AstNode): number {
  return EndOf(n);
}

// makeFromLiteral returns the constant value for the given literal string and kind.
export function makeFromLiteral(lit: string, kind: TokenKind): unknown {
  try {
    switch (kind) {
      case TokenKind.IntLiteral:
        return parseGoIntLiteral(lit);
      case TokenKind.FloatLiteral:
        return parseGoFloatLiteral(lit);
      case TokenKind.ImagLiteral: {
        const raw = lit.endsWith("i") ? lit.slice(0, -1) : lit;
        const im = /[.eEpP]/.test(raw) ? parseGoFloatLiteral(raw) : Number(parseGoIntLiteral(raw));
        return { re: 0, im };
      }
      case TokenKind.RuneLiteral:
        return parseGoRuneLiteral(lit);
      case TokenKind.StringLiteral:
        return parseGoStringLiteral(lit);
      default:
        return { kind: "Unknown" };
    }
  } catch {
    return { kind: "Unknown" };
  }
}

function parseGoIntLiteral(value: string): bigint {
  const text = value.replace(/_/g, "");
  if (/^0[0-7]+$/.test(text)) {
    return BigInt(`0o${text.slice(1)}`);
  }
  return BigInt(text);
}

function parseGoFloatLiteral(value: string): number {
  const text = value.replace(/_/g, "");
  const hex = /^0[xX]([0-9a-fA-F]*)(?:\.([0-9a-fA-F]*))?[pP]([+-]?[0-9]+)$/.exec(text);
  if (hex === null) {
    return Number(text);
  }
  const whole = hex[1] || "0";
  const frac = hex[2] || "";
  const exponent = Number(hex[3]);
  const wholeValue = Number.parseInt(whole, 16);
  let fracValue = 0;
  for (let index = 0; index < frac.length; index++) {
    fracValue += Number.parseInt(frac[index] ?? "0", 16) / 16 ** (index + 1);
  }
  return (wholeValue + fracValue) * 2 ** exponent;
}

function parseGoRuneLiteral(value: string): bigint {
  const body = value.slice(1, -1);
  const decoded = decodeGoEscaped(body);
  return BigInt(Array.from(decoded)[0]?.codePointAt(0) ?? 0);
}

function parseGoStringLiteral(value: string): string {
  if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
    return value.slice(1, -1).replace(/\r/g, "");
  }
  if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
    return decodeGoEscaped(value.slice(1, -1));
  }
  return value;
}

function decodeGoEscaped(value: string): string {
  let out = "";
  for (let i = 0; i < value.length; i++) {
    const ch = value[i]!;
    if (ch !== "\\") {
      out += ch;
      continue;
    }
    i++;
    if (i >= value.length) {
      throw new Error("invalid escape");
    }
    const esc = value[i]!;
    switch (esc) {
      case "a":
        out += "\x07";
        break;
      case "b":
        out += "\b";
        break;
      case "f":
        out += "\f";
        break;
      case "n":
        out += "\n";
        break;
      case "r":
        out += "\r";
        break;
      case "t":
        out += "\t";
        break;
      case "v":
        out += "\v";
        break;
      case "\\":
      case "\"":
      case "'":
        out += esc;
        break;
      case "x": {
        const hex = value.slice(i + 1, i + 3);
        if (!/^[0-9a-fA-F]{2}$/.test(hex)) {
          throw new Error("invalid hex escape");
        }
        out += String.fromCodePoint(Number.parseInt(hex, 16));
        i += 2;
        break;
      }
      case "u": {
        const hex = value.slice(i + 1, i + 5);
        if (!/^[0-9a-fA-F]{4}$/.test(hex)) {
          throw new Error("invalid unicode escape");
        }
        out += String.fromCodePoint(Number.parseInt(hex, 16));
        i += 4;
        break;
      }
      case "U": {
        const hex = value.slice(i + 1, i + 9);
        if (!/^[0-9a-fA-F]{8}$/.test(hex)) {
          throw new Error("invalid unicode escape");
        }
        out += String.fromCodePoint(Number.parseInt(hex, 16));
        i += 8;
        break;
      }
      default:
        if (/[0-7]/.test(esc)) {
          const oct = value.slice(i, i + 3);
          if (!/^[0-7]{3}$/.test(oct)) {
            throw new Error("invalid octal escape");
          }
          out += String.fromCodePoint(Number.parseInt(oct, 8));
          i += 2;
          break;
        }
        throw new Error("invalid escape");
    }
  }
  return out;
}
