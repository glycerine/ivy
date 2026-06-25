import type { BuildPackageReport, BuildProgressEvent, InspectPackageJavaScriptReport } from "./build.js";
import type { CompileResult } from "./compile.js";
import { formatDiagnostic, hasErrorDiagnostics, type Diagnostic } from "./diagnostics.js";
import type { SpreadsheetFixtureRunResult } from "./fixture.js";
import { formatReplValue, type EvaluationOptions, type EvaluationProgressEvent, type EvaluationResult, type RuntimeValue } from "./runtime.js";
import type { SpreadsheetDiagnostic, SpreadsheetValue } from "./spreadsheet.js";

export interface HostEvaluationPayload {
  ok: boolean;
  incomplete: boolean;
  exitCode?: number;
  diagnostics: string[];
  output: string;
  value: string;
  valueIsNil: boolean;
  observedDeps: readonly unknown[];
}

export interface HostCompilePayload {
  ok: boolean;
  diagnostics: string[];
  output: string;
}

export interface HostBuildPayload {
  ok: boolean;
  incomplete: boolean;
  diagnostics: string[];
  output: string;
  artifacts: unknown[];
  built: string[];
  skipped: string[];
}

export interface HostInspectJavaScriptPayload extends HostBuildPayload {
  source: string;
}

export interface HostSpreadsheetFixturePayload {
  ok: boolean;
  unstable: boolean;
  diagnostics: string[];
  evaluated: readonly unknown[];
  sheets: Record<string, Record<string, string>>;
  observedDeps: Record<string, unknown>;
}

export interface HostRuntimeEnvironment {
  GOJR_RANDOM_SEED?: string;
}

export function runtimeOptionsFromEnvironment(
  env: HostRuntimeEnvironment | undefined,
  extra: EvaluationOptions = {}
): EvaluationOptions {
  const seed = env?.GOJR_RANDOM_SEED;
  return seed === undefined || seed === ""
    ? { ...extra }
    : { ...extra, randomSeed: seed };
}

export function hostFormatValue(value: unknown): string {
  if (typeof value === "function") {
    const name = value.name ? value.name.replace(/^_fn_/, "") : "";
    return name ? `<func ${name}>` : "<func>";
  }
  return formatReplValue(value as RuntimeValue);
}

export function hostFormatResult(result: Pick<EvaluationResult, "value" | "values">): string {
  if (result.values) return result.values.map((value) => hostFormatValue(value)).join(", ");
  if (Object.prototype.hasOwnProperty.call(result, "value")) return hostFormatValue(result.value);
  return "";
}

export function hostResultValueIsNil(result: Pick<EvaluationResult, "value" | "values">): boolean {
  if (Array.isArray(result.values)) return result.values.length === 1 && result.values[0] === null;
  if (Object.prototype.hasOwnProperty.call(result, "value")) return result.value === null;
  return false;
}

export function evaluationResultToHostPayload(
  result: EvaluationResult,
  extraOutput: readonly string[] = []
): HostEvaluationPayload {
  const diagnostics = result.diagnostics || [];
  const exitCode = result.exitCode ?? 0;
  return {
    ok: !hasErrorDiagnostics(diagnostics) && exitCode === 0,
    incomplete: result.incomplete === true,
    ...(result.exitCode !== undefined ? { exitCode: result.exitCode } : {}),
    diagnostics: diagnostics.map(formatDiagnostic),
    output: [...extraOutput, ...(result.output || [])].join(""),
    value: hostFormatResult(result),
    valueIsNil: hostResultValueIsNil(result),
    observedDeps: Array.isArray(result.observedDeps) ? result.observedDeps : []
  };
}

export function evaluationResultToHostJSON(result: EvaluationResult, extraOutput: readonly string[] = []): string {
  return JSON.stringify(evaluationResultToHostPayload(result, extraOutput));
}

export function compileResultToHostPayload(result: CompileResult): HostCompilePayload {
  const diagnostics = result.diagnostics || [];
  return {
    ok: !hasErrorDiagnostics(diagnostics),
    diagnostics: diagnostics.map(formatDiagnostic),
    output: ""
  };
}

export function compileResultToHostJSON(result: CompileResult): string {
  return JSON.stringify(compileResultToHostPayload(result));
}

export function buildReportToHostPayload(result: BuildPackageReport): HostBuildPayload {
  const diagnostics = result.diagnostics || [];
  return {
    ok: !hasErrorDiagnostics(diagnostics),
    incomplete: false,
    diagnostics: diagnostics.map(formatDiagnostic),
    output: "",
    artifacts: result.artifacts || [],
    built: result.built || [],
    skipped: result.skipped || []
  };
}

export function buildReportToHostJSON(result: BuildPackageReport): string {
  return JSON.stringify(buildReportToHostPayload(result));
}

export function formatBuildProgressEvent(event: BuildProgressEvent): string {
  const prefix = event.action === "checking"
    ? "checking"
    : event.action === "cached"
      ? "cached"
      : "built";
  const detail = event.artifactPath ? ` -> ${event.artifactPath}` : "";
  const stdlib = event.standardLibrary ? " stdlib" : "";
  const files = event.fileCount === undefined ? "" : ` files=${event.fileCount}`;
  const deps = event.dependencyCount === undefined ? "" : ` deps=${event.dependencyCount}`;
  return `gojr: ${prefix}${stdlib} ${event.importPath}${detail}${files}${deps}`;
}

export function formatEvaluationProgressEvent(event: EvaluationProgressEvent): string {
  const files = event.fileCount === undefined ? "" : ` files=${event.fileCount}`;
  const deps = event.dependencyCount === undefined ? "" : ` deps=${event.dependencyCount}`;
  return `gojr: ${event.action} ${event.importPath}${files}${deps}`;
}

export function inspectPackageJavaScriptReportToHostPayload(result: InspectPackageJavaScriptReport): HostInspectJavaScriptPayload {
  return {
    ...buildReportToHostPayload(result),
    source: result.source || ""
  };
}

export function inspectPackageJavaScriptReportToHostJSON(result: InspectPackageJavaScriptReport): string {
  return JSON.stringify(inspectPackageJavaScriptReportToHostPayload(result));
}

export function spreadsheetDiagnosticToHostString(diagnostic: SpreadsheetDiagnostic): string {
  const cell = diagnostic.cell
    ? `${diagnostic.cell.sheet}!${diagnostic.cell.cell}`
    : "spreadsheet";
  return `${cell}: error ${diagnostic.code}: ${diagnostic.message}`;
}

export function spreadsheetFixtureResultToHostPayload(
  result: SpreadsheetFixtureRunResult & { packageDiagnostics?: readonly Diagnostic[] }
): HostSpreadsheetFixturePayload {
  const packageDiagnostics = result.packageDiagnostics || [];
  if (hasErrorDiagnostics(packageDiagnostics)) {
    return {
      ok: false,
      unstable: false,
      diagnostics: packageDiagnostics.map(formatDiagnostic),
      evaluated: [],
      sheets: {},
      observedDeps: {}
    };
  }

  const diagnostics = result.diagnostics || [];
  return {
    ok: result.ok === true && diagnostics.length === 0,
    unstable: result.unstable === true,
    diagnostics: diagnostics.map(spreadsheetDiagnosticToHostString),
    evaluated: result.evaluated || [],
    sheets: formatFixtureSheets(result.sheets || {}),
    observedDeps: result.observedDeps || {}
  };
}

export function spreadsheetFixtureResultToHostJSON(
  result: SpreadsheetFixtureRunResult & { packageDiagnostics?: readonly Diagnostic[] }
): string {
  return JSON.stringify(spreadsheetFixtureResultToHostPayload(result));
}

export function formatFixtureSheets(sheets: Record<string, Record<string, SpreadsheetValue>>): Record<string, Record<string, string>> {
  const formatted: Record<string, Record<string, string>> = {};
  for (const [sheetName, cells] of Object.entries(sheets || {})) {
    formatted[sheetName] = {};
    for (const [cell, value] of Object.entries(cells || {})) {
      formatted[sheetName]![cell] = hostFormatValue(value);
    }
  }
  return formatted;
}
