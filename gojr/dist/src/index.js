import { cstToAst } from "./cstToAst.js";
import { parseGoJunior } from "./parser.js";
export { cstToAst } from "./cstToAst.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { scanSource } from "./front/scanner.js";
export { ArrayType, assignableTo, BasicKind, BasicType, BuiltinObject, ConstObject, FuncObject, implementsInterface, InterfaceType, isNilAssignable, MapType, methodSet, NamedType, newUniverse, ObjectKind, PackageInfo, PointerType, Scope, SignatureType, SliceType, StructType, tuple, TypeKind, TypeNameObject, VarObject, varOf } from "./front/types.js";
export { TokenKind } from "./front/token.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { parseGoJunior } from "./parser.js";
export { evaluateProgram, evaluateSource, formatValue, GoJuniorSession, GoJuniorPanic, GoJuniorRuntimeError, RuntimeMap } from "./runtime.js";
export function parseProgram(source) {
    const result = parseGoJunior(source);
    return {
        ...result,
        ast: result.cst ? cstToAst(result.cst, result.diagnostics) : undefined
    };
}
