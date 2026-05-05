import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import {
  FakeAPI,
  FakeControls,
  FakeGraph,
  installSaveDom,
  makeApp,
} from './helpers/fakes.mjs';

function loadApp() {
  return loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
}

beforeEach(() => {
  installSaveDom();
  if (!document.getElementById('state-checkbox-body')) {
    document.body.insertAdjacentHTML('beforeend', '<table><tbody id="state-checkbox-body"></tbody></table>');
  }
});

describe('IvyApp checkbox state', () => {
  it('hydrates state checkboxes and visibility maps from backend concept toggles', () => {
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    app.api = { setToggles: vi.fn(async () => ({ status: 'ok' })) };

    app.populateStateCheckboxes({
      relations: ['link(X,Y)'],
      node_labels: ['semaphore'],
      toggles: {
        edges: {
          'link(X,Y)': {
            all_to_all: false,
            edge_unknown: true,
            none_to_none: false,
            transitive: false,
          },
        },
        labels: {
          semaphore: {
            node_necessarily: true,
            node_maybe: false,
            node_necessarily_not: false,
          },
        },
      },
    });

    const row = document.querySelector('#state-checkbox-body tr');
    const inputs = row.querySelectorAll('input[type="checkbox"]');
    expect(inputs[0].checked).toBe(false);
    expect(inputs[1].checked).toBe(true);
    expect(inputs[2].checked).toBe(false);
    expect(inputs[3].checked).toBe(false);
    expect(app._edgeVisibility['link(X,Y)'].edge_unknown).toBe(true);
    expect(app._labelVisibility.semaphore.node_necessarily).toBe(true);
  });
});
