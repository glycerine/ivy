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

type AnalysisSpreadsheetColumnWidths = Record<string, number>;
type AnalysisSpreadsheetRowHeights = Record<string, number>;

const ANALYSIS_DATA_COLUMN_COUNT = 26;
const ANALYSIS_DATA_COLUMN_LABELS = Array.from(
  { length: ANALYSIS_DATA_COLUMN_COUNT },
  (_, index) => spreadsheetColumnLabel(index),
);
const ANALYSIS_ROW_HEIGHT_DEFAULT = 28;
const ANALYSIS_ROW_HEIGHT_MIN = 22;
const ANALYSIS_ROW_HEIGHT_MAX = 220;
const ANALYSIS_COLUMN_WIDTH_DEFAULTS: AnalysisSpreadsheetColumnWidths = {
  lineNumber: 42,
  comment: 34,
  line: 360,
  ...Object.fromEntries(ANALYSIS_DATA_COLUMN_LABELS.map((label) => [label, 104])),
};
const ANALYSIS_COLUMN_WIDTH_MIN: AnalysisSpreadsheetColumnWidths = {
  lineNumber: 34,
  comment: 28,
  line: 120,
};
const ANALYSIS_COLUMN_WIDTH_MAX = 1400;

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
  var columnIds = table.getAllLeafColumns().map((column) => column.id);
  tableEl.style.minWidth = String(totalAnalysisColumnWidth(app, columnIds)) + 'px';
  tableEl.appendChild(renderAnalysisTableColgroup(app, columnIds, doc));
  tableEl.appendChild(renderAnalysisTableHead(app, table, doc));
  tableEl.appendChild(renderAnalysisTableBody(app, table, doc));
  return tableEl;
}

function renderAnalysisTableColgroup(app, columnIds: string[], doc: Document) {
  var colgroup = doc.createElement('colgroup');
  for (var columnId of columnIds) {
    var col = doc.createElement('col');
    col.dataset.columnId = columnId;
    col.style.width = String(analysisColumnWidth(app, columnId)) + 'px';
    colgroup.appendChild(col);
  }
  return colgroup;
}

function renderAnalysisTableHead(app, table, doc: Document) {
  var thead = doc.createElement('thead');
  for (var headerGroup of table.getHeaderGroups()) {
    var rowEl = doc.createElement('tr');
    for (var header of headerGroup.headers) {
      var th = doc.createElement('th');
      th.dataset.columnId = header.column.id;
      th.style.width = String(analysisColumnWidth(app, header.column.id)) + 'px';
      var label = doc.createElement('span');
      label.className = 'analysis-header-label';
      label.textContent = header.isPlaceholder ? '' : String(header.column.columnDef.header || '');
      th.appendChild(label);
      if (header.column.id === 'lineNumber') th.className = 'analysis-line-number-header';
      if (header.column.id === 'comment') th.className = 'analysis-comment-header';
      if (header.column.id === 'line') th.className = 'analysis-line-header';
      if (isAnalysisDataColumn(header.column.id)) th.className = 'analysis-data-header';
      th.appendChild(renderColumnResizeHandle(app, header.column.id, doc));
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
    tr.style.height = String(analysisRowHeight(app, row.index)) + 'px';
    for (var cell of row.getVisibleCells()) {
      var td = doc.createElement('td');
      td.dataset.columnId = cell.column.id;
      td.style.width = String(analysisColumnWidth(app, cell.column.id)) + 'px';
      td.style.height = String(analysisRowHeight(app, row.index)) + 'px';
      if (cell.column.id === 'lineNumber') {
        td.className = 'analysis-line-number-cell';
        td.textContent = String(row.original.lineNumber);
        td.appendChild(renderRowResizeHandle(app, row.index, doc));
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

function renderColumnResizeHandle(app, columnId: string, doc: Document) {
  var handle = doc.createElement('span');
  handle.className = 'analysis-column-resize-handle';
  handle.setAttribute('role', 'separator');
  handle.setAttribute('aria-label', 'Resize column ' + columnId);
  handle.addEventListener('mousedown', (event) => {
    event.preventDefault();
    event.stopPropagation();
    startAnalysisColumnResize(app, columnId, event, doc);
  });
  return handle;
}

function renderRowResizeHandle(app, rowIndex: number, doc: Document) {
  var handle = doc.createElement('span');
  handle.className = 'analysis-row-resize-handle';
  handle.setAttribute('role', 'separator');
  handle.setAttribute('aria-label', 'Resize row ' + String(rowIndex + 1));
  handle.addEventListener('mousedown', (event) => {
    event.preventDefault();
    event.stopPropagation();
    startAnalysisRowResize(app, rowIndex, event, doc);
  });
  return handle;
}

function startAnalysisColumnResize(app, columnId: string, event: MouseEvent, doc: Document) {
  var startX = event.clientX;
  var startWidth = analysisColumnWidth(app, columnId);
  var body = doc.body;
  body?.classList.add('analysis-resizing-columns');
  var onMove = (moveEvent: MouseEvent) => {
    var nextWidth = clampNumber(
      Math.round(startWidth + moveEvent.clientX - startX),
      analysisColumnMinWidth(columnId),
      ANALYSIS_COLUMN_WIDTH_MAX,
    );
    setAnalysisColumnWidth(app, columnId, nextWidth);
    applyAnalysisColumnWidths(app, doc);
  };
  var onUp = () => {
    body?.classList.remove('analysis-resizing-columns');
    doc.removeEventListener('mousemove', onMove);
    doc.removeEventListener('mouseup', onUp);
  };
  doc.addEventListener('mousemove', onMove);
  doc.addEventListener('mouseup', onUp);
}

function startAnalysisRowResize(app, rowIndex: number, event: MouseEvent, doc: Document) {
  var startY = event.clientY;
  var startHeight = analysisRowHeight(app, rowIndex);
  var body = doc.body;
  body?.classList.add('analysis-resizing-rows');
  var onMove = (moveEvent: MouseEvent) => {
    var nextHeight = clampNumber(
      Math.round(startHeight + moveEvent.clientY - startY),
      ANALYSIS_ROW_HEIGHT_MIN,
      ANALYSIS_ROW_HEIGHT_MAX,
    );
    setAnalysisRowHeight(app, rowIndex, nextHeight);
    applyAnalysisRowHeights(app, doc);
  };
  var onUp = () => {
    body?.classList.remove('analysis-resizing-rows');
    doc.removeEventListener('mousemove', onMove);
    doc.removeEventListener('mouseup', onUp);
  };
  doc.addEventListener('mousemove', onMove);
  doc.addEventListener('mouseup', onUp);
}

function renderLineInput(app, row: AnalysisSpreadsheetRow, doc: Document) {
  var input = doc.createElement('input');
  input.type = 'text';
  input.className = 'analysis-line-input';
  input.value = row.line;
  input.spellcheck = false;
  input.dataset.lineIndex = String(row.lineNumber - 1);
  input.setAttribute('aria-label', 'Spec line ' + row.lineNumber);
  var activate = () => {
    setAnalysisFormulaTarget(app, { kind: 'line', rowIndex: row.lineNumber - 1 }, input.value, doc);
  };
  input.addEventListener('focus', activate);
  input.addEventListener('click', activate);
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
  var activate = () => {
    setAnalysisFormulaTarget(app, { kind: 'cell', rowIndex: row.lineNumber - 1, columnId }, input.value, doc);
  };
  input.addEventListener('focus', activate);
  input.addEventListener('click', activate);
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
  var grid = doc.getElementById('analysis-spreadsheet-grid') as (HTMLElement & {
    __analysisFormulaFocusHandler?: (event: FocusEvent) => void;
  }) | null;
  if (grid) {
    if (grid.__analysisFormulaFocusHandler) {
      grid.removeEventListener('focusin', grid.__analysisFormulaFocusHandler);
    }
    grid.__analysisFormulaFocusHandler = (event: FocusEvent) => {
      setAnalysisFormulaTargetFromInput(app, event.target as HTMLInputElement | null, doc);
    };
    grid.addEventListener('focusin', grid.__analysisFormulaFocusHandler);
  }
}

function setAnalysisFormulaTargetFromInput(app, input: HTMLInputElement | null, doc: Document) {
  if (!input || typeof input.value !== 'string' || !input.classList) return;
  var rowIndex = Number(input.dataset.lineIndex || '0');
  if (input.classList.contains('analysis-line-input')) {
    setAnalysisFormulaTarget(app, { kind: 'line', rowIndex }, input.value, doc);
    return;
  }
  if (input.classList.contains('analysis-cell-input') && input.dataset.columnId) {
    setAnalysisFormulaTarget(app, { kind: 'cell', rowIndex, columnId: input.dataset.columnId }, input.value, doc);
  }
}

function setAnalysisFormulaTarget(app, target: AnalysisFormulaTarget, value: string, doc: Document) {
  app._analysisSpreadsheetFormulaTarget = target;
  var formulaInput = analysisFormulaInput(doc);
  if (formulaInput) formulaInput.value = value;
  updateAnalysisFormulaCellLabel(target, doc);
}

function refreshAnalysisFormulaInput(app, doc: Document) {
  var target = app?._analysisSpreadsheetFormulaTarget as AnalysisFormulaTarget | undefined;
  var formulaInput = analysisFormulaInput(doc);
  if (!target) {
    updateAnalysisFormulaCellLabel(null, doc);
    return;
  }
  if (formulaInput) formulaInput.value = formulaValueForTarget(app, target);
  updateAnalysisFormulaCellLabel(target, doc);
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

function updateAnalysisFormulaCellLabel(target: AnalysisFormulaTarget | null, doc: Document) {
  var label = analysisFormulaCellLabel(doc);
  if (label) label.textContent = target ? formulaCellLabelForTarget(target) : '';
}

function formulaCellLabelForTarget(target: AnalysisFormulaTarget) {
  if (target.kind === 'line') return 'spec' + String(target.rowIndex + 1);
  return target.columnId + String(target.rowIndex + 1);
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

function analysisFormulaCellLabel(doc: Document) {
  return doc.getElementById('analysis-formula-cell-label') as HTMLSpanElement | null;
}

function analysisColumnWidths(app): AnalysisSpreadsheetColumnWidths {
  if (!app._analysisSpreadsheetColumnWidths || typeof app._analysisSpreadsheetColumnWidths !== 'object') {
    app._analysisSpreadsheetColumnWidths = {};
  }
  return app._analysisSpreadsheetColumnWidths;
}

function analysisRowHeights(app): AnalysisSpreadsheetRowHeights {
  if (!app._analysisSpreadsheetRowHeights || typeof app._analysisSpreadsheetRowHeights !== 'object') {
    app._analysisSpreadsheetRowHeights = {};
  }
  return app._analysisSpreadsheetRowHeights;
}

function analysisColumnWidth(app, columnId: string) {
  var stored = Number(analysisColumnWidths(app)[columnId]);
  return Number.isFinite(stored) && stored > 0
    ? stored
    : ANALYSIS_COLUMN_WIDTH_DEFAULTS[columnId] || 104;
}

function setAnalysisColumnWidth(app, columnId: string, width: number) {
  analysisColumnWidths(app)[columnId] = width;
}

function analysisColumnMinWidth(columnId: string) {
  return ANALYSIS_COLUMN_WIDTH_MIN[columnId] || 48;
}

function analysisRowHeight(app, rowIndex: number) {
  var stored = Number(analysisRowHeights(app)[String(rowIndex + 1)]);
  return Number.isFinite(stored) && stored > 0 ? stored : ANALYSIS_ROW_HEIGHT_DEFAULT;
}

function setAnalysisRowHeight(app, rowIndex: number, height: number) {
  analysisRowHeights(app)[String(rowIndex + 1)] = height;
}

function totalAnalysisColumnWidth(app, columnIds: string[]) {
  return columnIds.reduce((total, columnId) => total + analysisColumnWidth(app, columnId), 0);
}

function applyAnalysisColumnWidths(app, doc: Document) {
  var table = doc.querySelector('.analysis-spreadsheet-table') as HTMLTableElement | null;
  if (!table) return;
  var columnIds = Array.from(table.querySelectorAll<HTMLTableColElement>('col[data-column-id]'))
    .map((col) => col.dataset.columnId || '')
    .filter(Boolean);
  table.style.minWidth = String(totalAnalysisColumnWidth(app, columnIds)) + 'px';
  for (var columnId of columnIds) {
    var width = String(analysisColumnWidth(app, columnId)) + 'px';
    for (var element of table.querySelectorAll<HTMLElement>('[data-column-id="' + columnId + '"]')) {
      element.style.width = width;
    }
  }
}

function applyAnalysisRowHeights(app, doc: Document) {
  var table = doc.querySelector('.analysis-spreadsheet-table');
  if (!table) return;
  for (var row of table.querySelectorAll<HTMLTableRowElement>('tbody tr[data-line-index]')) {
    var rowIndex = Number(row.dataset.lineIndex || '0');
    var height = String(analysisRowHeight(app, rowIndex)) + 'px';
    row.style.height = height;
    for (var cell of row.querySelectorAll<HTMLElement>('td')) {
      cell.style.height = height;
    }
  }
}

function clampNumber(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
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
