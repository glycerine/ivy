import { formatDiagnostic, hasErrorDiagnostics } from "./diagnostics.js";
import { formatReplValue } from "./runtime.js";
export function runtimeOptionsFromEnvironment(env, extra = {}) {
    const seed = env?.GOJR_RANDOM_SEED;
    return seed === undefined || seed === ""
        ? { ...extra }
        : { ...extra, randomSeed: seed };
}
export function hostFormatValue(value) {
    if (typeof value === "function") {
        const name = value.name ? value.name.replace(/^_fn_/, "") : "";
        return name ? `<func ${name}>` : "<func>";
    }
    return formatReplValue(value);
}
export function hostFormatResult(result) {
    if (result.values)
        return result.values.map((value) => hostFormatValue(value)).join(", ");
    if (Object.prototype.hasOwnProperty.call(result, "value"))
        return hostFormatValue(result.value);
    return "";
}
export function hostResultValueIsNil(result) {
    if (Array.isArray(result.values))
        return result.values.length === 1 && result.values[0] === null;
    if (Object.prototype.hasOwnProperty.call(result, "value"))
        return result.value === null;
    return false;
}
export function evaluationResultToHostPayload(result, extraOutput = []) {
    const diagnostics = result.diagnostics || [];
    return {
        ok: !hasErrorDiagnostics(diagnostics),
        incomplete: result.incomplete === true,
        diagnostics: diagnostics.map(formatDiagnostic),
        output: [...extraOutput, ...(result.output || [])].join(""),
        value: hostFormatResult(result),
        valueIsNil: hostResultValueIsNil(result),
        observedDeps: Array.isArray(result.observedDeps) ? result.observedDeps : []
    };
}
export function evaluationResultToHostJSON(result, extraOutput = []) {
    return JSON.stringify(evaluationResultToHostPayload(result, extraOutput));
}
export function compileResultToHostPayload(result) {
    const diagnostics = result.diagnostics || [];
    return {
        ok: !hasErrorDiagnostics(diagnostics),
        diagnostics: diagnostics.map(formatDiagnostic),
        output: ""
    };
}
export function compileResultToHostJSON(result) {
    return JSON.stringify(compileResultToHostPayload(result));
}
export function buildReportToHostPayload(result) {
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
export function buildReportToHostJSON(result) {
    return JSON.stringify(buildReportToHostPayload(result));
}
export function inspectPackageJavaScriptReportToHostPayload(result) {
    return {
        ...buildReportToHostPayload(result),
        source: result.source || ""
    };
}
export function inspectPackageJavaScriptReportToHostJSON(result) {
    return JSON.stringify(inspectPackageJavaScriptReportToHostPayload(result));
}
export function spreadsheetDiagnosticToHostString(diagnostic) {
    const cell = diagnostic.cell
        ? `${diagnostic.cell.sheet}!${diagnostic.cell.cell}`
        : "spreadsheet";
    return `${cell}: error ${diagnostic.code}: ${diagnostic.message}`;
}
export function spreadsheetFixtureResultToHostPayload(result) {
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
export function spreadsheetFixtureResultToHostJSON(result) {
    return JSON.stringify(spreadsheetFixtureResultToHostPayload(result));
}
export function formatFixtureSheets(sheets) {
    const formatted = {};
    for (const [sheetName, cells] of Object.entries(sheets || {})) {
        formatted[sheetName] = {};
        for (const [cell, value] of Object.entries(cells || {})) {
            formatted[sheetName][cell] = hostFormatValue(value);
        }
    }
    return formatted;
}
