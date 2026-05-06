import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeCTIApp() {
  document.body.innerHTML = '';
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  return app;
}

async function click(label) {
  const button = Array.from(document.querySelectorAll('[data-ivy-dialog-button]')).find((btn) => btn.textContent === label);
  expect(button, 'dialog button ' + label).toBeTruthy();
  button.click();
  await Promise.resolve();
}

describe('IvyApp CTI workflows', () => {
  it('prompts for a bounded-check bound and sends it to the backend', async () => {
    const app = makeCTIApp();
    app.currentBound = 3;
    app.api = {
      runCheck: vi.fn(async () => ({ result: 'pass' })),
    };

    const result = app.boundedCheck();
    const input = document.querySelector('[data-ivy-dialog-int]');
    expect(input.value).toBe('3');
    input.value = '7';
    await click('OK');
    await result;

    expect(app.api.runCheck).toHaveBeenCalledWith('bounded', { bound: 7 });
    expect(app.currentBound).toBe(7);
  });

  it('prompts for conjectures before weakening', async () => {
    const app = makeCTIApp();
    app.refreshConceptGraph = vi.fn(async () => undefined);
    app.api = {
      executeAction: vi.fn(async (action) => {
        if (action === 'get_conjectures') {
          return {
            conjectures: [
              { label: 'c0', formula: 'p(X)' },
              { label: 'c1', formula: 'q(X)' },
            ],
          };
        }
        return { removed: ['p(X)'], removed_count: 1 };
      }),
    };

    const result = app.weakenInvariant();
    await Promise.resolve();
    const select = document.querySelector('[data-ivy-dialog-list]');
    select.options[0].selected = true;
    await click('Weaken');
    await result;

    expect(app.api.executeAction).toHaveBeenNthCalledWith(1, 'get_conjectures', {});
    expect(app.api.executeAction).toHaveBeenNthCalledWith(2, 'weaken', { indices: [0] });
    expect(app.controls.lastStatus).toEqual({ message: 'Invariant weakened', kind: 'success' });
  });

  it('dispatches CTI concept graph actions with the active sheet id', async () => {
    const app = makeCTIApp();
    app.activeSheetId = 'sheet-2';
    app.refreshConceptGraph = vi.fn(async () => undefined);
    app.api = {
      executeAction: vi.fn(async () => ({ ok: true, message: 'strengthened' })),
    };

    await app.ctiConceptAction('cti_strengthen');

    expect(app.api.executeAction).toHaveBeenCalledWith('cti_strengthen', { sheet_id: 'sheet-2' });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.lastStatus).toEqual({ message: 'strengthened', kind: 'success' });
  });
});
