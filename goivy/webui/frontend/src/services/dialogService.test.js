import { describe, expect, it, vi } from 'vitest';
import { entryDialog, integerDialog, listboxDialog, okDialog } from './dialogService.js';

describe('dialogService', () => {
  it('normalizes high-level dialog helpers into dialog-store configs', () => {
    const bridge = {
      showDialog: vi.fn((config) => config),
    };

    expect(okDialog('Title', 'Message', bridge)).toEqual({ type: 'ok', title: 'Title', message: 'Message' });
    expect(entryDialog('Name', 'Enter:', 'x', { okLabel: 'Go' }, bridge)).toMatchObject({
      type: 'entry',
      initialValue: 'x',
      options: { okLabel: 'Go' },
    });
    expect(integerDialog('Bound', 'n:', 3, { min: 1 }, bridge)).toMatchObject({ type: 'integer', initialValue: 3 });
    expect(listboxDialog('Pick', 'one:', ['a'], {}, bridge)).toMatchObject({ type: 'listbox', items: ['a'] });
    expect(bridge.showDialog).toHaveBeenCalledTimes(4);
  });
});
