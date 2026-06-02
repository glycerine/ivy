import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  linesFromEditorContent,
  lineIsLanguageDirective,
  renderAnalysisSpreadsheet,
  rowsFromEditorContent,
  spreadsheetColumnLabel,
  syncAnalysisSpreadsheetFromEditor,
  toggleAnalysisSpreadsheetLineComment,
  toggleLineComment,
} from './analysisSpreadsheetService.ts';

function makeSpreadsheetApp(initialContent = 'type client\nrelation link(X,Y)') {
  let content = initialContent;
  return {
    _persistedFileContent: content,
    _analysisSpreadsheetCells: {},
    _analysisSpreadsheetColumnWidths: {},
    _analysisSpreadsheetRowHeights: {},
    _updateEditorLabel: vi.fn(),
    _invalidateModelState: vi.fn(),
    cmEditor: {
      getValue: vi.fn(() => content),
      getLine: vi.fn((line) => linesFromEditorContent(content)[line] || ''),
      replaceRange: vi.fn((nextLine, from, to) => {
        const lines = linesFromEditorContent(content);
        lines[from.line] = nextLine + lines[from.line].slice(to.ch);
        content = lines.join('\n');
      }),
    },
    currentContent: () => content,
  };
}

function mountSpreadsheetShell() {
  document.body.innerHTML = [
    '<div class="analysis-formula-bar">',
    '<span id="analysis-formula-cell-label"></span>',
    '<input id="analysis-formula-input">',
    '</div>',
    '<div id="analysis-spreadsheet-grid"></div>',
  ].join('');
}

afterEach(() => {
  document.body.innerHTML = '';
});

describe('analysisSpreadsheetService', () => {
  it('creates one spreadsheet row for each editor line', () => {
    expect(linesFromEditorContent('a\nb\n')).toEqual(['a', 'b', '']);
    expect(rowsFromEditorContent('a\nb').map((row) => row.lineNumber)).toEqual([1, 2]);
  });

  it('generates lower-case spreadsheet column labels', () => {
    expect(spreadsheetColumnLabel(0)).toBe('a');
    expect(spreadsheetColumnLabel(25)).toBe('z');
    expect(spreadsheetColumnLabel(26)).toBe('aa');
    expect(spreadsheetColumnLabel(27)).toBe('ab');
  });

  it('toggles a line comment marker at the first non-indent position', () => {
    expect(toggleLineComment('  action connect')).toBe('  #action connect');
    expect(toggleLineComment('  #action connect')).toBe('  action connect');
    expect(lineIsLanguageDirective('#lang ivy1.7')).toBe(true);
  });

  it('renders a TanStack-backed editable table from the editor buffer', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client\nrelation link(X,Y)');

    expect(renderAnalysisSpreadsheet(app, app.currentContent())).toBe(true);

    const headers = Array.from(document.querySelectorAll('th')).map((th) => th.textContent);
    const lineNumbers = Array.from(document.querySelectorAll('.analysis-line-number-cell')).map((cell) => cell.textContent);
    const inputs = Array.from(document.querySelectorAll<HTMLInputElement>('.analysis-line-input'));
    expect(headers.slice(0, 8)).toEqual(['Line', '#', 'Spec line', 'a', 'b', 'c', 'd', 'e']);
    expect(headers.at(-1)).toBe('z');
    expect(headers).toHaveLength(29);
    expect(lineNumbers).toEqual(['1', '2']);
    expect(inputs.map((input) => input.value)).toEqual(['type client', 'relation link(X,Y)']);
  });

  it('edits lettered analysis cells without changing the editor buffer', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client\ntype server');
    syncAnalysisSpreadsheetFromEditor(app);

    const firstA = document.querySelector<HTMLInputElement>(
      '.analysis-cell-input[data-line-index="0"][data-column-id="a"]',
    )!;
    firstA.value = 'reachable';
    firstA.dispatchEvent(new Event('input', { bubbles: true }));

    expect(app.currentContent()).toBe('type client\ntype server');
    expect(app._analysisSpreadsheetCells).toEqual({ '1': { a: 'reachable' } });

    renderAnalysisSpreadsheet(app, app.currentContent());

    expect(document.querySelector<HTMLInputElement>(
      '.analysis-cell-input[data-line-index="0"][data-column-id="a"]',
    )!.value).toBe('reachable');
  });

  it('resizes spreadsheet columns by dragging header dividers', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client');
    syncAnalysisSpreadsheetFromEditor(app);

    const handle = document.querySelector<HTMLElement>(
      '.analysis-line-header .analysis-column-resize-handle',
    )!;
    handle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientX: 360 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientX: 300 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(app._analysisSpreadsheetColumnWidths.line).toBe(300);
    expect(document.querySelector<HTMLTableColElement>('col[data-column-id="line"]')!.style.width).toBe('300px');
    expect(document.querySelector<HTMLElement>('td[data-column-id="line"]')!.style.width).toBe('300px');

    renderAnalysisSpreadsheet(app, app.currentContent());

    expect(document.querySelector<HTMLTableColElement>('col[data-column-id="line"]')!.style.width).toBe('300px');
  });

  it('resizes spreadsheet rows by dragging line-number dividers', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client\ntype server');
    syncAnalysisSpreadsheetFromEditor(app);

    const handle = document.querySelector<HTMLElement>(
      '.analysis-line-number-cell .analysis-row-resize-handle',
    )!;
    handle.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientY: 28 }));
    document.dispatchEvent(new MouseEvent('mousemove', { bubbles: true, clientY: 45 }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(app._analysisSpreadsheetRowHeights['1']).toBe(45);
    expect(document.querySelector<HTMLElement>('tbody tr[data-line-index="0"]')!.style.height).toBe('45px');
    expect(document.querySelector<HTMLElement>('td[data-column-id="lineNumber"]')!.style.height).toBe('45px');

    renderAnalysisSpreadsheet(app, app.currentContent());

    expect(document.querySelector<HTMLElement>('tbody tr[data-line-index="0"]')!.style.height).toBe('45px');
  });

  it('uses the formula bar to view and edit the selected analysis cell', () => {
    mountSpreadsheetShell();
    const app = makeSpreadsheetApp('type client');
    syncAnalysisSpreadsheetFromEditor(app);

    const formula = document.getElementById('analysis-formula-input') as HTMLInputElement;
    const label = document.getElementById('analysis-formula-cell-label') as HTMLSpanElement;
    const firstA = document.querySelector<HTMLInputElement>(
      '.analysis-cell-input[data-line-index="0"][data-column-id="a"]',
    )!;

    firstA.value = 'reachable';
    firstA.dispatchEvent(new Event('focusin', { bubbles: true }));
    firstA.dispatchEvent(new Event('input', { bubbles: true }));

    expect(formula.value).toBe('reachable');
    expect(label.textContent).toBe('a1');

    formula.value = '=a1';
    formula.dispatchEvent(new Event('input', { bubbles: true }));

    expect(firstA.value).toBe('=a1');
    expect(app.currentContent()).toBe('type client');
    expect(app._analysisSpreadsheetCells).toEqual({ '1': { a: '=a1' } });
  });

  it('uses the formula bar to view and edit the selected spec line cell', () => {
    mountSpreadsheetShell();
    const app = makeSpreadsheetApp('type client');
    syncAnalysisSpreadsheetFromEditor(app);

    const formula = document.getElementById('analysis-formula-input') as HTMLInputElement;
    const label = document.getElementById('analysis-formula-cell-label') as HTMLSpanElement;
    const lineInput = document.querySelector<HTMLInputElement>('.analysis-line-input')!;

    lineInput.dispatchEvent(new Event('focusin', { bubbles: true }));

    expect(formula.value).toBe('type client');
    expect(label.textContent).toBe('spec1');

    formula.value = 'type server';
    formula.dispatchEvent(new Event('input', { bubbles: true }));

    expect(lineInput.value).toBe('type server');
    expect(app.currentContent()).toBe('type server');
  });

  it('updates the editor buffer when a spreadsheet line is edited', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client\nrelation link(X,Y)');
    syncAnalysisSpreadsheetFromEditor(app);

    const secondLine = document.querySelectorAll<HTMLInputElement>('.analysis-line-input')[1];
    secondLine.value = 'relation semaphore(X)';
    secondLine.dispatchEvent(new Event('input', { bubbles: true }));

    expect(app.currentContent()).toBe('type client\nrelation semaphore(X)');
    expect(app._persistedFileContent).toBe('type client\nrelation semaphore(X)');
    expect(app._updateEditorLabel).toHaveBeenCalled();
    expect(app._invalidateModelState).toHaveBeenCalledWith('editor-change');
  });

  it('toggles a spreadsheet comment cell and rerenders the row marker', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('type client');
    syncAnalysisSpreadsheetFromEditor(app);

    document.querySelector<HTMLButtonElement>('.analysis-comment-toggle')!.click();

    expect(app.currentContent()).toBe('#type client');
    expect(document.querySelector<HTMLInputElement>('.analysis-line-input')!.value).toBe('#type client');
    expect(document.querySelector<HTMLButtonElement>('.analysis-comment-toggle')!.textContent).toBe('#');
  });

  it('omits the comment toggle for the first #lang directive row', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('#lang ivy1.7\ntype client');
    syncAnalysisSpreadsheetFromEditor(app);

    const commentCells = document.querySelectorAll<HTMLTableCellElement>('.analysis-comment-cell');
    const toggles = document.querySelectorAll<HTMLButtonElement>('.analysis-comment-toggle');
    expect(commentCells[0].textContent).toBe('');
    expect(commentCells[0].querySelector('.analysis-comment-toggle')).toBeNull();
    expect(toggles).toHaveLength(1);

    toggleAnalysisSpreadsheetLineComment(app, 0);

    expect(app.currentContent()).toBe('#lang ivy1.7\ntype client');
    expect(toggles[0].disabled).toBe(false);
  });
});
