import { EndOf, PosOf } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";

export function nodeKind(node: unknown): string {
  return (node as { kind?: string } | null)?.kind ?? "";
}

export function nodePos(node: unknown): number {
  return (node as { Pos?: () => number } | null)?.Pos?.() ?? PosOf(node as never);
}

export function nodeEnd(node: unknown): number {
  return (node as { End?: () => number } | null)?.End?.() ?? EndOf(node as never);
}

export function identName(id: unknown): string {
  const ident = id as { Name?: string; name?: string } | null;
  return ident?.Name ?? ident?.name ?? "";
}

export function fileName(file: unknown): unknown {
  const f = file as { Name?: unknown; name?: unknown } | null;
  return f?.Name ?? f?.name ?? null;
}

export function fileDecls(file: unknown): unknown[] {
  const f = file as { Decls?: unknown[]; declarations?: unknown[] } | null;
  return f?.Decls ?? f?.declarations ?? [];
}

export function fileGoVersion(file: unknown): string {
  return (file as { GoVersion?: string; goVersion?: string } | null)?.GoVersion ??
    (file as { GoVersion?: string; goVersion?: string } | null)?.goVersion ??
    "";
}

export function genDeclSpecs(decl: unknown): unknown[] {
  const d = decl as { Specs?: unknown[]; specs?: unknown[] } | null;
  return d?.Specs ?? d?.specs ?? [];
}

export function genDeclTok(decl: unknown): string {
  const tok = (decl as { Tok?: unknown; token?: unknown } | null)?.Tok ??
    (decl as { Tok?: unknown; token?: unknown } | null)?.token;
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

export function specNames(spec: unknown): unknown[] {
  const s = spec as { Names?: unknown[]; names?: unknown[] } | null;
  return s?.Names ?? s?.names ?? [];
}

export function specValues(spec: unknown): unknown[] {
  const s = spec as { Values?: unknown[]; values?: unknown[] } | null;
  return s?.Values ?? s?.values ?? [];
}

export function specType(spec: unknown): unknown {
  const s = spec as { Type?: unknown; type?: unknown } | null;
  return s?.Type ?? s?.type ?? null;
}

export function specName(spec: unknown): unknown {
  const s = spec as { Name?: unknown; name?: unknown } | null;
  return s?.Name ?? s?.name ?? null;
}

export function specPath(spec: unknown): unknown {
  const s = spec as { Path?: unknown; path?: unknown } | null;
  return s?.Path ?? s?.path ?? null;
}

export function basicLitValue(lit: unknown): string {
  const l = lit as { Value?: string; value?: string } | null;
  return l?.Value ?? l?.value ?? "";
}

export function funcDeclName(decl: unknown): unknown {
  const d = decl as { Name?: unknown; name?: unknown } | null;
  return d?.Name ?? d?.name ?? null;
}

export function funcDeclType(decl: unknown): unknown {
  const d = decl as { Type?: unknown; type?: unknown } | null;
  return d?.Type ?? d?.type ?? null;
}

export function funcDeclRecv(decl: unknown): unknown {
  const d = decl as { Recv?: unknown; receiver?: unknown } | null;
  return d?.Recv ?? d?.receiver ?? null;
}

export function funcDeclBody(decl: unknown): unknown {
  const d = decl as { Body?: unknown; body?: unknown } | null;
  return d?.Body ?? d?.body ?? null;
}

export function fieldListNumFields(list: unknown): number {
  const l = list as { NumFields?: () => number; List?: unknown[]; fields?: unknown[] } | null;
  return l?.NumFields?.() ?? l?.List?.length ?? l?.fields?.length ?? 0;
}
