export function isIntrinsicPackageImport(importPath) {
    return importPath === "iter" ||
        importPath === "runtime" ||
        importPath === "syscall/js" ||
        importPath === "unsafe" ||
        importPath === "internal/reflectlite";
}
export function isHostResolvedSourceImport(importPath) {
    return isIntrinsicPackageImport(importPath) ||
        importPath === "testing";
}
