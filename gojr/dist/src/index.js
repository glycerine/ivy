import { frontSourceToAst } from "./frontToAst.js";
export { analyzeEffects, functionEffectKey } from "./effects.js";
export { AsyncGoChannel, AsyncGoDeadlockError, AsyncGoPanic, AsyncGoScheduler, asyncSelect } from "./asyncRuntime.js";
export { artifactPathForImportPath, buildPackage, buildPackages, buildStandardLibraryPackage, collectSourceImportPaths, createStandardLibrarySourcePackageProvider, GOJR_GOARCH, GOJR_GOOS, inspectPackageJavaScript, parseGoJuniorPackageArchive, resolveArtifactRoot } from "./build.js";
export { buildPackagesOnNode, clearPackageArtifactCache, createNodeArtifactStore, createNodeSourcePackageProvider, defaultPackageCacheParent, inspectPackageJavaScriptOnNode, listPackageArtifactCache, packageCacheOnNode } from "./nodeHost.js";
export { compilePackageSourceFiles, compileSource, compileSourceFiles } from "./compile.js";
export { collectSpreadsheetFixtureFormulaSourceFiles, parseSpreadsheetFixtureJson, runSpreadsheetFixture, runSpreadsheetFixtureJson } from "./fixture.js";
export { SpreadsheetEngine, SpreadsheetFormulaCompilerCache, cellDependency, rangeDependency, spreadsheetFormulaCacheKey, spreadsheetFormulaEvaluation, spreadsheetError } from "./spreadsheet.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { checkGoJuniorFiles, checkGoJuniorSource, checkGoJuniorSourceFiles } from "./typecheck.js";
export { parseFrontSource } from "./front/parser.js";
export { frontSourceToAst, frontToProgramAst } from "./frontToAst.js";
export { scanSource } from "./front/scanner.js";
export * as GoTypes from "./go/types/index.js";
export { TokenKind } from "./front/token.js";
export { Node as formatGoNode, Source as formatGoSource } from "./go/format.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { evaluateProgram, evaluateSource, formatReplValue, formatValue, GoJuniorSession, GoJuniorDeadlockError, GoJuniorPanic, GoJuniorRuntimeError, RuntimeMap, RuntimeChannel, evaluatePackageSourceFiles, evaluateSourcePackageGraph, evaluateSourceFiles, testSource, testSourceFiles, typeCheckConfig } from "./runtime.js";
export function parseProgram(source, filename) {
    const result = frontSourceToAst(source, filename);
    return {
        diagnostics: result.diagnostics,
        parsed: result.parsed,
        ...(result.ast ? { ast: result.ast } : {})
    };
}
