// Mechanical TypeScript transliteration support for go/types/typestring.go.

import type { Type } from "./type.js";

// Qualifier controls how package-level objects are qualified in strings.
export type Qualifier = (pkg: unknown) => string;

// TypeString returns the string representation of typ.
//
// This is a temporary mechanically named landing point for the full
// typestring.go port. It preserves the upstream API boundary while the detailed
// writer is transliterated file by file.
export function TypeString(typ: Type, _qf?: Qualifier | null): string {
  return typeStringFallback(typ);
}

export function WriteType(buf: string[], typ: Type | null, qf?: Qualifier | null): void {
  if (typ === null) {
    buf.push("<nil>");
    return;
  }
  buf.push(TypeString(typ, qf));
}

export function WriteSignature(buf: string[], sig: Type & { Params?: () => unknown; Results?: () => unknown; Variadic?: () => boolean }, qf?: Qualifier | null): void {
  const params = typeof sig.Params === "function" ? sig.Params() : null;
  const results = typeof sig.Results === "function" ? sig.Results() : null;
  buf.push("(");
  writeTuple(buf, params, qf);
  buf.push(")");
  if (results !== null && tupleLen(results) > 0) {
    buf.push(" ");
    if (tupleLen(results) === 1 && tupleVarName(results, 0) === "") {
      WriteType(buf, tupleVarType(results, 0), qf);
    } else {
      buf.push("(");
      writeTuple(buf, results, qf);
      buf.push(")");
    }
  }
}

function tupleLen(tuple: unknown): number {
  if (tuple !== null && typeof tuple === "object" && "Len" in tuple && typeof tuple.Len === "function") {
    return tuple.Len();
  }
  return 0;
}

function tupleAt(tuple: unknown, i: number): unknown {
  if (tuple !== null && typeof tuple === "object" && "At" in tuple && typeof tuple.At === "function") {
    return tuple.At(i);
  }
  return null;
}

function tupleVarName(tuple: unknown, i: number): string {
  const v = tupleAt(tuple, i) as { Name?: () => string; name?: string } | null;
  if (v === null) return "";
  if (typeof v.Name === "function") return v.Name();
  return v.name ?? "";
}

function tupleVarType(tuple: unknown, i: number): Type | null {
  const v = tupleAt(tuple, i) as { Type?: () => Type | null; typ?: Type | null } | null;
  if (v === null) return null;
  if (typeof v.Type === "function") return v.Type();
  return v.typ ?? null;
}

function writeTuple(buf: string[], tuple: unknown, qf?: Qualifier | null): void {
  for (let i = 0; i < tupleLen(tuple); i++) {
    if (i > 0) {
      buf.push(", ");
    }
    const name = tupleVarName(tuple, i);
    if (name !== "") {
      buf.push(name);
      buf.push(" ");
    }
    WriteType(buf, tupleVarType(tuple, i), qf);
  }
}

function typeStringFallback(typ: Type): string {
  const value = typ as unknown as { name?: string; kind?: unknown; elem?: Type; key?: Type; len?: number; base?: Type; dir?: unknown; fields?: unknown[] | null; params?: unknown; results?: unknown; obj?: { name?: string; Name?: () => string } };
  if (typeof value.name === "string") return value.name;
  if (value.obj) {
    if (typeof value.obj.Name === "function") return value.obj.Name();
    if (typeof value.obj.name === "string") return value.obj.name;
  }
  if ("len" in value && value.elem) return `[${String(value.len)}]${TypeString(value.elem)}`;
  if ("key" in value && value.key && value.elem) return `map[${TypeString(value.key)}]${TypeString(value.elem)}`;
  if ("base" in value && value.base) return `*${TypeString(value.base)}`;
  if ("elem" in value && value.elem) return `[]${TypeString(value.elem)}`;
  if ("fields" in value && Array.isArray(value.fields)) return "struct{...}";
  if ("params" in value || "results" in value) {
    const buf: string[] = ["func"];
    WriteSignature(buf, typ as Type & { Params?: () => unknown; Results?: () => unknown; Variadic?: () => boolean });
    return buf.join("");
  }
  return typ.constructor.name;
}
