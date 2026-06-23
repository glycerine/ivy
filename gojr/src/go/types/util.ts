// Mechanical TypeScript transliteration support for go/types/util.go fragments.

export function assert(condition: boolean, message = "assertion failed"): asserts condition {
  if (!condition) {
    throw new Error(message);
  }
}

export function unreachable(): never {
  throw new Error("unreachable");
}
