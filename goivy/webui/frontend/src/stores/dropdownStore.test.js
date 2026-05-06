import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useDropdownStore } from './dropdownStore.js';

describe('dropdownStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.useRealTimers();
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

  it('tracks a flashing menu item until the flash timer expires', () => {
    vi.useFakeTimers();
    const dropdowns = useDropdownStore();

    dropdowns.flashItem('file-load', 50);

    expect(dropdowns.flashingItemId).toBe('file-load');
    vi.advanceTimersByTime(49);
    expect(dropdowns.flashingItemId).toBe('file-load');
    vi.advanceTimersByTime(1);
    expect(dropdowns.flashingItemId).toBe('');
  });
});
