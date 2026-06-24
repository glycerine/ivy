// Mechanical TypeScript support for Go's go/token.Pos values.

export type Pos = number;

export const NoPos: Pos = 0;

export function posIsValid(pos: Pos): boolean {
  return pos !== NoPos;
}
