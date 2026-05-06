import { describe, expect, it, vi } from 'vitest';
import { populateConstraintFacts } from './detailsService.js';

describe('detailsService', () => {
  it('routes constraint facts through the Vue bridge', async () => {
    const app = {
      api: {
        executeAction: vi.fn(),
      },
    };
    const bridge = {
      updateConstraintFacts: vi.fn(),
    };

    populateConstraintFacts(app, { facts: [{ index: 2, text: 'p(X)' }] }, { bridge });

    expect(bridge.updateConstraintFacts).toHaveBeenCalledWith(
      [{ index: 2, text: 'p(X)' }],
      expect.any(Function),
    );
    await bridge.updateConstraintFacts.mock.calls[0][1](2, true);
    expect(app.api.executeAction).toHaveBeenCalledWith('set_fact_selection', {
      index: 2,
      selected: true,
    });
  });

  it('renders fallback fact buttons when Vue is unavailable', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      api: {
        executeAction: vi.fn(),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    populateConstraintFacts(app, { facts: [{ index: 0, text: 'link(X,Y)', selected: false }] }, { bridge: null, doc: document });

    expect(document.querySelector('.constraint-facts-title').textContent).toBe('Constraints:');
    expect(document.querySelector('[data-constraint-fact="0"]').textContent).toBe('link(X,Y)');
  });
});
