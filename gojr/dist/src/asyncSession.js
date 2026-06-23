import { REPL_FILENAME } from "./diagnostics.js";
import { emitAsyncJavaScript } from "./asyncEmitter.js";
import { AsyncGoChannel, AsyncGoScheduler, asyncSelect } from "./asyncRuntime.js";
import { checkFrontSourceFiles } from "./front/checker.js";
import { newUniverse } from "./front/types.js";
import { frontSourceToAst } from "./frontToAst.js";
export class AsyncGoJuniorSession {
    options;
    acceptedSources = [];
    scheduler;
    state = Object.create(null);
    sheet;
    sheets;
    constructor(options = {}) {
        this.options = options;
        this.scheduler = new AsyncGoScheduler(options.randomSeed === undefined ? {} : { randomSeed: options.randomSeed });
        this.sheet = options.sheet ?? {};
        this.sheets = options.sheets ?? {};
    }
    setSheet(sheet) {
        this.sheet = sheet;
    }
    setSheets(sheets) {
        this.sheets = sheets;
    }
    async evaluate(source) {
        const sourceFile = sourceFileFromSource(source, this.options.filename ?? REPL_FILENAME);
        const parsed = frontSourceToAst(sourceFile.source, sourceFile.filename);
        const ast = parsed.ast;
        const hasParseError = parsed.diagnostics.some((diagnostic) => diagnostic.severity === "error");
        if (hasParseError) {
            return {
                diagnostics: parsed.diagnostics,
                output: [],
                incomplete: diagnosticsLookIncomplete(parsed.diagnostics),
                ...(ast ? { ast } : {})
            };
        }
        if (!ast)
            return { diagnostics: parsed.diagnostics, output: [] };
        const typeDiagnostics = this.checkSource(sourceFile);
        if (typeDiagnostics.some((diagnostic) => diagnostic.severity === "error")) {
            return { diagnostics: typeDiagnostics, output: [], ast };
        }
        const emitted = emitAsyncJavaScript(ast, {
            sessionState: true,
            returnLastExpression: true
        });
        if (emitted.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
            return { diagnostics: [...ast.diagnostics, ...emitted.diagnostics], output: [], ast };
        }
        try {
            const module = new Function(emitted.source)(this.runtime());
            const value = ast.kind === "function" && ast.functions[0] && ast.body.length === 0
                ? module.functions?.[ast.functions[0].name]
                : await module.main();
            this.acceptedSources.push(ensureTrailingNewlineSourceFile(sourceFile));
            return {
                diagnostics: ast.diagnostics,
                output: [],
                ast,
                ...(value !== undefined ? valueResult(value) : {})
            };
        }
        catch (error) {
            return {
                diagnostics: [
                    ...ast.diagnostics,
                    runtimeDiagnostic(ast, error instanceof Error ? error.message : String(error))
                ],
                output: [],
                ast
            };
        }
    }
    runtime() {
        return {
            scheduler: this.scheduler,
            AsyncGoChannel,
            asyncSelect,
            state: this.state
        };
    }
    checkSource(source) {
        const checked = checkFrontSourceFiles([...this.acceptedSources, ensureTrailingNewlineSourceFile(source)], typeCheckConfig({
            ...this.options,
            sheet: this.sheet,
            sheets: this.sheets
        }));
        return checked.diagnostics;
    }
}
function valueResult(value) {
    if (Array.isArray(value)) {
        return {
            values: value,
            ...(value.length === 1 ? { value: value[0] } : {})
        };
    }
    return { value: value };
}
function sourceFileFromSource(source, filename) {
    return { filename, source };
}
function ensureTrailingNewline(source) {
    return source.endsWith("\n") ? source : `${source}\n`;
}
function ensureTrailingNewlineSourceFile(file) {
    return {
        filename: file.filename,
        source: ensureTrailingNewline(file.source)
    };
}
function diagnosticsLookIncomplete(diagnostics) {
    const errors = diagnostics.filter((diagnostic) => diagnostic.severity === "error");
    return errors.length > 0 && errors.every((diagnostic) => {
        if (diagnostic.code === "GOJR_SCAN001") {
            return /unterminated .*string literal/i.test(diagnostic.message);
        }
        if (diagnostic.code !== "GOJR_PARSE_FRONT001")
            return false;
        return diagnostic.span?.length === 0 &&
            (/expected/i.test(diagnostic.message) || /found EOF/i.test(diagnostic.message));
    });
}
function runtimeDiagnostic(ast, message) {
    const span = ast.body?.[0]?.span ?? ast.functions?.[0]?.span;
    return {
        filename: span?.filename ?? REPL_FILENAME,
        code: "GOJR_RUNTIME001",
        severity: "error",
        message,
        ...(span ? { span } : {})
    };
}
function typeCheckConfig(options) {
    const universe = newUniverse();
    return {
        universe,
        sheetNamespaces: sheetNamespacesForOptions(options, universe)
    };
}
function sheetNamespacesForOptions(options, universe) {
    const namespaces = {
        sheet: sheetNamespaceForData(options.sheet ?? options.sheets?.[options.currentSheetName ?? "sheet"] ?? {}, universe)
    };
    for (const [name, data] of Object.entries(options.sheets ?? {})) {
        namespaces[name] = sheetNamespaceForData(data, universe);
    }
    return namespaces;
}
function sheetNamespaceForData(data, universe) {
    const cells = {};
    for (const [cell, value] of Object.entries(data)) {
        cells[normalizeCell(cell)] = checkerTypeForRuntimeValue(value, universe);
    }
    return {
        cells,
        defaultType: universe.basic.any
    };
}
function checkerTypeForRuntimeValue(value, universe) {
    if (typeof value === "bigint")
        return universe.basic.int64;
    if (typeof value === "number")
        return universe.basic.float64;
    if (typeof value === "string")
        return universe.basic.string;
    if (typeof value === "boolean")
        return universe.basic.bool;
    return universe.basic.any;
}
function normalizeCell(cell) {
    return cell.replace(/\$/g, "").toUpperCase();
}
