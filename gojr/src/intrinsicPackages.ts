export function isIntrinsicPackageImport(importPath: string): boolean {
  return importPath === "iter" ||
    importPath === "runtime" ||
    importPath === "syscall/js" ||
    importPath === "unsafe" ||
    importPath === "internal/reflectlite";
}

export function isHostResolvedSourceImport(importPath: string): boolean {
  return isIntrinsicPackageImport(importPath) ||
    importPath === "testing";
}
