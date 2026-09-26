import { describe, expect, it, vi } from 'vitest';
import { setupTutorialSearch } from './tutorialSearchService.ts';

function tick() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('tutorialSearchService', () => {
  it('builds a MiniSearch index from backend tutorial docs and navigates to a selected result', async () => {
    document.body.innerHTML = [
      '<input id="tutorial-search" type="search">',
      '<div id="tutorial-search-results" hidden></div>',
    ].join('');
    const fetchImpl = vi.fn(async () => ({
      ok: true,
      json: async () => ({
        documents: [
          {
            id: 'language',
            title: 'The Ivy language',
            url: '/static/tutorial/kenmcmil.github.io/ivy/language.html',
            body: 'Ivy programs contain actions, objects, and isolates.',
          },
          {
            id: 'proving',
            title: 'Proving protocols',
            url: '/static/tutorial/kenmcmil.github.io/ivy/proving.html',
            body: 'Proof tactics explain invariants and verification conditions.',
          },
        ],
      }),
    }));
    const navigateTo = vi.fn();

    const controller = setupTutorialSearch({
      doc: document,
      fetchImpl: fetchImpl as any,
      navigateTo,
    });
    const input = document.getElementById('tutorial-search') as HTMLInputElement;
    const results = document.getElementById('tutorial-search-results') as HTMLElement;

    input.value = 'isolte';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await tick();
    await controller.ready();
    await tick();

    expect(fetchImpl).toHaveBeenCalledWith('/api/tutorial/search-docs');
    expect(results.hidden).toBe(false);
    expect(results.querySelectorAll('.tutorial-search-result')).toHaveLength(1);
    expect(results.textContent).toContain('The Ivy language');

    (results.querySelector('.tutorial-search-result') as HTMLButtonElement).click();
    expect(navigateTo).toHaveBeenCalledWith('/static/tutorial/kenmcmil.github.io/ivy/language.html');
    expect(input.value).toBe('');
    expect(results.hidden).toBe(true);
  });
});
