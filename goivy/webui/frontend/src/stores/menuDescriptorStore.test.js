import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useMenuDescriptorStore } from './menuDescriptorStore.js';

describe('menuDescriptorStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('normalizes backend menu descriptors and dispatches enabled items', async () => {
    const menus = useMenuDescriptorStore();
    const dispatch = vi.fn().mockResolvedValue({ ok: true });

    menus.setRegion('concept', [
      {
        label: 'Action',
        items: [
          { label: 'Undo', action: 'undo', dispatch: 'action' },
          { type: 'separator' },
          { label: 'Disabled', action: 'disabled', enabled: false },
        ],
      },
    ], dispatch);

    expect(menus.menusFor('concept')[0].contentId).toBe('dynamic-concept-0-action');
    expect(menus.menusFor('concept')[0].items[1].type).toBe('separator');
    menus.toggle('concept', 0);
    expect(menus.isOpen('concept', 0)).toBe(true);

    await menus.runItem('concept', menus.menusFor('concept')[0].items[0]);
    expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ action: 'undo' }));
    expect(menus.isOpen('concept', 0)).toBe(false);

    await menus.runItem('concept', menus.menusFor('concept')[0].items[2]);
    expect(dispatch).toHaveBeenCalledTimes(1);
  });
});
