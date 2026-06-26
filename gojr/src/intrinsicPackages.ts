export const RESERVED_INTRINSIC_PACKAGE_IMPORTS = [
  "iter",
  "runtime",
  "syscall/js",
  "unsafe",
  "internal/reflectlite"
] as const;

export const GENERATED_ARTIFACT_INTRINSIC_IMPORTS = [
  ...RESERVED_INTRINSIC_PACKAGE_IMPORTS,
  "reflect"
] as const;

const reservedIntrinsicPackageImports = new Set<string>(RESERVED_INTRINSIC_PACKAGE_IMPORTS);
const generatedArtifactIntrinsicImports = new Set<string>(GENERATED_ARTIFACT_INTRINSIC_IMPORTS);

export function isIntrinsicPackageImport(importPath: string): boolean {
  return reservedIntrinsicPackageImports.has(importPath);
}

export function isGeneratedArtifactIntrinsicImport(importPath: string): boolean {
  return generatedArtifactIntrinsicImports.has(importPath);
}

export function intrinsicPackageName(importPath: string): string | undefined {
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

export function isHostResolvedSourceImport(importPath: string): boolean {
  return isIntrinsicPackageImport(importPath) ||
    importPath === "testing";
}
