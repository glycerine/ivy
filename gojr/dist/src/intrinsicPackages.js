export function isIntrinsicPackageImport(importPath) {
    return importPath === "runtime" ||
        importPath === "syscall/js" ||
        importPath === "unsafe" ||
        importPath === "internal/reflectlite";
}
export function isHostResolvedSourceImport(importPath) {
    return isIntrinsicPackageImport(importPath) ||
        importPath === "cmp" ||
        importPath === "fmt" ||
        importPath === "math" ||
        importPath === "testing";
}
