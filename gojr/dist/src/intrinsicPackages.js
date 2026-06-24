export function isIntrinsicPackageImport(importPath) {
    return importPath === "runtime" || importPath === "unsafe";
}
export function isHostResolvedSourceImport(importPath) {
    return isIntrinsicPackageImport(importPath) ||
        importPath === "cmp" ||
        importPath === "fmt" ||
        importPath === "math" ||
        importPath === "testing";
}
