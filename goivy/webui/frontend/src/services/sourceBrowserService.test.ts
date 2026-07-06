import { describe, expect, it } from 'vitest';
import { openSourceBrowser } from './sourceBrowserService.ts';

describe('sourceBrowserService', () => {
  it('reuses and raises one source browser while updating file and highlighted line', () => {
    document.body.innerHTML = '<div class="sheet-content active"></div>';
    const app: any = {};

    openSourceBrowser(app, {
      source: 'first\nsecond\nthird',
      file: 'first.ivy',
      lineno: 2,
    }, { doc: document });
    const browser = document.querySelector('[data-source-browser]');
    expect(browser).not.toBeNull();
    expect(browser?.getAttribute('data-source-file')).toBe('first.ivy');
    expect(browser?.getAttribute('data-source-line')).toBe('2');
    expect(browser?.querySelector('[data-source-browser-title]')?.textContent).toContain('first.ivy');
    expect(browser?.querySelector('.source-browser-line-highlight')?.textContent).toBe('second');
    const firstRaiseSeq = Number(browser?.getAttribute('data-raise-seq'));

    openSourceBrowser(app, {
      source: 'alpha\nbeta',
      file: 'second.ivy',
      lineno: 1,
    }, { doc: document });
    const browsers = document.querySelectorAll('[data-source-browser]');
    expect(browsers).toHaveLength(1);
    expect(browsers[0]).toBe(browser);
    expect(browser?.getAttribute('data-source-file')).toBe('second.ivy');
    expect(browser?.getAttribute('data-source-line')).toBe('1');
    expect(browser?.querySelector('.source-browser-line-highlight')?.textContent).toBe('alpha');
    expect(Number(browser?.getAttribute('data-raise-seq'))).toBeGreaterThan(firstRaiseSeq);
    expect(app.sourceBrowserState).toEqual({
      filename: 'second.ivy',
      content: 'alpha\nbeta',
      highlightLine: 1,
      lines: ['alpha', 'beta'],
    });
  });
});
