// Mechanical TypeScript transliteration support for go/types/util.go fragments.
import { TokenKind } from "../../front/token.js";
export function assert(condition, message = "assertion failed") {
    if (!condition) {
        throw new Error(message);
    }
}
export function unreachable() {
    throw new Error("unreachable");
}
export const isTypes2 = false;
// isdddArray reports whether atyp is of the form [...]E.
export function isdddArray(atyp) {
    const a = atyp;
    if (a !== null && a.kind === "ArrayType") {
        return a.inferredLength === true || (a.length?.kind === "Ellipsis" && a.length.element === undefined);
    }
    return false;
}
// makeFromLiteral returns the constant value for the given literal string and kind.
export function makeFromLiteral(lit, kind) {
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
    }
    catch {
        return { kind: "Unknown" };
    }
}
function parseGoIntLiteral(value) {
    const text = value.replace(/_/g, "");
    if (/^0[0-7]+$/.test(text)) {
        return BigInt(`0o${text.slice(1)}`);
    }
    return BigInt(text);
}
function parseGoFloatLiteral(value) {
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
function parseGoRuneLiteral(value) {
    const body = value.slice(1, -1);
    const decoded = decodeGoEscaped(body);
    return BigInt(Array.from(decoded)[0]?.codePointAt(0) ?? 0);
}
function parseGoStringLiteral(value) {
    if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
        return value.slice(1, -1).replace(/\r/g, "");
    }
    if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
        return decodeGoEscaped(value.slice(1, -1));
    }
    return value;
}
function decodeGoEscaped(value) {
    let out = "";
    for (let i = 0; i < value.length; i++) {
        const ch = value[i];
        if (ch !== "\\") {
            out += ch;
            continue;
        }
        i++;
        if (i >= value.length) {
            throw new Error("invalid escape");
        }
        const esc = value[i];
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
