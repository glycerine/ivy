import { describe, expect, it, vi } from 'vitest';
import {
  displayConceptName,
  findEdgeVisibility,
  hydrateBackendToggleState,
  onEdgeToggle,
  populateStateCheckboxes,
  stateRelationRows,
  toggleChecked,
  updateStateLabel,
} from './conceptVisibilityService.ts';

describe('conceptVisibilityService', () => {
  it('hydrates backend toggle state and builds relation rows', () => {
    const app = {};
    hydrateBackendToggleState(app, {
      toggles: {
        edges: { 'link(X,Y)': { edge_unknown: true } },
        labels: { semaphore: { node_necessarily: true } },
      },
    });

    expect(toggleChecked(app, 'link(X,Y)', 'edge_unknown')).toBe(true);
    expect(toggleChecked(app, 'semaphore', 'all_to_all')).toBe(true);
    expect(stateRelationRows(app, { relations: ['link(X,Y)'] })).toEqual([
      {
        name: 'link(X,Y)',
        checked: {
          all_to_all: false,
          edge_unknown: true,
          none_to_none: false,
          transitive: false,
        },
      },
    ]);
    expect(stateRelationRows(app, { relations: ['=@X', '=@Y', '=@Z'] }).map((row) => row.name)).toEqual([
      '=@X',
      '=@Y',
      '=@Z',
    ]);
  });

  it('renders relation rows into the DOM table', () => {
    document.body.innerHTML = '<table><tbody id="state-checkbox-body"></tbody></table>';
    const app = {
      _edgeVisibility: {},
      _labelVisibility: {},
      _applyEdgeVisibility: vi.fn(),
      _applyNodeLabels: vi.fn(),
      populateConstraintFacts: vi.fn(),
      onEdgeToggle: vi.fn(),
    };

    populateStateCheckboxes(app, { relations: ['link(X,Y)'] }, { doc: document });

    expect(document.querySelector('.name-col a').textContent).toBe('link(X,Y)');
    document.querySelector('input[value="all_to_all"]').checked = true;
    document.querySelector('input[value="all_to_all"]').dispatchEvent(new Event('change'));
    expect(app.onEdgeToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });

  it('updates local edge and label visibility before notifying the backend', async () => {
    const app = {
      _edgeVisibility: {},
      _labelVisibility: {},
      _applyEdgeVisibility: vi.fn(),
      _applyNodeLabels: vi.fn(),
      refreshConceptGraph: vi.fn(),
      api: {
        setToggles: vi.fn(),
      },
    };

    await onEdgeToggle(app, 'semaphore(X)', 'all_to_all', true);

    expect(app._edgeVisibility['semaphore(X)'].all_to_all).toBe(true);
    expect(app._labelVisibility.semaphore.node_necessarily).toBe(true);
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'semaphore(X)',
      display_class: 'all_to_all',
      value: true,
    });
  });

  it('finds visibility by full or bare relation names and formats equality labels', () => {
    const app = {
      _edgeVisibility: {
        'link(X,Y)': { all_to_all: true },
      },
    };

    expect(findEdgeVisibility(app, 'link', '')).toEqual({ all_to_all: true });
    expect(displayConceptName('=X:client')).toBe('=X');
  });

  it('updates the state label through the DOM', () => {
    document.body.innerHTML = '<span id="state-label"></span>';
    updateStateLabel(null, { doc: document });
    expect(document.getElementById('state-label').textContent).toBe('State: —');
  });
});
