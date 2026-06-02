import {
  createTable,
  getCoreRowModel,
  type ColumnDef,
} from '@tanstack/table-core';
import { editorContent } from './editorService.ts';

export interface AnalysisSpreadsheetRow {
  id: string;
  lineNumber: number;
  line: string;
  cells: Record<string, string>;
}

type AnalysisFormulaTarget =
  | { kind: 'line'; rowIndex: number }
  | { kind: 'cell'; rowIndex: number; columnId: string };

const ANALYSIS_DATA_COLUMN_COUNT = 26;
const ANALYSIS_DATA_COLUMN_LABELS = Array.from(
  { length: ANALYSIS_DATA_COLUMN_COUNT },
  (_, index) => spreadsheetColumnLabel(index),
);

const ANALYSIS_COLUMNS: ColumnDef<AnalysisSpreadsheetRow>[] = [
  {
    id: 'lineNumber',
    header: 'Line',
    accessorKey: 'lineNumber',
  },
  {
    id: 'comment',
    header: '#',
    accessorFn: (row) => (rowHasProtectedCommentMarker(row) ? '' : lineIsCommented(row.line) ? '#' : ''),
  },
  {
    id: 'line',
    header: 'Spec line',
    accessorKey: 'line',
  },
  ...ANALYSIS_DATA_COLUMN_LABELS.map((label) => ({
    id: label,
    header: label,
    accessorFn: (row: AnalysisSpreadsheetRow) => row.cells[label] || '',
  })),
];

export function spreadsheetColumnLabel(index: number) {
  var label = '';
  var next = index;
  do {
    label = String.fromCharCode(97 + (next % 26)) + label;
    next = Math.floor(next / 26) - 1;
  } while (next >= 0);
  return label;
}

export function linesFromEditorContent(content = '') {
  return content.length > 0 ? content.split('\n') : [''];
}

export function rowsFromEditorContent(
  content = '',
  cellsByLine: Record<string, Record<string, string>> = {},
): AnalysisSpreadsheetRow[] {
  return linesFromEditorContent(content).map((line, index) => ({
    id: String(index + 1),
    lineNumber: index + 1,
    line,
    cells: { ...(cellsByLine[String(index + 1)] || {}) },
  }));
}

export function lineIsCommented(line = '') {
  return /^\s*#/.test(line);
}

export function lineIsLanguageDirective(line = '') {
  return /^\s*#lang\b/i.test(line);
}

export function rowHasProtectedCommentMarker(row: AnalysisSpreadsheetRow) {
  return row.lineNumber === 1 && lineIsLanguageDirective(row.line);
}

export function toggleLineComment(line = '') {
  var match = line.match(/^(\s*)#(.*)$/);
  if (match) return match[1] + match[2];
  var indent = line.match(/^\s*/)?.[0] || '';
  return indent + '#' + line.slice(indent.length);
}

export function setupAnalysisSpreadsheet(app, {
  doc = globalThis.document,
}: { doc?: Document } = {}) {
  return syncAnalysisSpreadsheetFromEditor(app, { doc });
}

export function syncAnalysisSpreadsheetFromEditor(app, {
  doc = globalThis.document,
}: { doc?: Document } = {}) {
  if (!app || app._analysisSpreadsheetApplyingEdit) return false;
  return renderAnalysisSpreadsheet(app, editorContent(app), { doc });
}

export function renderAnalysisSpreadsheet(app, content = '', {
  doc = globalThis.document,
}: { doc?: Document } = {}) {
  var container = doc && doc.getElementById('analysis-spreadsheet-grid');
  if (!container) return false;

  var rows = rowsFromEditorContent(content, analysisSpreadsheetCells(app));
  var table = createAnalysisTable(rows);
  bindAnalysisFormulaInput(app, doc);
  container.textContent = '';
  container.appendChild(renderAnalysisTable(app, table, doc));
  refreshAnalysisFormulaInput(app, doc);
  return true;
}

export function applyAnalysisSpreadsheetLineEdit(app, rowIndex, nextLine, {
  doc = globalThis.document,
  render = false,
}: { doc?: Document; render?: boolean } = {}) {
  if (!app) return '';
  var lines = linesFromEditorContent(editorContent(app));
  while (lines.length <= rowIndex) lines.push('');
  if (lines[rowIndex] === nextLine) return lines.join('\n');
  lines[rowIndex] = nextLine;
  var nextContent = lines.join('\n');
  replaceEditorLineFromSpreadsheet(app, rowIndex, nextLine, nextContent);
  if (render) renderAnalysisSpreadsheet(app, nextContent, { doc });
  return nextContent;
}

export function toggleAnalysisSpreadsheetLineComment(app, rowIndex, {
  doc = globalThis.document,
}: { doc?: Document } = {}) {
  var lines = linesFromEditorContent(editorContent(app));
  while (lines.length <= rowIndex) lines.push('');
  if (rowIndex === 0 && lineIsLanguageDirective(lines[rowIndex])) {
    return lines.join('\n');
  }
  return applyAnalysisSpreadsheetLineEdit(app, rowIndex, toggleLineComment(lines[rowIndex]), {
    doc,
    render: true,
  });
}

export function applyAnalysisSpreadsheetCellEdit(app, rowIndex: number, columnId: string, nextValue: string) {
  if (!app) return '';
  var cells = analysisSpreadsheetCells(app);
  var rowKey = String(rowIndex + 1);
  cells[rowKey] = cells[rowKey] || {};
  cells[rowKey][columnId] = nextValue;
  return nextValue;
}

function createAnalysisTable(data: AnalysisSpreadsheetRow[]) {
  var state: any = {};
  var table = createTable({
    data,
    columns: ANALYSIS_COLUMNS,
    state,
    onStateChange: (updater) => {
      state = typeof updater === 'function' ? updater(state) : updater;
      table.setOptions((previous) => ({ ...previous, state }));
    },
    renderFallbackValue: null,
    getCoreRowModel: getCoreRowModel(),
  });
  state = table.initialState;
  table.setOptions((previous) => ({ ...previous, state }));
  return table;
}

function renderAnalysisTable(app, table, doc: Document) {
  var tableEl = doc.createElement('table');
  tableEl.className = 'analysis-spreadsheet-table';
  tableEl.appendChild(renderAnalysisTableHead(table, doc));
  tableEl.appendChild(renderAnalysisTableBody(app, table, doc));
  return tableEl;
}

function renderAnalysisTableHead(table, doc: Document) {
  var thead = doc.createElement('thead');
  for (var headerGroup of table.getHeaderGroups()) {
    var rowEl = doc.createElement('tr');
    for (var header of headerGroup.headers) {
      var th = doc.createElement('th');
      th.textContent = header.isPlaceholder ? '' : String(header.column.columnDef.header || '');
      if (header.column.id === 'lineNumber') th.className = 'analysis-line-number-header';
      if (header.column.id === 'comment') th.className = 'analysis-comment-header';
      if (header.column.id === 'line') th.className = 'analysis-line-header';
      if (isAnalysisDataColumn(header.column.id)) th.className = 'analysis-data-header';
      rowEl.appendChild(th);
    }
    thead.appendChild(rowEl);
  }
  return thead;
}

function renderAnalysisTableBody(app, table, doc: Document) {
  var tbody = doc.createElement('tbody');
  for (var row of table.getRowModel().rows) {
    var tr = doc.createElement('tr');
    tr.dataset.lineIndex = String(row.index);
    for (var cell of row.getVisibleCells()) {
      var td = doc.createElement('td');
      if (cell.column.id === 'lineNumber') {
        td.className = 'analysis-line-number-cell';
        td.textContent = String(row.original.lineNumber);
      } else if (cell.column.id === 'comment') {
        td.className = 'analysis-comment-cell';
        renderCommentCell(td, app, row.original, doc);
      } else if (cell.column.id === 'line') {
        td.className = 'analysis-line-cell';
        td.appendChild(renderLineInput(app, row.original, doc));
      } else {
        td.className = 'analysis-data-cell';
        td.appendChild(renderAnalysisCellInput(app, row.original, cell.column.id, doc));
      }
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }
  return tbody;
}

function renderCommentCell(cell: HTMLTableCellElement, app, row: AnalysisSpreadsheetRow, doc: Document) {
  cell.textContent = '';
  if (!rowHasProtectedCommentMarker(row)) {
    cell.appendChild(renderCommentToggle(app, row, doc));
  }
}

function renderCommentToggle(app, row: AnalysisSpreadsheetRow, doc: Document) {
  var button = doc.createElement('button');
  button.type = 'button';
  button.className = 'analysis-comment-toggle';
  button.textContent = lineIsCommented(row.line) ? '#' : '';
  button.setAttribute('aria-label', 'Toggle comment on line ' + row.lineNumber);
  button.setAttribute('aria-pressed', lineIsCommented(row.line) ? 'true' : 'false');
  button.addEventListener('click', () => {
    toggleAnalysisSpreadsheetLineComment(app, row.lineNumber - 1, { doc });
  });
  return button;
}

function renderLineInput(app, row: AnalysisSpreadsheetRow, doc: Document) {
  var input = doc.createElement('input');
  input.type = 'text';
  input.className = 'analysis-line-input';
  input.value = row.line;
  input.spellcheck = false;
  input.dataset.lineIndex = String(row.lineNumber - 1);
  input.setAttribute('aria-label', 'Spec line ' + row.lineNumber);
  input.addEventListener('focus', () => {
    setAnalysisFormulaTarget(app, { kind: 'line', rowIndex: row.lineNumber - 1 }, input.value, doc);
  });
  input.addEventListener('input', () => {
    var lineIndex = Number(input.dataset.lineIndex || '0');
    applyAnalysisSpreadsheetLineEdit(app, lineIndex, input.value, { doc });
    syncCommentCellForLineInput(app, row, input, doc);
    updateAnalysisFormulaInputIfTargetMatches(app, { kind: 'line', rowIndex: lineIndex }, input.value, doc);
  });
  return input;
}

function renderAnalysisCellInput(app, row: AnalysisSpreadsheetRow, columnId: string, doc: Document) {
  var input = doc.createElement('input');
  input.type = 'text';
  input.className = 'analysis-cell-input';
  input.value = row.cells[columnId] || '';
  input.spellcheck = false;
  input.dataset.lineIndex = String(row.lineNumber - 1);
  input.dataset.columnId = columnId;
  input.setAttribute('aria-label', 'Analysis cell ' + columnId + ' line ' + row.lineNumber);
  input.addEventListener('focus', () => {
    setAnalysisFormulaTarget(app, { kind: 'cell', rowIndex: row.lineNumber - 1, columnId }, input.value, doc);
  });
  input.addEventListener('input', () => {
    var lineIndex = Number(input.dataset.lineIndex || '0');
    applyAnalysisSpreadsheetCellEdit(app, lineIndex, columnId, input.value);
    updateAnalysisFormulaInputIfTargetMatches(app, { kind: 'cell', rowIndex: lineIndex, columnId }, input.value, doc);
  });
  return input;
}

function bindAnalysisFormulaInput(app, doc: Document) {
  var formulaInput = analysisFormulaInput(doc);
  if (!formulaInput) return;
  formulaInput.oninput = () => {
    applyAnalysisFormulaEdit(app, formulaInput.value, doc);
  };
}

function setAnalysisFormulaTarget(app, target: AnalysisFormulaTarget, value: string, doc: Document) {
  app._analysisSpreadsheetFormulaTarget = target;
  var formulaInput = analysisFormulaInput(doc);
  if (formulaInput) formulaInput.value = value;
}

function refreshAnalysisFormulaInput(app, doc: Document) {
  var target = app?._analysisSpreadsheetFormulaTarget as AnalysisFormulaTarget | undefined;
  var formulaInput = analysisFormulaInput(doc);
  if (!target || !formulaInput) return;
  formulaInput.value = formulaValueForTarget(app, target);
}

function applyAnalysisFormulaEdit(app, value: string, doc: Document) {
  var target = app?._analysisSpreadsheetFormulaTarget as AnalysisFormulaTarget | undefined;
  if (!target) return;
  if (target.kind === 'line') {
    applyAnalysisSpreadsheetLineEdit(app, target.rowIndex, value, { doc });
    var lineInput = doc.querySelector(
      '.analysis-line-input[data-line-index="' + target.rowIndex + '"]',
    ) as HTMLInputElement | null;
    if (lineInput) {
      lineInput.value = value;
      syncCommentCellForLineInput(app, rowFromLineInput(lineInput, value), lineInput, doc);
    }
    return;
  }
  applyAnalysisSpreadsheetCellEdit(app, target.rowIndex, target.columnId, value);
  var cellInput = doc.querySelector(
    '.analysis-cell-input[data-line-index="' + target.rowIndex + '"][data-column-id="' + target.columnId + '"]',
  ) as HTMLInputElement | null;
  if (cellInput) cellInput.value = value;
}

function updateAnalysisFormulaInputIfTargetMatches(
  app,
  target: AnalysisFormulaTarget,
  value: string,
  doc: Document,
) {
  var currentTarget = app?._analysisSpreadsheetFormulaTarget as AnalysisFormulaTarget | undefined;
  var formulaInput = analysisFormulaInput(doc);
  if (formulaInput && currentTarget && analysisFormulaTargetsMatch(currentTarget, target)) {
    formulaInput.value = value;
  }
}

function analysisFormulaTargetsMatch(a: AnalysisFormulaTarget, b: AnalysisFormulaTarget) {
  return a.kind === b.kind
    && a.rowIndex === b.rowIndex
    && (a.kind === 'line' || a.columnId === (b as { columnId: string }).columnId);
}

function formulaValueForTarget(app, target: AnalysisFormulaTarget) {
  if (target.kind === 'line') {
    return linesFromEditorContent(editorContent(app))[target.rowIndex] || '';
  }
  return analysisSpreadsheetCells(app)[String(target.rowIndex + 1)]?.[target.columnId] || '';
}

function rowFromLineInput(input: HTMLInputElement, line: string): AnalysisSpreadsheetRow {
  var lineNumber = Number(input.dataset.lineIndex || '0') + 1;
  return {
    id: String(lineNumber),
    lineNumber,
    line,
    cells: {},
  };
}

function syncCommentCellForLineInput(app, row: AnalysisSpreadsheetRow, input: HTMLInputElement, doc: Document) {
  var commentCell = input.closest('tr')?.querySelector('.analysis-comment-cell') as HTMLTableCellElement | null;
  if (!commentCell) return;
  var updatedRow = { ...row, line: input.value };
  if (rowHasProtectedCommentMarker(updatedRow)) {
    commentCell.textContent = '';
    return;
  }
  var toggle = commentCell.querySelector('.analysis-comment-toggle') as HTMLButtonElement | null;
  if (!toggle) {
    commentCell.appendChild(renderCommentToggle(app, updatedRow, doc));
    return;
  }
  var commented = lineIsCommented(input.value);
  toggle.textContent = commented ? '#' : '';
  toggle.setAttribute('aria-pressed', commented ? 'true' : 'false');
}

function analysisFormulaInput(doc: Document) {
  return doc.getElementById('analysis-formula-input') as HTMLInputElement | null;
}

function analysisSpreadsheetCells(app): Record<string, Record<string, string>> {
  if (!app._analysisSpreadsheetCells || typeof app._analysisSpreadsheetCells !== 'object') {
    app._analysisSpreadsheetCells = {};
  }
  return app._analysisSpreadsheetCells;
}

function isAnalysisDataColumn(columnId: string) {
  return ANALYSIS_DATA_COLUMN_LABELS.includes(columnId);
}

function replaceEditorLineFromSpreadsheet(app, rowIndex, nextLine, nextContent) {
  var editor = app && app.cmEditor;
  app._analysisSpreadsheetApplyingEdit = true;
  try {
    if (editor && typeof editor.replaceRange === 'function') {
      var previousLine = typeof editor.getLine === 'function'
        ? editor.getLine(rowIndex) || ''
        : linesFromEditorContent(editorContent(app))[rowIndex] || '';
      var replace = () => {
        editor.replaceRange(nextLine, { line: rowIndex, ch: 0 }, { line: rowIndex, ch: previousLine.length });
      };
      if (typeof editor.operation === 'function') {
        editor.operation(replace);
      } else {
        replace();
      }
    } else if (editor && typeof editor.setValue === 'function') {
      editor.setValue(nextContent);
    }
  } finally {
    app._analysisSpreadsheetApplyingEdit = false;
  }

  app._persistedFileContent = nextContent;
  if (typeof app._updateEditorLabel === 'function') app._updateEditorLabel();
  if (typeof app._invalidateModelState === 'function') app._invalidateModelState('editor-change');
}
