export function isIntrinsicPackageImport(importPath) {
    return importPath === "runtime" || importPath === "unsafe";
}
