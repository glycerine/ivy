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
  if (!document.getElementById('info-content')) {
    document.body.insertAdjacentHTML('beforeend', '<div id="info-content"></div>');
  }
});

describe('IvyApp constraint facts', () => {
  it('renders backend facts and toggles selection through the action API', async () => {
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    app.api = { executeAction: vi.fn(async () => ({ status: 'ok' })) };

    app.populateStateCheckboxes({
      relations: [],
      facts: [
        { index: 0, text: 'link(a,b)', selected: true },
        { index: 1, text: 'semaphore(b)', selected: true },
      ],
      toggles: { edges: {}, labels: {} },
    });

    const rows = document.querySelectorAll('[data-constraint-fact]');
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toBe('link(a,b)');
    expect(rows[0].classList.contains('inactive')).toBe(false);

    rows[0].click();
    await Promise.resolve();

    expect(app.api.executeAction).toHaveBeenCalledWith('set_fact_selection', {
      index: 0,
      selected: false,
    });
    expect(rows[0].classList.contains('inactive')).toBe(true);
  });
});
