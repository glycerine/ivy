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

    expect(baseRule).toContain('rgba(255, 224, 102, 0.78)');
    expect(focusedRule).toContain('rgba(255, 214, 64, 0.88)');
    expect(focusedRule).toContain('!important');
    expect(textRule).toContain('.CodeMirror-selectedtext *');
    expect(textRule).toContain('color: #141423 !important;');
  });
});
