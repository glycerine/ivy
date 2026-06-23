export function spanFromToken(token) {
    const startOffset = finiteOr(token.startOffset, 0);
    const endOffset = finiteOr(token.endOffset, startOffset);
    return {
        offset: startOffset,
        length: Math.max(0, endOffset - startOffset + 1),
        line: finiteOr(token.startLine, 1),
        column: finiteOr(token.startColumn, 1)
    };
}
function finiteOr(value, fallback) {
    return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
