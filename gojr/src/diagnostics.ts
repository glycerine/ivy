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

export interface DiagnosticLike {
  filename?: string;
  code: string;
  severity: DiagnosticSeverity;
  message: string;
  stack?: string;
  span?: Pick<SourceSpan, "filename" | "line" | "column">;
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

export function formatDiagnostic(diagnostic: DiagnosticLike): string {
  const filename = diagnostic.span?.filename ?? diagnostic.filename ?? REPL_FILENAME;
  const location = diagnostic.span
    ? `${filename}:${diagnostic.span.line}:${diagnostic.span.column}: `
    : `${filename}: `;
  const stack = typeof diagnostic.stack === "string" && diagnostic.stack !== ""
    ? `\n${diagnostic.stack}`
    : "";
  return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}${stack}`;
}

export function hasErrorDiagnostics(diagnostics: readonly DiagnosticLike[]): boolean {
  return diagnostics.some((diagnostic) => diagnostic.severity === "error");
}

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
