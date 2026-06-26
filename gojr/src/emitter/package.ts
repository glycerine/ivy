import type {
  ArrayLiteralExpression,
  BinaryExpression,
  BlockStatement,
  ConstDeclStatement,
  DeclarationSpec,
  Expression,
  FunctionDecl,
  LiteralExpression,
  MapLiteralExpression,
  ProgramAst,
  ReturnStatement,
  Statement,
  VarDeclStatement
} from "../ast.js";
import type { GoJuniorPackageExportData } from "../build.js";
import type { Diagnostic } from "../diagnostics.js";
import { EmitterContext } from "./context.js";
import { checkedInWasmStencil } from "./stencils.js";

export const GOJR_STAGE1_BACKEND = "copy-patch-wasm-stage1";

export interface Stage1PackageEmitResult {
  diagnostics: Diagnostic[];
  javascript: string;
  wasmBase64?: string;
}

export function emitStage1Package(artifact: GoJuniorPackageExportData, ast: ProgramAst): Stage1PackageEmitResult {
  const ctx = new EmitterContext({ artifact });
  const wasmFunctions = ast.functions.filter(isI64AddFunction);
  const usesWasm = wasmFunctions.length > 0;
  const wasmBase64 = usesWasm ? checkedInWasmStencil("i64.add.kernel").wasmBase64 : undefined;
  const bodyLines: string[] = [];

  for (const statement of ast.body) {
    if (statement.kind === "ConstDecl") {
      bodyLines.push(...emitConstDecl(ctx, statement));
      continue;
    }
    if (statement.kind === "VarDecl") {
      bodyLines.push(...emitVarDecl(ctx, statement));
      continue;
    }
    if (statement.kind === "TypeDecl") continue;
    ctx.emitError(`unsupported Stage 1 top-level statement ${statement.kind}`);
  }

  for (const fn of ast.functions) {
    bodyLines.push(...emitFunction(ctx, fn));
  }

  const javascript = stage1JavaScript(artifact, usesWasm, bodyLines);
  return {
    diagnostics: ctx.diagnostics,
    javascript,
    ...(wasmBase64 ? { wasmBase64 } : {})
  };
}

function emitConstDecl(ctx: EmitterContext, statement: ConstDeclStatement): string[] {
  return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "const", declaration));
}

function emitVarDecl(ctx: EmitterContext, statement: VarDeclStatement): string[] {
  return statement.declarations.flatMap((declaration) => emitDeclaration(ctx, "var", declaration));
}

function emitDeclaration(ctx: EmitterContext, _kind: "const" | "var", declaration: DeclarationSpec): string[] {
  const value = declaration.value
    ? expressionToJs(ctx, declaration.value)
    : zeroValueForType(declaration.type?.text);
  if (!value) {
    ctx.emitError(`unsupported Stage 1 declaration for ${declaration.name}`);
    return [];
  }
  return [`  pkg[${JSON.stringify(declaration.name)}] = ${value};`];
}

function emitFunction(ctx: EmitterContext, fn: FunctionDecl): string[] {
  if (isI64AddFunction(fn)) {
    return [
      `  pkg[${JSON.stringify(fn.name)}] = async (a, b) => __gojrWasmExports.add_i64(BigInt(a), BigInt(b));`
    ];
  }

  const statements = fn.body.statements;
  if (statements.length === 0) {
    return [`  pkg[${JSON.stringify(fn.name)}] = async () => null;`];
  }
  if (statements.length === 1 && statements[0]?.kind === "ReturnStatement") {
    const returned = returnLiteralExpression(ctx, statements[0]);
    if (returned) return [`  pkg[${JSON.stringify(fn.name)}] = async () => ${returned};`];
  }

  ctx.emitError(`unsupported Stage 1 function body for ${fn.name}`);
  return [];
}

function returnLiteralExpression(ctx: EmitterContext, statement: ReturnStatement): string | undefined {
  if (statement.values.length === 0) return "null";
  if (statement.values.length !== 1) return undefined;
  return expressionToJs(ctx, statement.values[0]);
}

function isI64AddFunction(fn: FunctionDecl): boolean {
  if (fn.signature.parameters.length !== 2 || fn.signature.results.length !== 1) return false;
  if (!fn.signature.parameters.every((param) => param.type.text === "int64")) return false;
  if (fn.signature.results[0]?.type.text !== "int64") return false;
  const statement = onlyReturn(fn.body);
  if (!statement || statement.values.length !== 1) return false;
  const value = statement.values[0];
  if (!value || value.kind !== "BinaryExpression" || value.operator !== "+") return false;
  return isIdentifier(value.left, fn.signature.parameters[0]?.name) &&
    isIdentifier(value.right, fn.signature.parameters[1]?.name);
}

function onlyReturn(body: BlockStatement): ReturnStatement | undefined {
  return body.statements.length === 1 && body.statements[0]?.kind === "ReturnStatement"
    ? body.statements[0]
    : undefined;
}

function isIdentifier(expression: Expression, name: string | undefined): boolean {
  return Boolean(name && expression.kind === "Identifier" && expression.name === name);
}

function expressionToJs(ctx: EmitterContext, expression: Expression | undefined): string | undefined {
  if (!expression) return undefined;
  switch (expression.kind) {
    case "Literal":
      return literalToJs(expression);
    case "ArrayLiteralExpression":
      return arrayLiteralToJs(ctx, expression);
    case "MapLiteralExpression":
      return mapLiteralToJs(ctx, expression);
    default:
      ctx.emitError(`unsupported Stage 1 expression ${expression.kind}`);
      return undefined;
  }
}

function literalToJs(expression: LiteralExpression): string | undefined {
  switch (expression.literalKind) {
    case "int":
    case "rune":
      if (typeof expression.value === "bigint") return `${expression.value.toString()}n`;
      if (typeof expression.value === "number") return `${Math.trunc(expression.value)}n`;
      return undefined;
    case "float":
      return typeof expression.value === "number" ? JSON.stringify(expression.value) : undefined;
    case "string":
      return typeof expression.value === "string" ? JSON.stringify(expression.value) : undefined;
    case "bool":
      return typeof expression.value === "boolean" ? String(expression.value) : undefined;
    case "nil":
      return "null";
    default:
      return undefined;
  }
}

function arrayLiteralToJs(ctx: EmitterContext, expression: ArrayLiteralExpression): string | undefined {
  const values = expression.elements.map((element) => expressionToJs(ctx, element.value));
  if (values.some((value) => value === undefined)) return undefined;
  const rendered = values.filter((value): value is string => value !== undefined);
  if (isByteSliceType(expression.type.text)) {
    const bytes = expression.elements.map((element) => byteLiteralValue(element.value));
    if (bytes.some((value) => value === undefined)) {
      ctx.emitError(`unsupported Stage 1 []byte literal element`);
      return undefined;
    }
    return `new Uint8Array([${bytes.join(", ")}])`;
  }
  return `[${rendered.join(", ")}]`;
}

function mapLiteralToJs(ctx: EmitterContext, expression: MapLiteralExpression): string | undefined {
  if (expression.keyType.text !== "string" || !isByteSliceType(expression.valueType.text)) {
    ctx.emitError(`unsupported Stage 1 map literal map[${expression.keyType.text}]${expression.valueType.text}`);
    return undefined;
  }
  const entries: string[] = [];
  for (const entry of expression.entries) {
    const key = expressionToJs(ctx, entry.key);
    const value = expressionToJs(ctx, entry.value);
    if (!key || !value) return undefined;
    entries.push(`[${key}, ${value}]`);
  }
  return `new Map([${entries.join(", ")}])`;
}

function byteLiteralValue(expression: Expression): number | undefined {
  if (expression.kind !== "Literal") return undefined;
  if (typeof expression.value === "bigint") {
    const value = Number(expression.value);
    return Number.isInteger(value) && value >= 0 && value <= 255 ? value : undefined;
  }
  if (typeof expression.value === "number") {
    const value = Math.trunc(expression.value);
    return value >= 0 && value <= 255 ? value : undefined;
  }
  return undefined;
}

function zeroValueForType(typeText: string | undefined): string | undefined {
  if (!typeText) return "null";
  switch (typeText) {
    case "int":
    case "int8":
    case "int16":
    case "int32":
    case "int64":
    case "uint":
    case "uint8":
    case "uint16":
    case "uint32":
    case "uint64":
    case "uintptr":
      return "0n";
    case "float32":
    case "float64":
      return "0";
    case "string":
      return "\"\"";
    case "bool":
      return "false";
    default:
      if (typeText.startsWith("[]")) return "[]";
      if (typeText.startsWith("map[")) return "new Map()";
      return "null";
  }
}

function isByteSliceType(typeText: string): boolean {
  return typeText === "[]byte" || typeText === "[]uint8";
}

function stage1JavaScript(artifact: GoJuniorPackageExportData, usesWasm: boolean, bodyLines: string[]): string {
  const artifactHeader: GoJuniorPackageExportData = { ...artifact };
  delete artifactHeader.runtime;
  return [
    "// Code generated by gojr Stage 1 copy-and-patch emitter; DO NOT EDIT.",
    `export const gojrPackageArtifact = ${JSON.stringify(artifactHeader, null, 2)};`,
    "export async function instantiateGoJrPackage(runtime = {}, options = {}) {",
    "  const pkg = Object.create(null);",
    "  const importsByPath = options.importsByPath || runtime.importsByPath || {};",
    "  void importsByPath;",
    usesWasm ? "  const wasmBase64 = options.wasmBase64 || runtime.wasmBase64;" : "",
    usesWasm ? "  if (typeof wasmBase64 !== \"string\" || wasmBase64.length === 0) throw new Error(\"gojr Stage 1 package requires options.wasmBase64 from _gojr.wasm\");" : "",
    usesWasm ? "  const __gojrWasmModule = new WebAssembly.Module(__gojrDecodeBase64(wasmBase64));" : "",
    usesWasm ? "  const __gojrWasmInstance = new WebAssembly.Instance(__gojrWasmModule, {});" : "",
    usesWasm ? "  const __gojrWasmExports = __gojrWasmInstance.exports;" : "  const __gojrWasmExports = {};",
    ...bodyLines,
    "  return { diagnostics: [], output: [], package: pkg, wasm: __gojrWasmExports };",
    "}",
    usesWasm ? "function __gojrDecodeBase64(base64) {" : "",
    usesWasm ? "  if (typeof Buffer !== \"undefined\") return new Uint8Array(Buffer.from(base64, \"base64\"));" : "",
    usesWasm ? "  const binary = atob(base64);" : "",
    usesWasm ? "  const out = new Uint8Array(binary.length);" : "",
    usesWasm ? "  for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);" : "",
    usesWasm ? "  return out;" : "",
    usesWasm ? "}" : "",
    "export default { artifact: gojrPackageArtifact, instantiateGoJrPackage };",
    ""
  ].filter((line) => line !== "").join("\n");
}
