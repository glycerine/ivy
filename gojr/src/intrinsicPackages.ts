export function isIntrinsicPackageImport(importPath: string): boolean {
  return importPath === "runtime" || importPath === "unsafe";
}
