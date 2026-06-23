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
function finiteOr(value, fallback) {
    return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
