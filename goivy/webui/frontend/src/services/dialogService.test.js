import { describe, expect, it, vi } from 'vitest';
import {
  buttonListDialog,
  entryDialog,
  integerDialog,
  listboxDialog,
  okCancelDialog,
  okDialog,
  textDialog,
} from './dialogService.js';

describe('dialogService', () => {
  it('normalizes high-level dialog helpers into dialog-store configs', () => {
    const bridge = {
      showDialog: vi.fn((config) => config),
    };

    expect(okDialog('Title', 'Message', bridge)).toEqual({ type: 'ok', title: 'Title', message: 'Message' });
    expect(okCancelDialog('Confirm', 'Continue?', bridge)).toEqual({
      type: 'okCancel',
      title: 'Confirm',
      message: 'Continue?',
    });
    expect(textDialog('Text', 'Edit:', 'old', { okLabel: 'Use' }, bridge)).toMatchObject({
      type: 'text',
      text: 'old',
      options: { okLabel: 'Use' },
    });
    expect(entryDialog('Name', 'Enter:', 'x', { okLabel: 'Go' }, bridge)).toMatchObject({
      type: 'entry',
      initialValue: 'x',
      options: { okLabel: 'Go' },
    });
    expect(integerDialog('Bound', 'n:', 3, { min: 1 }, bridge)).toMatchObject({ type: 'integer', initialValue: 3 });
    expect(listboxDialog('Pick', 'one:', ['a'], {}, bridge)).toMatchObject({ type: 'listbox', items: ['a'] });
    expect(buttonListDialog('Choose', 'command', [{ label: 'View', value: 'view' }], bridge)).toMatchObject({
      type: 'buttons',
      buttons: [{ label: 'View', value: 'view' }],
    });
    expect(bridge.showDialog).toHaveBeenCalledTimes(7);
  });
});
