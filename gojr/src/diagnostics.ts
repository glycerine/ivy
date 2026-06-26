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
  sourceLine?: string;
}

export interface DiagnosticLike {
  filename?: string;
  code: string;
  severity: DiagnosticSeverity;
  message: string;
  stack?: string;
  span?: Pick<SourceSpan, "filename" | "line" | "column"> & Partial<Pick<SourceSpan, "length">>;
  sourceLine?: string;
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
  const sourceContext = diagnosticSourceContext(diagnostic);
  const stack = typeof diagnostic.stack === "string" && diagnostic.stack !== ""
    ? `\n${diagnostic.stack}`
    : "";
  return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}${sourceContext}${stack}`;
}

export function withDiagnosticSourceContext<T extends Diagnostic>(diagnostics: readonly T[], sourceFiles: readonly SourceFile[]): T[] {
  return diagnostics.map((diagnostic) => diagnosticWithSourceContext(diagnostic, sourceFiles));
}

export function diagnosticWithSourceContext<T extends Diagnostic>(diagnostic: T, sourceFiles: readonly SourceFile[]): T {
  if (diagnostic.sourceLine !== undefined || diagnostic.span === undefined) return diagnostic;
  const sourceLine = sourceLineForSpan(sourceFiles, diagnostic.span);
  return sourceLine === undefined ? diagnostic : { ...diagnostic, sourceLine };
}

function diagnosticSourceContext(diagnostic: DiagnosticLike): string {
  if (diagnostic.sourceLine === undefined || diagnostic.span === undefined) return "";
  return `\n${diagnostic.sourceLine}\n${caretLine(diagnostic.sourceLine, diagnostic.span.column)}`;
}

function caretLine(line: string, column: number): string {
  const prefixLength = Math.max(0, Math.min(line.length, column - 1));
  const prefix = line.slice(0, prefixLength).replace(/[^\t]/g, " ");
  return `${prefix}^`;
}

function sourceLineForSpan(sourceFiles: readonly SourceFile[], span: Pick<SourceSpan, "filename" | "line">): string | undefined {
  const file = sourceFiles.find((candidate) => candidate.filename === span.filename)
    ?? (sourceFiles.length === 1 ? sourceFiles[0] : undefined);
  if (!file || span.line < 1) return undefined;
  return lineAt(file.source, span.line);
}

function lineAt(source: string, lineNumber: number): string | undefined {
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

export function hasErrorDiagnostics(diagnostics: readonly DiagnosticLike[]): boolean {
  return diagnostics.some((diagnostic) => diagnostic.severity === "error");
}

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}
