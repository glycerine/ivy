import { frontSourceToAst } from "./frontToAst.js";
import type { ProgramAst } from "./ast.js";
import type { Diagnostic } from "./diagnostics.js";
import type { ParseFrontResult } from "./front/parser.js";

export {
  AsyncGoJuniorSession
} from "./asyncSession.js";
export {
  emitAsyncJavaScript
} from "./asyncEmitter.js";
export type {
  AsyncEmitOptions,
  AsyncEmitResult
} from "./asyncEmitter.js";
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
  artifactPathForImportPath,
  buildPackage,
  buildPackages,
  resolveArtifactRoot
} from "./build.js";
export type {
  BuildArtifactReport,
  BuildArtifactStore,
  BuildExport,
  BuildPackageReport,
  BuildPackageRequest
} from "./build.js";
export type { Diagnostic, SourceFile, SourceSpan } from "./diagnostics.js";
export type { ImportDecl, ProgramAst, ProgramKind } from "./ast.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { checkFrontFiles, checkFrontSource } from "./front/checker.js";
export { parseFrontSource } from "./front/parser.js";
export { frontSourceToAst, frontToProgramAst } from "./frontToAst.js";
export { scanSource } from "./front/scanner.js";
export {
  ArrayType,
  assignableTo,
  BasicKind,
  BasicType,
  BuiltinObject,
  ChanType,
  ConstObject,
  FuncObject,
  implementsInterface,
  InterfaceType,
  isNilAssignable,
  MapType,
  methodSet,
  NamedType,
  newUniverse,
  ObjectKind,
  PackageInfo,
  PointerType,
  Scope,
  SignatureType,
  SliceType,
  StructType,
  tuple,
  TypeKind,
  TypeNameObject,
  VarObject,
  varOf
} from "./front/types.js";
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
export type { CheckConfig, CheckInfo, CheckResult, Importer, SheetNamespace, TypeAndValue, TypeMode } from "./front/checker.js";
export type { ParseFrontResult } from "./front/parser.js";
export type { FrontToken } from "./front/token.js";
export type { Type as FrontType, TypeObject } from "./front/types.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export {
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
  evaluateSourceFiles,
  testSource,
  testSourceFiles
} from "./runtime.js";
export type {
  EvaluationOptions,
  EvaluationResult,
  GoJuniorFunction,
  RuntimeCallable,
  RuntimeObject,
  RuntimeValue,
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
