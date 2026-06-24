import { EndOf, PosOf } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";
export function nodeKind(node) {
    return node?.kind ?? "";
}
export function nodePos(node) {
    return node?.Pos?.() ?? PosOf(node);
}
export function nodeEnd(node) {
    return node?.End?.() ?? EndOf(node);
}
export function identName(id) {
    const ident = id;
    return ident?.Name ?? ident?.name ?? "";
}
export function fileName(file) {
    const f = file;
    return f?.Name ?? f?.name ?? null;
}
export function fileDecls(file) {
    const f = file;
    return f?.Decls ?? f?.declarations ?? [];
}
export function fileGoVersion(file) {
    return file?.GoVersion ??
        file?.goVersion ??
        "";
}
export function genDeclSpecs(decl) {
    const d = decl;
    return d?.Specs ?? d?.specs ?? [];
}
export function genDeclTok(decl) {
    const tok = decl?.Tok ??
        decl?.token;
    switch (tok) {
        case "CONST":
        case TokenKind.Const:
            return "CONST";
        case "VAR":
        case TokenKind.Var:
            return "VAR";
        case "TYPE":
        case TokenKind.Type:
            return "TYPE";
        case "IMPORT":
        case TokenKind.Import:
            return "IMPORT";
        default:
            return String(tok ?? "");
    }
}
export function specNames(spec) {
    const s = spec;
    return s?.Names ?? s?.names ?? [];
}
export function specValues(spec) {
    const s = spec;
    return s?.Values ?? s?.values ?? [];
}
export function specType(spec) {
    const s = spec;
    return s?.Type ?? s?.type ?? null;
}
export function specName(spec) {
    const s = spec;
    return s?.Name ?? s?.name ?? null;
}
export function specPath(spec) {
    const s = spec;
    return s?.Path ?? s?.path ?? null;
}
export function basicLitValue(lit) {
    const l = lit;
    return l?.Value ?? l?.value ?? "";
}
export function funcDeclName(decl) {
    const d = decl;
    return d?.Name ?? d?.name ?? null;
}
export function funcDeclType(decl) {
    const d = decl;
    return d?.Type ?? d?.type ?? null;
}
export function funcDeclRecv(decl) {
    const d = decl;
    return d?.Recv ?? d?.receiver ?? null;
}
export function funcDeclBody(decl) {
    const d = decl;
    return d?.Body ?? d?.body ?? null;
}
export function fieldListNumFields(list) {
    const l = list;
    return l?.NumFields?.() ?? l?.List?.length ?? l?.fields?.length ?? 0;
}
