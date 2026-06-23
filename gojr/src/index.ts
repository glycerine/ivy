import { cstToAst } from "./cstToAst.js";
import { parseGoJunior } from "./parser.js";

export type { Diagnostic, SourceSpan } from "./diagnostics.js";
export type { ImportDecl, ProgramAst, ProgramKind } from "./ast.js";
export { cstToAst } from "./cstToAst.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { parseGoJunior } from "./parser.js";
export {
  evaluateProgram,
  evaluateSource,
  formatValue,
  GoJuniorPanic,
  GoJuniorRuntimeError
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
