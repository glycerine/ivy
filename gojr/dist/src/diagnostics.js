export function spanFromToken(token) {
    const endOffset = token.endOffset ?? token.startOffset;
    return {
        offset: token.startOffset,
        length: Math.max(0, endOffset - token.startOffset + 1),
        line: token.startLine ?? 1,
        column: token.startColumn ?? 1
    };
}
