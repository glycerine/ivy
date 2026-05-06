import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
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
  window.__ivyVueBridge = undefined;
  if (!document.getElementById('state-checkbox-body')) {
    document.body.insertAdjacentHTML('beforeend', '<table><tbody id="state-checkbox-body"></tbody></table>');
  }
});

afterEach(() => {
  window.__ivyVueBridge = undefined;
  vi.useRealTimers();
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

  it('routes state relation rows through Vue without appending legacy rows', () => {
    vi.useFakeTimers();
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    app.populateConstraintFacts = vi.fn();
    app._applyEdgeVisibility = vi.fn();
    app._applyNodeLabels = vi.fn();
    window.__ivyVueBridge = {
      updateStateRelations: vi.fn(),
    };

    app.populateStateCheckboxes({
      relations: ['link(X,Y)'],
      toggles: {
        edges: {
          'link(X,Y)': {
            all_to_all: true,
          },
        },
      },
    });
    vi.runAllTimers();

    expect(window.__ivyVueBridge.updateStateRelations).toHaveBeenCalled();
    expect(document.querySelector('#state-checkbox-body tr')).toBeNull();
  });

  it('refreshes the editor layout when the tutorial pane is hidden', () => {
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    document.body.insertAdjacentHTML('beforeend', [
      '<div id="tutorial-container"></div>',
      '<div id="divider-h"></div>',
      '<button id="btn-toggle-tutorial">Hide Tutorial</button>',
    ].join(''));
    app.cmEditor.refresh = vi.fn();
    app.argGraph = { resize: vi.fn() };
    app.conceptGraph = { resize: vi.fn() };

    app.toggleTutorial();

    expect(document.getElementById('tutorial-container').style.display).toBe('none');
    expect(document.getElementById('divider-h').style.display).toBe('none');
    expect(document.getElementById('btn-toggle-tutorial').textContent).toBe('Show Tutorial');
    expect(app.argGraph.resize).toHaveBeenCalled();
    expect(app.conceptGraph.resize).toHaveBeenCalled();
    expect(app.cmEditor.refresh).toHaveBeenCalled();
  });

  it('defers graph resize until Vue has settled the tutorial layout', () => {
    vi.useFakeTimers();
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    document.body.insertAdjacentHTML('beforeend', [
      '<div id="tutorial-container"></div>',
      '<div id="divider-h"></div>',
      '<button id="btn-toggle-tutorial">Hide Tutorial</button>',
    ].join(''));
    app.cmEditor.refresh = vi.fn();
    app.argGraph = { resize: vi.fn() };
    app.conceptGraph = { resize: vi.fn() };

    let settledCallback;
    window.__ivyVueBridge = {
      setTutorialVisible: vi.fn(),
      isTutorialVisible: vi.fn(() => true),
      flashTutorialButton: vi.fn(),
      afterLayoutSettled: vi.fn((callback) => {
        settledCallback = callback;
      }),
    };

    app.toggleTutorial(true);

    expect(window.__ivyVueBridge.setTutorialVisible).toHaveBeenCalledWith(false);
    expect(window.__ivyVueBridge.flashTutorialButton).toHaveBeenCalledWith(1200);
    expect(window.__ivyVueBridge.afterLayoutSettled).toHaveBeenCalled();
    expect(app.argGraph.resize).not.toHaveBeenCalled();
    expect(app.conceptGraph.resize).not.toHaveBeenCalled();

    settledCallback();
    expect(app.argGraph.resize).toHaveBeenCalledTimes(1);
    expect(app.conceptGraph.resize).toHaveBeenCalledTimes(1);
    expect(app.cmEditor.refresh).toHaveBeenCalledTimes(1);

    vi.advanceTimersByTime(60);
    expect(app.argGraph.resize).toHaveBeenCalledTimes(2);
    expect(app.conceptGraph.resize).toHaveBeenCalledTimes(2);
  });

  it('routes Vue-rendered menu flashing through the bridge', () => {
    vi.useFakeTimers();
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    const el = document.createElement('a');
    el.id = 'file-load';
    const callback = vi.fn();
    window.__ivyVueBridge = {
      flashMenuItem: vi.fn(),
      closeDropdownMenus: vi.fn(),
    };

    app.flashAndClose(el, callback);

    expect(window.__ivyVueBridge.flashMenuItem).toHaveBeenCalledWith('file-load', 50);
    expect(el.classList.contains('menu-flash')).toBe(false);
    expect(callback).not.toHaveBeenCalled();

    vi.advanceTimersByTime(50);
    expect(window.__ivyVueBridge.closeDropdownMenus).toHaveBeenCalled();
    expect(callback).toHaveBeenCalled();
  });
});
