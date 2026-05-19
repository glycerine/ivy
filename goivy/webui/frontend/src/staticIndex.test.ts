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
    const sliderRule = ivyCss.match(/\.graph-background-slider\s*\{[^}]+\}/)?.[0] || '';

    expect(titleRow?.querySelector('.column-title')?.textContent).toBe('Concept graph');
    expect(titleRow?.querySelector('#graph-background-slider')).toBe(slider);
    expect(slider?.type).toBe('range');
    expect(slider?.min).toBe('0');
    expect(slider?.max).toBe('100');
    expect(slider?.value).toBe('50');
    expect(slider?.getAttribute('aria-label')).toBe('Graph background');
    expect(rootRule).toContain('--ivy-graph-background: rgb(141, 141, 151);');
    expect(graphRule).toContain('background-color: var(--ivy-graph-background);');
    expect(sliderRule).toContain('width: 120px;');
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
