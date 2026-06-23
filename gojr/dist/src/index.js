import { frontSourceToAst } from "./frontToAst.js";
export { childNodes, ident, parseCellAddress, walk } from "./front/ast.js";
export { checkFrontFiles, checkFrontSource } from "./front/checker.js";
export { parseFrontSource } from "./front/parser.js";
export { frontSourceToAst, frontToProgramAst } from "./frontToAst.js";
export { scanSource } from "./front/scanner.js";
export { ArrayType, assignableTo, BasicKind, BasicType, BuiltinObject, ConstObject, FuncObject, implementsInterface, InterfaceType, isNilAssignable, MapType, methodSet, NamedType, newUniverse, ObjectKind, PackageInfo, PointerType, Scope, SignatureType, SliceType, StructType, tuple, TypeKind, TypeNameObject, VarObject, varOf } from "./front/types.js";
export { TokenKind } from "./front/token.js";
export { parseRuntimeJson, parseSheetJson, parseSheetsJson } from "./jsonInput.js";
export { evaluateProgram, evaluateSource, formatReplValue, formatValue, GoJuniorSession, GoJuniorPanic, GoJuniorRuntimeError, RuntimeMap } from "./runtime.js";
export function parseProgram(source) {
    const result = frontSourceToAst(source);
    return {
        diagnostics: result.diagnostics,
        parsed: result.parsed,
        ...(result.ast ? { ast: result.ast } : {})
    };
}
