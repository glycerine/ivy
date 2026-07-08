import { describe, expect, it, vi } from 'vitest';
import { clearDetailsLog, populateConstraintFacts } from './detailsService.ts';

describe('detailsService', () => {
  it('clears the active details pane to the placeholder', () => {
    document.body.innerHTML = [
      '<div class="sheet-content"><div class="info-panel"><div id="info-content">root stale</div></div></div>',
      '<div class="sheet-content active"><div class="info-panel"><div id="info-content-2" data-ivy-details-kind="selection">stale CTI log</div></div></div>',
    ].join('');

    clearDetailsLog(null, { doc: document });

    expect(document.getElementById('info-content')?.textContent).toBe('root stale');
    expect(document.getElementById('info-content-2')?.textContent).toBe('Select a node or edge to see details');
    expect(document.getElementById('info-content-2')?.getAttribute('data-ivy-details-kind')).toBe('placeholder');
  });

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

  it('applies returned concept snapshots after fact selection changes', async () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const concept = { sheet_id: 'sheet-7', elements: [], facts: [{ index: 0, text: 'link(X,Y)', selected: true }] };
    const app = {
      api: {
        executeAction: vi.fn().mockResolvedValue({ concept }),
      },
      controls: {
        setStatus: vi.fn(),
      },
      activeSheetId: 'sheet-1',
      applyConceptSnapshot: vi.fn(),
    };

    populateConstraintFacts(app, { facts: [{ index: 0, text: 'link(X,Y)', selected: false }] }, { doc: document });
    document.querySelector('[data-constraint-fact="0"]').click();
    await Promise.resolve();
    await Promise.resolve();

    expect(app.applyConceptSnapshot).toHaveBeenCalledWith('sheet-7', concept);
  });

  it('does not erase selected node details when a concept refresh has no facts', () => {
    document.body.innerHTML = '<div id="info-content" data-ivy-details-kind="selection">State 0\nInitial state</div>';
    const app = {
      api: { executeAction: vi.fn() },
      controls: { setStatus: vi.fn() },
    };

    populateConstraintFacts(app, { facts: [] }, { doc: document });

    expect(document.getElementById('info-content')?.textContent).toBe('State 0\nInitial state');
    expect(document.getElementById('info-content')?.getAttribute('data-ivy-details-kind')).toBe('selection');
  });

  it('clears stale constraint details when a later concept refresh has no facts', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      api: { executeAction: vi.fn() },
      controls: { setStatus: vi.fn() },
    };

    populateConstraintFacts(app, { facts: [{ index: 0, text: 'link(X,Y)', selected: true }] }, { doc: document });
    populateConstraintFacts(app, { facts: [] }, { doc: document });

    expect(document.getElementById('info-content')?.textContent).toBe('Select a node or edge to see details');
    expect(document.getElementById('info-content')?.getAttribute('data-ivy-details-kind')).toBe('placeholder');
  });
});
