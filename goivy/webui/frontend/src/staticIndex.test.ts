import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const frontendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const indexHtml = readFileSync(path.resolve(frontendDir, '../static/index.html'), 'utf8');
const ivyCss = readFileSync(path.resolve(frontendDir, '../static/css/ivy.css'), 'utf8');

describe('static index toolbar', () => {
  it('does not reserve top-bar space for a verification cancel button', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');

    expect(doc.getElementById('btn-cancel-check')).toBeNull();
    expect(doc.getElementById('btn-cancel-loading')).not.toBeNull();
  });

  it('keeps the right-side isolate, settings, and tutorial controls protected from long filenames', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const right = doc.querySelector('#menubar .menu-right');
    const loadedFileGroup = doc.querySelector('#menubar .loaded-file-group');
    const menuRightRule = ivyCss.match(/\.menu-right\s*\{[^}]+\}/)?.[0] || '';
    const loadedFileRule = ivyCss.match(/\.loaded-file\s*\{[^}]+\}/)?.[0] || '';
    const loadedFileGroupRule = ivyCss.match(/\.loaded-file-group\s*\{[^}]+\}/)?.[0] || '';
    const workflowSelectRule = ivyCss.match(/#ui-mode-select\s*\{[^}]+\}/)?.[0] || '';

    expect(doc.getElementById('btn-undo')).toBeNull();
    expect(loadedFileGroup).not.toBeNull();
    expect(right?.querySelector('#isolate-menu-wrapper')).not.toBeNull();
    expect(right?.querySelector('#btn-toggle-job-control')).not.toBeNull();
    expect(right?.querySelector('#btn-toggle-tutorial')).not.toBeNull();
    expect(menuRightRule).toContain('flex: 0 0 auto;');
    expect(menuRightRule).toContain('min-width: max-content;');
    expect(loadedFileGroupRule).toContain('flex: 1 1 12rem;');
    expect(loadedFileGroupRule).toContain('overflow: hidden;');
    expect(loadedFileRule).toContain('text-overflow: ellipsis;');
    expect(loadedFileRule).toContain('font-size: 11px;');
    expect(workflowSelectRule).toContain('width: 104px;');
  });
});

describe('static graph background controls', () => {
  it('places a midpoint background slider next to the Concept graph title', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const conceptHeader = doc.querySelector('#concept-panel .panel-header');
    const titleRow = conceptHeader?.querySelector('.panel-title-row');
    const slider = doc.getElementById('graph-background-slider') as HTMLInputElement | null;
    const rootRule = ivyCss.match(/:root\s*\{[^}]+\}/)?.[0] || '';
    const graphRule = ivyCss.match(/\.graph-container\s*\{[^}]+\}/)?.[0] || '';
    const statePaneContentRule = ivyCss.match(/#state-panel \.sheet-pane-content\s*\{[^}]+\}/)?.[0] || '';
    const stateRelationNameRule = ivyCss.match(/#state-checkbox-table td\.name-col a,\s*#state-checkbox-table td\.name-col button\s*\{[^}]+\}/)?.[0] || '';
    const titleRowRule = ivyCss.match(/\.panel-title-row\s*\{[^}]+\}/)?.[0] || '';
    const sliderRule = ivyCss.match(/\.graph-background-slider\s*\{[^}]+\}/)?.[0] || '';
    const webkitTrackRule = ivyCss.match(/\.graph-background-slider::-webkit-slider-runnable-track\s*\{[^}]+\}/)?.[0] || '';
    const mozTrackRule = ivyCss.match(/\.graph-background-slider::-moz-range-track\s*\{[^}]+\}/)?.[0] || '';
    const webkitThumbRule = ivyCss.match(/\.graph-background-slider::-webkit-slider-thumb\s*\{[^}]+\}/)?.[0] || '';
    const mozThumbRule = ivyCss.match(/\.graph-background-slider::-moz-range-thumb\s*\{[^}]+\}/)?.[0] || '';
    const readoutRule = ivyCss.match(/\.graph-background-readout\s*\{[^}]+\}/)?.[0] || '';
    const readoutVisibleRule = ivyCss.match(/\.graph-background-readout\.visible\s*\{[^}]+\}/)?.[0] || '';

    expect(titleRow?.querySelector('.column-title')?.textContent).toBe('Concept graph');
    expect(titleRow?.querySelector('#graph-background-slider')).toBe(slider);
    expect(slider?.type).toBe('range');
    expect(slider?.min).toBe('0');
    expect(slider?.max).toBe('100');
    expect(slider?.value).toBe('0');
    expect(slider?.getAttribute('aria-label')).toBe('Graph background');
    expect(slider?.title).toBe('Graph background: rgb(0, 0, 0)');
    expect(rootRule).toContain('--ivy-graph-background: rgb(0, 0, 0);');
    expect(rootRule).toContain('--ivy-relation-name-color: rgb(255, 255, 255);');
    expect(graphRule).toContain('background-color: var(--ivy-graph-background);');
    expect(statePaneContentRule).toContain('background-color: var(--ivy-graph-background);');
    expect(stateRelationNameRule).toContain('color: var(--ivy-relation-name-color);');
    expect(titleRowRule).toContain('width: 100%;');
    expect(sliderRule).toContain('width: 120px;');
    expect(sliderRule).toContain('height: 18px;');
    expect(sliderRule).toContain('margin-left: auto;');
    expect(sliderRule).toContain('background: transparent;');
    expect(sliderRule).toContain('accent-color: var(--ivy-graph-slider-color, var(--ivy-graph-background));');
    expect(webkitTrackRule).toContain('height: 5px;');
    expect(webkitTrackRule).toContain('linear-gradient(');
    expect(webkitTrackRule).toContain('var(--ivy-graph-slider-fill, 0%)');
    expect(webkitTrackRule).toContain('#111 100%');
    expect(mozTrackRule).toContain('height: 5px;');
    expect(mozTrackRule).toContain('linear-gradient(');
    expect(mozTrackRule).toContain('var(--ivy-graph-slider-fill, 0%)');
    expect(mozTrackRule).toContain('#111 100%');
    expect(webkitThumbRule).toContain('width: 18px;');
    expect(webkitThumbRule).toContain('height: 18px;');
    expect(mozThumbRule).toContain('width: 18px;');
    expect(mozThumbRule).toContain('height: 18px;');
    expect(readoutRule).toContain('position: fixed;');
    expect(readoutRule).toContain('pointer-events: none;');
    expect(readoutVisibleRule).toContain('opacity: 1;');
  });
});

describe('static CTI relation controls', () => {
  it('exposes the Python relations-to-minimize text input in the state pane', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const input = doc.getElementById('cti-relations-to-minimize') as HTMLInputElement | null;

    expect(input).not.toBeNull();
    expect(input?.value).toBe('');
    expect(input?.getAttribute('placeholder')).toBe('relation names, space-separated');
    expect(input?.getAttribute('aria-label')).toBe('Relations to minimize');
    expect(input?.getAttribute('title')).toContain('CTI checks and minimization');
    expect(input?.closest('#state-controls')).not.toBeNull();
  });

  it('exposes Python-style abstractor, BMC bound, and transition log controls', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const abstractor = doc.getElementById('analysis-abstractor-select') as HTMLSelectElement | null;
    const bound = doc.getElementById('analysis-bmc-bound') as HTMLSelectElement | null;
    const logFile = doc.getElementById('transition-log-file') as HTMLInputElement | null;
    const abstractorValues = Array.from(abstractor?.querySelectorAll('option') || []).map((option) => option.value);
    const boundValues = Array.from(bound?.querySelectorAll('option') || []).map((option) => option.value);

    expect(abstractorValues).toEqual([
      'ta.Abstractors.top_bottom',
      'ta.Abstractors.concrete',
      'ta.Abstractors.propagate',
      'ta.Abstractors.propagate_and_conjectures',
      'ta.Abstractors.concept_space',
    ]);
    expect(boundValues).toEqual(['1', '3', '5', '10', '15']);
    expect(bound?.value).toBe('3');
    expect(logFile?.getAttribute('aria-label')).toBe('Transition log file');
    expect(logFile?.closest('[data-analysis-controller-controls]')).not.toBeNull();
  });
});

describe('static job control', () => {
  it('places a backend status message above the browser/remote toggle', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const row = doc.querySelector('.job-control-row');
    const children = Array.from(row?.children || []);
    const status = doc.getElementById('job-control-backend-status');
    const toggle = doc.getElementById('job-submission-toggle');
    const statusRule = ivyCss.match(/\.job-control-backend-status\s*\{[^}]+\}/)?.[0] || '';

    expect(status).not.toBeNull();
    expect(status?.hasAttribute('hidden')).toBe(true);
    expect(children.indexOf(status!)).toBeLessThan(children.indexOf(toggle!));
    expect(statusRule).toContain('color: #ff9b9b;');
    expect(statusRule).toContain('text-align: center;');
  });
});

describe('static analysis spreadsheet pane', () => {
  it('does not render the retired analysis spreadsheet pane', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');

    expect(doc.getElementById('analysis-spreadsheet-panel')).toBeNull();
    expect(doc.getElementById('analysis-formula-input')).toBeNull();
    expect(doc.getElementById('analysis-spreadsheet-grid')).toBeNull();
    expect(ivyCss).not.toContain('analysis-spreadsheet');
    expect(ivyCss).not.toContain('analysis-formula');
  });
});

describe('static CodeMirror includes', () => {
  it('loads the CodeMirror 5 search addons before keymaps so Emacs isearch initializes', () => {
    const coreIndex = indexHtml.indexOf('/codemirror.min.js');
    const dialogIndex = indexHtml.indexOf('/addon/dialog/dialog.min.js');
    const cursorIndex = indexHtml.indexOf('/addon/search/searchcursor.min.js');
    const searchIndex = indexHtml.indexOf('/addon/search/search.min.js');
    const markSelectionIndex = indexHtml.indexOf('/addon/selection/mark-selection.min.js');
    const vimIndex = indexHtml.indexOf('/keymap/vim.min.js');
    const emacsIndex = indexHtml.indexOf('/keymap/emacs.min.js');
    const sublimeIndex = indexHtml.indexOf('/keymap/sublime.min.js');

    expect(coreIndex).toBeGreaterThan(-1);
    expect(dialogIndex).toBeGreaterThan(-1);
    expect(dialogIndex).toBeGreaterThan(coreIndex);
    expect(cursorIndex).toBeGreaterThan(dialogIndex);
    expect(searchIndex).toBeGreaterThan(cursorIndex);
    expect(markSelectionIndex).toBeGreaterThan(searchIndex);
    expect(vimIndex).toBeGreaterThan(markSelectionIndex);
    expect(emacsIndex).toBeGreaterThan(markSelectionIndex);
    expect(sublimeIndex).toBeGreaterThan(markSelectionIndex);
  });

  it('defaults the editor keymap preference to Emacs', () => {
    const doc = new DOMParser().parseFromString(indexHtml, 'text/html');
    const checked = doc.querySelector('input[name="keymap"]:checked') as HTMLInputElement | null;

    expect(checked?.value).toBe('emacs');
  });
});

describe('static details styling', () => {
  it('preserves newlines in cloned sheet details panes', () => {
    const rule = ivyCss.match(/#info-content,\s*\.info-panel \[id\^="info-content"\]\s*\{[^}]+\}/)?.[0] || '';

    expect(rule).toContain('white-space: pre-wrap;');
  });

  it('uses a high-contrast yellow CodeMirror selection color', () => {
    const baseRule = ivyCss.match(/#editor-panel \.CodeMirror-selected\s*\{[^}]+\}/)?.[0] || '';
    const focusedRule = ivyCss.match(/#editor-panel \.CodeMirror-focused \.CodeMirror-selected\s*\{[^}]+\}/)?.[0] || '';
    const textRule = ivyCss.match(/#editor-panel \.CodeMirror-selectedtext,\s*#editor-panel \.CodeMirror-selectedtext \*\s*\{[^}]+\}/)?.[0] || '';
    const isearchRule = ivyCss.match(/#editor-panel \.CodeMirror \.ivy-emacs-isearch-match,\s*#editor-panel \.CodeMirror \.ivy-emacs-isearch-match \*\s*\{[^}]+\}/)?.[0] || '';

    expect(baseRule).toContain('rgba(255, 224, 102, 0.78)');
    expect(focusedRule).toContain('rgba(255, 214, 64, 0.88)');
    expect(focusedRule).toContain('!important');
    expect(textRule).toContain('.CodeMirror-selectedtext *');
    expect(textRule).toContain('color: #141423 !important;');
    expect(isearchRule).toContain('rgba(255, 214, 64, 0.88)');
    expect(isearchRule).toContain('color: #141423 !important;');
  });

  it('reserves a bottom minibuffer row for CodeMirror search and replace prompts', () => {
    const editorRule = ivyCss.match(/#editor-panel \.CodeMirror\s*\{[^}]+\}/)?.[0] || '';
    const scrollRule = ivyCss.match(/#editor-panel \.CodeMirror-scroll\s*\{[^}]+\}/)?.[0] || '';
    const fakeScrollbarRule = ivyCss.match(/#editor-panel \.CodeMirror-hscrollbar,\s*#editor-panel \.CodeMirror-scrollbar-filler,\s*#editor-panel \.CodeMirror-gutter-filler\s*\{[^}]+\}/)?.[0] || '';
    const dialogRule = ivyCss.match(/#editor-panel \.CodeMirror-dialog\s*\{[^}]+\}/)?.[0] || '';
    const topDialogRule = ivyCss.match(/#editor-panel \.CodeMirror-dialog-top\s*\{[^}]+\}/)?.[0] || '';

    expect(editorRule).toContain('--ivy-codemirror-minibuffer-height: 24px;');
    expect(scrollRule).toContain('height: calc(100% - var(--ivy-codemirror-minibuffer-height)) !important;');
    expect(scrollRule).toContain('box-sizing: border-box;');
    expect(fakeScrollbarRule).toContain('display: none !important;');
    expect(dialogRule).toContain('min-height: var(--ivy-codemirror-minibuffer-height);');
    expect(dialogRule).toContain('display: flex;');
    expect(topDialogRule).toContain('top: auto !important;');
    expect(topDialogRule).toContain('bottom: 0;');
  });
});

describe('static dropdown styling', () => {
  it('styles dropdown headings as non-clickable inverse labels', () => {
    const rule = ivyCss.match(/\.dropdown-heading\s*\{[^}]+\}/)?.[0] || '';

    expect(rule).toContain('background-color: #c8c8c8;');
    expect(rule).toContain('color: #1b1b1b;');
    expect(rule).toContain('cursor: default;');
    expect(rule).toContain('user-select: none;');
  });
});
