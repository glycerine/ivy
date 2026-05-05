import { describe, expect, it } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeDialogApp() {
  document.body.innerHTML = '';
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  return new IvyApp();
}

async function click(label) {
  const buttons = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'));
  const button = buttons.find((btn) => btn.textContent === label);
  expect(button, 'dialog button ' + label).toBeTruthy();
  button.click();
  await Promise.resolve();
}

describe('IvyApp dialog primitives', () => {
  it('okDialog resolves true after OK', async () => {
    const app = makeDialogApp();
    const result = app.okDialog('ivyweb', 'hello');

    expect(document.querySelector('[data-ivy-dialog]')).toBeTruthy();
    expect(document.querySelector('.dialog-title').textContent).toBe('ivyweb');
    await click('OK');

    await expect(result).resolves.toBe(true);
    expect(document.querySelector('[data-ivy-dialog]')).toBeFalsy();
  });

  it('okCancelDialog resolves false on cancel', async () => {
    const app = makeDialogApp();
    const result = app.okCancelDialog('Confirm', 'continue?');

    await click('Cancel');

    await expect(result).resolves.toBe(false);
  });

  it('textDialog returns edited text and supports a custom OK label', async () => {
    const app = makeDialogApp();
    const result = app.textDialog('Text', 'edit it', 'old', { okLabel: 'Use', cancel: true });

    const textarea = document.querySelector('[data-ivy-dialog-text]');
    expect(textarea.value).toBe('old');
    textarea.value = 'new';
    await click('Use');

    await expect(result).resolves.toBe('new');
  });

  it('entryDialog returns a typed value', async () => {
    const app = makeDialogApp();
    const result = app.entryDialog('Name', 'graph name', 'old');

    const input = document.querySelector('[data-ivy-dialog-entry]');
    input.value = 'new name';
    await click('OK');

    await expect(result).resolves.toBe('new name');
  });

  it('integerDialog validates min and max before resolving', async () => {
    const app = makeDialogApp();
    const result = app.integerDialog('Bound', 'choose bound', 2, { min: 1, max: 5 });

    const input = document.querySelector('[data-ivy-dialog-int]');
    input.value = '10';
    await click('OK');
    expect(document.querySelector('[data-ivy-dialog-error]').textContent).toContain('at most 5');

    input.value = '4';
    await click('OK');

    await expect(result).resolves.toBe(4);
  });

  it('listboxDialog supports single and multi selection', async () => {
    const app = makeDialogApp();
    const single = app.listboxDialog('Pick', 'one', ['a', 'b', 'c']);
    document.querySelector('[data-ivy-dialog-list]').value = 'b';
    await click('OK');
    await expect(single).resolves.toBe('b');

    const multi = app.listboxDialog('Pick', 'many', ['a', 'b', 'c'], { multiple: true });
    const options = document.querySelectorAll('[data-ivy-dialog-list] option');
    options[0].selected = true;
    options[2].selected = true;
    await click('OK');
    await expect(multi).resolves.toEqual(['a', 'c']);
  });

  it('buttonListDialog resolves the chosen command value', async () => {
    const app = makeDialogApp();
    const result = app.buttonListDialog('Choose', 'command', [
      { label: 'View', value: 'view' },
      { label: 'Ignore', value: 'ignore' },
    ]);

    await click('View');

    await expect(result).resolves.toBe('view');
  });
});
