import { frontSourceToAst } from "./frontToAst.js";
import type { ProgramAst } from "./ast.js";
import type { Diagnostic } from "./diagnostics.js";
import type { ParseFrontResult } from "./front/parser.js";

export {
  formatDiagnostic,
  hasErrorDiagnostics,
  diagnosticFilename,
  spanFromToken,
  REPL_FILENAME
} from "./diagnostics.js";
export {
  analyzeEffects,
  functionEffectKey
} from "./effects.js";
export type {
  EffectReason,
  EffectReasonKind,
  FunctionEffectSummary,
  ProgramEffectAnalysis
} from "./effects.js";
export {
  AsyncGoChannel,
  AsyncGoDeadlockError,
  AsyncGoPanic,
  AsyncGoScheduler,
  asyncSelect
} from "./asyncRuntime.js";
export type {
  AsyncGoSchedulerOptions,
  AsyncSelectCase,
  AsyncSelectResult
} from "./asyncRuntime.js";
export {
  blake3HashBytes,
  blake3HashString,
  blake3RawBytes
} from "./blake3.js";
export {
  artifactPathForImportPath,
  buildPackage,
  buildPackages,
  buildStandardLibraryPackage,
  collectSourceImportPaths,
  createStandardLibrarySourcePackageProvider,
  GOJR_GOARCH,
  GOJR_GOOS,
  generatedCompiledPackageArtifactSource,
  inspectPackageJavaScript,
  generatedMixedWasmArtifactSource,
  parseGoJuniorPackageArchive,
  resolveArtifactRoot
} from "./build.js";
export {
  GOJR_STAGE1_BACKEND,
  emitStage1Package
} from "./emitter/package.js";
export type {
  Stage1PackageEmitResult
} from "./emitter/package.js";
export type {
  EmitterContextOptions
} from "./emitter/context.js";
export {
  EmitterContext
} from "./emitter/context.js";
export type {
  JsStencil,
  SlotKind,
  WasmFunctionSignature,
  WasmMemoryRequirement,
  WasmPatchHole,
  WasmPatchHoleKind,
  WasmStencil,
  WasmStencilMetadata
} from "./emitter/stencil.js";
export {
  createChunkedMixedWasmArtifactFixture,
  createMixedWasmArtifactFixture
} from "./emitter/mixedArtifact.js";
export type {
  MixedWasmArtifactFixture
} from "./emitter/mixedArtifact.js";
export {
  CHECKED_IN_WASM_STENCILS,
  checkedInWasmStencil,
  decodeBase64,
  encodeBase64
} from "./emitter/stencils.js";
export type {
  CheckedInWasmStencil,
  ResolvedWasmStencil
} from "./emitter/stencils.js";
export {
  clangCandidates,
  compileCToWasm,
  discoverClangForWasm,
  probeClangSupportsWasm
} from "./emitter/wasm/clang.js";
export type {
  ClangDiscoveryOptions,
  ClangDiscoveryResult,
  ClangProbeResult,
  CompileCToWasmOptions
} from "./emitter/wasm/clang.js";
export {
  extractWasmFunction,
  parseWasmModule
} from "./emitter/wasm/module.js";
export type {
  ExtractedWasmFunction,
  ExtractWasmFunctionOptions,
  WasmCodeBody,
  WasmCustomSection,
  WasmExport,
  WasmExportKind,
  WasmFunctionType,
  WasmImport,
  WasmImportKind,
  WasmLimits,
  WasmLocalDecl,
  WasmModuleInfo,
  WasmSection,
  WasmValueType
} from "./emitter/wasm/module.js";
export {
  buildPackagesOnNode,
  clearPackageArtifactCache,
  createNodeArtifactStore,
  createNodeSourcePackageProvider,
  defaultPackageCacheParent,
  evaluateSourceFilesWithPackagesOnNode,
  evaluateSourceWithPackagesOnNode,
  compileSourceFilesWithPackagesOnNode,
  inspectPackageJavaScriptOnNode,
  listPackageArtifactCache,
  loadSourcePackagesForRootFilesOnNode,
  packageCacheOnNode,
  runMainSourceFilesWithPackagesOnNode,
  runSpreadsheetFixtureWithPackagesOnNode,
  testSourceFilesWithPackagesOnNode
} from "./nodeHost.js";
export {
  defaultGoJuniorBenchmarkCases,
  formatGoJuniorBenchmarkReport,
  runGoJuniorBenchmark
} from "./bench.js";
export type {
  GoJuniorBenchmarkCacheMode,
  GoJuniorBenchmarkCase,
  GoJuniorBenchmarkCaseReport,
  GoJuniorBenchmarkMetrics,
  GoJuniorBenchmarkOptions,
  GoJuniorBenchmarkPhaseSet,
  GoJuniorBenchmarkPhaseSummary,
  GoJuniorBenchmarkReport
} from "./bench.js";
export {
  formatGoJuniorWasmPocReport,
  runGoJuniorWasmPocBenchmark
} from "./wasmPoc.js";
export type {
  GoJuniorWasmPocMetrics,
  GoJuniorWasmPocOptions,
  GoJuniorWasmPocReport
} from "./wasmPoc.js";
export {
  benchmarkGoJuniorOnNode,
  benchmarkGoJuniorOnNodeHostJSON,
  benchmarkGoJuniorOnNodeHostPayload,
  defaultCpuProfilePath,
  normalizeBenchmarkCacheMode,
  normalizeBenchmarkPhaseSet
} from "./nodeBench.js";
export type {
  NodeBenchmarkHostPayload,
  NodeBenchmarkRequest
} from "./nodeBench.js";
export {
  compilePackageSourceFiles,
  compileSource,
  compileSourceFiles
} from "./compile.js";
export type {
  CompilePackageOptions,
  CompilePackageResult,
  CompileResult
} from "./compile.js";
export {
  collectSpreadsheetFixtureFormulaSourceFiles,
  parseSpreadsheetFixtureJson,
  runSpreadsheetFixture,
  runSpreadsheetFixtureJson
} from "./fixture.js";
export {
  buildReportToHostJSON,
  buildReportToHostPayload,
  compileResultToHostJSON,
  compileResultToHostPayload,
  evaluationResultToHostJSON,
  evaluationResultToHostPayload,
  formatBuildProgressEvent,
  formatFixtureSheets,
  hostFormatResult,
  hostFormatValue,
  hostResultValueIsNil,
  inspectPackageJavaScriptReportToHostJSON,
  inspectPackageJavaScriptReportToHostPayload,
  runtimeOptionsFromEnvironment,
  spreadsheetDiagnosticToHostString,
  spreadsheetFixtureResultToHostJSON,
  spreadsheetFixtureResultToHostPayload
} from "./hostProtocol.js";
export type {
  HostBuildPayload,
  HostCompilePayload,
  HostEvaluationPayload,
  HostInspectJavaScriptPayload,
  HostRuntimeEnvironment,
  HostSpreadsheetFixturePayload
} from "./hostProtocol.js";
export type {
  SpreadsheetFixture,
  SpreadsheetFixtureCellInput,
  SpreadsheetFixtureCellSpec,
  SpreadsheetFixtureRunOptions,
  SpreadsheetFixtureRunResult
} from "./fixture.js";
export {
  SpreadsheetEngine,
  SpreadsheetFormulaCompilerCache,
  cellDependency,
  rangeDependency,
  spreadsheetFormulaCacheKey,
  spreadsheetFormulaEvaluation,
  spreadsheetError
} from "./spreadsheet.js";
export type {
  BuildArtifactReport,
  BuildArtifactStore,
  BuildExport,
  BuildProgressAction,
  BuildProgressEvent,
  BuildProgressSink,
  GoJuniorPackageArchive,
  GoJuniorPackageExportData,
  GoJuniorPackageExportIndexEntry,
  GoJuniorPackageSourcePayload,
  BuildPackageReport,
  BuildStandardLibraryPackageRequest,
  BuildSourcePackageProvider,
  InspectPackageJavaScriptReport,
  BuildPackageRequest,
  StandardLibrarySourceHost,
  StandardLibrarySourcePackageProviderOptions,
  SourceImportPathsResult
} from "./build.js";
export type {
  NodeBuildPackageRequest,
  PackageArtifactCacheEntry,
  PackageArtifactCacheRequest,
  PackageArtifactCacheResult
} from "./nodeHost.js";
export type {
  IterativeCalculationOptions,
  SetFormulaOptions,
  SetLiteralOptions,
  SpreadsheetFormulaCompileInput,
  SpreadsheetFormulaCompiler,
  SpreadsheetFormulaCompilerCacheInstallOptions,
  SpreadsheetFormulaCompilerCacheInstallResult,
  SpreadsheetFormulaCacheKeyInput,
  SpreadsheetCellDependency,
  SpreadsheetCellRef,
  SpreadsheetDependency,
  SpreadsheetDiagnostic,
  SpreadsheetEngineOptions,
  SpreadsheetErrorCode,
  SpreadsheetErrorValue,
  SpreadsheetFormulaContext,
  SpreadsheetFormulaEvaluation,
  SpreadsheetFormulaEvaluator,
  SpreadsheetRangeDependency,
  SpreadsheetRecalculationResult,
  SpreadsheetSetFormulaResult,
  SpreadsheetValue
} from "./spreadsheet.js";
export type { Diagnostic, SourceFile, SourceSpan } from "./diagnostics.js";
export type { ImportDecl, ProgramAst, ProgramKind } from "./ast.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export {
  checkGoJuniorFiles,
  checkGoJuniorSource,
  checkGoJuniorSourceFiles
} from "./typecheck.js";
export { parseFrontSource } from "./front/parser.js";
export { frontSourceToAst, frontToProgramAst } from "./frontToAst.js";
export { scanSource } from "./front/scanner.js";
export * as GoTypes from "./go/types/index.js";
export { TokenKind } from "./front/token.js";
export type {
  ArrayType as FrontArrayTypeNode,
  AstNode,
  BasicLit as FrontBasicLit,
  BinaryExpr as FrontBinaryExpr,
  BlockStmt as FrontBlockStmt,
  CallExpr as FrontCallExpr,
  CellAddress as FrontCellAddress,
  CellRefExpr as FrontCellRefExpr,
  Decl as FrontDecl,
  Expr as FrontExpr,
  Field as FrontField,
  FieldList as FrontFieldList,
  File as FrontFile,
  FuncDecl as FrontFuncDecl,
  FuncType as FrontFuncTypeNode,
  GenDecl as FrontGenDecl,
  Ident as FrontIdent,
  RangeRefExpr as FrontRangeRefExpr,
  Spec as FrontSpec,
  Stmt as FrontStmt
} from "./front/ast.js";
export type { ScanResult } from "./front/scanner.js";
export type {
  GoJuniorCheckConfig,
  GoJuniorCheckResult,
  GoJuniorImporter
} from "./typecheck.js";
export type { ParseFrontResult } from "./front/parser.js";
export type { FrontToken } from "./front/token.js";
export {
  Node as formatGoNode,
  Source as formatGoSource
} from "./go/format.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export {
  EvaluationContext,
  evaluateProgram,
  evaluateSource,
  formatReplValue,
  formatValue,
  GoJuniorSession,
  GoJuniorDeadlockError,
  GoJuniorPanic,
  GoJuniorRuntimeError,
  RuntimeMap,
  RuntimeChannel,
  RuntimeInterfaceValue,
  RuntimeNamedValue,
  RuntimePointer,
  RuntimeStruct,
  RuntimeTypedNilValue,
  gojrGeneratedRuntimeApi,
  evaluatePackageArtifact,
  evaluatePackageSourceFiles,
  evaluateSourcePackageGraph,
  evaluateSourceFiles,
  runLoadedMainPackage,
  runMainSourcePackageFiles,
  testSource,
  testSourceFiles,
  typeCheckConfig
} from "./runtime.js";
export type {
  EvaluationOptions,
  EvaluationResult,
  GoJuniorGeneratedPackageArtifact,
  GoJuniorGeneratedPackageContext,
  GoJuniorGeneratedPackageContextOptions,
  GoJuniorGeneratedRuntimeApi,
  GoJuniorFunction,
  PackageEvaluationOptions,
  PackageEvaluationResult,
  PackageRuntimePlan,
  PackageRuntimePlanConstant,
  PackageRuntimePlanVariable,
  MainPackageRunOptions,
  RuntimeCallable,
  RuntimeObject,
  RuntimeValue,
  SourcePackageGraphEvaluationResult,
  SourcePackageGraphOptions,
  SourcePackageProvider,
  SourcePackageSpec,
  SheetData
} from "./runtime.js";

export interface ParseProgramResult {
  diagnostics: Diagnostic[];
  parsed: ParseFrontResult;
  ast?: ProgramAst;
}

export function parseProgram(source: string, filename: string): ParseProgramResult {
  const result = frontSourceToAst(source, filename);
  return {
    diagnostics: result.diagnostics,
    parsed: result.parsed,
    ...(result.ast ? { ast: result.ast } : {})
  };
}
