import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useDialogStore } from './dialogStore.js';

describe('dialogStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('validates integer dialogs before resolving', async () => {
    const dialogs = useDialogStore();
    const result = dialogs.open({
      type: 'integer',
      title: 'Bound',
      message: 'choose',
      initialValue: 1,
      options: { min: 1, max: 5 },
    });

    dialogs.setInputValue('10');
    dialogs.submit();
    expect(dialogs.active.error).toContain('at most 5');

    dialogs.setInputValue('4');
    dialogs.submit();
    await expect(result).resolves.toBe(4);
  });

  it('returns original listbox values for multiple selections', async () => {
    const dialogs = useDialogStore();
    const result = dialogs.open({
      type: 'listbox',
      items: [{ label: 'A', value: { id: 'a' } }, { label: 'B', value: { id: 'b' } }],
      options: { multiple: true },
    });

    dialogs.setSelectedIndices(['0', '1']);
    dialogs.submit();
    await expect(result).resolves.toEqual([{ id: 'a' }, { id: 'b' }]);
  });
});
