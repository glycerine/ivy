import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useDropdownStore } from './dropdownStore.js';

describe('dropdownStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('keeps one static dropdown open at a time', () => {
    const dropdowns = useDropdownStore();

    expect(dropdowns.toggle('file-menu')).toBe(true);
    expect(dropdowns.isOpen('file-menu')).toBe(true);
    expect(dropdowns.toggle('arg-inv-menu')).toBe(true);
    expect(dropdowns.isOpen('file-menu')).toBe(false);
    expect(dropdowns.isOpen('arg-inv-menu')).toBe(true);

    dropdowns.closeAll();
    expect(dropdowns.openId).toBe('');
  });
});
