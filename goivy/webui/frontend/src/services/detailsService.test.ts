import { describe, expect, it, vi } from 'vitest';
import { populateConstraintFacts } from './detailsService.ts';

describe('detailsService', () => {
  it('renders fact buttons and writes selections through the app API', async () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      api: {
        executeAction: vi.fn(),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    populateConstraintFacts(app, { facts: [{ index: 0, text: 'link(X,Y)', selected: false }] }, { doc: document });

    expect(document.querySelector('.constraint-facts-title').textContent).toBe('Constraints:');
    const fact = document.querySelector('[data-constraint-fact="0"]');
    expect(fact.textContent).toBe('link(X,Y)');

    fact.click();
    await Promise.resolve();

    expect(app.api.executeAction).toHaveBeenCalledWith('set_fact_selection', {
      index: 0,
      selected: true,
    });
    expect(fact.classList.contains('inactive')).toBe(false);
  });
});
