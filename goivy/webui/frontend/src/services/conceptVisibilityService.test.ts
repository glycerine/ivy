import { describe, expect, it, vi } from 'vitest';
import {
  displayConceptName,
  findEdgeVisibility,
  hydrateBackendToggleState,
  applyEdgeVisibility,
  onDisplayClassToggle,
  onEdgeToggle,
  onRelationToggle,
  populateStateCheckboxes,
  renderStateCheckboxes,
  stateRelationRows,
  toggleChecked,
  updateStateLabel,
} from './conceptVisibilityService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';
import { selectConceptGraphView } from '../models/uiDataSelectors.ts';

function modelApp(conceptData) {
  const uiDataModel = new UIDataModel();
  createUIDataModelStore(uiDataModel).applyConceptSnapshot('sheet-1', conceptData);
  return { uiDataModel, activeSheetId: 'sheet-1' };
}

function storeApp(conceptData) {
  const uiDataModel = new UIDataModel();
  const uiDataStore = createUIDataModelStore(uiDataModel);
  uiDataStore.applyConceptSnapshot('sheet-1', conceptData);
  return {
    uiDataModel,
    uiDataStore,
    activeSheetId: 'sheet-1',
    api: {
      setToggles: vi.fn(),
    },
    refreshConceptGraph: vi.fn(),
  };
}

async function flushAsyncToggleBatch() {
  for (let i = 0; i < 10; i += 1) await Promise.resolve();
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

  it('keeps backend relation color metadata without varying relation-name text', () => {
    document.body.innerHTML = '<table><tbody id="state-checkbox-body"></tbody></table>';
    const app = modelApp({
      relations: ['link(X,Y)', 'semaphore'],
      relation_colors: {
        link: '#0000ff',
        semaphore: '#ff0000',
      },
    });

    const rows = stateRelationRows(app);
    expect(rows.find((row) => row.name === 'link(X,Y)').color).toBe('#0000ff');
    expect(rows.find((row) => row.name === 'semaphore').color).toBe('#ff0000');

    renderStateCheckboxes(app, rows, { doc: document });

    const linkButton = document.querySelector('button[data-state-toggle-relation="link(X,Y)"]') as HTMLButtonElement;
    const semaphoreButton = document.querySelector('button[data-state-toggle-relation="semaphore"]') as HTMLButtonElement;
    expect(linkButton.style.getPropertyValue('--ivy-relation-color')).toBe('#0000ff');
    expect(semaphoreButton.style.getPropertyValue('--ivy-relation-color')).toBe('#ff0000');
    expect(linkButton.style.color).toBe('');
    expect(semaphoreButton.style.color).toBe('');
  });

  it('does not let black backend relation colors override relation-name text color', () => {
    document.body.innerHTML = '<table><tbody id="state-checkbox-body"></tbody></table>';
    const app = modelApp({
      relations: ['link(X,Y)'],
      relation_colors: {
        link: '#000000',
      },
    });

    renderStateCheckboxes(app, stateRelationRows(app), { doc: document });

    const linkButton = document.querySelector('button[data-state-toggle-relation="link(X,Y)"]') as HTMLButtonElement;
    expect(linkButton.style.getPropertyValue('--ivy-relation-color')).toBe('#000000');
    expect(linkButton.style.color).toBe('');
  });

  it('reuses the static state header instead of adding a duplicate header row', () => {
    document.body.innerHTML = `
      <table id="state-checkbox-table">
        <thead>
          <tr>
            <th class="chk-col">+</th>
            <th class="chk-col">?</th>
            <th class="chk-col">-</th>
            <th class="chk-col">T</th>
            <th class="name-col"></th>
          </tr>
        </thead>
        <tbody id="state-checkbox-body"></tbody>
      </table>
    `;
    const app = {
      onEdgeToggle: vi.fn(),
      refreshConceptGraph: vi.fn(),
      api: {
        setToggles: vi.fn(),
      },
    };

    populateStateCheckboxes(app, { relations: ['link(X,Y)'] }, { doc: document });

    expect(document.querySelectorAll('thead')).toHaveLength(1);
    expect(document.querySelectorAll('thead tr')).toHaveLength(1);
    expect(Array.from(document.querySelectorAll('thead th')).map((th) => th.textContent.trim())).toEqual([
      '+',
      '?',
      '-',
      'T',
      'Relation',
    ]);

    document.querySelector('th[data-state-toggle-class="all_to_all"]').dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(app.api.setToggles).toHaveBeenCalledTimes(1);
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'link(X,Y)',
      display_class: 'all_to_all',
      value: true,
    });
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

  it('relation-name clicks toggle every class and update graph visibility by bare relation name', async () => {
    document.body.innerHTML = '<table><tbody id="state-checkbox-body"></tbody></table>';
    const app = storeApp({
      relations: ['link(X,Y)'],
      toggles: {
        edges: { link: { all_to_all: false, edge_unknown: false, none_to_none: false, transitive: false } },
        labels: {},
      },
      elements: [
        { group: 'nodes', data: { id: 'client', obj: 'client', label: 'client' } },
        { group: 'nodes', data: { id: 'server', obj: 'server', label: 'server' } },
        {
          group: 'edges',
          classes: 'all_to_all',
          data: { id: 'link-edge', obj: 'link', label: 'link', source: 'client', target: 'server' },
        },
      ],
    });

    renderStateCheckboxes(app, stateRelationRows(app), { doc: document });
    document.querySelector('button[data-state-toggle-relation="link(X,Y)"]').dispatchEvent(new MouseEvent('click', { bubbles: true }));
    await flushAsyncToggleBatch();

    expect(app.api.setToggles).toHaveBeenCalledTimes(4);
    expect(app.api.setToggles).toHaveBeenLastCalledWith({
      edge: 'link(X,Y)',
      display_class: 'transitive',
      value: true,
    });
    expect(app.refreshConceptGraph).toHaveBeenCalledTimes(1);

    const snapshot = app.uiDataModel.sheets['sheet-1'].concept;
    expect(snapshot.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
    expect(selectConceptGraphView(app.uiDataModel.sheets['sheet-1']).edgeVisibilityById['link-edge']).toBe(true);
  });

  it('class-header clicks toggle that class for all relations and mirror unary label state', async () => {
    document.body.innerHTML = '<table><tbody id="state-checkbox-body"></tbody></table>';
    const app = storeApp({
      relations: ['link(X,Y)', 'semaphore(X)'],
      node_labels: ['semaphore'],
      toggles: {
        edges: {
          link: { all_to_all: false },
          semaphore: { all_to_all: false },
        },
        labels: {
          semaphore: { node_necessarily: false },
        },
      },
      elements: [
        { group: 'nodes', data: { id: 'client', obj: 'client', label: 'client', sort: 'client' } },
        { group: 'nodes', data: { id: 'server', obj: 'server', label: 'server', sort: 'server' } },
        {
          group: 'edges',
          classes: 'all_to_all',
          data: { id: 'link-edge', obj: 'link', label: 'link', source: 'client', target: 'server' },
        },
      ],
    });

    renderStateCheckboxes(app, stateRelationRows(app), { doc: document });
    document.querySelector('th[data-state-toggle-class="all_to_all"]').dispatchEvent(new MouseEvent('click', { bubbles: true }));
    await flushAsyncToggleBatch();

    expect(app.api.setToggles).toHaveBeenCalledTimes(2);
    expect(app.api.setToggles).toHaveBeenNthCalledWith(1, {
      edge: 'link(X,Y)',
      display_class: 'all_to_all',
      value: true,
    });
    expect(app.api.setToggles).toHaveBeenNthCalledWith(2, {
      edge: 'semaphore(X)',
      display_class: 'all_to_all',
      value: true,
    });

    const snapshot = app.uiDataModel.sheets['sheet-1'].concept;
    expect(snapshot.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
    expect(snapshot.displayCheckboxes.edgeVisible('semaphore', 'all_to_all')).toBe(true);
    expect(snapshot.displayCheckboxes.nodeLabelVisible('semaphore', 'node_necessarily')).toBe(true);
    expect(selectConceptGraphView(app.uiDataModel.sheets['sheet-1']).edgeVisibilityById['link-edge']).toBe(true);
  });

  it('single-cell toggles keep backend-owned state and graph visibility aligned', async () => {
    const app = storeApp({
      relations: ['link(X,Y)'],
      toggles: {
        edges: { link: { edge_unknown: false } },
        labels: {},
      },
      elements: [
        { group: 'nodes', data: { id: 'client', obj: 'client', label: 'client' } },
        { group: 'nodes', data: { id: 'server', obj: 'server', label: 'server' } },
        {
          group: 'edges',
          classes: 'edge_unknown',
          data: { id: 'link-edge', obj: 'link', label: 'link', source: 'client', target: 'server' },
        },
      ],
    });

    await onEdgeToggle(app, 'link(X,Y)', 'edge_unknown', true);

    const snapshot = app.uiDataModel.sheets['sheet-1'].concept;
    expect(snapshot.displayCheckboxes.edgeVisible('link', 'edge_unknown')).toBe(true);
    expect(selectConceptGraphView(app.uiDataModel.sheets['sheet-1']).edgeVisibilityById['link-edge']).toBe(true);
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

  it('applies real edge visibility to concept edge halos', () => {
    const makeEdge = (data) => ({
      id: vi.fn(() => data.id),
      data: vi.fn((key) => data[key]),
      style: vi.fn(),
    });
    const real = makeEdge({
      id: 'e0',
      obj: 'link',
      source_obj: 'Client',
      target_obj: 'Server',
    });
    const halo = makeEdge({
      id: 'concept_edge_halo_e0',
      halo_for: 'e0',
      halo_obj: 'link',
      halo_source_obj: 'Client',
      halo_target_obj: 'Server',
    });
    const app = modelApp({ elements: [], toggles: {} });
    const view = {
      edgeVisibilityById: {
        'edge:link|Client|Server': true,
      },
    };

    applyEdgeVisibility(app, {
      cy: {
        edges: () => [real, halo],
      },
    }, view);

    expect(real.style).toHaveBeenCalledWith('display', 'element');
    expect(halo.style).toHaveBeenCalledWith('display', 'element');
  });

  it('updates the state label through the DOM', () => {
    document.body.innerHTML = '<span id="state-label"></span>';
    updateStateLabel(null, { doc: document });
    expect(document.getElementById('state-label').textContent).toBe('State: —');
  });
});
