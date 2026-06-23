export type DiagnosticSeverity = "error" | "warning" | "info";

export interface SourceSpan {
  offset: number;
  length: number;
  line: number;
  column: number;
}

export interface Diagnostic {
  code: string;
  severity: DiagnosticSeverity;
  message: string;
  span?: SourceSpan;
}

export function spanFromToken(token: {
  startOffset: number;
  endOffset?: number;
  startLine?: number;
  startColumn?: number;
}): SourceSpan {
  const endOffset = token.endOffset ?? token.startOffset;
  return {
    offset: token.startOffset,
    length: Math.max(0, endOffset - token.startOffset + 1),
    line: token.startLine ?? 1,
    column: token.startColumn ?? 1
  };
}
