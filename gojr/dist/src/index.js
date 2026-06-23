import { cstToAst } from "./cstToAst.js";
import { parseGoJunior } from "./parser.js";
export { cstToAst } from "./cstToAst.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { parseGoJunior } from "./parser.js";
export { evaluateProgram, evaluateSource, formatValue, GoJuniorSession, GoJuniorPanic, GoJuniorRuntimeError } from "./runtime.js";
export function parseProgram(source) {
    const result = parseGoJunior(source);
    return {
        ...result,
        ast: result.cst ? cstToAst(result.cst, result.diagnostics) : undefined
    };
}
