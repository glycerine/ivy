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

describe('static details styling', () => {
  it('preserves newlines in cloned sheet details panes', () => {
    const rule = ivyCss.match(/#info-content,\s*\.info-panel \[id\^="info-content"\]\s*\{[^}]+\}/)?.[0] || '';

    expect(rule).toContain('white-space: pre-wrap;');
  });
});
