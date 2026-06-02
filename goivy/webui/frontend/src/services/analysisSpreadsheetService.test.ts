import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  linesFromEditorContent,
  lineIsLanguageDirective,
  renderAnalysisSpreadsheet,
  rowsFromEditorContent,
  syncAnalysisSpreadsheetFromEditor,
  toggleLineComment,
} from './analysisSpreadsheetService.ts';

function makeSpreadsheetApp(initialContent = 'type client\nrelation link(X,Y)') {
  let content = initialContent;
  return {
    _persistedFileContent: content,
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

afterEach(() => {
  document.body.innerHTML = '';
});

describe('analysisSpreadsheetService', () => {
  it('creates one spreadsheet row for each editor line', () => {
    expect(linesFromEditorContent('a\nb\n')).toEqual(['a', 'b', '']);
    expect(rowsFromEditorContent('a\nb').map((row) => row.lineNumber)).toEqual([1, 2]);
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
    expect(headers).toEqual(['Line', '#', 'Spec line']);
    expect(lineNumbers).toEqual(['1', '2']);
    expect(inputs.map((input) => input.value)).toEqual(['type client', 'relation link(X,Y)']);
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

  it('disables the comment toggle for the first #lang directive row', () => {
    document.body.innerHTML = '<div id="analysis-spreadsheet-grid"></div>';
    const app = makeSpreadsheetApp('#lang ivy1.7\ntype client');
    syncAnalysisSpreadsheetFromEditor(app);

    const toggles = document.querySelectorAll<HTMLButtonElement>('.analysis-comment-toggle');
    expect(toggles[0].disabled).toBe(true);
    expect(toggles[0].textContent).toBe('#');

    toggles[0].click();

    expect(app.currentContent()).toBe('#lang ivy1.7\ntype client');
    expect(toggles[1].disabled).toBe(false);
  });
});
