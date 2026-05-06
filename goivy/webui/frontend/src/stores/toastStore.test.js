import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useToastStore } from './toastStore.js';

describe('toastStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useRealTimers();
  });

  it('adds clickable toast records and expires non-persistent items', () => {
    vi.useFakeTimers();
    const toasts = useToastStore();

    const id = toasts.show('Server connection lost', 'error', { timeoutMs: 25 });

    expect(toasts.items).toEqual([
      expect.objectContaining({ id, message: 'Server connection lost', level: 'error' }),
    ]);

    vi.advanceTimersByTime(25);

    expect(toasts.items).toEqual([]);
    vi.useRealTimers();
  });
});
