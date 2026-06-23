import { cstToAst } from "./cstToAst.js";
import { parseGoJunior } from "./parser.js";

export type { Diagnostic, SourceSpan } from "./diagnostics.js";
export type { ImportDecl, ProgramAst, ProgramKind } from "./ast.js";
export { cstToAst } from "./cstToAst.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { checkFrontFiles, checkFrontSource } from "./front/checker.js";
export { parseFrontSource } from "./front/parser.js";
export { scanSource } from "./front/scanner.js";
export {
  ArrayType,
  assignableTo,
  BasicKind,
  BasicType,
  BuiltinObject,
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
export { parseGoJunior } from "./parser.js";
export {
  evaluateProgram,
  evaluateSource,
  formatValue,
  GoJuniorSession,
  GoJuniorPanic,
  GoJuniorRuntimeError,
  RuntimeMap
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

export function parseProgram(source: string) {
  const result = parseGoJunior(source);
  return {
    ...result,
    ast: result.cst ? cstToAst(result.cst, result.diagnostics) : undefined
  };
}
