export const RESERVED_INTRINSIC_PACKAGE_IMPORTS = [
    "iter",
    "io/fs",
    "os",
    "runtime",
    "sync",
    "sync/atomic",
    "syscall/js",
    "unsafe",
    "weak",
    "internal/reflectlite"
];
export const GENERATED_ARTIFACT_INTRINSIC_IMPORTS = [
    ...RESERVED_INTRINSIC_PACKAGE_IMPORTS,
    "internal/bytealg",
    "internal/sync",
    "reflect"
];
const reservedIntrinsicPackageImports = new Set(RESERVED_INTRINSIC_PACKAGE_IMPORTS);
const generatedArtifactIntrinsicImports = new Set(GENERATED_ARTIFACT_INTRINSIC_IMPORTS);
export function isIntrinsicPackageImport(importPath) {
    return reservedIntrinsicPackageImports.has(importPath);
}
export function isGeneratedArtifactIntrinsicImport(importPath) {
    return generatedArtifactIntrinsicImports.has(importPath);
}
export function intrinsicPackageName(importPath) {
    switch (importPath) {
        case "internal/reflectlite":
            return "reflectlite";
        case "syscall/js":
            return "js";
        default:
            return isIntrinsicPackageImport(importPath) || importPath === "reflect"
                ? importPath.split("/").filter(Boolean).at(-1) ?? importPath
                : undefined;
    }
}
export function isHostResolvedSourceImport(importPath) {
    return isIntrinsicPackageImport(importPath) ||
        importPath === "testing";
}
