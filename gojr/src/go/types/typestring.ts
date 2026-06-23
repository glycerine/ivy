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

function typeStringFallback(typ: Type): string {
  const value = typ as unknown as { name?: string; kind?: unknown; elem?: Type; key?: Type; len?: number; base?: Type; dir?: unknown };
  if (typeof value.name === "string") return value.name;
  if ("len" in value && value.elem) return `[${String(value.len)}]${TypeString(value.elem)}`;
  if ("key" in value && value.key && value.elem) return `map[${TypeString(value.key)}]${TypeString(value.elem)}`;
  if ("base" in value && value.base) return `*${TypeString(value.base)}`;
  if ("elem" in value && value.elem) return `[]${TypeString(value.elem)}`;
  return typ.constructor.name;
}
