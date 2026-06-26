export const REPL_FILENAME = "gojr-repl.go";
export function spanFromToken(token) {
    const startOffset = finiteOr(token.startOffset, 0);
    const endOffset = finiteOr(token.endOffset, startOffset);
    return {
        filename: token.filename ?? REPL_FILENAME,
        offset: startOffset,
        length: Math.max(0, endOffset - startOffset + 1),
        line: finiteOr(token.startLine, 1),
        column: finiteOr(token.startColumn, 1)
    };
}
export function diagnosticFilename(span, filename = REPL_FILENAME) {
    return span?.filename ?? filename;
}
export function formatDiagnostic(diagnostic) {
    const filename = diagnostic.span?.filename ?? diagnostic.filename ?? REPL_FILENAME;
    const location = diagnostic.span
        ? `${filename}:${diagnostic.span.line}:${diagnostic.span.column}: `
        : `${filename}: `;
    const sourceContext = diagnosticSourceContext(diagnostic);
    const stack = typeof diagnostic.stack === "string" && diagnostic.stack !== ""
        ? `\n${diagnostic.stack}`
        : "";
    return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}${sourceContext}${stack}`;
}
export function withDiagnosticSourceContext(diagnostics, sourceFiles) {
    return diagnostics.map((diagnostic) => diagnosticWithSourceContext(diagnostic, sourceFiles));
}
export function diagnosticWithSourceContext(diagnostic, sourceFiles) {
    if (diagnostic.sourceLine !== undefined || diagnostic.span === undefined)
        return diagnostic;
    const sourceLine = sourceLineForSpan(sourceFiles, diagnostic.span);
    return sourceLine === undefined ? diagnostic : { ...diagnostic, sourceLine };
}
function diagnosticSourceContext(diagnostic) {
    if (diagnostic.sourceLine === undefined || diagnostic.span === undefined)
        return "";
    return `\n${diagnostic.sourceLine}\n${caretLine(diagnostic.sourceLine, diagnostic.span.column)}`;
}
function caretLine(line, column) {
    const prefixLength = Math.max(0, Math.min(line.length, column - 1));
    const prefix = line.slice(0, prefixLength).replace(/[^\t]/g, " ");
    return `${prefix}^`;
}
function sourceLineForSpan(sourceFiles, span) {
    const file = sourceFiles.find((candidate) => candidate.filename === span.filename)
        ?? (sourceFiles.length === 1 ? sourceFiles[0] : undefined);
    if (!file || span.line < 1)
        return undefined;
    return lineAt(file.source, span.line);
}
function lineAt(source, lineNumber) {
    let line = 1;
    let start = 0;
    for (let index = 0; index <= source.length; index += 1) {
        if (index === source.length || source.charCodeAt(index) === 10) {
            if (line === lineNumber) {
                const end = index > start && source.charCodeAt(index - 1) === 13 ? index - 1 : index;
                return source.slice(start, end);
            }
            line += 1;
            start = index + 1;
        }
    }
    return undefined;
}
export function hasErrorDiagnostics(diagnostics) {
    return diagnostics.some((diagnostic) => diagnostic.severity === "error");
}
function finiteOr(value, fallback) {
    return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
