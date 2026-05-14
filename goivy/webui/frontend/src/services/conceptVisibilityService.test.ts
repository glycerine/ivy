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
import { UIDataModel } from '../models/uiDataModel.ts';

function modelApp(conceptData) {
  const uiDataModel = new UIDataModel();
  uiDataModel.acceptConceptSnapshot('sheet-1', conceptData);
  return { uiDataModel, activeSheetId: 'sheet-1' };
}

describe('conceptVisibilityService', () => {
  it('hydrates backend toggle state and builds relation rows', () => {
    const conceptData = {
      relations: ['link(X,Y)'],
      toggles: {
        edges: { 'link(X,Y)': { edge_unknown: true } },
        labels: { semaphore: { node_necessarily: true } },
      },
    };
    const app = modelApp(conceptData);
    hydrateBackendToggleState(app, conceptData);

    expect(toggleChecked(app, 'link(X,Y)', 'edge_unknown')).toBe(true);
    expect(toggleChecked(app, 'semaphore', 'all_to_all')).toBe(true);
    expect(stateRelationRows(app)).toEqual([
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
      populateConstraintFacts: vi.fn(),
      onEdgeToggle: vi.fn(),
    };

    populateStateCheckboxes(app, { relations: ['link(X,Y)'] }, { doc: document });

    expect(document.querySelector('.name-col a').textContent).toBe('link(X,Y)');
    document.querySelector('input[value="all_to_all"]').checked = true;
    document.querySelector('input[value="all_to_all"]').dispatchEvent(new Event('change'));
    expect(app.onEdgeToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });

  it('notifies the backend before refreshing model-owned visibility', async () => {
    const app = {
      refreshConceptGraph: vi.fn(),
      api: {
        setToggles: vi.fn(),
      },
    };

    await onEdgeToggle(app, 'semaphore(X)', 'all_to_all', true);

    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'semaphore(X)',
      display_class: 'all_to_all',
      value: true,
    });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
  });

  it('finds visibility by full or bare relation names and formats equality labels', () => {
    const app = modelApp({
      toggles: {
        edges: { link: { all_to_all: true } },
      },
    });

    expect(findEdgeVisibility(app, 'link', '')).toEqual({
      all_to_all: true,
      edge_unknown: false,
      none_to_none: false,
      transitive: false,
    });
    expect(displayConceptName('=X:client')).toBe('=X');
  });

  it('updates the state label through the DOM', () => {
    document.body.innerHTML = '<span id="state-label"></span>';
    updateStateLabel(null, { doc: document });
    expect(document.getElementById('state-label').textContent).toBe('State: —');
  });
});
