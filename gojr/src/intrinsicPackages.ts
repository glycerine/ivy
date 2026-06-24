export function isIntrinsicPackageImport(importPath: string): boolean {
  return importPath === "runtime" || importPath === "unsafe";
}

export function isHostResolvedSourceImport(importPath: string): boolean {
  return isIntrinsicPackageImport(importPath) ||
    importPath === "cmp" ||
    importPath === "fmt" ||
    importPath === "math" ||
    importPath === "testing";
}
