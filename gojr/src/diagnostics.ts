export type DiagnosticSeverity = "error" | "warning" | "info";

export const REPL_FILENAME = "gojr-repl.go";

export interface SourceFile {
  filename: string;
  source: string;
}

export interface SourceSpan {
  filename: string;
  offset: number;
  length: number;
  line: number;
  column: number;
}

export interface Diagnostic {
  filename: string;
  code: string;
  severity: DiagnosticSeverity;
  message: string;
  stack?: string;
  span?: SourceSpan;
}

export function spanFromToken(token: {
  filename?: string;
  startOffset: number;
  endOffset?: number;
  startLine?: number;
  startColumn?: number;
}): SourceSpan {
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

export function diagnosticFilename(span?: SourceSpan, filename = REPL_FILENAME): string {
  return span?.filename ?? filename;
}

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
