import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useContextMenuStore } from './contextMenuStore.js';

describe('contextMenuStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('stores menu structure and runs item callbacks', () => {
    const menu = useContextMenuStore();
    const callback = vi.fn();

    menu.show(12, 34, [
      { header: 'Execute action:' },
      { name: 'connect', id: 'execute-connect', callback },
    ]);

    expect(menu.visible).toBe(true);
    expect(menu.style.left).toBe('12px');
    expect(menu.items[0].kind).toBe('header');
    menu.runItem(menu.items[1]);

    expect(callback).toHaveBeenCalled();
    expect(menu.visible).toBe(false);
  });
});
