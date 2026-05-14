import { describe, expect, it, vi } from 'vitest';
import {
  displayConceptName,
  findEdgeVisibility,
  hydrateBackendToggleState,
  onDisplayClassToggle,
  onEdgeToggle,
  onRelationToggle,
  populateStateCheckboxes,
  stateRelationRows,
  toggleChecked,
  updateStateLabel,
} from './conceptVisibilityService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

function modelApp(conceptData) {
  const uiDataModel = new UIDataModel();
  createUIDataModelStore(uiDataModel).applyConceptSnapshot('sheet-1', conceptData);
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

    expect(document.querySelector('.name-col button').textContent).toBe('link(X,Y)');
    document.querySelector('input[value="all_to_all"]').checked = true;
    document.querySelector('input[value="all_to_all"]').dispatchEvent(new Event('change'));
    expect(app.onEdgeToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });

  it('bulk toggles relation rows and display classes through one refresh', async () => {
    const app = {
      refreshConceptGraph: vi.fn(),
      api: {
        setToggles: vi.fn(),
      },
    };

    await onRelationToggle(app, 'link(X,Y)', true);
    expect(app.api.setToggles).toHaveBeenCalledTimes(4);
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'link(X,Y)',
      display_class: 'all_to_all',
      value: true,
    });
    expect(app.refreshConceptGraph).toHaveBeenCalledTimes(1);

    app.api.setToggles.mockClear();
    app.refreshConceptGraph.mockClear();
    await onDisplayClassToggle(app, [{ name: 'link' }, { name: 'other' }], 'edge_unknown', false);
    expect(app.api.setToggles).toHaveBeenCalledTimes(2);
    expect(app.api.setToggles).toHaveBeenNthCalledWith(2, {
      edge: 'other',
      display_class: 'edge_unknown',
      value: false,
    });
    expect(app.refreshConceptGraph).toHaveBeenCalledTimes(1);
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
