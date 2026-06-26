import { frontSourceToAst } from "./frontToAst.js";
export { formatDiagnostic, hasErrorDiagnostics, diagnosticFilename, spanFromToken, REPL_FILENAME } from "./diagnostics.js";
export { analyzeEffects, functionEffectKey } from "./effects.js";
export { AsyncGoChannel, AsyncGoDeadlockError, AsyncGoPanic, AsyncGoScheduler, asyncSelect } from "./asyncRuntime.js";
export { blake3HashBytes, blake3HashString, blake3RawBytes } from "./blake3.js";
export { artifactPathForImportPath, buildPackage, buildPackages, buildStandardLibraryPackage, collectSourceImportPaths, createStandardLibrarySourcePackageProvider, GOJR_GOARCH, GOJR_GOOS, generatedCompiledPackageArtifactSource, inspectPackageJavaScript, generatedMixedWasmArtifactSource, parseGoJuniorPackageArchive, resolveArtifactRoot } from "./build.js";
export { GOJR_STAGE1_BACKEND, emitStage1Package } from "./emitter/package.js";
export { EmitterContext } from "./emitter/context.js";
export { createChunkedMixedWasmArtifactFixture, createMixedWasmArtifactFixture } from "./emitter/mixedArtifact.js";
export { CHECKED_IN_WASM_STENCILS, checkedInWasmStencil, decodeBase64, encodeBase64 } from "./emitter/stencils.js";
export { clangCandidates, compileCToWasm, discoverClangForWasm, probeClangSupportsWasm } from "./emitter/wasm/clang.js";
export { extractWasmFunction, parseWasmModule } from "./emitter/wasm/module.js";
export { buildPackagesOnNode, clearPackageArtifactCache, createNodeArtifactStore, createNodeSourcePackageProvider, defaultPackageCacheParent, evaluateSourceFilesWithPackagesOnNode, evaluateSourceWithPackagesOnNode, compileSourceFilesWithPackagesOnNode, inspectPackageJavaScriptOnNode, listPackageArtifactCache, loadSourcePackagesForRootFilesOnNode, packageCacheOnNode, runMainSourceFilesWithPackagesOnNode, runSpreadsheetFixtureWithPackagesOnNode, testSourceFilesWithPackagesOnNode } from "./nodeHost.js";
export { defaultGoJuniorBenchmarkCases, formatGoJuniorBenchmarkReport, runGoJuniorBenchmark } from "./bench.js";
export { formatGoJuniorWasmPocReport, runGoJuniorWasmPocBenchmark } from "./wasmPoc.js";
export { benchmarkGoJuniorOnNode, benchmarkGoJuniorOnNodeHostJSON, benchmarkGoJuniorOnNodeHostPayload, defaultCpuProfilePath, normalizeBenchmarkCacheMode, normalizeBenchmarkPhaseSet } from "./nodeBench.js";
export { compilePackageSourceFiles, compileSource, compileSourceFiles } from "./compile.js";
export { collectSpreadsheetFixtureFormulaSourceFiles, parseSpreadsheetFixtureJson, runSpreadsheetFixture, runSpreadsheetFixtureJson } from "./fixture.js";
export { buildReportToHostJSON, buildReportToHostPayload, compileResultToHostJSON, compileResultToHostPayload, evaluationResultToHostJSON, evaluationResultToHostPayload, formatBuildProgressEvent, formatFixtureSheets, hostFormatResult, hostFormatValue, hostResultValueIsNil, inspectPackageJavaScriptReportToHostJSON, inspectPackageJavaScriptReportToHostPayload, runtimeOptionsFromEnvironment, spreadsheetDiagnosticToHostString, spreadsheetFixtureResultToHostJSON, spreadsheetFixtureResultToHostPayload } from "./hostProtocol.js";
export { SpreadsheetEngine, SpreadsheetFormulaCompilerCache, cellDependency, rangeDependency, spreadsheetFormulaCacheKey, spreadsheetFormulaEvaluation, spreadsheetError } from "./spreadsheet.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { Codebase, CodebaseTxn, CodebaseUpdateTxn, CodebaseViewTxn } from "./codebase.js";
export { checkGoJuniorFiles, checkGoJuniorSource, checkGoJuniorSourceFiles } from "./typecheck.js";
export { parseFrontSource } from "./front/parser.js";
export { frontSourceToAst, frontToProgramAst } from "./frontToAst.js";
export { scanSource } from "./front/scanner.js";
export * as GoTypes from "./go/types/index.js";
export { TokenKind } from "./front/token.js";
export { Node as formatGoNode, Source as formatGoSource } from "./go/format.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { EvaluationContext, evaluateProgram, evaluateSource, formatReplValue, formatValue, GoJuniorSession, GoJuniorDeadlockError, GoJuniorPanic, GoJuniorRuntimeError, RuntimeMap, RuntimeChannel, RuntimeInterfaceValue, RuntimeNamedValue, RuntimePointer, RuntimeStruct, RuntimeTypedNilValue, gojrGeneratedRuntimeApi, evaluatePackageArtifact, evaluatePackageSourceFiles, evaluateSourcePackageGraph, evaluateSourceFiles, runLoadedMainPackage, runMainSourcePackageFiles, testSource, testSourceFiles, typeCheckConfig } from "./runtime.js";
export function parseProgram(source, filename) {
    const result = frontSourceToAst(source, filename);
    return {
        diagnostics: result.diagnostics,
        parsed: result.parsed,
        ...(result.ast ? { ast: result.ast } : {})
    };
}
