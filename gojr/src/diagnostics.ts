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
  const startOffset = finiteOr(token.startOffset, 0);
  const endOffset = finiteOr(token.endOffset, startOffset);
  return {
    offset: startOffset,
    length: Math.max(0, endOffset - startOffset + 1),
    line: finiteOr(token.startLine, 1),
    column: finiteOr(token.startColumn, 1)
  };
}

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
