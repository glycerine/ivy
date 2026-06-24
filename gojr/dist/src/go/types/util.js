// Mechanical TypeScript transliteration support for go/types/util.go fragments.
export function assert(condition, message = "assertion failed") {
    if (!condition) {
        throw new Error(message);
    }
}
export function unreachable() {
    throw new Error("unreachable");
}
