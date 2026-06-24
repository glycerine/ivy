// Mechanical TypeScript support for Go's go/token.Pos values.
export const NoPos = 0;
export function posIsValid(pos) {
    return pos !== NoPos;
}
