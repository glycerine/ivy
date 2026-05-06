import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeChoiceApp() {
  document.body.innerHTML = '';
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  app.activeSheetId = 'sheet-1';
  return app;
}

async function click(label) {
  const buttons = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'));
  const button = buttons.find((btn) => btn.textContent === label);
  expect(button, 'dialog button ' + label).toBeTruthy();
  button.click();
  await Promise.resolve();
}

describe('IvyApp ARG choice-backed commands', () => {
  it('prompts for a conjecture and dispatches the selected value', async () => {
    const app = makeChoiceApp();
    app.api = {
      argNodeAction: vi.fn(async () => ({
        choices: [
          { label: 'link(X,Y) -> ~semaphore(Y)', value: 'link(X,Y) -> ~semaphore(Y)' },
          { label: 'other', value: 'other' },
        ],
      })),
    };

    const result = app.prepareArgNodeActionArgs(
      { id: 'state_0', obj: 'state_0' },
      'try_conjecture',
      { sheet_id: 'sheet-1' },
      'sheet-1'
    );
    document.querySelector('[data-ivy-dialog-list]').value = 'other';
    await click('OK');

    await expect(result).resolves.toEqual({ sheet_id: 'sheet-1', conjecture: 'other' });
    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'try_conjecture_choices', { sheet_id: 'sheet-1' });
  });

  it('prompts for a remembered goal and dispatches the selected name', async () => {
    const app = makeChoiceApp();
    app.api = {
      argNodeAction: vi.fn(async () => ({
        choices: [
          { label: 'goal-a', value: 'goal-a' },
          { label: 'goal-b', value: 'goal-b' },
        ],
      })),
    };

    const result = app.prepareArgNodeActionArgs(
      { id: 'state_0' },
      'try_remembered',
      { sheet_id: 'sheet-1' },
      'sheet-1'
    );
    document.querySelector('[data-ivy-dialog-list]').value = 'goal-b';
    await click('OK');

    await expect(result).resolves.toEqual({ sheet_id: 'sheet-1', goal: 'goal-b' });
    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'try_remembered_choices', { sheet_id: 'sheet-1' });
  });

  it('prompts for a graph name before remembering', async () => {
    const app = makeChoiceApp();
    app.api = {
      executeAction: vi.fn(async () => ({ status: 'ok' })),
    };

    const result = app.rememberGraph();
    document.querySelector('[data-ivy-dialog-entry]').value = 'goal-a';
    await click('Remember');
    await result;

    expect(app.api.executeAction).toHaveBeenCalledWith('remember', { name: 'goal-a', sheet_id: 'sheet-1' });
    expect(app.controls.lastStatus).toEqual({ message: 'Graph remembered', kind: 'success' });
  });
});
